package service

import (
	"context"
	"errors"
	"estimate/internal/entity"
	"estimate/internal/storage"
	"estimate/pkg/apperror"
	"estimate/pkg/cache"
	"estimate/pkg/worker"
	"github.com/corpix/uarand"
	"github.com/goware/urlx"
	"net/http"
	"time"
)

type WebsiteService interface {
	Watch(ctx context.Context, interval time.Duration) error
	Check(website entity.Website) (entity.Website, error)
	CheckByURL(rawURL string) (entity.Website, error)
	GetByURL(ctx context.Context, rawURL string) (entity.Website, error)
	Select(ctx context.Context) ([]entity.Website, error)
	Update(ctx context.Context, website entity.Website) error
	GetByMinAccessTime(ctx context.Context) (entity.Website, error)
	GetByMaxAccessTime(ctx context.Context) (entity.Website, error)
}

type websiteService struct {
	storage  storage.WebsiteStorage
	client   *http.Client
	cache    cache.Cache
	cacheTag string
}

func NewWebsiteService(
	storage storage.WebsiteStorage,
	cache cache.Cache,
	cacheTag string,
) WebsiteService {
	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return nil
		},
		Timeout: time.Second * 10,
	}

	return &websiteService{
		storage:  storage,
		client:   client,
		cache:    cache,
		cacheTag: cacheTag,
	}
}

func (service *websiteService) Watch(ctx context.Context, watchPeriod time.Duration) error {
	err := service.watch(ctx)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(watchPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err = service.watch(ctx)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (service *websiteService) watch(ctx context.Context) error {
	websites, err := service.Select(ctx)
	if err != nil && !errors.Is(err, apperror.NotFound) {
		return err
	}

	workerCount := 20

	pool := worker.NewPool(workerCount)

	jobs := make(chan worker.Job, workerCount)
	pool.AddJobs(jobs)

	go func() {
		defer close(jobs)

		for _, website := range websites {
			website := website

			jobs <- worker.Job{
				Fn: func(_ context.Context) (any, error) {
					updatedWebsite, err := service.Check(website)
					if err != nil {
						return nil, err
					}

					return updatedWebsite, nil
				},
			}
		}
	}()

	results := pool.Run(ctx)

	for result := range results {
		if err = result.Err; err != nil {
			return err
		}

		updatedWebsite := result.Value.(entity.Website)

		err = service.Update(ctx, updatedWebsite)
		if err != nil {
			return err
		}
	}

	err = service.cache.DelAll(ctx, service.cacheTag)
	if err != nil && !errors.Is(err, cache.Nil) {
		return err
	}

	return nil
}

// Check проверяет сайт и возвращает его обновленное состояние
func (service *websiteService) Check(website entity.Website) (entity.Website, error) {
	url, err := urlx.Parse(website.URL)
	if err != nil {
		return entity.Website{}, apperror.BadRequest.WithError(err)
	}
	url.Scheme = "https"

	request, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return entity.Website{}, apperror.Internal.WithError(err)
	}
	request.Header.Set("User-Agent", uarand.GetRandom())

	website.LastCheckAt = time.Now()

	response, err := service.client.Do(request)
	if err != nil {
		website.Available = false
	} else {
		response.Body.Close()

		website.Available = response.StatusCode == http.StatusOK
		if website.Available {
			website.AccessTime = time.Since(website.LastCheckAt)
		}
	}

	return website, nil
}

// CheckByURL проверяет сайт по ссылке и возвращает его обновленное состояние, возвращет ошибку, если сайт недоступен
func (service *websiteService) CheckByURL(rawURL string) (entity.Website, error) {
	website, err := service.Check(entity.Website{URL: rawURL})
	if err != nil {
		return entity.Website{}, err
	}

	return website, nil
}

func (service *websiteService) GetByURL(ctx context.Context, rawURL string) (entity.Website, error) {
	url, err := urlx.Parse(rawURL)
	if err != nil {
		return entity.Website{}, apperror.BadRequest.WithError(err)
	}

	var website entity.Website
	website, err = service.storage.GetByURL(ctx, url.Host)
	if err != nil {
		if !errors.Is(err, apperror.NotFound) {
			return entity.Website{}, err
		}

		website, err = service.CheckByURL(rawURL)
		if err != nil {
			return entity.Website{}, err
		}
	}

	if !website.Available {
		return entity.Website{}, apperror.Unavailable.WithMessage("website is unavailable")
	}

	return website, nil
}

func (service *websiteService) GetByMinAccessTime(ctx context.Context) (entity.Website, error) {
	website, err := service.storage.GetByMinAccessTime(ctx)
	if err != nil {
		if apperr, ok := apperror.Is(err, apperror.NotFound); ok {
			return entity.Website{}, apperr.WithMessage("website not found")
		}

		return entity.Website{}, err
	}

	return website, nil
}

func (service *websiteService) GetByMaxAccessTime(ctx context.Context) (entity.Website, error) {
	website, err := service.storage.GetByMaxAccessTime(ctx)
	if err != nil {
		if apperr, ok := apperror.Is(err, apperror.NotFound); ok {
			return entity.Website{}, apperr.WithMessage("website not found")
		}

		return entity.Website{}, err
	}

	return website, nil
}

func (service *websiteService) Select(ctx context.Context) ([]entity.Website, error) {
	websites, err := service.storage.Select(ctx)
	if err != nil {
		if apperr, ok := apperror.Is(err, apperror.NotFound); ok {
			return nil, apperr.WithMessage("websites not found")
		}

		return nil, err
	}

	return websites, nil
}

func (service *websiteService) Update(ctx context.Context, website entity.Website) error {
	err := service.storage.Update(ctx, website)
	if err != nil {
		return err
	}

	return nil
}
