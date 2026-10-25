package inventory

import (
	"errors"
	"time"

	"ecomm-micro/pkg/errx"
	"ecomm-micro/pkg/httpx"
	"github.com/gin-gonic/gin"
)

// PreStockReserveRequest Redis 预扣请求
type PreStockReserveRequest struct {
	SkuID     string `json:"sku_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required"`
	RequestID string `json:"request_id" binding:"required"`
}

// PreStockRestoreRequest Redis 预扣归还请求
type PreStockRestoreRequest struct {
	SkuID     string `json:"sku_id" binding:"required"`
	RequestID string `json:"request_id" binding:"required"`
}

type Handler struct {
	svc    *Service
	pre    *PreStock
	syncer *StockSyncer
	rec    *Reconciler
}

func NewHandler(svc *Service, pre *PreStock, syncer *StockSyncer, rec *Reconciler) *Handler {
	return &Handler{svc: svc, pre: pre, syncer: syncer, rec: rec}
}

func (h *Handler) Register(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		code, body := httpx.OK(gin.H{"status": "ok"})
		c.JSON(code, body)
	})

	r.GET("/inventory/:sku_id", h.getInventory)
	r.GET("/inventory", h.listInventory)
	r.POST("/inventory/reserve", h.reserve)
	r.POST("/inventory/release", h.release)
	r.POST("/inventory/release-by-order", h.releaseByOrder)
	r.POST("/inventory/confirm", h.confirm)
	r.GET("/inventory/reservations/:order_id", h.getReservation)

	// R1: Redis Lua 预扣加速层（与 TCC 预占并存，高并发场景使用）
	pre := r.Group("/inventory/prestock")
	pre.POST("/init", h.preStockInit)
	pre.POST("/reserve", h.preStockReserve)
	pre.POST("/restore", h.preStockRestore)
	pre.GET("/avail/:sku_id", h.preStockAvail)
	pre.POST("/sync", h.preStockSync)
	pre.POST("/reconcile", h.preStockReconcile)
}

// @Summary 初始化 Redis 预扣额度（演示/压测用，SETNX 幂等）
// @Tags Inventory
// @Accept json
// @Produce json
// @Param body body PreStockReserveRequest true "初始化请求 {sku_id, quantity=额度}"
// @Success 200 {object} httpx.Response
// @Router /inventory/prestock/init [post]
func (h *Handler) preStockInit(c *gin.Context) {
	var req PreStockReserveRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.SkuID == "" || req.Quantity <= 0 {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	if err := h.pre.InitAvail(c.Request.Context(), req.SkuID, req.Quantity); err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "init failed")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"sku_id": req.SkuID, "avail": req.Quantity})
	c.JSON(code, body)
}

// @Summary Redis 预扣库存（Lua 原子：读判扣记）
// @Tags Inventory
// @Accept json
// @Produce json
// @Param body body PreStockReserveRequest true "预扣请求 {sku_id, quantity, request_id}"
// @Success 200 {object} httpx.Response
// @Router /inventory/prestock/reserve [post]
func (h *Handler) preStockReserve(c *gin.Context) {
	var req PreStockReserveRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.SkuID == "" || req.Quantity <= 0 || req.RequestID == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	res, err := h.pre.Reserve(c.Request.Context(), req.SkuID, req.Quantity, req.RequestID)
	if err != nil {
		if errors.Is(err, ErrSkuNotFound) {
			code, body := httpx.Fail(errx.CodeSkuNotFound, errx.MsgSkuNotFound)
			c.JSON(code, body)
			return
		}
		code, body := httpx.Fail(errx.CodeInternalError, "prestock reserve failed")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{
		"result":     preStockResultName(res),
		"request_id": req.RequestID,
		"quantity":   req.Quantity,
	})
	c.JSON(code, body)
}

// @Summary 归还 Redis 预扣（订单取消）
// @Tags Inventory
// @Accept json
// @Produce json
// @Param body body PreStockRestoreRequest true "归还请求 {sku_id, request_id}"
// @Success 200 {object} httpx.Response
// @Router /inventory/prestock/restore [post]
func (h *Handler) preStockRestore(c *gin.Context) {
	var req PreStockRestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.SkuID == "" || req.RequestID == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	restored, err := h.pre.Restore(c.Request.Context(), req.SkuID, req.RequestID)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "prestock restore failed")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"restored": restored})
	c.JSON(code, body)
}

// @Summary 查询 Redis 预扣剩余额度
// @Tags Inventory
// @Produce json
// @Param sku_id path string true "SKU ID"
// @Success 200 {object} httpx.Response
// @Router /inventory/prestock/avail/{sku_id} [get]
func (h *Handler) preStockAvail(c *gin.Context) {
	avail, err := h.pre.GetAvail(c.Request.Context(), c.Param("sku_id"))
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "query avail failed")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"avail": avail})
	c.JSON(code, body)
}

// @Summary 手动触发一轮流水落库（演示用）
// @Tags Inventory
// @Produce json
// @Success 200 {object} httpx.Response
// @Router /inventory/prestock/sync [post]
func (h *Handler) preStockSync(c *gin.Context) {
	n, err := h.syncer.SyncOnce(c.Request.Context())
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "sync failed")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"synced": n})
	c.JSON(code, body)
}

// @Summary 手动触发一轮对账（演示用：输出差异报告）
// @Tags Inventory
// @Produce json
// @Success 200 {object} httpx.Response
// @Router /inventory/prestock/reconcile [post]
func (h *Handler) preStockReconcile(c *gin.Context) {
	report, err := h.rec.RunOnce(c.Request.Context(), time.Now())
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "reconcile failed")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(report)
	c.JSON(code, body)
}

func preStockResultName(res PreStockResult) string {
	switch res {
	case PreStockGranted:
		return "GRANTED"
	case PreStockDuplicate:
		return "DUPLICATE"
	case PreStockInsufficient:
		return "INSUFFICIENT"
	default:
		return "UNKNOWN"
	}
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

// @Summary 查询所有库存
// @Tags Inventory
// @Produce json
// @Success 200 {object} httpx.Response
// @Router /inventory [get]
func (h *Handler) listInventory(c *gin.Context) {
	items, err := h.svc.ListInventory()
	if err != nil {
		code, body := httpx.Fail(errx.CodeUpstreamUnavail, "failed to list inventory")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(items)
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
