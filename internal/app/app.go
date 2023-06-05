package app

import (
	"context"
	"errors"
	"estimate/internal/config"
	"estimate/internal/service"
	"estimate/internal/storage"
	"estimate/internal/transport/rest"
	"estimate/internal/transport/rest/handler"
	loggerpkg "estimate/pkg/logger"
	"estimate/pkg/postgres"
	"github.com/alejandro-carstens/gocache"
	"github.com/alejandro-carstens/gocache/encoder"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"net/http"
	"os/signal"
	"syscall"
)

type App struct {
	conf config.Config
}

func New() *App {
	conf := config.New()

	return &App{
		conf: conf,
	}
}

func (app *App) Run() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := loggerpkg.New(app.conf.LogLevel)

	logger.Info("starting app")

	logger.Info("connecting to postgres")
	pgClient, err := postgres.NewClient(ctx, postgres.Config{
		Host:     app.conf.Postgres.Host,
		Port:     app.conf.Postgres.Port,
		User:     app.conf.Postgres.User,
		Password: app.conf.Postgres.Password,
		DB:       app.conf.Postgres.DB,
	})
	if err != nil {
		logger.Fatal("failed to connect to postgres", zap.Error(err))
	}

	logger.Info("connecting to redis")
	redisClient := redis.NewClient(&redis.Options{
		Addr: app.conf.Redis.Addr,
	})
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		logger.Fatal("failed to connect to redis", zap.Error(err))
	}

	cache, err := gocache.New(&gocache.RedisConfig{
		Prefix: "gocache:",
		Addr:   app.conf.Redis.Addr,
	}, encoder.JSON{})
	if err != nil {
		logger.Fatal("failed to connect to redis cache", zap.Error(err))
	}

	estimateCache := cache.Tags("estimate")

	websiteStorage := storage.NewWebsiteStorage(pgClient)
	websiteService := service.NewWebsiteService(websiteStorage, estimateCache)

	logger.Info("starting estimation service")
	go func() {
		err = websiteService.Watch(ctx, app.conf.WatchPeriod)
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Fatal("failed to start estimation service", zap.Error(err))
		}
	}()

	metricsStorage := storage.NewMetricsStorage(redisClient)
	metricsService := service.NewMetricsService(metricsStorage)

	estimateHandler := handler.NewEstimateHandler(websiteService, estimateCache)
	adminHandler := handler.NewAdminHandler(metricsService)

	server := rest.New(
		app.conf.Server,
		redisClient,
		logger,
	).Handle(
		estimateHandler,
		adminHandler,
	)

	logger.Info("starting web service")
	go func() {
		err = server.Listen()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("failed to start web service", zap.Error(err))
		}
	}()

	<-ctx.Done()

	logger.Info("stopping app")

	logger.Info("shutting down web service")
	err = server.Shutdown()
	if err != nil {
		logger.Fatal("failed to shutdown web service", zap.Error(err))
	}
}
