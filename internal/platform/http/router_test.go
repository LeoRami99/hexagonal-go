package http_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authapp "hexagonal-go-backend/internal/modules/auth/application"
	authhttp "hexagonal-go-backend/internal/modules/auth/delivery/http/v1"
	cache "hexagonal-go-backend/internal/modules/auth/infrastructure/cache/memory"
	authsecurity "hexagonal-go-backend/internal/modules/auth/infrastructure/security"
	usersapp "hexagonal-go-backend/internal/modules/users/application"
	usershttp "hexagonal-go-backend/internal/modules/users/delivery/http/v1"
	userdomain "hexagonal-go-backend/internal/modules/users/domain"
	userrepository "hexagonal-go-backend/internal/modules/users/infrastructure/persistence/memory"
	usersecurity "hexagonal-go-backend/internal/modules/users/infrastructure/security"
	"hexagonal-go-backend/internal/platform/config"
	httpplatform "hexagonal-go-backend/internal/platform/http"
)

type mailer struct{}

func (mailer) Send(context.Context, usersapp.MailMessage) error { return nil }

func TestLoginAndProtectedUsers(t *testing.T) {
	repository := userrepository.NewUserRepository()
	hasher := usersecurity.NewBcryptHasher(4)
	tokens := authsecurity.NewHMACTokenProvider("test-secret")
	users := usersapp.NewUserService(repository, hasher, mailer{})
	_, err := users.Create(context.Background(), usersapp.CreateUserInput{Name: "Admin User", Email: "admin@example.com", Password: "password123", Role: userdomain.RoleAdmin})
	if err != nil {
		t.Fatal(err)
	}
	auth := authapp.NewAuthService(repository, hasher, tokens, cache.New(), time.Minute, time.Hour)
	cfg := config.Config{App: config.AppConfig{Env: "test"}, Security: config.SecurityConfig{RateLimit: 100}}
	router := httpplatform.NewRouter(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), usershttp.NewController(users), authhttp.NewController(auth), tokens)

	login := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"admin@example.com","password":"password123"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(login, request)
	if login.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", login.Code, login.Body.String())
	}
	var body struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err = json.Unmarshal(login.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}

	list := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	request.Header.Set("Authorization", "Bearer "+body.Data.AccessToken)
	router.ServeHTTP(list, request)
	if list.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	if list.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing request id")
	}
}
