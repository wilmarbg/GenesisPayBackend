package payments

import (
	"net/http"

	"genesis-pay-backend/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	Service *PaymentService
}

func NewPaymentHandler(s *PaymentService) *PaymentHandler {
	return &PaymentHandler{Service: s}
}

func (h *PaymentHandler) IssueCard(c *gin.Context) {
	var req IssueCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Datos obligatorios faltantes"})
		return
	}

	adminID, _ := uuid.Parse(c.GetString("user_id"))
	role := c.GetString("role")

	res, err := h.Service.IssueCard(req, adminID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Tarjeta emitida exitosamente", "data": res})
}

func (h *PaymentHandler) GetMyCards(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("user_id"))

	cards, err := h.Service.GetMyCards(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": cards})
}

func (h *PaymentHandler) GetCardBalance(c *gin.Context) {
	cardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ID inválido"})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))
	role := c.GetString("role")

	balance, err := h.Service.GetCardBalance(cardID, userID, role)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"balance": balance}})
}

func (h *PaymentHandler) FreezeCard(c *gin.Context) {
	h.updateCardStatusBase(c, "CONGELADA")
}

func (h *PaymentHandler) CancelCard(c *gin.Context) {
	h.updateCardStatusBase(c, "CANCELADA")
}

func (h *PaymentHandler) updateCardStatusBase(c *gin.Context, status string) {
	cardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ID inválido"})
		return
	}

	adminID, _ := uuid.Parse(c.GetString("user_id"))
	role := c.GetString("role")

	err = h.Service.UpdateCardStatus(cardID, status, adminID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Estado de tarjeta actualizado a " + status})
}

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	var req ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Datos inválidos"})
		return
	}

	clientID, _ := uuid.Parse(c.GetString("user_id"))

	res, err := h.Service.ProcessPayment(req, clientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Pago procesado exitosamente", "data": res})
}

func (h *PaymentHandler) GetTransactions(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("user_id"))
	role := c.GetString("role")

	txs, err := h.Service.GetTransactions(userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": txs})
}

func (h *PaymentHandler) GetTransaction(c *gin.Context) {
	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ID inválido"})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))
	role := c.GetString("role")

	tx, err := h.Service.GetTransactionByID(txID, userID, role)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": tx})
}

func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ID inválido"})
		return
	}

	adminID, _ := uuid.Parse(c.GetString("user_id"))
	role := c.GetString("role")

	res, err := h.Service.RefundTransaction(txID, adminID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Reembolso procesado", "data": res})
}

func RegisterPaymentRoutes(apiV1 *gin.RouterGroup, handler *PaymentHandler, jwtSecret string) {
	authMw := middleware.AuthRequired(jwtSecret)
	adminMw := middleware.RequireRole("ADMINISTRADOR")

	cards := apiV1.Group("/cards")
	cards.Use(authMw)
	{
		cards.POST("", adminMw, handler.IssueCard)
		cards.GET("/me", handler.GetMyCards)
		cards.GET("/:id/balance", handler.GetCardBalance)
		cards.PATCH("/:id/freeze", adminMw, handler.FreezeCard)
		cards.PATCH("/:id/cancel", adminMw, handler.CancelCard)
	}

	payments := apiV1.Group("/payments")
	payments.Use(authMw)
	{
		payments.POST("", handler.ProcessPayment)
		payments.GET("", handler.GetTransactions)
		payments.GET("/:id", handler.GetTransaction)
		payments.POST("/:id/refund", adminMw, handler.RefundPayment)
	}
}
