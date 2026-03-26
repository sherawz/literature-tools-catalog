package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const dbName = "tools_materials.db"

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

	log.Println("Database and schema for tools and materials created successfully.")
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
	CREATE TABLE IF NOT EXISTS tools_materials (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		type TEXT,
		description TEXT
	);`
	_, err = db.Exec(query)
	return err
}
