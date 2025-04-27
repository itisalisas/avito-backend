package service

import (
	"context"
	"database/sql"
	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
	"github.com/itisalisas/avito-backend/internal/storage"
	"time"
)

type PvzService struct {
	pvzRepo *storage.PvzRepository
}

func NewPvzService(db *sql.DB) *PvzService {
	return &PvzService{pvzRepo: storage.NewPvzRepository(db)}
}

func (s *PvzService) AddPvz(ctx context.Context, pvz *dto.PVZ) error {
	tx, err := s.pvzRepo.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if !isValidCity(pvz.City) {
		return models.ErrIncorrectCity
	}

	if err := s.pvzRepo.CreatePvz(ctx, pvz, tx); err != nil {
		return err
	}

	return tx.Commit()
}

func isValidCity(city dto.PVZCity) bool {
	return city == dto.Москва || city == dto.Казань || city == dto.СанктПетербург
}

func (s *PvzService) GetPvzList(ctx context.Context, startTime *time.Time, endTime *time.Time, page uint64, limit uint64) ([]*models.ExtendedPvz, error) {
	tx, err := s.pvzRepo.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	pvzList, err := s.pvzRepo.GetPvzList(ctx, startTime, endTime, page, limit, tx)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return pvzList, nil
}
