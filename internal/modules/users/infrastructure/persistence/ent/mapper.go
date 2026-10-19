package ent

import (
	entuser "hexagonal-go-backend/internal/ent"
	"hexagonal-go-backend/internal/modules/users/domain"
)

func toDomain(user *entuser.User) *domain.User {
	if user == nil {
		return nil
	}
	status := domain.StatusActive
	if user.Status != "" {
		status = domain.UserStatus(user.Status)
	}
	avatar := ""
	if user.AvatarURL != nil {
		avatar = *user.AvatarURL
	}
	var profile *domain.UserProfile
	if user.Edges.Profile != nil {
		profile = toDomainProfile(user.Edges.Profile)
	}
	return &domain.User{
		ID:           user.ID,
		AuthID:       user.AuthID,
		Name:         user.Name,
		Email:        user.Email,
		AvatarURL:    avatar,
		IsSuperAdmin: user.IsSuperAdmin,
		Status:       status,
		Active:       user.Active,
		Profile:      profile,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func toDomainProfile(p *entuser.UserProfile) *domain.UserProfile {
	if p == nil {
		return nil
	}
	fn, ln, phone, city, addr, photo, bio := "", "", "", "", "", "", ""
	if p.FirstName != nil {
		fn = *p.FirstName
	}
	if p.LastName != nil {
		ln = *p.LastName
	}
	if p.Phone != nil {
		phone = *p.Phone
	}
	if p.City != nil {
		city = *p.City
	}
	if p.Address != nil {
		addr = *p.Address
	}
	if p.PhotoURL != nil {
		photo = *p.PhotoURL
	}
	var gender domain.Gender
	if p.Gender != nil {
		gender = domain.Gender(*p.Gender)
	}

	return &domain.UserProfile{
		ID:                   p.ID,
		UserID:               p.UserID,
		FirstName:            fn,
		LastName:             ln,
		Phone:                phone,
		BirthDate:            p.BirthDate,
		Gender:               gender,
		City:                 city,
		Address:              addr,
		PhotoURL:             photo,
		Bio:                  bio,
		SociodemographicData: p.SociodemographicData,
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
	}
}
