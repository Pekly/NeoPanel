/* Optional tools: UI mirrors the API policy; it is not an authorization boundary. */
(() => {
  'use strict';
  const labels = {
    backups: ['New backups', 'Existing backups can still be downloaded or restored.'],
    schedules: ['Schedules', 'Pauses scheduled actions without deleting schedules.'],
    sftp: ['SFTP access', 'Controls external file connections, including administrators.'],
    sftp_import: ['SFTP import', 'Import files from another host. Active imports may finish.'],
    plugins: ['Plugins and mods', 'Browse and manage installed plugins and mods.'],
    databases: ['Database tools', 'Panel database management; existing databases remain intact.'],
    players: ['Player tools', 'Player lists and moderation actions.'],
    proxy_settings: ['Proxy builder', 'Guided Velocity, BungeeCord and Waterfall settings.'],
    extra_ports: ['Additional ports', 'New allocations only; existing ports remain assigned.'],
    software: ['Software and reinstall', 'Software changes and reinstall actions.'],
    subdomains: ['Subdomains', 'Manage server subdomains; existing DNS is unchanged.'],
    countdowns: ['Power countdowns', 'Scheduled stop/restart countdowns. Direct power remains available.'],
    world_conversion: ['World conversion', 'Convert a copy with Chunker while the server is stopped.']
  };
  const pages = { plug: 'plugins', db: 'databases', pl: 'players', sc: 'schedules', sftp: 'sftp', sub: 'subdomains', ver: 'software' };
  let generation = 0, state = null;
  const server = () => typeof curServer === 'function' ? curServer() : null;
  function context() {
    const srv = server();
    const scope = S.view === 'admin' ? '' : (srv ? srvApiPath(srv, '/features') : '');
    return { key: `${S.user?.id || ''}|${S.token || ''}|${scope}`, url: scope || '/api/admin/features', server: !!scope };
  }
  function valid(view) { return view && view.policy && view.global && view.effective && Object.keys(labels).every(key => typeof view.effective[key] === 'boolean'); }
  function active(st) { return state === st && context().key === st.key; }
  function render(srv) {
    if (S.user?.role !== 'admin') return '';
    return `<section class="card np-features-card" id="np-feature-settings" data-settings-section="advanced"><h2>Optional features</h2><p class="mu">${srv ? 'Inherit the panel setting or disable a tool for this server. Panel-wide off switches always win.' : 'Choose which tools are available. Existing data is kept; individual servers can disable tools too.'}</p><div id="np-feature-form"><p class="mu">Loading feature settings…</p></div></section>`;
  }
  function draw() {
    const host = document.getElementById('np-feature-form'), st = state;
    if (!host || !st || S.user?.role !== 'admin') return;
    if (!st.view) {
      host.innerHTML = `<p class="mu">${esc(st.error || 'Loading feature settings…')}</p>${st.error ? '<button class="btn sm" onclick="NPFeatures.reload()">Retry</button>' : ''}`;
      return;
    }
    const pending = Object.values(st.view.nodes || {}).filter(v => !v).length;
    host.innerHTML = `<div class="np-feature-grid">${Object.entries(labels).map(([key, [name, detail]]) => {
      const value = Object.hasOwn(st.draft, key) ? String(st.draft[key]) : 'inherit';
      const globalOff = st.server && st.view.global[key] === false;
      return `<label class="np-feature-row"><span><b>${esc(name)}</b><small>${esc(detail)}${globalOff ? ' Disabled panel-wide.' : ''}</small></span><select aria-label="${esc(name)}" data-feature-key="${key}" ${st.saving ? 'disabled' : ''}>${st.server ? `<option value="inherit" ${value === 'inherit' ? 'selected' : ''}>Inherit</option>` : ''}<option value="true" ${value === 'true' || (!st.server && value === 'inherit') ? 'selected' : ''}>On</option><option value="false" ${value === 'false' ? 'selected' : ''}>Off</option></select></label>`;
    }).join('')}</div><div class="row np-feature-footer"><button class="btn pr" id="np-feature-save" ${st.saving ? 'disabled' : ''}>${st.saving ? 'Saving…' : 'Save features'}</button><span role="status" class="mu">${esc(st.error || (pending ? `Saved locally · ${pending} node(s) pending. Save again to retry. Offline nodes keep their previous policy until they reconnect.` : st.saved ? 'Features saved' : ''))}</span></div>`;
    host.querySelectorAll('[data-feature-key]').forEach(el => el.addEventListener('change', () => {
      if (el.value === 'inherit') delete st.draft[el.dataset.featureKey]; else st.draft[el.dataset.featureKey] = el.value === 'true';
      st.saved = false;
    }));
    host.querySelector('#np-feature-save')?.addEventListener('click', save);
    if (typeof enhanceSelects === 'function') enhanceSelects(host);
  }
  async function load(st) {
    try {
      const view = await api(st.url, { suppressToast: true });
      if (!active(st)) return;
      if (!valid(view)) throw new Error('Feature settings require an updated panel and worker.');
      st.view = view; st.draft = { ...view.policy }; st.error = ''; st.loadedAt = Date.now();
      R();
    } catch (e) { if (active(st)) { st.error = e.message || 'Feature settings could not be loaded.'; draw(); } }
  }
  function sync() {
    if (!S.user) { reset(); return; }
    if (S.view !== 'admin' && !server()) return;
    const ctx = context();
    if (!state || state.key !== ctx.key) {
      state = { ...ctx, generation: ++generation, draft: {}, saving: false, view: null };
      if (!ctx.server && S.user.role !== 'admin') return;
      void load(state);
    }
  }
  function enabled(key) {
    const st = state;
    return !st || !active(st) || !st.view ? true : st.view.effective[key] !== false;
  }
  function pageEnabled(view) { return !pages[view] || enabled(pages[view]); }
  function apply() {
    const selectors = {
      backups: '[onclick="createBackup()"]',
      sftp: '[onclick*="openSftp"], [onclick*="openSFTP"]',
      proxy_settings: '#proxy-settings',
      extra_ports: '[onclick*="openAddPortModal"]',
      sftp_import: '[onclick*="NPSFTPImport.open"]',
      software: '[onclick*="reinstallServer"], [onclick*="openReinstall"], [onclick*="showReinstall"], [onclick*="showSoftware"]',
      countdowns: '.np-countdown form',
      subdomains: '[onclick*="saveSubdomain"], [onclick*="verifySubdomain"]',
      world_conversion: '[data-world-conversion-open]'
    };
    for (const [key, selector] of Object.entries(selectors)) {
      document.querySelectorAll(selector).forEach(el => {
        el.classList.toggle('np-feature-hidden', !enabled(key));
      });
    }
    document.querySelectorAll('[data-feature-page]').forEach(el => el.classList.toggle('np-feature-hidden', !pageEnabled(el.dataset.featurePage)));
  }
  function mount() { draw(); apply(); }
  async function save() {
    const st = state;
    if (!st || st.saving || !active(st) || S.user?.role !== 'admin') return;
    st.saving = true; st.error = ''; draw();
    try {
      const view = await api(st.url, { method: 'PUT', body: JSON.stringify(st.draft), suppressToast: true });
      if (!active(st)) return;
      if (!valid(view)) throw new Error('The save was not confirmed. Reload settings before retrying.');
      st.view = view; st.draft = { ...view.policy }; st.saved = true;
    } catch (e) { if (active(st)) st.error = e.message || 'Features could not be saved.'; }
    finally { if (active(st)) { st.saving = false; draw(); apply(); } }
  }
  function reload() { if (state) { state.error = ''; void load(state); draw(); } }
  function reset() { generation++; state = null; }
  window.NPFeatures = { render, sync, mount, enabled, pageEnabled, reload, reset };
})();
