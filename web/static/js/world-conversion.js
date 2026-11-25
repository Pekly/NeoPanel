// Chunker works on a private copy. Only the separate, explicit replacement flow changes the live world.
(() => {
  'use strict';
  const terminal = new Set(['succeeded', 'failed', 'cancelled', 'blocked', 'replaced']);
  const labels = {preparing:'Preparing copy', cloning:'Copying world', converting:'Converting', validating:'Checking output', publishing:'Publishing output', replacing:'Backing up and replacing', succeeded:'Conversion complete', failed:'Conversion failed', cancelled:'Cancelled', blocked:'Administrator recovery needed', replaced:'World replaced'};
  const warning = 'Conversion is lossy. Entities, structures, custom dimensions and plugin/mod data may be lost. Test the converted world before using it.';
  const esc = value => String(value ?? '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
  const basename = value => typeof value === 'string' && !!value && !/[\\/:\x00-\x1f\x7f]/.test(value) && !value.startsWith('.') && !value.startsWith('converted-');
  const outputPath = value => typeof value === 'string' && /^converted-[A-Za-z0-9_-]+$/.test(value);
  const viewAllowed = srv => !!srv?.id && hasServerPermission('files.view', srv);
  const editAllowed = srv => viewAllowed(srv) && hasServerPermission('files.edit', srv);
  const replaceAllowed = srv => editAllowed(srv) && hasServerPermission('backup.manage', srv);
  let dispose = null;
  // In-memory barriers survive closing/reopening the dialog, including an aborted POST.
  // Never persist tokens or job confirmation tokens in browser storage.
  let ownerToken, ownerUser;
  const uncertain = new Map();
  function validateJob(job, server) {
    if (!job || typeof job.id !== 'string' || !job.id || job.server_id !== server.id || typeof job.state !== 'string' || !job.state || typeof job.input !== 'string' || typeof job.format !== 'string' || typeof job.can_replace !== 'boolean') throw Error('Invalid conversion job response.');
    return job;
  }
  function validateStatus(data, server) {
    if (!data || typeof data.enabled !== 'boolean' || typeof data.supported !== 'boolean' || !Array.isArray(data.formats) || !Array.isArray(data.jobs) || data.jobs.length > 64 || data.formats.length > 256) throw Error('Invalid conversion status response.');
    const ids = new Set();
    for (const f of data.formats) {
      if (!f || typeof f.id !== 'string' || !f.id || typeof f.edition !== 'string' || typeof f.version !== 'string' || ids.has(f.id)) throw Error('Invalid conversion formats.');
      ids.add(f.id);
    }
    ids.clear();
    for (const j of data.jobs) { validateJob(j, server); if (ids.has(j.id)) throw Error('Duplicate conversion job.'); ids.add(j.id); }
    return data;
  }
  function markup(server, admin) {
    return `<div class="pal" id="wc-overlay"><section class="pbox wc-modal" role="dialog" aria-modal="true" aria-labelledby="wc-title" aria-describedby="wc-intro" tabindex="-1">
      <header class="wc-row"><div><div class="np-eyebrow">Files / World tools</div><h2 id="wc-title">Convert a world</h2></div><button class="btn sm" type="button" id="wc-close" aria-label="Close world conversion">×</button></header>
      <p id="wc-intro">Convert a copy of a world on <strong>${esc(server.name || server.id)}</strong> with Chunker. Your original stays in place; output goes into a separate folder.</p>
      <section class="wc-notice"><div class="wc-row"><strong>Keep the server stopped</strong><button class="btn sm" type="button" id="wc-refresh">Refresh status</button></div><p id="wc-target" role="status">Checking server and conversion availability…</p><p class="mu">This tool never stops or starts the server for you.</p></section>
      <section class="wc-warning" aria-label="Conversion warning"><strong>Some world data will not survive conversion</strong><p>${warning}</p><p class="mu">Only a direct world folder is accepted, not a ZIP or a nested path. Uploaded .chunker.json settings are not supported.</p></section>
      <p class="mu">The converter uses this server’s assigned CPU and RAM limits. Its temporary world copies also count toward RAM usage, so large worlds may need a larger allocation.</p>
      <form id="wc-form" autocomplete="off" novalidate><fieldset id="wc-settings"><legend>New conversion</legend><div class="wc-grid">
        <label class="f"><span>World folder</span><input id="wc-input" required placeholder="world" spellcheck="false" aria-describedby="wc-input-help"><small id="wc-input-help" class="mu">Folder name at the root of Files.</small></label>
        <label class="f"><span>Output edition &amp; version</span><select id="wc-format" required><option value="">Choose an output format…</option></select><small class="mu">Choose deliberately; versions are not auto-matched.</small></label>
      </div><label class="wc-check"><input id="wc-loss" type="checkbox"><span>I understand this conversion may lose world data.</span></label></fieldset>
      <div class="wc-row"><p id="wc-reason" class="mu">Checking availability…</p><button id="wc-start" type="submit" class="btn pr" disabled aria-describedby="wc-reason">Convert copy</button></div></form>
      <p id="wc-error" class="wc-error" role="alert" tabindex="-1" hidden></p><p id="wc-message" role="status" class="mu"></p>
      <section class="wc-history" aria-labelledby="wc-history-title"><h3 id="wc-history-title">Conversion jobs</h3><p id="wc-empty" class="mu">Loading jobs…</p><div id="wc-jobs"></div></section>
      <form id="wc-replace-form" class="wc-replace" aria-labelledby="wc-replace-title" hidden novalidate><h3 id="wc-replace-title">Replace the live world?</h3><p id="wc-replace-description"></p><p>A world-only backup and a safety copy are required before replacement. The server stays stopped. Inspect the result before starting it.</p>
        <label class="wc-check"><input id="wc-backup" type="checkbox"><span>Create the required backup and safety copy.</span></label>
        <label class="wc-check"><input id="wc-replace-loss" type="checkbox"><span>I accept possible data loss in the replacement world.</span></label>
        <label class="f"><span>Type REPLACE to confirm</span><input id="wc-confirm" autocomplete="off" spellcheck="false"></label>
        <p id="wc-replace-reason" class="mu"></p><div class="wc-row"><button id="wc-replace-back" type="button" class="btn">Keep original</button><button id="wc-replace-submit" type="submit" class="btn" disabled>Back up &amp; replace</button></div>
      </form>
      ${admin ? '<details class="wc-notice"><summary>Operator details</summary><p id="wc-operator"></p><pre id="wc-technical"></pre></details>' : ''}
      <footer class="mu">Closing this window does not cancel a job. Reopen it to check progress. Jobs are not resumed after a service restart.</footer>
    </section></div>`;
  }
  function open() {
    const current = curServer();
    if (S.view !== 'fi' || document.hidden || !viewAllowed(current)) return;
    if (dispose) dispose();
    closeModal();
    if (ownerToken !== S.token || ownerUser !== S.user) { uncertain.clear(); ownerToken = S.token; ownerUser = S.user; }
    const server = Object.freeze({...current}), token = S.token, user = S.user, userId = S.user?.id, role = S.user?.role;
    const admin = role === 'admin', route = window.location.href, trigger = document.activeElement;
    const base = srvApiPath(server, '/world-conversion'), serverURL = srvApiPath(server);
    const key = JSON.stringify([server.node_id || 'local', server.id]);
    const container = document.querySelector('#modal-container');
    if (!container) return;
    container.innerHTML = markup(server, admin);
    const root = container.querySelector('.wc-modal'), el = id => root.querySelector('#wc-' + id);
    const cards = new Map(), requests = new Map(), cancelPending = new Set();
    let closed = false, busy = false, checking = false, data = null, fresh = false, stopped = false, selected = null;
    let pollTimer, watchTimer, observer, formatSignature = '';
    const contextMatches = () => {
      const now = curServer();
      return S.token === token && S.user === user && S.user?.id === userId && S.user?.role === role && S.view === 'fi' && window.location.href === route && now?.id === server.id && (now.node_id || 'local') === (server.node_id || 'local') && viewAllowed(now);
    };
    function close() {
      if (closed) return;
      closed = true;
      clearTimeout(pollTimer); clearInterval(watchTimer);
      for (const [controller, timer] of requests) { clearTimeout(timer); controller.abort(); }
      requests.clear(); observer?.disconnect();
      document.removeEventListener('keydown', keyboard, true);
      window.removeEventListener('pagehide', close);
      if (root.isConnected) container.replaceChildren();
      if (dispose === close) dispose = null;
      if (contextMatches() && trigger?.isConnected) trigger.focus();
    }
    dispose = close;
    function live() {
      if (closed) return false;
      if (!root.isConnected || !contextMatches()) { close(); return false; }
      return true;
    }
    function keyboard(event) {
      if (!live()) return;
      if (event.key === 'Escape') { event.preventDefault(); event.stopImmediatePropagation(); close(); return; }
      if (event.key !== 'Tab') return;
      const list = [...root.querySelectorAll('button, input, select, summary, [tabindex="0"]')].filter(n => !n.matches(':disabled') && n.getClientRects().length);
      const first = list[0], last = list[list.length - 1];
      if (event.shiftKey && (document.activeElement === first || !root.contains(document.activeElement))) { event.preventDefault(); last?.focus(); }
      else if (!event.shiftKey && (document.activeElement === last || !root.contains(document.activeElement))) { event.preventDefault(); first?.focus(); }
    }
    function error(message = '') { el('error').textContent = message; el('error').hidden = !message; }
    function safeError(err, fallback) { return admin ? `${fallback} ${err.message || ''}` : fallback; }
    async function request(url, method = 'GET', body) {
      if (!live()) throw Error('Dialog closed.');
      const controller = new AbortController(), timer = setTimeout(() => controller.abort(), 30000);
      requests.set(controller, timer);
      try {
        const response = await api(url, {method, ...(body ? {body:JSON.stringify(body)} : {}), signal:controller.signal, suppressToast:true});
        if (!live()) throw Error('Dialog closed.');
        return response;
      } finally { clearTimeout(timer); requests.delete(controller); }
    }
    async function checkServer() {
      stopped = false;
      const response = await request(serverURL);
      if (!response || response.id !== server.id || typeof response.stats?.status !== 'string') throw Error('Invalid server status response.');
      stopped = response.stats.status === 'offline';
      return stopped;
    }
    const active = job => !terminal.has(job.state);
    const jobById = id => data?.jobs.find(j => j.id === id);
    function availability() {
      if (!editAllowed(curServer())) return 'You need Files view and edit permissions to convert or cancel.';
      if (uncertain.has(key)) return 'A request outcome is unconfirmed. Do not submit it again; checking jobs for confirmation.';
      if (!fresh || !data) return 'Availability could not be verified. Refresh status before continuing.';
      if (!data.enabled) return 'World conversion is disabled. Existing jobs can still be monitored or cancelled.';
      if (!data.supported) return 'Safe conversion is unavailable on this server. Ask an administrator to check support and resource limits.';
      if (!stopped) return 'Stop the server first, then refresh status. Nothing will stop it automatically.';
      if (data.jobs.some(j => active(j) || j.state === 'blocked')) return 'Finish the current job or ask an administrator to resolve a blocked job first.';
      return '';
    }
    function replaceEligible(job) {
      return !!job && job.state === 'succeeded' && job.can_replace === true && typeof job.replace_token === 'string' && !!job.replace_token && replaceAllowed(curServer()) && data?.formats.some(f => f.id === job.format && f.edition === 'java');
    }
    function sync() {
      const reason = busy ? 'Waiting for the current request…' : availability() || (!basename(el('input').value.trim()) ? 'Enter a direct world folder name.' : !data.formats.some(f => f.id === el('format').value) ? 'Choose the output edition and version.' : !el('loss').checked ? 'Acknowledge possible data loss to continue.' : '');
      el('start').disabled = !!reason || checking;
      el('reason').textContent = reason || (checking ? 'Refreshing status…' : 'Ready. A separate converted folder will be created.');
      el('settings').disabled = busy || !!uncertain.get(key) || !editAllowed(curServer());
      el('refresh').disabled = busy || checking;
      ['backup', 'replace-loss', 'confirm', 'replace-back'].forEach(id => { el(id).disabled = busy; });
      for (const [id, card] of cards) {
        const job = jobById(id); if (!job) continue;
        const cancel = card.querySelector('[data-cancel]'), replace = card.querySelector('[data-replace]');
        cancel.hidden = !active(job) || !editAllowed(curServer());
        cancel.disabled = busy || checking || cancelPending.has(id);
        cancel.textContent = cancelPending.has(id) ? 'Cancellation requested…' : 'Cancel job';
        replace.hidden = !replaceEligible(job);
        replace.disabled = busy || checking || !!availability();
      }
      if (selected) {
        const job = jobById(selected.id);
        const why = busy ? 'Waiting for the current request…' : availability() || (!replaceEligible(job) || job.replace_token !== selected.token || job.format !== selected.format ? 'This job is no longer eligible. Keep the original and refresh status.' : !el('backup').checked || !el('replace-loss').checked || el('confirm').value !== 'REPLACE' ? 'Confirm the backup, data loss and exact word REPLACE.' : '');
        el('replace-submit').disabled = !!why || checking;
        el('replace-reason').textContent = why || 'Ready to back up and replace the configured world.';
      }
    }
    function jobCard(job, index) {
      const card = document.createElement('article');
      card.className = 'wc-job'; card.style.setProperty('--wc-delay', Math.min(index, 4) * 25 + 'ms');
      card.innerHTML = `<div class="wc-row"><h4 data-title></h4><span data-state class="wc-state" role="status"></span></div><progress max="100" aria-label="Conversion progress"></progress><p data-description></p><p data-output class="mono"></p><p data-error class="wc-error"></p><div class="wc-row wc-job-actions"><button type="button" class="btn sm" data-open hidden>Open output in Files</button><button type="button" class="btn sm" data-cancel hidden>Cancel job</button><button type="button" class="btn sm" data-replace hidden>Review replacement…</button></div>${admin ? '<details><summary>Job diagnostics</summary><pre data-log></pre></details>' : ''}`;
      card.querySelector('[data-cancel]').onclick = () => action('cancel', job.id);
      card.querySelector('[data-replace]').onclick = () => {
        const currentJob = jobById(job.id);
        if (!live() || busy || checking || availability() || !replaceEligible(currentJob)) return;
        selected = {id:currentJob.id, token:currentJob.replace_token, format:currentJob.format};
        el('replace-description').textContent = `Replace “${basename(currentJob.input) ? currentJob.input : 'configured world'}” with the converted ${currentJob.format} output? This changes the live world, not just the export folder.`;
        el('backup').checked = false; el('replace-loss').checked = false; el('confirm').value = '';
        el('replace-form').hidden = false; sync(); el('confirm').focus();
      };
      card.querySelector('[data-open]').onclick = () => {
        const path = jobById(job.id)?.output;
        if (!live() || !outputPath(path) || typeof loadFileList !== 'function') return;
        close();
        // Existing navigation owns the fetch, render, and stale-server protection. Never construct a Files URL.
        if (contextMatches()) loadFileList(path, false, {isCurrent:contextMatches});
      };
      el('jobs').appendChild(card); cards.set(job.id, card);
      return card;
    }
    function renderJobs() {
      el('empty').textContent = data.jobs.length ? '' : 'No conversion jobs yet. Your original world will remain untouched when you convert a copy.';
      el('empty').hidden = !!data.jobs.length;
      for (const [id, card] of cards) if (!jobById(id)) { card.remove(); cards.delete(id); }
      data.jobs.forEach((job, index) => {
        const card = cards.get(job.id) || jobCard(job, index), q = s => card.querySelector(s);
        const f = data.formats.find(f => f.id === job.format);
        q('[data-title]').textContent = `${basename(job.input) ? job.input : 'World'} → ${f ? `${f.edition === 'java' ? 'Java' : f.edition === 'bedrock' ? 'Bedrock' : 'Minecraft'} ${f.version}` : 'Selected format'}`;
        q('[data-state]').textContent = labels[job.state] || 'In progress';
        card.dataset.state = terminal.has(job.state) ? job.state : 'active';
        const progress = q('progress');
        progress.hidden = terminal.has(job.state);
        if (typeof job.progress === 'number' && Number.isFinite(job.progress) && job.progress > 0) progress.value = Math.min(100, Math.max(0, job.progress));
        else progress.removeAttribute('value');
        q('[data-description]').textContent = job.state === 'succeeded' ? 'Separate output is ready. Inspect and test it before use. Replacement is only available for compatible Java worlds.' : job.state === 'replaced' ? 'Live world replaced. Backup and safety copy retained. Review it before starting the server.' : job.state === 'blocked' ? 'Do not start this server. Ask an administrator to inspect recovery details.' : job.state === 'failed' ? 'The operation did not complete safely. Ask an administrator to inspect the job before retrying.' : job.state === 'cancelled' ? 'Job cancelled. Check Files and any retained recovery copies before continuing.' : 'Working on this job. The server must stay stopped until cleanup finishes.';
        q('[data-output]').textContent = outputPath(job.output) ? `Output in Files: ${job.output}/` : '';
        q('[data-open]').hidden = !outputPath(job.output) || !['succeeded','replaced'].includes(job.state) || typeof loadFileList !== 'function';
        q('[data-error]').textContent = job.error ? (admin ? String(job.error) : 'Contact an administrator for safe recovery guidance.') : '';
        q('[data-error]').hidden = !job.error;
        if (admin) q('[data-log]').textContent = [`Job: ${job.id}`, `State: ${job.state}`, job.backup ? `World-only backup: ${job.backup}` : '', job.safety_copy ? `Safety copy: ${job.safety_copy}` : '', String(job.log || '').slice(-32768)].filter(Boolean).join('\n');
        if (terminal.has(job.state)) cancelPending.delete(job.id);
      });
    }
    function reconcile() {
      const pending = uncertain.get(key); if (!pending) return;
      const found = pending.kind === 'start' ? data.jobs.some(j => !pending.ids.includes(j.id) && j.input === pending.input && j.format === pending.format) : data.jobs.some(j => j.id === pending.id && (pending.kind === 'cancel' ? terminal.has(j.state) : j.state !== 'succeeded' && j.replace_token !== pending.token));
      if (found) { uncertain.delete(key); error(); el('message').textContent = 'The latest jobs confirm the earlier request. Review the result below.'; }
    }
    function schedule() { clearTimeout(pollTimer); if (live()) pollTimer = setTimeout(refresh, 2000); }
    async function refresh() {
      if (!live() || busy || checking) return;
      checking = true; sync();
      try {
        data = validateStatus(await request(base), server);
        renderJobs(); reconcile();
        const signature = JSON.stringify(data.formats);
        if (signature !== formatSignature) {
          const value = el('format').value;
          el('format').innerHTML = '<option value="">Choose an output format…</option>' + data.formats.map(f => `<option value="${esc(f.id)}">${esc(f.edition === 'java' ? 'Java' : f.edition === 'bedrock' ? 'Bedrock' : f.edition)} · ${esc(f.version)}</option>`).join('');
          el('format').value = data.formats.some(f => f.id === value) ? value : ''; formatSignature = signature;
        }
        fresh = true;
        try { await checkServer(); } catch (err) { stopped = false; error(safeError(err, 'Could not verify whether the server is stopped. Refresh status to retry.')); }
        if (!live()) return;
        el('target').textContent = !data.enabled ? 'Conversion is disabled. Existing jobs remain available below.' : !data.supported ? 'Conversion cannot run safely on this server. Contact an administrator.' : stopped ? 'Server is stopped. Conversion output stays separate.' : 'Server is not confirmed stopped. Stop it before converting or replacing.';
        if (admin) { el('operator').textContent = `Chunker ${data.chunker_version || 'version unknown'} · Hosting node: ${server.node_id || 'local'}. Backups here are world-only, not whole-server restore archives.`; el('technical').textContent = String(data.unavailable_reason || 'No platform restriction reported.'); }
      } catch (err) {
        if (!live()) return;
        fresh = false; stopped = false;
        el('target').textContent = 'Could not check conversion availability. Existing job information may be out of date.';
        el('empty').textContent = data?.jobs.length ? '' : 'Job status unavailable. Refresh to try again.';
        error(safeError(err, 'Status could not be loaded. Refresh to retry, or contact an administrator.'));
      } finally { if (live()) { checking = false; sync(); schedule(); } }
    }
    async function action(kind, id) {
      if (!live() || busy || checking || !editAllowed(curServer())) return;
      const job = jobById(id);
      if (kind === 'cancel' && (!job || !active(job) || cancelPending.has(id))) return;
      if (kind !== 'cancel' && availability()) return;
      const input = el('input').value.trim(), format = el('format').value;
      if (kind === 'start' && (!basename(input) || !data.formats.some(f => f.id === format) || !el('loss').checked)) return;
      if (kind === 'replace' && (!selected || selected.id !== id || !replaceEligible(job) || job.replace_token !== selected.token || job.format !== selected.format || !el('backup').checked || !el('replace-loss').checked || el('confirm').value !== 'REPLACE')) return;
      busy = true; clearTimeout(pollTimer); error(); sync();
      let sent = false;
      const marker = kind === 'start' ? {kind, input, format, ids:data.jobs.map(j => j.id)} : {kind, id, token:job.replace_token};
      try {
        if (kind !== 'cancel' && !await checkServer()) { error('Stop the server before continuing. It will not be stopped automatically.'); return; }
        if (!live() || !editAllowed(curServer()) || (kind === 'replace' && !replaceAllowed(curServer()))) return;
        // Set before dispatch: even closing during fetch must preserve the no-retry barrier.
        if (kind !== 'cancel') uncertain.set(key, marker);
        sent = true;
        const body = kind === 'start' ? {input, format, acknowledge_loss:true} : kind === 'replace' ? {token:marker.token, confirm:'REPLACE', backup:true, acknowledge_loss:true} : undefined;
        const response = await request(base + (kind === 'start' ? '' : '/' + encodeURIComponent(id) + (kind === 'replace' ? '/replace' : '')), kind === 'cancel' ? 'DELETE' : 'POST', body);
        if (kind !== 'cancel') {
          const next = validateJob(response, server);
          if ((kind === 'replace' && (next.id !== id || next.state === 'succeeded' || next.replace_token === marker.token)) || (kind === 'start' && (next.input !== input || next.format !== format || marker.ids.includes(next.id)))) throw Error('Unexpected accepted job.');
          uncertain.delete(key);
          data.jobs = [next, ...data.jobs.filter(j => j.id !== next.id)]; renderJobs();
        } else cancelPending.add(id);
        if (kind === 'replace') { selected = null; el('replace-form').hidden = true; }
        if (kind === 'start') el('loss').checked = false;
        el('message').textContent = kind === 'cancel' ? 'Cancellation requested. Wait for cleanup; the server stays stopped.' : kind === 'replace' ? 'Backup and replacement requested. The server will not start automatically.' : 'Conversion accepted. The original stays in place.';
      } catch (err) {
        if (!live()) return;
        const rejected = Number.isInteger(err.status) && err.status >= 400 && err.status < 500 && err.status !== 408;
        if (rejected && uncertain.get(key) === marker) uncertain.delete(key);
        if (kind === 'cancel' && sent && !rejected) cancelPending.add(id);
        error(safeError(err, sent && !rejected ? 'Request outcome is unknown. Do not repeat it. Status will be checked; contact an administrator if it remains unconfirmed.' : 'Request was not accepted. Check permissions, server status and the selected world, then refresh.'));
      } finally {
        if (live()) { busy = false; sync(); schedule(); }
      }
    }
    document.addEventListener('keydown', keyboard, true);
    window.addEventListener('pagehide', close);
    observer = new MutationObserver(() => { if (!root.isConnected) close(); });
    observer.observe(document.body, {childList:true, subtree:true});
    watchTimer = setInterval(() => { if (live()) sync(); }, 200);
    el('close').onclick = close;
    container.querySelector('#wc-overlay').onclick = event => { if (event.target.id === 'wc-overlay') close(); };
    el('refresh').onclick = () => { error(); return refresh(); };
    el('form').onsubmit = event => { event.preventDefault(); return action('start'); };
    el('replace-form').onsubmit = event => { event.preventDefault(); return action('replace', selected?.id); };
    el('replace-back').onclick = () => { if (busy) return; const card = cards.get(selected?.id); selected = null; el('replace-form').hidden = true; const back = card?.querySelector('[data-replace]'); (back && !back.hidden && !back.disabled ? back : el('close')).focus(); };
    ['input','format','loss','backup','replace-loss','confirm'].forEach(id => { el(id).addEventListener('input', sync); el(id).addEventListener('change', sync); });
    el('close').focus(); refresh();
  }
  const reset = () => { if (dispose) dispose(); };
  window.NPWorldConversion = {open, reset, dispose:reset};
})();
