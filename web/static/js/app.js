// NeoPanel - Core SPA Controller
const state = {
  token: localStorage.getItem('neopanel_token') || '',
  user: null,
  servers: [],
  selectedServer: null,
  currentTab: 'console',
  activeWs: null,
  consoleMgr: null,
  currentFilePath: '',
  systemInfo: null,
  reconnectAttempts: 0,
  reconnectTimeout: null,
  activeCategory: '',
};

// --- Toast Notifications ---
function showToast(message, type = 'success') {
  const container = document.getElementById('toast-container');
  const toast = document.createElement('div');
  toast.className = `toast ${type === 'error' ? 'error' : ''}`;
  
  const iconSvg = type === 'error' 
    ? `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="color:var(--danger); flex-shrink:0;"><circle cx="12" cy="12" r="10"></circle><line x1="15" y1="9" x2="9" y2="15"></line><line x1="9" y1="9" x2="15" y2="15"></line></svg>`
    : `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="color:var(--primary); flex-shrink:0;"><polyline points="20 6 9 17 4 12"></polyline></svg>`;

  toast.innerHTML = `${iconSvg}<span>${escapeHtml(message)}</span>`;
  container.appendChild(toast);
  setTimeout(() => {
    toast.style.opacity = '0';
    toast.style.transform = 'translateX(100%)';
    toast.style.transition = 'all 0.25s ease';
    setTimeout(() => toast.remove(), 250);
  }, 4000);
}

// --- API Helper ---
async function api(endpoint, options = {}) {
  const headers = options.headers || {};
  if (state.token) {
    headers['Authorization'] = `Bearer ${state.token}`;
  }
  if (!headers['Content-Type'] && !(options.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
  }

  try {
    const res = await fetch(endpoint, { ...options, headers });
    if (res.status === 401) {
      logout();
      throw new Error('Session expired. Please log in again.');
    }
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      throw new Error(data.error || `HTTP error ${res.status}`);
    }
    return data;
  } catch (err) {
    showToast(err.message, 'error');
    throw err;
  }
}

// --- Auth Flow ---
async function checkAuth() {
  try {
    const status = await (await fetch('/api/auth/status')).json();
    if (status.needs_setup) {
      showView('setup-view');
      return;
    }

    if (!state.token) {
      showView('login-view');
      return;
    }

    // Verify current user
    const user = await api('/api/auth/me');
    state.user = user;
    updateUserNav(user.username);
    showView('dashboard-view');
    loadDashboard();
  } catch (e) {
    showView('login-view');
  }
}

function updateUserNav(username) {
  document.getElementById('nav-user').textContent = username;
  const avatar = document.getElementById('nav-user-avatar');
  if (avatar) {
    avatar.textContent = username.charAt(0).toUpperCase();
  }
}

async function handleSetup(e) {
  e.preventDefault();
  const username = document.getElementById('setup-username').value;
  const password = document.getElementById('setup-password').value;

  try {
    const res = await api('/api/auth/setup', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    });
    state.token = res.token;
    state.user = res.user;
    localStorage.setItem('neopanel_token', res.token);
    showToast('Admin setup completed!');
    updateUserNav(res.user.username);
    showView('dashboard-view');
    loadDashboard();
  } catch (e) {}
}

async function handleLogin(e) {
  e.preventDefault();
  const username = document.getElementById('login-username').value;
  const password = document.getElementById('login-password').value;

  try {
    const res = await api('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    });
    state.token = res.token;
    state.user = res.user;
    localStorage.setItem('neopanel_token', res.token);
    showToast(`Welcome back, ${res.user.username}!`);
    updateUserNav(res.user.username);
    showView('dashboard-view');
    loadDashboard();
  } catch (e) {}
}

function logout() {
  state.token = '';
  state.user = null;
  localStorage.removeItem('neopanel_token');
  if (state.activeWs) {
    state.activeWs.close();
    state.activeWs = null;
  }
  showView('login-view');
}

// --- View Router ---
function showView(viewId) {
  ['setup-view', 'login-view', 'dashboard-view', 'server-detail-view'].forEach((id) => {
    const el = document.getElementById(id);
    if (el) el.style.display = id === viewId ? 'block' : 'none';
  });

  const navAuth = document.getElementById('nav-auth');
  if (navAuth) {
    navAuth.style.display = (viewId === 'dashboard-view' || viewId === 'server-detail-view') ? 'flex' : 'none';
  }
}

// --- Dashboard ---
async function loadDashboard() {
  try {
    const [servers, sys] = await Promise.all([
      api('/api/servers'),
      api('/api/system'),
    ]);
    state.servers = servers;
    state.systemInfo = sys;

    // Render metrics
    document.getElementById('total-servers-count').textContent = servers.length;
    let onlineCount = 0;
    servers.forEach(s => {
      if (s.stats && (s.stats.status === 'online' || s.stats.status === 'starting')) onlineCount++;
    });
    document.getElementById('active-servers-count').textContent = onlineCount;
    document.getElementById('system-ram-metric').textContent = `${sys.free_ram_mb} MB / ${sys.total_ram_mb} MB`;

    // Render server list cards
    const grid = document.getElementById('servers-grid');
    grid.innerHTML = '';

    if (servers.length === 0) {
      grid.innerHTML = `
        <div style="grid-column: 1/-1; text-align: center; padding: 4rem 2rem; background: var(--bg-card); border-radius: var(--radius-lg); border: 1px dashed rgba(255,255,255,0.12); box-shadow: var(--shadow-card);">
          <div style="width: 56px; height: 56px; border-radius: 14px; background: rgba(16, 185, 129, 0.1); border: 1px solid rgba(16, 185, 129, 0.2); display: flex; align-items: center; justify-content: center; margin: 0 auto 1.25rem; color: var(--primary);">
            <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polygon points="12 2 2 7 12 12 22 7 12 2"></polygon>
              <polyline points="2 17 12 22 22 17"></polyline>
              <polyline points="2 12 12 17 22 12"></polyline>
            </svg>
          </div>
          <h3 style="font-size: 1.35rem; font-weight: 800; color: #fff;">No Minecraft servers deployed yet</h3>
          <p style="color: var(--text-dim); margin-top: 0.5rem; max-width: 440px; margin-left: auto; margin-right: auto; font-size: 0.9rem;">
            Click the button below to deploy your first high-performance PaperMC, Purpur, or Vanilla Minecraft server.
          </p>
          <button class="btn btn-primary" onclick="openCreateServerModal()" style="margin-top: 1.5rem;">
            + Deploy Server
          </button>
        </div>
      `;
      return;
    }

    servers.forEach((s) => {
      const card = document.createElement('div');
      card.className = 'server-card';
      const statusClass = `badge-${s.stats?.status || 'offline'}`;
      const memUsed = s.stats?.memory_mb || 0;
      const memMax = s.memory_max || 1024;
      const memPct = Math.min(100, Math.round((memUsed / memMax) * 100));

      card.innerHTML = `
        <div>
          <div class="server-card-header">
            <div>
              <h3 class="server-name">${escapeHtml(s.name)}</h3>
              <p class="server-sub">Port: ${s.port} &bull; ${s.jar_file}</p>
            </div>
            <div style="display: flex; gap: 0.5rem; align-items: center;">
              <span class="badge badge-flavor">${escapeHtml(s.type)} ${escapeHtml(s.version)}</span>
              <span class="badge ${statusClass}">
                <span class="status-dot"></span>
                ${s.stats?.status || 'offline'}
              </span>
            </div>
          </div>

          <!-- Mini RAM Meter -->
          <div style="margin: 0.75rem 0 0.25rem;">
            <div style="display: flex; justify-content: space-between; font-size: 0.725rem; font-weight: 700; color: var(--text-dim); margin-bottom: 0.35rem;">
              <span>MEMORY ALLOCATION</span>
              <span>${memUsed} / ${memMax} MB (${memPct}%)</span>
            </div>
            <div style="background: rgba(255, 255, 255, 0.06); height: 5px; border-radius: 3px; overflow: hidden;">
              <div style="background: linear-gradient(90deg, var(--primary), var(--accent)); width: ${memPct}%; height: 100%; transition: width 0.3s ease;"></div>
            </div>
          </div>

          <div class="server-stats-grid">
            <div class="stat-item"><span class="label">CPU Usage</span><span class="val">${(s.stats?.cpu_usage || 0).toFixed(1)}%</span></div>
            <div class="stat-item"><span class="label">Memory</span><span class="val">${memUsed} MB</span></div>
            <div class="stat-item"><span class="label">Disk Usage</span><span class="val">${s.stats?.disk_mb || 0} MB</span></div>
            <div class="stat-item"><span class="label">Uptime</span><span class="val">${formatUptime(s.stats?.uptime_sec || 0)}</span></div>
          </div>
        </div>

        <div style="display: flex; gap: 0.5rem; justify-content: flex-end; margin-top: 1rem; border-top: 1px solid var(--border-subtle); padding-top: 0.85rem;">
          <button class="btn btn-secondary btn-sm" onclick="openServer('${s.id}')">
            Manage
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
              <line x1="5" y1="12" x2="19" y2="12"></line>
              <polyline points="12 5 19 12 12 19"></polyline>
            </svg>
          </button>
        </div>
      `;
      grid.appendChild(card);
    });
  } catch (e) {}
}

// --- Server Detail ---
async function openServer(serverId) {
  try {
    const srv = await api(`/api/servers/${serverId}`);
    state.selectedServer = srv;
    showView('server-detail-view');

    document.getElementById('detail-server-name').textContent = srv.name;
    document.getElementById('detail-server-meta').textContent = `${srv.type.toUpperCase()} ${srv.version} • Port ${srv.port} • Max RAM ${srv.memory_max} MB`;

    switchTab('console');
    updateServerPowerBadges(srv.stats?.status || 'offline');
    initConsoleWebSocket(srv.id);
  } catch (e) {}
}

function updateServerPowerBadges(status) {
  const badge = document.getElementById('detail-server-status');
  badge.className = `badge badge-${status}`;
  badge.innerHTML = `<span class="status-dot"></span>${status}`;
}

// --- Console & WebSocket ---
function updateConsoleConn(status, label) {
  const el = document.getElementById('console-conn-status');
  if (el) {
    el.className = `conn-status conn-${status}`;
    const txt = el.querySelector('.conn-text');
    if (txt) txt.textContent = label;
  }
}

function initConsoleWebSocket(serverId) {
  if (state.reconnectTimeout) {
    clearTimeout(state.reconnectTimeout);
    state.reconnectTimeout = null;
  }

  if (state.activeWs) {
    try { state.activeWs.close(); } catch (e) {}
    state.activeWs = null;
  }

  const logBox = document.getElementById('terminal-logs');
  const inputEl = document.getElementById('terminal-input');
  if (!state.consoleMgr) {
    state.consoleMgr = new ConsoleManager(logBox, inputEl, (cmd) => {
      sendConsoleCommand(cmd);
    });
  } else {
    state.consoleMgr.container = logBox;
    state.consoleMgr.input = inputEl;
  }
  state.consoleMgr.clear();

  updateConsoleConn('reconnecting', state.reconnectAttempts > 0 ? 'Reconnecting...' : 'Connecting...');

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${protocol}//${window.location.host}/api/servers/${serverId}/ws?token=${state.token}`;

  const ws = new WebSocket(wsUrl);
  state.activeWs = ws;

  ws.onopen = () => {
    state.reconnectAttempts = 0;
    updateConsoleConn('connected', 'Connected');
    state.consoleMgr.appendLine('\x1b[32m[NeoPanel] Console connected to server instance.\x1b[0m');
  };

  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data);
      if (msg.event === 'log') {
        state.consoleMgr.appendLine(msg.data);
      }
    } catch (e) {}
  };

  ws.onclose = () => {
    if (state.selectedServer && state.selectedServer.id === serverId && state.currentTab === 'console') {
      updateConsoleConn('reconnecting', 'Disconnected');
      scheduleConsoleReconnect(serverId);
    } else {
      updateConsoleConn('offline', 'Offline');
    }
  };

  ws.onerror = () => {
    updateConsoleConn('offline', 'Error');
  };
}

function scheduleConsoleReconnect(serverId) {
  if (state.reconnectTimeout) clearTimeout(state.reconnectTimeout);
  const delay = Math.min(1000 * Math.pow(1.5, state.reconnectAttempts), 8000);
  state.reconnectAttempts++;
  updateConsoleConn('reconnecting', `Retry in ${(delay / 1000).toFixed(0)}s`);
  state.reconnectTimeout = setTimeout(() => {
    if (state.selectedServer && state.selectedServer.id === serverId && state.currentTab === 'console') {
      initConsoleWebSocket(serverId);
    }
  }, delay);
}

function openPopoutConsole() {
  if (!state.selectedServer) return;
  const url = `/console.html?id=${state.selectedServer.id}`;
  const width = Math.min(1050, window.screen.width - 80);
  const height = Math.min(720, window.screen.height - 100);
  const left = (window.screen.width - width) / 2;
  const top = (window.screen.height - height) / 2;
  window.open(url, `neopanel_console_${state.selectedServer.id}`, `width=${width},height=${height},left=${left},top=${top},menubar=no,toolbar=no,location=no,status=no`);
}

function sendConsoleCommand(cmd) {
  if (!cmd) return;
  if (state.activeWs && state.activeWs.readyState === WebSocket.OPEN) {
    state.activeWs.send(JSON.stringify({ event: 'command', data: cmd }));
  } else {
    api(`/api/servers/${state.selectedServer.id}/command`, {
      method: 'POST',
      body: JSON.stringify({ command: cmd }),
    });
  }
}

// --- Power Actions ---
async function serverPower(action) {
  if (!state.selectedServer) return;
  try {
    const res = await api(`/api/servers/${state.selectedServer.id}/power`, {
      method: 'POST',
      body: JSON.stringify({ action }),
    });
    showToast(`Sent ${action.toUpperCase()} command`);
    if (res.current) {
      updateServerPowerBadges(res.current.status);
    }
  } catch (e) {}
}

// --- Tab Switching ---
function switchTab(tab) {
  state.currentTab = tab;
  ['console', 'files', 'plugins', 'settings'].forEach((t) => {
    const el = document.getElementById(`tab-content-${t}`);
    const btn = document.getElementById(`tab-btn-${t}`);
    if (el) el.style.display = t === tab ? 'block' : 'none';
    if (btn) btn.className = t === tab ? 'tab-btn active' : 'tab-btn';
  });

  if (tab === 'console' && state.selectedServer) {
    if (!state.activeWs || state.activeWs.readyState !== WebSocket.OPEN) {
      initConsoleWebSocket(state.selectedServer.id);
    }
  } else if (tab === 'files') {
    loadFiles('');
  } else if (tab === 'plugins') {
    loadPluginsTab();
  } else if (tab === 'settings') {
    populateSettingsForm();
  }
}

// --- Plugins & Addons Flow ---
async function loadPluginsTab() {
  loadInstalledPlugins();
  const resultsGrid = document.getElementById('plugin-search-results');
  if (resultsGrid && !resultsGrid.children.length) {
    searchPlugins('');
  }
}

async function searchPlugins(query) {
  const inputEl = document.getElementById('plugin-search-input');
  const q = query !== undefined ? query : (inputEl ? inputEl.value.trim() : '');
  const resultsGrid = document.getElementById('plugin-search-results');
  if (!resultsGrid) return;

  resultsGrid.innerHTML = `
    <div style="grid-column: 1/-1; text-align: center; padding: 2.5rem; color: var(--text-dim);">
      <p style="font-size: 0.95rem;">Searching Modrinth API...</p>
    </div>
  `;

  try {
    const searchTarget = q || (state.activeCategory || 'paper');
    const data = await api(`/api/plugins/search?query=${encodeURIComponent(searchTarget)}&limit=24`);
    renderPluginResults(data.hits || []);
  } catch (err) {
    resultsGrid.innerHTML = `<div style="grid-column:1/-1; text-align:center; padding:2rem; color:var(--danger);">Failed to search plugins: ${escapeHtml(err.message)}</div>`;
  }
}

function renderPluginResults(hits) {
  const grid = document.getElementById('plugin-search-results');
  if (!grid) return;
  grid.innerHTML = '';

  if (!hits || hits.length === 0) {
    grid.innerHTML = `<div style="grid-column: 1/-1; text-align:center; padding:2.5rem; color:var(--text-dim);">No plugins found on Modrinth matching this query.</div>`;
    return;
  }

  hits.forEach(plugin => {
    const card = document.createElement('div');
    card.className = 'plugin-card';
    const icon = plugin.icon_url || '/favicon.svg';
    const downloads = formatNumber(plugin.downloads);

    card.innerHTML = `
      <div>
        <div class="plugin-header">
          <img src="${escapeHtml(icon)}" class="plugin-icon" onerror="this.src='/favicon.svg'" alt="">
          <div class="plugin-info">
            <h4>${escapeHtml(plugin.title)}</h4>
            <div class="plugin-author">by ${escapeHtml(plugin.author)}</div>
          </div>
        </div>
        <p class="plugin-desc">${escapeHtml(plugin.description || 'No description provided.')}</p>
      </div>
      <div class="plugin-footer">
        <span class="plugin-downloads">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
          ${downloads}
        </span>
        <button class="btn btn-primary btn-sm" onclick="installPlugin('${escapeHtml(plugin.slug || plugin.project_id)}', this)">
          Install
        </button>
      </div>
    `;
    grid.appendChild(card);
  });
}

function filterPluginsCategory(chip, category) {
  document.querySelectorAll('.category-chip').forEach(c => c.classList.remove('active'));
  chip.classList.add('active');
  state.activeCategory = category;
  searchPlugins(category);
}

async function installPlugin(slug, btn) {
  if (!state.selectedServer) return;
  const originalText = btn.innerHTML;
  btn.disabled = true;
  btn.innerHTML = `Installing...`;

  try {
    const res = await api(`/api/servers/${state.selectedServer.id}/plugins/install`, {
      method: 'POST',
      body: JSON.stringify({ slug }),
    });
    btn.innerHTML = `✔ Installed`;
    btn.className = 'btn btn-secondary btn-sm';
    showToast(`Installed '${res.filename}'!`);
    loadInstalledPlugins();
  } catch (err) {
    btn.disabled = false;
    btn.innerHTML = originalText;
    showToast(`Install failed: ${err.message}`, 'error');
  }
}

async function loadInstalledPlugins() {
  if (!state.selectedServer) return;
  const tbody = document.getElementById('installed-plugins-tbody');
  const label = document.getElementById('installed-folder-label');
  if (!tbody) return;
  tbody.innerHTML = `<tr><td colspan="4" style="text-align:center; padding:1.5rem; color:var(--text-dim);">Loading installed plugins...</td></tr>`;

  try {
    const data = await api(`/api/servers/${state.selectedServer.id}/plugins`);
    if (label) label.textContent = data.folder === 'mods' ? 'Mods' : 'Plugins';
    renderInstalledPlugins(data.plugins || []);
  } catch (err) {
    tbody.innerHTML = `<tr><td colspan="4" style="text-align:center; padding:1.5rem; color:var(--danger);">${escapeHtml(err.message)}</td></tr>`;
  }
}

function renderInstalledPlugins(plugins) {
  const tbody = document.getElementById('installed-plugins-tbody');
  if (!tbody) return;
  tbody.innerHTML = '';

  if (!plugins || plugins.length === 0) {
    tbody.innerHTML = `<tr><td colspan="4" style="text-align:center; padding:1.5rem; color:var(--text-dim);">No plugins currently installed. Search above to install 1-click plugins from Modrinth!</td></tr>`;
    return;
  }

  plugins.forEach(p => {
    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td>
        <div style="display:flex; align-items:center; gap:0.65rem; font-weight:600; font-family:var(--font-mono); font-size:0.85rem;">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--primary)" stroke-width="2"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path></svg>
          ${escapeHtml(p.filename)}
        </div>
      </td>
      <td style="color:var(--text-dim); font-size:0.8rem;">${formatBytes(p.size)}</td>
      <td style="color:var(--text-dim); font-size:0.8rem;">${new Date(p.modified).toLocaleString()}</td>
      <td style="text-align:right;">
        <button class="btn btn-danger btn-sm" onclick="deletePlugin('${escapeHtml(p.filename)}')">
          Uninstall
        </button>
      </td>
    `;
    tbody.appendChild(tr);
  });
}

async function deletePlugin(filename) {
  if (!confirm(`Are you sure you want to uninstall and remove '${filename}'?`)) return;
  try {
    await api(`/api/servers/${state.selectedServer.id}/plugins?filename=${encodeURIComponent(filename)}`, {
      method: 'DELETE',
    });
    showToast(`Removed '${filename}'`);
    loadInstalledPlugins();
  } catch (err) {
    showToast(err.message, 'error');
  }
}

// --- SFTP Modal ---
async function openSFTPModal() {
  if (!state.selectedServer) return;
  try {
    const sftp = await api(`/api/servers/${state.selectedServer.id}/sftp`);
    document.getElementById('sftp-host').textContent = sftp.host;
    document.getElementById('sftp-port').textContent = sftp.port;
    document.getElementById('sftp-username').textContent = sftp.username;
    document.getElementById('sftp-modal').style.display = 'flex';
  } catch (err) {
    showToast('Failed to load SFTP details', 'error');
  }
}

function copyText(text) {
  if (!text) return;
  navigator.clipboard.writeText(text).then(() => {
    showToast('Copied to clipboard!');
  }).catch(() => {
    showToast('Failed to copy', 'error');
  });
}

function formatNumber(num) {
  if (!num) return '0';
  if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
  if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
  return num.toString();
}

// --- File Manager Icons & Utilities ---
function getFileIcon(name, isDir) {
  if (isDir) {
    return `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#f59e0b" stroke-width="2" style="flex-shrink:0;"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path></svg>`;
  }

  const ext = name.split('.').pop().toLowerCase();
  if (ext === 'jar') {
    return `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#a855f7" stroke-width="2" style="flex-shrink:0;"><rect x="2" y="7" width="20" height="14" rx="2" ry="2"></rect><path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"></path></svg>`;
  } else if (ext === 'properties' || ext === 'yml' || ext === 'yaml' || ext === 'json') {
    return `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#10b981" stroke-width="2" style="flex-shrink:0;"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line></svg>`;
  } else if (ext === 'log' || ext === 'txt') {
    return `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#06b6d4" stroke-width="2" style="flex-shrink:0;"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline></svg>`;
  }

  return `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#94a3b8" stroke-width="2" style="flex-shrink:0;"><path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"></path><polyline points="13 2 13 9 20 9"></polyline></svg>`;
}

async function loadFiles(path = '') {
  state.currentFilePath = path;
  try {
    const files = await api(`/api/servers/${state.selectedServer.id}/files?path=${encodeURIComponent(path)}`);
    renderFileList(files, path);
  } catch (e) {}
}

function renderFileList(files, currentPath) {
  const tbody = document.getElementById('file-list-tbody');
  tbody.innerHTML = '';

  // Breadcrumbs
  const breadcrumbEl = document.getElementById('file-breadcrumbs');
  const parts = currentPath.split('/').filter(Boolean);
  let breadcrumbHtml = `<a href="javascript:void(0)" onclick="loadFiles('')" style="color:var(--primary); font-weight:700;">root</a>`;
  let accum = '';
  parts.forEach((p) => {
    accum += (accum ? '/' : '') + p;
    breadcrumbHtml += ` <span style="color:var(--text-dim);">&sol;</span> <a href="javascript:void(0)" onclick="loadFiles('${accum}')" style="color:#fff; font-weight:600;">${escapeHtml(p)}</a>`;
  });
  breadcrumbEl.innerHTML = breadcrumbHtml;

  // If inside subfolder, show parent directory option
  if (currentPath) {
    const parent = parts.slice(0, -1).join('/');
    const tr = document.createElement('tr');
    tr.className = 'file-row';
    tr.innerHTML = `
      <td colspan="4">
        <span class="file-name" onclick="loadFiles('${parent}')">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="15 18 9 12 15 6"></polyline></svg>
          .. (parent folder)
        </span>
      </td>
    `;
    tbody.appendChild(tr);
  }

  // Sort: directories first, then files
  files.sort((a, b) => {
    if (a.is_dir === b.is_dir) return a.name.localeCompare(b.name);
    return a.is_dir ? -1 : 1;
  });

  files.forEach((f) => {
    const tr = document.createElement('tr');
    tr.className = 'file-row';
    const iconSvg = getFileIcon(f.name, f.is_dir);
    const clickAction = f.is_dir
      ? `loadFiles('${f.path}')`
      : `openFileEditor('${f.path}')`;

    tr.innerHTML = `
      <td><span class="file-name" onclick="${clickAction}">${iconSvg} ${escapeHtml(f.name)}</span></td>
      <td style="color: var(--text-dim); font-family: var(--font-mono); font-size: 0.8rem;">${f.is_dir ? '-' : formatBytes(f.size)}</td>
      <td style="color: var(--text-dim); font-size: 0.8rem;">${new Date(f.modified).toLocaleString()}</td>
      <td style="text-align: right;">
        <button class="btn btn-secondary btn-sm" onclick="deleteFile('${f.path}')" style="color: var(--danger); padding: 0.25rem 0.55rem;">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="3 6 5 6 21 6"></polyline>
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
          </svg>
        </button>
      </td>
    `;
    tbody.appendChild(tr);
  });
}

async function openFileEditor(path) {
  try {
    const res = await api(`/api/servers/${state.selectedServer.id}/files/content?path=${encodeURIComponent(path)}`);
    document.getElementById('editor-modal-title').textContent = path;
    document.getElementById('editor-file-path').value = path;
    document.getElementById('editor-content').value = res.content;
    document.getElementById('editor-modal').style.display = 'flex';
  } catch (e) {}
}

async function saveFileContent() {
  const path = document.getElementById('editor-file-path').value;
  const content = document.getElementById('editor-content').value;

  try {
    await api(`/api/servers/${state.selectedServer.id}/files/content`, {
      method: 'POST',
      body: JSON.stringify({ path, content }),
    });
    showToast('File saved successfully');
    closeModal('editor-modal');
  } catch (e) {}
}

async function deleteFile(path) {
  if (!confirm(`Are you sure you want to delete ${path}?`)) return;
  try {
    await api(`/api/servers/${state.selectedServer.id}/files?path=${encodeURIComponent(path)}`, {
      method: 'DELETE',
    });
    showToast('Deleted successfully');
    loadFiles(state.currentFilePath);
  } catch (e) {}
}

async function uploadFile() {
  const input = document.getElementById('file-upload-input');
  if (!input.files.length) return;

  const file = input.files[0];
  const formData = new FormData();
  formData.append('file', file);

  try {
    await api(`/api/servers/${state.selectedServer.id}/files/upload?path=${encodeURIComponent(state.currentFilePath)}`, {
      method: 'POST',
      body: formData,
    });
    showToast(`Uploaded ${file.name}`);
    input.value = '';
    loadFiles(state.currentFilePath);
  } catch (e) {}
}

// --- Server Settings ---
function populateSettingsForm() {
  const s = state.selectedServer;
  if (!s) return;
  document.getElementById('setting-name').value = s.name;
  document.getElementById('setting-mem-min').value = s.memory_min;
  document.getElementById('setting-mem-max').value = s.memory_max;
  document.getElementById('setting-java-args').value = s.java_args || '';
  document.getElementById('setting-jar-file').value = s.jar_file;
  document.getElementById('setting-auto-restart').checked = s.auto_restart;
}

async function saveServerSettings(e) {
  e.preventDefault();
  const s = state.selectedServer;
  const payload = {
    name: document.getElementById('setting-name').value,
    memory_min: parseInt(document.getElementById('setting-mem-min').value, 10),
    memory_max: parseInt(document.getElementById('setting-mem-max').value, 10),
    java_args: document.getElementById('setting-java-args').value,
    jar_file: document.getElementById('setting-jar-file').value,
    auto_restart: document.getElementById('setting-auto-restart').checked,
  };

  try {
    const updated = await api(`/api/servers/${s.id}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    });
    state.selectedServer = { ...s, ...updated };
    showToast('Settings saved successfully');
  } catch (e) {}
}

async function deleteCurrentServer() {
  if (!state.selectedServer) return;
  if (!confirm(`WARNING: This will permanently delete ${state.selectedServer.name} and ALL its world data! Are you sure?`)) {
    return;
  }

  try {
    await api(`/api/servers/${state.selectedServer.id}`, { method: 'DELETE' });
    showToast('Server deleted');
    showView('dashboard-view');
    loadDashboard();
  } catch (e) {}
}

// --- Create Server Modal ---
function openCreateServerModal() {
  document.getElementById('create-server-modal').style.display = 'flex';
}

async function handleCreateServer(e) {
  e.preventDefault();
  const name = document.getElementById('new-server-name').value;
  const flavor = document.getElementById('new-server-type').value;
  const version = document.getElementById('new-server-version').value;
  const port = parseInt(document.getElementById('new-server-port').value, 10);
  const memMin = parseInt(document.getElementById('new-server-mem-min').value, 10);
  const memMax = parseInt(document.getElementById('new-server-mem-max').value, 10);
  const autoDownload = document.getElementById('new-server-autodl').checked;

  try {
    const created = await api('/api/servers', {
      method: 'POST',
      body: JSON.stringify({
        name,
        type: flavor,
        version,
        port,
        memory_min: memMin,
        memory_max: memMax,
        auto_download: autoDownload,
      }),
    });
    showToast(`Server '${created.name}' created!`);
    closeModal('create-server-modal');
    loadDashboard();
  } catch (e) {}
}

function closeModal(id) {
  document.getElementById(id).style.display = 'none';
}

// --- Utilities ---
function escapeHtml(str) {
  if (!str) return '';
  return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

function formatBytes(bytes) {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

function formatUptime(seconds) {
  if (!seconds) return 'Offline';
  const hrs = Math.floor(seconds / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  if (hrs > 0) return `${hrs}h ${mins}m`;
  return `${mins}m ${seconds % 60}s`;
}

// Boot application
window.addEventListener('DOMContentLoaded', checkAuth);
