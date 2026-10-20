package v1

import (
	authapp "hexagonal-go-backend/internal/modules/auth/application"
	usersapp "hexagonal-go-backend/internal/modules/users/application"
	"hexagonal-go-backend/internal/platform/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(v1 *gin.RouterGroup, controller *Controller, tokens authapp.TokenProvider, userRepo usersapp.UserRepository) {
	protected := v1.Group("")
	protected.Use(middleware.Authenticate(tokens, userRepo))
	protected.GET("/users/:id", controller.Get)
	protected.GET("/users/me/profile", controller.GetMyProfile)
	protected.PUT("/users/me/profile", controller.UpdateMyProfile)

	admin := protected.Group("")
	admin.Use(middleware.RequireSuperAdminDirect())
	admin.POST("/users", controller.Create)
	admin.GET("/users", controller.List)
	admin.PUT("/users/:id", controller.Update)
	admin.DELETE("/users/:id", controller.Delete)
}
