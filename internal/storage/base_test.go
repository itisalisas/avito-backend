package storage

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseRepository(t *testing.T) {
	t.Run("should begin transaction successfully", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func(db *sql.DB) {
			err := db.Close()
			require.NoError(t, err)
		}(db)

		repo := NewBaseRepository(db)

		mock.ExpectBegin()

		tx, err := repo.BeginTx(context.Background(), nil)
		require.NoError(t, err)
		assert.NotNil(t, tx)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("should return error when BeginTx fails", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func(db *sql.DB) {
			_ = db.Close()
		}(db)

		repo := NewBaseRepository(db)

		mock.ExpectBegin().WillReturnError(fmt.Errorf("transaction error"))

		tx, err := repo.BeginTx(context.Background(), nil)

		assert.Error(t, err)
		assert.Nil(t, tx)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("should commit transaction successfully", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func(db *sql.DB) {
			_ = db.Close()
		}(db)

		repo := NewBaseRepository(db)

		mock.ExpectBegin()
		mock.ExpectCommit()

		_, err = repo.BeginTx(context.Background(), nil)
		require.NoError(t, err)

		err = repo.Commit()

		require.NoError(t, err)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("should return error when commit fails", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func(db *sql.DB) {
			_ = db.Close()
		}(db)

		repo := NewBaseRepository(db)

		mock.ExpectBegin()
		mock.ExpectCommit().WillReturnError(fmt.Errorf("commit error"))

		_, err = repo.BeginTx(context.Background(), nil)
		require.NoError(t, err)

		err = repo.Commit()

		assert.Error(t, err)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("should rollback transaction successfully", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func(db *sql.DB) {
			_ = db.Close()
		}(db)

		repo := NewBaseRepository(db)

		mock.ExpectBegin()
		mock.ExpectRollback()

		_, err = repo.BeginTx(context.Background(), nil)
		require.NoError(t, err)

		err = repo.Rollback()

		require.NoError(t, err)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("should return error when rollback fails", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func(db *sql.DB) {
			_ = db.Close()
		}(db)

		repo := NewBaseRepository(db)

		mock.ExpectBegin()
		mock.ExpectRollback().WillReturnError(fmt.Errorf("rollback error"))

		_, err = repo.BeginTx(context.Background(), nil)
		require.NoError(t, err)

		err = repo.Rollback()

		assert.Error(t, err)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})
}
