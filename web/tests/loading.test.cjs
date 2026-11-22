// All browser traffic is fulfilled locally. No server, live APIs or real credentials.
const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const root = path.join(__dirname, '../static');
const read = file => fs.readFileSync(path.join(root, file), 'utf8');
const index = read('index.html'), app = read('js/app.js');
const loading = require('../static/js/loading.js');
const kinds = ['servers', 'files', 'plugins', 'installed', 'players', 'backups', 'schedules', 'databases', 'activity'];
const modes = ['list', 'tiles', 'icons'];
const artifacts = process.env.NP_LOADING_ARTIFACTS || (process.env.PI_SCRATCH_DIR && path.join(process.env.PI_SCRATCH_DIR, 'skeleton-loading'));

test('bounded, escaped, deterministic presentation; action and SQL spinners remain opt-in', () => {
  for (const kind of kinds) for (const mode of modes) {
    const html = loading.html(kind, '<img src=x onerror="bad()"> & \'', mode);
    assert.equal(html, loading.html(kind, '<img src=x onerror="bad()"> & \'', mode));
    assert.equal((html.match(/role="status"/g) || []).length, 1);
    assert.equal((html.match(/class="np-load-item"/g) || []).length, kind === 'servers' ? 3 : mode === 'icons' ? 6 : 4);
    assert.match(html, /&lt;img src=x onerror=&quot;bad\(\)&quot;&gt; &amp; &#39;/);
    assert.doesNotMatch(html, /<(button|a|input|img|script)\b|tabindex|aria-busy|style=|onclick=/);
    assert.match(html, /class="np-load-shapes" aria-hidden="true"/);
  }
  assert.equal(loading.html('__proto__', 'x', '\" onclick=bad'), loading.html('files', 'x', 'list'));
  const context = { window: { NPLoading: loading }, esc: s => String(s).replaceAll('<', '&lt;') };
  vm.createContext(context);
  vm.runInContext(app.slice(app.indexOf('function loaderHtml('), app.indexOf('function setBtnLoading(')), context);
  assert.match(context.loaderHtml('Executing SQL statement...'), /np-spinner/);
  assert.doesNotMatch(context.loaderHtml('Executing SQL statement...'), /np-loading/);
  assert.match(context.loaderHtml('Loading files...', 'files'), /np-loading-files/);
  context.window.NPLoading = undefined;
  assert.match(context.loaderHtml('Loading files...', 'files'), /np-spinner/);
  assert.match(app, /loaderHtml\('Executing SQL statement\.\.\.'\)/);
  assert.match(app, /loaderHtml\('Inspecting tables\.\.\.'\)/);
  assert.ok(index.indexOf('/js/loading.js') < index.indexOf('/js/plugin-details.js'));
});

async function fixture(browser, scenario = 'login', javaScriptEnabled = true) {
  const context = await browser.newContext({ javaScriptEnabled, viewport: {width:1440, height:1000} });
  const page = await context.newPage(), requests = [], errors = [];
  let release;
  const gate = new Promise(resolve => { release = resolve; });
  page.on('pageerror', e => errors.push(e.message));
  await page.route('**/*', async route => {
    const url = new URL(route.request().url());
    if (url.origin !== 'https://neopanel.test') return route.fulfill({status:200, body:''});
    if (url.pathname === '/') return route.fulfill({contentType:'text/html', body:index});
    if (url.pathname === '/api/branding/appearance.js') return route.fulfill({contentType:'text/javascript', body:''});
    if (url.pathname.startsWith('/api/')) {
      requests.push(url.pathname + url.search);
      let data = {};
      if (url.pathname === '/api/auth/status') {
        await gate;
        if (scenario === 'error') return route.fulfill({status:503, body:'not json'});
        data = {needs_setup:scenario === 'setup'};
      }
      if (url.pathname === '/api/auth/me') {
        if (scenario === 'expired') return route.fulfill({status:401, contentType:'application/json', body:'{}'});
        data = {id:1, username:'Demo admin', role:'admin', email:''};
      }
      if (url.pathname === '/api/auth/invitations/inspect') {
        if (scenario === 'invite-unavailable') return route.fulfill({status:410,contentType:'application/json',body:JSON.stringify({error:'Invitation expired.'})});
        data = {server_name:'Demo invitation',email:'demo@example.test',account_exists:true,permissions:['files.view'],expires_at:'2030-01-01T00:00:00Z'};
      }
      if (url.pathname === '/api/cluster/servers') data = {servers:[], node_errors:{}};
      if (['/api/nodes', '/api/account/activity', '/api/servers'].includes(url.pathname)) data = [];
      return route.fulfill({contentType:'application/json', body:JSON.stringify(data)});
    }
    const file = path.join(root, url.pathname);
    if (fs.existsSync(file) && fs.statSync(file).isFile()) return route.fulfill({path:file});
    return route.fulfill({status:404, body:''});
  });
  if (scenario === 'authenticated' || scenario === 'expired') await page.addInitScript(() => localStorage.setItem('neopanel_token', 'demo-fixture-only'));
  if (scenario === 'invite-stored') await page.addInitScript(() => sessionStorage.setItem('neopanel_pending_invite','fixture-invite'));
  await page.goto('https://neopanel.test/' + (scenario.startsWith('invite-') && scenario !== 'invite-stored' ? '#invite-token=fixture-invite' : ''), {waitUntil:'domcontentloaded'});
  return {page, context, requests, errors, release};
}

async function accessibleShape(page, selector = '.np-loading') {
  const issues = await page.locator(selector).evaluateAll(els => els.flatMap(el => {
    const issues = [];
    if (el.querySelectorAll('[role="status"]').length !== 1) issues.push('status count');
    if (el.querySelector('[role="status"]').closest('[aria-busy="true"]')) issues.push('busy status');
    if (el.querySelector('button,a,input,select,textarea,[tabindex],[onclick]')) issues.push('fake control');
    for (const bar of el.querySelectorAll('.np-load-bar,.np-load-cover')) if (!bar.closest('[aria-hidden="true"]')) issues.push('exposed decoration');
    for (const n of [el, ...el.querySelectorAll('div,span')]) if (n.clientWidth && n.scrollWidth > n.clientWidth + 1) issues.push('overflow: ' + n.className);
    return issues;
  }));
  assert.deepEqual(issues, []);
}

async function shot(page, name) {
  if (!artifacts) return;
  fs.mkdirSync(artifacts, {recursive:true});
  await page.screenshot({path:path.join(artifacts, name + '.png'), fullPage:true});
  // Offline preview uses real rendered markup/styles without application scripts or remote fonts.
  const preview = (await page.content()).replace(/<script\b[^>]*>[\s\S]*?<\/script>/g, '')
    .replace(/<link[^>]*href="\/(css\/[^\"]+)"[^>]*>/g, (_, file) => '<style>' + read(file).replace(/@import[^;]+;/g, '') + '</style>')
    .replace(/<link\b[^>]*>/g, '').replace(/src="https?:[^\"]*"/g, 'src="data:,"').replace(/url\(['"]?https?:[^)]+\)/g, 'none');
  fs.writeFileSync(path.join(artifacts, name + '.html'), preview);
}

test('Chrome: static pre-JS boot and real auth/login/setup/error replacement (mock APIs)', {skip:!process.env.NP_LAYOUT_PLAYWRIGHT, timeout:60000}, async () => {
  const {chromium} = require(process.env.NP_LAYOUT_PLAYWRIGHT);
  const browser = await chromium.launch({headless:true, executablePath:process.env.NP_LAYOUT_CHROME});
  try {
    const f = await fixture(browser, 'login', false);
    for (const width of [1440,390,320]) for (const theme of ['dark','light']) {
      await f.page.setViewportSize({width,height:900});
      await f.page.evaluate(theme => document.documentElement.dataset.theme=theme, theme);
      await accessibleShape(f.page, '.np-loading-boot');
      assert.equal(await f.page.locator('.np-loading-boot').count(), 1);
      assert.ok(await f.page.evaluate(() => document.querySelector('#main').scrollWidth <= document.querySelector('#main').clientWidth + 1));
      await shot(f.page, `boot-before-js-${width}-${theme}`);
    }
    assert.deepEqual(f.requests, [], 'no API requests from static shell');
    await f.context.close();
    for (const scenario of ['login','setup','error','expired','authenticated','invite-valid','invite-unavailable','invite-stored']) {
      const f = await fixture(browser, scenario);
      await f.page.locator('.np-loading-boot').waitFor();
      await accessibleShape(f.page, '.np-loading-boot');
      f.release();
      await f.page.locator('.np-loading-boot').waitFor({state:'detached'});
      if (scenario === 'authenticated') await f.page.getByRole('heading', {name:'Your servers', exact:true}).waitFor();
      else if (scenario.startsWith('invite-')) await f.page.getByRole('heading',{name:scenario === 'invite-unavailable' ? 'Invitation unavailable' : 'Join Demo invitation',exact:true}).waitFor();
      else await f.page.locator(scenario === 'setup' ? '#setup-form' : '#login').waitFor();
      assert.deepEqual(f.errors, [], scenario);
      const publicPaths = scenario.startsWith('invite-') ? ['/api/branding','/api/auth/status','/api/auth/invitations/inspect'] : ['/api/branding', '/api/community', '/api/auth/status', '/api/mods'];
      if (scenario === 'expired' || scenario === 'authenticated') publicPaths.push('/api/auth/me');
      if (scenario === 'authenticated') publicPaths.push('/api/cluster/servers','/api/nodes','/api/account/activity?limit=200','/api/versions');
      assert.deepEqual([...new Set(f.requests)].sort(), publicPaths.sort(), `only existing requests: ${scenario}`);
      await shot(f.page, 'auth-' + scenario);
      await f.context.close();
    }
  } finally { await browser.close(); }
});

test('Chrome: real renderers, modes/themes/mobile, reduced motion, refresh retention, search retry and keyboard', {skip:!process.env.NP_LAYOUT_PLAYWRIGHT, timeout:120000}, async () => {
  const {chromium} = require(process.env.NP_LAYOUT_PLAYWRIGHT);
  const browser = await chromium.launch({headless:true, executablePath:process.env.NP_LAYOUT_CHROME});
  try {
    const f = await fixture(browser, 'authenticated'), page = f.page;
    f.release(); await page.getByRole('heading', {name:'Your servers', exact:true}).waitFor();
    // Actual app state and renderers; fixtures are explicitly demo data, not service responses.
    await page.evaluate(() => {
      window.demoServer = {id:'demo',node_id:'local',type:'paper',name:'Demo server',port:25565,status:'offline'};
      S.servers = [demoServer]; S.cs = 0; S.status = 'offline';
      S.serverAccess = {key:accessIdentity(),status:'ready',data:{is_admin:true,is_owner:true,permissions:[]}};
      S.players = []; S.bk = []; S.databases = []; S.serverSchedules = []; S.fileList = []; S.plugins = []; S.acct.act = [];
      S.loading = {dash:true,files:true,players:true,backups:true,databases:true,schedules:true,plugins:true,activity:true};
      S.ptab = 'online'; S.currentDir = ''; S.plugTab = 'browse'; S.lastModrinthHits = null;
      window.demoMode = 'list'; window.getViewMode = () => demoMode;
      window.drawDemo = view => {
        S.view = view === 'installed' ? 'plug' : view === 'activity' ? 'act' : view;
        S.plugTab = view === 'installed' ? 'installed' : 'browse';
        if (view === 'dash') { S.servers = []; S.cs = -1; } else { S.servers = [demoServer]; S.cs = 0; }
        const html = view === 'installed' ? V.plug() : view === 'activity' ? renderDashRecentActivityHtml() : V[view]();
        document.querySelector('#main').innerHTML = '<p class="mu">Demo data · Loading fixtures</p>' + html;
      };
    });
    const before = f.requests.length;
    const views = ['dash','fi','plug','installed','pl','bk','sc','db','activity','act'];
    let combinations = 0;
    for (const width of [1440,390,320]) for (const theme of ['dark','light']) for (const mode of modes) {
      await page.setViewportSize({width,height:1000});
      await page.evaluate(({theme,mode}) => {document.documentElement.dataset.theme=theme;demoMode=mode;}, {theme,mode});
      for (const view of views) {
        await page.evaluate(view => drawDemo(view), view);
        assert.ok(await page.locator('.np-loading').count() >= 1, `${view} has skeleton`);
        await accessibleShape(page);
        const main = await page.locator('#main').evaluate(el => ({scroll:el.scrollWidth,width:el.clientWidth}));
        assert.ok(main.scroll <= main.width + 1, `${view}/${width}/${theme}/${mode}: main overflow ${JSON.stringify(main)}`);
        combinations++;
        if (mode === 'list' && (width === 1440 && theme === 'dark' || width === 320 && theme === 'light')) await shot(page, `${view}-${width}-${theme}`);
      }
    }
    await page.evaluate(() => toggleCompact(true));
    for (const view of views) {
      await page.evaluate(view => drawDemo(view), view);
      await accessibleShape(page);
      assert.ok(await page.locator('#main').evaluate(el => el.scrollWidth <= el.clientWidth + 1), `compact ${view}`);
    }
    await page.evaluate(() => toggleCompact(false));
    assert.equal(f.requests.length, before, 'rendering placeholders adds no API requests');
    // Explicit OS and panel reduced-motion preferences, independently.
    for (const panelReduced of [false,true]) {
      await page.emulateMedia({reducedMotion:panelReduced ? 'no-preference' : 'reduce'});
      await page.evaluate(panelReduced => { document.documentElement.dataset.panelMotion=panelReduced?'reduced':'full'; drawDemo('fi'); }, panelReduced);
      assert.ok(await page.locator('.np-loading,.np-loading *').evaluateAll(els => els.every(el => getComputedStyle(el).animationName === 'none' && getComputedStyle(el).transitionDuration === '0s')));
    }
    // Loading + usable data keeps real content, not skeleton replacements.
    await page.evaluate(() => {
      S.fileList = [{name:'demo.txt',path:'demo.txt',size:20}];
      S.plugins = [['demo.jar','20 B',true,'',false,20]];
      S.players = [{n:'DemoPlayer'}]; S.bk = [{name:'demo.tar.gz',size:20}];
      S.databases = [{id:'demo',name:'Demo database',engine:'sqlite'}];
      S.serverSchedules = [{id:'demo',name:'Demo schedule',cron:'0 0 * * *',tasks:[]}];
      S.acct.act = [{action:'server.start',username:'Demo admin',details:'Demo event'}];
    });
    for (const view of ['fi','installed','pl','bk','db','activity']) {
      await page.evaluate(view => drawDemo(view), view);
      assert.equal(await page.locator('.np-loading').count(), 0, `${view}: preserve loaded data`);
    }
    // Schedules historically replace while loading; old data is not server-keyed.
    // Keep that gate so a server switch cannot expose another server's schedules.
    await page.evaluate(() => drawDemo('sc'));
    assert.equal(await page.locator('.np-loading-schedules').count(),1);
    assert.equal(await page.getByText('Demo schedule',{exact:true}).count(),0);
    await page.evaluate(() => {S.loading.schedules=false;drawDemo('sc');});
    assert.equal(await page.locator('.np-loading').count(),0);
    // Test the actual search renderer's delayed initial request, error and retry.
    await page.evaluate(() => {
      S.view='plug'; S.plugTab='browse'; S.loading.plugins=false; S.lastModrinthHits=[]; S.q='demo';
      window.demoCalls=[]; window.api=(url,opts) => { demoCalls.push(url); return new Promise((resolve,reject) => {window.finishSearch=resolve;window.failSearch=reject;}); };
      drawDemo('plug'); window.searchPending=NPPlugins.search();
    });
    await page.locator('.np-loading-plugins').waitFor(); await accessibleShape(page);
    assert.equal(await page.evaluate(() => demoCalls.length), 1);
    await page.evaluate(async () => {failSearch(Error('Demo offline')); await searchPending;});
    await page.getByRole('button',{name:'Retry',exact:true}).waitFor();
    assert.equal(await page.locator('.np-loading').count(),0);
    await page.getByRole('button',{name:'Retry',exact:true}).focus();
    assert.equal(await page.getByRole('button',{name:'Retry',exact:true}).evaluate(el => getComputedStyle(el).outlineStyle),'solid');
    await page.keyboard.press('Enter'); await page.locator('.np-loading-plugins').waitFor();
    assert.equal(await page.evaluate(() => demoCalls.length), 2);
    await page.evaluate(() => finishSearch({hits:[{slug:'demo',project_id:'demo',title:'Demo plugin',description:'Demo result'}],total_hits:1}));
    await page.locator('[data-project-card]').waitFor();
    await page.evaluate(() => {window.searchPending=NPPlugins.search();});
    assert.equal(await page.locator('.np-loading').count(),0,'background search keeps previous results');
    assert.equal(await page.locator('[data-project-card]').count(),1);
    await page.evaluate(async () => {finishSearch({hits:[],total_hits:0}); await searchPending;});
    await page.getByText('No matching projects. Try another search or browsing filter.').waitFor();
    assert.equal(await page.locator('.np-loading').count(),0);
    assert.deepEqual(f.errors, []);
    await page.evaluate(() => {
      S.view='fi'; S.loading.files=true; S.fileList=[];
      S.serverAccess={key:accessIdentity(),status:'error',error:'Demo permission failure'};
      R();
    });
    await page.getByRole('heading',{name:'Access restricted',exact:true}).waitFor();
    assert.equal(await page.locator('.np-loading').count(),0,'permission failure is not disguised as loading');
    assert.deepEqual(f.errors, []);
    if (artifacts) fs.writeFileSync(path.join(artifacts,'results.json'), JSON.stringify({combinations, widths:[1440,390,320], themes:['dark','light'], modes, bootScenarios:['login','setup','error','expired','authenticated'], network:'All requests fulfilled locally; no additional skeleton API requests'},null,2));
    await f.context.close();
  } finally { await browser.close(); }
});
