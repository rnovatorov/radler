package server

import (
	"strings"
	"testing"
)

func TestNewClipboardErrors(t *testing.T) {
	for _, tc := range []struct {
		addr string
		want string
	}{
		{"bogus://xclip", "not implemented"},
		{"xclip://", "not implemented"},
		{"://", "invalid clipboard address"},
	} {
		_, err := New(Options{Clipboard: tc.addr})
		if err == nil {
			t.Fatalf("New(clipboard %q): expected error", tc.addr)
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("New(clipboard %q) error = %v, want containing %q", tc.addr, err, tc.want)
		}
	}
}
