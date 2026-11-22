// Real Chrome, native DOM/DataTransfer/FormData, actual Files renderer; no server,
// downloads, sockets or upload network calls. CDP uses anonymous pipes only.
const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { spawn } = require('node:child_process');
const chrome = process.env.CHROME_PATH || 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe';
const scratch = process.env.PI_SCRATCH_DIR;
const read = name => fs.readFileSync(path.join(__dirname, '../static', name), 'utf8');

test('Chrome: Files drop destination -> immutable queue -> serialized multipart wire', { skip: !scratch || !fs.existsSync(chrome), timeout: 45000 }, async () => {
  const profile = fs.mkdtempSync(path.join(scratch, 'upload-chrome-'));
  const browser = spawn(chrome, ['--headless=new', '--remote-debugging-pipe', '--no-first-run', '--no-default-browser-check', '--disable-background-networking', '--disable-component-update', '--disable-sync', '--disable-extensions', '--disable-features=MediaRouter,OptimizationHints', '--host-resolver-rules=MAP * ~NOTFOUND', '--metrics-recording-only', `--user-data-dir=${profile}`, 'about:blank'], { stdio: ['ignore', 'ignore', 'ignore', 'pipe', 'pipe'] });
  let id = 0, buffer = '';
  const pending = new Map();
  browser.stdio[4].on('data', data => {
    buffer += data.toString();
    let end;
    while ((end = buffer.indexOf('\0')) >= 0) {
      const message = JSON.parse(buffer.slice(0, end)); buffer = buffer.slice(end + 1);
      const call = pending.get(message.id);
      if (call) { pending.delete(message.id); message.error ? call.reject(Error(JSON.stringify(message.error))) : call.resolve(message.result); }
    }
  });
  const send = (method, params = {}, sessionId) => new Promise((resolve, reject) => {
    const next = ++id; pending.set(next, { resolve, reject });
    browser.stdio[3].write(JSON.stringify({ id: next, method, params, sessionId }) + '\0');
  });
  try {
    const { targetId } = await send('Target.createTarget', { url: 'about:blank' });
    const { sessionId } = await send('Target.attachToTarget', { targetId, flatten: true });
    await send('Network.enable', {}, sessionId);
    await send('Network.setBlockedURLs', { urls: ['*'] }, sessionId);
    const app = read('js/app.js');
    const fi = app.slice(app.indexOf('  fi() {'), app.indexOf('\n  plug() {'));
    assert.ok(fi.startsWith('  fi() {') && fi.trimEnd().endsWith('},'));
    const css = Array.from(read('index.html').matchAll(/href="\/(css\/[^\"]+)"/g), match => read(match[1])).join('\n');
    const route = app.slice(app.indexOf('function srvApiPath('), app.indexOf('\nfunction renderLocalNodeAddresses'));
    const sources = { qol: read('js/qol.js'), drag: read('js/file-drag.js'), css, fi, route };
    const result = await send('Runtime.evaluate', { expression: `(${browserChecks.toString()})(${JSON.stringify(sources)})`, awaitPromise: true, returnByValue: true }, sessionId);
    assert.equal(result.exceptionDetails, undefined, JSON.stringify(result.exceptionDetails));
    const report = result.result.value;
    assert.ok(report.checks >= 50, JSON.stringify(report));
    fs.writeFileSync(path.join(scratch, 'upload-chrome-wire.json'), JSON.stringify(report.wire, null, 2));
    fs.writeFileSync(path.join(scratch, 'upload-chrome-results.json'), JSON.stringify(report, null, 2));
    const screenshot = await send('Page.captureScreenshot', { format: 'png' }, sessionId);
    fs.writeFileSync(path.join(scratch, 'upload-drop-hint.png'), Buffer.from(screenshot.data, 'base64'));
    console.log(`Chrome ${report.checks} checks passed; multipart and screenshot in ${scratch}`);
  } finally {
    await send('Browser.close').catch(() => {});
    browser.kill();
    await new Promise(resolve => browser.exitCode !== null ? resolve() : browser.once('exit', resolve));
    fs.rmSync(profile, { recursive: true, force: true, maxRetries: 10, retryDelay: 100 });
  }
});

async function browserChecks(sources) {
  let checks = 0;
  const check = (condition, label) => { if (!condition) throw Error(label); checks++; };
  const sleep = ms => new Promise(resolve => setTimeout(resolve, ms));
  const requests = [], moves = [], loads = [];
  let mode = 'list', denied = false;
  Object.assign(window, {
    S: { user: { id: 1, username: 'owner' }, token: 'token', view: 'fi', currentDir: '', selectedFiles: new Set(), fileList: [] },
    server: { id: 'alpha', node_id: 'local', name: 'Alpha' },
    curServer: () => window.server,
    hasServerPermission: () => !denied,
    srvApiPath: (srv, suffix) => `/api/nodes/${srv.node_id}/servers/${srv.id}${suffix}`,
    esc: value => String(value ?? '').replaceAll('&', '&amp;').replaceAll('"', '&quot;').replaceAll('<', '&lt;'),
    toast: () => {}, srvh: () => '', ic: () => '<span>◈</span>',
    sortFileItems: items => items, getViewMode: () => mode, renderViewModeToggle: () => '',
    fileListIconHtml: () => '<span class="fi-badge">◈</span>', getFileExtension: () => 'txt',
    formatFileSize: () => '1 B', formatFileDate: () => '', renderFileBulkBarHtml: () => '',
    api: async (url, options) => { moves.push({ url, body: JSON.parse(options.body) }); },
    loadFileList: async (dir, _, guard) => { if (!guard || guard.isCurrent()) { loads.push(dir); S.currentDir = dir; render(); } },
    syncFileSelectionUI: () => {},
    XMLHttpRequest: class {
      constructor() { this.upload = {}; }
      open(method, url) { this.method = method; this.url = url; }
      setRequestHeader(name, value) { this[name] = value; }
      send(body) { this.body = body; requests.push(this); }
      finish(status = 200) { this.status = status; this.responseText = '{"error":"conflict"}'; this.onload(); }
      abort() { this.onabort(); }
    }
  });
  document.body.innerHTML = '<main id="main"></main><div id="dropzone-overlay" class="dropzone-overlay" style="display:none"></div>';
  const style = document.createElement('style'); style.textContent = sources.css; document.head.append(style);
  const install = document.createElement('script'); install.textContent = sources.route + '\n' + sources.qol + '\nwindow.uploadQOL = QOL;\n' + sources.drag; document.head.append(install);
  const renderer = new Function('return ({' + sources.fi + '}).fi()');
  function render() {
    S.fileList = [{ name: '世界 folder', path: [S.currentDir, '世界 folder'].filter(Boolean).join('/'), is_dir: true }, { name: 'file.txt', path: [S.currentDir, 'file.txt'].filter(Boolean).join('/'), is_dir: false }];
    document.querySelector('#main').innerHTML = renderer(); NPFileDrag.afterRender();
  }
  function data() { const dt = new DataTransfer(); dt.items.add(new File(['uploaded content'], 'wire.txt', { type: 'text/plain' })); return dt; }
  function event(type, target, dt = data()) { target.dispatchEvent(new DragEvent(type, { bubbles: true, cancelable: true, dataTransfer: dt })); return dt; }
  async function drop(target, expected) {
    const before = requests.length, dir = S.currentDir;
    const dt = event('dragover', target);
    check(document.querySelector('.np-file-drop-label').textContent === `Upload to ${expected || 'root'}`, 'hint destination');
    const hint = document.querySelector('#dropzone-overlay');
    check(getComputedStyle(hint).pointerEvents === 'none' && hint.getBoundingClientRect().height < 85, 'compact pass-through hint');
    event('drop', target, dt); await sleep(0);
    check(requests.length === before + 1, 'one upload');
    const req = requests.at(-1);
    check(req.body instanceof FormData && req.body.get('path') === expected, 'multipart destination');
    check(S.currentDir === dir, 'folder drop must not enter folder');
    check(req.url.endsWith('/files/upload') && !req.url.includes('?'), 'actual body-only path contract');
    req.finish(); await sleep(0);
    return req;
  }
  NPFileDrag.init();
  let wireRequest;
  for (mode of ['list', 'tiles', 'icons']) {
    S.currentDir = 'nested/My space'; render();
    wireRequest = await drop(document.querySelector('[data-is-dir="true"] .fi-badge') || document.querySelector('[data-is-dir="true"]'), 'nested/My space/世界 folder');
    await drop(document.querySelector('.np-files-workspace'), 'nested/My space');
    await drop(document.querySelector('.breadcrumb [data-drop-path="nested"]'), 'nested');
    await drop(document.querySelector('.breadcrumb [data-drop-path=""]'), '');
    S.currentDir = ''; render();
    const before = requests.length; event('drop', document.querySelector('.np-trash-entry')); await sleep(0);
    check(requests.length === before, `${mode} Trash rejected`);
    await drop(document.querySelector('.np-files-workspace'), '');
  }
  // Serialize Chrome's real FormData without sending a byte over the network.
  const serialized = new Request('https://invalid.invalid' + wireRequest.url, { method: 'POST', body: wireRequest.body });
  const wire = { contentType: serialized.headers.get('content-type'), body: await serialized.text(), destination: wireRequest.body.get('path'), filename: 'wire.txt' };
  check(wire.body.includes('name="path"\r\n\r\nnested/My space/世界 folder'), 'Chrome wire contains exact path');
  check(new URL(serialized.url).search === '', 'old handler query path would be empty');
  // Queue and retry retain node/server/path despite unrelated navigation.
  const queued = enqueueUploads([new File(['one'], 'one'), new File(['two'], 'two')], 'queued/世界 space');
  const first = requests.at(-1); S.currentDir = 'different'; window.server = { id: 'beta', node_id: 'remote' }; S.view = 'ov';
  first.finish(); await sleep(0);
  const second = requests.at(-1);
  check(second.url === '/api/servers/alpha/files/upload' && second.body.get('path') === 'queued/世界 space', 'queued navigation does not retarget');
  second.finish(409); await queued;
  const failed = uploadQOL.uploads.at(-1); retryUpload(failed.id);
  check(requests.at(-1).body.get('path') === 'queued/世界 space', 'retry does not retarget'); requests.at(-1).finish(); await sleep(0);
  for (const change of [() => { S.token += 'x'; }, () => { S.user.id++; }, () => { S.user.username += 'x'; }, () => { denied = true; }]) {
    const running = enqueueUploads([new File(['a'], 'a'), new File(['b'], 'b')], 'lease');
    const count = requests.length; change(); requests.at(-1).finish(); await running;
    check(requests.length === count && uploadQOL.uploads.at(-1).status === 'cancelled', 'stale queued identity not sent');
    retryUpload(uploadQOL.uploads.at(-1).id); check(requests.length === count, 'stale retry not sent'); denied = false;
  }
  const running = enqueueUploads([new File(['a'], 'a'), new File(['b'], 'b')], 'cancel');
  const count = requests.length; cancelAllUploads(); await running;
  check(requests.length === count, 'cancelled queue not sent');
  // Real DOM hover-open and internal move destination, unchanged by UI polish.
  S.view = 'fi'; S.currentDir = 'nested'; render();
  for (const change of [() => { S.token += 'x'; }, () => { S.user.id++; }, () => { window.server = { ...server, node_id: server.node_id + 'x' }; }, () => { window.server = { ...server, id: server.id + 'x' }; }, () => NPFileDrag.cancel()]) {
    const count = requests.length, target = document.querySelector('[data-is-dir="true"]');
    const dt = event('dragover', target); change(); event('drop', target, dt); await sleep(0);
    check(requests.length === count, 'stale/cancelled drag must not send');
  }
  for (const item of [{ kind: 'file', webkitGetAsEntry: () => ({ isDirectory: true }) }, { kind: 'file', getAsFileSystemHandle: async () => ({ kind: 'directory' }) }]) {
    const count = requests.length, e = new Event('drop', { bubbles: true, cancelable: true });
    Object.defineProperty(e, 'dataTransfer', { value: { types: ['Files'], files: [new File([], 'folder')], items: [item] } });
    document.querySelector('[data-is-dir="true"]').dispatchEvent(e); await sleep(0);
    check(requests.length === count, 'directory uploads rejected');
  }
  const dt = event('dragover', document.querySelector('[data-is-dir="true"]'));
  await sleep(720); check(S.currentDir === 'nested/世界 folder', 'hover-open destination');
  event('drop', document.querySelector('.np-files-workspace'), dt); await sleep(0);
  check(requests.at(-1).body.get('path') === 'nested/世界 folder', 'hover upload destination'); requests.at(-1).finish(); await sleep(0);
  const internal = new DataTransfer(); event('dragstart', document.querySelector('[data-is-dir="false"]'), internal);
  event('drop', document.querySelector('[data-is-dir="true"]'), internal); await sleep(0);
  check(moves.at(-1).body.destination === 'nested/世界 folder/世界 folder' && moves.at(-1).body.overwrite === false, 'internal move destination and no overwrite');
  // Leave a compact hint for the screenshot artifact, without running its timer.
  event('dragover', document.querySelector('.np-files-workspace'));
  return { checks, modes: ['list', 'tiles', 'icons'], wire };
}
