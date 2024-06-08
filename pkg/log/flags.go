package log

import (
	"flag"
	"fmt"

	"go.uber.org/zap"
)

var (
	logLevelFlag = flag.String(
		"log", "info", `show logs at this level or higher (debug, info, warn, error, dpanic, panic, or fatal)`)
	useDevLoggerFlag = flag.Bool("use-dev-logger", false, "use prettier dev logger")
)

// ParseFlags creates a new logger based on flag values.
func ParseFlags() (zap.AtomicLevel, *zap.Logger, error) {
	lvl := zap.AtomicLevel{}
	if err := lvl.UnmarshalText([]byte(*logLevelFlag)); err != nil {
		return lvl, nil, fmt.Errorf("parse log flag value %q: %w", *logLevelFlag, err)
	}
	cfg := newProdConfig(lvl)
	if *useDevLoggerFlag {
		cfg = newShortDevCfg(lvl)
	}
	logger, err := cfg.Build()
	if err != nil {
		return lvl, nil, fmt.Errorf("build zap logger: %w", err)
	}
	// In case anyone uses the stdlib log package, redirect to Zap. None of our
	// code should use the stdlib lob package.
	zap.RedirectStdLog(logger)
	return lvl, logger, nil
}

func MustParseFlags() (zap.AtomicLevel, *zap.Logger) {
	lvl, logger, err := ParseFlags()
	if err != nil {
		panic("must parse flags: " + err.Error())
	}

	return lvl, logger
}
