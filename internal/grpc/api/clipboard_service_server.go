package api

import (
	"google.golang.org/grpc"

	"github.com/rnovatorov/radler/gen/go/radler/v1"
	"github.com/rnovatorov/radler/internal/core"
)

type ClipboardServiceServer struct {
	radlerv1.UnimplementedClipboardServiceServer
	service *core.ClipboardService
}

func NewClipboardServiceServer(service *core.ClipboardService) *ClipboardServiceServer {
	return &ClipboardServiceServer{service: service}
}

func (s *ClipboardServiceServer) Register(r grpc.ServiceRegistrar) {
	radlerv1.RegisterClipboardServiceServer(r, s)
}

func (s *ClipboardServiceServer) Paste(_ *radlerv1.PasteRequest, stream radlerv1.ClipboardService_PasteServer) error {
	return s.service.Paste(stream.Context(), pasteStream{stream: stream})
}

type pasteStream struct {
	stream radlerv1.ClipboardService_PasteServer
}

func (p pasteStream) Write(b []byte) (int, error) {
	if err := p.stream.Send(&radlerv1.PasteResponse{Data: b}); err != nil {
		return 0, err
	}
	return len(b), nil
}
