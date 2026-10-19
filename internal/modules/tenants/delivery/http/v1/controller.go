package v1

import (
	"errors"
	"net/http"

	"hexagonal-go-backend/internal/modules/tenants/application"
	"hexagonal-go-backend/internal/modules/tenants/domain"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	service application.TenantService
}

func NewController(service application.TenantService) *Controller {
	return &Controller{service: service}
}

func (ctrl *Controller) RegisterRoutes(rg *gin.RouterGroup) {
	tenants := rg.Group("/tenants")
	{
		tenants.POST("", ctrl.Create)
		tenants.GET("", ctrl.List)
		tenants.GET("/:id", ctrl.GetByID)
		tenants.PUT("/:id", ctrl.Update)
		tenants.DELETE("/:id", ctrl.Delete)
	}
}

// Create godoc
// @Summary      Crear tenant
// @Description  Crea una nueva organización/tenant en la plataforma.
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Param        request body application.CreateTenantInput true "Datos de creación del tenant"
// @Success      201 {object} map[string]any
// @Failure      400 {object} map[string]any "Datos inválidos"
// @Failure      409 {object} map[string]any "El tenant ya existe"
// @Failure      500 {object} map[string]any "Error interno"
// @Router       /tenants [post]
func (ctrl *Controller) Create(c *gin.Context) {
	var input application.CreateTenantInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	tenant, err := ctrl.service.Create(c.Request.Context(), input)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": tenant})
}

// GetByID godoc
// @Summary      Obtener tenant por ID
// @Description  Retorna los detalles de un tenant por su ID.
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Param        id path string true "ID del tenant"
// @Success      200 {object} map[string]any
// @Failure      404 {object} map[string]any "Tenant no encontrado"
// @Failure      500 {object} map[string]any "Error interno"
// @Router       /tenants/{id} [get]
func (ctrl *Controller) GetByID(c *gin.Context) {
	id := c.Param("id")
	tenant, err := ctrl.service.GetByID(c.Request.Context(), id)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tenant})
}

// List godoc
// @Summary      Listar tenants
// @Description  Retorna la lista de todos los tenants registrados.
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]any
// @Failure      500 {object} map[string]any "Error interno"
// @Router       /tenants [get]
func (ctrl *Controller) List(c *gin.Context) {
	tenants, err := ctrl.service.List(c.Request.Context())
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tenants})
}

// Update godoc
// @Summary      Actualizar tenant
// @Description  Actualiza la información de un tenant existente.
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Param        id path string true "ID del tenant"
// @Param        request body application.UpdateTenantInput true "Campos a actualizar del tenant"
// @Success      200 {object} map[string]any
// @Failure      400 {object} map[string]any "Datos inválidos"
// @Failure      404 {object} map[string]any "Tenant no encontrado"
// @Failure      500 {object} map[string]any "Error interno"
// @Router       /tenants/{id} [put]
func (ctrl *Controller) Update(c *gin.Context) {
	id := c.Param("id")
	var input application.UpdateTenantInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	tenant, err := ctrl.service.Update(c.Request.Context(), id, input)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tenant})
}

// Delete godoc
// @Summary      Eliminar tenant
// @Description  Elimina un tenant por su ID.
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Param        id path string true "ID del tenant"
// @Success      204 "Sin contenido"
// @Failure      404 {object} map[string]any "Tenant no encontrado"
// @Failure      500 {object} map[string]any "Error interno"
// @Router       /tenants/{id} [delete]
func (ctrl *Controller) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := ctrl.service.Delete(c.Request.Context(), id); err != nil {
		ctrl.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (ctrl *Controller) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrTenantNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "TENANT_NOT_FOUND", "message": "Tenant not found"}})
	case errors.Is(err, domain.ErrTenantAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "TENANT_EXISTS", "message": "Tenant already exists"}})
	case errors.Is(err, domain.ErrInvalidTenantData):
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_TENANT_DATA", "message": err.Error()}})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
	}
}
