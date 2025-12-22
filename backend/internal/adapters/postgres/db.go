package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/vitaly-stepin/agile_party/internal/adapters/config"
)

type DB struct {
	*sql.DB
}

func NewDB(cfg *config.DatabaseConfig) (*DB, error) {
	db, err := sql.Open("postgres", cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	} // consider later if we need db ping

	return &DB{DB: db}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) RunMigrations() error { // use a migration tool later
	createMigrationsTable := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT NOW()
		);
	`
	if _, err := db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrations := []struct {
		version int
		name    string
		sql     string
	}{
		{
			version: 1,
			name:    "create_rooms_table",
			sql: `
				CREATE TABLE IF NOT EXISTS rooms (
					id VARCHAR(10) PRIMARY KEY,
					name VARCHAR(255) NOT NULL,
					voting_system VARCHAR(20) NOT NULL DEFAULT 'dbs_fibo',
					auto_reveal BOOLEAN NOT NULL DEFAULT false,
					created_at TIMESTAMP NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMP NOT NULL DEFAULT NOW()
				);
			`,
		},
	}

	for _, migration := range migrations {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = $1", migration.version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration %d: %w", migration.version, err)
		}

		if count > 0 {
			// Migration already applied, skip
			continue
		}

		if _, err := db.Exec(migration.sql); err != nil {
			return fmt.Errorf("failed to apply migration %d (%s): %w", migration.version, migration.name, err)
		}

		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migration.version)
		if err != nil {
			return fmt.Errorf("failed to record migration %d: %w", migration.version, err)
		}

		fmt.Printf("Applied migration %d: %s\n", migration.version, migration.name)
	}

	return nil
}

func (db *DB) PingContext(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}
