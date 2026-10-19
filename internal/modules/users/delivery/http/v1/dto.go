package v1

import (
	"time"

	"hexagonal-go-backend/internal/modules/users/domain"
)

type CreateUserRequest struct {
	AuthID       string `json:"auth_id" binding:"omitempty"`
	Name         string `json:"name" binding:"required,min=2,max=100"`
	Email        string `json:"email" binding:"required,email"`
	IsSuperAdmin bool   `json:"is_super_admin"`
}

type UpdateUserRequest struct {
	Name   string             `json:"name" binding:"omitempty,min=2,max=100"`
	Email  string             `json:"email" binding:"omitempty,email"`
	Active *bool              `json:"active"`
	Status *domain.UserStatus `json:"status" binding:"omitempty,oneof=ACTIVE INACTIVE SUSPENDED BLOCKED"`
}

type UpdateProfileRequest struct {
	FirstName            *string        `json:"first_name" binding:"omitempty,max=100"`
	LastName             *string        `json:"last_name" binding:"omitempty,max=100"`
	Phone                *string        `json:"phone" binding:"omitempty,max=30"`
	BirthDate            *time.Time     `json:"birth_date"`
	Gender               *domain.Gender `json:"gender" binding:"omitempty,oneof=MALE FEMALE NON_BINARY PREFER_NOT_TO_SAY"`
	City                 *string        `json:"city" binding:"omitempty,max=100"`
	Address              *string        `json:"address" binding:"omitempty,max=255"`
	PhotoURL             *string        `json:"photo_url" binding:"omitempty,max=500"`
	Bio                  *string        `json:"bio"`
	SociodemographicData map[string]any `json:"sociodemographic_data"`
}

type UserProfileResponse struct {
	UserID               string         `json:"user_id"`
	FirstName            string         `json:"first_name,omitempty"`
	LastName             string         `json:"last_name,omitempty"`
	Phone                string         `json:"phone,omitempty"`
	BirthDate            *time.Time     `json:"birth_date,omitempty"`
	Gender               domain.Gender  `json:"gender,omitempty"`
	City                 string         `json:"city,omitempty"`
	Address              string         `json:"address,omitempty"`
	PhotoURL             string         `json:"photo_url,omitempty"`
	Bio                  string         `json:"bio,omitempty"`
	SociodemographicData map[string]any `json:"sociodemographic_data,omitempty"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

type UserResponse struct {
	ID           string               `json:"id"`
	AuthID       string               `json:"auth_id"`
	Name         string               `json:"name"`
	Email        string               `json:"email"`
	AvatarURL    string               `json:"avatar_url,omitempty"`
	IsSuperAdmin bool                 `json:"is_super_admin"`
	Status       domain.UserStatus    `json:"status"`
	Active       bool                 `json:"active"`
	Profile      *UserProfileResponse `json:"profile,omitempty"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

func userResponse(user *domain.User) UserResponse {
	var profileResp *UserProfileResponse
	if user.Profile != nil {
		p := userProfileResponse(user.Profile)
		profileResp = &p
	}
	return UserResponse{
		ID:           user.ID,
		AuthID:       user.AuthID,
		Name:         user.Name,
		Email:        user.Email,
		AvatarURL:    user.AvatarURL,
		IsSuperAdmin: user.IsSuperAdmin,
		Status:       user.Status,
		Active:       user.Active,
		Profile:      profileResp,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func userProfileResponse(p *domain.UserProfile) UserProfileResponse {
	if p == nil {
		return UserProfileResponse{}
	}
	return UserProfileResponse{
		UserID:               p.UserID,
		FirstName:            p.FirstName,
		LastName:             p.LastName,
		Phone:                p.Phone,
		BirthDate:            p.BirthDate,
		Gender:               p.Gender,
		City:                 p.City,
		Address:              p.Address,
		PhotoURL:             p.PhotoURL,
		Bio:                  p.Bio,
		SociodemographicData: p.SociodemographicData,
		UpdatedAt:            p.UpdatedAt,
	}
}
