const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const root = path.join(__dirname, '../static');
const app = fs.readFileSync(path.join(root, 'js/app.js'), 'utf8');
function section(start, end) {
  const a=app.indexOf(start), b=app.indexOf(end,a);
  assert.ok(a>=0 && b>a,start);
  return app.slice(a,b);
}
const creation = section('function setCreateRam(', 'function closeModal(');
const helpers = section('const ENGINE_ICONS =', 'function onPlayerHeadError(') + section('const CORES =', 'const CATS =') + section('const PREMADE_JAVA_VERSIONS =', 'async function loadJavaRuntimes(');
const dropdowns = section('function closeAllCustomSelects()', "document.addEventListener('click',");

test('fresh creation advertises latest rather than a dated preset and selects Auto', () => {
  const container={innerHTML:''}, javaCalls=[];
  const c={S:{nodes:[],servers:[],user:{role:'admin'}},$:()=>container,canCreateServers:()=>true,getNextSuggestedPort:()=>25565,
    getAllMcVersions:()=>['1.21.11','26.3'],CORES:[['Paper','Demo']],engineIcon:()=>'',esc:String,
    renderJavaSelectOptions:(...args)=>{javaCalls.push(args);return `<option value="${args[0]}" selected>Java ${args[0]}</option>`;},renderJavaAvailabilityStrip:()=>'',
    loadVersionsCatalog(){},loadJavaRuntimes(){},enhanceSelects(){},setTimeout(){}};
  vm.runInNewContext(section('function openCreateServerModal()', 'function closeModal('),c);
  c.openCreateServerModal();
  assert.match(container.innerHTML, /id="cs-ver" value="latest"/);
  assert.match(container.innerHTML, /value="latest" selected>Latest available/);
  assert.doesNotMatch(container.innerHTML, /ver-pill|Recent &amp; Popular|value="1.21.11" selected/);
  assert.deepEqual(javaCalls[0].slice(0,4),['auto',false,'latest','paper']);
  for(const id of ['cs-eula','cs-dl','cs-aikar']) assert.match(container.innerHTML,new RegExp(`<label[^>]*np-check-row[^>]*>\\s*<input[^>]*id="${id}"`));
});

test('creation submission defaults to latest/auto and preserves explicit choices and native settings', async () => {
  for(const variant of [{type:'paper'},{type:'velocity'},{type:'steel'},{type:'paper',version:'1.12.2',java:'8'}]) {
    const fields={'#cs-type':{value:variant.type}};
    if(variant.version) fields['#cs-ver']={value:variant.version};
    if(variant.java) fields['#cs-java']={value:variant.java};
    let payload;
    const c={S:{servers:[]},$:id=>fields[id],canCreateServers:()=>true,isSteelEngine:type=>type==='steel',toast(){},
      api:async(_,options)=>{payload=JSON.parse(options.body);return {};},nodeApiPath:(_,p)=>p,closeModal(){},loadDashboard:async()=>{},applyRoute(){},uiAlert(){}};
    vm.runInNewContext(section('async function handleCreateServer(', '// --- Port Allocation Management ---'),c);
    await c.handleCreateServer({preventDefault(){}});
    assert.equal(payload.version,variant.version||'latest');
    assert.equal(payload.java_path,variant.type==='steel'?'':variant.java||'auto');
    if(variant.type==='steel') assert.equal(payload.java_args,'');
  }
});

test('automatic Java selection remains Auto and never overrides a manual runtime', () => {
  const select={value:'21',dataset:{},_cselectRebuild(){this.rebuilt=true;}};
  const fields={'#cs-type':{value:'velocity'},'#cs-java':select};
  const c={$:id=>fields[id],$$:()=>[],isSteelEngine:type=>type==='steel'};
  vm.runInNewContext(section('function syncCreateJavaFromVersion(', 'function syncCreateSelectValue('),c);
  c.syncCreateJavaFromVersion('3.4.0-SNAPSHOT');assert.equal(select.value,'auto');assert.equal(select.dataset.userPicked,undefined);
  select.value='8';select.dataset.userPicked='1';c.syncCreateJavaFromVersion('latest');assert.equal(select.value,'8');
  fields['#cs-type'].value='steel';delete select.dataset.userPicked;c.syncCreateJavaFromVersion('latest');assert.equal(select.value,'8');
});

test('browser: creation defaults survive engine/catalog changes and fields stay aligned', {skip:!process.env.NP_LAYOUT_PLAYWRIGHT}, async () => {
  const {chromium}=require(process.env.NP_LAYOUT_PLAYWRIGHT);
  const browser=await chromium.launch({headless:true,...(process.env.NP_LAYOUT_CHROME?{executablePath:process.env.NP_LAYOUT_CHROME}:{})});
  const index=fs.readFileSync(path.join(root,'index.html'),'utf8');
  const css=[...index.matchAll(/href="\/css\/([^"]+)"/g)].map(m=>fs.readFileSync(path.join(root,'css',m[1]),'utf8')).join('\n');
  try {
    for(const theme of ['dark','light']) for(const width of [1440,1024,768,390,320]) {
      const page=await browser.newPage({viewport:{width,height:1000}});
      await page.route('**/*',route=>route.abort());
      await page.emulateMedia({reducedMotion:'reduce'});
      await page.setContent(`<html data-theme="${theme}"><head><style>${css}</style></head><body><div id="modal-container"></div></body></html>`);
      await page.addScriptTag({content:`
        window.S={token:'fixture',user:{id:1,role:'admin'},adminUsers:[{id:1,username:'Owner',role:'admin'}],nodes:[{id:'local',name:'Local',online:true}],servers:[{id:'old',node_id:'local',version:'1.20.4',java_path:'17',port:25565}],ver:'1.20.4',javaRuntimesMap:{local:{premade:[{version:'25',installed:false},{version:'21',installed:true}],can_auto_install:false}}};
        window.$=s=>document.querySelector(s);window.$$=s=>Array.from(document.querySelectorAll(s));window.curServer=()=>S.servers[0];
        window.esc=v=>String(v??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
        window.ic=()=>'';window.canCreateServers=()=>true;window.displayNodeName=(id,name)=>name||id;window.displayNodeLocation=()=> 'Local';
        window.loadJavaRuntimes=()=>Promise.resolve();window.toast=()=>{};
        window.api=async url=>url==='/api/steel/versions'?['v-next']:new Promise(resolve=>window.finishCatalog=resolve);
        ${helpers}\n${creation}\n${dropdowns}\nopenCreateServerModal();`});
      assert.equal(await page.locator('#cs-ver').inputValue(),'latest');
      assert.equal(await page.locator('#cs-ver-sel').inputValue(),'latest');
      assert.equal(await page.locator('#cs-java').inputValue(),'auto','Auto must not silently select an installed manual runtime');
      const shape=await page.evaluate(()=>{
        const box=id=>{const r=document.getElementById(id).getBoundingClientRect();return {x:r.x,y:r.y,height:r.height,width:r.width};};
        return {resources:['cs-mem','cs-cpu','cs-disk'].map(box),version:box('cs-ver'),port:box('cs-port'),java:box('cs-java').height?box('cs-java'):(()=>{const r=document.querySelector('[data-for="cs-java"] .cselect-trigger').getBoundingClientRect();return {x:r.x,y:r.y,height:r.height};})(),overflow:document.querySelector('.pbox').scrollWidth>document.querySelector('.pbox').clientWidth+1,
          checks:['cs-eula','cs-dl','cs-aikar'].map(id=>{const a=document.getElementById(id).getBoundingClientRect(),b=document.getElementById(id).nextElementSibling.getBoundingClientRect();return {width:a.width,beside:b.left>a.right,aligned:b.top<a.bottom};})};
      });
      assert.equal(shape.overflow,false,`${theme}/${width}`);
      for(const row of shape.checks){assert.ok(row.width>=14);assert.ok(row.beside&&row.aligned,JSON.stringify(row));}
      if(width>700){
        assert.ok(Math.abs(shape.version.y-shape.port.y)<1,JSON.stringify(shape));
        assert.ok(Math.abs(shape.port.y-shape.java.y)<1,JSON.stringify(shape));
        assert.ok(shape.resources.every(r=>Math.abs(r.y-shape.resources[0].y)<1),JSON.stringify(shape));
      }
      for(const engine of ['velocity','waterfall','fabric','steel','paper']) {
        await page.evaluate(engine=>setCreateEngine(engine),engine);
        assert.equal(await page.locator('#cs-ver').inputValue(),'latest');
        assert.equal(await page.locator('#cs-ver-sel').inputValue(),'latest');
        assert.equal(await page.locator('#cs-java').inputValue(),'auto');
        assert.equal(await page.locator('#cs-java').isDisabled(),engine==='steel');
      }
      await page.selectOption('#cs-java','17',{force:true});
      await page.evaluate(()=>setCreateVersion('1.20.4'));
      await page.evaluate(()=>finishCatalog({all_mc:['27.1','26.3','1.20.4'],by_engine:{paper:['27.1']}}));
      await page.waitForFunction(()=>S.versionsCatalog?.all_mc?.[0]==='27.1');
      assert.equal(await page.locator('#cs-ver').inputValue(),'1.20.4');
      assert.equal(await page.locator('#cs-java').inputValue(),'17');
      await page.evaluate(()=>setCreateEngine('paper'));
      assert.equal(await page.locator('#cs-ver').inputValue(),'1.20.4','clicking the same engine keeps an explicit version');
      assert.equal(await page.evaluate(()=>S.servers[0].version),'1.20.4');
      assert.equal(await page.evaluate(()=>S.servers[0].java_path),'17');
      await page.evaluate(()=>openCreateServerModal());
      assert.equal(await page.locator('#cs-ver').inputValue(),'latest');
      assert.equal(await page.locator('#cs-java').inputValue(),'auto');
      await page.close();
    }
  } finally {await browser.close();}
});
