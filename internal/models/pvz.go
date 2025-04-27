package models

import "github.com/itisalisas/avito-backend/internal/generated/dto"

type ExtendedPvz struct {
	PVZ        dto.PVZ             `json:"pvz"`
	Receptions []ExtendedReception `json:"receptions"`
}

type ExtendedReception struct {
	Reception dto.Reception `json:"reception"`
	Products  []dto.Product `json:"products"`
}
