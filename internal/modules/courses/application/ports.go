package application

import (
	"context"
	"hexagonal-go-backend/internal/modules/courses/domain"
)

type CoursesFilter struct {
	Page, Limit int
	Search      string
	Level       string
	Status      string
	Sort, Order string
}

type CreateCourseInput struct {
	TenantID     string
	ProjectID    string
	Title        string
	Slug         string
	Summary      string
	Description  string
	Level        string
	ThumbnailURL string
	Status       string
}

type UpdateCourseInput struct {
	Title        *string
	Slug         *string
	Summary      *string
	Description  *string
	Level        *string
	ThumbnailURL *string
	Status       *string
}

type CoursesRepository interface {
	Create(context.Context, *domain.Course) error
	FindByID(context.Context, string) (*domain.Course, error)
	FindBySlug(context.Context, string, string) (*domain.Course, error)
	List(context.Context, CoursesFilter) ([]domain.Course, int, error)
	Update(context.Context, *domain.Course) error
	Delete(context.Context, string) error
}

type CoursesService interface {
	Create(context.Context, CreateCourseInput) (*domain.Course, error)
	FindByID(context.Context, string) (*domain.Course, error)
	FindBySlug(context.Context, string, string) (*domain.Course, error)
	List(context.Context, CoursesFilter) ([]domain.Course, int, error)
	Update(context.Context, string, UpdateCourseInput) (*domain.Course, error)
	Delete(context.Context, string) error
}
