package main

import (
	"fmt"
	"log/slog"

	"github.com/jschaf/b2/pkg/db"
)

func main() {
	if err := runMain(); err != nil {
		panic(err)
	}
}

func runMain() (err error) {
	sqlite := db.NewSQLiteStore()
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
			slog.Error("close sqlite", "error", cErr)
			if err != nil {
				err = cErr
			}
		}
	}()
	return nil
}
