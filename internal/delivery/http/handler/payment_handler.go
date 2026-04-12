package handler

import (
	"net/http"

	ucPayment "github.com/go_video_subs/internal/usecase/payment"
	"github.com/go_video_subs/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type PaymentHandler struct {
	uc *ucPayment.UseCase
}

func NewPaymentHandler(uc *ucPayment.UseCase) *PaymentHandler {
	return &PaymentHandler{uc: uc}
}

type initiateRequest struct {
	Tier string `json:"tier"`
}

type callbackRequest struct {
	TransactionID uint64 `json:"transaction_id"`
	Status        string `json:"status"`
}

func (h *PaymentHandler) Initiate(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok || userID == 0 {
		return c.Status(http.StatusUnauthorized).JSON(response.Fail("unauthorized"))
	}

	var req initiateRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.Fail("invalid request body"))
	}
	if req.Tier == "" {
		return c.Status(http.StatusBadRequest).JSON(response.Fail("tier is required"))
	}

	out, err := h.uc.InitiatePayment(c.Context(), ucPayment.InitiatePaymentInput{
		UserID: userID,
		Tier:   req.Tier,
	})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.Fail(err.Error()))
	}

	return c.Status(http.StatusCreated).JSON(response.OK(out))
}

func (h *PaymentHandler) Callback(c fiber.Ctx) error {
	var req callbackRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.Fail("invalid request body"))
	}
	if req.TransactionID == 0 {
		return c.Status(http.StatusBadRequest).JSON(response.Fail("transaction_id is required"))
	}
	if req.Status == "" {
		return c.Status(http.StatusBadRequest).JSON(response.Fail("status is required"))
	}

	if err := h.uc.HandleCallback(c.Context(), ucPayment.HandleCallbackInput{
		TransactionID: req.TransactionID,
		Status:        req.Status,
	}); err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.Fail(err.Error()))
	}

	return c.JSON(response.OK(fiber.Map{"message": "callback processed"}))
}
