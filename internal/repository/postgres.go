package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Irongoshan-ux/url-shortener/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, url *model.URL) error {
	createdAt := url.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO urls (short_url, original_url, created_at) VALUES ($1, $2, $3)`,
		url.ShortURL, url.OriginalURL, createdAt)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyExists
		}
		return fmt.Errorf("insert url: %w", err)
	}
	return nil
}

func (r *PostgresRepository) CreateBatch(ctx context.Context, urls []*model.URL) error {
	if len(urls) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	for _, u := range urls {
		createdAt := u.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now()
		}
		_, err := tx.ExecContext(ctx,
			`INSERT INTO urls (short_url, original_url, created_at) VALUES ($1, $2, $3)`,
			u.ShortURL, u.OriginalURL, createdAt)
		if err != nil {
			if isUniqueViolation(err) {
				return ErrAlreadyExists
			}
			return fmt.Errorf("insert url: %w", err)
		}
	}
	return tx.Commit()
}

func (r *PostgresRepository) GetByShortURL(ctx context.Context, shortURL string) (*model.URL, error) {
	var u model.URL
	err := r.db.QueryRowContext(ctx,
		`SELECT short_url, original_url, created_at FROM urls WHERE short_url = $1`,
		shortURL).Scan(&u.ShortURL, &u.OriginalURL, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get by short url: %w", err)
	}
	return &u, nil
}

func (r *PostgresRepository) GetByOriginalURL(ctx context.Context, originalURL string) (*model.URL, error) {
	var u model.URL
	err := r.db.QueryRowContext(ctx,
		`SELECT short_url, original_url, created_at FROM urls WHERE original_url = $1`,
		originalURL).Scan(&u.ShortURL, &u.OriginalURL, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get by original url: %w", err)
	}
	return &u, nil
}

func isUniqueViolation(err error) bool {
	var e *pq.Error
	if errors.As(err, &e) {
		return e.Code == pgerrcode.UniqueViolation
	}
	return false
}
