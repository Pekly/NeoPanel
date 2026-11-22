/* Modrinth workspace. Provider text is untrusted: this bounded Markdown subset
 * constructs only allowlisted markup; raw HTML is always text (Prism is not a sanitizer). */
(() => {
  'use strict';
  const PAGE = 24, MAX = 1000;
  const e = value => String(value ?? '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
  const arr = value => Array.isArray(value) ? value : [];
  const text = value => String(value ?? '');
  function url(value) {
    try { const u = new URL(text(value)); return /^(https?):$/.test(u.protocol) && !u.username && !u.password ? u.href : ''; } catch (_) { return ''; }
  }
  function link(label, value) { const href = url(value); return href ? `<a href="${e(href)}" target="_blank" rel="noopener noreferrer">${e(label)}</a>` : e(label); }
  function inline(source) {
    const pattern = /(!?\[([^\]\n]{0,500})\]\(([^\s)]{1,2048})\)|`([^`\n]+)`|\*\*([^*\n]+)\*\*|\*([^*\n]+)\*)/g;
    let html = '', end = 0;
    for (const m of source.matchAll(pattern)) {
      html += e(source.slice(end, m.index));
      if (m[2] !== undefined) {
        const href = url(m[3]);
        html += m[0][0] === '!' ? (href ? `<img src="${e(href)}" alt="${e(m[2])}" loading="lazy" referrerpolicy="no-referrer">` : e(m[2])) : link(m[2], m[3]);
      } else if (m[4]) html += `<code>${e(m[4])}</code>`;
      else html += m[5] ? `<strong>${e(m[5])}</strong>` : `<em>${e(m[6])}</em>`;
      end = m.index + m[0].length;
    }
    return html + e(source.slice(end));
  }
  function markdown(value) {
    const input = text(value), lines = input.slice(0, 100000).split(/\r?\n/).slice(0, 2000);
    let html = '', code = null, list = false;
    for (const raw of lines) {
      const line = raw.slice(0, 5000);
      if (/^\s*```/.test(line)) { if (list) { html += '</ul>'; list = false; } if (code === null) code = []; else { html += `<pre><code>${e(code.join('\n'))}</code></pre>`; code = null; } continue; }
      if (code !== null) { code.push(line); continue; }
      const item = line.match(/^\s*(?:[-*+] |\d+\. )(.*)/);
      if (item) { if (!list) html += '<ul>'; list = true; html += `<li>${inline(item[1])}</li>`; continue; }
      if (list) { html += '</ul>'; list = false; }
      const heading = line.match(/^(#{1,6})\s+(.*)/);
      if (heading) { const n = Math.min(heading[1].length + 1, 6); html += `<h${n}>${inline(heading[2])}</h${n}>`; }
      else if (/^>\s?/.test(line)) html += `<blockquote>${inline(line.replace(/^>\s?/, ''))}</blockquote>`;
      else if (/^\s*---+\s*$/.test(line)) html += '<hr>';
      else if (line.trim()) html += `<p>${inline(line)}</p>`;
    }
    if (list) html += '</ul>';
    if (code !== null) html += `<pre><code>${e(code.join('\n'))}</code></pre>`;
    if (input.length > 100000 || lines.length >= 2000) html += '<p class="mu">Content shortened. Read the full text on Modrinth.</p>';
    return html || '<p class="mu">No content provided by the project.</p>';
  }
  let epoch = 0, detail = null, search = null, timer = null, lastIdentity = null, lastView = null;
  const installing = new Set(), installed = new Set();
  function identity() { const s = curServer(); return JSON.stringify([s?.id, s?.node_id, s?.type, S.user?.id, S.user?.username, S.user?.role, S.token]); }
  function allowed(edit = false) { return !!curServer() && !isSteelEngine(curServer().type) && hasServerPermission(edit ? 'files.edit' : 'files.view') && (!window.NPFeatures || NPFeatures.enabled('plugins')); }
  function capture() { const key = identity(), generation = epoch, srv = {...curServer()}; return {key, srv, live: () => key === identity() && generation === epoch && allowed()}; }
  function reset() { epoch++; clearTimeout(timer); search?.controller?.abort(); detail?.controller?.abort(); search = null; detail = null; installed.clear(); }
  function sync() {
    const key = identity();
    if ((lastIdentity !== null && lastIdentity !== key) || (lastView === 'plug' && S.view !== 'plug')) reset();
    lastIdentity = key; lastView = S.view;
  }
  function queryKey() { return JSON.stringify([identity(), S.q || '', S.plugLoader || '', S.plugVer || '', S.plugSort || 'relevance']); }
  function loaderOptions() {
    const raw = text(curServer()?.type || S.core || 'paper').toLowerCase().trim();
    const own = typeof engineMeta === 'function' ? engineMeta(raw).key : raw;
    const loaders = [...new Set([own, 'paper', 'purpur', 'spigot', 'bukkit', 'folia', 'velocity', 'waterfall', 'bungeecord', 'fabric', 'neoforge', 'forge', 'quilt'])].filter(Boolean);
    return `<option value="" ${!S.plugLoader ? 'selected' : ''}>Loader: Auto (All)</option>` + loaders.map((v, i) => `<option value="${e(v)}" ${S.plugLoader === v ? 'selected' : ''}>${e(typeof engineMeta === 'function' ? engineMeta(v).label : v)}${i === 0 ? ' — Your server' : ''}</option>`).join('');
  }
  const count = n => Number.isFinite(Number(n)) ? Number(n).toLocaleString() : 'Not provided';
  const date = v => v && Number.isFinite(Date.parse(v)) ? new Date(v).toLocaleDateString(undefined, {year:'numeric',month:'short',day:'numeric'}) : 'Not provided';
  const chips = values => arr(values).slice(0, 200).map(v => `<span class="chip">${e(v)}</span>`).join('') || '<span class="mu">Not provided</span>';
  function icon(src, title, cls = '') { const safe = url(src); return safe ? `<img class="np-project-icon ${cls}" src="${e(safe)}" alt="" loading="lazy" referrerpolicy="no-referrer">` : `<span class="np-project-icon ${cls}" aria-hidden="true">${e(text(title).slice(0, 1).toUpperCase() || '◇')}</span>`; }
  function projectKey(slug) {
    const p = detail?.data?.project;
    const h = arr(S.lastModrinthHits).find(h => [h.slug, h.project_id, h.id].includes(slug));
    return text(p && [p.id, p.slug].includes(slug) ? p.id || p.slug : h?.project_id || h?.id || slug);
  }
  const installKey = slug => `${identity()}|${projectKey(slug)}`;
  function installButton(slug, title, version, label = 'Install') {
    const busy = installing.has(installKey(slug)), done = installed.has(installKey(slug));
    return `<button type="button" class="btn pr" data-np-install data-slug="${e(slug)}" data-title="${e(title)}" ${version ? `data-version="${e(version)}"` : ''} ${busy || !allowed(true) ? 'disabled' : ''} aria-busy="${busy}">${busy ? 'Installing…' : done && !version ? 'Install again' : e(label)}</button>`;
  }
  function hitsHtml(hits) {
    const list = arr(hits), mode = getViewMode('plug_store', 'list');
    const cards = list.map((h, i) => {
      const slug = h.slug || h.project_id || h.id || '';
      return `<article class="np-project-result ${mode === 'list' ? 'plugin-card' : mode === 'tiles' ? 'vm-tile-card' : 'vm-icon-card'}" data-project-card data-slug="${e(slug)}" style="--np-item:${Math.min(i, 6)}">
        <button type="button" class="np-project-open" data-np-project data-slug="${e(slug)}" aria-label="View ${e(h.title)}">${icon(h.icon_url, h.title)}<span class="np-project-copy"><b>${e(h.title)}</b>${h.author ? `<span class="mu">by ${e(h.author)}</span>` : ''}<span class="np-project-summary">${e(h.description)}</span></span></button>
        <div class="np-project-result-foot"><span class="mu">↓ ${count(h.downloads)}${h.follows != null ? ` · ${count(h.follows)} followers` : ''}</span>${installButton(slug, h.title)}${installed.has(installKey(slug)) ? '<span class="mu" role="status">Installed · restart required</span>' : ''}</div></article>`;
    }).join('');
    const st = search?.key === queryKey() ? search : null;
    const empty = !list.length && !st?.busy && !st?.error ? '<p class="np-project-empty">No matching projects. Try another search or browsing filter.</p>' : '';
    return cards + empty + `<div class="np-project-pagination" role="status">${st?.busy ? '<span>Loading projects…</span>' : st?.error ? `<p>Search could not be loaded. Your existing results are kept.</p><button class="btn" data-np-search="${st.append ? 'more' : 'retry'}">Retry</button>` : st ? `<span class="mu">${list.length} of ${count(st.total)} projects${list.length >= MAX ? ' · Refine your search to see more' : ''}</span>${st.more && list.length < MAX ? '<button class="btn" data-np-search="more">Load more</button>' : ''}` : ''}</div>`;
  }
  function drawSearch() {
    if (S.view !== 'plug' || activeDetail() || (S.plugTab || 'browse') !== 'browse') return;
    const host = document.getElementById('store-grid');
    if (host) host.innerHTML = hitsHtml(S.lastModrinthHits);
  }
  function debounce() {
    clearTimeout(timer); search?.controller?.abort(); search = null;
    const ctx = capture(), key = queryKey();
    timer = setTimeout(() => { if (ctx.live() && key === queryKey()) void searchProjects(); }, 350);
  }
  async function searchProjects(skipDomWrite = false, append = false) {
    if (!allowed()) return;
    clearTimeout(timer);
    const key = queryKey(), ctx = capture();
    if (append && (!search || search.key !== key || search.busy || !search.more || arr(S.lastModrinthHits).length >= MAX)) return;
    if (!append && search?.key === key && search.busy) return;
    search?.controller?.abort();
    const st = search = {key, busy:true, append, error:false, offset:append ? search.offset : 0, total:append ? search.total : 0, more:true, controller:new AbortController()};
    if (!append) S.lastModrinthHits = [];
    if (!skipDomWrite) drawSearch();
    const timeout = setTimeout(() => st.controller.abort(), 20000);
    try {
      const params = new URLSearchParams({query:S.q || '', type:isModdedEngine(ctx.srv.type || S.core) ? 'mod' : 'plugin', loader:S.plugLoader || '', version:S.plugVer || '', sort:S.plugSort || 'relevance', limit:String(PAGE), offset:String(st.offset)});
      const res = await api(`/api/plugins/search?${params}`, {signal:st.controller.signal, suppressToast:true, isCurrent:() => search === st && ctx.live() && key === queryKey()});
      if (search !== st || !ctx.live() || key !== queryKey()) return;
      if (!Array.isArray(res.hits)) throw Error('Invalid search response');
      const seen = new Set(), all = [...(append ? arr(S.lastModrinthHits) : []), ...res.hits];
      S.lastModrinthHits = all.filter(h => { const id = h.project_id || h.id || h.slug; if (!id || seen.has(id)) return false; seen.add(id); return true; }).slice(0, MAX);
      st.offset += res.hits.length; st.total = Math.max(0, Number(res.total_hits) || 0); st.more = res.hits.length > 0 && st.offset < st.total && st.offset <= 10000;
    } catch (_) { if (search === st && ctx.live() && key === queryKey()) st.error = true; }
    finally { clearTimeout(timeout); if (search === st && ctx.live() && key === queryKey()) { st.busy = false; if (!skipDomWrite) drawSearch(); } }
  }
  function activeDetail() { return detail && detail.ctx.live() && S.view === 'plug' && allowed() ? detail : null; }
  async function open(slug, retry = false) {
    if (!slug || !allowed() || S.view !== 'plug') return;
    const previous = activeDetail(); detail?.controller?.abort();
    const st = detail = {slug, ctx:capture(), tab:'description', busy:true, data:null, error:false, controller:new AbortController(), scroll:retry && previous ? previous.scroll : (document.getElementById('main')?.scrollTop || 0), windowScroll:retry && previous ? previous.windowScroll : window.scrollY};
    R(); document.getElementById('np-project-back')?.focus({preventScroll:true});
    const main = document.getElementById('main'); if (main) main.scrollTop = 0;
    const timeout = setTimeout(() => st.controller.abort(), 20000);
    try {
      const data = await api(srvApiPath(st.ctx.srv, `/plugins/project?slug=${encodeURIComponent(slug)}`), {signal:st.controller.signal, suppressToast:true, isCurrent:() => detail === st && st.ctx.live() && S.view === 'plug'});
      if (detail !== st || !st.ctx.live() || S.view !== 'plug') return;
      if (!data?.project?.id || !Array.isArray(data.versions)) throw Error('Invalid project response');
      st.data = data;
    } catch (_) { if (detail === st && st.ctx.live()) st.error = true; }
    finally { clearTimeout(timeout); if (detail === st && st.ctx.live() && S.view === 'plug') { st.busy = false; R(); document.getElementById('np-project-back')?.focus({preventScroll:true}); } }
  }
  function back() {
    const st = activeDetail(); if (!st) return;
    st.controller.abort(); detail = null; R();
    const main = document.getElementById('main'); if (main) main.scrollTop = st.scroll;
    window.scrollTo(0, st.windowScroll);
    [...document.querySelectorAll('[data-np-project]')].find(el => el.dataset.slug === st.slug)?.focus({preventScroll:true});
  }
  function dependencies(values) {
    const list = arr(values).slice(0, 200);
    return '<h4>Dependencies</h4>' + (list.length ? '<ul>' + list.map(d => {
      const label = d.filename || d.project_id || d.version_id || 'Unspecified dependency';
      const target = d.project_id ? `https://modrinth.com/project/${encodeURIComponent(d.project_id)}` : '';
      return `<li>${link(label, target)} · ${e(d.dependency_type || 'Unspecified type')}${d.version_id ? ` · Version ${e(d.version_id)}` : ''}</li>`;
    }).join('') + '</ul><p class="mu">Dependencies are not installed automatically.</p>' : '<p class="mu">No dependencies declared by the author.</p>');
  }
  function metadata(label, content) { return `<div class="np-project-meta"><h3>${e(label)}</h3><div>${content}</div></div>`; }
  function render() {
    const st = activeDetail(); if (!st) return '';
    const backButton = '<button type="button" id="np-project-back" class="btn sm" data-np-back>← Back to results</button>';
    if (st.busy || st.error) return `<section class="np-project">${backButton}<div class="card np-project-empty" role="status">${st.busy ? 'Loading project from Modrinth…' : '<h2>Project could not be loaded</h2><p>Check your connection and try again.</p><button type="button" class="btn" data-np-retry>Retry</button>'}</div></section>`;
    const d = st.data, p = d.project, eligible = new Set(arr(d.installable_version_ids)), versions = arr(d.versions).slice(0, 500);
    const latest = versions.find(v => eligible.has(v.id));
    const links = [['Source',p.source_url],['Issues',p.issues_url],['Wiki',p.wiki_url],['Discord',p.discord_url],['Modrinth',`https://modrinth.com/project/${encodeURIComponent(p.slug || p.id)}`],...arr(p.donation_urls).map(v => [v.platform || v.id || 'Donate',v.url])].filter(([,v]) => url(v));
    let body;
    if (st.tab === 'gallery') {
      const gallery = arr(p.gallery).filter(g => url(g.url)).slice(0, 100);
      body = gallery.length ? `<div class="np-project-gallery">${gallery.map(g => `<figure><a href="${e(url(g.url))}" target="_blank" rel="noopener noreferrer"><img src="${e(url(g.url))}" alt="${e(g.title || 'Project screenshot')}" loading="lazy" referrerpolicy="no-referrer"></a>${g.title || g.description ? `<figcaption><b>${e(g.title)}</b><p>${e(g.description)}</p></figcaption>` : ''}</figure>`).join('')}</div>` : '<p class="np-project-empty">No gallery images provided.</p>';
    } else if (st.tab === 'versions') {
      body = `<p class="mu">All provider releases, newest first. Installation is available for your server software; Minecraft version compatibility is not guaranteed.</p>${versions.length ? versions.map(v => `<article class="np-project-version"><div><h3>${e(v.name || v.version_number)}</h3><p class="mu">${e(v.version_number)} · ${e(v.version_type)} · ${date(v.date_published)}</p><div class="np-project-tags">${chips(v.loaders)}${chips(v.game_versions)}</div><details><summary>Release notes, files &amp; dependencies</summary><div class="np-project-markdown">${markdown(v.changelog)}</div>${dependencies(v.dependencies)}<ul>${arr(v.files).map(f => `<li>${link(f.filename, f.url)}${f.size != null ? ` · ${count(f.size)} bytes` : ''}</li>`).join('')}</ul></details></div><div>${eligible.has(v.id) ? installButton(st.slug, p.title, v.id, 'Install selected version') : '<span class="mu">Not for your server software</span>'}</div></article>`).join('') : '<p class="np-project-empty">No releases provided.</p>'}`;
      if (d.versions_truncated) body = '<p class="np-project-notice">Showing up to 500 recent releases, including the newest for your server software. See Modrinth for the complete history.</p>' + body.replace('All provider releases', 'Provider releases');
    } else if (st.tab === 'changelog') {
      body = versions.length ? `<div class="np-project-markdown"><h2>${e(versions[0].name || versions[0].version_number)}</h2><p class="mu">Latest published release · ${date(versions[0].date_published)}</p>${markdown(versions[0].changelog)}</div>` : '<p class="np-project-empty">No releases provided.</p>';
    } else body = `<div class="np-project-markdown">${markdown(p.body)}</div>`;
    return `<section class="np-project">${backButton}<header class="card np-project-hero">${icon(p.icon_url,p.title)}<div class="np-project-copy"><div class="np-eyebrow">Modrinth · ${e(p.project_type || 'Project')}</div><h1>${e(p.title)}</h1><p>${e(p.description)}</p><div class="np-project-stats"><span>↓ ${count(p.downloads)} downloads</span><span>♡ ${count(p.followers)} followers</span></div></div><div class="np-project-install">${latest ? installButton(st.slug,p.title,null,'Install latest') : '<button class="btn pr" disabled>No eligible release</button>'}<span class="mu">For ${e(d.loader || curServer()?.type)}</span>${installed.has(installKey(st.slug)) ? '<span role="status">Installed · restart required</span>' : ''}</div></header>
      <p class="np-project-notice">Install latest chooses the newest release for your server software, not Minecraft ${e(d.minecraft_version || 'version')}. It may not support your current Minecraft version. Runtime compatibility is not guaranteed. Restart after installing.</p>
      <div class="np-project-layout"><div class="np-project-content"><div class="tabs np-project-tabs" role="tablist" aria-label="Project content">${['description','gallery','changelog','versions'].map(t => `<button type="button" role="tab" id="np-tab-${t}" aria-controls="np-project-panel" aria-selected="${st.tab === t}" tabindex="${st.tab === t ? 0 : -1}" class="${st.tab === t ? 'on' : ''}" data-np-tab="${t}">${t[0].toUpperCase()+t.slice(1)}</button>`).join('')}</div><section id="np-project-panel" class="card" role="tabpanel" aria-labelledby="np-tab-${st.tab}" tabindex="0">${body}</section></div>
      <aside class="card np-project-sidebar" aria-label="Project information">${metadata('Compatibility',`<p>Server: ${e(p.server_side || 'Not provided')} · Client: ${e(p.client_side || 'Not provided')}</p><div class="np-project-tags">${chips(p.loaders)}</div><details><summary>Minecraft versions (${arr(p.game_versions).length})</summary><div class="np-project-tags">${chips(p.game_versions)}</div></details>`)}${metadata('Links',`<div class="np-project-links">${links.map(([label,v]) => link(label,v)).join('')}</div>`)}${metadata('Tags',`<div class="np-project-tags">${chips([...arr(p.categories),...arr(p.additional_categories)])}</div>`)}${metadata('Creators',arr(d.members).length ? arr(d.members).slice(0,100).map(m => `<div class="np-project-member">${icon(m.user?.avatar_url,m.user?.username)}<div><b>${e(m.user?.name || m.user?.username)}</b><p class="mu">${e(m.role)}</p></div></div>`).join('') : '<p class="mu">No creators provided.</p>')}${metadata('Details',`<dl><dt>License</dt><dd>${link(p.license?.name || p.license?.id || 'Not provided',p.license?.url)}</dd><dt>Published</dt><dd>${date(p.published)}</dd><dt>Updated</dt><dd>${date(p.updated)}</dd>${p.status ? `<dt>Status</dt><dd>${e(p.status)}</dd>` : ''}</dl>`)}</aside></div></section>`;
  }
  function redraw() { if (activeDetail()) R(); else drawSearch(); }
  async function installRequest(endpoint, options, live) {
    const controller = new AbortController(), deadline = setTimeout(() => controller.abort(), 60000);
    try { return await api(endpoint, {...options, signal:controller.signal, suppressToast:true, isCurrent:live}); }
    finally { clearTimeout(deadline); }
  }
  async function install(slug, title, _button, versionId) {
    if (!slug || !allowed(true) || S.view !== 'plug') return;
    const ctx = capture(), key = installKey(slug), live = () => ctx.live() && allowed(true) && S.view === 'plug';
    if (installing.has(key)) return;
    installing.add(key); redraw();
    try {
      const result = await installRequest(srvApiPath(ctx.srv, `/plugins/versions?slug=${encodeURIComponent(slug)}`), {}, live);
      if (!live()) return;
      const releases = arr(result.versions), selected = versionId ? releases.find(v => v.id === versionId) : releases[0];
      if (!selected) { await uiAlert({title:'No eligible release',message:`No ${versionId ? 'selected ' : ''}release of <b>${e(title || slug)}</b> supports your server software <b>${e(result.loader)}</b>. Nothing was installed.`}); return; }
      if (!S.skipPlugConfirm) {
        const ok = await uiConfirm({title:versionId ? 'Install selected version' : 'Install latest release', message:`<p>Install <b>${e(title || slug)}</b>${versionId ? ` · ${e(selected.version_number)}` : ' (latest for your software)'} on <b>${e(ctx.srv.name)}</b>?</p><p>Server software: ${e(result.loader)} · Minecraft ${e(result.minecraft_version)}.</p><p class="mu">The newest release for your software may not support your current Minecraft version. Runtime compatibility is not guaranteed. Restart after installation.</p>`,confirmText:'Download & Install',cancelText:'Cancel'});
        if (!ok || !live()) return;
      }
      if (!live()) return;
      const payload = {slug}; if (versionId) payload.version_id = versionId;
      await installRequest(srvApiPath(ctx.srv, '/plugins/install'), {method:'POST',body:JSON.stringify(payload)}, live);
      if (!live()) return;
      installed.add(key); toast(`Installed ${title || slug}. Restart server to apply.`, 'ok');
      await loadPluginsList();
    } catch (_) { if (live()) toast('Installation could not be confirmed. Check installed files before trying again.', 'bad'); }
    finally { installing.delete(key); if (live()) redraw(); }
  }
  document.addEventListener('click', event => {
    const target = event.target.closest('button, [data-project-card]'); if (!target || S.view !== 'plug') return;
    if (target.hasAttribute('data-np-install')) { void install(target.dataset.slug,target.dataset.title,target,target.dataset.version); return; }
    if (target.hasAttribute('data-np-project')) { void open(target.dataset.slug); return; }
    if (target.hasAttribute('data-project-card') && !event.target.closest('a,input,select,button')) { void open(target.dataset.slug); return; }
    if (target.hasAttribute('data-np-back')) back();
    if (target.hasAttribute('data-np-retry') && activeDetail()) void open(detail.slug,true);
    if (target.hasAttribute('data-np-search')) void searchProjects(false,target.dataset.npSearch === 'more');
    if (target.hasAttribute('data-np-tab') && activeDetail()) { detail.tab = target.dataset.npTab; R(); document.getElementById(`np-tab-${detail.tab}`)?.focus({preventScroll:true}); }
  });
  document.addEventListener('keydown', event => {
    if (!event.target.matches('[data-np-tab]') || !activeDetail()) return;
    const tabs = [...document.querySelectorAll('[data-np-tab]')], i = tabs.indexOf(event.target);
    const next = event.key === 'ArrowRight' ? (i+1)%tabs.length : event.key === 'ArrowLeft' ? (i+tabs.length-1)%tabs.length : event.key === 'Home' ? 0 : event.key === 'End' ? tabs.length-1 : -1;
    if (next >= 0) { event.preventDefault(); tabs[next].click(); }
  });
  window.NPPlugins = {render, open, back, install, hitsHtml, loaderOptions, search:searchProjects, debounce, reset, sync, capture, markdown, safeURL:url};
})();
