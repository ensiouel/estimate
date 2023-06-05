package handler

import (
	"estimate/internal/service"
	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct {
	metricsService service.MetricsService
}

func NewAdminHandler(metricsService service.MetricsService) *AdminHandler {
	return &AdminHandler{metricsService: metricsService}
}

func (handler *AdminHandler) Register(router fiber.Router) {
	router.Get("/metrics", handler.Metrics)
}

func (handler *AdminHandler) Metrics(c *fiber.Ctx) error {
	metrics, err := handler.metricsService.Metrics(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(metrics)
}
