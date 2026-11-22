/* Display-only identities. Never used by update/source-confirmation authorization.
 * One session identity + canonical file signature, bounded to 1000 records / 5 min. */
(() => {
  'use strict';
  const PAGE = 20, MAX = 1000, TTL = 300000;
  const e = v => String(v ?? '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
  const arr = v => Array.isArray(v) ? v : [];
  let cache = null, navigation = '', queued = false, pauses = 0;
  const brokenImages = new Set();
  const painted = new WeakMap();
  let snapshot = null;
  // Batch a synchronous DOM/render pass, avoiding a full list sort for every cell.
  function batch(fn) { const before = snapshot; snapshot = {identity:identity(),scope:scope(),signature:signature()}; try { return fn(); } finally { snapshot = before; } }
  const folder = () => isModdedEngine(curServer()?.type || S.core) ? 'mods' : 'plugins';
  function allowed() { return !!curServer() && !isSteelEngine(curServer().type) && hasServerPermission('files.view') && (!window.NPFeatures || NPFeatures.enabled('plugins')); }
  function scope() {
    if (!allowed()) return '';
    if (S.view === 'plug' && S.plugTab === 'installed' && !window.NPPlugins?.isOpen?.()) return 'installed';
    // The file API uses no trailing slash; accept the equivalent single trailing slash only.
    if (S.view === 'fi' && [folder(), folder() + '/'].includes(S.currentDir) && !S.editingFile && !S.fileTrashOpen && !(typeof filePreviewActive === 'function' && filePreviewActive())) return 'files';
    return '';
  }
  function identity() { const s = curServer(); return JSON.stringify([s?.id,s?.node_id,S.activeNodeId,S.user?.id,S.user?.username,S.user?.role,S.token,s?.type,s?.version,S.core,S.ver]); }
  const jar = name => typeof name === 'string' && /^[^/\\]+\.jar(?:\.disabled)?$/i.test(name);
  function files() {
    return scope() === 'files' ? arr(S.fileList).filter(p => !p.is_dir && jar(p.name) && (!p.path || p.path === folder() + '/' + p.name || p.path === p.name)).map(p => [p.name,p.size ?? '',p.modified || ''])
      : arr(S.plugins).filter(p => jar(p[0])).map(p => [p[0],p[5] ?? p[1] ?? '',p[3] || '']);
  }
  function signature() { return JSON.stringify(files().sort((a,b) => a[0].localeCompare(b[0]))); }
  function paused() { return pauses > 0 || (typeof uploadsPendingFor === 'function' && uploadsPendingFor(curServer(), folder())); }
  function ready() { return scope() && !paused() && !(scope() === 'files' ? S.loading?.files : S.loading?.plugins || arr(S.plugins).some(p => p[4])); }
  function pause() {
    pauses++;
    if (cache?.busy) invalidate();
    let released = false;
    return () => { if (!released) { released = true; pauses--; sync(); } };
  }
  function invalidate() { cache?.controller?.abort(); cache = null; brokenImages.clear(); }
  function sync() {
    const nav = scope(), id = identity();
    if (cache && (cache.identity !== id || !allowed())) invalidate();
    if (navigation !== nav && cache?.busy) invalidate();
    navigation = nav;
    if (!nav || !ready()) { if (cache?.busy) invalidate(); return; }
    const sig = signature();
    if (cache && (cache.signature !== sig || (!cache.busy && !cache.error && Date.now() - cache.time >= TTL))) invalidate();
    if (!cache && !queued) {
      queued = true;
      // Rendering may not yet have committed its nodes. Never replace the workspace.
      Promise.resolve().then(() => { queued = false; if (ready() && !cache) void load(); });
    }
  }
  function current() { return allowed() && ready() && cache?.identity === identity() && cache.signature === (snapshot?.identity === identity() && snapshot.scope === scope() ? snapshot.signature : signature()) ? cache : null; }
  function record(filename) { return current()?.records.get(filename); }
  function safeURL(value) { try { const u = new URL(String(value || '')); return /^https?:$/.test(u.protocol) && !u.username && !u.password ? u.href : ''; } catch (_) { return ''; } }
  const exact = r => ['hash','confirmed'].includes(r?.match) && !!r.project?.id;
  function display(filename, r = record(filename)) {
    const project = exact(r) ? r.project : null;
    return {title:project?.title || r?.name || filename.replace(/\.jar(?:\.disabled)?$/i,''), url:safeURL(project?.icon_url), project, version:r?.installed_version || '', match:r?.match};
  }
  function icon(filename, cls = '', r) { const d = display(filename,r); return `<span class="npid-icon ${e(cls)}" aria-hidden="true">◇${d.url && !brokenImages.has(d.url) ? `<img src="${e(d.url)}" alt="" loading="lazy" referrerpolicy="no-referrer" data-npid-image>` : ''}</span>`; }
  function title(filename, r, installed = false) {
    const d = display(filename,r);
    return `${installed && d.project ? `<button class="npi-project" data-npi="project" data-project="${e(d.project.slug || d.project.id)}">${e(d.title)}</button>` : `<span class="npid-name">${e(d.title)}</span>`}<span class="npid-filename mono mu">${e(filename)}</span>`;
  }
  function info(filename) {
    const d = display(filename);
    return `<span class="npi-status">${d.match === 'hash' ? 'Exact SHA512 match' : d.match === 'confirmed' ? 'Source confirmed · installed release unverified' : d.match === 'error' ? 'Identity unavailable' : d.match ? 'No verified project source' : 'Identity not checked'}</span>${d.version ? `<span class="mu">Descriptor version (informational): ${e(d.version)}</span>` : ''}<span class="mu">Latest for software: Not checked</span>`;
  }
  function fileEligible(item) { return scope() === 'files' && !item.is_dir && jar(item.name) && (!item.path || item.path === folder() + '/' + item.name || item.path === item.name); }
  function fileTitle(item) { return fileEligible(item) ? `<span class="npid-copy" data-npid-title="${e(item.name)}">${title(item.name)}</span>` : e(item.name); }
  function fileIcon(item, fallback, size = '') { return fileEligible(item) ? `<span data-npid-icon="${e(item.name)}" data-npid-size="${e(size)}">${icon(item.name,size)}</span>` : fallback; }
  function feedback() {
    const st = current();
    if (!scope()) return '';
    if (paused()) return '<span>Names and icons will load when the current file operation finishes.</span>';
    if (!st || st.busy) return '<span>Identifying installed files… Names and icons appear as they are verified. Updates are not checked.</span>';
    if (st.error) return '<span>Names and icons could not be fully identified. Files remain usable.</span><button type="button" class="btn sm" data-npid-retry>Retry identities</button>';
    return `<span>${st.records.size} file identities checked${st.more ? ' · 1000-file display limit reached' : ''}. Updates are checked separately.</span>`;
  }
  function statusHtml() { return scope() ? `<div class="npid-feedback" data-npid-feedback role="status" aria-live="polite">${feedback()}</div>` : ''; }
  function paint() { return batch(paintContent); }
  function paintContent() {
    if (!scope()) return;
    for (const [selector, render] of [ ['[data-npid-title]',el => title(el.dataset.npidTitle)], ['[data-npid-installed]',el => title(el.dataset.npidInstalled,undefined,true)], ['[data-npid-icon]',el => icon(el.dataset.npidIcon,el.dataset.npidSize || '')], ['[data-npid-info]',el => info(el.dataset.npidInfo)], ['[data-npid-feedback]',() => feedback()] ]) {
      document.querySelectorAll(selector).forEach(el => { const html = render(el); if (painted.get(el) !== html && el.innerHTML !== html) { const focus = el.contains(document.activeElement); el.innerHTML = html; if (focus) el.querySelector('button')?.focus({preventScroll:true}); } painted.set(el,html); });
    }
  }
  async function load() {
    if (!ready() || cache) return;
    const nav = scope(), srv = {...curServer()}, sig = signature(), id = identity();
    const st = cache = {identity:id,signature:sig,records:new Map(),busy:true,error:false,time:Date.now(),more:false,controller:new AbortController()};
    const names = new Set(files().map(p => p[0]));
    const live = () => cache === st && ready() && nav === scope() && id === identity() && sig === signature() && !st.controller.signal.aborted;
    paint();
    try {
      let offset = 0;
      while (offset < MAX && live() && names.size) {
        const timer = setTimeout(() => st.controller.abort(),75000);
        let data;
        try { data = await api(srvApiPath(srv,`/plugins/catalog?offset=${offset}&limit=${PAGE}&metadata_only=true`),{signal:st.controller.signal,suppressToast:true,isCurrent:live}); }
        finally { clearTimeout(timer); }
        if (!live()) return;
        if (!data || !Array.isArray(data.items) || data.items.length > PAGE || data.offset !== offset || typeof data.has_more !== 'boolean' || !Number.isFinite(data.total) || data.items.some(r => !r || !jar(r.filename) || !['hash','confirmed','suggested','unknown','error'].includes(r.match))) throw Error('Invalid identities');
        for (const r of data.items) if (names.has(r.filename)) {
          // Retain only display fields. Even a malformed backend cannot seed update authority.
          st.records.set(r.filename,{name:String(r.name || '').slice(0,512),installed_version:String(r.installed_version || '').slice(0,256),match:r.match,project:exact(r) ? {id:r.project.id,slug:r.project.slug,title:String(r.project.title || '').slice(0,512),icon_url:safeURL(r.project.icon_url)} : null});
          if (r.match === 'error' || r.message === 'Hash lookup unavailable; retry the scan' || r.message === 'Project metadata unavailable; retry the scan') st.error = true;
        }
        offset += data.items.length; st.more = data.has_more;
        paint();
        if (!data.has_more || !data.items.length) break;
      }
    } catch (_) { if (cache === st && nav === scope() && id === identity() && sig === signature()) st.error = true; }
    finally { if (cache === st) { st.busy = false; st.time = Date.now(); if (st.controller.signal.aborted) st.error = true; paint(); } }
  }
  function retry() { if (!ready() || cache?.busy) return; invalidate(); return load(); }
  document.addEventListener('click',ev => { if (ev.target.closest('[data-npid-retry]')) void retry(); });
  document.addEventListener('error',ev => { if (ev.target.matches?.('[data-npid-image]')) { if (brokenImages.size < MAX) brokenImages.add(ev.target.src); ev.target.remove(); } },true);
  window.NPPluginIdentities = {sync,invalidate,pause,record,display,icon,title,info,fileTitle,fileIcon,statusHtml,safeURL,retry,batch};
})();
