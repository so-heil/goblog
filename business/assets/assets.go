package assets

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
)

//go:embed static
var files embed.FS

type Assets struct {
	embed.FS
}

func New() *Assets {
	return &Assets{files}
}

func (a *Assets) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.FileServer(http.FS(a)).ServeHTTP(w, r)
}

func (a *Assets) RecursiveCopy(target string, perm os.FileMode) error {
	err := fs.WalkDir(a, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("path %s walkdir: %w", path, err)
		}

		targetPath := filepath.Join(target, path)

		if d.IsDir() {
			return os.MkdirAll(targetPath, perm)
		}

		src, err := files.Open(path)
		if err != nil {
			return fmt.Errorf("open src %s: %w", path, err)
		}
		defer src.Close()

		dst, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("create dst: %w", err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return fmt.Errorf("copy src to dst: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("fs walkdir: %w", err)
	}

	return nil
}
