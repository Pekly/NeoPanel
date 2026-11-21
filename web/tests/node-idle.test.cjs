const test=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs'),path=require('node:path'),vm=require('node:vm');
const source=fs.readFileSync(path.join(__dirname,'../static/js/node-idle.js'),'utf8');
const view=(enabled=false,timeout_minutes=15,applied=true,error='')=>({config:{enabled,timeout_minutes},applied,error});
function setup(handler=async()=>view()){
 const nodes={};for(const id of ['node-idle-status','node-idle-error','node-idle-retry','node-idle-refresh','node-idle-save','node-idle-fields','nm-free','nm-idle-timeout','nm-timeout-wrap'])nodes[id]={value:'15',checked:false,disabled:false,textContent:'',style:{}};
 const root={isConnected:true,querySelector:id=>nodes[id.slice(1)]||null},calls=[];
 const c={S:{user:{id:1,role:'admin'},token:'one',nodes:[{id:'local'},{id:'remote'}]},document:{querySelector:()=>root},api:async(url,options)=>{calls.push({url,options});return handler(url,options)}};c.window=c;vm.runInNewContext(source,c);
 return{c,calls,nodes,root,ui:c.NPNodeIdle,save:()=>c.NPNodeIdle.saveLocal({preventDefault(){}})};
}
test('Local policy loads authoritative values and saves only idle fields',async()=>{
 const h=setup(async(_,o)=>o.method==='PUT'?{config:JSON.parse(o.body),applied:true,error:''}:view(true,23));await h.ui.attach('local',true);
 assert.equal(h.nodes['nm-free'].checked,true);assert.equal(h.nodes['nm-idle-timeout'].value,23);
 h.nodes['nm-free'].checked=false;h.nodes['nm-idle-timeout'].value='8';await h.save();
 assert.equal(h.calls[1].url,'/api/admin/nodes/local/idle-policy');assert.deepEqual(JSON.parse(h.calls[1].options.body),{enabled:false,timeout_minutes:8});
 assert.match(h.nodes['node-idle-status'].textContent,/disabled.*Applied on this host/);
});
test('Remote refresh and saved-policy retry preserve unsaved form fields',async()=>{
 const h=setup(async(_,o)=>o.method==='PUT'?{config:JSON.parse(o.body),applied:true,error:''}:view(false,15,false,'Worker offline'));
 h.nodes['nm-free'].checked=true;h.nodes['nm-idle-timeout'].value='42';await h.ui.attach('remote');
 assert.equal(h.nodes['nm-idle-timeout'].value,'42');assert.match(h.nodes['node-idle-status'].textContent,/pending disable is not a confirmed disable/);
 await h.nodes['node-idle-retry'].onclick();assert.deepEqual(JSON.parse(h.calls[1].options.body),{enabled:false,timeout_minutes:15});assert.equal(h.nodes['nm-idle-timeout'].value,'42');
});
test('Delivery failures retain confirmed desired policy and show errors as text',async()=>{
 const h=setup(async(_,o)=>{if(o.method==='PUT')throw Object.assign(Error('gateway'),{data:view(true,7,false,'<b>Update worker</b>')});return view();});await h.ui.attach('local',true);
 h.nodes['nm-free'].checked=true;h.nodes['nm-idle-timeout'].value='7';await h.save();assert.equal(h.nodes['node-idle-error'].textContent,'<b>Update worker</b>');assert.match(h.nodes['node-idle-status'].textContent,/Worker application pending/);
});
test('Unknown writes require refresh before retry and never claim accepted state',async()=>{
 const h=setup(async(_,o)=>{if(o.method==='PUT')throw Error('network lost');return view();});await h.ui.attach('local',true);h.nodes['nm-idle-timeout'].value='30';await h.save();
 assert.match(h.nodes['node-idle-status'].textContent,/unknown/);assert.equal(h.nodes['node-idle-retry'].disabled,true);assert.equal(h.nodes['nm-idle-timeout'].value,'30');
 await h.nodes['node-idle-refresh'].onclick();assert.equal(h.nodes['nm-idle-timeout'].value,'30');assert.equal(h.nodes['node-idle-retry'].disabled,false);
});
test('Missing endpoints and invalid minutes do not save fabricated policy',async()=>{
 const h=setup(async()=>{throw Object.assign(Error('missing'),{status:405});});await h.ui.attach('local',true);assert.equal(h.nodes['node-idle-save'].disabled,true);assert.match(h.nodes['node-idle-error'].textContent,/Update and restart/);await h.save();assert.equal(h.calls.length,1);
 const ok=setup();await ok.ui.attach('local',true);ok.nodes['nm-idle-timeout'].value='0';await ok.save();assert.equal(ok.calls.length,1);assert.match(ok.nodes['node-idle-error'].textContent,/1 to 1440/);
});
test('Session switches suppress late responses, and temporary wake-modal detachment preserves load',async()=>{
 let resolve;const h=setup(()=>new Promise(r=>resolve=r));const pending=h.ui.attach('local',true);h.c.S.token='new';resolve(view(true,99));await pending;assert.notEqual(h.nodes['nm-idle-timeout'].value,99);
 h.c.S.token='one';const retry=h.ui.attach('local',true);h.root.isConnected=false;resolve(view(true,12));await retry;assert.equal(h.nodes['nm-idle-timeout'].value,12);assert.equal(h.nodes['node-idle-save'].disabled,false);
 h.c.S.user.role='user';await h.ui.attach('local',true);assert.equal(h.calls.length,2);
});
test('Node save keeps a pending-policy modal open and converts new-node retries into updates',async()=>{
 const app=fs.readFileSync(path.join(__dirname,'../static/js/app.js'),'utf8');const fields={'#nm-free':{checked:true},'#nm-idle-timeout':{value:'5'},'#nm-name':{value:'Node'}};
 const form={dataset:{id:''},isConnected:true},btn={closest:()=>form};fields['#nm-submit-btn']=btn;let closed=false,saved;
 const c={S:{token:'one',user:{id:1,role:'admin'}},$:id=>fields[id],setBtnLoading(){},api:async()=>({id:'new-node',name:'Node',online:true,idle_policy_applied:false,idle_policy_error:'Update worker'}),closeModal:()=>closed=true,toast(){},NPNodeIdle:{savedNode:(...args)=>saved=args},loadDashboard(){throw Error('should preserve form')},loadAdminNodes(){throw Error('should preserve form')}};
 vm.runInNewContext(app.slice(app.indexOf('async function submitNodeModal('),app.indexOf('async function testClusterNode(')),c);await c.submitNodeModal({preventDefault(){}});
 assert.equal(closed,false);assert.equal(form.dataset.id,'new-node');assert.equal(saved[0],'new-node');assert.equal(saved[1].enabled,true);
 assert.match(app,/NPNodeIdle.attach\(existing.id, true\)/);assert.match(app,/onsubmit="NPNodeIdle.saveLocal\(event\)"/);
 const index=fs.readFileSync(path.join(__dirname,'../static/index.html'),'utf8');assert.ok(index.indexOf('/js/node-idle.js')<index.indexOf('/js/app.js'));
});

test('late node saves cannot rebind another node form or a different session',async()=>{
 const app=fs.readFileSync(path.join(__dirname,'../static/js/app.js'),'utf8');
 for(const switchSession of [false,true]){
  const formA={dataset:{id:'a'},isConnected:true},formB={dataset:{id:'local'},isConnected:true},btnA={closest:()=>formA},btnB={closest:()=>formB};let button=btnA,resolve,closed=0,attached=0;
  const fields={'#nm-name':{value:'A'},'#nm-free':{checked:true},'#nm-idle-timeout':{value:'5'}};
  const c={S:{token:'one',user:{id:1,role:'admin'}},$:id=>id==='#nm-submit-btn'?button:fields[id],setBtnLoading(){},api:()=>new Promise(r=>resolve=r),closeModal:()=>closed++,toast(){},NPNodeIdle:{savedNode:()=>attached++},loadDashboard(){},loadAdminNodes(){}};
  vm.runInNewContext(app.slice(app.indexOf('async function submitNodeModal('),app.indexOf('async function testClusterNode(')),c);
  const pending=c.submitNodeModal({preventDefault(){}},'a');
  if(switchSession)c.S.token='other';else{button=btnB;formA.isConnected=false;}
  resolve({id:'a',idle_policy_applied:false,idle_policy_error:'worker offline'});await pending;
  assert.equal(attached,0);assert.equal(closed,0);assert.equal(formB.dataset.id,'local');
 }
});
