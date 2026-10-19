package ent

import (
	"context"
	"errors"
	"fmt"

	entclient "hexagonal-go-backend/internal/ent"
	enttenants "hexagonal-go-backend/internal/ent/tenants"
	"hexagonal-go-backend/internal/modules/tenants/domain"
)

type TenantRepository struct {
	client *entclient.Client
}

func NewTenantRepository(client *entclient.Client) *TenantRepository {
	return &TenantRepository{client: client}
}

func (r *TenantRepository) Create(ctx context.Context, t *domain.Tenant) error {
	builder := r.client.Tenants.Create().
		SetName(t.Name).
		SetSlug(t.Slug).
		SetStatus(enttenants.Status(t.Status))

	if t.ID != "" {
		builder.SetID(t.ID)
	}
	if t.ContactEmail != "" {
		builder.SetContactEmail(t.ContactEmail)
	}
	if t.Settings != nil {
		builder.SetSettings(t.Settings)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return translateError(err)
	}
	t.ID = created.ID
	t.CreatedAt = created.CreatedAt
	t.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *TenantRepository) FindByID(ctx context.Context, id string) (*domain.Tenant, error) {
	entity, err := r.client.Tenants.Query().
		Where(enttenants.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return toDomain(entity), nil
}

func (r *TenantRepository) FindBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	entity, err := r.client.Tenants.Query().
		Where(enttenants.SlugEQ(slug)).
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return toDomain(entity), nil
}

func (r *TenantRepository) List(ctx context.Context) ([]*domain.Tenant, error) {
	entities, err := r.client.Tenants.Query().
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}

	result := make([]*domain.Tenant, 0, len(entities))
	for _, e := range entities {
		result = append(result, toDomain(e))
	}
	return result, nil
}

func (r *TenantRepository) Update(ctx context.Context, t *domain.Tenant) error {
	builder := r.client.Tenants.UpdateOneID(t.ID).
		SetName(t.Name).
		SetSlug(t.Slug).
		SetStatus(enttenants.Status(t.Status))

	if t.ContactEmail != "" {
		builder.SetContactEmail(t.ContactEmail)
	} else {
		builder.ClearContactEmail()
	}
	if t.Settings != nil {
		builder.SetSettings(t.Settings)
	}

	_, err := builder.Save(ctx)
	return translateError(err)
}

func (r *TenantRepository) Delete(ctx context.Context, id string) error {
	return translateError(r.client.Tenants.DeleteOneID(id).Exec(ctx))
}

func toDomain(e *entclient.Tenants) *domain.Tenant {
	if e == nil {
		return nil
	}
	contactEmail := ""
	if e.ContactEmail != nil {
		contactEmail = *e.ContactEmail
	}
	return &domain.Tenant{
		ID:           e.ID,
		Name:         e.Name,
		Slug:         e.Slug,
		ContactEmail: contactEmail,
		Status:       domain.TenantStatus(e.Status),
		Settings:     e.Settings,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}

func translateError(err error) error {
	if err == nil {
		return nil
	}
	if entclient.IsNotFound(err) {
		return domain.ErrTenantNotFound
	}
	if entclient.IsConstraintError(err) {
		return domain.ErrTenantAlreadyExists
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return fmt.Errorf("tenant repository: %w", err)
}

var _ domain.TenantRepository = (*TenantRepository)(nil)
