// Package main Event service.
// @title           Event Service API
// @version         1.0
// @description     API for managing events in the event-planner system.
// @BasePath        /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and your JWT.
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"

	"event-planner-go/pkg/shared"
	"event-planner-go/services/event/internal/config"
	eventsvc "event-planner-go/services/event/internal/event"
	"event-planner-go/services/event/internal/attendeeclient"
	ginSwagger "github.com/swaggo/gin-swagger"
    swaggerFiles "github.com/swaggo/files"
)

func main() {
	// 1. Load service-specific config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// 2. Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// 3. Init logger (reuse shared logger)
	if err := shared.InitLogger(shared.LoggerConfig{
		Level:     shared.LogLevel(cfg.Logging.Level),
		Format:    cfg.Logging.Format,
		Output:    cfg.Logging.Output,
		Component: "event-service",
	}); err != nil {
		log.Fatal(err)
	}
	logger := shared.GetLogger()

	// in main.go after loading config
	// logger.Info("JWT secret loaded",
	//     "len", len(cfg.JWTSecret),
	// )

	// 4. Open DB for events (event service has its own DB)
	db, err := sql.Open("sqlite3", "./data/event.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Optional: tune connections (even for sqlite this is fine)
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	// 5. Wire dependencies
	eventRepo := eventsvc.NewRepository(db)

	// Attendee HTTP client (calling the attendee microservice)
	attClient := attendeeclient.NewHTTPClient(attendeeclient.Config{
		BaseURL: cfg.AttendeeService.BaseURL,           // e.g. "http://localhost:8082"
		Timeout: cfg.AttendeeService.Timeout,          // e.g. 5 * time.Second
	})

	// Event service now receives attendee client
	eventService := eventsvc.NewService(eventRepo, attClient, logger)
	eventHandler := eventsvc.NewHandler(eventService)

	// 6. Gin router
	r := gin.New()
	r.Use(
	    shared.LoggingMiddleware(),
	    gin.Recovery(),
	)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// public routes
	v1 := r.Group("/api/v1")
	{
	    v1.GET("/events", eventHandler.GetAll)
	    v1.GET("/events/:id", eventHandler.GetByID)
	}

	// protected routes
	protected := r.Group("/api/v1")
	protected.Use(shared.AuthMiddleware(cfg.JWTSecret)) // <- IMPORTANT
	{
	    protected.POST("/events", eventHandler.Create)
	    protected.PUT("/events/:id", eventHandler.Update)
	    protected.DELETE("/events/:id", eventHandler.Delete)
	}

	// 7. HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.HTTP.Addr, // e.g. "8082"
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting event service HTTP server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Event service failed to start", "error", err.Error())
			log.Fatal(err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down event service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err.Error())
	} else {
		logger.Info("Event service shut down gracefully")
	}
}
