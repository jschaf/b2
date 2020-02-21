package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/livereload"

	"github.com/cloudflare/tableflip"
)

var (
	upgradeC chan os.Signal
)

func hotSwapServer() {
	upgradeC <- syscall.SIGHUP
}

func main() {
	port := "8080"

	liveReload := livereload.NewWebsocketServer()
	lrJSPath := "/dev/livereload.js"
	lrPath := "/dev/livereload"
	http.HandleFunc(lrJSPath, livereload.ServeJSHandler)
	http.HandleFunc(lrPath, liveReload.WebSocketHandler)
	go liveReload.Start()

	root, err := git.FindRootDir()
	if err != nil {
		log.Fatalf("failed to find root dir: %s", err)
	}
	pubDir := filepath.Join(root, "public")
	log.Printf("Serving dir %s", pubDir)
	pubDirHandler := http.FileServer(http.Dir(pubDir))

	lrScript := strings.Join([]string{
		fmt.Sprintf("<script defer src=%s?port=%s&path=%s type='application/javascript'>",
			lrJSPath, port, strings.TrimLeft(lrPath, "/")),
		"</script>",
	}, "")
	http.Handle("/", livereload.NewHTMLInjector(lrScript, pubDirHandler))

	watcher := NewFSWatcher(liveReload)
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

	// Upgrade on SIGHUP.
	go func() {
		upgradeC = make(chan os.Signal, 1)
		signal.Notify(upgradeC, syscall.SIGHUP)
		for range upgradeC {
			_ = upg.Upgrade()
		}
	}()

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

	go watcher.Start()

	go http.Serve(ln, nil)

	if err := upg.Ready(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Serving at port http://localhost:%s", port)

	<-upg.Exit()
}

func mustWatchDir(watcher *FSWatcher, dir string) {
	if err := watcher.AddRecursively(dir); err != nil {
		log.Fatalf("failed to watch path %s: %s", dir, err)
	}

}
