// Package main Attendee Service
//
// @title           Attendee Service API
// @version         1.0
// @description     Manages event attendance. Allows users to join events, leave events, and list attendees.
// @BasePath        /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Provide your JWT token with the prefix `Bearer `
//
// @contact.name   API Support
// @contact.url    https://github.com/your-org/event-planner-go
// @contact.email  support@example.com

package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"

	"event-planner-go/pkg/shared"
	"event-planner-go/services/attendee/internal/attendee"
	"event-planner-go/services/attendee/internal/config"
	"event-planner-go/services/attendee/internal/eventclient"
	"event-planner-go/services/attendee/internal/userclient"
	_ "event-planner-go/services/attendee/docs" 
	
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// 1) Load typed config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// 2) Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// 3) Logger
	if err := shared.InitLogger(shared.LoggerConfig{
		Level:     shared.LogLevel(cfg.Logging.Level),
		Format:    cfg.Logging.Format,
		Output:    cfg.Logging.Output,
		Component: "attendee-service",
	}); err != nil {
		log.Fatal(err)
	}
	logger := shared.GetLogger()
	defer func() { _ = shared.Cleanup() }()

	// 4) DB
	db, err := sql.Open("sqlite3", cfg.Database.DSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	// 5) Repo + clients + service
	attendeeRepo := attendee.NewRepository(db)

	evtClient := eventclient.NewHTTPClient(eventclient.Config{
		BaseURL: cfg.EventService.BaseURL,
		Timeout: cfg.EventService.Timeout,
	})

	userClient := userclient.NewHTTPClient(userclient.Config{
		BaseURL: cfg.UserService.BaseURL,
		Timeout: cfg.UserService.Timeout,
	})

	svc := attendee.NewService(attendeeRepo, evtClient, userClient, logger)
	h := attendee.NewHandler(svc)

	// 6) Router
	r := gin.New()
	r.Use(
		shared.LoggingMiddleware(),
		gin.Recovery(),
	)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	api := r.Group("/api/v1")
	{
	    att := api.Group("")
	    att.Use(shared.AuthMiddleware(cfg.JWTSecret))

	    att.POST("/events/:id/attendees/:userId", h.AddToEvent)
	    att.DELETE("/events/:id/attendees/:userId", h.RemoveFromEvent)
	    att.GET("/events/:id/attendees", h.GetUsersByEvent)
	    att.GET("/attendees/:id/events", h.GetEventsByUser)
	    att.DELETE("/events/:id/attendees", h.DeleteAttendeesForEvent)
	}


	logger.Info("Starting attendee service HTTP server", "addr", cfg.HTTP.Addr)
	if err := r.Run(cfg.HTTP.Addr); err != nil {
		logger.Error("Attendee service failed to start", "error", err.Error())
		log.Fatal(err)
	}
}
