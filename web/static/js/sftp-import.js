// Import remote files without saving source credentials in panel state.
(() => {
  'use strict';
  const permissions = ['files.view', 'files.edit', 'sftp.access'];
  const activeStates = new Set(['queued', 'scanning', 'transferring', 'publishing']);
  const states = new Set([...activeStates, 'completed', 'failed', 'cancelled']);
  const GiB = 1073741824;
  let dispose = null;
  const allowed = srv => !!srv?.id && permissions.every(p => hasServerPermission(p, srv));
  const escape = value => String(value ?? '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
  const fingerprintOK = value => /^SHA256:[A-Za-z0-9+/]{43}=?$/.test(value);
  const integer = (value, min, max, label) => {
    if (!Number.isSafeInteger(value) || value < min || value > max) throw Error(`${label} must be between ${min} and ${max}.`);
    return value;
  };
  function relativePath(value) {
    if (!value || value.startsWith('/') || value.includes('\\') || /[\x00-\x1f]/.test(value) || value.split('/').some(p => !p || p === '.' || p === '..')) throw Error('Use exact relative paths, without leading slashes, empty segments or ..');
    return value;
  }
  function validateJob(job, server) {
    if (!job || typeof job.id !== 'string' || !job.id || String(job.server_id) !== String(server.id) || !states.has(job.state) || !['files_total','files_done','bytes_total','bytes_done'].every(k => Number.isSafeInteger(job[k]) && job[k] >= 0) || (job.current_path != null && typeof job.current_path !== 'string')) throw Error('Unexpected import job response. Refresh target to check status.');
    return job;
  }
  const field = (id, title, type = 'text', extra = '') => `<label class="f"><span>${title}</span><input id="sfi-${id}" type="${type}" ${extra}></label>`;
  function markup(server) {
    return `<div class="pal" id="sfi-overlay"><section class="pbox sfi-modal" role="dialog" aria-modal="true" aria-labelledby="sfi-title" aria-describedby="sfi-intro" tabindex="-1">
      <header class="sfi-row"><div><div class="np-eyebrow">Files / Remote transfer</div><h2 id="sfi-title">Import from SFTP</h2></div><button type="button" class="btn sm" id="sfi-close" aria-label="Close SFTP importer">×</button></header>
      <p id="sfi-intro" class="mu">Copy into the root of <strong>${escape(server.name || server.id)}</strong> · node ${escape(server.node_id || 'local')}. The source is never changed or removed.</p>
      <section class="sfi-notice" aria-label="Destination requirements"><div class="sfi-row"><strong>Stopped server. Strictly empty root.</strong><button type="button" class="btn sm" id="sfi-refresh">Refresh target</button></div><p id="sfi-target" role="status">Checking destination…</p><p>New servers generate server.properties and eula.txt. Deliberately move or clear existing files via Files first. This importer never clears, overwrites or starts the server.</p></section>
      <section id="sfi-job" class="sfi-job" hidden aria-label="Import progress"><h3 id="sfi-job-state" role="status"></h3><progress id="sfi-progress" max="100" aria-label="Bytes transferred"></progress><p id="sfi-counts"></p><p id="sfi-current" class="mono"></p><p id="sfi-job-message" role="status"></p><button type="button" class="btn" id="sfi-cancel" hidden>Cancel import</button></section>
      <form id="sfi-form" autocomplete="off" novalidate>
        <fieldset id="sfi-settings"><legend class="sfi-sr">Source and copy options</legend>
          <section class="sfi-section"><h3>1. Connect to the source</h3><div class="sfi-grid">
            ${field('host','SFTP host','text','required placeholder="sftp.example.com" spellcheck="false"')}
            ${field('port','Port','number','required value="22" min="1" max="65535" step="1"')}
          </div><div class="sfi-row"><button type="button" class="btn sm" id="sfi-fetch">Fetch host key</button><span class="mu" id="sfi-key-note">Fetch sends only host and port, never credentials.</span></div>
          ${field('fingerprint','Expected host fingerprint (SHA256)','text','required placeholder="SHA256:…" spellcheck="false" aria-describedby="sfi-verify-note"')}
          <p id="sfi-verify-note" class="mu">A fetched fingerprint is unverified. Compare it with your provider or another trusted, out-of-band source.</p>
          <label class="sfi-check"><input type="checkbox" id="sfi-confirm"><span>I compared this fingerprint independently and trust this host.</span></label>
          <div class="sfi-grid">${field('username','Username','text','required spellcheck="false" autocomplete="off"')}<label class="f"><span>Authentication</span><select id="sfi-auth"><option value="password">Password</option><option value="key">Private key</option></select></label></div>
          <div id="sfi-password-wrap">${field('password','Password','password','autocomplete="new-password"')}</div>
          <div id="sfi-key-wrap" hidden><label class="f"><span>Load a local private key (optional, up to 64 KiB)</span><input id="sfi-key-file" type="file" accept=".key,.pem,text/plain"></label><label class="f"><span>Private key (PEM / OpenSSH)</span><textarea id="sfi-private-key" rows="4" spellcheck="false" autocomplete="off"></textarea></label>${field('passphrase','Key passphrase (if encrypted)','password','autocomplete="new-password"')}</div>
          <p class="mu">SFTP only — no FTP or shell commands. Private hosts require the node operator’s host allowlist configuration.</p></section>
          <section class="sfi-section"><h3>2. Choose what to copy</h3>${field('remote-path','Remote root','text','value="." required spellcheck="false"')}
          <div class="sfi-row"><button type="button" class="btn" id="sfi-test" disabled>Test & preview</button><span class="mu">Testing does not change either server.</span></div>
          <div id="sfi-preview" class="sfi-preview" aria-label="Remote directory preview"><p class="mu">Verify the host key, then test to preview and select files. No selection means copy everything under the remote root.</p></div>
          <details><summary>Advanced copy options</summary><p class="mu">Include/exclude paths are exact relative files or subtrees, not wildcard patterns. Unchecked preview entries are still copied when nothing is selected.</p><label class="f"><span>Exclude paths (one per line)</span><textarea id="sfi-excludes" rows="3" placeholder="logs&#10;backups" spellcheck="false"></textarea></label>
          <div class="sfi-row sfi-presets"><span>Also exclude:</span>${['logs','backups','cache'].map(p => `<label class="sfi-check"><input type="checkbox" data-exclude="${p}"><span>${p}/</span></label>`).join('')}</div>
          <div class="sfi-grid"><label class="sfi-check"><input id="sfi-hidden" type="checkbox"><span>Skip hidden files and folders</span></label><label class="sfi-check"><input id="sfi-times" type="checkbox" checked><span>Preserve modification times</span></label>
          ${field('max','Maximum copy size (GiB)','number','min="0" step="any" value="10" required aria-describedby="sfi-limits"')}
          ${field('bandwidth','Bandwidth (MiB/s; 0 = unlimited)','number','value="0" min="0" max="1024" step="any" required')}
          ${field('timeout','Timeout (minutes)','number','value="60" min="1" max="240" step="1" required')}</div><p class="mu" id="sfi-limits">Waiting for node limits…</p></details></section>
        </fieldset>
        <section class="sfi-notice"><strong>Before you import</strong><p>Stop the source server for a consistent copy; copying a running source is not a snapshot. Files stage first and publish all-or-nothing. The prior empty root is retained as a safety copy, with at most 3 copies shared with restore retention; cleanup is administrator-only.</p><p>After import, verify the startup JAR, runtime and assigned ports before starting manually.</p></section>
        <p id="sfi-error" role="alert" class="sfi-error" tabindex="-1" hidden></p><p id="sfi-status" role="status" class="mu"></p>
        <footer class="sfi-row"><p class="mu">Closing does not cancel an import. Reopen to check its latest status. Node restarts lose status; there is no durable resume.</p><button type="submit" class="btn pr" id="sfi-start" disabled aria-describedby="sfi-start-reason">Start import</button></footer><p id="sfi-start-reason" class="mu">Checking destination and host verification…</p>
      </form></section></div>`;
  }

  function open() {
    const current = curServer();
    if (S.view !== 'fi' || document.hidden || !allowed(current)) return;
    if (dispose) dispose();
    closeModal();
    const server = {...current};
    const base = srvApiPath(server, '/files/import-sftp');
    const token = S.token, user = S.user, userId = S.user?.id;
    const route = window.location.href;
    const trigger = document.activeElement;
    const container = document.querySelector('#modal-container');
    container.innerHTML = markup(server);
    const root = container.querySelector('.sfi-modal');
    const el = id => root.querySelector(`#sfi-${id}`);
    let closed = false, busy = false, checking = false, eligible = false, limits = null;
    let job = null, unknown = false, previousJob = null, cancelPending = false, revision = 0, keyRead = 0;
    let pollTimer, watchTimer, observer;
    const requests = new Map();
    const clearSecrets = () => { keyRead++; ['password','private-key','passphrase','key-file'].forEach(id => { el(id).value = ''; }); };
    const contextMatches = () => {
      const now = curServer();
      return S.token === token && S.user === user && S.user?.id === userId && S.view === 'fi' && window.location.href === route && now?.id === server.id && (now.node_id || 'local') === (server.node_id || 'local') && allowed(now);
    };
    function close() {
      if (closed) return;
      closed = true; keyRead++;
      clearSecrets();
      clearTimeout(pollTimer); clearInterval(watchTimer);
      for (const [request, timer] of requests) { clearTimeout(timer); request.abort(); }
      requests.clear(); observer?.disconnect();
      document.removeEventListener('keydown', keyboard, true);
      document.removeEventListener('visibilitychange', visibility);
      window.removeEventListener('pagehide', close);
      if (root.isConnected) container.replaceChildren();
      if (dispose === close) dispose = null;
      if (contextMatches() && trigger?.isConnected) trigger.focus();
    }
    dispose = close;
    function live() {
      if (closed) return false;
      if (!root.isConnected || document.hidden || !contextMatches()) { close(); return false; }
      return true;
    }
    function visibility() { if (document.hidden) close(); }
    function keyboard(event) {
      if (!live()) return;
      if (event.key === 'Escape') { event.preventDefault(); event.stopImmediatePropagation(); close(); return; }
      if (event.key !== 'Tab') return;
      const list = [...root.querySelectorAll('button, input, textarea, select, summary, [tabindex="0"]')].filter(n => !n.matches(':disabled') && n.getClientRects().length);
      const first = list[0], last = list[list.length - 1];
      if (event.shiftKey && (document.activeElement === first || !root.contains(document.activeElement))) { event.preventDefault(); last?.focus(); }
      else if (!event.shiftKey && (document.activeElement === last || !root.contains(document.activeElement))) { event.preventDefault(); first?.focus(); }
    }
    document.addEventListener('keydown', keyboard, true);
    document.addEventListener('visibilitychange', visibility);
    window.addEventListener('pagehide', close);
    observer = new MutationObserver(() => { live(); });
    observer.observe(document.body, {childList:true, subtree:true});
    watchTimer = setInterval(live, 200);
    el('close').onclick = close;
    container.querySelector('#sfi-overlay').onclick = event => { if (event.target.id === 'sfi-overlay') close(); };
    function error(message = '', focus = false) { el('error').textContent = message; el('error').hidden = !message; if (message && focus) el('error').focus(); }
    function sync() {
      const active = activeStates.has(job?.state);
      const verified = el('confirm').checked && fingerprintOK(el('fingerprint').value.trim());
      const reason = unknown ? 'Start status is unknown. Refresh target; do not submit again.' : active ? 'An import is already in progress.' : checking ? 'Checking destination…' : !eligible ? (el('target').textContent || 'Destination is not eligible.') : !limits ? 'Node limits are unavailable.' : limits.max_bytes < 1 ? 'No space is available within the node import limit.' : !verified ? 'Independently verify and confirm the host fingerprint first.' : busy ? 'Wait for the current request.' : '';
      el('start').disabled = !!reason;
      el('start-reason').textContent = reason || 'Ready to import into the empty, stopped destination root.';
      el('test').disabled = busy || checking || !verified || active;
      el('fetch').disabled = busy || checking || active;
      el('settings').disabled = busy || active || unknown;
      el('refresh').disabled = checking || busy;
      el('cancel').hidden = !active;
      el('cancel').disabled = busy || checking || cancelPending;
      el('cancel').textContent = cancelPending ? 'Cancellation requested…' : 'Cancel import';
    }
    async function request(suffix = '', method = 'GET', body) {
      if (!live()) throw Error('Dialog closed');
      const controller = new AbortController();
      const timer = setTimeout(() => controller.abort(), 30000);
      requests.set(controller, timer);
      try {
        const response = await api(base + suffix, {method, ...(body ? {body:JSON.stringify(body)} : {}), signal:controller.signal, suppressToast:true});
        if (!live()) throw Error('Dialog closed');
        return response;
      } finally { clearTimeout(timer); requests.delete(controller); }
    }
    function showJob(next) {
      job = next;
      el('job').hidden = !job;
      if (!job) return;
      el('job-state').textContent = `${job.state[0].toUpperCase()}${job.state.slice(1)}`;
      if (job.bytes_total > 0) el('progress').value = Math.min(100, job.bytes_done / job.bytes_total * 100);
      else if (job.state === 'completed') el('progress').value = 100;
      else el('progress').removeAttribute('value');
      el('counts').textContent = `${job.files_done.toLocaleString()} / ${job.files_total.toLocaleString()} files · ${(job.bytes_done / GiB).toFixed(3)} / ${(job.bytes_total / GiB).toFixed(3)} GiB`;
      el('current').textContent = job.current_path || '';
      el('job-message').textContent = [job.error, job.message, job.safety_copy ? `Safety copy: ${job.safety_copy}` : '', job.state === 'completed' ? 'Import complete. Verify startup JAR, runtime and ports; the server remains stopped.' : ''].filter(Boolean).join(' ');
      if (!activeStates.has(job.state)) cancelPending = false;
    }
    function schedule() {
      clearTimeout(pollTimer);
      if (live() && (activeStates.has(job?.state) || unknown)) pollTimer = setTimeout(refresh, 1500);
    }
    async function refresh() {
      if (!live() || checking || busy) return;
      checking = true; sync();
      try {
        const data = await request();
        if (!data || typeof data.eligible !== 'boolean' || typeof data.reason !== 'string' || !Object.hasOwn(data, 'job') || !data.limits || !Number.isSafeInteger(data.limits.max_bytes) || data.limits.max_bytes < 0 || !Number.isSafeInteger(data.limits.default_bytes) || data.limits.default_bytes < 0 || !Number.isSafeInteger(data.limits.max_entries) || !Number.isSafeInteger(data.limits.max_active)) throw Error('Unexpected destination response. Start is disabled; refresh to retry.');
        const next = data.job === null ? null : validateJob(data.job, server);
        eligible = data.eligible;
        if (!limits) el('max').value = String(Math.min(data.limits.max_bytes, data.limits.default_bytes || 10 * GiB) / GiB);
        limits = data.limits;
        el('max').max = String(limits.max_bytes / GiB);
        el('limits').textContent = `Node maximum: ${limits.max_bytes / GiB} GiB · ${limits.max_entries.toLocaleString()} entries · ${limits.max_active} active imports maximum. This is an import limit, not an OS quota.`;
        el('target').textContent = eligible ? 'Destination is empty and stopped. Eligibility is checked again when starting.' : data.reason || 'Destination must be strictly empty and stopped.';
        if (unknown && next && next.id !== previousJob) { unknown = false; clearSecrets(); el('status').textContent = 'Import found after the unconfirmed start.'; }
        showJob(next);
        if (!unknown && !busy) error();
      } catch (err) {
        if (!live()) return;
        eligible = false;
        el('target').textContent = [404,405].includes(err.status) ? 'SFTP import is unavailable. Ask the operator to update this node.' : 'Could not confirm destination eligibility. Refresh target to retry.';
        error([404,405].includes(err.status) ? el('target').textContent : err.name === 'AbortError' ? 'Status request timed out. Refresh target to retry.' : err.message);
      } finally {
        if (live()) { checking = false; sync(); schedule(); }
      }
    }
    function hostOptions() {
      const host = el('host').value.trim();
      if (!host || /[\s\/@]/.test(host) || host.includes('://')) throw Error('Enter a host name or IP address, not a URL or username.');
      return {host, port:integer(Number(el('port').value), 1, 65535, 'Port')};
    }
    function options() {
      const connection = hostOptions();
      const fingerprint = el('fingerprint').value.trim();
      if (!fingerprintOK(fingerprint) || !el('confirm').checked) throw Error('Confirm an independently verified SHA256 host fingerprint first.');
      const username = el('username').value.trim();
      if (!username) throw Error('Enter the source username.');
      const max = limits?.max_bytes ?? 10 * GiB;
      const maxBytes = Math.floor(Number(el('max').value) * GiB);
      const bandwidth = Math.floor(Number(el('bandwidth').value) * 1024);
      if (Number(el('bandwidth').value) > 0 && bandwidth === 0) throw Error('Use at least 0.0009765625 MiB/s (1 KiB/s), or 0 for unlimited.');
      const excludes = el('excludes').value.split(/\r?\n/).filter(p => p.length).map(relativePath);
      root.querySelectorAll('[data-exclude]:checked').forEach(n => excludes.push(n.dataset.exclude));
      const result = {...connection, username, fingerprint, remote_path:el('remote-path').value || '.', include_paths:[...root.querySelectorAll('[data-include]:checked')].map(n => relativePath(n.dataset.include)), exclude_paths:[...new Set(excludes)], skip_hidden:el('hidden').checked, preserve_times:el('times').checked, max_bytes:integer(maxBytes, 1, max, 'Copy size in bytes'), bandwidth_kib:integer(bandwidth, 0, 1048576, 'Bandwidth in KiB/s'), timeout_minutes:integer(Number(el('timeout').value), 1, 240, 'Timeout in minutes')};
      if (el('auth').value === 'key') {
        result.private_key = el('private-key').value;
        if (!result.private_key || new TextEncoder().encode(result.private_key).length > 65536) throw Error('Provide a private key no larger than 64 KiB.');
        result.passphrase = el('passphrase').value;
      } else {
        result.password = el('password').value;
        if (!result.password) throw Error('Enter the source password.');
      }
      return result;
    }
    async function action(kind) {
      if (!live() || busy || checking || (kind === 'start' && (!eligible || !limits || unknown || activeStates.has(job?.state)))) return;
      if (kind !== 'cancel' && activeStates.has(job?.state)) return;
      if (kind === 'cancel' && (!activeStates.has(job?.state) || cancelPending)) return;
      let body;
      try { body = kind === 'fetch' ? hostOptions() : kind === 'cancel' ? null : options(); }
      catch (err) { error(err.message, true); return; }
      busy = true; clearTimeout(pollTimer); error(); sync();
      const version = revision;
      const cancelId = job?.id;
      if (kind === 'start') previousJob = job?.id || null;
      el('status').textContent = ({fetch:'Fetching unverified host key…',test:'Verifying host key and reading remote directory…',start:'Requesting import…',cancel:'Requesting cancellation…'})[kind];
      let resolveUnknown = false;
      try {
        const data = await request(kind === 'fetch' ? '/host-key' : kind === 'test' ? '/test' : kind === 'cancel' ? '/' + encodeURIComponent(cancelId) : '', kind === 'cancel' ? 'DELETE' : 'POST', body);
        body = null;
        if (version !== revision) return;
        if (kind === 'fetch') {
          if (!fingerprintOK(data?.fingerprint) || typeof data.algorithm !== 'string') throw Error('Unexpected host key response. No fingerprint was trusted.');
          el('fingerprint').value = data.fingerprint; el('confirm').checked = false;
          el('key-note').textContent = `Unverified · ${data.algorithm}. Compare independently before confirming.`;
          el('status').textContent = 'Host key fetched, not verified.';
        } else if (kind === 'test') {
          if (!data || typeof data.remote_path !== 'string' || typeof data.truncated !== 'boolean' || !Array.isArray(data.entries) || !data.entries.every(e => e && typeof e.name === 'string' && typeof e.path === 'string' && typeof e.is_dir === 'boolean' && Number.isSafeInteger(e.size) && e.size >= 0)) throw Error('Unexpected remote preview response. Test again.');
          for (const entry of data.entries) relativePath(entry.name);
          el('preview').innerHTML = `<p><strong>Remote root:</strong> ${escape(data.remote_path)}</p><p class="mu">Select exact paths to include; leave all unchecked to copy all.</p><div class="sfi-entries">${data.entries.map(e => `<label class="sfi-entry"><input type="checkbox" data-include="${escape(e.name)}"><span>${escape(e.name)}${e.is_dir ? '/' : ''}</span><small class="mu">${e.is_dir ? 'Folder' : e.size.toLocaleString() + ' B'}</small></label>`).join('')}</div><p class="mu">${data.entries.length ? data.truncated ? 'Preview truncated. No selection still copies all paths, including those not shown.' : 'Preview loaded. Edit the remote root above to browse another folder.' : 'This remote folder is empty.'}</p>`;
          el('status').textContent = 'Connection verified. Preview ready; credentials remain in this form for Start.';
          el('preview').scrollIntoView({block:'nearest'});
        } else {
          const next = validateJob(data, server);
          if (kind === 'cancel' && next.id !== cancelId) throw Error('Unexpected cancellation job ID. Refresh target.');
          showJob(next);
          if (kind === 'start') clearSecrets();
          else cancelPending = activeStates.has(next.state);
          el('status').textContent = kind === 'start' ? 'Import accepted. Credentials cleared from this form.' : 'Cancellation requested. Wait for the final job state; publication may already be in progress.';
          el('job').scrollIntoView({block:'nearest'});
        }
      } catch (err) {
        if (!live()) return;
        if (kind === 'start' && (!err.status || err.status >= 500 || err.status === 408)) {
          unknown = true; resolveUnknown = true;
          error('Start was not confirmed. It may have been accepted. Checking status; never retrying the start automatically.', true);
          el('status').textContent = 'If no job appears, ask the operator to check before attempting another import.';
        } else {
          if (kind === 'start') { eligible = false; el('target').textContent = 'Start was rejected. Correct the issue, then refresh target. ' + err.message; }
          el('status').textContent = 'Request failed. No success was confirmed.';
          error(err.name === 'AbortError' ? 'Request timed out. Refresh or test again when ready.' : err.message, true);
        }
      } finally {
        body = null;
        if (live()) { busy = false; sync(); if (resolveUnknown) refresh(); else schedule(); }
      }
    }
    function invalidatePreview() {
      revision++;
      el('preview').innerHTML = '<p class="mu">Source settings changed. Test again to preview and select paths.</p>';
    }
    ['host','port','fingerprint'].forEach(id => el(id).addEventListener('input', () => {
      el('confirm').checked = false;
      el('key-note').textContent = 'Unverified. Compare the expected fingerprint independently.';
      invalidatePreview(); sync();
    }));
    ['username','remote-path'].forEach(id => el(id).addEventListener('input', invalidatePreview));
    el('confirm').onchange = sync;
    el('auth').onchange = () => {
      clearSecrets(); keyRead++; invalidatePreview();
      el('key-wrap').hidden = el('auth').value !== 'key';
      el('password-wrap').hidden = el('auth').value === 'key';
    };
    el('key-file').onchange = async () => {
      const file = el('key-file').files?.[0], version = ++keyRead;
      if (!file) return;
      if (file.size > 65536) { el('key-file').value = ''; error('Private key file exceeds 64 KiB.'); return; }
      try { const text = await file.text(); if (live() && version === keyRead && el('auth').value === 'key') { el('private-key').value = text; error(); } }
      catch (_) { if (live()) error('Could not read the private key file.'); }
    };
    el('fetch').onclick = () => action('fetch');
    el('test').onclick = () => action('test');
    el('form').onsubmit = event => { event.preventDefault(); action('start'); };
    el('cancel').onclick = () => action('cancel');
    el('refresh').onclick = refresh;
    el('close').focus();
    refresh();
  }
  window.NPSFTPImport = Object.freeze({open});
})();
