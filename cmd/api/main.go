package main

import (
	"log"

	"ymmo/configs/config"
	"ymmo/internal/api"
	"ymmo/internal/handler"
	"ymmo/internal/repository"
	"ymmo/internal/service"
	"ymmo/pkg/database"
)

func main() {
	// ── 1. Configuration ──────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Erreur de configuration : %v", err)
	}

	// ── 2. Connexion MySQL ────────────────────────────────────
	db, err := database.Connect(cfg.DSN())
	if err != nil {
		log.Fatalf("❌ Erreur BDD : %v", err)
	}
	defer database.Close()

	// ── 3. Repositories ───────────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	propertyRepo := repository.NewPropertyRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	blacklistRepo := repository.NewTokenBlacklistRepository(db)

	// ── 4. Services ───────────────────────────────────────────
	authSvc := service.NewAuthService(userRepo, blacklistRepo, cfg.JWTSecret, cfg.JWTExpiration)
	userSvc := service.NewUserService(userRepo)
	propertySvc := service.NewPropertyService(propertyRepo)
	messageSvc := service.NewMessageService(messageRepo, propertyRepo)
	statsSvc := service.NewStatsService(propertyRepo, messageRepo)

	// ── 5. Handlers ───────────────────────────────────────────
	authH := handler.NewAuthHandler(authSvc)
	propertyH := handler.NewPropertyHandler(propertySvc)
	adminH := handler.NewAdminHandler(userSvc)
	agentH := handler.NewAgentHandler(messageSvc)
	messageH := handler.NewMessageHandler(messageSvc)
	statsH := handler.NewStatsHandler(statsSvc)

	// ── 6. Routeur ────────────────────────────────────────────
	router := api.NewRouter(authH, propertyH, statsH, adminH, agentH, messageH, authSvc)

	// ── 7. Démarrage ──────────────────────────────────────────
	log.Printf("🚀 Ymmo démarré sur http://localhost:%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("❌ Erreur démarrage serveur : %v", err)
	}
}
