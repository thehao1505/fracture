Đây là hướng dẫn đầy đủ để khởi tạo một dự án backend Golang với Clean Architecture, phù hợp với background NestJS của bạn (nhiều concept sẽ quen thuộc).

---

## Khởi tạo dự án Backend Go với Clean Architecture

### Tổng quan kiến trúc

Clean Architecture trong Go chia code thành 4 layer, dependency chỉ chạy từ ngoài vào trong:

```
┌─────────────────────────────────────┐
│         Frameworks & Drivers        │  ← HTTP, DB, Cache (gin, pgx, redis)
├─────────────────────────────────────┤
│       Interface Adapters            │  ← Handler, Repository impl, DTO
├─────────────────────────────────────┤
│         Use Cases                   │  ← Business logic
├─────────────────────────────────────┤
│           Entities                  │  ← Domain model, Business rules
└─────────────────────────────────────┘
```

---

### 1. Khởi tạo project

```bash
mkdir go-clean-api && cd go-clean-api
go mod init github.com/yourname/go-clean-api
```

---

### 2. Cấu trúc thư mục

```
go-clean-api/
├── cmd/
│   └── api/
│       └── main.go              # Entry point
├── internal/
│   ├── domain/                  # Layer 1: Entities
│   │   ├── user.go              # Domain model + business rules
│   │   └── errors.go            # Domain errors
│   ├── usecase/                 # Layer 2: Use Cases
│   │   ├── user_usecase.go
│   │   └── user_usecase_test.go
│   ├── repository/              # Layer 3: Interface (port)
│   │   └── user_repository.go   # Interface definition chỉ, không implement
│   ├── handler/                 # Layer 3: Interface Adapters (HTTP)
│   │   └── user_handler.go
│   └── infrastructure/          # Layer 4: Frameworks & Drivers
│       ├── persistence/
│       │   └── postgres_user_repo.go  # Repository implementation
│       ├── cache/
│       │   └── redis.go
│       └── db/
│           └── postgres.go
├── pkg/
│   ├── logger/
│   └── validator/
├── config/
│   └── config.go
├── .env
├── docker-compose.yml
└── go.sum
```

> **So sánh với NestJS**: `domain` ≈ entity/schema, `usecase` ≈ service, `repository` ≈ interface (port), `handler` ≈ controller, `infrastructure` ≈ module providers.

---

### 3. Cài dependencies

```bash
# HTTP framework
go get github.com/gin-gonic/gin

# PostgreSQL
go get github.com/jackc/pgx/v5

# Redis
go get github.com/redis/go-redis/v9

# Config từ .env
go get github.com/spf13/viper

# Validation
go get github.com/go-playground/validator/v10

# UUID
go get github.com/google/uuid

# JWT
go get github.com/golang-jwt/jwt/v5

# Logger
go get go.uber.org/zap
```

---

### 4. Viết từng layer

#### Layer 1 — Domain (`internal/domain/user.go`)

```go
package domain

import (
    "time"
    "github.com/google/uuid"
)

type User struct {
    ID        uuid.UUID
    Email     string
    Password  string
    Name      string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Business rule nằm ở đây, không phải ở usecase
func (u *User) IsActive() bool {
    return u.Email != ""
}
```

#### Layer 2 — Repository Interface (`internal/repository/user_repository.go`)

```go
package repository

import (
    "context"
    "github.com/google/uuid"
    "github.com/yourname/go-clean-api/internal/domain"
)

// Interface (port) — usecase chỉ biết đến interface này
type UserRepository interface {
    FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
    FindByEmail(ctx context.Context, email string) (*domain.User, error)
    Create(ctx context.Context, user *domain.User) error
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

#### Layer 2 — Use Case (`internal/usecase/user_usecase.go`)

```go
package usecase

import (
    "context"
    "github.com/yourname/go-clean-api/internal/domain"
    "github.com/yourname/go-clean-api/internal/repository"
)

type UserUseCase struct {
    userRepo repository.UserRepository  // Phụ thuộc vào interface, không phải impl
}

func NewUserUseCase(repo repository.UserRepository) *UserUseCase {
    return &UserUseCase{userRepo: repo}
}

func (uc *UserUseCase) GetUser(ctx context.Context, id string) (*domain.User, error) {
    uid, err := uuid.Parse(id)
    if err != nil {
        return nil, domain.ErrInvalidID
    }

    user, err := uc.userRepo.FindByID(ctx, uid)
    if err != nil {
        return nil, err
    }
    return user, nil
}
```

#### Layer 3 — Handler (`internal/handler/user_handler.go`)

```go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/yourname/go-clean-api/internal/usecase"
)

type UserHandler struct {
    userUC *usecase.UserUseCase
}

func NewUserHandler(uc *usecase.UserUseCase) *UserHandler {
    return &UserHandler{userUC: uc}
}

func (h *UserHandler) GetUser(c *gin.Context) {
    id := c.Param("id")

    user, err := h.userUC.GetUser(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": user})
}
```

#### Layer 4 — Infrastructure (`internal/infrastructure/persistence/postgres_user_repo.go`)

```go
package persistence

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/yourname/go-clean-api/internal/domain"
    "github.com/yourname/go-clean-api/internal/repository"
)

type postgresUserRepo struct {
    db *pgxpool.Pool
}

// Đảm bảo implement đúng interface — compile-time check
var _ repository.UserRepository = (*postgresUserRepo)(nil)

func NewPostgresUserRepo(db *pgxpool.Pool) repository.UserRepository {
    return &postgresUserRepo{db: db}
}

func (r *postgresUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
    var user domain.User
    err := r.db.QueryRow(ctx,
        "SELECT id, email, name, created_at FROM users WHERE id = $1", id,
    ).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)

    if err != nil {
        return nil, err
    }
    return &user, nil
}
// ... implement các method còn lại
```

---

### 5. Dependency Injection ở `main.go`

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/yourname/go-clean-api/internal/handler"
    "github.com/yourname/go-clean-api/internal/infrastructure/db"
    "github.com/yourname/go-clean-api/internal/infrastructure/persistence"
    "github.com/yourname/go-clean-api/internal/usecase"
)

func main() {
    // 1. Init infrastructure
    pool := db.NewPostgresPool()

    // 2. Wire dependencies (từ trong ra ngoài)
    userRepo := persistence.NewPostgresUserRepo(pool)
    userUC   := usecase.NewUserUseCase(userRepo)
    userH    := handler.NewUserHandler(userUC)

    // 3. HTTP routes
    r := gin.Default()
    v1 := r.Group("/api/v1")
    {
        v1.GET("/users/:id", userH.GetUser)
    }

    r.Run(":8080")
}
```

> **Không dùng framework DI** như `wire` hay `fx` khi mới bắt đầu — manual wiring ở `main.go` là đủ và rõ ràng hơn. Khi project lớn thì xem xét `google/wire`.

---

### 6. Docker Compose để dev

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: dev
      POSTGRES_PASSWORD: dev
      POSTGRES_DB: go_clean_dev
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

```bash
docker compose up -d
go run cmd/api/main.go
```

---

### 7. Những điểm quan trọng cần nhớ

| Quy tắc                                                                                      | Lý do                                                |
| -------------------------------------------------------------------------------------------- | ---------------------------------------------------- |
| **Interface định nghĩa ở layer dùng nó** (usecase/repository), không phải layer implement nó | Tránh import ngược                                   |
| **Domain không import gì từ internal**                                                       | Core logic phải độc lập hoàn toàn                    |
| **Handler chỉ biết usecase, không biết repo**                                                | Đúng dependency direction                            |
| `var _ MyInterface = (*MyStruct)(nil)`                                                       | Compile-time interface check — rất hay dùng trong Go |
| **Error của domain là custom type**, không dùng `errors.New` thô                             | Dễ handle ở handler layer                            |

---

### Bước tiếp theo gợi ý

Sau khi có skeleton, thêm theo thứ tự này:

1. **Config** với Viper + `.env`
2. **Middleware**: auth JWT, request logger, recovery
3. **Error handling** tập trung ở handler
4. **Migration** với `golang-migrate`
5. **Unit test** cho usecase (dùng `testify/mock` mock repo)
6. **Integration test** cho handler

Bạn muốn đi sâu vào phần nào trước — ví dụ auth flow, error handling, hay test strategy?
