package main

import (
	"context"
	"errors"
	"log/slog"

	"hexagonal-go-backend/internal/ent"
	authapp "hexagonal-go-backend/internal/modules/auth/application"
	authhttp "hexagonal-go-backend/internal/modules/auth/delivery/http/v1"
	memorycache "hexagonal-go-backend/internal/modules/auth/infrastructure/cache/memory"
	authsecurity "hexagonal-go-backend/internal/modules/auth/infrastructure/security"
	coursesapp "hexagonal-go-backend/internal/modules/courses/application"
	courseshttp "hexagonal-go-backend/internal/modules/courses/delivery/http/v1"
	coursepostgres "hexagonal-go-backend/internal/modules/courses/infrastructure/persistence/postgres"
	usersapp "hexagonal-go-backend/internal/modules/users/application"
	usershttp "hexagonal-go-backend/internal/modules/users/delivery/http/v1"
	userdomain "hexagonal-go-backend/internal/modules/users/domain"
	usermail "hexagonal-go-backend/internal/modules/users/infrastructure/mail"
	entusers "hexagonal-go-backend/internal/modules/users/infrastructure/persistence/ent"
	usersecurity "hexagonal-go-backend/internal/modules/users/infrastructure/security"
	"hexagonal-go-backend/internal/platform/config"
	httpplatform "hexagonal-go-backend/internal/platform/http"

	"github.com/gin-gonic/gin"
)

func buildRouter(cfg config.Config, log *slog.Logger, entClient *ent.Client) (*gin.Engine, error) {
	userRepository := entusers.NewUserRepository(entClient)
	passwordHasher := usersecurity.NewBcryptHasher(0)
	mailer := usermail.NewLogMailer(log)
	tokenProvider := authsecurity.NewHMACTokenProvider(cfg.JWT.Secret)
	cache := memorycache.New()

	userService := usersapp.NewUserService(userRepository, passwordHasher, mailer)
	if err := seedAdmin(context.Background(), userService, cfg); err != nil {
		return nil, err
	}
	authService := authapp.NewAuthService(userRepository, passwordHasher, tokenProvider, cache, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	return httpplatform.NewRouter(
		cfg,
		log,
		usershttp.NewController(userService),
		authhttp.NewController(authService),
		tokenProvider,
		courseshttp.NewController(coursesapp.NewService(coursepostgres.NewRepository(entClient))),
	), nil
}

func seedAdmin(ctx context.Context, service usersapp.UserService, cfg config.Config) error {
	_, err := service.Create(ctx, usersapp.CreateUserInput{
		Name:     "Administrator",
		Email:    cfg.Security.AdminEmail,
		Password: cfg.Security.AdminPassword,
		Role:     userdomain.RoleAdmin,
	})
	if errors.Is(err, userdomain.ErrEmailAlreadyExists) {
		return nil
	}
	return err
}
