package order

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

	r.POST("/orders", h.create)
	r.GET("/orders/:id", h.get)
}

func (h *Handler) create(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.RequestID == "" || req.UserID == "" || len(req.Items) == 0 {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}

	resp, err := h.svc.Create(req)
	if err != nil {
		if err == ErrIdempotentHit {
			code, body := httpx.OK(resp)
			body.Message = "IDEMPOTENT_HIT"
			c.JSON(code, body)
			return
		}
		if err == ErrInventoryFail {
			code, body := httpx.Fail(errx.CodeConflict, "insufficient inventory")
			c.JSON(code, body)
			return
		}
		if err == ErrInvalidState {
			code, body := httpx.Fail(errx.CodeInvalidState, "invalid order state")
			c.JSON(code, body)
			return
		}
		code, body := httpx.Fail(errx.CodeInternalError, "create order failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

func (h *Handler) get(c *gin.Context) {
	id := c.Param("id")
	ord, err := h.svc.Get(id)
	if err != nil {
		code, body := httpx.Fail(errx.CodeNotFound, "order not found")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(ord)
	c.JSON(code, body)
}
