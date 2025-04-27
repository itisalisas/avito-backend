package pvz

import (
	"context"
	"time"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
)

type ServiceInterface interface {
	AddPvz(ctx context.Context, pvz *dto.PostPvzJSONRequestBody) (*dto.PVZ, error)
	GetPvzList(ctx context.Context, startTime *time.Time, endTime *time.Time, page uint64, limit uint64) ([]*models.ExtendedPvz, error)
}
