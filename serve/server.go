package main

import (
	"github.com/jschaf/b2/serve/livereload"
	"github.com/jschaf/b2/serve/paths"
	"log"
	"net/http"
	"path/filepath"
)

func main() {
	root, err := paths.FindRootDir()
	if err != nil {
		log.Fatalf("failed to find root dir: %s", err)
	}
	pubDir := filepath.Join(root, "public")
	log.Printf("Serving dir %s", pubDir)
	pubDirHandler := http.FileServer(http.Dir(pubDir))
	// http.Handle("/", pubDirHandler)

	liveReload := livereload.New()
	http.HandleFunc("/dev/livereload.js", liveReload.ServeJSHandler)
	http.HandleFunc("/dev/livereload", liveReload.WebSocketHandler)

	lrScript := "<script src=foo></script>"
	http.Handle("/", livereload.NewHTMLInjector(lrScript, pubDirHandler))

	http.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	port := "8080"
	log.Printf("Serving at port %s", port)
	log.Fatal(http.ListenAndServe("127.0.0.1:"+port, nil))
}
