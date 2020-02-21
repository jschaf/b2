package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/livereload"
	"github.com/jschaf/b2/pkg/markdown"
	"github.com/jschaf/b2/pkg/markdown/compiler"

	"github.com/cloudflare/tableflip"
)

type server struct {
	*http.ServeMux
	port     string
	stopC    chan struct{}
	once     sync.Once
	upgrader *tableflip.Upgrader
}

func newServer(port string) *server {
	s := new(server)
	s.ServeMux = http.NewServeMux()
	s.port = port
	s.once = sync.Once{}
	s.stopC = make(chan struct{})

	return s
}

func (s *server) Serve() error {
	srv := http.Server{
		Handler: s.ServeMux,
	}

	upg, err := tableflip.New(tableflip.Options{
		UpgradeTimeout: time.Second * 5,
	})
	if err != nil {
		return fmt.Errorf("failed to create upgrader")
	}
	s.upgrader = upg
	if err := s.upgrader.Ready(); err != nil {
		return fmt.Errorf("upgrader not ready: %w", err)
	}

	ln, err := s.upgrader.Listen("tcp", "localhost:"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer ln.Close()
	return srv.Serve(ln)
}

func (s *server) UpgradeOnSIGHUP() {
	upgrade := make(chan os.Signal, 1)
	signal.Notify(upgrade, syscall.SIGHUP)
	// We might get multiple upgrade requests.
	for range upgrade {
		if err := s.upgrader.Upgrade(); err != nil {
			fmt.Printf("failed to upgrade: %s", err)
			continue
		}
	}
}

func (s *server) Stop() {
	s.once.Do(func() {
		s.upgrader.Stop()
		close(s.stopC)
	})
}

func main() {
	port := "8080"
	server := newServer(port)
	defer server.Stop()

	lr := livereload.NewWebsocketServer()
	lrJSPath := "/dev/livereload.js"
	lrPath := "/dev/livereload"
	server.HandleFunc(lrJSPath, livereload.ServeJSHandler)
	server.HandleFunc(lrPath, lr.WebSocketHandler)
	go lr.Start()

	root, err := git.FindRootDir()
	if err != nil {
		log.Fatalf("failed to find root dir: %s", err)
	}
	pubDir := filepath.Join(root, "public")
	pubDirHandler := http.FileServer(http.Dir(pubDir))

	lrScript := strings.Join([]string{
		fmt.Sprintf("<script defer src=%s?port=%s&path=%s type='application/javascript'>",
			lrJSPath, port, strings.TrimLeft(lrPath, "/")),
		"</script>",
	}, "")
	server.Handle("/", livereload.NewHTMLInjector(lrScript, pubDirHandler))

	watcher := NewFSWatcher(lr)
	mustWatchDir(watcher, filepath.Join(root, "public"))
	mustWatchDir(watcher, filepath.Join(root, "style"))
	mustWatchDir(watcher, filepath.Join(root, "posts"))
	mustWatchDir(watcher, filepath.Join(root, "cmd"))
	mustWatchDir(watcher, filepath.Join(root, "pkg"))

	go server.UpgradeOnSIGHUP()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
		<-c
		server.Stop()
	}()

	go func() {
		if err := watcher.Start(); err != nil {
			log.Printf("stoping server because watcher error: %s", err)
			server.Stop()
		}
	}()

	go func() {
		if err := server.Serve(); err != nil {
			server.Stop()
		}
	}()

	log.Printf("Serving at http://localhost:%s", port)

	// Compile stuff because it might have changed.
	md := markdown.New()
	c := compiler.New(md)
	if err := compiler.CompileEverything(c); err != nil {
		log.Fatal(err)
	}

	select {
	case <-server.upgrader.Exit():
	case <-server.stopC:
		server.upgrader.Stop()
	}
}

func mustWatchDir(watcher *FSWatcher, dir string) {
	if err := watcher.AddRecursively(dir); err != nil {
		log.Fatalf("failed to watch path %s: %s", dir, err)
	}

}
