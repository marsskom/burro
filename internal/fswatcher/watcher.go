package fswatcher

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FSWatcher struct {
	watcher   *fsnotify.Watcher
	root      string
	ext       []string
	recursive bool

	Events chan FSEvent

	Errors chan error

	debounce map[string]*time.Timer
	delay    time.Duration

	mu sync.Mutex
}

type FSEvent struct {
	Path string
	Op   FSEventOperation
}

type FSEventOperation uint32

const (
	OpCreated FSEventOperation = 1 << iota
	OpUpdate
	OpDelete
)

func NewFSWatcher(root string, ext []string, recursive bool) (*FSWatcher, error) {
	if i, err := os.Stat(root); err != nil || !i.IsDir() {
		return nil, fmt.Errorf("root directory doesn't exist: %w", err)
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("cannot create fs watcher: %w", err)
	}

	return &FSWatcher{
		watcher: w,

		root:      root,
		ext:       ext,
		recursive: recursive,

		Events: make(chan FSEvent, 50),

		Errors: make(chan error, 1),

		debounce: make(map[string]*time.Timer),
		delay:    150 * time.Millisecond,
	}, nil
}

func (w *FSWatcher) addDirs(root string, recursive bool) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !recursive {
		return w.watcher.Add(root)
	}

	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return w.watcher.Add(path)
		}

		return nil
	})
}

func (w *FSWatcher) Start() error {
	if err := w.addDirs(w.root, w.recursive); err != nil {
		return fmt.Errorf("cannot start fs watcher: %w", err)
	}

	go w.loop()

	return nil
}

func (w *FSWatcher) Stop() {
	w.mu.Lock()
	for _, t := range w.debounce {
		t.Stop()
	}
	w.mu.Unlock()

	_ = w.watcher.Close()
}

func (w *FSWatcher) loop() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleEvent(event)

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}

			w.Errors <- err
		}
	}
}

func (w *FSWatcher) handleEvent(event fsnotify.Event) {
	path := event.Name

	if len(w.ext) > 0 {
		if !slices.Contains(w.ext, filepath.Ext(path)) {
			return
		}
	}

	if strings.HasPrefix(filepath.Base(path), ".") {
		return
	}

	w.emit(event)
}

func (w *FSWatcher) emit(event fsnotify.Event) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if t, ok := w.debounce[event.Name]; ok {
		t.Stop()
	}

	w.debounce[event.Name] = time.AfterFunc(w.delay, func() {
		w.processEvent(event)
	})
}

func (w *FSWatcher) processEvent(event fsnotify.Event) {
	var err error

	switch {
	case event.Has(fsnotify.Create):
		err = w.onCreate(event)
	case event.Has(fsnotify.Remove):
		err = w.onRemove(event)
	case event.Has(fsnotify.Write):
		err = w.onWrite(event)
	case event.Has(fsnotify.Rename):
		err = w.onRename(event)
	}

	if err != nil {
		select {
		case w.Errors <- err:
		default:
		}
	}
}
