const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const source = fs.readFileSync(require('node:path').join(__dirname, '../static/js/app.js'), 'utf8');
const slice = (a,b) => source.slice(source.indexOf(a), source.indexOf(b,source.indexOf(a)));
function setup() {
  const calls=[], storage=new Map();
  const c=vm.createContext({URL, AbortController, console, setTimeout, clearTimeout, clearInterval(){},
    S:{token:'token-a',user:{id:1,role:'user'},view:'fi'}, V:{}, window:{location:{pathname:'/server/a/files'}},
    server:{id:'a',node_id:'local'}, curServer:()=>c.server, srvApiPath:(s,p='')=>`/api/servers/${s.id}${p}`,
    esc:v=>String(v??'').replaceAll('&','&amp;').replaceAll('<','&lt;').replaceAll('"','&quot;'), ic:()=>'',
    localStorage:{getItem:k=>storage.get(k),setItem:(k,v)=>storage.set(k,v)}, R(){},$:()=>null,
    api:async (url,options={})=>{calls.push([url,options]);return {is_admin:false,is_owner:false,permissions:['files.view']};},
    toast(){}, setBtnLoading:(b,v)=>{b.disabled=v;}, closeModal(){}, uiConfirm:async()=>true,
  });
  vm.runInContext(source.slice(source.indexOf('// --- Permission-aware workspace')),c);
  return {c,calls,storage};
}
const deferred=()=>{let resolve;const promise=new Promise(r=>resolve=r);return {promise,resolve};};
const perms=(permissions=[],other={})=>({is_admin:false,is_owner:false,permissions,...other});
test('permissions fail closed before, during, and after a failed request; retry is explicit',async()=>{
  const {c}=setup();const wait=deferred();c.api=()=>wait.promise;
  assert.equal(c.canServerPage('fi'),false);const pending=c.ensureServerAccess();assert.equal(c.S.serverAccess.status,'loading');
  assert.equal(c.canServerPage('fi'),false);wait.resolve(perms(['files.view']));await pending;
  assert.equal(c.canServerPage('fi'),true);assert.equal(c.canServerPage('plug'),true);assert.equal(c.canServerPage('set'),false);
  c.api=async()=>{throw Error('offline');};await c.ensureServerAccess(true);assert.equal(c.canServerPage('fi'),false);assert.equal(c.S.serverAccess.status,'error');
  assert.match(c.renderAccessRestricted(),/Access restricted/);
});
test('late server and account permission responses cannot unlock current workspace',async()=>{
  const {c}=setup();const first=deferred(),second=deferred();c.api=()=>first.promise;const a=c.ensureServerAccess();
  c.server={id:'b',node_id:'local'};c.api=()=>second.promise;const b=c.ensureServerAccess();
  first.resolve(perms([], {is_owner:true}));await a;assert.equal(c.canServerPage('fi'),false);
  c.S.user={id:2,role:'user'};c.S.token='token-b';second.resolve(perms([],{is_admin:true}));await b;
  assert.equal(c.canServerPage('fi'),false);assert.equal(c.S.serverAccess.status,'loading');
});
test('confirmed owner and admin bypass while custom panel role never grants server access',async()=>{
  for(const flag of ['is_admin','is_owner']){const {c}=setup();c.api=async()=>perms([],{[flag]:true});await c.ensureServerAccess();assert.equal(c.canServerPage('team'),true);assert.equal(c.canServerPage('sftp'),true);}
  const {c}=setup();c.S.user.panel_permissions=['servers.create'];await c.ensureServerAccess();assert.equal(c.canCreateServers(),true);assert.equal(c.canServerPage('team'),false);
});
test('page permission mappings include SFTP separately from settings and files',()=>{
  const {c}=setup();for(const [page,p] of Object.entries({con:'console.view',pl:'console.view',fi:'files.view',plug:'files.view',bk:'backup.manage',db:'database.manage',sc:'schedules.manage',team:'subusers.manage',set:'settings.view',net:'settings.view',ver:'settings.view',sub:'settings.view',sftp:'sftp.access'}))assert.equal(c.serverPagePermission(page),p);
});
test('disabled navigation never follows a click, including modified clicks',()=>{
  const {c}=setup();vm.runInContext(slice('function handleNavClick(', 'function handleServerCardClick('),c);
  let prevented=0;assert.equal(c.handleNavClick({ctrlKey:true,preventDefault(){prevented++;}},'fi'),false);assert.equal(prevented,1);
});
test('forbidden initial page has no background loader calls',async()=>{
  const {c}=setup();vm.runInContext(slice('async function loadInitialRouteData(', 'window.onEnterView'),c);
  let loaded=0;c.loadFileList=()=>{loaded++;};c.api=async()=>perms([]);await c.loadInitialRouteData('fi');assert.equal(loaded,0);
});
test('file operations require view and the appropriate write permission, including copies',async()=>{
  const {c}=setup();await c.ensureServerAccess();
  assert.doesNotThrow(()=>c.assertClientRequestAccess('/api/servers/a/files'));
  for(const p of ['copy','content','compress','move','rename','delete-bulk'])assert.throws(()=>c.assertClientRequestAccess('/api/servers/a/files/'+p,{method:'POST'}),/Access restricted/);
  c.S.serverAccess.data.permissions=['files.view','files.edit'];assert.doesNotThrow(()=>c.assertClientRequestAccess('/api/servers/a/files/copy',{method:'POST'}));assert.throws(()=>c.assertClientRequestAccess('/api/servers/a/files/move',{method:'POST'}));
  c.S.serverAccess.data.permissions=['files.edit'];assert.throws(()=>c.assertClientRequestAccess('/api/servers/a/files/copy',{method:'POST'}));
});
test('server creation guards ordinary members and honors only admin or explicit panel permission',()=>{
  const {c}=setup();assert.equal(c.canCreateServers(),false);assert.match(c.V.create(),/Access restricted/);assert.throws(()=>c.assertClientRequestAccess('/api/servers',{method:'POST'}));
  c.S.user.panel_role_name='Administrator';assert.equal(c.canCreateServers(),false);
  c.S.user.panel_permissions=['servers.create'];assert.equal(c.canCreateServers(),true);c.S.user={role:'admin'};assert.equal(c.canCreateServers(),true);
  assert.match(source,/function openCreateServerModal\(\) \{\s+if \(!canCreateServers\(\)\)/);
  assert.match(source,/async function handleCreateServer\(e\) \{\s+if \(!canCreateServers\(\)\)/);
});
test('role catalog loads only for built-in admins and assignment sends separate panel-role payload',async()=>{
  const {c,calls}=setup();await c.loadPanelRoles();assert.equal(calls.length,0);
  c.S.user.role='admin';c.api=async(url,opt)=>{calls.push([url,opt]);return {roles:[{id:'r1',name:'Builder',permissions:['servers.create']}],permissions:[{id:'servers.create',label:'Create servers'}]};};
  await c.loadPanelRoles(true);assert.equal(c.S.panelRoles[0].id,'r1');c.S.adminUsers=[{id:2,role:'user'}];
  await c.assignUserPanelRole({disabled:false,value:'r1',dataset:{user:'2',previous:''}});
  assert.equal(calls[1][0],'/api/admin/users/2/panel-role');assert.deepEqual(JSON.parse(calls[1][1].body),{role_id:'r1'});assert.equal(c.S.adminUsers[0].role,'user');
});
test('role assignment errors roll selection back and deletion reports assigned-role conflict',async()=>{
  const {c}=setup();c.S.user.role='admin';c.api=async()=>{throw Error('Role is assigned to users');};let message='';c.toast=m=>message=m;
  const select={disabled:false,value:'r1',dataset:{user:'2',previous:'old'}};await c.assignUserPanelRole(select);assert.equal(select.value,'old');assert.equal(select.disabled,false);
  await c.deletePanelRole('r1',{disabled:false,isConnected:true});assert.match(message,/assigned/);
});
test('community serializes exact nested API fields and renders text safely',()=>{
  const {c}=setup();const value=v=>({value:v}),check=v=>({checked:v});const payload=c.communityPayload({elements:{discord_enabled:check(true),guild_id:value(' 12345 '),discord_title:value(' Community '),announcement_enabled:check(true),announcement_text:value('<img onerror=x>'),tone:value('warning'),dismissible:check(true)}});
  assert.deepEqual(JSON.parse(JSON.stringify(payload)),{discord:{enabled:true,guild_id:'12345',title:'Community'},announcement:{enabled:true,text:'<img onerror=x>',tone:'warning',dismissible:true}});
  c.S.community={announcement:{...payload.announcement,id:'v1'}};assert.match(c.renderAnnouncement(),/&lt;img/);assert.doesNotMatch(c.renderAnnouncement(),/<img/);
  c.dismissAnnouncement();assert.equal(c.renderAnnouncement(),'');c.S.community.announcement.id='v2';assert.match(c.renderAnnouncement(),/Panel announcement/);
});
test('Discord never invents invites from numeric IDs or unavailable widgets',()=>{
  const {c}=setup();for(const url of ['12345','javascript:alert(1)','https://evil.test/x','https://discord.gg.evil.test/x','https://discord.gg@evil.test/x'])assert.equal(c.safeDiscordInvite(url),'');
  assert.equal(c.safeDiscordInvite('https://discord.gg/community'),'https://discord.gg/community');
  c.S.community={discord:{enabled:true,status:'unavailable',guild_id:'12345',invite_url:'https://discord.gg/community',title:'<test>'}};assert.doesNotMatch(c.renderDiscordCard(),/href=/);assert.match(c.renderDiscordCard(),/Invite unavailable/);
});
test('permission checks are single-flight and malformed payloads never unlock tabs',async()=>{
  const {c}=setup();const wait=deferred();let calls=0;c.api=()=>{calls++;return wait.promise;};
  const a=c.ensureServerAccess(),b=c.ensureServerAccess();assert.equal(calls,1);wait.resolve({permissions:['files.view']});await Promise.all([a,b]);assert.equal(c.canServerPage('fi'),false);assert.equal(c.S.serverAccess.status,'error');
});
test('hung permission requests have a bounded retryable error state',async()=>{
  const {c}=setup();let timeout;c.setTimeout=(fn,delay)=>{assert.equal(delay,15000);timeout=fn;return 1;};c.clearTimeout=()=>{};
  c.api=(_,options)=>new Promise((resolve,reject)=>options.signal.addEventListener('abort',()=>reject(Object.assign(Error('aborted'),{name:'AbortError'}))));
  const pending=c.ensureServerAccess();timeout();await pending;assert.equal(c.canServerPage('fi'),false);assert.match(c.S.serverAccess.error,/timed out/);
});
test('SFTP is an explicit shared team permission and connection rendering has one authoritative view',()=>{
  assert.match(slice('const SUBUSER_PERMS', 'async function loadServerSubusers'),/\['sftp.access', 'SFTP Access'/);
  assert.equal((slice('  set() {','  admin() {').match(/renderSftpCardHtml/g)||[]).length,0);
  assert.equal((slice('  net() {','  sub() {').match(/renderSftpCardHtml/g)||[]).length,0);
});
test('community drafts survive rerenders but never follow another login identity',()=>{
  const {c}=setup();c.S.user.role='admin';c.S.communitySettings={discord:{enabled:false,guild_id:'',title:'Saved'},announcement:{enabled:false,text:'',tone:'info',dismissible:true}};
  const f={elements:{discord_enabled:{checked:true},guild_id:{value:'123'},discord_title:{value:'Draft title'},announcement_enabled:{checked:false},announcement_text:{value:''},tone:{value:'info'},dismissible:{checked:true}}};
  c.rememberCommunityDraft(f);assert.match(c.renderCommunitySettings(),/Draft title/);c.S.token='another-login';assert.doesNotMatch(c.renderCommunitySettings(),/Draft title/);
});
test('separate Team permission requests cannot commit after account changes',async()=>{
  const {c}=setup();c.hasServerPermission=()=>true;c.canServerPage=()=>true;c.sameQolServer=()=>true;
  const team=fs.readFileSync(require('node:path').join(__dirname,'../static/js/team.js'),'utf8');vm.runInContext(team,c);
  const wait=deferred();c.api=()=>wait.promise;const pending=c.loadTeamPage(true);c.S.token='new-login';wait.resolve(perms([],{is_owner:true}));await pending;
  assert.equal(c.S.teamState,undefined);assert.equal(c.S.serverSubusers,undefined);
});

test('rename, move and extraction require both edit and delete permission', async () => {
  const { c } = setup(); await c.ensureServerAccess();
  for (const action of ['rename', 'move', 'unzip', 'extract']) {
    const endpoint = '/api/servers/a/files/' + action;
    for (const granted of ['files.edit', 'files.delete']) {
      c.S.serverAccess.data.permissions = ['files.view', granted];
      assert.throws(() => c.assertClientRequestAccess(endpoint, { method: 'POST' }), /Access restricted/, action + ':' + granted);
    }
    c.S.serverAccess.data.permissions = ['files.view', 'files.edit', 'files.delete'];
    assert.doesNotThrow(() => c.assertClientRequestAccess(endpoint, { method: 'POST' }));
  }
});
