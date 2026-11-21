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
    cooldown_seconds: ['Start request cooldown (seconds)', 'number', 10, 600]
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
    return `Saved desired setting: ${desired}. ` + (saved.applied
      ? 'Configuration accepted by the node runtime. This does not confirm that every port is listening; check runtime logs for bind failures.'
      : 'Pending worker delivery/application. The worker may still use older settings; a pending disable does NOT mean join to start is disabled. Update/reconnect the worker and retry.');
  }
  function controls() {
    return Object.entries(fields).map(([key, [label, type, min, max]]) => {
      const id = `wake-${key}`;
      if (type === 'checkbox') return `<label style="display:block;margin:12px 0"><input id="${id}" type="checkbox"> ${escape(label)}</label>`;
      return `<label class="f" for="${id}"><span>${escape(label)}</span>${type === 'textarea'
        ? `<textarea id="${id}" rows="3"></textarea>`
        : `<input id="${id}" type="${type}" ${type === 'number' ? `min="${min}" max="${max}" step="1" required` : ''}>`}</label>`;
    }).join('');
  }
  const el = (ctx, id) => ctx.root.querySelector(`#wake-${id}`);
  const live = ctx => current === ctx && ctx.root.isConnected && admin() && ctx.owner === identity();
  function paint(ctx) {
    el(ctx, 'status').textContent = statusText(ctx.saved, ctx.unknown);
    el(ctx, 'error').textContent = ctx.error;
    el(ctx, 'error').hidden = !ctx.error;
    el(ctx, 'fields').disabled = !ctx.loaded || ctx.busy;
    el(ctx, 'save').disabled = !ctx.loaded || ctx.busy;
    el(ctx, 'retry').disabled = !ctx.saved || ctx.busy;
    el(ctx, 'refresh').disabled = ctx.busy;
    el(ctx, 'save').textContent = ctx.busy ? 'Working…' : 'Save settings';
    el(ctx, 'draft-note').textContent = ctx.dirty ? 'Edited fields are not saved. Refresh status and Retry saved settings do not discard your edits.' : 'Fields show the loaded settings. Changes require Save settings.';
    el(ctx, 'saved').textContent = ctx.saved ? JSON.stringify({ ...ctx.saved.config, icon_data_url: ctx.saved.config.icon_data_url ? '(custom PNG configured)' : '' }, null, 2) : 'Not loaded';
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
    ctx.unknown = false; ctx.error = result.error;
    if (initial) fill(ctx, result.config);
  }
  async function request(ctx, config) {
    if (!live(ctx) || ctx.busy) return;
    const initial = !ctx.loaded;
    ctx.busy = true; ctx.error = ''; paint(ctx);
    try {
      const result = await api(`/api/admin/nodes/${encodeURIComponent(ctx.id)}/wake`, { method: config ? 'PUT' : 'GET', ...(config ? { body: JSON.stringify(config) } : {}), suppressToast: true });
      if (!live(ctx)) return;
      accept(ctx, result, initial);
      if (config && !ctx.preserveDraft && !result.error) { fill(ctx, result.config); ctx.dirty = false; }
    } catch (error) {
      if (!live(ctx)) return;
      if (validEnvelope(error.data)) accept(ctx, error.data, initial);
      else ctx.unknown = true;
      ctx.error = error.message || 'Could not load or save settings. Your edits are preserved.';
      if ([404, 405].includes(error.status)) ctx.error += ' Update and restart the panel and this worker to a join-to-start capable version.';
    } finally {
      if (live(ctx)) { ctx.busy = false; paint(ctx); }
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
    container.innerHTML = `<div class="pal" id="wake-modal"><div class="card pbox" role="dialog" aria-modal="true" aria-labelledby="wake-title" style="padding:26px;max-width:720px;width:95vw;max-height:90vh;overflow:auto">
      <div class="row sb"><h2 id="wake-title">Join to start — ${escape(name)}</h2><button type="button" class="btn sm" id="wake-close" aria-label="${previous ? 'Back to node settings' : 'Close join-to-start settings'}">${previous ? 'Back' : '×'}</button></div>
      <p>Opt in this node’s entire local Minecraft Java inventory. Bedrock is not supported. Offline status pings never wake a server; only a login attempt requests a start.</p>
      <p>Players receive a disconnect message and must reconnect after the real server takes its port. Early reconnects may fail during boot. This is not a proxy or a Minecraft 1.21.6 dialog.</p>
      <p><b>Security:</b> An unauthenticated login can request a start. This is not authorization or proof of player identity. Disable for maintenance and confirm worker acceptance; online server MOTD and icons are unaffected.</p>
      <p id="wake-status" role="status" aria-live="polite"></p><p id="wake-error" class="node-error-note" role="alert" hidden></p>
      <details><summary>Saved desired settings (not necessarily active on worker)</summary><pre id="wake-saved" style="white-space:pre-wrap;overflow-wrap:anywhere"></pre></details>
      <form id="wake-form"><fieldset id="wake-fields" disabled style="border:0;padding:0;margin:0">
      <p class="mu">Text supports {server}, {version}, and legacy &amp; color codes (translated to § codes at runtime). Limits count UTF-8 bytes, not characters.</p>
      ${controls()}
      <label class="f" for="wake-upload"><span>Custom offline icon: 64×64 PNG, data URL ≤64 KiB</span><input id="wake-upload" type="file" accept="image/png"></label>
      <img id="wake-preview" width="64" height="64" alt="Custom offline icon preview" hidden>
      <button type="button" class="btn sm" id="wake-remove">Remove custom icon</button>
      <p class="mu">The server-icon option uses a valid server-icon.png when available; the custom icon is the fallback. Neither changes the online server’s icon.</p>
      </fieldset><p id="wake-draft-note" class="mu"></p><div class="row" style="gap:8px;flex-wrap:wrap">
      <button type="submit" id="wake-save" class="btn pr" disabled>Save settings</button>
      <button type="button" id="wake-retry" class="btn" disabled>Retry saved settings</button>
      <button type="button" id="wake-refresh" class="btn">Refresh status</button></div></form>
      </div></div>`;
    const ctx = { id, owner: identity(), root: document.querySelector('#wake-modal'), saved: null, loaded: false, busy: false, dirty: false, unknown: false, icon: '', error: '', uploadSequence: 0 };
    current = ctx;
    const close = () => {
      const restore = previous && live(ctx);
      if (current === ctx) current = null;
      ctx.root.remove();
      if (restore) container.replaceChildren(...previous);
      if (ctx.owner === identity()) opener?.focus();
    };
    el(ctx, 'close').onclick = close;
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
  window.NPNodeWake = { open, validate, validEnvelope, statusText, safeIcon };
})();
