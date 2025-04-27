package product

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/generated/mocks"
	"github.com/itisalisas/avito-backend/internal/models"
)

func TestProductService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := mocks.NewMockProductRepositoryInterface(ctrl)
	mockReceptionRepo := mocks.NewMockReceptionRepositoryInterface(ctrl)
	service := NewProductService(mockProductRepo, mockReceptionRepo)
	pvzId := uuid.New()
	receptionId := uuid.New()
	productId := uuid.New()

	tests := []struct {
		name            string
		method          string
		request         interface{}
		mockActions     func()
		expectedErr     error
		expectedProduct *dto.Product
	}{
		{
			name:   "add product success",
			method: "AddProduct",
			request: dto.PostProductsJSONRequestBody{
				PvzId: pvzId,
				Type:  "электроника",
			},
			mockActions: func() {
				mockReceptionRepo.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(&sql.Tx{}, nil).Times(1)
				mockReceptionRepo.EXPECT().GetLastReceptionByPvzId(gomock.Any(), gomock.Any()).Return(&dto.Reception{
					Id:     &receptionId,
					Status: dto.InProgress,
				}, nil).Times(1)
				mockProductRepo.EXPECT().AddProduct(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockReceptionRepo.EXPECT().Commit().Return(nil).Times(1)
				mockReceptionRepo.EXPECT().Rollback().Return(nil).Times(1)
			},
			expectedErr: nil,
			expectedProduct: &dto.Product{
				Type:        dto.ProductTypeЭлектроника,
				ReceptionId: receptionId,
			},
		},
		{
			name:   "add product invalid type",
			method: "AddProduct",
			request: dto.PostProductsJSONRequestBody{
				PvzId: pvzId,
				Type:  "InvalidType",
			},
			mockActions: func() {
				mockReceptionRepo.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(&sql.Tx{}, nil).Times(1)
				mockReceptionRepo.EXPECT().GetLastReceptionByPvzId(gomock.Any(), gomock.Any()).Return(&dto.Reception{
					Id:     &receptionId,
					Status: dto.Close,
				}, nil).Times(1)
				mockReceptionRepo.EXPECT().Rollback().Return(nil).Times(1)
			},
			expectedErr:     models.ErrIncorrectProductType,
			expectedProduct: nil,
		},
		{
			name:   "add product closed reception",
			method: "AddProduct",
			request: dto.PostProductsJSONRequestBody{
				PvzId: pvzId,
				Type:  "электроника",
			},
			mockActions: func() {
				mockReceptionRepo.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(&sql.Tx{}, nil).Times(1)
				mockReceptionRepo.EXPECT().Rollback().Return(nil).Times(1)
			},
			expectedErr:     models.ErrReceptionClosed,
			expectedProduct: nil,
		},
		{
			name:    "delete last product success",
			method:  "DeleteLastProduct",
			request: pvzId,
			mockActions: func() {
				mockReceptionRepo.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(&sql.Tx{}, nil).Times(1)
				mockReceptionRepo.EXPECT().GetLastReceptionByPvzId(gomock.Any(), gomock.Any()).Return(&dto.Reception{
					Id:     &receptionId,
					Status: dto.InProgress,
				}, nil).Times(1)
				mockProductRepo.EXPECT().GetLastProduct(gomock.Any(), gomock.Any()).Return(&dto.Product{
					Id: &productId,
				}, nil).Times(1)
				mockProductRepo.EXPECT().DeleteProductById(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockReceptionRepo.EXPECT().Commit().Return(nil).Times(1)
				mockReceptionRepo.EXPECT().Rollback().Return(nil).Times(1)
			},
			expectedErr:     nil,
			expectedProduct: nil,
		},
		{
			name:    "delete last product error reception",
			method:  "DeleteLastProduct",
			request: pvzId,
			mockActions: func() {
				mockReceptionRepo.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(&sql.Tx{}, nil).Times(1)
				mockReceptionRepo.EXPECT().GetLastReceptionByPvzId(gomock.Any(), gomock.Any()).Return(nil, models.ErrReceptionClosed).Times(1)
				mockReceptionRepo.EXPECT().Rollback().Return(nil).Times(1)
			},
			expectedErr:     models.ErrReceptionClosed,
			expectedProduct: nil,
		},
		{
			name:    "delete last product error no products",
			method:  "DeleteLastProduct",
			request: pvzId,
			mockActions: func() {
				mockReceptionRepo.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(&sql.Tx{}, nil).Times(1)
				mockReceptionRepo.EXPECT().GetLastReceptionByPvzId(gomock.Any(), gomock.Any()).Return(&dto.Reception{
					Id: &receptionId,
				}, nil).Times(1)
				mockProductRepo.EXPECT().GetLastProduct(gomock.Any(), gomock.Any()).Return(nil, models.ErrNoProductsInReception).Times(1)
				mockReceptionRepo.EXPECT().Rollback().Return(nil).Times(1)
			},
			expectedErr:     models.ErrNoProductsInReception,
			expectedProduct: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockActions()

			var err error
			var product *dto.Product

			switch tt.method {
			case "AddProduct":
				product, err = service.AddProduct(context.Background(), tt.request.(dto.PostProductsJSONRequestBody))
			case "DeleteLastProduct":
				err = service.DeleteLastProduct(context.Background(), tt.request.(types.UUID))
			}

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedProduct != nil {
				assert.Equal(t, tt.expectedProduct.Type, product.Type)
				assert.Equal(t, tt.expectedProduct.ReceptionId, product.ReceptionId)
			}
		})
	}
}
