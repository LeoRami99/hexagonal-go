package v1

import (
	"time"

	"hexagonal-go-backend/internal/modules/courses/domain"
)

type CreateCourseRequest struct {
	TenantID     string `json:"tenant_id" binding:"required,len=26"`
	ProjectID    string `json:"project_id" binding:"required,len=26"`
	Title        string `json:"title" binding:"required,min=2,max=350"`
	Slug         string `json:"slug" binding:"required,min=2,max=150"`
	Summary      string `json:"summary" binding:"required,min=2,max=350"`
	Description  string `json:"description" binding:"required,min=2"`
	Level        string `json:"level" binding:"required,oneof=beginner intermediate advanced"`
	ThumbnailURL string `json:"thumbnail_url" binding:"required,url"`
	Status       string `json:"status" binding:"required,oneof=draft published archived"`
}

type UpdateCourseRequest struct {
	Title        *string `json:"title" binding:"omitempty,min=2,max=350"`
	Slug         *string `json:"slug" binding:"omitempty,min=2,max=150"`
	Summary      *string `json:"summary" binding:"omitempty,min=2"`
	Description  *string `json:"description" binding:"omitempty,min=2"`
	Level        *string `json:"level" binding:"omitempty,oneof=beginner intermediate advanced"`
	ThumbnailURL *string `json:"thumbnail_url" binding:"omitempty,url"`
	Status       *string `json:"status" binding:"omitempty,oneof=draft published archived"`
}

type CreateCourseResponse struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	ProjectID    string    `json:"project_id"`
	Title        string    `json:"title"`
	Slug         string    `json:"slug"`
	Summary      string    `json:"summary"`
	Description  string    `json:"description"`
	Level        string    `json:"level"`
	ThumbnailURL string    `json:"thumbnail_url"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func createCourseResponse(course *domain.Course) CreateCourseResponse {
	return CreateCourseResponse{
		ID:           course.ID,
		TenantID:     course.TenantID,
		ProjectID:    course.ProjectID,
		Title:        course.Title,
		Slug:         course.Slug,
		Summary:      course.Summary,
		Description:  course.Description,
		Level:        string(course.Level),
		ThumbnailURL: course.ThumbnailURL,
		Status:       string(course.Status),
		CreatedAt:    course.CreatedAt,
		UpdatedAt:    course.UpdatedAt,
	}
}
