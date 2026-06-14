package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ymmo/internal/domain"
	"ymmo/internal/middleware"
	"ymmo/internal/service"
)

// MessageHandler gère les routes /messages — communes aux clients et aux agents.
// Contrairement à AgentHandler (réservé aux agents/admins), ces routes sont
// accessibles à tout utilisateur authentifié, le rôle étant déterminé par
// la conversation elle-même (un client ne voit que SES conversations).
type MessageHandler struct {
	messageSvc service.MessageService
}

// NewMessageHandler retourne une instance du MessageHandler
func NewMessageHandler(messageSvc service.MessageService) *MessageHandler {
	return &MessageHandler{messageSvc: messageSvc}
}

// GetMyMessages godoc
// GET /api/v1/messages
// Retourne toutes les conversations où l'utilisateur connecté est le CLIENT
// (ex: un client suit les réponses des agents qu'il a contactés).
func (h *MessageHandler) GetMyMessages(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "non authentifié"})
		return
	}

	threads, err := h.messageSvc.GetThreadsForClient(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}

	c.JSON(http.StatusOK, threads)
}

// Reply godoc
// POST /api/v1/messages/:id/reply
// Ajoute un message à une conversation existante. :id est l'ID de N'IMPORTE
// QUEL message de cette conversation (le service en déduit le fil complet).
// Le rôle de l'auteur (client ou agent) est déterminé automatiquement selon
// sa position dans la conversation.
func (h *MessageHandler) Reply(c *gin.Context) {
	messageID, err := parseID(c, "id")
	if err != nil {
		return
	}

	var req domain.ReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "non authentifié"})
		return
	}

	if err := h.messageSvc.Reply(messageID, userID, req.Message); err != nil {
		switch err.Error() {
		case "conversation introuvable":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "accès refusé":
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "réponse envoyée"})
}

// MarkAsRead godoc
// PATCH /api/v1/messages/:id/read
// Marque un message comme lu — uniquement par son destinataire
// (le service vérifie que l'utilisateur est bien l'autre partie de la conversation).
func (h *MessageHandler) MarkAsRead(c *gin.Context) {
	messageID, err := parseID(c, "id")
	if err != nil {
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "non authentifié"})
		return
	}

	if err := h.messageSvc.MarkAsRead(messageID, userID); err != nil {
		switch err.Error() {
		case "message introuvable":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "accès refusé":
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "marqué comme lu"})
}
