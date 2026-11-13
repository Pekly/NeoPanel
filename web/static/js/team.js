// Team management and invitation registration. Loaded before app.js.
const TeamUI = { sequence: 0, inviteServer: null, memberServer: null, sending: false, inviteAction: null };
const InvitationUI = { token: '', info: null, error: '', busy: false, accepted: false };

function canManageTeam(permission = S.teamState?.permission) {
  return Boolean(permission && (permission.is_admin || permission.is_owner || permission.permissions?.includes('subusers.manage')));
}
function teamInvitationPath(srv) {
  return '/api/team/invitations?' + new URLSearchParams({ server_id: srv.id, node_id: srv.node_id || 'local' });
}
async function loadTeamPage(skipRender = false) {
  const srv = curServer();
  if (!srv) return;
  const sequence = ++TeamUI.sequence;
  S.teamLoading = true;
  if (!skipRender && S.view === 'team') R();
  try {
    const permission = await api(srvApiPath(srv, '/permissions'), { suppressToast: true });
    if (sequence !== TeamUI.sequence || !sameQolServer(srv, curServer())) return;
    const state = { permission, invitations: [], smtp_configured: false, error: '' };
    if (canManageTeam(permission)) {
      const [members, invites] = await Promise.all([
        api(srvApiPath(srv, '/subusers'), { suppressToast: true }),
        api(teamInvitationPath(srv), { suppressToast: true })
      ]);
      if (sequence !== TeamUI.sequence || !sameQolServer(srv, curServer())) return;
      S.serverSubusers = Array.isArray(members) ? members : (members.subusers || []);
      state.invitations = invites.invitations || [];
      state.smtp_configured = Boolean(invites.smtp_configured);
    } else S.serverSubusers = [];
    S.teamState = state;
  } catch (err) {
    if (sequence === TeamUI.sequence && sameQolServer(srv, curServer())) S.teamState = { error: err.message, invitations: [] };
  } finally {
    if (sequence === TeamUI.sequence && sameQolServer(srv, curServer())) {
      S.teamLoading = false;
      if (!skipRender && S.view === 'team') R();
    }
  }
}
function renderTeamPage() {
  const srv = curServer(), state = S.teamState;
  const header = `${srvh()}<div class="np-section-heading"><div><h2>Team</h2><p>Manage server access and invite teammates without sharing your account.</p></div><button class="btn sm" onclick="loadTeamPage()">${ic('refresh', 13)} Refresh</button></div>`;
  if (S.teamLoading || !state) return header + `<div class="card">${loaderHtml('Loading team access…')}</div>`;
  if (state.error) return header + `<div class="card" role="alert"><b>Could not load team access</b><p class="mu">${esc(state.error)}</p><button class="btn sm" onclick="loadTeamPage()">Retry</button></div>`;
  if (!canManageTeam()) return header + `<div class="card"><h3>Team management is restricted</h3><p class="mu">You need the Manage Team permission to add members or send invitations. Your other server permissions are unchanged.</p></div>`;
  const invitations = state.invitations || [];
  return header + `<div class="np-team-stack">
    <section class="card"><div class="np-team-heading"><div><h3>Members</h3><p class="mu">${S.serverSubusers?.length || 0} team member${S.serverSubusers?.length === 1 ? '' : 's'} · Owners and panel administrators keep their existing access.</p></div><div class="row"><button class="btn sm" onclick="openSubuserModal()">${ic('user', 13)} Add existing user</button><button class="btn sm pr" id="team-invite-btn" onclick="openTeamInviteModal()">${ic('mail', 13)} Invite by email</button></div></div><div id="subusers-list-wrap">${renderSubusersListHtml(srv)}</div></section>
    <section class="card"><div class="np-team-heading"><div><h3>Email invitations</h3><p class="mu">Recipients register through a single-use link. Pending invitations do not grant server access.</p></div></div>
    ${!state.smtp_configured ? '<p class="np-team-notice" role="status">Email invitations require SMTP to be enabled in Admin → Email &amp; SMTP. Existing users can still be added directly.</p>' : ''}
    ${!invitations.length ? '<div class="np-empty-state"><b>No invitations yet</b><p class="mu">Enter a username and email to invite a new teammate.</p></div>' : `<div class="np-table-scroll"><table class="np-team-invitations"><thead><tr><th>Recipient</th><th>Permissions</th><th>Status</th><th>Expires</th><th>Actions</th></tr></thead><tbody>${invitations.map(invite => {
      const terminal = ['accepted', 'revoked'].includes(invite.status);
      const id = esc(String(invite.id));
      return `<tr><td><b>${esc(invite.username)}</b><div class="mu pv">${esc(invite.email)}</div></td><td>${invite.permissions?.length || 0} permission${invite.permissions?.length === 1 ? '' : 's'}</td><td><span class="chip">${esc(invitationStatusLabel(invite))}</span></td><td class="mu">${esc(formatInvitationExpiry(invite.expires_at))}</td><td><div class="row">${terminal ? '—' : `<button class="btn sm" data-invite-id="${id}" onclick="changeTeamInvitation(this.dataset.inviteId, 'resend')" ${TeamUI.inviteAction ? 'disabled' : ''}>Resend</button><button class="btn sm dg" data-invite-id="${id}" onclick="changeTeamInvitation(this.dataset.inviteId, 'revoke')" ${TeamUI.inviteAction ? 'disabled' : ''}>Revoke</button>`}</div></td></tr>`;
    }).join('')}</tbody></table></div>`}</section></div>`;
}
function invitationStatusLabel(invite) {
  if (['accepted', 'revoked'].includes(invite.status)) return invite.status;
  if (invite.expires_at && new Date(invite.expires_at) <= new Date()) return 'Expired';
  if (invite.delivery_status === 'failed') return 'Email delivery failed';
  if (invite.delivery_status === 'sending') return 'Sending email';
  return String(invite.status || 'pending').replaceAll('_', ' ');
}
function teamPermissionChecklist(selected = ['console.view']) {
  const checked = new Set(selected);
  return `<div class="subuser-perm-grid">${SUBUSER_PERMS.map(([key, label, description]) => `<label class="subuser-perm-item"><input type="checkbox" name="permissions" value="${esc(key)}" ${checked.has(key) ? 'checked' : ''}><span><b>${esc(label)}</b><span class="mu">${esc(description)}</span></span></label>`).join('')}</div>`;
}
function openTeamInviteModal() {
  if (!canManageTeam()) return;
  if (!S.teamState.smtp_configured) { uiAlert({ title: 'Email is not configured', message: 'Ask a panel administrator to enable SMTP under Admin → Email &amp; SMTP, then refresh this page.' }); return; }
  TeamUI.inviteServer = { ...curServer() };
  const container = document.querySelector('#modal-container');
  container.innerHTML = `<div class="pal" onclick="if(event.target===this&&!TeamUI.sending)closeModal()"><div class="card pbox np-team-modal" role="dialog" aria-modal="true" aria-labelledby="team-invite-title"><div class="np-team-heading"><h2 id="team-invite-title">Invite a teammate</h2><button class="btn sm" onclick="closeModal()" aria-label="Close invitation">×</button></div><p class="mu">Invite someone to <b>${esc(TeamUI.inviteServer.name)}</b>. New users register through the email link; existing users sign in with their current account. Use the existing account’s username and email if they already have one. People who already have server access cannot be invited again.</p>
    <form id="team-invite-form" onsubmit="submitTeamInvitation(event)"><div class="grid g2"><label class="f"><span>Requested username</span><input id="team-invite-username" required minlength="3" maxlength="32" pattern="[A-Za-z0-9][A-Za-z0-9_.\\-]{2,31}" autocomplete="off" placeholder="teammate"></label><label class="f"><span>Email address</span><input id="team-invite-email" type="email" required maxlength="254" autocomplete="off" placeholder="teammate@example.com"></label></div><h3>Server permissions</h3><p class="mu">Grant only what this person needs. Permissions apply after acceptance.</p>${teamPermissionChecklist()}<p id="team-invite-error" class="upload-error" role="alert"></p><div class="row" style="justify-content:flex-end;margin-top:18px"><button class="btn" type="button" onclick="closeModal()">Cancel</button><button id="team-invite-submit" class="btn pr" type="submit">Send invitation</button></div></form></div></div>`;
  document.querySelector('#team-invite-username').focus();
}
async function submitTeamInvitation(event) {
  event.preventDefault();
  const srv = TeamUI.inviteServer;
  if (TeamUI.sending || !srv || !sameQolServer(srv, curServer())) return;
  const form = document.querySelector('#team-invite-form');
  if (!form?.reportValidity()) return;
  const body = { server_id: srv.id, node_id: srv.node_id || 'local', username: document.querySelector('#team-invite-username').value.trim(), email: document.querySelector('#team-invite-email').value.trim(), permissions: [...form.querySelectorAll('[name=permissions]:checked')].map(el => el.value) };
  const conflict = teamInvitationConflict(body.username, body.email, srv);
  if (conflict) { document.querySelector('#team-invite-error').textContent = conflict; return; }
  TeamUI.sending = true;
  const button = document.querySelector('#team-invite-submit'); setBtnLoading(button, true, 'Sending…');
  try {
    const result = await api('/api/team/invitations', { method: 'POST', body: JSON.stringify(body), suppressToast: true });
    closeModal(); toast(result.message || 'Invitation email sent.', 'ok');
    if (sameQolServer(srv, curServer())) await loadTeamPage();
  } catch (err) {
    const error = document.querySelector('#team-invite-error'); if (error) error.textContent = err.message;
    if (sameQolServer(srv, curServer())) await loadTeamPage();
  } finally { TeamUI.sending = false; if (button?.isConnected) setBtnLoading(button, false); }
}
async function changeTeamInvitation(id, action) {
  const srv = curServer();
  if (!srv || !canManageTeam() || TeamUI.inviteAction) return;
  if (action === 'revoke' && !await uiConfirm({ title: 'Revoke invitation?', message: 'The registration link will stop working. Existing accounts and other server access will not be deleted.', confirmText: 'Revoke invitation', isDanger: true })) return;
  TeamUI.inviteAction = id;
  try {
    const url = `/api/team/invitations/${encodeURIComponent(id)}` + (action === 'resend' ? '/resend' : '');
    const result = await api(url, { method: action === 'resend' ? 'POST' : 'DELETE', suppressToast: true });
    toast(result.message || (action === 'resend' ? 'A new invitation email was sent. The old link is invalid.' : 'Invitation revoked.'), 'ok');
  } catch (err) { toast(err.message, 'bad'); }
  finally { TeamUI.inviteAction = null; if (sameQolServer(srv, curServer())) await loadTeamPage(); }
}

function pendingInvitationToken() {
  const hash = new URLSearchParams(location.hash.slice(1));
  const incoming = hash.get('invite-token');
  if (incoming) {
    InvitationUI.token = incoming;
    try { sessionStorage.setItem('neopanel_pending_invite', incoming); } catch (_) {}
    history.replaceState(history.state, '', location.pathname + location.search);
  }
  if (InvitationUI.token) return InvitationUI.token;
  try { return sessionStorage.getItem('neopanel_pending_invite') || ''; } catch (_) { return ''; }
}
async function invitationRequest(action, body) {
  const headers = { 'Content-Type': 'application/json' };
  if (S.token) headers.Authorization = `Bearer ${S.token}`;
  const response = await fetch(`/api/auth/invitations/${action}`, { method: 'POST', headers, body: JSON.stringify(body), cache: 'no-store' });
  const result = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(result.error || `Invitation request failed (HTTP ${response.status}).`);
  return result;
}
async function initializeInvitationAuth() {
  const response = await fetch('/api/auth/status', { cache: 'no-store' });
  if (!response.ok) throw new Error('Could not load sign-in settings. Refresh and try the invitation again.');
  const status = await response.json();
  S.turnstileEnabled = Boolean(status.turnstile_enabled && status.turnstile_site_key);
  S.turnstileSiteKey = S.turnstileEnabled ? status.turnstile_site_key : '';
  S.maintPublicInfo = { enabled: Boolean(status.maintenance_enabled), title: status.maintenance_title || '', message: status.maintenance_message || '', eta: status.maintenance_eta || '' };
}
async function openPendingInvitation() {
  const token = pendingInvitationToken();
  if (!token) return false;
  InvitationUI.token = token; InvitationUI.error = ''; InvitationUI.accepted = false;
  await loadPublicBranding().catch(() => {});
  try { await initializeInvitationAuth(); }
  catch (err) { InvitationUI.info = null; InvitationUI.error = err.message; renderInvitationScreen(); return true; }
  if (S.token) {
    try {
      const response = await fetch('/api/auth/me', { headers: { Authorization: `Bearer ${S.token}` } });
      if (response.ok) S.user = await response.json();
      else { S.token = ''; S.user = null; localStorage.removeItem('neopanel_token'); }
    } catch (_) { S.user = null; }
  }
  try { InvitationUI.info = await invitationRequest('inspect', { token }); }
  catch (err) { InvitationUI.info = null; InvitationUI.error = err.message; }
  renderInvitationScreen();
  return true;
}
function renderInvitationScreen() {
  document.querySelector('#login')?.remove();
  let root = document.querySelector('#invite-screen');
  if (!root) { root = document.createElement('div'); root.id = 'invite-screen'; root.className = 'np-invitation-shell'; document.body.appendChild(root); }
  const info = InvitationUI.info;
  const matchingAccount = Boolean(S.user && info?.email && String(S.user.email || '').toLowerCase() === info.email.toLowerCase());
  let content;
  if (InvitationUI.accepted) {
    content = `<h1>You’re ready to join</h1><p class="mu">Your invitation to <b>${esc(info?.server_name || 'the server')}</b> was accepted. Sign in with your account to continue.</p><button class="btn pr" onclick="invitationSignIn()">Continue to sign in</button>`;
  } else if (!info) {
    content = `<h1>Invitation unavailable</h1><p class="upload-error" role="alert">${esc(InvitationUI.error || 'This invitation could not be loaded.')}</p><p class="mu">Ask the server owner for a new invitation if the link has expired or been revoked.</p><button class="btn" onclick="dismissInvitation()">Back to panel</button>`;
  } else {
    content = `<div class="np-eyebrow">Server invitation</div><h1>Join ${esc(info.server_name)}</h1><p class="mu">This invitation is for <b>${esc(info.email)}</b>. Expires ${esc(formatInvitationExpiry(info.expires_at))}.</p><details class="np-invite-permissions"><summary>${info.permissions?.length || 0} server permission${info.permissions?.length === 1 ? '' : 's'}</summary><ul>${(info.permissions || []).map(key => `<li>${esc(SUBUSER_PERMS.find(p => p[0] === key)?.[1] || key)}</li>`).join('')}</ul></details>
      ${info.account_exists && !matchingAccount ? `<p class="np-team-notice">An account already exists for this email. Sign in to accept; this link cannot change its password.${S.user ? ` You are currently signed in as ${esc(S.user.username)}.` : ''}</p><button class="btn pr" onclick="invitationSignIn()">Sign in to accept</button>` : `<form id="invite-accept-form" onsubmit="acceptTeamInvitation(event)">
        ${!info.account_exists && !info.account_reserved ? `<label class="f"><span>Invited username</span><input id="invite-register-username" readonly autocomplete="username" value="${esc(info.username || '')}"></label><label class="f"><span>Email</span><input type="email" value="${esc(info.email)}" readonly autocomplete="email"></label><label class="f"><span>Password (at least 10 characters)</span><input id="invite-register-password" type="password" required minlength="10" maxlength="72" autocomplete="new-password"></label><label class="f"><span>Confirm password</span><input id="invite-register-confirm" type="password" required minlength="10" maxlength="72" autocomplete="new-password"></label>` : `<p class="np-team-notice">${info.account_reserved ? 'Your account was created during an earlier attempt. Retry granting server access; your original password will not change.' : `Accept as <b>${esc(S.user.username)}</b>. Your password and account security settings will not change.`}</p>`}
        <p id="invite-accept-error" class="upload-error" role="alert">${esc(InvitationUI.error)}</p><button class="btn pr" id="invite-accept-submit" type="submit">${info.account_reserved ? 'Retry server access' : info.account_exists ? 'Accept invitation' : 'Create account & accept'}</button></form>`}<button class="btn sm" style="margin-top:16px" onclick="dismissInvitation()">Not now</button>`;
  }
  root.innerHTML = `<main class="card np-invitation-card" aria-label="Server invitation"><div class="brand">${brandLogoHtml(false)}</div>${content}</main>`;
  root.querySelector('input:not([readonly]), button')?.focus();
}
async function acceptTeamInvitation(event) {
  event.preventDefault();
  if (InvitationUI.busy) return;
  const info = InvitationUI.info, form = document.querySelector('#invite-accept-form');
  if (!info || !form?.reportValidity()) return;
  const body = { token: InvitationUI.token };
  const error = document.querySelector('#invite-accept-error');
  if (!info.account_exists && !info.account_reserved) {
    body.username = document.querySelector('#invite-register-username').value.trim();
    body.password = document.querySelector('#invite-register-password').value;
    if (body.password !== document.querySelector('#invite-register-confirm').value) { error.textContent = 'Passwords do not match.'; return; }
    const bytes = new TextEncoder().encode(body.password).length;
    if (bytes < 10 || bytes > 72) { error.textContent = 'Password must be 10–72 UTF-8 bytes. Shorten it if it contains many Unicode characters.'; return; }
  }
  InvitationUI.busy = true;
  const button = document.querySelector('#invite-accept-submit'); setBtnLoading(button, true, 'Accepting…');
  try {
    const result = await invitationRequest('accept', body);
    InvitationUI.info = { ...info, ...result }; InvitationUI.accepted = true; InvitationUI.token = ''; InvitationUI.error = '';
    try { sessionStorage.removeItem('neopanel_pending_invite'); } catch (_) {}
    renderInvitationScreen();
  } catch (err) {
    InvitationUI.error = err.message; error.textContent = err.message;
    try {
      const latest = await invitationRequest('inspect', { token: InvitationUI.token });
      if (latest.account_reserved || latest.account_exists !== info.account_exists) { InvitationUI.info = latest; renderInvitationScreen(); }
    } catch (_) { /* Keep the original failure visible, not a second request error. */ }
  }
  finally { InvitationUI.busy = false; if (button?.isConnected) setBtnLoading(button, false); }
}
async function invitationSignIn() {
  document.querySelector('#invite-screen')?.remove();
  if (S.token) {
    await logout();
    if (S.token) { renderInvitationScreen(); return; }
  }
  if (InvitationUI.accepted && InvitationUI.info?.server_id) history.replaceState(null, '', serverRouteUrl(InvitationUI.info.node_id || 'local', InvitationUI.info.server_id, 'con'));
  showLoginScreen();
  const input = document.querySelector('#lu'); if (input) input.value = InvitationUI.info?.email || '';
}
function dismissInvitation() {
  if (InvitationUI.busy) return;
  InvitationUI.token = ''; InvitationUI.info = null; InvitationUI.accepted = false;
  try { sessionStorage.removeItem('neopanel_pending_invite'); } catch (_) {}
  document.querySelector('#invite-screen')?.remove();
  checkAuth();
}

function formatInvitationExpiry(value) {
  if (!value) return '—';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '—';
  return date.toLocaleString(undefined, { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit', timeZoneName: 'short' });
}

function teamInvitationConflict(username, email, server = curServer()) {
  const name = String(username || '').trim().toLowerCase();
  const address = String(email || '').trim().toLowerCase();
  const matches = account => (name && String(account.username || '').toLowerCase() === name) || (address && String(account.email || '').toLowerCase() === address);
  if ((S.serverSubusers || []).some(matches)) return 'This person already has access to this server. Edit their member permissions instead.';
  if (S.user && (S.user.role === 'admin' || server?.owner_id === S.user.id) && matches(S.user)) return 'This account already has access as a server owner or panel administrator.';
  const pending = (S.teamState?.invitations || []).some(invite => {
    if (!matches(invite)) return false;
    if (invite.status === 'granting') return true;
    return invite.status === 'pending' && (!invite.expires_at || new Date(invite.expires_at).getTime() > Date.now());
  });
  return pending ? 'An invitation already exists for this person. Use Resend or Revoke in the invitations list.' : '';
}
