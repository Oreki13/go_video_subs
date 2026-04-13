package handler

import (
	"net/http"

	"github.com/go_video_subs/internal/domain/subscription"
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
	tier, ok := c.Locals("subscription_tier").(subscription.Tier)
	if !ok {
		return c.Status(http.StatusForbidden).JSON(response.Fail("subscription tier not found"))
	}

	videos, err := h.videoUC.GetVideosByTier(c.Context(), tier)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(response.Fail(err.Error()))
	}

	return c.JSON(response.OK(videos))
}
