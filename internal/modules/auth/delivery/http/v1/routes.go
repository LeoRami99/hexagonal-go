package v1

import "github.com/gin-gonic/gin"

func RegisterRoutes(v1 *gin.RouterGroup, controller *Controller) {
	auth := v1.Group("/auth")
	auth.POST("/login", controller.Login)
	auth.POST("/refresh", controller.Refresh)
	auth.POST("/logout", controller.Logout)
}
