package grpc

import (
	"context"
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
	srv := grpc.NewServer()
	for _, service := range services {
		service.Register(srv)
	}
	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()
	return srv.Serve(lis)
}

func listen(addr string) (net.Listener, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("grpc listen: invalid address %q: %v", addr, err)
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
