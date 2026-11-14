const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto');
const read = name => fs.readFileSync(path.join(__dirname, '../static', name), 'utf8');
const source = read('js/editor-syntax.js');
const workerSource = read('js/editor-syntax-worker.js');
const vendor = read('vendor/prism-1.30.0/prism.js');

class Node {
  constructor(tag, doc) {
    this.tagName = tag.toUpperCase(); this.doc = doc; this.children = []; this.parentNode = null;
    this.attrs = new Map(); this.listeners = new Map(); this.className = ''; this.style = {};
    this.scrollTop = 0; this.scrollLeft = 0; this.clientWidth = 620; this.clientHeight = 360;
    this.value = ''; this.selectionStart = 2; this.selectionEnd = 5; this.selectionDirection = 'backward';
    this.classList = {
      add: name => { this.className = [...new Set([...this.className.split(' ').filter(Boolean), name])].join(' '); },
      remove: name => { this.className = this.className.split(' ').filter(n => n !== name).join(' '); },
      contains: name => this.className.split(' ').includes(name)
    };
  }
  set innerHTML(_) { throw Error('HTML sink forbidden'); }
  set textContent(value) { this.children.forEach(n => { n.parentNode = null; }); this.children = []; this.text = String(value); }
  get textContent() { return (this.text || '') + this.children.map(n => n.textContent).join(''); }
  get nextSibling() { return this.parentNode?.children[this.parentNode.children.indexOf(this) + 1] || null; }
  setAttribute(name, value) { this.attrs.set(name, String(value)); }
  getAttribute(name) { return this.attrs.get(name) ?? null; }
  removeAttribute(name) { this.attrs.delete(name); }
  appendChild(child) { return this.insertBefore(child, null); }
  insertBefore(child, next) {
    if (child.tagName === '#FRAGMENT') { [...child.children].forEach(c => this.insertBefore(c, next)); return child; }
    child.remove(); const index = next ? this.children.indexOf(next) : this.children.length;
    assert.ok(index >= 0); this.children.splice(index, 0, child); child.parentNode = this; return child;
  }
  remove() { if (this.parentNode) { const p = this.parentNode; p.children.splice(p.children.indexOf(this), 1); this.parentNode = null; } }
  replaceChildren(...children) { this.textContent = ''; children.forEach(n => this.appendChild(n)); }
  addEventListener(name, fn) { if (!this.listeners.has(name)) this.listeners.set(name, new Set()); this.listeners.get(name).add(fn); }
  removeEventListener(name, fn) { this.listeners.get(name)?.delete(fn); }
  emit(name) { for (const fn of this.listeners.get(name) || []) fn({ type: name }); }
  focus() { this.doc.activeElement = this; }
}
function setup({ text = '{"x": 1}\n\n', filename = 'test.json', forced = false, styled = true, noWorker = false } = {}) {
  const timers = new Map(), workers = [], observers = [];
  let now = 0, serial = 0;
  const document = { baseURI: 'https://panel.test/', currentScript: { src: 'https://panel.test/static/js/editor-syntax.js?v=1' } };
  document.createElement = tag => new Node(tag, document);
  document.createDocumentFragment = () => new Node('#fragment', document);
  document.createTextNode = value => { const n = new Node('#text', document); n.textContent = value; return n; };
  const media = new Node('media', document); media.matches = forced;
  class Worker {
    constructor(url) { this.url = url; this.jobs = []; workers.push(this); }
    postMessage(job) { this.jobs.push(job); }
    terminate() { this.terminated = true; }
    reply(job = this.jobs.at(-1), data = {}) { this.onmessage?.({ data: { generation: job.generation, version: job.version, runs: job.text ? [[job.text.length, 'string']] : [], ...data } }); }
  }
  const c = vm.createContext({ document, URL, Worker: noWorker ? undefined : Worker,
    getComputedStyle: () => ({ getPropertyValue: () => styled ? '1' : '' }),
    matchMedia: () => media,
    ResizeObserver: class { constructor(fn) { this.fn = fn; observers.push(this); } observe(ed) { this.target = ed; } disconnect() { this.disconnected = true; } },
    setTimeout: (fn, delay) => { const id = ++serial; timers.set(id, { at: now + delay, fn }); return id; },
    clearTimeout: id => timers.delete(id)
  });
  c.window = c;
  vm.runInContext(source, c);
  const parent = document.createElement('section'), ed = document.createElement('textarea');
  ed.value = text; ed.setAttribute('id', 'ed'); ed.setAttribute('wrap', 'soft'); ed.setAttribute('aria-label', 'File editor');
  ed.scrollTop = 51; ed.scrollLeft = 27; parent.appendChild(ed); ed.focus();
  let originalInputs = 0; ed.addEventListener('input', () => originalInputs++);
  const api = c.NPEditorSyntax, controller = api.bind(ed, filename);
  const wrapper = ed.parentNode, mirror = wrapper.children[0], status = parent.children[1];
  function advance(ms) {
    const target = now + ms;
    while (true) {
      const next = [...timers.entries()].filter(([, t]) => t.at <= target).sort((a, b) => a[1].at - b[1].at)[0];
      if (!next) break;
      now = next[1].at; timers.delete(next[0]); next[1].fn();
    }
    now = target;
  }
  return { api, controller, ed, parent, wrapper, mirror, status, workers, timers, observers, media, document, advance,
    originalInputs: () => originalInputs };
}
function workerVM() {
  const replies = [], imports = [];
  const c = vm.createContext({
    WorkerGlobalScope: class { static [Symbol.hasInstance]() { return true; } },
    postMessage: result => replies.push(result)
  });
  c.self = c;
  c.importScripts = url => { imports.push(url); assert.equal(url, '../vendor/prism-1.30.0/prism.js'); vm.runInContext(vendor, c, { timeout: 2000 }); };
  vm.runInContext(workerSource, c, { timeout: 2000 });
  return { c, replies, imports, run(language, text) {
    c.input = { data: { generation: 7, version: 3, language, text } };
    vm.runInContext('self.onmessage(input)', c, { timeout: 2000 });
    return replies.at(-1);
  } };
}

test('extension mapping is case-insensitive, path-aware, conservative, and prototype-safe', () => {
  const { api } = setup();
  const cases = {
    '/plugins/Plugin/CONFIG.YML': 'yaml', 'C:\\server\\paper.yaml': 'yaml', 'a.json': 'json', 'a.toml': 'toml',
    'a.ini': 'ini', 'a.conf': 'ini', 'a.cfg': 'ini', 'server.properties': 'properties',
    'en_us.lang': 'properties', 'pack.mcmeta': 'json',
    'a.js': 'javascript', 'a.mjs': 'javascript', 'a.cjs': 'javascript', 'a.ts': 'typescript', 'a.cts': 'typescript',
    'a.java': 'java', 'a.py': 'python', 'a.sh': 'bash', '.bashrc': 'bash', 'a.ps1': 'powershell',
    'a.psm1': 'powershell', 'a.sql': 'sql', 'a.xml': 'markup', 'a.html': 'markup', 'a.css': 'css',
    'README.md': 'markdown', 'a.go': 'go', 'a.rs': 'rust', 'a.lua': 'lua', 'Dockerfile': 'docker',
    'Dockerfile.prod': 'docker', 'prod.dockerfile': 'docker', '.env.local': 'properties',
    'latest.log': 'plain', 'a.txt': 'plain', 'LICENSE': 'plain', 'archive.jar': 'plain',
    'a.__proto__': 'plain', 'a.constructor': 'plain', 'dir.json/unknown': 'plain'
  };
  for (const [name, expected] of Object.entries(cases)) assert.equal(api.languageFor(name), expected, name);
});

test('official pinned vendor integrity, component hashes, and MIT license', () => {
  const manifest = JSON.parse(read('vendor/prism-1.30.0/SOURCES.json').replace(/^\uFEFF/, ''));
  const hash = text => crypto.createHash('sha256').update(text).digest('hex');
  assert.equal(manifest.version, '1.30.0'); assert.equal(manifest.sources.length, 21);
  assert.equal(hash(vendor), manifest.bundleSHA256);
  const license = read('vendor/prism-1.30.0/LICENSE');
  assert.equal(hash(license), manifest.licenseSHA256); assert.match(license, /MIT LICENSE[\s\S]*Permission is hereby granted/);
  for (let i = 0; i < manifest.sources.length; i++) {
    const entry = manifest.sources[i], marker = `\n/* ${entry.path} (unmodified) */\n`;
    const start = vendor.indexOf(marker) + marker.length;
    const next = manifest.sources[i + 1];
    const end = next ? vendor.indexOf(`\n/* ${next.path} (unmodified) */\n`) : vendor.length;
    assert.equal(hash(vendor.slice(start, end - 1)), entry.sha256, entry.path);
  }
});

test('actual worker tokenizes all bundled languages into bounded offsets, never HTML', () => {
  const worker = workerVM();
  const samples = {
    json: '{"motd":"<img src=x onerror=alert(1)>","port":25565}',
    yaml: '# config\nmotd: "hello"\nport: 25565\n', toml: '[server]\nport = 25565',
    ini: '[server]\nport=25565', properties: '# Minecraft\nmotd=Hello\nonline-mode=true',
    javascript: 'const x = "</span><script>alert(1)</script>";', typescript: 'const x: number = 1;',
    java: 'public class Server { int port = 25565; }', python: 'def start():\n  return "hello"',
    bash: '#!/bin/bash\necho "$HOME"', powershell: '$name = "server"; Write-Host $name',
    sql: "SELECT * FROM users WHERE name = 'Steve';", markup: '<img src="x" onerror="alert(1)">',
    css: 'body { color: red; }', markdown: '# Title\n**hello** [link](javascript:alert)',
    go: 'package main\nfunc main() { return }', rust: 'fn main() { let n = 1; }',
    lua: 'local x = "hello" -- comment', docker: 'FROM alpine:3.20\nRUN echo hello'
  };
  for (const [language, text] of Object.entries(samples)) {
    const result = worker.run(language, text);
    assert.equal(result.error, undefined, language); assert.equal(result.generation, 7); assert.equal(result.version, 3);
    assert.equal(result.runs.at(-1)[0], text.length, language);
    assert.ok(result.runs.some(run => run[1]), language);
    assert.ok(result.runs.length <= 6000); assert.equal(result.html, undefined);
    let start = 0, rebuilt = '';
    for (const [end, kind] of result.runs) { assert.ok(end > start); assert.match(kind, /^[a-z]*$/); rebuilt += text.slice(start, end); start = end; }
    assert.equal(rebuilt, text);
  }
  assert.equal(worker.imports.length, 1);
  assert.equal(worker.run('json', 'x'.repeat(200001)).error, 'limit');
  assert.equal(worker.run('__proto__', 'x').error, 'limit');
  assert.equal(worker.run('json', '[' + '1,'.repeat(4000) + '0]').error, 'complexity');
  assert.equal(worker.run('json', '').runs.length, 0);
});

test('bind preserves native textarea, input listener, focus, selection and scroll; safe mirror handles trailing lines', () => {
  const text = '<img src=x onerror=alert(1)>\n\t</pre><script>evil()</script>\n\n';
  const s = setup({ text, filename: 'a.html' });
  assert.equal(s.ed.getAttribute('wrap'), 'off'); assert.equal(s.ed.getAttribute('aria-label'), 'File editor');
  assert.equal(s.document.activeElement, s.ed); assert.equal(s.mirror.getAttribute('aria-hidden'), 'true');
  assert.equal(s.mirror.getAttribute('contenteditable'), null); assert.equal(s.status.getAttribute('role'), 'status');
  assert.equal(s.mirror.style.width, '620px'); assert.equal(s.mirror.style.height, '360px');
  s.advance(119); assert.equal(s.workers.length, 0); s.advance(1);
  assert.equal(s.workers[0].url, 'https://panel.test/static/js/editor-syntax-worker.js');
  s.workers[0].reply();
  assert.equal(s.mirror.textContent, text + '\u200b');
  assert.deepEqual(s.mirror.children.map(n => n.tagName), ['SPAN', '#TEXT']);
  assert.ok(s.wrapper.classList.contains('np-syntax-colored'));
  assert.equal(s.ed.value, text); assert.equal(s.ed.selectionStart, 2); assert.equal(s.ed.selectionEnd, 5);
  assert.equal(s.ed.selectionDirection, 'backward'); assert.equal(s.ed.scrollTop, 51); assert.equal(s.ed.scrollLeft, 27);
  s.ed.scrollTop = 120; s.ed.scrollLeft = 200; s.ed.emit('scroll');
  assert.equal(s.mirror.scrollTop, 120); assert.equal(s.mirror.scrollLeft, 200);
  s.ed.clientWidth = 500; s.ed.clientHeight = 200; s.observers[0].fn();
  assert.equal(s.mirror.style.width, '500px'); assert.equal(s.mirror.style.height, '200px');
  s.ed.emit('input'); assert.equal(s.originalInputs(), 1); assert.equal(s.mirror.textContent, '');
  assert.ok(!s.wrapper.classList.contains('np-syntax-colored'));
});

test('typing debounce, stale versions/generations, and single-flight queue', () => {
  const s = setup(); s.advance(120); const w = s.workers[0], old = w.jobs[0];
  for (let i = 0; i < 8; i++) { s.ed.value = String(i); s.ed.emit('input'); s.advance(50); }
  assert.equal(w.jobs.length, 1); s.advance(120);
  w.reply(old, { generation: old.generation - 1 }); assert.equal(w.jobs.length, 1);
  w.reply(old); assert.equal(s.mirror.textContent, ''); assert.equal(w.jobs.length, 2);
  w.reply(); assert.equal(s.mirror.textContent, '7\u200b');
  s.ed.value = 'programmatic'; s.api.refresh(); assert.equal(s.mirror.textContent, ''); s.advance(120); w.reply();
  assert.equal(s.mirror.textContent, 'programmatic\u200b');
});

test('worker watchdog is not extended by typing; timeout terminates with visible native fallback', () => {
  const s = setup(); s.advance(120); const w = s.workers[0];
  s.advance(1500); s.ed.emit('input'); s.advance(500);
  assert.equal(w.terminated, true); assert.match(s.status.textContent, /timed out/);
  assert.equal(s.mirror.textContent, ''); assert.equal(s.timers.size, 0);
  s.api.refresh(); s.advance(5000); assert.equal(s.workers.length, 1);
});

test('large files never reach a worker; crossing the cap stops inflight work and shrinking can recover', () => {
  const large = setup({ text: 'x'.repeat(200001) }); large.advance(5000);
  assert.equal(large.workers.length, 0); assert.match(large.status.textContent, /disabled above 200,000/);
  large.ed.value = '{}'; large.api.refresh(); large.advance(120); assert.equal(large.workers.length, 1);
  large.ed.value = 'x'.repeat(200001); large.ed.emit('input'); assert.equal(large.workers[0].terminated, true);
  assert.equal(large.timers.size, 0);
});

test('plain text, stylesheet failure, unavailable worker, worker error and malformed replies stay readable', () => {
  for (const options of [{ filename: 'latest.log' }, { styled: false }, { forced: true }]) {
    const s = setup(options); s.advance(5000); assert.equal(s.workers.length, 0); assert.ok(!s.wrapper.classList.contains('np-syntax-colored'));
  }
  const missing = setup({ noWorker: true }); missing.advance(120); assert.match(missing.status.textContent, /unavailable/);
  const error = setup(); error.advance(120); error.workers[0].onerror(); assert.equal(error.workers[0].terminated, true);
  const invalid = [ [[1, '"><img src=x>']], [[999999, 'string']], [[1, 'string']], [[0, 'string']], 'html' ];
  for (const runs of invalid) {
    const s = setup(); s.advance(120); s.workers[0].reply(undefined, { runs });
    assert.match(s.status.textContent, /invalid coloring result/); assert.equal(s.mirror.textContent, ''); assert.equal(s.workers[0].terminated, true);
  }
});

test('IME and dynamic forced colors disable the mirror without interfering with editing', () => {
  const s = setup(); s.advance(120); const w = s.workers[0]; w.reply();
  s.ed.emit('compositionstart'); assert.equal(s.mirror.textContent, '');
  s.ed.value = '世界'; s.ed.emit('input'); s.advance(500); assert.equal(w.jobs.length, 1);
  s.ed.emit('compositionend'); s.advance(120); w.reply(); assert.equal(s.mirror.textContent, '世界\u200b');
  s.media.matches = true; s.media.emit('change'); assert.equal(w.terminated, true); assert.match(s.status.textContent, /high contrast/);
  s.media.matches = false; s.media.emit('change'); s.advance(120); assert.equal(s.workers.length, 2);
});

test('dispose/rebind releases worker, callbacks, timers, media/resize observers and DOM without removing existing events', () => {
  const s = setup(); s.advance(120); const w = s.workers[0], callback = w.onmessage, job = w.jobs[0];
  s.api.dispose(); s.controller.dispose();
  assert.equal(w.terminated, true); assert.equal(w.onmessage, null); assert.equal(s.timers.size, 0);
  assert.equal(s.observers[0].disconnected, true); assert.equal(s.media.listeners.get('change').size, 0);
  assert.equal(s.ed.parentNode, s.parent); assert.equal(s.parent.children.length, 1); assert.equal(s.ed.getAttribute('wrap'), 'soft');
  assert.equal(s.ed.className, ''); assert.equal(s.ed.listeners.get('input').size, 1); assert.equal(s.ed.listeners.get('scroll').size, 0);
  callback({ data: { ...job, runs: [[job.text.length, 'string']] } }); assert.equal(s.mirror.textContent, '');
  s.ed.emit('input'); assert.equal(s.originalInputs(), 1);
  s.api.bind(s.ed, 'new.yaml'); s.advance(120); assert.equal(s.workers.length, 2);
  assert.notEqual(s.workers[1].jobs[0].generation, job.generation);
  s.api.bind(s.ed, 'again.json'); assert.equal(s.workers[1].terminated, true); assert.equal(s.parent.children.length, 2);
  s.api.dispose(); assert.equal(s.timers.size, 0);
});

test('paired CSS fixes metrics, hides inert mirror, and provides dark/light/forced-color selection palettes', () => {
  const css = read('css/editor-syntax.css');
  assert.match(css, /#ed\.np-syntax-input,\s*\.np-syntax-editor > \.np-syntax-mirror \{/);
  for (const rule of ['white-space: pre !important', 'padding: 16px !important', 'font-size: 13.5px !important', 'line-height: 1.7 !important', 'tab-size: 2 !important', 'pointer-events: none', 'user-select: none', 'caret-color:', '::selection', 'forced-colors: active', 'html[data-theme="light"]']) assert.ok(css.includes(rule), rule);
  assert.doesNotMatch(source, /innerHTML|insertAdjacentHTML|setSelectionRange|\.value\s*=|MutationObserver|setInterval|dispatchEvent/);
  assert.doesNotMatch(workerSource, /Prism\.highlight\(|eval\(|new Function/);
});
