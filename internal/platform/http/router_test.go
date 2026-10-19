package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authapp "hexagonal-go-backend/internal/modules/auth/application"
	authhttp "hexagonal-go-backend/internal/modules/auth/delivery/http/v1"
	cache "hexagonal-go-backend/internal/modules/auth/infrastructure/cache/memory"
	authsecurity "hexagonal-go-backend/internal/modules/auth/infrastructure/security"
	usersapp "hexagonal-go-backend/internal/modules/users/application"
	usershttp "hexagonal-go-backend/internal/modules/users/delivery/http/v1"
	userrepository "hexagonal-go-backend/internal/modules/users/infrastructure/persistence/memory"
	"hexagonal-go-backend/internal/platform/config"
	httpplatform "hexagonal-go-backend/internal/platform/http"
)

type mailer struct{}

func (mailer) Send(context.Context, usersapp.MailMessage) error { return nil }

func TestLoginAndProtectedUsers(t *testing.T) {
	repository := userrepository.NewUserRepository()
	tokens := authsecurity.NewHMACTokenProvider("test-secret")
	users := usersapp.NewUserService(repository, mailer{})
	createdUser, err := users.Create(context.Background(), usersapp.CreateUserInput{
		AuthID:       "admin-auth-id",
		Name:         "Admin User",
		Email:        "admin@example.com",
		IsSuperAdmin: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = users.Create(context.Background(), usersapp.CreateUserInput{
		AuthID:       "regular-user-id",
		Name:         "Student User",
		Email:        "student@example.com",
		IsSuperAdmin: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	auth := authapp.NewAuthService(repository, tokens, cache.New(), time.Minute, time.Hour, nil)
	cfg := config.Config{App: config.AppConfig{Env: "test"}, Security: config.SecurityConfig{RateLimit: 100}}
	router := httpplatform.NewRouter(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), usershttp.NewController(users), authhttp.NewController(auth), tokens, repository)

	tokenStr, err := tokens.Generate(authapp.TokenClaims{
		Subject:   createdUser.ID,
		ExpiresAt: time.Now().Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}

	// 1. Test GET /api/v1/users (Admin only, omits superadmin)
	list := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	request.Header.Set("Authorization", "Bearer "+tokenStr)
	router.ServeHTTP(list, request)
	if list.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	if list.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing request id")
	}
	var listResp struct {
		Data []struct {
			Email        string `json:"email"`
			IsSuperAdmin bool   `json:"is_super_admin"`
		} `json:"data"`
	}
	_ = json.Unmarshal(list.Body.Bytes(), &listResp)
	if len(listResp.Data) != 1 || listResp.Data[0].Email != "student@example.com" {
		t.Fatalf("expected 1 non-superadmin user in list, got %+v", listResp.Data)
	}

	// 2. Test GET /api/v1/auth/me
	meRec := httptest.NewRecorder()
	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+tokenStr)
	router.ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("auth/me status=%d body=%s", meRec.Code, meRec.Body.String())
	}

	// 3. Test PUT /api/v1/users/me/profile
	profileBody, _ := json.Marshal(map[string]any{
		"first_name": "Fabian",
		"last_name":  "Admin",
		"phone":      "+57-3001234567",
		"bio":        "Platform Super Administrator",
	})
	putRec := httptest.NewRecorder()
	putReq := httptest.NewRequest(http.MethodPut, "/api/v1/users/me/profile", bytes.NewReader(profileBody))
	putReq.Header.Set("Authorization", "Bearer "+tokenStr)
	putReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(putRec, putReq)
	if putRec.Code != http.StatusOK {
		t.Fatalf("put profile status=%d body=%s", putRec.Code, putRec.Body.String())
	}

	// 4. Test GET /api/v1/users/me/profile
	getProfRec := httptest.NewRecorder()
	getProfReq := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/profile", nil)
	getProfReq.Header.Set("Authorization", "Bearer "+tokenStr)
	router.ServeHTTP(getProfRec, getProfReq)
	if getProfRec.Code != http.StatusOK {
		t.Fatalf("get profile status=%d body=%s", getProfRec.Code, getProfRec.Body.String())
	}
}
