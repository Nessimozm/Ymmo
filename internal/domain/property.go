package domain

import "time"

// PropertyType distingue les biens résidentiels et professionnels
type PropertyType string

const (
	PropertyTypeApartment  PropertyType = "apartment"  // appartement
	PropertyTypeHouse      PropertyType = "house"      // maison
	PropertyTypeOffice     PropertyType = "office"     // bureau
	PropertyTypeCommercial PropertyType = "commercial" // local commercial
	PropertyTypeLand       PropertyType = "land"       // terrain
)

// PropertyStatus représente l'état de la transaction
type PropertyStatus string

const (
	StatusAvailable PropertyStatus = "available" // disponible
	StatusSold      PropertyStatus = "sold"      // vendu
	StatusRented    PropertyStatus = "rented"    // loué
	StatusPending   PropertyStatus = "pending"   // sous compromis
)

// TransactionType distingue vente et location
type TransactionType string

const (
	TransactionSale   TransactionType = "sale"   // vente
	TransactionRental TransactionType = "rental" // location
)

// Property représente un bien immobilier sur la plateforme
type Property struct {
	ID          uint            `json:"id"               db:"id"`
	Title       string          `json:"title"            db:"title"`
	Description string          `json:"description"      db:"description"`
	Price       float64         `json:"price"            db:"price"`
	Surface     float64         `json:"surface"          db:"surface"`
	Rooms       int             `json:"rooms"            db:"rooms"`
	Bedrooms    int             `json:"bedrooms"         db:"bedrooms"`
	Type        PropertyType    `json:"type"             db:"type"`
	Status      PropertyStatus  `json:"status"           db:"status"`
	Transaction TransactionType `json:"transaction"      db:"transaction"`
	Address     string          `json:"address"          db:"address"`
	City        string          `json:"city"             db:"city"`
	ZipCode     string          `json:"zip_code"         db:"zip_code"`
	Latitude    float64         `json:"latitude"         db:"latitude"`
	Longitude   float64         `json:"longitude"        db:"longitude"`
	AgentID     uint            `json:"agent_id"         db:"agent_id"`
	Agent       *User           `json:"agent,omitempty"`
	Images      []PropertyImage `json:"images,omitempty"`
	ViewCount   int             `json:"view_count"       db:"view_count"`
	CreatedAt   time.Time       `json:"created_at"       db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"       db:"updated_at"`
}

// PropertyImage représente une image attachée à un bien
type PropertyImage struct {
	ID         uint      `json:"id"          db:"id"`
	PropertyID uint      `json:"property_id" db:"property_id"`
	URL        string    `json:"url"         db:"url"`
	IsPrimary  bool      `json:"is_primary"  db:"is_primary"` // image principale (miniature)
	CreatedAt  time.Time `json:"created_at"  db:"created_at"`
}

// ContactMessage représente un message envoyé à un agent
type ContactMessage struct {
	ID         uint      `json:"id"          db:"id"`
	PropertyID uint      `json:"property_id" db:"property_id"`
	SenderID   uint      `json:"sender_id"   db:"sender_id"`
	AgentID    uint      `json:"agent_id"    db:"agent_id"`
	Message    string    `json:"message"     db:"message"`
	IsRead     bool      `json:"is_read"     db:"is_read"`
	CreatedAt  time.Time `json:"created_at"  db:"created_at"`
}
