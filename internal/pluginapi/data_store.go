package pluginapi

import "io"

type DataStore interface {
	Exists(name string) bool
	Read(name string) (io.ReadCloser, error)
	List(path string, ext []string) ([]string, error)
}
