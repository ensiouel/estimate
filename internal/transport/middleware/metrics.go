package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func Metrics(client *redis.Client) fiber.Handler {
	prefix := "metrics:"

	return func(c *fiber.Ctx) error {
		key := c.Path()

		err := client.Incr(c.Context(), prefix+key).Err()
		if err != nil {
			return err
		}

		return c.Next()
	}
}
