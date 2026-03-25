package payment

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
	r.GET("/health", func(c *gin.Context) {
		code, body := httpx.OK(gin.H{"status": "ok"})
		c.JSON(code, body)
	})
	r.GET("/healthz", func(c *gin.Context) {
		code, body := httpx.OK(gin.H{"status": "ok"})
		c.JSON(code, body)
	})
	r.GET("/readyz", func(c *gin.Context) {
		code, body := httpx.OK(gin.H{"status": "ready"})
		c.JSON(code, body)
	})

	r.POST("/payments", h.create)
	r.GET("/payments/:id", h.get)
	r.POST("/payments/:id/success", h.markSuccess)
	r.POST("/payments/:id/failed", h.markFailed)
	r.POST("/payments/:id/timeout", h.markTimeout)
	r.POST("/payments/:id/refund", h.refund)
}

func (h *Handler) create(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.OrderID == "" || req.RequestID == "" || req.Amount <= 0 {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	p, err := h.svc.Create(req)
	if err != nil {
		if err == ErrIdempotentHit {
			code, body := httpx.OK(p)
			body.Message = "IDEMPOTENT_HIT"
			c.JSON(code, body)
			return
		}
		code, body := httpx.Fail(errx.CodeInternalError, "create payment failed")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(p)
	c.JSON(code, body)
}

func (h *Handler) get(c *gin.Context) {
	id := c.Param("id")
	p, err := h.svc.Get(id)
	if err != nil {
		code, body := httpx.Fail(errx.CodeNotFound, "payment not found")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(p)
	c.JSON(code, body)
}

func (h *Handler) markSuccess(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.MarkSuccess(id); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidState, "invalid payment state")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"success": true})
	c.JSON(code, body)
}

func (h *Handler) markFailed(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.MarkFailed(id); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidState, "invalid payment state")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"success": true})
	c.JSON(code, body)
}

func (h *Handler) markTimeout(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.MarkTimeout(id); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidState, "invalid payment state")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"success": true})
	c.JSON(code, body)
}

func (h *Handler) refund(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Refund(id); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidState, "invalid payment state")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"success": true})
	c.JSON(code, body)
}
