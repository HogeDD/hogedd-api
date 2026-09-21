# ADR 0001: Auth0とNeonによるowner認証基盤

- Status: Accepted
- Date: 2026-09-21

## Context

HogeDDの管理機能を、当面はowner一人だけが利用できるようにする。認証画面、credential管理、token発行を独自実装せず、将来GoogleログインやYouTube連携を追加できる余地を残したい。また、業務データを保存するPostgreSQL接続の土台が必要である。

## Decision

- Auth0をOpenID Connect/OAuth 2.0の認証・token発行基盤として使用する。
- 最初はAuth0 Database Connectionのメールアドレス・パスワード認証を使用し、self-service signupを無効にする。
- `hogedd.com/login` はAuth0の `/authorize` へredirectし、Auth0 Universal Loginを表示する。Auth0の `/u/login` を直接リンクしない。
- APIはaccess tokenの署名、`iss`、`aud`、有効期限を検証する。
- API内のユーザー識別子はtokenの `iss` と `sub` の組み合わせとする。メールアドレスだけで本人判定しない。
- owner認可は、検証済みのidentityがDB上で`active`な`owner`として登録されているかをUse Case境界で確認する。
- PostgreSQLにはNeonを使用し、実行時はpooled connection string、migrationはdirect connection stringを使用する。
- password、access token、refresh tokenはHogeDDのDBへ保存しない。
- GoogleログインはAuth0のSocial Connectionとして後から追加できる。YouTube APIへの認可はログインとは別のGoogle OAuth consentとして設計する。

## Initial user record

`users`はHogeDD内部ID、Auth0のissuer・subject、メールアドレスのsnapshot、role、statusを保持する。`(auth_issuer, auth_subject)`を一意にし、メールアドレス変更がidentity変更にならないようにする。

## Consequences

- credential保護、login UI、password resetはAuth0に任せられる。
- Auth0とNeonへの外部依存が増えるため、障害時の扱い、timeout、ログ、readinessを設計する必要がある。
- 公開APIとliveness `/health` は認証・DBへ依存させない。
- Auth0 tenant、Application、API、callback/logout URLはDevelopment・Preview・Productionで分離して管理する。

## Rejected alternatives

- 独自のメールアドレス・パスワード認証: credential管理のリスクと実装コストが高いため採用しない。
- Google Identity Platformへ直接統一: 有力だが、現時点ではAuth0の運用経験とUniversal Loginを優先する。
- メールアドレスだけのallowlist: メールアドレスは変更可能な属性であり、永続的なidentity keyにしない。
- ログイン画面のURLを隠す: URLの秘匿は認可境界にならないため採用しない。
