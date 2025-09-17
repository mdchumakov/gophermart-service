package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

const DefaultPostgresDSN = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

type PGSettings struct {
	URI            string `envconfig:"DATABASE_URI"`
	MigrationsPath string `envconfig:"MIGRATIONS_PATH" default:"migrations"`
}

func SetupDB(ctx context.Context, logger LoggerInterface, pgDSN, migrationsPath string) *pgxpool.Pool {
	pool, err := InitPostgresDB(ctx, pgDSN)
	if err != nil {
		logger.Fatal(err)
		panic(err)
	}
	logger.Debug("connected to postgres database")

	if err := RunMigrations(logger, pool, migrationsPath); err != nil {
		logger.Fatal(err)
		panic(err)
	}
	logger.Debug("migrations applied successfully")

	return pool
}

func InitPostgresDB(ctx context.Context, pgDSN string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(pgDSN)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 30                      // Максимальное количество соединений
	config.MinConns = 5                       // Минимальное количество соединений
	config.MaxConnLifetime = time.Hour        // Время жизни соединения
	config.MaxConnIdleTime = time.Minute * 30 // Время простоя соединения

	return pgxpool.NewWithConfig(ctx, config)
}

func RunMigrations(logger LoggerInterface, pool *pgxpool.Pool, migrationsPath string) error {
	conn := stdlib.OpenDBFromPool(pool)
	defer func(conn *sql.DB) {
		if err := conn.Close(); err != nil {
			panic("failed to close database connection: " + err.Error())
		}
	}(conn)
	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer func(m *migrate.Migrate) {
		err, _ := m.Close()
		if err != nil {
			logger.Fatal("failed to close migrate instance")
			panic("failed to close migrate instance: " + err.Error())
		}
	}(m)

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Errorw("failed to run migrations")
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
