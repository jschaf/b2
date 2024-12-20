package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/log"
	"github.com/jschaf/b2/pkg/markdown/compiler"
	"github.com/jschaf/b2/pkg/process"
)

var postGlobFlag = flag.String("glob", "", "if given, only compile files that match glob")

func compile(glob string) error {
	start := time.Now()
	globStr := *postGlobFlag
	if globStr == "" {
		globStr = "all"
	}
	slog.Info("start compile", slog.String("glob", globStr))
	c := compiler.NewDetailCompiler(dirs.Dist)
	if err := c.Compile(glob); err != nil {
		return fmt.Errorf("compile detail posts: %w", err)
	}
	slog.Info("finish compile", slog.Duration("duration", time.Since(start)))
	return nil
}

func main() {
	process.RunMain(runMain)
}

func runMain(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	fset := flag.CommandLine
	logLevel := log.DefineFlags(fset)
	if err := fset.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	slog.SetDefault(slog.New(log.NewDevHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})))

	if err := compile(*postGlobFlag); err != nil {
		return fmt.Errorf("compile: %w", err)
	}
	return nil
}
