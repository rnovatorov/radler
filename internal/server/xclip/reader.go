package xclip

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// errTargetStringNotAvailable is xclip's stderr wording when the selection
// (clipboard) is empty. It may need updating if xclip changes its message.
const errTargetStringNotAvailable = "target STRING not available"

type reader struct {
	cmd     *exec.Cmd
	stdoutR *os.File
	stderrR *os.File

	closeOnce sync.Once
	err       error
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
		r.err = r.close()
	})
	return r.err
}

func (r *reader) close() error {
	r.cmd.Cancel()
	defer r.stdoutR.Close()
	defer r.stderrR.Close()
	err := r.cmd.Wait()
	if err == nil {
		return nil
	}
	data, _ := io.ReadAll(r.stderrR)
	var ee *exec.ExitError
	if errors.As(err, &ee) && ee.Exited() && strings.Contains(string(data), errTargetStringNotAvailable) {
		return nil
	}
	if len(data) == 0 {
		return err
	}
	return fmt.Errorf("%w: %s", err, data)
}
