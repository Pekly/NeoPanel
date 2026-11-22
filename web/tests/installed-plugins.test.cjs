const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const root = path.join(__dirname,'../static');
const source = fs.readFileSync(path.join(root,'js/installed-plugins.js'),'utf8');
const app = fs.readFileSync(path.join(root,'js/app.js'),'utf8');
const part = (a,b) => app.slice(app.indexOf(a),app.indexOf(b,app.indexOf(a)));
const hash = 'a'.repeat(128);
const item = (n,match='hash') => ({filename:`demo-${n}.jar${n===1?'.disabled':''}`,size:2048,enabled:n!==1,sha512:hash,name:`Demo utility ${n}`,installed_version:'1.0',match,project:['hash','confirmed'].includes(match)?{id:'p'+n,slug:'demo-'+n,title:'Demo project '+n,icon_url:'javascript:alert(1)'}:null,latest:{id:'v'+n,version_number:'2.0',game_versions:['1.21.5'],loaders:['paper']},update_available:true,candidates:match==='suggested'?[{id:'candidate',slug:'candidate-slug',title:'Possible demo source',description:'Demo API data — a name-based suggestion.'},{id:'other',title:'Another source'}]:[],message:'/home/private/provider-error'});
function harness(count=5) {
  const calls=[],events={},timers=new Map();let seq=0;
  const data=Array.from({length:count},(_,n)=>item(n,['hash','hash','suggested','unknown','error'][n]||'hash'));
  const c={S:{view:'plug',plugTab:'installed',status:'offline',user:{id:1,role:'admin'},token:'token',plugins:data.map(r=>[r.filename,'2 KB',r.enabled,'',false])},server:{id:'a',node_id:'local',type:'paper'},permission:true,feature:true,URL,AbortController,console,Set,Map,
    document:{addEventListener:(k,f)=>events[k]=f,getElementById:()=>null},curServer:()=>c.server,isSteelEngine:()=>false,hasServerPermission:()=>c.permission,NPFeatures:{enabled:()=>c.feature},getViewMode:()=> 'list',formatFileDate:()=>'',R(){c.NPInstalledPlugins.sync();},loadPluginsList:async()=>{},uiConfirm:async opts=>{c.confirm=opts;return true;},srvApiPath:(s,p)=>`/api/nodes/${s.node_id}/servers/${s.id}${p}`,setTimeout:(f,ms)=>{timers.set(++seq,{f,ms});return seq;},clearTimeout:id=>timers.delete(id)};
  c.window=c;c.api=async(url,o={})=>{calls.push({url,o});if(c.handler)return c.handler(url,o);if(url.includes('/catalog?')){const offset=Number(new URL(url,'http://demo').searchParams.get('offset'));return{items:data.slice(offset,offset+20),offset,total:data.length,has_more:offset+20<data.length};}if(url.endsWith('/identify'))return{message:'Source confirmed'};const payload=JSON.parse(o.body);return{results:payload.items.map(r=>({filename:r.filename,status:'updated',backup_path:'plugins/.backups/demo.jar'})),restart_required:true};};
  vm.runInNewContext(source,c);
  return {c,calls,data,events,timers,choose(id='candidate'){events.change({target:{matches:q=>q==='[name="npi-source"]',value:id,id:'npi-candidate-0'}});}};
}
test('opt-in metadata distinguishes exact hashes, suggestions, unknown, provider errors and disabled files',async()=>{
  const {c,calls}=harness();assert.match(c.NPInstalledPlugins.render(),/Not checked/);c.R();c.R();assert.equal(calls.length,0);
  await c.NPInstalledPlugins.scan();const html=c.NPInstalledPlugins.render();
  for(const text of ['Exact SHA512 match','Possible match','Unknown','Lookup failed','Disabled','demo-1.jar.disabled','1.0','2.0'])assert.ok(html.includes(text),text);
  assert.equal((html.match(/data-npi="project"/g)||[]).length,2);assert.doesNotMatch(html,/\/home\/private|src="javascript:/);
  c.NPInstalledPlugins.select('demo-2.jar',true);assert.match(c.NPInstalledPlugins.render(),/0 \/ 10 selected/);
});
test('scan singleflight, 20-file pages, bounded retention, errors and abort on context change',async()=>{
  const {c,calls,data}=harness(1020);await c.NPInstalledPlugins.scan();await c.NPInstalledPlugins.scan(true);assert.match(calls[1].url,/offset=20&limit=20/);
  for(let n=0;n<60;n++)await c.NPInstalledPlugins.scan(true);assert.equal(calls.length,50);assert.match(c.NPInstalledPlugins.render(),/1000 of 1020/);
  let resolve;c.handler=()=>new Promise(r=>resolve=r);const pending=c.NPInstalledPlugins.scan();await c.NPInstalledPlugins.scan();assert.equal(calls.length,51);c.server.node_id='b';c.R();assert.equal(calls.at(-1).o.signal.aborted,true);resolve({items:[data[0]],offset:0,total:1,has_more:false});await pending;assert.doesNotMatch(c.NPInstalledPlugins.render(),/Exact SHA512/);
});
test('source confirmation is explicit, deduplicated, SHA-bound and refreshed only in exact context',async()=>{
  const {c,calls,choose,data}=harness();await c.NPInstalledPlugins.scan();c.NPInstalledPlugins.review('demo-2.jar');await c.NPInstalledPlugins.confirmSource();assert.equal(calls.length,1);choose();
  let resolve;c.uiConfirm=opts=>{c.confirm=opts;return new Promise(r=>resolve=r);};const p=c.NPInstalledPlugins.confirmSource();await c.NPInstalledPlugins.confirmSource();assert.match(c.confirm.message,/custom or modified JAR/);data[2].match='confirmed';data[2].project={id:'candidate',title:'Confirmed source'};resolve(true);await p;
  assert.deepEqual(JSON.parse(calls.find(x=>x.url.endsWith('/identify')).o.body),{filename:'demo-2.jar',sha512:hash,project_id:'candidate',confirm:true});assert.equal(calls.filter(x=>x.url.includes('/catalog?')).length,2);assert.match(c.NPInstalledPlugins.render(),/Source confirmed · installed release unverified/);
});
test('update payload contains only approved selection, maximum ten, explicit preview and partial results',async()=>{
  const {c,calls}=harness(20);await c.NPInstalledPlugins.scan();for(let n=0;n<20;n++)c.NPInstalledPlugins.select(`demo-${n}.jar${n===1?'.disabled':''}`,true);
  assert.match(c.NPInstalledPlugins.render(),/10 \/ 10 selected/);
  c.handler=async(url,o)=>({results:JSON.parse(o.body).items.map((r,i)=>({filename:r.filename,status:i===0?'updated':i===1?'failed':'skipped',message:'/private/server/path',backup_path:i===0?'plugins/.backup/<b>.jar':'C:/secret'})),restart_required:true});
  await c.NPInstalledPlugins.update();const payload=JSON.parse(calls.find(x=>x.url.endsWith('/update')).o.body);assert.equal(payload.items.length,10);assert.equal(payload.confirm,true);assert.ok(payload.items.every(x=>x.sha512===hash&&x.project_id&&x.version_id));assert.ok(!payload.items.some(x=>x.filename==='demo-2.jar'));
  for(const text of ['demo-1.jar.disabled','1.0','2.0','1.21.5','Java','Configuration files','disabled files stay disabled','No automatic restart'])assert.ok(c.confirm.message.includes(text),text);
  const html=c.NPInstalledPlugins.render();assert.match(html,/updated/);assert.match(html,/failed/);assert.match(html,/skipped/);assert.match(html,/plugins\/\.backup\/&lt;b&gt;\.jar/);assert.doesNotMatch(html,/C:\/secret|\/private\/server/);
});
test('no delayed mutations after server/node/account/token/software/permission/feature/list/navigation changes',async()=>{
  for(const change of ['server','node','account','token','software','permission','feature','list','tab','view','reset','running'])for(const action of ['update','identify']) {
    if(change==='running'&&action==='identify')continue;
    const {c,calls,choose}=harness();await c.NPInstalledPlugins.scan();let resolve;c.uiConfirm=()=>new Promise(r=>resolve=r);
    if(action==='identify'){c.NPInstalledPlugins.review('demo-2.jar');choose();}else c.NPInstalledPlugins.select('demo-0.jar',true);
    const p=action==='identify'?c.NPInstalledPlugins.confirmSource():c.NPInstalledPlugins.update();
    if(change==='server')c.server.id='b';if(change==='node')c.server.node_id='b';if(change==='account')c.S.user.id=2;if(change==='token')c.S.token='b';if(change==='software')c.server.type='fabric';if(change==='permission')c.permission=false;if(change==='feature')c.feature=false;if(change==='list')c.S.plugins[0][0]='renamed.jar';if(change==='tab')c.S.plugTab='browse';if(change==='view')c.S.view='fi';if(change==='reset')c.NPInstalledPlugins.reset();if(change==='running')c.S.status='online';
    c.R();resolve(true);await p;assert.equal(calls.filter(x=>x.o.method==='POST').length,0,action+'/'+change);
  }
});
test('uncertain writes survive reset/navigation, require offset-zero successful rescan, and never auto-retry',async()=>{
  for(const failure of ['network','timeout','malformed','missing-file','stale']) {
    const {c,calls,timers}=harness();await c.NPInstalledPlugins.scan();c.NPInstalledPlugins.select('demo-0.jar',true);
    c.handler=async(url,o)=>{if(failure==='network')throw Error('lost');if(failure==='malformed')return{};if(failure==='missing-file')return{results:[],restart_required:true};if(failure==='stale'){c.S.view='fi';c.R();throw Object.assign(Error('stale'),{name:'AbortError'});}return new Promise((res,rej)=>o.signal.addEventListener('abort',()=>rej(Object.assign(Error('timeout'),{name:'AbortError'}))));};
    const p=c.NPInstalledPlugins.update();for(let n=0;n<5;n++)await Promise.resolve();if(failure==='timeout'){const timer=[...timers.values()].find(t=>t.ms===150000);assert.ok(timer);timer.f();}await p;
    c.S.view='plug';c.NPInstalledPlugins.reset();assert.match(c.NPInstalledPlugins.render(),/Previous write outcome is uncertain/);await c.NPInstalledPlugins.update();assert.equal(calls.filter(x=>x.o.method==='POST').length,1);
    c.handler=async()=>{throw Error('backend lease busy');};await c.NPInstalledPlugins.scan();assert.match(c.NPInstalledPlugins.render(),/Previous write outcome is uncertain/);
    c.handler=null;await c.NPInstalledPlugins.scan();assert.match(calls.at(-1).url,/offset=0/);assert.doesNotMatch(c.NPInstalledPlugins.render(),/Previous write outcome is uncertain/);assert.match(c.NPInstalledPlugins.render(),/0 \/ 10 selected/);assert.equal(calls.filter(x=>x.o.method==='POST').length,1);
  }
});
test('running/read-only still allows scans; preflight 409 clears stale mapping without uncertainty',async()=>{
  const {c,calls}=harness();c.S.status='online';await c.NPInstalledPlugins.scan();c.NPInstalledPlugins.select('demo-0.jar',true);await c.NPInstalledPlugins.update();assert.equal(calls.length,1);
  c.S.status='offline';c.NPInstalledPlugins.select('demo-0.jar',true);c.handler=async()=>{throw Object.assign(Error('/secret'),{status:409});};await c.NPInstalledPlugins.update();assert.match(c.NPInstalledPlugins.render(),/Nothing was updated/);assert.doesNotMatch(c.NPInstalledPlugins.render(),/Previous write outcome is uncertain/);
  c.permission=false;assert.match(c.NPInstalledPlugins.render(),/Read-only/);
});

test('Chrome: installed renderer, real confirm dialog, metadata and mutation flows, mobile and keyboard (mock APIs)',{skip:!process.env.NP_LAYOUT_PLAYWRIGHT},async()=>{
  const {chromium}=require(process.env.NP_LAYOUT_PLAYWRIGHT);const browser=await chromium.launch({headless:true,executablePath:process.env.NP_LAYOUT_CHROME});
  const index=fs.readFileSync(path.join(root,'index.html'),'utf8');const styles=[...index.matchAll(/href="\/css\/([^"]+)"/g)].map(m=>fs.readFileSync(path.join(root,'css',m[1]),'utf8')).join('\n');
  try {
    const page=await browser.newPage({viewport:{width:1440,height:1000}}),errors=[];page.on('pageerror',e=>errors.push(e.message));
    await page.setContent(`<html data-theme="dark"><head><style>${styles}</style></head><body><main id="main"></main><div id="modal-container"></div></body></html>`);
    await page.evaluate(data=>{
      window.demoData=data;window.S={view:'plug',plugTab:'installed',status:'offline',user:{id:1,role:'admin'},token:'demo',plugins:data.map(r=>[r.filename,'2 KB',r.enabled,'',false]),lastModrinthHits:[],loading:{}};window.server={id:'a',node_id:'local',type:'paper',name:'Demo server'};window.permission=true;window.mode='list';window.calls=[];window.failWrite=false;window.failScan=false;
      window.curServer=()=>server;window.hasServerPermission=()=>permission;window.isSteelEngine=()=>false;window.isModdedEngine=()=>false;window.NPFeatures={enabled:()=>true};window.getViewMode=()=>mode;window.formatFileDate=()=>'';window.formatFileSize=()=> '2 KB';window.getAllMcVersions=()=>[];window.renderViewModeToggle=()=>'';window.renderModrinthHitsHtml=()=>'';window.srvh=()=>'<p>Demo API data · Installed plugin test</p>';window.ic=()=>'';window.loaderHtml=s=>s;window.$=q=>document.querySelector(q);window.esc=v=>String(v??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));window.toast=()=>{};window.sameQolServer=(a,b)=>a.id===b.id&&a.node_id===b.node_id;window.setPlugTab=t=>{S.plugTab=t;R();};
      window.NPPlugins={render:()=>'',capture:()=>({live:()=>true}),safeURL:v=>/^https?:\/\//.test(v||'')?v:'',open:s=>window.openedProject=s,loaderOptions:()=>''};
      window.R=()=>{NPInstalledPlugins.sync();document.getElementById('main').innerHTML=views.plug();};window.srvApiPath=(s,p)=>`/api/nodes/${s.node_id}/servers/${s.id}${p}`;
      window.api=async(url,o={})=>{calls.push({url,method:o.method,body:o.body});if(url.includes('/catalog?')){if(failScan)throw Error('lease busy');const offset=Number(new URL(url,'http://demo').searchParams.get('offset'));return{items:structuredClone(demoData.slice(offset,offset+20)),offset,total:demoData.length,has_more:offset+20<demoData.length};}if(url.endsWith('/identify')){const r=demoData.find(r=>r.filename===JSON.parse(o.body).filename);r.match='confirmed';r.project={id:'candidate',title:'Confirmed demo source'};return{message:'confirmed'};}if(url.endsWith('/update')){if(failWrite)throw Error('lost response');return{results:JSON.parse(o.body).items.map((r,i)=>({filename:r.filename,status:i===0?'updated':'failed',backup_path:i===0?'plugins/.backup/<demo>.jar':'',message:'/private/host/path'})),restart_required:true};}if(url.endsWith('/toggle')){const p=JSON.parse(o.body),r=demoData.find(r=>r.filename===p.filename);r.filename=p.enabled?r.filename.replace(/\.disabled$/,''):r.filename+'.disabled';r.enabled=p.enabled;return{filename:r.filename,enabled:r.enabled};}if(o.method==='DELETE'){const f=new URL(url,'http://demo').searchParams.get('filename');demoData=demoData.filter(r=>r.filename!==f);return{};}return demoData.map(r=>({filename:r.filename,size:r.size,enabled:r.enabled}));};
      window._uiModalResolver=null;window.closeUiModal=ok=>{document.getElementById('modal-container').innerHTML='';const r=_uiModalResolver;_uiModalResolver=null;r?.(ok);};
    },Array.from({length:24},(_,n)=>item(n,['hash','hash','suggested','unknown','error'][n]||'hash')));
    await page.addScriptTag({content:source});await page.addScriptTag({content:'window.views={'+part('  plug() {','  ver() {')+'};\n'+part('function uiConfirm(', 'function uiPrompt(')+'\n'+part('async function loadPluginsList(', 'function searchModrinthDebounced(')});await page.evaluate(()=>R());
    assert.equal(await page.evaluate(()=>calls.length),0);await page.getByRole('button',{name:'Check for updates',exact:true}).click();await page.getByText('Exact SHA512 match',{exact:true}).first().waitFor();
    assert.equal(await page.locator('.npi-project').count(),17);assert.ok(await page.locator('#npi-select-2').isDisabled());assert.ok(await page.locator('#npi-select-3').isDisabled());assert.match(await page.locator('.npi-file').nth(1).textContent(),/Disabled/);
    await page.locator('.npi-project').first().click();assert.equal(await page.evaluate(()=>openedProject),'demo-0');
    // Reviewing a candidate never schedules a POST for a different server.
    await page.getByRole('button',{name:'Review possible match'}).click();await page.locator('#npi-candidate-0').check();await page.getByRole('button',{name:'Confirm source',exact:true}).click();await page.evaluate(()=>{server.id='other';R();});await page.locator('#ui-modal-confirm-btn').click();assert.equal(await page.evaluate(()=>calls.filter(x=>x.url.endsWith('/identify')).length),0);await page.evaluate(()=>{server.id='a';R();});await page.locator('#npi-check').click();
    await page.getByRole('button',{name:'Review possible match'}).click();assert.equal(await page.locator('input[name="npi-source"]:checked').count(),0);await page.locator('#npi-candidate-0').check();await page.getByRole('button',{name:'Confirm source',exact:true}).click();assert.match(await page.locator('.ui-modal-body').textContent(),/custom or modified JAR/);await page.locator('#ui-modal-confirm-btn').click();await page.getByText('Source confirmed · installed release unverified',{exact:true}).waitFor();
    assert.deepEqual(await page.evaluate(()=>JSON.parse(calls.find(x=>x.url.endsWith('/identify')).body)),{filename:'demo-2.jar',sha512:hash,project_id:'candidate',confirm:true});
    await page.getByRole('button',{name:'Load more scan'}).click();assert.equal(await page.locator('.npi-project').count(),22);
    for(const i of [0,1,2,5,6,7,8,9,10,11]) await page.locator(`#npi-select-${i}`).check();
    assert.equal(await page.locator('[data-npi-select]:checked').count(),10);assert.ok(await page.locator('#npi-select-12').isDisabled());await page.getByRole('button',{name:'Clear selection',exact:true}).click();
    // Exact filename and version preview, and only selected payloads.
    await page.locator('#npi-select-0').check();await page.locator('#npi-select-1').check();await page.locator('#npi-update').click();assert.match(await page.locator('.npi-preview').textContent(),/demo-1.jar.disabled/);assert.match(await page.locator('.npi-preview').textContent(),/1.21.5/);await page.locator('#ui-modal-confirm-btn').click();await page.getByRole('region',{name:'Update results'}).waitFor();assert.match(await page.locator('.npi-results').textContent(),/failed/);assert.match(await page.locator('.npi-results').textContent(),/plugins\/\.backup\/<demo>\.jar/);assert.equal(await page.locator('.npi-results demo').count(),0);
    assert.equal(await page.evaluate(()=>JSON.parse(calls.find(x=>x.url.endsWith('/update')).body).items.length),2);
    // Lost write barrier survives closing and tab reset, cannot retry until successful scan.
    await page.evaluate(()=>{failWrite=true;});await page.locator('#npi-check').click();await page.locator('#npi-select-0').check();await page.locator('#npi-update').click();await page.locator('#ui-modal-confirm-btn').click();await page.getByText(/Previous write outcome is uncertain/).waitFor();await page.evaluate(()=>{S.view='fi';NPInstalledPlugins.sync();S.view='plug';NPInstalledPlugins.reset();R();});assert.ok(await page.locator('#npi-select-0').isDisabled());await page.evaluate(()=>failScan=true);await page.locator('#npi-check').click();assert.match(await page.locator('.npi-feedback').textContent(),/uncertain/);await page.evaluate(()=>{failScan=false;failWrite=false;});await page.locator('#npi-check').click();assert.ok(await page.locator('#npi-select-0').isEnabled());assert.equal(await page.evaluate(()=>calls.filter(x=>x.url.endsWith('/update')).length),2);
    // Stale confirmation must not POST; running state disables mutation but not reads.
    await page.locator('#npi-select-0').check();await page.locator('#npi-update').click();await page.evaluate(()=>{S.user.id++;R();});await page.locator('#ui-modal-confirm-btn').click();assert.equal(await page.evaluate(()=>calls.filter(x=>x.url.endsWith('/update')).length),2);
    await page.evaluate(()=>{S.status='online';R();});await page.locator('#npi-check').click();assert.ok(await page.locator('#npi-select-0').isDisabled());assert.ok(await page.locator('#npi-check').isEnabled());await page.evaluate(()=>{S.status='offline';R();});
    // Existing toggle/delete handlers keep filenames and invalidate metadata.
    await page.getByRole('checkbox',{name:'Enable demo-0.jar',exact:true}).focus();await page.keyboard.press('Space');await page.waitForFunction(()=>S.plugins[0][0]==='demo-0.jar.disabled');assert.match(await page.locator('.npi-file').first().textContent(),/Not checked/);await page.locator('.npi-file').first().getByRole('button',{name:'Uninstall'}).click();await page.locator('#ui-modal-confirm-btn').click();await page.waitForFunction(()=>S.plugins.length===23);
    // Provider strings remain escaped, unsafe URLs rejected.
    await page.evaluate(()=>{demoData[0].project.title='<img onerror=window.pwned=1>';demoData[0].project.icon_url='javascript:alert(1)';demoData[0].installed_version='<script>bad</script>';demoData[1].match='suggested';demoData[1].candidates=[{id:'x',title:'<svg onload=window.pwned=1>',description:'<img onerror=window.pwned=1>'}];});await page.locator('#npi-check').click();await page.getByRole('button',{name:'Review possible match'}).click();assert.equal(await page.locator('.npi [onerror],.npi [onload],.npi script,.npi [src^="javascript:"]').count(),0);assert.equal(await page.evaluate(()=>window.pwned),undefined);
    // Check stable responsive tracks in every presentation and both established themes.
    for(const width of [1440,390,320])for(const theme of ['dark','light'])for(const mode of ['list','tiles','icons']) {
      await page.setViewportSize({width,height:900});await page.emulateMedia({reducedMotion:'reduce'});await page.evaluate(({mode,theme})=>{window.mode=mode;document.documentElement.dataset.theme=theme;R();},{mode,theme});
      assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),`${width}/${theme}/${mode} page overflow`);
      assert.ok(await page.locator('.npi-file').evaluateAll(els=>els.every(el=>el.scrollWidth<=el.clientWidth+1)),`${width}/${theme}/${mode} card overflow`);
      assert.equal(await page.locator('.npi-file').first().evaluate(e=>getComputedStyle(e).animationName),'none');
      assert.notEqual(await page.locator('.npi-file').first().evaluate(e=>getComputedStyle(e).backgroundColor),'rgba(0, 0, 0, 0)');
      await page.locator('#npi-check').focus();await page.keyboard.press('Tab');const focus=await page.evaluate(()=>({tag:document.activeElement.tagName,outline:getComputedStyle(document.activeElement).outlineStyle}));assert.notEqual(focus.tag,'BODY');assert.equal(focus.outline,'solid');
    }
    await page.evaluate(()=>{permission=false;R();});assert.match(await page.locator('.npi-bulk').textContent(),/Read-only/);assert.ok(await page.locator('.npi-file').first().getByRole('button',{name:'Uninstall'}).isDisabled());
    if(process.env.NP_PLUGIN_ARTIFACTS){fs.mkdirSync(process.env.NP_PLUGIN_ARTIFACTS,{recursive:true});await page.screenshot({path:path.join(process.env.NP_PLUGIN_ARTIFACTS,'installed-mobile.png'),fullPage:true});fs.writeFileSync(path.join(process.env.NP_PLUGIN_ARTIFACTS,'installed-preview.html'),(await page.content()).replace(/<script[\s\S]*?<\/script>/g,''));}
    assert.deepEqual(errors,[]);
  } finally {await browser.close();}
});

test('write barrier follows exact server/account/token across role/software changes; other identities are isolated',async()=>{
  const {c}=harness();await c.NPInstalledPlugins.scan();c.NPInstalledPlugins.select('demo-0.jar',true);c.handler=async()=>{throw Error('connection lost');};await c.NPInstalledPlugins.update();
  c.S.user.role='member';c.server.type='fabric';c.NPInstalledPlugins.reset();assert.match(c.NPInstalledPlugins.render(),/Previous write outcome is uncertain/);
  c.server.id='other';assert.doesNotMatch(c.NPInstalledPlugins.render(),/Previous write outcome is uncertain/);c.server.id='a';assert.match(c.NPInstalledPlugins.render(),/Previous write outcome is uncertain/);
  c.S.token='other';assert.doesNotMatch(c.NPInstalledPlugins.render(),/Previous write outcome is uncertain/);c.S.token='token';assert.match(c.NPInstalledPlugins.render(),/Previous write outcome is uncertain/);
});
test('source 409 invalidates stale review; completed writes keep truthful results if list refresh fails',async()=>{
  const {c,calls,choose}=harness();await c.NPInstalledPlugins.scan();c.NPInstalledPlugins.review('demo-2.jar');choose();c.handler=async()=>{throw Object.assign(Error('changed'),{status:409});};await c.NPInstalledPlugins.confirmSource();assert.match(c.NPInstalledPlugins.render(),/The file changed/);assert.doesNotMatch(c.NPInstalledPlugins.render(),/Review possible sources/);await c.NPInstalledPlugins.confirmSource();assert.equal(calls.filter(x=>x.o.method==='POST').length,1);
  c.handler=null;await c.NPInstalledPlugins.scan();c.NPInstalledPlugins.select('demo-0.jar',true);c.loadPluginsList=async()=>{throw Error('read failed');};await c.NPInstalledPlugins.update();assert.match(c.NPInstalledPlugins.render(),/Update results/);assert.doesNotMatch(c.NPInstalledPlugins.render(),/Previous write outcome is uncertain/);
});

test('actionable local limits and rollback repair warnings are visible without leaking arbitrary errors',async()=>{
  const {c,data}=harness();data[4].message='plugin exceeds file or page byte limit';await c.NPInstalledPlugins.scan();
  assert.match(c.NPInstalledPlugins.render(),/plugin exceeds file or page byte limit/);assert.doesNotMatch(c.NPInstalledPlugins.render(),/Provider unavailable|\/home\/private/);
  c.NPInstalledPlugins.select('demo-0.jar',true);c.handler=async(url,o)=>({results:JSON.parse(o.body).items.map(r=>({filename:r.filename,status:'failed',message:'publication and rollback failed; manual repair required before starting',backup_path:'plugin-backups/batch/demo.jar'})),restart_required:false});
  await c.NPInstalledPlugins.update();assert.match(c.NPInstalledPlugins.render(),/manual repair required before starting/);
});
test('uninstall uses delete permission independently of edit permission and rechecks after confirmation',async()=>{
  const {c,calls}=harness();const perms=new Set(['files.view','files.delete']);c.hasServerPermission=p=>perms.has(p);c.esc=String;c.toast=()=>{};c.sameQolServer=(a,b)=>a.id===b.id;
  c.isModdedEngine=()=>false;
  vm.runInNewContext(part('async function deletePlugin(', 'async function togglePlugin('),c);
  const uninstall=()=>c.NPInstalledPlugins.render().match(/<button[^>]+onclick="deletePlugin[^>]+>Uninstall/)[0];
  assert.doesNotMatch(uninstall(),/disabled/);await c.deletePlugin('demo-0.jar');assert.equal(calls.filter(r=>r.o.method==='DELETE').length,1);
  perms.delete('files.delete');perms.add('files.edit');assert.match(uninstall(),/disabled/);await c.deletePlugin('demo-0.jar');assert.equal(calls.filter(r=>r.o.method==='DELETE').length,1);
  perms.add('files.delete');c.uiConfirm=async()=>{perms.delete('files.delete');return true;};await c.deletePlugin('demo-0.jar');assert.equal(calls.filter(r=>r.o.method==='DELETE').length,1);
});

test('Installed defaults to Tiles, preserves supported choices, and migrates Icons',()=>{
  for(const saved of [null,'list','icons','tiles','invalid','blocked']) {
    const {c}=harness();
    c.localStorage={getItem:()=>{if(saved==='blocked')throw Error('storage unavailable');return saved;}};
    vm.runInNewContext(part('function getViewMode(', 'let viewModeTimer'),c);
    const expected=['list','tiles'].includes(saved)?saved:'tiles';
    assert.match(c.NPInstalledPlugins.render(),new RegExp(`npi-files npi-${expected}`));
    c.S.viewModes.plug_inst='icons';
    assert.match(c.NPInstalledPlugins.render(),new RegExp(`npi-files npi-${expected}`));
  }
});
