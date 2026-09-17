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
api/health         cmd/server
      \               /
       internal/app              composition root
          /     \
  transport      platform        delivery and infrastructure
       |
    health                       operational capability
```

Dependencies point inward. The `health` package does not import HTTP, Vercel, configuration, or infrastructure packages.

The current system has no business domain yet. `health` is deliberately not placed in a domain layer: liveness is an operational concern, not a business rule.

## Package responsibilities

| Package | Responsibility |
| --- | --- |
| `api/health` | Thin Vercel Function entrypoint |
| `cmd/server` | Local process lifecycle and signal handling |
| `internal/app` | Dependency construction and middleware composition |
| `internal/config` | Environment parsing and validation |
| `internal/health` | Transport-independent operational health capability |
| `internal/transport/httpapi` | Routing, HTTP handlers, responses, and middleware |
| `internal/platform/httpserver` | Standard-library HTTP server lifecycle |

## Domain-driven design

Business functionality is organized by bounded context first, then by responsibility inside that context. We do not build one global `domain`, `service`, or `repository` directory shared by unrelated business concepts.

```text
internal/<bounded-context>/
|-- domain/          entities, value objects, aggregates, domain services, domain events
|-- application/     commands, queries, use cases, transaction boundaries, ports
|-- infrastructure/  database and external-service adapters
`-- transport/http/  context-specific HTTP input and output mapping
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
