package handler

import (
	"errors"
	"net/http"

	ucvideo "github.com/go_video_subs/internal/usecase/video"
	"github.com/go_video_subs/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type VideoHandler struct {
	videoUC *ucvideo.UseCase
}

func NewVideoHandler(videoUC *ucvideo.UseCase) *VideoHandler {
	return &VideoHandler{videoUC: videoUC}
}

func (h *VideoHandler) GetVideos(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok || userID == 0 {
		return c.Status(http.StatusUnauthorized).JSON(response.Fail("unauthorized"))
	}

	videos, err := h.videoUC.GetVideosByUserTier(c.Context(), userID)
	if err != nil {
		if errors.Is(err, ucvideo.ErrNoActiveSubscription) {
			return c.Status(http.StatusForbidden).JSON(response.Fail("no active subscription found"))
		}
		return c.Status(http.StatusInternalServerError).JSON(response.Fail(err.Error()))
	}

	return c.JSON(response.OK(videos))
}
