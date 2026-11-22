package sftp

import (
	"github.com/Pekly/NeoPanel/internal/database"
	"testing"
)

func TestFeatureDisableRevokesConnectedSFTP(t *testing.T) {
	f := newFixture(t)
	_, client := f.connect(t, "owner.srv-alpha", "password")
	file, err := client.Open("file.txt")
	must(t, err)
	defer file.Close()
	must(t, f.db.SetFeaturePolicy("srv-alpha", database.FeaturePolicy{"sftp": false}))
	_, err = client.Stat("file.txt")
	denied(t, err)
	_, err = file.Read(make([]byte, 8))
	denied(t, err)
	denied(t, f.server.authorize(f.users["admin"].ID, "srv-alpha", "files.view"))
	must(t, f.db.SetFeaturePolicy("srv-alpha", database.FeaturePolicy{}))
	_, err = client.Stat("file.txt")
	must(t, err)
}
