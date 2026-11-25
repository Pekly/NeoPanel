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

test('openUploadUrlModal contains URL, progress bar, speed, and overwrite inputs', () => {
  assert.match(source, /id="up-url-input"/, 'Modal should contain up-url-input');
  assert.match(source, /id="up-url-filename"/, 'Modal should contain up-url-filename');
  assert.match(source, /id="up-url-overwrite"/, 'Modal should contain up-url-overwrite');
  assert.match(source, /id="up-url-progress-box"/, 'Modal should contain up-url-progress-box');
  assert.match(source, /formatTransferSpeed/, 'Modal should format live transfer speed');
  assert.match(source, /\/files\/upload-url/, 'Submit should target files/upload-url');
});
