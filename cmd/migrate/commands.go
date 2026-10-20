package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"
	"text/tabwriter"
	"time"

	"hexagonal-go-backend/internal/ent/migrate"

	atlasmigrate "ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/postgres"
	atlasschema "ariga.io/atlas/sql/schema"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	entschema "entgo.io/ent/dialect/sql/schema"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const lockTimeout = 10 * time.Second

func openDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("abrir postgres: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("conectar a postgres: %w", err)
	}
	return db, nil
}

func openDir() (*atlasmigrate.LocalDir, error) {
	dir, err := atlasmigrate.NewLocalDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("abrir %s/: %w", migrationsDir, err)
	}
	return dir, nil
}

func checksumHint(err error) error {
	if errors.Is(err, atlasmigrate.ErrChecksumMismatch) || errors.Is(err, atlasmigrate.ErrChecksumNotFound) {
		return fmt.Errorf("%w: %s/atlas.sum no coincide con los archivos; si editaste una migración a mano ejecuta `make migrate-hash`", err, migrationsDir)
	}
	return err
}

// diff genera un nuevo archivo de migración con la diferencia entre el estado
// de migrations/ (reproducido en una base temporal) y los schemas Ent.
func diff(ctx context.Context, name string) error {
	dir, err := openDir()
	if err != nil {
		return err
	}
	if err := atlasmigrate.Validate(dir); err != nil {
		return checksumHint(err)
	}
	devURL, cleanup, err := devDatabase(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	db, err := openDB(ctx, devURL)
	if err != nil {
		return err
	}
	defer db.Close()

	m, err := entschema.NewMigrate(entsql.OpenDB(dialect.Postgres, db),
		entschema.WithDir(dir),
		entschema.WithMigrationMode(entschema.ModeReplay),
		entschema.WithDialect(dialect.Postgres),
		entschema.WithFormatter(atlasmigrate.DefaultFormatter),
		entschema.WithDropColumn(true),
		entschema.WithDropIndex(true),
		entschema.WithErrNoPlan(true),
	)
	if err != nil {
		return err
	}
	before, _ := dir.Files()
	if err := m.NamedDiff(ctx, name, migrate.Tables...); err != nil {
		if errors.Is(err, atlasmigrate.ErrNoPlan) {
			fmt.Println("Sin cambios: los schemas Ent ya coinciden con las migraciones.")
			return nil
		}
		return fmt.Errorf("generar migración: %w", err)
	}
	after, _ := dir.Files()
	if len(after) > len(before) {
		fmt.Printf("Migración creada: %s/%s\n", migrationsDir, after[len(after)-1].Name())
	}
	fmt.Println("Revisa el SQL generado antes de versionarlo.")
	return nil
}

// devDatabase devuelve una base vacía para reproducir las migraciones. Usa
// DEV_DATABASE_URL si está definida; si no, crea una base temporal en el
// servidor de DATABASE_URL (requiere permiso CREATEDB) y la elimina al final.
func devDatabase(ctx context.Context) (string, func(), error) {
	if u := os.Getenv("DEV_DATABASE_URL"); u != "" {
		return u, func() {}, nil
	}
	base, err := databaseURL()
	if err != nil {
		return "", nil, err
	}
	u, err := url.Parse(base)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		return "", nil, errors.New("DATABASE_URL debe tener formato postgres://...; o define DEV_DATABASE_URL")
	}
	admin, err := openDB(ctx, base)
	if err != nil {
		return "", nil, err
	}
	name := fmt.Sprintf("atlas_dev_%d", time.Now().UnixNano())
	if _, err := admin.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE %q`, name)); err != nil {
		_ = admin.Close()
		return "", nil, fmt.Errorf("crear base temporal (o define DEV_DATABASE_URL): %w", err)
	}
	cleanup := func() {
		defer admin.Close()
		if _, err := admin.ExecContext(context.Background(), fmt.Sprintf(`DROP DATABASE IF EXISTS %q WITH (FORCE)`, name)); err != nil {
			fmt.Fprintf(os.Stderr, "Aviso: no se pudo eliminar la base temporal %s: %v\n", name, err)
		}
	}
	dev := *u
	dev.Path = "/" + name
	return dev.String(), cleanup, nil
}

// apply ejecuta las migraciones pendientes, cada archivo en su propia
// transacción. Con baseline, en una base sin historial registra esa versión
// como ya aplicada y continúa con las siguientes.
func apply(ctx context.Context, baseline string) error {
	dsn, err := databaseURL()
	if err != nil {
		return err
	}
	dir, err := openDir()
	if err != nil {
		return err
	}
	db, err := openDB(ctx, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	drv, err := postgres.Open(db)
	if err != nil {
		return err
	}
	unlock, err := drv.(atlasschema.Locker).Lock(ctx, "atlas_migrate_execute", lockTimeout)
	if err != nil {
		return fmt.Errorf("obtener lock de migraciones: %w", err)
	}
	defer unlock()

	rrw := &revisionStore{conn: db}
	if baseline != "" {
		revs, err := rrw.ReadRevisions(ctx)
		if err != nil {
			return err
		}
		if len(revs) > 0 {
			return errors.New("la base ya tiene historial de migraciones; baseline solo se usa una vez")
		}
	}
	opts := []atlasmigrate.ExecutorOption{atlasmigrate.WithOperatorVersion(operatorVersion)}
	if baseline != "" {
		opts = append(opts, atlasmigrate.WithBaselineVersion(baseline))
	}
	ex, err := atlasmigrate.NewExecutor(drv, dir, rrw, opts...)
	if err != nil {
		return err
	}
	pending, err := ex.Pending(ctx)
	switch {
	case errors.Is(err, atlasmigrate.ErrNoPendingFiles):
		if baseline != "" {
			fmt.Printf("Baseline %s registrado.\n", baseline)
		}
		fmt.Println("La base está al día: no hay migraciones pendientes.")
		return nil
	case errors.As(err, new(*atlasmigrate.NotCleanError)):
		return fmt.Errorf("%w\nLa base ya tiene tablas y no tiene historial: usa `make migrate-baseline VERSION=<versión>`", err)
	case err != nil:
		return checksumHint(err)
	}
	if baseline != "" {
		fmt.Printf("Baseline %s registrado.\n", baseline)
	}

	for _, f := range pending {
		start := time.Now()
		if err := executeFile(ctx, db, dir, f); err != nil {
			return fmt.Errorf("%s: %w", f.Name(), err)
		}
		fmt.Printf("  ✓ %s (%s)\n", f.Name(), time.Since(start).Round(time.Millisecond))
	}
	fmt.Printf("%d migración(es) aplicada(s).\n", len(pending))
	return nil
}

// executeFile aplica un archivo dentro de una transacción, salvo que declare
// `-- atlas:txmode none` (por ejemplo, para CREATE INDEX CONCURRENTLY).
func executeFile(ctx context.Context, db *sql.DB, dir atlasmigrate.Dir, f atlasmigrate.File) error {
	if lf, ok := f.(*atlasmigrate.LocalFile); ok {
		if mode := lf.Directive("txmode"); len(mode) > 0 && strings.TrimSpace(mode[0]) == "none" {
			drv, err := postgres.Open(db)
			if err != nil {
				return err
			}
			return newExecutor(drv, dir, db).Execute(ctx, f)
		}
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	drv, err := postgres.Open(tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := newExecutor(drv, dir, tx).Execute(ctx, f); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func newExecutor(drv atlasmigrate.Driver, dir atlasmigrate.Dir, conn atlasschema.ExecQuerier) *atlasmigrate.Executor {
	// NewExecutor solo falla con argumentos nil.
	ex, _ := atlasmigrate.NewExecutor(drv, dir, &revisionStore{conn: conn}, atlasmigrate.WithOperatorVersion(operatorVersion))
	return ex
}

func status(ctx context.Context) error {
	dsn, err := databaseURL()
	if err != nil {
		return err
	}
	dir, err := openDir()
	if err != nil {
		return err
	}
	db, err := openDB(ctx, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	drv, err := postgres.Open(db)
	if err != nil {
		return err
	}
	rrw := &revisionStore{conn: db}
	revs, err := rrw.ReadRevisions(ctx)
	if err != nil {
		return err
	}
	ex, err := atlasmigrate.NewExecutor(drv, dir, rrw)
	if err != nil {
		return err
	}
	pending, err := ex.Pending(ctx)
	switch {
	case errors.Is(err, atlasmigrate.ErrNoPendingFiles):
		pending = nil
	case errors.As(err, new(*atlasmigrate.NotCleanError)):
		fmt.Println("La base tiene tablas pero no historial de migraciones: usa `make migrate-baseline VERSION=<versión>`.")
		return nil
	case err != nil:
		return checksumHint(err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "VERSIÓN\tDESCRIPCIÓN\tESTADO")
	for _, r := range revs {
		state := "aplicada"
		switch {
		case r.Type.Has(atlasmigrate.RevisionTypeBaseline):
			state = "baseline"
		case r.Error != "":
			state = "error: " + r.Error
		case r.Applied < r.Total:
			state = fmt.Sprintf("parcial (%d/%d)", r.Applied, r.Total)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", r.Version, r.Description, state)
	}
	for _, f := range pending {
		if slices.ContainsFunc(revs, func(r *atlasmigrate.Revision) bool { return r.Version == f.Version() }) {
			continue
		}
		fmt.Fprintf(w, "%s\t%s\tpendiente\n", f.Version(), f.Desc())
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if len(pending) == 0 {
		fmt.Println("\nLa base está al día.")
	} else {
		fmt.Printf("\n%d migración(es) pendiente(s). Ejecuta `make migrate-apply`.\n", len(pending))
	}
	return nil
}

// hash recalcula atlas.sum, necesario tras editar a mano un archivo SQL.
func hash() error {
	dir, err := openDir()
	if err != nil {
		return err
	}
	sum, err := dir.Checksum()
	if err != nil {
		return err
	}
	if err := atlasmigrate.WriteSumFile(dir, sum); err != nil {
		return err
	}
	fmt.Printf("%s/atlas.sum actualizado.\n", migrationsDir)
	return nil
}
