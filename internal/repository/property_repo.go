package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"ymmo/internal/domain"
)

type propertyRepository struct {
	db *sql.DB
}

// NewPropertyRepository retourne une instance du PropertyRepository
func NewPropertyRepository(db *sql.DB) PropertyRepository {
	return &propertyRepository{db: db}
}

func (r *propertyRepository) Create(p *domain.Property) error {
	query := `
		INSERT INTO properties
			(title, description, price, surface, rooms, bedrooms, type, status,
			 transaction, address, city, zip_code, latitude, longitude, agent_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		p.Title, p.Description, p.Price, p.Surface,
		p.Rooms, p.Bedrooms, p.Type, p.Status,
		p.Transaction, p.Address, p.City, p.ZipCode,
		p.Latitude, p.Longitude, p.AgentID,
	)
	if err != nil {
		return fmt.Errorf("propertyRepository.Create : %w", err)
	}
	id, _ := result.LastInsertId()
	p.ID = uint(id)
	return nil
}

func (r *propertyRepository) GetByID(id uint) (*domain.Property, error) {
	query := `
		SELECT
			p.id, p.title, p.description, p.price, p.surface,
			p.rooms, p.bedrooms, p.type, p.status, p.transaction,
			p.address, p.city, p.zip_code, p.latitude, p.longitude,
			p.agent_id, p.view_count, p.created_at, p.updated_at,
			u.id, u.first_name, u.last_name, u.email, u.phone
		FROM properties p
		JOIN users u ON u.id = p.agent_id
		WHERE p.id = ?`

	p := &domain.Property{Agent: &domain.User{}}
	err := r.db.QueryRow(query, id).Scan(
		&p.ID, &p.Title, &p.Description, &p.Price, &p.Surface,
		&p.Rooms, &p.Bedrooms, &p.Type, &p.Status, &p.Transaction,
		&p.Address, &p.City, &p.ZipCode, &p.Latitude, &p.Longitude,
		&p.AgentID, &p.ViewCount, &p.CreatedAt, &p.UpdatedAt,
		&p.Agent.ID, &p.Agent.FirstName, &p.Agent.LastName,
		&p.Agent.Email, &p.Agent.Phone,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("propertyRepository.GetByID : %w", err)
	}

	// Charge les images
	images, err := r.GetImages(p.ID)
	if err != nil {
		return nil, err
	}
	p.Images = images
	return p, nil
}

// List retourne les biens filtrés avec pagination
func (r *propertyRepository) List(f domain.PropertyFilters) ([]domain.Property, int, error) {
	// Construction dynamique du WHERE
	conditions := []string{"p.status != 'sold'"}
	args := []interface{}{}

	if f.City != "" {
		conditions = append(conditions, "p.city = ?")
		args = append(args, f.City)
	}
	if f.Type != "" {
		conditions = append(conditions, "p.type = ?")
		args = append(args, f.Type)
	}
	if f.Transaction != "" {
		conditions = append(conditions, "p.transaction = ?")
		args = append(args, f.Transaction)
	}
	if f.MinPrice > 0 {
		conditions = append(conditions, "p.price >= ?")
		args = append(args, f.MinPrice)
	}
	if f.MaxPrice > 0 {
		conditions = append(conditions, "p.price <= ?")
		args = append(args, f.MaxPrice)
	}
	if f.MinSurface > 0 {
		conditions = append(conditions, "p.surface >= ?")
		args = append(args, f.MinSurface)
	}
	if f.Rooms > 0 {
		conditions = append(conditions, "p.rooms >= ?")
		args = append(args, f.Rooms)
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	// Compte total pour la pagination
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM properties p %s", where)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("propertyRepository.List count : %w", err)
	}

	// Pagination
	if f.Limit == 0 {
		f.Limit = 12
	}
	if f.Page == 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit
	args = append(args, f.Limit, offset)

	query := fmt.Sprintf(`
		SELECT
			p.id, p.title, p.price, p.surface, p.rooms,
			p.type, p.status, p.transaction,
			p.city, p.zip_code, p.agent_id, p.view_count, p.created_at
		FROM properties p
		%s
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?`, where)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("propertyRepository.List : %w", err)
	}
	defer rows.Close()

	var properties []domain.Property
	for rows.Next() {
		var p domain.Property
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Price, &p.Surface, &p.Rooms,
			&p.Type, &p.Status, &p.Transaction,
			&p.City, &p.ZipCode, &p.AgentID, &p.ViewCount, &p.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		properties = append(properties, p)
	}
	return properties, total, rows.Err()
}

func (r *propertyRepository) Update(p *domain.Property) error {
	query := `
		UPDATE properties
		SET title=?, description=?, price=?, surface=?, rooms=?, bedrooms=?,
		    type=?, status=?, transaction=?, address=?, city=?, zip_code=?
		WHERE id = ? AND agent_id = ?` // sécurité : un agent ne modifie que ses biens

	_, err := r.db.Exec(query,
		p.Title, p.Description, p.Price, p.Surface, p.Rooms, p.Bedrooms,
		p.Type, p.Status, p.Transaction, p.Address, p.City, p.ZipCode,
		p.ID, p.AgentID,
	)
	if err != nil {
		return fmt.Errorf("propertyRepository.Update : %w", err)
	}
	return nil
}

func (r *propertyRepository) Delete(id uint) error {
	_, err := r.db.Exec("DELETE FROM properties WHERE id = ?", id)
	return err
}

func (r *propertyRepository) IncrementViewCount(id uint) error {
	_, err := r.db.Exec("UPDATE properties SET view_count = view_count + 1 WHERE id = ?", id)
	return err
}

func (r *propertyRepository) AddImage(img *domain.PropertyImage) error {
	query := `INSERT INTO property_images (property_id, url, is_primary) VALUES (?, ?, ?)`
	result, err := r.db.Exec(query, img.PropertyID, img.URL, img.IsPrimary)
	if err != nil {
		return fmt.Errorf("propertyRepository.AddImage : %w", err)
	}
	id, _ := result.LastInsertId()
	img.ID = uint(id)
	return nil
}

func (r *propertyRepository) GetImages(propertyID uint) ([]domain.PropertyImage, error) {
	rows, err := r.db.Query(
		"SELECT id, property_id, url, is_primary, created_at FROM property_images WHERE property_id = ?",
		propertyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []domain.PropertyImage
	for rows.Next() {
		var img domain.PropertyImage
		if err := rows.Scan(&img.ID, &img.PropertyID, &img.URL, &img.IsPrimary, &img.CreatedAt); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, rows.Err()
}
