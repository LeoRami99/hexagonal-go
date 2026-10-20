package v1

import "github.com/gin-gonic/gin"

func RegisterRoutes(v1 *gin.RouterGroup, controller *Controller) {
	catalog := v1.Group("/courses")
	catalog.POST("", controller.CreateCourse)
	catalog.GET("/:id", controller.GetCourse)
	catalog.GET("/categories", controller.ListCategories)
	catalog.GET("/templates", controller.ListTemplates)
	catalog.GET("/templates/:id", controller.GetTemplate)
	catalog.GET("/templates/:id/versions", controller.ListVersions)
	catalog.GET("/versions/:id/syllabus", controller.GetSyllabus)
	catalog.POST("/categories", controller.CreateCategory)
	catalog.POST("/templates", controller.CreateTemplate)
	catalog.POST("/templates/:id/versions", controller.CreateVersion)
	catalog.PUT("/versions/:id/syllabus", controller.SetSyllabus)
}
