// NeoPanel — Ultra-Lightweight Minecraft Server Manager & Daemon SPA
// Zero external JS dependencies. Pure-Go backend integration.

// --- High-Quality Vector SVG Icon Generator ---
function ic(name, size = 18) {
  const p = {
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
    viewTiles: `<rect x="3" y="4" width="18" height="6.5" rx="1.2"/><rect x="3" y="13.5" width="8" height="6.5" rx="1.2"/><rect x="13" y="13.5" width="8" height="6.5" rx="1.2"/>`
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
  if (ext === 'jar') {
    return `<span class="fi-badge jar${cls}" title="Java Archive (.jar)">${ic('package', iconSize)}</span>`;
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
  brand_subtitle: 'Ultra-Lightweight Minecraft Server Manager',
  logo_url: '',
  favicon_url: '',
  login_bg_url: '',
  server_hero_url: SERVER_HERO_URL,
  card_cover_url: MINECRAFT_COVER_URL,
  default_avatar_url: DEFAULT_USER_AVATAR_URL,
  footer_text: 'Powered by NeoPanel Daemon',
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
  custom_css: ''
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
  view: 'dash',              // active view: dash, con, set, act, fi, db, bk, sc, net, sub, stp, pl, plug, ver, admin, acc
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
  ni: Array.from({ length: 40 }, () => 0),
  no: Array.from({ length: 40 }, () => 0),
  tps: 20.00,
  ver: '1.21.4',
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
  selectedFiles: new Set(),
  fileMenuOpen: null,
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

// --- Clean Multi-Page Routing Helpers ---
const VIEW_TO_SLUG = {
  dash: '',
  con: 'console',
  set: 'settings',
  act: 'activity',
  fi: 'files',
  db: 'databases',
  bk: 'backups',
  sc: 'schedules',
  net: 'network',
  sub: 'subdomain',
  stp: 'startup',
  pl: 'players',
  plug: 'plugins',
  ver: 'versions',
  admin: 'admin',
  acc: 'account',
  bill: 'billing'
};

const SLUG_TO_VIEW = Object.fromEntries(
  Object.entries(VIEW_TO_SLUG).filter(([, slug]) => slug).map(([k, slug]) => [slug, k])
);
SLUG_TO_VIEW.store = 'plug';
SLUG_TO_VIEW.servers = 'dash';

const ADMIN_TAB_SLUGS = ['health', 'nodes', 'servers', 'users', 'branding', 'audit'];
const ACC_TAB_SLUGS = ['general', 'security', 'ssh', 'api', 'activity'];

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

function serverRouteUrl(nodeId, srvId, view = 'con', opts = {}) {
  const nid = nodeId || 'local';
  const target = normalizeRoute(view);
  if (target === 'plug') {
    const tab = opts.plugTab || S.plugTab || 'browse';
    return `/server/${encodeURIComponent(nid)}/${encodeURIComponent(srvId)}/plugins/${tab === 'installed' ? 'installed' : 'store'}`;
  }
  const slug = VIEW_TO_SLUG[target] || 'console';
  return `/server/${encodeURIComponent(nid)}/${encodeURIComponent(srvId)}/${slug}`;
}

function routeUrl(view, opts = {}) {
  const target = normalizeRoute(view);
  if (target === 'dash') return '/';
  if (target === 'bill') return '/billing';
  if (target === 'admin') {
    const t = opts.adminTab !== undefined ? opts.adminTab : (S.adminTab || 0);
    const slug = ADMIN_TAB_SLUGS[t] || 'health';
    return t === 0 ? '/admin' : `/admin/${slug}`;
  }
  if (target === 'acc') {
    const t = opts.acctTab !== undefined ? opts.acctTab : (S.acct?.tab || 0);
    const slug = ACC_TAB_SLUGS[t] || 'general';
    return t === 0 ? '/account' : `/account/${slug}`;
  }
  if (target === 'act' && (opts.global || (!opts.srv && !window.location.pathname.startsWith('/server/') && S.view === 'dash'))) {
    return '/activity';
  }
  const srv = opts.srv || curServer();
  if (!srv || !srv.id) {
    return target === 'act' ? '/activity' : '/';
  }
  return serverRouteUrl(srv.node_id || 'local', srv.id, target, opts);
}

function parseCurrentRoute(pathname = window.location.pathname, hashStr = window.location.hash) {
  const cleanPath = String(pathname || '/').replace(/\/+$/, '') || '/';
  const rawHash = String(hashStr || '').replace(/^#/, '').trim();

  if (cleanPath === '/' || cleanPath === '/index.html') {
    if (rawHash) {
      const hView = normalizeRoute(rawHash);
      if (hView && (V[hView] || ['admin', 'acc', 'dash', 'act', 'bill'].includes(hView))) {
        return { view: hView, plugTab: rawHash === 'store' ? 'browse' : undefined, legacyHash: true };
      }
    }
    return { view: 'dash' };
  }
  if (cleanPath === '/servers') return { view: 'dash' };
  if (cleanPath === '/activity') return { view: 'act', global: true };
  if (cleanPath === '/billing') return { view: 'bill' };

  if (cleanPath === '/admin' || cleanPath.startsWith('/admin/')) {
    const sub = cleanPath.split('/')[2] || 'health';
    const idx = ADMIN_TAB_SLUGS.indexOf(sub);
    return { view: 'admin', adminTab: idx >= 0 ? idx : 0 };
  }
  if (cleanPath === '/account' || cleanPath.startsWith('/account/')) {
    const sub = cleanPath.split('/')[2] || 'general';
    const idx = ACC_TAB_SLUGS.indexOf(sub);
    return { view: 'acc', acctTab: idx >= 0 ? idx : 0 };
  }
  if (cleanPath.startsWith('/server/')) {
    const parts = cleanPath.split('/').filter(Boolean);
    const nodeId = decodeURIComponent(parts[1] || 'local');
    const serverId = decodeURIComponent(parts[2] || '');
    const slug = (parts[3] || 'console').toLowerCase();
    const extra = (parts[4] || '').toLowerCase();
    if (slug === 'plugins') {
      return {
        view: 'plug',
        nodeId,
        serverId,
        plugTab: extra === 'installed' ? 'installed' : 'browse'
      };
    }
    const mappedView = SLUG_TO_VIEW[slug] || 'con';
    return { view: mappedView, nodeId, serverId };
  }
  return { view: 'dash' };
}

function getViewDisplayLabel(view) {
  const map = {
    dash: 'Servers Overview',
    con: 'Server Console',
    set: 'Server Settings',
    act: 'Activity Log',
    fi: 'File Manager',
    db: 'Databases',
    bk: 'Backups',
    sc: 'Schedules',
    net: 'Network Allocations',
    sub: 'Subdomains',
    stp: 'Startup Settings',
    pl: 'Players',
    plug: 'Plugins & Store',
    ver: 'Software & Versions',
    admin: 'Admin Area',
    acc: 'Account Settings',
    bill: 'Billing & Resources'
  };
  return map[view] || 'NeoPanel';
}

function navigateToUrl(url, label = 'Page') {
  if (!url) return;
  closeAllCtx();
  closeFloatingCtx();
  closeAllCustomSelects();
  S.cmenu = false;
  S.navigating = true;
  startPageProgress(32);
  const main = $('#main');
  if (main) {
    main.innerHTML = `<div class="np-page-loader" role="status" aria-live="polite">
      <div class="np-spinner"></div>
      <span>Loading ${esc(label)}...</span>
    </div>`;
  }
  window.location.assign(url);
}

function handleNavClick(e, view, opts = {}) {
  if (e && (e.ctrlKey || e.metaKey || e.shiftKey || e.altKey || e.button !== 0)) {
    return true;
  }
  if (e) e.preventDefault();
  const target = normalizeRoute(view);
  if (opts.srv) saveActiveServerRef(opts.srv);
  else if (curServer()) saveActiveServerRef(curServer());
  if (target === 'plug' && view === 'store') opts.plugTab = 'browse';
  navigateToUrl(routeUrl(target, opts), getViewDisplayLabel(target));
  return false;
}

function handleServerCardClick(e, index, targetView = 'con') {
  if (e && (e.ctrlKey || e.metaKey || e.shiftKey || e.altKey || e.button !== 0)) {
    return true;
  }
  if (e) e.preventDefault();
  const srv = S.servers[index];
  if (!srv) return false;
  saveActiveServerRef(srv);
  navigateToUrl(serverRouteUrl(srv.node_id || 'local', srv.id, targetView), srv.name || 'Server Console');
  return false;
}

// --- Custom Modal Dialog System (Replacing browser default alert, confirm, prompt) ---
let _uiModalResolver = null;

function closeUiModal(value) {
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
  return `${Math.floor(s / 3600)}h ${Math.floor((s % 3600) / 60)}m ${s % 60}s`;
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

  S.memUsedMB = memMB;
  S.memMaxMB = maxMemMB;

  if (pushSeries) {
    S.cpu.push(cpuVal);
    S.cpu.shift();
    S.ram.push(ramPct);
    S.ram.shift();
    S.ni.push(netIn);
    S.ni.shift();
    S.no.push(netOut);
    S.no.shift();
  } else {
    S.cpu[S.cpu.length - 1] = cpuVal;
    S.ram[S.ram.length - 1] = ramPct;
    S.ni[S.ni.length - 1] = netIn;
    S.no[S.no.length - 1] = netOut;
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
}

// --- Sparklines SVGs ---
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

function head(name) {
  let h = 0;
  for (const c of String(name || 'Player')) h = (h * 31 + c.charCodeAt(0)) % 360;
  return `<span class="head" style="background:hsl(${h} 45% 48%);box-shadow:inset 0 -9px 0 hsl(${h} 35% 30%)"></span>`;
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
  const letter = (S.acct.fn ? S.acct.fn[0] : (S.acct.un ? S.acct.un[0] : 'A')).toUpperCase();
  const hasCustomGravatar = S.acct.h && S.acct.em && S.acct.em.toLowerCase() !== 'admin@neopanel.internal';
  const defAvatar = (typeof getBrand === 'function' ? getBrand('default_avatar_url') : '') || DEFAULT_USER_AVATAR_URL;
  return `<span class="av" style="width:${s}px;height:${s}px;font-size:${s * 0.4}px">
    ${esc(letter)}
    <img src="${esc(defAvatar)}" alt="" onerror="this.remove()">
    ${hasCustomGravatar ? `<img src="https://gravatar.com/avatar/${S.acct.h}?s=${s * 2}&d=404" alt="" onerror="this.remove()">` : ''}
  </span>`;
}

// --- Branding, Theme Presets & Multi-Node Helpers ---
function getBrand(key) {
  const active = (S.view === 'admin' && (S.adminTab || 0) === 4 && S.brandDraft) ? S.brandDraft : (S.branding || DEFAULT_BRANDING);
  const val = active && active[key] !== undefined ? active[key] : DEFAULT_BRANDING[key];
  if ((key === 'server_hero_url' || key === 'card_cover_url' || key === 'default_avatar_url' || key === 'brand_name') && !String(val || '').trim()) {
    return DEFAULT_BRANDING[key];
  }
  return val;
}

function hexToRgbStr(hex) {
  const m = /^#([0-9a-fA-F]{6})$/.exec(String(hex || '').trim());
  if (!m) return null;
  const n = parseInt(m[1], 16);
  return `${(n >> 16) & 255}, ${(n >> 8) & 255}, ${n & 255}`;
}

function brandLogoHtml(forLogin = false, override = null) {
  const b = override || ((S.view === 'admin' && (S.adminTab || 0) === 4 && S.brandDraft) ? S.brandDraft : (S.branding || DEFAULT_BRANDING));
  const name = (b.brand_name || 'NeoPanel').trim() || 'NeoPanel';
  const logo = (b.logo_url || '').trim();
  const iconHtml = logo
    ? `<img src="${esc(logo)}" class="brand-logo-img" alt="" onerror="this.outerHTML='<i class=\\'cube\\'></i>'">`
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

  const favEl = $('#app-favicon');
  if (favEl) {
    favEl.href = (b.favicon_url || '').trim() || '/favicon.svg';
  }

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

function findServerRef(srvId, nodeId = null) {
  if (!srvId) return null;
  if (nodeId) {
    return (S.servers || []).find(s => s.id === srvId && (s.node_id || 'local') === nodeId) || null;
  }
  const cur = curServer();
  if (cur && cur.id === srvId) return cur;
  return (S.servers || []).find(s => s.id === srvId) || null;
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
  if (navigator.clipboard) {
    navigator.clipboard.writeText(val).then(() => {
      toast(`${label} copied to clipboard!`, 'ok');
    }).catch(() => {
      uiPrompt({ title: `Copy ${label}`, message: 'Copy the value below:', defaultValue: val, confirmText: 'Done' });
    });
  } else {
    uiPrompt({ title: `Copy ${label}`, message: 'Copy the value below:', defaultValue: val, confirmText: 'Done' });
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
  const headers = options.headers || {};
  if (S.token) headers['Authorization'] = `Bearer ${S.token}`;
  if (!headers['Content-Type'] && !(options.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
  }

  try {
    const res = await fetch(endpoint, { ...options, headers });
    if (res.status === 401) {
      logout();
      throw new Error('Session expired. Please sign in again.');
    }
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      throw new Error(data.error || `HTTP ${res.status}`);
    }
    return data;
  } catch (err) {
    if (err.message && !err.message.includes('Session expired')) {
      toast(err.message, 'error');
    }
    throw err;
  }
}

function curServer() {
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
  const isOnline = S.status === 'online';
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
  if (startBtn) startBtn.disabled = S.status !== 'offline';
  const restartBtn = document.querySelector('[data-power-btn="restart"]');
  if (restartBtn) restartBtn.disabled = S.status !== 'online';
  const stopBtn = document.querySelector('[data-power-btn="stop"]');
  if (stopBtn) stopBtn.disabled = !(S.status === 'online' || S.status === 'starting');
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
  const showTrigger = !['dash', 'admin', 'acc'].includes(S.view);
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
          <input id="fly-cmd" autocomplete="off" spellcheck="false" placeholder="Type a command..." aria-label="Quick console command">
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

// Local 1-second uptime ticker fallback so uptime advances smoothly between WS ticks
setInterval(() => {
  if ((S.status === 'online' || S.status === 'starting') && (!S.activeWs || S.activeWs.readyState !== WebSocket.OPEN)) {
    S.uptimeSec = (Number(S.uptimeSec) || 0) + 1;
    if (S.view === 'con') liveCon();
    else if (S.view === 'dash') updateDashLiveStats();
  }
}, 1000);

// --- Power Actions ---
async function power(action, skipConfirm = false) {
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
        title: 'Force Kill Server Process',
        message: `Are you sure you want to <b>force kill</b> <b>${esc(srv.name)}</b>? This immediately terminates the Java process without saving chunks and may cause world data loss.`,
        confirmText: 'Force Kill',
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
      modalTitle = 'Server Executable JAR Not Found';
      modalMsg = `<div style="line-height:1.6;">
        The server executable JAR file <b>${esc(srv.jar_file || 'server.jar')}</b> could not be found in the server directory.<br><br>
        <div style="background:rgba(239,68,68,0.1); border:1px solid rgba(239,68,68,0.3); border-radius:8px; padding:10px 12px; margin-bottom:12px; font-size:13px; color:#fca5a5;">
          ${esc(err.message)}
        </div>
        <b>How to resolve:</b>
        <ol style="margin: 8px 0 0 18px; padding: 0; line-height: 1.6;">
          <li>Click <b>Reinstall Server</b> in Settings to automatically download an official build from mcjarfiles.com.</li>
          <li>Or navigate to <b>Files</b> and upload your JAR as <code>${esc(srv.jar_file || 'server.jar')}</code>.</li>
          <li>Or verify your <b>Server Executable File</b> name in <b>Startup Settings</b>.</li>
        </ol>
      </div>`;
    } else if (lower.includes('java')) {
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
  ['net', 'Network', 'network', 'Management'],
  ['sub', 'Subdomain', 'link', 'Management'],
  ['stp', 'Startup', 'play', 'Management'],
  ['pl', 'Players', 'users', 'Management'],
  ['plug', 'Plugins & Store', 'puzzle', 'Minecraft'],
  ['ver', 'Software & Versions', 'cpu', 'Minecraft']
];

const CMDS = [
  ['Save the world', 'save-all'],
  ['Say hello', 'say Hello from NeoPanel!'],
  ['List players', 'list'],
  ['Check TPS', 'tps']
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
  ['BungeeCord', 'Classic multi-server proxy']
];

const FAM = [
  ['26.3', 'Wilderness Bound', '2026', ['26.3']],
  ['26.2', 'Chaos Cubed', '2026', ['26.2']],
  ['26.1', 'Tiny Takeover', '2026', ['26.1.2', '26.1.1', '26.1']],
  ['1.21', 'Tricky Trials & Drops', '2024–2025', ['1.21.11', '1.21.10', '1.21.9', '1.21.8', '1.21.7', '1.21.6', '1.21.5', '1.21.4', '1.21.3', '1.21.2', '1.21.1', '1.21']],
  ['1.20', 'Trails & Tales', '2023–2024', ['1.20.6', '1.20.5', '1.20.4', '1.20.3', '1.20.2', '1.20.1', '1.20']],
  ['1.19', 'The Wild Update', '2022–2023', ['1.19.4', '1.19.3', '1.19.2', '1.19.1', '1.19']],
  ['1.18', 'Caves & Cliffs: Part II', '2021–2022', ['1.18.2', '1.18.1', '1.18']],
  ['1.17', 'Caves & Cliffs: Part I', '2021', ['1.17.1', '1.17']],
  ['1.16', 'Nether Update', '2020–2021', ['1.16.5', '1.16.4', '1.16.3', '1.16.2', '1.16.1', '1.16']],
  ['1.15', 'Buzzy Bees', '2019–2020', ['1.15.2', '1.15.1', '1.15']],
  ['1.14', 'Village & Pillage', '2019', ['1.14.4', '1.14.3', '1.14.2', '1.14.1', '1.14']],
  ['1.13', 'Update Aquatic', '2018', ['1.13.2', '1.13.1', '1.13']],
  ['1.12', 'World of Color', '2017', ['1.12.2', '1.12.1', '1.12']],
  ['1.11', 'Exploration Update', '2016–2017', ['1.11.2', '1.11.1', '1.11']],
  ['1.10', 'Frostburn Update', '2016', ['1.10.2', '1.10.1', '1.10']],
  ['1.9', 'Combat Update', '2016', ['1.9.4', '1.9.3', '1.9.2', '1.9.1', '1.9']],
  ['1.8', 'Bountiful Update', '2014–2015', ['1.8.9', '1.8.8', '1.8.7', '1.8.6', '1.8.5', '1.8.4', '1.8.3', '1.8.2', '1.8.1', '1.8']],
  ['1.7', 'Update that Changed the World', '2013–2014', ['1.7.10', '1.7.9', '1.7.8', '1.7.7', '1.7.6', '1.7.5', '1.7.4', '1.7.2']],
  ['1.0', 'Early Releases (1.0 – 1.6)', '2011–2013', ['1.6.4', '1.6.2', '1.6.1', '1.5.2', '1.5.1', '1.5', '1.4.7', '1.4.6', '1.4.5', '1.4.4', '1.4.2', '1.3.2', '1.3.1', '1.2.5', '1.2.4', '1.2.3', '1.2.2', '1.2.1', '1.1', '1.0']]
];

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
  const cat = S.versionsCatalog;
  if (eng === 'velocity') {
    const base = ['3.4.0-SNAPSHOT', '3.3.0-SNAPSHOT', '3.2.0-SNAPSHOT', '3.1.1'];
    const live = Array.isArray(cat?.by_engine?.velocity) ? cat.by_engine.velocity : [];
    return Array.from(new Set([...live, ...base]));
  }
  if (eng === 'waterfall') {
    const base = ['1.20', '1.19', '1.18', '1.17', '1.16', '1.15', '1.14', '1.13', '1.12'];
    const live = Array.isArray(cat?.by_engine?.waterfall) ? cat.by_engine.waterfall : [];
    return Array.from(new Set([...live, ...base])).sort(compareMcVersionsDesc);
  }
  const merged = new Set(FALLBACK_ALL_MC);
  if (Array.isArray(cat?.all_mc)) {
    for (const v of cat.all_mc) if (v) merged.add(String(v));
  }
  if (eng && Array.isArray(cat?.by_engine?.[eng])) {
    for (const v of cat.by_engine[eng]) if (v) merged.add(String(v));
  }
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
    const curVal = $('#cs-ver')?.value?.trim() || csSel.value || '1.21.4';
    const vers = getAllMcVersions(engineCS);
    const list = (curVal && !vers.includes(curVal)) ? [curVal, ...vers] : vers;
    csSel.innerHTML = list.map(v => `<option value="${esc(v)}" ${v === curVal ? 'selected' : ''}>${esc(v)}</option>`).join('');
    if (typeof csSel._cselectRebuild === 'function') csSel._cselectRebuild();
  }
  const reinstSel = $('#reinst-ver-sel');
  if (reinstSel) {
    const engR = $('#reinst-software')?.value || 'paper';
    const curVal = $('#reinst-version')?.value?.trim() || reinstSel.value || '1.21.4';
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
  const stpSel = $('#stp-ver-sel');
  if (stpSel) {
    const srv = curServer();
    const curVal = $('#stp-ver')?.value?.trim() || srv?.version || '1.21.4';
    const vers = getAllMcVersions(srv?.type || S.core || '');
    const list = (curVal && !vers.includes(curVal)) ? [curVal, ...vers] : vers;
    stpSel.innerHTML = list.map(v => `<option value="${esc(v)}" ${v === curVal ? 'selected' : ''}>${esc(v)}</option>`).join('');
    if (typeof stpSel._cselectRebuild === 'function') stpSel._cselectRebuild();
  }
  if (S.view === 'ver' && $('#vres')) {
    renderVersionRows();
  }
}

async function loadVersionsCatalog(force = false) {
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
  for (const [id, nid] of [['stp-java', curServer()?.node_id || defaultNid], ['prop-java', curServer()?.node_id || defaultNid], ['cs-java', csNodeVal]]) {
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
  const nodeLabel = srv?.node_name || getNodeById(nodeId)?.name || (nodeId === 'local' ? 'Local Master Node' : nodeId);
  const heroUrl = getBrand('server_hero_url');

  return `<div class="card srvh" style="background-image: url('${esc(heroUrl)}'), url('/img/banner.jpg');">
    <div>
      <div class="row">
        <h1>${esc(S.props.motd || srv?.name || 'Minecraft Server')}</h1>
        <span id="srv-badge" class="badge ${st}"><i class="dot"></i>${{ online: 'Running', offline: 'Offline', starting: 'Starting', stopping: 'Stopping' }[st] || 'Offline'}</span>
        ${((S.nodes && S.nodes.length > 1) || nodeId !== 'local') ? `<span class="chip srv-node-badge" title="Running on node ${esc(nodeLabel)}">${ic('network', 11)} ${esc(nodeLabel)}</span>` : ''}
        <span id="srv-halt-badge" class="badge halted" style="display:${st === 'offline' && S.startupHalted ? 'inline-flex' : 'none'};" title="Auto-restart stopped after 2 consecutive startup failures">Start Halted (2/2)</span>
      </div>
      <div id="hst" class="row hst">
        <span class="mono mu pv">${esc(hostAddr)}</span>
        <span class="mono mu">CPU ${c.toFixed(1)}%</span>
        <span class="mono mu">RAM ${m.toFixed(0)} MiB / ${(maxRamMB / 1024).toFixed(1)} GiB</span>
      </div>
    </div>
    <div class="row" style="position:relative">
      ${S.view !== 'con' ? `<button class="btn" onclick="toggleFlyConsole()" title="Open Live Console Drawer (Ctrl+\`)">${ic('terminal', 14)} Console</button>` : ''}
      <button class="btn" onclick="event.stopPropagation(); toggleCmdMenu()" aria-haspopup="menu" aria-expanded="${!!S.cmenu}">Commands &#9662;</button>
      ${S.cmenu ? `<div class="cmenu" role="menu" onclick="event.stopPropagation()">
        ${CMDS.map(([label, cmd]) => `<button role="menuitem" onclick="execQuickCmd('${cmd}')">${label}</button>`).join('')}
        <button role="menuitem" onclick="S.cmenu=false; toggleFlyConsole();">${ic('terminal', 13)} Toggle Live Console (Ctrl+\`)</button>
        <div style="border-top: 1px solid var(--border); margin: 4px 0;"></div>
        <button role="menuitem" style="color: #ef4444;" data-srv-id="${esc(srv?.id || '')}" data-srv-name="${esc(srv?.name || '')}" data-node-id="${esc(nodeId)}" onclick="S.cmenu=false; openDeleteServerModal(this.dataset.srvId, this.dataset.srvName, this.dataset.nodeId);">${ic('trash', 13)} Delete Server</button>
      </div>` : ''}
      <button class="btn pr" data-power-btn="start" onclick="power('start')" ${st !== 'offline' ? 'disabled' : ''}>${ic('play', 14)} Start</button>
      <button class="btn" data-power-btn="restart" onclick="power('restart')" ${on ? '' : 'disabled'}>${ic('refresh', 14)} Restart</button>
      <button class="btn stop" data-power-btn="stop" onclick="power('stop')" ${on || st === 'starting' ? '' : 'disabled'}>Stop</button>
      <button class="btn dg" onclick="power('kill')">Kill</button>
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

  return row(ic('network', 16), 'Address', `<span class="pv">${esc(hostAddr)}</span>`) +
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
  const z = a => a.map(() => 0);
  const netPeak = Math.max(10, ...S.ni, ...S.no);
  return `<div class="card ch">
    <div class="row sb"><span class="mu">CPU load</span><span class="mu mono">${active ? (S.cpu.at(-1) || 0).toFixed(1) + '%' : '100%'}</span></div>
    ${spark(active ? S.cpu : z(S.cpu), 100)}
  </div>
  <div class="card ch">
    <div class="row sb"><span class="mu">Memory</span><span class="mu mono">${active ? `${(S.memUsedMB || 0).toFixed(0)} MiB / ${maxRamGB} GiB` : `${maxRamGB} GiB`}</span></div>
    ${spark(active ? S.ram : z(S.ram), 100)}
  </div>
  <div class="card ch">
    <div class="row sb"><span class="mu">Network I/O</span><span class="mu"><span style="color:var(--ac)">in</span> <span style="color:#f0b429">out</span></span></div>
    ${lines(active ? [S.ni, S.no] : [z(S.ni), z(S.no)], netPeak)}
  </div>`;
}

function liveCon() {
  const stc = $('#stc');
  const chr = $('#chr');
  const hst = $('#hst');
  if (stc) stc.innerHTML = statsHtml();
  if (chr) chr.innerHTML = chartsHtml();
  if (hst) {
    const srv = curServer();
    const active = S.status === 'online' || S.status === 'starting';
    const c = active ? (S.cpu.at(-1) || 0) : 0;
    const maxRamMB = S.memMaxMB || getServerMaxRamMB(srv);
    const m = active ? (S.memUsedMB || ((S.ram.at(-1) || 0) / 100) * maxRamMB) : 0;
    const hostAddr = `${srvHost(srv)}:${srv ? srv.port : 25565}`;
    hst.innerHTML = `<span class="mono mu pv">${esc(hostAddr)}</span><span class="mono mu">CPU ${c.toFixed(1)}%</span><span class="mono mu">RAM ${m.toFixed(0)} MiB / ${(maxRamMB / 1024).toFixed(1)} GiB</span>`;
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
      badge.innerHTML = `<i class="dot"></i>${isRun ? 'Running' : 'Offline'}${isRun ? `<span class="mu" style="font-weight: 400; margin-left: 3px;">&bull; Up ${formatUptime(upVal)}</span>` : ''}`;
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
  if (runSummary) runSummary.innerHTML = `${runningCount} &bull; ${totalCount} running`;

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
  power_kill: { category: 'power', badgeLabel: 'Power', cls: 'act-cat-power', title: 'Force Killed Server' },
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
  file_unzip: { category: 'files', badgeLabel: 'Files', cls: 'act-cat-files', title: 'Extracted Archive' },
  plugin_install: { category: 'plugins', badgeLabel: 'Plugins', cls: 'act-cat-plugins', title: 'Installed Plugin' },
  plugin_delete: { category: 'plugins', badgeLabel: 'Plugins', cls: 'act-cat-plugins', title: 'Uninstalled Plugin' },
  software_install: { category: 'software', badgeLabel: 'Software', cls: 'act-cat-software', title: 'Installed Engine JAR' },
  properties_save: { category: 'software', badgeLabel: 'Config', cls: 'act-cat-software', title: 'Saved server.properties' },
  backup_create: { category: 'backups', badgeLabel: 'Backups', cls: 'act-cat-backups', title: 'Created Backup' },
  backup_delete: { category: 'backups', badgeLabel: 'Backups', cls: 'act-cat-backups', title: 'Deleted Backup' },
  backup_download: { category: 'backups', badgeLabel: 'Backups', cls: 'act-cat-backups', title: 'Downloaded Backup' },
  port_add: { category: 'network', badgeLabel: 'Network', cls: 'act-cat-network', title: 'Assigned Port' },
  port_delete: { category: 'network', badgeLabel: 'Network', cls: 'act-cat-network', title: 'Removed Port' },
  port_primary: { category: 'network', badgeLabel: 'Network', cls: 'act-cat-network', title: 'Changed Primary Port' },
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
  ['plugins', 'Plugins'],
  ['software', 'Software & Config'],
  ['backups', 'Backups'],
  ['network', 'Network & Ports'],
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
  if (key.startsWith('plugin_')) return { category: 'plugins', badgeLabel: 'Plugins', cls: 'act-cat-plugins', title: key.replace(/_/g, ' ') };
  if (key.startsWith('backup_')) return { category: 'backups', badgeLabel: 'Backups', cls: 'act-cat-backups', title: key.replace(/_/g, ' ') };
  if (key.startsWith('port_')) return { category: 'network', badgeLabel: 'Network', cls: 'act-cat-network', title: key.replace(/_/g, ' ') };
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

function formatFullDate(iso) {
  if (!iso) return '—';
  const d = new Date(iso);
  if (isNaN(d.getTime())) return String(iso);
  return d.toLocaleString();
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
    const showNodeBar = nodesList.length > 1 || errEntries.length > 0;
    const activeNodeFilter = S.nodeFilter || 'all';

    const indexedServers = S.servers.map((s, origIdx) => ({ s, origIdx }));
    const visibleServers = activeNodeFilter === 'all'
      ? indexedServers
      : indexedServers.filter(({ s }) => (s.node_id || 'local') === activeNodeFilter);
    const coverUrl = getBrand('card_cover_url');
    const dashVm = getViewMode('dash', 'tiles');
    const gridCls = dashVm === 'list' ? 'vm-dash-list' : (dashVm === 'icons' ? 'vm-dash-icons' : '');
    const gridStyle = dashVm === 'tiles' ? 'display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 16px;' : '';

    return `
    <div class="row sb" style="margin-bottom: 20px; align-items: flex-end; flex-wrap: wrap; gap: 16px;">
      <div>
        <h1 style="font-size: 26px; font-weight: 800; color: var(--tx); margin-bottom: 4px; letter-spacing: -0.02em;">Servers</h1>
        <span id="dash-running-summary" class="mu" style="font-size: 13px; font-weight: 500;">${runningCount} &bull; ${totalCount} running${nodesList.length > 1 ? ` &bull; ${nodesList.filter(n => n.online).length}/${nodesList.length} nodes online` : ''}</span>
      </div>
      <div class="row" style="gap: 10px; flex-wrap: wrap;">
        <div style="position: relative;">
          <input type="text" id="srv-filter" placeholder="Filter by name or address" oninput="filterServerCards(this.value)" style="width: 220px; padding-left: 32px; font-size: 13px; height: 36px; border-radius: 8px;">
          <span style="position: absolute; left: 10px; top: 9px; color: var(--tx-muted);">${ic('search', 14)}</span>
        </div>
        ${renderViewModeToggle('dash', 'tiles')}
        ${S.user?.role === 'admin' ? `<a href="/admin/nodes" class="btn" onclick="return handleNavClick(event, 'admin', { adminTab: 1 })" style="height: 36px; padding: 0 14px; font-size: 13px; text-decoration: none;">${ic('network', 14)} Nodes (${Math.max(1, nodesList.length)})</a>` : ''}
        <button class="btn pr" onclick="openCreateServerModal()" style="height: 36px; padding: 0 16px; font-size: 13px;">
          ${ic('plus', 14)} Create server
        </button>
      </div>
    </div>

    ${showNodeBar ? `
      <div class="node-filter-bar">
        <button type="button" class="node-filter-pill ${activeNodeFilter === 'all' ? 'on' : ''}" onclick="setNodeFilter('all')">
          ${ic('layers', 13)}
          <span>All Nodes</span>
          <span class="mu mono" style="font-size: 11px;">${totalCount}</span>
        </button>
        ${nodesList.map(n => {
          const cnt = S.servers.filter(x => (x.node_id || 'local') === n.id).length;
          return `<button type="button" class="node-filter-pill ${activeNodeFilter === n.id ? 'on' : ''}" onclick="setNodeFilter('${esc(n.id)}')">
            <span class="node-status-dot ${n.online ? 'online' : 'offline'}"></span>
            <span>${esc(n.name)}</span>
            <span class="mu mono" style="font-size: 11px;">${cnt}</span>
          </button>`;
        }).join('')}
      </div>
    ` : ''}

    ${errEntries.length > 0 ? `
      <div class="node-error-note" style="margin-bottom: 16px; display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 10px;">
        <span><b>Cluster Sync Warning:</b> ${errEntries.map(([nid, msg]) => `${esc(getNodeById(nid)?.name || nid)} (${esc(msg)})`).join(' &bull; ')}</span>
        ${S.user?.role === 'admin' ? `<a href="/admin/nodes" class="btn sm" onclick="return handleNavClick(event, 'admin', { adminTab: 1 })" style="text-decoration: none;">Inspect Nodes</a>` : ''}
      </div>
    ` : ''}

    <div class="dash-grid">
      <!-- Left Column: Server Cards Grid -->
      <div id="srv-card-grid" class="${gridCls}" style="${gridStyle}">
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
          const nodeName = s.node_name || getNodeById(nid)?.name || (nid === 'local' ? 'Local Master Node' : nid);
          const st = origIdx === S.cs ? S.status : normalizeServerStatus(s.stats?.status || s.status);
          const isRun = st === 'online' || st === 'starting';
          const maxRam = getServerMaxRamMB(s);
          const cVal = isRun ? (origIdx === S.cs ? (S.cpu.at(-1) || 0) : (Number(s.stats?.cpu_usage) || 0)) : 0;
          const mVal = isRun ? (origIdx === S.cs ? (S.memUsedMB || Number(s.stats?.memory_mb) || 0) : (Number(s.stats?.memory_mb) || 0)) : 0;
          const dVal = origIdx === S.cs ? (S.diskMB || Number(s.stats?.disk_mb) || 0) : (Number(s.stats?.disk_mb) || 0);
          const upSec = origIdx === S.cs ? S.uptimeSec : (Number(s.stats?.uptime_sec) || 0);
          const host = `${srvHost(s)}:${s.port || 25565}`;
          const coreLabel = (s.type || 'paper').charAt(0).toUpperCase() + (s.type || 'paper').slice(1);
          const srvUrl = serverRouteUrl(nid, s.id, 'con');

          if (dashVm === 'list') {
            return `
            <a id="dash-srv-${nid}-${s.id}" href="${esc(srvUrl)}" class="sift-server-card vm-dash-row" onclick="return handleServerCardClick(event, ${origIdx}, 'con')" data-name="${esc(s.name)}" data-addr="${esc(host)}">
              <div class="vm-dash-row-main">
                <div class="srv-row-thumb" style="background-image: url('${esc(coverUrl)}'), url('/img/banner.jpg');"></div>
                <div style="min-width: 0; flex: 1;">
                  <div class="row" style="gap: 8px; flex-wrap: wrap;">
                    <span class="srv-card-title">${esc(s.name || 'Minecraft Server')}</span>
                    <span data-role="badge" class="badge ${isRun ? 'online' : 'offline'}">
                      <i class="dot"></i>${isRun ? 'Running' : 'Offline'}
                      ${isRun ? `<span class="mu" style="font-weight: 400; margin-left: 3px;">&bull; Up ${formatUptime(upSec)}</span>` : ''}
                    </span>
                  </div>
                  <div class="row" style="gap: 8px; margin-top: 4px; flex-wrap: wrap;">
                    <span class="srv-card-addr">${esc(host)}</span>
                    <span class="chip" style="font-size: 11px;">${esc(coreLabel)} ${esc(s.version || '')}</span>
                    ${(showNodeBar || nid !== 'local') ? `<span class="srv-node-pill" title="${esc(nodeName)}">${esc(nodeName)}</span>` : ''}
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
                <span class="srv-card-action-btn danger" data-srv-id="${esc(s.id)}" data-srv-name="${esc(s.name)}" data-node-id="${esc(nid)}" onclick="event.preventDefault(); event.stopPropagation(); openDeleteServerModal(this.dataset.srvId, this.dataset.srvName, this.dataset.nodeId);" title="Delete Server">${ic('trash', 13)}</span>
              </div>
            </a>`;
          }

          return `
          <a id="dash-srv-${nid}-${s.id}" href="${esc(srvUrl)}" class="sift-server-card ${dashVm === 'icons' ? 'vm-dash-icon' : ''}" onclick="return handleServerCardClick(event, ${origIdx}, 'con')" data-name="${esc(s.name)}" data-addr="${esc(host)}">
            <div class="srv-card-banner" style="background-image: url('${esc(coverUrl)}'), url('/img/banner.jpg');">
              <div class="srv-card-badges-left">
                <span class="srv-card-badge">${dashVm === 'icons' ? esc(coreLabel) : `Minecraft - ${esc(coreLabel)}`}</span>
                ${(showNodeBar || nid !== 'local') ? `<span class="srv-node-pill" title="${esc(nodeName)}">${esc(nodeName)}</span>` : ''}
              </div>
              <div class="srv-card-actions">
                <span class="srv-card-action-btn" onclick="event.preventDefault(); event.stopPropagation(); loadDashboard(); toast('Refreshed server stats');" title="Refresh status">&#x21bb;</span>
                <span class="srv-card-action-btn danger" data-srv-id="${esc(s.id)}" data-srv-name="${esc(s.name)}" data-node-id="${esc(nid)}" onclick="event.preventDefault(); event.stopPropagation(); openDeleteServerModal(this.dataset.srvId, this.dataset.srvName, this.dataset.nodeId);" title="Delete Server">${ic('trash', 13)}</span>
              </div>
            </div>
            <div class="srv-card-body">
              <div class="srv-card-title-row">
                <span class="srv-card-title">${esc(s.name || 'Minecraft Server')}</span>
                <span data-role="badge" class="badge ${isRun ? 'online' : 'offline'}">
                  <i class="dot"></i>${isRun ? 'Running' : 'Offline'}
                  ${isRun && dashVm !== 'icons' ? `<span class="mu" style="font-weight: 400; margin-left: 3px;">&bull; Up ${formatUptime(upSec)}</span>` : ''}
                </span>
              </div>
              <div class="srv-card-addr">${esc(host)}</div>
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
                    <span>Memory</span>
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
        <div class="create-server-dashed-card ${dashVm === 'list' ? 'compact' : ''}" onclick="openCreateServerModal()">
          <div class="create-icon-bubble">
            ${ic('plus', 18)}
          </div>
          <span>Create a server</span>
        </div>
      </div>

      <!-- Right Column: Recent Activity & Allocation -->
      <div style="display: flex; flex-direction: column; gap: 16px;">
        <!-- Recent Activity Card -->
        <div class="card" style="padding: 18px;">
          <div class="row sb" style="margin-bottom: 14px;">
            <h2 style="font-size: 15px; font-weight: 700; color: var(--tx); margin: 0;">Recent activity</h2>
            <a href="/activity" onclick="return handleNavClick(event, 'act', { global: true })" class="mu" style="font-size: 12px; text-decoration: none;">View all</a>
          </div>
          <div id="dash-recent-activity">
            ${renderDashRecentActivityHtml()}
          </div>
        </div>

        <!-- Your Allocation Card -->
        <div class="card" style="padding: 18px;">
          <h2 style="font-size: 15px; font-weight: 700; color: var(--tx); margin-bottom: 14px;">Your allocation</h2>
          <div style="display: flex; flex-direction: column; gap: 14px;">
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
    </div>`;
  },

  con() {
    return `${srvh()}
    <div class="cgrid">
      <div class="cwrap">
        <div class="con-top-bar">
          <div class="row" style="gap:8px;">
            ${ic('terminal', 14)}
            <span>Console Stream</span>
            <span id="ws-live-pill" class="con-live-pill"><i></i>Connecting...</span>
            <span class="mu">&bull;</span>
            <span id="con-buf-count">${S.log.length} lines</span>
          </div>
          <button class="btn sm" onclick="S.log=[]; paintLog(); toast('Console buffer cleared');">Clear</button>
        </div>
        <div class="con" id="log" role="log" aria-live="polite"></div>
        <div class="cin">
          <span>&raquo;</span>
          <input id="cmd" autocomplete="off" spellcheck="false" placeholder="Type a command..." aria-label="Console command">
        </div>
      </div>
      <div class="card stc" id="stc">${statsHtml()}</div>
    </div>
    <div class="grid g3" id="chr">${chartsHtml()}</div>`;
  },

  pl() {
    const t = S.ptab;
    const vm = getViewMode('pl', 'list');
    let body = '';

    if (t === 'online') {
      if (S.loading?.players) {
        body = `<div class="card">${loaderHtml('Loading player lists...')}</div>`;
      } else if (!S.players.length) {
        body = `<div class="card mu" style="text-align:center;padding:24px;">No players currently online. Start the server to accept connections.</div>`;
      } else if (vm === 'icons' || vm === 'tiles') {
        body = `<div class="${vm === 'icons' ? 'vm-icons-grid' : 'vm-tiles-grid'}">
          ${S.players.map(p => `
            <div class="${vm === 'icons' ? 'vm-icon-card' : 'vm-tile-card'}">
              <div class="${vm === 'icons' ? 'vm-plug-icon-center' : 'vm-item-head'}">
                <div class="row" style="gap: 8px; min-width: 0; ${vm === 'icons' ? 'flex-direction: column;' : ''}">
                  ${head(p.n)}
                  <b class="vm-item-title">${esc(p.n)}</b>
                </div>
                ${p.op ? '<span class="badge online">OP</span>' : `<span class="chip">${p.ping || 32} ms</span>`}
              </div>
              <div class="mu mono" style="font-size: 11.5px;">${esc(p.t || 'Just now')} &bull; <span class="pv">${esc(p.ip || '127.0.0.1')}</span></div>
              <div class="vm-item-foot">
                <button class="btn sm" onclick="playerAction('${p.op ? 'deop' : 'op'}', '${esc(p.n)}')">${p.op ? 'Deop' : 'Op'}</button>
                <div class="row" style="gap: 6px;">
                  <button class="btn sm" onclick="playerAction('kick', '${esc(p.n)}')">Kick</button>
                  <button class="btn sm dg" onclick="playerAction('ban', '${esc(p.n)}')">Ban</button>
                </div>
              </div>
            </div>
          `).join('')}
        </div>`;
      } else {
        body = `<div class="card tw"><table>
          <thead><tr><th>Player</th><th>Ping</th><th>Playtime</th><th>IP</th><th style="text-align:right">Actions</th></tr></thead>
          <tbody>
            ${S.players.map(p => `
              <tr>
                <td>${head(p.n)}<b>${esc(p.n)}</b> ${p.op ? '<span class="badge" style="margin-left:6px;">Operator</span>' : ''}</td>
                <td class="mono">${p.ping || 32} ms</td>
                <td>${p.t || 'Just now'}</td>
                <td class="mono pv">${p.ip || '127.0.0.1'}</td>
                <td style="text-align:right">
                  <div class="row" style="justify-content:flex-end">
                    <button class="btn sm" onclick="playerAction('${p.op ? 'deop' : 'op'}', '${esc(p.n)}')">${p.op ? 'Deop' : 'Op'}</button>
                    <button class="btn sm" onclick="playerAction('kick', '${esc(p.n)}')">Kick</button>
                    <button class="btn sm dg" onclick="playerAction('ban', '${esc(p.n)}')">Ban</button>
                  </div>
                </td>
              </tr>
            `).join('')}
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
      <div class="tabs" style="margin:0;">
        ${[['online', 'Online Players'], ['wl', 'Whitelist'], ['ban', 'Banned Players']].map(([k, l]) => `
          <button class="btn ${t === k ? 'on' : ''}" onclick="S.ptab='${k}'; R(); loadPlayersData();">${l}</button>
        `).join('')}
      </div>
      ${renderViewModeToggle('pl', 'list')}
    </div>
    ${body}`;
  },

  fi() {
    const f = S.editingFile || S.ftab;
    if (S.editingFile) {
      return `${srvh()}
      <div style="margin-top: 16px;">
        <div class="row sb" style="margin-bottom: 12px; gap: 12px;">
          <div class="row" style="gap: 8px;">
            <button class="btn sm" onclick="closeEditor()">&larr; Back to files</button>
            <span class="mono" style="font-weight: 600; font-size: 14px; color: #fff;">${esc(S.editingFile)}</span>
          </div>
          <div class="row" style="gap: 8px;">
            <button class="btn sm" onclick="downloadFileItem('${esc(S.editingFile)}')">${ic('download', 14)} Download</button>
            <button class="btn sm pr" onclick="saveActiveFile()">${ic('download', 14)} Save File</button>
          </div>
        </div>
        <textarea id="ed" spellcheck="false" aria-label="File editor">${esc(S.files[S.editingFile] || '')}</textarea>
      </div>`;
    }

    const pathParts = (S.currentDir || '').split('/').filter(Boolean);
    const isSubdir = pathParts.length > 0;
    const allSelected = S.fileList.length > 0 && S.fileList.every(i => S.selectedFiles.has(i.path || i.name));
    const vm = getViewMode('files', 'list');

    const renderFileCtxMenuHtml = (item, idx, isDir, itemPath) => `
      <div class="file-ctx-menu" id="fctx-${idx}" style="display: none;">
        ${!isDir ? `<button onclick="openFile('${esc(item.name)}'); closeAllCtx();">${ic('edit', 14)} Edit</button>` : ''}
        ${!isDir ? `<button onclick="downloadFileItem('${esc(itemPath)}'); closeAllCtx();">${ic('download', 14)} Download</button>` : ''}
        ${isArchiveFile(item.name) ? `<button onclick="unzipFileItem('${esc(itemPath)}'); closeAllCtx();">${ic('archive', 14)} Extract</button>` : ''}
        <button onclick="promptRenameFile('${esc(itemPath)}'); closeAllCtx();">${ic('settings', 14)} Rename</button>
        <button onclick="copyFilePath('${esc(itemPath)}'); closeAllCtx();">${ic('copy', 14)} Copy Path</button>
        <button class="dg" onclick="deleteFileItem('${esc(itemPath)}'); closeAllCtx();">${ic('trash', 14)} Delete</button>
      </div>`;

    let filesContainerHtml = '';
    if (vm === 'list') {
      filesContainerHtml = `
      <div class="card tw" style="padding: 0; overflow: visible;">
        <table class="file-table" style="width: 100%;">
          <thead>
            <tr style="border-bottom: 1px solid var(--ln); color: var(--tx-muted); font-size: 12px; text-transform: uppercase; letter-spacing: 0.5px;">
              <th style="width: 36px; padding: 12px 0 12px 18px; text-align: left;">
                <input type="checkbox" id="chk-all-files" onchange="toggleSelectAllFiles(this.checked)" ${allSelected ? 'checked' : ''} aria-label="Select all files">
              </th>
              <th style="width: 40px; padding: 12px 4px;"></th>
              <th style="padding: 12px 10px; text-align: left;">Name</th>
              <th style="width: 140px; padding: 12px 10px; text-align: left;">Size</th>
              <th style="width: 180px; padding: 12px 10px; text-align: left;">Modified</th>
              <th style="width: 50px; padding: 12px 18px 12px 10px; text-align: right;"></th>
            </tr>
          </thead>
          <tbody>
            ${isSubdir ? `
              <tr onclick="browseFolder('..')">
                <td style="padding-left: 18px;"></td>
                <td style="width: 40px; padding: 8px 4px;"><span class="fi-badge up" title="Parent folder">${ic('folderUp', 15)}</span></td>
                <td colspan="4" style="font-weight: 600; color: var(--tx);">.. (Parent folder)</td>
              </tr>
            ` : ''}
            ${S.fileList.map((item, idx) => {
              const isDir = item.is_dir;
              const icon = fileListIconHtml(item);
              const sizeStr = formatFileSize(item.size, isDir);
              const dateStr = formatFileDate(item.modified);
              const itemPath = item.path || item.name;
              const isSelected = S.selectedFiles.has(itemPath);

              return `
              <tr class="file-row ${isSelected ? 'row-selected' : ''}" data-name="${esc(item.name)}"
                  onclick="${isDir ? `browseFolder('${esc(item.name)}')` : `openFile('${esc(item.name)}')`}"
                  oncontextmenu="handleFileContextMenu(event, ${idx}); return false;">
                <td style="width: 36px; padding-left: 18px;" onclick="event.stopPropagation();">
                  <input type="checkbox" class="file-chk" ${isSelected ? 'checked' : ''}
                         onchange="toggleFileSelect('${esc(itemPath)}', this.checked)"
                         aria-label="Select ${esc(item.name)}">
                </td>
                <td style="width: 40px; padding: 8px 4px;">${icon}</td>
                <td style="font-weight: 500; color: var(--tx);">${esc(item.name)}</td>
                <td class="mu mono" style="font-size: 12.5px; width: 140px;">${sizeStr}</td>
                <td class="mu mono" style="font-size: 12.5px; width: 180px;">${dateStr}</td>
                <td style="width: 50px; text-align: right; padding-right: 18px; position: relative;" onclick="event.stopPropagation();">
                  <button class="btn sm" onclick="toggleFileCtx(event, ${idx})" style="padding: 4px 8px; border: 0; background: transparent; font-size: 18px; line-height: 1; color: var(--tx-muted);">&#8942;</button>
                  ${renderFileCtxMenuHtml(item, idx, isDir, itemPath)}
                </td>
              </tr>`;
            }).join('')}
            ${S.loading?.files ? `<tr><td colspan="6">${loaderHtml('Loading directory contents...')}</td></tr>` : !S.fileList.length && !isSubdir ? '<tr><td colspan="6" class="mu" style="text-align:center; padding: 32px;">This folder is empty. Drag and drop files here or click Upload.</td></tr>' : ''}
          </tbody>
        </table>
      </div>`;
    } else if (vm === 'icons') {
      filesContainerHtml = `
      <div class="card" style="padding: 16px;">
        ${S.fileList.length > 0 ? `
          <div class="row sb" style="margin-bottom: 12px; padding-bottom: 10px; border-bottom: 1px solid var(--border);">
            <label class="row" style="gap: 8px; cursor: pointer; font-size: 12.5px; color: var(--tx-muted);">
              <input type="checkbox" id="chk-all-files" onchange="toggleSelectAllFiles(this.checked)" ${allSelected ? 'checked' : ''} aria-label="Select all files">
              <span>Select all (${S.fileList.length})</span>
            </label>
            <span class="mu" style="font-size: 12px;">Right-click or use &vellip; for options</span>
          </div>` : ''}
        ${S.loading?.files ? loaderHtml('Loading directory contents...') : (!S.fileList.length && !isSubdir)
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
            ${S.fileList.map((item, idx) => {
              const isDir = item.is_dir;
              const sizeStr = isDir ? 'Folder' : formatFileSize(item.size, false);
              const itemPath = item.path || item.name;
              const isSelected = S.selectedFiles.has(itemPath);
              return `
              <div class="vm-icon-card vm-file-card file-row ${isSelected ? 'selected' : ''}" data-name="${esc(item.name)}"
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
        ${S.fileList.length > 0 ? `
          <div class="row sb" style="margin-bottom: 12px; padding-bottom: 10px; border-bottom: 1px solid var(--border);">
            <label class="row" style="gap: 8px; cursor: pointer; font-size: 12.5px; color: var(--tx-muted);">
              <input type="checkbox" id="chk-all-files" onchange="toggleSelectAllFiles(this.checked)" ${allSelected ? 'checked' : ''} aria-label="Select all files">
              <span>Select all (${S.fileList.length})</span>
            </label>
            <span class="mu" style="font-size: 12px;">Right-click or use &vellip; for options</span>
          </div>` : ''}
        ${S.loading?.files ? loaderHtml('Loading directory contents...') : (!S.fileList.length && !isSubdir)
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
            ${S.fileList.map((item, idx) => {
              const isDir = item.is_dir;
              const sizeStr = isDir ? 'Folder' : formatFileSize(item.size, false);
              const dateStr = formatFileDate(item.modified);
              const itemPath = item.path || item.name;
              const isSelected = S.selectedFiles.has(itemPath);
              return `
              <div class="vm-tile-card vm-file-card file-row ${isSelected ? 'selected' : ''}" data-name="${esc(item.name)}"
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
    <div style="margin-top: 16px;">
      <div class="row sb" style="margin-bottom: 16px; flex-wrap: wrap; gap: 12px;">
        <!-- Breadcrumb Navigation -->
        <div class="breadcrumb" style="font-size: 13.5px; font-weight: 500; display: flex; align-items: center; flex-wrap: wrap; gap: 4px;">
          <span onclick="browseFolder('')" style="color: var(--tx-muted); cursor: pointer; display: inline-flex; align-items: center; gap: 6px;">${ic('folder', 15)} / home / container /</span>
          ${pathParts.map((part, idx) => {
            const sub = pathParts.slice(0, idx + 1).join('/');
            const isLast = idx === pathParts.length - 1;
            return `<span onclick="browseFolder('${sub}')" style="cursor: pointer; color: ${isLast ? 'var(--tx)' : 'var(--tx-muted)'};">${esc(part)} /</span>`;
          }).join('')}
        </div>

        <!-- Search, View Mode & Action Buttons -->
        <div class="row" style="gap: 10px; flex-wrap: wrap;">
          <div style="position: relative;">
            <input type="text" id="fi-filter" placeholder="Filter this folder" oninput="filterFileList(this.value)" style="width: 200px; height: 36px; padding: 0 28px 0 12px; font-size: 13px; border-radius: 8px;">
            <kbd style="position: absolute; right: 10px; top: 9px; font-size: 11px; color: var(--tx-muted); background: rgba(255,255,255,0.06); padding: 1px 5px; border-radius: 4px;">/</kbd>
          </div>
          ${renderViewModeToggle('files', 'list')}
          <button class="btn" onclick="promptNewFolder()" style="height: 36px;">${ic('folder', 14)} New folder</button>
          <button class="btn" onclick="uploadFileDialog()" style="height: 36px;">${ic('download', 14)} Upload</button>
          <button class="btn pr" onclick="promptNewFile()" style="height: 36px;">${ic('plus', 14)} New file</button>
        </div>
      </div>

      ${filesContainerHtml}

      <!-- Floating Bulk Action Bar -->
      ${S.selectedFiles.size > 0 ? `
        <div class="bulk-action-bar">
          <span style="font-weight: 600; font-size: 13.5px; color: #fff;">${S.selectedFiles.size} item${S.selectedFiles.size > 1 ? 's' : ''} selected</span>
          <div class="row" style="gap: 8px;">
            ${Array.from(S.selectedFiles).some(p => isArchiveFile(p)) ? `<button class="btn sm" onclick="unzipSelectedArchive()">${ic('archive', 14)} Extract</button>` : ''}
            <button class="btn sm dg" onclick="deleteSelectedFiles()">${ic('trash', 14)} Delete Selected</button>
            <button class="btn sm" onclick="deselectAllFiles()">Deselect</button>
          </div>
        </div>
      ` : ''}
    </div>`;
  },

  plug() {
    const isBrowse = (S.plugTab || 'browse') === 'browse';
    const isFabric = String(S.core || '').toLowerCase() === 'fabric';
    const targetFolder = isFabric ? 'mods/' : 'plugins/';
    const instFilter = (S.plugFilter || '').toLowerCase();
    const filteredInstalled = S.plugins
      .map((p, i) => ({ p, i }))
      .filter(({ p }) => !instFilter || String(p[0] || '').toLowerCase().includes(instFilter));
    const storeVm = getViewMode('plug_store', 'list');
    const instVm = getViewMode('plug_inst', 'list');
    const storeVersions = getAllMcVersions(S.plugLoader || '');
    const verChoices = (S.plugVer && !storeVersions.includes(S.plugVer)) ? [S.plugVer, ...storeVersions] : storeVersions;

    const storeGridClass = storeVm === 'icons' ? 'vm-icons-grid wide' : (storeVm === 'tiles' ? 'vm-tiles-grid' : '');
    const storeGridStyle = storeVm === 'list' ? 'display: flex; flex-direction: column; gap: 10px;' : '';
    const storeBodyHtml = Array.isArray(S.lastModrinthHits)
      ? renderModrinthHitsHtml(S.lastModrinthHits)
      : loaderHtml('Loading top plugins from Modrinth...');

    let installedBodyHtml = '';
    if (S.loading?.plugins) {
      installedBodyHtml = `<div class="card">${loaderHtml('Loading installed plugins...')}</div>`;
    } else if (!filteredInstalled.length) {
      installedBodyHtml = `<div class="card" style="padding: 32px; text-align: center;">
        <div class="mu" style="margin-bottom: 12px;">${S.plugins.length ? 'No installed plugins match your filter.' : 'No plugins installed yet in ' + targetFolder + '.'}</div>
        ${!S.plugins.length ? `<button class="btn pr sm" onclick="setPlugTab('browse')">${ic('store', 14)} Browse Modrinth Plugins</button>` : ''}
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
              <label class="sw" title="Toggle plugin state">
                <input type="checkbox" ${p[2] ? 'checked' : ''} onchange="togglePlugin(${i}, this.checked)">
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
              <label class="sw" title="Toggle plugin state">
                <input type="checkbox" ${p[2] ? 'checked' : ''} onchange="togglePlugin(${i}, this.checked)">
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
              <div style="width: 44px; height: 44px; border-radius: 8px; background: rgba(56, 122, 255, 0.12); display: grid; place-items: center; color: var(--ac); font-size: 20px; flex-shrink: 0;">
                🧩
              </div>
              <div style="min-width: 0;">
                <b class="plugin-title">${esc(p[0])}</b>
                <div class="mu mono" style="font-size: 12px; margin-top: 3px;">
                  ${esc(p[1] || 'installed')}${p[3] ? ` &bull; ${esc(formatFileDate(p[3]))}` : ''}
                </div>
              </div>
            </div>
            <div class="row" style="gap: 12px; flex-shrink: 0;">
              <label class="sw" title="Toggle plugin state">
                <input type="checkbox" ${p[2] ? 'checked' : ''} onchange="togglePlugin(${i}, this.checked)">
                <i></i>
              </label>
              <button class="btn sm dg" data-plugin="${esc(p[0])}" onclick="deletePlugin(this.dataset.plugin)">${ic('trash', 13)} Uninstall</button>
            </div>
          </div>
        `).join('')}
      </div>`;
    }

    return `${srvh()}
    <div style="margin-top: 18px;">
      <!-- Top Title and Tabs Bar -->
      <div class="row sb" style="margin-bottom: 16px; flex-wrap: wrap; gap: 14px;">
        <div>
          <h1 style="font-size: 24px; font-weight: 800; color: var(--tx); margin: 0;">Plugins &amp; Store</h1>
          <div class="mu" style="font-size: 12.5px; margin-top: 3px;">
            Browse and install Modrinth plugins directly into <span class="mono">${targetFolder}</span>
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
            <input type="text" id="store-search" placeholder="Search plugins on Modrinth (e.g. EssentialsX, LuckPerms, WorldEdit)..." value="${esc(S.q || '')}" oninput="S.q=this.value; searchModrinthDebounced();" style="height: 40px; font-size: 14px; padding-left: 38px; border-radius: 10px;">
            <span style="position: absolute; left: 12px; top: 11px; color: var(--tx-muted);">${ic('search', 16)}</span>
          </div>

          <select style="width: auto; min-width: 145px; height: 40px; font-size: 13px;" onchange="S.plugLoader=this.value; refreshOpenVersionSelects(); searchModrinth();">
            <option value="" ${!S.plugLoader ? 'selected' : ''}>Software: Any</option>
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

        <div id="store-grid" class="${storeGridClass}" style="${storeGridStyle}">
          ${storeBodyHtml}
        </div>
      ` : `
        <!-- Installed Plugins Filter & List -->
        <div class="row sb" style="gap: 10px; flex-wrap: wrap; margin-bottom: 14px;">
          <div style="position: relative; flex: 1; min-width: 240px; max-width: 420px;">
            <input type="text" placeholder="Filter installed plugins..." value="${esc(S.plugFilter || '')}" oninput="S.plugFilter=this.value; R();" style="height: 38px; font-size: 13.5px; padding-left: 36px; border-radius: 10px;">
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
        <div>
          <span class="mu">Currently Running</span>
          <b class="big">${esc(S.core)} ${esc(S.ver)}</b>
          <span class="mu">Java 25, 21, 17 &amp; 8 supported &bull; Direct installer for all server engines &amp; Minecraft versions</span>
        </div>
        <button class="btn pr" data-srv-id="${esc(srv?.id || '')}" data-node-id="${esc(srv?.node_id || 'local')}" onclick="openReinstallModal(this.dataset.srvId, this.dataset.nodeId)">
          ${ic('refresh', 14)} Reinstall Server (mcjarfiles.com)
        </button>
      </div>
    </div>

    <h2 style="margin-top:20px">1. Choose Server Engine</h2>
    <div class="grid g4">
      ${CORES.map(([name, desc]) => `
        <button class="card sel ${sel === name ? 'on' : ''}" onclick="S.sel='${name}'; R();">
          <div class="row sb" style="width:100%">
            <b>${name}</b>
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
            <input type="text" placeholder="Search versions (e.g. 1.21.4, 1.20.1, 1.8.9)..." value="${esc(S.vq || '')}" oninput="S.vq=this.value; renderVersionRows();" style="padding-left: 34px;">
            <span style="position: absolute; left: 10px; top: 10px; color: var(--tx-muted); pointer-events: none;">${ic('search', 15)}</span>
          </div>
          ${!isProxy ? `
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
            <thead><tr><th>Version</th><th>Release / Notes</th><th>Java</th><th>Engine</th><th style="text-align:right">Action</th></tr></thead>
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
                  <button class="btn sm dg" onclick="deleteBackup('${esc(b.name || b[0])}')">${ic('trash', 14)} Delete</button>
                </div>
              </td>
            </tr>
          `).join('')}
        </tbody>
      </table></div>`;
    }
    return `${srvh()}
    <div class="row sb" style="margin-top:14px; margin-bottom:12px; flex-wrap:wrap; gap:10px;">
      <span class="mu">${S.bk.length} backups recorded</span>
      <div class="row" style="gap:8px;">
        ${renderViewModeToggle('bk', 'list')}
        <button class="btn pr" onclick="createBackup()">${ic('plus', 14)} Create Full Backup</button>
      </div>
    </div>
    ${body}`;
  },

  sc() {
    return `${srvh()}
    <div class="card tw" style="margin-top:14px">
      <table>
        <thead><tr><th>Task</th><th>Cron Expression</th><th>Status</th></tr></thead>
        <tbody>
          ${S.sc.map((s, i) => `
            <tr>
              <td><b>${esc(s[0])}</b></td>
              <td class="mono">${esc(s[1])}</td>
              <td>
                <label class="sw">
                  <input type="checkbox" ${s[2] ? 'checked' : ''} onchange="S.sc[${i}][2]=this.checked; toast('${s[0]} schedule updated');">
                  <i></i>
                </label>
              </td>
            </tr>
          `).join('')}
        </tbody>
      </table>
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
        <div class="card" style="padding: 22px;">
          <h2 style="font-size: 16px; margin-bottom: 8px; color: #fff;">SFTP Management Endpoint</h2>
          <p class="mu" style="font-size: 13px; margin-bottom: 14px;">Pure-Go jailed SFTP daemon running natively in NeoPanel for secure file management.</p>
          <div class="mono" style="background: rgba(0, 0, 0, 0.4); border: 1px solid var(--border); padding: 12px 14px; border-radius: 8px; font-size: 12.5px; line-height: 1.8; margin-bottom: 14px;">
            <div><span class="mu">Host:</span> <b style="color: #fff;">${host}</b></div>
            <div><span class="mu">Port:</span> <b style="color: #fff;">${sftpPort}</b></div>
            <div><span class="mu">Username:</span> <b style="color: #fff;">${sftpUser}</b></div>
          </div>
          ${srv.node_id && srv.node_id !== 'local' ? `<div class="mu" style="font-size: 11.5px; margin-bottom: 12px; line-height: 1.45;">Remote cluster node SFTP password: use that node's <code>node_token</code> (<code>sudo neopanel node-token</code>) or a local user password.</div>` : ''}
          <button class="btn sm" onclick="copyDbText('sftp://${sftpUser}@${host}:${sftpPort}', 'SFTP URI')">
            Copy SFTP URI
          </button>
        </div>

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
    return `${srvh()}
    <div class="card" style="margin-top:14px">
      <h2>Custom Domain &amp; Subdomain</h2>
      <label class="f">
        <span>Choose a custom subdomain or root domain</span>
        <div class="row">
          <input type="text" id="sub-domain" value="proxy" style="max-width:240px">
          <span class="mu">.mnvibes.org</span>
        </div>
      </label>
      <button class="btn pr" onclick="toast('Domain mapping saved successfully')">Save Subdomain</button>
    </div>`;
  },

  stp() {
    const srv = curServer();
    const rawJava = srv?.java_path || '';
    const curJava = resolveJavaSelectValue(rawJava, srv?.version, srv?.type);
    const isCustom = curJava === 'custom';
    const memMin = srv?.memory_min || Math.min(1024, srv?.memory_max || 2048);
    const memMax = srv?.memory_max || 2048;
    const jvmFlags = srv?.java_args || '-XX:+UseG1GC';
    const jarName = srv?.jar_file || 'server.jar';
    const javaCmd = isCustom ? (rawJava || 'java') : `java (OpenJDK ${curJava})`;

    return `${srvh()}
    <div style="margin-top: 18px;">
      <!-- Top Grid: Startup Command & Docker Image / Java -->
      <div class="stp-top-grid">
        <div class="card" style="padding: 20px;">
          <h2 style="font-size: 15px; font-weight: 700; color: var(--tx); margin-bottom: 12px;">Active Startup Command</h2>
          <div class="mono" style="background: var(--con-bg); padding: 14px 16px; border-radius: 8px; border: 1px solid var(--border); color: #f0f3f6; font-size: 12.5px; line-height: 1.6; word-break: break-all;">
            ${esc(javaCmd)} -Xms${memMin}M -Xmx${memMax}M ${esc(jvmFlags)} -jar ${esc(jarName)} nogui
          </div>
          <span class="mu" style="font-size: 12px; margin-top: 8px; display: block;">This command is dynamically constructed from your selected Java binary, memory allocation, and JVM arguments. Auto-restart automatically times out after 2 failed startup attempts.</span>
        </div>

        <div class="card" style="padding: 20px;">
          <h2 style="font-size: 15px; font-weight: 700; color: var(--tx); margin-bottom: 12px;">Java Runtime Environment</h2>
          <select id="stp-java" onchange="toggleCustomJavaInput(this.value)" style="margin-bottom: 8px;">
            ${renderJavaSelectOptions(rawJava, true, srv?.version, srv?.type, srv?.node_id || 'local')}
          </select>
          <div id="stp-custom-wrap" style="${isCustom ? '' : 'display: none;'} margin-bottom: 8px;">
            <input type="text" id="stp-java-custom" value="${esc(isCustom ? rawJava : '')}" placeholder="/usr/lib/jvm/java-25-temurin/bin/java" style="font-size: 13px;">
          </div>
          ${renderJavaAvailabilityStrip(srv?.node_id || 'local')}
          <span class="mu" style="font-size: 12px; line-height: 1.5; display: block; margin-top: 8px;">Select which premade Java runtime executes your server JAR (Java 25, 21, 17, or 8).</span>
        </div>
      </div>

      <!-- Variables Section -->
      <h2 style="font-size: 18px; font-weight: 700; color: var(--tx); margin-bottom: 16px;">Runtime Variables &amp; Parameters</h2>

      <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 16px; margin-bottom: 20px;">
        <div class="card" style="padding: 20px;">
          <div style="font-size: 12.5px; font-weight: 700; color: var(--tx); text-transform: uppercase; margin-bottom: 8px; letter-spacing: 0.04em;">MINECRAFT VERSION</div>
          <div style="display: grid; grid-template-columns: minmax(135px, 165px) 1fr; gap: 8px; margin-bottom: 8px;">
            <select id="stp-ver-sel" data-searchable="1" onchange="const inp=$('#stp-ver'); if(inp) inp.value=this.value;">
              ${(() => {
                const curV = srv?.version || '1.21.4';
                const allV = getAllMcVersions(srv?.type || S.core || '');
                const list = (curV && !allV.includes(curV)) ? [curV, ...allV] : allV;
                return list.map(v => `<option value="${esc(v)}" ${v === curV ? 'selected' : ''}>${esc(v)}</option>`).join('');
              })()}
            </select>
            <input type="text" id="stp-ver" value="${esc(srv?.version || '1.21.4')}" style="font-weight: 600;" oninput="const sel=$('#stp-ver-sel'); if(sel && Array.from(sel.options).some(o=>o.value===this.value)){ sel.value=this.value; if(sel._cselectRebuild) sel._cselectRebuild(); }">
          </div>
          <span class="mu" style="font-size: 12px; line-height: 1.5; display: block;">Pick any Minecraft release from the searchable list or enter a custom version/snapshot tag.</span>
        </div>

        <div class="card" style="padding: 20px;">
          <div style="font-size: 12.5px; font-weight: 700; color: var(--tx); text-transform: uppercase; margin-bottom: 8px; letter-spacing: 0.04em;">SERVER JAR FILE</div>
          <input type="text" id="stp-jar" value="${esc(srv?.jar_file || 'server.jar')}" style="margin-bottom: 8px; font-weight: 600;">
          <span class="mu" style="font-size: 12px; line-height: 1.5; display: block;">The name of the server executable JAR inside the container directory.</span>
        </div>

        <div class="card" style="padding: 20px;">
          <div style="font-size: 12.5px; font-weight: 700; color: var(--tx); text-transform: uppercase; margin-bottom: 8px; letter-spacing: 0.04em;">MAX RAM ALLOCATION</div>
          <select id="stp-ram" style="margin-bottom: 8px; font-weight: 600;">
            <option value="1024" ${memMax === 1024 ? 'selected' : ''}>1 GB (1024 MB)</option>
            <option value="2048" ${memMax === 2048 ? 'selected' : ''}>2 GB (2048 MB)</option>
            <option value="4096" ${memMax === 4096 ? 'selected' : ''}>4 GB (4096 MB)</option>
            <option value="6144" ${memMax === 6144 ? 'selected' : ''}>6 GB (6144 MB)</option>
            <option value="8192" ${memMax === 8192 ? 'selected' : ''}>8 GB (8192 MB)</option>
            <option value="16384" ${memMax === 16384 ? 'selected' : ''}>16 GB (16384 MB)</option>
          </select>
          <span class="mu" style="font-size: 12px; line-height: 1.5; display: block;">Adjust the maximum heap memory (-Xmx) given to this Minecraft server.</span>
        </div>

        <div class="card" style="padding: 20px;">
          <div style="font-size: 12.5px; font-weight: 700; color: var(--tx); text-transform: uppercase; margin-bottom: 8px; letter-spacing: 0.04em;">CUSTOM JVM FLAGS (JAVA_ARGS)</div>
          <input type="text" id="stp-args" value="${esc(srv?.java_args || '-XX:+UseG1GC')}" style="margin-bottom: 8px; font-weight: 500;">
          <span class="mu" style="font-size: 12px; line-height: 1.5; display: block;">Additional JVM options passed directly to the Java process on start.</span>
        </div>
      </div>

      <button class="btn pr" onclick="saveStartupSettings()" style="padding: 10px 24px;">${ic('download', 14)} Save &amp; Apply Startup Variables</button>
    </div>`;
  },

  bill() {
    const srv = curServer();
    const header = srv ? srvh() : '';
    return `${header}
    <div style="${srv ? 'margin-top: 18px;' : ''}">
      <h1 style="font-size: 22px; font-weight: 800; color: var(--tx); margin-bottom: 16px;">Billing &amp; Resource Allocation</h1>
      <div class="grid g2" style="margin-bottom: 16px;">
        <div class="card" style="padding: 20px;">
          <h2 style="font-size: 15px; font-weight: 700; color: #fff; margin-bottom: 12px;">Active Server Tier</h2>
          <div class="row sb" style="margin-bottom: 8px;">
            <span class="mu">Plan Name</span>
            <b style="color: #fff;">NeoPanel Dedicated vCPU Tier 1</b>
          </div>
          <div class="row sb" style="margin-bottom: 8px;">
            <span class="mu">Allocated Memory</span>
            <b class="mono">${getServerMaxRamMB(srv)} MB Dedicated RAM</b>
          </div>
          <div class="row sb" style="margin-bottom: 8px;">
            <span class="mu">Port Allocation</span>
            <b class="mono">${srv?.port || 25565} (Primary IPv4)</b>
          </div>
          <div class="row sb">
            <span class="mu">Billing Cycle</span>
            <span class="badge online"><i class="dot"></i>Active &bull; Renews in 30 days</span>
          </div>
        </div>

        <div class="card" style="padding: 20px;">
          <h2 style="font-size: 15px; font-weight: 700; color: #fff; margin-bottom: 12px;">Payment Method</h2>
          <p class="mu" style="font-size: 13px; margin-bottom: 14px;">Direct VPS node deployment. No third-party recurring subscription fee.</p>
          <button class="btn sm" onclick="toast('Invoice downloaded')">${ic('download', 14)} Download Receipt (.PDF)</button>
        </div>
      </div>
    </div>`;
  },

  set() {
    const p = S.props;
    const srv = curServer();
    const o = (v, a) => a.map(x => `<option ${x === v ? 'selected' : ''}>${x}</option>`).join('');
    const sw = (k, l) => `<label class="row sb" style="padding:7px 0; border-bottom:1px solid var(--ln);">
      <span>${l}</span>
      <span class="sw">
        <input type="checkbox" ${p[k] ? 'checked' : ''} onchange="S.props.${k}=this.checked;">
        <i></i>
      </span>
    </label>`;

    const curMem = getServerMaxRamMB(srv);

    return `${srvh()}
    <div class="grid g2" style="margin-top:14px">
      <div class="card">
        <h2>Server Configuration</h2>
        <label class="f">
          <span>Server Message (MOTD)</span>
          <input type="text" id="prop-motd" value="${esc(p.motd)}" oninput="S.props.motd=this.value;">
        </label>
        <div class="grid g2">
          <label class="f">
            <span>Server Game Port</span>
            <input type="number" min="1024" max="65535" id="prop-port" value="${srv?.port || 25565}">
          </label>
          <label class="f">
            <span>Max Players</span>
            <input type="number" min="1" max="1000" id="prop-max" value="${p.max}" oninput="S.props.max=+this.value||1;">
          </label>
        </div>
        <label class="f">
          <span>Game Difficulty</span>
          <select id="prop-diff" onchange="S.props.diff=this.value">${o(p.diff, ['peaceful', 'easy', 'normal', 'hard'])}</select>
        </label>
        <label class="f">
          <span>Default Gamemode</span>
          <select id="prop-mode" onchange="S.props.mode=this.value">${o(p.mode, ['survival', 'creative', 'adventure'])}</select>
        </label>
      </div>

      <div class="card">
        <h2>Game Rules &amp; Protection</h2>
        ${sw('pvp', 'Player versus Player (PvP)')}
        ${sw('wl', 'Enforce Whitelist')}
        ${sw('cmd', 'Enable Command Blocks')}

        <h2 style="margin-top:20px">Memory &amp; Resources</h2>
        <label class="f">
          <span>Allocated Memory</span>
          <select id="prop-ram">
            <option value="1024" ${curMem === 1024 ? 'selected' : ''}>1 GB (1024 MB)</option>
            <option value="2048" ${curMem === 2048 ? 'selected' : ''}>2 GB (2048 MB)</option>
            <option value="4096" ${curMem === 4096 ? 'selected' : ''}>4 GB (4096 MB)</option>
            <option value="8192" ${curMem === 8192 ? 'selected' : ''}>8 GB (8192 MB)</option>
            <option value="16384" ${curMem === 16384 ? 'selected' : ''}>16 GB (16384 MB)</option>
          </select>
        </label>
        <label class="f" style="margin-top: 10px;">
          <span>Java Runtime Version</span>
          <select id="prop-java" onchange="toggleCustomPropJava(this.value)">
            ${renderJavaSelectOptions(srv?.java_path || '', true, srv?.version, srv?.type, srv?.node_id || 'local')}
          </select>
        </label>
        <div id="prop-custom-wrap" style="${isStandardJavaChoice(srv?.java_path) ? 'display: none;' : ''} margin-top: 6px;">
          <input type="text" id="prop-java-custom" value="${esc(!isStandardJavaChoice(srv?.java_path) ? srv?.java_path : '')}" placeholder="/usr/lib/jvm/java-25-temurin/bin/java" style="font-size: 13px;">
        </div>
        ${renderJavaAvailabilityStrip(srv?.node_id || 'local')}
        <div class="row" style="margin-top:14px;">
          <button class="btn pr" onclick="saveProperties()">${ic('download', 14)} Save Properties</button>
          <button class="btn" onclick="toggleTheme()">Toggle Dark/Light</button>
        </div>
      </div>
    </div>

    <!-- Reinstall Server Section -->
    <div class="card" style="margin-top: 20px; border: 1px solid var(--border-light); padding: 22px;">
      <div class="row sb" style="align-items: center; flex-wrap: wrap; gap: 16px;">
        <div>
          <h2 style="font-size: 16px; margin: 0 0 6px; display: flex; align-items: center; gap: 8px; color: #fff;">
            ${ic('refresh', 16)} Reinstall Server Instance
          </h2>
          <p class="mu" style="font-size: 13px; margin: 0; max-width: 620px;">
            Re-download and deploy a fresh Minecraft server binary (Paper, Purpur, Leaf, Folia, Fabric, Vanilla, Velocity) via <b>mcjarfiles.com</b> API. You can choose to preserve world data, configs, and plugins.
          </p>
        </div>
        <button class="btn" data-srv-id="${esc(srv?.id || '')}" data-node-id="${esc(srv?.node_id || 'local')}" onclick="openReinstallModal(this.dataset.srvId, this.dataset.nodeId)" style="height: 38px; padding: 0 18px; font-weight: 600;">
          ${ic('refresh', 14)} Reinstall Server
        </button>
      </div>
    </div>

    <!-- Danger Zone: Delete Server -->
    <div class="card" style="margin-top: 20px; border: 1px solid rgba(239, 68, 68, 0.35); background: rgba(239, 68, 68, 0.05); padding: 22px;">
      <div class="row sb" style="align-items: center; flex-wrap: wrap; gap: 16px;">
        <div>
          <h2 style="color: #ef4444; font-size: 16px; margin: 0 0 6px; display: flex; align-items: center; gap: 8px;">
            ${ic('trash', 16)} Danger Zone: Delete Server Instance
          </h2>
          <p class="mu" style="font-size: 13px; margin: 0; max-width: 620px;">
            Permanently destroy this Minecraft server instance (<b>${esc(srv?.name || '')}</b>). All files, worlds, player data, plugins, configs, and allocated ports will be erased. This action cannot be reversed.
          </p>
        </div>
        <button class="btn dg" data-srv-id="${esc(srv?.id || '')}" data-srv-name="${esc(srv?.name || '')}" data-node-id="${esc(srv?.node_id || 'local')}" onclick="openDeleteServerModal(this.dataset.srvId, this.dataset.srvName, this.dataset.nodeId)" style="height: 38px; padding: 0 18px; font-weight: 600;">
          ${ic('trash', 14)} Delete Server
        </button>
      </div>
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
    const sys = S.system || { os: 'linux', arch: 'amd64', cpus: 2, total_ram_mb: 4096, free_ram_mb: 2048, version: '1.0.0', go_version: 'go1.27.0', goroutines: 24 };
    const usedRam = (sys.total_ram_mb - sys.free_ram_mb) || 0;
    const ramPct = Math.round((usedRam / (sys.total_ram_mb || 1)) * 100) || 50;
    const t = S.adminTab || 0;
    const nodesList = (Array.isArray(S.nodes) && S.nodes.length > 0) ? S.nodes : [{
      id: 'local',
      name: 'Local Master Node',
      location: 'Primary Host',
      fqdn: S.brandingDaemon?.fqdn || window.location.hostname,
      is_local: true,
      online: true,
      latency_ms: 0,
      server_count: S.servers.filter(x => (x.node_id || 'local') === 'local').length,
      system: sys,
      token_prefix: 'local'
    }];
    const onlineNodesCount = nodesList.filter(n => n.online).length;

    const draft = S.brandDraft || { ...DEFAULT_BRANDING, ...S.branding };
    const daemonDraft = S.daemonDraft || {
      fqdn: S.brandingDaemon?.fqdn || '',
      sftp_port: S.brandingDaemon?.sftp_port || S.adminDaemonCfg?.sftpPort || 2022,
      max_upload_mb: S.brandingDaemon?.max_upload_mb || S.adminDaemonCfg?.maxUploadMB || 500,
      default_ram_mb: S.brandingDaemon?.default_ram_mb || S.adminDaemonCfg?.defaultRamMB || 2048,
      auto_restart_policy: S.brandingDaemon?.auto_restart_policy !== undefined ? Boolean(S.brandingDaemon.auto_restart_policy) : true
    };

    const tabs = [
      ['Cluster & Health', 0],
      ['Cluster Nodes', 1],
      ['Managed Servers', 2],
      ['Users & Roles', 3],
      ['Panel Look & Branding', 4],
      ['Audit Logs', 5]
    ];

    let content = '';

    if (t === 0) {
      content = `
        <div class="grid g4" style="margin-bottom:18px;">
          <div class="card stat">
            <span>Host Platform</span>
            <b style="font-size:18px; padding:6px 0;">${esc(sys.os)} (${esc(sys.arch)})</b>
            <span class="mu">${sys.cpus} CPU Cores &bull; ${esc(sys.go_version || 'Go Runtime')}</span>
          </div>
          <div class="card stat">
            <span>Host Memory</span>
            <b style="font-size:18px; padding:6px 0;">${(usedRam/1024).toFixed(1)} / ${(sys.total_ram_mb/1024).toFixed(1)} GB</b>
            <div class="bar"><i style="width:${ramPct}%"></i></div>
          </div>
          <div class="card stat">
            <span>Cluster Nodes</span>
            <b style="font-size:18px; padding:6px 0;">${onlineNodesCount} / ${nodesList.length} Online</b>
            <span class="mu"><a href="/admin/nodes" onclick="return handleNavClick(event, 'admin', { adminTab: 1 })" style="color:var(--ac);text-decoration:none;">Manage linked daemons &rarr;</a></span>
          </div>
          <div class="card stat">
            <span>SFTP Daemon</span>
            <b style="font-size:18px; padding:6px 0;">Port ${daemonDraft.sftp_port || 2022}</b>
            <span class="badge online"><i class="dot"></i>Active</span>
          </div>
        </div>

        <div class="card tw" style="margin-bottom:18px;">
          <div class="row sb" style="margin-bottom:14px; flex-wrap:wrap; gap:12px;">
            <div>
              <h2 style="margin:0 0 4px; color:var(--tx);">System Optimization &amp; Garbage Collection</h2>
              <span class="mu" style="font-size:13px;">Manage host process memory, reclaim unused buffers, and inspect runtime allocations.</span>
            </div>
            <button class="btn pr" onclick="triggerAdminGC()">${ic('refresh', 14)} Run Garbage Collection (runtime.GC)</button>
          </div>
          <div style="background:rgba(255,255,255,0.02); border:1px solid var(--border); border-radius:8px; padding:16px;">
            <div class="row sb" style="padding:6px 0; border-bottom:1px solid var(--border);">
              <span class="mu">Daemon Binary Version</span>
              <span class="mono" style="color:var(--tx);">NeoPanel v${esc(sys.version || '1.0.0')} (Target: Ubuntu 24.04 LTS)</span>
            </div>
            <div class="row sb" style="padding:6px 0; border-bottom:1px solid var(--border);">
              <span class="mu">Go Compiler Version</span>
              <span class="mono" style="color:var(--tx);">${esc(sys.go_version || 'go1.27.0')}</span>
            </div>
            <div class="row sb" style="padding:6px 0; border-bottom:1px solid var(--border);">
              <span class="mu">Active Goroutines</span>
              <span class="mono" style="color:var(--tx);">${sys.goroutines || 24}</span>
            </div>
            <div class="row sb" style="padding:6px 0;">
              <span class="mu">Process Architecture</span>
              <span class="mono" style="color:var(--tx);">${esc(sys.arch)} on ${esc(sys.os)}</span>
            </div>
          </div>
        </div>

        <div class="card tw" style="margin-bottom:18px;">
          <div class="row sb" style="margin-bottom:10px; flex-wrap:wrap; gap:12px;">
            <div>
              <h2 style="margin:0 0 4px; color:var(--tx);">Local Java Runtimes (25, 21, 17 &amp; 8)</h2>
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
              <b>Dedicated Cluster Nodes:</b> Install remote Ubuntu nodes with <code>--node</code> (builds only the cluster daemon &amp; SFTP agent) and pair automatically using an Admin API key, or connect an existing node via <code>sudo neopanel node-token</code> / <code>sudo neopanel node connect</code>.
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
              <h3 style="margin: 0 0 3px; font-size: 15px; color: var(--tx);">1-Command Cluster Node Installer &amp; Auto-Pairing</h3>
              <span class="mu" style="font-size: 12.5px;">Installs the lightweight node daemon &amp; SFTP agent and registers directly with this Master Panel.</span>
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
              <span>Daemon Port</span>
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
                      <b>${esc(n.name)}</b>
                      ${n.is_local ? '<span class="chip" style="font-size:10px;">Local</span>' : ''}
                    </div>
                  </td>
                  <td class="mu">${esc(n.location || 'Unspecified')}</td>
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
                    <b style="font-size: 16px; color: var(--tx); overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">${esc(n.name)}</b>
                    ${n.is_local ? '<span class="chip" style="font-size:10.5px;">Local Master</span>' : ''}
                  </div>
                  <div class="mu mono" style="font-size: 11.5px; margin-top: 3px;">
                    ${esc(n.location || 'Unspecified Region')} &bull; ${esc(n.fqdn || window.location.hostname)}
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
                    ? `<button type="button" class="btn sm" onclick="selectAdminTab(4)">Configure Host</button>`
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
        <div class="card tw" style="margin-bottom:18px;">
          <div class="row sb" style="margin-bottom:14px; flex-wrap:wrap; gap:12px;">
            <div>
              <h2 style="margin:0 0 4px; color:var(--tx);">All Cluster Servers (${S.servers.length})</h2>
              <span class="mu" style="font-size:13px;">Overview and controls for all server instances across local and linked remote nodes.</span>
            </div>
            <div class="row" style="gap:8px; flex-wrap:wrap;">
              ${renderViewModeToggle('admin_servers', 'list')}
              <button class="btn sm" onclick="bulkRestartServers()">${ic('refresh', 14)} Restart All</button>
              <button class="btn sm dg" onclick="bulkStopServers()">Stop All</button>
              <button class="btn sm pr" onclick="openCreateServerModal()">+ Deploy Server</button>
            </div>
          </div>
          ${admSrvVm === 'list' ? `
          <table>
            <thead><tr><th>Server</th><th>Node</th><th>Port</th><th>Engine</th><th>RAM Limit</th><th>Status</th><th style="text-align:right">Actions</th></tr></thead>
            <tbody>
              ${S.servers.length ? S.servers.map((s, idx) => {
                const nid = s.node_id || 'local';
                const nodeName = s.node_name || getNodeById(nid)?.name || (nid === 'local' ? 'Local Master' : nid);
                const normSt = normalizeServerStatus(s.stats?.status || s.status);
                const isRun = normSt === 'online' || normSt === 'starting';
                const srvUrl = serverRouteUrl(nid, s.id, 'con');
                return `
                <tr>
                  <td><b>${esc(s.name)}</b> <span class="mono mu" style="font-size:11px;">(${esc(s.id)})</span></td>
                  <td><span class="chip srv-node-badge">${esc(nodeName)}</span></td>
                  <td class="mono">${s.port}</td>
                  <td><span class="chip">${esc(s.type)} ${esc(s.version)}</span></td>
                  <td class="mono">${(getServerMaxRamMB(s)/1024).toFixed(1)} GB</td>
                  <td><span class="badge ${isRun ? 'online' : 'offline'}"><i class="dot"></i>${isRun ? 'Online' : 'Offline'}</span></td>
                  <td style="text-align:right">
                    <div class="row" style="justify-content:flex-end; gap:6px;">
                      <a href="${esc(srvUrl)}" class="btn sm" style="text-decoration:none;" onclick="return handleServerCardClick(event, ${idx}, 'con')">Manage</a>
                      <button class="btn sm" data-srv-id="${esc(s.id)}" data-node-id="${esc(nid)}" onclick="openReinstallModal(this.dataset.srvId, this.dataset.nodeId)">${ic('refresh', 12)} Reinstall</button>
                      <button class="btn sm dg" data-srv-id="${esc(s.id)}" data-srv-name="${esc(s.name)}" data-node-id="${esc(nid)}" onclick="openDeleteServerModal(this.dataset.srvId, this.dataset.srvName, this.dataset.nodeId)">${ic('trash', 12)} Delete</button>
                    </div>
                  </td>
                </tr>`;
              }).join('') : '<tr><td colspan="7" class="mu" style="text-align:center; padding:24px;">No server instances deployed yet.</td></tr>'}
            </tbody>
          </table>` : `
          <div class="${admSrvVm === 'icons' ? 'vm-icons-grid wide' : 'vm-tiles-grid'}">
            ${S.servers.length ? S.servers.map((s, idx) => {
              const nid = s.node_id || 'local';
              const nodeName = s.node_name || getNodeById(nid)?.name || (nid === 'local' ? 'Local Master' : nid);
              const normSt = normalizeServerStatus(s.stats?.status || s.status);
              const isRun = normSt === 'online' || normSt === 'starting';
              const srvUrl = serverRouteUrl(nid, s.id, 'con');
              return `<div class="${admSrvVm === 'icons' ? 'vm-icon-card' : 'vm-tile-card'}">
                <div class="vm-item-head">
                  <div style="min-width:0;">
                    <b class="vm-item-title" style="display:block;">${esc(s.name)}</b>
                    <div class="mu mono" style="font-size:11px; margin-top:2px;">Port ${s.port} &bull; ${(getServerMaxRamMB(s)/1024).toFixed(1)} GB</div>
                  </div>
                  <span class="badge ${isRun ? 'online' : 'offline'}"><i class="dot"></i>${isRun ? 'Online' : 'Offline'}</span>
                </div>
                <div class="row" style="gap:6px; flex-wrap:wrap;">
                  <span class="chip srv-node-badge">${esc(nodeName)}</span>
                  <span class="chip">${esc(s.type)} ${esc(s.version)}</span>
                </div>
                <div class="vm-item-foot">
                  <a href="${esc(srvUrl)}" class="btn sm pr" style="text-decoration:none;" onclick="return handleServerCardClick(event, ${idx}, 'con')">Manage</a>
                  <div class="row" style="gap:6px;">
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
        <div class="card tw" style="margin-bottom:18px;">
          <div class="row sb" style="margin-bottom:14px; flex-wrap:wrap; gap:12px;">
            <div>
              <h2 style="margin:0 0 4px; color:var(--tx);">Panel Users &amp; Permissions</h2>
              <span class="mu" style="font-size:13px;">Manage operators, administrators, and client user accounts.</span>
            </div>
            <button class="btn pr" onclick="openAddUserModal()">+ Create User</button>
          </div>
          <table>
            <thead><tr><th>ID</th><th>Username</th><th>Role</th><th>Created</th><th style="text-align:right">Actions</th></tr></thead>
            <tbody>
              ${(S.adminUsers || []).length ? S.adminUsers.map(u => `
                <tr>
                  <td class="mono mu">${u.id}</td>
                  <td><b>${esc(u.username)}</b></td>
                  <td><span class="badge ${u.role === 'admin' ? 'online' : ''}">${esc(u.role)}</span></td>
                  <td class="mu">${u.created_at ? new Date(u.created_at).toLocaleDateString() : '—'}</td>
                  <td style="text-align:right">
                    ${u.username !== S.acct.un ? `
                      <button class="btn sm dg" data-id="${u.id}" data-name="${esc(u.username)}" onclick="deleteAdminUser(this.dataset.id, this.dataset.name)">${ic('trash', 12)} Delete</button>
                    ` : '<span class="mu" style="font-size:12px;">Current Account</span>'}
                  </td>
                </tr>
              `).join('') : `
                <tr>
                  <td class="mono mu">1</td>
                  <td><b>${esc(S.acct.un)}</b></td>
                  <td><span class="badge online">admin</span></td>
                  <td class="mu">Just now</td>
                  <td style="text-align:right"><span class="mu" style="font-size:12px;">Current Account</span></td>
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
                <button type="button" id="brand-save-btn" class="btn sm pr" onclick="saveBrandingAndDaemon()">${ic('check', 13)} Save Look</button>
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
                <input type="text" value="${esc(draft.brand_subtitle || '')}" maxlength="120" oninput="onBrandInputChange('brand_subtitle', this.value)" placeholder="Ultra-Lightweight Minecraft Server Manager">
              </label>
              <label class="f" style="margin-bottom: 0;">
                <span>Footer / Login Subtext</span>
                <input type="text" value="${esc(draft.footer_text || '')}" maxlength="120" oninput="onBrandInputChange('footer_text', this.value)" placeholder="Powered by NeoPanel Daemon">
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
              <div style="font-size: 12px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: var(--tx-muted); margin-bottom: 10px;">Media &amp; Artwork URLs</div>
              <div class="grid g2">
                <label class="f" style="margin-bottom: 0;">
                  <span>Sidebar Logo URL (blank = 3D Cube)</span>
                  <input type="text" value="${esc(draft.logo_url || '')}" oninput="onBrandInputChange('logo_url', this.value)" placeholder="https://example.com/logo.png">
                </label>
                <label class="f" style="margin-bottom: 0;">
                  <span>Favicon URL (blank = default SVG)</span>
                  <input type="text" value="${esc(draft.favicon_url || '')}" oninput="onBrandInputChange('favicon_url', this.value)" placeholder="https://example.com/favicon.ico">
                </label>
                <label class="f" style="margin-bottom: 0;">
                  <span>Login Screen Wallpaper URL</span>
                  <input type="text" value="${esc(draft.login_bg_url || '')}" oninput="onBrandInputChange('login_bg_url', this.value)" placeholder="https://example.com/login-bg.jpg">
                </label>
                <label class="f" style="margin-bottom: 0;">
                  <span>Server Workspace Hero URL</span>
                  <input type="text" value="${esc(draft.server_hero_url || '')}" oninput="onBrandInputChange('server_hero_url', this.value)" placeholder="${esc(SERVER_HERO_URL)}">
                </label>
                <label class="f" style="margin-bottom: 0;">
                  <span>Dashboard Card Cover URL</span>
                  <input type="text" value="${esc(draft.card_cover_url || '')}" oninput="onBrandInputChange('card_cover_url', this.value)" placeholder="${esc(MINECRAFT_COVER_URL)}">
                </label>
                <label class="f" style="margin-bottom: 0;">
                  <span>Default User Avatar URL</span>
                  <input type="text" value="${esc(draft.default_avatar_url || '')}" oninput="onBrandInputChange('default_avatar_url', this.value)" placeholder="${esc(DEFAULT_USER_AVATAR_URL)}">
                </label>
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
              <button type="button" class="btn pr" style="flex: 1;" onclick="saveBrandingAndDaemon()">${ic('check', 14)} Save Global Look</button>
              <button type="button" class="btn" onclick="discardBrandDraft()">Discard</button>
            </div>
          </div>
        </div>

        <div class="card" style="margin-bottom:18px;">
          <h2 style="margin:0 0 6px; color:var(--tx);">Daemon Network &amp; Cluster Defaults</h2>
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
                <span>SFTP Daemon Port</span>
                <input type="number" id="adm-sftp" value="${daemonDraft.sftp_port || 2022}" min="1" max="65535" required oninput="onDaemonInputChange('sftp_port', Number(this.value))">
              </label>
            </div>
            <div class="grid g2" style="margin-top:4px;">
              <label class="f">
                <span>Max Upload Size (MB)</span>
                <input type="number" id="adm-upload" value="${daemonDraft.max_upload_mb || 500}" min="10" max="25000" required oninput="onDaemonInputChange('max_upload_mb', Number(this.value))">
              </label>
              <label class="f">
                <span>Default Instance RAM (MB)</span>
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
              <button type="submit" class="btn pr">${ic('download', 14)} Save Daemon Settings</button>
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
      `;
    } else if (t === 5) {
      content = renderActivityLogSection({
        scope: 'all',
        title: 'Global Daemon Audit Log',
        subtitle: 'Cluster-wide security, node connections, databases, user management, power, console, file & software audit events',
        showActor: true
      });
    }

    return `<header class="top">
      ${ic('shield', 26)}
      <h1>Daemon Administration Panel</h1>
      <span class="sp"></span>
      <span class="badge online"><i class="dot"></i>System Healthy</span>
    </header>

    <div class="tabs" style="margin-top: 16px; margin-bottom: 16px;">
      ${tabs.map(([lbl, idx]) => `
        <button class="${t === idx ? 'active' : ''}" onclick="selectAdminTab(${idx})">${lbl}</button>
      `).join('')}
    </div>

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
          <button class="btn sm" onclick="$('#aem').focus()">Change Gravatar</button>
        </div>
        <div class="pbody">
          <div class="grid g2">
            <label class="f"><span>First name</span><input type="text" id="afn" value="${esc(a.fn)}"></label>
            <label class="f"><span>Last name</span><input type="text" id="aln" value="${esc(a.ln)}"></label>
          </div>
          <label class="f"><span>Username</span><input type="text" id="aun" value="${esc(a.un)}"></label>
          <label class="f"><span>Email (used for Gravatar)</span><input type="text" id="aem" value="${esc(a.em)}"></label>
          <div class="row" style="justify-content:flex-end">
            <button class="btn pr" onclick="updateAccountProfile()">Update Profile</button>
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
      </div>`;
    } else if (t === 1) {
      tabContent = `<div class="card">
        <h2>Update Password</h2>
        <label class="f"><span>Current Password</span><input type="password" id="cur-pw"></label>
        <label class="f"><span>New Master Password</span><input type="password" id="new-pw"></label>
        <button class="btn pr" onclick="updatePassword()">Update Password</button>
      </div>
      <div class="card" style="margin-top:14px">
        <div class="ar" style="border:0; padding:0;">
          <div><b>Two-Factor Authentication (2FA)</b><div class="mu">Require an authenticator app code on login</div></div>
          <label class="sw"><input type="checkbox" ${a.tfa ? 'checked' : ''} onchange="a.tfa=this.checked; toast(this.checked ? '2FA enabled' : '2FA disabled');"><i></i></label>
        </div>
      </div>`;
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
        <h2>Create Automation API Key</h2>
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
          <h1 style="font-size: 24px; font-weight: 800; color: var(--tx); margin: 0; letter-spacing: -0.01em;">Account Settings</h1>
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
          <div class="vm-item-head"><b class="mono">${esc(item.v)}</b><span class="chip">${esc(sel)}</span></div>
          <div class="vm-item-sub">${esc(item.n)} &bull; Java ${esc(item.j)}</div>
          <div class="vm-item-foot">${actionBtn}</div>
        </div>`;
      }
      if (vm === 'tiles') {
        return `<div class="vm-tile-card">
          <div class="vm-item-head">
            <div><b class="mono" style="font-size:15px;">${esc(item.v)}</b><div class="mu" style="font-size:12px;margin-top:2px;">${esc(item.n)}</div></div>
            <span class="chip">${esc(sel)}</span>
          </div>
          <div class="vm-item-foot"><span class="mono mu" style="font-size:12px;">Java ${esc(item.j)}</span>${actionBtn}</div>
        </div>`;
      }
      return `<tr>
        <td class="mono"><b>${esc(item.v)}</b></td>
        <td><span class="chip">${esc(item.n)}</span></td>
        <td class="mono">Java ${esc(item.j)}</td>
        <td><span class="chip">${esc(sel)}</span></td>
        <td style="text-align:right">${actionBtn}</td>
      </tr>`;
    }).join('');
  }

  const POPULAR_VERS = new Set(['26.3', '1.21.4', '1.21.1', '1.20.4', '1.20.1', '1.19.4', '1.16.5', '1.12.2', '1.8.9']);
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
          <span class="chip">${esc(sel)}</span>
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
      <td><span class="chip">${esc(sel)}</span></td>
      <td style="text-align:right">${actionBtn}</td>
    </tr>`;
  }).join('');
}

function renderVersionRows() {
  const b = $('#vres');
  if (b) b.innerHTML = versionRowsHtml();
}

// --- Navigation Render ---
function navHtml() {
  const showGlobalNav = S.view === 'dash' || S.view === 'bill' || (S.view === 'act' && !window.location.pathname.startsWith('/server/')) || !curServer();
  const isAdminGlobal = (S.view === 'admin' || S.view === 'acc') && (!curServer() || S.view === 'dash');
  const brandHref = '/';
  const adminHref = routeUrl('admin');
  const accHref = routeUrl('acc');

  if (showGlobalNav || isAdminGlobal) {
    const isActGlobal = S.view === 'act';
    return `
    <a class="brand" href="${brandHref}" onclick="return handleNavClick(event, 'dash')" style="cursor:pointer; text-decoration:none; color:inherit;">
      ${brandLogoHtml(false)}
    </a>
    
    <button class="nsearch" onclick="openPalette(1)" aria-label="Search (Ctrl+K)">
      ${ic('search', 15)}
      <span>Search</span>
      <kbd>Ctrl K</kbd>
    </button>

    <a class="nav-item ${S.view === 'dash' ? 'on' : ''}" href="/" onclick="return handleNavClick(event, 'dash')" ${S.view === 'dash' ? 'aria-current="page"' : ''}>
      ${ic('server', 16)}
      <span>Servers</span>
    </a>
    <a class="nav-item ${isActGlobal ? 'on' : ''}" href="/activity" onclick="return handleNavClick(event, 'act', { global: true })" ${isActGlobal ? 'aria-current="page"' : ''}>
      ${ic('activity', 16)}
      <span>Activity</span>
    </a>

    <div class="ng">Billing</div>
    <a class="nav-item" href="/" onclick="return handleNavClick(event, 'dash')">
      ${ic('layers', 16)}
      <span>Overview</span>
    </a>
    <a class="nav-item ${S.view === 'bill' ? 'on' : ''}" href="/billing" onclick="return handleNavClick(event, 'bill')" ${S.view === 'bill' ? 'aria-current="page"' : ''}>
      ${ic('file', 16)}
      <span>Invoices</span>
    </a>

    <span class="sp"></span>

    <a class="nav-item ${S.view === 'admin' ? 'on' : ''}" href="${esc(adminHref)}" onclick="return handleNavClick(event, 'admin')" style="margin-bottom:6px;" ${S.view === 'admin' ? 'aria-current="page"' : ''}>
      ${ic('shield', 16)}
      <span>Open Admin Area</span>
    </a>

    <a class="chip2 ${S.view === 'acc' ? 'on' : ''}" href="${esc(accHref)}" onclick="return handleNavClick(event, 'acc')" aria-label="Account Settings" ${S.view === 'acc' ? 'aria-current="page"' : ''}>
      ${av(34)}
      <span class="who">
        <b>${esc(S.acct.fn)} ${esc(S.acct.ln)}</b>
        <span class="mu pv">${esc(S.acct.em)}</span>
      </span>
      ${ic('settings', 16)}
    </a>`;
  }

  // Inside a Server Workspace
  let group = '';
  return `
  <a class="brand" href="${brandHref}" onclick="return handleNavClick(event, 'dash')" style="cursor:pointer; text-decoration:none; color:inherit;">
    ${brandLogoHtml(false)}
  </a>
  
  <button class="nsearch" onclick="openPalette(1)" aria-label="Search (Ctrl+K)">
    ${ic('search', 15)}
    <span>Search</span>
    <kbd>Ctrl K</kbd>
  </button>

  <a class="nb" href="/" onclick="return handleNavClick(event, 'dash')" style="margin-bottom: 6px; text-decoration:none;">
    &larr; All servers
  </a>

  ${PG.map(([key, label, iconName, grp]) => {
    const header = grp !== group ? `<div class="ng">${grp}</div>` : '';
    group = grp;
    const active = S.view === key ? 'on' : '';
    const href = routeUrl(key);
    return `${header}
    <a class="nav-item ${active}" href="${esc(href)}" onclick="return handleNavClick(event, '${key}')" ${active ? 'aria-current="page"' : ''}>
      ${ic(iconName, 16)}
      <span>${label}</span>
    </a>`;
  }).join('')}

  <span class="sp"></span>

  <a class="nav-item ${S.view === 'admin' ? 'on' : ''}" href="${esc(adminHref)}" onclick="return handleNavClick(event, 'admin')" style="margin-bottom:6px;" ${S.view === 'admin' ? 'aria-current="page"' : ''}>
    ${ic('shield', 16)}
    <span>Open Admin Area</span>
  </a>

  <a class="chip2 ${S.view === 'acc' ? 'on' : ''}" href="${esc(accHref)}" onclick="return handleNavClick(event, 'acc')" aria-label="Account Settings" ${S.view === 'acc' ? 'aria-current="page"' : ''}>
    ${av(34)}
    <span class="who">
      <b>${esc(S.acct.fn)} ${esc(S.acct.ln)}</b>
      <span class="mu pv">${esc(S.acct.em)}</span>
    </span>
    ${ic('settings', 16)}
  </a>`;
}

// --- View Router & Multi-Page Navigation Engine ---
function normalizeRoute(view) {
  if (view === 'store') {
    S.plugTab = 'browse';
    return 'plug';
  }
  if (view === 'all') {
    return 'ver';
  }
  return view;
}

window.go = function(view, opts = {}) {
  const target = normalizeRoute(view);
  if (S.view === 'admin' && target !== 'admin' && (S.adminTab || 0) === 4) {
    applyBranding(S.branding);
  }
  navigateToUrl(routeUrl(target, opts), getViewDisplayLabel(target));
};

window.addEventListener('hashchange', () => {
  const raw = window.location.hash.replace('#', '');
  const hash = normalizeRoute(raw);
  if (hash && hash !== S.view && (V[hash] || hash === 'admin' || hash === 'acc')) {
    go(hash);
  }
});

// --- Main Render Function (R) ---
window.R = function() {
  const nav = $('#nav');
  const main = $('#main');
  if (nav) nav.innerHTML = navHtml();
  if (main) {
    main.innerHTML = (V[S.view] ? V[S.view]() : V.dash());
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

function bindEditorTabKey() {
  const ed = $('#ed');
  if (!ed) return;
  ed.addEventListener('keydown', e => {
    if (e.key === 'Tab') {
      e.preventDefault();
      const start = ed.selectionStart;
      const end = ed.selectionEnd;
      ed.value = ed.value.substring(0, start) + '  ' + ed.value.substring(end);
      ed.selectionStart = ed.selectionEnd = start + 2;
    }
  });
}

// --- View Transition Lifecycle Hook & Initial Route Preloader ---
async function loadInitialRouteData(view) {
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
  else if (view === 'set') {
    jobs.push(loadProperties(true));
    jobs.push(loadJavaRuntimes(false, curServer()?.node_id || 'local', true));
  }
  else if (view === 'stp' || view === 'ver') {
    jobs.push(loadVersionsCatalog());
    jobs.push(loadJavaRuntimes(false, curServer()?.node_id || 'local', true));
  }
  else if (view === 'net') {
    jobs.push(loadSFTPDetails(true));
    jobs.push(loadServerPorts(true));
  }
  else if (view === 'act') jobs.push(loadActivityLog(S.actFilter?.limit || 200, true));
  else if (view === 'admin') jobs.push(loadAdminData(true));
  else if (view === 'acc') jobs.push(loadAccountData(true));
  else if (view === 'con') jobs.push(loadServerPorts(true));
  await Promise.all(jobs).catch(() => {});
}

window.onEnterView = function(view) {
  if (view === 'pl') loadPlayersData();
  else if (view === 'fi') loadFileList(S.currentDir);
  else if (view === 'db') loadDatabases();
  else if (view === 'plug') {
    loadPluginsList();
    loadVersionsCatalog();
    if ((S.plugTab || 'browse') === 'browse') searchModrinth();
  }
  else if (view === 'bk') loadBackupsList();
  else if (view === 'set') { loadProperties(); loadJavaRuntimes(false, curServer()?.node_id || 'local'); }
  else if (view === 'stp' || view === 'ver') { loadVersionsCatalog(); loadJavaRuntimes(false, curServer()?.node_id || 'local'); }
  else if (view === 'net') { loadSFTPDetails(); loadServerPorts(); }
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
async function loadDashboard(skipRender = false, preferredRef = null) {
  if (!S.loading) S.loading = {};
  if (!skipRender && S.servers.length === 0) S.loading.dash = true;
  const prevCur = curServer();
  const prevId = preferredRef?.id || prevCur?.id || S.activeServerId;
  const prevNodeId = preferredRef?.nodeId || prevCur?.node_id || S.activeNodeId || 'local';

  try {
    const [clusterRes, nodesRes, act] = await Promise.all([
      api('/api/cluster/servers').catch(() => null),
      api('/api/nodes').catch(() => null),
      api(`/api/account/activity?limit=${S.actFilter?.limit || 200}`).catch(() => null)
    ]);

    if (clusterRes && Array.isArray(clusterRes.servers)) {
      S.servers = clusterRes.servers.map(s => ({ ...s, node_id: s.node_id || 'local' }));
      S.clusterNodeErrors = clusterRes.node_errors || {};
    } else {
      const srvs = await api('/api/servers').catch(() => []);
      S.servers = (Array.isArray(srvs) ? srvs : []).map(s => ({ ...s, node_id: s.node_id || 'local' }));
      S.clusterNodeErrors = {};
    }

    if (Array.isArray(nodesRes)) {
      S.nodes = nodesRes;
    }
    if (Array.isArray(act)) S.acct.act = act;

    if (S.servers.length > 0) {
      let idx = prevId ? S.servers.findIndex(s => s.id === prevId && (s.node_id || 'local') === prevNodeId) : -1;
      if (idx === -1 && prevId) idx = S.servers.findIndex(s => s.id === prevId);
      if (idx === -1) idx = 0;
      S.cs = idx;
      const cur = curServer();
      saveActiveServerRef(cur);
      applyServerStatsSnapshot(cur, cur.stats, false);
      S.ver = cur.version || S.ver;
      S.core = (cur.type || 'paper').charAt(0).toUpperCase() + (cur.type || 'paper').slice(1);
      connectWs(cur.id, false, cur.node_id || 'local');
    } else {
      S.cs = 0;
    }
  } catch (e) {} finally {
    S.loading.dash = false;
    if (!skipRender) R();
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
  navigateToUrl(routeUrl('admin', { adminTab: idx }), 'Admin Area');
}

function selectAccountTab(idx) {
  navigateToUrl(routeUrl('acc', { acctTab: idx }), 'Account Settings');
}

async function loadAdminData(skipRender = false) {
  const [sysRes, usersRes, tsRes, nodesRes, brandRes, actRes] = await Promise.all([
    api('/api/system').catch(() => null),
    api('/api/admin/users').catch(() => null),
    api('/api/admin/turnstile').catch(() => null),
    api('/api/nodes').catch(() => null),
    api('/api/admin/branding').catch(() => null),
    api(`/api/account/activity?limit=${S.actFilter?.limit || 200}`).catch(() => null)
  ]);

  if (sysRes) S.system = sysRes;
  if (usersRes) S.adminUsers = usersRes || [];
  if (tsRes) S.adminTurnstile = tsRes;
  if (Array.isArray(nodesRes)) S.nodes = nodesRes;
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
          <span>Daemon API Base URL</span>
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
    api_token: $('#nm-token')?.value.trim() || ''
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
  S.loading.players = true;
  if (!skipRender && S.view === 'pl') R();
  try {
    const data = await api(srvApiPath(srv, '/players'));
    S.wl = data.whitelist || [];
    S.bans = data.bans || [];
    S.ops = data.ops || [];
  } catch (e) {} finally {
    S.loading.players = false;
    if (!skipRender) R();
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
  S.loading.files = true;
  if (!skipRender && S.view === 'fi') R();
  try {
    const data = await api(srvApiPath(srv, `/files?path=${encodeURIComponent(path)}`));
    S.fileList = data || [];
    S.currentDir = path;
  } catch (e) {} finally {
    S.loading.files = false;
    if (!skipRender) R();
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
    return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
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
  const all = $$('.file-ctx-menu');
  all.forEach((m, i) => {
    if (i !== idx) m.style.display = 'none';
  });
  const m = $(`#fctx-${idx}`);
  if (m) {
    m.style.display = m.style.display === 'block' ? 'none' : 'block';
  }
}

function closeAllCtx() {
  $$('.file-ctx-menu').forEach(m => m.style.display = 'none');
}
window.addEventListener('click', closeAllCtx);

function closeEditor() {
  S.editingFile = null;
  R();
}

function setPlugTab(tab) {
  navigateToUrl(routeUrl('plug', { plugTab: tab }), tab === 'installed' ? 'Installed Plugins' : 'Plugin Store');
}

function toggleCustomJavaInput(val) {
  const wrap = $('#stp-custom-wrap');
  if (wrap) wrap.style.display = (val === 'custom') ? 'block' : 'none';
}

function toggleCustomPropJava(val) {
  const wrap = $('#prop-custom-wrap');
  if (wrap) wrap.style.display = (val === 'custom') ? 'block' : 'none';
}

async function saveStartupSettings() {
  const srv = curServer();
  if (!srv) return;

  let javaChoice = $('#stp-java')?.value;
  if (javaChoice === 'custom') {
    javaChoice = $('#stp-java-custom')?.value.trim();
  }
  const ver = $('#stp-ver')?.value?.trim() || srv.version || '1.21.4';
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

async function openFile(name) {
  const srv = curServer();
  if (!srv) return;
  const filePath = S.currentDir ? `${S.currentDir}/${name}` : name;
  S.editingFile = filePath;
  S.ftab = filePath;
  try {
    const data = await api(srvApiPath(srv, `/files/content?path=${encodeURIComponent(filePath)}`));
    S.files[filePath] = data.content != null ? data.content : data;
    R();
  } catch (e) {
    S.editingFile = null;
    toast(`Failed to open file: ${e.message}`, 'bad');
    uiAlert({
      title: 'File Open Error',
      message: `Could not open file <b>${esc(filePath)}</b>: ${esc(e.message || 'File not found or exceeds 5MB limit.')}`,
      type: 'error'
    });
  }
}

async function saveActiveFile() {
  const srv = curServer();
  const filePath = S.editingFile || S.ftab;
  if (!srv || !filePath) return;
  const ed = $('#ed');
  if (!ed) return;
  const content = ed.value;
  try {
    await api(srvApiPath(srv, '/files/content'), {
      method: 'POST',
      body: JSON.stringify({ path: filePath, content })
    });
    S.files[filePath] = content;
    toast(`Saved ${filePath.split('/').pop()} successfully`, 'ok');
  } catch (e) {
    toast(`Failed to save file: ${e.message}`, 'bad');
    uiAlert({
      title: 'Failed to Save File',
      message: `Could not save file <b>${esc(filePath)}</b>: ${esc(e.message || 'Permission denied or disk is full')}`,
      type: 'error'
    });
  }
}

function downloadFileItem(path) {
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
function syncSelectAllCheckbox() {
  const chk = $('#chk-all-files');
  if (!chk || !S.fileList.length) return;
  const selectedCount = S.fileList.filter(i => S.selectedFiles.has(i.path || i.name)).length;
  chk.checked = selectedCount === S.fileList.length;
  chk.indeterminate = selectedCount > 0 && selectedCount < S.fileList.length;
}

function toggleFileSelect(path, checked) {
  if (checked) {
    S.selectedFiles.add(path);
  } else {
    S.selectedFiles.delete(path);
  }
  R();
  syncSelectAllCheckbox();
}

function toggleSelectAllFiles(checked) {
  if (checked) {
    S.fileList.forEach(item => {
      S.selectedFiles.add(item.path || item.name);
    });
  } else {
    S.selectedFiles.clear();
  }
  R();
  syncSelectAllCheckbox();
}

function deselectAllFiles() {
  S.selectedFiles.clear();
  R();
  syncSelectAllCheckbox();
}

async function deleteSelectedFiles() {
  const count = S.selectedFiles.size;
  if (count === 0) return;
  const ok = await uiConfirm({
    title: 'Delete Selected Items',
    message: `Are you sure you want to permanently delete <b>${count}</b> selected item${count > 1 ? 's' : ''}? This action cannot be undone.`,
    confirmText: `Delete ${count} Item${count > 1 ? 's' : ''}`,
    cancelText: 'Cancel',
    isDanger: true
  });
  if (!ok) return;

  const srv = curServer();
  if (!srv) return;

  toast(`Deleting ${count} items...`);
  const paths = Array.from(S.selectedFiles);
  let failed = 0;
  for (const path of paths) {
    try {
      await api(srvApiPath(srv, `/files?path=${encodeURIComponent(path)}`), { method: 'DELETE' });
    } catch (e) {
      failed++;
    }
  }
  S.selectedFiles.clear();
  if (failed > 0) {
    toast(`Deleted ${count - failed} items (${failed} failed)`, 'bad');
    uiAlert({
      title: 'Bulk Delete Incomplete',
      message: `Deleted ${count - failed} items. <b>${failed} item(s) could not be deleted</b> (files may be locked or permission denied).`,
      type: 'error'
    });
  } else {
    toast(`Deleted ${count} items`, 'ok');
  }
  loadFileList(S.currentDir);
}

function copyFilePath(path) {
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
  const item = S.fileList[idx];
  if (!item) return;

  const menu = $('#floating-ctx');
  if (!menu) return;

  const isDir = item.is_dir;
  const path = item.path || item.name;
  const name = item.name;
  const isArchive = isArchiveFile(name);

  menu.innerHTML = `
    ${!isDir ? `
      <button onclick="openFile('${esc(name)}'); closeFloatingCtx();">
        ${ic('edit', 14)} Edit File
      </button>
      <button onclick="downloadFileItem('${esc(path)}'); closeFloatingCtx();">
        ${ic('download', 14)} Download
      </button>
      ${isArchive ? `
        <button onclick="unzipFileItem('${esc(path)}'); closeFloatingCtx();">
          ${ic('archive', 14)} Extract / Unzip
        </button>
      ` : ''}
    ` : `
      <button onclick="browseFolder('${esc(name)}'); closeFloatingCtx();">
        ${ic('folder', 14)} Open Folder
      </button>
    `}
    <button onclick="promptRenameFile('${esc(path)}'); closeFloatingCtx();">
      ${ic('settings', 14)} Rename
    </button>
    <button onclick="copyFilePath('${esc(path)}'); closeFloatingCtx();">
      ${ic('copy', 14)} Copy Path
    </button>
    <div class="floating-ctx-divider"></div>
    <button class="dg" onclick="deleteFileItem('${esc(path)}'); closeFloatingCtx();">
      ${ic('trash', 14)} Delete
    </button>
  `;

  const clientX = e?.clientX ?? (e?.touches?.[0]?.clientX ?? 100);
  const clientY = e?.clientY ?? (e?.touches?.[0]?.clientY ?? 150);
  const menuWidth = 185;
  const x = Math.max(10, Math.min(clientX, window.innerWidth - menuWidth - 10));
  const y = Math.max(10, Math.min(clientY, window.innerHeight - 250));

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

async function uploadFilesWithProgress(files) {
  const srv = curServer();
  if (!srv || !files || files.length === 0) return;

  const widget = $('#upload-progress-widget');
  const targetDir = S.currentDir || '';

  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    await new Promise((resolve) => {
      const xhr = new XMLHttpRequest();
      const formData = new FormData();
      formData.append('file', file);
      formData.append('path', targetDir);

      if (widget) {
        widget.style.display = 'block';
        widget.innerHTML = `
          <div class="row sb" style="font-size: 13px; font-weight: 600; color: #fff;">
            <span style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 220px;">
              Uploading (${i + 1}/${files.length}): ${esc(file.name)}
            </span>
            <span id="upl-pct" style="color: var(--ac); font-family: monospace;">0%</span>
          </div>
          <div class="upload-prog-bar"><i id="upl-bar" style="width: 0%;"></i></div>
        `;
      }

      xhr.upload.onprogress = (e) => {
        if (e.lengthComputable) {
          const pct = Math.round((e.loaded / e.total) * 100);
          const bar = $('#upl-bar');
          const txt = $('#upl-pct');
          if (bar) bar.style.width = `${pct}%`;
          if (txt) txt.textContent = `${pct}%`;
        }
      };

      xhr.onload = () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          toast(`Uploaded ${file.name}`, 'ok');
        } else {
          toast(`Failed to upload ${file.name}`, 'bad');
        }
        resolve();
      };

      xhr.onerror = () => {
        toast(`Network error uploading ${file.name}`, 'bad');
        resolve();
      };

      xhr.open('POST', srvApiPath(srv, '/files/upload'));
      if (S.token) {
        xhr.setRequestHeader('Authorization', `Bearer ${S.token}`);
      }
      xhr.send(formData);
    });
  }

  if (widget) {
    widget.style.display = 'none';
  }
  loadFileList(S.currentDir);
}

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
  if (!S.loading) S.loading = {};
  S.loading.plugins = true;
  if (!skipRender && S.view === 'plug' && (S.plugTab || 'browse') === 'installed') R();
  try {
    const data = await api(srvApiPath(srv, '/plugins'));
    const list = Array.isArray(data) ? data : (data?.plugins || []);
    S.plugins = list.map(p => {
      if (typeof p === 'string') return [p, 'installed', true, ''];
      const fname = p.filename || p.name || 'plugin.jar';
      const meta = p.size ? formatFileSize(p.size) : (p.version || 'installed');
      return [fname, meta, true, p.modified || ''];
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
  const ok = await uiConfirm({
    title: 'Remove Plugin',
    message: `Are you sure you want to delete <b>${esc(name)}</b> from <b>${esc(srv.name)}</b>? Restart the server afterward to unload the plugin.`,
    confirmText: 'Remove Plugin',
    cancelText: 'Cancel',
    isDanger: true,
    icon: 'trash'
  });
  if (!ok) return;

  try {
    await api(srvApiPath(srv, `/plugins?filename=${encodeURIComponent(name)}`), { method: 'DELETE' });
    toast(`Removed plugin ${name}`);
    loadPluginsList();
  } catch (e) {}
}

function togglePlugin(idx, enabled) {
  S.plugins[idx][2] = enabled;
  toast(`${S.plugins[idx][0]} ${enabled ? 'enabled' : 'disabled'}. Restart server to apply.`);
}

let searchTimeout = null;
function searchModrinthDebounced() {
  clearTimeout(searchTimeout);
  searchTimeout = setTimeout(() => searchModrinth(), 350);
}

function renderModrinthHitsHtml(hits) {
  const list = Array.isArray(hits) ? hits : [];
  const storeVm = getViewMode('plug_store', 'list');
  if (!list.length) {
    return '<div class="mu" style="padding:24px; text-align:center; grid-column:1/-1;">No matching plugins found on Modrinth.</div>';
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
    const isInstalling = Boolean(S.installingPlugins && S.installingPlugins.has(slugVal));
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
  const grid = $('#store-grid');
  if (!skipDomWrite && grid) {
    grid.innerHTML = loaderHtml('Searching Modrinth plugins...');
  }
  const q = S.q || '';
  const loader = S.plugLoader || '';
  const ver = S.plugVer || '';
  const sort = S.plugSort || 'relevance';
  try {
    const params = new URLSearchParams({
      query: q,
      loader,
      version: ver,
      sort,
      limit: '24'
    });
    const res = await api(`/api/plugins/search?${params.toString()}`);
    S.lastModrinthHits = res.hits || [];
    if (!skipDomWrite && grid) {
      grid.innerHTML = renderModrinthHitsHtml(S.lastModrinthHits);
    }
  } catch (e) {
    S.lastModrinthHits = [];
    if (!skipDomWrite && grid) {
      grid.innerHTML = '<div class="mu" style="padding:24px; text-align:center; grid-column:1/-1;">Error reaching Modrinth API.</div>';
    }
  }
}

async function installPlugin(projectSlug, title, btnEl) {
  const srv = curServer();
  if (!srv || !projectSlug) return;
  if (S.installingPlugins && S.installingPlugins.has(projectSlug)) return;

  const targetFolder = String(S.core || '').toLowerCase() === 'fabric' ? 'mods/' : 'plugins/';
  const ok = await uiConfirm({
    title: 'Install Plugin from Modrinth',
    message: `Download and install <b>${esc(title || projectSlug)}</b> into <code>${targetFolder}</code> on <b>${esc(srv.name)}</b>?`,
    confirmText: 'Download & Install',
    cancelText: 'Cancel',
    isDanger: false,
    icon: 'download'
  });
  if (!ok) return;

  if (!S.installingPlugins) S.installingPlugins = new Set();
  S.installingPlugins.add(projectSlug);

  const findBtn = () => {
    if (btnEl && document.body.contains(btnEl)) return btnEl;
    if (typeof CSS !== 'undefined' && CSS.escape) {
      return document.querySelector(`#store-grid button[data-slug="${CSS.escape(projectSlug)}"]`);
    }
    return null;
  };

  const activeBtn = findBtn();
  if (activeBtn) {
    setBtnLoading(activeBtn, true, 'Installing...');
  }
  toast(`Downloading & installing ${title || projectSlug}...`);
  try {
    await api(srvApiPath(srv, '/plugins/install'), {
      method: 'POST',
      body: JSON.stringify({ slug: projectSlug, project_id: projectSlug })
    });
    S.installingPlugins.delete(projectSlug);
    toast(`Successfully installed ${title || projectSlug}!`);
    const curBtn = findBtn();
    if (curBtn && curBtn.parentElement) {
      curBtn.parentElement.innerHTML = '<span class="badge online"><i class="dot"></i>Installed</span>';
    }
    loadPluginsList();
  } catch (e) {
    S.installingPlugins.delete(projectSlug);
    const curBtn = findBtn();
    if (curBtn) {
      setBtnLoading(curBtn, false);
      if (curBtn.innerHTML.includes('Installing')) {
        curBtn.innerHTML = `${ic('download', 14)} Install`;
      }
    }
  }
}

// --- Backups ---
async function loadBackupsList(skipRender = false) {
  const srv = curServer();
  if (!srv) return;
  if (!S.loading) S.loading = {};
  S.loading.backups = true;
  if (!skipRender && S.view === 'bk') R();
  try {
    const list = await api(srvApiPath(srv, '/backups'));
    S.bk = list || [];
  } catch (e) {} finally {
    S.loading.backups = false;
    if (!skipRender) R();
  }
}

async function createBackup() {
  const srv = curServer();
  if (!srv) return;
  const ok = await uiConfirm({
    title: 'Create Full Server Backup',
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
    const needsJavaSwitch = curJava !== recJava;
    const javaInfo = getPremadeJavaStatusMap(srv?.node_id || 'local').get(recJava);
    const javaMissingNote = javaInfo && !javaInfo.installed
      ? `<span class="mu" style="font-size: 11.5px; display: block; margin-top: 4px; color: #fbbf24;">Note: Java ${esc(recJava)} is not installed yet — install it from Startup Settings or Admin Area before starting.</span>`
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
            Download and install <b>${esc(software)} ${esc(version)}</b> onto <b>${esc(srv.name)}</b>? This will replace your current server JAR and restart the instance.
            <div style="margin-top: 14px; padding: 12px 14px; background: rgba(56, 122, 255, 0.08); border: 1px solid rgba(56, 122, 255, 0.25); border-radius: 10px;">
              <label style="display: flex; align-items: flex-start; gap: 10px; cursor: pointer; margin: 0;">
                <input type="checkbox" id="confirm-bkf" checked style="margin-top: 2px;">
                <div>
                  <b style="color: var(--tx); display: block; font-size: 13.5px;">Back up the world first (Recommended)</b>
                  <span class="mu" style="font-size: 12px; display: block; margin-top: 2px;">Creates a full <code>before-${esc(version)}</code> backup archive before downloading and replacing the server JAR.</span>
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
    toast(`Downloading ${software} ${version} server jar...`);
    const softRes = await api(srvApiPath(srv, '/software'), {
      method: 'POST',
      body: JSON.stringify({ software, version })
    });
    srv.type = softRes?.software || String(software).toLowerCase();
    srv.version = softRes?.version || version;
    if (softRes?.jar_file) srv.jar_file = softRes.jar_file;

    if (choice.syncJava && choice.recJava) {
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

    S.core = software;
    S.ver = version;
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

async function loadProperties(skipRender = false) {
  const srv = curServer();
  if (!srv) return;
  try {
    const props = await api(srvApiPath(srv, '/properties'));
    S.props.motd = props['motd'] || S.props.motd;
    S.props.max = parseInt(props['max-players'] || '20');
    S.props.diff = props['difficulty'] || S.props.diff;
    S.props.mode = props['gamemode'] || S.props.mode;
    S.props.pvp = props['pvp'] !== 'false';
    S.props.wl = props['white-list'] === 'true';
    S.props.cmd = props['enable-command-block'] === 'true';
    if (!skipRender) R();
  } catch (e) {}
}

async function saveProperties() {
  const srv = curServer();
  if (!srv) return;
  const newPort = parseInt($('#prop-port')?.value) || srv.port || 25565;
  const newRam = parseInt($('#prop-ram')?.value) || getServerMaxRamMB(srv);
  let newJava = $('#prop-java')?.value;
  if (newJava === 'custom') {
    newJava = $('#prop-java-custom')?.value.trim();
  }

  const payload = {
    'motd': S.props.motd,
    'server-port': String(newPort),
    'max-players': String(S.props.max),
    'difficulty': S.props.diff,
    'gamemode': S.props.mode,
    'pvp': String(S.props.pvp),
    'white-list': String(S.props.wl),
    'enable-command-block': String(S.props.cmd)
  };

  try {
    // 1. Write server.properties file
    await api(srvApiPath(srv, '/properties'), {
      method: 'POST',
      body: JSON.stringify(payload)
    });

    // 2. Persist port, memory, and java runtime to SQLite DB and process runner
    await api(srvApiPath(srv, ''), {
      method: 'PUT',
      body: JSON.stringify({
        name: srv.name,
        description: srv.description,
        type: srv.type,
        version: srv.version,
        jar_file: srv.jar_file,
        port: newPort,
        memory_min: Math.min(1024, newRam),
        memory_max: newRam,
        java_path: newJava || srv.java_path,
        java_args: srv.java_args,
        auto_restart: srv.auto_restart
      })
    });

    srv.port = newPort;
    srv.memory_max = newRam;
    srv.memory_min = Math.min(1024, newRam);
    if (newJava) srv.java_path = newJava;

    toast('Server properties, RAM allocation & Java runtime saved! Restart server to apply.', 'ok');
    R();
  } catch (e) {
    toast(e.message || 'Failed to save server settings', 'bad');
    uiAlert({
      title: 'Failed to Save Settings',
      message: `Could not save server properties: ${esc(e.message || 'Validation error')}`,
      type: 'error'
    });
  }
}

async function loadSFTPDetails(skipRender = false) {
  const srv = curServer();
  if (!srv) return;
  try {
    const data = await api(srvApiPath(srv, '/sftp'));
    S.sftp = data;
    if (!skipRender) R();
  } catch (e) {}
}

// --- Account Management ---
async function loadAccountData(skipRender = false) {
  try {
    const savedProf = JSON.parse(localStorage.getItem('neopanel_profile') || '{}');
    const prof = await api('/api/account/profile');
    S.acct.un = savedProf.un || prof.username || S.acct.un;
    S.acct.em = savedProf.em || prof.email || S.acct.em;
    S.acct.fn = savedProf.fn || prof.first_name || 'Admin';
    S.acct.ln = savedProf.ln !== undefined ? savedProf.ln : (prof.last_name || '');
    S.acct.h = await computeGravatarHash(S.acct.em);

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
  const fn = $('#afn')?.value?.trim();
  const ln = $('#aln')?.value?.trim() ?? '';
  const un = $('#aun')?.value?.trim();
  const em = $('#aem')?.value?.trim();
  if (fn) S.acct.fn = fn;
  S.acct.ln = ln;
  if (un) S.acct.un = un;
  if (em) {
    S.acct.em = em;
    S.acct.h = await computeGravatarHash(em);
  }
  localStorage.setItem('neopanel_profile', JSON.stringify({
    fn: S.acct.fn,
    ln: S.acct.ln,
    un: S.acct.un,
    em: S.acct.em
  }));
  toast('Profile information saved!', 'ok');
  R();
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

function switchServer(index) {
  S.cs = index;
  const srv = S.servers[index];
  if (!srv) return;
  const nodeId = srv.node_id || 'local';
  saveActiveServerRef(srv);
  applyServerStatsSnapshot(srv, srv.stats, false);
  S.ver = srv.version || S.ver;
  S.core = (srv.type || 'paper').charAt(0).toUpperCase() + (srv.type || 'paper').slice(1);
  if (S.activeServerId !== srv.id || (S.activeNodeId || 'local') !== nodeId) {
    S.log = [];
    S.ports = [];
    S.databases = [];
    S.dbRevealIds = new Set();
    S.sqlStudio = null;
    S.fileList = [];
    S.currentDir = '';
    S.editingFile = null;
    S.selectedFiles = new Set();
    S.bk = [];
    S.plugins = [];
    S.sftp = null;
    S.startupFailures = Number(srv.stats?.startup_failures) || 0;
    S.startupHalted = Boolean(srv.stats?.startup_halted);
    connectWs(srv.id, false, nodeId);
    loadServerPorts();
  }
  R();
  if (S.view === 'net') {
    loadSFTPDetails();
  } else if (S.view && !['dash', 'con', 'admin', 'acc'].includes(S.view)) {
    onEnterView(S.view);
  }
  toast(`Switched to ${srv.name}`);
}

// --- Command Palette (Ctrl + K) ---
function paletteItems() {
  const q = (S.pq || '').toLowerCase();
  const items = [
    ...PG.map(p => ['Pages', p[1], 'Go to page', () => go(p[0])]),
    ['Pages', 'Admin Area', 'Global cluster & daemon health', () => go('admin')],
    ['Pages', 'Account Settings', 'Profile, appearance, security', () => go('acc')],
    ...S.servers.map(s => ['Servers', s.name, `${(s.stats?.status || s.status) === 'running' ? 'Online' : 'Offline'} (Port ${s.port})`, () => {
      saveActiveServerRef(s);
      navigateToUrl(serverRouteUrl(s.node_id || 'local', s.id, 'con'), s.name || 'Server Console');
    }]),
    ...S.players.map(p => ['Players', p.n, p.op ? 'Operator' : 'Online', () => { S.ptab = 'online'; go('pl'); }]),
    ['Actions', 'Toggle Live Console', 'Ctrl + `', () => toggleFlyConsole()],
    ['Actions', 'Create Backup', 'Backups', () => createBackup()],
    ['Actions', 'Restart Server', 'Power', () => power('restart')],
    ['Actions', 'Toggle Dark / Light Theme', 'Appearance', () => toggleTheme()],
    ['Actions', 'Toggle Privacy Mode', 'Privacy', () => togglePrivacy(!S.acct.priv)]
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
    return;
  }
  if (p || $('#login')) return;

  p = document.createElement('div');
  p.id = 'pal';
  p.className = 'pal';
  p.innerHTML = `<div class="pbox" role="dialog" aria-label="Command Palette">
    <input id="pq" class="pin" placeholder="Search servers, pages, players, actions..." autocomplete="off">
    <div id="pl2" class="plist"></div>
  </div>`;
  document.body.appendChild(p);

  S.pq = '';
  S.pi = 0;
  renderPaletteList();

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
window.setCreateRam = function(mb) {
  const inp = $('#cs-mem');
  if (inp) inp.value = mb;
  $$('.ram-pill').forEach(btn => {
    btn.classList.toggle('pr', btn.dataset.ram == mb);
  });
};

window.syncCreateJavaFromVersion = function(ver) {
  const engine = $('#cs-type')?.value || 'paper';
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
};

function syncCreateSelectValue(ver) {
  const sel = $('#cs-ver-sel');
  if (!sel) return;
  const v = String(ver || '').trim();
  if (v && Array.from(sel.options).some(o => o.value === v)) {
    sel.value = v;
    if (typeof sel._cselectRebuild === 'function') sel._cselectRebuild();
  }
}

window.onCreateVersionInput = function(ver) {
  syncCreateSelectValue(ver);
  syncCreateJavaFromVersion(ver);
};

window.setCreateEngine = function(type) {
  const sel = $('#cs-type');
  if (sel) sel.value = type;
  $$('.engine-card').forEach(card => {
    card.classList.toggle('on', card.dataset.engine === type);
  });

  const verInp = $('#cs-ver');
  const portInp = $('#cs-port');
  if (type === 'velocity') {
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
      verInp.value = '1.21.4';
    }
  }

  const csVerSel = $('#cs-ver-sel');
  if (csVerSel) {
    const curVer = verInp?.value?.trim() || '1.21.4';
    const vers = getAllMcVersions(type);
    const list = (curVer && !vers.includes(curVer)) ? [curVer, ...vers] : vers;
    csVerSel.innerHTML = list.map(v => `<option value="${esc(v)}" ${v === curVer ? 'selected' : ''}>${esc(v)}</option>`).join('');
    if (typeof csVerSel._cselectRebuild === 'function') csVerSel._cselectRebuild();
  }
  syncCreateJavaFromVersion(verInp?.value || '');
};

window.setCreateVersion = function(ver) {
  const inp = $('#cs-ver');
  if (inp) inp.value = ver;
  syncCreateSelectValue(ver);
  syncCreateJavaFromVersion(ver);
};

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
  const container = $('#modal-container');
  if (!container) return;

  const nodesList = (Array.isArray(S.nodes) && S.nodes.length > 0)
    ? S.nodes
    : [{ id: 'local', name: 'Local Master Node', location: 'Primary Host', online: true }];
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
            <span>Server Instance Name</span>
            <input type="text" id="cs-name" required placeholder="e.g. Survival SMP, Proxy Network, Minigames" value="Survival Server">
          </label>
          ${nodesList.length > 1 ? `
          <label class="f" style="margin-bottom:0;">
            <span>Target Cluster Node</span>
            <select id="cs-node" onchange="onCreateNodeChange(this.value)">
              ${nodesList.map(n => `<option value="${esc(n.id)}" ${n.id === defaultNodeId ? 'selected' : ''}>${esc(n.name)}${n.location ? ` (${esc(n.location)})` : ''}${n.online ? '' : ' [Offline]'}</option>`).join('')}
            </select>
          </label>` : `<input type="hidden" id="cs-node" value="${esc(defaultNodeId)}">`}
        </div>

        <!-- Category: Game Server vs Network Proxy (All 11 Engines) -->
        <div style="margin-bottom:14px;">
          <span class="mu" style="display:block; margin-bottom:6px; font-size:13px; font-weight:600;">Engine / Flavor</span>
          <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(112px, 1fr)); gap:8px;">
            ${CORES.map(([label, desc]) => {
              const slug = label.toLowerCase();
              return `
              <button type="button" class="card engine-card sel ${slug === 'paper' ? 'on' : ''}" data-engine="${esc(slug)}" onclick="setCreateEngine('${esc(slug)}')" style="padding:9px 8px; text-align:center;">
                <b>${esc(label)}</b>
                <span class="mu" style="font-size:10.5px; display:block; margin-top:2px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap;" title="${esc(desc)}">${esc(desc)}</span>
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

        <!-- Version Searchable Catalog + Custom Input + Popular Pills -->
        <div style="margin-bottom:14px;">
          <div class="row sb" style="margin-bottom:6px; flex-wrap:wrap; gap:6px;">
            <span class="mu" style="font-size:13px; font-weight:600;">Minecraft / Release Version (All Releases)</span>
            <div class="row" style="gap:4px; flex-wrap:wrap;">
              ${['26.3', '1.21.4', '1.21.1', '1.20.4', '1.20.1', '1.16.5', '1.12.2', '1.8.9'].map(v => `
                <button type="button" class="btn sm ver-pill ${v === '1.21.4' ? 'pr' : ''}" data-ver="${v}" onclick="setCreateVersion('${v}')">${v}</button>
              `).join('')}
            </div>
          </div>
          <div style="display:grid; grid-template-columns: minmax(145px, 185px) 1fr; gap:8px;">
            <select id="cs-ver-sel" data-searchable="1" onchange="setCreateVersion(this.value)">
              ${initialVers.map(v => `<option value="${esc(v)}" ${v === '1.21.4' ? 'selected' : ''}>${esc(v)}</option>`).join('')}
            </select>
            <input type="text" id="cs-ver" value="1.21.4" required placeholder="e.g. 26.3, 1.21.4, 1.12.2, 1.8.9, or custom" oninput="onCreateVersionInput(this.value)">
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
              ${renderJavaSelectOptions('21', false, '1.21.4', 'paper', defaultNodeId)}
            </select>
          </label>
        </div>
        <div style="margin-bottom:14px;">
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

        <!-- EULA & Optimization Flags -->
        <div style="background:var(--pn2); border:1px solid var(--ln); border-radius:8px; padding:12px; margin-bottom:16px;">
          <label class="row" style="margin-bottom:8px; cursor:pointer;">
            <input type="checkbox" id="cs-eula" checked style="width:auto; margin-right:8px;">
            <span>Automatically agree to Mojang EULA (<code>eula=true</code>)</span>
          </label>
          <label class="row" style="margin-bottom:8px; cursor:pointer;">
            <input type="checkbox" id="cs-dl" checked style="width:auto; margin-right:8px;">
            <span>Automatically download official server jar immediately</span>
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
  const container = $('#modal-container');
  if (container) container.innerHTML = '';
  if (_uiModalResolver) {
    const resolve = _uiModalResolver;
    _uiModalResolver = null;
    resolve(null);
  }
}

async function handleCreateServer(e) {
  e.preventDefault();
  const btn = $('#cs-submit-btn');
  if (btn) {
    btn.disabled = true;
    btn.textContent = 'Provisioning server...';
  }

  const targetNodeId = $('#cs-node')?.value || 'local';
  const name = $('#cs-name')?.value.trim() || 'Minecraft Server';
  const type = $('#cs-type')?.value || 'paper';
  const version = $('#cs-ver')?.value.trim() || '1.21.4';
  const memoryMax = parseInt($('#cs-mem')?.value) || 2048;
  const memoryMin = Math.min(1024, memoryMax);
  const port = parseInt($('#cs-port')?.value) || 25565;
  const javaPath = $('#cs-java')?.value || '21';
  const autoDownload = $('#cs-dl')?.checked ?? true;
  const useAikar = $('#cs-aikar')?.checked;

  let javaArgs = '-XX:+UseG1GC';
  if (useAikar) {
    javaArgs = '-XX:+UseG1GC -XX:+ParallelRefProcEnabled -XX:MaxGCPauseMillis=200 -XX:+UnlockExperimentalVMOptions -XX:+DisableExplicitGC -XX:+AlwaysPreTouch';
  }

  toast(`Creating and configuring ${name}...`);

  try {
    const res = await api(nodeApiPath(targetNodeId, '/api/servers'), {
      method: 'POST',
      body: JSON.stringify({
        name,
        type,
        version,
        memory_min: memoryMin,
        memory_max: memoryMax,
        port,
        java_path: javaPath,
        java_args: javaArgs,
        auto_download: autoDownload,
        auto_restart: true
      })
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
      saveActiveServerRef({ id: createdId, node_id: targetNodeId });
      navigateToUrl(serverRouteUrl(targetNodeId, createdId, 'con'), name);
    } else {
      navigateToUrl('/', 'Servers Overview');
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

    await api(nodeApiPath(targetNodeId, `/api/servers/${encodeURIComponent(srvId)}`), { method: 'DELETE' });
    const saved = getSavedActiveServerRef();
    if (saved && saved.id === srvId && (saved.nodeId || 'local') === targetNodeId) {
      try { localStorage.removeItem('neopanel_active_srv'); } catch (e) {}
    }
    toast('Server deleted successfully', 'ok');
    closeModal();
    navigateToUrl('/', 'Servers Overview');
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
  const ver = $('#reinst-version')?.value.trim() || '1.21.4';
  const el = $('#reinst-java-rec');
  if (el) el.textContent = `Java ${getRecommendedJavaForMcVersion(ver, sw)}`;
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
  const sel = $('#reinst-ver-sel');
  const inp = $('#reinst-version');
  const curV = inp?.value?.trim() || '1.21.4';
  if (sw === 'velocity' && !curV.includes('SNAPSHOT')) {
    if (inp) inp.value = '3.4.0-SNAPSHOT';
  } else if (sw === 'waterfall' && curV !== '1.20') {
    if (inp) inp.value = '1.20';
  } else if (sw === 'bungeecord') {
    if (inp) inp.value = 'latest';
  }
  if (sel) {
    const newCur = inp?.value?.trim() || '1.21.4';
    const vers = getAllMcVersions(sw);
    const list = (newCur && !vers.includes(newCur)) ? [newCur, ...vers] : vers;
    sel.innerHTML = list.map(v => `<option value="${esc(v)}" ${v === newCur ? 'selected' : ''}>${esc(v)}</option>`).join('');
    if (typeof sel._cselectRebuild === 'function') sel._cselectRebuild();
  }
  updateReinstallJavaPreview();
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
  const curVer = targetSrv.version || '1.21.4';
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
        Reinstalls the server software from <b>mcjarfiles.com</b> official upstream distribution. This will replace the server executable JAR.
      </p>

      <form id="reinstall-form" data-srv-id="${esc(targetId)}" data-node-id="${esc(targetNodeId)}" onsubmit="submitReinstall(event, this.dataset.srvId, this.dataset.nodeId)">
        <label class="f" style="margin-bottom:12px;">
          <span>Minecraft Engine / Server Software</span>
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
          </select>
        </label>

        <div class="f" style="margin-bottom:16px;">
          <span style="display:block; margin-bottom:6px; font-size:13px; color:var(--tx-muted); font-weight:500;">Version</span>
          <div style="display:grid; grid-template-columns: minmax(145px, 180px) 1fr; gap:8px;">
            <select id="reinst-ver-sel" data-searchable="1" onchange="setReinstallVersion(this.value)">
              ${verList.map(v => `<option value="${esc(v)}" ${v === curVer ? 'selected' : ''}>${esc(v)}</option>`).join('')}
            </select>
            <input type="text" id="reinst-version" value="${esc(curVer)}" placeholder="26.3, 1.21.4, 1.12.2, or latest" required oninput="syncReinstallSelectFromInput(this.value); updateReinstallJavaPreview();">
          </div>
        </div>

        <div style="background:rgba(255,255,255,0.03); border:1px solid var(--border); border-radius:8px; padding:12px 14px; margin-bottom:14px;">
          <b style="font-size:13px; color:#fff; display:block; margin-bottom:8px;">Preservation Options</b>
          <label class="row" style="cursor:pointer; margin-bottom:8px; font-size:13px; gap:10px;">
            <input type="checkbox" id="reinst-world" checked>
            <span>Preserve Worlds (keep <code>world</code>, <code>world_nether</code>, <code>world_the_end</code>)</span>
          </label>
          <label class="row" style="cursor:pointer; margin-bottom:8px; font-size:13px; gap:10px;">
            <input type="checkbox" id="reinst-config" checked>
            <span>Preserve Configurations (keep <code>server.properties</code>, <code>bukkit.yml</code>, <code>spigot.yml</code>, <code>paper-*.yml</code>)</span>
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
  loadVersionsCatalog();
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
  const preservePlugins = $('#reinst-plugins')?.checked ?? true;
  const syncJava = $('#reinst-sync-java')?.checked ?? false;

  toast(`Initiating reinstallation of ${software} ${version}...`);

  try {
    const res = await api(nodeApiPath(targetNodeId, `/api/servers/${encodeURIComponent(srvId)}/reinstall`), {
      method: 'POST',
      body: JSON.stringify({
        software,
        version,
        preserve_world: preserveWorld,
        preserve_config: preserveConfig,
        preserve_plugins: preservePlugins
      })
    });

    const targetSrv = findServerRef(srvId, targetNodeId) || curServer();
    if (targetSrv) {
      targetSrv.type = software;
      targetSrv.version = version;
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
      S.ver = version;
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
        Please verify the version and software engine, or check network connectivity to mcjarfiles.com.`,
      type: 'error'
    });
  }
}

// --- Admin Panel Management (Users, GC, Customization) ---
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
          <span>Username</span>
          <input type="text" id="nu-username" required placeholder="e.g. operator" autocomplete="off">
        </label>

        <label class="f" style="margin-bottom:12px;">
          <span>Password</span>
          <input type="password" id="nu-password" required placeholder="Minimum 6 characters" autocomplete="new-password">
        </label>

        <label class="f" style="margin-bottom:20px;">
          <span>Role</span>
          <select id="nu-role" style="height:38px;">
            <option value="user">Standard User (Access assigned servers)</option>
            <option value="admin">Administrator (Full cluster management)</option>
          </select>
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
  const password = $('#nu-password')?.value;
  const role = $('#nu-role')?.value || 'user';

  if (!username || !password) {
    toast('Username and password are required', 'bad');
    if (btn) btn.disabled = false;
    return;
  }

  try {
    await api('/api/admin/users', {
      method: 'POST',
      body: JSON.stringify({ username, password, role })
    });
    closeModal();
    toast(`User "${username}" created successfully`, 'ok');
    await loadAdminData();
  } catch (err) {
    toast(err.message || 'Failed to create user', 'bad');
    if (btn) btn.disabled = false;
    uiAlert({
      title: 'User Creation Failed',
      message: `Failed to create user account <b>${esc(username)}</b>:<br><br>${esc(err.message || 'Username already exists or password too short')}`,
      type: 'error'
    });
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
    title: 'Run Daemon Garbage Collection',
    message: 'Trigger an immediate Go runtime garbage collection cycle across the NeoPanel daemon to reclaim unused heap memory?',
    confirmText: 'Run Garbage Collection',
    cancelText: 'Cancel',
    isDanger: false,
    icon: 'cpu'
  });
  if (!ok) return;

  toast('Running Go runtime garbage collection...');
  try {
    const res = await api('/api/admin/gc', { method: 'POST' });
    toast(res.message || 'Garbage collection completed', 'ok');
    await loadAdminData();
  } catch (err) {
    toast(err.message || 'Failed to run GC', 'bad');
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
    toast('Daemon network & cluster defaults saved!', 'ok');
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

async function checkAuth() {
  startPageProgress(22);
  const savedMode = localStorage.getItem('neopanel_mode');
  if (savedMode) setAppearanceMode(savedMode, true);
  if (localStorage.getItem('neopanel_compact') === 'true') toggleCompact(true);
  if (localStorage.getItem('neopanel_priv') === 'true') togglePrivacy(true);

  const route = parseCurrentRoute(window.location.pathname, window.location.hash);
  S.view = normalizeRoute(route.view || 'dash');
  if (route.adminTab !== undefined) S.adminTab = route.adminTab;
  if (route.acctTab !== undefined && S.acct) S.acct.tab = route.acctTab;
  if (route.plugTab !== undefined) S.plugTab = route.plugTab;
  const preferredRef = route.serverId ? { id: route.serverId, nodeId: route.nodeId || 'local' } : null;

  try {
    setPageProgress(42);
    const [, status] = await Promise.all([
      loadPublicBranding().catch(() => {}),
      fetch('/api/auth/status').then(r => r.json())
    ]);
    S.turnstileEnabled = Boolean(status.turnstile_enabled && status.turnstile_site_key);
    S.turnstileSiteKey = S.turnstileEnabled ? status.turnstile_site_key : '';

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
    const savedProf = JSON.parse(localStorage.getItem('neopanel_profile') || '{}');
    S.user = user;
    S.acct.un = savedProf.un || user.username || S.acct.un;
    S.acct.em = savedProf.em || user.email || S.acct.em;
    if (savedProf.fn) S.acct.fn = savedProf.fn;
    if (savedProf.ln !== undefined) S.acct.ln = savedProf.ln;
    S.acct.h = await computeGravatarHash(S.acct.em);

    setPageProgress(65);
    await loadDashboard(true, preferredRef || getSavedActiveServerRef());

    const globalOnlyViews = ['dash', 'bill', 'admin', 'acc'];
    if (!globalOnlyViews.includes(S.view) && !(S.view === 'act' && route.global)) {
      if (S.servers.length === 0) {
        S.view = 'dash';
        try { history.replaceState(null, '', '/'); } catch (e) {}
      }
    }
    if (route.legacyHash) {
      try { history.replaceState(null, '', routeUrl(S.view, route)); } catch (e) {}
    }

    setPageProgress(82);
    await loadInitialRouteData(S.view);
    if (!['plug', 'stp', 'ver'].includes(S.view)) {
      loadVersionsCatalog();
    }

    applyBranding(S.branding);
    R();
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
  const footTxt = (getBrand('footer_text') || 'Powered by NeoPanel Daemon').trim();
  const d = document.createElement('div');
  d.id = 'login';
  d.className = 'login';
  d.innerHTML = `<div class="lbox">
    <div class="llogo">${brandLogoHtml(true)}</div>
    <div class="lg">First-Time Initialization</div>
    <h1 class="lt">Create Admin Master Account</h1>
    <label class="lf">
      ${ic('user', 16)}
      <input id="su-name" type="text" placeholder="Username" autocomplete="username" onkeydown="if(event.key==='Enter')submitSetup()">
    </label>
    <label class="lf">
      ${ic('lock', 16)}
      <input id="su-pw" type="password" placeholder="Master Password" autocomplete="new-password" onkeydown="if(event.key==='Enter')submitSetup()">
      <button class="eye" type="button" onclick="const p=$('#su-pw'); p.type=p.type==='password'?'text':'password';" aria-label="Toggle password visibility">
        ${ic('eye', 16)}
      </button>
    </label>
    ${S.turnstileEnabled ? '<div id="turnstile-setup" class="turnstile-wrap"></div>' : ''}
    <div class="lerr" id="lerr" role="alert"></div>
    <button class="lbtn" onclick="submitSetup()">Initialize &amp; Launch</button>
    <div class="lsm foot">${esc(footTxt)}</div>
  </div>`;
  applyLoginOverlayBackground(d);
  document.body.appendChild(d);
  if (S.turnstileEnabled) mountTurnstileWidget('#turnstile-setup');
  $('#su-name')?.focus();
}

async function submitSetup() {
  const username = $('#su-name')?.value.trim();
  const password = $('#su-pw')?.value;
  const err = $('#lerr');
  if (!username || !password) {
    if (err) err.textContent = 'Username and password are required.';
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
  const hour = new Date().getHours();
  const greeting = hour < 5 ? 'Good evening' : hour < 12 ? 'Good morning' : hour < 18 ? 'Good afternoon' : 'Good evening';
  const brandName = (getBrand('brand_name') || 'NeoPanel').trim() || 'NeoPanel';
  const footTxt = (getBrand('footer_text') || 'Powered by NeoPanel Daemon').trim();

  const d = document.createElement('div');
  d.id = 'login';
  d.className = 'login';
  d.innerHTML = `<div class="lbox">
    <div class="llogo">${brandLogoHtml(true)}</div>
    <div class="lg">${greeting}</div>
    <h1 class="lt">Log in to ${esc(brandName)}</h1>
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
    ${S.turnstileEnabled ? '<div id="turnstile-login" class="turnstile-wrap"></div>' : ''}
    <div class="lerr" id="lerr" role="alert"></div>
    <button class="lbtn" onclick="submitLogin()">Login</button>
    <div class="lsm foot">${esc(footTxt)}</div>
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

    if (res.token) {
      S.token = res.token;
      S.turnstileToken = '';
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

async function logout() {
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

  S.token = '';
  S.activeServerId = null;
  S.activeNodeId = null;
  S.wsState = 'disconnected';
  localStorage.removeItem('neopanel_token');
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
window.toast = toast;
window.power = power;
window.switchServer = switchServer;
window.openPalette = openPalette;
window.execPaletteItem = execPaletteItem;
window.toggleCmdMenu = toggleCmdMenu;
window.execQuickCmd = execQuickCmd;
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
window.openAddUserModal = openAddUserModal;
window.submitAddUser = submitAddUser;
window.deleteAdminUser = deleteAdminUser;
window.triggerAdminGC = triggerAdminGC;
window.saveAdminDaemonSettings = saveAdminDaemonSettings;
window.saveAdminTurnstileSettings = saveAdminTurnstileSettings;
window.toggleCustomJavaInput = toggleCustomJavaInput;
window.toggleCustomPropJava = toggleCustomPropJava;
window.isArchiveFile = isArchiveFile;
window.unzipFileItem = unzipFileItem;
window.unzipSelectedArchive = unzipSelectedArchive;
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
window.onCreateNodeChange = onCreateNodeChange;
window.handleNavClick = handleNavClick;
window.handleServerCardClick = handleServerCardClick;
window.setViewMode = setViewMode;
window.getViewMode = getViewMode;
window.routeUrl = routeUrl;
window.serverRouteUrl = serverRouteUrl;
window.onCreateVersionInput = onCreateVersionInput;
window.setCreateVersion = setCreateVersion;
window.onReinstallEngineChange = onReinstallEngineChange;
window.setReinstallVersion = setReinstallVersion;
window.syncReinstallSelectFromInput = syncReinstallSelectFromInput;
window.refreshOpenVersionSelects = refreshOpenVersionSelects;
window.renderModrinthHitsHtml = renderModrinthHitsHtml;

// --- Custom Glassmorphic Selects System ---
function closeAllCustomSelects() {
  document.querySelectorAll('.cselect-wrap.open').forEach(el => {
    el.classList.remove('open');
    const trigger = el.querySelector('.cselect-trigger');
    if (trigger) trigger.setAttribute('aria-expanded', 'false');
  });
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

    function rebuildOptions() {
      menu.innerHTML = '';
      const opts = Array.from(sel.options);
      const isSearchable = sel.dataset.searchable === '1' || opts.length > 12;

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
        const rect = wrap.getBoundingClientRect();
        const spaceBelow = window.innerHeight - rect.bottom;
        const hasSearch = Boolean(menu.querySelector('.cselect-search'));
        if (spaceBelow < (hasSearch ? 300 : 220) && rect.top > (hasSearch ? 280 : 220)) {
          wrap.classList.add('cselect-up');
        } else {
          wrap.classList.remove('cselect-up');
        }
        wrap.classList.add('open');
        trigger.setAttribute('aria-expanded', 'true');
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
  if (!e.target.closest('.cselect-wrap')) {
    closeAllCustomSelects();
  }
  if (S.cmenu && !e.target.closest('.cmenu')) {
    S.cmenu = false;
    R();
  }
});

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
  if (globalSelectObserver) {
    globalSelectObserver.observe(document.body, { childList: true, subtree: true });
  }
});
