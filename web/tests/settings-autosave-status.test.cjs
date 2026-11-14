const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');

function setup() {
  const elements = ['set-autosave-pill', 'set-autosave-status'].map(() => ({
    style: { display: 'none' },
    label: { textContent: '' },
    classList: { remove() {}, add() {} },
    querySelector() { return this.label; },
  }));
  const context = vm.createContext({ document: { getElementById: id => elements[id === 'set-autosave-pill' ? 0 : 1] } });
  const start = source.indexOf('let _settingsAutoTimer =');
  const end = source.indexOf('function getSettingsSnapshotKey(', start);
  vm.runInContext(source.slice(start, end), context);
  return { context, elements };
}

test('idle autosave state no longer supplies an active label', () => {
  const { context } = setup();
  assert.equal(vm.runInContext('_settingsUiState.text', context), '');
  assert.doesNotMatch(source, /Real-time auto-save active/);
  for (const id of ['set-autosave-pill', 'set-autosave-status']) {
    const line = source.split('\n').find(line => line.includes(`id="${id}"`));
    assert.ok(line.includes('display:none;'), `${id} must not leave an empty idle badge`);
  }
});

test('saving and error feedback can still reveal both status indicators', () => {
  const { context, elements } = setup();
  for (const [status, text] of [['saving', 'Saving changes...'], ['warn', 'Enter a valid value to save'], ['error', 'Auto-save failed — check values'], ['saved', 'All changes saved automatically']]) {
    context.setSettingsSaveStatus(status);
    for (const element of elements) {
      assert.equal(element.style.display, '');
      assert.equal(element.label.textContent, text);
    }
  }
});
