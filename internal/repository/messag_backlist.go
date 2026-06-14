package repository

import (
	"database/sql"
	"fmt"
	"time"

	"ymmo/internal/domain"
)

// ── MessageRepository ─────────────────────────────────────────────────────────

type messageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(msg *domain.ContactMessage) error {
	query := `
		INSERT INTO contact_messages (property_id, client_id, agent_id, sender_id, sender_role, message)
		VALUES (?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		msg.PropertyID, msg.ClientID, msg.AgentID,
		msg.SenderID, msg.SenderRole, msg.Message,
	)
	if err != nil {
		return fmt.Errorf("messageRepository.Create : %w", err)
	}
	id, _ := result.LastInsertId()
	msg.ID = uint(id)
	return nil
}

func (r *messageRepository) GetByID(id uint) (*domain.ContactMessage, error) {
	query := `
		SELECT id, property_id, client_id, agent_id, sender_id, sender_role, message, is_read, created_at
		FROM contact_messages
		WHERE id = ?`

	m := &domain.ContactMessage{}
	err := r.db.QueryRow(query, id).Scan(
		&m.ID, &m.PropertyID, &m.ClientID, &m.AgentID,
		&m.SenderID, &m.SenderRole, &m.Message, &m.IsRead, &m.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("messageRepository.GetByID : %w", err)
	}
	return m, nil
}

// GetByAgentID retourne tous les messages des conversations de cet agent,
// triés pour que les messages d'une même conversation soient regroupés
// et ordonnés chronologiquement (le frontend les groupe par client_id+property_id).
func (r *messageRepository) GetByAgentID(agentID uint) ([]domain.MessageThreadItem, error) {
	query := `
		SELECT
			m.id, m.property_id, m.client_id, m.agent_id, m.sender_id,
			m.sender_role, m.message, m.is_read, m.created_at,
			p.title,
			u.first_name, u.last_name
		FROM contact_messages m
		JOIN properties p ON p.id = m.property_id
		JOIN users u      ON u.id = m.client_id
		WHERE m.agent_id = ?
		ORDER BY m.property_id, m.client_id, m.created_at ASC`

	return r.scanThreadItems(query, agentID)
}

// GetByClientID retourne tous les messages des conversations de ce client,
// avec le nom de l'agent comme "autre partie".
func (r *messageRepository) GetByClientID(clientID uint) ([]domain.MessageThreadItem, error) {
	query := `
		SELECT
			m.id, m.property_id, m.client_id, m.agent_id, m.sender_id,
			m.sender_role, m.message, m.is_read, m.created_at,
			p.title,
			u.first_name, u.last_name
		FROM contact_messages m
		JOIN properties p ON p.id = m.property_id
		JOIN users u      ON u.id = m.agent_id
		WHERE m.client_id = ?
		ORDER BY m.property_id, m.agent_id, m.created_at ASC`

	return r.scanThreadItems(query, clientID)
}

// scanThreadItems factorise le scan commun à GetByAgentID et GetByClientID
func (r *messageRepository) scanThreadItems(query string, id uint) ([]domain.MessageThreadItem, error) {
	rows, err := r.db.Query(query, id)
	if err != nil {
		return nil, fmt.Errorf("messageRepository.scanThreadItems : %w", err)
	}
	defer rows.Close()

	var items []domain.MessageThreadItem
	for rows.Next() {
		var item domain.MessageThreadItem
		var firstName, lastName string

		if err := rows.Scan(
			&item.ID, &item.PropertyID, &item.ClientID, &item.AgentID, &item.SenderID,
			&item.SenderRole, &item.Message, &item.IsRead, &item.CreatedAt,
			&item.PropertyTitle,
			&firstName, &lastName,
		); err != nil {
			return nil, err
		}

		item.OtherPartyName = firstName + " " + lastName
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *messageRepository) MarkAsRead(id uint) error {
	_, err := r.db.Exec("UPDATE contact_messages SET is_read = 1 WHERE id = ?", id)
	return err
}

// CountUnreadForAgent compte les messages envoyés par des clients
// que cet agent n'a pas encore lus.
func (r *messageRepository) CountUnreadForAgent(agentID uint) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM contact_messages
		 WHERE agent_id = ? AND sender_role = 'client' AND is_read = 0`,
		agentID,
	).Scan(&count)
	return count, err
}

// CountUnreadForClient compte les réponses d'agents que ce client
// n'a pas encore lues.
func (r *messageRepository) CountUnreadForClient(clientID uint) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM contact_messages
		 WHERE client_id = ? AND sender_role = 'agent' AND is_read = 0`,
		clientID,
	).Scan(&count)
	return count, err
}

// ── TokenBlacklistRepository ──────────────────────────────────────────────────

type tokenBlacklistRepository struct {
	db *sql.DB
}

func NewTokenBlacklistRepository(db *sql.DB) TokenBlacklistRepository {
	return &tokenBlacklistRepository{db: db}
}

func (r *tokenBlacklistRepository) Add(tokenHash string, expiresAt int64) error {
	query := `
		INSERT INTO token_blacklist (token_hash, expires_at)
		VALUES (?, FROM_UNIXTIME(?))`

	_, err := r.db.Exec(query, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("tokenBlacklistRepository.Add : %w", err)
	}
	return nil
}

func (r *tokenBlacklistRepository) IsBlacklisted(tokenHash string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM token_blacklist WHERE token_hash = ? AND expires_at > NOW()",
		tokenHash,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("tokenBlacklistRepository.IsBlacklisted : %w", err)
	}
	return count > 0, nil
}

// Purge supprime les tokens expirés — à appeler périodiquement
func (r *tokenBlacklistRepository) Purge() error {
	_, err := r.db.Exec(
		"DELETE FROM token_blacklist WHERE expires_at < ?",
		time.Now(),
	)
	return err
}
