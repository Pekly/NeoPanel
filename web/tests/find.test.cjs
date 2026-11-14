const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/find.js'), 'utf8');
const searchSource = fs.readFileSync(path.join(__dirname, '../static/js/find-search.js'), 'utf8');
function setup({ noEditor = false, noWorker = false, text = 'key=one\nkey=two' } = {}) {
  const timers = new Map(), intervals = new Map(), workers = []; let nextID = 0, doc;
  class Element {
    constructor(tag) { this.tagName = tag.toUpperCase(); this.children = []; this.dataset = {}; this.events = {}; this.attrs = {}; this.value = ''; this.selectionStart = this.selectionEnd = 0; this.scrollTop = this.scrollLeft = 0; this.clientHeight = 200; this.clientWidth = 400; }
    get isConnected() { return this === doc.body || Boolean(this.parent?.isConnected); }
    get childElementCount() { return this.children.length; }
    appendChild(child) { child.remove(); this.children.push(child); child.parent = this; }
    remove() { if (this.parent) this.parent.children = this.parent.children.filter(c => c !== this); this.parent = null; }
    setAttribute(key, value) { this.attrs[key] = value; }
    addEventListener(key, fn) { (this.events[key] ||= []).push(fn); }
    removeEventListener(key, fn) { this.events[key] = (this.events[key] || []).filter(f => f !== fn); }
    dispatchEvent(event) { for (const fn of this.events[event.type] || []) fn(event); }
    closest(selector) { return selector === '.np-editor-workspace' ? this.workspace ? this : null : this.hidden ? this : null; }
    getClientRects() { return this.hidden ? [] : [{}]; }
    focus() { doc.activeElement = this; }
    select() { this.selectionStart = 0; this.selectionEnd = this.value.length; }
    setSelectionRange(a, b) { this.selectionStart = a; this.selectionEnd = b; }
    scrollIntoView() { this.scrolled = true; }
  }
  const body = new Element('body');
  doc = { body, events: {}, activeElement: body, createElement: tag => new Element(tag),
    addEventListener(key, fn) { this.events[key] = fn; },
    getElementById(id) { const walk = el => el.id === id ? el : el.children.map(walk).find(Boolean); return walk(body); } };
  const main = new Element('main'); main.id = 'main'; body.appendChild(main);
  const ed = new Element('textarea'); ed.id = 'ed'; ed.workspace = true; ed.dataset.editorKey = 'server-a/config.yml'; ed.value = text;
  if (!noEditor) main.appendChild(ed);
  class Worker {
    constructor(url) { this.url = url; this.jobs = []; workers.push(this); }
    postMessage(job) { this.jobs.push(job); }
    terminate() { this.terminated = true; }
    reply(job = this.jobs.at(-1)) { this.onmessage({ data: { id: job.id, ...c.NPFileSearch.search(job.text, job.query) } }); }
  }
  const c = vm.createContext({ document: doc, Worker: noWorker ? undefined : Worker, Event: class { constructor(type) { this.type = type; } },
    getComputedStyle: () => ({ fontSize: '14px', lineHeight: '24px' }),
    setTimeout(fn, ms) { timers.set(++nextID, { fn, ms }); return nextID; }, clearTimeout: id => timers.delete(id),
    setInterval(fn) { intervals.set(++nextID, fn); return nextID; }, clearInterval: id => intervals.delete(id) });
  c.window = c; vm.runInContext(searchSource + '\n' + source, c);
  function flush(ms = 100) { for (const [id, entry] of [...timers]) if (entry.ms === ms) { timers.delete(id); entry.fn(); } }
  function key(key, extra = {}) { const event = { key, target: doc.activeElement, preventDefault() { this.prevented = true; }, stopImmediatePropagation() { this.stopped = true; }, ...extra }; doc.events.keydown(event); return event; }
  function type(value) { const input = doc.getElementById('np-find-query'); input.value = value; input.dispatchEvent({ type: 'input' }); }
  return { c, doc, ed, main, body, workers, timers, intervals, flush, key, type, count: () => doc.getElementById('np-find-count')?.textContent };
}
test('Ctrl/Cmd+F is built-in only for the verified text-file editor', () => {
  const outside = setup({ noEditor: true }); assert.equal(outside.c.NPFind.open(), false);
  assert.equal(outside.key('f', { ctrlKey: true }).prevented, undefined);
  const h = setup(); assert.equal(h.key('f', { ctrlKey: true }).prevented, true);
  assert.equal(h.doc.activeElement.id, 'np-find-query'); assert.equal(h.doc.activeElement.attrs['aria-label'], 'Find in text file');
  h.c.NPFind.close(); assert.equal(h.key('f', { metaKey: true }).prevented, true);
});
test('login, modals, hidden and unverified textareas do not intercept browser find', () => {
  for (const block of [h => { h.ed.hidden = true; }, h => { h.ed.workspace = false; }, h => { h.ed.dataset.editorKey = ''; },
    h => { const login = h.doc.createElement('div'); login.id = 'login'; h.body.appendChild(login); },
    h => { const modal = h.doc.createElement('div'); modal.id = 'modal-container'; modal.appendChild(h.doc.createElement('div')); h.body.appendChild(modal); }]) {
    const h = setup(); block(h); assert.equal(h.key('f', { ctrlKey: true }).prevented, undefined);
  }
});
test('search runs off-thread and Enter/F3 reuse cached offsets without scanning or editing', () => {
  const h = setup(), original = h.ed.value;
  h.ed.focus(); h.c.NPFind.open(); h.type('key'); h.flush();
  assert.equal(h.workers.length, 1); assert.equal(h.workers[0].url, '/js/find-worker.js');
  assert.equal(h.workers[0].jobs[0].text, original); h.workers[0].reply();
  assert.equal(h.count(), '1 of 2'); assert.deepEqual([h.ed.selectionStart, h.ed.selectionEnd], [0, 3]);
  h.key('Enter'); assert.equal(h.count(), '2 of 2'); assert.equal(h.doc.activeElement, h.ed);
  h.key('Enter', { shiftKey: true }); assert.equal(h.count(), '1 of 2');
  h.key('F3'); assert.equal(h.count(), '2 of 2');
  assert.equal(h.workers[0].jobs.length, 1); assert.equal(h.ed.value, original);
  const escape = h.key('Escape'); assert.ok(escape.prevented && escape.stopped); assert.equal(h.count(), undefined);
  assert.equal(h.workers[0].terminated, true); assert.equal(h.timers.size, 0); assert.equal(h.intervals.size, 0);
});
test('rapid queries coalesce and stale worker results cannot overwrite the newest search', () => {
  const h = setup(); h.c.NPFind.open(); h.type('one'); h.flush(); const old = h.workers[0], oldReply = old.onmessage, job = old.jobs[0];
  for (let i = 0; i < 30; i++) h.type(i === 29 ? 'two' : 'key');
  h.flush(); assert.equal(h.workers.length, 2); assert.equal(old.terminated, true);
  h.workers[1].reply(); assert.equal(h.count(), '1 of 1'); assert.equal(h.ed.selectionStart, h.ed.value.indexOf('two'));
  oldReply({ data: { id: job.id, matches: [[0, 3, 0, 0]], limited: false } });
  assert.equal(h.ed.selectionStart, h.ed.value.indexOf('two'));
});
test('file changes invalidate results and input rebuilds results without stealing focus', () => {
  const h = setup(); h.c.NPFind.open(); h.type('key'); h.flush(); h.workers[0].reply();
  h.ed.value += '\nkey=three'; h.ed.focus(); h.ed.dispatchEvent({ type: 'input' });
  assert.equal(h.count(), 'Searching…'); h.flush(); h.workers[0].reply(); assert.equal(h.count(), '3'); assert.equal(h.doc.activeElement, h.ed);
  h.ed.dataset.editorKey = 'server-b/another.yml'; h.c.NPFind.refresh();
  assert.equal(h.count(), undefined); assert.equal(h.workers[0].terminated, true);
});
test('refresh rebinds a replacement textarea for the same file without losing scope', () => {
  const h = setup(); h.c.NPFind.open(); h.type('key'); h.flush(); h.workers[0].reply();
  const replacement = h.doc.createElement('textarea'); replacement.id = 'ed'; replacement.workspace = true;
  replacement.dataset.editorKey = h.ed.dataset.editorKey; replacement.value = 'key=new';
  h.ed.remove(); h.main.appendChild(replacement); h.c.NPFind.refresh(); h.flush(); h.workers[0].reply();
  h.key('F3'); assert.deepEqual([replacement.selectionStart, replacement.selectionEnd], [0, 3]);
});
test('worker failure has a bounded fallback and honestly labels partial results', () => {
  const h = setup({ noWorker: true, text: 'x'.repeat(150000) + 'needle' });
  h.c.NPFind.open(); h.type('needle'); h.flush(); assert.equal(h.count(), '0 matches (partial)');
  assert.match(h.doc.getElementById('np-find-count').title, /131,072/);
});
test('hung worker is terminated at its deadline; close releases pending work', () => {
  const h = setup(); h.c.NPFind.open(); h.type('key'); h.flush(); h.flush(800);
  assert.equal(h.workers[0].terminated, true); assert.equal(h.count(), '1 of 2');
  h.type('one'); h.flush(); h.c.NPFind.close(); assert.equal(h.timers.size, 0); assert.equal(h.intervals.size, 0);
});
test('selected editor text seeds find; IME and modified shortcuts are left alone', () => {
  const h = setup(); h.ed.focus(); h.ed.setSelectionRange(4, 7); h.c.NPFind.open();
  assert.equal(h.doc.getElementById('np-find-query').value, 'one');
  assert.equal(h.key('f', { ctrlKey: true, altKey: true }).prevented, undefined);
  assert.equal(h.key('f', { ctrlKey: true, isComposing: true }).prevented, undefined);
});
test('literal search is Unicode-safe, multiline-aware and never interprets regex syntax', () => {
  const { c } = setup(); const search = c.NPFileSearch.search;
  let result = search('İ foo\nFOO [a].*', 'foo');
  assert.deepEqual(JSON.parse(JSON.stringify(result.matches)), [[2, 5, 0, 2], [6, 9, 1, 0]]);
  assert.equal(search('x [a].* y', '[a].*').matches.length, 1);
  assert.equal(search('😀 key\nkey', 'key').matches[0][0], 3);
  assert.equal(search('anything', '').matches.length, 0);
});
test('result and source limits bound memory while scanning to the end of supported text files', () => {
  const { c } = setup(), search = c.NPFileSearch.search, max = c.NPFileSearch.MAX_TEXT;
  assert.equal(search('key '.repeat(1000), 'key').limited, false);
  const capped = search('key '.repeat(1001), 'key'); assert.equal(capped.matches.length, 1000); assert.equal(capped.limited, true);
  const end = search('x'.repeat(max - 6) + 'needle', 'needle'); assert.equal(end.matches[0][0], max - 6);
  const beyond = search('x'.repeat(max) + 'needle', 'needle'); assert.equal(beyond.matches.length, 0); assert.equal(beyond.limited, true);
});
test('find never scans the page/mirror, stores content, or sends file text over HTTP', () => {
  assert.doesNotMatch(source, /MutationObserver|createTreeWalker|innerHTML|sessionStorage|localStorage|fetch\(/);
  const html = fs.readFileSync(path.join(__dirname, '../static/index.html'), 'utf8');
  assert.ok(html.indexOf('/js/find-search.js') < html.indexOf('/js/find.js'));
});
