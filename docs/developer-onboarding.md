# Developer Onboarding Guide

Tai lieu nay huong dan dev moi cach lam viec voi project Fracture, dac biet la:
- Viet API moi
- Them model moi
- Tao va quan ly migration

Muc tieu: follow dung workflow hien tai cua codebase (Clean Architecture + sqlc + golang-migrate + Swagger).

## 1. Tong quan kien truc

Project dang theo flow:
- `handler` nhan HTTP request/response (Gin)
- `usecase` xu ly business logic
- `repository` khai bao interface cho persistence
- `infrastructure/persistence` implement repository bang PostgreSQL + sqlc
- `domain` chua entity va domain errors

Duong di dependency:
- handler -> usecase -> repository interface -> repository implementation -> db/sqlc

## 2. Yeu cau moi truong

Can co san:
- Go (version theo `go.mod`)
- PostgreSQL (chay local hoac Docker)
- `migrate` CLI
- `sqlc` CLI
- `swag` CLI

Lenh setup nhanh:

```bash
make deps
make migrate-install
make sqlc-install
go install github.com/swaggo/swag/cmd/swag@latest
```

## 3. Chay project local

1. Tao/cap nhat file `.env` voi cac key can thiet:
- `APP_PORT`
- `APP_ENV`
- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `DB_SSLMODE`

2. Chay migration:

```bash
make migrate-up
```

3. Generate sqlc code (neu co doi query/schema):

```bash
make sqlc
```

4. Chay API:

```bash
make run
```

Swagger UI:
- `http://localhost:<APP_PORT>/swagger/`

Health check:
- `GET /health`

## 4. Quy trinh viet API moi (cho model da co)

Vi du: them endpoint `GET /api/v1/users/by-email/:email`.

### Buoc 1: Them query sqlc

Sua file query tuong ung trong `internal/infrastructure/persistence/queries/`.

Example (users):

```sql
-- name: GetUserByEmail :one
SELECT id, email, password, name, created_at, updated_at
FROM users
WHERE email = $1;
```

### Buoc 2: Regenerate sqlc

```bash
make sqlc
```

Sau buoc nay, cac function/type moi se duoc tao trong `internal/infrastructure/persistence/sqlcgen/`.

### Buoc 3: Cap nhat repository interface

Them method vao `internal/repository/user_repository.go` (hoac repo cua model tuong ung).

### Buoc 4: Implement trong persistence repo

Sua file implementation o `internal/infrastructure/persistence/`:
- Goi ham sqlc vua generate
- Map loi `pgx.ErrNoRows` -> `domain.ErrNotFound` neu can
- Map sqlc model -> domain model

### Buoc 5: Them usecase method

Sua file usecase trong `internal/usecase/`:
- Validate input
- Goi repository
- Xu ly business rules

### Buoc 6: Them handler + route

1. Them handler method trong `internal/handler/`
2. Dang ky route trong `cmd/api/main.go`
3. Tra ve HTTP status code phu hop (`400`, `404`, `409`, `500`, ...)

### Buoc 7: Swagger annotation

Them/doi comment `@Summary`, `@Param`, `@Success`, `@Failure`, `@Router` tren handler.

Regenerate docs:

```bash
make docs
```

### Buoc 8: Verify

- Chay app: `make run`
- Test endpoint moi bang curl/Postman
- Mo `/swagger/` check spec da update

## 5. Quy trinh them model moi (end-to-end)

Vi du them model `Article`.

### Buoc 1: Tao migration

```bash
make migrate-create name=create_articles_table
```

Se tao 2 file trong `migrations/`:
- `xxxxxx_create_articles_table.up.sql`
- `xxxxxx_create_articles_table.down.sql`

Viet SQL:
- `.up.sql`: tao table/index/constraint
- `.down.sql`: rollback nguoc lai (drop constraint/index/table)

Apply migration:

```bash
make migrate-up
```

### Buoc 2: Tao domain model

Them file trong `internal/domain/`, vi du `article.go`.

Chi de data structure + domain behavior, khong dua logic HTTP/DB vao day.

### Buoc 3: Tao repository interface

Them file interface trong `internal/repository/`, vi du `article_repository.go`.

Define cac method can cho usecase (CRUD + method dac thu).

### Buoc 4: Viet sqlc queries cho model moi

Them file SQL trong:
- `internal/infrastructure/persistence/queries/articles.sql`

Khai bao cac query voi ten ro rang:
- `CreateArticle`
- `GetArticleByID`
- `ListArticles`
- `UpdateArticle`
- `DeleteArticle`

Generate sqlc code:

```bash
make sqlc
```

### Buoc 5: Implement PostgreSQL repository

Tao file trong `internal/infrastructure/persistence/`, vi du `postgres_article_repo.go`.

Pattern nen theo:
- Struct repo giu `*sqlcgen.Queries`
- Compile-time check implement interface
- Function map sqlc type -> domain type
- Map `ErrNoRows` -> `domain.ErrNotFound`

### Buoc 6: Tao usecase

Tao `internal/usecase/article_usecase.go`:
- Inject repository interface
- Validate input
- Dat business rules
- Set metadata (ID, timestamps) neu can

### Buoc 7: Tao handler

Tao `internal/handler/article_handler.go`:
- Request DTO cho create/update
- Bind va validate JSON
- Goi usecase
- Mapping domain error -> HTTP status

### Buoc 8: Wire vao main

Trong `cmd/api/main.go`:
- Khoi tao repo + usecase + handler moi
- Register route group moi trong `/api/v1`

### Buoc 9: Swagger + verify

- Them Swagger annotations
- Chay `make docs`
- Chay `make run`
- Test endpoint

## 6. Quy trinh migration chi tiet

### Tao migration moi

```bash
make migrate-create name=<migration_name>
```

### Apply migration

```bash
make migrate-up
```

### Rollback 1 step

```bash
make migrate-down
```

### Xem version hien tai

```bash
make migrate-version
```

### Fix dirty state

```bash
make migrate-force version=<n>
```

Luu y quan trong:
- Luon viet day du ca file up/down
- Khong sua migration da merge tren branch chinh; tao migration moi de alter
- Kiem tra migration tren DB sach (fresh DB) truoc khi merge

## 7. Quy uoc code quan trong

- Domain errors dung chung trong `internal/domain/errors.go`
- Khong truy cap DB truc tiep tu usecase/handler
- Moi thay doi query/schema can `make sqlc`
- Moi thay doi swagger annotation can `make docs`
- Route API de duoi `/api/v1`

## 8. Checklist truoc khi mo PR

- Migration da co up/down hop le
- Da apply migration local thanh cong
- Da regenerate sqlc (`make sqlc`)
- Da regenerate swagger (`make docs`)
- API moi da duoc wire route trong `cmd/api/main.go`
- Handler tra status code dung theo domain error
- Chay app local thanh cong (`make run`)
- Manual test endpoint thanh cong

## 9. Ghi chu testing

Hien tai file test usecase con trong (`internal/usecase/user_usecase_test.go`).

Khuyen nghi khi them API/model moi:
- Unit test cho usecase (happy path + validation + domain error)
- Test conflict/not found behavior
- Neu co the, bo sung integration test cho repository layer
