// Admin domain wizard. Verification tickets and passwords are never persisted.
(() => {
  'use strict';
  const state = { owner: '', sequence: 0, phase: '', report: null, message: '', controller: null, timer: null, attempts: 0, prepared: '', email: '' };
  const identity = () => typeof S === 'undefined' ? '' : `${S.user?.id || ''}|${S.token || ''}`;
  const admin = () => S.user?.role === 'admin' && Boolean(S.token);
  const candidate = () => panelAddressDraftState(S.panelUrlDraft ?? S.adminPanelUrl?.public_url ?? '').url;
  function reset() {
    state.sequence++; state.controller?.abort(); clearTimeout(state.timer);
    Object.assign(state, { owner: '', phase: '', report: null, message: '', controller: null, timer: null, attempts: 0, prepared: '', email: '' });
  }
  function ensure() { if (state.owner !== identity()) { reset(); state.owner = identity(); } }
  function busy() { ensure(); return Boolean(state.phase); }
  function invalidate() {
    ensure(); state.sequence++; state.controller?.abort(); clearTimeout(state.timer);
    Object.assign(state, { phase: '', report: null, message: '', controller: null, timer: null, attempts: 0, prepared: '' });
    const root = document.querySelector('#np-domain-checks');
    if (root) root.outerHTML = render();
    const protocol = document.querySelector('.np-domain-protocol select');
    if (protocol) protocol.value = /^http:\/\//i.test(String(S.panelUrlDraft ?? S.adminPanelUrl?.public_url ?? '')) ? 'http' : 'https';
  }
  function validReport(report, origin) {
    const checks = report?.checks;
    return report?.public_url === origin && Array.isArray(checks) && checks.every(c => c && typeof c.id === 'string' && typeof c.detail === 'string') && ['dns','https','panel','passkeys'].every(id => checks.some(c => c.id === id && ['passed','warning','failed','blocked','waiting'].includes(c.status))) && Array.isArray(report.passkey_accounts) && report.passkey_accounts.every(a => a && typeof a.username === 'string' && typeof a.hostname === 'string') && Array.isArray(report.certificates) && report.certificates.every(c => c && typeof c.issuer === 'string' && Number.isFinite(Date.parse(c.expires_at))) && Array.isArray(report.addresses) && report.addresses.every(a => typeof a === 'string') && Number.isFinite(Date.parse(report.expires_at));
  }
  function canApply() {
    ensure(); const report = state.report;
    return admin() && !state.phase && Boolean(report && report.public_url === candidate() && report.ready === true && /^[a-f0-9]{48}$/.test(report.check_token || '') && Date.parse(report.expires_at) > Date.now() && ['dns','panel','passkeys'].every(id => report.checks.some(c => c.id === id && c.status === 'passed')) && report.checks.some(c => c.id === 'https' && ['passed','warning'].includes(c.status)));
  }
  function canSetup() {
    const report = state.report, origin = candidate();
    if (!admin() || state.phase || !report || report.public_url !== origin || !report.dns_ready || !report.https_setup?.available || !report.check_token || Date.parse(report.expires_at) <= Date.now() || state.prepared === origin) return false;
    const url = new URL(origin);
    return url.protocol === 'https:' && !url.port && !['localhost','127.0.0.1','[::1]'].includes(url.hostname) && report.checks.some(c => c.id === 'https' && !['passed','warning'].includes(c.status));
  }
  function protocolControl() {
    ensure(); const raw = String(S.panelUrlDraft ?? S.adminPanelUrl?.public_url ?? '');
    const http = /^http:\/\//i.test(raw);
    return `<label class="f np-domain-protocol"><span>Connection</span><select aria-label="Panel connection protocol" onchange="NPDomain.setProtocol(this.value)" ${busy() || S.panelUrlLoading ? 'disabled' : ''}><option value="https" ${http ? '' : 'selected'}>HTTPS — public panel</option><option value="http" ${http ? 'selected' : ''}>HTTP — localhost development only</option></select></label>`;
  }
  function setProtocol(protocol) {
    if (busy() || !admin() || !['http','https'].includes(protocol)) return;
    const raw = String(S.panelUrlDraft ?? S.adminPanelUrl?.public_url ?? '').trim().replace(/^https?:\/\//i, '');
    onPanelUrlInputChange(`${protocol}://${raw}`); R();
  }
  function render() {
    ensure(); if (!admin()) return '';
    const report = state.report, origin = candidate();
    const labels = { passed: 'Passed', warning: 'Review', failed: 'Needs attention', blocked: 'Action required', waiting: 'Not checked' };
    const checks = report && report.public_url === origin ? report.checks : [];
    return `<section id="np-domain-checks" class="np-domain-checks" aria-label="Domain verification">
      <div class="row sb"><div><h3>Verify before changing</h3><p class="mu">DNS, HTTPS, panel identity, and passkeys are checked together. Your saved address stays active until you apply the verified change.</p></div><button type="button" id="domain-check-btn" class="btn" onclick="NPDomain.check()" ${state.phase || !origin || S.panelUrlLoading ? 'disabled' : ''}>${state.phase === 'checking' ? 'Checking domain…' : 'Check domain & HTTPS'}</button>${state.phase === 'checking' ? '<button type="button" class="btn sm" onclick="NPDomain.cancelCheck()">Cancel check</button>' : ''}</div>
      ${state.phase === 'applying' ? '<p role="status">Rechecking the address and applying the change…</p>' : ''}
      ${state.phase === 'setup' ? '<p role="status">Preparing Caddy and requesting HTTPS. Keep this page open; this does not change your saved panel address.</p>' : ''}
      ${state.message ? `<p class="np-domain-message" role="status">${esc(state.message)}</p>` : ''}
      ${checks.length ? `<ol class="np-domain-results">${checks.map(check => `<li class="np-domain-result ${['passed','warning'].includes(check.status) ? check.status : 'attention'}"><div class="row sb"><b>${esc(check.label)}</b><span>${esc(labels[check.status] || 'Not checked')}</span></div><p>${esc(check.detail)}</p>${check.action ? `<p class="mu">${esc(check.action)}</p>` : ''}</li>`).join('')}</ol><p id="domain-check-expiry" class="mu">${Date.parse(report.expires_at) > Date.now() ? 'Checks expire in five minutes. Applying rechecks the address.' : 'These checks expired. Check the address again.'}</p>` : ''}
      ${report?.addresses?.length ? `<p class="mu">DNS addresses: <span class="mono">${esc(report.addresses.join(', '))}</span></p>` : ''}
      ${report?.certificates?.length ? `<div class="np-domain-certificates"><b>Certificates served at this address</b>${report.certificates.map(cert => `<p>${esc(cert.address)} — ${esc(cert.issuer || 'Trusted issuer')}${cert.lets_encrypt ? ' · Let’s Encrypt' : ''} · expires ${esc(new Date(cert.expires_at).toLocaleDateString())}</p>`).join('')}<p class="mu">Other trusted certificate providers also work. If you use a CDN, this describes its public-facing certificate.</p></div>` : ''}
      ${report?.passkey_accounts?.length ? `<div class="np-domain-passkeys"><h3>Accounts needing new passkeys</h3><ul>${report.passkey_accounts.map(account => `<li><b>${esc(account.username)}</b> — ${esc(account.count)} passkey(s) for ${esc(account.hostname)}</li>`).join('')}</ul><p class="mu">Confirm password sign-in first. These users can remove old keys in Account → Security and register replacements after the move. No keys are removed automatically.</p><div class="row"><a class="btn sm" href="/account/security" onclick="return handleNavClick(event, 'acc', {acctTab:1})">My passkeys</a><a class="btn sm" href="/admin/users" onclick="return handleNavClick(event, 'admin', {adminTab:3})">Manage users</a></div></div>` : ''}
      ${canSetup() ? `<fieldset class="np-domain-setup"><legend>Set up HTTPS with Let’s Encrypt</legend><p>For this recognized Caddy installation, NeoPanel can add the new hostname without removing the old one. Caddy manages certificates and renewal. DNS must point at this host and public ports 80/443 must reach Caddy.</p><label class="f"><span>Certificate contact email</span><input id="domain-acme-email" type="email" value="${esc(state.email || S.user?.email || '')}" autocomplete="email" oninput="NPDomain.setEmail(this.value)"></label><label class="f"><span>Your administrator password</span><input id="domain-admin-password" type="password" autocomplete="current-password"></label><label class="np-domain-confirm"><input id="domain-https-confirm" type="checkbox"> I control this domain and agree to the <a href="https://letsencrypt.org/repository/" target="_blank" rel="noopener noreferrer">Let’s Encrypt subscriber agreement</a>. Add HTTPS while keeping the old hostname.</label><button type="button" class="btn" onclick="NPDomain.setupHTTPS()">Set up HTTPS</button></fieldset>` : report && !report.https_setup?.available ? `<p class="mu"><b>Automatic HTTPS setup:</b> ${esc(report.https_setup?.reason || 'Not available on this installation.')} Configure the new hostname and certificate at your existing proxy, then check again. Nothing will be overwritten automatically.</p>` : ''}
      ${!state.phase && (S.panelUrlError || (report && !report.ready)) ? `<aside class="np-domain-recovery"><h3>Still unable to change the address?</h3><p>Connect to the panel host. On Linux, rerun <code>sudo neopanel-installer</code> and choose <b>12 — Change panel address / FQDN</b>. This is a recovery option, not a reinstall. If the command is missing, rerun the current installer script; older panel binaries must be updated first.</p><p>Configure DNS and trusted HTTPS and check the new address in your browser first. Host recovery saves the address manually; it does not verify or provision HTTPS, restart servers, or bypass passkey requirements. On Windows, run the installed <code>neopanel.exe -config config.yaml panel-address --url https://panel.example.com</code> from the host, using your actual config path and domain.</p></aside>` : ''}
      <p class="mu">Saving changes panel links and the passkey hostname together. It does not move game servers, edit DNS records, or change node/SFTP addresses.</p>
    </section>`;
  }
  function renderSettingsLink(context) {
    const address = S.adminPanelUrl?.public_url;
    const purpose = context === 'email' ? 'Invitation and password-reset links' : 'Passkeys';
    const detail = context === 'email' ? 'SMTP delivery and the address in its links are separate settings.' : 'HTTPS and the sign-in hostname must agree.';
    const summary = address ? `${purpose} use: <b>${esc(address)}</b>.` : `${purpose} depend on the public panel address. Open Panel address to review the configured hostname.`;
    return `<div class="np-domain-settings-note"><span>${summary} ${detail}</span><a class="btn sm" href="/admin/panel-address" onclick="return handleNavClick(event, 'admin', {adminTab:8})">Check domain & HTTPS</a></div>`;
  }
  function sync() {
    if (typeof S === 'undefined' || !document.querySelector('#np-domain-checks')) return;
    ensure();
    const save = document.querySelector('#panel-url-save-btn');
    if (save) save.disabled = S.panelUrlLoading || S.panelUrlSaving || !S.adminPanelUrl || !canApply();
    const expiry = document.querySelector('#domain-check-expiry');
    if (expiry && state.report && Date.parse(state.report.expires_at) <= Date.now()) expiry.textContent = 'These checks expired. Check the address again.';
  }
  function cancelCheck() {
    if (state.phase !== 'checking') return;
    invalidate(); state.message = 'Check cancelled. The saved address is unchanged.'; R();
  }
  async function request(endpoint, body, timeout = 30000) {
    const controller = new AbortController(); state.controller = controller;
    const timer = setTimeout(() => controller.abort(), timeout);
    try { return await api(endpoint, { method: 'POST', body: JSON.stringify(body), signal: controller.signal, suppressToast: true }); }
    finally { clearTimeout(timer); if (state.controller === controller) state.controller = null; }
  }
  function failure(error) {
    if ([404,405].includes(error?.status)) return 'Domain verification is not available on this backend. Update and restart the panel, then try again.';
    if (error?.name === 'AbortError') return 'The request timed out. Check again before retrying a change; a started HTTPS operation may still have finished.';
    return error?.message || 'The domain operation failed. Your saved address has not been changed by this request.';
  }
  function scheduleCheck() {
    clearTimeout(state.timer);
    if (state.attempts <= 0) return;
    state.timer = setTimeout(() => { state.timer = null; if (admin() && state.owner === identity() && candidate() === state.prepared) check(true); }, 5000);
  }
  async function check(automatic = false) {
    ensure(); if (!admin() || state.phase || S.panelUrlLoading) return;
    const origin = candidate(); if (!origin) return;
    if (!automatic) { clearTimeout(state.timer); state.attempts = 0; }
    else state.attempts--;
    const sequence = ++state.sequence, owner = state.owner;
    const owned = () => sequence === state.sequence && owner === identity();
    const live = () => owned() && candidate() === origin;
    state.phase = 'checking'; state.report = null; S.panelUrlError = ''; R();
    try {
      const report = await request('/api/admin/panel-url/check', { public_url: origin });
      if (!live()) return;
      if (!validReport(report, origin)) throw Error('Invalid domain check response. Update the panel backend and retry.');
      state.report = report;
      const networkReady = ['https','panel'].every(id => report.checks.some(c => c.id === id && ['passed','warning'].includes(c.status)));
      if (report.ready) { state.attempts = 0; state.message = 'Checks passed. Review the address, then apply the change below.'; }
      else if (networkReady) { state.attempts = 0; state.message = 'The address is reachable. Resolve the passkey requirements below before applying it.'; }
      else state.message = automatic && state.attempts > 0 ? 'Waiting for HTTPS to become ready. Rechecking shortly; your saved address is unchanged.' : 'Resolve the checks below, then check again. Your saved address is unchanged.';
    } catch (error) { if (live()) { S.panelUrlError = failure(error); state.attempts = 0; } }
    finally { if (owned()) { if (!live()) { state.report = null; state.attempts = 0; } state.phase = ''; R(); sync(); if (automatic) scheduleCheck(); } }
  }
  async function apply(event) {
    event?.preventDefault(); ensure();
    if (!canApply() || S.panelUrlLoading || !S.adminPanelUrl) return;
    const origin = candidate(), ticket = state.report.check_token, owner = state.owner, sequence = ++state.sequence;
    const live = () => sequence === state.sequence && owner === identity();
    clearTimeout(state.timer); state.attempts = 0; state.phase = 'applying'; S.panelUrlError = ''; R();
    try {
      const saved = await request('/api/admin/panel-url/apply', { public_url: origin, check_token: ticket });
      if (!live()) return;
      if (saved?.public_url !== origin || saved?.source !== 'settings') throw Error('Unexpected save response. Refresh the saved address before trying again.');
      S.adminPanelUrl = saved; S.panelUrlDraft = null; state.report = null;
      state.message = 'Verified address saved. Email links and passkeys now use it. This page stays open at your current address.';
      toast('Verified panel address saved', 'ok');
    } catch (error) {
      if (live()) { S.panelUrlError = failure(error); state.report = validReport(error?.data?.check, origin) ? error.data.check : null; }
    } finally { if (live()) { state.phase = ''; R(); sync(); } }
  }
  async function setupHTTPS() {
    ensure(); if (!canSetup()) return;
    const emailInput = document.querySelector('#domain-acme-email'), passwordInput = document.querySelector('#domain-admin-password'), confirm = document.querySelector('#domain-https-confirm');
    const email = emailInput?.value.trim() || '', password = passwordInput?.value || '';
    if (!email || !emailInput.checkValidity() || !password || !confirm?.checked) { S.panelUrlError = 'Enter a valid certificate email, your administrator password, and confirm domain ownership and the agreement.'; const error = document.querySelector('#adm-panel-url-error'); if (error) { error.textContent = S.panelUrlError; error.hidden = false; } return; }
    passwordInput.value = '';
    const origin = candidate(), ticket = state.report.check_token, owner = state.owner, sequence = ++state.sequence;
    const owned = () => sequence === state.sequence && owner === identity();
    const live = () => owned() && candidate() === origin;
    state.email = email; state.phase = 'setup'; S.panelUrlError = ''; R();
    try {
      const result = await request('/api/admin/panel-url/https', { public_url: origin, check_token: ticket, email, password, confirm: true }, 60000);
      if (!live()) return;
      if (result?.needs_verification !== true) throw Error('Unexpected HTTPS setup response. Check the domain before attempting setup again.');
      state.report = null; state.prepared = origin; state.attempts = 6;
      state.message = `${result.message || 'HTTPS setup requested.'} Certificate issuance can take a little time. The saved address is unchanged; checks will run automatically.`;
    } catch (error) { if (live()) { state.report = null; state.attempts = 0; S.panelUrlError = failure(error); } }
    finally { if (owned()) { if (!live()) { state.report = null; state.attempts = 0; } state.phase = ''; R(); sync(); scheduleCheck(); } }
  }
  window.NPDomain = { render, renderSettingsLink, protocolControl, setProtocol, check, apply, setupHTTPS, canApply, busy, invalidate, cancelCheck, reset, sync, setEmail(value) { state.email = value; } };
  setInterval(sync, 1000);
})();
