const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
function section(start, end) {
  const a = source.indexOf(start), b = source.indexOf(end, a);
  assert.ok(a >= 0 && b > a);
  return source.slice(a, b);
}
function setup(type = 'steel') {
  const context = vm.createContext({
    S: { status: 'starting' }, curServer: () => ({ type }),
    isSteelEngine: value => value === 'steel',
    esc: value => String(value).replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('>', '&gt;').replaceAll('"', '&quot;'),
    getPluginColor: () => ({ fg: '#fff', bg: '#000', border: '#333' }),
    updateInputState: () => {}, updateHeaderBadge: () => {},
  });
  vm.runInContext(section('function cleanConsoleOutput(', '// --- Minecraft Legacy Color') + section('function formatLogLine(', 'function updateInputState('), context);
  return context;
}

test('Steel cursor shape, movement, erase and empty prompts are not log rows', () => {
  const c = setup();
  for (const raw of ['\x1b[5 q', '\x1b[2C', '\r\x1b[J> \x1b[38;5;15m\x1b[0m\r\x1b[2C', '\r\r']) {
    assert.equal(c.formatLogLine(raw), '', JSON.stringify(raw));
  }
});

test('Steel saved log redraws retain blue Info and red Error formatting', () => {
  const c = setup();
  const info = c.formatLogLine('\r\x1b[2C\r\x1b[J> \r\x1b[J 0.00s \x1b[0;1;34m[Info]\x1b[0m Starting Steel Server (1e9d4df)\r');
  assert.match(info, /color:#3b82f6/);
  assert.match(info, /font-weight:bold/);
  assert.match(info, /Starting Steel Server/);
  assert.doesNotMatch(info, /\x1b|\r|con-cmd-row/);
  const error = c.formatLogLine(JSON.stringify({ event: 'log', data: '\r\x1b[J 3.11s \x1b[31m[Error]\x1b[0m Server startup failed: failed to bind to server port 25565: Address in use (os error 98)' }));
  assert.match(error, /color:#ef4444/);
  assert.match(error, /Address in use/);
  assert.doesNotMatch(error, /\x1b|\r/);
  assert.equal(c.S.status, 'starting', 'bind errors must not mark the server online');
});

test('terminal cleanup preserves truecolor and strips OSC without injecting HTML', () => {
  const c = setup();
  const result = c.formatLogLine('\x1b]0;title\x07\x1b]8;;https://example.test\x1b\\\x1b[38;2;12;34;56m<unsafe>\x1b[0m\x1b]8;;\x1b\\');
  assert.match(result, /color:rgb\(12,34,56\)/);
  assert.match(result, /&lt;unsafe&gt;/);
  assert.doesNotMatch(result, /\x1b|https:\/\/example|<unsafe>/);
});

test('Java logs retain colors and submitted commands still render as commands', () => {
  const c = setup('paper');
  assert.match(c.formatLogLine('\x1b[32mJava log\x1b[0m'), /color:#10b981/);
  assert.match(c.formatLogLine('> help'), /con-cmd-row/);
  assert.match(setup().formatLogLine('> help'), /con-cmd-row/);
});
