package storage

import (
	"context"
	"database/sql"
	"log"
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	db      *sql.DB
	cleanup func()
	repo    *UserRepository
	ctx     context.Context
}

func TestUserRepositorySuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (s *UserRepositoryTestSuite) SetupSuite() {
	s.ctx = context.Background()
	db := DBTestSetup()
	log.Println("migrations applied")
	s.db = db
	s.repo = NewUserRepository(s.db)
}

func (s *UserRepositoryTestSuite) TearDownSuite() {
	err := s.db.Close()
	if err != nil {
		log.Fatalf("failed to close database connection: %v", err)
	}
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *UserRepositoryTestSuite) SetupTest() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	require.NoError(s.T(), err)
	s.repo.tx = tx
}

func (s *UserRepositoryTestSuite) TearDownTest() {
	if s.repo.tx != nil {
		err := s.repo.tx.Rollback()
		require.NoError(s.T(), err)
	}
}

func (s *UserRepositoryTestSuite) createUser(t *testing.T, email openapi_types.Email, password string, role dto.UserRole) *models.User {
	user := &models.User{
		Email:    email,
		Password: password,
		Role:     role,
	}
	err := s.repo.CreateUser(s.ctx, user)
	require.NoError(t, err)
	return user
}

func (s *UserRepositoryTestSuite) TestCreateUser() {
	type testCase struct {
		name       string
		input      *models.User
		wantErr    bool
		validateFn func(*testing.T, *models.User)
	}

	existingUser := s.createUser(s.T(), "test@example.com", "password123", dto.UserRole(dto.Employee))

	testCases := []testCase{
		{
			name: "successful user creation",
			input: &models.User{
				Email:    openapi_types.Email("newuser@example.com"),
				Password: "password456",
				Role:     dto.UserRole(dto.Employee),
			},
			wantErr: false,
			validateFn: func(t *testing.T, u *models.User) {
				assert.NotEqual(t, 0, u.ID)
				assert.Equal(t, u.Email, openapi_types.Email("newuser@example.com"))
				assert.Equal(t, u.Role, dto.UserRole(dto.Employee))

				var count int
				err := s.repo.tx.QueryRowContext(s.ctx, "select count(*) from pvz_service.user where user_id = $1", u.ID).Scan(&count)
				require.NoError(t, err)
				assert.Equal(t, 1, count)
			},
		},
		{
			name: "email already in use",
			input: &models.User{
				Email:    existingUser.Email,
				Password: "password123",
				Role:     dto.UserRole(dto.Employee),
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			err := s.repo.CreateUser(s.ctx, tc.input)
			if tc.wantErr {
				require.Error(t, err)
				assert.Equal(t, models.ErrEmailAlreadyInUse, err)
				return
			}
			require.NoError(t, err)

			if tc.validateFn != nil {
				tc.validateFn(t, tc.input)
			}
		})
	}
}

func (s *UserRepositoryTestSuite) TestGetUserByEmail() {
	user := s.createUser(s.T(), "get_user@example.com", "password123", dto.UserRole(dto.Employee))

	s.T().Run("successful user retrieval", func(t *testing.T) {
		result, err := s.repo.GetUserByEmail(s.ctx, user.Email)
		require.NoError(t, err)
		assert.Equal(t, user.Email, result.Email)
		assert.Equal(t, user.Role, result.Role)
	})

	s.T().Run("user not found", func(t *testing.T) {
		_, err := s.repo.GetUserByEmail(s.ctx, "nonexistent@example.com")
		require.Error(t, err)
		assert.Equal(t, "user not found", err.Error())
	})
}
