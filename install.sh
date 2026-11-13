#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# NeoPanel All-In-One Installer for Ubuntu 24.04 LTS
# Supports:
#   - Master Panel (Dashboard + Daemon + Java + Caddy SSL + Firewall)
#   - Cluster Node Daemon (Dedicated Game Node, paired via API Key)
# GitHub: https://github.com/Pekly/NeoPanel
# -----------------------------------------------------------------------------

set -e

# Visual colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Default paths & constants
PANEL_PORT="8080"
CONFIG_DIR="/etc/neopanel"
DATA_DIR="/var/lib/neopanel"
SERVICE_FILE="/etc/systemd/system/neopanel.service"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="neopanel"
LOG_FILE="/tmp/neopanel-install.log"
BUILD_DIR="/tmp/neopanel-build"

# Git repository & branch configuration
GIT_REPO_URL="${NEOPANEL_REPO_URL:-${GIT_REPO_URL:-https://github.com/Pekly/NeoPanel.git}}"
GIT_BRANCH="${NEOPANEL_REPO_BRANCH:-${GIT_BRANCH:-main}}"

# Cloudflare Turnstile defaults (Full Panel mode only)
DEFAULT_TURNSTILE_SITE_KEY="0x4AAAAAAF7CgzdQmoC3Ho3H"
DEFAULT_TURNSTILE_SECRET_KEY="0x4AAAAAAF7Cg6ZKTqdgSA2s8Sg2oZ2HApU"
TURNSTILE_ENABLED="false"
TURNSTILE_SITE_KEY=""
TURNSTILE_SECRET_KEY=""

# Role & CLI pairing options
ACTION=""
INSTALL_ROLE=""
MASTER_URL=""
MASTER_API_KEY=""
NODE_NAME=""
NODE_ADVERTISE_URL=""
NODE_FQDN=""
NODE_LOCATION=""
NODE_TOKEN_OVERRIDE=""
ARG_USE_SSL=""
ARG_DOMAIN=""
ARG_EMAIL=""
ARG_CONF_UFW=""
ARG_USE_TURNSTILE=""
ARG_API_KEY_NAME=""
PAIR_INSECURE="false"
ASSUME_YES="false"

# Prevent apt interactive prompts from hanging in background
export DEBIAN_FRONTEND=noninteractive
export NEEDRESTART_MODE=a

# Initialize log file
: > "$LOG_FILE"

# -----------------------------------------------------------------------------
# Premium UI & Animated Progress System
# -----------------------------------------------------------------------------

print_banner() {
  if [ -t 1 ] || [ -e /dev/tty ]; then
    clear 2>/dev/null || true
  fi
  echo -e "${CYAN}${BOLD}"
  cat << 'EOF'
  _   _             ____                  _ 
 | \ | | ___  ___  |  _ \ __ _ _ __   ___| |
 |  \| |/ _ \/ _ \ | |_) / _` | '_ \ / _ \ |
 | |\  |  __/ (_) ||  __/ (_| | | | |  __/ |
 |_| \_|\___|\___/ |_|   \__,_|_| |_|\___|_|
EOF
  echo -e "${NC}"
  echo -e "${GREEN}${BOLD}>> Ultra-Lightweight Minecraft Server Manager Installer <<${NC}"
  echo -e "${YELLOW}Target: Linux (Debian, Ubuntu, CentOS, RHEL, Rocky, Alma, Fedora, Arch, Alpine)${NC}\n"
}

log_info() {
  echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
  echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
  echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
  echo -e "${RED}[ERROR]${NC} $1"
}

prompt_user() {
  local prompt_text="$1"
  local var_name="$2"
  if [ -t 0 ]; then
    read -r -p "$prompt_text" "$var_name" || true
  elif [ -e /dev/tty ]; then
    read -r -p "$prompt_text" "$var_name" </dev/tty || true
  else
    eval "$var_name=''"
  fi
}

# Renders a sleek progress bar with live percentage and label
draw_progress() {
  local pct="$1"
  local text="$2"
  local bar_width=28
  local filled=$(( pct * bar_width / 100 ))
  local empty=$(( bar_width - filled ))

  local bar=""
  for ((i=0; i<filled; i++)); do bar+="█"; done
  for ((i=0; i<empty; i++)); do bar+="░"; done

  # Carriage return with ANSI erase-to-end-of-line (\033[K) to cleanly overwrite line
  printf "\r\033[K  ${CYAN}[${GREEN}%s${CYAN}] ${BOLD}%3d%%${NC}  %-52s" "$bar" "$pct" "$text"
}

# Executes a background task silently while rendering an animated spinner & smooth percentage progress
run_step() {
  local start_pct="$1"
  local end_pct="$2"
  local label="$3"
  local cmd="$4"

  # Execute command in background, routing all output to log file
  eval "$cmd" >> "$LOG_FILE" 2>&1 &
  local pid=$!

  local cur=$start_pct
  local spin_idx=0
  local spinners=("⠋" "⠙" "⠹" "⠸" "⠼" "⠴" "⠦" "⠧" "⠇" "⠏")

  while kill -0 "$pid" 2>/dev/null; do
    local char="${spinners[$spin_idx]}"
    spin_idx=$(( (spin_idx + 1) % 10 ))

    draw_progress "$cur" "${char} ${label}..."
    if [ "$cur" -lt $((end_pct - 1)) ]; then
      cur=$(( cur + 1 ))
    fi
    sleep 0.12
  done

  # Use if wait "$pid" so set -e does not prematurely abort the script
  if wait "$pid"; then
    draw_progress "$end_pct" "✔ ${label}"
    echo ""
  else
    draw_progress "$cur" "✖ ${label} (FAILED)"
    echo -e "\n\n${RED}${BOLD}================================================================${NC}"
    echo -e "${RED}${BOLD}           Installation Error on Step: ${label}                  ${NC}"
    echo -e "${RED}${BOLD}================================================================${NC}\n"
    tail -n 35 "$LOG_FILE"
    echo -e "\n${YELLOW}Full detailed installation log saved to: ${LOG_FILE}${NC}\n"
    exit 1
  fi
}

check_root() {
  if [ "$EUID" -ne 0 ]; then
    log_error "Root privileges are required to install packages, configure UFW, and setup systemd."
    echo -e "\n${YELLOW}[REMINDER] Please re-run the installer using sudo:${NC}"
    echo -e "  ${CYAN}curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash${NC}\n"
    exit 1
  fi
}

check_os() {
  if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
    VERSION_ID=$VERSION_ID
  else
    log_error "Cannot detect operating system (/etc/os-release not found)."
    exit 1
  fi

  if [ "$OS" != "ubuntu" ]; then
    log_warn "This script is tailored for Ubuntu 24.04 LTS, but detected: $OS $VERSION_ID"
    if [ "$ASSUME_YES" != "true" ] && { [ -t 0 ] || [ -e /dev/tty ]; }; then
      prompt_user "Do you want to continue anyway? (y/N): " choice
      if [[ ! "$choice" =~ ^[Yy]$ ]]; then
        exit 1
      fi
    fi
  fi
}

get_public_ip() {
  curl -s4 --connect-timeout 4 https://ifconfig.me || curl -s4 --connect-timeout 4 https://api.ipify.org || hostname -I 2>/dev/null | awk '{print $1}' || echo "127.0.0.1"
}

generate_secret() {
  openssl rand -hex 32 2>/dev/null || date +%s | sha256sum | head -c 64
}

generate_node_token() {
  local rnd
  rnd=$(openssl rand -hex 28 2>/dev/null || { date +%s; uname -a; } | sha256sum | head -c 56)
  echo "neo_node_${rnd}"
}

detect_installed_role() {
  if [ "$INSTALL_ROLE" = "node" ] || [ "$INSTALL_ROLE" = "panel" ]; then
    return 0
  fi
  if [ -f "$CONFIG_DIR/config.yaml" ] && grep -Eq '^[[:space:]]*role:[[:space:]]*"?node"?' "$CONFIG_DIR/config.yaml"; then
    INSTALL_ROLE="node"
  else
    INSTALL_ROLE="panel"
  fi
}

# -----------------------------------------------------------------------------
# Background Installation Steps
# -----------------------------------------------------------------------------

detect_os_pkg_manager() {
  if command -v apt-get &>/dev/null; then
    PKG_MGR="apt"
  elif command -v dnf &>/dev/null; then
    PKG_MGR="dnf"
  elif command -v yum &>/dev/null; then
    PKG_MGR="yum"
  elif command -v pacman &>/dev/null; then
    PKG_MGR="pacman"
  elif command -v apk &>/dev/null; then
    PKG_MGR="apk"
  else
    PKG_MGR="unknown"
  fi
}

step_prerequisites() {
  detect_os_pkg_manager
  case "$PKG_MGR" in
    apt)
      apt-get update -y -q || return 1
      apt-get install -y -q -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" \
        curl wget jq tar ufw ca-certificates software-properties-common git openssl
      ;;
    dnf)
      dnf install -y curl wget jq tar firewalld ca-certificates git openssl
      ;;
    yum)
      yum install -y curl wget jq tar firewalld ca-certificates git openssl
      ;;
    pacman)
      pacman -Sy --noconfirm curl wget jq tar ufw ca-certificates git openssl
      ;;
    apk)
      apk update || return 1
      apk add curl wget jq tar ca-certificates git openssl
      ;;
    *)
      log_error "No supported package manager found. Install prerequisites manually on a supported Linux host."
      return 1
      ;;
  esac
}

# Keep isolation separate from general dependencies so updates repair older hosts
# too. Every failure is explicit even when a caller uses this function in an if.
install_isolation_tools() {
  local -a packages=()
  command -v bwrap >/dev/null 2>&1 || packages+=(bubblewrap)
  command -v timeout >/dev/null 2>&1 || packages+=(coreutils)
  if [ "${#packages[@]}" -eq 0 ]; then
    return 0
  fi
  detect_os_pkg_manager
  log_info "Installing isolation prerequisites: ${packages[*]}"
  case "$PKG_MGR" in
    apt)
      if ! apt-get update -y -q || ! apt-get install -y -q \
        -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" "${packages[@]}"; then
        log_error "Could not install isolation prerequisites with apt. Check the package/repository errors above."
        return 1
      fi
      ;;
    dnf|yum)
      if ! "$PKG_MGR" install -y "${packages[@]}"; then
        log_error "Could not install isolation prerequisites with $PKG_MGR. Check enabled repositories and package availability."
        return 1
      fi
      ;;
    pacman)
      if ! pacman -S --needed --noconfirm "${packages[@]}"; then
        log_error "Could not install isolation prerequisites with pacman. Check repository configuration."
        return 1
      fi
      ;;
    apk)
      if ! apk update || ! apk add "${packages[@]}"; then
        log_error "Could not install isolation prerequisites with apk. Check repository configuration."
        return 1
      fi
      ;;
    *)
      log_error "No supported package manager found. Install bubblewrap and a timeout utility manually, then retry."
      return 1
      ;;
  esac
  if ! command -v bwrap >/dev/null 2>&1 || ! command -v timeout >/dev/null 2>&1; then
    log_error "Package installation finished, but bwrap or timeout is still unavailable on PATH."
    return 1
  fi
}

# Subshell keeps cleanup and temporary state local to this step. No host security
# policy or saved node isolation setting is modified by this capability test.
step_isolation() (
  install_isolation_tools || return 1
  local bwrap_path probe_user=root probe_group="" load_state dynamic_user uid
  bwrap_path="$(command -v bwrap)" || return 1
  if ! "$bwrap_path" --version; then
    log_error "Bubblewrap is installed but cannot execute. Repair the package before retrying."
    return 1
  fi

  # Fresh installs create User=root. Updates must not accidentally test root
  # while an existing service is configured to run as an unprivileged user.
  if [ "${1:-fresh}" = "installed" ]; then
    if ! load_state="$(systemctl show neopanel.service --property=LoadState --value)" || [ "$load_state" != "loaded" ]; then
      log_error "Cannot inspect the installed neopanel.service. Restore/load its unit before updating."
      return 1
    fi
    if ! dynamic_user="$(systemctl show neopanel.service --property=DynamicUser --value)" || \
       ! probe_user="$(systemctl show neopanel.service --property=User --value)" || \
       ! probe_group="$(systemctl show neopanel.service --property=Group --value)"; then
      log_error "Cannot read the service identity; refusing to claim isolation was verified."
      return 1
    fi
    if [ "$dynamic_user" = "yes" ]; then
      log_error "DynamicUser=yes requires a service-specific sandbox check; this installer cannot safely reproduce that transient identity. The service was not changed."
      return 1
    fi
    probe_user="${probe_user:-root}"
  fi
  if ! uid="$(id -u "$probe_user")"; then
    log_error "Cannot resolve the NeoPanel service user '$probe_user'."
    return 1
  fi

  local probe_dir
  probe_dir="$(mktemp -d "${TMPDIR:-/var/tmp}/neopanel-isolation.XXXXXX")" || return 1
  trap 'rm -rf -- "$probe_dir"' EXIT
  chmod 755 "$probe_dir" || return 1
  printf '%s\n' 'Synthetic isolation probe; not a real secret.' > "$probe_dir/host-only" || return 1
  chmod 644 "$probe_dir/host-only" || return 1
  mkdir -m 700 "$probe_dir/work" || return 1
  local -a launcher=() sandbox=(
    "$bwrap_path" --unshare-all --share-net --die-with-parent --cap-drop ALL
    --clearenv --setenv PATH /usr/bin:/bin --setenv LANG C.UTF-8 --setenv HOME /server
    --ro-bind /usr /usr --ro-bind /bin /bin --ro-bind /lib /lib
    --ro-bind-try /lib64 /lib64 --dev /dev --proc /proc --tmpfs /tmp
    --bind "$probe_dir/work" /server --chdir /server
    -- /bin/sh -eu -c '
      test ! -e "$1"
      test ! -e /etc/neopanel
      test -z "${NEOPANEL_ISOLATION_PROBE_SECRET+x}"
      printf "%s\n" sandbox-ok > /server/probe-ok
    ' probe "$probe_dir/host-only"
  )
  if [ "$uid" != "0" ] || [ -n "$probe_group" ]; then
    if ! command -v runuser >/dev/null 2>&1; then
      log_error "Install runuser (usually util-linux) to test the configured service identity without changing it."
      return 1
    fi
    chown "$probe_user${probe_group:+:$probe_group}" "$probe_dir/work" || return 1
    launcher=(runuser -u "$probe_user")
    if [ -n "$probe_group" ]; then launcher+=(-g "$probe_group"); fi
    launcher+=(--)
  fi
  log_info "Testing Bubblewrap namespaces and filesystem boundaries as $probe_user (20-second limit)..."
  if ! NEOPANEL_ISOLATION_PROBE_SECRET=must-not-reach-sandbox \
    timeout -k 5 20 "${launcher[@]}" "${sandbox[@]}" > "$probe_dir/output" 2>&1; then
    cat "$probe_dir/output"
    log_error "Bubblewrap is installed, but the sandbox probe failed. Installation/update stopped before binary replacement or service restart."
    log_error "Check the error above: user namespaces may be restricted by AppArmor, kernel policy, or the VPS/container provider. Older Bubblewrap versions may lack required options."
    log_error "Review kernel/AppArmor logs and host namespace support. Do not globally disable AppArmor or silently switch to unrestricted mode."
    return 1
  fi
  if [ ! -f "$probe_dir/work/probe-ok" ] || [ "$(cat "$probe_dir/work/probe-ok")" != "sandbox-ok" ]; then
    log_error "The sandbox exited without its expected filesystem marker; isolation verification failed."
    return 1
  fi
  log_success "Bubblewrap host probe passed as $probe_user. Node settings were not changed."
  log_info "Choose Bubblewrap (bwrap) on the node to REQUIRE filesystem isolation. Systemd-only mode provides resource limits, not a filesystem sandbox."
  log_info "This verifies host capability and service identity, not all service-unit restrictions or every Java/plugin workload."
)

has_java_version() {
  local ver="$1"
  local d
  for d in /usr/lib/jvm/* /opt/java/*; do
    [ -d "$d" ] || continue
    case "$(basename "$d")" in
      *"-${ver}-"*|*"-${ver}"|"java-${ver}"*|"jdk-${ver}"*|"jre-${ver}"*|"temurin-${ver}"*)
        if [ -x "$d/bin/java" ] && "$d/bin/java" -version >/dev/null 2>&1; then
          return 0
        fi
        if [ -x "$d/jre/bin/java" ] && "$d/jre/bin/java" -version >/dev/null 2>&1; then
          return 0
        fi
        ;;
    esac
  done
  return 1
}

install_java_25() {
  if has_java_version "25"; then
    return 0
  fi

  # Remove any empty/broken directory left by an earlier failed download
  if [ -e "/usr/lib/jvm/java-25-temurin" ] && [ ! -x "/usr/lib/jvm/java-25-temurin/bin/java" ]; then
    rm -rf "/usr/lib/jvm/java-25-temurin"
  fi

  # Try Ubuntu apt repository if OpenJDK 25 is available
  if [ "$PKG_MGR" = "apt" ] && apt-cache show openjdk-25-jre-headless >/dev/null 2>&1; then
    apt-get install -y -q -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" \
      openjdk-25-jre-headless || true
    if has_java_version "25"; then
      return 0
    fi
  fi

  local arch="x64"
  if [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "arm64" ]; then
    arch="aarch64"
  fi

  mkdir -p /usr/lib/jvm
  local tmp_tar="/tmp/neopanel-temurin-25.tar.gz"
  local stage_dir="/usr/lib/jvm/.java-25-temurin.stage"
  local urls=(
    "https://api.adoptium.net/v3/binary/latest/25/ga/linux/${arch}/jre/hotspot/normal/eclipse"
    "https://api.adoptium.net/v3/binary/latest/25/ga/linux/${arch}/jdk/hotspot/normal/eclipse"
    "https://api.adoptium.net/v3/binary/latest/25/ea/linux/${arch}/jre/hotspot/normal/eclipse"
    "https://api.adoptium.net/v3/binary/latest/25/ea/linux/${arch}/jdk/hotspot/normal/eclipse"
  )

  local url
  for url in "${urls[@]}"; do
    rm -f "$tmp_tar"
    rm -rf "$stage_dir"
    echo "[INFO] Attempting Java 25 download from: $url"
    if curl -fsSL --retry 2 --connect-timeout 15 --max-time 300 -o "$tmp_tar" "$url"; then
      if tar -tzf "$tmp_tar" >/dev/null 2>&1; then
        mkdir -p "$stage_dir"
        if tar -xzf "$tmp_tar" -C "$stage_dir" --strip-components=1 && \
           [ -x "$stage_dir/bin/java" ] && \
           "$stage_dir/bin/java" -version >/dev/null 2>&1; then
          rm -rf "/usr/lib/jvm/java-25-temurin"
          mv "$stage_dir" "/usr/lib/jvm/java-25-temurin"
          chmod -R a+rX "/usr/lib/jvm/java-25-temurin"
          rm -f "$tmp_tar"
          echo "[INFO] Java 25 installed at /usr/lib/jvm/java-25-temurin"
          return 0
        fi
      fi
    fi
  done

  rm -f "$tmp_tar"
  rm -rf "$stage_dir"
  echo "[WARN] Failed to auto-install Java 25 from Adoptium." >&2
  return 1
}

step_java() {
  detect_os_pkg_manager
  case "$PKG_MGR" in
    apt)
      apt-get update -y -q
      apt-get install -y -q -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" \
        openjdk-21-jre-headless openjdk-17-jre-headless openjdk-8-jre-headless || true
      ;;
    dnf|yum)
      $PKG_MGR install -y java-21-openjdk-headless java-17-openjdk-headless java-1.8.0-openjdk-headless || true
      ;;
    pacman)
      pacman -Sy --noconfirm jre21-openjdk-headless jre17-openjdk-headless jre8-openjdk-headless || true
      ;;
    apk)
      apk add openjdk21-jre-headless openjdk17-jre-headless openjdk8-jre-base || true
      ;;
  esac

  # Install latest Java 25 runtime cleanly
  install_java_25 || true
}

step_go_compiler() {
  if command -v go &>/dev/null; then
    return 0
  fi
  detect_os_pkg_manager
  case "$PKG_MGR" in
    apt)
      apt-get update -y -q
      apt-get install -y -q -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" golang-go || true
      ;;
    dnf|yum)
      $PKG_MGR install -y golang || true
      ;;
    pacman)
      pacman -Sy --noconfirm go || true
      ;;
    apk)
      apk add go || true
      ;;
  esac

  if ! command -v go &>/dev/null; then
    local arch="amd64"
    if [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "arm64" ]; then
      arch="arm64"
    fi
    curl -fsSL "https://go.dev/dl/go1.22.4.linux-${arch}.tar.gz" -o /tmp/go.tar.gz 2>/dev/null || true
    if [ -f /tmp/go.tar.gz ]; then
      tar -C /usr/local -xzf /tmp/go.tar.gz 2>/dev/null || true
      export PATH=$PATH:/usr/local/go/bin
      rm -f /tmp/go.tar.gz
    fi
  fi
}

write_global_shortcuts() {
  # Install global neopanel-update command shortcut
  cat > "$INSTALL_DIR/neopanel-update.tmp" <<'EOF'
#!/usr/bin/env bash
set -e
REPO_URL="${NEOPANEL_REPO_URL:-https://github.com/Pekly/NeoPanel.git}"
REPO_BRANCH="${NEOPANEL_REPO_BRANCH:-main}"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo) REPO_URL="$2"; shift 2 ;;
    --branch) REPO_BRANCH="$2"; shift 2 ;;
    *) break ;;
  esac
done
CLEAN_SLUG="$(echo "$REPO_URL" | sed -E 's|https?://github.com/([^/]+)/([^/.]+)(\.git)?|\1/\2|')"
RAW_URL="https://raw.githubusercontent.com/${CLEAN_SLUG}/${REPO_BRANCH}/install.sh"
curl -sSL "$RAW_URL" | bash -s -- --update --repo "$REPO_URL" --branch "$REPO_BRANCH" "$@"
exit 0
EOF
  chmod +x "$INSTALL_DIR/neopanel-update.tmp"
  mv -f "$INSTALL_DIR/neopanel-update.tmp" "$INSTALL_DIR/neopanel-update"

  # Install global neopanel-installer command shortcut
  cat > "$INSTALL_DIR/neopanel-installer.tmp" <<'EOF'
#!/usr/bin/env bash
set -e
REPO_URL="${NEOPANEL_REPO_URL:-https://github.com/Pekly/NeoPanel.git}"
REPO_BRANCH="${NEOPANEL_REPO_BRANCH:-main}"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo) REPO_URL="$2"; shift 2 ;;
    --branch) REPO_BRANCH="$2"; shift 2 ;;
    *) break ;;
  esac
done
CLEAN_SLUG="$(echo "$REPO_URL" | sed -E 's|https?://github.com/([^/]+)/([^/.]+)(\.git)?|\1/\2|')"
RAW_URL="https://raw.githubusercontent.com/${CLEAN_SLUG}/${REPO_BRANCH}/install.sh"
curl -sSL "$RAW_URL" | bash -s -- --repo "$REPO_URL" --branch "$REPO_BRANCH" "$@"
exit 0
EOF
  chmod +x "$INSTALL_DIR/neopanel-installer.tmp"
  mv -f "$INSTALL_DIR/neopanel-installer.tmp" "$INSTALL_DIR/neopanel-installer"
}

step_build_neopanel() {
  mkdir -p "$CONFIG_DIR"
  mkdir -p "$DATA_DIR/servers"
  mkdir -p "$DATA_DIR/backups"

  local -a go_tags=()
  if [ "$INSTALL_ROLE" = "node" ]; then
    go_tags=(-tags node)
  fi

  local git_commit=""
  if [ -f "./cmd/neopanel/main.go" ] && [ -f "./cmd/neopanel/cli.go" ]; then
    git_commit="$(git rev-parse HEAD 2>/dev/null || true)"
    if [ -n "$git_commit" ]; then
      echo "$git_commit" > "$CONFIG_DIR/current_commit"
      echo "$git_commit" > "$DATA_DIR/current_commit"
    fi
    go mod tidy
    go build "${go_tags[@]}" -ldflags="-s -w -X 'github.com/Pekly/NeoPanel/internal/api.BuildCommit=${git_commit}'" -o "$INSTALL_DIR/$BINARY_NAME" ./cmd/neopanel
  else
    rm -rf "$BUILD_DIR"
    git clone --depth 1 --branch "$GIT_BRANCH" "$GIT_REPO_URL" "$BUILD_DIR"
    if [ -f "$BUILD_DIR/cmd/neopanel/main.go" ]; then
      git_commit="$(cd "$BUILD_DIR" && git rev-parse HEAD 2>/dev/null || true)"
      if [ -n "$git_commit" ]; then
        echo "$git_commit" > "$CONFIG_DIR/current_commit"
        echo "$git_commit" > "$DATA_DIR/current_commit"
      fi
      (cd "$BUILD_DIR" && go mod tidy && go build "${go_tags[@]}" -ldflags="-s -w -X 'github.com/Pekly/NeoPanel/internal/api.BuildCommit=${git_commit}'" -o "$INSTALL_DIR/$BINARY_NAME" ./cmd/neopanel)
      rm -rf "$BUILD_DIR"
    else
      echo "Failed to clone repository or main.go missing." >&2
      exit 1
    fi
  fi

  chmod +x "$INSTALL_DIR/$BINARY_NAME"
  write_global_shortcuts
}

prompt_turnstile_setup() {
  echo -e "\n${CYAN}${BOLD}Cloudflare Turnstile Bot Protection${NC}"
  echo -e "${YELLOW}Adds Cloudflare Turnstile challenge verification to your NeoPanel login and setup pages.${NC}"
  prompt_user "Would you like to enable Cloudflare Turnstile? (y/n) [default: y]: " ts_choice
  ts_choice=${ts_choice:-y}

  if [[ "$ts_choice" =~ ^[Yy]$ ]]; then
    TURNSTILE_ENABLED="true"
    prompt_user "Enter Turnstile Site Key [default: ${DEFAULT_TURNSTILE_SITE_KEY}]: " ts_site
    TURNSTILE_SITE_KEY=${ts_site:-$DEFAULT_TURNSTILE_SITE_KEY}

    prompt_user "Enter Turnstile Secret Key [default: ${DEFAULT_TURNSTILE_SECRET_KEY}]: " ts_secret
    TURNSTILE_SECRET_KEY=${ts_secret:-$DEFAULT_TURNSTILE_SECRET_KEY}
    log_success "Cloudflare Turnstile enabled (Site Key: ${TURNSTILE_SITE_KEY})."
  else
    TURNSTILE_ENABLED="false"
    TURNSTILE_SITE_KEY=""
    TURNSTILE_SECRET_KEY=""
    log_info "Cloudflare Turnstile skipped (you can enable it anytime via 'sudo neopanel-installer --turnstile' or Admin Settings)."
  fi
}

configure_turnstile_in_config() {
  local enabled="$1"
  local site_key="$2"
  local secret_key="$3"
  local cfg_file="$CONFIG_DIR/config.yaml"

  if [ ! -f "$cfg_file" ]; then
    return 0
  fi

  local tmp_file
  tmp_file=$(mktemp)

  awk -v en="$enabled" -v sk="$site_key" -v sec="$secret_key" '
    BEGIN { in_ts=0; inserted=0; in_sec=0 }
    /^security:/ { in_sec=1; print; next }
    in_sec && /^[^[:space:]]/ {
      if (!inserted) {
        print "  turnstile:"
        print "    enabled: " en
        print "    site_key: \"" sk "\""
        print "    secret_key: \"" sec "\""
        inserted=1
      }
      in_sec=0
    }
    in_sec && /^[[:space:]]+turnstile:/ {
      in_ts=1
      next
    }
    in_ts {
      if ($0 ~ /^[[:space:]]{4,}/) {
        next
      } else {
        in_ts=0
      }
    }
    { print }
    END {
      if (!inserted) {
        if (!in_sec) {
          print ""
          print "security:"
        }
        print "  turnstile:"
        print "    enabled: " en
        print "    site_key: \"" sk "\""
        print "    secret_key: \"" sec "\""
      }
    }
  ' "$cfg_file" > "$tmp_file"

  mv "$tmp_file" "$cfg_file"
  chmod 600 "$cfg_file"
}

step_config_service() {
  local host="$1"
  local port="$2"
  local ts_enabled="${3:-false}"
  local ts_site="${4:-}"
  local ts_secret="${5:-}"
  local role="${6:-panel}"
  local node_tok="${7:-}"

  local secret
  secret=$(generate_secret)

  local srv_mode="behind_proxy"
  if [ "$host" = "0.0.0.0" ] && [ "$role" = "node" ]; then
    srv_mode="standalone"
  fi

  if [ "$role" = "node" ]; then
    cat > "$CONFIG_DIR/config.yaml" <<EOF
server:
  host: "${host}"
  port: ${port}
  mode: "${srv_mode}"
  role: "node"

paths:
  data_dir: "${DATA_DIR}"
  servers_dir: "${DATA_DIR}/servers"
  db_path: "${DATA_DIR}/neopanel.db"

security:
  jwt_secret: "${secret}"
  token_expiration_hours: 72
  node_token: "${node_tok}"

java:
  default_path: "java"
EOF
  else
    cat > "$CONFIG_DIR/config.yaml" <<EOF
server:
  host: "${host}"
  port: ${port}
  mode: "behind_proxy"
  role: "panel"

paths:
  data_dir: "${DATA_DIR}"
  servers_dir: "${DATA_DIR}/servers"
  db_path: "${DATA_DIR}/neopanel.db"

security:
  jwt_secret: "${secret}"
  token_expiration_hours: 72
  node_token: "${node_tok}"
  turnstile:
    enabled: ${ts_enabled}
    site_key: "${ts_site}"
    secret_key: "${ts_secret}"

java:
  default_path: "java"
EOF
  fi

  chmod 600 "$CONFIG_DIR/config.yaml"

  local svc_desc="NeoPanel Minecraft Server Manager"
  if [ "$role" = "node" ]; then
    svc_desc="NeoPanel Cluster Node Daemon"
  fi

  cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=${svc_desc}
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${DATA_DIR}
ExecStart=${INSTALL_DIR}/${BINARY_NAME} -config ${CONFIG_DIR}/config.yaml
Restart=on-failure
RestartSec=5s
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

  systemctl daemon-reload
  systemctl enable neopanel
  systemctl restart neopanel
}

step_ssl_caddy() {
  local domain="$1"
  local email="$2"

  apt-get install -y -q debian-keyring debian-archive-keyring apt-transport-https curl
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor --yes -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg 2>/dev/null
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list >/dev/null
  apt-get update -y -q
  apt-get install -y -q -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" caddy

  cat > /etc/caddy/Caddyfile <<EOF
${domain} {
    encode gzip zstd
    tls ${email}

    reverse_proxy 127.0.0.1:${PANEL_PORT} {
        header_up Host {host}
        header_up X-Real-IP {remote_host}
    }
}
EOF

  systemctl restart caddy
}

step_firewall() {
  local include_http_ports="${1:-true}"
  ufw allow 22/tcp >/dev/null 2>&1 || true
  if [ "$include_http_ports" = "true" ]; then
    ufw allow 80/tcp >/dev/null 2>&1
    ufw allow 443/tcp >/dev/null 2>&1
  fi
  ufw allow 2022/tcp >/dev/null 2>&1
  ufw allow "${PANEL_PORT}/tcp" >/dev/null 2>&1
  ufw allow 25565:25600/tcp >/dev/null 2>&1
  ufw allow 25565:25600/udp >/dev/null 2>&1
  echo "y" | ufw enable >/dev/null 2>&1
}

wait_for_local_daemon() {
  local port="$1"
  local i=0
  while [ "$i" -lt 15 ]; do
    if curl -fsS "http://127.0.0.1:${port}/api/auth/status" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.4
    i=$((i + 1))
  done
  return 1
}

# -----------------------------------------------------------------------------
# Master Panel Installation Flow (Dashboard + Daemon)
# -----------------------------------------------------------------------------

run_full_installation() {
  INSTALL_ROLE="panel"
  print_banner
  check_root
  check_os

  local pub_ip
  pub_ip=$(get_public_ip)

  echo -e "${BOLD}Mode:${NC} ${CYAN}NeoPanel Master Panel (Dashboard + Local Daemon)${NC}"
  echo -e "${BOLD}Server Public IP detected:${NC} ${CYAN}${pub_ip}${NC}\n"

  local use_ssl="${ARG_USE_SSL}"
  local domain="${ARG_DOMAIN}"
  local email="${ARG_EMAIL}"
  local bind_host="0.0.0.0"

  if [ -n "$domain" ]; then
    use_ssl="y"
  fi
  if [ -z "$use_ssl" ]; then
    echo -e "${YELLOW}Would you like to configure a custom domain with automated Let's Encrypt SSL?${NC}"
    prompt_user "Use Domain & SSL? (y/n) [default: y]: " use_ssl
    use_ssl=${use_ssl:-y}
  fi

  if [[ "$use_ssl" =~ ^[Yy]$ ]]; then
    if [ -z "$domain" ]; then
      echo -e "\n${CYAN}[REMINDER] Ensure your domain's DNS 'A' record points to this server's IP (${pub_ip}).${NC}"
      echo -e "${YELLOW}           If DNS has not propagated yet, Let's Encrypt certificate issuance will fail.${NC}\n"

      prompt_user "Enter your Domain name (e.g. panel.example.com): " domain
      while [ -z "$domain" ]; do
        prompt_user "Domain cannot be empty. Enter Domain: " domain
      done

      local resolved_ip
      resolved_ip=$(getent ahostsv4 "$domain" 2>/dev/null | head -n 1 | awk '{print $1}')
      if [ -n "$resolved_ip" ] && [ "$resolved_ip" != "$pub_ip" ]; then
        log_warn "DNS check: '$domain' resolves to $resolved_ip, but server IP is $pub_ip."
        echo -e "${YELLOW}[REMINDER] If you just updated your DNS record, please wait a minute for propagation.${NC}"
        prompt_user "Do you want to continue anyway? (y/N): " override_dns
        if [[ ! "$override_dns" =~ ^[Yy]$ ]]; then
          exit 1
        fi
      fi
    fi

    if [ -z "$email" ]; then
      prompt_user "Enter Email for Let's Encrypt certificate notices [default: admin@${domain}]: " email
      email=${email:-"admin@${domain}"}
    fi

    bind_host="127.0.0.1"
  fi

  local conf_ufw="${ARG_CONF_UFW}"
  if [ -z "$conf_ufw" ]; then
    prompt_user "Would you like to configure UFW firewall rules automatically? (y/n) [default: y]: " conf_ufw
    conf_ufw=${conf_ufw:-y}
  fi

  if [ "$ARG_USE_TURNSTILE" = "n" ]; then
    TURNSTILE_ENABLED="false"
  else
    prompt_turnstile_setup
  fi

  # Clear terminal for clean, premium progress view
  print_banner
  echo -e "${BOLD}Installing NeoPanel Master Panel...${NC} (Silent background execution, logs -> ${CYAN}${LOG_FILE}${NC})\n"

  # Run steps with smooth percentage progress
  run_step 0 10 "Updating system & installing base utilities" "step_prerequisites"
  run_step 10 15 "Installing & testing Bubblewrap sandbox" "step_isolation fresh"
  log_info "Bubblewrap host check passed. Select Bubblewrap on the node to require filesystem isolation."
  run_step 15 35 "Installing Minecraft Java runtimes (25, 21, 17 & 8)" "step_java"
  run_step 35 55 "Preparing Go compilation toolchain" "step_go_compiler"
  run_step 55 75 "Compiling NeoPanel panel & daemon binary" "step_build_neopanel"
  run_step 75 85 "Creating configuration & systemd service" "step_config_service '$bind_host' '$PANEL_PORT' '$TURNSTILE_ENABLED' '$TURNSTILE_SITE_KEY' '$TURNSTILE_SECRET_KEY' 'panel' ''"

  if [[ "$use_ssl" =~ ^[Yy]$ ]]; then
    run_step 85 95 "Setting up automated Let's Encrypt SSL & Caddy" "step_ssl_caddy '$domain' '$email'"
  fi

  if [[ "$conf_ufw" =~ ^[Yy]$ ]]; then
    run_step 95 100 "Applying UFW firewall security rules" "step_firewall 'true'"
  else
    draw_progress 100 "✔ Installation completed"
    echo ""
  fi

  # -----------------------------------------------------------------------------
  # Completion Banner
  # -----------------------------------------------------------------------------
  echo -e "\n${GREEN}${BOLD}================================================================${NC}"
  echo -e "${GREEN}${BOLD}           NeoPanel Installation Successfully Completed!         ${NC}"
  echo -e "${GREEN}${BOLD}================================================================${NC}\n"

  if [[ "$use_ssl" =~ ^[Yy]$ ]]; then
    echo -e "  ${BOLD}Web Dashboard:${NC}   ${CYAN}https://${domain}${NC}"
  else
    echo -e "  ${BOLD}Web Dashboard:${NC}   ${CYAN}http://${pub_ip}:${PANEL_PORT}${NC}"
  fi

  echo -e "  ${BOLD}Initial Setup:${NC}   Open the URL above in your browser to create the Admin account."
  echo -e "  ${BOLD}Config Path:${NC}     ${CONFIG_DIR}/config.yaml"
  echo -e "  ${BOLD}Data Path:${NC}       ${DATA_DIR}"

  echo -e "\n  ${CYAN}${BOLD}[CLUSTER] Expanding Your Infrastructure with Cluster Nodes:${NC}"
  echo -e "    1. Create an Admin API key in NeoPanel (Account -> API Keys) OR on Master terminal:"
  echo -e "       ${YELLOW}sudo neopanel api-key create --name \"Node Join Key\"${NC}"
  echo -e "    2. Run on any remote Ubuntu server to deploy a Cluster Node & auto-pair:"
  echo -e "       ${YELLOW}curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash -s -- --node --master <PANEL_URL> --api-key <neo_...> --name node-01${NC}"

  echo -e "\n  ${YELLOW}${BOLD}[REMINDER] Cloud Firewalls & Security Groups:${NC}"
  echo -e "    Remember to also allow incoming ports in your cloud provider's web console:"
  echo -e "      - ${CYAN}Port 80 / 443 (TCP)${NC}       -> Web Dashboard"
  echo -e "      - ${CYAN}Port 2022 (TCP)${NC}           -> Pure-Go Built-in SFTP Server"
  echo -e "      - ${CYAN}Port 25565-25600 (TCP/UDP)${NC} -> Minecraft Player Connections"

  if [ -f /var/run/reboot-required ]; then
    echo -e "\n  ${YELLOW}${BOLD}[NOTE] Pending Kernel Updates:${NC}"
    echo -e "    Your system has pending updates. Consider running 'sudo reboot' when convenient."
  fi

  echo -e "\n  ${BOLD}Useful Commands:${NC}"
  echo -e "    ${YELLOW}sudo neopanel-update${NC}       # Instant one-command update from GitHub"
  echo -e "    ${YELLOW}sudo neopanel help${NC}         # Manage API keys & cluster nodes via CLI"
  echo -e "    ${YELLOW}systemctl status neopanel${NC}    # View status"
  echo -e "    ${YELLOW}journalctl -u neopanel -f${NC}    # View live logs"
  echo -e "\n${GREEN}Thank you for choosing NeoPanel!${NC}\n"
}

# -----------------------------------------------------------------------------
# Dedicated Cluster Node Daemon Installation Flow
# -----------------------------------------------------------------------------

run_node_installation() {
  INSTALL_ROLE="node"
  print_banner
  check_root
  check_os

  local pub_ip default_hostname streamlined
  pub_ip=$(get_public_ip)
  default_hostname=$(hostname -s 2>/dev/null || hostname 2>/dev/null || echo "mc-node-01")

  streamlined="false"
  if [ "$ASSUME_YES" = "true" ] || { [ -n "$MASTER_URL" ] && [ -n "$MASTER_API_KEY" ]; } || { [ ! -t 0 ] && [ ! -e /dev/tty ]; }; then
    streamlined="true"
  fi

  echo -e "${BOLD}Mode:${NC} ${MAGENTA}Dedicated Cluster Game Node (-tags node)${NC}"
  echo -e "${BOLD}Detected Host IP:${NC} ${CYAN}${pub_ip}${NC}\n"

  local node_name="${NODE_NAME}"
  local node_loc="${NODE_LOCATION}"
  local use_ssl="${ARG_USE_SSL}"
  local domain="${ARG_DOMAIN}"
  local email="${ARG_EMAIL}"
  local bind_host="0.0.0.0"

  if [ -z "$node_name" ]; then
    if [ "$streamlined" = "true" ]; then
      node_name="$default_hostname"
    else
      prompt_user "Enter Node Name [default: ${default_hostname}]: " node_name
      node_name=${node_name:-$default_hostname}
    fi
  fi

  if [ -z "$node_loc" ] && [ "$streamlined" != "true" ]; then
    prompt_user "Enter Region / Location label (optional, e.g. Frankfurt, DE): " node_loc
  fi

  if [ -n "$domain" ]; then
    use_ssl="y"
  elif [ -z "$use_ssl" ]; then
    if [ "$streamlined" = "true" ]; then
      use_ssl="n"
    else
      echo -e "\n${YELLOW}Optionally protect this node's Daemon API with a domain & Let's Encrypt SSL.${NC}"
      prompt_user "Use Domain & SSL for this Node API? (y/n) [default: n — uses direct HTTP on port ${PANEL_PORT}]: " use_ssl
      use_ssl=${use_ssl:-n}
    fi
  fi

  if [[ "$use_ssl" =~ ^[Yy]$ ]]; then
    if [ -z "$domain" ]; then
      echo -e "\n${CYAN}[REMINDER] Ensure DNS 'A' record for your node subdomain points to ${pub_ip}.${NC}"
      prompt_user "Enter Node Domain/Subdomain (e.g. eu1.example.com): " domain
      while [ -z "$domain" ]; do
        prompt_user "Domain cannot be empty. Enter Domain: " domain
      done
    fi
    if [ -z "$email" ]; then
      if [ "$streamlined" = "true" ]; then
        email="admin@${domain}"
      else
        prompt_user "Enter Email for Let's Encrypt certificate [default: admin@${domain}]: " email
        email=${email:-"admin@${domain}"}
      fi
    fi
    bind_host="127.0.0.1"
  fi

  local default_adv_url="http://${pub_ip}:${PANEL_PORT}"
  if [[ "$use_ssl" =~ ^[Yy]$ ]]; then
    default_adv_url="https://${domain}"
  fi

  local node_adv_url="${NODE_ADVERTISE_URL}"
  if [ -z "$node_adv_url" ]; then
    if [ "$streamlined" = "true" ]; then
      node_adv_url="$default_adv_url"
    else
      echo -e "\n${CYAN}Reachable Daemon API URL (used by Master Panel to control this node):${NC}"
      prompt_user "Daemon URL [default: ${default_adv_url}]: " node_adv_url
      node_adv_url=${node_adv_url:-$default_adv_url}
    fi
  fi

  local default_fqdn="${pub_ip}"
  if [[ "$use_ssl" =~ ^[Yy]$ ]]; then
    default_fqdn="${domain}"
  fi
  local node_fqdn="${NODE_FQDN}"
  if [ -z "$node_fqdn" ]; then
    if [ "$streamlined" = "true" ]; then
      node_fqdn="$default_fqdn"
    else
      prompt_user "Public Game Host / IP shown to players [default: ${default_fqdn}]: " node_fqdn
      node_fqdn=${node_fqdn:-$default_fqdn}
    fi
  fi

  local conf_ufw="${ARG_CONF_UFW}"
  if [ -z "$conf_ufw" ]; then
    if [ "$streamlined" = "true" ]; then
      conf_ufw="y"
    else
      prompt_user "Configure UFW firewall rules automatically? (y/n) [default: y]: " conf_ufw
      conf_ufw=${conf_ufw:-y}
    fi
  fi

  if [ -z "$MASTER_URL" ] && [ "$streamlined" != "true" ]; then
    echo ""
    prompt_user "Pair automatically with your Master Panel right now using an API Key? (y/n) [default: n]: " pair_now
    if [[ "$pair_now" =~ ^[Yy]$ ]]; then
      prompt_user "Master Panel URL (e.g. https://panel.example.com): " MASTER_URL
      prompt_user "Master Admin API Key (neo_...): " MASTER_API_KEY
    fi
  fi

  local node_tok="${NODE_TOKEN_OVERRIDE}"
  if [ -z "$node_tok" ]; then
    node_tok=$(generate_node_token)
  fi

  print_banner
  echo -e "${BOLD}Installing NeoPanel Cluster Node Daemon...${NC} (Logs -> ${CYAN}${LOG_FILE}${NC})\n"

  run_step 0 10 "Updating system & installing base utilities" "step_prerequisites"
  run_step 10 15 "Installing & testing Bubblewrap sandbox" "step_isolation fresh"
  log_info "Bubblewrap host check passed. Select Bubblewrap on the node to require filesystem isolation."
  run_step 15 35 "Installing Minecraft Java runtimes (25, 21, 17 & 8)" "step_java"
  run_step 35 55 "Preparing Go compilation toolchain" "step_go_compiler"
  run_step 55 75 "Compiling NeoPanel Node Daemon (-tags node)" "step_build_neopanel"
  run_step 75 85 "Configuring Cluster Node Daemon & systemd" "step_config_service '$bind_host' '$PANEL_PORT' 'false' '' '' 'node' '$node_tok'"

  local fw_include_http="false"
  if [[ "$use_ssl" =~ ^[Yy]$ ]]; then
    fw_include_http="true"
    run_step 85 95 "Setting up automated Let's Encrypt SSL & Caddy" "step_ssl_caddy '$domain' '$email'"
  fi

  if [[ "$conf_ufw" =~ ^[Yy]$ ]]; then
    run_step 95 100 "Applying UFW firewall security rules" "step_firewall '$fw_include_http'"
  else
    draw_progress 100 "✔ Cluster Node installation completed"
    echo ""
  fi

  wait_for_local_daemon "$PANEL_PORT" || true

  local pair_status="not_attempted"
  local pair_output=""
  if [ -n "$MASTER_URL" ] && [ -n "$MASTER_API_KEY" ]; then
    echo -e "\n${BLUE}[PAIRING]${NC} Registering node '${node_name}' with Master Panel ${CYAN}${MASTER_URL}${NC}..."
    local -a pair_cmd=(
      "$INSTALL_DIR/$BINARY_NAME" -config "$CONFIG_DIR/config.yaml"
      node connect
      --master "$MASTER_URL"
      --api-key "$MASTER_API_KEY"
      --name "$node_name"
      --url "$node_adv_url"
      --token "$node_tok"
      --fqdn "$node_fqdn"
      --location "$node_loc"
    )
    if [ "$PAIR_INSECURE" = "true" ]; then
      pair_cmd+=(--insecure)
    fi

    if pair_output=$("${pair_cmd[@]}" 2>&1); then
      pair_status="ok"
      log_success "Node paired with Master Panel!"
      echo "$pair_output" | sed 's/^/  /'
    else
      pair_status="failed"
      log_warn "Automatic pairing did not succeed yet: ${pair_output}"
    fi
  fi

  # -----------------------------------------------------------------------------
  # Node Completion Banner
  # -----------------------------------------------------------------------------
  echo -e "\n${GREEN}${BOLD}================================================================${NC}"
  echo -e "${GREEN}${BOLD}         NeoPanel Cluster Node Installation Complete!           ${NC}"
  echo -e "${GREEN}${BOLD}================================================================${NC}\n"

  echo -e "  ${BOLD}Daemon Role:${NC}      ${MAGENTA}node (Cluster Node Daemon + SFTP)${NC}"
  echo -e "  ${BOLD}Node Name:${NC}        ${CYAN}${node_name}${NC}"
  echo -e "  ${BOLD}API Endpoint:${NC}     ${CYAN}${node_adv_url}${NC}"
  echo -e "  ${BOLD}Player Host/FQDN:${NC} ${CYAN}${node_fqdn}${NC}"
  echo -e "  ${BOLD}Node Token:${NC}       ${YELLOW}${node_tok}${NC}"
  echo -e "  ${BOLD}Config Path:${NC}      ${CONFIG_DIR}/config.yaml"
  echo -e "  ${BOLD}Data Path:${NC}        ${DATA_DIR}"

  if [ "$pair_status" != "ok" ]; then
    echo -e "\n  ${CYAN}${BOLD}Connect this node to your Master Panel using EITHER method:${NC}"
    echo -e "    ${BOLD}1. On Master VPS CLI:${NC}"
    echo -e "       ${YELLOW}sudo neopanel node connect --name \"${node_name}\" --url \"${node_adv_url}\" --fqdn \"${node_fqdn}\" --location \"${node_loc}\" --token \"${node_tok}\"${NC}"
    echo -e "    ${BOLD}2. From this Node using a Master API Key (neo_...):${NC}"
    echo -e "       ${YELLOW}sudo neopanel node connect --master https://panel.example.com --api-key <neo_...> --name \"${node_name}\" --url \"${node_adv_url}\" --fqdn \"${node_fqdn}\"${NC}"
    echo -e "    ${BOLD}3. Master Panel UI:${NC} Admin -> Cluster Nodes -> Connect Node"
  fi

  echo -e "\n  ${YELLOW}${BOLD}[SECURITY TIP]${NC}"
  echo -e "    This node is seamlessly managed by your Master Panel and authenticates via ${YELLOW}node_token${NC}."
  if [[ ! "$use_ssl" =~ ^[Yy]$ ]]; then
    echo -e "    For extra hardening on public VPS hosts, restrict port ${PANEL_PORT}/tcp in UFW to your Master IP:"
    echo -e "      ${CYAN}sudo ufw delete allow ${PANEL_PORT}/tcp && sudo ufw allow from <MASTER_IP> to any port ${PANEL_PORT} proto tcp${NC}"
  fi

  echo -e "\n  ${BOLD}Useful Node Commands:${NC}"
  echo -e "    ${YELLOW}sudo neopanel node-token${NC}            # Show node auth token"
  echo -e "    ${YELLOW}sudo neopanel node-token --rotate${NC}   # Rotate token cleanly"
  echo -e "    ${YELLOW}sudo neopanel-update${NC}                # Update cluster node binary (-tags node preserved)"
  echo -e "    ${YELLOW}systemctl status neopanel${NC}           # View service health"
  echo -e "\n${GREEN}Cluster node ready!${NC}\n"
}

# -----------------------------------------------------------------------------
# Interactive Connect Remote Node Wizard
# -----------------------------------------------------------------------------

interactive_connect_node_wizard() {
  check_root
  print_banner
  echo -e "${CYAN}${BOLD}Connect Remote Node Wizard${NC}\n"

  prompt_user "Enter Node Name (e.g. EU-Frankfurt-01): " c_name
  while [ -z "$c_name" ]; do
    prompt_user "Node name cannot be empty: " c_name
  done
  prompt_user "Enter Remote Node Daemon URL (e.g. http://192.168.1.50:8080 or https://eu1.example.com): " c_url
  while [ -z "$c_url" ]; do
    prompt_user "Node URL cannot be empty: " c_url
  done
  prompt_user "Enter Remote Node Token (neo_node_... or neo_...): " c_tok
  while [ -z "$c_tok" ]; do
    prompt_user "Token cannot be empty: " c_tok
  done
  prompt_user "Public Game Host/FQDN for players [optional, defaults to URL host]: " c_fqdn
  prompt_user "Region / Location [optional, e.g. Frankfurt, DE]: " c_loc
  prompt_user "Register on local Master DB or remote Master URL? [leave blank for local DB]: " c_master

  local -a cmd=("$INSTALL_DIR/$BINARY_NAME" -config "$CONFIG_DIR/config.yaml" node connect --name "$c_name" --url "$c_url" --token "$c_tok")
  [ -n "$c_fqdn" ] && cmd+=(--fqdn "$c_fqdn")
  [ -n "$c_loc" ] && cmd+=(--location "$c_loc")
  if [ -n "$c_master" ]; then
    prompt_user "Enter Master Panel Admin API Key (neo_...): " c_mkey
    cmd+=(--master "$c_master" --api-key "$c_mkey")
  fi

  echo ""
  "${cmd[@]}"
}

# -----------------------------------------------------------------------------
# Update Flow (Preserves Master Panel vs Cluster Node Role Automatically)
# -----------------------------------------------------------------------------

step_update_ensure_java() {
  if ! has_java_version "21" || ! has_java_version "17" || ! has_java_version "8"; then
    apt-get update -y -q || true
    apt-get install -y -q -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" \
      openjdk-21-jre-headless openjdk-17-jre-headless openjdk-8-jre-headless || true
  fi
  if ! has_java_version "25"; then
    if ! command -v curl &>/dev/null || ! command -v tar &>/dev/null; then
      apt-get update -y -q && apt-get install -y -q curl tar ca-certificates || true
    fi
    install_java_25 || true
  fi
}

step_update_fetch() {
  if ! command -v git &>/dev/null; then
    apt-get update -y -q && apt-get install -y -q git
  fi
  if ! command -v go &>/dev/null; then
    apt-get update -y -q && apt-get install -y -q golang-go
  fi

  rm -rf "$BUILD_DIR"
  git clone --depth 1 --branch "$GIT_BRANCH" "$GIT_REPO_URL" "$BUILD_DIR"
}

step_update_compile() {
  local -a go_tags=()
  if [ "$INSTALL_ROLE" = "node" ]; then
    go_tags=(-tags node)
  fi
  local git_commit=""
  git_commit="$(cd "$BUILD_DIR" && git rev-parse HEAD 2>/dev/null || true)"
  (cd "$BUILD_DIR" && go mod tidy && go build "${go_tags[@]}" -ldflags="-s -w -X 'github.com/Pekly/NeoPanel/internal/api.BuildCommit=${git_commit}'" -o "$BUILD_DIR/neopanel_new" ./cmd/neopanel)
}

step_update_swap() {
  if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    cp "$INSTALL_DIR/$BINARY_NAME" "$INSTALL_DIR/${BINARY_NAME}.bak"
  fi

  local git_commit=""
  if [ -d "$BUILD_DIR" ]; then
    git_commit="$(cd "$BUILD_DIR" && git rev-parse HEAD 2>/dev/null || true)"
  fi
  if [ -n "$git_commit" ]; then
    mkdir -p "$CONFIG_DIR" "$DATA_DIR"
    echo "$git_commit" > "$CONFIG_DIR/current_commit"
    echo "$git_commit" > "$DATA_DIR/current_commit"
  fi

  mv "$BUILD_DIR/neopanel_new" "$INSTALL_DIR/$BINARY_NAME"
  chmod +x "$INSTALL_DIR/$BINARY_NAME"
  rm -rf "$BUILD_DIR"

  write_global_shortcuts

  if [ "$NEOPANEL_WEB_UPDATE" = "1" ]; then
    # Background service restart with a 2-second grace delay so the web updater process
    # and Go backend can cleanly finish writing the success log and return HTTP 200/exit 0!
    (sleep 2 && (systemd-run --no-block systemctl restart neopanel 2>/dev/null || systemctl restart neopanel)) >/dev/null 2>&1 &
  else
    systemctl restart neopanel || true
    sleep 1
    systemctl is-active --quiet neopanel || true
  fi
}

update_neopanel() {
  print_banner
  check_root
  detect_installed_role

  # Prompt for Cloudflare Turnstile ONLY in full Web Panel mode if upgrading from pre-Turnstile config
  if [ "$INSTALL_ROLE" != "node" ] && [ -f "$CONFIG_DIR/config.yaml" ] && ! grep -q "turnstile:" "$CONFIG_DIR/config.yaml"; then
    if [ -t 0 ] || [ -e /dev/tty ]; then
      echo -e "${CYAN}${BOLD}[NEW IN THIS VERSION] Cloudflare Turnstile Bot Protection${NC}"
      prompt_turnstile_setup
      configure_turnstile_in_config "$TURNSTILE_ENABLED" "$TURNSTILE_SITE_KEY" "$TURNSTILE_SECRET_KEY"
      echo ""
    fi
  fi

  if [ "$INSTALL_ROLE" = "node" ]; then
    echo -e "${BOLD}Updating NeoPanel Cluster Node Daemon (-tags node)...${NC} (Logs -> ${CYAN}${LOG_FILE}${NC})\n"
  else
    echo -e "${BOLD}Updating NeoPanel to Latest Release...${NC} (Logs -> ${CYAN}${LOG_FILE}${NC})\n"
  fi

  run_step 0 10 "Installing & testing Bubblewrap sandbox" "step_isolation installed"
  log_info "Bubblewrap host check passed; existing node isolation settings are unchanged."
  run_step 10 25 "Ensuring Java 25, 21, 17 & 8 runtimes" "step_update_ensure_java"
  run_step 25 45 "Fetching latest updates from GitHub" "step_update_fetch"
  if [ "$INSTALL_ROLE" = "node" ]; then
    run_step 45 80 "Compiling fresh Cluster Node binary (-tags node)" "step_update_compile"
  else
    run_step 45 80 "Compiling fresh NeoPanel binary" "step_update_compile"
  fi
  run_step 80 100 "Restarting service and verifying health" "step_update_swap"

  echo -e "\n${GREEN}${BOLD}================================================================${NC}"
  echo -e "${GREEN}${BOLD}           NeoPanel Successfully Updated to Latest Version!     ${NC}"
  echo -e "${GREEN}${BOLD}================================================================${NC}\n"
  echo -e "  ${BOLD}Role Preserved:${NC}   ${CYAN}${INSTALL_ROLE}${NC}"
  echo -e "  ${BOLD}Service Status:${NC}   ${GREEN}Active & Running (systemd)${NC}"
  echo -e "  ${BOLD}Data Safety:${NC}      All database records, servers & configs preserved.\n"
  echo -e "${GREEN}================================================================${NC}\n"
}

# -----------------------------------------------------------------------------
# Interactive Menu & CLI Argument Parsing
# -----------------------------------------------------------------------------

print_help() {
  cat <<'EOF'
NeoPanel Installer & Cluster Tooling

Usage:
  install.sh [options]

Installation Actions:
  --install, --panel        Install Master Panel (Dashboard + Daemon + Java)
  --node                    Install Dedicated Cluster Game Node (-tags node)
  --update, -u              Update existing installation (auto-preserves panel vs node build)
  --repo <url>              Git repository URL to fetch from (default: https://github.com/Pekly/NeoPanel.git)
  --branch <name>           Git branch to build from (default: main)

Cluster Node & Pairing Flags (combine with --node):
  --master <url>            Master Panel base URL (e.g. https://panel.example.com)
  --api-key <neo_...>       Master Admin API Key to auto-pair immediately after install
  --name <label>            Node display name (defaults to server hostname)
  --node-url <url>          Reachable API URL for this node (defaults to http://<ip>:8080 or https://<domain>)
  --fqdn <host>             Player-facing game IP/domain
  --location <region>       Region label (e.g. Frankfurt, DE)
  --node-token <neo_node_>  Preset node auth token (auto-generated if omitted)
  --port <port>             Daemon HTTP listening port (default: 8080)
  --domain <domain>         Configure Caddy + Let's Encrypt HTTPS for this host
  --email <email>           Let's Encrypt contact email
  --ssl / --no-ssl          Enable or skip Let's Encrypt SSL setup
  --firewall / --no-firewall Enable or skip UFW firewall setup
  --no-turnstile            Skip Turnstile prompt during Panel install
  --insecure                Allow self-signed HTTPS cert when pairing with --master
  -y, --yes                 Use defaults for any unspecified interactive prompts

Maintenance & CLI Shortcuts:
  --connect-node            Launch interactive remote node pairing wizard
  --show-token              Show this host's security.node_token
  --rotate-token            Rotate node_token and restart service
  --create-api-key [name]   Generate a new Admin API Key in Master SQLite
  --java                    Install/repair Java 25, 21, 17 & 8 runtimes
  --firewall                Apply UFW firewall rules
  --turnstile               Configure Cloudflare Turnstile on Panel
  -h, --help                Show this help
EOF
}

interactive_menu() {
  print_banner
  echo "Please select an option:"
  echo -e "  ${GREEN}[1]${NC}  Install Master Panel (Dashboard + Daemon + Java + SSL + Firewall)"
  echo -e "  ${GREEN}[2]${NC}  Install Dedicated Cluster Game Node (Daemon + SFTP + Java)"
  echo -e "  ${GREEN}[3]${NC}  Update NeoPanel to Latest Version (Preserves Panel / Node Mode)"
  echo -e "  ${GREEN}[4]${NC}  Connect / Pair a Remote Node (CLI Wizard)"
  echo -e "  ${GREEN}[5]${NC}  Show Node Token or Generate CLI API Key"
  echo -e "  ${GREEN}[6]${NC}  Install Java Runtimes (Java 25, 21, 17 & 8)"
  echo -e "  ${GREEN}[7]${NC}  Configure / Reconfigure Let's Encrypt SSL (Caddy)"
  echo -e "  ${GREEN}[8]${NC}  Configure UFW Firewall"
  echo -e "  ${GREEN}[9]${NC}  Configure Cloudflare Turnstile Bot Protection"
  echo -e "  ${GREEN}[10]${NC} Restart NeoPanel Service"
  echo -e "  ${GREEN}[11]${NC} Uninstall NeoPanel"
  echo -e "  ${RED}[0]${NC}  Exit"
  echo ""
  prompt_user "Enter selection [1]: " choice
  choice=${choice:-1}

  case "$choice" in
    1)
      run_full_installation
      ;;
    2)
      run_node_installation
      ;;
    3)
      update_neopanel
      ;;
    4)
      interactive_connect_node_wizard
      ;;
    5)
      check_root
      print_banner
      echo -e "  ${GREEN}[1]${NC} Show this host's node_token"
      echo -e "  ${GREEN}[2]${NC} Rotate node_token (and restart service)"
      echo -e "  ${GREEN}[3]${NC} Generate Admin API Key (neo_...) in Master DB"
      prompt_user "Choose action [1]: " tok_act
      tok_act=${tok_act:-1}
      case "$tok_act" in
        1)
          "$INSTALL_DIR/$BINARY_NAME" -config "$CONFIG_DIR/config.yaml" node-token
          ;;
        2)
          "$INSTALL_DIR/$BINARY_NAME" -config "$CONFIG_DIR/config.yaml" node-token --rotate
          systemctl restart neopanel || true
          ;;
        3)
          prompt_user "Enter API Key label [CLI Node Key]: " kname
          kname=${kname:-"CLI Node Key"}
          "$INSTALL_DIR/$BINARY_NAME" -config "$CONFIG_DIR/config.yaml" api-key create --name "$kname"
          ;;
        *)
          log_error "Invalid option."
          exit 1
          ;;
      esac
      ;;
    6)
      check_root
      print_banner
      run_step 0 100 "Installing Java 25, 21, 17 & 8 runtimes" "step_java"
      log_success "Java installed successfully."
      ;;
    7)
      check_root
      prompt_user "Enter Domain: " dom
      prompt_user "Enter Email: " em
      print_banner
      run_step 0 100 "Configuring Let's Encrypt SSL & Caddy" "step_ssl_caddy '$dom' '$em'"
      log_success "SSL configured."
      ;;
    8)
      check_root
      print_banner
      run_step 0 100 "Configuring UFW Firewall" "step_firewall 'true'"
      log_success "Firewall configured."
      ;;
    9)
      check_root
      print_banner
      prompt_turnstile_setup
      configure_turnstile_in_config "$TURNSTILE_ENABLED" "$TURNSTILE_SITE_KEY" "$TURNSTILE_SECRET_KEY"
      systemctl restart neopanel || true
      log_success "Cloudflare Turnstile configuration saved and NeoPanel restarted."
      ;;
    10)
      systemctl restart neopanel
      log_success "NeoPanel restarted."
      ;;
    11)
      check_root
      prompt_user "Are you sure you want to completely uninstall NeoPanel? (y/N): " uninst
      if [[ "$uninst" =~ ^[Yy]$ ]]; then
        systemctl stop neopanel || true
        systemctl disable neopanel || true
        rm -f "$SERVICE_FILE"
        systemctl daemon-reload
        rm -f "$INSTALL_DIR/$BINARY_NAME"
        rm -f "$INSTALL_DIR/neopanel-update"
        rm -f "$INSTALL_DIR/neopanel-installer"
        log_success "NeoPanel service and binary removed. (Data kept at $DATA_DIR for safety)"
      fi
      ;;
    0)
      exit 0
      ;;
    *)
      log_error "Invalid selection."
      exit 1
      ;;
  esac
}

parse_cli_args() {
  while [ $# -gt 0 ]; do
    case "$1" in
      --install|--panel)
        ACTION="${ACTION:-install}"
        INSTALL_ROLE="panel"
        shift
        ;;
      --node)
        INSTALL_ROLE="node"
        if [ -z "$ACTION" ]; then
          ACTION="node"
        fi
        shift
        ;;
      --update|-u)
        ACTION="update"
        shift
        ;;
      --repo|--git-repo)
        GIT_REPO_URL="${2:-}"
        shift 2
        ;;
      --repo=*|--git-repo=*)
        GIT_REPO_URL="${1#*=}"
        shift
        ;;
      --branch|--git-branch)
        GIT_BRANCH="${2:-}"
        shift 2
        ;;
      --branch=*|--git-branch=*)
        GIT_BRANCH="${1#*=}"
        shift
        ;;
      --master)
        MASTER_URL="${2:-}"
        shift 2
        ;;
      --master=*)
        MASTER_URL="${1#*=}"
        shift
        ;;
      --api-key|--key)
        MASTER_API_KEY="${2:-}"
        shift 2
        ;;
      --api-key=*|--key=*)
        MASTER_API_KEY="${1#*=}"
        shift
        ;;
      --name|--node-name)
        NODE_NAME="${2:-}"
        shift 2
        ;;
      --name=*|--node-name=*)
        NODE_NAME="${1#*=}"
        shift
        ;;
      --node-url|--url)
        NODE_ADVERTISE_URL="${2:-}"
        shift 2
        ;;
      --node-url=*|--url=*)
        NODE_ADVERTISE_URL="${1#*=}"
        shift
        ;;
      --fqdn)
        NODE_FQDN="${2:-}"
        shift 2
        ;;
      --fqdn=*)
        NODE_FQDN="${1#*=}"
        shift
        ;;
      --location)
        NODE_LOCATION="${2:-}"
        shift 2
        ;;
      --location=*)
        NODE_LOCATION="${1#*=}"
        shift
        ;;
      --node-token|--token)
        NODE_TOKEN_OVERRIDE="${2:-}"
        shift 2
        ;;
      --node-token=*|--token=*)
        NODE_TOKEN_OVERRIDE="${1#*=}"
        shift
        ;;
      --port)
        PANEL_PORT="${2:-8080}"
        shift 2
        ;;
      --port=*)
        PANEL_PORT="${1#*=}"
        shift
        ;;
      --domain)
        ARG_DOMAIN="${2:-}"
        ARG_USE_SSL="y"
        shift 2
        ;;
      --domain=*)
        ARG_DOMAIN="${1#*=}"
        ARG_USE_SSL="y"
        shift
        ;;
      --email)
        ARG_EMAIL="${2:-}"
        shift 2
        ;;
      --email=*)
        ARG_EMAIL="${1#*=}"
        shift
        ;;
      --ssl)
        ARG_USE_SSL="y"
        shift
        ;;
      --no-ssl)
        ARG_USE_SSL="n"
        shift
        ;;
      --firewall)
        if [ -z "$ACTION" ] && [ $# -eq 1 ]; then
          ACTION="firewall"
        else
          ARG_CONF_UFW="y"
        fi
        shift
        ;;
      --no-firewall)
        ARG_CONF_UFW="n"
        shift
        ;;
      --no-turnstile)
        ARG_USE_TURNSTILE="n"
        shift
        ;;
      --insecure)
        PAIR_INSECURE="true"
        shift
        ;;
      -y|--yes)
        ASSUME_YES="true"
        shift
        ;;
      --java)
        ACTION="java"
        shift
        ;;
      --turnstile)
        ACTION="turnstile"
        shift
        ;;
      --connect-node)
        ACTION="connect-node"
        shift
        ;;
      --show-token)
        ACTION="show-token"
        shift
        ;;
      --rotate-token)
        ACTION="rotate-token"
        shift
        ;;
      --create-api-key)
        ACTION="create-api-key"
        if [ $# -ge 2 ] && [[ "${2}" != -* ]]; then
          ARG_API_KEY_NAME="${2}"
          shift 2
        else
          shift
        fi
        ;;
      -h|--help)
        ACTION="help"
        shift
        ;;
      *)
        shift
        ;;
    esac
  done
}

# Entrypoint
if [ $# -eq 0 ]; then
  interactive_menu
else
  parse_cli_args "$@"
  case "$ACTION" in
    install)
      run_full_installation
      ;;
    node)
      run_node_installation
      ;;
    update)
      update_neopanel
      ;;
    java)
      check_root
      run_step 0 100 "Installing Java 25, 21, 17 & 8 runtimes" "step_java"
      ;;
    firewall)
      check_root
      run_step 0 100 "Configuring UFW Firewall" "step_firewall 'true'"
      ;;
    turnstile)
      check_root
      print_banner
      prompt_turnstile_setup
      configure_turnstile_in_config "$TURNSTILE_ENABLED" "$TURNSTILE_SITE_KEY" "$TURNSTILE_SECRET_KEY"
      systemctl restart neopanel || true
      log_success "Cloudflare Turnstile configuration saved and NeoPanel restarted."
      ;;
    connect-node)
      interactive_connect_node_wizard
      ;;
    show-token)
      check_root
      "$INSTALL_DIR/$BINARY_NAME" -config "$CONFIG_DIR/config.yaml" node-token
      ;;
    rotate-token)
      check_root
      "$INSTALL_DIR/$BINARY_NAME" -config "$CONFIG_DIR/config.yaml" node-token --rotate
      systemctl restart neopanel || true
      ;;
    create-api-key)
      check_root
      if [ -n "$ARG_API_KEY_NAME" ]; then
        "$INSTALL_DIR/$BINARY_NAME" -config "$CONFIG_DIR/config.yaml" api-key create --name "$ARG_API_KEY_NAME"
      else
        "$INSTALL_DIR/$BINARY_NAME" -config "$CONFIG_DIR/config.yaml" api-key create
      fi
      ;;
    help)
      print_help
      ;;
    *)
      interactive_menu
      ;;
  esac
fi
