package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // driver MySQL
)

// DB est l'instance partagée de la connexion MySQL
var DB *sql.DB

// Connect initialise la connexion MySQL à partir du DSN
// et configure le pool de connexions
func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("impossible d'ouvrir la connexion MySQL : %w", err)
	}

	// ── Pool de connexions ────────────────────────────────────
	db.SetMaxOpenConns(25)                 // max connexions simultanées
	db.SetMaxIdleConns(10)                 // connexions gardées en veille
	db.SetConnMaxLifetime(5 * time.Minute) // durée de vie max d'une connexion

	// ── Vérification que la BDD est bien accessible ───────────
	if err := ping(db); err != nil {
		return nil, err
	}

	DB = db
	log.Println("✅ Connexion MySQL établie")
	return db, nil
}

// ping tente de contacter la BDD avec plusieurs essais
// utile au démarrage si MySQL met quelques secondes à être prêt
func ping(db *sql.DB) error {
	maxRetries := 5
	for i := 1; i <= maxRetries; i++ {
		if err := db.Ping(); err == nil {
			return nil
		}
		log.Printf("⏳ Tentative %d/%d de connexion à MySQL...", i, maxRetries)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("impossible de joindre MySQL après %d tentatives", maxRetries)
}

// Close ferme proprement la connexion (appelé dans main via defer)
func Close() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("⚠️  Erreur lors de la fermeture MySQL : %v", err)
		}
	}
}
