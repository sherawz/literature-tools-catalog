package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const dbName = "ai_weights.db"

func main() {
	// Open database connection
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create table
	err = createTable(db)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	log.Println("Database and schema for AI weights created successfully.")
}

func createTable(db *sql.DB) error {
	// Set PRAGMA values
	pragmaQuery := `
	PRAGMA journal_mode = delete;
	PRAGMA page_size = 1024;
	`
	_, err := db.Exec(pragmaQuery)
	if err != nil {
		return err
	}

	query := `
	CREATE TABLE IF NOT EXISTS ai_weights (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		description TEXT
	);`
	_, err = db.Exec(query)
	return err
}
