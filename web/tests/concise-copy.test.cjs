const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = name => fs.readFileSync(path.join(__dirname, `../static/js/${name}.js`), 'utf8');
const plain = html => html.replace(/<[^>]*>/g, ' ').replace(/\s+/g, ' ').trim();
function concise(html, maxWords) {
  const text = plain(html);
  assert.ok(text.split(/\s+/).length <= maxWords, text);
  return text;
}
function context(name, overrides = {}) {
  const c = vm.createContext({
    S: { token: 'session', user: { id: 1 }, view: 'team' },
    esc: value => String(value ?? ''), ic: () => '',
    curServer: () => ({ id: 'one', name: 'One' }),
    sameQolServer: (a, b) => a?.id === b?.id,
    ...overrides
  });
  vm.runInContext(source(name), c);
  return c;
}

test('restore confirmation keeps file loss, stopped-server and disk warnings', async () => {
  let confirmation;
  const c = context('automation', {
    api: async () => ({ status: 'stopped' }), srvApiPath: () => '/server',
    uiConfirm: async options => { confirmation = options; return false; }
  });
  await c.restoreBackup('backup.zip');
  const text = concise(confirmation.message, 45);
  assert.equal(confirmation.isDanger, true);
  assert.match(text, /replaces current server files/i);
  assert.match(text, /removes files absent from the backup/i);
  assert.match(text, /server stopped/i);
  assert.match(text, /safety copy.*disk space/i);
  vm.runInContext("AutomationUI.lastRestore = { server: { id: 'one' }, safetyCopy: 'safety-copy' };", c);
  const notice = plain(c.renderRestoreNotice());
  assert.match(notice, /safety-copy/);
  assert.match(notice, /disk space/i);
  assert.match(notice, /check.*before deleting/i);
});

test('Trash copy retains expiry, no-overwrite and permanent-deletion warnings', async () => {
  let confirmation;
  const c = context('file-trash', {
    hasServerPermission: () => true,
    document: { getElementById: () => null },
    uiConfirm: async options => { confirmation = options; return false; }
  });
  const text = plain(c.renderFileTrash());
  assert.match(text, /up to 24 hours.*automatic deletion/i);
  assert.match(text, /original path.*never overwrites/i);
  await c.confirmFileDeletion({ title: 'Delete file?', description: '<p>Delete file?</p>' });
  assert.match(plain(confirmation.message), /24 hours.*automatic deletion/i);
  assert.match(plain(confirmation.message), /never overwritten/i);
  assert.match(plain(confirmation.message), /permanent deletion cannot be undone/i);
});

test('passkeys explain sign-in without removing password or 2FA requirements', () => {
  const c = context('passkeys', {
    window: { isSecureContext: true }, PublicKeyCredential: function () {},
    navigator: { credentials: { create() {}, get() {} } }, loaderHtml: () => ''
  });
  const html = c.renderPasskeySettings();
  const intro = concise(html.match(/<p class="mu">(.*?)<\/p>/)[1], 25);
  assert.match(intro, /screen lock.*security key/i);
  assert.match(intro, /password login.*available/i);
  assert.match(intro, /password login.*keeps your 2FA settings/i);
  assert.doesNotMatch(intro, /without.*two-factor/i);
});

test('team invitations retain acceptance and existing-account protections', () => {
  const root = { innerHTML: '', querySelector: () => null };
  const c = context('team', {
    S: { user: { id: 1, username: 'member', email: 'member@example.test' },
      teamState: { permission: { is_owner: true }, invitations: [], smtp_configured: true }, serverSubusers: [] },
    srvh: () => '', renderSubusersListHtml: () => '', SUBUSER_PERMS: [],
    brandLogoHtml: () => '', document: { querySelector: key => key === '#invite-screen' ? root : null }
  });
  assert.match(plain(c.renderTeamPage()), /no server access until accepted/i);
  for (const state of [
    { account_exists: true, email: 'other@example.test' },
    { account_exists: true, email: 'member@example.test' },
    { account_reserved: true, email: 'member@example.test' }
  ]) {
    c.info = state;
    vm.runInContext('InvitationUI.info = info;', c);
    c.renderInvitationScreen();
    assert.match(plain(root.innerHTML), /password.*unchanged/i);
    assert.doesNotMatch(root.innerHTML, /id="invite-register-password"/);
  }
});

test('preview and recovery helpers give a next step and retain important limits', () => {
  const preview = source('file-preview');
  assert.match(preview, /Browse without extracting\. Preview text files read-only\./);
  assert.match(preview, /Text editing could corrupt this file\. Download/);
  assert.match(preview, /unsupported by your browser\. Try downloading/);
  const reset = source('password-reset');
  assert.match(reset, /If the account exists.*reset link and code.*expire in one hour/);
  assert.match(reset, /64-character code/);
  assert.match(reset, /72 UTF-8 bytes/);
  const progress = source('operation-progress');
  assert.match(progress, /percentage unavailable/);
  assert.equal((progress.match(/Closing the dialog won’t cancel/g) || []).length, 2);
});
