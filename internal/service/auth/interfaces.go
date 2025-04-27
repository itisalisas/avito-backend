package auth

import (
	"context"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
)

type ServiceInterface interface {
	Register(ctx context.Context, request dto.PostRegisterJSONRequestBody) (*dto.User, error)
	DummyLogin(request dto.PostDummyLoginJSONRequestBody) (*dto.Token, error)
	Login(ctx context.Context, request dto.PostLoginJSONRequestBody) (*dto.Token, error)
}
