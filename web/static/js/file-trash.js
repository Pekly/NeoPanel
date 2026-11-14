// Server-scoped trash UI. Payloads stay on the owning node; no deleted content is cached here.
const FILE_TRASH = { key: '', items: [], loading: false, error: '', sequence: 0, busy: false, controller: null };
function fileTrashIdentity(server = curServer()) {
  return JSON.stringify([S.token, S.user?.id, server?.node_id || 'local', server?.id]);
}
function canUseFileTrash() { return hasServerPermission('files.view') && hasServerPermission('files.delete'); }
function resetFileTrash() {
  FILE_TRASH.controller?.abort();
  Object.assign(FILE_TRASH, { key: '', items: [], loading: false, error: '', busy: false, controller: null, sequence: FILE_TRASH.sequence + 1 });
  S.fileTrashOpen = false;
}
async function openFileTrash() {
  if (!curServer() || !canUseFileTrash()) return;
  window.NPFileDrag?.cancel();
  S.fileTrashOpen = true;
  await loadFileTrash();
}
function closeFileTrash() {
  S.fileTrashOpen = false;
  loadFileList(S.currentDir || '');
  R();
}
async function loadFileTrash() {
  const server = curServer();
  if (!server || !canUseFileTrash()) return;
  const key = fileTrashIdentity(server);
  FILE_TRASH.controller?.abort();
  const controller = new AbortController(), sequence = ++FILE_TRASH.sequence;
  const timer = setTimeout(() => controller.abort(), 20000);
  Object.assign(FILE_TRASH, { key, items: [], loading: true, error: '', controller });
  const live = () => sequence === FILE_TRASH.sequence && key === fileTrashIdentity();
  if (S.fileTrashOpen) R();
  try {
    const result = await api(srvApiPath(server, '/files/trash'), { signal: controller.signal, suppressToast: true });
    if (!live()) return;
    if (!Array.isArray(result.items)) throw Error('Invalid trash response. Update the panel and worker together.');
    FILE_TRASH.items = result.items;
  } catch (error) {
    if (live()) FILE_TRASH.error = error.name === 'AbortError' ? 'Trash request timed out. Try refreshing.' : error.message;
  } finally {
    clearTimeout(timer);
    if (live()) {
      FILE_TRASH.loading = false; FILE_TRASH.controller = null;
      if (S.fileTrashOpen && S.view === 'fi') R();
    }
  }
}
function trashTimeLeft(value) {
  const remaining = Date.parse(value) - Date.now();
  if (!Number.isFinite(remaining) || remaining <= 0) return 'Expired';
  if (remaining < 60000) return 'Expires in less than a minute';
  if (remaining < 3600000) return `Expires in ${Math.ceil(remaining / 60000)} min`;
  return `Expires in ${Math.ceil(remaining / 3600000)} h`;
}
function renderFileTrash() {
  if (!canUseFileTrash()) return '<section class="card"><h2>Trash access restricted</h2><p>Browse Files and Delete Files permissions are required.</p><button class="btn" onclick="closeFileTrash()">Back to files</button></section>';
  const current = FILE_TRASH.key === fileTrashIdentity();
  const state = current ? FILE_TRASH : { items: [], loading: false, busy: false, error: 'Refresh to load this server’s trash.' };
  return `<section class="card file-trash-workspace">
    <div class="row sb" style="gap:12px;flex-wrap:wrap;margin-bottom:12px"><div class="row" style="gap:12px"><button class="btn sm" onclick="closeFileTrash()">Back to files</button><h2>${ic('trash', 18)} Trash</h2></div><button class="btn sm" onclick="loadFileTrash()" ${state.loading || state.busy ? 'disabled' : ''}>${ic('refresh', 14)} Refresh</button></div>
    <p class="mu" style="margin-bottom:16px">Deleted files are kept for up to 24 hours, then automatically removed. Restore returns an item to its original path and never replaces an existing item.</p>
    ${state.loading ? '<p role="status">Loading trash…</p>' : state.error ? `<p role="alert">${esc(state.error)}</p>` : !state.items.length ? '<div class="file-preview-empty" role="status">Trash is empty.</div>' : `<div class="file-trash-list">${state.items.map(item => {
      const expired = !(Date.parse(item.expires_at) > Date.now());
      return `<article class="file-trash-item"><div><b>${ic(item.is_dir ? 'folder' : 'file', 15)} ${esc(item.name || item.original_path)}</b><div class="mu">${esc(item.original_path)}</div><small class="mu" title="${esc(item.expires_at)}">${esc(trashTimeLeft(item.expires_at))}${item.is_dir ? ' · Folder' : ` · ${esc(formatFileSize(item.size || 0))}`}</small></div><div class="row">
        <button class="btn sm" data-trash-id="${esc(item.id)}" onclick="restoreTrashItem(this.dataset.trashId)" ${state.busy || expired || !hasServerPermission('files.edit') ? 'disabled' : ''}>${ic('refresh', 13)} Restore</button>
        <button class="btn sm dg" data-trash-id="${esc(item.id)}" onclick="purgeTrashItem(this.dataset.trashId)" ${state.busy ? 'disabled' : ''}>${ic('trash', 13)} Delete permanently</button>
      </div></article>`;
    }).join('')}</div>`}
  </section>`;
}
async function mutateTrashItem(id, restore) {
  if (FILE_TRASH.busy || !canUseFileTrash() || (restore && !hasServerPermission('files.edit'))) return;
  const server = curServer(), key = fileTrashIdentity();
  const item = FILE_TRASH.key === key && FILE_TRASH.items.find(item => item.id === id);
  let timeout = null;
  if (!server || !item) return;
  FILE_TRASH.busy = true;
  try {
    if (!restore && !await uiConfirm({ title: 'Delete permanently?', message: `Permanently remove <b>${esc(item.name || item.original_path)}</b> from Trash? This cannot be undone.`, confirmText: 'Delete permanently', isDanger: true })) return;
    if (key !== fileTrashIdentity() || !canUseFileTrash() || (restore && !hasServerPermission('files.edit'))) return;
    if (S.fileTrashOpen) R();
    const controller = new AbortController();
    timeout = setTimeout(() => controller.abort(), 45000);
    await api(srvApiPath(server, `/files/trash/${encodeURIComponent(id)}${restore ? '/restore' : ''}`), { method: restore ? 'POST' : 'DELETE', ...(restore ? { body: '{}' } : {}), signal: controller.signal, suppressToast: true });
    if (key !== fileTrashIdentity()) return;
    toast(restore ? 'Item restored to its original location' : 'Item permanently deleted', 'ok');
    await loadFileTrash();
  } catch (error) {
    if (key === fileTrashIdentity()) toast(error.message || 'Trash operation failed', 'bad');
  } finally {
    clearTimeout(timeout);
    if (key === fileTrashIdentity()) { FILE_TRASH.busy = false; if (S.fileTrashOpen && S.view === 'fi') R(); }
  }
}
function restoreTrashItem(id) { return mutateTrashItem(id, true); }
function purgeTrashItem(id) { return mutateTrashItem(id, false); }

async function assertTrashAvailable(server) {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), 15000);
  try {
    const result = await api(srvApiPath(server, '/files/trash'), { signal: controller.signal, suppressToast: true });
    if (!Array.isArray(result?.items)) throw Error('This worker does not support Trash. Update the panel and worker before deleting.');
  } finally { clearTimeout(timer); }
}

async function confirmFileDeletion({ title, description }) {
  let permanent = false;
  const pending = uiConfirm({ title, message: `${description}<p class="mu">Items go to Trash for up to 24 hours and can be restored during that time.</p><label class="file-delete-choice"><input id="file-delete-permanent" type="checkbox"><span>Delete permanently instead (skip Trash)<br><small class="mu">Permanent deletion cannot be undone.</small></span></label>`, confirmText: 'Move to Trash', cancelText: 'Cancel', isDanger: true });
  const checkbox = document.getElementById('file-delete-permanent');
  checkbox?.addEventListener('change', () => {
    permanent = checkbox.checked;
    const confirm = document.getElementById('ui-modal-confirm-btn');
    if (confirm) confirm.textContent = permanent ? 'Delete permanently' : 'Move to Trash';
  });
  return await pending ? { permanent } : null;
}
