# hogedd-api

Goで実装したVercel Functions APIです。公開済みアプリの読み取りAPIとヘルスチェックを提供します。

アプリケーション、HTTP transport、ユースケース、実行基盤を分離しています。業務機能は境界づけられたコンテキストを起点にしたDDDで追加します。設計上のルールは [docs/architecture.md](docs/architecture.md)、HTTP APIの設計規約は [docs/api-design.md](docs/api-design.md) を参照してください。

## Endpoint

```http
GET /v1/apps
GET /v1/apps/{slug}
GET /v1/me
PUT /v1/users/me
GET /v1/users/me/profile
PUT /v1/users/me/profile
GET /v1/management/me
GET /v1/management/apps
POST /v1/management/apps
```

一覧は `{ "data": [...] }`、詳細はアプリ1件のJSONを返します。準備中または存在しないslugは`404`です。現在は`hogedd-web`の定義を元にした5件をメモリで保持し、公開済みのClean Tasksだけを返します。契約は [OpenAPI](docs/openapi.yaml)、処理のつながりは [Contentの処理の流れ](docs/content-flow.md) を参照してください。

`GET /v1/me`はAuth0のAccess Tokenを要求し、検証済みの`issuer`と`subject`を返します。
認証環境変数が両方未設定の環境では、このendpointだけがすべてのTokenを`401`で拒否します。

`PUT /v1/users/me`は認証済みの主体をHogeDD Userとして冪等に登録します。emailは同じAccess TokenでAuth0 `/userinfo`から取得し、初回は`member`として`201`、登録済みならrole・statusを維持して連絡先snapshotを更新し`200`を返します。このendpointには認証環境変数と`DATABASE_URL`が必要です。

`GET /v1/users/me`は検証済みの`issuer + subject`に紐づく登録済みUserを返します。未登録の場合は`404`です。読み取り時はAuth0 `/userinfo`を呼ばず、PostgreSQLの保存済みsnapshotを返します。

`GET /v1/users/me/profile`と`PUT /v1/users/me/profile`は、認証情報や権限とは分離した本人編集可能なプロフィールを取得・保存します。現在の項目は1〜50文字の`display_name`だけです。

`GET /v1/management/me`は`active`な`owner`・`admin`だけが利用できます。管理境界の存在を秘匿するため、tokenなし、不正token、未登録、`member`、`disabled`、認可確認失敗は同じ`404 not_found`を返します。内部ログでは拒否理由を区別します。
roleの付与・剥奪と緊急停止は、管理GUIを導入するまで[User権限の運用](docs/user-access-operations.md)に従います。

`GET /v1/management/apps`と`POST /v1/management/apps`は同じ秘匿認可を適用し、公開準備中を含むApp一覧の取得とdraft作成を提供します。管理用AppはPostgreSQLへ保存し、公開APIの既存メモリデータとはまだ分離されています。

```sh
curl -i http://localhost:8080/v1/apps
curl -i http://localhost:8080/v1/apps/clean-tasks
```

## Health

```http
GET /health
HEAD /health
```

正常時は `200 OK` を返します。

```json
{"status":"ok"}
```

このエンドポイントはプロセスがリクエストを処理できることだけを確認します。将来データベースなどを追加しても、外部依存の状態確認は readiness 用の別エンドポイントへ分離します。

## Local development

Host上で直接起動する場合:

```sh
go run ./cmd/server
curl -i http://localhost:8080/health
```

ポートは `PORT` 環境変数で変更できます。終了シグナルを受けると、処理中のリクエストを待って安全に停止します。

Dockerで起動する場合:

```sh
docker build -f Dockerfile.dev -t hogedd-api-dev .
docker run --rm --init \
  --publish 8080:8080 \
  --mount type=bind,source="$PWD",target=/workspace \
  hogedd-api-dev
```

`Dockerfile.dev`はLocal開発専用です。Airが`.go`ファイルの変更を検知し、APIを自動で再build・再起動します。認証やDBを使う場合は、必要な環境変数を`docker run`へ明示的に渡してください。複数serviceをまとめて起動する手順は`hogedd-local`で管理します。

利用できる環境変数:

| Name | Default | Description |
| --- | --- | --- |
| `PORT` | `8080` | ローカルHTTPサーバ専用のポート |
| `APP_ENV` | `VERCEL_ENV` または `development` | 任意の実行環境名 |
| `LOG_LEVEL` | `info` | ローカルとVercel共通。`debug`, `info`, `warn`, `error` |
| `AUTH0_ISSUER_URL` | なし | Auth0 tenantのissuer。末尾の`/`を含むHTTPS URL |
| `AUTH0_AUDIENCE` | なし | Auth0で設定したHogeDD APIの識別子 |
| `DATABASE_URL` | なし | Neonのpooled connection string。DBを使う処理で必須 |
| `DATABASE_URL_UNPOOLED` | なし | Neonのdirect connection string。migration実行時だけ使用 |
| `DATABASE_MIGRATION_URL` | なし | migration先を明示的に上書きする場合だけ使用 |
| `METRICS_INGEST_TOKEN` | なし | Web BFFから匿名起動を記録するための共有秘密値。未設定時は計測endpointだけ404 |
| `DATABASE_MAX_OPEN_CONNS` | `5` | 1インスタンスが保持する最大DB接続数 |
| `DATABASE_MAX_IDLE_CONNS` | `2` | 1インスタンスが保持する最大idle接続数 |

すべてのHTTPレスポンスにrequest ID、キャッシュ抑止、セキュリティヘッダーが付与されます。アクセスログはJSON形式で標準出力へ出力されます。

### Environment variables

環境変数の正本はVercel Project Settingsとし、`Development`、`Preview`、`Production`ごとに値を分離します。秘密値をリポジトリへコミットしてはいけません。

```sh
# Vercelに保存された名前と適用環境を確認する（値は表示しない）
vercel env ls

# 秘密値をファイルへ保存せず、Development環境の値でローカル起動する
vercel env run -- go run ./cmd/server

# 必要な場合だけDevelopment環境の値を.gitignore済みファイルへ取得する
vercel env pull .env.local
```

新しい環境変数を追加するときは [`.env.example`](.env.example) にキー名と安全なサンプル値だけを追加します。Go標準ライブラリは `.env.local` を自動では読み込まないため、通常は `vercel env run` またはシェルから環境変数を渡します。

### Database migrations

PostgreSQL schemaは `internal/platform/postgres/migrations` の連番SQLで管理します。Application起動時にはmigrationを実行せず、deploy前の独立した操作として適用します。

```sh
DATABASE_URL_UNPOOLED='postgresql://...' go run ./cmd/migrate
```

Neonでは、API実行時の `DATABASE_URL` にpooled connection string、migration用の `DATABASE_URL_UNPOOLED` にdirect connection stringを設定します。Vercel Marketplace連携では両方が自動設定されます。接続文字列をshell historyやログへ残さず、`vercel env run -- go run ./cmd/migrate` でDevelopment環境へ適用できます。

### Local database seed

Local PostgreSQLへ代表的なUser状態を投入する場合:

```sh
APP_ENV=development \
DATABASE_URL='postgresql://hogedd:hogedd@localhost:5432/hogedd?sslmode=disable' \
go run ./cmd/seed
```

owner、admin、member、disabled memberの4パターンを冪等に保存します。接続先がDockerの`db`、`localhost`、loopback以外の場合は実行を拒否します。

ログイン可能なLocal Auth0ユーザーは、`SEED_USERS_FILE`で指定したversion 1のJSONから複数追加できます。実際のidentity情報はGitへ含めず、`hogedd-local`のseedディレクトリで管理します。

## Test

```sh
go test ./...
go test -race ./...
go vet ./...
```

## Deploy to Vercel

Vercel CLIでプロジェクトを紐づけてデプロイします。

```sh
npm install --global vercel
vercel
vercel --prod
```

公開APIの `/health`、`/v1/apps`、`/v1/management/me`などは `vercel.json` により、Vercel内部のrouteへrewriteされます。VercelがGoコードを単一のbackend Functionとして束ねる場合にも対応するため、共通Routerがrewrite後の内部routeを処理します。`/v1/apps/{slug}` のslugはquery parameterで内部routeへ渡し、Routerがpath valueへ変換します。利用者に内部の `/api` prefixは見せません。

`cmd/server` と `internal/platform/httpserver` はローカル実行専用です。VercelではTCPポートを待ち受けず、Functionの `Handler` がリクエストごとに呼び出されます。
