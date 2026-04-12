package router

import (
	"github.com/go_video_subs/internal/delivery/http/handler"
	"github.com/go_video_subs/internal/delivery/http/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

type Handlers struct {
	User    *handler.UserHandler
	Auth    *handler.AuthHandler
	Video   *handler.VideoHandler
	Payment *handler.PaymentHandler

	JWTSecret string
}

func New(h *Handlers) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	// --- Global Middleware ---
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} ${latency} ${method} ${path}\n",
	}))

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "server is running",
		})
	})

	v1 := app.Group("/api/v1")

	auth := v1.Group("/auth")
	auth.Post("/login", h.Auth.Login)

	users := v1.Group("/users")
	users.Post("/", h.User.Create)

	videos := v1.Group("/videos", middleware.Auth(h.JWTSecret))
	videos.Get("/", h.Video.GetVideos)

	payments := v1.Group("/payments")
	payments.Post("/", middleware.Auth(h.JWTSecret), h.Payment.Initiate)
	payments.Post("/callback", h.Payment.Callback)

	return app
}
