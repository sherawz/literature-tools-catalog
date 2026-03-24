package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbName   = "catalog.db"
	dataFile = "20260323.PMID_PMCID_DOI.csv.400K-random.txt"
)

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

	// Ingest data
	log.Println("Starting data ingestion...")
	err = ingestData(db, dataFile)
	if err != nil {
		log.Fatalf("Failed to ingest data: %v", err)
	}
	log.Println("Data ingestion completed successfully.")

	// Create indexes for faster search
	log.Println("Creating indexes...")
	err = createIndexes(db)
	if err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}
	log.Println("Indexes created successfully.")
}

func createTable(db *sql.DB) error {
	// Set PRAGMA values for sql.js-httpvfs compatibility before table creation
	pragmaQuery := `
	PRAGMA journal_mode = delete;
	PRAGMA page_size = 1024;
	`
	_, err := db.Exec(pragmaQuery)
	if err != nil {
		return err
	}

	query := `
	CREATE TABLE IF NOT EXISTS literature (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pmid TEXT,
		pmcid TEXT,
		doi TEXT
	);`
	_, err = db.Exec(query)
	return err
}

func createIndexes(db *sql.DB) error {
	queries := []string{
		`CREATE INDEX IF NOT EXISTS idx_pmid ON literature(pmid);`,
		`CREATE INDEX IF NOT EXISTS idx_pmcid ON literature(pmcid);`,
		`CREATE INDEX IF NOT EXISTS idx_doi ON literature(doi);`,
	}
	for _, q := range queries {
		_, err := db.Exec(q)
		if err != nil {
			return err
		}
	}
	return nil
}

func ingestData(db *sql.DB, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Some records might have wrong number of fields, let's just make it robust.
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	// Use a transaction for bulk insert
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}

	stmt, err := tx.Prepare(`INSERT INTO literature(pmid, pmcid, doi) VALUES(?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("could not prepare statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Warning: error reading record at line %d: %v", count+1, err)
			continue
		}

		if len(record) < 3 {
			// Skip or handle incomplete records. The sample had exactly 3 columns.
			continue
		}

		pmid := record[0]
		pmcid := record[1]
		doi := record[2]

		_, err = stmt.Exec(pmid, pmcid, doi)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("could not execute statement for record %v: %w", record, err)
		}

		count++
		if count%100000 == 0 {
			log.Printf("Inserted %d records...", count)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}

	// Reorganize database and apply changed page size
	log.Println("Vacuuming database to apply page size...")
	_, err = db.Exec("VACUUM;")
	if err != nil {
		return fmt.Errorf("could not vacuum database: %w", err)
	}

	log.Printf("Total %d records inserted.", count)
	return nil
}
