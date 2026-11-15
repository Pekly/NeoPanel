const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/operation-progress.js'), 'utf8');
const app = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
function section(start, end) {
  const a = app.indexOf(start), b = app.indexOf(end, a);
  assert.ok(a >= 0 && b > a); return app.slice(a, b);
}
class Element {
  constructor(tag) { this.tagName = tag; this.children = []; this.attrs = {}; this.dataset = {}; this.className = ''; this.textContent = ''; }
  append(...children) { children.forEach(c => this.appendChild(c)); }
  appendChild(child) { child.remove(); this.children.push(child); child.parent = this; return child; }
  remove() { if (this.parent) this.parent.children = this.parent.children.filter(c => c !== this); this.parent = null; }
  get isConnected() { return this.tagName === 'body' || !!this.parent?.isConnected; }
  get childElementCount() { return this.children.length; }
  setAttribute(key, value) { this.attrs[key] = value; }
  querySelector(selector) { return this.find(el => selector === '[data-operation-progress]' && el.attrs['data-operation-progress'] !== undefined)[0] || null; }
  find(fn) { return this.children.flatMap(c => [ ...(fn(c) ? [c] : []), ...c.find(fn)]); }
}
function setup() {
  const body = new Element('body'), timers = new Map(); let now = 0, next = 0;
  const c = vm.createContext({ URL, FormData, console, Date: {now: () => now},
    document: {body, createElement: tag => new Element(tag), getElementById: id => body.find(el => el.id === id)[0]},
    location: {href: 'http://localhost:8085/'}, setInterval(fn) { timers.set(++next, fn); return next; }, clearInterval: id => timers.delete(id),
    S: {token: 'test'}, assertClientRequestAccess() {}, toast() {}, logout() {}, fetch: async () => ({ok: true, status: 200, json: async () => ({ok: true})})
  });
  c.window = c;
  vm.runInContext(source + '\n' + section('async function api(', 'function curServer()'), c);
  return {c, body, timers, cards: () => body.find(e => e.className === 'np-operation-card'), tick(seconds) { now += seconds * 1000; for (const fn of timers.values()) fn(); }};
}
const endpoint = '/api/servers/srv-one/reinstall';
const options = {method: 'POST'};

test('long operations have indeterminate progress for local and proxied hosts', () => {
  const {c, body, cards, timers} = setup();
  for (const prefix of ['', '/api/nodes/east/proxy']) {
    for (const suffix of ['reinstall', 'software', 'plugins/install', 'backups', 'backups/restore', 'files/compress', 'files/archive', 'files/unzip', 'files/extract']) {
      const end = c.NPProgress.startRequest(prefix + '/api/servers/srv-one/' + suffix, options);
      assert.equal(cards().length, 1);
      const bar = body.find(e => e.attrs.role === 'progressbar')[0];
      assert.ok(bar.attrs['aria-label']);
      assert.equal(bar.attrs['aria-valuenow'], undefined, 'no fabricated percentage');
      end(); assert.equal(cards().length, 0); assert.equal(timers.size, 0);
    }
    const end = c.NPProgress.startRequest(prefix + '/api/system/java/install', options);
    assert.equal(cards().length, 1); end();
  }
});

test('polling, uploads and unrelated requests do not create duplicate progress UI', () => {
  const {c, cards, timers} = setup();
  for (const [url, opts] of [[endpoint, {}], ['/api/servers/srv-one/files/upload', options], ['/api/servers/srv-one', {method: 'PUT'}], ['/api/servers', {}]]) {
    c.NPProgress.startRequest(url, opts)();
  }
  assert.equal(cards().length, 0); assert.equal(timers.size, 0);
});

test('concurrent progress, elapsed time, long waits and idempotent cleanup', () => {
  const {c, cards, timers, tick, body} = setup();
  const first = c.NPProgress.startRequest(endpoint, options);
  const second = c.NPProgress.startRequest('/api/system/java/install', options);
  assert.equal(cards().length, 2); assert.equal(timers.size, 1);
  tick(65);
  assert.equal(body.find(e => e.className === 'np-operation-elapsed')[0].textContent, '1m 5s elapsed');
  assert.match(body.find(e => e.className === 'np-operation-hint')[0].textContent, /Still waiting/);
  first(); first(); assert.equal(cards().length, 1); assert.equal(timers.size, 1);
  second(); assert.equal(body.children.length, 0); assert.equal(timers.size, 0);
});

test('reinstall mounts only into its own dialog and survives closing that dialog', () => {
  const {c, body, cards, tick} = setup();
  const form = new Element('form'); form.id = 'reinstall-form'; form.dataset = {srvId: 'srv-one', nodeId: 'local'};
  const slot = new Element('div'); slot.setAttribute('data-operation-progress', ''); form.append(slot); body.append(form);
  const end = c.NPProgress.startRequest(endpoint, options);
  assert.equal(slot.children.length, 1); assert.equal(body.children.length, 1);
  form.remove(); tick(1); assert.equal(cards().length, 1); assert.equal(body.children[0].className, 'np-operation-tray');
  end(); assert.equal(body.children.length, 0);
});

test('clearing a session removes cards and old completion cannot remove new work', () => {
  const {c, cards, timers} = setup();
  const old = c.NPProgress.startRequest(endpoint, options);
  c.NPProgress.clear(); assert.equal(cards().length, 0); assert.equal(timers.size, 0);
  const current = c.NPProgress.startRequest(endpoint, options);
  old(); assert.equal(cards().length, 1); current(); assert.equal(timers.size, 0);
});

test('API progress remains until JSON completion and cleans up on success/error/abort/401', async () => {
  for (const outcome of ['success', 'http', 'abort', 'unauthorized']) {
    const {c, cards, timers} = setup(); let release;
    c.fetch = async () => {
      await new Promise(resolve => { release = resolve; });
      if (outcome === 'abort') throw new Error('aborted');
      return {status: outcome === 'unauthorized' ? 401 : outcome === 'http' ? 500 : 200, ok: outcome === 'success', json: async () => ({error: 'upstream unavailable'})};
    };
    const pending = c.api(endpoint, options);
    assert.equal(cards().length, 1); release();
    if (outcome === 'success') await pending; else await assert.rejects(pending);
    assert.equal(cards().length, 0); assert.equal(timers.size, 0);
  }
});

test('reinstall uses a spinner, rejects duplicate submissions and permits retry after failure', async () => {
  let release; const calls = [], button = {}, form = {};
  const c = vm.createContext({S: {}, $: id => id === '#reinst-submit-btn' ? button : id === '#reinstall-form' ? form : undefined,
    setBtnLoading: (btn, busy) => { btn.busy = busy; }, toast() {}, isSteelEngine: () => false,
    nodeApiPath: (_, url) => url, api: url => { calls.push(url); return new Promise((resolve, reject) => { release = reject; }); },
    uiAlert() {}, esc: String});
  vm.runInContext(section('async function submitReinstall(', 'async function openAdminServerSettingsModal('), c);
  const run = () => c.submitReinstall({preventDefault() {}}, 'srv-one');
  const first = run(); assert.equal(button.busy, true); await run(); assert.equal(calls.length, 1);
  release(new Error('failed')); await first; assert.equal(button.busy, false); assert.equal(c.S.reinstallingServers.size, 0);
  const retry = run(); assert.equal(calls.length, 2); release(new Error('failed')); await retry;
  assert.match(section('function setBtnLoading(', 'function av('), /np-spinner-sm/);
});

test('progress assets load before app and honor reduced motion', () => {
  const html = fs.readFileSync(path.join(__dirname, '../static/index.html'), 'utf8');
  const css = fs.readFileSync(path.join(__dirname, '../static/css/operation-progress.css'), 'utf8');
  assert.ok(html.indexOf('/js/operation-progress.js') < html.indexOf('/js/app.js'));
  assert.match(css, /prefers-reduced-motion: reduce/);
  assert.match(css, /data-panel-motion="reduced"/);
  assert.match(css, /animation: np-operation-sweep/);
  assert.doesNotMatch(source, /innerHTML|aria-valuenow/);
});
