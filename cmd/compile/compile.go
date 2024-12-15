package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/markdown/compiler"
)

var postGlobFlag = flag.String("glob", "", "if given, only compile files that match glob")

func compile(glob string) error {
	slog.Info("run compile cmd", "glob", glob)
	c := compiler.NewPostDetail(dirs.Public)
	if err := c.CompileAll(glob); err != nil {
		return fmt.Errorf("compile detail posts: %w", err)
	}
	return nil
}

func main() {
	flag.Parse()
	if err := compile(*postGlobFlag); err != nil {
		slog.Error("compile cmd", "error", err.Error())
		os.Exit(1)
	}
	slog.Info("done")
}
