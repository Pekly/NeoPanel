const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const app = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
const wizard = app.slice(app.indexOf('async function openEnable2FAWizard()'), app.indexOf('async function confirmEnable2FA('));
const confirm = app.slice(app.indexOf('async function confirmEnable2FA('), app.indexOf('function openDisable2FAModal()'));

test('TOTP setup never uses an external QR service', () => {
  assert.doesNotMatch(wizard, /qrserver|https?:\/\/|encodeURIComponent\(res\.otpauth_url\)/);
  assert.match(wizard, /qr_data_url/);
  assert.match(wizard, /setupCurrent/);
});

test('Chrome: local-only TOTP QR, safe fallback and stale enrollment suppression', {skip: !process.env.NP_LAYOUT_PLAYWRIGHT}, async () => {
  const {chromium} = require(process.env.NP_LAYOUT_PLAYWRIGHT);
  const browser = await chromium.launch({headless:true, executablePath:process.env.NP_LAYOUT_CHROME});
  try {
    const page = await browser.newPage();
    const network = [];
    await page.route('**/*', route => { network.push(route.request().url()); return route.abort(); });
    await page.setContent('<div id="modal-container"></div>');
    await page.evaluate(() => {
      window.S = {user:{id:1}, token:'one'};
      window.$ = s => document.querySelector(s);
      window.esc = s => String(s).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
      window.ic = () => '';
      window.response = {secret:'ABCDEFGHIJKLMNOP234567ABCDEFGHIJKLMNOP', otpauth_url:'otpauth://totp/panel?secret=private', qr_data_url:'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII='};
      window.api = async () => response;
    });
    await page.addScriptTag({content:wizard + confirm});
    await page.evaluate(() => openEnable2FAWizard());
    assert.match(await page.locator('img').getAttribute('src'), /^data:image\/png;base64,/);
    for (const bad of ['', 'https://external.invalid/qr?secret=private', 'data:image/svg+xml,<svg onload=alert(1)>']) {
      await page.evaluate(value => { response.qr_data_url=value; return openEnable2FAWizard(); }, bad);
      assert.equal(await page.locator('img').count(), 0);
      assert.match(await page.locator('#tfa-wizard-content').innerText(), /Enter the secret key below/);
    }
    await page.evaluate(() => {
      window.api = () => new Promise(resolve => { window.finishSetup=resolve; });
      window.pendingSetup = openEnable2FAWizard();
      S.user = {id:2}; S.token='two';
      finishSetup(response);
      return pendingSetup;
    });
    assert.doesNotMatch(await page.locator('#tfa-wizard-content').innerText(), /Secret Key|ABCDEFGHIJKLMNOP/);
    await page.evaluate(() => {
      window.pendingSetup = openEnable2FAWizard();
      document.querySelector('#modal-container').innerHTML='<div id="tfa-wizard-content">Another dialog</div>';
      finishSetup(response);
      return pendingSetup;
    });
    assert.equal(await page.locator('#tfa-wizard-content').innerText(), 'Another dialog');
    await page.evaluate(() => {
      document.querySelector('#modal-container').innerHTML='<div id="tfa-wizard-content"><input id="tfa-verify-code" value="123456"><button id="tfa-confirm-btn"></button><div id="tfa-wizard-err"></div></div>';
      window.S.acct = {};
      window.toast = () => { throw Error('Stale completion notification'); };
      window.pendingEnable = confirmEnable2FA('test-secret');
      S.user = {id:3}; S.token='three';
      document.querySelector('#modal-container').innerHTML='<div id="tfa-wizard-content">New account</div>';
      finishSetup({recovery_codes:['private-recovery-code']});
      return pendingEnable;
    });
    assert.equal(await page.locator('#tfa-wizard-content').innerText(), 'New account');
    assert.equal(await page.evaluate(() => S.user.two_factor_enabled), undefined);
    assert.deepEqual(network, []);
  } finally { await browser.close(); }
});
