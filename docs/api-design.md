# API Design Guidelines

This project uses *Web API: The Good Parts* by Takaaki Mizuno as its primary API-design reference. The goal is an API that is easy to use, easy to change, robust, and unsurprising.

The book provides the design philosophy. Current HTTP specifications, security guidance, and platform constraints take precedence where practices have changed since publication.

## Contract first

- Define the resource, use case, invariants, authorization rules, and failure cases before choosing a URI.
- Treat a published request or response shape as a compatibility contract.
- Add or update an OpenAPI document when the first business endpoint is introduced.
- Keep transport DTOs separate from domain objects and persistence records.

## URI design

- Business APIs use the `/api/v1` prefix. Operational endpoints such as `/api/health` are unversioned.
- Use lowercase plural nouns for resource collections: `/api/v1/articles`.
- Represent relationships with shallow hierarchy: `/api/v1/articles/{article_id}/comments`.
- Do not put verbs, implementation names, file extensions, or UI terminology in URIs.
- Use hyphens when a multiword URI segment is unavoidable.
- Use query parameters for filtering, sorting, field selection, and pagination, not to identify the primary resource.
- Avoid deeply nested resources. Prefer stable resource IDs once nesting exceeds one relationship.

## HTTP semantics

| Method | Intended use | Safe | Idempotent |
| --- | --- | --- | --- |
| `GET` | Read a resource or collection | Yes | Yes |
| `HEAD` | Read response metadata | Yes | Yes |
| `POST` | Create a resource or execute a non-idempotent command | No | No |
| `PUT` | Replace a resource at a known URI | No | Yes |
| `PATCH` | Partially update a resource | No | Depends on the operation |
| `DELETE` | Remove a resource | No | Yes |

- Implement method semantics, not CRUD naming conventions.
- Return `Allow` with `405 Method Not Allowed`.
- Support idempotency keys for retryable `POST` operations with financial or otherwise irreversible effects.
- Do not use `GET` for state changes.

## Status codes

- `200 OK`: successful read or update with a response body.
- `201 Created`: resource created; include `Location` when a canonical URI exists.
- `202 Accepted`: asynchronous work accepted but not completed.
- `204 No Content`: successful operation with no response body.
- `400 Bad Request`: malformed syntax or structurally invalid input.
- `401 Unauthorized`: authentication is missing or invalid.
- `403 Forbidden`: authenticated caller is not allowed to perform the operation.
- `404 Not Found`: resource is absent or intentionally concealed.
- `409 Conflict`: request conflicts with current resource state.
- `422 Unprocessable Content`: syntactically valid input violates field or business validation.
- `429 Too Many Requests`: rate limit exceeded; include retry information where possible.
- `500 Internal Server Error`: unexpected server failure without exposing internal details.
- `503 Service Unavailable`: temporary inability to serve, including failed readiness.

Do not return `200 OK` with an error encoded only in the response body.

## JSON conventions

- Use `application/json; charset=utf-8`.
- Use `snake_case` field names consistently.
- Encode timestamps as RFC 3339 strings in UTC.
- Treat identifiers as opaque strings in public contracts.
- Distinguish absent, `null`, empty, and zero values intentionally.
- Do not change an existing field's type or meaning.
- Ignore unknown response fields on clients; reject unknown request fields when accepting them would hide caller mistakes.

Single resources may be returned directly. Collections use a stable envelope when pagination or metadata is present:

```json
{
  "data": [],
  "pagination": {
    "next_cursor": null
  }
}
```

Prefer cursor pagination for mutable or large collections. Bound every client-controlled page size.

## Error contract

Errors use one machine-readable shape:

```json
{
  "error": {
    "code": "validation_failed",
    "message": "request validation failed",
    "request_id": "opaque-request-id",
    "details": [
      {
        "field": "title",
        "code": "required"
      }
    ]
  }
}
```

- `code` is stable and intended for programmatic handling.
- `message` is safe for users but is not a stable programmatic contract.
- `request_id` connects a client-visible failure to server logs.
- `details` is optional and contains safe, structured validation information.
- Never expose stack traces, SQL errors, credentials, internal hostnames, or dependency responses.

## Compatibility and versioning

- Prefer additive changes: new optional fields, new resources, and new methods.
- Do not remove fields, tighten validation, change defaults, or alter semantics within a version without a migration plan.
- Version the public business API in the URI with a major version only.
- Do not create a new major version for additive changes.
- Announce deprecations and define an observable retirement date before removal.

## Security and operations

- Serve production APIs only over HTTPS.
- Authenticate before authorization; enforce authorization at the use-case boundary.
- Validate size, format, range, and allowed values for all external input.
- Configure CORS with an explicit allowlist when browser access is required. Do not enable it speculatively.
- Apply rate limits according to caller identity and endpoint cost when public or abuse-sensitive APIs are introduced.
- Never log credentials, tokens, cookies, or sensitive request bodies.
- Propagate request IDs and emit structured logs without exposing internals to clients.
- Keep liveness fast and dependency-free; put dependency checks in readiness.
- Define explicit timeouts for outbound calls and propagate request cancellation.

## Review checklist

- Does the URI describe a resource in domain language?
- Do method, status code, and retry behavior match HTTP semantics?
- Are authentication, authorization, validation, and rate-limit behavior defined?
- Is every error represented by the common error contract?
- Is pagination bounded and deterministic?
- Are timestamps, IDs, optional values, and field names consistent?
- Is the change backward compatible within the current version?
- Are logs useful without containing sensitive data?
- Are the OpenAPI contract and boundary tests updated?
