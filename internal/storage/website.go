package storage

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"inspector/internal/entity"
	"inspector/pkg/apperror"
	"inspector/pkg/postgres"
)

type WebsiteStorage interface {
	GetByURL(ctx context.Context, rawURL string) (entity.Website, error)
	Update(ctx context.Context, website entity.Website) error
	GetByMinAccessTime(ctx context.Context) (entity.Website, error)
	GetByMaxAccessTime(ctx context.Context) (entity.Website, error)
	Select(ctx context.Context) ([]entity.Website, error)
}

type websiteStorage struct {
	client postgres.Client
}

func NewWebsiteStorage(client postgres.Client) WebsiteStorage {
	return &websiteStorage{client: client}
}

func (storage *websiteStorage) GetByURL(ctx context.Context, rawURL string) (entity.Website, error) {
	q := `
SELECT url, last_check_at, access_time, available
FROM website
WHERE url = $1
`

	var website entity.Website
	err := storage.client.Get(ctx, &website, q, rawURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Website{}, apperror.NotFound.WithError(err)
		}

		return entity.Website{}, apperror.Internal.WithError(err)
	}

	return website, nil
}

func (storage *websiteStorage) Update(ctx context.Context, website entity.Website) error {
	q := `
UPDATE website
SET last_check_at = $1,
    access_time = $2,
    available   = $3
WHERE url = $4
`

	_, err := storage.client.Exec(ctx, q, website.LastCheckAt, website.AccessTime, website.Available, website.URL)
	if err != nil {
		return apperror.Internal.WithError(err)
	}

	return nil
}

func (storage *websiteStorage) GetByMinAccessTime(ctx context.Context) (entity.Website, error) {
	q := `
SELECT url,
       last_check_at,
       access_time,
       available
FROM website
WHERE available = true
ORDER BY access_time
LIMIT 1
`

	var website entity.Website
	err := storage.client.Get(ctx, &website, q)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Website{}, apperror.NotFound.WithError(err)
		}

		return entity.Website{}, apperror.Internal.WithError(err)
	}

	return website, nil
}

func (storage *websiteStorage) GetByMaxAccessTime(ctx context.Context) (entity.Website, error) {
	q := `
SELECT url,
       last_check_at,
       access_time,
       available
FROM website
WHERE available = true
ORDER BY access_time DESC
LIMIT 1
`

	var website entity.Website
	err := storage.client.Get(ctx, &website, q)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Website{}, apperror.NotFound.WithError(err)
		}

		return entity.Website{}, apperror.Internal.WithError(err)
	}

	return website, nil
}

func (storage *websiteStorage) Select(ctx context.Context) ([]entity.Website, error) {
	q := `
SELECT url,
       last_check_at,
       access_time,
       available
FROM website
`

	var websites []entity.Website
	err := storage.client.Select(ctx, &websites, q)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound.WithError(err)
		}

		return nil, apperror.Internal.WithError(err)
	}

	return websites, nil
}
