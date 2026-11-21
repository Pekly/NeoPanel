const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/onboarding.js'), 'utf8');
const app = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
const admin = {id: 'admin-one', role: 'admin', created_at: new Date(Date.now() - 3600000).toISOString()};
function guide(storage = new Map(), broken = false) {
  const events = {}, calls = [], root = {innerHTML: '', scrollIntoView() {}}, focus = [];
  const c = {localStorage: {getItem(k) {if (broken) throw Error('blocked'); return storage.get(k);}, setItem(k,v) {if (broken) throw Error('blocked'); storage.set(k,v);}},
    document: {addEventListener(n,fn) {events[n] = fn;}, getElementById(id) {
      if (id === 'np-onboarding-root') return root;
      if (id === 'np-guide-title' && !root.innerHTML) return null;
      return {focus() {focus.push(id);}, scrollIntoView() {}};
    }}, go: (...args) => calls.push(args), openCreateServerModal: () => calls.push(['create'])};
  c.window = c; vm.runInNewContext(source,c);
  const state = {token:'session',user:admin};
  const render = (s=state) => root.innerHTML = c.NPOnboarding.render(s);
  function click(action, step) {
    const b = {dataset:{guideAction:action,guideStep:step},hasAttribute() {return step !== undefined;},closest() {return root;}};
    events.click({target:{closest() {return b;}}});
  }
  return {api:c.NPOnboarding,render,click,root,storage,calls,focus,state};
}
test('fresh setup shows the guide, next/back preserve progress across sessions', () => {
  const g=guide(); g.api.begin(admin); assert.match(g.render(), /Make this account yours/);
  g.click('next'); assert.match(g.root.innerHTML, /Decide where Minecraft/);
  assert.equal(g.focus.at(-1),'np-guide-title');
  const resumed=guide(g.storage); assert.match(resumed.render(), /Decide where Minecraft/);
  resumed.click('back'); assert.match(resumed.root.innerHTML, /Make this account yours/);
});
test('pause and finish stay dismissed; Getting started resumes or restarts', () => {
  const g=guide(); g.api.begin(admin); g.render(); g.click('next'); g.click('pause');
  assert.equal(g.root.innerHTML,''); assert.equal(g.focus.at(-1),'np-guide-open');
  const resumed=guide(g.storage); assert.doesNotMatch(resumed.render(), /np-guide-detail/);
  resumed.api.open(); assert.match(resumed.root.innerHTML,/Decide where Minecraft/);
  resumed.click(null,'3'); resumed.click('next'); assert.equal(resumed.root.innerHTML,'');
  resumed.api.open(); assert.match(resumed.root.innerHTML,/Make this account yours/);
});
test('existing users are not interrupted and guide state is scoped to an authenticated admin', () => {
  const g=guide(); assert.doesNotMatch(g.render(),/np-guide-detail/);
  g.api.begin(admin); g.render(); g.click('next');
  assert.doesNotMatch(g.render({token:'other',user:{id:'admin-two',role:'admin'}}),/np-guide-detail/);
  assert.equal(g.render({token:'user',user:{id:'u',role:'user'}}),'');
  g.api.open(); g.click('create'); assert.equal(g.calls.length,0);
  assert.equal(g.render({user:admin}),'');
  g.render(); g.api.resetSession(); g.click('create'); assert.equal(g.calls.length,0);
});
test('blocked and malformed browser storage cannot break onboarding', () => {
  const g=guide(new Map(),true); g.api.begin(admin); g.render(); g.click('next'); assert.match(g.root.innerHTML,/Decide where Minecraft/);
  for (const value of ['{bad','{"step":999,"status":"active"}','{"step":"<img>","status":"active"}']) {
    const saved=new Map([['neopanel_getting_started_v1_admin-one',value]]);
    const h=guide(saved); assert.doesNotThrow(()=>h.render()); h.api.open(); h.click(null,'NaN'); assert.doesNotMatch(h.root.innerHTML,/<img>/);
  }
});
test('guide actions navigate only; tutorial completion does not configure a server', () => {
  const g=guide(); g.api.begin(admin); g.render(); g.click('security'); g.click('nodes'); g.click('create');
  assert.equal(JSON.stringify(g.calls),JSON.stringify([['acc',{acctTab:1}],['admin',{adminTab:1}],['create']]));
  assert.doesNotMatch(source,/\bfetch\s*\(/);
  g.click(null,'3'); assert.match(g.root.innerHTML,/restoring replaces current files/);
});
function setupForm() {
  const fields = {'#su-name':{value:'alex'}, '#su-email':{value:'alex@example.com'}, '#su-pw':{value:'strong-password'}, '#su-confirm':{value:'strong-password',focus(){}}, '#setup-form':{reportValidity:()=>true}, '#lerr':{textContent:''}, '#login .lbtn':{}, '#login':{remove(){}}};
  let calls=0, begins=0;
  const c={S:{}, $:id=>fields[id], setBtnLoading(btn,loading){btn.disabled=loading;}, resetTurnstileWidget(){},
    localStorage:{setItem(){}},history:{replaceState(){}},withSafeModeQuery:x=>x,toast(){},getBrand:()=>'',checkAuth:async()=>{},
    NPOnboarding:{begin(){begins++;}}, fetch:async()=>{calls++; return {ok:true,json:async()=>({token:'test',user:admin})};}};
  c.window=c;
  vm.runInNewContext(app.slice(app.indexOf('async function submitSetup()'),app.indexOf('function showLoginScreen()')),c);
  return {c,fields,calls:()=>calls,begins:()=>begins};
}
test('setup validates confirmation and native fields without posting', async () => {
  const g=setupForm(); g.fields['#su-confirm'].value='different'; await g.c.submitSetup();
  assert.equal(g.calls(),0); assert.match(g.fields['#lerr'].textContent,/do not match/);
  g.fields['#setup-form'].reportValidity=()=>false; await g.c.submitSetup(); assert.equal(g.calls(),0);
});
test('successful setup starts tutorial only after HTTP success', async () => {
  const g=setupForm(); await g.c.submitSetup(); assert.equal(g.begins(),1); assert.equal(g.c.S.setupSubmitting,false);
  const failed=setupForm(); failed.c.fetch=async()=>({ok:false,json:async()=>({token:'invalid',error:'Already initialized'})});
  await failed.c.submitSetup(); assert.equal(failed.begins(),0); assert.equal(failed.fields['#lerr'].textContent,'Already initialized');
});
test('double submission is blocked and failed requests unlock retry', async () => {
  const g=setupForm(); let resolve, calls=0;
  g.c.fetch=()=>{calls++; return new Promise(r=>resolve=r);};
  const first=g.c.submitSetup(); await g.c.submitSetup(); assert.equal(calls,1);
  resolve({ok:false,json:async()=>({error:'Try again'})}); await first;
  assert.equal(g.c.S.setupSubmitting,false); assert.equal(g.fields['#login .lbtn'].disabled,false);
  g.c.fetch=async()=>{throw Error('offline');}; await g.c.submitSetup(); assert.match(g.fields['#lerr'].textContent,/Check your connection/);
});
test('setup fields have visible labels and native form submission', () => {
  const form=app.slice(app.indexOf('function showSetupScreen()'),app.indexOf('async function submitSetup()'));
  for(const id of ['su-name','su-email','su-pw','su-confirm']) assert.ok(form.includes(`for="${id}"`));
  assert.match(form,/<form[^>]+onsubmit=/); assert.match(form,/type="submit"/);
  assert.match(app,/NPOnboarding\?\.resetSession\(\)/);
});
test('normal Linux users never see or open the guide, even with saved admin progress', () => {
  const user = {id:'normal-user',role:'user',panel_permissions:['servers.create']};
  const saved = new Map([['neopanel_getting_started_v1_normal-user',JSON.stringify({step:0,status:'active'})]]);
  const g = guide(saved);
  g.api.begin(user);
  assert.equal(g.render({token:'user-session',user}), '');
  g.api.open(); g.click('create'); g.click('security');
  assert.equal(g.root.innerHTML, ''); assert.equal(g.calls.length, 0);
  assert.match(app, /NPOnboarding\?\.eligible\(S\) \? `<button id="np-guide-open"/);
});
test('guide and launcher are unavailable after the admin account is three days old', () => {
  const g = guide();
  for (const created_at of [new Date(Date.now() - 3 * 86400000 - 1000).toISOString(), 'bad', undefined, new Date(Date.now() + 86400000).toISOString()]) {
    const user = {...admin, created_at};
    g.api.begin(user);
    assert.equal(g.api.eligible({token:'test', user}), false);
    assert.equal(g.render({token:'test', user}), '');
    g.api.open(); g.click('create');
  }
  assert.equal(g.calls.length, 0);
  assert.equal(g.api.eligible({token:'test', user:{...admin, created_at:new Date(Date.now() - 2 * 86400000).toISOString()}}), true);
});
