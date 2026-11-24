const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const appPath = path.join(__dirname, '../static/js/app.js');
const source = fs.readFileSync(appPath, 'utf8');

test('app.js defines updateClusterNode and updateAllClusterNodes', () => {
  assert.match(source, /async function updateClusterNode\(/, 'updateClusterNode should be defined');
  assert.match(source, /async function updateAllClusterNodes\(/, 'updateAllClusterNodes should be defined');
  assert.match(source, /window\.updateClusterNode\s*=\s*updateClusterNode;/, 'updateClusterNode should be exposed on window');
  assert.match(source, /window\.updateAllClusterNodes\s*=\s*updateAllClusterNodes;/, 'updateAllClusterNodes should be exposed on window');
});

test('app.js renders Update button for remote cluster nodes in table and cards', () => {
  assert.match(source, /onclick="updateClusterNode\(this\.dataset\.id,\s*this\.dataset\.name\)"/, 'Update button should invoke updateClusterNode');
  assert.match(source, /onclick="updateAllClusterNodes\(\)"/, 'Update All Nodes button should invoke updateAllClusterNodes');
});

test('app.js mentions cluster worker node updates in update source notes', () => {
  assert.match(source, /connected cluster worker nodes will also be updated automatically/, 'overview should inform user of automatic cluster node update');
});
