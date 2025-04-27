package utils

import (
	"encoding/json"
	"net/http"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
)

func WriteResponse[T any](w http.ResponseWriter, body T, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}

func Error(msg string) dto.Error {
	return dto.Error{Message: msg}
}
