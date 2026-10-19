package memory

import (
	"context"
	"sort"
	"strings"
	"sync"

	"hexagonal-go-backend/internal/modules/users/application"
	"hexagonal-go-backend/internal/modules/users/domain"
)

type UserRepository struct {
	mu       sync.RWMutex
	users    map[string]domain.User
	profiles map[string]domain.UserProfile
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users:    make(map[string]domain.User),
		profiles: make(map[string]domain.UserProfile),
	}
}

func (r *UserRepository) Create(_ context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.users {
		if existing.Email == user.Email {
			return domain.ErrEmailAlreadyExists
		}
	}
	r.users[user.ID] = *user
	return nil
}

func (r *UserRepository) FindByID(_ context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	u := clone(user)
	if p, found := r.profiles[id]; found {
		pCopy := p
		u.Profile = &pCopy
	}
	return u, nil
}

func (r *UserRepository) FindByAuthID(_ context.Context, authID string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	authID = strings.TrimSpace(authID)
	for _, user := range r.users {
		if user.AuthID == authID {
			u := clone(user)
			if p, found := r.profiles[user.ID]; found {
				pCopy := p
				u.Profile = &pCopy
			}
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (r *UserRepository) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	email = strings.ToLower(strings.TrimSpace(email))
	for _, user := range r.users {
		if user.Email == email {
			u := clone(user)
			if p, found := r.profiles[user.ID]; found {
				pCopy := p
				u.Profile = &pCopy
			}
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (r *UserRepository) List(_ context.Context, filter application.UserFilter) ([]domain.User, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	all := make([]domain.User, 0, len(r.users))
	query := strings.ToLower(filter.Search)
	for _, user := range r.users {
		if user.IsSuperAdmin {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(user.Name+" "+user.Email), query) {
			continue
		}
		if filter.Status != "" && user.Status != filter.Status {
			continue
		}
		if filter.Active != nil && user.Active != *filter.Active {
			continue
		}
		u := user
		if p, found := r.profiles[user.ID]; found {
			pCopy := p
			u.Profile = &pCopy
		}
		all = append(all, u)
	}
	sort.Slice(all, func(i, j int) bool {
		less := all[i].CreatedAt.Before(all[j].CreatedAt)
		if filter.Sort == "name" {
			less = all[i].Name < all[j].Name
		}
		if filter.Order == "desc" {
			return !less
		}
		return less
	})
	total := len(all)
	start := (filter.Page - 1) * filter.Limit
	if start > total {
		start = total
	}
	end := start + filter.Limit
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

func (r *UserRepository) Update(_ context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[user.ID]; !ok {
		return domain.ErrUserNotFound
	}
	for id, existing := range r.users {
		if id != user.ID && existing.Email == user.Email {
			return domain.ErrEmailAlreadyExists
		}
	}
	r.users[user.ID] = *user
	return nil
}

func (r *UserRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[id]; !ok {
		return domain.ErrUserNotFound
	}
	delete(r.users, id)
	delete(r.profiles, id)
	return nil
}

func (r *UserRepository) GetProfile(_ context.Context, userID string) (*domain.UserProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if p, found := r.profiles[userID]; found {
		pCopy := p
		return &pCopy, nil
	}
	return nil, nil
}

func (r *UserRepository) UpsertProfile(_ context.Context, profile *domain.UserProfile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.profiles[profile.UserID] = *profile
	return nil
}

func clone(user domain.User) *domain.User {
	copy := user
	return &copy
}
