package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func Migrate(parentContext context.Context) error {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "app.db"
	}

	// Open connection with typical production settings for SQLite
	db, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(parentContext, 2*time.Minute)
	defer cancel()

	if err := runSQLiteMigrations(ctx, db); err != nil {
		return err
	}

	log.Println("SQLite database is fully up to date!")
	return nil
}

func runSQLiteMigrations(ctx context.Context, db *sql.DB) error {
	var currentVersion int
	err := db.QueryRowContext(ctx, "PRAGMA user_version;").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("failed to read user_version pragma: %w", err)
	}

	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to read embedded migrations dir: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	// 3. Evaluate and run pending scripts
	for _, file := range files {
		// Extract integer version prefix (e.g., "0004_add_index.up.sql" -> 4)
		parts := strings.SplitN(file, "_", 2)
		fileVersion, err := strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Errorf("invalid migration filename format prefix for %s: %w", file, err)
		}

		// Skip if database is already at or past this version number
		if fileVersion <= currentVersion {
			log.Printf("Skipping applied migration: %s (DB Version: %d)", file, currentVersion)
			continue
		}

		log.Printf("Applying SQLite migration version %d (%s)...", fileVersion, file)

		content, err := migrationFS.ReadFile("migrations/" + file)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", file, err)
		}

		// Run everything inside an exclusive transaction block
		if err := executeSQLiteTx(ctx, db, fileVersion, string(content)); err != nil {
			return fmt.Errorf("migration failure at script version %d: %w", fileVersion, err)
		}

		currentVersion = fileVersion
	}

	return nil
}

func executeSQLiteTx(ctx context.Context, db *sql.DB, version int, sqlContent string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration SQL payload
	if _, err := tx.ExecContext(ctx, sqlContent); err != nil {
		return fmt.Errorf("SQL execution failed: %w", err)
	}

	// Update the internal SQLite version metadata header
	// Note: PRAGMA statements cannot accept standard SQL parameter placeholders ($1)
	versionUpdateQuery := fmt.Sprintf("PRAGMA user_version = %d;", version)
	if _, err := tx.ExecContext(ctx, versionUpdateQuery); err != nil {
		return fmt.Errorf("failed to update user_version pragma: %w", err)
	}

	return tx.Commit()
}
