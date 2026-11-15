const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
function setup() {
  const c = vm.createContext({ S: { view: 'admin', servers: [], nodes: [], system: { version: '1.0.0', os: 'linux', arch: 'amd64', cpus: 2, total_ram_mb: 4096, free_ram_mb: 2048 }, adminUpdate: { update_available: false } }, window: { location: { hostname: 'panel.test' } }, DEFAULT_BRANDING: {}, Date,
    setBtnLoading: (button, busy) => { button.disabled = busy; }, loadDashboard: async () => {}, loadAdminData: async () => {}, R: () => {},
    esc: value => String(value ?? ''), ic: () => '', normalizeServerStatus: value => value === 'running' ? 'online' : value,
    renderAdminUpdateBtn: () => '', renderJavaAvailabilityStrip: () => '' });
  const start = source.indexOf('const ADMIN_TAB_SLUGS'), end = source.indexOf('const ACC_TAB_SLUGS', start);
  vm.runInContext(source.slice(start, end) + '\nthis.sections=ADMIN_SECTIONS;this.slugs=ADMIN_TAB_SLUGS;this.refreshing=()=>adminStatusRefreshing;', c);
  return c;
}
test('admin navigation has one canonical label per preserved route', () => {
  const c=setup();
  assert.deepEqual(Array.from(c.sections, s=>s.label), ['Overview','Nodes','Servers','Users','Appearance','Audit logs','Extensions','Email & SMTP']);
  assert.deepEqual(Array.from(c.slugs), ['health','nodes','servers','users','branding','audit','mods','email']);
  assert.equal(c.sections.length,c.slugs.length);
  assert.match(source,/ADMIN_SECTIONS\.map\(\(section, index\)/);
});
test('admin overview replaces duplicate tabs and unconditional health claim with useful status', () => {
  const c=setup();
  const start=source.indexOf('  admin() {'), end=source.indexOf('  acc() {',start);
  const body=source.slice(start+'  admin() '.length,end).trim().replace(/,$/,'');
  vm.runInContext('this.renderAdmin=function() '+body,c);
  const html=c.renderAdmin();
  assert.match(html,/Panel status summary/);assert.match(html,/Servers running/);assert.match(html,/Nodes online/);assert.match(html,/Up to date/);
  assert.doesNotMatch(html,/class="tabs"|System Healthy|Daemon Administration Panel|NaN/);
});
test('local node label is human-readable without renaming custom nodes', () => {
  const c=setup();
  assert.equal(c.displayNodeName('local','Local Master Node'),'Local node');
  assert.equal(c.displayNodeName('local','Local Master'),'Local node');
  assert.equal(c.displayNodeName('local','My custom machine'),'My custom machine');
  assert.equal(c.displayNodeName('remote','Local Master Node'),'Local Master Node');
});
test('refresh status prevents duplicate requests and resets its busy state', async () => {
  const c=setup();let finish, loads=0,renders=0;
  c.loadAdminData=()=>{loads++;return new Promise(resolve=>{finish=resolve;});};c.R=()=>{renders++;};
  const button={disabled:false,isConnected:true};const first=c.refreshAdminStatus(button);await c.refreshAdminStatus(button);
  assert.equal(loads,1);assert.equal(c.refreshing(),true);finish();await first;
  assert.equal(c.refreshing(),false);assert.equal(renders,1);assert.ok(c.S.adminRefreshedAt);
});
