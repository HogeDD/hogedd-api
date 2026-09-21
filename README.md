# hogedd-api

Goで実装したVercel Functions APIです。公開済みアプリの読み取りAPIとヘルスチェックを提供します。

アプリケーション、HTTP transport、ユースケース、実行基盤を分離しています。業務機能は境界づけられたコンテキストを起点にしたDDDで追加します。設計上のルールは [docs/architecture.md](docs/architecture.md)、HTTP APIの設計規約は [docs/api-design.md](docs/api-design.md) を参照してください。

## Endpoint

```http
GET /v1/apps
GET /v1/apps/{slug}
```

一覧は `{ "data": [...] }`、詳細はアプリ1件のJSONを返します。準備中または存在しないslugは`404`です。現在は`hogedd-web`の定義を元にした5件をメモリで保持し、公開済みのClean Tasksだけを返します。契約は [OpenAPI](docs/openapi.yaml)、処理のつながりは [Contentの処理の流れ](docs/content-flow.md) を参照してください。

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

```sh
go run ./cmd/server
curl -i http://localhost:8080/health
```

ポートは `PORT` 環境変数で変更できます。終了シグナルを受けると、処理中のリクエストを待って安全に停止します。

利用できる環境変数:

| Name | Default | Description |
| --- | --- | --- |
| `PORT` | `8080` | ローカルHTTPサーバ専用のポート |
| `APP_ENV` | `VERCEL_ENV` または `development` | 任意の実行環境名 |
| `LOG_LEVEL` | `info` | ローカルとVercel共通。`debug`, `info`, `warn`, `error` |

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

公開APIの `/health` と `/v1/apps` は `vercel.json` により、Vercel内部のGo Functionへrewriteされます。`/v1/apps/{slug}` のslugはrewriteでFunctionへ渡します。利用者に内部の `/api` prefixは見せません。

`cmd/server` と `internal/platform/httpserver` はローカル実行専用です。VercelではTCPポートを待ち受けず、Functionの `Handler` がリクエストごとに呼び出されます。
