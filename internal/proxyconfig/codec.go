package proxyconfig

import (
	"bytes"
	"errors"
	"io"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

type object = map[string]any

func parse(kind string, raw []byte) (object, error) {
	var m object
	if kind == "velocity" {
		if toml.Unmarshal(raw, &m) != nil || m == nil {
			return nil, errConfig
		}
	} else {
		dec := yaml.NewDecoder(bytes.NewReader(raw))
		var n, extra yaml.Node
		if dec.Decode(&n) != nil || len(n.Content) != 1 || n.Content[0].Kind != yaml.MappingNode || !safeYAML(&n, 0) {
			return nil, errConfig
		}
		if dec.Decode(&extra) != io.EOF || n.Decode(&m) != nil || m == nil {
			return nil, errConfig
		}
	}
	return m, nil
}

// Refuse structures whose meaning cannot be round-tripped through a string-key
// map without ambiguity. In particular never expand aliases or merge keys.
func safeYAML(n *yaml.Node, depth int) bool {
	if depth > 64 || n.Kind == yaml.AliasNode || n.Anchor != "" {
		return false
	}
	switch n.Kind {
	case yaml.DocumentNode:
	case yaml.MappingNode:
		if n.Tag != "!!map" {
			return false
		}
		seen := map[string]bool{}
		for i := 0; i < len(n.Content); i += 2 {
			k := n.Content[i]
			if k.Kind != yaml.ScalarNode || k.Tag != "!!str" || k.Value == "<<" || seen[k.Value] {
				return false
			}
			seen[k.Value] = true
		}
	case yaml.SequenceNode:
		if n.Tag != "!!seq" {
			return false
		}
	case yaml.ScalarNode:
		switch n.Tag {
		case "!!str", "!!bool", "!!int", "!!float", "!!null", "!!timestamp":
		default:
			return false
		}
	default:
		return false
	}
	for _, c := range n.Content {
		if !safeYAML(c, depth+1) {
			return false
		}
	}
	return true
}
func encode(kind string, m object) ([]byte, error) {
	var b []byte
	var err error
	if kind == "velocity" {
		b, err = toml.Marshal(m)
	} else {
		b, err = yaml.Marshal(m)
	}
	if err != nil || len(b) > maxConfig {
		return nil, errConfig
	}
	return b, nil
}

type decoder struct{ err error }

func (d *decoder) str(m object, k, def string) string {
	v, ok := m[k]
	if !ok {
		return def
	}
	s, ok := v.(string)
	if !ok {
		d.err = errConfig
	}
	return s
}
func (d *decoder) boolean(m object, k string, def bool) bool {
	v, ok := m[k]
	if !ok {
		return def
	}
	b, ok := v.(bool)
	if !ok {
		d.err = errConfig
	}
	return b
}
func (d *decoder) integer(m object, k string, def int) int {
	v, ok := m[k]
	if !ok {
		return def
	}
	var i int64
	switch n := v.(type) {
	case int:
		i = int64(n)
	case int64:
		i = n
	default:
		d.err = errConfig
	}
	if i < -2147483648 || i > 2147483647 {
		d.err = errConfig
	}
	return int(i)
}
func (d *decoder) obj(m object, k string) object {
	v, ok := m[k]
	if !ok {
		return object{}
	}
	o, ok := v.(map[string]any)
	if !ok {
		d.err = errConfig
		return object{}
	}
	return o
}
func (d *decoder) strings(v any) []string {
	out := []string{}
	switch a := v.(type) {
	case []string:
		return append(out, a...)
	case []any:
		for _, v := range a {
			s, ok := v.(string)
			if !ok {
				d.err = errConfig
			}
			out = append(out, s)
		}
	default:
		d.err = errConfig
	}
	return out
}
func (d *decoder) list(m object, k string) []string {
	v, ok := m[k]
	if !ok {
		return []string{}
	}
	return d.strings(v)
}
func listener(m object, builder bool) (object, error) {
	a, ok := m["listeners"].([]any)
	if !ok || len(a) == 0 {
		return nil, errConfig
	}
	if builder && len(a) > 1 {
		return nil, errors.New("Multiple listeners: edit config.yml in Files")
	}
	// Validate every entry, but only edit the primary listener.
	for _, v := range a {
		if _, ok := v.(map[string]any); !ok {
			return nil, errConfig
		}
	}
	return a[0].(map[string]any), nil
}
func settings(kind string, m object) (Settings, error) {
	s := defaults(kind)
	d := &decoder{}
	servers := d.obj(m, "servers")
	var forced object
	if kind == "velocity" {
		s.MOTD = d.str(m, "motd", "<#09add3>A Velocity Server")
		s.MaxPlayers = d.integer(m, "show-max-players", 500)
		s.OnlineMode = d.boolean(m, "online-mode", true)
		s.ForceKeyAuthentication = d.boolean(m, "force-key-authentication", true)
		s.ForwardingMode = strings.ToUpper(d.str(m, "player-info-forwarding-mode", "NONE"))
		s.PingPassthrough = strings.ToUpper(d.str(m, "ping-passthrough", "DISABLED"))
		s.Try = d.list(servers, "try")
		forced = d.obj(m, "forced-hosts")
		if _, err := secretPath(m); err != nil {
			return s, err
		}
		for name, v := range servers {
			if name == "try" {
				continue
			}
			a, ok := v.(string)
			if !ok {
				d.err = errConfig
			}
			s.Backends = append(s.Backends, Backend{Name: name, Address: a})
		}
	} else {
		l, err := listener(m, true)
		if err != nil {
			return s, err
		}
		s.MOTD = d.str(l, "motd", "&1Another Bungee server")
		s.MaxPlayers = d.integer(l, "max_players", 1)
		s.OnlineMode = d.boolean(m, "online_mode", true)
		s.IPForward = d.boolean(m, "ip_forward", false)
		s.ForceDefaultServer = d.boolean(l, "force_default_server", false)
		if d.boolean(l, "ping_passthrough", false) {
			s.PingPassthrough = "ALL"
		}
		s.Try = d.list(l, "priorities")
		// Mirror upstream's legacy priority migration without mutating on Read.
		for _, key := range []string{"default_server", "fallback_server"} {
			if value, exists := l[key]; exists && value != nil {
				s.Try = append(s.Try, d.str(l, key, ""))
			}
		}
		forced = d.obj(l, "forced_hosts")
		for name := range servers {
			b := d.obj(servers, name)
			s.Backends = append(s.Backends, Backend{Name: name, Address: d.str(b, "address", ""), MOTD: d.str(b, "motd", "&1Just another BungeeCord - Forced Host"), Restricted: d.boolean(b, "restricted", false)})
		}
	}
	for host, v := range forced {
		var refs []string
		if kind == "velocity" {
			refs = d.strings(v)
		} else {
			value, ok := v.(string)
			if !ok {
				d.err = errConfig
			}
			refs = []string{value}
		}
		s.ForcedHosts = append(s.ForcedHosts, ForcedHost{Host: host, Servers: refs})
	}
	sort.Slice(s.Backends, func(i, j int) bool { return s.Backends[i].Name < s.Backends[j].Name })
	sort.Slice(s.ForcedHosts, func(i, j int) bool { return s.ForcedHosts[i].Host < s.ForcedHosts[j].Host })
	if d.err != nil {
		return s, d.err
	}
	// Validate known bind/query fields without changing the parsed document.
	if err := checkPorts(kind, m); err != nil {
		return s, err
	}
	if err := validate(kind, s, false); err != nil {
		return s, errConfig
	}
	return s, nil
}
func checkPorts(kind string, m object) error {
	d := &decoder{}
	if kind == "velocity" {
		if _, ok := address(d.str(m, "bind", "0.0.0.0:25565")); !ok {
			return errConfig
		}
		q := d.obj(m, "query")
		d.boolean(q, "enabled", false)
		if !validPort(d.integer(q, "port", 25565)) {
			return errConfig
		}
	} else {
		l, err := listener(m, false)
		if err != nil {
			return err
		}
		if _, ok := address(d.str(l, "host", "0.0.0.0:25577")); !ok {
			return errConfig
		}
		d.boolean(l, "query_enabled", false)
		if !validPort(d.integer(l, "query_port", 25577)) {
			return errConfig
		}
	}
	return d.err
}
func setPort(kind string, m object, port int) error {
	if !validPort(port) {
		return errors.New("Invalid assigned port")
	}
	if err := checkPorts(kind, m); err != nil {
		return err
	}
	d := &decoder{}
	if kind == "velocity" {
		host, _ := address(d.str(m, "bind", "0.0.0.0:25565"))
		m["bind"] = net.JoinHostPort(host, strconv.Itoa(port))
		// A missing query table remains missing; do not enable query.
		if _, ok := m["query"]; ok {
			q := d.obj(m, "query")
			q["port"] = port
		}
	} else {
		l, _ := listener(m, false)
		host, _ := address(d.str(l, "host", "0.0.0.0:25577"))
		l["host"] = net.JoinHostPort(host, strconv.Itoa(port))
		l["query_port"] = port
	}
	return d.err
}
func apply(kind string, m object, s Settings) {
	servers, forced := object{}, object{}
	if kind == "velocity" {
		m["motd"] = s.MOTD
		m["show-max-players"] = s.MaxPlayers
		m["online-mode"] = s.OnlineMode
		m["force-key-authentication"] = s.ForceKeyAuthentication
		m["player-info-forwarding-mode"] = s.ForwardingMode
		m["ping-passthrough"] = s.PingPassthrough
		for _, b := range s.Backends {
			servers[b.Name] = b.Address
		}
		servers["try"] = append([]string{}, s.Try...)
		for _, h := range s.ForcedHosts {
			forced[h.Host] = h.Servers
		}
		m["forced-hosts"] = forced
	} else {
		l, _ := listener(m, true)
		l["motd"] = s.MOTD
		l["max_players"] = s.MaxPlayers
		l["force_default_server"] = s.ForceDefaultServer
		l["ping_passthrough"] = s.PingPassthrough == "ALL"
		l["priorities"] = append([]string{}, s.Try...)
		// These deprecated keys otherwise append to priorities in upstream Bungee.
		delete(l, "default_server")
		delete(l, "fallback_server")
		m["online_mode"] = s.OnlineMode
		m["ip_forward"] = s.IPForward
		old, _ := m["servers"].(map[string]any)
		for _, b := range s.Backends {
			retained, _ := old[b.Name].(map[string]any)
			if retained == nil {
				retained = object{}
			}
			retained["address"] = b.Address
			retained["motd"] = b.MOTD
			retained["restricted"] = b.Restricted
			servers[b.Name] = retained
		}
		for _, h := range s.ForcedHosts {
			forced[h.Host] = h.Servers[0]
		}
		l["forced_hosts"] = forced
	}
	m["servers"] = servers
}
func starter(kind string, port int) object {
	m := object{}
	if kind == "velocity" {
		m["config-version"] = "2.7"
		m["bind"] = net.JoinHostPort("0.0.0.0", strconv.Itoa(port))
		m["forwarding-secret-file"] = "forwarding.secret"
		m["prevent-client-proxy-connections"] = false
		m["announce-forge"] = false
		m["kick-existing-players"] = false
		m["enable-player-address-logging"] = true
		m["advanced"] = object{"compression-threshold": 256, "compression-level": -1, "login-ratelimit": 3000, "connection-timeout": 5000, "read-timeout": 30000, "haproxy-protocol": false, "tcp-fast-open": false, "bungee-plugin-message-channel": true, "show-ping-requests": false, "failover-on-unexpected-server-disconnect": true, "announce-proxy-commands": true, "log-command-executions": false, "log-player-connections": true, "accepts-transfers": false, "enable-reuse-port": false, "command-rate-limit": 50, "forward-commands-if-rate-limited": true, "kick-after-rate-limited-commands": 0, "tab-complete-rate-limit": 10, "kick-after-rate-limited-tab-completes": 0}
		m["query"] = object{"enabled": false, "port": port, "map": "Velocity", "show-plugins": false}
	} else {
		m["player_limit"] = -1
		m["listeners"] = []any{object{"host": net.JoinHostPort("0.0.0.0", strconv.Itoa(port)), "query_enabled": false, "query_port": port, "tab_list": "GLOBAL_PING", "tab_size": 60, "bind_local_address": true}}
	}
	apply(kind, m, defaults(kind))
	return m
}
