package v1

import (
	"time"
)


type CreateCourseRequest struct {
	Title       string    `json:"title" binding:"required,min=2,max=100"`
	Slug        string    `json:"slug" binding:"required,min=2,max=100"`
	Summary     string    `json:"summary" binding:"required,min=2,max=350"`
	Description string    `json:"description" binding:"required,min=2"`
	Level       string    `json:"level" binding:"required,oneof=beginner intermediate advanced"`
	ThumbnailURL string   `json:"thumbnail_url" binding:"required,url"`
	Status      string    `json:"status" binding:"required,oneof=draft published archived"`
}

type CreateCourseResponse struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Level       string    `json:"level"`
	ThumbnailURL string   `json:"thumbnail_url"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}