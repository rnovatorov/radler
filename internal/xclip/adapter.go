package xclip

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/rnovatorov/radler/internal/core"
)

var _ core.Clipboard = (*Adapter)(nil)

type Adapter struct {
	binary string
	env    []string
}

func New(binary string) (*Adapter, error) {
	resolved, err := exec.LookPath(binary)
	if err != nil {
		return nil, fmt.Errorf("resolving %s: %w", binary, err)
	}
	return &Adapter{
		binary: resolved,
		env:    os.Environ(),
	}, nil
}

func (a *Adapter) NewReader(ctx context.Context) (io.ReadCloser, error) {
	return newReader(ctx, a.binary, a.env)
}

func (a *Adapter) NewWriter(ctx context.Context) (io.WriteCloser, error) {
	return newWriter(ctx, a.binary, a.env)
}
