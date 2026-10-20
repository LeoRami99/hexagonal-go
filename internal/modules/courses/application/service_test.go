package application

import (
	"context"
	"errors"
	"testing"

	"hexagonal-go-backend/internal/modules/courses/domain"
)

type courseRepoStub struct {
	items map[string]*domain.Course
}

func newCourseRepoStub() *courseRepoStub { return &courseRepoStub{items: map[string]*domain.Course{}} }

func (r *courseRepoStub) Create(_ context.Context, course *domain.Course) error {
	r.items[course.ID] = course
	return nil
}
func (r *courseRepoStub) FindByID(_ context.Context, id string) (*domain.Course, error) {
	course, ok := r.items[id]
	if !ok {
		return nil, domain.ErrCourseNotFound
	}
	copy := *course
	return &copy, nil
}
func (r *courseRepoStub) FindBySlug(_ context.Context, projectID, slug string) (*domain.Course, error) {
	for _, course := range r.items {
		if course.ProjectID == projectID && course.Slug == slug {
			copy := *course
			return &copy, nil
		}
	}
	return nil, domain.ErrCourseNotFound
}
func (r *courseRepoStub) List(context.Context, CoursesFilter) ([]domain.Course, int, error) {
	return nil, len(r.items), nil
}
func (r *courseRepoStub) Update(_ context.Context, course *domain.Course) error {
	r.items[course.ID] = course
	return nil
}
func (r *courseRepoStub) Delete(_ context.Context, id string) error {
	if _, ok := r.items[id]; !ok {
		return domain.ErrCourseNotFound
	}
	delete(r.items, id)
	return nil
}

func validCreateInput() CreateCourseInput {
	return CreateCourseInput{
		TenantID: "01K7Z8V1Y8QY33ZZZZZZZZZZZZ", ProjectID: "01K7Z8V1Y8QY33YYYYYYYYYYYY",
		Title: "Go básico", Slug: "go-basico", Summary: "Introducción a Go",
		Description: "Contenido del curso", Level: "beginner",
		ThumbnailURL: "https://example.com/go.png", Status: "draft",
	}
}

func TestCreateCourse(t *testing.T) {
	service := NewCoursesService(newCourseRepoStub())
	course, err := service.Create(context.Background(), validCreateInput())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if course.ID == "" || course.Slug != "go-basico" || course.Status != domain.StatusDraft {
		t.Fatalf("unexpected course: %#v", course)
	}
}

func TestCreateCourseRejectsDuplicateProjectSlug(t *testing.T) {
	repo := newCourseRepoStub()
	service := NewCoursesService(repo)
	if _, err := service.Create(context.Background(), validCreateInput()); err != nil {
		t.Fatal(err)
	}
	_, err := service.Create(context.Background(), validCreateInput())
	if !errors.Is(err, domain.ErrCourseAlreadyExists) {
		t.Fatalf("error = %v", err)
	}
}

func TestCreateCourseValidatesEnums(t *testing.T) {
	input := validCreateInput()
	input.Level = "expert"
	_, err := NewCoursesService(newCourseRepoStub()).Create(context.Background(), input)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("error = %v", err)
	}
}
