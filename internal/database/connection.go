package database

import (
	"fmt"
	"log"
	"time"

	"genesis-pay-backend/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewConnection(cfg *config.Config, schema string) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s search_path=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode, schema,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Error conectando a la base de datos en schema %s: %v\n", schema, err)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Error obteniendo objeto sql.DB para schema %s: %v\n", schema, err)
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	log.Printf("Conexión a BD establecida exitosamente apoyando el schema(s): %s\n", schema)

	return db, nil
}
