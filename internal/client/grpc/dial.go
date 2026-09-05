package grpc

import (
	"fmt"
	"net/url"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Conn struct {
	*grpc.ClientConn
}

func Dial(addr string) (*Conn, error) {
	target, err := dialTarget(addr)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Conn{ClientConn: conn}, nil
}

func dialTarget(addr string) (string, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return "", fmt.Errorf("grpc dial: invalid address %q: %w", addr, err)
	}
	switch u.Scheme {
	case "tcp":
		if u.Host == "" {
			return "", fmt.Errorf("grpc dial: invalid address %q: must be tcp://host:port or unix:///path", addr)
		}
		return "passthrough:///" + u.Host, nil
	case "unix":
		if u.Path == "" {
			return "", fmt.Errorf("grpc dial: invalid address %q: must be tcp://host:port or unix:///path", addr)
		}
		return "unix://" + u.Path, nil
	default:
		return "", fmt.Errorf("grpc dial: invalid address %q: must be tcp://host:port or unix:///path", addr)
	}
}
