package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
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

var (
	upgradeC chan os.Signal
)

func hotSwapServer() error {
	out := os.Args[0]
	pkg := "github.com/jschaf/b2/cmd/server"
	cmd := exec.Command("go", "build", "-o", out, pkg)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server build: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to rebuild server: %s", err)
	}

	upgradeC <- syscall.SIGHUP
	return nil
}

type server struct {
	*http.ServeMux
	port  string
	stopC chan struct{}
	once  sync.Once
}

func newServer(port string) *server {
	s := new(server)
	s.ServeMux = http.NewServeMux()
	s.port = port
	s.once = sync.Once{}
	s.stopC = make(chan struct{})
	return s
}

func (s *server) Serve(ln net.Listener) error {
	srv := http.Server{
		Handler: s.ServeMux,
	}
	return srv.Serve(ln)
}

func (s *server) stop() {
	s.once.Do(func() {
		close(s.stopC)
	})
}

func main() {
	port := "8080"
	server := newServer(port)
	stopC := make(chan struct{})

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

	upg, err := tableflip.New(tableflip.Options{
		UpgradeTimeout: time.Second * 5,
	})
	if err != nil {
		log.Fatalf("failed to create upgrader: %s", err)
	}
	defer upg.Stop()

	// Upgrade on SIGHUP
	go func() {
		upgradeC = make(chan os.Signal, 1)
		signal.Notify(upgradeC, syscall.SIGHUP)
		for range upgradeC {
			_ = upg.Upgrade()
		}
	}()

	// Stop upgrader on quit so it doesn't restart.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
		<-c
		upg.Stop()
	}()

	ln, err := upg.Listen("tcp", "localhost:"+port)
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}
	defer ln.Close()

	go func() {
		if err := watcher.Start(); err != nil {
			log.Printf("stoping server because watcher error: %s", err)
			server.stop()
		}
	}()

	go func() {
		if err := server.Serve(ln); err != nil {
			server.stop()
		}
	}()

	if err := upg.Ready(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Serving http://localhost:%s", port)

	// Compile stuff because it might have changed.
	md := markdown.New()
	c := compiler.New(md)
	if err := compiler.CompileEverything(c); err != nil {
		log.Fatal(err)
	}

	select {
	case <-upg.Exit():
	case <-stopC:
		upg.Stop()
	}
	<-upg.Exit()
}

func mustWatchDir(watcher *FSWatcher, dir string) {
	if err := watcher.AddRecursively(dir); err != nil {
		log.Fatalf("failed to watch path %s: %s", dir, err)
	}

}
