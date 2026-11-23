const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');

test('Cloudflare DNS, Modpacks, and Cloud Backups client modules load and export APIs', () => {
  const cfJs = fs.readFileSync(path.join(__dirname, '../static/js/cloudflare-dns.js'), 'utf8');
  const modpackJs = fs.readFileSync(path.join(__dirname, '../static/js/modpacks.js'), 'utf8');
  const cloudBkJs = fs.readFileSync(path.join(__dirname, '../static/js/cloud-backups.js'), 'utf8');

  const context = {
    window: {},
    document: {
      getElementById: () => null
    },
    esc: s => String(s || ''),
    ic: () => '<svg></svg>',
    formatBytes: b => `${b} B`,
    loaderHtml: () => '<div>Loading</div>',
    curServer: () => ({ id: 'srv-1', port: 25565 }),
    S: { user: { role: 'admin' }, subdomainDraft: { label: 'play', base_domain: 'example.com' } }
  };
  context.window = context;

  vm.createContext(context);
  vm.runInContext(cfJs, context);
  vm.runInContext(modpackJs, context);
  vm.runInContext(cloudBkJs, context);

  // 1. Verify Cloudflare DNS API object
  assert.ok(context.window.NPCloudflareDNS, 'NPCloudflareDNS exported on window');
  assert.equal(typeof context.window.NPCloudflareDNS.renderSection, 'function');
  assert.equal(typeof context.window.NPCloudflareDNS.syncDNS, 'function');
  const cfHtml = context.window.NPCloudflareDNS.renderSection({ id: 'srv-1' }, {}, 'play.example.com');
  assert.match(cfHtml, /Cloudflare DNS Automation/);

  // 2. Verify Modpacks API object
  assert.ok(context.window.NPModpacks, 'NPModpacks exported on window');
  assert.equal(typeof context.window.NPModpacks.renderTab, 'function');
  assert.equal(typeof context.window.NPModpacks.confirmInstall, 'function');
  const mpHtml = context.window.NPModpacks.renderTab({ id: 'srv-1' });
  assert.match(mpHtml, /Search 1-Click Modpacks/);

  // 3. Verify Cloud Backups API object
  assert.ok(context.window.NPCloudBackups, 'NPCloudBackups exported on window');
  assert.equal(typeof context.window.NPCloudBackups.renderTabsHeader, 'function');
  assert.equal(typeof context.window.NPCloudBackups.renderCloudStorageView, 'function');
  assert.equal(typeof context.window.NPCloudBackups.renderUploadButton, 'function');
  const bkTabs = context.window.NPCloudBackups.renderTabsHeader({ id: 'srv-1' });
  assert.match(bkTabs, /Cloud Storage/);
});
