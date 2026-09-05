package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(context.Background()); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "%s: error: %s\n", os.Args[0], err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: %s <serve|copy|paste> [flags]", os.Args[0])
	}
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	args := os.Args[2:]
	switch os.Args[1] {
	case "serve":
		return runServe(ctx, args)
	case "copy":
		return runCopy(ctx, args)
	case "paste":
		return runPaste(ctx, args)
	default:
		return fmt.Errorf("unknown command %q: expected serve, copy or paste", os.Args[1])
	}
}
