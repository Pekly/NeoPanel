// Runs synchronously before CSS: apply only validated appearance values, never cached code.
(() => {
  'use strict';
  const key = 'neopanel_appearance_v1';
  const root = document.documentElement;
  const defaults = {preset:'obsidian', accent_color:'#387aff', accent_hover:'#4f8aff', bg_deep:'', bg_canvas:'', card_bg:'', nav_bg:'', text_color:'', card_opacity:72, border_radius_px:14, glass_blur_px:16, default_theme:'dark'};
  const read = name => { try { return localStorage.getItem(name); } catch (_) { return null; } };
  const color = value => typeof value === 'string' && /^#[a-f0-9]{6}$/i.test(value) ? value : '';
  function clean(raw) {
    const b = {...defaults};
    if (!raw || typeof raw !== 'object') return b;
    for (const name of ['accent_color','accent_hover','bg_deep','bg_canvas','card_bg','nav_bg','text_color']) b[name] = color(raw[name]);
    b.preset = raw.preset === 'obsidian' ? 'obsidian' : 'custom';
    for (const [name,min,max] of [['card_opacity',35,100],['border_radius_px',0,28],['glass_blur_px',0,32]]) {
      if (typeof raw[name] === 'number' && Number.isFinite(raw[name])) b[name] = Math.max(min, Math.min(max, raw[name]));
    }
    if (['light','dark','oled','auto'].includes(raw.default_theme)) b.default_theme = raw.default_theme;
    return b;
  }
  function mode(value) {
    const selected = ['Light','Dark','Oled','Auto'].includes(value) ? value : 'Dark';
    root.classList.toggle('oled', selected === 'Oled');
    root.dataset.theme = selected === 'Light' || (selected === 'Auto' && window.matchMedia?.('(prefers-color-scheme: light)').matches) ? 'light' : 'dark';
    root.style.colorScheme = root.dataset.theme;
    return selected;
  }
  const rgb = value => color(value) ? [1,3,5].map(i => parseInt(value.slice(i,i+2),16)).join(', ') : '';
  function applyPalette(raw) {
    const b = clean(raw), style = root.style;
    const set = (name,value) => value ? style.setProperty(name,value) : style.removeProperty(name);
    const ac = rgb(b.accent_color), custom = ac && (b.accent_color.toLowerCase() !== '#387aff' || b.preset !== 'obsidian');
    set('--ac', custom ? b.accent_color : ''); set('--ac-hover', custom ? b.accent_hover || b.accent_color : '');
    for (const [name,opacity] of [['--ac-glow',.25],['--ac-subtle',.12],['--acs',.14],['--border-accent',.42]]) set(name,custom ? `rgba(${ac}, ${opacity})` : '');
    set('--bg-deep',b.bg_deep); set('--bg-canvas',b.bg_canvas); set('--tx',b.text_color);
    set('--nav-bg',rgb(b.nav_bg) ? `rgba(${rgb(b.nav_bg)}, 0.88)` : '');
    const fallback = root.dataset.theme === 'light' ? '255, 255, 255' : root.classList.contains('oled') ? '8, 10, 14' : '18, 24, 38';
    set('--card-bg',rgb(b.card_bg) || b.card_opacity !== 72 ? `rgba(${rgb(b.card_bg) || fallback}, ${(b.card_opacity/100).toFixed(2)})` : '');
    set('--card-bg-solid',b.card_bg); set('--np-radius',`${b.border_radius_px}px`); set('--np-blur',`${b.glass_blur_px}px`);
  }
  let cached = null;
  try { const stored = JSON.parse(read(key) || 'null'); if (stored?.version === 1) cached = clean(stored.branding); } catch (_) {}
  if (window.NPInitialAppearance && typeof window.NPInitialAppearance === 'object') cached = clean(window.NPInitialAppearance);
  const savedMode = read('neopanel_mode');
  const defaultMode = {light:'Light',dark:'Dark',oled:'Oled',auto:'Auto'}[cached?.default_theme] || 'Dark';
  const initialMode = mode(savedMode || defaultMode);
  if (read('neopanel_compact') === 'true') root.classList.add('compact');
  if (cached) applyPalette(cached);
  window.NPTheme = Object.freeze({ initialBranding: cached || {}, initialMode, applyPalette, mode,
    remember(branding) { try { localStorage.setItem(key,JSON.stringify({version:1,branding:clean(branding)})); } catch (_) {} }
  });
})();
