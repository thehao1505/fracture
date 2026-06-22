# Hướng dẫn: Thêm API `GET /users` (list + pagination + search)

Tài liệu này hướng dẫn thêm endpoint **`GET /users`** cho phép:

- Phân trang (pagination) bằng `page` / `limit`.
- Tìm kiếm theo `keyword` trên **name** và **email** (không phân biệt hoa/thường).
- Sắp xếp (`sort_by`) theo `created_at` / `updated_at` / `name` / `email`, chiều `asc` / `desc`. Mặc định **`created_at desc`**.

Đi từ dưới lên theo đúng clean architecture của project:

```
HTTP query  →  Handler (parse param)
            →  UseCase (default/cap/offset + whitelist sort, business rules)
            →  Repository.List + Count
            →  sqlc ListUsers/CountUsers (ILIKE + ORDER BY + LIMIT/OFFSET)
```

> ⚠️ **Quan trọng về sort:** sqlc (và SQL nói chung) chỉ bind được *giá trị*, **không bind được tên cột hay hướng `ASC/DESC`**.
> Vì vậy không thể viết `ORDER BY $1 $2`. Có 2 cách:
> 1. **Dùng `CASE` trong `ORDER BY`** (cách dùng ở đây) — giữ 100% sqlc, tham số hóa hoàn toàn, **chống SQL injection tuyệt đối**.
> 2. Tự nối chuỗi `ORDER BY` trong Go rồi chạy query thủ công — linh hoạt hơn nhưng **bắt buộc whitelist** cột/hướng, dễ sai sót.
> Tài liệu này dùng cách 1.

Thứ tự làm an toàn nhất: **1 → 2 → 3 → 4 → 5 → 6 → 7 → 8**. Mỗi bước xong chạy `go build ./...` để bắt lỗi sớm.

---

## 1. SQL query (sqlc)

File: `internal/infrastructure/persistence/queries/users.sql`

Thêm 2 query. Dùng **named params** (`@ten`) thay vì `$1` để sqlc tự gộp `keyword` lặp lại thành 1 field:

```sql
-- name: ListUsers :many
SELECT id, email, password, name, created_at, updated_at
FROM users
WHERE
    @keyword::text = ''
    OR name  ILIKE '%' || @keyword || '%'
    OR email ILIKE '%' || @keyword || '%'
ORDER BY
    CASE WHEN @sort_by::text = 'name'       AND @sort_order::text = 'asc'  THEN name       END ASC,
    CASE WHEN @sort_by::text = 'name'       AND @sort_order::text = 'desc' THEN name       END DESC,
    CASE WHEN @sort_by::text = 'email'      AND @sort_order::text = 'asc'  THEN email      END ASC,
    CASE WHEN @sort_by::text = 'email'      AND @sort_order::text = 'desc' THEN email      END DESC,
    CASE WHEN @sort_by::text = 'updated_at' AND @sort_order::text = 'asc'  THEN updated_at END ASC,
    CASE WHEN @sort_by::text = 'updated_at' AND @sort_order::text = 'desc' THEN updated_at END DESC,
    CASE WHEN @sort_by::text = 'created_at' AND @sort_order::text = 'asc'  THEN created_at END ASC,
    created_at DESC
LIMIT @page_limit OFFSET @page_offset;

-- name: CountUsers :one
SELECT COUNT(*)
FROM users
WHERE
    @keyword::text = ''
    OR name  ILIKE '%' || @keyword || '%'
    OR email ILIKE '%' || @keyword || '%';
```

- `ILIKE` = search không phân biệt hoa/thường.
- `@keyword = ''` → khi keyword rỗng thì lấy tất cả.
- `CountUsers` cần thiết để trả `total` cho client tính số trang. **Không cần `ORDER BY`** vì chỉ đếm.
- `::text` cast giúp sqlc suy ra đúng kiểu Go là `string`.

**Giải thích khối `ORDER BY`:**

- Mỗi cột × mỗi chiều = 1 nhánh `CASE`. Khi `sort_by`/`sort_order` khớp, `CASE` trả về giá trị cột đó → Postgres sort theo nhánh đó; các nhánh còn lại trả `NULL` nên không ảnh hưởng.
- Phải tách riêng `ASC` và `DESC` cho từng cột vì hướng sort là cú pháp cố định, không nhét vào `CASE` được.
- Mỗi `CASE` chỉ chứa **một kiểu dữ liệu** (text với text, timestamp với timestamp) — không trộn `name` (text) với `created_at` (timestamp) trong cùng một `CASE`, nếu không Postgres báo lỗi kiểu.
- Dòng cuối `created_at DESC` đóng **2 vai trò**: (1) giá trị **mặc định** khi không nhánh nào khớp — đây chính là `created_at desc`; (2) **tie-breaker** để thứ tự ổn định khi cột sort có giá trị trùng. (Vì vậy không cần nhánh `created_at` + `desc` riêng.)

> 💡 Sau này muốn nhanh hơn có thể thêm index `pg_trgm`
> (`CREATE EXTENSION pg_trgm; CREATE INDEX ... USING gin (name gin_trgm_ops)`)
> trong một migration mới — không bắt buộc cho bước đầu.

## 2. Sinh lại code sqlc

```bash
make sqlc
```

Sẽ tạo `r.q.ListUsers(ctx, ListUsersParams{...})` và `r.q.CountUsers(ctx, ...)` trong `sqlcgen/`.
`ListUsersParams` giờ gồm: `Keyword string`, `SortBy string`, `SortOrder string`, `PageLimit int32`, `PageOffset int32`.

> Kiểm tra tên field chính xác trong `internal/infrastructure/persistence/sqlcgen/users.sql.go` sau khi generate.

## 3. Repository interface

File: `internal/repository/user_repository.go`

Thêm 2 method vào interface `UserRepository`:

```go
List(ctx context.Context, keyword, sortBy, sortOrder string, limit, offset int32) ([]*domain.User, error)
Count(ctx context.Context, keyword string) (int64, error)
```

## 4. Repository implementation

File: `internal/infrastructure/persistence/postgres_user_repo.go`

Implement, tái dùng `userToDomain` sẵn có:

```go
func (r *postgresUserRepo) List(ctx context.Context, keyword, sortBy, sortOrder string, limit, offset int32) ([]*domain.User, error) {
	rows, err := r.q.ListUsers(ctx, sqlcgen.ListUsersParams{
		Keyword:    keyword,
		SortBy:     sortBy,
		SortOrder:  sortOrder,
		PageLimit:  limit,
		PageOffset: offset,
	})
	if err != nil {
		return nil, err
	}
	users := make([]*domain.User, 0, len(rows))
	for _, u := range rows {
		users = append(users, userToDomain(u))
	}
	return users, nil
}

func (r *postgresUserRepo) Count(ctx context.Context, keyword string) (int64, error) {
	return r.q.CountUsers(ctx, keyword)
}
```

> Lưu ý tên field trong `ListUsersParams` đúng như sqlc sinh ra (xem lại file `users.sql.go`).

## 5. UseCase

File: `internal/usecase/user_usecase.go`

Đây là nơi đặt **logic phân trang & default** (không để rò xuống handler hay repo):

```go
// Whitelist các cột được phép sort. Dù SQL đã an toàn nhờ CASE,
// việc whitelist ở đây giúp trả về giá trị "sạch" và rõ ràng cho client.
var allowedSortFields = map[string]bool{
	"created_at": true,
	"updated_at": true,
	"name":       true,
	"email":      true,
}

type ListUsersParams struct {
	Page      int    // bắt đầu từ 1
	Limit     int    // số item mỗi trang
	Keyword   string
	SortBy    string // created_at | updated_at | name | email
	SortOrder string // asc | desc
}

func (uc *UserUseCase) ListUsers(ctx context.Context, p ListUsersParams) ([]*domain.User, int64, error) {
	// Chuẩn hóa & chặn giá trị bất thường
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 20 // default
	}
	if p.Limit > 100 {
		p.Limit = 100 // cap, tránh client xin 1 triệu record
	}

	// Default + whitelist sort. Giá trị không hợp lệ -> quay về mặc định.
	if !allowedSortFields[p.SortBy] {
		p.SortBy = "created_at"
	}
	if p.SortOrder != "asc" && p.SortOrder != "desc" {
		p.SortOrder = "desc"
	}

	offset := (p.Page - 1) * p.Limit

	users, err := uc.userRepo.List(ctx, p.Keyword, p.SortBy, p.SortOrder, int32(p.Limit), int32(offset))
	if err != nil {
		return nil, 0, err
	}

	total, err := uc.userRepo.Count(ctx, p.Keyword)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
```

> Nhờ default ở đây, khi client không truyền (hoặc truyền sai) `sort_by`/`sort_order` thì luôn về **`created_at` + `desc`** — đúng yêu cầu mặc định.

## 6. Handler

File: `internal/handler/user_handler.go`

Đọc query string, gọi usecase, trả data kèm metadata phân trang.
Có `@Security BearerAuth` để hiện nút Authorize trong Swagger:

```go
// ListUsers godoc
// @Summary List users
// @Description List users with pagination and optional keyword search on name/email
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 20, max 100)"
// @Param keyword query string false "Search keyword (matches name or email)"
// @Param sort_by query string false "Sort field" Enums(created_at, updated_at, name, email) default(created_at)
// @Param sort_order query string false "Sort direction" Enums(asc, desc) default(desc)
// @Success 200 {object} map[string]interface{} "Users retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))   // "" -> 0, usecase tự set default
	limit, _ := strconv.Atoi(c.Query("limit"))

	users, total, err := h.userUc.ListUsers(c.Request.Context(), usecase.ListUsersParams{
		Page:      page,
		Limit:     limit,
		Keyword:   c.Query("keyword"),
		SortBy:    c.Query("sort_by"),    // "" -> usecase default created_at
		SortOrder: c.Query("sort_order"), // "" -> usecase default desc
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}
```

> `Enums(...)` và `default(...)` trong annotation giúp Swagger UI hiện sẵn dropdown các giá trị hợp lệ.

> Nhớ thêm `"strconv"` vào khối import.

## 7. Đăng ký route

File: `cmd/api/main.go`

Trong nhóm `users` (đã có `AuthRequired`), thêm:

```go
users.GET("", userH.ListUsers)
```

> ⚠️ Hiện `users.POST("", userH.CreateUser)` đang dùng path rỗng. `GET ""` và `POST ""`
> cùng path nhưng khác method nên **không xung đột** — Gin phân biệt theo HTTP method.

## 8. Sinh lại Swagger & test

```bash
make docs          # swag init -> cập nhật docs/
make run           # hoặc cách bạn vẫn chạy server
```

Gọi thử:

```
# Mặc định: created_at desc
GET /api/v1/users?page=1&limit=10&keyword=hao

# Sort theo tên A→Z
GET /api/v1/users?sort_by=name&sort_order=asc

# Sort theo email Z→A
GET /api/v1/users?sort_by=email&sort_order=desc

Authorization: Bearer <token>
```

Response mẫu:

```json
{
  "data": [
    { "id": "...", "email": "hao@yopmail.com", "name": "Hao", "created_at": "...", "updated_at": "..." }
  ],
  "pagination": { "page": 1, "limit": 10, "total": 42 }
}
```

---

## Mở rộng (tùy chọn)

- **Index tìm kiếm nhanh hơn**: tạo migration mới bật `pg_trgm` + GIN index trên `name`, `email`.
- **Cursor-based pagination**: thay `OFFSET` bằng điều kiện `WHERE created_at < @cursor` để ổn định khi data thay đổi và nhanh hơn ở trang sâu.
- **Index cho cột sort**: nếu sort theo `name`/`email` nhiều, cân nhắc thêm B-tree index trên các cột đó để tránh sort trên toàn bảng.
