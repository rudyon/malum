// Package catalog stores Malum's queryable library catalogue in SQLite.
package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rudyon/malum/internal/identifier"
	_ "modernc.org/sqlite"
)

const databaseFilename = "malum.db"

var (
	ErrDocumentExists   = errors.New("catalogue document already exists")
	ErrDocumentNotFound = errors.New("catalogue document not found")
	ErrAuthorNotFound   = errors.New("catalogue author not found")
)

type Catalog struct {
	database *sql.DB
	newID    func() (string, error)
}

func Open(ctx context.Context, dataRoot string) (*Catalog, error) {
	if strings.TrimSpace(dataRoot) == "" {
		return nil, errors.New("catalogue requires a data root")
	}
	if err := os.MkdirAll(dataRoot, 0o755); err != nil {
		return nil, fmt.Errorf("create catalogue data root: %w", err)
	}

	database, err := sql.Open("sqlite", filepath.Join(dataRoot, databaseFilename))
	if err != nil {
		return nil, fmt.Errorf("open catalogue database: %w", err)
	}
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)

	closeOnError := func(err error) (*Catalog, error) {
		_ = database.Close()
		return nil, err
	}
	if err := database.PingContext(ctx); err != nil {
		return closeOnError(fmt.Errorf("connect to catalogue database: %w", err))
	}
	if _, err := database.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return closeOnError(fmt.Errorf("enable catalogue foreign keys: %w", err))
	}
	if _, err := database.ExecContext(ctx, "PRAGMA busy_timeout = 5000"); err != nil {
		return closeOnError(fmt.Errorf("configure catalogue busy timeout: %w", err))
	}
	if err := applyMigrations(ctx, database); err != nil {
		return closeOnError(err)
	}

	return &Catalog{database: database, newID: identifier.NewUUID}, nil
}

func (c *Catalog) Close() error {
	return c.database.Close()
}
