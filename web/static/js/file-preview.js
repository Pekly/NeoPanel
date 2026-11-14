// Read-only file/media/archive views. All bytes still pass through the authenticated server API.
const FILE_PREVIEW = { current: null, retainedMedia: null };

function filePreviewActive(preview = FILE_PREVIEW.current) {
  return Boolean(preview && FILE_PREVIEW.current === preview && S.view === 'fi' && S.token === preview.token &&
    sameQolServer(preview.server, curServer()) &&
    (typeof hasServerPermission !== 'function' || hasServerPermission('files.view', preview.server)));
}
function disposeFilePreview() {
  FILE_PREVIEW.retainedMedia = null;
  const media = document.querySelector('#file-preview-media');
  if (media) {
    media.onerror = null;
    media.onload = null;
    media.onloadedmetadata = null;
    if (typeof media.pause === 'function') media.pause();
    media.removeAttribute('src');
    if (typeof media.load === 'function') media.load();
  }
  FILE_PREVIEW.current = null;
}
// Keep the original DOM node across unrelated page renders (for example an
// upload finishing). This preserves playback time, volume, rate and pause state.
function captureFilePreviewMedia() {
  FILE_PREVIEW.retainedMedia = null;
  const preview = FILE_PREVIEW.current;
  const element = document.querySelector('#file-preview-media');
  if (!filePreviewActive(preview) || !element || preview.mediaError) return null;
  const retained = { preview, element, playing: element.tagName === 'VIDEO' && !element.paused && !element.ended };
  FILE_PREVIEW.retainedMedia = retained;
  return retained;
}
function restoreFilePreviewMedia(retained) {
  FILE_PREVIEW.retainedMedia = null;
  if (!retained || !filePreviewActive(retained.preview)) return;
  const replacement = document.querySelector('#file-preview-media');
  if (!replacement || replacement === retained.element) return;
  replacement.replaceWith(retained.element);
  if (retained.element.readyState >= 1 || (retained.element.complete && retained.element.naturalWidth > 0)) fileMediaLoaded(retained.element);
  if (retained.playing && retained.element.paused) {
    const result = retained.element.play();
    if (result?.catch) result.catch(() => {});
  }
}

function closeFilePreview() {
  ++QOL.openSequence;
  disposeFilePreview();
  R();
}
function filePreviewURL(preview, operation) {
  const url = srvApiPath(preview.server, `/files/${operation}?path=${encodeURIComponent(preview.path)}`);
  return `${url}&token=${encodeURIComponent(preview.token || '')}`;
}
function downloadPreviewFile() {
  const preview = FILE_PREVIEW.current;
  if (!filePreviewActive(preview)) return;
  window.open(filePreviewURL(preview, 'download'), '_blank', 'noopener,noreferrer');
}
async function openFilePreview(server, path, metadata) {
  disposeFilePreview();
  const preview = { server: { ...server }, path, token: S.token, metadata, loading: metadata.kind === 'zip',
    error: '', mediaError: '', entries: [], folder: '', query: '', entry: null, entrySequence: 0 };
  FILE_PREVIEW.current = preview;
  QOL.editor = null;
  S.editingFile = null;
  S.ftab = null;
  R();
  if (metadata.kind !== 'zip') return;
  try {
    const result = await api(srvApiPath(server, `/files/zip?path=${encodeURIComponent(path)}`));
    if (!filePreviewActive(preview)) return;
    if (!Array.isArray(result?.entries) || result.entries.length > 10000 || result.entries.some(entry => !entry || typeof entry.name !== 'string' || typeof entry.is_dir !== 'boolean' || !Number.isInteger(entry.index) || entry.index < 0)) throw Error('Invalid or oversized ZIP listing');
    preview.entries = result.entries;
  } catch (err) {
    if (!filePreviewActive(preview)) return;
    preview.error = err.message || 'Could not read the ZIP archive.';
  } finally {
    if (filePreviewActive(preview)) { preview.loading = false; R(); }
  }
}
function fileMediaLoaded(element) {
  if (!filePreviewActive() || element !== document.querySelector('#file-preview-media')) return;
  const loading = document.querySelector('#file-preview-loading');
  if (loading) loading.hidden = true;
}
function fileMediaFailed(element) {
  if (!filePreviewActive() || element !== document.querySelector('#file-preview-media')) return;
  FILE_PREVIEW.current.mediaError = 'This file could not be displayed. It may be damaged, unavailable, or use a codec your browser does not support. You can still download it.';
  R();
}
function zipPreviewRows(entries, folder = '', query = '') {
  const folders = new Map(), files = [];
  for (const entry of entries) {
    if (typeof entry.name !== 'string' || !entry.name.startsWith(folder)) continue;
    const relative = entry.name.slice(folder.length);
    if (!relative) continue;
    const slash = relative.indexOf('/');
    if (slash >= 0) {
      const name = relative.slice(0, slash);
      if (name) folders.set(name, { name, path: folder + name + '/', is_dir: true });
    } else if (entry.is_dir) {
      folders.set(relative, { name: relative, path: folder + relative + '/', is_dir: true });
    } else {
      files.push({ ...entry, name: relative, path: entry.name });
    }
  }
  const term = query.trim().toLowerCase();
  return [...folders.values(), ...files].filter(row => row.name.toLowerCase().includes(term))
    .sort((a, b) => Number(b.is_dir) - Number(a.is_dir) || a.name.localeCompare(b.name, undefined, { numeric: true }));
}
function archiveRowsHtml(preview) {
  const rows = zipPreviewRows(preview.entries, preview.folder, preview.query);
  if (!rows.length) return `<tr><td colspan="3" class="file-preview-empty">${preview.entries.length ? 'No matching entries in this folder.' : 'This ZIP archive is empty.'}</td></tr>`;
  return rows.map(row => `<tr><td><button type="button" class="np-file-open" ${row.is_dir
    ? `data-path="${esc(row.path)}" onclick="browseZipPreview(this.dataset.path)"`
    : `data-index="${Number(row.index)}" onclick="openZipPreviewEntry(Number(this.dataset.index))"`}>${ic(row.is_dir ? 'folder' : 'file', 15)} ${esc(row.name)}</button></td>
    <td>${row.is_dir ? 'Folder' : esc(formatFileSize(row.size, false))}</td>
    <td>${row.is_dir ? 'Browse' : 'Read-only preview'}</td></tr>`).join('');
}
function browseZipPreview(folder) {
  const preview = FILE_PREVIEW.current;
  if (!filePreviewActive(preview) || preview.metadata.kind !== 'zip') return;
  if (folder !== '' && !preview.entries.some(entry => entry.name === folder || entry.name.startsWith(folder))) return;
  preview.folder = folder;
  preview.query = '';
  preview.entry = null;
  preview.entrySequence++;
  R();
}
function filterZipPreview(query) {
  const preview = FILE_PREVIEW.current;
  if (!filePreviewActive(preview)) return;
  preview.query = query;
  const body = document.querySelector('#zip-preview-rows');
  if (body) body.innerHTML = archiveRowsHtml(preview);
}
async function openZipPreviewEntry(index) {
  const preview = FILE_PREVIEW.current;
  if (!filePreviewActive(preview) || preview.metadata.kind !== 'zip' || !Number.isInteger(index)) return;
  const entry = preview.entries.find(item => item.index === index && !item.is_dir);
  if (!entry) return;
  const sequence = ++preview.entrySequence;
  preview.entry = { name: entry.name, loading: true, content: '', error: '' };
  R();
  try {
    const result = await api(srvApiPath(preview.server, `/files/zip-entry?path=${encodeURIComponent(preview.path)}&index=${index}`));
    if (!filePreviewActive(preview) || preview.entrySequence !== sequence) return;
    if (typeof result?.content !== 'string') throw Error('Invalid text preview response');
    preview.entry.content = result.content;
  } catch (err) {
    if (!filePreviewActive(preview) || preview.entrySequence !== sequence) return;
    preview.entry.error = err.message || 'This entry cannot be previewed as text. Download or extract the archive to open it externally.';
  } finally {
    if (filePreviewActive(preview) && preview.entrySequence === sequence) { preview.entry.loading = false; R(); }
  }
}
function closeZipPreviewEntry() {
  if (!filePreviewActive()) return;
  FILE_PREVIEW.current.entrySequence++;
  FILE_PREVIEW.current.entry = null;
  R();
}
function archivePreviewHtml(preview) {
  if (preview.loading) return '<div class="file-preview-empty" role="status">Reading ZIP contents…</div>';
  if (preview.error) return `<div class="file-preview-message" role="alert">${esc(preview.error)}</div>`;
  const crumbs = [{ name: 'Archive root', path: '' }];
  let prefix = '';
  for (const name of preview.folder.split('/').filter(Boolean)) { prefix += name + '/'; crumbs.push({ name, path: prefix }); }
  const entry = preview.entry;
  return `<p class="mu">Browse this archive without extracting it. Text entries can be previewed read-only; binary entries cannot be edited here.</p>
    <div class="file-preview-archive-tools"><div role="navigation" aria-label="ZIP folders" class="file-preview-breadcrumbs">${ic('archive', 15)} ${crumbs.map((crumb, i) => i === crumbs.length - 1 ? `<span aria-current="location">${esc(crumb.name)}</span>` : `<button type="button" class="file-preview-crumb" data-path="${esc(crumb.path)}" onclick="browseZipPreview(this.dataset.path)">${esc(crumb.name)}</button>`).join('<span class="mu" aria-hidden="true">/</span>')}</div>
    <input id="zip-preview-search" type="search" value="${esc(preview.query)}" aria-label="Search ZIP folder" placeholder="Search this folder…" oninput="filterZipPreview(this.value)"></div>
    <div class="file-preview-archive-table"><table><thead><tr><th>Name</th><th>Uncompressed size</th><th>Action</th></tr></thead><tbody id="zip-preview-rows">${archiveRowsHtml(preview)}</tbody></table></div>
    <p class="mu">${preview.entries.length} entries · Nothing has been extracted or changed.</p>
    ${entry ? `<section class="file-preview-entry"><div class="file-preview-toolbar"><h3>${esc(entry.name)}</h3><button type="button" class="btn sm" onclick="closeZipPreviewEntry()">Close entry preview</button></div>
      ${entry.loading ? '<p role="status">Loading text preview…</p>' : entry.error ? `<p role="alert">${esc(entry.error)}</p>` : `<pre tabindex="0" aria-label="Read-only ZIP entry content"><code>${esc(entry.content)}</code></pre>`}</section>` : ''}`;
}
function renderFilePreview() {
  const preview = FILE_PREVIEW.current;
  if (!filePreviewActive(preview)) return '';
  const info = preview.metadata;
  const title = { image: 'Image preview', video: 'Video preview', zip: 'ZIP contents' }[info.kind] || 'File details';
  let body;
  if (info.kind === 'zip') body = archivePreviewHtml(preview);
  else if (preview.mediaError) body = `<div class="file-preview-message" role="alert">${esc(preview.mediaError)}</div>`;
  else if (info.kind === 'image' || info.kind === 'video') {
    const source = FILE_PREVIEW.retainedMedia?.preview === preview ? '' : `src="${esc(filePreviewURL(preview, 'media'))}"`;
    body = `<div id="file-preview-loading" class="mu" role="status">Loading ${info.kind} preview…</div><div class="file-preview-stage">${info.kind === 'image'
      ? `<img id="file-preview-media" ${source} alt="${esc(info.name || preview.path)}" referrerpolicy="no-referrer" onload="fileMediaLoaded(this)" onerror="fileMediaFailed(this)">`
      : `<video id="file-preview-media" ${source} controls playsinline preload="metadata" onloadedmetadata="fileMediaLoaded(this)" onerror="fileMediaFailed(this)">Your browser does not support video playback.</video>`}</div>
      ${info.kind === 'video' ? '<p class="mu">Playback depends on the file’s video/audio codecs and your browser. Download the file if playback is unsupported.</p>' : ''}`;
  } else {
    body = `<div class="file-preview-message"><h3>Text editing is unavailable for this file</h3><p>${esc(info.reason || 'This is a binary or unsupported file. Opening it in a text editor could corrupt it. Download it to use an appropriate application.')}</p></div>`;
  }
  return `<section class="card file-preview-workspace" aria-label="${esc(title)}">
    <div class="file-preview-toolbar"><div><button type="button" class="btn sm" onclick="closeFilePreview()">← Back to files</button><h2>${title}</h2></div><button type="button" class="btn sm" onclick="downloadPreviewFile()">${ic('download', 14)} Download${info.kind === 'zip' ? ' archive' : ''}</button></div>
    <div class="file-preview-metadata"><b>${esc(preview.path)}</b><span>${esc(formatFileSize(info.size, false))}</span><span>${esc(info.mime || 'application/octet-stream')}</span><span>Read-only</span></div>
    ${body}
  </section>`;
}
