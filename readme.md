# Go API Template (Golang)

Opinionated starter for building a scalable Go HTTP API with Postgres (and optional RabbitMQ), with a clean transport/service/repository split and consistent JSON responses.

## The Idea

This template aims to stay **flat, explicit, and easy to extend** while keeping clear boundaries:

- **Transport**: thin edge (HTTP handlers / queue consumers). Parse input, call services, render output.
- **Service**: business logic + orchestration (and optional event publishing).
- **Repository**: data access via `sqlx` + SQL (easy to debug and optimize).

Adding a new domain should feel repeatable: create a small slice in transport/service/repository, wire it, ship it.

## Core decisions

- **Router**: `chi` (`github.com/go-chi/chi/v5`) for composable routes and middleware.
- **Responses**: a single renderer (`internal/libs/renderer`) standardizes JSON envelopes.
- **Errors**: return `internal/errors.HTTPError` to get consistent `statusCode` + `errorCode`.
- **Config**: one schema in `config/config.go`, loaded from env; accessible via `config.CONFIG`.
- **Queue**: RabbitMQ is optional; when disabled, publishing becomes a no-op (`NoopPublisher`).

## Features

- **HTTP routing**: `chi` with composable middleware
- **Postgres**: `sqlx` repositories + SQL migrations via `golang-migrate`
- **Queue (optional)**: RabbitMQ publish/consume with routing-key → handler router
- **Config**: environment-driven (supports `-envfilename`)
- **API ergonomics**: standardized JSON success/error envelopes
- **Ops/dev**: Docker Compose (DB + RabbitMQ), Air hot-reload, `golangci-lint`

## Run the HTTP API

Prerequisites: Go 1.26.10, a running SuperTokens Core instance, and Azure credentials
with permission to read the `KaininStoreUSDBConnString` secret from the
`kaininkv` Key Vault. The `cmd/api` entrypoint uses that SQL Server connection
string; the Postgres service and migration runner below are separate template
infrastructure and do not provide the database for this API entrypoint.

1. Create and configure the environment file:

```bash
cp .env.example .env
```

Set `SUPERTOKENS_CONNECTION_URI`, `SUPERTOKENS_APP_NAME`,
`SUPERTOKENS_API_DOMAIN`, and `SUPERTOKENS_WEBSITE_DOMAIN` in `.env`. For a local
frontend, use `http://localhost:8080` as the API domain and the Vite origin as the
website domain. Set `PORT=8080` and include the frontend origin in
`ALLOWED_ORIGINS`.

2. Authenticate with Azure using an identity that can read the connection-string
secret:

```bash
az login
```

3. Start the API:

```bash
go run ./cmd/api
```

## API docs and testing

The interactive API reference is Scalar at `http://localhost:8080/docs`; the
OpenAPI document is at `http://localhost:8080/openapi.json`. Scalar loads its UI
from a CDN, so the browser needs internet access. Public catalog reads can be
tested without signing in, for example `GET /api/products`; write operations
require the appropriate SuperTokens session and permissions.

## Configuration

Create `.env` from `.env.example`. Minimum keys to boot:

- **App**: `ENV`, `PORT`, `VERSION`, `ALLOWED_ORIGINS`
- **SuperTokens**: `SUPERTOKENS_CONNECTION_URI`, `SUPERTOKENS_APP_NAME`, `SUPERTOKENS_API_DOMAIN`, `SUPERTOKENS_WEBSITE_DOMAIN`
- **SuperTokens (optional)**: `SUPERTOKENS_API_KEY`, `SUPERTOKENS_API_BASE_PATH`, `SUPERTOKENS_WEBSITE_BASE_PATH`
- **Square webhook (optional)**: `SQUARE_WEBHOOK_SIGNATURE_KEY`, `SQUARE_WEBHOOK_NOTIFICATION_URL`
- **Square customer addresses**: `SQUARE_TOKEN` (server-side Square access token)
- **Database**: `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`
- **RabbitMQ (optional)**:
  - `RABBITMQ_ENABLED` (`true|false`)
  - `RABBITMQ_URL` (required when enabled)
  - `RABBITMQ_PREFETCH` (optional, default `20`)

Google, Apple, and GitHub sign-in use SuperTokens' dynamic third-party providers.
Configure each provider and its OAuth client credentials in SuperTokens Core, and
register the provider's SuperTokens API callback URL with its OAuth application.

Docker-first defaults (used by `docker-compose.yml`):

- `DB_HOST=go-api-template-db`
- `RABBITMQ_URL=amqp://guest:guest@go-api-template-rabbitmq:5672/`

If you run services outside Docker, switch hosts to `localhost`.

## Huma CRUD API

The API registers typed Huma operations in `internal/transport/http/resources`. Request and response DTOs are kept separate from GORM models; responses omit authentication identifiers and newsletter verification tokens. SuperTokens sessions protect private routes. Catalog reads are public, bag operations are scoped to the signed-in user, and administrative writes require `User.IsAdmin`.

`GET /api/users/me` creates a basic non-admin application user on first access from the authenticated SuperTokens email/password or third-party identity. It never links identities solely by matching email. If that email already belongs to a legacy store account, the API returns `409 Conflict`; link the account through a verified account-linking process rather than changing the SuperTokens ID manually.

Huma serves the OpenAPI document at `/openapi.json` and Scalar at `/docs`. Configure `ALLOWED_ORIGINS` with the exact React app origin. For example:

```ts
const response = await fetch(`${import.meta.env.VITE_API_URL}/api/products`, {
  credentials: "include",
});
if (!response.ok) throw new Error(`API request failed: ${response.status}`);
const products = await response.json();
```

SuperTokens handles authentication endpoints under `/auth`; use its React SDK for sign-in and session management. Login-method records are intentionally not exposed as CRUD routes, and newsletter verification tokens are never returned.

### Square inventory webhook

Configure `SQUARE_WEBHOOK_SIGNATURE_KEY` and `SQUARE_WEBHOOK_NOTIFICATION_URL` from the Square webhook subscription. The notification URL must exactly match the URL registered with Square because the Square SDK verifies the signature over that URL and the original request body. The endpoint is `POST /api/square/webhook`; it does not use a SuperTokens session, but rejects requests with missing or invalid Square signatures. `inventory.count.updated` events with a related catalog ITEM update the matching product, option, and variation records.

### Account shipping address

The authenticated endpoints `GET` and `PUT /api/account/shipping-address` read and update the signed-in user's default address in the Square Customer API. Set `SQUARE_TOKEN` in the API process environment; never expose it to the browser. Before deploying, apply `internal/migrations/sqlserver/20261126_add_square_customer_id.sql` to the API's SQL Server database. The API stores the Square customer ID on the app user and uses the app `UserID` as Square's `reference_id` when linking an existing profile. Order-specific shipping destinations should still be stored with each order.

## Project layout (high level)

```
cmd/                    # Entrypoints (api server, migration runner)
config/                 # Env/config schema + loader
internal/
  server.go             # Wiring + graceful shutdown
  transport/            # HTTP + queue entrypoints (thin)
  service/              # Use-cases / orchestration
  repositories/         # sqlx data access
  model/                # Domain + DB models
  errors/               # Typed HTTP errors + error codes
  libs/                 # Shared infra (db, queue, renderer, utils, crypto)
  migrations/           # SQL migrations
scripts/                # Helper scripts
```

## Architecture (quick tour)

If you’re looking for “where to change what”, start here:

- **Entrypoints**: `cmd/api/main.go`, `cmd/migration/main.go`
- **Composition + lifecycle**: `internal/server.go` (router, middleware, DB, optional queue, graceful shutdown)
- **HTTP wiring**: `internal/transport/http/http_transport.go` (+ per-domain folders under `internal/transport/http/`)
- **Queue wiring**: `internal/transport/queue/queue.go` + `internal/transport/queue/router.go`
- **Migrations**: `internal/migrations/` (plain SQL, `golang-migrate` format)

## Conventions

- **Layering**:
  - **Transport** parses input, calls service, renders output
  - **Service** contains business logic (and publishes events)
  - **Repository** talks to Postgres (`sqlx` + SQL)
- **Response envelopes** (via `internal/libs/renderer`):
  - \(2xx\): `{ "data": <any>, "timestamp": "<utc>" }`
  - \(4xx/5xx\): `{ "errorCode": <int>, "statusCode": <int>, "message": "<string>", "timestamp": "<utc>" }`
- **Errors**: return `internal/errors.HTTPError` to control `statusCode` + `errorCode` consistently.

## Examples included

- **Users HTTP routes**: see `internal/transport/http/users/handler_routes.go` (e.g. `GET /api/users/{id}`, `POST /api/users/`)
- **Auth middleware scaffold**: `internal/transport/http/middlewares/auth.go` (replace with JWT/session validation)
- **Queue handler example**: see `internal/transport/queue/` for a sample routing-key handler

## Adding a new domain (checklist)

For a resource like `orders`:

- **Model**: `internal/model/orders.go`
- **Repo**: `internal/repositories/order_repository.go`
- **Service**: `internal/service/orders.go` (wire in `internal/service/types.go`)
- **HTTP**: `internal/transport/http/orders/` (handler + routes + types/mapper)
- **Wire**: add to `internal/transport/http/http_transport.go` and register under `/api`
- **Queue (optional)**: add routing keys/topology/handlers under `internal/libs/queue` + `internal/transport/queue`

Also: rename the module in `go.mod` and update imports from `go-api-template` to your module path.

## Lint

```bash
golangci-lint run
```

## License

MIT — see [LICENSE](./LICENSE).
