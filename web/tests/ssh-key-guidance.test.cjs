const test=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs'),path=require('node:path'),vm=require('node:vm');
const app=fs.readFileSync(path.join(__dirname,'../static/js/app.js'),'utf8');
function setup(key='ssh-ed25519 AAAATEST fixture') {
 const fields={'#kn':{value:'Laptop'},'#kk':{value:key},'#ssh-key-error':{hidden:true,textContent:''}},requests=[],notices=[];
 const c={$:id=>fields[id],toast:(...args)=>notices.push(args),loadAccountData(){c.refreshed=true;},api:async(url,options)=>{requests.push({url,options});return{};}};
 vm.runInNewContext(app.slice(app.indexOf('function sshPublicKeyInputError('),app.indexOf('async function deleteSSHKey(')),c);
 return{c,fields,requests,notices};
}
test('hex strings and fingerprints produce actionable errors without a request',async()=>{
 for(const value of ['ab'.repeat(32),'SHA256:some-fingerprint','MD5:aa:bb:cc:dd','aa:bb:cc:dd:ee:ff']){
  const {c,requests,fields}=setup(value);await c.addSSHKey();assert.equal(requests.length,0);assert.match(fields['#ssh-key-error'].textContent,/\.pub file/);assert.equal(fields['#ssh-key-error'].hidden,false);assert.equal(fields['#kk'].value,value);
 }
});
test('private keys are stopped in the browser instead of being uploaded',async()=>{
 for(const value of ['-----BEGIN OPENSSH PRIVATE KEY-----\nprivate-data','-----BEGIN RSA PRIVATE KEY-----\nprivate-data','PuTTY-User-Key-File-3: ssh-ed25519']){const s=setup(value);await s.c.addSSHKey();assert.equal(s.requests.length,0);assert.match(s.fields['#ssh-key-error'].textContent,/Do not upload/);assert.doesNotMatch(s.fields['#ssh-key-error'].textContent,/private-data/);}
});
test('public key input reaches backend validation and successful adds refresh the list',async()=>{
 const {c,requests}=setup();await c.addSSHKey();assert.equal(requests.length,1);assert.equal(requests[0].url,'/api/account/ssh-keys');assert.equal(requests[0].options.suppressToast,true);assert.deepEqual(JSON.parse(requests[0].options.body),{name:'Laptop',public_key:'ssh-ed25519 AAAATEST fixture'});assert.equal(c.refreshed,true);
});
test('backend errors remain inline without unrelated server-console advice',async()=>{
 const s=setup();s.c.api=async()=>{throw Error('Invalid OpenSSH public key format');};await s.c.addSSHKey();assert.match(s.fields['#ssh-key-error'].textContent,/full OpenSSH public key line/);assert.equal(s.c.refreshed,undefined);assert.doesNotMatch(s.notices.flat().join(' '),/server console|refresh status/);
 s.c.api=async()=>{throw Error('Network unavailable');};await s.c.addSSHKey();assert.equal(s.fields['#ssh-key-error'].textContent,'Network unavailable');
});
test('key setup explains generation, public/private distinction, and SFTP-only access',()=>{
 const card=app.slice(app.indexOf('<h2>Add OpenSSH Public Key</h2>'),app.indexOf('<h2>Add OpenSSH Public Key</h2>')+2600);
 for(const text of ['.pub file','ssh-keygen -t ed25519','Get-Content','cat ~/.ssh/id_ed25519.pub','Do not overwrite','does not enable SSH shell access','role="alert"'])assert.ok(card.includes(text),text);
});
