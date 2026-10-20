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
	courseshttp "hexagonal-go-backend/internal/modules/courses/delivery/http/v1"
	coursedomain "hexagonal-go-backend/internal/modules/courses/domain"
	usersapp "hexagonal-go-backend/internal/modules/users/application"
	usershttp "hexagonal-go-backend/internal/modules/users/delivery/http/v1"
	userdomain "hexagonal-go-backend/internal/modules/users/domain"
	userrepository "hexagonal-go-backend/internal/modules/users/infrastructure/persistence/memory"
	usersecurity "hexagonal-go-backend/internal/modules/users/infrastructure/security"
	"hexagonal-go-backend/internal/platform/config"
	httpplatform "hexagonal-go-backend/internal/platform/http"
)

type catalogStub struct{ courseshttp.Service }

func (catalogStub) CreateCourse(_ context.Context, item coursedomain.Course) (coursedomain.Course, error) {
	item.Template.ID = "template-id"
	item.Version.ID = "version-id"
	return item, nil
}
func (catalogStub) GetCourse(_ context.Context, id string) (coursedomain.Course, error) {
	return coursedomain.Course{Template: coursedomain.Template{ID: "template-id"}, Version: coursedomain.Version{ID: id}}, nil
}

func (catalogStub) ListTemplates(context.Context) ([]coursedomain.Template, error) {
	return []coursedomain.Template{}, nil
}

func (catalogStub) CreateTemplate(_ context.Context, item coursedomain.Template) (coursedomain.Template, error) {
	item.ID = "example-id"
	return item, nil
}

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
	router := httpplatform.NewRouter(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), usershttp.NewController(users), authhttp.NewController(auth), tokens, courseshttp.NewController(catalogStub{}))

	routes := make(map[string]bool)
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, route := range []string{
		"POST /api/v1/auth/login", "POST /api/v1/auth/refresh", "POST /api/v1/auth/logout",
		"GET /api/v1/users/:id", "POST /api/v1/users", "GET /api/v1/users",
		"PUT /api/v1/users/:id", "DELETE /api/v1/users/:id",
		"POST /api/v1/courses", "GET /api/v1/courses/:id",
		"GET /api/v1/courses/categories", "GET /api/v1/courses/templates",
		"GET /api/v1/courses/templates/:id", "GET /api/v1/courses/templates/:id/versions",
		"GET /api/v1/courses/versions/:id/syllabus", "POST /api/v1/courses/categories",
		"POST /api/v1/courses/templates", "POST /api/v1/courses/templates/:id/versions",
		"PUT /api/v1/courses/versions/:id/syllabus",
	} {
		if !routes[route] {
			t.Errorf("missing route %s", route)
		}
	}

	for _, tc := range []struct {
		method, path, body string
		want               int
	}{
		{http.MethodGet, "/api/v1/courses/templates", "", http.StatusOK},
		{http.MethodGet, "/api/v1/courses/version-id", "", http.StatusOK},
		{http.MethodPost, "/api/v1/courses", `{"template":{"code":"C1","title":"Course"},"version":{"tag":"v1","status":"DRAFT"},"modules":[{"title":"Module","sequence_order":1,"lessons":[{"title":"Lesson","sequence_order":1,"activities":[{"title":"Activity","type":"VIDEO","sequence_order":1,"resources":[{"type":"VIDEO","name":"Intro","url_storage_key":"courses/intro.mp4","position":1}]}]}]}]}`, http.StatusCreated},
		{http.MethodPost, "/api/v1/courses/templates", `{"code":"C1","title":"Course"}`, http.StatusCreated},
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(recorder, request)
		if recorder.Code != tc.want {
			t.Errorf("%s %s without token: status=%d body=%s", tc.method, tc.path, recorder.Code, recorder.Body.String())
		}
	}

	unauthorized := httptest.NewRecorder()
	router.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/users/example-id", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Errorf("user route without token: status=%d", unauthorized.Code)
	}

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
