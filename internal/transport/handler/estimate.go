package handler

import (
	"estimate/internal/dto"
	"estimate/internal/entity"
	"estimate/internal/service"
	"estimate/internal/transport/middleware"
	"estimate/pkg/cache"
	"github.com/gofiber/fiber/v2"
	"time"
)

type EstimateHandler struct {
	websiteService service.WebsiteService
	cache          cache.Cache
	cacheTag       string
}

func NewEstimateHandler(estimateService service.WebsiteService, cache cache.Cache, cacheTag string) *EstimateHandler {
	return &EstimateHandler{
		websiteService: estimateService,
		cache:          cache,
		cacheTag:       cacheTag,
	}
}

func (handler *EstimateHandler) Register(router fiber.Router) {
	cacheMiddleware := middleware.Cache(1*time.Minute, handler.cache, handler.cacheTag)

	router.Get("", cacheMiddleware, handler.CheckWebsite)
	router.Get("/max", cacheMiddleware, handler.GetWebsiteByMaxAccessTime)
	router.Get("/min", cacheMiddleware, handler.GetWebsiteByMinAccessTime)
}

func (handler *EstimateHandler) CheckWebsite(c *fiber.Ctx) error {
	var request dto.GetWebsiteAccessTimeRequest
	err := c.QueryParser(&request)
	if err != nil {
		return err
	}

	err = request.Validate()
	if err != nil {
		return err
	}

	var website entity.Website
	website, err = handler.websiteService.GetByURL(c.Context(), request.URL)
	if err != nil {
		return err
	}

	return c.JSON(dto.GetWebsiteAccessTimeResponse{
		AccessTime:  dto.Duration{Duration: website.AccessTime},
		LastCheckAt: website.LastCheckAt,
	})
}

func (handler *EstimateHandler) GetWebsiteByMaxAccessTime(c *fiber.Ctx) error {
	website, err := handler.websiteService.GetByMaxAccessTime(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(dto.GetWebsiteWithMaxAccessTimeResponse{
		URL:         website.URL,
		AccessTime:  dto.Duration{Duration: website.AccessTime},
		LastCheckAt: website.LastCheckAt,
	})
}

func (handler *EstimateHandler) GetWebsiteByMinAccessTime(c *fiber.Ctx) error {
	website, err := handler.websiteService.GetByMinAccessTime(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(dto.GetWebsiteWithMinAccessTimeResponse{
		URL:         website.URL,
		AccessTime:  dto.Duration{Duration: website.AccessTime},
		LastCheckAt: website.LastCheckAt,
	})
}
