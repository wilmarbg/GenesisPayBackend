package main

import (
	"log"

	"genesis-pay-backend/internal/config"
	"genesis-pay-backend/internal/database"
	"genesis-pay-backend/internal/merchants"
	"genesis-pay-backend/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	db, err := database.NewConnection(cfg, "com")
	if err != nil {
		log.Fatalf("Error conectando BD: %v", err)
	}

	merchantRepo := merchants.NewMerchantRepository(db)
	productRepo := merchants.NewProductRepository(db)

	merchantService := merchants.NewMerchantService(merchantRepo)
	productService := merchants.NewProductService(productRepo, merchantRepo)

	handler := merchants.NewMerchantHandler(merchantService, productService)

	router := gin.Default()
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false
	router.Use(middleware.CORS())

	api := router.Group("/api/v1/merchants")
	merchants.RegisterMerchantRoutes(api, handler, cfg.JWTSecret)

	log.Printf("Merchants Service iniciado en puerto %s", cfg.MerchantsPort)
	if err := router.Run(":" + cfg.MerchantsPort); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
