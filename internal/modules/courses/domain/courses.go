package domain

import "time"

type Level string

const (
	LevelBeginner     Level = "beginner"
	LevelIntermediate Level = "intermediate"
	LevelAdvanced     Level = "advanced"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
	StatusArchived  Status = "archived"
)

type Course struct {
	ID           string
	TenantID     string
	ProjectID    string
	Title        string
	Slug         string
	Summary      string
	Description  string
	Level        Level
	ThumbnailURL string
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
