package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"success": false, "message": "Token inválido o expirado"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(401, gin.H{"success": false, "message": "Token inválido o expirado"})
			return
		}
		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(401, gin.H{"success": false, "message": "Token inválido o expirado"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			userID := claims["user_id"]
			email := claims["email"]
			role := claims["role"]

			c.Set("user_id", userID)
			c.Set("email", email)
			c.Set("role", role)
		} else {
			c.AbortWithStatusJSON(401, gin.H{"success": false, "message": "Token inválido o expirado"})
			return
		}

		c.Next()
	}
}
