const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/motion.js'), 'utf8');
const tick = () => new Promise(resolve => setImmediate(resolve));

function setup(reduced = false, storage = new Map()) {
  const listeners = {}, animations = [], created = [];
  let measurements = 0;
  function element(id, opts = {}) {
    const el = { id, style: {}, dataset: {}, children: [], isConnected: true, parentElement: null,
      disabled: false, attributes: {},
      getAttribute(name) { return this.attributes[name] ?? null; },
      setAttribute(name, value) { this.attributes[name] = value; },
      closest(selector) { return selector.startsWith('[hidden]') ? (opts.hidden ? {} : null) : selector.startsWith('[inert]') ? (opts.inert ? {} : null) : opts.button ? this : null; },
      querySelector(selector) { return opts.media && selector.includes('video') ? {} : null; },
      getBoundingClientRect() { measurements++; return { left: 20, top: opts.top ?? 30, bottom: (opts.top ?? 30) + 50, width: 120, height: 50 }; },
      appendChild(child) { this.children.push(child); child.parentElement = this; },
      remove() { this.isConnected = false; if (this.parentElement) this.parentElement.children = this.parentElement.children.filter(child => child !== this); },
      animate(frames, options) {
        let finish, reject;
        const finished = new Promise((resolve, fail) => { finish = resolve; reject = fail; });
        const animation = { frames, options, element: this, finished, cancelled: false, finish,
          cancel() { this.cancelled = true; reject(Error('cancelled')); } };
        animations.push(animation);
        return animation;
      }
    };
    return el;
  }
  const body = element('body');
  const root = { dataset: {}, attributes: {}, toggleAttribute(name, value) { if (value) this.attributes[name] = ''; else delete this.attributes[name]; } };
  const preference = { matches: reduced, addEventListener(name, callback) { listeners['preference:' + name] = callback; } };
  const doc = { hidden: false, documentElement: root, body,
    addEventListener(name, callback) { listeners[name] = callback; },
    querySelectorAll() { return []; },
    createElement(tag) { const el = element(tag + created.length); created.push(el); return el; }
  };
  const c = vm.createContext({ document: doc, innerHeight: 800, matchMedia: () => preference,
    sessionStorage: { getItem: key => storage.get(key), setItem: (key, value) => storage.set(key, value), removeItem: key => storage.delete(key) },
    getComputedStyle: () => ({ borderRadius: '8px', color: 'rgb(255, 255, 255)' }) });
  c.window = c;
  vm.runInContext(source, c);
  function page(panels = [element('one'), element('two')], rows = [element('row')]) {
    return { panels, rows, loading: false, editor: false,
      querySelector(selector) { return selector === '#ed' ? (this.editor ? {} : null) : this.loading ? {} : null; },
      querySelectorAll(selector) { return selector.startsWith('.file-row') ? this.rows : this.panels; },
      contains(node) { return this.panels.includes(node); }
    };
  }
  return { c, motion: c.NPMotion, preference, doc, root, body, listeners, animations, element, page, created, measurements: () => measurements };
}

test('page/card entrances stagger only on first visit, never when navigating back', () => {
  const s = setup(), page = s.page();
  s.motion.afterRender(page, 'dashboard');
  assert.equal(s.animations.length, 3);
  assert.deepEqual(s.animations.slice(0, 2).map(a => a.options.delay), [0, 28]);
  const first = [...s.animations];
  s.motion.afterRender(page, 'dashboard');
  assert.equal(s.animations.length, 3);
  assert.ok(first.every(a => a.cancelled));
  s.motion.afterRender(page, 'files');
  assert.equal(s.animations.length, 6);
  s.motion.afterRender(page, 'dashboard');
  s.motion.afterRender(page, 'files');
  assert.equal(s.animations.length, 6);
});

test('loaded content gets an entrance after a loading placeholder settles', () => {
  const s = setup(), page = s.page(); page.loading = true;
  s.motion.afterRender(page, 'dashboard');
  const before = s.animations.length;
  assert.equal(before, 0, 'loading placeholders must not spend the first entrance');
  page.loading = false;
  s.motion.afterRender(page, 'dashboard');
  assert.ok(s.animations.length > before);
  const after = s.animations.length;
  s.motion.afterRender(page, 'dashboard');
  assert.equal(s.animations.length, after);
});

test('route keys include server and active pane but exclude CPU/memory updates', () => {
  const { motion } = setup();
  const state = { view: 'fi', activeServerId: 'one', activeNodeId: 'east', currentDir: 'plugins', cpu: 10 };
  const key = motion.viewKey(state);
  assert.equal(motion.viewKey({ ...state, cpu: 90, memory: 2048 }), key);
  assert.notEqual(motion.viewKey({ ...state, activeNodeId: 'west' }), key);
  assert.notEqual(motion.viewKey({ ...state, currentDir: 'world' }), key);
  assert.notEqual(motion.viewKey(state, { path: 'world.zip', folder: 'region/' }), motion.viewKey(state, { path: 'world.zip', folder: '' }));
});

test('editors, media panels, hidden panels and offscreen panels stay still', () => {
  const s = setup();
  const page = s.page([s.element('video', { media: true }), s.element('hidden', { hidden: true }), s.element('far', { top: 2000 })], []);
  s.motion.afterRender(page, 'media');
  assert.equal(s.animations.length, 0);
  page.editor = true; page.panels = [s.element('card')];
  s.motion.afterRender(page, 'editor');
  assert.equal(s.animations.length, 0);
});

test('large lists have bounded measurement and animation work', () => {
  const s = setup();
  const page = s.page(Array.from({ length: 1000 }, (_, i) => s.element('card' + i)), Array.from({ length: 1000 }, (_, i) => s.element('row' + i)));
  s.motion.afterRender(page, 'big-list');
  assert.equal(s.animations.length, 30);
  assert.ok(s.measurements() <= 104);
  assert.ok(s.animations.every(a => a.options.delay <= 230));
});

test('reduced motion prevents effects and preference changes cancel active effects', () => {
  const initial = setup(true);
  initial.motion.afterRender(initial.page(), 'dashboard');
  initial.motion.notify(initial.element('toast'));
  assert.equal(initial.animations.length, 0);
  assert.equal(initial.root.dataset.panelMotion, 'reduced');
  const s = setup();
  s.motion.afterRender(s.page(), 'dashboard');
  s.preference.matches = true;
  s.listeners['preference:change']();
  assert.equal(s.motion.activeCount, 0);
  assert.ok(s.animations.every(a => a.cancelled));
  assert.equal(s.root.dataset.panelMotion, 'reduced');
});

test('background tabs cancel running effects and pause CSS loops', () => {
  const s = setup();
  s.motion.afterRender(s.page(), 'dashboard');
  s.doc.hidden = true; s.listeners.visibilitychange();
  assert.equal(s.motion.activeCount, 0);
  assert.ok('data-panel-motion-paused' in s.root.attributes);
  s.motion.feedback(s.element('button'));
  assert.equal(s.animations.length, 3);
  s.doc.hidden = false; s.listeners.visibilitychange();
  assert.ok(!('data-panel-motion-paused' in s.root.attributes));
});

test('pointer and keyboard ripples do not consume clicks or animate disabled controls', () => {
  const s = setup(), button = s.element('button', { button: true });
  const event = { type: 'pointerdown', button: 0, clientX: 40, clientY: 40, target: button };
  button.disabled = true; s.listeners.pointerdown(event); assert.equal(s.body.children.length, 0);
  button.disabled = false; s.listeners.pointerdown({ ...event, button: 2 }); assert.equal(s.body.children.length, 0);
  s.listeners.pointerdown(event); assert.equal(s.body.children.length, 1);
  s.listeners.click({ type: 'click', detail: 1, target: button }); assert.equal(s.body.children.length, 1);
  s.listeners.click({ type: 'click', detail: 0, target: button }); assert.equal(s.body.children.length, 2);
  assert.equal(s.body.children[0].attributes['aria-hidden'], 'true');
  s.motion.cancelAll(); assert.equal(s.body.children.length, 0);
});

test('ripple bursts are capped and finished effects release resources', async () => {
  const s = setup(), button = s.element('button', { button: true });
  for (let i = 0; i < 30; i++) s.listeners.pointerdown({ type: 'pointerdown', button: 0, clientX: 40, clientY: 40, target: button });
  assert.equal(s.body.children.length, 6);
  for (const animation of s.animations) animation.finish();
  await tick();
  assert.equal(s.motion.activeCount, 0);
  assert.equal(s.body.children.length, 0);
});

test('rapid toast replacement cancels the prior animation and unsupported WAAPI is harmless', () => {
  const s = setup(), toast = s.element('toast');
  s.motion.notify(toast); s.motion.notify(toast);
  assert.equal(s.motion.activeCount, 1);
  assert.equal(s.animations[0].cancelled, true);
  const plain = s.element('plain'); plain.animate = undefined;
  assert.doesNotThrow(() => s.motion.feedback(plain));
  assert.deepEqual(plain.style, {});
});

test('motion assets load before branding overrides and application rendering', () => {
  const html = fs.readFileSync(path.join(__dirname, '../static/index.html'), 'utf8');
  assert.ok(html.indexOf('/css/motion.css') > html.indexOf('/css/workbench.css'));
  assert.ok(html.indexOf('/css/motion.css') < html.indexOf('id="custom-brand-css"'));
  assert.ok(html.indexOf('/js/motion.js') < html.indexOf('/js/app.js'));
});

test('menu reveal is geometry-safe and reduced motion covers legacy CSS effects', () => {
  const css = fs.readFileSync(path.join(__dirname, '../static/css/motion.css'), 'utf8');
  const fade = css.slice(css.indexOf('@keyframes np-motion-fade'), css.indexOf('@keyframes np-motion-dialog'));
  assert.match(fade, /opacity:/);
  assert.doesNotMatch(fade, /transform:|scale:|translate:/);
  assert.match(css, /@media \(prefers-reduced-motion: reduce\)/);
  assert.match(css, /animation: none !important/);
  assert.match(css, /transition: none !important/);
  assert.match(css, /animation-play-state: paused !important/);
  assert.doesNotMatch(css, /will-change:/);
});

test('layout indicator slides from previous mode and collection motion stays bounded', () => {
  const s = setup(), indicator = s.element('indicator');
  const root = s.page(Array.from({ length: 100 }, (_, i) => s.element('collection' + i)), []);
  s.motion.layout(root, { querySelector: () => indicator }, 2, 0);
  assert.equal(s.animations.length, 5);
  assert.equal(s.animations[0].frames[0].transform, 'translateX(200%)');
  assert.equal(s.animations[0].frames[1].transform, 'translateX(0%)');
  assert.equal(s.animations[1].frames[0].transform, 'translateX(-18px)');
  assert.ok(s.measurements() <= 24);
  s.motion.layout(root, { querySelector: () => indicator }, 0, 1);
  assert.ok(s.animations.slice(0, 5).every(a => a.cancelled));
  assert.equal(s.motion.activeCount, 5);
});

test('layout motion respects reduced motion and avoids media surfaces', () => {
  const reduced = setup(true);
  reduced.motion.layout(reduced.page(), null, 0, 2);
  assert.equal(reduced.animations.length, 0);
  const s = setup();
  s.motion.layout(s.page([s.element('media', { media: true })], []), null, 0, 1);
  assert.equal(s.animations.length, 0);
});

test('visited pages survive reload in the same tab and reset for a new session', () => {
  const storage = new Map();
  const first = setup(false, storage);
  first.motion.afterRender(first.page(), 'dashboard');
  const reload = setup(false, storage);
  reload.motion.afterRender(reload.page(), 'dashboard');
  assert.equal(reload.animations.length, 0);
  reload.motion.afterRender(reload.page(), 'new-page');
  assert.equal(reload.animations.length, 3);
  reload.motion.resetSession();
  assert.equal(reload.motion.activeCount, 0);
  reload.motion.afterRender(reload.page(), 'dashboard');
  assert.equal(reload.animations.length, 6);
  const newTab = setup();
  newTab.motion.afterRender(newTab.page(), 'dashboard');
  assert.equal(newTab.animations.length, 3);
});

test('in-place settings reveals share first-visit tracking with full page renders', () => {
  const s = setup(), page = s.page();
  s.motion.afterRender(page, 'settings-general');
  s.motion.revealOnce(s.element('startup'), 'settings-startup');
  assert.equal(s.animations.length, 4);
  s.motion.revealOnce(s.element('general'), 'settings-general');
  s.motion.revealOnce(s.element('startup'), 'settings-startup');
  s.motion.afterRender(page, 'settings-startup');
  assert.equal(s.animations.length, 4);
});

test('reduced-motion visits stay visited after changing preferences', () => {
  const s = setup(true);
  s.motion.afterRender(s.page(), 'dashboard');
  s.preference.matches = false; s.listeners['preference:change']();
  s.motion.afterRender(s.page(), 'dashboard');
  assert.equal(s.animations.length, 0);
  s.motion.afterRender(s.page(), 'files');
  assert.equal(s.animations.length, 3);
});

test('background preloads do not consume the visible first visit', () => {
  const s = setup(); s.doc.hidden = true;
  s.motion.afterRender(s.page(), 'files');
  s.doc.hidden = false;
  s.motion.afterRender(s.page(), 'files');
  assert.equal(s.animations.length, 3);
});

test('page keys distinguish users and subtabs but ignore server selection for global pages', () => {
  const { motion } = setup();
  const state = { view: 'dash', user: { id: 1 }, activeServerId: 'a' };
  assert.equal(motion.viewKey(state), motion.viewKey({ ...state, activeServerId: 'b' }));
  assert.notEqual(motion.viewKey(state), motion.viewKey({ ...state, user: { id: 2 } }));
  for (const [view, field] of [['admin', 'appearanceSection'], ['set', 'serverSettingsSection'], ['mod', 'modPageId']]) {
    const base = { ...state, view, adminTab: 4 };
    assert.notEqual(motion.viewKey(base), motion.viewKey({ ...base, [field]: 2 }));
  }
});

test('bad or blocked session storage does not break motion', () => {
  const bad = new Map([['neopanel_motion_visits_v1', '{not-json']]);
  const blocked = { get() { throw Error('blocked'); }, set() { throw Error('blocked'); }, delete() { throw Error('blocked'); } };
  for (const storage of [bad, blocked]) {
    const s = setup(false, storage);
    assert.doesNotThrow(() => { s.motion.afterRender(s.page(), 'one'); s.motion.resetSession(); });
  }
});

test('global activity is one page; server activity has separate visit identities', () => {
  const { c, motion } = setup();
  const a = { view: 'act', activeServerId: 'a' }, b = { view: 'act', activeServerId: 'b' };
  c.location = { pathname: '/activity' };
  assert.equal(motion.viewKey(a), motion.viewKey(b));
  c.location.pathname = '/server/a/activity';
  assert.notEqual(motion.viewKey(a), motion.viewKey(b));
});
