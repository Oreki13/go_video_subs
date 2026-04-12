package middleware

import (
	"net/http"
	"strings"

	appjwt "github.com/go_video_subs/pkg/jwt"
	"github.com/go_video_subs/pkg/response"
	"github.com/gofiber/fiber/v3"
)

func Auth(jwtSecret string) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(http.StatusUnauthorized).JSON(response.Fail("missing authorization header"))
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.Status(http.StatusUnauthorized).JSON(response.Fail("invalid authorization header format"))
		}

		claims, err := appjwt.ParseToken(parts[1], jwtSecret)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(response.Fail("invalid or expired token"))
		}

		c.Locals("user_id", claims.UserID)
		return c.Next()
	}
}
