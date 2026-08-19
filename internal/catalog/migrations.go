package catalog

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

var migrations = []string{"migrations/001_initial.sql"}

func applyMigrations(ctx context.Context, database *sql.DB) error {
	var current int
	if err := database.QueryRowContext(ctx, "PRAGMA user_version").Scan(&current); err != nil {
		return fmt.Errorf("read catalogue schema version: %w", err)
	}
	if current > len(migrations) {
		return fmt.Errorf("catalogue schema version %d is newer than supported version %d", current, len(migrations))
	}

	for index := current; index < len(migrations); index++ {
		sqlText, err := migrationFiles.ReadFile(migrations[index])
		if err != nil {
			return fmt.Errorf("read catalogue migration %d: %w", index+1, err)
		}
		transaction, err := database.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin catalogue migration %d: %w", index+1, err)
		}
		if _, err := transaction.ExecContext(ctx, string(sqlText)); err != nil {
			_ = transaction.Rollback()
			return fmt.Errorf("apply catalogue migration %d: %w", index+1, err)
		}
		if _, err := transaction.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d", index+1)); err != nil {
			_ = transaction.Rollback()
			return fmt.Errorf("record catalogue migration %d: %w", index+1, err)
		}
		if err := transaction.Commit(); err != nil {
			return fmt.Errorf("commit catalogue migration %d: %w", index+1, err)
		}
	}
	return nil
}
