package trivia

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// RegisterRoutes registers trivia-related HTTP handlers onto the provided mux.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/trivia/files", getFilesHandler)
	mux.HandleFunc("/trivia/keys", getKeysHandler)
}

// getFilesHandler returns the list of filenames in the top-level `trivia` directory.
func getFilesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode([]byte("method not allowed"))
		return
	}

	dir := "trivia"
	entries, err := os.ReadDir(dir)
	if err != nil {
		// If directory doesn't exist or can't be read, return empty list.
		json.NewEncoder(w).Encode([]string{})
		return
	}

	files := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		files = append(files, name)
	}

	_ = json.NewEncoder(w).Encode(files)
}

// getKeysHandler returns all top-level keys found in the JSON file specified by
// the `file` query parameter. If the file doesn't exist, an empty list is returned.
func getKeysHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode([]byte("method not allowed"))
		return
	}

	q := r.URL.Query()
	fname := q.Get("file")
	if fname == "" {
		json.NewEncoder(w).Encode([]string{})
		return
	}

	// Defend against path traversal by taking base name and rejecting paths
	// that try to escape the directory.
	fname = filepath.Base(fname)
	// If user omitted .json, try both with and without extension.
	candidates := []string{fname}
	if !strings.HasSuffix(strings.ToLower(fname), ".json") {
		candidates = append(candidates, fname+".json")
	}

	var filePath string
	dir := "trivia"
	for _, c := range candidates {
		p := filepath.Join(dir, c)
		var err error
		if _, err = os.Stat(p); err == nil {
			filePath = p
			break
		}
		if !os.IsNotExist(err) {
			// unexpected error reading file; treat as not found
			json.NewEncoder(w).Encode([]string{})
			return
		}
	}

	if filePath == "" {
		json.NewEncoder(w).Encode([]string{})
		return
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		json.NewEncoder(w).Encode([]string{})
		return
	}

	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		json.NewEncoder(w).Encode([]string{})
		return
	}

	keysSet := make(map[string]struct{})

	switch val := v.(type) {
	case map[string]any:
		for k := range val {
			keysSet[k] = struct{}{}
		}
	case []any:
		for _, item := range val {
			if m, ok := item.(map[string]any); ok {
				for k := range m {
					keysSet[k] = struct{}{}
				}
			}
		}
	default:
		// other JSON types -> no keys
	}

	keys := make([]string, 0, len(keysSet))
	for k := range keysSet {
		keys = append(keys, k)
	}

	_ = json.NewEncoder(w).Encode(keys)
}
