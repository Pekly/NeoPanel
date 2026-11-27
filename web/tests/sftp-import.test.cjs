const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/sftp-import.js'), 'utf8');
const app = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
const fingerprint = 'SHA256:' + 'A'.repeat(43);
const status = extra => ({job:null, eligible:true, reason:'', limits:{max_bytes:20*1073741824, default_bytes:10*1073741824, max_entries:100000, max_active:2}, ...extra});
const job = extra => ({id:'job/one', server_id:'server', state:'queued', files_total:2, files_done:0, bytes_total:1000, bytes_done:0, current_path:'', created_at:'2026-01-01', updated_at:'2026-01-01', message:'', error:'', safety_copy:'', ...extra});
const preview = extra => ({remote_path:'.', entries:[{name:'world',path:'world',is_dir:true,size:0}], truncated:false, ...extra});
const flush = async () => { for(let i=0;i<12;i++) await Promise.resolve(); };
function harness(handler = async () => status(), granted = ['files.view','files.edit','sftp.access']) {
  let server = {id:'server',node_id:'worker',name:'<unsafe>'}, html = '', root;
  const nodes = {}, timers = new Map(), intervals = new Map(), calls = [], events = {}, selected = [], excluded = [];
  let sequence = 0, observer;
  const document = {body:{}, hidden:false, activeElement:null, addEventListener:(type, fn)=>events[type]=fn, removeEventListener:type=>delete events[type]};
  const makeNode = id => ({id:'sfi-'+id, value:'', checked:false, disabled:false, hidden:false, innerHTML:'', textContent:'', dataset:{}, listeners:{}, isConnected:true,
    addEventListener(type,fn){this.listeners[type]=fn;}, focus(){document.activeElement=this;}, removeAttribute(key){delete this[key];}, getClientRects(){return this.hidden?[]:[{}];}, matches(){return this.disabled;}, scrollIntoView(){}});
  const container = {
    set innerHTML(value) {
      html=value;
      if (!value) { if(root)root.isConnected=false; return; }
      for(const match of value.matchAll(/<[^>]+id="sfi-([\w-]+)"[^>]*>/g)) {
        const node = nodes[match[1]] = makeNode(match[1]);
        node.value = match[0].match(/\bvalue="([^"]*)"/)?.[1] || '';
        node.checked = /\bchecked\b/.test(match[0]); node.disabled = /\bdisabled\b/.test(match[0]);
      }
      nodes.auth.value='password';
      root={isConnected:true, querySelector:s=>nodes[s.replace('#sfi-','')], querySelectorAll:s=>s==='[data-include]:checked'?selected:s==='[data-exclude]:checked'?excluded:Object.values(nodes), contains:n=>Object.values(nodes).includes(n)};
    }, get innerHTML(){return html;}, querySelector:s=>s==='.sfi-modal'?root:nodes.overlay,
    replaceChildren(){this.innerHTML='';}
  };
  document.querySelector=()=>container;
  const c={S:{view:'fi',token:'token',user:{id:3,role:'user'}}, document, location:{href:'http://panel/server/worker/server/files'},
    curServer:()=>server, hasServerPermission:p=>granted.includes(p), srvApiPath:(srv,suffix)=>`/api/nodes/${srv.node_id}/proxy/api/servers/${srv.id}${suffix}`,
    closeModal:()=>container.replaceChildren(), api:async(url,options)=>{calls.push({url,options});return handler(url,options);},
    setTimeout:(fn,ms)=>{const id=++sequence;timers.set(id,{fn,ms});return id;}, clearTimeout:id=>timers.delete(id), setInterval:(fn,ms)=>{const id=++sequence;intervals.set(id,{fn,ms});return id;}, clearInterval:id=>intervals.delete(id),
    addEventListener:(name,fn)=>events[name]=fn, removeEventListener:name=>delete events[name],
    AbortController, TextEncoder, MutationObserver:class {constructor(fn){observer=fn;}observe(){}disconnect(){observer=null;}}
  };c.window=c;vm.runInNewContext(source,c);
  return {c,nodes,calls,timers,intervals,selected,excluded,container,open:()=>c.NPSFTPImport.open(),setServer:s=>server=s,watch:()=>{for(const {fn} of intervals.values())fn();},mutate:()=>observer?.(),get root(){return root;},
    fill(){Object.assign(nodes.host,{value:'source.example'});nodes.port.value='22';nodes.username.value='owner';nodes.password.value='very-secret';nodes.fingerprint.value=fingerprint;nodes.confirm.checked=true;nodes.confirm.onchange();},
    submit(){nodes.form.onsubmit({preventDefault(){}});}, close(){nodes.close.onclick();}, input(id,value){nodes[id].value=value;nodes[id].listeners.input?.();}, events};
}

test('entry is Files-only; all three permissions are required, ordinary team users are accepted', async () => {
  for(const missing of ['files.view','files.edit','sftp.access']) {const h=harness(undefined,['files.view','files.edit','sftp.access'].filter(p=>p!==missing));h.open();assert.equal(h.calls.length,0);}
  const h=harness();h.c.S.view='sftp';h.open();assert.equal(h.calls.length,0);h.c.S.view='fi';h.open();await flush();assert.equal(h.calls.length,1);
  assert.match(app, /\['files.view', 'files.edit', 'sftp.access'\]\.every\(p => hasServerPermission\(p\)\)/);
  assert.match(app,/uploadFileDialog\(\)[\s\S]{0,450}NPSFTPImport.open\(\)/);
});

test('client guard permits cancellation without files.delete and requires all import permissions', () => {
  const permissions=new Set(['files.view','files.edit','sftp.access']);
  const c={S:{user:{role:'user'}},hasServerPermission:p=>permissions.has(p)};
  vm.runInNewContext(app.slice(app.indexOf('function clientRequestPermission('),app.indexOf('function applyPermissionControls(')),c);
  const url='/api/nodes/worker/proxy/api/servers/server/files/import-sftp/job%2Fone';
  assert.doesNotThrow(()=>c.assertClientRequestAccess(url,{method:'DELETE'}));permissions.delete('sftp.access');assert.throws(()=>c.assertClientRequestAccess(url,{method:'DELETE'}),/sftp.access/);
});

test('status uses captured node, masks secrets, and never stores credentials', async () => {
  const h=harness();h.open();await flush();assert.equal(h.calls[0].url,'/api/nodes/worker/proxy/api/servers/server/files/import-sftp');
  assert.doesNotMatch(source,/localStorage|sessionStorage|indexedDB|document.cookie/);
  assert.match(h.container.innerHTML,/id="sfi-password" type="password"/);assert.match(h.container.innerHTML,/id="sfi-passphrase" type="password"/);
  assert.doesNotMatch(h.container.innerHTML,/<unsafe>/);assert.deepEqual(Object.keys(h.c.NPSFTPImport),['open']);
});

test('invalid status and old nodes fail closed with actionable feedback', async () => {
  for(const response of [{},status({job:{state:'completed'}}),status({limits:{max_bytes:-1}})]) {const h=harness(async()=>response);h.open();await flush();assert.equal(h.nodes.start.disabled,true);assert.match(h.nodes.error.textContent,/Unexpected/);}
  for(const code of [404,405]) {const h=harness(async()=>{throw Object.assign(Error('old'),{status:code});});h.open();await flush();assert.match(h.nodes.target.textContent,/Contact support/);assert.doesNotMatch(h.nodes.target.textContent,/node|worker|update/i);}
});

test('ineligible destination blocks Start, not a verified connection test', async () => {
  const h=harness(async(url)=>url.endsWith('/test')?preview():status({eligible:false,reason:'Stop the destination and empty its root.'}));h.open();await flush();h.fill();
  assert.equal(h.nodes.start.disabled,true);assert.equal(h.nodes.test.disabled,false);h.submit();await flush();assert.equal(h.calls.length,1);await h.nodes.test.onclick();assert.equal(h.calls.length,2);
});

test('host-key fetch sends no credentials, requires independent confirmation, resets on edits', async () => {
  const h=harness(async(url)=>url.endsWith('/host-key')?{fingerprint,algorithm:'ssh-ed25519'}:status());h.open();await flush();h.fill();await h.nodes.fetch.onclick();
  assert.deepEqual(JSON.parse(h.calls[1].options.body),{host:'source.example',port:22});assert.equal(h.nodes.confirm.checked,false);assert.equal(h.nodes.test.disabled,true);assert.match(h.nodes['key-note'].textContent,/Unverified/);
  for(const id of ['host','port','fingerprint']) {h.fill();h.input(id,h.nodes[id].value);assert.equal(h.nodes.confirm.checked,false);assert.equal(h.nodes.start.disabled,true);}
});

test('password test serializes exact safe options, escapes remote names and retains form secrets', async () => {
  const name='<img onerror="attack">';
  const h=harness(async(url)=>url.endsWith('/test')?preview({remote_path:'<b>root</b>',entries:[{name,path:name,is_dir:false,size:3}],truncated:true}):status());h.open();await flush();h.fill();
  h.nodes.max.value='1.5';h.nodes.bandwidth.value='2.5';h.nodes.excludes.value='logs\nbackup dir';h.excluded.push({dataset:{exclude:'logs'}});h.selected.push({dataset:{include:'world'}});
  await h.nodes.test.onclick();
  assert.deepEqual(JSON.parse(h.calls[1].options.body),{host:'source.example',port:22,username:'owner',password:'very-secret',fingerprint,remote_path:'.',include_paths:['world'],exclude_paths:['logs','backup dir'],skip_hidden:false,preserve_times:true,max_bytes:1610612736,bandwidth_kib:2560,timeout_minutes:60});
  assert.doesNotMatch(h.nodes.preview.innerHTML,/<img|<b>root/);assert.match(h.nodes.preview.innerHTML,/&lt;img/);assert.match(h.nodes.preview.innerHTML,/truncated/);assert.equal(h.nodes.password.value,'very-secret');
});

test('key mode sends only private_key and passphrase, enforces 64 KiB chooser limit', async () => {
  const h=harness(async(url)=>url.endsWith('/test')?preview():status());h.open();await flush();h.fill();h.nodes.auth.value='key';h.nodes.auth.onchange();assert.equal(h.nodes.password.value,'');
  h.nodes['private-key'].value='PRIVATE KEY';h.nodes.passphrase.value='encrypted';await h.nodes.test.onclick();let body=JSON.parse(h.calls[1].options.body);assert.equal(body.private_key,'PRIVATE KEY');assert.equal(body.passphrase,'encrypted');assert.equal('password' in body,false);
  let reads=0;h.nodes['key-file'].files=[{size:65537,text:async()=>{reads++;return 'bad';}}];await h.nodes['key-file'].onchange();assert.equal(reads,0);assert.match(h.nodes.error.textContent,/64 KiB/);
  h.nodes['key-file'].files=[{size:20,text:async()=>'loaded-key'}];await h.nodes['key-file'].onchange();assert.equal(h.nodes['private-key'].value,'loaded-key');
});

test('invalid values and invalid preview cannot be reported as success', async () => {
  const h=harness(async(url)=>url.endsWith('/test')?{}:status());h.open();await flush();h.fill();
  for(const [id,value] of [['port','0'],['max','21'],['bandwidth','1025'],['timeout','241'],['excludes','../world']]) {const previous=h.nodes[id].value;h.nodes[id].value=value;await h.nodes.test.onclick();assert.equal(h.calls.length,1);assert.ok(h.nodes.error.textContent);h.nodes[id].value=previous;}
  await h.nodes.test.onclick();assert.match(h.nodes.error.textContent,/Unexpected remote preview/);assert.doesNotMatch(h.nodes.status.textContent,/verified\. Preview ready/);
});

test('Start spamming is guarded; accepted jobs clear secrets and cancellation pins encoded job ID', async () => {
  let resolve;const h=harness(async(url,o)=>o.method==='GET'?status():o.method==='DELETE'?job({message:'Cancellation requested'}):new Promise(r=>resolve=r));h.open();await flush();h.fill();h.submit();h.submit();assert.equal(h.calls.length,2);
  resolve(job());await flush();assert.equal(h.nodes.password.value,'');assert.equal(h.nodes.start.disabled,true);assert.ok([...h.timers.values()].some(t=>t.ms===1500));
  await h.nodes.cancel.onclick();assert.equal(h.calls[2].url,'/api/nodes/worker/proxy/api/servers/server/files/import-sftp/job%2Fone');assert.equal(h.calls[2].options.method,'DELETE');assert.equal(h.nodes.cancel.disabled,true);assert.equal(h.nodes['job-state'].textContent,'Waiting');
});

test('unknown Start and malformed acceptance never auto-retry POST; GET resolves a newly found job', async () => {
  for(const malformed of [true,false]) {
    let reads=0;const h=harness(async(_,o)=>{if(o.method==='GET')return ++reads===1?status():status({job:job()});if(malformed)return {};throw Error('network timeout');});h.open();await flush();h.fill();h.submit();await flush();
    assert.equal(h.calls.filter(c=>c.options.method==='POST').length,1);assert.equal(h.calls.filter(c=>c.options.method==='GET').length,2);assert.match(h.nodes.status.textContent,/found after/);assert.equal(h.nodes.password.value,'');
  }
});

test('unresolved Start stays disabled even if GET reports no job', async () => {
  const h=harness(async(_,o)=>{if(o.method==='POST')throw Error('timeout');return status();});h.open();await flush();h.fill();h.submit();await flush();h.submit();await flush();assert.equal(h.calls.filter(c=>c.options.method==='POST').length,1);assert.equal(h.nodes.start.disabled,true);assert.match(h.nodes['start-reason'].textContent,/unknown/);
});

test('wrong cancellation generation is rejected, not substituted', async () => {
  const h=harness(async(_,o)=>o.method==='GET'?status({job:job()}):job({id:'replacement'}));h.open();await flush();await h.nodes.cancel.onclick();assert.match(h.nodes.error.textContent,/cancellation job ID/);assert.doesNotMatch(h.nodes.status.textContent,/Cancellation requested\./);
});

test('close, navigation, server/node/session/user changes abort requests and reject stale responses', async () => {
  for(const change of ['close','server','node','session','logout','user','view','route','removed']) {
    let resolve;const h=harness(()=>new Promise(r=>resolve=r));h.open();h.nodes.password.value='secret';const signal=h.calls[0].options.signal;
    if(change==='close')h.close();if(change==='server')h.setServer({id:'other',node_id:'worker'});if(change==='node')h.setServer({id:'server',node_id:'other'});if(change==='session')h.c.S.token='new';if(change==='logout')h.c.S.token='';if(change==='user')h.c.S.user={id:4};if(change==='view')h.c.S.view='con';if(change==='route')h.c.location.href+='?changed';if(change==='removed'){h.root.isConnected=false;h.mutate();}
    h.watch();assert.equal(signal.aborted,true,change);assert.equal(h.nodes.password.value,'',change);assert.equal(h.intervals.size,0);resolve(status());await flush();assert.equal(h.nodes.start.disabled,true);assert.equal(h.timers.size,0);assert.equal(h.calls.length,1);
  }
});

test('switching tabs preserves open modal and form state; reopened completed status never restarts or polls', async () => {
  const h=harness(async()=>status({job:job({state:'completed',files_done:2,bytes_done:1000})}));h.open();await flush();h.nodes.password.value='secret';h.c.document.hidden=true;h.watch();assert.equal(h.nodes.password.value,'secret');assert.ok(h.container.innerHTML!=='');
  h.c.document.hidden=false;h.open();await flush();assert.equal(h.nodes['job-state'].textContent,'Import complete');assert.equal(h.calls.filter(c=>c.options.method==='POST').length,0);assert.equal([...h.timers.values()].filter(t=>t.ms===1500).length,0);
});

test('native modal layout is responsive, focus-visible, reduced-motion aware and independently loaded', () => {
  const css=fs.readFileSync(path.join(__dirname,'../static/css/sftp-import.css'),'utf8');const index=fs.readFileSync(path.join(__dirname,'../static/index.html'),'utf8');
  assert.match(css,/repeat\(2, minmax\(0, 1fr\)\)/);assert.match(css,/@media \(max-width: 700px\)/);assert.match(css,/:focus-visible/);assert.match(css,/prefers-reduced-motion/);assert.doesNotMatch(css,/transition:\s*all|overflow:\s*hidden/);
  assert.ok(index.indexOf('/js/sftp-import.js')<index.indexOf('/js/app.js'));assert.match(index,/css\/sftp-import.css/);
  assert.match(source,/aria-modal="true"/);assert.match(source,/event.key === 'Escape'/);assert.match(source,/event.shiftKey/);
});

test('late local key reads cannot repopulate cleared secrets after Start or close', async () => {
  for (const finish of ['start','close']) {
    let resolve;
    const h=harness(async(_,o)=>o.method==='GET'?status():job());h.open();await flush();h.fill();h.nodes.auth.value='key';h.nodes.auth.onchange();h.nodes['private-key'].value='current-key';
    h.nodes['key-file'].files=[{size:10,text:()=>new Promise(r=>resolve=r)}];const pending=h.nodes['key-file'].onchange();
    if(finish==='start'){h.submit();await flush();}else h.close();
    resolve('late-secret');await pending;assert.equal(h.nodes['private-key'].value,'');
  }
});

test('explicit rejected Start preserves editable settings and shows the backend reason', async () => {
  const h=harness(async(_,o)=>{if(o.method==='POST')throw Object.assign(Error('Destination is no longer empty'),{status:409});return status();});h.open();await flush();h.fill();h.submit();await flush();
  assert.match(h.nodes.error.textContent,/no longer empty/);assert.equal(h.nodes.settings.disabled,false);assert.equal(h.calls.length,2);
});

test('ordinary import UI keeps safety instructions without host implementation details', async () => {
  const h=harness(async()=>status({job:job({state:'failed',error:'C:/private/host/path failed',safety_copy:'.restore-server-secret/original',message:'internal staging error'})}));
  h.open();await flush();
  const copy=h.container.innerHTML.replace(/<[^>]*>/g,' ');
  assert.doesNotMatch(copy,/Hosting node|allowlist|staging|retention|durable|worker|OS quota/i);
  assert.match(copy,/empty, stopped server/);assert.match(copy,/fingerprint with.*provider/);assert.match(copy,/Stop the source server/);assert.match(copy,/will not be changed/);
  assert.doesNotMatch(h.nodes['job-message'].textContent,/private\/host|\.restore-|staging/);assert.match(h.nodes['job-message'].textContent,/contact support/);
  assert.equal(h.nodes.limits.textContent,'Up to 20 GiB per import.');
});

test('admin import UI retains collapsed operator help and recovery diagnostics', async () => {
  const h=harness(async()=>status({job:job({state:'failed',error:'private staging failure',safety_copy:'.restore-server/original',user_message:'Contact support.'})}));
  h.c.S.user.role='admin';h.open();await flush();
  assert.match(h.container.innerHTML,/<details[^>]*><summary>Operator details/);assert.match(h.container.innerHTML,/Hosting node: worker/);
  assert.match(h.nodes['job-message'].textContent,/private staging failure/);assert.match(h.nodes['job-message'].textContent,/\.restore-server\/original/);
});

test('failed import connections do not show proxy addresses to users', async () => {
  for (const role of ['user','admin']) {
    const h=harness(async()=>{throw Object.assign(Error('dial tcp private-worker:8080: refused'),{status:502});});
    h.c.S.user.role=role;h.open();await flush();
    if(role==='admin') assert.match(h.nodes.error.textContent,/private-worker:8080/);
    else {assert.doesNotMatch(h.nodes.error.textContent,/private-worker|dial tcp/);assert.match(h.nodes.error.textContent,/Refresh.*progress/);}
    assert.equal(h.nodes.start.disabled,true);
  }
});

test('stopped server with existing files enables Start and sends preserve or replace conflict_mode', async () => {
  const h=harness(async(_,o)=>o?.method==='POST'?job():status({eligible:false,stopped:true,has_files:true,reason:'Your server still has files.'}));
  h.open();await flush();h.fill();
  assert.equal(h.nodes['conflict-mode'].value,'preserve');
  assert.equal(h.nodes.start.disabled,false);
  h.submit();await flush();
  assert.equal(JSON.parse(h.calls[1].options.body).conflict_mode,'preserve');

  h.fill();h.nodes['conflict-mode'].value='replace';h.nodes['conflict-mode'].onchange();await flush();
  assert.equal(h.nodes.start.disabled,false);
  h.submit();await flush();
  assert.equal(JSON.parse(h.calls[3].options.body).conflict_mode,'replace');
});
