package storage

import (
	"context"
	"database/sql"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/itisalisas/avito-backend/internal/generated/dto"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type ProductRepository struct {
	DB *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{DB: db}
}

func (r *ProductRepository) AddProduct(ctx context.Context, product *dto.Product, tx *sql.Tx) error {
	query, args, err := squirrel.Insert("pvz_service.product").
		Columns("product_type", "reception_id").
		Values(product.Type, product.ReceptionId).
		Suffix("returning product_id, added_at").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return err
	}

	err = tx.QueryRowContext(ctx, query, args...).Scan(&product.Id, &product.DateTime)

	if err != nil {
		return err
	}

	return nil
}

func (r *ProductRepository) GetLastProduct(ctx context.Context, receptionId uuid.UUID, tx *sql.Tx) (*dto.Product, error) {
	query, args, err := squirrel.Select("product_id", "product_type", "reception_id", "added_at").
		From("pvz_service.product").
		Where("reception_id = $1", receptionId).
		OrderBy("added_at DESC").
		Limit(1).
		ToSql()

	if err != nil {
		return nil, err
	}

	product := &dto.Product{}

	err = tx.QueryRowContext(ctx, query, args...).Scan(&product.Id, &product.Type, &product.ReceptionId, &product.DateTime)
	switch {
	case err != nil:
		return nil, err
	default:
		return product, nil
	}
}

func (r *ProductRepository) DeleteProductById(ctx context.Context, productID openapi_types.UUID, tx *sql.Tx) error {
	query, args, err := squirrel.Delete("pvz_service.product").
		Where("product_id = $1", productID).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, query, args...)
	return err
}
