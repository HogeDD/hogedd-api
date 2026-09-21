# Architecture

## Goals

- Keep business logic independent from HTTP and Vercel.
- Keep dependency construction in one place.
- Apply operational middleware to every endpoint by default.
- Make infrastructure replaceable through small consumer-owned interfaces.
- Keep liveness independent from databases and external services.

HTTP API contracts follow [API Design Guidelines](api-design.md), based primarily on *Web API: The Good Parts* and updated where current standards or security practices supersede older guidance.

## Dependency direction

```text
api/*, cmd/server -> internal/app -> HTTP adapters -> application -> domain
                         |                               ^
                         `-> infrastructure -> application ports
```

`internal/app` だけが具象実装を組み立てます。HTTP adapter と infrastructure は application が必要とする型や port に合わせますが、domain は HTTP・DB・Vercel を知りません。`health` は業務domainではなく運用機能です。

VercelはGoコードを単一のbackend Functionへ束ねる場合があります。その場合、rewrite後の `/api/...` も共通Routerへ届くため、Routerは公開routeとVercel内部routeの両方を登録します。内部routeは公開API contractではありません。

## Package responsibilities

| Package | Responsibility |
| --- | --- |
| `api/*` | Thin Vercel Function entrypoints; public paths are defined by rewrites |
| `cmd/server` | Local process lifecycle and signal handling |
| `internal/app` | Dependency construction and middleware composition |
| `internal/config` | Environment parsing and validation |
| `internal/content/domain` | App identity, publication state, and business rules |
| `internal/content/application` | Published-App use cases and their required interfaces |
| `internal/content/infrastructure` | Seeded in-memory App source |
| `internal/health` | Transport-independent operational health capability |
| `internal/content/transport/http` | Content固有のHTTP handlersとrequest/response DTO |
| `internal/transport/httpapi` | Context共通のrouting、response、middleware |
| `internal/platform/httpserver` | Standard-library HTTP server lifecycle |

## Domain-driven design

現在実装中のContentコンテキストについては、[処理の流れとinterfaceの図](content-flow.md)を参照してください。

Business functionality is organized by bounded context first, then by responsibility inside that context. We do not build one global `domain`, `service`, or `repository` directory shared by unrelated business concepts.

### Directory ownership

```text
api/                             Vercelの薄いFunction入口
cmd/server/                      ローカル実行入口
internal/
|-- app/                         依存の組み立てとroute登録
|-- config/                      環境変数の読み取り・検証
|-- platform/httpserver/         ローカルHTTPサーバのライフサイクル
|-- transport/httpapi/           全context共通のHTTP基盤
|-- health/                      業務外の運用機能
`-- content/                     Contentというbounded context
    |-- domain/                  App、Slug、公開状態などの業務規則
    |-- application/             Use Case、結果型、必要なport
    |-- infrastructure/          メモリ・将来のDBなどのadapter
    `-- transport/http/          Content固有のhandler・DTO（移動時に追加）
```

Content固有handlerとDTOは `internal/content/transport/http` に置きます。共通 `httpapi` にはmiddleware、共通response、routerなど業務語彙を持たない処理だけを置きます。新しいcontextでも同じ境界を使い、context固有のHTTP表現を共通packageへ集めません。`health` は業務contextの形に無理に合わせません。

全endpointへ適用するmiddlewareは `httpapi.MiddlewareStack` にまとめ、endpoint固有middlewareは `internal/app` でHandlerを組み立てる際に追加します。共通の安全な既定値を保ちながら、認証、CORS、rate limit、cacheなど適用範囲の異なる関心事をendpoint単位で設定できるようにします。

ファイルは役割が検索しやすい名前（例: `list_published_apps_usecase.go`、`memory_app_store.go`、`apps_handler.go`）にします。1ファイルが長くなっただけで階層を増やさず、独立した責務や変更理由が現れた時に分割します。DB導入時も `repository/` や `service/` を先に空で作らず、実装を所有するcontextの `infrastructure` 内で必要な単位に分けます。認証は共通のトークン検証とcontext固有の権限判断を分離し、認証基盤をdomainへ入れません。

```text
internal/<bounded-context>/
|-- domain/          entities, value objects, aggregates, domain services, domain events
|-- application/     commands, queries, use cases, transaction boundaries, ports
|-- infrastructure/  database and external-service adapters
`-- transport/http/  context-specific HTTP input and output mapping (when needed)
```

These directories are created only when the context needs them. Empty layers, marker interfaces, base repositories, and generic CRUD abstractions are avoided because they hide the language and invariants of the domain.

### Dependency rules

- `domain` depends only on the Go standard library and its own context.
- `application` depends on `domain` and declares the ports required by its use cases.
- `infrastructure` implements those ports and may depend on database or vendor SDKs.
- `transport` calls `application`; it does not contain business decisions.
- `internal/app` is the composition root and is the only place that wires concrete implementations together.
- One bounded context must not import another context's infrastructure package. Integration happens through an explicit application port, public contract, or domain event.

Repository interfaces are introduced when an aggregate needs collection-like persistence behavior, not merely because a database has been added. They use domain language such as `FindPublishedArticle` or `Save`, return domain types, and do not expose SQL rows or ORM models. Query-heavy read models may use dedicated query ports instead of forcing every read through an aggregate repository.

Transactions belong to application use cases. Infrastructure supplies the transaction implementation, while domain objects remain unaware of database sessions and transaction handles.

### Modeling rules

- Put invariants in constructors and behavior-rich domain types, not in HTTP handlers.
- Prefer value objects for validated concepts rather than passing primitive strings everywhere.
- Keep aggregate boundaries small and enforce consistency inside one aggregate transaction.
- Use domain events for meaningful completed facts; do not turn every state change into an event.
- Keep API request/response structs, persistence records, and domain models separate.
- Add abstractions at a real substitution or ownership boundary, not solely for test mocking.

## Adding business functionality

1. Name the bounded context and write down its ubiquitous language and invariants.
2. Model the behavior in `internal/<context>/domain` without HTTP or database concerns.
3. Add commands or queries in `internal/<context>/application` and define the smallest ports they consume.
4. Implement required adapters in `internal/<context>/infrastructure`.
5. Translate HTTP input and output in `internal/<context>/transport/http`.
6. Construct concrete dependencies in `internal/app` and register the route.
7. Add a thin file under `api/<route>` when it is deployed as a separate Vercel Function.
8. Test domain invariants, application behavior, adapter contracts, HTTP contracts, and assembled routes at their respective boundaries.

Do not pass `*http.Request`, environment variables, ORM models, or database clients into domain and application code. Avoid global mutable state; only immutable, fully constructed application objects should be package globals in Vercel entrypoints.

## Health semantics

`/health` is a liveness endpoint. Vercel rewrites it internally to the Function at `/api/health`; the internal path is not part of the public API contract. The endpoint only proves that the function can execute and return HTTP. It must remain fast and must not call databases or external APIs.

When dependencies are introduced, add a separate readiness endpoint. Readiness checks should have individual timeouts, run concurrently where appropriate, return a generic public response, and log detailed failures internally.
