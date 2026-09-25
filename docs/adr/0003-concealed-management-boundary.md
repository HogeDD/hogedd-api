# ADR 0003: 運営APIの存在を404で秘匿する

## 背景

HogeDDには、公開コンテンツの作成・編集・公開などをowner・adminだけが行う運営機能が必要になる。運営APIは一般利用者や第三者clientへ提供するcontractではなく、`hogedd-web`のBFFだけがserver-sideで利用する。

通常のOAuth APIでは、認証情報がない場合は401、権限不足は403を返す。しかし運営APIでは、その応答自体が管理境界の存在を示す。HogeDDは運営機能の発見可能性も抑えつつ、認可そのものはURLの秘匿へ依存させない必要がある。

## 決定

- Webの運営画面は`/manage`、APIの運営境界は`/v1/management/*`とする。
- roleとstatusの正本はHogeDD DBとし、Auth0 Access Tokenの`issuer + subject`からUserを取得する。
- `active`な`owner`・`admin`だけを許可する。
- tokenなし、不正token、未登録、`member`、`disabled`、認可確立前の依存障害は、すべて同じ`404 not_found`として返す。
- 認可失敗では`WWW-Authenticate`や`Allow`など、管理routeの性質を示すheaderを返さない。
- 内部の構造化ログには、token不備、User不在、停止中、権限不足、依存障害を区別して記録する。
- 認可判断はApplication Use Caseで行い、HTTP middlewareやWeb UIだけに依存しない。
- URLの秘匿は補助策とし、すべての運営Use Caseでdefault denyを強制する。

## 採用理由

- 一般利用者へ運営境界の存在を積極的に開示しない。
- BFF専用APIのため、第三者clientが401・403を使い分ける必要がない。
- 外部contractを一つにしても、内部ログで運用上の原因調査ができる。
- roleをrequestやclient表示から受け取らず、DBの状態を毎回確認できる。

## 却下した選択肢

- 401と403を標準どおり返す: 第三者向けAPIでは妥当だが、非公開の運営境界を識別しやすくするため採用しない。
- Webのroute guardだけで制限する: BFFやAPIを直接呼ぶ経路を防げないため採用しない。
- 推測困難なURLだけで保護する: URLは秘密情報ではなく、認可の代替にならないため採用しない。
- Auth0 roleだけで判断する: HogeDD内の利用停止やrole変更をDBで即時反映できないため採用しない。

## 影響

- API利用者は404だけでは認証失敗と権限不足を区別できない。
- 運用者はrequest IDと構造化ログから内部理由を確認する。
- 将来第三者へ運営APIを公開する場合は、別のAPI境界を設計し、401・403を含む標準contractを定義する。
