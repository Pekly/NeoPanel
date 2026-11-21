const test = require('node:test'), assert = require('node:assert/strict');
const fs = require('node:fs'), path = require('node:path'), vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/node-wake.js'), 'utf8');
const defaults = () => ({ enabled: false, bind_address: '', offline_motd: '{server} sleeps\n&bJoin!', version_label: 'Join to start', max_players: 20, icon_data_url: '', use_server_icon: true, starting_message: 'Reconnect', unavailable_message: 'Unavailable', cooldown_seconds: 30 });
const envelope = (config = defaults(), applied = true, error = '') => ({ config, applied, error });
function setup(handler = async () => envelope()) {
  let root, html = ''; const calls = [];
  const document = { activeElement: { focus() {} }, querySelector: id => id === '#modal-container' ? container : id === '#wake-modal' ? root : null };
  const container = { set innerHTML(value) {
    html = value; if (root) root.isConnected = false;
    const nodes = {};
    for (const [, id] of value.matchAll(/id="(wake-[^"]+)"/g)) nodes['#' + id] = { value: '', textContent: '', checked: false, disabled: false, hidden: false, focus() { document.activeElement = this; }, removeAttribute(key) { delete this[key]; } };
    // DOM input.value always coerces to string.
    for (const node of Object.values(nodes)) { let value = ''; Object.defineProperty(node, 'value', { get: () => value, set: v => { value = String(v); } }); }
    root = { nodes, isConnected: true, querySelector: id => nodes[id], addEventListener() {}, remove() { this.isConnected = false; } };
    this.childNodes = [root];
  }, get innerHTML() { return html; }, childNodes: [], replaceChildren(...nodes) { this.childNodes = nodes; } };
  const c = { document, TextEncoder, URL, S: { user: { id: 1, role: 'admin' }, token: 'one' }, displayNodeName: () => '<img onerror="bad">', api: async (url, options) => { calls.push({ url, options }); return handler(url, options); } };
  c.window = c;
  vm.runInNewContext(source, c);
  return { c, calls, ui: c.NPNodeWake, container, node: id => root.nodes['#wake-' + id], submit: () => root.nodes['#wake-form'].onsubmit({ preventDefault() {} }) };
}
test('GET defaults are authoritative and names are escaped', async () => {
  const config = { ...defaults(), max_players: 53, cooldown_seconds: 91, offline_motd: '</textarea><script>x</script>' };
  const h = setup(async () => envelope(config)); await h.ui.open('remote/a');
  assert.equal(h.calls[0].url, '/api/admin/nodes/remote%2Fa/wake'); assert.equal(h.calls[0].options.method, 'GET');
  assert.equal(h.node('max_players').value, '53'); assert.equal(h.node('cooldown_seconds').value, '91'); assert.equal(h.node('offline_motd').value, config.offline_motd);
  assert.doesNotMatch(h.container.innerHTML, /<img onerror|<script>x/); assert.match(h.container.innerHTML, /&lt;img/);
  const app = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
  assert.doesNotMatch(app.slice(0, app.indexOf('function renderNodeWakeSettings(')), /NPNodeWake\.open/);
  assert.match(app, /NPNodeWake\.open\(this.dataset.id, true\)/);
  const index = fs.readFileSync(path.join(__dirname, '../static/index.html'), 'utf8'); assert.ok(index.indexOf('/js/node-wake.js') < index.indexOf('/js/app.js'));
});
test('save sends config only; remote delivery failures remain visible and preserve drafts', async () => {
  const h = setup(async (_, options) => options.method === 'PUT' ? envelope(JSON.parse(options.body), false, '<b>worker unreachable</b>') : envelope());
  await h.ui.open('local'); h.node('enabled').checked = true; h.node('offline_motd').value = '&aEdited\n{version}';
  await h.submit();
  const body = JSON.parse(h.calls[1].options.body); assert.deepEqual(Object.keys(body).sort(), Object.keys(defaults()).sort()); assert.equal(body.enabled, true);
  assert.equal(h.node('offline_motd').value, '&aEdited\n{version}'); assert.equal(h.node('error').textContent, '<b>worker unreachable</b>'); assert.match(h.node('status').textContent, /Pending worker/);
});
test('pending disable is not reported as actually disabled; retry sends saved settings not edits', async () => {
  let saved = defaults(); const h = setup(async (_, options) => options.method === 'PUT' ? envelope(JSON.parse(options.body), true) : envelope(saved, false, 'offline'));
  await h.ui.open('remote'); assert.match(h.node('status').textContent, /pending disable does NOT mean/);
  h.node('offline_motd').value = 'unsaved draft'; h.node('form').oninput(); await h.node('retry').onclick();
  assert.equal(JSON.parse(h.calls[1].options.body).offline_motd, saved.offline_motd); assert.equal(h.node('offline_motd').value, 'unsaved draft');
  assert.match(h.node('status').textContent, /does not confirm that every port/);
});
test('HTTP error envelopes show saved desired state without erasing edited fields', async () => {
  const h = setup(async (_, options) => { if (options.method === 'PUT') throw Object.assign(Error('delivery failed'), { data: envelope({ ...defaults(), enabled: true }, false, 'delivery failed') }); return envelope(); });
  await h.ui.open('local'); h.node('offline_motd').value = 'keep me'; await h.submit();
  assert.equal(h.node('offline_motd').value, 'keep me'); assert.match(h.node('saved').textContent, /"enabled": true/); assert.match(h.node('status').textContent, /Pending/);
});
test('network error reports unknown status, preserves draft, and refresh does not overwrite it', async () => {
  const h = setup(async (_, options) => { if (options.method === 'PUT') throw Error('network down'); return envelope(); });
  await h.ui.open('local'); h.node('starting_message').value = 'edited'; await h.submit();
  assert.match(h.node('status').textContent, /unknown/); assert.equal(h.node('starting_message').value, 'edited');
  await h.node('refresh').onclick(); assert.equal(h.node('starting_message').value, 'edited');
});
test('failed initial load cannot save fabricated defaults, and unsupported workers get update guidance', async () => {
  const h = setup(async () => { throw Object.assign(Error('not found'), { status: 404 }); }); await h.ui.open('remote');
  assert.equal(h.node('save').disabled, true); assert.equal(h.node('fields').disabled, true); assert.match(h.node('error').textContent, /Update and restart/);
  await h.submit(); assert.equal(h.calls.length, 1);
});
test('rejects malformed envelopes instead of trusting applied truthiness', async () => {
  const h = setup(async () => ({ config: defaults(), applied: 'true', error: '' })); await h.ui.open('local');
  assert.equal(h.node('save').disabled, true); assert.match(h.node('error').textContent, /Unexpected/);
  assert.equal(h.ui.validEnvelope(envelope()), true);
});
test('validates UTF-8 byte limits, integer bounds, and literal bind IPs', () => {
  const { ui } = setup();
  for (const bind_address of ['', '0.0.0.0', '127.0.0.1', '::', '2001:db8::1', '::ffff:192.0.2.1']) assert.doesNotThrow(() => ui.validate({ ...defaults(), bind_address }));
  for (const bind_address of ['host.example', '127.0.0.1:25565', '999.2.3.4', '[::1]', '::zzz', '01.2.3.4']) assert.throws(() => ui.validate({ ...defaults(), bind_address }));
  for (const [key, limit] of [['offline_motd', 2048], ['version_label', 128], ['starting_message', 4096], ['unavailable_message', 4096]]) {
    assert.doesNotThrow(() => ui.validate({ ...defaults(), [key]: 'é'.repeat(limit / 2) })); assert.throws(() => ui.validate({ ...defaults(), [key]: 'é'.repeat(limit / 2) + 'a' }));
  }
  for (const [key, values] of [['max_players', [-1, 100001, 0.5, NaN]], ['cooldown_seconds', [9, 601, 10.5]]]) for (const value of values) assert.throws(() => ui.validate({ ...defaults(), [key]: value }));
  assert.doesNotThrow(() => ui.validate({ ...defaults(), max_players: 0, cooldown_seconds: 10 }));
});
test('validation failure preserves fields and sends no PUT', async () => {
  const h = setup(); await h.ui.open('local'); h.node('max_players').value = ''; h.node('offline_motd').value = 'still here'; await h.submit();
  assert.equal(h.calls.length, 1); assert.match(h.node('error').textContent, /integer/); assert.equal(h.node('offline_motd').value, 'still here');
});
test('icon preview only accepts bounded PNG data URLs; upload checks dimensions and removal is explicit', async () => {
  const h = setup(); await h.ui.open('local');
  for (const value of ['https://example.com/icon.png', 'data:image/svg+xml;base64,AAAA', 'data:image/png;base64,AAAA" onerror="x', 'data:image/png;base64,' + 'A'.repeat(65536)]) assert.equal(h.ui.safeIcon(value), false);
  const data = 'data:image/png;base64,AAAA';
  h.c.FileReader = class { readAsDataURL(file) { this.result = file.data; this.onload(); } };
  let size = 32; h.c.Image = class { set src(value) { this.naturalWidth = size; this.naturalHeight = size; this.onload(); } };
  const upload = () => h.node('upload').onchange({ target: { files: [{ size: 100, data }] } });
  await upload(); assert.match(h.node('error').textContent, /64×64/); assert.equal(h.node('preview').hidden, true);
  size = 64; await upload(); assert.equal(h.node('preview').src, data); assert.equal(h.node('preview').hidden, false);
  h.node('remove').onclick(); assert.equal(h.node('preview').hidden, true); await h.submit(); assert.equal(JSON.parse(h.calls[1].options.body).icon_data_url, '');
});
test('ordinary users cannot open settings or send requests', async () => {
  const h = setup(); h.c.S.user.role = 'user'; await h.ui.open('local'); assert.equal(h.calls.length, 0); assert.equal(h.container.innerHTML, '');
});
test('late responses cannot populate another node modal or a changed session', async () => {
  let resolve; const h = setup(url => url.includes('/first/') ? new Promise(r => { resolve = r; }) : Promise.resolve(envelope({ ...defaults(), offline_motd: 'second' })));
  const first = h.ui.open('first'); await h.ui.open('second'); resolve(envelope({ ...defaults(), offline_motd: 'stale' })); await first; assert.equal(h.node('offline_motd').value, 'second');
  const pending = h.ui.open('first'); h.c.S.token = 'other'; resolve(envelope()); await pending; assert.equal(h.node('save').disabled, true);
});

test('Local and remote edit dialogs contain wake settings, but connect-new does not', () => {
  const app = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
  const container = { innerHTML: '' };
  const local = { id: 'local', is_local: true, name: 'Local <host>' };
  const remote = { id: 'remote', name: 'Remote', base_url: 'https://node.example' };
  const c = { $: () => container, getNodeById: id => id === 'local' ? local : remote, esc: value => String(value ?? '').replaceAll('<', '&lt;').replaceAll('"', '&quot;'), displayNodeName: (_, name) => name, ic: () => '', setTimeout() {} };
  vm.runInNewContext(app.slice(app.indexOf('function renderNodeWakeSettings('), app.indexOf('async function testNodeModalConnection(')), c);
  c.openNodeModal('local');
  assert.match(container.innerHTML, /Edit Local Node: Local &lt;host>/);
  assert.match(container.innerHTML, /data-id="local" onclick="NPNodeWake.open\(this.dataset.id, true\)"/);
  assert.match(container.innerHTML, /Edit address/);
  assert.doesNotMatch(container.innerHTML, /submitNodeModal|nm-token/);
  c.openNodeModal('remote');
  assert.match(container.innerHTML, /data-id="remote" onclick="NPNodeWake.open\(this.dataset.id, true\)"/);
  assert.match(container.innerHTML, /submitNodeModal/);
  c.openNodeModal();
  assert.doesNotMatch(container.innerHTML, /NPNodeWake.open/);
});

test('Back restores the actual node form including unsaved values and handlers', async () => {
  const h = setup(); const form = { value: 'unsaved endpoint', onclick() {} };
  h.container.childNodes = [form];
  await h.ui.open('remote', true);
  assert.match(h.container.innerHTML, /Back to node settings/);
  h.node('close').onclick();
  assert.equal(h.container.childNodes[0], form);
  assert.equal(h.container.childNodes[0].value, 'unsaved endpoint');
});

test('Back cannot restore a stale node form after a session switch', async () => {
  const h = setup(); const form = { value: 'private draft' };
  h.container.childNodes = [form]; await h.ui.open('remote', true);
  h.c.S.token = 'new-session'; h.node('close').onclick();
  assert.notEqual(h.container.childNodes[0], form);
});

test('wrapped worker 405 retains saved state and gives upgrade, URL, proxy and retry guidance', async () => {
  const legacy = 'Node returned HTTP 405; configuration delivery pending';
  const h = setup(async (_, options) => {
    if (options.method === 'PUT') throw Object.assign(Error('Bad Gateway'), { status: 502, data: envelope(JSON.parse(options.body), false, legacy) });
    return envelope(defaults(), false, legacy);
  });
  await h.ui.open('remote');
  for (const pattern of [/HTTP 405/, /Update and restart/, /node URL/, /PUT \/api\/cluster\/wake/, /Retry saved settings/]) assert.match(h.node('error').textContent, pattern);
  h.node('enabled').checked = true; h.node('offline_motd').value = 'Keep this draft'; await h.submit();
  assert.match(h.node('error').textContent, /HTTP 405/); assert.doesNotMatch(h.node('error').textContent, /Bad Gateway/);
  assert.equal(h.node('panel-state').textContent, 'Enabled — saved');
  assert.equal(h.node('worker-state').textContent, 'Pending — not confirmed');
  assert.equal(h.node('offline_motd').value, 'Keep this draft');
});

test('existing actionable backend guidance is not duplicated', async () => {
  const error = 'Node returned HTTP 405 for PUT /api/cluster/wake. Update and restart the worker, then Retry saved settings';
  const h = setup(async () => envelope(defaults(), false, error)); await h.ui.open('remote');
  assert.equal(h.node('error').textContent, error);
});

test('separate dialog groups settings horizontally with per-field help and honest health status', async () => {
  const h = setup(); await h.ui.open('local'); const html = h.container.innerHTML;
  assert.match(html, /class="card pbox wake-dialog" role="dialog"/);
  assert.equal((html.match(/class="wake-section"/g) || []).length, 3);
  for (const label of ['Startup behavior', 'Offline appearance', 'Player messages', 'Not tested by this form', 'does not retry delivery or test ports', 'Any unauthenticated Java login']) assert.ok(html.includes(label), label);
  for (const key of Object.keys(defaults()).filter(key => key !== 'icon_data_url')) {
    assert.equal((html.match(new RegExp(`id="wake-${key}"`, 'g')) || []).length, 1);
    assert.ok(html.includes(`aria-describedby="wake-${key}-help"`), key);
  }
  assert.equal(h.node('worker-state').textContent, 'Stored on this host');
  const css = fs.readFileSync(path.join(__dirname, '../static/css/node-wake.css'), 'utf8');
  assert.match(css, /repeat\(3, minmax\(0, 1fr\)\)/);
  assert.match(css, /@media \(max-width: 700px\)/);
});
