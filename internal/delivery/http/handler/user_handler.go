package handler

import (
	"net/http"

	"github.com/go_video_subs/internal/usecase/user"
	"github.com/go_video_subs/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	userUC *user.UseCase
}

func NewUserHandler(userUC *user.UseCase) *UserHandler {
	return &UserHandler{userUC: userUC}
}

func (h *UserHandler) Create(c fiber.Ctx) error {
	var input user.CreateUserInput
	if err := c.Bind().JSON(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.Fail(err.Error()))
	}

	if err := h.userUC.CreateUser(c.Context(), input); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(response.Fail(err.Error()))
	}

	return c.Status(http.StatusCreated).JSON(response.OK(nil))
}
