package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App      AppConfig
	JWT      JWTConfig
	Zitadel  ZitadelConfig
	Security SecurityConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

type ZitadelConfig struct {
	Enabled     bool
	URL         string
	ProjectID   string
	ClientID    string
	RedirectURI string
	PAT         string
	JWKSURL     string
}

type AppConfig struct {
	Name, Env, Port, LogLevel                               string
	ReadTimeout, WriteTimeout, IdleTimeout, ShutdownTimeout time.Duration
	AllowedOrigins                                          []string
}
type JWTConfig struct {
	Secret                string
	AccessTTL, RefreshTTL time.Duration
}
type SecurityConfig struct {
	RateLimit                 int
	AdminEmail, AdminPassword string
}
type DatabaseConfig struct {
	URL                        string
	MaxOpenConns, MaxIdleConns int
	ConnMaxLifetime            time.Duration
}
type RedisConfig struct{ Address string }

func Load() (Config, error) {
	c := Config{
		App: AppConfig{Name: get("APP_NAME", "hexagonal-go-api"), Env: get("APP_ENV", "development"), Port: get("APP_PORT", "8080"), LogLevel: get("LOG_LEVEL", "info"), AllowedOrigins: strings.Split(get("CORS_ALLOWED_ORIGINS", "http://localhost:3000"), ",")},
		JWT: JWTConfig{Secret: get("JWT_SECRET", "change-me-in-production")},
		Zitadel: ZitadelConfig{
			Enabled:     getBool("ZITADEL_ENABLED", false),
			URL:         get("ZITADEL_URL", "http://localhost:8081"),
			ProjectID:   get("ZITADEL_PROJECT_ID", ""),
			ClientID:    get("ZITADEL_CLIENT_ID", ""),
			RedirectURI: get("ZITADEL_REDIRECT_URI", "http://localhost:8085/auth/callback"),
			PAT:         get("ZITADEL_PAT", ""),
			JWKSURL:     get("ZITADEL_JWKS_URL", "http://localhost:8081/oauth/v2/keys"),
		},
		Security: SecurityConfig{RateLimit: getInt("RATE_LIMIT", 100), AdminEmail: get("ADMIN_EMAIL", "admin@example.com"), AdminPassword: get("ADMIN_PASSWORD", "admin1234")},
		Database: DatabaseConfig{URL: get("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/app?sslmode=disable"), MaxOpenConns: getInt("DATABASE_MAX_OPEN_CONNS", 25), MaxIdleConns: getInt("DATABASE_MAX_IDLE_CONNS", 5)},
		Redis:    RedisConfig{Address: os.Getenv("REDIS_ADDRESS")},
	}
	var err error
	if c.App.ReadTimeout, err = duration("HTTP_READ_TIMEOUT", "15s"); err != nil {
		return c, err
	}
	if c.App.WriteTimeout, err = duration("HTTP_WRITE_TIMEOUT", "15s"); err != nil {
		return c, err
	}
	if c.Database.ConnMaxLifetime, err = duration("DATABASE_CONN_MAX_LIFETIME", "30m"); err != nil {
		return c, err
	}
	c.App.IdleTimeout, _ = duration("HTTP_IDLE_TIMEOUT", "60s")
	c.App.ShutdownTimeout, _ = duration("SHUTDOWN_TIMEOUT", "10s")
	c.JWT.AccessTTL, _ = duration("JWT_ACCESS_TTL", "15m")
	c.JWT.RefreshTTL, _ = duration("JWT_REFRESH_TTL", "168h")
	if c.App.Env == "production" && (c.JWT.Secret == "change-me-in-production" || len(c.JWT.Secret) < 32) {
		return c, fmt.Errorf("JWT_SECRET must contain at least 32 characters in production")
	}
	return c, nil
}
func get(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func getInt(k string, d int) int {
	v, e := strconv.Atoi(get(k, strconv.Itoa(d)))
	if e != nil {
		return d
	}
	return v
}
func duration(k, d string) (time.Duration, error) {
	v := get(k, d)
	x, e := time.ParseDuration(v)
	if e != nil {
		return 0, fmt.Errorf("invalid %s: %w", k, e)
	}
	return x, nil
}
func getBool(k string, d bool) bool {
	v := get(k, strconv.FormatBool(d))
	b, e := strconv.ParseBool(v)
	if e != nil {
		return d
	}
	return b
}
