package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	_ "github.com/mattes/migrate/source/file"
)

type DB struct {
	*sql.DB
}

func NewPostgres(host string, port string, usr string, password string, dbName string) (*DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, usr, password, dbName)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) Migrate(migrationsPath string) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize migrate: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Println("migrate ok")
	return nil
}

func DBTestSetup() *sql.DB {
	path, _ := filepath.Abs("../../.env")
	if err := godotenv.Load(path); err != nil {
		log.Println("Error loading .env file: " + err.Error())
		return nil
	}
	pg, err := NewPostgres(os.Getenv("DB_HOST_TEST"),
		os.Getenv("DB_PORT_TEST"),
		os.Getenv("DB_USER_TEST"),
		os.Getenv("DB_PASSWORD_TEST"),
		os.Getenv("DB_NAME_TEST"))
	if err != nil {
		log.Println("can't connect: " + err.Error())
		return nil
	}

	err = pg.Migrate("file://../../migrations")
	if err != nil {
		log.Println("can't migrate: " + err.Error())
		return nil
	}
	return pg.DB
}
