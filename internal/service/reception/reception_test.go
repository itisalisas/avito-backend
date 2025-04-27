package reception

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/generated/mocks"
	"github.com/itisalisas/avito-backend/internal/models"
)

func TestReceptionService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReceptionRepo := mocks.NewMockReceptionRepositoryInterface(ctrl)
	service := NewReceptionService(mockReceptionRepo)
	pvzId := uuid.New()
	receptionId := uuid.New()

	tests := []struct {
		name              string
		method            string
		request           interface{}
		mockActions       func()
		expectedErr       error
		expectedReception *dto.Reception
	}{
		{
			name:   "add reception success",
			method: "AddReception",
			request: dto.PostReceptionsJSONRequestBody{
				PvzId: pvzId,
			},
			mockActions: func() {
				mockReceptionRepo.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
				mockReceptionRepo.EXPECT().GetLastReceptionByPvzId(gomock.Any(), gomock.Any()).Return(nil, models.ErrReceptionNotFound).Times(1)
				mockReceptionRepo.EXPECT().AddReception(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockReceptionRepo.EXPECT().Commit().Return(nil).Times(1)
				mockReceptionRepo.EXPECT().Rollback().Return(nil).Times(1)
			},
			expectedErr: nil,
			expectedReception: &dto.Reception{
				PvzId: pvzId,
			},
		},
		{
			name:   "add reception not closed",
			method: "AddReception",
			request: dto.PostReceptionsJSONRequestBody{
				PvzId: pvzId,
			},
			mockActions: func() {
				mockReceptionRepo.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
				mockReceptionRepo.EXPECT().GetLastReceptionByPvzId(gomock.Any(), gomock.Any()).Return(&dto.Reception{
					Status: dto.InProgress,
				}, nil).Times(1)
				mockReceptionRepo.EXPECT().Rollback().Return(nil).Times(1)
			},
			expectedErr:       models.ErrReceptionNotClosed,
			expectedReception: nil,
		},
		{
			name:    "close reception success",
			method:  "CloseLastReception",
			request: pvzId,
			mockActions: func() {
				mockReceptionRepo.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
				mockReceptionRepo.EXPECT().GetLastReceptionByPvzId(gomock.Any(), gomock.Any()).Return(&dto.Reception{
					Id:     &receptionId,
					Status: dto.InProgress,
				}, nil).Times(1)
				mockReceptionRepo.EXPECT().CloseLastReception(gomock.Any(), gomock.Any()).Return(&dto.Reception{
					Id:     &receptionId,
					Status: dto.Close,
				}, nil).Times(1)
				mockReceptionRepo.EXPECT().Commit().Return(nil).Times(1)
				mockReceptionRepo.EXPECT().Rollback().Return(nil).Times(1)
			},
			expectedErr: nil,
			expectedReception: &dto.Reception{
				Id:     &receptionId,
				Status: dto.Close,
			},
		},
		{
			name:    "close reception already closed",
			method:  "CloseLastReception",
			request: pvzId,
			mockActions: func() {
				mockReceptionRepo.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
				mockReceptionRepo.EXPECT().GetLastReceptionByPvzId(gomock.Any(), gomock.Any()).Return(&dto.Reception{
					Id:     &receptionId,
					Status: dto.Close,
				}, nil).Times(1)
				mockReceptionRepo.EXPECT().Rollback().Return(nil).Times(1)
			},
			expectedErr:       models.ErrReceptionClosed,
			expectedReception: nil,
		},
		{
			name:   "add reception error on begin tx",
			method: "AddReception",
			request: dto.PostReceptionsJSONRequestBody{
				PvzId: pvzId,
			},
			mockActions: func() {
				mockReceptionRepo.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(nil, errors.New("tx error")).Times(1)
			},
			expectedErr:       errors.New("tx error"),
			expectedReception: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockActions()

			var err error
			var reception *dto.Reception

			switch tt.method {
			case "AddReception":
				reception, err = service.AddReception(context.Background(), tt.request.(dto.PostReceptionsJSONRequestBody))
			case "CloseLastReception":
				reception, err = service.CloseLastReception(context.Background(), tt.request.(types.UUID))
			}

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedReception != nil {
				assert.Equal(t, tt.expectedReception.PvzId, reception.PvzId)
			}
		})
	}
}
