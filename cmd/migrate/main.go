// Command migrate administra las migraciones versionadas de PostgreSQL sin
// depender de Atlas CLI: solo requiere Go, por lo que funciona igual en
// Linux, macOS y Windows.
//
// Uso:
//
//	go run ./cmd/migrate diff <nombre>       genera una migración desde los schemas Ent
//	go run ./cmd/migrate apply               aplica las migraciones pendientes
//	go run ./cmd/migrate baseline <versión>  marca una base existente en esa versión
//	go run ./cmd/migrate status              muestra migraciones aplicadas y pendientes
//	go run ./cmd/migrate hash                recalcula migrations/atlas.sum
//
// Lee DATABASE_URL del entorno o de .env. Para diff usa DEV_DATABASE_URL si
// existe; si no, crea una base temporal en el mismo servidor y la elimina al
// terminar.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"regexp"
)

const (
	migrationsDir   = "migrations"
	operatorVersion = "campus-migrate"
)

var (
	snakeCase = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	numeric   = regexp.MustCompile(`^[0-9]+$`)
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		usage()
		return errors.New("indica un comando")
	}
	if err := loadDotEnv(".env"); err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cmd, rest := args[0], args[1:]
	switch cmd {
	case "diff":
		if len(rest) != 1 || !snakeCase.MatchString(rest[0]) {
			return errors.New("indica el nombre en snake_case. Uso: migrate diff add_products")
		}
		return diff(ctx, rest[0])
	case "apply":
		return apply(ctx, "")
	case "baseline":
		if len(rest) != 1 || !numeric.MatchString(rest[0]) {
			return errors.New("indica el prefijo numérico de la migración. Uso: migrate baseline 20260925003527")
		}
		return apply(ctx, rest[0])
	case "status":
		return status(ctx)
	case "hash":
		return hash()
	case "help", "-h", "--help":
		usage()
		return nil
	default:
		usage()
		return fmt.Errorf("comando desconocido %q", cmd)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `Uso: go run ./cmd/migrate <comando>

Comandos:
  diff <nombre>       genera una migración comparando los schemas Ent con migrations/
  apply               aplica las migraciones pendientes sobre DATABASE_URL
  baseline <versión>  registra una base ya existente en la versión indicada (uso único)
  status              muestra el estado de las migraciones en DATABASE_URL
  hash                recalcula migrations/atlas.sum tras editar un archivo a mano
`)
}

func databaseURL() (string, error) {
	u := os.Getenv("DATABASE_URL")
	if u == "" {
		return "", errors.New("DATABASE_URL no está definido (en el entorno o en .env)")
	}
	return u, nil
}
