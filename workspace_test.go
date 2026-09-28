package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestUniqueRepoNames(t *testing.T) {
	got := uniqueRepoNames([]string{"/a/app", "/b/app", "/c/My Backend"})
	want := []string{"app", "app-2", "my-backend"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("uniqueRepoNames = %v, want %v", got, want)
		}
	}
}

func TestWorkspaceHubRouting(t *testing.T) {
	var roots []string
	for _, n := range []string{"frontend", "backend"} {
		d := filepath.Join(t.TempDir(), n)
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		os.WriteFile(filepath.Join(d, "main.go"), []byte("package main\n"), 0o644)
		roots = append(roots, d)
	}
	h, err := newWorkspaceHub(roots, "/", false, "", true)
	if err != nil {
		t.Fatal(err)
	}
	h.Build()

	// Root redirects to the first repo.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/r/frontend/" {
		t.Fatalf("GET / = %d %q", rec.Code, rec.Header().Get("Location"))
	}

	// Each repo's API is scoped to its own root.
	for _, c := range []struct{ path, root string }{
		{"/r/frontend/api/meta", roots[0]},
		{"/r/backend/api/meta", roots[1]},
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, c.path, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s = %d", c.path, rec.Code)
		}
		var meta map[string]any
		json.Unmarshal(rec.Body.Bytes(), &meta)
		if meta["root"] != c.root {
			t.Errorf("%s root = %v, want %s", c.path, meta["root"], c.root)
		}
	}

	// Switcher list.
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/workspace/list", nil))
	var list struct{ Repos []map[string]any }
	json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list.Repos) != 2 {
		t.Fatalf("workspace list = %s", rec.Body.String())
	}
}
