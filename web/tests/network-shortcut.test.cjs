const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const app = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
function renderNetwork(allowed, node = 'local') {
  const server = {id:'server-one',node_id:node,port:25565};
  const calls = [];
  const c = {S:{ports:[],sftp:null},curServer:()=>server,srvHost:()=> 'play.example.com',srvh:()=>'',ic:()=>'',esc:s=>String(s ?? ''),
    hasServerPermission:(permission,srv)=>{assert.equal(permission,'sftp.access');assert.equal(srv,server);return allowed;},
    routeUrl:(view,opts)=>{calls.push({view,server:opts.srv});return `/server/${opts.srv.node_id}/${opts.srv.id}/sftp`;}};
  const start=app.indexOf('  net() {'),end=app.indexOf('  db() {',start);
  assert.ok(start>=0 && end>start);
  const body=app.slice(start+'  net() '.length,end).trim().replace(/,$/,'');
  vm.runInNewContext('this.render=function() '+body,c);
  return {html:c.render(),calls};
}
test('Network offers a direct SFTP button for the selected local or remote server',()=>{
  for(const node of ['local','remote']) {
    const {html,calls}=renderNetwork(true,node);
    assert.match(html,new RegExp(`href="/server/${node}/server-one/sftp"`));
    assert.match(html,/class="btn"[^>]+onclick="return handleNavClick\(event, 'sftp'\)"[^>]*> Open SFTP<\/a>/);
    assert.equal(calls.length,1);assert.equal(calls[0].view,'sftp');assert.equal(calls[0].server.node_id,node);
    assert.doesNotMatch(html,/File transfer connections are managed/);
  }
});
test('Network does not expose an active SFTP shortcut without permission',()=>{
  const {html,calls}=renderNetwork(false);
  assert.match(html,/<button type="button" class="btn" disabled title="SFTP access permission is required"> Open SFTP<\/button>/);
  assert.doesNotMatch(html,/href="[^\"]*sftp"/);assert.equal(calls.length,0);
});
test('Settings no longer contains the redundant Network/SFTP instruction',()=>{
  assert.doesNotMatch(app,/Use Network for ports and SFTP for file-transfer access\./);
});
