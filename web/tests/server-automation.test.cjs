const test=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs'),path=require('node:path'),vm=require('node:vm');
const source=fs.readFileSync(path.join(__dirname,'../static/js/server-automation.js'),'utf8');
const data=overrides=>({join_to_start_enabled:true,node_join_to_start_enabled:true,idle_shutdown_enabled:true,idle_timeout_minutes:15,idle_shutdown_locked:true,stopped_by_idle_policy:false,wake_eligible:false,can_edit_join_to_start:true,reason:'Waiting for a successful idle auto-shutdown',...overrides});
function setup(handler=async()=>data()){
 let srv={id:'one',node_id:'worker'},root;const calls=[];
 const nodes={};for(const id of ['idle','idle-note','wake','wake-note','refresh','status','error'])nodes[id]={checked:false,disabled:false,textContent:'',hidden:false};
 root={isConnected:true,querySelector:id=>nodes[id.replace('#server-automation-','')]};
 const c={S:{token:'one',user:{id:2,role:'user'},view:'set'},curServer:()=>srv,srvApiPath:(s,suffix)=>s.node_id==='local'?`/api/servers/${s.id}${suffix}`:`/api/nodes/${s.node_id}/proxy/api/servers/${s.id}${suffix}`,document:{querySelector:()=>root},api:async(url,options)=>{calls.push({url,options});return handler(url,options)}};c.window=c;vm.runInNewContext(source,c);
 return{c,nodes,calls,root,ui:c.NPServerAutomation,setServer:s=>srv=s};
}
test('ordinary owners get an editable wake preference but a checked locked idle toggle',async()=>{
 const h=setup(async(_,o)=>data(o.method==='PATCH'?{join_to_start_enabled:JSON.parse(o.body).join_to_start_enabled}:{}));await h.ui.mount();
 assert.equal(h.nodes.idle.checked,true);assert.equal(h.nodes.idle.disabled,true);assert.equal(h.nodes.idle.onchange,undefined);assert.equal(h.nodes.wake.disabled,false);
 await h.nodes.wake.onchange({target:{checked:false}});
 assert.equal(h.calls[1].url,'/api/nodes/worker/proxy/api/servers/one/automation');assert.equal(h.calls[1].options.method,'PATCH');assert.deepEqual(JSON.parse(h.calls[1].options.body),{join_to_start_enabled:false});
 assert.equal(h.nodes.idle.checked,true);assert.equal(h.nodes.idle.disabled,true);assert.equal(h.nodes.wake.checked,false);
});
test('view-only permissions cannot write, including a forced change event',async()=>{
 const h=setup(async()=>data({can_edit_join_to_start:false}));await h.ui.mount();assert.equal(h.nodes.wake.disabled,true);await h.nodes.wake.onchange({target:{checked:false}});assert.equal(h.calls.length,1);
});
test('unknown policy never fabricates an enabled control; readonly policy fields are required',async()=>{
 for(const response of [{},data({idle_shutdown_locked:false})]){const h=setup(async()=>response);await h.ui.mount();assert.equal(h.nodes.wake.disabled,true);assert.equal(h.nodes.idle.disabled,true);assert.match(h.nodes.error.textContent,/Unexpected/);}
});
test('failed writes require refresh and keep auto-shutdown locked',async()=>{
 const h=setup(async(_,o)=>{if(o.method==='PATCH')throw Error('Worker unavailable');return data()});await h.ui.mount();await h.nodes.wake.onchange({target:{checked:false}});assert.equal(h.nodes.wake.disabled,true);assert.equal(h.nodes.idle.checked,true);assert.equal(h.nodes.idle.disabled,true);assert.match(h.nodes.status.textContent,/unconfirmed/);await h.nodes.refresh.onclick();assert.equal(h.nodes.wake.disabled,false);
});
test('responses are pinned to server, node and session',async()=>{
 for(const change of ['server','node','session','route']){let resolve;const h=setup(()=>new Promise(r=>resolve=r));const pending=h.ui.mount();if(change==='server')h.setServer({id:'two',node_id:'worker'});if(change==='node')h.setServer({id:'one',node_id:'other'});if(change==='session')h.c.S.token='other';if(change==='route')h.c.S.view='dash';resolve(data());await pending;assert.equal(h.nodes.wake.disabled,true);}
});
test('rendering explains idle-only eligibility and preserves startup-settings integration',()=>{
 const h=setup(),html=h.ui.render({id:'<img>',node_id:'worker'});assert.doesNotMatch(html,/<img>/);assert.match(html,/Locked by node policy/);assert.match(html,/Manual stops, crashes/);assert.match(html,/id="server-automation-idle"[^>]*disabled/);
 const app=fs.readFileSync(path.join(__dirname,'../static/js/app.js'),'utf8');assert.match(app,/NPServerAutomation.render\(srv\)/);assert.match(app,/NPServerAutomation\?\.mount\(\)/);assert.match(app,/el.id !== 'server-automation-settings'/);
 const index=fs.readFileSync(path.join(__dirname,'../static/index.html'),'utf8');assert.ok(index.indexOf('/js/server-automation.js')<index.indexOf('/js/app.js'));
});

test('client API guard mirrors settings-and-power writes without blocking policy reads',()=>{
 const app=fs.readFileSync(path.join(__dirname,'../static/js/app.js'),'utf8'),permissions=new Set(['settings.edit']);
 const c={S:{user:{role:'user'}},hasServerPermission:p=>permissions.has(p)};
 vm.runInNewContext(app.slice(app.indexOf('function clientRequestPermission('),app.indexOf('function applyPermissionControls(')),c);
 assert.doesNotThrow(()=>c.assertClientRequestAccess('/api/nodes/worker/proxy/api/servers/one/automation'));
 assert.throws(()=>c.assertClientRequestAccess('/api/nodes/worker/proxy/api/servers/one/automation',{method:'PATCH'}),/server.power/);
 permissions.add('server.power');assert.doesNotThrow(()=>c.assertClientRequestAccess('/api/nodes/worker/proxy/api/servers/one/automation',{method:'PATCH'}));
 permissions.delete('settings.edit');assert.throws(()=>c.assertClientRequestAccess('/api/servers/one/automation',{method:'PATCH'}),/settings.edit/);
});
