# Fracture — Coding Knowledge Base

A technical onboarding guide describing the architecture, structural conventions, and development workflow of the Fracture codebase.

Fracture is a Go HTTP API service. The module is `github.com/lukenguyen/fracture` (Go 1.25.x), built on the Gin web framework with a PostgreSQL data store accessed through `pgx` and type-safe `sqlc`-generated queries.

---

## 1. Architectural Assessment

### 1.1 Overall Architecture

The project is a **single-application monolith** organized around **Clean / Hexagonal Architecture** principles. There is exactly one runnable binary (`cmd/api`), and the code is layered by responsibility with dependencies pointing inward toward the domain.

The dependency direction is: `handler → usecase → repository (interface) → domain`, with `infrastructure` providing the outward-facing implementations of the repository interfaces. The domain layer depends on nothing else in the project.

**Top-level directory layout:**

| Directory | Purpose |
|-----------|---------|
| `cmd/` | Application entry points. Each subdirectory is a runnable binary; `cmd/api` wires dependencies and starts the HTTP server. |
| `internal/` | Private application code, not importable by external modules. Holds the layered architecture (domain, usecase, repository, handler, infrastructure). |
| `pkg/` | Reusable, dependency-light packages with no project-internal coupling (e.g. password hashing, JWT). Importable in principle by other modules. |
| `config/` | Configuration loading and validation. |
| `migrations/` | Versioned SQL schema migrations (golang-migrate format, paired up/down files). |
| `docs/` | Generated OpenAPI/Swagger artifacts plus human-authored onboarding documentation. |
| `bin/` | Compiled build output (git-ignored). |
| `.github/` | Repository tooling: documentation, prompt workflows, conventions, and templates. |
| Root files | `go.mod`/`go.sum` (dependencies), `Makefile` (task automation), `sqlc.yaml` (codegen config), `.env`/`.env.example` (configuration), `docker-compose.yml`. |

### 1.2 Shared Code Conventions

- **Domain model & types** live in `internal/domain`. The core entity types and their lightweight behavior are defined here and reused across all layers. This is the single source of truth for the data shapes the application manipulates.
- **Shared sentinel errors** are centralized in `internal/domain/errors.go` as exported `error` variables. Layers compare against these values to translate failures into the appropriate response (e.g. a not-found error maps to an HTTP 404). This keeps error semantics consistent across the codebase rather than scattering ad-hoc error strings.
- **Repository abstractions** (interfaces) are defined in `internal/repository`, decoupled from any concrete storage technology. Implementations live separately under `internal/infrastructure/persistence`.
- **Cross-domain utilities** with no business coupling are placed under `pkg/` (e.g. `pkg/hash`, `pkg/token`). These packages deliberately wrap third-party libraries so the rest of the codebase never imports the crypto/JWT libraries directly.
- **Request/response DTOs** are declared locally within the handler layer (`internal/handler`) as unexported structs, kept close to the endpoint that uses them and annotated with binding/validation tags.

### 1.3 Cross-Cutting Concerns

| Concern | Approach |
|---------|----------|
| **Configuration** | Centralized in `config/`. Loaded via Viper from a `.env` file with environment-variable override and defaults. A `validate()` step enforces required invariants and applies stricter rules in production (fails fast at startup on misconfiguration). |
| **Authentication / Authorization** | Stateless JWT bearer tokens. Token issuing/verification is isolated in `pkg/token` (HMAC-signed). Route protection is applied through Gin middleware (`internal/handler/middleware`) that validates the `Authorization` header and injects the authenticated identity into the request context. |
| **Password security** | Handled exclusively through `pkg/hash`, which wraps bcrypt. Plaintext passwords are hashed before persistence and never stored or serialized (the domain field is excluded from JSON output). |
| **Exception / error handling** | Errors propagate up as domain sentinel errors and are translated to HTTP status codes at the handler boundary. A shared `respondInternal` helper logs the underlying error server-side and returns a generic message, preventing internal details from leaking to clients. |
| **Logging** | Standard library `log` for lifecycle and error events; Gin's default middleware for request logging. |
| **Persistence access** | A single `pgx` connection pool is created at startup with tuned pool limits and lifetimes, then shared across the application. |
| **Graceful shutdown** | The HTTP server runs in a goroutine; an OS-signal listener triggers a context-bounded `Shutdown`, and the connection pool is closed on exit. |
| **API documentation** | OpenAPI annotations live as comments on handlers and the main entry point; the spec is generated into `docs/` and served via a custom Swagger UI route. |

---

## 2. Application-Level Design Analysis

The single application (`cmd/api`) follows a strict layered design. Each layer has a single responsibility and communicates with adjacent layers through narrow interfaces.

### Entry point — `cmd/api`
- `main.go` is the **composition root**: it loads configuration, constructs the connection pool, instantiates each layer (repository → usecase → handler), wires dependencies via constructor injection, registers routes/middleware, and manages the server lifecycle.
- `swagger.go` mounts API documentation routes.
- Routes are grouped under a versioned base path, with public auth routes and a protected route group guarded by authentication middleware.

### Layer responsibilities (`internal/`)

| Layer | Folder | Responsibility |
|-------|--------|----------------|
| **Domain** | `internal/domain` | The innermost layer. Defines entity types and shared error values. No outward dependencies. |
| **Repository (port)** | `internal/repository` | Declares storage interfaces in terms of domain types. The contract that the usecase layer depends on, independent of any database. |
| **Use case** | `internal/usecase` | Application/business orchestration. Validates input, enforces rules, coordinates repository calls, and handles concerns like pagination normalization. Depends only on repository interfaces and the domain, making it unit-testable with fakes. |
| **Handler (adapter)** | `internal/handler` | HTTP-facing layer. Binds and validates request DTOs, invokes use cases, and maps results and domain errors to HTTP responses. Subpackage `middleware` holds shared Gin middleware. |
| **Infrastructure (adapters)** | `internal/infrastructure` | Concrete outward-facing implementations: database connection (`db`), persistence/repository implementations (`persistence`), and a placeholder for caching (`cache`). |

### Key design patterns observed
- **Dependency Inversion**: Use cases depend on repository *interfaces*; concrete persistence implementations are injected at the composition root. An interface-conformance assertion (`var _ repository.X = (*impl)(nil)`) guarantees implementations satisfy their contract at compile time.
- **Constructor injection**: Every layer exposes a `New…` constructor receiving its dependencies, with no global/singleton state.
- **DTO ↔ domain mapping**: Inbound DTOs are translated to domain types in handlers; storage rows are mapped to domain types in the persistence layer (via a dedicated mapping helper), keeping each layer's types isolated.
- **Generated persistence code**: The persistence implementation is a thin adapter over `sqlc`-generated, type-safe query methods, isolating hand-written code from generated code.

---

## 3. Infrastructure and Data Mapping

### 3.1 Data Layer

**Schema definitions** are managed as versioned SQL migrations under `migrations/` (paired `*.up.sql` / `*.down.sql` files, sequentially numbered, golang-migrate format).

**Entities / schema objects:**
- `users` — the sole persisted table.

**Persistence pipeline:**
- `sqlc` reads the schema directly from the up-migrations and generates type-safe Go query code into `internal/infrastructure/persistence/sqlcgen`. Configuration lives in `sqlc.yaml`, including type overrides that map database types to the same Go types used by the domain layer (so generated rows map cleanly onto domain models).
- Hand-written SQL queries are authored in `internal/infrastructure/persistence/queries` and consumed by the generated `Queries` type.
- The repository implementation (`postgres_user_repo.go`) wraps the generated queries, translates driver-specific errors (e.g. no-rows) into domain sentinel errors, and maps generated row structs to domain entities.

### 3.2 External Dependencies & Integrations

| Dependency | Client / SDK | Role |
|------------|--------------|------|
| **PostgreSQL** | `jackc/pgx/v5` (with `pgxpool`) | Primary data store; accessed via a tuned connection pool. |
| **HTTP framework** | `gin-gonic/gin` | Routing, middleware, request handling. |
| **JWT** | `golang-jwt/jwt/v5` | Stateless authentication tokens. |
| **Password hashing** | `golang.org/x/crypto/bcrypt` | Credential hashing. |
| **Configuration** | `spf13/viper` | Env/file configuration loading. |
| **API docs** | `swaggo/swag` | OpenAPI spec generation from code annotations. |
| **UUIDs** | `google/uuid` | Entity identifiers. |

A `cache` infrastructure package and `docker-compose.yml` exist as scaffolding/placeholders and are not yet implemented.

### 3.3 Development Workflow

All routine tasks are automated through the `Makefile`. The database URL is assembled from `.env` values (with the password URL-encoded), and migration recipes are silenced to avoid leaking credentials into logs.

**Dependencies**
- `make deps` — download and tidy Go module dependencies.

**Run & build**
- `make run` — regenerate docs, then run the API from source (`go run ./cmd/api`).
- `make build` — regenerate docs, then compile the binary to `bin/api`.
- `make clean` — remove build artifacts.

**API documentation**
- `make docs` — regenerate the Swagger/OpenAPI spec into `docs/` (`swag init`).

**Type-safe query generation**
- `make sqlc` — regenerate query code from migrations + `queries/*.sql`.
- `make sqlc-install` — install the `sqlc` CLI (one-time per machine).

**Database migrations** (golang-migrate)
- `make migrate-install` — install the `migrate` CLI (one-time per machine).
- `make migrate-create name=<migration_name>` — scaffold a new up/down migration pair.
- `make migrate-up` — apply all pending migrations.
- `make migrate-down` — roll back the last migration.
- `make migrate-version` — show the current schema version.
- `make migrate-force version=<n>` — force the version to recover from a dirty state.

**Tests**
- Tests use the standard Go toolchain: `go test ./...`. (A `usecase` test file exists as a placeholder; the dependency-inversion design makes the use case layer the natural target for unit tests with fake repositories.)

**Configuration setup**
- Copy `.env.example` to `.env` and populate the `APP_*`, `DB_*`, and `JWT_*` variables before running. Production startup enforces a strong JWT secret and TLS-enabled database connections.

---

## Onboarding Quick Reference

1. Copy `.env.example` → `.env` and fill in values.
2. `make deps` to install dependencies.
3. Install tooling once: `make sqlc-install`, `make migrate-install`.
4. `make migrate-up` to apply the schema.
5. `make run` to start the API; browse `/swagger/` for interactive docs and `/health` for liveness.

**Where to make changes:**
- New entity behavior → `internal/domain`
- New business rule / orchestration → `internal/usecase`
- New endpoint → `internal/handler` (+ register the route in `cmd/api/main.go`)
- New storage operation → add SQL in `queries/`, run `make sqlc`, implement in `internal/infrastructure/persistence`
- Schema change → new migration via `make migrate-create`
