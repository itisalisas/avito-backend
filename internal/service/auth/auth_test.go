package auth

import (
	"context"
	"testing"

	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/generated/mocks"
	"github.com/itisalisas/avito-backend/internal/models"
)

func TestAuthService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepositoryInterface(ctrl)
	service := NewAuthService(mockRepo)

	tests := []struct {
		name          string
		method        string
		request       interface{}
		mockActions   func()
		expectedErr   error
		expectedUser  *models.User
		expectedToken *dto.Token
	}{
		{
			name:   "register success",
			method: "Register",
			request: dto.PostRegisterJSONRequestBody{
				Email:    "test@example.com",
				Password: "password123",
				Role:     "employee",
			},
			mockActions: func() {
				mockRepo.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(nil, nil)
				mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockRepo.EXPECT().Commit().Return(nil)
				mockRepo.EXPECT().Rollback().Return(nil)
			},
			expectedErr: nil,
			expectedUser: &models.User{
				Email: "test@example.com",
				Role:  dto.UserRoleEmployee,
			},
			expectedToken: nil,
		},
		{
			name:   "register invalid role",
			method: "Register",
			request: dto.PostRegisterJSONRequestBody{
				Email:    "test@example.com",
				Password: "password123",
				Role:     "invalid_role",
			},
			mockActions: func() {
				mockRepo.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(nil, nil)
				mockRepo.EXPECT().Rollback().Return(nil)
			},
			expectedErr:   models.ErrIncorrectUserRole,
			expectedUser:  nil,
			expectedToken: nil,
		},
		{
			name:   "login success",
			method: "Login",
			request: dto.PostLoginJSONRequestBody{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockActions: func() {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

				user := &models.User{
					Email:    "test@example.com",
					Password: string(hashedPassword),
					Role:     dto.UserRoleEmployee,
				}
				mockRepo.EXPECT().GetUserByEmail(gomock.Any(), types.Email("test@example.com")).Return(user, nil).Times(1)
			},
			expectedErr:   nil,
			expectedUser:  nil,
			expectedToken: strPtr("token"),
		},
		{
			name:   "login wrong password",
			method: "Login",
			request: dto.PostLoginJSONRequestBody{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			mockActions: func() {
				user := &models.User{
					Email:    "test@example.com",
					Password: "qwertrewq",
					Role:     dto.UserRoleEmployee,
				}
				mockRepo.EXPECT().GetUserByEmail(gomock.Any(), types.Email("test@example.com")).Return(user, nil).Times(1)
			},
			expectedErr:   models.ErrWrongPassword,
			expectedUser:  nil,
			expectedToken: nil,
		},
		{
			name:   "dummy login success",
			method: "DummyLogin",
			request: dto.PostDummyLoginJSONRequestBody{
				Role: "employee",
			},
			mockActions:   func() {},
			expectedErr:   nil,
			expectedUser:  nil,
			expectedToken: strPtr("token"),
		},
		{
			name:   "dummy invalid role",
			method: "DummyLogin",
			request: dto.PostDummyLoginJSONRequestBody{
				Role: "invalid role",
			},
			mockActions:   func() {},
			expectedErr:   models.ErrIncorrectUserRole,
			expectedUser:  nil,
			expectedToken: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockActions()

			var err error
			var token *dto.Token
			var user *dto.User

			switch tt.method {
			case "Register":
				user, err = service.Register(context.Background(), tt.request.(dto.PostRegisterJSONRequestBody))
			case "Login":
				token, err = service.Login(context.Background(), tt.request.(dto.PostLoginJSONRequestBody))
			case "DummyLogin":
				token, err = service.DummyLogin(tt.request.(dto.PostDummyLoginJSONRequestBody))
			}

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedUser != nil {
				assert.Equal(t, tt.expectedUser.Email, user.Email)
				assert.Equal(t, tt.expectedUser.Role, user.Role)
			}

			if tt.expectedToken != nil {
				assert.NotEmpty(t, token)
			} else {
				assert.Nil(t, token)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
