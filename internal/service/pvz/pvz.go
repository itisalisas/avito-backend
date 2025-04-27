package pvz

import (
	"context"
	"log"
	"time"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
	"github.com/itisalisas/avito-backend/internal/storage"
)

type Service struct {
	pvzRepo storage.PvzRepositoryInterface
}

func NewPvzService(pvzRepo storage.PvzRepositoryInterface) *Service {
	return &Service{pvzRepo: pvzRepo}
}

func (s *Service) AddPvz(ctx context.Context, request *dto.PostPvzJSONRequestBody) (*dto.PVZ, error) {
	_, err := s.pvzRepo.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := s.pvzRepo.Rollback()
		if err != nil {
			log.Fatalf("Error while rolling back transaction: %s", err)
		}
	}()

	if !isValidCity(request.City) {
		return nil, models.ErrIncorrectCity
	}

	pvz := dto.PVZ{
		City: request.City,
	}

	if err := s.pvzRepo.CreatePvz(ctx, &pvz); err != nil {
		return nil, err
	}

	return &pvz, s.pvzRepo.Commit()
}

func isValidCity(city dto.PVZCity) bool {
	return city == dto.Москва || city == dto.Казань || city == dto.СанктПетербург
}

func (s *Service) GetPvzList(ctx context.Context, startTime *time.Time, endTime *time.Time, page uint64, limit uint64) ([]*models.ExtendedPvz, error) {
	_, err := s.pvzRepo.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := s.pvzRepo.Rollback()
		if err != nil {
			log.Fatalf("Error while rolling back transaction: %s", err)
		}
	}()

	pvzList, err := s.pvzRepo.GetPvzList(ctx, startTime, endTime, page, limit)
	if err != nil {
		return nil, err
	}

	if err := s.pvzRepo.Commit(); err != nil {
		return nil, err
	}
	return pvzList, nil
}

func (s *Service) GetAllPVZ(ctx context.Context) ([]dto.PVZ, error) {
	return s.pvzRepo.GetAllPVZs(ctx)
}
