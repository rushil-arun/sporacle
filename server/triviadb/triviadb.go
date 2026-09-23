// Package triviadb provides access to the SQLite-backed trivia category store
// used by both the state package (to load a game's board items) and the
// trivia package (to expose categories/titles over HTTP).
package triviadb

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// FileName is the SQLite database file name, stored alongside the trivia/*.json seed files.
const FileName = "trivia.db"

const schema = `
CREATE TABLE IF NOT EXISTS categories (
	id INTEGER PRIMARY KEY,
	broad_category TEXT NOT NULL,
	narrow_category TEXT NOT NULL UNIQUE,
	items TEXT NOT NULL,
	created_by TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_categories_broad ON categories(broad_category);
`

// Open opens (creating and initializing if needed) the trivia database in dir.
func Open(dir string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", filepath.Join(dir, FileName))
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// Items returns the items for the given narrow category (title), or nil if not found.
func Items(db *sql.DB, title string) []string {
	var itemsJSON string
	if err := db.QueryRow(`SELECT items FROM categories WHERE narrow_category = ?`, title).Scan(&itemsJSON); err != nil {
		return nil
	}
	var items []string
	if json.Unmarshal([]byte(itemsJSON), &items) != nil {
		return nil
	}
	return items
}

// BroadCategories returns the distinct broad categories, sorted.
func BroadCategories(db *sql.DB) []string {
	rows, err := db.Query(`SELECT DISTINCT broad_category FROM categories ORDER BY broad_category`)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var b string
		if rows.Scan(&b) == nil {
			out = append(out, b)
		}
	}
	return out
}

// NarrowCategories returns the narrow categories (titles) under a broad category, sorted.
func NarrowCategories(db *sql.DB, broad string) []string {
	rows, err := db.Query(`SELECT narrow_category FROM categories WHERE broad_category = ? ORDER BY narrow_category`, broad)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var n string
		if rows.Scan(&n) == nil {
			out = append(out, n)
		}
	}
	return out
}

// Insert adds a new category, or replaces the existing one with the same narrow_category.
func Insert(db *sql.DB, broad, narrow string, items []string, createdBy string) error {
	itemsJSON, err := json.Marshal(items)
	if err != nil {
		return err
	}
	_, err = db.Exec(
		`INSERT INTO categories (broad_category, narrow_category, items, created_by) VALUES (?, ?, ?, ?)
		 ON CONFLICT(narrow_category) DO UPDATE SET
		   broad_category = excluded.broad_category,
		   items = excluded.items,
		   created_by = excluded.created_by`,
		broad, narrow, string(itemsJSON), createdBy,
	)
	if err != nil {
		return fmt.Errorf("insert category %q/%q: %w", broad, narrow, err)
	}
	return nil
}
