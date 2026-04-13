package middleware

import (
	"net/http"

	"github.com/go_video_subs/internal/domain/subscription"
	"github.com/go_video_subs/pkg/response"
	"github.com/gofiber/fiber/v3"
)

func RequireSubscription(subRepo subscription.Repository) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(uint64)
		if !ok || userID == 0 {
			return c.Status(http.StatusUnauthorized).JSON(response.Fail("unauthorized"))
		}

		sub, err := subRepo.FindActiveByUserID(c.Context(), userID)
		if err != nil || sub == nil {
			return c.Status(http.StatusForbidden).JSON(response.Fail("no active subscription found"))
		}

		c.Locals("subscription_tier", sub.Tier)
		return c.Next()
	}
}
