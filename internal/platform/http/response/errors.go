package response

import (
	"errors"
	"net/http"

	authdomain "hexagonal-go-backend/internal/modules/auth/domain"
	userdomain "hexagonal-go-backend/internal/modules/users/domain"

	"github.com/gin-gonic/gin"
)

func WriteError(c *gin.Context, err error) {
	status, code, message := http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error"
	switch {
	case errors.Is(err, userdomain.ErrUserNotFound):
		status, code, message = http.StatusNotFound, "USER_NOT_FOUND", err.Error()
	case errors.Is(err, userdomain.ErrEmailAlreadyExists):
		status, code, message = http.StatusConflict, "EMAIL_ALREADY_EXISTS", err.Error()
	case errors.Is(err, authdomain.ErrUnauthorized), errors.Is(err, authdomain.ErrInvalidToken):
		status, code, message = http.StatusUnauthorized, "UNAUTHORIZED", "Invalid credentials or token"
	case errors.Is(err, authdomain.ErrInvalidCredentials):
		status, code, message = http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid credentials"
	case errors.Is(err, authdomain.ErrAccountLocked):
		status, code, message = http.StatusLocked, "ACCOUNT_LOCKED", "Account temporarily locked"
	case errors.Is(err, authdomain.ErrAccountDisabled):
		status, code, message = http.StatusForbidden, "ACCOUNT_DISABLED", "Account disabled"
	case errors.Is(err, authdomain.ErrIdentityProviderUnavailable):
		status, code, message = http.StatusServiceUnavailable, "IDENTITY_PROVIDER_UNAVAILABLE", "Authentication service unavailable"
	case errors.Is(err, authdomain.ErrForbidden):
		status, code, message = http.StatusForbidden, "FORBIDDEN", err.Error()
	case errors.Is(err, userdomain.ErrInvalidInput):
		status, code, message = http.StatusUnprocessableEntity, "INVALID_INPUT", err.Error()
	}
	c.JSON(status, ErrorResponse{Error: ErrorBody{Code: code, Message: message}, RequestID: c.GetString("request_id")})
}

func ValidationError(c *gin.Context, _ error) {
	c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorBody{Code: "VALIDATION_ERROR", Message: "Invalid request"}, RequestID: c.GetString("request_id")})
}
