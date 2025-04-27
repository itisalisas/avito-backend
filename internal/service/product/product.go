package product

import (
	"context"
	"database/sql"
	"log"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
	"github.com/itisalisas/avito-backend/internal/storage"
)

type Service struct {
	productRepo   storage.ProductRepositoryInterface
	receptionRepo storage.ReceptionRepositoryInterface
}

func NewProductService(productRepo storage.ProductRepositoryInterface,
	receptionRepo storage.ReceptionRepositoryInterface) *Service {
	return &Service{productRepo: productRepo,
		receptionRepo: receptionRepo}
}

func (s *Service) AddProduct(ctx context.Context, request dto.PostProductsJSONRequestBody) (*dto.Product, error) {
	_, err := s.receptionRepo.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	_, err = s.productRepo.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := s.receptionRepo.Rollback()
		if err != nil {
			log.Fatalf("Error while rolling back transaction: %s", err)
		}
	}()

	if !isValidProductType(dto.ProductType(request.Type)) {
		return nil, models.ErrIncorrectProductType
	}

	reception, err := s.receptionRepo.GetLastReceptionByPvzId(ctx, request.PvzId)
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

	if err = s.productRepo.AddProduct(ctx, product); err != nil {
		return nil, err
	}
	if err = s.receptionRepo.Commit(); err != nil {
		return nil, err
	}
	return product, nil
}

func isValidProductType(productType dto.ProductType) bool {
	return productType == dto.ProductTypeЭлектроника ||
		productType == dto.ProductTypeОдежда ||
		productType == dto.ProductTypeОбувь
}

func (s *Service) DeleteLastProduct(ctx context.Context, pvzId openapi_types.UUID) error {
	tx, err := s.receptionRepo.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		err := s.receptionRepo.Rollback()
		if err != nil {
			log.Fatalf("Error while rolling back transaction: %s", err)
		}
	}(tx)

	reception, err := s.receptionRepo.GetLastReceptionByPvzId(ctx, pvzId)
	if err != nil {
		return err
	}

	product, err := s.productRepo.GetLastProduct(ctx, *reception.Id)
	if err != nil {
		return err
	}

	err = s.productRepo.DeleteProductById(ctx, *product.Id)
	if err != nil {
		return err
	}
	return s.receptionRepo.Commit()
}
