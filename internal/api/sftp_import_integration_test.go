package api

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/sftpimport"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func localImportSource(t *testing.T) (sftpimport.Options, string) {
	t.Helper()
	_, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	key, err := ssh.NewSignerFromKey(private)
	if err != nil {
		t.Fatal(err)
	}
	config := &ssh.ServerConfig{PasswordCallback: func(meta ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
		if meta.User() != "source" || string(password) != "temporary-password" {
			return nil, fmt.Errorf("denied")
		}
		return nil, nil
	}}
	config.AddHostKey(key)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	handlers := sftp.InMemHandler()
	var mu sync.Mutex
	var connections []net.Conn
	var clients sync.WaitGroup
	accepted := make(chan struct{})
	go func() {
		defer close(accepted)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			mu.Lock()
			connections = append(connections, conn)
			mu.Unlock()
			clients.Add(1)
			go func() {
				defer clients.Done()
				defer conn.Close()
				sshConn, channels, requests, err := ssh.NewServerConn(conn, config)
				if err != nil {
					return
				}
				defer sshConn.Close()
				go ssh.DiscardRequests(requests)
				for pending := range channels {
					if pending.ChannelType() != "session" {
						_ = pending.Reject(ssh.UnknownChannelType, "session required")
						continue
					}
					channel, requests, err := pending.Accept()
					if err != nil {
						return
					}
					for request := range requests {
						var payload struct{ Name string }
						ok := request.Type == "subsystem" && ssh.Unmarshal(request.Payload, &payload) == nil && payload.Name == "sftp"
						_ = request.Reply(ok, nil)
						if ok {
							server := sftp.NewRequestServer(channel, handlers)
							_ = server.Serve()
							_ = server.Close()
							break
						}
					}
					_ = channel.Close()
				}
			}()
		}
	}()
	t.Cleanup(func() {
		_ = listener.Close()
		<-accepted
		mu.Lock()
		for _, conn := range connections {
			_ = conn.Close()
		}
		mu.Unlock()
		clients.Wait()
	})
	address := listener.Addr().String()
	sshClient, err := ssh.Dial("tcp", address, &ssh.ClientConfig{User: "source", Auth: []ssh.AuthMethod{ssh.Password("temporary-password")}, HostKeyCallback: ssh.FixedHostKey(key.PublicKey()), Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer sshClient.Close()
	client, err := sftp.NewClient(sshClient)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if err := client.Mkdir("/world"); err != nil {
		t.Fatal(err)
	}
	file, err := client.Create("/world/level.dat")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte("test-world")); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	host, portString, _ := net.SplitHostPort(address)
	port, _ := strconv.Atoi(portString)
	return sftpimport.Options{Host: host, Port: port, Username: "source", Password: "temporary-password", Fingerprint: ssh.FingerprintSHA256(key.PublicKey()), RemotePath: "/", PreserveTimes: true, MaxBytes: 1024, TimeoutMinutes: 1}, address
}

func TestSFTPImportHTTPToRealSFTP(t *testing.T) {
	opts, address := localImportSource(t)
	f := newReliabilityFixture(t)
	f.cfg.Security.SFTPImportAllowedHosts = []string{address}
	body, _ := json.Marshal(map[string]any{"host": opts.Host, "port": opts.Port})
	w := f.request("POST", importURL+"/host-key", string(body), "owner")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var key sftpimport.HostKeyInfo
	if err := json.Unmarshal(w.Body.Bytes(), &key); err != nil || key.Fingerprint != opts.Fingerprint {
		t.Fatal(key, err)
	}
	body, _ = json.Marshal(opts)
	w = f.request("POST", importURL+"/test", string(body), "owner")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	w = f.request("POST", importURL, string(body), "owner")
	if w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	job := waitImportJob(t, f, "completed")
	if job.BytesDone != 10 || job.FilesDone != 1 {
		t.Fatal(job)
	}
	data, err := os.ReadFile(filepath.Join(f.manager.GetWorkingDir("alpha"), "world", "level.dat"))
	if err != nil || string(data) != "test-world" {
		t.Fatal(string(data), err)
	}
	if f.manager.GetStatus("alpha") != "offline" {
		t.Fatal("import started game")
	}
}
