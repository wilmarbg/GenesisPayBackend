package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost			string
	DBPort			string
	DBName			string
	DBUser			string
	DBPassword		string
	DBSSLMode		string
	JWTSecret		string
	JWTExpirationHours	string
	EncryptionKey		string
	AuthPort		string
	ClientsPort		string
	MerchantsPort		string
	PaymentsPort		string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Advertencia: No se encontró archivo .env, usando variables del sistema")
	}

	return &Config{
		DBHost:			os.Getenv("DB_HOST"),
		DBPort:			os.Getenv("DB_PORT"),
		DBName:			os.Getenv("DB_NAME"),
		DBUser:			os.Getenv("DB_USER"),
		DBPassword:		os.Getenv("DB_PASSWORD"),
		DBSSLMode:		os.Getenv("DB_SSL_MODE"),
		JWTSecret:		os.Getenv("JWT_SECRET"),
		JWTExpirationHours:	os.Getenv("JWT_EXPIRATION_HOURS"),
		EncryptionKey:		os.Getenv("ENCRYPTION_KEY"),
		AuthPort:		os.Getenv("AUTH_PORT"),
		ClientsPort:		os.Getenv("CLIENTS_PORT"),
		MerchantsPort:		os.Getenv("MERCHANTS_PORT"),
		PaymentsPort:		os.Getenv("PAYMENTS_PORT"),
	}
}
