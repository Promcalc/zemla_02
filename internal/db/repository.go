// Package db предоставляет доступ к PostgreSQL с PostGIS.
package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Promcalc/zemla_02/internal/parser"
	"github.com/Promcalc/zemla_02/internal/rss"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewRepository(ctx context.Context, dbURL string, logger *slog.Logger) (*Repository, error) {
	// Настройка конфигурации пула подключений
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга URL подключения к БД: %w", err)
	}

	// Установка параметров пула подключений
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.MaxConnLifetimeJitter = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания пула подключений к БД: %w", err)
	}

	// Проверим подключение
	if err := pool.Ping(ctx); err != nil {
		pool.Close() // закрываем пул при ошибке
		return nil, fmt.Errorf("ошибка пинга БД: %w", err)
	}

	return &Repository{
		pool:   pool,
		logger: logger,
	}, nil
}

// SaveLotWithExternalData сохраняет лот и связанные данные атомарно (в одной транзакции).
func (r *Repository) SaveLotWithExternalData(
	ctx context.Context,
	rssLot rss.Lot,
	extracted parser.ExtractedData,
	lotInfo map[string]interface{},
	nspdData interface{},
	lotInfoErr, nspdErr error,
) (string, error) {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return "", fmt.Errorf("не удалось получить соединение: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	defer tx.Rollback(ctx) // откат при ошибке

	// 1. Сохраняем lot
	lotID, err := r.saveLot(ctx, tx, rssLot, extracted)
	if err != nil {
		return "", fmt.Errorf("ошибка сохранения лота: %w", err)
	}

	// 2. Сохраняем динамические поля
	if err := r.saveLotFields(ctx, tx, lotID, rssLot.Fields); err != nil {
		return "", fmt.Errorf("ошибка сохранения полей: %w", err)
	}

	// 3. Сохраняем внешние данные
	if err := r.saveExternalData(ctx, tx, lotID, lotInfo, nspdData, lotInfoErr, nspdErr); err != nil {
		return "", fmt.Errorf("ошибка сохранения external_data: %w", err)
	}

	// 4. Коммит транзакции
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("ошибка коммита транзакции: %w", err)
	}

	return lotID, nil
}

func (r *Repository) saveLot(ctx context.Context, tx pgx.Tx, rssLot rss.Lot, extracted parser.ExtractedData) (string, error) {
	var lotID string
	query := `
		INSERT INTO lots (guid, link, title, pub_date, dc_date, auction_date, cadastral_number)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	err := tx.QueryRow(ctx, query,
		rssLot.GUID,
		rssLot.Link,
		rssLot.Title,
		rssLot.PubDate,
		rssLot.DCDate,
		extracted.AuctionDate,
		extracted.CadastralNumber,
	).Scan(&lotID)

	if err != nil {
		// Обработка дубликатов GUID
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			// Повторяющийся GUID — лот уже существует
			r.logger.Debug("Лот с таким GUID уже существует", "guid", rssLot.GUID)
			// Вернём существующий ID
			err = tx.QueryRow(ctx, "SELECT id FROM lots WHERE guid = $1", rssLot.GUID).Scan(&lotID)
			if err != nil {
				return "", fmt.Errorf("не удалось найти существующий лот: %w", err)
			}
			return lotID, nil
		}
		return "", err
	}

	r.logger.Debug("Лот сохранён", "id", lotID, "guid", rssLot.GUID)
	return lotID, nil
}

func (r *Repository) saveLotFields(ctx context.Context, tx pgx.Tx, lotID string, fields map[string]string) error {
	if len(fields) == 0 {
		return nil
	}

	query := "INSERT INTO lot_fields (lot_id, field_name, field_value) VALUES ($1, $2, $3)"
	for name, value := range fields {
		_, err := tx.Exec(ctx, query, lotID, name, value)
		if err != nil {
			return fmt.Errorf("ошибка сохранения поля %q: %w", name, err)
		}
	}
	return nil
}

func (r *Repository) saveExternalData(
	ctx context.Context,
	tx pgx.Tx,
	lotID string,
	lotInfo map[string]interface{},
	nspdData interface{},
	lotInfoErr, nspdErr error,
) error {
	now := time.Now()
	var lotInfoFetched, nspdFetched *time.Time
	var lotInfoJSON, nspdJSON JSONB

	if lotInfo != nil {
		lotInfoFetched = &now
		lotInfoJSON = JSONB{Data: lotInfo}
	}
	if nspdData != nil {
		nspdFetched = &now
		nspdJSON = JSONB{Data: nspdData}
	}

	var lotInfoErrStr, nspdErrStr *string
	if lotInfoErr != nil {
		s := lotInfoErr.Error()
		lotInfoErrStr = &s
	}
	if nspdErr != nil {
		s := nspdErr.Error()
		nspdErrStr = &s
	}

	query := `
		INSERT INTO external_data (
			lot_id, lot_info, lot_info_fetched_at, lot_info_error,
			nspd_data, nspd_fetched_at, nspd_error
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (lot_id) DO UPDATE SET
			lot_info = EXCLUDED.lot_info,
			lot_info_fetched_at = EXCLUDED.lot_info_fetched_at,
			lot_info_error = EXCLUDED.lot_info_error,
			nspd_data = EXCLUDED.nspd_data,
			nspd_fetched_at = EXCLUDED.nspd_fetched_at,
			nspd_error = EXCLUDED.nspd_error,
			updated_at = NOW()
	`

	_, err := tx.Exec(ctx, query,
		lotID, lotInfoJSON, lotInfoFetched, lotInfoErrStr,
		nspdJSON, nspdFetched, nspdErrStr,
	)
	return err
}

// Close закрывает пул подключений к базе данных.
func (r *Repository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}

// GetLastPubDate возвращает самую свежую pub_date из базы
func (r *Repository) GetLastPubDate(ctx context.Context) (time.Time, error) {
	var lastPubDate time.Time
	err := r.pool.QueryRow(ctx, "SELECT MAX(pub_date) FROM lots").Scan(&lastPubDate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, nil // таблица пуста
		}
		return time.Time{}, fmt.Errorf("ошибка получения last_pub_date: %w", err)
	}
	return lastPubDate, nil
}

// UpdateExternalData обновляет внешние данные для конкретного лота
func (r *Repository) UpdateExternalData(
	ctx context.Context,
	lotID string,
	lotInfo map[string]interface{},
	nspdData interface{},
	lotInfoErr, nspdErr error,
) error {
	now := time.Now()
	var lotInfoFetched, nspdFetched *time.Time
	var lotInfoJSON, nspdJSON JSONB

	if lotInfo != nil {
		lotInfoFetched = &now
		lotInfoJSON = JSONB{Data: lotInfo}
	}
	if nspdData != nil {
		nspdFetched = &now
		nspdJSON = JSONB{Data: nspdData}
	}

	var lotInfoErrStr, nspdErrStr *string
	if lotInfoErr != nil {
		s := lotInfoErr.Error()
		lotInfoErrStr = &s
	}
	if nspdErr != nil {
		s := nspdErr.Error()
		nspdErrStr = &s
	}

	query := `
		UPDATE external_data SET
			lot_info = $2,
			lot_info_fetched_at = $3,
			lot_info_error = $4,
			nspd_data = $5,
			nspd_fetched_at = $6,
			nspd_error = $7,
			updated_at = NOW()
		WHERE lot_id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		lotID, lotInfoJSON, lotInfoFetched, lotInfoErrStr,
		nspdJSON, nspdFetched, nspdErrStr,
	)
	return err
}

// Pool возвращает пул подключений к базе данных
func (r *Repository) Pool() *pgxpool.Pool {
	return r.pool
}
