/* NeoPanel syntax worker. Only local, pinned Prism grammars; never highlight() or HTML. */
'use strict';
self.Prism = { manual: true, disableWorkerMessageHandler: true };
importScripts('../vendor/prism-1.30.0/prism.js');

const MAX_CHARS = 200000;
const MAX_RUNS = 6000;
const languages = new Set(['markup', 'css', 'javascript', 'typescript', 'json', 'yaml',
  'toml', 'ini', 'properties', 'java', 'python', 'bash', 'powershell', 'sql',
  'markdown', 'go', 'rust', 'lua', 'docker']);
const categories = Object.freeze({
  comment: 'comment', prolog: 'comment', doctype: 'comment', cdata: 'comment',
  string: 'string', char: 'string', 'attr-value': 'string', url: 'string',
  number: 'number', boolean: 'number', constant: 'number',
  keyword: 'keyword', atrule: 'keyword', important: 'keyword', builtin: 'keyword',
  property: 'property', key: 'property', selector: 'property', 'attr-name': 'property',
  tag: 'tag', function: 'function', 'class-name': 'function',
  operator: 'operator', punctuation: 'punctuation', regex: 'string',
  variable: 'variable', symbol: 'variable', inserted: 'string', deleted: 'tag',
  title: 'keyword', bold: 'keyword', italic: 'string', code: 'string'
});

self.onmessage = function (event) {
  const { generation, version, language, text } = event.data || {};
  if (!Number.isSafeInteger(generation) || !Number.isSafeInteger(version)) return;
  const reply = { generation, version };
  if (typeof text !== 'string' || text.length > MAX_CHARS || !languages.has(language)) {
    self.postMessage({ ...reply, error: 'limit' });
    return;
  }
  try {
    // Offsets, not markup or worker-supplied text, cross the worker boundary.
    const runs = [];
    let offset = 0;
    function flatten(value, inherited, depth) {
      if (depth > 64) throw new Error('complexity');
      if (typeof value === 'string') {
        if (!value.length) return;
        offset += value.length;
        const previous = runs[runs.length - 1];
        if (previous && previous[1] === inherited) previous[0] = offset;
        else {
          if (runs.length >= MAX_RUNS) throw new Error('complexity');
          runs.push([offset, inherited]);
        }
      } else if (Array.isArray(value)) {
        for (const child of value) flatten(child, inherited, depth + 1);
      } else {
        const kind = Object.prototype.hasOwnProperty.call(categories, value.type) ? categories[value.type] : inherited;
        flatten(value.content, kind, depth + 1);
      }
    }
    flatten(Prism.tokenize(text, Prism.languages[language]), '', 0);
    if (offset !== text.length) throw new Error('invalid tokens');
    self.postMessage({ ...reply, runs });
  } catch (_) {
    self.postMessage({ ...reply, error: 'complexity' });
  }
};
