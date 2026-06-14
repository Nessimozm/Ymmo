package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ymmo/internal/domain"
	"ymmo/internal/middleware"
	"ymmo/internal/service"
)

// ── Admin ─────────────────────────────────────────────────────────────────────

// AdminHandler gère les routes /admin
type AdminHandler struct {
	userSvc service.UserService
}

func NewAdminHandler(userSvc service.UserService) *AdminHandler {
	return &AdminHandler{userSvc: userSvc}
}

// ListUsers godoc
// GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	users, err := h.userSvc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// UpdateRole godoc
// PATCH /api/v1/admin/users/:id/role
func (h *AdminHandler) UpdateRole(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}

	var req domain.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userSvc.UpdateRole(id, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rôle mis à jour"})
}

// DeleteUser godoc
// DELETE /api/v1/admin/users/:id
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}

	if err := h.userSvc.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "utilisateur supprimé"})
}

// ── Stats ─────────────────────────────────────────────────────────────────────

// StatsHandler gère les routes /stats
type StatsHandler struct {
	statsSvc service.StatsService
}

func NewStatsHandler(statsSvc service.StatsService) *StatsHandler {
	return &StatsHandler{statsSvc: statsSvc}
}

// Market godoc
// GET /api/v1/stats/market
func (h *StatsHandler) Market(c *gin.Context) {
	stats, err := h.statsSvc.GetMarketStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// Popular godoc
// GET /api/v1/stats/popular
func (h *StatsHandler) Popular(c *gin.Context) {
	properties, err := h.statsSvc.GetPopularProperties(10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	c.JSON(http.StatusOK, properties)
}

// Dashboard godoc
// GET /api/v1/stats/dashboard — agent/admin uniquement
func (h *StatsHandler) Dashboard(c *gin.Context) {
	stats, err := h.statsSvc.GetDashboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// ── Agent ─────────────────────────────────────────────────────────────────────

// AgentHandler gère les routes spécifiques aux agents
type AgentHandler struct {
	messageSvc service.MessageService
}

func NewAgentHandler(messageSvc service.MessageService) *AgentHandler {
	return &AgentHandler{messageSvc: messageSvc}
}

// ContactAgent godoc
// POST /api/v1/properties/:id/contact
func (h *AgentHandler) ContactAgent(c *gin.Context) {
	propertyID, err := parseID(c, "id")
	if err != nil {
		return
	}

	var req domain.ContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Récupère l'ID de l'expéditeur depuis le contexte JWT (injecté par le middleware)
	senderID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "non authentifié"})
		return
	}

	if err := h.messageSvc.Send(propertyID, senderID, req.Message); err != nil {
		switch err.Error() {
		case "bien introuvable":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "vous ne pouvez pas contacter votre propre annonce":
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "message envoyé"})
}

// GetMessages godoc
// GET /api/v1/agent/messages
// Retourne tous les messages des conversations de cet agent
// (messages des clients ET ses propres réponses), groupables côté
// frontend par (property_id, client_id).
func (h *AgentHandler) GetMessages(c *gin.Context) {
	agentID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "non authentifié"})
		return
	}

	threads, err := h.messageSvc.GetThreadsForAgent(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}

	c.JSON(http.StatusOK, threads)
}
