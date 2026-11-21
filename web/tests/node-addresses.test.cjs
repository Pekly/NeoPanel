const test=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs'),path=require('node:path'),vm=require('node:vm');
const app=fs.readFileSync(path.join(__dirname,'../static/js/app.js'),'utf8');
function section(start,end){const a=app.indexOf(start),b=app.indexOf(end,a);assert.ok(a>=0&&b>a);return app.slice(a,b);}
function setup(){
 const local={id:'local',is_local:true,name:'Singapore',fqdn:'old-game.example'},remote={id:'remote',fqdn:'remote-game.example'};
 const server={id:'one',node_id:'local',node_fqdn:'old-game.example'},remoteServer={id:'two',node_id:'remote',node_fqdn:'remote-game.example'};
 const fields={'#adm-fqdn':{value:'new-game.example'}};
 const c={S:{token:'admin',user:{role:'admin'},nodes:[local,remote],servers:[server,remoteServer],branding:{},brandingDaemon:{fqdn:'old-game.example'},adminPanelUrl:{public_url:'https://old-panel.example',source:'settings'},sftp:{host:'old-game.example'}},location:{hostname:'browser-old.example'},curServer:()=>server,getNodeById:id=>c.S.nodes.find(n=>n.id===id),esc:s=>String(s??'').replaceAll('&','&amp;').replaceAll('<','&lt;').replaceAll('"','&quot;'),$:id=>fields[id],localStorage:{setItem(){}},toast(){},R(){},loadAdminNodes(){},api:async()=>({daemon:{fqdn:'new-game.example'}})};c.window=c;
 vm.runInNewContext(section('function renderLocalNodeAddresses(', 'function copyDbText('),c);
 vm.runInNewContext(section('async function saveAdminDaemonSettings(', 'async function saveAdminTurnstileSettings('),c);
 return{c,local,remote,server,remoteServer};
}
test('local node displays the saved panel URL independently from its game/SFTP hostname',()=>{
 const {c,local}=setup();let html=c.renderLocalNodeAddresses(local);assert.match(html,/Panel URL:.*https:\/\/old-panel.example/);assert.match(html,/Game \/ SFTP host:.*old-game.example/);
 c.S.adminPanelUrl.public_url='https://new-panel.example';html=c.renderLocalNodeAddresses(local);assert.match(html,/https:\/\/new-panel.example/);assert.doesNotMatch(html,/old-panel.example/);assert.equal(local.fqdn,'old-game.example');assert.equal(c.srvHost(c.curServer()),'old-game.example');
});
test('successful local network save refreshes cached local node, server, and SFTP addresses only',async()=>{
 const {c,local,remote,server,remoteServer}=setup();await c.saveAdminDaemonSettings();assert.equal(local.fqdn,'new-game.example');assert.equal(server.node_fqdn,'new-game.example');assert.equal(c.srvHost(server),'new-game.example');assert.equal(c.S.sftp,null);assert.equal(remote.fqdn,'remote-game.example');assert.equal(remoteServer.node_fqdn,'remote-game.example');assert.equal(c.S.adminPanelUrl.public_url,'https://old-panel.example');
});
test('failed network save leaves cached addresses unchanged',async()=>{
 const {c,local,server}=setup();c.api=async()=>{throw Error('save failed');};await c.saveAdminDaemonSettings();assert.equal(local.fqdn,'old-game.example');assert.equal(server.node_fqdn,'old-game.example');assert.equal(c.S.sftp.host,'old-game.example');
});
test('local cache updates do not clear remote SFTP details',()=>{
 const {c,remoteServer}=setup();c.curServer=()=>remoteServer;c.S.sftp={host:'remote-game.example'};c.applyLocalNodeAddress({fqdn:'local-new.example'});assert.equal(c.S.sftp.host,'remote-game.example');
});
test('node address output is escaped and unavailable panel settings are not guessed from the browser',()=>{
 const {c,local,remote}=setup();c.S.adminPanelUrl=null;c.S.panelUrlLoading=true;assert.match(c.renderLocalNodeAddresses(local),/Loading/);c.S.panelUrlLoading=false;assert.match(c.renderLocalNodeAddresses(local),/Unavailable/);c.S.adminPanelUrl={public_url:''};assert.match(c.renderLocalNodeAddresses(local),/Not configured/);c.S.adminPanelUrl.public_url='<img src=x>';assert.doesNotMatch(c.renderLocalNodeAddresses(local),/<img/);assert.equal(c.renderLocalNodeAddresses(remote),'');
});
test('all local node layouts show labeled addresses and open the network settings section',()=>{
 assert.equal((app.match(/\$\{renderLocalNodeAddresses\(n\)\}/g)||[]).length,2);assert.equal((app.match(/onclick="S.appearanceSection=3;selectAdminTab\(4\)">Edit address/g)||[]).length,2);
 assert.match(section('async function saveBrandingAndDaemon(', 'async function resetBrandingDefaults('),/applyLocalNodeAddress\(S.brandingDaemon\)/);
});
test('Refresh Nodes reloads the saved panel URL and local server-address metadata',async()=>{
 const {c,server}=setup();let panelLoads=0;c.loadAdminPanelUrl=async skip=>{assert.equal(skip,true);panelLoads++;c.S.adminPanelUrl.public_url='https://fresh-panel.example';};c.api=async()=>[{id:'local',is_local:true,fqdn:'fresh-game.example'}];
 vm.runInNewContext(section('async function loadAdminNodes(', 'function buildClusterNodeInstallCmd('),c);await c.loadAdminNodes();assert.equal(panelLoads,1);assert.equal(server.node_fqdn,'fresh-game.example');assert.match(c.renderLocalNodeAddresses(c.S.nodes[0]),/fresh-panel.example/);
});
test('node refresh does not request admin-only panel settings for ordinary users or apply stale session results',async()=>{
 const {c,local}=setup();let panelLoads=0;c.S.user.role='user';c.loadAdminPanelUrl=async()=>panelLoads++;c.api=async()=>{c.S.token='different-session';return[{id:'local',fqdn:'late.example'}];};
 vm.runInNewContext(section('async function loadAdminNodes(', 'function buildClusterNodeInstallCmd('),c);await c.loadAdminNodes();assert.equal(panelLoads,0);assert.equal(c.S.nodes[0],local);assert.equal(local.fqdn,'old-game.example');
});
