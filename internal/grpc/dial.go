package grpc

import (
	"fmt"
	"net/url"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Dial(addr string) (*grpc.ClientConn, error) {
	target, err := dialTarget(addr)
	if err != nil {
		return nil, err
	}
	return grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func dialTarget(addr string) (string, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return "", fmt.Errorf("grpc dial: invalid address %q: %v", addr, err)
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
