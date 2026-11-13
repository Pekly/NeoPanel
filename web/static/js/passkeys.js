// WebAuthn UI. Keys stay with the authenticator; only public responses reach the panel.
const PasskeyUI = { items: [], owner: null, sequence: 0, loading: false, available: false, reason: '', busy: false, controller: null };
function passkeySupportIssue() {
  if (!window.isSecureContext) return 'Passkeys require HTTPS or localhost. You can still sign in with your password.';
  if (typeof PublicKeyCredential === 'undefined' || !navigator.credentials?.create || !navigator.credentials?.get) return 'This browser does not support passkeys. Use a supported browser or your password.';
  return '';
}
function passkeyBytes(value) {
  const base64 = String(value).replace(/-/g, '+').replace(/_/g, '/');
  const binary = atob(base64.padEnd(Math.ceil(base64.length / 4) * 4, '='));
  return Uint8Array.from(binary, character => character.charCodeAt(0)).buffer;
}
function passkeyBase64(value) {
  const bytes = ArrayBuffer.isView(value) ? new Uint8Array(value.buffer, value.byteOffset, value.byteLength) : new Uint8Array(value);
  let binary = '';
  for (const byte of bytes) binary += String.fromCharCode(byte);
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}
function passkeyOptions(payload, creation) {
  const raw = payload.options?.publicKey || payload.options || payload.publicKey || payload;
  if (!raw || typeof raw.challenge !== 'string' || !raw.challenge || (creation && typeof raw.user?.id !== 'string')) throw new Error('The panel returned invalid passkey options. Refresh and try again.');
  const options = JSON.parse(JSON.stringify(raw));
  options.challenge = passkeyBytes(options.challenge);
  if (creation && options.user) options.user.id = passkeyBytes(options.user.id);
  for (const key of ['allowCredentials', 'excludeCredentials']) {
    if (Array.isArray(options[key])) options[key] = options[key].map(credential => ({ ...credential, id: passkeyBytes(credential.id) }));
  }
  return options;
}
function passkeyCredentialJSON(credential) {
  if (!credential) throw new Error('No passkey response was returned.');
  const response = credential.response;
  const result = { id: credential.id, rawId: passkeyBase64(credential.rawId), type: credential.type, response: { clientDataJSON: passkeyBase64(response.clientDataJSON) }, clientExtensionResults: credential.getClientExtensionResults?.() || {} };
  if (credential.authenticatorAttachment) result.authenticatorAttachment = credential.authenticatorAttachment;
  if (response.attestationObject) {
    result.response.attestationObject = passkeyBase64(response.attestationObject);
    if (response.getTransports) result.response.transports = response.getTransports();
  } else {
    result.response.authenticatorData = passkeyBase64(response.authenticatorData);
    result.response.signature = passkeyBase64(response.signature);
    result.response.userHandle = response.userHandle ? passkeyBase64(response.userHandle) : null;
  }
  return result;
}
function passkeyErrorMessage(error) {
  if (error?.name === 'NotAllowedError' || error?.name === 'AbortError') return 'Passkey request cancelled or timed out. You can try again or use your password.';
  if (error?.name === 'InvalidStateError') return 'This authenticator already has a passkey for your account. Use it to sign in, or choose another device.';
  if (error?.name === 'SecurityError') return 'Passkeys are not available for this site address. Check HTTPS and the panel’s public URL.';
  return error?.message || 'The passkey request failed. Please try again.';
}
function abortPasskeyOperation() { PasskeyUI.controller?.abort(); }
function closePasskeyDialog() { abortPasskeyOperation(); closeModal(); }
async function passkeyRequest(url, body, authenticated = true, signal) {
  const headers = { 'Content-Type': 'application/json' };
  if (authenticated && S.token) headers.Authorization = `Bearer ${S.token}`;
  const response = await fetch(url, { method: 'POST', headers, body: JSON.stringify(body), cache: 'no-store', signal });
  const result = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(result.error || `Passkey request failed (HTTP ${response.status}).`);
  return result;
}
async function loadPasskeys(skipRender = false) {
  const owner = S.user?.id;
  if (!owner) return;
  const sequence = ++PasskeyUI.sequence;
  PasskeyUI.loading = true; PasskeyUI.reason = '';
  if (PasskeyUI.owner !== owner) { PasskeyUI.owner = owner; PasskeyUI.items = []; PasskeyUI.available = false; }
  try {
    const result = await api('/api/account/passkeys', { suppressToast: true });
    if (S.user?.id !== owner || sequence !== PasskeyUI.sequence) return;
    PasskeyUI.items = Array.isArray(result.passkeys) ? result.passkeys : [];
    PasskeyUI.available = result.available === true;
    PasskeyUI.reason = result.reason || '';
  } catch (error) {
    if (S.user?.id === owner && sequence === PasskeyUI.sequence) { PasskeyUI.available = false; PasskeyUI.reason = error.message; }
  } finally {
    if (S.user?.id === owner && sequence === PasskeyUI.sequence) { PasskeyUI.loading = false; if (!skipRender && S.view === 'acc' && S.acct.tab === 1) R(); }
  }
}
function renderPasskeySettings() {
  const support = passkeySupportIssue();
  const loading = PasskeyUI.loading || PasskeyUI.owner !== S.user?.id;
  const disabled = support || !PasskeyUI.available || PasskeyUI.busy || loading;
  return `<section class="card np-passkey-card" style="margin-top:14px"><div class="np-team-heading"><div><h2>Passkeys</h2><p class="mu">Sign in with your device’s screen lock or a security key. Password sign-in remains available, and existing two-factor verification still applies.</p></div><div class="row"><button type="button" class="btn sm" onclick="loadPasskeys()">Refresh</button><button type="button" class="btn sm pr" id="add-passkey-btn" onclick="openPasskeyRegistration()" ${disabled ? 'disabled' : ''}>${ic('key', 14)} Add passkey</button></div></div>
    ${support || PasskeyUI.reason ? `<p class="np-team-notice" role="status">${esc(support || PasskeyUI.reason)}</p>` : ''}
    ${loading ? loaderHtml('Loading passkeys…') : !PasskeyUI.items.length ? '<div class="np-empty-state"><b>No passkeys yet</b><p class="mu">Add a passkey on a trusted device. Your private key never leaves your authenticator.</p></div>' : `<div class="np-table-scroll"><table class="np-passkey-table"><thead><tr><th>Name</th><th>Added</th><th>Last used</th><th>Actions</th></tr></thead><tbody>${PasskeyUI.items.map(key => `<tr><td><b>${esc(key.label || 'Passkey')}</b></td><td class="mu">${esc(formatFileDate(key.created_at))}</td><td class="mu">${key.last_used_at ? esc(formatFileDate(key.last_used_at)) : 'Not used yet'}</td><td><button class="btn sm dg" type="button" data-passkey-id="${esc(String(key.id))}" onclick="openRemovePasskey(this.dataset.passkeyId)" ${PasskeyUI.busy ? 'disabled' : ''}>Remove</button></td></tr>`).join('')}</tbody></table></div>`}</section>`;
}
function openPasskeyRegistration() {
  if (PasskeyUI.busy || !PasskeyUI.available || passkeySupportIssue()) return;
  document.querySelector('#modal-container').innerHTML = `<div class="pal" onclick="if(event.target===this)closePasskeyDialog()"><div class="card pbox np-passkey-modal" role="dialog" aria-modal="true" aria-labelledby="passkey-register-title"><div class="np-team-heading"><h2 id="passkey-register-title">Add a passkey</h2><button class="btn sm" type="button" onclick="closePasskeyDialog()" aria-label="Close">×</button></div><p class="mu">Confirm your password, then follow your browser’s instructions to save a passkey on this device or a security key.</p><form id="passkey-register-form" onsubmit="registerPasskey(event)"><label class="f"><span>Name</span><input id="passkey-label" required maxlength="64" autocomplete="off" placeholder="e.g. Work laptop"></label><label class="f"><span>Current password</span><input id="passkey-password" type="password" required autocomplete="current-password"></label><p class="upload-error" id="passkey-register-error" role="alert"></p><div class="row" style="justify-content:flex-end"><button class="btn" type="button" onclick="closePasskeyDialog()">Cancel</button><button id="passkey-register-submit" class="btn pr" type="submit">Continue</button></div></form></div></div>`;
  document.querySelector('#passkey-label').focus();
}
async function registerPasskey(event) {
  event.preventDefault();
  const form = document.querySelector('#passkey-register-form');
  const owner = S.user?.id;
  if (!owner || PasskeyUI.busy || !form?.reportValidity()) return;
  const error = document.querySelector('#passkey-register-error'), button = document.querySelector('#passkey-register-submit');
  const body = { label: document.querySelector('#passkey-label').value.trim(), password: document.querySelector('#passkey-password').value };
  document.querySelector('#passkey-password').value = '';
  const controller = new AbortController(); PasskeyUI.controller = controller; PasskeyUI.busy = true;
  setBtnLoading(button, true, 'Waiting for your device…'); error.textContent = '';
  let success = false;
  try {
    const begin = await passkeyRequest('/api/account/passkeys/register/begin', body, true, controller.signal);
    const credential = await navigator.credentials.create({ publicKey: passkeyOptions(begin, true), signal: controller.signal });
    if (S.user?.id !== owner || document.querySelector('#passkey-register-form') !== form || S.view !== 'acc' || S.acct.tab !== 1) throw new DOMException('Registration cancelled', 'AbortError');
    const result = await passkeyRequest('/api/account/passkeys/register/finish', { flow_id: begin.flow_id, credential: passkeyCredentialJSON(credential) }, true, controller.signal);
    success = true; PasskeyUI.controller = null; closeModal(); toast(result.message || 'Passkey added.', 'ok');
  } catch (failure) { if (error.isConnected) error.textContent = passkeyErrorMessage(failure); }
  finally {
    if (PasskeyUI.controller === controller) PasskeyUI.controller = null;
    PasskeyUI.busy = false;
    if (button.isConnected) setBtnLoading(button, false);
    if (success && S.user?.id === owner) await loadPasskeys();
  }
}
function openRemovePasskey(id) {
  const key = PasskeyUI.items.find(item => String(item.id) === String(id));
  if (!key || PasskeyUI.busy) return;
  document.querySelector('#modal-container').innerHTML = `<div class="pal" onclick="if(event.target===this&&!PasskeyUI.busy)closeModal()"><div class="card pbox np-passkey-modal" role="dialog" aria-modal="true" aria-labelledby="passkey-remove-title"><h2 id="passkey-remove-title">Remove passkey?</h2><p class="mu">Remove <b>${esc(key.label || 'this passkey')}</b>? You can still sign in with your password or another passkey.</p><form id="passkey-remove-form" data-passkey-id="${esc(String(key.id))}" onsubmit="removePasskey(event)"><label class="f"><span>Current password</span><input id="passkey-remove-password" type="password" required autocomplete="current-password"></label><p id="passkey-remove-error" class="upload-error" role="alert"></p><div class="row" style="justify-content:flex-end"><button class="btn" type="button" onclick="closeModal()">Cancel</button><button class="btn dg" id="passkey-remove-submit" type="submit">Remove passkey</button></div></form></div></div>`;
  document.querySelector('#passkey-remove-password').focus();
}
async function removePasskey(event) {
  event.preventDefault();
  const form = document.querySelector('#passkey-remove-form'), owner = S.user?.id;
  if (!owner || PasskeyUI.busy || !form?.reportValidity()) return;
  const button = document.querySelector('#passkey-remove-submit'), error = document.querySelector('#passkey-remove-error');
  const password = document.querySelector('#passkey-remove-password').value;
  document.querySelector('#passkey-remove-password').value = '';
  PasskeyUI.busy = true; setBtnLoading(button, true, 'Removing…');
  let success = false;
  try {
    const result = await api(`/api/account/passkeys/${encodeURIComponent(form.dataset.passkeyId)}`, { method: 'DELETE', body: JSON.stringify({ password }), suppressToast: true });
    success = true; if (document.querySelector('#passkey-remove-form') === form) closeModal(); toast(result.message || 'Passkey removed.', 'ok');
  } catch (failure) { if (error.isConnected) error.textContent = passkeyErrorMessage(failure); }
  finally { PasskeyUI.busy = false; if (button.isConnected) setBtnLoading(button, false); if (success && S.user?.id === owner) await loadPasskeys(); }
}
function renderPasskeySignIn() {
  const issue = passkeySupportIssue();
  return `<div class="np-auth-divider"><span>or</span></div><button type="button" class="btn np-passkey-signin" id="passkey-signin-btn" onclick="signInWithPasskey()" ${issue ? 'disabled' : ''}>${ic('key', 16)} Sign in with a passkey</button>${issue ? `<p class="lsm np-passkey-support">${esc(issue)}</p>` : ''}`;
}
async function completePasskeySignIn(result) {
  if (result.two_factor_required && result.temp_token) { resetTurnstileWidget(); show2FALoginScreen(result.temp_token); return; }
  if (!result.token) throw new Error(result.error || 'Passkey sign-in was not completed.');
  PasskeyUI.busy = false; PasskeyUI.controller = null;
  S.token = result.token; S.turnstileToken = ''; S.maintAdminSignInOpen = false;
  localStorage.setItem('neopanel_token', result.token);
  document.querySelector('#login')?.remove(); toast('Signed in successfully', 'ok');
  await checkAuth();
}
async function signInWithPasskey() {
  const error = document.querySelector('#lerr'), button = document.querySelector('#passkey-signin-btn');
  if (PasskeyUI.busy) return;
  const issue = passkeySupportIssue(); if (issue) { if (error) error.textContent = issue; return; }
  if (S.turnstileEnabled && !S.turnstileToken) { if (error) error.textContent = 'Please complete the security check first.'; return; }
  const controller = new AbortController(); PasskeyUI.controller = controller; PasskeyUI.busy = true;
  if (error) error.textContent = ''; setBtnLoading(button, true, 'Waiting for your passkey…');
  let success = false;
  try {
    const begin = await passkeyRequest('/api/auth/passkeys/login/begin', { turnstile_token: S.turnstileToken || '' }, false, controller.signal);
    const credential = await navigator.credentials.get({ publicKey: passkeyOptions(begin, false), signal: controller.signal });
    const result = await passkeyRequest('/api/auth/passkeys/login/finish', { flow_id: begin.flow_id, credential: passkeyCredentialJSON(credential) }, false, controller.signal);
    await completePasskeySignIn(result); success = true;
  } catch (failure) { if (error?.isConnected) error.textContent = passkeyErrorMessage(failure); }
  finally {
    if (PasskeyUI.controller === controller) PasskeyUI.controller = null;
    PasskeyUI.busy = false;
    if (button?.isConnected) setBtnLoading(button, false);
    if (!success) resetTurnstileWidget();
  }
}
