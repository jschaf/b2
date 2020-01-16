package main

import (
	"github.com/jschaf/b2/serve/paths"
	"log"
	"net/http"
	"path/filepath"
)

type injectLiveReloadJHandler struct {
	prev http.Handler
}

func newInjectLiveReloadHandler(prev http.Handler) *injectLiveReloadJHandler {
	return &injectLiveReloadJHandler{prev: prev}
}

func (ilr *injectLiveReloadJHandler) Header() http.Header {
	panic("implement me")
}

func (ilr *injectLiveReloadJHandler) Write([]byte) (int, error) {
	panic("implement me")
}

func (ilr *injectLiveReloadJHandler) WriteHeader(statusCode int) {
	panic("implement me")
}

func injectLivereloadScriptHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
func main() {
	root, err := paths.FindRootDir()
	if err != nil {
		log.Fatalf("failed to find root dir: %s", err)
	}
	pubDir := filepath.Join(root, "public")
	log.Printf("Serving dir %s", pubDir)
	pubDirHandler := http.FileServer(http.Dir(pubDir))
	http.Handle("/", pubDirHandler)

	http.HandleFunc("/static/livereload.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(root, "third_party", "livereload", "livereload.js"))
	})

	http.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	port := "8080"
	log.Printf("Serving at port %s", port)
	log.Fatal(http.ListenAndServe("127.0.0.1:"+port, nil))
}
