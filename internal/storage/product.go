package storage

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
)

type ProductRepository struct {
	*BaseRepository
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *ProductRepository) AddProduct(ctx context.Context, product *dto.Product) error {
	query, args, err := squirrel.Insert("pvz_service.product").
		Columns("product_type", "reception_id").
		Values(product.Type, product.ReceptionId).
		Suffix("returning product_id, added_at").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return err
	}
	err = r.tx.QueryRowContext(ctx, query, args...).Scan(&product.Id, &product.DateTime)

	if err != nil {
		return err
	}

	return nil
}

func (r *ProductRepository) GetLastProduct(ctx context.Context, receptionId uuid.UUID) (*dto.Product, error) {
	query, args, err := squirrel.Select("product_id", "product_type", "reception_id", "added_at").
		From("pvz_service.product").
		Where(squirrel.Eq{"reception_id": receptionId}).
		OrderBy("added_at DESC").
		Limit(1).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, err
	}

	product := &dto.Product{}

	err = r.tx.QueryRowContext(ctx, query, args...).Scan(&product.Id, &product.Type, &product.ReceptionId, &product.DateTime)
	switch {
	case err != nil:
		return nil, err
	default:
		return product, nil
	}
}

func (r *ProductRepository) DeleteProductById(ctx context.Context, productID openapi_types.UUID) error {
	query, args, err := squirrel.Delete("pvz_service.product").
		Where(squirrel.Eq{"product_id": productID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return err
	}

	_, err = r.tx.ExecContext(ctx, query, args...)
	return err
}
