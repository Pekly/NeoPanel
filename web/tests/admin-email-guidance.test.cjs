const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const app = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
function setup() {
  const fields = Object.fromEntries(['host', 'port', 'enc', 'user'].map(key => [`#admin-smtp-${key}`, { value: '' }]));
  const context = vm.createContext({ S: { adminSmtp: {}, user: { email: 'admin@example.com' } }, ic: () => '', esc: value => String(value ?? ''), $: selector => fields[selector] });
  vm.runInContext(app.slice(app.indexOf('function renderAdminEmailTabHtml()'), app.indexOf('async function loadAdminSmtp()')), context);
  return { context, fields, html: context.renderAdminEmailTabHtml() };
}
test('email setup explains saved test configuration and sender verification', () => {
  const { html } = setup();
  assert.match(html, /test uses the saved configuration, not unsaved fields/);
  assert.match(html, /inbox and spam folder/);
  assert.match(html, /SPF\/DKIM/);
  assert.match(html, /changing the panel address does not configure email DNS/);
});
test('inbound Cloudflare routing is not offered as an outgoing SMTP preset', () => {
  const { html, context, fields } = setup();
  assert.doesNotMatch(html, /option value="cloudflare"/);
  assert.match(html, /Cloudflare Email Routing forwards incoming mail/);
  context.applySmtpPreset('cloudflare');
  assert.equal(fields['#admin-smtp-host'].value, '');
});
test('outgoing presets only fill fields and retain an entered username', () => {
  const { html, context, fields } = setup();
  assert.match(html, /Presets fill connection fields only/);
  context.applySmtpPreset('resend');
  assert.equal(fields['#admin-smtp-host'].value, 'smtp.resend.com');
  assert.equal(fields['#admin-smtp-port'].value, '465');
  assert.equal(fields['#admin-smtp-enc'].value, 'tls');
  assert.equal(fields['#admin-smtp-user'].value, 'resend');
  fields['#admin-smtp-user'].value = 'custom-user';
  context.applySmtpPreset('sendgrid');
  assert.equal(fields['#admin-smtp-user'].value, 'custom-user');
  assert.equal(fields['#admin-smtp-enc'].value, 'starttls');
});
