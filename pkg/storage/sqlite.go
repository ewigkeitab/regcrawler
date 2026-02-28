package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"regcrawler/pkg/logger"
	"regcrawler/pkg/models"
)

var db *sql.DB

// InitDB initializes the SQLite database
func InitDB() error {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get user cache dir: %w", err)
	}

	appDir := filepath.Join(cacheDir, "regcrawler")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return fmt.Errorf("failed to create app data dir: %w", err)
	}

	dbPath := filepath.Join(appDir, "data.db")
	logger.Info("Initializing SQLite database at %s", dbPath)

	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	db = database

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS unprocessed_regulations (
		link TEXT PRIMARY KEY,
		title TEXT,
		date TEXT,
		category TEXT,
		content TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS processed_regulations (
		link TEXT PRIMARY KEY,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// SaveUnprocessed saves a regulation to the database, ignoring if it already exists
func SaveUnprocessed(reg models.Regulation) error {
	insertSQL := `INSERT OR IGNORE INTO unprocessed_regulations (link, title, date, category, content) VALUES (?, ?, ?, ?, ?)`
	_, err := db.Exec(insertSQL, reg.Link, reg.Title, reg.Date, reg.Category, reg.Content)
	return err
}

// GetUnprocessed retrieves all unprocessed regulations, sorted by oldest first
func GetUnprocessed() ([]models.Regulation, error) {
	// Order by created_at ASC to process older items first
	querySQL := `SELECT link, title, date, category, content FROM unprocessed_regulations ORDER BY created_at ASC`
	rows, err := db.Query(querySQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regs []models.Regulation
	for rows.Next() {
		var reg models.Regulation
		err = rows.Scan(&reg.Link, &reg.Title, &reg.Date, &reg.Category, &reg.Content)
		if err != nil {
			logger.Error("Error scanning database row: %v", err)
			continue
		}
		regs = append(regs, reg)
	}
	return regs, nil
}

// DeleteProcessed removes a successfully processed regulation from the database
func DeleteProcessed(link string) error {
	deleteSQL := `DELETE FROM unprocessed_regulations WHERE link = ?`
	_, err := db.Exec(deleteSQL, link)
	return err
}

// MarkProcessed adds a link to the successfully processed tracking table
func MarkProcessed(link string) error {
	insertSQL := `INSERT OR IGNORE INTO processed_regulations (link) VALUES (?)`
	_, err := db.Exec(insertSQL, link)
	return err
}

// HasBeenProcessed checks if a regulation has already been successfully processed
func HasBeenProcessed(link string) bool {
	var count int
	err := db.QueryRow(`SELECT COUNT(1) FROM processed_regulations WHERE link = ?`, link).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// IsUnprocessed checks if a regulation is currently waiting in the retry queue
func IsUnprocessed(link string) bool {
	var count int
	err := db.QueryRow(`SELECT COUNT(1) FROM unprocessed_regulations WHERE link = ?`, link).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}
