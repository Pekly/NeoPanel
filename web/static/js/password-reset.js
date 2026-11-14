// Public password recovery. Tokens/passwords stay in memory and are never logged or persisted.
const PASSWORD_RESET = { generation: 0, busy: false, controller: null, widget: null, captcha: '', timer: null, focusTimer: null };
function disposePasswordReset() {
  PASSWORD_RESET.generation++;
  PASSWORD_RESET.controller?.abort(); PASSWORD_RESET.controller = null; PASSWORD_RESET.busy = false;
  clearTimeout(PASSWORD_RESET.timer); clearTimeout(PASSWORD_RESET.focusTimer);
  if (PASSWORD_RESET.widget != null && window.turnstile?.remove) {
    try { window.turnstile.remove(PASSWORD_RESET.widget); } catch (_) {}
  }
  PASSWORD_RESET.widget = null; PASSWORD_RESET.captcha = '';
}
function mountPasswordResetCaptcha(generation) {
  if (!S.turnstileEnabled) return;
  let attempts = 0;
  const mount = () => {
    const box = $('#turnstile-reset');
    if (!box || generation !== PASSWORD_RESET.generation) return;
    if (window.turnstile?.render && S.turnstileSiteKey) {
      box.textContent = '';
      try {
        PASSWORD_RESET.widget = window.turnstile.render(box, {
          sitekey: S.turnstileSiteKey, theme: document.documentElement.dataset.theme === 'light' ? 'light' : 'dark',
          callback: token => { if (generation === PASSWORD_RESET.generation) PASSWORD_RESET.captcha = token || ''; },
          'expired-callback': () => { if (generation === PASSWORD_RESET.generation) PASSWORD_RESET.captcha = ''; },
          'error-callback': () => { if (generation === PASSWORD_RESET.generation) PASSWORD_RESET.captcha = ''; }
        });
      } catch (_) {
        box.textContent = 'Security check unavailable. Reopen password reset or contact the panel administrator.';
      }
    } else if (++attempts < 100) {
      box.textContent = 'Loading security check…'; PASSWORD_RESET.timer = setTimeout(mount, 100);
    } else box.textContent = 'Security check unavailable. Check your connection and reopen password reset.';
  };
  mount();
}
function showPasswordResetForm(mode, presetToken = '', notice = '') {
  closeModal(); disposePasswordReset();
  const generation = PASSWORD_RESET.generation, confirming = mode === 'confirm';
  const token = /^[a-f0-9]{64}$/i.test(String(presetToken).trim()) ? String(presetToken).trim().toLowerCase() : '';
  const invalidLink = Boolean(presetToken && !token);
  const container = $('#modal-container');
  if (!container) return;
  _uiModalResolver = () => disposePasswordReset();
  container.innerHTML = `<div class="pal np-password-reset-backdrop" onclick="if(event.target===this)closeModal()">
    <section class="pbox ui-modal-box np-password-reset" role="dialog" aria-modal="true" aria-labelledby="password-reset-title">
      <div class="row sb" style="gap:12px;margin-bottom:16px"><h2 id="password-reset-title" style="margin:0;font-size:19px">${ic('lock', 19)} ${confirming ? 'Choose a new password' : 'Reset your password'}</h2><button type="button" class="btn-ico" aria-label="Close password reset" onclick="closeModal()">×</button></div>
      <p class="mu">${confirming ? 'Paste the 64-character code from your email, then choose your new password.' : 'Enter your username or email. If a matching account exists, we’ll email a reset link and code valid for one hour.'}</p>
      ${notice ? `<p class="np-reset-notice" role="status">${esc(notice)}</p>` : ''}
      <form id="password-reset-form" data-mode="${mode}" onsubmit="event.preventDefault(); ${confirming ? 'submitSetNewPassword()' : 'submitResetPasswordRequest()'}">
        ${confirming ? `<label class="f" for="reset-token-input">Reset code<input id="reset-token-input" class="mono" type="text" value="${esc(token)}" maxlength="64" autocomplete="one-time-code" autocapitalize="none" spellcheck="false" placeholder="64-character code" required></label>
          <label class="f" for="reset-new-password">New password<input id="reset-new-password" type="password" autocomplete="new-password" minlength="6" placeholder="At least 6 characters" required></label>
          <label class="f" for="reset-new-password-confirm">Confirm new password<input id="reset-new-password-confirm" type="password" autocomplete="new-password" minlength="6" placeholder="Repeat new password" required></label>` : `<label class="f" for="reset-identity">Username or email<input id="reset-identity" type="text" maxlength="254" autocomplete="username" placeholder="Username or user@example.com" required></label>
          ${S.turnstileEnabled ? '<div id="turnstile-reset" class="turnstile-wrap"></div>' : ''}`}
        <p id="password-reset-error" class="np-form-error" role="alert">${invalidLink ? 'This reset link is invalid. Paste the full code from your email or request a new one.' : ''}</p>
        <div class="row" style="justify-content:flex-end;gap:10px;margin-top:18px"><button type="button" class="btn" onclick="closeModal()">Cancel</button><button type="submit" id="password-reset-submit" class="btn pr">${confirming ? 'Update password' : 'Send reset link'}</button></div>
      </form>
      <button type="button" class="np-reset-switch" onclick="${confirming ? 'openResetPasswordModal()' : 'openSetNewPasswordModal()'}">${confirming ? 'Request a new reset code' : 'Already have a reset code?'}</button>
    </section></div>`;
  if (!confirming) mountPasswordResetCaptcha(generation);
  PASSWORD_RESET.focusTimer = setTimeout(() => {
    if (generation === PASSWORD_RESET.generation) $(confirming ? (token ? '#reset-new-password' : '#reset-token-input') : '#reset-identity')?.focus();
  }, 40);
}
function openResetPasswordModal() { showPasswordResetForm('request'); }
function openSetNewPasswordModal(presetToken = '', notice = '') { showPasswordResetForm('confirm', presetToken, notice); }
function resetFormBusy(form, value) {
  form.querySelectorAll('input, button[type="submit"]').forEach(el => { el.disabled = value; });
  const button = $('#password-reset-submit');
  if (button) { button.setAttribute('aria-busy', String(value)); button.textContent = value ? 'Please wait…' : form.dataset.mode === 'request' ? 'Send reset link' : 'Update password'; }
}
async function sendPasswordReset(mode, payload) {
  const form = $('#password-reset-form'), errorBox = $('#password-reset-error');
  if (!form || form.dataset.mode !== mode || PASSWORD_RESET.busy) return;
  const generation = PASSWORD_RESET.generation;
  const live = () => generation === PASSWORD_RESET.generation && $('#password-reset-form') === form;
  PASSWORD_RESET.busy = true; resetFormBusy(form, true); if (errorBox) errorBox.textContent = '';
  const controller = new AbortController(); PASSWORD_RESET.controller = controller;
  const timer = setTimeout(() => controller.abort(), 20000);
  try {
    const response = await fetch(`/api/auth/password-reset/${mode === 'request' ? 'request' : 'confirm'}`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload), signal: controller.signal
    });
    const result = await response.json();
    if (!live()) return;
    if (!response.ok || result.error) throw Error(result.error || `Password reset failed (HTTP ${response.status}).`);
    if (typeof result.message !== 'string') throw Error('Invalid response from the panel. Please try again.');
    if (mode === 'request') {
      // Never use a token from this public endpoint. Recovery credentials are email-only.
      openSetNewPasswordModal('', result.message);
    } else {
      closeModal();
      resetServerAccess(); S.token = ''; S.user = null; S.servers = [];
      try { localStorage.removeItem('neopanel_token'); } catch (_) {}
      $('#login')?.remove(); showLoginScreen();
      if (result.username && $('#lu')) $('#lu').value = result.username;
      $('#lp')?.focus(); toast('Password updated. Sign in with your new password.', 'ok');
    }
  } catch (error) {
    if (live() && errorBox) errorBox.textContent = error.name === 'AbortError' ? 'The request timed out. Please try again.' : error.message || 'Could not reach the panel.';
  } finally {
    clearTimeout(timer);
    if (live()) {
      PASSWORD_RESET.busy = false; PASSWORD_RESET.controller = null; resetFormBusy(form, false);
      if (mode === 'request') {
        PASSWORD_RESET.captcha = '';
        if (PASSWORD_RESET.widget != null && window.turnstile?.reset) {
          try { window.turnstile.reset(PASSWORD_RESET.widget); } catch (_) {}
        }
      }
    }
  }
}
async function submitResetPasswordRequest() {
  const identity = $('#reset-identity')?.value.trim(), errorBox = $('#password-reset-error');
  if (!identity) { if (errorBox) errorBox.textContent = 'Enter your username or email address.'; return; }
  if (S.turnstileEnabled && !PASSWORD_RESET.captcha) { if (errorBox) errorBox.textContent = 'Complete the security check first.'; return; }
  return sendPasswordReset('request', { identity, 'cf-turnstile-response': PASSWORD_RESET.captcha });
}
async function submitSetNewPassword() {
  const token = $('#reset-token-input')?.value.trim().toLowerCase() || '';
  const password = $('#reset-new-password')?.value || '', confirm = $('#reset-new-password-confirm')?.value;
  const errorBox = $('#password-reset-error');
  let error = '';
  if (!/^[a-f0-9]{64}$/.test(token)) error = 'Enter the full 64-character reset code from your email.';
  else if (password.length < 6) error = 'Use at least 6 characters for your new password.';
  else if (new TextEncoder().encode(password).length > 72) error = 'Your password must not exceed 72 UTF-8 bytes.';
  else if (password !== confirm) error = 'The passwords do not match.';
  if (error) { if (errorBox) errorBox.textContent = error; return; }
  return sendPasswordReset('confirm', { token, new_password: password });
}
function checkUrlResetToken() {
  const params = new URLSearchParams(window.location.hash.slice(1));
  if (!params.has('reset-token')) return;
  const token = params.get('reset-token') || '';
  params.delete('reset-token');
  const remaining = params.toString();
  history.replaceState(history.state, '', window.location.pathname + window.location.search + (remaining ? '#' + remaining : ''));
  openSetNewPasswordModal(token || 'invalid');
}
