package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"ymmo/internal/domain"
	"ymmo/internal/service"
)

// clés de contexte Gin
const (
	ContextUserID = "userID"
	ContextRole   = "userRole"
	ContextToken  = "token"
)

// JWTAuth vérifie la présence et la validité du token JWT
// Utilisation : router.Use(middleware.JWTAuth())
func JWTAuth(authSvc service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Récupère le token depuis le header Authorization: Bearer <token>
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token manquant"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Valide le token via le AuthService
		claims, err := authSvc.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token invalide ou expiré"})
			c.Abort()
			return
		}

		// Injecte les infos dans le contexte Gin — accessibles dans les handlers
		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextRole, claims.Role)
		c.Set(ContextToken, tokenString)

		c.Next()
	}
}

// RequireRole vérifie que l'utilisateur connecté a l'un des rôles autorisés
// Utilisation : router.Use(middleware.RequireRole("agent", "admin"))
func RequireRole(roles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextRole)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "non authentifié"})
			c.Abort()
			return
		}

		userRole, ok := role.(domain.Role)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur de rôle"})
			c.Abort()
			return
		}

		// Vérifie si le rôle de l'utilisateur est dans la liste autorisée
		for _, r := range roles {
			if userRole == r {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "accès refusé"})
		c.Abort()
	}
}

// CORS autorise les requêtes cross-origin (nécessaire pour le frontend)
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		// Répond directement aux preflight OPTIONS
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// Logger log chaque requête avec méthode, path et status
func Logger() gin.HandlerFunc {
	return gin.Logger()
}

// ── Helpers accessibles depuis les handlers ───────────────────────────────────

// GetUserID extrait l'ID utilisateur du contexte Gin
func GetUserID(c *gin.Context) (uint, bool) {
	val, exists := c.Get(ContextUserID)
	if !exists {
		return 0, false
	}
	id, ok := val.(uint)
	return id, ok
}

// GetUserRole extrait le rôle utilisateur du contexte Gin
func GetUserRole(c *gin.Context) (domain.Role, bool) {
	val, exists := c.Get(ContextRole)
	if !exists {
		return "", false
	}
	role, ok := val.(domain.Role)
	return role, ok
}
