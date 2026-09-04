package xclip

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

type reader struct {
	cmd     *exec.Cmd
	stdoutR *os.File
	stderrR *os.File

	closeOnce sync.Once
	result    error
}

func newReader(ctx context.Context, binary string, env []string) (*reader, error) {
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		stdoutR.Close()
		stdoutW.Close()
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}
	cmd := exec.CommandContext(ctx, binary, "-o", "-selection", "clipboard")
	cmd.Env = env
	cmd.Stdout = stdoutW
	cmd.Stderr = stderrW
	r := &reader{
		cmd:     cmd,
		stdoutR: stdoutR,
		stderrR: stderrR,
	}
	if err := cmd.Start(); err != nil {
		stdoutR.Close()
		stdoutW.Close()
		stderrR.Close()
		stderrW.Close()
		return nil, fmt.Errorf("starting %s: %w", binary, err)
	}
	stdoutW.Close()
	stderrW.Close()
	return r, nil
}

func (r *reader) Read(b []byte) (int, error) {
	return r.stdoutR.Read(b)
}

func (r *reader) Close() error {
	r.closeOnce.Do(func() {
		r.result = r.close()
	})
	return r.result
}

func (r *reader) close() error {
	r.cmd.Cancel()
	defer r.stdoutR.Close()
	defer r.stderrR.Close()
	werr := r.cmd.Wait()
	if werr == nil {
		return nil
	}
	data, _ := io.ReadAll(r.stderrR)
	if len(data) == 0 {
		return werr
	}
	return fmt.Errorf("%w: %s", werr, data)
}
