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
  for (let i = 0; i < 40; i++) c.setViewMode('dash', i % 2 ? 'list' : 'tiles');
  assert.equal(c.renders, 0); assert.equal(clock.jobs.size, 1);
  clock.flush();
  assert.equal(c.S.viewModes.dash, 'list'); assert.equal(c.renders, 1); assert.equal(c.focused, 1);
  assert.deepEqual(c.effects[0].slice(2), [1, 0]);
  c.setViewMode('dash', 'list'); clock.flush(); assert.equal(c.renders, 1);
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
  assert.match(html, /vm-indicator/); assert.match(html, /--vm-index:1/); assert.equal((html.match(/aria-pressed="true"/g) || []).length, 1);
});
function searchSetup() {
  const clock = timers(), requests = [];
  const c = vm.createContext({ ...clock, AbortController, URLSearchParams, URL,
    S: { view: 'plug', token: 'one', q: 'old', plugTab: 'browse' }, server: { id: 'one', node_id: 'local' }, grid: {setAttribute() {}, removeAttribute() {}},
    curServer: () => c.server, hasServerPermission: () => true, getViewMode: () => 'list',
    isSteelEngine: () => false, isModdedEngine: () => false, engineMeta: () => ({ key: 'paper' }),
    document: { addEventListener() {}, getElementById: () => c.grid },
    api(url, options) { const request = { ...deferred(), url, options }; requests.push(request); return request.promise; } });
  c.window = c;
  vm.runInContext(fs.readFileSync(path.join(__dirname, '../static/js/plugin-details.js'), 'utf8'), c);
  c.searchModrinth = c.NPPlugins.search;
  return { c, clock, requests };
}
test('search completion writes to the current layout grid, not a detached node', async () => {
  const { c, requests, clock } = searchSetup(); const oldGrid = c.grid;
  const pending = c.searchModrinth(); c.grid = {setAttribute() {}, removeAttribute() {}};
  requests[0].resolve({ hits: [{project_id:'result',title:'result'}], total_hits:1 }); await pending;
  assert.match(oldGrid.innerHTML, /Loading page 1/); assert.match(c.grid.innerHTML, /View result/); assert.equal(clock.jobs.size, 0);
});
test('older searches are aborted and cannot overwrite a newer query', async () => {
  const { c, requests } = searchSetup();
  const first = c.searchModrinth(); c.S.q = 'new'; const second = c.searchModrinth();
  assert.equal(requests[0].options.signal.aborted, true);
  requests[1].resolve({ hits: [{project_id:'new',title:'new'}], total_hits:1 }); await second;
  requests[0].resolve({ hits: [{project_id:'stale',title:'stale'}], total_hits:1 }); await first;
  assert.match(c.grid.innerHTML, /View new/); assert.equal(c.S.lastModrinthHits[0].project_id, 'new');
});
test('responses cannot populate another server/session or overwrite a pending typed query', async () => {
  for (const change of [c => { c.server = { id: 'two' }; }, c => { c.S.token = 'two'; }, c => { c.S.q = 'typing'; }]) {
    const { c, requests } = searchSetup(); const pending = c.searchModrinth(); change(c);
    requests[0].resolve({ hits: [{project_id:'stale',title:'stale'}], total_hits:1 }); await pending;
    assert.equal(c.S.lastModrinthHits.length, 0); assert.match(c.grid.innerHTML, /Loading page 1/);
  }
});
test('search deadline aborts the request and failure clears its timer', async () => {
  const { c, clock, requests } = searchSetup(); const pending = c.searchModrinth(); clock.flush();
  assert.equal(requests[0].options.signal.aborted, true);
  requests[0].reject(new Error('aborted')); await pending;
  assert.match(c.grid.innerHTML, /Retry/); assert.equal(clock.jobs.size, 0);
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
  c.setViewMode('files', 'tiles'); clock.flush();
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

test('every collection offers List and Tiles and migrates retired Icons preferences', () => {
  const scopes = new Map([...app.matchAll(/getViewMode\('([^']+)',\s*'(list|tiles)'\)/g)].map(m => [m[1], m[2]]));
  assert.ok(scopes.size >= 9, 'all collection scopes are covered');
  for (const [scope, fallback] of scopes) {
    const { c, clock, storage } = layoutSetup(scope, fallback);
    storage.set(`neopanel_vm_${scope}`, 'icons');
    assert.equal(c.getViewMode(scope, fallback), fallback, scope);
    assert.equal(storage.get(`neopanel_vm_${scope}`), fallback, `${scope} persisted migration`);
    c.S.viewModes[scope] = 'icons';
    assert.equal(c.getViewMode(scope, fallback), fallback, `${scope} cached migration`);
    const html = c.renderViewModeToggle(scope, fallback);
    assert.equal((html.match(/class="vm-btn /g) || []).length, 2, scope);
    assert.match(html, /List view/); assert.match(html, /Tiles view/);
    assert.doesNotMatch(html, /Icons|viewIcons|'icons'/);
    c.setViewMode(scope, 'icons'); clock.flush();
    assert.equal(c.renders, 0, `${scope} rejects retired layout`);
  }
});

test('view preferences preserve supported choices and tolerate blocked storage', () => {
  for (const mode of ['list', 'tiles']) {
    const { c, storage } = layoutSetup();
    storage.set('neopanel_vm_dash', mode);
    assert.equal(c.getViewMode('dash', 'tiles'), mode);
    c.S.viewModes.dash = mode === 'list' ? 'tiles' : 'list';
    assert.equal(c.getViewMode('dash', 'tiles'), c.S.viewModes.dash);
  }
  const { c } = layoutSetup();
  c.S.viewModes.dash = 'icons';
  c.localStorage = { getItem() { throw Error('blocked'); }, setItem() { throw Error('blocked'); } };
  assert.equal(c.getViewMode('dash', 'tiles'), 'tiles');
  assert.equal(c.S.viewModes.dash, 'tiles');
  assert.equal(c.getViewMode('unknown', 'icons'), 'list');
});

test('Chrome: two-option layout selectors migrate every collection and keep the indicator aligned', {skip:!process.env.NP_LAYOUT_PLAYWRIGHT,timeout:60000}, async () => {
  const { chromium } = require(process.env.NP_LAYOUT_PLAYWRIGHT);
  const browser = await chromium.launch({headless:true,executablePath:process.env.NP_LAYOUT_CHROME});
  const root = path.join(__dirname,'../static');
  const index = fs.readFileSync(path.join(root,'index.html'),'utf8');
  const css = [...index.matchAll(/href="\/css\/([^"]+)"/g)].map(m=>fs.readFileSync(path.join(root,'css',m[1]),'utf8')).join('\n');
  const scopes = [...new Map([...app.matchAll(/getViewMode\('([^']+)',\s*'(list|tiles)'\)/g)].map(m=>[m[1],m[2]]))];
  try {
    const page = await browser.newPage(), errors = [];
    page.on('pageerror',error=>errors.push(error.message));
    await page.route('https://neopanel-layout.test/',route=>route.fulfill({contentType:'text/html',body:`<!doctype html><html><head><style>${css}#main{position:static;margin:0;padding:20px;}#main section{margin-bottom:16px;}</style></head><body><main id="main"></main></body></html>`}));
    await page.goto('https://neopanel-layout.test/');
    await page.evaluate(scopes=>{
      window.S={view:'dash',token:'fixture',viewModes:{}};window.layoutScopes=scopes;
      window.esc=value=>String(value);window.ic=()=>'<svg width="13" height="13" aria-hidden="true"></svg>';
      window.$=selector=>document.querySelector(selector);
      window.R=()=>document.getElementById('main').innerHTML=layoutScopes.map(([scope,fallback])=>`<section>${renderViewModeToggle(scope,fallback)}</section>`).join('');
    },scopes);
    await page.addScriptTag({content:section('function getViewMode(', '// --- Clean Multi-Page')});
    for (const theme of ['dark','light']) for (const width of [1440,390,320]) {
      await page.setViewportSize({width,height:900});
      await page.evaluate(theme=>{
        document.documentElement.dataset.theme=theme;S.viewModes={};
        for(const [scope] of layoutScopes)localStorage.setItem(`neopanel_vm_${scope}`,'icons');
        R();
      },theme);
      assert.equal(await page.getByRole('button',{name:'Icons view'}).count(),0);
      assert.equal(await page.locator('.vm-btn').count(),scopes.length*2);
      for (const [scope,fallback] of scopes) {
        const toggle=page.locator(`.vm-toggle[data-scope="${scope}"]`);
        assert.equal(await toggle.locator('.vm-btn.on').textContent().then(value=>value.trim()),fallback==='list'?'List':'Tiles');
        assert.equal(await toggle.evaluate(el=>getComputedStyle(el).gridTemplateColumns.split(/\s+/).length),2);
        assert.equal(await page.evaluate(scope=>localStorage.getItem(`neopanel_vm_${scope}`),scope),fallback);
      }
      const files=page.locator('.vm-toggle[data-scope="files"]');
      await files.getByRole('button',{name:'Tiles',exact:true}).focus();await page.keyboard.press('Enter');
      await page.waitForFunction(()=>S.viewModes.files==='tiles');
      assert.equal(await files.getByRole('button',{name:'Tiles',exact:true}).evaluate(el=>el===document.activeElement),true);
      const aligned=await files.evaluate(el=>{const indicator=el.querySelector('.vm-indicator').getBoundingClientRect(),button=el.querySelector('.vm-btn.on').getBoundingClientRect();return Math.abs(indicator.left-button.left)<=2&&Math.abs(indicator.width-button.width)<=2;});
      assert.equal(aligned,true,`${theme}/${width} indicator`);
      assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1));
    }
    assert.deepEqual(errors,[]);
  } finally { await browser.close(); }
});
