package v1

import (
	"net/http"

	"hexagonal-go-backend/internal/platform/http/response"

	"hexagonal-go-backend/internal/modules/courses/application"

	"github.com/gin-gonic/gin"
)



type Controller struct { service application.CoursesService}

func NewController(service application.CoursesService) *Controller {
	return &Controller{service: service}
}


func (h *Controller) Create(c *gin.Context) {
	var request CreateCourseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.ValidationError(c, err)
		return
	}


	course, err := h.service.Create(c.Request.Context(), application.CreateCourseInput{
		ProjectID:    request.ProjectID,
		Title:        request.Title,
		Slug:         request.Slug,
		Summary:      request.Summary,
		Description:  request.Description,
		Level:        request.Level,
		ThumbnailURL: request.ThumbnailURL,
		Status:       request.Status,
	})
	if err != nil {
		response.WriteError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Envelope{Data: createCourseResponse(course)})

}