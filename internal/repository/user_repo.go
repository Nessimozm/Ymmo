package repository

import (
	"database/sql"
	"fmt"

	"ymmo/internal/domain"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository retourne une instance du UserRepository
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (first_name, last_name, email, password, role, phone)
		VALUES (?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password,
		user.Role,
		user.Phone,
	)
	if err != nil {
		return fmt.Errorf("userRepository.Create : %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = uint(id)
	return nil
}

func (r *userRepository) GetByID(id uint) (*domain.User, error) {
	query := `
		SELECT id, first_name, last_name, email, password, role, phone, created_at, updated_at
		FROM users
		WHERE id = ?`

	user := &domain.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Phone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // pas trouvé → nil sans erreur
	}
	if err != nil {
		return nil, fmt.Errorf("userRepository.GetByID : %w", err)
	}
	return user, nil
}

func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	query := `
		SELECT id, first_name, last_name, email, password, role, phone, created_at, updated_at
		FROM users
		WHERE email = ?`

	user := &domain.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Phone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("userRepository.GetByEmail : %w", err)
	}
	return user, nil
}

func (r *userRepository) Update(user *domain.User) error {
	query := `
		UPDATE users
		SET first_name = ?, last_name = ?, phone = ?
		WHERE id = ?`

	_, err := r.db.Exec(query, user.FirstName, user.LastName, user.Phone, user.ID)
	if err != nil {
		return fmt.Errorf("userRepository.Update : %w", err)
	}
	return nil
}

func (r *userRepository) Delete(id uint) error {
	_, err := r.db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("userRepository.Delete : %w", err)
	}
	return nil
}

func (r *userRepository) List() ([]domain.User, error) {
	query := `
		SELECT id, first_name, last_name, email, role, phone, created_at, updated_at
		FROM users
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("userRepository.List : %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.FirstName, &u.LastName,
			&u.Email, &u.Role, &u.Phone,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *userRepository) UpdateRole(id uint, role domain.Role) error {
	_, err := r.db.Exec("UPDATE users SET role = ? WHERE id = ?", role, id)
	if err != nil {
		return fmt.Errorf("userRepository.UpdateRole : %w", err)
	}
	return nil
}
