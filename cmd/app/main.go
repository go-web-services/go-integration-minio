package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	platform "github.com/go-web-services/go-web-platform/entrypoint"
	platformMiddleware "github.com/go-web-services/go-web-platform/middleware"

	"github.com/gin-gonic/gin"
	"github.com/go-web-services/go-web-platform/logger"

	"github.com/go-web-services/go-integration-minio/config"
	"github.com/go-web-services/go-integration-minio/docs" // Required to load swagger docs
	"github.com/go-web-services/go-integration-minio/internal/service"
	minioHTTP "github.com/go-web-services/go-integration-minio/internal/transport/http"
)

// @title           Minio Integration API
// @version         1.0
// @basePath /api

func main() {
	// Load configuration from environment variables
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Debug print TestEnv
	log.Printf("TestEnv value: %s", cfg.App.TestEnv)

	// Initialize custom logger
	logg := logger.NewLogger(cfg.App.Env)

	// Prepare services
	minioService, err := service.NewMinioService(logg)
	if err != nil {
		logg.Fatal("Failed to initialize MinIO service: ", err)
	}

	// Prepare HTTP server (router)
	router := gin.Default()
	// Platform integration
	platform.SetupPlatform(
		router,
		logg,
		nil,
		platformMiddleware.DefaultLoggingConfig(),
		cfg.App.Env,
	)

	minioHTTP.SetupRouter(router, minioService, logg)

	// Swagger docs
	swaggerBasePath := "/api"

	if cfg.App.SwaggerBasePath != "" {
		swaggerBasePath = "/" + cfg.App.SwaggerBasePath + swaggerBasePath
	}

	docs.SwaggerInfo.BasePath = swaggerBasePath

	// Start HTTP server
	serverAddr := fmt.Sprintf(":%d", cfg.App.Port)
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}
	logg.Info("Starting server on port ", cfg.App.Port)
	go func() {
		if e := router.Run(serverAddr); e != nil {
			logg.Fatal("Failed to start HTTP server: ", e)
		}
	}()

	// Graceful shutdown handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logg.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if e := srv.Shutdown(ctx); e != nil {
		logg.Fatal("Server forced to shutdown: ", e)
	}

	logg.Info("Server stopped.")
}
