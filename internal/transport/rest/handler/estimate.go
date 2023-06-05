package handler

import (
	"estimate/internal/dto"
	"estimate/internal/entity"
	"estimate/internal/service"
	"estimate/internal/transport/rest/middleware"
	"github.com/alejandro-carstens/gocache"
	"github.com/gofiber/fiber/v2"
	"time"
)

type EstimateHandler struct {
	websiteService service.WebsiteService
	cache          gocache.TaggedCache
}

func NewEstimateHandler(websiteService service.WebsiteService, cache gocache.TaggedCache) *EstimateHandler {
	return &EstimateHandler{
		websiteService: websiteService,
		cache:          cache,
	}
}

func (handler *EstimateHandler) Register(router fiber.Router) {
	cacheMiddleware := middleware.Cache(1*time.Minute, handler.cache)

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
