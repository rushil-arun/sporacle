// migrate-trivia reads the trivia/*.json seed files at the repo root and
// (re)populates trivia/trivia.db, the SQLite database the server reads from.
//
// Run from the server/ directory:
//
//	go run ./scripts/migrate-trivia
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"server/triviadb"
)

const triviaDir = "../trivia"

func main() {
	entries, err := os.ReadDir(triviaDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", triviaDir, err)
		os.Exit(1)
	}

	db, err := triviadb.Open(triviaDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open trivia.db: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	total := 0
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}

		path := filepath.Join(triviaDir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", path, err)
			os.Exit(1)
		}

		var obj map[string][]string
		if err := json.Unmarshal(data, &obj); err != nil {
			fmt.Fprintf(os.Stderr, "parse %s: %v\n", path, err)
			os.Exit(1)
		}

		broad := strings.TrimSuffix(e.Name(), ".json")
		for narrow, items := range obj {
			if err := triviadb.Insert(db, broad, narrow, dedupe(items), ""); err != nil {
				fmt.Fprintf(os.Stderr, "insert %s/%s: %v\n", broad, narrow, err)
				os.Exit(1)
			}
			total++
		}
		fmt.Printf("migrated %s (%d categories)\n", e.Name(), len(obj))
	}
	fmt.Printf("done: %d categories in %s\n", total, filepath.Join(triviaDir, triviadb.FileName))
}

// dedupe removes duplicate items while preserving order (the seed JSON has a few).
func dedupe(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}
