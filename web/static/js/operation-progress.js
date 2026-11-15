/* Request-scoped feedback: indeterminate until the host responds, never invented percentages. */
(function () {
  'use strict';
  const active = new Set();
  let timer = null;
  let tray = null;

  function describe(endpoint, options = {}) {
    if (String(options.method || 'GET').toUpperCase() !== 'POST') return null;
    let path;
    try { path = new URL(endpoint, window.location.href).pathname; } catch (_) { return null; }
    const node = path.match(/^\/api\/nodes\/([^/]+)\/proxy(?=\/api\/)/);
    if (node) path = path.slice(node[0].length);
    const server = path.match(/^\/api\/servers\/([^/]+)(\/.*)$/);
    const route = server ? server[2] : path;
    const labels = {
      '/reinstall': 'Reinstalling server',
      '/software': 'Installing server software',
      '/plugins/install': 'Installing plugin or mod',
      '/backups': 'Creating backup',
      '/backups/restore': 'Restoring backup',
      '/files/compress': 'Creating archive',
      '/files/archive': 'Creating archive',
      '/files/unzip': 'Extracting archive',
      '/files/extract': 'Extracting archive',
      '/api/system/java/install': 'Installing Java runtime'
    };
    const label = labels[route];
    if (!label || (!server && route !== '/api/system/java/install')) return null;
    return { label, server: server?.[1] || '', node: node?.[1] || 'local', inline: route === '/reinstall' };
  }

  function element(tag, className, text) {
    const el = document.createElement(tag);
    if (className) el.className = className;
    if (text) el.textContent = text;
    return el;
  }

  function getTray() {
    if (!tray?.isConnected) {
      tray = element('section', 'np-operation-tray');
      tray.setAttribute('aria-label', 'Operations in progress');
      document.body.appendChild(tray);
    }
    return tray;
  }

  function tick() {
    for (const task of active) {
      // A closed/re-rendered dialog must not hide a still-pending operation.
      if (!task.card.isConnected) getTray().appendChild(task.card);
      const seconds = Math.max(0, Math.floor((Date.now() - task.started) / 1000));
      task.elapsed.textContent = seconds < 60 ? `${seconds}s elapsed` : `${Math.floor(seconds / 60)}m ${seconds % 60}s elapsed`;
      if (seconds >= 30) task.hint.textContent = 'Still waiting for the host. This may take several minutes. Closing the dialog won’t cancel it.';
    }
  }

  function startRequest(endpoint, options) {
    const info = describe(endpoint, options);
    if (!info || !document.body) return () => {};
    const card = element('div', 'np-operation-card');
    const heading = element('div', 'np-operation-heading');
    const title = element('strong', '', info.label + '…');
    title.setAttribute('role', 'status');
    const elapsed = element('span', 'np-operation-elapsed', '0s elapsed');
    elapsed.setAttribute('aria-hidden', 'true');
    heading.append(title, elapsed);
    const context = element('div', 'np-operation-context', [info.server, info.node === 'local' ? 'Local host' : info.node].filter(Boolean).join(' · '));
    const bar = element('div', 'np-operation-bar');
    bar.setAttribute('role', 'progressbar');
    bar.setAttribute('aria-label', info.label);
    bar.setAttribute('aria-valuetext', 'Waiting for the host; percentage unavailable');
    bar.appendChild(element('span', 'np-operation-fill'));
    const hint = element('p', 'np-operation-hint', 'Waiting for the host. Closing the dialog won’t cancel the operation.');
    card.append(heading, context, bar, hint);
    let mount = null;
    if (info.inline) {
      const form = document.getElementById('reinstall-form');
      if (form?.dataset.srvId === info.server && (form.dataset.nodeId || 'local') === info.node) {
        mount = form.querySelector('[data-operation-progress]');
      }
    }
    (mount || getTray()).appendChild(card);
    const task = { card, elapsed, hint, started: Date.now() };
    active.add(task);
    if (timer === null) timer = window.setInterval(tick, 1000);
    // Each request owns only its own card; concurrent work remains visible.
    return () => {
      if (!active.delete(task)) return;
      card.remove();
      if (tray && !tray.childElementCount) { tray.remove(); tray = null; }
      if (!active.size && timer !== null) { window.clearInterval(timer); timer = null; }
    };
  }

  function clear() {
    for (const task of active) task.card.remove();
    active.clear();
    if (timer !== null) window.clearInterval(timer);
    timer = null;
    tray?.remove();
    tray = null;
  }

  window.NPProgress = { startRequest, clear };
})();
