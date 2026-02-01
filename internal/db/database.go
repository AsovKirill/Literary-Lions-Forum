package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB opens or creates an SQLite database at the provided path,
// initializes its schema from "internal/db/schema.sql",
// and returns the database connection.
// Returns an error if any step fails
func InitDB(path string) (*sql.DB, error) {
	fmt.Println("🔄 Opening database at:", path)

	// Open the SQLite DB (creates if not exist)
	dbConn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("❌ Failed to open database: %w", err)
	}

	// Test DB connectivity (catch early file/driver errors)
	err = dbConn.Ping()
	if err != nil {
		return nil, fmt.Errorf("❌ Cannot connect to database: %w", err)
	}

	// Load schema.sql file from project (required tables, indices, etc)
	fmt.Println("📖 Reading schema file: internal/db/schema.sql")
	schema, err := os.ReadFile("internal/db/schema.sql")
	if err != nil {
		return nil, fmt.Errorf("❌ Failed to read schema.sql: %w", err)
	}

	// Run SQL schema commands to create tables (if not exists)
	fmt.Println("🚀 Executing schema...")
	_, err = dbConn.Exec(string(schema))
	if err != nil {
		return nil, fmt.Errorf("❌ Failed to execute schema: %w", err)
	}

	fmt.Println("✅ Database initialized successfully.")
	return dbConn, nil
}
