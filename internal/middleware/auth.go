package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"genesis-pay-backend/internal/shared"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			shared.ErrorResponse(c, http.StatusUnauthorized, "Token no proveído")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			shared.ErrorResponse(c, http.StatusUnauthorized, "Formato de token inválido")
			c.Abort()
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			shared.ErrorResponse(c, http.StatusUnauthorized, "Token inválido o expirado")
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("userID", claims["user_id"])
			c.Set("role", claims["role"])
		}

		c.Next()
	}
}
