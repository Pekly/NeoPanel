const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const root = path.join(__dirname, '../static');
const app = fs.readFileSync(path.join(root, 'js/app.js'), 'utf8');
const source = fs.readFileSync(path.join(root, 'js/server-artifact.js'), 'utf8');
const section = (a,b) => app.slice(app.indexOf(a), app.indexOf(b,app.indexOf(a)));
const helpers = section('const ENGINE_ICONS =', 'function onPlayerHeadError(') + section('const PREMADE_JAVA_VERSIONS =', 'function renderJavaAvailabilityStrip(');
const escapeHTML = v => String(v ?? '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
function setup() {
  const calls=[];
  const srv={id:'one',node_id:'worker',type:'paper',version:'configured-wrong',jar_file:'custom.jar',java_path:'auto'};
  const c={S:{user:{id:1,role:'user'},token:'token',view:'set',servers:[srv],javaRuntimesMap:{}},allowed:true,
    curServer:()=>c.S.servers[0],hasServerPermission:()=>c.allowed,serverPagePermission:v=>['set','con','fi'].includes(v),canServerPage:()=>true,
    srvApiPath:(s,suffix)=>`/api/nodes/${s.node_id}/proxy/api/servers/${s.id}${suffix}`,esc:escapeHTML,
    AbortController,setTimeout,clearTimeout,document:{getElementById:()=>null},updateSettingsCmdPreview(){},
    api:(url,opts)=>new Promise((resolve,reject)=>calls.push({url,opts,resolve,reject}))};
  c.window=c;vm.runInNewContext(helpers+source,c);
  return {c,srv,calls,a:c.NPServerArtifact};
}
const tick = () => new Promise(resolve=>setImmediate(resolve));
test('unknown identity never borrows configured software/version; native remains configured',()=>{
  const {a,srv}=setup();assert.equal(a.display(srv).software,'custom');assert.equal(a.display(srv).text,'Custom / Unknown version');
  for(const status of ['unknown','missing','error']) {srv.artifact={status,software:'paper',minecraft_version:'bad'};assert.equal(a.display(srv).text,'Custom / Unknown version');}
  srv.type='steel';assert.match(a.display(srv).text,/Steel.*configured-wrong.*Native/);
});
test('proxy release is not Minecraft; loader, build and unsafe provider labels are escaped',()=>{
  const {a,srv}=setup();srv.artifact={status:'identified',software:'velocity',project_version:'3.4.0',required_java:21};
  assert.equal(a.display(srv).software,'velocity');assert.match(a.display(srv).text,/Release 3.4.0/);assert.doesNotMatch(a.display(srv).text,/Minecraft/);
  srv.artifact={status:'partial',software:'fabric',project_version:'<img onerror=x>',minecraft_version:'1.20.4',build_name:'<script>'};
  assert.match(a.label(srv,13,true),/Minecraft 1.20.4/);assert.match(a.label(srv,13,true),/&lt;img/);assert.doesNotMatch(a.label(srv,13,true),/<script>/);
  assert.equal(srv.type,'paper');assert.equal(srv.version,'configured-wrong');
});
test('Auto exact/minimum/unknown labels and system/manual selections remain distinct',()=>{
  const {a,c,srv}=setup();
  for(const raw of ['', 'java','default']) assert.equal(c.resolveJavaSelectValue(raw,'latest','paper'),'java');
  for(const raw of ['auto','8','11','16','17','21','25']) assert.equal(c.resolveJavaSelectValue(raw),raw);
  assert.equal(c.resolveJavaSelectValue('/custom/java'),'custom');
  const options=c.renderJavaSelectOptions('auto');assert.match(options,/^<option value="auto" selected/);
  srv.artifact={status:'partial',minimum_java:17};assert.match(a.panel(srv),/not an exact runtime recommendation/);assert.doesNotMatch(a.javaLabel(srv),/Java 17/);
  srv.artifact={status:'identified',required_java:21,java_source:'mcjars-build'};assert.equal(a.javaLabel(srv),'Auto (Java 21)');assert.match(a.panel(srv),/MCJars build metadata/);
});
test('unknown icons and broken images use one-shot bundled Minecraft fallback',()=>{
  const {c}=setup();for(const value of ['',undefined,'custom','nonsense','__proto__']) assert.equal(c.getEngineIconUrl(value),'/img/default-server-icon.png');
  assert.match(c.engineIcon('paper'),/this.onerror=null;this.src='\/img\/default-server-icon.png'/);
});
test('singleflight, worker path, bounded error attempt and explicit retry',async()=>{
  const {a,calls,srv}=setup();a.sync();a.sync();assert.equal(calls.length,1);assert.match(calls[0].url,/nodes\/worker\/proxy\/api\/servers\/one\/artifact$/);
  assert.equal(calls[0].opts.optionalMetadata,true);calls[0].reject(Error('404'));await tick();a.sync();assert.equal(calls.length,1);assert.match(a.panel(srv),/Older workers/);
  a.retry();assert.equal(calls.length,2);assert.match(calls[1].url,/\/artifact\?refresh=true$/);calls[1].resolve({status:'identified',software:'fabric',required_java:17});await tick();a.sync();assert.equal(calls.length,2);assert.equal(a.display(srv).software,'fabric');
});
test('permission and inventory views never auto-fetch; supplied summaries remain usable',()=>{
  const {a,c,calls,srv}=setup();c.allowed=false;a.sync();assert.equal(calls.length,0);srv.artifact={status:'identified',software:'paper',minecraft_version:'1.21'};
  assert.match(a.label(srv),/>Paper<\/span>/);assert.equal(a.display(srv).version,'1.21');assert.doesNotMatch(a.panel(srv),/onclick=|custom.jar/);
  c.allowed=true;c.S.view='dash';delete srv.artifact;a.sync();assert.equal(calls.length,0);
});
test('stale server/node/account/token/jar/type/version contexts abort and ignore completions',async()=>{
  for(const mutate of [c=>c.S.servers[0]={...c.S.servers[0],id:'other'},c=>c.S.servers[0].node_id='other',c=>c.S.user.id=2,c=>c.S.token='new',c=>c.S.servers[0].jar_file='new.jar',c=>c.S.servers[0].type='fabric',c=>c.S.servers[0].version='new']) {
    const {a,c,calls}=setup();a.sync();mutate(c);a.sync();assert.equal(calls[0].opts.signal.aborted,true);assert.equal(calls[0].opts.isCurrent(),false);
    calls[0].resolve({status:'identified',software:'velocity',required_java:25});await tick();assert.doesNotMatch(a.label(c.curServer()),/Velocity/);
    calls[1].resolve({status:'unknown'});await tick();assert.equal(calls.length,2);a.dispose();
  }
});
test('leaving and revisiting during a request cannot let an old completion remove the new request',async()=>{
  const {a,c,calls}=setup();a.sync();c.S.view='dash';a.sync();c.S.view='set';a.sync();calls[0].resolve({status:'unknown'});await tick();a.sync();assert.equal(calls.length,2);
  calls[1].resolve({status:'identified',software:'fabric'});await tick();assert.equal(a.display(c.curServer()).software,'fabric');
});
test('file signatures, successful mutations and metadata refresh invalidate safely',async()=>{
  const {a,c,calls,srv}=setup();srv.artifact={status:'identified',software:'paper',size:4,modified:'a'};a.sync();assert.equal(calls.length,0);
  a.files(srv,'',[{name:'custom.jar',size:4,modified:'a'}]);a.sync();assert.equal(calls.length,0);
  a.files(srv,'',[{name:'custom.jar',size:8,modified:'b'}]);a.sync();assert.equal(calls.length,1);
  calls[0].resolve({status:'identified',software:'fabric'});await tick();assert.equal(a.display(srv).software,'fabric');
  srv.artifact={status:'identified',software:'velocity',project_version:'3.4'};assert.equal(a.display(srv).software,'velocity','fresh response wins over cached hash lookup');
  a.requestResult(c.srvApiPath(srv,'/files/rename'),{method:'POST'},{},'old-token');assert.ok(srv.artifact);
  a.requestResult(c.srvApiPath(srv,'/files/rename'),{method:'POST'},{},c.S.token);assert.equal(srv.artifact,undefined);a.sync();assert.equal(calls.length,2);calls[1].resolve({status:'unknown'});await tick();
});
test('install/reinstall do not send heuristic Java PUTs; API optional metadata skips workspace health',()=>{
  assert.doesNotMatch(section('async function installSoftware(', 'function syncPropsFromServer('),/java_path:|choice.recJava/);
  assert.doesNotMatch(section('async function submitReinstall(', 'async function openAdminServerSettingsModal('),/java_path:|recJava/);
  assert.match(section('async function submitReinstall(', 'async function openAdminServerSettingsModal('),/if \(typeof res\?\.java_path === 'string'\) targetSrv.java_path = res.java_path/);
  assert.match(section('async function api(', 'function curServer('),/if \(!optionalMetadata\) window.NPWorkspace\?\.requestResult\(endpoint, err/);
  assert.match(app,/if \(path === '\/artifact'\) return 'files.view'/);
});
test('settings preserve blank/system aliases and intentionally save Auto or manual/custom choices',()=>{
  const {c,srv}=setup();c.S.props={};c.getServerMaxRamMB=()=>2048;
  const choice={value:'java'},custom={value:'/custom/java'};c.$=id=>id==='#prop-java'?choice:id==='#prop-java-custom'?custom:null;
  vm.runInNewContext(section('function collectSettingsFormPayload(', 'function seedSettingsSavedSnapshot('),c);
  for(const raw of ['', 'java','default']) {srv.java_path=raw;assert.equal(c.collectSettingsFormPayload(srv).data.javaPath,raw);}
  for(const raw of ['auto','8','11','16','17','21','25']) {choice.value=raw;assert.equal(c.collectSettingsFormPayload(srv).data.javaPath,raw);}
  choice.value='custom';assert.equal(c.collectSettingsFormPayload(srv).data.javaPath,'/custom/java');custom.value='';assert.equal(c.collectSettingsFormPayload(srv).ok,false);
});
test('successful settings responses, not submitted settings, determine installed identity and exact Java',async()=>{
  const {c,srv,a}=setup();c.S.props={};let error=false;
  Object.assign(c,{_settingsSaveInFlight:false,_settingsSaveQueued:null,_settingsLastSaved:{},getSettingsSnapshotKey:()=> 'one',setSettingsSaveStatus:state=>{if(state==='error')error=true;},clock:()=>'',toast(){},syncHeroFromSavedServer(){},
    buildSettingsSnapshots:()=>({propsJson:'{}',srvJson:'{}',srvPayload:{java_path:'25',java_args:'',memory_min:0}}),
    api:async()=>({...srv,jar_file:'verified.jar',version:'server-version',java_path:'auto',artifact:{status:'identified',software:'fabric',required_java:17}})});
  vm.runInNewContext(section('async function persistSettingsPayload(', 'async function saveProperties('),c);
  await c.persistSettingsPayload(srv,{ok:true,data:{jar:'submitted.jar',ver:'submitted-version',javaPath:'25',ram:2048}},{autoSave:true,skipRender:true});
  assert.equal(error,false);assert.equal(srv.jar_file,'verified.jar');assert.equal(srv.version,'server-version');assert.equal(srv.java_path,'auto');assert.equal(a.javaLabel(srv),'Auto (Java 17)');assert.equal(srv.type,'paper');
});
test('unrelated successful settings saves retain a single metadata lookup',async()=>{
  const {c,srv,a,calls}=setup();a.sync();calls[0].resolve({status:'identified',software:'fabric'});await tick();
  a.requestResult(c.srvApiPath(srv,''),{method:'PUT'},{...srv,name:'Renamed'},c.S.token);a.sync();assert.equal(calls.length,1);
});
test('software confirmation and completion stay bound to the selected server',async()=>{
  const {c,srv}=setup();srv.java_path='';let confirmation,download,powerCalls=0,posts=0;
  Object.assign(c,{askInstallSoftwareModal:()=>new Promise(resolve=>confirmation=resolve),toast(){},power(){powerCalls++;},setBtnLoading(){},renderVersionRows(){},R(){},uiAlert(){},document:{getElementById:()=>null,body:{contains:()=>false}},api:()=>{posts++;return new Promise(resolve=>download=resolve);}});
  vm.runInNewContext(section('async function installSoftware(', 'function syncPropsFromServer('),c);
  let job=c.installSoftware('fabric','latest');c.S.servers[0]={...srv,id:'other'};confirmation({confirmed:true});await job;assert.equal(posts,0);
  c.S.servers[0]=srv;job=c.installSoftware('fabric','latest');confirmation({confirmed:true});await tick();assert.equal(posts,1);
  c.S.servers[0]={...srv,id:'other'};download({software:'fabric',version:'1.21',jar_file:'server.jar',java_path:'auto'});await job;assert.equal(powerCalls,0);assert.equal(c.S.servers[0].type,'paper');assert.equal(srv.type,'fabric');assert.equal(srv.java_path,'auto');c.NPServerArtifact.dispose();
});

test('Chrome: responsive identity, keyboard retry, reduced motion, error/success and broken-image fallback', {skip:!process.env.NP_LAYOUT_PLAYWRIGHT}, async()=>{
  const {chromium}=require(process.env.NP_LAYOUT_PLAYWRIGHT);
  const browser=await chromium.launch({headless:true,...(process.env.NP_LAYOUT_CHROME?{executablePath:process.env.NP_LAYOUT_CHROME}:{})});
  const css=[...fs.readFileSync(path.join(root,'index.html'),'utf8').matchAll(/href="\/css\/([^"]+)"/g)].map(m=>fs.readFileSync(path.join(root,'css',m[1]),'utf8')).join('\n');
  try {
    for(const width of [1440,768,390,320]) for(const theme of ['light','dark']) {
      const page=await browser.newPage({viewport:{width,height:900}});await page.route('**/*',r=>r.abort());await page.emulateMedia({reducedMotion:'reduce'});
      await page.setContent(`<html data-theme="${theme}"><head><style>${css}</style></head><body style="display:block;padding:16px"><main style="max-width:640px;margin:auto"><p>Demo data · Installed build identity</p><span id="srv-hero-engine"></span><div id="fixture"></div></main></body></html>`);
      await page.addScriptTag({content:`window.S={token:'fixture',user:{id:1},view:'set',servers:[{id:'one',node_id:'worker',type:'paper',version:'1.21',jar_file:'custom.jar',java_path:'auto',artifact:{status:'unknown'}}]};window.curServer=()=>S.servers[0];window.hasServerPermission=()=>true;window.serverPagePermission=()=>true;window.canServerPage=()=>true;window.esc=${escapeHTML.toString()};window.srvApiPath=(s,p)=>'/api/servers/'+s.id+p;window.api=()=>new Promise((resolve,reject)=>{window.finish=resolve;window.fail=reject;});window.updateSettingsCmdPreview=()=>{};${helpers}\n${source}\nNPServerArtifact.sync();document.getElementById('fixture').innerHTML=NPServerArtifact.panel(curServer());document.getElementById('srv-hero-engine').innerHTML=NPServerArtifact.label(curServer());`});
      await page.keyboard.press('Tab');assert.equal(await page.locator('button').evaluate(e=>e===document.activeElement),true);
      assert.equal(await page.locator('button').evaluate(e=>getComputedStyle(e).outlineStyle),'solid');
      await page.keyboard.press('Enter');assert.equal(await page.locator('button').isDisabled(),true);
      await page.evaluate(()=>fail(Error('404')));await page.waitForFunction(()=>!document.querySelector('button').disabled);assert.match(await page.locator('.artifact-message').textContent(),/Older workers/);
      await page.locator('button').click();await page.evaluate(()=>finish({status:'identified',software:'velocity',project_version:'3.4.0-'+ 'long'.repeat(35),required_java:21,java_source:'mcjars-build'}));
      await page.waitForFunction(()=>document.querySelector('.artifact-message').textContent.includes('Java 21'));
      assert.doesNotMatch(await page.locator('#srv-hero-engine').textContent(),/Minecraft 3.4/);
      assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true,`${theme}/${width}`);
      const img=page.locator('#srv-hero-engine img');await img.evaluate(e=>e.dispatchEvent(new Event('error')));
      assert.equal(await img.getAttribute('src'),'/img/default-server-icon.png');assert.equal(await img.evaluate(e=>e.onerror),null);
      assert.equal(await page.locator('button').evaluate(e=>getComputedStyle(e).transitionDuration),'0s');
      if(process.env.PI_SCRATCH_DIR && width===390 && theme==='dark') await page.screenshot({path:path.join(process.env.PI_SCRATCH_DIR,'artifact-mobile.png')});
      await page.close();
    }
  } finally {await browser.close();}
});

test('supplied and fetched identities expire without a polling timer',async()=>{
  const {a,c,srv,calls}=setup();let now=1000;c.Date={now:()=>now};
  srv.artifact={status:'identified',software:'paper',minecraft_version:'1.21'};
  assert.equal(a.display(srv).software,'paper');now+=300001;
  assert.equal(a.display(srv).software,'custom');assert.equal(calls.length,0);
  a.sync();assert.equal(calls.length,1);calls[0].resolve({status:'identified',software:'fabric'});await tick();
  assert.equal(a.display(srv).software,'fabric');now+=300001;
  assert.equal(a.display(srv).software,'custom');assert.equal(calls.length,1);
  a.sync();assert.equal(calls.length,2);a.dispose();
});

test('software badges stay short and version text excludes build metadata',()=>{
  const {a,c,srv}=setup();
  srv.artifact={status:'identified',software:'velocity',project_version:'4.3.2-SNAPSHOT',build_name:'#58',build_id:'long-build-uuid'};
  assert.equal(a.display(srv).name,'Velocity');
  assert.equal(c.serverArtifactVersion(srv),'4.3.2-SNAPSHOT');
  assert.match(a.label(srv),/>Velocity<\/span>/);
  assert.doesNotMatch(a.label(srv),/Release|Build|SNAPSHOT|long-build/);
  assert.match(a.panel(srv),/Build #58/);
  srv.artifact={status:'identified',software:'fabric',minecraft_version:'1.21.4',project_version:'0.20.1',build_name:'#99'};
  assert.equal(c.serverArtifactVersion(srv),'1.21.4');
  for(const status of ['unknown','missing','error']) {
    srv.artifact={status,software:'velocity',project_version:'stale',minecraft_version:'stale'};
    assert.equal(a.display(srv).name,'Custom');assert.equal(c.serverArtifactVersion(srv),'Unknown version');
  }
});
