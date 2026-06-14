package service

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"ymmo/internal/domain"
	"ymmo/internal/repository"
)

// AuthService gère l'inscription, la connexion et la validation JWT
type AuthService interface {
	Register(req *domain.RegisterRequest) (*domain.AuthResponse, error)
	Login(req *domain.LoginRequest) (*domain.AuthResponse, error)
	ValidateToken(tokenString string) (*JWTClaims, error)
	Logout(tokenString string) error
}

// JWTClaims contient les données embarquées dans le token
type JWTClaims struct {
	UserID uint        `json:"user_id"`
	Role   domain.Role `json:"role"`
	jwt.RegisteredClaims
}

type authService struct {
	userRepo      repository.UserRepository
	blacklistRepo repository.TokenBlacklistRepository
	jwtSecret     string
	jwtExpiration int // heures
}

// NewAuthService retourne une instance du AuthService
func NewAuthService(
	userRepo repository.UserRepository,
	blacklistRepo repository.TokenBlacklistRepository,
	jwtSecret string,
	jwtExpiration int,
) AuthService {
	return &authService{
		userRepo:      userRepo,
		blacklistRepo: blacklistRepo,
		jwtSecret:     jwtSecret,
		jwtExpiration: jwtExpiration,
	}
}

// Register crée un nouveau compte client
func (s *authService) Register(req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	// Vérifie que l'email n'est pas déjà utilisé
	existing, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("authService.Register : %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("cet email est déjà utilisé")
	}

	// Hash du mot de passe avec bcrypt (cost=12 — bon équilibre sécurité/perf)
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("erreur lors du hash du mot de passe : %w", err)
	}

	user := &domain.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  string(hashed),
		Role:      domain.RoleClient, // toujours client à l'inscription
		Phone:     req.Phone,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("authService.Register create : %w", err)
	}

	// Génère le JWT directement après inscription
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{Token: token, User: *user}, nil
}

// Login vérifie les credentials et retourne un JWT
func (s *authService) Login(req *domain.LoginRequest) (*domain.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("authService.Login : %w", err)
	}

	// Message générique volontairement — ne pas révéler si c'est l'email ou le mdp qui est faux
	if user == nil {
		return nil, fmt.Errorf("identifiants invalides")
	}

	// Compare le mot de passe fourni avec le hash bcrypt stocké
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("identifiants invalides")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{Token: token, User: *user}, nil
}

// ValidateToken vérifie et décode un JWT
func (s *authService) ValidateToken(tokenString string) (*JWTClaims, error) {
	// Vérifie si le token est blacklisté (logout effectué)
	tokenHash := hashToken(tokenString)
	blacklisted, err := s.blacklistRepo.IsBlacklisted(tokenHash)
	if err != nil {
		return nil, fmt.Errorf("erreur vérification blacklist : %w", err)
	}
	if blacklisted {
		return nil, fmt.Errorf("token invalide")
	}

	// Parse et vérifie la signature JWT
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		// Vérifie que l'algorithme est bien HMAC — évite l'attaque "alg: none"
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("algorithme de signature inattendu : %v", t.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("token invalide ou expiré")
	}

	return claims, nil
}

// Logout ajoute le token à la blacklist jusqu'à son expiration naturelle
func (s *authService) Logout(tokenString string) error {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return err
	}

	tokenHash := hashToken(tokenString)
	expiresAt := claims.ExpiresAt.Unix()

	return s.blacklistRepo.Add(tokenHash, expiresAt)
}

// generateToken crée un JWT signé pour un utilisateur
func (s *authService) generateToken(user *domain.User) (string, error) {
	expiration := time.Now().Add(time.Duration(s.jwtExpiration) * time.Hour)

	claims := JWTClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("erreur signature JWT : %w", err)
	}

	return signed, nil
}

// hashToken retourne un SHA256 hex du token
// On ne stocke jamais le token brut en BDD
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h)
}
