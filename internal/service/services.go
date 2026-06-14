package service

import (
	"fmt"

	"ymmo/internal/domain"
	"ymmo/internal/repository"
)

// ── UserService ───────────────────────────────────────────────────────────────

// UserService gère les opérations sur les utilisateurs (côté admin)
type UserService interface {
	List() ([]domain.User, error)
	UpdateRole(id uint, role domain.Role) error
	Delete(id uint) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) List() ([]domain.User, error) {
	return s.userRepo.List()
}

func (s *userService) UpdateRole(id uint, role domain.Role) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("utilisateur introuvable")
	}
	return s.userRepo.UpdateRole(id, role)
}

func (s *userService) Delete(id uint) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("utilisateur introuvable")
	}
	return s.userRepo.Delete(id)
}

// ── MessageService ────────────────────────────────────────────────────────────

// MessageService gère les conversations entre clients et agents.
// Une conversation est identifiée par (PropertyID, ClientID, AgentID) et
// peut contenir des messages des deux côtés (SenderRole = client | agent).
type MessageService interface {
	// Send crée le PREMIER message d'une conversation (toujours envoyé par un client).
	Send(propertyID uint, senderID uint, message string) error

	// Reply ajoute un message à une conversation existante. userID peut être
	// le client OU l'agent de la conversation — le rôle est déduit automatiquement.
	Reply(messageID uint, userID uint, message string) error

	// GetThreadsForAgent retourne tous les messages des conversations de cet agent.
	GetThreadsForAgent(agentID uint) ([]domain.MessageThreadItem, error)

	// GetThreadsForClient retourne tous les messages des conversations de ce client.
	GetThreadsForClient(clientID uint) ([]domain.MessageThreadItem, error)

	// MarkAsRead marque un message comme lu — uniquement par son destinataire.
	MarkAsRead(messageID uint, userID uint) error

	CountUnreadForAgent(agentID uint) (int, error)
	CountUnreadForClient(clientID uint) (int, error)
}

type messageService struct {
	messageRepo  repository.MessageRepository
	propertyRepo repository.PropertyRepository
}

func NewMessageService(
	messageRepo repository.MessageRepository,
	propertyRepo repository.PropertyRepository,
) MessageService {
	return &messageService{
		messageRepo:  messageRepo,
		propertyRepo: propertyRepo,
	}
}

func (s *messageService) Send(propertyID uint, senderID uint, message string) error {
	// Récupère le bien pour trouver l'agent destinataire
	property, err := s.propertyRepo.GetByID(propertyID)
	if err != nil {
		return err
	}
	if property == nil {
		return fmt.Errorf("bien introuvable")
	}

	// Un agent ne peut pas démarrer une conversation client sur sa propre annonce
	if senderID == property.AgentID {
		return fmt.Errorf("vous ne pouvez pas contacter votre propre annonce")
	}

	msg := &domain.ContactMessage{
		PropertyID: propertyID,
		ClientID:   senderID, // l'expéditeur initial est le "client" de cette conversation
		AgentID:    property.AgentID,
		SenderID:   senderID,
		SenderRole: domain.RoleClient,
		Message:    message,
	}

	return s.messageRepo.Create(msg)
}

func (s *messageService) Reply(messageID uint, userID uint, message string) error {
	// Récupère un message existant de la conversation pour en connaître
	// le client_id et l'agent_id (constants sur tout le fil)
	original, err := s.messageRepo.GetByID(messageID)
	if err != nil {
		return err
	}
	if original == nil {
		return fmt.Errorf("conversation introuvable")
	}

	// Déduit le rôle de l'utilisateur courant dans CETTE conversation
	var role domain.Role
	switch userID {
	case original.AgentID:
		role = domain.RoleAgent
	case original.ClientID:
		role = domain.RoleClient
	default:
		return fmt.Errorf("accès refusé")
	}

	reply := &domain.ContactMessage{
		PropertyID: original.PropertyID,
		ClientID:   original.ClientID,
		AgentID:    original.AgentID,
		SenderID:   userID,
		SenderRole: role,
		Message:    message,
	}

	return s.messageRepo.Create(reply)
}

func (s *messageService) GetThreadsForAgent(agentID uint) ([]domain.MessageThreadItem, error) {
	return s.messageRepo.GetByAgentID(agentID)
}

func (s *messageService) GetThreadsForClient(clientID uint) ([]domain.MessageThreadItem, error) {
	return s.messageRepo.GetByClientID(clientID)
}

func (s *messageService) MarkAsRead(messageID uint, userID uint) error {
	msg, err := s.messageRepo.GetByID(messageID)
	if err != nil {
		return err
	}
	if msg == nil {
		return fmt.Errorf("message introuvable")
	}

	// Le destinataire d'un message est l'AUTRE partie de la conversation :
	// un message envoyé par le client doit être marqué lu par l'agent, et inversement.
	isRecipient := (msg.SenderRole == domain.RoleClient && msg.AgentID == userID) ||
		(msg.SenderRole == domain.RoleAgent && msg.ClientID == userID)

	if !isRecipient {
		return fmt.Errorf("accès refusé")
	}

	return s.messageRepo.MarkAsRead(messageID)
}

func (s *messageService) CountUnreadForAgent(agentID uint) (int, error) {
	return s.messageRepo.CountUnreadForAgent(agentID)
}

func (s *messageService) CountUnreadForClient(clientID uint) (int, error) {
	return s.messageRepo.CountUnreadForClient(clientID)
}

// ── StatsService ──────────────────────────────────────────────────────────────

// StatsService calcule les statistiques du marché immobilier
type StatsService interface {
	GetMarketStats() ([]domain.MarketStats, error)
	GetPopularProperties(limit int) ([]domain.Property, error)
	GetDashboard() (*domain.DashboardStats, error)
}

type statsService struct {
	propertyRepo repository.PropertyRepository
	messageRepo  repository.MessageRepository
}

func NewStatsService(
	propertyRepo repository.PropertyRepository,
	messageRepo repository.MessageRepository,
) StatsService {
	return &statsService{
		propertyRepo: propertyRepo,
		messageRepo:  messageRepo,
	}
}

func (s *statsService) GetMarketStats() ([]domain.MarketStats, error) {
	// Utilise les filtres existants pour agréger par ville
	// On récupère toutes les villes disponibles et on calcule les stats
	filters := domain.PropertyFilters{Limit: 1000, Page: 1}
	properties, _, err := s.propertyRepo.List(filters)
	if err != nil {
		return nil, err
	}

	// Agrégation en mémoire par ville
	cityMap := map[string]*domain.MarketStats{}
	for _, p := range properties {
		if _, ok := cityMap[p.City]; !ok {
			cityMap[p.City] = &domain.MarketStats{City: p.City}
		}
		stat := cityMap[p.City]
		stat.TotalListings++
		stat.AvgPrice = (stat.AvgPrice*float64(stat.TotalListings-1) + p.Price) / float64(stat.TotalListings)
		if p.Surface > 0 {
			stat.AvgPriceM2 = stat.AvgPrice / p.Surface
		}
	}

	stats := make([]domain.MarketStats, 0, len(cityMap))
	for _, s := range cityMap {
		stats = append(stats, *s)
	}
	return stats, nil
}

func (s *statsService) GetPopularProperties(limit int) ([]domain.Property, error) {
	// Trie par view_count décroissant
	filters := domain.PropertyFilters{Limit: limit, Page: 1}
	properties, _, err := s.propertyRepo.List(filters)
	if err != nil {
		return nil, err
	}
	return properties, nil
}

func (s *statsService) GetDashboard() (*domain.DashboardStats, error) {
	// Stats globales
	allFilters := domain.PropertyFilters{Limit: 1000, Page: 1}
	all, total, err := s.propertyRepo.List(allFilters)
	if err != nil {
		return nil, err
	}

	// Compte les disponibles
	available := 0
	for _, p := range all {
		if p.Status == domain.StatusAvailable {
			available++
		}
	}

	// Top 5 biens populaires
	popular := all
	if len(popular) > 5 {
		popular = popular[:5]
	}

	// Stats de marché
	marketStats, err := s.GetMarketStats()
	if err != nil {
		return nil, err
	}

	dashboard := &domain.DashboardStats{
		TotalProperties:   total,
		AvailableCount:    available,
		TopCities:         marketStats,
		PopularProperties: popular,
	}

	return dashboard, nil
}
