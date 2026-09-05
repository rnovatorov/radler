package client

import (
	"context"
	"fmt"
	"io"

	"github.com/rnovatorov/radler/internal/client/grpc"
)

type Client struct {
	conn      *grpc.Conn
	clipboard *grpc.ClipboardServiceClient
}

func New(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	return &Client{conn: conn, clipboard: grpc.NewClipboardServiceClient(conn)}, nil
}

func (c *Client) Copy(ctx context.Context, r io.Reader) error {
	if err := c.clipboard.Copy(ctx, r); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	return nil
}

func (c *Client) Paste(ctx context.Context, w io.Writer) error {
	if err := c.clipboard.Paste(ctx, w); err != nil {
		return fmt.Errorf("paste: %w", err)
	}
	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
