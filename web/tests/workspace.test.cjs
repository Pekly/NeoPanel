const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const path = require('node:path');
const source = fs.readFileSync(path.join(__dirname, '../static/js/workspace.js'), 'utf8');
const app = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
const plain = value => JSON.parse(JSON.stringify(value));
const deferred = () => { let resolve; const promise = new Promise(r => { resolve = r; }); return { promise, resolve }; };
function storage() { const values = new Map(); return { values, getItem: k => values.get(k) ?? null, setItem: (k,v) => values.set(k,v), removeItem: k => values.delete(k) }; }
function setup(local = storage(), session = storage()) {
  const elements = {}, requests = [], notices = [], permissions = new Set(['console.view','console.command','server.power','files.view']);
  const server = { id: 'a', node_id: 'local', name: 'Alpha' };
  const c = vm.createContext({ console, AbortController, setTimeout, clearTimeout, setInterval() {},
    S: { user: { id: 'u1', role: 'user' }, token: 't1', servers: [server], cs: 0, view: 'con', status: 'online', wsState: 'connected', activeServerId: 'a', activeNodeId: 'local', hist: [], hi: 0, branding: {} },
    localStorage: local, sessionStorage: session, navigator: { onLine: true },
    document: { getElementById: id => elements[id] || null, querySelectorAll: selector => elements[selector] || [], hidden: false },
    location: { pathname: '/', search: '', hash: '' }, history: { replaceState(a,b,url) { c.location.pathname = url; } },
    esc: s => String(s ?? '').replaceAll('&','&amp;').replaceAll('<','&lt;').replaceAll('"','&quot;'),
    normalizeServerStatus: s => s || 'offline', hasServerPermission: p => permissions.has(p),
    canServerPage: () => true, parseCurrentRoute: () => ({ serverId: c.S.view === 'dash' ? null : c.curServer()?.id }),
    serverRouteUrl: (node,id,view) => `/server/${node}/${id}/${view}`,
    srvApiPath: (srv,suffix) => `/api/nodes/${srv.node_id}/proxy/api/servers/${srv.id}${suffix}`,
    api: async (url, options = {}) => { requests.push({url,options}); return {countdown:null}; },
    R() {}, toast: (text,type) => notices.push({text,type}), captureEditor() {},
    loadFileList: async dir => { c.S.currentDir=dir; return true; }, openQolFile: async () => {}, refreshServerStatus: async () => {}, refreshDashboard: async () => {}, connectWs() {}
  });
  c.window = c; c.addEventListener = () => {}; c.curServer = () => c.S.servers[c.S.cs];
  vm.runInContext(source, c); c.NPWorkspace.loadUser(); c.NPWorkspace.activate(server);
  return { c, q: c.NPWorkspace, elements, requests, notices, permissions, local, session, server };
}
function form(seconds = '60', message = '', action = 'restart') { return { reportValidity: () => true, elements: { seconds: {value:seconds}, message: {value:message}, action: {value:action} } }; }
function section(start,end) { return app.slice(app.indexOf(start), app.indexOf(end,app.indexOf(start))); }

test('server favourites preserve original indexes and sort without changing inventory', () => {
  const {c,q} = setup(); c.S.servers.push({ id:'b',node_id:'remote',name:'Zulu' });
  q.toggleFavourite(1); const rows = c.S.servers.map((s,origIdx) => ({s,origIdx}));
  assert.deepEqual(plain(q.sortServers(rows)).map(x => x.origIdx), [1,0]);
  assert.equal(c.S.servers[0].id,'a'); c.S.favouritesOnly=true;
  assert.deepEqual(plain(q.sortServers(rows)).map(x => x.s.id), ['b']);
});
test('list preferences persist per account, not globally', () => {
  const {c,q,local} = setup(); c.S.serverFilter='paper'; c.S.nodeFilter='remote'; c.S.serverSort='node'; c.S.favouritesOnly=true; q.saveList();
  const next=setup(local); assert.equal(next.c.S.serverFilter,'paper'); assert.equal(next.c.S.favouritesOnly,true);
  next.c.S.user.id='u2'; next.q.loadUser(); assert.equal(next.c.S.serverFilter,''); assert.equal(next.c.S.favouritesOnly,false);
});
test('Pinned only uses the highlighted style while selected, including restored preferences', () => {
  const {c,q,local}=setup();
  assert.match(q.listControls(), /class="btn sm" aria-pressed="false"/);
  c.S.favouritesOnly=true; q.saveList();
  assert.match(q.listControls(), /class="btn sm pr" aria-pressed="true"/);
  const restored=setup(local);
  assert.match(restored.q.listControls(), /class="btn sm pr" aria-pressed="true"/);
  restored.c.S.favouritesOnly=false; restored.q.saveList();
  assert.match(restored.q.listControls(), /class="btn sm" aria-pressed="false"/);
  assert.doesNotMatch(restored.q.listControls(), /class="btn sm pr"/);
});
test('command history and filters are bounded and scoped by account, node and server', () => {
  const {c,q,server,session}=setup();
  for(let i=0;i<110;i++) q.recordCommand('say '+i);
  q.recordCommand('say 109'); q.setFilter('WARN'); assert.equal(c.S.hist.length,100);
  const remote={...server,node_id:'remote'}; c.S.servers=[remote]; q.activate(remote); assert.equal(c.S.hist.length,0);
  c.S.servers=[server]; q.activate(server); assert.equal(c.S.hist.length,100); assert.match(q.consoleTools(),/value="WARN"/);
  c.S.user.id='u2'; q.loadUser(); q.activate(server); assert.equal(c.S.hist.length,0);
  assert.ok([...session.values.keys()].every(k => k.includes('u1')));
});
test('clear history is local to the selected server', () => { const {c,q}=setup(); q.recordCommand('list'); q.clearHistory(); assert.equal(c.S.hist.length,0); assert.equal(c.S.hi,0); });
test('workspace restores known server tab only at root, not an explicit deep link', () => {
  const {c,q,local}=setup(); q.restoreRoute(); q.remember();
  let next=setup(local); next.q.restoreRoute(); assert.equal(next.c.location.pathname,'/server/local/a/con');
  next=setup(local); next.c.location.pathname='/account'; next.q.restoreRoute(); assert.equal(next.c.location.pathname,'/account');
});
test('removed servers and unsafe stored routes are never restored', () => {
  const {local}=setup(); local.setItem('neopanel:workspace:u1',JSON.stringify({route:{id:'missing',node:'local',view:'con'}}));
  let state=setup(local); state.q.restoreRoute(); assert.equal(state.c.location.pathname,'/');
  local.setItem('neopanel:workspace:u1',JSON.stringify({route:{id:'a',node:'local',view:'https://evil.example'}}));
  state=setup(local); state.q.restoreRoute(); assert.equal(state.c.location.pathname,'/');
});
test('last open file restores only after file access is verified', async () => {
  const {c,q,local}=setup(); q.restoreRoute(); c.S.view='fi'; c.S.currentDir='plugins'; c.S.editingFile='plugins/config.yml'; q.remember();
  const next=setup(local); next.c.S.view='fi'; let opened=''; next.c.openQolFile=async name => {opened=name;};
  await next.q.restoreFile(); assert.equal(opened,'config.yml'); assert.equal(next.c.S.currentDir,'plugins');
  opened=''; next.permissions.delete('files.view'); await next.q.restoreFile(); assert.equal(opened,'');
});
test('late file restoration cannot open a file on another server', async () => {
  const {c,q,local}=setup(); q.restoreRoute(); c.S.view='fi'; c.S.currentDir=''; c.S.editingFile='server.properties'; q.remember();
  const next=setup(local); next.c.S.view='fi'; const d=deferred(); let opens=0;
  next.c.loadFileList=()=>d.promise; next.c.openQolFile=async()=>opens++;
  const pending=next.q.restoreFile(); next.c.S.servers=[{id:'b',node_id:'local'}]; d.resolve(); await pending; assert.equal(opens,0);
});
test('storage corruption and quota errors do not break the panel', () => {
  const local=storage(); local.setItem('neopanel:workspace:u1','[]'); const {q}=setup(local);
  local.setItem=()=>{throw Error('quota');}; assert.doesNotThrow(()=>q.toggleFavourite(0)); assert.doesNotThrow(()=>q.saveList());
});
test('console filter operates on text, preserves paused scroll across redraw, and jumps on request', () => {
  const {q,elements}=setup(); const listeners={};
  const el={ dataset:{consoleBound:'1'}, children:[{textContent:'INFO ready'},{textContent:'WARN slow'}], scrollHeight:1000, scrollTop:100, clientHeight:200, addEventListener:(n,f)=>listeners[n]=f };
  elements.log=el; q.captureConsole(); q.setFilter('warn'); q.bindLog(el);
  assert.equal(el.children[0].hidden,true); assert.equal(el.children[1].hidden,false); assert.equal(el.scrollTop,100);
  q.jump('log'); assert.equal(el.scrollTop,1000);
});
test('disconnected controls are visibly disabled without treating the process as offline', () => {
  const {c,q,elements}=setup(); const button={dataset:{powerBtn:'restart'},setAttribute(){}}; elements['[data-power-btn]']=[button];
  c.S.wsState='reconnecting'; q.syncControls(); assert.equal(button.disabled,true); assert.equal(c.S.status,'online');
  c.S.wsState='connected'; q.syncControls(); assert.equal(button.disabled,true);
  q.requestResult(c.srvApiPath(c.curServer(),''),null,c.S.token); assert.equal(button.disabled,false);
  c.navigator.onLine=false; q.syncControls(); assert.equal(button.disabled,true);
});
test('API connection failures need a fresh status response and ignore other servers or sessions', () => {
  const {c,q,server}=setup(); const base=c.srvApiPath(server,'');
  q.requestResult(base,new Error('network'),'old-token'); assert.equal(q.connectionBlocked(),false);
  q.requestResult(base,new Error('network'),c.S.token); assert.equal(q.connectionBlocked(),true);
  q.requestResult(base+'/power/countdown',null,c.S.token); assert.equal(q.connectionBlocked(),true);
  q.requestResult(base,null,c.S.token); assert.equal(q.connectionBlocked(),false);
});
test('maintenance keeps permitted console and power routes, not file-management routes', () => {
  const {c,q,permissions}=setup(); c.S.branding.maintenance_enabled=true;
  vm.runInContext(section('function serverPagePermission(', 'function canCreateServers('),c);
  // Use the real page-policy function with an established permission snapshot.
  c.S.serverAccess={key:'id',status:'ready',data:{is_admin:false,is_owner:false,permissions:['server.power']}}; c.accessIdentity=()=> 'id';
  assert.equal(c.canServerPage('con'),true); assert.equal(c.canServerPage('fi'),false);
  c.S.serverAccess.data.permissions=[]; assert.equal(c.canServerPage('con'),false);
  assert.equal(q.maintenanceAllowed('dash'),true); assert.match(q.maintenanceBanner(),/existing permissions/);
  c.S.user.role='admin'; assert.equal(q.maintenanceLimited(),false);
});
test('maintenance sign-in is available to normal users and does not bypass auth', () => {
  const login=section('function showLoginScreen()', 'async function submitLogin(');
  assert.match(login,/Sign in to control your server/); assert.doesNotMatch(login,/Administrator Sign In|ADMIN LOGIN/);
  assert.match(login,/submitLogin\(\)/); assert.match(login,/renderPasskeySignIn\(\)/);
});
test('countdown validates input and requires power plus console-command access', async () => {
  const {q,requests,permissions}=setup();
  for(const f of [form('9'),form('3601'),form('1.5'),form('60','hello\nstop'),form('60','','kill')]) await q.scheduleCountdown(f);
  assert.equal(requests.length,0); permissions.delete('console.command'); await q.scheduleCountdown(form()); assert.equal(requests.length,0);
});
test('countdown posts to original server, blocks duplicates and can be cancelled', async () => {
  const {c,q,requests}=setup(); const d=deferred(); let current=null;
  c.api=async(url,options={})=>{requests.push({url,options}); if(options.method==='POST') {await d.promise; current={action:'restart',execute_at:new Date(Date.now()+60000).toISOString(),message:'Maintenance'};} if(options.method==='DELETE') current=null; return {countdown:current};};
  const pending=q.scheduleCountdown(form()); await q.scheduleCountdown(form()); assert.equal(requests.length,1); assert.equal(q.powerBlocked(),true);
  d.resolve(); await pending; assert.equal(q.powerBlocked(),true); assert.equal(JSON.parse(requests[0].options.body).seconds,60);
  await q.cancelCountdown(); assert.equal(q.powerBlocked(),false); assert.equal(requests.filter(r=>r.options.method==='DELETE').length,1);
});
test('a stale countdown poll cannot overwrite a newer successful cancellation', async () => {
  const {c,q}=setup(); const d=deferred();
  c.api=async(url,options={})=>options.method==='DELETE'?{countdown:null}:d.promise;
  const polling=q.pollCountdown(); await q.cancelCountdown(); d.resolve({countdown:{action:'restart',execute_at:new Date().toISOString()}}); await polling;
  assert.equal(q.powerBlocked(),false);
});
test('late countdown results do not populate a different account', async () => {
  const {c,q}=setup(); const d=deferred(); c.api=()=>d.promise;
  const polling=q.pollCountdown(); c.S.token='t2'; c.S.user.id='u2'; q.loadUser(); d.resolve({countdown:{action:'stop'}}); await polling;
  assert.equal(q.powerBlocked(),false);
});
test('countdown errors remain visible across timer paints', async () => {
  const {c,q,elements}=setup(); const el={textContent:'',closest:()=>null}; elements['np-countdown-state']=el;
  c.api=async()=>{throw Object.assign(Error('not found'),{status:404});};
  await q.pollCountdown(); q.paintCountdown(); assert.match(el.textContent,/Update and restart/);
});
test('actionable errors distinguish permission, connectivity, conflicts and Java failures', () => {
  const {q}=setup();
  assert.match(q.errorAdvice({status:403,message:'denied'}),/permissions/);
  assert.match(q.errorAdvice({status:503,message:'unavailable'}),/may already/);
  assert.match(q.errorAdvice({status:409,message:'pending'}),/cancel its countdown/);
  assert.match(q.errorAdvice({message:'Java missing'}),/Startup/);
});
test('player address copy is explicit and still stops card navigation', () => {
  const c=vm.createContext({copyDbText:(address,label)=>{c.copied={address,label};}});
  vm.runInContext(section('function copyServerAddress(', '// --- Sparklines'),c);
  let prevented=0; c.copyServerAddress('game.example:25566',{preventDefault(){prevented++;},stopPropagation(){prevented++;}});
  assert.equal(c.copied.address,'game.example:25566'); assert.match(c.copied.label,/Player/); assert.equal(prevented,2);
});
test('existing editor protection remains integrated on navigation and unload', () => {
  assert.match(section('function applyRoute(', 'function navigateToUrl('),/guardEditorLeave/);
  const qol=fs.readFileSync(path.join(__dirname,'../static/js/qol.js'),'utf8');
  assert.match(qol,/sessionStorage\.setItem/); assert.match(qol,/Recover file draft/); assert.match(qol,/beforeunload/);
});

test('background redraw cannot erase the startup file restoration snapshot', async () => {
  const {c,q,local}=setup(); q.restoreRoute(); c.S.view='fi'; c.S.currentDir='plugins'; c.S.editingFile='plugins/config.yml'; q.remember();
  const next=setup(local); next.q.restoreRoute(); next.c.S.view='fi'; next.c.S.currentDir=''; next.q.remember();
  let opened=''; next.c.openQolFile=async name=>{opened=name;};
  await next.q.restoreFile(); assert.equal(opened,'config.yml');
});

test('failed directory restoration retains the loaded directory and does not open a file', async () => {
  const {c,q,local}=setup(); q.restoreRoute(); c.S.view='fi'; c.S.currentDir='removed'; c.S.editingFile='removed/config.yml'; q.remember();
  const next=setup(local); next.c.S.view='fi'; next.c.S.currentDir=''; let opens=0;
  next.c.loadFileList=async()=>false; next.c.openQolFile=async()=>opens++;
  await next.q.restoreFile(); assert.equal(next.c.S.currentDir,''); assert.equal(opens,0);
});
test('countdown polling is paused while a mutation is in flight', async () => {
  const {c,q}=setup(); const d=deferred(); let gets=0;
  c.api=async(url,options={})=>{if(options.method==='POST')return d.promise;gets++;return {countdown:{action:'restart',execute_at:new Date(Date.now()+60000).toISOString()}};};
  const scheduling=q.scheduleCountdown(form()); await q.pollCountdown(); assert.equal(gets,0);
  d.resolve({countdown:{action:'restart',execute_at:new Date(Date.now()+60000).toISOString()}});await scheduling;
  assert.equal(gets,1); assert.equal(q.powerBlocked(),true);
});
test('removed location filters reset only after authoritative inventory loading', () => {
  const {c,q}=setup(); c.S.nodeFilter='removed'; c.S.nodes=[{id:'local'}]; q.restoreRoute(); assert.equal(c.S.nodeFilter,'removed');
  c.S.serverCatalog={clusterComplete:true}; q.restoreRoute(); assert.equal(c.S.nodeFilter,'all');
  assert.match(app,/const activeNodeFilter = locations.selected/);
});

test('intentional request cancellation is not reported as a lost connection', () => {
  const {c,q,server}=setup();
  q.requestResult(c.srvApiPath(server,'/files'),Object.assign(Error('cancelled'),{name:'AbortError'}),c.S.token);
  assert.equal(q.connectionBlocked(),false);
});

test('unsupported countdown APIs stop automatic polling without disabling normal power', async () => {
  for (const status of [404,405,501]) {
    const {c,q,elements}=setup(); let calls=0;
    const submit={disabled:false}, retry={hidden:true,disabled:false};
    const el={textContent:'',closest:()=>({querySelectorAll:()=>[submit]})};
    elements['np-countdown-state']=el; elements['np-countdown-retry']=retry;
    c.api=async()=>{calls++;throw Object.assign(Error('unsupported'),{status});};
    await q.pollCountdown(); await q.pollCountdown();
    assert.equal(calls,1); assert.match(el.textContent,new RegExp(`HTTP ${status}`));
    assert.match(el.textContent,/Update and restart/); assert.doesNotMatch(el.textContent,/Reconnect/);
    assert.equal(q.powerBlocked(),false); assert.equal(submit.disabled,true); assert.equal(retry.hidden,false); assert.equal(retry.disabled,false);
    await q.scheduleCountdown(form()); assert.equal(calls,1);
    c.api=async()=>{calls++;return {countdown:null};}; await q.pollCountdown(true);
    assert.equal(calls,2); assert.equal(submit.disabled,false); assert.equal(retry.hidden,true);
  }
});
test('permission errors are actionable and do not cause a reconnect loop', async () => {
  const {c,q,elements}=setup(); const el={textContent:'',closest:()=>null}; elements['np-countdown-state']=el;
  let calls=0;c.api=async()=>{calls++;throw Object.assign(Error('denied'),{status:403});};
  await q.pollCountdown();await q.pollCountdown();assert.equal(calls,1);
  assert.match(el.textContent,/access denied.*HTTP 403/);assert.match(el.textContent,/permissions/);assert.equal(q.connectionBlocked(),false);
});
test('malformed countdown responses do not masquerade as no pending action', async () => {
  for(const response of [{},[],null,{countdown:{action:'restart',execute_at:'invalid'}},{countdown:{action:'kill',execute_at:new Date().toISOString()}}]) {
    const {c,q,elements}=setup(); const el={textContent:'',closest:()=>null}; elements['np-countdown-state']=el;
    c.api=async()=>response;await q.pollCountdown();assert.match(el.textContent,/incompatible response/);
    assert.doesNotMatch(el.textContent,/No scheduled/);
  }
});
test('temporary countdown errors retain the cause and back off before retrying', async () => {
  const {c,q,elements}=setup();const el={textContent:'',closest:()=>null},retry={};elements['np-countdown-state']=el;elements['np-countdown-retry']=retry;
  let now=1000,calls=0;c.Date={now:()=>now,parse:Date.parse};
  c.api=async()=>{calls++;throw Object.assign(Error('worker unavailable'),{status:503});};
  await q.pollCountdown();await q.pollCountdown(true);assert.equal(calls,1);assert.equal(retry.disabled,true);
  assert.match(el.textContent,/HTTP 503.*worker unavailable/);
  now+=15001;c.api=async()=>({countdown:null});await q.pollCountdown();assert.equal(retry.hidden,true);
});
test('incompatible scheduling responses are not announced as successful', async () => {
  const {c,q,notices,elements}=setup();const el={textContent:'',closest:()=>null};elements['np-countdown-state']=el;
  c.api=async()=>({});await q.scheduleCountdown(form());
  assert.ok(notices.every(n=>!n.text.includes('Players warned')));assert.match(el.textContent,/incompatible response/);
});
test('console tools and power countdown are collapsed dropdowns', () => {
  const {q} = setup();
  assert.match(q.consoleTools(), /<details class="np-console-options np-compact-menu">/);
  assert.match(q.consoleTools(), /<summary class="btn sm">Console tools/);
  assert.doesNotMatch(q.consoleTools(), /<details[^>]*\sopen/);
  assert.match(q.countdownSection(), /<summary class="btn"/);
  assert.doesNotMatch(q.countdownSection(), /class="card np-countdown"/);
  assert.match(section('function srvh()', 'function toggleCmdMenu()'), /countdownSection/);
  assert.doesNotMatch(section('  con() {', '  pl() {'), /countdownSection/);
});
test('countdown label remains visible while scheduling controls are collapsed', async () => {
  const {q,c,elements}=setup();
  elements['np-countdown-label']={textContent:''};
  elements['np-countdown-state']={closest:()=>null};
  c.api=async()=>({countdown:{action:'restart',execute_at:new Date(Date.now()+60000).toISOString()}});
  await q.pollCountdown();
  assert.match(elements['np-countdown-label'].textContent,/Restart · \d+s/);
});
test('Schedule and Console tools use chevrons driven by the native open state', () => {
  const {q}=setup();
  for (const html of [q.consoleTools(), q.consoleTools('fly-log'), q.countdownSection()]) {
    assert.match(html, /class="np-menu-chevron" aria-hidden="true"/);
    assert.doesNotMatch(html, / ▾<\/summary>/);
  }
  const css=fs.readFileSync(path.join(__dirname,'../static/css/workspace.css'),'utf8');
  assert.match(css, /\.np-compact-menu\[open\] > summary \.np-menu-chevron\s*\{\s*transform: rotate\(225deg\)/);
  assert.match(css, /\.fly-con-head \.np-console-options\s*\{\s*position: static/);
  assert.match(css, /\.fly-con-head \.np-console-options \.np-menu-panel\s*\{[^}]*left: 12px;[^}]*max-width: calc\(100% - 24px\)/);
});
