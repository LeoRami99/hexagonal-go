package infrastructure

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"hexagonal-go-backend/internal/modules/auth/domain"
)



func TestMapSessionError(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		expected error
	}{
		{
			name:     "invalid credentials 400",
			status:   http.StatusBadRequest,
			body:     `{"message":"invalid password"}`,
			expected: domain.ErrInvalidCredentials,
		},
		{
			name:     "user not found 404",
			status:   http.StatusNotFound,
			body:     `{"message":"User could not be found"}`,
			expected: domain.ErrInvalidCredentials,
		},
		{
			name:     "account locked",
			status:   http.StatusBadRequest,
			body:     `{"message":"user account is locked"}`,
			expected: domain.ErrAccountLocked,
		},
		{
			name:     "account deactivated",
			status:   http.StatusBadRequest,
			body:     `{"message":"account deactivated"}`,
			expected: domain.ErrAccountDisabled,
		},
		{
			name:     "server error 500",
			status:   http.StatusInternalServerError,
			body:     `{"message":"internal error"}`,
			expected: domain.ErrIdentityProviderUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mapSessionError(tt.status, []byte(tt.body))
			if !errors.Is(err, tt.expected) {
				t.Errorf("expected error %v, got %v", tt.expected, err)
			}
		})
	}
}

func TestZitadelClient_FlowMock(t *testing.T) {
	var capturedState string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/v2/authorize":
			capturedState = r.URL.Query().Get("state")
			loc := "/ui/v2/login/login?authRequest=V2_mock123"
			http.Redirect(w, r, loc, http.StatusFound)

		case "/v2/sessions":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"sessionId":"sess_1","sessionToken":"tok_1"}`))

		case "/v2/oidc/auth_requests/V2_mock123":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"callbackUrl":"http://localhost:8085/auth/callback?code=mock_code&state=` + capturedState + `"}`))

		case "/oauth/v2/token":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"access_token":"valid_jwt_token","refresh_token":"valid_refresh","expires_in":3600}`))

		case "/oauth/v2/revoke":
			w.WriteHeader(http.StatusOK)

		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	client := NewZitadelClient(ts.URL, "proj_1", "client_1", "http://localhost:8085/auth/callback", "pat_1")

	// 1. Authenticate happy path
	pair, err := client.Authenticate(context.Background(), "user@example.com", "Password123!")
	if err != nil {
		t.Fatalf("unexpected authenticate error: %v", err)
	}
	if pair.AccessToken != "valid_jwt_token" {
		t.Errorf("expected access token valid_jwt_token, got %s", pair.AccessToken)
	}
	if pair.RefreshToken != "valid_refresh" {
		t.Errorf("expected refresh token valid_refresh, got %s", pair.RefreshToken)
	}

	// 2. Refresh
	refreshed, err := client.Refresh(context.Background(), pair.RefreshToken)
	if err != nil {
		t.Fatalf("unexpected refresh error: %v", err)
	}
	if refreshed.AccessToken != "valid_jwt_token" {
		t.Errorf("expected access token valid_jwt_token on refresh, got %s", refreshed.AccessToken)
	}

	// 3. Logout
	if err := client.Logout(context.Background(), pair.RefreshToken); err != nil {
		t.Fatalf("unexpected logout error: %v", err)
	}
}
