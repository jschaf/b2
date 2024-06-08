package main

import (
	"fmt"

	"go.uber.org/zap/zapcore"

	"github.com/jschaf/b2/pkg/db"
	"github.com/jschaf/b2/pkg/log"
	"go.uber.org/zap"
)

func main() {
	if err := runMain(); err != nil {
		panic(err)
	}
}

func runMain() (err error) {
	l, err := log.NewShortDevLogger(zapcore.InfoLevel)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	sqlite := db.NewSQLiteStore(l)
	if err := sqlite.Open(); err != nil {
		return err
	}
	fetches, err := sqlite.AllRawFetches()
	if err != nil {
		return err
	}
	fmt.Printf("\nFetches: %v\n", fetches)
	defer func() {
		if cErr := sqlite.Close(); cErr != nil {
			l.Error("failed to close sqlite", zap.Error(cErr))
			if err != nil {
				err = cErr
			}
		}
	}()
	return nil
}
