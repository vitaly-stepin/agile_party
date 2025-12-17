package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/vitaly-stepin/agile_party/internal/adapters/config"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// DB wraps sql.DB with additional methods
type DB struct {
	*sql.DB
}

// NewDB creates a new database connection with connection pooling
func NewDB(cfg *config.DatabaseConfig) (*DB, error) {
	db, err := sql.Open("postgres", cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// RunMigrations executes SQL migrations
func (db *DB) RunMigrations() error {
	// Create migrations table if not exists
	createMigrationsTable := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT NOW()
		);
	`
	if _, err := db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Define migrations in order
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

	// Apply migrations
	for _, migration := range migrations {
		// Check if migration already applied
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = $1", migration.version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration %d: %w", migration.version, err)
		}

		if count > 0 {
			// Migration already applied, skip
			continue
		}

		// Apply migration
		if _, err := db.Exec(migration.sql); err != nil {
			return fmt.Errorf("failed to apply migration %d (%s): %w", migration.version, migration.name, err)
		}

		// Record migration
		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migration.version)
		if err != nil {
			return fmt.Errorf("failed to record migration %d: %w", migration.version, err)
		}

		fmt.Printf("Applied migration %d: %s\n", migration.version, migration.name)
	}

	return nil
}

// Ping verifies database connectivity with context
func (db *DB) PingContext(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}
