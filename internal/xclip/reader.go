package xclip

import (
	"bytes"
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

	mu       sync.Mutex
	waitErr  error
	exited   chan struct{}
	drained  chan struct{}
	stderr   bytes.Buffer
	killOnce sync.Once
}

func newReader(ctx context.Context, binary string, env []string) (*reader, error) {
	argv := []string{binary, "-o", "-selection", "clipboard"}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Env = env
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
	cmd.Stdout = stdoutW
	cmd.Stderr = stderrW
	p := &reader{
		cmd:     cmd,
		stdoutR: stdoutR,
		stderrR: stderrR,
		exited:  make(chan struct{}),
		drained: make(chan struct{}),
	}
	if err := cmd.Start(); err != nil {
		stdoutR.Close()
		stdoutW.Close()
		stderrR.Close()
		stderrW.Close()
		return nil, fmt.Errorf("starting %s: %w", argv[0], err)
	}
	stdoutW.Close()
	stderrW.Close()
	go p.reap()
	go p.drain()
	go p.watch(ctx)
	return p, nil
}

func (p *reader) reap() {
	werr := p.cmd.Wait()
	p.mu.Lock()
	p.waitErr = werr
	p.mu.Unlock()
	close(p.exited)
}

// stderrR's only closer is drain at EOF; stdoutR's only closer is Close.
func (p *reader) drain() {
	defer close(p.drained)
	defer p.stderrR.Close()
	_, _ = io.Copy(&p.stderr, p.stderrR)
}

// Close and the ctx watcher are the concurrent kill callers; Once settles them.
func (p *reader) kill() {
	p.killOnce.Do(func() {
		p.cmd.Process.Kill()
	})
}

func (p *reader) watch(ctx context.Context) {
	select {
	case <-ctx.Done():
		p.kill()
	case <-p.exited:
	}
}

func (p *reader) Read(b []byte) (int, error) {
	return p.stdoutR.Read(b)
}

// Awaiting drained makes the returned error's stderr text complete.
func (p *reader) Close() error {
	p.kill()
	p.stdoutR.Close()
	<-p.exited
	<-p.drained
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.waitErr == nil {
		return nil
	}
	if s := p.stderr.String(); s != "" {
		return fmt.Errorf("%w: %s", p.waitErr, s)
	}
	return p.waitErr
}
