package http

import (
	"log/slog"
	"time"

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
	authRoutes := v1.Group("/auth")
	authRoutes.POST("/login", auth.Login)
	authRoutes.POST("/refresh", auth.Refresh)
	authRoutes.POST("/logout", auth.Logout)

	protected := v1.Group("")
	protected.Use(middleware.Authenticate(tokens, userRepo))
	protected.GET("/auth/me", auth.Me)
	protected.GET("/users/me/profile", users.GetMyProfile)
	protected.PUT("/users/me/profile", users.UpdateMyProfile)
	protected.GET("/users/:id", users.Get)

	admin := protected.Group("")
	admin.Use(middleware.RequireSuperAdminDirect())
	admin.POST("/users", users.Create)
	admin.GET("/users", users.List)
	admin.PUT("/users/:id", users.Update)
	admin.DELETE("/users/:id", users.Delete)

	if len(tenants) > 0 && tenants[0] != nil {
		tenants[0].RegisterRoutes(admin)
	}
	return router
}

