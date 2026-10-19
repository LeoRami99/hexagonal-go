package ent

import (
	"context"
	"errors"
	"fmt"
	"strings"

	entclient "hexagonal-go-backend/internal/ent"
	entuser "hexagonal-go-backend/internal/ent/user"
	entuserprofile "hexagonal-go-backend/internal/ent/userprofile"
	"hexagonal-go-backend/internal/modules/users/application"
	"hexagonal-go-backend/internal/modules/users/domain"

	"entgo.io/ent/dialect/sql"
)

type UserRepository struct {
	client *entclient.Client
}

func NewUserRepository(client *entclient.Client) *UserRepository {
	return &UserRepository{client: client}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	builder := r.client.User.Create().
		SetID(user.ID).
		SetAuthID(user.AuthID).
		SetName(user.Name).
		SetEmail(user.Email).
		SetIsSuperAdmin(user.IsSuperAdmin).
		SetActive(user.Active).
		SetCreatedAt(user.CreatedAt).
		SetUpdatedAt(user.UpdatedAt)
	if user.AvatarURL != "" {
		builder.SetAvatarURL(user.AvatarURL)
	}
	if user.Status != "" {
		builder.SetStatus(entuser.Status(user.Status))
	}
	_, err := builder.Save(ctx)
	return translateError(err)
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := r.client.User.Query().
		Where(entuser.IDEQ(id)).
		WithProfile().
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return toDomain(user), nil
}

func (r *UserRepository) FindByAuthID(ctx context.Context, authID string) (*domain.User, error) {
	user, err := r.client.User.Query().
		Where(entuser.AuthIDEQ(strings.TrimSpace(authID))).
		WithProfile().
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return toDomain(user), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := r.client.User.Query().
		Where(entuser.EmailEQ(strings.ToLower(strings.TrimSpace(email)))).
		WithProfile().
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return toDomain(user), nil
}

func (r *UserRepository) List(ctx context.Context, filter application.UserFilter) ([]domain.User, int, error) {
	query := r.client.User.Query().
		Where(entuser.IsSuperAdminEQ(false))
	if search := strings.TrimSpace(filter.Search); search != "" {
		query = query.Where(entuser.Or(entuser.NameContainsFold(search), entuser.EmailContainsFold(search)))
	}
	if filter.Status != "" {
		query = query.Where(entuser.StatusEQ(entuser.Status(filter.Status)))
	}
	if filter.Active != nil {
		query = query.Where(entuser.ActiveEQ(*filter.Active))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	order := sql.OrderAsc()
	if filter.Order == "desc" {
		order = sql.OrderDesc()
	}
	if filter.Sort == "name" {
		query = query.Order(entuser.ByName(order))
	} else {
		query = query.Order(entuser.ByCreatedAt(order))
	}

	users, err := query.
		WithProfile().
		Offset((filter.Page - 1) * filter.Limit).
		Limit(filter.Limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}

	result := make([]domain.User, 0, len(users))
	for _, user := range users {
		result = append(result, *toDomain(user))
	}
	return result, total, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	builder := r.client.User.UpdateOneID(user.ID).
		SetAuthID(user.AuthID).
		SetName(user.Name).
		SetEmail(user.Email).
		SetIsSuperAdmin(user.IsSuperAdmin).
		SetActive(user.Active).
		SetUpdatedAt(user.UpdatedAt)
	if user.AvatarURL != "" {
		builder.SetAvatarURL(user.AvatarURL)
	}
	if user.Status != "" {
		builder.SetStatus(entuser.Status(user.Status))
	}
	_, err := builder.Save(ctx)
	return translateError(err)
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	return translateError(r.client.User.DeleteOneID(id).Exec(ctx))
}

func (r *UserRepository) GetProfile(ctx context.Context, userID string) (*domain.UserProfile, error) {
	p, err := r.client.UserProfile.Query().
		Where(entuserprofile.UserIDEQ(userID)).
		Only(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return nil, nil
		}
		return nil, translateError(err)
	}
	return toDomainProfile(p), nil
}

func (r *UserRepository) UpsertProfile(ctx context.Context, profile *domain.UserProfile) error {
	existing, err := r.client.UserProfile.Query().
		Where(entuserprofile.UserIDEQ(profile.UserID)).
		Only(ctx)
	if err != nil && !entclient.IsNotFound(err) {
		return translateError(err)
	}

	if existing != nil {
		builder := r.client.UserProfile.UpdateOneID(existing.ID).
			SetUpdatedAt(profile.UpdatedAt)
		if profile.FirstName != "" {
			builder.SetFirstName(profile.FirstName)
		}
		if profile.LastName != "" {
			builder.SetLastName(profile.LastName)
		}
		if profile.Phone != "" {
			builder.SetPhone(profile.Phone)
		}
		if profile.BirthDate != nil {
			builder.SetBirthDate(*profile.BirthDate)
		}
		if profile.Gender != "" {
			builder.SetGender(entuserprofile.Gender(profile.Gender))
		}
		if profile.City != "" {
			builder.SetCity(profile.City)
		}
		if profile.Address != "" {
			builder.SetAddress(profile.Address)
		}
		if profile.PhotoURL != "" {
			builder.SetPhotoURL(profile.PhotoURL)
		}
		if profile.Bio != "" {
			builder.SetBio(profile.Bio)
		}
		if profile.SociodemographicData != nil {
			builder.SetSociodemographicData(profile.SociodemographicData)
		}
		_, err := builder.Save(ctx)
		return translateError(err)
	}

	createBuilder := r.client.UserProfile.Create().
		SetID(profile.ID).
		SetUserID(profile.UserID).
		SetCreatedAt(profile.CreatedAt).
		SetUpdatedAt(profile.UpdatedAt)
	if profile.FirstName != "" {
		createBuilder.SetFirstName(profile.FirstName)
	}
	if profile.LastName != "" {
		createBuilder.SetLastName(profile.LastName)
	}
	if profile.Phone != "" {
		createBuilder.SetPhone(profile.Phone)
	}
	if profile.BirthDate != nil {
		createBuilder.SetBirthDate(*profile.BirthDate)
	}
	if profile.Gender != "" {
		createBuilder.SetGender(entuserprofile.Gender(profile.Gender))
	}
	if profile.City != "" {
		createBuilder.SetCity(profile.City)
	}
	if profile.Address != "" {
		createBuilder.SetAddress(profile.Address)
	}
	if profile.PhotoURL != "" {
		createBuilder.SetPhotoURL(profile.PhotoURL)
	}
	if profile.Bio != "" {
		createBuilder.SetBio(profile.Bio)
	}
	if profile.SociodemographicData != nil {
		createBuilder.SetSociodemographicData(profile.SociodemographicData)
	}
	_, err = createBuilder.Save(ctx)
	return translateError(err)
}

func translateError(err error) error {
	if err == nil {
		return nil
	}
	if entclient.IsNotFound(err) {
		return domain.ErrUserNotFound
	}
	if entclient.IsConstraintError(err) {
		return domain.ErrEmailAlreadyExists
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return fmt.Errorf("users repository: %w", err)
}

var _ application.UserRepository = (*UserRepository)(nil)
