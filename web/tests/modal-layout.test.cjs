const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const root = path.join(__dirname, '../static');
const app = fs.readFileSync(path.join(root, 'js/app.js'), 'utf8');
const css = fs.readFileSync(path.join(root, 'css/modal-layout.css'), 'utf8');
const index = fs.readFileSync(path.join(root, 'index.html'), 'utf8');

function section(start, end) {
  const a = app.indexOf(start), b = app.indexOf(end, a);
  assert.ok(a >= 0 && b > a, `Missing source section: ${start}`);
  return app.slice(a, b);
}

// Execute the real renderers with labeled demo data; no network or saved state.
async function render(name, arg) {
  const container = { innerHTML: '' };
  const user = { id: 1, username: 'demo', email: 'demo@example.invalid', role: 'admin' };
  const server = { id: 'demo', node_id: 'local', name: 'Demo server', type: 'paper', version: '1.21.11' };
  const nodes = [{ id: 'local', name: 'Demo local', is_local: true, online: true }, { id: 'remote', name: 'Demo remote', online: true, base_url: 'https://example.invalid' }];
  const sandbox = {
    S: { nodes, adminUsers: [user], user, acct: { un: 'demo' } },
    $: selector => selector === '#modal-container' ? container : null,
    esc: value => String(value ?? '').replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('"', '&quot;'),
    ic: () => '', engineIcon: () => '', setTimeout() {}, enhanceSelects() {},
    getNodeById: id => nodes.find(node => node.id === id),
    displayNodeName: (id, name) => name || id, displayNodeLocation: () => 'Demo location',
    renderNodeWakeSettings: () => '<section class="demo-wake-settings">Demo join-to-start launcher</section>',
    canCreateServers: () => true, getNextSuggestedPort: () => 25565,
    getAllMcVersions: () => ['1.21.11', '1.20.4'], CORES: [['Paper', 'Demo engine'], ['Steel', 'Demo native engine']],
    renderJavaSelectOptions: () => '<option>Java 21</option>', renderJavaAvailabilityStrip: () => 'Demo runtime status',
    loadVersionsCatalog() {}, loadJavaRuntimes() {},
    findServerRef: () => server, curServer: () => server, getServerMaxRamMB: () => 2048,
    isSteelEngine: () => false, getRecommendedJavaForMcVersion: () => '21', onReinstallEngineChange() {},
  };
  const bounds = {
    openNodeModal: ['function openNodeModal(', 'async function testNodeModalConnection('],
    openCreateServerModal: ['function openCreateServerModal(', 'function closeModal('],
    openAdminServerSettingsModal: ['async function openAdminServerSettingsModal(', 'async function submitAdminServerSettings('],
    openAddUserModal: ['function openAddUserModal(', 'async function submitAddUser('],
    openEditUserModal: ['function openEditUserModal(', 'async function submitEditUser('],
    openCreateDatabaseModal: ['function openCreateDatabaseModal(', 'async function submitCreateDatabase('],
    openReinstallModal: ['function openReinstallModal(', 'async function submitReinstall('],
  };
  vm.runInNewContext(section(...bounds[name]), sandbox);
  await sandbox[name](arg);
  return container.innerHTML;
}

const cases = [
  ['openNodeModal', undefined, ['nm-name', 'nm-loc', 'nm-url', 'nm-fqdn', 'nm-token', 'nm-free', 'nm-idle-timeout', 'nm-allow-ram', 'nm-allow-ports', 'nm-allow-java', 'nm-isolation', 'node-modal-test-res', 'nm-test-btn', 'nm-submit-btn']],
  ['openNodeModal', 'remote', ['nm-name', 'nm-url', 'nm-token', 'nm-submit-btn']],
  ['openCreateServerModal', undefined, ['cs-name', 'cs-node', 'cs-type', 'cs-ver-sel', 'cs-ver', 'cs-port', 'cs-java', 'cs-steel-note', 'cs-java-status', 'cs-mem', 'cs-cpu', 'cs-disk', 'cs-owner', 'cs-eula', 'cs-dl', 'cs-aikar', 'cs-submit-btn']],
  ['openAdminServerSettingsModal', 'demo', ['adm-set-owner', 'adm-set-node', 'adm-set-port', 'adm-set-ram', 'adm-set-cpu', 'adm-set-disk', 'adm-set-submit']],
  ['openAddUserModal', undefined, ['add-user-form', 'nu-username', 'nu-firstname', 'nu-lastname', 'nu-email', 'nu-password', 'nu-role', 'nu-avatar-file', 'nu-submit-btn']],
  ['openEditUserModal', 1, ['edit-user-form', 'eu-username', 'eu-firstname', 'eu-lastname', 'eu-email', 'eu-password', 'eu-role', 'eu-submit-btn']],
  ['openCreateDatabaseModal', undefined, ['db-new-name', 'db-new-engine', 'db-new-maxconn', 'db-create-btn']],
  ['openReinstallModal', 'demo', ['reinstall-form', 'reinst-wipe-warn', 'reinst-steel-note', 'reinst-software', 'reinst-ver-sel', 'reinst-version', 'reinst-world', 'reinst-config', 'reinst-plugins', 'reinst-sync-java', 'reinst-submit-btn']],
];

for (const [name, arg, ids] of cases) {
  test(`${name} (${arg ?? 'create'}) opts into explicit columns without losing fields or actions`, async () => {
    const html = await render(name, arg);
    assert.match(html, /<form class="np-modal-form"[^>]*onsubmit="[^"]+"/);
    assert.match(html, /class="[^"]*np-modal-fields/);
    assert.match(html, /class="[^"]*np-modal-actions/);
    assert.match(html, /type="submit"/);
    assert.match(html, /onclick="closeModal\(\)"/);
    const actualIds = [...html.matchAll(/\sid="([^"]+)"/g)].map(match => match[1]);
    assert.equal(new Set(actualIds).size, actualIds.length, 'IDs stay unique');
    for (const id of ids) assert.ok(actualIds.includes(id), `${id} must remain addressable`);
  });
}

test('node token validation, read-only hosting node, and self-role protection survive the grid', async () => {
  assert.match(await render('openNodeModal'), /id="nm-token" required/);
  assert.doesNotMatch(await render('openNodeModal', 'remote'), /id="nm-token" required/);
  assert.match(await render('openNodeModal', 'remote'), /demo-wake-settings/);
  assert.match(await render('openNodeModal', 'local'), /Edit address/);
  assert.match(await render('openAdminServerSettingsModal', 'demo'), /id="adm-set-node"[^>]*disabled/);
  assert.match(await render('openEditUserModal', 1), /id="eu-role"[^>]*disabled/);
  assert.match(await render('openReinstallModal', 'demo'), /data-operation-progress/);
});

test('shared sheets load after workbench/workspace; dedicated wake rules load last', () => {
  const sheets = [...index.matchAll(/href="\/css\/([^"]+)"/g)].map(match => match[1]);
  for (const sheet of ['modal-layout.css', 'node-wake.css']) {
    assert.equal(sheets.filter(name => name === sheet).length, 1);
    assert.ok(sheets.indexOf(sheet) > sheets.indexOf('workbench.css'));
    assert.ok(sheets.indexOf(sheet) > sheets.indexOf('workspace.css'));
  }
  assert.ok(sheets.indexOf('node-wake.css') > sheets.indexOf('modal-layout.css'));
});

test('geometry contract overrides inline widths, preserves special layouts, and never forces dialog height', () => {
  assert.match(css, /--np-modal-width:\s*1024px/);
  assert.match(css, /\.pal > \.pbox:not\(\.db-studio-modal\)/);
  assert.match(css, /max-width:\s*var\(--np-modal-width\) !important/);
  assert.match(css, /100dvh - 2 \* var\(--np-modal-gutter\)/);
  assert.match(css, /#pal > \.pbox \{ overflow: visible;/);
  assert.match(css, /@media \(max-width: 700px\)/);
  assert.match(css, /repeat\(2, minmax\(0, 1fr\)\)/);
  assert.doesNotMatch(css, /(?:min-height|overflow-x:\s*hidden|transition:\s*all|grid-auto-flow:\s*dense)/);
  assert.doesNotMatch(css, /\.pbox\s+\.(?:row|grid|g2)\s*\{/);
});

// Optional real-browser regression, with no new production/test dependency.
// NP_LAYOUT_PLAYWRIGHT points to a Playwright installation; NP_LAYOUT_CHROME
// optionally points to a locally installed Chromium browser.
test('browser: real form markup fits desktop/mobile, preserves tab order and compact confirmations', { skip: !process.env.NP_LAYOUT_PLAYWRIGHT }, async () => {
  const { chromium } = require(process.env.NP_LAYOUT_PLAYWRIGHT);
  const browser = await chromium.launch({ headless: true, ...(process.env.NP_LAYOUT_CHROME ? { executablePath: process.env.NP_LAYOUT_CHROME } : {}) });
  try {
    const page = await browser.newPage();
    const styles = ['style.css', 'workbench.css', 'motion.css', 'workspace.css', 'modal-layout.css'].map(file => fs.readFileSync(path.join(root, 'css', file), 'utf8')).join('\n');
    const load = html => page.setContent(`<html data-theme="dark"><head><style>${styles}</style></head><body>${html}</body></html>`);
    await page.emulateMedia({ reducedMotion: 'reduce' });
    for (const width of [1440, 1024, 768, 701, 700, 390, 320]) {
      await page.setViewportSize({ width, height: 800 });
      for (const [name, arg] of cases) {
        await load(await render(name, arg));
        const box = page.locator('.pbox');
        const size = await box.boundingBox();
        assert.equal(Math.round(size.width), width === 1440 ? 1024 : width - (width <= 700 ? 24 : 32));
        assert.ok(size.height <= 800 - (width <= 700 ? 32 : 64));
        assert.ok(await box.evaluate(el => el.scrollWidth <= el.clientWidth + 1), `${name} at ${width}: horizontal overflow`);
        const columns = await page.locator('.np-modal-form').evaluate(el => getComputedStyle(el).gridTemplateColumns.split(' ').length);
        assert.equal(columns, width <= 700 ? 1 : 2);
        if (width <= 700) for (const fields of await page.locator('.np-modal-fields').all()) {
          assert.equal(await fields.evaluate(el => getComputedStyle(el).gridTemplateColumns.split(' ').length), 1);
        }
      }
    }
    await page.setViewportSize({ width: 1440, height: 800 });
    await load(await render('openAddUserModal'));
    await page.locator('#nu-username').focus();
    await page.keyboard.press('Tab');
    assert.equal(await page.evaluate(() => document.activeElement.id), 'nu-firstname');
    assert.equal(await page.evaluate(() => getComputedStyle(document.activeElement).outlineStyle), 'solid');
    assert.equal(await page.locator('.pbox').evaluate(el => getComputedStyle(el).animationName), 'none');
    await load('<div class="pal"><div class="pbox ui-modal-box" style="max-width:440px"><h3>Demo confirmation</h3><p>Continue?</p><button>Cancel</button></div></div>');
    assert.equal((await page.locator('.pbox').boundingBox()).width, 1024);
    assert.ok((await page.locator('.pbox').boundingBox()).height < 240, 'short confirmation must stay short');
    await load('<div class="pal"><div class="pbox db-studio-modal"><p>Demo SQL workspace</p></div></div>');
    assert.equal((await page.locator('.pbox').boundingBox()).width, 1120);
    await load('<div class="pal" id="pal"><div class="pbox"><div class="np-palette-input"><input id="pq"></div><div class="plist">Demo results</div></div></div>');
    assert.equal((await page.locator('.pbox').boundingBox()).width, 1024);
    assert.equal(await page.locator('.pbox').evaluate(el => getComputedStyle(el).overflow), 'visible');
  } finally { await browser.close(); }
});
