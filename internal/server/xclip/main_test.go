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
	display, stopXvfb := startXvfb()
	os.Setenv("DISPLAY", display)
	code := m.Run()
	stopXvfb()
	os.Exit(code)
}

func startXvfb() (string, func()) {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	cmd := exec.Command("Xvfb", "-displayfd", "3", "-screen", "0", "1024x768x24", "-nolisten", "tcp", "-ac")
	cmd.ExtraFiles = []*os.File{w}
	if err := cmd.Start(); err != nil {
		panic(fmt.Sprintf("starting Xvfb: %v", err))
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
			panic("Xvfb did not report a display")
		}
		display := ":" + res.display
		stopWithSocketCleanup := func() {
			stop()
			os.Remove("/tmp/.X11-unix/X" + res.display)
		}
		return display, stopWithSocketCleanup
	case <-time.After(20 * time.Second):
		stop()
		panic("Xvfb did not become ready")
	}
}
