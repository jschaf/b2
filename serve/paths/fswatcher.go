package paths

import (
	"github.com/fsnotify/fsnotify"
	"github.com/jschaf/b2/serve/livereload"
	"log"
)

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

func (f *FSWatcher) Add(name string) error {
	return f.watcher.Add(name)
}
