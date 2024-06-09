package main

import (
	"log/slog"
	"os"

	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/js"
)

func main() {
	if err := js.WriteTypeScriptMain(dirs.PublicMemfs); err != nil {
		slog.Error("write typescript main", "error", err.Error())
		os.Exit(1)
	}
}
