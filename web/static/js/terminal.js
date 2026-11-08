// ANSI & Minecraft Color Formatter & Terminal Handler
const TerminalUtil = {
  palette: [
    { fg: '#38bdf8', bg: 'rgba(56, 189, 248, 0.12)', border: 'rgba(56, 189, 248, 0.28)' },
    { fg: '#34d399', bg: 'rgba(52, 211, 153, 0.12)', border: 'rgba(52, 211, 153, 0.28)' },
    { fg: '#a78bfa', bg: 'rgba(167, 139, 250, 0.12)', border: 'rgba(167, 139, 250, 0.28)' },
    { fg: '#f472b6', bg: 'rgba(244, 114, 182, 0.12)', border: 'rgba(244, 114, 182, 0.28)' },
    { fg: '#fbbf24', bg: 'rgba(251, 191, 36, 0.12)', border: 'rgba(251, 191, 36, 0.28)' },
    { fg: '#2dd4bf', bg: 'rgba(45, 212, 191, 0.12)', border: 'rgba(45, 212, 191, 0.28)' },
    { fg: '#818cf8', bg: 'rgba(129, 140, 248, 0.12)', border: 'rgba(129, 140, 248, 0.28)' },
    { fg: '#fb923c', bg: 'rgba(251, 146, 60, 0.12)', border: 'rgba(251, 146, 60, 0.28)' },
    { fg: '#e879f9', bg: 'rgba(232, 121, 249, 0.12)', border: 'rgba(232, 121, 249, 0.28)' }
  ],

  getPluginColor(name) {
    let hash = 0;
    for (let i = 0; i < name.length; i++) {
      hash = name.charCodeAt(i) + ((hash << 5) - hash);
    }
    return TerminalUtil.palette[Math.abs(hash) % TerminalUtil.palette.length];
  },

  ansiColors: {
    30: '#4b5563', 31: '#ef4444', 32: '#10b981', 33: '#f59e0b',
    34: '#3b82f6', 35: '#ec4899', 36: '#06b6d4', 37: '#f3f4f6',
    90: '#6b7280', 91: '#f87171', 92: '#34d399', 93: '#fbbf24',
    94: '#60a5fa', 95: '#f472b6', 96: '#22d3ee', 97: '#ffffff',
  },

  ansiBgColors: {
    40: '#1f2937', 41: '#7f1d1d', 42: '#064e3b', 43: '#78350f',
    44: '#1e3a8a', 45: '#831843', 46: '#164e63', 47: '#374151',
    100: '#374151', 101: '#991b1b', 102: '#065f46', 103: '#92400e',
    104: '#1e40af', 105: '#9d174d', 106: '#155e75', 107: '#4b5563'
  },

  get256Color(n) {
    if (n < 16) return TerminalUtil.ansiColors[30 + (n % 8)] || '#ffffff';
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
  },

  format(text) {
    if (!text) return '';

    if (typeof text === 'string') {
      if (text.startsWith('<div class="con-cmd-row"')) return text;
      if (text.startsWith('<span class="c">&gt;') || text.startsWith('&gt; ') || text.startsWith('> ')) {
        const clean = text.replace(/<[^>]+>/g, '').replace(/^[>&;gt\s]+/, '').trim();
        const ts = new Date().toTimeString().split(' ')[0];
        return `<div class="con-cmd-row"><span class="con-ts">[${ts}]</span><span class="con-cmd-badge">CMD</span><span class="con-cmd-prefix">&gt;</span><span class="con-cmd-text">${clean.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')}</span></div>`;
      }
    }

    // First escape HTML entities
    let escaped = text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');

    // Minecraft § color codes
    const mcColors = {
      '0': '#000000', '1': '#0000AA', '2': '#00AA00', '3': '#00AAAA',
      '4': '#AA0000', '5': '#AA00AA', '6': '#FFAA00', '7': '#AAAAAA',
      '8': '#555555', '9': '#5555FF', 'a': '#55FF55', 'b': '#55FFFF',
      'c': '#FF5555', 'd': '#FF55FF', 'e': '#FFFF55', 'f': '#FFFFFF'
    };

    // Modern Hex formats: &#RRGGBB, §#RRGGBB, <#RRGGBB>
    escaped = escaped.replace(/(?:§|&amp;|&)#([0-9a-fA-F]{6})/gi, (_, hex) => `</span><span style="color:#${hex}">`);
    escaped = escaped.replace(/&lt;#([0-9a-fA-F]{6})&gt;/gi, (_, hex) => `</span><span style="color:#${hex}">`);

    // Standard Minecraft § and & codes
    escaped = escaped.replace(/(?:§|&amp;)([0-9a-f])/gi, (match, code) => {
      const color = mcColors[code.toLowerCase()] || '#FFFFFF';
      return `</span><span style="color:${color}">`;
    });
    escaped = escaped.replace(/(?:§|&amp;)l/gi, '<b>').replace(/(?:§|&amp;)r/gi, '</b></span>');
    escaped = escaped.replace(/(?:§|&amp;)o/gi, '<i>').replace(/(?:§|&amp;)n/gi, '<ins>').replace(/(?:§|&amp;)m/gi, '<del>');

    // Advanced ANSI: 16-color, 256-color, and 24-bit TrueColor RGB
    escaped = escaped.replace(/\x1b\[([0-9;]+)m/g, (_, p1) => {
      const tokens = p1.split(';').map(t => parseInt(t, 10) || 0);
      let styles = [];

      for (let i = 0; i < tokens.length; i++) {
        const c = tokens[i];
        if (c === 0) {
          styles = [];
        } else if (c === 1) {
          styles.push('font-weight: bold;');
        } else if (c === 2) {
          styles.push('opacity: 0.65;');
        } else if (c === 3) {
          styles.push('font-style: italic;');
        } else if (c === 4) {
          styles.push('text-decoration: underline;');
        } else if (c === 38) {
          if (tokens[i + 1] === 2 && tokens.length > i + 4) {
            styles.push(`color: rgb(${tokens[i+2]}, ${tokens[i+3]}, ${tokens[i+4]});`);
            i += 4;
          } else if (tokens[i + 1] === 5 && tokens.length > i + 2) {
            styles.push(`color: ${TerminalUtil.get256Color(tokens[i+2])};`);
            i += 2;
          }
        } else if (c === 48) {
          if (tokens[i + 1] === 2 && tokens.length > i + 4) {
            styles.push(`background-color: rgb(${tokens[i+2]}, ${tokens[i+3]}, ${tokens[i+4]});`);
            i += 4;
          } else if (tokens[i + 1] === 5 && tokens.length > i + 2) {
            styles.push(`background-color: ${TerminalUtil.get256Color(tokens[i+2])};`);
            i += 2;
          }
        } else if (TerminalUtil.ansiColors[c]) {
          styles.push(`color: ${TerminalUtil.ansiColors[c]};`);
        } else if (TerminalUtil.ansiBgColors[c]) {
          styles.push(`background-color: ${TerminalUtil.ansiBgColors[c]};`);
        }
      }

      return styles.length > 0 ? `</span><span style="${styles.join(' ')}">` : '</span>';
    });

    // Timestamps and Log Levels
    escaped = escaped.replace(/\[([0-9:.]+)\s*(INFO|WARN|WARNING|ERROR|SEVERE|FATAL|DEBUG|TRACE)\]:/g, (_, ts, lvl) => {
      const upper = lvl.toUpperCase();
      let lvlClass = 'info';
      if (upper.includes('WARN')) lvlClass = 'warn';
      else if (upper.includes('ERR') || upper.includes('SEV') || upper.includes('FATAL')) lvlClass = 'error';
      return `<span class="con-ts">[${ts}]</span> <span class="con-lvl con-lvl-${lvlClass}">${upper}</span><span class="con-colon">:</span>`;
    });

    // Plugin Bracket Badges like [SKSpawn], [SKCore], [Vault]
    escaped = escaped.replace(/\[([A-Za-z0-9_-]{2,30})\]/g, (match, pName) => {
      const reserved = ['INFO', 'WARN', 'WARNING', 'ERROR', 'SEVERE', 'FATAL', 'DEBUG', 'TRACE', 'SERVER', 'CLIENT'];
      if (reserved.includes(pName.toUpperCase())) return match;
      if (pName === 'NeoPanel') {
        return `<span class="con-plug-badge" style="color:#10b981; background:rgba(16,185,129,0.12); border:1px solid rgba(16,185,129,0.3); font-weight:700;">[NeoPanel]</span>`;
      }
      const c = TerminalUtil.getPluginColor(pName);
      return `<span class="con-plug-badge" style="color:${c.fg}; background:${c.bg}; border:1px solid ${c.border}; font-weight:600;">[${pName}]</span>`;
    });

    // Brand Prefixes: SMP »
    escaped = escaped.replace(/\b([A-Za-z0-9_-]{2,20})\s*(»|&raquo;|->|›)\s*/g, (_, brand, arrow) => {
      return `<span class="con-brand-prefix">${brand}</span> <span class="con-brand-arrow">${arrow}</span> `;
    });

    // Properties: protection.collision:
    escaped = escaped.replace(/(\b[a-zA-Z0-9_.-]{3,30}:)\s+/g, (match, prop) => {
      if (prop.startsWith('http:') || prop.startsWith('https:') || prop === 'INFO:' || prop === 'WARN:' || prop === 'ERROR:') {
        return match;
      }
      return `<span class="con-prop-key">${prop}</span> `;
    });

    return escaped;
  }
};

class ConsoleManager {
  constructor(containerEl, inputEl, sendCallback) {
    this.container = containerEl;
    this.input = inputEl;
    this.sendCallback = sendCallback;
    this.history = [];
    this.historyIdx = -1;
    this.autoScroll = true;
    this.maxLines = 1000;

    this.bindEvents();
  }

  bindEvents() {
    this.container.addEventListener('scroll', () => {
      const atBottom = this.container.scrollHeight - this.container.scrollTop <= this.container.clientHeight + 40;
      this.autoScroll = atBottom;
    });

    if (this.input) {
      this.input.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
          e.preventDefault();
          const cmd = this.input.value.trim();
          if (cmd) {
            this.history.push(cmd);
            this.historyIdx = this.history.length;
            this.appendCommandLine(cmd);
            this.sendCallback(cmd);
            this.input.value = '';
          }
        } else if (e.key === 'ArrowUp') {
          e.preventDefault();
          if (this.history.length > 0 && this.historyIdx > 0) {
            this.historyIdx--;
            this.input.value = this.history[this.historyIdx];
          }
        } else if (e.key === 'ArrowDown') {
          e.preventDefault();
          if (this.historyIdx < this.history.length - 1) {
            this.historyIdx++;
            this.input.value = this.history[this.historyIdx];
          } else {
            this.historyIdx = this.history.length;
            this.input.value = '';
          }
        }
      });
    }
  }

  appendCommandLine(cmd) {
    const ts = new Date().toTimeString().split(' ')[0];
    const lineDiv = document.createElement('div');
    lineDiv.className = 'terminal-line';
    const escaped = cmd.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    lineDiv.innerHTML = `<div class="con-cmd-row"><span class="con-ts">[${ts}]</span><span class="con-cmd-badge">CMD</span><span class="con-cmd-prefix">&gt;</span><span class="con-cmd-text">${escaped}</span></div>`;
    this.container.appendChild(lineDiv);
    if (this.autoScroll) {
      this.container.scrollTop = this.container.scrollHeight;
    }
  }

  appendLine(rawText) {
    const lineDiv = document.createElement('div');
    lineDiv.className = 'terminal-line';
    lineDiv.innerHTML = TerminalUtil.format(rawText);
    this.container.appendChild(lineDiv);

    // Prune excessive lines to keep memory ultra low
    while (this.container.children.length > this.maxLines) {
      this.container.removeChild(this.container.firstChild);
    }

    if (this.autoScroll) {
      this.container.scrollTop = this.container.scrollHeight;
    }
  }

  clear() {
    this.container.innerHTML = '';
  }
}

