# Developer Onboarding Guide

This guide helps new developers work effectively in the **Fracture** project, especially for:
- Building new APIs
- Adding new models
- Creating and managing database migrations

Goal: follow the existing codebase workflow — **Clean Architecture + sqlc + golang-migrate + Swagger**.

---

## 1. Architecture Overview

The project follows a layered (Clean Architecture) flow. Dependencies point **inward**: outer layers depend on inner ones, never the reverse.

```
HTTP request
   │
   ▼
handler  ──────────►  receives request/response (Gin), binds & validates DTOs,
   │                  maps domain errors → HTTP status codes
   ▼
usecase  ──────────►  business logic, input validation, orchestration
   │
   ▼
repository (interface)  ─►  persistence contract (defined in internal/repository)
   │
   ▼
infrastructure/persistence  ─►  implements the interface with PostgreSQL + sqlc
   │
   ▼
sqlcgen (generated)  ─►  type-safe queries over a pgx connection pool
```

- `domain` sits at the center: entities + domain errors, no HTTP/DB dependencies.
- `handler`, `usecase`, `repository` depend on `domain`, not on each other's concrete types.
- The persistence layer is the **only** place that touches SQL / the pgx pool.

Dependency direction:
`handler → usecase → repository (interface) → persistence (impl) → sqlcgen → pgx pool`

---

## 2. Tech Stack

| Concern | Choice |
|---|---|
| Language | Go (version pinned in `go.mod`) |
| HTTP framework | Gin |
| Database | PostgreSQL |
| DB driver | `pgx/v5` (`pgxpool`) |
| Query codegen | `sqlc` (type-safe Go from raw SQL) |
| Schema migrations | `golang-migrate` |
| API docs | `swaggo/swag` (OpenAPI) + custom dark Swagger UI |
| Config | `viper` (reads `.env` + environment) |
| UUIDs | `google/uuid` |

---

## 3. Project Layout

```
fracture/
├── cmd/api/
│   ├── main.go            # entrypoint: config, DI wiring, routes, graceful shutdown
│   └── swagger.go         # custom dark-themed Swagger UI + /swagger/doc.json
├── config/
│   └── config.go          # viper-based config loader (.env → Config struct)
├── docs/                  # generated OpenAPI spec (swagger.json/yaml, docs.go) + this guide
├── migrations/            # golang-migrate up/down SQL files
├── internal/
│   ├── domain/            # entities (user.go) + shared errors (errors.go)
│   ├── repository/        # repository INTERFACES only
│   ├── usecase/           # business logic
│   ├── handler/           # Gin HTTP handlers + request DTOs + Swagger annotations
│   └── infrastructure/
│       ├── db/postgres.go         # pgxpool setup
│       ├── cache/redis.go         # redis client
│       └── persistence/
│           ├── postgres_user_repo.go   # repository implementation
│           ├── queries/                # sqlc input: hand-written SQL
│           │   └── users.sql
│           └── sqlcgen/                # sqlc output: GENERATED, do not edit by hand
├── sqlc.yaml              # sqlc configuration
└── Makefile               # all dev commands
```

---

## 4. Environment Requirements

Required tools:
- Go (version from `go.mod`)
- PostgreSQL (local or Docker)
- `migrate` CLI (golang-migrate)
- `sqlc` CLI
- `swag` CLI

Quick setup:

```bash
make deps            # go mod download + tidy
make migrate-install # install golang-migrate CLI
make sqlc-install    # install sqlc CLI
go install github.com/swaggo/swag/cmd/swag@latest
```

---

## 5. Environment Variables

Copy `.env.example` → `.env` and fill in values. The Makefile reads `.env` to build the DB connection URL for migrations.

| Variable | Meaning | Example |
|---|---|---|
| `APP_PORT` | HTTP port the API listens on | `8080` |
| `APP_ENV` | `development` / `production` (toggles Gin release mode) | `development` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5532` |
| `DB_USER` | DB user | `tridentity_user` |
| `DB_PASSWORD` | DB password | `tridentity` |
| `DB_NAME` | Database name | `fracture_db` |
| `DB_SSLMODE` | `disable` / `require` / ... | `disable` |

> `.env` is gitignored; `.env.example` is the committed template. Keep them in sync when adding a new key.

---

## 6. Run The Project Locally

1. Ensure PostgreSQL is running and `.env` is configured.

2. Apply migrations (creates the schema):

```bash
make migrate-up
```

3. Regenerate sqlc code (only needed if schema/queries changed):

```bash
make sqlc
```

4. Run the API:

```bash
make run     # runs `make docs` first, then `go run ./cmd/api`
```

Endpoints:
- Swagger UI: `http://localhost:<APP_PORT>/swagger/`
- OpenAPI spec: `http://localhost:<APP_PORT>/swagger/doc.json`
- Health check: `GET /health`
- User API base path: `/api/v1`

The server installs a graceful-shutdown handler (SIGINT/SIGTERM) with a 5s timeout, defined in `cmd/api/main.go`.

---

## 7. Layer Responsibilities & Patterns

### Domain (`internal/domain`)
Plain entities and shared errors. No framework imports.

```go
type User struct {
    ID        uuid.UUID `json:"id"`
    Email     string    `json:"email"`
    Password  string    `json:"-"`   // never serialized in API responses
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

Shared errors live in `internal/domain/errors.go`:
`ErrInvalidID`, `ErrNotFound`, `ErrBadRequest`, `ErrConflict`, `ErrUnauthorized`.

> **Password convention:** `Password` has `json:"-"`, so it is never returned by any user API (get/create/update). Hashing and auth-related reads are deferred to the auth feature.

### Repository interface (`internal/repository`)
Declares the persistence contract the usecase depends on. No SQL here.

### Persistence (`internal/infrastructure/persistence`)
Implements the interface using sqlc-generated queries. The standard pattern:

```go
type postgresUserRepo struct {
    q *sqlcgen.Queries
}

var _ repository.UserRepository = (*postgresUserRepo)(nil) // compile-time check

func NewPostgresUserRepo(db *pgxpool.Pool) repository.UserRepository {
    return &postgresUserRepo{q: sqlcgen.New(db)}
}

// userToDomain maps the sqlc row → domain model.
// NOTE: name the mapper per-entity (userToDomain, articleToDomain, ...) —
// `persistence` is a shared package, a generic `toDomain` would collide.
func userToDomain(u sqlcgen.User) *domain.User { /* ... */ }

func (r *postgresUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
    u, err := r.q.GetUserByID(ctx, id)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, domain.ErrNotFound   // translate driver error → domain error
        }
        return nil, err
    }
    return userToDomain(u), nil
}
```

### Usecase (`internal/usecase`)
Business rules, validation, metadata. It uses a **check-then-act** pattern for conflicts:

```go
func (uc *UserUseCase) CreateUser(ctx context.Context, user *domain.User) error {
    if user.Email == "" || user.Name == "" {
        return domain.ErrBadRequest
    }
    user.ID = uuid.New()
    now := time.Now().UTC()
    user.CreatedAt, user.UpdatedAt = now, now

    if _, err := uc.userRepo.FindByEmail(ctx, user.Email); err == nil {
        return domain.ErrConflict
    } else if err != domain.ErrNotFound {
        return err
    }
    return uc.userRepo.Create(ctx, user)
}
```

### Handler (`internal/handler`)
Binds request DTOs, calls usecase, maps domain errors → HTTP. Note the current style uses **direct `==` comparison** on sentinel errors:

```go
if err := h.userUc.CreateUser(ctx, &user); err != nil {
    if err == domain.ErrConflict   { c.JSON(409, gin.H{"error": err.Error()}); return }
    if err == domain.ErrBadRequest { c.JSON(400, gin.H{"error": err.Error()}); return }
    c.JSON(500, gin.H{"error": err.Error()}); return
}
```

> Because of `==`, errors must be returned **unwrapped** to be matched. If you ever wrap with `fmt.Errorf("...: %w", err)`, switch the handler/usecase checks to `errors.Is`.

---

## 8. Workflow: Add A New API (Existing Model)

Example: `GET /api/v1/users/by-email/:email`.

1. **Add the SQL query** in `internal/infrastructure/persistence/queries/users.sql`:
   ```sql
   -- name: GetUserByEmail :one
   SELECT id, email, password, name, created_at, updated_at
   FROM users
   WHERE email = $1;
   ```
   Annotation cheatsheet: `:one` (single row), `:many` (slice), `:exec` (no rows), `:execrows` (rows affected).

2. **Regenerate sqlc**: `make sqlc` → new funcs/types appear in `sqlcgen/`.

3. **Extend the repository interface** in `internal/repository/`.

4. **Implement** in `internal/infrastructure/persistence/`: call the generated method, map `pgx.ErrNoRows → domain.ErrNotFound`, map row → domain via the entity mapper.

5. **Add a usecase method**: validate input, call the repo, apply business rules.

6. **Add handler + route**: handler method in `internal/handler/`, register in `cmd/api/main.go`, return status codes mapped from domain errors.

7. **Swagger annotations** on the handler (`@Summary`, `@Param`, `@Success`, `@Failure`, `@Router`) → `make docs`.

8. **Verify**: `make run`, hit the endpoint, confirm `/swagger/` reflects the change.

---

## 9. Workflow: Add A New Model (End-To-End)

Example: model `Article`.

1. **Migration**: `make migrate-create name=create_articles_table` → write `.up.sql` (create) and `.down.sql` (drop). Apply with `make migrate-up`.
2. **Domain**: add `internal/domain/article.go` (data + domain behavior only).
3. **Repository interface**: `internal/repository/article_repository.go`.
4. **Queries**: `internal/infrastructure/persistence/queries/articles.sql` (`CreateArticle`, `GetArticleByID`, `ListArticles`, `UpdateArticle`, `DeleteArticle`) → `make sqlc`.
5. **Persistence impl**: `internal/infrastructure/persistence/postgres_article_repo.go` — struct holds `*sqlcgen.Queries`, compile-time interface check, `articleToDomain` mapper, `ErrNoRows → ErrNotFound`.
6. **Usecase**: `internal/usecase/article_usecase.go` — inject repo interface, validate, set ID/timestamps.
7. **Handler**: `internal/handler/article_handler.go` — request DTOs, bind+validate, map domain errors → HTTP.
8. **Wire in `cmd/api/main.go`**: init repo + usecase + handler, register route group under `/api/v1`.
9. **Swagger + verify**: annotations → `make docs` → `make run` → test.

---

## 10. Migrations (golang-migrate) — Detailed

| Command | Purpose |
|---|---|
| `make migrate-create name=<n>` | create a new `.up.sql`/`.down.sql` pair |
| `make migrate-up` | apply all pending migrations |
| `make migrate-down` | roll back the last migration (1 step) |
| `make migrate-version` | show current version |
| `make migrate-force version=<n>` | reset a dirty state to version `n` |

How it works:
- The Makefile builds `DB_URL` from `.env` and passes it to the `migrate` CLI.
- golang-migrate tracks state in a `schema_migrations` table (current version + `dirty` flag).
- A migration that fails halfway leaves the DB **dirty**; fix the SQL, then `make migrate-force version=<last_good>` and re-run `migrate-up`.

Rules:
- Always write **both** up and down files; the down must reverse the up.
- **Never edit a migration already merged to main** — add a new one instead.
- Verify on a fresh database before merging.
- Migration files may contain multiple statements (they run in one transaction for Postgres).

---

## 11. sqlc — Detailed

Config: `sqlc.yaml`.

- `schema: "migrations/*.up.sql"` — sqlc reads the schema **straight from the up-migrations** (the `.down.sql` files are excluded so their `DROP` statements don't confuse it). This keeps the generated models in lock-step with the real schema.
- `queries: internal/infrastructure/persistence/queries` — your hand-written SQL.
- Output package `sqlcgen` (`sql_package: pgx/v5`), with `emit_json_tags` and `emit_interface`.
- **Type overrides** map Postgres types to the same types the domain uses:
  - `uuid → github.com/google/uuid.UUID`
  - `timestamptz → time.Time`

Gotchas:
- These overrides are **global**. They are correct for NOT NULL columns. If you add a **nullable** `uuid`/`timestamptz` column later, the override will map it to a non-pointer type and mishandle `NULL` — add a column-specific override at that point.
- `sqlcgen/` is **generated code — never edit by hand.** Re-run `make sqlc` after any schema or query change. If a query references a column that doesn't exist in the schema, `make sqlc` fails *at generation time* (this is the safety net that prevents struct ↔ DB drift).

---

## 12. Swagger / API Docs

- General API info (title, version, base path, host) lives as annotations above `main()` in `cmd/api/main.go`.
- Per-endpoint annotations live above each handler in `internal/handler/`.
- `make docs` runs `swag init -g cmd/api/main.go -o docs`, regenerating `docs/swagger.json`, `docs/swagger.yaml`, `docs/docs.go`. `make run`/`make build` run `make docs` automatically.
- Serving (`cmd/api/swagger.go`): a **custom dark-themed UI** is served (the stock swaggo UI is light-only with no CSS hook). Routes:
  - `GET /swagger/` and `/swagger/index.html` → the UI
  - `GET /swagger/doc.json` → the spec via `swag.ReadDoc()`
  - `GET /swagger` → redirects to `/swagger/`

---

## 13. Database Connection

`internal/infrastructure/db/postgres.go` builds a `pgxpool.Pool` with:
- `MaxConns = 20`, `MinConns = 5`
- `MaxConnLifetime = 1h`, `MaxConnIdleTime = 30m`
- a `Ping` on startup (the process `log.Fatalf`s if the DB is unreachable)

The pool is created once in `main.go` and injected into repositories.

---

## 14. Code Conventions

- Shared domain errors are defined once in `internal/domain/errors.go`.
- `usecase`/`handler` must **never** touch the DB directly — only through the repository interface.
- Translate driver errors (`pgx.ErrNoRows`) into domain errors **inside the persistence layer**.
- Any schema/query change → `make sqlc`. Any annotation change → `make docs`.
- Keep all API routes under `/api/v1`.
- Name sqlc→domain mappers per entity (`userToDomain`, `articleToDomain`), not a generic `toDomain`.
- Sentinel errors are matched with `==` today — return them unwrapped (or migrate the codebase to `errors.Is`).

---

## 15. Error → HTTP Reference

| Domain error | HTTP status |
|---|---|
| `ErrBadRequest`, `ErrInvalidID` | 400 |
| `ErrUnauthorized` | 401 |
| `ErrNotFound` | 404 |
| `ErrConflict` | 409 |
| (anything else) | 500 |

---

## 16. Known Gaps / Tech Debt

Be aware of these when extending the code:
- **Passwords are stored in plaintext.** Hashing must be added with the auth feature before any real use.
- **Conflict detection has a TOCTOU race.** `CreateUser`/`UpdateUser` check `FindByEmail` then insert; two concurrent requests can both pass the check, and the second hits a UNIQUE violation that currently surfaces as **500 instead of 409**. To harden, map Postgres error code `23505` → `domain.ErrConflict` in the repository.
- **`Update`/`Delete` on a missing row** silently succeed at the repo level (the `:exec` queries ignore rows-affected). The usecase guards this with a prior `FindByID`. For repo-level safety, use `:execrows` and return `ErrNotFound` when 0 rows change.
- **No tests yet.** `internal/usecase/user_usecase_test.go` is effectively empty.

---

## 17. Pre-PR Checklist

- [ ] Migration has valid up **and** down files, applies on a fresh DB
- [ ] `make sqlc` re-run and generated code committed
- [ ] `make docs` re-run (Swagger up to date)
- [ ] New API wired in `cmd/api/main.go`, routes under `/api/v1`
- [ ] Handler maps domain errors → correct HTTP status
- [ ] `.env.example` updated if a new env key was added
- [ ] `go build ./...` and `go vet ./...` pass
- [ ] App runs locally (`make run`) and endpoints tested manually
- [ ] Usecase unit tests added (happy path + validation + domain errors)

---

## 18. Troubleshooting

| Symptom | Likely cause / fix |
|---|---|
| `relation "users" does not exist` | Migrations not applied → `make migrate-up` |
| `Dirty database version N` | A migration failed midway → fix SQL, `make migrate-force version=<last_good>`, re-run `migrate-up` |
| `make sqlc` errors on a column | Query references a column not in the schema, or schema not migrated → fix the SQL / add a migration |
| Swagger UI not reflecting changes | Forgot `make docs` (re-run; `make run` does it automatically) |
| `unable to ping postgres` on startup | DB not running or `.env` credentials wrong |
| Password appears `""` everywhere | Expected pre-auth; writing/reading password is deferred to the auth feature |
