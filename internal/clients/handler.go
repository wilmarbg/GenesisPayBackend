package clients

import (
	"math"
	"net/http"
	"strconv"

	"genesis-pay-backend/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ClientHandler struct {
	Service *ClientService
}

func NewClientHandler(service *ClientService) *ClientHandler {
	return &ClientHandler{Service: service}
}

func (h *ClientHandler) Create(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, _ := uuid.Parse(userIDStr)

	var req CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Datos obligatorios faltantes o inválidos"})
		return
	}

	client, err := h.Service.Create(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Cliente creado exitosamente", "data": client})
}

func (h *ClientHandler) FindAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	search := c.Query("search")

	clients, totalRows, err := h.Service.FindAll(page, pageSize, status, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error al obtener clientes"})
		return
	}

	totalPages := int(math.Ceil(float64(totalRows) / float64(pageSize)))

	c.JSON(http.StatusOK, PaginatedResponse{
		Success:	true,
		Message:	"Clientes obtenidos exitosamente",
		Data:		clients,
		Page:		page,
		PageSize:	pageSize,
		TotalRows:	totalRows,
		TotalPages:	totalPages,
	})
}

func (h *ClientHandler) GetMe(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, _ := uuid.Parse(userIDStr)

	client, err := h.Service.FindByUserID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "No tiene perfil de cliente"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Perfil de cliente obtenido", "data": client})
}

func (h *ClientHandler) FindByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ID inválido"})
		return
	}

	client, err := h.Service.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
		return
	}

	userIDStr := c.GetString("user_id")
	role := c.GetString("role")
	userID, _ := uuid.Parse(userIDStr)

	if client.UserID != userID && role != "ADMINISTRADOR" {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "No tienes permiso para ver este cliente"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Cliente encontrado", "data": client})
}

func (h *ClientHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ID inválido"})
		return
	}

	var req UpdateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Datos inválidos"})
		return
	}

	userIDStr := c.GetString("user_id")
	userID, _ := uuid.Parse(userIDStr)
	role := c.GetString("role")

	client, err := h.Service.Update(id, req, userID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Cliente actualizado exitosamente", "data": client})
}

func (h *ClientHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ID inválido"})
		return
	}

	role := c.GetString("role")
	if err := h.Service.Delete(id, role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Cliente desactivado exitosamente"})
}

func (h *ClientHandler) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ID inválido"})
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Datos inválidos"})
		return
	}

	role := c.GetString("role")
	adminID, _ := uuid.Parse(c.GetString("user_id"))
	if err := h.Service.UpdateStatus(id, req, role, adminID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Estado actualizado exitosamente"})
}

func (h *ClientHandler) Activate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ID inválido"})
		return
	}

	adminIDStr := c.GetString("user_id")
	adminID, _ := uuid.Parse(adminIDStr)

	res, err := h.Service.Activate(id, adminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Cliente activado y tarjeta emitida exitosamente", "data": res})
}

func RegisterClientRoutes(router *gin.RouterGroup, handler *ClientHandler, jwtSecret string) {
	authMw := middleware.AuthRequired(jwtSecret)
	adminMw := middleware.RequireRole("ADMINISTRADOR")

	router.GET("/me", authMw, handler.GetMe)

	router.POST("", authMw, handler.Create)

	router.GET("", authMw, adminMw, handler.FindAll)

	router.GET("/:id", authMw, handler.FindByID)

	router.PUT("/:id", authMw, handler.Update)

	router.DELETE("/:id", authMw, adminMw, handler.Delete)

	router.PATCH("/:id/status", authMw, adminMw, handler.UpdateStatus)

	router.POST("/:id/activate", authMw, adminMw, handler.Activate)
}
