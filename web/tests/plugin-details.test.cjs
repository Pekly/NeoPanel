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
function catalog(url, total=83) {
  const q=new URL(url,'http://demo').searchParams,offset=Number(q.get('offset')),size=Number(q.get('limit'));
  return {hits:Array.from({length:Math.max(0,Math.min(size,total-offset))},(_,i)=>hit(offset+i)),total_hits:total};
}
test('numbered pages replace results; failure labels old page and retries intended offset',async()=>{
  const {c,calls}=harness();let fail=true;
  c.handler=async url=>{if(new URL(url,'http://demo').searchParams.get('offset')!=='0'&&fail)throw Error('offline');return catalog(url);};
  await c.NPPlugins.search();assert.equal(c.S.lastModrinthHits.length,20);await c.NPPlugins.goPage(2);
  let html=c.NPPlugins.hitsHtml(c.S.lastModrinthHits);assert.match(html,/1–20 of 83/);assert.match(html,/Page 2 could not be loaded/);assert.match(html,/aria-current="page"/);
  fail=false;await c.NPPlugins.search();assert.equal(c.S.lastModrinthHits.length,20);assert.equal(c.S.lastModrinthHits[0].project_id,'id20');
  assert.match(calls[1].url,/offset=20/);assert.equal(calls[1].url,calls[2].url);
  await c.NPPlugins.goPage(5);assert.equal(c.S.lastModrinthHits.length,3);assert.match(c.NPPlugins.hitsHtml(c.S.lastModrinthHits),/81–83 of 83/);
  const before=calls.length;await c.NPPlugins.goPage(6);await c.NPPlugins.goPage(0);await c.NPPlugins.goPage(5);assert.equal(calls.length,before);
});
test('page sizes, filter and identity resets, and detail Back preserve browsing state',async()=>{
  const {c,calls}=harness();c.handler=async url=>url.includes('/project?')?project():catalog(url,257);
  await c.NPPlugins.search();
  for(const size of [10,50,100,20]) {
    await c.NPPlugins.setPageSize(size);assert.match(calls.at(-1).url,new RegExp(`limit=${size}&offset=0`));
    await c.NPPlugins.goPage(2);assert.equal(c.S.lastModrinthHits.length,size);assert.equal(c.S.lastModrinthHits[0].project_id,'id'+size);
    const before=calls.length;await c.NPPlugins.open('project-'+size);c.NPPlugins.back();assert.equal(calls.length,before+1);assert.match(c.NPPlugins.hitsHtml(c.S.lastModrinthHits),/Page 2 of/);assert.match(c.NPPlugins.sizeControl(),new RegExp(`value="${size}" selected`));
  }
  for(const field of ['q','plugLoader','plugVer','plugSort']) {c.S[field]='changed';await c.NPPlugins.search();assert.match(calls.at(-1).url,/offset=0/);await c.NPPlugins.goPage(2);}
  c.server.id='b';await c.NPPlugins.search();assert.match(calls.at(-1).url,/offset=0/);
  const before=calls.length;for(const value of [-1,0,24,1000,'evil'])await c.NPPlugins.setPageSize(value);assert.equal(calls.length,before);
});
test('size failure retains old size range and retry uses new size; empty and shrinking catalogs',async()=>{
  const {c,calls}=harness();c.handler=async url=>catalog(url);await c.NPPlugins.search();await c.NPPlugins.goPage(4);
  c.handler=async()=>{throw Error('offline');};await c.NPPlugins.setPageSize(50);assert.match(c.NPPlugins.hitsHtml(c.S.lastModrinthHits),/61–80 of 83 projects · Page 4 of 5/);
  const before=calls.length;await c.NPPlugins.goPage(2);assert.equal(calls.length,before,'old page controls cannot navigate with the failed new size');
  assert.ok([...c.NPPlugins.hitsHtml(c.S.lastModrinthHits).matchAll(/<button[^>]*data-np-page="\d+"[^>]*>/g)].every(m=>m[0].includes('disabled')));
  c.handler=async url=>catalog(url);await c.NPPlugins.search();assert.match(calls.at(-1).url,/limit=50&offset=0/);
  c.handler=async url=>catalog(url,3);await c.NPPlugins.goPage(2);assert.match(c.NPPlugins.hitsHtml(c.S.lastModrinthHits),/1–3 of 3/);
  c.handler=async url=>catalog(url,0);c.S.q='empty';await c.NPPlugins.search();assert.match(c.NPPlugins.hitsHtml(c.S.lastModrinthHits),/0–0 of 0/);assert.match(c.NPPlugins.hitsHtml(c.S.lastModrinthHits),/No matching/);
});
test('artwork prefers safe featured then first safe gallery; no active URL or attribute injection',()=>{
  const {c}=harness();
  for(const bad of ['javascript:alert(1)','data:image/svg+xml,bad','file:///x','//evil.test','https://user:pass@evil.test']) {
    const html=c.NPPlugins.hitsHtml([{...hit(0),featured_gallery:bad,gallery:[bad,'https://demo.test/fallback?a=1&b=2']}]);
    assert.match(html,/src="https:\/\/demo.test\/fallback\?a=1&amp;b=2"/);assert.doesNotMatch(html,/src="(?:javascript|data|file):/);
  }
  const h={...hit(0),featured_gallery:'https://demo.test/featured',gallery:['https://demo.test/other']};
  assert.match(c.NPPlugins.hitsHtml([h]),/src="https:\/\/demo.test\/featured" alt="" aria-hidden="true" loading="lazy" referrerpolicy="no-referrer"/);
  h.featured_gallery='https://demo.test/" onerror="evil';assert.doesNotMatch(c.NPPlugins.hitsHtml([h]),/ onerror=/);
  c.getViewMode=()=> 'tiles';assert.match(c.NPPlugins.hitsHtml([h]),/np-project-banner/);
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
      await page.route('https://demo-art.test/**',route=>route.request().url().includes('broken')?route.abort():route.fulfill({contentType:'image/svg+xml',body:'<svg xmlns="http://www.w3.org/2000/svg" width="900" height="240"><rect width="900" height="240" fill="#6694ac"/><path d="M0 220L150 60 260 140 460 20 680 160 900 40V240H0" fill="#285745"/><path d="M0 230L200 170 400 210 630 90 900 200V240H0" fill="#a7cc75"/></svg>'}));
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
        window.api=async(url,o={})=>{calls.push({url,...o});if(url.includes('/search?')){const q=new URL(url,'http://demo').searchParams,offset=Number(q.get('offset')),size=Number(q.get('limit'));if(offset&&failAppend)throw Error('offline');return{hits:demoHits.slice(offset,offset+size),total_hits:demoHits.length};}if(url.includes('/project?'))return structuredClone(demoData);if(url.includes('/versions?'))return{...structuredClone(demoData),versions:demoData.versions.filter(v=>demoData.installable_version_ids.includes(v.id))};return{};};
        window.R=()=>{NPPlugins.sync();document.getElementById('main').innerHTML=views[S.view]();};
      },{data:project(),hits:Array.from({length:48},(_,i)=>({...hit(i),featured_gallery:i===1?'https://demo-art.test/broken':'',gallery:['javascript:bad','https://demo-art.test/banner.svg']}))});
      for (const asset of ['vendor/marked-18.2.1/marked.umd.js','vendor/dompurify-3.4.18/purify.min.js','js/project-markdown.js']) await page.addScriptTag({content:fs.readFileSync(path.join(root,asset),'utf8')});
      await page.addScriptTag({content:source});await page.addScriptTag({content:renderers});assert.deepEqual(errors,[]);await page.evaluate(async()=>{R();await NPPlugins.search();});
      const options=await page.locator('select[aria-label="Browse by loader"] option').allTextContents();assert.equal(options[0],'Loader: Auto (All)');assert.match(options[1],/Paper — Your server/);assert.equal(options.filter(x=>x.startsWith('Paper')).length,1);
      assert.equal(await page.locator('[data-project-card]').count(),20);
      await page.locator('[data-np-install]').first().click();await page.waitForFunction(()=>calls.some(x=>x.method==='POST'));
      assert.equal(await page.locator('.np-project').count(),0);assert.deepEqual(await page.evaluate(()=>JSON.parse(calls.find(x=>x.method==='POST').body)),{slug:'project-0'});
      await page.getByRole('button',{name:'Page 2',exact:true}).click();assert.equal(await page.locator('[data-project-card]').count(),20);assert.match(await page.locator('.np-project-page-status').textContent(),/1–20 of 48/);
      await page.evaluate(()=>failAppend=false);await page.getByRole('button',{name:'Retry',exact:true}).click();await page.waitForFunction(()=>document.querySelector('[data-project-card]').dataset.slug==='project-20');
      await page.getByRole('button',{name:'Next page',exact:true}).focus();await page.keyboard.press('Enter');await page.waitForFunction(()=>document.querySelectorAll('[data-project-card]').length===8);assert.equal(await page.getByRole('button',{name:'Next page'}).isDisabled(),true);
      assert.equal(await page.evaluate(()=>document.activeElement.getAttribute('aria-current')),'page');assert.equal(await page.locator('[aria-current=page]').evaluate(e=>getComputedStyle(e).outlineStyle),'solid');
      await page.getByRole('combobox',{name:'Projects per page'}).selectOption('10');await page.waitForFunction(()=>document.querySelectorAll('[data-project-card]').length===10);
      await page.getByRole('combobox',{name:'Projects per page'}).selectOption('20');await page.waitForFunction(()=>document.querySelectorAll('[data-project-card]').length===20);
      await page.evaluate(()=>{demoHits[2].title='LongProjectName'.repeat(8);demoHits[2].description='Long demo description to verify stable reflow. '.repeat(12);R();});
      for(const mode of ['list','tiles']) {
        await page.evaluate(mode=>{viewMode=mode;R();},mode);
        assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),`${width}/${theme}/${mode} browse overflow`);
        assert.equal(await page.locator('.np-project-banner').count(),20);
        assert.ok(await page.locator('.np-project-pages').evaluate(el=>el.getBoundingClientRect().height<=44 && el.scrollWidth<=el.clientWidth+1),'pagination is a compact horizontal row');
        assert.ok(await page.locator('#store-grid').evaluate(el=>[el,...el.querySelectorAll('.np-project-result,.np-project-open,.np-project-result-foot')].every(n=>n.scrollWidth<=n.clientWidth+1)),`${width}/${theme}/${mode} no internal clipping or overflow`);
          await page.locator('.np-project-banner').first().scrollIntoViewIfNeeded();await page.waitForFunction(()=>document.querySelector('.np-project-banner').naturalWidth>0);
          await page.locator('.np-project-banner').nth(1).evaluate(el=>el.closest('article').scrollIntoView());await page.waitForFunction(()=>document.querySelectorAll('.np-project-banner')[1].hidden);
          assert.equal(await page.locator('.np-project-banner').first().evaluate(el=>getComputedStyle(el).pointerEvents),'none');
        if(process.env.NP_PLUGIN_ARTIFACTS) {fs.mkdirSync(process.env.NP_PLUGIN_ARTIFACTS,{recursive:true});await page.screenshot({path:path.join(process.env.NP_PLUGIN_ARTIFACTS,`browse-${width}-${theme}-${mode}.png`),fullPage:true});if(width===1440&&theme==='dark'&&mode==='list')fs.writeFileSync(path.join(process.env.NP_PLUGIN_ARTIFACTS,'browse-preview.html'),(await page.content()).replace(/<script[\s\S]*?<\/script>/g,''));}
      }
      // Dense capped pagination (including four-digit final page) must also fit mobile.
      await page.evaluate(async()=>{window.savedApi=api;window.api=async(url,o)=>{if(!url.includes('/search?'))return savedApi(url,o);calls.push({url,...o});const q=new URL(url,'http://demo').searchParams,size=Number(q.get('limit')),offset=Number(q.get('offset'));return{hits:Array.from({length:size},(_,i)=>({...demoHits[i%demoHits.length],project_id:'cap'+(offset+i),slug:'cap-'+(offset+i)})),total_hits:12345};};await NPPlugins.setPageSize(10);await NPPlugins.goPage(500);});
      assert.equal(await page.locator('.np-page-gap').count(),2);assert.equal(await page.locator('.np-page').count(),7);
      assert.ok(await page.locator('.np-project-pages').evaluate(el=>el.getBoundingClientRect().height<=44 && el.scrollWidth<=el.clientWidth+1),`${width}/${theme} capped pagination fits one row`);
      await page.getByRole('button',{name:'Page 1000',exact:true}).click();assert.match(await page.locator('.np-project-page-status').textContent(),/9,991–10,000 of 12,345/);assert.equal(await page.getByRole('button',{name:'Next page'}).isDisabled(),true);
      if(process.env.NP_PLUGIN_ARTIFACTS)await page.locator('.np-project-pagination').screenshot({path:path.join(process.env.NP_PLUGIN_ARTIFACTS,`pagination-cap-${width}-${theme}.png`)});
      await page.evaluate(async()=>{api=savedApi;await NPPlugins.setPageSize(20);});
      await page.evaluate(()=>{window.beforePageBack=calls.filter(x=>x.url.includes('/search?')).length;});
      await page.getByRole('button',{name:'Page 2',exact:true}).click();await page.locator('[data-np-project]').first().click();await page.locator('.np-project-hero').waitFor();await page.locator('[data-np-back]').click();
      assert.equal(await page.locator('[aria-current=page]').textContent(),'2');assert.equal(await page.getByRole('combobox',{name:'Projects per page'}).inputValue(),'20');assert.equal(await page.evaluate(()=>calls.filter(x=>x.url.includes('/search?')).length),await page.evaluate(()=>beforePageBack+1));
      await page.getByRole('button',{name:'Page 1',exact:true}).click();
      await page.emulateMedia({reducedMotion:'no-preference'});await page.evaluate(()=>{document.documentElement.dataset.panelMotion='reduced';R();});assert.equal(await page.locator('[data-project-card]').first().evaluate(el=>getComputedStyle(el).animationName),'none');await page.emulateMedia({reducedMotion:'reduce'});
      await page.evaluate(()=>{viewMode='list';R();});
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
      await page.waitForFunction(()=>document.querySelector('[data-version="v-old"]').textContent==='Installed');
      assert.equal(await page.locator('[data-version="v-old"]').isDisabled(),true);
      assert.equal(await page.locator('[data-version="v-new"]').isDisabled(),false);
      assert.doesNotMatch(await page.locator('.np-project').innerText(),/Install again|Installed · restart required/);
      await page.evaluate(()=>{demoData.versions_truncated=true;demoData.versions[0].dependencies=[{project_id:'dependency',filename:'<img onerror=window.pwned=1>',dependency_type:'required',version_id:'dep-version'}];});
      await page.evaluate(async()=>{await NPPlugins.open('project-0',true);});
      await page.getByRole('tab',{name:'Versions',exact:true}).click();
      await page.locator('.np-project-version details').first().locator('summary').click();
      assert.match(await page.locator('#np-project-panel').textContent(),/Showing up to 500 recent releases/);
      assert.match(await page.locator('.np-project-version').first().textContent(),/Dependencies are not installed automatically/);
      assert.equal(await page.locator('.np-project-version [onerror]').count(),0);
      assert.equal(await page.locator('.np-project-version a[href="https://modrinth.com/project/dependency"]').count(),1);
      assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),`${width}/${theme} detail overflow`);
      await page.getByRole('button',{name:'Back to results'}).click();assert.equal(await page.locator('[data-project-card]').count(),20);
      assert.equal(await page.evaluate(()=>document.getElementById('main').scrollTop===savedScroll && scrollY===savedWindowScroll),true,'Back preserves both scroll containers');
      await page.locator('#store-search').fill('changed');await page.waitForTimeout(400);await page.waitForFunction(()=>document.querySelectorAll('[data-project-card]').length===20);assert.ok(await page.evaluate(()=>calls.filter(x=>x.url.includes('/search?')).at(-1).url.includes('offset=0')));
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
      for(const mode of ['list','tiles']) {
        await page.evaluate(mode=>{S.view='fi';viewMode=mode;S.currentDir='';R();},mode);
        const first=mode==='list'?page.locator('.file-table tbody > tr').first():page.locator('.vm-tiles-grid > *').first();assert.match(await first.textContent(),/Trash/);assert.equal(await first.locator('input,[data-file-drag],[data-path]').count(),0);
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
test('accessible results cap is 10000 with true total and bounded compact navigation',async()=>{
  const {c,calls}=harness();c.handler=async url=>catalog(url,12345);
  await c.NPPlugins.search();await c.NPPlugins.goPage(500);assert.match(calls.at(-1).url,/limit=20&offset=9980/);assert.equal(c.S.lastModrinthHits.length,20);
  let html=c.NPPlugins.hitsHtml(c.S.lastModrinthHits);assert.match(html,/9,981–10,000 of 12,345/);assert.match(html,/Refine your search/);assert.match(html,/Page 500 of 500/);
  const before=calls.length;await c.NPPlugins.goPage(501);assert.equal(calls.length,before);
  await c.NPPlugins.goPage(250);html=c.NPPlugins.hitsHtml(c.S.lastModrinthHits);assert.equal((html.match(/class="np-page-gap"/g)||[]).length,2);assert.equal((html.match(/data-np-page=/g)||[]).length,7);
  for(const size of [10,50,100]) {await c.NPPlugins.setPageSize(size);await c.NPPlugins.goPage(10000/size);assert.match(calls.at(-1).url,new RegExp(`offset=${10000-size}`));}
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

test('pending page reads are single flight, abort on size/filter changes and never commit stale results',async()=>{
  const {c,calls}=harness();c.handler=async url=>catalog(url,250);await c.NPPlugins.search();
  let resolve;c.handler=()=>new Promise(r=>resolve=r);const pending=c.NPPlugins.goPage(2);
  await c.NPPlugins.goPage(2);await c.NPPlugins.search();assert.equal(calls.length,2);
  const loading=c.NPPlugins.hitsHtml(c.S.lastModrinthHits);assert.match(loading,/Loading page 2/);assert.match(loading,/1–20 of 250/);assert.equal((loading.match(/data-np-page=[^>]*disabled/g)||[]).length,(loading.match(/data-np-page=/g)||[]).length);
  c.handler=async url=>catalog(url,250);await c.NPPlugins.setPageSize(50);assert.equal(calls[1].o.signal.aborted,true);
  resolve({hits:[hit(999)],total_hits:250});await pending;assert.equal(c.S.lastModrinthHits.length,50);assert.equal(c.S.lastModrinthHits[0].project_id,'id0');
  c.handler=()=>new Promise(r=>resolve=r);const old=c.NPPlugins.goPage(3);c.S.q='new';c.NPPlugins.debounce();assert.equal(calls.at(-1).o.signal.aborted,true);
  resolve({hits:[hit(999)],total_hits:250});await old;assert.equal(c.S.lastModrinthHits.length,0);
});

test('pending project renders a safe clicked-result preview and skeleton, never loaded release facts',async()=>{
  const {c}=harness();let resolve;c.S.lastModrinthHits=[{slug:'project-0',title:'<img onerror=evil()>',description:'Demo preview <script>evil</script>',icon_url:'https://user:pass@example.com/icon.png'}];c.handler=()=>new Promise(r=>resolve=r);
  const pending=c.NPPlugins.open('project-0');const html=c.NPPlugins.render();assert.match(html,/np-project-loading/);assert.match(html,/aria-busy="true"/);assert.match(html,/role="status"/);assert.match(html,/aria-hidden="true"/);assert.match(html,/&lt;img/);assert.match(html,/&lt;script&gt;/);assert.doesNotMatch(html,/<img|<script|user:pass|data-np-install|downloads|followers/);assert.equal(c.NPPlugins.isOpen(),true);
  resolve(project());await pending;assert.doesNotMatch(c.NPPlugins.render(),/np-project-loading/);assert.match(c.NPPlugins.render(),/data-np-install/);
});

test('Browse defaults to Tiles, preserves supported layouts, and migrates Icons',()=>{
  for(const saved of [null,'list','icons','tiles','invalid','blocked']) {
    const {c}=harness();
    c.localStorage={getItem:key=>{if(saved==='blocked')throw Error('storage unavailable');return key==='neopanel_vm_plug_store'?saved:null;}};
    vm.runInNewContext(section('function getViewMode(', 'let viewModeTimer'),c);
    const expected=['list','tiles'].includes(saved)?saved:'tiles';
    const html=c.NPPlugins.hitsHtml([hit(0)]);
    assert.match(html,new RegExp(`np-project-result ${{list:'plugin-card',tiles:'vm-tile-card'}[expected]}`));
    assert.equal(c.getViewMode('fi','list'),'list','Files retains its own fallback');
  }
  for(const scope of ['plug_store','plug_inst']) {
    assert.match(app,new RegExp(`getViewMode\\('${scope}', 'tiles'\\)`));
    assert.match(app,new RegExp(`renderViewModeToggle\\('${scope}', 'tiles'\\)`));
  }
});

test('button-owned install progress is project-scoped, indeterminate, and clears independently',async()=>{
  const {c,calls}=harness();c.S.lastModrinthHits=[hit(0),hit(1)];const pending=new Map();
  c.handler=async(url,o)=>{if(o.method==='POST')return new Promise((resolve,reject)=>pending.set(JSON.parse(o.body).slug,{resolve,reject}));return project();};
  const first=c.NPPlugins.install('project-0','Demo 0');
  assert.match(c.NPPlugins.hitsHtml(c.S.lastModrinthHits),/Checking…/);
  for(let i=0;i<8;i++)await Promise.resolve();
  const second=c.NPPlugins.install('project-1','Demo 1');for(let i=0;i<8;i++)await Promise.resolve();
  let html=c.NPPlugins.hitsHtml(c.S.lastModrinthHits);
  assert.equal((html.match(/class="np-install-ring"/g)||[]).length,2);
  assert.doesNotMatch(html,/aria-valuenow|\d+%/);
  assert.ok(calls.filter(x=>x.o.method==='POST').every(x=>x.o.progressTarget==='button'));
  await c.NPPlugins.install('project-0','Demo 0');assert.equal(pending.size,2);
  pending.get('project-0').resolve({});await first;
  html=c.NPPlugins.hitsHtml(c.S.lastModrinthHits);assert.equal((html.match(/class="np-install-ring"/g)||[]).length,1);assert.match(html,/is-installed[^>]*disabled[^>]*>[\s\S]*?<span class="np-install-caption">Installed<\/span>/);assert.doesNotMatch(html,/Install again|Installed · restart required/);
  pending.get('project-1').reject(Error('timeout'));await second;
  assert.doesNotMatch(c.NPPlugins.hitsHtml(c.S.lastModrinthHits),/np-install-ring|aria-busy="true"/);
  assert.ok(c.messages.some(args=>args[0].includes('Check installed files')));
});

test('returning to a page while installation is pending retains its ring then clears it',async()=>{
  const {c}=harness();c.S.lastModrinthHits=[hit(0)];let finish;let draws=0;
  c.document.getElementById=id=>id==='store-grid'?{set innerHTML(_v){draws++;},removeAttribute(){}}:null;
  c.handler=async(_url,o)=>o.method==='POST'?new Promise(resolve=>finish=resolve):project();
  c.NPPlugins.sync();const job=c.NPPlugins.install('project-0','Demo');for(let i=0;i<8;i++)await Promise.resolve();
  c.S.view='con';c.NPPlugins.sync();c.S.view='plug';c.NPPlugins.sync();
  assert.match(c.NPPlugins.hitsHtml(c.S.lastModrinthHits),/np-install-ring/);
  const before=draws;finish({});await job;
  assert.ok(draws>before);assert.doesNotMatch(c.NPPlugins.hitsHtml(c.S.lastModrinthHits),/np-install-ring/);
  assert.equal(c.messages.length,0,'stale completion does not claim success');
});

test('completed project and matching-version buttons show Installed without a restart status',async()=>{
  const {c,calls}=harness();c.S.lastModrinthHits=[hit(0)];await c.NPPlugins.open('project-0');
  await c.NPPlugins.install('project-0','Demo 0',null,'v-old');
  const hero=c.NPPlugins.render();
  assert.match(hero,/is-installed[^>]*data-np-install[^>]*disabled[^>]*aria-busy="false"/);
  assert.doesNotMatch(hero,/Install again|Installed · restart required/);
  const actions=()=>[...c.NPPlugins.hitsHtml(c.S.lastModrinthHits).matchAll(/<button[^>]*data-np-install[^>]*>[\s\S]*?<\/button>/g)];
  assert.equal(actions().length,1);assert.match(actions()[0][0],/>Installed<\/span>/);
  const before=calls.length;
  await c.NPPlugins.install('id0','Demo 0');
  await c.NPPlugins.install('project-0','Demo 0',null,'v-old');
  assert.equal(calls.length,before,'completed default and identical version do not reinstall');
  await c.NPPlugins.install('project-0','Demo 0',null,'v-new');
  assert.equal(calls.filter(call=>call.o.method==='POST').length,2,'a different release remains installable');
});
