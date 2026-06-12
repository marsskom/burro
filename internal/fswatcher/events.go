package fswatcher

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
)

func (w *FSWatcher) onCreate(event fsnotify.Event) error {
	path := event.Name

	i, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot read created path: %w", err)
	}

	if i.IsDir() {
		if !w.recursive {
			return nil
		}

		return w.addDirs(path, w.recursive)
	}

	w.Events <- FSEvent{
		Path: path,
		Op:   OpCreated,
	}

	return nil
}

func (w *FSWatcher) onRemove(event fsnotify.Event) error {
	w.Events <- FSEvent{
		Path: event.Name,
		Op:   OpDelete,
	}

	return nil
}

func (w *FSWatcher) onWrite(event fsnotify.Event) error {
	w.Events <- FSEvent{
		Path: event.Name,
		Op:   OpUpdate,
	}

	return nil
}

func (w *FSWatcher) onRename(event fsnotify.Event) error {
	w.Events <- FSEvent{
		Path: event.Name,
		Op:   OpDelete,
	}

	return nil
}
