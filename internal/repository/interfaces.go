package repository

import "ymmo/internal/domain"

// UserRepository définit les opérations BDD sur les utilisateurs
type UserRepository interface {
	Create(user *domain.User) error
	GetByID(id uint) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	Update(user *domain.User) error
	Delete(id uint) error
	List() ([]domain.User, error)
	UpdateRole(id uint, role domain.Role) error
}

// PropertyRepository définit les opérations BDD sur les biens
type PropertyRepository interface {
	Create(property *domain.Property) error
	GetByID(id uint) (*domain.Property, error)
	List(filters domain.PropertyFilters) ([]domain.Property, int, error) // retourne aussi le total pour la pagination
	Update(property *domain.Property) error
	Delete(id uint) error
	IncrementViewCount(id uint) error
	AddImage(image *domain.PropertyImage) error
	GetImages(propertyID uint) ([]domain.PropertyImage, error)
}

// MessageRepository définit les opérations BDD sur les messages
type MessageRepository interface {
	Create(msg *domain.ContactMessage) error
	GetByID(id uint) (*domain.ContactMessage, error)

	// GetByAgentID retourne tous les messages des conversations où
	// l'utilisateur est l'agent, enrichis du titre du bien et du nom du client.
	GetByAgentID(agentID uint) ([]domain.MessageThreadItem, error)

	// GetByClientID retourne tous les messages des conversations où
	// l'utilisateur est le client, enrichis du titre du bien et du nom de l'agent.
	GetByClientID(clientID uint) ([]domain.MessageThreadItem, error)

	MarkAsRead(id uint) error
	CountUnreadForAgent(agentID uint) (int, error)
	CountUnreadForClient(clientID uint) (int, error)
}

// TokenBlacklistRepository gère les JWT invalidés
type TokenBlacklistRepository interface {
	Add(tokenHash string, expiresAt int64) error
	IsBlacklisted(tokenHash string) (bool, error)
	Purge() error // supprime les tokens expirés
}
