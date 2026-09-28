# px0 multi-repository workspaces

This is a proposed implementation of [px0-ai/px0#162](https://github.com/px0-ai/px0/issues/162): open several repositories (for example mobile, frontend, and backend) in one px0 session.

```bash
px0 ~/work/mobile ~/work/frontend ~/work/backend
```

## Design

Each repository gets its own unmodified px0 `Server` under `/r/<name>/`. This reuses px0's existing `-base-path` support. A small hub at `/` redirects to the first repository and serves `/api/workspace/list` for a repository switcher in the sidebar header.

This design has some benefits:
- **No `server.go` changes.** Every handler (git panel, LSP, search, agent edits, and diff) stays scoped to one root, so path sandboxing works the same way it does now.
- **Single-repo behavior is unchanged.** The hub is only used when you pass more than one directory.
- **No new dependencies and nothing written to disk.**

This design also has trade-offs:
- Search and quick-open are scoped to one repository at a time. They don't yet search across all repositories.
- Switching repositories reloads the page, so open tabs are kept per repository.
- Each repository has its own git watcher and LSP manager. LSP managers still start lazily.

## Files

| File | Target in px0 |
|---|---|
| `workspace.go` | new, repo root |
| `workspace_test.go` | new, repo root |
| `web/src/workspace.js` | new |
| `web/workspace.css` | append to `web/style.css` |
| `patches/main.go.md` | edits to `main.go` and `web/src/main.js` |

## Status

⚠️ **These files haven't been compiled or tested yet.** They were written against px0 at commit `9d82e10` using only functions verified in that source. Before opening a PR, run:

```bash
go test ./... && node ./scripts/build-web.js && go build -o px0 .
```

## Contributing upstream

px0's [CONTRIBUTING.md](https://github.com/px0-ai/px0/blob/master/CONTRIBUTING.md) says they *generally do not accept unsolicited pull requests*. They prefer detailed issues and implement changes with their own AI agents. The best route is to share this design as a comment on #162. You can open a PR from a real fork if the maintainers welcome one.
