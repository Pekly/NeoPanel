// Shared web-panel QoL. No desktop or host-platform dependencies.
(() => {
  'use strict';
  let user = '', prefs = {}, bootFiles = {}, active = '', ready = false, restoreAttempted = false;
  let consoleState = { filter: '', follow: true, top: 0 };
  const failed = new Set(), countdowns = new Map(), pending = new Set();
  const revisions = new Map(), countdownErrors = new Map();
  let pollBusy = false, retryBusy = false, restoringFile = false;
  const key = srv => JSON.stringify([srv?.node_id || 'local', srv?.id || '']);
  const prefKey = () => 'neopanel:workspace:' + user;
  const sessionKey = (kind, srv) => 'neopanel:' + kind + ':' + JSON.stringify([user, srv?.node_id || 'local', srv?.id]);
  function read(storage, name, fallback) { try { return JSON.parse(storage.getItem(name)) ?? fallback; } catch (_) { return fallback; } }
  function write(storage, name, value) { try { storage.setItem(name, JSON.stringify(value)); } catch (_) {} }
  function save() { if (user) write(localStorage, prefKey(), prefs); }
  function loadUser() {
    const next = String(S.user?.id || S.user?.username || '');
    if (next === user) return;
    user = next; active = ''; ready = false; restoreAttempted = false;
    failed.clear(); countdowns.clear(); pending.clear(); revisions.clear(); countdownErrors.clear();
    prefs = read(localStorage, prefKey(), {});
    if (!prefs || typeof prefs !== 'object' || Array.isArray(prefs)) prefs = {};
    prefs.favourites = Array.isArray(prefs.favourites) ? prefs.favourites.filter(x => typeof x === 'string').slice(0, 500) : [];
    prefs.files = prefs.files && typeof prefs.files === 'object' && !Array.isArray(prefs.files) ? prefs.files : {};
    bootFiles = { ...prefs.files };
    S.serverFilter = typeof prefs.search === 'string' ? prefs.search.slice(0, 200) : '';
    S.nodeFilter = typeof prefs.node === 'string' ? prefs.node : 'all';
    S.serverSort = ['name', 'status', 'node'].includes(prefs.sort) ? prefs.sort : 'name';
    S.favouritesOnly = prefs.favouritesOnly === true;
  }
  function activate(srv) {
    if (!user || !srv?.id || active === key(srv)) return;
    active = key(srv);
    const stored = read(sessionStorage, sessionKey('console', srv), {});
    consoleState = { filter: typeof stored?.filter === 'string' ? stored.filter.slice(0, 200) : '', follow: true, top: 0 };
    S.hist = Array.isArray(stored?.history) ? stored.history.filter(x => typeof x === 'string').slice(-100) : [];
    S.hi = S.hist.length;
  }
  function saveConsole() {
    const srv = curServer();
    if (user && active === key(srv)) write(sessionStorage, sessionKey('console', srv), { filter: consoleState.filter, history: S.hist.slice(-100) });
  }
  function recordCommand(command) {
    activate(curServer());
    if (S.hist.at(-1) !== command) S.hist.push(command);
    S.hist = S.hist.slice(-100); S.hi = S.hist.length; saveConsole();
  }
  function clearHistory() { S.hist = []; S.hi = 0; saveConsole(); toast('Command history cleared for this server'); }
  function remember() {
    if (!ready || !user || !S.token) return;
    const route = parseCurrentRoute();
    // Persist only known, authorized server routes, never auth links or arbitrary URLs.
    if (route.serverId && curServer() && canServerPage(S.view) && !['notfound', 'serverpending', 'maintenance'].includes(S.view)) {
      prefs.route = { id: curServer().id, node: curServer().node_id || 'local', view: S.view, plugTab: S.plugTab };
    } else if (S.view === 'dash') prefs.route = null;
    if (S.view === 'fi' && hasServerPermission('files.view') && !restoringFile) {
      prefs.files[key(curServer())] = { dir: S.currentDir || '', path: S.editingFile || '' };
      const entries = Object.entries(prefs.files).slice(-100); prefs.files = Object.fromEntries(entries);
    }
    save();
  }
  function restoreRoute() {
    if (S.serverCatalog?.clusterComplete && S.nodeFilter !== 'all' && !S.servers.some(s => (s.node_id || 'local') === S.nodeFilter) && !(S.nodes || []).some(n => n.id === S.nodeFilter)) { S.nodeFilter = 'all'; saveList(); }
    if (restoreAttempted) { ready = true; return; }
    restoreAttempted = true;
    const saved = prefs.route;
    if (location.pathname === '/' && !location.search && !location.hash && saved && ['con','fi','set','act','db','bk','sc','team','net','sftp','sub','pl','plug','ver'].includes(saved.view)) {
      const srv = S.servers.find(s => s.id === saved.id && (s.node_id || 'local') === saved.node);
      if (srv) history.replaceState(null, '', serverRouteUrl(saved.node, saved.id, saved.view, { plugTab: saved.plugTab }));
    }
    ready = true;
  }
  async function restoreFile() {
    const srv = curServer(), token = S.token;
    if (S.view !== 'fi' || S.editingFile || !hasServerPermission('files.view', srv)) return;
    const file = bootFiles[key(srv)];
    delete bootFiles[key(srv)];
    if (!file || typeof file.dir !== 'string' || typeof file.path !== 'string') return;
    const valid = path => !path.startsWith('/') && !path.includes('\\') && !path.split('/').includes('..') && !/[\x00-\x1f]/.test(path);
    if (!valid(file.dir) || !valid(file.path)) return;
    const viewKey = key(srv);
    restoringFile = true;
    try {
      const isCurrent = () => S.token === token && S.view === 'fi' && key(curServer()) === viewKey;
      const loaded = await loadFileList(file.dir, true, { isCurrent });
      if (!loaded || !isCurrent()) return;
      if (file.path && file.path.split('/').slice(0, -1).join('/') === file.dir) await openQolFile(file.path.split('/').pop());
    } finally { restoringFile = false; }
  }
  function saveList() {
    prefs.search = (S.serverFilter || '').slice(0, 200); prefs.node = S.nodeFilter || 'all';
    prefs.sort = S.serverSort || 'name'; prefs.favouritesOnly = Boolean(S.favouritesOnly); save();
  }
  function favourite(srv) { return Boolean(prefs.favourites?.includes(key(srv))); }
  function toggleFavourite(index, event) {
    event?.preventDefault(); event?.stopPropagation();
    const srv = S.servers[index]; if (!srv) return false;
    const id = key(srv);
    prefs.favourites = favourite(srv) ? prefs.favourites.filter(x => x !== id) : [...(prefs.favourites || []), id].slice(-500);
    save(); R(); return false;
  }
  function favouriteButton(srv, index) {
    const pinned = favourite(srv);
    return `<span class="srv-card-action-btn${pinned ? ' np-favourite' : ''}" role="button" tabindex="0" aria-pressed="${pinned}" aria-label="${pinned ? 'Unpin' : 'Pin'} server" title="${pinned ? 'Unpin' : 'Pin'} server" onkeydown="if(event.key==='Enter'||event.key===' '){event.preventDefault();event.stopPropagation();this.click();}" onclick="return NPWorkspace.toggleFavourite(${index}, event)">${pinned ? '★' : '☆'}</span>`;
  }
  function sortServers(rows) {
    const status = s => normalizeServerStatus(s.stats?.status || s.status);
    return rows.filter(({s}) => !S.favouritesOnly || favourite(s)).slice().sort((a, b) => {
      const pin = Number(favourite(b.s)) - Number(favourite(a.s)); if (pin) return pin;
      const mode = S.serverSort || 'name';
      const value = s => mode === 'status' ? status(s) : mode === 'node' ? (s.node_name || s.node_id || 'local') : (s.name || '');
      return String(value(a.s)).localeCompare(String(value(b.s)), undefined, { numeric: true, sensitivity: 'base' }) || String(a.s.name || '').localeCompare(String(b.s.name || '')) || a.origIdx - b.origIdx;
    });
  }
  function listControls() {
    return `<select class="sel sm np-server-sort" aria-label="Sort servers" onchange="S.serverSort=this.value;NPWorkspace.saveList();R()">${[['name','Name'],['status','Status'],['node','Location']].map(([v,l]) => `<option value="${v}" ${S.serverSort === v ? 'selected' : ''}>${l}</option>`).join('')}</select><button class="btn sm${S.favouritesOnly ? ' pr' : ''}" aria-pressed="${Boolean(S.favouritesOnly)}" onclick="S.favouritesOnly=!S.favouritesOnly;NPWorkspace.saveList();R()">★ Pinned only</button>`;
  }
  function consoleTools(id = 'log') {
    return `<details class="np-console-options np-compact-menu"><summary class="btn sm">Console tools <span data-console-filter-badge>${consoleState.filter ? '· Filter on' : ''}</span><span class="np-menu-chevron" aria-hidden="true"></span></summary><div class="np-console-tools np-menu-panel"><label><span>Filter output</span><input type="search" value="${esc(consoleState.filter)}" maxlength="200" aria-label="Filter console output" oninput="NPWorkspace.setFilter(this.value)"></label><button type="button" class="btn sm" data-jump-log="${id}" onclick="NPWorkspace.jump('${id}')">Jump to latest</button><button type="button" class="btn sm" onclick="NPWorkspace.clearHistory()">Clear command history</button></div></details>`;
  }
  function filterRows(el) {
    const term = consoleState.filter.toLowerCase();
    for (const row of el.children) row.hidden = Boolean(term && !row.textContent.toLowerCase().includes(term));
  }
  function setFilter(value) {
    consoleState.filter = String(value).slice(0, 200); saveConsole();
    document.querySelectorAll('[data-console-filter-badge]').forEach(el => { el.textContent = consoleState.filter ? '· Filter on' : ''; });
    for (const id of ['log', 'fly-log']) { const el = document.getElementById(id); if (el) { filterRows(el); if (consoleState.follow) el.scrollTop = el.scrollHeight; } }
  }
  function captureConsole() {
    const el = document.getElementById('log') || document.getElementById('fly-log');
    if (el?.dataset.consoleBound === '1') {
      consoleState.follow = el.scrollHeight - el.scrollTop - el.clientHeight < 90;
      consoleState.top = el.scrollTop;
    }
  }
  function bindLog(el) {
    if (!el) return;
    filterRows(el);
    el.scrollTop = consoleState.follow ? el.scrollHeight : consoleState.top;
    if (el.dataset.consoleBound !== '1') {
      el.dataset.consoleBound = '1';
      el.addEventListener('scroll', () => { consoleState.follow = el.scrollHeight - el.scrollTop - el.clientHeight < 90; consoleState.top = el.scrollTop; syncJump(); });
    }
    syncJump();
  }
  function syncJump() {
    document.querySelectorAll('[data-jump-log]').forEach(el => { el.textContent = consoleState.follow ? 'Following latest' : 'Jump to latest'; el.disabled = consoleState.follow; });
  }
  function jump(id) { const el = document.getElementById(id); if (el) { consoleState.follow = true; el.scrollTop = el.scrollHeight; consoleState.top = el.scrollTop; syncJump(); } }
  function appended(el, row, atBottom) {
    if (row) row.hidden = Boolean(consoleState.filter && !row.textContent.toLowerCase().includes(consoleState.filter.toLowerCase()));
    if (atBottom) el.scrollTop = el.scrollHeight;
    syncJump();
  }
  function connectionBlocked() {
    // A successful socket handshake must be followed by a status snapshot.
    const srv = curServer();
    return navigator.onLine === false || failed.has(key(srv)) || Boolean(srv && hasServerPermission('console.view', srv) && S.activeServerId === srv.id && (S.activeNodeId || 'local') === (srv.node_id || 'local') && ['connecting','reconnecting','disconnected'].includes(S.wsState));
  }
  function syncControls() {
    const srv = curServer();
    if (srv && hasServerPermission('console.view', srv) && S.activeServerId === srv.id && (S.activeNodeId || 'local') === (srv.node_id || 'local') && ['connecting','reconnecting','disconnected'].includes(S.wsState)) failed.add(key(srv));
    const disconnected = connectionBlocked();
    const busy = pending.has(key(srv)) || countdowns.has(key(srv)) || (typeof powerRequests !== 'undefined' && powerRequests.has(`${srv?.node_id || 'local'}:${srv?.id}`));
    document.querySelectorAll('[data-power-btn]').forEach(el => {
      const action = el.dataset.powerBtn;
      const stateOK = action === 'start' ? S.status === 'offline' : action === 'restart' ? S.status === 'online' : action === 'stop' ? ['online','starting'].includes(S.status) : ['online','starting','stopping'].includes(S.status);
      el.disabled = disconnected || busy || !stateOK || !hasServerPermission('server.power');
      el.setAttribute('aria-busy', String(busy));
      el.title = disconnected ? 'Reconnect and refresh server status first' : busy ? 'A power action is pending; cancel a countdown before another action' : '';
    });
    document.querySelectorAll('#cmd, #fly-cmd, #fly-send-btn, [onclick^="runCmd("]').forEach(el => { el.disabled = disconnected || S.status !== 'online' || !hasServerPermission('console.command'); });
    const note = document.getElementById('np-connection');
    if (note) { note.hidden = !disconnected; note.querySelector('span').textContent = navigator.onLine === false ? 'You are offline. Server status may be out of date.' : 'Reconnecting… Server status may be out of date. Power and command controls are paused.'; }
  }
  function connectionBanner() { return '<div id="np-connection" class="np-qol-notice" role="status" hidden><span></span><button class="btn sm" onclick="NPWorkspace.retryConnection()">Reconnect &amp; refresh</button></div>'; }
  async function retryConnection() {
    if (retryBusy || !S.token) return;
    retryBusy = true;
    try {
      if (curServer()) {
        await refreshServerStatus();
        if (hasServerPermission('console.view')) connectWs(curServer().id, true, curServer().node_id || 'local');
      } else await refreshDashboard();
    } finally { retryBusy = false; syncControls(); }
  }
  function requestResult(endpoint, error, token) {
    if (token !== S.token) return;
    const srv = curServer(); if (!srv) return;
    const base = srvApiPath(srv, '');
    if (endpoint !== base && !endpoint.startsWith(base + '/')) return;
    // A responding panel/worker reporting a third-party catalog failure is not
    // evidence that the game status connection was lost. Keep real proxy/network
    // failures (which do not carry this typed provider code) on the recovery path.
    if (error?.data?.code === 'plugin_catalog_unavailable' && /^\/plugins\/(project|versions)(?:\?|$)/.test(endpoint.slice(base.length))) return;
    if (error && error.name !== 'AbortError' && (!error.status || [502,503,504].includes(error.status))) failed.add(key(srv));
    else if (!error && (endpoint === base || endpoint === base + '/permissions')) failed.delete(key(srv));
    syncControls();
  }
  function errorAdvice(error) {
    const text = String(error?.message || '').toLowerCase();
    if (!error?.status && /fetch|network|abort|timeout|connection/.test(text) || [502,503,504].includes(error?.status)) return 'Check your connection and refresh status before retrying. The request may already have reached the server.';
    if (error?.status === 403 || /permission|forbidden|access denied/.test(text)) return 'Ask the server owner or administrator to check your permissions, then refresh access.';
    if (error?.status === 409) return 'Refresh server status and wait for the current operation, or cancel its countdown before trying again.';
    if (error?.status === 429) return 'Too many requests. Wait before trying again; do not repeatedly click the action.';
    if (/disk|space|quota/.test(text)) return S.user?.role === 'admin' ? 'Check the host’s free disk space and quota. Free space safely, then retry.' : 'Check your storage usage. Remove unneeded files safely, or contact support for more space.';
    if (/java|runtime/.test(text)) return 'Check your Java version in Settings → Startup and inspect the console for details.';
    if (error?.status === 404) return S.user?.role === 'admin' ? 'Refresh the panel and check that the item exists and the worker is up to date.' : 'Refresh the page. If the item is still unavailable, contact the server owner or support.';
    return 'Inspect the server console for details, then refresh status before retrying.';
  }
  function maintenanceLimited() { return Boolean(S.branding?.maintenance_enabled && S.user?.role !== 'admin'); }
  function maintenanceAllowed(view) { return ['dash','con','acc'].includes(view); }
  function maintenanceBanner() {
    if (!S.branding?.maintenance_enabled) return '';
    const b = S.branding;
    return `<div class="np-maint-banner" role="status"><span><b>${esc(b.maintenance_title || 'Maintenance mode')}</b> — ${esc(b.maintenance_message || '')}<br>Server power and console remain available with your existing permissions.${b.maintenance_eta ? ` Estimated return: ${esc(b.maintenance_eta)}.` : ''}</span>${S.user?.role === 'admin' ? '<div class="row"><button class="btn sm" onclick="toggleMaintPreview(true)">Preview notice</button><a href="/admin/branding" class="btn sm" onclick="return handleNavClick(event, \'admin\', {adminTab:4})">Manage</a></div>' : ''}</div>`;
  }
  function readCountdownResponse(response) {
    const item = response?.countdown;
    if (!response || typeof response !== 'object' || !Object.hasOwn(response, 'countdown') || (item !== null && (!item || !['stop', 'restart'].includes(item.action) || typeof item.execute_at !== 'string' || !Number.isFinite(Date.parse(item.execute_at)) || (item.message !== undefined && typeof item.message !== 'string')))) {
      throw Object.assign(new Error('The countdown API returned an incompatible response.'), { code: 'COUNTDOWN_RESPONSE' });
    }
    return item;
  }
  function countdownFailure(error) {
    const status = Number(error?.status) || 0;
    const admin = S.user?.role === 'admin';
    const http = admin && status ? ` (HTTP ${status})` : '';
    if ([404, 405, 501].includes(status)) return { pause: true, text: admin ? `Countdowns are not supported by this hosting backend${http}. Update and restart the panel and hosting worker, then retry.` : 'Scheduled power actions are unavailable for this server. Contact support.' };
    if (error?.code === 'COUNTDOWN_RESPONSE') return { pause: true, text: admin ? 'The countdown API returned an incompatible response. Update and restart the hosting backend, then retry.' : 'Could not confirm countdown status. Contact support before scheduling another action.' };
    if ([401, 403].includes(status)) return { pause: true, text: `Countdown access denied${http}. Refresh your permissions or ask the server owner for access.` };
    if (status === 429) return { retryAfter: Date.now() + 30000, text: 'Too many countdown requests. Wait 30 seconds before retrying.' };
    if (error?.name === 'AbortError') return { retryAfter: Date.now() + 15000, text: 'Countdown status timed out. Try again shortly.' };
    const detail = String(error?.message || 'Check the hosting node connection.').slice(0, 180);
    return { retryAfter: Date.now() + 15000, text: admin ? `Countdown status is unavailable${http}: ${detail} Retry after checking the hosting backend.` : 'Could not refresh countdown status. Try again shortly.' };
  }
  function countdownSection() {
    const canPower = hasServerPermission('server.power');
    if (!canPower && !hasServerPermission('console.view')) return '';
    return `<details class="np-countdown np-compact-menu"><summary class="btn" title="Schedule a restart or stop with player warnings"><span id="np-countdown-label">Schedule</span><span class="np-menu-chevron" aria-hidden="true"></span></summary><div class="np-menu-panel"><h3>Schedule a restart or stop</h3><div id="np-countdown-state" role="status"></div><button type="button" id="np-countdown-retry" class="btn sm" onclick="NPWorkspace.pollCountdown(true)" hidden>Refresh countdown</button>${canPower && hasServerPermission('console.command') ? `<form onsubmit="event.preventDefault();NPWorkspace.scheduleCountdown(this)"><label>Action<select name="action"><option value="restart">Restart</option><option value="stop">Stop</option></select></label><label>Countdown (seconds)<input name="seconds" type="number" min="10" max="3600" step="1" value="60" required></label><label class="np-countdown-message">Player message<input name="message" maxlength="200" placeholder="Scheduled server maintenance"></label><p class="mu">Warns players before the action. Continues if you close this tab; cancelled if the hosting service restarts.</p><button class="btn" type="submit">Warn players &amp; schedule</button></form>` : ''}</div></details>`;
  }
  function paintCountdown() {
    const el = document.getElementById('np-countdown-state'); if (!el) return;
    const id = key(curServer()), item = countdowns.get(id), error = countdownErrors.get(id);
    const label = document.getElementById('np-countdown-label');
    if (label) label.textContent = item ? `${item.action === 'stop' ? 'Stop' : 'Restart'} · ${Math.max(0, Math.ceil((Date.parse(item.execute_at) - Date.now()) / 1000))}s` : error ? 'Schedule · !' : 'Schedule';
    if (item) {
      const seconds = Math.max(0, Math.ceil((Date.parse(item.execute_at) - Date.now()) / 1000));
      el.innerHTML = `<div class="row sb"><span><b>${esc(item.action === 'stop' ? 'Stop' : 'Restart')}</b> ${seconds ? `in ${seconds}s` : 'is executing…'} · ${esc(item.message || 'Scheduled maintenance')}</span>${hasServerPermission('server.power') ? '<button class="btn sm" onclick="NPWorkspace.cancelCountdown()">Cancel countdown</button>' : ''}</div>${error ? `<p class="mu">${esc(error.text)}</p>` : ''}`;
    } else el.textContent = error?.text || 'No scheduled power countdown.';
    const retry = document.getElementById('np-countdown-retry');
    if (retry) { retry.hidden = !error; retry.disabled = pollBusy || pending.has(id) || Boolean(error?.retryAfter > Date.now()); }
    const section = el.closest('.np-countdown');
    section?.querySelectorAll('form button, form input, form select').forEach(control => { control.disabled = Boolean(item) || Boolean(error) || pending.has(id) || connectionBlocked() || S.status !== 'online'; });
    syncControls();
  }
  async function pollCountdown(force = false) {
    const srv = curServer(), token = S.token;
    if (pollBusy || pending.has(key(srv)) || !srv || !token || S.view !== 'con' || (!hasServerPermission('server.power') && !hasServerPermission('console.view'))) return;
    const previousError = countdownErrors.get(key(srv));
    if (previousError?.retryAfter > Date.now() || (!force && previousError?.pause)) return;
    const revision = revisions.get(key(srv)) || 0;
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), 15000);
    pollBusy = true;
    try {
      const response = await api(srvApiPath(srv, '/power/countdown'), { suppressToast: true, signal: controller.signal });
      if (token !== S.token || key(srv) !== key(curServer()) || revision !== (revisions.get(key(srv)) || 0)) return;
      const item = readCountdownResponse(response);
      countdownErrors.delete(key(srv));
      if (item) countdowns.set(key(srv), item); else countdowns.delete(key(srv));
      paintCountdown();
    } catch (error) {
      if (token === S.token && key(srv) === key(curServer()) && revision === (revisions.get(key(srv)) || 0)) {
        countdownErrors.set(key(srv), countdownFailure(error));
        paintCountdown();
      }
    } finally { clearTimeout(timer); pollBusy = false; if (token === S.token && key(srv) === key(curServer())) paintCountdown(); }
  }
  async function changeCountdown(method, body) {
    const srv = curServer(), token = S.token, id = key(srv);
    if (!srv || pending.has(id) || !hasServerPermission('server.power') || connectionBlocked()) return;
    if (method === 'POST' && (!hasServerPermission('console.command') || S.status !== 'online' || countdownErrors.has(id))) return;
    revisions.set(id, (revisions.get(id) || 0) + 1);
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), 15000);
    pending.add(id); paintCountdown();
    try {
      const response = await api(srvApiPath(srv, '/power/countdown'), { method, ...(body ? { body: JSON.stringify(body) } : {}), suppressToast: true, signal: controller.signal });
      if (token !== S.token) return;
      const item = readCountdownResponse(response);
      countdownErrors.delete(id);
      if (item) countdowns.set(id, item); else countdowns.delete(id);
      toast(method === 'DELETE' ? 'Countdown cancelled' : 'Players warned; countdown scheduled', 'ok');
    } catch (error) {
      if (token === S.token) {
        if (error?.code === 'COUNTDOWN_RESPONSE') countdownErrors.set(id, countdownFailure(error));
        toast(S.user?.role === 'admin' ? `${error.message} ${errorAdvice(error)}` : countdownFailure(error).text, 'bad');
      }
    } finally { clearTimeout(timer); pending.delete(id); if (token === S.token && id === key(curServer())) { paintCountdown(); await pollCountdown(); } }
  }
  function scheduleCountdown(form) {
    if (!form.reportValidity()) return;
    const seconds = Number(form.elements.seconds.value), message = form.elements.message.value, action = form.elements.action.value;
    if (!Number.isInteger(seconds) || seconds < 10 || seconds > 3600 || !['restart','stop'].includes(action) || message.length > 200 || /[\x00-\x1f\x7f]/.test(message)) { toast('Choose 10–3600 seconds and a single-line player message.', 'bad'); return; }
    return changeCountdown('POST', { action, seconds, message: message.trim() });
  }
  function cancelCountdown() { return changeCountdown('DELETE'); }
  function powerBlocked() { return connectionBlocked() || countdowns.has(key(curServer())) || pending.has(key(curServer())); }
  function afterRender() {
    activate(curServer()); remember(); syncControls(); paintCountdown();
    if (S.view === 'dash') {
      const cards = [...document.querySelectorAll('.sift-server-card')];
      const empty = document.getElementById('np-server-empty'); if (empty) empty.hidden = cards.some(c => c.style.display !== 'none');
    }
  }
  function reset() { remember(); user = ''; prefs = {}; bootFiles = {}; active = ''; ready = false; failed.clear(); countdowns.clear(); pending.clear(); revisions.clear(); countdownErrors.clear(); S.hist = []; S.hi = 0; }
  window.NPWorkspace = { loadUser, activate, remember, restoreRoute, restoreFile, saveList, sortServers, favouriteButton, toggleFavourite, listControls, consoleTools, captureConsole, bindLog, appended, setFilter, jump, recordCommand, clearHistory, connectionBlocked, syncControls, connectionBanner, retryConnection, requestResult, errorAdvice, maintenanceLimited, maintenanceAllowed, maintenanceBanner, countdownSection, paintCountdown, pollCountdown, scheduleCountdown, cancelCountdown, powerBlocked, afterRender, reset };
  window.addEventListener('click', event => { document.querySelectorAll('.np-compact-menu[open]').forEach(menu => { if (!menu.contains(event.target)) menu.open = false; }); });
  window.addEventListener('keydown', event => { if (event.key === 'Escape') document.querySelectorAll('.np-compact-menu[open]').forEach(menu => { menu.open = false; menu.querySelector('summary')?.focus(); }); });
  window.addEventListener('offline', syncControls);
  window.addEventListener('online', () => { if (S.token) retryConnection(); });
  window.addEventListener('beforeunload', () => { captureEditor(); remember(); });
  setInterval(() => { if (S.token && S.view === 'con') paintCountdown(); }, 1000);
  setInterval(() => { if (S.token && !document.hidden) pollCountdown(); }, 5000);
})();
