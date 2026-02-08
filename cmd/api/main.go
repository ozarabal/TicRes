package main

import (
	"context"
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
	"ticres/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// 0. Initialize Logger
	mode := os.Getenv("APP_MODE")
	if mode == "" {
		mode = "development"
	}
	if err := logger.Init(mode); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("starting application", logger.String("mode", mode))

	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("load config failed", logger.Err(err))
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
		logger.Fatal("database connection failed", logger.Err(err))
	}
	defer dbPool.Close()
	logger.Info("database connected successfully")

	redisClient, err := database.NewRedClient(cfg.Cache.Host, cfg.Cache.Port, cfg.Cache.Password)
	if err != nil {
		logger.Fatal("redis connection failed", logger.Err(err))
	}
	logger.Info("redis connected successfully")

	// 3. Init Layers (Dependency Injection)
	userRepo := repository.NewUserRepository(dbPool)
	eventRepo := repository.NewEventRepository(dbPool, redisClient)
	bookingRepo := repository.NewBookingRepository(dbPool)
	transactionRepo := repository.NewTransactionRepository(dbPool)
	refundRepo := repository.NewRefundRepository(dbPool)

	timeoutContext := time.Duration(5) * time.Second
	notifWorker := worker.NewNotificationWorker(userRepo, bookingRepo, transactionRepo, refundRepo)
	notifWorker.Start()

	userUsecase := usecase.NewUserUsecase(userRepo, timeoutContext, cfg.JWT.Secret, cfg.JWT.ExpTime)
	eventUseCase := usecase.NewEventUsecase(eventRepo, timeoutContext, notifWorker)
	bookingUseCase := usecase.NewBookingUsecase(bookingRepo, transactionRepo, timeoutContext, notifWorker)
	paymentUseCase := usecase.NewPaymentUsecase(bookingRepo, transactionRepo, timeoutContext)

	// Handlers
	userHandler := delivery.NewUserHandler(userUsecase, bookingUseCase)
	eventHandler := delivery.NewEventHandler(eventUseCase)
	bookingHandler := delivery.NewBookingHandler(bookingUseCase)
	adminHandler := delivery.NewAdminHandler(bookingUseCase)
	paymentHandler := delivery.NewPaymentHandler(paymentUseCase)

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
			protected.POST("/payments", paymentHandler.ProcessPayment)
			protected.GET("/payments/:booking_id", paymentHandler.GetPaymentStatus)
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
		logger.Info("server starting", logger.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", logger.Err(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", logger.Err(err))
	}

	notifWorker.Stop()

	logger.Info("server exited")
}
