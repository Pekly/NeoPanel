package database

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func portTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := InitDB(filepath.Join(t.TempDir(), "ports.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
func portTestServer(t *testing.T, db *DB, id, node string, port, allowed int) *Server {
	t.Helper()
	s := &Server{ID: id, Name: id, NodeID: node, Port: port, PortsAllowed: allowed}
	if err := db.CreateServer(s); err != nil {
		t.Fatal(err)
	}
	return s
}
func TestPortDefaultPrimaryQuotaAndNodeScope(t *testing.T) {
	db := portTestDB(t)
	s := portTestServer(t, db, "a", "", 25565, 0)
	if s.PortsAllowed != 1 || s.PortsUsed != 1 {
		t.Fatal(s)
	}
	if _, err := db.AddServerPort("a", 25566, "tcp", "", false); err == nil {
		t.Fatal("default quota bypassed")
	}
	if err := db.SetServerPortAllowance("a", 3); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AddServerPort("a", 25566, "tcp", "", false); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateServer(&Server{ID: "b", Name: "b", Port: 25566}); err == nil {
		t.Fatal("creation collision accepted")
	}
	if _, err := db.GetServer("b"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("partial creation", err)
	}
	portTestServer(t, db, "remote", "remote", 25566, 1)
	if err := db.SetServerPortAllowance("a", 1); err == nil {
		t.Fatal("quota below usage accepted")
	}
	got, err := db.GetServer("a")
	if err != nil || got.PortsUsed != 2 || got.PortsAllowed != 3 {
		t.Fatal(got, err)
	}
	ports, _ := db.ListServerPorts("a")
	if err := db.DeleteServerPort("a", ports[0].ID); err == nil {
		t.Fatal("deleted primary")
	}
}
func TestPortRandomPoolAndExhaustion(t *testing.T) {
	db := portTestDB(t)
	portTestServer(t, db, "a", "local", 25565, 5)
	if _, err := db.AllocateServerPort("a", 40000, "tcp", "", false, true, nil); err == nil {
		t.Fatal("unconfigured pool accepted")
	}
	if err := db.SetPortPool(PortPool{30000, 30001}); err != nil {
		t.Fatal(err)
	}
	portTestServer(t, db, "b", "local", 30000, 1)
	p, err := db.AllocateServerPort("a", 45000, "udp", "", false, true, nil)
	if err != nil || p.Port != 30001 {
		t.Fatal(p, err)
	}
	if _, err = db.AllocateServerPort("a", 0, "tcp", "", false, true, nil); err == nil {
		t.Fatal("exhausted pool accepted")
	}
	if err = db.SetPortPool(PortPool{1, 30000}); err == nil {
		t.Fatal("invalid pool accepted")
	}
}
func TestPortConcurrentQuotaAndCollision(t *testing.T) {
	for _, kind := range []string{"quota", "collision"} {
		t.Run(kind, func(t *testing.T) {
			db := portTestDB(t)
			const n = 16
			for i := 0; i < n; i++ {
				portTestServer(t, db, fmt.Sprintf("s%d", i), "local", 25000+i, 2)
			}
			// A second handle exercises SQLite writer serialization, not a Go mutex.
			var file string
			if err := db.conn.QueryRow(`SELECT file FROM pragma_database_list WHERE name='main'`).Scan(&file); err != nil {
				t.Fatal(err)
			}
			other, err := InitDB(file)
			if err != nil {
				t.Fatal(err)
			}
			defer other.Close()
			var successes atomic.Int32
			var wg sync.WaitGroup
			for i := 0; i < n; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					id := fmt.Sprintf("s%d", i)
					p := 31000
					use := db
					if i%2 == 1 {
						use = other
					}
					if kind == "quota" {
						id = "s0"
						p += i
					}
					if _, err := use.AddServerPort(id, p, "tcp", "", false); err == nil {
						successes.Add(1)
					}
				}(i)
			}
			wg.Wait()
			if successes.Load() != 1 {
				t.Fatal("successful allocations", successes.Load())
			}
		})
	}
}
func TestPortConfigFailureRollsBackTransaction(t *testing.T) {
	db := portTestDB(t)
	portTestServer(t, db, "a", "local", 25565, 3)
	injected := errors.New("config failed")
	if _, err := db.AllocateServerPort("a", 25566, "tcp", "", true, false, func(int) error { return injected }); !errors.Is(err, injected) {
		t.Fatal(err)
	}
	s, _ := db.GetServer("a")
	if s.Port != 25565 || s.PortsUsed != 1 {
		t.Fatal(s)
	}
	p, err := db.AddServerPort("a", 25566, "tcp", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.ChangePrimaryServerPort("a", p.ID, func(int) error { return injected }); !errors.Is(err, injected) {
		t.Fatal(err)
	}
	s, _ = db.GetServer("a")
	if s.Port != 25565 || s.PortsUsed != 2 {
		t.Fatal(s)
	}
	ports, _ := db.ListServerPorts("a")
	if !ports[0].IsPrimary || ports[0].Port != 25565 {
		t.Fatal(ports)
	}
}
func TestPortMigrationPreservesLegacyAssignments(t *testing.T) {
	conn, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "legacy.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	db := &DB{conn: conn}
	if err = db.migrate(); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	for _, id := range []string{"a", "b"} {
		if _, err = conn.Exec(`INSERT INTO servers(id,name,port,created_at,updated_at) VALUES(?,?,25565,?,?)`, id, id, now, now); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = conn.Exec(`INSERT INTO server_ports(server_id,port,protocol,notes,is_primary,created_at) VALUES('a',25566,'udp','keep',0,?)`, now); err != nil {
		t.Fatal(err)
	}
	if err = db.migratePorts(); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"a", "b"} {
		s, err := db.GetServer(id)
		want := 1
		if id == "a" {
			want = 2
		}
		if err != nil || s.Port != 25565 || s.PortsAllowed != want || s.PortsUsed != want {
			t.Fatal(s, err)
		}
	}
	ports, _ := db.ListServerPorts("a")
	if ports[1].Notes != "keep" || ports[1].Protocol != "udp" {
		t.Fatal(ports)
	}
	if err = db.SetServerPortAllowance("a", 4); err != nil {
		t.Fatal(err)
	}
	if err = db.migratePorts(); err != nil {
		t.Fatal(err)
	}
	s, _ := db.GetServer("a")
	if s.PortsAllowed != 4 {
		t.Fatal("migration reset allowance")
	}
}
