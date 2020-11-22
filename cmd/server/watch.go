package main

import (
	"bytes"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/jschaf/b2/pkg/css"
	"github.com/jschaf/b2/pkg/errs"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/livereload"
	"github.com/jschaf/b2/pkg/sites"
	"github.com/jschaf/b2/pkg/static"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// FSWatcher watches the filesystem for modifications and sends LiveReload
// commands to browser clients.
type FSWatcher struct {
	liveReload *livereload.LiveReload
	watcher    *fsnotify.Watcher
	logger     *zap.SugaredLogger
	pubDir     string
}

func NewFSWatcher(pubDir string, lr *livereload.LiveReload, logger *zap.SugaredLogger) *FSWatcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	return &FSWatcher{
		pubDir:     pubDir,
		liveReload: lr,
		watcher:    watcher,
		logger:     logger.Named("watcher"),
	}
}

func (f *FSWatcher) Start() (mErr error) {
	defer errs.CapturingClose(&mErr, f.watcher, "close FSWatcher")
	rootDir := git.MustFindRootDir()

	for {
		select {
		case event := <-f.watcher.Events:
			f.logger.Infof("watcher event: %s", event.Name)
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
				f.reloadMainCSS()

			case strings.HasPrefix(rel, "static/"):
				f.logger.Infof("static reload: %s", rel)
				if err := static.CopyStaticFiles(f.pubDir); err != nil {
					return fmt.Errorf("failed to copy static files: %w", err)
				}
				// Send empty string which should reload all LiveReload clients
				f.liveReload.ReloadFile("")

			case filepath.Ext(rel) == ".md":
				if err := f.compileReloadMd(); err != nil {
					return fmt.Errorf("failed to compiled changed markdown: %w", err)
				}
				f.liveReload.ReloadFile(event.Name)

			case strings.HasPrefix(rel, "pkg/markdown/"):
				// If only the markdown has changed, recompile only that.
				if err := f.compileMdWithGoRun(); err != nil {
					return err
				}
				// Send empty string which should reload all LiveReload clients
				f.liveReload.ReloadFile("")

			case filepath.Ext(rel) == ".go" && !strings.HasSuffix(rel, "_test.go"):
				// Rebuild the server to pickup any new changes.
				if err := f.rebuildServer(); err != nil {
					f.logger.Errorf("rebuild server: %s", err)
				}
			}
		case err := <-f.watcher.Errors:
			f.logger.Infof("error: %s", err)
		}
	}
}

func (f *FSWatcher) compileMdWithGoRun() error {
	f.logger.Infof("pkg/markdown changed, compiling all markdown")
	cmd := exec.Command("go", "run", "github.com/jschaf/b2/cmd/compiler")
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start go run compiler: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("go run compiler failed: %w\n%s", err, buf.String())
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

func (f *FSWatcher) compileReloadMd() error {
	if err := sites.Rebuild(f.pubDir, f.logger.Desugar()); err != nil {
		return fmt.Errorf("rebuild for changed md: %w", err)
	}
	return nil
}

func (f *FSWatcher) reloadMainCSS() {
	stylePaths, err := css.CopyAllCSS(f.pubDir)
	if err != nil {
		f.logger.Info(err)
	}
	for _, stylePath := range stylePaths {
		f.liveReload.ReloadFile(stylePath)
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
		return nil
	}

	return filepath.Walk(name, walk)
}

func (f *FSWatcher) rebuildServer() error {
	f.logger.Info("hot swapping server because go file changed")
	out := os.Args[0]
	pkg := "github.com/jschaf/b2/cmd/server"
	cmd := exec.Command("go", "build", "-o", out, pkg)
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	now := time.Now()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start server rebuild: %w\n%s", err, buf.String())
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("wait for server rebuild: %w\n%s", err, buf.String())
	}
	f.logger.Infof("completed server rebuild in %.3f", time.Since(now).Seconds())
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
