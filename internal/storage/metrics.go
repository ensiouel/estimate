package storage

import (
	"context"
	"estimate/internal/entity"
	"github.com/redis/go-redis/v9"
	"strings"
)

type MetricsStorage interface {
	Metrics(ctx context.Context) (entity.Metrics, error)
}

type metricsStorage struct {
	client *redis.Client
	prefix string
}

func NewMetricsStorage(client *redis.Client) MetricsStorage {
	return &metricsStorage{
		client: client,
		prefix: "metrics:",
	}
}

func (storage *metricsStorage) Metrics(ctx context.Context) (entity.Metrics, error) {
	keys := storage.client.Keys(ctx, storage.prefix+"*").Val()

	metrics := make(entity.Metrics, len(keys))
	for i, key := range keys {
		count, err := storage.client.Get(ctx, key).Int()
		if err != nil {
			return nil, err
		}

		metrics[i] = entity.Metric{
			Endpoint: strings.TrimPrefix(key, storage.prefix),
			Count:    count,
		}
	}

	return metrics, nil
}
