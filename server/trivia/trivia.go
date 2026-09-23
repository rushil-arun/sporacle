package trivia

import (
	"encoding/json"
	"net/http"
	"strings"

	"server/triviadb"
)

// TriviaBasePath is the path to the trivia directory (relative to server when run from server/).
var TriviaBasePath = "../trivia"

const (
	maxCategoryNameLen = 100
	maxItems           = 500
	searchLimit        = 20
)

// RegisterRoutes registers trivia-related HTTP handlers onto the provided mux.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/trivia/files", getFilesHandler)
	mux.HandleFunc("/trivia/keys", getKeysHandler)
	mux.HandleFunc("/trivia/broad-categories", getBroadCategoriesHandler)
	mux.HandleFunc("/trivia/narrow-categories", getNarrowCategoriesHandler)
	mux.HandleFunc("/trivia/categories", categoriesHandler)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// getFilesHandler returns the distinct broad categories in trivia.db.
func getFilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	db, err := triviadb.Open(TriviaBasePath)
	if err != nil {
		writeJSON(w, http.StatusOK, []string{})
		return
	}
	defer db.Close()

	writeJSON(w, http.StatusOK, triviadb.BroadCategories(db))
}

// getKeysHandler returns the narrow categories (titles) under the broad category
// specified by the `file` query parameter. If none are found, an empty list is returned.
func getKeysHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	broad := r.URL.Query().Get("file")
	if broad == "" {
		writeJSON(w, http.StatusOK, []string{})
		return
	}

	db, err := triviadb.Open(TriviaBasePath)
	if err != nil {
		writeJSON(w, http.StatusOK, []string{})
		return
	}
	defer db.Close()

	writeJSON(w, http.StatusOK, triviadb.NarrowCategories(db, broad))
}

// getBroadCategoriesHandler returns broad categories matching the `q` query parameter,
// for autocomplete. An empty/missing `q` matches everything (up to searchLimit).
func getBroadCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	q := r.URL.Query().Get("q")

	db, err := triviadb.Open(TriviaBasePath)
	if err != nil {
		writeJSON(w, http.StatusOK, []string{})
		return
	}
	defer db.Close()

	writeJSON(w, http.StatusOK, triviadb.SearchBroadCategories(db, q, searchLimit))
}

// getNarrowCategoriesHandler returns narrow categories (titles) matching the `q` query
// parameter, across all broad categories, for autocomplete and duplicate-name checking.
func getNarrowCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	q := r.URL.Query().Get("q")

	db, err := triviadb.Open(TriviaBasePath)
	if err != nil {
		writeJSON(w, http.StatusOK, []string{})
		return
	}
	defer db.Close()

	writeJSON(w, http.StatusOK, triviadb.SearchNarrowCategories(db, q, searchLimit))
}

type createCategoryRequest struct {
	BroadCategory  string   `json:"broadCategory"`
	NarrowCategory string   `json:"narrowCategory"`
	Items          []string `json:"items"`
	CreatedBy      string   `json:"createdBy"`
}

// categoriesHandler creates a new category (broad + narrow + items). Every successful
// call creates a brand-new narrow category; it never modifies an existing one.
func categoriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req createCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	broad := strings.TrimSpace(req.BroadCategory)
	narrow := strings.TrimSpace(req.NarrowCategory)
	createdBy := strings.TrimSpace(req.CreatedBy)

	if broad == "" || len(broad) > maxCategoryNameLen {
		writeError(w, http.StatusBadRequest, "broadCategory must be 1-100 characters")
		return
	}
	if narrow == "" || len(narrow) > maxCategoryNameLen {
		writeError(w, http.StatusBadRequest, "narrowCategory must be 1-100 characters")
		return
	}

	items := dedupeItems(req.Items)
	if len(items) == 0 {
		writeError(w, http.StatusBadRequest, "items must contain at least 1 non-empty entry")
		return
	}
	if len(items) > maxItems {
		writeError(w, http.StatusBadRequest, "items must contain at most 500 entries")
		return
	}

	db, err := triviadb.Open(TriviaBasePath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to open trivia database")
		return
	}
	defer db.Close()

	if err := triviadb.Create(db, broad, narrow, items, createdBy); err != nil {
		if err == triviadb.ErrDuplicateCategory {
			writeError(w, http.StatusConflict, "a category with that narrow name already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create category")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"broadCategory":  broad,
		"narrowCategory": narrow,
		"items":          items,
	})
}

// dedupeItems trims each item, drops empties, and removes case-insensitive duplicates
// while preserving first-seen order.
func dedupeItems(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}
