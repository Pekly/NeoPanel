const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const path = require('node:path');
const source = fs.readFileSync(path.join(__dirname, '../static/js/qol.js'), 'utf8');
const app = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
const tick = () => new Promise(resolve => setImmediate(resolve));

function setup() {
  const elements = {}, storage = new Map(), requests = [], listeners = {};
  const server = { id: 'a', node_id: 'local', name: 'Alpha' };
  class XHR {
    constructor() { this.upload = {}; requests.push(this); }
    open(method, url) { this.url = url; }
    setRequestHeader() {}
    send(body) { this.body = body; }
    abort() { this.onabort(); }
    finish(status, error) { this.status = status; this.responseText = JSON.stringify({ error }); this.onload(); }
  }
  const c = vm.createContext({
    console, Date, setTimeout, clearTimeout,
    S: { user: { id: 'u1' }, token: 'token', view: 'fi', currentDir: 'plugins', files: {}, servers: [server] },
    document: { querySelector: s => elements[s] || null, body: { contains: () => true } },
    sessionStorage: { getItem: k => storage.get(k) || null, setItem: (k, v) => storage.set(k, v), removeItem: k => storage.delete(k) },
    location: { pathname: '/server/a/files', search: '' },
    esc: s => String(s ?? '').replaceAll('<', '&lt;'), toast: () => {}, R: () => {},
    api: async () => ({}), loadFileList: async () => {},
    srvApiPath: (srv, suffix) => `/nodes/${srv.node_id || 'local'}/servers/${srv.id}${suffix}`,
    XMLHttpRequest: XHR, FormData: class { constructor() { this.values = {}; } append(k, v) { this.values[k] = v; } },
    _uiModalResolver: null, uiConfirm: async () => true
  });
  c.window = c; c.addEventListener = (name, fn) => { listeners[name] = fn; };
  c.curServer = () => c.S.servers[0];
  vm.runInContext(source + '\nthis.qol = QOL;', c);
  c.saveActiveFile = c.saveQolFile;
  return { c, elements, storage, requests, listeners, server };
}
function editState(c, content = 'edited') {
  c.qol.editor = { server: c.curServer(), path: 'server.properties', key: 'draft', base: 'original', content };
}

test('cron preview supports ranges/steps, weekday lists, leap day and UTC boundaries', () => {
  const { c } = setup();
  for (const [expr, now, expected] of [
    ['0-59/5 * * * *', '2026-03-01T00:01:00Z', '2026-03-01T00:05:00.000Z'],
    ['0 4 * * *', '2026-03-01T04:00:00Z', '2026-03-02T04:00:00.000Z'],
    ['0 9 * * 1-5', '2026-03-06T09:00:00Z', '2026-03-09T09:00:00.000Z'],
    ['0 0 29 2 *', '2025-03-01T00:00:00Z', '2028-02-29T00:00:00.000Z'],
    ['0 0 * * 7', '2026-03-01T00:00:00Z', '2026-03-08T00:00:00.000Z'],
  ]) assert.equal(c.nextCronPreview(expr, new Date(now)).toISOString(), expected);
});
test('invalid cron expressions cannot be silently accepted by the form', () => {
  const { c } = setup();
  for (const expr of ['x * * * *', '60 * * * *', '*/0 * * * *', '4-2 * * * *', '* * *', '0 0 31 2 *']) {
    assert.throws(() => c.nextCronPreview(expr));
  }
});
test('draft keys separate users, nodes, servers, and paths', () => {
  const { c, server } = setup();
  const a = c.editorDraftKey(server, 'config');
  assert.notEqual(a, c.editorDraftKey({ ...server, node_id: 'remote' }, 'config'));
  assert.notEqual(a, c.editorDraftKey(server, 'other'));
  c.S.user.id = 'u2'; assert.notEqual(a, c.editorDraftKey(server, 'config'));
});
test('draft is retained through failed save and cleared only after successful save', async () => {
  const { c, storage } = setup(); editState(c);
  c.persistEditorDraft(); assert.ok(storage.has('draft'));
  c.api = async () => { throw Error('disk full'); };
  assert.equal(await c.saveQolFile(), false); assert.equal(c.editorIsDirty(), true); assert.ok(storage.has('draft'));
  c.api = async () => ({});
  assert.equal(await c.saveQolFile(), true); assert.equal(c.editorIsDirty(), false); assert.equal(storage.has('draft'), false);
});
test('new edits during save remain dirty and recoverable', async () => {
  const { c, storage } = setup(); editState(c);
  let finish; c.api = () => new Promise(resolve => { finish = resolve; });
  const saving = c.saveQolFile(); c.qol.editor.content = 'newer edit'; finish({}); await saving;
  assert.equal(c.qol.editor.base, 'edited'); assert.equal(c.editorIsDirty(), true);
  assert.equal(JSON.parse(storage.get('draft')).content, 'newer edit');
});
test('navigation can cancel, discard, or save without silent data loss', async () => {
  const { c, storage } = setup(); editState(c);
  c.editorLeaveChoice = async () => 'cancel'; assert.equal(await c.guardEditorLeave(), false); assert.equal(c.qol.editor.content, 'edited');
  c.editorLeaveChoice = async () => 'discard'; assert.equal(await c.guardEditorLeave(), true); assert.equal(c.qol.editor.content, 'original'); assert.equal(storage.size, 0);
  editState(c); c.editorLeaveChoice = async () => 'save'; c.api = async () => { throw Error('offline'); };
  assert.equal(await c.guardEditorLeave(), false); assert.equal(c.editorIsDirty(), true);
});
test('opening a file offers recovery and ignores responses after server switching', async () => {
  const { c, storage, server } = setup();
  const key = c.editorDraftKey(server, 'plugins/config.yml');
  storage.set(key, JSON.stringify({ content: 'draft content', base: 'server content' }));
  c.api = async () => ({ content: 'server content' });
  await c.openQolFile('config.yml'); assert.equal(c.qol.editor.content, 'draft content');
  c.qol.editor = null;
  let finish; c.api = () => new Promise(resolve => { finish = resolve; });
  const opening = c.openQolFile('other.yml'); await tick();
  c.S.servers = [{ id: 'b', node_id: 'remote' }]; finish({ content: 'wrong server' }); await opening;
  assert.equal(c.qol.editor, null);
});
test('beforeunload warns for dirty files and pending uploads', () => {
  const { c, listeners } = setup(); editState(c);
  let warned = false; listeners.beforeunload({ preventDefault() { warned = true; } }); assert.equal(warned, true);
  c.qol.editor = null; c.qol.uploads.push({ status: 'queued' }); warned = false;
  listeners.beforeunload({ preventDefault() { warned = true; } }); assert.equal(warned, true);
});
test('uploads preserve original server/directory, report API error, and retry', async () => {
  const { c, requests } = setup();
  const uploading = c.enqueueUploads([{ name: 'a.jar' }, { name: 'b.jar' }]);
  c.S.currentDir = 'world'; c.S.servers = [{ id: 'b', node_id: 'remote', name: 'Beta' }];
  requests[0].finish(403, 'Permission denied'); await tick();
  assert.equal(c.qol.uploads[0].status, 'failed'); assert.equal(c.qol.uploads[0].error, 'Permission denied');
  assert.equal(requests[1].url, '/nodes/local/servers/a/files/upload'); assert.equal(requests[1].body.values.path, 'plugins');
  requests[1].finish(200); await uploading;
  c.retryFailedUploads(); assert.equal(requests[2].url, '/nodes/local/servers/a/files/upload');
  requests[2].finish(200); await tick(); assert.equal(c.uploadQueueCounts().succeeded, 2);
});
test('cancelling queued/active uploads settles worker and does not upload cancelled files', async () => {
  const { c, requests } = setup();
  const uploading = c.enqueueUploads([{ name: 'a.jar' }, { name: 'b.jar' }]);
  c.cancelAllUploads(); await uploading;
  assert.equal(requests.length, 1); assert.equal(c.uploadQueueCounts().cancelled, 2); assert.equal(c.qol.uploading, false);
  c.retryUpload(2); requests[1].finish(200); await tick(); assert.equal(c.qol.uploads[1].status, 'succeeded');
});
test('plugin toggle calls API and rolls back on failure', async () => {
  const { c } = setup();
  const start = app.indexOf('async function togglePlugin('), end = app.indexOf('\nlet searchTimeout', start);
  vm.runInContext(app.slice(start, end), c);
  c.S.plugins = [['example.jar', 'installed', true, '', false]];
  let request;
  c.api = async (url, opts) => { request = { url, body: JSON.parse(opts.body) }; return { filename: 'example.jar.disabled', enabled: false }; };
  await c.togglePlugin(0, false);
  assert.equal(request.url, '/nodes/local/servers/a/plugins/toggle'); assert.equal(request.body.filename, 'example.jar'); assert.equal(c.S.plugins[0][2], false);
  c.api = async () => { throw Error('permission denied'); };
  await c.togglePlugin(0, true); assert.equal(c.S.plugins[0][2], false); assert.equal(c.S.plugins[0][4], false);
});

test('cron preview matches backend aliases and step bounds', () => {
  const { c } = setup();
  assert.equal(c.nextCronPreview('0 9 * MAR MON-FRI', new Date('2026-03-06T09:00:00Z')).toISOString(), '2026-03-09T09:00:00.000Z');
  assert.throws(() => c.nextCronPreview('*/61 * * * *'));
});
test('schedule UI accepts the backend response envelope and legacy arrays', async () => {
  const { c } = setup();
  const start = app.indexOf('async function loadServerSchedules('), end = app.indexOf('\nfunction describeCron', start);
  vm.runInContext(app.slice(start, end), c);
  const schedule = { id: 'one', last_status: 'failed', last_error: 'Server stopped' };
  c.api = async () => ({ schedules: [schedule] });
  await c.loadServerSchedules(); assert.equal(c.S.serverSchedules[0].last_status, 'failed');
  c.api = async () => [schedule];
  await c.loadServerSchedules(); assert.equal(c.S.serverSchedules[0].id, 'one');
});

test('cancelled Back and Forward restore position without replacing history entries', () => {
  const { c } = setup();
  const entries = [{ state: { other: 'preserved' }, url: '/start' }];
  let position = 0, replacements = 0;
  c.location.href = '/start';
  c.history = {
    get state() { return entries[position].state; },
    pushState(state, title, url) { entries.splice(position + 1); entries.push({ state, url }); position++; },
    replaceState(state, title, url) { replacements++; entries[position] = { state, url }; },
    go(delta) { position += delta; }
  };
  c.setupQolHistory();
  c.history.pushState(null, '', '/files'); c.history.pushState(null, '', '/schedules');
  const before = JSON.stringify(entries), replaceCount = replacements;
  position = 1;
  c.restoreEditorHistory(c.history.state.npQolIndex);
  assert.equal(position, 2); assert.equal(JSON.stringify(entries), before); assert.equal(replacements, replaceCount);
  c.qol.historyIndex = 1; position = 2;
  c.restoreEditorHistory(c.history.state.npQolIndex);
  assert.equal(position, 1); assert.equal(JSON.stringify(entries), before);
  assert.equal(entries[0].state.other, 'preserved');
});
test('late failed schedule requests cannot clear another server list or loading state', async () => {
  const { c } = setup();
  const start = app.indexOf('async function loadServerSchedules('), end = app.indexOf('\nfunction describeCron', start);
  vm.runInContext(app.slice(start, end), c);
  let reject;
  c.api = () => new Promise((resolve, fail) => { reject = fail; });
  const loadingA = c.loadServerSchedules();
  c.S.servers = [{ id: 'b', node_id: 'remote' }];
  c.S.serverSchedules = [{ id: 'b-schedule' }];
  c.S.loading.schedules = true;
  reject(Error('Server A went offline'));
  await loadingA;
  assert.equal(c.S.serverSchedules[0].id, 'b-schedule');
  assert.equal(c.S.loading.schedules, true);
});
