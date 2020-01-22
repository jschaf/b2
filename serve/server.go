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

	root, err := paths.FindRootDir()
	if err != nil {
		log.Fatalf("failed to find root dir: %s", err)
	}
	pubDir := filepath.Join(root, "public")
	log.Printf("Serving dir %s", pubDir)
	pubDirHandler := http.FileServer(http.Dir(pubDir))

	liveReload := livereload.New()
	lrJSPath := "/dev/livereload.js"
	lrPath := "/dev/livereload"
	http.HandleFunc(lrJSPath, liveReload.ServeJSHandler)
	http.HandleFunc(lrPath, liveReload.WebSocketHandler)

	lrScript := strings.Join([]string{
		fmt.Sprintf("<script defer src=%s?port=%s&path=%s type='application/javascript'>",
			lrJSPath, port, strings.TrimLeft(lrPath, "/")),
		"</script>",
	}, "")
	http.Handle("/", livereload.NewHTMLInjector(lrScript, pubDirHandler))

	log.Printf("Serving at port %s", port)
	log.Fatal(http.ListenAndServe("127.0.0.1:"+port, nil))
}
