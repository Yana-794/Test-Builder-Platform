package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool создаёт пул соединений с PostgreSQL, настраивает его параметры
// и проверяет, что база доступна (fail fast при старте).
//
// Возвращает готовый пул или ошибку. Если ошибка — пул уже закрыт,
// вызывающему не нужно ничего чистить
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	
	cfg.MaxConns = 10                                // максимум одновременных соединений
	cfg.MinConns = 0                                 // пул может полностью опустеть под нулевой нагрузкой
	cfg.MaxConnLifetime = time.Hour                  // соединение живёт не больше часа
	cfg.MaxConnIdleTime = 30 * time.Minute           // простаивающее соединение закрывается через 30 мин
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second  // таймаут на создание одного соединения

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect to PostgreSQL: %w", err)
	}

	return pool, nil
}