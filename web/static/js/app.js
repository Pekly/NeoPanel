// NeoPanel — Ultra-Lightweight Minecraft Server Manager & Daemon SPA
// Zero external JS dependencies. Pure-Go backend integration.

// --- High-Quality Vector SVG Icon Generator ---
function ic(name, size = 18) {
  const p = {
    mail: `<path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/>`,
    terminal: `<polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/>`,
    settings: `<circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>`,
    activity: `<polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>`,
    files: `<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/>`,
    database: `<ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/>`,
    backup: `<polyline points="21 8 21 21 3 21 3 8"/><rect x="1" y="3" width="22" height="5"/><line x1="10" y1="12" x2="14" y2="12"/>`,
    clock: `<circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>`,
    network: `<circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>`,
    link: `<path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>`,
    play: `<polygon points="5 3 19 12 5 21 5 3"/>`,
    users: `<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>`,
    puzzle: `<path d="M20.5 11H19V7a2 2 0 0 0-2-2h-4V3.5a2.5 2.5 0 0 0-5 0V5H4a2 2 0 0 0-2 2v3.8h1.5a2.5 2.5 0 0 1 0 5H2V20a2 2 0 0 0 2 2h3.8v-1.5a2.5 2.5 0 0 1 5 0V22H17a2 2 0 0 0 2-2v-4h1.5a2.5 2.5 0 0 0 0-5z"/>`,
    store: `<circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/>`,
    cpu: `<rect x="4" y="4" width="16" height="16" rx="2" ry="2"/><rect x="9" y="9" width="6" height="6"/><line x1="9" y1="1" x2="9" y2="4"/><line x1="15" y1="1" x2="15" y2="4"/><line x1="9" y1="20" x2="9" y2="23"/><line x1="15" y1="20" x2="15" y2="23"/><line x1="20" y1="9" x2="23" y2="9"/><line x1="20" y1="14" x2="23" y2="14"/><line x1="1" y1="9" x2="4" y2="9"/><line x1="1" y1="14" x2="4" y2="14"/>`,
    layers: `<polygon points="12 2 2 7 12 12 22 7 12 2"/><polyline points="2 17 12 22 22 17"/><polyline points="2 12 12 17 22 12"/>`,
    shield: `<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>`,
    user: `<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>`,
    search: `<circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>`,
    server: `<rect x="2" y="2" width="20" height="8" rx="2" ry="2"/><rect x="2" y="14" width="20" height="8" rx="2" ry="2"/><line x1="6" y1="6" x2="6.01" y2="6"/><line x1="6" y1="18" x2="6.01" y2="18"/>`,
    download: `<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/>`,
    upload: `<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/>`,
    external: `<path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/>`,
    trash: `<polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>`,
    edit: `<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>`,
    folder: `<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>`,
    file: `<path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/><polyline points="13 2 13 9 20 9"/>`,
    plus: `<line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>`,
    refresh: `<polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>`,
    more: `<circle cx="12" cy="12" r="1"/><circle cx="12" cy="5" r="1"/><circle cx="12" cy="19" r="1"/>`,
    key: `<path d="M21 2l-2 2m-1.5 1.5L16 7l-2-2m-2 2l-2-2-4 4a5.5 5.5 0 1 0 7.78 7.78l4.22-4.22V11h3v-3h2V5h-2.5z"/><circle cx="7.5" cy="16.5" r="1.5"/>`,
    lock: `<rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>`,
    code: `<polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/>`,
    eye: `<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>`,
    alert: `<path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/>`,
    info: `<circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/>`,
    check: `<polyline points="20 6 9 17 4 12"/>`,
    archive: `<polyline points="21 8 21 21 3 21 3 8"/><rect x="1" y="3" width="22" height="5"/><line x1="10" y1="12" x2="14" y2="12"/>`,
    copy: `<rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>`,
    folderUp: `<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/><polyline points="12 10 12 16"/><polyline points="9 13 12 10 15 13"/>`,
    package: `<line x1="16.5" y1="9.4" x2="7.5" y2="4.21"/><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/>`,
    fileText: `<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/>`,
    image: `<rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/>`,
    viewList: `<line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/>`,
    viewIcons: `<rect x="3" y="3" width="7" height="7" rx="1.2"/><rect x="14" y="3" width="7" height="7" rx="1.2"/><rect x="3" y="14" width="7" height="7" rx="1.2"/><rect x="14" y="14" width="7" height="7" rx="1.2"/>`,
    viewTiles: `<rect x="3" y="4" width="18" height="6.5" rx="1.2"/><rect x="3" y="13.5" width="8" height="6.5" rx="1.2"/><rect x="13" y="13.5" width="8" height="6.5" rx="1.2"/>`,
    move: `<polyline points="15 14 20 9 15 4"/><path d="M4 20v-7a4 4 0 0 1 4-4h12"/>`
  };
  return `<svg width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="ic" aria-hidden="true">${p[name] || p.terminal}</svg>`;
}

function fileListIconHtml(item, extraCls = '', iconSize = 15) {
  const cls = extraCls ? ` ${extraCls}` : '';
  if (!item) return `<span class="fi-badge${cls}">${ic('file', iconSize)}</span>`;
  if (item.is_dir) {
    return `<span class="fi-badge dir${cls}" title="Folder">${ic('folder', iconSize)}</span>`;
  }
  const name = String(item.name || '').toLowerCase();
  const ext = name.includes('.') ? name.split('.').pop() : '';
  const srv = curServer();
  if (isSteelEngine(srv?.type) && name === serverExecutable(srv).split(/[\\/]/).pop().toLowerCase()) {
    return `<span class="fi-badge${cls}" title="Steel executable (Rust)">${engineIcon('steel', iconSize)}</span>`;
  }
  if (ext === 'jar') {
    const jarEngine = detectJarEngine(name);
    const iconUrl = getEngineIconUrl(jarEngine);
    const imgSize = Math.max(14, iconSize - 2);
    return `<span class="fi-badge jar${cls}" style="padding:0; overflow:hidden; display:inline-flex; align-items:center; justify-content:center;" title="${esc(item.name || 'JAR')} (${esc(engineMeta(jarEngine).label)} Jar)"><img src="${esc(iconUrl)}" alt="jar" style="width:${imgSize}px; height:${imgSize}px; object-fit:cover; border-radius:3px;" loading="lazy"></span>`;
  }
  if (['zip', 'tar', 'gz', 'tgz', 'rar', '7z', 'bz2', 'xz'].includes(ext)) {
    return `<span class="fi-badge arc${cls}" title="Compressed Archive">${ic('archive', iconSize)}</span>`;
  }
  if (['yml', 'yaml', 'json', 'toml', 'properties', 'conf', 'cfg', 'ini', 'xml'].includes(ext)) {
    return `<span class="fi-badge cfg${cls}" title="Configuration File">${ic('code', iconSize)}</span>`;
  }
  if (['sh', 'bat', 'cmd', 'ps1', 'js', 'ts', 'py', 'go'].includes(ext)) {
    return `<span class="fi-badge scr${cls}" title="Script File">${ic('terminal', iconSize)}</span>`;
  }
  if (['log', 'txt', 'md'].includes(ext)) {
    return `<span class="fi-badge doc${cls}" title="Text / Log File">${ic('fileText', iconSize)}</span>`;
  }
  if (['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'ico'].includes(ext)) {
    return `<span class="fi-badge img${cls}" title="Image File">${ic('image', iconSize)}</span>`;
  }
  if (['db', 'sqlite', 'sqlite3', 'sql'].includes(ext)) {
    return `<span class="fi-badge db${cls}" title="Database File">${ic('database', iconSize)}</span>`;
  }
  return `<span class="fi-badge${cls}" title="File">${ic('file', iconSize)}</span>`;
}

// --- Server Software Icons & Helpers ---
const ENGINE_ICONS = {
  steel: '/static/img/steel-logo.png',
  vanilla: 'https://i.ibb.co/MmxPTbN/image.jpg',
  paper: 'https://cdn.jsdelivr.net/gh/selfhst/icons/svg/papermc.svg',
  purpur: 'https://cdn.jsdelivr.net/gh/selfhst/icons/svg/purpurmc.svg',
  leaf: 'https://i.ibb.co/7tLC0XDz/image.webp',
  folia: 'https://cdn.jsdelivr.net/gh/selfhst/icons/svg/papermc-folia.svg',
  forge: 'https://i.ibb.co/CKpCYrPq/image-1.jpg',
  neoforge: 'https://i.ibb.co/NgGPZcVc/image-2.jpg',
  neoforged: 'https://i.ibb.co/NgGPZcVc/image-2.jpg',
  fabric: 'https://i.ibb.co/qYjbv9C1/image-1.webp',
  velocity: 'https://cdn.jsdelivr.net/gh/selfhst/icons/svg/papermc-velocity.svg',
  waterfall: 'https://cdn.jsdelivr.net/gh/selfhst/icons/svg/papermc.svg',
  bungeecord: 'https://cdn.jsdelivr.net/gh/selfhst/icons/svg/papermc-velocity.svg',
  custom: 'https://i.ibb.co/MmxPTbN/image.jpg'
};

function getEngineIconUrl(type) {
  const raw = String(type || 'paper').toLowerCase().trim();
  const k = raw === 'leafmc' ? 'leaf' : (raw === 'neoforged' ? 'neoforge' : raw);
  return ENGINE_ICONS[k] || ENGINE_ICONS.paper;
}

function detectJarEngine(filename = '') {
  const n = String(filename || '').toLowerCase();
  if (n.includes('paper')) return 'paper';
  if (n.includes('purpur')) return 'purpur';
  if (n.includes('leaf')) return 'leaf';
  if (n.includes('folia')) return 'folia';
  if (n.includes('neoforge') || n.includes('neoforged')) return 'neoforge';
  if (n.includes('forge')) return 'forge';
  if (n.includes('fabric')) return 'fabric';
  if (n.includes('velocity')) return 'velocity';
  if (n.includes('waterfall') || n.includes('bungee')) return 'velocity';
  if (n.includes('vanilla') || n.includes('spigot') || n.includes('craftbukkit')) return 'vanilla';
  return curServer()?.type || 'vanilla';
}

function isModdedEngine(type) {
  const k = String(type || '').toLowerCase().trim();
  return k === 'fabric' || k === 'forge' || k === 'neoforge';
}

function isSteelEngine(type) {
  return String(type || '').toLowerCase().trim() === 'steel';
}

function serverExecutable(srv) {
  return srv?.jar_file || (isSteelEngine(srv?.type) ? 'steel' : 'server.jar');
}

function nativeCommandPreview(filename) {
  const name = String(filename || 'steel');
  const windows = /\.exe$/i.test(name);
  const command = /^[./\\]|^[A-Za-z]:/.test(name) ? name : (windows ? '.\\' : './') + name;
  return /\s/.test(command) ? `"${command}"` : command;
}

function steelConfigNotice(showLink = true) {
  return `<section class="card steel-info" data-steel-config>
    <div class="steel-info-heading">
      <img src="/static/img/steel-logo.png" alt="" width="40" height="40">
      <div><h2>Steel</h2><span class="steel-info-subtitle">Written in Rust · Minecraft Java Edition</span></div>
      <span class="steel-info-badge">Experimental · Pre-alpha</span>
    </div>
    <p class="steel-info-intro">Steel is an alternative Minecraft Java Edition server implementation written in Rust. It is designed for performance, memory safety, and efficient resource use, and runs without a Java runtime.</p>
    <div class="steel-info-warning"><strong>Still under active development.</strong> Gameplay features are incomplete and save formats may change. Use a separate test world and keep backups before updating or switching server software.</div>
    <div class="steel-info-grid">
      <div><h3>Java Edition clients</h3><p>Players connect using a compatible Minecraft Java Edition client. The supported Minecraft version depends on the Steel release you install.</p></div>
      <div><h3>Not a Paper or Fabric fork</h3><p>Steel is a separate server implementation. Bukkit/Spigot plugins and Forge/Fabric mods are not compatible.</p></div>
    </div>
    <details class="steel-info-details">
      <summary>Configuration &amp; resource limits in NeoPanel</summary>
      <dl>
        <div><dt>Configuration</dt><dd>Edit <code>config/config.toml</code> in the File Manager. Steel does not use Java <code>server.properties</code>.</dd></div>
        <div><dt>Port &amp; saves</dt><dd>NeoPanel applies your assigned port to <code>server_port</code> in the <code>[server]</code> section on startup. Worlds are stored in <code>saves/</code>.</dd></div>
        <div><dt>Memory limit</dt><dd>When launched with systemd, NeoPanel enforces the configured memory limit, including Bubblewrap + Systemd launches. Without systemd, Steel can still run, but no hard RAM cap is applied. Check the console's Isolation line for the launch mode and MemoryMax.</dd></div>
      </dl>
    </details>
    <div class="steel-info-actions">
      <a class="btn" href="https://steelmc.dev/getting-started/introduction/" target="_blank" rel="noopener noreferrer">About Steel ↗</a>
      ${showLink ? `<button type="button" class="btn" onclick="go('fi')">Open File Manager</button>` : ''}
    </div>
  </section>`;
}

function engineFamily(type) {
  const t = String(type || '').toLowerCase().trim();
  if (['paper', 'purpur', 'leaf', 'folia', 'spigot'].includes(t)) return 'bukkit';
  if (['fabric', 'quilt'].includes(t)) return 'fabric';
  if (['forge', 'neoforge'].includes(t)) return 'forge';
  if (['velocity', 'waterfall', 'bungeecord'].includes(t)) return 'proxy';
  return t;
}

function incompatibleEngineChange(from, to) {
  return Boolean(from && to && engineFamily(from) !== engineFamily(to));
}

function engineMeta(type) {
  const raw = String(type || 'paper').toLowerCase().trim();
  const key = raw === 'leafmc' ? 'leaf' : (raw === 'neoforged' ? 'neoforge' : raw);
  const map = {
    steel: { key: 'steel', label: 'Steel', color: '#94a3b8' },
    paper: { key: 'paper', label: 'Paper', color: '#3b82f6' },
    purpur: { key: 'purpur', label: 'Purpur', color: '#a855f7' },
    leaf: { key: 'leaf', label: 'Leaf', color: '#22c55e' },
    folia: { key: 'folia', label: 'Folia', color: '#10b981' },
    fabric: { key: 'fabric', label: 'Fabric', color: '#e2b98b' },
    forge: { key: 'forge', label: 'Forge', color: '#f59e0b' },
    neoforge: { key: 'neoforge', label: 'NeoForged', color: '#fb923c' },
    vanilla: { key: 'vanilla', label: 'Vanilla', color: '#4ade80' },
    velocity: { key: 'velocity', label: 'Velocity', color: '#06b6d4' },
    waterfall: { key: 'waterfall', label: 'Waterfall', color: '#38bdf8' },
    bungeecord: { key: 'bungeecord', label: 'BungeeCord', color: '#818cf8' },
    custom: { key: 'custom', label: 'Custom', color: '#94a3b8' }
  };
  return map[key] || { key: key || 'custom', label: raw ? raw.charAt(0).toUpperCase() + raw.slice(1) : 'Paper', color: '#94a3b8' };
}

function engineIcon(type, size = 16, withBg = false) {
  const m = engineMeta(type);
  const iconUrl = getEngineIconUrl(type);
  const rad = size >= 24 ? 6 : (size >= 16 ? 4 : 3);
  const img = `<img src="${esc(iconUrl)}" alt="${esc(m.label)}" width="${size}" height="${size}" class="eng-img-ic" style="width:${size}px; height:${size}px; min-width:${size}px; min-height:${size}px; object-fit:contain; border-radius:${rad}px; display:inline-block; vertical-align:middle; flex-shrink:0;" loading="lazy">`;
  if (!withBg) return img;
  const wrapCls = size >= 22 ? 'lg' : (size <= 14 ? 'sm' : '');
  return `<span class="eng-ico-wrap ${wrapCls}" title="${esc(m.label)}">${img}</span>`;
}

function onPlayerHeadError(imgEl, fallbackUrl) {
  if (!imgEl) return;
  if (!imgEl.dataset.fallbackTried && fallbackUrl) {
    imgEl.dataset.fallbackTried = '1';
    imgEl.src = fallbackUrl;
    return;
  }
  imgEl.style.display = 'none';
  const fb = imgEl.parentElement?.querySelector('.head-fallback');
  if (fb) fb.style.display = 'inline-flex';
}

// --- SFTP Helpers (One-Click Launch, FileZilla Export & Password Guidance) ---
function getSftpDetails(srv) {
  const s = srv || curServer();
  const rawHost = (S.sftp?.host || srvHost(s) || window.location.hostname || '127.0.0.1').trim();
  const bracketedHost = rawHost.includes(':') && !rawHost.startsWith('[') ? `[${rawHost}]` : rawHost;
  const port = Number(S.sftp?.port || S.adminDaemonCfg?.sftpPort || S.brandingDaemon?.sftp_port || 2022) || 2022;
  const username = S.sftp?.username || `${S.acct.un || 'admin'}.${s ? s.id : 'srv-1'}`;
  const launchUri = S.sftp?.launch_uri || `sftp://${encodeURIComponent(username)}@${bracketedHost}:${port}/`;
  const isRemote = Boolean(s?.node_id && s.node_id !== 'local');
  return {
    host: rawHost,
    bracketedHost,
    port,
    username,
    launchUri,
    isRemote,
    enabled: S.sftp ? S.sftp.enabled !== false : true,
    listening: S.sftp ? Boolean(S.sftp.listening) : null,
    serverId: s?.id || 'srv-1',
    serverName: s?.name || 'Minecraft Server'
  };
}

function launchSftpClient() {
  if (!hasServerPermission('sftp.access')) return;
  const d = getSftpDetails(curServer());
  try {
    const a = document.createElement('a');
    a.href = d.launchUri;
    a.style.display = 'none';
    document.body.appendChild(a);
    a.click();
    a.remove();
  } catch (e) {
    window.location.href = d.launchUri;
  }
  const passMsg = d.isRemote
    ? 'Use the remote node token or your panel password when prompted.'
    : 'Use your NeoPanel account login password (or SSH key) when prompted.';
  toast(`Launching SFTP (${d.host}:${d.port}). ${passMsg}`);
}

function downloadFileZillaProfile() {
  const d = getSftpDetails(curServer());
  const xmlEsc = v => String(v == null ? '' : v)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
  const siteName = `NeoPanel - ${d.serverName}`;
  const comment = d.isRemote
    ? 'Remote Node SFTP: enter the remote node token or local account password on connect.'
    : 'SFTP Password: enter your NeoPanel account login password on connect.';
  const xml = `<?xml version="1.0" encoding="UTF-8"?>
<FileZilla3 version="3.67.0" platform="*">
  <Servers>
    <Server>
      <Host>${xmlEsc(d.host)}</Host>
      <Port>${d.port}</Port>
      <Protocol>1</Protocol>
      <Type>0</Type>
      <User>${xmlEsc(d.username)}</User>
      <Logontype>2</Logontype>
      <TimezoneOffset>0</TimezoneOffset>
      <PasvMode>MODE_DEFAULT</PasvMode>
      <MaximumMultipleConnections>0</MaximumMultipleConnections>
      <EncodingType>Auto</EncodingType>
      <BypassProxy>0</BypassProxy>
      <Name>${xmlEsc(siteName)}</Name>
      <Comments>${xmlEsc(comment)}</Comments>
      <SyncBrowsing>0</SyncBrowsing>
      <DirectoryComparison>0</DirectoryComparison>
    </Server>
  </Servers>
</FileZilla3>
`;
  const blob = new Blob([xml], { type: 'application/xml' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `neopanel-sftp-${shortServerId(d.serverId)}.xml`;
  document.body.appendChild(a);
  a.click();
  a.remove();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
  toast('Downloaded FileZilla SFTP profile (.xml). Use File -> Import in FileZilla.', 'ok');
}

function renderSftpCardHtml(srv, compact = false) {
  const d = getSftpDetails(srv);
  const statusPill = d.listening === true
    ? `<span class="sftp-status-pill live"><i class="dot" style="width:7px;height:7px;border-radius:50%;background:currentColor;"></i>Listening :${d.port}</span>`
    : d.listening === false
      ? `<span class="sftp-status-pill down">Port :${d.port} Offline</span>`
      : `<span class="sftp-status-pill">Port :${d.port}</span>`;

  return `<div class="card" style="padding: ${compact ? '18px 20px' : '22px'};">
    <div class="row sb" style="margin-bottom: 8px; gap: 10px; flex-wrap: wrap;">
      <h2 style="font-size: 16px; margin: 0; color: var(--tx); display:flex; align-items:center; gap:8px;">
        ${ic('lock', 15)} SFTP Management Endpoint
      </h2>
      ${statusPill}
    </div>
    <p class="mu" style="font-size: 13px; margin-bottom: 12px;">Pure-Go jailed SFTP daemon for WinSCP, FileZilla, Cyberduck, or OpenSSH.</p>
    <div class="mono" style="background: rgba(0, 0, 0, 0.36); border: 1px solid var(--border); padding: 11px 14px; border-radius: 8px; font-size: 12.5px; line-height: 1.8; margin-bottom: 12px; overflow-wrap: anywhere;">
      <div class="row sb" style="gap: 8px;">
        <span><span class="mu">Connection:</span> <b style="color: var(--tx);">sftp://${esc(d.host)}:${d.port}</b></span>
        <button type="button" class="btn-ico" data-val="sftp://${esc(d.host)}:${d.port}" onclick="copyDbText(this.dataset.val, 'SFTP Address')" title="Copy Host:Port">${ic('copy', 12)}</button>
      </div>
      <div class="row sb" style="gap: 8px;">
        <span><span class="mu">Username:</span> <b style="color: var(--tx);">${esc(d.username)}</b></span>
        <button type="button" class="btn-ico" data-val="${esc(d.username)}" onclick="copyDbText(this.dataset.val, 'SFTP Username')" title="Copy Username">${ic('copy', 12)}</button>
      </div>
    </div>
    <div class="sftp-pass-callout" style="margin-bottom: 14px;">
      <b>SFTP password:</b> Use your <b>NeoPanel account password</b> (or add an SSH key under Account).${d.isRemote ? ' Remote nodes also accept their <code>node_token</code> (<code>sudo neopanel node-token</code>).' : ''}
    </div>
    <div class="row" style="gap: 8px; flex-wrap: wrap;">
      <button type="button" class="btn pr sm" onclick="launchSftpClient()">${ic('external', 13)} Launch SFTP Client</button>
      <button type="button" class="btn sm" onclick="downloadFileZillaProfile()">${ic('download', 13)} FileZilla XML</button>
      <button type="button" class="btn sm" data-uri="${esc(d.launchUri)}" onclick="copyDbText(this.dataset.uri, 'SFTP URI')">${ic('copy', 13)} Copy SFTP URI</button>
    </div>
  </div>`;
}

// --- File Manager Sorting Helpers (Folders ALWAYS First) ---
function getFileExtension(name, isDir = false) {
  if (isDir) return '';
  const s = String(name || '').trim();
  const idx = s.lastIndexOf('.');
  return idx > 0 ? s.slice(idx + 1).toLowerCase() : '';
}

function sortFileItems(list, sortCfg = S.fileSort) {
  const arr = Array.isArray(list) ? [...list] : [];
  const by = ['name', 'ext', 'size', 'modified'].includes(sortCfg?.by) ? sortCfg.by : 'name';
  const mult = sortCfg?.dir === 'desc' ? -1 : 1;

  arr.sort((a, b) => {
    const aDir = Boolean(a?.is_dir);
    const bDir = Boolean(b?.is_dir);
    if (aDir !== bDir) return aDir ? -1 : 1;

    const aName = String(a?.name || '');
    const bName = String(b?.name || '');

    let cmp = 0;
    if (by === 'size') {
      if (!aDir && !bDir) {
        cmp = (Number(a?.size) || 0) - (Number(b?.size) || 0);
      }
      if (cmp === 0) cmp = aName.localeCompare(bName, undefined, { numeric: true, sensitivity: 'base' });
    } else if (by === 'ext') {
      if (!aDir && !bDir) {
        cmp = getFileExtension(aName, false).localeCompare(getFileExtension(bName, false), undefined, { sensitivity: 'base' });
      }
      if (cmp === 0) cmp = aName.localeCompare(bName, undefined, { numeric: true, sensitivity: 'base' });
    } else if (by === 'modified') {
      const aTs = Date.parse(a?.modified || '') || 0;
      const bTs = Date.parse(b?.modified || '') || 0;
      cmp = aTs - bTs;
      if (cmp === 0) cmp = aName.localeCompare(bName, undefined, { numeric: true, sensitivity: 'base' });
    } else {
      cmp = aName.localeCompare(bName, undefined, { numeric: true, sensitivity: 'base' });
    }
    return cmp * mult;
  });

  return arr;
}

function setFileSort(by, dir = null) {
  const validBy = ['name', 'ext', 'size', 'modified'].includes(by) ? by : 'name';
  const curBy = S.fileSort?.by || 'name';
  const curDir = S.fileSort?.dir === 'desc' ? 'desc' : 'asc';
  const nextDir = (dir === 'asc' || dir === 'desc')
    ? dir
    : (curBy === validBy ? (curDir === 'asc' ? 'desc' : 'asc') : 'asc');

  S.fileSort = { by: validBy, dir: nextDir };
  try {
    localStorage.setItem('neopanel_file_sort', JSON.stringify(S.fileSort));
  } catch (e) {}
  R();
}

// --- Subdomain & DNS/SRV Helpers ---
function computeSubdomainDraft(data, srv) {
  const cfg = data?.config || {};
  return {
    enabled: Boolean(cfg.enabled),
    label: cfg.label || '',
    base_domain: cfg.base_domain || data?.default_base_domain || '',
    target_host: cfg.target_host || data?.default_target || srvHost(srv) || ''
  };
}

function getSubdomainPreviewFqdn(draft) {
  const lbl = String(draft?.label || '').trim().replace(/^\.+|\.+$/g, '').toLowerCase();
  const base = String(draft?.base_domain || '').trim().replace(/^\.+|\.+$/g, '').toLowerCase();
  if (lbl && base) return `${lbl}.${base}`;
  if (base) return base;
  return lbl ? `${lbl}.yourdomain.com` : 'play.yourdomain.com';
}

function onSubdomainFieldChange() {
  const draft = S.subdomainDraft || {};
  const enabledEl = $('#sub-enabled');
  const labelEl = $('#sub-label');
  const baseEl = $('#sub-base');
  const targetEl = $('#sub-target');
  if (enabledEl) draft.enabled = Boolean(enabledEl.checked);
  if (labelEl) draft.label = labelEl.value.trim().toLowerCase();
  if (baseEl) draft.base_domain = baseEl.value.trim().toLowerCase();
  if (targetEl) draft.target_host = targetEl.value.trim();
  S.subdomainDraft = draft;

  const fqdn = getSubdomainPreviewFqdn(draft);
  const port = Number(S.subdomainData?.server_port || curServer()?.port || 25565);
  const fullAddr = port === 25565 ? fqdn : `${fqdn} (SRV -> :${port})`;
  const prevEl = $('#sub-fqdn-preview');
  if (prevEl) prevEl.textContent = fullAddr;
  const copyBtn = $('#sub-copy-fqdn-btn');
  if (copyBtn) copyBtn.dataset.val = fqdn;
}

async function loadSubdomainData(skipRender = false) {
  const srv = curServer();
  if (!srv) return;
  S.subdomainLoading = true;
  if (!skipRender && S.view === 'sub' && !S.subdomainData) R();
  try {
    const data = await api(srvApiPath(srv, '/subdomain'));
    S.subdomainData = data;
    S.subdomainDraft = computeSubdomainDraft(data, srv);
  } catch (e) {
    if (!S.subdomainDraft) S.subdomainDraft = computeSubdomainDraft(null, srv);
  } finally {
    S.subdomainLoading = false;
    if (!skipRender && S.view === 'sub') R();
  }
}

async function saveSubdomainConfig(btnEl) {
  const srv = curServer();
  if (!srv) return;
  onSubdomainFieldChange();
  const draft = S.subdomainDraft || {};
  const fqdn = (draft.label && draft.base_domain)
    ? `${draft.label}.${draft.base_domain}`
    : (draft.base_domain || '');
  if (btnEl) setBtnLoading(btnEl, true, 'Saving...');
  try {
    const res = await api(srvApiPath(srv, '/subdomain'), {
      method: 'PUT',
      body: JSON.stringify({
        enabled: Boolean(draft.enabled),
        label: draft.label || '',
        base_domain: draft.base_domain || '',
        fqdn,
        target_host: draft.target_host || ''
      })
    });
    S.subdomainData = res;
    S.subdomainDraft = computeSubdomainDraft(res, srv);
    toast(res?.message || 'Subdomain configuration saved!', 'ok');
    R();
  } catch (e) {
    toast(e.message || 'Failed to save subdomain settings', 'bad');
  } finally {
    if (btnEl && document.body.contains(btnEl)) setBtnLoading(btnEl, false);
  }
}

async function verifySubdomainDNS(btnEl) {
  const srv = curServer();
  if (!srv) return;
  onSubdomainFieldChange();
  const draft = S.subdomainDraft || {};
  const fqdn = (draft.label && draft.base_domain)
    ? `${draft.label}.${draft.base_domain}`
    : (draft.base_domain || '');
  if (!fqdn) {
    toast('Enter a subdomain label and base domain first.', 'bad');
    return;
  }
  S.subdomainVerifying = true;
  if (btnEl) setBtnLoading(btnEl, true, 'Checking DNS...');
  try {
    const res = await api(srvApiPath(srv, '/subdomain/verify'), {
      method: 'POST',
      body: JSON.stringify({
        fqdn,
        target_host: draft.target_host || ''
      })
    });
    S.subdomainVerify = res;
    toast(res?.status === 'verified' ? 'DNS records verified!' : 'DNS check completed', res?.status === 'verified' ? 'ok' : '');
    R();
  } catch (e) {
    toast(e.message || 'DNS verification failed', 'bad');
  } finally {
    S.subdomainVerifying = false;
    if (btnEl && document.body.contains(btnEl)) setBtnLoading(btnEl, false);
  }
}

async function syncPlayersList(btnEl) {
  const srv = curServer();
  if (!srv) return;
  if (btnEl) setBtnLoading(btnEl, true, 'Syncing...');
  try {
    await api(srvApiPath(srv, '/players/action'), {
      method: 'POST',
      body: JSON.stringify({ action: 'sync', player: '' })
    });
    toast('Requested /list from server console...');
    await new Promise(r => setTimeout(r, 350));
    await loadPlayersData();
  } catch (e) {
    toast(e.message || 'Server must be online to sync online players', 'bad');
  } finally {
    if (btnEl && document.body.contains(btnEl)) setBtnLoading(btnEl, false);
  }
}

// --- Unified Settings Helpers (Identity, Live Startup Command & Real-Time Auto-Save) ---
let _settingsAutoTimer = null;
let _settingsSaveInFlight = false;
let _settingsSaveQueued = null;
let _settingsLastSaved = {};
let _settingsUiState = { status: 'saved', text: '' };

function clearSettingsAutoSaveTimer() {
  if (_settingsAutoTimer) {
    clearTimeout(_settingsAutoTimer);
    _settingsAutoTimer = null;
  }
}

function setSettingsSaveStatus(status = 'saved', text = '') {
  const fallback = status === 'saving'
    ? 'Saving changes...'
    : status === 'warn'
      ? 'Enter a valid value to save'
      : status === 'error'
        ? 'Auto-save failed — check values'
        : 'All changes saved automatically';
  const msg = text || fallback;
  _settingsUiState = { status, text: msg };
  const classes = ['is-saved', 'is-saving', 'is-warn', 'is-error'];
  for (const id of ['set-autosave-pill', 'set-autosave-status']) {
    const el = document.getElementById(id);
    if (!el) continue;
    el.style.display = msg ? '' : 'none';
    el.classList.remove(...classes);
    el.classList.add(`is-${status}`);
    const labelEl = el.querySelector('.settings-autosave-text');
    if (labelEl) labelEl.textContent = msg;
    else el.innerHTML = `<span class="settings-autosave-dot"></span><span class="settings-autosave-text">${esc(msg)}</span>`;
  }
}

function getSettingsSnapshotKey(srv) {
  return `${srv?.node_id || 'local'}:${srv?.id || ''}`;
}

function buildSettingsSnapshots(srv, data) {
  const propsPayload = {
    'motd': data.motd,
    'server-port': String(data.port),
    'max-players': String(data.max),
    'difficulty': data.diff,
    'gamemode': data.mode,
    'pvp': String(data.pvp),
    'white-list': String(data.wl),
    'enable-command-block': String(data.cmd)
  };
  const srvPayload = {
    name: data.name,
    description: data.desc,
    type: srv?.type || 'paper',
    version: data.ver,
    jar_file: data.jar,
    port: data.port,
    memory_min: isSteelEngine(srv?.type) ? 0 : Math.min(1024, data.ram),
    memory_max: data.ram,
    java_path: isSteelEngine(srv?.type) ? '' : data.javaPath,
    java_args: isSteelEngine(srv?.type) ? '' : data.args,
    auto_restart: data.autoRestart,
    cpu_limit_percent: Number(data.cpuLimit) || 0,
    disk_limit_mb: Number(data.diskLimit) || 0
  };
  if (data.ownerId) {
    srvPayload.owner_id = Number(data.ownerId);
  }
  return {
    propsPayload,
    srvPayload,
    propsJson: JSON.stringify(propsPayload),
    srvJson: JSON.stringify(srvPayload)
  };
}

function collectSettingsFormPayload(srv) {
  if (!srv) return { ok: false, reason: 'No active server selected' };

  const nameEl = $('#set-srv-name');
  const descEl = $('#set-srv-desc');
  const verEl = $('#prop-ver');
  const jarEl = $('#prop-jar');
  const argsEl = $('#prop-args');
  const portEl = $('#prop-port');
  const maxEl = $('#prop-max');
  const ramEl = $('#prop-ram');
  const javaSelEl = $('#prop-java');
  const javaCustomEl = $('#prop-java-custom');
  const motdEl = $('#prop-motd');
  const diffEl = $('#prop-diff');
  const modeEl = $('#prop-mode');
  const autoRestartEl = $('#prop-autorestart');
  const pvpEl = $('#prop-sw-pvp');
  const wlEl = $('#prop-sw-wl');
  const cmdEl = $('#prop-sw-cmd');

  const name = nameEl ? nameEl.value.trim() : (srv.name || 'Minecraft Server');
  if (!name) return { ok: false, reason: 'Server name cannot be empty' };

  const desc = descEl ? descEl.value.trim() : (srv.description || '');
  const ver = verEl ? verEl.value.trim() : (srv.version || (isSteelEngine(srv.type) ? 'latest' : '1.21.11'));
  if (!ver) return { ok: false, reason: 'Minecraft version cannot be empty' };

  const jar = jarEl ? jarEl.value.trim() : serverExecutable(srv);
  if (!jar) return { ok: false, reason: 'Executable filename cannot be empty' };

  const args = argsEl ? argsEl.value.trim() : (srv.java_args !== undefined && srv.java_args !== null ? srv.java_args : '-XX:+UseG1GC');

  let port = Number(srv.port) || 25565;
  if (portEl) {
    const rawPort = String(portEl.value ?? '').trim();
    if (!/^\d+$/.test(rawPort)) return { ok: false, reason: 'Port must be between 1024 and 65535' };
    port = parseInt(rawPort, 10);
    if (port < 1024 || port > 65535) return { ok: false, reason: 'Port must be between 1024 and 65535' };
  }

  let max = Number(S.props.max) || 20;
  if (maxEl) {
    const rawMax = String(maxEl.value ?? '').trim();
    if (!/^\d+$/.test(rawMax)) return { ok: false, reason: 'Max players must be at least 1' };
    max = parseInt(rawMax, 10);
    if (max < 1 || max > 100000) return { ok: false, reason: 'Max players must be 1–100000' };
    S.props.max = max;
  }

  const ram = ramEl ? (parseInt(ramEl.value, 10) || getServerMaxRamMB(srv)) : getServerMaxRamMB(srv);

  let javaPath = srv.java_path || '';
  if (javaSelEl) {
    if (javaSelEl.value === 'custom') {
      const customPath = (javaCustomEl?.value || '').trim();
      if (!customPath) return { ok: false, reason: 'Enter a custom Java binary path' };
      javaPath = customPath;
    } else if (javaSelEl.value) {
      javaPath = javaSelEl.value;
    }
  }

  const motd = motdEl ? motdEl.value : (S.props.motd ?? srv.motd ?? 'A Minecraft Server');
  S.props.motd = motd;
  const diff = diffEl ? diffEl.value : (S.props.diff || 'normal');
  S.props.diff = diff;
  const mode = modeEl ? modeEl.value : (S.props.mode || 'survival');
  S.props.mode = mode;

  const pvp = pvpEl ? Boolean(pvpEl.checked) : Boolean(S.props.pvp);
  const wl = wlEl ? Boolean(wlEl.checked) : Boolean(S.props.wl);
  const cmd = cmdEl ? Boolean(cmdEl.checked) : Boolean(S.props.cmd);
  S.props.pvp = pvp;
  S.props.wl = wl;
  S.props.cmd = cmd;

  const autoRestart = autoRestartEl ? Boolean(autoRestartEl.checked) : (srv.auto_restart !== false);

  const cpuLimitEl = $('#prop-cpu-limit');
  const diskLimitEl = $('#prop-disk-limit');
  const ownerIdEl = $('#prop-owner-id');

  const cpuLimit = cpuLimitEl ? (parseInt(cpuLimitEl.value, 10) || 0) : (srv.cpu_limit_percent || 0);
  const diskLimit = diskLimitEl ? (parseInt(diskLimitEl.value, 10) || 0) : (srv.disk_limit_mb || 0);
  const ownerId = ownerIdEl ? (parseInt(ownerIdEl.value, 10) || srv.owner_id || 1) : (srv.owner_id || 1);

  return {
    ok: true,
    data: { name, desc, ver, jar, args, port, max, ram, javaPath, motd, diff, mode, pvp, wl, cmd, autoRestart, cpuLimit, diskLimit, ownerId }
  };
}

function seedSettingsSavedSnapshot(srv) {
  if (!srv) return;
  const res = collectSettingsFormPayload(srv);
  if (!res.ok) return;
  const snaps = buildSettingsSnapshots(srv, res.data);
  _settingsLastSaved[getSettingsSnapshotKey(srv)] = {
    propsJson: snaps.propsJson,
    srvJson: snaps.srvJson
  };
}

function syncHeroFromSavedServer(targetSrv) {
  const active = curServer();
  if (!active || !targetSrv) return;
  if (active.id !== targetSrv.id || (active.node_id || 'local') !== (targetSrv.node_id || 'local')) return;

  const heroName = $('#srv-hero-name');
  if (heroName) heroName.innerHTML = renderMinecraftTextHtml(targetSrv.name || 'Minecraft Server', { allowNewlines: false });

  const heroDesc = $('#srv-hero-desc');
  if (heroDesc) {
    const descTxt = String(targetSrv.description || '').trim();
    heroDesc.textContent = descTxt;
    heroDesc.style.display = descTxt ? 'block' : 'none';
  }

  const heroTitle = $('#srv-hero-title');
  if (heroTitle) {
    const trimmed = String(targetSrv.motd || '').trim();
    if (!trimmed) {
      heroTitle.style.display = 'none';
      heroTitle.innerHTML = '';
    } else {
      heroTitle.style.display = 'block';
      heroTitle.innerHTML = renderMinecraftTextHtml(trimmed, { allowNewlines: true });
    }
  }

  const engPill = $('#srv-hero-engine');
  if (engPill) {
    const srvType = targetSrv.type || S.core || 'paper';
    const eng = engineMeta(srvType);
    engPill.innerHTML = `${engineIcon(srvType, 13)} ${esc(eng.label)}${targetSrv.version ? ` ${esc(targetSrv.version)}` : ''}`;
  }

  const hst = $('#hst');
  if (hst) {
    const activeStat = S.status === 'online' || S.status === 'starting';
    const c = activeStat ? (S.cpu.at(-1) || 0) : 0;
    const maxRamMB = S.memMaxMB || getServerMaxRamMB(targetSrv);
    const m = activeStat ? (S.memUsedMB || ((S.ram.at(-1) || 0) / 100) * maxRamMB) : 0;
    const hostAddr = `${srvHost(targetSrv)}:${targetSrv.port || 25565}`;
    hst.innerHTML = `<button type="button" class="copyable-ip mono mu" data-addr="${esc(hostAddr)}" onclick="return copyServerAddress(this.dataset.addr, event)" title="Copy server address" aria-label="Copy server address"><span class="pv">${esc(hostAddr)}</span> ${ic('copy', 12)}</button><span class="mono mu">CPU ${c.toFixed(1)}%</span><span class="mono mu">RAM ${m.toFixed(0)} MiB / ${(maxRamMB / 1024).toFixed(1)} GiB</span>`;
  }

  updateSettingsCmdPreview();
}

function scheduleSettingsAutoSave(delay = 550) {
  if (!hasServerPermission('settings.edit')) return;
  if (S.view !== 'set') return;
  clearSettingsAutoSaveTimer();
  const srv = curServer();
  if (!srv) return;
  const check = collectSettingsFormPayload(srv);
  if (!check.ok) {
    setSettingsSaveStatus('warn', check.reason);
    return;
  }
  setSettingsSaveStatus('saving', 'Pending sync...');
  _settingsAutoTimer = setTimeout(() => {
    _settingsAutoTimer = null;
    void saveProperties({ autoSave: true, skipRender: true });
  }, delay);
}

function saveSettingsAutoNow() {
  if (!hasServerPermission('settings.edit')) return;
  clearSettingsAutoSaveTimer();
  if (S.view !== 'set') return;
  void saveProperties({ autoSave: true, skipRender: true });
}

function flushSettingsAutoSave() {
  if (!hasServerPermission('settings.edit')) { clearSettingsAutoSaveTimer(); return; }
  const hadPending = Boolean(_settingsAutoTimer);
  clearSettingsAutoSaveTimer();
  const srv = curServer();
  if (!srv || S.view !== 'set') return;
  const collected = collectSettingsFormPayload(srv);
  if (!collected.ok) {
    if (hadPending) setSettingsSaveStatus('warn', collected.reason);
    return;
  }
  const key = getSettingsSnapshotKey(srv);
  const prev = _settingsLastSaved[key];
  const snaps = buildSettingsSnapshots(srv, collected.data);
  const changed = !prev || prev.propsJson !== snaps.propsJson || prev.srvJson !== snaps.srvJson;
  if (hadPending || changed) {
    void persistSettingsPayload(srv, collected, { autoSave: true, skipRender: true });
  }
}

function flushPendingSettingsBeforeLeave() {
  if (S.view !== 'set') {
    clearSettingsAutoSaveTimer();
    return;
  }
  const hadPending = Boolean(_settingsAutoTimer);
  clearSettingsAutoSaveTimer();
  const srv = curServer();
  if (!srv) return;
  const collected = collectSettingsFormPayload(srv);
  if (!collected.ok) return;
  const key = getSettingsSnapshotKey(srv);
  const prev = _settingsLastSaved[key];
  const snaps = buildSettingsSnapshots(srv, collected.data);
  if (hadPending || !prev || prev.propsJson !== snaps.propsJson || prev.srvJson !== snaps.srvJson) {
    void persistSettingsPayload(srv, collected, { autoSave: true, skipRender: true, fromDom: false });
  }
}

function onSettingsRuleToggle(key, checked) {
  S.props[key] = Boolean(checked);
  saveSettingsAutoNow();
}

function onSettingsMaxPlayersInput(val) {
  const raw = String(val ?? '').trim();
  if (/^\d+$/.test(raw)) {
    const n = parseInt(raw, 10);
    if (n >= 1 && n <= 100000) S.props.max = n;
  }
  scheduleSettingsAutoSave(550);
}

function onPropJavaChange(val) {
  toggleCustomPropJava(val);
  updateSettingsCmdPreview();
  if (val === 'custom') {
    const customEl = $('#prop-java-custom');
    if (!customEl || !customEl.value.trim()) {
      setSettingsSaveStatus('warn', 'Enter a custom Java binary path');
      if (customEl) setTimeout(() => customEl.focus(), 20);
      return;
    }
  }
  saveSettingsAutoNow();
}

function onPropVersionSelectChange(val) {
  const inp = $('#prop-ver');
  if (inp) inp.value = val;
  saveSettingsAutoNow();
}

function onPropVersionInput(val) {
  const sel = $('#prop-ver-sel');
  if (sel && Array.from(sel.options).some(o => o.value === val)) {
    sel.value = val;
    if (sel._cselectRebuild) sel._cselectRebuild();
  }
  scheduleSettingsAutoSave(550);
}

function onServerIdentityInput() {
  const srv = curServer();
  const nameVal = $('#set-srv-name')?.value ?? '';
  const descVal = $('#set-srv-desc')?.value ?? '';
  S.editSrvName = nameVal;
  S.editSrvDesc = descVal;

  const heroName = $('#srv-hero-name');
  if (heroName) heroName.textContent = nameVal.trim() || srv?.name || 'Minecraft Server';
  const heroDesc = $('#srv-hero-desc');
  if (heroDesc) {
    heroDesc.textContent = descVal.trim();
    heroDesc.style.display = descVal.trim() ? 'block' : 'none';
  }
  scheduleSettingsAutoSave(500);
}

function updateSettingsCmdPreview() {
  const el = $('#set-cmd-preview');
  if (!el) return;
  const srv = curServer();
  if (isSteelEngine(srv?.type)) {
    el.textContent = nativeCommandPreview($('#prop-jar')?.value?.trim() || serverExecutable(srv));
    return;
  }
  let javaVal = $('#prop-java')?.value || resolveJavaSelectValue(srv?.java_path || '', srv?.version, srv?.type);
  let javaCmd = `java (OpenJDK ${javaVal})`;
  if (javaVal === 'custom') {
    javaCmd = $('#prop-java-custom')?.value?.trim() || srv?.java_path || 'java';
  }
  const memMax = parseInt($('#prop-ram')?.value, 10) || getServerMaxRamMB(srv) || 2048;
  const memMin = Math.min(1024, memMax);
  const jar = $('#prop-jar')?.value?.trim() || srv?.jar_file || 'server.jar';
  const argsEl = $('#prop-args');
  const args = argsEl ? argsEl.value.trim() : (srv?.java_args !== undefined && srv?.java_args !== null ? srv.java_args : '-XX:+UseG1GC');
  el.textContent = `${javaCmd} -Xms${memMin}M -Xmx${memMax}M ${args ? args + ' ' : ''}-jar ${jar} nogui`;
}

// --- Selectors & Shortcuts ---
const $ = s => document.querySelector(s);
const $$ = s => document.querySelectorAll(s);
const esc = s => String(s == null ? '' : s).replace(/[&<>"']/g, c => ({
  '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'
}[c]));

const DEFAULT_USER_AVATAR_URL = 'https://www.minecraft.net/content/dam/minecraftnet/games/minecraft/key-art/MCV_SummerDrop_Hero_DotNet_Downloadable_Wallpaper_r414x414.png';
const MINECRAFT_COVER_URL = 'https://www.minecraft.net/content/dam/minecraftnet/games/minecraft/key-art/MC-Downloadables_Image-0_15th-Anniversary_1110x624.jpg';
const SERVER_HERO_URL = 'https://m.media-amazon.com/images/I/71rzlQ2XMOL.jpg';

const DEFAULT_BRANDING = {
  brand_name: 'NeoPanel',
  brand_subtitle: 'Minecraft server management',
  logo_url: '',
  favicon_url: '',
  login_bg_url: '',
  server_hero_url: SERVER_HERO_URL,
  card_cover_url: MINECRAFT_COVER_URL,
  default_avatar_url: DEFAULT_USER_AVATAR_URL,
  default_server_icon_url: '/static/img/default-server-icon.png',
  footer_text: '[NeoPanel](https://github.com/Pekly/NeoPanel/)© 2026',
  preset: 'obsidian',
  accent_color: '#387aff',
  accent_hover: '#4f8aff',
  bg_deep: '',
  bg_canvas: '',
  card_bg: '',
  nav_bg: '',
  text_color: '',
  card_opacity: 72,
  border_radius_px: 14,
  glass_blur_px: 16,
  default_theme: 'dark',
  custom_css: '',
  maintenance_enabled: false,
  maintenance_title: 'Scheduled Maintenance',
  maintenance_message: 'We are upgrading NeoPanel nodes. Running game servers stay online during panel maintenance.',
  maintenance_eta: '',
  client_area_enabled: false,
  client_area_available: true,
  client_area_label: 'Client Area',
  client_area_url: '',
  client_area_open_mode: 'tab',
  client_area_subtitle: 'Purchase servers, manage subscriptions, and view invoices in your storefront.',
  show_client_area_button: true,
  timezone: 'UTC'
};

const BRAND_PRESETS = [
  {
    id: 'obsidian',
    name: 'Obsidian Navy',
    desc: 'Signature slate glass',
    accent_color: '#387aff',
    accent_hover: '#4f8aff',
    bg_deep: '',
    bg_canvas: '',
    card_bg: '',
    nav_bg: '',
    text_color: '',
    card_opacity: 72,
    border_radius_px: 14,
    glass_blur_px: 16,
    swatches: ['#0a0d14', '#121826', '#387aff']
  },
  {
    id: 'emerald',
    name: 'Emerald Craft',
    desc: 'Creeper green & moss',
    accent_color: '#10b981',
    accent_hover: '#34d399',
    bg_deep: '#07120e',
    bg_canvas: '#0b1914',
    card_bg: '#0f231b',
    nav_bg: '#060f0c',
    text_color: '#ecfdf5',
    card_opacity: 78,
    border_radius_px: 14,
    glass_blur_px: 16,
    swatches: ['#07120e', '#0f231b', '#10b981']
  },
  {
    id: 'amethyst',
    name: 'Amethyst Violet',
    desc: 'End crystal glow',
    accent_color: '#8b5cf6',
    accent_hover: '#a78bfa',
    bg_deep: '#0c0916',
    bg_canvas: '#120e21',
    card_bg: '#19142c',
    nav_bg: '#090712',
    text_color: '#f3f0ff',
    card_opacity: 76,
    border_radius_px: 16,
    glass_blur_px: 18,
    swatches: ['#0c0916', '#19142c', '#8b5cf6']
  },
  {
    id: 'crimson',
    name: 'Crimson Forge',
    desc: 'Redstone ember dark',
    accent_color: '#f43f5e',
    accent_hover: '#fb7185',
    bg_deep: '#12080c',
    bg_canvas: '#1a0c12',
    card_bg: '#231118',
    nav_bg: '#0e0609',
    text_color: '#fff1f2',
    card_opacity: 78,
    border_radius_px: 12,
    glass_blur_px: 16,
    swatches: ['#12080c', '#231118', '#f43f5e']
  },
  {
    id: 'amber',
    name: 'Solar Gold',
    desc: 'Blaze bronze & gold',
    accent_color: '#f59e0b',
    accent_hover: '#fbbf24',
    bg_deep: '#110d07',
    bg_canvas: '#1a130a',
    card_bg: '#241b0e',
    nav_bg: '#0c0905',
    text_color: '#fffbeb',
    card_opacity: 78,
    border_radius_px: 14,
    glass_blur_px: 16,
    swatches: ['#110d07', '#241b0e', '#f59e0b']
  },
  {
    id: 'nordic',
    name: 'Cyber Cyan',
    desc: 'Diamond ice & slate',
    accent_color: '#06b6d4',
    accent_hover: '#22d3ee',
    bg_deep: '#071016',
    bg_canvas: '#0b1821',
    card_bg: '#0f212d',
    nav_bg: '#050c11',
    text_color: '#ecfeff',
    card_opacity: 76,
    border_radius_px: 14,
    glass_blur_px: 16,
    swatches: ['#071016', '#0f212d', '#06b6d4']
  }
];

// --- Global Application State ---
window.S = {
  token: localStorage.getItem('neopanel_token') || '',
  user: null,
  view: 'dash',              // active view: dash, con, set, act, fi, db, bk, sc, net, sub, pl, plug, ver, admin, acc
  loading: {
    dash: true,
    files: false,
    plugins: false,
    backups: false,
    players: false,
    activity: false,
    databases: false
  },
  versionsCatalog: (() => {
    try { return JSON.parse(sessionStorage.getItem('neopanel_versions_catalog') || 'null'); } catch (e) { return null; }
  })(),
  viewModes: {},
  lastModrinthHits: null,
  navigating: false,
  flyConOpen: false,
  flyConTall: false,
  nodePairOpen: false,
  nodePairHelper: {
    name: 'mc-node-01',
    location: '',
    port: '8080',
    apiKey: ''
  },
  installingPlugins: new Set(),
  installingSoftware: null,
  skipPlugConfirm: (() => {
    try { return localStorage.getItem('neopanel_skip_plug_confirm') === 'true'; } catch (e) { return false; }
  })(),
  actFilter: {
    q: '',
    category: 'all',
    limit: 200
  },
  status: 'offline',         // online, offline, starting, stopping
  start: Date.now(),
  uptimeSec: 0,
  diskMB: 0,
  memUsedMB: 0,
  memMaxMB: 2048,
  cpu: Array.from({ length: 40 }, () => 0),
  ram: Array.from({ length: 40 }, () => 0),
  memSeriesMB: Array.from({ length: 40 }, () => 0),
  ni: Array.from({ length: 40 }, () => 0),
  no: Array.from({ length: 40 }, () => 0),
  seriesTs: Array.from({ length: 40 }, (_, i) => Date.now() - (39 - i) * 2000),
  chartHoverKind: null,
  chartHoverIdx: -1,
  notFoundPath: '',
  maintPreview: false,
  maintPublicInfo: null,
  maintAdminSignInOpen: false,
  clientIframeWarning: false,
  mods: [],
  adminMods: [],
  modsDir: '',
  selectedModId: '',
  modDraft: null,
  modsSafeMode: false,
  modPageId: '',
  tps: 20.00,
  ver: '1.21.11',
  core: 'Paper',
  sel: 'Paper',
  q: '',
  cat: 'All',
  plugTab: 'browse',
  plugLoader: '',
  plugVer: '',
  plugSort: 'relevance',
  plugFilter: '',
  ptab: 'online',
  ftab: 'server.properties',
  currentDir: '',
  fileList: [],
  sortedFileList: null,
  fileSort: (() => {
    try {
      const saved = JSON.parse(localStorage.getItem('neopanel_file_sort') || 'null');
      if (saved && ['name', 'ext', 'size', 'modified'].includes(saved.by)) {
        return { by: saved.by, dir: saved.dir === 'desc' ? 'desc' : 'asc' };
      }
    } catch (e) {}
    return { by: 'name', dir: 'asc' };
  })(),
  selectedFiles: new Set(),
  fileMenuOpen: null,
  subdomainData: null,
  subdomainDraft: null,
  subdomainLoading: false,
  subdomainVerify: null,
  subdomainVerifying: false,
  editSrvName: '',
  editSrvDesc: '',
  adminTab: 0,
  adminUsers: [],
  nodes: [],
  nodeFilter: 'all',
  clusterNodeErrors: {},
  databases: [],
  dbCapabilities: null,
  dbRevealIds: new Set(),
  sqlStudio: null,
  branding: { ...DEFAULT_BRANDING },
  brandDraft: null,
  brandingLoaded: false,
  brandingDaemon: {
    fqdn: '',
    sftp_port: 2022,
    max_upload_mb: 500,
    default_ram_mb: 2048,
    auto_restart_policy: true
  },
  daemonDraft: null,
  adminPanelUrl: null,
  panelUrlDraft: null,
  panelUrlLoading: false,
  panelUrlSaving: false,
  panelUrlError: '',
  panelUrlRequest: 0,
  adminTurnstile: {
    enabled: false,
    site_key: '',
    has_secret_key: false
  },
  turnstileEnabled: false,
  turnstileSiteKey: '',
  turnstileToken: '',
  turnstileWidgetId: null,
  adminDaemonCfg: JSON.parse(localStorage.getItem('neopanel_admin_cfg') || JSON.stringify({
    brandName: 'NeoPanel',
    fqdn: window.location.hostname || '127.0.0.1',
    sftpPort: 2022,
    maxUploadMB: 500,
    defaultRamMB: 2048,
    autoRestartPolicy: true
  })),
  players: [],
  playersMax: 20,
  playersOnlineCount: 0,
  bans: [],
  wl: [],
  ops: [],
  plugins: [],
  ports: [],
  bk: [],
  sc: [
    ['Restart server', '0 4 * * *', true],
    ['Backup world', '0 3 * * *', true],
    ['Broadcast voting reminder', '*/30 * * * *', false]
  ],
  props: {
    _srvId: null,
    motd: 'A Minecraft Server',
    max: 20,
    diff: 'normal',
    mode: 'survival',
    pvp: true,
    wl: false,
    view: 10,
    port: 25565,
    cmd: false
  },
  files: {},
  log: [],
  hist: [],
  hi: 0,
  cs: 0,
  servers: [],
  sftp: null,
  system: null,
  javaRuntimes: null,
  javaRuntimesMap: {},
  installingJava: null,
  startupFailures: 0,
  startupHalted: false,
  cmenu: false,
  pq: '',
  pi: 0,
  vq: '',
  vf: 'All Families',
  acct: {
    fn: 'Admin',
    ln: 'User',
    un: 'admin',
    em: 'admin@neopanel.internal',
    tab: 0,
    mode: localStorage.getItem('neopanel_mode') || 'Dark',
    compact: localStorage.getItem('neopanel_compact') === 'true',
    sound: localStorage.getItem('neopanel_sound') === 'true',
    priv: localStorage.getItem('neopanel_priv') === 'true',
    h: '',
    tfa: false,
    newkey: '',
    ssh: [],
    api: [],
    act: []
  },
  activeWs: null,
  activeServerId: null,
  activeNodeId: null,
  wsState: 'disconnected',
  wsAttempts: 0,
  wsReconnectTimer: null,
  wsPingTimer: null,
  pollTimer: null
};

// --- Web Audio Click/Beep ---
function beep() {
  if (!S.acct.sound) return;
  try {
    const ctx = new (window.AudioContext || window.webkitAudioContext)();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.type = 'sine';
    osc.frequency.setValueAtTime(660, ctx.currentTime);
    gain.gain.setValueAtTime(0.04, ctx.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.08);
    osc.connect(gain);
    gain.connect(ctx.destination);
    osc.start();
    osc.stop(ctx.currentTime + 0.08);
  } catch (e) {}
}

function toast(msg, type = 'info') {
  beep();
  const t = $('#toast');
  if (!t) return;
  t.textContent = msg;
  t.classList.add('s');
  clearTimeout(toast.timer);
  toast.timer = setTimeout(() => t.classList.remove('s'), 2600);
}

// --- Top Progress Bar & View Mode Switcher ---
let _progressTimer = null;
function startPageProgress(initialPct = 24) {
  const bar = $('#np-top-progress');
  if (!bar) return;
  clearTimeout(_progressTimer);
  bar.classList.remove('done');
  bar.classList.add('active');
  bar.style.width = `${Math.max(10, Math.min(90, initialPct))}%`;
}

function setPageProgress(pct) {
  const bar = $('#np-top-progress');
  if (!bar) return;
  bar.classList.remove('done');
  bar.classList.add('active');
  bar.style.width = `${Math.max(12, Math.min(94, pct))}%`;
}

function finishPageProgress() {
  const bar = $('#np-top-progress');
  if (!bar) return;
  clearTimeout(_progressTimer);
  bar.classList.add('active');
  bar.classList.add('done');
  bar.style.width = '100%';
  _progressTimer = setTimeout(() => {
    bar.classList.remove('active', 'done');
    bar.style.width = '0';
  }, 280);
}

function getViewMode(scope, fallback = 'list') {
  if (!S.viewModes) S.viewModes = {};
  if (S.viewModes[scope]) return S.viewModes[scope];
  try {
    const saved = localStorage.getItem(`neopanel_vm_${scope}`);
    if (saved === 'list' || saved === 'icons' || saved === 'tiles') {
      S.viewModes[scope] = saved;
      return saved;
    }
  } catch (e) {}
  return fallback;
}

function setViewMode(scope, mode) {
  if (!['list', 'icons', 'tiles'].includes(mode)) return;
  if (!S.viewModes) S.viewModes = {};
  S.viewModes[scope] = mode;
  try { localStorage.setItem(`neopanel_vm_${scope}`, mode); } catch (e) {}
  R();
  if (scope === 'plug_store' && S.view === 'plug' && (S.plugTab || 'browse') === 'browse') {
    const grid = $('#store-grid');
    if (grid) {
      if (Array.isArray(S.lastModrinthHits)) {
        grid.innerHTML = renderModrinthHitsHtml(S.lastModrinthHits);
        applyPermissionControls(grid);
      } else {
        searchModrinth();
      }
    }
  }
}

function renderViewModeToggle(scope, fallback = 'list') {
  const cur = getViewMode(scope, fallback);
  const opts = [
    ['list', 'viewList', 'List'],
    ['icons', 'viewIcons', 'Icons'],
    ['tiles', 'viewTiles', 'Tiles']
  ];
  return `<div class="vm-toggle" role="group" aria-label="View layout">
    ${opts.map(([m, icon, lbl]) => `
      <button type="button" class="vm-btn ${cur === m ? 'on' : ''}" onclick="setViewMode('${esc(scope)}', '${m}')" title="${lbl} view" aria-pressed="${cur === m}">
        ${ic(icon, 13)}
        <span>${lbl}</span>
      </button>
    `).join('')}
  </div>`;
}

// --- Clean Multi-Page In-Place Routing Helpers ---
const VIEW_TO_SLUG = {
  dash: '',
  con: 'console',
  set: 'settings',
  act: 'activity',
  fi: 'files',
  db: 'databases',
  bk: 'backups',
  sc: 'schedules',
  team: 'team',
  net: 'network',
  sftp: 'sftp',
  sub: 'subdomain',
  stp: 'settings',
  pl: 'players',
  plug: 'plugins',
  ver: 'versions',
  admin: 'admin',
  acc: 'account',
  client: 'client-area',
  mod: 'mod'
};

const SLUG_TO_VIEW = Object.fromEntries(
  Object.entries(VIEW_TO_SLUG)
    .filter(([k, slug]) => slug && k !== 'notfound' && k !== 'stp')
    .map(([k, slug]) => [slug, k])
);
SLUG_TO_VIEW.store = 'plug';
SLUG_TO_VIEW.mods = 'plug';
SLUG_TO_VIEW.startup = 'set';
SLUG_TO_VIEW.stp = 'set';
SLUG_TO_VIEW.servers = 'dash';
SLUG_TO_VIEW.subusers = 'team';

const ADMIN_TAB_SLUGS = ['health', 'nodes', 'servers', 'users', 'branding', 'audit', 'mods', 'email', 'roles'];
const ADMIN_SECTIONS = [
  { label: 'Overview', icon: 'activity', description: 'Panel health, resource usage, and updates.' },
  { label: 'Nodes', icon: 'network', description: 'Connect and manage the machines that run your servers.' },
  { label: 'Servers', icon: 'server', description: 'Manage all your servers from one place.' },
  { label: 'Users', icon: 'users', description: 'Manage panel accounts and access.' },
  { label: 'Appearance', icon: 'layers', description: 'Customize your branding, theme, and panel defaults.' },
  { label: 'Audit logs', icon: 'clock', description: 'Review activity and changes across your panel.' },
  { label: 'Extensions', icon: 'package', description: 'Add and manage panel extensions.' },
  { label: 'Email & SMTP', icon: 'mail', description: 'Set up email for invitations and password resets.' },
  { label: 'Roles', icon: 'shield', description: 'Named panel permissions, separate from Administrator access.' }
];
let adminStatusRefreshing = false;
async function refreshAdminStatus(button) {
  if (adminStatusRefreshing) return;
  adminStatusRefreshing = true;
  setBtnLoading(button, true, 'Refreshing…');
  try {
    await Promise.all([loadDashboard(true), loadAdminData(true)]);
    S.adminRefreshedAt = Date.now();
  } finally {
    adminStatusRefreshing = false;
    if (S.view === 'admin') R();
    else if (button?.isConnected) setBtnLoading(button, false);
  }
}

function displayNodeName(nodeId, name = '') {
  const label = String(name || '').trim();
  if ((nodeId || 'local') === 'local' && (!label || ['local', 'Local Master Node', 'Local Master'].includes(label))) return 'Local node';
  if (label && label !== nodeId) return label;
  const node = (S.nodes || []).find(n => n.id === nodeId);
  return (node?.name && node.name !== nodeId ? node.name : node?.location && node.location !== nodeId ? node.location : 'Remote node');
}
const ACC_TAB_SLUGS = ['general', 'security', 'ssh', 'api', 'activity'];

function shortServerId(id) {
  return String(id || '').trim().replace(/^srv-+/i, '');
}

function resolveServerByParam(param, nodeId = null) {
  const p = String(param || '').trim();
  if (!p) return null;
  const shortP = shortServerId(p);
  const nid = nodeId ? String(nodeId).trim() : null;
  const list = Array.isArray(S.servers) ? S.servers : [];
  const matches = list.filter(s => (!nid || (s.node_id || 'local') === nid) &&
    (s.id === p || shortServerId(s.id) === shortP));
  // A node-qualified URL must never silently open a different node's server.
  if (nid) return matches.length === 1 ? matches[0] : null;
  return matches.find(s => (s.node_id || 'local') === 'local') || (matches.length === 1 ? matches[0] : null);
}

// Absence is authoritative only after the relevant authorized inventory loaded.
function missingServerView(route, url, attempts = 0) {
  const catalog = S.serverCatalog;
  const complete = catalog && !catalog.loading && (catalog.denied || (route.nodeId === 'local'
    ? catalog.localComplete
    : catalog.clusterComplete && (route.nodeId
      ? !Object.prototype.hasOwnProperty.call(catalog.nodeErrors, route.nodeId)
      : Object.keys(catalog.nodeErrors).length === 0)));
  clearActiveServerSelection();
  if (complete) {
    S.notFoundPath = new URL(url, window.location.origin).pathname;
    S.serverRoutePending = null;
    return 'notfound';
  }
  S.serverRoutePending = { url, attempts, busy: false, token: S.token };
  return 'serverpending';
}

function scheduleServerRouteRetry() {
  const pending = S.serverRoutePending;
  if (!pending || pending.attempts >= 2 || pending.timer || pending.busy) return;
  pending.timer = setTimeout(() => {
    pending.timer = null;
    if (S.serverRoutePending === pending) retryServerRoute();
  }, 1000 * (pending.attempts + 1));
}

async function retryServerRoute() {
  const pending = S.serverRoutePending;
  if (!pending || pending.busy || !S.token || S.token !== pending.token) return;
  clearTimeout(pending.timer);
  pending.timer = null;
  pending.busy = true;
  pending.attempts++;
  R();
  await loadDashboard(true);
  if (S.serverRoutePending !== pending || S.token !== pending.token) return;
  applyRoute(pending.url, { replace: true, retryAttempts: pending.attempts });
}

function clearActiveServerSelection() {
  S.cs = -1;
  S.activeServerId = null;
  S.activeNodeId = null;
  clearTimeout(S.wsReconnectTimer);
  clearInterval(S.wsPingTimer);
  S.wsReconnectTimer = S.wsPingTimer = null;
  if (S.activeWs) {
    S.activeWs.onopen = S.activeWs.onmessage = S.activeWs.onerror = S.activeWs.onclose = null;
    try { S.activeWs.close(); } catch (e) {}
    S.activeWs = null;
  }
  S.wsState = 'disconnected';
}

function saveActiveServerRef(srv) {
  if (!srv || !srv.id) return;
  try {
    localStorage.setItem('neopanel_active_srv', JSON.stringify({
      id: srv.id,
      nodeId: srv.node_id || 'local'
    }));
  } catch (e) {}
}

function getSavedActiveServerRef() {
  try {
    return JSON.parse(localStorage.getItem('neopanel_active_srv') || 'null');
  } catch (e) {
    return null;
  }
}

function withSafeModeQuery(url) {
  if (!S.modsSafeMode || !url) return url;
  if (String(url).includes('safemode=1')) return url;
  return url + (String(url).includes('?') ? '&' : '?') + 'safemode=1';
}

function serverRouteUrl(nodeId, srvId, view = 'con', opts = {}) {
  const nid = String(nodeId || 'local').trim() || 'local';
  const target = normalizeRoute(view);
  const shortId = shortServerId(srvId);
  let path;
  if (target === 'plug') {
    const tab = opts.plugTab || S.plugTab || 'browse';
    const cur = resolveServerByParam(srvId, nid) || curServer();
    const plugSlug = isModdedEngine(cur?.type || S.core) ? 'mods' : 'plugins';
    path = `/server/${encodeURIComponent(shortId)}/${plugSlug}/${tab === 'installed' ? 'installed' : 'store'}`;
  } else {
    const slug = VIEW_TO_SLUG[target] || 'console';
    path = `/server/${encodeURIComponent(shortId)}/${slug}`;
  }
  if (nid !== 'local') {
    path += `?node=${encodeURIComponent(nid)}`;
  }
  return withSafeModeQuery(path);
}

function routeUrl(view, opts = {}) {
  const target = normalizeRoute(view);
  if (target === 'create') return withSafeModeQuery('/servers/create');
  if (target === 'dash') return withSafeModeQuery('/');
  if (target === 'client') return withSafeModeQuery('/client-area');
  if (target === 'mod') {
    const mid = opts.modPageId || opts.modId || S.modPageId || '';
    return withSafeModeQuery(mid ? `/mod/${encodeURIComponent(mid)}` : '/');
  }
  if (target === 'notfound') {
    return withSafeModeQuery(opts.notFoundPath || window.location.pathname || '/not-found');
  }
  if (target === 'admin') {
    const t = opts.adminTab !== undefined ? opts.adminTab : (S.adminTab || 0);
    const slug = ADMIN_TAB_SLUGS[t] || 'health';
    return withSafeModeQuery(t === 0 ? '/admin' : `/admin/${slug}`);
  }
  if (target === 'acc') {
    const t = opts.acctTab !== undefined ? opts.acctTab : (S.acct?.tab || 0);
    const slug = ACC_TAB_SLUGS[t] || 'general';
    return withSafeModeQuery(t === 0 ? '/account' : `/account/${slug}`);
  }
  if (target === 'act' && (opts.global || (!opts.srv && !window.location.pathname.startsWith('/server/') && S.view === 'dash'))) {
    return withSafeModeQuery('/activity');
  }
  const srv = opts.srv || curServer();
  if (!srv || !srv.id) {
    return withSafeModeQuery(target === 'act' ? '/activity' : '/');
  }
  return serverRouteUrl(srv.node_id || 'local', srv.id, target, opts);
}

function parseCurrentRoute(urlOrPath = window.location.pathname + window.location.search, hashStr = window.location.hash) {
  let u;
  try {
    u = new URL(urlOrPath || '/', window.location.origin);
  } catch (e) {
    u = new URL('/', window.location.origin);
  }
  const cleanPath = String(u.pathname || '/').replace(/\/+$/, '') || '/';
  const rawHash = String(u.hash || hashStr || '').replace(/^#/, '').trim();
  const safeMode = u.searchParams.get('safemode') === '1';
  const nodeParam = u.searchParams.get('node') ? decodeURIComponent(u.searchParams.get('node')) : null;

  if (cleanPath === '/' || cleanPath === '/index.html') {
    if (rawHash) {
      const hView = normalizeRoute(rawHash);
      if (hView && V[hView] && hView !== 'notfound' && hView !== 'maintenance') {
        return { view: hView, plugTab: rawHash === 'store' ? 'browse' : undefined, legacyHash: true, safeMode };
      }
      return { view: 'notfound', notFoundPath: `${cleanPath}#${rawHash}`, safeMode };
    }
    return { view: 'dash', safeMode };
  }
  if (cleanPath === '/servers') return { view: 'dash', safeMode };
  if (['/create', '/servers/create', '/server/create'].includes(cleanPath)) return { view: 'create', safeMode };
  if (cleanPath === '/activity') return { view: 'act', global: true, safeMode };
  if (cleanPath === '/client-area') return { view: 'client', safeMode };
  if (cleanPath === '/billing') return { view: 'notfound', notFoundPath: cleanPath, safeMode };

  if (cleanPath.startsWith('/mod/')) {
    const parts = cleanPath.split('/').filter(Boolean);
    if (parts.length !== 2) {
      return { view: 'notfound', notFoundPath: cleanPath, safeMode };
    }
    return { view: 'mod', modPageId: decodeURIComponent(parts[1] || ''), safeMode };
  }

  if (cleanPath === '/admin' || cleanPath.startsWith('/admin/')) {
    const parts = cleanPath.split('/').filter(Boolean);
    if (parts.length > 2) return { view: 'notfound', notFoundPath: cleanPath, safeMode };
    const sub = parts[1] || 'health';
    const idx = ADMIN_TAB_SLUGS.indexOf(sub);
    if (idx < 0) return { view: 'notfound', notFoundPath: cleanPath, safeMode };
    return { view: 'admin', adminTab: idx, safeMode };
  }

  if (cleanPath === '/account' || cleanPath.startsWith('/account/')) {
    const parts = cleanPath.split('/').filter(Boolean);
    if (parts.length > 2) return { view: 'notfound', notFoundPath: cleanPath, safeMode };
    const sub = parts[1] || 'general';
    const idx = ACC_TAB_SLUGS.indexOf(sub);
    if (idx < 0) return { view: 'notfound', notFoundPath: cleanPath, safeMode };
    return { view: 'acc', acctTab: idx, safeMode };
  }

  if (cleanPath.startsWith('/server/')) {
    const parts = cleanPath.split('/').filter(Boolean);
    if (parts.length < 2) return { view: 'notfound', notFoundPath: cleanPath, safeMode };

    // Clean short format: /server/{id}/{slug?}[/installed|store]
    // Legacy format: /server/{nodeId}/{id}/{slug?}[/installed|store]
    const rawP1 = decodeURIComponent(parts[1] || '');
    const p2Lower = (parts[2] || '').toLowerCase();
    const isShortTwoSeg = parts.length === 2;
    const isShortWithKnownSlug = parts.length >= 3 && (SLUG_TO_VIEW[p2Lower] || p2Lower === 'plugins' || p2Lower === 'mods');

    if (isShortTwoSeg || isShortWithKnownSlug) {
      const serverId = rawP1;
      const slug = (parts[2] || 'console').toLowerCase();
      const extra = (parts[3] || '').toLowerCase();
      if (slug === 'plugins' || slug === 'mods') {
        if (parts.length > 4 || (extra && !['store', 'installed'].includes(extra))) {
          return { view: 'notfound', notFoundPath: cleanPath, safeMode };
        }
        return {
          view: 'plug',
          nodeId: nodeParam,
          serverId,
          plugTab: extra === 'installed' ? 'installed' : 'browse',
          canonicalRedirect: isShortTwoSeg || rawP1 !== shortServerId(rawP1) || (parts.length === 3),
          safeMode
        };
      }
      if (parts.length > 3) return { view: 'notfound', notFoundPath: cleanPath, safeMode };
      const mappedView = SLUG_TO_VIEW[slug];
      if (!mappedView) return { view: 'notfound', notFoundPath: cleanPath, safeMode };
      return {
        view: mappedView,
        nodeId: nodeParam,
        serverId,
        canonicalRedirect: isShortTwoSeg || rawP1 !== shortServerId(rawP1) || slug === 'startup',
        safeMode
      };
    }

    // Legacy 3+ segment: /server/{nodeId}/{serverId}/...
    if (parts.length >= 3) {
      const nodeId = rawP1 || 'local';
      const serverId = decodeURIComponent(parts[2] || '');
      const slug = (parts[3] || 'console').toLowerCase();
      const extra = (parts[4] || '').toLowerCase();
      if (slug === 'plugins' || slug === 'mods') {
        if (parts.length > 5 || (extra && !['store', 'installed'].includes(extra))) {
          return { view: 'notfound', notFoundPath: cleanPath, safeMode };
        }
        return {
          view: 'plug',
          nodeId,
          serverId,
          plugTab: extra === 'installed' ? 'installed' : 'browse',
          canonicalRedirect: true,
          safeMode
        };
      }
      if (parts.length > 4) return { view: 'notfound', notFoundPath: cleanPath, safeMode };
      const mappedView = SLUG_TO_VIEW[slug];
      if (!mappedView) return { view: 'notfound', notFoundPath: cleanPath, safeMode };
      return {
        view: mappedView,
        nodeId,
        serverId,
        canonicalRedirect: true,
        safeMode
      };
    }
  }

  return { view: 'notfound', notFoundPath: cleanPath, safeMode };
}

function getViewDisplayLabel(view) {
  const isMod = isModdedEngine(curServer()?.type || S.core);
  const map = {
    dash: 'Servers',
    con: 'Console',
    set: 'Settings',
    act: 'Activity',
    fi: 'Files',
    db: 'Databases',
    bk: 'Backups',
    sc: 'Schedules',
    team: 'Team',
    net: 'Network',
    sub: 'Subdomain',
    stp: 'Settings',
    pl: 'Players',
    plug: isMod ? 'Mods' : 'Plugins',
    ver: 'Software',
    admin: 'Admin',
    acc: 'Account',
    client: getBrand('client_area_label') || 'Client Area',
    notfound: 'Page Not Found',
    serverpending: 'Server Unavailable',
    maintenance: 'Scheduled Maintenance',
    mod: 'Extension Page'
  };
  return map[view] || 'NeoPanel';
}

function syncActiveServerState(srv) {
  if (!srv || !srv.id) return;
  captureEditor();
  const idx = S.servers.findIndex(x => x.id === srv.id && (x.node_id || 'local') === (srv.node_id || 'local'));
  if (idx >= 0) S.cs = idx;
  const nodeId = srv.node_id || 'local';
  const changed = S.activeServerId !== srv.id || (S.activeNodeId || 'local') !== nodeId;

  saveActiveServerRef(srv);
  applyServerStatsSnapshot(srv, srv.stats, false);
  syncPropsFromServer(srv, changed);
  S.ver = srv.version || S.ver;
  S.core = engineMeta(srv.type || 'paper').label;

  if (changed) {
    ++QOL.openSequence;
    if (typeof disposeFilePreview === 'function') disposeFilePreview();
    QOL.editor = null;
    ++TeamUI.sequence;
    S.teamState = null;
    S.teamLoading = false;
    S.serverSubusers = [];
    clearSettingsAutoSaveTimer();
    _settingsUiState = { status: 'saved', text: '' };
    S.log = [];
    S.ports = [];
    S.databases = [];
    S.dbRevealIds = new Set();
    S.sqlStudio = null;
    S.fileList = [];
    S.sortedFileList = null;
    S.currentDir = '';
    S.editingFile = null;
    S.selectedFiles = new Set();
    S.bk = [];
    S.plugins = [];
    S.lastModrinthHits = null;
    S.plugLoader = '';
    S.plugVer = '';
    S.players = [];
    S.wl = [];
    S.bans = [];
    S.ops = [];
    S.playersMax = 20;
    S.playersOnlineCount = 0;
    S.subdomainData = null;
    S.subdomainDraft = null;
    S.subdomainVerify = null;
    S.editSrvName = srv.name || '';
    S.editSrvDesc = srv.description || '';
    S.sftp = null;
    S.cpu = Array.from({ length: 40 }, () => 0);
    S.ram = Array.from({ length: 40 }, () => 0);
    S.memSeriesMB = Array.from({ length: 40 }, () => 0);
    S.ni = Array.from({ length: 40 }, () => 0);
    S.no = Array.from({ length: 40 }, () => 0);
    S.seriesTs = Array.from({ length: 40 }, (_, i) => Date.now() - (39 - i) * 2000);
    S.chartHoverKind = null;
    S.chartHoverIdx = -1;
    S.startupFailures = Number(srv.stats?.startup_failures) || 0;
    S.startupHalted = Boolean(srv.stats?.startup_halted);
    applyServerStatsSnapshot(srv, srv.stats, false);
    resetServerAccess();
    if (window.NP && typeof window.NP._emit === 'function') {
      window.NP._emit('server:switch', { server: srv });
    }
  }
}

function applyRoute(urlStr, opts = {}) {
  captureEditor();
  if (!opts.editorChecked && (editorIsDirty() || QOL.editor?.saving || QOL.leaving)) {
    return guardEditorLeave().then(allowed => {
      if (allowed) applyRoute(urlStr, { ...opts, editorChecked: true });
      else if (opts.fromPopState) restoreEditorHistory(opts.historyIndex);
    });
  }
  ++QOL.openSequence;
  clearTimeout(S.serverRoutePending?.timer);
  S.serverRoutePending = null;
  const { replace = false, fromPopState = false } = opts;
  if (fromPopState && Number.isInteger(opts.historyIndex)) QOL.historyIndex = opts.historyIndex;
  flushPendingSettingsBeforeLeave();
  const prevPath = window.location.pathname;
  const prevView = S.view;
  const prevAdminTab = S.adminTab || 0;

  closeAllCtx();
  closeFloatingCtx();
  closeAllCustomSelects();
  S.cmenu = false;
  S.chartHoverKind = null;
  S.chartHoverIdx = -1;

  const targetUrlStr = withSafeModeQuery(urlStr || '/');
  const route = parseCurrentRoute(targetUrlStr);
  if (route.safeMode) S.modsSafeMode = true;

  if (prevView === 'admin' && prevAdminTab === 4 && (route.view !== 'admin' || (route.adminTab || 0) !== 4)) {
    applyBranding(S.branding);
  }

  let resolvedView = normalizeRoute(route.view || 'dash');
  let finalUrl = targetUrlStr;

  if (route.notFoundPath) {
    S.notFoundPath = route.notFoundPath;
    resolvedView = 'notfound';
  } else if (resolvedView === 'mod') {
    const pid = route.modPageId || '';
    if (!window.NP || !window.NP._pages || !window.NP._pages.has(pid)) {
      S.notFoundPath = `/mod/${pid}`;
      resolvedView = 'notfound';
    } else {
      S.modPageId = pid;
      finalUrl = routeUrl('mod', { modPageId: pid });
    }
  } else if (route.serverId) {
    const srv = resolveServerByParam(route.serverId, route.nodeId || null);
    if (!srv) {
      resolvedView = missingServerView(route, targetUrlStr, opts.retryAttempts || 0);
    } else {
      syncActiveServerState(srv);
      if (route.plugTab !== undefined) S.plugTab = route.plugTab;
      if (route.canonicalRedirect) {
        finalUrl = serverRouteUrl(srv.node_id || 'local', srv.id, resolvedView, { plugTab: S.plugTab });
      }
    }
  } else {
    S.notFoundPath = '';
    if (route.adminTab !== undefined) S.adminTab = route.adminTab;
    if (route.acctTab !== undefined && S.acct) S.acct.tab = route.acctTab;
    if (route.plugTab !== undefined) S.plugTab = route.plugTab;
  }

  if (resolvedView === 'admin') {
    if (S.user && S.user.role !== 'admin') {
      toast('Administrator privileges required', 'bad');
      resolvedView = 'dash';
      finalUrl = withSafeModeQuery('/');
    } else {
      S.activeServerId = null;
      S.activeNodeId = null;
    }
  }

  S.view = resolvedView;
  if (route.adminTab !== undefined && resolvedView === 'admin') S.adminTab = route.adminTab;
  if (route.acctTab !== undefined && resolvedView === 'acc' && S.acct) S.acct.tab = route.acctTab;

  if (!fromPopState) {
    try {
      const curFull = window.location.pathname + window.location.search;
      if (curFull !== finalUrl) {
        if (replace || route.canonicalRedirect || route.legacyHash) {
          history.replaceState(null, '', finalUrl);
        } else {
          history.pushState(null, '', finalUrl);
        }
      }
    } catch (e) {}
  }

  startPageProgress(38);
  R();
  const mainEl = $('#main');
  if (mainEl && prevPath !== window.location.pathname) {
    mainEl.scrollTop = 0;
  }
  if (window.NP && typeof window.NP._emit === 'function') {
    window.NP._emit('route', {
      view: S.view,
      prevView,
      url: finalUrl,
      server: curServer()
    });
  }
  onEnterView(S.view);
  scheduleServerRouteRetry();
  finishPageProgress();
}

function navigateToUrl(url, label = 'Page', opts = {}) {
  if (!url) return;
  applyRoute(url, opts);
}

function handleNavClick(e, view, opts = {}) {
  if (serverPagePermission(view) && !canServerPage(view)) { if (e) e.preventDefault(); return false; }
  if (e && (e.ctrlKey || e.metaKey || e.shiftKey || e.altKey || e.button !== 0)) {
    return true;
  }
  if (e) e.preventDefault();
  const target = normalizeRoute(view);
  if (curServer()) saveActiveServerRef(curServer());
  if (target === 'plug' && view === 'store') opts.plugTab = 'browse';
  applyRoute(routeUrl(target, opts));
  return false;
}

function handleServerCardClick(e, index, targetView = 'con') {
  if (e && (e.ctrlKey || e.metaKey || e.shiftKey || e.altKey || e.button !== 0)) {
    return true;
  }
  if (e) e.preventDefault();
  const srv = S.servers[index];
  if (!srv) return false;
  applyRoute(serverRouteUrl(srv.node_id || 'local', srv.id, targetView));
  return false;
}

window.addEventListener('popstate', event => {
  const historyIndex = event.state?.npQolIndex;
  if (Number.isInteger(historyIndex) && QOL.restoringHistory === historyIndex) {
    QOL.restoringHistory = null;
    return;
  }
  if (!S.token || $('#login')) return;
  applyRoute(window.location.pathname + window.location.search + window.location.hash, {
    fromPopState: true,
    historyIndex,
    replace: true
  });
});

// --- Custom Modal Dialog System (Replacing browser default alert, confirm, prompt) ---
let _uiModalResolver = null;

function closeUiModal(value) {
  closeAllCustomSelects();
  const container = $('#modal-container');
  if (container) container.innerHTML = '';
  if (_uiModalResolver) {
    const resolve = _uiModalResolver;
    _uiModalResolver = null;
    resolve(value);
  }
}

function uiAlert(opts, titleFallback) {
  return new Promise(resolve => {
    let title = 'Notice';
    let message = '';
    let confirmText = 'OK';
    let iconName = 'info';

    if (typeof opts === 'string') {
      message = opts;
      if (titleFallback) title = titleFallback;
    } else if (opts && typeof opts === 'object') {
      title = opts.title || title;
      message = opts.message || '';
      confirmText = opts.confirmText || confirmText;
      iconName = opts.icon || (opts.type === 'error' ? 'alert' : 'info');
    }

    if (_uiModalResolver) {
      _uiModalResolver(null);
    }
    _uiModalResolver = () => resolve(true);

    const container = $('#modal-container') || document.body;
    container.innerHTML = `
      <div class="pal" onclick="if(event.target===this) closeUiModal(true)">
        <div class="pbox ui-modal-box">
          <div class="row sb" style="margin-bottom: 14px; align-items: center;">
            <h3 style="font-size: 17px; font-weight: 600; margin: 0; display: flex; align-items: center; gap: 8px; color: var(--tx);">
              ${ic(iconName, 18)} <span>${esc(title)}</span>
            </h3>
            <button class="btn sm" onclick="closeUiModal(true)" aria-label="Close" style="padding: 4px 8px; font-size: 15px; opacity: 0.8;">&times;</button>
          </div>
          <div class="ui-modal-body">
            ${message}
          </div>
          <div class="row" style="justify-content: flex-end; gap: 10px; margin-top: 20px;">
            <button type="button" id="ui-modal-confirm-btn" class="btn pr" onclick="closeUiModal(true)">
              ${esc(confirmText)}
            </button>
          </div>
        </div>
      </div>
    `;

    setTimeout(() => $('#ui-modal-confirm-btn')?.focus(), 40);
  });
}

function uiConfirm(opts, titleFallback) {
  return new Promise(resolve => {
    let title = 'Confirm Action';
    let message = '';
    let confirmText = 'Confirm';
    let cancelText = 'Cancel';
    let isDanger = false;
    let iconName = 'alert';

    if (typeof opts === 'string') {
      message = opts;
      if (titleFallback) title = titleFallback;
      if (opts.toLowerCase().includes('delete') || opts.toLowerCase().includes('remove')) {
        isDanger = true;
      }
    } else if (opts && typeof opts === 'object') {
      title = opts.title || title;
      message = opts.message || '';
      confirmText = opts.confirmText || confirmText;
      cancelText = opts.cancelText || cancelText;
      isDanger = Boolean(opts.isDanger);
      iconName = opts.icon || (isDanger ? 'trash' : 'alert');
    }

    if (_uiModalResolver) {
      _uiModalResolver(false);
    }
    _uiModalResolver = (val) => resolve(Boolean(val));

    const titleColor = isDanger ? '#ef4444' : 'var(--tx)';
    const confirmBtnClass = isDanger ? 'btn dg' : 'btn pr';
    const container = $('#modal-container') || document.body;

    container.innerHTML = `
      <div class="pal" onclick="if(event.target===this) closeUiModal(false)">
        <div class="pbox ui-modal-box">
          <div class="row sb" style="margin-bottom: 14px; align-items: center;">
            <h3 style="font-size: 17px; font-weight: 600; margin: 0; display: flex; align-items: center; gap: 8px; color: ${titleColor};">
              ${ic(iconName, 18)} <span>${esc(title)}</span>
            </h3>
            <button class="btn sm" onclick="closeUiModal(false)" aria-label="Close" style="padding: 4px 8px; font-size: 15px; opacity: 0.8;">&times;</button>
          </div>
          <div class="ui-modal-body">
            ${message}
          </div>
          <div class="row" style="justify-content: flex-end; gap: 10px; margin-top: 20px;">
            <button type="button" class="btn" onclick="closeUiModal(false)">${esc(cancelText)}</button>
            <button type="button" id="ui-modal-confirm-btn" class="${confirmBtnClass}" onclick="closeUiModal(true)">
              ${esc(confirmText)}
            </button>
          </div>
        </div>
      </div>
    `;

    setTimeout(() => $('#ui-modal-confirm-btn')?.focus(), 40);
  });
}

function uiPrompt(opts, defaultValFallback, titleFallback) {
  return new Promise(resolve => {
    let title = 'Input Required';
    let message = '';
    let defaultValue = '';
    let placeholder = '';
    let confirmText = 'Submit';
    let cancelText = 'Cancel';
    let inputType = 'text';

    if (typeof opts === 'string') {
      message = opts;
      if (defaultValFallback != null) defaultValue = defaultValFallback;
      if (titleFallback) title = titleFallback;
    } else if (opts && typeof opts === 'object') {
      title = opts.title || title;
      message = opts.message || '';
      defaultValue = opts.defaultValue != null ? String(opts.defaultValue) : '';
      placeholder = opts.placeholder || '';
      confirmText = opts.confirmText || confirmText;
      cancelText = opts.cancelText || cancelText;
      inputType = opts.inputType || 'text';
    }

    if (_uiModalResolver) {
      _uiModalResolver(null);
    }
    _uiModalResolver = (val) => resolve(val != null ? String(val) : null);

    const container = $('#modal-container') || document.body;
    container.innerHTML = `
      <div class="pal" onclick="if(event.target===this) closeUiModal(null)">
        <div class="pbox ui-modal-box">
          <div class="row sb" style="margin-bottom: 14px; align-items: center;">
            <h3 style="font-size: 17px; font-weight: 600; margin: 0; display: flex; align-items: center; gap: 8px; color: var(--tx);">
              ${ic('edit', 18)} <span>${esc(title)}</span>
            </h3>
            <button class="btn sm" onclick="closeUiModal(null)" aria-label="Close" style="padding: 4px 8px; font-size: 15px; opacity: 0.8;">&times;</button>
          </div>
          ${message ? `<div class="ui-modal-body" style="margin-bottom: 14px;">${message}</div>` : ''}
          <form id="ui-modal-prompt-form" onsubmit="event.preventDefault(); submitUiPrompt();" style="margin: 0;">
            <input type="${esc(inputType)}" id="ui-modal-input" value="${esc(defaultValue)}" placeholder="${esc(placeholder)}"
              autocomplete="off" spellcheck="false"
              class="ui-modal-input">
            <div class="row" style="justify-content: flex-end; gap: 10px; margin-top: 18px;">
              <button type="button" class="btn" onclick="closeUiModal(null)">${esc(cancelText)}</button>
              <button type="submit" id="ui-modal-submit-btn" class="btn pr">${esc(confirmText)}</button>
            </div>
          </form>
        </div>
      </div>
    `;

    setTimeout(() => {
      const inp = $('#ui-modal-input');
      if (inp) {
        inp.focus();
        inp.select();
      }
    }, 40);
  });
}

function submitUiPrompt() {
  const inp = $('#ui-modal-input');
  const val = inp ? inp.value : '';
  closeUiModal(val);
}

// Override native window.alert to always use modern custom modal
window.uiAlert = uiAlert;
window.uiConfirm = uiConfirm;
window.uiPrompt = uiPrompt;
window.closeUiModal = closeUiModal;
window.submitUiPrompt = submitUiPrompt;
window.alert = (msg) => { uiAlert(msg); };

const clock = () => new Date().toTimeString().slice(0, 8);

function formatBytes(bytes) {
  if (!bytes || bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KiB', 'MiB', 'GiB', 'TiB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return (bytes / Math.pow(k, i)).toFixed(i > 1 ? 2 : 0) + ' ' + sizes[i];
}

function getServerMaxRamMB(srv) {
  if (!srv) return 2048;
  let val = srv.memory_max || srv.memory_limit || srv.memory || 2048;
  if (typeof val === 'string') {
    val = val.trim();
    if (val.toLowerCase().endsWith('g')) {
      const parsedG = parseFloat(val);
      return !isNaN(parsedG) && parsedG > 0 ? Math.round(parsedG * 1024) : 2048;
    }
    if (val.toLowerCase().endsWith('m')) {
      const parsedM = parseFloat(val);
      return !isNaN(parsedM) && parsedM > 0 ? Math.round(parsedM) : 2048;
    }
    const parsed = parseInt(val, 10);
    return !isNaN(parsed) && parsed > 0 ? (parsed < 64 ? parsed * 1024 : parsed) : 2048;
  }
  if (typeof val === 'number' && !isNaN(val) && val > 0) {
    return val < 64 ? Math.round(val * 1024) : Math.round(val);
  }
  return 2048;
}

function formatDiskMB(mb) {
  const val = Number(mb) || 0;
  if (val <= 0) return '0 MB';
  if (val >= 1024) return `${(val / 1024).toFixed(2)} GiB`;
  return `${Math.round(val)} MB`;
}

function formatUptime(sec) {
  const s = Math.max(0, Math.floor(Number(sec) || 0));
  if (s < 60) return '< 1m';
  const m = Math.floor(s / 60) % 60;
  const h = Math.floor(s / 3600);
  const d = Math.floor(s / 86400);
  if (d > 0) return `${d}d ${h % 24}h`;
  if (h > 0) return `${h}h ${m}m`;
  return `${m}m`;
}

function up() {
  if (S.status !== 'online' && S.status !== 'starting') return 'Not running';
  return formatUptime(S.uptimeSec);
}

function normalizeServerStatus(rawStatus) {
  const st = String(rawStatus || 'offline').toLowerCase();
  if (st === 'online' || st === 'running') return 'online';
  if (st === 'starting' || st === 'stopping') return st;
  return 'offline';
}

function applyServerStatsSnapshot(srv, stats, pushSeries = false) {
  if (!stats) return;
  const prevHalted = S.startupHalted;
  const normStatus = normalizeServerStatus(stats.status);
  S.status = normStatus;
  S.uptimeSec = Math.max(0, Number(stats.uptime_sec) || 0);
  S.start = Date.now() - S.uptimeSec * 1000;
  S.diskMB = Math.max(0, Number(stats.disk_mb) || 0);
  S.startupFailures = Math.max(0, Number(stats.startup_failures) || 0);
  S.startupHalted = Boolean(stats.startup_halted);

  if (S.startupHalted && !prevHalted && normStatus === 'offline') {
    toast('Startup timed out after 2 failed attempts — check Console logs.', 'bad');
  }

  const cpuVal = Math.max(0, Math.min(100, Number(stats.cpu_percent ?? stats.cpu_usage) || 0));
  const memMB = Math.max(0, Number(stats.memory_mb) || 0);
  const maxMemMB = Math.max(1, Number(stats.memory_max_mb) || getServerMaxRamMB(srv));
  const ramPct = Math.max(0, Math.min(100, (memMB / maxMemMB) * 100));
  const netIn = Math.max(0, Number(stats.net_in_kib) || 0);
  const netOut = Math.max(0, Number(stats.net_out_kib) || 0);
  const nowTs = Date.now();

  S.memUsedMB = memMB;
  S.memMaxMB = maxMemMB;
  if (!Array.isArray(S.memSeriesMB) || S.memSeriesMB.length !== S.cpu.length) {
    S.memSeriesMB = S.ram.map(pct => (pct / 100) * maxMemMB);
  }
  if (!Array.isArray(S.seriesTs) || S.seriesTs.length !== S.cpu.length) {
    S.seriesTs = Array.from({ length: S.cpu.length }, (_, i) => nowTs - (S.cpu.length - 1 - i) * 2000);
  }

  if (pushSeries) {
    S.cpu.push(cpuVal);
    S.cpu.shift();
    S.ram.push(ramPct);
    S.ram.shift();
    S.memSeriesMB.push(memMB);
    S.memSeriesMB.shift();
    S.ni.push(netIn);
    S.ni.shift();
    S.no.push(netOut);
    S.no.shift();
    S.seriesTs.push(nowTs);
    S.seriesTs.shift();
  } else {
    S.cpu[S.cpu.length - 1] = cpuVal;
    S.ram[S.ram.length - 1] = ramPct;
    S.memSeriesMB[S.memSeriesMB.length - 1] = memMB;
    S.ni[S.ni.length - 1] = netIn;
    S.no[S.no.length - 1] = netOut;
    S.seriesTs[S.seriesTs.length - 1] = nowTs;
  }

  if (srv) {
    srv.stats = {
      ...(srv.stats || {}),
      status: normStatus,
      cpu_usage: cpuVal,
      memory_mb: memMB,
      memory_max_mb: maxMemMB,
      disk_mb: S.diskMB,
      uptime_sec: S.uptimeSec,
      net_in_kib: netIn,
      net_out_kib: netOut,
      startup_failures: S.startupFailures,
      startup_halted: S.startupHalted
    };
  }
  if (pushSeries && window.NP && typeof window.NP._emit === 'function') {
    window.NP._emit('stats', { server: srv, stats: srv?.stats || stats });
  }
}

function copyServerAddress(addr, ev) {
  if (ev) {
    ev.preventDefault();
    ev.stopPropagation();
  }
  if (!addr) return false;
  copyDbText(addr, 'Server Address');
  return false;
}

// --- Sparklines & Interactive Hover Charts ---
function spark(data, max) {
  if (!data || data.length === 0) data = [0];
  const m = max || Math.max(...data, 1);
  const p = data.map((v, i) => `${i * (300 / (data.length - 1 || 1))},${70 - (v / m) * 66}`).join(' ');
  return `<svg class="spark" viewBox="0 0 300 70" preserveAspectRatio="none" aria-hidden="true">
    <polygon points="0,70 ${p} 300,70" fill="var(--acs)"/>
    <polyline points="${p}" fill="none" stroke="var(--ac)" stroke-width="2" vector-effect="non-scaling-stroke"/>
  </svg>`;
}

function lines(arrays, max) {
  const m = max || 500;
  return `<svg class="spark" viewBox="0 0 300 70" preserveAspectRatio="none" aria-hidden="true">
    ${arrays.map((a, j) => {
      const p = a.map((v, i) => `${i * (300 / (a.length - 1 || 1))},${70 - (v / m) * 66}`).join(' ');
      return `<polyline points="${p}" fill="none" stroke="${j ? '#f0b429' : 'var(--ac)'}" stroke-width="2" vector-effect="non-scaling-stroke"/>`;
    }).join('')}
  </svg>`;
}

function formatChartTimestamp(ts) {
  try {
    return new Date(ts || Date.now()).toTimeString().slice(0, 8);
  } catch (e) {
    return clock();
  }
}

function chartSeriesSpec(kind) {
  const active = S.status === 'online' || S.status === 'starting';
  const z = arr => arr.map(() => 0);
  if (kind === 'cpu') {
    const data = active ? S.cpu : z(S.cpu);
    return { data, max: 100 };
  }
  if (kind === 'mem') {
    const data = active ? S.ram : z(S.ram);
    return { data, max: 100 };
  }
  const ni = active ? S.ni : z(S.ni);
  const no = active ? S.no : z(S.no);
  const max = Math.max(10, ...ni, ...no);
  return { ni, no, max };
}

function chartPointCoords(data, max) {
  const n = Math.max(1, data.length);
  const m = max || 100;
  return data.map((v, i) => {
    const x = n > 1 ? (i * 300) / (n - 1) : 150;
    const val = Math.max(0, Number(v) || 0);
    const y = 92 - Math.min(1, val / m) * 80;
    return [Number(x.toFixed(2)), Number(y.toFixed(2))];
  });
}

function buildChartSeriesInnerSvg(kind) {
  const spec = chartSeriesSpec(kind);
  if (kind === 'net') {
    const pIn = chartPointCoords(spec.ni, spec.max).map(([x, y]) => `${x},${y}`).join(' ');
    const pOut = chartPointCoords(spec.no, spec.max).map(([x, y]) => `${x},${y}`).join(' ');
    return `<polyline points="${pIn}" fill="none" stroke="var(--ac)" stroke-width="2" vector-effect="non-scaling-stroke"/>
      <polyline points="${pOut}" fill="none" stroke="#f0b429" stroke-width="2" vector-effect="non-scaling-stroke"/>`;
  }
  const coords = chartPointCoords(spec.data, spec.max);
  const pts = coords.map(([x, y]) => `${x},${y}`).join(' ');
  return `<polygon points="0,92 ${pts} 300,92" fill="var(--acs)"/>
    <polyline points="${pts}" fill="none" stroke="var(--ac)" stroke-width="2" vector-effect="non-scaling-stroke"/>`;
}

function updateChartHoverDom(kind) {
  const hoverSvg = document.getElementById(`ch-hover-svg-${kind}`);
  const tip = document.getElementById(`ch-tip-${kind}`);
  if (!hoverSvg || !tip) return;

  const activeHover = S.chartHoverKind === kind && Number.isInteger(S.chartHoverIdx) && S.chartHoverIdx >= 0;
  if (!activeHover) {
    hoverSvg.innerHTML = '';
    tip.style.display = 'none';
    return;
  }

  const len = Math.max(1, (S.cpu || []).length);
  const idx = Math.max(0, Math.min(len - 1, S.chartHoverIdx));
  const x = len > 1 ? (idx * 300) / (len - 1) : 150;
  const pctX = len > 1 ? (idx * 100) / (len - 1) : 50;
  const tsVal = Array.isArray(S.seriesTs) ? S.seriesTs[idx] : Date.now();
  const tsStr = formatChartTimestamp(tsVal);
  const active = S.status === 'online' || S.status === 'starting';

  if (kind === 'net') {
    const spec = chartSeriesSpec('net');
    const cIn = chartPointCoords(spec.ni, spec.max)[idx] || [x, 92];
    const cOut = chartPointCoords(spec.no, spec.max)[idx] || [x, 92];
    const vIn = active ? (spec.ni[idx] || 0) : 0;
    const vOut = active ? (spec.no[idx] || 0) : 0;
    hoverSvg.innerHTML = `
      <line class="ch-crosshair" x1="${x.toFixed(2)}" y1="6" x2="${x.toFixed(2)}" y2="94"/>
      <circle class="ch-dot-active" cx="${cIn[0]}" cy="${cIn[1]}" r="3.6" fill="var(--ac)" stroke="#fff" stroke-width="1.3"/>
      <circle class="ch-dot-active" cx="${cOut[0]}" cy="${cOut[1]}" r="3.6" fill="#f0b429" stroke="#fff" stroke-width="1.3"/>
    `;
    tip.innerHTML = `
      <div class="ch-tooltip-ts">${esc(tsStr)}</div>
      <div class="ch-tooltip-row"><span class="ch-dot-legend" style="background:var(--ac)"></span><span>In: ${vIn.toFixed(1)} KiB/s</span></div>
      <div class="ch-tooltip-row"><span class="ch-dot-legend" style="background:#f0b429"></span><span>Out: ${vOut.toFixed(1)} KiB/s</span></div>
    `;
  } else if (kind === 'mem') {
    const spec = chartSeriesSpec('mem');
    const pt = chartPointCoords(spec.data, spec.max)[idx] || [x, 92];
    const pct = active ? (spec.data[idx] || 0) : 0;
    const maxRamMB = S.memMaxMB || getServerMaxRamMB(curServer());
    const memMB = active
      ? ((Array.isArray(S.memSeriesMB) && S.memSeriesMB[idx] !== undefined) ? Number(S.memSeriesMB[idx]) : (pct / 100) * maxRamMB)
      : 0;
    hoverSvg.innerHTML = `
      <line class="ch-crosshair" x1="${x.toFixed(2)}" y1="6" x2="${x.toFixed(2)}" y2="94"/>
      <circle class="ch-dot-active" cx="${pt[0]}" cy="${pt[1]}" r="3.8" fill="var(--ac)" stroke="#fff" stroke-width="1.4"/>
    `;
    tip.innerHTML = `
      <div class="ch-tooltip-ts">${esc(tsStr)}</div>
      <div class="ch-tooltip-row"><span class="ch-dot-legend" style="background:var(--ac)"></span><span>${memMB.toFixed(0)} MiB (${pct.toFixed(1)}%)</span></div>
    `;
  } else {
    const spec = chartSeriesSpec('cpu');
    const pt = chartPointCoords(spec.data, spec.max)[idx] || [x, 92];
    const val = active ? (spec.data[idx] || 0) : 0;
    hoverSvg.innerHTML = `
      <line class="ch-crosshair" x1="${x.toFixed(2)}" y1="6" x2="${x.toFixed(2)}" y2="94"/>
      <circle class="ch-dot-active" cx="${pt[0]}" cy="${pt[1]}" r="3.8" fill="var(--ac)" stroke="#fff" stroke-width="1.4"/>
    `;
    tip.innerHTML = `
      <div class="ch-tooltip-ts">${esc(tsStr)}</div>
      <div class="ch-tooltip-row"><span class="ch-dot-legend" style="background:var(--ac)"></span><span>CPU: ${val.toFixed(1)}%</span></div>
    `;
  }

  tip.style.left = `${Math.max(18, Math.min(82, pctX))}%`;
  tip.style.display = 'block';
}

function onChartMouseMove(ev, kind) {
  const card = ev?.currentTarget?.closest?.('.card.ch') || document.querySelector(`.card.ch[data-chart-kind="${kind}"]`);
  const svg = card ? card.querySelector('.ch-svg') : null;
  const targetRect = svg ? svg.getBoundingClientRect() : ev?.currentTarget?.getBoundingClientRect?.();
  if (!targetRect || !targetRect.width) return;
  const relX = Math.max(0, Math.min(targetRect.width, (ev.clientX ?? 0) - targetRect.left));
  const len = Math.max(1, (S.cpu || []).length);
  const idx = Math.max(0, Math.min(len - 1, Math.round((relX / targetRect.width) * (len - 1))));
  S.chartHoverKind = kind;
  S.chartHoverIdx = idx;
  updateChartHoverDom(kind);
}

function onChartMouseLeave(kind) {
  if (S.chartHoverKind === kind) {
    S.chartHoverKind = null;
    S.chartHoverIdx = -1;
  }
  updateChartHoverDom(kind);
}

function head(name) {
  const raw = String(name || 'Steve').trim() || 'Steve';
  const encName = encodeURIComponent(raw);
  let h = 0;
  for (const c of raw) h = (h * 31 + c.charCodeAt(0)) % 360;
  const initial = raw.charAt(0).toUpperCase();
  const primaryUrl = `https://api.mcheads.org/head/${encName}/64`;
  const fallbackUrl = `https://mc-heads.net/avatar/${encName}/64`;
  return `<span class="head-wrap" title="${esc(raw)}"><img class="head" src="${primaryUrl}" alt="${esc(raw)}" loading="lazy" data-fallback="${fallbackUrl}" onerror="onPlayerHeadError(this, this.dataset.fallback)"><span class="head-fallback" style="background:hsl(${h} 45% 42%);">${esc(initial)}</span></span>`;
}

function loaderHtml(label = 'Loading...') {
  return `<div class="np-loader-box" role="status" aria-live="polite">
    <div class="np-spinner"></div>
    <span>${esc(label)}</span>
  </div>`;
}

function setBtnLoading(btn, isLoading, loadingText = 'Loading...') {
  if (!btn) return;
  if (isLoading) {
    if (!btn.dataset.origHtml) btn.dataset.origHtml = btn.innerHTML;
    btn.disabled = true;
    btn.classList.add('is-loading');
    btn.setAttribute('aria-busy', 'true');
    btn.innerHTML = `<span class="np-spinner-sm"></span> ${esc(loadingText)}`;
  } else {
    btn.disabled = false;
    btn.classList.remove('is-loading');
    btn.removeAttribute('aria-busy');
    if (btn.dataset.origHtml) {
      btn.innerHTML = btn.dataset.origHtml;
      delete btn.dataset.origHtml;
    }
  }
}

function av(s) {
  const customAvatar = (S.user?.avatar_url || S.acct?.avatar_url || '').trim();
  const letter = (S.acct?.fn ? S.acct.fn[0] : (S.acct?.un ? S.acct.un[0] : (S.user?.username ? S.user.username[0] : 'A'))).toUpperCase();
  if (customAvatar) {
    return `<span class="av" style="width:${s}px;height:${s}px;font-size:${s * 0.4}px">
      ${esc(letter)}
      <img src="${esc(customAvatar)}" alt="" style="object-fit:cover; width:100%; height:100%; border-radius:inherit;" onerror="this.remove()">
    </span>`;
  }
  const hasCustomGravatar = S.acct?.h && S.acct?.em && S.acct.em.toLowerCase() !== 'admin@neopanel.internal';
  const defAvatar = (typeof getBrand === 'function' ? getBrand('default_avatar_url') : '') || DEFAULT_USER_AVATAR_URL;
  return `<span class="av" style="width:${s}px;height:${s}px;font-size:${s * 0.4}px">
    ${esc(letter)}
    <img src="${esc(defAvatar)}" alt="" onerror="this.remove()">
    ${hasCustomGravatar ? `<img src="https://gravatar.com/avatar/${S.acct.h}?s=${s * 2}&d=404" alt="" onerror="this.remove()">` : ''}
  </span>`;
}

function userAvatarHtml(u, s = 32) {
  const name = u?.username || 'User';
  const letter = (name[0] || 'U').toUpperCase();
  const avatarUrl = (u?.avatar_url || '').trim();
  if (avatarUrl) {
    return `<span class="av user-avatar-ic" style="width:${s}px;height:${s}px;font-size:${s * 0.4}px">
      ${esc(letter)}
      <img src="${esc(avatarUrl)}" alt="${esc(name)}" style="object-fit:cover; width:100%; height:100%; border-radius:inherit;" onerror="this.remove()">
    </span>`;
  }
  return `<span class="av user-avatar-ic" style="width:${s}px;height:${s}px;font-size:${s * 0.4}px; background:linear-gradient(135deg, #6366f1, #8b5cf6); color:#fff; font-weight:700;">
    ${esc(letter)}
  </span>`;
}

// --- Branding, Theme Presets & Multi-Node Helpers ---
function detectFaviconMime(url) {
  const clean = String(url || '').split('?')[0].split('#')[0].toLowerCase();
  if (clean.startsWith('data:')) {
    const m = /^data:([^;,]+)/.exec(clean);
    return m ? m[1] : '';
  }
  if (clean.endsWith('.svg')) return 'image/svg+xml';
  if (clean.endsWith('.png')) return 'image/png';
  if (clean.endsWith('.ico')) return 'image/x-icon';
  if (clean.endsWith('.webp')) return 'image/webp';
  if (clean.endsWith('.gif')) return 'image/gif';
  if (clean.endsWith('.jpg') || clean.endsWith('.jpeg')) return 'image/jpeg';
  return '';
}

function getBrand(key) {
  const active = (S.view === 'admin' && (S.adminTab || 0) === 4 && S.brandDraft) ? S.brandDraft : (S.branding || DEFAULT_BRANDING);
  const val = active && active[key] !== undefined ? active[key] : DEFAULT_BRANDING[key];
  if ((key === 'server_hero_url' || key === 'card_cover_url' || key === 'default_avatar_url' || key === 'default_server_icon_url' || key === 'brand_name') && !String(val || '').trim()) {
    return DEFAULT_BRANDING[key];
  }
  return val;
}

function renderFooterHtml(raw) {
  let text = String(raw || getBrand('footer_text') || '[NeoPanel](https://github.com/Pekly/NeoPanel/)© 2026').trim();
  if (!text || text === 'Powered by NeoPanel Daemon') {
    text = '[NeoPanel](https://github.com/Pekly/NeoPanel/)© 2026';
  }
  const formatted = text.replace(/\[([^\]]+)\]\(([^)]+)\)/g, (match, label, url) => {
    let cleanUrl = String(url || '').trim().replace(/^https?:\/\/(https?:\/\/)/i, '$1');
    if (!cleanUrl.startsWith('http://') && !cleanUrl.startsWith('https://')) {
      cleanUrl = 'https://' + cleanUrl;
    }
    return `<a href="${esc(cleanUrl)}" target="_blank" rel="noopener noreferrer" style="color:inherit; text-decoration:none; font-weight:600; opacity:0.92;">${esc(label)}</a>`;
  });
  return formatted;
}

function hexToRgbStr(hex) {
  const m = /^#([0-9a-fA-F]{6})$/.exec(String(hex || '').trim());
  if (!m) return null;
  const n = parseInt(m[1], 16);
  return `${(n >> 16) & 255}, ${(n >> 8) & 255}, ${n & 255}`;
}

function isPngOrSvgUrl(url) {
  if (!url) return false;
  const lower = String(url).toLowerCase();
  return lower.endsWith('.png') || lower.includes('.png?') || lower.endsWith('.svg') || lower.includes('.svg?') || lower.startsWith('data:image/png') || lower.startsWith('data:image/svg');
}

function brandLogoHtml(forLogin = false, override = null) {
  const b = override || ((S.view === 'admin' && (S.adminTab || 0) === 4 && S.brandDraft) ? S.brandDraft : (S.branding || DEFAULT_BRANDING));
  const name = (b.brand_name || 'NeoPanel').trim() || 'NeoPanel';
  const logo = (b.logo_url || '').trim();
  const isPng = isPngOrSvgUrl(logo);
  const iconHtml = logo
    ? `<img src="${esc(logo)}" class="brand-logo-img ${isPng ? 'is-png' : ''}" alt="" onload="if(this.src.match(/\\.(png|svg)(\\?.*)?$/i)||this.src.startsWith('data:image\\/(png|svg)'))this.classList.add('is-png')" onerror="this.outerHTML='<i class=\\'cube\\'></i>'">`
    : `<i class="cube"></i>`;
  if (forLogin) {
    if (name === 'NeoPanel') {
      return `${iconHtml}<b>Neo</b><span>Panel</span>`;
    }
    return `${iconHtml}<b>${esc(name)}</b>`;
  }
  return `${iconHtml}<span style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">${esc(name)}</span>`;
}

function applyBranding(raw) {
  const b = { ...DEFAULT_BRANDING, ...(raw || {}) };
  const root = document.documentElement;
  const isLight = root.dataset.theme === 'light';
  const isOled = root.classList.contains('oled');

  const acRgb = hexToRgbStr(b.accent_color);
  if (b.accent_color && acRgb && (b.accent_color.toLowerCase() !== '#387aff' || b.preset !== 'obsidian')) {
    root.style.setProperty('--ac', b.accent_color);
    root.style.setProperty('--ac-hover', b.accent_hover || b.accent_color);
    root.style.setProperty('--ac-glow', `rgba(${acRgb}, 0.25)`);
    root.style.setProperty('--ac-subtle', `rgba(${acRgb}, 0.12)`);
    root.style.setProperty('--acs', `rgba(${acRgb}, 0.14)`);
    root.style.setProperty('--border-accent', `rgba(${acRgb}, 0.42)`);
  } else {
    ['--ac', '--ac-hover', '--ac-glow', '--ac-subtle', '--acs', '--border-accent'].forEach(p => root.style.removeProperty(p));
  }

  if (b.bg_deep && hexToRgbStr(b.bg_deep)) {
    root.style.setProperty('--bg-deep', b.bg_deep);
  } else {
    root.style.removeProperty('--bg-deep');
  }

  if (b.bg_canvas && hexToRgbStr(b.bg_canvas)) {
    root.style.setProperty('--bg-canvas', b.bg_canvas);
  } else {
    root.style.removeProperty('--bg-canvas');
  }

  const navRgb = hexToRgbStr(b.nav_bg);
  if (navRgb) {
    root.style.setProperty('--nav-bg', `rgba(${navRgb}, 0.88)`);
  } else {
    root.style.removeProperty('--nav-bg');
  }

  if (b.text_color && hexToRgbStr(b.text_color)) {
    root.style.setProperty('--tx', b.text_color);
  } else {
    root.style.removeProperty('--tx');
  }

  const cardRgb = hexToRgbStr(b.card_bg);
  const opacityNum = Number(b.card_opacity);
  const opacityVal = !isNaN(opacityNum) ? Math.max(0.35, Math.min(1, opacityNum / 100)) : 0.72;
  if (cardRgb) {
    root.style.setProperty('--card-bg', `rgba(${cardRgb}, ${opacityVal.toFixed(2)})`);
    root.style.setProperty('--card-bg-solid', b.card_bg);
  } else if (!isNaN(opacityNum) && Math.round(opacityNum) !== 72) {
    const fallbackRgb = isLight ? '255, 255, 255' : (isOled ? '8, 10, 14' : '18, 24, 38');
    root.style.setProperty('--card-bg', `rgba(${fallbackRgb}, ${opacityVal.toFixed(2)})`);
    root.style.removeProperty('--card-bg-solid');
  } else {
    root.style.removeProperty('--card-bg');
    root.style.removeProperty('--card-bg-solid');
  }

  const rad = Number(b.border_radius_px);
  if (!isNaN(rad) && rad >= 0 && rad <= 28) {
    root.style.setProperty('--np-radius', `${rad}px`);
  } else {
    root.style.removeProperty('--np-radius');
  }

  const blur = Number(b.glass_blur_px);
  if (!isNaN(blur) && blur >= 0 && blur <= 32) {
    root.style.setProperty('--np-blur', `${blur}px`);
  } else {
    root.style.removeProperty('--np-blur');
  }

  const brandName = (b.brand_name || 'NeoPanel').trim() || 'NeoPanel';
  const brandSub = (b.brand_subtitle || 'Minecraft Server Manager').trim();
  document.title = brandSub ? `${brandName} — ${brandSub}` : brandName;

  const favHref = (b.favicon_url || '').trim() || '/favicon.svg';
  const favType = detectFaviconMime(favHref);
  ['#app-favicon', '#app-favicon-shortcut'].forEach(sel => {
    const el = $(sel);
    if (el) {
      el.href = favHref;
      if (favType) el.type = favType;
      else el.removeAttribute('type');
    }
  });

  const cssEl = $('#custom-brand-css');
  if (cssEl) {
    cssEl.textContent = b.custom_css || '';
  }

  const navBrand = document.querySelector('#nav .brand');
  if (navBrand) {
    navBrand.innerHTML = brandLogoHtml(false, b);
  }

  const loginOverlay = $('#login');
  if (loginOverlay) {
    if ((b.login_bg_url || '').trim()) {
      loginOverlay.style.backgroundImage = `linear-gradient(rgba(6, 10, 18, 0.72), rgba(6, 10, 18, 0.85)), url('${b.login_bg_url.trim().replace(/'/g, '%27')}')`;
      loginOverlay.style.backgroundSize = 'cover';
      loginOverlay.style.backgroundPosition = 'center';
    } else {
      loginOverlay.style.backgroundImage = '';
    }
    const loginLogo = loginOverlay.querySelector('.llogo');
    if (loginLogo) loginLogo.innerHTML = brandLogoHtml(true, b);
  }
}

async function loadPublicBranding() {
  try {
    const res = await fetch('/api/branding');
    if (!res.ok) return;
    const data = await res.json().catch(() => null);
    if (data && data.branding) {
      S.branding = { ...DEFAULT_BRANDING, ...data.branding };
      S.brandingLoaded = true;
      if (data.daemon) {
        S.brandingDaemon = { ...S.brandingDaemon, ...data.daemon };
      }
      if (!localStorage.getItem('neopanel_mode') && S.branding.default_theme) {
        const dt = S.branding.default_theme.toLowerCase();
        const mapped = dt === 'light' ? 'Light' : dt === 'oled' ? 'Oled' : dt === 'auto' ? 'Auto' : 'Dark';
        S.acct.mode = mapped;
        const d = document.documentElement;
        d.classList.toggle('oled', mapped === 'Oled');
        if (mapped === 'Auto') {
          const prefersLight = window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches;
          d.dataset.theme = prefersLight ? 'light' : 'dark';
        } else {
          d.dataset.theme = mapped === 'Light' ? 'light' : 'dark';
        }
      }
      applyBranding(S.branding);
    }
  } catch (e) {}
}

function getNodeById(nodeId = 'local') {
  const id = nodeId || 'local';
  return (S.nodes || []).find(n => n.id === id) || null;
}

function getNode(nodeId = 'local') {
  return getNodeById(nodeId);
}

function findServerRef(srvId, nodeId = null) {
  if (!srvId) return null;
  const resolved = resolveServerByParam(srvId, nodeId);
  if (resolved) return resolved;
  const cur = curServer();
  if (cur && (cur.id === srvId || shortServerId(cur.id) === shortServerId(srvId))) return cur;
  return null;
}

function nodeApiPath(nodeId, apiPath) {
  const id = nodeId || 'local';
  const clean = apiPath.startsWith('/') ? apiPath : `/${apiPath}`;
  if (id === 'local') return clean;
  return `/api/nodes/${encodeURIComponent(id)}/proxy${clean}`;
}

function srvApiPath(srv, suffix = '') {
  if (!srv || !srv.id) return suffix;
  const nid = srv.node_id || 'local';
  const base = `/api/servers/${encodeURIComponent(srv.id)}${suffix || ''}`;
  if (nid === 'local') return base;
  return `/api/nodes/${encodeURIComponent(nid)}/proxy${base}`;
}

function srvHost(srv) {
  if (srv?.node_fqdn) return srv.node_fqdn;
  const node = getNodeById(srv?.node_id || 'local');
  if (node?.fqdn) return node.fqdn;
  return S.brandingDaemon?.fqdn || S.adminDaemonCfg?.fqdn || window.location.hostname;
}

function copyDbText(val, label = 'Value') {
  if (!val) return;
  const showSuccess = () => toast(`${label} copied to clipboard!`, 'ok');
  const fallback = () => {
    try {
      const ta = document.createElement('textarea');
      ta.value = val;
      ta.setAttribute('readonly', '');
      ta.style.position = 'fixed';
      ta.style.left = '-9999px';
      ta.style.opacity = '0';
      document.body.appendChild(ta);
      ta.select();
      const ok = document.execCommand('copy');
      document.body.removeChild(ta);
      if (ok) {
        showSuccess();
        return;
      }
    } catch (e) {}
    uiPrompt({ title: `Copy ${label}`, message: 'Copy the value below:', defaultValue: val, confirmText: 'Done' });
  };

  if (navigator.clipboard && window.isSecureContext) {
    navigator.clipboard.writeText(val).then(showSuccess).catch(fallback);
  } else {
    fallback();
  }
}

async function loadDatabases(silent = false, skipRender = false) {
  const srv = curServer();
  if (!srv) return;
  if (!S.loading) S.loading = {};
  if (!silent && !skipRender) {
    S.loading.databases = true;
    if (S.view === 'db') R();
  }
  try {
    const res = await api(srvApiPath(srv, '/databases'));
    S.databases = Array.isArray(res?.databases) ? res.databases : [];
    S.dbCapabilities = res?.capabilities || null;
  } catch (e) {
    S.databases = [];
  } finally {
    S.loading.databases = false;
    if (!skipRender && S.view === 'db') R();
  }
}

function toggleDbPasswordReveal(dbId) {
  if (!S.dbRevealIds) S.dbRevealIds = new Set();
  if (S.dbRevealIds.has(dbId)) S.dbRevealIds.delete(dbId);
  else S.dbRevealIds.add(dbId);
  if (S.view === 'db') R();
}

function openCreateDatabaseModal() {
  const srv = curServer();
  if (!srv) return;
  const container = $('#modal-container');
  if (!container) return;

  const caps = S.dbCapabilities || { sqlite: true, engines_label: 'Built-in SQLite (pure Go)' };
  const suggestedName = `mc_${(srv.name || 'server').toLowerCase().replace(/[^a-z0-9_]/g, '_').replace(/^[^a-z]+/, '').slice(0, 16) || 'main'}`;

  container.innerHTML = `<div class="pal" onclick="if(event.target===this)closeModal()">
    <div class="card pbox" style="padding:26px; max-width:500px;">
      <div class="row sb" style="margin-bottom:14px;">
        <h2 style="font-size:20px; margin:0; display:flex; align-items:center; gap:8px;">
          ${ic('database', 18)} Create Server Database
        </h2>
        <button class="btn sm" onclick="closeModal()" aria-label="Close modal">&times;</button>
      </div>
      <p class="mu" style="font-size:13px; line-height:1.5; margin-bottom:16px;">
        Provision an isolated database for <b>${esc(srv.name)}</b> using <b>${esc(caps.engines_label || 'Built-in SQLite')}</b>.
      </p>
      <form onsubmit="submitCreateDatabase(event)">
        <label class="f" style="margin-bottom:12px;">
          <span>Database Name <span class="mu">(letters, numbers, underscore)</span></span>
          <input type="text" id="db-new-name" required pattern="[A-Za-z][A-Za-z0-9_]{1,40}" value="${esc(suggestedName)}" placeholder="e.g. mc_coreprotect, luckperms_db">
        </label>
        <div class="grid g2" style="margin-bottom:16px;">
          <label class="f" style="margin-bottom:0;">
            <span>Engine</span>
            <select id="db-new-engine">
              <option value="sqlite" selected>SQLite 3 (Built-in pure Go)</option>
            </select>
          </label>
          <label class="f" style="margin-bottom:0;">
            <span>Max Connections</span>
            <input type="number" id="db-new-maxconn" min="1" max="500" value="25" required>
          </label>
        </div>
        <div class="row" style="justify-content:flex-end; gap:10px;">
          <button type="button" class="btn" onclick="closeModal()">Cancel</button>
          <button type="submit" id="db-create-btn" class="btn pr">${ic('plus', 14)} Create Database</button>
        </div>
      </form>
    </div>
  </div>`;

  enhanceSelects(container);
  setTimeout(() => $('#db-new-name')?.focus(), 40);
}

async function submitCreateDatabase(e) {
  e.preventDefault();
  const srv = curServer();
  if (!srv) return;
  const btn = $('#db-create-btn');
  if (btn) setBtnLoading(btn, true, 'Creating...');

  const name = $('#db-new-name')?.value.trim() || '';
  const engine = $('#db-new-engine')?.value || 'sqlite';
  const maxConnections = parseInt($('#db-new-maxconn')?.value, 10) || 25;

  try {
    const created = await api(srvApiPath(srv, '/databases'), {
      method: 'POST',
      body: JSON.stringify({ name, engine, max_connections: maxConnections })
    });
    closeModal();
    toast(`Database "${created?.name || name}" created!`, 'ok');
    await loadDatabases(true);
    if (S.view === 'db') R();
  } catch (err) {
    if (btn) setBtnLoading(btn, false);
  }
}

async function rotateDatabasePassword(dbId, name) {
  const srv = curServer();
  if (!srv) return;
  const ok = await uiConfirm({
    title: 'Rotate Database Password',
    message: `Generate a new credential password for database <b>${esc(name)}</b>?`,
    confirmText: 'Rotate Password',
    cancelText: 'Cancel',
    isDanger: false,
    icon: 'key'
  });
  if (!ok) return;

  try {
    await api(srvApiPath(srv, `/databases/${encodeURIComponent(dbId)}/rotate`), { method: 'POST' });
    if (!S.dbRevealIds) S.dbRevealIds = new Set();
    S.dbRevealIds.add(dbId);
    toast(`Password rotated for ${name}`, 'ok');
    await loadDatabases(true);
  } catch (e) {}
}

async function deleteServerDatabase(dbId, name) {
  const srv = curServer();
  if (!srv) return;
  const ok = await uiConfirm({
    title: 'Delete Database Permanently',
    message: `Delete database <b>${esc(name)}</b> and remove its SQLite file from <b>${esc(srv.name)}</b>? All tables and data will be permanently lost.`,
    confirmText: 'Delete Database',
    cancelText: 'Cancel',
    isDanger: true,
    icon: 'trash'
  });
  if (!ok) return;

  try {
    await api(srvApiPath(srv, `/databases/${encodeURIComponent(dbId)}`), { method: 'DELETE' });
    toast(`Database "${name}" deleted`, 'ok');
    await loadDatabases(true);
  } catch (e) {}
}

function exportServerDatabase(dbId, format = 'sql') {
  const srv = curServer();
  if (!srv) return;
  const baseUrl = srvApiPath(srv, `/databases/${encodeURIComponent(dbId)}/export?format=${encodeURIComponent(format)}`);
  const sep = baseUrl.includes('?') ? '&' : '?';
  window.open(`${baseUrl}${sep}token=${encodeURIComponent(S.token)}`, '_blank');
}

function renderSQLStudioSchema() {
  const el = $('#sql-studio-schema');
  const studio = S.sqlStudio;
  if (!el || !studio) return;

  if (studio.loadingTables) {
    el.innerHTML = loaderHtml('Inspecting tables...');
    return;
  }
  const tables = Array.isArray(studio.tables) ? studio.tables : [];
  if (tables.length === 0) {
    el.innerHTML = `<div class="mu" style="font-size:12px; line-height:1.5; padding:8px 4px;">
      No tables yet in <b>${esc(studio.db.name)}</b>.<br><br>
      Click <b>+ Sample Table</b> above to create a demo table, or run any <code>CREATE TABLE ...</code> SQL query on the right.
    </div>`;
    return;
  }

  el.innerHTML = tables.map(t => {
    const isSel = studio.selectedTable === t.name;
    const cols = Array.isArray(t.columns) ? t.columns : [];
    return `<div class="db-schema-item ${isSel ? 'on' : ''}" data-table="${esc(t.name)}" onclick="selectSQLStudioTable(this.dataset.table, true)">
      <div class="row sb" style="gap:6px; flex-wrap:nowrap;">
        <b class="mono" style="font-size:12.5px; color:var(--tx); overflow:hidden; text-overflow:ellipsis; white-space:nowrap;">${esc(t.name)}</b>
        <span class="chip" style="font-size:10.5px; padding:1px 6px; flex-shrink:0;">${Number(t.row_count) || 0} rows</span>
      </div>
      ${cols.length ? `<div class="db-schema-cols">
        ${cols.map(c => `<div class="db-col-row">
          <span style="color:var(--tx); overflow:hidden; text-overflow:ellipsis; white-space:nowrap;">${esc(c.name)}</span>
          <span class="row" style="gap:4px; flex-shrink:0;">
            <span class="mu" style="font-size:10.5px;">${esc(c.type || 'ANY')}</span>
            ${c.primary_key ? '<span class="db-col-badge">PK</span>' : ''}
            ${!c.primary_key && c.not_null ? '<span class="db-col-badge">NN</span>' : ''}
          </span>
        </div>`).join('')}
      </div>` : ''}
    </div>`;
  }).join('');
}

function renderSQLStudioResults() {
  const el = $('#sql-studio-results');
  const studio = S.sqlStudio;
  if (!el || !studio) return;

  if (studio.running) {
    el.innerHTML = loaderHtml('Executing SQL statement...');
    return;
  }
  if (studio.error) {
    el.innerHTML = `<div class="node-error-note" style="font-family:var(--font-mono);">${esc(studio.error)}</div>`;
    return;
  }
  if (!studio.result) {
    el.innerHTML = `<div class="mu" style="padding:18px; text-align:center; border:1px dashed var(--border); border-radius:10px; font-size:12.5px;">
      Press <b>Run Query</b> (or <kbd>Ctrl+Enter</kbd>) to inspect tables or modify data.
    </div>`;
    return;
  }

  const r = studio.result;
  if (r.mode === 'exec') {
    el.innerHTML = `<div class="card" style="padding:12px 14px; border-color:rgba(63,185,80,0.35); background:rgba(63,185,80,0.08);">
      <div class="row sb" style="flex-wrap:wrap; gap:8px;">
        <b style="color:var(--ok); font-size:13px;">Statement executed successfully</b>
        <span class="mu mono" style="font-size:11.5px;">${r.duration_ms ?? 0} ms</span>
      </div>
      <div class="mu mono" style="font-size:12px; margin-top:4px;">
        Rows affected: <b>${r.rows_affected ?? 0}</b>${r.last_insert_id ? ` &bull; Last Insert ID: <b>${r.last_insert_id}</b>` : ''}
      </div>
    </div>`;
    return;
  }

  const cols = Array.isArray(r.columns) ? r.columns : [];
  const rows = Array.isArray(r.rows) ? r.rows : [];
  el.innerHTML = `<div class="row sb" style="margin-bottom:8px; flex-wrap:wrap; gap:8px;">
      <span class="mu mono" style="font-size:12px;">
        Returned <b>${r.row_count ?? rows.length}</b> row${(r.row_count ?? rows.length) === 1 ? '' : 's'}
        ${r.truncated ? ' (capped at 500 rows)' : ''}
      </span>
      <span class="mu mono" style="font-size:11.5px;">${r.duration_ms ?? 0} ms</span>
    </div>
    ${cols.length === 0 ? `<div class="mu" style="padding:14px;">Query returned no columns.</div>` : `
      <div class="db-result-grid">
        <table>
          <thead><tr>${cols.map(c => `<th>${esc(c)}</th>`).join('')}</tr></thead>
          <tbody>
            ${rows.length ? rows.map(row => `<tr>${cols.map((_, idx) => {
              const val = Array.isArray(row) ? row[idx] : null;
              if (val === null || val === undefined) return `<td><span class="db-null">NULL</span></td>`;
              return `<td title="${esc(String(val))}">${esc(String(val))}</td>`;
            }).join('')}</tr>`).join('') : `<tr><td colspan="${cols.length}" class="mu" style="text-align:center; padding:18px;">0 rows returned</td></tr>`}
          </tbody>
        </table>
      </div>
    `}`;
}

async function loadSQLStudioTables() {
  const srv = curServer();
  const studio = S.sqlStudio;
  if (!srv || !studio?.db?.id) return;
  studio.loadingTables = true;
  renderSQLStudioSchema();
  try {
    const res = await api(srvApiPath(srv, `/databases/${encodeURIComponent(studio.db.id)}/tables`));
    studio.tables = Array.isArray(res?.tables) ? res.tables : [];
    const first = studio.tables[0]?.name || '';
    if (!studio.selectedTable && first) {
      studio.selectedTable = first;
      const inp = $('#db-sql-input');
      if (inp && !inp.value.trim()) {
        inp.value = `SELECT * FROM "${first.replace(/"/g, '""')}" LIMIT 100;`;
        studio.query = inp.value;
      }
    }
  } catch (e) {
    studio.tables = [];
  } finally {
    studio.loadingTables = false;
    renderSQLStudioSchema();
  }
}

function selectSQLStudioTable(tableName, autoRun = false) {
  if (!S.sqlStudio) return;
  S.sqlStudio.selectedTable = tableName;
  const escapedName = String(tableName || '').replace(/"/g, '""');
  const sql = `SELECT * FROM "${escapedName}" LIMIT 100;`;
  S.sqlStudio.query = sql;
  const inp = $('#db-sql-input');
  if (inp) inp.value = sql;
  renderSQLStudioSchema();
  if (autoRun) runSQLStudioQuery();
}

function insertSQLSnippet(kind) {
  const inp = $('#db-sql-input');
  if (!inp) return;
  const curTable = S.sqlStudio?.selectedTable || 'player_stats';
  const safeTable = curTable.replace(/"/g, '""');
  const snippets = {
    create: `CREATE TABLE IF NOT EXISTS player_stats (\n  id INTEGER PRIMARY KEY AUTOINCREMENT,\n  username TEXT NOT NULL UNIQUE,\n  balance INTEGER NOT NULL DEFAULT 0,\n  updated_at TEXT DEFAULT CURRENT_TIMESTAMP\n);`,
    insert: `INSERT INTO "${safeTable}" (username, balance) VALUES ('Steve', 250);`,
    select: `SELECT * FROM "${safeTable}" LIMIT 100;`,
    schema: `SELECT name, type, sql FROM sqlite_master WHERE name NOT LIKE 'sqlite_%' ORDER BY name;`
  };
  inp.value = snippets[kind] || snippets.select;
  if (S.sqlStudio) S.sqlStudio.query = inp.value;
  inp.focus();
}

async function runSQLStudioQuery() {
  const srv = curServer();
  const studio = S.sqlStudio;
  if (!srv || !studio?.db?.id) return;
  const inp = $('#db-sql-input');
  const sql = (inp ? inp.value : studio.query || '').trim();
  if (!sql) {
    toast('Enter a SQL query first', 'bad');
    return;
  }
  studio.query = sql;
  studio.running = true;
  studio.error = '';
  renderSQLStudioResults();

  const btn = $('#sql-run-btn');
  if (btn) setBtnLoading(btn, true, 'Running...');

  try {
    const res = await fetch(srvApiPath(srv, `/databases/${encodeURIComponent(studio.db.id)}/query`), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(S.token ? { 'Authorization': `Bearer ${S.token}` } : {})
      },
      body: JSON.stringify({ sql })
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      throw new Error(data.error || `HTTP ${res.status}`);
    }
    studio.result = data;
    studio.error = '';
    if (data.mode === 'exec') {
      await loadSQLStudioTables();
      loadDatabases(true);
    }
  } catch (err) {
    studio.result = null;
    studio.error = err.message || 'SQL execution error';
  } finally {
    studio.running = false;
    if (btn) setBtnLoading(btn, false);
    renderSQLStudioResults();
  }
}

async function openSQLStudioModal(dbId) {
  const srv = curServer();
  const db = (S.databases || []).find(d => d.id === dbId);
  const container = $('#modal-container');
  if (!srv || !db || !container) return;

  S.sqlStudio = {
    db,
    tables: [],
    selectedTable: '',
    loadingTables: true,
    running: false,
    query: `SELECT name, type, sql FROM sqlite_master WHERE name NOT LIKE 'sqlite_%' ORDER BY name;`,
    result: null,
    error: ''
  };

  container.innerHTML = `<div class="pal" onclick="if(event.target===this)closeModal()">
    <div class="pbox db-studio-modal">
      <div class="db-studio-head">
        <div class="row" style="gap:10px; align-items:center;">
          ${ic('database', 18)}
          <div>
            <b style="color:var(--tx); font-size:15.5px;">SQL Studio &mdash; ${esc(db.name)}</b>
            <div class="mu mono" style="font-size:11.5px;">${esc(db.jdbc_url || db.file_path || '')}</div>
          </div>
        </div>
        <div class="row" style="gap:8px;">
          <button type="button" class="btn sm" onclick="exportServerDatabase('${esc(db.id)}', 'sql')">${ic('download', 13)} Export .sql</button>
          <button type="button" class="btn sm" onclick="exportServerDatabase('${esc(db.id)}', 'db')">${ic('download', 13)} Export .db</button>
          <button type="button" class="btn sm" onclick="closeModal()" aria-label="Close SQL Studio">&times;</button>
        </div>
      </div>
      <div class="db-studio-body">
        <div class="db-schema-panel">
          <div class="row sb" style="margin-bottom:2px;">
            <b style="font-size:12px; text-transform:uppercase; letter-spacing:0.05em; color:var(--tx-muted);">Tables &amp; Schema</b>
            <button type="button" class="btn-ico" onclick="loadSQLStudioTables()" title="Refresh Schema">${ic('refresh', 13)}</button>
          </div>
          <div id="sql-studio-schema">${loaderHtml('Inspecting tables...')}</div>
        </div>
        <div class="db-query-panel">
          <div class="row sb" style="gap:8px; flex-wrap:wrap;">
            <div class="row" style="gap:6px; flex-wrap:wrap;">
              <button type="button" class="btn sm" onclick="insertSQLSnippet('create')">+ Sample Table</button>
              <button type="button" class="btn sm" onclick="insertSQLSnippet('insert')">+ Insert Row</button>
              <button type="button" class="btn sm" onclick="insertSQLSnippet('select')">Select 100</button>
              <button type="button" class="btn sm" onclick="insertSQLSnippet('schema')">Schema SQL</button>
            </div>
            <button type="button" id="sql-run-btn" class="btn sm pr" onclick="runSQLStudioQuery()">
              ${ic('play', 13)} Run Query (Ctrl+Enter)
            </button>
          </div>
          <textarea id="db-sql-input" class="db-sql-editor" spellcheck="false" placeholder="Write SQLite SQL query (SELECT, INSERT, UPDATE, CREATE TABLE...) and press Ctrl+Enter...">${esc(S.sqlStudio.query)}</textarea>
          <div id="sql-studio-results"></div>
        </div>
      </div>
    </div>
  </div>`;

  const inp = $('#db-sql-input');
  if (inp) {
    inp.addEventListener('keydown', e => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
        e.preventDefault();
        runSQLStudioQuery();
      }
    });
    setTimeout(() => inp.focus(), 40);
  }

  renderSQLStudioResults();
  await loadSQLStudioTables();
  if (S.sqlStudio.tables.length > 0 && S.sqlStudio.selectedTable) {
    selectSQLStudioTable(S.sqlStudio.selectedTable, true);
  }
}

async function computeGravatarHash(email) {
  try {
    const msg = new TextEncoder().encode(email.trim().toLowerCase());
    const hashBuffer = await crypto.subtle.digest('SHA-256', msg);
    return Array.from(new Uint8Array(hashBuffer)).map(b => b.toString(16).padStart(2, '0')).join('');
  } catch (e) {
    return '';
  }
}

// --- API Client ---
async function api(endpoint, options = {}) {
  assertClientRequestAccess(endpoint, options);
  const { suppressToast, ...fetchOpts } = options || {};
  const headers = { ...(fetchOpts.headers || {}) };
  if (S.token) headers['Authorization'] = `Bearer ${S.token}`;
  if (!headers['Content-Type'] && !(fetchOpts.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
  }

  try {
    const res = await fetch(endpoint, { ...fetchOpts, headers });
    if (res.status === 401) {
      logout();
      const error = new Error('Session expired. Please sign in again.');
      error.status = 401;
      throw error;
    }
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      const error = new Error(data.error || `HTTP ${res.status}`);
      error.status = res.status;
      throw error;
    }
    return data;
  } catch (err) {
    if (!suppressToast && err.message && !err.message.includes('Session expired')) {
      toast(err.message, 'error');
    }
    throw err;
  }
}

function curServer() {
  if (S.cs === -1) return null;
  if (!S.servers || S.servers.length === 0) return null;
  return S.servers[S.cs] || S.servers[0];
}

// --- Advanced Minecraft & Terminal Log Parser & Palette System ---
const _pluginColorPalette = [
  { fg: '#38bdf8', bg: 'rgba(56, 189, 248, 0.12)', border: 'rgba(56, 189, 248, 0.28)' }, // Sky
  { fg: '#34d399', bg: 'rgba(52, 211, 153, 0.12)', border: 'rgba(52, 211, 153, 0.28)' }, // Emerald
  { fg: '#a78bfa', bg: 'rgba(167, 139, 250, 0.12)', border: 'rgba(167, 139, 250, 0.28)' }, // Purple
  { fg: '#f472b6', bg: 'rgba(244, 114, 182, 0.12)', border: 'rgba(244, 114, 182, 0.28)' }, // Pink
  { fg: '#fbbf24', bg: 'rgba(251, 191, 36, 0.12)', border: 'rgba(251, 191, 36, 0.28)' },   // Amber
  { fg: '#2dd4bf', bg: 'rgba(45, 212, 191, 0.12)', border: 'rgba(45, 212, 191, 0.28)' },  // Teal
  { fg: '#818cf8', bg: 'rgba(129, 140, 248, 0.12)', border: 'rgba(129, 140, 248, 0.28)' }, // Indigo
  { fg: '#fb923c', bg: 'rgba(251, 146, 60, 0.12)', border: 'rgba(251, 146, 60, 0.28)' },   // Orange
  { fg: '#e879f9', bg: 'rgba(232, 121, 249, 0.12)', border: 'rgba(232, 121, 249, 0.28)' }  // Fuchsia
];

function getPluginColor(name) {
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }
  return _pluginColorPalette[Math.abs(hash) % _pluginColorPalette.length];
}

function cleanConsoleOutput(raw, native = false) {
  // Keep SGR colors; this line-based viewer cannot render terminal cursor/erase/OSC commands.
  const clean = String(raw).replace(/\x1b(?:\[[0-?]*[ -/]*[@-~]|\][^\x07\x1b]*(?:\x07|\x1b\\)|[ -/]*[@-Z\\-_])/g,
    code => /^\x1b\[[0-9;]*m$/.test(code) ? code : '');
  if (!native) return clean;
  const parts = clean.split('\r');
  for (let i = parts.length - 1; i >= 0; i--) {
    const line = parts[i].trim();
    const visible = line.replace(/\x1b\[[0-9;]*m/g, '').trim();
    if (visible && visible !== '>') return line;
  }
  return '';
}

function parseAnsi(text) {
  const ansi16 = {
    30: '#4b5563', 31: '#ef4444', 32: '#10b981', 33: '#f59e0b',
    34: '#3b82f6', 35: '#ec4899', 36: '#06b6d4', 37: '#f3f4f6',
    90: '#6b7280', 91: '#f87171', 92: '#34d399', 93: '#fbbf24',
    94: '#60a5fa', 95: '#f472b6', 96: '#22d3ee', 97: '#ffffff'
  };

  const ansiBg16 = {
    40: '#1f2937', 41: '#7f1d1d', 42: '#064e3b', 43: '#78350f',
    44: '#1e3a8a', 45: '#831843', 46: '#164e63', 47: '#374151',
    100: '#374151', 101: '#991b1b', 102: '#065f46', 103: '#92400e',
    104: '#1e40af', 105: '#9d174d', 106: '#155e75', 107: '#4b5563'
  };

  function get256Color(n) {
    if (n < 16) return ansi16[30 + (n % 8)] || '#ffffff';
    if (n >= 16 && n <= 231) {
      n -= 16;
      const r = Math.floor(n / 36);
      const g = Math.floor((n % 36) / 6);
      const b = n % 6;
      const toVal = (v) => v ? 55 + v * 40 : 0;
      return `rgb(${toVal(r)}, ${toVal(g)}, ${toVal(b)})`;
    }
    if (n >= 232 && n <= 255) {
      const g = Math.round((n - 232) * 10.9 + 8);
      return `rgb(${g}, ${g}, ${g})`;
    }
    return '#ffffff';
  }

  return text.replace(/\x1b\[([0-9;]*)m/g, (_, codeStr) => {
    if (!codeStr || codeStr === '0') return '</span>';
    const tokens = codeStr.split(';').map(t => parseInt(t, 10) || 0);
    const styles = [];
    for (let i = 0; i < tokens.length; i++) {
      const c = tokens[i];
      if (c === 0) {
        styles.length = 0;
      } else if (c === 1) {
        styles.push('font-weight:bold');
      } else if (c === 2) {
        styles.push('opacity:0.65');
      } else if (c === 3) {
        styles.push('font-style:italic');
      } else if (c === 4) {
        styles.push('text-decoration:underline');
      } else if (c === 9) {
        styles.push('text-decoration:line-through');
      } else if (c === 38) {
        if (tokens[i + 1] === 2 && tokens.length > i + 4) {
          styles.push(`color:rgb(${tokens[i+2]},${tokens[i+3]},${tokens[i+4]})`);
          i += 4;
        } else if (tokens[i + 1] === 5 && tokens.length > i + 2) {
          styles.push(`color:${get256Color(tokens[i+2])}`);
          i += 2;
        }
      } else if (c === 48) {
        if (tokens[i + 1] === 2 && tokens.length > i + 4) {
          styles.push(`background-color:rgb(${tokens[i+2]},${tokens[i+3]},${tokens[i+4]})`);
          i += 4;
        } else if (tokens[i + 1] === 5 && tokens.length > i + 2) {
          styles.push(`background-color:${get256Color(tokens[i+2])}`);
          i += 2;
        }
      } else if (ansi16[c]) {
        styles.push(`color:${ansi16[c]}`);
      } else if (ansiBg16[c]) {
        styles.push(`background-color:${ansiBg16[c]}`);
      }
    }
    return styles.length ? `</span><span style="${styles.join(';')}">` : '</span>';
  });
}

function formatCommandLine(cmd) {
  const ts = new Date().toTimeString().split(' ')[0];
  return `<div class="con-cmd-row"><span class="con-ts">[${ts}]</span><span class="con-cmd-badge">CMD</span><span class="con-cmd-prefix">&gt;</span><span class="con-cmd-text">${esc(cmd)}</span></div>`;
}

// --- Minecraft Legacy Color, Hex, Gradient & MiniMessage Renderer ---
const MC_LEGACY_COLORS = {
  '0': '#000000',
  '1': '#0000aa',
  '2': '#00aa00',
  '3': '#00aaaa',
  '4': '#aa0000',
  '5': '#aa00aa',
  '6': '#ffaa00',
  '7': '#aaaaaa',
  '8': '#555555',
  '9': '#5555ff',
  'a': '#55ff55',
  'b': '#55ffff',
  'c': '#ff5555',
  'd': '#ff55ff',
  'e': '#ffff55',
  'f': '#ffffff'
};

const MC_NAMED_COLORS = {
  black: '#000000',
  dark_blue: '#0000aa',
  dark_green: '#00aa00',
  dark_aqua: '#00aaaa',
  dark_red: '#aa0000',
  dark_purple: '#aa00aa',
  gold: '#ffaa00',
  gray: '#aaaaaa',
  grey: '#aaaaaa',
  dark_gray: '#555555',
  dark_grey: '#555555',
  blue: '#5555ff',
  green: '#55ff55',
  aqua: '#55ffff',
  red: '#ff5555',
  light_purple: '#ff55ff',
  yellow: '#ffff55',
  white: '#ffffff'
};

function parseMcColorHex(raw) {
  const s = String(raw || '').trim().toLowerCase();
  if (!s) return null;
  if (MC_NAMED_COLORS[s]) return MC_NAMED_COLORS[s];
  const m6 = /^#([0-9a-f]{6})$/i.exec(s);
  if (m6) return `#${m6[1].toLowerCase()}`;
  const m3 = /^#([0-9a-f]{3})$/i.exec(s);
  if (m3) {
    const [r, g, b] = m3[1].toLowerCase().split('');
    return `#${r}${r}${g}${g}${b}${b}`;
  }
  return null;
}

function hexToRgbTuple(hex) {
  const h = String(hex || '#ffffff').replace('#', '');
  return [
    parseInt(h.slice(0, 2), 16) || 0,
    parseInt(h.slice(2, 4), 16) || 0,
    parseInt(h.slice(4, 6), 16) || 0
  ];
}

function lerpGradientHex(stopsRgb, t) {
  if (!stopsRgb || stopsRgb.length === 0) return '#ffffff';
  if (stopsRgb.length === 1) {
    const [r, g, b] = stopsRgb[0];
    return `#${((1 << 24) + (r << 16) + (g << 8) + b).toString(16).slice(1)}`;
  }
  const clamped = Math.max(0, Math.min(1, t));
  const scaled = clamped * (stopsRgb.length - 1);
  const idx = Math.min(stopsRgb.length - 2, Math.floor(scaled));
  const localT = scaled - idx;
  const a = stopsRgb[idx];
  const b = stopsRgb[idx + 1];
  const r = Math.round(a[0] + (b[0] - a[0]) * localT);
  const g = Math.round(a[1] + (b[1] - a[1]) * localT);
  const bl = Math.round(a[2] + (b[2] - a[2]) * localT);
  return `#${((1 << 24) + (r << 16) + (g << 8) + bl).toString(16).slice(1)}`;
}

function hsvRainbowHex(t) {
  const h = (((t % 1) + 1) % 1) * 6;
  const i = Math.floor(h);
  const f = h - i;
  const q = 1 - f;
  const [r, g, b] = [
    [1, f, 0],
    [q, 1, 0],
    [0, 1, f],
    [0, q, 1],
    [f, 0, 1],
    [1, 0, q]
  ][i % 6];
  const toHex = x => Math.round(x * 255).toString(16).padStart(2, '0');
  return `#${toHex(r)}${toHex(g)}${toHex(b)}`;
}

function splitMcGraphemes(str) {
  if (!str) return [];
  if (typeof Intl !== 'undefined' && typeof Intl.Segmenter === 'function') {
    const seg = new Intl.Segmenter(undefined, { granularity: 'grapheme' });
    return Array.from(seg.segment(str), s => s.segment);
  }
  return Array.from(str);
}

function renderMinecraftTextHtml(raw, opts = {}) {
  const allowNewlines = opts.allowNewlines !== false;
  let src = String(raw == null ? '' : raw);
  if (!src) return '';

  // Normalize Java server.properties escapes (\u00A7, \n, \:, \=, \\)
  src = src.replace(/\\(u[0-9a-fA-F]{4}|n|r|t|\\|:|=)/g, (full, escSeq) => {
    if (escSeq === 'n') return '\n';
    if (escSeq === 'r') return '';
    if (escSeq === 't') return ' ';
    if (escSeq === '\\') return '\\';
    if (escSeq === ':' || escSeq === '=') return escSeq;
    if (escSeq.startsWith('u')) {
      const code = parseInt(escSeq.slice(1), 16);
      return Number.isFinite(code) ? String.fromCharCode(code) : full;
    }
    return full;
  });

  if (!allowNewlines) {
    src = src.replace(/\r?\n/g, ' ');
  } else {
    src = src.replace(/\r\n/g, '\n').replace(/\r/g, '\n');
  }

  // Tokenize legacy, hex, Iridium {#RRGGBB>}, and MiniMessage tags safely
  const tokens = [];
  const pushText = (t) => {
    if (!t) return;
    if (tokens.length > 0 && tokens[tokens.length - 1].type === 'text') {
      tokens[tokens.length - 1].text += t;
    } else {
      tokens.push({ type: 'text', text: t });
    }
  };

  let pos = 0;
  const n = src.length;
  while (pos < n) {
    const ch = src[pos];
    if (ch === '\n') {
      tokens.push({ type: 'newline' });
      pos += 1;
      continue;
    }

    if (ch === '§' || ch === '&') {
      const slice = src.slice(pos);
      const bxMatch = /^(?:§|&)x(?:(?:§|&)([0-9a-fA-F])){6}/i.exec(slice);
      if (bxMatch) {
        const hex = slice.slice(0, 14).replace(/[§&x]/gi, '');
        if (/^[0-9a-fA-F]{6}$/.test(hex)) {
          tokens.push({ type: 'legacy_color', color: `#${hex.toLowerCase()}` });
          pos += 14;
          continue;
        }
      }
      const hexMatch = /^(?:§|&)#([0-9a-fA-F]{6})/.exec(slice);
      if (hexMatch) {
        tokens.push({ type: 'legacy_color', color: `#${hexMatch[1].toLowerCase()}` });
        pos += hexMatch[0].length;
        continue;
      }
      const codeMatch = /^(?:§|&)([0-9a-fklmnor])/i.exec(slice);
      if (codeMatch) {
        const c = codeMatch[1].toLowerCase();
        if (MC_LEGACY_COLORS[c]) {
          tokens.push({ type: 'legacy_color', color: MC_LEGACY_COLORS[c] });
        } else {
          tokens.push({ type: 'legacy_fmt', code: c });
        }
        pos += 2;
        continue;
      }
      pushText(ch);
      pos += 1;
      continue;
    }

    if (ch === '{') {
      const irMatch = /^\{\#([0-9a-fA-F]{6})([><]?)\}/.exec(src.slice(pos));
      if (irMatch) {
        const hex = `#${irMatch[1].toLowerCase()}`;
        if (irMatch[2] === '>') {
          tokens.push({ type: 'iridium_open', hex });
        } else if (irMatch[2] === '<') {
          tokens.push({ type: 'iridium_close', hex });
        } else {
          tokens.push({ type: 'legacy_color', color: hex });
        }
        pos += irMatch[0].length;
        continue;
      }
      pushText(ch);
      pos += 1;
      continue;
    }

    if (ch === '<') {
      const mmMatch = /^<([!\/]?)([a-zA-Z0-9_#]+)(?::([^>]*))?>/.exec(src.slice(pos));
      if (mmMatch) {
        const prefix = mmMatch[1] || '';
        const tagLower = (mmMatch[2] || '').toLowerCase();
        const argStr = mmMatch[3] != null ? mmMatch[3] : '';
        let consumed = true;

        if (!prefix && (tagLower === 'br' || tagLower === 'newline') && !argStr) {
          tokens.push({ type: allowNewlines ? 'newline' : 'text', ...(allowNewlines ? {} : { text: ' ' }) });
        } else if ((tagLower === 'reset' || tagLower === 'r') && !prefix && !argStr) {
          tokens.push({ type: 'mm_reset' });
        } else if (tagLower === 'gradient') {
          if (prefix === '/' && !argStr) {
            tokens.push({ type: 'mm_close_grad', kind: 'gradient' });
          } else if (!prefix && argStr) {
            const parts = argStr.split(':').map(p => p.trim()).filter(Boolean);
            let phase = 0;
            if (parts.length > 2) {
              const lastNum = Number(parts[parts.length - 1]);
              if (Number.isFinite(lastNum) && /^[-+]?\d*\.?\d+$/.test(parts[parts.length - 1])) {
                phase = lastNum;
                parts.pop();
              }
            }
            const stops = parts.map(parseMcColorHex).filter(Boolean);
            if (stops.length >= 2) {
              tokens.push({
                type: 'mm_gradient',
                stops: stops.map(hexToRgbTuple),
                phase
              });
            } else if (stops.length === 1) {
              tokens.push({ type: 'mm_color', color: stops[0] });
            } else {
              consumed = false;
            }
          } else {
            consumed = false;
          }
        } else if (tagLower === 'rainbow') {
          if (prefix === '/' && !argStr) {
            tokens.push({ type: 'mm_close_grad', kind: 'rainbow' });
          } else if (!prefix) {
            let phase = 0;
            let inverted = false;
            if (argStr) {
              let rawPhase = argStr.trim();
              if (rawPhase.startsWith('!')) {
                inverted = true;
                rawPhase = rawPhase.slice(1);
              }
              if (rawPhase) {
                const p = Number(rawPhase);
                if (Number.isFinite(p)) phase = p;
                else consumed = false;
              }
            }
            if (consumed) {
              tokens.push({ type: 'mm_rainbow', phase, inverted });
            }
          } else {
            consumed = false;
          }
        } else if (tagLower === 'color' || tagLower === 'c' || tagLower === 'colour') {
          if (prefix === '/' && !argStr) {
            tokens.push({ type: 'mm_close_color' });
          } else if (!prefix && argStr) {
            const cHex = parseMcColorHex(argStr);
            if (cHex) tokens.push({ type: 'mm_color', color: cHex });
            else consumed = false;
          } else {
            consumed = false;
          }
        } else if (tagLower.startsWith('#') && !argStr) {
          const cHex = parseMcColorHex(tagLower);
          if (cHex) {
            if (prefix === '/') tokens.push({ type: 'mm_close_color' });
            else if (!prefix) tokens.push({ type: 'mm_color', color: cHex });
            else consumed = false;
          } else {
            consumed = false;
          }
        } else if (MC_NAMED_COLORS[tagLower] && !argStr) {
          if (prefix === '/' || prefix === '!') {
            tokens.push({ type: 'mm_close_color' });
          } else {
            tokens.push({ type: 'mm_color', color: MC_NAMED_COLORS[tagLower] });
          }
        } else {
          const fmtMap = {
            bold: 'bold', b: 'bold',
            italic: 'italic', i: 'italic', em: 'italic',
            underlined: 'underline', underline: 'underline', u: 'underline',
            strikethrough: 'strike', st: 'strike',
            obfuscated: 'obf', obf: 'obf'
          };
          const prop = fmtMap[tagLower];
          if (prop && !argStr) {
            tokens.push({ type: 'mm_fmt', prop, val: !(prefix === '/' || prefix === '!') });
          } else {
            consumed = false;
          }
        }

        if (consumed) {
          pos += mmMatch[0].length;
          continue;
        }
      }
      pushText(ch);
      pos += 1;
      continue;
    }

    // Read contiguous plain characters quickly until next control starter
    let next = pos + 1;
    while (next < n) {
      const c = src[next];
      if (c === '\n' || c === '§' || c === '&' || c === '{' || c === '<') break;
      next++;
    }
    pushText(src.slice(pos, next));
    pos = next;
  }

  // Resolve Iridium {#c1>}...{#c2<} spans into gradient tokens when closed properly
  for (let i = 0; i < tokens.length; i++) {
    if (tokens[i].type === 'iridium_open') {
      let closeIdx = -1;
      for (let j = i + 1; j < tokens.length; j++) {
        if (tokens[j].type === 'iridium_close') {
          closeIdx = j;
          break;
        }
      }
      if (closeIdx !== -1) {
        tokens[i] = {
          type: 'mm_gradient',
          stops: [hexToRgbTuple(tokens[i].hex), hexToRgbTuple(tokens[closeIdx].hex)],
          phase: 0
        };
        tokens[closeIdx] = { type: 'mm_close_grad', kind: 'gradient' };
      } else {
        tokens[i] = { type: 'legacy_color', color: tokens[i].hex };
      }
    } else if (tokens[i].type === 'iridium_close') {
      tokens[i] = { type: 'legacy_color', color: tokens[i].hex };
    }
  }

  // Measure visible grapheme span length for gradient / rainbow tags
  function measureGradientSpan(startIdx) {
    let total = 0;
    for (let j = startIdx + 1; j < tokens.length; j++) {
      const tk = tokens[j];
      if (tk.type === 'mm_close_grad' || tk.type === 'mm_reset' || tk.type === 'mm_gradient' || tk.type === 'mm_rainbow') {
        break;
      }
      if (tk.type === 'legacy_fmt' && tk.code === 'r') break;
      if (tk.type === 'text') total += splitMcGraphemes(tk.text).length;
    }
    return total;
  }

  const state = {
    color: null,
    colorStack: [],
    gradient: null,
    bold: false,
    italic: false,
    underline: false,
    strike: false,
    obf: false
  };

  function resetAll() {
    state.color = null;
    state.colorStack = [];
    state.gradient = null;
    state.bold = false;
    state.italic = false;
    state.underline = false;
    state.strike = false;
    state.obf = false;
  }

  function buildStyleAttr(hexColor) {
    const styles = [];
    if (hexColor) styles.push(`color:${hexColor}`);
    if (state.bold) styles.push('font-weight:800');
    if (state.italic) styles.push('font-style:italic');
    if (state.underline && state.strike) styles.push('text-decoration:underline line-through');
    else if (state.underline) styles.push('text-decoration:underline');
    else if (state.strike) styles.push('text-decoration:line-through');
    return styles.join(';');
  }

  let html = '';
  for (let i = 0; i < tokens.length; i++) {
    const tk = tokens[i];
    if (tk.type === 'newline') {
      html += allowNewlines ? '<br>' : ' ';
      continue;
    }
    if (tk.type === 'legacy_color') {
      // Vanilla MC legacy color codes reset active formatting flags
      state.gradient = null;
      state.color = tk.color;
      state.colorStack = [];
      state.bold = false;
      state.italic = false;
      state.underline = false;
      state.strike = false;
      state.obf = false;
      continue;
    }
    if (tk.type === 'legacy_fmt') {
      if (tk.code === 'r') resetAll();
      else if (tk.code === 'l') state.bold = true;
      else if (tk.code === 'o') state.italic = true;
      else if (tk.code === 'n') state.underline = true;
      else if (tk.code === 'm') state.strike = true;
      else if (tk.code === 'k') state.obf = true;
      continue;
    }
    if (tk.type === 'mm_reset') {
      resetAll();
      continue;
    }
    if (tk.type === 'mm_fmt') {
      state[tk.prop] = tk.val;
      continue;
    }
    if (tk.type === 'mm_color') {
      if (state.color) state.colorStack.push(state.color);
      state.color = tk.color;
      continue;
    }
    if (tk.type === 'mm_close_color') {
      state.color = state.colorStack.length ? state.colorStack.pop() : null;
      continue;
    }
    if (tk.type === 'mm_gradient') {
      const len = Math.max(1, measureGradientSpan(i));
      state.color = null;
      state.colorStack = [];
      state.gradient = {
        mode: 'gradient',
        stops: tk.stops,
        phase: tk.phase || 0,
        index: 0,
        length: len
      };
      continue;
    }
    if (tk.type === 'mm_rainbow') {
      const len = Math.max(1, measureGradientSpan(i));
      state.color = null;
      state.colorStack = [];
      state.gradient = {
        mode: 'rainbow',
        phase: tk.phase || 0,
        inverted: Boolean(tk.inverted),
        index: 0,
        length: len
      };
      continue;
    }
    if (tk.type === 'mm_close_grad') {
      state.gradient = null;
      continue;
    }
    if (tk.type === 'text') {
      // If active gradient & not temporarily overridden by a solid MM color, color each grapheme
      if (state.gradient && !state.color) {
        const g = state.gradient;
        const graphemes = splitMcGraphemes(tk.text);
        for (const gr of graphemes) {
          const baseT = g.length <= 1 ? 0 : g.index / (g.length - 1);
          g.index++;
          let hex;
          if (g.mode === 'rainbow') {
            const hueT = (g.inverted ? (1 - baseT) : baseT) + (g.phase / Math.max(1, g.length));
            hex = hsvRainbowHex(hueT);
          } else {
            let t = baseT + g.phase;
            if (g.phase !== 0) {
              // Triangle fold keeps phase smooth across stops
              t = Math.abs((((t % 2) + 2) % 2) - 1);
            }
            hex = lerpGradientHex(g.stops, t);
          }
          const stAttr = buildStyleAttr(hex);
          const clsAttr = state.obf ? ' class="mc-obf"' : '';
          html += `<span${clsAttr} style="${stAttr}">${esc(gr)}</span>`;
        }
      } else {
        const stAttr = buildStyleAttr(state.color);
        const clsAttr = state.obf ? ' class="mc-obf"' : '';
        if (stAttr || clsAttr) {
          html += `<span${clsAttr}${stAttr ? ` style="${stAttr}"` : ''}>${esc(tk.text)}</span>`;
        } else {
          html += esc(tk.text);
        }
      }
    }
  }

  return `<span class="mc-formatted-text">${html}</span>`;
}

function formatLogLine(raw) {
  if (!raw) return '';

  // 1. Unwrap JSON payload even if preceded by timestamp like [03:06:34] or whitespace
  if (typeof raw === 'string') {
    const jsonIdx = raw.indexOf('{"event"');
    if (jsonIdx !== -1) {
      try {
        const obj = JSON.parse(raw.substring(jsonIdx));
        if (obj.event === 'log' && obj.data != null) {
          raw = obj.data;
        }
      } catch (e) {}
    } else if (raw.trim().startsWith('{')) {
      try {
        const obj = JSON.parse(raw.trim());
        if (obj.event === 'log' && obj.data != null) {
          raw = obj.data;
        }
      } catch (e) {}
    }
  }

  raw = cleanConsoleOutput(raw, isSteelEngine(curServer()?.type));
  if (!raw) return '';

  // 2. If already formatted HTML command row, preserve directly
  if (typeof raw === 'string') {
    if (raw.startsWith('<div class="con-cmd-row"')) {
      return raw;
    }
    if (raw.startsWith('<span class="c">&gt;') || raw.startsWith('&gt; ') || raw.startsWith('> ')) {
      const cleanCmd = raw.replace(/<[^>]+>/g, '').replace(/^[>&;gt\s]+/, '').trim();
      return formatCommandLine(cleanCmd);
    }
  }

  // 3. Automatically detect if server is ready
  if (raw.includes('Done (') || raw.includes('For help, type "help"') || raw.includes('Server Ping Player Sample Count')) {
    if (S.status !== 'online') {
      S.status = 'online';
      updateInputState();
      updateHeaderBadge();
    }
  } else if (raw.includes('Stopping server') || raw.includes('Closing Server')) {
    S.status = 'stopping';
    updateInputState();
    updateHeaderBadge();
  }

  // 4. Escape HTML entities
  let text = esc(raw);

  // 5. Minecraft Section & Ampersand Color Codes and RGB Hex Colors
  const mcColors = {
    '0': '#000000', '1': '#0000aa', '2': '#00aa00', '3': '#00aaaa',
    '4': '#aa0000', '5': '#aa00aa', '6': '#ffaa00', '7': '#aaaaaa',
    '8': '#555555', '9': '#5555ff', 'a': '#55ff55', 'b': '#55ffff',
    'c': '#ff5555', 'd': '#ff55ff', 'e': '#ffff55', 'f': '#ffffff'
  };

  // Modern Hex formats: &#RRGGBB, §#RRGGBB, <#RRGGBB>, &x&r&r&g&g&b&b
  text = text.replace(/(?:§|&amp;|&)#([0-9a-fA-F]{6})/gi, (_, hex) => `</span><span style="color:#${hex}">`);
  text = text.replace(/&lt;#([0-9a-fA-F]{6})&gt;/gi, (_, hex) => `</span><span style="color:#${hex}">`);
  text = text.replace(/(?:&amp;x|§x)(?:(?:&amp;|§)([0-9a-fA-F])){6}/gi, (match) => {
    const chars = match.replace(/&amp;|§|x/gi, '');
    return chars.length === 6 ? `</span><span style="color:#${chars}">` : '';
  });

  // Standard § and & formatting codes
  text = text.replace(/(?:§|&amp;)([0-9a-f])/gi, (_, c) => `</span><span style="color:${mcColors[c.toLowerCase()] || '#fff'}">`);
  text = text.replace(/(?:§|&amp;)l/gi, '<b>').replace(/(?:§|&amp;)r/gi, '</b></span>');
  text = text.replace(/(?:§|&amp;)o/gi, '<i>').replace(/(?:§|&amp;)n/gi, '<ins>').replace(/(?:§|&amp;)m/gi, '<del>');

  // 6. ANSI Color Escape Sequences (Full 24-bit TrueColor RGB, 256-color, 16-color)
  text = parseAnsi(text);

  // 7. Highlight Log Levels & Timestamps
  text = text.replace(/\[([0-9:.]+)\s*(INFO|WARN|WARNING|ERROR|SEVERE|FATAL|DEBUG|TRACE)\]:/g, (_, ts, lvl) => {
    const upper = lvl.toUpperCase();
    let lvlClass = 'info';
    if (upper.includes('WARN')) lvlClass = 'warn';
    else if (upper.includes('ERR') || upper.includes('SEV') || upper.includes('FATAL')) lvlClass = 'error';
    return `<span class="con-ts">[${ts}]</span> <span class="con-lvl con-lvl-${lvlClass}">${upper}</span><span class="con-colon">:</span>`;
  });

  // 8. Highlight Plugin Bracket Badges like [SKSpawn], [SKCore], [Vault], [NeoPanel]
  text = text.replace(/\[([A-Za-z0-9_-]{2,30})\]/g, (match, pName) => {
    const reserved = ['INFO', 'WARN', 'WARNING', 'ERROR', 'SEVERE', 'FATAL', 'DEBUG', 'TRACE', 'SERVER', 'CLIENT'];
    if (reserved.includes(pName.toUpperCase())) return match;
    if (pName === 'NeoPanel') {
      return `<span class="con-plug-badge" style="color:#10b981; background:rgba(16,185,129,0.12); border:1px solid rgba(16,185,129,0.3); font-weight:700;">[NeoPanel]</span>`;
    }
    const c = getPluginColor(pName);
    return `<span class="con-plug-badge" style="color:${c.fg}; background:${c.bg}; border:1px solid ${c.border}; font-weight:600;">[${pName}]</span>`;
  });

  // 9. Brand / Server prefixes with chevrons: SMP » or Network » or Server ->
  text = text.replace(/\b([A-Za-z0-9_-]{2,20})\s*(»|&raquo;|->|›)\s*/g, (_, brand, arrow) => {
    return `<span class="con-brand-prefix">${brand}</span> <span class="con-brand-arrow">${arrow}</span> `;
  });

  // 10. Highlight property keys like protection.collision: or Economy hooked:
  text = text.replace(/(\b[a-zA-Z0-9_.-]{3,30}:)\s+/g, (match, prop) => {
    if (prop.startsWith('http:') || prop.startsWith('https:') || prop === 'INFO:' || prop === 'WARN:' || prop === 'ERROR:') {
      return match;
    }
    return `<span class="con-prop-key">${prop}</span> `;
  });

  // 11. Parenthetical notes: (through TAB). or (/help)
  text = text.replace(/\(([a-zA-Z0-9_\-/\s.]+)\)/g, '<span class="dim" style="color:#94a3b8;">($1)</span>');

  // 12. Full line error / warning fallbacks if unhandled
  if (text.includes('Exception:') || text.includes('at java.') || text.includes('at net.minecraft.')) {
    text = `<span class="e">${text}</span>`;
  } else if (text.includes('joined the game') || text.includes('has made the advancement')) {
    text = `<span class="c">${text}</span>`;
  } else if (text.includes('left the game')) {
    text = `<span class="w">${text}</span>`;
  }

  return text;
}

function updateInputState() {
  const isOnline = S.status === 'online' && hasServerPermission('console.command');
  const placeholder = isOnline ? 'Type a command, try help...' : 'Server is offline';
  for (const id of ['#cmd', '#fly-cmd']) {
    const inp = $(id);
    if (inp) {
      inp.disabled = !isOnline;
      inp.placeholder = placeholder;
    }
  }
  const sendBtn = $('#fly-send-btn');
  if (sendBtn) sendBtn.disabled = !isOnline;
}

function updateHeaderBadge() {
  const label = { online: 'Running', offline: 'Offline', starting: 'Starting', stopping: 'Stopping' }[S.status] || 'Offline';
  const badgeHtml = `<i class="dot"></i>${label}`;
  for (const id of ['#srv-badge', '#fly-con-badge']) {
    const badge = $(id);
    if (badge) {
      badge.className = `badge ${S.status}`;
      badge.innerHTML = badgeHtml;
    }
  }
  const haltBadge = $('#srv-halt-badge');
  if (haltBadge) {
    haltBadge.style.display = (S.status === 'offline' && S.startupHalted) ? 'inline-flex' : 'none';
  }
  const startBtn = document.querySelector('[data-power-btn="start"]');
  if (startBtn) startBtn.disabled = S.status !== 'offline' || !hasServerPermission('server.power');
  const restartBtn = document.querySelector('[data-power-btn="restart"]');
  if (restartBtn) restartBtn.disabled = S.status !== 'online' || !hasServerPermission('server.power');
  const stopBtn = document.querySelector('[data-power-btn="stop"]');
  if (stopBtn) stopBtn.disabled = !(S.status === 'online' || S.status === 'starting') || !hasServerPermission('server.power');
  const fabDot = $('#fly-con-fab .fly-con-dot');
  if (fabDot) fabDot.className = `fly-con-dot ${S.status}`;
}

function applyWsPillState(pill) {
  if (!pill) return;
  if (S.wsState === 'connected') {
    pill.className = 'con-live-pill live';
    pill.innerHTML = '<i></i>Live';
  } else if (S.wsState === 'reconnecting' || S.wsState === 'connecting') {
    pill.className = 'con-live-pill warn';
    pill.innerHTML = '<i></i>Reconnecting...';
  } else {
    pill.className = 'con-live-pill';
    pill.innerHTML = '<i></i>Offline';
  }
}

function updateWsLivePill() {
  applyWsPillState($('#ws-live-pill'));
  applyWsPillState($('#fly-ws-pill'));
}

function updateConBufferCount() {
  const txt = `${S.log.length} lines`;
  const el = $('#con-buf-count');
  if (el) el.textContent = txt;
  const flyEl = $('#fly-con-buf-count');
  if (flyEl) flyEl.textContent = txt;
  const fabCnt = $('#fly-fab-count');
  if (fabCnt) {
    fabCnt.textContent = S.log.length > 0 ? String(S.log.length) : '';
    fabCnt.style.display = S.log.length > 0 ? 'inline-block' : 'none';
  }
}

function appendLogToEl(el, html) {
  if (!el) return;
  const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 90;
  const row = document.createElement('div');
  row.className = 'con-line';
  row.innerHTML = html;
  el.appendChild(row);
  while (el.children.length > 800) {
    el.removeChild(el.firstChild);
  }
  if (atBottom) el.scrollTop = el.scrollHeight;
}

function addLog(line) {
  if (!line) return;
  const html = formatLogLine(line);
  if (!html) return;

  // Prevent consecutive duplicate lines (e.g. concurrent streams or duplicate frames)
  if (S.log.length > 0 && S.log[S.log.length - 1] === html) {
    return;
  }

  S.log.push(html);
  if (S.log.length > 800) S.log.shift();

  if (S.view === 'con') {
    appendLogToEl($('#log'), html);
  }
  appendLogToEl($('#fly-log'), html);
  updateConBufferCount();
}

function paintTargetLog(l) {
  if (!l) return;
  l.innerHTML = S.log.length
    ? S.log.map(line => `<div class="con-line">${line}</div>`).join('')
    : `<div class="con-line dim">No console output buffered yet. Start the server or type a command below.</div>`;
  l.scrollTop = l.scrollHeight;
}

function paintFlyConsole() {
  paintTargetLog($('#fly-log'));
  updateConBufferCount();
}

function paintLog() {
  paintTargetLog($('#log'));
  paintTargetLog($('#fly-log'));
  updateConBufferCount();
}

function syncFlyConsoleState() {
  updateHeaderBadge();
  updateWsLivePill();
  updateConBufferCount();
  updateInputState();
  const fab = $('#fly-con-fab');
  if (fab) fab.classList.toggle('is-open', Boolean(S.flyConOpen));
}

function bindFlyConsoleInput() {
  const inp = $('#fly-cmd');
  if (!inp || inp.dataset.bound === '1') return;
  inp.dataset.bound = '1';
  inp.addEventListener('keydown', e => {
    if (e.key === 'Enter' && inp.value.trim()) {
      e.preventDefault();
      runFlyCmd();
    } else if (e.key === 'ArrowUp' && S.hi > 0) {
      e.preventDefault();
      inp.value = S.hist[--S.hi] || '';
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      inp.value = S.hist[++S.hi] || '';
    }
  });
}

function runFlyCmd() {
  const inp = $('#fly-cmd');
  if (!inp || !inp.value.trim()) return;
  const c = inp.value.trim();
  runCmd(c);
  S.hist.push(c);
  S.hi = S.hist.length;
  inp.value = '';
}

function toggleFlyConsole(forceOpen) {
  if (!hasServerPermission('console.view')) return;
  const srv = curServer();
  if (!srv) {
    toast('Select or create a server first');
    return;
  }
  if (S.view === 'con') {
    const cmdInp = $('#cmd');
    if (cmdInp && !cmdInp.disabled) cmdInp.focus();
    else toast('Already viewing full Console');
    return;
  }
  const next = typeof forceOpen === 'boolean' ? forceOpen : !S.flyConOpen;
  S.flyConOpen = next;
  if (next && (!S.activeWs || S.activeWs.readyState !== WebSocket.OPEN)) {
    connectWs(srv.id, true, srv.node_id || 'local');
  }
  renderFlyConsole();
}

function toggleFlyConsoleSize() {
  S.flyConTall = !S.flyConTall;
  renderFlyConsole();
}

function closeFlyConsole() {
  if (!S.flyConOpen) return;
  S.flyConOpen = false;
  renderFlyConsole();
}

function renderFlyConsole() {
  const root = $('#fly-console-root');
  if (!root) return;
  const srv = curServer();
  const isLoginOpen = Boolean($('#login'));
  const showTrigger = !['dash', 'admin', 'acc', 'client', 'notfound', 'mod', 'maintenance'].includes(S.view);
  if (!S.token || isLoginOpen || !srv || S.view === 'con' || (!showTrigger && !S.flyConOpen)) {
    root.innerHTML = '';
    delete root.dataset.flySig;
    return;
  }

  const nodeId = srv.node_id || 'local';
  const sig = `${nodeId}|${srv.id}|${S.flyConOpen ? 1 : 0}|${S.flyConTall ? 1 : 0}`;
  if (root.dataset.flySig === sig && root.children.length > 0) {
    syncFlyConsoleState();
    return;
  }

  const prevOpen = root.dataset.flySig?.split('|')[1] === '1';
  root.dataset.flySig = sig;
  const statusLabel = { online: 'Running', offline: 'Offline', starting: 'Starting', stopping: 'Stopping' }[S.status] || 'Offline';

  root.innerHTML = `
    ${S.flyConOpen ? `
      <div class="fly-con-drawer ${S.flyConTall ? 'tall' : ''}" role="dialog" aria-label="Live Console Drawer">
        <div class="fly-con-head">
          <div class="fly-con-title">
            ${ic('terminal', 15)}
            <b>${esc(srv.name || 'Server Console')}</b>
            <span id="fly-con-badge" class="badge ${S.status}"><i class="dot"></i>${statusLabel}</span>
            <span id="fly-ws-pill" class="con-live-pill"><i></i>Connecting...</span>
            <span id="fly-con-buf-count" class="mu mono" style="font-size: 11px;">${S.log.length} lines</span>
          </div>
          <div class="row" style="gap: 6px; flex-wrap: nowrap;">
            <button type="button" class="btn sm" onclick="S.log=[]; paintLog(); toast('Console buffer cleared');">Clear</button>
            <button type="button" class="btn sm" onclick="toggleFlyConsoleSize()" title="${S.flyConTall ? 'Compact drawer size' : 'Expand drawer size'}">${S.flyConTall ? 'Compact' : 'Expand'}</button>
            <button type="button" class="btn sm" onclick="S.flyConOpen=false; go('con');" title="Open full Console view">Full Page</button>
            <button type="button" class="btn sm" onclick="closeFlyConsole()" aria-label="Minimize live console">&times;</button>
          </div>
        </div>
        <div class="con fly-con-logs" id="fly-log" role="log" aria-live="polite"></div>
        <div class="cin fly-con-input">
          <span>&raquo;</span>
          <input id="fly-cmd" type="text" class="con-cmd-input" autocomplete="off" spellcheck="false" placeholder="Type a command..." aria-label="Quick console command">
          <button type="button" id="fly-send-btn" class="btn sm pr fly-send-btn" onclick="runFlyCmd()">Send</button>
        </div>
      </div>
    ` : ''}
    <button type="button" id="fly-con-fab" class="fly-con-fab ${S.flyConOpen ? 'is-open' : ''}" onclick="toggleFlyConsole()" title="Toggle Live Console (Ctrl+\`)">
      <span class="fly-con-dot ${S.status}"></span>
      ${ic('terminal', 15)}
      <span>${S.flyConOpen ? 'Hide Console' : 'Live Console'}</span>
      <span id="fly-fab-count" class="fly-con-count" style="display:${S.log.length > 0 ? 'inline-block' : 'none'}">${S.log.length > 0 ? S.log.length : ''}</span>
      <kbd>Ctrl+\`</kbd>
    </button>
  `;

  if (S.flyConOpen) {
    paintFlyConsole();
    bindFlyConsoleInput();
    if (!prevOpen) {
      setTimeout(() => {
        const inp = $('#fly-cmd');
        if (inp && !inp.disabled) inp.focus();
      }, 30);
    }
  }
  syncFlyConsoleState();
}

// --- WebSocket Connection (Auto-Reconnect + Cloudflare Keepalive Heartbeat) ---
function connectWs(serverId, isReconnect = false, nodeId = null) {
  if (!serverId || !S.token) return;
  if (!hasServerPermission('console.view', { id: serverId, node_id: nodeId || curServer()?.node_id || 'local' })) return;

  const targetNodeId = nodeId || curServer()?.node_id || 'local';

  clearTimeout(S.wsReconnectTimer);
  S.wsReconnectTimer = null;

  // If already connected or connecting to this server on this node, do not recreate WebSocket
  if (
    S.activeWs &&
    S.activeServerId === serverId &&
    (S.activeNodeId || 'local') === targetNodeId &&
    (S.activeWs.readyState === WebSocket.OPEN || S.activeWs.readyState === WebSocket.CONNECTING)
  ) {
    return;
  }

  if (S.wsPingTimer) {
    clearInterval(S.wsPingTimer);
    S.wsPingTimer = null;
  }

  if (S.activeWs) {
    S.activeWs.onopen = null;
    S.activeWs.onmessage = null;
    S.activeWs.onerror = null;
    S.activeWs.onclose = null;
    try { S.activeWs.close(); } catch (e) {}
    S.activeWs = null;
  }

  S.activeServerId = serverId;
  S.activeNodeId = targetNodeId;
  S.wsState = isReconnect ? 'reconnecting' : 'connecting';
  updateWsLivePill();

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsPath = targetNodeId === 'local'
    ? `/api/servers/${encodeURIComponent(serverId)}/ws`
    : `/api/nodes/${encodeURIComponent(targetNodeId)}/ws/${encodeURIComponent(serverId)}`;
  const wsUrl = `${protocol}//${window.location.host}${wsPath}?token=${encodeURIComponent(S.token)}`;

  try {
    const ws = new WebSocket(wsUrl);
    S.activeWs = ws;

    ws.onopen = () => {
      if (S.activeWs !== ws) return;
      S.wsState = 'connected';
      S.wsAttempts = 0;
      updateWsLivePill();

      clearInterval(S.wsPingTimer);
      S.wsPingTimer = setInterval(() => {
        if (ws.readyState === WebSocket.OPEN) {
          try {
            ws.send(JSON.stringify({ event: 'ping' }));
          } catch (e) {}
        }
      }, 25000);
    };

    ws.onmessage = e => {
      if (!e.data) return;
      try {
        const text = e.data.trim();
        if (text.startsWith('{')) {
          const msg = JSON.parse(text);
          if (msg.event === 'pong') {
            return;
          } else if (msg.event === 'history') {
            try {
              const list = JSON.parse(msg.data);
              if (Array.isArray(list)) {
                S.log = list.map(formatLogLine).filter(Boolean);
                if (S.log.length > 800) S.log = S.log.slice(-800);
                paintLog();
                return;
              }
            } catch (histErr) {}
          } else if (msg.event === 'log') {
            addLog(msg.data);
            return;
          } else if (msg.event === 'stats' || msg.type === 'stats') {
            const prevStatus = S.status;
            const prevHalted = S.startupHalted;
            applyServerStatsSnapshot(curServer(), msg, true);
            if (S.status !== prevStatus || S.startupHalted !== prevHalted) {
              updateInputState();
              updateHeaderBadge();
            }
            liveCon();
            if (S.view === 'dash') updateDashLiveStats();
            return;
          } else if (msg.event === 'status' || msg.type === 'status') {
            S.status = normalizeServerStatus(msg.data || S.status);
            updateInputState();
            updateHeaderBadge();
            return;
          }
        }
      } catch (err) {}

      addLog(e.data);
    };

    ws.onerror = () => {
      try { ws.close(); } catch (e) {}
    };

    ws.onclose = () => {
      if (S.wsPingTimer) {
        clearInterval(S.wsPingTimer);
        S.wsPingTimer = null;
      }
      if (S.activeWs === ws) {
        S.activeWs = null;
      }
      if (S.token && S.activeServerId === serverId && (S.activeNodeId || 'local') === targetNodeId) {
        S.wsState = 'reconnecting';
        updateWsLivePill();
        const delay = Math.min(1000 * Math.pow(1.5, S.wsAttempts), 8000);
        S.wsAttempts++;
        S.wsReconnectTimer = setTimeout(() => {
          connectWs(serverId, true, targetNodeId);
        }, delay);
      }
    };
  } catch (e) {
    S.wsState = 'reconnecting';
    updateWsLivePill();
  }
}

document.addEventListener('visibilitychange', () => {
  if (!document.hidden && S.token && S.activeServerId) {
    if (!S.activeWs || S.activeWs.readyState === WebSocket.CLOSED) {
      connectWs(S.activeServerId, true, S.activeNodeId || 'local');
    }
  }
});

window.addEventListener('online', () => {
  if (S.token && S.activeServerId && (!S.activeWs || S.activeWs.readyState === WebSocket.CLOSED)) {
    connectWs(S.activeServerId, true, S.activeNodeId || 'local');
  }
});

// Local 1-second uptime ticker fallback + 5-second live players poll when viewing Online Players
let _panelTickCount = 0;
setInterval(() => {
  _panelTickCount = (_panelTickCount + 1) % 60;
  if ((S.status === 'online' || S.status === 'starting') && (!S.activeWs || S.activeWs.readyState !== WebSocket.OPEN)) {
    S.uptimeSec = (Number(S.uptimeSec) || 0) + 1;
    if (S.view === 'con') liveCon();
    else if (S.view === 'dash') updateDashLiveStats();
  }
  if (_panelTickCount % 5 === 0 && S.view === 'pl' && S.ptab === 'online' && S.token && !$('#login')) {
    loadPlayersData(true).then(() => {
      if (S.view === 'pl' && S.ptab === 'online') R();
    }).catch(() => {});
  }
}, 1000);

// --- Power Actions ---
async function power(action, skipConfirm = false) {
  if (!hasServerPermission('server.power')) { toast('Access restricted: server.power required', 'bad'); return; }
  const srv = curServer();
  if (!srv) return;

  if (!skipConfirm && action !== 'start') {
    const confirmMeta = {
      restart: {
        title: 'Restart Server',
        message: `Are you sure you want to restart <b>${esc(srv.name)}</b>? Connected players will be temporarily disconnected while the server reboots.`,
        confirmText: 'Restart Server',
        isDanger: false,
        icon: 'refresh'
      },
      stop: {
        title: 'Stop Server',
        message: `Are you sure you want to stop <b>${esc(srv.name)}</b>? The server will save world data and shut down cleanly.`,
        confirmText: 'Stop Server',
        isDanger: true,
        icon: 'alert'
      },
      kill: {
        title: 'Force stop server?',
        message: `Are you sure you want to <b>force kill</b> <b>${esc(srv.name)}</b>? This immediately terminates the Java process without saving chunks and may cause world data loss.`,
        confirmText: 'Force stop',
        isDanger: true,
        icon: 'alert'
      }
    }[action];

    if (confirmMeta) {
      const ok = await uiConfirm({
        title: confirmMeta.title,
        message: confirmMeta.message,
        confirmText: confirmMeta.confirmText,
        cancelText: 'Cancel',
        isDanger: confirmMeta.isDanger,
        icon: confirmMeta.icon
      });
      if (!ok) return;
    }
  }

  if (action === 'start') {
    S.status = 'starting';
    S.startupHalted = false;
    S.startupFailures = 0;
    addLog(`[NeoPanel]: Starting instance ${srv.name}...`);
    updateInputState();
    updateHeaderBadge();
    toast('Starting server...');
  } else if (action === 'stop') {
    S.status = 'stopping';
    addLog('[NeoPanel]: Stopping server process...');
    updateInputState();
    updateHeaderBadge();
    toast('Stopping server...');
  } else if (action === 'restart') {
    S.startupHalted = false;
    S.startupFailures = 0;
    toast('Restarting server...');
    addLog('[NeoPanel]: Restart sequence initiated...');
  } else if (action === 'kill') {
    toast('Terminating server process...');
    addLog('[NeoPanel]: Force kill signal delivered');
  }

  try {
    const res = await api(srvApiPath(srv, '/power'), {
      method: 'POST',
      body: JSON.stringify({ action })
    });
    if (res && res.current) {
      applyServerStatsSnapshot(srv, res.current, false);
    } else {
      const fallback = (action === 'start' || action === 'restart') ? 'starting' : 'offline';
      S.status = normalizeServerStatus(fallback);
    }
    updateInputState();
    updateHeaderBadge();
    liveCon();
    toast(`Server ${action} command sent`, 'ok');
  } catch (err) {
    S.status = 'offline';
    updateInputState();
    updateHeaderBadge();
    addLog(`\x1b[31;1m[NeoPanel Error]: ${err.message}\x1b[0m`);
    toast(`Action failed: ${err.message}`, 'bad');

    let modalTitle = 'Server Action Failed';
    let modalMsg = `Failed to ${action} server: <b>${esc(err.message)}</b>`;

    const lower = (err.message || '').toLowerCase();
    if (lower.includes('jar file') || lower.includes('jar') || lower.includes('not found')) {
      modalTitle = 'Server Executable Not Found';
      modalMsg = `<div style="line-height:1.6;">
        The server executable <b>${esc(serverExecutable(srv))}</b> could not be found in the server directory.<br><br>
        <div style="background:rgba(239,68,68,0.1); border:1px solid rgba(239,68,68,0.3); border-radius:8px; padding:10px 12px; margin-bottom:12px; font-size:13px; color:#fca5a5;">
          ${esc(err.message)}
        </div>
        <b>How to resolve:</b>
        <ol style="margin: 8px 0 0 18px; padding: 0; line-height: 1.6;">
          <li>Click <b>Reinstall Server</b> in Settings to download an official ${isSteelEngine(srv.type) ? 'Steel release' : 'build from mcjarfiles.com'}.</li>
          <li>Or navigate to <b>Files</b> and upload the ${isSteelEngine(srv.type) ? 'native executable' : 'JAR'} as <code>${esc(serverExecutable(srv))}</code>.</li>
          <li>Or verify your <b>Server Executable File</b> name in <b>Startup Settings</b>.</li>
        </ol>
      </div>`;
    } else if (!isSteelEngine(srv.type) && lower.includes('java')) {
      modalTitle = 'Java Runtime Error';
      modalMsg = `<div style="line-height:1.6;">
        The server could not launch with the selected Java runtime.<br><br>
        <div style="background:rgba(239,68,68,0.1); border:1px solid rgba(239,68,68,0.3); border-radius:8px; padding:10px 12px; margin-bottom:12px; font-size:13px; color:#fca5a5;">
          ${esc(err.message)}
        </div>
        <b>How to resolve:</b>
        <ol style="margin: 8px 0 0 18px; padding: 0; line-height: 1.6;">
          <li>Go to <b>Settings -> Startup</b> and choose an installed Java runtime (<b>Java 25</b>, <b>Java 21</b>, <b>Java 17</b>, or <b>Java 8</b>), or click <b>Install</b> next to a missing runtime.</li>
          <li>Or run <code>sudo neopanel-installer --java</code> on your server to install all supported Java versions.</li>
        </ol>
      </div>`;
    }

    uiAlert({
      title: modalTitle,
      message: modalMsg,
      type: 'error'
    });
    refreshServerStatus();
  }
}

async function refreshServerStatus() {
  const srv = curServer();
  if (!srv) return;
  try {
    const data = await api(srvApiPath(srv, ''));
    applyServerStatsSnapshot(srv, data.stats, false);
    S.ver = data.version || S.ver;
    S.core = (data.type || 'paper').charAt(0).toUpperCase() + (data.type || 'paper').slice(1);
    updateInputState();
    updateHeaderBadge();
    R();
  } catch (e) {}
}

async function runCmd(cmd) {
  if (!hasServerPermission('console.command')) { toast('Access restricted: console.command required', 'bad'); return; }
  if (!cmd) return;
  addLog(formatCommandLine(cmd));
  const srv = curServer();
  if (!srv) return;

  if (S.activeWs && S.activeWs.readyState === WebSocket.OPEN) {
    S.activeWs.send(JSON.stringify({ event: 'command', data: cmd }));
  } else {
    try {
      await api(srvApiPath(srv, '/command'), {
        method: 'POST',
        body: JSON.stringify({ command: cmd })
      });
    } catch (e) {}
  }
}

// --- Categorized Navigation (PG) with Vector SVGs ---
const PG = [
  ['con', 'Console', 'terminal', 'General'],
  ['set', 'Settings', 'settings', 'General'],
  ['act', 'Activity', 'activity', 'General'],
  ['fi', 'Files', 'files', 'Management'],
  ['db', 'Databases', 'database', 'Management'],
  ['bk', 'Backups', 'backup', 'Management'],
  ['sc', 'Schedules', 'clock', 'Management'],
  ['team', 'Team', 'users', 'Management'],
  ['net', 'Network', 'network', 'Management'],
  ['sftp', 'SFTP', 'files', 'Management'],
  ['sub', 'Subdomain', 'link', 'Management'],
  ['pl', 'Players', 'users', 'Management'],
  ['plug', 'Plugins', 'puzzle', 'Minecraft'],
  ['ver', 'Software', 'cpu', 'Minecraft']
];

const CMDS = [
  ['Save All Worlds', 'save-all', 'download'],
  ['List Online Players', 'list', 'users'],
  ['Check TPS & Health', 'tps', 'activity'],
  ['Reload Configurations', 'reload confirm', 'refresh']
];

const CORES = [
  ['Paper', 'High performance, plugin friendly'],
  ['Purpur', 'Paper fork with extreme configurability'],
  ['Leaf', 'Hyper-optimized Paper fork for large SMPs'],
  ['Folia', 'Multi-threaded regionised Paper fork'],
  ['Fabric', 'Lightweight modular mod loader'],
  ['Forge', 'Classic modding platform (legacy & modern)'],
  ['NeoForge', 'Modern community forge modding engine'],
  ['Vanilla', 'Standard Mojang vanilla server'],
  ['Velocity', 'Modern high-throughput proxy'],
  ['Waterfall', 'PaperMC BungeeCord proxy fork'],
  ['BungeeCord', 'Classic multi-server proxy'],
  ['Steel', 'Experimental Minecraft Java Edition server written in Rust']
];

const FAM = [
  ['1.21', 'Tricky Trials & Game Drops', '2024–2026', [
    '1.21.11', '1.21.10', '1.21.9', '1.21.8', '1.21.7', '1.21.6', '1.21.5',
    '1.21.4', '1.21.3', '1.21.2', '1.21.1', '1.21'
  ]],
  ['1.20', 'Trails & Tales', '2023–2024', ['1.20.6', '1.20.4', '1.20.1']],
  ['1.19', 'The Wild Update', '2022–2023', ['1.19.4', '1.19.2']],
  ['1.18', 'Caves & Cliffs: Part II', '2021–2022', ['1.18.2']],
  ['1.16', 'Nether Update', '2020–2021', ['1.16.5']],
  ['1.12', 'World of Color & Forge Era', '2017', ['1.12.2']],
  ['1.8', 'Bountiful Update & PvP', '2014–2015', ['1.8.9']],
  ['1.7', 'Legacy Modding Classic', '2013–2014', ['1.7.10']]
];

const POPULAR_MC_VERSIONS = new Set([
  '1.20.6', '1.20.4', '1.20.1',
  '1.19.4', '1.19.2',
  '1.18.2',
  '1.16.5',
  '1.12.2',
  '1.8.9',
  '1.7.10'
]);

const ALL_VERS = FAM.flatMap(f => f[3].map(v => ({ v, f: f[0], n: f[1], y: f[2] })));
const FALLBACK_ALL_MC = ALL_VERS.map(item => item.v);
const FAM_META_MAP = Object.fromEntries(FAM.map(f => [f[0], { n: f[1], y: f[2] }]));

function compareMcVersionsDesc(a, b) {
  const sa = String(a || '').trim();
  const sb = String(b || '').trim();
  if (sa === sb) return 0;
  if (sa.toLowerCase() === 'latest') return -1;
  if (sb.toLowerCase() === 'latest') return 1;
  const parse = s => {
    const core = s.split(/[-+]/)[0];
    return core.split('.').map(x => parseInt(x, 10) || 0);
  };
  const pa = parse(sa);
  const pb = parse(sb);
  const len = Math.max(pa.length, pb.length);
  for (let i = 0; i < len; i++) {
    const da = pa[i] || 0;
    const db = pb[i] || 0;
    if (da !== db) return db - da;
  }
  return sb.localeCompare(sa);
}

function getFamilyMetaForVersion(v) {
  const str = String(v || '').trim();
  const parts = str.split(/[-+]/)[0].split('.');
  const famKey = parts.length >= 2 && Number(parts[0]) === 1
    ? (Number(parts[1]) <= 6 ? '1.0' : `1.${parts[1]}`)
    : (parts.length >= 2 && Number(parts[0]) >= 26 ? `${parts[0]}.${parts[1]}` : parts[0]);
  const meta = FAM_META_MAP[famKey];
  if (meta) return { f: famKey, n: meta.n, y: meta.y };
  return { f: famKey || 'Other', n: Number(parts[0]) >= 26 ? 'Modern Release' : 'Minecraft Release', y: Number(parts[0]) >= 26 ? '2026+' : 'Release' };
}

function getAllMcVersions(engine = '') {
  const eng = String(engine || '').toLowerCase().trim();
  if (isSteelEngine(eng)) return Array.from(new Set(['latest', ...(S.steelVersions || [])]));
  const cat = S.versionsCatalog;
  if (eng === 'velocity') {
    const base = ['3.4.0-SNAPSHOT', '3.3.0-SNAPSHOT', '3.2.0-SNAPSHOT', '3.1.1'];
    const live = Array.isArray(cat?.by_engine?.velocity) ? cat.by_engine.velocity : [];
    return Array.from(new Set([...live, ...base]));
  }
  if (eng === 'waterfall') {
    const base = ['1.20', '1.19', '1.18', '1.17', '1.16', '1.12'];
    const live = Array.isArray(cat?.by_engine?.waterfall) ? cat.by_engine.waterfall : [];
    return Array.from(new Set([...live, ...base])).sort(compareMcVersionsDesc);
  }
  const merged = new Set(FALLBACK_ALL_MC);
  const filterAllowed = (v) => {
    if (!v) return false;
    const str = String(v).trim();
    return str.startsWith('1.21') || str.startsWith('26.') || POPULAR_MC_VERSIONS.has(str);
  };

  if (Array.isArray(cat?.all_mc)) {
    for (const v of cat.all_mc) {
      if (filterAllowed(v)) merged.add(String(v));
    }
  }
  if (eng && Array.isArray(cat?.by_engine?.[eng])) {
    for (const v of cat.by_engine[eng]) {
      if (filterAllowed(v)) merged.add(String(v));
    }
  }
  const curV = S.ver || curServer()?.version;
  if (curV) merged.add(curV);

  const sorted = Array.from(merged).sort(compareMcVersionsDesc);
  if (eng === 'bungeecord') {
    return ['latest', ...sorted.filter(v => v !== 'latest')];
  }
  return sorted;
}

function getMergedVersionItems(engine = '') {
  return getAllMcVersions(engine).map(v => {
    const m = getFamilyMetaForVersion(v);
    return { v, f: m.f, n: m.n, y: m.y };
  });
}

function refreshOpenVersionSelects() {
  const engineCS = $('#cs-type')?.value || 'paper';
  const csSel = $('#cs-ver-sel');
  if (csSel) {
    const curVal = $('#cs-ver')?.value?.trim() || csSel.value || '1.21.11';
    const vers = getAllMcVersions(engineCS);
    const list = (curVal && !vers.includes(curVal)) ? [curVal, ...vers] : vers;
    csSel.innerHTML = list.map(v => `<option value="${esc(v)}" ${v === curVal ? 'selected' : ''}>${esc(v)}</option>`).join('');
    if (typeof csSel._cselectRebuild === 'function') csSel._cselectRebuild();
  }
  const reinstSel = $('#reinst-ver-sel');
  if (reinstSel) {
    const engR = $('#reinst-software')?.value || 'paper';
    const curVal = $('#reinst-version')?.value?.trim() || reinstSel.value || '1.21.11';
    const vers = getAllMcVersions(engR);
    const list = (curVal && !vers.includes(curVal)) ? [curVal, ...vers] : vers;
    reinstSel.innerHTML = list.map(v => `<option value="${esc(v)}" ${v === curVal ? 'selected' : ''}>${esc(v)}</option>`).join('');
    if (typeof reinstSel._cselectRebuild === 'function') reinstSel._cselectRebuild();
  }
  const plugSel = $('#plug-ver-sel');
  if (plugSel) {
    const curVal = S.plugVer || '';
    const vers = getAllMcVersions(S.plugLoader || '');
    const list = (curVal && !vers.includes(curVal)) ? [curVal, ...vers] : vers;
    plugSel.innerHTML = `<option value="" ${!curVal ? 'selected' : ''}>Version: Any</option>` + list.map(v => `<option value="${esc(v)}" ${curVal === v ? 'selected' : ''}>${esc(v)}</option>`).join('');
    if (typeof plugSel._cselectRebuild === 'function') plugSel._cselectRebuild();
  }
  const propVerSel = $('#prop-ver-sel');
  if (propVerSel) {
    const srv = curServer();
    const curVal = $('#prop-ver')?.value?.trim() || srv?.version || '1.21.11';
    const vers = getAllMcVersions(srv?.type || S.core || '');
    const list = (curVal && !vers.includes(curVal)) ? [curVal, ...vers] : vers;
    propVerSel.innerHTML = list.map(v => `<option value="${esc(v)}" ${v === curVal ? 'selected' : ''}>${esc(v)}</option>`).join('');
    if (typeof propVerSel._cselectRebuild === 'function') propVerSel._cselectRebuild();
  }
  if (S.view === 'ver' && $('#vres')) {
    renderVersionRows();
  }
}

let steelVersionsRequest = null;
async function loadSteelVersions(force = false) {
  if (steelVersionsRequest) return steelVersionsRequest;
  if (!force && Array.isArray(S.steelVersions)) return S.steelVersions;
  steelVersionsRequest = (async () => {
    try {
      const releases = await api('/api/steel/versions');
      if (!Array.isArray(releases) || releases.some(v => typeof v !== 'string')) throw new Error('Invalid Steel release catalog');
      S.steelVersions = releases.filter(v => v.trim()).map(v => v.trim());
      refreshOpenVersionSelects();
      return S.steelVersions;
    } catch (e) {
      toast('Could not load Steel releases. You can still select latest or enter an official release tag.', 'bad');
      return [];
    } finally {
      steelVersionsRequest = null;
    }
  })();
  return steelVersionsRequest;
}

async function loadVersionsCatalog(force = false) {
  const selectedEngine = $('#cs-type')?.value || $('#reinst-software')?.value || (S.view === 'ver' ? S.sel || S.core : curServer()?.type);
  if (isSteelEngine(selectedEngine)) return loadSteelVersions(force);
  if (!S.token) return S.versionsCatalog;
  if (!force && S.versionsCatalog && Array.isArray(S.versionsCatalog.all_mc) && S.versionsCatalog.all_mc.length > 0) {
    return S.versionsCatalog;
  }
  try {
    const res = await api('/api/versions');
    if (res && (Array.isArray(res.all_mc) || res.by_engine)) {
      S.versionsCatalog = res;
      try { sessionStorage.setItem('neopanel_versions_catalog', JSON.stringify(res)); } catch (e) {}
      refreshOpenVersionSelects();
      return res;
    }
  } catch (e) {}
  return S.versionsCatalog;
}

const CATS = ['All', 'Admin', 'Protection', 'Performance', 'Network', 'Building', 'Utility', 'API', 'Map'];

const PREMADE_JAVA_VERSIONS = [
  { value: '25', short: 'Java 25', label: 'Java 25 (Latest LTS • Minecraft 26+)' },
  { value: '21', short: 'Java 21', label: 'Java 21 (LTS • Minecraft 1.20.5 - 1.21.x)' },
  { value: '17', short: 'Java 17', label: 'Java 17 (LTS • Minecraft 1.17 - 1.20.4)' },
  { value: '8', short: 'Java 8', label: 'Java 8 (Legacy • Minecraft 1.8 - 1.16.5)' }
];

function isStandardJavaChoice(val) {
  return ['25', '21', '17', '8', 'java', ''].includes(String(val ?? '').trim());
}

function getRecommendedJavaForMcVersion(ver, engine = '') {
  const eng = String(engine || '').toLowerCase().trim();
  if (isSteelEngine(eng)) return '';
  const v = String(ver || '').toLowerCase().trim();
  if (eng === 'velocity') {
    return v.startsWith('3.3') ? '17' : '21';
  }
  if (eng === 'waterfall' || eng === 'bungeecord') {
    return '21';
  }
  if (!v || v === 'latest') return '25';
  const parts = v.replace(/^v/i, '').split('.');
  const major = parseInt(parts[0] || '0', 10);
  const minor = parseInt(parts[1] || '0', 10);
  const patch = parseInt((parts[2] || '0').replace(/[^0-9].*$/, ''), 10) || 0;
  if (major >= 26) return '25';
  if (major === 1) {
    if (minor >= 21 || (minor === 20 && patch >= 5)) return '21';
    if (minor >= 17) return '17';
    if (minor > 0 && minor <= 16) return '8';
  }
  return '21';
}

function resolveJavaSelectValue(rawVal, mcVer = '', engine = '') {
  const v = String(rawVal ?? '').trim();
  if (!v || v === 'java' || v === 'default') {
    return mcVer ? getRecommendedJavaForMcVersion(mcVer, engine) : '21';
  }
  if (['25', '21', '17', '8'].includes(v)) return v;
  return 'custom';
}

function getPremadeJavaStatusMap(nodeId = null) {
  const nid = nodeId || curServer()?.node_id || 'local';
  const data = (S.javaRuntimesMap && S.javaRuntimesMap[nid]) || (nid === 'local' ? S.javaRuntimes : null);
  const map = new Map();
  const list = data?.premade || [];
  for (const item of list) {
    if (item && item.version) map.set(String(item.version), item);
  }
  return map;
}

function renderJavaSelectOptions(selectedVal, includeCustom = true, fallbackMcVer = '', engine = '', nodeId = null) {
  const cur = resolveJavaSelectValue(selectedVal, fallbackMcVer, engine);
  const statusMap = getPremadeJavaStatusMap(nodeId);
  const opts = PREMADE_JAVA_VERSIONS.map(item => {
    const info = statusMap.get(item.value);
    const suffix = info ? (info.installed ? ' [Installed]' : ' [Not Installed]') : '';
    return `<option value="${item.value}" ${cur === item.value ? 'selected' : ''}>${esc(item.label + suffix)}</option>`;
  }).join('');
  if (!includeCustom) return opts;
  return opts + `<option value="custom" ${cur === 'custom' ? 'selected' : ''}>Custom Binary Path...</option>`;
}

function renderJavaAvailabilityStrip(nodeId = null) {
  const nid = nodeId || curServer()?.node_id || 'local';
  const data = (S.javaRuntimesMap && S.javaRuntimesMap[nid]) || (nid === 'local' ? S.javaRuntimes : null);
  if (!data || !Array.isArray(data.premade)) {
    return `<div class="java-status-strip" data-node-id="${esc(nid)}"><span class="java-status-chip">Checking installed Java 25, 21, 17 &amp; 8 runtimes...</span></div>`;
  }
  const canInstall = Boolean(data.can_auto_install) && (S.user?.role === 'admin');
  return `<div class="java-status-strip" data-node-id="${esc(nid)}">
    ${data.premade.map(item => {
      const busy = S.installingJava === `${nid}:${item.version}` || S.installingJava === item.version;
      if (item.installed) {
        return `<span class="java-status-chip is-ok" title="${esc(item.path || 'Installed')}">
          <span class="java-status-dot"></span>Java ${esc(item.version)} Ready
        </span>`;
      }
      return `<span class="java-status-chip is-missing" title="Not detected in /usr/lib/jvm">
        <span class="java-status-dot"></span>Java ${esc(item.version)} Missing
        ${canInstall ? `<button type="button" class="btn sm pr java-install-inline ${busy ? 'is-loading' : ''}" onclick="installJavaRuntime('${esc(item.version)}', this, '${esc(nid)}')" ${busy || S.installingJava ? 'disabled' : ''}>${busy ? '<span class="np-spinner-sm"></span> Installing...' : 'Install'}</button>` : ''}
      </span>`;
    }).join('')}
  </div>`;
}

function refreshJavaUIStrips(nodeId = null) {
  const defaultNid = nodeId || curServer()?.node_id || 'local';
  document.querySelectorAll('.java-status-strip').forEach(el => {
    const elNid = el.dataset.nodeId || defaultNid;
    const stripHtml = renderJavaAvailabilityStrip(elNid);
    const tmp = document.createElement('div');
    tmp.innerHTML = stripHtml;
    if (tmp.firstElementChild) el.replaceWith(tmp.firstElementChild);
  });

  const csNodeVal = $('#cs-node')?.value || defaultNid;
  for (const [id, nid] of [['prop-java', curServer()?.node_id || defaultNid], ['cs-java', csNodeVal]]) {
    const sel = document.getElementById(id);
    if (!sel) continue;
    const statusMap = getPremadeJavaStatusMap(nid);
    for (const opt of Array.from(sel.options)) {
      const def = PREMADE_JAVA_VERSIONS.find(x => x.value === opt.value);
      if (def) {
        const info = statusMap.get(def.value);
        const suffix = info ? (info.installed ? ' [Installed]' : ' [Not Installed]') : '';
        opt.textContent = def.label + suffix;
      }
    }
    if (sel._cselectRebuild) sel._cselectRebuild();
  }
}

async function loadJavaRuntimes(force = false, nodeId = null, skipRender = false) {
  if (!S.token) return null;
  const nid = nodeId || curServer()?.node_id || 'local';
  if (!S.javaRuntimesMap) S.javaRuntimesMap = {};
  const cached = S.javaRuntimesMap[nid] || (nid === 'local' ? S.javaRuntimes : null);
  if (cached && !force) {
    refreshJavaUIStrips(nid);
    return cached;
  }
  try {
    const data = await api(nodeApiPath(nid, '/api/system/java?matrix=1'));
    if (data && Array.isArray(data.premade)) {
      S.javaRuntimesMap[nid] = data;
      if (nid === 'local') S.javaRuntimes = data;
      refreshJavaUIStrips(nid);
      if (!skipRender && S.view === 'admin' && (S.adminTab || 0) === 0) R();
    }
    return data;
  } catch (e) {
    return null;
  }
}

async function installJavaRuntime(ver, btnEl, nodeId = null) {
  const nid = nodeId || curServer()?.node_id || 'local';
  if (S.installingJava) {
    toast(`Already installing Java runtime...`);
    return;
  }
  S.installingJava = `${nid}:${ver}`;
  if (btnEl) setBtnLoading(btnEl, true, 'Installing...');
  refreshJavaUIStrips(nid);
  toast(`Downloading and installing Java ${ver} (Eclipse Temurin)...`);
  try {
    const res = await api(nodeApiPath(nid, '/api/system/java/install'), {
      method: 'POST',
      body: JSON.stringify({ version: ver })
    });
    if (res && Array.isArray(res.premade)) {
      const next = {
        installed: res.installed || [],
        premade: res.premade,
        can_auto_install: Boolean(res.can_auto_install)
      };
      if (!S.javaRuntimesMap) S.javaRuntimesMap = {};
      S.javaRuntimesMap[nid] = next;
      if (nid === 'local') S.javaRuntimes = next;
    } else {
      await loadJavaRuntimes(true, nid);
    }
    toast(`Java ${ver} installed and ready!`, 'ok');
  } catch (e) {
    toast(e.message || `Failed to install Java ${ver}`, 'bad');
  } finally {
    S.installingJava = null;
    refreshJavaUIStrips(nid);
    if (S.view === 'admin' && (S.adminTab || 0) === 0) R();
  }
}

// --- Header Banner Component (.srvh) ---
function srvh() {
  const srv = curServer();
  const st = S.status;
  const on = st === 'online';
  const active = on || st === 'starting';
  const c = active ? (S.cpu.at(-1) || 0) : 0;
  const maxRamMB = S.memMaxMB || getServerMaxRamMB(srv);
  const m = active ? (S.memUsedMB || ((S.ram.at(-1) || 0) / 100) * maxRamMB) : 0;
  const hostAddr = `${srvHost(srv)}:${srv ? srv.port : 25565}`;
  const nodeId = srv?.node_id || 'local';
  const nodeLabel = displayNodeName(nodeId, srv?.node_name || getNodeById(nodeId)?.name);
  const heroUrl = getBrand('server_hero_url');
  const srvType = srv?.type || S.core || 'paper';
  const eng = engineMeta(srvType);
  const panelName = (S.activeServerId === srv?.id && typeof S.editSrvName === 'string' && S.editSrvName.trim())
    ? S.editSrvName.trim()
    : (srv?.name || 'Minecraft Server');
  const panelDesc = (S.activeServerId === srv?.id && typeof S.editSrvDesc === 'string')
    ? S.editSrvDesc.trim()
    : String(srv?.description || '').trim();
  const rawMotd = (S.props && S.props._srvId === srv?.id && S.props.motd)
    ? S.props.motd
    : (srv?.motd || '');

  return `<div class="card srvh ${heroUrl ? 'has-brand-art' : ''}">
    <div class="srvh-bg" ${heroUrl ? `style="background-image: url('${esc(heroUrl)}');"` : ''} aria-hidden="true"></div>
    <div class="srvh-main">
      <div class="srv-hero-meta">
        <h1 id="srv-hero-name">${renderMinecraftTextHtml(panelName, { allowNewlines: false })}</h1>
        <span id="srv-badge" class="badge ${st}"><i class="dot"></i>${{ online: 'Running', offline: 'Offline', starting: 'Starting', stopping: 'Stopping' }[st] || 'Offline'}</span>
        ${((S.nodes && S.nodes.length > 1) || nodeId !== 'local') ? `<span class="chip srv-node-badge" title="Running on node ${esc(nodeLabel)}">${ic('network', 11)} ${esc(nodeLabel)}</span>` : ''}
        <span id="srv-halt-badge" class="badge halted" style="display:${st === 'offline' && S.startupHalted ? 'inline-flex' : 'none'};" title="Auto-restart stopped after 2 consecutive startup failures">Start Halted (2/2)</span>
      </div>
      <div id="srv-hero-title" class="srv-hero-sub" style="display:${rawMotd ? 'block' : 'none'};">${rawMotd ? renderMinecraftTextHtml(rawMotd, { allowNewlines: true }) : ''}</div>
      <div id="srv-hero-desc" class="srv-hero-sub" style="display:${panelDesc ? 'block' : 'none'};">${esc(panelDesc)}</div>
      <div class="srv-connection">
        <span id="srv-hero-engine" class="eng-pill" title="Server software">${engineIcon(srvType, 13)} ${esc(eng.label)}${srv?.version ? ` ${esc(srv.version)}` : ''}</span>
        <div id="hst" class="row hst" style="margin-top: 0;">
          <button type="button" class="copyable-ip mono mu" data-addr="${esc(hostAddr)}" onclick="return copyServerAddress(this.dataset.addr, event)" title="Copy server address" aria-label="Copy server address"><span class="pv">${esc(hostAddr)}</span> ${ic('copy', 12)}</button>
          <span class="mono mu">CPU ${c.toFixed(1)}%</span>
          <span class="mono mu">RAM ${m.toFixed(0)} MiB / ${(maxRamMB / 1024).toFixed(1)} GiB</span>
        </div>
      </div>
    </div>
    <div class="row srv-power-actions" role="group" aria-label="Server power controls">
      ${S.view !== 'con' ? `<button class="btn" onclick="toggleFlyConsole()" title="Open Live Console Drawer (Ctrl+\`)">${ic('terminal', 14)} Console</button>` : ''}
      <button class="btn pr" data-power-btn="start" onclick="power('start')" ${st !== 'offline' ? 'disabled' : ''}>${ic('play', 14)} Start</button>
      <button class="btn" data-power-btn="restart" onclick="power('restart')" ${on ? '' : 'disabled'}>${ic('refresh', 14)} Restart</button>
      <button class="btn stop" data-power-btn="stop" onclick="power('stop')" ${on || st === 'starting' ? '' : 'disabled'}>Stop</button>
      <button class="btn dg" onclick="power('kill')" title="Stop immediately; unsaved data may be lost">Force stop</button>
    </div>
  </div>`;
}

function toggleCmdMenu() {
  S.cmenu = !S.cmenu;
  R();
}

function execQuickCmd(cmd) {
  S.cmenu = false;
  runCmd(cmd);
  R();
}

function statsHtml() {
  const srv = curServer();
  const active = S.status === 'online' || S.status === 'starting';
  const c = active ? (S.cpu.at(-1) || 0) : 0;
  const maxRamMB = S.memMaxMB || getServerMaxRamMB(srv);
  const m = active ? (S.memUsedMB || ((S.ram.at(-1) || 0) / 100) * maxRamMB) : 0;
  const netIn = active ? (S.ni.at(-1) || 0) : 0;
  const netOut = active ? (S.no.at(-1) || 0) : 0;
  const hostAddr = `${srvHost(srv)}:${srv ? srv.port : 25565}`;

  const row = (i, l, v, s) => `<div class="srow">
    <span class="sic" aria-hidden="true">${i}</span>
    <div>
      <div class="mu">${l}</div>
      <b class="mono">${v}</b>${s ? ` <span class="mu mono">${s}</span>` : ''}
    </div>
  </div>`;

  return row(ic('network', 16), 'Address', `<span class="copyable-ip" data-addr="${esc(hostAddr)}" onclick="return copyServerAddress(this.dataset.addr, event)" title="Click to copy server address"><span class="pv">${esc(hostAddr)}</span> ${ic('copy', 12)}</span>`) +
    row(ic('clock', 16), 'Uptime', up()) +
    row(ic('cpu', 16), 'CPU load', `${c.toFixed(1)}%`, '/ 100%') +
    row(ic('layers', 16), 'Memory', `${m.toFixed(0)} MiB`, `/ ${(maxRamMB / 1024).toFixed(1)} GiB`) +
    row(ic('database', 16), 'Disk', formatDiskMB(S.diskMB)) +
    row('↓', 'Network (inbound)', `${netIn.toFixed(1)} KiB/s`) +
    row('↑', 'Network (outbound)', `${netOut.toFixed(1)} KiB/s`);
}

function chartsHtml() {
  const srv = curServer();
  const active = S.status === 'online' || S.status === 'starting';
  const maxRamMB = S.memMaxMB || getServerMaxRamMB(srv);
  const maxRamGB = (maxRamMB / 1024).toFixed(1);
  const cpuStr = active ? `${(S.cpu.at(-1) || 0).toFixed(1)}%` : '0.0%';
  const memStr = active ? `${(S.memUsedMB || 0).toFixed(0)} MiB / ${maxRamGB} GiB` : `0 MiB / ${maxRamGB} GiB`;
  const renderCard = (kind, title, valHtml) => `
    <div class="card ch" data-chart-kind="${kind}" onmousemove="onChartMouseMove(event, '${kind}')" onmouseleave="onChartMouseLeave('${kind}')">
      <div class="row sb">
        <span class="mu">${title}</span>
        <span id="ch-val-${kind}" class="mu mono">${valHtml}</span>
      </div>
      <div class="ch-wrap">
        <svg class="ch-svg" viewBox="0 0 300 98" preserveAspectRatio="none" aria-hidden="true">
          <g id="ch-series-${kind}">${buildChartSeriesInnerSvg(kind)}</g>
          <g id="ch-hover-svg-${kind}"></g>
          <rect class="ch-hit-layer" x="0" y="0" width="300" height="98" fill="transparent" pointer-events="all"/>
        </svg>
        <div id="ch-tip-${kind}" class="ch-tooltip" style="display:none;"></div>
      </div>
    </div>`;

  return renderCard('cpu', 'CPU load', cpuStr) +
    renderCard('mem', 'Memory', memStr) +
    renderCard('net', 'Network I/O', '<span style="color:var(--ac)">in</span> <span style="color:#f0b429">out</span>');
}

function liveCon() {
  const srv = curServer();
  const active = S.status === 'online' || S.status === 'starting';
  const c = active ? (S.cpu.at(-1) || 0) : 0;
  const maxRamMB = S.memMaxMB || getServerMaxRamMB(srv);
  const maxRamGB = (maxRamMB / 1024).toFixed(1);
  const m = active ? (S.memUsedMB || ((S.ram.at(-1) || 0) / 100) * maxRamMB) : 0;
  const hostAddr = `${srvHost(srv)}:${srv ? srv.port : 25565}`;

  const stc = $('#stc');
  const chr = $('#chr');
  const hst = $('#hst');
  if (stc) stc.innerHTML = statsHtml();

  if (chr) {
    const cpuSeriesEl = $('#ch-series-cpu');
    const memSeriesEl = $('#ch-series-mem');
    const netSeriesEl = $('#ch-series-net');
    if (cpuSeriesEl && memSeriesEl && netSeriesEl) {
      const cpuValEl = $('#ch-val-cpu');
      const memValEl = $('#ch-val-mem');
      if (cpuValEl) cpuValEl.textContent = active ? `${c.toFixed(1)}%` : '0.0%';
      if (memValEl) memValEl.textContent = active ? `${m.toFixed(0)} MiB / ${maxRamGB} GiB` : `0 MiB / ${maxRamGB} GiB`;
      cpuSeriesEl.innerHTML = buildChartSeriesInnerSvg('cpu');
      memSeriesEl.innerHTML = buildChartSeriesInnerSvg('mem');
      netSeriesEl.innerHTML = buildChartSeriesInnerSvg('net');
      if (S.chartHoverKind) {
        updateChartHoverDom(S.chartHoverKind);
      }
    } else {
      chr.innerHTML = chartsHtml();
    }
  }

  if (hst) {
    hst.innerHTML = `<button type="button" class="copyable-ip mono mu" data-addr="${esc(hostAddr)}" onclick="return copyServerAddress(this.dataset.addr, event)" title="Copy server address" aria-label="Copy server address"><span class="pv">${esc(hostAddr)}</span> ${ic('copy', 12)}</button><span class="mono mu">CPU ${c.toFixed(1)}%</span><span class="mono mu">RAM ${m.toFixed(0)} MiB / ${maxRamGB} GiB</span>`;
  }
}

function updateDashLiveStats() {
  if (S.view !== 'dash') return;
  let runningCount = 0;
  let totalRamAllocatedMB = 0;
  let totalRamUsedMB = 0;
  let totalDiskMB = 0;

  for (let i = 0; i < S.servers.length; i++) {
    const s = S.servers[i];
    const st = i === S.cs ? S.status : normalizeServerStatus(s.stats?.status || s.status);
    const isRun = st === 'online' || st === 'starting';
    if (isRun) runningCount++;

    const maxRam = getServerMaxRamMB(s);
    const cVal = isRun ? (i === S.cs ? (S.cpu.at(-1) || 0) : (Number(s.stats?.cpu_usage) || 0)) : 0;
    const mVal = isRun ? (i === S.cs ? (S.memUsedMB || 0) : (Number(s.stats?.memory_mb) || 0)) : 0;
    const dVal = i === S.cs ? (S.diskMB || Number(s.stats?.disk_mb) || 0) : (Number(s.stats?.disk_mb) || 0);
    const upVal = i === S.cs ? S.uptimeSec : (Number(s.stats?.uptime_sec) || 0);

    totalRamAllocatedMB += maxRam;
    totalRamUsedMB += mVal;
    totalDiskMB += dVal;

    const nid = s.node_id || 'local';
    const card = document.getElementById(`dash-srv-${nid}-${s.id}`) || document.getElementById(`dash-srv-${s.id}`);
    if (!card) continue;
    const badge = card.querySelector('[data-role="badge"]');
    if (badge) {
      badge.className = `badge ${isRun ? 'online' : 'offline'}`;
      badge.innerHTML = `<i class="dot"></i>${isRun ? 'Running' : 'Offline'}`;
    }
    const upEl = card.querySelector('[data-role="uptime"]');
    if (upEl) {
      if (isRun) {
        const isList = card.classList.contains('vm-dash-row');
        upEl.innerHTML = `${isList ? '&bull; ' : ''}Up ${formatUptime(upVal)}`;
        upEl.style.display = '';
      } else {
        upEl.innerHTML = '';
        upEl.style.display = 'none';
      }
    }
    const cpuValEl = card.querySelector('[data-role="cpu-val"]');
    const cpuBarEl = card.querySelector('[data-role="cpu-bar"]');
    if (cpuValEl) cpuValEl.textContent = `${cVal.toFixed(1)}%`;
    if (cpuBarEl) cpuBarEl.style.width = `${Math.min(100, cVal)}%`;

    const memValEl = card.querySelector('[data-role="mem-val"]');
    const memBarEl = card.querySelector('[data-role="mem-bar"]');
    if (memValEl) memValEl.textContent = `${Math.round(mVal)} MB`;
    if (memBarEl) memBarEl.style.width = `${Math.min(100, (mVal / (maxRam || 1)) * 100)}%`;

    const diskValEl = card.querySelector('[data-role="disk-val"]');
    const diskBarEl = card.querySelector('[data-role="disk-bar"]');
    if (diskValEl) diskValEl.textContent = formatDiskMB(dVal);
    if (diskBarEl) diskBarEl.style.width = `${Math.min(100, (dVal / 20480) * 100)}%`;
  }

  const totalCount = S.servers.length;
  const runSummary = $('#dash-running-summary');
  if (runSummary) runSummary.innerHTML = `${runningCount} of ${totalCount} servers running`;

  const allocMemVal = $('#dash-alloc-mem-val');
  const allocMemBar = $('#dash-alloc-mem-bar');
  const totalAllocGB = (totalRamAllocatedMB / 1024).toFixed(1);
  const usedGB = (totalRamUsedMB / 1024).toFixed(2);
  if (allocMemVal) allocMemVal.textContent = `${usedGB} of ${totalAllocGB} GB`;
  if (allocMemBar) allocMemBar.style.width = `${Math.min(100, (totalRamUsedMB / (totalRamAllocatedMB || 1)) * 100)}%`;

  const allocDiskVal = $('#dash-alloc-disk-val');
  const allocDiskBar = $('#dash-alloc-disk-bar');
  if (allocDiskVal) allocDiskVal.textContent = `${formatDiskMB(totalDiskMB)} used`;
  if (allocDiskBar) allocDiskBar.style.width = `${Math.min(100, (totalDiskMB / Math.max(20480, totalCount * 10240)) * 100)}%`;

  const allocRunVal = $('#dash-alloc-run-val');
  const allocRunBar = $('#dash-alloc-run-bar');
  if (allocRunVal) allocRunVal.textContent = `${runningCount} of ${totalCount}`;
  if (allocRunBar) allocRunBar.style.width = `${Math.min(100, (runningCount / (totalCount || 1)) * 100)}%`;
}

// --- Rich Activity & Audit Log Engine ---
const ACTIVITY_META_MAP = {
  power_start: { category: 'power', badgeLabel: 'Power', cls: 'act-cat-power', title: 'Started Server' },
  power_stop: { category: 'power', badgeLabel: 'Power', cls: 'act-cat-power', title: 'Stopped Server' },
  power_restart: { category: 'power', badgeLabel: 'Power', cls: 'act-cat-power', title: 'Restarted Server' },
  power_kill: { category: 'power', badgeLabel: 'Power', cls: 'act-cat-power', title: 'Server force-stopped' },
  startup_failed: { category: 'power', badgeLabel: 'Timeout', cls: 'act-cat-security', title: 'Startup Halted (2/2)' },
  server_create: { category: 'admin', badgeLabel: 'Server', cls: 'act-cat-admin', title: 'Created Server Instance' },
  server_update: { category: 'software', badgeLabel: 'Settings', cls: 'act-cat-software', title: 'Updated Server Config' },
  server_delete: { category: 'security', badgeLabel: 'Server', cls: 'act-cat-security', title: 'Deleted Server Instance' },
  server_reinstall: { category: 'software', badgeLabel: 'Software', cls: 'act-cat-software', title: 'Reinstalled Server' },
  console_command: { category: 'console', badgeLabel: 'Console', cls: 'act-cat-console', title: 'Console Command' },
  file_save: { category: 'files', badgeLabel: 'Files', cls: 'act-cat-files', title: 'Saved File' },
  file_create: { category: 'files', badgeLabel: 'Files', cls: 'act-cat-files', title: 'Created File / Folder' },
  file_delete: { category: 'files', badgeLabel: 'Files', cls: 'act-cat-files', title: 'Deleted File / Folder' },
  file_upload: { category: 'files', badgeLabel: 'Files', cls: 'act-cat-files', title: 'Uploaded File' },
  file_download: { category: 'files', badgeLabel: 'Files', cls: 'act-cat-files', title: 'Downloaded File' },
  file_rename: { category: 'files', badgeLabel: 'Files', cls: 'act-cat-files', title: 'Renamed Item' },
  file_move: { category: 'files', badgeLabel: 'Files', cls: 'act-cat-files', title: 'Moved Items' },
  file_compress: { category: 'files', badgeLabel: 'Files', cls: 'act-cat-files', title: 'Created Archive' },
  file_unzip: { category: 'files', badgeLabel: 'Files', cls: 'act-cat-files', title: 'Extracted Archive' },
  plugin_install: { category: 'plugins', badgeLabel: 'Plugins', cls: 'act-cat-plugins', title: 'Installed Plugin' },
  plugin_delete: { category: 'plugins', badgeLabel: 'Plugins', cls: 'act-cat-plugins', title: 'Uninstalled Plugin' },
  mod_install: { category: 'plugins', badgeLabel: 'Mods', cls: 'act-cat-plugins', title: 'Installed Mod' },
  mod_delete: { category: 'plugins', badgeLabel: 'Mods', cls: 'act-cat-plugins', title: 'Uninstalled Mod' },
  software_install: { category: 'software', badgeLabel: 'Software', cls: 'act-cat-software', title: 'Installed Engine JAR' },
  properties_save: { category: 'software', badgeLabel: 'Config', cls: 'act-cat-software', title: 'Saved server.properties' },
  backup_create: { category: 'backups', badgeLabel: 'Backups', cls: 'act-cat-backups', title: 'Created Backup' },
  backup_delete: { category: 'backups', badgeLabel: 'Backups', cls: 'act-cat-backups', title: 'Deleted Backup' },
  backup_download: { category: 'backups', badgeLabel: 'Backups', cls: 'act-cat-backups', title: 'Downloaded Backup' },
  port_add: { category: 'network', badgeLabel: 'Network', cls: 'act-cat-network', title: 'Assigned Port' },
  port_delete: { category: 'network', badgeLabel: 'Network', cls: 'act-cat-network', title: 'Removed Port' },
  port_primary: { category: 'network', badgeLabel: 'Network', cls: 'act-cat-network', title: 'Changed Primary Port' },
  subdomain_update: { category: 'network', badgeLabel: 'DNS', cls: 'act-cat-network', title: 'Updated Subdomain' },
  player_op: { category: 'players', badgeLabel: 'Players', cls: 'act-cat-players', title: 'Granted Operator' },
  player_deop: { category: 'players', badgeLabel: 'Players', cls: 'act-cat-players', title: 'Revoked Operator' },
  player_kick: { category: 'players', badgeLabel: 'Players', cls: 'act-cat-players', title: 'Kicked Player' },
  player_ban: { category: 'players', badgeLabel: 'Players', cls: 'act-cat-players', title: 'Banned Player' },
  player_pardon: { category: 'players', badgeLabel: 'Players', cls: 'act-cat-players', title: 'Pardoned Player' },
  player_whitelist_add: { category: 'players', badgeLabel: 'Players', cls: 'act-cat-players', title: 'Whitelisted Player' },
  player_whitelist_remove: { category: 'players', badgeLabel: 'Players', cls: 'act-cat-players', title: 'Removed Whitelist' },
  auth_setup: { category: 'security', badgeLabel: 'Security', cls: 'act-cat-security', title: 'Panel Initialized' },
  auth_login: { category: 'security', badgeLabel: 'Security', cls: 'act-cat-security', title: 'Signed In' },
  auth_login_failed: { category: 'security', badgeLabel: 'Alert', cls: 'act-cat-security', title: 'Failed Sign-In' },
  password_change: { category: 'security', badgeLabel: 'Security', cls: 'act-cat-security', title: 'Changed Password' },
  ssh_key_add: { category: 'security', badgeLabel: 'SSH Key', cls: 'act-cat-security', title: 'Added SSH Key' },
  ssh_key_delete: { category: 'security', badgeLabel: 'SSH Key', cls: 'act-cat-security', title: 'Deleted SSH Key' },
  api_key_create: { category: 'security', badgeLabel: 'API Key', cls: 'act-cat-security', title: 'Created API Key' },
  api_key_delete: { category: 'security', badgeLabel: 'API Key', cls: 'act-cat-security', title: 'Revoked API Key' },
  admin_user_create: { category: 'admin', badgeLabel: 'Admin', cls: 'act-cat-admin', title: 'Created Panel User' },
  admin_user_delete: { category: 'admin', badgeLabel: 'Admin', cls: 'act-cat-admin', title: 'Deleted Panel User' },
  admin_turnstile: { category: 'admin', badgeLabel: 'Admin', cls: 'act-cat-admin', title: 'Updated Turnstile' },
  admin_branding: { category: 'admin', badgeLabel: 'Theme', cls: 'act-cat-admin', title: 'Updated Panel Look' },
  admin_java_install: { category: 'admin', badgeLabel: 'Admin', cls: 'act-cat-admin', title: 'Installed Java Runtime' },
  admin_gc: { category: 'admin', badgeLabel: 'System', cls: 'act-cat-admin', title: 'Garbage Collection' },
  db_create: { category: 'databases', badgeLabel: 'Database', cls: 'act-cat-databases', title: 'Created Database' },
  db_delete: { category: 'databases', badgeLabel: 'Database', cls: 'act-cat-databases', title: 'Deleted Database' },
  db_rotate: { category: 'databases', badgeLabel: 'Database', cls: 'act-cat-databases', title: 'Rotated DB Credentials' },
  db_query: { category: 'databases', badgeLabel: 'SQL', cls: 'act-cat-databases', title: 'Executed SQL Query' },
  db_export: { category: 'databases', badgeLabel: 'Export', cls: 'act-cat-databases', title: 'Exported Database' },
  node_add: { category: 'nodes', badgeLabel: 'Nodes', cls: 'act-cat-nodes', title: 'Connected Cluster Node' },
  node_update: { category: 'nodes', badgeLabel: 'Nodes', cls: 'act-cat-nodes', title: 'Updated Cluster Node' },
  node_delete: { category: 'nodes', badgeLabel: 'Nodes', cls: 'act-cat-nodes', title: 'Removed Cluster Node' }
};

const ACTIVITY_CATEGORIES = [
  ['all', 'All Events'],
  ['power', 'Power'],
  ['console', 'Console'],
  ['files', 'Files'],
  ['databases', 'Databases'],
  ['nodes', 'Cluster Nodes'],
  ['plugins', 'Plugins & Mods'],
  ['software', 'Software & Config'],
  ['backups', 'Backups'],
  ['network', 'Network & DNS'],
  ['players', 'Players'],
  ['security', 'Security & Auth'],
  ['admin', 'Admin & System']
];

function getActivityMeta(action) {
  const key = String(action || '').toLowerCase().trim();
  if (ACTIVITY_META_MAP[key]) return ACTIVITY_META_MAP[key];
  if (key.startsWith('power_')) return { category: 'power', badgeLabel: 'Power', cls: 'act-cat-power', title: key.replace(/_/g, ' ') };
  if (key.startsWith('file_')) return { category: 'files', badgeLabel: 'Files', cls: 'act-cat-files', title: key.replace(/_/g, ' ') };
  if (key.startsWith('db_')) return { category: 'databases', badgeLabel: 'Database', cls: 'act-cat-databases', title: key.replace(/_/g, ' ') };
  if (key.startsWith('node_')) return { category: 'nodes', badgeLabel: 'Nodes', cls: 'act-cat-nodes', title: key.replace(/_/g, ' ') };
  if (key.startsWith('plugin_') || key.startsWith('mod_')) return { category: 'plugins', badgeLabel: key.startsWith('mod_') ? 'Mods' : 'Plugins', cls: 'act-cat-plugins', title: key.replace(/_/g, ' ') };
  if (key.startsWith('backup_')) return { category: 'backups', badgeLabel: 'Backups', cls: 'act-cat-backups', title: key.replace(/_/g, ' ') };
  if (key.startsWith('port_') || key.startsWith('subdomain_')) return { category: 'network', badgeLabel: 'Network', cls: 'act-cat-network', title: key.replace(/_/g, ' ') };
  if (key.startsWith('player_')) return { category: 'players', badgeLabel: 'Players', cls: 'act-cat-players', title: key.replace(/_/g, ' ') };
  if (key.startsWith('auth_') || key.includes('key') || key.includes('password')) {
    return { category: 'security', badgeLabel: 'Security', cls: 'act-cat-security', title: key.replace(/_/g, ' ') };
  }
  if (key.startsWith('admin_') || key.startsWith('server_')) {
    return { category: 'admin', badgeLabel: 'Admin', cls: 'act-cat-admin', title: key.replace(/_/g, ' ') };
  }
  const readable = key ? key.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase()) : 'Panel Event';
  return { category: 'admin', badgeLabel: 'System', cls: 'act-cat-admin', title: readable };
}

function getUserTimezone() {
  return localStorage.getItem('neopanel_timezone') || S.branding?.timezone || 'UTC';
}

function formatFullDate(iso) {
  if (!iso) return '—';
  const d = new Date(iso);
  if (isNaN(d.getTime())) return String(iso);
  const tz = getUserTimezone();
  try {
    return d.toLocaleString(undefined, { timeZone: tz });
  } catch (e) {
    return d.toLocaleString();
  }
}

function getScopedActivityRows(scope = 'all') {
  const all = Array.isArray(S.acct.act) ? S.acct.act : [];
  if (scope !== 'mine' || !S.acct.un) return all;
  const me = String(S.acct.un).toLowerCase();
  return all.filter(r => !r.username || String(r.username).toLowerCase() === me);
}

function getFilteredActivity(scope = 'all') {
  const rows = getScopedActivityRows(scope);
  const q = (S.actFilter?.q || '').toLowerCase().trim();
  const cat = S.actFilter?.category || 'all';

  return rows.filter(r => {
    const action = r.action || (Array.isArray(r) ? r[0] : '');
    const meta = getActivityMeta(action);
    if (cat !== 'all' && meta.category !== cat) return false;
    if (!q) return true;
    const actor = r.username || (Array.isArray(r) ? r[1] : '') || '';
    const ip = r.ip_address || r.ip || '';
    const details = r.details || '';
    const hay = `${action} ${meta.title} ${meta.badgeLabel} ${details} ${actor} ${ip}`.toLowerCase();
    return hay.includes(q);
  });
}

function renderDashRecentActivityHtml() {
  const rows = Array.isArray(S.acct.act) ? S.acct.act.slice(0, 6) : [];
  if (S.loading?.activity && rows.length === 0) {
    return loaderHtml('Loading recent activity...');
  }
  if (rows.length === 0) {
    return `<div class="mu" style="font-size: 12.5px; line-height: 1.5; padding: 6px 0;">
      No activity events recorded yet. Actions like starting servers, sending commands, editing files, and installing plugins will appear here.
    </div>`;
  }
  return `<div class="act-dash-list">
    ${rows.map(r => {
      const meta = getActivityMeta(r.action);
      const actor = r.username || (r.user_id === 0 ? 'Anonymous' : (S.acct.un || 'system'));
      const ip = r.ip_address || '127.0.0.1';
      const rel = r.created_at ? formatFileDate(r.created_at) : 'Recently';
      const full = formatFullDate(r.created_at);
      return `<div class="act-dash-item">
        <div style="min-width: 0; flex: 1;">
          <div class="row" style="gap: 6px; margin-bottom: 3px; flex-wrap: wrap;">
            <span class="act-cat-badge ${meta.cls}">${esc(meta.badgeLabel)}</span>
            <b style="color: var(--tx); font-size: 12.5px;">${esc(meta.title)}</b>
          </div>
          <div class="mu" style="font-size: 11.5px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;" title="${esc(r.details || '')}">
            ${esc(r.details || r.action || 'Event')}
          </div>
          <div class="mu mono" style="font-size: 11px; margin-top: 2px;">
            ${esc(actor)} &bull; <span class="pv">${esc(ip)}</span>
          </div>
        </div>
        <span class="mu" style="font-size: 11.5px; white-space: nowrap; flex-shrink: 0;" title="${esc(full)}">${esc(rel)}</span>
      </div>`;
    }).join('')}
  </div>`;
}

function buildActivityStatsHtml(scope = 'all') {
  const rows = getScopedActivityRows(scope);
  const counts = {};
  for (const r of rows) {
    const meta = getActivityMeta(r.action);
    counts[meta.category] = (counts[meta.category] || 0) + 1;
  }
  const activeCat = S.actFilter?.category || 'all';
  const activePills = [
    ['all', 'Total', rows.length],
    ['power', 'Power', counts.power || 0],
    ['console', 'Console', counts.console || 0],
    ['files', 'Files', counts.files || 0],
    ['databases', 'Databases', counts.databases || 0],
    ['nodes', 'Nodes', counts.nodes || 0],
    ['plugins', 'Plugins', counts.plugins || 0],
    ['software', 'Software', counts.software || 0],
    ['backups', 'Backups', counts.backups || 0],
    ['network', 'Network', counts.network || 0],
    ['players', 'Players', counts.players || 0],
    ['security', 'Security', counts.security || 0],
    ['admin', 'Admin', counts.admin || 0]
  ].filter(([k, , n]) => k === 'all' || n > 0 || k === activeCat);

  return activePills.map(([k, lbl, n]) => `
    <button type="button" class="act-stat-pill" onclick="setActivityCategory('${k}', true)"
      style="cursor:pointer; ${activeCat === k ? 'border-color: var(--ac); background: var(--ac-subtle); color: var(--tx);' : ''}">
      <span>${esc(lbl)}</span>
      <b>${n}</b>
    </button>
  `).join('');
}

function renderActivityRowsHtml(rows, showActor = true) {
  const cols = showActor ? 5 : 4;
  if (S.loading?.activity && !rows.length) {
    return `<tr><td colspan="${cols}">${loaderHtml('Loading audit trail...')}</td></tr>`;
  }
  if (!rows.length) {
    const hasFilter = Boolean((S.actFilter?.q || '').trim()) || (S.actFilter?.category || 'all') !== 'all';
    return `<tr><td colspan="${cols}" class="mu" style="text-align:center; padding: 28px;">
      ${hasFilter ? 'No audit events match your current search or category filter.' : 'No activity events recorded yet.'}
    </td></tr>`;
  }
  return rows.map(r => {
    const meta = getActivityMeta(r.action || (Array.isArray(r) ? r[0] : ''));
    const actor = r.username || (r.user_id === 0 ? 'anonymous' : (Array.isArray(r) ? r[1] : '') || S.acct.un || 'system');
    const ip = r.ip_address || r.ip || (Array.isArray(r) ? r[2] : '') || '127.0.0.1';
    const rel = r.created_at ? formatFileDate(r.created_at) : (Array.isArray(r) ? r[3] : 'Recently');
    const full = formatFullDate(r.created_at);
    const details = r.details || r.action || 'Event';

    return `<tr>
      <td><span class="act-cat-badge ${meta.cls}">${esc(meta.badgeLabel)}</span></td>
      <td style="min-width: 240px;">
        <b style="color: var(--tx); display: block;">${esc(meta.title)}</b>
        <div class="mu" style="font-size: 12px; margin-top: 2px; word-break: break-word;">${esc(details)}</div>
      </td>
      ${showActor ? `<td><span class="chip">${esc(actor)}</span></td>` : ''}
      <td class="mono pv" style="white-space: nowrap;">${esc(ip)}</td>
      <td style="white-space: nowrap;">
        <div style="color: var(--tx); font-size: 12.5px;">${esc(rel)}</div>
        <div class="mu mono" style="font-size: 11px;">${esc(full)}</div>
      </td>
    </tr>`;
  }).join('');
}

function renderActivityRows() {
  const tbody = $('#act-tbody');
  if (!tbody) return;
  const scope = tbody.dataset.scope || 'all';
  const showActor = tbody.dataset.showActor !== '0';
  const filtered = getFilteredActivity(scope);
  tbody.innerHTML = renderActivityRowsHtml(filtered, showActor);

  const countEl = $('#act-count-pill');
  if (countEl) {
    const total = getScopedActivityRows(scope).length;
    countEl.textContent = filtered.length === total ? `${total} events` : `${filtered.length} of ${total} events`;
  }
  const statsEl = $('#act-stats');
  if (statsEl) {
    statsEl.innerHTML = buildActivityStatsHtml(scope);
  }
}

function renderActivityLogSection(opts = {}) {
  const scope = opts.scope || 'all';
  const title = opts.title || 'Activity & Audit Log';
  const subtitle = opts.subtitle || 'Complete trail of server power events, commands, file operations, plugins, backups & logins';
  const showActor = opts.showActor !== false;
  const filtered = getFilteredActivity(scope);
  const totalScoped = getScopedActivityRows(scope).length;
  const curLimit = S.actFilter?.limit || 200;

  return `
    <div class="card" style="padding: 18px 20px; margin-bottom: 14px;">
      <div class="row sb" style="margin-bottom: 14px; flex-wrap: wrap; gap: 12px;">
        <div>
          <div class="row" style="gap: 10px; align-items: center;">
            <h2 style="margin: 0; color: var(--tx); font-size: 17px;">${esc(title)}</h2>
            <span id="act-count-pill" class="chip">${filtered.length === totalScoped ? `${totalScoped} events` : `${filtered.length} of ${totalScoped} events`}</span>
          </div>
          <div class="mu" style="font-size: 12.5px; margin-top: 3px;">${esc(subtitle)}</div>
        </div>
        <div class="row" style="gap: 8px; flex-wrap: wrap;">
          <button type="button" class="btn sm" onclick="loadActivityLog(); toast('Activity logs refreshed');">${ic('refresh', 13)} Refresh</button>
          <button type="button" class="btn sm" onclick="exportActivityLogs('${scope}')">${ic('download', 13)} Export JSON</button>
        </div>
      </div>

      <div class="act-toolbar">
        <div style="position: relative; flex: 1; min-width: 220px;">
          <input type="text" id="act-search-inp" placeholder="Search actions, files, commands, usernames, or IPs..."
            value="${esc(S.actFilter?.q || '')}"
            oninput="setActivityQuery(this.value)"
            style="padding-left: 34px; height: 38px;">
          <span style="position: absolute; left: 10px; top: 10px; color: var(--tx-muted);">${ic('search', 15)}</span>
        </div>
        <select id="act-cat-filter" onchange="setActivityCategory(this.value)" style="width: auto; min-width: 185px; height: 38px;">
          ${ACTIVITY_CATEGORIES.map(([k, lbl]) => `
            <option value="${k}" ${(S.actFilter?.category || 'all') === k ? 'selected' : ''}>Category: ${esc(lbl)}</option>
          `).join('')}
        </select>
      </div>

      <div id="act-stats" class="act-stats" style="margin-bottom: 0;">
        ${buildActivityStatsHtml(scope)}
      </div>
    </div>

    <div class="card tw">
      <table>
        <thead>
          <tr>
            <th style="width: 105px;">Category</th>
            <th>Event &amp; Details</th>
            ${showActor ? '<th style="width: 130px;">Actor</th>' : ''}
            <th style="width: 145px;">IP Address</th>
            <th style="width: 185px;">Timestamp</th>
          </tr>
        </thead>
        <tbody id="act-tbody" data-scope="${esc(scope)}" data-show-actor="${showActor ? '1' : '0'}">
          ${renderActivityRowsHtml(filtered, showActor)}
        </tbody>
      </table>
    </div>
    ${(S.acct.act.length >= curLimit && curLimit < 500) ? `
      <div class="row" style="justify-content: center; margin-top: 14px;">
        <button type="button" class="btn sm" onclick="loadMoreActivity()">Load up to 500 events</button>
      </div>
    ` : ''}
  `;
}

async function loadActivityLog(limit = S.actFilter?.limit || 200, skipRender = false) {
  if (!S.token) return;
  if (!S.loading) S.loading = {};
  S.loading.activity = true;
  if (!skipRender && $('#act-tbody') && (!S.acct.act || S.acct.act.length === 0)) {
    renderActivityRows();
  }
  try {
    const logs = await api(`/api/account/activity?limit=${limit}`);
    S.acct.act = Array.isArray(logs) ? logs : [];
  } catch (e) {} finally {
    S.loading.activity = false;
    if (!skipRender) {
      if ($('#act-tbody')) {
        renderActivityRows();
      } else if ($('#dash-recent-activity')) {
        const box = $('#dash-recent-activity');
        if (box) box.innerHTML = renderDashRecentActivityHtml();
      } else if (S.view === 'act' || (S.view === 'admin' && S.adminTab === 5) || (S.view === 'acc' && S.acct.tab === 4)) {
        R();
      }
    }
  }
}

function setActivityQuery(q) {
  if (!S.actFilter) S.actFilter = { q: '', category: 'all', limit: 200 };
  S.actFilter.q = q;
  renderActivityRows();
}

function setActivityCategory(cat, syncSelect = false) {
  if (!S.actFilter) S.actFilter = { q: '', category: 'all', limit: 200 };
  S.actFilter.category = cat || 'all';
  if (syncSelect) {
    const sel = $('#act-cat-filter');
    if (sel) {
      sel.value = S.actFilter.category;
      if (typeof sel._cselectRebuild === 'function') sel._cselectRebuild();
    }
  }
  renderActivityRows();
}

function loadMoreActivity() {
  if (!S.actFilter) S.actFilter = { q: '', category: 'all', limit: 200 };
  S.actFilter.limit = 500;
  loadActivityLog(500);
}

function exportActivityLogs(scope = 'all') {
  const rows = getFilteredActivity(scope);
  const blob = new Blob([JSON.stringify(rows, null, 2)], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `neopanel-activity-${new Date().toISOString().slice(0, 10)}.json`;
  document.body.appendChild(a);
  a.click();
  a.remove();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
  toast(`Exported ${rows.length} activity event${rows.length === 1 ? '' : 's'}`);
}

function renderBrandPreviewMarkup(b) {
  const brand = { ...DEFAULT_BRANDING, ...(b || {}) };
  const ac = brand.accent_color || '#387aff';
  const bgDeep = brand.bg_deep || '#0a0d14';
  const bgCanvas = brand.bg_canvas || '#101521';
  const cardHex = brand.card_bg || '#121826';
  const navHex = brand.nav_bg || '#0d121d';
  const tx = brand.text_color || '#e6edf3';
  const op = Math.max(0.35, Math.min(1, (Number(brand.card_opacity) || 72) / 100));
  const cardRgb = hexToRgbStr(cardHex) || '18, 24, 38';
  const navRgb = hexToRgbStr(navHex) || '13, 18, 29';
  const rad = Number(brand.border_radius_px) >= 0 ? Number(brand.border_radius_px) : 14;
  const blur = Number(brand.glass_blur_px) >= 0 ? Number(brand.glass_blur_px) : 16;
  const hero = (brand.server_hero_url || SERVER_HERO_URL).trim();

  return `<div class="brand-mini-app" style="background: radial-gradient(circle at top right, ${ac}22, transparent 55%), linear-gradient(145deg, ${bgDeep}, ${bgCanvas}); border-radius:${rad}px; color:${tx};">
    <div class="brand-mini-nav" style="background: rgba(${navRgb}, 0.9);">
      <div class="row" style="gap: 6px; font-weight: 700; font-size: 11px; color: ${tx};">
        ${brand.logo_url ? `<img src="${esc(brand.logo_url)}" style="width:16px;height:16px;border-radius:4px;object-fit:cover;" alt="">` : `<span style="width:14px;height:14px;border-radius:4px;background:${ac};display:inline-block;"></span>`}
        <span style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">${esc(brand.brand_name || 'NeoPanel')}</span>
      </div>
      <div style="height:20px;border-radius:6px;background:${ac}28;border:1px solid ${ac}66;display:flex;align-items:center;padding:0 7px;font-size:10px;font-weight:600;color:${tx};margin-top:4px;">Servers</div>
      <div style="height:18px;border-radius:6px;display:flex;align-items:center;padding:0 7px;font-size:10px;opacity:0.65;">Console</div>
      <div style="height:18px;border-radius:6px;display:flex;align-items:center;padding:0 7px;font-size:10px;opacity:0.65;">Databases</div>
    </div>
    <div class="brand-mini-main">
      <div class="brand-mini-banner" style="background-image:url('${esc(hero)}'); border-radius:${Math.max(4, rad - 4)}px;">
        <div style="display:flex;justify-content:space-between;align-items:center;gap:6px;">
          <b style="font-size:11.5px;color:#fff;">Survival SMP</b>
          <span style="background:${ac};color:#fff;font-size:9.5px;font-weight:700;padding:2px 7px;border-radius:6px;">Start</span>
        </div>
        <div style="font-size:9.5px;color:rgba(255,255,255,0.75);margin-top:3px;">play.example.com:25565</div>
      </div>
      <div style="background: rgba(${cardRgb}, ${op.toFixed(2)}); backdrop-filter: blur(${blur}px); border:1px solid rgba(255,255,255,0.1); border-radius:${Math.max(4, rad - 3)}px; padding:9px 10px;">
        <div style="font-size:10px;opacity:0.75;margin-bottom:4px;">Glass Card Preview (${Math.round(op * 100)}% &bull; ${rad}px &bull; ${blur}px blur)</div>
        <div style="height:6px;border-radius:99px;background:rgba(255,255,255,0.12);overflow:hidden;">
          <div style="width:68%;height:100%;background:${ac};"></div>
        </div>
      </div>
    </div>
  </div>`;
}

// --- Views Implementation (V) ---
const V = {
  team() { return renderTeamPage(); },
  dash() {
    const runningCount = S.servers.filter((s, i) => {
      const st = i === S.cs ? S.status : normalizeServerStatus(s.stats?.status || s.status);
      return st === 'online' || st === 'starting';
    }).length;
    const totalCount = S.servers.length;
    const totalRamAllocatedMB = S.servers.reduce((sum, s) => sum + getServerMaxRamMB(s), 0);
    const totalRamAllocatedGB = (totalRamAllocatedMB / 1024).toFixed(1);
    const totalRamUsedMB = S.servers.reduce((sum, s, i) => {
      const st = i === S.cs ? S.status : normalizeServerStatus(s.stats?.status || s.status);
      if (st !== 'online' && st !== 'starting') return sum;
      return sum + (i === S.cs ? (S.memUsedMB || Number(s.stats?.memory_mb) || 0) : (Number(s.stats?.memory_mb) || 0));
    }, 0);
    const totalRamUsedGB = (totalRamUsedMB / 1024).toFixed(2);
    const totalDiskMB = S.servers.reduce((sum, s, i) => sum + (i === S.cs ? (S.diskMB || Number(s.stats?.disk_mb) || 0) : (Number(s.stats?.disk_mb) || 0)), 0);
    const diskQuotaMB = Math.max(20480, totalCount * 10240);

    const nodesList = Array.isArray(S.nodes) ? S.nodes : [];
    const nodeErrs = S.clusterNodeErrors || {};
    const errEntries = Object.entries(nodeErrs);
    const activeNodeFilter = S.nodeFilter || 'all';

    const indexedServers = S.servers.map((s, origIdx) => ({ s, origIdx }));
    const visibleServers = activeNodeFilter === 'all'
      ? indexedServers
      : indexedServers.filter(({ s }) => (s.node_id || 'local') === activeNodeFilter);
    const coverUrl = getBrand('card_cover_url');
    const dashVm = getViewMode('dash', 'tiles');
    const gridCls = dashVm === 'list' ? 'vm-dash-list' : (dashVm === 'icons' ? 'vm-dash-icons' : '');
    const gridStyle = '';

    return `
    <div class="row sb np-page-heading">
      <div>
        <div class="np-eyebrow">Workspace overview</div>
        <h1>Your servers</h1>
        <span id="dash-running-summary" class="mu">${runningCount} of ${totalCount} servers running</span>
      </div>
      <div class="row" style="gap: 10px; flex-wrap: wrap; align-items: center;">
        <div style="position: relative;">
          <input type="text" id="srv-filter" aria-label="Filter servers" placeholder="Find a server…" oninput="filterServerCards(this.value)" style="width: 220px; padding-left: 32px; font-size: 13px; height: 36px; border-radius: 8px;">
          <span style="position: absolute; left: 10px; top: 9px; color: var(--tx-muted);">${ic('search', 14)}</span>
        </div>
        ${nodesList.length > 1 ? `
          <select class="sel sm" style="width: auto; min-width: 130px; height: 36px;" onchange="setNodeFilter(this.value)">
            <option value="all" ${activeNodeFilter === 'all' ? 'selected' : ''}>All Locations (${totalCount})</option>
            ${nodesList.map(n => {
              const cnt = S.servers.filter(x => (x.node_id || 'local') === n.id).length;
              return `<option value="${esc(n.id)}" ${activeNodeFilter === n.id ? 'selected' : ''}>${esc(displayNodeName(n.id, n.name))} (${cnt})</option>`;
            }).join('')}
          </select>
        ` : ''}
        ${renderViewModeToggle('dash', 'tiles')}
        ${canCreateServers() ? `<button class="btn pr" onclick="openCreateServerModal()" style="height: 36px; padding: 0 16px; font-size: 13px;">
          ${ic('plus', 14)} Create server
        </button>` : ''}
      </div>
    </div>

    ${(errEntries.length > 0 && S.user?.role === 'admin') ? `
      <div class="node-error-note" style="margin-bottom: 16px; display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 10px;">
        <span><b>Node connection issues:</b> ${errEntries.map(([nid, msg]) => `${esc(displayNodeName(nid, getNodeById(nid)?.name))} (${esc(msg)})`).join(' &bull; ')}</span>
        <a href="/admin/nodes" class="btn sm" onclick="return handleNavClick(event, 'admin', { adminTab: 1 })" style="text-decoration: none;">Inspect Nodes</a>
      </div>
    ` : ''}

    <div class="np-slot-container" data-np-slot="dashboard:top" style="margin-bottom: 16px;"></div>

    <div class="dash-grid">
      <!-- Left Column: Server Cards Grid -->
      <div id="srv-card-grid" class="np-server-grid ${gridCls}" style="${gridStyle}">
        ${S.loading?.dash && S.servers.length === 0 ? [1, 2, 3].map(() => `
          <div class="np-skeleton-card">
            <div class="np-skeleton" style="height: 120px; border-radius: 0;"></div>
            <div style="padding: 14px 18px 18px; display: flex; flex-direction: column; gap: 12px;">
              <div class="np-skeleton" style="height: 18px; width: 65%;"></div>
              <div class="np-skeleton" style="height: 12px; width: 45%;"></div>
              <div class="np-skeleton" style="height: 28px; width: 100%; margin-top: 6px;"></div>
            </div>
          </div>
        `).join('') : visibleServers.map(({ s, origIdx }) => {
          const nid = s.node_id || 'local';
          const nodeName = displayNodeName(nid, s.node_name || getNodeById(nid)?.name);
          const isRemote = nid !== 'local';
          const canDelete = S.user?.role === 'admin' || (s.owner_id && S.user?.id && s.owner_id === S.user.id);
          const st = origIdx === S.cs ? S.status : normalizeServerStatus(s.stats?.status || s.status);
          const isRun = st === 'online' || st === 'starting';
          const maxRam = getServerMaxRamMB(s);
          const cVal = isRun ? (origIdx === S.cs ? (S.cpu.at(-1) || 0) : (Number(s.stats?.cpu_usage) || 0)) : 0;
          const mVal = isRun ? (origIdx === S.cs ? (S.memUsedMB || Number(s.stats?.memory_mb) || 0) : (Number(s.stats?.memory_mb) || 0)) : 0;
          const dVal = origIdx === S.cs ? (S.diskMB || Number(s.stats?.disk_mb) || 0) : (Number(s.stats?.disk_mb) || 0);
          const upSec = origIdx === S.cs ? S.uptimeSec : (Number(s.stats?.uptime_sec) || 0);
          const host = `${srvHost(s)}:${s.port || 25565}`;
          const coreLabel = engineMeta(s.type || 'paper').label;
          const engBadgeIcon = engineIcon(s.type || 'paper', 14);
          const srvUrl = serverRouteUrl(nid, s.id, 'con');

          if (dashVm === 'list') {
            const defaultSrvIcon = getBrand('default_server_icon_url') || '/static/img/default-server-icon.png';
            const srvLogoUrl = (s.icon_url && String(s.icon_url).trim()) ? s.icon_url : defaultSrvIcon;
            return `
            <a id="dash-srv-${nid}-${s.id}" href="${esc(srvUrl)}" class="sift-server-card vm-dash-row" onclick="return handleServerCardClick(event, ${origIdx}, 'con')" data-name="${esc(s.name)}" data-addr="${esc(host)}">
              <div class="vm-dash-row-bg" ${coverUrl ? `style="background-image: url('${esc(coverUrl)}');"` : ''} aria-hidden="true"></div>
              <div class="vm-dash-row-main">
                <div class="srv-row-thumb" style="background-image: url('${esc(srvLogoUrl)}'), url('/static/img/default-server-icon.png'), url('/img/default-server-icon.png');"></div>
                <div style="min-width: 0; flex: 1;">
                  <div class="row" style="gap: 10px; align-items: center; flex-wrap: wrap;">
                    <span class="srv-card-title" title="${esc(s.name)}">${renderMinecraftTextHtml(s.name || 'Minecraft Server', { allowNewlines: false })}</span>
                    <span data-role="badge" class="badge ${isRun ? 'online' : 'offline'}" style="white-space: nowrap; flex-shrink: 0;">
                      <i class="dot"></i>${isRun ? 'Running' : 'Offline'}
                    </span>
                  </div>
                  <div class="row" style="gap: 8px; margin-top: 5px; align-items: center; flex-wrap: wrap;">
                    <span class="srv-card-addr copyable-ip" data-addr="${esc(host)}" onclick="return copyServerAddress(this.dataset.addr, event)" title="Click to copy server address"><span class="pv">${esc(host)}</span> ${ic('copy', 12)}</span>
                    <span class="chip" style="font-size: 11px; display:inline-flex; align-items:center; gap:5px;">${engBadgeIcon} ${esc(coreLabel)} ${esc(s.version || '')}</span>
                    <span data-role="uptime" class="srv-card-uptime" style="font-size: 11.5px; color: var(--tx-muted); white-space: nowrap; ${isRun ? '' : 'display:none;'}">&bull; Up ${formatUptime(upSec)}</span>
                    ${isRemote ? `<span class="chip" style="font-size: 11px; padding: 1px 7px;">${esc(nodeName)}</span>` : ''}
                  </div>
                </div>
              </div>
              <div class="vm-dash-row-meters">
                <div class="srv-meter">
                  <div class="srv-meter-head"><span>CPU</span><span data-role="cpu-val" class="srv-meter-val">${cVal.toFixed(1)}%</span></div>
                  <div class="meter-bar"><i data-role="cpu-bar" style="width:${Math.min(100, cVal)}%"></i></div>
                </div>
                <div class="srv-meter">
                  <div class="srv-meter-head"><span>RAM</span><span data-role="mem-val" class="srv-meter-val">${Math.round(mVal)} MB</span></div>
                  <div class="meter-bar"><i data-role="mem-bar" style="width:${Math.min(100, (mVal / (maxRam || 1)) * 100)}%"></i></div>
                </div>
                <div class="srv-meter">
                  <div class="srv-meter-head"><span>Disk</span><span data-role="disk-val" class="srv-meter-val">${formatDiskMB(dVal)}</span></div>
                  <div class="meter-bar"><i data-role="disk-bar" style="width:${Math.min(100, (dVal / 20480) * 100)}%"></i></div>
                </div>
              </div>
              <div class="vm-dash-row-actions">
                <span class="srv-card-action-btn" onclick="event.preventDefault(); event.stopPropagation(); loadDashboard(); toast('Refreshed server stats');" title="Refresh status">&#x21bb;</span>
                ${canDelete ? `<span class="srv-card-action-btn danger" data-srv-id="${esc(s.id)}" data-srv-name="${esc(s.name)}" data-node-id="${esc(nid)}" onclick="event.preventDefault(); event.stopPropagation(); openDeleteServerModal(this.dataset.srvId, this.dataset.srvName, this.dataset.nodeId);" title="Delete Server">${ic('trash', 13)}</span>` : ''}
              </div>
            </a>`;
          }

          return `
          <a id="dash-srv-${nid}-${s.id}" href="${esc(srvUrl)}" class="sift-server-card ${dashVm === 'icons' ? 'vm-dash-icon' : ''}" onclick="return handleServerCardClick(event, ${origIdx}, 'con')" data-name="${esc(s.name)}" data-addr="${esc(host)}">
            <div class="srv-card-banner ${coverUrl ? 'has-brand-art' : ''}" ${coverUrl ? `style="background-image: url('${esc(coverUrl)}');"` : ''}>
              <div class="srv-card-badges-left">
                <span class="srv-card-badge" style="display:inline-flex; align-items:center; gap:6px;">${engBadgeIcon} <span>${esc(coreLabel)}</span></span>
              </div>
              <div class="srv-card-actions">
                <span class="srv-card-action-btn" onclick="event.preventDefault(); event.stopPropagation(); loadDashboard(); toast('Refreshed server stats');" title="Refresh status">&#x21bb;</span>
                ${canDelete ? `<span class="srv-card-action-btn danger" data-srv-id="${esc(s.id)}" data-srv-name="${esc(s.name)}" data-node-id="${esc(nid)}" onclick="event.preventDefault(); event.stopPropagation(); openDeleteServerModal(this.dataset.srvId, this.dataset.srvName, this.dataset.nodeId);" title="Delete Server">${ic('trash', 13)}</span>` : ''}
              </div>
            </div>
            <div class="srv-card-body">
              <div class="srv-card-title-row">
                <span class="srv-card-title" title="${esc(s.name)}">${renderMinecraftTextHtml(s.name || 'Minecraft Server', { allowNewlines: false })}</span>
                <span data-role="badge" class="badge ${isRun ? 'online' : 'offline'}" style="white-space: nowrap; flex-shrink: 0;">
                  <i class="dot"></i>${isRun ? 'Running' : 'Offline'}
                </span>
              </div>
              <div class="srv-card-sub-row">
                <div class="srv-card-addr copyable-ip" data-addr="${esc(host)}" onclick="return copyServerAddress(this.dataset.addr, event)" title="Click to copy server address">
                  <span class="pv">${esc(host)}</span> ${ic('copy', 12)}
                </div>
                <div style="display: flex; align-items: center; gap: 6px; flex-shrink: 0;">
                  ${isRemote ? `<span class="chip" style="font-size: 10px; padding: 1px 6px;">${esc(nodeName)}</span>` : ''}
                  <span data-role="uptime" class="srv-card-uptime" style="font-size: 11.5px; color: var(--tx-muted); white-space: nowrap; ${isRun ? '' : 'display:none;'}">Up ${formatUptime(upSec)}</span>
                </div>
              </div>
              <div class="srv-meters-row">
                <div class="srv-meter">
                  <div class="srv-meter-head">
                    <span>CPU</span>
                    <span data-role="cpu-val" class="srv-meter-val">${cVal.toFixed(1)}%</span>
                  </div>
                  <div class="meter-bar"><i data-role="cpu-bar" style="width:${Math.min(100, cVal)}%"></i></div>
                </div>
                <div class="srv-meter">
                  <div class="srv-meter-head">
                    <span>RAM</span>
                    <span data-role="mem-val" class="srv-meter-val">${Math.round(mVal)} MB</span>
                  </div>
                  <div class="meter-bar"><i data-role="mem-bar" style="width:${Math.min(100, (mVal / (maxRam || 1)) * 100)}%"></i></div>
                </div>
                <div class="srv-meter">
                  <div class="srv-meter-head">
                    <span>Disk</span>
                    <span data-role="disk-val" class="srv-meter-val">${formatDiskMB(dVal)}</span>
                  </div>
                  <div class="meter-bar"><i data-role="disk-bar" style="width:${Math.min(100, (dVal / 20480) * 100)}%"></i></div>
                </div>
              </div>
            </div>
          </a>`;
        }).join('')}

        <!-- Dashed Create Server Card -->
        ${canCreateServers() ? `<button type="button" class="create-server-dashed-card ${dashVm === 'list' ? 'compact' : ''}" onclick="openCreateServerModal()">
          <span class="create-icon-bubble">${ic('plus', 18)}</span>
          <span>Create a server</span>
          <span class="np-create-hint">Make room for your next world</span>
        </button>` : ''}
      </div>

      <!-- Right Column: Recent Activity & Allocation -->
      <div class="np-dash-aside">
        <!-- Recent Activity Card -->
        <div class="card np-recent-activity">
          <div class="row sb" style="margin-bottom: 14px;">
            <h2 style="font-size: 15px; font-weight: 700; color: var(--tx); margin: 0;">Recent activity</h2>
            <a href="/activity" onclick="return handleNavClick(event, 'act', { global: true })" class="mu" style="font-size: 12px; text-decoration: none;">View all</a>
          </div>
          <div id="dash-recent-activity">
            ${renderDashRecentActivityHtml()}
          </div>
        </div>

        <!-- Your Allocation Card -->
        <div class="card np-allocation">
          <h2>Resource overview</h2>
          <div class="np-allocation-metrics">
            <div>
              <div class="row sb" style="font-size: 12px; margin-bottom: 5px;">
                <span class="mu">Memory</span>
                <span id="dash-alloc-mem-val" class="mono" style="font-weight: 600; color: var(--tx);">${totalRamUsedGB} of ${totalRamAllocatedGB} GB</span>
              </div>
              <div class="meter-bar"><i id="dash-alloc-mem-bar" style="width: ${Math.min(100, (totalRamUsedMB / (totalRamAllocatedMB || 1)) * 100)}%;"></i></div>
            </div>
            <div>
              <div class="row sb" style="font-size: 12px; margin-bottom: 5px;">
                <span class="mu">Disk</span>
                <span id="dash-alloc-disk-val" class="mono" style="font-weight: 600; color: var(--tx);">${formatDiskMB(totalDiskMB)} used</span>
              </div>
              <div class="meter-bar"><i id="dash-alloc-disk-bar" style="width: ${Math.min(100, (totalDiskMB / diskQuotaMB) * 100)}%;"></i></div>
            </div>
            <div>
              <div class="row sb" style="font-size: 12px; margin-bottom: 5px;">
                <span class="mu">Running</span>
                <span id="dash-alloc-run-val" class="mono" style="font-weight: 600; color: var(--tx);">${runningCount} of ${totalCount}</span>
              </div>
              <div class="meter-bar"><i id="dash-alloc-run-bar" style="width: ${Math.min(100, (runningCount / (totalCount || 1)) * 100)}%;"></i></div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="np-slot-container" data-np-slot="dashboard:bottom" style="margin-top: 16px;"></div>`;
  },

  con() {
    return `${srvh()}
    <div class="np-section-heading"><div><h2>Console</h2><p>Live output and server performance.</p></div><span class="np-shortcut-hint">Command history <kbd>↑</kbd> <kbd>↓</kbd></span></div>
    <div class="cgrid">
      <div class="cwrap">
        <div class="con-top-bar">
          <div class="row" style="gap:8px; align-items:center;">
            ${ic('terminal', 14)}
            <span class="np-console-label">Live console</span>
            <span id="ws-live-pill" class="con-live-pill"><i></i>Connecting...</span>
            <span class="mu">&bull;</span>
            <span id="con-buf-count">${S.log.length} lines</span>
          </div>
          <div class="row" style="gap:6px; align-items:center; flex-wrap:wrap;">
            <button type="button" class="con-quick-pill" onclick="runCmd('save-all')" title="Save all worlds to disk (save-all)">${ic('download', 11)} save-all</button>
            <button type="button" class="con-quick-pill" onclick="runCmd('list')" title="List online players (list)">${ic('users', 11)} list</button>
            <button type="button" class="con-quick-pill" onclick="runCmd('tps')" title="Check tick rate &amp; server health (tps)">${ic('activity', 11)} tps</button>
            <button type="button" class="btn sm" onclick="S.log=[]; paintLog(); toast('Console buffer cleared');" style="font-size:11px; padding:3px 9px;">Clear</button>
          </div>
        </div>
        <div class="con" id="log" role="log" aria-live="polite"></div>
        <div class="cin">
          <span>&raquo;</span>
          <input id="cmd" type="text" class="con-cmd-input" autocomplete="off" spellcheck="false" placeholder="Type a command..." aria-label="Console command">
        </div>
      </div>
      <div class="card stc" id="stc">${statsHtml()}</div>
    </div>
    <div class="np-slot-container" data-np-slot="console:below-stats"></div>
    <div class="grid g3" id="chr" style="margin-top: 24px;">${chartsHtml()}</div>
    <div class="np-slot-container" data-np-slot="console:below-charts" style="margin-top: 16px;"></div>`;
  },

  pl() {
    const t = S.ptab;
    const vm = getViewMode('pl', 'list');
    const maxPl = Number(S.playersMax || S.props?.max || 20) || 20;
    const onlineCnt = Math.max(S.players.length, Number(S.playersOnlineCount) || 0);
    let body = '';

    if (t === 'online') {
      if (S.loading?.players) {
        body = `<div class="card">${loaderHtml('Loading player lists...')}</div>`;
      } else if (!S.players.length) {
        const slpHint = onlineCnt > 0
          ? `<div style="font-weight:600; color:var(--tx); margin-bottom:6px;">${onlineCnt} player${onlineCnt === 1 ? '' : 's'} connected (SLP reported ${onlineCnt}/${maxPl})</div><div>Click <b>Sync (/list)</b> above to query exact online player names from the server console.</div>`
          : 'No players currently online. Connected players appear automatically from server logs and SLP pings.';
        body = `<div class="card mu" style="text-align:center;padding:26px;">${slpHint}</div>`;
      } else if (vm === 'icons' || vm === 'tiles') {
        body = `<div class="${vm === 'icons' ? 'vm-icons-grid' : 'vm-tiles-grid'}">
          ${S.players.map(p => {
            const pName = p.n || p.name || 'Player';
            const pingLabel = p.ping ? `${p.ping} ms` : 'Online';
            return `
            <div class="${vm === 'icons' ? 'vm-icon-card' : 'vm-tile-card'}">
              <div class="${vm === 'icons' ? 'vm-plug-icon-center' : 'vm-item-head'}">
                <div class="row" style="gap: 8px; min-width: 0; ${vm === 'icons' ? 'flex-direction: column;' : ''}">
                  ${head(pName)}
                  <b class="vm-item-title">${esc(pName)}</b>
                </div>
                ${p.op ? '<span class="badge online">OP</span>' : `<span class="chip">${esc(pingLabel)}</span>`}
              </div>
              <div class="mu mono" style="font-size: 11.5px;">${esc(p.t || 'Online')} &bull; <span class="pv">${esc(p.ip || '—')}</span></div>
              <div class="vm-item-foot">
                <button class="btn sm" onclick="playerAction('${p.op ? 'deop' : 'op'}', '${esc(pName)}')">${p.op ? 'Deop' : 'Op'}</button>
                <div class="row" style="gap: 6px;">
                  <button class="btn sm" onclick="playerAction('kick', '${esc(pName)}')">Kick</button>
                  <button class="btn sm dg" onclick="playerAction('ban', '${esc(pName)}')">Ban</button>
                </div>
              </div>
            </div>`;
          }).join('')}
        </div>`;
      } else {
        body = `<div class="card tw"><table>
          <thead><tr><th>Player</th><th>Ping</th><th>Playtime</th><th>IP</th><th style="text-align:right">Actions</th></tr></thead>
          <tbody>
            ${S.players.map(p => {
              const pName = p.n || p.name || 'Player';
              return `
              <tr>
                <td>${head(pName)}<b>${esc(pName)}</b> ${p.op ? '<span class="badge" style="margin-left:6px;">Operator</span>' : ''}</td>
                <td class="mono">${p.ping ? `${p.ping} ms` : '—'}</td>
                <td>${esc(p.t || 'Online')}</td>
                <td class="mono pv">${esc(p.ip || '—')}</td>
                <td style="text-align:right">
                  <div class="row" style="justify-content:flex-end">
                    <button class="btn sm" onclick="playerAction('${p.op ? 'deop' : 'op'}', '${esc(pName)}')">${p.op ? 'Deop' : 'Op'}</button>
                    <button class="btn sm" onclick="playerAction('kick', '${esc(pName)}')">Kick</button>
                    <button class="btn sm dg" onclick="playerAction('ban', '${esc(pName)}')">Ban</button>
                  </div>
                </td>
              </tr>`;
            }).join('')}
          </tbody>
        </table></div>`;
      }
    } else if (t === 'wl') {
      const wlBar = `<div class="card" style="padding: 14px 18px; margin-bottom: 12px;">
        <div class="row">
          <input type="text" id="wn" placeholder="Add Minecraft username to whitelist" style="max-width:280px">
          <button class="btn pr" onclick="addWhitelistPlayer()">Add to whitelist</button>
        </div>
      </div>`;
      if (!S.wl.length) {
        body = `${wlBar}<div class="card mu" style="text-align:center;padding:24px;">No players on whitelist.</div>`;
      } else if (vm === 'icons' || vm === 'tiles') {
        body = `${wlBar}<div class="${vm === 'icons' ? 'vm-icons-grid' : 'vm-tiles-grid'}">
          ${S.wl.map(p => {
            const name = typeof p === 'string' ? p : p.name;
            const uuid = typeof p === 'object' ? p.uuid : '';
            return `<div class="${vm === 'icons' ? 'vm-icon-card' : 'vm-tile-card'}">
              <div class="${vm === 'icons' ? 'vm-plug-icon-center' : 'row'}" style="gap: 10px;">
                ${head(name)}
                <b class="vm-item-title">${esc(name)}</b>
              </div>
              <div class="mu mono" style="font-size: 11px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;" title="${esc(uuid || '')}">${esc(uuid || '—')}</div>
              <div class="vm-item-foot">
                <span class="chip">Whitelisted</span>
                <button class="btn sm dg" onclick="playerAction('whitelist_remove', '${esc(name)}')">Remove</button>
              </div>
            </div>`;
          }).join('')}
        </div>`;
      } else {
        body = `<div class="card tw"><div class="row" style="margin-bottom:14px">
          <input type="text" id="wn" placeholder="Add Minecraft username to whitelist" style="max-width:280px">
          <button class="btn pr" onclick="addWhitelistPlayer()">Add to whitelist</button>
        </div>
        <table>
          <thead><tr><th>Player</th><th>UUID</th><th style="text-align:right">Actions</th></tr></thead>
          <tbody>
            ${S.wl.map(p => {
              const name = typeof p === 'string' ? p : p.name;
              const uuid = typeof p === 'object' ? p.uuid : '';
              return `<tr>
                <td>${head(name)}<b>${esc(name)}</b></td>
                <td class="mono mu">${uuid || '—'}</td>
                <td style="text-align:right">
                  <button class="btn sm dg" onclick="playerAction('whitelist_remove', '${esc(name)}')">Remove</button>
                </td>
              </tr>`;
            }).join('')}
          </tbody>
        </table></div>`;
      }
    } else if (t === 'ban') {
      if (!S.bans.length) {
        body = `<div class="card mu" style="text-align:center;padding:24px;">No banned players on record.</div>`;
      } else if (vm === 'icons' || vm === 'tiles') {
        body = `<div class="${vm === 'icons' ? 'vm-icons-grid' : 'vm-tiles-grid'}">
          ${S.bans.map(b => `
            <div class="${vm === 'icons' ? 'vm-icon-card' : 'vm-tile-card'}">
              <div class="${vm === 'icons' ? 'vm-plug-icon-center' : 'row'}" style="gap: 10px;">
                ${head(b.name)}
                <b class="vm-item-title">${esc(b.name)}</b>
              </div>
              <div class="vm-item-sub">${esc(b.reason || 'Banned by operator')}</div>
              <div class="vm-item-foot">
                <span class="badge offline">Banned</span>
                <button class="btn sm" onclick="playerAction('pardon', '${esc(b.name)}')">Pardon</button>
              </div>
            </div>
          `).join('')}
        </div>`;
      } else {
        body = `<div class="card tw"><table>
          <thead><tr><th>Player</th><th>Reason</th><th style="text-align:right">Actions</th></tr></thead>
          <tbody>
            ${S.bans.map(b => `
              <tr>
                <td>${head(b.name)}<b>${esc(b.name)}</b></td>
                <td class="mu">${esc(b.reason || 'Banned by operator')}</td>
                <td style="text-align:right">
                  <button class="btn sm" onclick="playerAction('pardon', '${esc(b.name)}')">Pardon</button>
                </td>
              </tr>
            `).join('')}
          </tbody>
        </table></div>`;
      }
    }

    return `${srvh()}
    <div class="row sb" style="margin-top:14px; margin-bottom:14px; flex-wrap:wrap; gap:10px;">
      <div class="row" style="gap:10px; flex-wrap:wrap; align-items:center;">
        <div class="tabs" style="margin:0;">
          ${[['online', `Online Players (${onlineCnt})`], ['wl', `Whitelist (${S.wl.length})`], ['ban', `Banned Players (${S.bans.length})`]].map(([k, l]) => `
            <button class="btn ${t === k ? 'on' : ''}" onclick="S.ptab='${k}'; R(); loadPlayersData();">${l}</button>
          `).join('')}
        </div>
        <span class="badge ${onlineCnt > 0 ? 'online' : 'offline'}"><i class="dot"></i>${onlineCnt} / ${maxPl} Online</span>
      </div>
      <div class="row" style="gap:8px; flex-wrap:wrap;">
        <button type="button" class="btn sm" onclick="syncPlayersList(this)" title="Query live /list from console and refresh SLP">${ic('refresh', 13)} Sync (/list)</button>
        ${renderViewModeToggle('pl', 'list')}
      </div>
    </div>
    ${body}`;
  },

  fi() {
    const preview = typeof renderFilePreview === 'function' ? renderFilePreview() : '';
    if (preview) return `${srvh()}${preview}`;
    const f = S.editingFile || S.ftab;
    if (S.editingFile) {
      return `${srvh()}
      <div class="np-editor-workspace">
        <div class="row sb np-editor-toolbar">
          <div class="row np-editor-file">
            <button class="btn sm" onclick="closeEditor()">&larr; Back to files</button>
            <span class="mono np-editor-name">${esc(S.editingFile)}</span>
            <span id="editor-save-status" class="mu" role="status" style="font-size:12px;"></span>
          </div>
          <div class="row" style="gap: 8px;">
            <button class="btn sm" onclick="downloadFileItem('${esc(S.editingFile)}')">${ic('download', 14)} Download</button>
            <button id="editor-save-btn" class="btn sm pr" onclick="saveActiveFile()" title="Ctrl/Cmd + S">${ic('download', 14)} Save File</button>
          </div>
        </div>
        <textarea id="ed" spellcheck="false" aria-label="File editor">${esc(S.files[S.editingFile] || '')}</textarea>
      </div>`;
    }

    const pathParts = (S.currentDir || '').split('/').filter(Boolean);
    const isSubdir = pathParts.length > 0;
    const sortedFiles = sortFileItems(S.fileList, S.fileSort);
    S.sortedFileList = sortedFiles;
    const sortBy = S.fileSort?.by || 'name';
    const sortDir = S.fileSort?.dir === 'desc' ? 'desc' : 'asc';
    const sortArrow = col => sortBy === col ? `<span class="file-sort-arrow">${sortDir === 'asc' ? '&#9650;' : '&#9660;'}</span>` : '';
    const allSelected = sortedFiles.length > 0 && sortedFiles.every(i => S.selectedFiles.has(i.path || i.name));
    const vm = getViewMode('files', 'list');

    const renderFileCtxMenuHtml = (item, idx, isDir, itemPath) => `
      <div class="file-ctx-menu" id="fctx-${idx}" style="display: none;">
        ${!isDir ? `<button data-name="${esc(item.name)}" onclick="openFile(this.dataset.name); closeAllCtx();">${ic('eye', 14)} Open / Preview</button>` : ''}
        ${!isDir ? `<button data-path="${esc(itemPath)}" onclick="downloadFileItem(this.dataset.path); closeAllCtx();">${ic('download', 14)} Download</button>` : ''}
        ${isArchiveFile(item.name) ? `<button data-path="${esc(itemPath)}" onclick="unzipFileItem(this.dataset.path); closeAllCtx();">${ic('package', 14)} Extract</button>` : ''}
        <button data-path="${esc(itemPath)}" onclick="openMoveFilesModal([this.dataset.path]); closeAllCtx();">${ic('move', 14)} Move</button>
        <button data-path="${esc(itemPath)}" onclick="openCompressFilesModal([this.dataset.path]); closeAllCtx();">${ic('archive', 14)} Archive / Zip</button>
        <button data-path="${esc(itemPath)}" onclick="promptRenameFile(this.dataset.path); closeAllCtx();">${ic('settings', 14)} Rename</button>
        <button data-path="${esc(itemPath)}" onclick="copyFilePath(this.dataset.path); closeAllCtx();">${ic('copy', 14)} Copy Path</button>
        <button class="dg" data-path="${esc(itemPath)}" onclick="deleteFileItem(this.dataset.path); closeAllCtx();">${ic('trash', 14)} Delete</button>
      </div>`;

    let filesContainerHtml = '';
    if (vm === 'list') {
      filesContainerHtml = `
      <div class="card tw np-file-table-card">
        <table class="file-table" style="width: 100%;">
          <thead>
            <tr style="border-bottom: 1px solid var(--ln); color: var(--tx-muted); font-size: 12px; text-transform: uppercase; letter-spacing: 0.5px;">
              <th class="file-select-cell">
                <input type="checkbox" id="chk-all-files" onchange="toggleSelectAllFiles(this.checked)" ${allSelected ? 'checked' : ''} aria-label="Select all files">
              </th>
              <th class="file-icon-cell" aria-label="File type"></th>
              <th class="file-name-cell">
                <button type="button" class="file-th-btn ${sortBy === 'name' ? 'active' : ''}" onclick="setFileSort('name')">Name ${sortArrow('name')}</button>
              </th>
              <th style="width: 90px; padding: 12px 10px; text-align: left;">
                <button type="button" class="file-th-btn ${sortBy === 'ext' ? 'active' : ''}" onclick="setFileSort('ext')">Ext ${sortArrow('ext')}</button>
              </th>
              <th style="width: 130px; padding: 12px 10px; text-align: left;">
                <button type="button" class="file-th-btn ${sortBy === 'size' ? 'active' : ''}" onclick="setFileSort('size')">Size ${sortArrow('size')}</button>
              </th>
              <th style="width: 170px; padding: 12px 10px; text-align: left;">
                <button type="button" class="file-th-btn ${sortBy === 'modified' ? 'active' : ''}" onclick="setFileSort('modified')">Modified ${sortArrow('modified')}</button>
              </th>
              <th style="width: 50px; padding: 12px 18px 12px 10px; text-align: right;"></th>
            </tr>
          </thead>
          <tbody>
            ${isSubdir ? `
              <tr onclick="browseFolder('..')">
                <td class="file-select-cell"></td>
                <td class="file-icon-cell"><span class="fi-badge up" title="Parent folder">${ic('folderUp', 15)}</span></td>
                <td class="file-name-cell" colspan="5" style="font-weight: 600; color: var(--tx);">.. (Parent folder)</td>
              </tr>
            ` : ''}
            ${sortedFiles.map((item, idx) => {
              const isDir = item.is_dir;
              const icon = fileListIconHtml(item);
              const extLabel = isDir ? 'Folder' : (getFileExtension(item.name, false) ? `.${getFileExtension(item.name, false)}` : '—');
              const sizeStr = formatFileSize(item.size, isDir);
              const dateStr = formatFileDate(item.modified);
              const itemPath = item.path || item.name;
              const isSelected = S.selectedFiles.has(itemPath);

              return `
              <tr class="file-row ${isSelected ? 'row-selected' : ''}" data-name="${esc(item.name)}" data-path="${esc(itemPath)}"
                  onclick="${isDir ? `browseFolder('${esc(item.name)}')` : `openFile('${esc(item.name)}')`}"
                  oncontextmenu="handleFileContextMenu(event, ${idx}); return false;">
                <td class="file-select-cell" onclick="event.stopPropagation();">
                  <input type="checkbox" class="file-chk" ${isSelected ? 'checked' : ''}
                         onchange="toggleFileSelect('${esc(itemPath)}', this.checked)"
                         aria-label="Select ${esc(item.name)}">
                </td>
                <td class="file-icon-cell">${icon}</td>
                <td class="file-name-cell" style="font-weight: 500; color: var(--tx);"><button type="button" class="np-file-open" onclick="event.stopPropagation(); ${isDir ? `browseFolder('${esc(item.name)}')` : `openFile('${esc(item.name)}')`}">${esc(item.name)}</button></td>
                <td class="mu mono" style="font-size: 12px; width: 90px;">${esc(extLabel)}</td>
                <td class="mu mono" style="font-size: 12.5px; width: 130px;">${sizeStr}</td>
                <td class="mu mono" style="font-size: 12.5px; width: 170px;">${dateStr}</td>
                <td style="width: 50px; text-align: right; padding-right: 18px; position: relative;" onclick="event.stopPropagation();">
                  <button class="btn sm" aria-label="File actions for ${esc(item.name)}" onclick="toggleFileCtx(event, ${idx})" style="padding: 4px 8px; border: 0; background: transparent; font-size: 18px; line-height: 1; color: var(--tx-muted);">&#8942;</button>
                  ${renderFileCtxMenuHtml(item, idx, isDir, itemPath)}
                </td>
              </tr>`;
            }).join('')}
            ${S.loading?.files ? `<tr><td colspan="7">${loaderHtml('Loading directory contents...')}</td></tr>` : !sortedFiles.length && !isSubdir ? '<tr><td colspan="7" class="mu" style="text-align:center; padding: 32px;">This folder is empty. Drag and drop files here or click Upload.</td></tr>' : ''}
          </tbody>
        </table>
      </div>`;
    } else if (vm === 'icons') {
      filesContainerHtml = `
      <div class="card" style="padding: 16px;">
        ${sortedFiles.length > 0 ? `
          <div class="row sb" style="margin-bottom: 12px; padding-bottom: 10px; border-bottom: 1px solid var(--border);">
            <label class="row" style="gap: 8px; cursor: pointer; font-size: 12.5px; color: var(--tx-muted);">
              <input type="checkbox" id="chk-all-files" onchange="toggleSelectAllFiles(this.checked)" ${allSelected ? 'checked' : ''} aria-label="Select all files">
              <span>Select all (${sortedFiles.length})</span>
            </label>
            <span class="mu" style="font-size: 12px;">Right-click or use &vellip; for options</span>
          </div>` : ''}
        ${S.loading?.files ? loaderHtml('Loading directory contents...') : (!sortedFiles.length && !isSubdir)
          ? '<div class="mu" style="text-align:center; padding: 32px;">This folder is empty. Drag and drop files here or click Upload.</div>'
          : `<div class="vm-icons-grid">
            ${isSubdir ? `
              <div class="vm-icon-card vm-file-card" onclick="browseFolder('..')">
                <div class="vm-file-head"><span></span><span></span></div>
                <div class="vm-file-center">
                  <span class="fi-badge up lg" title="Parent folder">${ic('folderUp', 24)}</span>
                  <div class="vm-file-name">.. (Parent folder)</div>
                  <div class="vm-file-meta">Go up</div>
                </div>
              </div>` : ''}
            ${sortedFiles.map((item, idx) => {
              const isDir = item.is_dir;
              const sizeStr = isDir ? 'Folder' : formatFileSize(item.size, false);
              const itemPath = item.path || item.name;
              const isSelected = S.selectedFiles.has(itemPath);
              return `
              <div class="vm-icon-card vm-file-card file-row ${isSelected ? 'selected' : ''}" data-name="${esc(item.name)}" data-path="${esc(itemPath)}"
                   onclick="${isDir ? `browseFolder('${esc(item.name)}')` : `openFile('${esc(item.name)}')`}"
                   oncontextmenu="handleFileContextMenu(event, ${idx}); return false;">
                <div class="vm-file-head" onclick="event.stopPropagation();">
                  <input type="checkbox" class="file-chk" ${isSelected ? 'checked' : ''}
                         onchange="toggleFileSelect('${esc(itemPath)}', this.checked)"
                         aria-label="Select ${esc(item.name)}">
                  <div style="position: relative;">
                    <button class="btn sm" onclick="toggleFileCtx(event, ${idx})" style="padding: 2px 6px; border: 0; background: transparent; font-size: 16px; line-height: 1; color: var(--tx-muted);">&#8942;</button>
                    ${renderFileCtxMenuHtml(item, idx, isDir, itemPath)}
                  </div>
                </div>
                <div class="vm-file-center">
                  ${fileListIconHtml(item, 'lg', 24)}
                  <div class="vm-file-name" title="${esc(item.name)}">${esc(item.name)}</div>
                  <div class="vm-file-meta">${sizeStr}</div>
                </div>
              </div>`;
            }).join('')}
          </div>`}
      </div>`;
    } else {
      filesContainerHtml = `
      <div class="card" style="padding: 16px;">
        ${sortedFiles.length > 0 ? `
          <div class="row sb" style="margin-bottom: 12px; padding-bottom: 10px; border-bottom: 1px solid var(--border);">
            <label class="row" style="gap: 8px; cursor: pointer; font-size: 12.5px; color: var(--tx-muted);">
              <input type="checkbox" id="chk-all-files" onchange="toggleSelectAllFiles(this.checked)" ${allSelected ? 'checked' : ''} aria-label="Select all files">
              <span>Select all (${sortedFiles.length})</span>
            </label>
            <span class="mu" style="font-size: 12px;">Right-click or use &vellip; for options</span>
          </div>` : ''}
        ${S.loading?.files ? loaderHtml('Loading directory contents...') : (!sortedFiles.length && !isSubdir)
          ? '<div class="mu" style="text-align:center; padding: 32px;">This folder is empty. Drag and drop files here or click Upload.</div>'
          : `<div class="vm-tiles-grid">
            ${isSubdir ? `
              <div class="vm-tile-card vm-file-card" onclick="browseFolder('..')">
                <div class="vm-tile-file-row">
                  <span class="fi-badge up md" title="Parent folder">${ic('folderUp', 18)}</span>
                  <div class="vm-tile-file-info">
                    <div class="vm-file-name">.. (Parent folder)</div>
                    <div class="vm-file-meta">Navigate up one level</div>
                  </div>
                </div>
              </div>` : ''}
            ${sortedFiles.map((item, idx) => {
              const isDir = item.is_dir;
              const sizeStr = isDir ? 'Folder' : formatFileSize(item.size, false);
              const dateStr = formatFileDate(item.modified);
              const itemPath = item.path || item.name;
              const isSelected = S.selectedFiles.has(itemPath);
              return `
              <div class="vm-tile-card vm-file-card file-row ${isSelected ? 'selected' : ''}" data-name="${esc(item.name)}" data-path="${esc(itemPath)}"
                   onclick="${isDir ? `browseFolder('${esc(item.name)}')` : `openFile('${esc(item.name)}')`}"
                   oncontextmenu="handleFileContextMenu(event, ${idx}); return false;">
                <div class="vm-tile-file-row">
                  <div onclick="event.stopPropagation();">
                    <input type="checkbox" class="file-chk" ${isSelected ? 'checked' : ''}
                           onchange="toggleFileSelect('${esc(itemPath)}', this.checked)"
                           aria-label="Select ${esc(item.name)}">
                  </div>
                  ${fileListIconHtml(item, 'md', 18)}
                  <div class="vm-tile-file-info">
                    <div class="vm-file-name" title="${esc(item.name)}">${esc(item.name)}</div>
                    <div class="vm-file-meta">${sizeStr} &bull; ${dateStr}</div>
                  </div>
                  <div style="position: relative;" onclick="event.stopPropagation();">
                    <button class="btn sm" onclick="toggleFileCtx(event, ${idx})" style="padding: 4px 8px; border: 0; background: transparent; font-size: 16px; line-height: 1; color: var(--tx-muted);">&#8942;</button>
                    ${renderFileCtxMenuHtml(item, idx, isDir, itemPath)}
                  </div>
                </div>
              </div>`;
            }).join('')}
          </div>`}
      </div>`;
    }

    return `${srvh()}
    <div class="np-files-workspace">
      <div class="np-slot-container" data-np-slot="files:top" style="margin-bottom: 12px;"></div>
      <div class="np-section-heading"><div><h2>Files</h2><p>Manage your server’s files and configuration.</p></div></div>
      <div class="row sb np-files-toolbar">
        <!-- Breadcrumb Navigation -->
        <div class="breadcrumb" style="font-size: 13.5px; font-weight: 500; display: flex; align-items: center; flex-wrap: wrap; gap: 4px;">
          <span onclick="browseFolder('')" style="color: var(--tx-muted); cursor: pointer; display: inline-flex; align-items: center; gap: 6px;">${ic('folder', 15)} / home / container /</span>
          ${pathParts.map((part, idx) => {
            const sub = pathParts.slice(0, idx + 1).join('/');
            const isLast = idx === pathParts.length - 1;
            return `<span onclick="browseFolder('${sub}')" style="cursor: pointer; color: ${isLast ? 'var(--tx)' : 'var(--tx-muted)'};">${esc(part)} /</span>`;
          }).join('')}
        </div>

        <!-- Search, Sort, View Mode & Action Buttons -->
        <div class="row" style="gap: 10px; flex-wrap: wrap;">
          <div style="position: relative;">
            <input type="text" id="fi-filter" aria-label="Filter this folder" placeholder="Filter this folder" oninput="filterFileList(this.value)" style="width: 190px; height: 36px; padding: 0 28px 0 12px; font-size: 13px; border-radius: 8px;">
            <kbd style="position: absolute; right: 10px; top: 9px; font-size: 11px; color: var(--tx-muted); background: rgba(255,255,255,0.06); padding: 1px 5px; border-radius: 4px;">/</kbd>
          </div>
          <div class="file-sort-bar">
            <select id="fi-sort-by" style="width: auto; min-width: 132px; height: 36px; font-size: 12.5px;" onchange="setFileSort(this.value, S.fileSort?.dir || 'asc')">
              <option value="name" ${sortBy === 'name' ? 'selected' : ''}>Sort: Name</option>
              <option value="ext" ${sortBy === 'ext' ? 'selected' : ''}>Sort: Extension</option>
              <option value="size" ${sortBy === 'size' ? 'selected' : ''}>Sort: Size</option>
              <option value="modified" ${sortBy === 'modified' ? 'selected' : ''}>Sort: Modified</option>
            </select>
            <button type="button" class="btn sm" onclick="setFileSort(S.fileSort?.by || 'name', S.fileSort?.dir === 'desc' ? 'asc' : 'desc')" style="height: 36px; padding: 0 11px;" title="Toggle ascending / descending order (folders always stay first)">${sortDir === 'asc' ? '&uarr; Asc' : '&darr; Desc'}</button>
          </div>
          ${renderViewModeToggle('files', 'list')}
          <button class="btn" onclick="promptNewFolder()" style="height: 36px;">${ic('folder', 14)} New folder</button>
          <button class="btn" onclick="uploadFileDialog()" style="height: 36px;">${ic('upload', 14)} Upload</button>
          <button class="btn pr" onclick="promptNewFile()" style="height: 36px;">${ic('plus', 14)} New file</button>
        </div>
      </div>

      ${filesContainerHtml}

      <!-- Floating Bulk Action Bar -->
      <div id="file-bulk-bar-root">${renderFileBulkBarHtml()}</div>
    </div>`;
  },

  plug() {
    const srv = curServer();
    if (isSteelEngine(srv?.type)) return `${srvh()}${steelConfigNotice()}`;
    const engSlug = String(srv?.type || S.core || 'paper').toLowerCase().trim();
    const isMod = isModdedEngine(engSlug);
    const nounSing = isMod ? 'Mod' : 'Plugin';
    const nounPlur = isMod ? 'Mods' : 'Plugins';
    const nounLow = isMod ? 'mod' : 'plugin';
    const nounsLow = isMod ? 'mods' : 'plugins';
    const isBrowse = (S.plugTab || 'browse') === 'browse';
    const targetFolder = isMod ? 'mods/' : 'plugins/';
    const instFilter = (S.plugFilter || '').toLowerCase();
    const filteredInstalled = S.plugins
      .map((p, i) => ({ p, i }))
      .filter(({ p }) => !instFilter || String(p[0] || '').toLowerCase().includes(instFilter));
    const storeVm = getViewMode('plug_store', 'list');
    const instVm = getViewMode('plug_inst', 'list');
    const storeVersions = getAllMcVersions(S.plugLoader || (isMod ? engSlug : ''));
    const verChoices = (S.plugVer && !storeVersions.includes(S.plugVer)) ? [S.plugVer, ...storeVersions] : storeVersions;

    const storeGridClass = storeVm === 'icons' ? 'vm-icons-grid wide' : (storeVm === 'tiles' ? 'vm-tiles-grid' : '');
    const storeGridStyle = storeVm === 'list' ? 'display: flex; flex-direction: column; gap: 10px;' : '';
    const storeBodyHtml = Array.isArray(S.lastModrinthHits)
      ? renderModrinthHitsHtml(S.lastModrinthHits)
      : loaderHtml(`Loading top ${nounsLow} from Modrinth...`);

    let installedBodyHtml = '';
    if (S.loading?.plugins) {
      installedBodyHtml = `<div class="card">${loaderHtml(`Loading installed ${nounsLow}...`)}</div>`;
    } else if (!filteredInstalled.length) {
      installedBodyHtml = `<div class="card" style="padding: 32px; text-align: center;">
        <div class="mu" style="margin-bottom: 12px;">${S.plugins.length ? `No installed ${nounsLow} match your filter.` : `No ${nounsLow} installed yet in ${targetFolder}.`}</div>
        ${!S.plugins.length ? `<button class="btn pr sm" onclick="setPlugTab('browse')">${ic('store', 14)} Browse Modrinth ${nounPlur}</button>` : ''}
      </div>`;
    } else if (instVm === 'icons') {
      installedBodyHtml = `<div class="vm-icons-grid wide">
        ${filteredInstalled.map(({ p, i }) => `
          <div class="vm-icon-card">
            <div class="vm-plug-icon-center">
              <div style="width: 48px; height: 48px; border-radius: 12px; background: rgba(56, 122, 255, 0.12); display: grid; place-items: center; color: var(--ac); font-size: 22px;">🧩</div>
              <b class="plugin-title" style="font-size: 13.5px; word-break: break-word;" title="${esc(p[0])}">${esc(p[0])}</b>
              <div class="mu mono" style="font-size: 11.5px;">${esc(p[1] || 'installed')}${p[3] ? ` &bull; ${esc(formatFileDate(p[3]))}` : ''}</div>
            </div>
            <div class="vm-item-foot">
              <label class="sw" title="Toggle ${nounLow} state">
                <input type="checkbox" ${p[2] ? 'checked' : ''} ${p[4] ? 'disabled' : ''} onchange="togglePlugin(${i}, this.checked)">
                <i></i>
              </label>
              <button class="btn sm dg" data-plugin="${esc(p[0])}" onclick="deletePlugin(this.dataset.plugin)">${ic('trash', 13)} Remove</button>
            </div>
          </div>
        `).join('')}
      </div>`;
    } else if (instVm === 'tiles') {
      installedBodyHtml = `<div class="vm-tiles-grid">
        ${filteredInstalled.map(({ p, i }) => `
          <div class="vm-tile-card">
            <div class="vm-item-head">
              <div style="display: flex; align-items: center; gap: 12px; min-width: 0;">
                <div style="width: 40px; height: 40px; border-radius: 10px; background: rgba(56, 122, 255, 0.12); display: grid; place-items: center; color: var(--ac); font-size: 18px; flex-shrink: 0;">🧩</div>
                <div style="min-width: 0;">
                  <b class="vm-item-title" style="display:block;" title="${esc(p[0])}">${esc(p[0])}</b>
                  <div class="mu mono" style="font-size: 11.5px; margin-top: 2px;">${esc(p[1] || 'installed')}${p[3] ? ` &bull; ${esc(formatFileDate(p[3]))}` : ''}</div>
                </div>
              </div>
              <label class="sw" title="Toggle ${nounLow} state">
                <input type="checkbox" ${p[2] ? 'checked' : ''} ${p[4] ? 'disabled' : ''} onchange="togglePlugin(${i}, this.checked)">
                <i></i>
              </label>
            </div>
            <div class="vm-item-foot">
              <span class="chip">${p[2] ? 'Enabled' : 'Disabled'}</span>
              <button class="btn sm dg" data-plugin="${esc(p[0])}" onclick="deletePlugin(this.dataset.plugin)">${ic('trash', 13)} Uninstall</button>
            </div>
          </div>
        `).join('')}
      </div>`;
    } else {
      installedBodyHtml = `<div style="display: flex; flex-direction: column; gap: 10px;">
        ${filteredInstalled.map(({ p, i }) => `
          <div class="plugin-card">
            <div style="display: flex; align-items: center; gap: 16px; min-width: 0;">
              <div class="np-plugin-icon" aria-hidden="true">${ic('puzzle', 22)}</div>
              <div style="min-width: 0;">
                <b class="plugin-title">${esc(p[0])}</b>
                <div class="mu mono" style="font-size: 12px; margin-top: 3px;">
                  ${esc(p[1] || 'installed')}${p[3] ? ` &bull; ${esc(formatFileDate(p[3]))}` : ''}
                </div>
              </div>
            </div>
            <div class="row" style="gap: 12px; flex-shrink: 0;">
              <label class="sw" title="Toggle ${nounLow} state">
                <input type="checkbox" ${p[2] ? 'checked' : ''} ${p[4] ? 'disabled' : ''} onchange="togglePlugin(${i}, this.checked)">
                <i></i>
              </label>
              <button class="btn sm dg" data-plugin="${esc(p[0])}" onclick="deletePlugin(this.dataset.plugin)">${ic('trash', 13)} Uninstall</button>
            </div>
          </div>
        `).join('')}
      </div>`;
    }

    return `${srvh()}
    <div class="np-plugins-workspace">
      <div class="np-slot-container" data-np-slot="plugins:top" style="margin-bottom: 12px;"></div>
      <!-- Top Title and Tabs Bar -->
      <div class="row sb np-section-heading">
        <div>
          <h2>${nounPlur}</h2>
          <div class="mu" style="font-size: 12.5px; margin-top: 3px;">
            Browse and install Modrinth ${nounsLow} directly into <span class="mono">${targetFolder}</span>
            <div style="margin-top:6px;">Enabled/disabled reflects the next startup. Restart after installing, removing, or toggling ${nounsLow}.</div>
          </div>
        </div>
        <div class="row" style="gap: 8px; flex-wrap: wrap;">
          <button class="btn ${isBrowse ? 'pr' : ''} sm" onclick="setPlugTab('browse')">
            ${ic('store', 14)} Browse Store
          </button>
          <button class="btn ${!isBrowse ? 'pr' : ''} sm" onclick="setPlugTab('installed')">
            ${ic('puzzle', 14)} Installed <span id="plug-installed-count" class="chip" style="margin-left: 4px; padding: 1px 7px;">${S.plugins.length}</span>
          </button>
        </div>
      </div>

      ${isBrowse ? `
        <!-- Browse Modrinth Filter & Search Bar -->
        <div class="row" style="gap: 10px; flex-wrap: wrap; margin-bottom: 12px;">
          <div style="position: relative; flex: 1; min-width: 240px;">
            <input type="text" id="store-search" placeholder="${isMod ? 'Search mods on Modrinth (e.g. Lithium, FerriteCore, Simple Voice Chat)...' : 'Search plugins on Modrinth (e.g. EssentialsX, LuckPerms, WorldEdit)...'}" value="${esc(S.q || '')}" oninput="S.q=this.value; searchModrinthDebounced();" style="height: 40px; font-size: 14px; padding-left: 38px; border-radius: 10px;">
            <span style="position: absolute; left: 12px; top: 11px; color: var(--tx-muted);">${ic('search', 16)}</span>
          </div>

          <select style="width: auto; min-width: 155px; height: 40px; font-size: 13px;" onchange="S.plugLoader=this.value; refreshOpenVersionSelects(); searchModrinth();">
            <option value="" ${!S.plugLoader ? 'selected' : ''}>Loader: Auto (${isMod ? engineMeta(engSlug).label : 'All'})</option>
            ${isMod ? `
              <option value="fabric" ${S.plugLoader === 'fabric' ? 'selected' : ''}>Fabric</option>
              <option value="neoforge" ${S.plugLoader === 'neoforge' ? 'selected' : ''}>NeoForge</option>
              <option value="forge" ${S.plugLoader === 'forge' ? 'selected' : ''}>Forge</option>
              <option value="paper" ${S.plugLoader === 'paper' ? 'selected' : ''}>Paper</option>
              <option value="purpur" ${S.plugLoader === 'purpur' ? 'selected' : ''}>Purpur</option>
            ` : `
              <option value="paper" ${S.plugLoader === 'paper' ? 'selected' : ''}>Paper</option>
              <option value="purpur" ${S.plugLoader === 'purpur' ? 'selected' : ''}>Purpur</option>
              <option value="spigot" ${S.plugLoader === 'spigot' ? 'selected' : ''}>Spigot</option>
              <option value="bukkit" ${S.plugLoader === 'bukkit' ? 'selected' : ''}>Bukkit</option>
              <option value="folia" ${S.plugLoader === 'folia' ? 'selected' : ''}>Folia</option>
              <option value="velocity" ${S.plugLoader === 'velocity' ? 'selected' : ''}>Velocity</option>
              <option value="waterfall" ${S.plugLoader === 'waterfall' ? 'selected' : ''}>Waterfall</option>
              <option value="bungeecord" ${S.plugLoader === 'bungeecord' ? 'selected' : ''}>BungeeCord</option>
              <option value="fabric" ${S.plugLoader === 'fabric' ? 'selected' : ''}>Fabric</option>
              <option value="neoforge" ${S.plugLoader === 'neoforge' ? 'selected' : ''}>NeoForge</option>
              <option value="forge" ${S.plugLoader === 'forge' ? 'selected' : ''}>Forge</option>
            `}
          </select>

          <select id="plug-ver-sel" data-searchable="1" style="width: auto; min-width: 145px; height: 40px; font-size: 13px;" onchange="S.plugVer=this.value; searchModrinth();">
            <option value="" ${!S.plugVer ? 'selected' : ''}>Version: Any</option>
            ${verChoices.map(v => `<option value="${esc(v)}" ${S.plugVer === v ? 'selected' : ''}>${esc(v)}</option>`).join('')}
          </select>

          <select style="width: auto; min-width: 145px; height: 40px; font-size: 13px;" onchange="S.plugSort=this.value; searchModrinth();">
            <option value="relevance" ${(S.plugSort || 'relevance') === 'relevance' ? 'selected' : ''}>Sort: Relevance</option>
            <option value="downloads" ${S.plugSort === 'downloads' ? 'selected' : ''}>Sort: Downloads</option>
            <option value="follows" ${S.plugSort === 'follows' ? 'selected' : ''}>Sort: Follows</option>
            <option value="newest" ${S.plugSort === 'newest' ? 'selected' : ''}>Sort: Newest</option>
          </select>

          ${renderViewModeToggle('plug_store', 'list')}
        </div>

        <div class="row sb" style="margin-bottom: 12px; gap: 10px; flex-wrap: wrap;">
          <label class="row" style="gap: 8px; cursor: pointer; font-size: 12.5px; color: var(--tx-muted);">
            <input type="checkbox" id="plug-confirm-toggle" ${S.skipPlugConfirm ? 'checked' : ''} onchange="setPlugSkipConfirm(this.checked)">
            <span>Install ${nounsLow} immediately without confirmation prompt</span>
          </label>
          <span class="mu" style="font-size: 11.5px;">Direct Modrinth ${nounSing.toLowerCase()} install into <code class="mono">${targetFolder}</code></span>
        </div>

        <div id="store-grid" class="${storeGridClass}" style="${storeGridStyle}">
          ${storeBodyHtml}
        </div>
      ` : `
        <!-- Installed ${nounPlur} Filter & List -->
        <div class="row sb" style="gap: 10px; flex-wrap: wrap; margin-bottom: 14px;">
          <div style="position: relative; flex: 1; min-width: 240px; max-width: 420px;">
            <input type="text" placeholder="Filter installed ${nounsLow}..." value="${esc(S.plugFilter || '')}" oninput="S.plugFilter=this.value; R();" style="height: 38px; font-size: 13.5px; padding-left: 36px; border-radius: 10px;">
            <span style="position: absolute; left: 11px; top: 10px; color: var(--tx-muted);">${ic('search', 15)}</span>
          </div>
          <div class="row" style="gap: 8px; flex-wrap: wrap;">
            ${renderViewModeToggle('plug_inst', 'list')}
            <button class="btn sm" onclick="loadPluginsList()">${ic('refresh', 13)} Refresh</button>
            <button class="btn sm pr" onclick="setPlugTab('browse')">${ic('plus', 13)} Install More</button>
          </div>
        </div>

        ${installedBodyHtml}
      `}
    </div>`;
  },

  ver() {
    const sel = S.sel || S.core || 'Paper';
    const isProxy = ['Velocity', 'Waterfall', 'BungeeCord'].includes(sel);
    const srv = curServer();
    const vm = getViewMode('ver', 'list');
    return `${srvh()}
    <div class="card hero" style="margin-top:14px">
      <div class="row sb" style="align-items: center; flex-wrap: wrap; gap: 12px;">
        <div class="row" style="gap: 14px; align-items: center;">
          ${engineIcon(S.core || srv?.type || 'paper', 24, true)}
          <div>
            <span class="mu">Currently Running</span>
            <b class="big">${esc(S.core)} ${esc(S.ver)}</b>
            <span class="mu">${isSteelEngine(srv?.type) ? 'Steel • Written in Rust • No Java runtime required' : 'Java 25, 21, 17 &amp; 8 supported • Direct server installer'}</span>
          </div>
        </div>
        <button class="btn pr" data-srv-id="${esc(srv?.id || '')}" data-node-id="${esc(srv?.node_id || 'local')}" onclick="openReinstallModal(this.dataset.srvId, this.dataset.nodeId)">
          ${ic('refresh', 14)} Reinstall Server
        </button>
      </div>
    </div>

    <h2 style="margin-top:20px">1. Choose Server Engine</h2>
    <div class="grid g4">
      ${CORES.map(([name, desc]) => `
        <button class="card sel ${sel === name ? 'on' : ''}" onclick="S.sel='${name}'; S.vq=''; S.vf='All Families'; loadVersionsCatalog(); R();">
          <div class="engine-card-head" style="width:100%;">
            <span class="engine-card-title">
              ${engineIcon(name, 16, true)}
              <b>${name}</b>
            </span>
            ${S.core === name ? '<span class="badge online" style="font-size:10px;padding:2px 6px;"><i class="dot"></i>Active</span>' : ''}
          </div>
          <span class="mu" style="font-size:12px;">${desc}</span>
        </button>
      `).join('')}
    </div>

    <div class="card ver-section-card" style="margin-top:22px;">
      <div class="ver-section-head">
        <h2>2. Select ${esc(sel)} Version</h2>
        <span class="mu" style="font-size:12.5px;">Click any release to confirm installation &amp; optional world backup</span>
      </div>

      <div class="ver-toolbar" style="justify-content: space-between;">
        <div class="row" style="gap: 10px; flex-wrap: wrap; flex: 1;">
          <div class="ver-search-wrap">
            <input type="text" placeholder="${isSteelEngine(sel) ? 'Search Steel release tags...' : 'Search Minecraft versions...'}" value="${esc(S.vq || '')}" oninput="S.vq=this.value; renderVersionRows();" style="padding-left: 34px;">
            <span style="position: absolute; left: 10px; top: 10px; color: var(--tx-muted); pointer-events: none;">${ic('search', 15)}</span>
          </div>
          ${!isProxy && !isSteelEngine(sel) ? `
            <select data-searchable="1" onchange="S.vf=this.value; renderVersionRows();" style="width:auto; min-width: 220px;">
              <option value="All Families" ${(S.vf === 'All Families' || S.vf === 'All') ? 'selected' : ''}>All Release Families (1.0 – 26.x)</option>
              <option value="Popular" ${S.vf === 'Popular' ? 'selected' : ''}>Popular / Recommended</option>
              ${FAM.map(f => `<option value="${f[0]}" ${S.vf === f[0] ? 'selected' : ''}>${f[0]} — ${f[1]} (${f[2]})</option>`).join('')}
            </select>
          ` : ''}
        </div>
        ${renderViewModeToggle('ver', 'list')}
      </div>

      ${vm === 'list' ? `
        <div class="tw">
          <table>
            <thead><tr><th>Version</th><th>Release / Notes</th><th>Runtime</th><th>Engine</th><th style="text-align:right">Action</th></tr></thead>
            <tbody id="vres">${versionRowsHtml()}</tbody>
          </table>
        </div>
      ` : `
        <div id="vres" class="${vm === 'icons' ? 'vm-icons-grid' : 'vm-tiles-grid'}">${versionRowsHtml()}</div>
      `}
    </div>`;
  },

  bk() {
    const vm = getViewMode('bk', 'list');
    let body = '';
    if (S.loading?.backups) {
      body = `<div class="card">${loaderHtml('Loading server backups...')}</div>`;
    } else if (!S.bk.length) {
      body = `<div class="card mu" style="text-align:center;padding:24px;">No backups yet. Create your first archive to protect your world!</div>`;
    } else if (vm === 'icons' || vm === 'tiles') {
      body = `<div class="${vm === 'icons' ? 'vm-icons-grid wide' : 'vm-tiles-grid'}">
        ${S.bk.map(b => {
          const bName = b.name || b[0];
          const bSize = formatBytes(b.size) || b[1];
          const bDate = b.created_at ? new Date(b.created_at).toLocaleString() : b[2];
          return `<div class="${vm === 'icons' ? 'vm-icon-card' : 'vm-tile-card'}">
            <div class="${vm === 'icons' ? 'vm-plug-icon-center' : 'vm-item-head'}">
              <div class="row" style="gap: 10px; min-width: 0; ${vm === 'icons' ? 'flex-direction: column;' : ''}">
                <span class="fi-badge arc ${vm === 'icons' ? 'lg' : 'md'}">${ic('archive', vm === 'icons' ? 22 : 17)}</span>
                <b class="vm-item-title mono" style="max-width: 100%;" title="${esc(bName)}">${esc(bName)}</b>
              </div>
              <span class="chip">${esc(bSize)}</span>
            </div>
            <div class="mu" style="font-size: 12px;">${esc(bDate)}</div>
            <div class="vm-item-foot">
              <button class="btn sm" onclick="downloadBackup('${esc(bName)}')">${ic('download', 13)} Download</button>
              <button class="btn sm" data-backup-name="${esc(bName)}" onclick="restoreBackup(this.dataset.backupName, this)">${ic('refresh', 13)} Restore</button>
              <button class="btn sm dg" onclick="deleteBackup('${esc(bName)}')">${ic('trash', 13)} Delete</button>
            </div>
          </div>`;
        }).join('')}
      </div>`;
    } else {
      body = `<div class="card tw"><table>
        <thead><tr><th>Archive Name</th><th>Size</th><th>Created</th><th style="text-align:right">Actions</th></tr></thead>
        <tbody>
          ${S.bk.map(b => `
            <tr>
              <td class="mono"><b>${esc(b.name || b[0])}</b></td>
              <td>${formatBytes(b.size) || b[1]}</td>
              <td class="mu">${b.created_at ? new Date(b.created_at).toLocaleString() : b[2]}</td>
              <td style="text-align:right">
                <div class="row" style="justify-content:flex-end">
                  <button class="btn sm" onclick="downloadBackup('${esc(b.name || b[0])}')">${ic('download', 14)} Download</button>
                  <button class="btn sm" data-backup-name="${esc(b.name || b[0])}" onclick="restoreBackup(this.dataset.backupName, this)">${ic('refresh', 14)} Restore</button>
                  <button class="btn sm dg" onclick="deleteBackup('${esc(b.name || b[0])}')">${ic('trash', 14)} Delete</button>
                </div>
              </td>
            </tr>
          `).join('')}
        </tbody>
      </table></div>`;
    }
    return `${srvh()}${renderRestoreNotice()}` + `
    <div class="row sb" style="margin-top:14px; margin-bottom:12px; flex-wrap:wrap; gap:10px;">
      <span class="mu">${S.bk.length} backups recorded</span>
      <div class="row" style="gap:8px;">
        ${renderViewModeToggle('bk', 'list')}
        <button class="btn pr" onclick="createBackup()">${ic('plus', 14)} Create backup</button>
      </div>
    </div>
    ${body}`;
  },

  sc() {
    const srv = curServer();
    const schedules = Array.isArray(S.serverSchedules) ? S.serverSchedules : [];
    const isLoading = S.loading?.schedules;

    return `${srvh()}
    <div class="np-schedules-workspace">
      <div class="np-section-heading">
        <div class="row sb np-schedules-heading">
          <div>
            <h2>Schedules</h2>
            <p class="mu">
              Keep your server on schedule. Automate commands, restarts and backups.
            </p>
          </div>
          <div class="row" style="gap: 8px;">
            <button class="btn sm" onclick="loadServerSchedules('${esc(srv?.id || '')}')">${ic('refresh', 13)} Refresh</button>
            <button class="btn pr sm" onclick="openScheduleModal()" style="height: 36px; padding: 0 16px;">
              ${ic('plus', 14)} Create Schedule
            </button>
          </div>
        </div>
      </div>

      ${isLoading ? `<div class="card">${loaderHtml('Loading automated schedules...')}</div>` : (
        schedules.length === 0 ? `
          <div class="card mu np-empty-state">
            <div class="np-empty-icon" aria-hidden="true">${ic('clock', 24)}</div>
            <h3>No schedules yet</h3>
            <p style="margin: 0 auto 16px; max-width: 480px; font-size: 13px; line-height: 1.5;">
              Create automated schedules to restart your server during quiet hours, broadcast automated warnings to online players, or archive world backups.
            </p>
            <button class="btn pr" onclick="openScheduleModal()">${ic('plus', 14)} Create Your First Schedule</button>
          </div>
        ` : `
          <div class="card tw np-schedule-table">
            <table>
              <thead>
                <tr>
                  <th>Schedule Name</th>
                  <th>Action</th>
                  <th>Cron Expression</th>
                  <th>Next Run (UTC)</th>
                  <th>Last Execution</th>
                  <th>Status</th>
                  <th style="text-align: right;">Actions</th>
                </tr>
              </thead>
              <tbody>
                ${schedules.map(sch => {
                  const actBadge = renderScheduleTaskSummary(sch);
                  const cronHuman = describeCron(sch.cron_expr);
                  return `<tr>
                    <td>
                      <b style="color: var(--tx); display: block;">${esc(sch.name)}</b>
                      ${sch.payload ? `<span class="mono mu" style="font-size: 11.5px;">${esc(sch.payload)}</span>` : ''}
                    </td>
                    <td>${actBadge}</td>
                    <td class="mono" style="font-weight: 600;">${esc(sch.cron_expr)}</td>
                    <td style="font-size:12px;" title="${esc(cronHuman)}">${sch.is_active ? esc(formatScheduleUTC(sch.next_run_at)) : 'Paused'}</td>
                    <td style="font-size:12px;" class="mu">${esc(formatScheduleUTC(sch.last_run_at || sch.last_run))}<br><b>${esc(sch.last_status || 'Never run')}</b>${sch.last_error ? `<div class="upload-error">${esc(sch.last_error)}</div>` : ''}</td>
                    <td>
                      <label class="sw" title="Toggle active status">
                        <input type="checkbox" ${sch.is_active ? 'checked' : ''} onchange="toggleServerSchedule('${esc(sch.id)}', this.checked)">
                        <i></i>
                      </label>
                    </td>
                    <td style="text-align: right;">
                      <div class="row" style="justify-content: flex-end; gap: 6px;">
                        <button class="btn sm" ${sch.last_status === 'running' || sch.last_status === 'started' ? 'disabled' : ''} onclick="runScheduleNow('${esc(sch.id)}')" title="Run immediately once">${ic('play', 12)} Run Now</button>
                        <button class="btn sm" onclick="openScheduleModal('${esc(sch.id)}')" title="Edit schedule settings">${ic('edit', 12)} Edit</button>
                        <button class="btn sm dg" onclick="deleteServerSchedule('${esc(sch.id)}')" title="Delete schedule" aria-label="Delete schedule">${ic('trash', 12)}</button>
                      </div>
                    </td>
                  </tr>`;
                }).join('')}
              </tbody>
            </table>
          </div>
        `
      )}
    </div>`;
  },

  net() {
    const srv = curServer();
    const host = srvHost(srv);
    const sftpUser = S.sftp ? S.sftp.username : `admin.${srv?.id || 'server'}`;
    const sftpPort = S.sftp ? S.sftp.port : 2022;
    const ports = (S.ports && S.ports.length > 0) ? S.ports : [
      { id: 0, port: srv?.port || 25565, protocol: 'tcp/udp', notes: 'Primary Game Port', is_primary: true }
    ];

    return `${srvh()}
    <div style="display: flex; flex-direction: column; gap: 20px; margin-top: 14px;">
      <!-- Port Allocations Card -->
      <div class="card tw" style="padding: 22px;">
        <div class="row sb" style="margin-bottom: 16px; align-items: flex-start; flex-wrap: wrap; gap: 12px;">
          <div>
            <h2 style="margin: 0; font-size: 18px; color: #fff;">Network Allocations</h2>
            <p class="mu" style="font-size: 13px; margin: 4px 0 0;">Manage inbound port bindings for game traffic, voice chat, web maps, crossplay, and queries.</p>
          </div>
          <button class="btn pr sm" onclick="openAddPortModal()" style="height: 36px; padding: 0 16px;">
            ${ic('plus', 14)} Assign Port
          </button>
        </div>

        <table>
          <thead>
            <tr>
              <th>IP / Host</th>
              <th>Port</th>
              <th>Protocol</th>
              <th>Service / Notes</th>
              <th>Status</th>
              <th style="text-align: right;">Actions</th>
            </tr>
          </thead>
          <tbody>
            ${ports.map(p => `
              <tr>
                <td class="mono" style="color: var(--tx-muted);">${host}</td>
                <td>
                  <span class="mono" style="font-weight: 700; color: #fff; cursor: pointer;" onclick="navigator.clipboard.writeText('${p.port}'); toast('Port ${p.port} copied!');" title="Click to copy">
                    ${p.port}
                  </span>
                </td>
                <td>
                  <span class="badge" style="background: rgba(255,255,255,0.06); color: var(--tx-muted); font-size: 11px; padding: 3px 8px; border-radius: 4px; text-transform: uppercase;">
                    ${esc(p.protocol || 'tcp/udp')}
                  </span>
                </td>
                <td style="color: #fff; font-weight: 500;">
                  ${esc(p.notes || (p.is_primary ? 'Primary Game Port' : 'Additional Allocation'))}
                </td>
                <td>
                  ${p.is_primary ? 
                    `<span class="badge online"><i class="dot"></i>Primary</span>` : 
                    `<span class="badge" style="background: rgba(255,255,255,0.05); color: var(--tx-muted);">Secondary</span>`}
                </td>
                <td style="text-align: right;">
                  ${p.is_primary ? `
                    <span class="mu" style="font-size: 12px;">Default Game Port</span>
                  ` : `
                    <div class="row" style="justify-content: flex-end; gap: 8px;">
                      <button class="btn sm" onclick="setPrimaryPort(${p.id})" title="Set as primary game port">Make Primary</button>
                      <button class="btn dg sm" onclick="deletePort(${p.id}, ${p.port})" title="Delete allocation">${ic('trash', 14)}</button>
                    </div>
                  `}
                </td>
              </tr>
            `).join('')}
          </tbody>
        </table>
      </div>

      <!-- Secondary Row: SFTP Access & Common Ports -->
      <div class="grid g2" style="gap: 20px;">
        <p class="mu">File transfer connections are managed in the dedicated SFTP section.</p>

        <div class="card" style="padding: 22px;">
          <h2 style="font-size: 16px; margin-bottom: 8px; color: #fff;">Common Plugin Ports</h2>
          <p class="mu" style="font-size: 13px; margin-bottom: 12px;">Default ports frequently needed by popular Minecraft server plugins:</p>
          <div style="display: flex; flex-direction: column; gap: 8px; font-size: 13px;">
            <div class="row sb" style="padding: 6px 0; border-bottom: 1px solid var(--border);">
              <span>Simple Voice Chat (UDP)</span>
              <span class="mono" style="color: var(--ac); font-weight: 700;">24454</span>
            </div>
            <div class="row sb" style="padding: 6px 0; border-bottom: 1px solid var(--border);">
              <span>Dynmap Web Interface</span>
              <span class="mono" style="color: var(--ac); font-weight: 700;">8123</span>
            </div>
            <div class="row sb" style="padding: 6px 0; border-bottom: 1px solid var(--border);">
              <span>Geyser / Floodgate (Bedrock)</span>
              <span class="mono" style="color: var(--ac); font-weight: 700;">19132</span>
            </div>
            <div class="row sb" style="padding: 6px 0;">
              <span>NuVotifier / Voting Port</span>
              <span class="mono" style="color: var(--ac); font-weight: 700;">8192</span>
            </div>
          </div>
        </div>
      </div>
    </div>`;
  },

  db() {
    const srv = curServer();
    const list = Array.isArray(S.databases) ? S.databases : [];
    const caps = S.dbCapabilities || { sqlite: true, engines_label: 'Built-in SQLite (pure Go)' };
    const vm = getViewMode('db', 'list');

    const emptyHtml = `<div class="card" style="text-align: center; padding: 36px 20px;">
      <div style="max-width: 440px; margin: 0 auto;">
        <div style="font-size: 15px; font-weight: 700; color: var(--tx); margin-bottom: 6px;">No databases provisioned for ${esc(srv?.name || 'this server')}</div>
        <div class="mu" style="font-size: 13px; margin-bottom: 16px;">Create an isolated SQLite database for LuckPerms, CoreProtect, Economy plugins, or custom data — inspect schemas and run live SQL queries right in the browser.</div>
        <button type="button" class="btn pr" onclick="openCreateDatabaseModal()">${ic('plus', 14)} Create First Database</button>
      </div>
    </div>`;

    let dbBody = '';
    if (S.loading?.databases) {
      dbBody = `<div class="card">${loaderHtml('Loading server databases...')}</div>`;
    } else if (!list.length) {
      dbBody = emptyHtml;
    } else if (vm === 'icons' || vm === 'tiles') {
      dbBody = `<div class="${vm === 'icons' ? 'vm-icons-grid wide' : 'vm-tiles-grid'}">
        ${list.map(d => {
          const revealed = S.dbRevealIds && S.dbRevealIds.has(d.id);
          const jdbcUrl = d.jdbc_url || (d.file_path ? `jdbc:sqlite:${d.file_path}` : '');
          const dsnPath = d.sqlite_dsn || d.file_path || '';
          return `<div class="${vm === 'icons' ? 'vm-icon-card' : 'vm-tile-card'}">
            <div class="vm-item-head">
              <div class="row" style="gap: 10px; min-width: 0;">
                <span class="fi-badge db md">${ic('database', 17)}</span>
                <div style="min-width: 0;">
                  <b class="vm-item-title" style="display:block;" title="${esc(d.name)}">${esc(d.name)}</b>
                  <span class="chip" style="font-size: 10px; margin-top: 2px;">${esc((d.engine || 'sqlite').toUpperCase())}</span>
                </div>
              </div>
              <span class="mu mono" style="font-size: 11px;">${Number(d.table_count) || 0} tbl &bull; ${formatBytes(d.size_bytes || 0)}</span>
            </div>
            <div style="display: flex; flex-direction: column; gap: 5px;">
              <div class="db-cred-row">
                <span class="mu" style="font-size: 11px; width: 32px;">User</span>
                <span class="db-secret-chip">${esc(d.username || 'u_db')}</span>
                <button type="button" class="btn-ico" data-val="${esc(d.username || '')}" onclick="copyDbText(this.dataset.val, 'Username')" title="Copy Username">${ic('copy', 12)}</button>
              </div>
              ${vm === 'tiles' ? `<div class="db-cred-row">
                <span class="mu" style="font-size: 11px; width: 32px;">Pass</span>
                <span class="db-secret-chip">${revealed ? esc(d.password || '') : '&bull;&bull;&bull;&bull;&bull;&bull;&bull;&bull;&bull;&bull;'}</span>
                <button type="button" class="btn-ico" onclick="toggleDbPasswordReveal('${esc(d.id)}')" title="${revealed ? 'Hide password' : 'Reveal password'}">${ic('eye', 12)}</button>
                <button type="button" class="btn-ico" data-val="${esc(d.password || '')}" onclick="copyDbText(this.dataset.val, 'Password')" title="Copy Password">${ic('copy', 12)}</button>
                <button type="button" class="btn-ico" data-id="${esc(d.id)}" data-name="${esc(d.name)}" onclick="rotateDatabasePassword(this.dataset.id, this.dataset.name)" title="Rotate Password">${ic('key', 12)}</button>
              </div>` : ''}
            </div>
            <div class="vm-item-foot">
              <button type="button" class="btn sm pr" onclick="openSQLStudioModal('${esc(d.id)}')">${ic('code', 13)} SQL Studio</button>
              <div class="row" style="gap: 4px;">
                <button type="button" class="btn sm" data-val="${esc(jdbcUrl)}" onclick="copyDbText(this.dataset.val, 'JDBC URL')" title="Copy JDBC URL">JDBC</button>
                <button type="button" class="btn sm" onclick="exportServerDatabase('${esc(d.id)}', 'sql')" title="Download SQL Dump">.sql</button>
                <button type="button" class="btn sm dg" data-id="${esc(d.id)}" data-name="${esc(d.name)}" onclick="deleteServerDatabase(this.dataset.id, this.dataset.name)" title="Delete Database">${ic('trash', 13)}</button>
              </div>
            </div>
          </div>`;
        }).join('')}
      </div>`;
    } else {
      dbBody = `<div class="card tw">
        <table>
          <thead>
            <tr>
              <th>Database &amp; File</th>
              <th>Credentials</th>
              <th>JDBC / DSN</th>
              <th>Tables &amp; Size</th>
              <th style="text-align: right;">Actions</th>
            </tr>
          </thead>
          <tbody>
            ${list.map(d => {
              const revealed = S.dbRevealIds && S.dbRevealIds.has(d.id);
              const jdbcUrl = d.jdbc_url || (d.file_path ? `jdbc:sqlite:${d.file_path}` : '');
              const dsnPath = d.sqlite_dsn || d.file_path || '';
              return `<tr>
                <td>
                  <div style="font-weight: 700; color: var(--tx);">${esc(d.name)}</div>
                  <div class="row" style="gap: 5px; margin-top: 4px;">
                    <span class="chip" style="font-size: 10.5px;">${esc((d.engine || 'sqlite').toUpperCase())}</span>
                    <span class="mu mono" style="font-size: 11px; max-width: 240px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;" title="${esc(dsnPath)}">${esc(dsnPath)}</span>
                  </div>
                </td>
                <td>
                  <div style="display: flex; flex-direction: column; gap: 6px;">
                    <div class="db-cred-row">
                      <span class="mu" style="font-size: 11px; width: 34px;">User</span>
                      <span class="db-secret-chip">${esc(d.username || 'u_db')}</span>
                      <button type="button" class="btn-ico" data-val="${esc(d.username || '')}" onclick="copyDbText(this.dataset.val, 'Username')" title="Copy Username">${ic('copy', 12)}</button>
                    </div>
                    <div class="db-cred-row">
                      <span class="mu" style="font-size: 11px; width: 34px;">Pass</span>
                      <span class="db-secret-chip">${revealed ? esc(d.password || '') : '&bull;&bull;&bull;&bull;&bull;&bull;&bull;&bull;&bull;&bull;&bull;&bull;'}</span>
                      <button type="button" class="btn-ico" onclick="toggleDbPasswordReveal('${esc(d.id)}')" title="${revealed ? 'Hide password' : 'Reveal password'}">${ic('eye', 12)}</button>
                      <button type="button" class="btn-ico" data-val="${esc(d.password || '')}" onclick="copyDbText(this.dataset.val, 'Password')" title="Copy Password">${ic('copy', 12)}</button>
                      <button type="button" class="btn-ico" data-id="${esc(d.id)}" data-name="${esc(d.name)}" onclick="rotateDatabasePassword(this.dataset.id, this.dataset.name)" title="Rotate Password">${ic('key', 12)}</button>
                    </div>
                  </div>
                </td>
                <td>
                  <div class="row" style="gap: 6px; flex-wrap: wrap;">
                    <button type="button" class="btn sm" data-val="${esc(jdbcUrl)}" onclick="copyDbText(this.dataset.val, 'JDBC URL')" title="${esc(jdbcUrl)}">${ic('copy', 12)} Copy JDBC</button>
                    <button type="button" class="btn sm" data-val="${esc(dsnPath)}" onclick="copyDbText(this.dataset.val, 'SQLite Path')" title="${esc(dsnPath)}">${ic('copy', 12)} Copy Path</button>
                  </div>
                </td>
                <td>
                  <div class="mono" style="font-weight: 600; color: var(--tx);">${Number(d.table_count) || 0} table${Number(d.table_count) === 1 ? '' : 's'}</div>
                  <div class="mu mono" style="font-size: 11.5px; margin-top: 2px;">${formatBytes(d.size_bytes || 0)} &bull; max ${d.max_connections || 25}</div>
                </td>
                <td style="text-align: right;">
                  <div class="row" style="justify-content: flex-end; gap: 6px; flex-wrap: wrap;">
                    <button type="button" class="btn sm pr" onclick="openSQLStudioModal('${esc(d.id)}')">${ic('code', 13)} SQL Studio</button>
                    <button type="button" class="btn sm" onclick="exportServerDatabase('${esc(d.id)}', 'sql')" title="Download SQL Dump">${ic('download', 12)} .sql</button>
                    <button type="button" class="btn sm" onclick="exportServerDatabase('${esc(d.id)}', 'db')" title="Download SQLite Binary">${ic('download', 12)} .db</button>
                    <button type="button" class="btn sm dg" data-id="${esc(d.id)}" data-name="${esc(d.name)}" onclick="deleteServerDatabase(this.dataset.id, this.dataset.name)" title="Delete Database">${ic('trash', 13)}</button>
                  </div>
                </td>
              </tr>`;
            }).join('')}
          </tbody>
        </table>
      </div>`;
    }

    return `${srvh()}
    <div style="margin-top: 14px; display: flex; flex-direction: column; gap: 14px;">
      <div class="card" style="padding: 18px 20px;">
        <div class="row sb" style="flex-wrap: wrap; gap: 12px;">
          <div>
            <div class="row" style="gap: 10px; align-items: center;">
              <h2 style="margin: 0; font-size: 18px;">Server Databases</h2>
              <span class="chip">${list.length} database${list.length === 1 ? '' : 's'}</span>
              <span class="badge online" style="font-size: 11px;"><i class="dot"></i>${esc(caps.engines_label || 'Built-in SQLite (pure Go)')}</span>
            </div>
            <div class="mu" style="font-size: 12.5px; margin-top: 4px;">
              Real isolated per-server SQLite databases stored directly on <b>${esc(srv?.node_name || 'this node')}</b> with built-in SQL Studio &amp; binary/SQL exports.
            </div>
          </div>
          <div class="row" style="gap: 8px; flex-wrap: wrap;">
            ${renderViewModeToggle('db', 'list')}
            <button type="button" class="btn sm" onclick="loadDatabases()">${ic('refresh', 13)} Refresh</button>
            <button type="button" class="btn sm pr" onclick="openCreateDatabaseModal()">${ic('plus', 14)} Create Database</button>
          </div>
        </div>
      </div>
      ${dbBody}
    </div>`;
  },

  sub() {
    const srv = curServer();
    const data = S.subdomainData;
    const draft = S.subdomainDraft || computeSubdomainDraft(data, srv);
    const port = Number(data?.server_port || srv?.port || 25565);
    const fqdn = getSubdomainPreviewFqdn(draft);
    const fullDisplay = port === 25565 ? fqdn : `${fqdn} (SRV -> :${port})`;
    const targetVal = draft.target_host || data?.default_target || srvHost(srv) || '';
    const rawRecords = Array.isArray(data?.records) ? data.records : [];
    const srvRecordHost = fqdn.startsWith('play.') || !draft.base_domain
      ? `_minecraft._tcp.${fqdn}`
      : `_minecraft._tcp.${draft.label || 'play'}.${draft.base_domain}`;
    const displayRecords = rawRecords.length ? rawRecords : [
      { type: (/^\d{1,3}(\.\d{1,3}){3}$/.test(targetVal) ? 'A' : (targetVal.includes(':') ? 'AAAA' : 'CNAME')), name: draft.label || 'play', full_name: fqdn, value: targetVal || 'YOUR_NODE_IP', ttl: 300, note: 'Points your custom address to this NeoPanel node.' },
      { type: 'SRV', name: `_minecraft._tcp.${draft.label || 'play'}`, full_name: srvRecordHost, value: `0 5 ${port} ${fqdn}`, ttl: 300, note: `Routes Minecraft Java clients automatically to port ${port}.` }
    ];
    const verify = S.subdomainVerify;

    return `${srvh()}
    <div style="margin-top: 14px; display: flex; flex-direction: column; gap: 16px;">
      <div class="card" style="padding: 20px;">
        <div class="row sb" style="margin-bottom: 14px; flex-wrap: wrap; gap: 10px;">
          <div>
            <h2 style="margin: 0 0 4px; font-size: 18px; color: var(--tx);">Custom Domain &amp; SRV Manager</h2>
            <div class="mu" style="font-size: 12.5px;">Bind a vanity Minecraft address (<code>play.yourdomain.com</code>) with automatic SRV port mapping &amp; live DNS verification.</div>
          </div>
          <label class="row" style="gap: 10px; cursor: pointer;">
            <span style="font-size: 12.5px; font-weight: 600; color: var(--tx);">Custom Domain Active</span>
            <span class="sw">
              <input type="checkbox" id="sub-enabled" ${draft.enabled ? 'checked' : ''} onchange="onSubdomainFieldChange()">
              <i></i>
            </span>
          </label>
        </div>

        ${S.subdomainLoading && !data ? loaderHtml('Loading subdomain & DNS records...') : `
          <div class="dns-preview-banner" style="margin-bottom: 16px;">
            <div style="min-width: 0; flex: 1;">
              <div class="mu" style="font-size: 11px; text-transform: uppercase; letter-spacing: 0.05em; font-weight: 600;">Player Connect Address</div>
              <div id="sub-fqdn-preview" class="dns-fqdn-big">${esc(fullDisplay)}</div>
            </div>
            <div class="row" style="gap: 8px; flex-wrap: wrap;">
              <span class="badge ${draft.enabled ? 'online' : 'offline'}"><i class="dot"></i>${draft.enabled ? 'Enabled' : 'Disabled'}</span>
              <button type="button" id="sub-copy-fqdn-btn" class="btn sm" data-val="${esc(fqdn)}" onclick="copyDbText(this.dataset.val, 'Subdomain Address')">${ic('copy', 12)} Copy Address</button>
            </div>
          </div>

          <div class="grid g3" style="gap: 12px; margin-bottom: 14px;">
            <label class="f" style="margin-bottom: 0;">
              <span>Subdomain Label</span>
              <input type="text" id="sub-label" value="${esc(draft.label || '')}" placeholder="e.g. play, smp, skyblock" oninput="onSubdomainFieldChange()">
            </label>
            <label class="f" style="margin-bottom: 0;">
              <span>Base Domain</span>
              <input type="text" id="sub-base" value="${esc(draft.base_domain || '')}" placeholder="e.g. yourdomain.com" oninput="onSubdomainFieldChange()">
            </label>
            <label class="f" style="margin-bottom: 0;">
              <span>Target Node IP / FQDN</span>
              <input type="text" id="sub-target" value="${esc(targetVal)}" placeholder="e.g. 203.0.113.10 or node.example.com" oninput="onSubdomainFieldChange()">
            </label>
          </div>

          <div class="row" style="gap: 10px; flex-wrap: wrap;">
            <button type="button" class="btn pr" onclick="saveSubdomainConfig(this)">${ic('download', 14)} Save Subdomain</button>
            <button type="button" class="btn" onclick="verifySubdomainDNS(this)" ${S.subdomainVerifying ? 'disabled' : ''}>${ic('check', 14)} ${S.subdomainVerifying ? 'Checking DNS...' : 'Verify Live DNS'}</button>
          </div>
        `}
      </div>

      ${verify ? `
        <div class="dns-verify-box ${esc(verify.status || 'pending')}">
          <div class="row sb" style="margin-bottom: 6px; flex-wrap: wrap; gap: 8px;">
            <b style="color: var(--tx); display: inline-flex; align-items: center; gap: 7px;">
              ${verify.status === 'verified' ? ic('check', 15) : ic('alert', 15)}
              DNS Check for <span class="mono">${esc(verify.fqdn || fqdn)}</span>
            </b>
            <span class="badge ${verify.status === 'verified' ? 'online' : (verify.status === 'partial' ? 'starting' : 'offline')}">
              <i class="dot"></i>${verify.status === 'verified' ? 'Verified' : (verify.status === 'partial' ? 'Partial Match' : 'Not Propagated Yet')}
            </span>
          </div>
          <div style="color: var(--tx);">${esc(verify.summary || '')}</div>
          <div class="row" style="gap: 8px; flex-wrap: wrap; margin-top: 8px; font-size: 12px;">
            <span class="chip mono">A/Host: ${verify.a_records?.length ? esc(verify.a_records.join(', ')) : (verify.cname ? esc(verify.cname) : 'None')}</span>
            <span class="chip mono">SRV: ${verify.srv_records?.length ? verify.srv_records.map(r => `${r.target}:${r.port}`).join(', ') : 'None detected'}</span>
          </div>
          ${Array.isArray(verify.warnings) && verify.warnings.length ? `
            <div class="mu" style="font-size: 12px; margin-top: 8px;">
              ${verify.warnings.map(w => `<div>&bull; ${esc(w)}</div>`).join('')}
            </div>` : ''}
        </div>
      ` : ''}

      <div class="card tw">
        <div class="row sb" style="margin-bottom: 12px; flex-wrap: wrap; gap: 8px;">
          <div>
            <h2 style="margin: 0 0 3px; font-size: 16px; color: var(--tx);">Required DNS &amp; SRV Records</h2>
            <div class="mu" style="font-size: 12.5px;">Add these records at your DNS registrar (Cloudflare, Namecheap, Route53, Porkbun) so players connect without typing port <code class="mono">:${port}</code>.</div>
          </div>
        </div>
        <table>
          <thead>
            <tr>
              <th>Type</th>
              <th>Name / Host</th>
              <th>Target / Value</th>
              <th>TTL</th>
              <th style="text-align: right;">Copy</th>
            </tr>
          </thead>
          <tbody>
            ${displayRecords.map(rec => `
              <tr>
                <td><span class="chip mono" style="font-weight: 700;">${esc(rec.type)}</span></td>
                <td>
                  <div class="mono" style="font-weight: 600; color: var(--tx);">${esc(rec.name || rec.full_name)}</div>
                  ${rec.full_name && rec.full_name !== rec.name ? `<div class="mu mono" style="font-size: 11px;">${esc(rec.full_name)}</div>` : ''}
                </td>
                <td>
                  <span class="dns-val-chip mono" style="font-size: 12px;">${esc(rec.value)}</span>
                  ${rec.note ? `<div class="mu" style="font-size: 11.5px; margin-top: 3px;">${esc(rec.note)}</div>` : ''}
                </td>
                <td class="mono mu">${esc(rec.ttl || 300)}s</td>
                <td style="text-align: right;">
                  <button type="button" class="btn sm" data-val="${esc(rec.value)}" onclick="copyDbText(this.dataset.val, '${esc(rec.type)} Value')">${ic('copy', 12)} Copy</button>
                </td>
              </tr>
            `).join('')}
          </tbody>
        </table>
        <div class="cf-proxy-warn" style="margin-top: 14px;">
          <b>Cloudflare Tip:</b> Set your Minecraft <code>A</code> / <code>CNAME</code> proxy status to <b>DNS Only (grey cloud)</b>. Orange-cloud HTTP proxying blocks Minecraft Java TCP traffic!
        </div>
      </div>
    </div>`;
  },

  stp() {
    return V.set();
  },

  client() {
    const b = S.branding || DEFAULT_BRANDING;
    const enabled = Boolean(b.client_area_enabled);
    const available = b.client_area_available !== false;
    const rawUrl = (b.client_area_url || '').trim();
    const mode = b.client_area_open_mode === 'embed' ? 'embed' : 'tab';
    const label = (b.client_area_label || 'Client Area').trim() || 'Client Area';
    const subtitle = (b.client_area_subtitle || 'Purchase servers, manage subscriptions, and view invoices in your storefront.').trim();
    const isAdmin = S.user?.role === 'admin';

    if (!enabled) {
      return `<div class="np-status-shell">
        <div class="card np-status-card">
          <div class="np-status-code warn">${ic('store', 14)} NOT CONFIGURED</div>
          <h1 style="font-size:24px; margin:0 0 10px; color:var(--tx);">${esc(label)} Is Not Configured Yet</h1>
          <p class="mu" style="font-size:13.5px; line-height:1.6; margin:0;">
            No external hosting storefront or billing portal has been linked to this panel yet (for example <span class="mono">https://client.hosting.net</span>).
          </p>
          <div class="row" style="justify-content:center; gap:10px; margin-top:22px; flex-wrap:wrap;">
            <a href="/" onclick="return handleNavClick(event, 'dash')" class="btn pr" style="text-decoration:none;">${ic('server', 14)} Back to Servers</a>
            ${isAdmin ? `<a href="/admin/branding" onclick="return handleNavClick(event, 'admin', { adminTab: 4 })" class="btn" style="text-decoration:none;">${ic('settings', 14)} Connect Storefront</a>` : ''}
          </div>
        </div>
      </div>`;
    }

    if (!available) {
      return `<div class="np-status-shell">
        <div class="card np-status-card">
          <div class="np-status-code err">${ic('alert', 14)} 503 UNAVAILABLE</div>
          <h1 style="font-size:24px; margin:0 0 10px; color:var(--tx);">${esc(label)} Temporarily Unavailable</h1>
          <p class="mu" style="font-size:13.5px; line-height:1.6; margin:0;">
            The connected ordering storefront is currently marked as unavailable by an administrator. Running game servers in your panel remain unaffected.
          </p>
          <div class="row" style="justify-content:center; gap:10px; margin-top:22px; flex-wrap:wrap;">
            <button type="button" class="btn" onclick="applyRoute('/client-area', { replace: true })">${ic('refresh', 14)} Retry</button>
            <a href="/" onclick="return handleNavClick(event, 'dash')" class="btn pr" style="text-decoration:none;">${ic('server', 14)} Back to Servers</a>
            ${isAdmin ? `<a href="/admin/branding" onclick="return handleNavClick(event, 'admin', { adminTab: 4 })" class="btn" style="text-decoration:none;">${ic('settings', 14)} Storefront Settings</a>` : ''}
          </div>
        </div>
      </div>`;
    }

    if (!rawUrl) {
      return `<div class="np-status-shell">
        <div class="card np-status-card">
          <div class="np-status-code warn">${ic('link', 14)} 422 URL MISSING</div>
          <h1 style="font-size:24px; margin:0 0 10px; color:var(--tx);">Storefront URL Is Empty</h1>
          <p class="mu" style="font-size:13.5px; line-height:1.6; margin:0;">
            ${esc(label)} is enabled, but its website address is missing or invalid. Add a valid <span class="mono">https://...</span> address under Admin → Appearance.
          </p>
          <div class="row" style="justify-content:center; gap:10px; margin-top:22px; flex-wrap:wrap;">
            <a href="/" onclick="return handleNavClick(event, 'dash')" class="btn pr" style="text-decoration:none;">${ic('server', 14)} Back to Servers</a>
            ${isAdmin ? `<a href="/admin/branding" onclick="return handleNavClick(event, 'admin', { adminTab: 4 })" class="btn" style="text-decoration:none;">${ic('settings', 14)} Set Storefront URL</a>` : ''}
          </div>
        </div>
      </div>`;
    }

    let safeHost = rawUrl;
    try {
      safeHost = new URL(rawUrl, window.location.origin).origin;
    } catch (e) {}

    if (mode === 'embed') {
      return `
      <div class="row sb" style="margin-bottom: 14px; flex-wrap: wrap; gap: 12px;">
        <div>
          <div class="row" style="gap: 8px; flex-wrap: wrap;">
            <h1 style="font-size: 22px; font-weight: 800; margin: 0; color: var(--tx);">${esc(label)}</h1>
            <span class="badge online"><i class="dot"></i>Connected</span>
            <span class="chip mono" style="font-size: 11px;">${esc(safeHost)}</span>
          </div>
          ${subtitle ? `<p class="mu" style="font-size: 13px; margin: 4px 0 0;">${esc(subtitle)}</p>` : ''}
        </div>
        <div class="row" style="gap: 8px; flex-wrap: wrap;">
          ${isAdmin ? `<a href="/admin/branding" onclick="return handleNavClick(event, 'admin', { adminTab: 4 })" class="btn sm" style="text-decoration:none;">${ic('settings', 13)} Configure</a>` : ''}
          <a href="${esc(rawUrl)}" target="_blank" rel="noopener noreferrer" class="btn sm pr" style="text-decoration:none;">${ic('external', 13)} Open in New Tab</a>
        </div>
      </div>
      ${S.clientIframeWarning ? `
        <div class="np-maint-banner" style="margin-bottom: 12px;">
          <span>External site may block iframe embedding (<code class="mono">X-Frame-Options</code> / CSP). Use Open in New Tab if the frame stays blank.</span>
          <a href="${esc(rawUrl)}" target="_blank" rel="noopener noreferrer" class="btn sm pr" style="text-decoration:none;">${ic('external', 13)} Launch Storefront</a>
        </div>` : ''}
      <div class="client-area-iframe-wrap">
        <iframe id="client-area-frame" src="${esc(rawUrl)}" title="${esc(label)}" referrerpolicy="no-referrer-when-downgrade" onerror="S.clientIframeWarning=true; R();"></iframe>
      </div>`;
    }

    return `<div class="np-status-shell">
      <div class="card np-status-card" style="max-width: 580px;">
        <div class="np-status-code">${ic('store', 14)} STOREFRONT PORTAL</div>
        <h1 style="font-size:25px; margin:0 0 10px; color:var(--tx);">${esc(label)}</h1>
        <p class="mu" style="font-size:13.5px; line-height:1.6; margin:0 0 16px;">${esc(subtitle)}</p>
        <div class="chip mono" style="display:inline-flex; max-width:100%; overflow:hidden; text-overflow:ellipsis; padding:6px 12px; font-size:12px;">
          ${esc(rawUrl)}
        </div>
        <div class="row" style="justify-content:center; gap:10px; margin-top:22px; flex-wrap:wrap;">
          <a href="${esc(rawUrl)}" target="_blank" rel="noopener noreferrer" class="btn pr" style="text-decoration:none;">${ic('external', 14)} Open ${esc(label)}</a>
          <a href="/" onclick="return handleNavClick(event, 'dash')" class="btn" style="text-decoration:none;">${ic('server', 14)} Back to Servers</a>
          ${isAdmin ? `<a href="/admin/branding" onclick="return handleNavClick(event, 'admin', { adminTab: 4 })" class="btn" style="text-decoration:none;">${ic('settings', 14)} Configure</a>` : ''}
        </div>
      </div>
    </div>`;
  },

  serverpending() {
    const busy = S.serverRoutePending?.busy;
    return `<div class="np-status-shell"><div class="card np-status-card" role="status">
      <h1>${busy ? 'Loading server…' : 'Server temporarily unavailable'}</h1>
      <p class="mu">The server list could not be fully loaded. The panel or a node may be offline. This does not mean the server was deleted.</p>
      <div class="row" style="justify-content:center; gap:10px; margin-top:20px;">
        <button type="button" class="btn pr" onclick="retryServerRoute()" ${busy ? 'disabled' : ''}>${busy ? 'Loading…' : 'Retry'}</button>
        <a href="/" onclick="return handleNavClick(event, 'dash')" class="btn">Back to Servers</a>
      </div>
    </div></div>`;
  },

  notfound() {
    const path = S.notFoundPath || window.location.pathname || '/unknown';
    const srv = curServer();
    return `<div class="np-status-shell">
      <div class="card np-status-card">
        <div class="np-status-code err">${ic('alert', 14)} 404 NOT FOUND</div>
        <h1 style="font-size:26px; font-weight:800; margin:0 0 10px; color:var(--tx);">Page Not Found</h1>
        <p class="mu" style="font-size:13.5px; line-height:1.6; margin:0 0 14px;">
          The route <span class="chip mono" style="padding:2px 8px; font-size:12px;">${esc(path)}</span> does not exist, or the requested server or extension is unavailable.
        </p>
        <div class="row" style="justify-content:center; gap:10px; margin-top:20px; flex-wrap:wrap;">
          <a href="/" onclick="return handleNavClick(event, 'dash')" class="btn pr" style="text-decoration:none;">${ic('server', 14)} Back to Servers</a>
          ${srv ? `<button type="button" class="btn" onclick="applyRoute(serverRouteUrl('${esc(srv.node_id || 'local')}', '${esc(srv.id)}', 'con'))">${ic('terminal', 14)} Open Active Server</button>` : ''}
        </div>
      </div>
    </div>`;
  },

  maintenance() {
    const b = S.branding || DEFAULT_BRANDING;
    const title = (b.maintenance_title || 'Scheduled Maintenance').trim();
    const message = (b.maintenance_message || 'We are upgrading NeoPanel nodes. Running game servers stay online during panel maintenance.').trim();
    const eta = (b.maintenance_eta || '').trim();
    const isAdminPreview = S.maintPreview && S.user?.role === 'admin';
    return `<div class="np-status-shell full-screen">
      <div class="card np-status-card">
        <div class="llogo" style="margin-bottom:14px;">${brandLogoHtml(true, b)}</div>
        <div class="np-status-code warn">${ic('alert', 14)} 503 MAINTENANCE</div>
        <h1 style="font-size:25px; font-weight:800; margin:0 0 10px; color:var(--tx);">${esc(title)}</h1>
        <p class="mu" style="font-size:13.5px; line-height:1.65; margin:0;">${esc(message)}</p>
        ${eta ? `<div class="chip mono" style="margin-top:14px; display:inline-flex;">Estimated return: ${esc(eta)}</div>` : ''}
        <div class="row" style="justify-content:center; gap:10px; margin-top:24px; flex-wrap:wrap;">
          ${isAdminPreview
            ? `<button type="button" class="btn pr" onclick="toggleMaintPreview(false)">Exit Maintenance Preview</button>`
            : `<button type="button" class="btn pr" onclick="checkAuth()">${ic('refresh', 14)} Check Status</button>
               <button type="button" class="btn" onclick="logout()">${ic('lock', 14)} Sign Out</button>`}
        </div>
      </div>
    </div>`;
  },

  mod() {
    const page = window.NP?._pages?.get(S.modPageId);
    if (!page) {
      S.notFoundPath = `/mod/${S.modPageId || ''}`;
      return V.notfound();
    }
    const header = page.scope === 'server' && curServer() ? srvh() : '';
    return `${header}
    <div style="${header ? 'margin-top: 16px;' : ''}">
      <div id="np-mod-page-root" data-mod-page="${esc(page.id)}"></div>
    </div>`;
  },

  set() {
    const p = S.props;
    const srv = curServer();
    const native = isSteelEngine(srv?.type);
    const o = (v, a) => a.map(x => `<option ${x === v ? 'selected' : ''}>${x}</option>`).join('');
    const sw = (k, l) => `<label class="row sb" style="padding:7px 0; border-bottom:1px solid var(--ln);">
      <span>${l}</span>
      <span class="sw">
        <input type="checkbox" id="prop-sw-${k}" ${p[k] ? 'checked' : ''} onchange="onSettingsRuleToggle('${k}', this.checked)">
        <i></i>
      </span>
    </label>`;

    const rawJava = srv?.java_path || '';
    const curJava = resolveJavaSelectValue(rawJava, srv?.version, srv?.type);
    const isCustom = curJava === 'custom';
    const curMem = getServerMaxRamMB(srv);
    const memMin = srv?.memory_min || Math.min(1024, curMem);
    const jvmFlags = (srv?.java_args !== undefined && srv?.java_args !== null) ? srv.java_args : '-XX:+UseG1GC';
    const jarName = serverExecutable(srv);
    const javaCmd = isCustom ? (rawJava || 'java') : `java (OpenJDK ${curJava})`;
    const editName = (S.activeServerId === srv?.id && S.editSrvName !== '') ? S.editSrvName : (srv?.name || '');
    const editDesc = (S.activeServerId === srv?.id && typeof S.editSrvDesc === 'string') ? S.editSrvDesc : (srv?.description || '');
    const curV = srv?.version || (native ? 'latest' : '1.21.11');
    const allV = getAllMcVersions(srv?.type || S.core || '');
    const verList = (curV && !allV.includes(curV)) ? [curV, ...allV] : allV;
    const saveStatus = _settingsUiState?.status || 'saved';
    const saveText = _settingsUiState?.text || '';
    const node = getNodeById(srv?.node_id || 'local');
    const isAdmin = S.user?.role === 'admin';
    const isOwner = isAdmin || (srv?.owner_id && S.user?.id && srv.owner_id === S.user.id);
    const isRamLocked = Boolean(node && node.allow_user_ram === false && !isAdmin);
    const isPortsLocked = Boolean(node && node.allow_user_ports === false && !isAdmin);
    const isJavaLocked = Boolean(node && node.allow_user_java === false && !isAdmin);

    return `${srvh()}
    <div style="margin-top: 14px; display: flex; flex-direction: column; gap: 16px;">
      ${(!isAdmin && !isOwner) ? `
      <div class="banner info" style="margin-bottom:0; display:flex; align-items:center; gap:10px; padding:12px 16px; border-radius:8px; background:rgba(99,102,241,0.08); border:1px solid rgba(99,102,241,0.25);">
        ${ic('users', 18)}
        <div style="font-size:13px; line-height:1.45;">
          <b>Team access:</b> Your assigned permissions determine which pages and controls are available. Viewing settings does not grant permission to change them.
        </div>
      </div>` : ''}
      <!-- Top Row: Panel Identity & Active Startup Command / Java Runtime -->
      <div class="stp-top-grid">
        <div class="card" style="padding: 20px;">
          <div class="row sb" style="margin-bottom: 12px; gap: 10px; flex-wrap: wrap; align-items: center;">
            <h2 style="font-size: 16px; margin: 0;">Server Identity &amp; Startup Command</h2>
            <span id="set-autosave-pill" class="settings-autosave-pill is-${esc(saveStatus)}" aria-live="polite" ${saveText ? '' : 'style="display:none;"'}>
              <span class="settings-autosave-dot"></span>
              <span class="settings-autosave-text">${esc(saveText)}</span>
            </span>
          </div>

          <!-- Server Icon Preview & Quick Configs -->
          <div class="row" style="gap: 16px; align-items: center; margin-bottom: 14px; padding-bottom: 12px; border-bottom: 1px solid var(--border);">
            <div style="position: relative; width: 64px; height: 64px; border-radius: 12px; overflow: hidden; background: rgba(0,0,0,0.3); border: 1px solid var(--border); display: flex; align-items: center; justify-content: center; flex-shrink: 0;">
              <img id="set-srv-icon-img" src="${srv?.icon_url ? esc(srv.icon_url) : esc(getBrand('default_server_icon_url') || '/static/img/default-server-icon.png')}" onerror="this.src=getBrand('default_server_icon_url') || '/static/img/default-server-icon.png'" style="width: 100%; height: 100%; object-fit: cover; image-rendering: pixelated;" alt="Server Icon">
            </div>
            <div style="flex: 1; min-width: 180px;">
              <div style="font-weight: 600; font-size: 13.5px; margin-bottom: 3px; color: var(--tx);">Server Icon &amp; Quick Configs</div>
              <div class="mu" style="font-size: 12px; margin-bottom: 8px;">64×64 PNG for Minecraft server list &amp; instant gameplay config editor</div>
              <div class="row" style="gap: 8px; flex-wrap: wrap;">
                <label class="btn sm" style="cursor: pointer; margin: 0;">
                  ${ic('upload', 12)} Upload Icon
                  <input type="file" accept="image/png,image/jpeg,image/webp" style="display: none;" onchange="if(this.files[0]) uploadServerIcon('${esc(srv?.id || '')}', this.files[0])">
                </label>
                ${srv?.icon_url ? `<button type="button" class="btn sm dg" onclick="deleteServerIcon('${esc(srv?.id || '')}')">${ic('trash', 12)} Remove</button>` : ''}
                ${native ? `<button type="button" class="btn sm" onclick="go('fi')">Edit config/config.toml in File Manager</button>` : `<button type="button" class="btn sm" onclick="openQuickConfigsModal('${esc(srv?.id || '')}')">${ic('sliders', 12)} Quick Configs</button>`}
              </div>
            </div>
          </div>

          <div class="grid g2" style="gap: 10px; margin-bottom: 12px;">
            <label class="f" style="margin-bottom: 0;">
              <span>Server Panel Name</span>
              <input type="text" id="set-srv-name" value="${esc(editName)}" placeholder="e.g. Survival SMP" oninput="onServerIdentityInput()" onblur="flushSettingsAutoSave()">
            </label>
            <label class="f" style="margin-bottom: 0;">
              <span>Description / Tagline</span>
              <input type="text" id="set-srv-desc" value="${esc(editDesc)}" placeholder="Optional short panel note" oninput="onServerIdentityInput()" onblur="flushSettingsAutoSave()">
            </label>
          </div>
          <div class="mu" style="font-size: 11.5px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.04em; margin-bottom: 6px;">Active Startup Command</div>
          <div id="set-cmd-preview" class="mono" style="background: var(--con-bg); padding: 12px 14px; border-radius: 8px; border: 1px solid var(--border); color: #f0f3f6; font-size: 12.5px; line-height: 1.55; word-break: break-all;">${native ? esc(nativeCommandPreview(jarName)) : `${esc(javaCmd)} -Xms${memMin}M -Xmx${curMem}M ${jvmFlags ? esc(jvmFlags) + ' ' : ''}-jar ${esc(jarName)} nogui`}</div>
          <span class="mu" style="font-size: 11.5px; margin-top: 8px; display: block;">Auto-restart automatically times out after 2 consecutive startup failures.</span>
        </div>

        ${native ? '' : `
        <div class="card" style="padding: 20px;">
          <h2 style="font-size: 16px; margin-bottom: 12px;">Java version</h2>
          <label class="f" style="margin-bottom: 8px;">
            <span class="row sb">
              <span>Java Runtime Version</span>
              ${isJavaLocked ? '<span class="chip" style="font-size:10px; color:#f59e0b; border-color:rgba(245,158,11,0.3);">🔒 Locked by Node Policy</span>' : ''}
            </span>
            <select id="prop-java" ${isJavaLocked ? 'disabled title="Java runtime changes are restricted by node policy."' : 'onchange="onPropJavaChange(this.value)"'}>
              ${renderJavaSelectOptions(rawJava, true, srv?.version, srv?.type, srv?.node_id || 'local')}
            </select>
          </label>
          <div id="prop-custom-wrap" style="${isCustom ? '' : 'display: none;'} margin-bottom: 8px;">
            <input type="text" id="prop-java-custom" value="${esc(isCustom ? rawJava : '')}" placeholder="/usr/lib/jvm/java-25-temurin/bin/java" style="font-size: 13px;" ${isJavaLocked ? 'disabled' : ''} oninput="updateSettingsCmdPreview(); scheduleSettingsAutoSave(600);" onblur="flushSettingsAutoSave()">
          </div>
          ${renderJavaAvailabilityStrip(srv?.node_id || 'local')}
          <span class="mu" style="font-size: 12px; line-height: 1.5; display: block; margin-top: 8px;">Choose Java 25, 21, 17, 8, or a custom JVM binary path for this server.</span>
        </div>`}
      </div>

      <!-- Main Unified Configuration & Startup Grid -->
      <div class="grid g2">
        ${native ? steelConfigNotice() : `
        <div class="card">
          <h2>Server Properties (<code>server.properties</code>)</h2>
          <label class="f">
            <span>Server Message (MOTD)</span>
            <input type="text" id="prop-motd" value="${esc(p.motd)}" oninput="onMotdInput(this.value)" onblur="flushSettingsAutoSave()" placeholder="e.g. &lt;gradient:#00f5a0:#00d9f5&gt;&lt;bold&gt;My Server&lt;/bold&gt;&lt;/gradient&gt; or §bWelcome!">
          </label>
          <div class="motd-preview-box">
            <div class="row sb" style="margin-bottom:6px; gap:8px;">
              <span class="mu" style="font-size:11px; font-weight:600; text-transform:uppercase; letter-spacing:0.4px;">Live MOTD Preview</span>
              <span class="motd-hints mu">&amp;/&sect; &bull; &#38;#RRGGBB &bull; &lt;gradient:...&gt; &bull; &lt;rainbow&gt; &bull; \n</span>
            </div>
            <div id="motd-preview">${renderMinecraftTextHtml(p.motd, { allowNewlines: true })}</div>
          </div>
          <div class="grid g2">
            <label class="f">
              <span class="row sb">
                <span>Server Game Port</span>
                ${isPortsLocked ? '<span class="chip" style="font-size:10px; color:#f59e0b; border-color:rgba(245,158,11,0.3);">🔒 Locked</span>' : ''}
              </span>
              <input type="number" min="1024" max="65535" id="prop-port" value="${srv?.port || 25565}" ${isPortsLocked ? 'disabled title="Port changes are restricted by node policy."' : 'oninput="scheduleSettingsAutoSave(600)" onblur="flushSettingsAutoSave()"'} >
            </label>
            <label class="f">
              <span>Max Players</span>
              <input type="number" min="1" max="100000" id="prop-max" value="${p.max}" oninput="onSettingsMaxPlayersInput(this.value)" onblur="flushSettingsAutoSave()">
            </label>
          </div>
          <div class="grid g2">
            <label class="f">
              <span>Game Difficulty</span>
              <select id="prop-diff" onchange="S.props.diff=this.value; saveSettingsAutoNow();">${o(p.diff, ['peaceful', 'easy', 'normal', 'hard'])}</select>
            </label>
            <label class="f">
              <span>Default Gamemode</span>
              <select id="prop-mode" onchange="S.props.mode=this.value; saveSettingsAutoNow();">${o(p.mode, ['survival', 'creative', 'adventure'])}</select>
            </label>
          </div>
          <h3 style="font-size: 14px; font-weight: 700; margin: 14px 0 6px; color: var(--tx);">Game Rules &amp; Protection</h3>
          ${sw('pvp', 'Player versus Player (PvP)')}
          ${sw('wl', 'Enforce Whitelist')}
          ${sw('cmd', 'Enable Command Blocks')}
        </div>`}

        <div class="card">
          <h2>Startup Variables &amp; Allocation</h2>
          <div class="grid g2">
            <label class="f">
              <span class="row sb">
                <span>${native ? 'Memory Resource Limit (systemd only)' : 'Allocated Memory (-Xmx)'}</span>
                ${isRamLocked ? '<span class="chip" style="font-size:10px; color:#f59e0b; border-color:rgba(245,158,11,0.3);">🔒 Locked by Node Policy</span>' : ''}
              </span>
              ${isRamLocked ? `
                <select id="prop-ram" disabled title="Memory allocation is managed by cluster administrators on this node.">
                  <option value="${curMem}" selected>${curMem >= 1024 ? (curMem/1024) + ' GB' : curMem + ' MB'} (${curMem} MB)</option>
                </select>
              ` : `
                <select id="prop-ram" onchange="updateSettingsCmdPreview(); saveSettingsAutoNow();">
                  <option value="1024" ${curMem === 1024 ? 'selected' : ''}>1 GB (1024 MB)</option>
                  <option value="2048" ${curMem === 2048 ? 'selected' : ''}>2 GB (2048 MB)</option>
                  <option value="4096" ${curMem === 4096 ? 'selected' : ''}>4 GB (4096 MB)</option>
                  <option value="6144" ${curMem === 6144 ? 'selected' : ''}>6 GB (6144 MB)</option>
                  <option value="8192" ${curMem === 8192 ? 'selected' : ''}>8 GB (8192 MB)</option>
                  <option value="16384" ${curMem === 16384 ? 'selected' : ''}>16 GB (16384 MB)</option>
                </select>
              `}
            </label>
            <label class="f">
              <span>${native ? 'Server Executable' : 'Server Executable JAR'}</span>
              <input type="text" id="prop-jar" value="${esc(jarName)}" placeholder="${native ? 'steel or steel.exe' : 'server.jar'}" oninput="updateSettingsCmdPreview(); scheduleSettingsAutoSave(600);" onblur="flushSettingsAutoSave()">
            </label>
          </div>
          <div class="grid g2" style="margin-bottom: 12px;">
            <label class="f" style="margin-bottom: 0;">
              <span class="row sb">
                <span>CPU Limit (%)</span>
                <span class="mu" style="font-size: 11px;">0% = Unlimited</span>
              </span>
              <input type="number" id="prop-cpu-limit" value="${srv?.cpu_limit_percent || 0}" min="0" max="3200" step="10" placeholder="0 for unlimited" ${!isAdmin && !isOwner ? 'disabled' : 'oninput="scheduleSettingsAutoSave(600)" onblur="flushSettingsAutoSave()"'} >
            </label>
            <label class="f" style="margin-bottom: 0;">
              <span class="row sb">
                <span>Storage / Disk Limit (MB)</span>
                <span class="mu" style="font-size: 11px;">0 = Unlimited</span>
              </span>
              <input type="number" id="prop-disk-limit" value="${srv?.disk_limit_mb || 0}" min="0" step="512" placeholder="0 for unlimited" ${!isAdmin && !isOwner ? 'disabled' : 'oninput="scheduleSettingsAutoSave(600)" onblur="flushSettingsAutoSave()"'} >
            </label>
          </div>
          <div class="f" style="margin-bottom: 12px;">
            <span style="display:block; margin-bottom:6px; font-size:12.5px; color:var(--tx-muted); font-weight:600;">Minecraft / Release Version</span>
            <div style="display: grid; grid-template-columns: minmax(135px, 165px) 1fr; gap: 8px;">
              <select id="prop-ver-sel" data-searchable="1" onchange="onPropVersionSelectChange(this.value)">
                ${verList.map(v => `<option value="${esc(v)}" ${v === curV ? 'selected' : ''}>${esc(v)}</option>`).join('')}
              </select>
              <input type="text" id="prop-ver" value="${esc(curV)}" style="font-weight: 600;" oninput="onPropVersionInput(this.value)" onblur="flushSettingsAutoSave()">
            </div>
          </div>
          ${native ? '' : `
          <label class="f">
            <span>Custom JVM Flags (JAVA_ARGS)</span>
            <input type="text" id="prop-args" value="${esc(jvmFlags)}" placeholder="-XX:+UseG1GC" oninput="updateSettingsCmdPreview(); scheduleSettingsAutoSave(600);" onblur="flushSettingsAutoSave()">
          </label>`}
          <label class="row sb" style="padding: 10px 0; border-top: 1px solid var(--ln); border-bottom: 1px solid var(--ln); margin-bottom: 14px;">
            <div>
              <span style="display:block; font-weight: 600; color: var(--tx);">Auto-Restart Crashed Server</span>
              <span class="mu" style="font-size: 11.5px;">Halts automatically after 2 consecutive startup failures</span>
            </div>
            <span class="sw">
              <input type="checkbox" id="prop-autorestart" ${srv?.auto_restart !== false ? 'checked' : ''} onchange="saveSettingsAutoNow()">
              <i></i>
            </span>
          </label>
          <div id="set-autosave-status" class="settings-autosave-footer is-${esc(saveStatus)}" aria-live="polite" ${saveText ? '' : 'style="display:none;"'}>
            <span class="settings-autosave-dot"></span>
            <span class="settings-autosave-text">${esc(saveText)}</span>
          </div>
        </div>
      </div>

      ${isAdmin ? `
      <!-- Server Ownership & Tenant Assignment (Admin Only) -->
      <div class="card" style="padding: 22px;">
        <div class="row sb" style="align-items: center; flex-wrap: wrap; gap: 14px;">
          <div>
            <h2 style="font-size: 16px; margin: 0 0 4px; display: flex; align-items: center; gap: 8px; color: var(--tx);">
              ${ic('user', 16)} Server Ownership &amp; Tenant Assignment
            </h2>
            <p class="mu" style="font-size: 13px; margin: 0; max-width: 620px;">
              Reassign ownership of this server instance to any registered panel account. The assigned owner receives full management rights for this server.
            </p>
          </div>
          <div class="row" style="gap: 10px; align-items: center; flex-wrap: wrap;">
            <label class="f" style="margin-bottom:0; min-width: 220px;">
              <select id="prop-owner-id" onchange="saveSettingsAutoNow()">
                ${(S.adminUsers || []).map(u => `<option value="${u.id}" ${u.id === (srv?.owner_id || 1) ? 'selected' : ''}>${esc(u.username)}${(u.first_name || u.last_name) ? ` (${esc([u.first_name, u.last_name].filter(Boolean).join(' '))})` : ''} [${esc(u.role)}]</option>`).join('')}
              </select>
            </label>
          </div>
        </div>
      </div>` : ''}


      <p class="mu">Networking and file transfer have dedicated Network and SFTP sections in the sidebar.</p>

      ${(isAdmin || isOwner) ? `
      <!-- Reinstall Server Section -->
      <div class="card" style="border: 1px solid var(--border-light); padding: 22px;">
        <div class="row sb" style="align-items: center; flex-wrap: wrap; gap: 16px;">
          <div>
            <h2 style="font-size: 16px; margin: 0 0 6px; display: flex; align-items: center; gap: 8px; color: var(--tx);">
              ${ic('refresh', 16)} Reinstall Server Instance
            </h2>
            <p class="mu" style="font-size: 13px; margin: 0; max-width: 620px;">
              Re-download a fresh server executable from official upstream releases (Steel) or mcjarfiles.com (Java engines). Back up your files before switching engines; preserved worlds, configs and plugins are not converted.
            </p>
          </div>
          <button class="btn" data-srv-id="${esc(srv?.id || '')}" data-node-id="${esc(srv?.node_id || 'local')}" onclick="openReinstallModal(this.dataset.srvId, this.dataset.nodeId)" style="height: 38px; padding: 0 18px; font-weight: 600;">
            ${ic('refresh', 14)} Reinstall Server
          </button>
        </div>
      </div>

      <!-- Danger Zone: Delete Server -->
      <div class="card" style="border: 1px solid rgba(239, 68, 68, 0.35); background: rgba(239, 68, 68, 0.05); padding: 22px;">
        <div class="row sb" style="align-items: center; flex-wrap: wrap; gap: 16px;">
          <div>
            <h2 style="color: #ef4444; font-size: 16px; margin: 0 0 6px; display: flex; align-items: center; gap: 8px;">
              ${ic('trash', 16)} Danger Zone: Delete Server Instance
            </h2>
            <p class="mu" style="font-size: 13px; margin: 0; max-width: 620px;">
              Permanently destroy this Minecraft server instance (<b>${esc(srv?.name || '')}</b>). All files, worlds, player data, plugins/mods, configs, and allocated ports will be erased. This action cannot be reversed.
            </p>
          </div>
          <button class="btn dg" data-srv-id="${esc(srv?.id || '')}" data-srv-name="${esc(srv?.name || '')}" data-node-id="${esc(srv?.node_id || 'local')}" onclick="openDeleteServerModal(this.dataset.srvId, this.dataset.srvName, this.dataset.nodeId)" style="height: 38px; padding: 0 18px; font-weight: 600;">
            ${ic('trash', 14)} Delete Server
          </button>
        </div>
      </div>` : ''}
      <div class="np-slot-container" data-np-slot="settings:bottom"></div>
    </div>`;
  },

  act() {
    const header = curServer()
      ? srvh()
      : `<div class="card" style="padding: 18px 22px;"><h1 style="margin:0; font-size:20px;">Panel Activity Log</h1></div>`;
    return `${header}
    <div style="margin-top: 14px;">
      ${renderActivityLogSection({
        scope: 'all',
        title: 'Server & Panel Activity',
        subtitle: 'Real-time audit trail of power actions, console commands, files, plugins, backups & logins',
        showActor: true
      })}
    </div>`;
  },

  // Dedicated Admin Area View
  admin() {
    const sys = S.system || {};
    const usedRam = Math.max(0, Number(sys.total_ram_mb || 0) - Number(sys.free_ram_mb || 0));
    const ramPct = Math.round((usedRam / (sys.total_ram_mb || 1)) * 100);
    const t = S.adminTab || 0;
    const nodesList = (Array.isArray(S.nodes) && S.nodes.length > 0) ? S.nodes : [{
      id: 'local',
      name: 'Local node',
      location: 'This machine',
      fqdn: S.brandingDaemon?.fqdn || window.location.hostname,
      is_local: true,
      online: Boolean(S.system),
      latency_ms: 0,
      server_count: S.servers.filter(x => (x.node_id || 'local') === 'local').length,
      system: sys,
      token_prefix: 'local'
    }];
    const onlineNodesCount = nodesList.filter(n => n.online).length;

    const draft = S.brandDraft || { ...DEFAULT_BRANDING, ...S.branding };
    const up = S.adminUpdate || {};
    const daemonDraft = S.daemonDraft || {
      fqdn: S.brandingDaemon?.fqdn || '',
      sftp_port: S.brandingDaemon?.sftp_port || S.adminDaemonCfg?.sftpPort || 2022,
      max_upload_mb: S.brandingDaemon?.max_upload_mb || S.adminDaemonCfg?.maxUploadMB || 500,
      default_ram_mb: S.brandingDaemon?.default_ram_mb || S.adminDaemonCfg?.defaultRamMB || 2048,
      auto_restart_policy: S.brandingDaemon?.auto_restart_policy !== undefined ? Boolean(S.brandingDaemon.auto_restart_policy) : true
    };

    const section = ADMIN_SECTIONS[t] || ADMIN_SECTIONS[0];
    const runningServers = S.servers.filter(server => ['online', 'starting'].includes(normalizeServerStatus(server.stats?.status || server.status))).length;
    const updateStatus = (up.is_updating || S.updateProgress?.running) ? 'Updating' : up.error ? 'Check failed' : up.update_available === true ? 'Update available' : up.update_available === false ? 'Up to date' : 'Not checked';

    let content = '';

    if (t === 0) {
      content = `
        <div class="grid g4" style="margin-bottom:18px;">
          <div class="card stat">
            <span>Operating system</span>
            <b style="font-size:18px; padding:6px 0;">${sys.os ? `${esc(sys.os)} (${esc(sys.arch)})` : 'Not loaded'}</b>
            <span class="mu">${sys.cpus ?? '—'} CPU cores &bull; ${esc(sys.go_version || 'Go')}</span>
          </div>
          <div class="card stat">
            <span>Memory usage</span>
            <b style="font-size:18px; padding:6px 0;">${sys.total_ram_mb ? `${(usedRam/1024).toFixed(1)} / ${(sys.total_ram_mb/1024).toFixed(1)} GB` : '—'}</b>
            <div class="bar"><i style="width:${ramPct}%"></i></div>
          </div>
          <div class="card stat">
            <span>Nodes</span>
            <b style="font-size:18px; padding:6px 0;">${onlineNodesCount} / ${nodesList.length} online</b>
            <span class="mu"><a href="/admin/nodes" onclick="return handleNavClick(event, 'admin', { adminTab: 1 })" style="color:var(--ac);text-decoration:none;">Manage nodes &rarr;</a></span>
          </div>
          <div class="card stat">
            <span>SFTP</span>
            <b style="font-size:18px; padding:6px 0;">Port ${daemonDraft.sftp_port || 2022}</b>
            <span class="mu">Configured file-transfer port</span>
          </div>
        </div>

        <div class="card tw np-admin-info-card" style="margin-bottom:18px;">
          <div class="row sb" style="margin-bottom:14px; flex-wrap:wrap; gap:12px;">
            <div>
              <h2 style="margin:0 0 4px; color:var(--tx);">Panel memory</h2>
              <span class="mu" style="font-size:13px;">Clean up unused panel memory and view system details. This does not change Minecraft-server memory.</span>
            </div>
            <button class="btn pr" onclick="triggerAdminGC()">${ic('refresh', 14)} Clean up unused memory</button>
          </div>
          <div style="background:rgba(255,255,255,0.02); border:1px solid var(--border); border-radius:8px; padding:16px;">
            <div class="row sb" style="padding:6px 0; border-bottom:1px solid var(--border);">
              <span class="mu">Panel version</span>
              <span class="mono" style="color:var(--tx);">${sys.version ? `v${esc(sys.version)}` : '—'}</span>
            </div>
            <div class="row sb" style="padding:6px 0; border-bottom:1px solid var(--border);">
              <span class="mu">Go version</span>
              <span class="mono" style="color:var(--tx);">${esc(sys.go_version || '—')}</span>
            </div>
            <div class="row sb" style="padding:6px 0; border-bottom:1px solid var(--border);">
              <span class="mu" title="Active Go routines in the panel process">Go routines</span>
              <span class="mono" style="color:var(--tx);">${sys.goroutines ?? '—'}</span>
            </div>
            <div class="row sb" style="padding:6px 0;">
              <span class="mu">Platform</span>
              <span class="mono" style="color:var(--tx);">${sys.os ? `${esc(sys.arch)} on ${esc(sys.os)}` : '—'}</span>
            </div>
          </div>
        </div>

        <div class="card tw np-admin-info-card" style="margin-bottom:18px;">
          <div class="row sb" style="margin-bottom:14px; flex-wrap:wrap; gap:12px;">
            <div>
              <h2 style="margin:0 0 4px; color:var(--tx); display:flex; align-items:center; gap:8px;">
                ${ic('cpu', 18)} Panel updates
              </h2>
              <span class="mu" style="font-size:13px;">Check for updates and choose which repository and branch supply them.</span>
            </div>
            <div class="row" style="gap:8px; align-items:center;" id="admin-update-actions-wrap">
              <button type="button" class="btn sm" onclick="checkAdminUpdates()">${ic('refresh', 13)} Check for updates</button>
              ${renderAdminUpdateBtn(up)}
            </div>
          </div>

          <div style="background:rgba(255,255,255,0.02); border:1px solid var(--border); border-radius:8px; padding:16px; margin-bottom:16px;">
            <div class="row sb" style="padding:6px 0; border-bottom:1px solid var(--border); flex-wrap:wrap; gap:8px;">
              <span class="mu">Installed version</span>
              <div class="row" style="gap:8px; align-items:center;">
                <span class="mono" style="color:var(--tx);">${sys.version ? `v${esc(sys.version)}` : '—'}</span>
                ${up?.current_commit ? `<span class="chip mono" style="font-size:11px;">#${esc(up.current_commit.slice(0, 7))}</span>` : ''}
              </div>
            </div>
            <div class="row sb" style="padding:6px 0; border-bottom:1px solid var(--border); flex-wrap:wrap; gap:8px;">
              <span class="mu">Update status</span>
              <div class="row" style="gap:8px; align-items:center;">
                ${(up?.is_updating || S.updateProgress?.running) ? '<span class="badge starting"><i class="dot"></i>Updating</span>' :
                  up?.error ? `<span class="badge offline" title="${esc(up.error)}"><i class="dot"></i>Check failed: ${esc(up.error)}</span>` :
                  up?.update_available === true ? `<span class="badge starting"><i class="dot"></i>Update available (${esc(up.remote_commit?.slice(0, 7) || 'latest')})</span>` :
                  up?.update_available === false ? '<span class="badge online"><i class="dot"></i>Up to date</span>' : '<span class="chip">Not checked</span>'}
              </div>
            </div>
            ${up?.remote_message ? `
            <div class="row sb" style="padding:6px 0; border-bottom:1px solid var(--border); flex-wrap:wrap; gap:8px;">
              <span class="mu">Latest change</span>
              <div class="row" style="gap:8px; align-items:center;">
                <span class="mono mu" style="font-size:11px;">#${esc(up.remote_commit?.slice(0, 7) || '')}</span>
                <span style="color:var(--tx); font-size:12px;">${esc(up.remote_message)}</span>
              </div>
            </div>` : ''}
            <div class="row sb" style="padding:6px 0; flex-wrap:wrap; gap:8px;">
              <span class="mu">Update logs</span>
              <button type="button" class="btn sm" onclick="openUpdateLogsModal()">${ic('terminal', 12)} View update logs</button>
            </div>
          </div>

          <div style="background:rgba(255,255,255,0.015); border:1px solid var(--border); border-radius:8px; padding:16px;">
            <h3 style="margin:0 0 8px; font-size:14px; color:var(--tx);">Update source</h3>
            <p class="mu" style="font-size:12px; margin:0 0 12px;">Point NeoPanel to your own GitHub fork or a custom development branch for automated builds.</p>
            <div class="grid g2" style="gap:12px; margin-bottom:12px;">
              <div>
                <label class="lbl" style="font-size:12px;">Git Repository URL</label>
                <input type="text" id="admin-update-repo" class="inp mono" value="${esc(up?.repo_url || 'https://github.com/Pekly/NeoPanel.git')}" placeholder="https://github.com/Pekly/NeoPanel.git">
              </div>
              <div>
                <label class="lbl" style="font-size:12px;">Git Branch</label>
                <input type="text" id="admin-update-branch" class="inp mono" value="${esc(up?.branch || 'main')}" placeholder="main">
              </div>
            </div>
            <div class="row" style="justify-content:flex-end; gap:8px;">
              <button type="button" class="btn sm" onclick="resetAdminUpdateRepo()">${ic('refresh', 12)} Reset to default</button>
              <button type="button" class="btn sm pr" onclick="saveAdminUpdateConfig()">${ic('save', 12)} Save update source</button>
            </div>
          </div>
        </div>

        <div class="card tw np-admin-info-card" style="margin-bottom:18px;">
          <div class="row sb" style="margin-bottom:10px; flex-wrap:wrap; gap:12px;">
            <div>
              <h2 style="margin:0 0 4px; color:var(--tx);">Java versions (25, 21, 17 &amp; 8)</h2>
              <span class="mu" style="font-size:13px;">Monitor detected OpenJDK / Eclipse Temurin runtimes on the master host and install missing premade versions.</span>
            </div>
            <button class="btn sm" onclick="loadJavaRuntimes(true, 'local')">${ic('refresh', 13)} Refresh</button>
          </div>
          ${renderJavaAvailabilityStrip('local')}
        </div>
      `;
    } else if (t === 1) {
      const h = S.nodePairHelper || {};
      const cmdPreview = buildClusterNodeInstallCmd();
      const nodeVm = getViewMode('admin_nodes', 'tiles');
      content = `
        <div class="node-callout">
          <div class="row sb" style="gap: 12px; flex-wrap: wrap; align-items: center;">
            <div style="flex: 1 1 380px;">
              <b>Add another machine:</b> Install NeoPanel on an Ubuntu host with <code>--node</code> and connect it using an admin API key, or pair an existing node with <code>sudo neopanel node-token</code> / <code>sudo neopanel node connect</code>.
            </div>
            <button type="button" class="btn sm ${S.nodePairOpen ? 'pr' : ''}" onclick="toggleClusterNodeHelper()">
              ${ic('terminal', 13)} ${S.nodePairOpen ? 'Hide Installer Command' : 'Node Install Command'}
            </button>
          </div>
        </div>

        ${S.nodePairOpen ? `
        <div class="card node-cli-card" style="margin-bottom: 16px;">
          <div class="row sb" style="margin-bottom: 10px; gap: 10px; flex-wrap: wrap;">
            <div>
              <h3 style="margin: 0 0 3px; font-size: 15px; color: var(--tx);">Install and connect a node</h3>
              <span class="mu" style="font-size: 12.5px;">Installs a node with SFTP and connects it to this panel.</span>
            </div>
            <button type="button" class="btn sm pr" onclick="generateNodePairingKey(this)">
              ${ic('key', 13)} Generate API Key
            </button>
          </div>
          <div class="grid g3" style="gap: 10px; margin-bottom: 12px;">
            <label class="f" style="margin: 0;">
              <span>Node Name</span>
              <input type="text" id="nh-name" value="${esc(h.name || 'mc-node-01')}" oninput="onNodeHelperInput('name', this.value)" placeholder="mc-node-01">
            </label>
            <label class="f" style="margin: 0;">
              <span>Location (optional)</span>
              <input type="text" id="nh-loc" value="${esc(h.location || '')}" oninput="onNodeHelperInput('location', this.value)" placeholder="Frankfurt, DE">
            </label>
            <label class="f" style="margin: 0;">
              <span>Node port</span>
              <input type="number" id="nh-port" value="${esc(h.port || '8080')}" min="1024" max="65535" oninput="onNodeHelperInput('port', this.value)" placeholder="8080">
            </label>
          </div>
          <label class="f" style="margin-bottom: 10px;">
            <span>Admin API Key (<code>neo_...</code>)</span>
            <input type="text" id="nh-key" class="mono" value="${esc(h.apiKey || '')}" oninput="onNodeHelperInput('apiKey', this.value)" placeholder="Click 'Generate API Key' above or paste an existing neo_... key">
          </label>
          <div class="node-cmd-box">
            <code id="node-install-cmd-text">${esc(cmdPreview)}</code>
            <button type="button" id="node-install-copy-btn" class="btn sm" onclick="copyDbText($('#node-install-cmd-text')?.textContent || '', 'Node install command')">
              ${ic('copy', 13)} Copy
            </button>
          </div>
          <div class="mu" style="font-size: 11.5px; margin-top: 8px; line-height: 1.45;">
            Prefer manual pairing on Master SSH without creating an API key? Run <code>sudo neopanel node-token</code> on the remote node, then run <code>sudo neopanel node connect --name "${esc(h.name || 'mc-node-01')}" --url "http://&lt;NODE_IP&gt;:${esc(h.port || '8080')}" --token "&lt;neo_node_...&gt;"</code> here.
          </div>
        </div>
        ` : ''}

        <div class="row sb" style="margin-bottom: 14px; flex-wrap: wrap; gap: 10px;">
          <div class="row" style="gap: 8px;">
            <h2 style="margin: 0;">Connected Nodes</h2>
            <span class="chip">${onlineNodesCount} online &bull; ${nodesList.length} total</span>
          </div>
          <div class="row" style="gap: 8px; flex-wrap: wrap;">
            ${renderViewModeToggle('admin_nodes', 'tiles')}
            <button type="button" class="btn sm" onclick="loadAdminNodes(true)">${ic('refresh', 13)} Refresh Nodes</button>
            <button type="button" class="btn sm pr" onclick="openNodeModal()">${ic('plus', 14)} Connect Node</button>
          </div>
        </div>

        ${nodeVm === 'list' ? `
        <div class="card tw">
          <table>
            <thead><tr><th>Node</th><th>Location</th><th>Endpoint</th><th>Hardware</th><th>Servers</th><th>Status</th><th style="text-align:right">Actions</th></tr></thead>
            <tbody>
              ${nodesList.map(n => {
                const sysInfo = n.system || {};
                const totalGb = sysInfo.total_ram_mb ? (sysInfo.total_ram_mb / 1024).toFixed(1) + ' GB' : '—';
                return `<tr>
                  <td>
                    <div class="row" style="gap:8px;">
                      <span class="node-status-dot ${n.online ? 'online' : 'offline'}"></span>
                      <b>${esc(displayNodeName(n.id, n.name))}</b>
                      ${n.is_local ? '<span class="chip" style="font-size:10px;">Local</span>' : ''}
                    </div>
                  </td>
                  <td class="mu">${esc(displayNodeLocation(n))}</td>
                  <td class="mono">${esc(n.is_local ? 'Internal Direct' : (n.base_url || n.fqdn))}</td>
                  <td class="mono">${sysInfo.cpus ? `${sysInfo.cpus} vCPU &bull; ${totalGb}` : '—'}</td>
                  <td>${Number(n.server_count) || 0}</td>
                  <td><span class="badge ${n.online ? 'online' : 'offline'}"><i class="dot"></i>${n.online ? `Online${n.is_local ? '' : ` (${n.latency_ms ?? 0}ms)`}` : 'Offline'}</span></td>
                  <td style="text-align:right">
                    <div class="row" style="justify-content:flex-end; gap:6px;">
                      <button type="button" class="btn sm" onclick="testClusterNode('${esc(n.id)}', this)">${ic('refresh', 12)} Test</button>
                      ${n.is_local
                        ? `<button type="button" class="btn sm" onclick="selectAdminTab(4)">Configure</button>`
                        : `<button type="button" class="btn sm" onclick="openNodeModal('${esc(n.id)}')">${ic('edit', 12)} Edit</button>
                           <button type="button" class="btn sm dg" data-id="${esc(n.id)}" data-name="${esc(n.name)}" onclick="deleteClusterNode(this.dataset.id, this.dataset.name)">${ic('trash', 12)}</button>`}
                    </div>
                  </td>
                </tr>`;
              }).join('')}
            </tbody>
          </table>
        </div>` : `
        <div class="${nodeVm === 'icons' ? 'vm-icons-grid wide' : 'node-grid'}">
          ${nodesList.map(n => {
            const sysInfo = n.system || {};
            const totalGb = sysInfo.total_ram_mb ? (sysInfo.total_ram_mb / 1024).toFixed(1) + ' GB' : '—';
            const usedNodeRam = sysInfo.total_ram_mb && sysInfo.free_ram_mb !== undefined
              ? ((sysInfo.total_ram_mb - sysInfo.free_ram_mb) / 1024).toFixed(1) + ' GB'
              : '—';
            return `<div class="card node-card">
              <div class="row sb" style="align-items: flex-start; gap: 10px;">
                <div style="min-width: 0;">
                  <div class="row" style="gap: 8px; align-items: center;">
                    <span class="node-status-dot ${n.online ? 'online' : 'offline'}"></span>
                    <b style="font-size: 16px; color: var(--tx); overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">${esc(displayNodeName(n.id, n.name))}</b>
                    ${n.is_local ? '<span class="chip" style="font-size:10.5px;">This machine</span>' : ''}
                  </div>
                  <div class="mu mono" style="font-size: 11.5px; margin-top: 3px;">
                    ${esc(displayNodeLocation(n))} &bull; ${esc(n.fqdn || window.location.hostname)}
                  </div>
                </div>
                <span class="badge ${n.online ? 'online' : 'offline'}">
                  <i class="dot"></i>${n.online ? `Online${n.is_local ? '' : ` (${n.latency_ms ?? 0}ms)`}` : 'Offline'}
                </span>
              </div>

              <div class="node-meta-grid">
                <div class="node-meta-item">
                  <span>Servers</span>
                  <b>${Number(n.server_count) || 0} instance${Number(n.server_count) === 1 ? '' : 's'}</b>
                </div>
                <div class="node-meta-item">
                  <span>Hardware</span>
                  <b>${sysInfo.cpus ? `${sysInfo.cpus} vCPU &bull; ${totalGb}` : 'Unavailable'}</b>
                </div>
                <div class="node-meta-item">
                  <span>Memory Used</span>
                  <b>${usedNodeRam !== '—' ? `${usedNodeRam} / ${totalGb}` : '—'}</b>
                </div>
                <div class="node-meta-item">
                  <span>Endpoint</span>
                  <b class="mono" title="${esc(n.base_url || 'Local Host')}">${esc(n.is_local ? 'Internal Direct' : (n.base_url || n.fqdn))}</b>
                </div>
              </div>

              ${!n.online && n.error ? `<div class="node-error-note">${esc(n.error)}</div>` : ''}

              <div class="row sb" style="margin-top: auto; padding-top: 4px;">
                <span class="mu mono" style="font-size: 11px;">${n.is_local ? 'id: local' : `key: ${esc(n.token_prefix || 'neo_...')}`}</span>
                <div class="row" style="gap: 6px;">
                  <button type="button" class="btn sm" onclick="testClusterNode('${esc(n.id)}', this)">${ic('refresh', 12)} Test</button>
                  ${n.is_local
                    ? `<button type="button" class="btn sm" onclick="selectAdminTab(4)">Panel settings</button>`
                    : `
                      <button type="button" class="btn sm" onclick="openNodeModal('${esc(n.id)}')">${ic('edit', 12)} Edit</button>
                      <button type="button" class="btn sm dg" data-id="${esc(n.id)}" data-name="${esc(n.name)}" onclick="deleteClusterNode(this.dataset.id, this.dataset.name)">${ic('trash', 12)} Disconnect</button>
                    `}
                </div>
              </div>
            </div>`;
          }).join('')}
        </div>`}
      `;
    } else if (t === 2) {
      const admSrvVm = getViewMode('admin_servers', 'list');
      content = `
        <div class="card tw np-admin-cluster-card" style="margin-bottom:18px;">
          <div class="row sb" style="margin-bottom:14px; flex-wrap:wrap; gap:12px;">
            <div>
              <h2 style="margin:0 0 4px; color:var(--tx);">All servers (${S.servers.length})</h2>
              <span class="mu" style="font-size:13px;">Overview and controls for all server instances across local and linked remote nodes.</span>
            </div>
            <div class="row" style="gap:8px; flex-wrap:wrap;">
              ${renderViewModeToggle('admin_servers', 'list')}
              <button class="btn sm" onclick="bulkRestartServers()">${ic('refresh', 14)} Restart All</button>
              <button class="btn sm dg" onclick="bulkStopServers()">Stop All</button>
              <button class="btn sm pr" onclick="openCreateServerModal()">${ic('plus', 13)} Create server</button>
            </div>
          </div>
          ${admSrvVm === 'list' ? `
          <div class="np-table-scroll" role="region" aria-label="Cluster servers" tabindex="0"><table class="np-admin-server-table">
            <thead><tr><th>Server</th><th>Owner</th><th>Node</th><th>Port</th><th>Engine</th><th>Resource Limits</th><th>Status</th><th style="text-align:right">Actions</th></tr></thead>
            <tbody>
              ${S.servers.length ? S.servers.map((s, idx) => {
                const nid = s.node_id || 'local';
                const nodeName = displayNodeName(nid, s.node_name || getNodeById(nid)?.name);
                const normSt = normalizeServerStatus(s.stats?.status || s.status);
                const isRun = normSt === 'online' || normSt === 'starting';
                const srvUrl = serverRouteUrl(nid, s.id, 'con');
                const ownerUser = (S.adminUsers || []).find(u => u.id === s.owner_id);
                const ownerLabel = ownerUser ? ownerUser.username : (s.owner_id ? `User #${s.owner_id}` : 'Root / Admin');
                const cpuLabel = s.cpu_limit_percent > 0 ? `${s.cpu_limit_percent}% CPU` : '∞ CPU';
                const diskLabel = s.disk_limit_mb > 0 ? (s.disk_limit_mb >= 1024 ? `${(s.disk_limit_mb / 1024).toFixed(1)} GB` : `${s.disk_limit_mb} MB`) : '∞ Disk';
                return `
                <tr>
                  <td><b>${esc(s.name)}</b><span class="mono mu np-server-id" style="font-size:11px;">${esc(s.id)}</span></td>
                  <td><span class="chip" style="font-size:11px; display:inline-flex; align-items:center; gap:4px;">${ic('user', 11)} ${esc(ownerLabel)}</span></td>
                  <td><span class="chip srv-node-badge">${esc(nodeName)}</span></td>
                  <td class="mono">${s.port}</td>
                  <td><span class="chip" style="display:inline-flex; align-items:center; gap:5px;">${engineIcon(s.type || 'paper', 12)} ${esc(engineMeta(s.type || 'paper').label)} ${esc(s.version)}</span></td>
                  <td><span class="mono mu" style="font-size:11.5px;">${(getServerMaxRamMB(s)/1024).toFixed(1)}G RAM &bull; ${cpuLabel} &bull; ${diskLabel}</span></td>
                  <td><span class="badge ${isRun ? 'online' : 'offline'}"><i class="dot"></i>${isRun ? 'Online' : 'Offline'}</span></td>
                  <td style="text-align:right">
                    <div class="np-server-actions">
                      <a href="${esc(srvUrl)}" class="btn sm" style="text-decoration:none;" onclick="return handleServerCardClick(event, ${idx}, 'con')">Manage</a>
                      <button class="btn sm" data-srv-id="${esc(s.id)}" data-node-id="${esc(nid)}" onclick="openAdminServerSettingsModal(this.dataset.srvId, this.dataset.nodeId)" title="Edit Limits & Ownership">${ic('sliders', 12)} Limits &amp; Owner</button>
                      <button class="btn sm" data-srv-id="${esc(s.id)}" data-node-id="${esc(nid)}" onclick="openReinstallModal(this.dataset.srvId, this.dataset.nodeId)">${ic('refresh', 12)} Reinstall</button>
                      <button class="btn sm dg" data-srv-id="${esc(s.id)}" data-srv-name="${esc(s.name)}" data-node-id="${esc(nid)}" onclick="openDeleteServerModal(this.dataset.srvId, this.dataset.srvName, this.dataset.nodeId)">${ic('trash', 12)} Delete</button>
                    </div>
                  </td>
                </tr>`;
              }).join('') : '<tr><td colspan="8" class="mu" style="text-align:center; padding:24px;">No server instances deployed yet.</td></tr>'}
            </tbody>
          </table></div>` : `
          <div class="${admSrvVm === 'icons' ? 'vm-icons-grid wide' : 'vm-tiles-grid'}">
            ${S.servers.length ? S.servers.map((s, idx) => {
              const nid = s.node_id || 'local';
              const nodeName = displayNodeName(nid, s.node_name || getNodeById(nid)?.name);
              const normSt = normalizeServerStatus(s.stats?.status || s.status);
              const isRun = normSt === 'online' || normSt === 'starting';
              const srvUrl = serverRouteUrl(nid, s.id, 'con');
              const ownerUser = (S.adminUsers || []).find(u => u.id === s.owner_id);
              const ownerLabel = ownerUser ? ownerUser.username : (s.owner_id ? `User #${s.owner_id}` : 'Root / Admin');
              return `<div class="${admSrvVm === 'icons' ? 'vm-icon-card' : 'vm-tile-card'}">
                <div class="vm-item-head">
                  <div style="min-width:0;">
                    <b class="vm-item-title" style="display:block;">${esc(s.name)}</b>
                    <div class="mu mono" style="font-size:11px; margin-top:2px;">Owner: ${esc(ownerLabel)} &bull; Port ${s.port} &bull; ${(getServerMaxRamMB(s)/1024).toFixed(1)} GB</div>
                  </div>
                  <span class="badge ${isRun ? 'online' : 'offline'}"><i class="dot"></i>${isRun ? 'Online' : 'Offline'}</span>
                </div>
                <div class="row" style="gap:6px; flex-wrap:wrap;">
                  <span class="chip srv-node-badge">${esc(nodeName)}</span>
                  <span class="chip" style="display:inline-flex; align-items:center; gap:5px;">${engineIcon(s.type || 'paper', 12)} ${esc(engineMeta(s.type || 'paper').label)} ${esc(s.version)}</span>
                </div>
                <div class="vm-item-foot">
                  <a href="${esc(srvUrl)}" class="btn sm pr" style="text-decoration:none;" onclick="return handleServerCardClick(event, ${idx}, 'con')">Manage</a>
                  <div class="row" style="gap:6px;">
                    <button class="btn sm" data-srv-id="${esc(s.id)}" data-node-id="${esc(nid)}" onclick="openAdminServerSettingsModal(this.dataset.srvId, this.dataset.nodeId)" title="Edit Limits & Ownership">${ic('sliders', 12)}</button>
                    <button class="btn sm" data-srv-id="${esc(s.id)}" data-node-id="${esc(nid)}" onclick="openReinstallModal(this.dataset.srvId, this.dataset.nodeId)">${ic('refresh', 12)}</button>
                    <button class="btn sm dg" data-srv-id="${esc(s.id)}" data-srv-name="${esc(s.name)}" data-node-id="${esc(nid)}" onclick="openDeleteServerModal(this.dataset.srvId, this.dataset.srvName, this.dataset.nodeId)">${ic('trash', 12)}</button>
                  </div>
                </div>
              </div>`;
            }).join('') : '<div class="mu" style="grid-column:1/-1; text-align:center; padding:24px;">No server instances deployed yet.</div>'}
          </div>`}
        </div>
      `;
    } else if (t === 3) {
      content = `
        <input type="file" id="admin-user-avatar-file" accept="image/*" style="display:none;" onchange="onAdminUserAvatarSelected(this)">
        <div class="card tw" style="margin-bottom:18px;">
          <div class="row sb" style="margin-bottom:14px; flex-wrap:wrap; gap:12px;">
            <div>
              <h2 style="margin:0 0 4px; color:var(--tx);">Panel accounts</h2>
              <span class="mu" style="font-size:13px;">Manage operators, administrators, and client user accounts with independent roles and avatars.</span>
            </div>
            <button class="btn pr" onclick="openAddUserModal()">+ Create User</button>
          </div>
          <table>
            <thead><tr><th>User</th><th>Email Address</th><th>Role</th><th>Security</th><th>Created</th><th style="text-align:right">Actions</th></tr></thead>
            <tbody>
              ${(S.adminUsers || []).length ? S.adminUsers.map(u => {
                const isMe = (S.user && (u.id === S.user.id || u.username === S.user.username)) || u.username === S.acct.un;
                return `
                <tr>
                  <td>
                    <div class="row" style="gap:10px; align-items:center;">
                      ${userAvatarHtml(u, 32)}
                      <div>
                        <b>${esc(u.username)}</b>
                        ${(u.first_name || u.last_name) ? `<span class="mu" style="font-size:12px; margin-left:6px;">(${esc([u.first_name, u.last_name].filter(Boolean).join(' '))})</span>` : ''}
                        ${isMe ? '<span class="chip" style="margin-left:6px; font-size:10px;">You</span>' : ''}
                      </div>
                    </div>
                  </td>
                  <td>
                    <span class="mono" style="color:${u.email ? 'var(--tx)' : 'var(--tx-muted)'}; font-size:13px;">
                      ${esc(u.email || '—')}
                    </span>
                  </td>
                  <td><span class="badge ${u.role === 'admin' ? 'online' : ''}">${esc(u.role)}</span>${renderUserPanelRole(u)}</td>
                  <td>
                    ${u.two_factor_enabled
                      ? '<span class="badge online"><i class="dot"></i>2FA Enabled</span>'
                      : '<span class="chip" style="font-size:11px;">Standard</span>'}
                  </td>
                  <td class="mu">${u.created_at ? new Date(u.created_at).toLocaleDateString() : '—'}</td>
                  <td style="text-align:right">
                    <div class="row" style="justify-content:flex-end; gap:6px;">
                      <button type="button" class="btn sm" data-id="${u.id}" onclick="openEditUserModal(${u.id})">${ic('edit', 12)} Edit</button>
                      <button type="button" class="btn sm" data-id="${u.id}" data-name="${esc(u.username)}" onclick="adminTriggerUserAvatarUpload(this.dataset.id)">${ic('image', 12)} Icon</button>
                      ${!isMe ? `
                        <button class="btn sm dg" data-id="${u.id}" data-name="${esc(u.username)}" onclick="deleteAdminUser(this.dataset.id, this.dataset.name)">${ic('trash', 12)} Delete</button>
                      ` : '<span class="mu" style="font-size:12px; padding:0 4px;">Active</span>'}
                    </div>
                  </td>
                </tr>`;
              }).join('') : `
                <tr>
                  <td>
                    <div class="row" style="gap:10px; align-items:center;">
                      ${av(32)}
                      <b>${esc(S.acct.un || S.user?.username || 'admin')}</b>
                      <span class="chip" style="margin-left:6px; font-size:10px;">You</span>
                    </div>
                  </td>
                  <td><span class="mono" style="font-size:13px;">${esc(S.acct.em || S.user?.email || 'admin@neopanel.internal')}</span></td>
                  <td><span class="badge online">${esc(S.user?.role || 'admin')}</span></td>
                  <td><span class="badge online"><i class="dot"></i>Standard</span></td>
                  <td class="mu">Active</td>
                  <td style="text-align:right"><span class="mu" style="font-size:12px;">Active</span></td>
                </tr>
              `}
            </tbody>
          </table>
        </div>
      `;
    } else if (t === 4) {
      content = `
        <div class="brand-layout-grid" style="margin-bottom: 18px;">
          <div class="card">
            <div class="row sb" style="margin-bottom: 6px; flex-wrap: wrap; gap: 10px;">
              <div>
                <h2 style="margin: 0; color: var(--tx);">Panel Look, Theme Presets &amp; Branding</h2>
                <span class="mu" style="font-size: 13px; display: block; margin-top: 3px;">
                  Persist global visual styles, colors, glass surface properties, artwork &amp; custom CSS for all users.
                </span>
              </div>
              <div class="row" style="gap: 8px;">
                <button type="button" class="btn sm" onclick="discardBrandDraft()">Discard</button>
                <button type="button" class="btn sm dg" onclick="resetBrandingDefaults()">Reset Defaults</button>
                <button type="button" id="brand-save-btn" class="btn sm pr" onclick="saveBrandingAndDaemon()">${ic('check', 13)} Save appearance</button>
              </div>
            </div>

            <div style="margin-top: 16px;">
              <div style="font-size: 12px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: var(--tx-muted); margin-bottom: 8px;">Theme Presets</div>
              <div class="brand-preset-grid">
                ${BRAND_PRESETS.map(p => `
                  <button type="button" class="brand-preset-card ${draft.preset === p.id ? 'on' : ''}" onclick="selectBrandPreset('${p.id}')">
                    <div class="brand-swatch-row">
                      ${p.swatches.map(sw => `<span class="brand-swatch-dot" style="background:${sw};"></span>`).join('')}
                    </div>
                    <div>
                      <b style="display:block; font-size:13px; color:var(--tx);">${esc(p.name)}</b>
                      <span class="mu" style="font-size:11px;">${esc(p.desc)}</span>
                    </div>
                  </button>
                `).join('')}
              </div>
            </div>

            <div class="grid g2" style="margin-top: 18px;">
              <label class="f" style="margin-bottom: 0;">
                <span>Panel Brand Title</span>
                <input type="text" value="${esc(draft.brand_name || '')}" maxlength="48" oninput="onBrandInputChange('brand_name', this.value)" placeholder="NeoPanel">
              </label>
              <label class="f" style="margin-bottom: 0;">
                <span>Brand Subtitle / Page Title Suffix</span>
                <input type="text" value="${esc(draft.brand_subtitle || '')}" maxlength="120" oninput="onBrandInputChange('brand_subtitle', this.value)" placeholder="Minecraft server management">
              </label>
              <label class="f" style="margin-bottom: 0;">
                <span>Footer and sign-in text</span>
                <input type="text" value="${esc(draft.footer_text || '')}" maxlength="120" oninput="onBrandInputChange('footer_text', this.value)" placeholder="[NeoPanel](https://github.com/Pekly/NeoPanel/)© 2026">
              </label>
              <label class="f" style="margin-bottom: 0;">
                <span>Default Color Mode (New Browsers)</span>
                <select onchange="onBrandInputChange('default_theme', this.value)">
                  <option value="dark" ${(draft.default_theme || 'dark') === 'dark' ? 'selected' : ''}>Dark Glass (Default)</option>
                  <option value="oled" ${draft.default_theme === 'oled' ? 'selected' : ''}>OLED Pure Black</option>
                  <option value="light" ${draft.default_theme === 'light' ? 'selected' : ''}>Light Daylight</option>
                  <option value="auto" ${draft.default_theme === 'auto' ? 'selected' : ''}>Match OS Preference</option>
                </select>
              </label>
              <label class="f" style="margin-bottom: 0;">
                <span>Panel Timezone</span>
                <select onchange="onBrandInputChange('timezone', this.value)">
                  ${['UTC', 'America/New_York', 'America/Chicago', 'America/Denver', 'America/Los_Angeles', 'America/Sao_Paulo', 'Europe/London', 'Europe/Paris', 'Europe/Berlin', 'Asia/Tokyo', 'Asia/Shanghai', 'Asia/Singapore', 'Asia/Ulaanbaatar', 'Australia/Sydney'].map(tz => `<option value="${tz}" ${(draft.timezone || 'UTC') === tz ? 'selected' : ''}>${tz}</option>`).join('')}
                </select>
              </label>
            </div>

            <div style="margin-top: 18px;">
              <div style="font-size: 12px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: var(--tx-muted); margin-bottom: 10px;">Custom Colors &amp; Surfaces</div>
              <div class="grid g3">
                <label class="f" style="margin-bottom: 0;">
                  <span>Accent Color</span>
                  <div class="brand-color-row">
                    <input type="color" class="brand-color-picker" value="${esc(draft.accent_color || '#387aff')}" oninput="onBrandInputChange('accent_color', this.value, this)">
                    <input type="text" value="${esc(draft.accent_color || '#387aff')}" placeholder="#387aff" oninput="onBrandInputChange('accent_color', this.value, this)">
                  </div>
                </label>
                <label class="f" style="margin-bottom: 0;">
                  <span>Accent Hover</span>
                  <div class="brand-color-row">
                    <input type="color" class="brand-color-picker" value="${esc(draft.accent_hover || '#4f8aff')}" oninput="onBrandInputChange('accent_hover', this.value, this)">
                    <input type="text" value="${esc(draft.accent_hover || '#4f8aff')}" placeholder="#4f8aff" oninput="onBrandInputChange('accent_hover', this.value, this)">
                  </div>
                </label>
                <label class="f" style="margin-bottom: 0;">
                  <span>Text Color (optional)</span>
                  <div class="brand-color-row">
                    <input type="color" class="brand-color-picker" value="${esc(draft.text_color || '#e6edf3')}" oninput="onBrandInputChange('text_color', this.value, this)">
                    <input type="text" value="${esc(draft.text_color || '')}" placeholder="Theme default" oninput="onBrandInputChange('text_color', this.value, this)">
                  </div>
                </label>
                <label class="f" style="margin-bottom: 0;">
                  <span>Deep Backdrop</span>
                  <div class="brand-color-row">
                    <input type="color" class="brand-color-picker" value="${esc(draft.bg_deep || '#0a0d14')}" oninput="onBrandInputChange('bg_deep', this.value, this)">
                    <input type="text" value="${esc(draft.bg_deep || '')}" placeholder="Theme default" oninput="onBrandInputChange('bg_deep', this.value, this)">
                  </div>
                </label>
                <label class="f" style="margin-bottom: 0;">
                  <span>Card Surface Tint</span>
                  <div class="brand-color-row">
                    <input type="color" class="brand-color-picker" value="${esc(draft.card_bg || '#121826')}" oninput="onBrandInputChange('card_bg', this.value, this)">
                    <input type="text" value="${esc(draft.card_bg || '')}" placeholder="Theme default" oninput="onBrandInputChange('card_bg', this.value, this)">
                  </div>
                </label>
                <label class="f" style="margin-bottom: 0;">
                  <span>Sidebar Background</span>
                  <div class="brand-color-row">
                    <input type="color" class="brand-color-picker" value="${esc(draft.nav_bg || '#0d121d')}" oninput="onBrandInputChange('nav_bg', this.value, this)">
                    <input type="text" value="${esc(draft.nav_bg || '')}" placeholder="Theme default" oninput="onBrandInputChange('nav_bg', this.value, this)">
                  </div>
                </label>
              </div>
            </div>

            <div class="grid g3" style="margin-top: 16px;">
              <label class="f" style="margin-bottom: 0;">
                <span>Card Opacity (<b id="lbl-card-opacity">${draft.card_opacity ?? 72}</b>%)</span>
                <div class="brand-range-row">
                  <input type="range" min="35" max="100" value="${draft.card_opacity ?? 72}" oninput="$('#lbl-card-opacity').textContent=this.value; onBrandInputChange('card_opacity', Number(this.value));">
                </div>
              </label>
              <label class="f" style="margin-bottom: 0;">
                <span>Border Radius (<b id="lbl-border-radius">${draft.border_radius_px ?? 14}</b>px)</span>
                <div class="brand-range-row">
                  <input type="range" min="0" max="28" value="${draft.border_radius_px ?? 14}" oninput="$('#lbl-border-radius').textContent=this.value; onBrandInputChange('border_radius_px', Number(this.value));">
                </div>
              </label>
              <label class="f" style="margin-bottom: 0;">
                <span>Glass Blur (<b id="lbl-glass-blur">${draft.glass_blur_px ?? 16}</b>px)</span>
                <div class="brand-range-row">
                  <input type="range" min="0" max="32" value="${draft.glass_blur_px ?? 16}" oninput="$('#lbl-glass-blur').textContent=this.value; onBrandInputChange('glass_blur_px', Number(this.value));">
                </div>
              </label>
            </div>

            <div style="margin-top: 18px;">
              <div class="row sb" style="margin-bottom: 10px; flex-wrap:wrap; gap:8px;">
                <div style="font-size: 12px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: var(--tx-muted);">Media &amp; Artwork Assets (Upload or URL)</div>
                <span class="mu" style="font-size: 11px;">PNG, JPG, WebP, SVG, ICO, GIF (up to 12 MB)</span>
              </div>
              <input type="file" id="brand-file-input" accept="image/*,.ico,.svg" style="display:none" onchange="handleBrandingFileSelect(this)">
              <div class="brand-media-grid">
                ${[
                  ['logo_url', 'Sidebar Logo (blank = 3D Cube)', 'https://example.com/logo.png', '/favicon.svg'],
                  ['favicon_url', 'Browser Favicon (syncs tab icon)', 'https://example.com/favicon.png', '/favicon.svg'],
                  ['login_bg_url', 'Login Screen Wallpaper', 'https://example.com/login-bg.jpg', MINECRAFT_COVER_URL],
                  ['server_hero_url', 'Server Workspace Hero Banner', SERVER_HERO_URL, SERVER_HERO_URL],
                  ['card_cover_url', 'Dashboard Card Cover', MINECRAFT_COVER_URL, MINECRAFT_COVER_URL],
                  ['default_avatar_url', 'Default User Avatar', DEFAULT_USER_AVATAR_URL, DEFAULT_USER_AVATAR_URL],
                  ['default_server_icon_url', 'Default Server Icon (Fallback 64×64)', '/static/img/default-server-icon.png', '/static/img/default-server-icon.png']
                ].map(([field, label, ph, fallbackThumb]) => {
                  const curVal = draft[field] || '';
                  const thumbSrc = curVal || fallbackThumb;
                  return `<div class="f brand-media-field" style="margin-bottom: 0;">
                    <label for="brand-media-${field}">${label}</label>
                    <div class="brand-upload-row">
                      <img class="brand-thumb-mini" data-brand-thumb="${field}" src="${esc(thumbSrc)}" alt="" onerror="this.src='/favicon.svg'">
                      <input id="brand-media-${field}" type="text" data-brand-media="${field}" value="${esc(curVal)}" oninput="onBrandInputChange('${field}', this.value)" placeholder="${esc(ph)}">
                      <button type="button" class="btn sm" data-upload-btn="${field}" onclick="triggerBrandingUpload('${field}', this)" title="Upload local image file">${ic('upload', 12)} Upload</button>
                      <button type="button" class="btn sm" onclick="clearBrandingMedia('${field}')" title="Reset to default" aria-label="Reset ${esc(label)}">&times;</button>
                    </div>
                  </div>`;
                }).join('')}
              </div>
            </div>

            <div style="margin-top: 16px;">
              <label class="f" style="margin-bottom: 0;">
                <span>Custom CSS Injection</span>
                <textarea rows="4" spellcheck="false" style="font-family:var(--font-mono); font-size:12.5px;" placeholder="/* e.g. .sift-server-card { border-width: 2px; } */" oninput="onBrandInputChange('custom_css', this.value)">${esc(draft.custom_css || '')}</textarea>
              </label>
            </div>
          </div>

          <div class="card brand-preview-card">
            <div class="row sb" style="margin-bottom: 12px;">
              <h2 style="margin: 0; font-size: 16px;">Live Panel Preview</h2>
              <span class="chip">${esc(draft.preset || 'custom')}</span>
            </div>
            <div id="brand-live-preview">
              ${renderBrandPreviewMarkup(draft)}
            </div>
            <p class="mu" style="font-size: 12px; line-height: 1.5; margin: 12px 0;">
              Changes apply live as you edit so you can inspect colors, glass blur, typography, and banners before saving to SQLite.
            </p>
            <div class="row" style="gap: 8px; flex-wrap: wrap;">
              <button type="button" class="btn pr" style="flex: 1;" onclick="saveBrandingAndDaemon()">${ic('check', 14)} Save appearance</button>
              <button type="button" class="btn" onclick="discardBrandDraft()">Discard</button>
            </div>
          </div>
        </div>

        ${renderAdminPanelUrlCard()}

        <div class="card" style="margin-bottom:18px;">
          <h2 style="margin:0 0 6px; color:var(--tx);">Network and server defaults</h2>
          <span class="mu" style="font-size:13px; display:block; margin-bottom:18px;">
            Persisted host FQDN, SFTP port, default instance memory, and auto-restart policy for the local master daemon.
          </span>
          <form onsubmit="saveAdminDaemonSettings(event)">
            <div class="grid g2">
              <label class="f">
                <span>Public Host IP / FQDN (Connect Address &amp; SFTP Host)</span>
                <input type="text" id="adm-fqdn" value="${esc(daemonDraft.fqdn || '')}" placeholder="${esc(window.location.hostname || 'play.example.com')}" oninput="onDaemonInputChange('fqdn', this.value)">
              </label>
              <label class="f">
                <span>SFTP port</span>
                <input type="number" id="adm-sftp" value="${daemonDraft.sftp_port || 2022}" min="1" max="65535" required oninput="onDaemonInputChange('sftp_port', Number(this.value))">
              </label>
            </div>
            <div class="grid g2" style="margin-top:4px;">
              <label class="f">
                <span>Max Upload Size (MB)</span>
                <input type="number" id="adm-upload" value="${daemonDraft.max_upload_mb || 500}" min="10" max="25000" required oninput="onDaemonInputChange('max_upload_mb', Number(this.value))">
              </label>
              <label class="f">
                <span>Default server memory (MB)</span>
                <input type="number" id="adm-ram" value="${daemonDraft.default_ram_mb || 2048}" min="256" max="131072" required oninput="onDaemonInputChange('default_ram_mb', Number(this.value))">
              </label>
            </div>
            <div style="margin-top:8px; padding:12px 14px; background:rgba(255,255,255,0.02); border:1px solid var(--border); border-radius:8px;">
              <label class="row sb" style="cursor:pointer;">
                <div>
                  <b style="color:var(--tx); display:block;">Auto-Restart Policy</b>
                  <span class="mu" style="font-size:12.5px;">Automatically restart crashed server processes (times out after 2 consecutive startup failures)</span>
                </div>
                <span class="sw">
                  <input type="checkbox" id="adm-autorestart" ${daemonDraft.auto_restart_policy ? 'checked' : ''} onchange="onDaemonInputChange('auto_restart_policy', this.checked)">
                  <i></i>
                </span>
              </label>
            </div>
            <div class="row" style="justify-content:flex-end; margin-top:18px;">
              <button type="submit" class="btn pr">${ic('download', 14)} Save defaults</button>
            </div>
          </form>
        </div>

        <div class="card" style="margin-bottom:18px;">
          <div class="row sb" style="margin-bottom: 6px; flex-wrap: wrap; gap: 8px;">
            <h2 style="margin:0; color:var(--tx);">Cloudflare Turnstile Bot Protection</h2>
            <span class="badge ${S.adminTurnstile?.enabled ? 'online' : 'offline'}">
              <i class="dot"></i>${S.adminTurnstile?.enabled ? 'Enabled' : 'Disabled'}
            </span>
          </div>
          <span class="mu" style="font-size:13px; display:block; margin-bottom:18px;">
            Protect Login and Initial Setup endpoints against automated brute-force attacks with Cloudflare Turnstile.
          </span>
          <form onsubmit="saveAdminTurnstileSettings(event)">
            <div style="margin-bottom:14px; padding:12px 14px; background:rgba(255,255,255,0.02); border:1px solid var(--border); border-radius:8px;">
              <label class="row sb" style="cursor:pointer;">
                <div>
                  <b style="color:var(--tx); display:block;">Enable Cloudflare Turnstile</b>
                  <span class="mu" style="font-size:12.5px;">Require Turnstile challenge verification on login &amp; setup screens</span>
                </div>
                <span class="sw">
                  <input type="checkbox" id="adm-ts-enabled" ${S.adminTurnstile?.enabled ? 'checked' : ''}>
                  <i></i>
                </span>
              </label>
            </div>
            <div class="grid g2">
              <label class="f">
                <span>Turnstile Site Key</span>
                <input type="text" id="adm-ts-site" placeholder="0x4AAAAAA..." value="${esc(S.adminTurnstile?.site_key || '')}">
              </label>
              <label class="f">
                <span>Turnstile Secret Key ${S.adminTurnstile?.has_secret_key ? '<span class="mu">(saved — leave blank to keep)</span>' : ''}</span>
                <input type="password" id="adm-ts-secret" placeholder="${S.adminTurnstile?.has_secret_key ? '••••••••••••••••••••••••••••' : '0x4AAAAAA...'}" autocomplete="new-password">
              </label>
            </div>
            <div class="row" style="justify-content:flex-end; margin-top:18px;">
              <button type="submit" class="btn pr">${ic('shield', 14)} Save Turnstile Settings</button>
            </div>
          </form>
        </div>

        <div class="grid g2" style="margin-bottom:18px;">
          <div class="card">
            <div class="row sb" style="margin-bottom: 6px; flex-wrap: wrap; gap: 8px;">
              <h2 style="margin:0; color:var(--tx);">Global Maintenance Mode</h2>
              <span class="badge ${draft.maintenance_enabled ? 'starting' : 'offline'}">
                <i class="dot"></i>${draft.maintenance_enabled ? 'Active' : 'Inactive'}
              </span>
            </div>
            <span class="mu" style="font-size:12.5px; display:block; margin-bottom:14px;">
              Locks standard user access and guests behind a 503 Maintenance screen while Administrators retain full access.
            </span>
            <div style="margin-bottom:12px; padding:10px 12px; background:rgba(255,255,255,0.02); border:1px solid var(--border); border-radius:8px;">
              <label class="row sb" style="cursor:pointer;">
                <span style="font-weight:600; color:var(--tx);">Enable Maintenance Mode</span>
                <span class="sw">
                  <input type="checkbox" ${draft.maintenance_enabled ? 'checked' : ''} onchange="onBrandInputChange('maintenance_enabled', this.checked)">
                  <i></i>
                </span>
              </label>
            </div>
            <label class="f">
              <span>Maintenance Title</span>
              <input type="text" value="${esc(draft.maintenance_title || '')}" maxlength="96" oninput="onBrandInputChange('maintenance_title', this.value)" placeholder="Scheduled Maintenance">
            </label>
            <label class="f">
              <span>Maintenance Message</span>
              <textarea rows="3" maxlength="512" oninput="onBrandInputChange('maintenance_message', this.value)" placeholder="We are upgrading NeoPanel nodes. Running game servers stay online during panel maintenance.">${esc(draft.maintenance_message || '')}</textarea>
            </label>
            <label class="f">
              <span>Estimated Return / ETA (optional)</span>
              <input type="text" value="${esc(draft.maintenance_eta || '')}" maxlength="96" oninput="onBrandInputChange('maintenance_eta', this.value)" placeholder="e.g. ~15 minutes / 02:30 UTC">
            </label>
            <div class="row" style="justify-content:flex-end; gap:8px; margin-top:14px; flex-wrap:wrap;">
              <button type="button" class="btn sm" onclick="toggleMaintPreview(true)">${ic('eye', 13)} Preview Screen</button>
              <button type="button" class="btn sm pr" onclick="saveBrandingAndDaemon()">${ic('check', 13)} Save Maintenance</button>
            </div>
          </div>

          <div class="card">
            <div class="row sb" style="margin-bottom: 6px; flex-wrap: wrap; gap: 8px;">
              <h2 style="margin:0; color:var(--tx);">Client Area Storefront Portal</h2>
              <span class="badge ${draft.client_area_enabled ? (draft.client_area_available !== false ? 'online' : 'starting') : 'offline'}">
                <i class="dot"></i>${draft.client_area_enabled ? (draft.client_area_available !== false ? 'Connected' : 'Unavailable') : 'Not Set Up'}
              </span>
            </div>
            <span class="mu" style="font-size:12.5px; display:block; margin-bottom:14px;">
              Link your external billing / server-buying storefront (e.g. <code class="mono">https://client.hosting.net</code>) inside the sidebar.
            </span>
            <div style="margin-bottom:10px; padding:10px 12px; background:rgba(255,255,255,0.02); border:1px solid var(--border); border-radius:8px;">
              <label class="row sb" style="cursor:pointer; margin-bottom:8px;">
                <span style="font-weight:600; color:var(--tx);">Enable Storefront Integration</span>
                <span class="sw">
                  <input type="checkbox" ${draft.client_area_enabled ? 'checked' : ''} onchange="onBrandInputChange('client_area_enabled', this.checked)">
                  <i></i>
                </span>
              </label>
              <label class="row sb" style="cursor:pointer; margin-bottom:8px;">
                <span style="font-size:13px; color:var(--tx-muted);">Show "Client Area" Navigation Button in Sidebar</span>
                <span class="sw">
                  <input type="checkbox" ${draft.show_client_area_button !== false ? 'checked' : ''} onchange="onBrandInputChange('show_client_area_button', this.checked)">
                  <i></i>
                </span>
              </label>
              <label class="row sb" style="cursor:pointer;">
                <span style="font-size:13px; color:var(--tx-muted);">Mark Storefront Available (off = 503 Unavailable error)</span>
                <span class="sw">
                  <input type="checkbox" ${draft.client_area_available !== false ? 'checked' : ''} onchange="onBrandInputChange('client_area_available', this.checked)">
                  <i></i>
                </span>
              </label>
            </div>
            <div class="grid g2">
              <label class="f">
                <span>Sidebar Link Label</span>
                <input type="text" value="${esc(draft.client_area_label || 'Client Area')}" maxlength="48" oninput="onBrandInputChange('client_area_label', this.value)" placeholder="Client Area">
              </label>
              <label class="f">
                <span>Open Behavior</span>
                <select onchange="onBrandInputChange('client_area_open_mode', this.value)">
                  <option value="tab" ${(draft.client_area_open_mode || 'tab') === 'tab' ? 'selected' : ''}>Portal Card &amp; New Tab</option>
                  <option value="embed" ${draft.client_area_open_mode === 'embed' ? 'selected' : ''}>Embed in Panel (iframe)</option>
                </select>
              </label>
            </div>
            <label class="f">
              <span>External Storefront URL</span>
              <input type="text" value="${esc(draft.client_area_url || '')}" maxlength="1024" oninput="onBrandInputChange('client_area_url', this.value)" placeholder="https://client.hosting.net">
            </label>
            <label class="f">
              <span>Storefront Subtitle</span>
              <input type="text" value="${esc(draft.client_area_subtitle || '')}" maxlength="200" oninput="onBrandInputChange('client_area_subtitle', this.value)" placeholder="Purchase servers, manage subscriptions, and view invoices.">
            </label>
            <div class="row" style="justify-content:flex-end; gap:8px; margin-top:14px; flex-wrap:wrap;">
              <a href="/client-area" onclick="return handleNavClick(event, 'client')" class="btn sm" style="text-decoration:none;">${ic('external', 13)} Open /client-area</a>
              <button type="button" class="btn sm pr" onclick="saveBrandingAndDaemon()">${ic('check', 13)} Save Storefront</button>
            </div>

            <!-- Automated Provisioning Integration & API Code Samples -->
            <div style="margin-top: 18px; border-top: 1px solid var(--border); padding-top: 16px;">
              <div class="row sb" style="margin-bottom: 8px;">
                <div>
                  <b style="color: var(--tx); font-size: 13.5px; display: block;">Server management API</b>
                  <span class="mu" style="font-size: 12px;">Call <code>POST /api/client-api/v1/provision</code> from WHMCS, Blesta, or billing systems when an order is placed</span>
                </div>
              </div>
              <div class="tabs" style="margin-bottom: 8px;">
                <button type="button" class="btn sm ${S.clientCodeLang === 'php' || !S.clientCodeLang ? 'pr' : ''}" onclick="setClientCodeLang('php')">PHP (WHMCS / Blesta)</button>
                <button type="button" class="btn sm ${S.clientCodeLang === 'node' ? 'pr' : ''}" onclick="setClientCodeLang('node')">Node.js</button>
                <button type="button" class="btn sm ${S.clientCodeLang === 'python' ? 'pr' : ''}" onclick="setClientCodeLang('python')">Python</button>
                <button type="button" class="btn sm ${S.clientCodeLang === 'curl' ? 'pr' : ''}" onclick="setClientCodeLang('curl')">cURL</button>
              </div>
              <div style="position: relative;">
                <pre class="mono" style="background: var(--con-bg); border: 1px solid var(--border); border-radius: 8px; padding: 12px 14px; font-size: 12px; color: #f0f3f6; overflow-x: auto; max-height: 220px; white-space: pre;">${esc(getClientProvisioningCode(S.clientCodeLang || 'php'))}</pre>
                <button type="button" class="btn sm" style="position: absolute; right: 8px; top: 8px;" onclick="copyClientProvisioningCode()">${ic('copy', 12)} Copy</button>
              </div>
            </div>
          </div>
        </div>
      `;
    } else if (t === 5) {
      content = renderActivityLogSection({
        scope: 'all',
        title: 'Audit logs',
        subtitle: 'Cluster-wide security, node connections, databases, user management, power, console, file & software audit events',
        showActor: true
      });
    } else if (t === 6) {
      content = renderAdminModsTabHtml();
    } else if (t === 7) {
      content = renderAdminEmailTabHtml();
    } else if (t === 8) {
      content = renderPanelRoles();
    }
    if (t === 4) content = organizeAppearance(content);

    return `<header class="np-admin-heading">
      ${ic(section.icon, 26)}
      <div><div class="np-eyebrow">Admin</div><h1>${esc(section.label)}</h1><p>${esc(section.description)}</p></div>
    </header>
    <section class="np-admin-summary" aria-label="Panel status summary">
      <div><span class="mu">Servers running</span><strong>${runningServers} / ${S.servers.length}</strong></div>
      <div><span class="mu">Nodes online</span><strong>${S.system || S.nodes?.length ? `${onlineNodesCount} / ${nodesList.length}` : '—'}</strong></div>
      <div><span class="mu">Panel version</span><strong>${sys.version ? `v${esc(sys.version)}` : '—'}</strong></div>
      <div><span class="mu">Updates</span><strong>${esc(updateStatus)}</strong></div>
      <div class="np-admin-summary-actions"><button class="btn sm" onclick="refreshAdminStatus(this)" ${adminStatusRefreshing ? 'disabled' : ''}>${ic('refresh', 13)} Refresh status</button>${S.adminRefreshedAt ? `<span class="mu">Updated ${esc(new Date(S.adminRefreshedAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }))}</span>` : ''}</div>
    </section>
    ${content}`;
  },

  acc() {
    const a = S.acct;
    const t = a.tab;
    let tabContent = '';

    if (t === 0) {
      tabContent = `<div class="card prof">
        <div class="banner" style="background-image: url('${esc(getBrand('card_cover_url'))}'), url('/img/banner.jpg'); background-size: cover; background-position: center;">
          <span class="avb">${av(64)}</span>
          <input type="file" id="my-avatar-upload" accept="image/*" style="display:none;" onchange="uploadMyAvatar(this)">
          <div class="row" style="gap:8px;">
            <button type="button" class="btn sm" onclick="$('#my-avatar-upload').click()">${ic('image', 13)} Upload Avatar</button>
            <button type="button" class="btn sm" onclick="$('#aem').focus()">Gravatar</button>
          </div>
        </div>
        <div class="pbody">
          <div class="grid g2">
            <label class="f"><span>First name</span><input type="text" id="afn" value="${esc(a.fn || '')}" placeholder="First name"></label>
            <label class="f"><span>Last name</span><input type="text" id="aln" value="${esc(a.ln || '')}" placeholder="Last name"></label>
          </div>
          <label class="f"><span>Username / Account Name <b style="color:var(--err); font-size:12px;">*required</b></span><input type="text" id="aun" value="${esc(a.un || '')}" placeholder="Username" autocomplete="username"></label>
          <label class="f"><span>Email Address <b style="color:var(--err); font-size:12px;">*required</b></span><input type="email" id="aem" value="${esc(a.em || '')}" placeholder="name@example.com" autocomplete="email"></label>
          <div class="row" style="justify-content:flex-end">
            <button class="btn pr" onclick="updateAccountProfile()">${ic('check', 14)} Update Profile</button>
          </div>
        </div>
      </div>

      <div class="card" style="margin-top:14px">
        <h2>Appearance &amp; Preferences</h2>
        <div class="ar">
          <div><b>Color Mode</b><div class="mu">Choose the aesthetic that suits your workstation</div></div>
          <div class="seg">
            ${['Light', 'Dark', 'Oled', 'Auto'].map((m, i) => `
              <button class="${a.mode === m ? 'on' : ''}" onclick="setAppearanceMode('${m}')">${m}</button>
            `).join('')}
          </div>
        </div>
        <div class="ar">
          <div><b>Compact Mode</b><div class="mu">Reduce spacing for higher data density</div></div>
          <label class="sw"><input type="checkbox" ${a.compact ? 'checked' : ''} onchange="toggleCompact(this.checked)"><i></i></label>
        </div>
        <div class="ar">
          <div><b>Panel Sounds</b><div class="mu">Play subtle audio click on interactive actions</div></div>
          <label class="sw"><input type="checkbox" ${a.sound ? 'checked' : ''} onchange="toggleSound(this.checked)"><i></i></label>
        </div>
        <div class="ar">
          <div><b>Privacy Mode</b><div class="mu">Blur server IPs and personal emails across the panel</div></div>
          <label class="sw"><input type="checkbox" ${a.priv ? 'checked' : ''} onchange="togglePrivacy(this.checked)"><i></i></label>
        </div>
        <div class="ar">
          <div><b>Timezone</b><div class="mu">Display activity logs, console timestamps, and file dates in your timezone</div></div>
          <select id="acc-timezone" onchange="setUserTimezone(this.value)" style="max-width: 220px;">
            ${['UTC', 'America/New_York', 'America/Chicago', 'America/Denver', 'America/Los_Angeles', 'America/Sao_Paulo', 'Europe/London', 'Europe/Paris', 'Europe/Berlin', 'Asia/Tokyo', 'Asia/Shanghai', 'Asia/Singapore', 'Asia/Ulaanbaatar', 'Australia/Sydney'].map(tz => `<option value="${tz}" ${getUserTimezone() === tz ? 'selected' : ''}>${tz}</option>`).join('')}
          </select>
        </div>
      </div>`;
    } else if (t === 1) {
      const is2FA = Boolean(S.user?.two_factor_enabled || a.two_factor_enabled);
      tabContent = `<div class="card">
        <h2>Update password</h2>
        <label class="f"><span>Current password</span><input type="password" id="cur-pw" autocomplete="current-password"></label>
        <label class="f"><span>New password</span><input type="password" id="new-pw"></label>
        <button class="btn pr" onclick="updatePassword()">Update password</button>
      </div>
      <div class="card" style="margin-top:14px">
        <div class="row sb" style="align-items:center; flex-wrap:wrap; gap:12px;">
          <div>
            <div class="row" style="gap:8px; align-items:center;">
              <h2 style="margin:0;">Two-factor authentication</h2>
              ${is2FA ? '<span class="badge online"><i class="dot"></i>Active</span>' : '<span class="badge offline"><i class="dot"></i>Disabled</span>'}
            </div>
            <span class="mu" style="font-size:13px; display:block; margin-top:4px;">
              ${is2FA
                ? 'Password sign-in requires an authenticator or recovery code. A verified passkey signs you in directly.'
                : 'Protect your account from unauthorized sign-ins using Google Authenticator, Authy, 1Password, or standard TOTP app.'}
            </span>
          </div>
          ${is2FA ? `
            <button type="button" class="btn sm dg" onclick="openDisable2FAModal()">${ic('lock', 13)} Disable 2FA</button>
          ` : `
            <button type="button" class="btn sm pr" onclick="openEnable2FAWizard()">${ic('shield', 13)} Set up 2FA</button>
          `}
        </div>
      </div>`;
      tabContent += renderPasskeySettings();
    } else if (t === 2) {
      tabContent = `<div class="card">
        <h2>Add OpenSSH Public Key</h2>
        <p class="mu" style="font-size:13px; margin-bottom:12px;">Allows passwordless public-key authentication for pure-Go SFTP.</p>
        <label class="f"><span>Key Label / Name</span><input type="text" id="kn" placeholder="Work Laptop"></label>
        <label class="f"><span>Public Key</span><textarea id="kk" style="min-height:80px" placeholder="ssh-ed25519 AAAA... or ssh-rsa AAAA..."></textarea></label>
        <button class="btn pr" onclick="addSSHKey()">${ic('plus', 14)} Add SSH Key</button>
      </div>
      <div class="card tw" style="margin-top:14px">
        <table>
          <thead><tr><th>Name</th><th>Fingerprint</th><th>Added</th><th style="text-align:right">Action</th></tr></thead>
          <tbody>
            ${a.ssh.length ? a.ssh.map((k, i) => `
              <tr>
                <td><b>${esc(k.name || k[0])}</b></td>
                <td class="mono">${esc(k.fingerprint || k[1])}</td>
                <td class="mu">${esc(k.created_at ? formatFileDate(k.created_at) : (k[2] || 'Recently'))}</td>
                <td style="text-align:right">
                  <button class="btn sm dg" onclick="deleteSSHKey('${k.id || i}')">${ic('trash', 14)} Delete</button>
                </td>
              </tr>
            `).join('') : '<tr><td colspan="4" class="mu" style="text-align:center;padding:24px;">No SSH keys registered. Add your public key above for secure SFTP access.</td></tr>'}
          </tbody>
        </table>
      </div>`;
    } else if (t === 3) {
      tabContent = `${a.newkey ? `
        <div class="card hero" style="margin-bottom:14px;">
          <b>Copy your newly generated API key now</b>
          <div class="mono" style="margin:6px 0; font-size:14px; user-select:all; background:var(--con); padding:8px; border-radius:6px; color:#fff;">${esc(a.newkey)}</div>
          <span class="mu">For security, this secret token will not be displayed again.</span>
        </div>
      ` : ''}
      <div class="card">
        <h2>Create API key</h2>
        <div class="row">
          <input type="text" id="ad" placeholder="Key description (e.g. CI/CD Deployer)" style="max-width:340px">
          <button class="btn pr" onclick="createAPIKey()">${ic('plus', 14)} Generate Key</button>
        </div>
      </div>
      <div class="card tw" style="margin-top:14px">
        <table>
          <thead><tr><th>Description</th><th>Key Prefix</th><th>Created</th><th style="text-align:right">Action</th></tr></thead>
          <tbody>
            ${a.api.length ? a.api.map((k, i) => `
              <tr>
                <td><b>${esc(k.name || k.description || k[0])}</b></td>
                <td class="mono">${esc(k.key_prefix || k.prefix || k[1])}</td>
                <td class="mu">${esc(k.created_at ? formatFileDate(k.created_at) : (k.last_used || k[2] || 'Never'))}</td>
                <td style="text-align:right">
                  <button class="btn sm dg" onclick="deleteAPIKey('${k.id || i}')">${ic('trash', 14)} Revoke</button>
                </td>
              </tr>
            `).join('') : '<tr><td colspan="4" class="mu" style="text-align:center;padding:24px;">No API keys generated yet. Create a key above to automate your server.</td></tr>'}
          </tbody>
        </table>
      </div>`;
    } else {
      tabContent = renderActivityLogSection({
        scope: S.user?.role === 'admin' ? 'all' : 'mine',
        title: 'Account & Security Activity',
        subtitle: 'Sign-ins, password changes, SSH & API keys, and recent panel operations',
        showActor: true
      });
    }

    const accTabs = [
      ['settings', 'General'],
      ['lock', 'Security'],
      ['key', 'SSH Keys'],
      ['code', 'API Keys'],
      ['activity', 'Activity']
    ];

    return `
    <div class="row sb" style="margin-bottom: 24px;">
      <div class="row" style="gap: 14px; align-items: center;">
        <div style="width: 44px; height: 44px; border-radius: 12px; background: #6366f1; display: grid; place-items: center; color: #fff; box-shadow: 0 4px 14px rgba(99, 102, 241, 0.4);">
          ${ic('user', 22)}
        </div>
        <div>
          <h1 style="font-size: 24px; font-weight: 800; color: var(--tx); margin: 0; letter-spacing: -0.01em;">Account</h1>
        </div>
      </div>
      <button class="btn dg sm" onclick="logout()">Log Out</button>
    </div>
    <div class="tabs">
      ${accTabs.map(([icon, label], i) => `
        <button class="${t === i ? 'on' : ''}" onclick="selectAccountTab(${i});">
          ${ic(icon, 15)}
          <span>${label}</span>
        </button>
      `).join('')}
    </div>
    ${tabContent}`;
  }
};

// --- Version Rows Helper ---
function versionRowsHtml() {
  const q = (S.vq || '').toLowerCase().trim();
  const sel = S.sel || S.core || 'Paper';
  const fam = S.vf || 'All Families';
  const vm = getViewMode('ver', 'list');

  if (isSteelEngine(sel)) {
    const versions = getAllMcVersions(sel).filter(v => !q || v.toLowerCase().includes(q));
    if (!versions.length) return vm === 'list' ? '<tr><td colspan="5">No matching Steel releases.</td></tr>' : '<div>No matching Steel releases.</div>';
    return versions.map(v => {
      const installing = S.installingSoftware === `steel:${v}`;
      const active = isSteelEngine(S.core) && S.ver === v && v !== 'latest';
      const action = active ? '<span class="badge online">Active</span>' : `<button class="btn sm pr" data-soft="steel" data-ver="${esc(v)}" onclick="installSoftware(this.dataset.soft, this.dataset.ver, this)" ${installing ? 'disabled' : ''}>${installing ? 'Installing...' : 'Install Steel ' + esc(v)}</button>`;
      const notes = v === 'latest' ? 'Latest official Steel release' : 'Official Steel release';
      if (vm !== 'list') return `<div class="${vm === 'icons' ? 'vm-icon-card' : 'vm-tile-card'}"><div class="vm-item-head">${engineIcon('steel', 16)} <b>${esc(v)}</b></div><div class="vm-item-sub">${notes} • Native executable</div><div class="vm-item-foot">${action}</div></div>`;
      return `<tr><td class="mono">${esc(v)}</td><td>${notes}</td><td>Native executable</td><td>${engineIcon('steel', 12)} Steel</td><td style="text-align:right">${action}</td></tr>`;
    }).join('');
  }

  if (sel === 'Velocity' || sel === 'Waterfall' || sel === 'BungeeCord') {
    const dynamicProxy = getAllMcVersions(sel);
    const descMap = {
      '3.4.0-SNAPSHOT': 'Popular latest Velocity build',
      '3.3.0-SNAPSHOT': 'Stable Velocity release',
      '1.20': 'Final Waterfall release (Build 578)',
      latest: 'Continuous Jenkins CI build'
    };
    const proxyReleases = dynamicProxy.map(v => ({
      v,
      n: descMap[v] || `${sel} proxy release`,
      j: getRecommendedJavaForMcVersion(v, sel)
    }));

    const filtered = proxyReleases.filter(item => !q || (item.v + ' ' + item.n).toLowerCase().includes(q));
    if (!filtered.length) {
      return vm === 'list'
        ? '<tr><td colspan="5" class="mu" style="text-align:center;padding:20px;">No matching proxy builds found.</td></tr>'
        : '<div class="mu" style="grid-column:1/-1; text-align:center;padding:20px;">No matching proxy builds found.</div>';
    }
    return filtered.map(item => {
      const isRunning = S.ver === item.v && S.core === sel;
      const isInstalling = S.installingSoftware === `${sel}:${item.v}`;
      const actionBtn = isRunning
        ? '<span class="badge online"><i class="dot"></i>Active</span>'
        : isInstalling
          ? `<button class="btn sm pr is-loading" data-soft="${esc(sel)}" data-ver="${esc(item.v)}" onclick="installSoftware(this.dataset.soft, this.dataset.ver, this)" disabled aria-busy="true"><span class="np-spinner-sm"></span> Installing...</button>`
          : `<button class="btn sm pr" data-soft="${esc(sel)}" data-ver="${esc(item.v)}" onclick="installSoftware(this.dataset.soft, this.dataset.ver, this)">Install ${esc(sel)} ${esc(item.v)}</button>`;
      if (vm === 'icons') {
        return `<div class="vm-icon-card">
          <div class="vm-item-head"><b class="mono">${esc(item.v)}</b><span class="chip" style="display:inline-flex;align-items:center;gap:5px;">${engineIcon(sel, 12)} ${esc(sel)}</span></div>
          <div class="vm-item-sub">${esc(item.n)} &bull; Java ${esc(item.j)}</div>
          <div class="vm-item-foot">${actionBtn}</div>
        </div>`;
      }
      if (vm === 'tiles') {
        return `<div class="vm-tile-card">
          <div class="vm-item-head">
            <div><b class="mono" style="font-size:15px;">${esc(item.v)}</b><div class="mu" style="font-size:12px;margin-top:2px;">${esc(item.n)}</div></div>
            <span class="chip" style="display:inline-flex;align-items:center;gap:5px;">${engineIcon(sel, 12)} ${esc(sel)}</span>
          </div>
          <div class="vm-item-foot"><span class="mono mu" style="font-size:12px;">Java ${esc(item.j)}</span>${actionBtn}</div>
        </div>`;
      }
      return `<tr>
        <td class="mono"><b>${esc(item.v)}</b></td>
        <td><span class="chip">${esc(item.n)}</span></td>
        <td class="mono">Java ${esc(item.j)}</td>
        <td><span class="chip" style="display:inline-flex;align-items:center;gap:5px;">${engineIcon(sel, 12)} ${esc(sel)}</span></td>
        <td style="text-align:right">${actionBtn}</td>
      </tr>`;
    }).join('');
  }

  const POPULAR_VERS = new Set(['26.3', '1.21.11', '1.21.8', '1.21.4', '1.20.4', '1.20.1', '1.19.4', '1.16.5', '1.12.2', '1.8.9']);
  const items = getMergedVersionItems(sel).filter(item => {
    const matchesQ = !q || (item.v + ' ' + item.n + ' ' + item.y).toLowerCase().includes(q);
    const matchesF =
      fam === 'All Families' ||
      fam === 'All' ||
      (fam === 'Popular' ? POPULAR_VERS.has(item.v) : item.f === fam.split(' ')[0]);
    return matchesQ && matchesF;
  });

  if (!items.length) {
    return vm === 'list'
      ? '<tr><td colspan="5" class="mu" style="text-align:center;padding:20px;">No matching versions found.</td></tr>'
      : '<div class="mu" style="grid-column:1/-1; text-align:center;padding:20px;">No matching versions found.</div>';
  }

  return items.map(item => {
    const isRunning = S.ver === item.v && S.core === sel;
    const isInstalling = S.installingSoftware === `${sel}:${item.v}`;
    const javaVer = getRecommendedJavaForMcVersion(item.v, sel);
    const isPopular = POPULAR_VERS.has(item.v);
    const tagHtml = item.v === '26.3'
      ? '<span class="badge online" style="font-size:10px;padding:1px 6px;">Latest</span>'
      : isPopular
        ? '<span class="chip" style="font-size:10px;padding:1px 6px;">Popular</span>'
        : '';
    const actionBtn = isRunning
      ? '<span class="badge online"><i class="dot"></i>Active</span>'
      : isInstalling
        ? `<button class="btn sm pr is-loading" data-soft="${esc(sel)}" data-ver="${esc(item.v)}" onclick="installSoftware(this.dataset.soft, this.dataset.ver, this)" disabled aria-busy="true"><span class="np-spinner-sm"></span> Installing...</button>`
        : `<button class="btn sm pr" data-soft="${esc(sel)}" data-ver="${esc(item.v)}" onclick="installSoftware(this.dataset.soft, this.dataset.ver, this)">Install ${esc(sel)} ${esc(item.v)}</button>`;

    if (vm === 'icons') {
      return `<div class="vm-icon-card">
        <div class="vm-item-head">
          <b class="mono" style="font-size:15px;">${esc(item.v)}</b>
          ${tagHtml || `<span class="chip" style="font-size:10.5px;">Java ${esc(javaVer)}</span>`}
        </div>
        <div class="vm-item-sub">${esc(item.n)} (${esc(item.y)})</div>
        <div class="vm-item-foot">
          <span class="mu mono" style="font-size:11.5px;">Java ${esc(javaVer)}</span>
          ${actionBtn}
        </div>
      </div>`;
    }
    if (vm === 'tiles') {
      return `<div class="vm-tile-card">
        <div class="vm-item-head">
          <div>
            <div class="row" style="gap:6px;"><b class="mono" style="font-size:15px;">${esc(item.v)}</b>${tagHtml}</div>
            <div class="mu" style="font-size:12px; margin-top:3px;">${esc(item.n)} (${esc(item.y)})</div>
          </div>
          <span class="chip" style="display:inline-flex;align-items:center;gap:5px;">${engineIcon(sel, 12)} ${esc(sel)}</span>
        </div>
        <div class="vm-item-foot">
          <span class="mono mu" style="font-size:12px;">Java ${esc(javaVer)}</span>
          ${actionBtn}
        </div>
      </div>`;
    }
    return `<tr>
      <td class="mono">
        <b>${esc(item.v)}</b>
        ${item.v === '26.3' ? '<span class="badge online" style="font-size:10px;padding:1px 6px;margin-left:6px;">Latest</span>' : isPopular ? '<span class="chip" style="font-size:10px;padding:1px 6px;margin-left:6px;">Popular</span>' : ''}
      </td>
      <td><span class="chip">${esc(item.n)} (${esc(item.y)})</span></td>
      <td class="mono">Java ${esc(javaVer)}</td>
      <td><span class="chip" style="display:inline-flex;align-items:center;gap:5px;">${engineIcon(sel, 12)} ${esc(sel)}</span></td>
      <td style="text-align:right">${actionBtn}</td>
    </tr>`;
  }).join('');
}

function renderVersionRows() {
  const b = $('#vres');
  if (b) b.innerHTML = versionRowsHtml();
}

// --- Navigation Render ---
function renderModSidebarItemsHtml(scope) {
  const items = Array.isArray(window.NP?._sidebarItems) ? window.NP._sidebarItems : [];
  const filtered = items.filter(it => (it.scope || 'both') === scope || (it.scope || 'both') === 'both');
  if (!filtered.length) return '';
  return `<div class="ng">Extensions</div>` + filtered.map(it => {
    const href = it.url || (it.pageId ? `/mod/${encodeURIComponent(it.pageId)}` : '#');
    const isActive = Boolean(it.pageId && S.view === 'mod' && S.modPageId === it.pageId);
    const clickAttr = it.pageId
      ? `onclick="return handleNavClick(event, 'mod', { modPageId: '${esc(it.pageId)}' })"`
      : `onclick="if(window.NP&&window.NP._execSidebarItem){return window.NP._execSidebarItem('${esc(it.id)}', event);}return false;"`;
    return `<a class="nav-item ${isActive ? 'on' : ''}" href="${esc(withSafeModeQuery(href))}" ${clickAttr} ${isActive ? 'aria-current="page"' : ''}>
      ${ic(it.icon || 'puzzle', 16)}
      <span>${esc(it.label || 'Extension')}</span>
    </a>`;
  }).join('');
}

function selectAdminTabNav(e, idx) {
  if (e && (e.ctrlKey || e.metaKey || e.shiftKey || e.button !== 0)) return true;
  if (e && e.preventDefault) e.preventDefault();
  selectAdminTab(idx);
  return false;
}

function navHtml() {
  const brandHref = withSafeModeQuery('/');
  const adminHref = routeUrl('admin');
  const accHref = routeUrl('acc');
  const b = S.branding || DEFAULT_BRANDING;
  const clientLabel = (b.client_area_label || 'Client Area').trim() || 'Client Area';
  const clientStatusBadge = !b.client_area_enabled
    ? `<span class="chip" style="margin-left:auto; font-size:10px; padding:1px 6px;">Setup</span>`
    : (b.client_area_available === false
      ? `<span class="chip" style="margin-left:auto; font-size:10px; padding:1px 6px; color:#f87171;">503</span>`
      : '');

  // 1. Independent Dedicated Admin Navigation
  if (S.view === 'admin' && S.user?.role === 'admin') {
    const t = S.adminTab || 0;
    return `
    <a class="brand" href="${esc(brandHref)}" onclick="return handleNavClick(event, 'dash')" style="cursor:pointer; text-decoration:none; color:inherit;">
      ${brandLogoHtml(false)}
    </a>
    
    <button class="nsearch" onclick="openPalette(1)" aria-label="Search (Ctrl+K)">
      ${ic('search', 15)}
      <span>Search</span>
      <kbd>Ctrl K</kbd>
    </button>

    <a class="nb" href="${esc(brandHref)}" onclick="return handleNavClick(event, 'dash')" style="margin-bottom: 8px; text-decoration:none;">
      &larr; All servers
    </a>

    <div class="ng">Admin</div>
    ${ADMIN_SECTIONS.map((section, index) => `<a class="nav-item ${t === index ? 'on' : ''}" data-admin-section="${index}" href="${esc(routeUrl('admin', { adminTab: index }))}" onclick="return selectAdminTabNav(event, ${index})" ${t === index ? 'aria-current="page"' : ''}>
      ${ic(section.icon, 16)}<span>${esc(section.label)}</span>
    </a>`).join('')}

    <div class="np-slot-container" data-np-slot="nav:bottom"></div>
    <span class="sp"></span>

    <a class="chip2 ${S.view === 'acc' ? 'on' : ''}" href="${esc(accHref)}" onclick="return handleNavClick(event, 'acc')" aria-label="Account" ${S.view === 'acc' ? 'aria-current="page"' : ''}>
      ${av(34)}
      <span class="who" title="${esc(S.acct.em || S.user?.email || '')}">
        <b>${esc(S.acct.fn ? (S.acct.fn + ' ' + (S.acct.ln || '')) : (S.acct.un || S.user?.username || 'Account'))}</b>
        <span class="mu pv">${esc(S.acct.em || S.user?.email || 'Administrator')}</span>
      </span>
      ${ic('settings', 16)}
    </a>
    <div class="nav-footer" style="padding: 10px 14px 4px; font-size: 11px; color: var(--tx-muted); text-align: center;">${renderFooterHtml()}</div>`;
  }

  // 2. Global Navigation (Servers, Activity, Client Area, Account)
  const showGlobalNav = ['dash', 'client', 'notfound', 'mod', 'maintenance', 'acc', 'create', 'admin'].includes(S.view) ||
    (S.view === 'act' && !window.location.pathname.startsWith('/server/')) ||
    !curServer();

  if (showGlobalNav) {
    const isActGlobal = S.view === 'act';
    return `
    <a class="brand" href="${esc(brandHref)}" onclick="return handleNavClick(event, 'dash')" style="cursor:pointer; text-decoration:none; color:inherit;">
      ${brandLogoHtml(false)}
    </a>
    
    <button class="nsearch" onclick="openPalette(1)" aria-label="Search (Ctrl+K)">
      ${ic('search', 15)}
      <span>Search</span>
      <kbd>Ctrl K</kbd>
    </button>

    <a class="nav-item ${S.view === 'dash' ? 'on' : ''}" href="${esc(brandHref)}" onclick="return handleNavClick(event, 'dash')" ${S.view === 'dash' ? 'aria-current="page"' : ''}>
      ${ic('server', 16)}
      <span>Servers</span>
    </a>
    <a class="nav-item ${isActGlobal ? 'on' : ''}" href="${esc(withSafeModeQuery('/activity'))}" onclick="return handleNavClick(event, 'act', { global: true })" ${isActGlobal ? 'aria-current="page"' : ''}>
      ${ic('activity', 16)}
      <span>Activity</span>
    </a>
    ${b.show_client_area_button !== false ? `
    <a class="nav-item ${S.view === 'client' ? 'on' : ''}" href="${esc(routeUrl('client'))}" onclick="return handleNavClick(event, 'client')" ${S.view === 'client' ? 'aria-current="page"' : ''}>
      ${ic('store', 16)}
      <span>${esc(clientLabel)}</span>
      ${clientStatusBadge}
    </a>` : ''}

    ${renderModSidebarItemsHtml('global')}
    <div class="np-slot-container" data-np-slot="nav:bottom"></div>

    <span class="sp"></span>

    ${S.user?.role === 'admin' ? `
    <a class="nav-item" href="${esc(adminHref)}" onclick="return handleNavClick(event, 'admin')" style="margin-bottom:6px;">
      ${ic('shield', 16)}
      <span>Admin</span>
    </a>` : ''}

    <a class="chip2 ${S.view === 'acc' ? 'on' : ''}" href="${esc(accHref)}" onclick="return handleNavClick(event, 'acc')" aria-label="Account" ${S.view === 'acc' ? 'aria-current="page"' : ''}>
      ${av(34)}
      <span class="who" title="${esc(S.acct.em || S.user?.email || '')}">
        <b>${esc(S.acct.fn ? (S.acct.fn + ' ' + (S.acct.ln || '')) : (S.acct.un || S.user?.username || 'Account'))}</b>
        <span class="mu pv">${esc(S.acct.em || S.user?.email || (S.user?.role === 'admin' ? 'Administrator' : 'Standard User'))}</span>
      </span>
      ${ic('settings', 16)}
    </a>
    <div class="nav-footer" style="padding: 10px 14px 4px; font-size: 11px; color: var(--tx-muted); text-align: center;">${renderFooterHtml()}</div>`;
  }

  // 3. Inside a Server Workspace
  let group = '';
  return `
  <a class="brand" href="${esc(brandHref)}" onclick="return handleNavClick(event, 'dash')" style="cursor:pointer; text-decoration:none; color:inherit;">
    ${brandLogoHtml(false)}
  </a>
  
  <button class="nsearch" onclick="openPalette(1)" aria-label="Search (Ctrl+K)">
    ${ic('search', 15)}
    <span>Search</span>
    <kbd>Ctrl K</kbd>
  </button>

  <a class="nb" href="${esc(brandHref)}" onclick="return handleNavClick(event, 'dash')" style="margin-bottom: 6px; text-decoration:none;">
    &larr; All servers
  </a>

  ${(() => {
    const isModSrv = isModdedEngine(curServer()?.type || S.core);
    return PG.map(([key, label, iconName, grp]) => {
      const header = grp !== group ? `<div class="ng">${grp}</div>` : '';
      group = grp;
      const active = S.view === key ? 'on' : '';
      const href = routeUrl(key);
      const displayLabel = (key === 'plug' && isModSrv) ? 'Mods' : label;
      return `${header}
    <a class="nav-item ${active}" ${canServerPage(key) ? `href="${esc(href)}"` : 'aria-disabled="true" tabindex="0" title="Access restricted — permission required"'} onclick="return handleNavClick(event, '${key}')" ${active ? 'aria-current="page"' : ''}>
      ${ic(iconName, 16)}
      <span>${displayLabel}</span>
    </a>`;
    }).join('');
  })()}

  ${renderModSidebarItemsHtml('server')}
  <div class="np-slot-container" data-np-slot="nav:bottom"></div>

  <span class="sp"></span>

  <a class="chip2 ${S.view === 'acc' ? 'on' : ''}" href="${esc(accHref)}" onclick="return handleNavClick(event, 'acc')" aria-label="Account" ${S.view === 'acc' ? 'aria-current="page"' : ''}>
    ${av(34)}
    <span class="who" title="${esc(S.acct.em || S.user?.email || '')}">
      <b>${esc(S.acct.fn ? (S.acct.fn + ' ' + (S.acct.ln || '')) : (S.acct.un || S.user?.username || 'Account'))}</b>
      <span class="mu pv">${esc(S.acct.em || S.user?.email || (S.user?.role === 'admin' ? 'Administrator' : 'Standard User'))}</span>
    </span>
    ${ic('settings', 16)}
  </a>
  <div class="nav-footer" style="padding: 10px 14px 4px; font-size: 11px; color: var(--tx-muted); text-align: center;">${renderFooterHtml()}</div>`;
}

// --- View Router & Multi-Page Navigation Engine ---
function normalizeRoute(view) {
  if (view === 'store') {
    S.plugTab = 'browse';
    return 'plug';
  }
  if (view === 'mods') {
    return 'plug';
  }
  if (view === 'stp' || view === 'startup') {
    return 'set';
  }
  if (view === 'all') {
    return 'ver';
  }
  return view;
}

window.go = function(view, opts = {}) {
  const target = normalizeRoute(view);
  if (serverPagePermission(target) && !canServerPage(target)) return;
  if (S.view === 'admin' && target !== 'admin' && (S.adminTab || 0) === 4) {
    applyBranding(S.branding);
  }
  applyRoute(routeUrl(target, opts), opts);
};

window.addEventListener('hashchange', () => {
  if (new URLSearchParams(window.location.hash.slice(1)).has('invite-token')) { checkAuth(); return; }
  checkUrlResetToken();
  if (!S.token || $('#login')) return;
  const raw = window.location.hash.replace('#', '');
  const hash = normalizeRoute(raw);
  if (hash && V[hash] && hash !== 'notfound' && hash !== 'maintenance') {
    go(hash);
  }
});

// --- Main Render Function (R) ---
window.R = function() {
  captureEditor();
  if (typeof filePreviewActive === 'function' && !filePreviewActive()) disposeFilePreview();
  closeAllCustomSelects();
  const nav = $('#nav');
  const main = $('#main');
  const prevNavScroll = nav ? nav.scrollTop : 0;
  const prevMainScroll = main ? main.scrollTop : 0;

  const isMaintEnabled = Boolean(S.branding?.maintenance_enabled);
  const isAdmin = S.user?.role === 'admin';
  const showMaintLockout = (isMaintEnabled && !isAdmin) || (S.maintPreview && isAdmin);

  if (showMaintLockout) {
    if (typeof disposeFilePreview === 'function') disposeFilePreview();
    if (nav) nav.innerHTML = '';
    if (main) {
      main.innerHTML = V.maintenance();
      main.scrollTop = 0;
    }
    const fly = $('#fly-console-root');
    if (fly) {
      fly.innerHTML = '';
      delete fly.dataset.flySig;
    }
    return;
  }

  if (nav) {
    nav.innerHTML = navHtml();
    nav.scrollTop = prevNavScroll;
  }
  if (main) {
    const maintBannerHtml = (isMaintEnabled && isAdmin)
      ? `<div class="np-maint-banner">
          <span><b>Maintenance Mode Active:</b> Standard users &amp; visitors see the maintenance lockout screen. Administrators retain full panel access.</span>
          <div class="row" style="gap: 8px;">
            <button type="button" class="btn sm" onclick="toggleMaintPreview(true)">${ic('eye', 13)} Preview Screen</button>
            <a href="/admin/branding" onclick="return handleNavClick(event, 'admin', { adminTab: 4 })" class="btn sm pr" style="text-decoration:none;">Manage</a>
          </div>
        </div>`
      : '';
    const viewFn = V[S.view] || V.notfound;
    const blocked = serverPagePermission(S.view) && !canServerPage(S.view);
    if (blocked && typeof disposeFilePreview === 'function') disposeFilePreview();
    const retainedPreviewMedia = typeof captureFilePreviewMedia === 'function' ? captureFilePreviewMedia() : null;
    main.innerHTML = renderAnnouncement() + maintBannerHtml + (blocked ? renderAccessRestricted() : (S.view === 'admin' && !isAdmin ? renderAccessRestricted(false) : viewFn())) + (S.view === 'dash' ? renderDiscordCard() : '');
    if (typeof restoreFilePreviewMedia === 'function') restoreFilePreviewMedia(retainedPreviewMedia);
    if (S.view === 'set' && !blocked) organizeServerSettings(main);
    applyPermissionControls(main);
    main.scrollTop = prevMainScroll;
    if (S.view === 'con') {
      paintLog();
      bindConsoleInput();
      updateInputState();
      updateWsLivePill();
    } else if (S.view === 'fi') {
      bindEditorTabKey();
    }
    enhanceSelects(main);
  }
  renderFlyConsole();
  if (typeof mountPanelModHooks === 'function') {
    mountPanelModHooks();
  }
  if (window.NP && typeof window.NP._emit === 'function') {
    window.NP._emit('render', { view: S.view, server: curServer() });
  }
};

function bindConsoleInput() {
  const inp = $('#cmd');
  if (!inp) return;
  inp.addEventListener('keydown', e => {
    if (e.key === 'Enter' && inp.value.trim()) {
      const c = inp.value.trim();
      runCmd(c);
      S.hist.push(c);
      S.hi = S.hist.length;
      inp.value = '';
    } else if (e.key === 'ArrowUp' && S.hi > 0) {
      inp.value = S.hist[--S.hi];
    } else if (e.key === 'ArrowDown') {
      inp.value = S.hist[++S.hi] || '';
    }
  });
}

function bindEditorTabKey() { bindQolEditor(); }

// --- View Transition Lifecycle Hook & Initial Route Preloader ---
async function loadInitialRouteData(view) {
  if (serverPagePermission(view)) {
    await ensureServerAccess();
    if (!canServerPage(view) || S.view !== view) return;
    if (hasServerPermission('console.view')) connectWs(curServer()?.id, false, curServer()?.node_id || 'local');
  }
  if (view === 'create') { if (canCreateServers()) openCreateServerModal(); return; }
  if (view === 'sftp') { await loadSFTPDetails(true); return; }
  const jobs = [];
  if (view === 'pl') jobs.push(loadPlayersData(true));
  else if (view === 'fi') jobs.push(loadFileList(S.currentDir, true));
  else if (view === 'db') jobs.push(loadDatabases(true, true));
  else if (view === 'plug') {
    jobs.push(loadPluginsList(true));
    jobs.push(loadVersionsCatalog());
    if ((S.plugTab || 'browse') === 'browse') jobs.push(searchModrinth(true));
  }
  else if (view === 'bk') jobs.push(loadBackupsList(true));
  else if (view === 'sc') jobs.push(loadServerSchedules(curServer()?.id, true));
  else if (view === 'team') jobs.push(loadTeamPage(true));
  else if (view === 'set' || view === 'stp') {
    jobs.push(loadProperties(true));
    jobs.push(loadVersionsCatalog());
    // SFTP loads only on its dedicated, separately authorized page.
    if (!isSteelEngine(curServer()?.type)) jobs.push(loadJavaRuntimes(false, curServer()?.node_id || 'local', true));
  }
  else if (view === 'ver') {
    jobs.push(loadVersionsCatalog());
    if (!isSteelEngine(curServer()?.type)) jobs.push(loadJavaRuntimes(false, curServer()?.node_id || 'local', true));
  }
  else if (view === 'sub') jobs.push(loadSubdomainData(true));
  else if (view === 'net') {
    // Network allocations do not require SFTP credentials.
    jobs.push(loadServerPorts(true));
  }
  else if (view === 'act') jobs.push(loadActivityLog(S.actFilter?.limit || 200, true));
  else if (view === 'admin') jobs.push(loadAdminData(true));
  else if (view === 'acc') jobs.push(loadAccountData(true));
  else if (view === 'con' && hasServerPermission('settings.view')) jobs.push(loadServerPorts(true));
  await Promise.all(jobs).catch(() => {});
}

window.onEnterView = function(view) {
  if (serverPagePermission(view) && !canServerPage(view)) {
    ensureServerAccess().then(() => { if (S.view === view) { R(); if (canServerPage(view)) onEnterView(view); } });
    return;
  }
  if (view === 'create') { if (canCreateServers()) openCreateServerModal(); return; }
  if (serverPagePermission(view) && hasServerPermission('console.view')) connectWs(curServer()?.id, false, curServer()?.node_id || 'local');
  if (view === 'sftp') { loadSFTPDetails(); return; }
  if (view === 'pl') loadPlayersData();
  else if (view === 'fi') loadFileList(S.currentDir);
  else if (view === 'db') loadDatabases(Array.isArray(S.databases) && S.databases.length > 0);
  else if (view === 'plug') {
    loadPluginsList();
    loadVersionsCatalog();
    if ((S.plugTab || 'browse') === 'browse') searchModrinth();
  }
  else if (view === 'bk') loadBackupsList();
  else if (view === 'sc') loadServerSchedules();
  else if (view === 'team') loadTeamPage();
  else if (view === 'set' || view === 'stp') {
    loadProperties();
    loadVersionsCatalog();
    // SFTP is intentionally not prefetched.
    if (!isSteelEngine(curServer()?.type)) loadJavaRuntimes(false, curServer()?.node_id || 'local');
  }
  else if (view === 'ver') { loadVersionsCatalog(); if (!isSteelEngine(curServer()?.type)) loadJavaRuntimes(false, curServer()?.node_id || 'local'); }
  else if (view === 'sub') loadSubdomainData();
  else if (view === 'net') { loadServerPorts(); }
  else if (view === 'act') loadActivityLog();
  else if (view === 'admin') loadAdminData();
  else if (view === 'acc') loadAccountData();
  else if (view === 'dash') loadDashboard();
};

function setNodeFilter(nodeId) {
  S.nodeFilter = nodeId || 'all';
  R();
}

// --- Backend Data Loaders ---
let dashboardLoadSequence = 0;
async function loadDashboard(skipRender = false, preferredRef = null) {
  const sequence = ++dashboardLoadSequence;
  const token = S.token;
  if (!S.loading) S.loading = {};
  S.loading.dash = true;
  S.serverCatalog = { ...S.serverCatalog, loading: true };
  // Bound hung network requests as well as the number of route recovery attempts.
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), 30000);
  const request = endpoint => api(endpoint, { signal: controller.signal, suppressToast: true });
  const live = () => sequence === dashboardLoadSequence && token === S.token;
  const metadataJobs = [];
  try {
    const clusterJob = request('/api/cluster/servers').then(value => ({ value }), error => ({ error }));
    // Optional metadata must not hold up server routing or invalidate the inventory.
    metadataJobs.push(request('/api/nodes').then(nodes => {
      if (live() && Array.isArray(nodes)) {
        S.nodes = nodes;
        if (!skipRender && S.view === 'dash') R();
      }
    }).catch(() => {}));
    metadataJobs.push(request(`/api/account/activity?limit=${S.actFilter?.limit || 200}`).then(act => {
      if (live() && Array.isArray(act)) S.acct.act = act;
    }).catch(() => {}));
    const { value: clusterRes, error: clusterError } = await clusterJob;
    const clusterComplete = Array.isArray(clusterRes?.servers);
    const denied = [401, 403].includes(clusterError?.status);
    let servers = clusterComplete ? clusterRes.servers : [];
    let localComplete = clusterComplete;
    if (!clusterComplete && !denied && live()) {
      try {
        const local = await request('/api/servers');
        localComplete = Array.isArray(local);
        if (localComplete) servers = local.map(s => ({ ...s, node_id: 'local' }));
      } catch (error) { localComplete = [401, 403].includes(error.status); }
    }
    if (!live()) return;
    // Capture selection at commit time, not before awaiting the network.
    const current = curServer();
    const route = parseCurrentRoute();
    const ref = route.serverId ? { id: route.serverId, nodeId: route.nodeId } :
      (current ? { id: current.id, nodeId: current.node_id || 'local' } : preferredRef);
    S.servers = servers.map(s => ({ ...s, node_id: s.node_id || 'local' }));
    S.clusterNodeErrors = clusterComplete ? (clusterRes.node_errors || {}) : {};
    S.serverCatalog = { clusterComplete, localComplete, denied, nodeErrors: S.clusterNodeErrors, loading: false };
    // Never restore unverified cached servers or select the first server for a missing deep link.
    const selected = (ref && resolveServerByParam(ref.id, ref.nodeId || null)) ||
      (!route.serverId ? S.servers[0] : null);
    if (selected) syncActiveServerState(selected);
    else clearActiveServerSelection();
  } finally {
    // Let optional metadata finish independently, still subject to the deadline.
    Promise.allSettled(metadataJobs).then(() => { clearTimeout(timer); controller.abort(); });
    if (live()) {
      S.loading.dash = false;
      S.serverCatalog.loading = false;
      if (!skipRender) {
        if (parseCurrentRoute().serverId) {
          applyRoute(window.location.pathname + window.location.search, {
            replace: true, retryAttempts: S.serverRoutePending?.attempts || 0
          });
        } else R();
      }
    }
  }
}

async function loadAdminNodes(toastOnDone = false) {
  try {
    const nodes = await api('/api/nodes');
    if (Array.isArray(nodes)) {
      S.nodes = nodes;
    }
    if (toastOnDone) toast('Cluster node status refreshed', 'ok');
  } catch (e) {}
  if (S.view === 'admin' || S.view === 'dash') R();
}

function buildClusterNodeInstallCmd() {
  const h = S.nodePairHelper || {};
  const origin = window.location.origin;
  const name = (h.name || 'mc-node-01').trim() || 'mc-node-01';
  const loc = (h.location || '').trim();
  const port = String(h.port || '8080').replace(/[^0-9]/g, '') || '8080';
  const key = (h.apiKey || '').trim() || '<ADMIN_API_KEY>';
  const q = s => `"${String(s).replace(/["\\$`]/g, '\\$&')}"`;
  let cmd = `curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash -s -- --node --master ${q(origin)} --api-key ${q(key)} --name ${q(name)}`;
  if (port && port !== '8080') cmd += ` --port ${port}`;
  if (loc) cmd += ` --location ${q(loc)}`;
  return cmd;
}

function toggleClusterNodeHelper() {
  S.nodePairOpen = !S.nodePairOpen;
  R();
}

function onNodeHelperInput(field, val) {
  if (!S.nodePairHelper) S.nodePairHelper = { name: 'mc-node-01', location: '', port: '8080', apiKey: '' };
  S.nodePairHelper[field] = val;
  const el = $('#node-install-cmd-text');
  if (el) el.textContent = buildClusterNodeInstallCmd();
}

async function generateNodePairingKey(btnEl) {
  if (!S.nodePairHelper) S.nodePairHelper = { name: 'mc-node-01', location: '', port: '8080', apiKey: '' };
  if (btnEl) setBtnLoading(btnEl, true, 'Generating...');
  try {
    const label = `Cluster Node (${(S.nodePairHelper.name || 'mc-node-01').trim() || 'mc-node-01'})`;
    const res = await api('/api/account/api-keys', {
      method: 'POST',
      body: JSON.stringify({ name: label })
    });
    const tok = res?.token || res?.key || '';
    if (!tok) {
      toast('Server did not return an API key token', 'bad');
      return;
    }
    S.nodePairHelper.apiKey = tok;
    toast('Admin API key generated for cluster node pairing!', 'ok');
    R();
  } catch (e) {} finally {
    if (btnEl) setBtnLoading(btnEl, false);
  }
}

function selectAdminTab(idx) {
  if ((S.adminTab || 0) === 4 && idx !== 4) {
    applyBranding(S.branding);
  }
  applyRoute(routeUrl('admin', { adminTab: idx }));
}

function selectAccountTab(idx) {
  applyRoute(routeUrl('acc', { acctTab: idx }));
}

async function loadAdminData(skipRender = false) {
  if (S.user?.role !== 'admin') return;
  await Promise.all([loadPanelRoles(true), loadCommunitySettings(true)]);
  const [sysRes, usersRes, tsRes, nodesRes, brandRes, actRes, modsRes, upRes, smtpRes] = await Promise.all([
    api('/api/system').catch(() => null),
    api('/api/admin/users').catch(() => null),
    api('/api/admin/turnstile').catch(() => null),
    api('/api/nodes').catch(() => null),
    api('/api/admin/branding').catch(() => null),
    api(`/api/account/activity?limit=${S.actFilter?.limit || 200}`).catch(() => null),
    api('/api/admin/mods').catch(() => null),
    api('/api/admin/updates').catch(() => null),
    api('/api/admin/email').catch(() => null),
    loadAdminPanelUrl(true)
  ]);

  if (sysRes) S.system = sysRes;
  if (usersRes) S.adminUsers = usersRes || [];
  if (tsRes) S.adminTurnstile = tsRes;
  if (Array.isArray(nodesRes)) S.nodes = nodesRes;
  if (upRes) {
    S.adminUpdate = upRes;
    if (upRes.is_updating) {
      startUpdateProgressPolling();
    }
  }
  if (smtpRes) S.adminSmtp = smtpRes;
  if (modsRes && Array.isArray(modsRes.mods)) {
    S.adminMods = modsRes.mods;
    if (modsRes.mods_dir) S.modsDir = modsRes.mods_dir;
    if (!S.selectedModId && S.adminMods.length > 0 && !S.modDraft) {
      selectAdminMod(S.adminMods[0].id, true);
    }
  }
  if (brandRes && brandRes.branding) {
    S.branding = { ...DEFAULT_BRANDING, ...brandRes.branding };
    S.brandingLoaded = true;
    S.brandDraft = { ...DEFAULT_BRANDING, ...S.branding };
  }
  if (brandRes && brandRes.daemon) {
    S.brandingDaemon = { ...S.brandingDaemon, ...brandRes.daemon };
    S.daemonDraft = { ...S.brandingDaemon };
    S.adminDaemonCfg = {
      brandName: S.branding.brand_name || 'NeoPanel',
      fqdn: S.brandingDaemon.fqdn || window.location.hostname || '127.0.0.1',
      sftpPort: S.brandingDaemon.sftp_port || 2022,
      maxUploadMB: S.brandingDaemon.max_upload_mb || 500,
      defaultRamMB: S.brandingDaemon.default_ram_mb || 2048,
      autoRestartPolicy: Boolean(S.brandingDaemon.auto_restart_policy)
    };
  }
  if (Array.isArray(actRes)) S.acct.act = actRes;

  applyBranding((S.adminTab || 0) === 4 && S.brandDraft ? S.brandDraft : S.branding);
  await loadJavaRuntimes(true, 'local', skipRender);
  if (!skipRender) R();
}

function openNodeModal(nodeId = null) {
  const container = $('#modal-container');
  if (!container) return;
  const existing = nodeId ? getNodeById(nodeId) : null;
  const isEdit = Boolean(existing && !existing.is_local);

  container.innerHTML = `<div class="pal" onclick="if(event.target===this)closeModal()">
    <div class="card pbox" style="padding:26px; max-width:540px; max-height:90vh; overflow-y:auto;">
      <div class="row sb" style="margin-bottom:14px;">
        <h2 style="font-size:20px; margin:0; display:flex; align-items:center; gap:8px;">
          ${ic('network', 18)} ${isEdit ? `Edit Cluster Node: ${esc(existing.name)}` : 'Connect Remote NeoPanel Node'}
        </h2>
        <button class="btn sm" onclick="closeModal()" aria-label="Close modal">&times;</button>
      </div>

      <p class="mu" style="font-size:13px; line-height:1.5; margin-bottom:16px;">
        Enter the reachable HTTP/HTTPS URL of your remote NeoPanel daemon along with its Node Token (<code>neo_node_...</code> via <code>sudo neopanel node-token</code> on cluster nodes) or an admin <code>neo_...</code> API key.
      </p>

      <form onsubmit="submitNodeModal(event, '${isEdit ? esc(existing.id) : ''}')">
        <div class="grid g2">
          <label class="f">
            <span>Node Name</span>
            <input type="text" id="nm-name" required value="${esc(existing?.name || '')}" placeholder="e.g. EU-Frankfurt-01">
          </label>
          <label class="f">
            <span>Region / Location</span>
            <input type="text" id="nm-loc" value="${esc(existing?.location || '')}" placeholder="e.g. Frankfurt, DE">
          </label>
        </div>

        <label class="f">
          <span>Node address</span>
          <input type="url" id="nm-url" required value="${esc(existing?.base_url || '')}" placeholder="http://192.168.1.50:8080 or https://node2.example.com">
        </label>

        <label class="f">
          <span>Public Game IP / FQDN <span class="mu">(displayed to players, optional)</span></span>
          <input type="text" id="nm-fqdn" value="${esc(existing?.fqdn || '')}" placeholder="e.g. eu1.example.com (defaults to URL host)">
        </label>

        <label class="f">
          <span>Node Auth Token ${isEdit ? '<span class="mu">(leave blank to keep current)</span>' : ''}</span>
          <input type="password" id="nm-token" ${isEdit ? '' : 'required'} autocomplete="new-password" placeholder="${isEdit ? `Saved (${esc(existing.token_prefix || 'neo_...')})` : 'neo_node_... or neo_...'}">
        </label>

        <!-- Free Node Idle Player Timeout -->
        <div style="background: rgba(255,255,255,0.02); border: 1px solid var(--border); border-radius: 8px; padding: 12px; margin-bottom: 12px;">
          <label class="row sb" style="cursor: pointer; margin-bottom: 8px;">
            <div>
              <b style="color: var(--tx); display: block;">Free Node (Idle Player Auto-Shutdown)</b>
              <span class="mu" style="font-size: 12px;">Automatically stops running servers when 0 players are connected for the configured duration</span>
            </div>
            <span class="sw">
              <input type="checkbox" id="nm-free" ${existing?.is_free ? 'checked' : ''} onchange="$('#nm-timeout-wrap').style.display = this.checked ? 'block' : 'none'">
              <i></i>
            </span>
          </label>
          <div id="nm-timeout-wrap" style="display: ${existing?.is_free ? 'block' : 'none'}; margin-top: 8px;">
            <label class="f" style="margin-bottom: 0;">
              <span>Idle Timeout (Minutes with 0 Players, default 15)</span>
              <input type="number" id="nm-idle-timeout" value="${existing?.idle_timeout_min || 15}" min="1" max="1440" placeholder="15">
            </label>
          </div>
        </div>

        <!-- Node Permissions Policy -->
        <div style="background: rgba(255,255,255,0.02); border: 1px solid var(--border); border-radius: 8px; padding: 12px; margin-bottom: 12px;">
          <div style="font-size: 12px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: var(--tx-muted); margin-bottom: 8px;">Non-Admin Permissions Policy</div>
          <label class="row sb" style="cursor: pointer; margin-bottom: 6px;">
            <span style="font-size: 13px; color: var(--tx);">Allow users to change memory limits</span>
            <span class="sw">
              <input type="checkbox" id="nm-allow-ram" ${existing?.allow_user_ram !== false ? 'checked' : ''}>
              <i></i>
            </span>
          </label>
          <label class="row sb" style="cursor: pointer; margin-bottom: 6px;">
            <span style="font-size: 13px; color: var(--tx);">Allow users to change server ports</span>
            <span class="sw">
              <input type="checkbox" id="nm-allow-ports" ${existing?.allow_user_ports !== false ? 'checked' : ''}>
              <i></i>
            </span>
          </label>
          <label class="row sb" style="cursor: pointer;">
            <span style="font-size: 13px; color: var(--tx);">Allow users to change Java versions</span>
            <span class="sw">
              <input type="checkbox" id="nm-allow-java" ${existing?.allow_user_java !== false ? 'checked' : ''}>
              <i></i>
            </span>
          </label>
        </div>

        <!-- Process Isolation -->
        <div style="background: rgba(255,255,255,0.02); border: 1px solid var(--border); border-radius: 8px; padding: 12px; margin-bottom: 14px;">
          <div style="font-size: 12px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: var(--tx-muted); margin-bottom: 8px;">Security &amp; Sandboxing</div>
          <label class="f" style="margin-bottom: 0;">
            <span>Process Isolation Engine</span>
            <select id="nm-isolation">
              <option value="bubblewrap" ${(existing?.isolation_engine || 'bubblewrap') === 'bubblewrap' ? 'selected' : ''}>Bubblewrap (bwrap jailed filesystem + systemd slice)</option>
              <option value="none" ${existing?.isolation_engine === 'none' ? 'selected' : ''}>Host Direct (No filesystem jail)</option>
            </select>
          </label>
        </div>

        <div id="node-modal-test-res" style="margin-bottom: 14px;"></div>

        <div class="row sb" style="gap:10px; margin-top:6px;">
          <button type="button" id="nm-test-btn" class="btn" onclick="testNodeModalConnection('${isEdit ? esc(existing.id) : ''}')">
            ${ic('refresh', 14)} Test Connection
          </button>
          <div class="row" style="gap:8px;">
            <button type="button" class="btn" onclick="closeModal()">Cancel</button>
            <button type="submit" id="nm-submit-btn" class="btn pr">
              ${ic('check', 14)} ${isEdit ? 'Save Node' : 'Connect Node'}
            </button>
          </div>
        </div>
      </form>
    </div>
  </div>`;

  setTimeout(() => $('#nm-name')?.focus(), 40);
}

async function testNodeModalConnection(nodeId = '') {
  const base_url = $('#nm-url')?.value.trim() || '';
  const api_token = $('#nm-token')?.value.trim() || '';
  const box = $('#node-modal-test-res');
  const btn = $('#nm-test-btn');
  if (!base_url) {
    toast('Enter the Node API Base URL first', 'bad');
    return;
  }
  if (!api_token && !nodeId) {
    toast('Enter the Node API Token (neo_...)', 'bad');
    return;
  }

  if (btn) setBtnLoading(btn, true, 'Testing...');
  if (box) box.innerHTML = loaderHtml('Probing remote NeoPanel daemon...');
  try {
    const res = await api('/api/nodes/test', {
      method: 'POST',
      body: JSON.stringify({ id: nodeId || '', base_url, api_token })
    });
    if (res.online) {
      const sys = res.system || {};
      const ramGb = sys.total_ram_mb ? (sys.total_ram_mb / 1024).toFixed(1) + ' GB RAM' : '';
      box.innerHTML = `<div class="node-callout" style="margin:0; border-color: rgba(63,185,80,0.45); background: rgba(63,185,80,0.1);">
        <b style="color:var(--ok);">Connected!</b> Reachable in <b>${res.latency_ms ?? 0}ms</b> &bull; ${res.server_count ?? 0} server(s)
        ${sys.cpus ? `<div class="mu mono" style="font-size:11.5px; margin-top:3px;">${esc(sys.os || 'linux')}/${esc(sys.arch || 'amd64')} &bull; ${sys.cpus} vCPU &bull; ${ramGb}</div>` : ''}
      </div>`;
    } else {
      box.innerHTML = `<div class="node-error-note"><b>Connection Failed:</b> ${esc(res.error || 'Unreachable node')}</div>`;
    }
  } catch (err) {
    if (box) box.innerHTML = `<div class="node-error-note">${esc(err.message || 'Test failed')}</div>`;
  } finally {
    if (btn) setBtnLoading(btn, false);
  }
}

async function submitNodeModal(e, nodeId = '') {
  e.preventDefault();
  const btn = $('#nm-submit-btn');
  if (btn) setBtnLoading(btn, true, nodeId ? 'Saving...' : 'Connecting...');

  const payload = {
    name: $('#nm-name')?.value.trim() || '',
    location: $('#nm-loc')?.value.trim() || '',
    base_url: $('#nm-url')?.value.trim() || '',
    fqdn: $('#nm-fqdn')?.value.trim() || '',
    api_token: $('#nm-token')?.value.trim() || '',
    is_free: $('#nm-free') ? $('#nm-free').checked : false,
    idle_timeout_min: parseInt($('#nm-idle-timeout')?.value, 10) || 15,
    allow_user_ram: $('#nm-allow-ram') ? $('#nm-allow-ram').checked : true,
    allow_user_ports: $('#nm-allow-ports') ? $('#nm-allow-ports').checked : true,
    allow_user_java: $('#nm-allow-java') ? $('#nm-allow-java').checked : true,
    isolation_engine: $('#nm-isolation')?.value || 'bubblewrap'
  };

  try {
    const isEdit = Boolean(nodeId);
    const res = await api(isEdit ? `/api/nodes/${encodeURIComponent(nodeId)}` : '/api/nodes', {
      method: isEdit ? 'PUT' : 'POST',
      body: JSON.stringify(payload)
    });
    closeModal();
    if (res && res.online) {
      toast(`Node "${res.name}" connected (${res.latency_ms ?? 0}ms)!`, 'ok');
    } else {
      toast(`Saved node "${res?.name || payload.name}"${res?.error ? ` — warning: ${res.error}` : ''}`);
    }
    await Promise.all([loadDashboard(), loadAdminNodes()]);
  } catch (err) {
    if (btn) setBtnLoading(btn, false);
  }
}

async function testClusterNode(nodeId, btnEl) {
  if (btnEl) setBtnLoading(btnEl, true, 'Testing...');
  try {
    const res = await api(`/api/nodes/${encodeURIComponent(nodeId)}/test`, { method: 'POST' });
    if (res.online) {
      toast(`Node online (${res.latency_ms ?? 0}ms, ${res.server_count ?? 0} servers)`, 'ok');
    } else {
      toast(`Node offline: ${res.error || 'Unreachable'}`, 'bad');
    }
    await loadAdminNodes();
  } catch (e) {} finally {
    if (btnEl) setBtnLoading(btnEl, false);
  }
}

async function deleteClusterNode(nodeId, name) {
  const ok = await uiConfirm({
    title: 'Disconnect Cluster Node',
    message: `Remove remote node <b>${esc(name)}</b> from this NeoPanel cluster? Servers hosted on that remote daemon will keep running there untouched, and will simply disappear from this master dashboard.`,
    confirmText: 'Disconnect Node',
    cancelText: 'Cancel',
    isDanger: true,
    icon: 'network'
  });
  if (!ok) return;

  try {
    await api(`/api/nodes/${encodeURIComponent(nodeId)}`, { method: 'DELETE' });
    if (S.nodeFilter === nodeId) S.nodeFilter = 'all';
    toast(`Disconnected node "${name}"`, 'ok');
    await Promise.all([loadDashboard(), loadAdminNodes()]);
  } catch (e) {}
}

async function bulkRestartServers() {
  toast('Restarting all cluster instances...');
  for (const s of S.servers) {
    await api(srvApiPath(s, '/power'), { method: 'POST', body: JSON.stringify({ action: 'restart' }) }).catch(() => {});
  }
  toast('All servers restarted');
}

async function bulkStopServers() {
  const count = S.servers.length;
  if (!count) return;
  const ok = await uiConfirm({
    title: 'Stop All Server Instances',
    message: `Are you sure you want to stop all <b>${count}</b> server instance${count > 1 ? 's' : ''} across the cluster? Connected players on all servers will be disconnected.`,
    confirmText: `Stop All (${count})`,
    cancelText: 'Cancel',
    isDanger: true,
    icon: 'alert'
  });
  if (!ok) return;

  toast('Stopping all cluster instances...');
  for (const s of S.servers) {
    await api(srvApiPath(s, '/power'), { method: 'POST', body: JSON.stringify({ action: 'stop' }) }).catch(() => {});
  }
  toast('All servers stopped');
}

async function loadPlayersData(skipRender = false) {
  const srv = curServer();
  if (!srv) return;
  if (!S.loading) S.loading = {};
  const hasCached = (S.players && S.players.length > 0) || (S.wl && S.wl.length > 0) || (S.bans && S.bans.length > 0) || (S.ops && S.ops.length > 0);
  if (!hasCached) {
    S.loading.players = true;
    if (!skipRender && S.view === 'pl') R();
  }
  try {
    const data = await api(srvApiPath(srv, '/players'));
    const rawOnline = Array.isArray(data?.online) ? data.online : [];
    S.players = rawOnline.map(p => ({
      ...p,
      n: p?.name || p?.n || 'Player',
      name: p?.name || p?.n || 'Player',
      t: p?.playtime || p?.t || 'Online',
      ip: p?.ip || '',
      ping: p?.ping || '',
      uuid: p?.uuid || '',
      op: Boolean(p?.op)
    }));
    S.playersMax = Number(data?.max_players) || S.props?.max || 20;
    S.playersOnlineCount = (data?.online_count !== undefined && data?.online_count !== null)
      ? Number(data.online_count)
      : S.players.length;
    S.wl = data?.whitelist || [];
    S.bans = data?.bans || [];
    S.ops = data?.ops || [];
  } catch (e) {} finally {
    S.loading.players = false;
    if (!skipRender && S.view === 'pl') R();
  }
}

async function playerAction(action, player, reason = '') {
  const srv = curServer();
  if (!srv) return;

  const actionMeta = {
    kick: {
      title: 'Kick Player',
      message: `Kick <b>${esc(player)}</b> from <b>${esc(srv.name)}</b>?`,
      confirmText: 'Kick Player',
      isDanger: true,
      icon: 'alert'
    },
    ban: {
      title: 'Ban Player',
      message: `Permanently ban <b>${esc(player)}</b> from <b>${esc(srv.name)}</b>? They will be disconnected and blocked from rejoining.`,
      confirmText: 'Ban Player',
      isDanger: true,
      icon: 'shield'
    },
    unban: {
      title: 'Pardon / Unban Player',
      message: `Remove the ban on <b>${esc(player)}</b> so they can rejoin <b>${esc(srv.name)}</b>?`,
      confirmText: 'Unban Player',
      isDanger: false,
      icon: 'check'
    },
    op: {
      title: 'Grant Operator (OP)',
      message: `Grant full server operator permissions to <b>${esc(player)}</b> on <b>${esc(srv.name)}</b>?`,
      confirmText: 'Grant OP',
      isDanger: false,
      icon: 'shield'
    },
    deop: {
      title: 'Revoke Operator (De-OP)',
      message: `Revoke server operator permissions from <b>${esc(player)}</b>?`,
      confirmText: 'Revoke OP',
      isDanger: true,
      icon: 'shield'
    },
    whitelist_remove: {
      title: 'Remove from Whitelist',
      message: `Remove <b>${esc(player)}</b> from the server whitelist?`,
      confirmText: 'Remove',
      isDanger: true,
      icon: 'trash'
    }
  }[action];

  if (actionMeta) {
    const ok = await uiConfirm({
      title: actionMeta.title,
      message: actionMeta.message,
      confirmText: actionMeta.confirmText,
      cancelText: 'Cancel',
      isDanger: actionMeta.isDanger,
      icon: actionMeta.icon
    });
    if (!ok) return;
  }

  try {
    await api(srvApiPath(srv, '/players/action'), {
      method: 'POST',
      body: JSON.stringify({ action, player, reason })
    });
    toast(`Action ${action} executed for ${player}`);
    loadPlayersData();
  } catch (e) {}
}

function addWhitelistPlayer() {
  const inp = $('#wn');
  if (!inp || !inp.value.trim()) return;
  playerAction('whitelist_add', inp.value.trim());
  inp.value = '';
}

// --- File Manager Operations ---
async function loadFileList(path = '', skipRender = false) {
  const srv = curServer();
  if (!srv) return;
  if (!S.loading) S.loading = {};
  const dirChanged = path !== S.currentDir;
  if (dirChanged) {
    ++QOL.openSequence;
    if (typeof disposeFilePreview === 'function') disposeFilePreview();
  }
  if (dirChanged && S.selectedFiles) {
    S.selectedFiles.clear();
  }
  if (dirChanged || !S.fileList || S.fileList.length === 0) {
    S.loading.files = true;
    if (!skipRender && S.view === 'fi') R();
  }
  try {
    const data = await api(srvApiPath(srv, `/files?path=${encodeURIComponent(path)}`));
    if (!sameQolServer(srv, curServer())) return;
    S.fileList = data || [];
    S.currentDir = path;
  } catch (e) {} finally {
    S.loading.files = false;
    if (!skipRender && S.view === 'fi') R();
  }
}

function browseFolder(dirName) {
  if (dirName === '..') {
    const parts = S.currentDir.split('/').filter(Boolean);
    parts.pop();
    loadFileList(parts.join('/'));
  } else {
    const newPath = S.currentDir ? `${S.currentDir}/${dirName}` : dirName;
    loadFileList(newPath);
  }
}

function formatFileDate(iso) {
  if (!iso) return '—';
  try {
    const d = new Date(iso);
    if (isNaN(d.getTime())) return '—';
    const now = new Date();
    const diffSec = Math.floor((now - d) / 1000);
    if (diffSec < 60) return 'Just now';
    if (diffSec < 3600) return `${Math.floor(diffSec / 60)} minutes ago`;
    if (diffSec < 86400) return `${Math.floor(diffSec / 3600)} hours ago`;
    const tz = getUserTimezone();
    return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric', timeZone: tz });
  } catch (e) {
    return '—';
  }
}

function formatFileSize(bytes, isDir) {
  if (isDir) return '—';
  if (bytes === 0) return '0 Bytes';
  if (!bytes) return '—';
  const units = ['Bytes', 'KiB', 'MiB', 'GiB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / Math.pow(1024, i)).toFixed(i === 0 ? 0 : 2)} ${units[i] || 'Bytes'}`;
}

function formatNum(n) {
  if (!n) return '0';
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M';
  if (n >= 1000) return (n / 1000).toFixed(1) + 'K';
  return String(n);
}

function filterServerCards(q) {
  const cards = $$('.sift-server-card');
  const term = (q || '').toLowerCase();
  cards.forEach(c => {
    const name = (c.getAttribute('data-name') || '').toLowerCase();
    const addr = (c.getAttribute('data-addr') || '').toLowerCase();
    c.style.display = (name.includes(term) || addr.includes(term)) ? '' : 'none';
  });
}

function filterFileList(q) {
  const rows = $$('.file-row');
  const term = (q || '').toLowerCase();
  rows.forEach(r => {
    const name = (r.getAttribute('data-name') || '').toLowerCase();
    r.style.display = name.includes(term) ? '' : 'none';
  });
}

function toggleFileCtx(e, idx) {
  if (e) e.stopPropagation();
  const targetId = `fctx-${idx}`;
  $$('.file-ctx-menu').forEach(m => {
    if (m.id !== targetId) m.style.display = 'none';
  });
  const m = $(`#${targetId}`);
  if (m) {
    m.style.display = m.style.display === 'block' ? 'none' : 'block';
  }
}

function closeAllCtx() {
  $$('.file-ctx-menu').forEach(m => m.style.display = 'none');
}
window.addEventListener('click', closeAllCtx);

async function closeEditor() {
  if (!await guardEditorLeave()) return;
  ++QOL.openSequence;
  QOL.editor = null;
  S.editingFile = null;
  R();
}

function setPlugTab(tab) {
  const nextTab = tab === 'installed' ? 'installed' : 'browse';
  applyRoute(routeUrl('plug', { plugTab: nextTab }));
}

function toggleCustomJavaInput(val) {
  const wrap = $('#stp-custom-wrap') || $('#prop-custom-wrap');
  if (wrap) wrap.style.display = (val === 'custom') ? 'block' : 'none';
}

function toggleCustomPropJava(val) {
  const wrap = $('#prop-custom-wrap');
  if (wrap) wrap.style.display = (val === 'custom') ? 'block' : 'none';
}

async function saveStartupSettings() {
  if (!$('#stp-java')) {
    return saveProperties();
  }
  const srv = curServer();
  if (!srv) return;

  let javaChoice = $('#stp-java')?.value;
  if (javaChoice === 'custom') {
    javaChoice = $('#stp-java-custom')?.value.trim();
  }
  const ver = $('#stp-ver')?.value?.trim() || srv.version || '1.21.11';
  const jar = $('#stp-jar')?.value?.trim() || srv.jar_file || 'server.jar';
  const args = $('#stp-args')?.value?.trim() || srv.java_args || '-XX:+UseG1GC';
  const newRam = parseInt($('#stp-ram')?.value, 10) || srv.memory_max || 2048;

  try {
    await api(srvApiPath(srv, ''), {
      method: 'PUT',
      body: JSON.stringify({
        name: srv.name,
        description: srv.description,
        type: srv.type,
        version: ver,
        jar_file: jar,
        port: srv.port,
        memory_min: Math.min(1024, newRam),
        memory_max: newRam,
        java_path: javaChoice,
        java_args: args,
        auto_restart: srv.auto_restart
      })
    });

    srv.version = ver;
    srv.jar_file = jar;
    srv.java_path = javaChoice;
    srv.java_args = args;
    srv.memory_max = newRam;
    srv.memory_min = Math.min(1024, newRam);
    S.ver = ver;

    toast('Startup variables & Java runtime saved to database! Restart server to apply.', 'ok');
    R();
  } catch (err) {
    toast(err.message || 'Failed to save startup configuration', 'bad');
    uiAlert({
      title: 'Failed to Save Startup Settings',
      message: `Could not save startup settings: ${esc(err.message || 'Invalid parameters')}`,
      type: 'error'
    });
  }
}

async function openFile(name) { return openQolFile(name); }
async function saveActiveFile() { return saveQolFile(); }

function downloadFileItem(path) {
  if (!hasServerPermission('files.view')) return;
  const srv = curServer();
  if (!srv) return;
  const baseUrl = srvApiPath(srv, `/files/download?path=${encodeURIComponent(path)}`);
  const sep = baseUrl.includes('?') ? '&' : '?';
  window.open(`${baseUrl}${sep}token=${encodeURIComponent(S.token)}`, '_blank');
}

function isArchiveFile(filename) {
  const f = (filename || '').toLowerCase();
  return f.endsWith('.zip') || f.endsWith('.tar.gz') || f.endsWith('.tgz') || f.endsWith('.tar') || f.endsWith('.jar');
}

async function unzipFileItem(path) {
  const srv = curServer();
  if (!srv) return;
  const fileName = path.split('/').pop();

  const dest = await uiPrompt({
    title: 'Extract Archive File',
    message: `Extract <b>${esc(fileName)}</b> to directory:`,
    defaultValue: S.currentDir || '',
    placeholder: 'Destination folder (leave blank for current folder)',
    confirmText: 'Extract Archive'
  });
  if (dest === null) return;

  toast(`Extracting ${fileName}...`);
  try {
    const res = await api(srvApiPath(srv, '/files/unzip'), {
      method: 'POST',
      body: JSON.stringify({ path: path, destination: dest.trim() })
    });
    toast(res.message || `Extracted ${fileName} successfully!`, 'ok');
    loadFileList(S.currentDir);
  } catch (err) {
    uiAlert({
      title: 'Extraction Error',
      message: err.message || 'Failed to extract archive.',
      type: 'error'
    });
  }
}

async function unzipSelectedArchive() {
  const archives = Array.from(S.selectedFiles).filter(p => isArchiveFile(p));
  if (archives.length > 0) {
    await unzipFileItem(archives[0]);
  }
}

function normalizeContainerDestPath(raw) {
  let d = String(raw || '').trim().replace(/\\/g, '/');
  d = d.replace(/^\/+home\/+container\/?/i, '');
  d = d.replace(/^\/+|\/+$/g, '');
  if (d === '.' || d === '~') d = '';
  return d;
}

function setMoveModalDest(destVal) {
  const inp = $('#move-dest-input');
  if (inp) {
    inp.value = normalizeContainerDestPath(destVal);
    inp.focus();
  }
  const errEl = $('#move-modal-err');
  if (errEl) errEl.style.display = 'none';
}

function openMoveFilesModal(pathsArg = null) {
  const paths = (Array.isArray(pathsArg) && pathsArg.length)
    ? pathsArg.map(p => String(p || '').trim()).filter(Boolean)
    : Array.from(S.selectedFiles || []);
  if (!paths.length) {
    toast('Select one or more files or folders first', 'bad');
    return;
  }

  const container = $('#modal-container');
  if (!container) return;

  S.moveModalState = { paths };
  const selectedSet = new Set(paths);
  const curDir = S.currentDir || '';
  const parts = curDir.split('/').filter(Boolean);
  const parentDir = parts.slice(0, -1).join('/');

  const chips = [];
  if (curDir) {
    chips.push({ label: '/ (Root)', dest: '' });
    if (parts.length > 1) {
      chips.push({ label: `.. (/${parentDir})`, dest: parentDir });
    }
  }
  (S.fileList || []).forEach(it => {
    if (!it.is_dir) return;
    const p = it.path || (curDir ? `${curDir}/${it.name}` : it.name);
    if (!selectedSet.has(p)) {
      chips.push({ label: `${it.name}/`, dest: p });
    }
  });

  const maxPreview = 6;
  const shown = paths.slice(0, maxPreview);
  const extraCount = Math.max(0, paths.length - maxPreview);
  const defaultDest = curDir ? (parentDir || '') : '';

  container.innerHTML = `<div class="pal" onclick="if(event.target===this)closeModal()">
    <div class="card pbox" style="padding:26px; max-width:560px; border:1px solid var(--border-light);">
      <div class="row sb" style="margin-bottom:14px;">
        <h2 style="font-size:19px; margin:0; color:#fff; display:flex; align-items:center; gap:8px;">
          ${ic('move', 18)} Move ${paths.length} Item${paths.length > 1 ? 's' : ''}
        </h2>
        <button class="btn sm" onclick="closeModal()" aria-label="Close modal">&times;</button>
      </div>

      <div class="move-item-chips">
        ${shown.map(p => `<span class="chip mono">${esc(p.split('/').pop() || p)}</span>`).join('')}
        ${extraCount > 0 ? `<span class="chip">+${extraCount} more</span>` : ''}
      </div>

      <form id="move-files-form" onsubmit="submitMoveFilesModal(event)">
        <label class="f" style="margin-bottom:10px;">
          <span>Destination Folder (relative to server root)</span>
          <div class="move-path-input-wrap">
            <span class="move-path-prefix mono">/home/container/</span>
            <input type="text" id="move-dest-input" value="${esc(defaultDest)}" placeholder="e.g. plugins, backups/old, or leave blank for root" autocomplete="off" spellcheck="false" oninput="const e=$('#move-modal-err');if(e)e.style.display='none';">
          </div>
        </label>

        ${chips.length > 0 ? `
          <div style="margin-bottom:14px;">
            <span class="mu" style="font-size:11.5px; display:block; margin-bottom:6px;">Quick destination folders:</span>
            <div class="row" style="gap:6px; flex-wrap:wrap;">
              ${chips.slice(0, 10).map(c => `<button type="button" class="dest-chip" data-dest="${esc(c.dest)}" onclick="setMoveModalDest(this.dataset.dest)">${ic('folder', 12)} ${esc(c.label)}</button>`).join('')}
            </div>
          </div>
        ` : `<p class="mu" style="font-size:12px; margin-bottom:14px;">Type a folder name (missing folders are created automatically) or leave empty for <code>/home/container/</code>.</p>`}

        <label class="row" style="cursor:pointer; gap:9px; font-size:13px; margin-bottom:14px;">
          <input type="checkbox" id="move-overwrite">
          <span>Overwrite existing files or folders if a conflict occurs</span>
        </label>

        <div id="move-modal-err" class="node-error-note" style="display:none; margin-bottom:14px;"></div>

        <div class="row" style="justify-content:flex-end; gap:10px;">
          <button type="button" class="btn" onclick="closeModal()">Cancel</button>
          <button type="submit" id="move-submit-btn" class="btn pr">${ic('move', 14)} Move Here</button>
        </div>
      </form>
    </div>
  </div>`;

  setTimeout(() => {
    const inp = $('#move-dest-input');
    if (inp) { inp.focus(); inp.select(); }
  }, 30);
}

async function submitMoveFilesModal(e) {
  e.preventDefault();
  const srv = curServer();
  const state = S.moveModalState;
  if (!srv || !state?.paths?.length) return;

  const btn = $('#move-submit-btn');
  const errEl = $('#move-modal-err');
  if (errEl) errEl.style.display = 'none';
  if (btn) setBtnLoading(btn, true, 'Moving...');

  const destination = normalizeContainerDestPath($('#move-dest-input')?.value || '');
  const overwrite = Boolean($('#move-overwrite')?.checked);

  try {
    assertClientRequestAccess(srvApiPath(srv, '/files/move'), { method: 'POST' });
    const res = await fetch(srvApiPath(srv, '/files/move'), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(S.token ? { 'Authorization': `Bearer ${S.token}` } : {})
      },
      body: JSON.stringify({
        paths: state.paths,
        destination,
        overwrite
      })
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      throw new Error(data.error || `HTTP ${res.status}`);
    }

    S.moveModalState = null;
    if (S.selectedFiles) S.selectedFiles.clear();
    closeModal();
    toast(data.message || `Moved ${state.paths.length} item(s)`, 'ok');
    loadFileList(S.currentDir);
  } catch (err) {
    if (errEl) {
      errEl.textContent = err.message || 'Failed to move selected items';
      errEl.style.display = 'block';
    }
    toast(err.message || 'Move failed', 'bad');
  } finally {
    const b = $('#move-submit-btn');
    if (b) setBtnLoading(b, false);
  }
}

function setCompressModalFormat(fmt) {
  const hidden = $('#comp-format');
  if (hidden) hidden.value = fmt;
  document.querySelectorAll('.fmt-pill').forEach(el => {
    el.classList.toggle('on', el.dataset.fmt === fmt);
  });
  const inp = $('#comp-filename');
  if (inp) {
    let base = inp.value.trim().replace(/\.(zip|tar\.gz|tgz)$/i, '') || 'archive';
    inp.value = fmt === 'tar.gz' ? `${base}.tar.gz` : `${base}.zip`;
  }
}

function openCompressFilesModal(pathsArg = null) {
  const paths = (Array.isArray(pathsArg) && pathsArg.length)
    ? pathsArg.map(p => String(p || '').trim()).filter(Boolean)
    : Array.from(S.selectedFiles || []);
  if (!paths.length) {
    toast('Select one or more files or folders to archive', 'bad');
    return;
  }

  const container = $('#modal-container');
  if (!container) return;

  const curDir = S.currentDir || '';
  S.compressModalState = { paths, dir: curDir };

  const now = new Date();
  const stamp = `${now.getFullYear()}${String(now.getMonth() + 1).padStart(2, '0')}${String(now.getDate()).padStart(2, '0')}-${String(now.getHours()).padStart(2, '0')}${String(now.getMinutes()).padStart(2, '0')}`;
  let defaultBase = `archive-${stamp}`;
  if (paths.length === 1) {
    const singleName = (paths[0].split('/').pop() || 'archive').replace(/\.(zip|tar\.gz|tgz)$/i, '');
    if (singleName) defaultBase = singleName;
  }

  const maxPreview = 6;
  const shown = paths.slice(0, maxPreview);
  const extraCount = Math.max(0, paths.length - maxPreview);
  const folderLabel = curDir ? `/home/container/${curDir}/` : '/home/container/';

  container.innerHTML = `<div class="pal" onclick="if(event.target===this)closeModal()">
    <div class="card pbox" style="padding:26px; max-width:540px; border:1px solid var(--border-light);">
      <div class="row sb" style="margin-bottom:14px;">
        <h2 style="font-size:19px; margin:0; color:#fff; display:flex; align-items:center; gap:8px;">
          ${ic('archive', 18)} Create Archive (${paths.length} item${paths.length > 1 ? 's' : ''})
        </h2>
        <button class="btn sm" onclick="closeModal()" aria-label="Close modal">&times;</button>
      </div>

      <div class="move-item-chips">
        ${shown.map(p => `<span class="chip mono">${esc(p.split('/').pop() || p)}</span>`).join('')}
        ${extraCount > 0 ? `<span class="chip">+${extraCount} more</span>` : ''}
      </div>

      <form id="compress-files-form" onsubmit="submitCompressFilesModal(event)">
        <input type="hidden" id="comp-format" value="zip">

        <div class="f" style="margin-bottom:12px;">
          <span>Archive Format</span>
          <div class="fmt-pills">
            <button type="button" class="fmt-pill on" data-fmt="zip" onclick="setCompressModalFormat('zip')">
              <b>.zip</b>
              <span class="mu">Universal Zip Archive</span>
            </button>
            <button type="button" class="fmt-pill" data-fmt="tar.gz" onclick="setCompressModalFormat('tar.gz')">
              <b>.tar.gz</b>
              <span class="mu">Pterodactyl Gzip Tarball</span>
            </button>
          </div>
        </div>

        <label class="f" style="margin-bottom:10px;">
          <span>Archive Filename (saved in <code class="mono">${esc(folderLabel)}</code>)</span>
          <input type="text" id="comp-filename" value="${esc(defaultBase)}.zip" required autocomplete="off" spellcheck="false" oninput="const e=$('#comp-modal-err');if(e)e.style.display='none';">
        </label>

        <div id="comp-modal-err" class="node-error-note" style="display:none; margin-bottom:14px;"></div>

        <div class="row" style="justify-content:flex-end; gap:10px; margin-top:16px;">
          <button type="button" class="btn" onclick="closeModal()">Cancel</button>
          <button type="submit" id="comp-submit-btn" class="btn pr">${ic('archive', 14)} Create Archive</button>
        </div>
      </form>
    </div>
  </div>`;

  setTimeout(() => {
    const inp = $('#comp-filename');
    if (inp) { inp.focus(); inp.select(); }
  }, 30);
}

async function submitCompressFilesModal(e) {
  e.preventDefault();
  const srv = curServer();
  const state = S.compressModalState;
  if (!srv || !state?.paths?.length) return;

  const btn = $('#comp-submit-btn');
  const errEl = $('#comp-modal-err');
  if (errEl) errEl.style.display = 'none';
  if (btn) setBtnLoading(btn, true, 'Archiving...');

  const filename = $('#comp-filename')?.value.trim() || '';
  const format = $('#comp-format')?.value || 'zip';

  try {
    assertClientRequestAccess(srvApiPath(srv, '/files/compress'), { method: 'POST' });
    const res = await fetch(srvApiPath(srv, '/files/compress'), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(S.token ? { 'Authorization': `Bearer ${S.token}` } : {})
      },
      body: JSON.stringify({
        paths: state.paths,
        destination: state.dir || '',
        filename,
        format
      })
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      throw new Error(data.error || `HTTP ${res.status}`);
    }

    S.compressModalState = null;
    if (S.selectedFiles) S.selectedFiles.clear();
    closeModal();
    toast(data.message || 'Archive created successfully!', 'ok');
    loadFileList(S.currentDir);
  } catch (err) {
    if (errEl) {
      errEl.textContent = err.message || 'Failed to create archive';
      errEl.style.display = 'block';
    }
    toast(err.message || 'Archive failed', 'bad');
  } finally {
    const b = $('#comp-submit-btn');
    if (b) setBtnLoading(b, false);
  }
}

async function promptRenameFile(oldPath) {
  const baseName = oldPath.split('/').pop();
  const newName = await uiPrompt({
    title: 'Rename File or Folder',
    message: `Enter new name for <b>${esc(baseName)}</b>:`,
    defaultValue: baseName,
    placeholder: 'Filename or path',
    confirmText: 'Rename'
  });
  if (!newName || newName === baseName) return;
  const parent = oldPath.split('/').slice(0, -1).join('/');
  const newPath = parent ? `${parent}/${newName}` : newName;
  const srv = curServer();
  if (!srv) return;
  try {
    await api(srvApiPath(srv, '/files/rename'), {
      method: 'POST',
      body: JSON.stringify({ old_path: oldPath, new_path: newPath })
    });
    toast(`Renamed to ${newName}`, 'ok');
    loadFileList(S.currentDir);
  } catch (e) {
    toast(`Rename failed: ${e.message}`, 'bad');
    uiAlert({
      title: 'Rename Failed',
      message: `Could not rename <b>${esc(baseName)}</b> to <b>${esc(newName)}</b>:<br><br>${esc(e.message || 'File already exists or invalid name.')}`,
      type: 'error'
    });
  }
}

async function deleteFileItem(path) {
  const baseName = path.split('/').pop();
  const ok = await uiConfirm({
    title: 'Delete Item',
    message: `Are you sure you want to permanently delete <b>${esc(baseName)}</b>? This action cannot be undone.`,
    confirmText: 'Delete',
    cancelText: 'Cancel',
    isDanger: true
  });
  if (!ok) return;

  const srv = curServer();
  if (!srv) return;
  try {
    await api(srvApiPath(srv, `/files?path=${encodeURIComponent(path)}`), { method: 'DELETE' });
    toast(`Deleted ${baseName}`, 'ok');
    loadFileList(S.currentDir);
  } catch (e) {
    toast(`Delete failed: ${e.message}`, 'bad');
    uiAlert({
      title: 'Delete Failed',
      message: `Could not delete <b>${esc(baseName)}</b>: ${esc(e.message || 'Permission denied or file in use.')}`,
      type: 'error'
    });
  }
}

async function promptNewFile() {
  const name = await uiPrompt({
    title: 'Create New File',
    message: `Enter file name in <code>/${esc(S.currentDir || '')}</code>:`,
    placeholder: 'e.g. server.properties, rules.txt',
    confirmText: 'Create File'
  });
  if (!name) return;
  const srv = curServer();
  if (!srv) return;
  const target = S.currentDir ? `${S.currentDir}/${name}` : name;
  try {
    await api(srvApiPath(srv, '/files/create'), {
      method: 'POST',
      body: JSON.stringify({ path: target, is_dir: false })
    });
    toast(`Created file ${name}`, 'ok');
    loadFileList(S.currentDir);
    openFile(name);
  } catch (e) {
    toast(`Create file failed: ${e.message}`, 'bad');
    uiAlert({
      title: 'Failed to Create File',
      message: `Could not create file <b>${esc(name)}</b>: ${esc(e.message || 'File may already exist or filename is invalid.')}`,
      type: 'error'
    });
  }
}

async function promptNewFolder() {
  const name = await uiPrompt({
    title: 'Create New Folder',
    message: `Enter directory name in <code>/${esc(S.currentDir || '')}</code>:`,
    placeholder: 'e.g. plugins, config, backups',
    confirmText: 'Create Folder'
  });
  if (!name) return;
  const srv = curServer();
  if (!srv) return;
  const target = S.currentDir ? `${S.currentDir}/${name}` : name;
  try {
    await api(srvApiPath(srv, '/files/create'), {
      method: 'POST',
      body: JSON.stringify({ path: target, is_dir: true })
    });
    toast(`Created folder ${name}`, 'ok');
    loadFileList(S.currentDir);
  } catch (e) {
    toast(`Create folder failed: ${e.message}`, 'bad');
    uiAlert({
      title: 'Failed to Create Folder',
      message: `Could not create directory <b>${esc(name)}</b>: ${esc(e.message || 'Folder already exists or invalid path.')}`,
      type: 'error'
    });
  }
}

// --- File Selection & Bulk Actions ---
function renderFileBulkBarHtml() {
  const count = S.selectedFiles ? S.selectedFiles.size : 0;
  if (count <= 0) return '';
  const deleting = Boolean(QOL.bulkDelete && sameQolServer(QOL.bulkDelete.server, curServer()));
  const disabled = deleting ? 'disabled' : '';
  return `<div class="bulk-action-bar">
    <span style="font-weight: 600; font-size: 13.5px; color: var(--tx);">${count} item${count > 1 ? 's' : ''} selected</span>
    <div class="row" style="gap: 8px;">
      <button class="btn sm" ${disabled} onclick="openMoveFilesModal()">${ic('move', 14)} Move</button>
      <button class="btn sm" ${disabled} onclick="openCompressFilesModal()">${ic('archive', 14)} Archive</button>
      ${Array.from(S.selectedFiles).some(p => isArchiveFile(p)) ? `<button class="btn sm" ${disabled} onclick="unzipSelectedArchive()">${ic('package', 14)} Extract</button>` : ''}
      <button class="btn sm dg" ${disabled} aria-busy="${deleting}" onclick="deleteSelectedFiles()">${ic('trash', 14)} ${deleting ? 'Deleting…' : 'Delete Selected'}</button>
      <button class="btn sm" ${disabled} onclick="deselectAllFiles()">Deselect</button>
    </div>
  </div>`;
}

function syncSelectAllCheckbox() {
  const chk = $('#chk-all-files');
  if (!chk) return;
  const list = (Array.isArray(S.sortedFileList) && S.sortedFileList.length) ? S.sortedFileList : (S.fileList || []);
  if (!list.length) {
    chk.checked = false;
    chk.indeterminate = false;
    return;
  }
  const selectedCount = list.filter(i => S.selectedFiles.has(i.path || i.name)).length;
  chk.checked = selectedCount === list.length;
  chk.indeterminate = selectedCount > 0 && selectedCount < list.length;
}

function syncFileSelectionUI() {
  const vm = getViewMode('files', 'list');
  $$('.file-row[data-path]').forEach(row => {
    const p = row.getAttribute('data-path') || '';
    const isSel = Boolean(S.selectedFiles && S.selectedFiles.has(p));
    if (vm === 'list') {
      row.classList.toggle('row-selected', isSel);
      row.classList.remove('selected');
    } else {
      row.classList.toggle('selected', isSel);
      row.classList.remove('row-selected');
    }
    const cb = row.querySelector('.file-chk');
    if (cb) cb.checked = isSel;
  });
  syncSelectAllCheckbox();
  const root = $('#file-bulk-bar-root');
  if (root) {
    root.innerHTML = renderFileBulkBarHtml();
    applyPermissionControls(root);
  }
}

function toggleFileSelect(path, checked) {
  if (checked) {
    S.selectedFiles.add(path);
  } else {
    S.selectedFiles.delete(path);
  }
  syncFileSelectionUI();
}

function toggleSelectAllFiles(checked) {
  const list = (Array.isArray(S.sortedFileList) && S.sortedFileList.length) ? S.sortedFileList : (S.fileList || []);
  if (checked) {
    list.forEach(item => {
      S.selectedFiles.add(item.path || item.name);
    });
  } else {
    S.selectedFiles.clear();
  }
  syncFileSelectionUI();
}

function deselectAllFiles() {
  S.selectedFiles.clear();
  syncFileSelectionUI();
}

async function deleteSelectedFiles() {
  const srv = curServer();
  const selection = S.selectedFiles;
  const paths = Array.from(selection || []);
  const directory = S.currentDir;
  if (!srv || !paths.length || QOL.bulkDelete) return;
  QOL.bulkDelete = { server: srv };
  let deleted = 0;
  const failures = [];
  try {
    const ok = await uiConfirm({
      title: 'Delete Selected Items',
      message: `Permanently delete <b>${paths.length}</b> selected item${paths.length > 1 ? 's' : ''} from <b>${esc(srv.name)}</b>? This action cannot be undone.`,
      confirmText: `Delete ${paths.length} Item${paths.length > 1 ? 's' : ''}`,
      cancelText: 'Cancel', isDanger: true
    });
    if (!ok || !sameQolServer(srv, curServer())) return;
    syncFileSelectionUI();
    toast(`Deleting ${paths.length} items…`);
    // A bulk request costs one rate-limit token, rather than one per file.
    // Pace unusually large selections so chunks don't create another burst.
    for (let offset = 0; offset < paths.length; offset += 1000) {
      const chunk = paths.slice(offset, offset + 1000);
      try {
        const result = await api(srvApiPath(srv, '/files/delete-bulk'), {
          method: 'POST', body: JSON.stringify({ paths: chunk }), suppressToast: true
        });
        if (!Array.isArray(result.deleted) || !Array.isArray(result.failed)) throw new Error('Invalid bulk-delete response. Update the panel and worker node together.');
        const completed = new Set(result.deleted);
        for (const path of chunk) {
          if (completed.has(path)) { selection.delete(path); deleted++; }
          else failures.push(result.failed.find(item => item.path === path) || { path, error: 'Deletion was not confirmed by the server.' });
        }
      } catch (err) {
        failures.push({ path: '', error: err.message || 'Bulk deletion failed.' });
        break; // Retain unprocessed selections; never retry destructive requests blindly.
      }
      if (offset + 1000 < paths.length) await new Promise(resolve => setTimeout(resolve, 1100));
    }
    const remaining = paths.length - deleted;
    if (remaining) {
      toast(`Deleted ${deleted} of ${paths.length} items. Remaining items are still selected.`, 'bad');
      if (sameQolServer(srv, curServer())) await uiAlert({
        title: 'Bulk Delete Incomplete', type: 'error',
        message: `<p>${remaining} item(s) remain selected for retry.</p><ul>${failures.slice(0, 8).map(item => `<li>${item.path ? `<b>${esc(item.path)}</b>: ` : ''}${esc(item.error)}</li>`).join('')}</ul>`
      });
    } else toast(`Deleted ${deleted} items`, 'ok');
    if (sameQolServer(srv, curServer()) && S.currentDir === directory) await loadFileList(directory);
  } finally {
    QOL.bulkDelete = null;
    if (sameQolServer(srv, curServer())) syncFileSelectionUI();
  }
}

function copyFilePath(path) {
  if (!hasServerPermission('files.view')) return;
  if (navigator.clipboard) {
    navigator.clipboard.writeText(path).then(() => {
      toast(`Copied "${path}" to clipboard!`, 'ok');
    }).catch(() => {
      uiPrompt({
        title: 'Copy File Path',
        message: 'Copy the path below:',
        defaultValue: path,
        confirmText: 'Done'
      });
    });
  } else {
    uiPrompt({
      title: 'Copy File Path',
      message: 'Copy the path below:',
      defaultValue: path,
      confirmText: 'Done'
    });
  }
}

// --- Floating Context Menu ---
function handleFileContextMenu(e, idx) {
  if (e) {
    e.preventDefault();
    e.stopPropagation();
  }
  const list = (Array.isArray(S.sortedFileList) && S.sortedFileList.length) ? S.sortedFileList : (S.fileList || []);
  const item = list[idx];
  if (!item) return;

  const menu = $('#floating-ctx');
  if (!menu) return;

  const isDir = item.is_dir;
  const path = item.path || item.name;
  const name = item.name;
  const isArchive = isArchiveFile(name);

  menu.innerHTML = `
    ${!isDir ? `
      <button data-name="${esc(name)}" onclick="openFile(this.dataset.name); closeFloatingCtx();">
        ${ic('eye', 14)} Open / Preview
      </button>
      <button data-path="${esc(path)}" onclick="downloadFileItem(this.dataset.path); closeFloatingCtx();">
        ${ic('download', 14)} Download
      </button>
      ${isArchive ? `
        <button data-path="${esc(path)}" onclick="unzipFileItem(this.dataset.path); closeFloatingCtx();">
          ${ic('package', 14)} Extract / Unzip
        </button>
      ` : ''}
    ` : `
      <button data-name="${esc(name)}" onclick="browseFolder(this.dataset.name); closeFloatingCtx();">
        ${ic('folder', 14)} Open Folder
      </button>
    `}
    <button data-path="${esc(path)}" onclick="openMoveFilesModal([this.dataset.path]); closeFloatingCtx();">
      ${ic('move', 14)} Move
    </button>
    <button data-path="${esc(path)}" onclick="openCompressFilesModal([this.dataset.path]); closeFloatingCtx();">
      ${ic('archive', 14)} Archive / Zip
    </button>
    <button data-path="${esc(path)}" onclick="promptRenameFile(this.dataset.path); closeFloatingCtx();">
      ${ic('settings', 14)} Rename
    </button>
    <button data-path="${esc(path)}" onclick="copyFilePath(this.dataset.path); closeFloatingCtx();">
      ${ic('copy', 14)} Copy Path
    </button>
    <div class="floating-ctx-divider"></div>
    <button class="dg" data-path="${esc(path)}" onclick="deleteFileItem(this.dataset.path); closeFloatingCtx();">
      ${ic('trash', 14)} Delete
    </button>
  `;
  applyPermissionControls(menu);

  const clientX = e?.clientX ?? (e?.touches?.[0]?.clientX ?? 100);
  const clientY = e?.clientY ?? (e?.touches?.[0]?.clientY ?? 150);
  const menuWidth = 190;
  const x = Math.max(10, Math.min(clientX, window.innerWidth - menuWidth - 10));
  const y = Math.max(10, Math.min(clientY, window.innerHeight - 315));

  menu.style.left = `${x}px`;
  menu.style.top = `${y}px`;
  menu.style.display = 'block';
}

function closeFloatingCtx() {
  const menu = $('#floating-ctx');
  if (menu) menu.style.display = 'none';
}

// Global click dismisses floating context menu
document.addEventListener('click', () => closeFloatingCtx());

// --- Upload with Live Progress & Drag-and-Drop ---
function uploadFileDialog() {
  const input = document.createElement('input');
  input.type = 'file';
  input.multiple = true;
  input.onchange = () => {
    if (input.files && input.files.length > 0) {
      uploadFilesWithProgress(Array.from(input.files));
    }
  };
  input.click();
}

async function uploadFilesWithProgress(files) { return enqueueUploads(files); }

function initDragAndDrop() {
  let dragCounter = 0;
  const overlay = $('#dropzone-overlay');

  window.addEventListener('dragenter', (e) => {
    e.preventDefault();
    if (S.view !== 'fi') return;
    dragCounter++;
    if (overlay) overlay.style.display = 'flex';
  });

  window.addEventListener('dragover', (e) => {
    e.preventDefault();
    if (S.view !== 'fi') return;
  });

  window.addEventListener('dragleave', (e) => {
    e.preventDefault();
    if (S.view !== 'fi') return;
    dragCounter--;
    if (dragCounter <= 0) {
      dragCounter = 0;
      if (overlay) overlay.style.display = 'none';
    }
  });

  window.addEventListener('drop', (e) => {
    e.preventDefault();
    dragCounter = 0;
    if (overlay) overlay.style.display = 'none';
    if (S.view !== 'fi') return;

    if (e.dataTransfer && e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      uploadFilesWithProgress(Array.from(e.dataTransfer.files));
    }
  });
}

// --- Plugins & Modrinth Store ---
async function loadPluginsList(skipRender = false) {
  const srv = curServer();
  if (!srv) return;
  if (isSteelEngine(srv.type)) return;
  if (!S.loading) S.loading = {};
  S.loading.plugins = true;
  if (!skipRender && S.view === 'plug' && (S.plugTab || 'browse') === 'installed') R();
  try {
    const data = await api(srvApiPath(srv, '/plugins'));
    if (!sameQolServer(srv, curServer())) return;
    const list = Array.isArray(data) ? data : (data?.plugins || []);
    S.plugins = list.map(p => {
      if (typeof p === 'string') return [p, 'installed', true, ''];
      const fname = p.filename || p.name || 'plugin.jar';
      const meta = p.size ? formatFileSize(p.size) : (p.version || 'installed');
      return [fname, meta, p.enabled !== false, p.modified || '', false];
    });
  } catch (e) {} finally {
    S.loading.plugins = false;
    if (!skipRender) {
      if (S.view === 'plug' && (S.plugTab || 'browse') === 'browse') {
        const cnt = $('#plug-installed-count');
        if (cnt) cnt.textContent = String(S.plugins.length);
      } else {
        R();
      }
    }
  }
}

async function deletePlugin(name) {
  const srv = curServer();
  if (!srv || !name) return;
  const isMod = isModdedEngine(srv.type || S.core);
  const noun = isMod ? 'Mod' : 'Plugin';
  const nounLow = isMod ? 'mod' : 'plugin';
  const ok = await uiConfirm({
    title: `Remove ${noun}`,
    message: `Are you sure you want to delete <b>${esc(name)}</b> from <b>${esc(srv.name)}</b>? Restart the server afterward to unload the ${nounLow}.`,
    confirmText: `Remove ${noun}`,
    cancelText: 'Cancel',
    isDanger: true,
    icon: 'trash'
  });
  if (!ok) return;

  try {
    await api(srvApiPath(srv, `/plugins?filename=${encodeURIComponent(name)}`), { method: 'DELETE' });
    toast(`Removed ${nounLow} ${name}`);
    loadPluginsList();
  } catch (e) {}
}

async function togglePlugin(idx, enabled) {
  const srv = curServer();
  const plugin = S.plugins[idx];
  if (!srv || !plugin || plugin[4]) return;
  plugin[4] = true;
  R();
  try {
    const result = await api(srvApiPath(srv, '/plugins/toggle'), { method: 'POST', body: JSON.stringify({ filename: plugin[0], enabled }) });
    plugin[0] = result.filename;
    plugin[2] = result.enabled;
    toast(`${result.filename} ${result.enabled ? 'enabled' : 'disabled'}. Restart server to apply.`, 'ok');
  } catch (_) { /* api reports the failure; retain the previous server state. */ }
  finally {
    plugin[4] = false;
    if (sameQolServer(srv, curServer())) R();
  }
}

let searchTimeout = null;
function searchModrinthDebounced() {
  clearTimeout(searchTimeout);
  searchTimeout = setTimeout(() => searchModrinth(), 350);
}

function renderModrinthHitsHtml(hits) {
  const list = Array.isArray(hits) ? hits : [];
  const storeVm = getViewMode('plug_store', 'list');
  const isMod = isModdedEngine(curServer()?.type || S.core);
  if (!list.length) {
    return `<div class="mu" style="padding:24px; text-align:center; grid-column:1/-1;">No matching ${isMod ? 'mods' : 'plugins'} found on Modrinth.</div>`;
  }
  return list.map(hit => {
    const titleLow = String(hit.title || '').toLowerCase();
    const slugLow = String(hit.slug || '').toLowerCase();
    const isInstalled = S.plugins.some(p => {
      const fname = String(p[0] || '').toLowerCase();
      return (titleLow && fname.includes(titleLow.replace(/\s+/g, ''))) || (slugLow && fname.includes(slugLow));
    });
    const icon = hit.icon_url
      ? `<img src="${esc(hit.icon_url)}" class="plugin-icon" onerror="this.src='/favicon.svg'">`
      : `<div class="plugin-icon" style="display:grid; place-items:center; color:var(--ac); font-size:20px;">🧩</div>`;
    const slugVal = hit.slug || hit.project_id || hit.id || '';
    const srv = curServer();
    const isInstalling = Boolean(S.installingPlugins && S.installingPlugins.has(`${srv?.node_id || 'local'}:${srv?.id}:${slugVal}`));
    const btnPad = storeVm === 'icons' ? 'min-width: 88px; font-size: 12px; padding: 5px 10px;' : 'min-width: 112px;';
    const actionHtml = isInstalled
      ? '<span class="badge online"><i class="dot"></i>Installed</span>'
      : isInstalling
        ? `<button class="btn ${storeVm === 'icons' ? 'sm ' : ''}pr is-loading" data-slug="${esc(slugVal)}" data-title="${esc(hit.title)}" onclick="installPlugin(this.dataset.slug, this.dataset.title, this)" disabled aria-busy="true" style="${btnPad} justify-content: center;"><span class="np-spinner-sm"></span> Installing...</button>`
        : `<button class="btn ${storeVm === 'icons' ? 'sm ' : ''}pr" data-slug="${esc(slugVal)}" data-title="${esc(hit.title)}" onclick="installPlugin(this.dataset.slug, this.dataset.title, this)" style="${btnPad} justify-content: center;">${ic('download', 14)} Install</button>`;

    if (storeVm === 'icons') {
      return `
      <div class="vm-icon-card">
        <div class="vm-plug-icon-center">
          ${icon}
          <b class="plugin-title" style="font-size: 13.5px;" title="${esc(hit.title)}">${esc(hit.title)}</b>
          <div class="mu" style="font-size: 11.5px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 100%;">by ${esc(hit.author || 'Author')}</div>
        </div>
        <div class="vm-item-foot">
          <span class="mu" style="font-size: 11px;">&darr; ${formatNum(hit.downloads)}</span>
          ${actionHtml}
        </div>
      </div>`;
    }

    if (storeVm === 'tiles') {
      return `
      <div class="vm-tile-card">
        <div class="vm-item-head" style="align-items: flex-start;">
          <div style="display: flex; align-items: center; gap: 12px; min-width: 0; flex: 1;">
            ${icon}
            <div style="min-width: 0;">
              <b class="vm-item-title" style="display: block;" title="${esc(hit.title)}">${esc(hit.title)}</b>
              <div class="mu" style="font-size: 12px; margin-top: 2px;">by ${esc(hit.author || 'Author')}</div>
            </div>
          </div>
        </div>
        <div class="plugin-desc" style="display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; white-space: normal;">${esc(hit.description || '')}</div>
        <div class="vm-item-foot">
          <div class="row" style="gap: 8px; font-size: 11.5px;">
            <span class="mu">&darr; ${formatNum(hit.downloads)}</span>
            <span class="chip">Modrinth</span>
          </div>
          ${actionHtml}
        </div>
      </div>`;
    }

    return `
    <div class="plugin-card">
      <div style="display: flex; align-items: center; gap: 16px; min-width: 0;">
        ${icon}
        <div style="min-width: 0;">
          <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 2px; flex-wrap: wrap;">
            <b class="plugin-title">${esc(hit.title)}</b>
            <span class="mu" style="font-size: 12.5px;">by ${esc(hit.author || 'Author')}</span>
          </div>
          <div class="plugin-desc">${esc(hit.description || '')}</div>
          <div class="row" style="gap: 10px; font-size: 11.5px; margin-top: 5px; flex-wrap: wrap;">
            <span class="mu">&darr; ${formatNum(hit.downloads)}</span>
            <span class="mu">&hearts; ${formatNum(hit.follows || 0)} follows</span>
            <span class="chip">Modrinth</span>
          </div>
        </div>
      </div>
      <div style="flex-shrink: 0;">
        ${actionHtml}
      </div>
    </div>`;
  }).join('');
}

async function searchModrinth(skipDomWrite = false) {
  const srv = curServer();
  if (isSteelEngine(srv?.type)) return;
  const isMod = isModdedEngine(srv?.type || S.core);
  const engSlug = engineMeta(srv?.type || S.core).key;
  const grid = $('#store-grid');
  if (!skipDomWrite && grid) {
    grid.innerHTML = loaderHtml(`Searching Modrinth ${isMod ? 'mods' : 'plugins'}...`);
  }
  const q = S.q || '';
  const loader = S.plugLoader || (isMod ? (engSlug === 'leaf' ? '' : engSlug) : '');
  const ver = S.plugVer || '';
  const sort = S.plugSort || 'relevance';
  try {
    const params = new URLSearchParams({
      query: q,
      type: isMod ? 'mod' : 'plugin',
      loader,
      version: ver,
      sort,
      limit: '24'
    });
    const res = await api(`/api/plugins/search?${params.toString()}`);
    S.lastModrinthHits = res.hits || [];
    if (!skipDomWrite && grid) {
      grid.innerHTML = renderModrinthHitsHtml(S.lastModrinthHits);
      applyPermissionControls(grid);
    }
  } catch (e) {
    S.lastModrinthHits = [];
    if (!skipDomWrite && grid) {
      grid.innerHTML = '<div class="mu" style="padding:24px; text-align:center; grid-column:1/-1;">Error reaching Modrinth API.</div>';
    }
  }
}

function setPlugSkipConfirm(val) {
  S.skipPlugConfirm = Boolean(val);
  try {
    if (S.skipPlugConfirm) localStorage.setItem('neopanel_skip_plug_confirm', 'true');
    else localStorage.removeItem('neopanel_skip_plug_confirm');
  } catch (e) {}
  const toggleEl = $('#plug-confirm-toggle');
  if (toggleEl) toggleEl.checked = S.skipPlugConfirm;
}

async function installPlugin(projectSlug, title, btnEl) {
  const srv = curServer();
  if (!srv || !projectSlug) return;
  if (isSteelEngine(srv.type)) { toast('Java plugins and mods are incompatible with Steel.', 'bad'); return; }
  if (!S.installingPlugins) S.installingPlugins = new Set();
  const installKey = `${srv.node_id || 'local'}:${srv.id}:${projectSlug}`;
  if (S.installingPlugins.has(installKey)) return;
  S.installingPlugins.add(installKey);
  if (btnEl) setBtnLoading(btnEl, true, 'Checking…');
  try {
    const result = await api(srvApiPath(srv, `/plugins/versions?slug=${encodeURIComponent(projectSlug)}`));
    const versions = result.versions || [];
    if (!sameQolServer(srv, curServer())) return;
    if (!versions.length) {
      await uiAlert({ title: 'No compatible release', message: `No release of <b>${esc(title || projectSlug)}</b> explicitly supports Minecraft <b>${esc(result.minecraft_version)}</b> and <b>${esc(result.loader)}</b>. Nothing was installed.` });
      return;
    }
    let selected = versions[0];
    if (!S.skipPlugConfirm) {
      const confirmation = uiConfirm({
        title: 'Install compatible release',
        message: `<p>Install <b>${esc(title || projectSlug)}</b> on <b>${esc(srv.name)}</b>?</p><p class="mu">Minecraft ${esc(result.minecraft_version)} · ${esc(result.loader)}. Restart required after installation.</p><label class="f"><span>Compatible release</span><select id="plugin-release">${versions.map(v => `<option value="${esc(v.id)}">${esc(v.version_number)} — ${esc(v.name)}</option>`).join('')}</select></label>`,
        confirmText: 'Download & Install', cancelText: 'Cancel'
      });
      $('#plugin-release').onchange = event => { selected = versions.find(v => v.id === event.target.value) || versions[0]; };
      if (!await confirmation) return;
    }
    if (!sameQolServer(srv, curServer())) return;
    if (btnEl) setBtnLoading(btnEl, true, 'Installing…');
    await api(srvApiPath(srv, '/plugins/install'), { method: 'POST', body: JSON.stringify({ slug: projectSlug, version_id: selected.id }) });
    toast(`Installed ${title || projectSlug} ${selected.version_number}. Restart server to apply.`, 'ok');
    if (sameQolServer(srv, curServer())) await loadPluginsList();
  } catch (_) { /* api displays actionable errors. */ }
  finally {
    S.installingPlugins.delete(installKey);
    if (btnEl && document.body.contains(btnEl)) setBtnLoading(btnEl, false);
    if (sameQolServer(srv, curServer()) && S.view === 'plug' && $('#store-grid') && Array.isArray(S.lastModrinthHits)) {
      $('#store-grid').innerHTML = renderModrinthHitsHtml(S.lastModrinthHits);
      applyPermissionControls($('#store-grid'));
    }
  }
}

// --- Backups ---
async function loadBackupsList(skipRender = false) {
  const srv = curServer();
  if (!srv) return;
  if (!S.loading) S.loading = {};
  if (!S.bk || S.bk.length === 0) {
    S.loading.backups = true;
    if (!skipRender && S.view === 'bk') R();
  }
  try {
    const list = await api(srvApiPath(srv, '/backups'));
    S.bk = list || [];
  } catch (e) {} finally {
    S.loading.backups = false;
    if (!skipRender && S.view === 'bk') R();
  }
}

async function createBackup() {
  const srv = curServer();
  if (!srv) return;
  const ok = await uiConfirm({
    title: 'Create backup',
    message: `Create a new compressed <code>.tar.gz</code> backup snapshot of <b>${esc(srv.name)}</b> (including worlds, plugins, and configs)?`,
    confirmText: 'Create Backup',
    cancelText: 'Cancel',
    isDanger: false,
    icon: 'backup'
  });
  if (!ok) return;

  toast('Creating full server archive (.tar.gz)...');
  try {
    await api(srvApiPath(srv, '/backups'), {
      method: 'POST',
      body: JSON.stringify({ name: '' })
    });
    toast('Backup created successfully!');
    loadBackupsList();
  } catch (e) {}
}

function downloadBackup(name) {
  const srv = curServer();
  if (!srv) return;
  const baseUrl = srvApiPath(srv, `/backups/download?name=${encodeURIComponent(name)}`);
  const sep = baseUrl.includes('?') ? '&' : '?';
  window.open(`${baseUrl}${sep}token=${encodeURIComponent(S.token)}`, '_blank');
}

async function deleteBackup(name) {
  const srv = curServer();
  if (!srv) return;
  const ok = await uiConfirm({
    title: 'Delete Backup Archive',
    message: `Are you sure you want to permanently delete backup <b>${esc(name)}</b>? This action cannot be undone.`,
    confirmText: 'Delete Backup',
    cancelText: 'Cancel',
    isDanger: true,
    icon: 'trash'
  });
  if (!ok) return;

  try {
    await api(srvApiPath(srv, `/backups?name=${encodeURIComponent(name)}`), { method: 'DELETE' });
    toast(`Backup ${name} deleted`);
    loadBackupsList();
  } catch (e) {}
}

// --- Software & Properties ---
function askInstallSoftwareModal(srv, software, version) {
  return new Promise(resolve => {
    if (_uiModalResolver) {
      _uiModalResolver(null);
    }
    _uiModalResolver = (val) => resolve(val);

    const recJava = getRecommendedJavaForMcVersion(version, software);
    const curJava = resolveJavaSelectValue(srv.java_path || '', srv.version || '', srv.type || '');
    const needsJavaSwitch = !isSteelEngine(software) && curJava !== recJava;
    const javaInfo = getPremadeJavaStatusMap(srv?.node_id || 'local').get(recJava);
    const javaMissingNote = !isSteelEngine(software) && javaInfo && !javaInfo.installed
      ? `<span class="mu" style="font-size: 11.5px; display: block; margin-top: 4px; color: #fbbf24;">Java ${esc(recJava)} is not installed yet. Install it from Settings or Admin before starting the server.</span>`
      : '';

    const container = $('#modal-container') || document.body;
    container.innerHTML = `
      <div class="pal" onclick="if(event.target===this) closeUiModal(null)">
        <div class="pbox ui-modal-box" style="max-width: 480px;">
          <div class="row sb" style="margin-bottom: 14px; align-items: center;">
            <h3 style="font-size: 17px; font-weight: 600; margin: 0; display: flex; align-items: center; gap: 8px; color: var(--tx);">
              ${ic('download', 18)} <span>Install ${esc(software)} ${esc(version)}</span>
            </h3>
            <button class="btn sm" onclick="closeUiModal(null)" aria-label="Close" style="padding: 4px 8px; font-size: 15px; opacity: 0.8;">&times;</button>
          </div>
          <div class="ui-modal-body" style="line-height: 1.55;">
            Download and install <b>${esc(software)} ${esc(version)}</b> onto <b>${esc(srv.name)}</b>? This will replace your current server executable and restart the instance.
            ${isSteelEngine(software) ? '<p>Steel is an experimental Minecraft Java Edition server written in Rust; it does not require Java. Edit <code>config/config.toml</code> in File Manager; worlds use <code>saves/</code>.</p>' : ''}
            ${incompatibleEngineChange(srv.type, software) ? '<p class="banner warn"><b>Incompatible engine change:</b> worlds/data, plugins/mods and configuration may be incompatible and are not converted. Java plugins/mods cannot run on Steel. Back up first; use Reinstall for a clean installation.</p>' : ''}
            <div style="margin-top: 14px; padding: 12px 14px; background: rgba(56, 122, 255, 0.08); border: 1px solid rgba(56, 122, 255, 0.25); border-radius: 10px;">
              <label style="display: flex; align-items: flex-start; gap: 10px; cursor: pointer; margin: 0;">
                <input type="checkbox" id="confirm-bkf" checked style="margin-top: 2px;">
                <div>
                  <b style="color: var(--tx); display: block; font-size: 13.5px;">Back up the world first (Recommended)</b>
                  <span class="mu" style="font-size: 12px; display: block; margin-top: 2px;">Creates a full <code>before-${esc(version)}</code> backup archive before replacing the server executable.</span>
                </div>
              </label>
            </div>
            ${needsJavaSwitch ? `
            <div style="margin-top: 10px; padding: 12px 14px; background: rgba(16, 185, 129, 0.08); border: 1px solid rgba(16, 185, 129, 0.28); border-radius: 10px;">
              <label style="display: flex; align-items: flex-start; gap: 10px; cursor: pointer; margin: 0;">
                <input type="checkbox" id="confirm-sync-java" checked style="margin-top: 2px;">
                <div>
                  <b style="color: var(--tx); display: block; font-size: 13.5px;">Switch Java Runtime to Java ${esc(recJava)} (Recommended)</b>
                  <span class="mu" style="font-size: 12px; display: block; margin-top: 2px;">${esc(software)} ${esc(version)} expects <b>Java ${esc(recJava)}</b> (currently using <code>${esc(curJava === 'custom' ? (srv.java_path || 'custom') : 'Java ' + curJava)}</code>).</span>
                  ${javaMissingNote}
                </div>
              </label>
            </div>` : (javaMissingNote ? `<div style="margin-top: 10px;">${javaMissingNote}</div>` : '')}
          </div>
          <div class="row" style="justify-content: flex-end; gap: 10px; margin-top: 20px;">
            <button type="button" class="btn" onclick="closeUiModal(null)">Cancel</button>
            <button type="button" id="ui-modal-confirm-btn" class="btn pr" onclick="closeUiModal({ confirmed: true, backup: Boolean(document.getElementById('confirm-bkf')?.checked), syncJava: Boolean(document.getElementById('confirm-sync-java')?.checked), recJava: '${esc(recJava)}' })">
              ${ic('download', 14)} Download &amp; Install
            </button>
          </div>
        </div>
      </div>
    `;

    setTimeout(() => $('#ui-modal-confirm-btn')?.focus(), 40);
  });
}

async function installSoftware(software, version, btnEl) {
  if (isSteelEngine(software)) software = 'steel';
  const srv = curServer();
  if (!srv) return;
  if (S.installingSoftware) {
    toast('Another software installation is already in progress.');
    return;
  }
  const choice = await askInstallSoftwareModal(srv, software, version);
  if (!choice || !choice.confirmed) return;

  S.installingSoftware = `${software}:${version}`;
  const findBtn = () => {
    if (btnEl && document.body.contains(btnEl)) return btnEl;
    if (typeof CSS !== 'undefined' && CSS.escape) {
      return document.querySelector(`#vres button[data-soft="${CSS.escape(software)}"][data-ver="${CSS.escape(version)}"]`);
    }
    return null;
  };

  let curBtn = findBtn();
  try {
    if (choice.backup) {
      if (curBtn) setBtnLoading(curBtn, true, 'Backing up...');
      toast('Creating world & server backup prior to upgrade...');
      await api(srvApiPath(srv, '/backups'), { method: 'POST', body: JSON.stringify({ name: `before-${version}` }) }).catch(() => {});
    }
    curBtn = findBtn();
    if (curBtn) setBtnLoading(curBtn, true, 'Installing...');
    toast(`Downloading ${software} ${version} server ${isSteelEngine(software) ? 'executable' : 'jar'}...`);
    const softRes = await api(srvApiPath(srv, '/software'), {
      method: 'POST',
      body: JSON.stringify({ software, version })
    });
    srv.type = softRes?.software || String(software).toLowerCase();
    srv.version = softRes?.version || version;
    if (softRes?.jar_file) srv.jar_file = softRes.jar_file;

    if (!isSteelEngine(software) && choice.syncJava && choice.recJava) {
      await api(srvApiPath(srv, ''), {
        method: 'PUT',
        body: JSON.stringify({
          name: srv.name,
          description: srv.description,
          type: srv.type,
          version: srv.version,
          jar_file: srv.jar_file,
          port: srv.port,
          memory_min: srv.memory_min,
          memory_max: srv.memory_max,
          java_path: choice.recJava,
          java_args: srv.java_args,
          auto_restart: srv.auto_restart
        })
      });
      srv.java_path = choice.recJava;
    }

    S.core = isSteelEngine(software) ? 'Steel' : software;
    S.ver = srv.version;
    toast(`Successfully installed ${software} ${version}! Restarting...`);
    power('restart', true);
  } catch (e) {
    toast(`Software installation failed: ${e.message}`, 'bad');
    uiAlert({
      title: 'Installation Failed',
      message: `Failed to install <b>${esc(software)} ${esc(version)}</b>:<br><br>${esc(e.message || 'Download failed or version unavailable.')}`,
      type: 'error'
    });
  } finally {
    S.installingSoftware = null;
    if (S.view === 'ver') {
      renderVersionRows();
      R();
    } else {
      curBtn = findBtn();
      if (curBtn) setBtnLoading(curBtn, false);
    }
  }
}

function syncPropsFromServer(srv, force = false) {
  if (!srv) return;
  if (force || S.props._srvId !== srv.id) {
    S.props.motd = srv.motd || srv.name || 'A Minecraft Server';
    S.props._srvId = srv.id;
  } else if (srv.motd && (!S.props.motd || S.props.motd === 'A Minecraft Server' || S.props.motd === srv.name)) {
    S.props.motd = srv.motd;
  }
}

function onMotdInput(val) {
  const srv = curServer();
  S.props.motd = val;
  if (srv) {
    S.props._srvId = srv.id;
    srv.motd = val;
  }
  const prev = $('#motd-preview');
  if (prev) prev.innerHTML = renderMinecraftTextHtml(val, { allowNewlines: true });
  const heroTitle = $('#srv-hero-title');
  if (heroTitle) {
    const trimmed = String(val || '').trim();
    if (!trimmed) {
      heroTitle.style.display = 'none';
      heroTitle.innerHTML = '';
    } else {
      heroTitle.style.display = 'block';
      heroTitle.innerHTML = renderMinecraftTextHtml(trimmed, { allowNewlines: true });
    }
  }
  scheduleSettingsAutoSave(500);
}

function shouldSkipSettingsBgRender() {
  if (S.view !== 'set') return false;
  if (_settingsAutoTimer || _settingsSaveInFlight) return true;
  const act = document.activeElement;
  const main = $('#main');
  if (act && main && main.contains && main.contains(act) && ['INPUT', 'TEXTAREA', 'SELECT'].includes(act.tagName)) {
    return true;
  }
  return false;
}

async function loadProperties(skipRender = false) {
  const srv = curServer();
  if (!srv) return;
  if (isSteelEngine(srv.type)) { seedSettingsSavedSnapshot(srv); return; }
  try {
    const props = await api(srvApiPath(srv, '/properties'));
    if (shouldSkipSettingsBgRender()) return;
    S.props._srvId = srv.id;
    S.props.motd = props['motd'] || srv.motd || srv.name || 'A Minecraft Server';
    srv.motd = S.props.motd;
    S.props.max = parseInt(props['max-players'] || '20');
    S.props.diff = props['difficulty'] || S.props.diff;
    S.props.mode = props['gamemode'] || S.props.mode;
    S.props.pvp = props['pvp'] !== 'false';
    S.props.wl = props['white-list'] === 'true';
    S.props.cmd = props['enable-command-block'] === 'true';
    if (!skipRender && !shouldSkipSettingsBgRender()) R();
    seedSettingsSavedSnapshot(srv);
  } catch (e) {}
}

async function persistSettingsPayload(targetSrv, collected, opts = {}) {
  if (!hasServerPermission('settings.edit', targetSrv)) return false;
  const { autoSave = false, skipRender = false, force = false } = opts || {};
  if (!targetSrv || !collected || !collected.ok) return;

  if (_settingsSaveInFlight) {
    _settingsSaveQueued = { targetSrv, collected, opts, fromDom: opts.fromDom !== false };
    return;
  }

  const snaps = buildSettingsSnapshots(targetSrv, collected.data);
  const key = getSettingsSnapshotKey(targetSrv);
  const prev = _settingsLastSaved[key];
  const sendProps = !isSteelEngine(targetSrv.type) && (force || !prev || prev.propsJson !== snaps.propsJson);
  const sendSrv = force || !prev || prev.srvJson !== snaps.srvJson;

  if (!sendProps && !sendSrv) {
    setSettingsSaveStatus('saved', 'All changes saved automatically');
    return;
  }

  _settingsSaveInFlight = true;
  setSettingsSaveStatus('saving', 'Saving changes...');

  const d = collected.data;
  const nodeId = targetSrv.node_id || 'local';

  try {
    if (sendProps) {
      await api(srvApiPath(targetSrv, '/properties'), {
        method: 'POST',
        suppressToast: autoSave,
        body: JSON.stringify(snaps.propsPayload)
      });
    }

    if (sendSrv) {
      await api(srvApiPath(targetSrv, ''), {
        method: 'PUT',
        suppressToast: autoSave,
        body: JSON.stringify(snaps.srvPayload)
      });
    }

    targetSrv.name = d.name;
    targetSrv.description = d.desc;
    targetSrv.version = d.ver;
    targetSrv.jar_file = d.jar;
    targetSrv.java_args = snaps.srvPayload.java_args;
    targetSrv.auto_restart = d.autoRestart;
    targetSrv.port = d.port;
    targetSrv.memory_max = d.ram;
    targetSrv.memory_min = snaps.srvPayload.memory_min;
    targetSrv.motd = d.motd;
    if (isSteelEngine(targetSrv.type) || d.javaPath) targetSrv.java_path = snaps.srvPayload.java_path;

    _settingsLastSaved[key] = {
      propsJson: snaps.propsJson,
      srvJson: snaps.srvJson
    };

    const stillActive = curServer()?.id === targetSrv.id && (curServer()?.node_id || 'local') === nodeId;
    if (stillActive) {
      S.ver = d.ver;
      S.editSrvName = d.name;
      S.editSrvDesc = d.desc;
      S.props._srvId = targetSrv.id;
      S.memMaxMB = d.ram;
      syncHeroFromSavedServer(targetSrv);
    }

    setSettingsSaveStatus('saved', `Saved at ${clock()}`);

    if (!autoSave) {
      toast('Server settings, identity & startup configuration saved!', 'ok');
      if (!skipRender) R();
    }
  } catch (e) {
    setSettingsSaveStatus('error', e.message || 'Auto-save failed');
    toast(e.message || 'Failed to save server settings', 'bad');
    if (!autoSave) {
      uiAlert({
        title: 'Failed to Save Settings',
        message: `Could not save server properties: ${esc(e.message || 'Validation error')}`,
        type: 'error'
      });
    }
  } finally {
    _settingsSaveInFlight = false;
    if (_settingsSaveQueued) {
      const next = _settingsSaveQueued;
      _settingsSaveQueued = null;
      let nextCollected = next.collected;
      if (next.fromDom && curServer()?.id === next.targetSrv.id && S.view === 'set') {
        const fresh = collectSettingsFormPayload(next.targetSrv);
        if (fresh.ok) nextCollected = fresh;
        else {
          setSettingsSaveStatus('warn', fresh.reason);
          return;
        }
      }
      void persistSettingsPayload(next.targetSrv, nextCollected, next.opts);
    }
  }
}

async function saveProperties(opts = {}) {
  clearSettingsAutoSaveTimer();
  const srv = curServer();
  if (!srv) return;
  const collected = collectSettingsFormPayload(srv);
  if (!collected.ok) {
    setSettingsSaveStatus('warn', collected.reason);
    if (!opts?.autoSave) {
      toast(collected.reason || 'Check server settings values', 'bad');
    }
    return;
  }
  return persistSettingsPayload(srv, collected, opts);
}

async function loadSFTPDetails(skipRender = false) {
  const srv = curServer();
  if (!srv || !hasServerPermission('sftp.access')) return;
  const key = accessIdentity();
  S.sftpLoading = true; S.sftpError = '';
  if (!skipRender) R();
  try {
    const data = await api(srvApiPath(srv, '/sftp'), { suppressToast: true });
    if (key === accessIdentity()) S.sftp = data;
  } catch (error) { if (key === accessIdentity()) S.sftpError = error.message; }
  finally { if (key === accessIdentity()) { S.sftpLoading = false; if (!skipRender && S.view === 'sftp') R(); } }
}

// --- Account Management ---
async function loadAccountData(skipRender = false) {
  try {
    const prof = await api('/api/account/profile');
    if (prof) {
      S.acct.un = prof.username || S.acct.un || '';
      S.acct.em = prof.email || S.acct.em || '';
      S.acct.fn = prof.first_name !== undefined ? prof.first_name : (S.acct.fn || '');
      S.acct.ln = prof.last_name !== undefined ? prof.last_name : (S.acct.ln || '');
      S.acct.avatar_url = prof.avatar_url || S.acct.avatar_url || '';
      S.acct.two_factor_enabled = Boolean(prof.two_factor_enabled);

      if (S.user) {
        S.user.username = S.acct.un;
        S.user.email = S.acct.em;
        S.user.first_name = S.acct.fn;
        S.user.last_name = S.acct.ln;
      }

      localStorage.setItem('neopanel_profile', JSON.stringify({
        un: S.acct.un,
        em: S.acct.em,
        fn: S.acct.fn,
        ln: S.acct.ln
      }));
    }
    S.acct.h = await computeGravatarHash(S.acct.em);
    if (S.acct.tab === 1) await loadPasskeys(true);

    if (S.acct.tab === 2) {
      S.acct.ssh = (await api('/api/account/ssh-keys')) || [];
    } else if (S.acct.tab === 3) {
      S.acct.api = (await api('/api/account/api-keys')) || [];
    } else if (S.acct.tab === 4) {
      S.acct.act = (await api(`/api/account/activity?limit=${S.actFilter?.limit || 200}`)) || [];
    }
    if (!skipRender) R();
  } catch (e) {}
}

async function updateAccountProfile() {
  const un = $('#aun')?.value?.trim() || '';
  const fn = $('#afn')?.value?.trim() || '';
  const ln = $('#aln')?.value?.trim() || '';
  const em = $('#aem')?.value?.trim() || '';

  if (!un) {
    toast('Account name / username is required', 'bad');
    return;
  }
  if (!em || !em.includes('@') || !em.includes('.')) {
    toast('A valid email address is required', 'bad');
    return;
  }

  try {
    const res = await api('/api/account/profile', {
      method: 'PUT',
      body: JSON.stringify({ username: un, first_name: fn, last_name: ln, email: em })
    });

    if (res?.token) {
      S.token = res.token;
      localStorage.setItem('neopanel_token', res.token);
    }
    if (S.user) {
      S.user.username = un;
      S.user.first_name = fn;
      S.user.last_name = ln;
      S.user.email = em;
    }
    S.acct.un = un;
    S.acct.fn = fn;
    S.acct.ln = ln;
    S.acct.em = em;
    S.acct.h = await computeGravatarHash(em);

    const profObj = { un, fn, ln, em };
    localStorage.setItem('neopanel_profile', JSON.stringify(profObj));
    if (S.user?.id) {
      localStorage.setItem(`neopanel_names_${S.user.id}`, JSON.stringify(profObj));
    }

    toast('Account profile updated successfully!', 'ok');
    R();
  } catch (err) {
    toast(err.message || 'Failed to update profile', 'bad');
  }
}

async function openEnable2FAWizard() {
  const container = $('#modal-container');
  if (!container) return;

  container.innerHTML = `<div class="pal" onclick="if(event.target===this)closeModal()">
    <div class="card pbox" style="padding:26px; max-width:480px; border:1px solid var(--border-light);">
      <div class="row sb" style="margin-bottom:16px;">
        <h2 style="font-size:20px; margin:0; color:#fff;">${ic('shield', 18)} Setup Two-Factor Authentication</h2>
        <button class="btn sm" onclick="closeModal()" aria-label="Close modal">&times;</button>
      </div>
      <div id="tfa-wizard-content" style="text-align:center; padding:20px 0;">
        <span class="np-spinner-sm"></span> Generating security keys...
      </div>
    </div>
  </div>`;

  try {
    const res = await api('/api/account/2fa/generate', { method: 'POST' });
    const content = $('#tfa-wizard-content');
    if (!content) return;

    const qrUrl = `https://api.qrserver.com/v1/create-qr-code/?size=180x180&data=${encodeURIComponent(res.otpauth_url)}`;
    const formattedSecret = (res.secret.match(/.{1,4}/g) || []).join(' ');

    content.innerHTML = `
      <p class="mu" style="font-size:13px; margin:0 0 16px; text-align:left;">
        1. Scan this QR code in Google Authenticator, Authy, or 1Password. Or enter the secret key manually:
      </p>
      <div style="background:#fff; display:inline-block; padding:10px; border-radius:12px; margin-bottom:14px; box-shadow:0 4px 14px rgba(0,0,0,0.25);">
        <img src="${esc(qrUrl)}" alt="TOTP QR Code" width="180" height="180" style="display:block;">
      </div>
      <div style="margin-bottom:16px;">
        <span class="mu" style="font-size:11px; text-transform:uppercase; letter-spacing:0.5px; display:block; margin-bottom:4px;">Secret Key</span>
        <div class="mono" style="background:var(--con-bg); border:1px solid var(--border); border-radius:6px; padding:8px 12px; font-size:14px; font-weight:700; user-select:all; color:#fff; display:inline-flex; align-items:center; gap:8px;">
          <span>${esc(formattedSecret)}</span>
          <button type="button" class="btn sm" onclick="navigator.clipboard.writeText('${esc(res.secret)}'); toast('Secret copied to clipboard', 'ok');">${ic('copy', 12)}</button>
        </div>
      </div>
      <p class="mu" style="font-size:13px; margin:0 0 10px; text-align:left;">
        2. Enter the 6-digit confirmation code shown on your phone:
      </p>
      <div class="row" style="gap:8px; margin-bottom:16px; justify-content:center;">
        <input type="text" id="tfa-verify-code" maxlength="6" placeholder="000000" autocomplete="one-time-code" style="width:180px; text-align:center; font-size:20px; font-weight:700; letter-spacing:4px;" onkeydown="if(event.key==='Enter')confirmEnable2FA('${esc(res.secret)}')">
        <button type="button" id="tfa-confirm-btn" class="btn pr" onclick="confirmEnable2FA('${esc(res.secret)}')">Activate</button>
      </div>
      <div id="tfa-wizard-err" style="color:var(--err); font-size:13px; min-height:18px;"></div>
    `;
    $('#tfa-verify-code')?.focus();
  } catch (err) {
    const content = $('#tfa-wizard-content');
    if (content) content.innerHTML = `<span style="color:var(--err);">Failed to generate 2FA key: ${esc(err.message)}</span>`;
  }
}

async function confirmEnable2FA(secret) {
  const code = $('#tfa-verify-code')?.value.trim();
  const errEl = $('#tfa-wizard-err');
  const btn = $('#tfa-confirm-btn');
  if (!code || code.length !== 6) {
    if (errEl) errEl.textContent = 'Please enter a 6-digit code from your authenticator app.';
    return;
  }
  if (btn) btn.disabled = true;
  try {
    const res = await api('/api/account/2fa/enable', {
      method: 'POST',
      body: JSON.stringify({ secret, code })
    });
    if (S.user) S.user.two_factor_enabled = true;
    S.acct.two_factor_enabled = true;

    const content = $('#tfa-wizard-content');
    if (content) {
      const codes = res.recovery_codes || [];
      content.innerHTML = `
        <div style="width:48px; height:48px; border-radius:50%; background:rgba(34,197,94,0.15); color:#22c55e; display:grid; place-items:center; margin:0 auto 14px;">
          ${ic('check', 24)}
        </div>
        <h3 style="margin:0 0 8px; color:var(--tx);">Two-Factor Authentication Active!</h3>
        <p class="mu" style="font-size:13px; margin:0 0 16px; text-align:left;">
          Save these <b>emergency recovery codes</b> in a secure password manager. If you ever lose your phone, each code can be used once to log in:
        </p>
        <div class="grid g2" style="background:var(--con-bg); border:1px solid var(--border); border-radius:8px; padding:12px; margin-bottom:16px;">
          ${codes.map(c => `<div class="mono" style="font-size:13px; font-weight:700; color:#fff; padding:4px;">${esc(c)}</div>`).join('')}
        </div>
        <div class="row" style="gap:10px; justify-content:center;">
          <button type="button" class="btn" onclick="navigator.clipboard.writeText('${codes.join('\\n')}'); toast('Recovery codes copied!', 'ok');">${ic('copy', 13)} Copy All Codes</button>
          <button type="button" class="btn pr" onclick="closeModal(); R();">Done</button>
        </div>
      `;
    }
    toast('Two-factor authentication successfully enabled!', 'ok');
  } catch (err) {
    if (btn) btn.disabled = false;
    if (errEl) errEl.textContent = err.message || 'Verification failed. Please check the code and try again.';
  }
}

function openDisable2FAModal() {
  const container = $('#modal-container');
  if (!container) return;

  container.innerHTML = `<div class="pal" onclick="if(event.target===this)closeModal()">
    <div class="card pbox" style="padding:26px; max-width:440px; border:1px solid var(--border-light);">
      <div class="row sb" style="margin-bottom:16px;">
        <h2 style="font-size:20px; margin:0; color:#fff;">${ic('lock', 18)} Disable Two-Factor Auth</h2>
        <button class="btn sm" onclick="closeModal()" aria-label="Close modal">&times;</button>
      </div>
      <p class="mu" style="font-size:13px; margin:0 0 16px;">
        Disabling 2FA reduces your account security. Please verify your identity with your account password and current authenticator code:
      </p>
      <label class="f" style="margin-bottom:12px;">
        <span>Account Password</span>
        <input type="password" id="disable-2fa-pw" required autocomplete="current-password">
      </label>
      <label class="f" style="margin-bottom:18px;">
        <span>6-Digit Code or Recovery Code</span>
        <input type="text" id="disable-2fa-code" maxlength="9" placeholder="000000 or 1a2b-3c4d" autocomplete="one-time-code">
      </label>
      <div id="disable-2fa-err" style="color:var(--err); font-size:13px; margin-bottom:12px;"></div>
      <div class="row" style="justify-content:flex-end; gap:10px;">
        <button type="button" class="btn" onclick="closeModal()">Cancel</button>
        <button type="button" id="disable-2fa-btn" class="btn dg" onclick="confirmDisable2FA()">Disable 2FA</button>
      </div>
    </div>
  </div>`;
  $('#disable-2fa-pw')?.focus();
}

async function confirmDisable2FA() {
  const password = $('#disable-2fa-pw')?.value;
  const code = $('#disable-2fa-code')?.value.trim();
  const errEl = $('#disable-2fa-err');
  const btn = $('#disable-2fa-btn');

  if (!password) {
    if (errEl) errEl.textContent = 'Account password is required.';
    return;
  }
  if (btn) btn.disabled = true;
  try {
    await api('/api/account/2fa/disable', {
      method: 'POST',
      body: JSON.stringify({ password, code })
    });
    if (S.user) S.user.two_factor_enabled = false;
    S.acct.two_factor_enabled = false;
    closeModal();
    toast('Two-factor authentication disabled', 'ok');
    R();
  } catch (err) {
    if (btn) btn.disabled = false;
    if (errEl) errEl.textContent = err.message || 'Failed to disable 2FA';
  }
}

async function updatePassword() {
  const cur = $('#cur-pw')?.value;
  const nw = $('#new-pw')?.value;
  if (!cur || !nw) {
    toast('Please enter your current and new password', 'error');
    return;
  }
  try {
    await api('/api/account/password', {
      method: 'POST',
      body: JSON.stringify({ current_password: cur, new_password: nw })
    });
    toast('Password updated successfully!');
    if ($('#cur-pw')) $('#cur-pw').value = '';
    if ($('#new-pw')) $('#new-pw').value = '';
  } catch (e) {}
}

async function addSSHKey() {
  const name = $('#kn')?.value.trim();
  const key = $('#kk')?.value.trim();
  if (!name || !key) {
    toast('Key name and public key string are required', 'error');
    return;
  }
  try {
    await api('/api/account/ssh-keys', {
      method: 'POST',
      body: JSON.stringify({ name, public_key: key })
    });
    toast('SSH Public Key added successfully!');
    loadAccountData();
  } catch (e) {}
}

async function deleteSSHKey(id) {
  const ok = await uiConfirm({
    title: 'Delete SSH Public Key',
    message: 'Are you sure you want to remove this SSH public key? SFTP clients authenticating with this key will lose access immediately.',
    confirmText: 'Delete Key',
    cancelText: 'Cancel',
    isDanger: true,
    icon: 'key'
  });
  if (!ok) return;

  try {
    await api(`/api/account/ssh-keys/${id}`, { method: 'DELETE' });
    toast('SSH Key deleted');
    loadAccountData();
  } catch (e) {}
}

async function createAPIKey() {
  const desc = $('#ad')?.value.trim();
  if (!desc) {
    toast('Please provide a key description', 'error');
    return;
  }
  try {
    const res = await api('/api/account/api-keys', {
      method: 'POST',
      body: JSON.stringify({ name: desc })
    });
    S.acct.newkey = res.token || res.key || '';
    toast('API Key generated successfully', 'ok');
    loadAccountData();
  } catch (e) {}
}

async function deleteAPIKey(id) {
  const ok = await uiConfirm({
    title: 'Revoke API Key',
    message: 'Are you sure you want to revoke this API key? Any scripts or external integrations using this token will immediately fail authentication.',
    confirmText: 'Revoke Key',
    cancelText: 'Cancel',
    isDanger: true,
    icon: 'key'
  });
  if (!ok) return;

  try {
    await api(`/api/account/api-keys/${id}`, { method: 'DELETE' });
    toast('API Key revoked');
    loadAccountData();
  } catch (e) {}
}

// --- Appearance ---
function setAppearanceMode(mode, skipRender = false) {
  S.acct.mode = mode;
  localStorage.setItem('neopanel_mode', mode);
  const d = document.documentElement;
  d.classList.toggle('oled', mode === 'Oled');
  if (mode === 'Auto') {
    const prefersLight = window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches;
    d.dataset.theme = prefersLight ? 'light' : 'dark';
  } else {
    d.dataset.theme = mode === 'Light' ? 'light' : 'dark';
  }
  applyBranding((S.view === 'admin' && (S.adminTab || 0) === 4 && S.brandDraft) ? S.brandDraft : S.branding);
  if (!skipRender) R();
}

if (window.matchMedia) {
  window.matchMedia('(prefers-color-scheme: light)').addEventListener('change', () => {
    if (S.acct.mode === 'Auto') setAppearanceMode('Auto');
  });
}

function toggleCompact(val) {
  S.acct.compact = val;
  localStorage.setItem('neopanel_compact', val);
  document.documentElement.classList.toggle('compact', val);
}

function toggleSound(val) {
  S.acct.sound = val;
  localStorage.setItem('neopanel_sound', val);
  if (val) beep();
}

function togglePrivacy(val) {
  S.acct.priv = val;
  localStorage.setItem('neopanel_priv', val);
  document.documentElement.classList.toggle('priv', val);
}

function toggleTheme() {
  const cur = document.documentElement.dataset.theme;
  setAppearanceMode(cur === 'light' ? 'Dark' : 'Light');
}

async function switchServer(index) {
  const requested = S.servers[index];
  if (!requested || !await guardEditorLeave()) return;
  const srv = resolveServerByParam(requested.id, requested.node_id || 'local');
  if (!srv) return;
  if (S.serverRoutePending) {
    applyRoute(serverRouteUrl(srv.node_id || 'local', srv.id, 'con'));
    return;
  }
  flushPendingSettingsBeforeLeave();
  syncActiveServerState(srv);

  const serverViews = ['con', 'set', 'act', 'fi', 'db', 'bk', 'sc', 'team', 'net', 'sftp', 'sub', 'stp', 'pl', 'plug', 'ver'];
  if (serverViews.includes(S.view) && !(S.view === 'act' && !window.location.pathname.startsWith('/server/'))) {
    try {
      const nextUrl = serverRouteUrl(srv.node_id || 'local', srv.id, S.view, { plugTab: S.plugTab });
      if ((window.location.pathname + window.location.search) !== nextUrl) {
        history.replaceState(null, '', nextUrl);
      }
    } catch (e) {}
  }

  R();
  if (serverPagePermission(S.view)) onEnterView(S.view);
  toast(`Switched to ${srv.name}`);
}

// --- Command Palette (Ctrl + K) ---
function paletteItems() {
  const q = (S.pq || '').toLowerCase();
  const isModSrv = isModdedEngine(curServer()?.type || S.core);
  const clientLabel = (getBrand('client_area_label') || 'Client Area').trim() || 'Client Area';
  const showClientArea = getBrand('show_client_area_button') !== false;
  const modCmds = Array.isArray(window.NP?._paletteCommands)
    ? window.NP._paletteCommands.map(c => [c.group || 'Extensions', c.label || 'Extension Action', c.subtitle || 'Mod command', () => c.action && c.action()])
    : [];
  const items = [
    ...PG.filter(p => canServerPage(p[0])).map(p => ['Pages', (p[0] === 'plug' && isModSrv) ? 'Mods' : p[1], 'Go to page', () => go(p[0])]),
    ...(showClientArea ? [['Pages', clientLabel, 'External storefront and billing', () => go('client')]] : []),
    ...(S.user?.role === 'admin' ? [['Pages', 'Admin', 'Panel settings, nodes, and users', () => go('admin')]] : []),
    ['Pages', 'Account', 'Profile, appearance, and security', () => go('acc')],
    ...S.servers.map(s => ['Servers', s.name, `${(s.stats?.status || s.status) === 'running' ? 'Online' : 'Offline'} (Port ${s.port})`, () => {
      applyRoute(serverRouteUrl(s.node_id || 'local', s.id, 'con'));
    }]),
    ...S.players.map(p => ['Players', p.n, p.op ? 'Operator' : 'Online', () => { S.ptab = 'online'; go('pl'); }]),
    ...(hasServerPermission('console.view') ? [['Actions', 'Toggle Live Console', 'Ctrl + `', () => toggleFlyConsole()]] : []),
    ...(hasServerPermission('backup.manage') ? [['Actions', 'Create Backup', 'Backups', () => createBackup()]] : []),
    ...(hasServerPermission('server.power') ? [['Actions', 'Restart Server', 'Power', () => power('restart')]] : []),
    ...(canCreateServers() ? [['Actions', 'Create server', 'New Minecraft server', () => openCreateServerModal()]] : []),
    ['Actions', 'Toggle Dark / Light Theme', 'Appearance', () => toggleTheme()],
    ['Actions', 'Toggle Privacy Mode', 'Privacy', () => togglePrivacy(!S.acct.priv)],
    ...modCmds
  ];
  return items.filter(x => (x[1] + ' ' + x[0]).toLowerCase().includes(q));
}

function renderPaletteList() {
  const box = $('#pl2');
  if (!box) return;
  const items = paletteItems();
  let currentGroup = '';

  box.innerHTML = items.map((item, i) => {
    const header = item[0] !== currentGroup ? `<div class="pg">${item[0]}</div>` : '';
    currentGroup = item[0];
    return `${header}
    <button class="pit ${i === S.pi ? 'on' : ''}" onclick="execPaletteItem(${i})">
      <span>${esc(item[1])}</span>
      <span class="mu">${esc(item[2])}</span>
    </button>`;
  }).join('') || '<div class="mu" style="padding:16px;">Nothing found. Try a page, server, or action name.</div>';

  const selected = $('#pl2 .on');
  if (selected) selected.scrollIntoView({ block: 'nearest' });
}

function execPaletteItem(idx) {
  const items = paletteItems();
  const target = items[idx];
  openPalette(0);
  if (target && target[3]) target[3]();
}

function openPalette(open) {
  let p = $('#pal');
  if (!open) {
    if (p) p.remove();
    if (p && S.paletteReturnFocus?.isConnected) S.paletteReturnFocus.focus();
    return;
  }
  if (p || $('#login')) return;
  S.paletteReturnFocus = document.activeElement;

  p = document.createElement('div');
  p.id = 'pal';
  p.className = 'pal';
  p.innerHTML = `<div class="pbox" role="dialog" aria-modal="true" aria-label="Command Palette">
    <div class="np-palette-input"><input id="pq" class="pin" aria-label="Search servers, pages, players, actions" placeholder="Search servers, pages, players, actions..." autocomplete="off"></div>
    <div id="pl2" class="plist"></div>
  </div>`;
  document.body.appendChild(p);

  S.pq = '';
  S.pi = 0;
  renderPaletteList();

  p.addEventListener('keydown', e => {
    if (e.key !== 'Tab') return;
    const controls = [...p.querySelectorAll('input, button')];
    const first = controls[0], last = controls.at(-1);
    if (e.shiftKey && document.activeElement === first) { e.preventDefault(); last.focus(); }
    else if (!e.shiftKey && document.activeElement === last) { e.preventDefault(); first.focus(); }
  });
  const inp = $('#pq');
  inp.focus();
  inp.oninput = () => {
    S.pq = inp.value;
    S.pi = 0;
    renderPaletteList();
  };
  inp.onkeydown = e => {
    const items = paletteItems();
    if (e.key === 'ArrowDown') {
      S.pi = Math.min(items.length - 1, S.pi + 1);
      renderPaletteList();
      e.preventDefault();
    } else if (e.key === 'ArrowUp') {
      S.pi = Math.max(0, S.pi - 1);
      renderPaletteList();
      e.preventDefault();
    } else if (e.key === 'Enter' && items[S.pi]) {
      execPaletteItem(S.pi);
    }
  };
  p.onclick = e => {
    if (e.target === p) openPalette(0);
  };
}

window.addEventListener('keydown', e => {
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
    e.preventDefault();
    openPalette(1);
  } else if ((e.ctrlKey || e.metaKey) && (e.code === 'Backquote' || e.key === '`')) {
    if (!$('#login')) {
      e.preventDefault();
      toggleFlyConsole();
    }
  } else if (e.key === 'Escape') {
    if ($('#pal') || $('#modal-container')?.innerHTML.trim()) {
      openPalette(0);
      closeModal();
    } else if (S.flyConOpen) {
      closeFlyConsole();
    }
  }
});

// --- Flexible Server Creation Modal (With Proxies & Custom Ports) ---
function setCreateRam(mb) {
  const inp = $('#cs-mem');
  if (inp) inp.value = mb;
  $$('.ram-pill').forEach(btn => {
    btn.classList.toggle('pr', btn.dataset.ram == mb);
  });
}

function syncCreateJavaFromVersion(ver) {
  const engine = $('#cs-type')?.value || 'paper';
  if (isSteelEngine(engine)) return;
  const v = String(ver ?? $('#cs-ver')?.value ?? '').trim();
  $$('.ver-pill').forEach(btn => {
    btn.classList.toggle('pr', btn.dataset.ver === v);
  });
  const sel = $('#cs-java');
  if (!sel || sel.dataset.userPicked === '1') return;
  const rec = getRecommendedJavaForMcVersion(v, engine);
  if (sel.value !== rec) {
    sel.value = rec;
    sel.dispatchEvent(new Event('change', { bubbles: true }));
  }
}

function syncCreateSelectValue(ver) {
  const sel = $('#cs-ver-sel');
  if (!sel) return;
  const v = String(ver || '').trim();
  if (v && Array.from(sel.options).some(o => o.value === v)) {
    sel.value = v;
    if (typeof sel._cselectRebuild === 'function') sel._cselectRebuild();
  }
}

function onCreateVersionInput(ver) {
  syncCreateSelectValue(ver);
  syncCreateJavaFromVersion(ver);
}

function setCreateEngine(type) {
  const sel = $('#cs-type');
  const previousType = sel?.value;
  if (sel) sel.value = type;
  $$('.engine-card').forEach(card => {
    card.classList.toggle('on', card.dataset.engine === type);
  });

  const verInp = $('#cs-ver');
  const portInp = $('#cs-port');
  const native = isSteelEngine(type);
  for (const id of ['cs-java', 'cs-aikar']) {
    const input = $('#' + id);
    if (input) { input.disabled = native; input.closest('label').style.display = native ? 'none' : ''; }
  }
  const strip = $('#cs-java-status');
  if (strip) strip.style.display = native ? 'none' : '';
  const note = $('#cs-steel-note');
  if (note) { note.style.display = native ? '' : 'none'; note.innerHTML = native ? steelConfigNotice(false) : ''; }
  const label = $('#cs-version-label');
  if (label) label.textContent = native ? 'Steel Release Tag' : 'Minecraft Version (Recent & Popular)';
  $$('.ver-pill').forEach(btn => { btn.style.display = native ? 'none' : ''; });
  if (verInp) verInp.placeholder = native ? 'latest or official Steel release tag' : 'Minecraft version or latest';
  if (native) { if (verInp) verInp.value = 'latest'; loadSteelVersions(); }
  else if (isSteelEngine(previousType) && verInp) verInp.value = '1.21.11';
  if (native) {
    // Steel release tags are independent of Minecraft versions.
  } else if (type === 'velocity') {
    if (verInp) verInp.value = '3.4.0-SNAPSHOT';
    if (portInp) portInp.value = '25565';
    setCreateRam(1024);
  } else if (type === 'waterfall') {
    if (verInp) verInp.value = '1.20';
    if (portInp) portInp.value = '25577';
    setCreateRam(1024);
  } else if (type === 'bungeecord') {
    if (verInp) verInp.value = 'latest';
    if (portInp) portInp.value = '25577';
    setCreateRam(1024);
  } else if (verInp) {
    const cur = verInp.value.trim();
    if (!cur || cur === '3.4.0-SNAPSHOT' || cur === 'latest' || cur === '1.20') {
      verInp.value = '1.21.11';
    }
  }

  const csVerSel = $('#cs-ver-sel');
  if (csVerSel) {
    const curVer = verInp?.value?.trim() || '1.21.11';
    const vers = getAllMcVersions(type);
    const list = (curVer && !vers.includes(curVer)) ? [curVer, ...vers] : vers;
    csVerSel.innerHTML = list.map(v => `<option value="${esc(v)}" ${v === curVer ? 'selected' : ''}>${esc(v)}</option>`).join('');
    if (typeof csVerSel._cselectRebuild === 'function') csVerSel._cselectRebuild();
  }
  syncCreateJavaFromVersion(verInp?.value || '');
}

function setCreateVersion(ver) {
  const inp = $('#cs-ver');
  if (inp) inp.value = ver;
  syncCreateSelectValue(ver);
  syncCreateJavaFromVersion(ver);
}

function getNextSuggestedPort(targetNodeId = null) {
  const nid = targetNodeId || curServer()?.node_id || 'local';
  const usedPorts = new Set();
  (S.servers || []).forEach(s => {
    if ((s.node_id || 'local') === nid && s.port) usedPorts.add(s.port);
  });
  if ((curServer()?.node_id || 'local') === nid) {
    (S.ports || []).forEach(p => {
      if (p.port) usedPorts.add(p.port);
    });
  }
  let candidate = 25565;
  while (usedPorts.has(candidate)) {
    candidate++;
  }
  return candidate;
}

function onCreateNodeChange(nodeId) {
  const nid = nodeId || 'local';
  const portInp = $('#cs-port');
  if (portInp) portInp.value = getNextSuggestedPort(nid);
  if (isSteelEngine($('#cs-type')?.value)) return;
  const strip = document.querySelector('#modal-container .java-status-strip');
  if (strip) {
    strip.dataset.nodeId = nid;
    if (!S.javaRuntimesMap?.[nid]) {
      strip.innerHTML = '<span class="java-status-chip">Checking installed Java 25, 21, 17 &amp; 8 runtimes...</span>';
    }
  }
  loadJavaRuntimes(false, nid);
}

function openCreateServerModal() {
  if (!canCreateServers()) { toast('Access restricted: server creation permission required', 'bad'); return; }
  const container = $('#modal-container');
  if (!container) return;

  const nodesList = (Array.isArray(S.nodes) && S.nodes.length > 0)
    ? S.nodes
    : [{ id: 'local', name: 'Local node', location: 'This machine', online: true }];
  const defaultNodeId = (S.nodeFilter && S.nodeFilter !== 'all' && nodesList.some(n => n.id === S.nodeFilter))
    ? S.nodeFilter
    : (nodesList.find(n => n.id === 'local')?.id || nodesList[0].id);
  const nextPort = getNextSuggestedPort(defaultNodeId);
  const defaultRam = Number(S.brandingDaemon?.default_ram_mb) || Number(S.adminDaemonCfg?.defaultRamMB) || 2048;
  const initialVers = getAllMcVersions('paper');

  container.innerHTML = `<div class="pal" onclick="if(event.target===this)closeModal()">
    <div class="card pbox" style="padding:26px; max-width:680px; max-height:88vh; overflow-y:auto;">
      <div class="row sb" style="margin-bottom:16px;">
        <div>
          <h2 style="font-size:22px; margin:0;">Create Server or Proxy</h2>
          <span class="mu" style="font-size:13px;">Deploy any Minecraft server engine &amp; release version</span>
        </div>
        <button class="btn sm" onclick="closeModal()">✕</button>
      </div>

      <form onsubmit="handleCreateServer(event)">
        <!-- Instance Name & Target Node -->
        <div class="${nodesList.length > 1 ? 'grid g2' : ''}" style="margin-bottom:14px;">
          <label class="f" style="margin-bottom:0;">
            <span>Server name</span>
            <input type="text" id="cs-name" required placeholder="e.g. Survival SMP, Proxy Network, Minigames" value="Survival Server">
          </label>
          ${nodesList.length > 1 ? `
          <label class="f" style="margin-bottom:0;">
            <span>Node</span>
            <select id="cs-node" onchange="onCreateNodeChange(this.value)">
              ${nodesList.map(n => `<option value="${esc(n.id)}" ${n.id === defaultNodeId ? 'selected' : ''}>${esc(displayNodeName(n.id, n.name))} (${esc(displayNodeLocation(n))})${n.online ? '' : ' [Offline]'}</option>`).join('')}
            </select>
          </label>` : `<input type="hidden" id="cs-node" value="${esc(defaultNodeId)}">`}
        </div>

        <!-- Game Servers & Network Proxies -->
        <div style="margin-bottom:14px;">
          <span class="mu" style="display:block; margin-bottom:6px; font-size:13px; font-weight:600;">Engine / Flavor</span>
          <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(118px, 1fr)); gap:8px;">
            ${CORES.map(([label, desc]) => {
              const slug = label.toLowerCase();
              return `
              <button type="button" class="card engine-card sel ${slug === 'paper' ? 'on' : ''}" data-engine="${esc(slug)}" onclick="setCreateEngine('${esc(slug)}')" style="padding:10px 8px; text-align:center; display:flex; flex-direction:column; align-items:center; gap:5px; min-width:0;">
                ${engineIcon(slug, 18, true)}
                <b style="font-size:13px;">${esc(label)}</b>
                <span class="mu" style="font-size:10.5px; display:block; max-width:100%; overflow:hidden; text-overflow:ellipsis; white-space:nowrap;" title="${esc(desc)}">${esc(desc)}</span>
              </button>`;
            }).join('')}
          </div>
          <select id="cs-type" style="display:none;">
            ${CORES.map(([label]) => {
              const slug = label.toLowerCase();
              return `<option value="${esc(slug)}" ${slug === 'paper' ? 'selected' : ''}>${esc(slug)}</option>`;
            }).join('')}
          </select>
        </div>
        <div id="cs-steel-note" style="display:none; margin-bottom:14px;"></div>

        <!-- Version Searchable Catalog + Custom Input + Popular Pills -->
        <div style="margin-bottom:14px;">
          <div class="row sb" style="margin-bottom:6px; flex-wrap:wrap; gap:6px;">
            <span id="cs-version-label" class="mu" style="font-size:13px; font-weight:600;">Minecraft Version (Recent &amp; Popular)</span>
            <div class="row" style="gap:4px; flex-wrap:wrap;">
              ${['1.21.11', '1.21.8', '1.21.4', '1.20.4', '1.20.1', '1.19.4', '1.16.5', '1.12.2', '1.8.9'].map(v => `
                <button type="button" class="btn sm ver-pill ${v === '1.21.11' ? 'pr' : ''}" data-ver="${v}" onclick="setCreateVersion('${v}')">${v}</button>
              `).join('')}
            </div>
          </div>
          <div style="display:grid; grid-template-columns: minmax(145px, 185px) 1fr; gap:8px;">
            <select id="cs-ver-sel" data-searchable="1" onchange="setCreateVersion(this.value)">
              ${initialVers.map(v => `<option value="${esc(v)}" ${v === '1.21.11' ? 'selected' : ''}>${esc(v)}</option>`).join('')}
            </select>
            <input type="text" id="cs-ver" value="1.21.11" required placeholder="e.g. 1.21.11, 1.20.4, 1.16.5, 1.12.2, 1.8.9, or custom" oninput="onCreateVersionInput(this.value)">
          </div>
        </div>

        <!-- Custom Port Allocation -->
        <div class="grid g2" style="margin-bottom:10px;">
          <label class="f" style="margin-bottom:0;">
            <span>Custom Port (1024 - 65535)</span>
            <input type="number" id="cs-port" value="${nextPort}" min="1024" max="65535" required>
          </label>
          <label class="f" style="margin-bottom:0;">
            <span>Java Runtime</span>
            <select id="cs-java" onchange="this.dataset.userPicked='1'">
              ${renderJavaSelectOptions('21', false, '1.21.11', 'paper', defaultNodeId)}
            </select>
          </label>
        </div>
        <div id="cs-java-status" style="margin-bottom:14px;">
          ${renderJavaAvailabilityStrip(defaultNodeId)}
        </div>

        <!-- Memory Allocation Pills -->
        <div style="margin-bottom:14px;">
          <div class="row sb" style="margin-bottom:6px;">
            <span class="mu" style="font-size:13px; font-weight:600;">Memory Allocation</span>
            <div class="row" style="gap:4px;">
              <button type="button" class="btn sm ram-pill ${defaultRam === 1024 ? 'pr' : ''}" data-ram="1024" onclick="setCreateRam(1024)">1 GB</button>
              <button type="button" class="btn sm ram-pill ${defaultRam === 2048 ? 'pr' : ''}" data-ram="2048" onclick="setCreateRam(2048)">2 GB</button>
              <button type="button" class="btn sm ram-pill ${defaultRam === 4096 ? 'pr' : ''}" data-ram="4096" onclick="setCreateRam(4096)">4 GB</button>
              <button type="button" class="btn sm ram-pill ${defaultRam === 8192 ? 'pr' : ''}" data-ram="8192" onclick="setCreateRam(8192)">8 GB</button>
            </div>
          </div>
          <input type="number" id="cs-mem" value="${defaultRam}" min="512" max="65536" step="512" required placeholder="RAM in MB">
        </div>

        <!-- Resource Limits: CPU Percentage & Storage Allocation -->
        <div class="grid g2" style="margin-bottom:14px;">
          <label class="f" style="margin-bottom:0;">
            <div class="row sb" style="margin-bottom:4px;">
              <span>CPU Limit (%)</span>
              <span class="mu" style="font-size:11.5px;">0% = Unlimited</span>
            </div>
            <div class="row" style="gap:5px; margin-bottom:6px; flex-wrap:wrap;">
              <button type="button" class="btn sm" style="font-size:11px; padding:3px 7px;" onclick="$('#cs-cpu').value=0">Unlimited</button>
              <button type="button" class="btn sm" style="font-size:11px; padding:3px 7px;" onclick="$('#cs-cpu').value=100">100% (1 Core)</button>
              <button type="button" class="btn sm" style="font-size:11px; padding:3px 7px;" onclick="$('#cs-cpu').value=200">200% (2 Cores)</button>
            </div>
            <input type="number" id="cs-cpu" value="0" min="0" max="3200" step="10" placeholder="0 for unlimited">
          </label>
          <label class="f" style="margin-bottom:0;">
            <div class="row sb" style="margin-bottom:4px;">
              <span>Storage / Disk Limit (MB)</span>
              <span class="mu" style="font-size:11.5px;">0 = Unlimited</span>
            </div>
            <div class="row" style="gap:5px; margin-bottom:6px; flex-wrap:wrap;">
              <button type="button" class="btn sm" style="font-size:11px; padding:3px 7px;" onclick="$('#cs-disk').value=0">Unlimited</button>
              <button type="button" class="btn sm" style="font-size:11px; padding:3px 7px;" onclick="$('#cs-disk').value=5120">5 GB</button>
              <button type="button" class="btn sm" style="font-size:11px; padding:3px 7px;" onclick="$('#cs-disk').value=10240">10 GB</button>
              <button type="button" class="btn sm" style="font-size:11px; padding:3px 7px;" onclick="$('#cs-disk').value=20480">20 GB</button>
            </div>
            <input type="number" id="cs-disk" value="0" min="0" step="512" placeholder="0 for unlimited (in MB)">
          </label>
        </div>

        ${(S.user?.role === 'admin' && Array.isArray(S.adminUsers) && S.adminUsers.length > 0) ? `
        <!-- Server Ownership (Admin Only) -->
        <div style="margin-bottom:14px;">
          <label class="f" style="margin-bottom:0;">
            <span class="row sb">
              <span>Server Owner</span>
              <span class="mu" style="font-size:11.5px;">Assign server instance ownership</span>
            </span>
            <select id="cs-owner">
              ${S.adminUsers.map(u => `<option value="${u.id}" ${u.id === (S.user?.id || 1) ? 'selected' : ''}>${esc(u.username)}${(u.first_name || u.last_name) ? ` (${esc([u.first_name, u.last_name].filter(Boolean).join(' '))})` : ''} [${esc(u.role)}]</option>`).join('')}
            </select>
          </label>
        </div>` : ''}

        <!-- EULA & Optimization Flags -->
        <div style="background:var(--pn2); border:1px solid var(--ln); border-radius:8px; padding:12px; margin-bottom:16px;">
          <label class="row" style="margin-bottom:8px; cursor:pointer;">
            <input type="checkbox" id="cs-eula" checked style="width:auto; margin-right:8px;">
            <span>Automatically agree to Mojang EULA (<code>eula=true</code>)</span>
          </label>
          <label class="row" style="margin-bottom:8px; cursor:pointer;">
            <input type="checkbox" id="cs-dl" checked style="width:auto; margin-right:8px;">
            <span>Automatically download official server executable immediately</span>
          </label>
          <label class="row" style="cursor:pointer;">
            <input type="checkbox" id="cs-aikar" checked style="width:auto; margin-right:8px;">
            <span>Apply Aikar's optimized G1GC flags</span>
          </label>
        </div>

        <div class="row" style="justify-content:flex-end; gap:8px;">
          <button type="button" class="btn" onclick="closeModal()">Cancel</button>
          <button type="submit" class="btn pr" id="cs-submit-btn">🚀 Create &amp; Launch Server</button>
        </div>
      </form>
    </div>
  </div>`;

  enhanceSelects(container);
  loadVersionsCatalog();
  loadJavaRuntimes(false, defaultNodeId);
  setTimeout(() => $('#cs-name')?.focus(), 50);
}

function closeModal() {
  if (typeof abortPasskeyOperation === 'function') abortPasskeyOperation();
  closeAllCustomSelects();
  const container = $('#modal-container');
  if (container) container.innerHTML = '';
  if (_uiModalResolver) {
    const resolve = _uiModalResolver;
    _uiModalResolver = null;
    resolve(null);
  }
}

async function handleCreateServer(e) {
  if (!canCreateServers()) { e?.preventDefault(); toast('Access restricted: server creation permission required', 'bad'); return; }
  e.preventDefault();
  const btn = $('#cs-submit-btn');
  if (btn) {
    btn.disabled = true;
    btn.textContent = 'Creating server...';
  }

  const targetNodeId = $('#cs-node')?.value || 'local';
  const name = $('#cs-name')?.value.trim() || 'Minecraft Server';
  const type = $('#cs-type')?.value || 'paper';
  const version = $('#cs-ver')?.value.trim() || (isSteelEngine(type) ? 'latest' : '1.21.11');
  const memoryMax = parseInt($('#cs-mem')?.value) || 2048;
  const memoryMin = isSteelEngine(type) ? 0 : Math.min(1024, memoryMax);
  const port = parseInt($('#cs-port')?.value) || 25565;
  const javaPath = isSteelEngine(type) ? '' : ($('#cs-java')?.value || '21');
  const autoDownload = $('#cs-dl')?.checked ?? true;
  const useAikar = $('#cs-aikar')?.checked;
  const cpuLimit = parseInt($('#cs-cpu')?.value) || 0;
  const diskLimit = parseInt($('#cs-disk')?.value) || 0;
  const ownerId = parseInt($('#cs-owner')?.value) || 0;

  let javaArgs = '-XX:+UseG1GC';
  if (useAikar) {
    javaArgs = '-XX:+UseG1GC -XX:+ParallelRefProcEnabled -XX:MaxGCPauseMillis=200 -XX:+UnlockExperimentalVMOptions -XX:+DisableExplicitGC -XX:+AlwaysPreTouch';
  }
  if (isSteelEngine(type)) javaArgs = '';

  toast(`Creating and configuring ${name}...`);

  try {
    const payload = {
      name,
      type,
      version,
      memory_min: memoryMin,
      memory_max: memoryMax,
      port,
      java_path: javaPath,
      java_args: javaArgs,
      auto_download: autoDownload,
      auto_restart: true,
      cpu_limit_percent: cpuLimit,
      disk_limit_mb: diskLimit
    };
    if (ownerId > 0) {
      payload.owner_id = ownerId;
    }

    const res = await api(nodeApiPath(targetNodeId, '/api/servers'), {
      method: 'POST',
      body: JSON.stringify(payload)
    });

    closeModal();
    toast(`Server ${name} created successfully!`);

    await loadDashboard(true);

    let createdSrv = S.servers.find(s => s.id === res.id && (s.node_id || 'local') === targetNodeId);
    if (!createdSrv) {
      createdSrv = S.servers.find(s => s.name === name && (s.node_id || 'local') === targetNodeId);
    }
    const createdId = createdSrv?.id || res?.id;
    if (createdId) {
      if (createdSrv) syncActiveServerState(createdSrv);
      else saveActiveServerRef({ id: createdId, node_id: targetNodeId });
      applyRoute(serverRouteUrl(targetNodeId, createdId, 'con'));
    } else {
      applyRoute('/');
    }
  } catch (err) {
    if (btn) {
      btn.disabled = false;
      btn.textContent = '🚀 Create & Launch Server';
    }
    toast(err.message || 'Server creation failed', 'bad');
    uiAlert({
      title: 'Server Creation Failed',
      message: `Failed to create server <b>${esc(name)}</b>:<br><br>
        <div style="background:rgba(239,68,68,0.1); border:1px solid rgba(239,68,68,0.3); border-radius:8px; padding:10px 12px; margin-bottom:12px; font-size:13px; color:#fca5a5;">
          ${esc(err.message)}
        </div>
        Please verify that port <b>${port}</b> is free and disk space is available.`,
      type: 'error'
    });
  }
}

// --- Port Allocation Management ---
async function loadServerPorts(skipRender = false) {
  const srv = curServer();
  if (!srv) return;
  try {
    const list = await api(srvApiPath(srv, '/ports'));
    if (Array.isArray(list)) {
      S.ports = list;
      if (!skipRender && S.view === 'net') R();
    }
  } catch (err) {
    console.error('Failed to load server ports:', err);
  }
}

function openAddPortModal() {
  const container = $('#modal-container');
  if (!container) return;

  const srv = curServer();
  const nextPort = getNextSuggestedPort(srv?.node_id || 'local');

  container.innerHTML = `<div class="pal" onclick="if(event.target===this)closeModal()">
    <div class="card pbox" style="padding:26px; max-width:520px;">
      <div class="row sb" style="margin-bottom:16px;">
        <h2 style="font-size:20px; margin:0; color:#fff;">Assign Port Allocation</h2>
        <button class="btn sm" onclick="closeModal()" aria-label="Close modal">&times;</button>
      </div>

      <p class="mu" style="font-size:13px; margin-bottom:16px;">
        Assign an additional incoming network port to <b>${esc(srv?.name || 'this server')}</b>.
      </p>

      <form id="add-port-form" onsubmit="submitAddPort(event)">
        <label class="f" style="margin-bottom:14px;">
          <span>Port Number (1024 - 65535)</span>
          <input type="number" id="ap-port" min="1024" max="65535" required value="${nextPort}">
        </label>

        <!-- Quick Preset Chips -->
        <div style="margin-bottom:14px;">
          <span class="mu" style="display:block; margin-bottom:6px; font-size:12px;">Quick Presets:</span>
          <div class="row" style="flex-wrap:wrap; gap:6px;">
            <button type="button" class="btn sm" style="font-size:11.5px; padding:4px 8px;" onclick="setPortPreset(24454, 'Simple Voice Chat')">Voice Chat (24454)</button>
            <button type="button" class="btn sm" style="font-size:11.5px; padding:4px 8px;" onclick="setPortPreset(8123, 'Dynmap Web Map')">Dynmap (8123)</button>
            <button type="button" class="btn sm" style="font-size:11.5px; padding:4px 8px;" onclick="setPortPreset(19132, 'Geyser Bedrock')">Bedrock (19132)</button>
            <button type="button" class="btn sm" style="font-size:11.5px; padding:4px 8px;" onclick="setPortPreset(8192, 'NuVotifier')">Votifier (8192)</button>
            <button type="button" class="btn sm" style="font-size:11.5px; padding:4px 8px;" onclick="setPortPreset(8100, 'BlueMap')">BlueMap (8100)</button>
          </div>
        </div>

        <label class="f" style="margin-bottom:14px;">
          <span>Protocol</span>
          <select id="ap-proto">
            <option value="tcp/udp">TCP / UDP (Recommended)</option>
            <option value="tcp">TCP only</option>
            <option value="udp">UDP only (e.g. Voice Chat)</option>
          </select>
        </label>

        <label class="f" style="margin-bottom:14px;">
          <span>Service / Notes</span>
          <input type="text" id="ap-notes" placeholder="e.g. Simple Voice Chat, Dynmap, Geyser" value="Additional Allocation">
        </label>

        <label class="row" style="gap:8px; margin-bottom:20px; cursor:pointer;">
          <input type="checkbox" id="ap-primary">
          <span style="font-size:13px; color:#fff;">Set as primary game port (updates server.properties)</span>
        </label>

        <div class="row" style="justify-content:flex-end; gap:10px;">
          <button type="button" class="btn" onclick="closeModal()">Cancel</button>
          <button type="submit" id="ap-submit-btn" class="btn pr">${ic('plus', 14)} Assign Port</button>
        </div>
      </form>
    </div>
  </div>`;

  enhanceSelects(container);
  setTimeout(() => $('#ap-port')?.focus(), 50);
}

function setPortPreset(port, notes) {
  const pInp = $('#ap-port');
  const nInp = $('#ap-notes');
  if (pInp) pInp.value = port;
  if (nInp) nInp.value = notes;
}

async function submitAddPort(e) {
  e.preventDefault();
  const srv = curServer();
  if (!srv) return;

  const btn = $('#ap-submit-btn');
  if (btn) btn.disabled = true;

  const port = parseInt($('#ap-port')?.value, 10);
  const protocol = $('#ap-proto')?.value || 'tcp/udp';
  const notes = $('#ap-notes')?.value.trim() || 'Additional Allocation';
  const isPrimary = $('#ap-primary')?.checked || false;

  if (isNaN(port) || port < 1024 || port > 65535) {
    toast('Invalid port: must be between 1024 and 65535', 'bad');
    if (btn) btn.disabled = false;
    return;
  }

  try {
    await api(srvApiPath(srv, '/ports'), {
      method: 'POST',
      body: JSON.stringify({ port, protocol, notes, is_primary: isPrimary })
    });
    toast(`Port ${port} successfully assigned!`, 'ok');
    closeModal();
    if (isPrimary) {
      srv.port = port;
    }
    await loadServerPorts();
    R();
  } catch (err) {
    toast(err.message || 'Failed to assign port', 'bad');
    if (btn) btn.disabled = false;
    uiAlert({
      title: 'Port Allocation Failed',
      message: `Failed to allocate port <b>${port}</b>:<br><br>${esc(err.message || 'Port is already assigned or unavailable.')}`,
      type: 'error'
    });
  }
}

async function deletePort(portId, portNum) {
  const srv = curServer();
  if (!srv) return;

  const ok = await uiConfirm({
    title: 'Delete Port Allocation',
    message: `Are you sure you want to delete port allocation <b>${esc(portNum)}</b>? Any external connections using this port will stop working.`,
    confirmText: 'Delete Port',
    cancelText: 'Cancel',
    isDanger: true
  });
  if (!ok) return;

  try {
    await api(srvApiPath(srv, `/ports/${portId}`), { method: 'DELETE' });
    toast(`Port ${portNum} deleted`, 'ok');
    await loadServerPorts();
    R();
  } catch (err) {
    toast(err.message || 'Failed to delete port', 'bad');
  }
}

async function setPrimaryPort(portId) {
  const srv = curServer();
  if (!srv) return;
  const ok = await uiConfirm({
    title: 'Change Primary Game Port',
    message: `Set this port allocation as the primary Minecraft game port for <b>${esc(srv.name)}</b>? This updates <code>server.properties</code> and requires a server restart to take effect.`,
    confirmText: 'Set Primary Port',
    cancelText: 'Cancel',
    isDanger: false,
    icon: 'network'
  });
  if (!ok) return;

  try {
    const res = await api(srvApiPath(srv, `/ports/${portId}/primary`), { method: 'POST' });
    if (res && res.port) {
      srv.port = res.port;
    }
    toast(`Primary game port updated to ${res.port}`, 'ok');
    await loadServerPorts();
    R();
  } catch (err) {
    toast(err.message || 'Failed to update primary port', 'bad');
  }
}

// --- Server Schedules Management ---
async function loadServerSchedules(srvId = null, skipRender = false) {
  const srv = srvId ? S.servers.find(s => s.id === srvId && (s.node_id || 'local') === (curServer()?.node_id || 'local')) : curServer();
  if (!srv) return;
  clearTimeout(QOL.scheduleTimer);
  if (!S.loading) S.loading = {};
  S.loading.schedules = true;
  if (!skipRender && S.view === 'sc') R();
  try {
    const list = await api(srvApiPath(srv, '/schedules'));
    if (!sameQolServer(srv, curServer())) return;
    S.serverSchedules = Array.isArray(list) ? list : (Array.isArray(list?.schedules) ? list.schedules : []);
  } catch (err) {
    if (!sameQolServer(srv, curServer())) return;
    console.error('Failed to load server schedules:', err);
    S.serverSchedules = [];
  } finally {
    if (sameQolServer(srv, curServer())) {
      S.loading.schedules = false;
      if (!skipRender && S.view === 'sc') R();
    }
    if (S.view === 'sc' && sameQolServer(srv, curServer())) {
      const running = S.serverSchedules?.some(s => ['started', 'running'].includes(s.last_status));
      QOL.scheduleTimer = setTimeout(() => {
        if (S.view === 'sc' && sameQolServer(srv, curServer())) loadServerSchedules();
      }, running ? 2000 : 30000);
    }
  }
}

function describeCron(expr) {
  if (!expr) return 'Custom schedule';
  const e = expr.trim();
  if (e === '0 * * * *') return 'Hourly (at minute 0)';
  if (e === '0 4 * * *') return 'Daily at 04:00 AM';
  if (e === '0 0 * * *') return 'Daily at Midnight (00:00)';
  if (e === '0 0 * * 0') return 'Weekly (Sundays at 00:00)';
  if (e === '*/30 * * * *') return 'Every 30 minutes';
  if (e === '*/15 * * * *') return 'Every 15 minutes';
  if (e === '*/5 * * * *') return 'Every 5 minutes';
  return 'Cron: ' + e;
}

function getScheduleActionBadge(action) {
  if (action === 'command') return `<span class="chip" style="color: var(--ac); border-color: var(--border-accent);">${ic('terminal', 11)} Command</span>`;
  if (action === 'restart') return `<span class="chip" style="color: var(--wn); border-color: var(--wn-glow);">${ic('refresh', 11)} Restart</span>`;
  if (action === 'backup') return `<span class="chip" style="color: var(--ok); border-color: var(--ok-glow);">${ic('archive', 11)} Backup</span>`;
  if (action === 'stop') return `<span class="chip" style="color: var(--er); border-color: var(--er-glow);">${ic('stop', 11)} Stop</span>`;
  return `<span class="chip">${esc(action)}</span>`;
}

function openScheduleModal(schedId = null) { return openScheduleChainModal(schedId); }

function onScheduleActionChange(act) {
  const wrap = $('#sch-payload-wrap');
  if (wrap) wrap.style.display = (act === 'command') ? 'block' : 'none';
}

async function submitScheduleModal(event, schedId = '') { return submitScheduleChain(event); }

async function toggleServerSchedule(schedId, active) {
  const srv = curServer();
  if (!srv) return;
  try {
    await api(srvApiPath(srv, `/schedules/${encodeURIComponent(schedId)}/toggle`), {
      method: 'POST',
      body: JSON.stringify({ is_active: Boolean(active) })
    });
    toast(active ? 'Schedule activated' : 'Schedule paused');
    await loadServerSchedules(srv.id);
  } catch (err) {
    toast(err.message || 'Failed to toggle schedule', 'bad');
    await loadServerSchedules(srv.id);
  }
}

async function runScheduleNow(schedId) {
  const srv = curServer();
  if (!srv) return;
  try {
    toast('Executing schedule now...');
    const res = await api(srvApiPath(srv, `/schedules/${encodeURIComponent(schedId)}/run`), { method: 'POST' });
    toast(res.message || 'Schedule accepted; waiting for completion.', 'info');
    await loadServerSchedules(srv.id);
  } catch (err) {
    toast(err.message || 'Failed to execute schedule', 'bad');
  }
}

async function deleteServerSchedule(schedId) {
  const srv = curServer();
  if (!srv) return;
  const ok = await uiConfirm({
    title: 'Delete Automated Schedule',
    message: 'Are you sure you want to delete this automated schedule? This action cannot be undone.',
    confirmText: 'Delete Schedule',
    isDanger: true,
    icon: 'trash'
  });
  if (!ok) return;

  try {
    await api(srvApiPath(srv, `/schedules/${encodeURIComponent(schedId)}`), { method: 'DELETE' });
    toast('Schedule deleted', 'ok');
    await loadServerSchedules(srv.id);
  } catch (err) {
    toast(err.message || 'Failed to delete schedule', 'bad');
  }
}

// --- Server Subusers & Team Access ---
const SUBUSER_PERMS = [
  ['console.view', 'View Console', 'Access live server console and logs'],
  ['console.command', 'Send Commands', 'Execute commands in the live console'],
  ['server.power', 'Power Controls', 'Start, stop, restart, and kill server'],
  ['files.view', 'Browse Files', 'Browse directories and view file contents'],
  ['files.edit', 'Create & Edit Files', 'Upload, save edits, create files/folders'],
  ['files.delete', 'Delete Files', 'Delete, rename, move, and extract archives'],
  ['sftp.access', 'SFTP Access', 'Connect with an SFTP client using your account credentials'],
  ['settings.view', 'View Settings', 'Inspect server settings and properties'],
  ['settings.edit', 'Edit Settings', 'Modify MOTD, difficulty, rules & configs'],
  ['database.manage', 'Databases', 'Create, delete, and view database credentials'],
  ['backup.manage', 'Backups', 'Create, restore, download, and delete backups'],
  ['subusers.manage', 'Manage Team', 'Invite, modify, and revoke subuser access'],
  ['schedules.manage', 'Schedules', 'Create, toggle, and trigger cron schedules']
];

async function loadServerSubusers(srvId = null, skipRender = false) {
  if (srvId && curServer()?.id !== srvId) return;
  return loadTeamPage(skipRender);
}

function renderSubusersListHtml(srv) {
  const subusers = Array.isArray(S.serverSubusers) ? S.serverSubusers : [];
  if (!subusers.length) {
    return `<div class="card mu" style="text-align: center; padding: 24px 16px; background: rgba(255,255,255,0.01);">
      <div style="font-size: 26px; margin-bottom: 6px;">👥</div>
      <b style="color: var(--tx); display: block; margin-bottom: 4px;">No team members yet</b>
      <span style="font-size: 12.5px;">Add teammates and choose what they can access without sharing your password.</span>
    </div>`;
  }

  return `<div class="tw">
    <table>
      <thead>
        <tr>
          <th>User</th>
          <th>Email</th>
          <th>Granted Permissions</th>
          <th>Added</th>
          <th style="text-align: right;">Actions</th>
        </tr>
      </thead>
      <tbody>
        ${subusers.map(sub => {
          const permsCount = Array.isArray(sub.permissions) ? sub.permissions.length : 0;
          const permSummary = permsCount === SUBUSER_PERMS.length
            ? '<span class="badge online">All server permissions</span>'
            : `<span class="chip">${permsCount} of ${SUBUSER_PERMS.length} allowed</span>`;
          return `<tr>
            <td>
              <div class="row" style="gap: 8px; align-items: center;">
                <span class="avb" style="width: 28px; height: 28px; border-radius: 6px; overflow: hidden; display: inline-flex; align-items: center; justify-content: center; background: var(--bg);">${av(28)}</span>
                <b style="color: var(--tx);">${esc(sub.username)}</b>
              </div>
            </td>
            <td class="mu pv">${esc(sub.email || '—')}</td>
            <td>${permSummary}</td>
            <td class="mu" style="font-size: 12px;">${formatFileDate(sub.created_at)}</td>
            <td style="text-align: right;">
              <div class="row" style="justify-content: flex-end; gap: 6px;">
                <button type="button" class="btn sm" onclick="openSubuserModal('${esc(sub.id)}')">${ic('sliders', 12)} Edit</button>
                <button type="button" class="btn sm dg" onclick="deleteServerSubuser('${esc(sub.id)}')" aria-label="Revoke access for ${esc(sub.username)}">${ic('trash', 12)}</button>
              </div>
            </td>
          </tr>`;
        }).join('')}
      </tbody>
    </table>
  </div>`;
}

function openSubuserModal(subuserId = null) {
  const container = $('#modal-container');
  if (!container) return;
  const srv = curServer();
  if (!canManageTeam()) return;
  TeamUI.memberServer = { ...srv };
  const existing = subuserId && Array.isArray(S.serverSubusers)
    ? S.serverSubusers.find(s => String(s.id) === String(subuserId))
    : null;
  const isEdit = Boolean(existing);
  const curPerms = new Set(existing ? (existing.permissions || []) : ['console.view', 'files.view', 'settings.view']);

  container.innerHTML = `<div class="pal" onclick="if(event.target===this)closeModal()">
    <div class="card pbox" style="padding: 26px; max-width: 600px; max-height: 90vh; overflow-y: auto;">
      <div class="row sb" style="margin-bottom: 14px;">
        <h2 style="font-size: 20px; margin: 0; display: flex; align-items: center; gap: 8px;">
          ${ic('user', 18)} ${isEdit ? `Edit member: ${esc(existing.username)}` : 'Add existing user'}
        </h2>
        <button class="btn sm" onclick="closeModal()" aria-label="Close modal">&times;</button>
      </div>

      <p class="mu" style="font-size: 13px; line-height: 1.5; margin-bottom: 16px;">
        ${isEdit ? `Update server permissions for <b>${esc(existing.username)}</b>.` : `Add an existing NeoPanel account to <b>${esc(srv?.name || 'this server')}</b>. For someone without an account, use Invite by email on the Team page.`}
      </p>

      <form onsubmit="submitSubuserModal(event, '${isEdit ? esc(existing.id) : ''}')">
        ${!isEdit ? `
          <label class="f">
            <span>User Email or Username</span>
            <input type="text" id="sub-ident" required placeholder="user@example.com or username">
          </label>
        ` : ''}

        <div style="margin-bottom: 14px;">
          <div class="row sb" style="margin-bottom: 6px;">
            <span style="font-size: 13px; font-weight: 600; color: var(--tx);">Granted Permissions</span>
            <div class="row" style="gap: 6px;">
              <button type="button" class="btn sm" onclick="$$('.sub-perm-cb').forEach(cb=>cb.checked=true)">Select All</button>
              <button type="button" class="btn sm" onclick="$$('.sub-perm-cb').forEach(cb=>cb.checked=false)">Deselect All</button>
            </div>
          </div>
          <div class="subuser-perm-grid">
            ${SUBUSER_PERMS.map(([key, label, desc]) => `
              <label class="subuser-perm-item">
                <input type="checkbox" class="sub-perm-cb" value="${esc(key)}" ${curPerms.has(key) ? 'checked' : ''}>
                <div style="min-width: 0;">
                  <b style="font-size: 12.5px; color: var(--tx); display: block;">${esc(label)}</b>
                  <span class="mu" style="font-size: 11px; display: block; line-height: 1.3;">${esc(desc)}</span>
                </div>
              </label>
            `).join('')}
          </div>
        </div>

        <div class="row sb" style="gap: 10px; margin-top: 18px;">
          <button type="button" class="btn" onclick="closeModal()">Cancel</button>
          <button type="submit" id="sub-submit-btn" class="btn pr">
            ${ic('check', 14)} ${isEdit ? 'Save Permissions' : 'Grant Access'}
          </button>
        </div>
      </form>
    </div>
  </div>`;
  setTimeout(() => $('#sub-ident')?.focus(), 40);
}

async function submitSubuserModal(e, subuserId = '') {
  e.preventDefault();
  const srv = TeamUI.memberServer;
  if (!srv || !sameQolServer(srv, curServer())) return;
  const btn = $('#sub-submit-btn');
  if (btn) setBtnLoading(btn, true, subuserId ? 'Saving...' : 'Granting...');

  const perms = Array.from($$('.sub-perm-cb:checked')).map(cb => cb.value);
  const identifier = $('#sub-ident')?.value.trim() || '';

  try {
    const isEdit = Boolean(subuserId);
    if (isEdit) {
      await api(srvApiPath(srv, `/subusers/${encodeURIComponent(subuserId)}`), {
        method: 'PUT',
        body: JSON.stringify({ permissions: perms })
      });
      toast('Member permissions updated', 'ok');
    } else {
      await api(srvApiPath(srv, '/subusers'), {
        method: 'POST',
        body: JSON.stringify({ username: identifier, permissions: perms })
      });
      toast(`User "${identifier}" added to the team`, 'ok');
    }
    closeModal();
    await loadServerSubusers(srv.id);
  } catch (err) {
    toast(err.message || 'Failed to save subuser', 'bad');
    if (btn) setBtnLoading(btn, false);
  }
}

async function deleteServerSubuser(subuserId) {
  const srv = curServer();
  if (!srv) return;
  const ok = await uiConfirm({
    title: 'Remove team access',
    message: 'Are you sure you want to revoke this user\'s access to the server?',
    confirmText: 'Revoke Access',
    isDanger: true,
    icon: 'trash'
  });
  if (!ok) return;

  try {
    await api(srvApiPath(srv, `/subusers/${encodeURIComponent(subuserId)}`), { method: 'DELETE' });
    toast('Team access removed', 'ok');
    await loadServerSubusers(srv.id);
  } catch (err) {
    toast(err.message || 'Failed to revoke subuser', 'bad');
  }
}

// --- Server Icon Management ---
function resizeImageTo64x64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () => reject(new Error('Failed to read image file'));
    reader.onload = (e) => {
      const img = new Image();
      img.onerror = () => reject(new Error('Invalid image file format'));
      img.onload = () => {
        try {
          const canvas = document.createElement('canvas');
          canvas.width = 64;
          canvas.height = 64;
          const ctx = canvas.getContext('2d');
          if (!ctx) return resolve(file);

          // Center-crop square to prevent distortion
          const minDim = Math.min(img.width, img.height);
          const sx = (img.width - minDim) / 2;
          const sy = (img.height - minDim) / 2;

          ctx.imageSmoothingEnabled = true;
          ctx.imageSmoothingQuality = 'high';
          ctx.drawImage(img, sx, sy, minDim, minDim, 0, 0, 64, 64);

          canvas.toBlob((blob) => {
            if (!blob) return resolve(file);
            resolve(blob);
          }, 'image/png');
        } catch (err) {
          resolve(file);
        }
      };
      img.src = e.target.result;
    };
    reader.readAsDataURL(file);
  });
}

async function uploadServerIcon(srvId, file) {
  if (!file) return;
  const srv = srvId ? S.servers.find(s => s.id === srvId) : curServer();
  if (!srv) return;

  if (file.size > 10 * 1024 * 1024) {
    toast('Image file too large (maximum 10 MB)', 'bad');
    return;
  }

  try {
    toast('Resizing & formatting icon to 64×64 PNG...');
    const processedBlob = await resizeImageTo64x64(file);
    const formData = new FormData();
    formData.append('icon', processedBlob, 'server-icon.png');

    const res = await api(srvApiPath(srv, '/icon'), {
      method: 'POST',
      body: formData
    });
    if (res && res.icon_url) {
      srv.icon_url = res.icon_url;
      const img = $('#set-srv-icon-img');
      if (img) img.src = res.icon_url + '?t=' + Date.now();
      toast('Server icon updated to 64×64 PNG (server-icon.png)', 'ok');
      R();
    }
  } catch (err) {
    toast(err.message || 'Failed to upload icon', 'bad');
  }
}

async function deleteServerIcon(srvId) {
  const srv = srvId ? S.servers.find(s => s.id === srvId) : curServer();
  if (!srv) return;
  try {
    await api(srvApiPath(srv, '/icon'), { method: 'DELETE' });
    srv.icon_url = '';
    const img = $('#set-srv-icon-img');
    if (img) img.src = getBrand('default_server_icon_url') || '/static/img/default-server-icon.png';
    toast('Server icon removed', 'ok');
    R();
  } catch (err) {
    toast(err.message || 'Failed to delete icon', 'bad');
  }
}

// --- Quick Config Editor Tool ---
let _quickConfigsCache = null;

async function openQuickConfigsModal(srvId) {
  const container = $('#modal-container');
  if (!container) return;
  const srv = srvId ? S.servers.find(s => s.id === srvId) : curServer();
  if (!srv) return;
  if (isSteelEngine(srv.type)) { uiAlert({ title: 'Steel Configuration', message: steelConfigNotice() }); return; }

  container.innerHTML = `<div class="pal" onclick="if(event.target===this)closeModal()">
    <div class="card pbox" style="padding: 26px; max-width: 640px; max-height: 90vh; overflow-y: auto;">
      <div class="row sb" style="margin-bottom: 14px;">
        <h2 style="font-size: 20px; margin: 0; display: flex; align-items: center; gap: 8px;">
          ${ic('sliders', 18)} Quick Configs &amp; Optimization
        </h2>
        <button class="btn sm" onclick="closeModal()" aria-label="Close modal">&times;</button>
      </div>
      <div id="quick-cfg-body">${loaderHtml('Loading server configurations...')}</div>
    </div>
  </div>`;

  try {
    const data = await api(srvApiPath(srv, '/quick-configs'));
    _quickConfigsCache = data || {};
    renderQuickConfigsModalBody(srv);
  } catch (err) {
    const body = $('#quick-cfg-body');
    if (body) body.innerHTML = `<div class="node-error-note">${esc(err.message || 'Failed to load configs')}</div>`;
  }
}

function renderQuickConfigsModalBody(srv) {
  const body = $('#quick-cfg-body');
  if (!body || !_quickConfigsCache) return;
  const cfg = _quickConfigsCache;

  body.innerHTML = `
    <p class="mu" style="font-size: 13px; line-height: 1.5; margin-bottom: 16px;">
      Directly adjust high-impact Minecraft game properties without manually parsing raw <code>server.properties</code> and <code>eula.txt</code>.
    </p>

    <!-- Quick Presets -->
    <div style="background: rgba(255,255,255,0.02); border: 1px solid var(--border); border-radius: 8px; padding: 12px; margin-bottom: 16px;">
      <div style="font-size: 12px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: var(--tx-muted); margin-bottom: 8px;">Performance Optimization Presets</div>
      <div class="row" style="gap: 8px; flex-wrap: wrap;">
        <button type="button" class="btn sm" onclick="applyQuickConfigPreset('perf')">⚡ Extreme Performance (6 View, 4 Sim)</button>
        <button type="button" class="btn sm" onclick="applyQuickConfigPreset('balanced')">⚖️ Balanced (10 View, 8 Sim)</button>
        <button type="button" class="btn sm" onclick="applyQuickConfigPreset('ultra')">🌟 Ultra Visual (24 View, 16 Sim)</button>
      </div>
    </div>

    <form onsubmit="saveQuickConfigs(event, '${esc(srv.id)}')">
      <div class="grid g2" style="margin-bottom: 14px;">
        <label class="f" style="margin-bottom: 0;">
          <span>View Distance (<b id="lbl-qc-view">${cfg.view_distance ?? 10}</b> chunks)</span>
          <div class="quick-cfg-slider-row">
            <input type="range" id="qc-view" min="2" max="32" value="${cfg.view_distance ?? 10}" oninput="$('#lbl-qc-view').textContent=this.value">
            <span class="quick-cfg-val-badge">${cfg.view_distance ?? 10}</span>
          </div>
        </label>
        <label class="f" style="margin-bottom: 0;">
          <span>Simulation Distance (<b id="lbl-qc-sim">${cfg.simulation_distance ?? 10}</b> chunks)</span>
          <div class="quick-cfg-slider-row">
            <input type="range" id="qc-sim" min="2" max="32" value="${cfg.simulation_distance ?? 10}" oninput="$('#lbl-qc-sim').textContent=this.value">
            <span class="quick-cfg-val-badge">${cfg.simulation_distance ?? 10}</span>
          </div>
        </label>
      </div>

      <div class="grid g2" style="margin-bottom: 14px;">
        <label class="f" style="margin-bottom: 0;">
          <span>Spawn Protection Radius (<b id="lbl-qc-spawn">${cfg.spawn_protection ?? 16}</b> blocks)</span>
          <div class="quick-cfg-slider-row">
            <input type="range" id="qc-spawn" min="0" max="64" value="${cfg.spawn_protection ?? 16}" oninput="$('#lbl-qc-spawn').textContent=this.value">
            <span class="quick-cfg-val-badge">${cfg.spawn_protection ?? 16}</span>
          </div>
        </label>
        <label class="f" style="margin-bottom: 0;">
          <span>Server MOTD</span>
          <input type="text" id="qc-motd" value="${esc(cfg.motd || '')}" placeholder="A Minecraft Server">
        </label>
      </div>

      <div style="background: rgba(255,255,255,0.02); border: 1px solid var(--border); border-radius: 8px; padding: 12px; margin-bottom: 16px;">
        <div style="font-size: 12px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: var(--tx-muted); margin-bottom: 8px;">Gameplay &amp; Protection Toggles</div>
        <div class="grid g2" style="gap: 10px;">
          <label class="row sb" style="cursor: pointer;">
            <span style="font-size: 13px; color: var(--tx);">Allow Player Flight</span>
            <span class="sw"><input type="checkbox" id="qc-flight" ${cfg.allow_flight ? 'checked' : ''}><i></i></span>
          </label>
          <label class="row sb" style="cursor: pointer;">
            <span style="font-size: 13px; color: var(--tx);">Player vs Player (PvP)</span>
            <span class="sw"><input type="checkbox" id="qc-pvp" ${cfg.pvp ? 'checked' : ''}><i></i></span>
          </label>
          <label class="row sb" style="cursor: pointer;">
            <span style="font-size: 13px; color: var(--tx);">Enforce Whitelist</span>
            <span class="sw"><input type="checkbox" id="qc-whitelist" ${cfg.white_list ? 'checked' : ''}><i></i></span>
          </label>
          <label class="row sb" style="cursor: pointer;">
            <span style="font-size: 13px; color: var(--tx);">Command Blocks</span>
            <span class="sw"><input type="checkbox" id="qc-cmd" ${cfg.enable_command_block ? 'checked' : ''}><i></i></span>
          </label>
          <label class="row sb" style="cursor: pointer;">
            <span style="font-size: 13px; color: var(--tx);">Mojang EULA Accepted (eula.txt)</span>
            <span class="sw"><input type="checkbox" id="qc-eula" ${cfg.eula_accepted ? 'checked' : ''}><i></i></span>
          </label>
        </div>
      </div>

      <div class="row sb" style="gap: 10px;">
        <button type="button" class="btn" onclick="closeModal()">Cancel</button>
        <button type="submit" id="qc-save-btn" class="btn pr">
          ${ic('check', 14)} Apply &amp; Save Configs
        </button>
      </div>
    </form>
  `;
}

function applyQuickConfigPreset(preset) {
  if (preset === 'perf') {
    if ($('#qc-view')) { $('#qc-view').value = 6; $('#lbl-qc-view').textContent = 6; }
    if ($('#qc-sim')) { $('#qc-sim').value = 4; $('#lbl-qc-sim').textContent = 4; }
    if ($('#qc-spawn')) { $('#qc-spawn').value = 0; $('#lbl-qc-spawn').textContent = 0; }
    toast('Applied Extreme Performance preset');
  } else if (preset === 'balanced') {
    if ($('#qc-view')) { $('#qc-view').value = 10; $('#lbl-qc-view').textContent = 10; }
    if ($('#qc-sim')) { $('#qc-sim').value = 8; $('#lbl-qc-sim').textContent = 8; }
    if ($('#qc-spawn')) { $('#qc-spawn').value = 16; $('#lbl-qc-spawn').textContent = 16; }
    toast('Applied Balanced preset');
  } else if (preset === 'ultra') {
    if ($('#qc-view')) { $('#qc-view').value = 24; $('#lbl-qc-view').textContent = 24; }
    if ($('#qc-sim')) { $('#qc-sim').value = 16; $('#lbl-qc-sim').textContent = 16; }
    if ($('#qc-spawn')) { $('#qc-spawn').value = 16; $('#lbl-qc-spawn').textContent = 16; }
    toast('Applied Ultra Visual preset');
  }
}

async function saveQuickConfigs(e, srvId) {
  e.preventDefault();
  const srv = srvId ? S.servers.find(s => s.id === srvId) : curServer();
  if (!srv) return;
  const btn = $('#qc-save-btn');
  if (btn) setBtnLoading(btn, true, 'Saving...');

  const payload = {
    view_distance: parseInt($('#qc-view')?.value, 10) || 10,
    simulation_distance: parseInt($('#qc-sim')?.value, 10) || 10,
    spawn_protection: parseInt($('#qc-spawn')?.value, 10) || 16,
    motd: $('#qc-motd')?.value || '',
    allow_flight: $('#qc-flight') ? $('#qc-flight').checked : false,
    pvp: $('#qc-pvp') ? $('#qc-pvp').checked : true,
    white_list: $('#qc-whitelist') ? $('#qc-whitelist').checked : false,
    enable_command_block: $('#qc-cmd') ? $('#qc-cmd').checked : false,
    eula_accepted: $('#qc-eula') ? $('#qc-eula').checked : true
  };

  try {
    await api(srvApiPath(srv, '/quick-configs'), {
      method: 'POST',
      body: JSON.stringify(payload)
    });
    closeModal();
    toast('Quick configurations saved (server restart recommended)', 'ok');
    if (S.view === 'set') loadProperties();
  } catch (err) {
    toast(err.message || 'Failed to save quick configs', 'bad');
    if (btn) setBtnLoading(btn, false);
  }
}

// --- Client Area Provisioning Integration Helpers ---
function setClientCodeLang(lang) {
  S.clientCodeLang = lang;
  if (S.view === 'admin' && S.adminTab === 4) R();
}

function getClientProvisioningCode(lang = 'php') {
  const panelUrl = window.location.origin;
  if (lang === 'php') {
    return `<?php
// NeoPanel Automated Provisioning (WHMCS / Blesta / Custom Hook)
$endpoint = "${panelUrl}/api/client-api/v1/provision";
$apiKey = "YOUR_MASTER_NODE_TOKEN_OR_API_KEY";

$payload = [
    "client_email"      => $client["email"],      // e.g. "player@example.com"
    "client_username"   => $client["username"],   // e.g. "notch"
    "server_name"       => "Survival SMP",
    "node_id"           => "local",               // Target node ID
    "server_type"       => "paper",               // paper, purpur, fabric, etc.
    "minecraft_version" => "1.21.4",
    "memory_mb"         => 4096,                  // 4 GB RAM
    "disk_mb"           => 10240,                 // 10 GB Disk
    "cpu_limit_percent" => 200,                   // 2 CPU cores max
    "auto_start"        => true
];

$ch = curl_init($endpoint);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_POST, true);
curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payload));
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    "Content-Type: application/json",
    "X-API-Key: " . $apiKey
]);

$response = curl_exec($ch);
$httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
curl_close($ch);

$result = json_decode($response, true);
if ($httpCode === 201) {
    // Created! Save $result["server"]["id"], $result["server"]["port"]
    echo "Server created! Port: " . $result["server"]["port"];
}
`;
  } else if (lang === 'node') {
    return `// Node.js Provisioning Hook
const fetch = require('node-fetch');

async function provisionServer(clientOrder) {
  const res = await fetch('${panelUrl}/api/client-api/v1/provision', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': process.env.NEOPANEL_API_KEY
    },
    body: JSON.stringify({
      client_email: clientOrder.email,
      client_username: clientOrder.username,
      server_name: clientOrder.serverName || 'Minecraft Server',
      node_id: clientOrder.nodeId || 'local',
      server_type: 'paper',
      minecraft_version: '1.21.4',
      memory_mb: clientOrder.ramMB || 4096,
      disk_mb: 10240,
      auto_start: true
    })
  });
  return await res.json();
}`;
  } else if (lang === 'python') {
    return `# Python Provisioning Hook
import requests

def provision_minecraft_server(email, username, ram_mb=4096):
    url = "${panelUrl}/api/client-api/v1/provision"
    headers = {
        "Content-Type": "application/json",
        "X-API-Key": "YOUR_MASTER_NODE_TOKEN_OR_API_KEY"
    }
    data = {
        "client_email": email,
        "client_username": username,
        "server_name": "Survival Server",
        "node_id": "local",
        "server_type": "paper",
        "minecraft_version": "1.21.4",
        "memory_mb": ram_mb,
        "auto_start": True
    }
    resp = requests.post(url, json=data, headers=headers)
    return resp.json()`;
  }
  return `curl -X POST "${panelUrl}/api/client-api/v1/provision" \\
  -H "Content-Type: application/json" \\
  -H "X-API-Key: YOUR_MASTER_NODE_TOKEN_OR_API_KEY" \\
  -d '{
    "client_email": "user@example.com",
    "client_username": "player1",
    "server_name": "SMP Server",
    "node_id": "local",
    "server_type": "paper",
    "minecraft_version": "1.21.4",
    "memory_mb": 4096,
    "disk_mb": 10240,
    "auto_start": true
  }'`;
}

function copyClientProvisioningCode() {
  const code = getClientProvisioningCode(S.clientCodeLang || 'php');
  navigator.clipboard.writeText(code).then(() => {
    toast('Provisioning code copied to clipboard!', 'ok');
  }).catch(() => {
    toast('Failed to copy code', 'bad');
  });
}

async function setUserTimezone(tz) {
  if (!S.branding) S.branding = { ...DEFAULT_BRANDING };
  S.branding.timezone = tz;
  localStorage.setItem('neopanel_timezone', tz);
  toast(`Timezone set to ${tz}`, 'ok');
  if (S.user?.role === 'admin') {
    try {
      const draft = { ...DEFAULT_BRANDING, ...S.branding, timezone: tz };
      await api('/api/admin/branding', {
        method: 'PUT',
        suppressToast: true,
        body: JSON.stringify({ branding: draft })
      });
    } catch (e) {}
  }
  R();
}

// --- Server Deletion Management ---
function openDeleteServerModal(srvId, srvName, nodeId = null) {
  const container = $('#modal-container');
  if (!container) return;

  const targetSrv = findServerRef(srvId, nodeId);
  const targetNodeId = nodeId || targetSrv?.node_id || 'local';
  const resolvedName = srvName || targetSrv?.name || srvId;

  container.innerHTML = `<div class="pal" onclick="if(event.target===this)closeModal()">
    <div class="card pbox" style="padding:26px; max-width:520px; border:1px solid rgba(239, 68, 68, 0.4);">
      <div class="row sb" style="margin-bottom:14px;">
        <h2 style="font-size:20px; margin:0; color:#ef4444; display:flex; align-items:center; gap:8px;">
          ${ic('trash', 18)} Delete Server Instance
        </h2>
        <button class="btn sm" onclick="closeModal()" aria-label="Close modal">&times;</button>
      </div>

      <p style="font-size:13.5px; color:#e2e8f0; line-height:1.6; margin-bottom:14px;">
        You are about to permanently delete <b style="color:#fff;">${esc(resolvedName)}</b> (<code>${esc(srvId)}</code>)${targetNodeId !== 'local' ? ` on node <code>${esc(targetNodeId)}</code>` : ''}.
      </p>

      <div style="background:rgba(239, 68, 68, 0.1); border:1px solid rgba(239, 68, 68, 0.3); border-radius:8px; padding:12px 14px; margin-bottom:16px; font-size:13px; color:#fca5a5; line-height:1.5;">
        <b>Warning:</b> This process is completely irreversible. All world data, player inventories, plugin configs, and assigned port allocations will be wiped immediately.
      </div>

      <label class="f" style="margin-bottom:20px;">
        <span style="font-size:13px; color:#cbd5e1;">To confirm, type <b style="color:#fff;">${esc(resolvedName)}</b> below:</span>
        <input type="text" id="del-srv-confirm" placeholder="Type server name to confirm" autocomplete="off" data-expected="${esc(resolvedName)}" oninput="checkDeleteServerConfirm(this.dataset.expected)">
      </label>

      <div class="row" style="justify-content:flex-end; gap:10px;">
        <button type="button" class="btn" onclick="closeModal()">Cancel</button>
        <button type="button" id="del-srv-btn" class="btn dg" disabled data-srv-id="${esc(srvId)}" data-node-id="${esc(targetNodeId)}" onclick="executeDeleteServer(this.dataset.srvId, this.dataset.nodeId)">
          ${ic('trash', 14)} Delete Server Permanently
        </button>
      </div>
    </div>
  </div>`;

  setTimeout(() => $('#del-srv-confirm')?.focus(), 50);
}

function checkDeleteServerConfirm(expectedName) {
  const inp = $('#del-srv-confirm');
  const btn = $('#del-srv-btn');
  if (!inp || !btn) return;
  const match = inp.value.trim().toLowerCase() === expectedName.trim().toLowerCase();
  btn.disabled = !match;
}

async function executeDeleteServer(srvId, nodeId = 'local') {
  const targetNodeId = nodeId || 'local';
  const btn = $('#del-srv-btn');
  if (btn) {
    btn.disabled = true;
    btn.textContent = 'Deleting server...';
  }

  try {
    const result = await api(nodeApiPath(targetNodeId, `/api/servers/${encodeURIComponent(srvId)}`), { method: 'DELETE' });
    if (result?.deleted !== true || result?.server_id !== srvId) {
      throw new Error('Deletion completion was not confirmed by the hosting node. Refresh the server list and verify the node is updated before retrying.');
    }
    if (S.activeServerId === srvId && (S.activeNodeId || 'local') === targetNodeId) {
      clearTimeout(S.wsReconnectTimer);
      S.wsReconnectTimer = null;
      if (S.wsPingTimer) {
        clearInterval(S.wsPingTimer);
        S.wsPingTimer = null;
      }
      S.activeServerId = null;
      S.activeNodeId = null;
      if (S.activeWs) {
        try {
          S.activeWs.onopen = null;
          S.activeWs.onmessage = null;
          S.activeWs.onerror = null;
          S.activeWs.onclose = null;
          S.activeWs.close();
        } catch (e) {}
        S.activeWs = null;
      }
    }

    const saved = getSavedActiveServerRef();
    if (saved && saved.id === srvId && (saved.nodeId || 'local') === targetNodeId) {
      try { localStorage.removeItem('neopanel_active_srv'); } catch (e) {}
    }
    toast('Server deleted successfully', 'ok');
    closeModal();
    await loadDashboard(true);
    applyRoute('/');
  } catch (err) {
    toast(err.message || 'Failed to delete server', 'bad');
    if (btn) {
      btn.disabled = false;
      btn.textContent = 'Delete Server Permanently';
    }
  }
}

// --- Server Reinstallation (mcjarfiles.com) ---
function updateReinstallJavaPreview() {
  const sw = $('#reinst-software')?.value || 'paper';
  const ver = $('#reinst-version')?.value.trim() || '1.21.11';
  const el = $('#reinst-java-rec');
  if (el) el.textContent = isSteelEngine(sw) ? 'Not required (native)' : `Java ${getRecommendedJavaForMcVersion(ver, sw)}`;
}

function syncReinstallSelectFromInput(ver) {
  const sel = $('#reinst-ver-sel');
  if (!sel) return;
  const v = String(ver || '').trim();
  if (v && Array.from(sel.options).some(o => o.value === v)) {
    sel.value = v;
    if (typeof sel._cselectRebuild === 'function') sel._cselectRebuild();
  }
}

function setReinstallVersion(ver) {
  const inp = $('#reinst-version');
  if (inp) inp.value = ver;
  syncReinstallSelectFromInput(ver);
  updateReinstallJavaPreview();
}

function onReinstallEngineChange(sw) {
  const iconEl = $('#reinst-eng-icon');
  if (iconEl) iconEl.innerHTML = engineIcon(sw, 18, true);
  const sel = $('#reinst-ver-sel');
  const inp = $('#reinst-version');
  const curV = inp?.value?.trim() || '1.21.11';
  const native = isSteelEngine(sw);
  const previousType = $('#reinstall-form')?.dataset.selectedType;
  if (native && previousType && !isSteelEngine(previousType) && inp) inp.value = 'latest';
  else if (!native && isSteelEngine(previousType) && inp) inp.value = '1.21.11';
  if ($('#reinstall-form')) $('#reinstall-form').dataset.selectedType = sw;
  if (inp) inp.placeholder = native ? 'latest or official Steel release tag' : 'Minecraft version or latest';
  const sync = $('#reinst-sync-java');
  if (sync) { sync.disabled = native; sync.closest('label').parentElement.style.display = native ? 'none' : ''; }
  const plugins = $('#reinst-plugins');
  if (plugins) { plugins.disabled = native; plugins.closest('label').style.display = native ? 'none' : ''; }
  const note = $('#reinst-steel-note');
  if (note) { note.style.display = native ? '' : 'none'; note.innerHTML = native ? steelConfigNotice(false) : ''; }
  if (native) loadSteelVersions();
  if (sw === 'velocity' && !curV.includes('SNAPSHOT')) {
    if (inp) inp.value = '3.4.0-SNAPSHOT';
  } else if (sw === 'waterfall' && curV !== '1.20') {
    if (inp) inp.value = '1.20';
  } else if (sw === 'bungeecord') {
    if (inp) inp.value = 'latest';
  }
  if (sel) {
    const newCur = inp?.value?.trim() || '1.21.11';
    const vers = getAllMcVersions(sw);
    const list = (newCur && !vers.includes(newCur)) ? [newCur, ...vers] : vers;
    sel.innerHTML = list.map(v => `<option value="${esc(v)}" ${v === newCur ? 'selected' : ''}>${esc(v)}</option>`).join('');
    if (typeof sel._cselectRebuild === 'function') sel._cselectRebuild();
  }
  updateReinstallJavaPreview();

  // Show wipe recommendation when switching to an incompatible engine family
  const form = $('#reinstall-form');
  const origType = (form?.dataset?.origType || '').toLowerCase();
  const wipeWarn = $('#reinst-wipe-warn');
  if (!wipeWarn || !origType) { updateReinstallJavaPreview(); return; }

  // Define incompatible engine families: switching between families requires a wipe
  const incompatible = incompatibleEngineChange(origType, sw);
  wipeWarn.style.display = incompatible ? '' : 'none';
  wipeWarn.innerHTML = incompatible ? `
    <div style="background:rgba(239,68,68,0.12);border:1px solid rgba(239,68,68,0.45);border-radius:8px;padding:12px 14px;margin-bottom:14px;display:flex;gap:10px;align-items:flex-start;">
      <span style="font-size:18px;line-height:1;">⚠️</span>
      <div>
        <b style="color:#f87171;">Incompatible engine change detected</b>
        <p style="margin:4px 0 0;font-size:12.5px;color:var(--text-muted);line-height:1.5;">
          You are switching from <b>${esc(origType)}</b> to <b>${esc(sw)}</b>. Worlds/data, plugins/mods and configuration may be incompatible and are not converted. Steel uses <code>saves/</code> and <code>config/config.toml</code>; Java plugins/mods cannot run on Steel.<br>
          <b style="color:#fbbf24;">Back up first. Recommended: wipe all server files before reinstalling; do not preserve incompatible data.</b>
        </p>
        <label style="display:flex;align-items:center;gap:6px;margin-top:8px;font-size:12.5px;cursor:pointer;">
          <input type="checkbox" id="reinst-wipe-files" style="accent-color:#ef4444;width:14px;height:14px;">
          <span>Wipe all server files before reinstalling <span style="color:var(--text-muted)">(keeps world &amp; config if checked above)</span></span>
        </label>
      </div>
    </div>` : '';
}

function openReinstallModal(srvId, nodeId = null) {
  const container = $('#modal-container');
  if (!container) return;

  const targetSrv = findServerRef(srvId, nodeId) || curServer();
  if (!targetSrv) return;

  const targetId = targetSrv.id;
  const targetNodeId = nodeId || targetSrv.node_id || 'local';
  const targetName = targetSrv.name || 'Server';
  const curType = (targetSrv.type || 'paper').toLowerCase();
  const curVer = isSteelEngine(curType) ? 'latest' : (targetSrv.version || '1.21.11');
  const recJava = getRecommendedJavaForMcVersion(curVer, curType);
  const allVers = getAllMcVersions(curType);
  const verList = (curVer && !allVers.includes(curVer)) ? [curVer, ...allVers] : allVers;

  container.innerHTML = `<div class="pal" onclick="if(event.target===this)closeModal()">
    <div class="card pbox" style="padding:26px; max-width:580px; max-height:88vh; overflow-y:auto; border:1px solid var(--border-light);">
      <div class="row sb" style="margin-bottom:14px;">
        <h2 style="font-size:20px; margin:0; color:#fff; display:flex; align-items:center; gap:8px;">
          ${ic('refresh', 18)} Reinstall Server: ${esc(targetName)}
        </h2>
        <button class="btn sm" onclick="closeModal()" aria-label="Close modal">&times;</button>
      </div>

      <p class="mu" style="font-size:13px; line-height:1.5; margin-bottom:16px;">
        Reinstalls the server executable from official upstream releases (Steel) or mcjarfiles.com (Java engines).
      </p>

      <form id="reinstall-form" data-srv-id="${esc(targetId)}" data-node-id="${esc(targetNodeId)}" data-orig-type="${esc(curType)}" data-selected-type="${esc(curType)}" onsubmit="submitReinstall(event, this.dataset.srvId, this.dataset.nodeId)">
        <div id="reinst-wipe-warn" style="display:none;"></div>
        <div id="reinst-steel-note" style="display:none; margin-bottom:14px;"></div>
        <label class="f" style="margin-bottom:12px;">
          <span style="display:flex; align-items:center; gap:8px;">
            <span id="reinst-eng-icon">${engineIcon(curType, 18, true)}</span>
            <span>Minecraft Engine / Server Software</span>
          </span>
          <select id="reinst-software" style="height:38px;" onchange="onReinstallEngineChange(this.value)">
            <option value="paper" ${curType === 'paper' ? 'selected' : ''}>Paper (High performance, Bukkit/Spigot plugins)</option>
            <option value="purpur" ${curType === 'purpur' ? 'selected' : ''}>Purpur (Ultra-configurable Paper fork)</option>
            <option value="leaf" ${curType === 'leaf' || curType === 'leafmc' ? 'selected' : ''}>LeafMC (Hyper-optimized Paper fork)</option>
            <option value="folia" ${curType === 'folia' ? 'selected' : ''}>Folia (Multi-threaded chunk engine)</option>
            <option value="fabric" ${curType === 'fabric' ? 'selected' : ''}>Fabric (Lightweight mod loader)</option>
            <option value="forge" ${curType === 'forge' ? 'selected' : ''}>Forge (Classic modding framework)</option>
            <option value="neoforge" ${curType === 'neoforge' ? 'selected' : ''}>NeoForge (Modern community mod loader)</option>
            <option value="vanilla" ${curType === 'vanilla' ? 'selected' : ''}>Vanilla (Official Mojang release)</option>
            <option value="velocity" ${curType === 'velocity' ? 'selected' : ''}>Velocity (High-performance modern proxy)</option>
            <option value="waterfall" ${curType === 'waterfall' ? 'selected' : ''}>Waterfall (PaperMC BungeeCord proxy fork)</option>
            <option value="bungeecord" ${curType === 'bungeecord' ? 'selected' : ''}>BungeeCord (Classic multi-server proxy)</option>
            <option value="steel" ${curType === 'steel' ? 'selected' : ''}>Steel (Experimental, written in Rust)</option>
          </select>
        </label>

        <div class="f" style="margin-bottom:16px;">
          <span style="display:block; margin-bottom:6px; font-size:13px; color:var(--tx-muted); font-weight:500;">Version</span>
          <div style="display:grid; grid-template-columns: minmax(145px, 180px) 1fr; gap:8px;">
            <select id="reinst-ver-sel" data-searchable="1" onchange="setReinstallVersion(this.value)">
              ${verList.map(v => `<option value="${esc(v)}" ${v === curVer ? 'selected' : ''}>${esc(v)}</option>`).join('')}
            </select>
            <input type="text" id="reinst-version" value="${esc(curVer)}" placeholder="1.21.11, 1.20.4, 1.12.2, or latest" required oninput="syncReinstallSelectFromInput(this.value); updateReinstallJavaPreview();">
          </div>
        </div>

        <div style="background:rgba(255,255,255,0.03); border:1px solid var(--border); border-radius:8px; padding:12px 14px; margin-bottom:14px;">
          <b style="font-size:13px; color:#fff; display:block; margin-bottom:8px;">Preservation Options</b>
          <label class="row" style="cursor:pointer; margin-bottom:8px; font-size:13px; gap:10px;">
            <input type="checkbox" id="reinst-world" checked>
            <span>Preserve Worlds (Java: <code>world*</code>; Steel: <code>saves/</code>; no format conversion)</span>
          </label>
          <label class="row" style="cursor:pointer; margin-bottom:8px; font-size:13px; gap:10px;">
            <input type="checkbox" id="reinst-config" checked>
            <span>Preserve Configurations (Java properties/YAML; Steel: <code>config/config.toml</code>; no format conversion)</span>
          </label>
          <label class="row" style="cursor:pointer; font-size:13px; gap:10px;">
            <input type="checkbox" id="reinst-plugins" checked>
            <span>Preserve Plugins &amp; Mods (keep <code>plugins</code>, <code>mods</code> folders)</span>
          </label>
        </div>

        <div style="background: rgba(16, 185, 129, 0.08); border: 1px solid rgba(16, 185, 129, 0.28); border-radius: 8px; padding: 11px 14px; margin-bottom: 20px;">
          <label class="row" style="cursor:pointer; font-size:13px; gap:10px; align-items:flex-start;">
            <input type="checkbox" id="reinst-sync-java" checked style="margin-top:2px;">
            <div>
              <b style="color: var(--tx); display: block;">Sync Java Runtime (<span id="reinst-java-rec">Java ${esc(recJava)}</span>)</b>
              <span class="mu" style="font-size:12px;">Automatically switches server Java Runtime among Java 25, 21, 17, or 8 for the chosen release.</span>
            </div>
          </label>
        </div>

        <div class="row" style="justify-content:flex-end; gap:10px;">
          <button type="button" class="btn" onclick="closeModal()">Cancel</button>
          <button type="submit" id="reinst-submit-btn" class="btn pr">
            ${ic('refresh', 14)} Start Reinstallation
          </button>
        </div>
      </form>
    </div>
  </div>`;

  enhanceSelects(container);
  onReinstallEngineChange(curType);
  if (!isSteelEngine(curType)) loadVersionsCatalog();
}

async function submitReinstall(e, srvId, nodeId = 'local') {
  e.preventDefault();
  const targetNodeId = nodeId || 'local';
  const btn = $('#reinst-submit-btn');
  if (btn) {
    btn.disabled = true;
    btn.innerHTML = `${ic('refresh', 14)} Reinstalling...`;
  }

  const software = $('#reinst-software')?.value || 'paper';
  const version = $('#reinst-version')?.value.trim() || 'latest';
  const preserveWorld = $('#reinst-world')?.checked ?? true;
  const preserveConfig = $('#reinst-config')?.checked ?? true;
  const preservePlugins = !isSteelEngine(software) && ($('#reinst-plugins')?.checked ?? true);
  const syncJava = !isSteelEngine(software) && ($('#reinst-sync-java')?.checked ?? false);
  const wipeFiles = $('#reinst-wipe-files')?.checked ?? false;

  toast(`Initiating reinstallation of ${software} ${version}...`);

  try {
    const res = await api(nodeApiPath(targetNodeId, `/api/servers/${encodeURIComponent(srvId)}/reinstall`), {
      method: 'POST',
      body: JSON.stringify({
        software,
        version,
        preserve_world: preserveWorld,
        preserve_config: preserveConfig,
        preserve_plugins: preservePlugins,
        wipe_files: wipeFiles
      })
    });

    const targetSrv = findServerRef(srvId, targetNodeId) || curServer();
    if (targetSrv) {
      targetSrv.type = software;
      targetSrv.version = res?.version || version;
      if (res?.jar_file) targetSrv.jar_file = res.jar_file;
    }

    if (syncJava && targetSrv) {
      const recJava = getRecommendedJavaForMcVersion(version, software);
      if (resolveJavaSelectValue(targetSrv.java_path || '', targetSrv.version || '', targetSrv.type || '') !== recJava) {
        await api(nodeApiPath(targetNodeId, `/api/servers/${encodeURIComponent(srvId)}`), {
          method: 'PUT',
          body: JSON.stringify({
            name: targetSrv.name,
            description: targetSrv.description,
            type: targetSrv.type,
            version: targetSrv.version,
            jar_file: targetSrv.jar_file || 'server.jar',
            port: targetSrv.port,
            memory_min: targetSrv.memory_min,
            memory_max: targetSrv.memory_max,
            java_path: recJava,
            java_args: targetSrv.java_args,
            auto_restart: targetSrv.auto_restart
          })
        }).catch(() => {});
        targetSrv.java_path = recJava;
      }
    }

    closeModal();
    toast(`Server reinstalled successfully (${software} ${version})!`, 'ok');

    const cur = curServer();
    if (cur && cur.id === srvId && (cur.node_id || 'local') === targetNodeId) {
      S.ver = targetSrv?.version || version;
      S.core = software.charAt(0).toUpperCase() + software.slice(1);
    }
    await refreshServerStatus();
    if (S.view === 'fi') loadFileList(S.currentDir);
    R();
  } catch (err) {
    toast(err.message || 'Server reinstallation failed', 'bad');
    if (btn) {
      btn.disabled = false;
      btn.innerHTML = `${ic('refresh', 14)} Start Reinstallation`;
    }
    uiAlert({
      title: 'Server Reinstallation Failed',
      message: `Failed to download or reinstall <b>${esc(software)} ${esc(version)}</b>:<br><br>
        <div style="background:rgba(239,68,68,0.1); border:1px solid rgba(239,68,68,0.3); border-radius:8px; padding:10px 12px; margin-bottom:12px; font-size:13px; color:#fca5a5;">
          ${esc(err.message)}
        </div>
        Please verify the release tag and software engine, or check connectivity to the upstream release provider.`,
      type: 'error'
    });
  }
}

async function openAdminServerSettingsModal(srvId, nodeId = null) {
  const container = $('#modal-container');
  if (!container) return;

  const targetSrv = findServerRef(srvId, nodeId) || curServer();
  if (!targetSrv) return;

  const targetId = targetSrv.id;
  const targetNodeId = nodeId || targetSrv.node_id || 'local';
  const targetName = targetSrv.name || 'Server';

  if (!S.adminUsers || S.adminUsers.length === 0) {
    try {
      const users = await api('/api/admin/users');
      if (Array.isArray(users)) S.adminUsers = users;
    } catch (_) {}
  }

  const curRam = getServerMaxRamMB(targetSrv);
  const curCpu = targetSrv.cpu_limit_percent || 0;
  const curDisk = targetSrv.disk_limit_mb || 0;
  const curPort = targetSrv.port || 25565;
  const curOwner = targetSrv.owner_id || 1;
  const curNid = targetSrv.node_id || 'local';

  container.innerHTML = `<div class="pal" onclick="if(event.target===this)closeModal()">
    <div class="card pbox" style="padding:26px; max-width:580px; max-height:88vh; overflow-y:auto; border:1px solid var(--border-light);">
      <div class="row sb" style="margin-bottom:14px;">
        <h2 style="font-size:20px; margin:0; color:#fff; display:flex; align-items:center; gap:8px;">
          ${ic('sliders', 18)} Manage Server &amp; Limits: ${esc(targetName)}
        </h2>
        <button class="btn sm" onclick="closeModal()" aria-label="Close modal">&times;</button>
      </div>

      <p class="mu" style="font-size:13px; line-height:1.5; margin-bottom:16px;">
        Configure server tenant ownership, compute resource limits (CPU &amp; RAM), and disk storage quotas. Moving a server between nodes requires a real stop-and-transfer operation, not a settings change.
      </p>

      <form onsubmit="submitAdminServerSettings(event, '${esc(targetId)}', '${esc(targetNodeId)}')">
        <!-- Server Owner -->
        <div style="margin-bottom:14px;">
          <label class="f" style="margin-bottom:0;">
            <span class="row sb">
              <span>Server Owner</span>
              <span class="mu" style="font-size:11.5px;">Panel user account</span>
            </span>
            <select id="adm-set-owner" required>
              ${(S.adminUsers || []).map(u => `<option value="${u.id}" ${u.id === curOwner ? 'selected' : ''}>${esc(u.username)}${(u.first_name || u.last_name) ? ` (${esc([u.first_name, u.last_name].filter(Boolean).join(' '))})` : ''} [${esc(u.role)}]</option>`).join('')}
            </select>
          </label>
        </div>

        <!-- Node & Port -->
        <div class="grid g2" style="margin-bottom:14px;">
          <label class="f" style="margin-bottom:0;">
            <span>Node</span>
            <input id="adm-set-node" value="${esc(displayNodeName(curNid))}" disabled aria-label="Hosting node (read-only)">
            <span class="mu" style="font-size:11.5px;">Hosting node is read-only; changing metadata would not move the running process.</span>
          </label>
          <label class="f" style="margin-bottom:0;">
            <span>Port</span>
            <input type="number" id="adm-set-port" value="${curPort}" min="1024" max="65535" required>
          </label>
        </div>

        <!-- RAM Limit -->
        <div style="margin-bottom:14px;">
          <label class="f" style="margin-bottom:0;">
            <div class="row sb" style="margin-bottom:4px;">
              <span>Allocated RAM (MB)</span>
              <span class="mu" style="font-size:11.5px;">${isSteelEngine(targetSrv.type) ? 'Native resource limit (enforced with systemd only)' : '-Xmx memory ceiling'}</span>
            </div>
            <div class="row" style="gap:5px; margin-bottom:6px; flex-wrap:wrap;">
              <button type="button" class="btn sm" style="font-size:11px; padding:3px 7px;" onclick="$('#adm-set-ram').value=1024">1 GB</button>
              <button type="button" class="btn sm" style="font-size:11px; padding:3px 7px;" onclick="$('#adm-set-ram').value=2048">2 GB</button>
              <button type="button" class="btn sm" style="font-size:11px; padding:3px 7px;" onclick="$('#adm-set-ram').value=4096">4 GB</button>
              <button type="button" class="btn sm" style="font-size:11px; padding:3px 7px;" onclick="$('#adm-set-ram').value=8192">8 GB</button>
              <button type="button" class="btn sm" style="font-size:11px; padding:3px 7px;" onclick="$('#adm-set-ram').value=16384">16 GB</button>
            </div>
            <input type="number" id="adm-set-ram" value="${curRam}" min="512" max="65536" step="512" required>
          </label>
        </div>

        <!-- Resource Limits: CPU & Storage -->
        <div class="grid g2" style="margin-bottom:16px;">
          <label class="f" style="margin-bottom:0;">
            <div class="row sb" style="margin-bottom:4px;">
              <span>CPU Limit (%)</span>
              <span class="mu" style="font-size:11.5px;">0% = Unlimited</span>
            </div>
            <div class="row" style="gap:5px; margin-bottom:6px; flex-wrap:wrap;">
              <button type="button" class="btn sm" style="font-size:11px; padding:3px 7px;" onclick="$('#adm-set-cpu').value=0">Unlimited</button>
              <button type="button" class="btn sm" style="font-size:11px; padding:3px 7px;" onclick="$('#adm-set-cpu').value=100">100% (1 Core)</button>
              <button type="button" class="btn sm" style="font-size:11px; padding:3px 7px;" onclick="$('#adm-set-cpu').value=200">200% (2 Cores)</button>
            </div>
            <input type="number" id="adm-set-cpu" value="${curCpu}" min="0" max="3200" step="10" placeholder="0 for unlimited">
          </label>
          <label class="f" style="margin-bottom:0;">
            <div class="row sb" style="margin-bottom:4px;">
              <span>Storage / Disk Limit (MB)</span>
              <span class="mu" style="font-size:11.5px;">0 = Unlimited</span>
            </div>
            <div class="row" style="gap:5px; margin-bottom:6px; flex-wrap:wrap;">
              <button type="button" class="btn sm" style="font-size:11px; padding:3px 7px;" onclick="$('#adm-set-disk').value=0">Unlimited</button>
              <button type="button" class="btn sm" style="font-size:11px; padding:3px 7px;" onclick="$('#adm-set-disk').value=5120">5 GB</button>
              <button type="button" class="btn sm" style="font-size:11px; padding:3px 7px;" onclick="$('#adm-set-disk').value=10240">10 GB</button>
              <button type="button" class="btn sm" style="font-size:11px; padding:3px 7px;" onclick="$('#adm-set-disk').value=20480">20 GB</button>
            </div>
            <input type="number" id="adm-set-disk" value="${curDisk}" min="0" step="512" placeholder="0 for unlimited">
          </label>
        </div>

        <div class="row" style="justify-content:flex-end; gap:8px;">
          <button type="button" class="btn" onclick="closeModal()">Cancel</button>
          <button type="submit" class="btn pr" id="adm-set-submit">Save Server Configuration</button>
        </div>
      </form>
    </div>
  </div>`;
}

async function submitAdminServerSettings(e, srvId, nodeId) {
  if (e) e.preventDefault();
  const btn = $('#adm-set-submit');
  if (btn) {
    btn.disabled = true;
    btn.textContent = 'Saving...';
  }

  const targetSrv = findServerRef(srvId, nodeId) || curServer();
  if (!targetSrv) {
    toast('Server instance not found', 'bad');
    closeModal();
    return;
  }

  const ownerId = parseInt($('#adm-set-owner')?.value, 10) || 1;
  const port = parseInt($('#adm-set-port')?.value, 10) || targetSrv.port;
  const ram = parseInt($('#adm-set-ram')?.value, 10) || getServerMaxRamMB(targetSrv);
  const cpu = parseInt($('#adm-set-cpu')?.value, 10) || 0;
  const disk = parseInt($('#adm-set-disk')?.value, 10) || 0;

  const payload = {
    name: targetSrv.name,
    description: targetSrv.description,
    type: targetSrv.type,
    version: targetSrv.version,
    jar_file: targetSrv.jar_file,
    port: port,
    memory_min: isSteelEngine(targetSrv.type) ? 0 : Math.min(1024, ram),
    memory_max: ram,
    java_path: isSteelEngine(targetSrv.type) ? '' : targetSrv.java_path,
    java_args: isSteelEngine(targetSrv.type) ? '' : targetSrv.java_args,
    auto_restart: targetSrv.auto_restart !== false,
    owner_id: ownerId,
    cpu_limit_percent: cpu,
    disk_limit_mb: disk
  };

  try {
    await api(nodeApiPath(targetSrv.node_id || 'local', `/api/servers/${srvId}`), {
      method: 'PUT',
      body: JSON.stringify(payload)
    });

    targetSrv.owner_id = ownerId;
    targetSrv.port = port;
    targetSrv.memory_max = ram;
    targetSrv.memory_min = payload.memory_min;
    targetSrv.cpu_limit_percent = cpu;
    targetSrv.disk_limit_mb = disk;

    closeModal();
    toast('Server limits and ownership updated successfully', 'ok');
    R();
  } catch (err) {
    toast(`Failed to update server: ${err.message || err}`, 'bad');
    if (btn) {
      btn.disabled = false;
      btn.textContent = 'Save Server Configuration';
    }
  }
}

// --- Admin Panel Management (Users, GC, Customization) ---
let _targetAdminUploadUserId = null;

function adminTriggerUserAvatarUpload(userId) {
  _targetAdminUploadUserId = userId;
  const inp = $('#admin-user-avatar-file');
  if (inp) {
    inp.value = '';
    inp.click();
  }
}

async function onAdminUserAvatarSelected(input) {
  const file = input?.files?.[0];
  if (!file || !_targetAdminUploadUserId) return;
  const fd = new FormData();
  fd.append('avatar', file);
  try {
    const res = await api(`/api/admin/users/${_targetAdminUploadUserId}/avatar`, {
      method: 'POST',
      body: fd
    });
    toast('User icon updated successfully', 'ok');
    await loadAdminData();
  } catch (err) {
    toast(err.message || 'Failed to upload icon', 'bad');
  } finally {
    _targetAdminUploadUserId = null;
    if (input) input.value = '';
  }
}

async function uploadMyAvatar(input) {
  const file = input?.files?.[0];
  if (!file) return;
  const fd = new FormData();
  fd.append('avatar', file);
  try {
    const res = await api('/api/account/avatar', {
      method: 'POST',
      body: fd
    });
    if (res.avatar_url) {
      if (S.user) S.user.avatar_url = res.avatar_url;
      if (S.acct) S.acct.avatar_url = res.avatar_url;
      toast('Profile avatar updated successfully', 'ok');
      R();
    }
  } catch (err) {
    toast(err.message || 'Failed to upload avatar', 'bad');
  } finally {
    if (input) input.value = '';
  }
}

function openAddUserModal() {
  const container = $('#modal-container');
  if (!container) return;

  container.innerHTML = `<div class="pal" onclick="if(event.target===this)closeModal()">
    <div class="card pbox" style="padding:26px; max-width:480px; border:1px solid var(--border-light);">
      <div class="row sb" style="margin-bottom:16px;">
        <h2 style="font-size:20px; margin:0; color:#fff;">${ic('users', 18)} Create Panel User</h2>
        <button class="btn sm" onclick="closeModal()" aria-label="Close modal">&times;</button>
      </div>

      <form id="add-user-form" onsubmit="submitAddUser(event)">
        <label class="f" style="margin-bottom:12px;">
          <span>Username / Account Name <b style="color:var(--err); font-size:12px;">*required</b></span>
          <input type="text" id="nu-username" required placeholder="e.g. operator" autocomplete="off">
        </label>

        <div class="grid g2" style="margin-bottom:12px;">
          <label class="f">
            <span>First Name</span>
            <input type="text" id="nu-firstname" placeholder="Optional" autocomplete="off">
          </label>
          <label class="f">
            <span>Last Name</span>
            <input type="text" id="nu-lastname" placeholder="Optional" autocomplete="off">
          </label>
        </div>

        <label class="f" style="margin-bottom:12px;">
          <span>Email Address <b style="color:var(--err); font-size:12px;">*required</b></span>
          <input type="email" id="nu-email" required placeholder="user@example.com" autocomplete="off">
        </label>

        <label class="f" style="margin-bottom:12px;">
          <span>Password <b style="color:var(--err); font-size:12px;">*required</b></span>
          <input type="password" id="nu-password" required placeholder="Minimum 6 characters" autocomplete="new-password">
        </label>

        <label class="f" style="margin-bottom:12px;">
          <span>Role</span>
          <select id="nu-role" style="height:38px;">
            <option value="user" selected>Standard User (Access assigned servers)</option>
            <option value="admin">Administrator (Full cluster management)</option>
          </select>
        </label>

        <label class="f" style="margin-bottom:20px;">
          <span>Profile Icon (Optional)</span>
          <div class="row" style="gap:10px; align-items:center;">
            <input type="file" id="nu-avatar-file" accept="image/*" style="display:none;" onchange="const n=$('#nu-avatar-name'); if(n) n.textContent=this.files[0]?.name||'No file chosen';">
            <button type="button" class="btn sm" onclick="$('#nu-avatar-file').click()">${ic('image', 14)} Choose Icon...</button>
            <span id="nu-avatar-name" class="mu" style="font-size:12px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; max-width:200px;">No file chosen</span>
          </div>
        </label>

        <div class="row" style="justify-content:flex-end; gap:10px;">
          <button type="button" class="btn" onclick="closeModal()">Cancel</button>
          <button type="submit" id="nu-submit-btn" class="btn pr">Create User</button>
        </div>
      </form>
    </div>
  </div>`;

  enhanceSelects(container);
}

async function submitAddUser(e) {
  e.preventDefault();
  const btn = $('#nu-submit-btn');
  if (btn) btn.disabled = true;

  const username = $('#nu-username')?.value.trim();
  const firstName = $('#nu-firstname')?.value.trim() || '';
  const lastName = $('#nu-lastname')?.value.trim() || '';
  const email = $('#nu-email')?.value.trim();
  const password = $('#nu-password')?.value;
  const role = $('#nu-role')?.value || 'user';
  const avatarFile = $('#nu-avatar-file')?.files?.[0];

  if (!username || !password) {
    toast('Username and password are required', 'bad');
    if (btn) btn.disabled = false;
    return;
  }
  if (!email || !email.includes('@') || !email.includes('.')) {
    toast('A valid email address is required for user accounts', 'bad');
    if (btn) btn.disabled = false;
    return;
  }

  try {
    const newUser = await api('/api/admin/users', {
      method: 'POST',
      body: JSON.stringify({ username, first_name: firstName, last_name: lastName, email, password, role })
    });

    if (avatarFile && newUser?.id) {
      try {
        const fd = new FormData();
        fd.append('avatar', avatarFile);
        await api(`/api/admin/users/${newUser.id}/avatar`, {
          method: 'POST',
          body: fd
        });
      } catch (avErr) {
        console.warn('Avatar upload warning:', avErr);
      }
    }

    closeModal();
    toast(`User "${username}" created successfully`, 'ok');
    await loadAdminData();
  } catch (err) {
    toast(err.message || 'Failed to create user', 'bad');
    if (btn) btn.disabled = false;
    uiAlert({
      title: 'User Creation Failed',
      message: `Failed to create user account <b>${esc(username)}</b>:<br><br>${esc(err.message || 'Username or email already exists')}`,
      type: 'error'
    });
  }
}

function openEditUserModal(userId) {
  const container = $('#modal-container');
  if (!container) return;
  const u = (S.adminUsers || []).find(x => x.id === Number(userId));
  if (!u) {
    toast('User not found', 'bad');
    return;
  }

  const isMe = (S.user && (u.id === S.user.id || u.username === S.user.username)) || u.username === S.acct.un;

  container.innerHTML = `<div class="pal" onclick="if(event.target===this)closeModal()">
    <div class="card pbox" style="padding:26px; max-width:480px; border:1px solid var(--border-light);">
      <div class="row sb" style="margin-bottom:16px;">
        <h2 style="font-size:20px; margin:0; color:#fff;">${ic('edit', 18)} Edit User Account</h2>
        <button class="btn sm" onclick="closeModal()" aria-label="Close modal">&times;</button>
      </div>

      <form id="edit-user-form" onsubmit="submitEditUser(event, ${u.id})">
        <label class="f" style="margin-bottom:12px;">
          <span>Username / Account Name <b style="color:var(--err); font-size:12px;">*required</b></span>
          <input type="text" id="eu-username" required value="${esc(u.username)}" placeholder="Username" autocomplete="off">
        </label>

        <div class="grid g2" style="margin-bottom:12px;">
          <label class="f">
            <span>First Name</span>
            <input type="text" id="eu-firstname" value="${esc(u.first_name || '')}" placeholder="First name" autocomplete="off">
          </label>
          <label class="f">
            <span>Last Name</span>
            <input type="text" id="eu-lastname" value="${esc(u.last_name || '')}" placeholder="Last name" autocomplete="off">
          </label>
        </div>

        <label class="f" style="margin-bottom:12px;">
          <span>Email Address <b style="color:var(--err); font-size:12px;">*required</b></span>
          <input type="email" id="eu-email" required value="${esc(u.email || '')}" placeholder="user@example.com" autocomplete="off">
        </label>

        <label class="f" style="margin-bottom:12px;">
          <span>New Password <span class="mu" style="font-size:12px;">(leave blank to keep current)</span></span>
          <input type="password" id="eu-password" placeholder="Leave blank to keep current" autocomplete="new-password">
        </label>

        <label class="f" style="margin-bottom:20px;">
          <span>Role</span>
          <select id="eu-role" style="height:38px;" ${isMe ? 'disabled title="You cannot change your own role"' : ''}>
            <option value="user" ${u.role !== 'admin' ? 'selected' : ''}>Standard User (Access assigned servers)</option>
            <option value="admin" ${u.role === 'admin' ? 'selected' : ''}>Administrator (Full cluster management)</option>
          </select>
        </label>

        <div class="row" style="justify-content:flex-end; gap:10px;">
          <button type="button" class="btn" onclick="closeModal()">Cancel</button>
          <button type="submit" id="eu-submit-btn" class="btn pr">Save Changes</button>
        </div>
      </form>
    </div>
  </div>`;

  enhanceSelects(container);
}

async function submitEditUser(e, userId) {
  e.preventDefault();
  const btn = $('#eu-submit-btn');
  if (btn) btn.disabled = true;

  const username = $('#eu-username')?.value.trim();
  const firstName = $('#eu-firstname')?.value.trim() || '';
  const lastName = $('#eu-lastname')?.value.trim() || '';
  const email = $('#eu-email')?.value.trim();
  const password = $('#eu-password')?.value || '';
  const role = $('#eu-role')?.value || 'user';

  if (!username) {
    toast('Username / Account name is required', 'bad');
    if (btn) btn.disabled = false;
    return;
  }
  if (!email || !email.includes('@') || !email.includes('.')) {
    toast('A valid email address is required', 'bad');
    if (btn) btn.disabled = false;
    return;
  }

  try {
    const payload = {
      username,
      first_name: firstName,
      last_name: lastName,
      email,
      role
    };
    if (password) {
      payload.password = password;
    }

    await api(`/api/admin/users/${userId}`, {
      method: 'PUT',
      body: JSON.stringify(payload)
    });

    closeModal();
    toast(`User "${username}" updated successfully`, 'ok');

    if (S.user && S.user.id === userId) {
      S.user.username = username;
      S.user.first_name = firstName;
      S.user.last_name = lastName;
      S.user.email = email;
      S.acct.un = username;
      S.acct.fn = firstName;
      S.acct.ln = lastName;
      S.acct.em = email;
      localStorage.setItem('neopanel_profile', JSON.stringify({
        un: username,
        em: email,
        fn: firstName,
        ln: lastName
      }));
    }

    await loadAdminData();
    R();
  } catch (err) {
    toast(err.message || 'Failed to update user', 'bad');
    if (btn) btn.disabled = false;
  }
}

async function deleteAdminUser(id, username) {
  const ok = await uiConfirm({
    title: 'Delete User Account',
    message: `Are you sure you want to permanently delete user account <b>${esc(username)}</b>? This action cannot be reversed.`,
    confirmText: 'Delete User',
    cancelText: 'Cancel',
    isDanger: true
  });
  if (!ok) return;

  try {
    await api(`/api/admin/users/${id}`, { method: 'DELETE' });
    toast(`User "${username}" deleted`, 'ok');
    await loadAdminData();
  } catch (err) {
    toast(err.message || 'Failed to delete user', 'bad');
  }
}

async function triggerAdminGC() {
  const ok = await uiConfirm({
    title: 'Clean up panel memory?',
    message: 'Run Go garbage collection to clean up unused memory in the panel process? This does not change Minecraft-server memory, and the amount released may vary.',
    confirmText: 'Clean up memory',
    cancelText: 'Cancel',
    isDanger: false,
    icon: 'cpu'
  });
  if (!ok) return;

  toast('Cleaning up unused panel memory...');
  try {
    const res = await api('/api/admin/gc', { method: 'POST' });
    toast(res.message || 'Panel memory cleanup completed', 'ok');
    await loadAdminData();
  } catch (err) {
    toast(err.message || 'Panel memory cleanup failed', 'bad');
  }
}

async function checkAdminUpdates() {
  toast('Checking GitHub for updates...');
  try {
    const res = await api('/api/admin/updates');
    S.adminUpdate = res;
    if (res.error) {
      toast(`Update check note: ${res.error}`, 'bad');
    } else if (res.update_available) {
      toast(`Update available: commit #${res.remote_commit?.slice(0, 7) || 'latest'}`);
    } else {
      toast('NeoPanel is up to date with remote repository!');
    }
    R();
  } catch (e) {
    toast('Failed to check updates: ' + (e.message || e), 'bad');
  }
}

async function saveAdminUpdateConfig() {
  const repoInput = $('#admin-update-repo');
  const branchInput = $('#admin-update-branch');
  const repo_url = repoInput ? repoInput.value.trim() : '';
  const branch = branchInput ? branchInput.value.trim() : 'main';

  if (!repo_url) {
    toast('Repository URL cannot be empty', 'bad');
    return;
  }

  try {
    const res = await api('/api/admin/updates/config', {
      method: 'POST',
      body: JSON.stringify({ repo_url, branch })
    });
    toast(res.message || 'Repository settings saved!');
    await checkAdminUpdates();
  } catch (e) {
    toast(e.message || 'Failed to save repository settings', 'bad');
  }
}

function resetAdminUpdateRepo() {
  const r = $('#admin-update-repo');
  const b = $('#admin-update-branch');
  if (r) r.value = 'https://github.com/Pekly/NeoPanel.git';
  if (b) b.value = 'main';
}

async function openApplyUpdateModal() {
  const up = S.adminUpdate || {};
  const repo = up.repo_url || 'https://github.com/Pekly/NeoPanel.git';
  const branch = up.branch || 'main';

  const ok = await uiConfirm({
    title: 'Confirm NeoPanel System Update',
    message: `
      Are you sure you want to trigger a panel binary update now?<br><br>
      <div style="background:rgba(255,255,255,0.04); border:1px solid var(--border); border-radius:8px; padding:10px; font-size:12px; margin:8px 0;">
        <div><b>Target Repository:</b> <code>${esc(repo)}</code></div>
        <div><b>Target Branch:</b> <code>${esc(branch)}</code></div>
        <div><b>Current Version:</b> <code>v${esc(S.system?.version || '1.0.0')}</code></div>
      </div>
      The daemon will pull the latest source code from GitHub, compile the new binary, and restart the NeoPanel service. Running Minecraft servers will remain online.
    `,
    confirmText: 'Apply Update Now',
    cancelText: 'Cancel',
    isDanger: false,
    icon: 'download'
  });
  if (!ok) return;

  await executePanelUpdate();
}

function renderAdminUpdateBtn(up) {
  const isUpdating = !!(S.updateProgress?.running || up?.is_updating);
  let pct = S.updateProgress?.pct || 0;
  let step = S.updateProgress?.step || (isUpdating ? 'Updating...' : '');
  if (isUpdating && pct < 5) pct = 5;

  if (isUpdating) {
    return `
      <button type="button" class="btn sm pr np-update-btn is-updating" id="admin-update-now-btn" disabled style="position:relative; overflow:hidden; min-width:220px; justify-content:center; padding:6px 14px; background:#111927; border:1px solid rgba(56,122,255,0.45); cursor:wait;" title="Panel update is compiling and applying in the background">
        <span class="np-update-bar-fill" id="admin-update-bar-fill" style="position:absolute; left:0; top:0; bottom:0; width:${pct}%; background:linear-gradient(90deg, rgba(56,122,255,0.4), rgba(56,122,255,0.75)); transition:width 0.35s ease; border-radius:inherit; pointer-events:none;"></span>
        <span style="position:relative; z-index:2; display:inline-flex; align-items:center; gap:8px; font-weight:600; font-size:12px; color:#ffffff;">
          <span class="spinner-sm" style="width:12px; height:12px; border:2px solid rgba(255,255,255,0.3); border-top-color:#fff; border-radius:50%; animation:spin 0.8s linear infinite;"></span>
          <span id="admin-update-bar-label">${pct}% ${esc(step)}</span>
        </span>
      </button>
    `;
  }

  const hasUpdate = up?.update_available;
  const label = hasUpdate ? `Update panel (${esc(up.remote_commit?.slice(0, 7) || 'new')})` : 'Update panel';
  return `
    <button type="button" class="btn sm pr np-update-btn" id="admin-update-now-btn" onclick="openApplyUpdateModal()">
      ${ic('download', 13)} <span>${label}</span>
    </button>
  `;
}

async function executePanelUpdate() {
  try {
    S.updateProgress = { running: true, pct: 5, step: 'Starting update...' };
    if (!S.adminUpdate) S.adminUpdate = {};
    S.adminUpdate.is_updating = true;

    const wrap = $('#admin-update-actions-wrap');
    if (wrap) {
      wrap.innerHTML = `<button type="button" class="btn sm" onclick="checkAdminUpdates()">${ic('refresh', 13)} Check Updates</button>${renderAdminUpdateBtn(S.adminUpdate)}`;
    } else {
      R();
    }

    startUpdateProgressPolling();

    const res = await api('/api/admin/updates/apply', { method: 'POST' });
    toast(res.message || 'Update started in background!', 'ok');
  } catch (e) {
    S.updateProgress = null;
    if (S.adminUpdate) S.adminUpdate.is_updating = false;
    toast(e.message || 'Failed to start update', 'bad');
    R();
  }
}

let _updatePoller = null;
function startUpdateProgressPolling() {
  if (_updatePoller) return;

  const poll = async () => {
    try {
      const res = await api('/api/admin/updates/logs');
      const isRunning = !!res.running;
      let pct = typeof res.progress === 'number' ? res.progress : 0;
      let step = res.step || '';

      if (isRunning && pct < 5) pct = 5;
      if (!step) step = isRunning ? 'Updating...' : 'Ready';

      S.updateProgress = { running: isRunning, pct, step };

      // Update button elements directly in DOM
      const barFill = $('#admin-update-bar-fill');
      const barLabel = $('#admin-update-bar-label');
      if (barFill) barFill.style.width = pct + '%';
      if (barLabel) barLabel.textContent = `${pct}% ${step}`;

      if (!isRunning) {
        clearInterval(_updatePoller);
        _updatePoller = null;

        if (pct >= 80 || res.logs?.includes('Successfully Updated') || res.logs?.includes('NeoPanel Update SUCCESS')) {
          if (barFill) {
            barFill.style.width = '100%';
            barFill.style.background = '#10b981';
          }
          if (barLabel) barLabel.innerHTML = '✔ 100% Updated &bull; Reconnecting...';
          toast('NeoPanel successfully updated! Reconnecting to panel...', 'ok');

          waitForPanelRestart();
        } else {
          S.updateProgress = null;
          if (S.adminUpdate) S.adminUpdate.is_updating = false;
          R();
        }
      }
    } catch (e) {
      // Panel daemon may be restarting right now
      const barLabel = $('#admin-update-bar-label');
      if (barLabel && (S.updateProgress?.pct >= 80)) {
        barLabel.innerHTML = '✔ Restarting &bull; Reconnecting...';
      }
    }
  };

  poll();
  _updatePoller = setInterval(poll, 1500);
}

function waitForPanelRestart() {
  let attempts = 0;
  const maxAttempts = 35; // ~50s
  const ping = setInterval(async () => {
    attempts++;
    try {
      const res = await fetch('/api/system', { method: 'GET', cache: 'no-cache' });
      if (res.ok) {
        clearInterval(ping);
        S.updateProgress = null;
        if (S.adminUpdate) S.adminUpdate.is_updating = false;
        toast('NeoPanel is back online with the latest update!', 'ok');
        await loadAdminData();
        R();
      }
    } catch (e) {
      if (attempts >= maxAttempts) {
        clearInterval(ping);
        S.updateProgress = null;
        if (S.adminUpdate) S.adminUpdate.is_updating = false;
        toast('Panel update complete! Please refresh the page.', 'ok');
        R();
      }
    }
  }, 1500);
}

let _updateLogTimer = null;
function openUpdateLogsModal() {
  uiAlert({
    title: 'NeoPanel Installer & Build Logs',
    message: `
      <div style="display:flex; flex-direction:column; gap:8px;">
        <div class="row sb" style="align-items:center;">
          <span id="update-log-status" class="con-live-pill"><i></i>Streaming logs...</span>
          <button type="button" class="btn sm" onclick="pollUpdateLogsOnce()">${ic('refresh', 12)} Refresh</button>
        </div>
        <pre id="update-log-pre" style="background:#090d16; border:1px solid var(--border); border-radius:8px; padding:12px; font-family:var(--font-mono, monospace); font-size:11.5px; line-height:1.5; color:#e2e8f0; height:320px; overflow-y:auto; white-space:pre-wrap; word-break:break-all;">Loading build logs from /tmp/neopanel-install.log...</pre>
      </div>
    `,
    type: 'info'
  });

  pollUpdateLogsOnce();
  if (_updateLogTimer) clearInterval(_updateLogTimer);
  _updateLogTimer = setInterval(pollUpdateLogsOnce, 2500);
}

async function pollUpdateLogsOnce() {
  try {
    const data = await api('/api/admin/updates/logs');
    const pre = $('#update-log-pre');
    const pill = $('#update-log-status');
    if (pre) {
      let clean = String(data.logs || 'No update logs found yet (/tmp/neopanel-install.log is empty).');
      clean = clean.replace(/\x1b\[[0-9;]*[a-zA-Z]|\x1b\([a-zA-Z]|\r/g, '');
      pre.textContent = clean;
      pre.scrollTop = pre.scrollHeight;
    }
    if (pill) {
      if (data.running) {
        pill.className = 'con-live-pill';
        pill.innerHTML = `<i></i>${data.progress ? data.progress + '% ' : ''}${esc(data.step || 'Compiling & updating...')}`;
      } else {
        pill.className = 'badge online';
        pill.innerHTML = '<i class="dot"></i>Process Idle / Completed';
        if (_updateLogTimer) {
          clearInterval(_updateLogTimer);
          _updateLogTimer = null;
        }
      }
    }
  } catch (e) {}
}

function renderAdminEmailTabHtml() {
  const smtp = S.adminSmtp || {};
  return `
    <div class="card tw" style="margin-bottom:18px;">
      <div class="row sb" style="margin-bottom:14px; flex-wrap:wrap; gap:12px;">
        <div>
          <h2 style="margin:0 0 4px; color:var(--tx); display:flex; align-items:center; gap:8px;">
            ${ic('mail', 18)} Email delivery
          </h2>
          <span class="mu" style="font-size:13px;">Configure outgoing SMTP mail server credentials to send password reset links, notification alerts, and subuser invitations.</span>
        </div>
        <div class="row" style="gap:8px;">
          <button type="button" class="btn sm" onclick="loadAdminSmtp()">${ic('refresh', 13)} Refresh</button>
          <button type="button" class="btn sm pr" onclick="saveAdminSmtp()">${ic('save', 13)} Save Email Settings</button>
        </div>
      </div>

      <div style="background:rgba(255,255,255,0.02); border:1px solid var(--border); border-radius:8px; padding:16px; margin-bottom:18px;">
        <div class="row sb" style="margin-bottom:8px; align-items:center; flex-wrap:wrap; gap:8px;">
          <span style="font-size:13px; font-weight:600; color:var(--tx);">Quick Email Provider Presets</span>
          <div style="min-width:260px;">
            <select class="sel" id="admin-smtp-preset" onchange="applySmtpPreset(this.value)">
              <option value="custom">Custom SMTP Server...</option>
              <option value="cloudflare">Cloudflare Email Routing / Mailchannels</option>
              <option value="gmail">Gmail / Google Workspace (smtp.gmail.com:587)</option>
              <option value="brevo">Brevo / Sendinblue (smtp-relay.brevo.com:587)</option>
              <option value="sendgrid">SendGrid (smtp.sendgrid.net:587)</option>
              <option value="resend">Resend SMTP (smtp.resend.com:465)</option>
              <option value="ses">Amazon SES (Port 587)</option>
            </select>
          </div>
        </div>
        <p class="mu" style="font-size:12px; margin:0;">Selecting a provider preset automatically populates optimal host, port, and TLS encryption modes.</p>
      </div>

      <div style="display:grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap:16px; margin-bottom:18px;">
        <div>
          <label class="lbl" style="font-size:12px;">Email Delivery Active</label>
          <label class="row" style="gap:10px; cursor:pointer; padding:6px 0;">
            <input type="checkbox" id="admin-smtp-enabled" ${smtp.enabled ? 'checked' : ''} style="width:16px; height:16px; accent-color:var(--ac);">
            <span style="font-size:13px; color:var(--tx);">Enable outgoing emails for password resets</span>
          </label>
        </div>
        <div>
          <label class="lbl" style="font-size:12px;">Encryption Security</label>
          <select class="sel" id="admin-smtp-enc">
            <option value="starttls" ${smtp.encryption === 'starttls' ? 'selected' : ''}>STARTTLS (Port 587 recommended)</option>
            <option value="tls" ${smtp.encryption === 'tls' || smtp.encryption === 'ssl' ? 'selected' : ''}>Direct SSL/TLS (Port 465)</option>
            <option value="none" ${smtp.encryption === 'none' ? 'selected' : ''}>Plain / Unencrypted (Port 25)</option>
          </select>
        </div>
        <div>
          <label class="lbl" style="font-size:12px;">SMTP Server Host</label>
          <input type="text" id="admin-smtp-host" class="inp mono" value="${esc(smtp.host || '')}" placeholder="smtp.gmail.com or mail.yourdomain.com">
        </div>
        <div>
          <label class="lbl" style="font-size:12px;">SMTP Port</label>
          <input type="number" id="admin-smtp-port" class="inp mono" value="${smtp.port || 587}" placeholder="587">
        </div>
        <div>
          <label class="lbl" style="font-size:12px;">SMTP Username / API Key</label>
          <input type="text" id="admin-smtp-user" class="inp mono" value="${esc(smtp.username || '')}" placeholder="username@example.com or apikey">
        </div>
        <div>
          <label class="lbl" style="font-size:12px;">SMTP Password / Token</label>
          <div class="row" style="position:relative;">
            <input type="password" id="admin-smtp-pass" class="inp mono" value="${esc(smtp.password || '')}" placeholder="${smtp.password ? '••••••••' : 'Enter password or API token'}" style="padding-right:36px;">
            <button type="button" class="eye" onclick="const p=$('#admin-smtp-pass'); p.type=p.type==='password'?'text':'password';" style="position:absolute; right:8px; top:50%; transform:translateY(-50%); background:none; border:none; color:var(--tx-muted); cursor:pointer;">${ic('eye', 14)}</button>
          </div>
        </div>
        <div>
          <label class="lbl" style="font-size:12px;">From Sender Email</label>
          <input type="email" id="admin-smtp-from-email" class="inp mono" value="${esc(smtp.from_email || '')}" placeholder="noreply@yourdomain.com">
        </div>
        <div>
          <label class="lbl" style="font-size:12px;">From Display Name</label>
          <input type="text" id="admin-smtp-from-name" class="inp" value="${esc(smtp.from_name || 'NeoPanel')}" placeholder="NeoPanel">
        </div>
      </div>

      <div style="background:rgba(255,255,255,0.02); border:1px solid var(--border); border-radius:8px; padding:16px;">
        <h3 style="margin:0 0 8px; font-size:14px; color:var(--tx); display:flex; align-items:center; gap:6px;">
          ${ic('mail', 14)} Test Email Delivery
        </h3>
        <p class="mu" style="font-size:12px; margin:0 0 12px;">Send an immediate diagnostic email to confirm your SMTP host connection, TLS certificate validation, and authentication.</p>
        <div class="row" style="gap:10px; flex-wrap:wrap;">
          <input type="email" id="admin-smtp-test-email" class="inp mono" style="flex:1 1 260px;" placeholder="Recipient email (e.g. your personal email)" value="${esc(S.user?.email || '')}">
          <button type="button" class="btn" onclick="sendAdminTestEmail(this)">${ic('mail', 13)} Send Test Email</button>
        </div>
      </div>
    </div>
  `;
}

function applySmtpPreset(preset) {
  const host = $('#admin-smtp-host');
  const port = $('#admin-smtp-port');
  const enc = $('#admin-smtp-enc');
  const user = $('#admin-smtp-user');

  if (preset === 'gmail') {
    if (host) host.value = 'smtp.gmail.com';
    if (port) port.value = '587';
    if (enc) enc.value = 'starttls';
  } else if (preset === 'cloudflare') {
    if (host) host.value = 'smtp.mailchannels.net';
    if (port) port.value = '587';
    if (enc) enc.value = 'starttls';
  } else if (preset === 'brevo') {
    if (host) host.value = 'smtp-relay.brevo.com';
    if (port) port.value = '587';
    if (enc) enc.value = 'starttls';
  } else if (preset === 'sendgrid') {
    if (host) host.value = 'smtp.sendgrid.net';
    if (port) port.value = '587';
    if (enc) enc.value = 'starttls';
    if (user && !user.value) user.value = 'apikey';
  } else if (preset === 'resend') {
    if (host) host.value = 'smtp.resend.com';
    if (port) port.value = '465';
    if (enc) enc.value = 'tls';
    if (user && !user.value) user.value = 'resend';
  } else if (preset === 'ses') {
    if (host) host.value = 'email-smtp.us-east-1.amazonaws.com';
    if (port) port.value = '587';
    if (enc) enc.value = 'starttls';
  }
}

async function loadAdminSmtp() {
  toast('Loading SMTP configuration...');
  try {
    const res = await api('/api/admin/email');
    S.adminSmtp = res;
    toast('SMTP settings refreshed', 'ok');
    R();
  } catch (e) {
    toast('Failed to load SMTP settings: ' + (e.message || e), 'bad');
  }
}

async function saveAdminSmtp() {
  const enabled = Boolean($('#admin-smtp-enabled')?.checked);
  const host = $('#admin-smtp-host')?.value.trim() || '';
  const port = parseInt($('#admin-smtp-port')?.value.trim() || '587', 10);
  const username = $('#admin-smtp-user')?.value.trim() || '';
  const password = $('#admin-smtp-pass')?.value.trim() || '';
  const from_email = $('#admin-smtp-from-email')?.value.trim() || '';
  const from_name = $('#admin-smtp-from-name')?.value.trim() || '';
  const encryption = $('#admin-smtp-enc')?.value || 'starttls';

  try {
    const res = await api('/api/admin/email', {
      method: 'POST',
      body: JSON.stringify({
        enabled,
        host,
        port,
        username,
        password,
        from_email,
        from_name,
        encryption
      })
    });
    toast(res.message || 'SMTP settings saved successfully!', 'ok');
    await loadAdminSmtp();
  } catch (e) {
    toast(e.message || 'Failed to save SMTP settings', 'bad');
  }
}

async function sendAdminTestEmail(btn) {
  const to_email = $('#admin-smtp-test-email')?.value.trim() || '';
  if (!to_email) {
    toast('Please enter a recipient email address for testing', 'bad');
    return;
  }

  setBtnLoading(btn, true, 'Sending...');
  try {
    const res = await api('/api/admin/email/test', {
      method: 'POST',
      body: JSON.stringify({ to_email })
    });
    setBtnLoading(btn, false);
    toast(res.message || 'Test email sent successfully!', 'ok');
    uiAlert({
      title: 'Email Delivery Successful',
      message: `Diagnostic test email was successfully dispatched to <b>${esc(to_email)}</b> via your configured SMTP mail server!`,
      type: 'success'
    });
  } catch (e) {
    setBtnLoading(btn, false);
    toast('Failed to send test email: ' + (e.message || e), 'bad');
    uiAlert({
      title: 'SMTP Delivery Failed',
      message: `Failed to deliver test email:<br><br><code style="color:#ef4444; word-break:break-all;">${esc(e.message || e)}</code><br><br>Please verify your SMTP host, port, credentials, and TLS encryption mode.`,
      type: 'error'
    });
  }
}

function openResetPasswordModal() {
  openModal({
    title: 'Reset Account Password',
    body: `
      <div style="font-size: 13.5px; line-height: 1.6; color: var(--tx-secondary);">
        <p style="margin-top: 0; margin-bottom: 12px;">
          Enter your registered username or email address below. We'll send you a password reset verification link and code.
        </p>
        <div style="margin-bottom: 14px;">
          <label class="lbl" style="font-size: 12px; margin-bottom: 6px;">Username or Email Address</label>
          <input type="text" id="reset-identity" class="inp" placeholder="e.g. admin or user@example.com" autocomplete="username" autofocus onkeydown="if(event.key==='Enter')submitResetPasswordRequest()">
        </div>
        ${S.turnstileEnabled ? '<div id="turnstile-reset" class="turnstile-wrap" style="margin-bottom:12px;"></div>' : ''}
        <div class="lerr" id="reset-err" role="alert" style="margin-bottom: 10px; color:#ef4444; font-size:12.5px;"></div>
        <div class="row" style="margin-top: 6px;">
          <a href="javascript:void(0)" onclick="openSetNewPasswordModal()" style="font-size: 12px; color: var(--ac); text-decoration: none;">
            Already have a reset code? Enter it here &rarr;
          </a>
        </div>
      </div>
    `,
    buttons: [
      { label: 'Cancel', class: 'btn', onClick: () => closeModal() },
      {
        label: 'Send Reset Link',
        class: 'btn pr',
        onClick: () => submitResetPasswordRequest()
      }
    ]
  });

  if (S.turnstileEnabled) {
    mountTurnstileWidget('#turnstile-reset');
  }
  setTimeout(() => $('#reset-identity')?.focus(), 150);
}

async function submitResetPasswordRequest() {
  const identity = $('#reset-identity')?.value.trim();
  const err = $('#reset-err');
  if (!identity) {
    if (err) err.textContent = 'Please enter your username or email address.';
    return;
  }

  if (S.turnstileEnabled && !S.turnstileToken) {
    if (err) err.textContent = 'Please complete the Cloudflare Turnstile security check.';
    return;
  }

  const btn = $('#modal-container .btn.pr');
  setBtnLoading(btn, true, 'Sending...');

  try {
    const res = await (await fetch('/api/auth/password-reset/request', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        identity,
        'cf-turnstile-response': S.turnstileToken
      })
    })).json();

    setBtnLoading(btn, false);
    resetTurnstileWidget();

    if (res.error) {
      if (err) err.textContent = res.error;
      return;
    }

    closeModal();
    if (res.token) {
      uiAlert({
        title: 'Password Reset Code Generated',
        message: `
          <div style="font-size:13.5px; line-height:1.6;">
            <p style="margin-top:0;">${esc(res.message || 'A reset verification code was generated.')}</p>
            <div style="background:rgba(255,255,255,0.05); border:1px solid var(--border); border-radius:8px; padding:12px; margin:14px 0;">
              <span class="mu" style="font-size:11.5px; display:block; margin-bottom:4px;">Your Reset Code:</span>
              <code class="mono" style="font-size:14px; color:#38bdf8; font-weight:bold; user-select:all;">${esc(res.token)}</code>
            </div>
            <p class="mu" style="font-size:12px; margin-bottom:0;">Click the button below to set your new password now using this code.</p>
          </div>
        `,
        type: 'info',
        confirmText: 'Set New Password Now',
        onConfirm: () => openSetNewPasswordModal(res.token)
      });
    } else {
      uiAlert({
        title: 'Reset Instructions Sent',
        message: `
          <div style="font-size:13.5px; line-height:1.6;">
            <p style="margin-top:0;">${esc(res.message || 'If an account exists with that address, instructions and a verification code have been sent.')}</p>
            <p class="mu" style="font-size:12.5px;">Please check your email inbox (and spam folder) for the password reset link or code.</p>
          </div>
        `,
        type: 'info',
        confirmText: 'Enter Reset Code',
        onConfirm: () => openSetNewPasswordModal()
      });
    }
  } catch (e) {
    setBtnLoading(btn, false);
    resetTurnstileWidget();
    if (err) err.textContent = 'Connection error: ' + (e.message || e);
  }
}

function openSetNewPasswordModal(presetToken = '') {
  openModal({
    title: 'Choose a New Password',
    body: `
      <div style="font-size: 13.5px; line-height: 1.6; color: var(--tx-secondary);">
        <p style="margin-top: 0; margin-bottom: 12px;">
          Enter the verification code from your email (or provided by the system) along with your new password.
        </p>
        <div style="margin-bottom: 12px;">
          <label class="lbl" style="font-size: 12px; margin-bottom: 5px;">Reset Code / Token</label>
          <input type="text" id="reset-token-input" class="inp mono" value="${esc(presetToken)}" placeholder="Paste your 32-character reset code" autocomplete="off">
        </div>
        <div style="margin-bottom: 12px;">
          <label class="lbl" style="font-size: 12px; margin-bottom: 5px;">New Password</label>
          <div class="row" style="position:relative;">
            <input type="password" id="reset-new-password" class="inp" placeholder="At least 6 characters" autocomplete="new-password" style="padding-right:36px;">
            <button type="button" class="eye" onclick="const p=$('#reset-new-password'); p.type=p.type==='password'?'text':'password';" style="position:absolute; right:8px; top:50%; transform:translateY(-50%); background:none; border:none; color:var(--tx-muted); cursor:pointer;">${ic('eye', 14)}</button>
          </div>
        </div>
        <div style="margin-bottom: 14px;">
          <label class="lbl" style="font-size: 12px; margin-bottom: 5px;">Confirm New Password</label>
          <div class="row" style="position:relative;">
            <input type="password" id="reset-new-password-confirm" class="inp" placeholder="Repeat new password" autocomplete="new-password" style="padding-right:36px;" onkeydown="if(event.key==='Enter')submitSetNewPassword()">
            <button type="button" class="eye" onclick="const p=$('#reset-new-password-confirm'); p.type=p.type==='password'?'text':'password';" style="position:absolute; right:8px; top:50%; transform:translateY(-50%); background:none; border:none; color:var(--tx-muted); cursor:pointer;">${ic('eye', 14)}</button>
          </div>
        </div>
        <div class="lerr" id="set-pw-err" role="alert" style="margin-bottom: 10px; color:#ef4444; font-size:12.5px;"></div>
      </div>
    `,
    buttons: [
      { label: 'Cancel', class: 'btn', onClick: () => closeModal() },
      {
        label: 'Update Password',
        class: 'btn pr',
        onClick: () => submitSetNewPassword()
      }
    ]
  });

  setTimeout(() => {
    if (presetToken) {
      $('#reset-new-password')?.focus();
    } else {
      $('#reset-token-input')?.focus();
    }
  }, 150);
}

async function submitSetNewPassword() {
  const token = $('#reset-token-input')?.value.trim();
  const p1 = $('#reset-new-password')?.value;
  const p2 = $('#reset-new-password-confirm')?.value;
  const err = $('#set-pw-err');

  if (!token) {
    if (err) err.textContent = 'Please enter your reset code / token.';
    return;
  }
  if (!p1 || p1.length < 6) {
    if (err) err.textContent = 'New password must be at least 6 characters.';
    return;
  }
  if (p1 !== p2) {
    if (err) err.textContent = 'Passwords do not match. Please re-enter.';
    return;
  }

  const btn = $('#modal-container .btn.pr');
  setBtnLoading(btn, true, 'Updating...');

  try {
    const res = await (await fetch('/api/auth/password-reset/confirm', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        token,
        new_password: p1
      })
    })).json();

    setBtnLoading(btn, false);

    if (res.error) {
      if (err) err.textContent = res.error;
      return;
    }

    closeModal();
    toast('Password reset successful! You can now log in.', 'ok');
    if (window.location.hash.includes('reset-token=')) {
      window.location.hash = '';
    }
    $('#login')?.remove();
    showLoginScreen();
    if (res.username) {
      const uInp = $('#lu');
      if (uInp) uInp.value = res.username;
      $('#lp')?.focus();
    }
  } catch (e) {
    setBtnLoading(btn, false);
    if (err) err.textContent = 'Connection error: ' + (e.message || e);
  }
}

function checkUrlResetToken() {
  if (window.location.hash && window.location.hash.includes('reset-token=')) {
    const m = /reset-token=([^&]+)/i.exec(window.location.hash);
    if (m && m[1]) {
      setTimeout(() => openSetNewPasswordModal(decodeURIComponent(m[1])), 200);
    }
  }
}

function selectBrandPreset(presetId) {
  const p = BRAND_PRESETS.find(x => x.id === presetId);
  if (!p) return;
  S.brandDraft = {
    ...DEFAULT_BRANDING,
    ...(S.brandDraft || S.branding),
    preset: p.id,
    accent_color: p.accent_color,
    accent_hover: p.accent_hover,
    bg_deep: p.bg_deep,
    bg_canvas: p.bg_canvas,
    card_bg: p.card_bg,
    nav_bg: p.nav_bg,
    text_color: p.text_color,
    card_opacity: p.card_opacity,
    border_radius_px: p.border_radius_px,
    glass_blur_px: p.glass_blur_px
  };
  applyBranding(S.brandDraft);
  R();
}

function updateBrandLivePreview() {
  const draft = S.brandDraft || S.branding;
  const prevEl = $('#brand-live-preview');
  if (prevEl) prevEl.innerHTML = renderBrandPreviewMarkup(draft);
  const chipEl = document.querySelector('.brand-preview-card .chip');
  if (chipEl) chipEl.textContent = draft?.preset || 'custom';
}

function onBrandInputChange(field, val, srcEl = null) {
  if (!S.brandDraft) S.brandDraft = { ...DEFAULT_BRANDING, ...S.branding };
  S.brandDraft[field] = val;

  if (srcEl) {
    const row = srcEl.closest('.brand-color-row');
    if (row) {
      if (srcEl.type === 'color') {
        const txt = row.querySelector('input[type="text"]');
        if (txt) txt.value = val;
      } else if (srcEl.type === 'text' && /^#[0-9a-fA-F]{6}$/.test(String(val || '').trim())) {
        const pick = row.querySelector('input[type="color"]');
        if (pick) pick.value = String(val).trim().toLowerCase();
      }
    }
  }

  if (['accent_color', 'accent_hover', 'bg_deep', 'bg_canvas', 'card_bg', 'nav_bg', 'text_color', 'card_opacity', 'border_radius_px', 'glass_blur_px'].includes(field)) {
    S.brandDraft.preset = 'custom';
    document.querySelectorAll('.brand-preset-card.on').forEach(el => el.classList.remove('on'));
  }

  applyBranding(S.brandDraft);
  updateBrandLivePreview();
  if (['logo_url', 'favicon_url', 'login_bg_url', 'server_hero_url', 'card_cover_url', 'default_avatar_url', 'default_server_icon_url'].includes(field)) {
    const fallbacks = {
      logo_url: '/favicon.svg',
      favicon_url: '/favicon.svg',
      login_bg_url: MINECRAFT_COVER_URL,
      server_hero_url: SERVER_HERO_URL,
      card_cover_url: MINECRAFT_COVER_URL,
      default_avatar_url: DEFAULT_USER_AVATAR_URL,
      default_server_icon_url: '/static/img/default-server-icon.png'
    };
    const thumb = document.querySelector(`[data-brand-thumb="${field}"]`);
    if (thumb) thumb.src = (val || '').trim() || fallbacks[field] || '/favicon.svg';
    const inp = document.querySelector(`input[data-brand-media="${field}"]`);
    if (inp && inp.value !== (val || '')) inp.value = val || '';
  }
  if (['brand_name', 'logo_url', 'client_area_label', 'client_area_enabled', 'client_area_available', 'maintenance_enabled'].includes(field)) {
    const nav = $('#nav');
    if (nav) nav.innerHTML = navHtml();
  }
}

// Panel address is independent of the game/SFTP daemon FQDN.
function renderAdminPanelUrlCard() {
  const current = S.adminPanelUrl;
  const draft = S.panelUrlDraft ?? current?.public_url ?? '';
  const busy = S.panelUrlLoading || S.panelUrlSaving;
  const source = { settings: 'Database override', config: 'Configuration fallback', unset: 'Not configured' };
  return `<div class="card" style="margin-bottom:18px;">
    <div class="row sb" style="gap:10px; flex-wrap:wrap; margin-bottom:6px;">
      <h2 style="margin:0; color:var(--tx);">${ic('link', 18)} Panel address</h2>
      <span class="chip">${S.panelUrlLoading ? 'Loading…' : esc(source[current?.source] || 'Unavailable')}</span>
    </div>
    <p class="mu" style="font-size:13px; line-height:1.6; margin:0 0 16px;">
      The browser-facing address used for passkeys, invitation links, and password-reset links. This is separate from game and SFTP connection addresses.
    </p>
    <form onsubmit="saveAdminPanelUrl(event)">
      <label class="f" for="adm-panel-url">
        <span>Public panel URL</span>
        <input type="url" id="adm-panel-url" value="${esc(draft)}" placeholder="https://panel.example.com" required spellcheck="false" autocomplete="url" aria-describedby="adm-panel-url-help adm-panel-url-error" oninput="onPanelUrlInputChange(this.value)" ${busy ? 'disabled' : ''}>
      </label>
      <p id="adm-panel-url-help" class="mu" style="font-size:12.5px; line-height:1.6;">
        Enter an exact HTTPS origin: scheme, hostname, and optional port, with no path, credentials, query, or fragment. HTTP is allowed only for localhost development. The saved address cannot be cleared.
        Saving applies immediately in SQLite and overrides <code>server.public_url</code>; no restart is needed.
      </p>
      <div style="padding:12px 14px; border:1px solid var(--border); border-radius:8px; font-size:12.5px; line-height:1.6;">
        <b>Prepare the address first.</b> Saving does not configure DNS, HTTPS certificates, or your reverse proxy. Make sure the panel is reachable at this address before saving.
        <br><b>Moving to a new hostname?</b> Remove all passkeys from every account before changing it, then register new passkeys at the new address. Hostname changes are blocked while any passkeys exist; disabling passkeys does not remove them.
      </div>
      <p id="adm-panel-url-error" role="alert" style="color:var(--red); font-size:13px;" ${S.panelUrlError ? '' : 'hidden'}>${esc(S.panelUrlError)}</p>
      <div class="row sb" style="gap:10px; flex-wrap:wrap; margin-top:18px;">
        <span class="mu" style="font-size:12px; overflow-wrap:anywhere;">Current address: ${esc(current?.public_url || (current ? 'Not configured' : 'Unavailable'))}</span>
        <div class="row" style="gap:8px;">
          ${!current ? `<button type="button" class="btn" onclick="loadAdminPanelUrl()" ${busy ? 'disabled' : ''}>${ic('refresh', 14)} Retry</button>` : ''}
          <button type="submit" id="panel-url-save-btn" class="btn pr" ${busy || !current || !draft.trim() ? 'disabled' : ''}>${ic('check', 14)} ${S.panelUrlSaving ? 'Saving…' : 'Save panel address'}</button>
        </div>
      </div>
    </form>
  </div>`;
}

function onPanelUrlInputChange(value) {
  if (S.panelUrlSaving) return;
  S.panelUrlDraft = value;
  S.panelUrlError = '';
  const error = $('#adm-panel-url-error');
  if (error) { error.textContent = ''; error.hidden = true; }
  const button = $('#panel-url-save-btn');
  if (button) button.disabled = S.panelUrlLoading || !S.adminPanelUrl || !value.trim();
}

async function loadAdminPanelUrl(skipRender = false) {
  if (S.panelUrlSaving) return;
  const request = ++S.panelUrlRequest;
  S.panelUrlLoading = true;
  if (!skipRender) R();
  try {
    const result = await api('/api/admin/panel-url');
    if (request !== S.panelUrlRequest) return;
    S.adminPanelUrl = result;
    S.panelUrlError = '';
    // A null draft follows the saved value; never overwrite an unsaved edit.
  } catch (err) {
    if (request === S.panelUrlRequest) S.panelUrlError = err.message || 'Failed to load panel address.';
  } finally {
    if (request === S.panelUrlRequest) {
      S.panelUrlLoading = false;
      if (!skipRender) R();
    }
  }
}

async function saveAdminPanelUrl(event) {
  if (event) event.preventDefault();
  if (S.panelUrlSaving || S.panelUrlLoading || !S.adminPanelUrl) return;
  const draft = S.panelUrlDraft ?? S.adminPanelUrl.public_url;
  if (!draft.trim()) {
    S.panelUrlError = 'Enter a public panel URL. The saved address cannot be cleared.';
    R();
    return;
  }
  S.panelUrlDraft = draft;
  S.panelUrlSaving = true;
  S.panelUrlError = '';
  ++S.panelUrlRequest;
  R();
  try {
    const result = await api('/api/admin/panel-url', {
      method: 'PUT',
      body: JSON.stringify({ public_url: draft.trim() })
    });
    S.adminPanelUrl = result;
    S.panelUrlDraft = null;
    toast('Panel address saved. The new address is active immediately.', 'ok');
  } catch (err) {
    S.panelUrlError = err.message || 'Failed to save panel address.';
    toast(S.panelUrlError, 'bad');
  } finally {
    S.panelUrlSaving = false;
    R();
  }
}

function onDaemonInputChange(field, val) {
  if (!S.daemonDraft) S.daemonDraft = { ...S.brandingDaemon };
  S.daemonDraft[field] = val;
}

function discardBrandDraft() {
  S.brandDraft = { ...DEFAULT_BRANDING, ...S.branding };
  applyBranding(S.branding);
  toast('Reverted unsaved panel look changes');
  R();
}

async function saveBrandingAndDaemon() {
  const btn = $('#brand-save-btn');
  if (btn) setBtnLoading(btn, true, 'Saving...');

  const draft = { ...DEFAULT_BRANDING, ...(S.brandDraft || S.branding) };
  const daemonDraft = { ...S.brandingDaemon, ...(S.daemonDraft || {}) };

  try {
    const res = await api('/api/admin/branding', {
      method: 'PUT',
      body: JSON.stringify({ branding: draft, daemon: daemonDraft })
    });
    if (res && res.branding) {
      S.branding = { ...DEFAULT_BRANDING, ...res.branding };
      S.brandDraft = { ...DEFAULT_BRANDING, ...S.branding };
    } else {
      S.branding = { ...draft };
      S.brandDraft = { ...draft };
    }
    if (res && res.daemon) {
      S.brandingDaemon = { ...S.brandingDaemon, ...res.daemon };
      S.daemonDraft = { ...S.brandingDaemon };
      S.adminDaemonCfg = {
        brandName: S.branding.brand_name || 'NeoPanel',
        fqdn: S.brandingDaemon.fqdn || window.location.hostname || '127.0.0.1',
        sftpPort: S.brandingDaemon.sftp_port || 2022,
        maxUploadMB: S.brandingDaemon.max_upload_mb || 500,
        defaultRamMB: S.brandingDaemon.default_ram_mb || 2048,
        autoRestartPolicy: Boolean(S.brandingDaemon.auto_restart_policy)
      };
      localStorage.setItem('neopanel_admin_cfg', JSON.stringify(S.adminDaemonCfg));
    }
    applyBranding(S.branding);
    toast(res?.message || 'Panel look & branding saved globally!', 'ok');
    loadAdminNodes();
    R();
  } catch (err) {
    toast(err.message || 'Failed to save panel branding', 'bad');
  } finally {
    const curBtn = $('#brand-save-btn');
    if (curBtn) setBtnLoading(curBtn, false);
  }
}

async function resetBrandingDefaults() {
  const ok = await uiConfirm({
    title: 'Reset Panel Look to Defaults',
    message: 'Reset all custom theme presets, colors, glass settings, media URLs, and custom CSS back to factory NeoPanel defaults?',
    confirmText: 'Reset to Defaults',
    cancelText: 'Cancel',
    isDanger: true,
    icon: 'refresh'
  });
  if (!ok) return;

  try {
    const res = await api('/api/admin/branding/reset', { method: 'POST' });
    if (res && res.branding) {
      S.branding = { ...DEFAULT_BRANDING, ...res.branding };
      S.brandDraft = { ...DEFAULT_BRANDING, ...S.branding };
    } else {
      S.branding = { ...DEFAULT_BRANDING };
      S.brandDraft = { ...DEFAULT_BRANDING };
    }
    applyBranding(S.branding);
    toast(res?.message || 'Panel look reset to factory defaults', 'ok');
    R();
  } catch (err) {
    toast(err.message || 'Failed to reset panel branding', 'bad');
  }
}

async function saveAdminDaemonSettings(e) {
  if (e) e.preventDefault();
  const fqdn = $('#adm-fqdn')?.value.trim() || '';
  const sftp_port = parseInt($('#adm-sftp')?.value, 10) || 2022;
  const max_upload_mb = parseInt($('#adm-upload')?.value, 10) || 500;
  const default_ram_mb = parseInt($('#adm-ram')?.value, 10) || 2048;
  const auto_restart_policy = $('#adm-autorestart')?.checked ?? true;

  const daemonPayload = { fqdn, sftp_port, max_upload_mb, default_ram_mb, auto_restart_policy };
  S.daemonDraft = { ...daemonPayload };

  try {
    const res = await api('/api/admin/branding', {
      method: 'PUT',
      body: JSON.stringify({ daemon: daemonPayload })
    });
    if (res && res.daemon) {
      S.brandingDaemon = { ...S.brandingDaemon, ...res.daemon };
      S.daemonDraft = { ...S.brandingDaemon };
    } else {
      S.brandingDaemon = { ...daemonPayload };
    }
    S.adminDaemonCfg = {
      brandName: S.branding.brand_name || 'NeoPanel',
      fqdn: S.brandingDaemon.fqdn || window.location.hostname || '127.0.0.1',
      sftpPort: S.brandingDaemon.sftp_port || 2022,
      maxUploadMB: S.brandingDaemon.max_upload_mb || 500,
      defaultRamMB: S.brandingDaemon.default_ram_mb || 2048,
      autoRestartPolicy: Boolean(S.brandingDaemon.auto_restart_policy)
    };
    localStorage.setItem('neopanel_admin_cfg', JSON.stringify(S.adminDaemonCfg));
    toast('Network and server defaults saved.', 'ok');
    loadAdminNodes();
    R();
  } catch (err) {
    toast(err.message || 'Failed to save daemon settings', 'bad');
  }
}

async function saveAdminTurnstileSettings(e) {
  if (e) e.preventDefault();
  const enabled = Boolean($('#adm-ts-enabled')?.checked);
  const siteKey = $('#adm-ts-site')?.value.trim() || '';
  const secretKey = $('#adm-ts-secret')?.value.trim() || '';

  try {
    const res = await api('/api/admin/turnstile', {
      method: 'POST',
      body: JSON.stringify({
        enabled,
        site_key: siteKey,
        secret_key: secretKey
      })
    });
    S.adminTurnstile = {
      enabled: Boolean(res.enabled),
      site_key: res.site_key || siteKey,
      has_secret_key: Boolean(res.has_secret_key)
    };
    S.turnstileEnabled = Boolean(res.enabled);
    S.turnstileSiteKey = res.enabled ? (res.site_key || siteKey) : '';
    toast(res.message || 'Turnstile settings updated', 'ok');
    R();
  } catch (err) {
    toast(err.message || 'Failed to save Turnstile settings', 'bad');
  }
}

// --- Cloudflare Turnstile Explicit SPA Integration ---
function mountTurnstileWidget(selector) {
  S.turnstileToken = '';
  if (!S.turnstileEnabled || !S.turnstileSiteKey) return;
  const container = $(selector);
  if (!container) return;

  container.innerHTML = '<div class="turnstile-loading">Loading Cloudflare security check...</div>';

  let attempts = 0;
  const tryRender = () => {
    const el = $(selector);
    if (!el) return;
    if (window.turnstile && typeof window.turnstile.render === 'function') {
      el.innerHTML = '';
      if (S.turnstileWidgetId !== null && typeof window.turnstile.remove === 'function') {
        try { window.turnstile.remove(S.turnstileWidgetId); } catch (e) {}
        S.turnstileWidgetId = null;
      }
      const widgetTheme = S.acct?.mode === 'Light' ? 'light' : 'dark';
      S.turnstileWidgetId = window.turnstile.render(el, {
        sitekey: S.turnstileSiteKey,
        theme: widgetTheme,
        callback: (token) => {
          S.turnstileToken = token || '';
          const err = $('#lerr');
          if (err && err.textContent.includes('Turnstile')) err.textContent = '';
        },
        'expired-callback': () => {
          S.turnstileToken = '';
        },
        'error-callback': () => {
          S.turnstileToken = '';
        }
      });
    } else if (++attempts < 100) {
      setTimeout(tryRender, 100);
    } else {
      el.innerHTML = '<div class="turnstile-loading">Unable to load Cloudflare Turnstile. Please check your network or disable adblockers.</div>';
    }
  };
  tryRender();
}

function resetTurnstileWidget() {
  S.turnstileToken = '';
  if (S.turnstileEnabled && window.turnstile && S.turnstileWidgetId !== null) {
    try { window.turnstile.reset(S.turnstileWidgetId); } catch (e) {}
  }
}

// --- Auth Flow & Setup Screen ---
function applyLoginOverlayBackground(el) {
  if (!el) return;
  const bg = (getBrand('login_bg_url') || '').trim();
  if (bg) {
    el.style.backgroundImage = `linear-gradient(rgba(6, 10, 18, 0.72), rgba(6, 10, 18, 0.85)), url('${bg.replace(/'/g, '%27')}')`;
    el.style.backgroundSize = 'cover';
    el.style.backgroundPosition = 'center';
  }
}

// --- Branding Asset Upload & Maintenance Mode Helpers ---
let _brandUploadField = '';
let _brandUploadBtn = null;

function triggerBrandingUpload(field, btnEl) {
  _brandUploadField = field || '';
  _brandUploadBtn = btnEl || null;
  const inp = $('#brand-file-input');
  if (!inp) return;
  inp.value = '';
  inp.click();
}

async function handleBrandingFileSelect(inputEl) {
  const file = inputEl?.files?.[0];
  const field = _brandUploadField || '';
  const btnEl = _brandUploadBtn;
  _brandUploadField = '';
  _brandUploadBtn = null;
  if (inputEl) inputEl.value = '';
  if (!file || !field) return;

  if (file.size > 12 * 1024 * 1024) {
    toast('Image file exceeds 12 MiB upload limit', 'bad');
    return;
  }

  if (btnEl) setBtnLoading(btnEl, true, 'Uploading...');
  try {
    const fd = new FormData();
    fd.append('file', file);
    fd.append('field', field);
    const res = await fetch('/api/admin/branding/upload', {
      method: 'POST',
      headers: { 'Authorization': 'Bearer ' + S.token },
      body: fd
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok || !data.url) {
      toast(data.error || 'Failed to upload branding image', 'bad');
      return;
    }
    onBrandInputChange(field, data.url);
    toast(`Uploaded ${data.filename || file.name} — remember to click Save Global Look`, 'ok');
  } catch (err) {
    toast(err.message || 'Branding upload failed', 'bad');
  } finally {
    if (btnEl && document.body.contains(btnEl)) setBtnLoading(btnEl, false);
  }
}

function clearBrandingMedia(field) {
  const resetMap = {
    logo_url: '',
    favicon_url: '',
    login_bg_url: '',
    server_hero_url: SERVER_HERO_URL,
    card_cover_url: MINECRAFT_COVER_URL,
    default_avatar_url: DEFAULT_USER_AVATAR_URL,
    default_server_icon_url: '/static/img/default-server-icon.png'
  };
  const nextVal = Object.prototype.hasOwnProperty.call(resetMap, field) ? resetMap[field] : '';
  onBrandInputChange(field, nextVal);
  toast('Reset media field to default (save to persist)');
}

function toggleMaintPreview(val) {
  S.maintPreview = (val === undefined) ? !S.maintPreview : Boolean(val);
  R();
}

function enterAdminLoginFromMaint() {
  S.maintAdminSignInOpen = true;
  $('#login')?.remove();
  showLoginScreen();
}

// --- Official Open Panel Modding System (window.NP / window.NeoPanelMod) ---
const MOD_TEMPLATES = {
  'custom-page': {
    id: 'cluster-tools',
    name: 'Cluster Tools Mod',
    version: '1.0.0',
    author: 'Community',
    description: 'Adds a dedicated /mod/cluster-tools sidebar page, a top dashboard banner, and a Ctrl+K command.',
    js: `// Official NeoPanel Mod SDK v2 — Custom Page + Slot + Command Palette
NP.registerPage({
  id: 'cluster-tools',
  title: 'Cluster Tools',
  icon: 'terminal',
  scope: 'both', // 'global', 'server', or 'both'
  render(root, ctx) {
    const servers = ctx.servers || [];
    const online = servers.filter(s => (s.stats?.status || s.status) === 'running').length;
    root.innerHTML = \`
      <div class="card">
        <div class="row sb" style="margin-bottom:12px;">
          <div>
            <h1 style="margin:0; font-size:22px;">Cluster Tools Extension</h1>
            <div class="mu" style="font-size:13px;">Powered by window.NP SDK v\${NP.version}</div>
          </div>
          <span class="badge online"><i class="dot"></i>\${online} / \${servers.length} Online</span>
        </div>
        <p class="mu" style="font-size:13.5px; line-height:1.6;">
          You can call authenticated APIs via <code>NP.api(...)</code>, render into extension slots, and store preferences in <code>NP.storage('cluster-tools')</code>.
        </p>
        <div class="row" style="gap:10px; margin-top:14px;">
          <button class="btn pr" id="mod-ping-btn">Check System Info</button>
          <button class="btn" onclick="NP.navigate('dash')">Back to Dashboard</button>
        </div>
        <pre id="mod-ping-out" class="mono" style="margin-top:14px; font-size:12px; color:var(--tx-muted);"></pre>
      </div>\`;
    const btn = root.querySelector('#mod-ping-btn');
    if (btn) {
      btn.onclick = async () => {
        const info = await NP.api('/api/system');
        root.querySelector('#mod-ping-out').textContent = JSON.stringify(info, null, 2);
      };
    }
  }
});

NP.registerSlot('dashboard:top', (root, ctx) => {
  root.innerHTML = \`<div class="card" style="padding:12px 16px; border-color:var(--border-accent);">
    <div class="row sb" style="gap:10px; flex-wrap:wrap;">
      <span><b>Cluster Tools Active</b> &mdash; Managing \${ctx.servers.length} server instance(s)</span>
      <button class="btn sm pr" onclick="NP.navigate('/mod/cluster-tools')">Open Extension Page</button>
    </div>
  </div>\`;
});

NP.addPaletteCommand({
  id: 'open-cluster-tools',
  group: 'Extensions',
  label: 'Open Cluster Tools',
  subtitle: 'Jump to Cluster Tools mod page',
  action: () => NP.navigate('/mod/cluster-tools')
});`,
    css: `/* Optional CSS scoped cleanly by your mod */
#np-mod-page-root[data-mod-page="cluster-tools"] .card {
  border-color: var(--border-accent);
}`
  },
  'console-widget': {
    id: 'tps-quickbar',
    name: 'Console Quick Actions Bar',
    version: '1.0.0',
    author: 'Community',
    description: 'Injects custom one-click admin shortcuts right beneath console live charts.',
    js: `// Adds custom operator buttons below the Server Console charts
NP.registerSlot('console:below-charts', (root, ctx) => {
  if (!ctx.server) return;
  root.innerHTML = \`
    <div class="card" style="padding:12px 16px;">
      <div class="row sb" style="gap:10px; flex-wrap:wrap;">
        <span style="font-size:13px; font-weight:600;">Quick Operator Macros (\${ctx.server.name})</span>
        <div class="row" style="gap:8px; flex-wrap:wrap;">
          <button class="btn sm" data-cmd="save-all">Save World</button>
          <button class="btn sm" data-cmd="list">List Players</button>
          <button class="btn sm pr" data-cmd="timings paste">Timings Report</button>
        </div>
      </div>
    </div>\`;
  root.querySelectorAll('button[data-cmd]').forEach(btn => {
    btn.onclick = async () => {
      await NP.srvApi('/command', {
        method: 'POST',
        body: JSON.stringify({ command: btn.dataset.cmd })
      });
      NP.toast('Sent command: ' + btn.dataset.cmd, 'ok');
    };
  });
});`,
    css: ``
  },
  'custom-theme': {
    id: 'neon-glass-theme',
    name: 'Neon Glass CSS Theme Pack',
    version: '1.0.0',
    author: 'Community',
    description: 'Pure CSS visual enhancement highlighting active cards and console edges.',
    js: ``,
    css: `/* Neon Glass Style Extension */
.card {
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.28), inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.sift-server-card:hover {
  border-color: var(--ac) !important;
  box-shadow: 0 12px 34px var(--ac-glow) !important;
}`
  }
};

function initNeoPanelModSdk() {
  if (window.NP && window.NP._initialized) return window.NP;

  const normalizeId = (raw) => String(raw || '').trim().toLowerCase().replace(/[^a-z0-9_-]/g, '-').replace(/^-+|-+$/g, '').slice(0, 48);

  const sdk = {
    version: '2.0.0',
    _initialized: true,
    _mods: [],
    _errors: [],
    _loadedModSigs: new Map(),
    _pages: new Map(),
    _sidebarItems: [],
    _slots: new Map(),
    _paletteCommands: [],
    _listeners: new Map(),
    _modRegistrations: new Map(),
    _currentLoadingModId: null,

    _trackReg(modId, kind, key) {
      if (!modId) return;
      let rec = this._modRegistrations.get(modId);
      if (!rec) {
        rec = { pages: new Set(), sidebar: new Set(), slots: [], palette: new Set(), css: new Set() };
        this._modRegistrations.set(modId, rec);
      }
      if (kind === 'slot') rec.slots.push(key);
      else if (rec[kind]) rec[kind].add(key);
    },

    clearModRegistrations(modId) {
      const rec = this._modRegistrations.get(modId);
      if (!rec) return;
      rec.pages.forEach(pid => this._pages.delete(pid));
      this._sidebarItems = this._sidebarItems.filter(it => it.modId !== modId && !rec.sidebar.has(it.id));
      for (const [slotName, list] of this._slots.entries()) {
        this._slots.set(slotName, list.filter(it => it.modId !== modId));
      }
      this._paletteCommands = this._paletteCommands.filter(c => c.modId !== modId && !rec.palette.has(c.id));
      rec.css.forEach(cid => this.removeCSS(cid));
      this._modRegistrations.delete(modId);
    },

    on(event, fn) {
      if (!event || typeof fn !== 'function') return () => {};
      if (!this._listeners.has(event)) this._listeners.set(event, []);
      this._listeners.get(event).push(fn);
      return () => this.off(event, fn);
    },

    off(event, fn) {
      const list = this._listeners.get(event);
      if (!list) return;
      this._listeners.set(event, list.filter(f => f !== fn));
    },

    _emit(event, payload) {
      const list = this._listeners.get(event);
      if (!list || !list.length) return;
      for (const fn of list) {
        try {
          fn(payload, this.getState());
        } catch (err) {
          this._recordError(this._currentLoadingModId || 'event:' + event, err);
        }
      }
    },

    _recordError(modId, err) {
      const msg = err?.message || String(err || 'Unknown mod error');
      this._errors.push({ modId: modId || 'unknown', message: msg, stack: err?.stack || '', time: Date.now() });
      if (this._errors.length > 30) this._errors.shift();
      console.error(`[NeoPanel Mod:${modId}]`, err);
    },

    getState() {
      return {
        view: S.view,
        server: curServer(),
        servers: S.servers || [],
        nodes: S.nodes || [],
        user: S.user || null,
        branding: S.branding || DEFAULT_BRANDING,
        theme: document.documentElement.dataset.theme || 'dark',
        safeMode: Boolean(S.modsSafeMode)
      };
    },

    registerPage(idOrOpts, maybeRender) {
      const opts = (typeof idOrOpts === 'object' && idOrOpts)
        ? idOrOpts
        : { id: idOrOpts, render: maybeRender };
      const cleanId = normalizeId(opts.id);
      if (!cleanId || typeof opts.render !== 'function') return null;
      const modId = opts.modId || this._currentLoadingModId || 'inline';
      const entry = {
        id: cleanId,
        title: opts.title || cleanId,
        icon: opts.icon || 'puzzle',
        scope: opts.scope || 'both',
        render: opts.render,
        modId
      };
      this._pages.set(cleanId, entry);
      this._trackReg(modId, 'pages', cleanId);
      if (opts.sidebar !== false) {
        this.addSidebarItem({
          id: `page-${cleanId}`,
          label: opts.sidebarLabel || entry.title,
          icon: entry.icon,
          scope: entry.scope,
          pageId: cleanId,
          modId
        });
      }
      return entry;
    },

    addSidebarItem(opts) {
      if (!opts || !opts.label) return null;
      const modId = opts.modId || this._currentLoadingModId || 'inline';
      const itemId = opts.id || `${modId}:${opts.pageId || opts.label}`;
      const item = {
        id: itemId,
        label: opts.label,
        icon: opts.icon || 'puzzle',
        scope: opts.scope || 'both',
        pageId: opts.pageId ? normalizeId(opts.pageId) : '',
        url: opts.url || '',
        onClick: typeof opts.onClick === 'function' ? opts.onClick : null,
        modId
      };
      const existingIdx = this._sidebarItems.findIndex(x => x.id === itemId);
      if (existingIdx >= 0) this._sidebarItems[existingIdx] = item;
      else this._sidebarItems.push(item);
      this._trackReg(modId, 'sidebar', itemId);
      return item;
    },

    _execSidebarItem(id, e) {
      const item = this._sidebarItems.find(x => x.id === id);
      if (!item) return false;
      if (e) e.preventDefault();
      try {
        if (item.onClick) item.onClick(this.getState(), e);
        else if (item.pageId) this.navigate(`/mod/${item.pageId}`);
        else if (item.url) applyRoute(item.url);
      } catch (err) {
        this._recordError(item.modId, err);
        toast(`Mod action failed: ${err.message}`, 'bad');
      }
      return false;
    },

    registerSlot(slotName, fnOrOpts) {
      const opts = typeof fnOrOpts === 'function' ? { render: fnOrOpts } : (fnOrOpts || {});
      if (!slotName || typeof opts.render !== 'function') return null;
      const modId = opts.modId || this._currentLoadingModId || 'inline';
      const entry = {
        id: opts.id || `${modId}:${slotName}:${Math.random().toString(36).slice(2, 7)}`,
        slot: slotName,
        order: Number(opts.order) || 10,
        render: opts.render,
        modId
      };
      if (!this._slots.has(slotName)) this._slots.set(slotName, []);
      const list = this._slots.get(slotName).filter(x => x.id !== entry.id);
      list.push(entry);
      list.sort((a, b) => a.order - b.order);
      this._slots.set(slotName, list);
      this._trackReg(modId, 'slot', entry.id);
      return entry;
    },

    addPaletteCommand(opts) {
      if (!opts || !opts.label || typeof opts.action !== 'function') return null;
      const modId = opts.modId || this._currentLoadingModId || 'inline';
      const id = opts.id || `${modId}:${opts.label}`;
      const cmd = {
        id,
        group: opts.group || 'Extensions',
        label: opts.label,
        subtitle: opts.subtitle || 'Mod command',
        action: opts.action,
        modId
      };
      const idx = this._paletteCommands.findIndex(x => x.id === id);
      if (idx >= 0) this._paletteCommands[idx] = cmd;
      else this._paletteCommands.push(cmd);
      this._trackReg(modId, 'palette', id);
      return cmd;
    },

    navigate(target, opts = {}) {
      if (!target) return;
      const s = String(target);
      if (s.startsWith('/')) {
        applyRoute(s, opts);
        return;
      }
      applyRoute(routeUrl(s, opts), opts);
    },

    api(path, opts) {
      return api(path, opts);
    },

    srvApi(subPath, opts) {
      const srv = curServer();
      if (!srv) return Promise.reject(new Error('No active server selected'));
      const cleanSub = subPath.startsWith('/') ? subPath : '/' + subPath;
      return api(srvApiPath(srv, cleanSub), opts);
    },

    toast(msg, kind) {
      toast(msg, kind);
    },

    modal: {
      alert: (opts, t) => uiAlert(opts, t),
      confirm: (opts, t) => uiConfirm(opts, t),
      prompt: (opts, d, t) => uiPrompt(opts, d, t),
      close: () => closeModal()
    },

    injectCSS(id, cssText) {
      const cid = normalizeId(id) || 'mod-style';
      const modId = this._currentLoadingModId || cid;
      let styleEl = document.querySelector(`style[data-np-mod-css="${cid}"]`);
      if (!styleEl) {
        styleEl = document.createElement('style');
        styleEl.setAttribute('data-np-mod-css', cid);
        document.head.appendChild(styleEl);
      }
      styleEl.textContent = String(cssText || '');
      this._trackReg(modId, 'css', cid);
      return styleEl;
    },

    removeCSS(id) {
      const cid = normalizeId(id) || 'mod-style';
      document.querySelectorAll(`style[data-np-mod-css="${cid}"]`).forEach(el => el.remove());
    },

    storage(modId) {
      const prefix = `neopanel_mod_${normalizeId(modId) || 'default'}_`;
      return {
        get(key, fallback = null) {
          try {
            const raw = localStorage.getItem(prefix + key);
            return raw === null ? fallback : JSON.parse(raw);
          } catch (e) {
            return fallback;
          }
        },
        set(key, val) {
          try {
            localStorage.setItem(prefix + key, JSON.stringify(val));
          } catch (e) {}
        },
        remove(key) {
          try {
            localStorage.removeItem(prefix + key);
          } catch (e) {}
        }
      };
    }
  };

  window.NP = sdk;
  window.NeoPanelMod = sdk;
  return sdk;
}

initNeoPanelModSdk();

function mountPanelModHooks() {
  const sdk = initNeoPanelModSdk();
  if (S.modsSafeMode) return;
  const ctx = sdk.getState();

  document.querySelectorAll('.np-slot-container[data-np-slot]').forEach(container => {
    const slotName = container.getAttribute('data-np-slot');
    const handlers = sdk._slots.get(slotName) || [];
    container.innerHTML = '';
    if (!handlers.length) return;

    handlers.forEach(entry => {
      const mountBox = document.createElement('div');
      mountBox.dataset.npMod = entry.modId;
      container.appendChild(mountBox);
      try {
        const res = entry.render(mountBox, ctx);
        if (typeof res === 'string') {
          mountBox.innerHTML = res;
        } else if (res && res.nodeType) {
          mountBox.innerHTML = '';
          mountBox.appendChild(res);
        } else if (res && typeof res.then === 'function') {
          res.then(val => {
            if (!document.body.contains(mountBox)) return;
            if (typeof val === 'string') mountBox.innerHTML = val;
            else if (val && val.nodeType) {
              mountBox.innerHTML = '';
              mountBox.appendChild(val);
            }
            enhanceSelects(mountBox);
          }).catch(err => {
            sdk._recordError(entry.modId, err);
            mountBox.innerHTML = `<div class="np-maint-banner"><span>Mod <b>${esc(entry.modId)}</b> slot error: ${esc(err.message)}</span></div>`;
          });
        }
      } catch (err) {
        sdk._recordError(entry.modId, err);
        mountBox.innerHTML = `<div class="np-maint-banner"><span>Mod <b>${esc(entry.modId)}</b> error: ${esc(err.message)}</span></div>`;
      }
    });
    enhanceSelects(container);
  });

  const pageRoot = $('#np-mod-page-root');
  if (pageRoot && S.view === 'mod') {
    const pageId = pageRoot.getAttribute('data-mod-page') || S.modPageId;
    const page = sdk._pages.get(pageId);
    if (page) {
      pageRoot.innerHTML = '';
      try {
        const res = page.render(pageRoot, ctx);
        if (typeof res === 'string') {
          pageRoot.innerHTML = res;
        } else if (res && res.nodeType) {
          pageRoot.innerHTML = '';
          pageRoot.appendChild(res);
        } else if (res && typeof res.then === 'function') {
          res.then(val => {
            if (!document.body.contains(pageRoot)) return;
            if (typeof val === 'string') pageRoot.innerHTML = val;
            else if (val && val.nodeType) {
              pageRoot.innerHTML = '';
              pageRoot.appendChild(val);
            }
            enhanceSelects(pageRoot);
          }).catch(err => {
            sdk._recordError(page.modId, err);
            pageRoot.innerHTML = `<div class="card"><div class="np-status-code err">${ic('alert', 14)} MOD ERROR</div><h2>${esc(page.title)}</h2><pre class="mono">${esc(err.message)}</pre></div>`;
          });
        }
      } catch (err) {
        sdk._recordError(page.modId, err);
        pageRoot.innerHTML = `<div class="card"><div class="np-status-code err">${ic('alert', 14)} MOD ERROR</div><h2>${esc(page.title)}</h2><pre class="mono">${esc(err.message)}</pre></div>`;
      }
      enhanceSelects(pageRoot);
    }
  }
}

async function loadRuntimeMods(forceReload = false) {
  const sdk = initNeoPanelModSdk();
  const params = new URLSearchParams(window.location.search);
  if (params.get('safemode') === '1' || S.modsSafeMode) {
    S.modsSafeMode = true;
    S.mods = [];
    document.querySelectorAll('link[data-np-mod-style]').forEach(el => el.remove());
    return [];
  }

  try {
    const res = await fetch('/api/mods', { cache: 'no-store' });
    if (!res.ok) return [];
    const data = await res.json();
    const list = Array.isArray(data?.mods) ? data.mods : [];
    S.mods = list;
    sdk._mods = list;

    const activeIds = new Set(list.map(m => m.id));
    for (const prevId of Array.from(sdk._loadedModSigs.keys())) {
      if (!activeIds.has(prevId)) {
        sdk.clearModRegistrations(prevId);
        sdk._loadedModSigs.delete(prevId);
      }
    }

    document.querySelectorAll('link[data-np-mod-style]').forEach(el => {
      if (!activeIds.has(el.getAttribute('data-np-mod-style'))) el.remove();
    });

    for (const mod of list) {
      const sig = `${mod.updated_at || ''}|${mod.entry_js || ''}|${mod.entry_css || ''}`;
      if (mod.has_css && mod.css_url) {
        let link = document.querySelector(`link[data-np-mod-style="${mod.id}"]`);
        const href = `${mod.css_url}?v=${encodeURIComponent(mod.updated_at || '1')}`;
        if (!link) {
          link = document.createElement('link');
          link.rel = 'stylesheet';
          link.setAttribute('data-np-mod-style', mod.id);
          document.head.appendChild(link);
        }
        if (forceReload || link.getAttribute('href') !== href) {
          link.href = href;
        }
      } else {
        document.querySelectorAll(`link[data-np-mod-style="${mod.id}"]`).forEach(el => el.remove());
      }

      if (mod.has_js && mod.js_url) {
        if (!forceReload && sdk._loadedModSigs.get(mod.id) === sig) continue;
        sdk.clearModRegistrations(mod.id);
        sdk._currentLoadingModId = mod.id;
        try {
          const urlWithBust = `${mod.js_url}?v=${encodeURIComponent(mod.updated_at || String(Date.now()))}`;
          const codeRes = await fetch(urlWithBust, { cache: 'no-store' });
          if (!codeRes.ok) throw new Error(`HTTP ${codeRes.status} loading ${mod.js_url}`);
          const code = await codeRes.text();
          const looksModule = /\bexport\s+(default|const|let|var|function|async)\b/.test(code) || /(?:^|\n)\s*import\s+/.test(code);
          if (looksModule) {
            const modNs = await import(/* @vite-ignore */ urlWithBust + (forceReload ? '&r=' + Date.now() : ''));
            if (modNs && typeof modNs.default === 'function') {
              await modNs.default(sdk, sdk.getState());
            }
          } else {
            const runner = new Function('NP', 'NeoPanelMod', `${code}\n//# sourceURL=${mod.js_url}`);
            runner(sdk, sdk);
          }
          sdk._loadedModSigs.set(mod.id, sig);
        } catch (err) {
          sdk._recordError(mod.id, err);
          toast(`Mod "${mod.name || mod.id}" failed to load: ${err.message}`, 'bad');
        } finally {
          sdk._currentLoadingModId = null;
        }
      } else {
        if (sdk._loadedModSigs.has(mod.id)) {
          sdk.clearModRegistrations(mod.id);
          sdk._loadedModSigs.delete(mod.id);
        }
      }
    }

    return list;
  } catch (err) {
    console.warn('Failed to fetch runtime mods:', err);
    return [];
  }
}

// --- Admin Mods & Extensions Workspace ---
function getBlankModDraft() {
  return {
    isNew: true,
    id: '',
    name: '',
    version: '1.0.0',
    author: S.user?.username || 'Admin',
    description: '',
    enabled: true,
    entry_js: 'mod.js',
    entry_css: 'mod.css',
    js_code: '',
    css_code: ''
  };
}

function startNewModDraft() {
  S.selectedModId = null;
  S.modDraft = getBlankModDraft();
  R();
}

function selectAdminMod(modId, skipRender = false) {
  const found = (S.adminMods || []).find(m => m.id === modId);
  if (!found) return;
  S.selectedModId = found.id;
  S.modDraft = {
    isNew: false,
    id: found.id,
    name: found.name || found.id,
    version: found.version || '1.0.0',
    author: found.author || '',
    description: found.description || '',
    enabled: found.enabled !== false,
    entry_js: found.entry_js || 'mod.js',
    entry_css: found.entry_css || 'mod.css',
    js_code: found.js_code || '',
    css_code: found.css_code || ''
  };
  if (!skipRender) R();
}

function applyModTemplate(key) {
  const tpl = MOD_TEMPLATES[key];
  if (!tpl) return;
  const cur = S.modDraft || getBlankModDraft();
  const isNew = cur.isNew !== false && !S.selectedModId;
  S.modDraft = {
    ...cur,
    id: isNew ? tpl.id : (cur.id || tpl.id),
    name: isNew || !cur.name ? tpl.name : cur.name,
    version: cur.version || tpl.version,
    author: cur.author || tpl.author,
    description: cur.description || tpl.description,
    enabled: true,
    js_code: tpl.js || '',
    css_code: tpl.css || ''
  };
  toast(`Applied starter template: ${tpl.name}`);
  R();
}

function onModDraftInput(field, val) {
  if (!S.modDraft) S.modDraft = getBlankModDraft();
  S.modDraft[field] = val;
}

async function loadAdminMods(skipRender = false) {
  try {
    const res = await api('/api/admin/mods');
    if (res && Array.isArray(res.mods)) {
      S.adminMods = res.mods;
      if (res.mods_dir) S.modsDir = res.mods_dir;
    }
    if (!skipRender) R();
  } catch (err) {
    toast(err.message || 'Failed to load admin mods', 'bad');
  }
}

async function saveAdminMod(e) {
  if (e) e.preventDefault();
  const d = S.modDraft || getBlankModDraft();
  const rawId = String(d.id || d.name || '').trim().toLowerCase().replace(/[^a-z0-9_-]+/g, '-').replace(/^-+|-+$/g, '');
  if (!rawId || rawId.length < 2) {
    toast('Mod ID must be at least 2 characters (a-z, 0-9, -, _)', 'bad');
    return;
  }

  const btn = $('#mod-save-btn');
  if (btn) setBtnLoading(btn, true, 'Saving Mod...');
  try {
    const endpoint = d.isNew ? '/api/admin/mods' : `/api/admin/mods/${encodeURIComponent(rawId)}`;
    const method = d.isNew ? 'POST' : 'PUT';
    const res = await api(endpoint, {
      method,
      body: JSON.stringify({
        id: rawId,
        name: (d.name || rawId).trim(),
        version: (d.version || '1.0.0').trim(),
        author: (d.author || 'Admin').trim(),
        description: (d.description || '').trim(),
        enabled: Boolean(d.enabled),
        entry_js: (d.entry_js || 'mod.js').trim(),
        entry_css: (d.entry_css || 'mod.css').trim(),
        js_code: String(d.js_code || ''),
        css_code: String(d.css_code || '')
      })
    });
    await loadAdminMods(true);
    selectAdminMod(res?.mod?.id || rawId, true);
    await loadRuntimeMods(true);
    toast(res?.message || `Saved mod "${rawId}" and applied live!`, 'ok');
    R();
  } catch (err) {
    toast(err.message || 'Failed to save mod', 'bad');
  } finally {
    const curBtn = $('#mod-save-btn');
    if (curBtn) setBtnLoading(curBtn, false);
  }
}

async function toggleAdminModEnabled(modId, enabled) {
  const target = (S.adminMods || []).find(m => m.id === modId);
  if (!target) return;
  try {
    await api(`/api/admin/mods/${encodeURIComponent(modId)}`, {
      method: 'PUT',
      body: JSON.stringify({
        id: target.id,
        name: target.name,
        version: target.version,
        author: target.author,
        description: target.description || '',
        enabled: Boolean(enabled)
      })
    });
    if (S.modDraft && S.modDraft.id === modId) {
      S.modDraft.enabled = Boolean(enabled);
    }
    await loadAdminMods(true);
    await loadRuntimeMods(true);
    toast(`Mod "${target.name || modId}" ${enabled ? 'enabled' : 'disabled'}`, 'ok');
    R();
  } catch (err) {
    toast(err.message || 'Failed to update mod state', 'bad');
  }
}

async function deleteAdminMod(modId, name) {
  const ok = await uiConfirm({
    title: 'Delete Panel Extension',
    message: `Delete extension <b>${esc(name || modId)}</b> (<code>${esc(modId)}</code>) from disk? Any custom pages, slots, or stylesheets it added will be unloaded.`,
    confirmText: 'Delete Mod',
    cancelText: 'Cancel',
    isDanger: true,
    icon: 'trash'
  });
  if (!ok) return;

  try {
    await api(`/api/admin/mods/${encodeURIComponent(modId)}`, { method: 'DELETE' });
    if (S.selectedModId === modId) {
      S.selectedModId = null;
      S.modDraft = getBlankModDraft();
    }
    await loadAdminMods(true);
    await loadRuntimeMods(true);
    toast(`Removed mod "${modId}"`, 'ok');
    R();
  } catch (err) {
    toast(err.message || 'Failed to delete mod', 'bad');
  }
}

function triggerModPackageUpload() {
  const inp = $('#mod-pkg-upload');
  if (!inp) return;
  inp.value = '';
  inp.click();
}

async function handleModPackageUpload(inputEl) {
  const file = inputEl?.files?.[0];
  if (inputEl) inputEl.value = '';
  if (!file) return;

  if (file.size > 20 * 1024 * 1024) {
    toast('Mod package exceeds 20 MiB limit', 'bad');
    return;
  }

  const fd = new FormData();
  fd.append('file', file);
  toast(`Installing mod package ${file.name}...`);
  try {
    const res = await fetch('/api/admin/mods/upload', {
      method: 'POST',
      headers: { 'Authorization': 'Bearer ' + S.token },
      body: fd
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      toast(data.error || 'Failed to upload mod package', 'bad');
      return;
    }
    await loadAdminMods(true);
    if (data.mod?.id) selectAdminMod(data.mod.id, true);
    await loadRuntimeMods(true);
    toast(data.message || `Installed mod package ${file.name}!`, 'ok');
    R();
  } catch (err) {
    toast(err.message || 'Failed to upload mod package', 'bad');
  }
}

function renderAdminModsTabHtml() {
  const mods = Array.isArray(S.adminMods) ? S.adminMods : [];
  const draft = S.modDraft || getBlankModDraft();
  const isNew = draft.isNew !== false && !S.selectedModId;
  const errCount = Array.isArray(window.NP?._errors) ? window.NP._errors.length : 0;
  const pageCount = window.NP?._pages?.size || 0;
  const slotCount = Array.from(window.NP?._slots?.values() || []).reduce((acc, a) => acc + a.length, 0);

  return `
    <div class="card" style="margin-bottom: 16px;">
      <div class="row sb" style="gap: 14px; flex-wrap: wrap; align-items: center;">
        <div>
          <div class="row" style="gap: 8px; flex-wrap: wrap;">
            <h2 style="margin: 0; color: var(--tx);">Panel Extensions &amp; Modding Engine</h2>
            <span class="badge online"><i class="dot"></i>window.NP SDK v${esc(window.NP?.version || '2.0.0')}</span>
            ${S.modsSafeMode ? `<span class="badge starting"><i class="dot"></i>Safe Mode Active</span>` : ''}
          </div>
          <p class="mu" style="font-size: 12.5px; margin: 4px 0 0;">
            Extend NeoPanel beyond presets: drop bundles into <code class="mono">${esc(S.modsDir || './data/mods')}</code>, upload <code class="mono">.zip/.js/.css</code>, or author extensions live below. Active hooks: <b>${pageCount}</b> page(s), <b>${slotCount}</b> slot(s)${errCount ? `, <span style="color:#f87171;">${errCount} error(s)</span>` : ''}.
          </p>
        </div>
        <div class="row" style="gap: 8px; flex-wrap: wrap;">
          <input type="file" id="mod-pkg-upload" accept=".zip,.js,.mjs,.css" style="display:none" onchange="handleModPackageUpload(this)">
          <button type="button" class="btn sm" onclick="triggerModPackageUpload()">${ic('upload', 13)} Upload .zip / .js / .css</button>
          <button type="button" class="btn sm" onclick="loadRuntimeMods(true).then(() => { R(); toast('Reloaded runtime mods', 'ok'); })">${ic('refresh', 13)} Reload Runtime</button>
          <button type="button" class="btn sm pr" onclick="startNewModDraft()">${ic('plus', 13)} New Mod</button>
        </div>
      </div>
    </div>

    <div class="mod-layout">
      <div class="card">
        <div class="row sb" style="margin-bottom: 12px;">
          <h3 style="margin: 0; font-size: 15px;">Installed Mods (${mods.length})</h3>
          <a href="${esc(S.modsSafeMode ? '/admin/mods' : '/admin/mods?safemode=1')}" class="mu" style="font-size: 11.5px; text-decoration: none;" title="Append ?safemode=1 to recover from broken mods">${S.modsSafeMode ? 'Exit Safe Mode' : 'Safe Mode (?safemode=1)'}</a>
        </div>
        <div class="mod-list">
          ${mods.length ? mods.map(m => {
            const isSel = S.selectedModId === m.id && !isNew;
            const modPageExists = window.NP?._pages?.has(m.id);
            return `<div class="mod-card ${isSel ? 'on' : ''}" onclick="selectAdminMod('${esc(m.id)}')">
              <div class="row sb" style="gap: 8px; align-items: flex-start;">
                <div style="min-width: 0;">
                  <div class="row" style="gap: 6px; flex-wrap: wrap;">
                    <b style="color: var(--tx); font-size: 13.5px;">${esc(m.name || m.id)}</b>
                    <span class="chip mono" style="font-size: 10.5px; padding: 1px 6px;">v${esc(m.version || '1.0.0')}</span>
                  </div>
                  <div class="mono mu" style="font-size: 11px; margin-top: 2px;">${esc(m.id)}${m.author ? ` &bull; ${esc(m.author)}` : ''}</div>
                </div>
                <label class="sw" onclick="event.stopPropagation();" title="${m.enabled ? 'Enabled' : 'Disabled'}">
                  <input type="checkbox" ${m.enabled ? 'checked' : ''} onchange="toggleAdminModEnabled('${esc(m.id)}', this.checked)">
                  <i></i>
                </label>
              </div>
              ${m.description ? `<div class="mu" style="font-size: 12px; margin-top: 6px; line-height: 1.4;">${esc(m.description)}</div>` : ''}
              <div class="row sb" style="margin-top: 8px; gap: 6px; flex-wrap: wrap;">
                <div class="row" style="gap: 5px;">
                  ${m.has_js ? '<span class="chip" style="font-size: 10px; padding: 1px 6px;">JS</span>' : ''}
                  ${m.has_css ? '<span class="chip" style="font-size: 10px; padding: 1px 6px;">CSS</span>' : ''}
                  ${!m.enabled ? '<span class="chip" style="font-size: 10px; padding: 1px 6px; color: #f87171;">Disabled</span>' : ''}
                </div>
                <div class="row" style="gap: 6px;" onclick="event.stopPropagation();">
                  ${modPageExists ? `<a href="/mod/${esc(m.id)}" onclick="return handleNavClick(event, 'mod', { modPageId: '${esc(m.id)}' })" class="btn sm" style="text-decoration:none; padding: 3px 8px; font-size: 11px;">Open</a>` : ''}
                  <button type="button" class="btn sm dg" onclick="deleteAdminMod('${esc(m.id)}', '${esc(m.name || m.id)}')" style="padding: 3px 8px; font-size: 11px;">${ic('trash', 12)}</button>
                </div>
              </div>
            </div>`;
          }).join('') : `<div class="mu" style="text-align: center; padding: 24px 10px; font-size: 13px;">
            No panel mods installed yet.<br>Choose a starter template on the right or upload a <code class="mono">.zip/.js/.css</code> bundle.
          </div>`}
        </div>
      </div>

      <div class="card">
        <div class="row sb" style="margin-bottom: 14px; flex-wrap: wrap; gap: 10px;">
          <div>
            <h3 style="margin: 0; font-size: 16px;">${isNew ? 'Create New Panel Mod' : `Editing Mod: ${esc(draft.name || draft.id)}`}</h3>
            <span class="mu" style="font-size: 12px;">Changes write directly to <code class="mono">${esc(S.modsDir || './data/mods')}/${esc(draft.id || '<mod-id>')}</code> and hot-reload.</span>
          </div>
          <div class="row" style="gap: 6px; flex-wrap: wrap;">
            <span class="mu" style="font-size: 11.5px;">Starter Templates:</span>
            <button type="button" class="btn sm" onclick="applyModTemplate('custom-page')">Custom Page + Slot</button>
            <button type="button" class="btn sm" onclick="applyModTemplate('console-widget')">Console Macros</button>
            <button type="button" class="btn sm" onclick="applyModTemplate('custom-theme')">CSS Theme</button>
          </div>
        </div>

        <form onsubmit="saveAdminMod(event)">
          <div class="grid g3">
            <label class="f">
              <span>Mod ID (slug)</span>
              <input type="text" value="${esc(draft.id || '')}" ${isNew ? '' : 'readonly'} required placeholder="my-custom-mod" oninput="onModDraftInput('id', this.value)">
            </label>
            <label class="f">
              <span>Display Name</span>
              <input type="text" value="${esc(draft.name || '')}" placeholder="My Custom Extension" oninput="onModDraftInput('name', this.value)">
            </label>
            <label class="f">
              <span>Version &amp; Author</span>
              <div class="row" style="gap: 6px;">
                <input type="text" value="${esc(draft.version || '1.0.0')}" placeholder="1.0.0" style="width: 90px;" oninput="onModDraftInput('version', this.value)">
                <input type="text" value="${esc(draft.author || '')}" placeholder="Author" oninput="onModDraftInput('author', this.value)">
              </div>
            </label>
          </div>

          <div class="row sb" style="gap: 12px; margin-bottom: 12px; flex-wrap: wrap;">
            <label class="f" style="flex: 1 1 280px; margin-bottom: 0;">
              <span>Description</span>
              <input type="text" value="${esc(draft.description || '')}" placeholder="Short summary of what this extension adds" oninput="onModDraftInput('description', this.value)">
            </label>
            <label class="row" style="gap: 8px; cursor: pointer; padding-top: 18px;">
              <span style="font-size: 13px; color: var(--tx);">Enabled at Runtime</span>
              <span class="sw">
                <input type="checkbox" ${draft.enabled !== false ? 'checked' : ''} onchange="onModDraftInput('enabled', this.checked)">
                <i></i>
              </span>
            </label>
          </div>

          <label class="f">
            <span>JavaScript Entry (<code class="mono">${esc(draft.entry_js || 'mod.js')}</code>)</span>
            <textarea class="mod-code-area" rows="10" spellcheck="false" placeholder="// Use window.NP / NP.registerPage, NP.registerSlot, NP.addSidebarItem, NP.addPaletteCommand..." oninput="onModDraftInput('js_code', this.value)">${esc(draft.js_code || '')}</textarea>
          </label>

          <label class="f" style="margin-top: 12px;">
            <span>Stylesheet Entry (<code class="mono">${esc(draft.entry_css || 'mod.css')}</code>)</span>
            <textarea class="mod-code-area" rows="6" spellcheck="false" placeholder="/* Optional custom CSS rules loaded when enabled */" oninput="onModDraftInput('css_code', this.value)">${esc(draft.css_code || '')}</textarea>
          </label>

          <div class="row sb" style="margin-top: 14px; flex-wrap: wrap; gap: 10px;">
            <div class="row" style="gap: 8px;">
              ${!isNew ? `<button type="button" class="btn dg sm" onclick="deleteAdminMod('${esc(draft.id)}', '${esc(draft.name || draft.id)}')">${ic('trash', 13)} Delete Mod</button>` : ''}
            </div>
            <div class="row" style="gap: 8px;">
              ${!isNew && window.NP?._pages?.has(draft.id) ? `<a href="/mod/${esc(draft.id)}" onclick="return handleNavClick(event, 'mod', { modPageId: '${esc(draft.id)}' })" class="btn sm" style="text-decoration:none;">Open Page</a>` : ''}
              <button type="submit" id="mod-save-btn" class="btn pr">${ic('check', 14)} Save &amp; Apply Live</button>
            </div>
          </div>
        </form>

        <div style="margin-top: 18px; padding: 12px 14px; background: rgba(255,255,255,0.02); border: 1px solid var(--border); border-radius: 9px; font-size: 12px;">
          <b style="color: var(--tx); display: block; margin-bottom: 6px;">window.NP SDK Reference</b>
          <div class="mu mono" style="line-height: 1.7;">
            NP.registerPage({ id, title, icon, scope, render }) &bull; NP.addSidebarItem({ id, label, icon, scope, pageId, url, onClick })<br>
            NP.registerSlot(slot, (el, ctx) =&gt; ...) &rarr; Slots: <b>nav:bottom</b>, <b>dashboard:top</b>, <b>dashboard:bottom</b>, <b>console:below-stats</b>, <b>console:below-charts</b>, <b>files:top</b>, <b>plugins:top</b>, <b>settings:bottom</b><br>
            NP.addPaletteCommand({ id, group, label, action }) &bull; NP.on('render'|'route'|'server:switch', fn) &bull; NP.api(url) &bull; NP.srvApi(sub) &bull; NP.injectCSS(id, css) &bull; NP.storage(modId)
          </div>
        </div>
      </div>
    </div>
  `;
}

// --- Auth Flow & Setup Screen ---
function applyLoginOverlayBackground(el) {
  if (!el) return;
  const bg = (getBrand('login_bg_url') || '').trim();
  if (bg) {
    el.style.backgroundImage = `linear-gradient(rgba(6, 10, 18, 0.72), rgba(6, 10, 18, 0.85)), url('${bg.replace(/'/g, '%27')}')`;
    el.style.backgroundSize = 'cover';
    el.style.backgroundPosition = 'center';
  }
}

async function checkAuth() {
  startPageProgress(22);
  const savedMode = localStorage.getItem('neopanel_mode');
  if (savedMode) setAppearanceMode(savedMode, true);
  if (localStorage.getItem('neopanel_compact') === 'true') toggleCompact(true);
  if (localStorage.getItem('neopanel_priv') === 'true') togglePrivacy(true);
  if (await openPendingInvitation()) { finishPageProgress(); return; }

  let route = parseCurrentRoute(window.location.pathname + window.location.search, window.location.hash);
  if (route.safeMode) S.modsSafeMode = true;
  S.view = route.notFoundPath ? 'notfound' : normalizeRoute(route.view || 'dash');
  if (route.notFoundPath) S.notFoundPath = route.notFoundPath;
  if (route.adminTab !== undefined) S.adminTab = route.adminTab;
  if (route.acctTab !== undefined && S.acct) S.acct.tab = route.acctTab;
  if (route.plugTab !== undefined) S.plugTab = route.plugTab;
  const preferredRef = route.serverId ? { id: route.serverId, nodeId: route.nodeId || null } : null;

  try {
    setPageProgress(42);
    const results = await Promise.all([
      loadPublicBranding().catch(() => {}),
      loadPublicCommunity(true),
      fetch('/api/auth/status').then(r => r.json()),
      loadRuntimeMods().catch(() => [])
    ]);
    const status = results[2];
    S.turnstileEnabled = Boolean(status.turnstile_enabled && status.turnstile_site_key);
    S.turnstileSiteKey = S.turnstileEnabled ? status.turnstile_site_key : '';
    S.maintPublicInfo = {
      enabled: Boolean(status.maintenance_enabled || S.branding?.maintenance_enabled),
      title: status.maintenance_title || S.branding?.maintenance_title || 'Scheduled Maintenance',
      message: status.maintenance_message || S.branding?.maintenance_message || 'We are upgrading NeoPanel nodes. Running game servers stay online during panel maintenance.',
      eta: status.maintenance_eta || S.branding?.maintenance_eta || ''
    };
    if (status.maintenance_enabled) {
      S.branding.maintenance_enabled = true;
      if (status.maintenance_title) S.branding.maintenance_title = status.maintenance_title;
      if (status.maintenance_message) S.branding.maintenance_message = status.maintenance_message;
      if (status.maintenance_eta !== undefined) S.branding.maintenance_eta = status.maintenance_eta;
    }

    if (status.needs_setup) {
      finishPageProgress();
      showSetupScreen();
      return;
    }

    if (!S.token) {
      finishPageProgress();
      showLoginScreen();
      return;
    }

    const meRes = await fetch('/api/auth/me', {
      headers: { 'Authorization': 'Bearer ' + S.token }
    });
    if (!meRes.ok) {
      S.token = '';
      localStorage.removeItem('neopanel_token');
      finishPageProgress();
      showLoginScreen();
      return;
    }
    const user = await meRes.json();
    S.user = user;
    S.acct.un = user.username || '';
    S.acct.em = user.email || '';
    S.acct.avatar_url = user.avatar_url || '';
    S.acct.two_factor_enabled = Boolean(user.two_factor_enabled);
    if (user.first_name) S.acct.fn = user.first_name;
    if (user.last_name) S.acct.ln = user.last_name;
    S.acct.h = await computeGravatarHash(S.acct.em);

    setPageProgress(65);
    await loadDashboard(true, preferredRef || getSavedActiveServerRef());
    // Navigation can change while authentication/inventory requests are in flight.
    route = parseCurrentRoute(window.location.pathname + window.location.search, window.location.hash);
    S.view = route.notFoundPath ? 'notfound' : normalizeRoute(route.view || 'dash');
    if (route.adminTab !== undefined) S.adminTab = route.adminTab;
    if (route.acctTab !== undefined) S.acct.tab = route.acctTab;
    if (route.plugTab !== undefined) S.plugTab = route.plugTab;

    const globalOnlyViews = ['dash', 'client', 'admin', 'acc', 'create', 'notfound', 'maintenance', 'mod'];
    if (route.notFoundPath) {
      S.view = 'notfound';
      S.notFoundPath = route.notFoundPath;
    } else if (S.view === 'mod') {
      const pid = route.modPageId || '';
      if (!window.NP?._pages?.has(pid)) {
        S.notFoundPath = window.location.pathname || `/mod/${pid}`;
        S.view = 'notfound';
      } else {
        S.modPageId = pid;
      }
    } else if (route.serverId) {
      const matched = resolveServerByParam(route.serverId, route.nodeId || null);
      if (!matched) {
        S.view = missingServerView(route, window.location.pathname + window.location.search);
      } else {
        syncActiveServerState(matched);
        if (route.canonicalRedirect || route.legacyHash) {
          try {
            history.replaceState(null, '', serverRouteUrl(matched.node_id || 'local', matched.id, S.view, { plugTab: S.plugTab }));
          } catch (e) {}
        }
      }
    } else if (!globalOnlyViews.includes(S.view) && !(S.view === 'act' && route.global)) {
      if (S.servers.length === 0) {
        S.view = 'dash';
        try { history.replaceState(null, '', withSafeModeQuery('/')); } catch (e) {}
      } else if (route.legacyHash) {
        try { history.replaceState(null, '', routeUrl(S.view, route)); } catch (e) {}
      }
    }

    setPageProgress(82);
    if (!['notfound', 'serverpending', 'maintenance'].includes(S.view)) {
      await loadInitialRouteData(S.view);
    }
    if (!serverPagePermission(S.view) && !['plug', 'set', 'ver', 'notfound', 'serverpending', 'maintenance'].includes(S.view)) {
      loadVersionsCatalog();
    }

    applyBranding(S.branding);
    R();
    scheduleServerRouteRetry();
    finishPageProgress();
  } catch (e) {
    S.token = '';
    localStorage.removeItem('neopanel_token');
    finishPageProgress();
    showLoginScreen();
  }
}

function showSetupScreen() {
  $('#login')?.remove();
  const footTxt = (getBrand('footer_text') || '[NeoPanel](https://github.com/Pekly/NeoPanel/)© 2026').trim();
  const d = document.createElement('div');
  d.id = 'login';
  d.className = 'login';
  d.innerHTML = `<div class="lbox">
    <div class="llogo">${brandLogoHtml(true)}</div>
    <div class="lg">Welcome</div>
    <h1 class="lt">Create your admin account</h1>
    <label class="lf">
      ${ic('user', 16)}
      <input id="su-name" type="text" placeholder="Admin Username" autocomplete="username" onkeydown="if(event.key==='Enter')submitSetup()">
    </label>
    <label class="lf">
      ${ic('mail', 16)}
      <input id="su-email" type="email" placeholder="Administrator Email" autocomplete="email" onkeydown="if(event.key==='Enter')submitSetup()">
    </label>
    <label class="lf">
      ${ic('lock', 16)}
      <input id="su-pw" type="password" placeholder="Password" autocomplete="new-password" onkeydown="if(event.key==='Enter')submitSetup()">
      <button class="eye" type="button" onclick="const p=$('#su-pw'); p.type=p.type==='password'?'text':'password';" aria-label="Toggle password visibility">
        ${ic('eye', 16)}
      </button>
    </label>
    ${S.turnstileEnabled ? '<div id="turnstile-setup" class="turnstile-wrap"></div>' : ''}
    <div class="lerr" id="lerr" role="alert"></div>
    <button class="lbtn" onclick="submitSetup()">Complete setup</button>
    <div class="lsm foot">${renderFooterHtml(footTxt)}</div>
  </div>`;
  applyLoginOverlayBackground(d);
  document.body.appendChild(d);
  if (S.turnstileEnabled) mountTurnstileWidget('#turnstile-setup');
  $('#su-name')?.focus();
}

async function submitSetup() {
  const username = $('#su-name')?.value.trim();
  const email = $('#su-email')?.value.trim();
  const password = $('#su-pw')?.value;
  const err = $('#lerr');
  if (!username || !password) {
    if (err) err.textContent = 'Username and password are required.';
    return;
  }
  if (!email || !email.includes('@') || !email.includes('.')) {
    if (err) err.textContent = 'A valid administrator email address is required.';
    return;
  }
  if (S.turnstileEnabled && !S.turnstileToken) {
    if (err) err.textContent = 'Please complete the Cloudflare Turnstile security check.';
    return;
  }
  const btn = $('#login .lbtn');
  setBtnLoading(btn, true, 'Initializing...');
  try {
    const res = await (await fetch('/api/auth/setup', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username,
        email,
        password,
        'cf-turnstile-response': S.turnstileToken
      })
    })).json();

    if (res.token) {
      S.token = res.token;
      S.turnstileToken = '';
      localStorage.setItem('neopanel_token', res.token);
      $('#login')?.remove();
      toast(`Setup complete! Welcome to ${getBrand('brand_name') || 'NeoPanel'}.`);
      checkAuth();
    } else {
      setBtnLoading(btn, false);
      resetTurnstileWidget();
      if (err) err.textContent = res.error || 'Setup failed.';
    }
  } catch (e) {
    setBtnLoading(btn, false);
    resetTurnstileWidget();
    if (err) err.textContent = 'Connection error.';
  }
}

function showLoginScreen() {
  if ($('#login')) return;
  const maintEnabled = Boolean(S.maintPublicInfo?.enabled || S.branding?.maintenance_enabled);
  const b = S.branding || DEFAULT_BRANDING;
  const footTxt = (getBrand('footer_text') || '[NeoPanel](https://github.com/Pekly/NeoPanel/)© 2026').trim();

  if (maintEnabled && !S.maintAdminSignInOpen) {
    const title = (S.maintPublicInfo?.title || b.maintenance_title || 'Scheduled Maintenance').trim();
    const message = (S.maintPublicInfo?.message || b.maintenance_message || 'We are upgrading NeoPanel nodes. Running game servers stay online during panel maintenance.').trim();
    const eta = (S.maintPublicInfo?.eta || b.maintenance_eta || '').trim();
    const d = document.createElement('div');
    d.id = 'login';
    d.className = 'login';
    d.innerHTML = `<div class="card np-status-card">
      <div class="llogo" style="margin-bottom:14px;">${brandLogoHtml(true)}</div>
      <div class="np-status-code warn">${ic('alert', 14)} 503 MAINTENANCE</div>
      <h1 style="font-size:25px; font-weight:800; margin:0 0 10px; color:var(--tx);">${esc(title)}</h1>
      <p class="mu" style="font-size:13.5px; line-height:1.65; margin:0;">${esc(message)}</p>
      ${eta ? `<div class="chip mono" style="margin-top:14px; display:inline-flex;">Estimated return: ${esc(eta)}</div>` : ''}
      <div class="row" style="justify-content:center; gap:10px; margin-top:24px; flex-wrap:wrap;">
        <button type="button" class="btn pr" onclick="$('#login')?.remove(); checkAuth();">${ic('refresh', 14)} Check Status</button>
        <button type="button" class="btn" onclick="enterAdminLoginFromMaint()">${ic('shield', 14)} Administrator Sign In</button>
      </div>
      <div class="lsm foot" style="margin-top:18px;">${renderFooterHtml(footTxt)}</div>
    </div>`;
    applyLoginOverlayBackground(d);
    document.body.appendChild(d);
    return;
  }

  const hour = new Date().getHours();
  const greeting = hour < 5 ? 'Good evening' : hour < 12 ? 'Good morning' : hour < 18 ? 'Good afternoon' : 'Good evening';
  const brandName = (getBrand('brand_name') || 'NeoPanel').trim() || 'NeoPanel';

  const d = document.createElement('div');
  d.id = 'login';
  d.className = 'login';
  d.innerHTML = `<div class="lbox">
    <div class="llogo">${brandLogoHtml(true)}</div>
    ${maintEnabled ? `<div class="np-status-code warn" style="margin-bottom:10px;">${ic('alert', 13)} MAINTENANCE MODE &bull; ADMIN LOGIN</div>` : ''}
    <div class="lg">${greeting}</div>
    <h1 class="lt">Sign in to ${esc(brandName)}</h1>
    <label class="lf">
      ${ic('user', 16)}
      <input id="lu" type="text" placeholder="Username or email" autocomplete="username" onkeydown="if(event.key==='Enter')submitLogin()">
    </label>
    <label class="lf">
      ${ic('lock', 16)}
      <input id="lp" type="password" placeholder="Password" autocomplete="current-password" onkeydown="if(event.key==='Enter')submitLogin()">
      <button class="eye" type="button" onclick="const p=$('#lp'); p.type=p.type==='password'?'text':'password';" aria-label="Toggle password visibility">
        ${ic('eye', 16)}
      </button>
    </label>
    <div class="row sb" style="margin: -6px 0 12px; font-size: 12.5px;">
      <span></span>
      <a href="javascript:void(0)" onclick="openResetPasswordModal()" style="color: var(--ac); text-decoration: none; font-weight: 500;">Reset Password</a>
    </div>
    ${S.turnstileEnabled ? '<div id="turnstile-login" class="turnstile-wrap"></div>' : ''}
    <div class="lerr" id="lerr" role="alert"></div>
    <button class="lbtn" onclick="submitLogin()">Sign in</button>
    ${renderPasskeySignIn()}
    ${maintEnabled ? `<button type="button" class="btn sm" style="width:100%; margin-top:10px; justify-content:center;" onclick="S.maintAdminSignInOpen=false; $('#login')?.remove(); showLoginScreen();">Back to Maintenance Notice</button>` : ''}
    <div class="lsm foot">${renderFooterHtml(footTxt)}</div>
  </div>`;
  applyLoginOverlayBackground(d);
  document.body.appendChild(d);
  if (S.turnstileEnabled) mountTurnstileWidget('#turnstile-login');
  $('#lu')?.focus();
}

async function submitLogin() {
  const u = $('#lu')?.value.trim();
  const p = $('#lp')?.value;
  const err = $('#lerr');

  if (!u || !p) {
    if (err) err.textContent = 'Please enter your username and password.';
    return;
  }

  if (S.turnstileEnabled && !S.turnstileToken) {
    if (err) err.textContent = 'Please complete the Cloudflare Turnstile security check.';
    return;
  }

  const btn = $('#login .lbtn');
  setBtnLoading(btn, true, 'Signing in...');
  try {
    const res = await (await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username: u,
        password: p,
        'cf-turnstile-response': S.turnstileToken
      })
    })).json();

    if (res.two_factor_required && res.temp_token) {
      setBtnLoading(btn, false);
      resetTurnstileWidget();
      show2FALoginScreen(res.temp_token);
      return;
    }

    if (res.token) {
      S.token = res.token;
      S.turnstileToken = '';
      S.maintAdminSignInOpen = false;
      localStorage.setItem('neopanel_token', res.token);
      $('#login')?.remove();
      toast('Signed in successfully');
      checkAuth();
    } else {
      setBtnLoading(btn, false);
      resetTurnstileWidget();
      if (err) err.textContent = res.error || 'Invalid credentials.';
    }
  } catch (e) {
    setBtnLoading(btn, false);
    resetTurnstileWidget();
    if (err) err.textContent = 'Connection error.';
  }
}

function show2FALoginScreen(tempToken) {
  const box = $('#login .lbox');
  if (!box) return;
  box.innerHTML = `
    <div class="llogo">${brandLogoHtml(true)}</div>
    <div style="width:48px; height:48px; border-radius:14px; background:linear-gradient(135deg,#6366f1,#8b5cf6); color:#fff; display:grid; place-items:center; margin:0 auto 16px; box-shadow:0 4px 14px rgba(99,102,241,0.35);">
      ${ic('shield', 24)}
    </div>
    <h1 class="lt" style="margin-bottom:6px; font-size:22px;">Two-factor authentication</h1>
    <p class="mu" style="font-size:13px; line-height:1.45; margin:0 0 18px; text-align:center;">
      Enter the 6-digit code from your authenticator app, or an 8-character recovery code.
    </p>
    <label class="lf" style="margin-bottom:14px;">
      ${ic('key', 16)}
      <input id="l2fa-code" type="text" maxlength="9" placeholder="Authentication code" aria-label="Authentication or recovery code" autofocus autocomplete="one-time-code" autocapitalize="off" spellcheck="false" onkeydown="if(event.key==='Enter')submit2FALogin('${esc(tempToken)}')">
    </label>
    <div class="lerr" id="lerr" role="alert"></div>
    <button class="lbtn" onclick="submit2FALogin('${esc(tempToken)}')">Verify &amp; sign in</button>
    <button type="button" class="btn sm" style="width:100%; margin-top:12px; justify-content:center;" onclick="$('#login')?.remove(); showLoginScreen();">
      &larr; Back to sign in
    </button>
  `;
  $('#l2fa-code')?.focus();
}

async function submit2FALogin(tempToken) {
  const code = $('#l2fa-code')?.value.trim();
  const err = $('#lerr');
  if (!code) {
    if (err) err.textContent = 'Please enter your 6-digit code or recovery code.';
    return;
  }
  const btn = $('#login .lbtn');
  setBtnLoading(btn, true, 'Verifying...');
  try {
    const res = await (await fetch('/api/auth/login/2fa', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ temp_token: tempToken, code })
    })).json();

    if (res.token) {
      S.token = res.token;
      S.maintAdminSignInOpen = false;
      localStorage.setItem('neopanel_token', res.token);
      $('#login')?.remove();
      toast('Signed in successfully');
      checkAuth();
    } else {
      setBtnLoading(btn, false);
      if (err) err.textContent = res.error || 'Invalid verification code.';
    }
  } catch (e) {
    setBtnLoading(btn, false);
    if (err) err.textContent = 'Verification request failed.';
  }
}

async function logout() {
  resetServerAccess();
  if (!await guardEditorLeave()) return;
  const brandName = (getBrand('brand_name') || 'NeoPanel').trim() || 'NeoPanel';
  const ok = await uiConfirm({
    title: `Sign Out of ${brandName}`,
    message: `Are you sure you want to log out of your current ${esc(brandName)} session?`,
    confirmText: 'Log Out',
    cancelText: 'Stay Signed In',
    isDanger: false,
    icon: 'lock'
  });
  if (!ok) return;
  cancelAllUploads();
  QOL.uploads = [];
  renderUploadQueue();
  QOL.editor = null;
  ++QOL.openSequence;

  S.token = '';
  S.user = null;
  clearTimeout(S.serverRoutePending?.timer);
  S.serverRoutePending = null;
  S.serverCatalog = null;
  S.servers = [];
  ++dashboardLoadSequence;
  S.acct = {
    un: '',
    em: '',
    fn: '',
    ln: '',
    h: '',
    tab: 0,
    ssh: [],
    api: [],
    act: [],
    avatar_url: '',
    two_factor_enabled: false
  };
  S.activeServerId = null;
  S.activeNodeId = null;
  clearSettingsAutoSaveTimer();
  _settingsSaveQueued = null;
  S.wsState = 'disconnected';
  S.maintAdminSignInOpen = false;
  localStorage.removeItem('neopanel_token');
  localStorage.removeItem('neopanel_profile');
  clearTimeout(S.wsReconnectTimer);
  S.wsReconnectTimer = null;
  if (S.wsPingTimer) {
    clearInterval(S.wsPingTimer);
    S.wsPingTimer = null;
  }
  if (S.activeWs) {
    try {
      S.activeWs.onopen = null;
      S.activeWs.onmessage = null;
      S.activeWs.onerror = null;
      S.activeWs.onclose = null;
      S.activeWs.close();
    } catch (e) {}
    S.activeWs = null;
  }
  S.flyConOpen = false;
  S.log = [];
  renderFlyConsole();
  $('#login')?.remove();
  showLoginScreen();
}

// --- Attach All Functions to Window for Seamless HTML Event Handling ---
window.getNode = getNode;
window.getNodeById = getNodeById;
window.runCmd = runCmd;
window.toast = toast;
window.power = power;
window.switchServer = switchServer;
window.openPalette = openPalette;
window.execPaletteItem = execPaletteItem;
window.toggleCmdMenu = toggleCmdMenu;
window.execQuickCmd = execQuickCmd;
window.renderAdminUpdateBtn = renderAdminUpdateBtn;
window.startUpdateProgressPolling = startUpdateProgressPolling;
window.waitForPanelRestart = waitForPanelRestart;
window.checkAdminUpdates = checkAdminUpdates;
window.saveAdminUpdateConfig = saveAdminUpdateConfig;
window.resetAdminUpdateRepo = resetAdminUpdateRepo;
window.openApplyUpdateModal = openApplyUpdateModal;
window.executePanelUpdate = executePanelUpdate;
window.openUpdateLogsModal = openUpdateLogsModal;
window.pollUpdateLogsOnce = pollUpdateLogsOnce;
window.renderFooterHtml = renderFooterHtml;
window.renderAdminEmailTabHtml = renderAdminEmailTabHtml;
window.applySmtpPreset = applySmtpPreset;
window.loadAdminSmtp = loadAdminSmtp;
window.saveAdminSmtp = saveAdminSmtp;
window.sendAdminTestEmail = sendAdminTestEmail;
window.openResetPasswordModal = openResetPasswordModal;
window.submitResetPasswordRequest = submitResetPasswordRequest;
window.openSetNewPasswordModal = openSetNewPasswordModal;
window.submitSetNewPassword = submitSetNewPassword;
window.checkUrlResetToken = checkUrlResetToken;
window.playerAction = playerAction;
window.addWhitelistPlayer = addWhitelistPlayer;
window.browseFolder = browseFolder;
window.openFile = openFile;
window.saveActiveFile = saveActiveFile;
window.downloadFileItem = downloadFileItem;
window.promptRenameFile = promptRenameFile;
window.deleteFileItem = deleteFileItem;
window.promptNewFile = promptNewFile;
window.promptNewFolder = promptNewFolder;
window.uploadFileDialog = uploadFileDialog;
window.deletePlugin = deletePlugin;
window.togglePlugin = togglePlugin;
window.searchModrinthDebounced = searchModrinthDebounced;
window.searchModrinth = searchModrinth;
window.installPlugin = installPlugin;
window.setPlugSkipConfirm = setPlugSkipConfirm;
window.createBackup = createBackup;
window.downloadBackup = downloadBackup;
window.deleteBackup = deleteBackup;
window.installSoftware = installSoftware;
window.saveProperties = saveProperties;
window.updateAccountProfile = updateAccountProfile;
window.updatePassword = updatePassword;
window.addSSHKey = addSSHKey;
window.deleteSSHKey = deleteSSHKey;
window.createAPIKey = createAPIKey;
window.deleteAPIKey = deleteAPIKey;
window.setAppearanceMode = setAppearanceMode;
window.toggleCompact = toggleCompact;
window.toggleSound = toggleSound;
window.togglePrivacy = togglePrivacy;
window.toggleTheme = toggleTheme;
window.openCreateServerModal = openCreateServerModal;
window.closeModal = closeModal;
window.handleCreateServer = handleCreateServer;
window.submitSetup = submitSetup;
window.submitLogin = submitLogin;
window.logout = logout;
window.checkAuth = checkAuth;
window.loadDashboard = loadDashboard;
window.updateDashLiveStats = updateDashLiveStats;
window.refreshServerStatus = refreshServerStatus;
window.paintLog = paintLog;
window.renderVersionRows = renderVersionRows;
window.loadAccountData = loadAccountData;
window.loadPlayersData = loadPlayersData;
window.bulkRestartServers = bulkRestartServers;
window.bulkStopServers = bulkStopServers;
window.filterServerCards = filterServerCards;
window.filterFileList = filterFileList;
window.toggleFileCtx = toggleFileCtx;
window.closeAllCtx = closeAllCtx;
window.closeEditor = closeEditor;
window.setPlugTab = setPlugTab;
window.saveStartupSettings = saveStartupSettings;
window.formatFileDate = formatFileDate;
window.formatFileSize = formatFileSize;
window.formatNum = formatNum;
window.openAddPortModal = openAddPortModal;
window.submitAddPort = submitAddPort;
window.deletePort = deletePort;
window.setPrimaryPort = setPrimaryPort;
window.setPortPreset = setPortPreset;
window.loadServerPorts = loadServerPorts;
window.openDeleteServerModal = openDeleteServerModal;
window.checkDeleteServerConfirm = checkDeleteServerConfirm;
window.executeDeleteServer = executeDeleteServer;
window.toggleFileSelect = toggleFileSelect;
window.toggleSelectAllFiles = toggleSelectAllFiles;
window.deselectAllFiles = deselectAllFiles;
window.deleteSelectedFiles = deleteSelectedFiles;
window.copyFilePath = copyFilePath;
window.handleFileContextMenu = handleFileContextMenu;
window.closeFloatingCtx = closeFloatingCtx;
window.uploadFilesWithProgress = uploadFilesWithProgress;
window.openReinstallModal = openReinstallModal;
window.submitReinstall = submitReinstall;
window.openAdminServerSettingsModal = openAdminServerSettingsModal;
window.submitAdminServerSettings = submitAdminServerSettings;
window.openAddUserModal = openAddUserModal;
window.submitAddUser = submitAddUser;
window.openEditUserModal = openEditUserModal;
window.submitEditUser = submitEditUser;
window.deleteAdminUser = deleteAdminUser;
window.adminTriggerUserAvatarUpload = adminTriggerUserAvatarUpload;
window.onAdminUserAvatarSelected = onAdminUserAvatarSelected;
window.uploadMyAvatar = uploadMyAvatar;
window.openEnable2FAWizard = openEnable2FAWizard;
window.confirmEnable2FA = confirmEnable2FA;
window.openDisable2FAModal = openDisable2FAModal;
window.confirmDisable2FA = confirmDisable2FA;
window.show2FALoginScreen = show2FALoginScreen;
window.submit2FALogin = submit2FALogin;
window.selectAdminTabNav = selectAdminTabNav;
window.triggerAdminGC = triggerAdminGC;
window.saveAdminDaemonSettings = saveAdminDaemonSettings;
window.saveAdminTurnstileSettings = saveAdminTurnstileSettings;
window.toggleCustomJavaInput = toggleCustomJavaInput;
window.toggleCustomPropJava = toggleCustomPropJava;
window.isArchiveFile = isArchiveFile;
window.unzipFileItem = unzipFileItem;
window.unzipSelectedArchive = unzipSelectedArchive;
window.openMoveFilesModal = openMoveFilesModal;
window.setMoveModalDest = setMoveModalDest;
window.submitMoveFilesModal = submitMoveFilesModal;
window.openCompressFilesModal = openCompressFilesModal;
window.setCompressModalFormat = setCompressModalFormat;
window.submitCompressFilesModal = submitCompressFilesModal;
window.renderMinecraftTextHtml = renderMinecraftTextHtml;
window.onMotdInput = onMotdInput;
window.enhanceSelects = enhanceSelects;
window.closeAllCustomSelects = closeAllCustomSelects;
window.toggleFlyConsole = toggleFlyConsole;
window.toggleFlyConsoleSize = toggleFlyConsoleSize;
window.closeFlyConsole = closeFlyConsole;
window.runFlyCmd = runFlyCmd;
window.renderFlyConsole = renderFlyConsole;
window.loadActivityLog = loadActivityLog;
window.setActivityQuery = setActivityQuery;
window.setActivityCategory = setActivityCategory;
window.loadMoreActivity = loadMoreActivity;
window.exportActivityLogs = exportActivityLogs;
window.loadJavaRuntimes = loadJavaRuntimes;
window.installJavaRuntime = installJavaRuntime;
window.updateReinstallJavaPreview = updateReinstallJavaPreview;
window.setNodeFilter = setNodeFilter;
window.selectAdminTab = selectAdminTab;
window.selectAccountTab = selectAccountTab;
window.loadAdminNodes = loadAdminNodes;
window.toggleClusterNodeHelper = toggleClusterNodeHelper;
window.buildClusterNodeInstallCmd = buildClusterNodeInstallCmd;
window.onNodeHelperInput = onNodeHelperInput;
window.generateNodePairingKey = generateNodePairingKey;
window.openNodeModal = openNodeModal;
window.testNodeModalConnection = testNodeModalConnection;
window.submitNodeModal = submitNodeModal;
window.testClusterNode = testClusterNode;
window.deleteClusterNode = deleteClusterNode;
window.loadDatabases = loadDatabases;
window.openCreateDatabaseModal = openCreateDatabaseModal;
window.submitCreateDatabase = submitCreateDatabase;
window.toggleDbPasswordReveal = toggleDbPasswordReveal;
window.rotateDatabasePassword = rotateDatabasePassword;
window.deleteServerDatabase = deleteServerDatabase;
window.exportServerDatabase = exportServerDatabase;
window.openSQLStudioModal = openSQLStudioModal;
window.loadSQLStudioTables = loadSQLStudioTables;
window.selectSQLStudioTable = selectSQLStudioTable;
window.insertSQLSnippet = insertSQLSnippet;
window.runSQLStudioQuery = runSQLStudioQuery;
window.copyDbText = copyDbText;
window.selectBrandPreset = selectBrandPreset;
window.onBrandInputChange = onBrandInputChange;
window.onDaemonInputChange = onDaemonInputChange;
window.discardBrandDraft = discardBrandDraft;
window.saveBrandingAndDaemon = saveBrandingAndDaemon;
window.resetBrandingDefaults = resetBrandingDefaults;
window.triggerBrandingUpload = triggerBrandingUpload;
window.handleBrandingFileSelect = handleBrandingFileSelect;
window.clearBrandingMedia = clearBrandingMedia;
window.toggleMaintPreview = toggleMaintPreview;
window.enterAdminLoginFromMaint = enterAdminLoginFromMaint;
window.loadRuntimeMods = loadRuntimeMods;
window.mountPanelModHooks = mountPanelModHooks;
window.loadAdminMods = loadAdminMods;
window.selectAdminMod = selectAdminMod;
window.startNewModDraft = startNewModDraft;
window.applyModTemplate = applyModTemplate;
window.onModDraftInput = onModDraftInput;
window.saveAdminMod = saveAdminMod;
window.toggleAdminModEnabled = toggleAdminModEnabled;
window.deleteAdminMod = deleteAdminMod;
window.triggerModPackageUpload = triggerModPackageUpload;
window.handleModPackageUpload = handleModPackageUpload;
window.onCreateNodeChange = onCreateNodeChange;
window.handleNavClick = handleNavClick;
window.handleServerCardClick = handleServerCardClick;
window.applyRoute = applyRoute;
window.navigateToUrl = navigateToUrl;
window.shortServerId = shortServerId;
window.resolveServerByParam = resolveServerByParam;
window.copyServerAddress = copyServerAddress;
window.onChartMouseMove = onChartMouseMove;
window.onChartMouseLeave = onChartMouseLeave;
window.updateChartHoverDom = updateChartHoverDom;
window.detectFaviconMime = detectFaviconMime;
window.setViewMode = setViewMode;
window.getViewMode = getViewMode;
window.routeUrl = routeUrl;
window.serverRouteUrl = serverRouteUrl;
window.setCreateRam = setCreateRam;
window.syncCreateJavaFromVersion = syncCreateJavaFromVersion;
window.onCreateVersionInput = onCreateVersionInput;
window.setCreateEngine = setCreateEngine;
window.setCreateVersion = setCreateVersion;
window.onReinstallEngineChange = onReinstallEngineChange;
window.setReinstallVersion = setReinstallVersion;
window.syncReinstallSelectFromInput = syncReinstallSelectFromInput;
window.refreshOpenVersionSelects = refreshOpenVersionSelects;
window.renderModrinthHitsHtml = renderModrinthHitsHtml;
window.isModdedEngine = isModdedEngine;
window.engineMeta = engineMeta;
window.engineIcon = engineIcon;
window.onPlayerHeadError = onPlayerHeadError;
window.getSftpDetails = getSftpDetails;
window.launchSftpClient = launchSftpClient;
window.downloadFileZillaProfile = downloadFileZillaProfile;
window.renderSftpCardHtml = renderSftpCardHtml;
window.getFileExtension = getFileExtension;
window.sortFileItems = sortFileItems;
window.setFileSort = setFileSort;
window.computeSubdomainDraft = computeSubdomainDraft;
window.getSubdomainPreviewFqdn = getSubdomainPreviewFqdn;
window.onSubdomainFieldChange = onSubdomainFieldChange;
window.loadSubdomainData = loadSubdomainData;
window.saveSubdomainConfig = saveSubdomainConfig;
window.verifySubdomainDNS = verifySubdomainDNS;
window.syncPlayersList = syncPlayersList;
window.onServerIdentityInput = onServerIdentityInput;
window.updateSettingsCmdPreview = updateSettingsCmdPreview;
window.scheduleSettingsAutoSave = scheduleSettingsAutoSave;
window.saveSettingsAutoNow = saveSettingsAutoNow;
window.flushSettingsAutoSave = flushSettingsAutoSave;
window.onSettingsRuleToggle = onSettingsRuleToggle;
window.onSettingsMaxPlayersInput = onSettingsMaxPlayersInput;
window.onPropJavaChange = onPropJavaChange;
window.onPropVersionSelectChange = onPropVersionSelectChange;
window.onPropVersionInput = onPropVersionInput;
window.setSettingsSaveStatus = setSettingsSaveStatus;
window.loadServerSchedules = loadServerSchedules;
window.describeCron = describeCron;
window.getScheduleActionBadge = getScheduleActionBadge;
window.openScheduleModal = openScheduleModal;
window.onScheduleActionChange = onScheduleActionChange;
window.submitScheduleModal = submitScheduleModal;
window.toggleServerSchedule = toggleServerSchedule;
window.runScheduleNow = runScheduleNow;
window.deleteServerSchedule = deleteServerSchedule;
window.loadServerSubusers = loadServerSubusers;
window.renderSubusersListHtml = renderSubusersListHtml;
window.openSubuserModal = openSubuserModal;
window.submitSubuserModal = submitSubuserModal;
window.deleteServerSubuser = deleteServerSubuser;
window.uploadServerIcon = uploadServerIcon;
window.deleteServerIcon = deleteServerIcon;
window.openQuickConfigsModal = openQuickConfigsModal;
window.renderQuickConfigsModalBody = renderQuickConfigsModalBody;
window.applyQuickConfigPreset = applyQuickConfigPreset;
window.saveQuickConfigs = saveQuickConfigs;
window.setClientCodeLang = setClientCodeLang;
window.getClientProvisioningCode = getClientProvisioningCode;
window.copyClientProvisioningCode = copyClientProvisioningCode;
window.setUserTimezone = setUserTimezone;

// --- Custom Glassmorphic Selects System (Portaled to document.body to prevent clipping) ---
function closeAllCustomSelects() {
  document.querySelectorAll('.cselect-menu.cselect-portaled').forEach(menu => {
    menu.classList.remove('cselect-portaled', 'cselect-up');
    menu.style.left = '';
    menu.style.top = '';
    menu.style.bottom = '';
    menu.style.width = '';
    menu.style.minWidth = '';
    menu.style.maxHeight = '';
    const home = menu._homeWrap;
    if (home && document.body.contains(home)) {
      home.appendChild(menu);
    } else if (menu.parentNode === document.body) {
      menu.remove();
    }
  });
  document.querySelectorAll('.cselect-wrap.open').forEach(el => {
    el.classList.remove('open', 'cselect-up');
    const trigger = el.querySelector('.cselect-trigger');
    if (trigger) trigger.setAttribute('aria-expanded', 'false');
  });
}

function positionPortaledMenu(wrap, trigger, menu) {
  const rect = trigger.getBoundingClientRect();
  if (rect.width === 0 && rect.height === 0) return;

  menu._homeWrap = wrap;
  const hasSearch = Boolean(menu.querySelector('.cselect-search'));
  const spaceBelow = window.innerHeight - rect.bottom;
  const spaceAbove = rect.top;
  const threshold = hasSearch ? 260 : 190;
  const openUp = spaceBelow < threshold && spaceAbove > spaceBelow;

  const width = Math.min(Math.max(rect.width, 210), window.innerWidth - 16);
  let left = rect.left;
  if (left + width > window.innerWidth - 8) {
    left = window.innerWidth - width - 8;
  }
  left = Math.max(8, left);

  menu.classList.add('cselect-portaled');
  menu.classList.toggle('cselect-up', openUp);
  wrap.classList.toggle('cselect-up', openUp);

  menu.style.left = `${Math.round(left)}px`;
  menu.style.width = `${Math.round(width)}px`;
  menu.style.minWidth = `${Math.round(width)}px`;

  const availableSpace = openUp ? spaceAbove - 16 : spaceBelow - 16;
  const menuMaxH = Math.min(340, Math.max(200, Math.floor(availableSpace)));

  if (openUp) {
    menu.style.top = 'auto';
    menu.style.bottom = `${Math.round(window.innerHeight - rect.top + 4)}px`;
    menu.style.maxHeight = `${menuMaxH}px`;
  } else {
    menu.style.bottom = 'auto';
    menu.style.top = `${Math.round(rect.bottom + 4)}px`;
    menu.style.maxHeight = `${menuMaxH}px`;
  }

  if (menu.parentNode !== document.body) {
    document.body.appendChild(menu);
  }
}

function enhanceSelects(root = document) {
  if (!root || !root.querySelectorAll) return;
  const selects = root.tagName === 'SELECT' ? [root] : Array.from(root.querySelectorAll('select'));
  for (const sel of selects) {
    if (sel.dataset.cselectInit === '1') continue;
    if (sel.classList.contains('no-cselect')) continue;
    if (sel.id === 'cs-type') continue;
    if (sel.style.display === 'none' && !sel.classList.contains('cselect-native-hidden')) continue;

    sel.dataset.cselectInit = '1';
    sel.classList.add('cselect-native-hidden');
    sel.style.display = 'none';

    const wrap = document.createElement('div');
    wrap.className = 'cselect-wrap';
    if (sel.id) wrap.dataset.for = sel.id;
    if (sel.style.width) wrap.style.width = sel.style.width;
    if (sel.style.minWidth) wrap.style.minWidth = sel.style.minWidth;
    if (sel.style.marginBottom) wrap.style.marginBottom = sel.style.marginBottom;
    if (sel.style.marginTop) wrap.style.marginTop = sel.style.marginTop;

    const trigger = document.createElement('button');
    trigger.type = 'button';
    trigger.className = 'cselect-trigger';
    trigger.setAttribute('aria-haspopup', 'listbox');
    trigger.setAttribute('aria-expanded', 'false');
    if (sel.disabled) trigger.disabled = true;

    const valSpan = document.createElement('span');
    valSpan.className = 'cselect-val';

    const arrowSpan = document.createElement('span');
    arrowSpan.className = 'cselect-arrow';
    arrowSpan.innerHTML = `<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"></polyline></svg>`;

    trigger.appendChild(valSpan);
    trigger.appendChild(arrowSpan);

    const menu = document.createElement('div');
    menu.className = 'cselect-menu';
    menu.setAttribute('role', 'listbox');
    menu._homeWrap = wrap;

    function rebuildOptions() {
      menu.innerHTML = '';
      const opts = Array.from(sel.options);
      const isSearchable = sel.dataset.searchable === '1' || opts.length > 5;

      let searchInp = null;
      let emptyEl = null;
      const optionItems = [];

      if (isSearchable) {
        const searchWrap = document.createElement('div');
        searchWrap.className = 'cselect-search-wrap';
        searchInp = document.createElement('input');
        searchInp.type = 'search';
        searchInp.className = 'cselect-search';
        searchInp.placeholder = 'Search options...';
        searchInp.setAttribute('autocomplete', 'off');
        searchInp.addEventListener('click', (e) => e.stopPropagation());
        searchInp.addEventListener('keydown', (e) => {
          e.stopPropagation();
          if (e.key === 'Escape') {
            closeAllCustomSelects();
            trigger.focus();
          } else if (e.key === 'Enter') {
            e.preventDefault();
            const firstVis = optionItems.find(it => it.style.display !== 'none' && it.dataset.disabled !== '1');
            if (firstVis) firstVis.click();
          }
        });
        searchWrap.appendChild(searchInp);
        menu.appendChild(searchWrap);
      }

      opts.forEach((opt, idx) => {
        const item = document.createElement('div');
        item.className = 'cselect-opt' + (opt.selected ? ' on' : '');
        item.textContent = opt.textContent;
        item.dataset.value = opt.value;
        item.dataset.index = String(idx);
        item.setAttribute('role', 'option');
        if (opt.disabled) {
          item.style.opacity = '0.5';
          item.dataset.disabled = '1';
        }

        item.addEventListener('click', (e) => {
          e.stopPropagation();
          if (opt.disabled) return;
          sel.selectedIndex = idx;
          valSpan.textContent = opt.textContent;
          menu.querySelectorAll('.cselect-opt').forEach(el => el.classList.remove('on'));
          item.classList.add('on');
          closeAllCustomSelects();
          sel.dispatchEvent(new Event('change', { bubbles: true }));
        });
        menu.appendChild(item);
        optionItems.push(item);
      });

      if (isSearchable) {
        emptyEl = document.createElement('div');
        emptyEl.className = 'cselect-empty';
        emptyEl.textContent = 'No matching options';
        emptyEl.style.display = 'none';
        menu.appendChild(emptyEl);

        searchInp.addEventListener('input', () => {
          const term = searchInp.value.trim().toLowerCase();
          let visibleCount = 0;
          for (const it of optionItems) {
            const hay = `${it.textContent || ''} ${it.dataset.value || ''}`.toLowerCase();
            const show = !term || hay.includes(term);
            it.style.display = show ? '' : 'none';
            if (show) visibleCount++;
          }
          if (emptyEl) emptyEl.style.display = visibleCount === 0 ? 'block' : 'none';
        });
      }

      menu._resetSearch = () => {
        if (!searchInp) return;
        searchInp.value = '';
        optionItems.forEach(it => { it.style.display = ''; });
        if (emptyEl) emptyEl.style.display = 'none';
      };

      const currentOpt = sel.options[sel.selectedIndex] || sel.options[0];
      valSpan.textContent = currentOpt ? currentOpt.textContent : 'Select...';
    }

    rebuildOptions();

    wrap.appendChild(trigger);
    wrap.appendChild(menu);

    if (sel.parentNode) {
      sel.parentNode.insertBefore(wrap, sel.nextSibling);
    }

    sel._cselectWrap = wrap;
    sel._cselectRebuild = rebuildOptions;

    trigger.addEventListener('click', (e) => {
      e.stopPropagation();
      if (sel.disabled) return;
      const isOpen = wrap.classList.contains('open');
      closeAllCustomSelects();
      if (!isOpen) {
        if (typeof menu._resetSearch === 'function') menu._resetSearch();
        wrap.classList.add('open');
        trigger.setAttribute('aria-expanded', 'true');
        positionPortaledMenu(wrap, trigger, menu);
        const active = menu.querySelector('.cselect-opt.on');
        if (active) active.scrollIntoView({ block: 'nearest' });
        const searchInp = menu.querySelector('.cselect-search');
        if (searchInp) requestAnimationFrame(() => searchInp.focus());
      }
    });

    trigger.addEventListener('keydown', (e) => {
      if (e.key === ' ' || e.key === 'Enter') {
        e.preventDefault();
        trigger.click();
      } else if (e.key === 'ArrowDown') {
        e.preventDefault();
        if (!wrap.classList.contains('open')) {
          trigger.click();
        } else {
          const items = Array.from(menu.querySelectorAll('.cselect-opt'));
          const cur = items.findIndex(o => o.classList.contains('on'));
          const next = Math.min(items.length - 1, cur + 1);
          if (items[next]) items[next].click();
        }
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        if (!wrap.classList.contains('open')) {
          trigger.click();
        } else {
          const items = Array.from(menu.querySelectorAll('.cselect-opt'));
          const cur = items.findIndex(o => o.classList.contains('on'));
          const prev = Math.max(0, cur - 1);
          if (items[prev]) items[prev].click();
        }
      } else if (e.key === 'Escape') {
        closeAllCustomSelects();
      }
    });

    sel.addEventListener('change', () => {
      const currentOpt = sel.options[sel.selectedIndex];
      if (currentOpt) {
        valSpan.textContent = currentOpt.textContent;
        menu.querySelectorAll('.cselect-opt').forEach((el, i) => {
          el.classList.toggle('on', i === sel.selectedIndex);
        });
      }
    });

    if (window.MutationObserver) {
      const selObserver = new MutationObserver(() => {
        trigger.disabled = sel.disabled;
        rebuildOptions();
      });
      selObserver.observe(sel, { childList: true, subtree: true, attributes: true, attributeFilter: ['disabled'] });
    }
  }
}

document.addEventListener('click', (e) => {
  if (!e.target.closest('.cselect-wrap') && !e.target.closest('.cselect-menu')) {
    closeAllCustomSelects();
  }
  if (S.cmenu && !e.target.closest('.cmenu')) {
    S.cmenu = false;
    R();
  }
});

window.addEventListener('resize', closeAllCustomSelects);
document.addEventListener('scroll', (e) => {
  if (e.target && e.target.closest && e.target.closest('.cselect-menu')) return;
  if (document.querySelector('.cselect-wrap.open, .cselect-menu.cselect-portaled')) {
    closeAllCustomSelects();
  }
}, true);

let _enhanceTimeout = null;
const globalSelectObserver = (typeof MutationObserver !== 'undefined') ? new MutationObserver((mutations) => {
  let hasSelect = false;
  for (const m of mutations) {
    for (const node of m.addedNodes) {
      if (node.nodeType === 1) {
        if (node.tagName === 'SELECT' || (node.querySelector && node.querySelector('select'))) {
          hasSelect = true;
          break;
        }
      }
    }
    if (hasSelect) break;
  }
  if (hasSelect) {
    if (_enhanceTimeout) cancelAnimationFrame(_enhanceTimeout);
    _enhanceTimeout = requestAnimationFrame(() => {
      enhanceSelects(document.body);
    });
  }
}) : null;

// --- Initialize Application ---
document.addEventListener('DOMContentLoaded', () => {
  checkAuth();
  initDragAndDrop();
  enhanceSelects(document.body);
  checkUrlResetToken();
  if (globalSelectObserver) {
    globalSelectObserver.observe(document.body, { childList: true, subtree: true });
  }
});

// --- Permission-aware workspace and panel governance ---
// UI guards are fail-closed conveniences; the API remains the authority.
function accessIdentity(srv = curServer()) {
  return JSON.stringify([S.token, S.user?.id, S.user?.role, srv?.node_id || 'local', srv?.id]);
}
function resetServerAccess() {
  S.serverAccess?.controller?.abort();
  S.serverAccess = null;
  clearTimeout(S.wsReconnectTimer);
  clearInterval(S.wsPingTimer);
  if (S.activeWs) {
    S.activeWs.onopen = S.activeWs.onmessage = S.activeWs.onclose = S.activeWs.onerror = null;
    try { S.activeWs.close(); } catch (_) {}
    S.activeWs = null;
  }
  S.sftp = null;
  S.flyConOpen = false;
}
function serverPagePermission(view) {
  if (view === 'act' && !window.location.pathname.startsWith('/server/')) return '';
  return ({ con: 'console.view', pl: 'console.view', act: 'console.view', fi: 'files.view', plug: 'files.view', bk: 'backup.manage', db: 'database.manage', sc: 'schedules.manage', team: 'subusers.manage', set: 'settings.view', stp: 'settings.view', net: 'settings.view', ver: 'settings.view', sub: 'settings.view', sftp: 'sftp.access' })[view] || '';
}
function hasServerPermission(permission, srv = curServer()) {
  if (!srv?.id || !S.token || !S.user) return false;
  const state = S.serverAccess;
  if (!state || state.key !== accessIdentity(srv) || state.status !== 'ready') return false;
  return state.data.is_admin === true || state.data.is_owner === true || state.data.permissions.includes(permission);
}
function canServerPage(view) {
  const permission = serverPagePermission(view);
  return !permission || hasServerPermission(permission);
}
function canCreateServers() {
  return Boolean(S.user && (S.user.role === 'admin' || S.user.panel_permissions?.includes('servers.create')));
}
async function ensureServerAccess(force = false) {
  const srv = curServer();
  if (!srv?.id || !S.user || !S.token) return false;
  const key = accessIdentity(srv);
  const old = S.serverAccess;
  if (!force && old?.key === key) return old.promise || old.status === 'ready';
  resetServerAccess();
  const state = { key, status: 'loading', data: null, error: '' };
  S.serverAccess = state;
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), 15000);
  state.controller = controller;
  state.promise = (async () => {
    try {
      const data = await api(srvApiPath(srv, '/permissions'), { suppressToast: true, signal: controller.signal });
      if (S.serverAccess !== state || accessIdentity() !== key) return false;
      if (!Array.isArray(data?.permissions) || typeof data.is_admin !== 'boolean' || typeof data.is_owner !== 'boolean') throw new Error('Invalid permission response');
      state.data = data;
      state.status = 'ready';
      return true;
    } catch (error) {
      if (S.serverAccess === state && accessIdentity() === key) { state.status = 'error'; state.error = error.name === 'AbortError' ? 'Permission check timed out. Please retry.' : error.message; }
      return false;
    } finally { clearTimeout(timer); state.promise = null; }
  })();
  return state.promise;
}
async function retryWorkspaceAccess() {
  const view = S.view;
  const pending = ensureServerAccess(true);
  R();
  await pending;
  if (S.view === view) { R(); if (canServerPage(view)) onEnterView(view); }
}
function renderAccessRestricted(server = true) {
  const state = S.serverAccess;
  const loading = server && (!state || state.key !== accessIdentity() || state.status === 'loading');
  return `<section class="card np-access-state" role="status"><div class="np-eyebrow">Workspace access</div><h1>${loading ? 'Checking permissions…' : 'Access restricted'}</h1><p class="mu">${loading ? 'Server pages stay locked until your access is confirmed.' : server && state?.error ? `Could not confirm access: ${esc(state.error)}` : 'Your account does not have permission to open this page. Contact the server owner or a panel administrator.'}</p><div class="row">${server && !loading ? '<button class="btn" onclick="retryWorkspaceAccess()">Check access again</button>' : ''}<a class="btn" href="/" onclick="return handleNavClick(event, \'dash\')">All servers</a></div></section>`;
}
function clientRequestPermission(suffix, method = 'GET') {
  const read = ['GET', 'HEAD'].includes(method.toUpperCase());
  const path = suffix.split('?')[0];
  if (path === '/permissions') return '';
  if (path.startsWith('/files')) {
    if (read) return 'files.view';
    if (/delete|rename|move|unzip/.test(path) || method === 'DELETE') return 'files.delete';
    return 'files.edit';
  }
  if (path.startsWith('/plugins')) return read ? 'files.view' : method === 'DELETE' ? 'files.delete' : 'files.edit';
  if (path.startsWith('/sftp')) return 'sftp.access';
  if (path.startsWith('/backups')) return 'backup.manage';
  if (path.startsWith('/databases')) return 'database.manage';
  if (path.startsWith('/schedules')) return 'schedules.manage';
  if (path.startsWith('/subusers')) return 'subusers.manage';
  if (path.startsWith('/power')) return 'server.power';
  if (path.startsWith('/command')) return 'console.command';
  if (path.startsWith('/players')) return read ? 'console.view' : 'console.command';
  if (!path && read) return ''; // Inventory/status is not a page prefetch.
  return read ? 'settings.view' : 'settings.edit';
}
function assertClientRequestAccess(endpoint, options = {}) {
  const path = String(endpoint).split('?')[0];
  const method = (options.method || 'GET').toUpperCase();
  if (/^(?:\/api\/nodes\/[^/]+\/proxy)?\/api\/servers$/.test(path) && method === 'POST' && !canCreateServers()) throw Object.assign(new Error('Access restricted: server creation permission required'), { status: 403 });
  const match = path.match(/^(?:\/api\/nodes\/([^/]+)\/proxy)?\/api\/servers\/([^/]+)(.*)$/);
  if (!match || match[3] === '/permissions' || S.user?.role === 'admin') return;
  const srv = { id: decodeURIComponent(match[2]), node_id: decodeURIComponent(match[1] || 'local') };
  const permission = clientRequestPermission(match[3], method);
  if ((permission && !hasServerPermission(permission, srv)) || (match[3].startsWith('/files') && !hasServerPermission('files.view', srv))) throw Object.assign(new Error('Access restricted: ' + (permission || 'files.view') + ' required'), { status: 403 });
}
function applyPermissionControls(root) {
  if (!serverPagePermission(S.view)) return;
  const disable = (selector, permission) => {
    if (hasServerPermission(permission)) return;
    root.querySelectorAll(selector).forEach(el => { el.disabled = true; el.setAttribute('aria-disabled', 'true'); el.title = `Access restricted — ${permission} required`; });
  };
  disable('.srv-power-actions [onclick*="power("]', 'server.power');
  disable('[onclick*="toggleFlyConsole("]', 'console.view');
  disable('#cmd, [onclick*="runCmd("], [onclick*="playerAction("], [onclick*="addWhitelistPlayer("]', 'console.command');
  if (['set', 'net', 'sub', 'ver'].includes(S.view)) disable('input, select, textarea, button:not(.nb):not([data-settings-section]):not(.srv-power-actions button):not([onclick*="handleNavClick"]):not([onclick*="copy"])', 'settings.edit');
  if (['fi', 'plug'].includes(S.view)) {
    disable('#ed, #editor-save-btn, [onclick*="promptNew"], [onclick*="upload"], input[type="file"], [onclick*="Compress"], [onclick*="installPlugin"], [onchange*="togglePlugin"]', 'files.edit');
    disable('[onclick*="deleteFile"], [onclick*="deleteSelected"], [onclick*="RenameFile"], [onclick*="MoveFiles"], [onclick*="unzipFile"], [onclick*="unzipSelected"], [onclick*="deletePlugin"]', 'files.delete');
  }
}
V.sftp = () => `${srvh()}<div class="np-section-heading"><div><h2>SFTP</h2><p>One place for external file transfer connections.</p></div></div>${S.sftpLoading ? '<div class="card" role="status">Loading connection details…</div>' : S.sftp ? renderSftpCardHtml(curServer()) : `<div class="card" role="status">${esc(S.sftpError || 'Connection details are not loaded.')} <button class="btn sm" onclick="loadSFTPDetails()">Retry</button></div>`}`;
V.create = () => canCreateServers() ? '<section class="card"><h1>Create server</h1><p class="mu">Choose a node and configure your Minecraft server.</p><button class="btn pr" onclick="openCreateServerModal()">Configure server</button></section>' : renderAccessRestricted(false);

// Custom panel roles never imply built-in admin or server team privileges.
async function loadPanelRoles(skipRender = false) {
  if (S.user?.role !== 'admin') return;
  const key = accessIdentity();
  S.rolesLoading = true;
  try {
    const result = await api('/api/admin/roles', { suppressToast: true });
    if (key !== accessIdentity()) return;
    if (!Array.isArray(result.roles) || !Array.isArray(result.permissions)) throw new Error('Invalid roles response');
    S.panelRoles = result.roles.filter(role => !role.builtin); S.panelRoleCatalog = result.permissions; S.rolesError = '';
  } catch (error) { if (key === accessIdentity()) S.rolesError = error.message; }
  finally { if (key === accessIdentity()) { S.rolesLoading = false; if (!skipRender && S.view === 'admin') R(); } }
}
function renderPanelRoles() {
  if (S.user?.role !== 'admin') return renderAccessRestricted(false);
  if (S.rolesLoading) return '<div class="card" role="status">Loading panel roles…</div>';
  if (S.rolesError || !S.panelRoles) return `<div class="card" role="alert"><h2>Could not load roles</h2><p>${esc(S.rolesError || 'Roles have not loaded.')}</p><button class="btn" onclick="loadPanelRoles()">Retry</button></div>`;
  return `<div class="np-section-heading"><div><h2>Custom panel roles</h2><p>Grant specific panel actions without making someone an Administrator. Server Team access is managed separately.</p></div><button class="btn pr" onclick="openPanelRoleEditor()">${ic('plus', 14)} Create role</button></div><div class="np-role-list">${S.panelRoles.length ? S.panelRoles.map(role => `<section class="card np-role-row"><div><h3>${esc(role.name)}</h3><p class="mu">${esc(role.description || 'No description')}</p><div class="row">${role.permissions.map(p => `<span class="chip">${esc(S.panelRoleCatalog.find(x => x.id === p)?.label || p)}</span>`).join('') || '<span class="mu">No extra panel permissions</span>'}</div></div><div class="row"><button class="btn sm" data-role="${esc(role.id)}" onclick="openPanelRoleEditor(this.dataset.role)">Edit</button><button class="btn sm dg" data-role="${esc(role.id)}" onclick="deletePanelRole(this.dataset.role, this)">Delete</button></div></section>`).join('') : '<div class="card np-empty-state"><h3>No custom roles yet</h3><p class="mu">Create a role to allow trusted members to create servers.</p></div>'}</div>`;
}
function openPanelRoleEditor(id = '') {
  if (S.user?.role !== 'admin' || !S.panelRoleCatalog) return;
  const role = S.panelRoles.find(x => String(x.id) === String(id));
  const container = $('#modal-container');
  container.innerHTML = `<div class="pal"><form class="pbox ui-modal-box" role="dialog" aria-modal="true" aria-labelledby="role-title" onsubmit="savePanelRole(event)" data-role="${esc(id)}"><h2 id="role-title">${role ? 'Edit' : 'Create'} panel role</h2><p class="mu">This role does not grant administrator or server team access.</p><label class="f">Name<input name="name" required maxlength="64" value="${esc(role?.name || '')}"></label><label class="f">Description<textarea name="description" rows="2" maxlength="500">${esc(role?.description || '')}</textarea></label><fieldset class="np-permission-fields"><legend>Panel permissions</legend>${S.panelRoleCatalog.map(p => `<label class="np-permission-choice"><input type="checkbox" name="permissions" value="${esc(p.id)}" ${role?.permissions.includes(p.id) ? 'checked' : ''}><span><b>${esc(p.label)}</b><small>${esc(p.description)}</small></span></label>`).join('')}</fieldset><p class="np-form-error" role="alert"></p><div class="row"><button type="button" class="btn" onclick="closeModal()">Cancel</button><button class="btn pr" type="submit">Save role</button></div></form></div>`;
  container.querySelector('[name=name]')?.focus();
}
async function savePanelRole(event) {
  event.preventDefault();
  if (S.user?.role !== 'admin') return;
  const form = event.currentTarget, button = form.querySelector('[type=submit]');
  if (button.disabled) return;
  const payload = { name: form.elements.name.value.trim(), description: form.elements.description.value.trim(), permissions: [...form.querySelectorAll('[name=permissions]:checked')].map(x => x.value) };
  if (!payload.name) return;
  setBtnLoading(button, true, 'Saving…');
  try {
    await api('/api/admin/roles' + (form.dataset.role ? '/' + encodeURIComponent(form.dataset.role) : ''), { method: form.dataset.role ? 'PUT' : 'POST', body: JSON.stringify(payload), suppressToast: true });
    closeModal(); toast('Panel role saved', 'ok'); await loadPanelRoles();
  } catch (error) { form.querySelector('[role=alert]').textContent = error.message; }
  finally { if (button.isConnected) setBtnLoading(button, false); }
}
async function deletePanelRole(id, button) {
  if (S.user?.role !== 'admin' || button.disabled) return;
  if (!await uiConfirm({ title: 'Delete panel role?', message: 'Assigned roles cannot be deleted. Reassign their users first.', confirmText: 'Delete role', danger: true })) return;
  setBtnLoading(button, true, 'Deleting…');
  try { await api('/api/admin/roles/' + encodeURIComponent(id), { method: 'DELETE', suppressToast: true }); toast('Role deleted', 'ok'); await loadPanelRoles(); }
  catch (error) { toast(error.message, 'bad'); }
  finally { if (button.isConnected) setBtnLoading(button, false); }
}
function renderUserPanelRole(user) {
  if (S.user?.role !== 'admin') return '';
  if (user.role === 'admin') return '<small class="mu np-role-hint">All panel permissions · built-in</small>';
  if (!S.panelRoles || S.rolesError) return `<small class="mu np-role-hint">${esc(user.panel_role_name || 'No custom role')} · roles unavailable</small>`;
  return `<label class="np-user-role"><span class="mu">Panel role</span><select aria-label="Panel role for ${esc(user.username)}" data-user="${esc(user.id)}" data-previous="${esc(user.panel_role_id || '')}" onchange="assignUserPanelRole(this)"><option value="">No custom role</option>${S.panelRoles.map(role => `<option value="${esc(role.id)}" ${String(user.panel_role_id) === String(role.id) ? 'selected' : ''}>${esc(role.name)}</option>`).join('')}</select></label>`;
}
async function assignUserPanelRole(select) {
  if (S.user?.role !== 'admin') return;
  select.disabled = true;
  try {
    await api('/api/admin/users/' + encodeURIComponent(select.dataset.user) + '/panel-role', { method: 'PUT', body: JSON.stringify({ role_id: select.value }), suppressToast: true });
    const user = S.adminUsers.find(x => String(x.id) === select.dataset.user);
    if (user) { user.panel_role_id = select.value; user.panel_role_name = S.panelRoles.find(x => String(x.id) === select.value)?.name || ''; }
    select.dataset.previous = select.value; toast('Panel role assigned', 'ok');
  } catch (error) { select.value = select.dataset.previous; toast(error.message, 'bad'); }
  finally { select.disabled = false; R(); }
}

// Public community content is plain text. Only a validated Discord invite becomes a link.
function safeDiscordInvite(value) {
  try { const u = new URL(value); return u.protocol === 'https:' && !u.username && !u.password && !u.port && ((u.hostname === 'discord.gg' && /^\/[\w-]+\/?$/.test(u.pathname)) || (['discord.com', 'www.discord.com'].includes(u.hostname) && /^\/invite\/[\w-]+\/?$/.test(u.pathname))) ? u.href : ''; } catch (_) { return ''; }
}
async function loadPublicCommunity(skipRender = false) {
  try { const response = await fetch('/api/community'); if (!response.ok) throw new Error('Community unavailable'); S.community = await response.json(); }
  catch (_) { S.community = null; }
  if (!skipRender && !$('#login')) R();
}
function announcementDismissed(id) {
  if (S.dismissedAnnouncement === id) return true;
  try { return localStorage.getItem('neopanel:announcement:' + id) === 'dismissed'; } catch (_) { return false; }
}
function dismissAnnouncement() {
  const a = S.community?.announcement;
  if (!a?.dismissible || !a.id) return;
  S.dismissedAnnouncement = a.id;
  try { localStorage.setItem('neopanel:announcement:' + a.id, 'dismissed'); } catch (_) {}
  R();
}
function renderAnnouncement() {
  const a = S.community?.announcement;
  if (!a?.enabled || !a.text || (a.dismissible && announcementDismissed(a.id))) return '';
  const tone = ['info', 'success', 'warning'].includes(a.tone) ? a.tone : 'info';
  return `<aside class="np-announcement is-${tone}" aria-label="Panel announcement"><span>${esc(a.text)}</span>${a.dismissible && a.id ? '<button class="btn-ico" aria-label="Dismiss announcement" onclick="dismissAnnouncement()">&times;</button>' : ''}</aside>`;
}
function discordLogo() {
  return '<svg viewBox="0 0 24 24" width="28" height="28" aria-hidden="true" fill="currentColor"><path d="M19.7 4.5a18 18 0 0 0-4.5-1.4l-.6 1.2a16 16 0 0 0-5.2 0l-.6-1.2a18 18 0 0 0-4.5 1.4C1.5 8.7.7 12.8 1.1 16.8a18 18 0 0 0 5.5 2.8l1.1-1.8-1.7-.8.4-.3a13 13 0 0 0 11.2 0l.4.3-1.7.8 1.1 1.8a18 18 0 0 0 5.5-2.8c.5-4.7-.8-8.8-3.2-12.3ZM8.4 14.3c-1.1 0-1.9-1-1.9-2.2s.8-2.2 1.9-2.2 1.9 1 1.9 2.2-.8 2.2-1.9 2.2Zm7.2 0c-1.1 0-1.9-1-1.9-2.2s.8-2.2 1.9-2.2 1.9 1 1.9 2.2-.8 2.2-1.9 2.2Z"/></svg>';
}
function renderDiscordCard() {
  const d = S.community?.discord;
  if (!d?.enabled || d.status === 'disabled') return '';
  const invite = d.status === 'ready' ? safeDiscordInvite(d.invite_url) : '';
  return `<aside class="card np-discord-card">${discordLogo()}<div><h2>${esc(d.title || 'Join our Discord')}</h2><p class="mu">${invite ? esc(d.name || 'Our community') + (Number.isFinite(d.member_count) ? ` · ${d.member_count.toLocaleString()} online` : '') : 'Community invite is temporarily unavailable.'}</p></div>${invite ? `<a class="btn pr" href="${esc(invite)}" target="_blank" rel="noopener noreferrer">Join Discord ${ic('external', 13)}</a>` : '<button class="btn" disabled title="No verified Discord invite is available">Invite unavailable</button>'}</aside>`;
}
async function loadCommunitySettings(skipRender = false) {
  if (S.user?.role !== 'admin') return;
  const key = accessIdentity();
  try {
    const data = await api('/api/admin/community', { suppressToast: true });
    if (key !== accessIdentity()) return;
    if (!data.discord || !data.announcement) throw new Error('Invalid community settings response');
    S.communitySettings = data; S.communityError = '';
  } catch (error) { if (key === accessIdentity()) S.communityError = error.message; }
  if (!skipRender && S.view === 'admin') R();
}
function communityPayload(form) {
  const f = form.elements;
  return { discord: { enabled: f.discord_enabled.checked, guild_id: f.guild_id.value.trim(), title: f.discord_title.value.trim() }, announcement: { enabled: f.announcement_enabled.checked, text: f.announcement_text.value.trim(), tone: f.tone.value, dismissible: f.dismissible.checked } };
}
function rememberCommunityDraft(form) {
  S.communityDraftKey = accessIdentity();
  S.communityDraft = communityPayload(form);
}
function renderCommunitySettings() {
  if (S.user?.role !== 'admin') return '';
  if (S.communityError || !S.communitySettings) return `<div class="card" role="alert"><h2>Community settings unavailable</h2><p class="mu">${esc(S.communityError || 'Settings have not loaded.')}</p><button class="btn" onclick="loadCommunitySettings()">Retry</button></div>`;
  const draft = S.communityDraftKey === accessIdentity() ? S.communityDraft : null;
  const { discord: d, announcement: a } = draft || S.communitySettings;
  return `<form class="np-community-form" oninput="rememberCommunityDraft(this)" onchange="rememberCommunityDraft(this)" onsubmit="saveCommunitySettings(event)">
    <fieldset class="np-community-fields" ${S.communitySaving ? 'disabled' : ''}>
      <section class="card">
        <div class="np-section-heading"><div><h2>Discord community</h2><p>A compact custom join card, not an embedded Discord widget.</p></div>${discordLogo()}</div>
        <label class="np-permission-choice"><input name="discord_enabled" type="checkbox" ${d.enabled ? 'checked' : ''}> Show Discord join card</label>
        <div class="grid g2">
          <label class="f">Discord server ID<input name="guild_id" inputmode="numeric" pattern="[0-9]*" value="${esc(d.guild_id)}" aria-describedby="discord-help"></label>
          <label class="f">Card title<input name="discord_title" maxlength="80" value="${esc(d.title)}" placeholder="Join our Discord"></label>
        </div>
        <p id="discord-help" class="mu">In Discord, enable <b>Server Settings → Widget → Enable Server Widget</b>, choose an invite channel, then copy the server ID. An ID is not a join URL. NeoPanel resolves the public widget invite safely; no link is shown when it is unavailable.</p>
        <p class="np-community-diagnostic" role="status">Public status: ${esc(S.community?.discord?.status || 'not loaded')}${S.community?.discord?.status === 'unavailable' ? ' — check the server ID, public widget setting, and invite channel.' : ''}</p>
      </section>
      <section class="card">
        <h2>Panel announcement</h2><p class="mu">Plain text displayed at the top of every workspace page.</p>
        <label class="np-permission-choice"><input name="announcement_enabled" type="checkbox" ${a.enabled ? 'checked' : ''}> Show announcement</label>
        <label class="f">Announcement text<textarea name="announcement_text" rows="3" maxlength="500">${esc(a.text)}</textarea></label>
        <div class="grid g2">
          <label class="f">Tone<select name="tone">${['info', 'success', 'warning'].map(t => `<option value="${t}" ${a.tone === t ? 'selected' : ''}>${t[0].toUpperCase() + t.slice(1)}</option>`).join('')}</select></label>
          <label class="np-permission-choice"><input type="checkbox" name="dismissible" ${a.dismissible ? 'checked' : ''}> Allow dismissal</label>
        </div>
        <p class="mu">An updated announcement ID lets members see a new announcement after dismissing an old one.</p>
      </section>
      <p class="np-form-error" role="alert">${esc(S.communitySaveError || '')}</p>
      <div class="row"><button type="submit" class="btn pr">${S.communitySaving ? 'Saving…' : 'Save community settings'}</button><button type="button" class="btn" onclick="loadPublicCommunity()">Refresh public status</button></div>
    </fieldset>
  </form>`;
}
async function saveCommunitySettings(event) {
  event.preventDefault();
  if (S.user?.role !== 'admin' || S.communitySaving) return;
  const form = event.currentTarget;
  const payload = communityPayload(form), key = accessIdentity();
  const errorBox = form.querySelector('[role=alert]');
  if ((payload.discord.enabled || payload.discord.guild_id) && !/^\d{17,20}$/.test(payload.discord.guild_id)) { errorBox.textContent = 'Enter a Discord server ID containing 17–20 digits.'; return; }
  if (payload.announcement.enabled && !payload.announcement.text) { errorBox.textContent = 'Enter announcement text or turn the announcement off.'; return; }
  S.communitySaving = true; S.communitySaveError = ''; rememberCommunityDraft(form); R();
  try {
    const saved = await api('/api/admin/community', { method: 'PUT', body: JSON.stringify(payload), suppressToast: true });
    if (key !== accessIdentity()) return;
    S.communitySettings = saved; S.communityDraft = null;
    await loadPublicCommunity(true); toast('Community settings saved', 'ok');
  } catch (error) { if (key === accessIdentity()) S.communitySaveError = error.message; }
  finally { S.communitySaving = false; if (key === accessIdentity()) R(); }
}
function organizeAppearance(html) {
  const sections = ['Brand & theme', 'Public URL', 'Community', 'Server defaults', 'Security & maintenance'];
  const selected = S.appearanceSection || 0;
  const tabs = `<div class="tabs np-settings-tabs" aria-label="Appearance sections">${sections.map((label, index) => `<button class="${selected === index ? 'on' : ''}" aria-pressed="${selected === index}" onclick="S.appearanceSection=${index};R()">${label}</button>`).join('')}</div>`;
  if (selected === 2) return tabs + renderCommunitySettings();
  const template = document.createElement('template'); template.innerHTML = html;
  const visible = [...template.content.children].filter(el => {
    const text = el.textContent;
    const group = el.classList.contains('brand-layout-grid') ? 0 : el.querySelector('#adm-fqdn') ? 3 : /Turnstile|Maintenance|Client Area/.test(text) ? 4 : 1;
    return group === selected;
  }).map(el => el.outerHTML).join('');
  return tabs + visible;
}

// Move existing controls, never clone them: autosave keeps a complete form snapshot.
function organizeServerSettings(root) {
  const identity = root.querySelector('#set-srv-name')?.closest('.card');
  if (!identity) return;
  const wrapper = identity.parentElement.parentElement;
  const groups = [
    ['General', [identity, root.querySelector('#prop-motd')?.closest('.card') || root.querySelector('[data-steel-config]')]],
    ['Startup & resources', [root.querySelector('#prop-java')?.closest('.card'), root.querySelector('#prop-jar')?.closest('.card')]],
    ['Administration', [...wrapper.children].filter(el => el.classList.contains('card'))]
  ];
  const selected = Math.min(S.serverSettingsSection || 0, 2);
  const nav = document.createElement('div'); nav.className = 'tabs np-settings-tabs'; nav.setAttribute('aria-label', 'Server settings sections');
  groups.forEach(([label], i) => {
    const button = document.createElement('button'); button.type = 'button'; button.textContent = label; button.dataset.settingsSection = String(i); button.className = i === selected ? 'on' : ''; button.setAttribute('aria-pressed', String(i === selected));
    button.onclick = () => { S.serverSettingsSection = i; root.querySelectorAll('[data-settings-panel]').forEach(panel => { panel.hidden = Number(panel.dataset.settingsPanel) !== i; }); nav.querySelectorAll('button').forEach((b, j) => { b.classList.toggle('on', i === j); b.setAttribute('aria-pressed', String(i === j)); }); };
    nav.append(button);
  });
  wrapper.prepend(nav);
  groups.forEach(([, cards], i) => {
    const panel = document.createElement('section'); panel.className = 'np-server-settings-group'; panel.dataset.settingsPanel = String(i); panel.hidden = i !== selected;
    cards.filter(Boolean).forEach(card => panel.append(card)); wrapper.append(panel);
  });
  wrapper.querySelectorAll(':scope > .grid, :scope > .stp-top-grid').forEach(el => { if (!el.children.length) el.remove(); });
}

function displayNodeLocation(node) {
  const location = String(node?.location || '').trim();
  if (location && location !== node.id) return location;
  return node?.id === 'local' ? 'This machine' : location ? displayNodeName(node.id, node.name) : 'Location not specified';
}
