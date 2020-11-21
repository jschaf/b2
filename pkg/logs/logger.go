package logs

import (
	"log"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewShortDevLogger(lvl zapcore.Level) (*zap.Logger, error) {
	logCfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(lvl),
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
		OutputPaths:      []string{"stderr", "/tmp/b2_server.log"},
		ErrorOutputPaths: []string{"stderr", "/tmp/b2_server.log"},
	}
	return logCfg.Build()
}

func NewShortDevSugaredLogger(lvl zapcore.Level) (*zap.SugaredLogger, error) {
	l, err := NewShortDevLogger(lvl)
	if err != nil {
		return nil, err
	}
	return l.Sugar(), nil
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
		if err, ok := err.(*os.PathError); ok && err.Path == "/dev/stderr" {
			return // ignore
		}
		log.Printf("ERROR: failed to sync zap logger: %s", err.Error())
	}
}
