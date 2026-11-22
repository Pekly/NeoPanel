const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/proxy-settings.js'), 'utf8');
const app = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
function fixture(type = 'velocity', extra = {}) {
  return {type, file:type === 'velocity' ? 'velocity.toml' : 'config.yml', exists:true, revision:'r1', port:25565, can_edit:true, can_view_secret:true, running:false, ready:true,
    config:{motd:'Demo proxy', max_players:100, online_mode:true, forwarding_mode:type === 'velocity' ? 'MODERN' : '', ip_forward:type !== 'velocity', force_key_authentication:true, force_default_server:false, ping_passthrough:'DISABLED', backends:[{name:'lobby', address:'lobby.internal:25566', motd:'', restricted:false}], try:['lobby'], forced_hosts:[], future:{keep:true}}, ...extra};
}
function setup(type = 'velocity', handler) {
  let srv = {id:'one', node_id:'east', type}, root;
  const calls = [], confirmations = [], permissions = new Set(['settings.view', 'settings.edit', 'files.view']);
  const makeRoot = () => ({isConnected:true, innerHTML:'', querySelectorAll:() => [], querySelector:() => null});
  root = makeRoot();
  const c = {S:{token:'token-a', user:{id:1, role:'user'}, view:'set'}, URL, document:{querySelector:() => root},
    curServer:() => srv, hasServerPermission:p => permissions.has(p), srvApiPath:(s, suffix) => `/api/nodes/${s.node_id}/proxy/api/servers/${s.id}${suffix}`,
    api:async (url, options) => {calls.push({url, ...options}); return handler ? handler(url, options) : fixture(type);},
    uiConfirm:async opts => {confirmations.push(opts); return true;}, go:view => {c.S.view = view;}, navigator:{clipboard:{writeText:async () => {}}}};
  c.window = c; vm.runInNewContext(source, c);
  return {c, ui:c.NPProxySettings, calls, confirmations, permissions, get root() {return root;}, setServer: value => {srv = value;},
    remount:() => {root.isConnected = false; root = makeRoot(); return c.NPProxySettings.mount();},
    click:(action, index = '') => root.onclick({target:{closest:() => ({dataset:{action, index:String(index)}, disabled:false})}}),
    input:(dataset, value, type = 'text') => root.oninput({type:'input', target:{dataset, value, type, checked:value}})};
}
function slice(a, b) {const start = app.indexOf(a); const end = app.indexOf(b, start); assert.ok(start >= 0 && end > start); return app.slice(start, end);}

test('supports only the three proxies; type-specific controls and assigned port are rendered', async () => {
  for (const type of ['velocity', 'bungeecord', 'waterfall']) {
    const h = setup(type); await h.ui.mount(); assert.equal(h.ui.supports(type.toUpperCase()), true);
    assert.match(h.root.innerHTML, /Assigned proxy port/); assert.match(h.root.innerHTML, /readonly aria-describedby="proxy-port-note"/);
    assert.doesNotMatch(h.root.innerHTML, /server\.properties|Game Rules/);
    if (type === 'velocity') {assert.match(h.root.innerHTML, /MiniMessage|MODERN/); assert.doesNotMatch(h.root.innerHTML, /data-field="ip_forward"/);}
    else {assert.match(h.root.innerHTML, /data-field="ip_forward"/); assert.doesNotMatch(h.root.innerHTML, /Show forwarding secret|data-field="forwarding_mode"/);}
  }
  for (const type of ['paper', 'fabric', 'steel', '', 'bungee']) {const h = setup(type); assert.equal(h.ui.supports(type), false); assert.equal(h.ui.render({type}), '');}
});
test('incomplete, mismatched and incompatible views fail closed, with no initialize', async () => {
  for (const view of [{}, fixture('velocity', {running:undefined}), fixture('velocity', {can_edit:'yes'}), fixture('waterfall'), fixture('velocity', {config:{motd:'partial'}})]) {
    const h = setup('velocity', async () => view); await h.ui.mount(); await h.click('save'); await h.click('initialize');
    assert.equal(h.calls.length, 1); assert.match(h.root.innerHTML, /Could not confirm/); assert.doesNotMatch(h.root.innerHTML, /data-action="initialize"/);
  }
  const h = setup('waterfall', async () => {throw Object.assign(Error('internal machine name'), {status:400, data:{code:'unsupported_config'}});}); await h.ui.mount();
  assert.match(h.root.innerHTML, /Open config.yml in Files/); assert.doesNotMatch(h.root.innerHTML, /internal machine/);
});
test('readonly and running views retain readable config but cannot mutate, initialize or save', async () => {
  for (const extra of [{can_edit:false}, {running:true}]) {
    const h = setup('velocity', async () => fixture('velocity', extra)); await h.ui.mount(); h.input({field:'motd'}, 'changed'); await h.click('add-backend'); await h.click('save');
    assert.equal(h.calls.length, 1); assert.match(h.root.innerHTML, /fieldset disabled/); assert.doesNotMatch(h.root.innerHTML, /value="changed"/);
  }
  const h = setup(); h.permissions.delete('settings.view'); await h.ui.mount(); assert.equal(h.calls.length, 0);
});
test('missing config is initialized only explicitly; empty returned backends are not invented', async () => {
  const h = setup('velocity', async (_, o) => fixture('velocity', {exists:o.method === 'POST', ready:false, config:{...fixture().config, backends:[], try:[]}}));
  await h.ui.mount(); assert.equal(h.calls.length, 1); assert.match(h.root.innerHTML, /Create config/);
  await h.click('initialize'); assert.equal(h.calls[1].method, 'POST'); assert.match(h.calls[1].url, /\/proxy-config\/initialize$/); assert.deepEqual(JSON.parse(h.calls[1].body), {});
  assert.match(h.root.innerHTML, /No backend servers/); await h.click('initialize'); assert.equal(h.calls.length, 2);
  h.input({field:'motd'}, 'changed'); await h.click('save'); assert.equal(h.calls.length, 2); assert.match(h.root.innerHTML, /Add at least one backend/);
});
test('address and name validation accepts private DNS and bracketed IPv6, rejects URLs and invalid ports', () => {
  const {ui} = setup();
  for (const address of ['private:25566', 'lobby.internal:1', '[::1]:25566', '[2001:db8::1]:65535', '192.168.0.1:25565']) assert.equal(ui.addressOK(address), true, address);
  for (const address of ['https://host:25565', 'host', 'host:0', 'host:65536', '::1:25565', '[:::]:123', 'host:3/path', 'a b:3']) assert.equal(ui.addressOK(address), false, address);
  for (const name of ['try', 'TRY', 'has space', 'ümlaut', 'a'.repeat(65), '']) {const c = fixture().config; c.backends[0].name = name; assert.notEqual(ui.validate(c, 'velocity'), '');}
  const c = fixture().config; c.backends.push({...c.backends[0], name:'LOBBY'}); assert.match(ui.validate(c, 'velocity'), /unique/);
});
test('rename/remove update references; reorder is bounded; unresolved and duplicate references cannot save', () => {
  const {ui} = setup(), c = fixture().config; c.backends.push({...c.backends[0], name:'survival'}); c.try.push('survival'); c.forced_hosts.push({host:'play.example.com', servers:['lobby', 'survival']});
  ui.rename(c, 0, 'hub'); assert.deepEqual(c.try, ['hub', 'survival']); assert.deepEqual(c.forced_hosts[0].servers, ['hub', 'survival']);
  ui.move(c.try, 1, -1); assert.deepEqual(c.try, ['survival', 'hub']); ui.move(c.try, 0, -1); assert.deepEqual(c.try, ['survival', 'hub']);
  assert.equal(ui.validate(c, 'velocity'), ''); assert.match(ui.validate(c, 'bungeecord'), /exactly one/);
  ui.remove(c, 0); assert.deepEqual(c.try, ['survival']); assert.deepEqual(c.forced_hosts[0].servers, ['survival']);
  c.try.push('SURVIVAL'); assert.match(ui.validate(c, 'velocity'), /fallback/); c.try = ['missing']; assert.match(ui.validate(c, 'velocity'), /fallback/);
  c.try = ['survival']; c.forced_hosts[0].servers = []; assert.match(ui.validate(c, 'velocity'), /valid server/);
});
test('add/remove backend and fallback controls edit a deep cloned DTO, preserving unknown fields', async () => {
  const h = setup('velocity', async (_, o) => o.method === 'PUT' ? {...fixture(), config:JSON.parse(o.body).config, revision:'r2'} : fixture());
  await h.ui.mount(); await h.click('add-backend'); h.input({list:'backends', index:'1', key:'name'}, 'survival'); h.input({list:'backends', index:'1', key:'address'}, '[::1]:25567'); await h.click('add-fallback'); await h.click('up', 1);
  await h.click('save'); const body = JSON.parse(h.calls.at(-1).body); assert.deepEqual(body.config.future, {keep:true}); assert.deepEqual(body.config.try, ['survival', 'lobby']); assert.equal(body.config.force_default_server, false); assert.equal(body.revision, 'r1'); assert.equal(body.confirm_offline, false);
  assert.match(h.root.innerHTML, /Proxy config saved/); assert.equal(h.calls.at(-1).url, '/api/nodes/east/proxy/api/servers/one/proxy-config');
  await h.click('remove-backend', 1); await h.click('save'); assert.deepEqual(JSON.parse(h.calls.at(-1).body).config.try, ['lobby']);
});
test('409 keeps draft and revision, disables repeat writes, and refresh requires explicit discard', async () => {
  const h = setup('velocity', async (_, o) => {if (o.method === 'PUT') throw Object.assign(Error('stale'), {status:409}); return fixture();});
  await h.ui.mount(); h.input({field:'motd'}, 'My draft'); await h.click('save'); assert.match(h.root.innerHTML, /value="My draft"/); assert.match(h.root.innerHTML, /Draft kept/);
  await h.click('save'); assert.equal(h.calls.length, 2); await h.remount(); assert.match(h.root.innerHTML, /value="My draft"/); assert.equal(h.calls.length, 2);
  h.c.uiConfirm = async () => false; await h.click('refresh'); assert.equal(h.calls.length, 2); assert.match(h.root.innerHTML, /value="My draft"/);
  h.c.uiConfirm = async () => true; await h.click('refresh'); assert.equal(h.calls.length, 3); assert.doesNotMatch(h.root.innerHTML, /value="My draft"/);
});
test('background render keeps dirty drafts without an additional GET', async () => {
  const h = setup(); await h.ui.mount(); h.input({field:'motd'}, '<script>draft</script>'); await h.remount(); assert.equal(h.calls.length, 1); assert.match(h.root.innerHTML, /&lt;script&gt;draft/); assert.doesNotMatch(h.root.innerHTML, /<script>/);
});
test('offline mode must be explicitly confirmed for each save', async () => {
  const h = setup(); await h.ui.mount(); h.input({field:'online_mode'}, false, 'checkbox'); h.c.uiConfirm = async () => false; await h.click('save'); assert.equal(h.calls.length, 1);
  h.c.uiConfirm = async options => {assert.match(options.message, /impersonate/); return true;}; await h.click('save'); assert.equal(JSON.parse(h.calls.at(-1).body).confirm_offline, true);
});
test('session/server/node/user/route changes during offline confirmation cannot submit a stale save', async () => {
  for (const mutate of [h => {h.c.S.token = 'new';}, h => {h.c.S.user.id = 2;}, h => {h.c.S.view = 'dash';}, h => h.setServer({id:'two', node_id:'east', type:'velocity'}), h => h.setServer({id:'one', node_id:'west', type:'velocity'}), h => h.permissions.delete('settings.edit')]) {
    let resolve; const h = setup(); await h.ui.mount(); h.input({field:'online_mode'}, false, 'checkbox'); h.c.uiConfirm = () => new Promise(r => resolve = r);
    const pending = h.click('save'); mutate(h); resolve(true); await pending; assert.equal(h.calls.length, 1);
  }
});
test('rapid clicks produce a single write and stale GET/secret responses are ignored', async () => {
  let resolve; const h = setup('velocity', async (_, o) => o.method === 'PUT' ? new Promise(r => resolve = r) : fixture());
  await h.ui.mount(); h.input({field:'motd'}, 'changed'); const pending = h.click('save'); await h.click('save'); assert.equal(h.calls.length, 2); resolve(fixture()); await pending;
  for (const operation of ['read', 'secret']) {
    let finish; const x = setup('velocity', (url) => operation === 'read' || url.endsWith('/secret') ? new Promise(r => finish = r) : fixture());
    const p = operation === 'read' ? x.ui.mount() : (await x.ui.mount(), x.click('secret'));
    x.c.S.token = 'other'; x.ui.sync(); finish(operation === 'read' ? fixture() : {secret:'late-secret'}); await p;
    assert.doesNotMatch(x.root.innerHTML, operation === 'read' ? /value="Demo proxy"/ : /late-secret/);
  }
});
test('secret is explicit, permission gated and cleared on hide, refresh, save, rerender and session change', async () => {
  for (const permission of ['settings.edit', 'files.view']) {const h = setup(); h.permissions.delete(permission); await h.ui.mount(); await h.click('secret'); assert.equal(h.calls.length, 1); assert.doesNotMatch(h.root.innerHTML, /Show forwarding secret/);}
  const denied = setup('velocity', async () => fixture('velocity', {can_view_secret:false})); await denied.ui.mount(); await denied.click('secret'); assert.equal(denied.calls.length, 1);
  const h = setup('velocity', async url => url.endsWith('/secret') ? {secret:'demo-private-secret'} : fixture()); await h.ui.mount(); assert.equal(h.calls.length, 1);
  for (const action of ['close-secret', 'refresh', 'save', 'rerender']) {
    await h.click('secret'); assert.match(h.root.innerHTML, /demo-private-secret/);
    if (action === 'save') h.input({field:'motd'}, 'changed');
    if (action === 'rerender') await h.remount(); else await h.click(action);
    assert.doesNotMatch(h.root.innerHTML, /demo-private-secret/);
  }
  assert.doesNotMatch(source, /localStorage|sessionStorage/);
});
test('proxy entry points never request Java properties or quick configs; common saves retain resources', async () => {
  const calls = []; let type = 'velocity'; const srv = {id:'one', node_id:'east', type, port:25565};
  const c = {S:{servers:[srv], props:{}, view:'set'}, curServer:() => srv, seedSettingsSavedSnapshot:() => {},
    hasServerPermission:() => true, api:async (...args) => {calls.push(args); return {};}, srvApiPath:(_, suffix) => '/srv' + suffix,
    $:() => ({}), document:{querySelector:() => null}, go:() => {}, applyRoute:() => {}, uiAlert:() => {},
    _settingsSaveInFlight:false, _settingsSaveQueued:null, _settingsLastSaved:{}, getSettingsSnapshotKey:() => 'one', setSettingsSaveStatus:() => {}, clock:() => '', toast:() => {}, syncHeroFromSavedServer:() => {}};
  c.window = c;
  vm.runInNewContext([slice('function isSteelEngine(', 'function serverExecutable('), slice('async function loadProperties(', 'async function persistSettingsPayload('), slice('async function openQuickConfigsModal(', 'function renderQuickConfigsModalBody('), slice('async function saveQuickConfigs(', '// --- Client Area Provisioning'), slice('function buildSettingsSnapshots(', 'function collectSettingsFormPayload('), slice('async function persistSettingsPayload(', 'async function saveProperties(')].join('\n'), c);
  for (const proxy of ['velocity', 'waterfall', 'bungeecord']) {
    srv.type = proxy; await c.loadProperties(); await c.openQuickConfigsModal('one'); await c.saveQuickConfigs({preventDefault() {}}, 'one'); assert.equal(calls.length, 0);
    await c.persistSettingsPayload(srv, {ok:true, data:{name:'Proxy', ram:2048, javaPath:'21', args:'-XX:+UseG1GC', jar:'proxy.jar', port:25565}}, {autoSave:true, skipRender:true});
    assert.equal(calls.length, 1); assert.equal(calls[0][0], '/srv'); const payload = JSON.parse(calls[0][1].body); assert.equal(payload.memory_max, 2048); assert.equal(payload.java_path, '21'); calls.length = 0;
  }
});
test('startup preview omits nogui only for proxies; Paper and Steel behavior retained', () => {
  let srv = {type:'velocity', jar_file:'proxy.jar', memory_max:2048}; const preview = {};
  const c = {$:selector => selector === '#set-cmd-preview' ? preview : null, curServer:() => srv, getServerMaxRamMB:() => 2048, resolveJavaSelectValue:() => '21'};
  vm.runInNewContext(slice('function isSteelEngine(', 'function steelConfigNotice(') + slice('function javaCommandLabel(', 'function onPlayerHeadError(') + slice('function updateSettingsCmdPreview()', '// --- Selectors & Shortcuts ---'), c);
  for (const type of ['velocity', 'waterfall', 'bungeecord']) {srv.type = type; c.updateSettingsCmdPreview(); assert.doesNotMatch(preview.textContent, /nogui/); assert.match(preview.textContent, /-Xmx2048M/);}
  srv.type = 'paper'; c.updateSettingsCmdPreview(); assert.match(preview.textContent, /nogui$/);
  srv = {type:'steel', jar_file:'steel'}; c.updateSettingsCmdPreview(); assert.equal(preview.textContent, './steel');
});
test('integration mounts beside automation, keeps card out of admin, and guards secret API with files.view', () => {
  assert.match(app, /NPProxySettings\?\.mount\(\)/); assert.match(app, /NPProxySettings\?\.sync\(\)/); assert.match(app, /el.id !== 'proxy-settings'/);
  const permissions = new Set(['settings.edit']); const c = {S:{user:{role:'user'}}, hasServerPermission:p => permissions.has(p)};
  vm.runInNewContext(slice('function clientRequestPermission(', 'function applyPermissionControls('), c);
  assert.throws(() => c.assertClientRequestAccess('/api/servers/one/proxy-config/secret', {method:'POST'}), /files.view/);
  permissions.add('files.view'); assert.doesNotThrow(() => c.assertClientRequestAccess('/api/servers/one/proxy-config/secret', {method:'POST'}));
});

test('Settings replaces Java gameplay controls only for proxies and retains common policy locks', () => {
  const h = setup();
  const srv = {id:'one', node_id:'east', type:'velocity', name:'Demo', version:'latest', port:25565, memory_max:2048, jar_file:'proxy.jar'};
  const c = {S:{props:{}, user:{id:1, role:'user'}, activeServerId:'one'}, curServer:() => srv, NPProxySettings:h.ui,
    esc:value => String(value ?? ''), ic:() => '', srvh:() => '', getBrand:() => '',
    resolveJavaSelectValue:() => '21', getServerMaxRamMB:() => 2048, getAllMcVersions:() => ['latest'],
    getNodeById:() => ({allow_user_ram:false, allow_user_java:false, allow_user_ports:false}), _settingsUiState:{},
    renderJavaSelectOptions:() => '<option>21</option>', renderJavaAvailabilityStrip:() => '', renderMinecraftTextHtml:() => ''};
  c.window = c;
  vm.runInNewContext(slice('function isSteelEngine(', 'function steelConfigNotice(') + slice('function javaCommandLabel(', 'function onPlayerHeadError(') + `var settingsView = {${slice('  set() {', '\n  },')}\n}};`, c);
  for (const type of ['velocity', 'waterfall', 'bungeecord']) {
    srv.type = type; const html = c.settingsView.set(); assert.match(html, /id="proxy-settings"/);
    assert.doesNotMatch(html, /server.properties|Game Rules|id="prop-max"|id="prop-port"|nogui/);
    for (const id of ['set-srv-name', 'prop-java', 'prop-ram', 'prop-jar', 'prop-args', 'prop-autorestart']) assert.ok(html.includes(`id="${id}"`), id);
    assert.match(html, /id="prop-ram" disabled/); assert.match(html, /id="prop-java" disabled/);
  }
  srv.type = 'paper'; const html = c.settingsView.set(); assert.match(html, /server.properties|Game Rules/); assert.doesNotMatch(html, /id="proxy-settings"/); assert.match(html, /nogui/);
});
test('transient duplicate names while typing never steal another backend’s references', () => {
  const {ui} = setup(), c = fixture().config;
  c.backends.push({...c.backends[0], name:'survival'}); c.try.push('survival'); c.forced_hosts.push({host:'play.example.com', servers:['survival']});
  ui.rename(c, 0, 'survival'); assert.match(ui.validate(c, 'velocity'), /unique/);
  ui.rename(c, 0, 'survival2'); assert.deepEqual(c.try, ['survival2', 'survival']); assert.deepEqual(c.forced_hosts[0].servers, ['survival']);
  assert.equal(ui.validate(c, 'velocity'), '');
});
test('failed or partial save never claims success, keeps draft and prevents an automatic retry', async () => {
  for (const response of [null, {}, fixture('velocity', {ready:false}), fixture('velocity', {exists:false}), fixture('velocity', {validation_error:'Invalid forwarding mode'})]) {
    const h = setup('velocity', async (_, o) => {if(o.method === 'PUT') {if(response === null) throw Error('offline'); return response;} return fixture();});
    await h.ui.mount(); h.input({field:'motd'}, 'keep me'); await h.click('save'); await h.click('save');
    assert.equal(h.calls.length, 2); assert.match(h.root.innerHTML, /value="keep me"/); assert.doesNotMatch(h.root.innerHTML, /Proxy config saved/);
  }
});

test('API-supported displayed slot bounds load and save instead of falling back to Files', async () => {
  for (const max_players of [0, 100001, 2147483647]) {
    const view = fixture(); view.config.max_players = max_players;
    const h = setup('velocity', async () => view); await h.ui.mount();
    assert.equal(h.ui.valid(view, 'velocity'), true);
    assert.equal(h.ui.validate(view.config, 'velocity'), '');
    assert.match(h.root.innerHTML, /Displayed slots/);
    assert.doesNotMatch(h.root.innerHTML, /Open velocity.toml in Files/);
    h.input({field:'motd'}, 'changed'); await h.click('save');
    assert.equal(JSON.parse(h.calls.at(-1).body).config.max_players, max_players);
    assert.match(h.root.innerHTML, /Proxy config saved/);
  }
});
test('network and missing worker API failures offer Retry, not a misleading Files fallback', async () => {
  for (const status of [0, 401, 403, 404, 405, 502, 503]) {
    let fail = true;
    const h = setup('velocity', async () => {if (fail) throw Object.assign(Error('private machine detail'), {status}); return fixture();});
    await h.ui.mount();
    assert.match(h.root.innerHTML, /Retry/);
    assert.doesNotMatch(h.root.innerHTML, /Open velocity.toml in Files|private machine detail|data-action="initialize"/);
    fail = false; await h.click('refresh'); assert.match(h.root.innerHTML, /Displayed slots/);
  }
});

test('recoverable field errors retain their values and can be corrected without Files', async () => {
  for (const [field, value, fix, message] of [
    ['max_players', -1, 0, /Displayed slots must/],
    ['forwarding_mode', 'TYPO', 'NONE', /valid player forwarding/],
    ['ping_passthrough', 'TYPO', 'ALL', /valid server-list ping/]
  ]) {
    const view = fixture('velocity', {ready:false}); view.config[field] = value;
    const h = setup('velocity', async (_, o) => o.method === 'PUT' ? {...fixture(), config:JSON.parse(o.body).config} : view);
    await h.ui.mount(); assert.match(h.root.innerHTML, message); assert.doesNotMatch(h.root.innerHTML, /Open velocity.toml in Files/);
    assert.match(h.root.innerHTML, new RegExp(`value="${value}"`));
    h.input({field:'motd'}, 'changed'); await h.click('save'); assert.equal(h.calls.length, 1);
    h.input({field}, fix); await h.click('save'); assert.equal(h.calls.length, 2); assert.match(h.root.innerHTML, /Proxy config saved/);
  }
});
test('structured failures distinguish unsafe files, unsupported config, busy state and missing config', async () => {
  for (const [code, expected] of [['file_unavailable', /permissions, file size and server path links/], ['unsafe_secret', /without replacing an existing secret/], ['operation_busy', /Another server action/], ['server_running', /Stop the proxy/], ['config_missing', /Create config is offered only/], ['unsupported_config', /Open velocity.toml in Files/]]) {
    const h = setup('velocity', async () => {throw Object.assign(Error('private detail'), {status:400, data:{code}});});
    await h.ui.mount(); assert.match(h.root.innerHTML, expected); assert.match(h.root.innerHTML, /Retry/);
    assert.doesNotMatch(h.root.innerHTML, /private detail|data-action="initialize"/);
    assert.equal(h.root.innerHTML.includes('data-action="files"'), code === 'unsupported_config');
  }
});
