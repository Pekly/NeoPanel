// Proxy config drafts are private to one server/node/session, never S.props or storage.
(() => {
  'use strict';
  const supports = type => ['velocity', 'bungeecord', 'waterfall'].includes(String(type || '').trim().toLowerCase());
  const clone = value => JSON.parse(JSON.stringify(value));
  const esc = value => String(value ?? '').replace(/[&<>"']/g, c => ({'&':'&amp;', '<':'&lt;', '>':'&gt;', '"':'&quot;', "'":'&#39;'})[c]);
  const lower = value => String(value).toLowerCase();
  const bool = value => typeof value === 'boolean';
  const strings = value => Array.isArray(value) && value.length <= 128 && value.every(v => typeof v === 'string');
  const object = value => value && typeof value === 'object' && !Array.isArray(value);
  let current = null;
  function valid(view, type) {
    if (!object(view) || !supports(view.type) || view.type !== type || view.file !== (type === 'velocity' ? 'velocity.toml' : 'config.yml') ||
        !['exists', 'can_edit', 'can_view_secret', 'running', 'ready'].every(k => bool(view[k])) ||
        typeof view.revision !== 'string' || (view.exists && !view.revision) || !Number.isInteger(view.port) || view.port < 1 || view.port > 65535) return false;
    const c = view.config;
    // Validate the transport shape here, not editable field values. The API can
    // return a recoverable draft with validation_error and ready:false.
    return object(c) && typeof c.motd === 'string' && Number.isInteger(c.max_players) && c.max_players >= -2147483648 && c.max_players <= 2147483647 &&
      ['online_mode', 'ip_forward', 'force_key_authentication', 'force_default_server'].every(k => bool(c[k])) &&
      typeof c.forwarding_mode === 'string' && typeof c.ping_passthrough === 'string' &&
      (view.validation_error === undefined || typeof view.validation_error === 'string') &&
      Array.isArray(c.backends) && c.backends.length <= 128 && c.backends.every(b => object(b) && ['name', 'address', 'motd'].every(k => typeof b[k] === 'string') && bool(b.restricted)) &&
      strings(c.try) && Array.isArray(c.forced_hosts) && c.forced_hosts.length <= 128 && c.forced_hosts.every(r => object(r) && typeof r.host === 'string' && strings(r.servers));
  }
  function addressOK(value) {
    const match = /^(\[[0-9a-fA-F:.]+\]|[a-zA-Z0-9_](?:[a-zA-Z0-9_.-]*[a-zA-Z0-9_])?):([0-9]{1,5})$/.exec(value);
    if (!match || Number(match[2]) < 1 || Number(match[2]) > 65535) return false;
    if (match[1].startsWith('[')) { try { new URL('http://' + value); } catch (_) { return false; } }
    return true;
  }
  function validate(c, type) {
    if (!Number.isInteger(c.max_players) || c.max_players < 0 || c.max_players > 2147483647) return 'Displayed slots must be 0–2147483647.';
    if (type === 'velocity' && !['NONE', 'LEGACY', 'MODERN', 'BUNGEEGUARD'].includes(c.forwarding_mode)) return 'Choose a valid player forwarding mode in Authentication & forwarding.';
    if (!(type === 'velocity' ? ['DISABLED', 'MODS', 'DESCRIPTION', 'ALL'] : ['DISABLED', 'ALL']).includes(c.ping_passthrough)) return 'Choose a valid server-list ping passthrough mode in Authentication & forwarding.';
    if (!c.backends.length) return 'Add at least one backend server.';
    if (c.backends.length > 128 || c.try.length > 128 || c.forced_hosts.length > 128) return 'Use at most 128 servers, routes and fallback entries.';
    const names = new Set();
    for (const b of c.backends) {
      if (!/^[a-zA-Z0-9_-]{1,64}$/.test(b.name) || lower(b.name) === 'try') return 'Server names: 1–64 letters, numbers, _ or -. “try” is reserved.';
      if (names.has(lower(b.name))) return 'Give every backend a unique name (ignoring case).';
      names.add(lower(b.name));
      if (!addressOK(b.address)) return `Enter a hostname:port or [IPv6]:port address for ${b.name}. No URLs.`;
    }
    const refsOK = refs => refs.length > 0 && refs.length <= 128 && new Set(refs.map(lower)).size === refs.length && refs.every(n => names.has(lower(n)));
    if (!refsOK(c.try)) return 'Choose at least one valid fallback server, without duplicates.';
    const hosts = new Set();
    for (const r of c.forced_hosts) {
      if (!/^(?=.{1,253}$)[a-zA-Z0-9](?:[a-zA-Z0-9.-]*[a-zA-Z0-9])?$/.test(r.host) || r.host.includes('..') || hosts.has(lower(r.host))) return 'Use a unique domain for each route, without a port or URL.';
      hosts.add(lower(r.host));
      if (!refsOK(r.servers) || (type !== 'velocity' && r.servers.length !== 1)) return type === 'velocity' ? 'Each domain needs valid server names in order, without duplicates (max 128).' : 'Choose exactly one backend for each domain.';
    }
    return '';
  }
  // Bind references to row objects while typing: a temporary duplicate name must
  // not steal another server's routes when the user finishes the rename.
  const bindings = new WeakMap();
  const refText = c => JSON.stringify([c.try, c.forced_hosts.map(r => r.servers)]);
  function refs(c) {
    let value = bindings.get(c);
    if (!value || value.text !== refText(c)) {
      const bind = names => names.map(n => c.backends.find(b => lower(b.name) === lower(n)) || n);
      value = {join:bind(c.try), hosts:c.forced_hosts.map(r => bind(r.servers)), text:refText(c)};
      bindings.set(c, value);
    }
    return value;
  }
  function rename(c, index, name) {
    const value = refs(c);
    c.backends[index].name = name;
    const names = list => list.map(b => typeof b === 'string' ? b : b.name);
    c.try = names(value.join);
    c.forced_hosts.forEach((r, i) => { r.servers = names(value.hosts[i]); });
    value.text = refText(c);
  }
  function remove(c, index) {
    const value = refs(c), target = c.backends[index];
    const without = list => list.filter(b => b !== target).map(b => typeof b === 'string' ? b : b.name);
    c.backends.splice(index, 1);
    c.try = without(value.join);
    c.forced_hosts.forEach((r, i) => { r.servers = without(value.hosts[i]); });
    bindings.delete(c);
  }
  function move(list, index, delta) {
    const to = index + delta;
    if (index >= 0 && to >= 0 && index < list.length && to < list.length) [list[index], list[to]] = [list[to], list[index]];
  }
  const identity = () => JSON.stringify([S.user?.id, S.user?.role, S.token, curServer()?.id, curServer()?.node_id || 'local', lower(curServer()?.type || '')]);
  const permitted = (ctx, permission) => typeof hasServerPermission === 'function' && hasServerPermission(permission, ctx.snapshot) === true;
  const same = ctx => current === ctx && Boolean(S.token) && ctx.identity === identity() && S.view === 'set' && permitted(ctx, 'settings.view');
  const live = ctx => same(ctx) && ctx.root?.isConnected;
  const editable = ctx => live(ctx) && !ctx.busy && !ctx.blocked && ctx.data?.can_edit === true && ctx.data.running === false && permitted(ctx, 'settings.edit');
  const secretAllowed = ctx => live(ctx) && !ctx.busy && !ctx.blocked && ctx.data?.exists && ctx.type === 'velocity' && ctx.data.can_edit === true && ctx.data.can_view_secret === true && permitted(ctx, 'settings.edit') && permitted(ctx, 'files.view');
  function clearSecret(ctx) {
    ctx.secret = '';
    const box = ctx.root?.querySelector('[data-secret]');
    if (box) { box.value = ''; box.removeAttribute('value'); }
  }
  function reset() { if (current) clearSecret(current); current = null; }
  function sync() {
    if (current && !same(current)) reset();
    else if (current && (!permitted(current, 'settings.edit') || !permitted(current, 'files.view'))) clearSecret(current);
  }
  function render(srv) {
    if (!supports(srv?.type)) return '';
    return '<section class="card proxy-settings" id="proxy-settings" tabindex="-1" aria-labelledby="proxy-settings-title"><h2 id="proxy-settings-title">Proxy configuration</h2><p role="status">Loading proxy configuration…</p></section>';
  }
  const button = (action, text, disabled = false, index = '') => `<button type="button" class="btn sm" data-action="${action}" data-index="${index}" ${disabled ? 'disabled' : ''}>${text}</button>`;
  const input = (label, value, attrs) => `<label class="f"><span>${label}</span><input ${attrs} value="${esc(value)}"></label>`;
  const check = (label, value, attrs) => `<label class="proxy-check"><input type="checkbox" ${attrs} ${value ? 'checked' : ''}><span>${label}</span></label>`;
  const select = (label, value, choices, attrs) => `<label class="f"><span>${label}</span><select class="no-cselect" ${attrs}>${[...new Set([...choices, value])].map(v => `<option value="${esc(v)}" ${v === value ? 'selected' : ''}>${esc(v || 'Choose a server')}</option>`).join('')}</select></label>`;
  function paint(ctx, focus) {
    if (!live(ctx)) return;
    const d = ctx.data, c = ctx.draft, vel = ctx.type === 'velocity', disabled = !editable(ctx);
    const open = new Set([...ctx.root.querySelectorAll('details[open]')].map(el => el.dataset.section));
    const active = document.activeElement;
    if (!focus && active && ctx.root.contains?.(active) && active.dataset.action) focus = `[data-action="${active.dataset.action}"][data-index="${active.dataset.index || ''}"]`;
    ctx.root.innerHTML = `<header class="proxy-heading"><div><h2 id="proxy-settings-title">Proxy configuration</h2><p class="mu">${vel ? 'Velocity · velocity.toml' : (ctx.type === 'waterfall' ? 'Waterfall' : 'BungeeCord') + ' · config.yml'}</p></div>${button('refresh', ctx.blocked ? 'Retry / Refresh' : 'Refresh', ctx.busy)}</header>
      <p class="proxy-status" role="status" aria-live="polite">${esc(ctx.busy ? 'Working…' : ctx.message || (ctx.dirty ? 'Unsaved proxy changes' : d ? (d.running ? 'Stop the proxy before saving.' : !d.can_edit || !permitted(ctx, 'settings.edit') ? 'View only' : !d.ready ? 'Setup needed · Add a server and join order.' : 'Ready to edit') : 'Configuration unavailable.'))}</p>
      <p class="proxy-error" role="alert" ${ctx.error ? '' : 'hidden'}>${esc(ctx.error)}</p>
      ${!d ? (ctx.showFiles ? button('files', 'Open ' + (vel ? 'velocity.toml' : 'config.yml') + ' in Files', !permitted(ctx, 'files.view')) : '') : !d.exists ? `<div class="proxy-empty"><h3>Create your proxy config</h3><p>No config file yet. Create it, then add your backend servers.</p>${button('initialize', 'Create config', disabled)}<p class="mu">Only creates a missing file. Does not start the proxy.</p></div>` : `
      <fieldset ${disabled ? 'disabled' : ''} class="proxy-fields"><legend class="sr-only">Proxy settings</legend>
        <div class="proxy-grid">${input('Server message (MOTD)', c.motd, 'data-field="motd"')}${input('Displayed slots', c.max_players, 'type="number" min="0" max="2147483647" data-field="max_players"')}
        </div><p class="mu proxy-hint">${vel ? 'MOTD supports MiniMessage, such as &lt;bold&gt;Welcome&lt;/bold&gt;.' : 'MOTD supports &amp; color codes, such as &amp;aWelcome.'}</p>
        ${input('Assigned proxy port', d.port, 'readonly aria-describedby="proxy-port-note"')}<p id="proxy-port-note" class="mu proxy-hint">Change the port in Networking, if allowed.</p>
        <div class="proxy-heading"><h3>Backend servers</h3>${button('add-backend', 'Add server', disabled || c.backends.length >= 128)}</div>
        <p class="mu">Name each server and enter its address. No backend settings are changed automatically.</p>
        <div class="proxy-list">${c.backends.map((b, i) => `<section class="proxy-row" aria-label="Backend ${i + 1}"><div class="proxy-grid">
          ${input('Server name', b.name, `data-list="backends" data-index="${i}" data-key="name" maxlength="64" aria-label="Server ${i + 1} name"`)}
          ${input('Address · hostname:port', b.address, `data-list="backends" data-index="${i}" data-key="address" placeholder="e.g. lobby.internal:25566" aria-label="Server ${i + 1} address"`)}
          </div>${!vel ? `<details data-section="backend-${i}"><summary>Server options</summary>${input('Backend MOTD', b.motd, `data-list="backends" data-index="${i}" data-key="motd"`)}${check('Restricted (requires server permission)', b.restricted, `data-list="backends" data-index="${i}" data-key="restricted"`)}</details>` : ''}
          ${button('remove-backend', 'Remove server ' + (i + 1), disabled, i)}</section>`).join('') || '<p class="proxy-empty">No backend servers. Choose Add server to begin.</p>'}</div>
        <h3>Join &amp; fallback order</h3><p class="mu">Players join the first available server in this order.</p>
        <div class="proxy-list">${c.try.map((n, i) => `<div class="proxy-order">${select('Priority ' + (i + 1), n, [...new Set(['', ...c.backends.map(b => b.name), n])], `data-list="try" data-index="${i}"`)}<div class="proxy-actions">${button('up', 'Move up', disabled || i === 0, i)}${button('down', 'Move down', disabled || i === c.try.length - 1, i)}${button('remove-fallback', 'Remove priority ' + (i + 1), disabled, i)}</div></div>`).join('') || '<p class="proxy-empty">No join order selected.</p>'}</div>
        ${button('add-fallback', 'Add to join order', disabled || c.try.length >= 128 || !c.backends.some(b => b.name && !c.try.some(n => lower(n) === lower(b.name))))}
        <details data-section="routes"><summary>Domain routing <span class="mu">Optional · ${c.forced_hosts.length}</span></summary><p class="mu">Send players joining a domain to ${vel ? 'these servers, in order' : 'one server'}.</p>
          ${c.forced_hosts.map((r, i) => `<div class="proxy-row"><div class="proxy-grid">${input('Domain', r.host, `data-list="forced_hosts" data-index="${i}" data-key="host" placeholder="play.example.com"`)}${vel ? input('Server names in order (comma separated)', r.servers.join(', '), `data-list="forced_hosts" data-index="${i}" data-key="servers"`) : select('Target server', r.servers[0] || '', [...new Set(['', ...c.backends.map(b => b.name), ...r.servers])], `data-list="forced_hosts" data-index="${i}" data-key="servers"`)}</div>${button('remove-route', 'Remove domain ' + (i + 1), disabled, i)}</div>`).join('')}
          ${button('add-route', 'Add domain', disabled || c.forced_hosts.length >= 128)}</details>
        <details data-section="auth"><summary>Authentication &amp; forwarding</summary>
          ${check('Verify player accounts (online mode)', c.online_mode, 'data-field="online_mode"')}
          ${vel ? select('Player forwarding', c.forwarding_mode, ['NONE', 'LEGACY', 'MODERN', 'BUNGEEGUARD'], 'data-field="forwarding_mode"') + check('Require signed player keys', c.force_key_authentication, 'data-field="force_key_authentication"') : check('Forward player IPs and UUIDs', c.ip_forward, 'data-field="ip_forward"') + check('Always join the default server', c.force_default_server, 'data-field="force_default_server"')}
          <p class="mu">Match forwarding in each backend’s proxy settings. Protect backends from direct access.</p>
          ${select('Server-list ping passthrough', c.ping_passthrough, vel ? ['DISABLED', 'MODS', 'DESCRIPTION', 'ALL'] : ['DISABLED', 'ALL'], 'data-field="ping_passthrough"')}
        </details>
      </fieldset>
      ${vel && d.can_view_secret && d.can_edit && permitted(ctx, 'files.view') && permitted(ctx, 'settings.edit') ? `<section class="proxy-secret"><h3>Forwarding secret</h3><p class="mu">Set the matching secret in each backend’s proxy settings. Keep it private.</p>${ctx.secret ? `${input('Forwarding secret', ctx.secret, 'data-secret readonly autocomplete="off" spellcheck="false"')}<div class="proxy-actions">${button('copy-secret', 'Copy secret', ctx.busy)}${button('close-secret', 'Hide secret')}</div>` : button('secret', 'Show forwarding secret', !secretAllowed(ctx))}</section>` : ''}
      <footer class="proxy-footer"><p class="mu">Save while stopped. Changes apply on the next manual start.</p>${button('save', 'Save proxy config', disabled || !ctx.dirty)}</footer>`}`;
    ctx.root.querySelectorAll('details').forEach(el => { el.open = open.has(el.dataset.section); });
    if (focus) ctx.root.querySelector(focus)?.focus();
  }
  function markDirty(ctx) {
    ctx.dirty = true; ctx.error = ''; ctx.message = '';
    const status = ctx.root.querySelector('[role="status"]');
    if (status) status.textContent = 'Unsaved proxy changes';
    const save = ctx.root.querySelector('[data-action="save"]');
    if (save) save.disabled = !editable(ctx);
    const error = ctx.root.querySelector('[role="alert"]');
    if (error) error.hidden = true;
  }
  function change(ctx, event) {
    if (!editable(ctx) || !ctx.data.exists) return;
    const el = event.target, {field, list, index, key} = el.dataset;
    if (!field && !list) return;
    const value = el.type === 'checkbox' ? el.checked : el.value;
    if (field && ['motd', 'max_players', 'online_mode', 'forwarding_mode', 'ip_forward', 'force_key_authentication', 'force_default_server', 'ping_passthrough'].includes(field)) ctx.draft[field] = field === 'max_players' ? Number(value) : value;
    else if (list === 'backends' && ctx.draft.backends[index]) {
      if (key === 'name') rename(ctx.draft, Number(index), value);
      else if (['address', 'motd', 'restricted'].includes(key)) ctx.draft.backends[index][key] = value;
    } else if (list === 'try') ctx.draft.try[index] = value;
    else if (list === 'forced_hosts' && ctx.draft.forced_hosts[index]) ctx.draft.forced_hosts[index][key] = key === 'servers' ? value.split(',').map(n => n.trim()).filter(Boolean) : value;
    markDirty(ctx);
    if (list === 'backends' && key === 'name') {
      // Update references without replacing the input during blur/Tab navigation.
      ctx.root.querySelectorAll('[data-list="try"], select[data-key="servers"]').forEach(selectEl => {
        const selected = selectEl.dataset.list === 'try' ? ctx.draft.try[selectEl.dataset.index] : ctx.draft.forced_hosts[selectEl.dataset.index].servers[0] || '';
        selectEl.innerHTML = [...new Set(['', ...ctx.draft.backends.map(b => b.name), selected])].map(n => `<option value="${esc(n)}" ${n === selected ? 'selected' : ''}>${esc(n || 'Choose a server')}</option>`).join('');
      });
      ctx.root.querySelectorAll('input[data-list="forced_hosts"][data-key="servers"]').forEach(el => { el.value = ctx.draft.forced_hosts[el.dataset.index].servers.join(', '); });
      const add = ctx.root.querySelector('[data-action="add-fallback"]');
      if (add) add.disabled = ctx.draft.try.length >= 128 || !ctx.draft.backends.some(b => b.name && !ctx.draft.try.some(n => lower(n) === lower(b.name)));
    }
  }
  async function request(ctx, action) {
    if (!live(ctx) || ctx.busy) return;
    if (['save', 'initialize'].includes(action) && !editable(ctx)) return;
    if (action === 'save' && (!ctx.data.exists || !ctx.dirty)) return;
    if (action === 'initialize' && ctx.data.exists) return;
    if (action === 'secret' && !secretAllowed(ctx)) return;
    clearSecret(ctx);
    const issue = action === 'save' && validate(ctx.draft, ctx.type);
    if (issue) { ctx.error = issue; paint(ctx); return; }
    ctx.busy = true; ctx.error = ''; paint(ctx);
    let confirmOffline = false;
    try {
      if (action === 'refresh' && ctx.dirty) {
        if (!await uiConfirm({title:'Discard proxy changes?', message:'Refresh replaces your unsaved proxy draft with the latest config.', confirmText:'Discard & refresh', isDanger:true})) return;
      }
      if (action === 'save' && ctx.draft.online_mode === false) {
        confirmOffline = await uiConfirm({title:'Disable account verification?', message:'Offline mode lets players impersonate other accounts. Only continue if you have a separate, trusted authentication setup.', confirmText:'Save with offline mode', isDanger:true});
        if (!confirmOffline) return;
      }
      // Dialogs can outlive a route, session, permission or server change.
      if (!live(ctx) || (action !== 'refresh' && (!permitted(ctx, 'settings.edit') || (action === 'secret' && !permitted(ctx, 'files.view'))))) return;
      const suffix = action === 'initialize' ? '/initialize' : action === 'secret' ? '/secret' : '';
      const options = {method:action === 'save' ? 'PUT' : ['initialize', 'secret'].includes(action) ? 'POST' : 'GET', suppressToast:true};
      if (options.method !== 'GET') options.body = JSON.stringify(action === 'save' ? {revision:ctx.data.revision, config:clone(ctx.draft), confirm_offline:confirmOffline} : {});
      const response = await api(ctx.path + suffix, options);
      if (!live(ctx)) return;
      if (action === 'secret') {
        if (!permitted(ctx, 'settings.edit') || !permitted(ctx, 'files.view')) return;
        if (typeof response?.secret !== 'string' || !response.secret) throw Error('invalid');
        ctx.secret = response.secret; ctx.message = 'Secret revealed. Hide it when finished.';
      } else {
        if (!valid(response, ctx.type)) throw Object.assign(Error('invalid'), {incompatible:true});
        if ((action === 'initialize' && !response.exists) || (action === 'save' && (!response.exists || !response.ready || response.validation_error || validate(response.config, ctx.type)))) throw Object.assign(Error('unconfirmed'), {incompatible:true});
        ctx.data = clone(response); ctx.draft = clone(response.config); ctx.dirty = false; ctx.blocked = false; ctx.showFiles = false;
        if (response.exists) ctx.error = response.validation_error || validate(ctx.draft, ctx.type);
        ctx.message = action === 'save' ? 'Proxy config saved. Start the proxy manually when ready.' : action === 'initialize' ? 'Config created. Add your first backend server.' : '';
      }
    } catch (error) {
      if (!live(ctx)) return;
      ctx.blocked = true;
      ctx.message = ctx.dirty ? 'Draft kept · Refresh required before saving.' : 'Refresh required.';
      const failure = failureAdvice(error, ctx.type);
      ctx.error = failure.message; ctx.showFiles = failure.files;
    } finally {
      if (live(ctx)) { ctx.busy = false; paint(ctx); }
      else clearSecret(ctx);
    }
  }
  function failureAdvice(error, type) {
    const file = type === 'velocity' ? 'velocity.toml' : 'config.yml';
    const code = error.data?.code;
    const advice = {
      unsupported_config:`This document has invalid syntax or a structure the builder cannot safely edit. Open ${file} in Files, or ask your administrator to check it. No file was reset.`,
      file_unavailable:'The config file cannot be accessed safely. Ask your administrator to check permissions, file size and server path links, then Retry.',
      unsafe_secret:'The forwarding secret path or file is unsafe or invalid. Ask your administrator to check it without replacing an existing secret, then Retry.',
      config_missing:'The config is missing. Retry to load its current state; Create config is offered only after the server confirms it is missing.',
      revision_conflict:'The config changed. Refresh before saving; your draft has not been overwritten.',
      server_running:'Stop the proxy before saving, then Refresh. Your draft has not been overwritten.',
      operation_busy:'Another server action is in progress. Wait for it to finish, then Retry.',
      validation_failed:'The settings were rejected. Check the field values, then Refresh before retrying. Your draft has been kept.'
    };
    if (typeof advice[code] === 'string') return {message:advice[code], files:code === 'unsupported_config'};
    if (error.incompatible || [404, 405, 501].includes(error.status)) return {message:'Could not confirm a compatible proxy settings API. Retry; if this continues, ask your administrator to update the panel and hosting worker together and check the server route.', files:false};
    if (error.status === 401) return {message:'Your session expired. Sign in again, then Retry.', files:false};
    if (error.status === 403) return {message:'Proxy settings access was denied. Ask the server owner to check your Settings permissions, then Retry.', files:false};
    if (error.status === 409) return {message:'Config changed or the proxy is busy. Check its status, then Refresh. Your draft has not been overwritten.', files:false};
    if (error.status === 400) return {message:'The hosting worker rejected this request. Retry; if it continues, ask your administrator to check worker compatibility and the config format. No file was reset.', files:false};
    return {message:'Could not reach or confirm the hosting worker. Check your connection, then Retry. If this continues, contact your administrator. No save has been confirmed.', files:false};
  }
  async function action(ctx, event) {
    const target = event.target.closest('[data-action]');
    if (!target || !live(ctx) || target.disabled) return;
    const name = target.dataset.action, i = Number(target.dataset.index), c = ctx.draft;
    if (name === 'close-secret') { clearSecret(ctx); ctx.message = ''; paint(ctx, '[data-action="secret"]'); return; }
    if (ctx.busy) return;
    if (name === 'files') { if (permitted(ctx, 'files.view')) go('fi'); return; }
    if (name === 'copy-secret') {
      if (!secretAllowed(ctx) || !ctx.secret) return;
      try { await navigator.clipboard.writeText(ctx.secret); if (live(ctx)) ctx.message = 'Secret copied. Keep it private.'; }
      catch (_) { if (live(ctx)) ctx.error = 'Could not copy. Select the secret and copy it manually.'; }
      if (live(ctx)) paint(ctx); return;
    }
    if (['refresh', 'save', 'initialize', 'secret'].includes(name)) return request(ctx, name);
    if (!editable(ctx) || !ctx.data.exists) return;
    let focus;
    if (name === 'add-backend' && c.backends.length < 128) { c.backends.push({name:'', address:'', motd:'', restricted:false}); focus = `[data-list="backends"][data-index="${c.backends.length - 1}"][data-key="name"]`; }
    else if (name === 'remove-backend' && c.backends[i]) remove(c, i);
    else if (name === 'add-fallback' && c.try.length < 128) { const b = c.backends.find(b => b.name && !c.try.some(n => lower(n) === lower(b.name))); if (!b) return; c.try.push(b.name); }
    else if (name === 'remove-fallback') c.try.splice(i, 1);
    else if (name === 'up' || name === 'down') { const delta = name === 'up' ? -1 : 1; move(c.try, i, delta); focus = `[data-list="try"][data-index="${i + delta}"]`; }
    else if (name === 'add-route' && c.forced_hosts.length < 128) { c.forced_hosts.push({host:'', servers:[]}); focus = `[data-list="forced_hosts"][data-index="${c.forced_hosts.length - 1}"][data-key="host"]`; }
    else if (name === 'remove-route') c.forced_hosts.splice(i, 1);
    else return;
    markDirty(ctx); paint(ctx, focus);
    if (focus && name.startsWith('add-')) ctx.root.querySelector(focus)?.closest('.proxy-row, .proxy-order')?.classList.add('proxy-new');
  }
  function mount() {
    sync();
    const root = document.querySelector('#proxy-settings'), srv = curServer();
    if (!root || !supports(srv?.type) || !S.token || S.view !== 'set') { if (current) clearSecret(current); current = null; return; }
    if (current?.root === root) return;
    if (current) { clearSecret(current); current.root = root; }
    else {
      const snapshot = {...srv};
      current = {root, snapshot, identity:identity(), type:lower(srv.type), path:srvApiPath(snapshot, '/proxy-config'), data:null, draft:null, dirty:false, busy:false, blocked:false, error:'', message:'', secret:''};
    }
    const ctx = current;
    if (!permitted(ctx, 'settings.view')) { clearSecret(ctx); current = null; root.textContent = 'Settings view permission required.'; return; }
    root.oninput = event => change(ctx, event);
    root.onchange = event => change(ctx, event);
    root.onclick = event => action(ctx, event);
    paint(ctx);
    if (!ctx.data && !ctx.busy && !ctx.blocked) return request(ctx, 'refresh');
  }
  window.NPProxySettings = {supports, render, mount, sync, reset, valid, validate, addressOK, rename, remove, move};
})();
