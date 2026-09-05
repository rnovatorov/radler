package api

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

type errReader struct{ err error }

func (r errReader) Read([]byte) (int, error) { return 0, r.err }

type errWriter struct{ err error }

func (w errWriter) Write([]byte) (int, error) { return 0, w.err }

func TestClientCopyStreamsPayload(t *testing.T) {
	clip := &fakeClipboard{}
	client := NewClipboardServiceClient(startConn(t, clip))
	payload := bytes.Repeat([]byte("radler-"), 10000)
	if err := client.Copy(context.Background(), bytes.NewReader(payload)); err != nil {
		t.Fatalf("copy: %v", err)
	}
	if !bytes.Equal(clip.data, payload) {
		t.Fatalf("committed data: got %d bytes, want %d", len(clip.data), len(payload))
	}
}

func TestClientPasteStreamsPayload(t *testing.T) {
	payload := bytes.Repeat([]byte("radler-"), 10000)
	client := NewClipboardServiceClient(startConn(t, &fakeClipboard{data: payload}))
	var got bytes.Buffer
	if err := client.Paste(context.Background(), &got); err != nil {
		t.Fatalf("paste: %v", err)
	}
	if !bytes.Equal(got.Bytes(), payload) {
		t.Fatalf("pasted data: got %d bytes, want %d", got.Len(), len(payload))
	}
}

func TestClientCopyReaderError(t *testing.T) {
	client := NewClipboardServiceClient(startConn(t, &fakeClipboard{}))
	want := errors.New("stdin exploded")
	r := io.MultiReader(strings.NewReader("partial"), errReader{want})
	if err := client.Copy(context.Background(), r); !errors.Is(err, want) {
		t.Fatalf("copy error = %v, want %v", err, want)
	}
}

func TestClientPasteWriterError(t *testing.T) {
	client := NewClipboardServiceClient(startConn(t, &fakeClipboard{data: []byte("payload")}))
	want := errors.New("stdout exploded")
	if err := client.Paste(context.Background(), errWriter{want}); !errors.Is(err, want) {
		t.Fatalf("paste error = %v, want %v", err, want)
	}
}

func TestClientCopyPasteRoundTrip(t *testing.T) {
	clip := &fakeClipboard{}
	conn := startConn(t, clip)
	client := NewClipboardServiceClient(conn)
	payload := []byte("hello radler")
	if err := client.Copy(context.Background(), bytes.NewReader(payload)); err != nil {
		t.Fatalf("copy: %v", err)
	}
	var got bytes.Buffer
	if err := client.Paste(context.Background(), &got); err != nil {
		t.Fatalf("paste: %v", err)
	}
	if !bytes.Equal(got.Bytes(), payload) {
		t.Fatalf("pasted data = %q, want %q", got.Bytes(), payload)
	}
}
