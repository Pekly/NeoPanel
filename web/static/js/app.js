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
};

// --- Toast Notifications ---
function showToast(message, type = 'success') {
  const container = document.getElementById('toast-container');
  const toast = document.createElement('div');
  toast.className = `toast ${type === 'error' ? 'error' : ''}`;
  toast.textContent = message;
  container.appendChild(toast);
  setTimeout(() => {
    toast.remove();
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
    document.getElementById('nav-user').textContent = user.username;
    showView('dashboard-view');
    loadDashboard();
  } catch (e) {
    showView('login-view');
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
    document.getElementById('nav-user').textContent = res.user.username;
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
    document.getElementById('nav-user').textContent = res.user.username;
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
        <div style="grid-column: 1/-1; text-align: center; padding: 3rem; background: var(--bg-card); border-radius: 8px; border: 1px dashed var(--border);">
          <h3>No Minecraft servers yet</h3>
          <p style="color: var(--text-muted); margin-top: 0.5rem;">Click "Create Server" to deploy your first Paper, Purpur, or Vanilla server.</p>
        </div>
      `;
      return;
    }

    servers.forEach((s) => {
      const card = document.createElement('div');
      card.className = 'server-card';
      const statusClass = `badge-${s.stats?.status || 'offline'}`;

      card.innerHTML = `
        <div>
          <div class="server-card-header">
            <div>
              <h3 class="server-name">${escapeHtml(s.name)}</h3>
              <p class="server-sub">Port: ${s.port} &bull; ${s.jar_file}</p>
            </div>
            <div style="display: flex; gap: 0.5rem;">
              <span class="badge badge-flavor">${escapeHtml(s.type)} ${escapeHtml(s.version)}</span>
              <span class="badge ${statusClass}">${s.stats?.status || 'offline'}</span>
            </div>
          </div>
          <div class="server-stats-grid">
            <div class="stat-item"><span class="label">CPU Usage</span><span class="val">${(s.stats?.cpu_usage || 0).toFixed(1)}%</span></div>
            <div class="stat-item"><span class="label">Memory</span><span class="val">${s.stats?.memory_mb || 0} MB / ${s.memory_max} MB</span></div>
            <div class="stat-item"><span class="label">Disk Usage</span><span class="val">${s.stats?.disk_mb || 0} MB</span></div>
            <div class="stat-item"><span class="label">Uptime</span><span class="val">${formatUptime(s.stats?.uptime_sec || 0)}</span></div>
          </div>
        </div>
        <div style="display: flex; gap: 0.5rem; justify-content: flex-end; margin-top: 1rem;">
          <button class="btn btn-secondary btn-sm" onclick="openServer('${s.id}')">Manage &rarr;</button>
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
    document.getElementById('detail-server-meta').textContent = `${srv.type.toUpperCase()} ${srv.version} &bull; Port ${srv.port}`;

    switchTab('console');
    updateServerPowerBadges(srv.stats?.status || 'offline');
    initConsoleWebSocket(srv.id);
  } catch (e) {}
}

function updateServerPowerBadges(status) {
  const badge = document.getElementById('detail-server-status');
  badge.className = `badge badge-${status}`;
  badge.textContent = status;
}

// --- Console & WebSocket ---
function initConsoleWebSocket(serverId) {
  if (state.activeWs) {
    state.activeWs.close();
    state.activeWs = null;
  }

  const logBox = document.getElementById('terminal-logs');
  const inputEl = document.getElementById('terminal-input');
  state.consoleMgr = new ConsoleManager(logBox, inputEl, (cmd) => {
    sendConsoleCommand(cmd);
  });
  state.consoleMgr.clear();

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${protocol}//${window.location.host}/api/servers/${serverId}/ws?token=${state.token}`;

  const ws = new WebSocket(wsUrl);
  state.activeWs = ws;

  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data);
      if (msg.event === 'log') {
        state.consoleMgr.appendLine(msg.data);
      }
    } catch (e) {}
  };

  ws.onclose = () => {
    if (state.selectedServer && state.selectedServer.id === serverId) {
      state.consoleMgr.appendLine('\x1b[33m[NeoPanel] Console disconnected. Reconnecting...\x1b[0m');
    }
  };
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
    showToast(`Sent ${action.toUpperCase()} signal`);
    if (res.current) {
      updateServerPowerBadges(res.current.status);
    }
  } catch (e) {}
}

// --- Tab Switching ---
function switchTab(tab) {
  state.currentTab = tab;
  ['console', 'files', 'settings'].forEach((t) => {
    const el = document.getElementById(`tab-content-${t}`);
    const btn = document.getElementById(`tab-btn-${t}`);
    if (el) el.style.display = t === tab ? 'block' : 'none';
    if (btn) btn.className = t === tab ? 'tab-btn active' : 'tab-btn';
  });

  if (tab === 'files') {
    loadFiles('');
  } else if (tab === 'settings') {
    populateSettingsForm();
  }
}

// --- File Manager ---
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
  let breadcrumbHtml = `<a href="javascript:void(0)" onclick="loadFiles('')" style="color:var(--primary);">root</a>`;
  let accum = '';
  parts.forEach((p, idx) => {
    accum += (accum ? '/' : '') + p;
    breadcrumbHtml += ` / <a href="javascript:void(0)" onclick="loadFiles('${accum}')" style="color:var(--primary);">${escapeHtml(p)}</a>`;
  });
  breadcrumbEl.innerHTML = breadcrumbHtml;

  // If inside subfolder, show parent directory option
  if (currentPath) {
    const parent = parts.slice(0, -1).join('/');
    const tr = document.createElement('tr');
    tr.className = 'file-row';
    tr.innerHTML = `
      <td colspan="4"><span class="file-name" onclick="loadFiles('${parent}')">&#128193; .. (parent directory)</span></td>
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
    const icon = f.is_dir ? '&#128193;' : '&#128196;';
    const clickAction = f.is_dir
      ? `loadFiles('${f.path}')`
      : `openFileEditor('${f.path}')`;

    tr.innerHTML = `
      <td><span class="file-name" onclick="${clickAction}">${icon} ${escapeHtml(f.name)}</span></td>
      <td>${f.is_dir ? '-' : formatBytes(f.size)}</td>
      <td>${new Date(f.modified).toLocaleString()}</td>
      <td style="text-align: right;">
        <button class="btn btn-secondary btn-sm" onclick="deleteFile('${f.path}')">&#128465;</button>
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

