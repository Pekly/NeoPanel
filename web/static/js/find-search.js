// Literal editor-text search, shared by the worker and a bounded compatibility fallback.
(function (global) {
  'use strict';
  const MAX_TEXT = 5 * 1024 * 1024, MAX_MATCHES = 1000, MAX_QUERY = 256;
  function search(source, query) {
    if (typeof source !== 'string' || typeof query !== 'string' || !query) return { matches: [], limited: false };
    const text = source.slice(0, MAX_TEXT);
    const pattern = new RegExp(query.slice(0, MAX_QUERY).replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'giu');
    const matches = [];
    let hit, line = 0, lineStart = 0, nextBreak = text.indexOf('\n'), limited = source.length > MAX_TEXT;
    while ((hit = pattern.exec(text))) {
      if (matches.length === MAX_MATCHES) { limited = true; break; }
      while (nextBreak >= 0 && nextBreak < hit.index) {
        line++; lineStart = nextBreak + 1; nextBreak = text.indexOf('\n', lineStart);
      }
      // Preserve ORIGINAL UTF-16 offsets: lowercasing first can change string length.
      matches.push([hit.index, hit.index + hit[0].length, line, hit.index - lineStart]);
    }
    return { matches, limited };
  }
  global.NPFileSearch = Object.freeze({ search, MAX_TEXT, MAX_MATCHES, MAX_QUERY });
})(typeof self === 'undefined' ? globalThis : self);
