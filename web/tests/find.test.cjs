const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/find.js'), 'utf8');

function setup(customHighlights = true) {
  const listeners = {}, observers = [], tasks = new Map(), scrolls = [], focusCalls = [];
  let serial = 0, now = 0, measurements = 0;
  const doc = { addEventListener(name, fn, capture) { listeners[name] = { fn, capture }; } };
  class Node {
    constructor(tag, value = '') {
      this.nodeType = tag === '#text' ? 3 : 1; this.tagName = tag.toUpperCase(); this.nodeValue = value;
      this.childNodes = []; this.parentNode = null; this.attributes = {}; this.style = {}; this.events = {};
      this.id = ''; this.value = ''; this.textContent = ''; this.clientHeight = 100; this.scrollHeight = 100;
      this.selectionStart = this.selectionEnd = 0;
    }
    get parentElement() { return this.parentNode; }
    get firstChild() { return this.childNodes[0] || null; }
    get nextSibling() { return this.parentNode?.childNodes[this.parentNode.childNodes.indexOf(this) + 1] || null; }
    get childElementCount() { return this.childNodes.filter(n => n.nodeType === 1).length; }
    get isConnected() { let n = this; while (n.parentNode) n = n.parentNode; return n === doc.body; }
    appendChild(child) { child.remove(); this.childNodes.push(child); child.parentNode = this; return child; }
    remove() { if (this.parentNode) this.parentNode.childNodes.splice(this.parentNode.childNodes.indexOf(this), 1); this.parentNode = null; }
    contains(node) { for (let n = node; n; n = n.parentNode) if (n === this) return true; return false; }
    setAttribute(key, value) { this.attributes[key] = value; }
    getAttribute(key) { return this.attributes[key] ?? null; }
    addEventListener(name, fn) { this.events[name] = fn; }
    removeEventListener(name, fn) { if (this.events[name] === fn) delete this.events[name]; }
    emit(name, detail = {}) { this.events[name]?.({ target: this, ...detail }); }
    focus(opts) { doc.activeElement = this; focusCalls.push({ node: this, opts }); }
    select() { this.selectionStart = 0; this.selectionEnd = this.value.length; }
    setSelectionRange(start, end) { this.selectionStart = start; this.selectionEnd = end; }
    scrollIntoView(opts) { scrolls.push({ node: this, opts }); }
    scrollTo(opts) { scrolls.push({ node: this, opts }); }
    scrollBy(opts) { scrolls.push({ node: this, opts }); }
    getBoundingClientRect() { return { top: 0, bottom: 100, left: 0, right: 100 }; }
  }
  class Range {
    setStart(node, offset) { this.startContainer = node; this.startOffset = offset; }
    setEnd(node, offset) { this.endContainer = node; this.endOffset = offset; }
    get commonAncestorContainer() {
      for (let n = this.startContainer; n; n = n.parentNode) if (n.contains(this.endContainer)) return n;
    }
    intersectsNode(node) { return this.commonAncestorContainer.contains(node); }
    cloneRange() { return Object.assign(new Range(), this); }
    getBoundingClientRect() { return { top: 1000, bottom: 1016, left: 0, right: 80 }; }
  }
  const selection = {
    ranges: [], get rangeCount() { return this.ranges.length; },
    get isCollapsed() { return !this.ranges.length; },
    get anchorNode() { return this.ranges[0]?.startContainer; },
    get focusNode() { return this.ranges[0]?.endContainer; },
    getRangeAt(i) { return this.ranges[i]; }, removeAllRanges() { this.ranges = []; },
    addRange(range) { this.ranges.push(range); }
  };
  doc.body = new Node('body'); doc.activeElement = doc.body;
  doc.createElement = tag => new Node(tag); doc.createRange = () => new Range();
  doc.createTreeWalker = root => {
    const queue = [...root.childNodes];
    return { nextNode() { const n = queue.shift(); if (n) queue.unshift(...n.childNodes); return n || null; } };
  };
  doc.getElementById = id => {
    const stack = [doc.body];
    while (stack.length) { const n = stack.pop(); if (n.id === id) return n; stack.push(...n.childNodes); }
    return null;
  };
  const element = (tag, text, parent = doc.body) => {
    const n = new Node(tag); if (text) n.appendChild(new Node('#text', text)); parent.appendChild(n); return n;
  };
  const main = element('main'); main.id = 'main';
  const modal = element('div'); modal.id = 'modal-container';
  const highlights = new Map();
  function addTask(fn, delay, interval = false) { const id = ++serial; tasks.set(id, { fn, at: now + delay, delay, interval }); return id; }
  const context = vm.createContext({ document: doc, innerHeight: 600,
    CSS: customHighlights ? { highlights } : undefined,
    Highlight: customHighlights ? class { constructor(...ranges) { this.ranges = ranges; } } : undefined,
    getSelection: () => selection,
    getComputedStyle: el => {
      measurements++;
      return { display: ['SPAN', 'B', 'EM', 'A', 'STRONG'].includes(el.tagName) ? 'inline' : 'block',
        visibility: 'visible', opacity: '1', overflowY: 'auto', lineHeight: '20px', fontSize: '14px', ...el.style };
    },
    scrollBy: opts => scrolls.push({ node: 'window', opts }),
    setTimeout: (fn, delay) => addTask(fn, delay), clearTimeout: id => tasks.delete(id),
    setInterval: (fn, delay) => addTask(fn, delay, true), clearInterval: id => tasks.delete(id),
    MutationObserver: class {
      constructor(fn) { this.fn = fn; observers.push(this); }
      observe(root, opts) { this.root = root; this.opts = opts; this.connected = true; }
      disconnect() { this.connected = false; }
    }
  });
  context.window = context;
  for (const key of ['localStorage', 'sessionStorage', 'fetch', 'XMLHttpRequest']) {
    Object.defineProperty(context, key, { get() { throw new Error('Forbidden access: ' + key); } });
  }
  vm.runInContext(source, context);
  function tick(ms = 100) {
    const end = now + ms;
    for (;;) {
      let chosen;
      for (const [id, task] of tasks) if (task.at <= end && (!chosen || task.at < chosen[1].at)) chosen = [id, task];
      if (!chosen) break;
      const [id, task] = chosen; now = task.at;
      if (task.interval) task.at += task.delay; else tasks.delete(id);
      task.fn();
    }
    now = end;
  }
  function key(key, detail = {}) {
    const event = { key, target: doc.activeElement, preventDefault() { this.prevented = true; },
      stopImmediatePropagation() { this.stopped = true; }, ...detail };
    listeners.keydown.fn(event); return event;
  }
  const input = () => doc.getElementById('np-find-query');
  function type(value, wait = true) { input().value = value; input().emit('input'); if (wait) tick(); }
  function select(node, start, end) { const r = new Range(); r.setStart(node, start); r.setEnd(node, end); selection.ranges = [r]; }
  return { api: context.NPFind, doc, main, modal, element, key, input, type, tick, select, selection,
    highlights, scrolls, focusCalls, observers, tasks, listeners, measurements: () => measurements,
    count: () => doc.getElementById('np-find-count')?.textContent,
    mutate: () => observers.filter(o => o.connected).forEach(o => o.fn([])),
    ranges: () => highlights.get('np-find-results')?.ranges || [] };
}

test('capture shortcut opens an accessible nonmodal bar, empty query never scans', () => {
  const s = setup(); s.element('p', 'hello', s.main);
  assert.equal(s.listeners.keydown.capture, true);
  const event = s.key('f', { ctrlKey: true }); assert.ok(event.prevented && event.stopped);
  assert.equal(s.doc.getElementById('np-find').getAttribute('role'), 'search');
  assert.equal(s.input().getAttribute('aria-label'), 'Find in workspace');
  assert.equal(s.doc.getElementById('np-find-count').getAttribute('aria-live'), 'polite');
  s.tick(); assert.equal(s.count(), 'Type to find'); assert.equal(s.ranges().length, 0);
});

test('native find is retained for login, overlays, hidden/no workspace and modified/composing chords', () => {
  for (const detail of [{ altKey: true }, { shiftKey: true }, { isComposing: true }, { keyCode: 229 }]) {
    const s = setup(); assert.ok(!s.key('f', { ctrlKey: true, ...detail }).prevented);
  }
  for (const blocker of ['login', 'pal', 'modal', 'hidden', 'none']) {
    const s = setup();
    if (blocker === 'modal') s.element('div', 'dialog', s.modal);
    else if (blocker === 'hidden') s.main.hidden = true;
    else if (blocker === 'none') s.main.remove();
    else s.element('div').id = blocker;
    assert.equal(s.api.open(), false); assert.ok(!s.key('F', { metaKey: true }).prevented);
  }
});

test('literal case-insensitive matches have original Unicode offsets and no regex injection', () => {
  const s = setup(); const p = s.element('p', 'İ 😀 A.*[x] a.*[X] Kelvin', s.main); s.api.open();
  s.type('a.*[x]'); assert.equal(s.count(), '1 of 2');
  assert.deepEqual(s.ranges().map(r => [r.startOffset, r.endOffset]), [[5, 11], [12, 18]]);
  assert.ok(s.ranges().every(r => r.startContainer === p.firstChild));
  s.type('kelvin'); assert.equal(s.count(), '1 of 1'); assert.equal(s.ranges()[0].startOffset, 19);
  s.type('['); assert.equal(s.ranges().length, 2);
});

test('phrases span inline markup but not block rows or line break elements; no node replacements', () => {
  const s = setup(), p = s.element('p', 'Hello ', s.main), b = s.element('b', 'WORLD', p);
  s.element('div', 'Hello ', s.main); s.element('div', 'world', s.main);
  const line = s.element('p', 'Hello ', s.main); s.element('br', '', line); s.element('span', 'world', line);
  b.addEventListener('click', () => {}); const handler = b.events.click;
  s.api.open(); s.type('hello world'); assert.equal(s.ranges().length, 1);
  assert.equal(s.ranges()[0].startContainer, p.firstChild); assert.equal(s.ranges()[0].endContainer, b.firstChild);
  assert.equal(b.events.click, handler); assert.equal(b.firstChild.nodeValue, 'WORLD');
});

test('excluded and hidden subtrees, credentials, options and collapsed details are not searched', () => {
  const s = setup(); s.element('p', 'needle', s.main);
  for (const tag of ['script', 'style', 'noscript', 'select', 'option', 'input', 'textarea', 'template']) s.element(tag, 'needle', s.main);
  for (const attr of ['hidden', 'inert', 'aria-hidden']) {
    const el = s.element('div', 'needle', s.main);
    if (attr === 'aria-hidden') el.setAttribute(attr, 'true'); else el[attr] = true;
  }
  for (const style of [{ display: 'none' }, { visibility: 'hidden' }, { opacity: '0' }, { contentVisibility: 'hidden' }]) {
    const el = s.element('div', '', s.main); el.style = style; s.element('span', 'needle', el);
  }
  const details = s.element('details', 'needle', s.main); s.element('summary', 'Summary', details); s.element('p', 'needle', details);
  s.element('p', 'needle'); s.api.open(); s.type('needle'); assert.equal(s.ranges().length, 1);
});

test('Enter/F3 and reversed navigation wrap, while Escape consumes only the find close', () => {
  const s = setup(); const prior = s.element('button', 'Prior'); prior.focus();
  s.element('p', 'one one', s.main); s.api.open(); s.type('one');
  assert.ok(s.key('Enter').prevented); assert.equal(s.count(), '2 of 2');
  s.key('F3'); assert.equal(s.count(), '1 of 2');
  s.key('F3', { shiftKey: true }); assert.equal(s.count(), '2 of 2');
  s.key('Enter', { shiftKey: true }); assert.equal(s.count(), '1 of 2');
  const event = s.key('Escape'); assert.ok(event.stopped && event.prevented);
  assert.equal(s.doc.getElementById('np-find'), null); assert.equal(s.doc.activeElement, prior);
  assert.equal(s.tasks.size, 0); assert.equal(s.highlights.size, 0); assert.ok(s.observers.every(o => !o.connected));
  assert.ok(s.scrolls.every(x => x.opts.behavior === 'instant'));
});

test('short workspace selection prefills, repeated Ctrl/Cmd+F selects the memory-only query', () => {
  const s = setup(); const p = s.element('p', 'hello world', s.main); s.select(p.firstChild, 0, 5);
  s.key('f', { metaKey: true }); assert.equal(s.input().value, 'hello');
  s.selection.removeAllRanges(); s.type('world'); s.key('f', { ctrlKey: true });
  assert.equal(s.input().value, 'world'); assert.equal(s.input().selectionEnd, 5);
  s.api.close(); s.api.open(); assert.equal(s.input().value, 'world');
});

test('short selection spanning inline markup can prefill without copying the entire workspace', () => {
  const s = setup(), p = s.element('p', 'Hello ', s.main); s.element('b', 'world', p);
  s.select(p, 0, 2); s.api.open(); assert.equal(s.input().value, 'Hello world');
});

test('oversized and external selections are not copied into the query', () => {
  const s = setup(); const large = s.element('p', 'x'.repeat(100000), s.main);
  s.select(large.firstChild, 0, 100000); s.api.open(); assert.equal(s.input().value, ''); s.api.close();
  const outside = s.element('p', 'private'); s.select(outside.firstChild, 0, 7); s.api.open(); assert.equal(s.input().value, '');
});

test('file editor searches its current value, selects offsets, and never writes or emits input', () => {
  const s = setup(); const ed = s.element('textarea', 'stale default', s.main); ed.id = 'ed';
  ed.value = 'İ 😀 needle\nneedle'; let inputs = 0; ed.addEventListener('input', () => inputs++);
  s.api.open(); s.type('needle'); assert.equal(s.count(), '1 of 2');
  assert.equal(ed.selectionStart, 5); assert.equal(ed.selectionEnd, 11);
  assert.equal(s.doc.activeElement, s.input());
  s.key('Enter'); assert.equal(ed.selectionStart, 12); assert.equal(s.doc.activeElement, ed);
  assert.equal(ed.value, 'İ 😀 needle\nneedle'); assert.equal(inputs, 0);
  s.key('F3', { shiftKey: true }); assert.equal(ed.selectionStart, 5);
});

test('fallback uses only the active DOM selection and clears owned selections on close', () => {
  const s = setup(false); const p = s.element('p', 'one one', s.main); s.api.open(); s.type('one');
  assert.equal(s.selection.rangeCount, 1); assert.equal(s.selection.getRangeAt(0).startOffset, 0);
  s.key('F3'); assert.equal(s.selection.getRangeAt(0).startOffset, 4);
  s.api.close(); assert.equal(s.selection.rangeCount, 0);
  s.api.open(); s.type('one'); s.select(p.firstChild, 1, 2); s.api.close();
  assert.equal(s.selection.rangeCount, 1, 'do not clear a newer user selection');
});

test('refresh immediately clears stale ranges/count, and never scrolls or steals focus', () => {
  const s = setup(); const old = s.element('p', 'hello', s.main); s.api.open(); s.type('hello');
  old.remove(); s.element('p', 'hello hello', s.main);
  const focusCount = s.focusCalls.length, scrollCount = s.scrolls.length;
  s.api.refresh(); assert.equal(s.highlights.size, 0); assert.equal(s.count(), 'Searching…');
  s.tick(); assert.equal(s.count(), '2'); assert.equal(s.focusCalls.length, focusCount); assert.equal(s.scrolls.length, scrollCount);
  s.type('hello', false); s.api.refresh(); s.tick(); assert.equal(s.scrolls.length, scrollCount, 'navigation cancels pending explicit reveal');
});

test('live updates and frequent afterRender hooks have a fixed deadline; observer is main-only', () => {
  const s = setup(); const p = s.element('p', 'log', s.main); s.api.open(); s.type('log');
  assert.equal(s.observers[0].root, s.main); assert.equal(s.observers[0].opts.subtree, true);
  p.firstChild.nodeValue = 'log log'; s.mutate();
  for (let i = 0; i < 9; i++) { s.tick(10); s.mutate(); s.api.refresh(); }
  assert.equal(s.count(), 'Searching…'); s.tick(10); assert.equal(s.count(), '2');
  s.api.close(); assert.equal(s.main.events.input, undefined); assert.equal(s.tasks.size, 0);
});

test('root replacement rebinds and disconnected prior focus falls back to main', () => {
  const s = setup(); const prior = s.element('button', 'old', s.main); prior.focus(); s.api.open(); s.type('new');
  s.main.remove(); const replacement = s.element('main', 'new'); replacement.id = 'main';
  const before = s.focusCalls.length; s.api.refresh(); s.tick();
  assert.equal(s.count(), '1'); assert.equal(s.focusCalls.length, before);
  assert.equal(s.observers.at(-1).root, replacement); assert.equal(s.observers[0].connected, false);
  s.api.close(); assert.equal(s.doc.activeElement, replacement);
});

test('sibling modal guard closes without stealing modal focus; Escape race leaves dialog intact', () => {
  const s = setup(); s.api.open(); const dialog = s.element('div', 'dialog', s.modal); dialog.focus();
  const event = s.key('Escape'); assert.ok(event.stopped); assert.equal(s.doc.activeElement, dialog); assert.ok(dialog.isConnected);
  dialog.remove(); s.api.open(); const palette = s.element('div'); palette.id = 'pal'; palette.focus();
  s.tick(250); assert.equal(s.doc.getElementById('np-find'), null); assert.equal(s.doc.activeElement, palette);
  assert.equal(s.tasks.size, 0);
});

test('empty/no matches and composition lifecycle clear results safely', () => {
  const s = setup(); s.element('p', 'hello', s.main); s.api.open(); s.type('missing'); assert.equal(s.count(), '0 matches');
  assert.equal(s.doc.getElementById('np-find').childNodes[2].disabled, true);
  s.type('hello'); s.type(''); assert.equal(s.count(), 'Type to find'); assert.equal(s.highlights.size, 0);
  s.input().value = 'hello'; s.input().emit('input', { isComposing: true }); s.tick(); assert.equal(s.ranges().length, 0);
  s.input().emit('compositionend'); s.tick(); assert.equal(s.ranges().length, 1);
});

test('all work limits report honestly and cap matches, scanned characters and visited nodes', () => {
  const s = setup(); s.element('p', 'x '.repeat(2000), s.main); s.api.open(); s.type('x');
  assert.equal(s.ranges().length, 1000); assert.match(s.count(), /limited/);
  const chars = setup(); chars.element('p', 'a'.repeat(1000000) + 'needle', chars.main); chars.api.open(); chars.type('needle');
  assert.equal(chars.count(), '0 matches (limited)');
  const nodes = setup();
  for (let i = 0; i < 20010; i++) nodes.element('span', '', nodes.main);
  nodes.element('p', 'needle', nodes.main); nodes.api.open(); nodes.type('needle');
  assert.equal(nodes.count(), '0 matches (limited)'); assert.ok(nodes.measurements() < 65000);
});

test('assets use themed fixed placement and do not persist, inject HTML, or replace workspace DOM', () => {
  const css = fs.readFileSync(path.join(__dirname, '../static/css/find.css'), 'utf8');
  assert.match(css, /position: fixed/); assert.match(css, /var\(--card-bg-solid/); assert.match(css, /::highlight\(np-find-active\)/);
  assert.doesNotMatch(source, /innerHTML|outerHTML|replaceChildren|localStorage|sessionStorage|fetch\(|XMLHttpRequest/);
});

test('repeated Enter on a found editor range navigates instead of inserting a newline', () => {
  const s = setup(); const ed = s.element('textarea', '', s.main); ed.id = 'ed'; ed.value = 'one one';
  s.api.open(); s.type('one'); s.key('Enter');
  assert.equal(s.doc.activeElement, ed); assert.equal(ed.selectionStart, 4);
  assert.ok(s.key('Enter').prevented); assert.equal(ed.selectionStart, 0);
  assert.ok(s.key('Enter', { shiftKey: true }).prevented); assert.equal(ed.selectionStart, 4);
  ed.setSelectionRange(2, 2); assert.ok(!s.key('Enter').prevented, 'manual caret movement returns editor Enter to editing');
});

test('rendered collapsed whitespace spans inline nodes without corrupting Unicode or source offsets', () => {
  const s = setup(), p = s.element('p', 'İ 😀 Hello\n  ', s.main), b = s.element('b', '  WORLD', p);
  const pre = s.element('pre', 'Hello\n  WORLD', s.main); pre.style.whiteSpace = 'pre';
  s.api.open(); s.type('hello world'); assert.equal(s.ranges().length, 1);
  const range = s.ranges()[0]; assert.equal(range.startContainer, p.firstChild); assert.equal(range.startOffset, 5);
  assert.equal(range.endContainer, b.firstChild); assert.equal(range.endOffset, 7);
  s.type('hello  world'); assert.equal(s.count(), '0 matches');
});
