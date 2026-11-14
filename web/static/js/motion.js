// Progressive enhancement only: animations never delay actions, replace DOM, or hide content without JS.
(function (global) {
  'use strict';
  const doc = global.document;
  if (!doc) return;
  const preference = typeof global.matchMedia === 'function' ? global.matchMedia('(prefers-reduced-motion: reduce)') : null;
  const active = new Set();
  const pageAnimations = new Set();
  const ripples = new Set();
  const notifications = new WeakMap();
  const ease = 'cubic-bezier(.2,.8,.2,1)';
  const panelSelector = '.np-page-heading, .srvh, .card, .sift-server-card, .plugin-card, .vm-tile-card, .vm-icon-card, .np-access-state, .np-empty-state';
  const visitStorageKey = 'neopanel_motion_visits_v1';
  const visited = new Set();
  try {
    const saved = JSON.parse(global.sessionStorage?.getItem(visitStorageKey) || '[]');
    if (Array.isArray(saved)) saved.slice(0, 1024).forEach(key => {
      if (typeof key === 'string' && key.length <= 4096) visited.add(key);
    });
  } catch (_) {}
  function firstVisit(key) {
    // Stop adding effects at the cap rather than evicting and replaying old pages.
    if (typeof key !== 'string' || key.length > 4096 || visited.has(key) || visited.size >= 1024) return false;
    visited.add(key);
    try { global.sessionStorage?.setItem(visitStorageKey, JSON.stringify([...visited])); } catch (_) {}
    return true;
  }
  function resetSession() {
    cancelAll();
    visited.clear();
    try { global.sessionStorage?.removeItem(visitStorageKey); } catch (_) {}
  }

  function reduced() { return Boolean(preference?.matches); }
  function enabled() { return !reduced() && !doc.hidden; }
  function forget(record) { active.delete(record); pageAnimations.delete(record); }
  function cancel(record) {
    try { record.animation.cancel(); } catch (_) {}
    forget(record);
  }
  function cancelAll() {
    Array.from(active).forEach(cancel);
    for (const layer of ripples) layer.remove();
    ripples.clear();
  }
  function play(element, frames, options = {}, category = 'ui') {
    if (!enabled() || !element || element.isConnected === false || typeof element.animate !== 'function' || active.size >= 64) return null;
    let animation;
    try { animation = element.animate(frames, { duration: 320, easing: ease, fill: 'backwards', ...options }); }
    catch (_) { return null; }
    const record = { animation };
    active.add(record);
    if (category === 'page') pageAnimations.add(record);
    Promise.resolve(animation.finished).then(() => forget(record), () => forget(record));
    return record;
  }
  function visible(element) {
    if (element.closest('[hidden], .np-editor-workspace, .file-preview-stage')) return false;
    const rect = element.getBoundingClientRect();
    return rect.width > 0 && rect.height > 0 && rect.bottom >= 0 && rect.top <= global.innerHeight + 80;
  }
  function viewKey(state, preview = null) {
    const view = state.view || '';
    let pane = '';
    if (view === 'admin') pane = [state.adminTab || 0, state.adminTab === 4 ? state.appearanceSection || 0 : ''];
    else if (view === 'acc') pane = state.acct?.tab || 0;
    else if (view === 'set') pane = state.serverSettingsSection || 0;
    else if (view === 'pl') pane = state.ptab || 'online';
    else if (view === 'plug') pane = state.plugTab || 'browse';
    else if (view === 'ver') pane = state.sel;
    else if (view === 'mod') pane = state.modPageId;
    else if (view === 'fi') pane = [state.currentDir, state.editingFile, preview?.path, preview?.folder, Boolean(state.fileTrashOpen)];
    const globalPage = ['dash', 'admin', 'acc', 'client', 'mod', 'maintenance', 'notfound'].includes(view) ||
      (view === 'act' && !(global.location?.pathname || '').startsWith('/server/'));
    return JSON.stringify([state.user?.id || '', view, globalPage ? '' : state.activeNodeId || 'local', globalPage ? '' : state.activeServerId || '', pane]);
  }
  function afterRender(root, key) {
    // Any previous render's elements have been replaced: release their animations immediately.
    Array.from(pageAnimations).forEach(cancel);
    if (!root) return;
    const loading = Boolean(root.querySelector('.np-skeleton, .np-page-loader, .np-loader-box'));
    if (loading || doc.hidden || !firstVisit('page:' + key) || !enabled() || root.querySelector('#ed')) return;

    const panels = Array.from(root.querySelectorAll(panelSelector)).slice(0, 64).filter(element => {
      const parent = element.parentElement?.closest(panelSelector);
      return (!parent || !root.contains(parent)) && !element.querySelector('video, audio, #ed') && visible(element);
    }).slice(0, 18);
    panels.forEach((element, index) => play(element,
      [{ opacity: 0, transform: 'translateY(14px)' }, { opacity: 1, transform: 'translateY(0)' }],
      { duration: 380, delay: Math.min(index * 28, 210) }, 'page'));

    Array.from(root.querySelectorAll('.file-row, .srow, .act-row, .np-role-row')).slice(0, 40).filter(visible).slice(0, 12)
      .forEach((element, index) => play(element, [{ opacity: 0 }, { opacity: 1 }],
        { duration: 240, delay: Math.min(70 + index * 18, 230) }, 'page'));

    if (firstVisit('navigation')) {
      Array.from(doc.querySelectorAll('#nav .nav-item, #nav .brand')).slice(0, 8)
        .forEach((element, index) => play(element, [{ opacity: 0 }, { opacity: 1 }],
          { duration: 300, delay: index * 22 }, 'page'));
    }
  }
  function layout(root, toggle, from, to) {
    Array.from(pageAnimations).forEach(cancel);
    if (!enabled() || !root) return;
    const indicator = toggle?.querySelector('.vm-indicator');
    if (indicator) play(indicator,
      [{ transform: `translateX(${from * 100}%)` }, { transform: `translateX(${to * 100}%)` }],
      { duration: 300 }, 'page');
    // Animate only collection surfaces, not the toolbar, editor, or navigation.
    const selector = '.np-server-grid, .vm-icons-grid, .vm-tiles-grid, table, #store-grid';
    Array.from(root.querySelectorAll(selector)).slice(0, 24).filter(el =>
      !el.parentElement?.closest(selector) && visible(el) && !el.querySelector('video, audio, #ed')
    ).slice(0, 4).forEach(el => play(el,
      [{ opacity: .45, transform: `translateX(${to > from ? 18 : -18}px)` }, { opacity: 1, transform: 'translateX(0)' }],
      { duration: 280 }, 'page'));
  }
  function revealOnce(element, key) {
    if (!element || doc.hidden || !firstVisit('page:' + key)) return null;
    Array.from(pageAnimations).forEach(cancel);
    return play(element, [{ opacity: .35 }, { opacity: 1 }], { duration: 260 }, 'page');
  }
  function reveal(element) {
    return play(element, [{ opacity: .35 }, { opacity: 1 }], { duration: 260 });
  }
  function notify(element) {
    const previous = notifications.get(element);
    if (previous) cancel(previous);
    const record = play(element, [{ opacity: .25, transform: 'translateY(10px)' }, { opacity: 1, transform: 'translateY(0)' }], { duration: 300 });
    if (record) notifications.set(element, record);
  }
  function feedback(element) {
    return play(element, [{ opacity: .55 }, { opacity: 1 }], { duration: 220 });
  }
  function ripple(element, event) {
    if (!enabled() || ripples.size >= 6 || element.disabled || element.getAttribute('aria-disabled') === 'true') return;
    const rect = element.getBoundingClientRect();
    if (rect.width <= 0 || rect.height <= 0) return;
    const style = global.getComputedStyle(element);
    const layer = doc.createElement('span');
    layer.className = 'np-motion-ripple-layer';
    layer.setAttribute('aria-hidden', 'true');
    Object.assign(layer.style, {
      left: `${rect.left}px`, top: `${rect.top}px`, width: `${rect.width}px`, height: `${rect.height}px`,
      borderRadius: style.borderRadius, color: style.color
    });
    const wave = doc.createElement('span');
    wave.className = 'np-motion-ripple';
    const size = Math.hypot(rect.width, rect.height) * 2;
    const keyboard = event.type === 'click' && event.detail === 0;
    const x = keyboard ? rect.width / 2 : Math.max(0, Math.min(rect.width, event.clientX - rect.left));
    const y = keyboard ? rect.height / 2 : Math.max(0, Math.min(rect.height, event.clientY - rect.top));
    Object.assign(wave.style, { width: `${size}px`, height: `${size}px`, left: `${x - size / 2}px`, top: `${y - size / 2}px` });
    layer.appendChild(wave);
    doc.body.appendChild(layer);
    ripples.add(layer);
    const record = play(wave, [{ opacity: .18, transform: 'scale(0)' }, { opacity: 0, transform: 'scale(1)' }], { duration: 440, easing: 'cubic-bezier(.2,.7,.3,1)' });
    const remove = () => { layer.remove(); ripples.delete(layer); };
    if (record) Promise.resolve(record.animation.finished).then(remove, remove);
    else remove();
  }
  function onPress(event) {
    if (event.type === 'pointerdown' && event.button !== 0) return;
    if (event.type === 'click' && event.detail !== 0) return;
    const element = event.target?.closest?.('.btn, #nav .nav-item, .tabs button, .atabs button, .cselect-trigger, .vm-toggle button');
    if (!element || element.closest('[inert], [aria-disabled="true"]')) return;
    ripple(element, event);
  }
  function syncPreference() {
    doc.documentElement.dataset.panelMotion = reduced() ? 'reduced' : 'full';
    if (reduced()) cancelAll();
    // Preference changes never reset which pages this session has visited.
  }
  function syncVisibility() {
    doc.documentElement.toggleAttribute('data-panel-motion-paused', Boolean(doc.hidden));
    if (doc.hidden) cancelAll();
  }
  doc.addEventListener('pointerdown', onPress, { passive: true });
  doc.addEventListener('click', onPress, { passive: true });
  doc.addEventListener('visibilitychange', syncVisibility);
  if (preference?.addEventListener) preference.addEventListener('change', syncPreference);
  else if (preference?.addListener) preference.addListener(syncPreference);
  syncPreference();
  syncVisibility();
  global.NPMotion = Object.freeze({ afterRender, viewKey, reveal, revealOnce, resetSession, notify, feedback, layout, cancelAll,
    get reduced() { return reduced(); }, get activeCount() { return active.size; } });
})(window);
