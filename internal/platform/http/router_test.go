package http_test

import (
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
	coursesapp "hexagonal-go-backend/internal/modules/courses/application"
	courseshttp "hexagonal-go-backend/internal/modules/courses/delivery/http/v1"

	usersapp "hexagonal-go-backend/internal/modules/users/application"
	usershttp "hexagonal-go-backend/internal/modules/users/delivery/http/v1"
	userrepository "hexagonal-go-backend/internal/modules/users/infrastructure/persistence/memory"
	"hexagonal-go-backend/internal/platform/config"
	httpplatform "hexagonal-go-backend/internal/platform/http"
)

type mailer struct{}
type courseRepositoryStub struct{ coursesapp.CoursesRepository }

func (mailer) Send(context.Context, usersapp.MailMessage) error { return nil }

func TestModuleRoutesAndAuthorization(t *testing.T) {
	repository := userrepository.NewUserRepository()
	users := usersapp.NewUserService(repository, mailer{})
	admin, err := users.Create(context.Background(), usersapp.CreateUserInput{
		Name: "Admin User", Email: "admin@example.com", IsSuperAdmin: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = users.Create(context.Background(), usersapp.CreateUserInput{
		Name: "Student User", Email: "student@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	tokens := authsecurity.NewHMACTokenProvider("test-secret")
	auth := authapp.NewAuthService(repository, tokens, cache.New(), time.Minute, time.Hour, nil)
	cfg := config.Config{App: config.AppConfig{Env: "test"}, Security: config.SecurityConfig{RateLimit: 100}}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	courses := courseshttp.NewController(coursesapp.NewCoursesService(courseRepositoryStub{}))
	router := httpplatform.NewRouterWithCourses(cfg, log, usershttp.NewController(users), authhttp.NewController(auth), courses, tokens, repository)

	routes := make(map[string]bool)
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, route := range []string{
		"POST /api/v1/auth/login", "POST /api/v1/auth/refresh", "POST /api/v1/auth/logout",
		"GET /api/v1/users/:id", "GET /api/v1/users/me/profile", "PUT /api/v1/users/me/profile",
		"POST /api/v1/users", "GET /api/v1/users", "PUT /api/v1/users/:id", "DELETE /api/v1/users/:id",
		"POST /api/v1/courses", "GET /api/v1/courses", "GET /api/v1/courses/:id",
		"PUT /api/v1/courses/:id", "DELETE /api/v1/courses/:id",
	} {
		if !routes[route] {
			t.Errorf("missing route %s", route)
		}
	}

	unauthorized := httptest.NewRecorder()
	router.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/users", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Errorf("user route without token: status=%d", unauthorized.Code)
	}

	token, err := tokens.Generate(authapp.TokenClaims{Subject: admin.ID, ExpiresAt: time.Now().Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	list := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(list, request)
	if list.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	if list.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing request id")
	}
	var result struct {
		Data []struct {
			Email string `json:"email"`
		} `json:"data"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Data) != 1 || result.Data[0].Email != "student@example.com" {
		t.Fatalf("unexpected list: %+v", result.Data)
	}
}
