package runtime

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type PluginDataStore struct {
	basePath string
}

func NewPluginDataStore(basePath string) *PluginDataStore {
	return &PluginDataStore{
		basePath: basePath,
	}
}

func (s *PluginDataStore) Exists(name string) bool {
	clean, err := cleanPath(name)
	if err != nil {
		return false
	}

	full := filepath.Join(s.basePath, clean)

	if _, err := os.Stat(full); err != nil {
		return false
	}

	return true
}

func (s *PluginDataStore) Read(name string) (io.ReadCloser, error) {
	clean, err := cleanPath(name)
	if err != nil {
		return nil, err
	}

	return os.Open(filepath.Join(s.basePath, clean))
}

func (s *PluginDataStore) List(path string, ext []string) ([]string, error) {
	var out []string

	dirPath := filepath.Join(s.basePath, path)

	err := filepath.Walk(dirPath, func(fpath string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(s.basePath, fpath)
		if err != nil {
			return fmt.Errorf("cannot get relative path to base dir '%s' for file: '%s': %w", s.basePath, fpath, err)
		}

		if len(ext) == 0 {
			out = append(out, relPath)

			return nil
		}

		name := info.Name()
		for _, e := range ext {
			if strings.HasSuffix(name, e) {
				out = append(out, relPath)

				return nil
			}
		}

		return nil
	})

	return out, err
}
