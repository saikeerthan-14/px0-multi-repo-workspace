# Upstream contribution notes

This repository is a copy of [px0-ai/px0](https://github.com/px0-ai/px0) at commit
[`9d82e10`](https://github.com/px0-ai/px0/commit/9d82e10f6a2ef1307a0c76bcd8f34628e14b155b)
(v0.1.9, branch `master`), plus multi-repository workspace support for
[px0-ai/px0#162](https://github.com/px0-ai/px0/issues/162).

> **Read this first.** px0's [CONTRIBUTING.md](https://github.com/px0-ai/px0/blob/master/CONTRIBUTING.md)
> says the project is built entirely by AI agents and *"generally do[es] not accept unsolicited pull
> requests containing manual or disparate code patches."* It asks for detailed issues and feature ideas
> instead. Post the comment in section 1 on #162 first. Only follow section 2 if a maintainer asks
> for a PR.

The change touches exactly 11 files relative to `9d82e10` (477 insertions, 1 deletion):

| File | Change |
| --- | --- |
| `workspace.go` | new: `workspaceHub`, `uniqueRepoNames`, `/api/workspace/list` |
| `workspace_test.go` | new: naming, routing, `-base-path`, `<base href>`, session-keying tests |
| `web/src/workspace.js` | new: sidebar repository switcher |
| `main.go` | multi-directory branch in `main()` and a usage line |
| `web/src/main.js` | import and call `initWorkspace()` after meta loads |
| `web/style.css` | `.repo-switch` styles |
| `README.md` | Usage example |
| `docs/features/workspaces.md` | new feature guide |
| `docs/features/README.md` | feature matrix and section 7 entry |
| `docs/internals/architecture.md` | "Multi-Repository Workspace Hub" section |
| `docs/agents/README.md` | module table and documentation-mapping row |

---

## 1. Ready-to-paste comment for #162

````markdown
I prototyped this against `9d82e10` (v0.1.9) to see how small it could be. Sharing the design in case it's useful for your agents. I haven't opened a PR, per CONTRIBUTING.md.

**Usage**

```bash
px0 ~/work/mobile ~/work/frontend ~/work/backend
```

**Design: one unmodified `Server` per repository, mounted under a base path**

px0 already supports `-base-path` end to end (route prefixing, the injected `<base href>`, and `document.baseURI` in `state.js`). So instead of threading a "current repo" through every handler, a small `workspaceHub` (~160 lines, new `workspace.go`) does the following:

- It builds one `NewServer(ix, lsp, "<base>r/<name>/")` per directory, each with its own `Index`, `lspManager`, `GitWatcher`, and `agentManager`, and mounts it on an `http.ServeMux`.
- It redirects `<base>` to the first repository.
- It serves `<base>api/workspace/list` (`name`, `root`, `path`, `files`, `ready`, `gitChanges`) for a sidebar switcher. The switcher (`web/src/workspace.js`, ~30 lines) replaces `#root-name` with a `<select>` and does nothing in single-repo sessions.
- Names are URL-safe directory basenames. Duplicates get suffixes (`app`, `app-2`).
- `-base-path /x/` (and `server.basePath`) still work: repositories are served at `/x/r/<name>/`.

`main.go` gains one early branch that only runs when you pass more than one directory and no PR URL. `server.go` is untouched, so the single-directory, file-target, and PR paths are byte-for-byte the same.

**One wrinkle:** `sessionFilePath` keys session files on the base path when it isn't `/`, so every workspace with an `app` repository would share `r_app.json`. The hub re-creates each server's `sessionManager` keyed on the root, which is exactly what a single-repo run of that directory uses.

**Trade-offs**

- Quick open, search, and the tree are scoped to one repository. Cross-repo search would need a fan-out over the hub's servers.
- Switching repositories reloads the page. Tabs and open folders are restored per repository through the existing session.
- Each repository pays for its own index and git watcher. LSP still starts lazily per repository.
- No new dependencies, same single binary, no new files on disk.

**Verification** (Go 1.25 toolchain, Node 24)

- `node ./scripts/build-web.js`, `go vet ./...`, `go build`, and `go test ./...` are all green, including 4 new tests: naming, routing, `-base-path` with `<base href>`, and session keying.
- Smoke test with two git repos (`px0 -no-open -port 0 alpha beta`):
  - `/` redirects 302 to `/r/alpha/`.
  - `/r/alpha/api/meta` and `/r/beta/api/meta` report their own `root`.
  - `/api/workspace/list` lists both repos.
  - `/r/beta/` serves `index.html` with `<base href="/r/beta/">`.
  - `-base-path /x/` serves `/x/r/<name>/`.
- Single-repo: `/`, `/api/meta`, `/api/tree`, and `index.html` are identical to an upstream v0.1.9 build of the same directory.

Happy to share the full diff (477 insertions and 1 deletion across 11 files, including docs) if you'd like it as a reference.
````

---

## 2. Opening a PR against px0-ai/px0 (only if maintainers ask for one)

These steps need a GitHub account, `git`, Go 1.24+ (go.mod pulls the 1.25 toolchain automatically), and Node or Bun.

1. **Fork upstream.** Open <https://github.com/px0-ai/px0> and click **Fork**, or run:
   ```bash
   gh repo fork px0-ai/px0 --clone --remote
   cd px0
   ```
   Without `gh`, clone your fork: `git clone https://github.com/<you>/px0 && cd px0 && git remote add upstream https://github.com/px0-ai/px0`.

2. **Branch from the exact base commit.**
   ```bash
   git fetch upstream master
   git checkout -b multi-repo-workspaces 9d82e10f6a2ef1307a0c76bcd8f34628e14b155b
   ```

3. **Bring in the 11 changed files from this repository.** The paths are identical, so check them out directly:
   ```bash
   git fetch https://github.com/saikeerthan-14/px0-multi-repo-workspace main
   git checkout FETCH_HEAD -- \
     workspace.go workspace_test.go web/src/workspace.js \
     main.go web/src/main.js web/style.css \
     README.md docs/features/workspaces.md docs/features/README.md \
     docs/internals/architecture.md docs/agents/README.md
   git diff --cached --stat   # expect 11 files changed, 477 insertions(+), 1 deletion(-)
   ```
   If the change hasn't been merged into this repository's `main` yet, fetch the `copilot/add-multi-repo-workspace-support` branch instead.
   Don't copy anything else from this repository: `UPSTREAM.md` only exists here.

4. **Verify.**
   ```bash
   node ./scripts/build-web.js
   go vet ./...
   go test ./...
   go build -o px0 .
   ./px0 -no-open -port 0 /path/to/repoA /path/to/repoB   # then curl / , /r/<name>/api/meta, /api/workspace/list
   ```

5. **Rebase if upstream moved.** If `master` is past `9d82e10`, run `git rebase upstream/master` and resolve any conflicts. `main.go` and the docs are the files most likely to conflict. Then repeat step 4.

6. **Commit, push, and open the PR.**
   ```bash
   git commit -m "Add multi-repository workspaces (#162)"
   git push -u origin multi-repo-workspaces
   gh pr create --repo px0-ai/px0 --base master \
     --title "Add multi-repository workspaces (#162)" \
     --body "Closes #162. Design and verification: see my comment on #162."
   ```
   Or open `https://github.com/<you>/px0/pull/new/multi-repo-workspaces` and target `px0-ai/px0:master`.
