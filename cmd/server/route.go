package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/jschaf/jsc/pkg/livereload"
)

type buildRoutesOpts struct {
	distDir string
	lr      *livereload.LiveReload
}

func buildRoutes(opts buildRoutesOpts) *http.ServeMux {
	mux := http.NewServeMux()
	lrJSPath := "/dev/livereload.js"
	lrPath := "/dev/livereload"

	mux.HandleFunc(lrJSPath, opts.lr.ServeJSHandler)
	mux.HandleFunc(lrPath, opts.lr.WebSocketHandler)
	mux.HandleFunc("GET /_/heap/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // do nothing
	})

	root := http.Dir(opts.distDir)
	distDirHandler := &cleanFileServer{
		root:        root,
		baseHandler: http.FileServer(root),
	}

	lrScript := strings.Join([]string{
		fmt.Sprintf("<script defer src=%s?port=%s&path=%s type='application/javascript'>",
			lrJSPath, port, strings.TrimLeft(lrPath, "/")),
		"</script>",
	}, "")
	mux.Handle("/", opts.lr.NewHTMLInjector(lrScript, distDirHandler))
	return mux
}

// cleanFileServer is an http.FileServer that serves directories with an
// index.html without the trailing slash to match the behavior of Firebase.
type cleanFileServer struct {
	root        http.FileSystem
	baseHandler http.Handler
}

func (c *cleanFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If the path has a dot, it's a file, serve it.
	if strings.Contains(r.URL.Path, ".") {
		c.baseHandler.ServeHTTP(w, r)
		return
	}

	// Redirect .../index.html to .../
	// can't use Redirect() because that would make the path absolute,
	// which would be a problem running under StripPrefix
	const indexPage = "/index.html"
	if strings.HasSuffix(r.URL.Path, indexPage) {
		localRedirect(w, r, ".")
		return
	}

	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}
	name := path.Clean(upath)

	f, err := c.root.Open(name)
	if err != nil {
		http.Error(w, fmt.Errorf("open file: %w", err).Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			slog.Error("close file", "err", err.Error())
		}
	}()

	d, err := f.Stat()
	if err != nil {
		http.Error(w, fmt.Errorf("stat file: %w", err).Error(), http.StatusInternalServerError)
		return
	}
	if !d.IsDir() {
		c.baseHandler.ServeHTTP(w, r)
		return
	}

	// If it has a trailing slash, redirect to the same path without the slash.
	if upath != "/" && upath[len(upath)-1] == '/' {
		w.Header().Set("Location", upath[:len(upath)-1])
		w.WriteHeader(http.StatusMovedPermanently)
		return
	}

	// Find the index.html page.
	index, err := c.root.Open(path.Join(name, indexPage))
	if err != nil {
		http.Error(w, fmt.Errorf("open index file: %w", err).Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := index.Close(); err != nil {
			slog.Error("close index file", "err", err.Error())
		}
	}()

	// Serve the index.html page.
	body, err := io.ReadAll(index)
	if err != nil {
		http.Error(w, fmt.Errorf("read index file: %w", err).Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.FormatInt(int64(len(body)), 10))
	_, err = w.Write(body)
	if err != nil {
		slog.Error("write index file", "err", err.Error())
		return
	}
}

// localRedirect gives a Moved Permanently response.
// It does not convert relative paths to absolute paths like Redirect does.
func localRedirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	w.WriteHeader(http.StatusMovedPermanently)
}
