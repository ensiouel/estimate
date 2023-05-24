package transport

import (
	"estimate/internal/config"
	"estimate/internal/transport/handler"
	"estimate/internal/transport/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Server struct {
	router      *fiber.App
	conf        config.Server
	redisClient *redis.Client
}

func New(conf config.Server, redisClient *redis.Client, log *zap.Logger) *Server {
	router := fiber.New(fiber.Config{
		ErrorHandler: middleware.Error(log),
	})

	router.Use(
		recover.New(),
		logger.New(),
	)

	return &Server{
		router:      router,
		conf:        conf,
		redisClient: redisClient,
	}
}

func (server *Server) Handle(estimateHandler *handler.EstimateHandler, adminHandler *handler.AdminHandler) *Server {
	auth := basicauth.New(basicauth.Config{
		Users: map[string]string{
			server.conf.Admin.Username: server.conf.Admin.Password,
		},
	})

	api := server.router.Group("/api", middleware.Metrics(server.redisClient))
	{
		v1 := api.Group("/v1")
		{
			estimateHandler.Register(v1.Group("/estimate"))
		}
	}

	adminHandler.Register(server.router.Group("/admin", auth))

	return server
}

func (server *Server) Listen() error {
	return server.router.Listen(server.conf.Addr)
}

func (server *Server) Shutdown() error {
	return server.router.Shutdown()
}
