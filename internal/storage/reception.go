package storage

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/Masterminds/squirrel"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
)

type ReceptionRepository struct {
	*BaseRepository
}

func NewReceptionRepository(db *sql.DB) *ReceptionRepository {
	return &ReceptionRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *ReceptionRepository) AddReception(ctx context.Context, reception *dto.Reception) error {
	query, args, err := squirrel.Insert("pvz_service.reception").
		Columns("pvz_id").
		Values(reception.PvzId).
		Suffix("returning reception_id, started_at, status").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	log.Println(query)

	if err != nil {
		return err
	}

	err = r.tx.QueryRowContext(ctx, query, args...).Scan(&reception.Id, &reception.DateTime, &reception.Status)

	return err
}

func (r *ReceptionRepository) GetLastReceptionByPvzId(ctx context.Context, pvzId openapi_types.UUID) (*dto.Reception, error) {
	query, args, err := squirrel.Select("reception_id", "started_at", "status", "pvz_id").
		From("pvz_service.reception").
		Where("pvz_id = $1", pvzId).
		OrderBy("started_at DESC").
		Limit(1).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, err
	}

	reception := &dto.Reception{}

	err = r.tx.QueryRowContext(ctx, query, args...).Scan(&reception.Id, &reception.DateTime, &reception.Status, &reception.PvzId)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, models.ErrReceptionNotFound
	case err != nil:
		return nil, err
	default:
		return reception, nil
	}
}

func (r *ReceptionRepository) CloseLastReception(ctx context.Context, receptionId openapi_types.UUID) (*dto.Reception, error) {
	query, args, err := squirrel.Update("pvz_service.reception").
		Set("status", string(dto.Close)).
		Where("reception_id = $2", receptionId).
		Suffix("returning reception_id, started_at, status, pvz_id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, err
	}

	reception := &dto.Reception{}

	err = r.tx.QueryRowContext(ctx, query, args...).Scan(&reception.Id, &reception.DateTime, &reception.Status, &reception.PvzId)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, models.ErrReceptionNotFound
	case err != nil:
		return nil, err
	default:
		return reception, nil
	}
}
