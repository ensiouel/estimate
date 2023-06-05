package middleware

import (
	"errors"
	"github.com/alejandro-carstens/gocache"
	"github.com/gofiber/fiber/v2"
	"time"
)

func Cache(expiration time.Duration, cache gocache.TaggedCache) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.OriginalURL()

		result, err := cache.GetString(key)
		if err != nil {
			if errors.Is(err, gocache.ErrNotFound) {
				err = c.Next()
				if err != nil {
					return err
				}

				body := c.Response().Body()
				err = cache.Put(key, string(body), expiration)
				if err != nil {
					return err
				}

				return nil
			}

			return err
		}

		return c.Type("json").SendString(result)
	}
}
