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
	q := psql.Insert("urls").Columns("short_url", "original_url", "created_at", "user_id", "is_deleted").
		Values(url.ShortURL, url.OriginalURL, createdAt, url.UserID, url.IsDeleted)
	sqlStr, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build insert: %w", err)
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	if err != nil {
		if errConflict := uniqueViolationError(err); errConflict != nil {
			return errConflict
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
		q := psql.Insert("urls").Columns("short_url", "original_url", "created_at", "user_id", "is_deleted").
			Values(u.ShortURL, u.OriginalURL, createdAt, u.UserID, u.IsDeleted)
		sqlStr, args, err := q.ToSql()
		if err != nil {
			return fmt.Errorf("build insert: %w", err)
		}
		_, err = tx.Exec(ctx, sqlStr, args...)
		if err != nil {
			if errConflict := uniqueViolationError(err); errConflict != nil {
				return errConflict
			}
			return fmt.Errorf("insert url: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (r *PostgresRepository) GetByShortURL(ctx context.Context, shortURL string) (*model.URL, error) {
	q := psql.Select("short_url", "original_url", "created_at", "user_id", "is_deleted").
		From("urls").Where(squirrel.Eq{"short_url": shortURL})
	sqlStr, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select: %w", err)
	}
	var u model.URL
	err = r.pool.QueryRow(ctx, sqlStr, args...).Scan(&u.ShortURL, &u.OriginalURL, &u.CreatedAt, &u.UserID, &u.IsDeleted)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get by short url: %w", err)
	}
	return &u, nil
}

func (r *PostgresRepository) GetByOriginalURL(ctx context.Context, originalURL string) (*model.URL, error) {
	q := psql.Select("short_url", "original_url", "created_at", "user_id", "is_deleted").
		From("urls").Where(squirrel.Eq{"original_url": originalURL}).Where(squirrel.Eq{"is_deleted": false})
	sqlStr, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select: %w", err)
	}
	var u model.URL
	err = r.pool.QueryRow(ctx, sqlStr, args...).Scan(&u.ShortURL, &u.OriginalURL, &u.CreatedAt, &u.UserID, &u.IsDeleted)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get by original url: %w", err)
	}
	return &u, nil
}

func (r *PostgresRepository) GetByUserID(ctx context.Context, userID string) ([]*model.URL, error) {
	q := psql.Select("short_url", "original_url", "created_at", "user_id", "is_deleted").
		From("urls").Where(squirrel.Eq{"user_id": userID}).Where(squirrel.Eq{"is_deleted": false})
	sqlStr, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select: %w", err)
	}
	rows, err := r.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("get by user id: %w", err)
	}
	defer rows.Close()
	var list []*model.URL
	for rows.Next() {
		var u model.URL
		if err := rows.Scan(&u.ShortURL, &u.OriginalURL, &u.CreatedAt, &u.UserID, &u.IsDeleted); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		list = append(list, &u)
	}
	return list, rows.Err()
}

func (r *PostgresRepository) DeleteByShortURLs(ctx context.Context, userID string, shortIDs []string) error {
	if len(shortIDs) == 0 {
		return nil
	}
	q := psql.Update("urls").Set("is_deleted", true).
		Where(squirrel.Eq{"user_id": userID}).Where(squirrel.Eq{"short_url": shortIDs})
	sqlStr, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build update: %w", err)
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("delete by short urls: %w", err)
	}
	return nil
}

const pgConstraintUniqueOriginalURL = "urls_original_url_not_deleted_key"

func uniqueViolationError(err error) error {
	var e *pgconn.PgError
	if !errors.As(err, &e) || e.Code != pgerrcode.UniqueViolation {
		return nil
	}
	if e.ConstraintName == pgConstraintUniqueOriginalURL {
		return ErrConflict
	}
	return ErrAlreadyExists
}
