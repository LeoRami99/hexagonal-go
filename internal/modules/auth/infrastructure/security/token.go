package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"hexagonal-go-backend/internal/modules/auth/application"
	"hexagonal-go-backend/internal/modules/auth/domain"
)

type HMACTokenProvider struct{ secret []byte }

type payload struct {
	Subject  string `json:"sub"`
	TenantID string `json:"tid,omitempty"`
	Expires  int64  `json:"exp"`
	ID       string `json:"jti"`
}

func NewHMACTokenProvider(secret string) *HMACTokenProvider {
	return &HMACTokenProvider{secret: []byte(secret)}
}

func (p *HMACTokenProvider) Generate(claims application.TokenClaims) (string, error) {
	bytes, err := json.Marshal(payload{Subject: claims.Subject, TenantID: claims.TenantID, Expires: claims.ExpiresAt.Unix(), ID: claims.TokenID})
	if err != nil {
		return "", err
	}
	body := base64.RawURLEncoding.EncodeToString(bytes)
	return body + "." + p.sign(body), nil
}

func (p *HMACTokenProvider) Parse(token string) (application.TokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 || !hmac.Equal([]byte(p.sign(parts[0])), []byte(parts[1])) {
		return application.TokenClaims{}, domain.ErrInvalidToken
	}
	bytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return application.TokenClaims{}, domain.ErrInvalidToken
	}
	var value payload
	if json.Unmarshal(bytes, &value) != nil || value.Subject == "" || time.Now().Unix() >= value.Expires {
		return application.TokenClaims{}, domain.ErrInvalidToken
	}
	return application.TokenClaims{Subject: value.Subject, TenantID: value.TenantID, ExpiresAt: time.Unix(value.Expires, 0), TokenID: value.ID}, nil
}

func (p *HMACTokenProvider) sign(value string) string {
	hash := hmac.New(sha256.New, p.secret)
	_, _ = hash.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
}
