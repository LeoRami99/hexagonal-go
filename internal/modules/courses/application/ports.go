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


type CoursesRepository interface {
	Create(context.Context, *domain.Course) error
	FindByID(context.Context, string) (*domain.Course, error)
	FindBySlug(context.Context, string) (*domain.Course, error)
	List(context.Context, CoursesFilter) ([]domain.Course, int, error)
	Update(context.Context, *domain.Course) error
	Delete(context.Context, string) error
}