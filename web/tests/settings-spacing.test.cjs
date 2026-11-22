const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const root = path.join(__dirname, '../static');
const app = fs.readFileSync(path.join(root, 'js/app.js'), 'utf8');
const index = fs.readFileSync(path.join(root, 'index.html'), 'utf8');
function section(start, end) {
  const a = app.indexOf(start), b = app.indexOf(end, a);
  assert.ok(a >= 0 && b > a, `Missing source section: ${start}`);
  return app.slice(a, b);
}
const organizer = section('function organizeServerSettings(', 'function displayNodeLocation(');
const dropdowns = section('function closeAllCustomSelects()', "document.addEventListener('click',");
const styles = [...index.matchAll(/href="\/css\/([^"]+)"/g)].map(m => fs.readFileSync(path.join(root, 'css', m[1]), 'utf8')).join('\n');
function fixture(type = 'paper') {
  const srv = {id:'demo', node_id:'local', name:'Demo server', owner_id:1, type, version:type === 'velocity' ? '3.4.0-SNAPSHOT' : '1.21.11', port:25565, memory_max:2048, jar_file:type === 'steel' ? 'steel' : 'server.jar'};
  const container = {innerHTML:''};
  const c = {S:{props:{}, user:{id:1, role:'admin'}, activeServerId:'demo', serverSettingsSection:1},
    curServer:() => srv, findServerRef:() => srv, $:() => container,
    esc:value => String(value ?? '').replace(/[&<>"']/g, ch => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[ch])),
    ic:() => '', engineIcon:() => '', srvh:() => '', getBrand:() => '',
    resolveJavaSelectValue:() => '21', getServerMaxRamMB:() => 2048, getAllMcVersions:() => [srv.version],
    getNodeById:() => null, _settingsUiState:{}, getRecommendedJavaForMcVersion:() => '21',
    renderJavaSelectOptions:() => '<option value="21">Java 21</option><option value="25">Java 25</option>',
    renderJavaAvailabilityStrip:() => '<p class="mu">Java 21 Ready</p>', renderMinecraftTextHtml:() => '',
    onReinstallEngineChange() {}, loadVersionsCatalog() {}, setTimeout() {}, enhanceSelects() {}};
  c.window = c;
  vm.runInNewContext(section('const ENGINE_ICONS =', 'function onPlayerHeadError('), c);
  c.S.token = 'fixture'; c.hasServerPermission = () => true;
  vm.runInNewContext(fs.readFileSync(path.join(root, 'js/server-artifact.js'), 'utf8'), c);
  vm.runInNewContext(fs.readFileSync(path.join(root, 'js/server-automation.js'), 'utf8'), c);
  vm.runInNewContext(fs.readFileSync(path.join(root, 'js/proxy-settings.js'), 'utf8'), c);
  vm.runInNewContext(`var settingsView = {${section('  set() {', '\n  },')}\n}};`, c);
  vm.runInNewContext(section('function openReinstallModal(', 'async function submitReinstall('), c);
  return {settings:() => c.settingsView.set(), reinstall:() => {c.openReinstallModal('demo');return container.innerHTML;}};
}
async function withBrowser(check) {
  const {chromium} = require(process.env.NP_LAYOUT_PLAYWRIGHT);
  const browser = await chromium.launch({headless:true, ...(process.env.NP_LAYOUT_CHROME ? {executablePath:process.env.NP_LAYOUT_CHROME} : {})});
  try {
    const page = await browser.newPage();
    await page.route('**/*', route => route.request().resourceType() === 'image' ? route.fulfill({contentType:'image/png', body:Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9Zl1sAAAAASUVORK5CYII=', 'base64')}) : route.abort());
    await page.emulateMedia({reducedMotion:'reduce'});
    await check(page);
  } finally { await browser.close(); }
}
async function load(page, html, theme) {
  await page.setContent(`<html data-theme="${theme}"><head><base href="http://layout.example.invalid/"><style>${styles}</style></head><body><main id="mount" style="padding:24px">${html}</main></body></html>`);
}

test('reinstall checkbox rows opt out of wrapping; startup layout is explicit', () => {
  const html = fixture().reinstall();
  for (const id of ['reinst-world', 'reinst-config', 'reinst-plugins']) {
    assert.match(html, new RegExp(`<label[^>]*class="[^"]*np-check-row[^>]*>\\s*<input[^>]*id="${id}"`));
  }
  assert.doesNotMatch(html, /id="reinst-sync-java"/);
  assert.match(html, /Your Java selection is preserved/);
  assert.match(organizer, /np-startup-grid/);
});

test('browser: reinstall checkboxes stay beside multiline text at every width', {skip:!process.env.NP_LAYOUT_PLAYWRIGHT}, async () => {
  await withBrowser(async page => {
    for (const theme of ['dark','light']) for (const width of [1440,1024,768,390,320]) {
      await page.setViewportSize({width,height:1000});
      await load(page, fixture().reinstall(), theme);
      await page.addScriptTag({content:dropdowns + '\nenhanceSelects(document);'});
      const rows = await page.evaluate(() => ['reinst-world','reinst-config','reinst-plugins'].map(id => {
        const input = document.getElementById(id), text = input.nextElementSibling;
        const a = input.getBoundingClientRect(), b = text.getBoundingClientRect();
        return {id, input:{x:a.x,y:a.y,right:a.right,bottom:a.bottom,width:a.width}, text:{x:b.x,y:b.y}, wrap:getComputedStyle(input.closest('label')).flexWrap};
      }));
      for (const row of rows) {
        const context = `${theme}/${width}: ${JSON.stringify(row)}`;
        assert.equal(row.wrap, 'nowrap', context);
        assert.ok(row.input.width >= 14, `checkbox shrank: ${context}`);
        assert.ok(row.text.x >= row.input.right + 6, `text not beside checkbox: ${context}`);
        assert.ok(row.text.y < row.input.bottom && row.input.y <= row.text.y + 5, `checkbox orphaned above caption: ${context}`);
      }
      const input = page.locator('#reinst-world');
      assert.equal(await input.isChecked(), true);
      await page.locator('label').filter({has:input}).locator('span').first().click();
      assert.equal(await input.isChecked(), false, 'caption remains clickable');
      assert.ok(await page.locator('.pbox').evaluate(el => el.scrollWidth <= el.clientWidth + 1));
    }
  });
});

test('browser: startup cards use compact columns and preserve mobile order and controls', {skip:!process.env.NP_LAYOUT_PLAYWRIGHT}, async () => {
  await withBrowser(async page => {
    for (const type of ['paper','velocity','steel']) for (const theme of ['dark','light']) for (const width of [1594,1024,900,390,320]) {
      await page.setViewportSize({width,height:1200});
      await load(page, fixture(type).settings(), theme);
      await page.addScriptTag({content:`window.S={serverSettingsSection:1}; window.getBrand=()=>'';\n${organizer}\n${dropdowns}\norganizeServerSettings(document.getElementById('mount')); enhanceSelects(document);`});
      const geometry = await page.evaluate(() => {
        const rect = el => {if (!el) return null;const r=el.getBoundingClientRect();return {x:r.x,y:r.y,right:r.right,bottom:r.bottom,height:r.height};};
        const group = document.querySelector('[data-settings-panel="1"]');
        return {java:rect(group.querySelector('#prop-java')?.closest('.card')), resources:rect(group.querySelector('#prop-jar').closest('.card')), automation:rect(group.querySelector('#server-automation-settings')), overflow:group.scrollWidth>group.clientWidth+1};
      });
      const {java,resources,automation}=geometry, context=`${type}/${theme}/${width}: ${JSON.stringify(geometry)}`;
      assert.equal(geometry.overflow,false,context);
      if (width > 900 && java) {
        assert.ok(Math.abs(automation.y-java.bottom-16)<1, `blank space below Java card: ${context}`);
        assert.ok(Math.abs(java.x-automation.x)<1,context);
        assert.ok(Math.abs(java.y-resources.y)<1,context);
        assert.ok(resources.x>java.right,context);
      } else if (width > 900) {
        assert.ok(Math.abs(resources.y-automation.y)<1,context);
        assert.ok(automation.x>resources.right,context);
      } else {
        if (java) assert.ok(Math.abs(resources.y-java.bottom-16)<1,context);
        assert.ok(Math.abs(automation.y-resources.bottom-16)<1,context);
      }
      await page.evaluate(() => {window.originalJar=document.getElementById('prop-jar');originalJar.value='unchanged-by-tabs.jar';});
      await page.locator('[data-settings-section="0"]').click();
      await page.locator('[data-settings-section="1"]').click();
      assert.ok(await page.evaluate(() => originalJar===document.getElementById('prop-jar')),'controls must not be cloned');
      assert.equal(await page.locator('#prop-jar').inputValue(),'unchanged-by-tabs.jar');
      assert.ok(await page.evaluate(() => document.documentElement.scrollWidth<=innerWidth+1),context);
    }
  });
});
