package security

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	authapp "hexagonal-go-backend/internal/modules/auth/application"
	"hexagonal-go-backend/internal/modules/auth/domain"
)

type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Typ string `json:"typ"`
}

type jwtPayload struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	PreferredUser string `json:"preferred_username"`
	ExpiresAt     int64  `json:"exp"`
	IssuedAt      int64  `json:"iat"`
	JTI           string `json:"jti"`
}

type ZitadelTokenProvider struct {
	jwksURL    string
	httpClient *http.Client

	mu        sync.RWMutex
	keys      map[string]crypto.PublicKey
	fetchedAt time.Time
}

func NewZitadelTokenProvider(jwksURL string) *ZitadelTokenProvider {
	return &ZitadelTokenProvider{
		jwksURL: jwksURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		keys: make(map[string]crypto.PublicKey),
	}
}

func (p *ZitadelTokenProvider) Generate(claims authapp.TokenClaims) (string, error) {
	return "", errors.New("token generation is managed externally by ZITADEL IdP")
}

func (p *ZitadelTokenProvider) Parse(tokenStr string) (authapp.TokenClaims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return authapp.TokenClaims{}, domain.ErrInvalidToken
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return authapp.TokenClaims{}, domain.ErrInvalidToken
	}

	var hdr jwtHeader
	if err := json.Unmarshal(headerBytes, &hdr); err != nil {
		return authapp.TokenClaims{}, domain.ErrInvalidToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return authapp.TokenClaims{}, domain.ErrInvalidToken
	}

	var payload jwtPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return authapp.TokenClaims{}, domain.ErrInvalidToken
	}

	if payload.ExpiresAt > 0 && time.Now().Unix() >= payload.ExpiresAt {
		return authapp.TokenClaims{}, domain.ErrInvalidToken
	}

	if payload.Subject == "" {
		return authapp.TokenClaims{}, domain.ErrInvalidToken
	}

	// Verify cryptographic signature against ZITADEL JWKS
	if p.jwksURL != "" {
		pubKey, err := p.getKey(hdr.Kid)
		if err != nil || pubKey == nil {
			return authapp.TokenClaims{}, fmt.Errorf("%w: public key lookup failed", domain.ErrInvalidToken)
		}
		signedContent := parts[0] + "." + parts[1]
		sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
		if err != nil || !verifySignature(hdr.Alg, pubKey, []byte(signedContent), sigBytes) {
			return authapp.TokenClaims{}, domain.ErrInvalidToken
		}
	}

	return authapp.TokenClaims{
		Subject:   payload.Subject,
		ExpiresAt: time.Unix(payload.ExpiresAt, 0),
		TokenID:   payload.JTI,
	}, nil
}

func (p *ZitadelTokenProvider) getKey(kid string) (crypto.PublicKey, error) {
	p.mu.RLock()
	if time.Since(p.fetchedAt) < 1*time.Hour {
		if key, found := p.keys[kid]; found {
			p.mu.RUnlock()
			return key, nil
		}
		if len(p.keys) > 0 && kid == "" {
			for _, k := range p.keys {
				p.mu.RUnlock()
				return k, nil
			}
		}
	}
	p.mu.RUnlock()

	if err := p.refreshKeys(); err != nil {
		return nil, err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()
	if key, found := p.keys[kid]; found {
		return key, nil
	}
	if kid == "" {
		for _, k := range p.keys {
			return k, nil
		}
	}
	return nil, fmt.Errorf("key kid %s not found in JWKS", kid)
}

func (p *ZitadelTokenProvider) refreshKeys() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.jwksURL, nil)
	if err != nil {
		return err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jwks fetch returned status: %d", resp.StatusCode)
	}

	var res JWKSResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}

	newKeys := make(map[string]crypto.PublicKey)
	for _, jwk := range res.Keys {
		pubKey, err := parseJWK(jwk)
		if err == nil && pubKey != nil {
			newKeys[jwk.Kid] = pubKey
		}
	}

	p.keys = newKeys
	p.fetchedAt = time.Now()
	return nil
}

func parseJWK(jwk JWK) (crypto.PublicKey, error) {
	switch jwk.Kty {
	case "RSA":
		nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
		if err != nil {
			return nil, err
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
		if err != nil {
			return nil, err
		}
		var eInt int
		for _, b := range eBytes {
			eInt = (eInt << 8) | int(b)
		}
		return &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: eInt,
		}, nil
	case "EC":
		xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
		if err != nil {
			return nil, err
		}
		yBytes, err := base64.RawURLEncoding.DecodeString(jwk.Y)
		if err != nil {
			return nil, err
		}
		return &ecdsa.PublicKey{
			X: new(big.Int).SetBytes(xBytes),
			Y: new(big.Int).SetBytes(yBytes),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported key type: %s", jwk.Kty)
	}
}

func verifySignature(alg string, pubKey crypto.PublicKey, data, sig []byte) bool {
	hashed := sha256.Sum256(data)
	switch k := pubKey.(type) {
	case *rsa.PublicKey:
		return rsa.VerifyPKCS1v15(k, crypto.SHA256, hashed[:], sig) == nil
	case *ecdsa.PublicKey:
		return ecdsa.VerifyASN1(k, hashed[:], sig)
	default:
		return false
	}
}
