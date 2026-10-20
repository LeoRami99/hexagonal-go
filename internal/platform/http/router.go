package http

import (
	"log/slog"
	"time"

	courseshttp "hexagonal-go-backend/internal/modules/courses/delivery/http/v1"

	authapp "hexagonal-go-backend/internal/modules/auth/application"
	authhttp "hexagonal-go-backend/internal/modules/auth/delivery/http/v1"
	tenanthttp "hexagonal-go-backend/internal/modules/tenants/delivery/http/v1"
	usersapp "hexagonal-go-backend/internal/modules/users/application"
	usershttp "hexagonal-go-backend/internal/modules/users/delivery/http/v1"
	"hexagonal-go-backend/internal/platform/config"
	"hexagonal-go-backend/internal/platform/http/middleware"

	"hexagonal-go-backend/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title                      Go Hexagonal API
// @version                    1.0
// @description                API REST construida con Go, Gin y Arquitectura Hexagonal.

// @contact.name              API Support
// @contact.email             soporte@empresa.com

// @BasePath                  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Escribe "Bearer " seguido de tu token JWT

func NewRouter(cfg config.Config, logger *slog.Logger, users *usershttp.Controller, auth *authhttp.Controller, tokens authapp.TokenProvider, userRepo usersapp.UserRepository, tenants ...*tenanthttp.Controller) *gin.Engine {
	return newRouter(cfg, logger, users, auth, nil, tokens, userRepo, tenants...)
}

func NewRouterWithCourses(cfg config.Config, logger *slog.Logger, users *usershttp.Controller, auth *authhttp.Controller, courses *courseshttp.Controller, tokens authapp.TokenProvider, userRepo usersapp.UserRepository, tenants ...*tenanthttp.Controller) *gin.Engine {
	return newRouter(cfg, logger, users, auth, courses, tokens, userRepo, tenants...)
}

func newRouter(cfg config.Config, logger *slog.Logger, users *usershttp.Controller, auth *authhttp.Controller, courses *courseshttp.Controller, tokens authapp.TokenProvider, userRepo usersapp.UserRepository, tenants ...*tenanthttp.Controller) *gin.Engine {
	docs.SwaggerInfo.Host = ""
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
	usershttp.RegisterRoutes(v1, users, tokens, userRepo)
	if courses != nil {
		courseshttp.RegisterRoutes(v1, courses)
	}
	if len(tenants) > 0 && tenants[0] != nil {
		tenants[0].RegisterRoutes(v1)
	}
	return router
}
