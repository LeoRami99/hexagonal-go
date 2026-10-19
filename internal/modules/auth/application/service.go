package application

import (
	"context"
	"time"

	"hexagonal-go-backend/internal/modules/auth/domain"
	users "hexagonal-go-backend/internal/modules/users/application"
	userdomain "hexagonal-go-backend/internal/modules/users/domain"
	"hexagonal-go-backend/internal/platform/identifier"
)

type authService struct {
	users                 users.UserRepository
	tokens                TokenProvider
	cache                 Cache
	ctxRepo               AuthContextRepository
	idp                   IdentityProvider
	accessTTL, refreshTTL time.Duration
}

func NewAuthService(userRepository users.UserRepository, tokens TokenProvider, cache Cache, accessTTL, refreshTTL time.Duration, ctxRepo AuthContextRepository, idp ...IdentityProvider) AuthService {
	var provider IdentityProvider
	if len(idp) > 0 {
		provider = idp[0]
	}
	return &authService{
		users:      userRepository,
		tokens:     tokens,
		cache:      cache,
		ctxRepo:    ctxRepo,
		idp:        provider,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (s *authService) Login(ctx context.Context, email, password string) (TokenPair, error) {
	if s.idp != nil {
		pair, err := s.idp.Authenticate(ctx, email, password)
		if err != nil {
			return TokenPair{}, err
		}

		claims, err := s.tokens.Parse(pair.AccessToken)
		if err != nil || claims.Subject == "" {
			return TokenPair{}, domain.ErrInvalidToken
		}
		if err := s.provision(ctx, claims, email); err != nil {
			return TokenPair{}, err
		}

		return pair, nil
	}
	return TokenPair{}, domain.ErrUnauthorized
}

// provision garantiza que exista el usuario local asociado a la identidad del IdP.
func (s *authService) provision(ctx context.Context, claims TokenClaims, email string) error {
	u, err := s.users.FindByAuthID(ctx, claims.Subject)
	if err == nil {
		if !u.Active {
			return domain.ErrAccountDisabled
		}
		return nil
	}

	// Usuario pre-existente (ej. sembrado) que aún no está vinculado al IdP.
	if existing, errEmail := s.users.FindByEmail(ctx, email); errEmail == nil {
		if !existing.Active {
			return domain.ErrAccountDisabled
		}
		existing.AuthID = claims.Subject
		return s.users.Update(ctx, existing)
	}

	now := time.Now().UTC()
	newUser := &userdomain.User{
		ID:        identifier.New(),
		AuthID:    claims.Subject,
		Name:      email,
		Email:     email,
		Status:    userdomain.StatusActive,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.users.Create(ctx, newUser); err != nil {
		if _, errFind := s.users.FindByAuthID(ctx, claims.Subject); errFind == nil {
			return nil
		}
		return err
	}
	return nil
}

func (s *authService) Refresh(ctx context.Context, token string) (TokenPair, error) {
	if s.idp != nil {
		pair, err := s.idp.Refresh(ctx, token)
		if err != nil {
			return TokenPair{}, err
		}
		if claims, errParse := s.tokens.Parse(pair.AccessToken); errParse == nil && claims.Subject != "" {
			u, errFind := s.users.FindByAuthID(ctx, claims.Subject)
			if errFind == nil && (!u.Active || u.Status != userdomain.StatusActive) {
				return TokenPair{}, domain.ErrUnauthorized
			}
		}
		return pair, nil
	}

	claims, err := s.tokens.Parse(token)
	if err != nil {
		return TokenPair{}, domain.ErrInvalidToken
	}
	if _, err = s.cache.Get(ctx, "refresh:"+claims.TokenID); err != nil {
		return TokenPair{}, domain.ErrInvalidToken
	}
	_ = s.cache.Delete(ctx, "refresh:"+claims.TokenID)
	u, err := s.users.FindByID(ctx, claims.Subject)
	if err != nil {
		u, err = s.users.FindByAuthID(ctx, claims.Subject)
	}
	if err != nil || !u.Active {
		return TokenPair{}, domain.ErrUnauthorized
	}
	return s.issue(ctx, u)
}

func (s *authService) Logout(ctx context.Context, token string) error {
	if s.idp != nil {
		return s.idp.Logout(ctx, token)
	}
	claims, err := s.tokens.Parse(token)
	if err != nil {
		return domain.ErrInvalidToken
	}
	return s.cache.Delete(ctx, "refresh:"+claims.TokenID)
}

func (s *authService) GetSessionContext(ctx context.Context, userID string) (*SessionContext, error) {
	if s.ctxRepo != nil {
		return s.ctxRepo.GetSessionContext(ctx, userID)
	}
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &SessionContext{
		User: UserInfo{
			ID:           u.ID,
			AuthID:       u.AuthID,
			Name:         u.Name,
			Email:        u.Email,
			AvatarURL:    u.AvatarURL,
			IsSuperAdmin: u.IsSuperAdmin,
		},
		Tenants:  []TenantContext{},
		Projects: []ProjectMembershipContext{},
	}, nil
}

func (s *authService) issue(ctx context.Context, user *userdomain.User) (TokenPair, error) {
	now := time.Now().UTC()
	accessExp, refreshExp := now.Add(s.accessTTL), now.Add(s.refreshTTL)
	accessID, refreshID := identifier.New(), identifier.New()
	accessToken, err := s.tokens.Generate(TokenClaims{Subject: user.ID, ExpiresAt: accessExp, TokenID: accessID})
	if err != nil {
		return TokenPair{}, err
	}
	refreshToken, err := s.tokens.Generate(TokenClaims{Subject: user.ID, ExpiresAt: refreshExp, TokenID: refreshID})
	if err != nil {
		return TokenPair{}, err
	}
	if err = s.cache.Set(ctx, "refresh:"+refreshID, []byte(user.ID), s.refreshTTL); err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: accessToken, RefreshToken: refreshToken, AccessExpiresAt: accessExp, RefreshExpiresAt: refreshExp}, nil
}
