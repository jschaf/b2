package log

import (
	"flag"
	"log/slog"
)

func DefineFlags(fset *flag.FlagSet) slog.Level {
	var lvl slog.Level
	fset.TextVar(&lvl, "log-level", slog.LevelInfo, "show logs at this level or higher")
	return lvl
}
