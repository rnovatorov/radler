package main

import (
	"context"
	"flag"

	"github.com/rnovatorov/radler/internal/server"
)

func runServe(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	listen := fs.String("listen", "", "listen address (tcp://host:port or unix:///path)")
	clipboard := fs.String("clipboard", "xclip://xclip", "clipboard address (xclip://<binary> or xclip:///<abs-path>)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	srv, err := server.New(server.Options{Listen: *listen, Clipboard: *clipboard})
	if err != nil {
		return err
	}
	return srv.Serve(ctx)
}
