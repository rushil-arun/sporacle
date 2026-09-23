package trivia_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"server/trivia"
)

// newTestServer spins up an httptest server backed by an isolated, empty trivia.db
// in a fresh temp directory, so tests never touch (or depend on) the repo's real
// trivia/ data.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	saved := trivia.TriviaBasePath
	trivia.TriviaBasePath = t.TempDir()
	t.Cleanup(func() { trivia.TriviaBasePath = saved })

	mux := http.NewServeMux()
	trivia.RegisterRoutes(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func createCategory(t *testing.T, srv *httptest.Server, body map[string]any) *http.Response {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	resp, err := http.Post(srv.URL+"/trivia/categories", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST /trivia/categories: %v", err)
	}
	return resp
}

func TestCreateCategory_Success(t *testing.T) {
	srv := newTestServer(t)

	resp := createCategory(t, srv, map[string]any{
		"broadCategory":  "sports",
		"narrowCategory": "NHL Teams",
		"items":          []string{"Bruins", "Maple Leafs", "Bruins", " Rangers "},
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}

	var created struct {
		BroadCategory  string   `json:"broadCategory"`
		NarrowCategory string   `json:"narrowCategory"`
		Items          []string `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	wantItems := []string{"Bruins", "Maple Leafs", "Rangers"}
	if len(created.Items) != len(wantItems) {
		t.Fatalf("items = %v, want %v (dedup/trim expected)", created.Items, wantItems)
	}
	for i, item := range wantItems {
		if created.Items[i] != item {
			t.Errorf("items[%d] = %q, want %q", i, created.Items[i], item)
		}
	}

	// The new category should now show up via the existing read endpoints.
	keysResp, err := http.Get(srv.URL + "/trivia/keys?file=sports")
	if err != nil {
		t.Fatalf("GET /trivia/keys: %v", err)
	}
	defer keysResp.Body.Close()
	var keys []string
	json.NewDecoder(keysResp.Body).Decode(&keys)
	if len(keys) != 1 || keys[0] != "NHL Teams" {
		t.Errorf("keys(sports) = %v, want [NHL Teams]", keys)
	}
}

func TestCreateCategory_DuplicateNarrow(t *testing.T) {
	srv := newTestServer(t)

	body := map[string]any{
		"broadCategory":  "sports",
		"narrowCategory": "NHL Teams",
		"items":          []string{"Bruins"},
	}
	resp1 := createCategory(t, srv, body)
	resp1.Body.Close()
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("first create status = %d, want %d", resp1.StatusCode, http.StatusCreated)
	}

	// Different broad category and items, but same narrow name — must be rejected.
	resp2 := createCategory(t, srv, map[string]any{
		"broadCategory":  "geography",
		"narrowCategory": "NHL Teams",
		"items":          []string{"Something Else"},
	})
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate create status = %d, want %d", resp2.StatusCode, http.StatusConflict)
	}
}

func TestCreateCategory_ValidationErrors(t *testing.T) {
	srv := newTestServer(t)

	cases := []struct {
		name string
		body map[string]any
	}{
		{"missing broad", map[string]any{"broadCategory": "", "narrowCategory": "X", "items": []string{"a"}}},
		{"missing narrow", map[string]any{"broadCategory": "X", "narrowCategory": "  ", "items": []string{"a"}}},
		{"no items", map[string]any{"broadCategory": "X", "narrowCategory": "Y", "items": []string{}}},
		{"only blank items", map[string]any{"broadCategory": "X", "narrowCategory": "Y", "items": []string{"  ", ""}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := createCategory(t, srv, c.body)
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
			}
		})
	}
}

func TestCreateCategory_MethodNotAllowed(t *testing.T) {
	srv := newTestServer(t)

	resp, err := http.Get(srv.URL + "/trivia/categories")
	if err != nil {
		t.Fatalf("GET /trivia/categories: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestSearchBroadAndNarrowCategories(t *testing.T) {
	srv := newTestServer(t)

	for _, c := range []map[string]any{
		{"broadCategory": "sports", "narrowCategory": "NBA Teams", "items": []string{"a"}},
		{"broadCategory": "sports", "narrowCategory": "NHL Teams", "items": []string{"a"}},
		{"broadCategory": "geography", "narrowCategory": "US States", "items": []string{"a"}},
	} {
		resp := createCategory(t, srv, c)
		resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("setup create failed: status = %d", resp.StatusCode)
		}
	}

	broadResp, err := http.Get(srv.URL + "/trivia/broad-categories?q=spo")
	if err != nil {
		t.Fatalf("GET /trivia/broad-categories: %v", err)
	}
	defer broadResp.Body.Close()
	var broad []string
	json.NewDecoder(broadResp.Body).Decode(&broad)
	if len(broad) != 1 || broad[0] != "sports" {
		t.Errorf("broad-categories(q=spo) = %v, want [sports]", broad)
	}

	narrowResp, err := http.Get(srv.URL + "/trivia/narrow-categories?q=team")
	if err != nil {
		t.Fatalf("GET /trivia/narrow-categories: %v", err)
	}
	defer narrowResp.Body.Close()
	var narrow []string
	json.NewDecoder(narrowResp.Body).Decode(&narrow)
	if len(narrow) != 2 {
		t.Errorf("narrow-categories(q=team) = %v, want 2 results", narrow)
	}
}
