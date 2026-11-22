/* Installed files remain authoritative. Catalog metadata is opt-in and scoped to
 * an exact session/server/software identity; names never establish provenance. */
(() => {
  'use strict';
  const PAGE = 20, MAX = 1000;
  const e = v => String(v ?? '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
  const arr = v => Array.isArray(v) ? v : [];
  // Deliberately outside disposable view state: navigation is not a rollback.
  const writes = new Map();
  let state = null, generation = 0, lastView = '', lastTab = '';
  function identity() {
    const s = curServer();
    return JSON.stringify([s?.id, s?.node_id, S.user?.id, S.user?.username, S.token]);
  }
  function key() { return JSON.stringify([identity(), S.user?.role, curServer()?.type, curServer()?.version, S.core, S.ver]); }
  function allowed(edit = false) { return !!curServer() && !isSteelEngine(curServer().type) && hasServerPermission(edit ? 'files.edit' : 'files.view') && (!window.NPFeatures || NPFeatures.enabled('plugins')); }
  const active = () => S.view === 'plug' && S.plugTab === 'installed';
  const fingerprint = () => JSON.stringify(arr(S.plugins).map(p => [p[0], p[1], p[2], p[3]]));
  function reset() { generation++; state?.read?.abort(); state = null; }
  function sync() {
    if (state && (state.key !== key() || state.files !== fingerprint() || !allowed() || (lastView === 'plug' && S.view !== 'plug') || (lastTab === 'installed' && S.plugTab !== 'installed'))) reset();
    lastView = S.view; lastTab = S.plugTab;
  }
  function get() {
    sync();
    if (!state) state = {key:key(), files:fingerprint(), records:new Map(), selected:new Set(), offset:0, total:0, more:false, scanned:false, busy:false, error:'', notice:'', review:null, preparing:false};
    return state;
  }
  function capture() {
    const k = key(), n = generation, base = identity(), files = fingerprint(), srv = {...curServer()};
    return {base, srv, live:() => k === key() && n === generation && files === fingerprint() && allowed()};
  }
  function invalidate() { reset(); }
  const blocked = () => !!writes.get(identity())?.blocked;
  const pending = () => !!writes.get(identity())?.pending;
  const canUpdate = () => allowed(true) && S.status === 'offline' && !blocked();
  const authoritative = r => ['hash','confirmed'].includes(r?.match) && !!r.project?.id;
  const eligible = r => authoritative(r) && /^[a-f0-9]{128}$/i.test(r.sha512 || '') && !!r.latest?.id && r.update_available === true;
  function draw(focus) {
    if (!active()) return;
    const host = document.getElementById('np-installed');
    const id = focus || (host?.contains(document.activeElement) ? document.activeElement.id : '');
    if (host) host.outerHTML = render();
    if (id) document.getElementById(id)?.focus({preventScroll:true});
  }
  async function request(ctx, path, options, timeout) {
    const controller = options.controller || new AbortController();
    const {controller: unused, ...rest} = options;
    const timer = setTimeout(() => controller.abort(), timeout);
    try { return await api(srvApiPath(ctx.srv, path), {...rest, signal:controller.signal, suppressToast:true, isCurrent:ctx.live}); }
    finally { clearTimeout(timer); }
  }
  async function scan(more = false) {
    const st = get();
    if (!allowed() || !active() || st.busy || st.preparing || pending() || (more && (blocked() || !st.more || st.records.size >= MAX))) return;
    const ctx = capture(), offset = more ? st.offset : 0;
    st.busy = true; st.error = ''; st.review = null; st.selected.clear();
    if (!more) { st.records.clear(); st.scanned = false; st.offset = 0; st.more = false; }
    st.read = new AbortController(); draw();
    const live = () => ctx.live() && state === st;
    try {
      const data = await request({...ctx, live}, `/plugins/catalog?offset=${offset}&limit=${PAGE}`, {controller:st.read}, 75000);
      if (!live()) return;
      if (!data || !Array.isArray(data.items) || data.items.length > PAGE || data.offset !== offset || typeof data.has_more !== 'boolean' || !Number.isFinite(data.total) || data.items.some(r => !r || typeof r.filename !== 'string' || !['hash','confirmed','suggested','unknown','error'].includes(r.match))) throw Error('Invalid catalog');
      for (const r of data.items) if (st.records.size < MAX) st.records.set(r.filename, r);
      st.offset = offset + data.items.length; st.total = data.total;
      st.more = data.has_more && data.items.length > 0 && st.records.size < MAX;
      st.scanned = true; st.notice = '';
      // A successful offset-zero scan takes the backend lease after any old write.
      if (!more && writes.get(ctx.base)?.blocked && !writes.get(ctx.base)?.pending) writes.delete(ctx.base);
    } catch (_) { if (live()) st.error = 'The update check could not finish. Metadata may be unavailable or another operation may still be active. Check for updates again; no update was retried.'; }
    finally { if (live()) { st.busy = false; st.read = null; draw('npi-check'); } }
  }
  function select(filename, checked) {
    const st = get(), r = st.records.get(filename);
    if (st.busy || st.preparing || !canUpdate() || !eligible(r)) return;
    if (!checked) st.selected.delete(filename);
    else if (st.selected.size < 10) st.selected.add(filename);
    draw();
  }
  function review(filename) {
    const st = get(), r = st.records.get(filename);
    if (!allowed() || st.busy || st.preparing || blocked() || r?.match !== 'suggested') return;
    st.review = {filename, record:r, ctx:capture(), project:''}; draw('npi-candidate-0');
  }
  async function confirmSource() {
    const st = get(), v = st.review;
    if (!v || !v.ctx.live() || !allowed(true) || st.preparing || blocked() || !/^[a-f0-9]{128}$/i.test(v.record.sha512 || '')) return;
    const candidate = arr(v.record.candidates).find(p => p.id === v.project);
    if (!candidate?.id) return;
    st.preparing = true; draw();
    const live = () => v.ctx.live() && state === st && allowed(true) && active();
    let sent = false, success = false;
    try {
      const ok = await uiConfirm({title:'Confirm source', message:`<p>Associate <b>${e(v.filename)}</b> with <b>${e(candidate.title || candidate.id)}</b>?</p><p>A name suggestion is not a verified match. The installed release will remain unverified. Updating this file later replaces any custom or modified JAR with an official upstream release.</p>`, confirmText:'Confirm source', cancelText:'Cancel', isDanger:true});
      if (!ok || !live() || blocked() || st.records.get(v.filename) !== v.record) return;
      writes.set(v.ctx.base, {blocked:true, pending:true}); sent = true;
      const data = await request({...v.ctx, live}, '/plugins/identify', {method:'POST',body:JSON.stringify({filename:v.filename,sha512:v.record.sha512,project_id:candidate.id,confirm:true})}, 75000);
      if (!data || typeof data.message !== 'string') throw Error('Invalid confirmation response');
      writes.delete(v.ctx.base); success = true;
      if (live()) { st.review = null; st.notice = 'Source confirmed. Checking the current file again…'; }
    } catch (err) {
      if (sent) settleFailure(v.ctx.base, err);
      if (live()) {
        st.notice = err?.status === 409 ? 'The file changed. Check for updates again before confirming its source.' : 'Source confirmation could not be verified. Check for updates before trying again.';
        if (sent) { st.records.delete(v.filename); st.selected.delete(v.filename); st.review = null; }
      }
    } finally {
      st.preparing = false;
      if (live()) { draw(); if (success) await scan(); }
    }
  }
  function settleFailure(base, err) {
    if (err?.status >= 400 && err.status < 500) writes.delete(base); // Backend preflight: no writes.
    else writes.set(base, {blocked:true, pending:false});
  }
  function backup(value) {
    if (typeof value !== 'string' || !value || /[\x00-\x1f\\:]/.test(value) || value.startsWith('/') || value.split('/').includes('..')) return '';
    return value;
  }
  async function update() {
    const st = get(), ctx = capture();
    if (!active() || !canUpdate() || st.busy || st.preparing || !st.selected.size || st.selected.size > 10) return;
    const records = [...st.selected].map(f => st.records.get(f));
    if (records.some(r => !eligible(r) || !S.plugins.some(p => p[0] === r.filename && !p[4]))) return;
    const live = () => ctx.live() && state === st && active() && allowed(true);
    const valid = () => live() && S.status === 'offline' && !blocked() && records.every(r => st.records.get(r.filename) === r && st.selected.has(r.filename));
    st.preparing = true; draw(); let sent = false;
    try {
      const preview = records.map(r => `<li><b>${e(r.filename)}</b><br>${r.match === 'confirmed' ? 'Unverified installed release (descriptor: ' + e(r.installed_version || 'not provided') + ')' : e(r.installed_version || 'Not provided')} → <b>${e(r.latest.version_number || r.latest.name || r.latest.id)}</b><br>Advertised Minecraft: ${e(arr(r.latest.game_versions).join(', ') || 'Not provided')}</li>`).join('');
      const ok = await uiConfirm({title:`Update ${records.length} selected file${records.length === 1 ? '' : 's'}`,message:`<div class="npi-preview"><ul>${preview}</ul><p>The latest release for your software may not support your Minecraft version, Java runtime, or dependencies. Dependencies will not be installed automatically. Custom or modified JARs are replaced with official upstream releases.</p><p>The server must remain stopped. Existing JARs are backed up before replacement. Configuration files are untouched and disabled files stay disabled. No automatic restart is performed; start the server yourself after reviewing results.</p></div>`,confirmText:'Back up & update selected',cancelText:'Cancel',isDanger:true});
      if (!ok || !valid()) return;
      const items = records.map(r => ({filename:r.filename,sha512:r.sha512,project_id:r.project.id,version_id:r.latest.id}));
      writes.set(ctx.base, {blocked:true,pending:true}); sent = true; draw();
      const data = await request({...ctx,live}, '/plugins/update', {method:'POST',body:JSON.stringify({items,confirm:true})}, 150000);
      const names = new Set(items.map(r => r.filename));
      if (!data || !Array.isArray(data.results) || data.results.length !== items.length || typeof data.restart_required !== 'boolean' || data.results.some(r => !r || !names.delete(r.filename) || !['updated','failed','skipped'].includes(r.status))) throw Error('Invalid update response');
      writes.set(ctx.base, {blocked:false,pending:false,results:data.results,restart:data.restart_required});
      if (live()) {
        st.selected.clear(); st.records.clear(); st.scanned = false; st.more = false;
        st.notice = 'Update finished. Check for updates again before another update.';
        // A failed list refresh does not make an already validated write uncertain.
        try { await loadPluginsList(); } catch (_) { /* Check for updates remains explicit. */ }
      }
    } catch (err) {
      if (sent) settleFailure(ctx.base, err);
      if (live()) {
        st.selected.clear();
        st.notice = err?.status >= 400 && err.status < 500 ? (err.status === 409 ? 'The server or files changed. Nothing was updated. Check for updates again.' : 'The update was rejected before writing. Check permissions and ensure the server is stopped, then check for updates again.') : 'The update outcome is uncertain. Do not repeat it. Check for updates to reconcile the current files first.';
        st.records.clear(); st.scanned = false; st.more = false;
      }
    } finally { st.preparing = false; if (live()) draw('npi-check'); }
  }
  // Only fixed backend diagnostics are shown; never echo arbitrary provider or host errors.
  const safeMessages = new Set([
    'plugin exceeds file or page byte limit', 'cannot read regular plugin file', 'cannot inspect plugin file',
    'plugin changed or could not be read', 'plugin changed while reading', 'invalid JAR archive',
    'JAR archive resource limit exceeded', 'invalid JAR central directory bounds', 'invalid JAR central directory',
    'invalid JAR central directory entry count', 'invalid or oversized JAR directory', 'cannot inspect JAR',
    'JAR exceeds resource limit', 'plugin descriptor exceeds resource limit', 'invalid plugin descriptor',
    'invalid or oversized plugin descriptor', 'invalid YAML plugin descriptor', 'invalid JSON plugin descriptor',
    'invalid TOML plugin descriptor', 'Cannot read confirmed association',
    'Hash lookup unavailable; retry the scan', 'Project metadata unavailable; retry the scan',
    'Latest compatible release unavailable; retry the scan', 'Project search unavailable; retry the scan',
    'No verified newer compatible release; automatic downgrade is disabled.',
    'cannot retain original plugin backup', 'cannot replace original plugin; backup retained',
    'publication and rollback failed; manual repair required before starting',
    'publication failed; original plugin restored', 'Download failed; original plugin retained.',
    'Downloaded JAR is identical to the installed file; no replacement needed.',
    'Update interrupted; original plugin retained.', 'Plugin changed after preflight; original retained. Rescan.',
    'Unsafe plugin publication path; original retained.', 'Cannot create safe plugin backups; original plugin retained.'
  ]);
  const diagnostic = r => safeMessages.has(r?.message) ? `<span class="mu">${e(r.message)}</span>` : '';
  function status(r) {
    if (!r) return 'Not checked';
    if (r.match === 'suggested') return 'Possible match · source not confirmed';
    if (r.match === 'unknown') return 'Unknown · no source identified';
    if (r.match === 'error') return 'Lookup failed · identification incomplete';
    if (r.match === 'confirmed') return 'Source confirmed · installed release unverified';
    return 'Exact SHA512 match';
  }
  function renderReview(st, r) {
    const v = st.review;
    if (!v || v.filename !== r?.filename) return diagnostic(r);
    return `<fieldset class="npi-review"><legend>Review possible sources</legend><p>A similar name is only a suggestion. Choose a source you trust; nothing is linked automatically.</p>${arr(r.candidates).slice(0,20).map((p,i) => `<label class="npi-candidate"><input id="npi-candidate-${i}" type="radio" name="npi-source" value="${e(p.id)}" ${v.project === p.id ? 'checked' : ''} ${st.preparing ? 'disabled' : ''}><span><b>${e(p.title || p.id)}</b><span class="mu">${e(p.description || '')}</span><span class="mu">Project ID: ${e(p.id)}</span></span></label>`).join('') || '<p>No suggestions available. Check again later.</p>'}<div class="npi-actions"><button class="btn pr sm" data-npi="confirm-source" ${!v.project || !allowed(true) || st.preparing ? 'disabled' : ''}>Confirm source</button><button class="btn sm" data-npi="cancel-review" ${st.preparing ? 'disabled' : ''}>Cancel</button></div></fieldset>`;
  }
  function render() {
    const st = get(), w = writes.get(identity()), edit = allowed(true), mode = getViewMode('plug_inst','list');
    const rows = arr(S.plugins).map((p,i) => ({p,i})).filter(({p}) => !S.plugFilter || String(p[0]).toLowerCase().includes(S.plugFilter.toLowerCase()));
    const lock = st.busy || st.preparing || blocked();
    const body = rows.map(({p,i}) => {
      const r = st.records.get(p[0]), identified = authoritative(r), project = identified ? r.project : null;
      const title = project?.title || r?.name || p[0], url = window.NPPlugins?.safeURL(project?.icon_url) || '';
      const chosen = st.selected.has(p[0]);
      return `<article class="npi-file ${chosen ? 'is-selected' : ''}" style="--npi-item:${Math.min(i,5)}"><div class="npi-file-head"><input id="npi-select-${i}" type="checkbox" data-npi-select="${e(p[0])}" aria-label="Select ${e(p[0])} for update" ${chosen ? 'checked' : ''} ${!eligible(r) || !canUpdate() || lock || p[4] || (!chosen && st.selected.size >= 10) ? 'disabled' : ''}>${url ? `<img class="npi-icon" src="${e(url)}" loading="lazy" referrerpolicy="no-referrer" alt="">` : '<span class="npi-icon" aria-hidden="true">◇</span>'}<div class="npi-copy">${project ? `<button class="npi-project" data-npi="project" data-project="${e(project.slug || project.id)}">${e(title)}</button>` : `<b>${e(title)}</b>`}<span class="mono mu">${e(p[0])}</span><span class="mu">${e(p[1] || 'installed')} · ${p[2] ? 'Enabled' : 'Disabled'}${p[3] ? ` · ${e(formatFileDate(p[3]))}` : ''}</span></div></div><div class="npi-metadata"><span class="npi-status">${status(r)}</span>${identified ? `<dl><dt>${r.match === 'confirmed' ? 'Descriptor (informational)' : 'Installed'}</dt><dd>${e(r.installed_version || 'Not provided')}</dd><dt>Latest for software</dt><dd>${e(r.latest?.version_number || r.latest?.name || (r.latest ? r.latest.id : 'Unavailable'))}</dd></dl><span class="mu">${r.update_available && r.latest ? 'Update available' : r.latest && r.match === 'hash' ? 'No update available' : 'Release comparison unavailable'}</span>` : ''}${r?.match === 'suggested' ? `<button class="btn sm" data-npi="review" data-file="${e(p[0])}" ${lock ? 'disabled' : ''}>Review possible match</button>` : ''}</div><div class="npi-actions"><label class="sw" title="Toggle file state"><input aria-label="Enable ${e(p[0])}" type="checkbox" ${p[2] ? 'checked' : ''} ${p[4] || !edit || lock ? 'disabled' : ''} onchange="togglePlugin(${i}, this.checked)"><i></i></label><button class="btn sm dg" data-plugin="${e(p[0])}" onclick="deletePlugin(this.dataset.plugin)" ${!allowed() || !hasServerPermission('files.delete') || lock ? 'disabled' : ''}>Uninstall</button></div>${renderReview(st,r)}</article>`;
    }).join('');
    const results = w?.results ? `<section class="npi-results" aria-label="Update results"><h3>Update results</h3>${w.results.map(r => `<div><b>${e(r.filename)}</b> · ${e(r.status)}${diagnostic(r) || (r.status !== 'updated' ? '<span class="mu">Not updated. Recheck the file and compatibility before a new attempt.</span>' : '')}${backup(r.backup_path) ? `<span class="mu">Backup: <code>${e(backup(r.backup_path))}</code></span>` : ''}</div>`).join('')}<p>${w.restart ? 'Restart required to apply updates. The server was not started automatically.' : 'No automatic restart was performed.'}</p></section>` : '';
    return `<section id="np-installed" class="npi" aria-label="Installed file metadata"><div class="npi-toolbar"><div><b>Installed sources &amp; updates</b><p class="mu">Check explicitly to identify files by SHA512. Name suggestions need your confirmation.</p></div><button id="npi-check" class="btn" data-npi="check" ${!allowed() || st.busy || st.preparing || pending() ? 'disabled' : ''}>${st.busy ? 'Checking…' : 'Check for updates'}</button></div><div class="npi-bulk"><span>${st.selected.size} / 10 selected</span><button id="npi-update" class="btn pr sm" data-npi="update" ${!canUpdate() || !st.selected.size || lock ? 'disabled' : ''}>Review selected updates</button><button class="btn sm" data-npi="clear" ${!st.selected.size || lock ? 'disabled' : ''}>Clear selection</button><span class="mu">${!edit ? 'Read-only · file editing permission required.' : S.status !== 'offline' ? 'Stop the server before updating. Checking is available while running.' : 'Server stopped · no automatic restart.'}</span></div><div role="status" aria-live="polite" class="npi-feedback">${pending() ? 'Updating metadata or files… Results are available when the operation finishes. Do not repeat this operation.' : blocked() ? 'Previous write outcome is uncertain. Check for updates from the beginning before any new update or source confirmation.' : ''}${st.busy ? ' Scanning up to 20 files…' : ''}${e(st.error || st.notice)}</div>${results}<div class="npi-files npi-${e(mode)}">${S.loading?.plugins ? '<p>Loading installed files…</p>' : body || `<p class="card">${S.plugins.length ? 'No installed files match your filter.' : 'No plugins or mods installed yet. Use Browse Store to install a project.'}</p>`}</div>${st.scanned ? `<div class="npi-pagination"><span class="mu">${st.records.size} of ${st.total} files checked${st.records.size >= MAX ? ' · Scan limit reached (1000 files)' : ''}</span>${st.more ? `<button class="btn sm" data-npi="more" ${st.busy || st.preparing || blocked() ? 'disabled' : ''}>Load more scan</button>` : ''}</div>` : ''}</section>`;
  }
  document.addEventListener('click', ev => {
    const b = ev.target.closest('[data-npi]'); if (!b || b.disabled || !active()) return;
    const action = b.dataset.npi;
    if (action === 'check') void scan();
    if (action === 'more') void scan(true);
    if (action === 'update') void update();
    if (action === 'clear') { get().selected.clear(); draw(); }
    if (action === 'review') review(b.dataset.file);
    if (action === 'cancel-review') { get().review = null; draw(); }
    if (action === 'confirm-source') void confirmSource();
    if (action === 'project') window.NPPlugins?.open(b.dataset.project);
  });
  document.addEventListener('change', ev => {
    if (!active()) return;
    if (ev.target.matches('[data-npi-select]')) select(ev.target.dataset.npiSelect, ev.target.checked);
    if (ev.target.matches('[name="npi-source"]') && get().review && !get().preparing) { get().review.project = ev.target.value; draw(ev.target.id); }
  });
  window.NPInstalledPlugins = {render,sync,reset,invalidate,capture,scan,select,review,confirmSource,update};
})();
