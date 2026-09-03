package xclip

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"testing"
)

func TestBareBinaryNameResolves(t *testing.T) {
	a, err := New("xclip")
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	r, err := a.NewReader(context.Background())
	if err != nil {
		t.Fatalf("new reader: %v", err)
	}
	_, readErr := io.ReadAll(r)
	closeErr := r.Close()
	if readErr != nil {
		t.Fatalf("read: %v", readErr)
	}
	if closeErr == nil || !strings.Contains(closeErr.Error(), "STRING not available") {
		t.Fatalf("close error = %v, want proof the bare-named binary ran", closeErr)
	}
}

func TestPasteEmptySelectionFails(t *testing.T) {
	a, err := New("xclip")
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	r, err := a.NewReader(context.Background())
	if err != nil {
		t.Fatalf("paste: %v", err)
	}
	defer r.Close()
	data, readErr := io.ReadAll(r)
	closeErr := r.Close()
	if len(data) != 0 {
		t.Fatalf("paste on empty selection returned data: %q", data)
	}
	if readErr != nil {
		t.Fatalf("read: %v", readErr)
	}
	if closeErr == nil {
		t.Fatal("paste on empty selection: Close must report the failure")
	}
	if !strings.Contains(closeErr.Error(), "STRING not available") {
		t.Fatalf("close error = %v, want xclip unavailable-selection stderr", closeErr)
	}
}

func TestNewResolvesBinary(t *testing.T) {
	for _, binary := range []string{"xclip", "/usr/bin/xclip"} {
		a, err := New(binary)
		if err != nil {
			t.Fatalf("new(%q): %v", binary, err)
		}
		if !filepath.IsAbs(a.binary) {
			t.Fatalf("new(%q): binary = %q, want absolute", binary, a.binary)
		}
	}
	if _, err := New("no-such-clipboard-tool"); err == nil {
		t.Fatal("new(missing): expected error")
	}
}
