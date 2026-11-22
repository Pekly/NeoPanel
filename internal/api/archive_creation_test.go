package api

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/rooted"
	"github.com/Pekly/NeoPanel/internal/runner"
)

type archiveOutputFunc func([]byte) (int, error)

func (f archiveOutputFunc) Write(p []byte) (int, error) { return f(p) }

func TestArchiveCreationPreflightAndNormalFormats(t *testing.T) {
	for _, format := range []string{"zip", "tar.gz"} {
		t.Run(format, func(t *testing.T) {
			dir := t.TempDir()
			f, err := os.Create(filepath.Join(dir, "sparse"))
			if err != nil {
				t.Fatal(err)
			}
			if err = f.Truncate(16 << 20); err != nil {
				t.Fatal(err)
			}
			f.Close()
			root, err := rooted.Open(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			sources := []archiveSource{{"sparse", "sparse"}}
			wrote := false
			_, err = buildArchiveContext(context.Background(), root, archiveOutputFunc(func(p []byte) (int, error) { wrote = true; return len(p), nil }), sources, nil, format, 1<<20)
			if !errors.Is(err, runner.ErrArchiveLimit) || wrote {
				t.Fatal("preflight streamed", err, wrote)
			}
			if err = os.WriteFile(filepath.Join(dir, "sparse"), []byte("normal contents"), 0600); err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			n, err := buildArchiveContext(context.Background(), root, &out, sources, nil, format, 1<<20)
			if err != nil || n != 1 {
				t.Fatal(n, err)
			}
			var content []byte
			if format == "zip" {
				z, err := zip.NewReader(bytes.NewReader(out.Bytes()), int64(out.Len()))
				if err != nil {
					t.Fatal(err)
				}
				if len(z.File) != 1 || z.File[0].Name != "sparse" {
					t.Fatal(z.File)
				}
				in, err := z.File[0].Open()
				if err != nil {
					t.Fatal(err)
				}
				content, err = io.ReadAll(in)
				in.Close()
				if err != nil {
					t.Fatal(err)
				}
			} else {
				gz, err := gzip.NewReader(bytes.NewReader(out.Bytes()))
				if err != nil {
					t.Fatal(err)
				}
				defer gz.Close()
				tr := tar.NewReader(gz)
				h, err := tr.Next()
				if err != nil || h.Name != "sparse" {
					t.Fatal(h, err)
				}
				content, err = io.ReadAll(tr)
				if err != nil {
					t.Fatal(err)
				}
				if _, err = tr.Next(); err != io.EOF {
					t.Fatal(err)
				}
			}
			if string(content) != "normal contents" {
				t.Fatal(string(content))
			}
			// Cancellation after compression starts must remove Publish's temporary file.
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			err = root.Publish("cancelled", 0600, func(out *os.File) error {
				_, err := buildArchiveContext(ctx, root, archiveOutputFunc(func(p []byte) (int, error) { cancel(); return out.Write(p) }), sources, nil, format, 1<<20)
				return err
			})
			if !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) != 1 || entries[0].Name() != "sparse" {
				t.Fatal(entries, err)
			}
		})
	}
}

func TestCompressHTTPBudgetsCancellationAndGates(t *testing.T) {
	f := newReliabilityFixture(t)
	srv, err := f.db.GetServer("alpha")
	if err != nil {
		t.Fatal(err)
	}
	srv.DiskLimitMB = 1
	if err = f.db.UpdateServer(srv); err != nil {
		t.Fatal(err)
	}
	previewPut(t, f, "source", []byte("keep"))
	path := filepath.Join(f.manager.GetWorkingDir("alpha"), "source")
	if err = os.Truncate(path, 16<<20); err != nil {
		t.Fatal(err)
	}
	h := NewFileHandler(f.db, f.manager)
	request := func(ctx context.Context, name string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "/", strings.NewReader(`{"paths":["source"],"filename":"`+name+`"}`)).WithContext(ctx)
		r.SetPathValue("id", "alpha")
		w := httptest.NewRecorder()
		h.Compress(w, r)
		return w
	}
	if w := request(context.Background(), "large.zip"); w.Code != http.StatusRequestEntityTooLarge {
		t.Fatal(w.Code, w.Body.String())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if w := request(ctx, "cancel.zip"); w.Code != http.StatusRequestTimeout {
		t.Fatal(w.Code, w.Body.String())
	}
	if err = os.WriteFile(path, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	_, _, release, err := f.manager.BeginArchiveCreation(context.Background(), "alpha")
	if err != nil {
		t.Fatal(err)
	}
	w := request(context.Background(), "busy.zip")
	release()
	if w.Code != http.StatusConflict {
		t.Fatal(w.Code, w.Body.String())
	}
	if w = request(context.Background(), "normal.zip"); w.Code != http.StatusOK {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, name := range []string{"large.zip", "cancel.zip", "busy.zip"} {
		if _, err = os.Stat(filepath.Join(filepath.Dir(path), name)); !os.IsNotExist(err) {
			t.Fatal(name, err)
		}
	}
	other := runner.NewServerManager(t.TempDir(), nil)
	defer other.ShutdownAll()
	_, _, release1, err := other.BeginArchiveCreation(context.Background(), "one")
	if err != nil {
		t.Fatal(err)
	}
	_, _, release2, err := other.BeginArchiveCreation(context.Background(), "two")
	if err != nil {
		release1()
		t.Fatal(err)
	}
	w = request(context.Background(), "capacity.zip")
	release2()
	release1()
	if w.Code != http.StatusTooManyRequests || w.Header().Get("Retry-After") == "" {
		t.Fatal(w.Code, w.Header())
	}
}
