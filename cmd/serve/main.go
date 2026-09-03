package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/rnovatorov/radler/internal/core"
	"github.com/rnovatorov/radler/internal/grpc"
	"github.com/rnovatorov/radler/internal/grpc/api"
	"github.com/rnovatorov/radler/internal/xclip"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "%s: error: %s\n", os.Args[0], err)
		os.Exit(-1)
	}
}

func run(ctx context.Context) error {
	listen, clipboard := parseArgs()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	clip, err := newClipboard(clipboard)
	if err != nil {
		return fmt.Errorf("clipboard: %w", err)
	}
	clipboardAPI := api.NewClipboardServiceServer(core.NewClipboardService(clip))
	return grpc.Serve(ctx, listen, clipboardAPI)
}

func parseArgs() (listen string, clipboard string) {
	listenFlag := flag.String("listen", "", "listen address (tcp://host:port or unix:///path)")
	clipboardFlag := flag.String("clipboard", "xclip://xclip", "clipboard address (xclip://<binary> or xclip:///<abs-path>)")
	flag.Parse()
	return *listenFlag, *clipboardFlag
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
