package main

import (
	"context"
	"flag"
	"os"

	"github.com/rnovatorov/radler/internal/client"
)

func runPaste(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("paste", flag.ContinueOnError)
	listen := fs.String("listen", "", "server listen address (tcp://host:port or unix:///path)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	c, err := client.New(*listen)
	if err != nil {
		return err
	}
	defer c.Close()
	return c.Paste(ctx, os.Stdout)
}
