package pluginapi

import "io"

type ArtifactStore interface {
	Create(name string) (io.WriteCloser, error)
	Exists(name string) bool
	Rename(oldpath, newpath string) error
	Write(name string, r io.Reader) (string, error)
	Read(name string) (io.ReadCloser, error)
	Delete(name string) error
	List() ([]string, error)
}
