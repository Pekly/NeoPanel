const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
const css = fs.readFileSync(path.join(__dirname, '../static/css/style.css'), 'utf8');
function section(start, end) {
  const a = source.indexOf(start), b = source.indexOf(end, a);
  assert.ok(a >= 0 && b > a, start);
  return source.slice(a, b);
}
function setup(width = 800, height = 600) {
  const listeners = {}, events = [];
  const body = { appendChild(node) { node.parentNode = body; events.push('portal'); } };
  const menu = { style: { display: 'none' }, dataset: {}, parentNode: body,
    naturalWidth: 218, naturalHeight: 360,
    contains: target => target === menu,
    getBoundingClientRect() { events.push('measure'); return { width: Math.min(this.naturalWidth, parseFloat(this.style.maxWidth)), height: Math.min(this.naturalHeight, parseFloat(this.style.maxHeight)) }; }
  };
  const c = vm.createContext({
    S: { fileList: [{ name: 'archive.zip', path: 'archive.zip', is_dir: false }], sortedFileList: [] },
    innerWidth: width, innerHeight: height,
    document: { body, documentElement: { clientWidth: width }, addEventListener: (name, fn) => { listeners['document:' + name] = fn; } },
    addEventListener: (name, fn) => { listeners['window:' + name] = fn; },
    $: selector => selector === '#floating-ctx' ? menu : null, $$: () => [],
    esc: value => String(value ?? '').replaceAll('&', '&amp;').replaceAll('"', '&quot;').replaceAll('<', '&lt;'),
    ic: () => '', isArchiveFile: () => true, applyPermissionControls: () => events.push('permissions'),
    getBrand: () => '/brand/default.png'
  });
  c.window = c;
  vm.runInContext(section('function fileMenuViewport(', '// --- Upload with Live Progress') + section('function toggleFileCtx(', 'window.addEventListener(\'click\', closeAllCtx);') + section('function serverRowIconHtml(', 'function brandLogoHtml('), c);
  function event(x, y, rect) { return { clientX: x, clientY: y, preventDefault() {}, stopPropagation() {}, currentTarget: rect ? { getBoundingClientRect: () => rect } : null }; }
  return { c, menu, listeners, events, event, body };
}
function assertWithin(state, bounds) {
  const { menu } = state;
  const box = menu.getBoundingClientRect();
  const x = parseFloat(menu.style.left), y = parseFloat(menu.style.top);
  assert.ok(x >= bounds.left && x + box.width <= bounds.left + bounds.width, `horizontal: ${x} + ${box.width}`);
  assert.ok(y >= bounds.top && y + box.height <= bounds.top + bounds.height, `vertical: ${y} + ${box.height}`);
  assert.equal(menu.style.visibility, 'visible');
}

test('right-click menus fit all viewport corners using their actual dimensions', () => {
  for (const [x, y] of [[0, 0], [799, 0], [0, 599], [799, 599]]) {
    const state = setup();
    state.c.handleFileContextMenu(state.event(x, y), 0);
    assertWithin(state, { left: 0, top: 0, width: 800, height: 600 });
    assert.ok(state.events.indexOf('permissions') < state.events.indexOf('measure'));
  }
});

test('oversized menus are constrained to small visible viewports', () => {
  const state = setup(180, 140);
  state.c.handleFileContextMenu(state.event(179, 139), 0);
  assertWithin(state, { left: 0, top: 0, width: 180, height: 140 });
  assert.equal(state.menu.style.maxWidth, '164px');
  assert.equal(state.menu.style.maxHeight, '124px');
  assert.match(css, /\.floating-ctx-menu\s*\{[^}]*overflow-y:\s*auto/s);
});

test('zoomed/panned visual viewport offsets constrain the menu', () => {
  const state = setup();
  state.c.visualViewport = { offsetLeft: 70, offsetTop: 90, width: 250, height: 220 };
  state.c.handleFileContextMenu(state.event(799, 599), 0);
  assertWithin(state, { left: 70, top: 90, width: 250, height: 220 });
});

test('three-dot menus use the shared portal, flip above their button, and toggle closed', () => {
  const state = setup();
  const anchor = { left: 750, right: 785, top: 550, bottom: 585 };
  state.menu.parentNode = {};
  state.c.toggleFileCtx(state.event(0, 0, anchor), 0);
  assert.equal(state.menu.parentNode, state.body);
  assertWithin(state, { left: 0, top: 0, width: 800, height: 600 });
  assert.ok(parseFloat(state.menu.style.top) + state.menu.getBoundingClientRect().height < anchor.top);
  state.c.toggleFileCtx(state.event(0, 0, anchor), 0);
  assert.equal(state.menu.style.display, 'none');
  assert.ok(!source.includes('renderFileCtxMenuHtml'), 'inline menus must not be clipped by file cards');
});

test('Escape, resize and outside scrolling dismiss menus without blocking internal scrolling', () => {
  const state = setup();
  const open = () => state.c.handleFileContextMenu(state.event(300, 300), 0);
  open(); state.listeners['document:scroll']({ target: state.menu }); assert.equal(state.menu.style.display, 'block');
  state.listeners['document:scroll']({ target: {} }); assert.equal(state.menu.style.display, 'none');
  open(); state.listeners['window:resize'](); assert.equal(state.menu.style.display, 'none');
  open(); state.listeners['document:keydown']({ key: 'Escape' }); assert.equal(state.menu.style.display, 'none');
});

test('uploaded server icons are a single transparent image with the branding outline', () => {
  const { c } = setup();
  const html = c.serverRowIconHtml({ icon_url: '/custom/transparent.png' });
  assert.match(html, /<img class="srv-row-thumb" src="\/custom\/transparent.png"/);
  assert.doesNotMatch(html, /background-image|style=/);
  assert.match(c.serverRowIconHtml({}), /src="\/brand\/default.png"/);
  assert.match(css, /\.srv-row-thumb\s*\{[^}]*background:\s*transparent;[^}]*filter:\s*var\(--np-logo-outline\)/s);
  assert.match(css, /\.brand-logo-img\s*\{[^}]*filter:\s*var\(--np-logo-outline\)/s);
});

test('broken uploaded icons fall back sequentially without layering or retry loops', () => {
  const { c } = setup();
  const image = { src: '/custom/broken.png', dataset: { fallbackIcon: '/brand/default.png' }, getAttribute() { return this.src; }, removeAttribute() { delete this.src; } };
  c.onServerRowIconError(image); assert.equal(image.src, '/brand/default.png');
  c.onServerRowIconError(image); assert.equal(image.src, '/static/img/default-server-icon.png');
  c.onServerRowIconError(image); assert.equal(image.hidden, true); assert.equal(image.onerror, null);
});

test('late content growth remains inside the space below and right of the menu', () => {
  const state = setup();
  state.c.handleFileContextMenu(state.event(790, 590), 0);
  state.menu.naturalWidth = 900;
  state.menu.naturalHeight = 1200;
  assertWithin(state, { left: 0, top: 0, width: 800, height: 600 });
});

test('Team, Database and Backups empty states use consistent SVG icons instead of emoji', () => {
  for (const icon of ['users', 'database', 'archive']) assert.ok(source.includes(`class="np-empty-icon" aria-hidden="true">\${ic('${icon}', 30)}`), icon);
  assert.doesNotMatch(section('function renderSubusersListHtml(', '\n  return `<div class="tw">'), /👥/);
  const c = vm.createContext({});
  vm.runInContext(section('function ic(', 'function fileListIconHtml('), c);
  for (const name of ['users', 'database', 'archive']) assert.match(c.ic(name, 30), /<svg[^>]*aria-hidden="true"/);
});
