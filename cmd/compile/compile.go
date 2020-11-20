package main

import (
	"flag"
	"fmt"
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/logs"
	"github.com/jschaf/b2/pkg/markdown/compiler"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
)

var postGlobFlag = flag.String("glob", "", "if given, only compile files that match glob")

func compile(glob string, l *zap.Logger) error {
	l.Sugar().Infof("Run compile cmd with glob %q", glob)
	c := compiler.NewPostDetail(dirs.PublicMemfs, l)
	if err := c.CompileAll(glob); err != nil {
		return fmt.Errorf("compile detail posts: %w", err)
	}
	return nil
}

func main() {
	flag.Parse()
	logger, err := logs.NewShortDevSugaredLogger(zapcore.DebugLevel)
	if err != nil {
		log.Fatalf("create dev logger: %s", err)
	}
	if err := compile(*postGlobFlag, logger.Desugar()); err != nil {
		logger.Fatalf("compile cmd: %s", err)
	}
	logger.Info("done")
}
