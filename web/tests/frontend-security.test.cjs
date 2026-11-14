const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const app = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
const automation = fs.readFileSync(path.join(__dirname, '../static/js/automation.js'), 'utf8');
const payloads = ["x');globalThis.__securityProbe=1;//", "x');globalThis.__securityProbe=1;//\\\n\"<img src=x onerror=globalThis.__securityProbe=1>&#39;"];
payloads.push("x');globalThis.__securityProbe=1;('config.txt");
function section(start, end) {
  const a = app.indexOf(start), b = app.indexOf(end, a);
  assert.ok(a >= 0 && b > a, start);
  return app.slice(a, b);
}
// Model the HTML parser's one entity-decoding pass before compiling an inline handler.
function decode(value) {
  return value.replace(/&(amp|lt|gt|quot|#39|#(\d+)|#x([0-9a-f]+));/gi, (all, named, dec, hex) => {
    if (dec || hex) return String.fromCodePoint(parseInt(dec || hex, hex ? 16 : 10));
    return { amp: '&', lt: '<', gt: '>', quot: '"', '#39': "'" }[named.toLowerCase()];
  });
}
function elements(html) {
  const result = [];
  for (const tag of html.matchAll(/<([a-z][\w-]*)\b((?:"[^"]*"|'[^']*'|[^'">])*)>/gi)) {
    const attrs = {}, dataset = {};
    for (const attr of tag[2].matchAll(/([\w-]+)\s*=\s*"([^"]*)"/g)) {
      assert.equal(Object.hasOwn(attrs, attr[1]), false, `duplicate ${attr[1]}`);
      attrs[attr[1]] = decode(attr[2]);
      if (attr[1].startsWith('data-')) dataset[attr[1].slice(5).replace(/-([a-z])/g, (_, c) => c.toUpperCase())] = attrs[attr[1]];
    }
    result.push({ tag: tag[1], attrs, dataset, checked: true, files: ['file'] });
  }
  assert.equal(result.some(el => /^(img|svg|script)$/i.test(el.tag)), false, 'payload must not inject a node');
  return result;
}
function invoke(c, element, eventName = 'onclick') {
  c.element = element;
  assert.doesNotMatch(element.attrs[eventName], /__securityProbe/, 'untrusted text must not enter handler source');
  vm.runInContext(`(function(event) { ${element.attrs[eventName]} }).call(element, event)`, c);
  assert.equal(c.__securityProbe, undefined);
}
function setup(mode = 'list') {
  const c = vm.createContext({
    S: { fileList: [], files: {}, selectedFiles: new Set(), players: [], wl: [], bans: [], bk: [] },
    ic: () => '', head: () => '', srvh: () => '', getViewMode: () => mode,
    hasServerPermission: () => true,
    renderViewModeToggle: () => '', renderRestoreNotice: () => '', renderFileBulkBarHtml: () => '',
    sortFileItems: list => list, fileListIconHtml: () => '', getFileExtension: () => 'txt',
    formatFileSize: () => '1 B', formatFileDate: () => 'Today', formatBytes: () => '1 B',
    event: { key: 'Enter', stopPropagation() {}, preventDefault() {} }, calls: []
  });
  vm.runInContext(section('const esc =', 'const DEFAULT_USER_AVATAR_URL') + `\nthis.views = ({${section('\n  pl() {', '\n  plug() {')}${section('\n  bk() {', '\n  sc() {')}});`, c);
  for (const name of ['openFile', 'browseFolder', 'loadFileList', 'toggleFileSelect', 'downloadFileItem', 'playerAction', 'downloadBackup', 'deleteBackup', 'restoreBackup']) {
    c[name] = (...args) => c.calls.push([name, ...args]);
  }
  return c;
}
for (const mode of ['list', 'icons', 'tiles']) {
  test(`file ${mode} handlers preserve filenames, selection paths, and breadcrumb paths as data`, () => {
    for (const payload of payloads) {
      const c = setup(mode);
      c.S.currentDir = `parent/${payload}`;
      c.S.fileList = [false, true].map(is_dir => ({ name: payload, path: `${c.S.currentDir}/${payload}`, is_dir }));
      const rendered = elements(c.views.fi());
      const open = rendered.filter(el => el.attrs.onclick?.match(/^(?:event.stopPropagation\(\); )?(openFile|browseFolder)\(this.dataset.name\)/));
      assert.equal(open.length, mode === 'list' ? 4 : 2);
      for (const el of open) { invoke(c, el); assert.equal(c.calls.at(-1)[1], payload); }
      assert.ok(c.calls.some(([name]) => name === 'openFile'));
      assert.ok(c.calls.some(([name]) => name === 'browseFolder'));
      const checks = rendered.filter(el => el.attrs.onchange?.startsWith('toggleFileSelect('));
      assert.equal(checks.length, 2);
      for (const el of checks) { invoke(c, el, 'onchange'); assert.deepEqual(c.calls.at(-1).slice(1), [`${c.S.currentDir}/${payload}`, true]); }
      const crumbs = rendered.filter(el => el.attrs.onclick === 'loadFileList(this.dataset.path)');
      for (const el of crumbs) { invoke(c, el); assert.equal(c.calls.at(-1)[1], el.dataset.path); }
      assert.equal(crumbs.at(-1).dataset.path, c.S.currentDir.split('/').filter(Boolean).join('/'));
      c.S.editingFile = payload;
      const download = elements(c.views.fi()).find(el => el.attrs.onclick?.startsWith('downloadFileItem('));
      invoke(c, download); assert.equal(c.calls.at(-1)[1], payload);
    }
  });
  test(`player and backup ${mode} controls cannot compile names as JavaScript`, () => {
    for (const payload of payloads) {
      const c = setup(mode);
      c.S.players = [{ name: payload, op: true }]; c.S.wl = [{ name: payload }]; c.S.bans = [{ name: payload }];
      for (const tab of ['online', 'wl', 'ban']) {
        c.S.ptab = tab;
        const controls = elements(c.views.pl()).filter(el => el.attrs.onclick?.startsWith('playerAction('));
        assert.equal(controls.length, tab === 'online' ? 3 : 1);
        for (const el of controls) { invoke(c, el); assert.equal(c.calls.at(-1)[2], payload); }
      }
      c.S.bk = [{ name: payload, size: 1 }];
      const controls = elements(c.views.bk()).filter(el => /^(downloadBackup|restoreBackup|deleteBackup)\(/.test(el.attrs.onclick || ''));
      assert.equal(controls.length, 3);
      for (const el of controls) { invoke(c, el); assert.equal(c.calls.at(-1)[1], payload); }
    }
  });
}

test('data-based handlers across database, node, mod, schedule, settings and account templates round-trip strings', () => {
  const handlers = /(?:exportServerDatabase|installJavaRuntime|loadServerSchedules|toggleServerSchedule|runScheduleNow|openScheduleModal|deleteServerSchedule|toggleDbPasswordReveal|openSQLStudioModal|copyDbText|serverRouteUrl|uploadServerIcon|deleteServerIcon|openQuickConfigsModal|testClusterNode|openNodeModal|deleteSSHKey|deleteAPIKey|handleNavClick|_execSidebarItem|submitNodeModal|testNodeModalConnection|closeUiModal|clipboard.writeText|confirmEnable2FA|setCreateEngine|openSubuserModal|deleteServerSubuser|submitSubuserModal|saveQuickConfigs|submitAdminServerSettings|selectAdminMod|toggleAdminModEnabled|deleteAdminMod|submit2FALogin|submitScheduleModal)\(/;
  let tested = 0;
  for (const line of (app + '\n' + automation).split('\n')) {
    const match = line.match(/\bdata-[\w-]+=".*?\b(onclick|onchange|onsubmit|onkeydown)="([^"]*)"/);
    if (!match || !handlers.test(match[2]) || !match[2].includes('this.dataset.')) continue;
    if (match[2].startsWith('copyDbText(') && !match[2].includes('this.dataset.label')) continue;
    for (const payload of payloads) {
      const c = setup();
      const record = { id: payload, name: payload, node_id: payload, version: payload, value: payload, type: payload, secret: payload, pageId: payload };
      Object.assign(c, { db: record, d: record, item: record, srv: record, sch: record, rec: record, n: record, k: record, it: record, existing: record, res: record, sub: record, m: record, draft: record,
        nid: payload, recJava: payload, slug: payload, targetId: payload, targetNodeId: payload, tempToken: payload, codes: [payload, payload], isEdit: true, i: 0,
        AutomationUI: { scheduleID: payload }, document: { getElementById: () => ({ checked: true }) }, toast: () => {} });
      c.window = c;
      c.NP = { _execSidebarItem: (...args) => c.calls.push(args) };
      c.navigator = { clipboard: { writeText: (...args) => c.calls.push(args) } };
      for (const name of match[2].matchAll(/\b([A-Za-z]\w*)\(/g)) {
        if (['Boolean', 'if', 'getElementById', 'writeText'].includes(name[1])) continue;
        c[name[1]] = (...args) => { c.calls.push(args); return 'route'; };
      }
      const html = vm.runInContext('`<button ' + match[0] + '>`', c);
      const el = elements(html)[0];
      invoke(c, el, match[1]);
      assert.ok(c.calls.length, match[0]);
      assert.ok(c.calls.some(args => JSON.stringify(args).includes(JSON.stringify(payload).slice(1, -1))), match[0]);
    }
    tested++;
  }
  assert.ok(tested >= 45, `expected coverage of all converted secondary surfaces, got ${tested}`);
});

test('owned templates never use HTML escaping as inline JavaScript string encoding', () => {
  for (const source of [app, automation]) {
    for (const handler of source.matchAll(/\bon(?:click|change|submit|contextmenu|keydown|input)="([^"]*)"/g)) {
      assert.doesNotMatch(handler[1], /\$\{[^}]*\besc\(/, handler[0]);
    }
  }
});
