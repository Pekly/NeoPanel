/* Progressive enhancement for #ed. No file contents, selections or edit events are modified.
 * Integration: bind(ed, filename) after value/selection; refresh() after programmatic edits;
 * dispose() BEFORE replacing editor DOM, closing, or logout. bind returns the same scoped hooks.
 */
(function (global) {
  'use strict';
  const doc = global.document;
  const workerURL = new URL('./editor-syntax-worker.js', doc.currentScript?.src || new URL('/static/js/editor-syntax.js', doc.baseURI)).href;
  const MAX_CHARS = 200000, MAX_RUNS = 6000, DEBOUNCE_MS = 120, TIMEOUT_MS = 2000;
  const extensions = Object.freeze({
    json: 'json', mcmeta: 'json', jsonc: 'javascript', yaml: 'yaml', yml: 'yaml', toml: 'toml',
    ini: 'ini', cfg: 'ini', conf: 'ini', properties: 'properties', lang: 'properties',
    js: 'javascript', mjs: 'javascript', cjs: 'javascript', ts: 'typescript', mts: 'typescript', cts: 'typescript',
    java: 'java', py: 'python', pyw: 'python', sh: 'bash', bash: 'bash', zsh: 'bash',
    ps1: 'powershell', psm1: 'powershell', psd1: 'powershell', sql: 'sql',
    xml: 'markup', html: 'markup', htm: 'markup', xhtml: 'markup', svg: 'markup',
    css: 'css', md: 'markdown', markdown: 'markdown', go: 'go', rs: 'rust', lua: 'lua', dockerfile: 'docker'
  });
  const labels = Object.freeze({ json: 'JSON', yaml: 'YAML', toml: 'TOML', ini: 'INI / config',
    properties: 'Properties', javascript: 'JavaScript', typescript: 'TypeScript', java: 'Java',
    python: 'Python', bash: 'Shell', powershell: 'PowerShell', sql: 'SQL', markup: 'XML / HTML',
    css: 'CSS', markdown: 'Markdown', go: 'Go', rust: 'Rust', lua: 'Lua', docker: 'Dockerfile', plain: 'Plain text' });
  const kinds = new Set(['', 'comment', 'string', 'number', 'keyword', 'property', 'tag',
    'function', 'operator', 'punctuation', 'variable']);
  let active = null, nextGeneration = 0;

  function languageFor(filename) {
    const name = String(filename || '').split(/[\\/]/).pop().toLowerCase();
    if (/^dockerfile(?:\.|$)/.test(name)) return 'docker';
    if (/^\.env(?:\.|$)/.test(name)) return 'properties';
    if (['.bashrc', '.bash_profile', '.zshrc'].includes(name)) return 'bash';
    const extension = name.includes('.') ? name.split('.').pop() : '';
    return Object.prototype.hasOwnProperty.call(extensions, extension) ? extensions[extension] : 'plain';
  }

  function bind(ed, filename) {
    active?.dispose();
    if (!ed || ed.tagName !== 'TEXTAREA' || !ed.parentNode) return null;
    const generation = ++nextGeneration, language = languageFor(filename);
    const parent = ed.parentNode, oldWrap = ed.getAttribute('wrap');
    const focused = doc.activeElement === ed, top = ed.scrollTop, left = ed.scrollLeft;
    const wrapper = doc.createElement('div'), mirror = doc.createElement('pre'), status = doc.createElement('div');
    // #ed stays the only editable source; NPMotion already skips #ed / .np-editor-workspace.
    wrapper.className = 'np-syntax-editor';
    mirror.className = 'np-syntax-mirror';
    mirror.setAttribute('aria-hidden', 'true');
    status.className = 'np-syntax-status';
    status.setAttribute('role', 'status');
    status.setAttribute('aria-live', 'polite');
    parent.insertBefore(wrapper, ed);
    wrapper.appendChild(mirror);
    wrapper.appendChild(ed);
    parent.insertBefore(status, wrapper.nextSibling);
    ed.classList.add('np-syntax-input');
    ed.setAttribute('wrap', 'off');
    if (focused) ed.focus({ preventScroll: true });
    ed.scrollTop = top; ed.scrollLeft = left;
    const forced = global.matchMedia?.('(forced-colors: active)');
    let disposed = false, composing = false, failed = '', version = 0;
    let worker = null, inflight = null, debounce = null, watchdog = null, ready = false;
    const styled = global.getComputedStyle(mirror).getPropertyValue('--np-syntax-ready').trim() === '1';

    function message(detail) {
      const text = labels[language] + (detail ? ' · ' + detail : '');
      if (status.textContent !== text) status.textContent = text;
    }
    function plain() {
      wrapper.classList.remove('np-syntax-colored');
      mirror.textContent = '';
    }
    function sync() {
      if (disposed) return;
      // Use the native text viewport, excluding both scrollbars (including after resize).
      mirror.style.width = ed.clientWidth + 'px';
      mirror.style.height = ed.clientHeight + 'px';
      mirror.scrollTop = ed.scrollTop;
      mirror.scrollLeft = ed.scrollLeft;
    }
    function stopWorker() {
      global.clearTimeout(watchdog); watchdog = null;
      if (worker) {
        worker.onmessage = worker.onerror = worker.onmessageerror = null;
        worker.terminate(); worker = null;
      }
      inflight = null;
    }
    function fail(reason) {
      if (disposed) return;
      failed = reason; ready = false;
      global.clearTimeout(debounce); debounce = null;
      stopWorker(); plain(); message('Plain text — ' + reason);
    }
    function render(runs, text) {
      if (!Array.isArray(runs) || runs.length > MAX_RUNS) throw new Error('invalid runs');
      const fragment = doc.createDocumentFragment();
      let start = 0;
      for (const run of runs) {
        if (!Array.isArray(run) || run.length !== 2 || !Number.isSafeInteger(run[0]) ||
            run[0] <= start || run[0] > text.length || !kinds.has(run[1])) throw new Error('invalid run');
        const node = doc.createElement('span');
        node.className = run[1] ? 'np-syntax-' + run[1] : '';
        node.textContent = text.slice(start, run[0]);
        fragment.appendChild(node);
        start = run[0];
      }
      if (start !== text.length) throw new Error('incomplete runs');
      // A zero-width sentinel gives empty/trailing blank lines their native line box.
      fragment.appendChild(doc.createTextNode('\u200b'));
      mirror.replaceChildren(fragment);
      sync();
      wrapper.classList.add('np-syntax-colored');
    }
    function request() {
      if (disposed || !ready || inflight || composing || failed) return;
      ready = false;
      const text = ed.value;
      if (text.length > MAX_CHARS) { refresh(); return; }
      try {
        if (!worker) {
          worker = new global.Worker(workerURL);
          worker.onerror = worker.onmessageerror = () => fail('coloring unavailable');
          worker.onmessage = event => {
            if (disposed || !inflight) return;
            const data = event.data || {}, job = inflight;
            if (data.generation !== generation || data.version !== job.version) return;
            global.clearTimeout(watchdog); watchdog = null; inflight = null;
            if (job.version === version && job.text === ed.value && !composing) {
              if (data.error) { fail('coloring complexity limit'); return; }
              try { render(data.runs, job.text); }
              catch (_) { fail('invalid coloring result'); return; }
            }
            request();
          };
        }
        inflight = { version, text };
        // Includes worker startup/import time. Never reset this deadline on further input.
        watchdog = global.setTimeout(() => fail('coloring timed out'), TIMEOUT_MS);
        worker.postMessage({ generation, version, language, text });
      } catch (_) { fail('coloring unavailable'); }
    }
    function refresh() {
      if (disposed) return;
      version++; ready = false;
      global.clearTimeout(debounce); debounce = null;
      // Never show out-of-date tokens under current text, even during IME composition.
      plain(); sync();
      if (ed.value.length > MAX_CHARS) {
        stopWorker(); message('Plain text — coloring disabled above 200,000 characters'); return;
      }
      if (language === 'plain') { message('No syntax coloring'); return; }
      if (forced?.matches) { stopWorker(); message('Plain text — high contrast'); return; }
      if (!styled) { message('Plain text — syntax stylesheet unavailable'); return; }
      if (failed) { message('Plain text — ' + failed); return; }
      message('Syntax');
      if (!composing) debounce = global.setTimeout(() => { debounce = null; ready = true; request(); }, DEBOUNCE_MS);
    }
    function compositionStart() { composing = true; refresh(); }
    function compositionEnd() { composing = false; refresh(); }
    function dispose() {
      if (disposed) return;
      disposed = true;
      global.clearTimeout(debounce);
      stopWorker(); observer?.disconnect();
      forced?.removeEventListener?.('change', refresh);
      ed.removeEventListener('input', refresh);
      ed.removeEventListener('scroll', sync);
      ed.removeEventListener('compositionstart', compositionStart);
      ed.removeEventListener('compositionend', compositionEnd);
      ed.classList.remove('np-syntax-input');
      if (oldWrap === null) ed.removeAttribute('wrap'); else ed.setAttribute('wrap', oldWrap);
      const wasFocused = doc.activeElement === ed, y = ed.scrollTop, x = ed.scrollLeft;
      if (ed.parentNode === wrapper && wrapper.parentNode) wrapper.parentNode.insertBefore(ed, wrapper);
      wrapper.remove(); status.remove();
      if (wasFocused) ed.focus({ preventScroll: true });
      ed.scrollTop = y; ed.scrollLeft = x;
      if (active === controller) active = null;
    }
    const observer = typeof global.ResizeObserver === 'function' ? new global.ResizeObserver(sync) : null;
    observer?.observe(ed);
    ed.addEventListener('input', refresh);
    ed.addEventListener('scroll', sync, { passive: true });
    ed.addEventListener('compositionstart', compositionStart);
    ed.addEventListener('compositionend', compositionEnd);
    forced?.addEventListener?.('change', refresh);
    const controller = Object.freeze({ refresh, dispose });
    active = controller;
    refresh();
    return controller;
  }
  global.NPEditorSyntax = Object.freeze({ bind, languageFor,
    refresh: () => active?.refresh(), dispose: () => active?.dispose(),
    limits: Object.freeze({ maxChars: MAX_CHARS, maxRuns: MAX_RUNS, debounceMs: DEBOUNCE_MS, timeoutMs: TIMEOUT_MS })
  });
})(window);
