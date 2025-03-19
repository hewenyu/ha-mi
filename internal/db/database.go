package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB is our database wrapper
type DB struct {
	*sql.DB
}

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return &DB{DB: db}, nil
}

// Initialize creates all the necessary tables
func (db *DB) Initialize() error {
	// Create nonce table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS nonces (
			nonce TEXT PRIMARY KEY,
			expires_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating nonces table: %w", err)
	}

	// Create zones table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS zones (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating zones table: %w", err)
	}

	// Create device types table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS device_types (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating device_types table: %w", err)
	}

	// Create operations table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS operations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			device_type_id INTEGER NOT NULL,
			description TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			UNIQUE(name, device_type_id),
			FOREIGN KEY(device_type_id) REFERENCES device_types(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating operations table: %w", err)
	}

	// Create mappings table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS mappings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			zone_id INTEGER NOT NULL,
			device_type_id INTEGER NOT NULL,
			operation_id INTEGER NOT NULL,
			entity_id TEXT NOT NULL,
			service TEXT NOT NULL,
			params TEXT,
			value_mapping TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			UNIQUE(zone_id, device_type_id, operation_id),
			FOREIGN KEY(zone_id) REFERENCES zones(id) ON DELETE CASCADE,
			FOREIGN KEY(device_type_id) REFERENCES device_types(id) ON DELETE CASCADE,
			FOREIGN KEY(operation_id) REFERENCES operations(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating mappings table: %w", err)
	}

	// Create scenes table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS scenes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			scene_id TEXT NOT NULL UNIQUE,
			description TEXT,
			actions TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating scenes table: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}
