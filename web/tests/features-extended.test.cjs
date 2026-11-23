const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');

const appJs = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
const cfJs = fs.readFileSync(path.join(__dirname, '../static/js/cloudflare-dns.js'), 'utf8');
const modpackJs = fs.readFileSync(path.join(__dirname, '../static/js/modpacks.js'), 'utf8');
const cloudBkJs = fs.readFileSync(path.join(__dirname, '../static/js/cloud-backups.js'), 'utf8');

function createTestContext(overrides = {}) {
  const modalContainer = { innerHTML: '', id: 'modal-container' };
  const elements = {
    '#modal-container': modalContainer,
    '#sub-label': { value: 'play' },
    '#sub-base': { value: 'example.com' },
    '#sub-target': { value: '203.0.113.1' }
  };

  const toasts = [];
  const apiCalls = [];

  const ctx = {
    window: {},
    document: {
      getElementById: (id) => elements['#' + id] || null,
      querySelector: (sel) => elements[sel] || null,
      querySelectorAll: () => [],
      body: { contains: () => true, innerHTML: '' },
      addEventListener: () => {}
    },
    esc: s => String(s || ''),
    ic: () => '<svg></svg>',
    formatBytes: b => `${b} B`,
    loaderHtml: () => '<div>Loading</div>',
    curServer: () => ({ id: 'srv-1', port: 25565, node_id: 'local' }),
    srvApiPath: (srv, path) => `/api/servers/${srv.id}${path}`,
    toast: (msg, type) => toasts.push({ msg, type }),
    api: async (url, opts = {}) => {
      apiCalls.push({ url, opts });
      if (url === '/api/admin/dns/cloudflare') return { configured: false, masked_token: '', zone_id: '' };
      if (url === '/api/admin/backups/cloud') return { enabled: false, endpoint: '', bucket: '', region: 'us-east-1' };
      if (url.includes('/subdomain/cloudflare/sync')) return { success: true, message: 'DNS synchronized!' };
      return {};
    },
    confirm: () => true,
    closeModal: () => {
      modalContainer.innerHTML = '';
    },
    S: {
      view: 'dash',
      viewModes: {},
      user: { role: 'admin' },
      subdomainDraft: { label: 'play', base_domain: 'example.com', target_host: '203.0.113.1' }
    },
    closeAllCustomSelects: () => {},
    localStorage: {
      getItem: () => null,
      setItem: () => {}
    },
    toasts,
    apiCalls,
    ...overrides
  };
  ctx.window = ctx;
  ctx.$ = (sel) => elements[sel] || null;
  return ctx;
}

test('Cloudflare DNS, Modpacks, and Cloud Backups client modules load and export APIs', () => {
  const context = createTestContext();
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

  // 3. Verify Cloud Backups API object
  assert.ok(context.window.NPCloudBackups, 'NPCloudBackups exported on window');
  assert.equal(typeof context.window.NPCloudBackups.renderTabsHeader, 'function');
  assert.equal(typeof context.window.NPCloudBackups.renderCloudStorageView, 'function');
  assert.equal(typeof context.window.NPCloudBackups.renderUploadButton, 'function');
  const bkTabs = context.window.NPCloudBackups.renderTabsHeader({ id: 'srv-1' });
  assert.match(bkTabs, /Cloud Storage/);
});

test('Mobile strictly forces Tiles view mode without allowing switches', () => {
  function section(start, end) {
    const a = appJs.indexOf(start);
    const b = appJs.indexOf(end, a);
    assert.ok(a >= 0 && b > a, `Failed to find section: ${start}`);
    return appJs.slice(a, b);
  }

  const vmSection = section('function isMobileView()', '// --- Clean Multi-Page');

  // Test 1: On desktop (innerWidth = 1200), respects default or preferences, toggle is rendered
  {
    const ctx = createTestContext({ innerWidth: 1200 });
    vm.createContext(ctx);
    vm.runInContext(vmSection, ctx);

    assert.equal(ctx.isMobileView(), false, 'isMobileView is false on desktop');
    assert.equal(ctx.getViewMode('dash', 'tiles'), 'tiles');
    assert.equal(ctx.getViewMode('pl', 'list'), 'list');
    const toggle = ctx.renderViewModeToggle('dash', 'tiles');
    assert.match(toggle, /class="vm-toggle"/, 'Toggle is rendered on desktop');
  }

  // Test 2: On mobile (innerWidth = 412 or matchMedia matching), strictly forces 'tiles'
  {
    const ctx = createTestContext({
      innerWidth: 390,
      matchMedia: (query) => ({ matches: query.includes('768') })
    });
    vm.createContext(ctx);
    vm.runInContext(vmSection, ctx);

    assert.equal(ctx.isMobileView(), true, 'isMobileView is true on mobile');
    // All scopes return 'tiles', even if fallback or saved was 'list'
    assert.equal(ctx.getViewMode('dash', 'list'), 'tiles', 'Dash forced to tiles on mobile');
    assert.equal(ctx.getViewMode('pl', 'list'), 'tiles', 'Players forced to tiles on mobile');
    assert.equal(ctx.getViewMode('files', 'list'), 'tiles', 'Files forced to tiles on mobile');
    assert.equal(ctx.getViewMode('bk', 'list'), 'tiles', 'Backups forced to tiles on mobile');

    // Toggle button is completely suppressed on mobile
    const toggle = ctx.renderViewModeToggle('dash', 'tiles');
    assert.equal(toggle, '', 'Toggle button returns empty string on mobile');

    // setViewMode is a no-op on mobile
    ctx.setViewMode('dash', 'list');
    assert.equal(ctx.getViewMode('dash', 'tiles'), 'tiles', 'Mode remains tiles');
  }
});

test('Modal opening and closing works and fixes buttons doing nothing', async () => {
  const ctx = createTestContext();
  vm.createContext(ctx);

  vm.runInContext(cfJs, ctx);
  vm.runInContext(cloudBkJs, ctx);
  vm.runInContext(modpackJs, ctx);

  const container = ctx.document.getElementById('modal-container');

  // 1. Cloud Storage Settings modal opens without throwing
  assert.equal(container.innerHTML, '');
  await ctx.window.NPCloudBackups.openSettingsModal();
  assert.match(container.innerHTML, /S3-Compatible Cloud Storage/, 'Cloud storage modal rendered');
  assert.match(container.innerHTML, /id="s3-endpoint"/, 'S3 endpoint input present');
  assert.match(container.innerHTML, /id="s3-bucket"/, 'S3 bucket input present');

  // Close modal
  ctx.closeModal();
  assert.equal(container.innerHTML, '', 'Modal closed and cleared');

  // 2. Cloudflare Settings modal opens without throwing
  await ctx.window.NPCloudflareDNS.openConfigModal();
  assert.match(container.innerHTML, /Cloudflare DNS API Settings/, 'Cloudflare settings modal rendered');
  assert.match(container.innerHTML, /id="cf-modal-token"/, 'Cloudflare token input present');

  // Close modal
  ctx.closeModal();
  assert.equal(container.innerHTML, '');

  // 3. Modpack 1-Click Install confirm modal opens
  ctx.window.NPModpacks.confirmInstall('cobblemon', 'Cobblemon');
  assert.match(container.innerHTML, /Install Modpack: Cobblemon/, 'Modpack confirmation modal rendered');
  assert.match(container.innerHTML, /modpack-do-install-btn/, 'Installation button present');

  ctx.closeModal();
  assert.equal(container.innerHTML, '');
});

test('Sync DNS button handles subdomain validation, unconfigured token, and sync execution', async () => {
  const ctx = createTestContext();
  vm.createContext(ctx);

  vm.runInContext(cfJs, ctx);

  // Case 1: Subdomain fields valid -> calls sync endpoint
  const btn = { disabled: false, textContent: 'Sync' };
  await ctx.window.NPCloudflareDNS.syncDNS(btn);

  assert.ok(ctx.apiCalls.some(c => c.url.includes('/subdomain/cloudflare/sync')), 'Called sync API endpoint');
  assert.ok(ctx.toasts.some(t => t.msg.includes('synchronized')), 'Success toast triggered');

  // Case 2: Subdomain fields empty -> shows validation toast
  ctx.S.subdomainDraft = { label: '', base_domain: '' };
  ctx.document.getElementById = (id) => (id === 'modal-container' ? ctx.document.getElementById('modal-container') : { value: '' });
  await ctx.window.NPCloudflareDNS.syncDNS(btn);
  assert.ok(ctx.toasts.some(t => t.msg.includes('Please set a subdomain label')), 'Validation toast triggered');
});
