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

  function openModal(html) {
    if (typeof window !== 'undefined' && typeof window.openModal === 'function') {
      window.openModal(html);
      return;
    }
    const container = document.getElementById('modal-container') || document.body;
    if (!container) return;
    const trimmed = typeof html === 'string' ? html.trim() : '';
    if (trimmed.startsWith('<div class="pal"') || trimmed.startsWith("<div class='pal'")) {
      container.innerHTML = trimmed;
    } else {
      container.innerHTML = `<div class="pal" onclick="if(event.target===this && typeof closeModal === 'function')closeModal()">${html}</div>`;
    }
  }

  function closeModal() {
    if (typeof window !== 'undefined' && typeof window.closeModal === 'function') {
      window.closeModal();
      return;
    }
    const container = document.getElementById('modal-container');
    if (container) container.innerHTML = '';
  }

  function setBtnLoading(btn, isLoading, text) {
    if (typeof window !== 'undefined' && typeof window.setBtnLoading === 'function') {
      window.setBtnLoading(btn, isLoading, text);
      return;
    }
    if (!btn) return;
    btn.disabled = isLoading;
    if (isLoading && text) {
      btn._origText = btn.innerHTML;
      btn.textContent = text;
    } else if (!isLoading && btn._origText !== undefined) {
      btn.innerHTML = btn._origText;
      delete btn._origText;
    }
  }

  async function loadModpacks(srv, force = false) {
    if (!srv) return;
    if (modpacksData && !force) return;

    loading = true;
    try {
      const instUrl = (typeof srvApiPath === 'function')
        ? srvApiPath(srv, '/modpack')
        : `/api/servers/${encodeURIComponent(srv.id)}/modpack`;
      const installedPromise = api(instUrl).catch(() => ({ installed: false }));
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
      <div style="margin-top: 14px;">
        ${isInst ? `
          <div class="card" style="padding: 18px 20px; margin-bottom: 16px; border-left: 3px solid var(--ac);">
            <div class="row sb" style="flex-wrap: wrap; gap: 12px;">
              <div style="display: flex; gap: 14px; align-items: center;">
                ${meta.icon_url ? `
                  <img src="${esc(meta.icon_url)}" alt="${esc(meta.title)}" style="width: 52px; height: 52px; border-radius: 10px; object-fit: cover;">
                ` : `
                  <div style="width: 52px; height: 52px; border-radius: 10px; background: var(--border); display: flex; align-items: center; justify-content: center;">
                    ${ic('box', 28)}
                  </div>
                `}
                <div>
                  <div class="row" style="gap: 8px; align-items: center;">
                    <h3 style="margin: 0; font-size: 17px; color: var(--tx);">${esc(meta.title)}</h3>
                    <span class="badge online"><i class="dot"></i>Active Modpack</span>
                  </div>
                  <div class="mu" style="font-size: 12.5px; margin-top: 2px;">
                    ${esc(meta.version_name || meta.version_id)} &bull; ${esc(meta.loaders?.join(', ') || 'Any loader')} &bull; Minecraft ${esc(meta.game_versions?.join(', ') || 'Any')}
                  </div>
                </div>
              </div>
              <div class="row" style="gap: 8px;">
                <button type="button" class="btn sm bad" onclick="window.NPModpacks.unlinkModpack()">${ic('trash', 12)} Unlink Modpack</button>
              </div>
            </div>
          </div>
        ` : ''}

        <div class="row sb" style="flex-wrap: wrap; gap: 10px; margin-bottom: 14px;">
          <div class="row" style="gap: 8px; flex: 1; max-width: 450px;">
            <input type="text" id="modpack-search-input" value="${esc(searchQuery)}" placeholder="Search Modrinth modpacks (e.g. Cobblemon, All the Mods, Origins)…" style="flex: 1;" oninput="window.NPModpacks.onSearchInput(this.value)">
          </div>
          <div class="row" style="gap: 8px;">
            <select id="modpack-loader-select" onchange="window.NPModpacks.onLoaderChange(this.value)" style="min-width: 120px;">
              <option value="" ${searchLoader === '' ? 'selected' : ''}>All Loaders</option>
              <option value="fabric" ${searchLoader === 'fabric' ? 'selected' : ''}>Fabric</option>
              <option value="forge" ${searchLoader === 'forge' ? 'selected' : ''}>Forge</option>
              <option value="neoforge" ${searchLoader === 'neoforge' ? 'selected' : ''}>NeoForge</option>
              <option value="quilt" ${searchLoader === 'quilt' ? 'selected' : ''}>Quilt</option>
            </select>
          </div>
        </div>

        <div style="margin-bottom: 16px; overflow-x: auto; white-space: nowrap; padding-bottom: 4px;">
          <div class="row" style="gap: 6px;">
            ${['All', 'Adventure', 'Tech', 'Magic', 'Quests', 'Challenging', 'Multiplayer', 'Optimization'].map(cat => {
              const active = (cat === 'All' && !searchCategory) || searchCategory.toLowerCase() === cat.toLowerCase();
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
      loadModpacks(typeof curServer === 'function' ? curServer() : null, true);
    }, 350);
  }

  function onLoaderChange(val) {
    searchLoader = val;
    loadModpacks(typeof curServer === 'function' ? curServer() : null, true);
  }

  function onCategoryChange(val) {
    searchCategory = val;
    loadModpacks(typeof curServer === 'function' ? curServer() : null, true);
  }

  function confirmInstall(identifier, title) {
    const srv = typeof curServer === 'function' ? curServer() : null;
    if (!srv) return;

    const html = `
      <div class="card pbox" style="max-width: 480px; padding: 24px; max-height: 90vh; overflow-y: auto;">
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
    const srv = typeof curServer === 'function' ? curServer() : null;
    if (!srv) return;

    const btn = document.getElementById('modpack-do-install-btn');
    if (btn) setBtnLoading(btn, true, 'Installing Modpack…');

    try {
      const url = (typeof srvApiPath === 'function')
        ? srvApiPath(srv, '/modpack/install')
        : `/api/servers/${encodeURIComponent(srv.id)}/modpack/install`;
      const res = await api(url, {
        method: 'POST',
        body: JSON.stringify({ slug: identifier })
      });
      toast(res?.message || 'Modpack installed successfully!', 'ok');
      closeModal();
      await loadModpacks(srv, true);
      if (typeof R === 'function') R();
    } catch (e) {
      toast(e.message || 'Modpack installation failed', 'bad');
      if (btn) setBtnLoading(btn, false);
    }
  }

  async function unlinkModpack() {
    const srv = typeof curServer === 'function' ? curServer() : null;
    if (!srv) return;
    if (!confirm('Unlink this modpack metadata? (Mod files will remain in your mods folder).')) return;

    try {
      const url = (typeof srvApiPath === 'function')
        ? srvApiPath(srv, '/modpack')
        : `/api/servers/${encodeURIComponent(srv.id)}/modpack`;
      await api(url, { method: 'DELETE' });
      toast('Modpack unlinked', 'ok');
      await loadModpacks(srv, true);
      if (typeof R === 'function') R();
    } catch (e) {
      toast(e.message || 'Failed to unlink modpack', 'bad');
    }
  }

  function render() {
    const root = document.getElementById('modpacks-tab-root');
    if (root && S.view === 'plug' && S.plugTab === 'modpacks') {
      const srv = typeof curServer === 'function' ? curServer() : null;
      root.innerHTML = renderModpacksTab(srv);
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
