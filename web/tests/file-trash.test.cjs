const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/file-trash.js'), 'utf8');
const app = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
const tick = () => new Promise(resolve => setImmediate(resolve));
function section(start, end) { const a = app.indexOf(start), b = app.indexOf(end, a); assert.ok(a >= 0 && b > a); return app.slice(a, b); }
function deferred() { let resolve, reject; const promise = new Promise((a, b) => { resolve = a; reject = b; }); return { promise, resolve, reject }; }
function setup() {
  const requests = [], messages = [], elements = {}, timers = new Map(); let timerID = 0;
  const c = vm.createContext({ AbortController, S: { user: { id: 1, role: 'user' }, token: 'one', view: 'fi', currentDir: '', selectedFiles: new Set() }, QOL: {}, server: { id: 'a', node_id: 'local' }, grants: ['files.view', 'files.edit', 'files.delete'],
    document: { getElementById: id => elements[id] || null }, hasServerPermission: p => c.grants.includes(p),
    curServer: () => c.server, sameQolServer: (a, b) => a?.id === b?.id && a?.node_id === b?.node_id,
    srvApiPath: (srv, suffix) => `/api/servers/${srv.id}${suffix}`, R() {}, ic: () => '',
    esc: s => String(s ?? '').replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('"', '&quot;'), formatFileSize: n => `${n} B`,
    toast: (...args) => messages.push(args), loadFileList: async () => {}, uiConfirm: async () => true, syncFileSelectionUI() {}, uiAlert: async () => {},
    setTimeout(fn) { timers.set(++timerID, fn); return timerID; }, clearTimeout(id) { timers.delete(id); },
    api: async (url, options) => { requests.push({ url, options }); return { items: [] }; } });
  c.window = c;
  vm.runInContext(source + '\nthis.state = FILE_TRASH;', c);
  return { c, requests, messages, elements, timers };
}
test('delete dialog defaults to trash and only an explicit checked choice is permanent', async () => {
  for (const permanent of [false, true]) {
    const { c, elements } = setup(), wait = deferred(); let change, modal;
    elements['file-delete-permanent'] = { checked: false, addEventListener(name, fn) { change = fn; } };
    elements['ui-modal-confirm-btn'] = {};
    c.uiConfirm = options => { modal = options; return wait.promise; };
    const result = c.confirmFileDeletion({ title: 'Delete?', description: '<p>File</p>' });
    assert.match(modal.message, /up to 24 hours/); assert.doesNotMatch(modal.message, /type="checkbox" checked/);
    if (permanent) { elements['file-delete-permanent'].checked = true; change(); assert.equal(elements['ui-modal-confirm-btn'].textContent, 'Delete permanently'); }
    wait.resolve(true); assert.equal((await result).permanent, permanent);
  }
});
test('cancelled delete dialog returns no action', async () => {
  const { c } = setup(); c.uiConfirm = async () => false;
  assert.equal(await c.confirmFileDeletion({ title: 'Delete?', description: '' }), null);
});
test('single deletion captures server before confirmation and sends explicit permanent choice', async () => {
  for (const permanent of [false, true]) {
    const { c, requests } = setup();
    vm.runInContext(section('async function deleteFileItem(', 'async function promptNewFile('), c);
    c.confirmFileDeletion = async () => ({ permanent });
    await c.deleteFileItem('config/a & b.txt');
    const deletes = requests.filter(r => r.options?.method === 'DELETE');
    assert.equal(deletes.length, 1); assert.match(deletes[0].url, new RegExp(`permanent=${permanent}$`));
    assert.match(deletes[0].url, /path=config%2Fa%20%26%20b.txt/); assert.equal(c.QOL.singleDelete, false);
  }
});
test('single delete cannot target a different server after confirmation or duplicate a pending dialog', async () => {
  const { c, requests } = setup(), wait = deferred();
  vm.runInContext(section('async function deleteFileItem(', 'async function promptNewFile('), c);
  c.confirmFileDeletion = () => wait.promise;
  const pending = c.deleteFileItem('a.txt'); await c.deleteFileItem('a.txt');
  c.server = { id: 'b', node_id: 'local' }; wait.resolve({ permanent: true }); await pending;
  assert.equal(requests.length, 0); assert.equal(c.QOL.singleDelete, false);
});
test('trash loading rejects stale results after server/session change and resets on access reset', async () => {
  const { c } = setup(), wait = deferred(); c.api = () => wait.promise;
  const pending = c.openFileTrash(); c.S.token = 'two'; wait.resolve({ items: [{ id: 'secret', name: 'old' }] }); await pending;
  assert.equal(c.state.items.length, 0); assert.doesNotMatch(c.renderFileTrash(), /secret|old/);
  c.resetFileTrash(); assert.equal(c.S.fileTrashOpen, false); assert.equal(c.state.key, '');
});
test('trash restore uses the owning server, never opts into overwrite, and keeps conflict errors visible', async () => {
  const { c, requests, messages } = setup();
  c.state.key = c.fileTrashIdentity(); c.state.items = [{ id: 'entry', name: 'config', original_path: 'config' }];
  c.api = async (url, options) => { requests.push({ url, options }); throw Error('Destination already exists'); };
  await c.restoreTrashItem('entry');
  assert.equal(requests[0].url, '/api/servers/a/files/trash/entry/restore');
  assert.equal(requests[0].options.method, 'POST'); assert.equal(requests[0].options.body, '{}');
  assert.match(messages[0][0], /already exists/); assert.equal(c.state.busy, false);
});
test('trash controls enforce permissions, escape paths and mark expired entries', () => {
  const { c } = setup(); c.state.key = c.fileTrashIdentity();
  c.state.items = [{ id: 'a" onclick="evil', name: '<img src=x>', original_path: '<folder>', expires_at: '2000-01-01T00:00:00Z' }];
  const html = c.renderFileTrash(); assert.match(html, /&lt;img/); assert.doesNotMatch(html, /<img|onclick="evil/); assert.match(html, /Expired/);
  assert.match(html, /onclick="restoreTrashItem\(this.dataset.trashId\)" disabled/);
  c.grants = ['files.view']; assert.match(c.renderFileTrash(), /access restricted/);
});
test('late directory responses cannot overwrite newer folder or clear its loading state', async () => {
  const { c } = setup(), first = deferred(), second = deferred(); const requests = [];
  vm.runInContext(section('let fileListSequence', 'function formatFileDate('), c);
  c.api = (url, options) => { requests.push({ url, options }); return requests.length === 1 ? first.promise : second.promise; };
  const a = c.loadFileList('a'), b = c.loadFileList('b');
  assert.equal(requests[0].options.signal.aborted, true);
  first.resolve([{ name: 'stale' }]); await a; assert.equal(c.S.loading.files, true);
  second.resolve([{ name: 'fresh' }]); await b;
  assert.equal(c.S.currentDir, 'b'); assert.equal(c.S.fileList[0].name, 'fresh'); assert.equal(c.S.loading.files, false);
});
test('directory failure keeps the current path and releases request/loading state', async () => {
  const { c, messages, timers } = setup();
  vm.runInContext(section('let fileListSequence', 'function formatFileDate('), c);
  c.S.currentDir = 'safe'; c.api = async () => { throw Error('No access'); };
  assert.equal(await c.loadFileList('blocked'), false);
  assert.equal(c.S.currentDir, 'safe'); assert.equal(c.S.loading.files, false); assert.equal(timers.size, 0);
  assert.equal(messages[0][0], 'No access');
});
test('root breadcrumb navigates to root, not the current directory again', () => {
  const { c } = setup(); let requested;
  vm.runInContext(section('function browseFolder(', 'function formatFileDate('), c);
  c.S.currentDir = 'world/region'; c.loadFileList = p => { requested = p; };
  c.browseFolder(''); assert.equal(requested, '');
  c.browseFolder('..'); assert.equal(requested, 'world');
});

test('cancelled hover cannot commit a folder and releases only its own loading state', async () => {
  const { c } = setup(), wait = deferred(); let current = true;
  vm.runInContext(section('let fileListSequence', 'function formatFileDate('), c);
  c.S.selectedFiles.add('keep.txt'); c.api = () => wait.promise;
  const pending = c.loadFileList('target', false, { isCurrent: () => current });
  current = false; wait.resolve([{ name: 'wrong folder' }]);
  assert.equal(await pending, false); assert.equal(c.S.currentDir, '');
  assert.equal(c.S.loading.files, false); assert.equal(c.S.selectedFiles.has('keep.txt'), true);
});

test('successful trash restore/purge sends one mutation and releases timers and busy state', async () => {
  for (const restore of [true, false]) {
    const { c, requests, timers } = setup();
    c.state.key = c.fileTrashIdentity(); c.state.items = [{ id: 'entry', name: 'file' }];
    await c.mutateTrashItem('entry', restore);
    assert.equal(requests.filter(r => ['POST', 'DELETE'].includes(r.options?.method)).length, 1);
    assert.equal(requests.length, 2); assert.equal(timers.size, 0); assert.equal(c.state.busy, false);
  }
});

test('purge confirmation cannot follow another server/session', async () => {
  const { c, requests } = setup(), wait = deferred();
  c.state.key = c.fileTrashIdentity(); c.state.items = [{ id: 'entry', name: 'file' }];
  c.uiConfirm = () => wait.promise;
  const pending = c.purgeTrashItem('entry'); c.S.token = 'different-session'; wait.resolve(true); await pending;
  assert.equal(requests.length, 0);
});

test('recoverable delete fails closed against a worker without Trash support', async () => {
  const { c, requests, messages } = setup();
  vm.runInContext(section('async function deleteFileItem(', 'async function promptNewFile('), c);
  c.confirmFileDeletion = async () => ({ permanent: false });
  c.api = async (url, options) => { requests.push({ url, options }); return {}; };
  await c.deleteFileItem('important.txt');
  assert.equal(requests.length, 1); assert.equal(requests[0].url, '/api/servers/a/files/trash');
  assert.match(messages[0][0], /does not support Trash/);
});
