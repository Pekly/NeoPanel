// Idle policy delivery is independent from connection health or join-to-start.
(() => {
  'use strict';
  let current = null;
  const admin = () => typeof S !== 'undefined' && S.user?.role === 'admin' && Boolean(S.token);
  const identity = () => `${S.user?.id || ''}|${S.token || ''}`;
  const valid = value => value && typeof value.applied === 'boolean' && typeof value.error === 'string' && value.config && typeof value.config.enabled === 'boolean' && Number.isInteger(value.config.timeout_minutes) && value.config.timeout_minutes >= 1 && value.config.timeout_minutes <= 1440;
  // The original form may be temporarily detached by the separate wake dialog.
  // Keep its response usable on Back, but never apply it across nodes/sessions.
  const active = ctx => current === ctx && admin() && ctx.owner === identity();
  const el = (ctx, id) => ctx.root.querySelector('#' + id);
  function summary(view, local) {
    if (!view) return 'Idle policy delivery has not been checked. Save the node to apply it to the worker.';
    const setting = view.config.enabled ? `enabled after ${view.config.timeout_minutes} minutes empty` : 'disabled';
    return `Saved idle auto-shutdown: ${setting}. ` + (view.applied
      ? `${local ? 'Applied on this host' : 'Last worker delivery accepted'}. The watchdog checks about every 20 seconds; this is not a live timer readout.`
      : 'Worker application pending. Older policy may still be active; a pending disable is not a confirmed disable. Update/reconnect the worker and retry.');
  }
  function paint(ctx) {
    el(ctx, 'node-idle-status').textContent = ctx.unknown
      ? 'Latest idle-policy state is unknown after the request failed. The worker may have accepted a change. Refresh policy status before retrying; your edited fields are preserved.'
      : summary(ctx.saved, ctx.local);
    el(ctx, 'node-idle-error').textContent = ctx.error;
    el(ctx, 'node-idle-error').hidden = !ctx.error;
    el(ctx, 'node-idle-retry').disabled = ctx.busy || !ctx.saved || ctx.unknown;
    el(ctx, 'node-idle-refresh').disabled = ctx.busy;
    const fieldset = el(ctx, 'node-idle-fields');
    if (fieldset) fieldset.disabled = ctx.busy || !ctx.loaded;
    const save = el(ctx, 'node-idle-save');
    if (save) save.disabled = ctx.busy || !ctx.loaded;
  }
  function accept(ctx, result, initial) {
    if (!valid(result)) throw Error('Unexpected idle-policy response. Update this panel and worker before retrying.');
    ctx.saved = {config: {...result.config}, applied: result.applied, error: result.error};
    ctx.error = result.error;
    ctx.loaded = true;
    ctx.unknown = false;
    if (ctx.local && initial) {
      el(ctx, 'nm-free').checked = result.config.enabled;
      el(ctx, 'nm-idle-timeout').value = result.config.timeout_minutes;
      el(ctx, 'nm-timeout-wrap').style.display = result.config.enabled ? 'block' : 'none';
    }
    for (const node of S.nodes || []) if (node.id === ctx.id) {
      node.is_free = result.config.enabled;
      node.idle_timeout_min = result.config.timeout_minutes;
    }
  }
  async function request(ctx, config) {
    if (!active(ctx) || ctx.busy) return;
    const initial = !ctx.loaded;
    ctx.busy = true;
    ctx.error = '';
    paint(ctx);
    try {
      const result = await api(`/api/admin/nodes/${encodeURIComponent(ctx.id)}/idle-policy`, {
        method: config ? 'PUT' : 'GET', ...(config ? {body:JSON.stringify(config)} : {}), suppressToast:true
      });
      if (active(ctx)) accept(ctx, result, initial);
    } catch (error) {
      if (!active(ctx)) return;
      if (valid(error.data)) accept(ctx, error.data, initial);
      else {
        ctx.unknown = true;
        ctx.error = error.message || 'Idle policy status is unknown. Refresh before retrying.';
        if ([404,405].includes(error.status)) ctx.error += ' Update and restart this panel and the affected worker.';
      }
    } finally {
      ctx.busy = false;
      if (active(ctx)) paint(ctx);
    }
  }
  function attach(id, local = false, saved = null) {
    if (!admin()) return;
    const root = document.querySelector('#node-idle-policy');
    if (!root) return;
    const ctx = {id, local, root, owner:identity(), saved:null, error:'', busy:false, loaded:false, unknown:false};
    current = ctx;
    el(ctx, 'node-idle-retry').onclick = () => ctx.saved && !ctx.unknown && request(ctx, {...ctx.saved.config});
    el(ctx, 'node-idle-refresh').onclick = () => request(ctx);
    if (saved) {
      if (valid(saved)) accept(ctx, saved, false);
      else ctx.error = 'Node saved, but worker policy delivery was not confirmed. Update the panel and worker, then refresh/retry.';
      paint(ctx);
    } else return request(ctx);
  }
  async function saveLocal(event) {
    event.preventDefault();
    const ctx = current;
    if (!ctx || !ctx.local || !active(ctx) || ctx.busy || !ctx.loaded) return;
    const minutes = Number(el(ctx, 'nm-idle-timeout').value);
    if (!Number.isInteger(minutes) || minutes < 1 || minutes > 1440) {
      ctx.error = 'Enter a whole-number timeout from 1 to 1440 minutes.';
      paint(ctx);
      return;
    }
    return request(ctx, {enabled:el(ctx, 'nm-free').checked, timeout_minutes:minutes});
  }
  function savedNode(id, config, result) {
    return attach(id, false, {config, applied:result.idle_policy_applied === true,
      error:result.idle_policy_error || (result.idle_policy_applied === true ? '' : 'Node saved, but idle policy application is unconfirmed. Refresh status and retry after updating the worker.')});
  }
  window.NPNodeIdle = {attach, saveLocal, savedNode, summary, valid};
})();
