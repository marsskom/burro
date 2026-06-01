package runtime

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func cleanPath(name string) (string, error) {
	return filepath.Rel("/", filepath.Clean("/"+name))
}

type FileArtifactStore struct {
	basePath string
}

func NewFileArtifactStore(basePath string) *FileArtifactStore {
	return &FileArtifactStore{
		basePath: basePath,
	}
}

func (s *FileArtifactStore) Create(name string) (io.WriteCloser, error) {
	clean, err := cleanPath(name)
	if err != nil {
		return nil, err
	}

	full := filepath.Join(s.basePath, clean)

	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return nil, err
	}

	return os.Create(full)
}

func (s *FileArtifactStore) Exists(name string) bool {
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

func (s *FileArtifactStore) Rename(oldpath, newpath string) error {
	if !s.Exists(oldpath) {
		return fmt.Errorf("artifacts: oldpath '%s', doesn't exist", oldpath)
	}
	if s.Exists(newpath) {
		return fmt.Errorf("artifacts: newpath '%s' already in use", newpath)
	}

	cleanOP, err := cleanPath(oldpath)
	if err != nil {
		return err
	}

	cleanNP, err := cleanPath(newpath)
	if err != nil {
		return err
	}

	fullOP := filepath.Join(s.basePath, cleanOP)
	fullNP := filepath.Join(s.basePath, cleanNP)

	err = os.Rename(fullOP, fullNP)
	if err != nil {
		return fmt.Errorf("artifacts: cannot rename '%s' to '%s': %w", oldpath, newpath, err)
	}

	return err
}

func (s *FileArtifactStore) Write(name string, r io.Reader) (string, error) {
	clean, err := cleanPath(name)
	if err != nil {
		return "", err
	}

	full := filepath.Join(s.basePath, clean)

	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return "", fmt.Errorf("artifacts: directory structure doesn't exist and cannot be created: %w", err)
	}

	f, err := os.Create(full)
	if err != nil {
		return "", fmt.Errorf("artifacts: file cannot be created: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	if err != nil {
		return "", fmt.Errorf("artifacts: cannot write to file: %w", err)
	}

	return full, nil
}

func (s *FileArtifactStore) Read(name string) (io.ReadCloser, error) {
	clean, err := cleanPath(name)
	if err != nil {
		return nil, err
	}

	return os.Open(filepath.Join(s.basePath, clean))
}

func (s *FileArtifactStore) Delete(name string) error {
	clean, err := cleanPath(name)
	if err != nil {
		return err
	}

	return os.Remove(filepath.Join(s.basePath, clean))
}

func (s *FileArtifactStore) List() ([]string, error) {
	var out []string

	err := filepath.Walk(s.basePath, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() {
			out = append(out, path)
		}

		return nil
	})

	return out, err
}
