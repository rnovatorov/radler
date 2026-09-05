package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"

	"google.golang.org/grpc"
)

type Service interface {
	Register(grpc.ServiceRegistrar)
}

func Serve(ctx context.Context, addr string, services ...Service) error {
	lis, err := listen(addr)
	if err != nil {
		return err
	}
	return serve(ctx, lis, services...)
}

func serve(ctx context.Context, lis net.Listener, services ...Service) error {
	srv := grpc.NewServer()
	for _, service := range services {
		service.Register(srv)
	}
	errc := make(chan error, 1)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
		errc <- srv.Serve(lis)
	}()
	<-ctx.Done()
	srv.GracefulStop()
	if err := <-errc; err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return err
	}
	return nil
}

func listen(addr string) (net.Listener, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("grpc listen: invalid address %q: %w", addr, err)
	}
	var network, address string
	switch u.Scheme {
	case "tcp":
		network, address = "tcp", u.Host
	case "unix":
		network, address = "unix", u.Path
	}
	if address == "" {
		return nil, fmt.Errorf("grpc listen: invalid address %q: must be tcp://host:port or unix:///path", addr)
	}
	return net.Listen(network, address)
}
