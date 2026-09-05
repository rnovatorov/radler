package server

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rnovatorov/radler/internal/server/core"
	"github.com/rnovatorov/radler/internal/server/grpc"
	"github.com/rnovatorov/radler/internal/server/xclip"
)

type Options struct {
	Listen    string
	Clipboard string
}

type Server struct {
	listen  string
	service grpc.Service
}

func New(opts Options) (*Server, error) {
	clip, err := newClipboard(opts.Clipboard)
	if err != nil {
		return nil, fmt.Errorf("clipboard: %w", err)
	}
	return &Server{
		listen:  opts.Listen,
		service: grpc.NewClipboardServiceServer(core.NewClipboardService(clip)),
	}, nil
}

func (s *Server) Serve(ctx context.Context) error {
	return grpc.Serve(ctx, s.listen, s.service)
}

func newClipboard(addr string) (core.Clipboard, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid clipboard address %q: %w", addr, err)
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
