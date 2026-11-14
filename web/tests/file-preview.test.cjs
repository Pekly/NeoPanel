const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const qol = fs.readFileSync(path.join(__dirname, '../static/js/qol.js'), 'utf8');
const previews = fs.readFileSync(path.join(__dirname, '../static/js/file-preview.js'), 'utf8');
const tick = () => new Promise(resolve => setImmediate(resolve));
const entry = (index, name, is_dir = false) => ({ index, name, is_dir, size: 12, compressed_size: 8 });

function setup(kind = 'image') {
  const server = { id: 'alpha', node_id: 'edge', name: 'Alpha' };
  const elements = {}, storage = new Map(), requests = [], notices = [], opened = [];
  const metadata = { kind, editable: kind === 'text', name: 'example', size: 120, mime: { image: 'image/png', video: 'video/mp4', zip: 'application/zip' }[kind] || 'application/octet-stream' };
  const entries = [entry(0, 'readme.txt'), entry(1, 'config/server.properties'), entry(2, 'textures/icon.png'), entry(3, 'empty/', true)];
  const c = vm.createContext({
    S: { user: { id: 1 }, token: 'session-token', view: 'fi', currentDir: 'media', files: {}, servers: [server] },
    document: { querySelector: selector => elements[selector] || null, body: { contains: () => true } },
    sessionStorage: { getItem: key => storage.get(key), setItem: (key, value) => storage.set(key, value), removeItem: key => storage.delete(key) },
    location: { pathname: '/server/alpha/files', search: '?node=edge' },
    esc: value => String(value ?? '').replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('>', '&gt;').replaceAll('"', '&quot;').replaceAll("'", '&#39;'),
    ic: () => '', formatFileSize: value => `${value} bytes`,
    toast: (message, type) => notices.push([message, type]),
    srvApiPath: (srv, suffix) => `/api/nodes/${srv.node_id}/proxy/api/servers/${srv.id}${suffix}`,
    setTimeout, clearTimeout, uiConfirm: async () => true, _uiModalResolver: null,
    hasServerPermission: () => true,
  });
  c.window = c;
  c.open = (...args) => opened.push(args);
  c.addEventListener = () => {};
  c.curServer = () => c.S.servers[0];
  c.R = () => { c.lastHTML = c.renderFilePreview(); };
  c.api = async (url, options) => {
    requests.push([url, options]);
    if (url.includes('/zip-entry?')) return { name: 'readme.txt', content: '<script>example</script>\nHello' };
    if (url.includes('/zip?')) return { entries };
    if (url.includes('/preview?')) return metadata;
    return { content: 'motd=Hello\n' };
  };
  vm.runInContext(qol + '\n' + previews + '\nthis.qol = QOL; this.preview = FILE_PREVIEW;', c);
  c.saveActiveFile = c.saveQolFile;
  return { c, server, metadata, entries, elements, requests, notices, opened };
}

test('images open inline through the correct authenticated node, never the editor', async () => {
  const { c, requests, opened } = setup('image');
  await c.openQolFile('a & b.png');
  assert.equal(c.qol.editor, null);
  assert.equal(requests.length, 1);
  assert.match(requests[0][0], /\/files\/preview\?path=media%2Fa%20%26%20b.png/);
  assert.match(c.lastHTML, /<img id="file-preview-media"/);
  assert.match(c.lastHTML, /\/api\/nodes\/edge\/proxy\/api\/servers\/alpha\/files\/media/);
  assert.match(c.lastHTML, /token=session-token/);
  assert.doesNotMatch(c.lastHTML, /<textarea|editor-save-btn/);
  c.downloadPreviewFile();
  assert.match(opened[0][0], /\/files\/download\?/);
  assert.equal(opened[0][2], 'noopener,noreferrer');
});

test('videos have controls without autoplay and release playback on close', async () => {
  const { c, elements } = setup('video');
  await c.openQolFile('clip.mp4');
  assert.match(c.lastHTML, /<video[^>]+controls playsinline preload="metadata"/);
  assert.doesNotMatch(c.lastHTML, /autoplay|<textarea/);
  let paused = false, removed = false, reset = false;
  elements['#file-preview-media'] = { pause() { paused = true; }, removeAttribute(name) { removed = name === 'src'; }, load() { reset = true; } };
  c.closeFilePreview();
  assert.equal(c.preview.current, null);
  assert.equal(paused && removed && reset, true);
});

test('binary files and oversized text never receive editable state or a content request', async () => {
  for (const kind of ['binary', 'text']) {
    const { c, metadata, requests } = setup(kind);
    metadata.editable = false;
    metadata.reason = 'Cannot edit this file';
    await c.openQolFile('server.dat');
    assert.equal(c.qol.editor, null);
    assert.equal(c.S.editingFile, null);
    assert.equal(requests.length, 1);
    assert.match(c.lastHTML, /Text editing is unavailable/);
    assert.match(c.lastHTML, /Cannot edit this file/);
    assert.doesNotMatch(c.lastHTML, /<textarea|editor-save-btn/);
  }
});

test('text files retain the existing editor workflow after authoritative validation', async () => {
  const { c, requests } = setup('text');
  await c.openQolFile('server.properties');
  assert.equal(c.qol.editor.content, 'motd=Hello\n');
  assert.equal(c.S.editingFile, 'media/server.properties');
  assert.equal(c.preview.current, null);
  assert.equal(requests.length, 2);
  assert.match(requests[1][0], /\/files\/content\?/);
});

test('invalid metadata or binary content cannot fall back to the text editor', async () => {
  for (const response of [{}, { content: 'binary\u0000data' }]) {
    const { c, notices } = setup('text');
    c.api = async url => url.includes('/preview?') && response.content ? { kind: 'text', editable: true } : response;
    await c.openQolFile('disguised.txt');
    assert.equal(c.qol.editor, null);
    assert.equal(notices.at(-1)[1], 'bad');
  }
});

test('ZIP browsing synthesizes folders, filters names and never extracts files', async () => {
  const { c, requests } = setup('zip');
  await c.openQolFile('world.zip');
  assert.match(c.lastHTML, /ZIP contents/);
  assert.match(c.lastHTML, /Nothing has been extracted or changed/);
  assert.deepEqual(Array.from(c.zipPreviewRows(c.preview.current.entries), row => row.name), ['config', 'empty', 'textures', 'readme.txt']);
  c.browseZipPreview('config/');
  assert.equal(c.preview.current.folder, 'config/');
  assert.match(c.lastHTML, /server.properties/);
  assert.doesNotMatch(c.lastHTML, /readme.txt/);
  assert.equal(c.zipPreviewRows(c.preview.current.entries, '', 'texture').length, 1);
  assert.ok(requests.every(([url, options]) => !url.includes('/extract') && !url.includes('/unzip') && !options?.method));
});

test('ZIP text entries render escaped read-only content, not an editable buffer', async () => {
  const { c, requests } = setup('zip');
  await c.openQolFile('archive.jar');
  await c.openZipPreviewEntry(0);
  assert.match(requests.at(-1)[0], /\/files\/zip-entry\?path=media%2Farchive.jar&index=0/);
  assert.match(c.lastHTML, /&lt;script&gt;example&lt;\/script&gt;/);
  assert.doesNotMatch(c.lastHTML, /<script>|<textarea|editor-save-btn/);
  assert.equal(c.qol.editor, null);
});

test('binary ZIP entries and corrupt archives show a recoverable error', async () => {
  const { c } = setup('zip');
  await c.openQolFile('world.zip');
  c.api = async () => { throw Error('Binary entries cannot be opened as text'); };
  await c.openZipPreviewEntry(2);
  assert.match(c.lastHTML, /Binary entries cannot be opened as text/);
  assert.match(c.lastHTML, /Download archive/);
  assert.equal(c.qol.editor, null);
  await c.openFilePreview(c.curServer(), 'bad.zip', { kind: 'zip' });
  assert.match(c.lastHTML, /role="alert"/);
});

test('file and ZIP responses cannot cross server, session or preview boundaries', async () => {
  const { c } = setup('image');
  let finish;
  c.api = () => new Promise(resolve => { finish = resolve; });
  const pending = c.openQolFile('old.png'); await tick();
  c.S.servers = [{ id: 'other', node_id: 'west' }];
  finish({ kind: 'image' }); await pending;
  assert.equal(c.preview.current, null);
  const zipped = setup('zip');
  await zipped.c.openQolFile('world.zip');
  zipped.c.api = () => new Promise(resolve => { finish = resolve; });
  const entryPending = zipped.c.openZipPreviewEntry(0);
  zipped.c.closeFilePreview();
  finish({ content: 'late entry' }); await entryPending;
  assert.equal(zipped.c.preview.current, null);
  const session = setup('image');
  await session.c.openQolFile('picture.png');
  session.c.S.token = 'another-session';
  assert.equal(session.c.renderFilePreview(), '');
});

test('unsaved text cancellation prevents opening a preview', async () => {
  const { c, requests } = setup('image');
  c.qol.editor = { path: 'dirty.txt', key: 'draft', base: '', content: 'keep me', server: c.curServer() };
  c.editorLeaveChoice = async () => 'cancel';
  await c.openQolFile('image.png');
  assert.equal(requests.length, 0);
  assert.equal(c.qol.editor.content, 'keep me');
  assert.equal(c.preview.current, null);
});

test('media failures show download fallback and filenames are safely escaped', async () => {
  const { c, metadata, elements } = setup('image');
  metadata.name = '\" onerror=\"alert(1)';
  await c.openQolFile('<unsafe>.png');
  assert.doesNotMatch(c.lastHTML, /alt="" onerror="alert|<unsafe>/);
  const media = {}; elements['#file-preview-media'] = media;
  c.fileMediaFailed(media);
  assert.match(c.lastHTML, /codec your browser does not support/);
  assert.doesNotMatch(c.lastHTML, /<img id=/);
});

test('rerenders retain the original video element and playback state', async () => {
  const { c, elements } = setup('video');
  await c.openQolFile('clip.mp4');
  const video = { tagName: 'VIDEO', paused: false, ended: false, currentTime: 42, volume: 0.4, playbackRate: 1.5, readyState: 2 };
  elements['#file-preview-media'] = video;
  elements['#file-preview-loading'] = { hidden: false };
  const retained = c.captureFilePreviewMedia();
  const html = c.renderFilePreview();
  assert.doesNotMatch(html, /<video[^>]+src=/, 'replacement must not start another media request');
  elements['#file-preview-media'] = { replaceWith(element) { elements['#file-preview-media'] = element; } };
  c.restoreFilePreviewMedia(retained);
  assert.equal(elements['#file-preview-media'], video);
  assert.equal(video.currentTime, 42); assert.equal(video.volume, 0.4); assert.equal(video.playbackRate, 1.5);
  assert.equal(video.paused, false); assert.equal(elements['#file-preview-loading'].hidden, true);
  assert.equal(c.preview.retainedMedia, null);
});
