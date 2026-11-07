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
    eye: `<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>`
  };
  return `<svg width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="ic" aria-hidden="true">${p[name] || p.terminal}</svg>`;
}

// --- Selectors & Shortcuts ---
const $ = s => document.querySelector(s);
const $$ = s => document.querySelectorAll(s);
const esc = s => String(s == null ? '' : s).replace(/[&<>"']/g, c => ({
  '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'
}[c]));

// --- Global Application State ---
window.S = {
  token: localStorage.getItem('neopanel_token') || '',
  user: null,
  view: 'dash',              // active view: dash, con, set, act, fi, db, bk, sc, net, sub, stp, pl, plug, store, ver, all, admin, acc
  status: 'offline',         // online, offline, starting, stopping
  start: Date.now(),
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
  ptab: 'online',
  ftab: 'server.properties',
  currentDir: '',
  fileList: [],
  selectedFiles: new Set(),
  fileMenuOpen: null,
  adminTab: 0,
  adminUsers: [],
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
  cmenu: false,
  pq: '',
  pi: 0,
  vq: '',
  vf: 'All',
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

function up() {
  if (S.status !== 'online') return 'Not running';
  const s = Math.max(0, Math.floor((Date.now() - S.start) / 1000));
  return `${Math.floor(s / 3600)}h ${Math.floor((s % 3600) / 60)}m ${s % 60}s`;
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

function av(s) {
  const letter = (S.acct.fn ? S.acct.fn[0] : (S.acct.un ? S.acct.un[0] : 'A')).toUpperCase();
  return `<span class="av" style="width:${s}px;height:${s}px;font-size:${s * 0.4}px">
    ${esc(letter)}
    ${S.acct.h ? `<img src="https://gravatar.com/avatar/${S.acct.h}?s=${s * 2}&d=404" alt="" onerror="this.remove()">` : ''}
  </span>`;
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

// --- Advanced Minecraft & Terminal Log Parser ---
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

  // 2. Automatically detect if server is ready
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

  // 3. Escape HTML entities
  let text = esc(raw);

  // 4. Minecraft Section Color Codes (§0-§f, §l, §r)
  const mcColors = {
    '0': '#000000', '1': '#0000aa', '2': '#00aa00', '3': '#00aaaa',
    '4': '#aa0000', '5': '#aa00aa', '6': '#ffaa00', '7': '#aaaaaa',
    '8': '#555555', '9': '#5555ff', 'a': '#55ff55', 'b': '#55ffff',
    'c': '#ff5555', 'd': '#ff55ff', 'e': '#ffff55', 'f': '#ffffff'
  };
  text = text.replace(/§([0-9a-f])/gi, (_, c) => `</span><span style="color:${mcColors[c.toLowerCase()] || '#fff'}">`);
  text = text.replace(/§l/gi, '<b>').replace(/§r/gi, '</b></span>');

  // 5. ANSI Color Escape Sequences (compound codes like \x1b[0;33;1m, \x1b[m, etc.)
  const ansiMap = {
    30: '#4b5563', 31: '#ef4444', 32: '#10b981', 33: '#f59e0b',
    34: '#3b82f6', 35: '#ec4899', 36: '#06b6d4', 37: '#f3f4f6',
    90: '#6b7280', 91: '#f87171', 92: '#34d399', 93: '#fbbf24',
    94: '#60a5fa', 95: '#f472b6', 96: '#22d3ee', 97: '#ffffff'
  };
  text = text.replace(/\x1b\[([0-9;]*)m/g, (_, codeStr) => {
    if (!codeStr || codeStr === '0') return '</span>';
    const codes = codeStr.split(';').map(Number);
    let styles = [];
    for (const c of codes) {
      if (c === 0) styles = [];
      else if (c === 1) styles.push('font-weight:bold');
      else if (ansiMap[c]) styles.push(`color:${ansiMap[c]}`);
    }
    return styles.length ? `</span><span style="${styles.join(';')}">` : '</span>';
  });

  // 6. Highlight Log Levels & Timestamps (authentic to Sift theme)
  if (text.includes('WARN]:') || text.includes('WARN:') || text.includes('WARNING:')) {
    text = `<span class="w">${text}</span>`;
  } else if (text.includes('ERROR]:') || text.includes('ERROR:') || text.includes('Exception:') || text.includes('at java.') || text.includes('at net.minecraft.')) {
    text = `<span class="e">${text}</span>`;
  } else if (text.includes('INFO]:')) {
    text = text.replace(/\[([0-9:]+)\s+INFO\]:/g, '<span class="dim">[$1</span> <span class="c">INFO</span><span class="dim">]:</span>');
  } else if (text.includes('[NeoPanel]')) {
    text = text.replace(/\[NeoPanel\]/g, '<span class="c" style="font-weight:600;">[NeoPanel]</span>');
  }

  // 7. Highlight advancements & player activity
  if (text.includes('joined the game') || text.includes('has made the advancement')) {
    text = `<span class="c">${text}</span>`;
  } else if (text.includes('left the game')) {
    text = `<span class="w">${text}</span>`;
  }

  return text;
}

function updateInputState() {
  const inp = $('#cmd');
  if (!inp) return;
  const isOnline = S.status === 'online';
  inp.disabled = !isOnline;
  inp.placeholder = isOnline ? 'Type a command, try help...' : 'Server is offline';
}

function updateHeaderBadge() {
  const badge = $('#srv-badge');
  if (!badge) return;
  badge.className = `badge ${S.status}`;
  badge.innerHTML = `<i class="dot"></i>${{ online: 'Running', offline: 'Offline', starting: 'Starting', stopping: 'Stopping' }[S.status] || 'Offline'}`;
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
  if (S.view === 'con') paintLog();
}

function paintLog() {
  const l = $('#log');
  if (!l) return;
  l.innerHTML = S.log.join('\n');
  l.scrollTop = l.scrollHeight;
}

// --- WebSocket Connection ---
function connectWs(serverId) {
  if (!serverId) return;

  // If already connected or connecting to this server, do not recreate WebSocket
  if (S.activeWs && S.activeServerId === serverId && (S.activeWs.readyState === WebSocket.OPEN || S.activeWs.readyState === WebSocket.CONNECTING)) {
    return;
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

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${protocol}//${window.location.host}/api/servers/${serverId}/ws?token=${encodeURIComponent(S.token)}`;

  try {
    const ws = new WebSocket(wsUrl);
    S.activeWs = ws;

    ws.onmessage = e => {
      if (!e.data) return;
      try {
        const text = e.data.trim();
        if (text.startsWith('{')) {
          const msg = JSON.parse(text);
          if (msg.event === 'history') {
            try {
              const list = JSON.parse(msg.data);
              if (Array.isArray(list)) {
                S.log = list.map(formatLogLine).filter(Boolean);
                if (S.log.length > 800) S.log = S.log.slice(-800);
                if (S.view === 'con') paintLog();
                return;
              }
            } catch (histErr) {}
          } else if (msg.event === 'log') {
            addLog(msg.data);
            return;
          } else if (msg.event === 'stats' || msg.type === 'stats') {
            const c = Math.max(0, Math.min(100, msg.cpu_percent || 0));
            S.cpu.push(c);
            S.cpu.shift();

            const mem = msg.memory_bytes || 0;
            const maxMemMB = getServerMaxRamMB(curServer());
            const maxMemBytes = maxMemMB * 1024 * 1024;
            const r = Math.max(0, Math.min(100, (mem / maxMemBytes) * 100));
            S.ram.push(r);
            S.ram.shift();

            S.ni.push(Math.max(2, Math.min(480, S.ni.at(-1) + (Math.random() - 0.45) * 40)));
            S.ni.shift();
            S.no.push(Math.max(2, Math.min(480, S.no.at(-1) + (Math.random() - 0.45) * 30)));
            S.no.shift();

            liveCon();
            return;
          } else if (msg.event === 'status' || msg.type === 'status') {
            S.status = msg.data || S.status;
            updateInputState();
            updateHeaderBadge();
            return;
          }
        }
      } catch (err) {}

      addLog(e.data);
    };

    ws.onclose = () => {
      if (S.activeWs === ws) {
        S.activeWs = null;
      }
    };
  } catch (e) {}
}

// --- Power Actions ---
async function power(action) {
  const srv = curServer();
  if (!srv) return;

  if (action === 'start') {
    S.status = 'starting';
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
    toast('Restarting server...');
    addLog('[NeoPanel]: Restart sequence initiated...');
  } else if (action === 'kill') {
    toast('Terminating server process...');
    addLog('[NeoPanel]: Force kill signal delivered');
  }

  try {
    const res = await api(`/api/servers/${srv.id}/power`, {
      method: 'POST',
      body: JSON.stringify({ action })
    });
    S.status = res.status || (action === 'start' ? 'online' : 'offline');
    if (S.status === 'online') S.start = Date.now();
    updateInputState();
    updateHeaderBadge();
    toast(`Server ${action} command sent`);
  } catch (err) {
    toast(`Action failed: ${err.message}`, 'error');
    refreshServerStatus();
  }
}

async function refreshServerStatus() {
  const srv = curServer();
  if (!srv) return;
  try {
    const data = await api(`/api/servers/${srv.id}`);
    S.status = data.stats?.status || data.status || 'offline';
    S.ver = data.version || S.ver;
    S.core = (data.type || 'paper').charAt(0).toUpperCase() + (data.type || 'paper').slice(1);
    updateInputState();
    updateHeaderBadge();
    R();
  } catch (e) {}
}

async function runCmd(cmd) {
  if (!cmd) return;
  addLog(`<span class="c">&gt; ${esc(cmd)}</span>`);
  const srv = curServer();
  if (!srv) return;

  if (S.activeWs && S.activeWs.readyState === WebSocket.OPEN) {
    S.activeWs.send(JSON.stringify({ event: 'command', data: cmd }));
  } else {
    try {
      await api(`/api/servers/${srv.id}/command`, {
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
  ['plug', 'Plugins', 'puzzle', 'Minecraft'],
  ['store', 'Plugin store', 'store', 'Minecraft'],
  ['ver', 'Software', 'cpu', 'Minecraft'],
  ['all', 'All versions', 'layers', 'Minecraft']
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
  ['Fabric', 'Lightweight modular mod loader'],
  ['Vanilla', 'Standard Mojang vanilla server'],
  ['NeoForge', 'Modern forge modding engine'],
  ['Velocity', 'Modern high-throughput proxy'],
  ['Waterfall', 'PaperMC BungeeCord proxy fork'],
  ['BungeeCord', 'Classic multi-server proxy']
];

const FAM = [
  ['26.3', 'Wilderness Bound', '2026', ['26.3']],
  ['26.2', 'Chaos Cubed', '2026', ['26.2']],
  ['26.1', 'Tiny Takeover', '2026', ['26.1', '26.1.1', '26.1.2']],
  ['1.21', 'Tricky Trials & drops', '2024 to 2025', ['1.21', '1.21.1', '1.21.2', '1.21.3', '1.21.4']],
  ['1.20', 'Trails and Tales', '2023', ['1.20', '1.20.1', '1.20.2', '1.20.4', '1.20.6']],
  ['1.19', 'The Wild Update', '2022', ['1.19', '1.19.2', '1.19.4']],
  ['1.18', 'Caves and Cliffs, part 2', '2021', ['1.18', '1.18.1', '1.18.2']],
  ['1.17', 'Caves and Cliffs, part 1', '2021', ['1.17', '1.17.1']],
  ['1.16', 'Nether Update', '2020', ['1.16.5', '1.16.4', '1.16.1']],
  ['1.12', 'World of Color', '2017', ['1.12.2', '1.12.1']]
];

const ALL_VERS = FAM.flatMap(f => f[3].map(v => ({ v, f: f[0], n: f[1], y: f[2] })));
const CATS = ['All', 'Admin', 'Protection', 'Performance', 'Network', 'Building', 'Utility', 'API', 'Map'];

// --- Header Banner Component (.srvh) ---
function srvh() {
  const srv = curServer();
  const st = S.status;
  const on = st === 'online';
  const c = on ? S.cpu.at(-1) : 0;
  const maxRamMB = getServerMaxRamMB(srv);
  const m = on ? (S.ram.at(-1) / 100) * maxRamMB : 0;
  const hostAddr = `${window.location.hostname}:${srv ? srv.port : 25565}`;

  return `<div class="card srvh">
    <div>
      <div class="row">
        <h1>${esc(S.props.motd || srv?.name || 'Minecraft Server')}</h1>
        <span id="srv-badge" class="badge ${st}"><i class="dot"></i>${{ online: 'Running', offline: 'Offline', starting: 'Starting', stopping: 'Stopping' }[st]}</span>
      </div>
      <div id="hst" class="row hst">
        <span class="mono mu pv">${hostAddr}</span>
        <span class="mono mu">CPU ${c.toFixed(0)}%</span>
        <span class="mono mu">RAM ${m.toFixed(0)} MiB / ${(maxRamMB / 1024).toFixed(1)} GiB</span>
      </div>
    </div>
    <div class="row" style="position:relative">
      <button class="btn" onclick="toggleCmdMenu()" aria-haspopup="menu" aria-expanded="${!!S.cmenu}">Commands &#9662;</button>
      ${S.cmenu ? `<div class="cmenu" role="menu">
        ${CMDS.map(([label, cmd]) => `<button role="menuitem" onclick="execQuickCmd('${cmd}')">${label}</button>`).join('')}
        <div style="border-top: 1px solid var(--border); margin: 4px 0;"></div>
        <button role="menuitem" style="color: #ef4444;" onclick="S.cmenu=false; openDeleteServerModal('${srv?.id}', '${esc(srv?.name || '')}');">${ic('trash', 13)} Delete Server</button>
      </div>` : ''}
      <button class="btn pr" onclick="power('start')" ${st !== 'offline' ? 'disabled' : ''}>${ic('play', 14)} Start</button>
      <button class="btn" onclick="power('restart')" ${on ? '' : 'disabled'}>${ic('refresh', 14)} Restart</button>
      <button class="btn stop" onclick="power('stop')" ${on ? '' : 'disabled'}>Stop</button>
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
  const on = S.status === 'online';
  const c = on ? S.cpu.at(-1) : 0;
  const maxRamMB = getServerMaxRamMB(srv);
  const m = on ? (S.ram.at(-1) / 100) * maxRamMB : 0;
  const hostAddr = `${window.location.hostname}:${srv ? srv.port : 25565}`;

  const row = (i, l, v, s) => `<div class="srow">
    <span class="sic" aria-hidden="true">${i}</span>
    <div>
      <div class="mu">${l}</div>
      <b class="mono">${v}</b>${s ? ` <span class="mu mono">${s}</span>` : ''}
    </div>
  </div>`;

  return row(ic('network', 16), 'Address', `<span class="pv">${hostAddr}</span>`) +
    row(ic('clock', 16), 'Uptime', up()) +
    row(ic('cpu', 16), 'CPU load', `${c.toFixed(1)}%`, '/ 100%') +
    row(ic('layers', 16), 'Memory', `${m.toFixed(0)} MiB`, `/ ${(maxRamMB / 1024).toFixed(1)} GiB`) +
    row(ic('database', 16), 'Disk', '1.82 GiB', '/ 20 GiB') +
    row('↓', 'Network (inbound)', `${(on ? S.ni.at(-1) : 0).toFixed(1)} KiB`) +
    row('↑', 'Network (outbound)', `${(on ? S.no.at(-1) : 0).toFixed(1)} KiB`);
}

function chartsHtml() {
  const srv = curServer();
  const on = S.status === 'online';
  const maxRamGB = (getServerMaxRamMB(srv) / 1024).toFixed(1);
  const z = a => a.map(() => 0);
  return `<div class="card ch">
    <div class="row sb"><span class="mu">CPU load</span><span class="mu mono">100%</span></div>
    ${spark(on ? S.cpu : z(S.cpu), 100)}
  </div>
  <div class="card ch">
    <div class="row sb"><span class="mu">Memory</span><span class="mu mono">${maxRamGB} GiB</span></div>
    ${spark(on ? S.ram : z(S.ram), 100)}
  </div>
  <div class="card ch">
    <div class="row sb"><span class="mu">Network I/O</span><span class="mu"><span style="color:var(--ac)">in</span> <span style="color:#f0b429">out</span></span></div>
    ${lines(on ? [S.ni, S.no] : [z(S.ni), z(S.no)], 500)}
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
    const on = S.status === 'online';
    const c = on ? S.cpu.at(-1) : 0;
    const maxRamMB = getServerMaxRamMB(srv);
    const m = on ? (S.ram.at(-1) / 100) * maxRamMB : 0;
    const hostAddr = `${window.location.hostname}:${srv ? srv.port : 25565}`;
    hst.innerHTML = `<span class="mono mu pv">${hostAddr}</span><span class="mono mu">CPU ${c.toFixed(0)}%</span><span class="mono mu">RAM ${m.toFixed(0)} MiB / ${(maxRamMB / 1024).toFixed(1)} GiB</span>`;
  }
}

// --- Views Implementation (V) ---
const V = {
  dash() {
    const srv = curServer();
    const runningCount = S.servers.filter(s => (s.stats?.status || s.status) === 'running' || (s.stats?.status || s.status) === 'online').length;
    const totalCount = S.servers.length;
    const totalRamAllocatedMB = S.servers.reduce((sum, s) => sum + getServerMaxRamMB(s), 0);
    const totalRamAllocatedGB = (totalRamAllocatedMB / 1024).toFixed(1);
    const totalRamUsedGB = S.status === 'online' ? ((S.ram.at(-1) / 100) * (totalRamAllocatedMB / 1024)).toFixed(1) : '0.0';

    return `
    <div class="row sb" style="margin-bottom: 24px; align-items: flex-end; flex-wrap: wrap; gap: 16px;">
      <div>
        <h1 style="font-size: 26px; font-weight: 800; color: #fff; margin-bottom: 4px; letter-spacing: -0.02em;">Servers</h1>
        <span class="mu" style="font-size: 13px; font-weight: 500;">${runningCount} &bull; ${totalCount} running</span>
      </div>
      <div class="row" style="gap: 10px; flex-wrap: wrap;">
        <div style="position: relative;">
          <input type="text" id="srv-filter" placeholder="Filter by name or address" oninput="filterServerCards(this.value)" style="width: 220px; padding-left: 32px; font-size: 13px; height: 36px; border-radius: 8px;">
          <span style="position: absolute; left: 10px; top: 9px; color: var(--tx-muted);">${ic('search', 14)}</span>
        </div>
        <div class="row" style="gap: 2px; background: rgba(255,255,255,0.04); padding: 3px; border-radius: 8px; border: 1px solid var(--border);">
          <button class="btn sm pr" style="height: 28px; padding: 0 12px;">Mine</button>
          <button class="btn sm" style="height: 28px; padding: 0 12px; border: 0; background: transparent;">Others</button>
        </div>
        <button class="btn pr" onclick="openCreateServerModal()" style="height: 36px; padding: 0 16px; font-size: 13px;">
          ${ic('plus', 14)} Create server
        </button>
      </div>
    </div>

    <div style="display: grid; grid-template-columns: 1fr 310px; gap: 20px; align-items: start;">
      <!-- Left Column: Server Cards Grid -->
      <div id="srv-card-grid" style="display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 16px;">
        ${S.servers.map((s, i) => {
          const isRun = (s.stats?.status || s.status) === 'running' || (s.stats?.status || s.status) === 'online' || (i === S.cs && S.status === 'online');
          const maxRam = getServerMaxRamMB(s);
          const cVal = isRun ? (i === S.cs ? S.cpu.at(-1) : 2) : 0;
          const mVal = isRun ? (i === S.cs ? ((S.ram.at(-1) / 100) * maxRam).toFixed(0) : Math.round(maxRam * 0.4)) : 0;
          const host = `${window.location.hostname}:${s.port || 25565}`;
          const coreLabel = (s.type || 'paper').charAt(0).toUpperCase() + (s.type || 'paper').slice(1);

          return `
          <div class="sift-server-card" onclick="switchServer(${i}); go('con');" data-name="${esc(s.name)}" data-addr="${host}">
            <div class="srv-card-banner" style="background-image: url('/img/banner.jpg');">
              <span class="srv-card-badge">Minecraft - ${esc(coreLabel)}</span>
              <div class="srv-card-actions">
                <span class="srv-card-action-btn" onclick="event.stopPropagation(); refreshServerStatus();" title="Refresh status">↻</span>
                <span class="srv-card-action-btn danger" onclick="event.stopPropagation(); openDeleteServerModal('${s.id}', '${esc(s.name)}');" title="Delete Server">${ic('trash', 13)}</span>
              </div>
            </div>
            <div class="srv-card-body">
              <div class="srv-card-title-row">
                <span class="srv-card-title">${esc(s.name || 'Minecraft Server')}</span>
                <span class="badge ${isRun ? 'online' : 'offline'}">
                  <i class="dot"></i>${isRun ? 'Running' : 'Offline'}
                  ${isRun ? `<span class="mu" style="font-weight: 400; margin-left: 3px;">&bull; Up ${up()}</span>` : ''}
                </span>
              </div>
              <div class="srv-card-addr">${host}</div>
              <div class="srv-meters-row">
                <div class="srv-meter">
                  <div class="srv-meter-head">
                    <span>CPU</span>
                    <span class="srv-meter-val">${cVal.toFixed(0)}%</span>
                  </div>
                  <div class="meter-bar"><i style="width:${Math.min(100, cVal)}%"></i></div>
                </div>
                <div class="srv-meter">
                  <div class="srv-meter-head">
                    <span>Memory</span>
                    <span class="srv-meter-val">${mVal} MB</span>
                  </div>
                  <div class="meter-bar"><i style="width:${Math.min(100, (mVal / (maxRam || 1)) * 100)}%"></i></div>
                </div>
                <div class="srv-meter">
                  <div class="srv-meter-head">
                    <span>Disk</span>
                    <span class="srv-meter-val">222 MB</span>
                  </div>
                  <div class="meter-bar"><i style="width:14%"></i></div>
                </div>
              </div>
            </div>
          </div>`;
        }).join('')}

        <!-- Dashed Create Server Card -->
        <div class="create-server-dashed-card" onclick="openCreateServerModal()">
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
            <h2 style="font-size: 15px; font-weight: 700; color: #fff;">Recent activity</h2>
            <a href="#act" onclick="go('act')" class="mu" style="font-size: 12px; text-decoration: none;">View all</a>
          </div>
          <div style="display: flex; flex-direction: column; gap: 12px;">
            <div style="display: flex; align-items: flex-start; justify-content: space-between; font-size: 13px;">
              <div style="display: flex; gap: 8px;">
                <span class="mu" style="color: var(--tx-dim);">&rarr;</span>
                <div>
                  <div style="font-weight: 600; color: #fff;">Signed in</div>
                  <div class="mu mono" style="font-size: 11.5px;">${esc(S.acct.un || 'admin')} &bull; ${window.location.hostname}</div>
                </div>
              </div>
              <span class="mu" style="font-size: 12px;">12m</span>
            </div>
            ${S.servers.length ? `
              <div style="display: flex; align-items: flex-start; justify-content: space-between; font-size: 13px;">
                <div style="display: flex; gap: 8px;">
                  <span class="mu" style="color: var(--tx-dim);">&rarr;</span>
                  <div>
                    <div style="font-weight: 600; color: #fff;">Server registered</div>
                    <div class="mu" style="font-size: 11.5px;">${esc(S.servers[0].name)} &bull; ${esc(S.servers[0].type || 'paper')}</div>
                  </div>
                </div>
                <span class="mu" style="font-size: 12px;">2h</span>
              </div>
            ` : ''}
          </div>
        </div>

        <!-- Your Allocation Card -->
        <div class="card" style="padding: 18px;">
          <h2 style="font-size: 15px; font-weight: 700; color: #fff; margin-bottom: 14px;">Your allocation</h2>
          <div style="display: flex; flex-direction: column; gap: 14px;">
            <div>
              <div class="row sb" style="font-size: 12px; margin-bottom: 5px;">
                <span class="mu">Memory</span>
                <span class="mono" style="font-weight: 600; color: #fff;">${totalRamUsedGB} of ${totalRamAllocatedGB} GB</span>
              </div>
              <div class="meter-bar"><i style="width: ${Math.min(100, (parseFloat(totalRamUsedGB) / (parseFloat(totalRamAllocatedGB) || 1)) * 100)}%;"></i></div>
            </div>
            <div>
              <div class="row sb" style="font-size: 12px; margin-bottom: 5px;">
                <span class="mu">Disk</span>
                <span class="mono" style="font-weight: 600; color: #fff;">0.2 of 10 GB</span>
              </div>
              <div class="meter-bar"><i style="width: 12%;"></i></div>
            </div>
            <div>
              <div class="row sb" style="font-size: 12px; margin-bottom: 5px;">
                <span class="mu">Running</span>
                <span class="mono" style="font-weight: 600; color: #fff;">${runningCount} of ${totalCount}</span>
              </div>
              <div class="meter-bar"><i style="width: ${Math.min(100, (runningCount / (totalCount || 1)) * 100)}%;"></i></div>
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
          <div class="row" style="gap:6px;">
            ${ic('terminal', 14)}
            <span>Console Stream &bull; Buffer: ${S.log.length} lines</span>
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
    let body = '';

    if (t === 'online') {
      body = `<table>
        <thead><tr><th>Player</th><th>Ping</th><th>Playtime</th><th>IP</th><th style="text-align:right">Actions</th></tr></thead>
        <tbody>
          ${S.players.length ? S.players.map((p, i) => `
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
          `).join('') : '<tr><td colspan="5" class="mu" style="text-align:center;padding:24px;">No players currently online. Start the server to accept connections.</td></tr>'}
        </tbody>
      </table>`;
    } else if (t === 'wl') {
      body = `<div class="row" style="margin-bottom:14px">
        <input type="text" id="wn" placeholder="Add Minecraft username to whitelist" style="max-width:280px">
        <button class="btn pr" onclick="addWhitelistPlayer()">Add to whitelist</button>
      </div>
      <table>
        <thead><tr><th>Player</th><th>UUID</th><th style="text-align:right">Actions</th></tr></thead>
        <tbody>
          ${S.wl.length ? S.wl.map(p => {
            const name = typeof p === 'string' ? p : p.name;
            const uuid = typeof p === 'object' ? p.uuid : '';
            return `<tr>
              <td>${head(name)}<b>${esc(name)}</b></td>
              <td class="mono mu">${uuid || '—'}</td>
              <td style="text-align:right">
                <button class="btn sm dg" onclick="playerAction('whitelist_remove', '${esc(name)}')">Remove</button>
              </td>
            </tr>`;
          }).join('') : '<tr><td colspan="3" class="mu" style="text-align:center;padding:20px;">No players on whitelist.</td></tr>'}
        </tbody>
      </table>`;
    } else if (t === 'ban') {
      body = `<table>
        <thead><tr><th>Player</th><th>Reason</th><th style="text-align:right">Actions</th></tr></thead>
        <tbody>
          ${S.bans.length ? S.bans.map(b => `
            <tr>
              <td>${head(b.name)}<b>${esc(b.name)}</b></td>
              <td class="mu">${esc(b.reason || 'Banned by operator')}</td>
              <td style="text-align:right">
                <button class="btn sm" onclick="playerAction('pardon', '${esc(b.name)}')">Pardon</button>
              </td>
            </tr>
          `).join('') : '<tr><td colspan="3" class="mu" style="text-align:center;padding:20px;">No banned players on record.</td></tr>'}
        </tbody>
      </table>`;
    }

    return `${srvh()}
    <div class="tabs" style="margin-top:14px">
      ${[['online', 'Online Players'], ['wl', 'Whitelist'], ['ban', 'Banned Players']].map(([k, l]) => `
        <button class="btn ${t === k ? 'on' : ''}" onclick="S.ptab='${k}'; R(); loadPlayersData();">${l}</button>
      `).join('')}
    </div>
    <div class="card tw">${body}</div>`;
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

    return `${srvh()}
    <div style="margin-top: 16px;">
      <div class="row sb" style="margin-bottom: 16px; flex-wrap: wrap; gap: 12px;">
        <!-- Breadcrumb Navigation -->
        <div class="breadcrumb" style="font-size: 13.5px; font-weight: 500;">
          <span onclick="browseFolder('')" style="color: var(--tx-muted); cursor: pointer;">/ home / container /</span>
          ${pathParts.map((part, idx) => {
            const sub = pathParts.slice(0, idx + 1).join('/');
            const isLast = idx === pathParts.length - 1;
            return `<span onclick="browseFolder('${sub}')" style="cursor: pointer; color: ${isLast ? '#fff' : 'var(--tx-muted)'}; margin-left: 4px;">${esc(part)} /</span>`;
          }).join('')}
        </div>

        <!-- Search & Action Buttons -->
        <div class="row" style="gap: 10px; flex-wrap: wrap;">
          <div style="position: relative;">
            <input type="text" id="fi-filter" placeholder="Filter this folder" oninput="filterFileList(this.value)" style="width: 200px; height: 36px; padding: 0 28px 0 12px; font-size: 13px; border-radius: 8px;">
            <kbd style="position: absolute; right: 10px; top: 9px; font-size: 11px; color: var(--tx-muted); background: rgba(255,255,255,0.06); padding: 1px 5px; border-radius: 4px;">/</kbd>
          </div>
          <button class="btn" onclick="promptNewFolder()" style="height: 36px;">+ New folder</button>
          <button class="btn" onclick="uploadFileDialog()" style="height: 36px;">${ic('download', 14)} Upload</button>
          <button class="btn pr" onclick="promptNewFile()" style="height: 36px;">+ New file</button>
        </div>
      </div>

      <!-- Full-Width File Table -->
      <div class="card tw" style="padding: 0; overflow: visible;">
        <table class="file-table" style="width: 100%;">
          <thead>
            <tr style="border-bottom: 1px solid var(--ln); color: var(--tx-muted); font-size: 12px; text-transform: uppercase; letter-spacing: 0.5px;">
              <th style="width: 36px; padding: 12px 0 12px 18px; text-align: left;">
                <input type="checkbox" id="chk-all-files" onchange="toggleSelectAllFiles(this.checked)" style="cursor: pointer; width: 15px; height: 15px; accent-color: var(--ac);" ${allSelected ? 'checked' : ''}>
              </th>
              <th style="width: 32px; padding: 12px 4px;"></th>
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
                <td style="color: var(--ac); font-size: 16px;">📁</td>
                <td colspan="4" style="font-weight: 600; color: #fff;">.. (Parent folder)</td>
              </tr>
            ` : ''}
            ${S.fileList.map((item, idx) => {
              const isDir = item.is_dir;
              const icon = isDir ? '📁' : '📄';
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
                         style="cursor: pointer; width: 15px; height: 15px; accent-color: var(--ac);">
                </td>
                <td style="width: 32px; padding: 0 4px; font-size: 16px;">${icon}</td>
                <td style="font-weight: 500; color: #fff;">${esc(item.name)}</td>
                <td class="mu mono" style="font-size: 12.5px; width: 140px;">${sizeStr}</td>
                <td class="mu mono" style="font-size: 12.5px; width: 180px;">${dateStr}</td>
                <td style="width: 50px; text-align: right; padding-right: 18px; position: relative;" onclick="event.stopPropagation();">
                  <button class="btn sm" onclick="toggleFileCtx(event, ${idx})" style="padding: 4px 8px; border: 0; background: transparent; font-size: 18px; line-height: 1; color: var(--tx-muted);">&#8942;</button>
                  <div class="file-ctx-menu" id="fctx-${idx}" style="display: none;">
                    ${!isDir ? `<button onclick="openFile('${esc(item.name)}'); closeAllCtx();">${ic('edit', 14)} Edit</button>` : ''}
                    ${!isDir ? `<button onclick="downloadFileItem('${esc(itemPath)}'); closeAllCtx();">${ic('download', 14)} Download</button>` : ''}
                    <button onclick="promptRenameFile('${esc(itemPath)}'); closeAllCtx();">${ic('settings', 14)} Rename</button>
                    <button onclick="copyFilePath('${esc(itemPath)}'); closeAllCtx();">📋 Copy Path</button>
                    <button class="dg" onclick="deleteFileItem('${esc(itemPath)}'); closeAllCtx();">${ic('trash', 14)} Delete</button>
                  </div>
                </td>
              </tr>`;
            }).join('')}
            ${!S.fileList.length && !isSubdir ? '<tr><td colspan="6" class="mu" style="text-align:center; padding: 32px;">This folder is empty. Drag and drop files here or click Upload.</td></tr>' : ''}
          </tbody>
        </table>
      </div>

      <!-- Floating Bulk Action Bar -->
      ${S.selectedFiles.size > 0 ? `
        <div class="bulk-action-bar">
          <span style="font-weight: 600; font-size: 13.5px; color: #fff;">${S.selectedFiles.size} item${S.selectedFiles.size > 1 ? 's' : ''} selected</span>
          <button class="btn sm dg" onclick="deleteSelectedFiles()">${ic('trash', 14)} Delete Selected</button>
          <button class="btn sm" onclick="deselectAllFiles()">Deselect</button>
        </div>
      ` : ''}
    </div>`;
  },

  plug() {
    const isBrowse = (S.plugTab || 'browse') === 'browse';
    return `${srvh()}
    <div style="margin-top: 18px;">
      <!-- Top Title and Controls Bar -->
      <div class="row sb" style="margin-bottom: 16px; flex-wrap: wrap; gap: 14px;">
        <h1 style="font-size: 24px; font-weight: 800; color: #fff;">Plugins</h1>
        <div class="row" style="gap: 10px; flex-wrap: wrap;">
          <button class="btn ${isBrowse ? 'pr' : ''} sm" onclick="setPlugTab('browse')">Browse</button>
          <button class="btn ${!isBrowse ? 'pr' : ''} sm" onclick="setPlugTab('installed')">Installed ${S.plugins.length}</button>
          
          <select style="width: auto; height: 32px; padding: 2px 10px; font-size: 12px;" disabled>
            <option>Source: Modrinth</option>
          </select>
          <select style="width: auto; height: 32px; padding: 2px 10px; font-size: 12px;" onchange="S.core=this.value; searchModrinth();">
            <option ${S.core === 'Paper' ? 'selected' : ''}>Software: Paper</option>
            <option ${S.core === 'Purpur' ? 'selected' : ''}>Software: Purpur</option>
            <option ${S.core === 'Fabric' ? 'selected' : ''}>Software: Fabric</option>
            <option ${S.core === 'Spigot' ? 'selected' : ''}>Software: Spigot</option>
            <option ${S.core === 'Velocity' ? 'selected' : ''}>Software: Velocity</option>
          </select>
          <select style="width: auto; height: 32px; padding: 2px 10px; font-size: 12px;">
            <option>Version: Any</option>
            <option>1.21.4</option>
            <option>1.20.4</option>
            <option>1.19.4</option>
          </select>
          <select style="width: auto; height: 32px; padding: 2px 10px; font-size: 12px;">
            <option>Sort: Relevance</option>
            <option>Sort: Downloads</option>
            <option>Sort: Follows</option>
            <option>Sort: Newest</option>
          </select>
        </div>
      </div>

      <!-- Big Search Input -->
      <div style="position: relative; margin-bottom: 18px;">
        <input type="text" id="store-search" placeholder="Search 19,447 plugins on Modrinth..." value="${esc(S.q || '')}" oninput="S.q=this.value; searchModrinthDebounced();" style="height: 42px; font-size: 14px; padding-left: 38px; border-radius: 10px;">
        <span style="position: absolute; left: 12px; top: 12px; color: var(--tx-muted);">${ic('search', 16)}</span>
      </div>

      <!-- Content Grid: Browse Modrinth Hits or Installed Plugins -->
      ${isBrowse ? `
        <div style="display: flex; flex-direction: column; gap: 10px;" id="store-grid">
          <div class="mu" style="padding: 24px; text-align: center;">Loading top plugins from Modrinth...</div>
        </div>
      ` : `
        <div style="display: flex; flex-direction: column; gap: 10px;">
          ${S.plugins.map((p, i) => `
            <div class="plugin-card">
              <div style="display: flex; align-items: center; gap: 16px; min-width: 0;">
                <div style="width: 44px; height: 44px; border-radius: 8px; background: rgba(56, 122, 255, 0.12); display: grid; place-items: center; color: var(--ac); font-size: 20px;">
                  🧩
                </div>
                <div>
                  <b class="plugin-title">${esc(p[0])}</b>
                  <div class="mu mono" style="font-size: 12px; margin-top: 3px;">${esc(p[1] || 'installed')}</div>
                </div>
              </div>
              <div class="row" style="gap: 12px;">
                <label class="sw">
                  <input type="checkbox" ${p[2] ? 'checked' : ''} onchange="togglePlugin(${i}, this.checked)">
                  <i></i>
                </label>
                <button class="btn sm dg" onclick="deletePlugin('${esc(p[0])}')">Uninstall</button>
              </div>
            </div>
          `).join('') || '<div class="mu" style="padding: 32px; text-align: center;">No plugins installed. Switch to Browse to install plugins with one click.</div>'}
        </div>
      `}
    </div>`;
  },

  store() {
    S.plugTab = 'browse';
    return this.plug();
  },

  ver() {
    return `${srvh()}
    <div class="card hero" style="margin-top:14px">
      <div class="row sb" style="align-items: center; flex-wrap: wrap; gap: 12px;">
        <div>
          <span class="mu">Currently Running</span>
          <b class="big">${S.core} ${S.ver}</b>
          <span class="mu">Java 21 / 17 compatible &bull; High performance runtime</span>
        </div>
        <button class="btn pr" onclick="openReinstallModal('${curServer()?.id}')">
          ${ic('refresh', 14)} Reinstall Server (mcjarfiles.com)
        </button>
      </div>
    </div>

    <h2 style="margin-top:20px">Choose Server Engine</h2>
    <div class="grid g4">
      ${CORES.map(([name, desc]) => `
        <button class="card sel ${S.sel === name ? 'on' : ''}" onclick="S.sel='${name}'; R();">
          <b>${name}</b>
          <span class="mu" style="font-size:12px;">${desc}</span>
        </button>
      `).join('')}
    </div>

    <div class="row sb" style="margin:22px 0 12px">
      <div class="row">
        <h2 style="margin:0">Available Releases for ${S.sel}</h2>
        <button class="btn sm" onclick="go('all')">See all version matrix</button>
      </div>
      <label class="row" style="cursor:pointer">
        <input type="checkbox" id="bkf" checked style="width:auto; margin-right:6px;"> Back up the world first
      </label>
    </div>

    <div class="card tw">
      <table>
        <thead><tr><th>Version</th><th>Notes</th><th style="text-align:right">Action</th></tr></thead>
        <tbody>
          ${(() => {
            let releases = [];
            if (S.sel === 'Velocity') {
              releases = [
                ['3.4.0-SNAPSHOT', 'Popular latest (Java 17-21)'],
                ['3.3.0-SNAPSHOT', 'Stable proxy release']
              ];
            } else if (S.sel === 'Waterfall') {
              releases = [
                ['1.20', 'Final Waterfall release (Build 578)'],
                ['1.19', 'Legacy stable proxy']
              ];
            } else if (S.sel === 'BungeeCord') {
              releases = [
                ['latest', 'Continuous Jenkins CI build']
              ];
            } else {
              releases = [
                ['1.21.4', 'Popular latest'],
                ['1.21.1', 'Widely supported'],
                ['1.20.4', 'Stable'],
                ['1.20.1', 'Top modded'],
                ['1.19.4', 'Legacy stable']
              ];
            }
            return releases.map(([v, n]) => {
              const cur = S.ver === v && S.core === S.sel;
              return `<tr>
                <td class="mono"><b>${v}</b></td>
                <td><span class="chip">${n}</span></td>
                <td style="text-align:right">
                  ${cur ? '<span class="badge online"><i class="dot"></i>Active</span>' : `<button class="btn sm pr" onclick="installSoftware('${S.sel}', '${v}')">Install ${S.sel} ${v}</button>`}
                </td>
              </tr>`;
            }).join('');
          })()}
        </tbody>
      </table>
    </div>`;
  },

  all() {
    return `${srvh()}
    <div class="card hero" style="margin-top:14px">
      <span class="mu">Running now</span>
      <b class="big">${S.core} ${S.ver}</b>
      <span class="mu">Direct installer for all official Minecraft versions (from modern drops down to legacy releases).</span>
    </div>

    <div class="row" style="margin:18px 0 12px; gap:10px;">
      <input type="text" placeholder="Search versions..." value="${esc(S.vq)}" oninput="S.vq=this.value; renderVersionRows();" style="max-width:240px">
      <select onchange="S.vf=this.value; renderVersionRows();" style="width:auto">
        <option>All Families</option>
        ${FAM.map(f => `<option ${S.vf === f[0] ? 'selected' : ''}>${f[0]} (${f[1]})</option>`).join('')}
      </select>
      <select onchange="S.sel=this.value; renderVersionRows();" style="width:auto">
        ${CORES.map(([c]) => `<option ${S.sel === c ? 'selected' : ''}>${c}</option>`).join('')}
      </select>
      <label class="row" style="margin-left:auto; cursor:pointer;">
        <input type="checkbox" id="bkf" checked style="width:auto; margin-right:6px;"> Back up world first
      </label>
    </div>

    <div class="card tw">
      <table>
        <thead><tr><th>Version</th><th>Release Family</th><th>Java</th><th>Engines</th><th style="text-align:right">Action</th></tr></thead>
        <tbody id="vres">${versionRowsHtml()}</tbody>
      </table>
    </div>`;
  },

  bk() {
    return `${srvh()}
    <div class="row sb" style="margin:14px 0 12px">
      <span class="mu">${S.bk.length} backups recorded</span>
      <button class="btn pr" onclick="createBackup()">${ic('plus', 14)} Create Full Backup</button>
    </div>
    <div class="card tw">
      <table>
        <thead><tr><th>Archive Name</th><th>Size</th><th>Created</th><th style="text-align:right">Actions</th></tr></thead>
        <tbody>
          ${S.bk.length ? S.bk.map((b, i) => `
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
          `).join('') : '<tr><td colspan="4" class="mu" style="text-align:center;padding:24px;">No backups yet. Create your first archive to protect your world!</td></tr>'}
        </tbody>
      </table>
    </div>`;
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
    const host = window.location.hostname;
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
          <button class="btn sm" onclick="navigator.clipboard.writeText('sftp://${sftpUser}@${host}:${sftpPort}'); toast('SFTP URI copied!');">
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
    return `${srvh()}
    <div class="card tw" style="margin-top:14px">
      <table>
        <thead><tr><th>Database</th><th>Username</th><th>Connections</th><th>Size</th></tr></thead>
        <tbody>
          <tr><td class="mono">mc_survival</td><td class="mono">u_survival</td><td>2</td><td>48 MiB</td></tr>
          <tr><td class="mono">mc_coreprotect</td><td class="mono">u_cp</td><td>1</td><td>212 MiB</td></tr>
        </tbody>
      </table>
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
    return `${srvh()}
    <div style="margin-top: 18px;">
      <!-- Top Grid: Startup Command & Docker Image / Java -->
      <div style="display: grid; grid-template-columns: 1fr 340px; gap: 16px; margin-bottom: 24px;">
        <div class="card" style="padding: 20px;">
          <h2 style="font-size: 15px; font-weight: 700; color: #fff; margin-bottom: 12px;">Startup Command</h2>
          <div class="mono" style="background: var(--con-bg); padding: 14px 16px; border-radius: 8px; border: 1px solid var(--border); color: #fff; font-size: 12.5px; line-height: 1.6; word-break: break-all;">
            java -Xms128M -XX:MaxRAMPercentage=95.0 -Dterminal.jline=false -Dterminal.ansi=true -jar ${esc(srv?.jar_file || 'server.jar')}
          </div>
        </div>

        <div class="card" style="padding: 20px;">
          <h2 style="font-size: 15px; font-weight: 700; color: #fff; margin-bottom: 12px;">Docker Image / Java Runtime</h2>
          <select id="stp-java" style="margin-bottom: 8px;">
            <option selected>Java 21 (Recommended)</option>
            <option>Java 17 (Legacy 1.18 - 1.20.4)</option>
            <option>Java 8 (Legacy 1.12.2)</option>
            <option>Java 25 (Cutting edge)</option>
          </select>
          <span class="mu" style="font-size: 12px; line-height: 1.5; display: block;">This is an advanced feature allowing you to select a Docker image or Java runtime to use when running this server instance.</span>
        </div>
      </div>

      <!-- Variables Section -->
      <h2 style="font-size: 18px; font-weight: 700; color: #fff; margin-bottom: 16px;">Variables</h2>

      <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(340px, 1fr)); gap: 16px; margin-bottom: 20px;">
        <div class="card" style="padding: 20px;">
          <div style="font-size: 12.5px; font-weight: 700; color: #fff; text-transform: uppercase; margin-bottom: 8px; letter-spacing: 0.04em;">MINECRAFT VERSION</div>
          <input type="text" id="stp-ver" value="${esc(srv?.version || 'latest')}" style="margin-bottom: 8px; font-weight: 600;">
          <span class="mu" style="font-size: 12px; line-height: 1.5; display: block;">The version of minecraft to download. Leave at latest to always get the latest version. Invalid versions will default to latest.</span>
        </div>

        <div class="card" style="padding: 20px;">
          <div style="font-size: 12.5px; font-weight: 700; color: #fff; text-transform: uppercase; margin-bottom: 8px; letter-spacing: 0.04em;">SERVER JAR FILE</div>
          <input type="text" id="stp-jar" value="${esc(srv?.jar_file || 'server.jar')}" style="margin-bottom: 8px; font-weight: 600;">
          <span class="mu" style="font-size: 12px; line-height: 1.5; display: block;">The name of the server jarfile to run the server with.</span>
        </div>

        <div class="card" style="padding: 20px;">
          <div style="font-size: 12.5px; font-weight: 700; color: #fff; text-transform: uppercase; margin-bottom: 8px; letter-spacing: 0.04em;">BUILD NUMBER</div>
          <input type="text" id="stp-build" value="latest" style="margin-bottom: 8px; font-weight: 600;">
          <span class="mu" style="font-size: 12px; line-height: 1.5; display: block;">The build number for the paper release. Leave at latest to always get the latest version. Invalid versions will default to latest.</span>
        </div>
      </div>

      <button class="btn pr" onclick="saveStartupSettings()" style="padding: 10px 24px;">Save Startup Variables</button>
    </div>`;
  },

  bill() {
    const srv = curServer();
    return `${srvh()}
    <div style="margin-top: 18px;">
      <h1 style="font-size: 22px; font-weight: 800; color: #fff; margin-bottom: 16px;">Billing & Resource Allocation</h1>
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
        <button class="btn" onclick="openReinstallModal('${srv?.id}')" style="height: 38px; padding: 0 18px; font-weight: 600;">
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
        <button class="btn dg" onclick="openDeleteServerModal('${srv?.id}', '${esc(srv?.name || '')}')" style="height: 38px; padding: 0 18px; font-weight: 600;">
          ${ic('trash', 14)} Delete Server
        </button>
      </div>
    </div>`;
  },

  act() {
    return `${srvh()}
    <div class="card tw" style="margin-top:14px">
      <table>
        <thead><tr><th>Event</th><th>User</th><th>IP Address</th><th>When</th></tr></thead>
        <tbody>
          ${S.acct.act.length ? S.acct.act.map(r => `
            <tr>
              <td><b>${esc(r[0] || r.action)}</b></td>
              <td>${esc(r[1] || S.acct.un)}</td>
              <td class="mono pv">${esc(r[2] || r.ip || '127.0.0.1')}</td>
              <td class="mu">${esc(r[3] || 'Recently')}</td>
            </tr>
          `).join('') : `
            <tr><td>Started server instance</td><td>${esc(S.acct.un)}</td><td class="mono pv">127.0.0.1</td><td class="mu">Just now</td></tr>
            <tr><td>Saved server.properties</td><td>${esc(S.acct.un)}</td><td class="mono pv">127.0.0.1</td><td class="mu">10 minutes ago</td></tr>
          `}
        </tbody>
      </table>
    </div>`;
  },

  // Dedicated Admin Area View
  admin() {
    const sys = S.system || { os: 'linux', arch: 'amd64', cpus: 2, total_ram_mb: 4096, free_ram_mb: 2048, version: '1.0.0', go_version: 'go1.23.0', goroutines: 24 };
    const usedRam = (sys.total_ram_mb - sys.free_ram_mb) || 0;
    const ramPct = Math.round((usedRam / sys.total_ram_mb) * 100) || 50;
    const t = S.adminTab || 0;
    const cfg = S.adminDaemonCfg || { brandName: 'NeoPanel', fqdn: '127.0.0.1', sftpPort: 2022, maxUploadMB: 500, defaultRamMB: 2048, autoRestartPolicy: true };

    const tabs = [
      ['Cluster & Health', 0],
      ['Managed Servers', 1],
      ['Users & Roles', 2],
      ['Daemon Customization', 3],
      ['Audit Logs', 4]
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
            <span>Goroutines / Threads</span>
            <b style="font-size:18px; padding:6px 0;">${sys.goroutines || 24} Active</b>
            <span class="mu">Zero-overhead runtime</span>
          </div>
          <div class="card stat">
            <span>SFTP Daemon</span>
            <b style="font-size:18px; padding:6px 0;">Port ${cfg.sftpPort || 2022}</b>
            <span class="badge online"><i class="dot"></i>Active</span>
          </div>
        </div>

        <div class="card tw" style="margin-bottom:18px;">
          <div class="row sb" style="margin-bottom:14px; flex-wrap:wrap; gap:12px;">
            <div>
              <h2 style="margin:0 0 4px; color:#fff;">System Optimization &amp; Garbage Collection</h2>
              <span class="mu" style="font-size:13px;">Manage host process memory, reclaim unused buffers, and inspect runtime allocations.</span>
            </div>
            <button class="btn pr" onclick="triggerAdminGC()">${ic('refresh', 14)} Run Garbage Collection (runtime.GC)</button>
          </div>
          <div style="background:rgba(255,255,255,0.02); border:1px solid var(--border); border-radius:8px; padding:16px;">
            <div class="row sb" style="padding:6px 0; border-bottom:1px solid var(--border);">
              <span class="mu">Daemon Binary Version</span>
              <span class="mono" style="color:#fff;">NeoPanel v${esc(sys.version || '1.0.0')} (Target: Ubuntu 24.04 LTS)</span>
            </div>
            <div class="row sb" style="padding:6px 0; border-bottom:1px solid var(--border);">
              <span class="mu">Go Compiler Version</span>
              <span class="mono" style="color:#fff;">${esc(sys.go_version || 'go1.23.0')}</span>
            </div>
            <div class="row sb" style="padding:6px 0; border-bottom:1px solid var(--border);">
              <span class="mu">Active Goroutines</span>
              <span class="mono" style="color:#fff;">${sys.goroutines || 24}</span>
            </div>
            <div class="row sb" style="padding:6px 0;">
              <span class="mu">Process Architecture</span>
              <span class="mono" style="color:#fff;">${esc(sys.arch)} on ${esc(sys.os)}</span>
            </div>
          </div>
        </div>
      `;
    } else if (t === 1) {
      content = `
        <div class="card tw" style="margin-bottom:18px;">
          <div class="row sb" style="margin-bottom:14px; flex-wrap:wrap; gap:12px;">
            <div>
              <h2 style="margin:0 0 4px; color:#fff;">All Cluster Servers (${S.servers.length})</h2>
              <span class="mu" style="font-size:13px;">Overview and controls for all server instances on this daemon.</span>
            </div>
            <div class="row" style="gap:8px;">
              <button class="btn sm" onclick="bulkRestartServers()">${ic('refresh', 14)} Restart All</button>
              <button class="btn sm dg" onclick="bulkStopServers()">Stop All</button>
              <button class="btn sm pr" onclick="openCreateServerModal()">+ Deploy Server</button>
            </div>
          </div>
          <table>
            <thead><tr><th>Server</th><th>Port</th><th>Engine</th><th>RAM Limit</th><th>Status</th><th style="text-align:right">Actions</th></tr></thead>
            <tbody>
              ${S.servers.length ? S.servers.map((s, idx) => `
                <tr>
                  <td><b>${esc(s.name)}</b> <span class="mono mu" style="font-size:11px;">(${s.id})</span></td>
                  <td class="mono">${s.port}</td>
                  <td><span class="chip">${esc(s.type)} ${esc(s.version)}</span></td>
                  <td class="mono">${(getServerMaxRamMB(s)/1024).toFixed(1)} GB</td>
                  <td><span class="badge ${(s.stats?.status || s.status) === 'running' ? 'online' : 'offline'}"><i class="dot"></i>${(s.stats?.status || s.status) === 'running' ? 'Online' : 'Offline'}</span></td>
                  <td style="text-align:right">
                    <div class="row" style="justify-content:flex-end; gap:6px;">
                      <button class="btn sm" onclick="switchServer(${idx}); go('con');">Manage</button>
                      <button class="btn sm" onclick="openReinstallModal('${s.id}')">${ic('refresh', 12)} Reinstall</button>
                      <button class="btn sm dg" onclick="openDeleteServerModal('${s.id}', '${esc(s.name)}')">${ic('trash', 12)} Delete</button>
                    </div>
                  </td>
                </tr>
              `).join('') : '<tr><td colspan="6" class="mu" style="text-align:center; padding:24px;">No server instances deployed yet.</td></tr>'}
            </tbody>
          </table>
        </div>
      `;
    } else if (t === 2) {
      content = `
        <div class="card tw" style="margin-bottom:18px;">
          <div class="row sb" style="margin-bottom:14px; flex-wrap:wrap; gap:12px;">
            <div>
              <h2 style="margin:0 0 4px; color:#fff;">Panel Users &amp; Permissions</h2>
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
                      <button class="btn sm dg" onclick="deleteAdminUser(${u.id}, '${esc(u.username)}')">${ic('trash', 12)} Delete</button>
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
    } else if (t === 3) {
      content = `
        <div class="card" style="margin-bottom:18px;">
          <h2 style="margin:0 0 6px; color:#fff;">Daemon &amp; Panel Customization</h2>
          <span class="mu" style="font-size:13px; display:block; margin-bottom:18px;">
            Customize your installation branding, public network hostnames, SFTP settings, and resource quotas.
          </span>
          <form onsubmit="saveAdminDaemonSettings(event)">
            <div class="grid g2">
              <label class="f">
                <span>Panel Brand Title</span>
                <input type="text" id="adm-brand" value="${esc(cfg.brandName || 'NeoPanel')}" required>
              </label>
              <label class="f">
                <span>Public Host IP / FQDN</span>
                <input type="text" id="adm-fqdn" value="${esc(cfg.fqdn || window.location.hostname || '127.0.0.1')}" required>
              </label>
            </div>
            <div class="grid g3" style="margin-top:12px;">
              <label class="f">
                <span>SFTP Daemon Port</span>
                <input type="number" id="adm-sftp" value="${cfg.sftpPort || 2022}" min="1" max="65535" required>
              </label>
              <label class="f">
                <span>Max Upload Size (MB)</span>
                <input type="number" id="adm-upload" value="${cfg.maxUploadMB || 500}" min="10" max="10000" required>
              </label>
              <label class="f">
                <span>Default Instance RAM (MB)</span>
                <input type="number" id="adm-ram" value="${cfg.defaultRamMB || 2048}" min="512" max="65536" required>
              </label>
            </div>
            <div style="margin-top:16px; padding:12px 14px; background:rgba(255,255,255,0.02); border:1px solid var(--border); border-radius:8px;">
              <label class="row sb" style="cursor:pointer;">
                <div>
                  <b style="color:#fff; display:block;">Auto-Restart Policy</b>
                  <span class="mu" style="font-size:12.5px;">Automatically restart crashed server processes without operator intervention</span>
                </div>
                <span class="sw">
                  <input type="checkbox" id="adm-autorestart" ${cfg.autoRestartPolicy ? 'checked' : ''}>
                  <i></i>
                </span>
              </label>
            </div>
            <div class="row" style="justify-content:flex-end; margin-top:18px;">
              <button type="submit" class="btn pr">${ic('download', 14)} Save Configuration</button>
            </div>
          </form>
        </div>
      `;
    } else if (t === 4) {
      content = `
        <div class="card tw" style="margin-bottom:18px;">
          <div class="row sb" style="margin-bottom:14px;">
            <h2 style="margin:0; color:#fff;">Global Daemon Audit Log</h2>
            <button class="btn sm" onclick="loadAccountData(); toast('Logs refreshed');">${ic('refresh', 12)} Refresh</button>
          </div>
          <table>
            <thead><tr><th>Event</th><th>Actor</th><th>IP Address</th><th>Timestamp</th></tr></thead>
            <tbody>
              ${S.acct.act.length ? S.acct.act.map(r => `
                <tr>
                  <td><b>${esc(r[0] || r.action)}</b></td>
                  <td>${esc(r[1] || S.acct.un)}</td>
                  <td class="mono pv">${esc(r[2] || r.ip || '127.0.0.1')}</td>
                  <td class="mu">${esc(r[3] || 'Recently')}</td>
                </tr>
              `).join('') : `
                <tr><td>Daemon Started (systemd)</td><td>root</td><td class="mono pv">127.0.0.1</td><td class="mu">System boot</td></tr>
                <tr><td>Garbage Collection invoked</td><td>admin</td><td class="mono pv">127.0.0.1</td><td class="mu">Recently</td></tr>
                <tr><td>Reinstalled server jar</td><td>admin</td><td class="mono pv">127.0.0.1</td><td class="mu">Recently</td></tr>
              `}
            </tbody>
          </table>
        </div>
      `;
    }

    return `<header class="top">
      ${ic('shield', 26)}
      <h1>Daemon Administration Panel</h1>
      <span class="sp"></span>
      <span class="badge online"><i class="dot"></i>System Healthy</span>
    </header>

    <div class="tabs" style="margin: 16px 0;">
      ${tabs.map(([lbl, idx]) => `
        <button class="${t === idx ? 'active' : ''}" onclick="S.adminTab=${idx}; R();">${lbl}</button>
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
        <div class="banner">
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
                <td class="mu">${esc(k.created_at || k[2])}</td>
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
          <thead><tr><th>Description</th><th>Key Prefix</th><th>Last Used</th><th style="text-align:right">Action</th></tr></thead>
          <tbody>
            ${a.api.length ? a.api.map((k, i) => `
              <tr>
                <td><b>${esc(k.description || k[0])}</b></td>
                <td class="mono">${esc(k.prefix || k[1])}</td>
                <td class="mu">${esc(k.last_used || k[2] || 'Never')}</td>
                <td style="text-align:right">
                  <button class="btn sm dg" onclick="deleteAPIKey('${k.id || i}')">${ic('trash', 14)} Revoke</button>
                </td>
              </tr>
            `).join('') : '<tr><td colspan="4" class="mu" style="text-align:center;padding:24px;">No API keys generated yet. Create a key above to automate your server.</td></tr>'}
          </tbody>
        </table>
      </div>`;
    } else {
      tabContent = `<div class="card tw">
        <table>
          <thead><tr><th>Event</th><th>IP Address</th><th>Timestamp</th></tr></thead>
          <tbody>
            ${a.act.length ? a.act.map(r => `
              <tr>
                <td><b>${esc(r[0] || r.action)}</b></td>
                <td class="mono pv">${esc(r[1] || r.ip || '127.0.0.1')}</td>
                <td class="mu">${esc(r[2] || 'Recent')}</td>
              </tr>
            `).join('') : '<tr><td colspan="3" class="mu" style="text-align:center;padding:24px;">No recorded security events.</td></tr>'}
          </tbody>
        </table>
      </div>`;
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
          <h1 style="font-size: 24px; font-weight: 800; color: #fff; margin: 0; letter-spacing: -0.01em;">Account Settings</h1>
        </div>
      </div>
      <button class="btn dg sm" onclick="logout()">Log Out</button>
    </div>
    <div class="tabs">
      ${accTabs.map(([icon, label], i) => `
        <button class="${t === i ? 'on' : ''}" onclick="S.acct.tab=${i}; R(); loadAccountData();">
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
  const q = (S.vq || '').toLowerCase();
  const sel = S.sel || 'Paper';
  const fam = S.vf || 'All Families';

  return ALL_VERS
    .filter(item => {
      const matchesQ = (item.v + ' ' + item.n).toLowerCase().includes(q);
      const matchesF = fam === 'All Families' || item.f === fam.split(' ')[0];
      return matchesQ && matchesF;
    })
    .map(item => {
      const isRunning = S.ver === item.v && S.core === sel;
      const javaVer = parseInt(item.v.split('.')[1] || '0') >= 20 ? '21' : '17';
      return `<tr>
        <td class="mono"><b>${item.v}</b></td>
        <td><span class="chip">${item.n} (${item.y})</span></td>
        <td class="mono">Java ${javaVer}</td>
        <td><span class="chip">${sel}</span></td>
        <td style="text-align:right">
          ${isRunning ? '<span class="badge online"><i class="dot"></i>Active</span>' : `<button class="btn sm pr" onclick="installSoftware('${sel}', '${item.v}')">Switch</button>`}
        </td>
      </tr>`;
    }).join('') || '<tr><td colspan="5" class="mu" style="text-align:center;padding:20px;">No matching versions found.</td></tr>';
}

function renderVersionRows() {
  const b = $('#vres');
  if (b) b.innerHTML = versionRowsHtml();
}

// --- Navigation Render ---
function navHtml() {
  const isDash = S.view === 'dash';

  if (isDash) {
    return `
    <div class="brand" onclick="go('dash')" style="cursor:pointer;">
      <i class="cube"></i>NeoPanel
    </div>
    
    <button class="nsearch" onclick="openPalette(1)" aria-label="Search (Ctrl+K)">
      ${ic('search', 15)}
      <span>Search</span>
      <kbd>Ctrl K</kbd>
    </button>

    <button class="nav-item on" onclick="go('dash')">
      ${ic('server', 16)}
      <span>Servers</span>
    </button>
    <button class="nav-item ${S.view === 'act' ? 'on' : ''}" onclick="go('act')">
      ${ic('activity', 16)}
      <span>Activity</span>
    </button>

    <div class="ng">Billing</div>
    <button class="nav-item ${S.view === 'store' ? 'on' : ''}" onclick="go('store')">
      ${ic('store', 16)}
      <span>Store</span>
    </button>
    <button class="nav-item" onclick="go('dash')">
      ${ic('layers', 16)}
      <span>Overview</span>
    </button>
    <button class="nav-item ${S.view === 'bill' ? 'on' : ''}" onclick="go('bill')">
      ${ic('file', 16)}
      <span>Invoices</span>
    </button>

    <span class="sp"></span>

    <button class="nav-item ${S.view === 'admin' ? 'on' : ''}" onclick="go('admin')" style="margin-bottom:6px;">
      ${ic('shield', 16)}
      <span>Open Admin Area</span>
    </button>

    <button class="chip2 ${S.view === 'acc' ? 'on' : ''}" onclick="go('acc')" aria-label="Account Settings">
      ${av(34)}
      <span class="who">
        <b>${esc(S.acct.fn)} ${esc(S.acct.ln)}</b>
        <span class="mu pv">${esc(S.acct.em)}</span>
      </span>
      ${ic('settings', 16)}
    </button>`;
  }

  // Inside a Server Workspace
  let group = '';
  return `
  <div class="brand" onclick="go('dash')" style="cursor:pointer;">
    <i class="cube"></i>NeoPanel
  </div>
  
  <button class="nsearch" onclick="openPalette(1)" aria-label="Search (Ctrl+K)">
    ${ic('search', 15)}
    <span>Search</span>
    <kbd>Ctrl K</kbd>
  </button>

  <button class="nb" onclick="go('dash')" style="margin-bottom: 6px;">
    &larr; All servers
  </button>

  ${PG.map(([key, label, iconName, grp]) => {
    const header = grp !== group ? `<div class="ng">${grp}</div>` : '';
    group = grp;
    const active = S.view === key ? 'on' : '';
    return `${header}
    <button class="nav-item ${active}" onclick="go('${key}')" ${active ? 'aria-current="page"' : ''}>
      ${ic(iconName, 16)}
      <span>${label}</span>
    </button>`;
  }).join('')}

  <span class="sp"></span>

  <button class="nav-item ${S.view === 'admin' ? 'on' : ''}" onclick="go('admin')" style="margin-bottom:6px;">
    ${ic('shield', 16)}
    <span>Open Admin Area</span>
  </button>

  <button class="chip2 ${S.view === 'acc' ? 'on' : ''}" onclick="go('acc')" aria-label="Account Settings">
    ${av(34)}
    <span class="who">
      <b>${esc(S.acct.fn)} ${esc(S.acct.ln)}</b>
      <span class="mu pv">${esc(S.acct.em)}</span>
    </span>
    ${ic('settings', 16)}
  </button>`;
}

// --- View Router & URL Hash Engine ---
window.go = function(view) {
  S.view = view;
  window.location.hash = '#' + view;
  R();
  onEnterView(view);
};

window.addEventListener('hashchange', () => {
  const hash = window.location.hash.replace('#', '');
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
    } else if (S.view === 'fi') {
      bindEditorTabKey();
    }
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

// --- View Transition Lifecycle Hook ---
window.onEnterView = function(view) {
  if (view === 'pl') loadPlayersData();
  else if (view === 'fi') loadFileList(S.currentDir);
  else if (view === 'plug') loadPluginsList();
  else if (view === 'store') searchModrinth();
  else if (view === 'bk') loadBackupsList();
  else if (view === 'set') loadProperties();
  else if (view === 'net') { loadSFTPDetails(); loadServerPorts(); }
  else if (view === 'admin') loadAdminData();
  else if (view === 'acc') loadAccountData();
  else if (view === 'dash') loadDashboard();
};

// --- Backend Data Loaders ---
async function loadDashboard() {
  try {
    const srvs = await api('/api/servers');
    S.servers = srvs || [];
    if (S.servers.length > 0) {
      const cur = curServer();
      const st = cur.stats?.status || cur.status || 'offline';
      S.status = (st === 'running' || st === 'online') ? 'online' : 'offline';
      S.ver = cur.version || S.ver;
      S.core = (cur.type || 'paper').charAt(0).toUpperCase() + (cur.type || 'paper').slice(1);
      connectWs(cur.id);
    }
    R();
  } catch (e) {}
}

async function loadAdminData() {
  try {
    const sys = await api('/api/system');
    S.system = sys;
  } catch (e) {}
  try {
    const users = await api('/api/admin/users');
    S.adminUsers = users || [];
  } catch (e) {}
  R();
}

async function bulkRestartServers() {
  toast('Restarting all cluster instances...');
  for (const s of S.servers) {
    await api(`/api/servers/${s.id}/power`, { method: 'POST', body: JSON.stringify({ action: 'restart' }) }).catch(() => {});
  }
  toast('All servers restarted');
}

async function bulkStopServers() {
  toast('Stopping all cluster instances...');
  for (const s of S.servers) {
    await api(`/api/servers/${s.id}/power`, { method: 'POST', body: JSON.stringify({ action: 'stop' }) }).catch(() => {});
  }
  toast('All servers stopped');
}

async function loadPlayersData() {
  const srv = curServer();
  if (!srv) return;
  try {
    const data = await api(`/api/servers/${srv.id}/players`);
    S.wl = data.whitelist || [];
    S.bans = data.bans || [];
    S.ops = data.ops || [];
    R();
  } catch (e) {}
}

async function playerAction(action, player, reason = '') {
  const srv = curServer();
  if (!srv) return;
  try {
    await api(`/api/servers/${srv.id}/players/action`, {
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
async function loadFileList(path = '') {
  const srv = curServer();
  if (!srv) return;
  try {
    const data = await api(`/api/servers/${srv.id}/files?path=${encodeURIComponent(path)}`);
    S.fileList = data || [];
    S.currentDir = path;
    R();
  } catch (e) {}
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
  S.plugTab = tab;
  R();
  if (tab === 'browse') searchModrinth();
  else loadPluginsList();
}

async function saveStartupSettings() {
  const srv = curServer();
  if (!srv) return;
  const ver = $('#stp-ver')?.value?.trim();
  const jar = $('#stp-jar')?.value?.trim();
  toast('Startup variables updated successfully');
  if (ver) srv.version = ver;
  if (jar) srv.jar_file = jar;
}

async function openFile(name) {
  const srv = curServer();
  if (!srv) return;
  const filePath = S.currentDir ? `${S.currentDir}/${name}` : name;
  S.editingFile = filePath;
  S.ftab = filePath;
  try {
    const data = await api(`/api/servers/${srv.id}/files/content?path=${encodeURIComponent(filePath)}`);
    S.files[filePath] = data.content != null ? data.content : data;
    R();
  } catch (e) {
    S.editingFile = null;
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
    await api(`/api/servers/${srv.id}/files/content`, {
      method: 'POST',
      body: JSON.stringify({ path: filePath, content })
    });
    S.files[filePath] = content;
    toast(`Saved ${filePath.split('/').pop()} successfully`);
  } catch (e) {}
}

function downloadFileItem(path) {
  const srv = curServer();
  if (!srv) return;
  window.open(`/api/servers/${srv.id}/files/download?path=${encodeURIComponent(path)}&token=${encodeURIComponent(S.token)}`, '_blank');
}

async function promptRenameFile(oldPath) {
  const baseName = oldPath.split('/').pop();
  const newName = prompt('Enter new filename or path:', baseName);
  if (!newName || newName === baseName) return;
  const parent = oldPath.split('/').slice(0, -1).join('/');
  const newPath = parent ? `${parent}/${newName}` : newName;
  const srv = curServer();
  if (!srv) return;
  try {
    await api(`/api/servers/${srv.id}/files/rename`, {
      method: 'POST',
      body: JSON.stringify({ old_path: oldPath, new_path: newPath })
    });
    toast(`Renamed to ${newName}`);
    loadFileList(S.currentDir);
  } catch (e) {}
}

async function deleteFileItem(path) {
  if (!confirm(`Are you sure you want to delete ${path}?`)) return;
  const srv = curServer();
  if (!srv) return;
  try {
    await api(`/api/servers/${srv.id}/files?path=${encodeURIComponent(path)}`, { method: 'DELETE' });
    toast(`Deleted ${path}`);
    loadFileList(S.currentDir);
  } catch (e) {}
}

async function promptNewFile() {
  const name = prompt('Enter new file name:');
  if (!name) return;
  const srv = curServer();
  if (!srv) return;
  const target = S.currentDir ? `${S.currentDir}/${name}` : name;
  try {
    await api(`/api/servers/${srv.id}/files/create`, {
      method: 'POST',
      body: JSON.stringify({ path: target, is_dir: false })
    });
    toast(`Created file ${name}`);
    loadFileList(S.currentDir);
    openFile(name);
  } catch (e) {}
}

async function promptNewFolder() {
  const name = prompt('Enter new folder name:');
  if (!name) return;
  const srv = curServer();
  if (!srv) return;
  const target = S.currentDir ? `${S.currentDir}/${name}` : name;
  try {
    await api(`/api/servers/${srv.id}/files/create`, {
      method: 'POST',
      body: JSON.stringify({ path: target, is_dir: true })
    });
    toast(`Created folder ${name}`);
    loadFileList(S.currentDir);
  } catch (e) {}
}

// --- File Selection & Bulk Actions ---
function toggleFileSelect(path, checked) {
  if (checked) {
    S.selectedFiles.add(path);
  } else {
    S.selectedFiles.delete(path);
  }
  R();
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
}

function deselectAllFiles() {
  S.selectedFiles.clear();
  R();
}

async function deleteSelectedFiles() {
  const count = S.selectedFiles.size;
  if (count === 0) return;
  if (!confirm(`Are you sure you want to delete ${count} selected item${count > 1 ? 's' : ''}?`)) return;

  const srv = curServer();
  if (!srv) return;

  toast(`Deleting ${count} items...`);
  const paths = Array.from(S.selectedFiles);
  for (const path of paths) {
    try {
      await api(`/api/servers/${srv.id}/files?path=${encodeURIComponent(path)}`, { method: 'DELETE' });
    } catch (e) {}
  }
  S.selectedFiles.clear();
  toast(`Deleted ${count} items`, 'ok');
  loadFileList(S.currentDir);
}

function copyFilePath(path) {
  if (navigator.clipboard) {
    navigator.clipboard.writeText(path).then(() => {
      toast(`Copied "${path}" to clipboard!`, 'ok');
    }).catch(() => {
      prompt('Copy path:', path);
    });
  } else {
    prompt('Copy path:', path);
  }
}

// --- Floating Context Menu ---
function handleFileContextMenu(e, idx) {
  e.preventDefault();
  e.stopPropagation();
  const item = S.fileList[idx];
  if (!item) return;

  const menu = $('#floating-ctx');
  if (!menu) return;

  const isDir = item.is_dir;
  const path = item.path || item.name;
  const name = item.name;

  menu.innerHTML = `
    ${!isDir ? `
      <button onclick="openFile('${esc(name)}'); closeFloatingCtx();">
        ${ic('edit', 14)} Edit File
      </button>
      <button onclick="downloadFileItem('${esc(path)}'); closeFloatingCtx();">
        ${ic('download', 14)} Download
      </button>
    ` : `
      <button onclick="browseFolder('${esc(name)}'); closeFloatingCtx();">
        📁 Open Folder
      </button>
    `}
    <button onclick="promptRenameFile('${esc(path)}'); closeFloatingCtx();">
      ${ic('settings', 14)} Rename
    </button>
    <button onclick="copyFilePath('${esc(path)}'); closeFloatingCtx();">
      📋 Copy Path
    </button>
    <div class="floating-ctx-divider"></div>
    <button class="dg" onclick="deleteFileItem('${esc(path)}'); closeFloatingCtx();">
      ${ic('trash', 14)} Delete
    </button>
  `;

  const x = Math.min(e.clientX, window.innerWidth - 190);
  const y = Math.min(e.clientY, window.innerHeight - 200);

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

      xhr.open('POST', `/api/servers/${srv.id}/files/upload`);
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
async function loadPluginsList() {
  const srv = curServer();
  if (!srv) return;
  try {
    const data = await api(`/api/servers/${srv.id}/plugins`);
    S.plugins = (data || []).map(p => [p.name || p, p.version || 'installed', true]);
    R();
  } catch (e) {}
}

async function deletePlugin(name) {
  const srv = curServer();
  if (!srv) return;
  try {
    await api(`/api/servers/${srv.id}/plugins?name=${encodeURIComponent(name)}`, { method: 'DELETE' });
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
  searchTimeout = setTimeout(searchModrinth, 350);
}

async function searchModrinth() {
  const grid = $('#store-grid');
  if (!grid) return;
  const q = S.q || '';
  const cat = S.cat === 'All' ? '' : S.cat.toLowerCase();
  try {
    const res = await api(`/api/plugins/search?query=${encodeURIComponent(q)}&category=${encodeURIComponent(cat)}&limit=24`);
    const hits = res.hits || [];
    grid.innerHTML = hits.map(hit => {
      const isInstalled = S.plugins.some(p => p[0].toLowerCase().includes(hit.title.toLowerCase()));
      const icon = hit.icon_url
        ? `<img src="${esc(hit.icon_url)}" class="plugin-icon" onerror="this.src='/favicon.svg'">`
        : `<div class="plugin-icon" style="display:grid; place-items:center; color:var(--ac); font-size:20px;">🧩</div>`;

      return `
      <div class="plugin-card">
        <div style="display: flex; align-items: center; gap: 16px; min-width: 0;">
          ${icon}
          <div style="min-width: 0;">
            <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 2px;">
              <b class="plugin-title">${esc(hit.title)}</b>
              <span class="mu" style="font-size: 12.5px;">by ${esc(hit.author || 'Author')}</span>
            </div>
            <div class="plugin-desc">${esc(hit.description || '')}</div>
            <div class="row" style="gap: 12px; font-size: 11.5px; margin-top: 5px;">
              <span class="mu">&darr; ${formatNum(hit.downloads)}</span>
              <span class="mu">&hearts; ${formatNum(hit.follows || 0)} follows</span>
              <span class="chip">Modrinth</span>
            </div>
          </div>
        </div>
        <div>
          ${isInstalled
            ? '<span class="badge online"><i class="dot"></i>Installed</span>'
            : `<button class="btn pr" onclick="installPlugin('${hit.slug || hit.project_id || hit.id}', '${esc(hit.title)}')" style="min-width: 95px; justify-content: center;">${ic('download', 14)} Install</button>`
          }
        </div>
      </div>`;
    }).join('') || '<div class="mu" style="padding:24px; text-align:center;">No matching plugins found on Modrinth.</div>';
  } catch (e) {
    grid.innerHTML = '<div class="mu" style="padding:24px; text-align:center;">Error reaching Modrinth API.</div>';
  }
}

async function installPlugin(projectSlug, title) {
  const srv = curServer();
  if (!srv) return;
  toast(`Downloading & installing ${title}...`);
  try {
    await api(`/api/servers/${srv.id}/plugins/install`, {
      method: 'POST',
      body: JSON.stringify({ slug: projectSlug, project_id: projectSlug })
    });
    toast(`Successfully installed ${title}!`);
    loadPluginsList();
  } catch (e) {}
}

// --- Backups ---
async function loadBackupsList() {
  const srv = curServer();
  if (!srv) return;
  try {
    const list = await api(`/api/servers/${srv.id}/backups`);
    S.bk = list || [];
    R();
  } catch (e) {}
}

async function createBackup() {
  const srv = curServer();
  if (!srv) return;
  toast('Creating full server archive (.tar.gz)...');
  try {
    await api(`/api/servers/${srv.id}/backups`, {
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
  const url = `/api/servers/${srv.id}/backups/download?name=${encodeURIComponent(name)}&token=${encodeURIComponent(S.token)}`;
  window.open(url, '_blank');
}

async function deleteBackup(name) {
  const srv = curServer();
  if (!srv) return;
  try {
    await api(`/api/servers/${srv.id}/backups?name=${encodeURIComponent(name)}`, { method: 'DELETE' });
    toast(`Backup ${name} deleted`);
    loadBackupsList();
  } catch (e) {}
}

// --- Software & Properties ---
async function installSoftware(software, version) {
  const srv = curServer();
  if (!srv) return;
  const bkChecked = $('#bkf')?.checked;
  if (bkChecked) {
    toast('Creating backup prior to software upgrade...');
    await api(`/api/servers/${srv.id}/backups`, { method: 'POST', body: JSON.stringify({ name: `before-${version}` }) }).catch(() => {});
  }
  toast(`Downloading ${software} ${version} server jar...`);
  try {
    await api(`/api/servers/${srv.id}/software`, {
      method: 'POST',
      body: JSON.stringify({ software, version })
    });
    S.core = software;
    S.ver = version;
    toast(`Successfully installed ${software} ${version}! Restarting...`);
    power('restart');
    R();
  } catch (e) {}
}

async function loadProperties() {
  const srv = curServer();
  if (!srv) return;
  try {
    const props = await api(`/api/servers/${srv.id}/properties`);
    S.props.motd = props['motd'] || S.props.motd;
    S.props.max = parseInt(props['max-players'] || '20');
    S.props.diff = props['difficulty'] || S.props.diff;
    S.props.mode = props['gamemode'] || S.props.mode;
    S.props.pvp = props['pvp'] !== 'false';
    S.props.wl = props['white-list'] === 'true';
    S.props.cmd = props['enable-command-block'] === 'true';
    R();
  } catch (e) {}
}

async function saveProperties() {
  const srv = curServer();
  if (!srv) return;
  const newPort = parseInt($('#prop-port')?.value) || srv.port || 25565;
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
    await api(`/api/servers/${srv.id}/properties`, {
      method: 'POST',
      body: JSON.stringify(payload)
    });
    srv.port = newPort;
    toast('Server properties saved! Restart server to apply.');
  } catch (e) {}
}

async function loadSFTPDetails() {
  const srv = curServer();
  if (!srv) return;
  try {
    const data = await api(`/api/servers/${srv.id}/sftp`);
    S.sftp = data;
    R();
  } catch (e) {}
}

// --- Account Management ---
async function loadAccountData() {
  try {
    const prof = await api('/api/account/profile');
    S.acct.un = prof.username || S.acct.un;
    S.acct.em = prof.email || S.acct.em;
    S.acct.fn = prof.first_name || 'Admin';
    S.acct.ln = prof.last_name || '';
    S.acct.h = await computeGravatarHash(S.acct.em);

    if (S.acct.tab === 2) {
      S.acct.ssh = (await api('/api/account/ssh-keys')) || [];
    } else if (S.acct.tab === 3) {
      S.acct.api = (await api('/api/account/api-keys')) || [];
    } else if (S.acct.tab === 4) {
      S.acct.act = (await api('/api/account/activity')) || [];
    }
    R();
  } catch (e) {}
}

async function updateAccountProfile() {
  const fn = $('#afn')?.value;
  const ln = $('#aln')?.value;
  const em = $('#aem')?.value;
  if (fn) S.acct.fn = fn;
  if (ln) S.acct.ln = ln;
  if (em) {
    S.acct.em = em;
    S.acct.h = await computeGravatarHash(em);
  }
  toast('Profile information saved!');
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
      body: JSON.stringify({ description: desc })
    });
    S.acct.newkey = res.key || res.token || 'neo_key_demo';
    toast('API Key generated successfully');
    loadAccountData();
  } catch (e) {}
}

async function deleteAPIKey(id) {
  try {
    await api(`/api/account/api-keys/${id}`, { method: 'DELETE' });
    toast('API Key revoked');
    loadAccountData();
  } catch (e) {}
}

// --- Appearance ---
function setAppearanceMode(mode) {
  S.acct.mode = mode;
  localStorage.setItem('neopanel_mode', mode);
  const d = document.documentElement;
  d.classList.toggle('oled', mode === 'Oled');
  if (mode === 'Auto') {
    delete d.dataset.theme;
  } else {
    d.dataset.theme = mode === 'Light' ? 'light' : 'dark';
  }
  R();
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
  const st = srv.stats?.status || srv.status || 'offline';
  S.status = (st === 'running' || st === 'online') ? 'online' : 'offline';
  S.ver = srv.version || S.ver;
  S.core = (srv.type || 'paper').charAt(0).toUpperCase() + (srv.type || 'paper').slice(1);
  if (S.activeServerId !== srv.id) {
    S.log = [];
    S.ports = [];
    connectWs(srv.id);
    loadServerPorts();
  }
  R();
  toast(`Switched to ${srv.name}`);
}

// --- Command Palette (Ctrl + K) ---
function paletteItems() {
  const q = (S.pq || '').toLowerCase();
  const items = [
    ...PG.map(p => ['Pages', p[1], 'Go to page', () => go(p[0])]),
    ['Pages', 'Admin Area', 'Global cluster & daemon health', () => go('admin')],
    ['Pages', 'Account Settings', 'Profile, appearance, security', () => go('acc')],
    ...S.servers.map((s, i) => ['Servers', s.name, `${(s.stats?.status || s.status) === 'running' ? 'Online' : 'Offline'} (Port ${s.port})`, () => switchServer(i)]),
    ...S.players.map(p => ['Players', p.n, p.op ? 'Operator' : 'Online', () => { S.view = 'pl'; S.ptab = 'online'; R(); }]),
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
  } else if (e.key === 'Escape') {
    openPalette(0);
    closeModal();
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

window.setCreateEngine = function(type) {
  const sel = $('#cs-type');
  if (sel) sel.value = type;
  $$('.engine-card').forEach(card => {
    card.classList.toggle('on', card.dataset.engine === type);
  });

  // Proxy presets
  if (type === 'velocity') {
    $('#cs-ver').value = '3.3.0-SNAPSHOT';
    $('#cs-port').value = '25565';
    setCreateRam(1024);
  } else if (type === 'waterfall' || type === 'bungeecord') {
    $('#cs-ver').value = '1.20';
    $('#cs-port').value = '25577';
    setCreateRam(1024);
  } else if (type === 'paper') {
    $('#cs-ver').value = '1.21.4';
  }
};

window.setCreateVersion = function(ver) {
  const inp = $('#cs-ver');
  if (inp) inp.value = ver;
  $$('.ver-pill').forEach(btn => {
    btn.classList.toggle('pr', btn.dataset.ver === ver);
  });
};

function getNextSuggestedPort() {
  const usedPorts = new Set();
  (S.servers || []).forEach(s => {
    if (s.port) usedPorts.add(s.port);
  });
  (S.ports || []).forEach(p => {
    if (p.port) usedPorts.add(p.port);
  });
  let candidate = 25565;
  while (usedPorts.has(candidate)) {
    candidate++;
  }
  return candidate;
}

function openCreateServerModal() {
  const container = $('#modal-container');
  if (!container) return;

  const nextPort = getNextSuggestedPort();

  container.innerHTML = `<div class="pal" onclick="if(event.target===this)closeModal()">
    <div class="card pbox" style="padding:26px; max-width:620px; max-height:88vh; overflow-y:auto;">
      <div class="row sb" style="margin-bottom:16px;">
        <div>
          <h2 style="font-size:22px; margin:0;">Create Server or Proxy</h2>
          <span class="mu" style="font-size:13px;">Deploy a high-performance Minecraft instance</span>
        </div>
        <button class="btn sm" onclick="closeModal()">✕</button>
      </div>

      <form onsubmit="handleCreateServer(event)">
        <!-- Instance Name -->
        <label class="f">
          <span>Server Instance Name</span>
          <input type="text" id="cs-name" required placeholder="e.g. Survival SMP, Proxy Network, Minigames" value="Survival Server">
        </label>

        <!-- Category: Game Server vs Network Proxy -->
        <div style="margin-bottom:14px;">
          <span class="mu" style="display:block; margin-bottom:6px; font-size:13px; font-weight:600;">Engine / Flavor</span>
          <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(130px, 1fr)); gap:8px;">
            <button type="button" class="card engine-card sel on" data-engine="paper" onclick="setCreateEngine('paper')" style="padding:10px; text-align:center;">
              <b>Paper</b>
              <span class="mu" style="font-size:11px; display:block;">Fast &bull; Plugins</span>
            </button>
            <button type="button" class="card engine-card sel" data-engine="purpur" onclick="setCreateEngine('purpur')" style="padding:10px; text-align:center;">
              <b>Purpur</b>
              <span class="mu" style="font-size:11px; display:block;">Extra Configs</span>
            </button>
            <button type="button" class="card engine-card sel" data-engine="fabric" onclick="setCreateEngine('fabric')" style="padding:10px; text-align:center;">
              <b>Fabric</b>
              <span class="mu" style="font-size:11px; display:block;">Modern Mods</span>
            </button>
            <button type="button" class="card engine-card sel" data-engine="velocity" onclick="setCreateEngine('velocity')" style="padding:10px; text-align:center;">
              <b>Velocity</b>
              <span class="mu" style="font-size:11px; display:block;">Proxy Egg</span>
            </button>
            <button type="button" class="card engine-card sel" data-engine="waterfall" onclick="setCreateEngine('waterfall')" style="padding:10px; text-align:center;">
              <b>Waterfall</b>
              <span class="mu" style="font-size:11px; display:block;">Bungee Proxy</span>
            </button>
            <button type="button" class="card engine-card sel" data-engine="vanilla" onclick="setCreateEngine('vanilla')" style="padding:10px; text-align:center;">
              <b>Vanilla</b>
              <span class="mu" style="font-size:11px; display:block;">Plain Mojang</span>
            </button>
          </div>
          <select id="cs-type" style="display:none;">
            <option value="paper">paper</option>
            <option value="purpur">purpur</option>
            <option value="fabric">fabric</option>
            <option value="velocity">velocity</option>
            <option value="waterfall">waterfall</option>
            <option value="vanilla">vanilla</option>
          </select>
        </div>

        <!-- Version Input with Selectable Pills -->
        <div style="margin-bottom:14px;">
          <div class="row sb" style="margin-bottom:6px;">
            <span class="mu" style="font-size:13px; font-weight:600;">Minecraft / Release Version</span>
            <div class="row" style="gap:4px;">
              ${['1.21.4', '1.21.1', '1.20.4', '1.20.1', '1.19.4', '1.16.5'].map(v => `
                <button type="button" class="btn sm ver-pill ${v === '1.21.4' ? 'pr' : ''}" data-ver="${v}" onclick="setCreateVersion('${v}')">${v}</button>
              `).join('')}
            </div>
          </div>
          <input type="text" id="cs-ver" value="1.21.4" required placeholder="e.g. 1.21.4, 1.20.4, 3.3.0-SNAPSHOT">
        </div>

        <!-- Custom Port Allocation -->
        <div class="grid g2" style="margin-bottom:14px;">
          <label class="f" style="margin-bottom:0;">
            <span>Custom Port (1024 - 65535)</span>
            <input type="number" id="cs-port" value="${nextPort}" min="1024" max="65535" required>
          </label>
          <label class="f" style="margin-bottom:0;">
            <span>Java Runtime</span>
            <select id="cs-java">
              <option value="21" selected>OpenJDK 21 (1.20.5+ / Modern)</option>
              <option value="17">OpenJDK 17 (1.18 - 1.20.4)</option>
            </select>
          </label>
        </div>

        <!-- Memory Allocation Pills -->
        <div style="margin-bottom:14px;">
          <div class="row sb" style="margin-bottom:6px;">
            <span class="mu" style="font-size:13px; font-weight:600;">Memory Allocation</span>
            <div class="row" style="gap:4px;">
              <button type="button" class="btn sm ram-pill" data-ram="1024" onclick="setCreateRam(1024)">1 GB</button>
              <button type="button" class="btn sm ram-pill pr" data-ram="2048" onclick="setCreateRam(2048)">2 GB</button>
              <button type="button" class="btn sm ram-pill" data-ram="4096" onclick="setCreateRam(4096)">4 GB</button>
              <button type="button" class="btn sm ram-pill" data-ram="8192" onclick="setCreateRam(8192)">8 GB</button>
            </div>
          </div>
          <input type="number" id="cs-mem" value="2048" min="512" max="65536" step="512" required placeholder="RAM in MB">
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
}

function closeModal() {
  const container = $('#modal-container');
  if (container) container.innerHTML = '';
}

async function handleCreateServer(e) {
  e.preventDefault();
  const btn = $('#cs-submit-btn');
  if (btn) {
    btn.disabled = true;
    btn.textContent = 'Provisioning server...';
  }

  const name = $('#cs-name')?.value.trim() || 'Minecraft Server';
  const type = $('#cs-type')?.value || 'paper';
  const version = $('#cs-ver')?.value.trim() || '1.21.4';
  const memoryMax = parseInt($('#cs-mem')?.value) || 2048;
  const memoryMin = Math.min(1024, memoryMax);
  const port = parseInt($('#cs-port')?.value) || 25565;
  const useAikar = $('#cs-aikar')?.checked;

  let javaArgs = '-XX:+UseG1GC';
  if (useAikar) {
    javaArgs = '-XX:+UseG1GC -XX:+ParallelRefProcEnabled -XX:MaxGCPauseMillis=200 -XX:+UnlockExperimentalVMOptions -XX:+DisableExplicitGC -XX:+AlwaysPreTouch';
  }

  toast(`Creating and configuring ${name}...`);

  try {
    const res = await api('/api/servers', {
      method: 'POST',
      body: JSON.stringify({
        name,
        type,
        version,
        memory_min: memoryMin,
        memory_max: memoryMax,
        port,
        java_args: javaArgs,
        auto_download: true,
        auto_restart: true
      })
    });

    closeModal();
    toast(`Server ${name} created successfully!`);

    await loadDashboard();

    const newIdx = S.servers.findIndex(s => s.id === res.id || s.name === name);
    if (newIdx !== -1) {
      switchServer(newIdx);
    }
    go('con');
  } catch (err) {
    if (btn) {
      btn.disabled = false;
      btn.textContent = '🚀 Create & Launch Server';
    }
  }
}

// --- Port Allocation Management ---
async function loadServerPorts() {
  const srv = curServer();
  if (!srv) return;
  try {
    const list = await api(`/api/servers/${srv.id}/ports`);
    if (Array.isArray(list)) {
      S.ports = list;
      if (S.view === 'net') R();
    }
  } catch (err) {
    console.error('Failed to load server ports:', err);
  }
}

function openAddPortModal() {
  const container = $('#modal-container');
  if (!container) return;

  const srv = curServer();
  const nextPort = getNextSuggestedPort();

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
    await api(`/api/servers/${srv.id}/ports`, {
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
  }
}

async function deletePort(portId, portNum) {
  const srv = curServer();
  if (!srv) return;

  if (!confirm(`Are you sure you want to delete port allocation ${portNum}?`)) {
    return;
  }

  try {
    await api(`/api/servers/${srv.id}/ports/${portId}`, { method: 'DELETE' });
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

  try {
    const res = await api(`/api/servers/${srv.id}/ports/${portId}/primary`, { method: 'POST' });
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
function openDeleteServerModal(srvId, srvName) {
  const container = $('#modal-container');
  if (!container) return;

  container.innerHTML = `<div class="pal" onclick="if(event.target===this)closeModal()">
    <div class="card pbox" style="padding:26px; max-width:520px; border:1px solid rgba(239, 68, 68, 0.4);">
      <div class="row sb" style="margin-bottom:14px;">
        <h2 style="font-size:20px; margin:0; color:#ef4444; display:flex; align-items:center; gap:8px;">
          ${ic('trash', 18)} Delete Server Instance
        </h2>
        <button class="btn sm" onclick="closeModal()" aria-label="Close modal">&times;</button>
      </div>

      <p style="font-size:13.5px; color:#e2e8f0; line-height:1.6; margin-bottom:14px;">
        You are about to permanently delete <b style="color:#fff;">${esc(srvName)}</b> (<code>${esc(srvId)}</code>).
      </p>

      <div style="background:rgba(239, 68, 68, 0.1); border:1px solid rgba(239, 68, 68, 0.3); border-radius:8px; padding:12px 14px; margin-bottom:16px; font-size:13px; color:#fca5a5; line-height:1.5;">
        <b>Warning:</b> This process is completely irreversible. All world data, player inventories, plugin configs, and assigned port allocations will be wiped immediately.
      </div>

      <label class="f" style="margin-bottom:20px;">
        <span style="font-size:13px; color:#cbd5e1;">To confirm, type <b style="color:#fff;">${esc(srvName)}</b> below:</span>
        <input type="text" id="del-srv-confirm" placeholder="Type server name to confirm" autocomplete="off" oninput="checkDeleteServerConfirm('${esc(srvName).replace(/'/g, "\\'")}')">
      </label>

      <div class="row" style="justify-content:flex-end; gap:10px;">
        <button type="button" class="btn" onclick="closeModal()">Cancel</button>
        <button type="button" id="del-srv-btn" class="btn dg" disabled onclick="executeDeleteServer('${srvId}')">
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

async function executeDeleteServer(srvId) {
  const btn = $('#del-srv-btn');
  if (btn) {
    btn.disabled = true;
    btn.textContent = 'Deleting server...';
  }

  try {
    if (S.ws) {
      try {
        S.ws.onclose = null;
        S.ws.close();
      } catch (e) {}
      S.ws = null;
    }

    await api(`/api/servers/${srvId}`, { method: 'DELETE' });
    toast('Server deleted successfully', 'ok');
    closeModal();

    await loadDashboard();
    S.cs = 0;
    go('dash');
  } catch (err) {
    toast(err.message || 'Failed to delete server', 'bad');
    if (btn) {
      btn.disabled = false;
      btn.textContent = 'Delete Server Permanently';
    }
  }
}

// --- Server Reinstallation (mcjarfiles.com) ---
function openReinstallModal(srvId) {
  const container = $('#modal-container');
  if (!container) return;

  const targetSrv = (S.servers || []).find(s => s.id === srvId) || curServer();
  if (!targetSrv) return;

  const targetId = targetSrv.id;
  const targetName = targetSrv.name || 'Server';
  const curType = (targetSrv.type || 'paper').toLowerCase();
  const curVer = targetSrv.version || '1.21.4';

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

      <form id="reinstall-form" onsubmit="submitReinstall(event, '${targetId}')">
        <label class="f" style="margin-bottom:12px;">
          <span>Minecraft Engine / Server Software</span>
          <select id="reinst-software" style="height:38px;">
            <option value="paper" ${curType === 'paper' ? 'selected' : ''}>Paper (High performance, Bukkit/Spigot plugins)</option>
            <option value="purpur" ${curType === 'purpur' ? 'selected' : ''}>Purpur (Ultra-configurable Paper fork)</option>
            <option value="leaf" ${curType === 'leaf' || curType === 'leafmc' ? 'selected' : ''}>LeafMC (Hyper-optimized Paper fork)</option>
            <option value="folia" ${curType === 'folia' ? 'selected' : ''}>Folia (Multi-threaded chunk engine)</option>
            <option value="fabric" ${curType === 'fabric' ? 'selected' : ''}>Fabric (Lightweight mod loader)</option>
            <option value="forge" ${curType === 'forge' ? 'selected' : ''}>Forge (Classic modding framework)</option>
            <option value="neoforge" ${curType === 'neoforge' ? 'selected' : ''}>NeoForge (Modern community mod loader)</option>
            <option value="vanilla" ${curType === 'vanilla' ? 'selected' : ''}>Vanilla (Official Mojang release)</option>
            <option value="velocity" ${curType === 'velocity' ? 'selected' : ''}>Velocity (High-performance modern proxy)</option>
          </select>
        </label>

        <label class="f" style="margin-bottom:16px;">
          <span>Version</span>
          <input type="text" id="reinst-version" value="${esc(curVer)}" placeholder="1.21.4, 1.20.4, or latest" required>
        </label>

        <div style="background:rgba(255,255,255,0.03); border:1px solid var(--border); border-radius:8px; padding:12px 14px; margin-bottom:20px;">
          <b style="font-size:13px; color:#fff; display:block; margin-bottom:8px;">Preservation Options</b>
          <label class="row" style="cursor:pointer; margin-bottom:8px; font-size:13px;">
            <input type="checkbox" id="reinst-world" checked style="width:auto; margin-right:8px; accent-color:var(--ac);">
            <span>Preserve Worlds (keep <code>world</code>, <code>world_nether</code>, <code>world_the_end</code>)</span>
          </label>
          <label class="row" style="cursor:pointer; margin-bottom:8px; font-size:13px;">
            <input type="checkbox" id="reinst-config" checked style="width:auto; margin-right:8px; accent-color:var(--ac);">
            <span>Preserve Configurations (keep <code>server.properties</code>, <code>bukkit.yml</code>, <code>spigot.yml</code>, <code>paper-*.yml</code>)</span>
          </label>
          <label class="row" style="cursor:pointer; font-size:13px;">
            <input type="checkbox" id="reinst-plugins" checked style="width:auto; margin-right:8px; accent-color:var(--ac);">
            <span>Preserve Plugins &amp; Mods (keep <code>plugins</code>, <code>mods</code> folders)</span>
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
}

async function submitReinstall(e, srvId) {
  e.preventDefault();
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

  toast(`Initiating reinstallation of ${software} ${version}...`);

  try {
    const res = await api(`/api/servers/${srvId}/reinstall`, {
      method: 'POST',
      body: JSON.stringify({
        software,
        version,
        preserve_world: preserveWorld,
        preserve_config: preserveConfig,
        preserve_plugins: preservePlugins
      })
    });

    closeModal();
    toast(`Server reinstalled successfully (${software} ${version})!`, 'ok');

    const cur = curServer();
    if (cur && cur.id === srvId) {
      cur.type = software;
      cur.version = version;
      S.ver = version;
      S.core = software.charAt(0).toUpperCase() + software.slice(1);
    }
    await checkServer();
    if (S.view === 'fi') loadFileList(S.currentDir);
    R();
  } catch (err) {
    toast(err.message || 'Server reinstallation failed', 'bad');
    if (btn) {
      btn.disabled = false;
      btn.innerHTML = `${ic('refresh', 14)} Start Reinstallation`;
    }
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
  }
}

async function deleteAdminUser(id, username) {
  if (!confirm(`Are you sure you want to delete user "${username}"?`)) return;

  try {
    await api(`/api/admin/users/${id}`, { method: 'DELETE' });
    toast(`User "${username}" deleted`, 'ok');
    await loadAdminData();
  } catch (err) {
    toast(err.message || 'Failed to delete user', 'bad');
  }
}

async function triggerAdminGC() {
  toast('Running Go runtime garbage collection...');
  try {
    const res = await api('/api/admin/gc', { method: 'POST' });
    toast(res.message || 'Garbage collection completed', 'ok');
    await loadAdminData();
  } catch (err) {
    toast(err.message || 'Failed to run GC', 'bad');
  }
}

function saveAdminDaemonSettings(e) {
  if (e) e.preventDefault();
  const brandName = $('#adm-brand')?.value.trim() || 'NeoPanel';
  const fqdn = $('#adm-fqdn')?.value.trim() || '127.0.0.1';
  const sftpPort = parseInt($('#adm-sftp')?.value, 10) || 2022;
  const maxUploadMB = parseInt($('#adm-upload')?.value, 10) || 500;
  const defaultRamMB = parseInt($('#adm-ram')?.value, 10) || 2048;
  const autoRestart = $('#adm-autorestart')?.checked ?? true;

  S.adminDaemonCfg = {
    brandName,
    fqdn,
    sftpPort,
    maxUploadMB,
    defaultRamMB,
    autoRestartPolicy: autoRestart
  };

  localStorage.setItem('neopanel_admin_cfg', JSON.stringify(S.adminDaemonCfg));
  document.title = `${brandName} — Minecraft Server Manager`;
  toast('Daemon & panel customization saved!', 'ok');
  R();
}

// --- Auth Flow & Setup Screen ---
async function checkAuth() {
  try {
    const status = await (await fetch('/api/auth/status')).json();
    if (status.needs_setup) {
      showSetupScreen();
      return;
    }

    if (!S.token) {
      showLoginScreen();
      return;
    }

    const user = await api('/api/auth/me');
    S.user = user;
    S.acct.un = user.username || S.acct.un;
    S.acct.em = user.email || S.acct.em;
    S.acct.h = await computeGravatarHash(S.acct.em);

    const savedMode = localStorage.getItem('neopanel_mode');
    if (savedMode) setAppearanceMode(savedMode);
    if (localStorage.getItem('neopanel_compact') === 'true') toggleCompact(true);
    if (localStorage.getItem('neopanel_priv') === 'true') togglePrivacy(true);

    await loadDashboard();

    // Respect URL hash if present
    const hash = window.location.hash.replace('#', '');
    if (hash && (V[hash] || hash === 'admin' || hash === 'acc')) {
      go(hash);
    }
  } catch (e) {
    showLoginScreen();
  }
}

function showSetupScreen() {
  const d = document.createElement('div');
  d.id = 'login';
  d.className = 'login';
  d.innerHTML = `<div class="lbox">
    <div class="llogo"><i class="cube"></i><b>Neo</b><span>Panel</span></div>
    <div class="lg">First-Time Initialization</div>
    <h1 class="lt">Create Admin Master Account</h1>
    <label class="lf">
      ${ic('user', 16)}
      <input id="su-name" type="text" placeholder="Username" autocomplete="username">
    </label>
    <label class="lf">
      ${ic('settings', 16)}
      <input id="su-pw" type="password" placeholder="Master Password" autocomplete="new-password">
    </label>
    <div class="lerr" id="lerr" role="alert"></div>
    <button class="lbtn" onclick="submitSetup()">Initialize &amp; Launch</button>
    <div class="lsm foot">&copy; 2026 NeoPanel &bull; Ultra-Lightweight</div>
  </div>`;
  document.body.appendChild(d);
}

async function submitSetup() {
  const username = $('#su-name')?.value.trim();
  const password = $('#su-pw')?.value;
  const err = $('#lerr');
  if (!username || !password) {
    if (err) err.textContent = 'Username and password are required.';
    return;
  }
  try {
    const res = await (await fetch('/api/auth/setup', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    })).json();

    if (res.token) {
      S.token = res.token;
      localStorage.setItem('neopanel_token', res.token);
      $('#login')?.remove();
      toast('Setup complete! Welcome to NeoPanel.');
      checkAuth();
    } else {
      if (err) err.textContent = res.error || 'Setup failed.';
    }
  } catch (e) {
    if (err) err.textContent = 'Connection error.';
  }
}

function showLoginScreen() {
  if ($('#login')) return;
  const hour = new Date().getHours();
  const greeting = hour < 5 ? 'Good evening' : hour < 12 ? 'Good morning' : hour < 18 ? 'Good afternoon' : 'Good evening';

  const d = document.createElement('div');
  d.id = 'login';
  d.className = 'login';
  d.innerHTML = `<div class="lbox">
    <div class="llogo"><i class="cube"></i><b>Neo</b><span>Panel</span></div>
    <div class="lg">${greeting}</div>
    <h1 class="lt">Log in to NeoPanel</h1>
    <label class="lf">
      ${ic('user', 16)}
      <input id="lu" type="text" placeholder="Username or email" autocomplete="username" onkeydown="if(event.key==='Enter')submitLogin()">
    </label>
    <label class="lf">
      ${ic('settings', 16)}
      <input id="lp" type="password" placeholder="Password" autocomplete="current-password" onkeydown="if(event.key==='Enter')submitLogin()">
      <button class="eye" type="button" onclick="const p=$('#lp'); p.type=p.type==='password'?'text':'password';" aria-label="Toggle password visibility">
        ${ic('activity', 16)}
      </button>
    </label>
    <div class="lerr" id="lerr" role="alert"></div>
    <button class="lbtn" onclick="submitLogin()">Login</button>
    <div class="lsm foot">&copy; 2026 NeoPanel &bull; Ultra-Lightweight</div>
  </div>`;
  document.body.appendChild(d);
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

  try {
    const res = await (await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: u, password: p })
    })).json();

    if (res.token) {
      S.token = res.token;
      localStorage.setItem('neopanel_token', res.token);
      $('#login')?.remove();
      toast('Signed in successfully');
      checkAuth();
    } else {
      if (err) err.textContent = res.error || 'Invalid credentials.';
    }
  } catch (e) {
    if (err) err.textContent = 'Connection error.';
  }
}

function logout() {
  S.token = '';
  localStorage.removeItem('neopanel_token');
  if (S.activeWs) {
    try { S.activeWs.close(); } catch (e) {}
    S.activeWs = null;
  }
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

// --- Initialize Application ---
document.addEventListener('DOMContentLoaded', () => {
  checkAuth();
  initDragAndDrop();
});
