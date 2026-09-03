package core

import (
	"context"
	"io"
)

type Clipboard interface {
	NewReader(ctx context.Context) (io.ReadCloser, error)
	NewWriter(ctx context.Context) (io.WriteCloser, error)
}
