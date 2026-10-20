package infrastructure

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"hexagonal-go-backend/internal/modules/auth/application"
	"hexagonal-go-backend/internal/modules/auth/domain"
)

type ZitadelClient struct {
	baseURL     string
	projectID   string
	clientID    string
	redirectURI string
	pat         string
	httpClient  *http.Client
	authClient  *http.Client
}

func NewZitadelClient(baseURL, projectID, clientID, redirectURI, pat string) *ZitadelClient {
	return &ZitadelClient{
		baseURL:     strings.TrimRight(baseURL, "/"),
		projectID:   projectID,
		clientID:    clientID,
		redirectURI: redirectURI,
		pat:         pat,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		authClient: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// Session API v2 models
type createSessionCheckUser struct {
	LoginName string `json:"loginName"`
}

type createSessionCheckPassword struct {
	Password string `json:"password"`
}

type createSessionChecks struct {
	User     *createSessionCheckUser     `json:"user,omitempty"`
	Password *createSessionCheckPassword `json:"password,omitempty"`
}

type createSessionRequest struct {
	Checks createSessionChecks `json:"checks"`
}

type createSessionResponse struct {
	SessionID    string `json:"sessionId"`
	SessionToken string `json:"sessionToken"`
}

type linkSessionRequest struct {
	Session struct {
		SessionID    string `json:"sessionId"`
		SessionToken string `json:"sessionToken"`
	} `json:"session"`
}

type linkSessionResponse struct {
	CallbackURL string `json:"callbackUrl"`
}

type oidcTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	IDToken      string `json:"id_token"`
}

// Authenticate implementa authapp.IdentityProvider mediante flujo OIDC Auth Code + PKCE + Session API
func (c *ZitadelClient) Authenticate(ctx context.Context, username, password string) (application.TokenPair, error) {
	// 1. Generar PKCE verifier, challenge y state
	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		return application.TokenPair{}, fmt.Errorf("generate code verifier: %w", err)
	}
	codeChallenge := generateCodeChallenge(codeVerifier)
	state, err := generateState()
	if err != nil {
		return application.TokenPair{}, fmt.Errorf("generate state: %w", err)
	}

	// 2. Iniciar solicitud de autorización OIDC
	authRequestID, err := c.startAuthRequest(ctx, state, codeChallenge)
	if err != nil {
		return application.TokenPair{}, fmt.Errorf("start auth request: %w", err)
	}

	// 3. Crear sesión con credenciales de usuario en la Session API
	sessionID, sessionToken, err := c.createSession(ctx, username, password)
	if err != nil {
		return application.TokenPair{}, err
	}

	// 4. Vincular sesión al auth request OIDC y extraer código de autorización
	code, err := c.finalizeAuthRequest(ctx, authRequestID, sessionID, sessionToken, state)
	if err != nil {
		return application.TokenPair{}, fmt.Errorf("finalize auth request: %w", err)
	}

	// 5. Canjear código por tokens JWT
	return c.exchangeCode(ctx, code, codeVerifier)
}

func (c *ZitadelClient) startAuthRequest(ctx context.Context, state, codeChallenge string) (string, error) {
	authURL := fmt.Sprintf("%s/oauth/v2/authorize", c.baseURL)

	scope := "openid profile email offline_access"
	if c.projectID != "" {
		scope = fmt.Sprintf("openid profile email offline_access urn:zitadel:iam:org:project:id:%s:aud", c.projectID)
	}

	q := url.Values{}
	q.Set("client_id", c.clientID)
	q.Set("redirect_uri", c.redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", scope)
	q.Set("state", state)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")

	fullURL := fmt.Sprintf("%s?%s", authURL, q.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.authClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", domain.ErrIdentityProviderUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusSeeOther && resp.StatusCode != http.StatusTemporaryRedirect {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected authorize status %d: %s", resp.StatusCode, string(body))
	}

	loc := resp.Header.Get("Location")
	if loc == "" {
		return "", fmt.Errorf("missing Location header in authorize response")
	}

	locURL, err := url.Parse(loc)
	if err != nil {
		return "", fmt.Errorf("parse redirect location: %w", err)
	}

	authRequestID := locURL.Query().Get("authRequest")
	if authRequestID == "" {
		return "", fmt.Errorf("authRequest parameter missing from redirect location: %s", loc)
	}

	return authRequestID, nil
}

func (c *ZitadelClient) createSession(ctx context.Context, username, password string) (string, string, error) {
	sessionURL := fmt.Sprintf("%s/v2/sessions", c.baseURL)

	sessionReqBody := createSessionRequest{
		Checks: createSessionChecks{
			User:     &createSessionCheckUser{LoginName: username},
			Password: &createSessionCheckPassword{Password: password},
		},
	}
	bodyBytes, err := json.Marshal(sessionReqBody)
	if err != nil {
		return "", "", fmt.Errorf("marshal session request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, sessionURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", "", fmt.Errorf("create session request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.pat != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.pat))
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		slog.Error("zitadel unreachable", "error", err)
		return "", "", fmt.Errorf("%w: %v", domain.ErrIdentityProviderUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBytes, _ := io.ReadAll(resp.Body)
		slog.Warn("zitadel authentication failed", "status", resp.StatusCode, "body", string(respBytes), "username", username)
		return "", "", mapSessionError(resp.StatusCode, respBytes)
	}

	var sessionRes createSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sessionRes); err != nil {
		return "", "", fmt.Errorf("decode session response: %w", err)
	}

	return sessionRes.SessionID, sessionRes.SessionToken, nil
}

func (c *ZitadelClient) finalizeAuthRequest(ctx context.Context, authRequestID, sessionID, sessionToken, expectedState string) (string, error) {
	linkURL := fmt.Sprintf("%s/v2/oidc/auth_requests/%s", c.baseURL, authRequestID)

	linkReq := linkSessionRequest{}
	linkReq.Session.SessionID = sessionID
	linkReq.Session.SessionToken = sessionToken

	bodyBytes, err := json.Marshal(linkReq)
	if err != nil {
		return "", fmt.Errorf("marshal link request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, linkURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.pat != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.pat))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", domain.ErrIdentityProviderUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBytes, _ := io.ReadAll(resp.Body)
		slog.Warn("zitadel link session failed", "status", resp.StatusCode, "body", string(respBytes))
		return "", fmt.Errorf("link session returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var linkRes linkSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&linkRes); err != nil {
		return "", fmt.Errorf("decode link session response: %w", err)
	}

	cbURL, err := url.Parse(linkRes.CallbackURL)
	if err != nil {
		return "", fmt.Errorf("parse callback url: %w", err)
	}

	state := cbURL.Query().Get("state")
	if state != expectedState {
		return "", fmt.Errorf("state mismatch in callback: expected %s, got %s", expectedState, state)
	}

	code := cbURL.Query().Get("code")
	if code == "" {
		return "", fmt.Errorf("authorization code missing from callback: %s", linkRes.CallbackURL)
	}

	return code, nil
}

func (c *ZitadelClient) exchangeCode(ctx context.Context, code, codeVerifier string) (application.TokenPair, error) {
	tokenURL := fmt.Sprintf("%s/oauth/v2/token", c.baseURL)

	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", code)
	formData.Set("redirect_uri", c.redirectURI)
	formData.Set("client_id", c.clientID)
	formData.Set("code_verifier", codeVerifier)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return application.TokenPair{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return application.TokenPair{}, fmt.Errorf("%w: %v", domain.ErrIdentityProviderUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		slog.Warn("zitadel code exchange failed", "status", resp.StatusCode, "body", string(respBytes))
		return application.TokenPair{}, fmt.Errorf("token exchange returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var tokenRes oidcTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
		return application.TokenPair{}, fmt.Errorf("decode token response: %w", err)
	}

	now := time.Now().UTC()
	accessTTL := time.Duration(tokenRes.ExpiresIn) * time.Second
	if accessTTL <= 0 {
		accessTTL = 1 * time.Hour
	}

	return application.TokenPair{
		AccessToken:      tokenRes.AccessToken,
		RefreshToken:     tokenRes.RefreshToken,
		AccessExpiresAt:  now.Add(accessTTL),
		RefreshExpiresAt: now.Add(7 * 24 * time.Hour),
	}, nil
}

func (c *ZitadelClient) Refresh(ctx context.Context, refreshToken string) (application.TokenPair, error) {
	tokenURL := fmt.Sprintf("%s/oauth/v2/token", c.baseURL)

	formData := url.Values{}
	formData.Set("grant_type", "refresh_token")
	formData.Set("refresh_token", refreshToken)
	formData.Set("client_id", c.clientID)
	formData.Set("redirect_uri", c.redirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return application.TokenPair{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return application.TokenPair{}, fmt.Errorf("%w: %v", domain.ErrIdentityProviderUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return application.TokenPair{}, domain.ErrInvalidToken
	}

	var tokenRes oidcTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
		return application.TokenPair{}, fmt.Errorf("decode refresh token response: %w", err)
	}

	now := time.Now().UTC()
	accessTTL := time.Duration(tokenRes.ExpiresIn) * time.Second
	if accessTTL <= 0 {
		accessTTL = 1 * time.Hour
	}

	newRefreshToken := tokenRes.RefreshToken
	if newRefreshToken == "" {
		newRefreshToken = refreshToken
	}

	return application.TokenPair{
		AccessToken:      tokenRes.AccessToken,
		RefreshToken:     newRefreshToken,
		AccessExpiresAt:  now.Add(accessTTL),
		RefreshExpiresAt: now.Add(7 * 24 * time.Hour),
	}, nil
}

func (c *ZitadelClient) Logout(ctx context.Context, token string) error {
	revokeURL := fmt.Sprintf("%s/oauth/v2/revoke", c.baseURL)

	formData := url.Values{}
	formData.Set("token", token)
	formData.Set("client_id", c.clientID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, revokeURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrIdentityProviderUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return domain.ErrInvalidToken
	}
	return nil
}

type zitadelError struct {
	Message string `json:"message"`
}

func mapSessionError(status int, body []byte) error {
	var e zitadelError
	_ = json.Unmarshal(body, &e)
	msg := strings.ToLower(e.Message)

	switch {
	case status >= 500:
		return fmt.Errorf("%w: zitadel status %d", domain.ErrIdentityProviderUnavailable, status)
	case status == http.StatusUnauthorized || status == http.StatusForbidden || strings.Contains(msg, "membership not found"):
		return fmt.Errorf("zitadel rejected the service credentials (status %d): %s", status, e.Message)
	case strings.Contains(msg, "locked"):
		return domain.ErrAccountLocked
	case strings.Contains(msg, "inactive"), strings.Contains(msg, "not active"), strings.Contains(msg, "deactivated"):
		return domain.ErrAccountDisabled
	case status == http.StatusBadRequest, status == http.StatusNotFound:
		return domain.ErrInvalidCredentials
	default:
		return fmt.Errorf("zitadel returned status %d: %s", status, e.Message)
	}
}

// User API Request / Response models
type CreateZitadelUserRequest struct {
	Username string `json:"username,omitempty"`
	Profile  struct {
		GivenName   string `json:"givenName"`
		FamilyName  string `json:"familyName"`
		DisplayName string `json:"displayName"`
	} `json:"profile"`
	Email struct {
		Email      string `json:"email"`
		IsVerified bool   `json:"isVerified"`
	} `json:"email"`
	Password struct {
		Password       string `json:"password"`
		ChangeRequired bool   `json:"changeRequired"`
	} `json:"password"`
}

type CreateZitadelUserResponse struct {
	UserID string `json:"userId"`
}

type SearchUserResponse struct {
	Result []struct {
		UserID   string `json:"userId"`
		Username string `json:"username"`
		Human    struct {
			Email struct {
				Email string `json:"email"`
			} `json:"email"`
		} `json:"human"`
	} `json:"result"`
}

func (c *ZitadelClient) CreateUser(ctx context.Context, email, name, password string) (string, error) {
	url := fmt.Sprintf("%s/v2/users/human", c.baseURL)

	req := CreateZitadelUserRequest{}
	req.Username = email
	req.Profile.GivenName = name
	req.Profile.FamilyName = name
	req.Profile.DisplayName = name
	req.Email.Email = email
	req.Email.IsVerified = true
	req.Password.Password = password
	req.Password.ChangeRequired = false

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal user request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create user http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.pat != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.pat))
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("execute user creation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("zitadel create user returned %d: %s", resp.StatusCode, string(respBytes))
	}

	var res CreateZitadelUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("decode create user response: %w", err)
	}

	return res.UserID, nil
}

func (c *ZitadelClient) FindUser(ctx context.Context, emailOrUsername string) (string, error) {
	url := fmt.Sprintf("%s/v2/users", c.baseURL)

	searchBody := map[string]interface{}{
		"query": map[string]interface{}{
			"limit": 1,
		},
		"queries": []map[string]interface{}{
			{
				"emailQuery": map[string]interface{}{
					"emailAddress": emailOrUsername,
				},
			},
		},
	}
	bodyBytes, err := json.Marshal(searchBody)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.pat != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.pat))
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("zitadel find user returned %d", resp.StatusCode)
	}

	var searchRes SearchUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchRes); err != nil {
		return "", err
	}

	if len(searchRes.Result) > 0 {
		return searchRes.Result[0].UserID, nil
	}

	return "", nil
}

func (c *ZitadelClient) SetUserPassword(ctx context.Context, userID, password string) error {
	url := fmt.Sprintf("%s/v2/users/%s/password", c.baseURL, userID)

	body := map[string]interface{}{
		"newPassword": map[string]interface{}{
			"password":       password,
			"changeRequired": false,
		},
	}
	bodyBytes, _ := json.Marshal(body)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.pat != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.pat))
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("zitadel set password returned %d: %s", resp.StatusCode, string(respBytes))
	}

	return nil
}

// PKCE Helpers
func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

var _ application.IdentityProvider = (*ZitadelClient)(nil)
