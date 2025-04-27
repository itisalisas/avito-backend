package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
)

func TestWriteResponse(t *testing.T) {
	t.Run("should write JSON response with correct status code", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		responseBody := dto.Error{Message: "test error"}

		WriteResponse(recorder, responseBody, http.StatusBadRequest)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)

		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
		var response dto.Error
		err := json.NewDecoder(recorder.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "test error", response.Message)
	})
}

func TestError(t *testing.T) {
	t.Run("should return Error struct with correct message", func(t *testing.T) {
		msg := "error message"
		err := Error(msg)
		assert.Equal(t, msg, err.Message)
	})
}
