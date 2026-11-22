// Progressive enhancement: never delay actions, replace DOM, move focus or retain styles.
(function (global) {
  'use strict';
  const doc = global.document;
  if (!doc) return;
  const preference = typeof global.matchMedia === 'function' ? global.matchMedia('(prefers-reduced-motion: reduce)') : null;
  const active = new Set();
  const pageAnimations = new Set();
  const feedbackAnimations = new WeakMap();
  const ease = 'cubic-bezier(.2,.8,.2,1)';
  const panelSelector = '.np-page-heading, .srvh, .card, .np-server-grid, .sift-server-card, .plugin-card, .vm-tiles-grid, .vm-icons-grid, .np-access-state, .np-empty-state, .file-table';
  const loadingSelector = '.np-loading, .np-skeleton, .np-page-loader, .np-loader-box';
  let currentKey = null;
  let pending = false;

  // The panel setting and OS setting are independent vetoes. Never overwrite the
  // panel attribute when the OS changes (including full -> reduced -> full).
  function reduced() { return Boolean(preference?.matches) || doc.documentElement.dataset.panelMotion === 'reduced'; }
  function enabled() { return !reduced() && !doc.hidden; }
  function forget(record) { active.delete(record); pageAnimations.delete(record); }
  function cancel(record) {
    try { record.animation.cancel(); } catch (_) {}
    forget(record);
  }
  function cancelPage() { Array.from(pageAnimations).forEach(cancel); }
  function cancelAll() { Array.from(active).forEach(cancel); }
  function resetSession() { cancelAll(); currentKey = null; pending = false; }
  function play(element, frames, options = {}, category = 'ui') {
    if (!enabled() || !element || element.isConnected === false || typeof element.animate !== 'function' || active.size >= 32) return null;
    let animation;
    try { animation = element.animate(frames, { duration: 180, easing: ease, fill: 'none', ...options }); }
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
    return rect.width > 0 && rect.height > 0 && rect.bottom >= 0 && rect.top <= global.innerHeight;
  }
  function viewKey(state) {
    const view = state.view || '';
    let pane = '';
    if (view === 'admin') pane = [state.adminTab || 0, state.adminTab === 4 ? state.appearanceSection || 0 : ''];
    else if (view === 'acc') pane = state.acct?.tab || 0;
    else if (view === 'set') pane = state.serverSettingsSection || 0;
    else if (view === 'pl') pane = state.ptab || 'online';
    else if (view === 'plug') pane = state.plugTab || 'browse';
    else if (view === 'ver') pane = state.sel;
    else if (view === 'mod') pane = state.modPageId;
    // Folder/archive browsing and returning from the editor are in-page actions.
    else if (view === 'fi') pane = state.fileTrashOpen ? 'trash' : 'files';
    const globalPage = ['dash', 'admin', 'acc', 'client', 'mod', 'maintenance', 'notfound'].includes(view) ||
      (view === 'act' && !(global.location?.pathname || '').startsWith('/server/'));
    return JSON.stringify([state.user?.id || '', view, globalPage ? '' : state.activeNodeId || 'local', globalPage ? '' : state.activeServerId || '', pane]);
  }
  function navigation(key) {
    if (key !== currentKey) { currentKey = key; pending = true; }
    return pending;
  }
  function afterRender(root, key) {
    // R replaces the nodes; cancel detached effects even on same-route refreshes.
    cancelPage();
    if (!root || !navigation(key)) return;
    // A navigation may resolve from a static placeholder once. A later data
    // refresh on that route is not another entrance, even if it briefly loads.
    if (root.querySelector(loadingSelector) || doc.hidden) return;
    pending = false;
    if (!enabled() || root.querySelector('#ed')) return;
    // Fade non-nested content groups together, not every row/card in sequence.
    // No transforms on ancestors: fixed menus and editor scroll stay untouched.
    Array.from(root.querySelectorAll(panelSelector)).slice(0, 64).filter(element => {
      const parent = element.parentElement?.closest(panelSelector);
      return (!parent || !root.contains(parent)) && !element.querySelector('video, audio, #ed') && visible(element);
    }).slice(0, 12).forEach(element => play(element,
      [{ opacity: .6 }, { opacity: 1 }], {}, 'page'));
  }
  function layout(root, toggle, from, to) {
    cancelPage();
    if (!enabled() || !root || from === to) return;
    const indicator = toggle?.querySelector('.vm-indicator');
    if (indicator) play(indicator,
      [{ transform: `translateX(${from * 100}%)` }, { transform: `translateX(${to * 100}%)` }],
      { duration: 160 }, 'page');
    const selector = '.np-server-grid, .vm-icons-grid, .vm-tiles-grid, table, #store-grid';
    Array.from(root.querySelectorAll(selector)).slice(0, 24).filter(el =>
      !el.parentElement?.closest(selector) && visible(el) && !el.querySelector('video, audio, #ed, ' + loadingSelector)
    ).slice(0, 4).forEach(el => play(el, [{ opacity: .6 }, { opacity: 1 }], { duration: 160 }, 'page'));
  }
  function revealOnce(element, key) {
    cancelPage();
    if (!element || !navigation(key) || doc.hidden) return null;
    pending = false;
    return play(element, [{ opacity: .6 }, { opacity: 1 }], {}, 'page');
  }
  function reveal(element) {
    return play(element, [{ opacity: .6 }, { opacity: 1 }], { duration: 160 });
  }
  function feedback(element) {
    const previous = feedbackAnimations.get(element);
    if (previous) cancel(previous);
    const record = reveal(element);
    if (record) feedbackAnimations.set(element, record);
    return record;
  }
  function syncPreference() { if (reduced()) cancelAll(); }
  function syncVisibility() {
    doc.documentElement.toggleAttribute('data-panel-motion-paused', Boolean(doc.hidden));
    if (doc.hidden) cancelAll();
  }
  doc.addEventListener('visibilitychange', syncVisibility);
  if (preference?.addEventListener) preference.addEventListener('change', syncPreference);
  else if (preference?.addListener) preference.addListener(syncPreference);
  if (typeof global.MutationObserver === 'function') {
    new global.MutationObserver(syncPreference).observe(doc.documentElement, { attributes: true, attributeFilter: ['data-panel-motion'] });
  }
  syncPreference();
  syncVisibility();
  global.NPMotion = Object.freeze({ afterRender, viewKey, reveal, revealOnce, resetSession, feedback, layout, cancelAll,
    get reduced() { return reduced(); }, get activeCount() { return active.size; } });
})(window);
