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
BOLD='\033[1m'
NC='\033[0m' # No Color

# Default values
PANEL_PORT="8080"
CONFIG_DIR="/etc/neopanel"
DATA_DIR="/var/lib/neopanel"
SERVICE_FILE="/etc/systemd/system/neopanel.service"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="neopanel"

# -----------------------------------------------------------------------------
# Helper Functions
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

check_root() {
  if [ "$EUID" -ne 0 ]; then
    log_error "This script must be run as root! Run with: sudo bash $0"
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
    read -p "Do you want to continue anyway? (y/N): " choice
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
# Dependency Management
# -----------------------------------------------------------------------------

install_prerequisites() {
  log_info "Updating system packages..."
  apt-get update -y -q

  log_info "Installing essential utilities (curl, wget, jq, tar, ufw, ca-certificates)..."
  apt-get install -y -q curl wget jq tar ufw ca-certificates software-properties-common git

  log_success "Base dependencies installed."
}

install_java() {
  log_info "Installing Java OpenJDK Runtimes for Minecraft (Java 21 & 17)..."
  apt-get update -y -q
  apt-get install -y -q openjdk-21-jre-headless openjdk-17-jre-headless

  if command -v java &>/dev/null; then
    JAVA_VER=$(java -version 2>&1 | head -n 1)
    log_success "Java installed successfully: $JAVA_VER"
  else
    log_error "Java installation failed."
  fi
}

# -----------------------------------------------------------------------------
# Binary & Service Setup
# -----------------------------------------------------------------------------

install_neopanel_binary() {
  log_info "Setting up NeoPanel core binary..."

  # Create directories
  mkdir -p "$CONFIG_DIR"
  mkdir -p "$DATA_DIR/servers"
  mkdir -p "$DATA_DIR/backups"

  # If running from local source tree, compile or copy
  if [ -f "./cmd/neopanel/main.go" ]; then
    log_info "Compiling NeoPanel from local source..."
    if ! command -v go &>/dev/null; then
      log_info "Installing Go compiler..."
      apt-get install -y -q golang-go
    fi
    go mod tidy
    go build -ldflags="-s -w" -o "$INSTALL_DIR/$BINARY_NAME" ./cmd/neopanel
  else
    log_info "Downloading NeoPanel source and compiling..."
    if ! command -v go &>/dev/null; then
      log_info "Installing Go compiler..."
      apt-get install -y -q golang-go
    fi
    TMP_DIR=$(mktemp -d)
    log_info "Cloning repository from https://github.com/Pekly/NeoPanel.git..."
    git clone https://github.com/Pekly/NeoPanel.git "$TMP_DIR"
    if [ -f "$TMP_DIR/cmd/neopanel/main.go" ]; then
      log_info "Resolving dependencies and compiling binary..."
      (cd "$TMP_DIR" && go mod tidy && go build -ldflags="-s -w" -o "$INSTALL_DIR/$BINARY_NAME" ./cmd/neopanel)
      rm -rf "$TMP_DIR"
    else
      log_error "Failed to clone repository or main.go not found."
      exit 1
    fi
  fi

  chmod +x "$INSTALL_DIR/$BINARY_NAME"
  log_success "Binary installed at $INSTALL_DIR/$BINARY_NAME"

  # Install global neopanel-update command shortcut
  cat > "$INSTALL_DIR/neopanel-update" <<'EOF'
#!/usr/bin/env bash
set -e
bash <(curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh) --update
EOF
  chmod +x "$INSTALL_DIR/neopanel-update"
  log_success "Global updater command created: 'neopanel-update'"
}

create_config() {
  local host="$1"
  local port="$2"

  local secret
  secret=$(generate_secret)

  log_info "Generating configuration at $CONFIG_DIR/config.yaml..."

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
  log_success "Configuration generated with secure random JWT secret."
}

create_systemd_service() {
  log_info "Creating systemd service at $SERVICE_FILE..."

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
  log_success "NeoPanel systemd service created and started."
}

# -----------------------------------------------------------------------------
# SSL & Reverse Proxy (Caddy)
# -----------------------------------------------------------------------------

setup_ssl_caddy() {
  local domain="$1"
  local email="$2"

  log_info "Installing Caddy web server for automated Let's Encrypt SSL..."

  apt-get install -y -q debian-keyring debian-archive-keyring apt-transport-https curl
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor --yes -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg 2>/dev/null
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list >/dev/null
  apt-get update -y -q
  apt-get install -y -q caddy

  log_info "Configuring automated SSL reverse proxy for domain: $domain..."

  cat > /etc/caddy/Caddyfile <<EOF
${domain} {
    encode gzip zstd
    tls ${email}

    # Proxy WebSocket & API traffic to NeoPanel
    reverse_proxy 127.0.0.1:${PANEL_PORT} {
        header_up Host {host}
        header_up X-Real-IP {remote_host}
    }
}
EOF

  systemctl restart caddy
  log_success "Caddy configured! Let's Encrypt certificate will be provisioned automatically."
}

# -----------------------------------------------------------------------------
# Firewall Configuration
# -----------------------------------------------------------------------------

configure_firewall() {
  log_info "Configuring UFW firewall..."

  # Keep SSH open first!
  ufw allow 22/tcp >/dev/null 2>&1 || true

  # Allow Web ports
  ufw allow 80/tcp >/dev/null 2>&1
  ufw allow 443/tcp >/dev/null 2>&1

  # Allow Panel standalone port if no SSL
  ufw allow "${PANEL_PORT}/tcp" >/dev/null 2>&1

  # Allow common Minecraft server port range (25565-25600 TCP/UDP)
  ufw allow 25565:25600/tcp >/dev/null 2>&1
  ufw allow 25565:25600/udp >/dev/null 2>&1

  # Enable UFW non-interactively
  echo "y" | ufw enable >/dev/null 2>&1

  log_success "Firewall rules applied: SSH (22), HTTP (80), HTTPS (443), Minecraft (25565-25600)."
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
  read -p "Use Domain & SSL? (y/n) [default: y]: " use_ssl
  use_ssl=${use_ssl:-y}

  local domain=""
  local email=""
  local bind_host="0.0.0.0"

  if [[ "$use_ssl" =~ ^[Yy]$ ]]; then
    read -p "Enter your Domain name (e.g. panel.example.com): " domain
    while [ -z "$domain" ]; do
      read -p "Domain cannot be empty. Enter Domain: " domain
    done

    read -p "Enter Email for Let's Encrypt certificates: " email
    while [ -z "$email" ]; do
      read -p "Email cannot be empty. Enter Email: " email
    done

    # When behind reverse proxy, bind internally
    bind_host="127.0.0.1"
  fi

  echo -e "\n${BOLD}Starting installation process...${NC}\n"

  install_prerequisites
  install_java
  install_neopanel_binary
  create_config "$bind_host" "$PANEL_PORT"
  create_systemd_service

  if [[ "$use_ssl" =~ ^[Yy]$ ]]; then
    setup_ssl_caddy "$domain" "$email"
  fi

  read -p "Would you like to configure UFW firewall rules automatically? (y/n) [default: y]: " conf_ufw
  conf_ufw=${conf_ufw:-y}
  if [[ "$conf_ufw" =~ ^[Yy]$ ]]; then
    configure_firewall
  fi

  # -----------------------------------------------------------------------------
  # Completion Banner
  # -----------------------------------------------------------------------------
  print_banner
  echo -e "${GREEN}${BOLD}================================================================${NC}"
  echo -e "${GREEN}${BOLD}           NeoPanel Installation Successfully Completed!         ${NC}"
  echo -e "${GREEN}${BOLD}================================================================${NC}\n"

  if [[ "$use_ssl" =~ ^[Yy]$ ]]; then
    echo -e "  ${BOLD}Web Dashboard:${NC}   ${CYAN}https://${domain}${NC}"
  else
    echo -e "  ${BOLD}Web Dashboard:${NC}   ${CYAN}http://${pub_ip}:${PANEL_PORT}${NC}"
  fi

  echo -e "  ${BOLD}Initial Setup:${NC}   Open the URL in your browser to create the Admin user."
  echo -e "  ${BOLD}Config Path:${NC}     ${CONFIG_DIR}/config.yaml"
  echo -e "  ${BOLD}Data Path:${NC}       ${DATA_DIR}"
  echo -e "\n  ${BOLD}Useful Commands:${NC}"
  echo -e "    ${YELLOW}neopanel-update${NC}            # Instant one-command update from GitHub"
  echo -e "    ${YELLOW}systemctl status neopanel${NC}    # View status"
  echo -e "    ${YELLOW}journalctl -u neopanel -f${NC}    # View live logs"
  echo -e "    ${YELLOW}systemctl restart neopanel${NC}   # Restart panel"
  echo -e "\n${GREEN}Thank you for choosing NeoPanel!${NC}\n"
}

# -----------------------------------------------------------------------------
# Update Flow
# -----------------------------------------------------------------------------

update_neopanel() {
  print_banner
  check_root
  log_info "Initiating seamless update from GitHub repository..."

  if ! command -v git &>/dev/null; then
    apt-get update -y -q && apt-get install -y -q git
  fi
  if ! command -v go &>/dev/null; then
    apt-get update -y -q && apt-get install -y -q golang-go
  fi

  TMP_DIR=$(mktemp -d)
  log_info "Cloning latest commits from https://github.com/Pekly/NeoPanel.git..."
  if ! git clone --depth 1 https://github.com/Pekly/NeoPanel.git "$TMP_DIR"; then
    log_error "Failed to clone repository from GitHub."
    rm -rf "$TMP_DIR"
    exit 1
  fi

  LATEST_COMMIT=$(cd "$TMP_DIR" && git log -1 --format="%h - %s (%cr)")
  log_info "Target version: ${LATEST_COMMIT}"

  log_info "Tidying module dependencies..."
  (cd "$TMP_DIR" && go mod tidy)

  log_info "Compiling fresh NeoPanel binary..."
  if ! (cd "$TMP_DIR" && go build -ldflags="-s -w" -o "$TMP_DIR/neopanel_new" ./cmd/neopanel); then
    log_error "Build failed! Current running version was not modified."
    rm -rf "$TMP_DIR"
    exit 1
  fi

  # Backup previous binary
  if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    cp "$INSTALL_DIR/$BINARY_NAME" "$INSTALL_DIR/${BINARY_NAME}.bak"
  fi

  # Atomically replace binary
  mv "$TMP_DIR/neopanel_new" "$INSTALL_DIR/$BINARY_NAME"
  chmod +x "$INSTALL_DIR/$BINARY_NAME"
  rm -rf "$TMP_DIR"

  log_info "Restarting NeoPanel service..."
  systemctl restart neopanel
  sleep 1

  if systemctl is-active --quiet neopanel; then
    echo -e "\n${GREEN}${BOLD}================================================================${NC}"
    echo -e "${GREEN}${BOLD}           NeoPanel Successfully Updated to Latest Version!     ${NC}"
    echo -e "${GREEN}${BOLD}================================================================${NC}\n"
    echo -e "  ${BOLD}Installed Commit:${NC} ${CYAN}${LATEST_COMMIT}${NC}"
    echo -e "  ${BOLD}Service Status:${NC}   ${GREEN}Active & Running (systemd)${NC}"
    echo -e "  ${BOLD}Dashboard:${NC}        All database records, servers & configs preserved.\n"
    echo -e "${GREEN}================================================================${NC}\n"
  else
    log_error "Service failed to restart cleanly! Rolling back binary..."
    if [ -f "$INSTALL_DIR/${BINARY_NAME}.bak" ]; then
      mv "$INSTALL_DIR/${BINARY_NAME}.bak" "$INSTALL_DIR/$BINARY_NAME"
      systemctl restart neopanel
    fi
    log_error "Check system logs with: journalctl -u neopanel -n 50"
    exit 1
  fi
}

# -----------------------------------------------------------------------------
# Interactive Menu
# -----------------------------------------------------------------------------

interactive_menu() {
  print_banner
  echo "Please select an option:"
  echo -e "  ${GREEN}[1]${NC} Full Installation (Panel + Java + SSL + Firewall)"
  echo -e "  ${GREEN}[2]${NC} Update NeoPanel to Latest Version (From GitHub)"
  echo -e "  ${GREEN}[3]${NC} Install Java Runtimes (OpenJDK 17 & 21)"
  echo -e "  ${GREEN}[4]${NC} Configure / Reconfigure Let's Encrypt SSL (Caddy)"
  echo -e "  ${GREEN}[5]${NC} Configure UFW Firewall"
  echo -e "  ${GREEN}[6]${NC} Restart NeoPanel Service"
  echo -e "  ${GREEN}[7]${NC} Uninstall NeoPanel"
  echo -e "  ${RED}[0]${NC} Exit"
  echo ""
  read -p "Enter selection [1]: " choice
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
      install_java
      ;;
    4)
      check_root
      read -p "Enter Domain: " dom
      read -p "Enter Email: " em
      setup_ssl_caddy "$dom" "$em"
      ;;
    5)
      check_root
      configure_firewall
      ;;
    6)
      systemctl restart neopanel
      log_success "NeoPanel restarted."
      ;;
    7)
      check_root
      read -p "Are you sure you want to completely uninstall NeoPanel? (y/N): " uninst
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
      install_java
      ;;
    --firewall)
      check_root
      configure_firewall
      ;;
    *)
      interactive_menu
      ;;
  esac
fi

