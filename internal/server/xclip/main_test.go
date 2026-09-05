package xclip

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if _, err := exec.LookPath("Xvfb"); err != nil {
		fmt.Fprintln(os.Stderr, "skipping xclip tests: Xvfb binary not found")
		os.Exit(0)
	}
	if _, err := exec.LookPath("xclip"); err != nil {
		fmt.Fprintln(os.Stderr, "skipping xclip tests: xclip binary not found")
		os.Exit(0)
	}

	display, stopXvfb, err := startXvfb()
	if err != nil {
		fmt.Fprintf(os.Stderr, "skipping xclip tests: %v\n", err)
		os.Exit(0)
	}

	os.Setenv("DISPLAY", display)
	code := m.Run()
	stopXvfb()
	os.Exit(code)
}

func startXvfb() (string, func(), error) {
	r, w, err := os.Pipe()
	if err != nil {
		return "", nil, fmt.Errorf("creating pipe: %w", err)
	}
	cmd := exec.Command("Xvfb", "-displayfd", "3", "-screen", "0", "1024x768x24", "-nolisten", "tcp", "-ac")
	cmd.ExtraFiles = []*os.File{w}
	if err := cmd.Start(); err != nil {
		w.Close()
		r.Close()
		return "", nil, fmt.Errorf("starting Xvfb: %w", err)
	}
	w.Close()
	type result struct {
		display string
		err     error
	}
	ch := make(chan result, 1)
	go func() {
		data, err := io.ReadAll(r)
		ch <- result{strings.TrimSpace(string(data)), err}
	}()
	stop := func() {
		cmd.Process.Kill()
		cmd.Wait()
		r.Close()
	}
	select {
	case res := <-ch:
		if res.err != nil || res.display == "" {
			stop()
			return "", nil, fmt.Errorf("Xvfb did not report a display")
		}
		display := ":" + res.display
		stopWithSocketCleanup := func() {
			stop()
			os.Remove("/tmp/.X11-unix/X" + res.display)
		}
		return display, stopWithSocketCleanup, nil
	case <-time.After(20 * time.Second):
		stop()
		return "", nil, fmt.Errorf("Xvfb did not become ready")
	}
}
