/* Presentation only: bounded shapes, no state, timers, requests or fake controls.
   Callers own loading/refresh/error decisions and provide their existing status copy. */
(function (root) {
  'use strict';
  const kinds = new Set(['servers', 'files', 'plugins', 'installed', 'players', 'backups', 'schedules', 'databases', 'activity']);
  const modes = new Set(['list', 'tiles', 'icons']);
  const escape = value => String(value ?? '').replace(/[&<>"']/g, ch => ({'&':'&amp;', '<':'&lt;', '>':'&gt;', '"':'&quot;', "'":'&#39;'}[ch]));
  const bar = cls => `<span class="np-load-bar ${cls || ''}"></span>`;
  function html(kind, label = 'Loading...', mode = 'list') {
    kind = kinds.has(kind) ? kind : 'files';
    mode = modes.has(mode) ? mode : 'list';
    const count = kind === 'servers' ? 3 : mode === 'icons' ? 6 : 4;
    const copy = `<div class="np-load-copy">${bar('np-load-title')}${bar('np-load-short')}</div>`;
    const item = `<div class="np-load-item">${kind === 'servers' ? '<div class="np-load-cover"></div>' : ''}<div class="np-load-head">${bar('np-load-icon')}${copy}</div><div class="np-load-meta">${bar()}${bar('np-load-short')}</div></div>`;
    return `<div class="np-loading np-loading-${kind} np-loading-${mode}"><div class="np-loading-status" role="status" aria-live="polite">${escape(label)}</div><div class="np-load-shapes" aria-hidden="true">${Array.from({length: count}, () => item).join('')}</div></div>`;
  }
  root.NPLoading = Object.freeze({ html });
  if (typeof module !== 'undefined' && module.exports) module.exports = root.NPLoading;
})(typeof window !== 'undefined' ? window : globalThis);
