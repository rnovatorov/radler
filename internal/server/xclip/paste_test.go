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
	if closeErr != nil {
		t.Fatalf("close: %v", closeErr)
	}
}

func TestPasteEmptySelectionSucceeds(t *testing.T) {
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
	if closeErr != nil {
		t.Fatalf("paste on empty selection: Close error = %v, want nil", closeErr)
	}
}

func TestReaderCloseIncludesXclipStderr(t *testing.T) {
	a, err := New("xclip")
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	env := make([]string, 0, len(a.env))
	for _, e := range a.env {
		if !strings.HasPrefix(e, "DISPLAY=") {
			env = append(env, e)
		}
	}
	r, err := newReader(context.Background(), a.binary, env)
	if err != nil {
		t.Fatalf("new reader: %v", err)
	}
	_, _ = io.ReadAll(r)
	closeErr := r.Close()
	if closeErr == nil {
		t.Fatal("close without display: expected error")
	}
	if !strings.Contains(closeErr.Error(), "Can't open display") {
		t.Fatalf("close error = %v, want xclip stderr included", closeErr)
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
