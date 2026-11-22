const test=require('node:test');
const assert=require('node:assert/strict');
const fs=require('node:fs');
const path=require('node:path');
const root=path.join(__dirname,'../static');
const app=fs.readFileSync(path.join(root,'js/app.js'),'utf8');
const api=app.slice(app.indexOf('async function api('),app.indexOf('function curServer(',app.indexOf('async function api(')));
const index=fs.readFileSync(path.join(root,'index.html'),'utf8');
const css=[...index.matchAll(/href="\/css\/([^"]+)"/g)].map(m=>fs.readFileSync(path.join(root,'css',m[1]),'utf8')).join('\n');

test('Chrome: inline install rings replace trays, remain independent, and respect reduced motion', {skip:!process.env.NP_LAYOUT_PLAYWRIGHT,timeout:60000},async()=>{
  const {chromium}=require(process.env.NP_LAYOUT_PLAYWRIGHT);
  const browser=await chromium.launch({headless:true,executablePath:process.env.NP_LAYOUT_CHROME});
  try{
    for(const theme of ['dark','light']){
      const page=await browser.newPage({viewport:{width:390,height:900}}),pending=new Map(),posts=[],errors=[];
      page.on('pageerror',e=>errors.push(e.message));
      await page.route('**/*',route=>{
        const req=route.request(),url=new URL(req.url());
        if(url.pathname==='/')return route.fulfill({contentType:'text/html',body:`<!doctype html><html data-theme="${theme}"><head><style>${css}#main{position:static;margin:0;padding:20px;}#store-grid{display:grid;gap:16px;grid-template-columns:minmax(0,1fr);}</style></head><body><main id="main"><div id="store-grid"></div></main></body></html>`});
        if(url.pathname.endsWith('/plugins/install')){const slug=req.postDataJSON().slug;posts.push(slug);pending.set(slug,route);return;}
        if(url.pathname.endsWith('/plugins/versions'))return route.fulfill({contentType:'application/json',body:JSON.stringify({versions:[{id:'v1',version_number:'1.0'}],loader:'paper',minecraft_version:'1.21.4'})});
        return route.fulfill({status:404,body:'{}'});
      });
      await page.goto('https://neopanel.test/');
      await page.evaluate(()=>{
        window.S={view:'plug',plugTab:'browse',plugins:[],lastModrinthHits:[{project_id:'id0',slug:'demo-0',title:'Simple Voice Chat',description:'Demo project — voice communication.',downloads:100},{project_id:'id1',slug:'demo-1',title:'Permissions',description:'Demo project — permission management.',downloads:80}],user:{id:1,role:'admin'},token:'fixture',skipPlugConfirm:true};
        window.srv={id:'one',node_id:'east',type:'paper',name:'Demo server'};window.curServer=()=>srv;window.hasServerPermission=()=>true;window.isSteelEngine=()=>false;window.isModdedEngine=()=>false;window.viewMode='tiles';window.getViewMode=()=>viewMode;
        window.srvApiPath=(s,p)=>`/api/nodes/${s.node_id}/proxy/api/servers/${s.id}${p}`;
        window.assertClientRequestAccess=()=>{};window.messages=[];window.toast=(...x)=>messages.push(x);window.logout=()=>{};window.loadPluginsList=async()=>{};window.uiConfirm=async()=>true;window.uiAlert=async()=>{};
        window.R=()=>{NPPlugins.sync();document.getElementById('store-grid').innerHTML=NPPlugins.hitsHtml(S.lastModrinthHits);};
      });
      for(const file of ['operation-progress.js','plugin-details.js'])await page.addScriptTag({content:fs.readFileSync(path.join(root,'js',file),'utf8')});
      await page.addScriptTag({content:api});await page.evaluate(()=>R());
      await page.locator('[data-np-install][data-slug="demo-0"]').click();
      await page.waitForFunction(()=>document.querySelector('[data-np-install][data-slug="demo-0"]').textContent.includes('Installing'));
      await page.locator('[data-np-install][data-slug="demo-1"]').click();
      await page.waitForFunction(()=>document.querySelectorAll('[data-np-install][aria-busy="true"]').length===2&&[...document.querySelectorAll('.np-install-caption')].every(e=>e.textContent==='Installing…'));
      assert.equal(await page.locator('.np-operation-card,.np-operation-tray').count(),0);
      assert.equal(await page.locator('.np-install-ring').count(),2);
      assert.equal(await page.locator('[aria-valuenow]').count(),0);
      await page.evaluate(()=>NPPlugins.install('demo-0','Simple Voice Chat'));
      assert.equal(posts.length,2,'duplicate click must not repeat POST');
      for(const width of [1440,390,320])for(const mode of ['tiles','list']){
        await page.setViewportSize({width,height:900});await page.evaluate(mode=>{viewMode=mode;R();},mode);
        assert.equal(await page.locator('.np-install-ring').count(),2);
        assert.ok(await page.locator('.np-install-btn').first().isDisabled());
        assert.ok(await page.locator('.np-install-btn').first().evaluate(el=>el.scrollWidth<=el.clientWidth+1));
        assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1));
      }
      await page.emulateMedia({reducedMotion:'reduce'});
      assert.equal(await page.locator('.np-install-arc').first().evaluate(el=>getComputedStyle(el).animationName),'none');
      await page.emulateMedia({reducedMotion:'no-preference'});
      await page.evaluate(()=>document.documentElement.dataset.panelMotion='reduced');
      assert.equal(await page.locator('.np-install-arc').first().evaluate(el=>getComputedStyle(el).animationName),'none');
      await page.evaluate(()=>{delete document.documentElement.dataset.panelMotion;viewMode='tiles';R();});
      assert.equal(await page.locator('.np-install-arc').first().evaluate(el=>getComputedStyle(el).animationName),'np-install-turn');
      if(process.env.PI_SCRATCH_DIR){await page.setViewportSize({width:390,height:900});await page.screenshot({path:path.join(process.env.PI_SCRATCH_DIR,`plugin-button-${theme}.png`)});}
      await pending.get('demo-0').fulfill({contentType:'application/json',body:'{}'});
      await page.waitForFunction(()=>document.querySelectorAll('.np-install-ring').length===1);
      const completed=page.locator('[data-np-install][data-slug="demo-0"]');
      assert.equal(await completed.innerText(),'Installed');
      assert.equal(await completed.isDisabled(),true);
      assert.equal(await completed.getAttribute('aria-busy'),'false');
      assert.match(await completed.getAttribute('class'),/is-installed/);
      assert.equal(await completed.evaluate(el=>getComputedStyle(el).opacity),'1');
      assert.doesNotMatch(await page.locator('#main').innerText(),/Install again|Installed · restart required/);
      await page.evaluate(()=>NPPlugins.install('demo-0','Simple Voice Chat'));
      assert.equal(posts.length,2,'completed installs cannot be submitted again');
      await pending.get('demo-1').fulfill({status:502,contentType:'application/json',body:'{"error":"fixture failure"}'});
      await page.waitForFunction(()=>document.querySelectorAll('.np-install-ring').length===0);
      assert.equal(await page.locator('[data-np-install][data-slug="demo-1"]').isDisabled(),false);
      assert.ok(await page.evaluate(()=>messages.some(m=>m[0].includes('Check installed files'))));
      assert.equal(await page.locator('.np-operation-tray').count(),0);assert.deepEqual(errors,[]);
      await page.close();
    }
  }finally{await browser.close();}
});
