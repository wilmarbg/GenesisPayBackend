package middleware

import (
	"github.com/gin-gonic/gin"
)

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {

		roleVal, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(403, gin.H{"success": false, "message": "No tiene permisos para esta acción"})
			return
		}

		userRole, ok := roleVal.(string)
		if !ok {
			c.AbortWithStatusJSON(403, gin.H{"success": false, "message": "No tiene permisos para esta acción"})
			return
		}

		hasPermission := false
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.AbortWithStatusJSON(403, gin.H{"success": false, "message": "No tiene permisos para esta acción"})
			return
		}

		c.Next()
	}
}
