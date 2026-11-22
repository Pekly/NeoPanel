// Server-level wake consent; node auto-shutdown remains administrator-owned.
(() => {
  'use strict';
  let current = null;
  const escape = value => String(value ?? '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'})[c]);
  const owner = () => `${S.user?.id || ''}|${S.token || ''}`;
  const valid = data => data && ['join_to_start_enabled','node_join_to_start_enabled','idle_shutdown_enabled','stopped_by_idle_policy','wake_eligible','can_edit_join_to_start'].every(k => typeof data[k] === 'boolean') && data.idle_shutdown_locked === true && Number.isInteger(data.idle_timeout_minutes) && data.idle_timeout_minutes >= 1 && data.idle_timeout_minutes <= 1440 && typeof data.reason === 'string';
  const reasons = {
    eligible: 'Ready to start when a player joins.',
    not_idle_stopped: 'Available after the server shuts down automatically from inactivity.',
    disabled: 'Player-triggered startup is not available for this server.',
    node_idle_disabled: 'Player-triggered startup is not available for this server.',
    user_disabled: 'Player-triggered startup is off.',
    ineligible: 'Player-triggered startup is not supported for this server.',
    blocked: 'Not available while the server is running or busy.',
    busy: 'Another action is in progress. Try again shortly.',
    conflict: 'Player-triggered startup is temporarily unavailable. Contact support.'
  };
  function render(srv) {
    if (!srv?.id) return '';
    return `<section class="card server-automation-card" id="server-automation-settings" data-server-id="${escape(srv.id)}" data-node-id="${escape(srv.node_id || 'local')}" aria-labelledby="server-automation-title">
      <div class="server-automation-heading"><div><h2 id="server-automation-title">Server automation</h2><p class="mu">Automatic shutdown and player-triggered startup.</p></div><button type="button" class="btn sm" id="server-automation-refresh">Refresh</button></div>
      <div class="server-automation-options">
        <label class="server-automation-option"><span><b>Shut down when empty <span class="chip">Managed by your host</span></b><small id="server-automation-idle-note">Loading settings…</small></span><span class="sw"><input id="server-automation-idle" type="checkbox" disabled aria-disabled="true" aria-describedby="server-automation-idle-note"><i></i></span></label>
        <label class="server-automation-option"><span><b>Start when a player joins</b><small id="server-automation-wake-note">Loading your preference…</small></span><span class="sw"><input id="server-automation-wake" type="checkbox" disabled aria-describedby="server-automation-wake-note"><i></i></span></label>
      </div>
      <p class="mu server-automation-boundary">Works after an automatic idle shutdown. If you stop the server manually, it stays off.</p>
      <p id="server-automation-status" role="status" aria-live="polite">Loading settings…</p><p id="server-automation-error" class="upload-error" role="alert" hidden></p>
    </section>`;
  }
  const el = (ctx, id) => ctx.root.querySelector('#server-automation-' + id);
  const live = ctx => current === ctx && ctx.root.isConnected && S.view === 'set' && S.token && ctx.owner === owner() && curServer()?.id === ctx.id && (curServer()?.node_id || 'local') === ctx.node;
  function paint(ctx) {
    const data = ctx.data;
    el(ctx, 'idle').disabled = true;
    el(ctx, 'idle').checked = Boolean(data?.idle_shutdown_enabled);
    el(ctx, 'idle-note').textContent = !data ? 'Settings are not available yet.' : data.idle_shutdown_enabled ? `Shuts down after ${data.idle_timeout_minutes} minutes without players.` : 'Automatic shutdown is off.';
    el(ctx, 'wake').checked = Boolean(data?.join_to_start_enabled);
    el(ctx, 'wake').disabled = ctx.busy || ctx.unknown || !data?.can_edit_join_to_start;
    el(ctx, 'wake').ariaDisabled = String(el(ctx, 'wake').disabled);
    el(ctx, 'wake').title = data?.can_edit_join_to_start ? 'Save your startup preference' : 'Settings edit and server power permissions required';
    el(ctx, 'wake-note').textContent = !data ? 'Loading your preference…' : !data.can_edit_join_to_start ? 'You need settings and power permissions to change this.' : !data.node_join_to_start_enabled || !data.idle_shutdown_enabled ? 'Your preference is saved, but this feature is not currently available.' : 'Changes save automatically. This does not start the server now.';
    el(ctx, 'refresh').disabled = ctx.busy;
    el(ctx, 'refresh').ariaDisabled = String(ctx.busy);
    el(ctx, 'refresh').title = 'Check the latest settings';
    el(ctx, 'status').textContent = ctx.busy ? 'Updating…' : ctx.unknown ? 'Changes are unconfirmed. Refresh before trying again.' : data ? (reasons[data.reason] || 'Status unavailable. Refresh to try again.') : 'Settings have not loaded.';
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
      if (!valid(result)) throw Error('Could not load automation settings. Refresh to try again.');
      ctx.data = result; ctx.unknown = false;
    } catch (error) {
      if (live(ctx)) {ctx.unknown = true; ctx.error = S.user?.role === 'admin' ? (error.message || 'Could not confirm the change. Refresh to try again.') : 'Could not confirm your settings. Refresh to try again, or contact support if the problem continues.';}
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
