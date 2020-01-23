package paths

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/jschaf/b2/serve/livereload"
	"log"
	"os"
	"path/filepath"
	"strings"
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

func (f *FSWatcher) Start() error {
	defer f.watcher.Close()
	root, err := FindRootDir()
	if err != nil {
		return fmt.Errorf("failed to get root dir: %w", err)
	}

	for {
		select {
		case event := <-f.watcher.Events:
			if event.Op == fsnotify.Chmod || strings.HasSuffix(event.Name, "~") {
				// Intellij temp file
				break
			}
			if event.Op&fsnotify.Write != fsnotify.Write {
				break
			}
			log.Printf("event: %s", event)

			rel, err := filepath.Rel(root, event.Name)
			if err != nil {
				rel = ""
			}
			switch rel {
			case "style/main.css":
				dest := filepath.Join(root, "public", "style", "main.css")
				err := os.MkdirAll(filepath.Dir(dest), 0755)
				if err != nil {
					log.Printf("failed to create dir public/style")
				}
				err = Copy(event.Name, dest)
				if err != nil {
					log.Printf("failed to copy main.css into public: %s", err)
				}
				f.lr.ReloadFile(dest)
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
