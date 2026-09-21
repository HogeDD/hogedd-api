# AGENTS.md

このファイルは、`hogedd-api` を変更するコーディングエージェント向けの作業規約です。人間向けの概要は `README.md`、設計判断の詳細は `docs/` を正本として参照してください。

## プロジェクト概要

- Goで実装したHogeDDのWeb APIです。
- Production URLは `https://api.hogedd.com` です。
- 現在の公開エンドポイントは `/health`、`GET /v1/apps`、`GET /v1/apps/{slug}` です（各GETはHEADにも対応）。
- Vercel Functionsへデプロイします。
- 業務機能はDDDを前提に、境界づけられたコンテキスト単位で追加します。
- HTTP API設計は『Web API: The Good Parts』を主な基準とし、現在のHTTP仕様とセキュリティ慣行を優先します。

## 最初に読むもの

作業前に、変更範囲に応じて次を確認してください。

1. `README.md`: 起動、テスト、環境変数、デプロイ方法
2. `docs/architecture.md`: DDD、依存方向、package責務
3. `docs/api-design.md`: URI、HTTP、JSON、エラー、互換性の規約
4. 対象packageの `doc.go` と既存テスト

文書と実装が食い違う場合は、推測で片方へ合わせず、意図を確認して同じ変更内で整合させてください。

## プロジェクトの進め方

このリポジトリは `main` をProductionへ対応させるGitHub Flowで進めます。長期間存続する開発branchや環境別branchは作りません。

1. Issueまたは依頼内容から、目的、対象外、受け入れ条件を確認します。
2. 最新の `main` を基点に、目的が一つだけの短命branchを作ります。
3. Domainへ影響する場合は、実装前に用語、invariant、aggregate境界、失敗条件を整理します。
4. 小さく実装し、同じbranchでテスト、GoDoc、関連文書を更新します。
5. エージェントは変更内容と検証結果を利用者へ報告します。明示的な依頼がある場合に限り、commitやpushなど指定されたGit操作を続けます。
6. Push後にVercel Previewで公開contractと実環境固有の挙動を確認します。
7. Pull Requestで設計意図、変更内容、検証結果、互換性、残課題を共有します。
8. Review完了後に `main` へmergeし、Productionのhealth checkと変更対象を確認します。
9. Merge済みbranchは削除します。

不確実な大規模変更は一つの巨大なPRにせず、後戻り可能な順序へ分割します。ただし、動かない中間状態や利用されない抽象化だけを先行してmergeしません。

### Issueと設計判断

- Bugは再現条件、期待結果、実際の結果を記録します。
- Featureは利用者、ユースケース、受け入れ条件、対象外を明確にします。
- Domain変更では業務用語とinvariantを先に合意します。
- API変更ではrequest、response、error、認証、冪等性、互換性を実装前に確認します。
- DB schema、認証方式、外部公開contractなど、変更コストが高い判断は `docs/adr/` にADRとして残します。
- ADRは背景、決定、採用理由、却下した選択肢、影響を簡潔に記録します。

## ブランチ規約

branchは作業者の種類にかかわらず、変更種別と短い目的をkebab-caseで表します。

```text
feature/add-article-api
fix/reject-invalid-cursor
refactor/split-health-transport
docs/update-api-guidelines
chore/update-go-version
```

- `feature/`: 利用者または運用者へ新しい価値を追加する変更
- `fix/`: 不具合やcontract違反の修正
- `refactor/`: 外部挙動を変えない内部構造の改善
- `docs/`: 文書だけの変更
- `chore/`: dependency、tooling、CI、設定などの保守

- branch名にIssue番号を含める場合は `feature/123-add-article-api` のように種別の直後へ置きます。
- `main` へ直接commitしません。
- 一つのbranchへ無関係な変更を混在させません。
- 作業開始後に別目的へ広がった場合は、別Issue／branchへ切り分けます。
- branch作成前に未commit変更を確認し、他者の変更を取り込んだり破棄したりしません。

## コミット規約

エージェントは利用者からその操作を明示的に依頼された場合に限り、`git commit`、`git push`、Pull Request作成、merge、tag作成を行います。コード変更の依頼だけを、commitやpushの許可まで含むものとして解釈してはいけません。依頼がない限りGitの外部状態を変更せず、作業完了時に変更ファイル、検証結果、未完了事項を報告します。

- 一つのcommitには、説明可能で動作する一つの論理変更を含めます。
- Commit messageは日本語で簡潔に、何をしたかが分かる形にします。
- `修正`、`対応`、`update` だけの曖昧なmessageを避けます。
- Formatだけの変更や無関係なrenameを機能変更へ混ぜません。
- Secret、`.env`、`.vercel/`、binary、coverage出力をcommitしません。
- Commit前に少なくとも変更packageのテストを実行し、PR前にはリポジトリ全体を検証します。

例:

```text
記事公開条件をドメインモデルで検証
カーソル不正時の400レスポンスを追加
Vercel公開パスをhealthへ変更
```

## Pull Request規約

- 一つのPRは一つの目的に限定し、差分をreview可能な大きさに保ちます。
- 大きな作業はDraft PRを早めに作り、設計方向を確認します。
- PR本文には最低限、目的、主な変更、検証方法、API／DB／環境変数への影響を記載します。
- API contractを変える場合はrequest／response例と互換性判断を記載します。
- DB migrationがある場合は適用順序、rollback方針、既存データへの影響を記載します。
- Review指摘への対応で別の設計判断が生じた場合は、会話だけで終わらせずコードコメント、文書、ADRの適切な場所へ反映します。
- 原則としてsquash mergeを使い、`main` の一commitが一つのPRの目的を表すようにします。
- 必須検証が失敗中、Review未完了、Preview未確認のPRはmergeしません。

PR本文の基本形:

```markdown
## 目的

## 変更内容

## 検証

## 影響

## 残課題
```

## 環境とリリース

- Local: 実装と高速なフィードバックに使用します。
- Preview: Pull Request単位の結合確認に使用します。Productionの秘密値を流用しません。
- Production: `main` の正常系だけを反映します。
- Feature branchからProductionへ直接deployしません。
- 緊急修正も `fix/...` branchとPRを経由し、検証を省略しません。
- Production反映後は `/health`、変更したendpoint、Vercel logsを確認します。
- Rollbackが必要な場合は、原因調査より先に直前の正常deploymentへ戻して影響を止めます。その後、修正PRを作成します。
- 公開contractや永続データを伴う変更は、アプリケーションのrollbackだけで戻せるとは限らないため、前方互換な手順で段階的にreleaseします。

## 基本コマンド

```sh
# ローカル起動
go run ./cmd/server

# ヘルスチェック
curl -i http://localhost:8080/health

# 通常の品質確認
gofmt -w <変更したGoファイル>
go test ./...
go vet ./...

# 共有処理、並行処理、状態を持つ処理を変更した場合
go test -race ./...
```

テストや静的検査を実行できなかった場合は、完了報告で理由を明記してください。

## アーキテクチャ規約

- `internal/app` をコンポジションルートとし、具象依存の生成を集約します。
- DomainとApplicationをHTTP、Vercel、環境変数、DB、ORM、外部SDKへ依存させません。
- HTTP DTO、Domain Model、永続化レコードを同じ型で兼用しません。
- インターフェースは利用側が所有し、必要最小限の操作だけを宣言します。
- RepositoryはDBを導入しただけでは作りません。AggregateやUse Caseが永続化ポートを必要とした時に追加します。
- 汎用的な`BaseRepository`、空の抽象化、業務語彙を消すCRUD層を作りません。
- トランザクション境界はApplication層へ置き、DomainへDBセッションを渡しません。
- 複数のbounded contextを一つのグローバルなdomain packageへ混在させません。
- `internal/health` は業務ドメインではなく運用機能です。

業務機能の基本構成は次のとおりです。必要になったdirectoryだけを追加してください。

```text
internal/<bounded-context>/
|-- domain/
|-- application/
|-- infrastructure/
`-- transport/http/
```

Content固有handlerとDTOは `internal/content/transport/http` に置きます。共通 `internal/transport/httpapi` にはmiddleware、共通response、routerなど業務語彙を持たない処理だけを置きます。詳細は `docs/architecture.md` の Directory ownership を参照してください。

## HTTP API規約

- 公開ホストが `api.hogedd.com` のため、公開パスへ `/api` prefixを付けません。
- 業務APIは `/v1/...`、運用エンドポイントはバージョン対象外です。
- URIは原則として小文字・複数形の名詞を使います。
- HTTP method、status code、冪等性を正しく対応させます。
- JSON fieldは `snake_case`、時刻はUTCのRFC 3339、IDはopaqueな文字列として扱います。
- エラーは `internal/transport/httpapi.Responder` の共通形式を使います。
- 公開済みcontractは原則として加算的に変更し、破壊的変更を黙って入れません。
- 最初の業務APIを追加する際にOpenAPI文書を導入し、以後は実装と同時に更新します。
- 詳細は `docs/api-design.md` を参照してください。

## HTTP実装

- 現在はGo標準の `net/http` を使用しています。
- RouterやframeworkはTransport層の実装詳細です。導入しても `*http.Request` やframework固有ContextをApplication／Domainへ渡しません。
- 共通middlewareはrequest ID、security header、access log、panic recoveryをすべてのrouteへ適用します。
- 新しいrouteは単体テストに加え、`internal/app` の組み立てテストでも到達可能性を確認します。
- `HEAD`、`405 Method Not Allowed`、`Allow` header、共通エラー形式を考慮します。
- Client入力を無制限に読み込まず、body size、timeout、collection sizeへ上限を設けます。

## ヘルスチェック

- 公開パスは `/health` です。`/healthz` へ戻しません。
- Livenessとして高速かつ外部依存なしに保ちます。
- DBや外部APIの確認を `/health` に追加しません。必要になったらreadinessを別エンドポイントとして設計します。
- 監視側はJSON本文ではなくHTTP status codeで成功を判定できるようにします。

## Vercel

- `api/<route>/index.go` はVercel Functionの薄いエントリーポイントです。業務ロジックを書きません。
- `cmd/server` と `internal/platform/httpserver` はローカル実行専用です。
- VercelではTCP portをlistenせず、exportされた `Handler` が呼ばれます。
- 公開パスとVercel内部の `/api/...` の対応は `vercel.json` のrewriteで管理します。
- routeを追加したら、公開URI、Function配置、rewrite、README、API文書をまとめて確認します。
- Previewは `vercel`、Productionは `vercel --prod` です。明示的に依頼されていないProduction deployを行いません。
- Productionの疎通確認は `https://api.hogedd.com` を使います。
- `.vercel/` はローカルproject metadataのためコミットしません。

## 環境変数と秘密情報

- 環境変数の正本はVercel Project Settingsです。
- `Development`、`Preview`、`Production`で値を分離します。
- 秘密値、token、credential、実際の接続文字列をcommit、log、テストfixtureへ含めません。
- `.env` と `.env.*` はcommitしません。共有するのは安全な例だけを置いた `.env.example` です。
- 新しい環境変数は `internal/config` で読み取り、必要なら起動時に検証します。
- Vercelとローカルで共通の設定は `LoadRuntime`、ローカルサーバ固有設定は `Load` で扱います。
- 必須設定が不正な場合、危険なdefaultへ黙ってfallbackさせません。

## Goコーディング規約

- `gofmt`を必ず適用します。
- 標準ライブラリで十分な場合は標準ライブラリを優先します。依存追加には保守性、security、binary size、cold startへの理由が必要です。
- Errorは意味を付けてwrapし、公開レスポンスへ内部情報を露出させません。
- `context.Context` はrequest scopeのキャンセルとdeadline伝播に使い、構造体へ保存しません。
- Mutableなpackage globalを作りません。Vercel entrypointのglobalは完全構築済みで並行利用可能なobjectに限ります。
- Exportedな型、関数、method、fieldには、識別子名から始まる日本語GoDocコメントを書きます。
- コメントは処理の逐語訳ではなく、責務、制約、利用条件、戻り値の意味を説明します。
- ログは `log/slog` の構造化ログを使い、秘密情報やrequest body全体を記録しません。

## テスト方針

- Domain: invariant、value object、aggregate behaviorをtable-driven testで確認します。
- Application: portをstub/fakeへ置き換え、成功、業務エラー、キャンセルを確認します。
- Infrastructure: 実装がport contractを満たすことをintegration testで確認します。
- Transport: method、path、status、header、JSON contract、validation errorを確認します。
- Composition: `internal/app` でroute登録とmiddleware適用漏れを確認します。
- Bug修正では、可能な限り先に再現テストを追加します。
- 時刻、乱数、外部I/Oへ依存するテストは決定的にします。

## 変更時の完了条件

変更を完了する前に、該当する項目を確認してください。

- コード、テスト、GoDoc、設計文書が整合している
- `gofmt`、`go test ./...`、`go vet ./...` が成功している
- 並行性へ影響する場合は `go test -race ./...` が成功している
- 公開API変更では互換性、status code、error contract、OpenAPIへの影響を確認した
- 新しい環境変数は `.env.example` とREADMEへ安全な例を追加した
- Vercel route変更ではrewriteと公開URLを確認した
- 秘密情報、`.env`、`.vercel/`、生成物が差分へ入っていない
- 依頼範囲外の既存変更を戻していない
