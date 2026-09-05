package grpc

import (
	"context"
	"net"
	"path/filepath"
	"testing"
	"time"
)

func TestListen(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "radler.sock")
	for _, tc := range []struct {
		addr    string
		network string
	}{
		{"tcp://127.0.0.1:0", "tcp"},
		{"unix://" + sock, "unix"},
	} {
		lis, err := listen(tc.addr)
		if err != nil {
			t.Fatalf("listen(%q): %v", tc.addr, err)
		}
		if got := lis.Addr().Network(); got != tc.network {
			t.Errorf("listen(%q) network = %q, want %q", tc.addr, got, tc.network)
		}
		lis.Close()
	}
}

func TestListenErrors(t *testing.T) {
	for _, addr := range []string{
		"",
		"://",
		"http://localhost:9090",
		"tcp://",
		"unix://",
		"unix://relative",
	} {
		if _, err := listen(addr); err == nil {
			t.Fatalf("listen(%q): expected error", addr)
		}
	}
}

func TestServeReleasesWatcherWhenListenerClosed(t *testing.T) {
	lis, err := net.Listen("unix", filepath.Join(t.TempDir(), "radler.sock"))
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	errCh := make(chan error, 1)
	go func() { errCh <- serve(context.Background(), lis) }()
	lis.Close()
	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("serve: expected error after listener close")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serve: did not return after listener close")
	}
}

func TestServeGracefulStopsOnContextCancel(t *testing.T) {
	lis, err := net.Listen("unix", filepath.Join(t.TempDir(), "radler.sock"))
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- serve(ctx, lis) }()
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("serve: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serve: did not stop after context cancel")
	}
}
