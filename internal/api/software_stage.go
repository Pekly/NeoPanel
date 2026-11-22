package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/Pekly/NeoPanel/internal/downloader"
	"github.com/Pekly/NeoPanel/internal/rooted"
	"github.com/Pekly/NeoPanel/internal/serverjar"
)

var errSoftwareStage = errors.New("invalid staged software distribution")

// The downloader already validates/extracts its distribution. This additional
// boundary checks its approved outputs before reinstall removes any live files;
// it does not implement another archive extractor or execute install scripts.
func validateSoftwareStage(ctx context.Context, stage *rooted.Root, dir, software, filename string, info serverjar.Info) error {
	native := downloader.IsNativeSoftware(software)
	var count int
	var total int64
	err := stage.Walk(".", func(name string, st os.FileInfo, err error) error {
		if err != nil {
			return errSoftwareStage
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if name == "." {
			return nil
		}
		count++
		if count > 50000 || (!st.Mode().IsRegular() && !st.IsDir()) {
			return errSoftwareStage
		}
		name = filepath.ToSlash(name)
		allowed := name == filename && st.Mode().IsRegular()
		if native {
			allowed = allowed || name == "config" && st.IsDir() || name == "config/config.toml" && st.Mode().IsRegular()
		} else {
			allowed = allowed || name == "eula.txt" && st.Mode().IsRegular() || name == "libraries" && st.IsDir() || strings.HasPrefix(name, "libraries/")
		}
		if !allowed {
			return errSoftwareStage
		}
		if st.Mode().IsRegular() {
			if st.Size() < 0 || st.Size() > (1<<30)-total {
				return errSoftwareStage
			}
			total += st.Size()
		}
		return nil
	})
	if err != nil {
		return err
	}
	if info.Filename != filename {
		return errSoftwareStage
	}
	if !native {
		fresh, err := serverjar.Inspect(ctx, dir, filename)
		if err != nil || info.SHA256 == "" || fresh.SHA256 != info.SHA256 || fresh.Size != info.Size {
			return errSoftwareStage
		}
		return nil
	}
	f, err := stage.Open(filename)
	if err != nil {
		return errSoftwareStage
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil || st.Size() == 0 {
		return errSoftwareStage
	}
	// Native distributions may not have JAR metadata, but verify a supplied
	// digest rather than interpreting the executable as ZIP or running it.
	if info.SHA256 != "" {
		h := sha256.New()
		n, err := rooted.Copy(ctx, h, f, 1<<30)
		if err != nil || n != info.Size || hex.EncodeToString(h.Sum(nil)) != info.SHA256 {
			return errSoftwareStage
		}
	}
	return ctx.Err()
}

// Same-root renames retain the validated bytes/modes (including native execute
// bits). Reinstall has already removed old runtime outputs under its exclusive
// lease. Publication is per-output, not a crash-atomic tree transaction.
func publishSoftwareStage(ctx context.Context, root, stage *rooted.Root, stageName, software, filename string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if downloader.IsNativeSoftware(software) {
		config := filepath.Join("config", "config.toml")
		if _, err := root.Lstat(config); os.IsNotExist(err) {
			if err := root.Check("config"); err != nil {
				return err
			}
			if err := root.MkdirAll("config", 0755); err != nil {
				return err
			}
			if err := root.Rename(filepath.Join(stageName, config), config); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	} else {
		for _, name := range []string{"libraries", "eula.txt"} {
			if _, err := stage.Lstat(name); os.IsNotExist(err) {
				continue
			} else if err != nil {
				return err
			}
			if err := root.Rename(filepath.Join(stageName, name), name); err != nil {
				return err
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return root.Rename(filepath.Join(stageName, filename), filename)
}
