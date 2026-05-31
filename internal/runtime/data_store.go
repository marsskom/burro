package runtime

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type PluginDataStore struct {
	basePath string
}

func NewPluginDataStore(basePath string) *PluginDataStore {
	return &PluginDataStore{
		basePath: basePath,
	}
}

func (s *PluginDataStore) Create(name string) (io.WriteCloser, error) {
	return nil, fmt.Errorf("plugin data store: create cannot be implemented")
}

func (s *PluginDataStore) Exists(name string) bool {
	clean := filepath.Clean("/" + name)[1:]
	full := filepath.Join(s.basePath, clean)

	if _, err := os.Stat(full); err != nil {
		return false
	}

	return true
}

func (s *PluginDataStore) Rename(oldpath, newpath string) error {
	return fmt.Errorf("plugin data store: rename cannot be implemented")
}

func (s *PluginDataStore) Write(name string, r io.Reader) (string, error) {
	return "", fmt.Errorf("plugin data store: write cannot be implemented")
}

func (s *PluginDataStore) Read(name string) (io.ReadCloser, error) {
	clean := filepath.Clean("/" + name)[1:]

	return os.Open(filepath.Join(s.basePath, clean))
}

func (s *PluginDataStore) Delete(name string) error {
	return fmt.Errorf("plugin data store: delete cannot be implemented")
}

func (s *PluginDataStore) List() ([]string, error) {
	var out []string

	err := filepath.Walk(s.basePath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			out = append(out, path)
		}

		return nil
	})

	return out, err
}
