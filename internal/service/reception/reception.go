package reception

import (
	"context"
	"errors"
	"log"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
	"github.com/itisalisas/avito-backend/internal/storage"
)

type Service struct {
	receptionRepo storage.ReceptionRepositoryInterface
}

func NewReceptionService(receptionRepo storage.ReceptionRepositoryInterface) *Service {
	return &Service{receptionRepo: receptionRepo}
}

func (s *Service) AddReception(ctx context.Context, request dto.PostReceptionsJSONRequestBody) (*dto.Reception, error) {
	_, err := s.receptionRepo.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := s.receptionRepo.Rollback()
		if err != nil {
			log.Fatalf("Error while rolling back transaction: %s", err)
		}
	}()

	lastReception, err := s.receptionRepo.GetLastReceptionByPvzId(ctx, request.PvzId)
	if err != nil && !errors.Is(err, models.ErrReceptionNotFound) {
		return nil, err
	}

	if lastReception != nil && lastReception.Status != dto.Close {
		return nil, models.ErrReceptionNotClosed
	}

	reception := dto.Reception{
		PvzId: request.PvzId,
	}

	if err := s.receptionRepo.AddReception(ctx, &reception); err != nil {
		return nil, err
	}

	if err := s.receptionRepo.Commit(); err != nil {
		return nil, err
	}

	return &reception, nil
}

func (s *Service) CloseLastReception(ctx context.Context, pvzId openapi_types.UUID) (*dto.Reception, error) {
	_, err := s.receptionRepo.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := s.receptionRepo.Rollback()
		if err != nil {
			log.Fatalf("Error while rolling back transaction: %s", err)
		}
	}()

	reception, err := s.receptionRepo.GetLastReceptionByPvzId(ctx, pvzId)
	if err != nil {
		return nil, err
	}

	if reception.Status == dto.Close {
		return nil, models.ErrReceptionClosed
	}

	updReception, err := s.receptionRepo.CloseLastReception(ctx, *reception.Id)

	if err := s.receptionRepo.Commit(); err != nil {
		return nil, err
	}

	return updReception, err
}
