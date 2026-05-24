package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Irongoshan-ux/url-shortener/internal/model"
	"github.com/Irongoshan-ux/url-shortener/internal/repository/sqlc/db"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const pgConstraintUniqueOriginalURL = "urls_original_url_not_deleted_key"

// PostgresRepository implements storage using pgxpool and sqlc-generated queries.
type PostgresRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

// NewPostgresRepository wraps a connection pool; the pool must remain open for the repository lifetime.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool: pool,
		q:    db.New(pool),
	}
}

func toModel(u db.Url) *model.URL {
	return &model.URL{
		ShortURL:    u.ShortUrl,
		OriginalURL: u.OriginalUrl,
		CreatedAt:   u.CreatedAt,
		UserID:      u.UserID,
		IsDeleted:   u.IsDeleted,
	}
}

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

func (r *PostgresRepository) Create(ctx context.Context, url *model.URL) error {
	createdAt := url.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	err := r.q.CreateUrl(ctx, db.CreateUrlParams{
		ShortUrl:    url.ShortURL,
		OriginalUrl: url.OriginalURL,
		CreatedAt:   createdAt,
		UserID:      url.UserID,
		IsDeleted:   url.IsDeleted,
	})
	if err != nil {
		if errConflict := uniqueViolationError(err); errConflict != nil {
			return errConflict
		}
		return fmt.Errorf("insert url: %w", err)
	}
	return nil
}

func (r *PostgresRepository) CreateBatch(ctx context.Context, urls []*model.URL) error {
	if len(urls) == 0 {
		return nil
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	q := db.New(tx)
	params := make([]db.CreateUrlBatchParams, len(urls))
	for i, u := range urls {
		createdAt := u.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now()
		}
		params[i] = db.CreateUrlBatchParams{
			ShortUrl:    u.ShortURL,
			OriginalUrl: u.OriginalURL,
			CreatedAt:   createdAt,
			UserID:      u.UserID,
			IsDeleted:   u.IsDeleted,
		}
	}
	br := q.CreateUrlBatch(ctx, params)
	var firstErr error
	br.Exec(func(_ int, err error) {
		if err != nil && firstErr == nil {
			firstErr = err
		}
	})
	if firstErr != nil {
		if errConflict := uniqueViolationError(firstErr); errConflict != nil {
			return errConflict
		}
		return fmt.Errorf("insert batch: %w", firstErr)
	}
	return tx.Commit(ctx)
}

func (r *PostgresRepository) GetByShortURL(ctx context.Context, shortURL string) (*model.URL, error) {
	u, err := r.q.GetByShortURL(ctx, shortURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get by short url: %w", err)
	}
	return toModel(u), nil
}

func (r *PostgresRepository) GetByOriginalURL(ctx context.Context, originalURL string) (*model.URL, error) {
	u, err := r.q.GetByOriginalURL(ctx, originalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get by original url: %w", err)
	}
	return toModel(u), nil
}

func (r *PostgresRepository) GetByUserID(ctx context.Context, userID string) ([]*model.URL, error) {
	list, err := r.q.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get by user id: %w", err)
	}
	out := make([]*model.URL, len(list))
	for i := range list {
		out[i] = toModel(list[i])
	}
	return out, nil
}

func (r *PostgresRepository) DeleteByShortURLs(ctx context.Context, userID string, shortIDs []string) error {
	if len(shortIDs) == 0 {
		return nil
	}
	err := r.q.MarkDeletedByShortURLs(ctx, db.MarkDeletedByShortURLsParams{
		UserID:  userID,
		Column2: shortIDs,
	})
	if err != nil {
		return fmt.Errorf("delete by short urls: %w", err)
	}
	return nil
}
