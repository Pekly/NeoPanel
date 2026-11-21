const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');

function setup() {
  const elements = {
    '#adm-panel-url-error': { textContent: '', hidden: true },
    '#panel-url-save-btn': { disabled: false },
    '#panel-url-reset-btn': { disabled: true }
  };
  const notices = [];
  const c = vm.createContext({
    S: {}, URL, location: { origin: 'https://current.example.com:8443', pathname: '/admin/panel-address', search: '?token=private' }, api: async () => ({ public_url: '', source: 'unset' }),
    $: selector => elements[selector] || null,
    R: () => {}, toast: (...args) => notices.push(args), ic: () => '',
    esc: value => String(value ?? '').replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('"', '&quot;')
  });
  const stateStart = source.indexOf('  adminPanelUrl: null,');
  const stateEnd = source.indexOf('  adminTurnstile:', stateStart);
  assert.ok(stateStart > 0 && stateEnd > stateStart);
  vm.runInContext('Object.assign(S, {' + source.slice(stateStart, stateEnd) + '});', c);
  const start = source.indexOf('function renderAdminPanelUrlCard()');
  const end = source.indexOf('function onDaemonInputChange(', start);
  assert.ok(start > 0 && end > start);
  vm.runInContext(source.slice(start, end), c);
  return { c, elements, notices };
}

function configured(c) {
  c.S.adminPanelUrl = { public_url: 'https://panel.example.com', source: 'config' };
}

test('Appearance places an independent panel-address card before network defaults and loads it with admin data', () => {
  const appearance = source.slice(source.indexOf('} else if (t === 4) {'));
  assert.ok(appearance.indexOf('${renderAdminPanelUrlCard()}') < appearance.indexOf('Network and server defaults'));
  const load = source.slice(source.indexOf('async function loadAdminData('), source.indexOf('function openNodeModal('));
  assert.match(load, /loadAdminPanelUrl\(true\)/);
  assert.match(source, /Public Host IP \/ FQDN \(Connect Address &amp; SFTP Host\)/);
});

test('GET uses the API helper and distinguishes unset, config, and database sources', async () => {
  const { c } = setup();
  for (const [kind, label] of [['unset', 'Not configured'], ['config', 'Configuration fallback'], ['settings', 'Database override']]) {
    c.api = async endpoint => {
      assert.equal(endpoint, '/api/admin/panel-url');
      return { public_url: kind === 'unset' ? '' : 'https://panel.example.com', source: kind };
    };
    await c.loadAdminPanelUrl(true);
    assert.match(c.renderAdminPanelUrlCard(), new RegExp(label));
    assert.equal(c.S.panelUrlLoading, false);
  }
});

test('card explains origin requirements, deployment prerequisites, and all-account passkey migration', () => {
  const { c } = setup();
  configured(c);
  const html = c.renderAdminPanelUrlCard();
  for (const text of ['passkeys, invitation links, and password-reset links', 'exact HTTPS origin', 'localhost development', 'cannot be cleared', 'SQLite', 'server.public_url', 'does not configure DNS', 'HTTPS certificates', 'reverse proxy', 'Remove all passkeys from every account', 'disabling passkeys does not remove them']) {
    assert.ok(html.includes(text), text);
  }
  assert.match(html, /type="text"[^>]*inputmode="url"[^>]*required/);
  assert.match(html, /role="alert"/);
});

test('unsaved and intentionally empty drafts survive reloads and rendering', async () => {
  const { c } = setup();
  configured(c);
  for (const draft of ['https://new.example.com', '']) {
    c.onPanelUrlInputChange(draft);
    c.api = async () => ({ public_url: 'https://changed.example.com', source: 'settings' });
    await c.loadAdminPanelUrl(true);
    assert.equal(c.S.panelUrlDraft, draft);
    assert.ok(c.renderAdminPanelUrlCard().includes(`value="${draft}"`));
  }
});

test('newest GET wins over a late refresh', async () => {
  const { c } = setup();
  let finish;
  c.api = () => new Promise(resolve => { finish = resolve; });
  const old = c.loadAdminPanelUrl(true);
  c.api = async () => ({ public_url: 'https://new.example.com', source: 'settings' });
  await c.loadAdminPanelUrl(true);
  finish({ public_url: 'https://old.example.com', source: 'config' });
  await old;
  assert.equal(c.S.adminPanelUrl.public_url, 'https://new.example.com');
  assert.equal(c.S.panelUrlLoading, false);
});

test('PUT saves only public_url, uses normalized server response, and never mutates game/SFTP settings', async () => {
  const { c, notices } = setup();
  configured(c);
  const daemon = { fqdn: 'play.example.com', sftp_port: 2022 };
  c.S.brandingDaemon = daemon;
  c.onPanelUrlInputChange(' https://PANEL.example.com:443/ ');
  c.api = async (endpoint, options) => {
    assert.equal(endpoint, '/api/admin/panel-url');
    assert.equal(options.method, 'PUT');
    assert.deepEqual(JSON.parse(options.body), { public_url: 'https://panel.example.com' });
    return { public_url: 'https://panel.example.com', source: 'settings' };
  };
  let prevented = false;
  await c.saveAdminPanelUrl({ preventDefault() { prevented = true; } });
  assert.equal(prevented, true);
  assert.equal(c.S.panelUrlDraft, null);
  assert.equal(c.S.adminPanelUrl.source, 'settings');
  assert.equal(c.S.adminPanelUrl.public_url, 'https://panel.example.com');
  assert.equal(c.S.brandingDaemon, daemon);
  assert.equal(notices[0][1], 'ok');
});

test('saving disables input and submit across rerenders, preventing duplicate PUT and GET requests', async () => {
  const { c } = setup();
  configured(c);
  c.onPanelUrlInputChange('https://new.example.com');
  let finish, requests = 0;
  c.api = () => { requests++; return new Promise(resolve => { finish = resolve; }); };
  const pending = c.saveAdminPanelUrl();
  assert.equal(c.S.panelUrlSaving, true);
  const html = c.renderAdminPanelUrlCard();
  assert.match(html, /<input[^>]*id="adm-panel-url"[^>]*disabled/);
  assert.match(html, /<button[^>]*id="panel-url-save-btn"[^>]*disabled/);
  await c.saveAdminPanelUrl();
  await c.loadAdminPanelUrl();
  c.onPanelUrlInputChange('https://unexpected.example.com');
  assert.equal(c.S.panelUrlDraft, 'https://new.example.com');
  assert.equal(requests, 1);
  finish({ public_url: 'https://new.example.com', source: 'settings' });
  await pending;
  assert.equal(c.S.panelUrlSaving, false);
  assert.doesNotMatch(c.renderAdminPanelUrlCard(), /id="panel-url-save-btn"[^>]*disabled/);
});

test('blank addresses cannot be saved or used to clear the setting', async () => {
  const { c, elements } = setup();
  configured(c);
  let requests = 0;
  c.api = async () => { requests++; };
  c.onPanelUrlInputChange('   ');
  assert.equal(elements['#panel-url-save-btn'].disabled, true);
  await c.saveAdminPanelUrl();
  assert.equal(requests, 0);
  assert.match(c.S.panelUrlError, /cannot be cleared/);
});

test('validation and passkey-hostname errors remain visible and preserve the draft for correction', async () => {
  for (const message of ['public_url must be an HTTPS origin', 'Remove all passkeys before changing the hostname']) {
    const { c, elements, notices } = setup();
    configured(c);
    c.onPanelUrlInputChange('https://new.example.com');
    c.api = async () => { throw Error(message); };
    await c.saveAdminPanelUrl();
    assert.equal(c.S.panelUrlError, message);
    assert.ok(c.renderAdminPanelUrlCard().includes(message));
    assert.equal(c.S.panelUrlDraft, 'https://new.example.com');
    assert.equal(c.S.adminPanelUrl.source, 'config');
    assert.equal(c.S.panelUrlSaving, false);
    assert.equal(notices[0][1], 'bad');
    c.onPanelUrlInputChange('https://new.example.com');
    assert.equal(c.S.panelUrlError, '');
    assert.equal(elements['#adm-panel-url-error'].hidden, true);
  }
});

test('GET errors show retry rather than falsely claiming the address is unset', async () => {
  const { c } = setup();
  c.api = async () => { throw Error('Connection failed'); };
  await c.loadAdminPanelUrl(true);
  assert.equal(c.S.adminPanelUrl, null);
  const html = c.renderAdminPanelUrlCard();
  assert.match(html, /Unavailable/);
  assert.match(html, /Connection failed/);
  assert.match(html, /Retry/);
  assert.match(html, /id="panel-url-save-btn"[^>]*disabled/);
  assert.equal(c.S.panelUrlLoading, false);
  c.api = async () => ({ public_url: '', source: 'unset' });
  await c.loadAdminPanelUrl(true);
  assert.equal(c.S.panelUrlError, '');
});

test('server-provided URLs, draft values, and errors are HTML escaped', () => {
  const { c } = setup();
  c.S.adminPanelUrl = { public_url: '<script>bad</script>', source: 'settings' };
  c.S.panelUrlDraft = '" onfocus="bad';
  c.S.panelUrlError = '<img src=x onerror=bad>';
  const html = c.renderAdminPanelUrlCard();
  assert.doesNotMatch(html, /<script>|<img|value="" onfocus=/);
  assert.match(html, /&quot; onfocus=&quot;bad/);
  assert.match(html, /&lt;img/);
});

test('hostname shortcut previews and saves HTTPS without changing game FQDN', async () => {
  const {c}=setup(); configured(c);
  c.onPanelUrlInputChange(' NEW.example.com:8443 ');
  assert.match(c.renderAdminPanelUrlCard(),/Will use: https:\/\/new.example.com:8443/);
  let payload;c.api=async(url,options)=>{payload=JSON.parse(options.body);return {public_url:payload.public_url,source:'settings'};};
  await c.saveAdminPanelUrl();assert.equal(payload.public_url,'https://new.example.com:8443');
});
test('address normalization accepts explicit localhost HTTP and drops default ports', () => {
  const {c}=setup();
  for(const [raw,want] of [['panel.example','https://panel.example'],[' HTTPS://PANEL.example:443/ ','https://panel.example'],['http://localhost:8123','http://localhost:8123'],['http://[::1]:8123','http://[::1]:8123']]) assert.equal(c.normalizePanelAddress(raw),want);
});
test('invalid addresses cannot be saved or become clickable links', async () => {
  const {c,elements}=setup();configured(c);let calls=0;c.api=async()=>calls++;
  elements['#panel-url-open']={href:'',hidden:false};elements['#adm-panel-url-preview']={textContent:''};
  for(const raw of ['javascript:alert(1)','https://panel.example/path','panel.example?secret=x','panel.example#','user@panel.example','panel.example\\evil','http://panel.example','https://192.0.2.1','panel.example:0','panel.example:65536','panel.example:', '//evil.example','panel.example.','bad_host.example']) {
    c.onPanelUrlInputChange(raw);assert.equal(elements['#panel-url-save-btn'].disabled,true,raw);
    assert.equal(elements['#panel-url-open'].hidden,true,raw);await c.saveAdminPanelUrl();assert.ok(c.S.panelUrlError,raw);
  }
  assert.equal(calls,0);
});
test('use browser address copies only the origin and does not save automatically', () => {
  const {c}=setup();configured(c);let calls=0;c.api=async()=>calls++;
  c.useCurrentPanelAddress();assert.equal(c.S.panelUrlDraft,'https://current.example.com:8443');assert.equal(calls,0);
  assert.doesNotMatch(c.renderAdminPanelUrlCard(),/token=private/);
  assert.match(c.renderAdminPanelUrlCard(),/target="_blank" rel="noopener noreferrer"/);
});
test('reset changes restores saved address without a request or changing game settings', () => {
  const {c}=setup();configured(c);c.onPanelUrlInputChange('new.example');c.S.panelUrlError='failed';
  c.resetPanelAddressDraft();assert.equal(c.S.panelUrlDraft,null);assert.equal(c.S.panelUrlError,'');assert.match(c.renderAdminPanelUrlCard(),/value="https:\/\/panel.example.com"/);
  c.S.panelUrlSaving=true;c.S.panelUrlDraft='new.example';c.resetPanelAddressDraft();c.useCurrentPanelAddress();assert.equal(c.S.panelUrlDraft,'new.example');
});
test('preview and open link follow typing without a full page redraw', () => {
  const {c,elements}=setup();configured(c);let renders=0;c.R=()=>renders++;
  elements['#panel-url-open']={href:'',hidden:true};elements['#adm-panel-url-preview']={textContent:''};
  c.onPanelUrlInputChange('new.example');assert.equal(elements['#panel-url-open'].href,'https://new.example');assert.equal(elements['#panel-url-open'].hidden,false);
  assert.equal(elements['#adm-panel-url-preview'].textContent,'Will use: https://new.example');assert.equal(renders,0);
});

test('typing enables Reset without waiting for a rerender', () => {
  const {c,elements}=setup();configured(c);
  c.onPanelUrlInputChange('new.example');assert.equal(elements['#panel-url-reset-btn'].disabled,false);
});
