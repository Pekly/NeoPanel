const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/file-drag.js'), 'utf8');
const TYPE = 'application/x-neopanel-files';
const tick = () => new Promise(resolve => setImmediate(resolve));

// Small DOM/event harness: no production internals exported just for tests.
class Element {
  constructor(tag = 'div') {
    this.tag = tag; this.dataset = {}; this.children = []; this.parentElement = null;
    this.style = {}; this.attrs = {}; this.className = ''; this.textContent = '';
    this.classList = {
      contains: name => this.className.split(' ').includes(name),
      add: name => { if (!this.classList.contains(name)) this.className = `${this.className} ${name}`.trim(); },
      remove: name => { this.className = this.className.split(' ').filter(x => x !== name).join(' '); }
    };
  }
  get isConnected() { return this.tag === 'body' || Boolean(this.parentElement?.isConnected); }
  matches(selector) {
    return selector.split(',').some(raw => {
      const s = raw.trim();
      if (s.startsWith('#')) return this.id === s.slice(1);
      if (s.startsWith('.')) return this.classList.contains(s.slice(1));
      if (s === '[data-file-drag][data-path]') return 'fileDrag' in this.dataset && 'path' in this.dataset;
      if (s === '[data-drop-path]') return 'dropPath' in this.dataset;
      if (s === '[aria-modal="true"]') return this.attrs['aria-modal'] === 'true';
      if (s === '[role="button"]') return this.attrs.role === 'button';
      if (s.startsWith('[contenteditable]')) return 'contenteditable' in this.attrs && this.attrs.contenteditable !== 'false';
      if (s === 'dialog[open]') return this.tag === 'dialog' && 'open' in this.attrs;
      return this.tag === s;
    });
  }
  closest(s) { return this.matches(s) ? this : this.parentElement?.closest(s) || null; }
  contains(el) { return el === this || this.children.some(child => child.contains(el)); }
  querySelectorAll(s) { return this.children.flatMap(child => [...(child.matches(s) ? [child] : []), ...child.querySelectorAll(s)]); }
  querySelector(s) { return this.querySelectorAll(s)[0] || null; }
  setAttribute(k, v) { this.attrs[k] = v; }
  appendChild(child) { child.remove(); this.children.push(child); child.parentElement = this; return child; }
  remove() { if (this.parentElement) this.parentElement.children = this.parentElement.children.filter(x => x !== this); this.parentElement = null; }
  before(child) { child.remove(); const p = this.parentElement; p.children.splice(p.children.indexOf(this), 0, child); child.parentElement = p; }
  replaceWith(child) { this.before(child); this.remove(); }
  replaceChildren(...children) { [...this.children].forEach(child => child.remove()); children.forEach(child => this.appendChild(child)); }
}
function transfer(types = [], files = [], items = []) {
  return { types: [...types], files, items, data: {}, setData(k, v) { this.data[k] = v; if (!this.types.includes(k)) this.types.push(k); }, getData() { throw Error('Untrusted data must never be read'); } };
}
function setup() {
  const body = new Element('body'), main = body.appendChild(new Element()), overlay = body.appendChild(new Element());
  main.id = 'main'; overlay.id = 'dropzone-overlay';
  const listeners = new Map(), timeouts = new Map(), intervals = new Map(), requests = [], loads = [], uploads = [], notices = [];
  let sequence = 0;
  const on = (type, fn) => { if (!listeners.has(type)) listeners.set(type, []); listeners.get(type).push(fn); };
  const document = {
    body, addEventListener: on, querySelector: s => body.querySelector(s), querySelectorAll: s => body.querySelectorAll(s),
    createElement: tag => new Element(tag), createElementNS: (_, tag) => new Element(tag), createComment: () => new Element('comment'),
    elementFromPoint: () => document.hit || null
  };
  const c = vm.createContext({ document, S: { view: 'fi', token: 'token', user: { id: 1, username: 'one' }, currentDir: 'src', selectedFiles: new Set() },
    server: { id: 'server', node_id: 'node' }, denied: new Set(), preview: false,
    setTimeout: (fn, ms) => { const id = ++sequence; timeouts.set(id, { fn, ms }); return id; }, clearTimeout: id => timeouts.delete(id),
    setInterval: fn => { const id = ++sequence; intervals.set(id, fn); return id; }, clearInterval: id => intervals.delete(id),
    toast: (...args) => notices.push(args), syncFileSelectionUI: () => {},
    srvApiPath: (server, suffix) => `/nodes/${server.node_id}/servers/${server.id}${suffix}`,
    api: async (url, options) => { requests.push([url, options]); },
    enqueueUploads: async (files, dir) => { uploads.push([Array.from(files), dir]); },
    sameQolServer: (a, b) => a.id === b.id && a.node_id === b.node_id
  });
  c.window = c; c.addEventListener = on; c.curServer = () => c.server;
  c.hasServerPermission = name => !c.denied.has(name); c.filePreviewActive = () => c.preview;
  function workspace() { const node = main.appendChild(new Element()); node.className = 'np-files-workspace'; return node; }
  let area = workspace();
  function row(p, dir = false) { const node = area.appendChild(new Element()); node.dataset = { fileDrag: '', path: p, isDir: String(dir) }; if (dir) node.dataset.dropPath = p; return node; }
  function folder(p) { const node = main.appendChild(new Element('button')); node.dataset.dropPath = p; return node; }
  const sourceRow = row('src/a.txt'), destination = folder('dest');
  c.loadFileList = async (dir, skip, guard) => {
    loads.push([dir, guard]);
    if (!guard.isCurrent()) return;
    main.replaceChildren(); area = workspace(); c.S.currentDir = dir; c.S.selectedFiles.clear(); c.NPFileDrag.afterRender(main);
  };
  vm.runInContext(source, c); c.NPFileDrag.init();
  function emit(type, target = destination, dt = transfer(), extra = {}) {
    const event = { type, target, dataTransfer: dt, relatedTarget: null, clientX: 10, clientY: 10, defaultPrevented: false,
      preventDefault() { this.defaultPrevented = true; }, ...extra };
    for (const listener of listeners.get(type) || []) listener(event);
    return event;
  }
  async function timers() { const pending = [...timeouts.values()]; timeouts.clear(); for (const entry of pending) entry.fn(); await tick(); }
  return { c, body, main, overlay, document, sourceRow, destination, requests, loads, uploads, notices, listeners, timeouts, intervals,
    row, folder, emit, timers, get area() { return area; }, start() { const dt = transfer(); emit('dragstart', sourceRow, dt); return dt; } };
}

test('init is idempotent; only marked rows decorate; controls cannot start a native row drag', () => {
  const h = setup(); h.c.NPFileDrag.init(); assert.equal(h.listeners.get('drop').length, 1); assert.equal(h.sourceRow.draggable, true);
  for (const tag of ['input', 'button', 'textarea']) {
    const control = h.sourceRow.appendChild(new Element(tag)); h.emit('pointerdown', control);
    const dt = transfer(); assert.equal(h.emit('dragstart', h.sourceRow, dt).defaultPrevented, true); assert.deepEqual(dt.types, []);
  }
  const editable = h.sourceRow.appendChild(new Element()); editable.attrs.contenteditable = '';
  h.emit('pointerdown', editable); assert.equal(h.emit('dragstart', h.sourceRow).defaultPrevented, true);
});

test('hover parks ORIGINAL source outside main; captures multiselection despite rerender and directory changes', async () => {
  const h = setup(); h.c.S.selectedFiles = new Set(['src/a.txt', 'src/b.txt']);
  const dt = h.start(); assert.equal(dt.data[TYPE], 'internal');
  h.emit('dragover', h.destination, dt); assert.equal([...h.timeouts.values()][0].ms, 650);
  await h.timers(); assert.equal(h.loads.length, 1); assert.equal(h.sourceRow.isConnected, true);
  assert.equal(h.sourceRow.closest('#main'), null); assert.equal(h.sourceRow.draggable, true);
  assert.equal(h.c.S.selectedFiles.size, 0);
  h.emit('drop', h.area, dt); await tick();
  assert.equal(h.requests.length, 1); assert.match(h.requests[0][0], /nodes\/node\/servers\/server\/files\/move$/);
  assert.deepEqual(JSON.parse(h.requests[0][1].body), { paths: ['src/a.txt', 'src/b.txt'], destination: 'dest', overwrite: false });
  assert.equal(h.requests[0][1].method, 'POST'); assert.equal(h.sourceRow.isConnected, false);
  assert.equal(h.body.querySelector('.np-file-drag-parking'), null); assert.equal(h.overlay.style.display, 'none');
});

test('unselected row moves only itself; payload paths are never deserialized', async () => {
  const h = setup(); h.c.S.selectedFiles.add('src/other');
  h.emit('drop', h.destination, transfer([TYPE])); await tick(); assert.equal(h.requests.length, 0);
  const dt = h.start(); dt.data[TYPE] = '["secret"]'; h.emit('drop', h.destination, dt); await tick();
  assert.deepEqual(JSON.parse(h.requests[0][1].body).paths, ['src/a.txt']);
});

test('self/descendant/same-parent moves, traversal and over-1000 selection are rejected', async () => {
  for (const dest of ['src/a.txt', 'src/a.txt/child', 'src', '../outside', '/absolute']) {
    const h = setup(), dt = h.start(); h.destination.dataset.dropPath = dest;
    h.emit('drop', h.destination, dt); await tick(); assert.equal(h.requests.length, 0, dest);
  }
  const h = setup(); h.c.S.selectedFiles = new Set(['src/a.txt', ...Array.from({ length: 1000 }, (_, i) => `src/${i}`)]);
  assert.equal(h.emit('dragstart', h.sourceRow).defaultPrevented, true); assert.match(h.notices[0][0], /1000/);
});

test('view, editor, login, modal, palette, preview, trash and permissions fail closed', async () => {
  const blockers = [h => { h.c.S.view = 'ov'; }, h => { h.c.S.fileTrashOpen = true; }, h => { h.c.preview = true; },
    ...['ed', 'login', 'pal'].map(id => h => { const n = h.body.appendChild(new Element()); n.id = id; }),
    h => { h.body.appendChild(new Element()).attrs['aria-modal'] = 'true'; },
    ...['files.view', 'files.edit', 'files.delete'].map(permission => h => h.c.denied.add(permission))];
  for (const block of blockers) {
    const h = setup(); block(h); assert.equal(h.emit('dragstart', h.sourceRow).defaultPrevented, true);
    h.emit('drop', h.destination, transfer([TYPE])); await tick(); assert.equal(h.requests.length, 0);
  }
  const h = setup(); h.c.denied.add('files.delete');
  h.emit('drop', h.destination, transfer(['Files'], [{ name: 'allowed.txt' }])); await tick(); assert.equal(h.uploads.length, 1);
});

test('session/server/user/permission changes cancel hover and reject internal drops', async () => {
  for (const mutate of [h => { h.c.S.token = 'new'; }, h => { h.c.S.user.id = 2; }, h => { h.c.server.node_id = 'other'; }, h => h.c.denied.add('files.delete')]) {
    const h = setup(), dt = h.start(); h.emit('dragover', h.destination, dt); mutate(h);
    for (const fn of h.intervals.values()) fn(); await h.timers();
    h.emit('drop', h.destination, dt); await tick(); assert.equal(h.loads.length, 0); assert.equal(h.requests.length, 0);
    assert.equal(h.sourceRow.classList.contains('np-file-dragging'), false);
  }
});

test('hover does not spam while stationary, and cancel makes late navigation guard false', async () => {
  const h = setup(); let finish, guard;
  h.c.loadFileList = async (dir, skip, options) => { h.loads.push(dir); guard = options; await new Promise(resolve => { finish = resolve; }); if (guard.isCurrent()) h.c.S.currentDir = dir; };
  const dt = h.start(); for (let i = 0; i < 20; i++) h.emit('dragover', h.destination, dt);
  assert.equal(h.timeouts.size, 1); await h.timers();
  for (let i = 0; i < 20; i++) h.emit('dragover', h.destination, dt);
  await h.timers(); assert.equal(h.loads.length, 1); assert.equal(guard.isCurrent(), true);
  h.c.NPFileDrag.cancel(); assert.equal(guard.isCurrent(), false); finish(); await tick();
  assert.equal(h.c.S.currentDir, 'src'); assert.equal(h.sourceRow.parentElement, h.area); assert.equal(h.body.querySelector('.np-file-drag-parking'), null);
});

test('unrelated navigation cleans parked sources and invalidates late hover', async () => {
  const h = setup(); let guard;
  h.c.loadFileList = (dir, skip, options) => { guard = options; return new Promise(() => {}); };
  h.emit('dragover', h.destination, h.start()); await h.timers();
  h.c.S.currentDir = 'unrelated'; h.c.NPFileDrag.afterRender(h.main);
  assert.equal(guard.isCurrent(), false); assert.equal(h.body.querySelector('.np-file-drag-parking'), null);
});

test('external Files hover without files; explicit upload destination and text-only hint', async () => {
  const h = setup(), dt = transfer(['Files']); h.destination.dataset.dropPath = '<img onerror=bad>';
  h.emit('dragover', h.destination, dt); assert.equal(h.overlay.style.pointerEvents, 'none');
  assert.match(h.overlay.querySelector('.np-file-drop-label').textContent, /<img onerror=bad>/);
  assert.equal(h.overlay.querySelector('img'), null); await h.timers();
  assert.equal(h.c.S.currentDir, '<img onerror=bad>');
  dt.files = [{ name: '<script>.txt' }]; h.emit('drop', h.area, dt); await tick();
  assert.deepEqual(h.uploads[0], [[{ name: '<script>.txt' }], '<img onerror=bad>']);
});

test('text drags are untouched; OS files outside workspace cannot navigate or upload', async () => {
  const h = setup();
  for (const type of ['dragenter', 'dragover', 'drop']) assert.equal(h.emit(type, h.area, transfer(['text/plain'])).defaultPrevented, false);
  const dt = transfer(['Files'], [{ name: 'a.txt' }]);
  assert.equal(h.emit('dragover', h.body, dt).defaultPrevented, true);
  assert.equal(h.emit('drop', h.body, dt).defaultPrevented, true); await tick(); assert.equal(h.uploads.length, 0);
});

test('OS directory entries and filesystem handles are explicitly rejected, not uploaded as placeholders', async () => {
  for (const item of [{ kind: 'file', webkitGetAsEntry: () => ({ isDirectory: true }) },
    { kind: 'file', getAsFileSystemHandle: async () => ({ kind: 'directory' }) }]) {
    const h = setup(); h.emit('drop', h.destination, transfer(['Files'], [{ name: 'folder', size: 0 }], [item])); await tick();
    assert.equal(h.uploads.length, 0); assert.match(h.notices[0][0], /Folder uploads are not supported/);
  }
});

test('external drag is not recaptured after session cancellation or during async directory checks', async () => {
  const h = setup(), dt = transfer(['Files'], [{ name: 'a' }]);
  h.emit('dragover', h.destination, dt); h.c.S.token = 'changed';
  for (const fn of h.intervals.values()) fn();
  h.emit('dragover', h.destination, dt); h.emit('drop', h.destination, dt); await tick(); assert.equal(h.uploads.length, 0);
  let resolve;
  dt.items = [{ kind: 'file', getAsFileSystemHandle: () => new Promise(r => { resolve = r; }) }];
  h.emit('drop', h.destination, dt); h.c.NPFileDrag.cancel(); resolve({ kind: 'file' }); await tick(); assert.equal(h.uploads.length, 0);
});

test('single in-flight lock survives cancel; stale completion cannot refresh or change selection', async () => {
  const h = setup(); let finish;
  h.c.api = (url, options) => { h.requests.push([url, options]); return new Promise(resolve => { finish = resolve; }); };
  const dt = h.start(); h.emit('drop', h.destination, dt);
  assert.equal(h.body.classList.contains('np-file-drag-pending'), true);
  h.c.NPFileDrag.cancel(); assert.equal(h.emit('dragstart', h.sourceRow).defaultPrevented, true);
  h.emit('drop', h.destination, dt); h.c.S.selectedFiles.add('src/a.txt'); finish(); await tick();
  assert.equal(h.requests.length, 1); assert.equal(h.loads.length, 0); assert.equal(h.notices.length, 0);
  assert.equal(h.c.S.selectedFiles.has('src/a.txt'), true); assert.equal(h.body.classList.contains('np-file-drag-pending'), false);
});

test('child dragleave does not flicker or cancel hover; escape, dragend and pagehide clean all transient UI', async () => {
  for (const end of ['keydown', 'dragend', 'pagehide']) {
    const h = setup(), dt = h.start(); h.emit('dragover', h.destination, dt);
    const child = h.destination.appendChild(new Element('span')); h.document.hit = child;
    for (let i = 0; i < 10; i++) h.emit('dragleave', child, dt, { relatedTarget: child });
    assert.equal(h.timeouts.size, 1); assert.equal(h.overlay.style.display, 'flex');
    h.emit(end, h.destination, dt, { key: 'Escape' }); await h.timers();
    assert.equal(h.loads.length, 0); assert.equal(h.overlay.style.display, 'none'); assert.equal(h.intervals.size, 0);
    assert.equal(h.destination.classList.contains('np-file-drop-target'), false);
  }
});

test('file-name buttons can start a drag without enabling action-button drags', async () => {
  const h = setup();
  const label = h.sourceRow.appendChild(new Element('button')); label.className = 'np-file-open';
  h.c.NPFileDrag.afterRender(h.main); assert.equal(label.draggable, true);
  h.emit('pointerdown', label); const dt = transfer();
  assert.equal(h.emit('dragstart', label, dt).defaultPrevented, false);
  h.emit('drop', h.destination, dt); await tick();
  assert.equal(h.requests.length, 1);
});
