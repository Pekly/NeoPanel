// Server-level wake consent; node auto-shutdown remains administrator-owned.
(() => {
  'use strict';
  let current = null;
  const escape = value => String(value ?? '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'})[c]);
  const owner = () => `${S.user?.id || ''}|${S.token || ''}`;
  const valid = data => data && ['join_to_start_enabled','node_join_to_start_enabled','idle_shutdown_enabled','stopped_by_idle_policy','wake_eligible','can_edit_join_to_start'].every(k => typeof data[k] === 'boolean') && data.idle_shutdown_locked === true && Number.isInteger(data.idle_timeout_minutes) && data.idle_timeout_minutes >= 1 && data.idle_timeout_minutes <= 1440 && typeof data.reason === 'string';
  const reasons = {
    eligible: 'This server was stopped by the idle policy and is eligible to wake when a player joins.',
    not_idle_stopped: 'Waiting for a successful idle auto-shutdown. Manual stops, crashes and newly created servers cannot wake.',
    disabled: 'Join-to-start is disabled by the node administrator.',
    node_idle_disabled: 'The node must enable idle auto-shutdown before join-to-start can apply.',
    user_disabled: 'Join-to-start is off for this server. The node’s idle auto-shutdown policy still applies.',
    ineligible: 'This server type, host assignment or game port is not eligible for join-to-start.',
    blocked: 'The server is active, starting, halted, or blocked by a lifecycle safeguard; automatic wake is unavailable.',
    busy: 'Another server operation is in progress. Refresh after it finishes.',
    conflict: 'Another local server uses this game port. An administrator must resolve the conflict.'
  };
  function render(srv) {
    if (!srv?.id) return '';
    return `<section class="card server-automation-card" id="server-automation-settings" data-server-id="${escape(srv.id)}" data-node-id="${escape(srv.node_id || 'local')}" aria-labelledby="server-automation-title">
      <div class="server-automation-heading"><div><h2 id="server-automation-title">Server automation</h2><p class="mu">Node policy controls idle shutdown. You control whether this server can wake after it.</p></div><button type="button" class="btn sm" id="server-automation-refresh">Refresh policy</button></div>
      <div class="server-automation-options">
        <label class="server-automation-option"><span><b>Idle auto-shutdown <span class="chip">Locked by node policy</span></b><small id="server-automation-idle-note">Loading administrator policy…</small></span><span class="sw"><input id="server-automation-idle" type="checkbox" disabled aria-disabled="true" aria-describedby="server-automation-idle-note"><i></i></span></label>
        <label class="server-automation-option"><span><b>Join to start after idle shutdown</b><small id="server-automation-wake-note">Loading your server preference…</small></span><span class="sw"><input id="server-automation-wake" type="checkbox" disabled aria-describedby="server-automation-wake-note"><i></i></span></label>
      </div>
      <p class="mu server-automation-boundary">Only a successful node idle auto-shutdown can make this server eligible. Manual stops, crashes, fresh servers, maintenance and failed starts do not qualify. This switch cannot disable the node’s auto-shutdown policy.</p>
      <p id="server-automation-status" role="status" aria-live="polite">Loading policy from the hosting node…</p><p id="server-automation-error" class="upload-error" role="alert" hidden></p>
    </section>`;
  }
  const el = (ctx, id) => ctx.root.querySelector('#server-automation-' + id);
  const live = ctx => current === ctx && ctx.root.isConnected && S.view === 'set' && S.token && ctx.owner === owner() && curServer()?.id === ctx.id && (curServer()?.node_id || 'local') === ctx.node;
  function paint(ctx) {
    const data = ctx.data;
    el(ctx, 'idle').disabled = true;
    el(ctx, 'idle').checked = Boolean(data?.idle_shutdown_enabled);
    el(ctx, 'idle-note').textContent = !data ? 'Policy unavailable until the hosting node responds.' : data.idle_shutdown_enabled ? `Required by this node: shutdown after ${data.idle_timeout_minutes} minutes with no tracked players. Users cannot turn it off.` : 'Disabled by the node administrator. Only administrators can change the node policy.';
    el(ctx, 'wake').checked = Boolean(data?.join_to_start_enabled);
    el(ctx, 'wake').disabled = ctx.busy || ctx.unknown || !data?.can_edit_join_to_start;
    el(ctx, 'wake').ariaDisabled = String(el(ctx, 'wake').disabled);
    el(ctx, 'wake').title = data?.can_edit_join_to_start ? 'Join-to-start preference only; node auto-shutdown stays enforced' : 'Settings edit and server power permissions required';
    el(ctx, 'wake-note').textContent = !data ? 'Waiting for permissions and policy.' : !data.can_edit_join_to_start ? 'Changing this preference requires server settings and power permissions.' : !data.node_join_to_start_enabled || !data.idle_shutdown_enabled ? 'Your preference is saved independently, but this node must enable both idle shutdown and join-to-start before it can take effect.' : 'Automatically save this preference. Enabling it never starts the server immediately; a player must join after an eligible idle shutdown.';
    el(ctx, 'refresh').disabled = ctx.busy;
    el(ctx, 'refresh').ariaDisabled = String(ctx.busy);
    el(ctx, 'refresh').title = 'Read the hosting node policy without changing it';
    el(ctx, 'status').textContent = ctx.busy ? 'Checking / saving the hosting node policy…' : ctx.unknown ? 'The latest state is unconfirmed. Refresh policy before making another change.' : data ? (reasons[data.reason] || data.reason) : 'Policy has not been loaded.';
    el(ctx, 'error').textContent = ctx.error;
    el(ctx, 'error').hidden = !ctx.error;
  }
  async function request(ctx, enabled) {
    if (!live(ctx) || ctx.busy) return;
    if (enabled !== undefined && (!ctx.data?.can_edit_join_to_start || ctx.unknown)) return;
    ctx.busy = true; ctx.error = ''; paint(ctx);
    try {
      const result = await api(ctx.path, {method: enabled === undefined ? 'GET' : 'PATCH', ...(enabled === undefined ? {} : {body:JSON.stringify({join_to_start_enabled:enabled})}), suppressToast:true});
      if (!live(ctx)) return;
      if (!valid(result)) throw Error('Unexpected server automation response. Update the panel and hosting node, then refresh.');
      ctx.data = result; ctx.unknown = false;
    } catch (error) {
      if (live(ctx)) {ctx.unknown = true; ctx.error = error.message || 'Could not confirm the preference. Refresh before retrying.';}
    } finally {if (live(ctx)) {ctx.busy = false; paint(ctx);}}
  }
  function mount() {
    const root = document.querySelector('#server-automation-settings');
    const srv = curServer();
    if (!root || !srv?.id || !S.token || S.view !== 'set') {current = null; return;}
    if (current?.root === root) return;
    const ctx = {root, id:srv.id, node:srv.node_id || 'local', owner:owner(), path:srvApiPath({...srv},'/automation'), data:null, busy:false, unknown:false, error:''};
    current = ctx;
    el(ctx, 'wake').onchange = event => request(ctx, Boolean(event.target.checked));
    el(ctx, 'refresh').onclick = () => request(ctx);
    return request(ctx);
  }
  window.NPServerAutomation = {render, mount, valid};
})();
