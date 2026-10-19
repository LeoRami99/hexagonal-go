package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"hexagonal-go-backend/internal/modules/courses/domain"
	"hexagonal-go-backend/internal/platform/identifier"
)

type coursesService struct {
	repo CoursesRepository
}

func NewCoursesService(repo CoursesRepository) CoursesService {
	return &coursesService{repo: repo}
}

func (s *coursesService) Create(ctx context.Context, in CreateCourseInput) (*domain.Course, error) {
	tenantID, projectID := strings.TrimSpace(in.TenantID), strings.TrimSpace(in.ProjectID)
	title, slug := strings.TrimSpace(in.Title), strings.ToLower(strings.TrimSpace(in.Slug))
	if tenantID == "" || projectID == "" || title == "" || slug == "" || strings.TrimSpace(in.Summary) == "" || strings.TrimSpace(in.Description) == "" || strings.TrimSpace(in.ThumbnailURL) == "" {
		return nil, domain.ErrInvalidInput
	}
	level, status := domain.Level(in.Level), domain.Status(in.Status)
	if !validLevel(level) || !validStatus(status) {
		return nil, domain.ErrInvalidInput
	}
	if _, err := s.repo.FindBySlug(ctx, projectID, slug); err == nil {
		return nil, domain.ErrCourseAlreadyExists
	} else if !errors.Is(err, domain.ErrCourseNotFound) {
		return nil, err
	}
	now := time.Now()
	course := &domain.Course{
		ID: identifier.New(), TenantID: tenantID, ProjectID: projectID,
		Title: title, Slug: slug, Summary: strings.TrimSpace(in.Summary),
		Description: strings.TrimSpace(in.Description), Level: level,
		ThumbnailURL: strings.TrimSpace(in.ThumbnailURL), Status: status,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, course); err != nil {
		return nil, err
	}
	return course, nil
}

func (s *coursesService) FindByID(ctx context.Context, id string) (*domain.Course, error) {
	if strings.TrimSpace(id) == "" {
		return nil, domain.ErrInvalidInput
	}
	return s.repo.FindByID(ctx, id)
}

func (s *coursesService) FindBySlug(ctx context.Context, projectID, slug string) (*domain.Course, error) {
	if strings.TrimSpace(projectID) == "" || strings.TrimSpace(slug) == "" {
		return nil, domain.ErrInvalidInput
	}
	return s.repo.FindBySlug(ctx, projectID, strings.ToLower(strings.TrimSpace(slug)))
}

func (s *coursesService) List(ctx context.Context, filter CoursesFilter) ([]domain.Course, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Sort == "" {
		filter.Sort = "created_at"
	}
	if filter.Order != "asc" {
		filter.Order = "desc"
	}
	return s.repo.List(ctx, filter)
}

func (s *coursesService) Update(ctx context.Context, id string, in UpdateCourseInput) (*domain.Course, error) {
	course, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Title != nil {
		course.Title = strings.TrimSpace(*in.Title)
	}
	if in.Slug != nil {
		course.Slug = strings.ToLower(strings.TrimSpace(*in.Slug))
	}
	if in.Summary != nil {
		course.Summary = strings.TrimSpace(*in.Summary)
	}
	if in.Description != nil {
		course.Description = strings.TrimSpace(*in.Description)
	}
	if in.Level != nil {
		course.Level = domain.Level(*in.Level)
	}
	if in.ThumbnailURL != nil {
		course.ThumbnailURL = strings.TrimSpace(*in.ThumbnailURL)
	}
	if in.Status != nil {
		course.Status = domain.Status(*in.Status)
	}
	if course.Title == "" || course.Slug == "" || course.Summary == "" || course.Description == "" || course.ThumbnailURL == "" || !validLevel(course.Level) || !validStatus(course.Status) {
		return nil, domain.ErrInvalidInput
	}
	if existing, findErr := s.repo.FindBySlug(ctx, course.ProjectID, course.Slug); findErr == nil && existing.ID != course.ID {
		return nil, domain.ErrCourseAlreadyExists
	} else if findErr != nil && !errors.Is(findErr, domain.ErrCourseNotFound) {
		return nil, findErr
	}
	course.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, course); err != nil {
		return nil, err
	}
	return course, nil
}

func (s *coursesService) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: id is required", domain.ErrInvalidInput)
	}
	return s.repo.Delete(ctx, id)
}

func validLevel(level domain.Level) bool {
	return level == domain.LevelBeginner || level == domain.LevelIntermediate || level == domain.LevelAdvanced
}

func validStatus(status domain.Status) bool {
	return status == domain.StatusDraft || status == domain.StatusPublished || status == domain.StatusArchived
}
