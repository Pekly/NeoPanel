const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const root = path.join(__dirname, '../static');
const source = fs.readFileSync(path.join(root, 'js/features.js'), 'utf8');
const keys = ['backups','schedules','sftp','sftp_import','plugins','databases','players','proxy_settings','extra_ports','software','subdomains','countdowns','world_conversion'];
const view = (overrides={}) => ({policy:{...overrides},global:{},effective:Object.fromEntries(keys.map(k=>[k,overrides[k]!==false]))});
const tick = () => new Promise(r=>setImmediate(r));
function harness() {
 const c = {S:{view:'set',user:{id:1,role:'admin'}},srv:{id:'a'},esc:String,requests:[],renders:0,document:{getElementById:()=>null,querySelectorAll:()=>[]}};
 c.window=c;c.curServer=()=>c.srv;c.srvApiPath=(s,p)=>'/api/servers/'+s.id+p;c.R=()=>{c.renders++};
 c.api=(url,opts)=>new Promise(resolve=>c.requests.push({url,opts,resolve}));
 vm.runInNewContext(source,c);return c;
}
test('features default on, apply valid policy, and ignore a late server response',async()=>{
 const c=harness();c.NPFeatures.sync();assert.equal(c.requests[0].url,'/api/servers/a/features');
 c.srv={id:'b'};c.NPFeatures.sync();c.requests[0].resolve(view({plugins:false}));await tick();
 assert.equal(c.renders,0);assert.equal(c.NPFeatures.pageEnabled('plug'),true);
 c.requests[1].resolve(view({plugins:false,schedules:false}));await tick();
 assert.equal(c.NPFeatures.pageEnabled('plug'),false);assert.equal(c.NPFeatures.pageEnabled('sc'),false);assert.equal(c.NPFeatures.pageEnabled('bk'),true);
 c.NPFeatures.reset();assert.equal(c.NPFeatures.pageEnabled('plug'),true);
});
test('only administrators render feature editing cards',()=>{
 const c=harness();assert.match(c.NPFeatures.render(c.srv),/Optional features/);
 c.S.user.role='user';assert.equal(c.NPFeatures.render(c.srv),'');
});
test('feature cards render responsive controls and explicit safe save', {skip:!process.env.NP_LAYOUT_PLAYWRIGHT},async()=>{
 const {chromium}=require(process.env.NP_LAYOUT_PLAYWRIGHT);
 const browser=await chromium.launch({headless:true,executablePath:process.env.NP_LAYOUT_CHROME});
 try {
  const page=await browser.newPage();
  const style=fs.readFileSync(path.join(root,'css/style.css'),'utf8')+fs.readFileSync(path.join(root,'css/features.css'),'utf8');
  await page.setContent(`<style>${style}</style><main id="main"></main>`);
  await page.evaluate(({keys})=>{
   window.S={view:'admin',user:{id:1,role:'admin'}};window.curServer=()=>null;window.esc=s=>String(s).replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
   window.saved=[];window.api=async(url,opts)=>{const p=opts?.body?JSON.parse(opts.body):{};if(opts?.method==='PUT')window.saved.push({url,p});return {policy:p,global:p,effective:Object.fromEntries(keys.map(k=>[k,p[k]!==false])),nodes:{offline:false}}};
   window.R=()=>{document.querySelector('#main').innerHTML=NPFeatures.render(null);NPFeatures.mount()};
  },{keys});
  await page.addScriptTag({content:source});
  await page.evaluate(()=>{R();NPFeatures.sync()});
  await page.waitForSelector('[data-feature-key="plugins"]');
  assert.equal(await page.locator('[data-feature-key]').count(),13);
  await page.locator('[data-feature-key="plugins"]').selectOption('false');
  assert.equal(await page.evaluate(()=>saved.length),0);
  await page.locator('#np-feature-save').click();
  await page.waitForFunction(()=>saved.length===1);
  assert.equal(await page.evaluate(()=>saved[0].p.plugins),false);
  assert.match(await page.locator('[role="status"]').innerText(),/pending/);
  for(const width of [1280,390]) {await page.setViewportSize({width,height:900});assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));}
  await page.evaluate(()=>{S.view='set';window.srv={id:'alpha'};curServer=()=>srv;srvApiPath=(s,p)=>'/api/servers/'+s.id+p;R=()=>{document.querySelector('#main').innerHTML=NPFeatures.render(srv);NPFeatures.mount()};R();NPFeatures.sync()});
  await page.waitForSelector('[data-feature-key="plugins"] option[value="inherit"]', {state:'attached'});
  assert.equal(await page.locator('[data-feature-key="plugins"]').inputValue(),'inherit');
 } finally {await browser.close()}
});
