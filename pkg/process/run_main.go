package process

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func RunMain(f func(ctx context.Context) error) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	err := f(ctx)
	cancel() // os.Exit doesn't call the defer statement
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
