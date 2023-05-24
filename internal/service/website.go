package service

import (
	"context"
	"errors"
	"github.com/corpix/uarand"
	"github.com/goware/urlx"
	"golang.org/x/sync/errgroup"
	"inspector/internal/entity"
	"inspector/internal/storage"
	"inspector/pkg/apperror"
	"inspector/pkg/cache"
	"net/http"
	"time"
)

type WebsiteService interface {
	RunEstimation(ctx context.Context) error
	Get(ctx context.Context, rawURL string) (entity.Website, error)
	Select(ctx context.Context) ([]entity.Website, error)
	Update(ctx context.Context, website entity.Website) error
	GetByMinAccessTime(ctx context.Context) (entity.Website, error)
	GetByMaxAccessTime(ctx context.Context) (entity.Website, error)
}

type websiteService struct {
	storage  storage.WebsiteStorage
	period   time.Duration
	cache    *cache.Cache
	cacheTag string
	client   *http.Client
}

func NewWebsiteService(storage storage.WebsiteStorage, period time.Duration, cache *cache.Cache, cacheTag string) WebsiteService {
	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return nil
		},
		Timeout: time.Second * 15,
	}

	return &websiteService{
		storage:  storage,
		period:   period,
		cache:    cache,
		cacheTag: cacheTag,
		client:   client,
	}
}

func (service *websiteService) RunEstimation(ctx context.Context) error {
	err := service.runEstimate(ctx)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(service.period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err = service.runEstimate(ctx)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (service *websiteService) runEstimate(ctx context.Context) error {
	websites, err := service.Select(ctx)
	if err != nil && !errors.Is(err, apperror.NotFound) {
		return err
	}

	g := new(errgroup.Group)
	for _, website := range websites {
		website := website
		g.Go(func() error {
			return service.pingWebsite(ctx, website)
		})
	}

	err = g.Wait()
	if err != nil {
		return err
	}

	err = service.cache.DelAll(ctx, service.cacheTag)
	if err != nil && !errors.Is(err, cache.Nil) {
		return err
	}

	return nil
}

func (service *websiteService) pingWebsite(ctx context.Context, website entity.Website) error {
	url, err := urlx.Parse(website.URL)
	if err != nil {
		return apperror.BadRequest.WithError(err)
	}
	url.Scheme = "https"

	now := time.Now()

	request, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", uarand.GetRandom())

	response, err := service.client.Do(request)
	if err != nil || response.StatusCode != http.StatusOK {
		website.Available = false
	} else {
		website.Available = true
	}
	website.AccessTime = time.Since(now)
	website.LastCheckAt = now.Add(website.AccessTime)

	err = service.Update(ctx, website)
	if err != nil {
		return err
	}

	return nil
}

func (service *websiteService) Get(ctx context.Context, rawURL string) (entity.Website, error) {
	url, err := urlx.Parse(rawURL)
	if err != nil {
		return entity.Website{}, apperror.BadRequest.WithError(err)
	}

	website, err := service.storage.GetByURL(ctx, url.Host)
	if err != nil {
		if apperr, ok := apperror.Is(err, apperror.NotFound); ok {
			return entity.Website{}, apperr.WithMessage("website not found")
		}

		return entity.Website{}, err
	}

	if !website.Available {
		return entity.Website{}, errors.New("website is unavailable")
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
