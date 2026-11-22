const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const root = path.join(__dirname, '../static');
const source = fs.readFileSync(path.join(root, 'js/plugin-details.js'), 'utf8');
const app = fs.readFileSync(path.join(root, 'js/app.js'), 'utf8');
const section = (a,b) => app.slice(app.indexOf(a),app.indexOf(b,app.indexOf(a)));
const hit = n => ({project_id:'id'+n,slug:'project-'+n,title:'Demo project '+n,description:'Demo API data — a server utility.',downloads:1000+n,author:'Demo creator'});
const project = () => ({project:{id:'id0',slug:'project-0',title:'Demo project 0',description:'Demo API data — authenticate players on your server.',body:'# Authentication\nA **secure** server utility.\n\n- Player sessions\n- [Documentation](https://example.com/wiki)\n\n```yaml\nsessions: true\n```',downloads:325001,followers:501,project_type:'plugin',categories:['utility','management'],loaders:['paper','spigot'],game_versions:['1.21.5','1.21.4'],server_side:'required',client_side:'unsupported',license:{id:'GPL-3.0',name:'GNU GPL v3',url:'https://example.com/license'},published:'2020-01-01',updated:'2026-01-01',source_url:'https://example.com/source',gallery:[]},members:[{user:{username:'demo'},role:'Owner'}],versions:[{id:'v-new',name:'Latest release',version_number:'2.0',date_published:'2026-01-01',version_type:'release',loaders:['paper'],game_versions:['1.21.5'],changelog:'## Improved sessions\nFixes and improvements.',files:[{filename:'demo.jar',url:'https://example.com/demo.jar',size:10000}]},{id:'v-old',name:'Previous release',version_number:'1.0',date_published:'2025-01-01',version_type:'release',loaders:['paper'],game_versions:['1.20'],changelog:'Initial release',files:[]},{id:'v-other',name:'Other loader',loaders:['fabric'],game_versions:['1.21']}],installable_version_ids:['v-new','v-old'],loader:'paper',minecraft_version:'1.20'});
function harness() {
  const calls = [], timers = new Map(); let seq=0;
  const c = {S:{view:'plug',user:{id:1,role:'admin'},token:'one',q:'',plugins:[],lastModrinthHits:[],skipPlugConfirm:true},server:{id:'a',node_id:'local',type:'paper',name:'Demo'},document:{addEventListener(){},getElementById(){return null;},querySelectorAll(){return[];}},
    URL,URLSearchParams,AbortController,Set,console,allowed:true,isSteelEngine:t=>t==='steel',isModdedEngine:t=>t==='fabric',hasServerPermission:()=>c.allowed,curServer:()=>c.server,getViewMode:()=> 'list',R(){c.renders=(c.renders||0)+1;},toast(...args){c.messages.push(args);},messages:[],loadPluginsList:async()=>{},uiConfirm:async()=>true,uiAlert:async()=>{},srvApiPath:(s,p)=>`/api/nodes/${s.node_id}/servers/${s.id}${p}`,setTimeout:(f)=>{timers.set(++seq,f);return seq;},clearTimeout:n=>timers.delete(n)};
  c.window=c;c.scrollY=0;c.scrollTo=()=>{};c.api=async(url,o={})=>{calls.push({url,o});return c.handler?c.handler(url,o):url.includes('/project?')?project():url.includes('/versions?')?project():url.includes('/search?')?{hits:[hit(0)],total_hits:1}:{};};
  vm.runInNewContext(source,c);return{c,calls,timers};
}
test('missing Markdown module falls back to bounded escaped text',()=>{
  const {c}=harness(), render=c.NPPlugins.markdown;
  const html=render('# title\n<script>alert(1)</script>\n[x](javascript:alert) ![x](data:image/svg+xml,test) [safe](https://example.com/)\n**bold** `code`');
  assert.doesNotMatch(html,/<script|href="javascript:|src="data:/);assert.match(html,/&lt;script&gt;/);assert.match(html,/np-markdown-fallback/);assert.match(render('x'.repeat(200000)),/Content shortened/);
  for(const bad of ['javascript:alert(1)','data:text/html,x','file:///x','//example.com','https://user:pass@example.com']) assert.equal(c.NPPlugins.safeURL(bad),'');
});
test('default install omits version id; explicit selected version stays explicit; software not MC gates release',async()=>{
  const {c,calls}=harness();await c.NPPlugins.install('project-0','Title');await c.NPPlugins.install('project-0','Title',null,'v-old');
  assert.deepEqual(calls.filter(x=>x.o.method==='POST').map(x=>JSON.parse(x.o.body)),[{slug:'project-0'},{slug:'project-0',version_id:'v-old'}]);
  const preflights=calls.filter(x=>x.url.includes('/versions?'));assert.match(preflights[0].url,/latest=true/);assert.match(preflights[1].url,/version_id=v-old/);
  await c.NPPlugins.install('project-0','Title',null,'v-other');assert.equal(calls.filter(x=>x.o.method==='POST').length,3,'versions endpoint is the software eligibility authority');
});
test('confirmation, preflight and mutation completion reject server/node/account/token/permission/feature changes',async()=>{
  for(const stage of ['preflight','confirm','post']) for(const change of ['server','node','account','token','permission','feature','reset']) {
    const {c,calls}=harness();let resolve;c.S.skipPlugConfirm=stage!=='confirm';
    c.handler=async(url,o)=>{if((stage==='preflight'&&url.includes('/versions?'))||(stage==='post'&&o.method==='POST'))return new Promise(r=>resolve=r);return project();};
    if(stage==='confirm') c.uiConfirm=()=>new Promise(r=>resolve=r);
    const pending=c.NPPlugins.install('project-0','Title');for(let i=0;i<5;i++)await Promise.resolve();
    if(change==='server')c.server.id='b';if(change==='node')c.server.node_id='other';if(change==='account')c.S.user.id=2;if(change==='token')c.S.token='two';if(change==='permission')c.allowed=false;if(change==='feature')c.NPFeatures={enabled:()=>false};if(change==='reset')c.NPPlugins.reset();
    resolve(stage==='confirm'?true:project());await pending;
    assert.equal(calls.filter(x=>x.o.method==='POST').length,stage==='post'?1:0,stage+'/'+change);assert.equal(c.messages.length,0,stage+'/'+change);
  }
});
test('duplicate project installs are single-flight across slug and project id',async()=>{
  const {c,calls}=harness();c.S.lastModrinthHits=[hit(0)];let resolve;c.handler=()=>new Promise(r=>resolve=r);
  const pending=c.NPPlugins.install('project-0','Title');await c.NPPlugins.install('id0','Title');assert.equal(calls.length,1);resolve({versions:[]});await pending;
});
test('pagination deduplicates, preserves results on failure, retries same offset and resets filters',async()=>{
  const {c,calls}=harness();let fail=true;
  c.handler=async url=>{const q=new URL(url,'http://demo').searchParams,offset=Number(q.get('offset'));if(offset&&fail)throw Error('offline');return{hits:Array.from({length:24},(_,i)=>hit(offset?i+20:i)),total_hits:80};};
  await c.NPPlugins.search();await c.NPPlugins.search(false,true);assert.equal(c.S.lastModrinthHits.length,24);assert.match(c.NPPlugins.hitsHtml(c.S.lastModrinthHits),/Retry/);
  fail=false;await c.NPPlugins.search(false,true);assert.equal(c.S.lastModrinthHits.length,44);assert.ok(calls[1].url.includes('offset=24'));assert.ok(calls[2].url.includes('offset=24'));
  c.S.plugLoader='fabric';await c.NPPlugins.search();assert.equal(c.S.lastModrinthHits.length,24);assert.match(calls.at(-1).url,/offset=0/);
  c.server.type='fabric';c.S.plugLoader='';await c.NPPlugins.search();assert.match(calls.at(-1).url,/loader=&/);
});
test('search requests and detail responses never commit stale identity or query; duplicate loads are ignored',async()=>{
  for(const change of ['query','loader','server','node','account','token']) {
    const {c,calls}=harness();let resolve;c.handler=()=>new Promise(r=>resolve=r);const pending=c.NPPlugins.search();await c.NPPlugins.search();assert.equal(calls.length,1);
    if(change==='query')c.S.q='new';if(change==='loader')c.S.plugLoader='fabric';if(change==='server')c.server.id='b';if(change==='node')c.server.node_id='b';if(change==='account')c.S.user.id=2;if(change==='token')c.S.token='b';
    resolve({hits:[hit(0)],total_hits:1});await pending;assert.equal(c.S.lastModrinthHits.length,0,change);
  }
  const {c}=harness();let resolve;c.handler=()=>new Promise(r=>resolve=r);const pending=c.NPPlugins.open('project-0');c.S.user.id=2;resolve(project());await pending;assert.equal(c.NPPlugins.render(),'');
});

test('Chrome: actual Plugins and Files renderers, theme, navigation, tabs, pagination, install and safety (demo APIs)',{skip:!process.env.NP_LAYOUT_PLAYWRIGHT},async()=>{
  const {chromium}=require(process.env.NP_LAYOUT_PLAYWRIGHT), browser=await chromium.launch({headless:true,executablePath:process.env.NP_LAYOUT_CHROME});
  const index=fs.readFileSync(path.join(root,'index.html'),'utf8');
  const styles=[...index.matchAll(/href="\/css\/([^"]+)"/g)].map(m=>fs.readFileSync(path.join(root,'css',m[1]),'utf8')).join('\n');
  const renderers='window.views = {' + section('  fi() {','  ver() {') + '};';
  try {
    for(const width of [1440,390,320]) for(const theme of ['dark','light']) {
      const page=await browser.newPage({viewport:{width,height:900}}),errors=[];page.on('pageerror',e=>errors.push(e.message));await page.emulateMedia({reducedMotion:'reduce'});
      await page.setContent(`<html data-theme="${theme}"><head><style>${styles}</style></head><body><main id="main"></main><div id="modal-container"></div></body></html>`);
      await page.evaluate(({data,hits})=>{
        window.S={view:'plug',plugTab:'browse',plugins:[],lastModrinthHits:[],q:'',plugLoader:'',plugVer:'',plugSort:'relevance',loading:{},skipPlugConfirm:true,user:{id:1,role:'admin'},token:'demo',fileList:[{name:'world',path:'world',is_dir:true},{name:'server.properties',path:'server.properties',is_dir:false}],files:{},selectedFiles:new Set(),fileSort:{by:'name'},currentDir:''};
        window.demoServer={id:'a',node_id:'local',type:'paper',name:'Demo server'};window.demoData=data;window.demoHits=hits;window.calls=[];window.failAppend=true;window.viewMode='list';window.permission=true;window.confirmResolve=null;window.toasts=[];
        window.curServer=()=>demoServer;window.hasServerPermission=()=>permission;window.isSteelEngine=t=>t==='steel';window.isModdedEngine=t=>t==='fabric';window.engineMeta=t=>({key:t,label:t[0].toUpperCase()+t.slice(1)});window.NPFeatures={enabled:()=>true};
        window.esc=v=>String(v??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));window.srvh=()=>'<p class="mu">Demo API data · Demo server</p>';window.ic=()=>'';window.getViewMode=()=>viewMode;window.getAllMcVersions=()=>['1.21.5','1.20'];window.renderViewModeToggle=()=>'';window.loaderHtml=t=>`<p role="status">${t}</p>`;
        window.refreshOpenVersionSelects=()=>{};window.setPlugSkipConfirm=v=>S.skipPlugConfirm=v;window.searchModrinth=(...a)=>NPPlugins.search(...a);window.searchModrinthDebounced=()=>NPPlugins.debounce();window.renderModrinthHitsHtml=h=>NPPlugins.hitsHtml(h);
        window.loadPluginsList=async()=>{};window.toast=(...a)=>toasts.push(a);window.uiConfirm=()=>new Promise(r=>confirmResolve=r);window.uiAlert=async()=>{};window.srvApiPath=(s,p)=>`/api/nodes/${s.node_id}/servers/${s.id}${p}`;
        window.sortFileItems=a=>a;window.fileListIconHtml=()=>'<span>▤</span>';window.getFileExtension=()=>'';window.formatFileSize=()=>'';window.formatFileDate=()=>'';window.renderFileBulkBarHtml=()=>'';window.openFileTrash=()=>window.trashOpened=true;window.NPWorldConversion={open:()=>window.conversionOpened=true};
        window.toggleSelectAllFiles=checked=>{S.selectedFiles=new Set(checked?S.fileList.map(f=>f.path):[]);};
        window.api=async(url,o={})=>{calls.push({url,...o});if(url.includes('/search?')){const q=new URL(url,'http://demo').searchParams,offset=Number(q.get('offset'));if(offset&&failAppend)throw Error('offline');return{hits:demoHits.slice(offset,offset+24),total_hits:demoHits.length};}if(url.includes('/project?'))return structuredClone(demoData);if(url.includes('/versions?'))return{...structuredClone(demoData),versions:demoData.versions.filter(v=>demoData.installable_version_ids.includes(v.id))};return{};};
        window.R=()=>{NPPlugins.sync();document.getElementById('main').innerHTML=views[S.view]();};
      },{data:project(),hits:Array.from({length:48},(_,i)=>hit(i))});
      for (const asset of ['vendor/marked-18.2.1/marked.umd.js','vendor/dompurify-3.4.18/purify.min.js','js/project-markdown.js']) await page.addScriptTag({content:fs.readFileSync(path.join(root,asset),'utf8')});
      await page.addScriptTag({content:source});await page.addScriptTag({content:renderers});assert.deepEqual(errors,[]);await page.evaluate(async()=>{R();await NPPlugins.search();});
      const options=await page.locator('select[aria-label="Browse by loader"] option').allTextContents();assert.equal(options[0],'Loader: Auto (All)');assert.match(options[1],/Paper — Your server/);assert.equal(options.filter(x=>x.startsWith('Paper')).length,1);
      assert.equal(await page.locator('[data-project-card]').count(),24);
      await page.locator('[data-np-install]').first().click();await page.waitForFunction(()=>calls.some(x=>x.method==='POST'));
      assert.equal(await page.locator('.np-project').count(),0);assert.deepEqual(await page.evaluate(()=>JSON.parse(calls.find(x=>x.method==='POST').body)),{slug:'project-0'});
      await page.locator('[data-np-search="more"]').click();assert.equal(await page.locator('[data-project-card]').count(),24);await page.evaluate(()=>failAppend=false);await page.getByRole('button',{name:'Retry',exact:true}).click();await page.waitForFunction(()=>document.querySelectorAll('[data-project-card]').length===48);
      await page.locator('[data-project-card]').first().evaluate(el=>{el.scrollIntoView({block:'center'});window.savedScroll=document.getElementById('main').scrollTop;window.savedWindowScroll=scrollY;});
      await page.locator('[data-project-card]').first().locator('b').click();await page.locator('.np-project-hero').waitFor();assert.equal(await page.locator('.np-project-hero h1').textContent(),'Demo project 0');
      assert.equal(await page.locator('.np-project-markdown strong').first().textContent(),'secure');
      if(process.env.NP_PLUGIN_ARTIFACTS) {
        const dest=process.env.NP_PLUGIN_ARTIFACTS;fs.mkdirSync(dest,{recursive:true});await page.screenshot({path:path.join(dest,`plugins-${width}-${theme}.png`),fullPage:true});
        if(width===1440&&theme==='dark') fs.writeFileSync(path.join(dest,'plugins-preview.html'),(await page.content()).replace(/<script[\s\S]*?<\/script>/g,''));
      }
      await page.locator('#np-tab-description').focus();await page.keyboard.press('ArrowRight');assert.equal(await page.locator('#np-tab-gallery').getAttribute('aria-selected'),'true');assert.match(await page.locator('#np-project-panel').textContent(),/No gallery/);assert.equal(await page.locator('#np-tab-gallery').evaluate(e=>getComputedStyle(e).outlineStyle),'solid');
      await page.getByRole('tab',{name:'Changelog',exact:true}).click();assert.match(await page.locator('#np-project-panel').textContent(),/Improved sessions/);
      await page.getByRole('tab',{name:'Versions',exact:true}).click();assert.equal(await page.locator('[data-version]').count(),2);await page.locator('[data-version="v-old"]').click();await page.waitForFunction(()=>calls.filter(x=>x.method==='POST').length===2);assert.deepEqual(await page.evaluate(()=>JSON.parse(calls.filter(x=>x.method==='POST')[1].body)),{slug:'project-0',version_id:'v-old'});
      await page.evaluate(()=>{demoData.versions_truncated=true;demoData.versions[0].dependencies=[{project_id:'dependency',filename:'<img onerror=window.pwned=1>',dependency_type:'required',version_id:'dep-version'}];});
      await page.evaluate(async()=>{await NPPlugins.open('project-0',true);});
      await page.getByRole('tab',{name:'Versions',exact:true}).click();
      await page.locator('.np-project-version details').first().locator('summary').click();
      assert.match(await page.locator('#np-project-panel').textContent(),/Showing up to 500 recent releases/);
      assert.match(await page.locator('.np-project-version').first().textContent(),/Dependencies are not installed automatically/);
      assert.equal(await page.locator('.np-project-version [onerror]').count(),0);
      assert.equal(await page.locator('.np-project-version a[href="https://modrinth.com/project/dependency"]').count(),1);
      assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),`${width}/${theme} detail overflow`);
      await page.getByRole('button',{name:'Back to results'}).click();assert.equal(await page.locator('[data-project-card]').count(),48);
      assert.equal(await page.evaluate(()=>document.getElementById('main').scrollTop===savedScroll && scrollY===savedWindowScroll),true,'Back preserves both scroll containers');
      await page.locator('#store-search').fill('changed');await page.waitForTimeout(400);await page.waitForFunction(()=>document.querySelectorAll('[data-project-card]').length===24);assert.ok(await page.evaluate(()=>calls.filter(x=>x.url.includes('/search?')).at(-1).url.includes('offset=0')));
      assert.equal(await page.locator('[data-project-card]').first().evaluate(e=>getComputedStyle(e).animationName),'none');
      await page.locator('select[aria-label="Browse by loader"]').selectOption('fabric');assert.ok(await page.evaluate(()=>calls.filter(x=>x.url.includes('/search?')).at(-1).url.includes('loader=fabric')));
      // Delayed old search and detail requests are ignored by the actual renderer.
      await page.evaluate(async()=>{window.originalApi=api;window.resolveOld=null;window.api=(url,o)=>url.includes('query=held')?new Promise(r=>resolveOld=r):originalApi(url,o);S.q='held';window.oldSearch=NPPlugins.search();S.q='newer';await NPPlugins.search();resolveOld({hits:[{project_id:'stale',title:'STALE RESULT'}],total_hits:1});await oldSearch;api=originalApi;});
      assert.equal(await page.getByText('STALE RESULT',{exact:true}).count(),0);
      await page.evaluate(()=>{window.api=(url,o)=>url.includes('/project?')?Promise.reject(Error('offline')):originalApi(url,o);});await page.locator('[data-np-project]').first().click();await page.getByRole('button',{name:'Retry',exact:true}).waitFor();await page.evaluate(()=>api=originalApi);await page.getByRole('button',{name:'Retry',exact:true}).click();await page.locator('.np-project-hero').waitFor();await page.locator('[data-np-back]').click();
      // Raw HTML, attributes and active schemes from every provider surface are inert.
      await page.evaluate(()=>{demoData.project.title='<img src=x onerror="window.pwned=1">';demoData.project.body='<script>window.pwned=1</script>\n[x](javascript:alert) ![x](data:image/svg+xml,bad)';demoData.project.source_url='javascript:alert(1)';demoData.project.icon_url='data:image/svg+xml,bad';demoData.project.license={name:'<script>x</script>',url:'javascript:x'};demoData.project.gallery=[{url:'javascript:x',title:'<img>'}];demoData.members=[{user:{username:'<svg onload=window.pwned=1>',avatar_url:'javascript:x'},role:'<b>Owner</b>'}];});
      await page.locator('[data-np-project]').first().click();await page.locator('.np-project-hero').waitFor();assert.equal(await page.locator('.np-project script,.np-project [onerror],.np-project [onload],.np-project [href^="javascript:"],.np-project [src^="data:"]').count(),0);assert.equal(await page.evaluate(()=>window.pwned),undefined);
      // No delayed POST after server/account switch during confirmation.
      for(const change of ['server','account']) {
        await page.evaluate(()=>{NPPlugins.reset();S.skipPlugConfirm=false;S.view='plug';R();window.beforePosts=calls.filter(x=>x.method==='POST').length;});
        await page.locator('[data-np-install]').first().click();await page.waitForFunction(()=>!!confirmResolve);
        await page.evaluate(change=>{if(change==='server')demoServer.id+='b';else S.user.id++;confirmResolve(true);confirmResolve=null;},change);await page.waitForTimeout(10);assert.equal(await page.evaluate(()=>calls.filter(x=>x.method==='POST').length===beforePosts),true);
      }
      await page.addScriptTag({content:fs.readFileSync(path.join(root,'js/file-drag.js'),'utf8')});
      await page.evaluate(()=>{window.uploads=[];window.enqueueUploads=async(...args)=>uploads.push(args);window.sameQolServer=(a,b)=>a.id===b.id&&a.node_id===b.node_id;NPFileDrag.init();});
      for(const mode of ['list','tiles','icons']) {
        await page.evaluate(mode=>{S.view='fi';viewMode=mode;S.currentDir='';R();},mode);
        const first=mode==='list'?page.locator('.file-table tbody > tr').first():page.locator(mode==='tiles'?'.vm-tiles-grid > *':'.vm-icons-grid > *').first();assert.match(await first.textContent(),/Trash/);assert.equal(await first.locator('input,[data-file-drag],[data-path]').count(),0);
        await first.evaluate(el=>{const dt=new DataTransfer();dt.items.add(new File(['test'],'not-uploaded.txt'));el.dispatchEvent(new DragEvent('dragover',{bubbles:true,cancelable:true,dataTransfer:dt}));el.dispatchEvent(new DragEvent('drop',{bubbles:true,cancelable:true,dataTransfer:dt}));});
        await page.waitForTimeout(10);assert.equal(await page.evaluate(()=>uploads.length),0,'Trash rejects external drops');
        await page.locator('#chk-all-files').check();assert.deepEqual(await page.evaluate(()=>[...S.selectedFiles]),['world','server.properties']);
        assert.equal(await page.locator('.np-files-toolbar').getByRole('button',{name:'Trash',exact:true}).count(),0);
        await page.getByText('World tools',{exact:true}).click();await page.getByRole('button',{name:'Convert world',exact:true}).click();assert.equal(await page.evaluate(()=>conversionOpened),true);
        await first.getByRole('button').count().then(async n=>{if(n)await first.getByRole('button').click();else await first.click();});assert.equal(await page.evaluate(()=>trashOpened),true);
        await page.evaluate(()=>{S.currentDir='world';R();});assert.equal(await page.locator('.np-trash-entry,.np-world-tools').count(),0);
      }
      assert.deepEqual(errors,[]);await page.close();
    }
  } finally {await browser.close();}
});

test('API context guards suppress stale 401 logout, recovery banners and errors',async()=>{
  let resolve, live=true;const events=[];
  const c=vm.createContext({S:{token:'one'},FormData,DOMException,assertClientRequestAccess(){},fetch:()=>new Promise(r=>resolve=r),logout:()=>events.push('logout'),toast:()=>events.push('toast'),NPWorkspace:{requestResult:()=>events.push('recovery')}});c.window=c;
  vm.runInContext(section('async function api(', 'function curServer('),c);
  const pending=c.api('/api/servers/a/plugins/project',{isCurrent:()=>live});live=false;resolve({status:401});await assert.rejects(pending,{name:'AbortError'});assert.deepEqual(events,[]);
});
test('pagination retained result limit is bounded at 1000',async()=>{
  const {c,calls}=harness();c.handler=async url=>{const offset=Number(new URL(url,'http://demo').searchParams.get('offset'));return{hits:Array.from({length:24},(_,i)=>hit(offset+i)),total_hits:5000};};
  await c.NPPlugins.search();for(let i=0;i<60;i++)await c.NPPlugins.search(false,true);assert.equal(c.S.lastModrinthHits.length,1000);assert.equal(calls.length,42);assert.match(c.NPPlugins.hitsHtml(c.S.lastModrinthHits),/Refine your search/);
});
test('typing away and back cannot leave an aborted search permanently busy',async()=>{
  const {c,calls,timers}=harness();let resolve;c.handler=()=>new Promise(r=>resolve=r);
  const first=c.NPPlugins.search();c.S.q='changed';c.NPPlugins.debounce();resolve({hits:[hit(9)],total_hits:1});await first;
  c.S.q='';c.NPPlugins.debounce();const debounce=[...timers.values()].at(-1);timers.clear();debounce();assert.equal(calls.length,2);resolve({hits:[hit(0)],total_hits:1});for(let i=0;i<5;i++)await Promise.resolve();assert.equal(c.S.lastModrinthHits[0].project_id,'id0');
});
test('Network guidance describes simultaneous protocol-specific listeners',()=>{
  assert.match(app,/Multiple ports can serve this server simultaneously/);
  assert.match(app,/Java on 25565\/TCP and Geyser on 19132\/UDP/);
  assert.doesNotMatch(app,/Extra ports are reservations/);
});

test('partial catalogs retain description and report unknown latest instead of claiming no eligible release',async()=>{
  const {c}=harness(), data=project();
  data.latest_version_status='unknown';data.members=[];data.warnings=[{section:'members',message:'<script>unsafe-provider-message</script>'},{section:'latest_version'},{section:'versions'}];
  c.handler=async()=>data;await c.NPPlugins.open('project-0');
  const html=c.NPPlugins.render();assert.match(html,/Authentication/);assert.match(html,/Version check unavailable/);assert.match(html,/Retry missing data/);
  assert.doesNotMatch(html,/No eligible release|unsafe-provider-message|data-np-install/);
  data.latest_version_status='available';await c.NPPlugins.open('project-0',true);assert.match(c.NPPlugins.render(),/Install latest/);
});

test('catalog errors identify Modrinth without exposing raw provider messages or blaming the server connection',async()=>{
  const {c}=harness();c.handler=async()=>{throw Object.assign(Error('<script>private</script>'),{status:502,data:{code:'plugin_catalog_unavailable'}});};
  await c.NPPlugins.open('project-0');const html=c.NPPlugins.render();
  assert.match(html,/Modrinth could not provide this project/);assert.match(html,/separate from your server connection/);assert.doesNotMatch(html,/Check your connection|private|<script>/);
});
