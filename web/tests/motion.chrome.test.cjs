// Real app, real CSS and interactions; every request is fulfilled from local demo fixtures.
const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const root = path.join(__dirname, '../static');
const index = fs.readFileSync(path.join(root, 'index.html'), 'utf8');

test('Chrome: navigation vs refresh, modal/menu lifecycle, preferences, static loading and mobile themes', {skip:!process.env.NP_LAYOUT_PLAYWRIGHT, timeout:90000}, async () => {
  const {chromium} = require(process.env.NP_LAYOUT_PLAYWRIGHT);
  const browser = await chromium.launch({headless:true, executablePath:process.env.NP_LAYOUT_CHROME});
  try {
    const page = await browser.newPage({viewport:{width:1440,height:1000}}), errors = [];
    page.on('pageerror', error => errors.push(error.message));
    await page.route('**/*', route => {
      const url = new URL(route.request().url());
      if (url.origin !== 'https://neopanel.test') return route.fulfill({status:200,body:''});
      if (url.pathname === '/') return route.fulfill({contentType:'text/html',body:index});
      if (url.pathname === '/api/branding/appearance.js') return route.fulfill({contentType:'text/javascript',body:''});
      if (url.pathname.startsWith('/api/')) {
        let data = {};
        if (url.pathname === '/api/auth/status') data = {needs_setup:false};
        if (url.pathname === '/api/auth/me') data = {id:1,username:'Demo admin',role:'admin',email:'demo@example.invalid'};
        if (url.pathname === '/api/cluster/servers') data = {servers:[],node_errors:{}};
        if (['/api/nodes','/api/account/activity','/api/servers'].includes(url.pathname)) data = [];
        return route.fulfill({contentType:'application/json',body:JSON.stringify(data)});
      }
      const file = path.join(root, url.pathname);
      return fs.existsSync(file) && fs.statSync(file).isFile() ? route.fulfill({path:file}) : route.fulfill({status:404,body:''});
    });
    await page.addInitScript(() => {
      localStorage.setItem('neopanel_token','demo-fixture-only');
      window.motionCalls = [];
      const animate = Element.prototype.animate;
      Element.prototype.animate = function(frames, options) {
        motionCalls.push({frames,options,tag:this.tagName,className:this.className});
        return animate.call(this, frames, options);
      };
    });
    await page.goto('https://neopanel.test/', {waitUntil:'domcontentloaded'});
    await page.getByRole('heading',{name:'Your servers',exact:true}).waitFor();
    await page.emulateMedia({reducedMotion:'no-preference'});
    const count = () => page.evaluate(() => motionCalls.length);
    let before = await count();
    await page.getByRole('link',{name:'Account',exact:true}).click();
    await page.waitForFunction(() => S.view === 'acc');
    assert.ok(await count() > before, 'actual account navigation enters');
    assert.equal(await page.evaluate(() => {const n=motionCalls.length; R(); R(); return motionCalls.length-n;}), 0, 'same-route R does not enter');
    before = await count();
    await page.getByRole('link',{name:'Servers',exact:true}).click();
    await page.getByRole('heading',{name:'Your servers',exact:true}).waitFor();
    assert.ok(await count() > before, 'return navigation enters, not just first session visit');
    assert.ok(await page.evaluate(() => motionCalls.every(c => c.options.duration <= 180 && !c.options.delay && c.frames.every(f => !('transform' in f)))));

    // Actual R + loaders: keep loading still, then reveal content once, never on stats refresh.
    const lifecycle = await page.evaluate(() => {
      const server = {id:'demo',node_id:'local',type:'paper',name:'Demo server',port:25565,status:'offline'};
      S.servers=[server]; S.cs=0; S.view='fi'; S.fileList=[]; S.loading.files=true;
      S.serverAccess={key:accessIdentity(),status:'ready',data:{is_admin:true,is_owner:true,permissions:[]}};
      const n=motionCalls.length; R();
      const loading=motionCalls.length-n;
      const staticShapes=[...document.querySelectorAll('.np-loading,.np-loading *')].every(el => getComputedStyle(el).animationName==='none' && el.getAnimations().length===0);
      S.loading.files=false; S.fileList=[{name:'demo.txt',path:'demo.txt',size:20}]; R();
      const resolved=motionCalls.length-n;
      S.cpu=78; S.mem=512; R(); R();
      return {loading,staticShapes,resolved,refresh:motionCalls.length-n-resolved};
    });
    assert.equal(lifecycle.loading,0); assert.equal(lifecycle.staticShapes,true);
    assert.ok(lifecycle.resolved>0); assert.equal(lifecycle.refresh,0);
    assert.equal(await page.evaluate(() => { S.view='dash'; R(); S.view='fi'; R(); R(); return NPMotion.activeCount; }), 0, 'rapid switches release detached effects');
    const editor = await page.evaluate(() => {
      S.editingFile='demo.txt'; S.files['demo.txt']='Demo data\n'.repeat(200); R();
      const ed=document.querySelector('#ed'); ed.focus(); ed.setSelectionRange(12,20); ed.scrollTop=100;
      const before={focus:document.activeElement===ed,start:ed.selectionStart,end:ed.selectionEnd,scroll:ed.scrollTop};
      NPMotion.afterRender(document.querySelector('#main'),'demo-editor-navigation');
      return {before,after:{focus:document.activeElement===ed,start:ed.selectionStart,end:ed.selectionEnd,scroll:ed.scrollTop},count:NPMotion.activeCount};
    });
    assert.deepEqual(editor.after,editor.before); assert.equal(editor.count,0);
    await page.evaluate(() => {S.editingFile=null;R();});

    // Real create form/custom select, confirmation resolver and command palette.
    for (const width of [1440,390,320]) for (const theme of ['dark','light']) {
      await page.setViewportSize({width,height:1000});
      await page.locator('.file-row').first().click({button:'right'});
      const contextMenu=page.locator('#floating-ctx');
      assert.equal(await contextMenu.evaluate(el=>getComputedStyle(el).animationName),'np-motion-fade');
      assert.equal(await contextMenu.evaluate(el=>getComputedStyle(el).transform),'none');
      assert.ok(await contextMenu.locator('button').evaluateAll(els=>els.every(el=>getComputedStyle(el).animationName==='none')));
      await page.keyboard.press('Escape');
      assert.equal(await contextMenu.isVisible(),false);
      await page.evaluate(theme => {document.documentElement.dataset.theme=theme; openCreateServerModal();}, theme);
      await page.locator('#cs-name').waitFor();
      const box = page.locator('#modal-container .pbox');
      assert.ok(await box.evaluate(el => el.scrollWidth<=el.clientWidth+1));
      assert.equal(await box.evaluate(el => getComputedStyle(el).transform),'none');
      assert.equal(await box.evaluate(el => getComputedStyle(el).animationName),'none', 'only overlay fades, no nested dialog zoom');
      assert.equal(await page.locator('#modal-container .pal').evaluate(el => getComputedStyle(el).animationName),'np-motion-fade');
      await page.locator('#cs-name').focus(); await page.keyboard.press('Tab');
      assert.equal(await page.evaluate(() => getComputedStyle(document.activeElement).outlineStyle),'solid');
      const trigger = page.locator('#modal-container .cselect-trigger').first();
      await trigger.click();
      const menu = page.locator('.cselect-menu:visible').first();
      await menu.waitFor();
      const geometry = await menu.evaluate(el => {const r=el.getBoundingClientRect(),s=getComputedStyle(el);return {x:r.x,y:r.y,right:r.right,bottom:r.bottom,transform:s.transform,animation:s.animationName};});
      assert.ok(geometry.x>=0 && geometry.right<=width+1 && geometry.y>=0 && geometry.bottom<=1001, JSON.stringify(geometry));
      assert.equal(geometry.transform,'none'); assert.equal(geometry.animation,'np-motion-fade');
      await page.keyboard.press('Escape');
      // Closing the containing modal must also clean up the body-mounted menu immediately.
      if (process.env.NP_MOTION_ARTIFACTS) {
        fs.mkdirSync(process.env.NP_MOTION_ARTIFACTS,{recursive:true});
        await page.screenshot({path:path.join(process.env.NP_MOTION_ARTIFACTS,`modal-menu-${width}-${theme}.png`)});
      }
      assert.equal(await page.evaluate(() => {closeModal();return document.querySelector('#modal-container').childElementCount;}),0);
      assert.equal(await page.locator('.cselect-menu:visible').count(),0);
      await page.evaluate(() => {window.demoAnswer='pending';uiConfirm({title:'Demo confirmation',message:'Demo data only. Continue?'}).then(value=>demoAnswer=value);});
      await page.waitForFunction(() => document.activeElement.id==='ui-modal-confirm-btn');
      await page.getByRole('button',{name:'Cancel',exact:true}).click();
      assert.equal(await page.evaluate(() => demoAnswer),false);
      assert.equal(await page.locator('#modal-container .pal').count(),0);
      const race = await page.evaluate(async () => {
        const first=uiConfirm('Demo first'); closeModal();
        const second=uiConfirm('Demo second'); closeUiModal(true);
        return [await first,await second,document.querySelector('#modal-container').childElementCount];
      });
      assert.deepEqual(race,[false,true,0]);
    }
    await page.setViewportSize({width:1440,height:1000});
    await page.evaluate(() => {S.view='dash';R();});
    const search = page.getByRole('button',{name:'Search (Ctrl+K)',exact:true});
    await search.focus(); await page.keyboard.press('Enter');
    await page.locator('#pq').waitFor();
    assert.equal(await page.evaluate(() => document.activeElement.id),'pq');
    await page.keyboard.press('Escape');
    assert.equal(await page.locator('#pal').count(),0);
    assert.equal(await page.evaluate(() => document.activeElement.getAttribute('aria-label')),'Search (Ctrl+K)');

    // Panel and OS preferences veto WAAPI and all CSS independently, including runtime changes.
    for (const panelReduced of [true,false]) {
      await page.emulateMedia({reducedMotion:panelReduced?'no-preference':'reduce'});
      await page.evaluate(panelReduced => {
        document.documentElement.dataset.panelMotion=panelReduced?'reduced':'full';
        S.view='acc';R();openCreateServerModal();
      }, panelReduced);
      assert.equal(await page.evaluate(() => NPMotion.activeCount),0);
      assert.ok(await page.locator('#modal-container,#modal-container *').evaluateAll(els => els.every(el => getComputedStyle(el).animationName==='none' && getComputedStyle(el).transitionDuration==='0s')));
      await page.evaluate(() => closeModal());
    }
    await page.emulateMedia({reducedMotion:'no-preference'});
    await page.evaluate(() => {document.documentElement.dataset.panelMotion='full';S.view='dash';R();document.documentElement.dataset.panelMotion='reduced';});
    assert.equal(await page.evaluate(() => NPMotion.activeCount),0, 'changing panel preference cancels an in-flight entrance');
    await page.emulateMedia({reducedMotion:'reduce'}); await page.emulateMedia({reducedMotion:'no-preference'});
    assert.equal(await page.evaluate(() => document.documentElement.dataset.panelMotion),'reduced', 'OS never overwrites panel preference');

    // Feature result cards refresh without CSS replays; positioned overlays keep their transforms.
    const featureCSS = await page.evaluate(() => {
      document.documentElement.dataset.panelMotion='full';
      const demo=document.createElement('div');
      demo.innerHTML='<div class="np-project-result">Demo plugin</div><div class="npi-file">Demo installed file</div><div class="bulk-action-bar">Demo selection</div>';
      const fleet=document.createElement('div');fleet.className='np-server-grid';fleet.innerHTML='<div class="sift-server-card">Demo server</div>';demo.append(fleet);
      document.body.append(demo);
      const styles=[...demo.querySelectorAll('.np-project-result,.npi-file,.bulk-action-bar,.sift-server-card')].map(el=>({animation:getComputedStyle(el).animationName,transform:getComputedStyle(el).transform}));demo.remove();return styles;
    });
    assert.equal(featureCSS[0].animation,'none'); assert.equal(featureCSS[1].animation,'none');
    assert.notEqual(featureCSS[2].transform,'none');
    assert.equal(featureCSS[3].animation,'none');
    assert.equal(await page.locator('.np-motion-ripple-layer').count(),0);
    assert.deepEqual(errors,[]);
  } finally { await browser.close(); }
});
