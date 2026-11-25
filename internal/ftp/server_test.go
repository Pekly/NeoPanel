package ftp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func TestFTPServerWorkflow(t *testing.T) {
	dir := t.TempDir()
	db, err := database.InitDB(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte("secretpass123"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	owner, err := db.CreateUser("ftpuser", string(hash), "user")
	if err != nil {
		t.Fatal(err)
	}
	srv := &database.Server{
		ID:      "srv1",
		Name:    "TestServer",
		OwnerID: owner.ID,
		NodeID:  "local",
	}
	if err := db.CreateServer(srv); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Paths: config.PathsConfig{ServersDir: filepath.Join(dir, "servers")},
		FTP:   config.FTPConfig{Enabled: true, Port: 0},
	}

	fsrv, err := NewFTPServer(cfg, db)
	if err != nil {
		t.Fatal(err)
	}
	if err := fsrv.Start(); err != nil {
		t.Fatal(err)
	}
	defer fsrv.Stop()

	conn, err := net.Dial("tcp", fsrv.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	rd := bufio.NewReader(conn)
	readResp := func() string {
		line, err := rd.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		return line
	}
	sendCmd := func(c string) string {
		_, _ = fmt.Fprintf(conn, "%s\r\n", c)
		return readResp()
	}

	if got := readResp(); !strings.HasPrefix(got, "220 ") {
		t.Fatalf("unexpected greeting: %s", got)
	}
	if got := sendCmd("USER ftpuser#srv1"); !strings.HasPrefix(got, "331 ") {
		t.Fatalf("unexpected USER resp: %s", got)
	}
	if got := sendCmd("PASS wrongpass"); !strings.HasPrefix(got, "530 ") {
		t.Fatalf("expected 530 on bad pass, got: %s", got)
	}
	if got := sendCmd("PASS secretpass123"); !strings.HasPrefix(got, "230 ") {
		t.Fatalf("expected 230 login, got: %s", got)
	}
	if got := sendCmd("MKD plugins"); !strings.HasPrefix(got, "257 ") {
		t.Fatalf("expected 257 MKD, got: %s", got)
	}
	if got := sendCmd("CWD ../../../etc"); !strings.HasPrefix(got, "550 ") {
		t.Fatalf("expected 550 traversal rejection, got: %s", got)
	}
	if got := sendCmd("CWD plugins"); !strings.HasPrefix(got, "250 ") {
		t.Fatalf("expected 250 CWD, got: %s", got)
	}

	// Open EPSV data connection and STOR
	epsv := sendCmd("EPSV")
	if !strings.HasPrefix(epsv, "229 ") {
		t.Fatalf("expected 229 EPSV, got: %s", epsv)
	}
	s := strings.Index(epsv, "(|||")
	e := strings.Index(epsv[s+4:], "|)")
	dataPort := epsv[s+4 : s+4+e]
	host, _, _ := net.SplitHostPort(fsrv.Addr())
	dc, err := net.Dial("tcp", net.JoinHostPort(host, dataPort))
	if err != nil {
		t.Fatal(err)
	}
	if got := sendCmd("STOR hello.txt"); !strings.HasPrefix(got, "150 ") {
		t.Fatalf("expected 150 STOR, got: %s", got)
	}
	_, _ = io.WriteString(dc, "hello from ftp")
	_ = dc.Close()
	if got := readResp(); !strings.HasPrefix(got, "226 ") {
		t.Fatalf("expected 226 transfer complete, got: %s", got)
	}

	// Retrieve via RETR
	epsv2 := sendCmd("EPSV")
	s2 := strings.Index(epsv2, "(|||")
	e2 := strings.Index(epsv2[s2+4:], "|)")
	dc2, err := net.Dial("tcp", net.JoinHostPort(host, epsv2[s2+4:s2+4+e2]))
	if err != nil {
		t.Fatal(err)
	}
	if got := sendCmd("RETR hello.txt"); !strings.HasPrefix(got, "150 ") {
		t.Fatalf("expected 150 RETR, got: %s", got)
	}
	gotBytes, _ := io.ReadAll(dc2)
	_ = dc2.Close()
	if got := readResp(); !strings.HasPrefix(got, "226 ") {
		t.Fatalf("expected 226 RETR complete, got: %s", got)
	}
	if string(gotBytes) != "hello from ftp" {
		t.Fatalf("unexpected RETR bytes: %q", string(gotBytes))
	}
	_ = sendCmd("QUIT")
}
