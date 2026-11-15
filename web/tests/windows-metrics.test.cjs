const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
function slice(start, end) {
  const a = source.indexOf(start), b = source.indexOf(end, a);
  assert.ok(a >= 0 && b > a);
  return source.slice(a, b);
}
function setup() {
  const srv = {port: 25565, stats: {}};
  const elements = {'#ch-tip-net': {style: {display: 'block'}}, '#ch-hover-svg-net': {innerHTML: 'stale'}};
  const c = vm.createContext({
    S: {cpu: [0, 0], ram: [0, 0], ni: [0, 0], no: [0, 0]},
    curServer: () => srv, getServerMaxRamMB: () => 1024, window: {},
    toast: () => {}, srvHost: () => 'localhost', esc: String, ic: () => '', up: () => '1m', formatDiskMB: String,
    $: id => elements[id], document: {getElementById: id => elements['#' + id]}
  });
  vm.runInContext([
    slice('function normalizeServerStatus(', 'function copyServerAddress('),
    slice('function chartSeriesSpec(', 'function updateChartHoverDom('),
    slice('function updateChartHoverDom(', 'function onChartMouseMove('),
    slice('function statsHtml()', 'function liveCon()')
  ].join('\n'), c);
  return {c, srv, elements};
}

test('Windows network capability survives snapshots and suppresses misleading charts', () => {
  const {c, srv, elements} = setup();
  c.applyServerStatsSnapshot(srv, {status: 'online', memory_mb: 53, cpu_usage: 3, network_supported: false}, true);
  assert.equal(srv.stats.network_supported, false);
  assert.equal(c.buildChartSeriesInnerSvg('net'), '');
  assert.match(c.buildChartSeriesInnerSvg('cpu'), /polyline/);
  assert.match(c.statsHtml(), /53 MiB/);
  assert.equal((c.statsHtml().match(/Unavailable/g) || []).length, 2);
  assert.match(c.chartsHtml(), /Unavailable on this host/);
  c.updateChartHoverDom('net');
  assert.equal(elements['#ch-tip-net'].style.display, 'none');
  assert.equal(elements['#ch-hover-svg-net'].innerHTML, '');
  c.applyServerStatsSnapshot(srv, {status: 'offline', network_supported: false});
  assert.match(c.statsHtml(), /Unavailable/);
});

test('supported and legacy nodes retain network charts without stale Windows capability', () => {
  const {c, srv} = setup();
  c.applyServerStatsSnapshot(srv, {status: 'online', network_supported: false});
  for (const capability of [true, undefined]) {
    c.applyServerStatsSnapshot(srv, {status: 'online', network_supported: capability, net_in_kib: 12.5});
    assert.match(c.buildChartSeriesInnerSvg('net'), /polyline/);
    assert.match(c.statsHtml(), /12.5 KiB\/s/);
    assert.doesNotMatch(c.chartsHtml(), /Unavailable/);
  }
});

test('node settings expose Windows-compatible process management', () => {
  assert.match(source, /option value="auto"[^\n]+Windows: Job Object, no filesystem jail/);
  assert.match(source, /option value="none"[^\n]+Windows Job Object/);
});
