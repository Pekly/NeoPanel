const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
function slice(start, end) {
  const a = source.indexOf(start), b = source.indexOf(end, a);
  assert.ok(a >= 0 && b > a, `${start} / ${end}`);
  return source.slice(a, b);
}
function setup(type = 'steel') {
  const elements = new Map(), calls = [], notices = [];
  const srv = { id: 'srv-test', node_id: 'east', type, name: 'Test', version: 'v0.17.0+mc26.3', jar_file: type === 'steel' ? 'steel.exe' : 'server.jar', memory_max: 4096 };
  const c = vm.createContext({
    S: { token: 'token', servers: [srv], core: type === 'steel' ? 'Steel' : 'Paper', props: {}, user: { role: 'admin' }, nodes: [], ver: srv.version },
    curServer: () => srv, $: key => elements.get(key) || null, $$: () => [],
    esc: value => String(value ?? '').replaceAll('<', '&lt;').replaceAll('"', '&quot;'), ic: () => '',
    toast: message => notices.push(message), api: async (url, opts) => { calls.push([url, opts && JSON.parse(opts.body)]); return url === '/api/steel/versions' ? ['v0.17.0+mc26.3', 'v0.16.0+mc26.2'] : { jar_file: 'steel.exe', version: 'v0.17.0+mc26.3' }; },
    getViewMode: () => 'list', srvh: () => '', getServerMaxRamMB: () => 4096,
    setTimeout: () => {}, setCreateRam: () => {}, Event: class {},
    canCreateServers: () => true, nodeApiPath: (node, url) => `/node/${node}${url}`, srvApiPath: (server, url) => `/node/${server.node_id}/api/servers/${server.id}${url}`,
    closeModal: () => {}, loadDashboard: async () => {}, applyRoute: () => {},
    findServerRef: () => srv, refreshServerStatus: async () => {}, R: () => {}, enhanceSelects: () => {},
    power: () => {}, setBtnLoading: () => {},
    getNodeById: () => null, getBrand: () => '', _settingsUiState: {},
    renderJavaAvailabilityStrip: () => '', renderJavaSelectOptions: () => '',
    resolveJavaSelectValue: () => '21', getPremadeJavaStatusMap: () => new Map(),
    document: { body: { contains: () => false } }, _uiModalResolver: null,
    seedSettingsSavedSnapshot: () => {}, uiAlert: message => notices.push(message),
    sessionStorage: { setItem: () => {} }
  });
  vm.runInContext([
    slice('function fileListIconHtml(', '// --- Server Software Icons & Helpers ---'),
    slice('const ENGINE_ICONS =', 'function onPlayerHeadError('),
    slice('const CORES =', 'const CATS ='),
    slice('function getRecommendedJavaForMcVersion(', 'function resolveJavaSelectValue('),
    slice('function versionRowsHtml()', '// --- Navigation Render ---'),
    slice('function setCreateRam(', 'function getNextSuggestedPort('),
    slice('async function handleCreateServer(', '// --- Port Allocation Management ---'),
    slice('function updateReinstallJavaPreview()', 'async function openAdminServerSettingsModal('),
    slice('function askInstallSoftwareModal(', 'function syncPropsFromServer('),
    slice('function updateSettingsCmdPreview()', '// --- Selectors & Shortcuts ---'),
    slice('function buildSettingsSnapshots(', 'function collectSettingsFormPayload('),
    slice('async function loadProperties(', 'async function persistSettingsPayload('),
    slice('async function loadPluginsList(', 'async function deletePlugin('),
    slice('async function searchModrinth(', 'async function installPlugin('),
  ].join('\n'), c);
  function el(id, value = '') {
    const parent = { style: {}, parentElement: { style: {} } };
    const element = { value, checked: true, disabled: false, style: {}, dataset: {}, options: [], innerHTML: '', textContent: '', closest: () => parent, dispatchEvent: () => {} };
    elements.set('#' + id, element);
    return element;
  }
  return { c, el, srv, calls, notices };
}

test('Steel uses a local official PNG and distinct engine identity', () => {
  const { c } = setup();
  assert.equal(c.engineMeta('Steel').key, 'steel');
  assert.equal(c.getEngineIconUrl('Steel'), '/static/img/steel-logo.png');
  assert.match(c.fileListIconHtml({ name: 'steel.exe' }), /Steel executable \(Rust\)/);
  assert.match(c.fileListIconHtml({ name: 'steel.exe' }), /steel-logo.png/);
  const logo = fs.readFileSync(path.join(__dirname, '../static/img/steel-logo.png'));
  assert.equal(logo.subarray(0, 8).toString('hex'), '89504e470d0a1a0a');
  assert.equal(c.getRecommendedJavaForMcVersion('latest', 'steel'), '');
  assert.equal(c.getRecommendedJavaForMcVersion('1.21.11', 'paper'), '21');
});

test('experimental Steel stays last while Paper remains first and the creation default', () => {
  const { c } = setup();
  const engines = Array.from(vm.runInContext('CORES.map(([name]) => name)', c));
  assert.equal(engines[0], 'Paper');
  assert.equal(engines.at(-1), 'Steel');
  assert.match(vm.runInContext('CORES.find(([name]) => name === "Steel")[1]', c), /Experimental/);
  const reinstall = slice('<select id="reinst-software"', '</select>');
  const options = Array.from(reinstall.matchAll(/<option value="([^"]+)"/g), match => match[1]);
  assert.equal(options[0], 'paper');
  assert.equal(options.at(-1), 'steel');
  assert.match(reinstall, /Steel \(Experimental/);
  assert.ok(slice('<select id="cs-type"', '</select>').includes("slug === 'paper' ? 'selected' : ''"));
});

test('Steel catalog loads release tags only and keeps latest on failures', async () => {
  const { c, calls, notices } = setup();
  assert.deepEqual(Array.from(c.getAllMcVersions('steel')), ['latest']);
  await c.loadVersionsCatalog();
  assert.deepEqual(calls.map(call => call[0]), ['/api/steel/versions']);
  assert.deepEqual(Array.from(c.getAllMcVersions('Steel')), ['latest', 'v0.17.0+mc26.3', 'v0.16.0+mc26.2']);
  c.S.steelVersions = undefined;
  c.api = async () => { throw Error('offline'); };
  await c.loadSteelVersions();
  assert.deepEqual(Array.from(c.getAllMcVersions('steel')), ['latest']);
  assert.ok(notices.some(message => message.includes('Could not load Steel')));
});

test('Steel version browser ignores Minecraft family filters and offers native releases', async () => {
  const { c } = setup();
  await c.loadSteelVersions();
  c.S.sel = 'Steel'; c.S.vf = '1.21';
  for (const mode of ['list', 'icons', 'tiles']) {
    c.getViewMode = () => mode;
    const html = c.versionRowsHtml();
    assert.match(html, /v0.17.0\+mc26.3/);
    assert.match(html, /Native executable/);
    assert.match(html, /data-soft="steel"/);
    assert.doesNotMatch(html, /Java|1\.21\.11/);
  }
});

test('native command previews honor backend executable filenames without JVM or port flags', () => {
  const { c, el, srv } = setup();
  const preview = el('set-cmd-preview');
  c.updateSettingsCmdPreview();
  assert.equal(preview.textContent, '.\\steel.exe');
  srv.jar_file = 'steel';
  c.updateSettingsCmdPreview();
  assert.equal(preview.textContent, './steel');
  el('prop-jar', 'custom steel.exe');
  c.updateSettingsCmdPreview();
  assert.equal(preview.textContent, '".\\custom steel.exe"');
  assert.doesNotMatch(preview.textContent, /java|nogui|--port|-Xm/);
});

test('creation toggles native controls and sends latest with resource memory but no Java', async () => {
  const { c, el, calls } = setup('paper');
  el('cs-type', 'paper'); const version = el('cs-ver', '1.21.11');
  const java = el('cs-java', '21'), flags = el('cs-aikar');
  el('cs-java-status'); el('cs-steel-note'); el('cs-version-label'); el('cs-ver-sel');
  c.setCreateEngine('steel');
  assert.equal(version.value, 'latest');
  assert.equal(java.disabled, true); assert.equal(flags.disabled, true);
  el('cs-mem', '8192'); el('cs-node', 'east'); el('cs-name', 'Native');
  await c.handleCreateServer({ preventDefault() {} });
  const payload = calls.find(([url]) => url === '/node/east/api/servers')[1];
  assert.equal(payload.type, 'steel'); assert.equal(payload.version, 'latest');
  assert.equal(payload.memory_max, 8192); assert.equal(payload.memory_min, 0);
  assert.equal(payload.java_path, ''); assert.equal(payload.java_args, '');
  assert.equal(payload.jar_file, undefined, 'backend chooses host executable filename');
  c.setCreateEngine('paper');
  assert.equal(version.value, '1.21.11'); assert.equal(java.disabled, false);
});

test('reinstall defaults to latest, warns across engine families and never syncs Java for Steel', async () => {
  const { c, el, calls, srv } = setup('paper');
  const form = el('reinstall-form'); form.dataset = { origType: 'paper', selectedType: 'paper' };
  el('reinst-software', 'steel'); const version = el('reinst-version', '1.21.11');
  el('reinst-ver-sel'); const sync = el('reinst-sync-java'); el('reinst-plugins');
  const warning = el('reinst-wipe-warn');
  c.onReinstallEngineChange('steel');
  assert.equal(version.value, 'latest'); assert.equal(sync.disabled, true);
  assert.match(warning.innerHTML, /incompatible|Incompatible/);
  assert.match(warning.innerHTML, /saves\//);
  await c.submitReinstall({ preventDefault() {} }, srv.id, 'east');
  const reinstall = calls.find(([url]) => url.endsWith('/reinstall'));
  assert.equal(reinstall[1].software, 'steel'); assert.equal(reinstall[1].version, 'latest');
  assert.equal(reinstall[1].preserve_plugins, false);
  assert.equal(srv.jar_file, 'steel.exe'); assert.equal(srv.version, 'v0.17.0+mc26.3');
  assert.equal(calls.filter(([, body]) => body?.java_path).length, 0);
  assert.equal(c.incompatibleEngineChange('steel', 'paper'), true);
  assert.equal(c.incompatibleEngineChange('steel', 'steel'), false);
  assert.equal(c.incompatibleEngineChange('paper', 'purpur'), false);
});

test('software confirmation describes native/config incompatibility without Java recommendation', () => {
  const { c, el, srv } = setup('paper');
  const modal = el('modal-container');
  c.askInstallSoftwareModal(srv, 'steel', 'latest');
  assert.match(modal.innerHTML, /Incompatible engine change/);
  assert.match(modal.innerHTML, /config\/config\.toml/);
  assert.doesNotMatch(modal.innerHTML, /Switch Java Runtime|confirm-sync-java"|Java 21|server JAR/);
});

test('Steel settings and plugin entry points avoid Java properties and plugin APIs', async () => {
  const { c, srv, calls } = setup();
  await c.loadProperties(); await c.loadPluginsList(); await c.searchModrinth();
  assert.equal(calls.length, 0);
  const payload = c.buildSettingsSnapshots(srv, { ram: 8192, javaPath: '21', args: '-XX:+UseG1GC', jar: 'steel.exe' }).srvPayload;
  assert.equal(payload.memory_max, 8192); assert.equal(payload.memory_min, 0);
  assert.equal(payload.java_path, ''); assert.equal(payload.java_args, '');
  const method = slice('  set() {', '\n  },');
  vm.runInContext(`var settingsView = {${method}\n}};`, c);
  const html = c.settingsView.set();
  assert.match(html, /Memory Resource Limit \(systemd only\)/);
  assert.match(html, /config\/config\.toml/);
  assert.match(html, /server_port/);
  assert.match(html, /\[server\]/);
  assert.match(html, /including Bubblewrap \+ Systemd launches/);
  assert.match(html, /Without systemd, Steel can still run, but no hard RAM cap is applied/);
  assert.match(html, /<summary>Configuration &amp; resource limits in NeoPanel<\/summary>/);
  assert.ok(!source.includes('config/' + 'server.toml'), 'all Steel guidance uses the upstream config path');
  assert.match(html, /steel.exe/);
  assert.doesNotMatch(html, /id="prop-java"|id="prop-motd"|id="prop-args"|-Xmx|nogui/);
  assert.match(source, /const sendProps = !isSteelEngine\(targetSrv.type\)/);
});

test('software install posts native type and preserves backend version/filename without Java PUT', async () => {
  const { c, calls, srv } = setup('paper');
  c.askInstallSoftwareModal = async () => ({ confirmed: true, backup: false, syncJava: true, recJava: '21' });
  await c.installSoftware('Steel', 'latest');
  assert.deepEqual(calls.map(([url]) => url), ['/node/east/api/servers/srv-test/software']);
  assert.equal(calls[0][1].software, 'steel');
  assert.equal(calls[0][1].version, 'latest');
  assert.equal(srv.jar_file, 'steel.exe');
  assert.equal(c.S.core, 'Steel');
  assert.equal(c.S.ver, 'v0.17.0+mc26.3');
});

test('Steel explanation uses its real name, Rust implementation and pre-alpha status', () => {
  const { c } = setup();
  const html = c.steelConfigNotice();
  assert.match(html, /<h2>Steel<\/h2>/);
  assert.match(html, /Minecraft Java Edition server implementation written in Rust/);
  assert.match(html, /Experimental · Pre-alpha/);
  assert.match(html, /Gameplay features are incomplete and save formats may change/);
  assert.match(html, /Not a Paper or Fabric fork/);
  assert.match(html, /https:\/\/steelmc.dev\/getting-started\/introduction\//);
  assert.match(html, /<details class="steel-info-details">/);
  assert.doesNotMatch(html, /<details[^>]*\bopen\b|Steel native server/);
  assert.match(html, /Open File Manager/);
  assert.doesNotMatch(c.steelConfigNotice(false), /Open File Manager/);
  const css = fs.readFileSync(path.join(__dirname, '../static/css/workbench.css'), 'utf8');
  assert.match(css, /\.srvh\s*\+\s*\.steel-info\s*\{\s*margin-top:\s*20px/);
});
