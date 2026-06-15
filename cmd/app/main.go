package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

	_ "effective-mobile-task/docs"
	"effective-mobile-task/internal/config"
	"effective-mobile-task/internal/handlers"
	"effective-mobile-task/internal/logger"
	"effective-mobile-task/internal/middleware"
	"effective-mobile-task/internal/repository"
	"effective-mobile-task/internal/service"
)

// @title Subscription Service API
// @version 1.0
// @description API for managing user subscriptions
// @BasePath /api/v1

func main() {
	cfg := config.Load()

	log := logger.New(slog.LevelInfo, os.Stdout)

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	log.Info("connecting to database",
		"host", cfg.DBHost,
		"port", cfg.DBPort,
		"dbname", cfg.DBName,
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	log.Info("database connected successfully")

	repo := repository.NewSubscriptionRepository(db)
	srv := service.NewSubscriptionService(repo, log)
	h := handlers.NewSubscriptionHandler(srv)

	r := gin.New()
	r.Use(middleware.RequestLogger(log), gin.Recovery())

	api := r.Group("/api/v1")
	{
		api.POST("/subscriptions", h.CreateSubscription)
		api.GET("/subscriptions/:id", h.GetSubscription)
		api.PUT("/subscriptions/:id", h.UpdateSubscription)
		api.DELETE("/subscriptions/:id", h.DeleteSubscription)
		api.GET("/subscriptions", h.ListSubscriptions)
		api.GET("/subscriptions/total-cost", h.GetTotalCost)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := cfg.ServerPort
	httpSrv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Info("server starting", "port", port)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Info("shutting down server", "signal", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("server stopped gracefully")
}
