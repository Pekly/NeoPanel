const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
function extract(name) {
  const start = source.indexOf(`async function ${name}(`);
  const end = source.indexOf('\n}', start);
  assert.ok(start >= 0 && end > start);
  return source.slice(start, end + 2);
}
function setup(api) {
  const events = [], button = { disabled: false, textContent: '' };
  const server = { id: 'same-id', node_id: 'east', name: 'Keep', port: 25565, type: 'paper' };
  const ws = { close: () => events.push('close socket') };
  const c = vm.createContext({
    S: { activeServerId: server.id, activeNodeId: server.node_id, activeWs: ws },
    $: id => id.includes('submit') || id === '#del-srv-btn' ? button : id === '#adm-set-node' ? { value: 'west' } : null,
    api, nodeApiPath: (node, suffix) => `/node/${node}${suffix}`,
    clearTimeout: () => {}, clearInterval: () => {},
    toast: (message, level) => events.push([message, level]), closeModal: () => events.push('close modal'),
    getSavedActiveServerRef: () => ({ id: server.id, nodeId: 'east' }),
    localStorage: { removeItem: () => events.push('forget selection') },
    loadDashboard: async () => events.push('refresh'), applyRoute: route => events.push(route),
    findServerRef: () => server, curServer: () => server, getServerMaxRamMB: () => 2048,
    isSteelEngine: () => false, R: () => {},
  });
  vm.runInContext(extract('executeDeleteServer') + '\n' + extract('submitAdminServerSettings'), c);
  return { c, events, button, server, ws };
}

test('failed remote delete retains active server and console until confirmation', async () => {
  let reject;
  const { c, events, ws, button } = setup(() => new Promise((_, fail) => { reject = fail; }));
  const pending = c.executeDeleteServer('same-id', 'east');
  assert.equal(c.S.activeServerId, 'same-id'); assert.equal(c.S.activeWs, ws);
  assert.deepEqual(events, []);
  reject(Error('process cleanup not confirmed'));
  await pending;
  assert.equal(c.S.activeServerId, 'same-id'); assert.equal(c.S.activeWs, ws);
  assert.equal(button.disabled, false);
  assert.deepEqual(events, [['process cleanup not confirmed', 'bad']]);
});

test('empty, legacy and wrong-server responses cannot claim completed deletion', async () => {
  for (const reply of [{}, { message: 'Server deleted successfully' }, { deleted: true, server_id: 'other-id' }, { deleted: false, server_id: 'same-id' }]) {
    const { c, events } = setup(async () => reply);
    await c.executeDeleteServer('same-id', 'east');
    assert.equal(c.S.activeServerId, 'same-id');
    assert.equal(events.length, 1); assert.match(events[0][0], /not confirmed/);
    assert.equal(events[0][1], 'bad');
  }
});

test('only hosting-node confirmation closes console and removes selection', async () => {
  const calls = [];
  const { c, events } = setup(async (url, opts) => { calls.push([url, opts.method]); return { deleted: true, server_id: 'same-id' }; });
  await c.executeDeleteServer('same-id', 'east');
  assert.deepEqual(calls, [['/node/east/api/servers/same-id', 'DELETE']]);
  assert.equal(c.S.activeServerId, null); assert.equal(c.S.activeWs, null);
  assert.ok(events.includes('close socket')); assert.ok(events.includes('forget selection')); assert.ok(events.includes('/'));
});

test('admin settings cannot reroute later deletion through a fake node move', async () => {
  let captured;
  const { c, server } = setup(async (url, opts) => { captured = [url, JSON.parse(opts.body)]; return {}; });
  await c.submitAdminServerSettings(null, server.id, 'east');
  assert.equal(captured[0], '/node/east/api/servers/same-id');
  assert.equal(captured[1].node_id, undefined); assert.equal(server.node_id, 'east');
  assert.match(source, /id="adm-set-node"[^\n]+disabled/);
});
