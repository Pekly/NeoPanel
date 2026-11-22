// Mock-only browser regression. Run with NP_PLAYWRIGHT pointing to an installed playwright package.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const http = require('node:http');
const { chromium } = require(process.env.NP_PLAYWRIGHT || 'playwright');
const root = path.resolve(__dirname, '../static');
const app = fs.readFileSync(path.join(root, 'js/app.js'), 'utf8').replace(/\r\n/g, '\n');
const slice = (a, b) => app.slice(app.indexOf(a), app.indexOf(b, app.indexOf(a)));
const markup = slice('      return `${srvh()}\n      <div class="np-editor-workspace">', '\n    }');
const hooks = slice('function applyPermissionControls(root)', '\nV.sftp =') + slice('window.R = function()', '\nfunction bindConsoleInput()');
const server = http.createServer((req, res) => {
  const name = new URL(req.url, 'http://localhost').pathname;
  if (name === '/') { res.setHeader('Content-Type', 'text/html; charset=utf-8'); res.end('<!doctype html><meta charset="utf-8"><title>MOCK API — editor regression</title><link rel="stylesheet" href="/css/editor-syntax.css"><link rel="stylesheet" href="/css/find.css"><style>body{margin:24px;font-family:sans-serif}.np-editor-workspace{max-width:900px}.np-editor-toolbar{padding:12px}#ed{width:100%;height:300px;box-sizing:border-box}#main{max-width:100%}</style><h1>MOCK API — no production requests</h1><main id="main"></main><div id="modal-container"></div>'); return; }
  const file = path.join(root, name);
  if (!file.startsWith(root + path.sep) || !fs.existsSync(file) || !fs.statSync(file).isFile()) { res.writeHead(404).end(); return; }
  res.setHeader('Content-Type', file.endsWith('.js') ? 'text/javascript' : file.endsWith('.css') ? 'text/css' : 'application/octet-stream');
  res.end(fs.readFileSync(file));
});
(async () => {
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
  const browser = await chromium.launch({ executablePath: process.env.NP_CHROME || 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe', headless: true });
  try {
    const page = await browser.newPage(); const errors = []; page.on('pageerror', e => errors.push(e.message));
    await page.goto(`http://127.0.0.1:${server.address().port}/`);
    await page.addScriptTag({ content: `
      var S = { user:{id:'mock-user',role:'admin'}, token:'mock-session', view:'fi', currentDir:'', files:{}, branding:{} };
      var mockServer={id:'mock',node_id:'local'}, permitted=true, posts=0, failSave=false, delaySave=false, finishSave;
      var mockText='key=one\\r\\n'+('line\\r\\n').repeat(1500)+'key=two', mockStamp='2020-02-03T04:05:06Z';
      var $=s=>document.querySelector(s), curServer=()=>mockServer, srvApiPath=(s,p)=>p;
      var esc=s=>String(s??'').replaceAll('&','&amp;').replaceAll('<','&lt;').replaceAll('"','&quot;'), ic=()=>'', srvh=()=>'';
      var toast=message=>{window.lastToast=message;}, _uiModalResolver=null, hasServerPermission=()=>permitted, serverPagePermission=()=>true, canServerPage=()=>true;
      var closeAllCustomSelects=()=>{}, closeFloatingCtx=()=>{}, enhanceSelects=()=>{}, renderFlyConsole=()=>{}, renderAnnouncement=()=>'';
      var bindEditorTabKey=()=>bindQolEditor(), saveActiveFile=()=>saveQolFile();
      var api=async (url,opts)=>{if(opts?.method==='POST'){posts++;if(delaySave)await new Promise(r=>finishSave=r);if(failSave)throw Error('MOCK failure');return {modified_at:'2025-03-04T05:06:07Z'};}return url.includes('preview')?{kind:'text',editable:true}:{content:mockText,modified_at:mockStamp};};
      var V={fi:()=>{${markup}}};
      window.NPWorkspace={captureConsole:()=>{},activate:()=>{},maintenanceBanner:()=>'',connectionBanner:()=>'',afterRender:()=>{},syncControls:()=>{const b=$('#editor-save-btn');if(b)b.disabled=false;}};
    ` });
    for (const file of ['qol.js', 'find-search.js', 'find.js', 'editor-syntax.js']) await page.addScriptTag({ url: '/js/' + file });
    await page.addScriptTag({ content: hooks });
    await page.evaluate(() => openQolFile('config.properties'));
    assert.equal(await page.evaluate(() => window.lastToast), undefined);
    assert.equal(await page.locator('#editor-save-btn').count(), 1, JSON.stringify(errors));
    const disabled = () => page.locator('#editor-save-btn').isDisabled();
    assert.equal(await disabled(), true);
    assert.equal(await page.evaluate(() => QOL.editor.savedAt), '2020-02-03T04:05:06.000Z');
    await page.locator('#ed').focus(); await page.keyboard.press('Control+s');
    await page.evaluate(() => saveActiveFile()); assert.equal(await page.evaluate(() => posts), 0);
    await page.keyboard.press('Control+End'); await page.keyboard.insertText('changed'); assert.equal(await disabled(), false);
    await page.keyboard.press('Control+z'); assert.equal(await disabled(), true);
    await page.keyboard.press('Control+s'); assert.equal(await page.evaluate(() => posts), 0);
    await page.evaluate(() => { R(); applyPermissionControls(document.querySelector('#main')); }); assert.equal(await disabled(), true);
    await page.locator('#ed').focus(); await page.keyboard.press('Control+End'); await page.keyboard.type('edit');
    await page.evaluate(() => { delaySave=true; window.pendingSave=saveActiveFile(); }); assert.equal(await disabled(), true);
    await page.keyboard.type('newer'); await page.evaluate(async () => { finishSave(); await pendingSave; delaySave=false; });
    assert.equal(await disabled(), false); assert.equal(await page.evaluate(() => QOL.editor.base.endsWith('edit')), true);
    await page.evaluate(() => saveActiveFile()); assert.equal(await disabled(), true);
    const label = await page.locator('#editor-save-status').textContent(); const title = await page.locator('#editor-save-status').getAttribute('title');
    assert.ok(label.startsWith('Saved · ') && title); await page.evaluate(() => R()); assert.equal(await page.locator('#editor-save-status').textContent(), label);
    await page.locator('#ed').focus(); await page.keyboard.press('Control+End'); await page.keyboard.type('fail');
    await page.evaluate(async () => { failSave=true; await saveActiveFile(); failSave=false; });
    assert.equal(await page.evaluate(() => QOL.editor.savedAt), '2025-03-04T05:06:07.000Z');
    await page.evaluate(() => { permitted=false; applyPermissionControls(document.querySelector('#main')); }); assert.equal(await disabled(), true);
    const before = await page.evaluate(() => posts); await page.evaluate(() => saveActiveFile()); assert.equal(await page.evaluate(() => posts), before);
    await page.evaluate(() => { permitted=true; R(); });
    await page.locator('#ed').focus(); await page.keyboard.press('Control+f');
    assert.equal(await page.locator('.np-editor-workspace #np-find').count(), 1); assert.equal(await page.locator('body > #np-find').count(), 0);
    for (const width of [390, 1280]) {
      await page.setViewportSize({width, height:900});
      assert.equal(await page.locator('#np-find').evaluate(bar => { const a=bar.getBoundingClientRect(), b=bar.closest('.np-editor-workspace').getBoundingClientRect(); return a.left >= b.left && a.right <= b.right && a.top >= b.top && a.bottom <= b.bottom && getComputedStyle(bar).position !== 'fixed'; }), true);
    }
    await page.locator('#np-find-query').fill('key'); await page.waitForFunction(() => document.querySelector('#np-find-count').textContent === '1 of 2');
    await page.keyboard.press('Enter'); assert.equal(await page.locator('#np-find-count').textContent(), '2 of 2');
    assert.ok(await page.locator('#ed').evaluate(ed => ed.scrollTop > 10000 && ed.value.slice(ed.selectionStart, ed.selectionEnd) === 'key'));
    await page.keyboard.press('Shift+Enter'); assert.equal(await page.locator('#np-find-count').textContent(), '1 of 2');
    await page.evaluate(() => R()); assert.equal(await page.locator('.np-editor-workspace #np-find').count(), 1);
    await page.keyboard.press('Escape'); assert.equal(await page.locator('#np-find').count(), 0); assert.equal(await page.locator('#ed').evaluate(ed => document.activeElement === ed), true);
    // Worker still searches the far end of a 5 MiB textarea, with bounded result counts.
    await page.evaluate(() => { document.querySelector('#ed').value='x'.repeat(5*1024*1024-6)+'needle'; R(); });
    await page.keyboard.press('Control+f'); await page.locator('#np-find-query').fill('needle');
    await page.waitForFunction(() => document.querySelector('#np-find-count').textContent === '1 of 1');
    assert.equal(await page.locator('#ed').evaluate(ed => ed.selectionStart), 5*1024*1024-6);
    await page.evaluate(() => { S.token='different-session'; NPFind.refresh(); }); assert.equal(await page.locator('#np-find').count(), 0);
    await page.evaluate(() => { document.querySelector('#main').replaceChildren(); S.view='dash'; });
    const intercepted = await page.evaluate(() => !document.dispatchEvent(new KeyboardEvent('keydown', {key:'f',ctrlKey:true,bubbles:true,cancelable:true})));
    assert.equal(intercepted, false); assert.deepEqual(errors, []);
    console.log('PASS: mock browser editor save/undo/busy/permission/timestamps, real R + permission hooks, scoped find/worker/keyboard/rerender/5 MiB/session cleanup, outside browser find');
  } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode=1; }).finally(() => server.close());
