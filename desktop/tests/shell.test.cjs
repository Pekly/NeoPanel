const {test} = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const path = require('node:path');
const root = path.join(__dirname, '../internal/desktop/shell');
function setup() {
  const elements = new Map(), events = {}, calls = [];
  const get = id => {
    if (!elements.has(id)) elements.set(id, {hidden: true, handlers: {}, attrs: {}, addEventListener(n,f){this.handlers[n]=f;}, setAttribute(n,v){this.attrs[n]=v;}});
    return elements.get(id);
  };
  const runtime = new Proxy({}, {get: (_,name) => (...args) => { calls.push([name,...args]); return name === 'WindowIsMaximised' ? Promise.resolve(true) : undefined; }});
  const window = {runtime,addEventListener(n,f){events[n]=f;}};
  vm.runInNewContext(fs.readFileSync(path.join(root,'shell.js'),'utf8'), {window,document:{getElementById:get},URL,setTimeout,clearTimeout});
  return {get,window,calls,events};
}
test('only loopback panel can be loaded; no frame-to-native message relay', () => {
  const {window,get,events} = setup();
  for (const url of ['https://evil.example','http://wails.localhost','http://127.0.0.1','http://user:pass@127.0.0.1:1234']) window.desktopShell.init(url);
  assert.equal(get('panel').src, undefined);
  window.desktopShell.init('http://127.0.0.1:1234');
  assert.equal(get('panel').src,'http://127.0.0.1:1234/');
  assert.equal(events.message,undefined);
});
test('toolbar uses native window actions; closing disables repeat operations', async () => {
  const {window,get,calls} = setup();
  window.desktopShell.init('http://127.0.0.1:1234');
  await new Promise(resolve => setImmediate(resolve));
  assert.equal(get('maximize').attrs['aria-label'],'Restore window');
  for (const id of ['minimize','maximize','data','close']) get(id).handlers.click();
  assert.deepEqual(calls.slice(1),[['WindowMinimise'],['WindowToggleMaximise'],['EventsEmit','desktop:data'],['Quit']]);
  get('panel').handlers.load();
  assert.equal(get('loading').hidden,true);
  window.desktopShell.stopping();
  assert.equal(get('shutdown').hidden,false);
  const count=calls.length;
  for (const id of ['home','reload','data','close']) { assert.equal(get(id).disabled,true); get(id).handlers.click(); }
  assert.equal(calls.length,count);
});
test('shell restricts navigation without disabling scripts, downloads or forms', () => {
  const html=fs.readFileSync(path.join(root,'index.html'),'utf8');
  const sandbox=html.match(/sandbox="([^"]+)"/)[1];
  for (const token of ['allow-scripts','allow-same-origin','allow-forms','allow-downloads']) assert.ok(sandbox.includes(token));
  assert.ok(!sandbox.includes('allow-top-navigation'));
  assert.ok(!sandbox.includes('allow-popups-to-escape-sandbox'));
  assert.ok(html.includes("object-src 'none'"));
});
