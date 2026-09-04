package xclip

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

func TestCopyWritesClipboard(t *testing.T) {
	a, err := New("xclip")
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	w, err := a.NewWriter(context.Background())
	if err != nil {
		t.Fatalf("new writer: %v", err)
	}
	if _, err := io.WriteString(w, "hello, "); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := io.WriteString(w, "clipboard"); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	r, err := a.NewReader(context.Background())
	if err != nil {
		t.Fatalf("new reader: %v", err)
	}
	data, readErr := io.ReadAll(r)
	closeErr := r.Close()
	if readErr != nil {
		t.Fatalf("read: %v", readErr)
	}
	if closeErr != nil {
		t.Fatalf("reader close: %v", closeErr)
	}
	if string(data) != "hello, clipboard" {
		t.Fatalf("paste = %q, want written data", data)
	}
}

func TestWriterCloseIsIdempotent(t *testing.T) {
	a, err := New("xclip")
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	w, err := a.NewWriter(context.Background())
	if err != nil {
		t.Fatalf("new writer: %v", err)
	}
	if _, err := io.WriteString(w, "idempotent"); err != nil {
		t.Fatalf("write: %v", err)
	}
	first := w.Close()
	second := w.Close()
	if first != nil {
		t.Fatalf("first close: %v", first)
	}
	if second != nil {
		t.Fatalf("second close: %v", second)
	}
}

func TestNewWriterMissingBinaryFails(t *testing.T) {
	_, err := newWriter(context.Background(), "/no/such/xclip-binary", os.Environ())
	if err == nil {
		t.Fatal("new writer with missing binary: expected error")
	}
	if !strings.Contains(err.Error(), "starting") {
		t.Fatalf("new writer error = %v, want start failure", err)
	}
}

func TestWriterCloseIncludesXclipStderr(t *testing.T) {
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
	w, err := newWriter(context.Background(), a.binary, env)
	if err != nil {
		t.Fatalf("new writer: %v", err)
	}
	closeErr := w.Close()
	if closeErr == nil {
		t.Fatal("close without display: expected error")
	}
	if !strings.Contains(closeErr.Error(), "Can't open display") {
		t.Fatalf("close error = %v, want xclip stderr included", closeErr)
	}
}

func TestWriterContextCancelFailsOnClose(t *testing.T) {
	a, err := New("xclip")
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	w, err := newWriter(ctx, a.binary, a.env)
	if err != nil {
		t.Fatalf("new writer: %v", err)
	}
	cancel()
	if err := w.Close(); err == nil {
		t.Fatal("close after cancel: expected error")
	}
}
