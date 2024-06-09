package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/livereload"
	"github.com/jschaf/b2/pkg/log"
	"github.com/jschaf/b2/pkg/net/srv"
	"github.com/jschaf/b2/pkg/process"
	"github.com/jschaf/b2/pkg/sites"
)

func main() {
	process.RunMain(runMain)
}

func runMain(context.Context) (mErr error) {
	fset := flag.CommandLine
	logLevel := log.DefineFlags(fset)
	flag.Parse()
	slog.SetLogLoggerLevel(logLevel)

	port := "8080"
	server := srv.NewServer(srv.ServerCfg{
		Name:     srv.NameB2,
		Version:  "dev",
		HTTPAddr: "0.0.0.0:8080",
	})
	mux := http.NewServeMux()
	server.SetHTTPHandler(mux)
	pubDir := dirs.PublicMemfs

	if err := dirs.CleanDir(pubDir); err != nil {
		return fmt.Errorf("clean public dir: %w", err)
	}

	lrJSPath := "/dev/livereload.js"
	lrPath := "/dev/livereload"
	lr := livereload.NewServer()
	mux.HandleFunc(lrJSPath, lr.ServeJSHandler)
	mux.HandleFunc(lrPath, lr.WebSocketHandler)
	go lr.Start()

	pubDirHandler := http.FileServer(http.Dir(pubDir))

	lrScript := strings.Join([]string{
		fmt.Sprintf("<script defer src=%s?port=%s&path=%s type='application/javascript'>",
			lrJSPath, port, strings.TrimLeft(lrPath, "/")),
		"</script>",
	}, "")
	mux.Handle("/", lr.NewHTMLInjector(lrScript, pubDirHandler))

	watcher := NewFSWatcher(pubDir, lr)
	root := git.RootDir()
	if err := watcher.watchDirs(
		filepath.Join(root, dirs.Book),
		filepath.Join(root, dirs.Cmd),
		filepath.Join(root, dirs.Pkg),
		filepath.Join(root, dirs.Posts),
		filepath.Join(root, dirs.Static),
		filepath.Join(root, dirs.Style),
		filepath.Join(root, dirs.TIL),
	); err != nil {
		return fmt.Errorf("watch dirs: %w", err)
	}

	// Rebuild in case content changed since last run.
	if err := sites.Rebuild(pubDir); err != nil {
		return fmt.Errorf("rebuild site: %w", err)
	}

	go func() {
		if err := watcher.Start(); err != nil {
			slog.Error("watcher error", "error", err)
			server.Stop()
		}
	}()

	server.AddDrainHandler(watcher.Stop)
	server.AddDrainHandler(lr.Shutdown)
	return server.ListenAndServe()
}
