package http_test

import (
	"bytes"
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
