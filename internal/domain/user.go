package domain

import "time"

// Role définit les niveaux d'accès de l'application
type Role string

const (
	RoleClient Role = "client"
	RoleAgent  Role = "agent"
	RoleAdmin  Role = "admin"
)

// User représente un utilisateur de la plateforme
type User struct {
	ID        uint      `json:"id"         db:"id"`
	FirstName string    `json:"first_name"  db:"first_name"`
	LastName  string    `json:"last_name"   db:"last_name"`
	Email     string    `json:"email"       db:"email"`
	Password  string    `json:"-"           db:"password"` // jamais sérialisé en JSON
	Role      Role      `json:"role"        db:"role"`
	Phone     string    `json:"phone"       db:"phone"`
	CreatedAt time.Time `json:"created_at"  db:"created_at"`
	UpdatedAt time.Time `json:"updated_at"  db:"updated_at"`
}

// IsAgent retourne true si l'utilisateur est agent ou admin
func (u *User) IsAgent() bool {
	return u.Role == RoleAgent || u.Role == RoleAdmin
}

// IsAdmin retourne true si l'utilisateur est admin
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
