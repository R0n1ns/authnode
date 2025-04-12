package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"authmicro/internal/config"
	"authmicro/internal/infrastructure/database"
	_ "github.com/lib/pq"
)

func main() {
	// Parse command line flags
	upFlag := flag.Bool("up", false, "Run migrations up")
	downFlag := flag.Bool("down", false, "Roll back migrations")
	helpFlag := flag.Bool("help", false, "Show help")
	flag.Parse()

	if *helpFlag {
		fmt.Println("Usage: migrate [-up] [-down]")
		fmt.Println("  -up    Run migrations up")
		fmt.Println("  -down  Roll back migrations")
		fmt.Println("  -help  Show this help")
		os.Exit(0)
	}

	if !*upFlag && !*downFlag {
		fmt.Println("Error: Must specify either -up or -down")
		fmt.Println("Use -help for more information")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if *upFlag {
		err = migrateUp(db)
		if err != nil {
			log.Fatalf("Failed to run migrations up: %v", err)
		}
		fmt.Println("Migrations completed successfully")
	}

	if *downFlag {
		err = migrateDown(db)
		if err != nil {
			log.Fatalf("Failed to roll back migrations: %v", err)
		}
		fmt.Println("Migrations rolled back successfully")
	}
}

func migrateUp(db *sql.DB) error {
	// Create migrations table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Run migrations
	migrations := []struct {
		name string
		sql  string
	}{
		{
			name: "create_users_table",
			sql: `
				CREATE TABLE IF NOT EXISTS users (
					id SERIAL PRIMARY KEY,
					first_name VARCHAR(255) NOT NULL,
					last_name VARCHAR(255) NOT NULL,
					nickname VARCHAR(255) NOT NULL UNIQUE,
					email VARCHAR(255) NOT NULL UNIQUE,
					email_verified BOOLEAN NOT NULL DEFAULT FALSE,
					role VARCHAR(50) NOT NULL,
					accepted_privacy_policy BOOLEAN NOT NULL DEFAULT FALSE,
					created_at TIMESTAMP NOT NULL,
					updated_at TIMESTAMP NOT NULL
				)
			`,
		},
		{
			name: "create_registration_sessions_table",
			sql: `
				CREATE TABLE IF NOT EXISTS registration_sessions (
					id VARCHAR(36) PRIMARY KEY,
					first_name VARCHAR(255) NOT NULL,
					last_name VARCHAR(255) NOT NULL,
					nickname VARCHAR(255) NOT NULL,
					email VARCHAR(255) NOT NULL,
					code VARCHAR(10) NOT NULL,
					code_expires TIMESTAMP NOT NULL,
					accepted_privacy_policy BOOLEAN NOT NULL DEFAULT FALSE,
					created_at TIMESTAMP NOT NULL
				)
			`,
		},
		{
			name: "create_login_sessions_table",
			sql: `
				CREATE TABLE IF NOT EXISTS login_sessions (
					email VARCHAR(255) PRIMARY KEY,
					code VARCHAR(10) NOT NULL,
					code_expires TIMESTAMP NOT NULL,
					created_at TIMESTAMP NOT NULL
				)
			`,
		},
		{
			name: "create_indexes",
			sql: `
				CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
				CREATE INDEX IF NOT EXISTS idx_users_nickname ON users(nickname);
				CREATE INDEX IF NOT EXISTS idx_registration_sessions_email ON registration_sessions(email);
			`,
		},
	}

	for _, migration := range migrations {
		// Check if migration has already been applied
		var exists bool
		err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM migrations WHERE name = $1)", migration.name).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if exists {
			fmt.Printf("Migration '%s' already applied, skipping\n", migration.name)
			continue
		}

		// Apply migration
		fmt.Printf("Applying migration '%s'...\n", migration.name)
		_, err = tx.Exec(migration.sql)
		if err != nil {
			return fmt.Errorf("failed to apply migration '%s': %w", migration.name, err)
		}

		// Record migration
		_, err = tx.Exec("INSERT INTO migrations (name) VALUES ($1)", migration.name)
		if err != nil {
			return fmt.Errorf("failed to record migration '%s': %w", migration.name, err)
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func migrateDown(db *sql.DB) error {
	// Check if migrations table exists
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'migrations'
		)
	`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if migrations table exists: %w", err)
	}

	if !exists {
		fmt.Println("No migrations to roll back")
		return nil
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Roll back migrations in reverse order
	rollbacks := []struct {
		name string
		sql  string
	}{
		{
			name: "create_indexes",
			sql: `
				DROP INDEX IF EXISTS idx_users_email;
				DROP INDEX IF EXISTS idx_users_nickname;
				DROP INDEX IF EXISTS idx_registration_sessions_email;
			`,
		},
		{
			name: "create_login_sessions_table",
			sql:  `DROP TABLE IF EXISTS login_sessions`,
		},
		{
			name: "create_registration_sessions_table",
			sql:  `DROP TABLE IF EXISTS registration_sessions`,
		},
		{
			name: "create_users_table",
			sql:  `DROP TABLE IF EXISTS users`,
		},
	}

	// Get applied migrations
	rows, err := tx.Query("SELECT name FROM migrations ORDER BY id DESC")
	if err != nil {
		return fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	var appliedMigrations []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("failed to scan migration name: %w", err)
		}
		appliedMigrations = append(appliedMigrations, name)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating migrations: %w", err)
	}

	// Apply rollbacks for applied migrations
	for _, rollback := range rollbacks {
		// Check if migration was applied
		var applied bool
		for _, m := range appliedMigrations {
			if m == rollback.name {
				applied = true
				break
			}
		}

		if !applied {
			fmt.Printf("Migration '%s' not applied, skipping rollback\n", rollback.name)
			continue
		}

		// Apply rollback
		fmt.Printf("Rolling back migration '%s'...\n", rollback.name)
		_, err = tx.Exec(rollback.sql)
		if err != nil {
			return fmt.Errorf("failed to roll back migration '%s': %w", rollback.name, err)
		}

		// Remove migration record
		_, err = tx.Exec("DELETE FROM migrations WHERE name = $1", rollback.name)
		if err != nil {
			return fmt.Errorf("failed to remove migration record '%s': %w", rollback.name, err)
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
