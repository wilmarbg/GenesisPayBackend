package middleware

import (
	"net/http"

	"genesis-pay-backend/internal/shared"
	"github.com/gin-gonic/gin"
)

func RoleRequired(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			shared.ErrorResponse(c, http.StatusForbidden, "No se encontró rol en el token")
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			shared.ErrorResponse(c, http.StatusForbidden, "Formato de rol inválido")
			c.Abort()
			return
		}

		allowed := false
		for _, r := range allowedRoles {
			if r == userRole {
				allowed = true
				break
			}
		}

		if !allowed {
			shared.ErrorResponse(c, http.StatusForbidden, shared.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}
