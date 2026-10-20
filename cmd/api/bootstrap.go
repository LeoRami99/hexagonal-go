package main

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"hexagonal-go-backend/internal/ent"
	authapp "hexagonal-go-backend/internal/modules/auth/application"
	authhttp "hexagonal-go-backend/internal/modules/auth/delivery/http/v1"
	authinfrastructure "hexagonal-go-backend/internal/modules/auth/infrastructure"
	memorycache "hexagonal-go-backend/internal/modules/auth/infrastructure/cache/memory"
	authent "hexagonal-go-backend/internal/modules/auth/infrastructure/persistence/ent"
	authsecurity "hexagonal-go-backend/internal/modules/auth/infrastructure/security"
	coursesapp "hexagonal-go-backend/internal/modules/courses/application"
	courseshttp "hexagonal-go-backend/internal/modules/courses/delivery/http/v1"
	courseent "hexagonal-go-backend/internal/modules/courses/infrastructure/persistence/ent"
	usersapp "hexagonal-go-backend/internal/modules/users/application"
	usershttp "hexagonal-go-backend/internal/modules/users/delivery/http/v1"
	userdomain "hexagonal-go-backend/internal/modules/users/domain"
	usermail "hexagonal-go-backend/internal/modules/users/infrastructure/mail"
	entusers "hexagonal-go-backend/internal/modules/users/infrastructure/persistence/ent"

	"hexagonal-go-backend/internal/platform/config"
	httpplatform "hexagonal-go-backend/internal/platform/http"

	"github.com/gin-gonic/gin"
)

func buildRouter(cfg config.Config, log *slog.Logger, entClient *ent.Client) (*gin.Engine, error) {
	userRepository := entusers.NewUserRepository(entClient)

	mailer := usermail.NewLogMailer(log)
	var tokenProvider authapp.TokenProvider = authsecurity.NewHMACTokenProvider(cfg.JWT.Secret)
	var idp authapp.IdentityProvider
	if cfg.Zitadel.Enabled {
		tokenProvider = authsecurity.NewZitadelTokenProvider(cfg.Zitadel.JWKSURL)
		idp = authinfrastructure.NewZitadelClient(cfg.Zitadel.URL, cfg.Zitadel.ProjectID, cfg.Zitadel.ClientID, cfg.Zitadel.RedirectURI, cfg.Zitadel.PAT)
	}
	cache := memorycache.New()

	userService := usersapp.NewUserService(userRepository, mailer)
	if err := seedAdmin(context.Background(), userService, cfg); err != nil {
		return nil, err
	}
	authService := authapp.NewAuthService(userRepository, tokenProvider, cache, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL, authent.NewAuthContextRepository(entClient), idp)

	return httpplatform.NewRouterWithCourses(
		cfg,
		log,
		usershttp.NewController(userService),
		authhttp.NewController(authService),
		courseshttp.NewController(coursesapp.NewCoursesService(courseent.NewCourseRepository(entClient))),
		tokenProvider,
		userRepository,
	), nil
}

func seedAdmin(ctx context.Context, service usersapp.UserService, cfg config.Config) error {
	_, err := service.Create(ctx, usersapp.CreateUserInput{
		AuthID:       "seed:admin:" + strings.ToLower(strings.TrimSpace(cfg.Security.AdminEmail)),
		Name:         "Administrator",
		Email:        cfg.Security.AdminEmail,
		IsSuperAdmin: true,
	})
	if errors.Is(err, userdomain.ErrEmailAlreadyExists) {
		return nil
	}
	return err
}
