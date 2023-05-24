package middleware

import (
	"errors"
	"estimate/pkg/cache"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"time"
)

func Cache(expiration time.Duration, cache cache.Cache, tag string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.OriginalURL()

		result, err := cache.Get(c.Context(), key, tag)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				err = c.Next()
				if err != nil {
					return err
				}

				body := c.Response().Body()
				err = cache.Set(c.Context(), key, tag, body, expiration)
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
