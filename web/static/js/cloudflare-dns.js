(() => {
  'use strict';

  let cfStatus = null;
  let cfLoading = false;
  let cfSyncing = false;

  async function loadCloudflareStatus(serverId) {
    if (!serverId) return;
    try {
      cfLoading = true;
      const res = await api(`/api/servers/${encodeURIComponent(serverId)}/subdomain/cloudflare`);
      cfStatus = res;
    } catch (e) {
      cfStatus = { configured: false, sync: null };
    } finally {
      cfLoading = false;
      updateUI();
    }
  }

  function renderCloudflareSection(srv, draft, fqdn) {
    if (!srv) return '';
    const isSynced = Boolean(cfStatus?.sync?.record_a_id || cfStatus?.sync?.record_srv_id);
    const configured = Boolean(cfStatus?.configured);
    const syncData = cfStatus?.sync;

    return `
      <div class="card" style="padding: 20px; margin-top: 16px; border-left: 3px solid #f38020;">
        <div class="row sb" style="flex-wrap: wrap; gap: 10px; margin-bottom: 12px;">
          <div>
            <h3 style="margin: 0 0 4px; font-size: 16px; color: var(--tx); display: flex; align-items: center; gap: 8px;">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#f38020" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z"/></svg>
              Cloudflare DNS Automation
            </h3>
            <div class="mu" style="font-size: 12.5px;">Automatically create and sync non-proxied <code>A</code>/<code>CNAME</code> and <code>SRV</code> records directly in your Cloudflare account.</div>
          </div>
          <div class="row" style="gap: 8px;">
            ${isSynced ? `<span class="badge online"><i class="dot"></i>Cloudflare Synced</span>` : (configured ? `<span class="badge offline"><i class="dot"></i>Not Synced</span>` : `<span class="badge warn"><i class="dot"></i>API Token Needed</span>`)}
            ${S.user?.role === 'admin' ? `<button type="button" class="btn sm" onclick="window.NPCloudflareDNS.openConfigModal()">${ic('settings', 12)} Cloudflare Settings</button>` : ''}
          </div>
        </div>

        ${isSynced && syncData ? `
          <div style="background: rgba(46, 213, 115, 0.08); border: 1px solid rgba(46, 213, 115, 0.25); border-radius: 8px; padding: 12px 14px; margin-bottom: 14px;">
            <div class="row sb" style="flex-wrap: wrap; gap: 8px;">
              <div>
                <b style="color: var(--tx); font-size: 13px;">Live Cloudflare Records</b>
                <div class="mu mono" style="font-size: 11.5px; margin-top: 2px;">Zone: ${esc(syncData.zone_name || syncData.zone_id)} &bull; Synced: ${esc(new Date(syncData.synced_at).toLocaleString())}</div>
              </div>
              <div class="row" style="gap: 6px;">
                <span class="chip mono sm">A: ${esc(syncData.record_a_id ? syncData.record_a_id.substring(0, 10) + '…' : 'None')}</span>
                <span class="chip mono sm">SRV: ${esc(syncData.record_srv_id ? syncData.record_srv_id.substring(0, 10) + '…' : 'None')}</span>
              </div>
            </div>
          </div>
        ` : ''}

        <div class="row" style="gap: 10px; flex-wrap: wrap;">
          <button type="button" class="btn pr" id="cf-sync-btn" onclick="window.NPCloudflareDNS.syncDNS(this)" ${cfSyncing ? 'disabled' : ''}>
            ${ic('refresh', 14)} ${cfSyncing ? 'Syncing with Cloudflare…' : 'Sync DNS to Cloudflare'}
          </button>
          ${isSynced ? `
            <button type="button" class="btn bad" id="cf-del-btn" onclick="window.NPCloudflareDNS.deleteDNS(this)">
              ${ic('trash', 14)} Remove from Cloudflare
            </button>
          ` : ''}
        </div>
      </div>
    `;
  }

  async function syncDNS(btnEl) {
    const srv = curServer();
    if (!srv) return;
    const draft = S.subdomainDraft || {};
    const fqdn = (draft.label && draft.base_domain)
      ? `${draft.label}.${draft.base_domain}`
      : (draft.base_domain || '');

    if (!fqdn) {
      toast('Please set a subdomain label and base domain first.', 'bad');
      return;
    }

    cfSyncing = true;
    if (btnEl) setBtnLoading(btnEl, true, 'Syncing…');
    try {
      const res = await api(`/api/servers/${encodeURIComponent(srv.id)}/subdomain/cloudflare/sync`, {
        method: 'POST',
        body: JSON.stringify({
          fqdn,
          target_host: draft.target_host || ''
        })
      });
      toast(res?.message || 'DNS successfully synchronized with Cloudflare!', 'ok');
      await loadCloudflareStatus(srv.id);
      R();
    } catch (e) {
      toast(e.message || 'Failed to sync with Cloudflare', 'bad');
      if (e.message && e.message.includes('API token is not configured') && S.user?.role === 'admin') {
        openConfigModal();
      }
    } finally {
      cfSyncing = false;
      if (btnEl && document.body.contains(btnEl)) setBtnLoading(btnEl, false);
    }
  }

  async function deleteDNS(btnEl) {
    const srv = curServer();
    if (!srv) return;
    if (!confirm('Are you sure you want to remove these DNS records from Cloudflare?')) return;

    if (btnEl) setBtnLoading(btnEl, true, 'Removing…');
    try {
      const res = await api(`/api/servers/${encodeURIComponent(srv.id)}/subdomain/cloudflare`, {
        method: 'DELETE'
      });
      toast(res?.message || 'Cloudflare DNS records removed.', 'ok');
      await loadCloudflareStatus(srv.id);
      R();
    } catch (e) {
      toast(e.message || 'Failed to remove Cloudflare records', 'bad');
    } finally {
      if (btnEl && document.body.contains(btnEl)) setBtnLoading(btnEl, false);
    }
  }

  async function openConfigModal() {
    let currentConfig = { configured: false, masked_token: '', zone_id: '' };
    try {
      currentConfig = await api('/api/admin/dns/cloudflare');
    } catch (_) {}

    const html = `
      <div class="modal-box card" style="max-width: 520px; padding: 24px;">
        <div class="row sb" style="margin-bottom: 16px;">
          <h2 style="margin: 0; font-size: 18px; color: var(--tx); display: flex; align-items: center; gap: 8px;">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#f38020" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z"/></svg>
            Cloudflare DNS API Settings
          </h2>
          <button type="button" class="btn ic sm" onclick="closeModal()">${ic('x', 14)}</button>
        </div>
        <p class="mu" style="font-size: 13px; line-height: 1.5; margin: 0 0 16px;">
          Create a Cloudflare API Token with <b>Zone - DNS - Edit</b> and <b>Zone - Zone - Read</b> permissions. NeoPanel will automatically detect your Zone and manage DNS records for your Minecraft servers.
        </p>

        <label class="f" style="margin-bottom: 12px;">
          <span>Cloudflare API Token</span>
          <input type="password" id="cf-modal-token" placeholder="${currentConfig.configured ? currentConfig.masked_token : 'Enter API Token'}" autocomplete="off">
        </label>

        <label class="f" style="margin-bottom: 16px;">
          <span>Zone ID (Optional, auto-detected if blank)</span>
          <input type="text" id="cf-modal-zone" value="${esc(currentConfig.zone_id || '')}" placeholder="Leave blank to detect automatically">
        </label>

        <div id="cf-modal-test-status" style="margin-bottom: 16px; font-size: 12.5px; display: none;"></div>

        <div class="row sb" style="flex-wrap: wrap; gap: 10px;">
          <button type="button" class="btn sm" id="cf-test-btn" onclick="window.NPCloudflareDNS.testToken(this)">${ic('check', 14)} Test Token</button>
          <div class="row" style="gap: 8px;">
            <button type="button" class="btn" onclick="closeModal()">Cancel</button>
            <button type="button" class="btn pr" id="cf-save-btn" onclick="window.NPCloudflareDNS.saveConfig(this)">Save Settings</button>
          </div>
        </div>
      </div>
    `;

    openModal(html);
  }

  async function testToken(btnEl) {
    const tokenInput = document.getElementById('cf-modal-token');
    const token = tokenInput ? tokenInput.value.trim() : '';
    const statusEl = document.getElementById('cf-modal-test-status');

    if (btnEl) setBtnLoading(btnEl, true, 'Testing…');
    if (statusEl) {
      statusEl.style.display = 'block';
      statusEl.innerHTML = `<span class="mu">Testing token with Cloudflare API…</span>`;
    }

    try {
      const res = await api('/api/admin/dns/cloudflare/test', {
        method: 'POST',
        body: JSON.stringify({ token })
      });
      if (statusEl) {
        statusEl.innerHTML = `<div style="color: var(--ok); font-weight: 600;">✓ Valid Token! Found ${res.zones_count || 0} accessible Cloudflare zone(s).</div>`;
      }
    } catch (e) {
      if (statusEl) {
        statusEl.innerHTML = `<div style="color: var(--err); font-weight: 600;">✗ ${esc(e.message || 'Token test failed')}</div>`;
      }
    } finally {
      if (btnEl && document.body.contains(btnEl)) setBtnLoading(btnEl, false);
    }
  }

  async function saveConfig(btnEl) {
    const tokenInput = document.getElementById('cf-modal-token');
    const zoneInput = document.getElementById('cf-modal-zone');
    const token = tokenInput ? tokenInput.value.trim() : '';
    const zoneId = zoneInput ? zoneInput.value.trim() : '';

    if (btnEl) setBtnLoading(btnEl, true, 'Saving…');
    try {
      await api('/api/admin/dns/cloudflare', {
        method: 'POST',
        body: JSON.stringify({ token, zone_id: zoneId })
      });
      toast('Cloudflare settings saved successfully', 'ok');
      closeModal();
      const srv = curServer();
      if (srv) await loadCloudflareStatus(srv.id);
      R();
    } catch (e) {
      toast(e.message || 'Failed to save Cloudflare settings', 'bad');
    } finally {
      if (btnEl && document.body.contains(btnEl)) setBtnLoading(btnEl, false);
    }
  }

  function updateUI() {
    const container = document.getElementById('cf-dns-root');
    if (container && S.view === 'sub') {
      const srv = curServer();
      container.innerHTML = renderCloudflareSection(srv, S.subdomainDraft, getSubdomainPreviewFqdn(S.subdomainDraft));
    }
  }

  window.NPCloudflareDNS = {
    loadStatus: loadCloudflareStatus,
    renderSection: renderCloudflareSection,
    syncDNS,
    deleteDNS,
    openConfigModal,
    testToken,
    saveConfig
  };
})();
