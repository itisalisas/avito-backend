package grpc

import (
	"context"

	"github.com/itisalisas/avito-backend/internal/service/pvz"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PVZServer struct {
	UnimplementedPVZServiceServer
	service *pvz.Service
}

func NewPVZServer(s *pvz.Service) *PVZServer {
	return &PVZServer{service: s}
}

func (s *PVZServer) GetPVZList(ctx context.Context, req *GetPVZListRequest) (*GetPVZListResponse, error) {
	pvzs, err := s.service.GetAllPVZ(ctx)
	if err != nil {
		return nil, err
	}

	resp := &GetPVZListResponse{}
	for _, p := range pvzs {
		resp.Pvzs = append(resp.Pvzs, &PVZ{
			Id:               p.Id.String(),
			RegistrationDate: timestamppb.New(*p.RegistrationDate),
			City:             string(p.City),
		})
	}
	return resp, nil
}

func RegisterGRPCServer(s *grpc.Server, pvzService *pvz.Service) {
	RegisterPVZServiceServer(s, NewPVZServer(pvzService))
}
