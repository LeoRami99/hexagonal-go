package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"hexagonal-go-backend/internal/modules/users/domain"
	"hexagonal-go-backend/internal/platform/identifier"
)

type userService struct {
	repo   UserRepository
	mailer Mailer
}

func NewUserService(repo UserRepository, mailer Mailer) UserService {
	return &userService{repo: repo, mailer: mailer}
}

func (s *userService) Create(ctx context.Context, in CreateUserInput) (*domain.User, error) {
	in.Name, in.Email, in.AuthID = strings.TrimSpace(in.Name), strings.ToLower(strings.TrimSpace(in.Email)), strings.TrimSpace(in.AuthID)
	if len(in.Name) < 2 || in.Email == "" {
		return nil, domain.ErrInvalidInput
	}
	if in.AuthID != "" {
		if existing, err := s.repo.FindByAuthID(ctx, in.AuthID); err == nil && existing != nil {
			return existing, nil
		}
	}
	if _, err := s.repo.FindByEmail(ctx, in.Email); err == nil {
		return nil, domain.ErrEmailAlreadyExists
	} else if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	now := time.Now().UTC()
	u := &domain.User{
		ID:           identifier.New(),
		AuthID:       in.AuthID,
		Name:         in.Name,
		Email:        in.Email,
		AvatarURL:    in.AvatarURL,
		IsSuperAdmin: in.IsSuperAdmin,
		Status:       domain.StatusActive,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	if s.mailer != nil {
		_ = s.mailer.Send(ctx, MailMessage{To: u.Email, Subject: "Welcome", Body: "Your account was created."})
	}
	return u, nil
}

func (s *userService) FindByID(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *userService) List(ctx context.Context, filter UserFilter) ([]domain.User, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	return s.repo.List(ctx, filter)
}

func (s *userService) Update(ctx context.Context, id string, in UpdateUserInput) (*domain.User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if name := strings.TrimSpace(in.Name); name != "" {
		u.Name = name
	}
	if email := strings.ToLower(strings.TrimSpace(in.Email)); email != "" && email != u.Email {
		if _, findErr := s.repo.FindByEmail(ctx, email); findErr == nil {
			return nil, domain.ErrEmailAlreadyExists
		} else if !errors.Is(findErr, domain.ErrUserNotFound) {
			return nil, findErr
		}
		u.Email = email
	}
	if in.AvatarURL != nil {
		u.AvatarURL = *in.AvatarURL
	}
	if in.Active != nil {
		u.Active = *in.Active
	}
	if in.Status != nil {
		u.Status = *in.Status
	}
	u.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *userService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *userService) GetProfile(ctx context.Context, userID string) (*domain.UserProfile, error) {
	if _, err := s.repo.FindByID(ctx, userID); err != nil {
		return nil, err
	}
	profile, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		profile = &domain.UserProfile{
			UserID: userID,
		}
	}
	return profile, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID string, in UpdateProfileInput) (*domain.UserProfile, error) {
	if _, err := s.repo.FindByID(ctx, userID); err != nil {
		return nil, err
	}
	profile, err := s.repo.GetProfile(ctx, userID)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}
	now := time.Now().UTC()
	if profile == nil {
		profile = &domain.UserProfile{
			ID:        identifier.New(),
			UserID:    userID,
			CreatedAt: now,
		}
	}
	if in.FirstName != nil {
		profile.FirstName = strings.TrimSpace(*in.FirstName)
	}
	if in.LastName != nil {
		profile.LastName = strings.TrimSpace(*in.LastName)
	}
	if in.Phone != nil {
		profile.Phone = strings.TrimSpace(*in.Phone)
	}
	if in.BirthDate != nil {
		profile.BirthDate = in.BirthDate
	}
	if in.Gender != nil {
		profile.Gender = *in.Gender
	}
	if in.City != nil {
		profile.City = strings.TrimSpace(*in.City)
	}
	if in.Address != nil {
		profile.Address = strings.TrimSpace(*in.Address)
	}
	if in.PhotoURL != nil {
		profile.PhotoURL = strings.TrimSpace(*in.PhotoURL)
	}
	if in.Bio != nil {
		profile.Bio = strings.TrimSpace(*in.Bio)
	}
	if in.SociodemographicData != nil {
		profile.SociodemographicData = in.SociodemographicData
	}
	profile.UpdatedAt = now

	if err := s.repo.UpsertProfile(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}
