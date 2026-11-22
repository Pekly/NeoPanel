const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/motion.js'), 'utf8');
const tick = () => new Promise(resolve => setImmediate(resolve));

function setup(reduced = false, panelMotion) {
  const listeners = {}, animations = [];
  let measurements = 0;
  function element(id, opts = {}) {
    return { id, style: {}, isConnected: true, parentElement: null,
      closest(selector) { return selector.startsWith('[hidden]') && opts.hidden ? {} : null; },
      querySelector(selector) { return opts.media && selector.includes('video') ? {} : null; },
      getBoundingClientRect() { measurements++; return { top: opts.top ?? 30, bottom: (opts.top ?? 30) + 50, width: 120, height: 50 }; },
      animate(frames, options) {
        let finish, reject;
        const finished = new Promise((resolve, fail) => { finish = resolve; reject = fail; });
        const animation = { frames, options, element: this, finished, cancelled: false, finish,
          cancel() { this.cancelled = true; reject(Error('cancelled')); } };
        animations.push(animation); return animation;
      }
    };
  }
  const root = { dataset: { panelMotion }, attributes: {}, toggleAttribute(name, value) { if (value) this.attributes[name] = ''; else delete this.attributes[name]; } };
  const preference = { matches: reduced, addEventListener(name, callback) { listeners['preference:' + name] = callback; } };
  const doc = { hidden: false, documentElement: root, addEventListener(name, callback) { listeners[name] = callback; } };
  const c = vm.createContext({ document: doc, innerHeight: 800, matchMedia: () => preference,
    MutationObserver: class { constructor(callback) { listeners.panel = callback; } observe() {} } });
  c.window = c; vm.runInContext(source, c);
  function page(panels = [element('one'), element('two')]) {
    return { panels, loading: false, editor: false,
      querySelector(selector) { return selector === '#ed' ? (this.editor ? {} : null) : this.loading ? {} : null; },
      querySelectorAll() { return this.panels; }, contains(node) { return this.panels.includes(node); }
    };
  }
  return { c, motion: c.NPMotion, preference, doc, root, listeners, animations, element, page, measurements: () => measurements };
}

test('true navigation (including back) fades groups together; same-route refresh never replays', () => {
  const s = setup(), page = s.page();
  s.motion.afterRender(page, 'dashboard');
  assert.equal(s.animations.length, 2);
  assert.ok(s.animations.every(a => a.options.duration === 180 && !a.options.delay && a.options.fill === 'none'));
  assert.ok(s.animations.every(a => a.frames.every(f => !('transform' in f))));
  s.motion.afterRender(page, 'dashboard');
  assert.equal(s.animations.length, 2); assert.ok(s.animations.every(a => a.cancelled));
  s.motion.afterRender(page, 'files'); s.motion.afterRender(page, 'dashboard');
  assert.equal(s.animations.length, 6);
});

test('static placeholder resolves once; subsequent loading/refresh is not navigation', () => {
  const s = setup(), page = s.page(); page.loading = true;
  s.motion.afterRender(page, 'files'); assert.equal(s.animations.length, 0);
  page.loading = false; s.motion.afterRender(page, 'files'); assert.equal(s.animations.length, 2);
  page.loading = true; s.motion.afterRender(page, 'files');
  page.loading = false; s.motion.afterRender(page, 'files'); assert.equal(s.animations.length, 2);
});

test('rapid switching replaces pending loading route and cancels detached effects', () => {
  const s = setup(), page = s.page();
  page.loading = true; s.motion.afterRender(page, 'files');
  page.loading = false; s.motion.afterRender(page, 'dashboard');
  s.motion.afterRender(page, 'account');
  assert.equal(s.motion.activeCount, 2); assert.ok(s.animations.slice(0, 2).every(a => a.cancelled));
  s.motion.afterRender(page, 'account'); assert.equal(s.motion.activeCount, 0);
});

test('editors, media, hidden/offscreen and nested panels stay still', () => {
  const s = setup(), page = s.page([s.element('video', {media:true}), s.element('hidden', {hidden:true}), s.element('far', {top:2000})]);
  s.motion.afterRender(page, 'media'); assert.equal(s.animations.length, 0);
  page.editor = true; page.panels = [s.element('editor')];
  s.motion.afterRender(page, 'editor'); assert.equal(s.animations.length, 0);
  page.editor = false;
  const outer = s.element('outer'), inner = s.element('inner'); inner.parentElement = {closest: () => outer};
  page.panels = [outer, inner]; s.motion.afterRender(page, 'nested'); assert.equal(s.animations.length, 1);
});

test('large lists have bounded work and no sequential delays', () => {
  const s = setup(); s.motion.afterRender(s.page(Array.from({length:1000}, (_, i) => s.element(i))), 'big');
  assert.equal(s.animations.length, 12); assert.ok(s.measurements() <= 64);
  assert.ok(s.animations.every(a => !a.options.delay));
});

test('OS and panel reduced motion independently veto and cancel effects without overwriting preferences', () => {
  for (const [os, panel] of [[true,'full'], [false,'reduced']]) {
    const s = setup(os, panel); s.motion.afterRender(s.page(), 'dash'); s.motion.feedback(s.element('button'));
    assert.equal(s.animations.length, 0); assert.equal(s.root.dataset.panelMotion, panel);
  }
  const s = setup(false, 'full'); s.motion.afterRender(s.page(), 'dash');
  s.root.dataset.panelMotion = 'reduced'; s.listeners.panel();
  assert.equal(s.motion.activeCount, 0); assert.ok(s.animations.every(a => a.cancelled));
  s.preference.matches = true; s.listeners['preference:change']();
  s.preference.matches = false; s.listeners['preference:change']();
  assert.equal(s.root.dataset.panelMotion, 'reduced');
  s.root.dataset.panelMotion = 'full'; s.motion.afterRender(s.page(), 'dash'); assert.equal(s.animations.length, 2);
});

test('reduced visits do not replay when re-enabled; background tabs cancel and pause', () => {
  const s = setup(true); s.motion.afterRender(s.page(), 'dash');
  s.preference.matches = false; s.listeners['preference:change']();
  s.motion.afterRender(s.page(), 'dash'); assert.equal(s.animations.length, 0);
  s.motion.afterRender(s.page(), 'files'); s.doc.hidden = true; s.listeners.visibilitychange();
  assert.equal(s.motion.activeCount, 0); assert.ok('data-panel-motion-paused' in s.root.attributes);
  s.doc.hidden = false; s.listeners.visibilitychange(); assert.ok(!('data-panel-motion-paused' in s.root.attributes));
});

test('in-place settings selection and R share route identity; session reset releases effects', () => {
  const s = setup(); s.motion.afterRender(s.page(), 'general');
  s.motion.revealOnce(s.element('startup'), 'startup'); assert.equal(s.animations.length, 3);
  s.motion.afterRender(s.page(), 'startup'); assert.equal(s.animations.length, 3);
  s.motion.revealOnce(s.element('general'), 'general'); assert.equal(s.animations.length, 4);
  s.motion.resetSession(); assert.equal(s.motion.activeCount, 0);
  s.motion.afterRender(s.page(), 'general'); assert.equal(s.animations.length, 6);
});

test('layout indicator retains meaningful position; collection only fades and skips unchanged mode', () => {
  const s = setup(), indicator = s.element('indicator'), root = s.page();
  s.motion.layout(root, {querySelector: () => indicator}, 2, 0);
  assert.equal(s.animations[0].frames[0].transform, 'translateX(200%)');
  assert.equal(s.animations[0].frames[1].transform, 'translateX(0%)');
  assert.ok(s.animations.slice(1).every(a => a.frames.every(f => !('transform' in f))));
  s.motion.layout(root, null, 0, 0); assert.equal(s.motion.activeCount, 0);
  const reduced = setup(true); reduced.motion.layout(reduced.page(), null, 0, 1); assert.equal(reduced.animations.length, 0);
  s.motion.layout(s.page([s.element('media', {media:true})]), null, 0, 1); assert.equal(s.motion.activeCount, 0);
});

test('feedback replaces prior effect, releases resources, and unsupported WAAPI is harmless', async () => {
  const s = setup(), button = s.element('button'); s.motion.feedback(button); s.motion.feedback(button);
  assert.equal(s.motion.activeCount, 1); assert.equal(s.animations[0].cancelled, true);
  s.animations[1].finish(); await tick(); assert.equal(s.motion.activeCount, 0);
  button.animate = undefined; assert.doesNotThrow(() => s.motion.feedback(button)); assert.deepEqual(button.style, {});
  assert.equal(s.listeners.pointerdown, undefined, 'no decorative ripple/click interception');
});

test('route keys ignore telemetry, folders, previews and server selection on global views', () => {
  const {motion, c} = setup(), state = {view:'fi',user:{id:1},activeServerId:'a',activeNodeId:'east'};
  const key = motion.viewKey(state);
  for (const extra of [{cpu:90}, {memory:2048}, {currentDir:'world'}, {editingFile:'a.txt'}]) assert.equal(motion.viewKey({...state,...extra}), key);
  assert.notEqual(motion.viewKey({...state,activeNodeId:'west'}), key);
  assert.notEqual(motion.viewKey({...state,user:{id:2}}), key);
  assert.notEqual(motion.viewKey({...state,fileTrashOpen:true}), key);
  for (const [view, field] of [['admin','appearanceSection'],['set','serverSettingsSection'],['mod','modPageId'],['plug','plugTab']]) {
    const base = {...state,view,adminTab:4}; assert.notEqual(motion.viewKey(base), motion.viewKey({...base,[field]:2}));
  }
  assert.equal(motion.viewKey({...state,view:'dash'}), motion.viewKey({...state,view:'dash',activeServerId:'b'}));
  c.location = {pathname:'/activity'};
  assert.equal(motion.viewKey({...state,view:'act'}), motion.viewKey({...state,view:'act',activeServerId:'b'}));
  c.location.pathname = '/server/a/activity';
  assert.notEqual(motion.viewKey({...state,view:'act'}), motion.viewKey({...state,view:'act',activeServerId:'b'}));
});

test('motion contract loads last before branding; no broad transitions, retained transforms or aesthetic timers', () => {
  const html = fs.readFileSync(path.join(__dirname, '../static/index.html'), 'utf8');
  const css = fs.readFileSync(path.join(__dirname, '../static/css/motion.css'), 'utf8');
  assert.ok(html.indexOf('/css/motion.css') > html.indexOf('/css/server-artifact.css'));
  assert.ok(html.indexOf('/css/motion.css') < html.indexOf('id="custom-brand-css"'));
  assert.ok(html.indexOf('/js/motion.js') < html.indexOf('/js/app.js'));
  assert.match(css, /@media \(prefers-reduced-motion: reduce\)/);
  assert.match(css, /animation: none !important/); assert.match(css, /transition: none !important/);
  assert.match(css, /animation-play-state: paused !important/);
  assert.doesNotMatch(css, /transition:\s*all|will-change:|animation-delay:|filter:|scale\(/);
  assert.doesNotMatch(source, /setTimeout|requestAnimationFrame|sessionStorage|createElement/);
});
