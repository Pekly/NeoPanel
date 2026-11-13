const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');

function slice(start, end) {
  const a = source.indexOf(start), b = source.indexOf(end, a);
  assert.ok(a >= 0 && b > a, `${start} / ${end}`);
  return source.slice(a, b);
}
function deferred() {
  let resolve, reject;
  const promise = new Promise((yes, no) => { resolve = yes; reject = no; });
  return { promise, resolve, reject };
}
const server = (id = 'srv-11ac29', node_id = 'local') => ({ id, node_id, name: id, stats: {} });
const failure = (status = 503) => Object.assign(Error('Unavailable'), { status });

function setup(url = '/server/11ac29/console') {
  const timers = new Map(), saved = new Map(), calls = [], selected = [], entered = [], rendered = [];
  let nextTimer = 0;
  const c = vm.createContext({
    URL, AbortController, console,
    S: { token: 'user-token', user: { role: 'user' }, servers: [], cs: -1, nodes: [], acct: {}, branding: {}, loading: {}, view: 'con' },
    QOL: { openSequence: 0 }, V: { con: () => 'console', dash: () => 'dashboard', notfound: () => '404', plug: () => 'plugins' },
    window: { location: new URL(url, 'https://panel.test') },
    localStorage: { getItem: key => saved.get(key) || null, setItem: (key, value) => saved.set(key, value), removeItem: key => saved.delete(key) },
    setTimeout: (fn, delay) => { const id = ++nextTimer; timers.set(id, { fn, delay }); return id; },
    clearTimeout: id => timers.delete(id), clearInterval: () => {},
    captureEditor: () => {}, editorIsDirty: () => false, flushPendingSettingsBeforeLeave: () => {},
    closeAllCtx: () => {}, closeFloatingCtx: () => {}, closeAllCustomSelects: () => {},
    $: () => null, startPageProgress: () => {}, finishPageProgress: () => {}, setPageProgress: () => {},
    applyBranding: () => {}, toast: () => {}, onEnterView: view => entered.push(view),
    isModdedEngine: () => false, esc: value => String(value ?? ''), ic: () => '',
    setAppearanceMode: () => {}, toggleCompact: () => {}, togglePrivacy: () => {},
    openPendingInvitation: async () => false, loadPublicBranding: async () => {}, loadRuntimeMods: async () => [],
    loadPublicCommunity: async () => {}, serverPagePermission: () => '',
    computeGravatarHash: async () => '', loadInitialRouteData: async view => entered.push(view), loadVersionsCatalog: () => {},
    showSetupScreen: () => { c.setupShown = true; }, showLoginScreen: () => { c.loginShown = true; },
    fetch: async endpoint => ({ ok: true, json: async () => endpoint === '/api/auth/status' ? {} : { username: 'User', role: 'user' } }),
  });
  c.history = {
    replaceState: (_, __, target) => { c.window.location = new URL(target, c.window.location); },
    pushState: (_, __, target) => { c.window.location = new URL(target, c.window.location); }
  };
  c.api = async endpoint => {
    calls.push(endpoint);
    if (endpoint === '/api/cluster/servers') return c.inventory();
    if (endpoint === '/api/servers') return c.localInventory();
    return [];
  };
  c.inventory = async () => ({ servers: [server()], node_errors: {} });
  c.localInventory = async () => { throw failure(); };
  c.syncActiveServerState = srv => {
    c.S.cs = c.S.servers.indexOf(srv);
    c.S.activeServerId = srv.id;
    c.S.activeNodeId = srv.node_id;
    selected.push(`${srv.node_id}/${srv.id}`);
    c.saveActiveServerRef(srv);
  };
  c.R = () => rendered.push(c.S.view);
  vm.runInContext([
    slice('const VIEW_TO_SLUG =', 'const ADMIN_TAB_SLUGS'),
    "const ADMIN_TAB_SLUGS = ['health', 'nodes']; const ACC_TAB_SLUGS = ['general', 'security'];",
    slice('function shortServerId(', 'function getViewDisplayLabel('),
    slice('function applyRoute(', 'function navigateToUrl('),
    slice('function curServer()', '// --- Advanced Minecraft'),
    slice('function normalizeRoute(', 'window.go ='),
    slice('let dashboardLoadSequence =', 'async function loadAdminNodes('),
    slice('async function checkAuth()', 'function showSetupScreen('),
    'Object.assign(V, {' + slice('  serverpending() {', '  notfound() {') + '});'
  ].join('\n'), c);
  return { c, timers, calls, selected, entered, rendered, saved };
}

async function flush() { for (let i = 0; i < 40; i++) await Promise.resolve(); }
async function runRouteTimer(timers) {
  const entry = [...timers].find(([, item]) => item.delay < 30000);
  assert.ok(entry, 'expected a bounded recovery timer');
  timers.delete(entry[0]); entry[1].fn(); await flush();
}

test('direct-link authentication startup resolves short IDs without changing the URL', async () => {
  const { c, entered } = setup();
  await c.checkAuth();
  assert.equal(c.loginShown, undefined);
  assert.equal(c.S.view, 'con');
  assert.equal(c.curServer().id, 'srv-11ac29');
  assert.equal(c.window.location.pathname, '/server/11ac29/console');
  assert.ok(entered.includes('con'));
});

test('cold direct link survives transient cluster and local failures, then recovers automatically', async () => {
  const { c, timers, selected, entered } = setup();
  c.inventory = async () => { throw failure(); };
  await c.checkAuth();
  assert.equal(c.S.token, 'user-token');
  assert.equal(c.S.view, 'serverpending');
  assert.equal(c.curServer(), null);
  assert.equal(selected.length, 0);
  assert.equal(entered.length, 0);
  assert.match(c.V.serverpending(), /temporarily unavailable/);
  assert.doesNotMatch(c.V.serverpending(), /404|Page Not Found/);
  c.inventory = async () => ({ servers: [server()], node_errors: {} });
  await runRouteTimer(timers);
  assert.equal(c.S.view, 'con');
  assert.equal(c.S.serverRoutePending, null);
  assert.equal(c.curServer().id, 'srv-11ac29');
});

test('complete authorized inventory makes nonexistent or unauthorized IDs a real not-found with no retries', async () => {
  const { c, timers, selected } = setup();
  c.inventory = async () => ({ servers: [server('srv-other')], node_errors: {} });
  await c.checkAuth();
  assert.equal(c.S.view, 'notfound');
  assert.equal(c.curServer(), null);
  assert.equal(selected.length, 0, 'must not connect to an unrelated first server');
  await flush();
  assert.equal(timers.size, 0);
});

test('remote node errors are incomplete inventory, not proof of absence; local matches still work', async () => {
  const { c, timers } = setup('/server/11ac29/files?node=east');
  c.inventory = async () => ({ servers: [server()], node_errors: { east: 'offline' } });
  await c.checkAuth();
  assert.equal(c.S.view, 'serverpending');
  assert.equal(c.curServer(), null, 'must not fall back to the local server with the same ID');
  c.inventory = async () => ({ servers: [server(), server('srv-11ac29', 'east')], node_errors: {} });
  await runRouteTimer(timers);
  assert.equal(c.S.view, 'fi');
  assert.equal(c.curServer().node_id, 'east');
  assert.equal(c.window.location.search, '?node=east');
  c.applyRoute('/server/11ac29/console');
  assert.equal(c.curServer().node_id, 'local');
});

test('local fallback can resolve local links but cannot declare remote links nonexistent', async () => {
  const { c } = setup();
  c.inventory = async () => { throw failure(); };
  c.localInventory = async () => [server()];
  await c.checkAuth();
  assert.equal(c.S.view, 'con');
  c.applyRoute('/server/remote/console?node=east');
  assert.equal(c.S.view, 'serverpending');
  assert.equal(c.curServer(), null);
});

test('missing explicit local or healthy-node server is authoritative even with unrelated offline nodes', async () => {
  for (const query of ['?node=local', '?node=west']) {
    const { c, timers } = setup('/server/missing/console' + query);
    c.inventory = async () => ({ servers: [], node_errors: { east: 'offline' } });
    await c.checkAuth();
    assert.equal(c.S.view, 'notfound');
    await flush();
    assert.equal(timers.size, 0);
  }
});

test('unqualified missing ID stays recoverable when any node failed', async () => {
  const { c } = setup();
  c.inventory = async () => ({ servers: [], node_errors: { east: 'offline' } });
  await c.checkAuth();
  assert.equal(c.S.view, 'serverpending');
});

test('short, prefixed and legacy remote URLs select the same authorized server and preserve canonical URLs', async () => {
  for (const url of ['/server/11ac29/console?node=east', '/server/srv-11ac29/console?node=east', '/server/east/srv-11ac29/console']) {
    const { c } = setup(url);
    c.inventory = async () => ({ servers: [server('srv-11ac29', 'east')], node_errors: {} });
    await c.checkAuth();
    assert.equal(c.S.view, 'con');
    assert.equal(c.curServer().node_id, 'east');
    assert.equal(c.window.location.pathname + c.window.location.search, '/server/11ac29/console?node=east');
  }
});

test('ambiguous remote IDs never select an arbitrary node; explicit scope is strict', () => {
  const { c } = setup();
  c.S.servers = [server('srv-11ac29', 'east'), server('srv-11ac29', 'west')];
  assert.equal(c.resolveServerByParam('11ac29'), null);
  assert.equal(c.resolveServerByParam('11ac29', 'local'), null);
  assert.equal(c.resolveServerByParam('11ac29', 'west').node_id, 'west');
  assert.equal(c.resolveServerByParam('11ac'), null);
});

test('only two automatic retries, then manual retry can recover without losing the deep link', async () => {
  const { c, timers, calls } = setup();
  c.inventory = async () => { throw failure(); };
  await c.checkAuth();
  await runRouteTimer(timers);
  await runRouteTimer(timers);
  assert.equal(c.S.view, 'serverpending');
  assert.equal(c.S.serverRoutePending.attempts, 2);
  assert.equal(timers.size, 0);
  assert.equal(calls.filter(url => url === '/api/cluster/servers').length, 3);
  c.inventory = async () => ({ servers: [server()], node_errors: {} });
  await c.retryServerRoute();
  assert.equal(c.S.view, 'con');
  assert.equal(c.window.location.pathname, '/server/11ac29/console');
});

test('manual retry is single-flight and disables the retry button while loading', async () => {
  const { c, calls } = setup();
  c.inventory = async () => { throw failure(); };
  await c.checkAuth();
  const result = deferred(); c.inventory = () => result.promise;
  const first = c.retryServerRoute();
  await c.retryServerRoute();
  assert.equal(calls.filter(url => url === '/api/cluster/servers').length, 2);
  assert.match(c.V.serverpending(), /disabled/);
  result.resolve({ servers: [server()], node_errors: {} });
  await first;
  assert.equal(c.S.view, 'con');
});

test('navigation away cancels pending recovery and a late result cannot steal the route', async () => {
  const { c, timers } = setup();
  c.inventory = async () => { throw failure(); };
  await c.checkAuth();
  const result = deferred(); c.inventory = () => result.promise;
  const retry = c.retryServerRoute();
  c.applyRoute('/account');
  result.resolve({ servers: [server()], node_errors: {} });
  await retry;
  assert.equal(c.S.view, 'acc');
  assert.equal(c.window.location.pathname, '/account');
  assert.equal(c.S.serverRoutePending, null);
  await flush();
  assert.equal(timers.size, 0);
});

test('latest inventory request wins even when an older request finishes last', async () => {
  const { c } = setup('/');
  const old = deferred(); c.inventory = () => old.promise;
  const first = c.loadDashboard(true);
  c.inventory = async () => ({ servers: [server('srv-new')], node_errors: {} });
  await c.loadDashboard(true);
  old.resolve({ servers: [server('srv-old')], node_errors: { east: 'offline' } });
  await first;
  assert.equal(c.curServer().id, 'srv-new');
  assert.equal(c.S.servers.length, 1);
  assert.equal(Object.keys(c.S.clusterNodeErrors).length, 0);
});

test('inventory refresh respects selection/navigation performed while it was in flight', async () => {
  const { c } = setup();
  const a = server(), b = server('srv-second', 'east');
  c.S.servers = [a, b]; c.S.cs = 0;
  const result = deferred(); c.inventory = () => result.promise;
  const load = c.loadDashboard(true);
  c.applyRoute('/server/second/console?node=east');
  result.resolve({ servers: [b, a], node_errors: {} });
  await load;
  assert.equal(c.curServer().id, b.id);
  assert.equal(c.S.activeNodeId, 'east');
  assert.equal(c.S.cs, 0);
});

test('startup re-reads navigation after asynchronous authentication and inventory loading', async () => {
  const { c } = setup();
  const result = deferred(); c.inventory = () => result.promise;
  const auth = c.checkAuth(); await flush();
  c.applyRoute('/account/security');
  result.resolve({ servers: [server()], node_errors: {} });
  await auth;
  assert.equal(c.S.view, 'acc');
  assert.equal(c.S.acct.tab, 1);
  assert.equal(c.window.location.pathname, '/account/security');
});

test('background dashboard failure on an open server route renders unavailable, never another server', async () => {
  const { c } = setup();
  await c.checkAuth();
  let closed = false;
  c.S.activeWs = { close: () => { closed = true; } };
  c.inventory = async () => { throw failure(); };
  await c.loadDashboard();
  assert.equal(c.S.view, 'serverpending');
  assert.equal(c.curServer(), null);
  assert.equal(c.S.activeServerId, null);
  assert.equal(closed, true);
});

test('authorization failures cannot use fallback or stale authorized inventory', async () => {
  for (const status of [401, 403]) {
    const { c, calls } = setup();
    c.S.servers = [server()]; c.S.cs = 0;
    c.inventory = async () => { throw failure(status); };
    c.localInventory = async () => [server()];
    await c.checkAuth();
    assert.equal(c.S.view, 'notfound');
    assert.equal(c.S.servers.length, 0);
    assert.equal(c.curServer(), null);
    assert.ok(!calls.includes('/api/servers'));
  }
});

test('a response from a previous login session cannot populate server inventory', async () => {
  const { c } = setup();
  const result = deferred(); c.inventory = () => result.promise;
  const load = c.loadDashboard(true);
  c.S.token = 'different-user';
  result.resolve({ servers: [server()], node_errors: {} });
  await load;
  assert.equal(c.S.servers.length, 0);
  assert.equal(c.curServer(), null);
});

test('slow optional activity/nodes requests do not block a valid deep link', async () => {
  const { c } = setup();
  c.api = async endpoint => {
    if (endpoint === '/api/cluster/servers') return { servers: [server()], node_errors: {} };
    return new Promise(() => {});
  };
  await c.checkAuth();
  assert.equal(c.S.view, 'con');
  assert.equal(c.curServer().id, 'srv-11ac29');
});

test('malformed inventory is unavailable rather than authoritative empty data', async () => {
  const { c } = setup();
  c.inventory = async () => ({ error: 'Unexpected response' });
  c.localInventory = async () => ({});
  await c.checkAuth();
  assert.equal(c.S.view, 'serverpending');
});

test('unknown page slugs remain not-found even when the server inventory is unavailable', async () => {
  const { c, timers } = setup('/not-a-page');
  c.inventory = async () => { throw failure(); };
  await c.checkAuth();
  assert.equal(c.S.view, 'notfound');
  await flush();
  assert.equal(timers.size, 0);
});

test('server switch keeps the clicked identity across an asynchronous confirmation and reordered list', async () => {
  const { c } = setup();
  const a = server(), b = server('srv-second', 'east');
  c.S.servers = [a, b]; c.S.cs = 0;
  const confirm = deferred(); c.guardEditorLeave = () => confirm.promise;
  vm.runInContext(slice('async function switchServer(', '// --- Command Palette'), c);
  const switching = c.switchServer(1);
  c.S.servers = [b, a];
  confirm.resolve(true); await switching;
  assert.equal(c.curServer().id, b.id);
  assert.equal(c.window.location.pathname + c.window.location.search, '/server/second/console?node=east');
});

test('late optional node metadata is accepted without delaying server routing', async () => {
  const { c } = setup();
  const nodes = deferred(); let signal;
  c.api = async (endpoint, options) => {
    signal = options.signal;
    if (endpoint === '/api/cluster/servers') return { servers: [server()], node_errors: {} };
    if (endpoint === '/api/nodes') return nodes.promise;
    return [];
  };
  await c.checkAuth();
  assert.equal(c.S.view, 'con');
  assert.equal(signal.aborted, false);
  nodes.resolve([{ id: 'east' }]); await flush();
  assert.equal(c.S.nodes[0].id, 'east');
  assert.equal(signal.aborted, true);
});

test('hung inventory requests time out into a recoverable view instead of hanging indefinitely', async () => {
  const { c, timers } = setup();
  c.api = (endpoint, { signal }) => {
    if (!['/api/cluster/servers', '/api/servers'].includes(endpoint)) return Promise.resolve([]);
    return new Promise((resolve, reject) => {
      if (signal.aborted) reject(Error('Aborted'));
      else signal.addEventListener('abort', () => reject(Error('Aborted')), { once: true });
    });
  };
  const auth = c.checkAuth(); await flush();
  const timeout = [...timers].find(([, item]) => item.delay === 30000);
  assert.ok(timeout);
  timers.delete(timeout[0]); timeout[1].fn();
  await auth;
  assert.equal(c.S.view, 'serverpending');
  assert.equal(c.S.token, 'user-token');
});

test('revoked access in a refreshed authorized inventory removes cached servers and stops recovery', async () => {
  const { c, timers } = setup();
  await c.checkAuth();
  c.inventory = async () => ({ servers: [], node_errors: {} });
  await c.loadDashboard(); await flush();
  assert.equal(c.S.view, 'notfound');
  assert.equal(c.S.servers.length, 0);
  assert.equal(c.curServer(), null);
  assert.equal(timers.size, 0);
});
