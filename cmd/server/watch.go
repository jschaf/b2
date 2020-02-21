package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/livereload"
	"github.com/jschaf/b2/pkg/markdown"
	"github.com/jschaf/b2/pkg/markdown/compiler"
	"github.com/jschaf/b2/pkg/paths"
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
	rootDir, err := git.FindRootDir()
	if err != nil {
		return fmt.Errorf("failed to get root dir: %w", err)
	}

	publicDir := filepath.Join(rootDir, "public")

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

			rel, err := filepath.Rel(rootDir, event.Name)
			if err != nil {
				rel = ""
			}
			if rel == "style/main.css" {
				f.reloadMainCSS(rootDir, event)
			} else if filepath.Ext(rel) == ".md" {
				if err := f.compileReloadMd(event.Name, publicDir); err != nil {
					return fmt.Errorf("failed to compiled changed markdown: %w", err)
				}
			} else if filepath.Ext(rel) == ".go" {
				log.Printf("hot swapping server")
				if err := rebuildServer(); err != nil {
					return fmt.Errorf("failed to hotswap erver: %w", err)
				}
			}

		case err := <-f.watcher.Errors:
			log.Println("error:", err)
		}
	}
}

func (f *FSWatcher) compileReloadMd(path string, publicDir string) error {
	c := compiler.New(markdown.New())
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	if err := c.CompileIntoDir(file, publicDir); err != nil {
		return fmt.Errorf("failed to compile md file: %s", err)
	}
	f.lr.ReloadFile(path)
	return nil
}

func (f *FSWatcher) reloadMainCSS(root string, event fsnotify.Event) {
	dest := filepath.Join(root, "public", "style", "main.css")
	err := os.MkdirAll(filepath.Dir(dest), 0755)
	if err != nil {
		log.Printf("failed to create dir public/style")
	}
	err = paths.Copy(event.Name, dest)
	if err != nil {
		log.Printf("failed to copy main.css into public: %s", err)
	}
	f.lr.ReloadFile(dest)
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
		return nil
	}

	return filepath.Walk(name, walk)
}

func rebuildServer() error {
	out := os.Args[0]
	pkg := "github.com/jschaf/b2/cmd/server"
	cmd := exec.Command("go", "build", "-o", out, pkg)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server build: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to rebuild server: %s", err)
	}
	return nil
}
