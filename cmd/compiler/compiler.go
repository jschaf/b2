package main

import (
	"flag"
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/logs"
	"github.com/jschaf/b2/pkg/sites"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
)

var profileFlag = flag.String("cpu-profile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	logger, err := logs.NewShortDevSugaredLogger(zapcore.InfoLevel)
	if err != nil {
		log.Fatalf("create dev logger: %s", err)
	}
	if *profileFlag != "" {
		f, err := os.Create(*profileFlag)
		if err != nil {
			log.Fatalf("create profile file: %s", err)
		}
		log.Println("created profile file: " + f.Name())
		if err = pprof.StartCPUProfile(f); err != nil {
			logger.Errorf("start CPU profile: %s", err.Error())
		}
		defer pprof.StopCPUProfile()
	}
	pubDir := filepath.Join(git.MustFindRootDir(), dirs.Public)
	if err := sites.Rebuild(pubDir, logger.Desugar()); err != nil {
		logger.Fatal(err)
	}
}
