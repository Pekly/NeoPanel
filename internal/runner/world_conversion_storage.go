package runner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/Pekly/NeoPanel/internal/chunker"
)

func conversionRetainedBytes(ctx context.Context, base, id string) (int64, error) {
	dir, err := os.Open(base)
	if err != nil {
		return 0, err
	}
	defer dir.Close()
	var total int64
	seen, count := 0, 0
	for {
		names, e := dir.Readdirnames(128)
		seen += len(names)
		if seen > chunker.MaxEntries {
			return 0, errors.New("storage inspection entry limit exceeded")
		}
		for _, name := range names {
			if !strings.HasPrefix(name, ".chunker-"+id+"-") && !strings.HasPrefix(name, ".chunker-original-"+id+"-") {
				continue
			}
			count++
			if count > 6 {
				return 0, errors.New("too many retained conversion directories; administrator review required")
			}
			path := filepath.Join(base, name)
			st, e := os.Lstat(path)
			if e != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
				return 0, errors.New("unsafe conversion recovery directory")
			}
			_, size, e := chunker.Scan(ctx, path, chunker.MaxBytes*3)
			if e != nil {
				return 0, e
			}
			total += size
		}
		if len(names) == 0 {
			return total, errOrEOF(e)
		}
	}
}
