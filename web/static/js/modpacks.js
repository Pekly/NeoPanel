(() => {
  'use strict';

  let modpacksData = null;
  let installedModpack = null;
  let loading = false;
  let installing = false;
  let searchQuery = '';
  let searchLoader = '';
  let searchCategory = '';
  let searchTimer = null;

  async function loadModpacks(srv, force = false) {
    if (!srv) return;
    if (modpacksData && !force) return;

    loading = true;
    try {
      const installedPromise = api(`/api/servers/${encodeURIComponent(srv.id)}/modpack`).catch(() => ({ installed: false }));
      const searchPromise = api(`/api/modpacks/search?query=${encodeURIComponent(searchQuery)}&loader=${encodeURIComponent(searchLoader)}&category=${encodeURIComponent(searchCategory)}&limit=18`);

      const [inst, search] = await Promise.all([installedPromise, searchPromise]);
      installedModpack = inst;
      modpacksData = search;
    } catch (e) {
      modpacksData = { hits: [], total_hits: 0 };
    } finally {
      loading = false;
      render();
    }
  }

  function renderModpacksTab(srv) {
    if (!srv) return '';

    const hits = modpacksData?.hits || [];
    const isInst = Boolean(installedModpack?.installed && installedModpack?.metadata);
    const meta = installedModpack?.metadata;

    return `
      <div style="margin-top: 14px; display: flex; flex-direction: column; gap: 16px;">
        ${isInst ? `
          <div class="card" style="padding: 18px 20px; border-left: 3px solid var(--accent); background: var(--card-bg);">
            <div class="row sb" style="flex-wrap: wrap; gap: 10px;">
              <div>
                <div class="row" style="gap: 8px; margin-bottom: 4px;">
                  <span class="badge online"><i class="dot"></i>Active Modpack</span>
                  <span class="chip mono sm">${esc(meta.loader || 'fabric').toUpperCase()} ${esc(meta.game_version || '')}</span>
                </div>
                <h3 style="margin: 0 0 4px; font-size: 18px; color: var(--tx);">${esc(meta.name || 'Installed Modpack')}</h3>
                <div class="mu" style="font-size: 12.5px;">
                  ${meta.server_mods_count || 0} server mods active &bull; ${meta.client_mods_skipped || 0} client mods excluded &bull; Installed ${esc(new Date(meta.installed_at).toLocaleDateString())}
                </div>
              </div>
              <div class="row" style="gap: 8px;">
                <button type="button" class="btn sm bad" onclick="window.NPModpacks.unlinkModpack()">${ic('trash', 12)} Unlink Modpack</button>
              </div>
            </div>
          </div>
        ` : ''}

        <div class="card" style="padding: 16px 20px;">
          <div class="row sb" style="flex-wrap: wrap; gap: 12px; margin-bottom: 12px;">
            <div style="flex: 1; min-width: 240px; position: relative;">
              <input type="text" id="modpack-search-input" value="${esc(searchQuery)}" placeholder="Search 1-Click Modpacks (e.g. Cobblemon, Better MC, ATM)…" oninput="window.NPModpacks.onSearchInput(this.value)" style="width: 100%;">
            </div>
            <div class="row" style="gap: 8px; flex-wrap: wrap;">
              <select id="modpack-loader-select" onchange="window.NPModpacks.onLoaderChange(this.value)" style="padding: 6px 10px; font-size: 13px;">
                <option value="" ${!searchLoader ? 'selected' : ''}>All Loaders</option>
                <option value="fabric" ${searchLoader === 'fabric' ? 'selected' : ''}>Fabric</option>
                <option value="forge" ${searchLoader === 'forge' ? 'selected' : ''}>Forge</option>
                <option value="neoforge" ${searchLoader === 'neoforge' ? 'selected' : ''}>NeoForge</option>
              </select>
            </div>
          </div>

          <div class="row" style="gap: 6px; flex-wrap: wrap; font-size: 12px;">
            ${['All', 'Adventure', 'Optimization', 'Magic', 'Technology', 'Quests', 'Survival'].map(cat => {
              const active = (cat === 'All' && !searchCategory) || (searchCategory.toLowerCase() === cat.toLowerCase());
              return `<button type="button" class="chip ${active ? 'active' : ''}" onclick="window.NPModpacks.onCategoryChange('${cat === 'All' ? '' : cat}')" style="cursor: pointer; border: 1px solid var(--border);">${esc(cat)}</button>`;
            }).join('')}
          </div>
        </div>

        <div id="modpacks-grid-root">
          ${loading ? `<div class="card">${loaderHtml('Searching Modrinth modpacks…', 'modpacks', 'tiles')}</div>` : (
            hits.length === 0 ? `
              <div class="card mu" style="text-align: center; padding: 36px 20px;">
                <div class="np-empty-icon" aria-hidden="true">${ic('box', 36)}</div>
                <h3 class="np-empty-title">No modpacks found</h3>
                <p>Try searching for a different title or loader.</p>
              </div>
            ` : `
              <div class="vm-tiles-grid" style="grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 14px;">
                ${hits.map(p => renderModpackCard(p)).join('')}
              </div>
            `
          )}
        </div>
      </div>
    `;
  }

  function renderModpackCard(p) {
    const iconUrl = p.icon_url || '';
    const downloads = p.downloads ? (p.downloads > 1000000 ? (p.downloads / 1000000).toFixed(1) + 'M' : (p.downloads > 1000 ? (p.downloads / 1000).toFixed(0) + 'k' : p.downloads)) : '0';
    const categories = (p.categories || []).slice(0, 3);

    return `
      <div class="card vm-tile-card" style="padding: 16px; display: flex; flex-direction: column; justify-content: space-between; gap: 12px;">
        <div style="display: flex; gap: 12px; align-items: flex-start;">
          ${iconUrl ? `
            <img src="${esc(iconUrl)}" alt="${esc(p.title)}" style="width: 48px; height: 48px; border-radius: 8px; object-fit: cover; flex-shrink: 0;" loading="lazy">
          ` : `
            <div style="width: 48px; height: 48px; border-radius: 8px; background: var(--border); display: flex; align-items: center; justify-content: center; flex-shrink: 0;">
              ${ic('box', 24)}
            </div>
          `}
          <div style="min-width: 0; flex: 1;">
            <h4 style="margin: 0 0 3px; font-size: 15px; color: var(--tx); white-space: nowrap; overflow: hidden; text-overflow: ellipsis;">${esc(p.title)}</h4>
            <div class="mu" style="font-size: 11.5px;">by ${esc(p.author || 'Modrinth')} &bull; ${esc(downloads)} downloads</div>
          </div>
        </div>

        <p class="mu" style="font-size: 12.5px; line-height: 1.4; margin: 0; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;">
          ${esc(p.description || '')}
        </p>

        <div class="row sb" style="gap: 8px; margin-top: auto; padding-top: 8px; border-top: 1px solid var(--border);">
          <div class="row" style="gap: 4px; flex-wrap: wrap;">
            ${categories.map(c => `<span class="chip mono sm">${esc(c)}</span>`).join('')}
          </div>
          <button type="button" class="btn sm pr" onclick="window.NPModpacks.confirmInstall('${esc(p.project_id || p.slug)}', '${esc(p.title)}')">
            ${ic('download', 12)} 1-Click Install
          </button>
        </div>
      </div>
    `;
  }

  function onSearchInput(val) {
    searchQuery = val.trim();
    clearTimeout(searchTimer);
    searchTimer = setTimeout(() => {
      loadModpacks(curServer(), true);
    }, 350);
  }

  function onLoaderChange(val) {
    searchLoader = val;
    loadModpacks(curServer(), true);
  }

  function onCategoryChange(val) {
    searchCategory = val;
    loadModpacks(curServer(), true);
  }

  function confirmInstall(identifier, title) {
    const srv = curServer();
    if (!srv) return;

    const html = `
      <div class="modal-box card" style="max-width: 480px; padding: 24px;">
        <div class="row sb" style="margin-bottom: 14px;">
          <h2 style="margin: 0; font-size: 18px; color: var(--tx); display: flex; align-items: center; gap: 8px;">
            ${ic('download', 18)} Install Modpack: ${esc(title)}
          </h2>
          <button type="button" class="btn ic sm" onclick="closeModal()">${ic('x', 14)}</button>
        </div>
        <p class="mu" style="font-size: 13.5px; line-height: 1.5; margin: 0 0 16px;">
          This will download the <b>.mrpack</b> archive, automatically extract server overrides, install compatible server-side mods, and skip client-only mods (such as shaders and client UI).
        </p>

        <div class="cf-proxy-warn" style="margin-bottom: 16px;">
          <b>Requirement:</b> The server must be completely <b>stopped</b> before installation. Existing conflicting mods may be replaced.
        </div>

        <div class="row sb" style="gap: 10px;">
          <button type="button" class="btn" onclick="closeModal()">Cancel</button>
          <button type="button" class="btn pr" id="modpack-do-install-btn" onclick="window.NPModpacks.executeInstall('${esc(identifier)}')">
            Start 1-Click Installation
          </button>
        </div>
      </div>
    `;

    openModal(html);
  }

  async function executeInstall(identifier) {
    const srv = curServer();
    if (!srv) return;

    const btn = document.getElementById('modpack-do-install-btn');
    if (btn) setBtnLoading(btn, true, 'Installing Modpack…');

    try {
      const res = await api(`/api/servers/${encodeURIComponent(srv.id)}/modpack/install`, {
        method: 'POST',
        body: JSON.stringify({ slug: identifier })
      });
      toast(res?.message || 'Modpack installed successfully!', 'ok');
      closeModal();
      await loadModpacks(srv, true);
      R();
    } catch (e) {
      toast(e.message || 'Modpack installation failed', 'bad');
      if (btn) setBtnLoading(btn, false);
    }
  }

  async function unlinkModpack() {
    const srv = curServer();
    if (!srv) return;
    if (!confirm('Unlink this modpack metadata? (Mod files will remain in your mods folder).')) return;

    try {
      await api(`/api/servers/${encodeURIComponent(srv.id)}/modpack`, { method: 'DELETE' });
      toast('Modpack unlinked', 'ok');
      await loadModpacks(srv, true);
      R();
    } catch (e) {
      toast(e.message || 'Failed to unlink modpack', 'bad');
    }
  }

  function render() {
    const root = document.getElementById('modpacks-tab-root');
    if (root && S.view === 'plug' && S.plugTab === 'modpacks') {
      root.innerHTML = renderModpacksTab(curServer());
    }
  }

  window.NPModpacks = {
    load: loadModpacks,
    renderTab: renderModpacksTab,
    onSearchInput,
    onLoaderChange,
    onCategoryChange,
    confirmInstall,
    executeInstall,
    unlinkModpack
  };
})();
