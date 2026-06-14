package service

import (
	"fmt"

	"ymmo/internal/domain"
	"ymmo/internal/repository"
)

// PropertyService gère la logique métier des biens immobiliers
type PropertyService interface {
	Create(req *domain.CreatePropertyRequest, agentID uint) (*domain.Property, error)
	GetByID(id uint) (*domain.Property, error)
	List(filters domain.PropertyFilters) (*domain.PropertyListResponse, error)
	Update(id uint, req *domain.UpdatePropertyRequest, agentID uint, role domain.Role) (*domain.Property, error)
	Delete(id uint, agentID uint, role domain.Role) error
	IncrementViewCount(id uint)
	AddImages(propertyID uint, urls []string) error
}

type propertyService struct {
	propertyRepo repository.PropertyRepository
}

// NewPropertyService retourne une instance du PropertyService
func NewPropertyService(propertyRepo repository.PropertyRepository) PropertyService {
	return &propertyService{propertyRepo: propertyRepo}
}

func (s *propertyService) Create(req *domain.CreatePropertyRequest, agentID uint) (*domain.Property, error) {
	property := &domain.Property{
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		Surface:     req.Surface,
		Rooms:       req.Rooms,
		Bedrooms:    req.Bedrooms,
		Type:        req.Type,
		Status:      domain.StatusAvailable, // toujours disponible à la création
		Transaction: req.Transaction,
		Address:     req.Address,
		City:        req.City,
		ZipCode:     req.ZipCode,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		AgentID:     agentID,
	}

	if err := s.propertyRepo.Create(property); err != nil {
		return nil, fmt.Errorf("propertyService.Create : %w", err)
	}

	return property, nil
}

func (s *propertyService) GetByID(id uint) (*domain.Property, error) {
	return s.propertyRepo.GetByID(id)
}

func (s *propertyService) List(filters domain.PropertyFilters) (*domain.PropertyListResponse, error) {
	// Valeurs par défaut
	if filters.Limit == 0 {
		filters.Limit = 12
	}
	if filters.Page == 0 {
		filters.Page = 1
	}

	properties, total, err := s.propertyRepo.List(filters)
	if err != nil {
		return nil, fmt.Errorf("propertyService.List : %w", err)
	}

	totalPages := total / filters.Limit
	if total%filters.Limit != 0 {
		totalPages++
	}

	return &domain.PropertyListResponse{
		Data:       properties,
		Total:      total,
		Page:       filters.Page,
		TotalPages: totalPages,
	}, nil
}

func (s *propertyService) Update(id uint, req *domain.UpdatePropertyRequest, agentID uint, role domain.Role) (*domain.Property, error) {
	// Récupère le bien existant
	property, err := s.propertyRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if property == nil {
		return nil, nil
	}

	// Un agent ne peut modifier que ses propres biens — un admin peut tout modifier
	if role != domain.RoleAdmin && property.AgentID != agentID {
		return nil, fmt.Errorf("accès refusé")
	}

	// Applique uniquement les champs envoyés (pointeurs non nil)
	if req.Title != nil {
		property.Title = *req.Title
	}
	if req.Description != nil {
		property.Description = *req.Description
	}
	if req.Price != nil {
		property.Price = *req.Price
	}
	if req.Surface != nil {
		property.Surface = *req.Surface
	}
	if req.Rooms != nil {
		property.Rooms = *req.Rooms
	}
	if req.Bedrooms != nil {
		property.Bedrooms = *req.Bedrooms
	}
	if req.Status != nil {
		property.Status = *req.Status
	}
	if req.Transaction != nil {
		property.Transaction = *req.Transaction
	}
	if req.Address != nil {
		property.Address = *req.Address
	}
	if req.City != nil {
		property.City = *req.City
	}
	if req.ZipCode != nil {
		property.ZipCode = *req.ZipCode
	}

	if err := s.propertyRepo.Update(property); err != nil {
		return nil, fmt.Errorf("propertyService.Update : %w", err)
	}

	return property, nil
}

func (s *propertyService) Delete(id uint, agentID uint, role domain.Role) error {
	property, err := s.propertyRepo.GetByID(id)
	if err != nil {
		return err
	}
	if property == nil {
		return fmt.Errorf("bien introuvable")
	}

	// Un agent ne peut supprimer que ses propres biens
	// Un admin peut tout supprimer
	if role != domain.RoleAdmin && property.AgentID != agentID {
		return fmt.Errorf("accès refusé")
	}

	return s.propertyRepo.Delete(id)
}

// IncrementViewCount est appelé en goroutine — pas de retour d'erreur
func (s *propertyService) IncrementViewCount(id uint) {
	_ = s.propertyRepo.IncrementViewCount(id)
}

func (s *propertyService) AddImages(propertyID uint, urls []string) error {
	for i, url := range urls {
		img := &domain.PropertyImage{
			PropertyID: propertyID,
			URL:        url,
			IsPrimary:  i == 0, // la première image est la principale
		}
		if err := s.propertyRepo.AddImage(img); err != nil {
			return fmt.Errorf("propertyService.AddImages : %w", err)
		}
	}
	return nil
}
