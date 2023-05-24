package middleware

import (
	"errors"
	"estimate/pkg/apperror"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Error(logger *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		var e *fiber.Error
		if errors.As(err, &e) {
			return c.Status(e.Code).JSON(fiber.Map{"error": e.Error()})
		}

		var apperr apperror.Error
		if errors.As(err, &apperr) {
			switch apperr.Code {
			case apperror.Internal.Code:
				logger.Error("internal error", zap.Error(err))

				c.Status(fiber.StatusInternalServerError)
			case apperror.NotFound.Code:
				c.Status(fiber.StatusNotFound)
			case apperror.AlreadyExists.Code, apperror.BadRequest.Code:
				c.Status(fiber.StatusBadRequest)
			case apperror.Unauthorized.Code:
				c.Status(fiber.StatusUnauthorized)
			}

			return c.JSON(fiber.Map{"error": apperr})
		}

		logger.Warn("unexpected error", zap.Error(err))

		return c.Status(fiber.StatusTeapot).JSON(fiber.Map{"error": apperror.Unknown.WithError(err)})
	}
}
