package pvz

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/generated/mocks"
	"github.com/itisalisas/avito-backend/internal/models"
)

func TestPvzService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPvzRepo := mocks.NewMockPvzRepositoryInterface(ctrl)
	service := NewPvzService(mockPvzRepo)
	pvzId := uuid.New()

	tests := []struct {
		name            string
		method          string
		request         interface{}
		mockActions     func()
		expectedErr     error
		expectedPvz     *dto.PVZ
		expectedPvzList []*models.ExtendedPvz
	}{
		{
			name:   "add pvz success",
			method: "AddPvz",
			request: &dto.PVZ{
				City: dto.Москва,
			},
			mockActions: func() {
				mockPvzRepo.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
				mockPvzRepo.EXPECT().CreatePvz(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockPvzRepo.EXPECT().Commit().Return(nil).Times(1)
				mockPvzRepo.EXPECT().Rollback().Return(nil).Times(1)
			},
			expectedErr: nil,
			expectedPvz: &dto.PVZ{
				City: dto.Москва,
			},
		},
		{
			name:   "add pvz invalid city",
			method: "AddPvz",
			request: &dto.PVZ{
				City: "InvalidCity",
			},
			mockActions: func() {
				mockPvzRepo.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
				mockPvzRepo.EXPECT().Rollback().Return(nil).Times(1)
			},
			expectedErr: models.ErrIncorrectCity,
			expectedPvz: nil,
		},
		{
			name:   "get pvz list success",
			method: "GetPvzList",
			request: struct {
				startTime *time.Time
				endTime   *time.Time
				page      uint64
				limit     uint64
			}{
				startTime: nil,
				endTime:   nil,
				page:      1,
				limit:     10,
			},
			mockActions: func() {
				mockPvzRepo.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
				mockPvzRepo.EXPECT().GetPvzList(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*models.ExtendedPvz{
					{
						PVZ: dto.PVZ{
							Id:   &pvzId,
							City: dto.Москва,
						},
						Receptions: []models.ExtendedReception{},
					},
				}, nil).Times(1)
				mockPvzRepo.EXPECT().Commit().Return(nil).Times(1)
				mockPvzRepo.EXPECT().Rollback().Return(nil).Times(1)
			},
			expectedErr: nil,
			expectedPvzList: []*models.ExtendedPvz{
				{
					PVZ: dto.PVZ{
						Id:   &pvzId,
						City: dto.Москва,
					},
					Receptions: []models.ExtendedReception{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockActions()

			var err error
			var pvz *dto.PVZ
			var pvzList []*models.ExtendedPvz

			switch tt.method {
			case "AddPvz":
				pvz, err = service.AddPvz(context.Background(), tt.request.(*dto.PVZ))
			case "GetPvzList":
				request := tt.request.(struct {
					startTime *time.Time
					endTime   *time.Time
					page      uint64
					limit     uint64
				})
				pvzList, err = service.GetPvzList(context.Background(), request.startTime, request.endTime, request.page, request.limit)
			}

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedPvz != nil {
				assert.Equal(t, tt.expectedPvz.City, pvz.City)
			}

			if tt.expectedPvzList != nil {
				assert.Len(t, pvzList, len(tt.expectedPvzList))
				for i, item := range tt.expectedPvzList {
					assert.Equal(t, item.PVZ.City, pvzList[i].PVZ.City)
				}
			}
		})
	}
}
