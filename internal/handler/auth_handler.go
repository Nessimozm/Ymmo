package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ymmo/internal/domain"
	"ymmo/internal/middleware"
	"ymmo/internal/service"
)

// AuthHandler gère les routes /auth
type AuthHandler struct {
	authSvc service.AuthService
}

// NewAuthHandler retourne une instance du AuthHandler
func NewAuthHandler(authSvc service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// Register godoc
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authSvc.Register(&req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Login godoc
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authSvc.Login(&req)
	if err != nil {
		// 401 et non 404 — ne pas distinguer email inconnu / mauvais mdp
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Logout godoc
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	token, exists := c.Get(middleware.ContextToken)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token manquant"})
		return
	}

	if err := h.authSvc.Logout(token.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur lors du logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "déconnexion réussie"})
}

// Me godoc
// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "non authentifié"})
		return
	}

	role, _ := middleware.GetUserRole(c)

	c.JSON(http.StatusOK, gin.H{
		"id":   userID,
		"role": role,
	})
}
