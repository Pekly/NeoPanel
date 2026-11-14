const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/password-reset.js'), 'utf8');
const app = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
const token = 'a'.repeat(64);
function deferred() { let resolve, reject; const promise = new Promise((a,b) => { resolve=a; reject=b; }); return { promise, resolve, reject }; }
function setup() {
  const nodes = {}, timers = new Map(), requests = [], removed = [], notices = []; let id = 0, html = '', c;
  const element = (name, value = '') => ({ id: name, value, textContent: '', disabled: false, dataset: {}, attrs: {}, focus() { c.focused = name; }, setAttribute(k,v) { this.attrs[k]=v; }, remove() { delete nodes[name]; }, querySelectorAll: () => [] });
  const container = element('modal-container');
  Object.defineProperty(container, 'innerHTML', { get: () => html, set(value) {
    html = value;
    for (const key of Object.keys(nodes)) if (key !== 'modal-container' && !['lu','lp','login'].includes(key)) delete nodes[key];
    for (const match of value.matchAll(/<[^>]*\bid="([^"]+)"[^>]*>/g)) {
      const node = element(match[1], /\bvalue="([^"]*)"/.exec(match[0])?.[1] || '');
      if (match[1] === 'password-reset-form') node.dataset.mode = /data-mode="([^"]+)"/.exec(match[0])?.[1];
      nodes[match[1]] = node;
    }
    if (nodes['password-reset-form']) nodes['password-reset-form'].querySelectorAll = () => Object.entries(nodes).filter(([key]) => /^reset-/.test(key) || key === 'password-reset-submit').map(([,n]) => n);
  } });
  nodes['modal-container'] = container; nodes.lu = element('lu'); nodes.lp = element('lp');
  c = vm.createContext({ S: { turnstileEnabled: false, token: 'old-session', user: { id: 1 }, servers: [{}] }, _uiModalResolver: null,
    document: { documentElement: { dataset: { theme: 'dark' } } }, $: selector => nodes[selector.slice(1)] || null,
    location: { pathname: '/', search: '?safe=1', hash: '' }, history: { state: {}, replaceState(state,title,url) { c.replacedURL = url; c.location.hash = url.includes('#') ? '#'+url.split('#')[1] : ''; } },
    localStorage: { removeItem: key => removed.push(key) }, URLSearchParams, TextEncoder, AbortController,
    esc: value => String(value).replaceAll('&','&amp;').replaceAll('<','&lt;').replaceAll('"','&quot;'), ic: () => '',
    closeModal() { container.innerHTML = ''; const resolver = c._uiModalResolver; c._uiModalResolver = null; resolver?.(null); },
    showLoginScreen() { c.loginShown = true; }, resetServerAccess() { c.accessReset = true; }, toast: (...args) => notices.push(args),
    setTimeout(fn, ms) { timers.set(++id, {fn, ms}); return id; }, clearTimeout: key => timers.delete(key),
    fetch: async (url, options) => { requests.push({url,options}); return { ok: true, status: 200, json: async () => ({message:'If an account matches, instructions have been sent.'}) }; }
  });
  c.window=c; vm.runInContext(source+'\nthis.state=PASSWORD_RESET;', c);
  return { c, nodes, requests, timers, removed, notices, html: () => html, flush(ms) { for(const [id,t] of [...timers]) if(t.ms===ms){timers.delete(id);t.fn();} } };
}
test('reset UI uses working forms and is loaded before app; obsolete modal API is gone', () => {
  const {c,nodes,html}=setup(); c.openResetPasswordModal();
  assert.equal(nodes['password-reset-form'].dataset.mode,'request'); assert.match(html(), /role="dialog"/); assert.match(html(), /onsubmit=/);
  c.openSetNewPasswordModal(token); assert.equal(nodes['reset-token-input'].value,token); assert.match(html(), /64-character/);
  assert.doesNotMatch(app, /openModal\(/);
  const index=fs.readFileSync(path.join(__dirname,'../static/index.html'),'utf8'); assert.ok(index.indexOf('/js/password-reset.js')<index.indexOf('/js/app.js'));
});
test('request success goes directly to the code form without trusting returned tokens', async () => {
  const {c,nodes,html}=setup(); c.openResetPasswordModal(); nodes['reset-identity'].value='owner';
  c.fetch=async()=>({ok:true,status:200,json:async()=>({message:'If an account matches…',token:'must-not-be-used'})});
  await c.submitResetPasswordRequest(); assert.equal(nodes['password-reset-form'].dataset.mode,'confirm');
  assert.equal(nodes['reset-token-input'].value,''); assert.doesNotMatch(html(),/must-not-be-used/); assert.match(html(),/If an account matches/);
});
test('duplicate submit is single-flight and late response cannot replace another dialog', async () => {
  const {c,nodes}=setup(), wait=deferred(); let requests=0; c.fetch=()=>{requests++;return wait.promise;};
  c.openResetPasswordModal();nodes['reset-identity'].value='owner';const pending=c.submitResetPasswordRequest();await c.submitResetPasswordRequest();
  assert.equal(requests,1);assert.equal(nodes['password-reset-submit'].disabled,true);
  c.openSetNewPasswordModal(token);wait.resolve({ok:true,status:200,json:async()=>({message:'sent'})});await pending;
  assert.equal(nodes['reset-token-input'].value,token);assert.equal(c.state.busy,false);
});
test('HTTP failure and malformed responses do not look like successful email delivery', async () => {
  for(const response of [{ok:false,status:503,json:async()=>({})},{ok:true,status:200,json:async()=>({})},{ok:true,status:200,json:async()=>{throw Error('Invalid JSON');}}]) {
    const {c,nodes}=setup();c.fetch=async()=>response;c.openResetPasswordModal();nodes['reset-identity'].value='owner';await c.submitResetPasswordRequest();
    assert.equal(nodes['password-reset-form'].dataset.mode,'request');assert.ok(nodes['password-reset-error'].textContent);assert.equal(c.state.busy,false);
  }
});
test('reset code, matching passwords, and UTF-8 byte length are validated before POST', async () => {
  for(const [code,password,confirmation] of [['x','password','password'],[token,'short','short'],[token,'password','different'],[token,'é'.repeat(37),'é'.repeat(37)]]) {
    const {c,nodes,requests}=setup();c.openSetNewPasswordModal(code);nodes['reset-token-input'].value=code;nodes['reset-new-password'].value=password;nodes['reset-new-password-confirm'].value=confirmation;
    await c.submitSetNewPassword();assert.equal(requests.length,0);assert.ok(nodes['password-reset-error'].textContent);
  }
});
test('successful password change clears stale login state and returns to sign-in', async () => {
  const {c,nodes,requests,removed}=setup();c.openSetNewPasswordModal(token.toUpperCase());nodes['reset-new-password'].value='new-password';nodes['reset-new-password-confirm'].value='new-password';
  c.fetch=async(url,options)=>{requests.push({url,options});return{ok:true,status:200,json:async()=>({message:'updated',username:'owner'})};};
  await c.submitSetNewPassword(); assert.equal(requests[0].url,'/api/auth/password-reset/confirm');
  assert.deepEqual(JSON.parse(requests[0].options.body),{token,new_password:'new-password'});assert.equal(c.S.token,'');assert.equal(c.S.user,null);assert.ok(c.loginShown&&c.accessReset);assert.deepEqual(removed,['neopanel_token']);assert.equal(nodes.lu.value,'owner');
});
test('timeout aborts the request and restores form controls; cancel disposes state', async () => {
  const h=setup(),{c,nodes}=h;c.openResetPasswordModal();nodes['reset-identity'].value='owner';
  c.fetch=(url,options)=>new Promise((resolve,reject)=>options.signal.addEventListener('abort',()=>reject(Object.assign(Error('timeout'),{name:'AbortError'}))));
  const pending=c.submitResetPasswordRequest();h.flush(20000);await pending;assert.match(nodes['password-reset-error'].textContent,/timed out/);assert.equal(nodes['password-reset-submit'].disabled,false);
  c.closeModal();assert.equal(h.timers.size,0);assert.equal(c.state.controller,null);
});
test('link tokens are consumed once, removed from history, and malformed encoding is safe', () => {
  for(const raw of [token,'%E0%A4%A','']) {
    const {c,nodes,html}=setup();c.location.hash='#reset-token='+raw;c.checkUrlResetToken();
    assert.equal(c.replacedURL,'/?safe=1');assert.equal(c.location.hash,'');
    assert.equal(nodes['reset-token-input'].value,raw===token?token:'');if(raw!==token)assert.match(html(),/reset link is invalid/);
    const generation=c.state.generation;c.checkUrlResetToken();assert.equal(c.state.generation,generation);
  }
});
test('reset CAPTCHA has its own token/widget and cannot consume the login CAPTCHA', async () => {
  const {c,nodes,requests}=setup();let options;const removed=[];c.S.turnstileEnabled=true;c.S.turnstileSiteKey='site';c.S.turnstileToken='login-token';c.S.turnstileWidgetId=99;
  c.turnstile={render:(box,opts)=>{options=opts;return 7;},remove:id=>removed.push(id),reset:()=>{}};
  c.openResetPasswordModal();nodes['reset-identity'].value='owner';await c.submitResetPasswordRequest();assert.equal(requests.length,0);
  options.callback('reset-token');await c.submitResetPasswordRequest();assert.equal(JSON.parse(requests[0].options.body)['cf-turnstile-response'],'reset-token');assert.equal(c.S.turnstileToken,'login-token');assert.deepEqual(removed,[7]);
});

test('CAPTCHA rendering errors remain recoverable without borrowing the login token', () => {
  const {c,nodes}=setup();c.S.turnstileEnabled=true;c.S.turnstileSiteKey='invalid';c.S.turnstileToken='login-token';
  c.turnstile={render:()=>{throw Error('bad sitekey');}};
  assert.doesNotThrow(()=>c.openResetPasswordModal());assert.match(nodes['turnstile-reset'].textContent,/unavailable/);assert.equal(c.state.captcha,'');assert.equal(c.S.turnstileToken,'login-token');
});

test('password reset backdrop is above the login overlay', () => {
  const {c,html}=setup();c.openResetPasswordModal();assert.match(html(),/class="pal np-password-reset-backdrop"/);
  const css=fs.readFileSync(path.join(__dirname,'../static/css/workbench.css'),'utf8');
  const z=Number(/\.np-password-reset-backdrop\s*\{\s*z-index:\s*(\d+)/.exec(css)?.[1]);
  assert.ok(z>9999,'login overlay uses z-index 9999');
});
