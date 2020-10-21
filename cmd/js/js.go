package main

import (
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/js"
	"github.com/jschaf/b2/pkg/logs"
	"go.uber.org/zap/zapcore"
	"log"
)

func main() {
	l, err := logs.NewShortDevSugaredLogger(zapcore.DebugLevel)
	if err != nil {
		log.Fatalf("create dev logger: %s", err)
	}
	if err = js.WriteTypeScriptMain(dirs.PublicMemfs); err != nil {
		l.Fatal(err)
	}
}
