package v1

import "github.com/gin-gonic/gin"

func RegisterRoutes(v1 *gin.RouterGroup, controller *Controller) {
	controller.RegisterRoutes(v1)
}
