package main

import (
	"log"
	"strconv"

	"genesis-pay-backend/internal/auth"
	"genesis-pay-backend/internal/config"
	"genesis-pay-backend/internal/database"
	"genesis-pay-backend/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	db, err := database.NewConnection(cfg, "aut")
	if err != nil {
		log.Fatalf("Error conectando BD: %v", err)
	}

	repo := auth.NewUserRepository(db)

	expHours, err := strconv.Atoi(cfg.JWTExpirationHours)
	if err != nil {
		expHours = 24
	}

	service := auth.NewAuthService(repo, cfg.JWTSecret, expHours)
	handler := auth.NewAuthHandler(service)

	router := gin.Default()
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false
	router.Use(middleware.CORS())

	api := router.Group("/api/v1/auth")

	auth.RegisterAuthRoutes(api, handler, cfg.JWTSecret)

	log.Printf("Auth Service iniciado en puerto %s", cfg.AuthPort)
	if err := router.Run(":" + cfg.AuthPort); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
