package app

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"inspector/internal/config"
	"inspector/internal/service"
	"inspector/internal/storage"
	"inspector/internal/transport"
	"inspector/internal/transport/handler"
	cachepkg "inspector/pkg/cache"
	loggerpkg "inspector/pkg/logger"
	"inspector/pkg/postgres"
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

	logger.Info("Starting app...")

	logger.Info("Connecting to postgres...")
	pgClient, err := postgres.NewClient(ctx, postgres.Config{
		Host:     app.conf.Postgres.Host,
		Port:     app.conf.Postgres.Port,
		User:     app.conf.Postgres.User,
		Password: app.conf.Postgres.Password,
		DB:       app.conf.Postgres.DB,
	})
	if err != nil {
		logger.Fatal(err.Error())
	}

	logger.Info("Connecting to redis...")
	redisClient := redis.NewClient(&redis.Options{
		Addr: app.conf.Redis.Addr,
	})
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		logger.Fatal(err.Error())
	}

	cache := cachepkg.New(redisClient)

	websiteStorage := storage.NewWebsiteStorage(pgClient)
	websiteService := service.NewWebsiteService(websiteStorage, app.conf.EstimationPeriod, cache, "website")

	logger.Info("Running estimation...")
	go func() {
		err = websiteService.RunEstimation(ctx)
		if err != nil {
			logger.Fatal(err.Error())
		}
	}()

	estimateHandler := handler.NewEstimateHandler(websiteService, cache, "website")

	metricsStorage := storage.NewMetricsStorage(redisClient)
	metricsService := service.NewMetricsService(metricsStorage)

	adminHandler := handler.NewAdminHandler(metricsService)

	logger.Info("Starting web service...")
	go func() {
		err = transport.New(
			app.conf.Server,
			redisClient,
			logger,
		).Handle(
			estimateHandler,
			adminHandler,
		).Listen()
		if err != nil && errors.Is(err, http.ErrServerClosed) == false {
			logger.Fatal(err.Error())
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("Stopping...")
	}
}
