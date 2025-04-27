package auth

import (
	"context"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
)

type ServiceInterface interface {
	Register(ctx context.Context, request dto.PostRegisterJSONRequestBody) (*models.User, error)
	DummyLogin(request dto.PostDummyLoginJSONRequestBody) (*dto.Token, error)
	Login(ctx context.Context, request dto.PostLoginJSONRequestBody) (*dto.Token, error)
}
