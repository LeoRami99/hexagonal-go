package v1

import (
	authapp "hexagonal-go-backend/internal/modules/auth/application"
	"hexagonal-go-backend/internal/modules/users/domain"
	"hexagonal-go-backend/internal/platform/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(v1 *gin.RouterGroup, controller *Controller, tokens authapp.TokenProvider) {
	protected := v1.Group("")
	protected.Use(middleware.Authenticate(tokens))
	protected.GET("/users/:id", controller.Get)

	admin := protected.Group("")
	admin.Use(middleware.RequireRole(domain.RoleAdmin))
	admin.POST("/users", controller.Create)
	admin.GET("/users", controller.List)
	admin.PUT("/users/:id", controller.Update)
	admin.DELETE("/users/:id", controller.Delete)
}
