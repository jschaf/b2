package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/jschaf/b2/pkg/livereload"
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

	distDirHandler := http.FileServer(http.Dir(opts.distDir))

	lrScript := strings.Join([]string{
		fmt.Sprintf("<script defer src=%s?port=%s&path=%s type='application/javascript'>",
			lrJSPath, port, strings.TrimLeft(lrPath, "/")),
		"</script>",
	}, "")
	mux.Handle("/", opts.lr.NewHTMLInjector(lrScript, distDirHandler))
	return mux
}
