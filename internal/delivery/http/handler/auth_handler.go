package handler

import (
	"errors"
	"net/http"

	"github.com/go_video_subs/internal/usecase/user"
	"github.com/go_video_subs/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	userUC *user.UseCase
}

func NewAuthHandler(userUC *user.UseCase) *AuthHandler {
	return &AuthHandler{userUC: userUC}
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var input user.LoginInput
	if err := c.Bind().JSON(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.Fail(err.Error()))
	}

	out, err := h.userUC.Login(c.Context(), input)
	if err != nil {
		if errors.Is(err, user.ErrInvalidCredentials) {
			return c.Status(http.StatusUnauthorized).JSON(response.Fail(err.Error()))
		}
		return c.Status(http.StatusInternalServerError).JSON(response.Fail(err.Error()))
	}

	return c.JSON(response.OK(out))
}
