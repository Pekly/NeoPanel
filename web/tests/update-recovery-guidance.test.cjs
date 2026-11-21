const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const app = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
test('update confirmation explains downtime and running-state restoration honestly', () => {
  const start = app.indexOf('async function openApplyUpdateModal()');
  const text = app.slice(start, app.indexOf('function renderAdminUpdateBtn', start));
  assert.match(text, /Local Minecraft servers briefly stop/);
  assert.match(text, /those online before shutdown start again afterward/);
  assert.match(text, /Offline servers stay offline/);
  assert.match(text, /Remote workers are unaffected/);
  assert.doesNotMatch(text, /Running Minecraft servers will remain online/);
});
