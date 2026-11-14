/* Native file-manager drag/drop. Parent calls init(), afterRender(root), cancel().
 * loadFileList(path, false, { isCurrent }) must check isCurrent before committing
 * async results (including finally/render). enqueueUploads(files, destination)
 * must retain the supplied destination. Remove the legacy drag/drop listeners.
 */
(() => {
  'use strict';
  const TYPE = 'application/x-neopanel-files';
  const ROW = '[data-file-drag][data-path]';
  const BLOCKED = 'input,button,select,textarea,a,[contenteditable]:not([contenteditable="false"]),[role="button"]';
  let initialized = false, active = null, pending = null, epoch = 0;
  let hover = null, timer = null, watchdog = null, highlight = null, hint = null;
  let blockedOrigin = false, rejectedExternal = false;

  function path(value) {
    if (typeof value !== 'string' || /[\\\x00-\x1f]/.test(value) || value.startsWith('/')) return null;
    const parts = value.split('/').filter(Boolean);
    if (parts.some(p => p === '.' || p === '..')) return null;
    return parts.join('/');
  }
  function identity() { return JSON.stringify([S.user?.id, S.user?.username]); }
  function allowed(internal) {
    if (typeof S === 'undefined' || !S.token || !S.user || S.view !== 'fi' || S.fileTrashOpen) return false;
    if (document.querySelector('#ed, #login, #pal, .pal, [aria-modal="true"], dialog[open]') ||
        (typeof _uiModalResolver !== 'undefined' && _uiModalResolver) ||
        (typeof filePreviewActive === 'function' && filePreviewActive())) return false;
    const server = curServer();
    return Boolean(server && hasServerPermission('files.view', server) && hasServerPermission('files.edit', server) &&
      (!internal || hasServerPermission('files.delete', server)));
  }
  function capture(internal) {
    return { internal, server: { ...curServer() }, token: S.token, user: identity(),
      dir: path(S.currentDir || ''), epoch, navigation: null, source: null, parking: null, marker: null };
  }
  function current(state) {
    return Boolean(state && state.epoch === epoch && allowed(state.internal) && state.token === S.token &&
      state.user === identity() && sameQolServer(state.server, curServer()));
  }
  function locationCurrent(state) {
    const dir = path(S.currentDir || '');
    return dir === state.dir || Boolean(state.navigation && dir === state.navigation.to);
  }
  function valid(state) { return current(state) && locationCurrent(state); }
  function filesDrag(dt) {
    return Boolean(dt && (Array.from(dt.types || []).includes('Files') ||
      Array.from(dt.items || []).some(item => item.kind === 'file') || dt.files?.length));
  }
  function relevant(event) {
    return Boolean(active?.internal || filesDrag(event.dataTransfer) || Array.from(event.dataTransfer?.types || []).includes(TYPE));
  }
  function element(target) { return target?.closest ? target : target?.parentElement; }
  function blockedControl(target) {
    const control = element(target)?.closest(BLOCKED);
    return Boolean(control && !control.matches('.np-file-open'));
  }
  function targetAt(target) {
    const el = element(target);
    // Breadcrumbs may sit next to (rather than inside) the collection.
    const explicit = el?.closest('[data-drop-path]');
    if (explicit && explicit.closest('#main') && document.querySelector('.np-files-workspace')) {
      const destination = path(explicit.dataset.dropPath);
      return destination === null ? null : { el: explicit, destination, folder: true };
    }
    const workspace = el?.closest('.np-files-workspace');
    const destination = path(S.currentDir || '');
    return workspace && destination !== null ? { el: workspace, destination, folder: false } : null;
  }
  function moveError(state, destination) {
    if (!state.paths.length || state.paths.length > 1000) return 'Move between 1 and 1000 items at a time.';
    for (const source of state.paths) {
      if (destination === source || destination.startsWith(source + '/')) return 'Cannot move an item into itself or its descendants.';
      if (source.split('/').slice(0, -1).join('/') === destination) return 'These items are already in that folder.';
    }
    return '';
  }
  function clearTimer() { if (timer !== null) clearTimeout(timer); timer = null; }
  function clearVisuals() {
    highlight?.classList.remove('np-file-drop-target'); highlight = null;
    if (hint) hint.style.display = 'none';
  }
  function stopHover() { clearTimer(); hover = null; }
  function cleanSource(state) {
    if (!state) return;
    state.source?.classList.remove('np-file-dragging');
    if (state.parking) {
      if (state.marker?.isConnected) state.marker.replaceWith(state.source);
      else state.source?.remove();
      state.parking.remove();
    }
    state.marker?.remove();
  }
  function endDrag() {
    stopHover(); clearVisuals(); cleanSource(active); active = null;
    if (watchdog !== null) clearInterval(watchdog);
    watchdog = null;
  }
  function cancel() {
    if (active && !active.internal) rejectedExternal = true;
    ++epoch; endDrag();
    // An already submitted POST cannot be undone; keep its lock until settled.
  }
  function watch() {
    if (watchdog !== null) return;
    watchdog = setInterval(() => { if (active && !valid(active)) cancel(); }, 150);
  }
  function park(state) {
    if (!state.source || state.parking) return;
    state.marker = document.createComment('np-file-drag-source');
    state.source.before(state.marker);
    state.parking = document.createElement('div');
    state.parking.className = 'np-file-drag-parking';
    state.parking.setAttribute('aria-hidden', 'true');
    document.body.appendChild(state.parking);
    // Keep the ORIGINAL native source connected outside #main during R().
    state.parking.appendChild(state.source);
  }
  function showHint(target, state) {
    clearVisuals();
    highlight = target.el; highlight.classList.add('np-file-drop-target');
    hint = document.querySelector('#dropzone-overlay');
    if (!hint) return;
    hint.classList.add('np-file-drop-hint');
    hint.setAttribute('role', 'status');
    hint.setAttribute('aria-live', 'polite');
    hint.style.pointerEvents = 'none';
    // Icon is a trusted constant; all destination/user-controlled text is textContent.
    if (!hint.querySelector('.np-file-drop-label')) {
      hint.replaceChildren();
      const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
      svg.setAttribute('viewBox', '0 0 24 24'); svg.setAttribute('aria-hidden', 'true');
      const shape = document.createElementNS('http://www.w3.org/2000/svg', 'path');
      shape.setAttribute('d', 'M12 16V3m-5 5 5-5 5 5M4 15v6h16v-6');
      svg.appendChild(shape); hint.appendChild(svg);
      const label = document.createElement('span'); label.className = 'np-file-drop-label'; hint.appendChild(label);
    }
    hint.querySelector('.np-file-drop-label').textContent = `${state.internal ? 'Move' : 'Upload'} to ${target.destination || 'root'}${target.folder ? ' · Hold to open' : ' · Current folder'}`;
    hint.style.display = 'flex';
  }
  function scheduleHover(target, state) {
    if (!target.folder || target.destination === path(S.currentDir || '') ||
        (state.internal && state.paths.some(source => target.destination === source || target.destination.startsWith(source + '/')))) { stopHover(); return; }
    if (hover?.el === target.el && hover.destination === target.destination) return;
    stopHover(); hover = target;
    timer = setTimeout(async () => {
      timer = null;
      if (active !== state || !valid(state) || pending || !target.el.isConnected || state.navigation) return;
      const navigation = { to: target.destination };
      state.navigation = navigation;
      park(state);
      const isCurrent = () => active === state && state.navigation === navigation && valid(state);
      try {
        await loadFileList(target.destination, false, { isCurrent });
        if (isCurrent()) state.dir = path(S.currentDir || '');
      } catch (_) { if (isCurrent()) toast('Could not open that folder.', 'bad'); }
      finally { if (state.navigation === navigation) state.navigation = null; }
    }, 650);
  }
  function onStart(event) {
    const el = element(event.target), source = el?.closest(ROW);
    if (!source || !source.closest('.np-files-workspace')) return;
    if (blockedOrigin || blockedControl(el)) { event.preventDefault(); return; }
    if (!event.dataTransfer || pending || !allowed(true)) { event.preventDefault(); return; }
    cancel();
    rejectedExternal = false;
    const sourcePath = path(source.dataset.path);
    const selected = S.selectedFiles;
    const paths = Array.from(new Set(selected?.has(sourcePath) ? Array.from(selected) : [sourcePath]));
    if (!sourcePath || !paths.length || paths.length > 1000 || paths.some(p => !p || path(p) !== p)) {
      event.preventDefault(); toast('Choose up to 1000 valid file paths to move.', 'bad'); return;
    }
    active = capture(true); active.paths = paths; active.source = source;
    active.isDirectory = source.dataset.isDir === 'true';
    if (active.dir === null) { cancel(); event.preventDefault(); return; }
    // Marker only, never serialized paths. External/custom payloads are untrusted.
    event.dataTransfer.setData(TYPE, 'internal'); event.dataTransfer.effectAllowed = 'move';
    source.classList.add('np-file-dragging'); watch();
  }
  function onOver(event) {
    if (!relevant(event)) return;
    event.preventDefault(); // Also prevents external file navigation OUTSIDE the workspace.
    if (active && !valid(active)) { cancel(); return; }
    if (!active && !rejectedExternal && filesDrag(event.dataTransfer) && allowed(false) && !pending) { active = capture(false); watch(); }
    const target = targetAt(event.target);
    if (!active || pending || !valid(active) || !target) {
      stopHover(); clearVisuals(); if (event.dataTransfer) event.dataTransfer.dropEffect = 'none'; return;
    }
    if (active.internal && moveError(active, target.destination)) {
      clearVisuals(); event.dataTransfer.dropEffect = 'none'; scheduleHover(target, active); return;
    }
    event.dataTransfer.dropEffect = active.internal ? 'move' : 'copy';
    showHint(target, active); scheduleHover(target, active);
  }
  function onLeave(event) {
    if (!relevant(event)) return;
    const related = element(event.relatedTarget);
    if (related && highlight?.contains(related)) return;
    if (related) { stopHover(); clearVisuals(); return; }
    // Child transitions often have a null relatedTarget. Hit test rather than
    // relying on an unbalanced dragenter/dragleave counter.
    const hit = document.elementFromPoint(event.clientX, event.clientY);
    if (hit && highlight?.contains(hit) && event.clientX > 0 && event.clientY > 0) return;
    stopHover(); clearVisuals();
    if (!hit && !active?.internal) { cancel(); rejectedExternal = false; }
  }
  function uploadSnapshot(dt) {
    const files = Array.from(dt.files || []), handles = [];
    for (const item of Array.from(dt.items || [])) {
      if (item.kind !== 'file') continue;
      const entry = item.webkitGetAsEntry?.() || item.getAsEntry?.();
      if (entry?.isDirectory) throw new Error('Folder uploads are not supported. Select files instead.');
      if (!entry && item.getAsFileSystemHandle) handles.push(Promise.resolve(item.getAsFileSystemHandle()));
    }
    // Acquire handles during the drop event; browsers revoke access afterwards.
    return { files, handles };
  }
  async function onDrop(event) {
    if (!relevant(event)) return;
    event.preventDefault();
    const target = targetAt(event.target);
    const state = active || (!rejectedExternal && filesDrag(event.dataTransfer) && allowed(false) ? capture(false) : null);
    const safe = state && valid(state) && target && !pending;
    let snapshot, error = '';
    if (safe) {
      if (state.internal) error = moveError(state, target.destination);
      else { try { snapshot = uploadSnapshot(event.dataTransfer); } catch (e) { error = e.message; } }
    }
    endDrag();
    rejectedExternal = false;
    if (!safe) return;
    if (error) { toast(error, 'bad'); return; }
    const operation = { state }; pending = operation;
    document.body.classList.add('np-file-drag-pending');
    try {
      if (state.internal) {
        if (!valid(state)) return;
        await api(srvApiPath(state.server, '/files/move'), { method: 'POST',
          body: JSON.stringify({ paths: state.paths, destination: target.destination, overwrite: false }) });
        if (!current(state)) return;
        for (const source of state.paths) S.selectedFiles?.delete(source);
        if (typeof syncFileSelectionUI === 'function') syncFileSelectionUI();
        toast(`Moved ${state.paths.length} item${state.paths.length === 1 ? '' : 's'}.`, 'ok');
        if (valid(state)) {
          const dir = path(S.currentDir || '');
          await loadFileList(dir, false, { isCurrent: () => current(state) && path(S.currentDir || '') === dir });
        }
      } else {
        const handles = await Promise.all(snapshot.handles);
        if (!valid(state)) return;
        if (handles.some(handle => handle?.kind === 'directory')) throw new Error('Folder uploads are not supported. Select files instead.');
        if (!snapshot.files.length || snapshot.files.length > 1000) throw new Error('Choose between 1 and 1000 files to upload.');
        await enqueueUploads(snapshot.files, target.destination);
      }
    } catch (e) { if (current(state)) toast(e.message || 'File operation failed. Nothing was overwritten.', 'bad'); }
    finally {
      if (pending === operation) { pending = null; document.body.classList.remove('np-file-drag-pending'); }
    }
  }
  function afterRender(root = document) {
    if (active && (!valid(active) || (active.source && !active.source.isConnected))) cancel();
    const rows = Array.from(root.querySelectorAll(ROW));
    if (root.matches?.(ROW)) rows.unshift(root);
    for (const row of rows) {
      if (row !== active?.source) row.draggable = allowed(true) && Boolean(row.closest('.np-files-workspace'));
      for (const label of row.querySelectorAll('.np-file-open')) label.draggable = row.draggable;
    }
    if (!active) clearVisuals();
  }
  function init() {
    if (initialized) return;
    initialized = true;
    document.addEventListener('pointerdown', event => { blockedOrigin = blockedControl(event.target); }, true);
    document.addEventListener('pointerup', () => { blockedOrigin = false; }, true);
    document.addEventListener('dragstart', onStart, true);
    document.addEventListener('dragenter', onOver, true);
    document.addEventListener('dragover', onOver, true);
    document.addEventListener('dragleave', onLeave, true);
    document.addEventListener('drop', onDrop, true);
    document.addEventListener('dragend', endDrag, true);
    document.addEventListener('keydown', event => { if (event.key === 'Escape') cancel(); }, true);
    window.addEventListener('pagehide', cancel);
    window.addEventListener('popstate', cancel);
    // Unrelated user navigation must not leave a parked source or late hover.
    document.addEventListener('click', () => { if (active) cancel(); }, true);
    afterRender();
  }
  window.NPFileDrag = Object.freeze({ init, afterRender, cancel });
})();
