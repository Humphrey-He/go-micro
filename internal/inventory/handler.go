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
	r.GET("/healthz", func(c *gin.Context) {
		code, body := httpx.OK(gin.H{"status": "ok"})
		c.JSON(code, body)
	})
	r.GET("/readyz", func(c *gin.Context) {
		code, body := httpx.OK(gin.H{"status": "ready"})
		c.JSON(code, body)
	})

	r.GET("/inventory/:sku_id", h.getInventory)
	r.POST("/inventory/reserve", h.reserve)
	r.POST("/inventory/release", h.release)
	r.POST("/inventory/confirm", h.confirm)
}

// @Summary ????
// @Tags Inventory
// @Produce json
// @Param sku_id path string true "SKU ID"
// @Success 200 {object} httpx.Response
// @Router /inventory/{sku_id} [get]
func (h *Handler) getInventory(c *gin.Context) {
	skuID := c.Param("sku_id")
	inv, err := h.svc.GetInventory(skuID)
	if err != nil {
		code, body := httpx.Fail(errx.CodeNotFound, "sku not found")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(inv)
	c.JSON(code, body)
}

// @Summary ????
// @Tags Inventory
// @Accept json
// @Produce json
// @Param body body ReserveRequest true "????"
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
			code, body := httpx.Fail(errx.CodeConflict, "insufficient inventory")
			c.JSON(code, body)
			return
		}
		code, body := httpx.Fail(errx.CodeInternalError, "reserve failed")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

// @Summary ????
// @Tags Inventory
// @Accept json
// @Produce json
// @Param body body ReleaseRequest true "????"
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
			code, body := httpx.Fail(errx.CodeNotFound, "reservation not found")
			c.JSON(code, body)
			return
		}
		code, body := httpx.Fail(errx.CodeInvalidState, "invalid reservation state")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"success": true})
	c.JSON(code, body)
}

// @Summary ????
// @Tags Inventory
// @Accept json
// @Produce json
// @Param body body ConfirmRequest true "????"
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
			code, body := httpx.Fail(errx.CodeNotFound, "reservation not found")
			c.JSON(code, body)
			return
		}
		code, body := httpx.Fail(errx.CodeInvalidState, "invalid reservation state")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"success": true})
	c.JSON(code, body)
}
