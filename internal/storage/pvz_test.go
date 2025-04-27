package storage

import (
	"context"
	"database/sql"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
)

type PvzRepositoryTestSuite struct {
	suite.Suite
	db      *sql.DB
	cleanup func()
	repo    *PvzRepository
	ctx     context.Context
	pvzID   uuid.UUID
}

func TestPvzRepositorySuite(t *testing.T) {
	suite.Run(t, new(PvzRepositoryTestSuite))
}

func (s *PvzRepositoryTestSuite) SetupSuite() {
	s.ctx = context.Background()
	db := DBTestSetup()
	log.Println("migrations applied")
	s.db = db

	tx, err := s.db.BeginTx(s.ctx, nil)
	require.NoError(s.T(), err)

	s.repo = NewPvzRepository(s.db)
	s.repo.tx = tx

	pvzID := uuid.New()
	_, err = s.repo.tx.ExecContext(s.ctx, `
        insert into pvz_service.pvz (pvz_id, registration_date, city)
        values ($1, current_date, 'Москва')`, pvzID)
	require.NoError(s.T(), err)
	s.pvzID = pvzID
}

func (s *PvzRepositoryTestSuite) TearDownSuite() {
	err := s.db.Close()
	if err != nil {
		log.Fatalf("failed to close database connection: %v", err)
	}
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *PvzRepositoryTestSuite) SetupTest() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	require.NoError(s.T(), err)
	s.repo.tx = tx
}

func (s *PvzRepositoryTestSuite) TearDownTest() {
	if s.repo.tx != nil {
		err := s.repo.tx.Rollback()
		require.NoError(s.T(), err)
	}
}

func (s *PvzRepositoryTestSuite) TestCreatePvz() {
	type testCase struct {
		name       string
		input      *dto.PVZ
		wantErr    bool
		validateFn func(*testing.T, *dto.PVZ)
	}

	testCases := []testCase{
		{
			name: "successful PVZ creation",
			input: &dto.PVZ{
				City: "Москва",
			},
			wantErr: false,
			validateFn: func(t *testing.T, pvz *dto.PVZ) {
				assert.NotEqual(t, uuid.Nil, pvz.Id)
				assert.False(t, pvz.RegistrationDate.IsZero())

				var count int
				err := s.repo.tx.QueryRowContext(
					s.ctx,
					"select count(*) from pvz_service.pvz where pvz_id = $1",
					pvz.Id,
				).Scan(&count)
				require.NoError(t, err)
				assert.Equal(t, 1, count)
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			err := s.repo.CreatePvz(s.ctx, tc.input)
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

func (s *PvzRepositoryTestSuite) TestGetPvzList() {
	const timeLayout = "2006-01-02 15:04:05"

	pvzID1 := uuid.New()
	pvzID2 := uuid.New()

	reg1, _ := time.Parse(timeLayout, "2025-01-01 00:00:00")
	reg2, _ := time.Parse(timeLayout, "2025-01-02 00:00:00")

	_, err := s.repo.tx.ExecContext(s.ctx, `
		insert into pvz_service.pvz (pvz_id, registration_date, city)
		values ($1, $2, $3)
	`, pvzID1, reg1, "Москва")
	require.NoError(s.T(), err)

	_, err = s.repo.tx.ExecContext(s.ctx, `
		insert into pvz_service.pvz (pvz_id, registration_date, city)
		values ($1, $2, $3)
	`, pvzID2, reg2, "Санкт-Петербург")
	require.NoError(s.T(), err)

	rec1a := uuid.New()
	rec1b := uuid.New()
	rec2 := uuid.New()

	start1a, _ := time.Parse(timeLayout, "2025-01-01 10:00:00")
	start1b, _ := time.Parse(timeLayout, "2025-01-02 11:00:00")
	start2, _ := time.Parse(timeLayout, "2025-01-03 12:00:00")

	for _, data := range []struct {
		id     uuid.UUID
		pvz    uuid.UUID
		start  time.Time
		status string
	}{
		{rec1a, pvzID1, start1a, "in_progress"},
		{rec1b, pvzID1, start1b, "close"},
		{rec2, pvzID2, start2, "in_progress"},
	} {
		_, err = s.repo.tx.ExecContext(s.ctx, `
			insert into pvz_service.reception (reception_id, pvz_id, started_at, status)
			values ($1, $2, $3, $4)
		`, data.id, data.pvz, data.start, data.status)
		require.NoError(s.T(), err)
	}

	prod1 := uuid.New()
	prod2 := uuid.New()
	prod3 := uuid.New()

	add1, _ := time.Parse(timeLayout, "2025-01-01 10:30:00")
	add2, _ := time.Parse(timeLayout, "2025-01-01 11:00:00")
	add3, _ := time.Parse(timeLayout, "2025-01-02 11:30:00")

	for _, data := range []struct {
		id    uuid.UUID
		rec   uuid.UUID
		at    time.Time
		ptype string
	}{
		{prod1, rec1a, add1, "TypeA"},
		{prod2, rec1a, add2, "TypeB"},
		{prod3, rec1b, add3, "TypeC"},
	} {
		_, err = s.repo.tx.ExecContext(s.ctx, `
			insert into pvz_service.product (product_id, reception_id, added_at, product_type)
			values ($1, $2, $3, $4)
		`, data.id, data.rec, data.at, data.ptype)
		require.NoError(s.T(), err)
	}

	startF, _ := time.Parse(timeLayout, "2025-01-01 00:00:00")
	endF, _ := time.Parse(timeLayout, "2025-01-02 12:00:00")
	s.Run("time filter only pvz1", func() {
		list, err := s.repo.GetPvzList(s.ctx, &startF, &endF, 1, 10)
		require.NoError(s.T(), err)

		require.Len(s.T(), list, 1)
		assert.Equal(s.T(), pvzID1, *list[0].PVZ.Id)
		assert.Len(s.T(), list[0].Receptions, 2)
	})

	s.Run("pagination page2 limit1 gives second pvz", func() {
		list, err := s.repo.GetPvzList(s.ctx, nil, nil, 2, 1)
		require.NoError(s.T(), err)

		require.Len(s.T(), list, 1)
		assert.Equal(s.T(), pvzID1, *list[0].PVZ.Id)
	})
}
