package main

import (
	"fmt"
	"guycanella-url-shortner/internal/cache"
	"guycanella-url-shortner/internal/config"
	"guycanella-url-shortner/internal/database"
	"guycanella-url-shortner/internal/handlers"
	"guycanella-url-shortner/internal/repository"
	"guycanella-url-shortner/internal/services"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	if err := database.Connect(cfg); err != nil {
		log.Fatalf("❌ DB connection failed: %v", err)
	}

	if err := database.Migrate(); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	cache.InitRedis()

	repo := repository.NewURLRepository()
	service := services.NewURLService(repo, cfg)
	handler := handlers.NewURLHandler(service)

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			affected, err := service.DeactivateExpired()
			if err != nil {
				log.Printf("❌ Error deactivating expired URLs: %v", err)
				continue
			}

			if affected > 0 {
				log.Printf("🧹 Deactivated %d expired URLs", affected)
			}
		}
	}()

	routes := gin.Default()

	routes.POST("/shorten", handler.CreateURL)
	routes.GET("/stats/:shortCode", handler.Stats)
	routes.DELETE("/:id", handler.Delete)

	routes.GET("/:shortCode", handler.Redirect)

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("🚀 Server running at http://%s\n", addr)

	if err := routes.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("❌ failed to start server: %v", err)
	}
}
