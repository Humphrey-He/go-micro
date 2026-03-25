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
	r.GET("/healthz", func(c *gin.Context) {
		code, body := httpx.OK(gin.H{"status": "ok"})
		c.JSON(code, body)
	})
	r.GET("/readyz", func(c *gin.Context) {
		code, body := httpx.OK(gin.H{"status": "ready"})
		c.JSON(code, body)
	})

	r.POST("/orders", h.create)
	r.GET("/orders/:id", h.get)
	r.GET("/orders/biz/:biz_no", h.getByBizNo)
	r.POST("/orders/:id/cancel", h.cancel)
}

// @Summary 创建订单
// @Tags Order
// @Accept json
// @Produce json
// @Param body body CreateOrderRequest true "创建订单请求"
// @Success 200 {object} httpx.Response
// @Router /orders [post]
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

// @Summary 查询订单
// @Tags Order
// @Produce json
// @Param id path string true "订单ID"
// @Success 200 {object} httpx.Response
// @Router /orders/{id} [get]
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

// @Summary 按业务单号查询
// @Tags Order
// @Produce json
// @Param biz_no path string true "业务订单号"
// @Success 200 {object} httpx.Response
// @Router /orders/biz/{biz_no} [get]
func (h *Handler) getByBizNo(c *gin.Context) {
	bizNo := c.Param("biz_no")
	ord, err := h.svc.GetByBizNo(bizNo)
	if err != nil {
		code, body := httpx.Fail(errx.CodeNotFound, "order not found")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(ord)
	c.JSON(code, body)
}

// @Summary 取消订单
// @Description 幂等语义：已取消返回成功，已完成不可取消
// @Tags Order
// @Produce json
// @Param id path string true "订单ID"
// @Success 200 {object} httpx.Response
// @Router /orders/{id}/cancel [post]
func (h *Handler) cancel(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.Cancel(id)
	if err != nil {
		if err == ErrNotCancelable {
			code, body := httpx.Fail(errx.CodeInvalidState, "order not cancelable")
			c.JSON(code, body)
			return
		}
		if err == ErrNotFound {
			code, body := httpx.Fail(errx.CodeNotFound, "order not found")
			c.JSON(code, body)
			return
		}
		code, body := httpx.Fail(errx.CodeInternalError, "cancel order failed")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"success": true})
	c.JSON(code, body)
}
