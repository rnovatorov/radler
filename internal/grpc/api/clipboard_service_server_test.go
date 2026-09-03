package api

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/rnovatorov/radler/gen/go/api/v1"
	"github.com/rnovatorov/radler/internal/core"
)

type fakeClipboard struct {
	data     []byte
	closeErr error
}

func (f *fakeClipboard) NewReader(ctx context.Context) (io.ReadCloser, error) {
	return closeErrReader{Reader: bytes.NewReader(f.data), err: f.closeErr}, nil
}

func (f *fakeClipboard) NewWriter(ctx context.Context) (io.WriteCloser, error) {
	return nil, errors.New("not implemented")
}

type closeErrReader struct {
	io.Reader
	err error
}

func (r closeErrReader) Close() error { return r.err }

func startServer(t *testing.T, clip *fakeClipboard) apiv1.ClipboardServiceClient {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	apiv1.RegisterClipboardServiceServer(srv, NewClipboardServiceServer(core.NewClipboardService(clip)))
	go srv.Serve(lis)
	t.Cleanup(srv.Stop)
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
	return apiv1.NewClipboardServiceClient(conn)
}

func pasteAll(t *testing.T, client apiv1.ClipboardServiceClient) ([]byte, error) {
	t.Helper()
	stream, err := client.Paste(context.Background(), &apiv1.PasteRequest{})
	if err != nil {
		return nil, err
	}
	var got []byte
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			return got, nil
		}
		if err != nil {
			return got, err
		}
		got = append(got, resp.GetData()...)
	}
}

func TestPasteStreamsPayload(t *testing.T) {
	payload := bytes.Repeat([]byte("radler-"), 10000)
	client := startServer(t, &fakeClipboard{data: payload})
	got, err := pasteAll(t, client)
	if err != nil {
		t.Fatalf("paste: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("paste: got %d bytes, want %d", len(got), len(payload))
	}
}

func TestPasteCloseErrorSurfaces(t *testing.T) {
	client := startServer(t, &fakeClipboard{data: []byte("x"), closeErr: errors.New("clipboard exploded")})
	_, err := pasteAll(t, client)
	if err == nil {
		t.Fatal("paste: expected Close error to surface")
	}
	if !strings.Contains(err.Error(), "clipboard exploded") {
		t.Fatalf("paste error = %v, want Close error surfaced", err)
	}
}

func TestCopyNotImplemented(t *testing.T) {
	client := startServer(t, &fakeClipboard{})
	stream, err := client.Copy(context.Background())
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	if err := stream.Send(&apiv1.CopyRequest{Data: []byte("data")}); err != nil {
		t.Fatalf("send: %v", err)
	}
	_, err = stream.CloseAndRecv()
	if err == nil {
		t.Fatal("copy: expected unimplemented error")
	}
	if status.Code(err) != codes.Unimplemented {
		t.Fatalf("copy error code = %v, want Unimplemented", status.Code(err))
	}
}

func TestPasteContextErrorCode(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want codes.Code
	}{
		{context.Canceled, codes.Canceled},
		{context.DeadlineExceeded, codes.DeadlineExceeded},
	} {
		client := startServer(t, &fakeClipboard{data: []byte("x"), closeErr: tc.err})
		_, err := pasteAll(t, client)
		if err == nil {
			t.Fatalf("paste: expected error for %v", tc.err)
		}
		if status.Code(err) != tc.want {
			t.Fatalf("paste error code for %v = %v, want %v", tc.err, status.Code(err), tc.want)
		}
	}
}
