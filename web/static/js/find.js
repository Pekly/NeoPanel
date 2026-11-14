// Find is deliberately scoped to the verified text-file editor, never page DOM.
(() => {
  'use strict';
  const FALLBACK_LIMIT = 128 * 1024;
  let bar, input, count, previous, next, editor, priorFocus, worker;
  let fileKey = '', query = '', snapshot = '', matches = [], active = -1;
  let opened = false, dirty = true, pending = false, workerBusy = false, limited = false, fallback = false;
  let timer = null, deadline = null, guard = null, requestID = 0, wantedDirection = 0;

  function currentEditor() {
    const ed = document.getElementById('ed');
    if (!ed || !ed.isConnected || ed.tagName !== 'TEXTAREA' || !ed.dataset.editorKey ||
        !ed.closest('.np-editor-workspace') || ed.closest('[hidden], [inert], [aria-hidden="true"]') ||
        document.getElementById('login') || document.getElementById('pal') ||
        document.getElementById('modal-container')?.childElementCount) return null;
    if (ed.getClientRects && !ed.getClientRects().length) return null;
    return ed;
  }
  function stopWorker() {
    if (worker) worker.terminate();
    worker = null; workerBusy = false;
    clearTimeout(deadline); deadline = null;
  }
  function status() {
    if (!count) return;
    count.textContent = !query ? 'Type to find' : pending || dirty ? 'Searching…' :
      (matches.length ? `${active >= 0 ? active + 1 + ' of ' : ''}${matches.length}${limited ? '+' : ''}` : '0 matches') + (fallback && limited ? ' (partial)' : '');
    count.title = fallback ? 'Worker unavailable: searched the first 131,072 characters only.' : limited ? 'Showing up to 1,000 matches in the first 5,242,880 characters. Narrow your search.' : '';
    previous.disabled = next.disabled = pending || dirty || !matches.length;
  }
  function reveal(focus) {
    const match = matches[active];
    if (!match || editor !== currentEditor()) return;
    if (focus) editor.focus({ preventScroll: true });
    editor.setSelectionRange(match[0], match[1]);
    const style = getComputedStyle(editor), fontSize = parseFloat(style.fontSize) || 14;
    const lineHeight = parseFloat(style.lineHeight) || fontSize * 1.7;
    const top = match[2] * lineHeight;
    if (top < editor.scrollTop || top + lineHeight > editor.scrollTop + editor.clientHeight) editor.scrollTop = Math.max(0, top - editor.clientHeight / 2);
    const left = match[3] * fontSize * .61;
    if (left < editor.scrollLeft || left > editor.scrollLeft + editor.clientWidth - 30) editor.scrollLeft = Math.max(0, left - editor.clientWidth / 2);
    editor.dispatchEvent(new Event('scroll'));
    editor.scrollIntoView({ block: 'nearest', behavior: 'instant' });
  }
  function accept(id, result, explicit) {
    if (!opened || id !== requestID || currentEditor() !== editor) return;
    clearTimeout(deadline); deadline = null; workerBusy = pending = dirty = false;
    matches = result.matches; limited = Boolean(result.limited) || snapshot.length > NPFileSearch.MAX_TEXT;
    active = matches.length && (explicit || wantedDirection) ? (wantedDirection < 0 ? matches.length - 1 : 0) : -1;
    const focus = Boolean(wantedDirection); wantedDirection = 0;
    status(); if (active >= 0) reveal(focus);
  }
  function run(explicit) {
    timer = null;
    if (!opened) return;
    if (currentEditor() !== editor) { refresh(); return; }
    snapshot = editor.value;
    if (!query) { pending = dirty = false; status(); return; }
    const id = ++requestID;
    pending = true; fallback = false; status();
    const compatibility = () => {
      if (id !== requestID || !opened) return;
      stopWorker(); fallback = true;
      const result = NPFileSearch.search(snapshot.slice(0, FALLBACK_LIMIT), query);
      result.limited = result.limited || snapshot.length > FALLBACK_LIMIT;
      accept(id, result, explicit);
    };
    try {
      if (workerBusy) stopWorker(); // Never queue obsolete full-file searches behind each other.
      if (!worker) worker = new Worker('/js/find-worker.js');
      const current = worker;
      workerBusy = true;
      current.onmessage = ({ data }) => {
        if (current !== worker || data.id !== id || id !== requestID) return;
        if (data.error || !Array.isArray(data.matches)) compatibility();
        else accept(id, data, explicit);
      };
      current.onerror = () => { if (current === worker) compatibility(); };
      deadline = setTimeout(compatibility, 800);
      current.postMessage({ id, text: snapshot.slice(0, NPFileSearch.MAX_TEXT), query });
    } catch (_) { compatibility(); }
  }
  function schedule(explicit = false) {
    ++requestID; clearTimeout(timer); clearTimeout(deadline);
    matches = []; active = -1; dirty = true; pending = false; wantedDirection = 0;
    status(); timer = setTimeout(() => run(explicit), 100);
  }
  function changed() { schedule(false); }
  function bind(ed) {
    editor?.removeEventListener('input', changed);
    editor = ed; fileKey = ed.dataset.editorKey;
    ed.addEventListener('input', changed);
  }
  function refresh() {
    if (!opened) return;
    const ed = currentEditor();
    if (!ed || ed.dataset.editorKey !== fileKey) { close(false); query = ''; return; }
    if (ed !== editor || ed.value !== snapshot) { bind(ed); schedule(false); }
  }
  function move(direction) {
    if (!opened || currentEditor() !== editor) { refresh(); return; }
    if (pending || dirty) {
      if (timer !== null) { clearTimeout(timer); run(false); }
      wantedDirection = direction;
      // A synchronous compatibility search may already have completed.
      if (pending || dirty) return;
      wantedDirection = 0;
    }
    if (!matches.length) return;
    active = active < 0 ? (direction > 0 ? 0 : matches.length - 1) : (active + direction + matches.length) % matches.length;
    status(); reveal(true);
  }
  function build() {
    bar = document.createElement('div'); bar.id = 'np-find'; bar.setAttribute('role', 'search'); bar.setAttribute('aria-label', 'Find in text file');
    input = document.createElement('input'); input.id = 'np-find-query'; input.type = 'search'; input.maxLength = 256;
    input.placeholder = 'Find in file'; input.autocomplete = 'off'; input.spellcheck = false;
    input.setAttribute('aria-label', 'Find in text file'); input.setAttribute('aria-describedby', 'np-find-count');
    count = document.createElement('span'); count.id = 'np-find-count'; count.setAttribute('role', 'status'); count.setAttribute('aria-live', 'polite');
    function button(label, text, action) {
      const b = document.createElement('button'); b.type = 'button'; b.textContent = text; b.title = label; b.setAttribute('aria-label', label); b.addEventListener('click', action); return b;
    }
    previous = button('Previous match (Shift+Enter)', '↑', () => move(-1));
    next = button('Next match (Enter)', '↓', () => move(1));
    const dismiss = button('Close find (Escape)', '×', () => close(true));
    for (const element of [input, count, previous, next, dismiss]) bar.appendChild(element);
    input.addEventListener('input', event => {
      query = input.value.slice(0, 256);
      if (!event.isComposing) schedule(true);
      else { ++requestID; clearTimeout(timer); clearTimeout(deadline); matches = []; active = -1; dirty = true; status(); }
    });
    input.addEventListener('compositionend', () => { query = input.value.slice(0, 256); schedule(true); });
  }
  function open() {
    const ed = currentEditor();
    if (!ed) { if (opened) close(false); return false; }
    const selectedLength = ed.selectionEnd - ed.selectionStart;
    const seed = document.activeElement === ed && selectedLength > 0 && selectedLength <= 256 ? ed.value.slice(ed.selectionStart, ed.selectionEnd) : '';
    if (!opened) {
      priorFocus = document.activeElement;
      if (fileKey !== ed.dataset.editorKey) query = '';
      if (!bar) build();
      document.body.appendChild(bar); opened = true; bind(ed);
      guard = setInterval(() => { if (currentEditor() !== editor) refresh(); }, 250);
    }
    if (seed) query = seed;
    input.value = query;
    schedule(Boolean(seed)); input.focus({ preventScroll: true }); input.select();
    return true;
  }
  function close(restore) {
    if (!opened) return;
    opened = false; ++requestID;
    clearTimeout(timer); clearInterval(guard); timer = guard = null; stopWorker();
    editor?.removeEventListener('input', changed);
    const target = priorFocus?.isConnected ? priorFocus : currentEditor() || document.getElementById('main');
    bar.remove(); matches = []; snapshot = ''; active = -1; pending = false;
    editor = priorFocus = null;
    if (restore) target?.focus({ preventScroll: true });
  }
  function consume(event) { event.preventDefault(); event.stopImmediatePropagation(); }
  document.addEventListener('keydown', event => {
    if (event.isComposing || event.keyCode === 229) return;
    if ((event.ctrlKey || event.metaKey) && !event.altKey && !event.shiftKey && event.key.toLowerCase() === 'f') {
      if (open()) consume(event); // Outside text files, leave browser find untouched.
      return;
    }
    if (!opened) return;
    if (event.key === 'Escape') { consume(event); close(Boolean(currentEditor())); return; }
    if (currentEditor() !== editor) { refresh(); return; }
    if (event.altKey || event.ctrlKey || event.metaKey) return;
    const match = matches[active];
    const selectedMatch = event.target === editor && match && editor.selectionStart === match[0] && editor.selectionEnd === match[1];
    if (event.key === 'F3' || (event.key === 'Enter' && (event.target === input || selectedMatch))) { consume(event); move(event.shiftKey ? -1 : 1); }
  }, true);
  window.NPFind = Object.freeze({ open, close: () => close(true), refresh });
})();
