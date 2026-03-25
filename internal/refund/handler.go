package refund

import (
	"github.com/gin-gonic/gin"
	"go-micro/pkg/errx"
	"go-micro/pkg/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r *gin.Engine) {
	api := r.Group("/api/v1")
	api.POST("/refund/initiate", h.initiate)
	api.POST("/refund/status", h.status)
	api.POST("/refund/rollback", h.rollback)
}

// @Summary Initiate refund
// @Tags Refund
// @Accept json
// @Produce json
// @Param body body InitiateRequest true "initiate refund"
// @Success 200 {object} httpx.Response
// @Router /api/v1/refund/initiate [post]
func (h *Handler) initiate(c *gin.Context) {
	var req InitiateRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.RefundID == "" || req.OrderID == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	ref, err := h.svc.Initiate(req)
	if err != nil && err != ErrIdempotentHit {
		code, body := httpx.Fail(errx.CodeInternalError, err.Error())
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(ref)
	c.JSON(code, body)
}

// @Summary Refund status
// @Tags Refund
// @Accept json
// @Produce json
// @Param body body StatusRequest true "refund status"
// @Success 200 {object} httpx.Response
// @Router /api/v1/refund/status [post]
func (h *Handler) status(c *gin.Context) {
	var req StatusRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.RefundID == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	ref, err := h.svc.Get(req.RefundID)
	if err != nil {
		code, body := httpx.Fail(errx.CodeNotFound, errx.MsgNotFound)
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(ref)
	c.JSON(code, body)
}

// @Summary Rollback on cancel
// @Tags Refund
// @Accept json
// @Produce json
// @Param body body RollbackRequest true "rollback refund"
// @Success 200 {object} httpx.Response
// @Router /api/v1/refund/rollback [post]
func (h *Handler) rollback(c *gin.Context) {
	var req RollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.RefundID == "" || req.OrderID == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	ref, err := h.svc.Rollback(req)
	if err != nil && err != ErrIdempotentHit {
		code, body := httpx.Fail(errx.CodeInternalError, err.Error())
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(ref)
	c.JSON(code, body)
}
