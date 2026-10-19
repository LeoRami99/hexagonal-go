package application

import (
	"context"
	"hexagonal-go-backend/internal/modules/courses/domain"
)

type coursesService struct {
	repo CoursesRepository
}



func NewCoursesService(repo CoursesRepository) *coursesService {
	return &coursesService{repo: repo}
}



func (s *coursesService) Create(ctx context.Context, in domain.Course) (*domain.Course, error) {
	if in.ProjectID == "" || in.Title == "" || in.Slug == "" || in.Summary == "" || in.Description == "" || in.Level == "" || in.ThumbnailURL == "" || in.Status == "" {
		return nil, domain.ErrCourseEmptyFields
	}

	if _, err := s.repo.FindBySlug(ctx, in.Slug); err == nil {
		return nil, domain.ErrCourseAlreadyExists
	}
	if err := s.repo.Create(ctx, &in); err != nil {
		return nil, err
	}
	return &in, nil
}