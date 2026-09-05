package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"

	"github.com/rnovatorov/radler/internal/core"
	"github.com/rnovatorov/radler/internal/grpc"
	"github.com/rnovatorov/radler/internal/grpc/api"
	"github.com/rnovatorov/radler/internal/xclip"
)

func runServe(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	listen := fs.String("listen", "", "listen address (tcp://host:port or unix:///path)")
	clipboard := fs.String("clipboard", "xclip://xclip", "clipboard address (xclip://<binary> or xclip:///<abs-path>)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	clip, err := newClipboard(*clipboard)
	if err != nil {
		return fmt.Errorf("clipboard: %w", err)
	}
	clipboardAPI := api.NewClipboardServiceServer(core.NewClipboardService(clip))
	return grpc.Serve(ctx, *listen, clipboardAPI)
}

func newClipboard(addr string) (core.Clipboard, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid clipboard address %q: %v", addr, err)
	}
	name := u.Path
	if name == "" {
		name = u.Host
	}
	if u.Scheme != "xclip" || name == "" {
		return nil, fmt.Errorf("%q: not implemented", addr)
	}
	return xclip.New(name)
}
