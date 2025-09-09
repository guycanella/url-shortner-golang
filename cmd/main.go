package main

import (
	"fmt"
	"guycanella-url-shortner/internal/config"
	"guycanella-url-shortner/internal/database"
	"guycanella-url-shortner/internal/handlers"
	"guycanella-url-shortner/internal/repository"
	"guycanella-url-shortner/internal/services"
	"log"

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

	repo := repository.NewURLRepository()
	service := services.NewURLService(repo, cfg)
	handler := handlers.NewURLHandler(service)

	routes := gin.Default()

	routes.POST("/shorten", handler.CreateURL)
	routes.GET("/stats/:shortCode", handler.Stats)
	routes.DELETE("/:id", handler.Delete)

	routes.GET("/:shortCode", handler.Redirect)

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("🚀 Server running at http://%s\n", addr)

	if err := routes.Run(addr); err != nil {
		log.Fatalf("❌ failed to start server: %v", err)
	}
}
