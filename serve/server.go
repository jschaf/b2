package main

import (
	"fmt"
	"github.com/jschaf/b2/serve/livereload"
	"github.com/jschaf/b2/serve/paths"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

func main() {
	port := "8080"

	liveReload := livereload.New()
	lrJSPath := "/dev/livereload.js"
	lrPath := "/dev/livereload"
	http.HandleFunc(lrJSPath, liveReload.ServeJSHandler)
	http.HandleFunc(lrPath, liveReload.WebSocketHandler)
	go liveReload.Start()

	root, err := paths.FindRootDir()
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

	watcher := paths.NewFSWatcher(liveReload)
	path := filepath.Join(pubDir, "circle_ci_fast_git")
	err = watcher.Add(path)
	if err != nil {
		log.Printf("failed to add to path: %s", err)
	}
	go watcher.Start()

	log.Printf("Serving at port %s", port)
	log.Fatal(http.ListenAndServe("127.0.0.1:"+port, nil))
}
