package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/js"
	"github.com/jschaf/b2/pkg/log"
	"github.com/jschaf/b2/pkg/process"
)

func main() {
	process.RunMain(runMain)
}

func runMain(context.Context) error {
	fset := flag.CommandLine
	logLevel := log.DefineFlags(fset)
	if err := fset.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	slog.SetDefault(slog.New(log.NewDevHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})))

	slog.Info("start js compile")

	err := js.WriteTypeScriptMain(dirs.Dist)
	if err != nil {
		return fmt.Errorf("write typescript main: %w", err)
	}
	slog.Info("finish js compile")
	return nil
}
