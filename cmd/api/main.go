package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	authapp "hexagonal-go-backend/internal/modules/auth/application"
	authhttp "hexagonal-go-backend/internal/modules/auth/delivery/http/v1"
	authinfra "hexagonal-go-backend/internal/modules/auth/infrastructure"
	memorycache "hexagonal-go-backend/internal/modules/auth/infrastructure/cache/memory"
	authpersistence "hexagonal-go-backend/internal/modules/auth/infrastructure/persistence/ent"
	authsecurity "hexagonal-go-backend/internal/modules/auth/infrastructure/security"
	courseapp "hexagonal-go-backend/internal/modules/courses/application"
	courseshttp "hexagonal-go-backend/internal/modules/courses/delivery/http/v1"
	coursepersistence "hexagonal-go-backend/internal/modules/courses/infrastructure/persistence/ent"
	usersapp "hexagonal-go-backend/internal/modules/users/application"
	usershttp "hexagonal-go-backend/internal/modules/users/delivery/http/v1"
	userdomain "hexagonal-go-backend/internal/modules/users/domain"
	usermail "hexagonal-go-backend/internal/modules/users/infrastructure/mail"
	entusers "hexagonal-go-backend/internal/modules/users/infrastructure/persistence/ent"

	"hexagonal-go-backend/internal/platform/config"
	"hexagonal-go-backend/internal/platform/database"
	"hexagonal-go-backend/internal/platform/database/seeder"
	httpplatform "hexagonal-go-backend/internal/platform/http"
	"hexagonal-go-backend/internal/platform/logger"
	"hexagonal-go-backend/internal/platform/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.App.LogLevel, cfg.App.Env)
	autoMigrate := cfg.App.Env == "development"
	entClient, _, err := database.OpenPostgres(context.Background(), cfg.Database.URL, cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns, cfg.Database.ConnMaxLifetime, autoMigrate)
	if err != nil {
		log.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := entClient.Close(); err != nil {
			log.Error("failed to close postgres", "error", err)
		}
	}()

	userRepository := entusers.NewUserRepository(entClient)
	coursesRepository := coursepersistence.NewCourseRepository(entClient)
	mailer := usermail.NewLogMailer(log)

	var tokenProvider authapp.TokenProvider
	var idp authapp.IdentityProvider
	var zitadelClient *authinfra.ZitadelClient
	if cfg.Zitadel.Enabled {
		log.Info("using ZITADEL JWKS token provider", "jwks_url", cfg.Zitadel.JWKSURL)
		tokenProvider = authsecurity.NewZitadelTokenProvider(cfg.Zitadel.JWKSURL)
		zitadelClient = authinfra.NewZitadelClient(cfg.Zitadel.URL, cfg.Zitadel.ProjectID, cfg.Zitadel.ClientID, cfg.Zitadel.RedirectURI, cfg.Zitadel.PAT)
		idp = zitadelClient
	} else {
		log.Info("using local HMAC token provider")
		tokenProvider = authsecurity.NewHMACTokenProvider(cfg.JWT.Secret)
	}
	cache := memorycache.New()

	if err := seeder.SeedPlatform(context.Background(), entClient, log); err != nil {
		log.Error("failed to seed platform catalog", "error", err)
		os.Exit(1)
	}

	userService := usersapp.NewUserService(userRepository, mailer)

	courseService := courseapp.NewCoursesService(coursesRepository)

	if err := seedAdmin(context.Background(), userService, userRepository, zitadelClient, log, cfg); err != nil {
		log.Error("failed to create development administrator", "error", err)
		os.Exit(1)
	}
	authContextRepo := authpersistence.NewAuthContextRepository(entClient)
	authService := authapp.NewAuthService(userRepository, tokenProvider, cache, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL, authContextRepo, idp)

	router := httpplatform.NewRouterWithCourses(
		cfg,
		log,
		usershttp.NewController(userService),
		authhttp.NewController(authService),
		courseshttp.NewController(courseService),
		tokenProvider,
		userRepository,
	)
	httpServer := server.New(router, cfg.App)

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		log.Info("http server started", "address", httpServer.Addr, "environment", cfg.App.Env)
		serverErrors <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http server failed", "error", err)
			os.Exit(1)
		}
	case <-shutdownSignal.Done():
		log.Info("shutdown signal received")
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(shutdownContext); err != nil {
		log.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	log.Info("http server stopped")
}

func seedAdmin(ctx context.Context, service usersapp.UserService, repo usersapp.UserRepository, zitadel *authinfra.ZitadelClient, log *slog.Logger, cfg config.Config) error {
	authID := "admin-seed-id"

	if zitadel != nil && cfg.Zitadel.Enabled && cfg.Zitadel.PAT != "" {
		log.Info("syncing administrator with ZITADEL...", "email", cfg.Security.AdminEmail)

		existingUID, errFind := zitadel.FindUser(ctx, cfg.Security.AdminEmail)
		if errFind == nil && existingUID != "" {
			authID = existingUID
			if err := zitadel.SetUserPassword(ctx, existingUID, cfg.Security.AdminPassword); err != nil {
				log.Warn("could not update admin password in ZITADEL", "error", err)
			} else {
				log.Info("administrator password verified in ZITADEL", "auth_id", authID)
			}
		} else {
			createdUID, errCreate := zitadel.CreateUser(ctx, cfg.Security.AdminEmail, "Administrator", cfg.Security.AdminPassword)
			if errCreate != nil {
				log.Warn("could not create admin user in ZITADEL", "error", errCreate)
			} else if createdUID != "" {
				authID = createdUID
				log.Info("administrator created in ZITADEL", "auth_id", authID)
			}
		}
	}

	_, err := service.Create(ctx, usersapp.CreateUserInput{
		AuthID:       authID,
		Name:         "Administrator",
		Email:        cfg.Security.AdminEmail,
		IsSuperAdmin: true,
	})
	if errors.Is(err, userdomain.ErrEmailAlreadyExists) {
		if repo != nil {
			if existing, errFind := repo.FindByEmail(ctx, cfg.Security.AdminEmail); errFind == nil {
				updated := false
				if authID != "admin-seed-id" && existing.AuthID != authID {
					existing.AuthID = authID
					updated = true
				}
				if !existing.IsSuperAdmin {
					existing.IsSuperAdmin = true
					updated = true
				}
				if updated {
					_ = repo.Update(ctx, existing)
					log.Info("updated administrator credentials and superadmin flag in database", "auth_id", existing.AuthID)
				}
			}
		}
		return nil
	}
	return err
}
