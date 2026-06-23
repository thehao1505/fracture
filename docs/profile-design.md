# Thiết kế: User Profile (Link-in-bio, mô hình Block)

> Tài liệu thiết kế tính năng **Profile** — trang bio công khai của mỗi user
> (kiểu Linktree/Bento/Beacons). Nội dung trang dựng theo **block** (link chỉ là
> một loại block). Bám theo kiến trúc hiện có: clean architecture + `pgx` +
> `sqlc` + `golang-migrate`.

## 1. Bối cảnh & mục tiêu

App là một **link-in-bio**: mỗi user có một trang công khai tại
`https://app.com/@<username>`. Trang gồm phần header (avatar/bio) và một **chuỗi
block** xếp dọc, người dùng tự sắp xếp. Tính năng cần:

- Mỗi user sở hữu **đúng 1 profile** (quan hệ 1–1 với `users`).
- Profile có **nhiều block** (1–N), tự sắp xếp thứ tự, bật/tắt.
- `link` chỉ là **một** loại block; còn có `header`, `socials`, và (tương lai)
  `image`, `embed`, `text`, `divider`…
- Trang public **không cần đăng nhập** để xem; thao tác chỉnh sửa thì cần token.
- Tách bạch **dữ liệu public** (ai cũng xem) và **dữ liệu private** (chỉ chủ sở hữu).

Không thuộc phạm vi bản này (để sau): custom domain, A/B testing, thanh toán,
team/multi-profile.

**Vì sao chọn block thay vì "danh sách link phẳng":** link chỉ là một block type
nên chọn block không mất gì mà mở rộng được (thêm type không cần migration). Chi
phí chênh lệch nhỏ — bảng `blocks` gần như y hệt một bảng `links`, chỉ gộp các
trường riêng của link vào `content` JSONB và thêm cột `type`. Đổi từ link → block
về sau rất đau (migrate data + viết lại API), nên bắt đầu bằng block.

## 2. Mô hình dữ liệu

```
users (1) ───< (1) profiles (1) ───< (N) blocks
```

### 2.1. Entity `Profile` (1–1 với User)

| Cột           | Kiểu          | Ràng buộc                                  | Ghi chú |
|---------------|---------------|--------------------------------------------|---------|
| `id`          | UUID          | PK                                         | |
| `user_id`     | UUID          | FK → `users(id)`, `UNIQUE`, `ON DELETE CASCADE` | 1–1 |
| `username`    | TEXT          | `UNIQUE`, NOT NULL                          | slug công khai `@username` |
| `display_name`| TEXT          | NOT NULL DEFAULT `''`                       | tên hiển thị trên trang |
| `bio`         | TEXT          | NOT NULL DEFAULT `''`                       | mô tả ngắn |
| `avatar_url`  | TEXT          | NOT NULL DEFAULT `''`                       | |
| `appearance`  | JSONB         | NOT NULL DEFAULT `'{}'`                     | theme: màu nền, kiểu nút, font… |
| `is_published`| BOOLEAN       | NOT NULL DEFAULT `false`                    | nháp vs công khai |
| `created_at`  | TIMESTAMPTZ   | NOT NULL DEFAULT `now()`                    | |
| `updated_at`  | TIMESTAMPTZ   | NOT NULL DEFAULT `now()`                    | |

**`username`** là khóa public quan trọng nhất — quy tắc validate ở mục 5.

**`appearance` (JSONB)** chọn JSONB thay vì nhiều cột rời để theme tiến hóa tự
do mà không cần migration mỗi lần thêm tùy chọn. Đánh đổi: không query/validate
ở tầng DB — phải validate ở tầng `usecase`. Hình dạng đề xuất:

```json
{
  "theme": "dark",
  "background": { "type": "color", "value": "#1b1b1b" },
  "button": { "style": "rounded", "color": "#ffffff" },
  "font": "inter"
}
```

### 2.2. Entity `Block` (1–N từ Profile)

| Cột          | Kiểu        | Ràng buộc                                   | Ghi chú |
|--------------|-------------|---------------------------------------------|---------|
| `id`         | UUID        | PK                                          | |
| `profile_id` | UUID        | FK → `profiles(id)`, `ON DELETE CASCADE`    | |
| `type`       | TEXT        | NOT NULL                                    | `'link'` \| `'socials'` \| `'header'` \| … |
| `content`    | JSONB       | NOT NULL DEFAULT `'{}'`                      | payload tùy theo `type` |
| `position`   | INTEGER     | NOT NULL DEFAULT `0`                        | thứ tự sắp xếp (asc) |
| `is_active`  | BOOLEAN     | NOT NULL DEFAULT `true`                     | ẩn/hiện không cần xóa |
| `click_count`| BIGINT      | NOT NULL DEFAULT `0`                        | analytics cho block có thể click |
| `created_at` | TIMESTAMPTZ | NOT NULL DEFAULT `now()`                    | |
| `updated_at` | TIMESTAMPTZ | NOT NULL DEFAULT `now()`                    | |

**Toàn bộ sự khác biệt giữa các loại block nằm trong `content` (JSONB).** Cấu
trúc `content` cho từng `type` (v1 làm 3 type đầu, phần sau để mở rộng):

```jsonc
// type = "link"
{ "title": "My shop", "url": "https://...", "icon": "shopping-bag", "thumbnail": "" }

// type = "socials"   (một hàng icon mạng xã hội)
{ "items": [
    { "platform": "instagram", "url": "https://instagram.com/abc" },
    { "platform": "github",    "url": "https://github.com/abc" }
] }

// type = "header"    (tiêu đề ngăn cách các nhóm)
{ "text": "My socials" }

// ---- mở rộng tương lai ----
// type = "text"    → { "markdown": "..." }
// type = "image"   → { "url": "...", "alt": "...", "href": "" }
// type = "embed"   → { "provider": "youtube", "url": "..." }
// type = "divider" → {}
```

> `click_count` chỉ tăng cho các type "có thể click" (`link`, và mỗi item trong
> `socials` — xem mục 8 về cách xử lý click cho socials).

### 2.3. Domain models (Go) — `internal/domain/profile.go`

```go
type Profile struct {
    ID          uuid.UUID       `json:"id"`
    UserID      uuid.UUID       `json:"user_id"`
    Username    string          `json:"username"`
    DisplayName string          `json:"display_name"`
    Bio         string          `json:"bio"`
    AvatarURL   string          `json:"avatar_url"`
    Appearance  json.RawMessage `json:"appearance"`
    IsPublished bool            `json:"is_published"`
    CreatedAt   time.Time       `json:"created_at"`
    UpdatedAt   time.Time       `json:"updated_at"`
    Blocks      []Block         `json:"blocks,omitempty"` // nạp khi xem trang
}

// BlockType liệt kê các loại block hợp lệ.
type BlockType string

const (
    BlockTypeLink    BlockType = "link"
    BlockTypeSocials BlockType = "socials"
    BlockTypeHeader  BlockType = "header"
)

type Block struct {
    ID         uuid.UUID       `json:"id"`
    ProfileID  uuid.UUID       `json:"profile_id"`
    Type       BlockType       `json:"type"`
    Content    json.RawMessage `json:"content"` // hình dạng tùy Type
    Position   int32           `json:"position"`
    IsActive   bool            `json:"is_active"`
    ClickCount int64           `json:"click_count"`
    CreatedAt  time.Time       `json:"created_at"`
    UpdatedAt  time.Time       `json:"updated_at"`
}
```

`content` giữ ở dạng `json.RawMessage` xuyên suốt domain/repo — usecase mới là
nơi parse/validate theo `Type` (xem mục 5). Cách này giữ repo "ngu" và đơn giản,
logic theo từng type gom về một chỗ.

**Tách view public**: trang công khai KHÔNG nên trả `user_id`, `is_published`,
`click_count`. Dùng DTO riêng ở tầng handler:

```go
type PublicProfile struct {
    Username    string          `json:"username"`
    DisplayName string          `json:"display_name"`
    Bio         string          `json:"bio"`
    AvatarURL   string          `json:"avatar_url"`
    Appearance  json.RawMessage `json:"appearance"`
    Blocks      []PublicBlock   `json:"blocks"`
}

type PublicBlock struct {
    ID      uuid.UUID       `json:"id"`   // cần để gọi endpoint /click
    Type    string          `json:"type"`
    Content json.RawMessage `json:"content"`
}
```

## 3. Migrations

Tạo bằng `make migrate-create name=create_profiles_table` rồi điền:

**`000003_create_profiles_table.up.sql`**
```sql
CREATE TABLE IF NOT EXISTS profiles (
    id           UUID PRIMARY KEY,
    user_id      UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    username     TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL DEFAULT '',
    bio          TEXT NOT NULL DEFAULT '',
    avatar_url   TEXT NOT NULL DEFAULT '',
    appearance   JSONB NOT NULL DEFAULT '{}',
    is_published BOOLEAN NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Tra cứu public theo username là truy vấn nóng nhất.
-- (đã có UNIQUE nên Postgres tự tạo index, dòng này chỉ minh hoạ chủ đích.)
-- CREATE UNIQUE INDEX idx_profiles_username ON profiles(username);
```

**`000003_create_profiles_table.down.sql`**
```sql
DROP TABLE IF EXISTS profiles;
```

**`000004_create_blocks_table.up.sql`**
```sql
CREATE TABLE IF NOT EXISTS blocks (
    id          UUID PRIMARY KEY,
    profile_id  UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    type        TEXT NOT NULL,
    content     JSONB NOT NULL DEFAULT '{}',
    position    INTEGER NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    click_count BIGINT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Lấy block theo profile, sắp xếp sẵn theo position.
CREATE INDEX idx_blocks_profile_position ON blocks(profile_id, position);
```

**`000004_create_blocks_table.down.sql`**
```sql
DROP TABLE IF EXISTS blocks;
```

> Vì `sqlc.yaml` đọc schema từ `migrations/*.up.sql`, chỉ cần thêm 2 file up là
> sqlc tự nhận bảng mới. Sau đó chạy `make sqlc` để sinh code.
>
> Cột `type` cố ý để `TEXT` (validate ở app) thay vì `ENUM` Postgres — thêm type
> block mới không phải `ALTER TYPE`/migration, đúng tinh thần mô hình block.

## 4. sqlc queries — `internal/infrastructure/persistence/queries/`

**`profiles.sql`**
```sql
-- name: CreateProfile :exec
INSERT INTO profiles (id, user_id, username, display_name, bio, avatar_url, appearance, is_published, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: GetProfileByUserID :one
SELECT * FROM profiles WHERE user_id = $1;

-- name: GetPublishedProfileByUsername :one
SELECT * FROM profiles WHERE username = $1 AND is_published = true;

-- name: UpdateProfile :exec
UPDATE profiles
SET username = $2, display_name = $3, bio = $4, avatar_url = $5,
    appearance = $6, is_published = $7, updated_at = $8
WHERE id = $1;

-- name: ProfileUsernameExists :one
SELECT EXISTS (SELECT 1 FROM profiles WHERE username = $1 AND id <> $2);
```

**`blocks.sql`**
```sql
-- name: CreateBlock :exec
INSERT INTO blocks (id, profile_id, type, content, position, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: ListBlocksByProfile :many
SELECT * FROM blocks WHERE profile_id = $1 ORDER BY position ASC, created_at ASC;

-- name: ListActiveBlocksByProfile :many
SELECT * FROM blocks WHERE profile_id = $1 AND is_active = true ORDER BY position ASC, created_at ASC;

-- name: UpdateBlock :exec
UPDATE blocks
SET type = $2, content = $3, is_active = $4, updated_at = $5
WHERE id = $1 AND profile_id = $6;

-- name: UpdateBlockPosition :exec
UPDATE blocks SET position = $2, updated_at = $3 WHERE id = $1 AND profile_id = $4;

-- name: DeleteBlock :exec
DELETE FROM blocks WHERE id = $1 AND profile_id = $2;

-- name: IncrementBlockClick :exec
UPDATE blocks SET click_count = click_count + 1 WHERE id = $1;
```

> Lưu ý: các câu UPDATE/DELETE đều kèm `profile_id` trong `WHERE` → đảm bảo
> user chỉ đụng được block thuộc profile của mình (chống IDOR ngay tầng query).

## 5. Validation (tầng usecase)

**Username** (quan trọng nhất, vì là URL public):
- Độ dài 3–30 ký tự.
- Chỉ `[a-z0-9_.]`, **luôn lưu lowercase** (normalize trước khi ghi).
- Không bắt đầu/kết thúc bằng `.`; không có `..`.
- Danh sách **reserved** không cho dùng: `admin`, `api`, `app`, `login`,
  `register`, `health`, `swagger`, `me`, `p`, `www`, `support`, `about`…
- Kiểm tra trùng qua `ProfileUsernameExists` (loại trừ chính profile đang sửa).

**Block — validate `content` theo `type`** (đây là điểm cốt lõi của mô hình
block; làm trong usecase trước khi ghi DB):

| `type`    | Quy tắc `content` |
|-----------|-------------------|
| `link`    | `title` bắt buộc ≤ 80 ký tự; `url` bắt buộc, parse được và scheme ∈ {`http`,`https`}; `icon`/`thumbnail` tùy chọn |
| `socials` | `items` 1–20 phần tử; mỗi item `platform` ∈ whitelist, `url` http/https |
| `header`  | `text` bắt buộc ≤ 80 ký tự |

Nguyên tắc chung:
- `type` phải ∈ tập `BlockType` đã định nghĩa; type lạ → `ErrBadRequest`.
- Mọi URL chỉ cho `http`/`https` (chặn `javascript:`, `data:` → tránh XSS khi
  client render).
- Từ chối field thừa / JSON sai hình dạng thay vì lưu tùy tiện
  (`json.Decoder` + `DisallowUnknownFields`).
- Giới hạn số block/profile (vd. tối đa 100) để chống lạm dụng.

Gợi ý tổ chức: một hàm `validateBlockContent(t BlockType, raw json.RawMessage) error`
dùng `switch` theo type — thêm type mới chỉ thêm một nhánh `case`.

**appearance**: validate theo schema cố định ở usecase (enum theme, mã màu hợp lệ
`#rrggbb`), từ chối JSON lạ.

## 6. API endpoints

### Public (không cần token)
| Method | Path | Mô tả |
|--------|------|-------|
| `GET`  | `/api/v1/p/:username` | Lấy `PublicProfile` + block active (chỉ khi `is_published`). 404 nếu không thấy/chưa publish. |
| `POST` | `/api/v1/p/:username/blocks/:id/click` | Ghi nhận 1 click rồi trả `{ "url": "..." }` để client redirect. |

### Owner (cần `AuthRequired`, thao tác trên profile của **chính mình**)
| Method | Path | Mô tả |
|--------|------|-------|
| `GET`    | `/api/v1/me/profile` | Lấy profile đầy đủ của tôi (kể cả nháp) |
| `POST`   | `/api/v1/me/profile` | Tạo profile lần đầu (1 user chỉ 1 profile) |
| `PUT`    | `/api/v1/me/profile` | Cập nhật username/bio/avatar/appearance/publish |
| `GET`    | `/api/v1/me/blocks` | Liệt kê tất cả block (cả ẩn) |
| `POST`   | `/api/v1/me/blocks` | Thêm block (`{ "type": "...", "content": {...} }`) |
| `PUT`    | `/api/v1/me/blocks/:id` | Sửa block |
| `DELETE` | `/api/v1/me/blocks/:id` | Xóa block |
| `PATCH`  | `/api/v1/me/blocks/reorder` | Sắp xếp lại: body `{ "order": ["<id1>", "<id2>", ...] }` |

> Định danh chủ sở hữu lấy từ JWT (`ContextUserID` mà `AuthRequired` đã set),
> **không** nhận `user_id` từ client → tránh IDOR. Route dùng `/me/...` cố ý để
> không bao giờ truyền id user khác.

### Wiring trong `main.go`
```go
public := v1.Group("/p")
{
    public.GET("/:username", profileH.GetPublic)
    public.POST("/:username/blocks/:id/click", profileH.RecordClick)
}

me := v1.Group("/me")
me.Use(middleware.AuthRequired(tokenManager))
{
    me.GET("/profile", profileH.GetMine)
    me.POST("/profile", profileH.Create)
    me.PUT("/profile", profileH.Update)
    me.GET("/blocks", blockH.List)
    me.POST("/blocks", blockH.Create)
    me.PUT("/blocks/:id", blockH.Update)
    me.DELETE("/blocks/:id", blockH.Delete)
    me.PATCH("/blocks/reorder", blockH.Reorder)
}
```

## 7. Các file cần thêm (theo layout hiện tại)

```
migrations/000003_create_profiles_table.{up,down}.sql
migrations/000004_create_blocks_table.{up,down}.sql
internal/infrastructure/persistence/queries/profiles.sql
internal/infrastructure/persistence/queries/blocks.sql
internal/domain/profile.go            # Profile, Block, BlockType + lỗi domain mới
internal/repository/profile_repository.go   # interface ProfileRepository, BlockRepository
internal/infrastructure/persistence/postgres_profile_repo.go
internal/usecase/profile_usecase.go   # validate username/appearance + validateBlockContent
internal/handler/profile_handler.go   # GetPublic, GetMine, Create, Update, RecordClick
internal/handler/block_handler.go     # CRUD + Reorder
```

(Chạy `make sqlc` sau khi thêm migration + query để sinh `sqlcgen`.)

## 8. Quyết định thiết kế & đánh đổi

1. **Mô hình block thay vì danh sách link phẳng**: `link` là một block type, nên
   không mất gì mà mở rộng được sang header/ảnh/embed mà **không cần migration**.
   Đổi từ link → block về sau rất đau, nên làm đúng từ đầu.
2. **`content` JSONB + `type` TEXT**: dữ liệu đa hình gọn trong một bảng. Đánh
   đổi: không có type-safety ở tầng DB → phải validate `content` theo `type` ở
   usecase (`validateBlockContent`). `type` để `TEXT` (không `ENUM`) để thêm
   loại mới khỏi `ALTER TYPE`.
3. **1–1 user↔profile, tách bảng** thay vì nhồi cột vào `users`: giữ `users`
   gọn cho auth, profile tiến hóa độc lập, mở đường multi-profile sau (bỏ
   `UNIQUE` trên `user_id`).
4. **`appearance` JSONB**: linh hoạt theme, không migration mỗi lần đổi UI.
5. **`click_count` ngay trên `blocks`**: analytics "đủ xài" cho v1. Cần số liệu
   theo thời gian/nguồn → tách bảng `block_clicks` (event) sau, không phá v1.
   *Lưu ý:* với block `socials` (nhiều item) thì `click_count` ở mức block là
   tổng — muốn tách theo từng nền tảng thì cần `block_clicks` ghi kèm khóa item.
6. **Reorder bằng cột `position` + endpoint riêng**: đổi thứ tự là thao tác phổ
   biến, làm batch trong 1 transaction thay vì sửa từng block.
7. **Route `/me/...` + `WHERE profile_id` trong query**: chặn IDOR hai lớp —
   đúng khuyến nghị bảo mật ở đợt rà soát production trước.

## 9. Mở rộng tương lai (ngoài v1)
- Thêm block type: `text` (markdown), `image`, `embed` (YouTube/Spotify), `divider`.
  Mỗi type mới = 1 nhánh trong `validateBlockContent` + 1 renderer ở client,
  **không cần migration**.
- Bảng `block_clicks` cho analytics theo thời gian, referrer, thiết bị, item.
- Custom domain, verified badge.
- Lên lịch hiện/ẩn block (start/end time).
- Theme template dựng sẵn + upload ảnh nền.
- Rate-limit endpoint `click` để chống bơm số liệu ảo.
