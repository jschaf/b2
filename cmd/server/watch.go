package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/livereload"
	"github.com/jschaf/b2/pkg/markdown"
	"github.com/jschaf/b2/pkg/markdown/compiler"
	"github.com/jschaf/b2/pkg/paths"
	"go.uber.org/zap"
)

// FSWatcher watches the filesystem for modifications and sends LiveReload
// commands to the browser client.
type FSWatcher struct {
	liveReload *livereload.LiveReload
	watcher    *fsnotify.Watcher
	logger     *zap.SugaredLogger
}

func NewFSWatcher(lr *livereload.LiveReload, logger *zap.SugaredLogger) *FSWatcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	return &FSWatcher{
		liveReload: lr,
		watcher:    watcher,
		logger:     logger.Named("watcher"),
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
			// Ignore everything except writes.
			if event.Op&fsnotify.Write != fsnotify.Write {
				break
			}

			rel, err := filepath.Rel(rootDir, event.Name)
			if err != nil {
				f.logger.Infof("failed to get relative path: %s", err)
				break
			}

			switch {
			case rel == "style/main.css":
				f.reloadMainCSS(rootDir, event)

			case filepath.Ext(rel) == ".md":
				if err := f.compileReloadMd(event.Name, publicDir); err != nil {
					return fmt.Errorf("failed to compiled changed markdown: %w", err)
				}

			case strings.HasPrefix(rel, "pkg/markdown/"):
				// If only the markdown has changed, recompile only that.
				if err := f.compileMdWithGoRun(); err != nil {
					return err
				}
				// Send empty string which should reload all LiveReload clients
				f.liveReload.ReloadFile("")

			case filepath.Ext(rel) == ".go":
				if err := f.rebuildServer(); err != nil {
					return fmt.Errorf("failed to hotswap erver: %w", err)
				}
			}

		case err := <-f.watcher.Errors:
			f.logger.Infof("error:", err)
		}
	}
}

func (f *FSWatcher) compileMdWithGoRun() error {
	f.logger.Infof("pkg/markdown changed, compiling all markdown")
	cmd := exec.Command("go", "run", "github.com/jschaf/b2/cmd/compiler")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start go run compiler: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("go run compiler failed: %w", err)
	}
	return nil
}

func (f *FSWatcher) watchDirs(dirs ...string) error {
	for _, dir := range dirs {
		if err := f.AddRecursively(dir); err != nil {
			return fmt.Errorf("failed to watch dir: %w", err)
		}
	}
	return nil
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
	f.liveReload.ReloadFile(path)
	return nil
}

func (f *FSWatcher) reloadMainCSS(root string, event fsnotify.Event) {
	dest := filepath.Join(root, "public", "style", "main.css")
	err := os.MkdirAll(filepath.Dir(dest), 0755)
	if err != nil {
		f.logger.Info("failed to create dir public/style")
	}
	err = paths.Copy(event.Name, dest)
	if err != nil {
		f.logger.Infof("failed to copy main.css into public: %s", err)
	}
	f.liveReload.ReloadFile(dest)
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

func (f *FSWatcher) rebuildServer() error {
	f.logger.Info("hot swapping server because go file changed")
	out := os.Args[0]
	pkg := "github.com/jschaf/b2/cmd/server"
	cmd := exec.Command("go", "build", "-o", out, pkg)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server build: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to rebuild server: %s", err)
	}

	f.logger.Debug("sending SIGHUP")
	if err := sendSighup(); err != nil {
		return err
	}
	return nil
}

func sendSighup() error {
	pid := os.Getpid()
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to get process from PID: %w", err)
	}
	if err = process.Signal(syscall.SIGHUP); err != nil {
		return fmt.Errorf("failed to send SIGHUP: %w", err)
	}
	return nil
}
