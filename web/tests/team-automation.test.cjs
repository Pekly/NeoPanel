const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const teamSource = fs.readFileSync(path.join(__dirname, '../static/js/team.js'), 'utf8');
const automationSource = fs.readFileSync(path.join(__dirname, '../static/js/automation.js'), 'utf8');
const app = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
const tick = () => new Promise(resolve => setImmediate(resolve));
function setup() {
  const elements = {}, storage = new Map(), notices = [];
  const location = { hash: '', pathname: '/', search: '' };
  const c = vm.createContext({
    console, URLSearchParams, TextEncoder, Date, setTimeout: () => 0,
    S: { user: { id: 1, email: 'owner@example.test' }, token: 'session', view: 'team', servers: [{ id: 'one', node_id: 'local', name: 'One' }] },
    SUBUSER_PERMS: [['console.view','View console','Read logs'],['subusers.manage','Manage team','Invite users']],
    location,
    history: { state: null, replaceState: () => { location.hash = ''; } },
    sessionStorage: { getItem: key => storage.get(key) || '', setItem: (key, value) => storage.set(key, value), removeItem: key => storage.delete(key) },
    localStorage: { removeItem: () => {} },
    document: { querySelector: selector => elements[selector] || null, createElement: () => ({ innerHTML: '', querySelector: () => null }), body: { appendChild: element => { elements['#' + element.id] = element; } } },
    esc: value => String(value ?? '').replaceAll('&','&amp;').replaceAll('<','&lt;').replaceAll('>','&gt;').replaceAll('"','&quot;'),
    ic: () => '', srvh: () => '<header></header>', loaderHtml: () => 'Loading', renderSubusersListHtml: () => 'Members', formatFileDate: value => value || '', brandLogoHtml: () => 'NeoPanel',
    R: () => {}, closeModal: () => {}, closeAllCustomSelects: () => {}, enhanceSelects: () => {},
    toast: (message, kind) => notices.push({message,kind}), uiAlert: async opts => notices.push(opts), uiConfirm: async () => true,
    setBtnLoading: (button, busy) => { if (button) button.disabled = busy; }, loadBackupsList: async () => {}, loadServerSchedules: async () => {}, loadPublicBranding: async () => {},
    srvApiPath: (srv, suffix) => `/api/servers/${srv.id}${suffix}`,
    sameQolServer: (a,b) => Boolean(a&&b&&a.id===b.id&&(a.node_id||'local')===(b.node_id||'local')),
    api: async () => ({}), fetch: async () => ({ok:true,json:async()=>({})})
  });
  c.window=c;c.curServer=()=>c.S.servers[0];c.$=selector=>elements[selector]||null;c.$$=()=>[];
  vm.runInContext(teamSource+'\n'+automationSource+'\nthis.team=TeamUI;this.invite=InvitationUI;this.automation=AutomationUI;',c);
  return {c,elements,storage,notices};
}

test('Team permissions distinguish owners/managers from ordinary subusers',()=>{
  const {c}=setup();assert.equal(c.canManageTeam({is_owner:true}),true);assert.equal(c.canManageTeam({is_admin:true}),true);
  assert.equal(c.canManageTeam({permissions:['subusers.manage']}),true);assert.equal(c.canManageTeam({permissions:['console.view']}),false);
});
test('Team loads envelopes and keeps remote invitations on the master route',async()=>{
  const {c}=setup();c.S.servers[0].node_id='remote';const calls=[];
  c.api=async url=>{calls.push(url);if(url.endsWith('/permissions'))return {permissions:['subusers.manage']};if(url.endsWith('/subusers'))return {subusers:[{id:7,username:'member',email:'member@example.test'}]};return {smtp_configured:true,invitations:[{id:'i1'}]};};
  await c.loadTeamPage();assert.equal(c.S.serverSubusers[0].username,'member');assert.equal(c.S.teamState.smtp_configured,true);
  assert.ok(calls.includes('/api/team/invitations?server_id=one&node_id=remote'));
});
test('unprivileged team view never requests member or invitation lists',async()=>{
  const {c}=setup();const calls=[];c.api=async url=>{calls.push(url);return {permissions:['console.view']};};await c.loadTeamPage();
  assert.equal(calls.length,1);assert.match(c.renderTeamPage(),/Team management is restricted/);
});
test('late Team response cannot overwrite another server',async()=>{
  const {c}=setup();let resolve;c.api=()=>new Promise(r=>{resolve=r;});const load=c.loadTeamPage();c.S.servers=[{id:'two',node_id:'remote'}];c.S.teamState={marker:'new'};resolve({is_owner:true});await load;assert.equal(c.S.teamState.marker,'new');
});
test('invitation delivery failure, expiry and lifecycle are distinct',()=>{
  const {c}=setup();assert.equal(c.invitationStatusLabel({status:'pending',delivery_status:'failed'}),'Email delivery failed');
  assert.equal(c.invitationStatusLabel({status:'pending',expires_at:'2000-01-01'}),'Expired');assert.equal(c.invitationStatusLabel({status:'accepted',expires_at:'2000-01-01'}),'accepted');
});
test('invite tokens use fragment/session state, not query strings',()=>{
  const {c,storage}=setup();c.location.hash='#invite-token=secret';assert.equal(c.pendingInvitationToken(),'secret');assert.equal(c.location.hash,'');assert.equal(storage.get('neopanel_pending_invite'),'secret');
});
test('invitation screen initializes Turnstile before sign-in',async()=>{
  const {c}=setup();c.invite.token='secret';c.S.token='';c.renderInvitationScreen=()=>{};const calls=[];
  c.fetch=async url=>{calls.push(url);return {ok:true,json:async()=>url==='/api/auth/status'?{turnstile_enabled:true,turnstile_site_key:'site-key'}:{email:'new@example.test',username:'newuser',account_exists:true}};};
  assert.equal(await c.openPendingInvitation(),true);assert.equal(c.S.turnstileEnabled,true);assert.equal(c.S.turnstileSiteKey,'site-key');assert.equal(calls[0],'/api/auth/status');
});
test('public invite failures are surfaced rather than treated as registration success',async()=>{
  const {c}=setup();c.fetch=async()=>({ok:false,status:410,json:async()=>({error:'Invitation expired'})});await assert.rejects(c.invitationRequest('inspect',{token:'old'}),/expired/);
});
test('reserved-account retry screen cannot suggest setting a new password',()=>{
  const {c,elements}=setup();c.invite.info={email:'new@example.test',username:'newuser',server_name:'Server',permissions:[],account_exists:false,account_reserved:true};
  c.renderInvitationScreen();assert.match(elements['#invite-screen'].innerHTML,/original password will not change/);assert.doesNotMatch(elements['#invite-screen'].innerHTML,/invite-register-password/);
});
test('reserved-account acceptance sends only the invite token',async()=>{
  const {c,elements,storage}=setup();c.invite.token='secret';c.invite.info={account_reserved:true,account_exists:false};storage.set('neopanel_pending_invite','secret');elements['#invite-accept-form']={reportValidity:()=>true};elements['#invite-accept-error']={};c.renderInvitationScreen=()=>{};let request;
  c.invitationRequest=async(action,body)=>{request={action,body};return {login_required:true};};await c.acceptTeamInvitation({preventDefault(){}});
  assert.equal(request.action,'accept');assert.deepEqual(Object.keys(request.body),['token']);assert.equal(storage.has('neopanel_pending_invite'),false);assert.equal(c.invite.accepted,true);
});
test('password mismatch and byte limit prevent registration requests',async()=>{
  const {c,elements}=setup();c.invite.info={account_exists:false};elements['#invite-accept-form']={reportValidity:()=>true};elements['#invite-accept-error']={};elements['#invite-register-username']={value:'user'};elements['#invite-register-password']={value:'long-password'};elements['#invite-register-confirm']={value:'different-password'};let calls=0;c.invitationRequest=async()=>{calls++;};
  await c.acceptTeamInvitation({preventDefault(){}});assert.match(elements['#invite-accept-error'].textContent,/do not match/);
  elements['#invite-register-password'].value='é'.repeat(40);elements['#invite-register-confirm'].value='é'.repeat(40);await c.acceptTeamInvitation({preventDefault(){}});assert.equal(calls,0);assert.match(elements['#invite-accept-error'].textContent,/UTF-8/);
});
test('existing subuser edit resolves numeric IDs and retains permissions',()=>{
  const {c,elements}=setup();c.S.teamState={permission:{is_owner:true}};c.S.serverSubusers=[{id:7,username:'member',permissions:['subusers.manage']}];elements['#modal-container']={innerHTML:''};
  const start=app.indexOf('function openSubuserModal('),end=app.indexOf('\nasync function submitSubuserModal',start);vm.runInContext(app.slice(start,end),c);c.openSubuserModal('7');assert.match(elements['#modal-container'].innerHTML,/Edit member: member/);assert.match(elements['#modal-container'].innerHTML,/value="subusers.manage" checked/);
});
test('existing user form sends the backend username field, not identifier',async()=>{
  const {c,elements}=setup();c.team.memberServer=c.curServer();elements['#sub-ident']={value:'member@example.test'};c.loadServerSubusers=async()=>{};let body;c.api=async(url,opts)=>{body=JSON.parse(opts.body);};
  const start=app.indexOf('async function submitSubuserModal('),end=app.indexOf('\nasync function deleteServerSubuser',start);vm.runInContext(app.slice(start,end),c);await c.submitSubuserModal({preventDefault(){}});assert.equal(body.username,'member@example.test');assert.equal(body.identifier,undefined);
});
test('legacy schedules normalize to one step and task bounds match the backend',()=>{
  const {c}=setup();const tasks=c.scheduleTasksFor({action:'restart',payload:''});assert.equal(tasks.length,1);assert.equal(tasks[0].time_offset,0);assert.equal(tasks[0].action,'restart');
  for(const tasks of [[],Array.from({length:21},()=>({action:'backup',time_offset:0})),[{action:'command',payload:'',time_offset:0}],[{action:'backup',time_offset:1.5}],[{action:'backup',time_offset:3601}],[{action:'unknown',time_offset:0}],[{action:'command',payload:'é'.repeat(3000),time_offset:0}]])assert.throws(()=>c.validateScheduleTasks(tasks));
});
test('task reordering preserves delay/payload/failure policy',()=>{
  const {c}=setup();c.automation.tasks=[{action:'command',payload:'say hi',time_offset:5,continue_on_failure:true},{action:'backup',payload:'',time_offset:10,continue_on_failure:false}];c.moveScheduleTask(1,-1);assert.equal(c.automation.tasks[0].action,'backup');assert.equal(c.automation.tasks[1].payload,'say hi');assert.equal(c.automation.tasks[1].continue_on_failure,true);c.removeScheduleTask(0);assert.equal(c.automation.tasks.length,1);c.removeScheduleTask(0);assert.equal(c.automation.tasks.length,1);
});
test('running servers cannot reach the restore POST',async()=>{
  const {c,notices}=setup();const calls=[];c.api=async(url,opts)=>{calls.push(url);return {stats:{status:'running'}};};await c.restoreBackup('backup.tar.gz');assert.equal(calls.length,1);assert.ok(notices.some(n=>n.title==='Stop the server first'));assert.equal(c.automation.restoring,null);
});
test('restore requires confirmation and reports retained safety copy',async()=>{
  const {c}=setup();const calls=[];c.api=async(url,opts)=>{calls.push({url,opts});return opts?.method==='POST'?{safety_copy:'alpha.previous-test'}:{stats:{status:'stopped'}};};c.uiConfirm=async()=>false;await c.restoreBackup('backup.tar.gz');assert.equal(calls.length,1);
  c.uiConfirm=async()=>true;await c.restoreBackup('backup.tar.gz');assert.equal(calls.length,3);assert.deepEqual(JSON.parse(calls[2].opts.body),{name:'backup.tar.gz',confirm:true});assert.match(c.renderRestoreNotice(),/alpha.previous-test/);
});

test('future invitation expiry is an explicit date, not a past-file relative label',()=>{
  const {c}=setup();const expiry=c.formatInvitationExpiry('2035-06-15T12:00:00Z');
  assert.match(expiry,/2035/);assert.doesNotMatch(expiry,/just now/i);assert.equal(c.formatInvitationExpiry('invalid'),'—');
});

test('team invitation preflight rejects existing members case-insensitively',()=>{
  const {c}=setup();c.S.serverSubusers=[{username:'Member',email:'member@example.test'}];
  assert.match(c.teamInvitationConflict('MEMBER','other@example.test'),/already has access/);
  assert.match(c.teamInvitationConflict('other','MEMBER@example.test'),/already has access/);
  assert.equal(c.teamInvitationConflict('newuser','new@example.test'),'');
});
test('pending invitation duplicates use resend while expired invites can be replaced',()=>{
  const {c}=setup();c.S.teamState={invitations:[{username:'invitee',email:'invitee@example.test',status:'pending',expires_at:'2099-01-01'}]};
  assert.match(c.teamInvitationConflict('invitee','invitee@example.test'),/Resend/);
  c.S.teamState.invitations[0].expires_at='2000-01-01';assert.equal(c.teamInvitationConflict('invitee','invitee@example.test'),'');
});
test('an unresolved granting invitation cannot be duplicated even after its expiry',()=>{
  const {c}=setup();c.S.teamState={invitations:[{username:'invitee',email:'invitee@example.test',status:'granting',expires_at:'2000-01-01'}]};
  assert.match(c.teamInvitationConflict('invitee','invitee@example.test'),/invitation already exists/);
});
