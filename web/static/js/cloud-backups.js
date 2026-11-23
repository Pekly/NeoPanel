(() => {
  'use strict';

  let cloudBackups = null;
  let cloudLoading = false;
  let activeTab = 'local'; // 'local' or 'cloud'

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

  async function loadCloudBackups(serverId, force = false) {
    if (!serverId) return;
    if (cloudBackups && !force) return;

    cloudLoading = true;
    try {
      const srv = typeof curServer === 'function' ? curServer() : null;
      const url = (srv && typeof srvApiPath === 'function')
        ? srvApiPath(srv, '/backups/cloud')
        : `/api/servers/${encodeURIComponent(serverId)}/backups/cloud`;
      const res = await api(url);
      cloudBackups = res;
    } catch (e) {
      cloudBackups = { enabled: false, backups: [] };
    } finally {
      cloudLoading = false;
      render();
    }
  }

  function renderTabsHeader(srv) {
    const isCloudEnabled = Boolean(cloudBackups?.enabled);
    const count = (cloudBackups?.backups || []).length;

    return `
      <div class="row sb" style="margin-bottom: 16px; flex-wrap: wrap; gap: 10px;">
        <div class="row" style="gap: 8px;">
          <button type="button" class="btn sm ${activeTab === 'local' ? 'pr' : ''}" onclick="window.NPCloudBackups.setTab('local')">
            ${ic('archive', 14)} Local Backups (${S.bk?.length || 0})
          </button>
          <button type="button" class="btn sm ${activeTab === 'cloud' ? 'pr' : ''}" onclick="window.NPCloudBackups.setTab('cloud')">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z"/></svg>
            Cloud Storage (${count})
          </button>
        </div>
        ${S.user?.role === 'admin' ? `
          <button type="button" class="btn sm" onclick="window.NPCloudBackups.openSettingsModal()">
            ${ic('settings', 13)} Cloud Storage Settings
          </button>
        ` : ''}
      </div>
    `;
  }

  function renderCloudStorageView(srv) {
    if (!srv) return '';

    if (cloudLoading && !cloudBackups) {
      return `<div class="card">${loaderHtml('Loading cloud backups…', 'backups', 'list')}</div>`;
    }

    if (!cloudBackups?.enabled) {
      return `
        <div class="card mu" style="text-align: center; padding: 36px 20px;">
          <div class="np-empty-icon" aria-hidden="true">
            <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z"/></svg>
          </div>
          <h3 class="np-empty-title">Cloud Backups Not Configured</h3>
          <p style="margin-bottom: 16px;">Connect AWS S3, Cloudflare R2, MinIO, or Backblaze B2 to store offsite world backups.</p>
          ${S.user?.role === 'admin' ? `
            <button type="button" class="btn pr sm" onclick="window.NPCloudBackups.openSettingsModal()">
              ${ic('settings', 13)} Configure Cloud Storage
            </button>
          ` : ''}
        </div>
      `;
    }

    const items = cloudBackups.backups || [];
    if (items.length === 0) {
      return `
        <div class="card mu" style="text-align: center; padding: 36px 20px;">
          <div class="np-empty-icon" aria-hidden="true">
            <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z"/></svg>
          </div>
          <h3 class="np-empty-title">No cloud backups found</h3>
          <p>Upload a local backup to cloud storage or enable automatic uploads.</p>
        </div>
      `;
    }

    return `
      <div class="card tw">
        <div class="row sb" style="padding: 14px 18px; border-bottom: 1px solid var(--border);">
          <div>
            <b style="color: var(--tx); font-size: 14px;">Bucket: ${esc(cloudBackups.bucket || '')}</b>
            <div class="mu mono" style="font-size: 11.5px;">Prefix: ${esc(cloudBackups.prefix || '')}</div>
          </div>
          <button type="button" class="btn sm" onclick="window.NPCloudBackups.load(curServer(), true)">${ic('refresh', 12)} Refresh</button>
        </div>
        <table>
          <thead>
            <tr>
              <th>Backup File</th>
              <th>Size</th>
              <th>Created / Uploaded</th>
              <th style="text-align: right;">Actions</th>
            </tr>
          </thead>
          <tbody>
            ${items.map(b => `
              <tr>
                <td>
                  <div class="mono" style="font-weight: 600; color: var(--tx);">${esc(b.name)}</div>
                  <div class="mu mono" style="font-size: 11px;">${esc(b.key)}</div>
                </td>
                <td class="mono mu">${formatBytes(b.size)}</td>
                <td class="mu">${esc(new Date(b.last_modified).toLocaleString())}</td>
                <td style="text-align: right;">
                  <div class="row" style="gap: 6px; justify-content: flex-end;">
                    <button type="button" class="btn sm" onclick="window.NPCloudBackups.restore('${esc(b.name)}')">
                      ${ic('refresh', 12)} Restore
                    </button>
                    <button type="button" class="btn sm bad ic" onclick="window.NPCloudBackups.deleteBackup('${esc(b.name)}')">
                      ${ic('trash', 12)}
                    </button>
                  </div>
                </td>
              </tr>
            `).join('')}
          </tbody>
        </table>
      </div>
    `;
  }

  function renderUploadButton(backupName) {
    if (!cloudBackups?.enabled) return '';
    return `
      <button type="button" class="btn sm" title="Upload copy to Cloud Storage (S3 / R2)" onclick="window.NPCloudBackups.uploadBackup('${esc(backupName)}', this)">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z"/><path d="M12 12v6m0-6 3 3m-3-3-3 3"/></svg>
        Cloud
      </button>
    `;
  }

  async function uploadBackup(name, btnEl) {
    const srv = typeof curServer === 'function' ? curServer() : null;
    if (!srv) return;

    if (btnEl) setBtnLoading(btnEl, true, 'Uploading…');
    try {
      const url = (typeof srvApiPath === 'function')
        ? srvApiPath(srv, `/backups/${encodeURIComponent(name)}/cloud/upload`)
        : `/api/servers/${encodeURIComponent(srv.id)}/backups/${encodeURIComponent(name)}/cloud/upload`;
      const res = await api(url, {
        method: 'POST'
      });
      toast(res?.message || 'Backup uploaded to cloud storage!', 'ok');
      await loadCloudBackups(srv.id, true);
    } catch (e) {
      toast(e.message || 'Failed to upload backup to cloud', 'bad');
    } finally {
      if (btnEl && document.body.contains(btnEl)) setBtnLoading(btnEl, false);
    }
  }

  async function restore(name) {
    const srv = typeof curServer === 'function' ? curServer() : null;
    if (!srv) return;
    if (!confirm(`Are you sure you want to restore ${name} from cloud storage? Current files will be replaced.`)) return;

    try {
      const url = (typeof srvApiPath === 'function')
        ? srvApiPath(srv, `/backups/cloud/${encodeURIComponent(name)}/restore`)
        : `/api/servers/${encodeURIComponent(srv.id)}/backups/cloud/${encodeURIComponent(name)}/restore`;
      const res = await api(url, {
        method: 'POST'
      });
      toast(res?.message || 'Server restored successfully!', 'ok');
      if (typeof R === 'function') R();
    } catch (e) {
      toast(e.message || 'Failed to restore cloud backup', 'bad');
    }
  }

  async function deleteBackup(name) {
    const srv = typeof curServer === 'function' ? curServer() : null;
    if (!srv) return;
    if (!confirm(`Delete cloud backup ${name}?`)) return;

    try {
      const url = (typeof srvApiPath === 'function')
        ? srvApiPath(srv, `/backups/cloud/${encodeURIComponent(name)}`)
        : `/api/servers/${encodeURIComponent(srv.id)}/backups/cloud/${encodeURIComponent(name)}`;
      await api(url, {
        method: 'DELETE'
      });
      toast('Cloud backup deleted', 'ok');
      await loadCloudBackups(srv.id, true);
      if (typeof R === 'function') R();
    } catch (e) {
      toast(e.message || 'Failed to delete cloud backup', 'bad');
    }
  }

  async function openSettingsModal() {
    let cfg = { enabled: false, endpoint: '', bucket: '', region: 'us-east-1', access_key: '', secret_key: '', prefix: 'backups', auto_upload: false, retention_days: 30 };
    try {
      cfg = await api('/api/admin/backups/cloud');
    } catch (e) {
      toast(e.message || 'Failed to load cloud backup settings', 'bad');
    }

    const html = `
      <div class="card pbox" style="max-width: 540px; padding: 24px; max-height: 90vh; overflow-y: auto;">
        <div class="row sb" style="margin-bottom: 16px;">
          <h2 style="margin: 0; font-size: 18px; color: var(--tx); display: flex; align-items: center; gap: 8px;">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z"/></svg>
            S3-Compatible Cloud Storage
          </h2>
          <button type="button" class="btn ic sm" onclick="closeModal()">${ic('x', 14)}</button>
        </div>
        <p class="mu" style="font-size: 13px; line-height: 1.5; margin: 0 0 16px;">
          Store world backups offsite using AWS S3, Cloudflare R2, MinIO, or Backblaze B2.
        </p>

        <label class="row" style="gap: 10px; margin-bottom: 16px; cursor: pointer;">
          <span class="sw">
            <input type="checkbox" id="s3-enabled" ${cfg.enabled ? 'checked' : ''}>
            <i></i>
          </span>
          <span style="font-weight: 600; color: var(--tx);">Enable Cloud Backups</span>
        </label>

        <div class="grid g2" style="gap: 12px; margin-bottom: 12px;">
          <label class="f">
            <span>S3 Endpoint</span>
            <input type="text" id="s3-endpoint" value="${esc(cfg.endpoint || '')}" placeholder="https://<id>.r2.cloudflarestorage.com">
          </label>
          <label class="f">
            <span>Bucket Name</span>
            <input type="text" id="s3-bucket" value="${esc(cfg.bucket || '')}" placeholder="e.g. neopanel-backups">
          </label>
        </div>

        <div class="grid g2" style="gap: 12px; margin-bottom: 12px;">
          <label class="f">
            <span>Region</span>
            <input type="text" id="s3-region" value="${esc(cfg.region || 'us-east-1')}" placeholder="e.g. us-east-1 or auto">
          </label>
          <label class="f">
            <span>Folder Prefix</span>
            <input type="text" id="s3-prefix" value="${esc(cfg.prefix || 'backups')}" placeholder="e.g. backups">
          </label>
        </div>

        <div class="grid g2" style="gap: 12px; margin-bottom: 12px;">
          <label class="f">
            <span>Access Key ID</span>
            <input type="text" id="s3-access" value="${esc(cfg.access_key || '')}" placeholder="Access Key">
          </label>
          <label class="f">
            <span>Secret Access Key</span>
            <input type="password" id="s3-secret" placeholder="${cfg.secret_key ? cfg.secret_key : 'Secret Key'}" autocomplete="off">
          </label>
        </div>

        <div class="grid g2" style="gap: 12px; margin-bottom: 16px;">
          <label class="row" style="gap: 10px; cursor: pointer; align-items: center;">
            <span class="sw">
              <input type="checkbox" id="s3-auto" ${cfg.auto_upload ? 'checked' : ''}>
              <i></i>
            </span>
            <span style="font-size: 13px; color: var(--tx);">Auto-upload newly created backups</span>
          </label>
          <label class="f" style="margin-bottom: 0;">
            <span>Retention (Days, 0 = forever)</span>
            <input type="number" id="s3-retention" min="0" max="365" value="${cfg.retention_days || 30}">
          </label>
        </div>

        <div id="s3-modal-test-status" style="margin-bottom: 16px; font-size: 12.5px; display: none;"></div>

        <div class="row sb" style="flex-wrap: wrap; gap: 10px;">
          <button type="button" class="btn sm" id="s3-test-btn" onclick="window.NPCloudBackups.testConnection(this)">
            ${ic('check', 14)} Test S3 Connection
          </button>
          <div class="row" style="gap: 8px;">
            <button type="button" class="btn" onclick="closeModal()">Cancel</button>
            <button type="button" class="btn pr" id="s3-save-btn" onclick="window.NPCloudBackups.saveSettings(this)">Save Settings</button>
          </div>
        </div>
      </div>
    `;

    openModal(html);
  }

  async function testConnection(btnEl) {
    const endpoint = document.getElementById('s3-endpoint')?.value.trim();
    const bucket = document.getElementById('s3-bucket')?.value.trim();
    const region = document.getElementById('s3-region')?.value.trim();
    const accessKey = document.getElementById('s3-access')?.value.trim();
    const secretKey = document.getElementById('s3-secret')?.value.trim();
    const statusEl = document.getElementById('s3-modal-test-status');

    if (btnEl) setBtnLoading(btnEl, true, 'Testing…');
    if (statusEl) {
      statusEl.style.display = 'block';
      statusEl.innerHTML = `<span class="mu">Testing connection to S3 bucket…</span>`;
    }

    try {
      const res = await api('/api/admin/backups/cloud/test', {
        method: 'POST',
        body: JSON.stringify({
          endpoint,
          bucket,
          region,
          access_key: accessKey,
          secret_key: secretKey
        })
      });
      if (statusEl) {
        statusEl.innerHTML = `<div style="color: var(--ok); font-weight: 600;">✓ ${esc(res?.message || 'Connection successful!')}</div>`;
      }
    } catch (e) {
      if (statusEl) {
        statusEl.innerHTML = `<div style="color: var(--err); font-weight: 600;">✗ ${esc(e.message || 'Connection test failed')}</div>`;
      }
    } finally {
      if (btnEl && document.body.contains(btnEl)) setBtnLoading(btnEl, false);
    }
  }

  async function saveSettings(btnEl) {
    const enabled = document.getElementById('s3-enabled')?.checked || false;
    const endpoint = document.getElementById('s3-endpoint')?.value.trim();
    const bucket = document.getElementById('s3-bucket')?.value.trim();
    const region = document.getElementById('s3-region')?.value.trim();
    const accessKey = document.getElementById('s3-access')?.value.trim();
    const secretKey = document.getElementById('s3-secret')?.value.trim();
    const prefix = document.getElementById('s3-prefix')?.value.trim();
    const autoUpload = document.getElementById('s3-auto')?.checked || false;
    const retentionDays = parseInt(document.getElementById('s3-retention')?.value || '30', 10);

    if (btnEl) setBtnLoading(btnEl, true, 'Saving…');
    try {
      await api('/api/admin/backups/cloud', {
        method: 'POST',
        body: JSON.stringify({
          enabled,
          endpoint,
          bucket,
          region,
          access_key: accessKey,
          secret_key: secretKey,
          prefix,
          auto_upload: autoUpload,
          retention_days: retentionDays
        })
      });
      toast('Cloud storage settings saved!', 'ok');
      closeModal();
      const srv = typeof curServer === 'function' ? curServer() : null;
      if (srv) await loadCloudBackups(srv.id, true);
      if (typeof R === 'function') R();
    } catch (e) {
      toast(e.message || 'Failed to save cloud storage settings', 'bad');
    } finally {
      if (btnEl && document.body.contains(btnEl)) setBtnLoading(btnEl, false);
    }
  }

  function setTab(tab) {
    activeTab = tab;
    const srv = typeof curServer === 'function' ? curServer() : null;
    if (tab === 'cloud' && srv) {
      loadCloudBackups(srv.id);
    }
    if (typeof R === 'function') R();
  }

  function getActiveTab() {
    return activeTab;
  }

  function render() {
    // Triggers redraw if in bk view
    if (typeof S !== 'undefined' && S.view === 'bk' && typeof R === 'function') R();
  }

  window.NPCloudBackups = {
    load: loadCloudBackups,
    renderTabsHeader,
    renderCloudStorageView,
    renderUploadButton,
    uploadBackup,
    restore,
    deleteBackup,
    openSettingsModal,
    testConnection,
    saveSettings,
    setTab,
    getActiveTab
  };
})();
