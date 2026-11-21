// Per-node desired settings. Never infer worker state from the enabled checkbox.
(() => {
  'use strict';
  const fields = {
    enabled: ['Enable join to start for this node', 'checkbox'],
    bind_address: ['Bind address (empty = all interfaces; literal IP only)', 'text'],
    offline_motd: ['Offline MOTD (2048 UTF-8 bytes)', 'textarea', 2048],
    version_label: ['Offline version label (128 UTF-8 bytes)', 'text', 128],
    max_players: ['Offline maximum players', 'number', 0, 100000],
    use_server_icon: ['Use each server’s server-icon.png when available', 'checkbox'],
    starting_message: ['Starting disconnect message (4096 UTF-8 bytes)', 'textarea', 4096],
    unavailable_message: ['Unavailable disconnect message (4096 UTF-8 bytes)', 'textarea', 4096],
    legacy_disconnect_colors: ['Horizon-compatible reconnect colors', 'checkbox'],
    cooldown_seconds: ['Start request cooldown (seconds)', 'number', 10, 600]
  };
  const help = {
    enabled: 'Applies only after this node successfully stops a server through idle auto-shutdown, and only if that server allows join-to-start. Manual stops, crashes and new servers never qualify.',
    bind_address: 'Usually leave empty. Advanced: use an IP assigned to this node, not a hostname or game port. This does not change player addresses.',
    offline_motd: 'Shown in the multiplayer server list while offline. Pings never start the server. Up to 2048 UTF-8 bytes.',
    version_label: 'Cosmetic offline label only; does not change the installed Minecraft version. Up to 128 UTF-8 bytes.',
    max_players: 'Offline display only (0–100000). The listener reports zero online players; this does not change the real player limit.',
    use_server_icon: 'Prefer a valid server-icon.png from each server. The uploaded icon below is the fallback; neither changes the online icon.',
    starting_message: 'Sent when a start request is accepted. Ask the player to reconnect after boot; this is not proof the game finished starting. Up to 4096 UTF-8 bytes.',
    unavailable_message: 'Sent when the listener declines a request, for example during cooldown. If no listener is available, Minecraft shows its normal connection error. Up to 4096 UTF-8 bytes.',
    legacy_disconnect_colors: 'Enable for Horizon Rust Edition: its disconnect screen flattens JSON styles. Uses literal legacy colors (16-color palette) for reconnect messages, not the MOTD. Off preserves exact RGB for vanilla clients. Save and apply to the worker.',
    cooldown_seconds: 'Minimum time between accepted start requests per server: 10–600 seconds (default 30). This is not the server boot time.'
  };
  const keys = [...Object.keys(fields), 'icon_data_url'];
  let current = null;
  const admin = () => typeof S !== 'undefined' && S.user?.role === 'admin' && Boolean(S.token);
  const identity = () => `${S.user?.id || ''}|${S.token || ''}`;
  const escape = value => String(value ?? '').replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[c]);
  const bytes = text => new TextEncoder().encode(text).length;
  const copy = config => Object.fromEntries(keys.map(key => [key, config[key]]));
  function validEnvelope(value) {
    return value && typeof value.applied === 'boolean' && typeof value.error === 'string' && value.config && keys.every(key => typeof value.config[key] === (key === 'icon_data_url' ? 'string' : fields[key][1] === 'checkbox' ? 'boolean' : fields[key][1] === 'number' ? 'number' : 'string'));
  }
  function validIP(value) {
    if (!value) return true;
    if (/^(0|[1-9]\d{0,2})(\.(0|[1-9]\d{0,2})){3}$/.test(value)) return value.split('.').every(n => Number(n) <= 255);
    if (!value.includes(':') || !/^[\da-f:.]+$/i.test(value)) return false;
    try { return Boolean(new URL(`http://[${value}]/`).hostname); } catch { return false; }
  }
  function safeIcon(value) {
    return typeof value === 'string' && bytes(value) <= 65536 && /^data:image\/png;base64,[A-Za-z0-9+/]+={0,2}$/.test(value);
  }
  function validate(config) {
    if (!validIP(config.bind_address)) throw Error('Bind address must be empty or a literal IPv4/IPv6 address, without a port.');
    for (const [key, [label, type, min, max]] of Object.entries(fields)) {
      const value = config[key];
      if (type === 'number' && (!Number.isInteger(value) || value < min || value > max)) throw Error(`${label} must be an integer from ${min} to ${max}.`);
      if ((type === 'textarea' || type === 'text') && min && bytes(value) > min) throw Error(`${label}: text exceeds the UTF-8 byte limit.`);
    }
    if (config.icon_data_url && !safeIcon(config.icon_data_url)) throw Error('Icon must be a PNG data URL of at most 64 KiB including its prefix.');
    return config;
  }
  function statusText(saved, unknown) {
    if (unknown) return 'Worker status unknown after request failure. The last confirmed saved settings are shown below; delivery may have completed. Refresh status or explicitly save/retry. A requested disable is not confirmed.';
    if (!saved) return 'Saved settings have not been loaded. No changes have been sent.';
    const desired = saved.config.enabled ? 'enabled' : 'disabled';
    return `Saved on panel: ${desired}. ` + (saved.applied
      ? 'The node acknowledged the configuration. This does not confirm that every port is listening or the game can start.'
      : 'Pending worker delivery. The worker may still use older settings; a pending disable does NOT mean join to start is disabled.');
  }
  function deliveryError(message, status) {
    const text = String(message || 'Could not load or save settings. Your edits are preserved.');
    const code = Number(text.match(/HTTP\s+(\d{3})\b/i)?.[1] || status);
    if ([404, 405].includes(code) && !text.includes('Update and restart')) {
      return `${text}\nUpdate and restart the affected worker as well as the panel. Check that the node URL points to that daemon and its reverse proxy forwards PUT /api/cluster/wake. Then choose Retry saved settings. The response alone cannot distinguish an old worker from a wrong URL or proxy rule.`;
    }
    return text;
  }
  function controls(keys = Object.keys(fields)) {
    return keys.map(key => {
      const [label, type, min, max] = fields[key];
      const id = `wake-${key}`;
      const note = `<small id="${id}-help">${escape(help[key])}</small>`;
      if (type === 'checkbox') return `<div class="wake-control"><label class="wake-check"><input id="${id}" type="checkbox" aria-describedby="${id}-help"> <span>${escape(label)}</span></label>${note}</div>`;
      return `<label class="f wake-control" for="${id}"><span>${escape(label)}</span>${type === 'textarea'
        ? `<textarea id="${id}" rows="3" aria-describedby="${id}-help"></textarea>`
        : `<input id="${id}" type="${type}" aria-describedby="${id}-help" ${type === 'number' ? `min="${min}" max="${max}" step="1" required` : ''}>`}${note}</label>`;
    }).join('');
  }
  const el = (ctx, id) => ctx.root.querySelector(`#wake-${id}`);
  const live = ctx => current === ctx && ctx.root.isConnected && admin() && ctx.owner === identity();
  function runtimeSummary(report, now = Date.now()) {
    if (!report) return { text: 'Not checked', tone: 'unknown' };
    const age = now - Date.parse(report.checked_at);
    if (!Number.isFinite(age) || age < -5000 || age > 10000) return { text: 'Status stale — check listeners', tone: 'unknown' };
    if (report.error) return { text: 'Listener status needs attention', tone: 'warning' };
    if (!report.config_matches) return { text: 'Settings not active yet', tone: 'warning' };
    if (!report.enabled) return { text: 'Disabled', tone: 'unknown' };
    if (!report.coordinator_running) return { text: 'Wake coordinator not running', tone: 'warning' };
    if (report.listeners > 0) {
      const ready = report.servers.some(s => s.state === 'listening');
      const cooling = report.servers.some(s => s.state === 'cooldown');
      return { text: ready ? `Ready — ${report.listeners} listening` : `${cooling ? 'Cooldown' : 'Waiting'} — ${report.listeners} ports bound`, tone: ready ? 'ready' : 'warning' };
    }
    if (report.servers.some(s => ['bind_failed', 'blocked', 'conflict'].includes(s.state))) return { text: 'Needs attention — no listeners', tone: 'warning' };
    return { text: 'Standby — no offline listener', tone: 'unknown' };
  }
  function validRuntime(r) {
    return r && typeof r.checked_at === 'string' && ['enabled', 'coordinator_running', 'config_matches'].every(k => typeof r[k] === 'boolean') && typeof r.config_hash === 'string' && typeof r.error === 'string' && Number.isInteger(r.listeners) && r.listeners >= 0 && r.listeners <= 1024 && Array.isArray(r.servers) && r.servers.length <= 1024 && r.servers.every(s => s && ['id', 'name', 'state', 'detail'].every(k => typeof s[k] === 'string') && Number.isInteger(s.port));
  }
  async function checkRuntime(ctx) {
    if (!live(ctx)) return;
    const sequence = ++ctx.runtimeSequence;
    clearTimeout(ctx.runtimeTimer);
    ctx.runtimeBusy = true; ctx.runtimeError = ''; ctx.runtime = null; paint(ctx);
    try {
      const result = await api(`/api/admin/nodes/${encodeURIComponent(ctx.id)}/wake/status`, { method: 'GET', suppressToast: true });
      if (!live(ctx) || sequence !== ctx.runtimeSequence) return;
      if (!validRuntime(result)) throw Error('Unexpected listener status. Update this panel and worker, then check again.');
      ctx.runtime = result;
      const delay = 10001 - (Date.now() - Date.parse(result.checked_at));
      if (Number.isFinite(delay) && delay > 0 && delay < 16000) ctx.runtimeTimer = setTimeout(() => { if (live(ctx)) paint(ctx); }, delay);
    } catch (error) {
      if (live(ctx) && sequence === ctx.runtimeSequence) {
        ctx.runtimeError = error.message || 'Listener status unavailable';
        if ([404, 405].includes(error.status) || /HTTP\s+(404|405)\b/.test(ctx.runtimeError)) ctx.runtimeError += '\nUpdate and restart the worker, and check that its proxy forwards GET /api/cluster/wake/status. This check does not apply settings or start a game.';
      }
    } finally {
      if (live(ctx) && sequence === ctx.runtimeSequence) { ctx.runtimeBusy = false; paint(ctx); }
    }
  }
  async function applyBrandingPreset(ctx) {
    if (!live(ctx) || ctx.busy || !ctx.loaded) return;
    ctx.busy = true; ctx.error = ''; paint(ctx);
    try {
      const result = await api(`/api/admin/nodes/${encodeURIComponent(ctx.id)}/wake/preset`, { method: 'GET', suppressToast: true });
      if (!live(ctx)) return;
      const textKeys = ['offline_motd', 'version_label', 'starting_message', 'unavailable_message'];
      if (!result?.config || !textKeys.every(k => typeof result.config[k] === 'string' && bytes(result.config[k]) <= fields[k][2])) throw Error('Invalid branding template response. Existing text was preserved.');
      for (const key of textKeys) el(ctx, key).value = result.config[key];
      ctx.dirty = true;
    } catch (error) { if (live(ctx)) ctx.error = error.message || 'Could not load branding template'; }
    finally { if (live(ctx)) { ctx.busy = false; paint(ctx); } }
  }
  function paint(ctx) {
    el(ctx, 'status').textContent = statusText(ctx.saved, ctx.unknown);
    el(ctx, 'error').textContent = ctx.error;
    el(ctx, 'error').hidden = !ctx.error;
    el(ctx, 'panel-state').textContent = ctx.saved ? (ctx.saved.config.enabled ? 'Enabled — saved' : 'Disabled — saved') : 'Not loaded';
    el(ctx, 'worker-state').textContent = ctx.unknown ? 'Unknown' : !ctx.saved ? 'Not loaded' : ctx.saved.applied ? (ctx.id === 'local' ? 'Stored on this host' : 'Last delivery accepted') : 'Pending — not confirmed';
    el(ctx, 'fields').disabled = !ctx.loaded || ctx.busy;
    el(ctx, 'save').disabled = !ctx.loaded || ctx.busy;
    el(ctx, 'retry').disabled = !ctx.saved || ctx.busy;
    el(ctx, 'refresh').disabled = ctx.busy;
    el(ctx, 'save').textContent = ctx.busy ? 'Working…' : 'Save settings';
    el(ctx, 'draft-note').textContent = ctx.dirty ? 'Edited fields are not saved. Refresh status and Retry saved settings do not discard your edits.' : 'Fields show the loaded settings. Changes require Save settings.';
    el(ctx, 'saved').textContent = ctx.saved ? JSON.stringify({ ...ctx.saved.config, icon_data_url: ctx.saved.config.icon_data_url ? '(custom PNG configured)' : '' }, null, 2) : 'Not loaded';
    const runtime = ctx.runtimeBusy ? { text: 'Checking listeners…', tone: 'unknown' } : ctx.runtimeError ? { text: 'Listener status unavailable', tone: 'warning' } : runtimeSummary(ctx.runtime);
    el(ctx, 'runtime-state').textContent = runtime.text;
    el(ctx, 'runtime-state').className = `wake-runtime-state ${runtime.tone}`;
    el(ctx, 'check').disabled = ctx.busy || ctx.runtimeBusy;
    el(ctx, 'preset').disabled = ctx.busy || !ctx.loaded;
    el(ctx, 'runtime-note').textContent = ctx.runtimeError || (ctx.runtime ? `${ctx.runtime.error ? ctx.runtime.error + ' ' : ''}Last reconciliation: ${ctx.runtime.checked_at || 'not available'}. This confirms daemon listener state, not external reachability or successful game startup.` : 'Check listeners reads the node runtime without starting a game.');
    el(ctx, 'runtime-list').innerHTML = ctx.runtime?.servers.length ? ctx.runtime.servers.map(s => `<div class="wake-runtime-row"><b>${escape(s.name || s.id)} <span class="mu">:${s.port}</span></b><span>${escape(s.state.replaceAll('_', ' '))}</span><small>${escape(s.detail)}</small></div>`).join('') : '<p class="mu">No server listener details available.</p>';
  }
  function preview(ctx) {
    const image = el(ctx, 'preview');
    image.hidden = !safeIcon(ctx.icon);
    if (!image.hidden) image.src = ctx.icon;
    else image.removeAttribute('src');
  }
  function fill(ctx, config) {
    for (const [key, [, type]] of Object.entries(fields)) el(ctx, key)[type === 'checkbox' ? 'checked' : 'value'] = config[key];
    ctx.icon = config.icon_data_url; ctx.loaded = true; preview(ctx);
  }
  function draft(ctx) {
    const config = { icon_data_url: ctx.icon };
    for (const [key, [, type]] of Object.entries(fields)) {
      const input = el(ctx, key);
      config[key] = type === 'checkbox' ? input.checked : type === 'number' ? (input.value.trim() === '' ? NaN : Number(input.value)) : input.value;
    }
    return validate(config);
  }
  function accept(ctx, result, initial) {
    if (!validEnvelope(result)) throw Error('Unexpected join-to-start response. Update the panel and worker, then refresh status.');
    ctx.saved = { config: copy(result.config), applied: result.applied, error: result.error };
    ctx.unknown = false; ctx.error = deliveryError(result.error || '', undefined);
    if (!result.error) ctx.error = '';
    if (initial) fill(ctx, result.config);
  }
  async function request(ctx, config) {
    if (!live(ctx) || ctx.busy) return;
    const initial = !ctx.loaded;
    ++ctx.runtimeSequence; clearTimeout(ctx.runtimeTimer); ctx.runtime = null;
    ctx.busy = true; ctx.error = ''; paint(ctx);
    try {
      const result = await api(`/api/admin/nodes/${encodeURIComponent(ctx.id)}/wake`, { method: config ? 'PUT' : 'GET', ...(config ? { body: JSON.stringify(config) } : {}), suppressToast: true });
      if (!live(ctx)) return;
      accept(ctx, result, initial);
      if (config && !ctx.preserveDraft && !result.error) { fill(ctx, result.config); ctx.dirty = false; }
    } catch (error) {
      if (!live(ctx)) return;
      if (validEnvelope(error.data)) accept(ctx, error.data, initial);
      else { ctx.unknown = true; ctx.error = deliveryError(error.message, error.status); }
    } finally {
      if (live(ctx)) { ctx.busy = false; paint(ctx); await checkRuntime(ctx); }
    }
  }
  async function upload(ctx, file) {
    if (!file || !live(ctx) || ctx.busy) return;
    const sequence = ++ctx.uploadSequence;
    ctx.busy = true; paint(ctx);
    try {
      if (file.size > 49152) throw Error('PNG is too large: the complete base64 data URL must fit within 64 KiB.');
      const data = await new Promise((resolve, reject) => {
        const reader = new FileReader(); reader.onload = () => resolve(reader.result); reader.onerror = () => reject(Error('Could not read the PNG.')); reader.readAsDataURL(file);
      });
      if (!safeIcon(data)) throw Error('Choose a PNG whose complete data URL is at most 64 KiB.');
      await new Promise((resolve, reject) => {
        const image = new Image(); image.onload = () => image.naturalWidth === 64 && image.naturalHeight === 64 ? resolve() : reject(Error('Icon must be exactly 64×64 pixels.')); image.onerror = () => reject(Error('Could not decode this PNG.')); image.src = data;
      });
      if (!live(ctx) || sequence !== ctx.uploadSequence) return;
      ctx.icon = data; ctx.dirty = true; ctx.error = ''; preview(ctx);
    } catch (error) { if (live(ctx) && sequence === ctx.uploadSequence) ctx.error = error.message; }
    finally { if (live(ctx) && sequence === ctx.uploadSequence) { ctx.busy = false; el(ctx, 'upload').value = ''; paint(ctx); } }
  }
  async function open(id, returnToEdit = false) {
    if (!admin()) return;
    const container = document.querySelector('#modal-container');
    if (!container) return;
    const name = typeof displayNodeName === 'function' ? displayNodeName(id) : id;
    const opener = document.activeElement;
    // Keep the actual form nodes: serializing HTML would lose unsaved values.
    const previous = returnToEdit ? [...container.childNodes] : null;
    container.innerHTML = `<div class="pal" id="wake-modal"><div class="card pbox wake-dialog" role="dialog" aria-modal="true" aria-labelledby="wake-title">
      <div class="wake-heading"><div><h2 id="wake-title">Join to start — ${escape(name)}</h2><p class="mu">Offline presentation and player-triggered startup for this node.</p></div><button type="button" class="btn sm" id="wake-close" aria-label="${previous ? 'Back to node settings' : 'Close join-to-start settings'}">${previous ? 'Back' : '×'}</button></div>
      <div class="wake-delivery-grid" aria-label="Configuration delivery status">
        <div><span>Saved on panel</span><b id="wake-panel-state">Not loaded</b></div>
        <div><span>${id === 'local' ? 'Local configuration' : 'Worker acknowledgement'}</span><b id="wake-worker-state">Not loaded</b></div>
        <div><span>Live join-to-start listeners</span><b id="wake-runtime-state" class="wake-runtime-state unknown" role="status">Not checked</b></div>
      </div>
      <div class="wake-runtime-tools"><button type="button" class="btn sm" id="wake-check">Check listeners</button><span class="mu">Read-only check · never starts a game</span></div>
      <details class="wake-help"><summary>Per-server listener status</summary><p id="wake-runtime-note" class="mu"></p><div id="wake-runtime-list"></div></details>
      <p id="wake-status" class="wake-status" role="status" aria-live="polite"></p><p id="wake-error" class="node-error-note" role="alert" hidden></p>
      <p class="wake-safety"><b>Before enabling:</b> Any unauthenticated Java login can request a start. Pings never wake servers. Disable and confirm worker acceptance for maintenance.</p>
      <form id="wake-form"><fieldset id="wake-fields" class="wake-settings-grid" disabled>
        <section class="wake-section" aria-labelledby="wake-behavior-title"><h3 id="wake-behavior-title">Startup behavior</h3>
          ${controls(['enabled', 'cooldown_seconds', 'bind_address'])}
          <div class="wake-note"><b>Player flow</b><ol><li>See the offline MOTD and icon.</li><li>Join to request startup and receive a disconnect message.</li><li>Reconnect after the real server finishes booting.</li></ol>Early reconnects may fail while the game takes over its normal port. This is not a proxy or a Minecraft 1.21.6 dialog.</div>
        </section>
        <section class="wake-section" aria-labelledby="wake-appearance-title"><h3 id="wake-appearance-title">Offline appearance</h3>
          <div class="wake-template"><button type="button" class="btn sm" id="wake-preset">Use panel branding template</button><small>Replaces MOTD, version label, and both messages using your saved panel name/accent. Other settings and icons stay untouched. Review and Save settings to apply.</small></div>
          ${controls(['offline_motd', 'version_label', 'max_players', 'use_server_icon'])}
          <p class="wake-note"><b>Red X ping icon:</b> the sleeping listener advertises an incompatible protocol so vanilla Minecraft shows its red X instead of normal ping bars. The version label and MOTD explain “Join to start”; a real join still requests startup. This is separate from the PNG server icon.</p>
          <label class="f wake-control" for="wake-upload"><span>Custom fallback icon</span><input id="wake-upload" type="file" accept="image/png" aria-describedby="wake-icon-help"><small id="wake-icon-help">Exactly 64×64 PNG. The complete data URL must be ≤64 KiB (file approximately ≤48 KiB). No remote image URLs.</small></label>
          <div class="wake-icon-actions"><img id="wake-preview" width="64" height="64" alt="Custom offline icon preview" hidden><button type="button" class="btn sm" id="wake-remove">Remove custom icon</button></div>
        </section>
        <section class="wake-section" aria-labelledby="wake-messages-title"><h3 id="wake-messages-title">Player messages</h3>
          ${controls(['legacy_disconnect_colors', 'starting_message', 'unavailable_message'])}
          <div class="wake-note"><b>Text formatting</b><p>Use {server}, {version}, newlines, legacy &amp; codes such as &amp;a, and exact colors such as &amp;#387aff. Colors and styles are sent as Minecraft text components; pre-1.16 clients use the nearest legacy color. Limits count UTF-8 bytes. Very large expanded messages use a safe fallback.</p><p>The online server keeps its own MOTD, icon, and player limit. The server icon is loaded when its offline listener opens.</p></div>
        </section>
      </fieldset>
      <details class="wake-help"><summary>Requirements, troubleshooting, and save behavior</summary><div class="wake-help-grid">
        <section><h4>Which servers can wake?</h4><p>Only a successful node idle auto-shutdown creates wake eligibility. Both node policies and the server’s join-to-start preference must be enabled. Manual stops, crashes, new servers and maintenance do not qualify. Recognized Java-protocol servers need an available primary TCP port (1024–65535); normal startup and isolation safeguards still apply.</p><p>Users can opt out of join-to-start in server Settings → Startup &amp; resources, but cannot disable node-enforced idle shutdown. Updating does not retroactively mark existing stopped servers as idle-stopped.</p><p>The daemon must stay running. Limits: 1024 listeners, 64 connections, and 12 admitted connections per source IP per 10 seconds.</p></section>
        <section><h4>Worker returns HTTP 405 or 404?</h4><ol><li>Update and restart the affected worker, not only this panel.</li><li>Check its saved node URL points to that daemon.</li><li>Ensure its proxy forwards <code>PUT /api/cluster/wake</code> with authentication.</li><li>Choose <b>Retry saved settings</b> and wait for acceptance.</li></ol><p>A 405 means the method was rejected; it does not prove whether the cause is an old binary, wrong URL, or proxy. 401/403 means check the Node Token / admin API key and access rules.</p></section>
        <section><h4>What do the buttons do?</h4><p><b>Save settings</b> saves this form and attempts delivery. <b>Retry saved settings</b> resends the last saved configuration without using unsaved edits. <b>Refresh status</b> reloads acknowledgement and listener state; it does not retry delivery or test ports from the Internet. <b>Check listeners</b> reads the latest runtime reconciliation and per-server bind state.</p><p>Green means ports are bound on the node with matching settings, not that a player can reach them through a firewall or that the game will boot successfully. Status expires after ten seconds; check again for a fresh report. A pending disable is not a confirmed disable.</p></section>
      </div></details>
      <details class="wake-help"><summary>Saved configuration details</summary><pre id="wake-saved"></pre></details>
      <div class="wake-actions"><p id="wake-draft-note" class="mu"></p><div class="row">
      <button type="submit" id="wake-save" class="btn pr" disabled>Save settings</button>
      <button type="button" id="wake-retry" class="btn" disabled>Retry saved settings</button>
      <button type="button" id="wake-refresh" class="btn">Refresh status</button></div></div></form>
      </div></div>`;
    const ctx = { id, owner: identity(), root: document.querySelector('#wake-modal'), saved: null, loaded: false, busy: false, dirty: false, unknown: false, icon: '', error: '', uploadSequence: 0, runtime: null, runtimeBusy: false, runtimeError: '', runtimeSequence: 0, runtimeTimer: null };
    current = ctx;
    const close = () => {
      clearTimeout(ctx.runtimeTimer);
      const restore = previous && live(ctx);
      if (current === ctx) current = null;
      ctx.root.remove();
      if (restore) container.replaceChildren(...previous);
      if (ctx.owner === identity()) opener?.focus();
    };
    el(ctx, 'close').onclick = close;
    el(ctx, 'check').onclick = () => { if (!ctx.busy && !ctx.runtimeBusy) return checkRuntime(ctx); };
    el(ctx, 'preset').onclick = () => applyBrandingPreset(ctx);
    ctx.root.addEventListener('keydown', event => {
      if (event.key === 'Escape') { event.stopPropagation(); close(); }
      if (event.key !== 'Tab') return;
      const focusable = [...ctx.root.querySelectorAll('button:not(:disabled), input:not(:disabled), textarea:not(:disabled), summary')].filter(node => !node.closest('fieldset:disabled'));
      const first = focusable[0], last = focusable[focusable.length - 1];
      if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last?.focus(); }
      else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first?.focus(); }
    });
    el(ctx, 'form').oninput = () => { ctx.dirty = true; paint(ctx); };
    el(ctx, 'form').onsubmit = event => {
      event.preventDefault(); if (!live(ctx) || ctx.busy || !ctx.loaded) return;
      try { const config = draft(ctx); ctx.preserveDraft = false; return request(ctx, config); }
      catch (error) { ctx.error = error.message; paint(ctx); }
    };
    el(ctx, 'retry').onclick = () => { if (!ctx.saved || ctx.busy) return; ctx.preserveDraft = true; return request(ctx, copy(ctx.saved.config)); };
    el(ctx, 'refresh').onclick = () => request(ctx);
    el(ctx, 'upload').onchange = event => upload(ctx, event.target.files?.[0]);
    el(ctx, 'remove').onclick = () => { if (!live(ctx) || ctx.busy) return; ctx.uploadSequence++; ctx.icon = ''; ctx.dirty = true; preview(ctx); paint(ctx); };
    el(ctx, 'close').focus();
    await request(ctx);
  }
  window.NPNodeWake = { open, validate, validEnvelope, statusText, safeIcon, runtimeSummary, validRuntime };
})();
