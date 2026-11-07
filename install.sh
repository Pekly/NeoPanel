#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# NeoPanel All-In-One Installer for Ubuntu 24.04 LTS
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

# Default values
PANEL_PORT="8080"
CONFIG_DIR="/etc/neopanel"
DATA_DIR="/var/lib/neopanel"
SERVICE_FILE="/etc/systemd/system/neopanel.service"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="neopanel"
LOG_FILE="/tmp/neopanel-install.log"
BUILD_DIR="/tmp/neopanel-build"

# Prevent apt interactive prompts from hanging in background
export DEBIAN_FRONTEND=noninteractive
export NEEDRESTART_MODE=a

# Initialize log file
: > "$LOG_FILE"

# -----------------------------------------------------------------------------
# Premium UI & Animated Progress System
# -----------------------------------------------------------------------------

print_banner() {
  clear
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
  echo -e "${YELLOW}Target: Ubuntu 24.04 LTS (Noble Numbat)${NC}\n"
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
    read -r -p "$prompt_text" "$var_name"
  elif [ -e /dev/tty ]; then
    read -r -p "$prompt_text" "$var_name" </dev/tty
  else
    read -r -p "$prompt_text" "$var_name"
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
  printf "\r\033[K  ${CYAN}[${GREEN}%s${CYAN}] ${BOLD}%3d%%${NC}  %-48s" "$bar" "$pct" "$text"
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
    prompt_user "Do you want to continue anyway? (y/N): " choice
    if [[ ! "$choice" =~ ^[Yy]$ ]]; then
      exit 1
    fi
  fi
}

get_public_ip() {
  curl -s4 https://ifconfig.me || curl -s4 https://api.ipify.org || echo "127.0.0.1"
}

generate_secret() {
  openssl rand -hex 32 2>/dev/null || date +%s | sha256sum | head -c 64
}

# -----------------------------------------------------------------------------
# Background Installation Steps
# -----------------------------------------------------------------------------

step_prerequisites() {
  apt-get update -y -q
  apt-get install -y -q -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" \
    curl wget jq tar ufw ca-certificates software-properties-common git
}

step_java() {
  apt-get update -y -q
  apt-get install -y -q -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" \
    openjdk-21-jre-headless openjdk-17-jre-headless || true
  apt-get install -y -q -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" \
    openjdk-8-jre-headless || true

  # Install Java 25 (Cutting edge / Experimental) via Adoptium if not already present
  if [ ! -x "/usr/lib/jvm/java-25-temurin/bin/java" ] && [ ! -x "/usr/lib/jvm/java-25-openjdk-amd64/bin/java" ]; then
    local arch="x64"
    if [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "arm64" ]; then
      arch="aarch64"
    fi
    mkdir -p /usr/lib/jvm/java-25-temurin
    local j25_url="https://api.adoptium.net/v3/binary/latest/25/ga/linux/${arch}/jdk/hotspot/normal/eclipse"
    if ! curl -sIL "$j25_url" | grep -qi "HTTP.*[23]0[0-9]"; then
      j25_url="https://api.adoptium.net/v3/binary/latest/25/ea/linux/${arch}/jdk/hotspot/normal/eclipse"
    fi
    curl -sL "$j25_url" | tar -xz -C /usr/lib/jvm/java-25-temurin --strip-components=1 2>/dev/null || true
  fi
}

step_go_compiler() {
  if ! command -v go &>/dev/null; then
    apt-get update -y -q
    apt-get install -y -q -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" golang-go
  fi
}

step_build_neopanel() {
  mkdir -p "$CONFIG_DIR"
  mkdir -p "$DATA_DIR/servers"
  mkdir -p "$DATA_DIR/backups"

  if [ -f "./cmd/neopanel/main.go" ]; then
    go mod tidy
    go build -ldflags="-s -w" -o "$INSTALL_DIR/$BINARY_NAME" ./cmd/neopanel
  else
    rm -rf "$BUILD_DIR"
    git clone https://github.com/Pekly/NeoPanel.git "$BUILD_DIR"
    if [ -f "$BUILD_DIR/cmd/neopanel/main.go" ]; then
      (cd "$BUILD_DIR" && go mod tidy && go build -ldflags="-s -w" -o "$INSTALL_DIR/$BINARY_NAME" ./cmd/neopanel)
      rm -rf "$BUILD_DIR"
    else
      echo "Failed to clone repository or main.go missing." >&2
      exit 1
    fi
  fi

  chmod +x "$INSTALL_DIR/$BINARY_NAME"

  # Install global neopanel-update command shortcut
  cat > "$INSTALL_DIR/neopanel-update" <<'EOF'
#!/usr/bin/env bash
set -e
curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | bash -s -- --update
EOF
  chmod +x "$INSTALL_DIR/neopanel-update"
}

step_config_service() {
  local host="$1"
  local port="$2"

  local secret
  secret=$(generate_secret)

  cat > "$CONFIG_DIR/config.yaml" <<EOF
server:
  host: "${host}"
  port: ${port}
  mode: "behind_proxy"

paths:
  data_dir: "${DATA_DIR}"
  servers_dir: "${DATA_DIR}/servers"
  db_path: "${DATA_DIR}/neopanel.db"

security:
  jwt_secret: "${secret}"
  token_expiration_hours: 72

java:
  default_path: "java"
EOF

  chmod 600 "$CONFIG_DIR/config.yaml"

  cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=NeoPanel Minecraft Server Manager
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
  ufw allow 22/tcp >/dev/null 2>&1 || true
  ufw allow 80/tcp >/dev/null 2>&1
  ufw allow 443/tcp >/dev/null 2>&1
  ufw allow 2022/tcp >/dev/null 2>&1
  ufw allow "${PANEL_PORT}/tcp" >/dev/null 2>&1
  ufw allow 25565:25600/tcp >/dev/null 2>&1
  ufw allow 25565:25600/udp >/dev/null 2>&1
  echo "y" | ufw enable >/dev/null 2>&1
}

# -----------------------------------------------------------------------------
# Main Install Flow
# -----------------------------------------------------------------------------

run_full_installation() {
  print_banner
  check_root
  check_os

  local pub_ip
  pub_ip=$(get_public_ip)

  echo -e "${BOLD}Server Public IP detected:${NC} ${CYAN}${pub_ip}${NC}\n"

  echo -e "${YELLOW}Would you like to configure a custom domain with automated Let's Encrypt SSL?${NC}"
  prompt_user "Use Domain & SSL? (y/n) [default: y]: " use_ssl
  use_ssl=${use_ssl:-y}

  local domain=""
  local email=""
  local bind_host="0.0.0.0"

  if [[ "$use_ssl" =~ ^[Yy]$ ]]; then
    echo -e "\n${CYAN}[REMINDER] Ensure your domain's DNS 'A' record points to this server's IP (${pub_ip}).${NC}"
    echo -e "${YELLOW}           If DNS has not propagated yet, Let's Encrypt certificate issuance will fail.${NC}\n"

    prompt_user "Enter your Domain name (e.g. panel.example.com): " domain
    while [ -z "$domain" ]; do
      prompt_user "Domain cannot be empty. Enter Domain: " domain
    done

    # Quick DNS check
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

    prompt_user "Enter Email for Let's Encrypt certificate notices: " email
    while [ -z "$email" ]; do
      prompt_user "Email cannot be empty. Enter Email: " email
    done

    bind_host="127.0.0.1"
  fi

  prompt_user "Would you like to configure UFW firewall rules automatically? (y/n) [default: y]: " conf_ufw
  conf_ufw=${conf_ufw:-y}

  # Clear terminal for clean, premium progress view
  print_banner
  echo -e "${BOLD}Installing NeoPanel...${NC} (Silent background execution, logs -> ${CYAN}${LOG_FILE}${NC})\n"

  # Run steps with smooth percentage progress
  run_step 0 15 "Updating system & installing base utilities" "step_prerequisites"
  run_step 15 35 "Installing Minecraft Java runtimes (8, 17, 21 & 25)" "step_java"
  run_step 35 55 "Preparing Go compilation toolchain" "step_go_compiler"
  run_step 55 75 "Compiling NeoPanel core binary from source" "step_build_neopanel"
  run_step 75 85 "Creating configuration & systemd service" "step_config_service '$bind_host' '$PANEL_PORT'"

  if [[ "$use_ssl" =~ ^[Yy]$ ]]; then
    run_step 85 95 "Setting up automated Let's Encrypt SSL & Caddy" "step_ssl_caddy '$domain' '$email'"
  fi

  if [[ "$conf_ufw" =~ ^[Yy]$ ]]; then
    run_step 95 100 "Applying UFW firewall security rules" "step_firewall"
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

  echo -e "\n  ${YELLOW}${BOLD}[REMINDER] Cloud Firewalls & Security Groups:${NC}"
  echo -e "    If your VPS is hosted on AWS, Oracle Cloud, Google Cloud, Azure, or OVH,"
  echo -e "    remember to also allow incoming ports in your cloud provider's web console:"
  echo -e "      - ${CYAN}Port 80 / 443 (TCP)${NC}       -> Web Dashboard"
  echo -e "      - ${CYAN}Port 2022 (TCP)${NC}           -> Pure-Go Built-in SFTP Server"
  echo -e "      - ${CYAN}Port 25565-25600 (TCP/UDP)${NC} -> Minecraft Player Connections"

  echo -e "\n  ${CYAN}${BOLD}[REMINDER] Minecraft Server Ports:${NC}"
  echo -e "    Your first server can use default port 25565."
  echo -e "    Each subsequent server requires its own unique port (e.g. 25566, 25567)."

  if [ -f /var/run/reboot-required ]; then
    echo -e "\n  ${YELLOW}${BOLD}[NOTE] Pending Kernel Updates:${NC}"
    echo -e "    Your system has pending updates. Consider running 'sudo reboot' when convenient."
  fi

  echo -e "\n  ${BOLD}Useful Commands:${NC}"
  echo -e "    ${YELLOW}sudo neopanel-update${NC}       # Instant one-command update from GitHub"
  echo -e "    ${YELLOW}systemctl status neopanel${NC}    # View status"
  echo -e "    ${YELLOW}journalctl -u neopanel -f${NC}    # View live logs"
  echo -e "    ${YELLOW}systemctl restart neopanel${NC}   # Restart panel"
  echo -e "\n${GREEN}Thank you for choosing NeoPanel!${NC}\n"
}

# -----------------------------------------------------------------------------
# Update Flow
# -----------------------------------------------------------------------------

step_update_fetch() {
  if ! command -v git &>/dev/null; then
    apt-get update -y -q && apt-get install -y -q git
  fi
  if ! command -v go &>/dev/null; then
    apt-get update -y -q && apt-get install -y -q golang-go
  fi

  rm -rf "$BUILD_DIR"
  git clone --depth 1 https://github.com/Pekly/NeoPanel.git "$BUILD_DIR"
}

step_update_compile() {
  (cd "$BUILD_DIR" && go mod tidy && go build -ldflags="-s -w" -o "$BUILD_DIR/neopanel_new" ./cmd/neopanel)
}

step_update_swap() {
  if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    cp "$INSTALL_DIR/$BINARY_NAME" "$INSTALL_DIR/${BINARY_NAME}.bak"
  fi

  mv "$BUILD_DIR/neopanel_new" "$INSTALL_DIR/$BINARY_NAME"
  chmod +x "$INSTALL_DIR/$BINARY_NAME"
  rm -rf "$BUILD_DIR"

  systemctl restart neopanel
  sleep 1
  systemctl is-active --quiet neopanel
}

update_neopanel() {
  print_banner
  check_root
  echo -e "${BOLD}Updating NeoPanel to Latest Release...${NC} (Logs -> ${CYAN}${LOG_FILE}${NC})\n"

  run_step 0 35 "Fetching latest updates from GitHub" "step_update_fetch"
  run_step 35 75 "Compiling fresh NeoPanel binary" "step_update_compile"
  run_step 75 100 "Restarting service and verifying health" "step_update_swap"

  echo -e "\n${GREEN}${BOLD}================================================================${NC}"
  echo -e "${GREEN}${BOLD}           NeoPanel Successfully Updated to Latest Version!     ${NC}"
  echo -e "${GREEN}${BOLD}================================================================${NC}\n"
  echo -e "  ${BOLD}Service Status:${NC}   ${GREEN}Active & Running (systemd)${NC}"
  echo -e "  ${BOLD}Dashboard:${NC}        All database records, servers & configs preserved.\n"
  echo -e "${GREEN}================================================================${NC}\n"
}

# -----------------------------------------------------------------------------
# Interactive Menu
# -----------------------------------------------------------------------------

interactive_menu() {
  print_banner
  echo "Please select an option:"
  echo -e "  ${GREEN}[1]${NC} Full Installation (Panel + Java + SSL + Firewall)"
  echo -e "  ${GREEN}[2]${NC} Update NeoPanel to Latest Version (From GitHub)"
  echo -e "  ${GREEN}[3]${NC} Install Java Runtimes (OpenJDK 8, 17, 21 & Java 25)"
  echo -e "  ${GREEN}[4]${NC} Configure / Reconfigure Let's Encrypt SSL (Caddy)"
  echo -e "  ${GREEN}[5]${NC} Configure UFW Firewall"
  echo -e "  ${GREEN}[6]${NC} Restart NeoPanel Service"
  echo -e "  ${GREEN}[7]${NC} Uninstall NeoPanel"
  echo -e "  ${RED}[0]${NC} Exit"
  echo ""
  prompt_user "Enter selection [1]: " choice
  choice=${choice:-1}

  case "$choice" in
    1)
      run_full_installation
      ;;
    2)
      update_neopanel
      ;;
    3)
      check_root
      print_banner
      run_step 0 100 "Installing Java 8, 17, 21 & Java 25 runtimes" "step_java"
      log_success "Java installed successfully."
      ;;
    4)
      check_root
      prompt_user "Enter Domain: " dom
      prompt_user "Enter Email: " em
      print_banner
      run_step 0 100 "Configuring Let's Encrypt SSL & Caddy" "step_ssl_caddy '$dom' '$em'"
      log_success "SSL configured."
      ;;
    5)
      check_root
      print_banner
      run_step 0 100 "Configuring UFW Firewall" "step_firewall"
      log_success "Firewall configured."
      ;;
    6)
      systemctl restart neopanel
      log_success "NeoPanel restarted."
      ;;
    7)
      check_root
      prompt_user "Are you sure you want to completely uninstall NeoPanel? (y/N): " uninst
      if [[ "$uninst" =~ ^[Yy]$ ]]; then
        systemctl stop neopanel || true
        systemctl disable neopanel || true
        rm -f "$SERVICE_FILE"
        systemctl daemon-reload
        rm -f "$INSTALL_DIR/$BINARY_NAME"
        rm -f "$INSTALL_DIR/neopanel-update"
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

# Check if arguments provided
if [ $# -eq 0 ]; then
  interactive_menu
else
  case "$1" in
    --install)
      run_full_installation
      ;;
    --update|-u)
      update_neopanel
      ;;
    --java)
      check_root
      run_step 0 100 "Installing Java 8, 17, 21 & Java 25 runtimes" "step_java"
      ;;
    --firewall)
      check_root
      run_step 0 100 "Configuring UFW Firewall" "step_firewall"
      ;;
    *)
      interactive_menu
      ;;
  esac
fi
