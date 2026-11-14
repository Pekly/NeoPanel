const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const app = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
function section(start, end) {
  const a = app.indexOf(start), b = app.indexOf(end, a);
  assert.ok(a >= 0 && b > a, start);
  return app.slice(a, b);
}
function deferred() { let resolve, reject; const promise = new Promise((a, b) => { resolve = a; reject = b; }); return { promise, resolve, reject }; }
function timers() {
  let id = 0; const jobs = new Map();
  return { setTimeout(fn) { jobs.set(++id, fn); return id; }, clearTimeout(key) { jobs.delete(key); }, flush() { const next = [...jobs.values()]; jobs.clear(); next.forEach(fn => fn()); }, jobs };
}
function layoutSetup(scope = 'dash', fallback = 'tiles') {
  const clock = timers(), storage = new Map();
  const toggle = { dataset: { scope, fallback }, querySelector: () => ({ focus() { c.focused++; } }) };
  const c = vm.createContext({ ...clock, S: { view: scope === 'dash' ? 'dash' : 'plug', token: 'one', viewModes: {} },
    document: { querySelectorAll: () => [toggle], activeElement: { closest: () => toggle } },
    localStorage: { getItem: key => storage.get(key), setItem: (key, value) => storage.set(key, value) },
    $: () => ({}), R() { c.renders++; }, renders: 0, focused: 0, effects: [], esc: x => String(x), ic: () => '',
    NPMotion: { layout: (...args) => c.effects.push(args) } });
  c.window = c;
  vm.runInContext(section('function getViewMode(', '// --- Clean Multi-Page'), c);
  return { c, clock, storage };
}
test('layout bursts coalesce to the last choice, preserve focus, and never fetch', () => {
  const { c, clock } = layoutSetup();
  for (let i = 0; i < 40; i++) c.setViewMode('dash', i % 2 ? 'icons' : 'list');
  assert.equal(c.renders, 0); assert.equal(clock.jobs.size, 1);
  clock.flush();
  assert.equal(c.S.viewModes.dash, 'icons'); assert.equal(c.renders, 1); assert.equal(c.focused, 1);
  assert.deepEqual(c.effects[0].slice(2), [2, 1]);
  c.setViewMode('dash', 'icons'); clock.flush(); assert.equal(c.renders, 1);
  assert.doesNotMatch(section('function setViewMode(', 'function renderViewModeToggle('), /searchModrinth\(|\bapi\(|\bfetch\(/);
});
test('reselecting the current layout cancels a pending switch; invalid scopes do nothing', () => {
  const { c, clock } = layoutSetup();
  c.setViewMode('dash', 'list'); c.setViewMode('dash', 'tiles'); clock.flush();
  c.setViewMode('missing', 'list'); c.setViewMode('dash', 'invalid'); clock.flush();
  assert.equal(c.renders, 0);
});
test('pending layout switches cannot rerender another page, server, or session', () => {
  for (const field of ['view', 'activeServerId', 'activeNodeId', 'token', 'currentDir', 'plugTab']) {
    const { c, clock } = layoutSetup();
    c.setViewMode('dash', 'list'); c.S[field] = 'changed'; clock.flush();
    assert.equal(c.renders, 0, field); assert.equal(c.S.viewModes.dash, undefined);
  }
});
test('store layout switches are local even before the first search finishes', () => {
  const { c, clock } = layoutSetup('plug_store', 'list');
  c.S.lastModrinthHits = null; c.setViewMode('plug_store', 'tiles'); clock.flush();
  assert.equal(c.renders, 1); assert.equal(c.S.viewModes.plug_store, 'tiles');
  const html = c.renderViewModeToggle('plug_store');
  assert.match(html, /vm-indicator/); assert.match(html, /--vm-index:2/); assert.equal((html.match(/aria-pressed="true"/g) || []).length, 1);
});
function searchSetup() {
  const clock = timers(), requests = [];
  const c = vm.createContext({ ...clock, AbortController, URLSearchParams,
    S: { view: 'plug', token: 'one', q: 'old', plugTab: 'browse' }, server: { id: 'one', node_id: 'local' }, grid: {},
    curServer: () => c.server, sameQolServer: (a, b) => a?.id === b?.id && a?.node_id === b?.node_id,
    isSteelEngine: () => false, isModdedEngine: () => false, engineMeta: () => ({ key: 'paper' }),
    $: () => c.grid, loaderHtml: () => 'loading', renderModrinthHitsHtml: hits => hits.join(','), applyPermissionControls() {},
    api(url, options) { const request = { ...deferred(), url, options }; requests.push(request); return request.promise; } });
  vm.runInContext(section('let modrinthSearchSequence', 'function setPlugSkipConfirm('), c);
  return { c, clock, requests };
}
test('search completion writes to the current layout grid, not a detached node', async () => {
  const { c, requests, clock } = searchSetup(); const oldGrid = c.grid;
  const pending = c.searchModrinth(); c.grid = {};
  requests[0].resolve({ hits: ['result'] }); await pending;
  assert.equal(oldGrid.innerHTML, 'loading'); assert.equal(c.grid.innerHTML, 'result'); assert.equal(clock.jobs.size, 0);
});
test('older searches are aborted and cannot overwrite a newer query', async () => {
  const { c, requests } = searchSetup();
  const first = c.searchModrinth(); c.S.q = 'new'; const second = c.searchModrinth();
  assert.equal(requests[0].options.signal.aborted, true);
  requests[1].resolve({ hits: ['new'] }); await second;
  requests[0].resolve({ hits: ['stale'] }); await first;
  assert.equal(c.grid.innerHTML, 'new'); assert.equal(c.S.lastModrinthHits[0], 'new');
});
test('responses cannot populate another server/session or overwrite a pending typed query', async () => {
  for (const change of [c => { c.server = { id: 'two' }; }, c => { c.S.token = 'two'; }, c => { c.S.q = 'typing'; }]) {
    const { c, requests } = searchSetup(); const pending = c.searchModrinth(); change(c);
    requests[0].resolve({ hits: ['stale'] }); await pending;
    assert.equal(c.S.lastModrinthHits, null); assert.equal(c.grid.innerHTML, 'loading');
  }
});
test('search deadline aborts the request and failure clears its timer', async () => {
  const { c, clock, requests } = searchSetup(); const pending = c.searchModrinth(); clock.flush();
  assert.equal(requests[0].options.signal.aborted, true);
  requests[0].reject(new Error('aborted')); await pending;
  assert.match(c.grid.innerHTML, /timed out/); assert.equal(clock.jobs.size, 0);
});
test('dashboard refresh is single-flight, cooled down, and reports results only after completion', async () => {
  let now = 1000; const requests = [], messages = [], controls = [{ setAttribute() {}, removeAttribute() {} }];
  const c = vm.createContext({ Date: { now: () => now }, S: { token: 'one', serverCatalog: { clusterComplete: true } },
    document: { querySelectorAll: () => controls }, toast: (...args) => messages.push(args),
    loadDashboard() { const request = deferred(); requests.push(request); return request.promise; } });
  vm.runInContext(section('let dashboardRefreshBusy', 'let dashboardLoadSequence'), c);
  const first = c.refreshDashboard();
  for (let i = 0; i < 30; i++) await c.refreshDashboard();
  assert.equal(requests.length, 1); assert.equal(messages.length, 0);
  requests[0].resolve(); await first; assert.match(messages[0][0], /refreshed/);
  await c.refreshDashboard(); assert.equal(requests.length, 1);
  now += 1600; const next = c.refreshDashboard(); requests[1].reject(new Error('offline')); await next;
  assert.equal(messages.at(-1)[0], 'offline');
  now += 1600; const retry = c.refreshDashboard(); requests[2].resolve(); await retry;
  assert.equal(requests.length, 3);
});
function powerSetup() {
  const clock = timers(), requests = [], confirms = [];
  const c = vm.createContext({ ...clock, AbortController, S: { token: 'one', status: 'online' }, server: { id: 'one', name: 'First', node_id: 'local' },
    curServer: () => c.server, hasServerPermission: () => true, sameQolServer: (a, b) => a.id === b?.id && a.node_id === b?.node_id,
    esc: x => x, toast() {}, addLog() {}, updateInputState() {}, updateHeaderBadge() {}, liveCon() {}, uiAlert() {}, refreshServerStatus() {},
    normalizeServerStatus: x => x, srvApiPath: srv => srv.id, applyServerStatsSnapshot() { c.applied++; }, applied: 0,
    uiConfirm() { const request = deferred(); confirms.push(request); return request.promise; },
    api() { const request = deferred(); requests.push(request); return request.promise; } });
  vm.runInContext(section('const powerRequests', 'async function refreshServerStatus('), c);
  return { c, clock, requests, confirms };
}
test('power guard covers duplicate confirmations and duplicate POSTs', async () => {
  const { c, requests, confirms, clock } = powerSetup();
  const first = c.power('restart'); await c.power('restart'); assert.equal(confirms.length, 1);
  confirms[0].resolve(true); await Promise.resolve(); await Promise.resolve();
  await c.power('start'); assert.equal(requests.length, 1);
  requests[0].resolve({ current: {} }); await first;
  await c.power('start'); assert.equal(requests.length, 1);
  clock.flush(); const next = c.power('start'); assert.equal(requests.length, 2);
  requests[1].resolve({}); await next;
});
test('changing server during confirmation sends no power command', async () => {
  const { c, confirms, requests } = powerSetup(); const pending = c.power('stop');
  c.server = { id: 'two', node_id: 'local' }; confirms[0].resolve(true); await pending;
  assert.equal(requests.length, 0); assert.equal(c.S.status, 'online');
});
test('late power success/failure cannot change the newly selected server', async () => {
  for (const fail of [false, true]) {
    const { c, requests } = powerSetup(); const pending = c.power('start');
    c.server = { id: 'two', node_id: 'local' }; c.S.status = 'online';
    if (fail) requests[0].reject(new Error('offline')); else requests[0].resolve({ current: {} });
    await pending; assert.equal(c.S.status, 'online'); assert.equal(c.applied, 0);
  }
});
test('announcement close icon is vector-based, labelled, and text stays escaped', () => {
  const c = vm.createContext({ S: { community: { announcement: { enabled: true, text: '<img>', dismissible: true, id: 'test' } } },
    esc: s => s.replaceAll('<', '&lt;'), announcementDismissed: () => false });
  vm.runInContext(section('function renderAnnouncement()', 'function discordLogo()'), c);
  const html = c.renderAnnouncement();
  assert.match(html, /aria-label="Dismiss announcement"/); assert.match(html, /<svg/); assert.doesNotMatch(html, /&times;|<img>/);
});

test('power timeout releases its guard without silently retrying a command', async () => {
  const { c, clock } = powerSetup(); let calls = 0;
  c.isSteelEngine = () => false;
  c.serverExecutable = () => 'server.jar';
  c.api = (url, options) => { calls++; return new Promise((resolve, reject) => options.signal.addEventListener('abort', () => reject(new Error('timeout')))); };
  const first = c.power('start'); clock.flush(); await first;
  assert.equal(calls, 1);
  await c.power('start'); assert.equal(calls, 1);
  clock.flush(); const retry = c.power('start'); assert.equal(calls, 2);
  clock.flush(); await retry;
});

test('file layout changes retain the filter text and reapply filtering', () => {
  const { c, clock } = layoutSetup('files', 'list');
  c.S.view = 'fi'; const input = { value: 'world' }; let filtered;
  c.$ = selector => selector === '#fi-filter' ? input : {};
  c.R = () => { input.value = ''; };
  c.filterFileList = value => { filtered = value; };
  c.setViewMode('files', 'icons'); clock.flush();
  assert.equal(input.value, 'world'); assert.equal(filtered, 'world');
});

test('status reconciliation ignores a response after server navigation', async () => {
  const { c } = powerSetup(), request = deferred();
  c.api = () => request.promise;
  vm.runInContext(section('async function refreshServerStatus(', 'async function runCmd('), c);
  const pending = c.refreshServerStatus(); c.server = { id: 'two', node_id: 'local' };
  request.resolve({ stats: {}, version: 'stale' }); await pending;
  assert.equal(c.applied, 0); assert.equal(c.S.ver, undefined);
});
