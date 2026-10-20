package http

import (
	"log/slog"
	"time"

	authapp "hexagonal-go-backend/internal/modules/auth/application"
	authhttp "hexagonal-go-backend/internal/modules/auth/delivery/http/v1"
	courseshttp "hexagonal-go-backend/internal/modules/courses/delivery/http/v1"
	usershttp "hexagonal-go-backend/internal/modules/users/delivery/http/v1"

	"hexagonal-go-backend/internal/platform/config"
	"hexagonal-go-backend/internal/platform/http/middleware"

	_ "hexagonal-go-backend/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title                      Go Hexagonal API
// @version                    1.0
// @description                API REST construida con Go, Gin y Arquitectura Hexagonal.

// @contact.name              API Support
// @contact.email             soporte@empresa.com

// @host                      localhost:8080
// @BasePath                  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Escribe "Bearer " seguido de tu token JWT

func NewRouter(cfg config.Config, logger *slog.Logger, users *usershttp.Controller, auth *authhttp.Controller, tokens authapp.TokenProvider, courses ...*courseshttp.Controller) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Use(middleware.RequestID(), middleware.Recovery(logger), middleware.Logger(logger), middleware.CORS(cfg.App.AllowedOrigins), middleware.SecurityHeaders(), middleware.BodyLimit(1<<20), middleware.RateLimit(cfg.Security.RateLimit, time.Minute))
	router.GET("/health", Health)
	router.GET("/ready", Ready)

	v1 := router.Group("/api/v1")
	authhttp.RegisterRoutes(v1, auth)
	usershttp.RegisterRoutes(v1, users, tokens)
	if len(courses) > 0 && courses[0] != nil {
		courseshttp.RegisterRoutes(v1, courses[0])
	}
	return router
}
