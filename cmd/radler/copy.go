package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/rnovatorov/radler/internal/grpc"
	"github.com/rnovatorov/radler/internal/grpc/api"
)

func runCopy(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("copy", flag.ContinueOnError)
	listen := fs.String("listen", "", "server listen address (tcp://host:port or unix:///path)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	conn, err := grpc.Dial(*listen)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()
	if err := api.NewClipboardServiceClient(conn).Copy(ctx, os.Stdin); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	return nil
}
