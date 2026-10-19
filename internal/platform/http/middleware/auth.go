package middleware

import (
	"net/http"
	"strings"

	auth "hexagonal-go-backend/internal/modules/auth/application"
	usersapp "hexagonal-go-backend/internal/modules/users/application"

	"github.com/gin-gonic/gin"
)

func Authenticate(tokens auth.TokenProvider, userRepo usersapp.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			unauthorized(c, "Authorization header is missing")
			return
		}

		tokenStr := authHeader
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			tokenStr = strings.TrimSpace(authHeader[7:])
		}

		claims, err := tokens.Parse(tokenStr)
		if err != nil {
			unauthorized(c, "Invalid or expired token")
			return
		}

		userID := claims.Subject
		isSuperAdmin := false

		// Lectura únicamente: el usuario local lo provisiona el flujo de login.
		if userRepo != nil {
			ctx := c.Request.Context()
			user, err := userRepo.FindByAuthID(ctx, claims.Subject)
			if err != nil {
				// Tokens locales (HMAC) llevan el ID interno como subject.
				user, err = userRepo.FindByID(ctx, claims.Subject)
			}
			if err != nil || user == nil || !user.Active {
				unauthorized(c, "User account not found or disabled in database")
				return
			}
			userID = user.ID
			isSuperAdmin = user.IsSuperAdmin
		}

		c.Set("user_id", userID)
		c.Set("auth_id", claims.Subject)
		c.Set("subject", userID)
		c.Set("is_super_admin", isSuperAdmin)
		c.Next()
	}
}

func RequireSuperAdminDirect() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !c.GetBool("is_super_admin") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "FORBIDDEN", "message": "Super admin privilege required"}, "request_id": c.GetString("request_id")})
			return
		}
		c.Next()
	}
}

func unauthorized(c *gin.Context, reason ...string) {
	msg := "Missing or invalid token"
	if len(reason) > 0 && reason[0] != "" {
		msg = reason[0]
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": gin.H{
			"code":    "UNAUTHORIZED",
			"message": msg,
		},
		"request_id": c.GetString("request_id"),
	})
}
