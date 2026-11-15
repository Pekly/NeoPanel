const {test} = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const path = require('node:path');
const script = fs.readFileSync(path.join(__dirname,'../internal/desktop/panel-session.js'),'utf8');
function setup(topLevel=false, pathname='/') {
  const events={}, messages=[];
  let timeout;
  const parent={postMessage:(...args)=>messages.push(args)};
  const window={parent, addEventListener:(n,f)=>events[n]=f,removeEventListener:n=>delete events[n]};
  if(topLevel) window.parent=window;
  vm.runInNewContext(script,{window,location:{protocol:'http:',hostname:'127.0.0.1',port:'1234',pathname},setTimeout:f=>(timeout=f),clearTimeout:()=>{}});
  return {window,parent,events,messages,timeout};
}
test('ordinary browser and file previews do not bootstrap a desktop session',()=>{
  assert.equal(setup(true).window.NPDesktopSession,undefined);
  assert.equal(setup(false,'/api/servers/x/files/preview').window.NPDesktopSession,undefined);
});
test('receiver accepts only its Wails parent and removes the listener after one handoff',async()=>{
  const {window,parent,events,messages}=setup();
  assert.equal(messages[0][1],'http://wails.localhost');
  const receive=events.message;
  for(const [source,origin] of [[{},'http://wails.localhost'],[parent,'http://evil.example']]) receive({source,origin,data:{type:'np:desktop-session',token:'bad'}});
  assert.equal(events.message,receive);
  receive({source:parent,origin:'http://wails.localhost',data:{type:'np:desktop-session',token:'private',guest:true}});
  assert.equal(await window.NPDesktopSession,'private');
  assert.equal(window.NPDesktopGuest,true);
  assert.equal(events.message,undefined);
});
test('missing native shell fails closed to normal login',async()=>{
  const {window,events,timeout}=setup();
  timeout();
  assert.equal(await window.NPDesktopSession,'');
  assert.equal(events.message,undefined);
});
test('application consumes native session before checkAuth; ordinary startup stays unchanged', async()=>{
  const app=fs.readFileSync(path.join(__dirname,'../../web/static/js/app.js'),'utf8');
  const startup=app.slice(app.indexOf('// --- Initialize Application ---'),app.indexOf('// --- Permission-aware workspace'));
  for(const desktop of [false,true]) {
    let callback, deliver;
    const seen=[], storage=[];
    const window=desktop ? {NPDesktopSession:new Promise(resolve=>deliver=resolve)} : {};
    const S={token:'previous'};
    vm.runInNewContext(startup,{window,S,document:{addEventListener:(_,f)=>callback=f,body:{}},localStorage:{setItem:(...args)=>storage.push(args)},checkAuth:()=>seen.push(S.token),initDragAndDrop(){},enhanceSelects(){},checkUrlResetToken(){},globalSelectObserver:null});
    const pending=callback();
    if(desktop) {
      assert.equal(seen.length,0);
      deliver('native-session');
    }
    await pending;
    assert.deepEqual(seen,[desktop?'native-session':'previous']);
    assert.equal(storage.length,desktop?1:0);
    assert.equal(window.NPDesktopSession,undefined);
  }
});
