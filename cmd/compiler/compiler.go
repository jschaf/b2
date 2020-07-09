package main

import (
	"flag"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/jschaf/b2/pkg/logs"
	"github.com/jschaf/b2/pkg/markdown/compiler"
	"github.com/jschaf/b2/pkg/static"
)

var flagGlob = flag.String(
	"glob", "",
	"Only compile posts with an exact substring match on the filename")

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
	start := time.Now()
	c := compiler.NewForPostDetail(logger.Desugar())
	if err := c.CompileAllPosts(*flagGlob); err != nil {
		log.Fatal(err)
	}

	ic := compiler.NewForIndex(logger.Desugar())
	if err := ic.Compile(); err != nil {
		log.Fatal(err)
	}

	if err := static.CopyStaticFiles(); err != nil {
		log.Fatal(err)
	}

	if err := static.LinkPapers(); err != nil {
		log.Fatal(err)
	}
	duration := time.Since(start)
	logger.Infof("finished compiling in %d ms", duration.Milliseconds())
}
