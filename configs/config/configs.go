package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config contient toute la configuration de l'application
type Config struct {
	// Serveur
	Port string

	// Base de données
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// JWT
	JWTSecret     string
	JWTExpiration int // en heures
}

// Load charge le fichier .env et retourne la config
func Load() (*Config, error) {
	// Charge le .env uniquement si le fichier existe (en prod les vars sont déjà injectées)
	_ = godotenv.Load()

	cfg := &Config{
		Port:       getEnv("PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "ymmo"),
		JWTSecret:  getEnv("JWT_SECRET", ""),
	}

	// JWT_SECRET obligatoire
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET est obligatoire dans le fichier .env")
	}

	// JWTExpiration optionnel, défaut 24h
	expStr := getEnv("JWT_EXPIRATION_HOURS", "24")
	exp, err := strconv.Atoi(expStr)
	if err != nil {
		return nil, fmt.Errorf("JWT_EXPIRATION_HOURS doit être un entier : %w", err)
	}
	cfg.JWTExpiration = exp

	return cfg, nil
}

// DSN retourne la chaîne de connexion MySQL (Data Source Name)
func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}

// getEnv lit une variable d'environnement avec une valeur par défaut
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
