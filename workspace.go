package main

// Multi-repository workspaces (px0-ai/px0#162).
//
// Design: instead of threading a "current repo" through every handler in
// server.go, each repository gets its own, unmodified *Server mounted under
// its own base path (/r/<name>/). px0 already supports base paths end to end
// (-base-path, <base href>, document.baseURI in web/src/state.js), so every
// existing API, the git panel, LSP, search and agent editing keep working
// per repo with zero changes. A thin hub at the root lists the repos and
// redirects "/" to the first one.
//
// Single-repo invocations never touch this file: main.go only builds a hub
// when more than one directory is passed.

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

// hubRepo is one repository mounted in a workspace hub.
type hubRepo struct {
	Name  string `json:"name"` // unique, URL-safe display name
	Root  string `json:"root"` // absolute workspace root
	Path  string `json:"path"` // base path it is served under, e.g. /r/backend/
	srv   *Server
	ix    *Index
	lsp   *lspManager
	agent *agentManager
}

// workspaceHub routes requests to the per-repo servers.
type workspaceHub struct {
	repos []*hubRepo
	mux   *http.ServeMux
}

// repoSlug turns a directory name into something safe for a URL segment.
func repoSlug(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	s := strings.Trim(b.String(), "-.")
	if s == "" {
		s = "repo"
	}
	return s
}

// uniqueRepoNames assigns each root a unique slug, suffixing duplicates
// (two checkouts both called "app" become app and app-2).
func uniqueRepoNames(roots []string) []string {
	seen := map[string]int{}
	out := make([]string, len(roots))
	for i, root := range roots {
		base := repoSlug(filepath.Base(root))
		seen[base]++
		if n := seen[base]; n > 1 {
			out[i] = base + "-" + strconv.Itoa(n)
		} else {
			out[i] = base
		}
	}
	return out
}

// newWorkspaceHub builds one Server per root. basePath is the global
// -base-path prefix ("/" normally); repos mount beneath it.
func newWorkspaceHub(roots []string, basePath string, useLSP bool, agentCmd string, noAgent bool) (*workspaceHub, error) {
	if len(roots) == 0 {
		return nil, fmt.Errorf("no repositories given")
	}
	basePath = cleanBasePath(basePath)
	h := &workspaceHub{mux: http.NewServeMux()}
	names := uniqueRepoNames(roots)
	for i, root := range roots {
		bp := basePath + "r/" + names[i] + "/"
		ix := NewIndex(root)
		lsp := newLSPManager(root, useLSP)
		srv := NewServer(ix, lsp, bp)
		// sessionFilePath keys on a non-root base path, so every workspace with
		// a repo named "app" would share r_app.json. Key on the root instead,
		// exactly like a single-repo run of the same directory.
		srv.session = newSessionManager("/", root)
		r := &hubRepo{Name: names[i], Root: root, Path: bp, srv: srv, ix: ix, lsp: lsp}
		if !noAgent {
			a, err := newAgentManager(root, agentCmd, lsp)
			if err != nil {
				return nil, fmt.Errorf("-agent: %w", err)
			}
			srv.SetAgent(a)
			r.agent = a
		}
		h.repos = append(h.repos, r)
		// Server registers its routes with the full prefix, so no StripPrefix.
		h.mux.Handle(bp, srv)
		h.mux.Handle(strings.TrimSuffix(bp, "/"), srv)
	}

	listPath := basePath + "api/workspace/list"
	h.mux.HandleFunc(listPath, h.handleList)
	h.mux.HandleFunc(basePath, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != basePath && r.URL.Path != strings.TrimSuffix(basePath, "/") {
			http.NotFound(w, r)
			return
		}
		target := h.repos[0].Path
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusFound)
	})
	return h, nil
}

func (h *workspaceHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// handleList powers the repo switcher (web/src/workspace.js).
func (h *workspaceHub) handleList(w http.ResponseWriter, r *http.Request) {
	type item struct {
		Name       string `json:"name"`
		Root       string `json:"root"`
		Path       string `json:"path"`
		Files      int    `json:"files"`
		Ready      bool   `json:"ready"`
		GitChanges int    `json:"gitChanges"`
	}
	out := make([]item, 0, len(h.repos))
	for _, rp := range h.repos {
		n, _, _ := rp.ix.Stats()
		gc, _ := rp.ix.GitChanges()
		out = append(out, item{rp.Name, rp.Root, rp.Path, n, rp.ix.Ready(), gc})
	}
	writeJSON(w, map[string]any{"repos": out})
}

// Build indexes every repo; call from a goroutine like the single-repo path.
func (h *workspaceHub) Build() {
	for _, rp := range h.repos {
		rp.ix.Build()
	}
}

// Close shuts down language servers and agents for every repo.
func (h *workspaceHub) Close() {
	for _, rp := range h.repos {
		rp.lsp.Close()
		rp.agent.Close()
	}
}
