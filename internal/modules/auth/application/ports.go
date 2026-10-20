package application

import (
	"context"
	"time"
)

type TokenClaims struct {
	Subject   string
	TenantID  string
	ExpiresAt time.Time
	TokenID   string
}

type TokenProvider interface {
	Generate(TokenClaims) (string, error)
	Parse(string) (TokenClaims, error)
}

type TokenPair struct {
	AccessToken, RefreshToken         string
	AccessExpiresAt, RefreshExpiresAt time.Time
}

type UserInfo struct {
	ID           string `json:"id"`
	AuthID       string `json:"auth_id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	AvatarURL    string `json:"avatar_url,omitempty"`
	IsSuperAdmin bool   `json:"is_super_admin"`
}

type TenantContext struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Status string `json:"status"`
}

type ProjectMembershipContext struct {
	ProjectID   string   `json:"project_id"`
	ProjectName string   `json:"project_name"`
	ProjectSlug string   `json:"project_slug"`
	TenantID    string   `json:"tenant_id"`
	Roles       []string `json:"roles"`
}

type SessionContext struct {
	User     UserInfo                   `json:"user"`
	Tenants  []TenantContext            `json:"tenants"`
	Projects []ProjectMembershipContext `json:"projects"`
}

type AuthContextRepository interface {
	GetSessionContext(ctx context.Context, userID string) (*SessionContext, error)
}

type AuthService interface {
	Login(context.Context, string, string) (TokenPair, error)
	Refresh(context.Context, string) (TokenPair, error)
	Logout(context.Context, string) error
	GetSessionContext(context.Context, string) (*SessionContext, error)
}

type IdentityProvider interface {
	Authenticate(ctx context.Context, username, password string) (TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (TokenPair, error)
	Logout(ctx context.Context, token string) error
}

type Cache interface {
	Get(context.Context, string) ([]byte, error)
	Set(context.Context, string, []byte, time.Duration) error
	Delete(context.Context, string) error
}
