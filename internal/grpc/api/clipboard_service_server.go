package api

import (
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

func (s *ClipboardServiceServer) Paste(_ *apiv1.PasteRequest, stream apiv1.ClipboardService_PasteServer) error {
	return s.service.Paste(stream.Context(), pasteStream{stream: stream})
}

type pasteStream struct {
	stream apiv1.ClipboardService_PasteServer
}

func (p pasteStream) Write(b []byte) (int, error) {
	if err := p.stream.Send(&apiv1.PasteResponse{Data: b}); err != nil {
		return 0, err
	}
	return len(b), nil
}
