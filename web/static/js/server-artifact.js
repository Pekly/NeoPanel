/* Installed-artifact presentation only. Never changes configured engine or launch policy. */
(function (root) {
  'use strict';
  const known = new Set(['paper', 'purpur', 'leaf', 'folia', 'spigot', 'craftbukkit', 'fabric', 'quilt', 'forge', 'neoforge', 'vanilla', 'velocity', 'waterfall', 'bungeecord']);
  const cache = new Map(), signatures = new Map(), bindings = new WeakMap();
  const cacheTTL = 5 * 60 * 1000;
  let active = null, session = '', generation = 0;
  const text = value => typeof value === 'string' ? value.slice(0, 180) : '';
  const major = value => Number.isInteger(value) && value >= 8 && value <= 100 ? value : null;
  function context(srv) {
    return JSON.stringify([S.user?.id, S.user?.role, S.token, srv?.node_id || 'local', srv?.id, srv?.jar_file || 'server.jar', srv?.type || '', srv?.version || '']);
  }
  function permitted(srv) { return Boolean(srv && S.token && hasServerPermission('files.view', srv)); }
  function eligible() {
    const srv = curServer();
    return Boolean(srv && S.token && serverPagePermission(S.view) && canServerPage(S.view) && permitted(srv) && !isSteelEngine(srv.type));
  }
  function bounded(map, key, value) {
    if (map === cache) value.checked = Date.now();
    map.delete(key); map.set(key, value);
    while (map.size > 24) map.delete(map.keys().next().value);
  }
  function supplied(srv) {
    if (!srv) return null;
    const key = context(srv), old = bindings.get(srv);
    // A local settings change must not relabel the previous JAR as the new file.
    if (old && old.key !== key && old.info === srv.artifact) delete srv.artifact;
    const checked = old && old.info === srv.artifact ? old.checked : Date.now();
    bindings.set(srv, {key, info: srv.artifact, checked});
    if (Date.now() - checked >= cacheTTL) { delete srv.artifact; return null; }
    return srv.artifact && typeof srv.artifact === 'object' ? srv.artifact : null;
  }
  function cached(srv) {
    const key = context(srv), entry = cache.get(key);
    if (entry && Date.now() - entry.checked >= cacheTTL) { cache.delete(key); return null; }
    return entry;
  }
  function info(srv) {
    const summary = supplied(srv);
    return summary || (permitted(srv) ? cached(srv)?.info : null);
  }
  function display(srv) {
    if (isSteelEngine(srv?.type)) return {software: 'steel', name: 'Steel', version: text(srv.version) || 'Unknown version', text: `Steel${srv.version ? ' · ' + text(srv.version) : ''} · Native`};
    const a = info(srv), identified = a && ['identified', 'partial'].includes(a.status);
    const software = identified && known.has(a.software) ? a.software : 'custom';
    const name = software === 'custom' ? 'Custom' : engineMeta(software).label;
    const parts = [name];
    // project_version is a loader/proxy release, never a Minecraft version.
    if (identified && text(a.project_version)) parts.push('Release ' + text(a.project_version));
    if (identified && text(a.minecraft_version)) parts.push('Minecraft ' + text(a.minecraft_version));
    if (identified && (text(a.build_name) || a.build_id)) parts.push('Build ' + (text(a.build_name) || String(a.build_id).slice(0, 80)));
    if (parts.length === 1) parts.push('Unknown version');
    const proxy = ['velocity', 'waterfall', 'bungeecord'].includes(software);
    const version = identified ? (proxy ? text(a.project_version) : text(a.minecraft_version) || text(a.project_version)) : '';
    return {software, name, version: version || 'Unknown version', text: parts.join(' / ')};
  }
  function label(srv, size = 13, detailed = false) {
    const d = display(srv);
    return `${engineIcon(d.software, size)} <span class="artifact-label">${esc(detailed ? d.text : d.name)}</span>`;
  }
  function javaLabel(srv) {
    const a = info(srv);
    return ['identified', 'partial'].includes(a?.status) && major(a?.required_java) ? `Auto (Java ${a.required_java})` : 'Auto (exact Java not identified)';
  }
  function explanation(srv) {
    const a = info(srv), entry = cached(srv);
    if (entry?.state === 'loading') return 'Identifying the installed server file…';
    if (entry?.state === 'error') return 'Identification unavailable. Retry, or choose Java manually. Older workers may not support identification.';
    if (a?.status === 'missing') return 'Server file not found. Install or upload the configured file, then identify it again.';
    if (a?.status === 'error') return 'Could not identify this server file. Retry, or choose Java manually.';
    if (['identified', 'partial'].includes(a?.status) && major(a?.required_java)) return `Exact recommended runtime: Java ${a.required_java}${a.java_source === 'mcjars-build' ? ' · MCJars build metadata' : a.java_source === 'mcjars-version' ? ' · MCJars version metadata' : ' · Embedded metadata'}.`;
    if (major(a?.minimum_java)) return `Bytecode minimum: Java ${a.minimum_java}. This is not an exact runtime recommendation; choose Java manually.`;
    return 'Exact Java requirement not identified. Identify the installed file or choose Java manually.';
  }
  function panel(srv) {
    const a = info(srv), loading = cached(srv)?.state === 'loading';
    return `<section id="server-artifact-panel" class="artifact-note" aria-label="Installed server file">
      <div class="artifact-heading"><b>Installed file</b>${permitted(srv) ? `<button type="button" class="btn sm" onclick="NPServerArtifact.retry()" ${loading ? 'disabled' : ''}>${loading ? 'Identifying…' : 'Identify again'}</button>` : ''}</div>
      <div class="artifact-badge">${label(srv, 13, true)}</div>
      <p class="artifact-message" role="status">${esc(explanation(srv))}</p>
      <p class="mu artifact-config">Configured software: ${esc(engineMeta(srv?.type).label)}${srv?.version ? ' · ' + esc(srv.version) : ''}</p>
      ${permitted(srv) ? `<p class="mu mono artifact-file">${esc(a?.filename || srv?.jar_file || 'server.jar')}</p>` : ''}
    </section>`;
  }
  function paint() {
    const srv = curServer();
    if (!srv || !serverPagePermission(S.view)) return;
    const hero = document.getElementById('srv-hero-engine');
    if (hero) hero.innerHTML = label(srv);
    const version = document.getElementById('srv-hero-version');
    if (version) version.textContent = display(srv).version;
    const panelEl = document.getElementById('server-artifact-panel');
    // Keep the retry button mounted (and keyboard focus intact) during status changes.
    if (panelEl) {
      const button = panelEl.querySelector('button');
      const loading = cached(srv)?.state === 'loading';
      if (button) { button.disabled = loading; button.textContent = loading ? 'Identifying…' : 'Identify again'; button.hidden = !permitted(srv); }
      panelEl.querySelector('.artifact-badge').innerHTML = label(srv, 13, true);
      panelEl.querySelector('.artifact-message').textContent = explanation(srv);
      const config = panelEl.querySelector('.artifact-config');
      if (config) config.textContent = `Configured software: ${engineMeta(srv.type).label}${srv.version ? ' · ' + srv.version : ''}`;
      const file = panelEl.querySelector('.artifact-file');
      if (file) { file.hidden = !permitted(srv); file.textContent = permitted(srv) ? info(srv)?.filename || srv.jar_file || 'server.jar' : ''; }
    }
    if (typeof updateSettingsCmdPreview === 'function') updateSettingsCmdPreview();
  }
  function dispose() {
    generation++;
    if (active) { active.controller.abort(); if (cache.get(active.key)?.state === 'loading') cache.delete(active.key); }
    active = null;
  }
  function sync(force = false) {
    const nextSession = JSON.stringify([S.user?.id, S.user?.role, S.token]);
    if (session !== nextSession) { dispose(); cache.clear(); signatures.clear(); session = nextSession; }
    const srv = curServer(), key = srv && context(srv);
    if (!eligible()) { dispose(); return; }
    if (active && active.key !== key) dispose();
    if (active || cached(srv) || supplied(srv)) return;
    const controller = new AbortController(), seq = ++generation;
    active = {key, controller};
    bounded(cache, key, {state: 'loading', seq});
    const live = () => generation === seq && eligible() && context(curServer()) === key;
    const timer = setTimeout(() => controller.abort(), 15000);
    api(srvApiPath(srv, '/artifact' + (force === true ? '?refresh=true' : '')), {signal: controller.signal, suppressToast: true, optionalMetadata: true, isCurrent: live})
      .then(data => {
        if (!live()) return;
        if (!data || !['identified', 'partial', 'unknown', 'missing', 'error', 'native'].includes(data.status)) throw Error('Invalid artifact response');
        bounded(cache, key, {state: 'ready', info: Object.freeze({...data})});
      }).catch(() => { if (live()) bounded(cache, key, {state: 'error'}); })
      .finally(() => {
        clearTimeout(timer);
        if (generation === seq) { active = null; paint(); }
        else if (cache.get(key)?.seq === seq) cache.delete(key);
      });
  }
  function invalidate(srv) {
    if (!srv) return;
    if (active?.key === context(srv)) dispose();
    cache.delete(context(srv)); delete srv.artifact; bindings.delete(srv);
  }
  function retry() {
    if (!eligible()) return;
    invalidate(curServer()); sync(true); paint();
  }
  function installed(srv, response) {
    invalidate(srv);
    if (response?.artifact) srv.artifact = response.artifact;
    if (curServer()?.id === srv.id && (curServer()?.node_id || 'local') === (srv.node_id || 'local')) { sync(); paint(); }
  }
  function files(srv, dir, rows) {
    const filename = String(srv.jar_file || 'server.jar').replace(/\\/g, '/');
    const parent = filename.includes('/') ? filename.slice(0, filename.lastIndexOf('/')) : '';
    if (String(dir).replace(/^\/+|\/+$/g, '') !== parent) return;
    const row = rows.find(f => f.name === filename.split('/').pop());
    const signature = JSON.stringify(row ? [row.name, row.size, row.modified || row.mod_time] : ['missing']);
    const key = context(srv), old = signatures.get(key), a = info(srv);
    const changed = old !== undefined ? old !== signature : a && (a.status === 'missing' ? Boolean(row) : !row || (a.size != null && a.size !== row.size) || (a.modified && (row.modified || row.mod_time) && a.modified !== (row.modified || row.mod_time)));
    bounded(signatures, key, signature);
    if (changed) invalidate(srv);
  }
  function requestResult(endpoint, options, response, token) {
    if (token !== S.token || (options.method || 'GET').toUpperCase() === 'GET') return;
    // Successful mutations only. Uploads use their directory refresh's file signature.
    const srv = (S.servers || []).find(s => {
      const base = srvApiPath(s, '');
      return endpoint === base || endpoint.startsWith(base + '/');
    });
    if (!srv) return;
    const suffix = endpoint.slice(srvApiPath(srv, '').length);
    if (suffix === '' && response?.id === srv.id && ['jar_file', 'type', 'version'].every(field => response[field] === srv[field])) {
      if (response.artifact) srv.artifact = response.artifact;
      return; // Renaming a server or choosing Java does not change its installed file.
    }
    if (suffix === '' || /^\/(files|software|reinstall|backups\/restore)(\/|\?|$)/.test(suffix)) {
      invalidate(srv);
      if (response?.artifact) srv.artifact = response.artifact;
    }
  }
  function refresh() { sync(); paint(); }
  function reset() { dispose(); cache.clear(); signatures.clear(); session = ''; }
  root.NPServerArtifact = {display, label, javaLabel, panel, sync, dispose, reset, refresh, retry, invalidate, installed, files, requestResult};
})(window);
