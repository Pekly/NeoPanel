/* Standard project Markdown + tightly allowlisted HTML. Never insert parser output
 * without DOMPurify; provider descriptions and release notes are untrusted. */
(() => {
  'use strict';
  const MAX_TEXT = 100000, MAX_NODES = 12000;
  const escape = value => String(value ?? '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
  const tags = ['p','br','hr','h1','h2','h3','h4','h5','h6','strong','b','em','i','u','s','del','blockquote','pre','code','kbd','samp','ul','ol','li','a','img','table','thead','tbody','tfoot','tr','th','td','div','span','center','details','summary','sup','sub','dl','dt','dd','input'];
  const attrs = ['href','src','alt','title','width','height','align','start','colspan','rowspan','open','type','checked','disabled'];
  function safeURL(value) {
    if (!value || /[\u0000-\u001f\u007f]/.test(value)) return '';
    try {
      const u = new URL(value, 'https://modrinth.com/');
      return /^https?:$/.test(u.protocol) && !u.username && !u.password ? u.href : '';
    } catch (_) { return ''; }
  }
  function render(value) {
    const original = String(value ?? ''), source = original.slice(0, MAX_TEXT);
    if (!source.trim()) return '<p class="mu">No content provided by the project.</p>';
    const shortened = original.length > MAX_TEXT ? '<p class="mu">Content shortened. Read the full text on Modrinth.</p>' : '';
    const fallback = () => '<p class="mu">Formatted preview unavailable. Showing project text.</p><pre class="np-markdown-fallback">' + escape(source) + '</pre>' + shortened;
    if (!window.marked?.parse || !window.DOMPurify?.sanitize || !window.DOMPurify.isSupported) return fallback();
    try {
      const parsed = window.marked.parse(source, {gfm:true, breaks:false, async:false});
      if (typeof parsed !== 'string' || parsed.length > 1000000) return fallback();
      const fragment = window.DOMPurify.sanitize(parsed, {
        ALLOWED_TAGS:tags, ALLOWED_ATTR:attrs, ALLOW_DATA_ATTR:false, ALLOW_ARIA_ATTR:false,
        ALLOW_UNKNOWN_PROTOCOLS:false, SANITIZE_DOM:true, SANITIZE_NAMED_PROPS:true,
        RETURN_DOM_FRAGMENT:true
      });
      const elements = fragment.querySelectorAll('*');
      if (elements.length > MAX_NODES) return fallback();
      for (const node of elements) {
        for (const attr of ['href','src']) {
          if (!node.hasAttribute(attr)) continue;
          const safe = safeURL(node.getAttribute(attr));
          if (safe && ((attr === 'href' && node.tagName === 'A') || (attr === 'src' && node.tagName === 'IMG'))) node.setAttribute(attr, safe);
          else node.removeAttribute(attr);
        }
        for (const attr of ['width','height','colspan','rowspan','start']) {
          if (!node.hasAttribute(attr)) continue;
          const raw = node.getAttribute(attr);
          if (!/^\d{1,4}$/.test(raw) || Number(raw) > 2000 || Number(raw) < 1) node.removeAttribute(attr);
        }
        if (node.hasAttribute('align') && !/^(left|center|right)$/i.test(node.getAttribute('align'))) node.removeAttribute('align');
        if (node.tagName === 'A' && node.hasAttribute('href')) {
          node.setAttribute('target','_blank'); node.setAttribute('rel','noopener noreferrer');
        }
        if (node.tagName === 'IMG') {
          if (!node.hasAttribute('src')) { node.replaceWith(document.createTextNode(node.getAttribute('alt') || '')); continue; }
          node.setAttribute('loading','lazy'); node.setAttribute('referrerpolicy','no-referrer');
        }
        if (node.tagName === 'INPUT') {
          if (node.getAttribute('type') !== 'checkbox') { node.remove(); continue; }
          node.setAttribute('disabled',''); node.setAttribute('tabindex','-1');
        }
      }
      const holder = document.createElement('div'); holder.append(fragment);
      for (const table of holder.querySelectorAll('table')) {
        const scroller = document.createElement('div'); scroller.className = 'np-markdown-table';
        table.replaceWith(scroller); scroller.append(table);
      }
      return holder.innerHTML + shortened;
    } catch (_) { return fallback(); }
  }
  window.NPProjectMarkdown = Object.freeze({render});
})();
