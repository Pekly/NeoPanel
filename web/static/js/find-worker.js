'use strict';
importScripts('./find-search.js');
self.onmessage = ({ data }) => {
  try { self.postMessage({ id: data.id, ...NPFileSearch.search(data.text, data.query) }); }
  catch (_) { self.postMessage({ id: data.id, error: 'File search failed' }); }
};
