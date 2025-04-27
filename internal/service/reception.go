package service

import (
	"context"
	"database/sql"
	"errors"
	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
	"github.com/itisalisas/avito-backend/internal/storage"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type ReceptionService struct {
	receptionRepo *storage.ReceptionRepository
}

func NewReceptionService(db *sql.DB) *ReceptionService {
	return &ReceptionService{receptionRepo: storage.NewReceptionRepository(db)}
}

func (s *ReceptionService) AddReception(ctx context.Context, request dto.PostReceptionsJSONRequestBody) (*dto.Reception, error) {
	tx, err := s.receptionRepo.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	lastReception, err := s.receptionRepo.GetLastReceptionByPvzId(ctx, request.PvzId, tx)
	if err != nil && !errors.Is(err, models.ErrReceptionNotFound) {
		return nil, err
	}

	if lastReception != nil && lastReception.Status != dto.Close {
		return nil, models.ErrReceptionNotClosed
	}

	reception := dto.Reception{
		PvzId: request.PvzId,
	}

	if err := s.receptionRepo.AddReception(ctx, &reception, tx); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &reception, nil
}

func (s *ReceptionService) CloseLastReception(ctx context.Context, pvzId openapi_types.UUID) (*dto.Reception, error) {
	tx, err := s.receptionRepo.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	reception, err := s.receptionRepo.GetLastReceptionByPvzId(ctx, pvzId, tx)
	if err != nil {
		return nil, err
	}

	if reception.Status == dto.Close {
		return nil, models.ErrReceptionClosed
	}

	updReception, err := s.receptionRepo.CloseLastReception(ctx, *reception.Id, tx)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return updReception, err
}
