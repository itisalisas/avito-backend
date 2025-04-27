package product

import (
	"context"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
)

type ServiceInterface interface {
	AddProduct(ctx context.Context, request dto.PostProductsJSONRequestBody) (*dto.Product, error)
	DeleteLastProduct(ctx context.Context, pvzId openapi_types.UUID) error
}
