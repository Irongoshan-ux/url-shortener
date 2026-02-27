package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Irongoshan-ux/url-shortener/internal/model"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgerrcode"
)

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, url *model.URL) error {
	createdAt := url.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	q := psql.Insert("urls").Columns("short_url", "original_url", "created_at").
		Values(url.ShortURL, url.OriginalURL, createdAt)
	sqlStr, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build insert: %w", err)
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyExists
		}
		return fmt.Errorf("insert url: %w", err)
	}
	return nil
}

func (r *PostgresRepository) CreateBatch(ctx context.Context, urls []*model.URL) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	for _, u := range urls {
		createdAt := u.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now()
		}
		q := psql.Insert("urls").Columns("short_url", "original_url", "created_at").
			Values(u.ShortURL, u.OriginalURL, createdAt)
		sqlStr, args, err := q.ToSql()
		if err != nil {
			return fmt.Errorf("build insert: %w", err)
		}
		_, err = tx.Exec(ctx, sqlStr, args...)
		if err != nil {
			if isUniqueViolation(err) {
				return ErrAlreadyExists
			}
			return fmt.Errorf("insert url: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (r *PostgresRepository) GetByShortURL(ctx context.Context, shortURL string) (*model.URL, error) {
	q := psql.Select("short_url", "original_url", "created_at").
		From("urls").Where(squirrel.Eq{"short_url": shortURL})
	sqlStr, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select: %w", err)
	}
	var u model.URL
	err = r.pool.QueryRow(ctx, sqlStr, args...).Scan(&u.ShortURL, &u.OriginalURL, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get by short url: %w", err)
	}
	return &u, nil
}

func (r *PostgresRepository) GetByOriginalURL(ctx context.Context, originalURL string) (*model.URL, error) {
	q := psql.Select("short_url", "original_url", "created_at").
		From("urls").Where(squirrel.Eq{"original_url": originalURL})
	sqlStr, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select: %w", err)
	}
	var u model.URL
	err = r.pool.QueryRow(ctx, sqlStr, args...).Scan(&u.ShortURL, &u.OriginalURL, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get by original url: %w", err)
	}
	return &u, nil
}

func isUniqueViolation(err error) bool {
	var e *pgconn.PgError
	if errors.As(err, &e) {
		return e.Code == pgerrcode.UniqueViolation
	}
	return false
}
