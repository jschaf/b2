package main

import (
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/js"
	"github.com/jschaf/b2/pkg/log"
	"go.uber.org/zap/zapcore"
)

func main() {
	l, err := log.NewShortDevSugaredLogger(zapcore.DebugLevel)
	if err != nil {
		panic("create dev logger: " + err.Error())
	}
	if err = js.WriteTypeScriptMain(dirs.PublicMemfs); err != nil {
		l.Fatal(err)
	}
}
