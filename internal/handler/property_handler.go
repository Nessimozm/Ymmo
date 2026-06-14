package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"ymmo/internal/domain"
	"ymmo/internal/middleware"
	"ymmo/internal/service"
)

// PropertyHandler gère les routes /properties
type PropertyHandler struct {
	propertySvc service.PropertyService
}

// NewPropertyHandler retourne une instance du PropertyHandler
func NewPropertyHandler(propertySvc service.PropertyService) *PropertyHandler {
	return &PropertyHandler{propertySvc: propertySvc}
}

// List godoc
// GET /api/v1/properties
func (h *PropertyHandler) List(c *gin.Context) {
	var filters domain.PropertyFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.propertySvc.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetByID godoc
// GET /api/v1/properties/:id
func (h *PropertyHandler) GetByID(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}

	property, err := h.propertySvc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	if property == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bien introuvable"})
		return
	}

	// Incrémente le compteur de vues en arrière-plan
	go h.propertySvc.IncrementViewCount(id)

	c.JSON(http.StatusOK, property)
}

// Create godoc
// POST /api/v1/properties — agent uniquement
func (h *PropertyHandler) Create(c *gin.Context) {
	var req domain.CreatePropertyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "non authentifié"})
		return
	}

	property, err := h.propertySvc.Create(&req, agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, property)
}

// Update godoc
// PUT /api/v1/properties/:id — agent (ses biens) ou admin (tous les biens)
func (h *PropertyHandler) Update(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}

	var req domain.UpdatePropertyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetUserRole(c)

	property, err := h.propertySvc.Update(id, &req, agentID, role)
	if err != nil {
		// "accès refusé" → 403, le reste → 500
		if err.Error() == "accès refusé" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if property == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bien introuvable"})
		return
	}

	c.JSON(http.StatusOK, property)
}

// Delete godoc
// DELETE /api/v1/properties/:id — agent (ses biens) ou admin (tous les biens)
func (h *PropertyHandler) Delete(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}

	agentID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetUserRole(c)

	if err := h.propertySvc.Delete(id, agentID, role); err != nil {
		if err.Error() == "accès refusé" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "bien supprimé"})
}

// UploadImages godoc
// POST /api/v1/properties/:id/images — agent uniquement
func (h *PropertyHandler) UploadImages(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "formulaire invalide"})
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "aucun fichier envoyé"})
		return
	}

	var urls []string
	for i, file := range files {
		// Valide le type MIME
		ext := filepath.Ext(file.Filename)
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "format accepté : jpg, png, webp"})
			return
		}

		// Nom unique basé sur timestamp
		filename := fmt.Sprintf("property_%d_%d_%d%s", id, time.Now().Unix(), i, ext)
		savePath := fmt.Sprintf("./uploads/%s", filename)

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur sauvegarde image"})
			return
		}

		urls = append(urls, fmt.Sprintf("/uploads/%s", filename))
	}

	// Enregistre les URLs en BDD (la première image devient la principale)
	if err := h.propertySvc.AddImages(id, urls); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"urls": urls})
}

// ── Helper ────────────────────────────────────────────────────────────────────

// parseID extrait et convertit un paramètre d'URL en uint
func parseID(c *gin.Context, param string) (uint, error) {
	raw := c.Param(param)
	id, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "identifiant invalide"})
		return 0, err
	}
	return uint(id), nil
}
