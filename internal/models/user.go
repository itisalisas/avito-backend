package models

import (
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
)

type User struct {
	ID       openapi_types.UUID  `json:"id"`
	Email    openapi_types.Email `json:"email"`
	Role     dto.UserRole        `json:"role"`
	Password string
}
