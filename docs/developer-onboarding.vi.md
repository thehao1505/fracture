# Hướng Dẫn Onboarding Cho Developer

Tài liệu này giúp developer mới làm việc hiệu quả với dự án Fracture, đặc biệt trong các việc sau:

- Viết API mới
- Thêm model mới
- Tạo và quản lý migration

Mục tiêu: tuân thủ đúng workflow hiện tại của codebase (Clean Architecture + sqlc + golang-migrate + Swagger).

## 1. Tổng Quan Kiến Trúc

Luồng hiện tại:

- `handler` nhận và trả HTTP request/response (Gin)
- `usecase` chứa business logic
- `repository` khai báo interface cho persistence
- `infrastructure/persistence` implement repository bằng PostgreSQL + sqlc
- `domain` chứa entity và domain errors

Chiều phụ thuộc:

- handler -> usecase -> repository interface -> repository implementation -> db/sqlc

## 2. Yêu Cầu Môi Trường

Cần có:

- Go (phiên bản theo `go.mod`)
- PostgreSQL (chạy local hoặc Docker)
- `migrate` CLI
- `sqlc` CLI
- `swag` CLI

Thiết lập nhanh:

```bash
make deps
make migrate-install
make sqlc-install
go install github.com/swaggo/swag/cmd/swag@latest
```

## 3. Chạy Dự Án Local

1. Tạo/cập nhật file `.env` với các key bắt buộc:

- `APP_PORT`
- `APP_ENV`
- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `DB_SSLMODE`

2. Chạy migration:

```bash
make migrate-up
```

3. Generate sqlc code (nếu schema/query thay đổi):

```bash
make sqlc
```

4. Chạy API:

```bash
make run
```

Swagger UI:

- `http://localhost:<APP_PORT>/swagger/`

Health check:

- `GET /health`

## 4. Quy Trình Thêm API Mới (Cho Model Đã Có)

Ví dụ: thêm endpoint `GET /api/v1/users/by-email/:email`.

### Bước 1: Thêm sqlc query

Cập nhật file query tương ứng trong `internal/infrastructure/persistence/queries/`.

Ví dụ (users):

```sql
-- name: GetUserByEmail :one
SELECT id, email, password, name, created_at, updated_at
FROM users
WHERE email = $1;
```

### Bước 2: Regenerate sqlc

```bash
make sqlc
```

Sau bước này, các type/function mới sẽ được generate trong `internal/infrastructure/persistence/sqlcgen/`.

### Bước 3: Cập nhật repository interface

Thêm method trong `internal/repository/user_repository.go` (hoặc repository tương ứng với model).

### Bước 4: Implement trong persistence repository

Cập nhật implementation ở `internal/infrastructure/persistence/`:

- Gọi các method sqlc vừa generate
- Map `pgx.ErrNoRows` -> `domain.ErrNotFound` khi phù hợp
- Map sqlc model -> domain model

### Bước 5: Thêm usecase method

Cập nhật usecase trong `internal/usecase/`:

- Validate input
- Gọi repository
- Áp dụng business rules

### Bước 6: Thêm handler + route

1. Thêm handler method trong `internal/handler/`
2. Đăng ký route trong `cmd/api/main.go`
3. Trả về HTTP status code phù hợp (`400`, `404`, `409`, `500`, ...)

### Bước 7: Swagger annotations

Thêm/cập nhật annotation trên handler: `@Summary`, `@Param`, `@Success`, `@Failure`, `@Router`.

Regenerate docs:

```bash
make docs
```

### Bước 8: Verify

- Chạy app: `make run`
- Test endpoint mới bằng curl/Postman
- Mở `/swagger/` và xác nhận spec đã cập nhật

## 5. Quy Trình Thêm Model Mới (End-To-End)

Ví dụ: thêm model `Article`.

### Bước 1: Tạo migration

```bash
make migrate-create name=create_articles_table
```

Lệnh sẽ tạo 2 file trong `migrations/`:

- `xxxxxx_create_articles_table.up.sql`
- `xxxxxx_create_articles_table.down.sql`

Viết SQL:

- `.up.sql`: tạo table/index/constraint
- `.down.sql`: rollback ngược lại (drop constraint/index/table)

Apply migration:

```bash
make migrate-up
```

### Bước 2: Thêm domain model

Tạo file trong `internal/domain/`, ví dụ `article.go`.

Chỉ giữ data structure và domain behavior. Không đưa logic HTTP/DB vào domain.

### Bước 3: Thêm repository interface

Tạo file interface trong `internal/repository/`, ví dụ `article_repository.go`.

Định nghĩa các method cần cho usecase (CRUD + method đặc thù).

### Bước 4: Viết sqlc queries cho model mới

Tạo file SQL trong:

- `internal/infrastructure/persistence/queries/articles.sql`

Đặt tên query rõ ràng:

- `CreateArticle`
- `GetArticleByID`
- `ListArticles`
- `UpdateArticle`
- `DeleteArticle`

Generate sqlc code:

```bash
make sqlc
```

### Bước 5: Implement PostgreSQL repository

Tạo file trong `internal/infrastructure/persistence/`, ví dụ `postgres_article_repo.go`.

Pattern nên theo:

- Repository struct giữ `*sqlcgen.Queries`
- Compile-time check interface implementation
- Hàm mapping sqlc type -> domain type
- Map `ErrNoRows` -> `domain.ErrNotFound`

### Bước 6: Thêm usecase

Tạo `internal/usecase/article_usecase.go`:

- Inject repository interface
- Validate input
- Áp dụng business rules
- Set metadata (ID, timestamps) khi cần

### Bước 7: Thêm handler

Tạo `internal/handler/article_handler.go`:

- Request DTO cho create/update
- Bind và validate JSON
- Gọi usecase
- Map domain errors -> HTTP status

### Bước 8: Wire vào main

Trong `cmd/api/main.go`:

- Khởi tạo repo + usecase + handler mới
- Register route group mới dưới `/api/v1`

### Bước 9: Swagger + verify

- Thêm Swagger annotations
- Chạy `make docs`
- Chạy `make run`
- Test endpoints

## 6. Migration Workflow (Chi Tiết)

### Tạo migration mới

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

### Xem version hiện tại

```bash
make migrate-version
```

### Sửa trạng thái dirty

```bash
make migrate-force version=<n>
```

Lưu ý quan trọng:

- Luôn viết đầy đủ cả file up/down
- Không sửa migration đã merge vào main; hãy tạo migration mới cho thay đổi tiếp theo
- Verify migration trên DB sạch (fresh database) trước khi merge

## 7. Quy Ước Code Quan Trọng

- Shared domain errors được định nghĩa trong `internal/domain/errors.go`
- Không truy cập DB trực tiếp từ usecase/handler
- Mọi thay đổi schema/query đều cần chạy `make sqlc`
- Mọi thay đổi Swagger annotation đều cần chạy `make docs`
- API routes nên nằm dưới `/api/v1`

## 8. Checklist Trước Khi Mở PR

- Migration có đủ file up/down hợp lệ
- Migration apply local thành công
- sqlc đã regenerate (`make sqlc`)
- Swagger docs đã regenerate (`make docs`)
- API mới đã được wire trong `cmd/api/main.go`
- Handler trả status code đúng theo domain errors
- App chạy local thành công (`make run`)
- Manual test endpoint đã pass

## 9. Ghi Chú Testing

Hiện tại file test usecase vẫn đang trống (`internal/usecase/user_usecase_test.go`).

Khuyến nghị khi thêm API/model mới:

- Viết usecase unit tests (happy path + validation + domain errors)
- Test conflict/not-found behavior
- Bổ sung repository integration tests nếu có thể
