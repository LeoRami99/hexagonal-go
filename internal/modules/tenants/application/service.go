package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hexagonal-go-backend/internal/modules/tenants/domain"
	"hexagonal-go-backend/internal/platform/identifier"
)

type CreateTenantInput struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Slug         string         `json:"slug"`
	ContactEmail string         `json:"contact_email"`
	Settings     map[string]any `json:"settings"`
}

type UpdateTenantInput struct {
	Name         *string        `json:"name"`
	Slug         *string        `json:"slug"`
	ContactEmail *string        `json:"contact_email"`
	Status       *string        `json:"status"`
	Settings     map[string]any `json:"settings"`
}

type TenantService interface {
	Create(ctx context.Context, input CreateTenantInput) (*domain.Tenant, error)
	GetByID(ctx context.Context, id string) (*domain.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
	List(ctx context.Context) ([]*domain.Tenant, error)
	Update(ctx context.Context, id string, input UpdateTenantInput) (*domain.Tenant, error)
	Delete(ctx context.Context, id string) error
}

type tenantService struct {
	repo domain.TenantRepository
}

func NewTenantService(repo domain.TenantRepository) TenantService {
	return &tenantService{repo: repo}
}

func (s *tenantService) Create(ctx context.Context, input CreateTenantInput) (*domain.Tenant, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrInvalidTenantData)
	}

	slug := strings.ToLower(strings.TrimSpace(input.Slug))
	if slug == "" {
		slug = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	}

	id := input.ID
	if id == "" {
		id = identifier.New()
	}

	now := time.Now()
	tenant := &domain.Tenant{
		ID:           id,
		Name:         name,
		Slug:         slug,
		ContactEmail: strings.TrimSpace(input.ContactEmail),
		Status:       domain.StatusActive,
		Settings:     input.Settings,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, err
	}

	return tenant, nil
}

func (s *tenantService) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", domain.ErrInvalidTenantData)
	}
	return s.repo.FindByID(ctx, id)
}

func (s *tenantService) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, fmt.Errorf("%w: slug is required", domain.ErrInvalidTenantData)
	}
	return s.repo.FindBySlug(ctx, slug)
}

func (s *tenantService) List(ctx context.Context) ([]*domain.Tenant, error) {
	return s.repo.List(ctx)
}

func (s *tenantService) Update(ctx context.Context, id string, input UpdateTenantInput) (*domain.Tenant, error) {
	tenant, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, fmt.Errorf("%w: name cannot be empty", domain.ErrInvalidTenantData)
		}
		tenant.Name = name
	}
	if input.Slug != nil {
		slug := strings.TrimSpace(*input.Slug)
		if slug != "" {
			tenant.Slug = slug
		}
	}
	if input.ContactEmail != nil {
		tenant.ContactEmail = strings.TrimSpace(*input.ContactEmail)
	}
	if input.Status != nil {
		tenant.Status = domain.TenantStatus(*input.Status)
	}
	if input.Settings != nil {
		tenant.Settings = input.Settings
	}

	tenant.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, tenant); err != nil {
		return nil, err
	}

	return tenant, nil
}

func (s *tenantService) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("%w: id is required", domain.ErrInvalidTenantData)
	}
	return s.repo.Delete(ctx, id)
}
