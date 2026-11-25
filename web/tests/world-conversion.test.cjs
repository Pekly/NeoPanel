const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const root = path.join(__dirname, '../static');
const source = fs.readFileSync(path.join(root, 'js/world-conversion.js'), 'utf8');
const css = fs.readFileSync(path.join(root, 'css/world-conversion.css'), 'utf8');
const permissions = ['files.view', 'files.edit', 'backup.manage'];
const job = extra => ({id:'job/one', server_id:'server', input:'world', format:'JAVA_1_21', state:'converting', can_replace:false, progress:24, ...extra});
const status = extra => ({enabled:true, supported:true, unavailable_reason:'', formats:[{id:'JAVA_1_21',edition:'java',version:'1.21'}, {id:'BEDROCK_R21',edition:'bedrock',version:'1.21'}], jobs:[], warning:'Lossy', chunker_version:'1.21.0', ...extra});
const flush = async () => { for (let i=0;i<20;i++) await Promise.resolve(); };
function harness(handler, granted = [...permissions]) {
  let server = {id:'server',node_id:'worker',name:'<unsafe>'}, sequence = 0, observer;
  const timers = new Map(), intervals = new Map(), calls = [], events = {}, nodes = {};
  const document = {body:{},hidden:false,activeElement:null,addEventListener:(k,v)=>events[k]=v,removeEventListener:k=>delete events[k]};
  function node() {
    const n = {value:'',checked:false,hidden:false,disabled:false,isConnected:true,textContent:'',dataset:{},children:[],listeners:{},style:{setProperty(){}}, selectors:{},
      focus(){document.activeElement=this;},removeAttribute(k){delete this[k];},addEventListener(k,v){this.listeners[k]=v;},getClientRects(){return this.hidden?[]:[{}];},matches(){return this.disabled;},
      appendChild(child){this.children.push(child);child.parent=this;},remove(){this.parent.children=this.parent.children.filter(c=>c!==this);this.isConnected=false;},
      querySelector(s){return this.selectors[s];},querySelectorAll(){return Object.values(this.selectors);},contains(n){return Object.values(this.selectors).includes(n);}};
    Object.defineProperty(n,'innerHTML',{get(){return this.html||'';},set(html){this.html=html;this.selectors={};for(const m of html.matchAll(/<([\w-]+)\b([^>]*)>/g)){
      const id=m[2].match(/\bid="wc-([\w-]+)"/), data=m[2].match(/\b(data-[\w-]+)(?:\s|$)/);if(!id&&!data&&m[1]!=='progress')continue;
      const child=node();child.id=id?'wc-'+id[1]:'';child.hidden=/\bhidden\b/.test(m[2]);child.disabled=/\bdisabled\b/.test(m[2]);this.selectors[id?'#wc-'+id[1]:data?'['+data[1]+']':'progress']=child;
      if(id)nodes[id[1]]=child;
    }}});return n;
  }
  let modal;
  const container = {html:'',set innerHTML(html){this.html=html;modal=node();modal.innerHTML=html;},get innerHTML(){return this.html;},querySelector:s=>s==='.wc-modal'?modal:nodes.overlay,replaceChildren(){if(modal)modal.isConnected=false;this.html='';}};
  document.querySelector=()=>container;document.createElement=node;
  const c={document, S:{view:'fi',token:'token',user:{id:3,role:'user'}}, location:{href:'http://fixture/files'},curServer:()=>server,hasServerPermission:p=>granted.includes(p),
    srvApiPath:(s,suffix='')=>`/api/nodes/${s.node_id}/proxy/api/servers/${s.id}${suffix}`,closeModal:()=>container.replaceChildren(),
    api:async(url,options)=>{calls.push({url,options});return handler?handler(url,options):url.endsWith('world-conversion')?status():{id:'server',stats:{status:'offline'}};},
    setTimeout:(fn,ms)=>{const id=++sequence;timers.set(id,{fn,ms});return id;},clearTimeout:id=>timers.delete(id),setInterval:(fn,ms)=>{const id=++sequence;intervals.set(id,{fn,ms});return id;},clearInterval:id=>intervals.delete(id),
    addEventListener:(k,v)=>events[k]=v,removeEventListener:k=>delete events[k],AbortController,MutationObserver:class {constructor(fn){observer=fn;}observe(){}disconnect(){observer=null;}},
    loadFileList:(...args)=>{c.opened=args;}};
  c.window=c;vm.runInNewContext(source,c);
  const h={c,nodes,calls,timers,intervals,events,granted,container,open:()=>c.NPWorldConversion.open(),close:()=>c.NPWorldConversion.reset(),setServer:s=>server=s,watch:()=>{for(const t of intervals.values())t.fn();},mutate:()=>observer?.(),get modal(){return modal;},get cards(){return nodes.jobs.children;},
    fill(){nodes.input.value='world';nodes.format.value='JAVA_1_21';nodes.loss.checked=true;nodes.input.listeners.input();},submit:()=>nodes.form.onsubmit({preventDefault(){}}),refresh:()=>nodes.refresh.onclick(),poll:async()=>{const t=[...timers.entries()].find(([,t])=>t.ms===2000);assert.ok(t);timers.delete(t[0]);await t[1].fn();},
    confirm(){nodes.backup.checked=true;nodes['replace-loss'].checked=true;nodes.confirm.value='REPLACE';nodes.confirm.listeners.input();},replace:()=>nodes['replace-form'].onsubmit({preventDefault(){}})};
  return h;
}
function fixture(data, mutate) {return async (url,o)=>o.method==='GET'?(url.endsWith('/world-conversion')?structuredClone(data):{id:'server',stats:{status:'offline'}}):mutate?mutate(url,o):job();}

test('Files-only view access; view-only users can monitor without mutation rights',async()=>{
  const denied=harness(undefined,[]);denied.open();assert.equal(denied.calls.length,0);
  const h=harness(undefined,['files.view']);h.c.S.view='console';h.open();assert.equal(h.calls.length,0);h.c.S.view='fi';h.open();await flush();h.fill();assert.equal(h.nodes.start.disabled,true);await h.submit();assert.equal(h.calls.length,2);
  assert.deepEqual(Object.keys(h.c.NPWorldConversion),['open','reset','dispose']);assert.doesNotMatch(h.container.innerHTML,/<unsafe>/);
});
test('exact contract, explicit output choice, stopped status preflight and duplicate submit protection',async()=>{
  let resolve;const h=harness(fixture(status(),()=>new Promise(r=>resolve=r)));h.open();await flush();
  assert.equal(h.nodes.format.value,'');h.nodes.input.value='world';h.nodes.loss.checked=true;await h.submit();assert.equal(h.calls.length,2);
  h.fill();const pending=h.submit();h.submit();await flush();assert.equal(h.calls.length,4);
  assert.equal(h.calls[2].url,'/api/nodes/worker/proxy/api/servers/server');
  assert.deepEqual(JSON.parse(h.calls[3].options.body),{input:'world',format:'JAVA_1_21',acknowledge_loss:true});
  resolve(job());await pending;assert.equal(h.nodes.start.disabled,true);assert.equal(h.nodes.loss.checked,false);assert.ok([...h.timers.values()].some(t=>t.ms===2000));
});
test('unsafe folders, unavailable support, disabled module and running servers fail closed',async()=>{
  for(const bad of ['../world','/host/world','C:\\world','.hidden','converted-output','a/b','wo\nrld']) {const h=harness();h.open();await flush();h.fill();h.nodes.input.value=bad;await h.submit();assert.equal(h.calls.length,2,bad);}
  for(const data of [status({enabled:false}),status({supported:false,unavailable_reason:'C:/private/path'})]) {const h=harness(fixture(data));h.open();await flush();h.fill();await h.submit();assert.equal(h.calls.length,2);assert.equal(h.nodes.start.disabled,true);assert.doesNotMatch(h.nodes.target.textContent,/private/);}
  let reads=0;const h=harness(async(url,o)=>url.endsWith('world-conversion')?status():{id:'server',stats:{status:++reads===1?'offline':'online'}});h.open();await flush();h.fill();await h.submit();assert.equal(h.calls.filter(c=>c.options.method==='POST').length,0);assert.match(h.nodes.error.textContent,/Stop the server/);
});
test('active jobs remain cancellable while disabled or unsupported, without backup permission',async()=>{
  const h=harness(fixture(status({enabled:false,supported:false,jobs:[job()]}),()=>({message:'Cancellation requested'})),['files.view','files.edit']);h.open();await flush();
  const cancel=h.cards[0].querySelector('[data-cancel]');assert.equal(cancel.hidden,false);await cancel.onclick();await cancel.onclick();
  const writes=h.calls.filter(c=>c.options.method==='DELETE');assert.equal(writes.length,1);assert.ok(writes[0].url.endsWith('/job%2Fone'));assert.equal(cancel.disabled,true);
});
test('all known terminal states suppress cancellation; unknown future phase stays active',async()=>{
  for(const state of ['succeeded','failed','cancelled','blocked','replaced','future_phase']) {const h=harness(fixture(status({jobs:[job({state})]})));h.open();await flush();assert.equal(h.cards[0].querySelector('[data-cancel]').hidden,state!=='future_phase',state);}
});
test('replacement requires backup.manage, Java, backend eligibility and second typed acknowledgement',async()=>{
  const complete=job({state:'succeeded',can_replace:true,replace_token:'one-time',output:'converted-world-123'});
  for(const [granted,j] of [[['files.view','files.edit'],complete],[permissions,{...complete,format:'BEDROCK_R21'}],[permissions,{...complete,can_replace:false}]]) {
    const h=harness(fixture(status({jobs:[j]})),granted);h.open();await flush();assert.equal(h.cards[0].querySelector('[data-replace]').hidden,true);await h.replace();assert.equal(h.calls.length,2);
  }
  const h=harness(fixture(status({jobs:[complete]}),()=>job({state:'replacing'})));h.open();await flush();h.cards[0].querySelector('[data-replace]').onclick();await h.replace();assert.equal(h.calls.length,2);h.confirm();await h.replace();
  const post=h.calls.find(c=>c.options.method==='POST');assert.ok(post.url.endsWith('/job%2Fone/replace'));assert.deepEqual(JSON.parse(post.options.body),{token:'one-time',confirm:'REPLACE',backup:true,acknowledge_loss:true});assert.equal(h.nodes['replace-form'].hidden,true);
});
test('replacement token and live permission changes invalidate already-open confirmation',async()=>{
  const data=status({jobs:[job({state:'succeeded',can_replace:true,replace_token:'old'})]});const h=harness(fixture(data));h.open();await flush();h.cards[0].querySelector('[data-replace]').onclick();h.confirm();data.jobs[0].replace_token='new';await h.refresh();await h.replace();assert.equal(h.calls.filter(c=>c.options.method==='POST').length,0);
  h.cards[0].querySelector('[data-replace]').onclick();h.confirm();h.granted.splice(h.granted.indexOf('backup.manage'),1);h.watch();await h.replace();assert.equal(h.nodes['replace-submit'].disabled,true);
});
test('unknown and malformed start acceptance never retries POST, even after reset/reopen',async()=>{
  for(const malformed of [false,true]) {const h=harness(fixture(status(),()=>{if(malformed)return {};throw Error('lost response');}));h.open();await flush();h.fill();await h.submit();assert.match(h.nodes.error.textContent,/unknown/);await h.poll();h.close();h.open();await flush();h.fill();await h.submit();assert.equal(h.nodes.start.disabled,true);assert.equal(h.calls.filter(c=>c.options.method==='POST').length,1);}
});
test('poll resolves uncertain start only from a new matching job; replacement uncertainty stays blocked',async()=>{
  const data=status();const h=harness(fixture(data,()=>{throw Error('lost');}));h.open();await flush();h.fill();await h.submit();data.jobs=[job({input:'other',state:'succeeded'})];await h.poll();assert.match(h.nodes.reason.textContent,/unconfirmed/);data.jobs=[job({id:'match',state:'succeeded'})];await h.poll();assert.doesNotMatch(h.nodes.reason.textContent,/unconfirmed/);
  const d=status({jobs:[job({state:'succeeded',can_replace:true,replace_token:'t'})]});const r=harness(fixture(d,()=>{throw Error('lost');}));r.open();await flush();r.cards[0].querySelector('[data-replace]').onclick();r.confirm();await r.replace();await r.poll();assert.match(r.nodes.reason.textContent,/unconfirmed/);d.jobs[0].state='replaced';d.jobs[0].replace_token='';await r.poll();assert.doesNotMatch(r.nodes.reason.textContent,/unconfirmed/);
});
test('HTTP rejection permits correction; proxy and host errors stay out of ordinary UI',async()=>{
  const h=harness(fixture(status(),()=>{throw Object.assign(Error('C:/private-host/file'),{status:409});}));h.open();await flush();h.fill();await h.submit();assert.equal(h.nodes.start.disabled,false);assert.doesNotMatch(h.nodes.error.textContent,/private-host/);
  const data=status({supported:false,unavailable_reason:'/private/systemd'});data.jobs=[job({state:'failed',error:'/private/failed',log:'<script>attack</script>/private/log',safety_copy:'/private/copy',backup:'/private/backup'})];
  for(const role of ['user','admin']) {const h=harness(fixture(data));h.c.S.user.role=role;h.open();await flush();const card=h.cards[0];if(role==='user'){assert.equal(card.querySelector('[data-log]'),undefined);assert.doesNotMatch(card.querySelector('[data-error]').textContent,/private/);}else {assert.match(card.querySelector('[data-log]').textContent,/<script>attack/);assert.match(h.nodes.technical.textContent,/private/);}}
});
test('malformed status and wrong server responses cannot enable mutations',async()=>{
  for(const data of [{},status({jobs:[job({server_id:'other'})]}),status({formats:[{}]}),status({jobs:[job(),job()]})]) {const h=harness(fixture(data));h.open();await flush();h.fill();assert.equal(h.nodes.start.disabled,true);assert.match(h.nodes.error.textContent,/Status could not/);}
});
test('close, navigation, node, account, role, permission and session changes abort pending status and reject stale responses',async()=>{
  for(const change of ['close','view','server','node','route','user','role','token','permission','removed']) {
    let resolve;const h=harness(()=>new Promise(r=>resolve=r));h.open();const signal=h.calls[0].options.signal;
    if(change==='close')h.close();if(change==='view')h.c.S.view='con';if(change==='server')h.setServer({id:'other',node_id:'worker'});if(change==='node')h.setServer({id:'server',node_id:'other'});if(change==='route')h.c.location.href+='?changed';if(change==='user')h.c.S.user={id:4,role:'user'};if(change==='role')h.c.S.user.role='admin';if(change==='token')h.c.S.token='new';if(change==='permission')h.granted.length=0;if(change==='removed'){h.modal.isConnected=false;h.mutate();}h.watch();
    assert.equal(signal.aborted,true,change);resolve(status());await flush();assert.equal(h.calls.length,1,change);assert.equal(h.timers.size,0,change);assert.equal(h.intervals.size,0,change);
  }
});
test('in-flight mutation target is immutable and closing retains uncertainty across reopen',async()=>{
  let resolve;const h=harness(fixture(status(),()=>new Promise(r=>resolve=r)));h.open();await flush();h.fill();const pending=h.submit();await flush();const post=h.calls.at(-1);h.close();assert.equal(post.options.signal.aborted,true);resolve(job());await pending;h.open();await flush();h.fill();await h.submit();assert.equal(h.calls.filter(c=>c.options.method==='POST').length,1);assert.ok(post.url.startsWith('/api/nodes/worker/'));
});
test('output opens through guarded existing Files API, unsafe host paths never become navigation',async()=>{
  const data=status({jobs:[job({state:'succeeded',output:'converted-world-123'})]});const h=harness(fixture(data));h.open();await flush();h.cards[0].querySelector('[data-open]').onclick();assert.equal(h.c.opened[0],'converted-world-123');assert.equal(h.c.opened[2].isCurrent(),true);h.c.S.token='changed';assert.equal(h.c.opened[2].isCurrent(),false);
  for(const output of ['/private/world','../world','C:\\world','converted-good/../bad']) {const h=harness(fixture(status({jobs:[job({state:'succeeded',output})]})));h.open();await flush();assert.equal(h.cards[0].querySelector('[data-open]').hidden,true);h.cards[0].querySelector('[data-open]').onclick();assert.equal(h.c.opened,undefined);}
});
test('polls are single bounded 2s timers and Escape disposes events',async()=>{
  const h=harness();h.open();await flush();for(let i=0;i<3;i++)await h.poll();assert.equal([...h.timers.values()].filter(t=>t.ms===2000).length,1);
  h.events.keydown({key:'Escape',preventDefault(){},stopImmediatePropagation(){}});assert.equal(h.timers.size,0);assert.equal(h.intervals.size,0);assert.equal(h.events.keydown,undefined);
  assert.match(css,/minmax\(0, 1fr\)/);assert.match(css,/:focus-visible/);assert.match(css,/prefers-reduced-motion/);assert.doesNotMatch(css,/transition:\s*all|overflow:\s*hidden/);
});

test('Chrome: real theme, mobile layout, focus trap, reduced motion and replacement interaction (demo API)',{skip:!process.env.NP_LAYOUT_PLAYWRIGHT},async()=>{
  const {chromium}=require(process.env.NP_LAYOUT_PLAYWRIGHT);
  const browser=await chromium.launch({headless:true,...(process.env.NP_LAYOUT_CHROME?{executablePath:process.env.NP_LAYOUT_CHROME}:{})});
  const index=fs.readFileSync(path.join(root,'index.html'),'utf8');
  const styles=[...index.matchAll(/href="\/css\/([^"]+)"/g)].map(m=>fs.readFileSync(path.join(root,'css',m[1]),'utf8')).join('\n')+'\n'+css;
  try {
    for(const width of [1440,390,320]) for(const theme of ['dark','light']) {
      const page=await browser.newPage({viewport:{width,height:900}});const errors=[];page.on('pageerror',e=>errors.push(e.message));
      await page.emulateMedia({reducedMotion:'reduce'});
      await page.setContent(`<html data-theme="${theme}"><head><style>${styles}</style></head><body><button id="trigger">Demo API — convert world</button><div id="modal-container"></div></body></html>`);
      await page.evaluate(data=>{
        window.S={view:'fi',token:'demo',user:{id:1,role:'admin'}};window.demoServer={id:'server',node_id:'local',name:'Demo world server'};window.demoStatus=data;window.posts=[];
        window.curServer=()=>demoServer;window.hasServerPermission=()=>true;window.srvApiPath=(s,suffix='')=>'/api/servers/'+s.id+suffix;
        window.closeModal=()=>document.querySelector('#modal-container').replaceChildren();window.loadFileList=path=>{window.opened=path;};
        window.api=async(url,o)=>{if(o.method==='POST'){posts.push(JSON.parse(o.body));return {...demoStatus.jobs[0],state:'replacing',can_replace:false,replace_token:''};}return url.endsWith('/world-conversion')?structuredClone(demoStatus):{id:'server',stats:{status:'offline'}};};
      },status({jobs:[job({state:'succeeded',can_replace:true,replace_token:'demo-confirm',input:'a'.repeat(180),output:'converted-'+ 'a'.repeat(180),log:'/private/path\n'+'x'.repeat(500)})]}));
      await page.addScriptTag({content:source});await page.evaluate(()=>{document.querySelector('#trigger').focus();NPWorldConversion.open();});
      await page.waitForFunction(()=>!document.querySelector('#wc-refresh').disabled);
      assert.equal(await page.locator('#wc-close').evaluate(e=>document.activeElement===e),true);
      await page.keyboard.press('Shift+Tab');assert.equal(await page.locator('.wc-modal summary').last().evaluate(e=>document.activeElement===e),true);await page.keyboard.press('Tab');assert.equal(await page.locator('#wc-close').evaluate(e=>document.activeElement===e),true);
      assert.equal(await page.locator('#wc-close').evaluate(e=>getComputedStyle(e).outlineStyle),'solid');
      await page.locator('[data-replace]').click();await page.locator('#wc-backup').check();await page.locator('#wc-replace-loss').check();await page.locator('#wc-confirm').fill('REPLACE');assert.equal(await page.locator('#wc-replace-submit').isDisabled(),false);
      assert.ok(await page.locator('.wc-modal').evaluate(e=>e.scrollWidth<=e.clientWidth+1),`${theme}/${width} modal overflow`);
      assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),`${theme}/${width} page overflow`);
      assert.equal(await page.locator('.wc-job').evaluate(e=>getComputedStyle(e).animationName),'none');
      await page.locator('#wc-replace-submit').click();await page.waitForFunction(()=>posts.length===1);assert.deepEqual(await page.evaluate(()=>posts[0]),{token:'demo-confirm',confirm:'REPLACE',backup:true,acknowledge_loss:true});
      await page.keyboard.press('Escape');assert.equal(await page.locator('.wc-modal').count(),0);assert.equal(await page.locator('#trigger').evaluate(e=>document.activeElement===e),true);assert.deepEqual(errors,[]);
      await page.close();
    }
  } finally {await browser.close();}
});
test('replacement request keeps its approved token through a delayed status preflight',async()=>{
  let resolve, reads=0;const complete=job({state:'succeeded',can_replace:true,replace_token:'approved'});
  const h=harness(async(url,o)=>{if(o.method==='POST')return job({state:'replacing'});if(url.endsWith('world-conversion'))return status({jobs:[complete]});if(++reads===2)return new Promise(r=>resolve=r);return {id:'server',stats:{status:'offline'}};});
  h.open();await flush();h.cards[0].querySelector('[data-replace]').onclick();h.confirm();const pending=h.replace();await flush();assert.equal(h.nodes['replace-back'].disabled,true);h.nodes['replace-back'].onclick();resolve({id:'server',stats:{status:'offline'}});await pending;
  assert.equal(JSON.parse(h.calls.find(c=>c.options.method==='POST').options.body).token,'approved');
});
