package domain

import "time"

type UserStatus string

const (
	StatusActive    UserStatus = "ACTIVE"
	StatusInactive  UserStatus = "INACTIVE"
	StatusSuspended UserStatus = "SUSPENDED"
	StatusBlocked   UserStatus = "BLOCKED"
)

type Gender string

const (
	GenderMale           Gender = "MALE"
	GenderFemale         Gender = "FEMALE"
	GenderNonBinary      Gender = "NON_BINARY"
	GenderPreferNotToSay Gender = "PREFER_NOT_TO_SAY"
)

type User struct {
	ID           string
	AuthID       string
	Name         string
	Email        string
	AvatarURL    string
	IsSuperAdmin bool
	Status       UserStatus
	Active       bool
	Profile      *UserProfile
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserProfile struct {
	ID                   string
	UserID               string
	FirstName            string
	LastName             string
	Phone                string
	BirthDate            *time.Time
	Gender               Gender
	City                 string
	Address              string
	PhotoURL             string
	Bio                  string
	SociodemographicData map[string]any
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
