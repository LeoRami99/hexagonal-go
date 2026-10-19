package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/schema"
)

// Mismo esquema y tabla que usa Atlas CLI, para que ambas herramientas
// puedan operar sobre la misma base sin perder el historial.
const (
	revisionsSchema = "atlas_schema_revisions"
	revisionsTable  = "atlas_schema_revisions"
)

// revisionStore implementa migrate.RevisionReadWriter sobre PostgreSQL.
type revisionStore struct {
	conn schema.ExecQuerier
}

var _ migrate.RevisionReadWriter = (*revisionStore)(nil)

func (s *revisionStore) Ident() *migrate.TableIdent {
	return &migrate.TableIdent{Schema: revisionsSchema, Name: revisionsTable}
}

func (s *revisionStore) init(ctx context.Context) error {
	stmts := []string{
		fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %q`, revisionsSchema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %q.%q (
  "version" character varying NOT NULL,
  "description" character varying NOT NULL,
  "type" bigint NOT NULL DEFAULT 2,
  "applied" bigint NOT NULL DEFAULT 0,
  "total" bigint NOT NULL DEFAULT 0,
  "executed_at" timestamptz NOT NULL,
  "execution_time" bigint NOT NULL,
  "error" text NULL,
  "error_stmt" text NULL,
  "hash" character varying NOT NULL,
  "partial_hashes" jsonb NULL,
  "operator_version" character varying NOT NULL,
  PRIMARY KEY ("version")
)`, revisionsSchema, revisionsTable),
	}
	for _, stmt := range stmts {
		if _, err := s.conn.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("crear tabla de revisiones: %w", err)
		}
	}
	return nil
}

func (s *revisionStore) exists(ctx context.Context) (bool, error) {
	rows, err := s.conn.QueryContext(ctx,
		`SELECT 1 FROM information_schema.tables WHERE table_schema = $1 AND table_name = $2`,
		revisionsSchema, revisionsTable,
	)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), rows.Err()
}

const selectRevision = `SELECT "version", "description", "type", "applied", "total", "executed_at",
  "execution_time", COALESCE("error", ''), COALESCE("error_stmt", ''), "hash", "partial_hashes", "operator_version"
FROM "atlas_schema_revisions"."atlas_schema_revisions"`

func (s *revisionStore) ReadRevisions(ctx context.Context) ([]*migrate.Revision, error) {
	ok, err := s.exists(ctx)
	if err != nil || !ok {
		return nil, err
	}
	rows, err := s.conn.QueryContext(ctx, selectRevision+` ORDER BY "version"`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var revs []*migrate.Revision
	for rows.Next() {
		r, err := scanRevision(rows)
		if err != nil {
			return nil, err
		}
		revs = append(revs, r)
	}
	return revs, rows.Err()
}

func (s *revisionStore) ReadRevision(ctx context.Context, version string) (*migrate.Revision, error) {
	ok, err := s.exists(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, migrate.ErrRevisionNotExist
	}
	rows, err := s.conn.QueryContext(ctx, selectRevision+` WHERE "version" = $1`, version)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, migrate.ErrRevisionNotExist
	}
	return scanRevision(rows)
}

func (s *revisionStore) WriteRevision(ctx context.Context, r *migrate.Revision) error {
	if err := s.init(ctx); err != nil {
		return err
	}
	var partial any
	if len(r.PartialHashes) > 0 {
		b, err := json.Marshal(r.PartialHashes)
		if err != nil {
			return err
		}
		partial = string(b)
	}
	_, err := s.conn.ExecContext(ctx, `INSERT INTO "atlas_schema_revisions"."atlas_schema_revisions"
  ("version", "description", "type", "applied", "total", "executed_at", "execution_time",
   "error", "error_stmt", "hash", "partial_hashes", "operator_version")
VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), NULLIF($9, ''), $10, $11, $12)
ON CONFLICT ("version") DO UPDATE SET
  "description" = EXCLUDED."description", "type" = EXCLUDED."type", "applied" = EXCLUDED."applied",
  "total" = EXCLUDED."total", "executed_at" = EXCLUDED."executed_at", "execution_time" = EXCLUDED."execution_time",
  "error" = EXCLUDED."error", "error_stmt" = EXCLUDED."error_stmt", "hash" = EXCLUDED."hash",
  "partial_hashes" = EXCLUDED."partial_hashes", "operator_version" = EXCLUDED."operator_version"`,
		r.Version, r.Description, int64(r.Type), r.Applied, r.Total, r.ExecutedAt, int64(r.ExecutionTime),
		r.Error, r.ErrorStmt, r.Hash, partial, r.OperatorVersion,
	)
	return err
}

func (s *revisionStore) DeleteRevision(ctx context.Context, version string) error {
	_, err := s.conn.ExecContext(ctx, `DELETE FROM "atlas_schema_revisions"."atlas_schema_revisions" WHERE "version" = $1`, version)
	return err
}

func scanRevision(rows *sql.Rows) (*migrate.Revision, error) {
	var (
		r        migrate.Revision
		typ      int64
		execTime int64
		partial  sql.NullString
	)
	if err := rows.Scan(&r.Version, &r.Description, &typ, &r.Applied, &r.Total, &r.ExecutedAt,
		&execTime, &r.Error, &r.ErrorStmt, &r.Hash, &partial, &r.OperatorVersion); err != nil {
		return nil, err
	}
	r.Type = migrate.RevisionType(typ)
	r.ExecutionTime = time.Duration(execTime)
	if partial.Valid && partial.String != "" {
		if err := json.Unmarshal([]byte(partial.String), &r.PartialHashes); err != nil {
			return nil, errors.Join(errors.New("leer partial_hashes"), err)
		}
	}
	return &r, nil
}
