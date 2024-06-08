package main

import (
	"flag"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/log"
	"github.com/jschaf/b2/pkg/sites"
)

var profileFlag = flag.String("cpu-profile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	_, logger := log.MustParseFlags()
	l := logger.Sugar()
	runtime.GOMAXPROCS(1)
	if *profileFlag != "" {
		f, err := os.Create(*profileFlag)
		if err != nil {
			l.Fatalf("create profile file: %s", err)
		}
		l.Info("created profile file: " + f.Name())
		if err = pprof.StartCPUProfile(f); err != nil {
			l.Errorf("start CPU profile: %s", err.Error())
		}
		defer pprof.StopCPUProfile()
	}

	pubDir := dirs.PublicMemfs
	if err := sites.Rebuild(pubDir, logger); err != nil {
		l.Fatal(err)
	}
	l.Info("done")
}
