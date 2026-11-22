const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const root = path.join(__dirname, '../static');

test('Chrome: cards show software only above a separate detected version', {skip:!process.env.NP_LAYOUT_PLAYWRIGHT,timeout:60000}, async()=>{
  const {chromium}=require(process.env.NP_LAYOUT_PLAYWRIGHT);
  const browser=await chromium.launch({headless:true,executablePath:process.env.NP_LAYOUT_CHROME});
  try {
    const page=await browser.newPage({viewport:{width:1440,height:1000}}),errors=[];
    page.on('pageerror',e=>errors.push(e.message));
    await page.route('**/*',route=>{
      const url=new URL(route.request().url());
      if(url.origin!=='https://neopanel.test')return route.fulfill({status:200,body:''});
      if(url.pathname==='/')return route.fulfill({path:path.join(root,'index.html')});
      if(url.pathname==='/api/branding/appearance.js')return route.fulfill({contentType:'text/javascript',body:''});
      if(url.pathname.startsWith('/api/')){
        let data={};
        if(url.pathname==='/api/auth/status')data={needs_setup:false};
        if(url.pathname==='/api/auth/me')data={id:1,username:'Demo admin',role:'admin'};
        if(url.pathname==='/api/cluster/servers')data={servers:[],node_errors:{}};
        if(['/api/nodes','/api/account/activity','/api/servers'].includes(url.pathname))data=[];
        return route.fulfill({contentType:'application/json',body:JSON.stringify(data)});
      }
      const file=path.join(root,url.pathname);
      return fs.existsSync(file)&&fs.statSync(file).isFile()?route.fulfill({path:file}):route.fulfill({status:404,body:''});
    });
    await page.addInitScript(()=>localStorage.setItem('neopanel_token','demo-fixture-only'));
    await page.goto('https://neopanel.test/',{waitUntil:'domcontentloaded'});
    await page.getByRole('heading',{name:'Your servers',exact:true}).waitFor();
    await page.evaluate(()=>{
      S.servers=[{id:'demo',node_id:'local',name:'Survival Server',type:'velocity',version:'latest',jar_file:'server.jar',java_path:'auto',port:25565,memory_max:2048,stats:{status:'offline'},artifact:{status:'identified',software:'velocity',project_version:'4.3.2-SNAPSHOT',build_name:'#58',build_id:'long-build-uuid'}}];
      S.cs=0;S.status='offline';S.loading.dash=false;S.view='dash';
      S.serverAccess={key:accessIdentity(),status:'ready',data:{is_admin:true,is_owner:true,permissions:[]}};
      R();
    });
    for(const width of [1440,390,320])for(const theme of ['dark','light'])for(const mode of ['tiles','icons','list']){
      await page.setViewportSize({width,height:1000});
      await page.evaluate(({theme,mode})=>{document.documentElement.dataset.theme=theme;S.viewModes.dash=mode;R();},{theme,mode});
      const card=page.locator('#dash-srv-local-demo');
      assert.equal((await card.locator('.artifact-badge').innerText()).trim(),'Velocity');
      assert.equal((await card.locator('.artifact-version').innerText()).trim(),'4.3.2-SNAPSHOT');
      assert.doesNotMatch(await card.innerText(),/Release|Build|long-build-uuid/);
      assert.ok(await card.evaluate(el=>el.scrollWidth<=el.clientWidth+1),'card should not overflow');
      if(width===390&&theme==='dark'&&mode==='tiles'&&process.env.PI_SCRATCH_DIR){
        await card.screenshot({path:path.join(process.env.PI_SCRATCH_DIR,'server-name-fixed.png')});
      }
    }
    await page.evaluate(()=>{S.view='con';S.activeNodeId='local';S.activeServerId='demo';R();});
    assert.equal((await page.locator('#srv-hero-engine').innerText()).trim(),'Velocity');
    assert.equal((await page.locator('#srv-hero-version').innerText()).trim(),'4.3.2-SNAPSHOT');
    await page.evaluate(()=>{S.servers[0].artifact={status:'unknown',software:'velocity',project_version:'stale'};NPServerArtifact.refresh();});
    assert.equal((await page.locator('#srv-hero-engine').innerText()).trim(),'Custom');
    assert.equal((await page.locator('#srv-hero-version').innerText()).trim(),'Unknown version');
    assert.deepEqual(errors,[]);
  } finally {await browser.close();}
});
