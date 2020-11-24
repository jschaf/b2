package main

import (
	"flag"
	"fmt"
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/livereload"
	"github.com/jschaf/b2/pkg/log"
	"github.com/jschaf/b2/pkg/net/srv"
	"github.com/jschaf/b2/pkg/sites"
	"go.uber.org/zap"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func run() (mErr error) {
	level, logger, err := log.ParseFlags()
	if err != nil {
		return fmt.Errorf("parse log flags: %w", err)
	}

	port := "8080"
	server := srv.NewServer(srv.ServerCfg{
		Name:     srv.NameB2,
		Version:  "dev",
		HTTPAddr: "0.0.0.0:8080",
	}, level, logger)
	mux := http.NewServeMux()
	server.SetHTTPHandler(mux)
	pubDir := dirs.PublicMemfs

	if err := dirs.CleanDir(pubDir); err != nil {
		return fmt.Errorf("clean public dir: %w", err)
	}

	lrJSPath := "/dev/livereload.js"
	lrPath := "/dev/livereload"
	lr := livereload.NewServer(logger.Sugar().Named("livereload"))
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

	watcher := NewFSWatcher(pubDir, lr, logger.Sugar())
	root := git.MustFindRootDir()
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
	if err := sites.Rebuild(pubDir, logger); err != nil {
		return fmt.Errorf("rebuild site: %w", err)
	}

	go func() {
		if err := watcher.Start(); err != nil {
			logger.Error("watcher error", zap.Error(err))
			server.Stop()
		}
	}()

	server.AddDrainHandler(watcher.Stop)
	server.AddDrainHandler(lr.Shutdown)
	return server.ListenAndServe()
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Printf("ERROR: failed to run b2 server: " + err.Error())
		os.Exit(1)
	}
}
