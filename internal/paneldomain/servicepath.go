package paneldomain

import (
	"errors"
	"os"
)

func canonicalServiceUnit(path string) (string, error) {
	const canonical = "/usr/lib/systemd/system/caddy.service"
	if path == canonical {
		return path, nil
	}
	if path != "/lib/systemd/system/caddy.service" {
		return "", errors.New("unsupported caddy.service unit location")
	}
	info, err := os.Lstat("/lib")
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return path, nil
	} // unmerged /usr; ordinary ancestry checks still apply
	uid, _, _ := ownership(info)
	target, err := os.Readlink("/lib")
	if err != nil || uid != 0 || (target != "usr/lib" && target != "/usr/lib") {
		return "", errors.New("unsupported /lib alias for caddy.service")
	}
	// Only this fixed root-owned merged-/usr alias is allowed. All canonical
	// ancestors and the actual unit are checked separately, without symlinks.
	return canonical, nil
}
