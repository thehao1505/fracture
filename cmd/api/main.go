package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lukenguyen/fracture/config"
	_ "github.com/lukenguyen/fracture/docs"
	"github.com/lukenguyen/fracture/internal/handler"
	"github.com/lukenguyen/fracture/internal/handler/middleware"
	infradb "github.com/lukenguyen/fracture/internal/infrastructure/db"
	"github.com/lukenguyen/fracture/internal/infrastructure/persistence"
	"github.com/lukenguyen/fracture/internal/usecase"
	"github.com/lukenguyen/fracture/pkg/token"
)

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// @title           Fracture API
// @version         1.0
// @description     A clean architecture Go API for managing users.
// @termsOfService  http://swagger.io/terms/
//
// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@swagger.io
//
// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html
//
// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http https
//
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Type "Bearer" followed by a space and your JWT access token. Example: "Bearer eyJhbGci..."
func main() {
	cfg := config.Load()

	pool := infradb.NewPostgresPool(cfg)
	defer pool.Close()

	tokenManager := token.NewManager(cfg.JWT.Secret, cfg.JWT.Expiry)

	userRepo := persistence.NewPostgresUserRepo(pool)
	userUC := usecase.NewUserUseCase(userRepo)
	userH := handler.NewUserHandler(userUC)

	authUC := usecase.NewAuthUseCase(userRepo, tokenManager)
	authH := handler.NewAuthHandler(authUC)

	profileRepo := persistence.NewPostgresProfileRepo(pool)
	blockRepo := persistence.NewPostgresBlockRepo(pool)
	profileUC := usecase.NewProfileUseCase(profileRepo, blockRepo)
	profileH := handler.NewProfileHandler(profileUC)
	blockH := handler.NewBlockHandler(profileUC)

	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.GET("/health", healthHandler)

	registerSwagger(r)

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login", authH.Login)
		}

		users := v1.Group("/users")
		users.Use(middleware.AuthRequired(tokenManager))
		{
			users.GET("/:id", userH.GetUser)
			users.POST("", userH.CreateUser)
			users.PUT("/:id", userH.UpdateUser)
			users.DELETE("/:id", userH.DeleteUser)
			users.GET("", userH.ListUsers)
		}

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
			me.PATCH("/blocks/reorder", blockH.Reorder)
			me.PUT("/blocks/:id", blockH.Update)
			me.DELETE("/blocks/:id", blockH.Delete)
		}
	}

	srv := &http.Server{
		Addr:              ":" + cfg.App.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("server running on port %s", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}

	log.Printf("server exited")
}
