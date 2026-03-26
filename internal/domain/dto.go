package domain

//  Auth

// RegisterRequest payload pour POST /auth/register
type RegisterRequest struct {
	FirstName string `json:"first_name" binding:"required,min=2"`
	LastName  string `json:"last_name"  binding:"required,min=2"`
	Email     string `json:"email"      binding:"required,email"`
	Password  string `json:"password"   binding:"required,min=8"`
	Phone     string `json:"phone"      binding:"omitempty,e164"`
}

// LoginRequest payload pour POST /auth/login
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse retourné après login ou register
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// CreatePropertyRequest payload pour POST /properties
type CreatePropertyRequest struct {
	Title       string          `json:"title"       binding:"required,min=5"`
	Description string          `json:"description" binding:"required,min=20"`
	Price       float64         `json:"price"       binding:"required,gt=0"`
	Surface     float64         `json:"surface"     binding:"required,gt=0"`
	Rooms       int             `json:"rooms"       binding:"required,gte=1"`
	Bedrooms    int             `json:"bedrooms"    binding:"gte=0"`
	Type        PropertyType    `json:"type"        binding:"required"`
	Transaction TransactionType `json:"transaction" binding:"required"`
	Address     string          `json:"address"     binding:"required"`
	City        string          `json:"city"        binding:"required"`
	ZipCode     string          `json:"zip_code"    binding:"required,len=5"`
	Latitude    float64         `json:"latitude"    binding:"omitempty"`
	Longitude   float64         `json:"longitude"   binding:"omitempty"`
}

// UpdatePropertyRequest payload pour PUT /properties/:id
// Tous les champs sont optionnels → seuls les champs envoyés sont mis à jour
type UpdatePropertyRequest struct {
	Title       *string          `json:"title"`
	Description *string          `json:"description"`
	Price       *float64         `json:"price"`
	Surface     *float64         `json:"surface"`
	Rooms       *int             `json:"rooms"`
	Bedrooms    *int             `json:"bedrooms"`
	Status      *PropertyStatus  `json:"status"`
	Transaction *TransactionType `json:"transaction"`
	Address     *string          `json:"address"`
	City        *string          `json:"city"`
	ZipCode     *string          `json:"zip_code"`
}

// PropertyFilters paramètres de filtrage pour GET /properties
type PropertyFilters struct {
	City        string          `form:"city"`
	Type        PropertyType    `form:"type"`
	Transaction TransactionType `form:"transaction"`
	MinPrice    float64         `form:"min_price"`
	MaxPrice    float64         `form:"max_price"`
	MinSurface  float64         `form:"min_surface"`
	Rooms       int             `form:"rooms"`
	Page        int             `form:"page,default=1"`
	Limit       int             `form:"limit,default=12"`
}

// PropertyListResponse réponse paginée pour GET /properties
type PropertyListResponse struct {
	Data       []Property `json:"data"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	TotalPages int        `json:"total_pages"`
}

//  Contact

// ContactRequest payload pour POST /properties/:id/contact
type ContactRequest struct {
	Message string `json:"message" binding:"required,min=10"`
}

//  Admin

// UpdateRoleRequest payload pour PATCH /admin/users/:id/role
type UpdateRoleRequest struct {
	Role Role `json:"role" binding:"required,oneof=client agent admin"`
}

// Stats

// MarketStats retourné par GET /stats/market
type MarketStats struct {
	City          string  `json:"city"`
	AvgPrice      float64 `json:"avg_price"`
	AvgPriceM2    float64 `json:"avg_price_m2"`
	TotalListings int     `json:"total_listings"`
}

// DashboardStats retourné par GET /stats/dashboard
type DashboardStats struct {
	TotalProperties   int           `json:"total_properties"`
	AvailableCount    int           `json:"available_count"`
	SoldThisMonth     int           `json:"sold_this_month"`
	TotalMessages     int           `json:"total_messages"`
	UnreadMessages    int           `json:"unread_messages"`
	TopCities         []MarketStats `json:"top_cities"`
	PopularProperties []Property    `json:"popular_properties"`
}
