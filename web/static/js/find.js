/* Local, non-destructive workspace find. Integrators call NPFind.refresh() after render. */
(() => {
  'use strict';
  const MAX_TEXT = 1000000, MAX_NODES = 20000, MAX_MATCHES = 1000, MAX_QUERY = 256;
  const ALL = 'np-find-results', ACTIVE = 'np-find-active';
  const SKIP = new Set(['SCRIPT', 'STYLE', 'NOSCRIPT', 'SELECT', 'OPTION', 'OPTGROUP', 'INPUT', 'TEMPLATE', 'IFRAME', 'CANVAS']);
  let bar, input, count, previous, next, main, priorFocus, observer;
  let query = '', matches = [], active = -1, limited = false, opened = false;
  let timer = null, guard = null, dirty = false, ownedSelection = null, pendingExplicit = false;

  function hidden(el) {
    if (el.hidden || el.inert || el.getAttribute('aria-hidden') === 'true') return true;
    const style = getComputedStyle(el);
    return style.display === 'none' || style.visibility === 'hidden' || style.visibility === 'collapse' ||
      style.contentVisibility === 'hidden' || style.opacity === '0';
  }

  function workspace() {
    const root = document.getElementById('main');
    if (!root || !root.isConnected || document.getElementById('login') || document.getElementById('pal')) return null;
    const modal = document.getElementById('modal-container');
    // NeoPanel mounts dialogs here, and removes their children when dismissed.
    if (modal && modal.childElementCount) return null;
    for (let el = root; el; el = el.parentElement) if (hidden(el)) return null;
    return root;
  }

  function highlights() {
    return window.CSS && CSS.highlights && typeof window.Highlight === 'function' ? CSS.highlights : null;
  }

  function sameRange(a, b) {
    return a && b && a.startContainer === b.startContainer && a.endContainer === b.endContainer &&
      a.startOffset === b.startOffset && a.endOffset === b.endOffset;
  }

  function clearSelection() {
    const selection = window.getSelection();
    if (ownedSelection && selection && selection.rangeCount === 1 && sameRange(selection.getRangeAt(0), ownedSelection)) {
      selection.removeAllRanges();
    }
    ownedSelection = null;
  }

  function clearResults() {
    const registry = highlights();
    if (registry) { registry.delete(ALL); registry.delete(ACTIVE); }
    clearSelection();
    matches = []; active = -1; limited = false;
    if (count) {
      count.textContent = query ? 'Searching…' : 'Type to find'; count.title = '';
      previous.disabled = next.disabled = true;
    }
  }

  function status() {
    count.textContent = (matches.length ? (active >= 0 ? `${active + 1} of ` : '') + matches.length : '0 matches') +
      (limited ? ' (limited)' : '');
    count.title = limited ? 'Search limit reached: first 1,000,000 characters, 20,000 nodes, or 1,000 matches. Narrow your query.' : '';
    previous.disabled = next.disabled = matches.length === 0;
  }

  // RegExp indices remain offsets in the ORIGINAL UTF-16 string, including Unicode case folds.
  // Lowercasing the haystack first would corrupt offsets (for example, dotted capital I).
  function scan() {
    const regex = new RegExp(query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'giu');
    let chars = 0, nodes = 0, text = '', pieces = [], stopped = false;
    function search(value, editor) {
      regex.lastIndex = 0;
      let hit, pieceIndex = 0;
      while ((hit = regex.exec(value))) {
        if (editor) matches.push({ editor, start: hit.index, end: hit.index + hit[0].length });
        else {
          const start = hit.index, end = start + hit[0].length;
          while (pieces[pieceIndex] && pieces[pieceIndex].end <= start) pieceIndex++;
          const first = pieces[pieceIndex];
          let endIndex = pieceIndex;
          while (pieces[endIndex] && pieces[endIndex].end < end) endIndex++;
          const last = pieces[endIndex];
          if (first && last) {
            const range = document.createRange();
            range.setStart(first.node, first.starts ? first.starts[start - first.start] : start - first.start);
            range.setEnd(last.node, last.ends ? last.ends[end - last.start - 1] : end - last.start);
            matches.push({ range });
          }
        }
        if (matches.length >= MAX_MATCHES) { limited = stopped = true; break; }
      }
    }
    function flush() {
      if (text && !stopped) search(text, null);
      text = ''; pieces = [];
    }
    function take(value) {
      const part = value.slice(0, MAX_TEXT - chars);
      chars += part.length;
      if (chars >= MAX_TEXT) limited = true;
      return part;
    }
    function appendText(node, part) {
      if (!part) return;
      let value = part, starts, ends;
      const whitespace = getComputedStyle(node.parentElement).whiteSpace || 'normal';
      if ((whitespace === 'normal' || whitespace === 'nowrap') &&
          (/[\t\r\n\f]| {2}/.test(part) || (text.endsWith(' ') && part.startsWith(' ')))) {
        // Map collapsed rendered whitespace back to original DOM offsets. Pre/code and
        // editor whitespace remains literal; Unicode characters are never case-expanded.
        value = ''; starts = []; ends = [];
        const tokens = /[ \t\r\n\f]+|[^ \t\r\n\f]+/g;
        let token;
        while ((token = tokens.exec(part))) {
          if (/^[ \t\r\n\f]/.test(token[0])) {
            if ((value || text).endsWith(' ')) continue;
            value += ' '; starts.push(token.index); ends.push(token.index + token[0].length);
          } else {
            value += token[0];
            for (let i = 0; i < token[0].length; i++) { starts.push(token.index + i); ends.push(token.index + i + 1); }
          }
        }
      }
      if (value) { pieces.push({ node, start: text.length, end: text.length + value.length, starts, ends }); text += value; }
    }
    // Iterative traversal avoids call-stack overflow and doesn't enumerate an unbounded child list.
    let node = main.firstChild;
    while (node && !stopped) {
      if (++nodes > MAX_NODES || chars >= MAX_TEXT) { limited = true; break; }
      let descend = false;
      if (node.parentElement.tagName === 'DETAILS' && !node.parentElement.open && node.tagName !== 'SUMMARY') {
        flush();
      } else if (node.nodeType === 3) {
        appendText(node, take(node.nodeValue || ''));
      } else if (node.nodeType === 1) {
        const tag = node.tagName;
        if (SKIP.has(tag) || node === bar || hidden(node)) {
          flush();
        } else if (tag === 'TEXTAREA') {
          flush();
          if (!stopped && node.id === 'ed') search(take(node.value), node);
        } else {
          if (boundary(node)) flush();
          descend = tag !== 'BR' && tag !== 'HR';
        }
      }
      if (descend && node.firstChild) { node = node.firstChild; continue; }
      while (node && node !== main) {
        if (node.nodeType === 1 && boundary(node)) flush();
        if (node.nextSibling) { node = node.nextSibling; break; }
        node = node.parentNode;
      }
      if (node === main) break;
    }
    flush();
  }

  function boundary(el) {
    if (el.tagName === 'BR' || el.tagName === 'HR') return true;
    const display = getComputedStyle(el).display;
    return display !== 'inline' && display !== 'contents';
  }

  function paint() {
    const registry = highlights();
    if (!registry) return;
    registry.set(ALL, new Highlight(...matches.filter(m => m.range).map(m => m.range)));
    registry.delete(ACTIVE);
    if (active >= 0 && matches[active].range) registry.set(ACTIVE, new Highlight(matches[active].range));
  }

  function scrollRange(range) {
    let el = range.startContainer.parentElement, steps = 0;
    // Scroll each relevant viewport to the actual text rect (not the top of a large <pre>).
    for (; el && steps++ < 100; el = el.parentElement) {
      if (el.scrollHeight <= el.clientHeight) continue;
      const style = getComputedStyle(el);
      if (!/(auto|scroll|overlay)/.test(style.overflowY)) continue;
      const rect = range.getBoundingClientRect(), box = el.getBoundingClientRect();
      if (rect.top < box.top + 48 || rect.bottom > box.bottom) {
        el.scrollBy({ top: rect.top - box.top - el.clientHeight / 2, behavior: 'instant' });
      }
    }
    const rect = range.getBoundingClientRect();
    if (rect.top < 60 || rect.bottom > window.innerHeight) {
      window.scrollBy({ top: rect.top - window.innerHeight / 2, behavior: 'instant' });
    }
  }

  function reveal(focusEditor) {
    if (active < 0) return;
    const match = matches[active];
    clearSelection();
    if (match.editor) {
      const ed = match.editor;
      if (focusEditor) ed.focus({ preventScroll: true });
      ed.setSelectionRange(match.start, match.end);
      ed.scrollIntoView({ block: 'nearest', behavior: 'instant' });
      const style = getComputedStyle(ed);
      const lineHeight = parseFloat(style.lineHeight) || (parseFloat(style.fontSize) || 14) * 1.4;
      const line = ed.value.slice(0, match.start).split('\n').length - 1;
      ed.scrollTo({ top: Math.max(0, line * lineHeight - ed.clientHeight / 2), behavior: 'instant' });
    } else {
      if (!highlights()) {
        const selection = window.getSelection();
        if (selection) { selection.removeAllRanges(); selection.addRange(match.range); ownedSelection = match.range.cloneRange(); }
      }
      scrollRange(match.range);
    }
  }

  function run(explicit) {
    if (!opened) return;
    const root = workspace();
    if (!root) { shut(false); return; }
    if (root !== main) { bind(root); explicit = false; }
    clearResults(); dirty = false;
    if (!query) return;
    scan();
    if (explicit && matches.length) active = 0;
    paint(); status();
    if (explicit) reveal(false);
  }

  function schedule(explicit, debounce) {
    if (!opened) return;
    if (!workspace()) { shut(false); return; }
    if (!dirty) { clearResults(); dirty = true; }
    pendingExplicit = explicit;
    if (debounce && timer !== null) { clearTimeout(timer); timer = null; }
    // Live mutations use a fixed deadline, not a trailing debounce: logs cannot starve results.
    if (timer === null) timer = setTimeout(() => { timer = null; run(pendingExplicit); }, 100);
  }

  function changed() { schedule(false, false); }

  function bind(root) {
    if (observer) observer.disconnect();
    if (main) main.removeEventListener('input', changed);
    main = root;
    main.addEventListener('input', changed);
    if (typeof MutationObserver === 'function') {
      observer = new MutationObserver(changed);
      observer.observe(main, { subtree: true, childList: true, characterData: true, attributes: true,
        attributeFilter: ['hidden', 'inert', 'aria-hidden', 'style', 'class', 'open', 'value'] });
    }
  }

  function refresh() {
    if (!opened) return;
    const root = workspace();
    if (!root) { shut(false); return; }
    if (root !== main) bind(root);
    // Explicit afterRender invalidates pending user navigation, but preserves the fixed deadline.
    schedule(false, false);
  }

  function move(direction) {
    const root = workspace();
    if (!root) { shut(false); return; }
    if (root !== main) { bind(root); dirty = true; }
    if (dirty) { if (timer !== null) clearTimeout(timer); timer = null; run(false); }
    if (!opened || !matches.length) return;
    active = active < 0 ? (direction > 0 ? 0 : matches.length - 1) : (active + direction + matches.length) % matches.length;
    paint(); status(); reveal(true);
  }

  function build() {
    bar = document.createElement('div'); bar.id = 'np-find';
    bar.setAttribute('role', 'search'); bar.setAttribute('aria-label', 'Find in workspace');
    input = document.createElement('input'); input.type = 'search'; input.id = 'np-find-query';
    input.setAttribute('aria-label', 'Find in workspace'); input.setAttribute('aria-describedby', 'np-find-count');
    input.placeholder = 'Find in workspace'; input.maxLength = MAX_QUERY;
    input.autocomplete = 'off'; input.spellcheck = false;
    count = document.createElement('span'); count.id = 'np-find-count';
    count.setAttribute('role', 'status'); count.setAttribute('aria-live', 'polite'); count.setAttribute('aria-atomic', 'true');
    function button(label, text, action) {
      const el = document.createElement('button'); el.type = 'button'; el.textContent = text;
      el.title = label; el.setAttribute('aria-label', label); el.addEventListener('click', action); return el;
    }
    previous = button('Previous match (Shift+Enter)', '↑', () => move(-1));
    next = button('Next match (Enter)', '↓', () => move(1));
    const close = button('Close find (Escape)', '×', () => shut(true));
    for (const el of [input, count, previous, next, close]) bar.appendChild(el);
    input.addEventListener('input', event => {
      query = input.value.slice(0, MAX_QUERY);
      if (!event.isComposing) schedule(true, true);
      else { clearResults(); dirty = true; if (timer !== null) clearTimeout(timer); timer = null; }
    });
    input.addEventListener('compositionend', () => { query = input.value.slice(0, MAX_QUERY); schedule(true, true); });
  }

  function selectedText(root) {
    const focused = document.activeElement;
    let text = '';
    if (focused && focused.id === 'ed' && focused.tagName === 'TEXTAREA' && root.contains(focused)) {
      if (focused.selectionEnd - focused.selectionStart <= MAX_QUERY) text = focused.value.slice(focused.selectionStart, focused.selectionEnd);
    } else {
      const selection = window.getSelection();
      if (selection && selection.rangeCount && !selection.isCollapsed && !ownedSelection &&
          root.contains(selection.anchorNode) && root.contains(selection.focusNode)) {
        // Read at most a short selection, including inline markup, without allocating
        // Selection.toString() for a potentially enormous selected workspace.
        const range = selection.getRangeAt(0), ancestor = range.commonAncestorContainer;
        const walker = ancestor.nodeType === 3 ? null : document.createTreeWalker(ancestor, 0xFFFFFFFF);
        let node = walker ? walker.nextNode() : ancestor, visited = 0;
        while (node) {
          if (++visited > 1000) return '';
          if (node.nodeType === 3 && range.intersectsNode(node)) {
            const start = node === range.startContainer ? range.startOffset : 0;
            const end = node === range.endContainer ? range.endOffset : node.nodeValue.length;
            if (text.length + end - start > MAX_QUERY) return '';
            for (let el = node.parentElement; el && el !== root; el = el.parentElement) {
              if (SKIP.has(el.tagName) || el.tagName === 'TEXTAREA' || hidden(el)) return '';
            }
            text += node.nodeValue.slice(start, end);
          }
          node = walker ? walker.nextNode() : null;
        }
      }
    }
    return text.length <= MAX_QUERY && text.trim() && !/[\r\n\x00-\x1f]/.test(text) ? text : '';
  }

  function open() {
    const root = workspace();
    if (!root) { if (opened) shut(false); return false; }
    const seed = selectedText(root);
    if (!opened) {
      priorFocus = document.activeElement;
      if (!bar) build();
      document.body.appendChild(bar); opened = true; bind(root);
      // Only #main is mutation-observed. A cheap open-only guard notices sibling overlays/logout.
      guard = setInterval(() => {
        const current = workspace();
        if (!current) shut(false);
        else if (current !== main) refresh();
      }, 250);
    }
    if (seed) query = seed;
    input.value = query;
    refresh(); input.focus({ preventScroll: true }); input.select();
    return true;
  }

  function shut(restore) {
    if (!opened) return;
    opened = false;
    if (timer !== null) clearTimeout(timer);
    if (guard !== null) clearInterval(guard);
    timer = guard = null; dirty = false;
    if (observer) observer.disconnect(); observer = null;
    if (main) main.removeEventListener('input', changed);
    clearResults(); bar.remove();
    if (restore) {
      const target = priorFocus && priorFocus.isConnected ? priorFocus : document.getElementById('main');
      if (target && typeof target.focus === 'function') target.focus({ preventScroll: true });
    }
    main = priorFocus = null;
  }

  function consume(event) { event.preventDefault(); event.stopImmediatePropagation(); }
  document.addEventListener('keydown', event => {
    if (event.isComposing || event.keyCode === 229) return;
    if (opened && event.key === 'Escape') { consume(event); shut(!!workspace()); return; }
    if ((event.ctrlKey || event.metaKey) && !event.altKey && !event.shiftKey && event.key.toLowerCase() === 'f') {
      if (open()) consume(event);
      return;
    }
    if (!opened) return;
    if (!workspace()) { shut(false); return; }
    if (event.altKey || event.ctrlKey || event.metaKey) return;
    const editorMatch = active >= 0 && matches[active].editor;
    const findTarget = event.target === input || (editorMatch && event.target === editorMatch &&
      editorMatch.selectionStart === matches[active].start && editorMatch.selectionEnd === matches[active].end);
    if (event.key === 'F3' || (event.key === 'Enter' && findTarget)) {
      consume(event); move(event.shiftKey ? -1 : 1);
    }
  }, true);
  window.NPFind = Object.freeze({ open, close: () => shut(true), refresh });
})();
