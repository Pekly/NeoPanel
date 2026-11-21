package api

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
	"golang.org/x/crypto/ssh"
)

func TestSSHKeyRegistrationGuidance(t *testing.T) {
	f := newReliabilityFixture(t)
	for _, value := range []string{strings.Repeat("ab", 32), "SHA256:fingerprint-only", "-----BEGIN OPENSSH PRIVATE KEY-----\nnot-a-public-key", "ssh-ed25519 AAAA-not-a-key"} {
		body, _ := json.Marshal(map[string]string{"name": "Laptop", "public_key": value})
		w := f.request("POST", "/api/account/ssh-keys", string(body), "owner")
		if w.Code != 400 || !strings.Contains(w.Body.String(), ".pub file") {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	keys, err := f.db.ListSSHKeys(f.users["owner"].ID)
	if err != nil || len(keys) != 0 {
		t.Fatal("invalid keys were stored", keys, err)
	}
	public, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	key, err := ssh.NewPublicKey(public)
	if err != nil {
		t.Fatal(err)
	}
	authorized := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key))) + " laptop@example"
	body, _ := json.Marshal(map[string]string{"name": "Laptop", "public_key": authorized})
	w := f.request("POST", "/api/account/ssh-keys", string(body), "owner")
	var saved database.SSHKey
	if err := json.Unmarshal(w.Body.Bytes(), &saved); err != nil || w.Code != 200 {
		t.Fatal(w.Code, w.Body.String(), err)
	}
	if saved.Fingerprint != ssh.FingerprintSHA256(key) || saved.PublicKey != authorized || saved.UserID != f.users["owner"].ID {
		t.Fatal(saved)
	}
	keys, err = f.db.GetAllSSHKeysForUser(f.users["owner"].Username)
	if err != nil || len(keys) != 1 || keys[0].PublicKey != authorized {
		t.Fatal(keys, err)
	}
	other, err := f.db.ListSSHKeys(f.users["admin"].ID)
	if err != nil || len(other) != 0 {
		t.Fatal("key leaked to another account", other, err)
	}
}
