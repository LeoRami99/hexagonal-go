package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrTenantAlreadyExists = errors.New("tenant already exists")
	ErrInvalidTenantData   = errors.New("invalid tenant data")
)

type TenantStatus string

const (
	StatusActive    TenantStatus = "ACTIVE"
	StatusInactive  TenantStatus = "INACTIVE"
	StatusSuspended TenantStatus = "SUSPENDED"
	StatusPending   TenantStatus = "PENDING"
)

type Tenant struct {
	ID           string
	Name         string
	Slug         string
	ContactEmail string
	Status       TenantStatus
	Settings     map[string]any
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type TenantRepository interface {
	Create(ctx context.Context, tenant *Tenant) error
	FindByID(ctx context.Context, id string) (*Tenant, error)
	FindBySlug(ctx context.Context, slug string) (*Tenant, error)
	List(ctx context.Context) ([]*Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, id string) error
}
