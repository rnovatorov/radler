package core

import (
	"context"
	"io"
)

type ClipboardService struct {
	clipboard Clipboard
}

func NewClipboardService(clipboard Clipboard) *ClipboardService {
	return &ClipboardService{clipboard: clipboard}
}

func (s *ClipboardService) Copy(ctx context.Context, r io.Reader) error {
	w, err := s.clipboard.NewWriter(ctx)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, r); err != nil {
		w.Close()
		return err
	}
	return w.Close()
}

func (s *ClipboardService) Paste(ctx context.Context, w io.Writer) error {
	r, err := s.clipboard.NewReader(ctx)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, r); err != nil {
		r.Close()
		return err
	}
	return r.Close()
}
