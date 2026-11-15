const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const app = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
const source = app.slice(app.indexOf('async function sendAdminTestEmail('), app.indexOf('// Password recovery is implemented'));

test('SMTP test shows an escaped sidebar-logo warning instead of unconditional success', async () => {
  for (const warning of ['', 'Sidebar logo failed <img src=x onerror=alert(1)>']) {
    const alerts = [], button = {};
    const context = {$:()=>({value:'recipient@example.test'}), api:async()=>({message:'Sent.',logo_warning:warning}),
      toast(){}, uiAlert:value=>alerts.push(value), setBtnLoading:(b,busy)=>{b.disabled=busy;},
      esc:value=>String(value).replaceAll('&','&amp;').replaceAll('<','&lt;').replaceAll('>','&gt;')};
    vm.runInNewContext(source,context);
    await context.sendAdminTestEmail(button);
    assert.equal(button.disabled,false);
    assert.equal(alerts[0].type,warning?'warning':'success');
    assert.match(alerts[0].message,/recipient@example.test/);
    if(warning) {assert.match(alerts[0].message,/Sidebar logo failed/);assert.doesNotMatch(alerts[0].message,/<img/);}
  }
});
