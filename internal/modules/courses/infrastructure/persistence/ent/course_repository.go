package ent

import (
	"context"
	"errors"
	"fmt"
	"strings"

	entclient "hexagonal-go-backend/internal/ent"
	entcourse "hexagonal-go-backend/internal/ent/courses"
	"hexagonal-go-backend/internal/modules/courses/application"
	"hexagonal-go-backend/internal/modules/courses/domain"

	"entgo.io/ent/dialect/sql"
)

type CourseRepository struct{ client *entclient.Client }

func NewCourseRepository(client *entclient.Client) *CourseRepository {
	return &CourseRepository{client: client}
}

func (r *CourseRepository) Create(ctx context.Context, course *domain.Course) error {
	created, err := r.client.Courses.Create().
		SetID(course.ID).
		SetTenantID(course.TenantID).
		SetProjectID(course.ProjectID).
		SetTitle(course.Title).
		SetSlug(course.Slug).
		SetSummary(course.Summary).
		SetDescription(course.Description).
		SetLevel(entcourse.Level(course.Level)).
		SetThumbnailURL(course.ThumbnailURL).
		SetStatus(entcourse.Status(course.Status)).
		SetCreatedAt(course.CreatedAt).
		SetUpdatedAt(course.UpdatedAt).
		Save(ctx)
	if err != nil {
		return translateError(err)
	}
	*course = *toDomain(created)
	return nil
}

func (r *CourseRepository) FindByID(ctx context.Context, id string) (*domain.Course, error) {
	course, err := r.client.Courses.Get(ctx, id)
	if err != nil {
		return nil, translateError(err)
	}
	return toDomain(course), nil
}

func (r *CourseRepository) FindBySlug(ctx context.Context, projectID, slug string) (*domain.Course, error) {
	course, err := r.client.Courses.Query().Where(
		entcourse.ProjectIDEQ(projectID), entcourse.SlugEQ(slug),
	).Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return toDomain(course), nil
}

func (r *CourseRepository) List(ctx context.Context, filter application.CoursesFilter) ([]domain.Course, int, error) {
	query := r.client.Courses.Query()
	if search := strings.TrimSpace(filter.Search); search != "" {
		query = query.Where(entcourse.Or(entcourse.TitleContainsFold(search), entcourse.SlugContainsFold(search), entcourse.SummaryContainsFold(search)))
	}
	if filter.Level != "" {
		query = query.Where(entcourse.LevelEQ(entcourse.Level(filter.Level)))
	}
	if filter.Status != "" {
		query = query.Where(entcourse.StatusEQ(entcourse.Status(filter.Status)))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count courses: %w", err)
	}
	order := sql.OrderDesc()
	if filter.Order == "asc" {
		order = sql.OrderAsc()
	}
	switch filter.Sort {
	case "title":
		query = query.Order(entcourse.ByTitle(order))
	case "updated_at":
		query = query.Order(entcourse.ByUpdatedAt(order))
	default:
		query = query.Order(entcourse.ByCreatedAt(order))
	}
	items, err := query.Offset((filter.Page - 1) * filter.Limit).Limit(filter.Limit).All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list courses: %w", err)
	}
	result := make([]domain.Course, 0, len(items))
	for _, item := range items {
		result = append(result, *toDomain(item))
	}
	return result, total, nil
}

func (r *CourseRepository) Update(ctx context.Context, course *domain.Course) error {
	updated, err := r.client.Courses.UpdateOneID(course.ID).
		SetTitle(course.Title).SetSlug(course.Slug).SetSummary(course.Summary).
		SetDescription(course.Description).SetLevel(entcourse.Level(course.Level)).
		SetThumbnailURL(course.ThumbnailURL).SetStatus(entcourse.Status(course.Status)).
		SetUpdatedAt(course.UpdatedAt).Save(ctx)
	if err != nil {
		return translateError(err)
	}
	*course = *toDomain(updated)
	return nil
}

func (r *CourseRepository) Delete(ctx context.Context, id string) error {
	return translateError(r.client.Courses.DeleteOneID(id).Exec(ctx))
}

func toDomain(course *entclient.Courses) *domain.Course {
	if course == nil {
		return nil
	}
	return &domain.Course{
		ID: course.ID, TenantID: course.TenantID, ProjectID: course.ProjectID,
		Title: course.Title, Slug: course.Slug, Summary: course.Summary,
		Description: course.Description, Level: domain.Level(course.Level),
		ThumbnailURL: course.ThumbnailURL, Status: domain.Status(course.Status),
		CreatedAt: course.CreatedAt, UpdatedAt: course.UpdatedAt,
	}
}

func translateError(err error) error {
	if err == nil {
		return nil
	}
	if entclient.IsNotFound(err) {
		return domain.ErrCourseNotFound
	}
	if entclient.IsConstraintError(err) {
		return domain.ErrCourseAlreadyExists
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return fmt.Errorf("courses repository: %w", err)
}
