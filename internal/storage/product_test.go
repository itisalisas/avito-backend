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

type ProductRepositoryTestSuite struct {
	suite.Suite
	db      *sql.DB
	cleanup func()
	repo    *ProductRepository
	ctx     context.Context
	pvzID   uuid.UUID
}

func TestProductRepositorySuite(t *testing.T) {
	suite.Run(t, new(ProductRepositoryTestSuite))
}

func (s *ProductRepositoryTestSuite) SetupSuite() {
	s.ctx = context.Background()
	db := DBTestSetup()
	log.Println("migrations applied")
	s.db = db

	s.repo = NewProductRepository(s.db)
}

func (s *ProductRepositoryTestSuite) TearDownSuite() {
	err := s.db.Close()
	if err != nil {
		log.Fatalf("failed to close database connection: %v", err)
	}
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *ProductRepositoryTestSuite) SetupTest() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	require.NoError(s.T(), err)
	s.repo.tx = tx
}

func (s *ProductRepositoryTestSuite) TearDownTest() {
	if s.repo.tx != nil {
		err := s.repo.tx.Rollback()
		require.NoError(s.T(), err)
	}
}

func (s *ProductRepositoryTestSuite) createPVZ(t *testing.T) uuid.UUID {
	pvzID := uuid.New()
	_, err := s.repo.tx.ExecContext(s.ctx, `
        insert into pvz_service.pvz (pvz_id, registration_date, city)
        values ($1, current_date, 'Москва')`, pvzID)
	require.NoError(t, err)
	return pvzID
}

func (s *ProductRepositoryTestSuite) createReception(t *testing.T) uuid.UUID {
	receptionID := uuid.New()
	_, err := s.repo.tx.ExecContext(s.ctx, `
		insert into pvz_service.reception (reception_id, started_at, pvz_id, status)
		values ($1, current_timestamp, $2, 'in_progress')`,
		receptionID, s.pvzID,
	)
	require.NoError(t, err)
	return receptionID
}

func (s *ProductRepositoryTestSuite) TestAddProduct() {
	type testCase struct {
		name       string
		input      *dto.Product
		wantErr    bool
		validateFn func(*testing.T, *dto.Product)
	}

	s.pvzID = s.createPVZ(s.T())
	receptionID := s.createReception(s.T())

	testCases := []testCase{
		{
			name: "successful product addition",
			input: &dto.Product{
				Type:        "Electronics",
				ReceptionId: receptionID,
			},
			wantErr: false,
			validateFn: func(t *testing.T, p *dto.Product) {
				assert.NotEqual(t, uuid.Nil, p.Id)
				assert.False(t, p.DateTime.IsZero())

				var count int
				err := s.repo.tx.QueryRowContext(
					s.ctx,
					"select count(*) from pvz_service.product where product_id = $1",
					p.Id,
				).Scan(&count)
				require.NoError(t, err)
				assert.Equal(t, 1, count)
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			err := s.repo.AddProduct(s.ctx, tc.input)
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

func (s *ProductRepositoryTestSuite) TestGetLastProduct() {
	s.pvzID = s.createPVZ(s.T())
	receptionID := s.createReception(s.T())

	types := []dto.ProductType{
		dto.ProductTypeОбувь,
		dto.ProductTypeОдежда,
		dto.ProductTypeЭлектроника,
	}

	for i, ptype := range types {
		date := time.Now().Add(time.Duration(i) * time.Second)
		productID := uuid.New()
		_, err := s.repo.tx.ExecContext(s.ctx, `
insert into pvz_service.product (product_id, product_type, reception_id, added_at)
values ($1, $2, $3, $4)
		`, productID, ptype, receptionID, date)
		require.NoError(s.T(), err)
	}

	s.T().Run("should return last added product", func(t *testing.T) {
		result, err := s.repo.GetLastProduct(s.ctx, receptionID)
		require.NoError(t, err)
		assert.Equal(t, dto.ProductTypeЭлектроника, result.Type)
	})

	s.T().Run("should return error for non-existent reception", func(t *testing.T) {
		_, err := s.repo.GetLastProduct(s.ctx, uuid.New())
		require.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
	})
}

func (s *ProductRepositoryTestSuite) TestDeleteProductById() {
	s.pvzID = s.createPVZ(s.T())
	receptionID := s.createReception(s.T())

	p := &dto.Product{
		Type:        dto.ProductTypeОдежда,
		ReceptionId: receptionID,
	}
	err := s.repo.AddProduct(s.ctx, p)
	require.NoError(s.T(), err)

	s.T().Run("successful deletion", func(t *testing.T) {
		err := s.repo.DeleteProductById(s.ctx, *p.Id)
		require.NoError(t, err)

		var exists bool
		err = s.repo.tx.QueryRowContext(
			s.ctx,
			"select exists(select 1 from pvz_service.product where product_id = $1)",
			p.Id,
		).Scan(&exists)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	s.T().Run("delete non-existent product", func(t *testing.T) {
		err := s.repo.DeleteProductById(s.ctx, uuid.New())
		require.NoError(t, err)
	})
}
