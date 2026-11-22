const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const root = path.join(__dirname,'../static');
const source = fs.readFileSync(path.join(root,'js/plugin-identities.js'),'utf8');
const installed = fs.readFileSync(path.join(root,'js/installed-plugins.js'),'utf8');
const details = fs.readFileSync(path.join(root,'js/plugin-details.js'),'utf8');
const app = fs.readFileSync(path.join(root,'js/app.js'),'utf8');
const part = (a,b) => app.slice(app.indexOf(a),app.indexOf(b,app.indexOf(a)));
const item = n => ({filename:`opaque-${n}.jar${n===1?'.disabled':''}`,size:2048,modified:'2026-01-01',enabled:n!==1,name:`Embedded utility ${n}`,installed_version:'1.2',match:n===2?'unknown':n===3?'suggested':'hash',project:{id:'p'+n,slug:'project-'+n,title:'Verified demo project '+n,icon_url:'https://fixture.test/icon.png'},latest:{id:'MUST-NOT-AUTHORIZE',version_number:'9'},update_available:true,candidates:[{id:'untrusted',icon_url:'https://fixture.test/candidate.png'}],sha512:'a'.repeat(128)});
const flush = async () => { for(let i=0;i<400;i++) await Promise.resolve(); };
function harness(count=4) {
  const calls=[],timers=new Map(),events={};let seq=0;
  const data=Array.from({length:count},(_,n)=>item(n));
  const c={S:{view:'plug',plugTab:'installed',status:'offline',user:{id:1,role:'admin'},token:'token',plugins:data.map(r=>[r.filename,'2 KB',r.enabled,r.modified,false,r.size]),fileList:data.map(r=>({name:r.filename,path:'plugins/'+r.filename,size:r.size,modified:r.modified})),currentDir:'plugins',loading:{}},server:{id:'a',node_id:'local',type:'paper',version:'1.21'},permission:true,feature:true,URL,AbortController,console,Set,Map,Date,
    document:{addEventListener:(k,f)=>(events[k]??=[]).push(f),getElementById:()=>null,querySelectorAll:()=>[]},curServer:()=>c.server,isSteelEngine:t=>t==='steel',isModdedEngine:t=>t==='fabric',hasServerPermission:p=>c.permission&&(!c.readonly||p==='files.view'),NPFeatures:{enabled:()=>c.feature},getViewMode:()=> 'list',formatFileDate:()=>'',R(){c.NPPluginIdentities.sync();c.NPInstalledPlugins?.sync();},loadPluginsList:async()=>{},uiConfirm:async()=>true,srvApiPath:(s,p)=>`/api/nodes/${s.node_id}/servers/${s.id}${p}`,setTimeout:(f,ms)=>{timers.set(++seq,{f,ms});return seq;},clearTimeout:id=>timers.delete(id)};
  c.window=c;c.api=async(url,o={})=>{calls.push({url,o});if(c.handler)return c.handler(url,o);const offset=Number(new URL(url,'http://demo').searchParams.get('offset'));return{items:data.slice(offset,offset+20),offset,total:data.length,has_more:offset+20<data.length};};
  vm.runInNewContext(source,c);vm.runInNewContext(installed,c);
  return {c,calls,data,timers,events};
}
test('first Installed render automatically loads sequential pages; no update authority, polling or duplicate requests',async()=>{
  const {c,calls}=harness(44);assert.match(c.NPInstalledPlugins.render(),/Identifying installed files/);await flush();
  assert.equal(calls.length,3);assert.deepEqual(calls.map(x=>new URL(x.url,'http://demo').searchParams.get('offset')),['0','20','40']);assert.ok(calls.every(x=>x.url.endsWith('limit=20&metadata_only=true')));
  const html=c.NPInstalledPlugins.render();assert.match(html,/Verified demo project 0/);assert.match(html,/Embedded utility 2/);assert.match(html,/Latest for software: Not checked/);assert.doesNotMatch(html,/MUST-NOT-AUTHORIZE|candidate.png/);
  c.NPInstalledPlugins.select('opaque-0.jar',true);assert.match(c.NPInstalledPlugins.render(),/0 \/ 10 selected/);assert.equal(c.NPPluginIdentities.record('opaque-0.jar').latest,undefined);
  for(let n=0;n<8;n++){c.R();c.NPInstalledPlugins.render();}await flush();assert.equal(calls.length,3);
  c.S.view='fi';c.R();await flush();assert.equal(calls.length,3,'Files reuses identical file fingerprint');assert.match(c.NPPluginIdentities.fileTitle(c.S.fileList[0]),/Verified demo project 0/);
});
test('1000-record cap and each page has a cancellable 75-second deadline',async()=>{
  const {c,calls,timers}=harness(1020);c.R();await flush();assert.equal(calls.length,50);assert.match(c.NPPluginIdentities.statusHtml(),/1000-file display limit/);
  c.NPPluginIdentities.invalidate();c.handler=(u,o)=>new Promise((resolve,reject)=>o.signal.addEventListener('abort',()=>reject(Error('timeout'))));c.R();await Promise.resolve();assert.ok([...timers.values()].some(t=>t.ms===75000));[...timers.values()].find(t=>t.ms===75000).f();await flush();assert.match(c.NPPluginIdentities.statusHtml(),/Retry identities/);c.R();await flush();assert.equal(calls.length,51);
});
test('exact directory, policy and view scope; read-only may identify, feature-off never leaks metadata',async()=>{
  for(const dir of ['','plugins/config','other/plugins','mods','plugins//','/plugins','plugins/../plugins']){
    const {c,calls}=harness();c.S.view='fi';c.S.currentDir=dir;c.R();await flush();assert.equal(calls.length,0,dir);assert.equal(c.NPPluginIdentities.fileTitle(c.S.fileList[0]),'opaque-0.jar');
  }
  for(const change of [c=>c.S.plugTab='browse',c=>c.feature=false,c=>c.permission=false,c=>c.server.type='steel',c=>c.S.editingFile='plugins/a.yml']){
    const {c,calls}=harness();if(change.toString().includes('editingFile'))c.S.view='fi';change(c);c.R();await flush();assert.equal(calls.length,0);
  }
  const {c,calls}=harness();c.readonly=true;c.R();await flush();assert.equal(calls.length,1);assert.match(c.NPInstalledPlugins.render(),/Read-only/);c.feature=false;c.R();assert.equal(c.NPPluginIdentities.record('opaque-0.jar'),undefined);assert.doesNotMatch(c.NPInstalledPlugins.render(),/Verified demo/);
  const mod=harness();mod.c.server.type='fabric';mod.c.S.view='fi';mod.c.S.currentDir='mods/';mod.c.S.fileList.forEach(p=>p.path=p.path.replace('plugins/','mods/'));mod.c.R();await flush();assert.equal(mod.calls.length,1);
});
test('stale responses abort and cannot commit after all identity/navigation/file changes',async()=>{
  for(const change of ['server','node','account','token','software','version','view','tab','folder','list','loading','permission','feature','editor','detail']){
    const {c,calls,data}=harness();let resolve;c.handler=()=>new Promise(r=>resolve=r);if(['folder','editor'].includes(change))c.S.view='fi';c.R();await Promise.resolve();
    if(change==='server')c.server.id='b';if(change==='node')c.server.node_id='b';if(change==='account')c.S.user.id=2;if(change==='token')c.S.token='b';if(change==='software')c.server.type='fabric';if(change==='version')c.server.version='next';if(change==='view')c.S.view='console';if(change==='tab')c.S.plugTab='browse';if(change==='folder')c.S.currentDir='plugins/config';if(change==='list')c.S.plugins[0][3]='changed';if(change==='loading')c.S.loading.plugins=true;if(change==='permission')c.permission=false;if(change==='feature')c.feature=false;if(change==='editor')c.S.editingFile='plugins/a.yml';if(change==='detail')c.NPPlugins={isOpen:()=>true};
    c.R();assert.equal(calls[0].o.signal.aborted,true,change);const old=resolve;c.handler=async()=>{throw Error('new context unavailable');};old({items:data,offset:0,total:4,has_more:false});await flush();assert.equal(c.NPPluginIdentities.record(data[0].filename),undefined,change);
  }
});
test('fingerprint mutation invalidates reused metadata; errors require explicit retry; URLs and descriptor names are safe',async()=>{
  const {c,calls,data}=harness();c.R();await flush();c.S.plugins[0][5]++;c.R();await flush();assert.equal(calls.length,2);
  c.NPPluginIdentities.invalidate();c.handler=async()=>{throw Error('/private/path');};c.R();await flush();assert.match(c.NPPluginIdentities.statusHtml(),/Retry identities/);assert.doesNotMatch(c.NPPluginIdentities.statusHtml(),/private/);for(let i=0;i<4;i++)c.R();await flush();assert.equal(calls.length,3);
  c.handler=null;await c.NPPluginIdentities.retry();assert.equal(calls.length,4);
  for(const bad of ['javascript:alert(1)','data:image/svg+xml,test','file:///x','//example.com','https://user:pass@example.com'])assert.equal(c.NPPluginIdentities.safeURL(bad),'');
  data[0].project.title='<img onerror=alert(1)>';data[0].project.icon_url='https://user:pass@example.com';data[2].name='<script>evil</script>';await c.NPPluginIdentities.retry();const html=c.NPInstalledPlugins.render();assert.match(html,/&lt;img/);assert.match(html,/&lt;script&gt;/);assert.doesNotMatch(html,/<script|<img onerror|user:pass/);
  c.S.view='fi';c.R();assert.equal(c.NPPluginIdentities.fileTitle({name:'opaque-0.jar',path:'elsewhere/opaque-0.jar'}),'opaque-0.jar');
});
test('automatic identity success never clears the uncertain-write barrier',async()=>{
  const {c}=harness();await c.NPInstalledPlugins.scan();c.NPInstalledPlugins.select('opaque-0.jar',true);c.handler=async()=>{throw Error('lost response');};await c.NPInstalledPlugins.update();assert.match(c.NPInstalledPlugins.render(),/Previous write outcome is uncertain/);
  c.handler=null;await c.NPPluginIdentities.retry();assert.match(c.NPInstalledPlugins.render(),/Previous write outcome is uncertain/);assert.match(c.NPInstalledPlugins.render(),/0 \/ 10 selected/);await c.NPInstalledPlugins.scan();assert.doesNotMatch(c.NPInstalledPlugins.render(),/Previous write outcome is uncertain/);
});

test('in-flight renders are singleflight; successful cache expires once, failed cache does not auto-retry',async()=>{
  const {c,calls,data}=harness();let now=1000,resolve;c.Date={now:()=>now};c.handler=()=>new Promise(r=>resolve=r);c.NPInstalledPlugins.render();await Promise.resolve();for(let i=0;i<10;i++)c.NPInstalledPlugins.render();await Promise.resolve();assert.equal(calls.length,1);resolve({items:data,offset:0,total:4,has_more:false});await flush();c.handler=null;now+=300001;c.R();await flush();assert.equal(calls.length,2);c.R();await flush();assert.equal(calls.length,2);
  c.handler=async()=>{throw Error('offline');};await c.NPPluginIdentities.retry();now+=600000;c.R();await flush();assert.equal(calls.length,3);assert.match(c.NPPluginIdentities.statusHtml(),/Retry identities/);
});

test('explicit scans and updates cancel automatic reads and suspend them through dialogs and writes',async()=>{
  const {c,calls,data}=harness();let resolveScan,resolveWrite,confirm;
  c.handler=(url,o)=>url.includes('metadata_only')?new Promise((r,j)=>o.signal.addEventListener('abort',()=>j(Error('cancelled')))):new Promise(r=>resolveScan=r);
  c.R();await flush();assert.equal(calls.length,1);
  const scan=c.NPInstalledPlugins.scan();assert.equal(calls[0].o.signal.aborted,true);c.R();await flush();assert.equal(calls.length,2);
  resolveScan({items:data,offset:0,total:4,has_more:false});await scan;await flush();
  c.NPInstalledPlugins.select('opaque-0.jar',true);c.uiConfirm=()=>new Promise(r=>confirm=r);
  c.handler=(url,o)=>url.endsWith('/plugins/update')?new Promise(r=>resolveWrite=r):new Promise((r,j)=>o.signal.addEventListener('abort',()=>j(Error('cancelled'))));
  const before=calls.length;const update=c.NPInstalledPlugins.update();c.R();await flush();assert.equal(calls.length,before);assert.match(c.NPPluginIdentities.statusHtml(),/current file operation/);
  confirm(true);await flush();assert.equal(calls.length,before+1);c.R();await flush();assert.equal(calls.length,before+1);
  resolveWrite({results:[{filename:'opaque-0.jar',status:'updated'}],restart_required:true});await update;await flush();assert.equal(calls.length,before+2);
  c.NPPluginIdentities.invalidate();
});

test('active upload batches defer identities across changing file fingerprints until settled',async()=>{
  const {c,calls}=harness();let pending=true;c.uploadsPendingFor=(server,dir)=>pending&&server.id==='a'&&dir==='plugins';
  c.S.view='fi';c.R();await flush();assert.equal(calls.length,0);
  for(let i=0;i<3;i++){c.S.fileList[0].modified='upload-'+i;c.R();await flush();}assert.equal(calls.length,0);
  pending=false;c.R();await flush();assert.equal(calls.length,1);c.R();await flush();assert.equal(calls.length,1);
});

test('Chrome: automatic identities in real Files/Installed renderers, loading card and Trash across themes/modes/320px', {skip:!process.env.NP_LAYOUT_PLAYWRIGHT},async()=>{
  const {chromium}=require(process.env.NP_LAYOUT_PLAYWRIGHT);const browser=await chromium.launch({headless:true,executablePath:process.env.NP_LAYOUT_CHROME});
  const styles=[...fs.readFileSync(path.join(root,'index.html'),'utf8').matchAll(/href="\/css\/([^"]+)"/g)].map(m=>fs.readFileSync(path.join(root,'css',m[1]),'utf8')).join('\n');
  try {
    const page=await browser.newPage({viewport:{width:1440,height:1000}}),errors=[];page.on('pageerror',e=>errors.push(e.message));
    await page.route('https://fixture.test/**',route=>route.fulfill({status:200,contentType:'image/png',body:Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII=','base64')}));
    await page.setContent(`<html data-theme="dark"><head><style>${styles}</style></head><body><main id="main"></main></body></html>`);
    await page.evaluate(data=>{
      window.demoData=data;window.S={view:'plug',plugTab:'installed',status:'offline',user:{id:1,role:'admin'},token:'demo',plugins:data.map(r=>[r.filename,'2 KB',r.enabled,r.modified,false,r.size]),fileList:data.map(r=>({name:r.filename,path:'plugins/'+r.filename,size:r.size,modified:r.modified})),selectedFiles:new Set(),currentDir:'plugins',lastModrinthHits:[{slug:'project-0',title:'Verified demo project 0',description:'Demo API data — safe clicked-result preview.',icon_url:'https://fixture.test/icon.png'}],loading:{}};
      window.server={id:'a',node_id:'local',type:'paper',name:'Demo server'};window.mode='list';window.feature=true;window.readonly=false;window.calls=[];window.hold=true;window.pending=[];
      window.curServer=()=>server;window.hasServerPermission=p=>!readonly||p==='files.view';window.isSteelEngine=()=>false;window.isModdedEngine=t=>t==='fabric';window.NPFeatures={enabled:()=>feature};window.getViewMode=()=>mode;window.formatFileDate=()=> 'Jan 1';window.formatFileSize=()=> '2 KB';window.getFileExtension=()=> 'jar';window.renderViewModeToggle=()=>'';window.sortFileItems=l=>l;window.renderFileBulkBarHtml=()=>'';window.fileListIconHtml=()=>'<span class="fi-badge">JAR</span>';window.srvh=()=>'<p>Demo API data · identities test</p>';window.loaderHtml=s=>s;window.$=q=>document.querySelector(q);window.esc=v=>String(v??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
      window.srvApiPath=(s,p)=>`/api/nodes/${s.node_id}/servers/${s.id}${p}`;window.toggleFileSelect=(p,on)=>on?S.selectedFiles.add(p):S.selectedFiles.delete(p);window.openFile=n=>window.openedFile=n;window.deletePlugin=n=>window.deletedFile=n;window.togglePlugin=(i,on)=>window.toggled=[S.plugins[i][0],on];window.openFileTrash=()=>window.trashOpened=true;
      window.api=async(url,o={})=>{calls.push({url,method:o.method});if(url.includes('/project?'))return new Promise(r=>window.resolveProject=r);const offset=Number(new URL(url,'http://demo').searchParams.get('offset'));if(hold)await new Promise(r=>pending.push(r));if(window.failMetadata)throw Error('offline');return {items:structuredClone(demoData.slice(offset,offset+20)),offset,total:demoData.length,has_more:offset+20<demoData.length};};
      window.R=()=>{NPPlugins?.sync();NPPluginIdentities.sync();NPInstalledPlugins.sync();document.getElementById('main').innerHTML=S.view==='fi'?views.fi():NPPlugins.render()||NPInstalledPlugins.render();};
    },Array.from({length:24},(_,n)=>item(n)));
    // Use the actual icon function and Files renderer, without bootstrapping the live service.
    await page.addScriptTag({content:part('function ic(', 'function fileListIconHtml(')});
    await page.addScriptTag({content:details});await page.addScriptTag({content:source});await page.addScriptTag({content:installed});await page.addScriptTag({content:'window.views={'+part('  fi() {','  plug() {')+'};'});
    await page.evaluate(()=>R());await page.getByText(/Identifying installed files/).waitFor();assert.equal(await page.locator('.npi-project').count(),0);assert.equal(await page.locator('[data-npi-select]:enabled').count(),0);
    await page.evaluate(()=>{hold=false;pending.splice(0).forEach(r=>r());});await page.getByText(/24 file identities checked/).waitFor();assert.equal(await page.locator('.npi-project').count(),22);assert.equal(await page.locator('[data-npi-select]:enabled').count(),0);assert.equal(await page.evaluate(()=>calls.length),2);
    await page.locator('[data-plugin="opaque-0.jar"]').click();assert.equal(await page.evaluate(()=>deletedFile),'opaque-0.jar');await page.getByRole('checkbox',{name:'Enable opaque-0.jar',exact:true}).focus();await page.keyboard.press('Space');assert.deepEqual(await page.evaluate(()=>toggled),['opaque-0.jar',false]);
    await page.evaluate(()=>{S.view='fi';R();});assert.equal(await page.evaluate(()=>calls.length),2);assert.equal(await page.locator('[data-npid-title]').count(),24);await page.locator('.np-file-open[data-name="opaque-0.jar"]').click();assert.equal(await page.evaluate(()=>openedFile),'opaque-0.jar');
    await page.locator('.file-chk').first().check();await page.locator('.file-chk').first().focus();await page.evaluate(()=>{window.firstCheckbox=document.querySelector('.file-chk');window.hold=true;void NPPluginIdentities.retry();});await page.evaluate(()=>{hold=false;pending.splice(0).forEach(r=>r());});await page.getByText(/24 file identities checked/).waitFor();assert.ok(await page.evaluate(()=>firstCheckbox===document.querySelector('.file-chk')&&firstCheckbox===document.activeElement&&firstCheckbox.checked));
    for(const width of [1440,390,320])for(const theme of ['dark','light'])for(const mode of ['list','tiles','icons']){
      await page.setViewportSize({width,height:900});await page.emulateMedia({reducedMotion:'reduce'});await page.evaluate(({mode,theme})=>{window.mode=mode;document.documentElement.dataset.theme=theme;R();},{mode,theme});
      assert.equal(await page.locator('[data-npid-title]').count(),24);assert.ok(await page.locator('[data-npid-title]').evaluateAll(els=>els.every(el=>el.scrollWidth<=el.clientWidth+1)),`identity wrapping ${width}/${theme}/${mode}`);
      if(process.env.NP_PLUGIN_ARTIFACTS && width===390 && theme==='dark'){fs.mkdirSync(process.env.NP_PLUGIN_ARTIFACTS,{recursive:true});await page.screenshot({path:path.join(process.env.NP_PLUGIN_ARTIFACTS,`identities-files-${mode}.png`),fullPage:true});fs.writeFileSync(path.join(process.env.NP_PLUGIN_ARTIFACTS,`identities-files-${mode}.html`),(await page.content()).replace(/<script[\s\S]*?<\/script>/g,''));}
      await page.evaluate(()=>{S.currentDir='';S.fileList=[];R();});assert.equal(await page.locator('.np-trash-entry').count(),1);assert.equal(await page.locator('.np-trash-entry [data-drop-path],.np-trash-entry input,.np-trash-entry[data-file-drag]').count(),0);assert.equal(await page.locator('.np-trash-badge svg').count(),1);assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),`trash overflow ${width}/${theme}/${mode}`);
      const trash=page.locator('.np-trash-entry button,.np-trash-entry:is(button)').first();await page.keyboard.press('Tab');await trash.focus();assert.equal(await trash.evaluate(el=>getComputedStyle(el).outlineStyle),'solid');await page.keyboard.press('Enter');assert.equal(await page.evaluate(()=>trashOpened),true);
      await page.evaluate(()=>{S.currentDir='plugins';S.fileList=demoData.map(r=>({name:r.filename,path:'plugins/'+r.filename,size:r.size,modified:r.modified}));S.view='plug';R();});assert.ok(await page.locator('.npi-file').evaluateAll(els=>els.every(el=>el.scrollWidth<=el.clientWidth+1)),`installed overflow ${width}/${theme}/${mode}`);
      if(process.env.NP_PLUGIN_ARTIFACTS && width===390 && theme==='dark')await page.screenshot({path:path.join(process.env.NP_PLUGIN_ARTIFACTS,`identities-installed-${mode}.png`),fullPage:true});
      await page.evaluate(()=>{S.view='fi';R();});
    }
    await page.evaluate(async()=>{failMetadata=true;await NPPluginIdentities.retry();});await page.getByRole('button',{name:'Retry identities'}).waitFor();const failedCalls=await page.evaluate(()=>calls.length);await page.evaluate(()=>{R();R();});assert.equal(await page.evaluate(()=>calls.length),failedCalls);await page.evaluate(()=>failMetadata=false);await page.getByRole('button',{name:'Retry identities'}).click();await page.getByText(/24 file identities checked/).waitFor();
    await page.evaluate(async()=>{demoData[0].project.title='<img onerror=window.pwned=1>';demoData[0].project.icon_url='javascript:alert(1)';demoData[2].name='<svg onload=window.pwned=1>';await NPPluginIdentities.retry();});assert.equal(await page.locator('[data-npid-title] img,[data-npid-title] svg,[src^="javascript:"]').count(),0);assert.equal(await page.evaluate(()=>window.pwned),undefined);
    await page.evaluate(()=>{const img=document.querySelector('[data-npid-image]');img?.dispatchEvent(new Event('error'));});assert.equal(await page.locator('[data-npid-icon] img[src="https://fixture.test/icon.png"]').count() < 22,true);
    const cachedCalls=await page.evaluate(()=>calls.length);await page.evaluate(()=>{S.currentDir='plugins/config';R();});assert.equal(await page.locator('[data-npid-title],[data-npid-feedback]').count(),0);assert.equal(await page.evaluate(()=>calls.length),cachedCalls);await page.evaluate(()=>{S.currentDir='plugins';S.fileList[0].modified='changed';R();});await page.waitForFunction(n=>calls.length>n,cachedCalls);
    await page.evaluate(()=>{feature=false;R();});assert.equal(await page.locator('[data-npid-title],[data-npid-image],[data-npid-feedback]').count(),0);await page.evaluate(()=>{feature=true;readonly=true;S.currentDir='';R();});assert.ok(await page.locator('.np-trash-entry button,.np-trash-entry:is(button)').first().isDisabled());
    await page.evaluate(()=>{readonly=false;S.view='plug';S.plugTab='browse';void NPPlugins.open('project-0');});await page.locator('.np-project-loading').waitFor();assert.match(await page.locator('.np-project-loading').textContent(),/safe clicked-result preview/);assert.equal(await page.locator('.np-project-loading [data-np-install]').count(),0);
    for(const width of [1440,390,320])for(const theme of ['dark','light']){
      await page.setViewportSize({width,height:900});await page.evaluate(theme=>document.documentElement.dataset.theme=theme,theme);assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),`loading overflow ${width}/${theme}`);assert.equal(await page.locator('.np-project-loading').evaluate(el=>getComputedStyle(el).animationName),'none');assert.equal(await page.locator('.np-project-loading [role="status"]').count(),1);
      if(process.env.NP_PLUGIN_ARTIFACTS){fs.mkdirSync(process.env.NP_PLUGIN_ARTIFACTS,{recursive:true});await page.screenshot({path:path.join(process.env.NP_PLUGIN_ARTIFACTS,`identity-loading-${width}-${theme}.png`),fullPage:true});}
    }
    if(process.env.NP_PLUGIN_ARTIFACTS)fs.writeFileSync(path.join(process.env.NP_PLUGIN_ARTIFACTS,'identity-loading-preview.html'),(await page.content()).replace(/<script[\s\S]*?<\/script>/g,''));
    await page.getByRole('button',{name:'Back to results'}).click();assert.equal(await page.locator('.np-project-loading').count(),0);
    assert.deepEqual(errors,[]);
  } finally {await browser.close();}
});
