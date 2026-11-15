const {test} = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const root = path.join(__dirname, '../..');

test('Windows panel packaging is desktop-only; legacy launchers are absent', () => {
  for (const name of ['build-windows.ps1', 'start-windows.ps1', 'start-windows.cmd']) {
    assert.equal(fs.existsSync(path.join(root, name)), false, name);
  }
  const build = fs.readFileSync(path.join(root, 'build-windows-desktop.ps1'), 'utf8');
  assert.match(build, /NeoPanel-Desktop\.exe/);
  assert.match(build, /\$files = @\(\$exe, \$readme\)/);
  assert.doesNotMatch(build, /start-windows|NeoPanel-node\.exe|'NeoPanel\.exe'/);
});

test('Windows docs direct panel users to the desktop and retain API-only workers', () => {
  const guide = fs.readFileSync(path.join(root, 'docs/windows.md'), 'utf8');
  const readme = fs.readFileSync(path.join(root, 'README.md'), 'utf8');
  assert.match(guide, /NeoPanel Desktop is the Windows panel edition/);
  assert.match(guide, /go build -trimpath -tags node/);
  assert.doesNotMatch(guide + readme, /start-windows|build-windows\.ps1|For the Windows browser\/CLI version/);
});
