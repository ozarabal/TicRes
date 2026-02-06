package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ticres/internal/config"
	delivery "ticres/internal/delivery/http"
	"ticres/internal/delivery/http/middleware"
	"ticres/internal/repository"
	"ticres/internal/usecase"
	"ticres/internal/worker"
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

	redisClient, err := database.NewRedClient(cfg.Cache.Host, cfg.Cache.Port, cfg.Cache.Password)
	if err != nil {
		log.Fatalf("Gagal connect Redis: %v", err)
	}

	// 3. Init Layers (Dependency Injection)
	userRepo := repository.NewUserRepository(dbPool)
	eventRepo := repository.NewEventRepository(dbPool, redisClient)
	bookingRepo := repository.NewBookingRepository(dbPool)

	timeoutContext := time.Duration(5) * time.Second
	notifWorker := worker.NewNotificationWorker(userRepo, bookingRepo)
	notifWorker.Start()

	userUsecase := usecase.NewUserUsecase(userRepo, timeoutContext, cfg.JWT.Secret, cfg.JWT.ExpTime)
	eventUseCase := usecase.NewEventUsecase(eventRepo, timeoutContext, notifWorker)
	bookingUseCase := usecase.NewBookingUsecase(bookingRepo, timeoutContext, notifWorker)

	// Handlers
	userHandler := delivery.NewUserHandler(userUsecase, bookingUseCase)
	eventHandler := delivery.NewEventHandler(eventUseCase)
	bookingHandler := delivery.NewBookingHandler(bookingUseCase)
	adminHandler := delivery.NewAdminHandler(bookingUseCase)

	// 4. Setup Router (Gin)
	r := gin.Default()

	// CORS middleware for frontend
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	v1 := r.Group("/api/v1")
	{
		// Public routes
		v1.POST("/register", userHandler.Register)
		v1.POST("/login", userHandler.Login)
		v1.GET("/events", eventHandler.List)
		v1.GET("/events/:id", eventHandler.GetByID)

		// Protected routes (authenticated users)
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			protected.GET("/me", userHandler.Me)
			protected.GET("/me/bookings", userHandler.GetMyBookings)
			protected.POST("/events", eventHandler.Create)
			protected.POST("/bookings", bookingHandler.Create)
		}

		// Admin routes
		adminGroup := v1.Group("/admin")
		adminGroup.Use(middleware.AuthMiddleware(cfg.JWT.Secret), middleware.AdminMiddleware(cfg.JWT.Secret))
		{
			adminGroup.PUT("/events/:id", eventHandler.Update)
			adminGroup.DELETE("/events/:id", eventHandler.Delete)
			adminGroup.GET("/bookings", adminHandler.GetAllBookings)
			adminGroup.GET("/events/:id/bookings", adminHandler.GetEventBookings)
		}
	}

	// Graceful shutdown Setup
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	// 5. Run Server
	go func() {
		log.Printf("Server running on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	notifWorker.Stop()

	log.Print("Server exiting...")
}
