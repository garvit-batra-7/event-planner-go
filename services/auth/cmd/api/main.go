// Package main Auth Service
//
// @title           Auth Service API
// @version         1.0
// @description     Handles user registration, login & JWT authentication.
// @BasePath        /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and a valid JWT.
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
    "event-planner-go/services/auth/internal/auth"
    "event-planner-go/services/auth/internal/user"
    "event-planner-go/services/auth/internal/config"
    _ "event-planner-go/services/auth/docs"
    
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
    cfg := config.Load()

    if cfg.IsProduction() {
        gin.SetMode(gin.ReleaseMode)
    }

    err := shared.InitLogger(shared.LoggerConfig{
        Level:     shared.LogLevel(cfg.Logging.Level),
        Format:    cfg.Logging.Format,
        Output:    cfg.Logging.Output,
        Component: "auth-service",
    })

    if err != nil {
        log.Fatal(err)
    }

    logger := shared.GetLogger()

    // in main.go after loading config
    // logger.Info("JWT secret loaded",
    //     "len", len(cfg.Auth.JWTSecret),
    // )
    
    db, err := sql.Open("sqlite3", "./data/auth.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    userRepo := user.NewRepository(db)
    authService := auth.NewService(userRepo, cfg, logger)
    authHandler := auth.NewHandler(authService)

    r := gin.New()
    r.Use(shared.LoggingMiddleware(), gin.Recovery())
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    v1 := r.Group("/api/v1/auth")
    {
        v1.GET("/users/:id", authHandler.GetByID)
        v1.POST("/login", authHandler.Login)
        v1.POST("/register", authHandler.Register)
    }

    if err := r.Run(":8081"); err != nil {
        log.Fatal(err)
    }
}
