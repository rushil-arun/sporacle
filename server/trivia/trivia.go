package trivia

import (
	"encoding/json"
	"net/http"

	"server/triviadb"
)

// TriviaBasePath is the path to the trivia directory (relative to server when run from server/).
var TriviaBasePath = "../trivia"

// RegisterRoutes registers trivia-related HTTP handlers onto the provided mux.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/trivia/files", getFilesHandler)
	mux.HandleFunc("/trivia/keys", getKeysHandler)
}

// getFilesHandler returns the distinct broad categories in trivia.db.
func getFilesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode([]byte("method not allowed"))
		return
	}

	db, err := triviadb.Open(TriviaBasePath)
	if err != nil {
		_ = json.NewEncoder(w).Encode([]string{})
		return
	}
	defer db.Close()

	_ = json.NewEncoder(w).Encode(triviadb.BroadCategories(db))
}

// getKeysHandler returns the narrow categories (titles) under the broad category
// specified by the `file` query parameter. If none are found, an empty list is returned.
func getKeysHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode([]byte("method not allowed"))
		return
	}

	broad := r.URL.Query().Get("file")
	if broad == "" {
		_ = json.NewEncoder(w).Encode([]string{})
		return
	}

	db, err := triviadb.Open(TriviaBasePath)
	if err != nil {
		_ = json.NewEncoder(w).Encode([]string{})
		return
	}
	defer db.Close()

	_ = json.NewEncoder(w).Encode(triviadb.NarrowCategories(db, broad))
}
