const test=require('node:test');
const assert=require('node:assert/strict');
const fs=require('node:fs');
const path=require('node:path');
const vm=require('node:vm');
const app=fs.readFileSync(path.join(__dirname,'../static/js/app.js'),'utf8');
const code=app.slice(app.indexOf('// --- Port Allocation Management ---'),app.indexOf('// --- Server Schedules Management ---'));
function harness(role='user') {
 const container={innerHTML:''},srv={id:'alpha',name:'Alpha',port:25565,ports_used:1,ports_allowed:3};
 const form={dataset:{serverUrl:'/api/servers/alpha/ports'}};
 const fields={'#modal-container':container,'#add-port-form':form,'#ap-port':{value:'30001'},'#ap-proto':{value:'udp'},'#ap-notes':{value:'Voice'},'#ap-primary':{checked:false},'#ap-submit-btn':{disabled:false}};
 const c={S:{user:{role},view:'net',ports:[]},srv,fields,calls:[],toasts:[],renders:0,closed:0,$:s=>fields[s],esc:String,ic:()=>'',getNextSuggestedPort:()=>30001,hasServerPermission:()=>true,enhanceSelects(){},setTimeout(){},console};
 c.window=c;c.curServer=()=>c.srv;c.srvApiPath=(s,p)=>'/api/servers/'+s.id+p;c.toast=(...v)=>c.toasts.push(v);c.closeModal=()=>c.closed++;c.R=()=>c.renders++;
 c.api=async(url,opts)=>{c.calls.push({url,opts});if(opts?.method==='POST')return {port:31000};if(url.endsWith('/summary'))return {used:2,allowed:3};return [{id:1,port:31000}]};
 vm.runInNewContext(code,c);return c;
}
test('member allocation has no numeric choice; administrator retains explicit choice',()=>{
 const member=harness();member.openAddPortModal();const html=member.fields['#modal-container'].innerHTML;
 assert.doesNotMatch(html,/id="ap-port"|Quick Presets/);assert.match(html,/random free port|assigned automatically/);
 const admin=harness('admin');admin.openAddPortModal();assert.match(admin.fields['#modal-container'].innerHTML,/id="ap-port"/);
 member.fields['#modal-container'].innerHTML='';member.srv.ports_used=3;member.openAddPortModal();assert.equal(member.fields['#modal-container'].innerHTML,'');
});
test('member request omits port even when a numeric field is injected; response supplies assigned port',async()=>{
 const c=harness();await c.submitAddPort({preventDefault(){}});
 const body=JSON.parse(c.calls[0].opts.body);assert.equal(Object.hasOwn(body,'port'),false);assert.equal(body.protocol,'udp');
 assert.match(c.toasts[0][0],/31000 reserved/);assert.equal(c.srv.ports_used,2);assert.equal(c.closed,1);
 const admin=harness('admin');await admin.submitAddPort({preventDefault(){}});assert.equal(JSON.parse(admin.calls[0].opts.body).port,30001);
});
test('allocation blocks stale forms and duplicate submits',async()=>{
 const c=harness();c.srv={id:'beta'};await c.submitAddPort({preventDefault(){}});assert.equal(c.calls.length,0);
 c.srv.id='alpha';c.fields['#add-port-form'].dataset.busy='1';await c.submitAddPort({preventDefault(){}});assert.equal(c.calls.length,0);
});
test('late port list cannot overwrite a different active server',async()=>{
 const c=harness();let resolve;c.api=()=>new Promise(r=>{if(!resolve)resolve=r;else r({used:2,allowed:3})});
 const task=c.loadServerPorts();c.srv={id:'beta'};resolve([{port:31000}]);await task;assert.equal(c.S.ports.length,0);assert.equal(c.renders,0);
});
test('Network displays total used/allowed and generic settings primary is read-only',()=>{
 const net=app.slice(app.indexOf('  net() {'),app.indexOf('  sftp() {'));
 assert.match(net,/ports_used/);assert.match(net,/used \/.*allowed/);assert.match(net,/includes the primary port/);
 assert.match(app, /id="prop-port"[^\n]*readonly/);
});

test('Chrome Network shows member quota, random allocation and administrator choices', {skip:!process.env.NP_LAYOUT_PLAYWRIGHT}, async()=>{
 const {chromium}=require(process.env.NP_LAYOUT_PLAYWRIGHT);
 const browser=await chromium.launch({headless:true,executablePath:process.env.NP_LAYOUT_CHROME});
 try {
  const page=await browser.newPage({viewport:{width:1280,height:900}});
  const root=path.join(__dirname,'../static');
  const style=['style.css','modal-layout.css','workspace.css'].map(f=>fs.readFileSync(path.join(root,'css',f),'utf8')).join('\n');
  await page.setContent(`<style>${style}</style><main id="main"></main><div id="modal-container"></div>`);
  await page.evaluate(()=>{
   window.S={user:{role:'user'},view:'net',ports:[],sftp:null};window.srv={id:'alpha',name:'Alpha',port:25565,ports_used:1,ports_allowed:3};
   window.$=s=>document.querySelector(s);window.curServer=()=>srv;window.srvApiPath=(s,p)=>'/api/servers/'+s.id+p;
   window.srvHost=()=> 'play.example.test';window.srvh=()=>'';window.ic=()=>'';window.esc=String;window.routeUrl=()=>'/servers/alpha/sftp';window.hasServerPermission=()=>true;
   window.getNextSuggestedPort=()=>30001;window.enhanceSelects=()=>{};window.toast=()=>{};window.closeModal=()=>$('#modal-container').replaceChildren();
   window.calls=[];window.api=async(url,opts)=>{calls.push({url,opts});if(opts?.method==='POST')return {port:31000};if(url.endsWith('/summary'))return {used:2,allowed:3};return [{id:1,port:25565,is_primary:true},{id:2,port:31000}]};
  });
  const render=app.slice(app.indexOf('  net() {'),app.indexOf('  db() {'));
  await page.addScriptTag({content:`window.networkRenderer={${render}};window.R=()=>{document.querySelector('#main').innerHTML=networkRenderer.net()};${code}`});
  await page.evaluate(()=>R());
  assert.match(await page.locator('#main').innerText(),/1 used \/ 3 allowed/);
  assert.equal(await page.locator('#port-allowance').count(),0);
  await page.getByRole('button',{name:'Assign Port',exact:true}).click();
  assert.equal(await page.locator('#ap-port').count(),0);
  await page.locator('#ap-submit-btn').click();
  await page.waitForFunction(()=>calls.some(c=>c.opts?.method==='POST'));
  const memberBody=await page.evaluate(()=>JSON.parse(calls.find(c=>c.opts?.method==='POST').opts.body));
  assert.equal(Object.hasOwn(memberBody,'port'),false);
  await page.waitForFunction(()=>srv.ports_used===2);
  await page.evaluate(()=>{srv.ports_used=3;R()});
  assert.equal(await page.getByRole('button',{name:'Assign Port',exact:true}).isDisabled(),true);
  await page.evaluate(()=>{S.user.role='admin';srv.ports_used=2;R()});
  assert.equal(await page.locator('#port-allowance').inputValue(),'3');
  await page.getByRole('button',{name:'Assign Port',exact:true}).click();
  await page.locator('#ap-port').fill('32001');
  await page.locator('#ap-submit-btn').click();
  await page.waitForFunction(()=>calls.filter(c=>c.opts?.method==='POST').length===2);
  assert.equal(await page.evaluate(()=>JSON.parse(calls.filter(c=>c.opts?.method==='POST')[1].opts.body).port),32001);
 } finally {await browser.close()}
});
