package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/jschaf/jsc/pkg/dirs"
	"github.com/jschaf/jsc/pkg/log"
	"github.com/jschaf/jsc/pkg/process"
	"github.com/jschaf/jsc/pkg/sites"
)

var profileFlag = flag.String("cpu-profile", "", "write cpu profile to file")

func main() {
	process.RunMain(runMain)
}

func runMain(_ context.Context) error {
	fset := flag.CommandLine
	logLevel := log.DefineFlags(fset)
	if err := fset.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	slog.SetDefault(slog.New(log.NewDevHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})))

	runtime.GOMAXPROCS(1)
	if *profileFlag != "" {
		f, err := os.Create(*profileFlag)
		if err != nil {
			slog.Error("create profile file", "error", err)
			return err
		}
		slog.Info("created profile file", "file", f.Name())
		if err = pprof.StartCPUProfile(f); err != nil {
			slog.Error("start CPU profile", "error", err.Error())
		}
		defer pprof.StopCPUProfile()
	}

	distDir := dirs.Dist
	if err := sites.Rebuild(distDir); err != nil {
		slog.Error("rebuild site", "error", err)
		return err
	}
	return nil
}
