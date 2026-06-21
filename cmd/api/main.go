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
	infradb "github.com/lukenguyen/fracture/internal/infrastructure/db"
	"github.com/lukenguyen/fracture/internal/infrastructure/persistence"
	"github.com/lukenguyen/fracture/internal/usecase"
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
func main() {
	cfg := config.Load()

	pool := infradb.NewPostgresPool(cfg)
	defer pool.Close()

	userRepo := persistence.NewPostgresUserRepo(pool)
	userUC := usecase.NewUserUseCase(userRepo)
	userH := handler.NewUserHandler(userUC)

	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.GET("/health", healthHandler)

	registerSwagger(r)

	v1 := r.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.GET("/:id", userH.GetUser)
			users.POST("", userH.CreateUser)
			users.PUT("/:id", userH.UpdateUser)
			users.DELETE("/:id", userH.DeleteUser)
		}
	}

	srv := &http.Server{
		Addr: ":" + cfg.App.Port,
		Handler: r,
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