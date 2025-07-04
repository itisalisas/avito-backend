package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
)

type PvzRepository struct {
	*BaseRepository
}

func (r *PvzRepository) GetAllPVZs(ctx context.Context) ([]dto.PVZ, error) {
	query, _, err := squirrel.Select("pvz_id", "city", "registration_date").
		From("pvz_service.pvz").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rows)

	var pvzs []dto.PVZ
	for rows.Next() {
		var p dto.PVZ
		if err := rows.Scan(&p.Id, &p.City, &p.RegistrationDate); err != nil {
			return nil, err
		}
		pvzs = append(pvzs, p)
	}
	return pvzs, nil
}

func NewPvzRepository(db *sql.DB) *PvzRepository {
	return &PvzRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *PvzRepository) CreatePvz(ctx context.Context, pvz *dto.PVZ) error {
	query, args, err := squirrel.Insert("pvz_service.pvz").
		Columns("city").
		Values(pvz.City).
		Suffix("returning pvz_id, registration_date").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	err = r.tx.QueryRowContext(ctx, query, args...).Scan(&pvz.Id, &pvz.RegistrationDate)

	if err != nil {
		return fmt.Errorf("failed to insert pvz: %w", err)
	}

	return nil
}

func (r *PvzRepository) GetPvzList(ctx context.Context, startTime *time.Time, endTime *time.Time, page uint64, limit uint64) ([]*models.ExtendedPvz, error) {
	baseQuery := squirrel.Select(
		"p.pvz_id",
		"p.city",
		"p.registration_date",
		"r.reception_id",
		"r.started_at",
		"r.status",
		"pr.product_id",
		"pr.added_at",
		"pr.product_type",
	).
		From("pvz_service.pvz p").
		LeftJoin("pvz_service.reception r ON p.pvz_id = r.pvz_id").
		LeftJoin("pvz_service.product pr ON r.reception_id = pr.reception_id").
		PlaceholderFormat(squirrel.Dollar)

	if startTime != nil && endTime != nil {
		baseQuery = baseQuery.Where("r.started_at BETWEEN ? AND ?", *startTime, *endTime)
	}

	subQuery := baseQuery.
		GroupBy(
			"p.pvz_id",
			"r.reception_id",
			"r.started_at",
			"r.status",
			"pr.product_id",
			"pr.added_at",
			"pr.product_type",
		).
		OrderBy("p.registration_date DESC").
		Limit(limit).
		Offset((page - 1) * limit)

	query, args, err := subQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query pvzs: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatalf("failed to close rows: %v", err)
		}
	}(rows)

	pvzMap := make(map[uuid.UUID]*models.ExtendedPvz)
	receptionMap := make(map[uuid.UUID]*models.ExtendedReception)

	for rows.Next() {
		var (
			pvzID          openapi_types.UUID
			city           dto.PVZCity
			regDate        time.Time
			receptionID    *uuid.UUID
			startedAt      *time.Time
			status         *dto.ReceptionStatus
			productID      *uuid.UUID
			productAddedAt *time.Time
			productType    *dto.ProductType
		)

		err := rows.Scan(
			&pvzID,
			&city,
			&regDate,
			&receptionID,
			&startedAt,
			&status,
			&productID,
			&productAddedAt,
			&productType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if _, exists := pvzMap[pvzID]; !exists {
			pvzMap[pvzID] = &models.ExtendedPvz{
				PVZ: dto.PVZ{
					Id:               &pvzID,
					City:             city,
					RegistrationDate: &regDate,
				},
				Receptions: []models.ExtendedReception{},
			}
		}
		pvz := pvzMap[pvzID]

		if receptionID != nil {
			if _, exists := receptionMap[*receptionID]; !exists {
				receptionMap[*receptionID] = &models.ExtendedReception{
					Reception: dto.Reception{
						Id:       receptionID,
						DateTime: *startedAt,
						Status:   *status,
					},
					Products: []dto.Product{},
				}
				pvz.Receptions = append(pvz.Receptions, *receptionMap[*receptionID])
			}
			reception := receptionMap[*receptionID]

			if productID != nil {
				reception.Products = append(reception.Products, dto.Product{
					Id:       productID,
					DateTime: productAddedAt,
					Type:     *productType,
				})
			}
		}
	}

	result := make([]*models.ExtendedPvz, 0, len(pvzMap))
	for _, pvz := range pvzMap {
		log.Println(pvz.PVZ.City)
		result = append(result, pvz)
	}

	return result, nil
}
