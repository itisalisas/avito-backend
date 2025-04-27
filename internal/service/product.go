package service

import (
	"context"
	"database/sql"
	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
	"github.com/itisalisas/avito-backend/internal/storage"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"log"
)

type ProductService struct {
	productRepo   *storage.ProductRepository
	receptionRepo *storage.ReceptionRepository
}

func NewProductService(db *sql.DB) *ProductService {
	return &ProductService{productRepo: storage.NewProductRepository(db),
		receptionRepo: storage.NewReceptionRepository(db)}
}

func (s *ProductService) AddProduct(ctx context.Context, request dto.PostProductsJSONRequestBody) (*dto.Product, error) {
	tx, err := s.receptionRepo.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			log.Fatalf("Error while rolling back transaction: %s", err)
		}
	}(tx)

	if !isValidProductType(dto.ProductType(request.Type)) {
		return nil, models.ErrIncorrectProductType
	}

	reception, err := s.receptionRepo.GetLastReceptionByPvzId(ctx, request.PvzId, tx)
	if err != nil {
		return nil, err
	}

	if reception.Status != dto.InProgress {
		return nil, models.ErrReceptionClosed
	}

	product := &dto.Product{
		Type:        dto.ProductType(request.Type),
		ReceptionId: *reception.Id,
	}

	if err = s.productRepo.AddProduct(ctx, product, tx); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return product, nil
}

func isValidProductType(productType dto.ProductType) bool {
	return productType == dto.ProductTypeЭлектроника ||
		productType == dto.ProductTypeОдежда ||
		productType == dto.ProductTypeОбувь
}

func (s *ProductService) DeleteLastProduct(ctx context.Context, pvzId openapi_types.UUID) error {
	tx, err := s.receptionRepo.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			log.Fatalf("Error while rolling back transaction: %s", err)
		}
	}(tx)

	reception, err := s.receptionRepo.GetLastReceptionByPvzId(ctx, pvzId, tx)
	if err != nil {
		return err
	}

	product, err := s.productRepo.GetLastProduct(ctx, *reception.Id, tx)
	if err != nil {
		return err
	}

	err = s.productRepo.DeleteProductById(ctx, *product.Id, tx)
	if err != nil {
		return err
	}
	return tx.Commit()
}
