// ANSI & Minecraft Color Formatter & Terminal Handler
const TerminalUtil = {
  // Map ANSI color codes to CSS colors
  ansiColors: {
    30: '#4b5563', // Black / Gray
    31: '#ef4444', // Red
    32: '#10b981', // Green
    33: '#f59e0b', // Yellow
    34: '#3b82f6', // Blue
    35: '#ec4899', // Magenta
    36: '#06b6d4', // Cyan
    37: '#f3f4f6', // White
    90: '#6b7280', // Bright Black
    91: '#f87171', // Bright Red
    92: '#34d399', // Bright Green
    93: '#fbbf24', // Bright Yellow
    94: '#60a5fa', // Bright Blue
    95: '#f472b6', // Bright Magenta
    96: '#22d3ee', // Bright Cyan
    97: '#ffffff', // Bright White
  },

  format(text) {
    if (!text) return '';

    // First escape HTML entities
    let escaped = text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');

    // Handle Minecraft § color codes
    const mcColors = {
      '0': '#000000', '1': '#0000AA', '2': '#00AA00', '3': '#00AAAA',
      '4': '#AA0000', '5': '#AA00AA', '6': '#FFAA00', '7': '#AAAAAA',
      '8': '#555555', '9': '#5555FF', 'a': '#55FF55', 'b': '#55FFFF',
      'c': '#FF5555', 'd': '#FF55FF', 'e': '#FFFF55', 'f': '#FFFFFF'
    };

    escaped = escaped.replace(/§([0-9a-f])/gi, (match, code) => {
      const color = mcColors[code.toLowerCase()];
      return `<span style="color:${color}">`;
    }).replace(/§r/gi, '</span>');

    // Handle ANSI escape codes: \x1b[32m or \033[32m
    escaped = escaped.replace(/\x1b\[([0-9;]+)m/g, (match, p1) => {
      const codes = p1.split(';').map(Number);
      let styles = [];

      for (const code of codes) {
        if (code === 0) {
          return '</span>';
        } else if (code === 1) {
          styles.push('font-weight: bold;');
        } else if (TerminalUtil.ansiColors[code]) {
          styles.push(`color: ${TerminalUtil.ansiColors[code]};`);
        }
      }

      if (styles.length > 0) {
        return `<span style="${styles.join(' ')}">`;
      }
      return '';
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

