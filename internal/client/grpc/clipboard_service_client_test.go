package grpc

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/rnovatorov/radler/gen/go/api/v1"
)

type fakeServer struct {
	apiv1.UnimplementedClipboardServiceServer

	mu      sync.Mutex
	data    []byte
	copyErr error
}

func (s *fakeServer) Copy(stream apiv1.ClipboardService_CopyServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&apiv1.CopyResponse{})
		}
		if err != nil {
			return err
		}
		s.mu.Lock()
		s.data = append(s.data, req.GetData()...)
		err = s.copyErr
		s.mu.Unlock()
		if err != nil {
			return err
		}
	}
}

func (s *fakeServer) Paste(_ *apiv1.PasteRequest, stream apiv1.ClipboardService_PasteServer) error {
	s.mu.Lock()
	data := s.data
	s.mu.Unlock()
	for len(data) > 0 {
		n := min(len(data), 32<<10)
		if err := stream.Send(&apiv1.PasteResponse{Data: data[:n]}); err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}

func (s *fakeServer) stored() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]byte(nil), s.data...)
}

func startConn(t *testing.T, srv *fakeServer) *grpc.ClientConn {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	gsrv := grpc.NewServer()
	apiv1.RegisterClipboardServiceServer(gsrv, srv)
	go gsrv.Serve(lis)
	t.Cleanup(gsrv.Stop)
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc client: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func TestCopyStreamsPayload(t *testing.T) {
	srv := &fakeServer{}
	client := NewClipboardServiceClient(startConn(t, srv))
	payload := bytes.Repeat([]byte("radler-"), 10000)
	if err := client.Copy(context.Background(), bytes.NewReader(payload)); err != nil {
		t.Fatalf("copy: %v", err)
	}
	if !bytes.Equal(srv.stored(), payload) {
		t.Fatalf("stored data: got %d bytes, want %d", len(srv.stored()), len(payload))
	}
}

func TestPasteStreamsPayload(t *testing.T) {
	payload := bytes.Repeat([]byte("radler-"), 10000)
	srv := &fakeServer{data: payload}
	client := NewClipboardServiceClient(startConn(t, srv))
	var got bytes.Buffer
	if err := client.Paste(context.Background(), &got); err != nil {
		t.Fatalf("paste: %v", err)
	}
	if !bytes.Equal(got.Bytes(), payload) {
		t.Fatalf("pasted data: got %d bytes, want %d", got.Len(), len(payload))
	}
}

func TestCopyPasteRoundTrip(t *testing.T) {
	srv := &fakeServer{}
	client := NewClipboardServiceClient(startConn(t, srv))
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

type errReader struct{ err error }

func (r errReader) Read([]byte) (int, error) { return 0, r.err }

type errWriter struct{ err error }

func (w errWriter) Write([]byte) (int, error) { return 0, w.err }

func TestCopyReaderError(t *testing.T) {
	srv := &fakeServer{}
	client := NewClipboardServiceClient(startConn(t, srv))
	want := errors.New("stdin exploded")
	r := io.MultiReader(strings.NewReader("partial"), errReader{want})
	if err := client.Copy(context.Background(), r); !errors.Is(err, want) {
		t.Fatalf("copy error = %v, want %v", err, want)
	}
}

func TestPasteWriterError(t *testing.T) {
	srv := &fakeServer{data: bytes.Repeat([]byte("x"), 2<<20)}
	client := NewClipboardServiceClient(startConn(t, srv))
	want := errors.New("stdout exploded")
	if err := client.Paste(context.Background(), errWriter{want}); !errors.Is(err, want) {
		t.Fatalf("paste error = %v, want %v", err, want)
	}
}

func TestCopySurfacesServerError(t *testing.T) {
	srv := &fakeServer{copyErr: status.Error(codes.Internal, "clipboard exploded")}
	client := NewClipboardServiceClient(startConn(t, srv))
	payload := bytes.Repeat([]byte("x"), 2<<20)
	err := client.Copy(context.Background(), bytes.NewReader(payload))
	if err == nil {
		t.Fatal("copy: expected error")
	}
	if status.Code(err) != codes.Internal {
		t.Fatalf("copy error = %v, want code %v", err, codes.Internal)
	}
	if !strings.Contains(err.Error(), "clipboard exploded") {
		t.Fatalf("copy error = %v, want containing %q", err, "clipboard exploded")
	}
}
