package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/itisalisas/avito-backend/internal/models"
)

type UserRepository struct {
	*BaseRepository
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	query, args, err := squirrel.Insert("pvz_service.user").
		Columns("email", "password", "role").
		Values(user.Email, user.Password, user.Role).
		Suffix("RETURNING user_id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	err = r.tx.QueryRowContext(ctx, query, args...).Scan(&user.ID)
	if err != nil {
		return models.ErrEmailAlreadyInUse
	}

	return nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email openapi_types.Email) (*models.User, error) {
	query, args, err := squirrel.Select("user_id", "email", "password", "role").
		From("pvz_service.user").
		Where(squirrel.Eq{"email": email}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var user models.User
	err = r.tx.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}
