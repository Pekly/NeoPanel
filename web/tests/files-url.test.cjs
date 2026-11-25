const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const appPath = path.join(__dirname, '../static/js/app.js');
const source = fs.readFileSync(appPath, 'utf8');

test('app.js defines openUploadUrlModal, onUploadUrlInput and submitUploadFromURL', () => {
  assert.match(source, /function openUploadUrlModal\(/, 'openUploadUrlModal should be defined');
  assert.match(source, /function onUploadUrlInput\(/, 'onUploadUrlInput should be defined');
  assert.match(source, /async function submitUploadFromURL\(/, 'submitUploadFromURL should be defined');
  assert.match(source, /window\.openUploadUrlModal\s*=\s*openUploadUrlModal;/, 'openUploadUrlModal should be on window');
  assert.match(source, /window\.submitUploadFromURL\s*=\s*submitUploadFromURL;/, 'submitUploadFromURL should be on window');
});

test('app.js renders Download from URL button in file manager toolbar', () => {
  assert.match(source, /openUploadUrlModal\(\)/, 'Toolbar should include call to openUploadUrlModal');
  assert.match(source, /Download from URL/, 'Toolbar menu should display "Download from URL"');
});

test('app.js renders persistent background transfer banner for imports and URL downloads', () => {
  assert.match(source, /function renderServerTransferBannerHtml\(/, 'renderServerTransferBannerHtml should be defined');
  assert.match(source, /async function syncServerBackgroundTransfers\(/, 'syncServerBackgroundTransfers should be defined');
  assert.match(source, /id="srv-transfer-banner-root"/, 'srvh should include srv-transfer-banner-root');
});
