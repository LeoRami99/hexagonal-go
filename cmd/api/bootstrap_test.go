package main

import (
	"context"
	"testing"

	usersapp "hexagonal-go-backend/internal/modules/users/application"
	userrepository "hexagonal-go-backend/internal/modules/users/infrastructure/persistence/memory"
	"hexagonal-go-backend/internal/platform/config"
)

func TestSeedAdminIsIdempotentAndLinksOnLogin(t *testing.T) {
	ctx := context.Background()
	repo := userrepository.NewUserRepository()
	service := usersapp.NewUserService(repo, nil)
	cfg := config.Config{Security: config.SecurityConfig{AdminEmail: " Admin@Example.com "}}

	for i := 0; i < 2; i++ {
		if err := seedAdmin(ctx, service, cfg); err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
	}
	admin, err := repo.FindByEmail(ctx, "admin@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !admin.IsSuperAdmin || admin.AuthID != "seed:admin:admin@example.com" {
		t.Fatalf("unexpected seeded admin: %+v", admin)
	}

	admin.AuthID = "zitadel-user-id"
	if err := repo.Update(ctx, admin); err != nil {
		t.Fatal(err)
	}
	if err := seedAdmin(ctx, service, cfg); err != nil {
		t.Fatalf("seed after linking: %v", err)
	}
	linked, err := repo.FindByEmail(ctx, "admin@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if linked.AuthID != "zitadel-user-id" {
		t.Fatalf("admin auth ID was overwritten: %q", linked.AuthID)
	}
}
