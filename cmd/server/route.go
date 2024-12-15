package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/jschaf/b2/pkg/livereload"
)

type buildRoutesOpts struct {
	pubDir string
	lr     *livereload.LiveReload
}

func buildRoutes(opts buildRoutesOpts) *http.ServeMux {
	mux := http.NewServeMux()
	lrJSPath := "/dev/livereload.js"
	lrPath := "/dev/livereload"

	mux.HandleFunc(lrJSPath, opts.lr.ServeJSHandler)
	mux.HandleFunc(lrPath, opts.lr.WebSocketHandler)

	pubDirHandler := http.FileServer(http.Dir(opts.pubDir))

	lrScript := strings.Join([]string{
		fmt.Sprintf("<script defer src=%s?port=%s&path=%s type='application/javascript'>",
			lrJSPath, port, strings.TrimLeft(lrPath, "/")),
		"</script>",
	}, "")
	mux.Handle("/", opts.lr.NewHTMLInjector(lrScript, pubDirHandler))
	return mux
}
