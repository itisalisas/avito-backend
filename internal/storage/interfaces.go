package storage

import (
	"context"
	"database/sql"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
)

type ProductRepositoryInterface interface {
	TransactionStorage
	AddProduct(ctx context.Context, product *dto.Product) error
	GetLastProduct(ctx context.Context, receptionId openapi_types.UUID) (*dto.Product, error)
	DeleteProductById(ctx context.Context, productId openapi_types.UUID) error
}

type ReceptionRepositoryInterface interface {
	TransactionStorage
	GetLastReceptionByPvzId(ctx context.Context, pvzId openapi_types.UUID) (*dto.Reception, error)
	AddReception(ctx context.Context, reception *dto.Reception) error
	CloseLastReception(ctx context.Context, receptionId openapi_types.UUID) (*dto.Reception, error)
}

type PvzRepositoryInterface interface {
	TransactionStorage
	CreatePvz(ctx context.Context, pvz *dto.PVZ) error
	GetPvzList(ctx context.Context, startTime *time.Time, endTime *time.Time, page uint64, limit uint64) ([]*models.ExtendedPvz, error)
	GetAllPVZs(ctx context.Context) ([]dto.PVZ, error)
}

type UserRepositoryInterface interface {
	TransactionStorage
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email openapi_types.Email) (*models.User, error)
}

type TransactionStorage interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Commit() error
	Rollback() error
}
