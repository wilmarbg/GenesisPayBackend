package main

import (
	"log"

	"genesis-pay-backend/internal/config"
	"genesis-pay-backend/internal/database"
	"genesis-pay-backend/internal/payments"
	"genesis-pay-backend/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	db, err := database.NewConnection(cfg, "pag,com,cli")
	if err != nil {
		log.Fatalf("Error conectando BD: %v", err)
	}

	cardRepo := payments.NewCardRepository(db)
	txRepo := payments.NewTransactionRepository(db)
	auditRepo := payments.NewAuditLogRepository(db)

	service := payments.NewPaymentService(db, cardRepo, txRepo, auditRepo, cfg.EncryptionKey)
	handler := payments.NewPaymentHandler(service)

	router := gin.Default()
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false
	router.Use(middleware.CORS())

	apiV1 := router.Group("/api/v1")
	payments.RegisterPaymentRoutes(apiV1, handler, cfg.JWTSecret)

	log.Printf("Payments Service iniciado en puerto %s", cfg.PaymentsPort)
	if err := router.Run(":" + cfg.PaymentsPort); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
