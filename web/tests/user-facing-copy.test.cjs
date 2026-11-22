const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const app = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
function section(start, end) {
  const from = app.indexOf(start), to = app.indexOf(end, from);
  assert.ok(from >= 0 && to > from);
  return app.slice(from, to);
}
const details = {host:'play.example.com',port:2022,username:'member.server',serverName:'Survival',serverId:'one',isRemote:true,listening:true,launchUri:'sftp://member.server@play.example.com:2022'};

test('SFTP users get usable connection details, not node-token instructions', () => {
  for (const role of ['user','admin']) {
    const c = {S:{user:{role}},getSftpDetails:()=>details,ic:()=>'',esc:String};
    vm.runInNewContext(section('function renderSftpCardHtml(', '// --- File Manager Sorting'), c);
    const html = c.renderSftpCardHtml({id:'one'});
    assert.match(html,/play.example.com:2022/);assert.match(html,/member.server/);
    assert.match(html,/NeoPanel account password/);assert.match(html,/SSH key/);
    assert.doesNotMatch(html,/Pure-Go|jailed SFTP daemon|Management Endpoint/);
    if (role === 'user') assert.doesNotMatch(html,/node_token|sudo neopanel|Remote nodes/);
    else assert.match(html,/<details><summary>Administrator connection options/);
  }
});

test('FileZilla exports keep connection fields without infrastructure authentication advice', () => {
  let xml = '';
  const c = {
    S:{user:{role:'user'}},curServer:()=>({id:'one'}),getSftpDetails:()=>details,
    Blob:class {constructor(parts){xml=parts.join('');}},
    URL:{createObjectURL:()=> 'blob:fixture',revokeObjectURL(){}},
    document:{createElement:()=>({click(){},remove(){}}),body:{appendChild(){}}},
    shortServerId:()=> 'one',setTimeout(){},toast(){}
  };
  vm.runInNewContext(section('function downloadFileZillaProfile(', 'function renderSftpCardHtml('),c);
  c.downloadFileZillaProfile();
  assert.match(xml,/<Host>play.example.com<\/Host>/);assert.match(xml,/<User>member.server<\/User>/);
  assert.match(xml,/NeoPanel account password/);assert.doesNotMatch(xml,/remote node token|node_token|local account password/i);
});

test('file destination labels and managed settings use user-facing terms', () => {
  assert.doesNotMatch(app,/\/home\/container\//);
  assert.match(app,/Server files \//);
  assert.doesNotMatch(app,/Locked by Node Policy|restricted by node policy\./);
  assert.match(app,/Managed by your host/);
});

test('browser: user and admin import dialogs stay concise and fit desktop/mobile', {skip: !process.env.NP_LAYOUT_PLAYWRIGHT}, async () => {
  const {chromium} = require(process.env.NP_LAYOUT_PLAYWRIGHT);
  const browser = await chromium.launch({headless:true, ...(process.env.NP_LAYOUT_CHROME ? {executablePath:process.env.NP_LAYOUT_CHROME} : {})});
  const root = path.join(__dirname,'../static');
  const index = fs.readFileSync(path.join(root,'index.html'),'utf8');
  const styles = [...index.matchAll(/href="\/css\/([^"]+)"/g)].map(m=>fs.readFileSync(path.join(root,'css',m[1]),'utf8')).join('\n');
  const importer = fs.readFileSync(path.join(root,'js/sftp-import.js'),'utf8');
  const dropdowns = section('function closeAllCustomSelects()', "document.addEventListener('click',");
  try {
    for (const role of ['user','admin']) for (const width of [1440,390,320]) {
      const page = await browser.newPage({viewport:{width,height:900}});
      await page.emulateMedia({reducedMotion:'reduce'});
      await page.setContent(`<html data-theme="dark"><head><style>${styles}</style></head><body><div id="modal-container"></div></body></html>`);
      await page.evaluate(role=>{
        window.S={view:'fi',token:'fixture',user:{id:1,role}};
        window.curServer=()=>({id:'demo',name:'Demo server',node_id:'private-worker'});
        window.hasServerPermission=()=>true;
        window.srvApiPath=()=>'/fixture';
        window.closeModal=()=>document.querySelector('#modal-container').replaceChildren();
        window.api=async()=>({eligible:true,reason:'',limits:{max_bytes:10737418240,default_bytes:10737418240,max_entries:100000,max_active:2},job:{id:'fixture',server_id:'demo',state:'failed',files_total:0,files_done:0,bytes_total:0,bytes_done:0,current_path:'',error:'C:/private-host/staging failed',safety_copy:'.restore-demo/original'}});
      },role);
      await page.addScriptTag({content:importer});
      await page.evaluate(()=>NPSFTPImport.open());
      await page.waitForFunction(()=>!document.querySelector('#sfi-refresh').disabled);
      await page.addScriptTag({content:dropdowns+'\nenhanceSelects(document);'});
      const modal=page.locator('.sfi-modal');
      assert.ok(await modal.evaluate(el=>el.scrollWidth<=el.clientWidth+1),`${role}/${width}: overflow`);
      assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1));
      const text=await modal.innerText();
      assert.match(text,/empty, stopped server/);assert.match(text,/Compare this fingerprint.*provider/);assert.match(text,/Stop the source server/);
      if(role==='user') assert.doesNotMatch(text,/private-worker|private-host|\.restore-|staging|allowlist|retention/);
      else {
        const details=modal.locator('details').filter({has:page.locator('summary',{hasText:'Operator details'})});
        assert.equal(await details.getAttribute('open'),null);
        await details.locator('summary').click();
        assert.match(await details.innerText(),/private-worker/);
      }
      await page.close();
    }
  } finally {await browser.close();}
});
