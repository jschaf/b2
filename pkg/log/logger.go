package log

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewProductionConfig is a reasonable production logging configuration.
// Logging is enabled at InfoLevel and above.
//
// It uses a JSON encoder, writes to standard error, and enables sampling.
// Stacktraces are automatically included on logs of ErrorLevel and above.
func newProdConfig(lvl zap.AtomicLevel) zap.Config {
	return zap.Config{
		Level:       lvl,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

func newShortDevCfg(lvl zap.AtomicLevel) zap.Config {
	return zap.Config{
		Level:       lvl,
		Development: true,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			// Keys can be anything except the empty string.
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeTime:     hourMinSecEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

func hourMinSecEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}
	layout := "15:04:05.000"

	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, layout)
		return
	}
	enc.AppendString(t.Format(layout))
}

// Flush flushes at any buffered logs.
func Flush(l *zap.Logger) {
	if err := l.Sync(); err != nil {
		// Ignore harmless errors when trying to flush stderr.
		// https://github.com/uber-go/zap/issues/328
		if err, ok := err.(*os.PathError); ok && err.Path == "/dev/stderr" {
			return // ignore
		}
		fmt.Printf("ERROR: failed to sync zap logger: %s\n", err.Error())
	}
}

func NewShortDevLogger(lvl zapcore.Level) (*zap.Logger, error) {
	logCfg := newShortDevCfg(zap.NewAtomicLevelAt(lvl))
	return logCfg.Build()
}

func NewShortDevSugaredLogger(lvl zapcore.Level) (*zap.SugaredLogger, error) {
	l, err := NewShortDevLogger(lvl)
	if err != nil {
		return nil, err
	}
	return l.Sugar(), nil
}
