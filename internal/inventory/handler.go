package inventory

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

	r.GET("/inventory/:sku_id", h.getInventory)
	r.POST("/inventory/reserve", h.reserve)
	r.POST("/inventory/release", h.release)
	r.POST("/inventory/release-by-order", h.releaseByOrder)
	r.POST("/inventory/confirm", h.confirm)
	r.GET("/inventory/reservations/:order_id", h.getReservation)
}

// @Summary 查询库存
// @Tags Inventory
// @Produce json
// @Param sku_id path string true "SKU ID"
// @Success 200 {object} httpx.Response
// @Router /inventory/{sku_id} [get]
func (h *Handler) getInventory(c *gin.Context) {
	skuID := c.Param("sku_id")
	inv, err := h.svc.GetInventory(skuID)
	if err != nil {
		code, body := httpx.Fail(errx.CodeNotFound, errx.MsgSkuNotFound)
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(inv)
	c.JSON(code, body)
}

// @Summary 预占库存
// @Tags Inventory
// @Accept json
// @Produce json
// @Param body body ReserveRequest true "预占请求"
// @Success 200 {object} httpx.Response
// @Router /inventory/reserve [post]
func (h *Handler) reserve(c *gin.Context) {
	var req ReserveRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Items) == 0 || req.OrderID == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}

	resp, err := h.svc.Reserve(req)
	if err != nil {
		if err == ErrInsufficient {
			code, body := httpx.Fail(errx.CodeInventoryInsufficient, errx.MsgInventoryInsufficient)
			c.JSON(code, body)
			return
		}
		code, body := httpx.Fail(errx.CodeInternalError, errx.MsgInventoryReserveFailed)
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

// @Summary 释放库存
// @Description 基于 reserved_id 释放，重复调用幂等
// @Tags Inventory
// @Accept json
// @Produce json
// @Param body body ReleaseRequest true "释放请求"
// @Success 200 {object} httpx.Response
// @Router /inventory/release [post]
func (h *Handler) release(c *gin.Context) {
	var req ReleaseRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ReservedID == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}

	if err := h.svc.Release(req.ReservedID); err != nil {
		if err == ErrNotFound {
			code, body := httpx.Fail(errx.CodeNotFound, errx.MsgReservationNotFound)
			c.JSON(code, body)
			return
		}
		code, body := httpx.Fail(errx.CodeInvalidReservationState, errx.MsgInvalidReservationState)
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"success": true})
	c.JSON(code, body)
}

// @Summary 确认扣减
// @Tags Inventory
// @Accept json
// @Produce json
// @Param body body ConfirmRequest true "确认请求"
// @Success 200 {object} httpx.Response
// @Router /inventory/confirm [post]
func (h *Handler) confirm(c *gin.Context) {
	var req ConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ReservedID == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}

	if err := h.svc.Confirm(req.ReservedID); err != nil {
		if err == ErrNotFound {
			code, body := httpx.Fail(errx.CodeNotFound, errx.MsgReservationNotFound)
			c.JSON(code, body)
			return
		}
		code, body := httpx.Fail(errx.CodeInvalidReservationState, errx.MsgInvalidReservationState)
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"success": true})
	c.JSON(code, body)
}

// @Summary 按订单号释放
// @Description 基于订单号幂等释放库存，重复调用不会重复回补
// @Tags Inventory
// @Accept json
// @Produce json
// @Param body body ReleaseByOrderRequest true "订单号释放请求"
// @Success 200 {object} httpx.Response
// @Router /inventory/release-by-order [post]
func (h *Handler) releaseByOrder(c *gin.Context) {
	var req ReleaseByOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.OrderID == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	if err := h.svc.ReleaseByOrder(req.OrderID); err != nil {
		if err == ErrNotFound {
			code, body := httpx.Fail(errx.CodeNotFound, errx.MsgReservationNotFound)
			c.JSON(code, body)
			return
		}
		code, body := httpx.Fail(errx.CodeInvalidReservationState, errx.MsgInvalidReservationState)
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"success": true})
	c.JSON(code, body)
}

// @Summary 查询预占记录
// @Tags Inventory
// @Produce json
// @Param order_id path string true "订单ID"
// @Success 200 {object} httpx.Response
// @Router /inventory/reservations/{order_id} [get]
func (h *Handler) getReservation(c *gin.Context) {
	orderID := c.Param("order_id")
	resv, err := h.svc.GetReservation(orderID)
	if err != nil {
		code, body := httpx.Fail(errx.CodeNotFound, errx.MsgReservationNotFound)
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(resv)
	c.JSON(code, body)
}
