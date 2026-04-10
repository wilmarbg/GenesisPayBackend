package auth

import (
	"net/http"

	"genesis-pay-backend/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	Service *AuthService
}

func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{Service: service}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Datos inválidos"})
		return
	}

	res, err := h.Service.Register(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": res})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Datos inválidos"})
		return
	}

	res, err := h.Service.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {

	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Usuario no autenticado"})
		return
	}

	user, err := h.Service.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": user})
}

func RegisterAuthRoutes(router *gin.RouterGroup, handler *AuthHandler, jwtSecret string) {

	router.POST("/register", handler.Register)

	router.POST("/login", handler.Login)

	router.GET("/me", middleware.AuthRequired(jwtSecret), handler.GetProfile)
}
