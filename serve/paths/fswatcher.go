package paths

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/jschaf/b2/serve/livereload"
	"log"
	"os"
	"path/filepath"
)

// FSWatcher watches the filesystem for modifications and sends LiveReload
// commands to the browser client.
type FSWatcher struct {
	lr      *livereload.LiveReload
	watcher *fsnotify.Watcher
}

func NewFSWatcher(lr *livereload.LiveReload) *FSWatcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	return &FSWatcher{
		lr:      lr,
		watcher: watcher,
	}
}

func (f *FSWatcher) Start() {
	defer f.watcher.Close()

	for {
		select {
		case event := <-f.watcher.Events:
			log.Println("event:", event)
			if event.Op&fsnotify.Write == fsnotify.Write {
				f.lr.ReloadFile(event.Name)
				log.Println("modified file:", event.Name)
			}
		case err := <-f.watcher.Errors:
			log.Println("error:", err)
		}
	}
}

func (f *FSWatcher) AddRecursively(name string) error {
	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}

		err = f.watcher.Add(path)
		if err != nil {
			return fmt.Errorf("failed to watch directory: %w", err)
		}
		log.Printf("Watching dir %s", path)
		return nil
	}

	return filepath.Walk(name, walk)
}
