package handler

import (
	"estimate/internal/dto"
	"estimate/internal/service"
	"estimate/internal/transport/middleware"
	"estimate/pkg/cache"
	"github.com/gofiber/fiber/v2"
	"time"
)

type EstimateHandler struct {
	websiteService service.WebsiteService
	cache          *cache.Cache
	cacheTag       string
}

func NewEstimateHandler(estimateService service.WebsiteService, cache *cache.Cache, cacheTag string) *EstimateHandler {
	return &EstimateHandler{
		websiteService: estimateService,
		cache:          cache,
		cacheTag:       cacheTag,
	}
}

func (handler *EstimateHandler) Register(router fiber.Router) {
	router.Get("", middleware.Cache(1*time.Minute, handler.cache, handler.cacheTag), handler.GetWebsiteAccessTime)
	router.Get("/max", middleware.Cache(1*time.Minute, handler.cache, handler.cacheTag), handler.GetWebsiteByMaxAccessTime)
	router.Get("/min", middleware.Cache(1*time.Minute, handler.cache, handler.cacheTag), handler.GetWebsiteByMinAccessTime)
}

func (handler *EstimateHandler) GetWebsiteAccessTime(c *fiber.Ctx) error {
	var request dto.GetWebsiteAccessTimeRequest
	if err := c.QueryParser(&request); err != nil {
		return err
	}

	if err := request.Validate(); err != nil {
		return err
	}

	website, err := handler.websiteService.Get(c.Context(), request.URL)
	if err != nil {
		return err
	}

	return c.JSON(dto.GetWebsiteAccessTimeResponse{
		LastCheckAt: website.LastCheckAt,
		AccessTime:  dto.Duration{Duration: website.AccessTime},
	})
}

func (handler *EstimateHandler) GetWebsiteByMaxAccessTime(c *fiber.Ctx) error {
	website, err := handler.websiteService.GetByMaxAccessTime(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(dto.GetWebsiteWithMaxAccessTimeResponse{
		URL:         website.URL,
		LastCheckAt: website.LastCheckAt,
		AccessTime:  dto.Duration{Duration: website.AccessTime},
	})
}

func (handler *EstimateHandler) GetWebsiteByMinAccessTime(c *fiber.Ctx) error {
	website, err := handler.websiteService.GetByMinAccessTime(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(dto.GetWebsiteWithMinAccessTimeResponse{
		URL:         website.URL,
		LastCheckAt: website.LastCheckAt,
		AccessTime:  dto.Duration{Duration: website.AccessTime},
	})
}
