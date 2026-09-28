# Multi-Repository Workspaces

You can open several repositories in one px0 session, for example a mobile app, a frontend, and a backend, and switch between them from the sidebar. This feature addresses [px0-ai/px0#162](https://github.com/px0-ai/px0/issues/162).

```bash
px0 ~/work/mobile ~/work/frontend ~/work/backend
```

---

## Overview & Core Purpose

Many changes span more than one repository: an API change in the backend, the client that calls it, and the app that ships it. Without this feature, reviewing those repositories side by side means running one px0 per repository on separate ports and managing several browser tabs. With a multi-repository workspace, one px0 process serves all of them on one port.

---

## Key Capabilities

- **One command, one port**: Pass two or more directories. px0 prints each workspace root and a single URL.
- **Repository switcher**: When more than one repository is open, the repository name in the sidebar header becomes a drop-down menu. Pick a repository to switch to it. Repositories with uncommitted changes show a count, for example `backend (3)`.
- **Everything stays per-repository**: Each repository keeps its own file index, git panel (Stage/Commit/Push/Pull), gutter diffs, language servers, coding harness, and saved tabs. Path sandboxing works the same way as in a single-repo session.
- **Stable, shareable URLs**: Each repository is served at `/r/<name>/`, where `<name>` is the directory name in lowercase and URL-safe form. If two directories have the same name, the second gets a suffix (`app`, `app-2`). Opening `/` redirects to the first repository.
- **Works behind a subpath**: `-base-path /x/` (or `server.basePath` in settings) serves the repositories at `/x/r/<name>/`.
- **Single-repo sessions are unchanged**: `px0`, `px0 <dir>`, `px0 <file>:<line>`, and `px0 <pr-url>` behave exactly as before.

---

## Usage

```bash
# Three repositories on the default port
px0 ~/work/mobile ~/work/frontend ~/work/backend

# Headless, on a free port
px0 -no-open -port 0 ./api ./web

# Behind a reverse proxy: served at /x/r/api/ and /x/r/web/
px0 -base-path /x/ ./api ./web
```

Every argument must be a directory. File targets such as `main.go:42` and PR URLs only work in single-repo mode. If a directory is inside a git repository, the repository root is used, the same way it is in single-repo mode.

---

## How It Works

px0 already supports serving everything under a URL base path (`-base-path`). A multi-repository workspace gives each repository its own, unmodified px0 server mounted at `<base>/r/<name>/`. A small hub in front of those servers does three things:

- It redirects `<base>` to the first repository.
- It serves `<base>/api/workspace/list`, which the sidebar switcher uses.
- It routes every other request to the matching repository's server.

Because each server only knows about its own root, every existing feature works per repository with no changes to the request handlers. For implementation details, see [System Architecture](../internals/architecture.md#multi-repository-workspace-hub).

---

## Trade-offs & Limitations

- **Search is per-repository**: Quick open (`Cmd/Ctrl+P`), workspace search, and the file tree cover the repository you are viewing. They don't search across all repositories yet.
- **Switching reloads the page**: Open tabs and expanded folders are saved per repository and restored when you switch back.
- **Resources scale with repository count**: Each repository has its own index and git watcher. Language servers still start lazily, the first time you use them in a given repository.
- **No new dependencies and nothing extra on disk**: The feature is part of the same single static binary. The only file it writes is the per-repository session file, the same one a single-repo run of that directory would write.
