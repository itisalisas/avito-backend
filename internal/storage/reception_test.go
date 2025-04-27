package storage

import (
	"context"
	"database/sql"
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
)

type ReceptionRepositoryTestSuite struct {
	suite.Suite
	db      *sql.DB
	cleanup func()
	repo    *ReceptionRepository
	ctx     context.Context
	pvzID   uuid.UUID
}

func TestReceptionRepositorySuite(t *testing.T) {
	suite.Run(t, new(ReceptionRepositoryTestSuite))
}

func (s *ReceptionRepositoryTestSuite) SetupSuite() {
	s.ctx = context.Background()
	db := DBTestSetup()
	log.Println("migrations applied")
	s.db = db

	// Создание репозитория без добавления данных о PVZ
	s.repo = NewReceptionRepository(s.db)
}

func (s *ReceptionRepositoryTestSuite) TearDownSuite() {
	err := s.db.Close()
	if err != nil {
		log.Fatalf("failed to close database connection: %v", err)
	}
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *ReceptionRepositoryTestSuite) SetupTest() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	require.NoError(s.T(), err)
	s.repo.tx = tx

	// Создание PVZ для каждого теста
	pvzID := uuid.New()
	_, err = s.repo.tx.ExecContext(s.ctx, `
        insert into pvz_service.pvz (pvz_id, registration_date, city)
        values ($1, current_date, 'Москва')`, pvzID)
	require.NoError(s.T(), err)

	// Сохраняем PVZ ID для использования в тестах
	s.pvzID = pvzID
}

func (s *ReceptionRepositoryTestSuite) TearDownTest() {
	if s.repo.tx != nil {
		err := s.repo.tx.Rollback()
		require.NoError(s.T(), err)
	}
}

func (s *ReceptionRepositoryTestSuite) createReception(t *testing.T) uuid.UUID {
	receptionID := uuid.New()
	_, err := s.repo.tx.ExecContext(s.ctx, `
		insert into pvz_service.reception (reception_id, started_at, pvz_id, status)
		values ($1, current_timestamp, $2, 'in_progress')`,
		receptionID, s.pvzID,
	)
	require.NoError(t, err)
	return receptionID
}

func (s *ReceptionRepositoryTestSuite) TestAddReception() {
	type testCase struct {
		name       string
		input      *dto.Reception
		wantErr    bool
		validateFn func(*testing.T, *dto.Reception)
	}

	testCases := []testCase{
		{
			name: "successful reception addition",
			input: &dto.Reception{
				PvzId: s.pvzID,
			},
			wantErr: false,
			validateFn: func(t *testing.T, r *dto.Reception) {
				assert.NotEqual(t, uuid.Nil, r.Id)
				assert.False(t, r.DateTime.IsZero())

				var count int
				err := s.repo.tx.QueryRowContext(
					s.ctx,
					"select count(*) from pvz_service.reception where reception_id = $1",
					r.Id,
				).Scan(&count)
				require.NoError(t, err)
				assert.Equal(t, 1, count)
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			err := s.repo.AddReception(s.ctx, tc.input)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tc.validateFn != nil {
				tc.validateFn(t, tc.input)
			}
		})
	}
}

func (s *ReceptionRepositoryTestSuite) TestGetLastReception() {

	receptionId := s.createReception(s.T())

	receptionIds := []uuid.UUID{
		receptionId,
	}
	for i := 0; i < 3; i++ {
		receptionId := uuid.New()
		_, err := s.repo.tx.ExecContext(s.ctx, `
		insert into pvz_service.reception (reception_id, started_at, pvz_id, status)
		values ($1, current_timestamp + interval '1 second' * $2, $3, 'in_progress')`,
			receptionId, i, s.pvzID)
		require.NoError(s.T(), err)
		receptionIds = append(receptionIds, receptionId)
	}

	s.T().Run("should return last added reception", func(t *testing.T) {
		result, err := s.repo.GetLastReceptionByPvzId(s.ctx, s.pvzID)
		require.NoError(t, err)
		assert.Equal(t, receptionIds[3], *result.Id)
	})

	s.T().Run("should return error for non-existent pvz", func(t *testing.T) {
		_, err := s.repo.GetLastReceptionByPvzId(s.ctx, uuid.New())
		require.Error(t, err)
		assert.Equal(t, models.ErrReceptionNotFound, err)
	})
}

func (s *ReceptionRepositoryTestSuite) TestCloseLastReception() {
	receptionID := uuid.New()
	_, err := s.repo.tx.ExecContext(s.ctx, `
insert into pvz_service.reception (reception_id, started_at, pvz_id, status)
values ($1, current_timestamp, $2, 'in_progress')`, receptionID, s.pvzID)
	require.NoError(s.T(), err)

	s.T().Run("successful close of reception", func(t *testing.T) {
		result, err := s.repo.CloseLastReception(s.ctx, receptionID)
		require.NoError(t, err)
		assert.Equal(t, dto.Close, result.Status)
	})

	s.T().Run("close non-existent reception", func(t *testing.T) {
		_, err := s.repo.CloseLastReception(s.ctx, uuid.New())
		require.Error(t, err)
		assert.Equal(t, models.ErrReceptionNotFound, err)
	})
}
