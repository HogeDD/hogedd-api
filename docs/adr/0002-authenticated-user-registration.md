# ADR 0002: 認証済みUserの自己登録

- Status: Accepted
- Date: 2026-09-22
- Supersedes: ADR 0001のowner専用role制約とemail一意制約

## Context

Auth0でloginした認証主体を、通知先emailとHogeDD内のrole・statusを持つUserへ紐づけたい。emailは変更される連絡先であり、認証主体の永続的な識別子にはできない。将来はowner以外の利用者も登録するため、schemaをowner一人へ限定できない。

## Decision

- Userの同一性は引き続き`(auth_issuer, auth_subject)`とする。
- emailは通知先のsnapshotとして保存し、一意制約を設けない。
- roleは`owner`、`admin`、`member`とする。
- statusは`active`、`disabled`とする。
- `PUT /v1/users/me`で認証済みUserを冪等に登録する。
- 自己登録で作成するroleは常に`member`とし、requestからrole・statusを受け取らない。
- 既存Userの再登録ではemailと`email_verified`だけを更新し、role・statusを変更しない。
- emailはAccess Tokenへcustom claimとして追加せず、Auth0 `GET /userinfo`から取得する。
- `/userinfo`の`sub`とJWT検証済みIdentityの`sub`が一致しなければ登録しない。
- owner・adminへの昇格は、監査可能な別の管理操作として設計する。
- password、Access Token、Refresh TokenはDBへ保存しない。

## API semantics

- 初回作成は`201 Created`と`Location: /v1/users/me`を返す。
- 登録済みUserの更新は`200 OK`を返す。
- Auth0 profileを取得できない場合は`502 Bad Gateway`を返す。
- DBへ保存できない場合は内部情報を隠した`500 Internal Server Error`を返す。
- responseとserver logへTokenを含めない。

## Consequences

- User登録時にAuth0 `/userinfo`への外部通信が1回発生する。
- `openid email` scopeとRS256のcustom API Access Tokenが必要になる。
- email変更は次回のUser登録実行時にDBへ反映される。
- 同じemailを持つ複数のAuth0 identityをDB上で表現できるため、account linkingは別途設計できる。
- 初期ownerの付与方法とrole変更用管理APIは別Issueで決定する必要がある。

## Rejected alternatives

- emailをAccess Tokenのcustom claimへ追加する: 全API requestへ通知先情報を載せる必要がないため採用しない。
- Next.js sessionのemailをrequest bodyで渡す: Go APIがemailのAuth0由来を独立して確認できないため採用しない。
- emailをUserの一意なidentityにする: email変更や複数identityのaccount linkingを妨げるため採用しない。
- 最初に登録したUserを自動でownerにする: 公開時の競合や設定ミスで権限昇格につながるため採用しない。
