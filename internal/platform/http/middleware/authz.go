package middleware

import (
	"net/http"

	accessapp "hexagonal-go-backend/internal/modules/access/application"

	"github.com/gin-gonic/gin"
)

// RequirePermission verifies that the authenticated user has the required atomic permission in the project context.
func RequirePermission(authorizer accessapp.Authorizer, permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			unauthorized(c)
			return
		}

		// Extraer project_id desde param de ruta, query param o header
		projectID := c.Param("project_id")
		if projectID == "" {
			projectID = c.Query("project_id")
		}
		if projectID == "" {
			projectID = c.GetHeader("X-Project-ID")
		}

		allowed, err := authorizer.Can(c.Request.Context(), userID, projectID, permissionCode)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to evaluate authorization permissions",
				},
				"request_id": c.GetString("request_id"),
			})
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Insufficient permissions for action: " + permissionCode,
				},
				"request_id": c.GetString("request_id"),
			})
			return
		}

		if projectID != "" {
			c.Set("project_id", projectID)
		}
		c.Next()
	}
}

// RequireSuperAdmin ensures only platform-level super administrators can access the endpoint.
func RequireSuperAdmin(authorizer accessapp.Authorizer) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			unauthorized(c)
			return
		}

		isSuper, err := authorizer.IsSuperAdmin(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to verify super administrator status",
				},
				"request_id": c.GetString("request_id"),
			})
			return
		}

		if !isSuper {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Super administrator privileges required",
				},
				"request_id": c.GetString("request_id"),
			})
			return
		}

		c.Set("is_super_admin", true)
		c.Next()
	}
}
