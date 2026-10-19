package application

import (
	"context"
	"time"

	"hexagonal-go-backend/internal/modules/users/domain"
)

type UserFilter struct {
	Page, Limit int
	Search      string
	Status      domain.UserStatus
	Active      *bool
	Sort, Order string
}

type CreateUserInput struct {
	AuthID       string
	Name, Email  string
	AvatarURL    string
	IsSuperAdmin bool
}

type UpdateUserInput struct {
	Name, Email string
	AvatarURL   *string
	Active      *bool
	Status      *domain.UserStatus
}

type UpdateProfileInput struct {
	FirstName            *string
	LastName             *string
	Phone                *string
	BirthDate            *time.Time
	Gender               *domain.Gender
	City                 *string
	Address              *string
	PhotoURL             *string
	Bio                  *string
	SociodemographicData map[string]any
}

type UserRepository interface {
	Create(context.Context, *domain.User) error
	FindByID(context.Context, string) (*domain.User, error)
	FindByAuthID(context.Context, string) (*domain.User, error)
	FindByEmail(context.Context, string) (*domain.User, error)
	List(context.Context, UserFilter) ([]domain.User, int, error)
	Update(context.Context, *domain.User) error
	Delete(context.Context, string) error
	GetProfile(context.Context, string) (*domain.UserProfile, error)
	UpsertProfile(context.Context, *domain.UserProfile) error
}

type UserService interface {
	Create(context.Context, CreateUserInput) (*domain.User, error)
	FindByID(context.Context, string) (*domain.User, error)
	List(context.Context, UserFilter) ([]domain.User, int, error)
	Update(context.Context, string, UpdateUserInput) (*domain.User, error)
	Delete(context.Context, string) error
	GetProfile(context.Context, string) (*domain.UserProfile, error)
	UpdateProfile(context.Context, string, UpdateProfileInput) (*domain.UserProfile, error)
}

type MailMessage struct{ To, Subject, Body string }

type Mailer interface {
	Send(context.Context, MailMessage) error
}
