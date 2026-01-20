package main

import (
	"log"
	"time"

	"ticres/internal/config"
	delivery "ticres/internal/delivery/http" // Alias biar gak bentrok sama package net/http
	"ticres/internal/delivery/http/middleware"
	"ticres/internal/repository"
	"ticres/internal/usecase"
	"ticres/pkg/database"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Load config failed: %v", err)
	}

	// 2. Connect Database
	dbPool, err := database.NewPostgresConnection(
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Name,
	)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer dbPool.Close()

	// 3. Init Layers (Dependency Injection)
	// Repo butuh DB
	userRepo := repository.NewUserRepository(dbPool)
	
	// Usecase butuh Repo & Timeout Context
	timeoutContext := time.Duration(5) * time.Second
	userUsecase := usecase.NewUserUsecase(userRepo, timeoutContext, cfg.JWT.Secret, cfg.JWT.ExpTime)
	
	// Handler butuh Usecase
	userHandler := delivery.NewUserHandler(userUsecase)

	// 4. Setup Router (Gin)
	r := gin.Default()
    v1 := r.Group("/api/v1")
    {
        v1.POST("/register", userHandler.Register)
        v1.POST("/login", userHandler.Login)

		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			protected.GET("/me", userHandler.Me)
		}
    }

	// 5. Run Server
	log.Printf("Server berjalan di port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}