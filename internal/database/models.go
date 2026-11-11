package database

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"` // "admin" or "user"
	CreatedAt    time.Time `json:"created_at"`
}

type Server struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`       // "paper", "purpur", "vanilla", "fabric", "spigot", "custom"
	Version     string    `json:"version"`    // e.g. "1.20.4", "1.21.1"
	JarFile     string    `json:"jar_file"`   // e.g. "server.jar"
	Port        int       `json:"port"`       // e.g. 25565
	MemoryMinMB int       `json:"memory_min"` // e.g. 1024
	MemoryMaxMB int       `json:"memory_max"` // e.g. 2048
	JavaArgs    string    `json:"java_args"`  // custom JVM flags
	JavaPath    string    `json:"java_path"`  // custom Java binary path (optional)
	AutoRestart bool      `json:"auto_restart"`
	IconURL     string    `json:"icon_url,omitempty"`
	OwnerID     int       `json:"owner_id,omitempty"`
	NodeID      string    `json:"node_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ServerStats struct {
	ServerID        string  `json:"server_id"`
	Status          string  `json:"status"` // "offline", "starting", "online", "stopping"
	CPUUsage        float64 `json:"cpu_usage"`
	MemoryMB        int64   `json:"memory_mb"`
	MemoryMaxMB     int     `json:"memory_max_mb"`
	DiskMB          int64   `json:"disk_mb"`
	UptimeSec       int64   `json:"uptime_sec"`
	NetInKiB        float64 `json:"net_in_kib"`
	NetOutKiB       float64 `json:"net_out_kib"`
	StartupFailures int     `json:"startup_failures"`
	StartupHalted   bool    `json:"startup_halted"`
}

type SSHKey struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Name        string    `json:"name"`
	Fingerprint string    `json:"fingerprint"`
	PublicKey   string    `json:"public_key"`
	CreatedAt   time.Time `json:"created_at"`
}

type APIKey struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	KeyPrefix string    `json:"key_prefix"`
	CreatedAt time.Time `json:"created_at"`
}

type ActivityLog struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
	IPAddress string    `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
}

type ServerPort struct {
	ID        int       `json:"id"`
	ServerID  string    `json:"server_id"`
	Port      int       `json:"port"`
	Protocol  string    `json:"protocol"` // "tcp/udp", "tcp", "udp"
	Notes     string    `json:"notes"`    // "Primary Game Port", "Voice Chat", "Dynmap", "Geyser"
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
}

type ClusterNode struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	BaseURL         string    `json:"base_url"`
	APIToken        string    `json:"-"`
	FQDN            string    `json:"fqdn"`
	Location        string    `json:"location"`
	IsFree          bool      `json:"is_free"`
	IdleTimeoutMin  int       `json:"idle_timeout_min"` // minutes of 0 players before auto-stop (e.g. 15)
	AllowUserRAM    bool      `json:"allow_user_ram"`   // allow non-admin users to change RAM
	AllowUserPorts  bool      `json:"allow_user_ports"` // allow non-admin users to change game port
	AllowUserJava   bool      `json:"allow_user_java"`  // allow non-admin users to change java version
	IsolationEngine string    `json:"isolation_engine"` // "auto", "bwrap", "systemd", "process"
	CPULimitPercent int       `json:"cpu_limit_percent"`
	TasksMax        int       `json:"tasks_max"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ServerSubuser struct {
	ID          int       `json:"id"`
	ServerID    string    `json:"server_id"`
	UserID      int       `json:"user_id"`
	Username    string    `json:"username"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
}

type ServerSchedule struct {
	ID        string     `json:"id"`
	ServerID  string     `json:"server_id"`
	Name      string     `json:"name"`
	CronExpr  string     `json:"cron_expr"`
	Action    string     `json:"action"` // "command", "restart", "backup", "start", "stop"
	Payload   string     `json:"payload"`
	IsActive  bool       `json:"is_active"`
	LastRunAt *time.Time `json:"last_run_at,omitempty"`
	NextRunAt *time.Time `json:"next_run_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type ServerDatabase struct {
	ID             string    `json:"id"`
	ServerID       string    `json:"server_id"`
	Name           string    `json:"name"`
	Engine         string    `json:"engine"` // "sqlite" or "mariadb"
	Username       string    `json:"username"`
	Password       string    `json:"password"`
	Host           string    `json:"host"`
	Port           int       `json:"port"`
	FilePath       string    `json:"file_path"`
	MaxConnections int       `json:"max_connections"`
	CreatedAt      time.Time `json:"created_at"`
}
