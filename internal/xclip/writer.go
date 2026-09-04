package xclip

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

type writer struct {
	cmd     *exec.Cmd
	stdinW  *os.File
	stderrR *os.File

	closeOnce sync.Once
	err       error
}

func newWriter(ctx context.Context, binary string, env []string) (*writer, error) {
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		stdinR.Close()
		stdinW.Close()
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}
	cmd := exec.CommandContext(ctx, binary, "-i", "-selection", "clipboard")
	cmd.Env = env
	cmd.Stdin = stdinR
	cmd.Stderr = stderrW
	w := &writer{
		cmd:     cmd,
		stdinW:  stdinW,
		stderrR: stderrR,
	}
	if err := cmd.Start(); err != nil {
		stdinR.Close()
		stdinW.Close()
		stderrR.Close()
		stderrW.Close()
		return nil, fmt.Errorf("starting %s: %w", binary, err)
	}
	stdinR.Close()
	stderrW.Close()
	return w, nil
}

func (w *writer) Write(p []byte) (int, error) {
	return w.stdinW.Write(p)
}

func (w *writer) Close() error {
	w.closeOnce.Do(func() {
		w.err = w.close()
	})
	return w.err
}

func (w *writer) close() error {
	defer w.stderrR.Close()
	w.stdinW.Close()
	err := w.cmd.Wait()
	if err == nil {
		return nil
	}
	var ee *exec.ExitError
	if !errors.As(err, &ee) || !ee.Exited() {
		return err
	}
	data, _ := io.ReadAll(w.stderrR)
	if len(data) == 0 {
		return err
	}
	return fmt.Errorf("%w: %s", err, data)
}
