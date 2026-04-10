package main

import (
	"log"

	"genesis-pay-backend/internal/clients"
	"genesis-pay-backend/internal/config"
	"genesis-pay-backend/internal/database"
	"genesis-pay-backend/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	db, err := database.NewConnection(cfg, "cli,pag")
	if err != nil {
		log.Fatalf("Error conectando BD: %v", err)
	}

	log.Println("EncryptionKey length:", len(cfg.EncryptionKey))
	repo := clients.NewClientRepository(db)
	service := clients.NewClientService(repo, cfg.EncryptionKey)
	handler := clients.NewClientHandler(service)

	router := gin.Default()
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false
	router.Use(middleware.CORS())

	api := router.Group("/api/v1/clients")

	clients.RegisterClientRoutes(api, handler, cfg.JWTSecret)

	log.Printf("Clients Service iniciado en puerto %s", cfg.ClientsPort)
	if err := router.Run(":" + cfg.ClientsPort); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
