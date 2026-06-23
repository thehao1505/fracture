# Hướng dẫn triển khai: User Profile (phần còn lại)

> Tài liệu **làm tiếp** từ [`profile-design.md`](./profile-design.md). Bám đúng
> kiến trúc & convention hiện có của repo (clean architecture + `pgx` + `sqlc` +
> `gin`). Tài liệu này chỉ **hướng dẫn**; bạn tự viết code theo từng bước.

## 0. Trạng thái hiện tại

Đã có sẵn (không cần làm lại):

- ✅ Migrations `000003_*`, `000004_*`.
- ✅ Query SQL: `queries/profile.sql`, `queries/block.sql`.
- ✅ Code sqlc sinh ra: `sqlcgen/profile.sql.go`, `sqlcgen/block.sql.go`
  (đã có `CreateProfile`, `GetProfileByUserID`, `GetPublishedProfileByUsername`,
  `ProfileUsernameExists`, `UpdateProfile`, `CreateBlock`, `ListBlocksByProfile`,
  `ListActiveBlocksByProfile`, `UpdateBlock`, `UpdateBlockPosition`,
  `DeleteBlock`, `IncrementBlockClick`).
- ✅ Domain struct `Profile` (`domain/profile.go`) và `Block` + `BlockType`
  (`domain/block.go`).
- 🟡 `usecase/profile_usecase.go` — mới có `type ProfileUseCase struct{}` rỗng.
- 🟡 `persistence/postgres_profile_repo.go` — mới có dòng `package persistence`.

Còn thiếu (nội dung tài liệu này):

1. DTO public + (tùy chọn) lỗi domain mới.
2. Interface `ProfileRepository`, `BlockRepository`.
3. Implementation Postgres cho 2 interface trên.
4. Logic usecase: validate username / appearance / `content` theo type + các
   thao tác CRUD/reorder/click.
5. Handler `profile_handler.go`, `block_handler.go`.
6. Wiring trong `cmd/api/main.go`.
7. Lệnh build / migrate / test.

**Thứ tự làm khuyến nghị:** domain → repository (interface) → persistence (impl)
→ usecase → handler → wiring. Mỗi lớp chỉ phụ thuộc lớp dưới, build dần cho chắc.

> ⚠️ Lưu ý format: `domain/profile.go` và `block.go` đang lẫn **space** thụt đầu
> dòng (gofmt dùng **tab**). Chạy `gofmt -w internal/domain/*.go` trước khi đi
> tiếp để tránh lệch chuẩn cả module.

---

## 1. Domain — bổ sung DTO public (và lỗi nếu cần)

Struct `Profile`/`Block` đã đủ. Cần thêm **DTO view public** để trang công khai
KHÔNG lộ `user_id`, `is_published`, `click_count` (xem mục 2.3 của design).

Tạo trong `internal/domain/profile.go` (hoặc file `public_profile.go` cùng package):

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
    ID      uuid.UUID       `json:"id"`   // cần để gọi /click
    Type    string          `json:"type"`
    Content json.RawMessage `json:"content"`
}
```

**Lỗi domain:** các lỗi hiện có (`ErrNotFound`, `ErrBadRequest`, `ErrConflict`,
`ErrInvalidID`, `ErrUnauthorized`) đã đủ cho luồng profile. Map như sau:

- username sai định dạng / reserved / `content` sai → `ErrBadRequest`.
- username đã có người dùng → `ErrConflict`.
- tạo profile lần 2 cho cùng user → `ErrConflict`.
- không tìm thấy profile/block → `ErrNotFound`.

Nếu muốn message rõ hơn cho client (vd. "username taken" khác "profile exists"),
có thể thêm lỗi mới vào `domain/errors.go`, nhưng **không bắt buộc** cho v1.

---

## 2. Repository — interface

Tạo `internal/repository/profile_repository.go` (đặt cạnh `user_repository.go`,
cùng `package repository`). Hai interface, theo đúng style `UserRepository`:

```go
type ProfileRepository interface {
    Create(ctx context.Context, p *domain.Profile) error
    GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Profile, error)
    GetPublishedByUsername(ctx context.Context, username string) (*domain.Profile, error)
    Update(ctx context.Context, p *domain.Profile) error
    // exclude là id profile đang sửa (lúc tạo mới truyền uuid.Nil).
    UsernameExists(ctx context.Context, username string, exclude uuid.UUID) (bool, error)
}

type BlockRepository interface {
    Create(ctx context.Context, b *domain.Block) error
    ListByProfile(ctx context.Context, profileID uuid.UUID) ([]domain.Block, error)
    ListActiveByProfile(ctx context.Context, profileID uuid.UUID) ([]domain.Block, error)
    Update(ctx context.Context, b *domain.Block) error
    Delete(ctx context.Context, id, profileID uuid.UUID) error
    IncrementClick(ctx context.Context, id uuid.UUID) error
    // Reorder cập nhật position cho nhiều block trong 1 transaction.
    Reorder(ctx context.Context, profileID uuid.UUID, orderedIDs []uuid.UUID) error
}
```

> Ghi chú thiết kế interface:
> - `UsernameExists` gói luôn việc loại trừ chính mình (khớp query
>   `ProfileUsernameExists` nhận `id`). Lúc tạo mới chưa có id → truyền `uuid.Nil`.
> - `Reorder` để ở interface (không phải usecase loop từng cái) vì cần
>   **transaction** — gom về repo cho gọn, usecase chỉ truyền thứ tự id.
> - Trả `[]domain.Block` (value) cho đồng nhất với cách `Profile.Blocks` dùng.

---

## 3. Persistence — implementation Postgres

File `internal/infrastructure/persistence/postgres_profile_repo.go`. Theo y hệt
khung `postgres_user_repo.go`. **Khác biệt quan trọng:** repo này cần cả `pool`
(để mở transaction cho `Reorder`), không chỉ `*sqlcgen.Queries`.

### 3.1. Khung struct + constructor

```go
type postgresProfileRepo struct {
    q    *sqlcgen.Queries
    pool *pgxpool.Pool   // cần cho transaction (Reorder)
}

var _ repository.ProfileRepository = (*postgresProfileRepo)(nil)

func NewPostgresProfileRepo(db *pgxpool.Pool) repository.ProfileRepository {
    return &postgresProfileRepo{q: sqlcgen.New(db), pool: db}
}
```

Làm tương tự `postgresBlockRepo` (cùng file cũng được, hoặc tách
`postgres_block_repo.go`). Block repo cũng cần `pool` cho `Reorder`.

### 3.2. Hàm map sqlc row → domain

Lưu ý kiểu: cột JSONB sqlc sinh ra là `[]byte`; domain dùng `json.RawMessage`
(bản chất cũng là `[]byte`) → convert trực tiếp `json.RawMessage(u.Appearance)`.
`Type` ở sqlc là `string`, domain là `BlockType` → ép `domain.BlockType(b.Type)`.

```go
func profileToDomain(p sqlcgen.Profile) *domain.Profile {
    return &domain.Profile{
        ID:          p.ID,
        UserID:      p.UserID,
        Username:    p.Username,
        DisplayName: p.DisplayName,
        Bio:         p.Bio,
        AvatarURL:   p.AvatarUrl,                 // sqlc đặt tên AvatarUrl
        Appearance:  json.RawMessage(p.Appearance),
        IsPublished: p.IsPublished,
        CreatedAt:   p.CreatedAt,
        UpdatedAt:   p.UpdatedAt,
    }
}

func blockToDomain(b sqlcgen.Block) domain.Block {
    return domain.Block{
        ID:         b.ID,
        ProfileID:  b.ProfileID,
        Type:       domain.BlockType(b.Type),
        Content:    json.RawMessage(b.Content),
        Position:   b.Position,
        IsActive:   b.IsActive,
        ClickCount: b.ClickCount,
        CreatedAt:  b.CreatedAt,
        UpdatedAt:  b.UpdatedAt,
    }
}
```

### 3.3. Các method profile

Map sqlc → domain, xử lý `pgx.ErrNoRows` → `domain.ErrNotFound` (đúng như
`postgresUserRepo.FindByID`):

- `Create` → `q.CreateProfile(ctx, sqlcgen.CreateProfileParams{...})`. Field
  JSONB truyền `[]byte(p.Appearance)`.
- `GetByUserID` → `q.GetProfileByUserID`; `ErrNoRows`→`ErrNotFound`; rồi
  `profileToDomain`.
- `GetPublishedByUsername` → `q.GetPublishedProfileByUsername`; tương tự.
- `Update` → `q.UpdateProfile(ctx, sqlcgen.UpdateProfileParams{...})`.
- `UsernameExists` → `q.ProfileUsernameExists(ctx, sqlcgen.ProfileUsernameExistsParams{Username: username, ID: exclude})`.

### 3.4. Các method block

- `Create` → `q.CreateBlock(...)` (`Content: []byte(b.Content)`, `Type: string(b.Type)`).
- `ListByProfile` / `ListActiveByProfile` → loop rows, append `blockToDomain(r)`
  vào `[]domain.Block` (khởi tạo `make([]domain.Block, 0, len(rows))`).
- `Update` → `q.UpdateBlock(...)` (đã có `profile_id` trong WHERE → chống IDOR).
- `Delete` → `q.DeleteBlock(ctx, sqlcgen.DeleteBlockParams{ID: id, ProfileID: profileID})`.
- `IncrementClick` → `q.IncrementBlockClick(ctx, id)`.

### 3.5. `Reorder` — dùng transaction

`sqlcgen.Queries` có `WithTx(tx pgx.Tx) *Queries` (xem `sqlcgen/db.go`). Mở tx từ
`pool`, gọi `UpdateBlockPosition` cho từng id theo thứ tự (index = position mới),
commit. `profile_id` trong WHERE đảm bảo chỉ đụng block của chủ sở hữu.

```go
func (r *postgresBlockRepo) Reorder(ctx context.Context, profileID uuid.UUID, ids []uuid.UUID) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx) // no-op nếu đã Commit

    qtx := r.q.WithTx(tx)
    now := time.Now().UTC()
    for i, id := range ids {
        if err := qtx.UpdateBlockPosition(ctx, sqlcgen.UpdateBlockPositionParams{
            ID:        id,
            Position:  int32(i),
            UpdatedAt: now,
            ProfileID: profileID,
        }); err != nil {
            return err
        }
    }
    return tx.Commit(ctx)
}
```

> Vì `WHERE id = $1 AND profile_id = $4`, id không thuộc profile sẽ update 0 dòng
> (không lỗi). Nếu muốn chặt chẽ, ở usecase hãy nạp danh sách block hiện có và
> kiểm tra `ids` khớp đúng tập đó trước khi gọi `Reorder`.

---

## 4. Usecase — validate + nghiệp vụ

File `internal/usecase/profile_usecase.go`. Đây là **trái tim** của mô hình block
(toàn bộ validate `content` theo `type` nằm ở đây). Thay `struct{}` rỗng bằng:

```go
type ProfileUseCase struct {
    profileRepo repository.ProfileRepository
    blockRepo   repository.BlockRepository
}

func NewProfileUseCase(p repository.ProfileRepository, b repository.BlockRepository) *ProfileUseCase {
    return &ProfileUseCase{profileRepo: p, blockRepo: b}
}
```

### 4.1. Validate username

Hàm `normalizeUsername(s string) string` → `strings.ToLower(strings.TrimSpace(s))`.
Hàm `validateUsername(u string) error`:

- Độ dài 3–30 (sau normalize).
- Regex `^[a-z0-9_.]+$` (compile sẵn ở package-level `var usernameRe = regexp.MustCompile(...)`).
- Không bắt đầu/kết thúc bằng `.`, không chứa `..`.
- Không nằm trong set `reserved` (`map[string]struct{}{"admin":{}, "api":{}, ...}` —
  lấy danh sách ở mục 5 của design).
- Sai bất kỳ điều nào → `domain.ErrBadRequest`.

### 4.2. Validate `content` theo `type`

Cốt lõi. Một hàm switch theo type, dùng `json.Decoder` + `DisallowUnknownFields`
để **từ chối field thừa**:

```go
func validateBlockContent(t domain.BlockType, raw json.RawMessage) error {
    switch t {
    case domain.BlockTypeLink:
        var c struct {
            Title     string `json:"title"`
            URL       string `json:"url"`
            Icon      string `json:"icon"`
            Thumbnail string `json:"thumbnail"`
        }
        if err := strictUnmarshal(raw, &c); err != nil { return domain.ErrBadRequest }
        if c.Title == "" || len(c.Title) > 80 { return domain.ErrBadRequest }
        if err := validateHTTPURL(c.URL); err != nil { return err }
        return nil
    case domain.BlockTypeSocials:
        // items 1..20; mỗi item platform ∈ whitelist, url http/https
        ...
    case domain.BlockTypeHeader:
        // text bắt buộc ≤ 80
        ...
    default:
        return domain.ErrBadRequest // type lạ
    }
}
```

Helper dùng chung:

- `strictUnmarshal(raw, v)` → `d := json.NewDecoder(bytes.NewReader(raw)); d.DisallowUnknownFields(); return d.Decode(v)`.
- `validateHTTPURL(s)` → `url.Parse`, bắt buộc `scheme ∈ {http, https}`
  (chặn `javascript:`, `data:` để tránh XSS khi client render). Rỗng → lỗi.
- `socials.platform` whitelist: `map[string]bool{"instagram":true,"github":true,"x":true,"youtube":true,"tiktok":true,...}`.

### 4.3. Validate appearance

`validateAppearance(raw json.RawMessage) error`: parse strict, kiểm `theme` ∈
enum (`{"dark","light",...}`), mã màu khớp `^#[0-9a-fA-F]{6}$`. JSON lạ → lỗi.
Rỗng/`{}` → hợp lệ (default).

### 4.4. Các method nghiệp vụ

Chữ ký gợi ý (owner luôn truyền `userID uuid.UUID` lấy từ JWT, **không** từ body):

```go
// Profile
GetMyProfile(ctx, userID) (*domain.Profile, error)
GetPublicProfile(ctx, username string) (*domain.Profile, error) // kèm block active
CreateProfile(ctx, userID uuid.UUID, p *domain.Profile) error
UpdateProfile(ctx, userID uuid.UUID, p *domain.Profile) error

// Block
ListMyBlocks(ctx, userID) ([]domain.Block, error)
CreateBlock(ctx, userID uuid.UUID, b *domain.Block) error
UpdateBlock(ctx, userID uuid.UUID, b *domain.Block) error
DeleteBlock(ctx, userID, blockID uuid.UUID) error
ReorderBlocks(ctx, userID uuid.UUID, orderedIDs []uuid.UUID) error
RecordClick(ctx, username string, blockID uuid.UUID) (redirectURL string, err error)
```

Điểm mấu chốt từng method:

- **CreateProfile:** normalize + validate username; `GetByUserID` để chặn tạo
  trùng (đã có → `ErrConflict`); `UsernameExists(username, uuid.Nil)`; validate
  appearance (mặc định `{}` nếu rỗng); set `ID=uuid.New()`, `CreatedAt/UpdatedAt=time.Now().UTC()`;
  `profileRepo.Create`.
- **UpdateProfile:** `GetByUserID` lấy profile hiện tại (→ có `profile.ID`);
  nếu username đổi thì validate + `UsernameExists(new, profile.ID)`; validate
  appearance; gán field, `UpdatedAt=now`; `Update`. (Quyền sở hữu đảm bảo vì tra
  theo `userID` của chính mình.)
- **CreateBlock:** `GetByUserID` → `profile.ID`; **đếm block hiện có**
  (`ListByProfile`) để chặn vượt giới hạn (vd. 100) → `ErrBadRequest`;
  `validateBlockContent(b.Type, b.Content)`; set `ProfileID=profile.ID`,
  `ID=uuid.New()`, `Position` = cuối danh sách (len hiện tại), `IsActive=true`,
  timestamps; `Create`.
- **UpdateBlock:** `GetByUserID` → `profile.ID`; set `b.ProfileID=profile.ID`
  (chống IDOR — không tin profile_id từ client); validate content; `Update`
  (query đã kèm `profile_id`).
- **DeleteBlock:** `GetByUserID` → `profile.ID`; `Delete(blockID, profile.ID)`.
- **ReorderBlocks:** `GetByUserID` → `profile.ID`; (tùy chọn) nạp block hiện có
  để xác minh `orderedIDs` đúng tập; `Reorder(profile.ID, orderedIDs)`.
- **GetPublicProfile:** `GetPublishedByUsername(username)` (không thấy →
  `ErrNotFound`); `ListActiveByProfile(profile.ID)`; gán vào `profile.Blocks`;
  trả `*domain.Profile` (handler sẽ map sang `PublicProfile`).
- **RecordClick:** `GetPublishedByUsername`; nạp block active, tìm block theo id;
  chỉ cho click type clickable (`link`); parse `content.url`; `IncrementClick(id)`;
  trả url. (Với `socials` nhiều item — xem mục 8.5 design, v1 có thể bỏ qua hoặc
  trả lỗi `ErrBadRequest`.)

> Pattern lỗi trùng/không thấy: bắt chước `UserUseCase.CreateUser` —
> `if _, err := repo.GetByUserID(...); err == nil { return ErrConflict } else if err != ErrNotFound { return err }`.

---

## 5. Handler — `profile_handler.go` + `block_handler.go`

Theo khung `user_handler.go`: struct giữ `*usecase.ProfileUseCase`, request DTO
có tag `binding`, map lỗi domain → HTTP status, dùng `respondInternal` cho lỗi lạ.

### 5.1. Đọc user id từ JWT

`AuthRequired` đã `c.Set(ContextUserID, claims.UserID)` với kiểu **`uuid.UUID`**.
Trong handler owner:

```go
userID := c.MustGet(middleware.ContextUserID).(uuid.UUID)
```

(Có thể viết helper nhỏ `currentUserID(c) (uuid.UUID, bool)` dùng `c.Get` cho an
toàn, nhưng vì route đã qua `AuthRequired` nên `MustGet` là đủ.)

### 5.2. Map lỗi → HTTP (dùng lại đúng style user_handler)

| Lỗi domain | HTTP |
|---|---|
| `ErrBadRequest`, `ErrInvalidID` | 400 |
| `ErrNotFound` | 404 |
| `ErrConflict` | 409 |
| khác | 500 (`respondInternal`) |

Bind lỗi (`ShouldBindJSON`) → 400 trực tiếp như user_handler.

### 5.3. ProfileHandler

- `GetPublic(c)` — đọc `c.Param("username")`; gọi `GetPublicProfile`; **map sang
  `domain.PublicProfile`** (không trả `user_id`/`is_published`/`click_count`);
  404 nếu `ErrNotFound`.
- `RecordClick(c)` — `username` + `id` từ param (`uuid.Parse(c.Param("id"))`,
  lỗi → 400 `ErrInvalidID`); gọi `RecordClick`; trả `{"url": "..."}`.
- `GetMine(c)` — `userID` từ JWT; `GetMyProfile`.
- `Create(c)` — bind `createProfileRequest`; dựng `domain.Profile`; `CreateProfile`; 201.
- `Update(c)` — bind `updateProfileRequest`; `UpdateProfile`; 200.

Request DTO ví dụ:

```go
type createProfileRequest struct {
    Username    string          `json:"username" binding:"required"`
    DisplayName string          `json:"display_name"`
    Bio         string          `json:"bio"`
    AvatarURL   string          `json:"avatar_url"`
    Appearance  json.RawMessage `json:"appearance"`
}
```

### 5.4. BlockHandler

- `List(c)` — `ListMyBlocks`.
- `Create(c)` — bind `{type, content}`; dựng `domain.Block{Type, Content}`;
  `CreateBlock`; 201.
- `Update(c)` — `id` param + bind body; `UpdateBlock`.
- `Delete(c)` — `id` param; `DeleteBlock`; 200.
- `Reorder(c)` — bind `{ "order": ["<id>", ...] }`; parse từng id sang
  `uuid.UUID` (lỗi → 400); `ReorderBlocks`.

Request DTO:

```go
type createBlockRequest struct {
    Type    string          `json:"type" binding:"required"`
    Content json.RawMessage `json:"content" binding:"required"`
}
type reorderRequest struct {
    Order []string `json:"order" binding:"required"`
}
```

> Có thể gom Profile + Block vào **một** handler struct (vd.
> `ProfileHandler` giữ luôn các method block) để wiring gọn. Tùy bạn; design gốc
> tách 2 handler cho rõ. Nếu tách, có thể cùng dùng chung 1 `ProfileUseCase`.

---

## 6. Wiring trong `cmd/api/main.go`

Thêm vào phần khởi tạo (sau `userRepo`):

```go
profileRepo := persistence.NewPostgresProfileRepo(pool)
blockRepo   := persistence.NewPostgresBlockRepo(pool)
profileUC   := usecase.NewProfileUseCase(profileRepo, blockRepo)
profileH    := handler.NewProfileHandler(profileUC)
blockH      := handler.NewBlockHandler(profileUC) // nếu tách handler
```

Route — thêm trong group `v1` (public KHÔNG có `AuthRequired`, owner CÓ):

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

> ⚠️ Thứ tự route gin: `/blocks/reorder` và `/blocks/:id` cùng prefix. Gin xử lý
> được static-vs-param trong cùng nhóm, nhưng nếu gặp panic "conflict", hãy đăng
> ký `reorder` (static) trước `:id`. Test kỹ route này.

---

## 7. Build / migrate / test

```bash
gofmt -w internal/...                 # chuẩn hóa tab (profile.go đang dùng space)
make sqlc        # chỉ cần nếu bạn sửa thêm file .sql (hiện đã sinh sẵn rồi)
make migrate-up  # áp 000003 + 000004 vào DB local
go build ./...   # phải xanh trước khi chạy
go vet ./...
make run         # hoặc go run ./cmd/api
```

Kiểm thử nhanh bằng curl (cần token từ `/auth/login` cho nhóm `/me`):

```bash
# tạo profile
curl -X POST localhost:8080/api/v1/me/profile \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"username":"hao","display_name":"Hao"}'

# thêm block link
curl -X POST localhost:8080/api/v1/me/blocks \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"type":"link","content":{"title":"Shop","url":"https://example.com"}}'

# publish (PUT /me/profile với is_published=true) rồi xem public
curl localhost:8080/api/v1/p/hao
```

Các ca biên nên thử: username reserved (`admin`) → 400; username trùng → 409;
tạo profile lần 2 → 409; block `content` thiếu `url` → 400; url `javascript:...`
→ 400; field thừa trong content → 400; xem `/p/<user chưa publish>` → 404; sửa
block của profile khác (đổi token) → 404/không đổi.

---

## 8. Checklist hoàn thành

- [ ] `gofmt` sạch toàn module.
- [ ] `domain`: thêm `PublicProfile`, `PublicBlock`.
- [ ] `repository/profile_repository.go`: 2 interface.
- [ ] `persistence`: impl `ProfileRepository` + `BlockRepository` (có `pool` cho tx).
- [ ] `usecase/profile_usecase.go`: validate username/appearance/content + đủ method.
- [ ] `handler`: `ProfileHandler`, `BlockHandler` (map lỗi đúng status).
- [ ] `main.go`: khởi tạo + route `/p` (public) và `/me` (auth).
- [ ] `go build ./...` xanh, chạy thử curl các ca biên ở mục 7.
- [ ] (Tùy chọn) Swagger annotation `// @...` cho endpoint mới, như `user_handler.go`.
```
