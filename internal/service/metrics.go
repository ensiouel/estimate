package service

import (
	"context"
	"inspector/internal/entity"
	"inspector/internal/storage"
)

type MetricsService interface {
	Metrics(ctx context.Context) (entity.Metrics, error)
}

type metricsService struct {
	storage storage.MetricsStorage
}

func NewMetricsService(storage storage.MetricsStorage) MetricsService {
	return &metricsService{
		storage: storage,
	}
}

func (service *metricsService) Metrics(ctx context.Context) (entity.Metrics, error) {
	metrics, err := service.storage.Metrics(ctx)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}
