package reception

import (
	"context"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
)

type ServiceInterface interface {
	AddReception(ctx context.Context, request dto.PostReceptionsJSONRequestBody) (*dto.Reception, error)
	CloseLastReception(ctx context.Context, pvzId openapi_types.UUID) (*dto.Reception, error)
}
