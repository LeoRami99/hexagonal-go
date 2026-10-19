package v1

import (
	"net/http"
	"strconv"

	"hexagonal-go-backend/internal/platform/http/response"

	"hexagonal-go-backend/internal/modules/courses/application"

	"github.com/gin-gonic/gin"
)

type Controller struct{ service application.CoursesService }

func NewController(service application.CoursesService) *Controller {
	return &Controller{service: service}
}

func (h *Controller) RegisterRoutes(rg *gin.RouterGroup) {
	routes := rg.Group("/courses")
	routes.POST("", h.Create)
	routes.GET("", h.List)
	routes.GET("/:id", h.FindByID)
	routes.PUT("/:id", h.Update)
	routes.DELETE("/:id", h.Delete)
}

// Create godoc
// @Summary      Crear curso
// @Description  Crea un curso asociado a un proyecto y tenant. El slug debe ser único dentro del proyecto.
// @Tags         courses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateCourseRequest true "Datos del curso"
// @Success      201 {object} response.Envelope{data=CreateCourseResponse}
// @Failure      400 {object} response.ErrorResponse "Petición malformada"
// @Failure      401 {object} response.ErrorResponse "No autenticado"
// @Failure      409 {object} response.ErrorResponse "Ya existe un curso con el mismo slug en el proyecto"
// @Failure      422 {object} response.ErrorResponse "Datos del curso inválidos"
// @Failure      500 {object} response.ErrorResponse "Error interno"
// @Router       /courses [post]
func (h *Controller) Create(c *gin.Context) {
	var request CreateCourseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.ValidationError(c, err)
		return
	}

	course, err := h.service.Create(c.Request.Context(), application.CreateCourseInput{
		TenantID:     request.TenantID,
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

// FindByID godoc
// @Summary      Obtener curso
// @Description  Retorna un curso por su identificador ULID.
// @Tags         courses
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "ULID del curso" minlength(26) maxlength(26)
// @Success      200 {object} response.Envelope{data=CreateCourseResponse}
// @Failure      401 {object} response.ErrorResponse "No autenticado"
// @Failure      404 {object} response.ErrorResponse "Curso no encontrado"
// @Failure      422 {object} response.ErrorResponse "Identificador inválido"
// @Failure      500 {object} response.ErrorResponse "Error interno"
// @Router       /courses/{id} [get]
func (h *Controller) FindByID(c *gin.Context) {
	course, err := h.service.FindByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Envelope{Data: createCourseResponse(course)})
}

// List godoc
// @Summary      Listar cursos
// @Description  Retorna cursos paginados y permite buscar por título, slug o resumen, filtrar por nivel y estado, y ordenar el resultado.
// @Tags         courses
// @Produce      json
// @Security     BearerAuth
// @Param        page   query int    false "Número de página" default(1) minimum(1)
// @Param        limit  query int    false "Elementos por página" default(20) minimum(1) maximum(100)
// @Param        search query string false "Texto para buscar en título, slug o resumen"
// @Param        level  query string false "Nivel del curso" Enums(beginner, intermediate, advanced)
// @Param        status query string false "Estado del curso" Enums(draft, published, archived)
// @Param        sort   query string false "Campo de ordenamiento" Enums(created_at, updated_at, title) default(created_at)
// @Param        order  query string false "Dirección del ordenamiento" Enums(asc, desc) default(desc)
// @Success      200 {object} response.Envelope{data=[]CreateCourseResponse,meta=object}
// @Failure      401 {object} response.ErrorResponse "No autenticado"
// @Failure      500 {object} response.ErrorResponse "Error interno"
// @Router       /courses [get]
func (h *Controller) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	filter := application.CoursesFilter{Page: page, Limit: limit, Search: c.Query("search"), Level: c.Query("level"), Status: c.Query("status"), Sort: c.DefaultQuery("sort", "created_at"), Order: c.DefaultQuery("order", "desc")}
	courses, total, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	output := make([]CreateCourseResponse, len(courses))
	for i := range courses {
		output[i] = createCourseResponse(&courses[i])
	}
	c.JSON(http.StatusOK, response.Envelope{Data: output, Meta: gin.H{"page": page, "limit": limit, "total": total}})
}

// Update godoc
// @Summary      Actualizar curso
// @Description  Actualiza parcialmente un curso. Solo se modifican los campos enviados en el cuerpo.
// @Tags         courses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "ULID del curso" minlength(26) maxlength(26)
// @Param        request body UpdateCourseRequest true "Campos del curso que se desean actualizar"
// @Success      200 {object} response.Envelope{data=CreateCourseResponse}
// @Failure      400 {object} response.ErrorResponse "Petición malformada"
// @Failure      401 {object} response.ErrorResponse "No autenticado"
// @Failure      404 {object} response.ErrorResponse "Curso no encontrado"
// @Failure      409 {object} response.ErrorResponse "El slug ya está utilizado en el proyecto"
// @Failure      422 {object} response.ErrorResponse "Datos del curso inválidos"
// @Failure      500 {object} response.ErrorResponse "Error interno"
// @Router       /courses/{id} [put]
func (h *Controller) Update(c *gin.Context) {
	var request UpdateCourseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.ValidationError(c, err)
		return
	}
	course, err := h.service.Update(c.Request.Context(), c.Param("id"), application.UpdateCourseInput{Title: request.Title, Slug: request.Slug, Summary: request.Summary, Description: request.Description, Level: request.Level, ThumbnailURL: request.ThumbnailURL, Status: request.Status})
	if err != nil {
		response.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Envelope{Data: createCourseResponse(course)})
}

// Delete godoc
// @Summary      Eliminar curso
// @Description  Elimina permanentemente un curso por su identificador.
// @Tags         courses
// @Security     BearerAuth
// @Param        id path string true "ULID del curso" minlength(26) maxlength(26)
// @Success      204 "Curso eliminado"
// @Failure      401 {object} response.ErrorResponse "No autenticado"
// @Failure      404 {object} response.ErrorResponse "Curso no encontrado"
// @Failure      422 {object} response.ErrorResponse "Identificador inválido"
// @Failure      500 {object} response.ErrorResponse "Error interno"
// @Router       /courses/{id} [delete]
func (h *Controller) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		response.WriteError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
