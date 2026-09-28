// web/src/workspace.js
// Repo switcher for multi-repository workspaces (px0-ai/px0#162).
// Only appears when px0 was started with more than one directory; the hub
// serves /api/workspace/list at the global root, outside this repo's base path.
import { $, esc } from './state.js';

export async function initWorkspace() {
  // This page lives at <base>/r/<name>/; the hub list sits at <base>/api/...
  const here = new URL(document.baseURI || location.href);
  const m = here.pathname.match(/^(.*\/)r\/[^/]+\/$/);
  if (!m) return; // single-repo session
  let repos;
  try {
    const r = await fetch(m[1] + 'api/workspace/list');
    if (!r.ok) return;
    repos = (await r.json()).repos || [];
  } catch { return; }
  if (repos.length < 2) return;

  const sel = document.createElement('select');
  sel.id = 'repo-switch';
  sel.className = 'repo-switch';
  sel.title = 'Switch repository';
  sel.innerHTML = repos.map(r =>
    `<option value="${esc(r.path)}"${r.path === here.pathname ? ' selected' : ''}>` +
    `${esc(r.name)}${r.gitChanges ? ` (${r.gitChanges})` : ''}</option>`).join('');
  sel.addEventListener('change', () => { location.href = sel.value; });

  const name = $('#root-name');
  if (name) { name.hidden = true; name.after(sel); }
}
