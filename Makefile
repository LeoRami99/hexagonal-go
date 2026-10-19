
APP := bin/api
CMD := ./cmd/api
MODULES_DIR := internal/modules
MIGRATIONS_DIR := file://migrations
ATLAS ?= atlas
export GOTOOLCHAIN ?= go1.26.0

.PHONY: run dev build test cover fmt fmt-check vet lint check swagger ent-generate ent-schema atlas-check migration-diff migrate-baseline migrate-apply migrate-status docker-up docker-down module clean

run:
	@if [ -f .env ]; then \
		set -a; \
		. ./.env; \
		set +a; \
	fi; \
	go run $(CMD)

dev:
	@if [ -f .env ]; then \
		set -a; \
		. ./.env; \
		set +a; \
	fi; \
	which air >/dev/null 2>&1 && air || go run github.com/air-verse/air@v1.61.7

build:
	go build -o $(APP) $(CMD)

test:
	go test ./... -race -v

cover:
	go test ./... -race -coverprofile=coverage.out

fmt:
	gofmt -w .

fmt-check:
	test -z "$$(gofmt -l .)"

vet:
	go vet ./...

lint:
	golangci-lint run

check: fmt-check vet test build

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.3 init -g internal/platform/http/router.go -o docs --parseDependency --parseInternal

ent-generate:
	go generate ./internal/ent

# Uso: make ent-schema NAME=Course
ent-schema:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: indica el nombre del schema. Uso: make ent-schema NAME=Course"; \
		exit 1; \
	fi
	@if ! printf '%s' "$(NAME)" | grep -Eq '^[A-Z][A-Za-z0-9]*$$'; then \
		echo "Error: NAME debe usar PascalCase, comenzar con mayúscula y contener solo letras o números."; \
		exit 1; \
	fi
	@if [ -e "internal/ent/schema/$$(printf '%s' "$(NAME)" | tr '[:upper:]' '[:lower:]').go" ]; then \
		echo "Error: el schema $(NAME) ya existe."; \
		exit 1; \
	fi
	go run entgo.io/ent/cmd/ent new --target internal/ent/schema $(NAME)
	$(MAKE) ent-generate

atlas-check:
	@command -v $(ATLAS) >/dev/null 2>&1 || { \
		echo "Error: Atlas CLI no está instalado. macOS: brew install ariga/tap/atlas"; \
		exit 1; \
	}

# Uso: make migration-diff NAME=add_courses
migration-diff: atlas-check
	@if [ -z "$(NAME)" ]; then \
		echo "Error: indica el nombre de la migración. Uso: make migration-diff NAME=add_courses"; \
		exit 1; \
	fi
	@if ! printf '%s' "$(NAME)" | grep -Eq '^[a-z][a-z0-9_]*$$'; then \
		echo "Error: NAME debe usar snake_case."; \
		exit 1; \
	fi
	$(ATLAS) migrate diff $(NAME) \
		--dir "$(MIGRATIONS_DIR)" \
		--to "ent://internal/ent/schema" \
		--dev-url "docker://postgres/17/dev?search_path=public"

# Uso único al adoptar Atlas sobre una base ya creada por Ent:
# make migrate-baseline VERSION=20260925003527
migrate-baseline: atlas-check
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: indica la versión inicial. Uso: make migrate-baseline VERSION=20260925003527"; \
		exit 1; \
	fi
	@if ! printf '%s' "$(VERSION)" | grep -Eq '^[0-9]+$$'; then \
		echo "Error: VERSION debe contener únicamente el prefijo numérico de la migración."; \
		exit 1; \
	fi
	@if [ ! -f .env ]; then \
		echo "Error: no existe .env con DATABASE_URL."; \
		exit 1; \
	fi; \
	set -a; \
	. ./.env; \
	set +a; \
	$(ATLAS) migrate apply --dir "$(MIGRATIONS_DIR)" --url "$$DATABASE_URL" --baseline "$(VERSION)"

migrate-apply: atlas-check
	@if [ ! -f .env ]; then \
		echo "Error: no existe .env con DATABASE_URL."; \
		exit 1; \
	fi; \
	set -a; \
	. ./.env; \
	set +a; \
	$(ATLAS) migrate apply --dir "$(MIGRATIONS_DIR)" --url "$$DATABASE_URL"

migrate-status: atlas-check
	@if [ ! -f .env ]; then \
		echo "Error: no existe .env con DATABASE_URL."; \
		exit 1; \
	fi; \
	set -a; \
	. ./.env; \
	set +a; \
	GOTOOLCHAIN=go1.26.0 $(ATLAS) migrate status --dir "$(MIGRATIONS_DIR)" --url "$$DATABASE_URL"

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

# Uso: make module MODULE=courses
module:
	@if [ -z "$(MODULE)" ]; then \
		echo "Error: indica el nombre del módulo. Uso: make module MODULE=courses"; \
		exit 1; \
	fi
	@if ! printf '%s' "$(MODULE)" | grep -Eq '^[a-z][a-z0-9_]*$$'; then \
		echo "Error: MODULE debe usar snake_case, comenzar con una letra y contener solo minúsculas, números o _."; \
		exit 1; \
	fi
	@if [ -e "$(MODULES_DIR)/$(MODULE)" ]; then \
		echo "Error: el módulo $(MODULE) ya existe en $(MODULES_DIR)/$(MODULE)."; \
		exit 1; \
	fi
	@set -eu; \
		base="$(MODULES_DIR)/$(MODULE)"; \
		mkdir -p \
			"$$base/domain" \
			"$$base/application" \
			"$$base/infrastructure/persistence/memory" \
			"$$base/infrastructure/persistence/ent" \
			"$$base/infrastructure/security" \
			"$$base/infrastructure/mail" \
			"$$base/delivery/http/v1"; \
		printf 'package domain\n' > "$$base/domain/module.go"; \
		printf 'package domain\n' > "$$base/domain/errors.go"; \
		printf 'package application\n' > "$$base/application/ports.go"; \
		printf 'package application\n' > "$$base/application/service.go"; \
		printf 'package application\n' > "$$base/application/service_test.go"; \
		printf 'package v1\n' > "$$base/delivery/http/v1/controller.go"; \
		printf 'package v1\n' > "$$base/delivery/http/v1/dto.go"; \
		touch \
			"$$base/infrastructure/persistence/memory/.gitkeep" \
			"$$base/infrastructure/persistence/ent/.gitkeep" \
			"$$base/infrastructure/security/.gitkeep" \
			"$$base/infrastructure/mail/.gitkeep"; \
		echo "Módulo $(MODULE) creado en $$base"

clean:
	rm -rf bin coverage.out tmp
