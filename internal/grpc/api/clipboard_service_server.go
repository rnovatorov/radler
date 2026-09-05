package api

import (
	"bytes"
	"io"

	"google.golang.org/grpc"

	"github.com/rnovatorov/radler/gen/go/api/v1"
	"github.com/rnovatorov/radler/internal/core"
)

type ClipboardServiceServer struct {
	apiv1.UnimplementedClipboardServiceServer
	service *core.ClipboardService
}

func NewClipboardServiceServer(service *core.ClipboardService) *ClipboardServiceServer {
	return &ClipboardServiceServer{service: service}
}

func (s *ClipboardServiceServer) Register(r grpc.ServiceRegistrar) {
	apiv1.RegisterClipboardServiceServer(r, s)
}

func (s *ClipboardServiceServer) Copy(stream apiv1.ClipboardService_CopyServer) error {
	if err := s.service.Copy(stream.Context(), &copyStream{stream: stream}); err != nil {
		return err
	}
	return stream.SendAndClose(&apiv1.CopyResponse{})
}

func (s *ClipboardServiceServer) Paste(_ *apiv1.PasteRequest, stream apiv1.ClipboardService_PasteServer) error {
	return s.service.Paste(stream.Context(), pasteStream{stream: stream})
}

type copyStream struct {
	stream  apiv1.ClipboardService_CopyServer
	pending []byte
}

func (c *copyStream) Read(b []byte) (int, error) {
	for len(c.pending) == 0 {
		req, err := c.stream.Recv()
		if err == io.EOF {
			return 0, io.EOF
		}
		if err != nil {
			return 0, err
		}
		c.pending = req.GetData()
	}
	n := copy(b, c.pending)
	c.pending = c.pending[n:]
	return n, nil
}

type pasteStream struct {
	stream apiv1.ClipboardService_PasteServer
}

func (p pasteStream) Write(b []byte) (int, error) {
	if err := p.stream.Send(&apiv1.PasteResponse{Data: bytes.Clone(b)}); err != nil {
		return 0, err
	}
	return len(b), nil
}
