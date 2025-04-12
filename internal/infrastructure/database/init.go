package database

import (
	"database/sql"
	"fmt"
)

// InitDatabase initializes the database with extension for UUID generation
func InitDatabase(db *sql.DB) error {
	// Create uuid extension if not exists
	_, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)
	if err != nil {
		return fmt.Errorf("failed to create uuid extension: %w", err)
	}

	return nil
}
