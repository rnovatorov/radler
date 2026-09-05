package grpc

import "testing"

func TestDialTarget(t *testing.T) {
	for _, tc := range []struct {
		addr string
		want string
	}{
		{"tcp://localhost:9090", "passthrough:///localhost:9090"},
		{"tcp://127.0.0.1:9090", "passthrough:///127.0.0.1:9090"},
		{"tcp://[::1]:9090", "passthrough:///[::1]:9090"},
		{"tcp://:9090", "passthrough:///:9090"},
		{"unix:///tmp/radler.sock", "unix:///tmp/radler.sock"},
	} {
		got, err := dialTarget(tc.addr)
		if err != nil {
			t.Fatalf("dialTarget(%q): %v", tc.addr, err)
		}
		if got != tc.want {
			t.Fatalf("dialTarget(%q) = %q, want %q", tc.addr, got, tc.want)
		}
	}
}

func TestDialTargetErrors(t *testing.T) {
	for _, addr := range []string{
		"",
		"://",
		"http://localhost:9090",
		"tcp://",
		"unix://",
		"unix://relative",
	} {
		if _, err := dialTarget(addr); err == nil {
			t.Fatalf("dialTarget(%q): expected error", addr)
		}
	}
}
