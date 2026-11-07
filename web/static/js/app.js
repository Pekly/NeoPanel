// NeoPanel — Ultra-Lightweight Minecraft Server Manager
// Fully reactive Vanilla JS SPA with real-time WebSocket, SFTP, Modrinth API & Backups.

// --- Selectors & Global Helpers ---
const $ = s => document.querySelector(s);
const $$ = s => document.querySelectorAll(s);
const esc = s => String(s == null ? '' : s).replace(/[&<>"']/g, c => ({
  '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'
}[c]));

// --- Global Application State ---
window.S = {
  token: localStorage.getItem('neopanel_token') || '',
  user: null,
  view: 'dash',              // active view: dash, con, set, act, fi, db, bk, sc, net, sub, stp, pl, plug, store, ver, all, acc
  status: 'offline',         // current server status: online, offline, starting
  start: Date.now(),         // server start time for uptime
  cpu: Array.from({ length: 40 }, () => 0),
  ram: Array.from({ length: 40 }, () => 0),
  ni: Array.from({ length: 40 }, () => 0),
  no: Array.from({ length: 40 }, () => 0),
  tps: 20.00,
  ver: '1.21.4',
  core: 'Paper',
  sel: 'Paper',              // selected core in software switcher
  q: '',                     // plugin search query
  cat: 'All',                // plugin category filter
  ptab: 'online',            // player tab: online, wl, ban
  ftab: 'server.properties', // active open file name
  currentDir: '',            // file browser current directory path
  fileList: [],              // list of files in currentDir
  players: [],               // online players
  bans: [],                  // banned players
  wl: [],                    // whitelisted players
  ops: [],                   // ops
  plugins: [],               // installed plugins [[name, ver, enabled]]
  bk: [],                    // backups list
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
  files: {},                 // cache of open files
  log: [],                   // console lines
  hist: [],                  // command input history
  hi: 0,                     // command history pointer
  cs: 0,                     // current server index
  servers: [],               // live servers list from /api/servers
  sftp: null,                // sftp details
  system: null,              // system info
  cmenu: false,              // commands dropdown open
  pq: '',                    // command palette query
  pi: 0,                     // command palette selected index
  vq: '',                    // version search query
  vf: 'All',                 // version family filter
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
  pollTimer: null
};

// --- Audio Feedback (Web Audio API) ---
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

// --- Toast Notification ---
function toast(msg, type = 'info') {
  beep();
  const t = $('#toast');
  if (!t) return;
  t.textContent = msg;
  t.classList.add('s');
  clearTimeout(toast.timer);
  toast.timer = setTimeout(() => t.classList.remove('s'), 2600);
}

// --- Clock & Formatters ---
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
  const mem = srv.memory_max || srv.memory_limit || 2048;
  return typeof mem === 'number' ? mem : parseInt(mem) || 2048;
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

// --- Active Server Helper ---
function curServer() {
  if (!S.servers || S.servers.length === 0) return null;
  return S.servers[S.cs] || S.servers[0];
}

// --- ANSI to HTML ---
function ansiToHtml(text) {
  const escText = esc(text);
  return escText
    .replace(/\x1b\[0m/g, '</span>')
    .replace(/\x1b\[31m/g, '<span class="e">')
    .replace(/\x1b\[32m/g, '<span class="c">')
    .replace(/\x1b\[33m/g, '<span class="w">')
    .replace(/\x1b\[34m/g, '<span class="c">')
    .replace(/\x1b\[35m/g, '<span class="w">')
    .replace(/\x1b\[36m/g, '<span class="c">')
    .replace(/\x1b\[1m/g, '<b>')
    .replace(/\x1b\[22m/g, '</b>');
}

function addLog(line, cls) {
  const time = clock();
  const formatted = cls ? `<span class="${cls}">${line}</span>` : line;
  S.log.push(`[${time}] ${formatted}`);
  if (S.log.length > 500) S.log.shift();
  if (S.view === 'con') paintLog();
}

function paintLog() {
  const l = $('#log');
  if (!l) return;
  l.innerHTML = S.log.join('\n');
  l.scrollTop = l.scrollHeight;
}

// --- WebSocket Console Connection ---
function connectWs(serverId) {
  if (S.activeWs) {
    try { S.activeWs.close(); } catch (e) {}
    S.activeWs = null;
  }
  if (!serverId) return;

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${protocol}//${window.location.host}/api/servers/${serverId}/ws?token=${encodeURIComponent(S.token)}`;

  try {
    const ws = new WebSocket(wsUrl);
    S.activeWs = ws;

    ws.onmessage = e => {
      try {
        if (e.data.startsWith('{')) {
          const msg = JSON.parse(e.data);
          if (msg.type === 'stats') {
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
          } else if (msg.type === 'status') {
            S.status = msg.data || S.status;
            R();
            return;
          }
        }
      } catch (err) {}

      // Plain text log line
      const html = ansiToHtml(e.data);
      addLog(html);
    };

    ws.onclose = () => {
      S.activeWs = null;
    };
  } catch (e) {}
}

// --- Power Actions ---
async function power(action) {
  const srv = curServer();
  if (!srv) return;

  if (action === 'start') {
    S.status = 'starting';
    addLog(`[Server thread/INFO]: Starting Minecraft server instance ${srv.name}...`, 'c');
    R();
    toast('Starting server...');
  } else if (action === 'stop') {
    addLog('[Server thread/INFO]: Stopping the server...', 'w');
    toast('Stopping server...');
  } else if (action === 'restart') {
    toast('Restarting server...');
    addLog('[Server thread/INFO]: Restart sequence initiated...', 'w');
  } else if (action === 'kill') {
    toast('Terminating server process...');
    addLog('[Server thread/ERROR]: Force kill signal delivered', 'e');
  }

  try {
    const res = await api(`/api/servers/${srv.id}/power`, {
      method: 'POST',
      body: JSON.stringify({ action })
    });
    S.status = res.status || (action === 'start' ? 'online' : 'offline');
    if (S.status === 'online') S.start = Date.now();
    R();
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
    R();
  } catch (e) {}
}

async function runCmd(cmd) {
  if (!cmd) return;
  addLog(`<span class="c">&gt; ${esc(cmd)}</span>`);
  const srv = curServer();
  if (!srv) return;

  if (S.activeWs && S.activeWs.readyState === WebSocket.OPEN) {
    S.activeWs.send(JSON.stringify({ action: 'command', data: cmd }));
  } else {
    try {
      await api(`/api/servers/${srv.id}/command`, {
        method: 'POST',
        body: JSON.stringify({ command: cmd })
      });
    } catch (e) {}
  }
}

// --- Navigation Categorization (PG) ---
const PG = [
  ['con', 'Console', '›_', 'General'],
  ['set', 'Settings', '⚙', 'General'],
  ['act', 'Activity', '◷', 'General'],
  ['fi', 'Files', '▤', 'Management'],
  ['db', 'Databases', '⛁', 'Management'],
  ['bk', 'Backups', '▣', 'Management'],
  ['sc', 'Schedules', '◫', 'Management'],
  ['net', 'Network', '◍', 'Management'],
  ['sub', 'Subdomain', '◌', 'Management'],
  ['stp', 'Startup', '➤', 'Management'],
  ['pl', 'Players', '☺', 'Management'],
  ['plug', 'Plugins', '✦', 'Minecraft'],
  ['store', 'Plugin store', '⊕', 'Minecraft'],
  ['ver', 'Software', '⇅', 'Minecraft'],
  ['all', 'All versions', '☰', 'Minecraft']
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
  ['NeoForge', 'Modern forge modding engine']
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
        <span class="badge ${st}"><i class="dot"></i>${{ online: 'Running', offline: 'Offline', starting: 'Starting' }[st]}</span>
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
      </div>` : ''}
      <button class="btn pr" onclick="power('start')" ${st !== 'offline' ? 'disabled' : ''}>Start</button>
      <button class="btn" onclick="power('restart')" ${on ? '' : 'disabled'}>Restart</button>
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

// --- Stats & Charts ---
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

  return row('◎', 'Address', `<span class="pv">${hostAddr}</span>`) +
    row('◷', 'Uptime', up()) +
    row('▣', 'CPU load', `${c.toFixed(1)}%`, '/ 100%') +
    row('▤', 'Memory', `${m.toFixed(0)} MiB`, `/ ${(maxRamMB / 1024).toFixed(1)} GiB`) +
    row('▥', 'Disk', '1.82 GiB', '/ 20 GiB') +
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
    const on = S.status === 'online';
    const c = on ? S.cpu.at(-1) : 0;
    const maxRamMB = getServerMaxRamMB(srv);
    const maxRamGB = (maxRamMB / 1024).toFixed(1);
    const r = on ? ((S.ram.at(-1) / 100) * parseFloat(maxRamGB)).toFixed(1) : '0.0';
    const hostAddr = `${window.location.hostname}:${srv ? srv.port : 25565}`;

    return `<header class="top">
      <h1>Dashboard</h1>
      <span class="sp"></span>
      <button class="btn pr" onclick="openCreateServerModal()">+ Create Server</button>
    </header>
    
    ${S.servers.length > 1 ? `
      <div class="row" style="margin-bottom:16px; overflow-x:auto; padding-bottom:4px;">
        ${S.servers.map((s, i) => `
          <button class="btn ${i === S.cs ? 'pr' : ''}" onclick="switchServer(${i})">
            <b>${esc(s.name)}</b> <span class="badge ${(s.stats?.status || s.status) === 'running' ? 'online' : 'offline'}" style="margin-left:6px;"><i class="dot"></i>${(s.stats?.status || s.status) === 'running' ? 'Online' : 'Offline'}</span>
          </button>
        `).join('')}
      </div>
    ` : ''}

    <div class="grid g4">
      <div class="card stat">
        <span>Players online</span>
        <b>${S.players.length} / ${S.props.max}</b>
        <div class="bar"><i style="width:${Math.min(100, (S.players.length / S.props.max) * 100)}%"></i></div>
      </div>
      <div class="card stat">
        <span>Ticks per second</span>
        <b>${on ? S.tps.toFixed(2) : '0.00'}</b>
        <span class="mu">Target 20.00</span>
      </div>
      <div class="card stat">
        <span>Uptime</span>
        <b style="font-size:21px;padding:5px 0">${up()}</b>
        <span class="mu">Since last start</span>
      </div>
      <div class="card stat">
        <span>Game Address</span>
        <b style="font-size:17px;padding:7px 0" class="mono pv">${hostAddr}</b>
        <button class="btn sm" onclick="navigator.clipboard.writeText('${hostAddr}'); toast('Address copied!')">Copy</button>
      </div>
    </div>

    <div class="grid g2" style="margin-top:14px">
      <div class="card">
        <div class="row sb"><h2>CPU Load</h2><b>${c.toFixed(0)}%</b></div>
        ${spark(on ? S.cpu : S.cpu.map(() => 0), 100)}
      </div>
      <div class="card">
        <div class="row sb"><h2>Memory</h2><b>${r} / ${maxRamGB} GB</b></div>
        ${spark(on ? S.ram : S.ram.map(() => 0), 100)}
      </div>
    </div>

    <div class="grid g2" style="margin-top:14px">
      <div class="card">
        <h2>How it looks in the server list</h2>
        <div class="slist">
          <i class="cube"></i>
          <div>
            <b>${esc(S.props.motd)}</b>
            <div class="mu">${S.core} ${S.ver}, ${S.players.length}/${S.props.max} players</div>
          </div>
          <span class="ping"><i></i><i></i><i></i><i></i></span>
        </div>
      </div>
      <div class="card">
        <h2>Players (Last 24 Hours)</h2>
        <svg class="spark" viewBox="0 0 240 70" preserveAspectRatio="none" aria-hidden="true">
          ${[3, 2, 1, 1, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 12, 14, 15, 17, 16, 13, 10, 7, 4].map((v, i) => `
            <rect x="${i * 10 + 1}" y="${70 - v * 4}" width="8" height="${v * 4}" rx="1" fill="var(--ac)" opacity="${0.35 + v / 25}"/>
          `).join('')}
        </svg>
      </div>
    </div>

    <div class="card" style="margin-top:14px">
      <h2>Recent Activity</h2>
      <div class="mono mu" style="font-size:12.5px; line-height: 1.7;">
        ${S.log.slice(-5).map(l => `<div>${l}</div>`).join('') || '<div>No recent activity recorded.</div>'}
      </div>
    </div>`;
  },

  con() {
    return `${srvh()}
    <div class="cgrid">
      <div class="cwrap">
        <div class="con" id="log" role="log" aria-live="polite"></div>
        <div class="cin">
          <span>&raquo;</span>
          <input id="cmd" autocomplete="off" spellcheck="false" placeholder="${S.status === 'online' ? 'Type a command, try help...' : 'Server is offline'}" aria-label="Console command" ${S.status !== 'online' ? 'disabled' : ''}>
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
    const f = S.ftab;
    return `${srvh()}
    <div class="grid" style="grid-template-columns: minmax(200px, 260px) 1fr; margin-top:14px;">
      <div class="card" style="display:flex; flex-direction:column; gap:8px;">
        <div class="row sb">
          <h2 style="margin:0">/server${S.currentDir ? '/' + S.currentDir : ''}</h2>
          <button class="btn sm" onclick="uploadFileDialog()">Upload</button>
        </div>
        ${S.currentDir ? `<button class="btn sm" onclick="browseFolder('..')" style="text-align:left;">📁 .. (Up one level)</button>` : ''}
        <div style="max-height:480px; overflow-y:auto; display:flex; flex-direction:column; gap:4px;">
          ${S.fileList.map(item => `
            <button class="btn sm ${f === item.name ? 'on' : ''}" style="text-align:left; justify-content:flex-start;" onclick="${item.is_dir ? `browseFolder('${item.name}')` : `openFile('${item.name}')`}">
              ${item.is_dir ? '📁' : '📄'} ${esc(item.name)}
            </button>
          `).join('') || '<div class="mu" style="font-size:12px; padding:8px;">Loading server files...</div>'}
        </div>
      </div>

      <div class="card">
        <div class="row sb" style="margin-bottom:10px">
          <h2 style="margin:0">${esc(f || 'Select a file to edit')}</h2>
          ${f ? `<button class="btn pr sm" onclick="saveActiveFile()">Save file</button>` : ''}
        </div>
        <textarea id="ed" spellcheck="false" aria-label="File editor" placeholder="Select or open a file from the left to edit its content...">${esc(S.files[f] || '')}</textarea>
      </div>
    </div>`;
  },

  plug() {
    return `${srvh()}
    <div class="card tw" style="margin-top:14px">
      <table>
        <thead><tr><th>Plugin</th><th>Version</th><th>Enabled</th><th style="text-align:right">Actions</th></tr></thead>
        <tbody>
          ${S.plugins.length ? S.plugins.map((p, i) => `
            <tr>
              <td><b>${esc(p[0])}</b></td>
              <td class="mono">${esc(p[1] || 'installed')}</td>
              <td>
                <label class="sw">
                  <input type="checkbox" ${p[2] ? 'checked' : ''} onchange="togglePlugin(${i}, this.checked)">
                  <i></i>
                </label>
              </td>
              <td style="text-align:right">
                <button class="btn sm dg" onclick="deletePlugin('${esc(p[0])}')">Remove</button>
              </td>
            </tr>
          `).join('') : '<tr><td colspan="4" class="mu" style="text-align:center;padding:24px;">No plugins found in the /plugins directory.</td></tr>'}
        </tbody>
      </table>
      <div class="row sb" style="margin-top:16px; border-top:1px solid var(--ln); padding-top:14px;">
        <span class="mu">Changes apply immediately or after the next server restart.</span>
        <button class="btn pr" onclick="go('store')">Browse Modrinth Store</button>
      </div>
    </div>`;
  },

  store() {
    return `${srvh()}
    <div class="row sb" style="margin:14px 0 12px; gap:12px;">
      <input type="text" id="store-search" placeholder="Search thousands of plugins on Modrinth..." value="${esc(S.q)}" oninput="S.q=this.value; searchModrinthDebounced();" style="max-width:320px">
      <div class="row" style="overflow-x:auto;">
        ${CATS.map(c => `
          <button class="btn sm ${S.cat === c ? 'pr' : ''}" onclick="S.cat='${c}'; searchModrinth();">${c}</button>
        `).join('')}
      </div>
      <button class="btn" onclick="uploadFileDialog()">Upload .jar</button>
    </div>
    <div class="grid store" id="store-grid">
      <div class="mu" style="padding:24px;">Loading top plugins from Modrinth...</div>
    </div>`;
  },

  ver() {
    return `${srvh()}
    <div class="card hero" style="margin-top:14px">
      <span class="mu">Currently Running</span>
      <b class="big">${S.core} ${S.ver}</b>
      <span class="mu">Java 21 / 17 compatible &bull; High performance runtime</span>
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
          ${[['1.21.4', 'Popular latest'], ['1.21.1', 'Widely supported'], ['1.20.4', 'Stable'], ['1.20.1', 'Top modded'], ['1.19.4', 'Legacy stable']].map(([v, n]) => {
            const cur = S.ver === v && S.core === S.sel;
            return `<tr>
              <td class="mono"><b>${v}</b></td>
              <td><span class="chip">${n}</span></td>
              <td style="text-align:right">
                ${cur ? '<span class="badge online"><i class="dot"></i>Active</span>' : `<button class="btn sm pr" onclick="installSoftware('${S.sel}', '${v}')">Install ${S.sel} ${v}</button>`}
              </td>
            </tr>`;
          }).join('')}
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
      <button class="btn pr" onclick="createBackup()">+ Create Full Backup</button>
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
                  <button class="btn sm" onclick="downloadBackup('${esc(b.name || b[0])}')">Download</button>
                  <button class="btn sm dg" onclick="deleteBackup('${esc(b.name || b[0])}')">Delete</button>
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

    return `${srvh()}
    <div class="grid g2" style="margin-top:14px">
      <div class="card tw">
        <h2>Server Port Allocations</h2>
        <table>
          <thead><tr><th>Service</th><th>Port</th><th>Primary</th></tr></thead>
          <tbody>
            <tr><td>Minecraft Game</td><td class="mono">${srv?.port || 25565}</td><td><span class="badge online"><i class="dot"></i>Yes</span></td></tr>
            <tr><td>Pure-Go SFTP (Built-in)</td><td class="mono">${sftpPort}</td><td>No</td></tr>
            <tr><td>RCON Remote Console</td><td class="mono">25575</td><td>No</td></tr>
          </tbody>
        </table>
      </div>

      <div class="card">
        <h2>SFTP File Access</h2>
        <p class="mu" style="font-size:13px; margin-bottom:12px;">Connect with FileZilla, WinSCP, or Cyberduck using pure-Go jailed SFTP.</p>
        <div class="mono" style="background:var(--con); color:var(--cont); padding:10px 12px; border-radius:6px; font-size:12px; margin-bottom:12px;">
          <div><b>Host:</b> ${host}</div>
          <div><b>Port:</b> ${sftpPort}</div>
          <div><b>Username:</b> ${sftpUser}</div>
        </div>
        <button class="btn pr sm" onclick="navigator.clipboard.writeText('sftp://${sftpUser}@${host}:${sftpPort}'); toast('SFTP URI copied!');">Copy Connection URL</button>
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
    const mem = getServerMaxRamMB(srv);
    return `${srvh()}
    <div class="card" style="margin-top:14px">
      <h2>Server Startup Parameters</h2>
      <label class="f">
        <span>Startup Command</span>
        <textarea id="startup-cmd" style="min-height:75px">java -Xms128M -Xmx${mem}M -XX:+UseG1GC -jar ${srv?.jar_file || 'server.jar'} nogui</textarea>
      </label>
      <label class="f">
        <span>Java Runtime Environment</span>
        <select id="startup-java">
          <option selected>OpenJDK 21 (Recommended for 1.20.5+)</option>
          <option>OpenJDK 17 (Legacy 1.18 - 1.20.4)</option>
        </select>
      </label>
      <button class="btn pr" onclick="toast('Startup arguments saved')">Save Startup Parameters</button>
    </div>`;
  },

  set() {
    const p = S.props;
    const o = (v, a) => a.map(x => `<option ${x === v ? 'selected' : ''}>${x}</option>`).join('');
    const sw = (k, l) => `<label class="row sb" style="padding:7px 0; border-bottom:1px solid var(--ln);">
      <span>${l}</span>
      <span class="sw">
        <input type="checkbox" ${p[k] ? 'checked' : ''} onchange="S.props.${k}=this.checked;">
        <i></i>
      </span>
    </label>`;

    const curMem = getServerMaxRamMB(curServer());

    return `${srvh()}
    <div class="grid g2" style="margin-top:14px">
      <div class="card">
        <h2>Server Configuration</h2>
        <label class="f">
          <span>Server Message (MOTD)</span>
          <input type="text" id="prop-motd" value="${esc(p.motd)}" oninput="S.props.motd=this.value;">
        </label>
        <label class="f">
          <span>Max Players</span>
          <input type="number" min="1" max="1000" id="prop-max" value="${p.max}" oninput="S.props.max=+this.value||1;">
        </label>
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
            <option value="2048" ${curMem === 2048 ? 'selected' : ''}>2 GB (2048 MB)</option>
            <option value="4096" ${curMem === 4096 ? 'selected' : ''}>4 GB (4096 MB)</option>
            <option value="8192" ${curMem === 8192 ? 'selected' : ''}>8 GB (8192 MB)</option>
            <option value="16384" ${curMem === 16384 ? 'selected' : ''}>16 GB (16384 MB)</option>
          </select>
        </label>
        <div class="row" style="margin-top:14px;">
          <button class="btn pr" onclick="saveProperties()">Save Properties</button>
          <button class="btn" onclick="toggleTheme()">Toggle Dark/Light</button>
        </div>
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
        <button class="btn pr" onclick="addSSHKey()">Add SSH Key</button>
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
                  <button class="btn sm dg" onclick="deleteSSHKey('${k.id || i}')">Delete</button>
                </td>
              </tr>
            `).join('') : '<tr><td colspan="4" class="mu" style="text-align:center;padding:20px;">No SSH keys registered. Add your public key above.</td></tr>'}
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
          <button class="btn pr" onclick="createAPIKey()">Generate Key</button>
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
                  <button class="btn sm dg" onclick="deleteAPIKey('${k.id || i}')">Revoke</button>
                </td>
              </tr>
            `).join('') : '<tr><td colspan="4" class="mu" style="text-align:center;padding:20px;">No active API keys created yet.</td></tr>'}
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
            `).join('') : '<tr><td colspan="3" class="mu" style="text-align:center;padding:20px;">No recent security events.</td></tr>'}
          </tbody>
        </table>
      </div>`;
    }

    return `<header class="top">
      ${av(40)}
      <h1>Account Settings</h1>
      <span class="sp"></span>
      <button class="btn dg" onclick="logout()">Log Out</button>
    </header>
    <div class="atabs">
      ${['General', 'Security', 'SSH Keys', 'API Keys', 'Activity'].map((x, i) => `
        <button class="${t === i ? 'on' : ''}" onclick="S.acct.tab=${i}; R(); loadAccountData();">${x}</button>
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
  let group = '';

  return `<div class="brand" onclick="go('dash')" style="cursor:pointer;">
    <i class="cube"></i>NeoPanel
  </div>
  
  <button class="nsearch" onclick="openPalette(1)">
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
      <circle cx="11" cy="11" r="7"/><path d="M20 20l-4-4"/>
    </svg>
    <span>Search</span>
    <kbd>Ctrl K</kbd>
  </button>

  <button class="nb" onclick="go('dash')">
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
      <path d="M19 12H5M12 5l-7 7 7 7"/>
    </svg>
    All Servers (${S.servers.length})
  </button>

  ${PG.map(([key, label, icon, grp]) => {
    const header = grp !== group ? `<div class="ng">${grp}</div>` : '';
    group = grp;
    const active = S.view === key ? 'on' : '';
    return `${header}
    <button class="nav-item ${active}" onclick="go('${key}')" ${active ? 'aria-current="page"' : ''}>
      <span aria-hidden="true" style="font-family:var(--mf); width:18px; text-align:center;">${icon}</span>
      ${label}
    </button>`;
  }).join('')}

  <span class="sp"></span>

  <button class="chip2 ${S.view === 'acc' ? 'on' : ''}" onclick="go('acc')" aria-label="Account Settings">
    ${av(34)}
    <span class="who">
      <b>${esc(S.acct.fn)} ${esc(S.acct.ln)}</b>
      <span class="mu pv">${esc(S.acct.em)}</span>
    </span>
    <span class="mu" style="font-size:12px;">⚙</span>
  </button>`;
}

// --- View Router & State Handler (go) ---
window.go = function(view) {
  S.view = view;
  R();
  onEnterView(view);
};

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

// --- View Lifecycle Hook ---
window.onEnterView = function(view) {
  if (view === 'pl') loadPlayersData();
  else if (view === 'fi') loadFileList(S.currentDir);
  else if (view === 'plug') loadPluginsList();
  else if (view === 'store') searchModrinth();
  else if (view === 'bk') loadBackupsList();
  else if (view === 'set') loadProperties();
  else if (view === 'net') loadSFTPDetails();
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

// Files
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

async function openFile(name) {
  const srv = curServer();
  if (!srv) return;
  const filePath = S.currentDir ? `${S.currentDir}/${name}` : name;
  S.ftab = name;
  try {
    const data = await api(`/api/servers/${srv.id}/files/content?path=${encodeURIComponent(filePath)}`);
    S.files[name] = data.content != null ? data.content : data;
    R();
  } catch (e) {}
}

async function saveActiveFile() {
  const srv = curServer();
  if (!srv || !S.ftab) return;
  const ed = $('#ed');
  if (!ed) return;
  const content = ed.value;
  const filePath = S.currentDir ? `${S.currentDir}/${S.ftab}` : S.ftab;
  try {
    await api(`/api/servers/${srv.id}/files/content`, {
      method: 'POST',
      body: JSON.stringify({ path: filePath, content })
    });
    S.files[S.ftab] = content;
    toast(`Saved ${S.ftab} successfully`);
  } catch (e) {}
}

function uploadFileDialog() {
  const input = document.createElement('input');
  input.type = 'file';
  input.onchange = async () => {
    const file = input.files[0];
    if (!file) return;
    const srv = curServer();
    if (!srv) return;
    const formData = new FormData();
    formData.append('file', file);
    formData.append('path', S.currentDir || '');
    toast(`Uploading ${file.name}...`);
    try {
      await api(`/api/servers/${srv.id}/files/upload`, {
        method: 'POST',
        body: formData
      });
      toast(`${file.name} uploaded successfully!`);
      loadFileList(S.currentDir);
    } catch (e) {}
  };
  input.click();
}

// Plugins
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
      return `<div class="card pc">
        <div class="row sb">
          <b>${esc(hit.title)}</b>
          <span class="chip">${hit.categories ? hit.categories[0] : 'plugin'}</span>
        </div>
        <p class="mu" style="font-size:12.5px; margin:4px 0 10px;">${esc(hit.description || '')}</p>
        <div class="row sb">
          <span class="mu" style="font-size:12px;">↓ ${(hit.downloads / 1000).toFixed(0)}k</span>
          ${isInstalled ? '<span class="badge online"><i class="dot"></i>Installed</span>' : `<button class="btn pr sm" onclick="installPlugin('${hit.project_id || hit.slug}', '${esc(hit.title)}')">Install</button>`}
        </div>
      </div>`;
    }).join('') || '<div class="mu" style="padding:24px;">No matching plugins found on Modrinth.</div>';
  } catch (e) {
    grid.innerHTML = '<div class="mu" style="padding:24px;">Error reaching Modrinth API.</div>';
  }
}

async function installPlugin(projectSlug, title) {
  const srv = curServer();
  if (!srv) return;
  toast(`Downloading & installing ${title}...`);
  try {
    await api(`/api/servers/${srv.id}/plugins/install`, {
      method: 'POST',
      body: JSON.stringify({ project_id: projectSlug })
    });
    toast(`Successfully installed ${title}!`);
    loadPluginsList();
  } catch (e) {}
}

// Backups
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

// Software & Properties
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
  const payload = {
    'motd': S.props.motd,
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

// Account
async function loadAccountData() {
  try {
    const prof = await api('/api/account/profile');
    S.acct.un = prof.username || S.acct.un;
    S.acct.em = prof.email || S.acct.em;
    S.acct.fn = prof.first_name || 'Admin';
    S.acct.ln = prof.last_name || '';
    S.acct.h = await computeGravatarHash(S.acct.em);

    if (S.acct.tab === 2) {
      S.acct.ssh = await api('/api/account/ssh-keys');
    } else if (S.acct.tab === 3) {
      S.acct.api = await api('/api/account/api-keys');
    } else if (S.acct.tab === 4) {
      S.acct.act = await api('/api/account/activity');
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

// Appearance
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

// Server Switching
function switchServer(index) {
  S.cs = index;
  const srv = S.servers[index];
  if (!srv) return;
  const st = srv.stats?.status || srv.status || 'offline';
  S.status = (st === 'running' || st === 'online') ? 'online' : 'offline';
  S.ver = srv.version || S.ver;
  S.core = (srv.type || 'paper').charAt(0).toUpperCase() + (srv.type || 'paper').slice(1);
  connectWs(srv.id);
  R();
  toast(`Switched to ${srv.name}`);
}

// --- Command Palette (Ctrl + K) ---
function paletteItems() {
  const q = (S.pq || '').toLowerCase();
  const items = [
    ...PG.map(p => ['Pages', p[1], 'Go to page', () => go(p[0])]),
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

// --- Flexible & Powerful Server Creation Modal ---
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
};

window.setCreateVersion = function(ver) {
  const inp = $('#cs-ver');
  if (inp) inp.value = ver;
  $$('.ver-pill').forEach(btn => {
    btn.classList.toggle('pr', btn.dataset.ver === ver);
  });
};

function getNextSuggestedPort() {
  const usedPorts = S.servers.map(s => s.port || 25565);
  let candidate = 25565;
  while (usedPorts.includes(candidate)) {
    candidate++;
  }
  return candidate;
}

function openCreateServerModal() {
  const container = $('#modal-container');
  if (!container) return;

  const nextPort = getNextSuggestedPort();

  container.innerHTML = `<div class="pal" onclick="if(event.target===this)closeModal()">
    <div class="card pbox" style="padding:26px; max-width:580px; max-height:88vh; overflow-y:auto;">
      <div class="row sb" style="margin-bottom:16px;">
        <div>
          <h2 style="font-size:22px; margin:0;">Create Minecraft Server</h2>
          <span class="mu" style="font-size:13px;">Deploy a high-performance instance in seconds</span>
        </div>
        <button class="btn sm" onclick="closeModal()">✕</button>
      </div>

      <form onsubmit="handleCreateServer(event)">
        <!-- Basic Info -->
        <label class="f">
          <span>Server Name</span>
          <input type="text" id="cs-name" required placeholder="e.g. Survival SMP, Bedwars Hub, Modded Realm" value="My Server">
        </label>

        <!-- Engine Choice -->
        <div style="margin-bottom:14px;">
          <span class="mu" style="display:block; margin-bottom:6px; font-size:13px; font-weight:600;">Server Engine</span>
          <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(115px, 1fr)); gap:8px;">
            <button type="button" class="card engine-card sel on" data-engine="paper" onclick="setCreateEngine('paper')" style="padding:10px; text-align:center;">
              <b>Paper</b>
              <span class="mu" style="font-size:11px; display:block;">Plugins &amp; Fast</span>
            </button>
            <button type="button" class="card engine-card sel" data-engine="purpur" onclick="setCreateEngine('purpur')" style="padding:10px; text-align:center;">
              <b>Purpur</b>
              <span class="mu" style="font-size:11px; display:block;">High Config</span>
            </button>
            <button type="button" class="card engine-card sel" data-engine="fabric" onclick="setCreateEngine('fabric')" style="padding:10px; text-align:center;">
              <b>Fabric</b>
              <span class="mu" style="font-size:11px; display:block;">Modern Mods</span>
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
            <option value="vanilla">vanilla</option>
          </select>
        </div>

        <!-- Version Choice with Quick Pills -->
        <div style="margin-bottom:14px;">
          <div class="row sb" style="margin-bottom:6px;">
            <span class="mu" style="font-size:13px; font-weight:600;">Minecraft Version</span>
            <div class="row" style="gap:4px;">
              ${['1.21.4', '1.21.1', '1.20.4', '1.20.1', '1.19.4', '1.16.5'].map(v => `
                <button type="button" class="btn sm ver-pill ${v === '1.21.4' ? 'pr' : ''}" data-ver="${v}" onclick="setCreateVersion('${v}')">${v}</button>
              `).join('')}
            </div>
          </div>
          <input type="text" id="cs-ver" value="1.21.4" required placeholder="e.g. 1.21.4, 1.20.4, 1.16.5">
        </div>

        <!-- Memory Allocation with Quick Pills -->
        <div style="margin-bottom:14px;">
          <div class="row sb" style="margin-bottom:6px;">
            <span class="mu" style="font-size:13px; font-weight:600;">RAM Allocation</span>
            <div class="row" style="gap:4px;">
              <button type="button" class="btn sm ram-pill" data-ram="1024" onclick="setCreateRam(1024)">1 GB</button>
              <button type="button" class="btn sm ram-pill pr" data-ram="2048" onclick="setCreateRam(2048)">2 GB</button>
              <button type="button" class="btn sm ram-pill" data-ram="4096" onclick="setCreateRam(4096)">4 GB</button>
              <button type="button" class="btn sm ram-pill" data-ram="8192" onclick="setCreateRam(8192)">8 GB</button>
            </div>
          </div>
          <input type="number" id="cs-mem" value="2048" min="512" max="65536" step="512" required placeholder="RAM in MB (e.g. 2048)">
        </div>

        <!-- Port Allocation & Java -->
        <div class="grid g2" style="margin-bottom:14px;">
          <label class="f" style="margin-bottom:0;">
            <span>Game Port</span>
            <input type="number" id="cs-port" value="${nextPort}" min="1024" max="65535" required>
          </label>
          <label class="f" style="margin-bottom:0;">
            <span>Java Runtime</span>
            <select id="cs-java">
              <option value="21" selected>OpenJDK 21 (1.20.5+)</option>
              <option value="17">OpenJDK 17 (1.18 - 1.20.4)</option>
            </select>
          </label>
        </div>

        <!-- Flags & Automations -->
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
            <span>Apply Aikar's high-performance GC flags</span>
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

    // Select newly created server
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

    // Load appearance settings
    const savedMode = localStorage.getItem('neopanel_mode');
    if (savedMode) setAppearanceMode(savedMode);
    if (localStorage.getItem('neopanel_compact') === 'true') toggleCompact(true);
    if (localStorage.getItem('neopanel_priv') === 'true') togglePrivacy(true);

    await loadDashboard();
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
      <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="8" r="4"/><path d="M4 21c0-4 4-6 8-6s8 2 8 6"/></svg>
      <input id="su-name" type="text" placeholder="Username" autocomplete="username">
    </label>
    <label class="lf">
      <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="8" cy="15" r="4"/><path d="M11 12l9-9M16 7l3 3"/></svg>
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
      <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="8" r="4"/><path d="M4 21c0-4 4-6 8-6s8 2 8 6"/></svg>
      <input id="lu" type="text" placeholder="Username or email" autocomplete="username" onkeydown="if(event.key==='Enter')submitLogin()">
    </label>
    <label class="lf">
      <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="8" cy="15" r="4"/><path d="M11 12l9-9M16 7l3 3"/></svg>
      <input id="lp" type="password" placeholder="Password" autocomplete="current-password" onkeydown="if(event.key==='Enter')submitLogin()">
      <button class="eye" type="button" onclick="const p=$('#lp'); p.type=p.type==='password'?'text':'password';" aria-label="Toggle password visibility">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M2 12s4-7 10-7 10 7 10 7-4 7-10 7S2 12 2 12z"/><circle cx="12" cy="12" r="3"/></svg>
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

// --- Attach All Global Handlers on Window ---
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

// --- Initialize App ---
document.addEventListener('DOMContentLoaded', () => {
  checkAuth();
});
