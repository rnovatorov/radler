package grpc

import (
	"bytes"
	"context"
	"io"

	"google.golang.org/grpc"

	"github.com/rnovatorov/radler/gen/go/api/v1"
)

type ClipboardServiceClient struct {
	client apiv1.ClipboardServiceClient
}

func NewClipboardServiceClient(cc grpc.ClientConnInterface) *ClipboardServiceClient {
	return &ClipboardServiceClient{client: apiv1.NewClipboardServiceClient(cc)}
}

func (c *ClipboardServiceClient) Copy(ctx context.Context, r io.Reader) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stream, err := c.client.Copy(ctx)
	if err != nil {
		return err
	}
	buf := make([]byte, 32<<10)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if err := stream.Send(&apiv1.CopyRequest{Data: bytes.Clone(buf[:n])}); err != nil {
				if err == io.EOF {
					_, err = stream.CloseAndRecv()
				}
				return err
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	_, err = stream.CloseAndRecv()
	return err
}

func (c *ClipboardServiceClient) Paste(ctx context.Context, w io.Writer) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stream, err := c.client.Paste(ctx, &apiv1.PasteRequest{})
	if err != nil {
		return err
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if _, err := w.Write(resp.GetData()); err != nil {
			return err
		}
	}
}
