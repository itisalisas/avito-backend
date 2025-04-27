package models

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/itisalisas/avito-backend/internal/generated/dto"
)

type TokenClaims struct {
	Role dto.UserRole `json:"role"`
	jwt.RegisteredClaims
}
