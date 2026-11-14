// Stateful QoL workflows. Loaded before app.js; functions run after the app initializes.
const QOL = { editor: null, leaving: false, openSequence: 0, uploads: [], uploadID: 0, uploading: false };

// Stamp same-document entries so cancelled Back/Forward can restore position
// without overwriting either the destination URL or the forward history stack.
function setupQolHistory() {
  if (typeof history === 'undefined' || QOL.historyReady) return;
  QOL.historyReady = true;
  QOL.historyIndex = Number.isInteger(history.state?.npQolIndex) ? history.state.npQolIndex : 0;
  QOL.restoringHistory = null;
  const push = history.pushState.bind(history), replace = history.replaceState.bind(history);
  const stamp = (state, index) => ({ ...(state && typeof state === 'object' ? state : {}), npQolIndex: index });
  history.pushState = (state, title, url) => {
    const next = QOL.historyIndex + 1;
    push(stamp(state, next), title, url);
    QOL.historyIndex = next;
  };
  history.replaceState = (state, title, url) => replace(stamp(state, QOL.historyIndex), title, url);
  history.replaceState(history.state, '', location.href);
}
function restoreEditorHistory(targetIndex) {
  if (!Number.isInteger(targetIndex) || targetIndex === QOL.historyIndex) return;
  QOL.restoringHistory = QOL.historyIndex;
  history.go(QOL.historyIndex - targetIndex);
}
setupQolHistory();

function sameQolServer(a, b) {
  return Boolean(a && b && a.id === b.id && (a.node_id || 'local') === (b.node_id || 'local'));
}
function editorDraftKey(srv, path) {
  return 'neopanel:draft:' + JSON.stringify([S.user?.id || S.user?.username, srv.node_id || 'local', srv.id, path]);
}
function editorIsDirty() {
  const e = QOL.editor;
  return Boolean(e && e.content !== e.base);
}
function persistEditorDraft() {
  const e = QOL.editor;
  if (!e) return;
  try {
    if (editorIsDirty()) sessionStorage.setItem(e.key, JSON.stringify({ content: e.content, base: e.base }));
    else sessionStorage.removeItem(e.key);
    e.draftStored = true;
  } catch (_) {
    e.draftStored = false;
    if (!e.warned) { e.warned = true; toast('Browser draft storage is full or unavailable. Save this file before leaving.', 'bad'); }
  }
}
function captureEditor() {
  const ed = document.querySelector('#ed');
  if (ed && QOL.editor && ed.dataset.editorKey === QOL.editor.key) {
    QOL.editor.content = ed.value;
    QOL.editor.selection = [ed.selectionStart, ed.selectionEnd, ed.scrollTop];
    persistEditorDraft();
  }
}
function updateEditorStatus() {
  const e = QOL.editor;
  const status = document.querySelector('#editor-save-status');
  const button = document.querySelector('#editor-save-btn');
  if (status && e) status.textContent = e.saving ? 'Saving…' : editorIsDirty() ? (e.draftStored ? 'Unsaved · draft kept in this tab' : 'Unsaved · draft not stored') : 'Saved';
  if (button && e) button.disabled = Boolean(e.saving) || (typeof hasServerPermission === 'function' && !hasServerPermission('files.edit', e.server));
}
function bindQolEditor() {
  const ed = document.querySelector('#ed');
  const e = QOL.editor;
  if (!ed || !e || S.editingFile !== e.path || !sameQolServer(curServer(), e.server)) return;
  ed.dataset.editorKey = e.key;
  ed.value = e.content;
  if (e.selection) { ed.setSelectionRange(e.selection[0], e.selection[1]); ed.scrollTop = e.selection[2]; }
  ed.addEventListener('input', () => { captureEditor(); updateEditorStatus(); });
  ed.addEventListener('keydown', event => {
    if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 's') {
      event.preventDefault(); saveActiveFile();
    } else if (event.key === 'Tab') {
      event.preventDefault();
      ed.setRangeText('  ', ed.selectionStart, ed.selectionEnd, 'end');
      captureEditor(); updateEditorStatus();
    }
  });
  updateEditorStatus();
}
function editorLeaveChoice() {
  return new Promise(resolve => {
    if (_uiModalResolver) _uiModalResolver(null);
    _uiModalResolver = resolve;
    document.querySelector('#modal-container').innerHTML = `<div class="pal" onclick="if(event.target===this)closeUiModal('cancel')">
      <div class="pbox ui-modal-box" role="dialog" aria-modal="true" aria-labelledby="editor-leave-title">
        <h3 id="editor-leave-title">Unsaved changes</h3>
        <p>Save changes to <b>${esc(QOL.editor.path)}</b> before leaving?</p>
        <div class="row" style="justify-content:flex-end;gap:8px;flex-wrap:wrap">
          <button class="btn" id="editor-stay" onclick="closeUiModal('cancel')">Cancel</button>
          <button class="btn dg" onclick="closeUiModal('discard')">Discard</button>
          <button class="btn pr" onclick="closeUiModal('save')">Save &amp; leave</button>
        </div>
      </div></div>`;
    document.querySelector('#editor-stay')?.focus();
  });
}
async function guardEditorLeave() {
  captureEditor();
  if (QOL.editor?.saving || QOL.leaving) return false;
  if (!editorIsDirty()) return true;
  QOL.leaving = true;
  try {
    const choice = await editorLeaveChoice();
    if (choice === 'save') return await saveActiveFile() && !editorIsDirty();
    if (choice === 'discard') {
      QOL.editor.content = QOL.editor.base;
      persistEditorDraft();
      const ed = document.querySelector('#ed');
      if (ed) ed.value = QOL.editor.content;
      updateEditorStatus();
      return true;
    }
    return false;
  } finally { QOL.leaving = false; }
}
async function openQolFile(name) {
  if (!await guardEditorLeave()) return;
  const srv = curServer();
  if (!srv) return;
  if (typeof hasServerPermission === 'function' && !hasServerPermission('files.view', srv)) return;
  const token = S.token;
  const path = S.currentDir ? `${S.currentDir}/${name}` : name;
  const sequence = ++QOL.openSequence;
  try {
    const metadata = await api(srvApiPath(srv, `/files/preview?path=${encodeURIComponent(path)}`));
    if (sequence !== QOL.openSequence || S.view !== 'fi' || token !== S.token || !sameQolServer(srv, curServer())) return;
    if (!metadata || !['text', 'image', 'video', 'zip', 'binary'].includes(metadata.kind)) throw Error('File type could not be verified. Refresh the panel and check that the node is updated.');
    if (metadata.kind !== 'text' || metadata.editable !== true) return openFilePreview(srv, path, metadata);
    const data = await api(srvApiPath(srv, `/files/content?path=${encodeURIComponent(path)}`));
    if (sequence !== QOL.openSequence || S.view !== 'fi' || token !== S.token || !sameQolServer(srv, curServer())) return;
    if (typeof data?.content !== 'string' || /[\x00-\x08\x0b\x0e-\x1a\x1c-\x1f]/.test(data.content)) throw Error('This file is not safe to open as text. Download it instead.');
    const base = data.content;
    const e = { server: { ...srv }, path, base, content: base, key: editorDraftKey(srv, path), url: location.pathname + location.search };
    let draft;
    try { draft = JSON.parse(sessionStorage.getItem(e.key)); } catch (_) {}
    if (draft && typeof draft.content === 'string' && draft.content !== base) {
      const recover = await uiConfirm({ title: 'Recover file draft?', message: `An unsaved draft exists for <b>${esc(path)}</b>.${draft.base !== base ? ' The server copy has changed since this draft was created; review it before saving.' : ''}`, confirmText: 'Recover draft', cancelText: 'Use server copy' });
      if (sequence !== QOL.openSequence || S.view !== 'fi' || token !== S.token || !sameQolServer(srv, curServer())) return;
      if (recover) e.content = draft.content;
      else { try { sessionStorage.removeItem(e.key); } catch (_) {} }
    }
    if (typeof disposeFilePreview === 'function') disposeFilePreview();
    QOL.editor = e;
    S.editingFile = path;
    S.ftab = path;
    S.files[path] = base;
    persistEditorDraft();
    R();
  } catch (err) { if (sequence === QOL.openSequence && token === S.token && S.view === 'fi' && sameQolServer(srv, curServer())) toast(`Could not open ${path}: ${err.message}`, 'bad'); }
}
async function saveQolFile() {
  captureEditor();
  const e = QOL.editor;
  if (!e || e.saving) return false;
  const content = e.content;
  e.saving = true;
  updateEditorStatus();
  try {
    await api(srvApiPath(e.server, '/files/content'), { method: 'POST', body: JSON.stringify({ path: e.path, content }) });
    e.base = content;
    if (QOL.editor === e) {
      S.files[e.path] = content;
      persistEditorDraft();
    }
    toast(`Saved ${e.path.split('/').pop()}`, 'ok');
    return true;
  } catch (_) { return false; }
  finally { e.saving = false; updateEditorStatus(); }
}

function uploadQueueCounts() {
  const counts = { queued: 0, uploading: 0, succeeded: 0, failed: 0, cancelled: 0 };
  QOL.uploads.forEach(item => counts[item.status]++);
  return counts;
}
function renderUploadQueue() {
  const widget = document.querySelector('#upload-progress-widget');
  if (!widget) return;
  widget.style.display = QOL.uploads.length ? 'block' : 'none';
  if (!QOL.uploads.length) return;
  const c = uploadQueueCounts();
  widget.innerHTML = `<div class="row sb" style="gap:8px"><b>Upload queue</b><button class="btn sm" onclick="clearFinishedUploads()" title="Dismiss completed entries">Clear finished</button></div>
    <p class="mu" role="status">${c.succeeded} uploaded · ${c.failed} failed · ${c.cancelled} cancelled · ${c.queued + c.uploading} pending</p>
    <div class="upload-queue-items">${QOL.uploads.map(item => `<div class="upload-queue-item">
      <div class="row sb" style="gap:8px"><b style="overflow-wrap:anywhere">${esc(item.file.name)}</b><span>${item.status === 'uploading' ? item.progress + '%' : esc(item.status)}</span></div>
      <div class="mu" style="font-size:11px;overflow-wrap:anywhere">${esc(item.server.name)} / ${esc(item.dir || '(root)')}</div>
      ${item.status === 'uploading' ? `<progress max="100" value="${item.progress}" aria-label="Upload progress for ${esc(item.file.name)}"></progress>` : ''}
      ${item.error ? `<div class="upload-error">${esc(item.error)}</div>` : ''}
      ${['queued', 'uploading'].includes(item.status) ? `<button class="btn sm" onclick="cancelUpload(${item.id})">Cancel</button>` : ''}
      ${['failed', 'cancelled'].includes(item.status) ? `<button class="btn sm" onclick="retryUpload(${item.id})">Retry</button>` : ''}
    </div>`).join('')}</div>
    <div class="row" style="gap:8px;margin-top:10px">${c.failed ? '<button class="btn sm" onclick="retryFailedUploads()">Retry failed</button>' : ''}${c.queued + c.uploading ? '<button class="btn sm dg" onclick="cancelAllUploads()">Cancel pending</button>' : ''}</div>`;
}
function enqueueUploads(files) {
  if (typeof hasServerPermission === 'function' && (!hasServerPermission('files.view') || !hasServerPermission('files.edit'))) { toast('Access restricted: files.view and files.edit required', 'bad'); return; }
  const srv = curServer();
  if (!srv || !files?.length) return;
  const dir = S.currentDir || '';
  for (const file of files) QOL.uploads.push({ id: ++QOL.uploadID, file, dir, server: { ...srv }, user: S.user?.id || S.user?.username, status: 'queued', progress: 0, error: '' });
  renderUploadQueue();
  return processUploadQueue();
}
function cancelUpload(id) {
  const item = QOL.uploads.find(i => i.id === id);
  if (!item || !['queued', 'uploading'].includes(item.status)) return;
  item.status = 'cancelled';
  item.error = 'Transfer cancelled. If the server already received the file, it may still exist; check before retrying.';
  item.xhr?.abort();
  renderUploadQueue();
}
function cancelAllUploads() { QOL.uploads.forEach(item => cancelUpload(item.id)); }
function retryUpload(id) {
  const item = QOL.uploads.find(i => i.id === id);
  if (!item || !['failed', 'cancelled'].includes(item.status) || item.user !== (S.user?.id || S.user?.username)) return;
  item.status = 'queued'; item.error = ''; item.progress = 0;
  renderUploadQueue(); processUploadQueue();
}
function retryFailedUploads() {
  for (const item of QOL.uploads) if (item.status === 'failed') retryUpload(item.id);
}
function clearFinishedUploads() {
  QOL.uploads = QOL.uploads.filter(item => ['queued', 'uploading'].includes(item.status));
  renderUploadQueue();
}
function transferUpload(item) {
  return new Promise(resolve => {
    const xhr = new XMLHttpRequest();
    item.xhr = xhr;
    let finished = false;
    const finish = (status, error = '') => {
      if (finished) return;
      finished = true;
      if (item.status !== 'cancelled') { item.status = status; item.error = error; }
      item.xhr = null;
      renderUploadQueue(); resolve();
    };
    xhr.upload.onprogress = event => {
      if (event.lengthComputable) { item.progress = Math.round(event.loaded / event.total * 100); renderUploadQueue(); }
    };
    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) { item.progress = 100; finish('succeeded'); }
      else {
        let error = `Upload failed (HTTP ${xhr.status})`;
        try { error = JSON.parse(xhr.responseText).error || error; } catch (_) {}
        finish('failed', error);
      }
    };
    xhr.onerror = () => finish('failed', 'Network error. Check your connection and retry.');
    xhr.ontimeout = () => finish('failed', 'Upload timed out after 15 minutes. Check the destination before retrying.');
    xhr.onabort = () => finish('cancelled');
    try {
      if (typeof assertClientRequestAccess === 'function') assertClientRequestAccess(srvApiPath(item.server, '/files/upload'), { method: 'POST' });
      const body = new FormData(); body.append('file', item.file); body.append('path', item.dir);
      xhr.open('POST', srvApiPath(item.server, '/files/upload'));
      xhr.timeout = 15 * 60 * 1000;
      xhr.setRequestHeader('Authorization', `Bearer ${S.token}`);
      xhr.send(body);
    } catch (err) { finish('failed', err.message); }
  });
}
async function processUploadQueue() {
  if (QOL.uploading) return;
  QOL.uploading = true;
  try {
    let item;
    while ((item = QOL.uploads.find(i => i.status === 'queued'))) {
      item.status = 'uploading'; renderUploadQueue();
      await transferUpload(item);
      if (S.view === 'fi' && sameQolServer(item.server, curServer()) && S.currentDir === item.dir) await loadFileList(item.dir);
    }
  } finally { QOL.uploading = false; renderUploadQueue(); }
}

function cronFieldValues(field, min, max) {
  const values = new Set();
  for (const part of field.split(',')) {
    if (!/^(\*|\d+(?:-\d+)?)(?:\/\d+)?$/.test(part)) throw new Error('Use numbers, *, ranges, lists, and /steps in each of the five cron fields.');
    const [range, stepText] = part.split('/');
    const step = stepText === undefined ? 1 : Number(stepText);
    let start, end;
    if (range === '*') { start = min; end = max; }
    else { const pair = range.split('-').map(Number); start = pair[0]; end = pair.length > 1 ? pair[1] : stepText ? max : start; }
    if (step < 1 || step > max - min + 1 || start < min || end > max || start > end) throw new Error(`Cron value outside ${min}–${max}, reversed range, or invalid step.`);
    for (let v = start; v <= end; v += step) values.add(v);
  }
  return values;
}
function nextCronPreview(expression, now = new Date()) {
  const fields = expression.trim().split(/\s+/);
  if (fields.length !== 5) throw new Error('A cron expression needs exactly five fields.');
  const months = ['JAN','FEB','MAR','APR','MAY','JUN','JUL','AUG','SEP','OCT','NOV','DEC'];
  const weekdays = ['SUN','MON','TUE','WED','THU','FRI','SAT'];
  fields[3] = fields[3].replace(/[a-z]+/gi, name => { const i = months.indexOf(name.toUpperCase()); return i < 0 ? name : String(i + 1); });
  fields[4] = fields[4].replace(/[a-z]+/gi, name => { const i = weekdays.indexOf(name.toUpperCase()); return i < 0 ? name : String(i); });
  const ranges = [[0, 59], [0, 23], [1, 31], [1, 12], [0, 7]];
  const sets = fields.map((field, i) => cronFieldValues(field, ...ranges[i]));
  if (sets[4].has(7)) sets[4].add(0);
  const hours = [...sets[1]].sort((a, b) => a - b), minutes = [...sets[0]].sort((a, b) => a - b);
  const day = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate()));
  for (let n = 0; n < 366 * 8; n++, day.setUTCDate(day.getUTCDate() + 1)) {
    if (!sets[3].has(day.getUTCMonth() + 1)) continue;
    const dom = sets[2].has(day.getUTCDate()), dow = sets[4].has(day.getUTCDay());
    const match = fields[2].startsWith('*') || fields[4].startsWith('*') ? dom && dow : dom || dow;
    if (!match) continue;
    for (const hour of hours) for (const minute of minutes) {
      const candidate = new Date(day); candidate.setUTCHours(hour, minute, 0, 0);
      if (candidate > now) return candidate;
    }
  }
  throw new Error('This schedule has no execution in the next eight years.');
}
function formatScheduleUTC(value) {
  if (!value || String(value).startsWith('0001')) return '—';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? '—' : date.toISOString().replace('T', ' ').replace(/\.\d{3}Z$/, ' UTC');
}
function updateSchedulePreview() {
  const input = document.querySelector('#sch-cron');
  const preview = document.querySelector('#sch-preview');
  if (!input || !preview) return;
  try { preview.textContent = 'Next run: ' + formatScheduleUTC(nextCronPreview(input.value)); input.setCustomValidity(''); }
  catch (err) { preview.textContent = err.message; input.setCustomValidity(err.message); }
}
function applyScheduleTimePicker() {
  const mode = document.querySelector('#sch-frequency')?.value;
  const time = (document.querySelector('#sch-time')?.value || '04:00').split(':').map(Number);
  const day = document.querySelector('#sch-weekday')?.value || '0';
  const input = document.querySelector('#sch-cron');
  if (!input || mode === 'custom') return;
  input.value = mode === 'hourly' ? `${time[1]} * * * *` : `${time[1]} ${time[0]} * * ${mode === 'weekly' ? day : '*'}`;
  updateSchedulePreview();
}

window.addEventListener('beforeunload', event => {
  if (!window.S) return;
  captureEditor();
  if (editorIsDirty() || QOL.uploads.some(i => ['queued', 'uploading'].includes(i.status))) { event.preventDefault(); event.returnValue = ''; }
});
