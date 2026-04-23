package pricewatch

import (
	"github.com/gin-gonic/gin"
	"go-micro/pkg/errx"
	"go-micro/pkg/httpx"
	"go-micro/pkg/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r *gin.RouterGroup) {
	priceWatch := r.Group("/price-watch")
	priceWatch.POST("", h.setPriceWatch)
	priceWatch.DELETE("/:sku_id", h.cancelPriceWatch)
	priceWatch.GET("/list", h.getPriceWatchList)
	priceWatch.PUT("/:sku_id", h.updatePriceWatch)

	// Notifications (under /api/v1/price-watch/notifications)
	r.GET("/notifications/price-watch", h.getNotifications)
}

// SetPriceWatch creates or updates a price watch
func (h *Handler) setPriceWatch(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		code, body := httpx.Fail(errx.CodeUnauthorized, "not authenticated")
		c.JSON(code, body)
		return
	}

	var req SetPriceWatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "invalid request")
		c.JSON(code, body)
		return
	}

	resp, err := h.svc.SetPriceWatch(c.Request.Context(), userID, &req)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "set price watch failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

// CancelPriceWatch removes a price watch
func (h *Handler) cancelPriceWatch(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		code, body := httpx.Fail(errx.CodeUnauthorized, "not authenticated")
		c.JSON(code, body)
		return
	}

	skuID := c.Param("sku_id")
	if skuID == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "sku_id required")
		c.JSON(code, body)
		return
	}

	if err := h.svc.CancelPriceWatch(c.Request.Context(), userID, skuID); err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "cancel price watch failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(gin.H{"message": "监控已取消"})
	c.JSON(code, body)
}

// GetPriceWatchList returns user's price watches
func (h *Handler) getPriceWatchList(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		code, body := httpx.Fail(errx.CodeUnauthorized, "not authenticated")
		c.JSON(code, body)
		return
	}

	page := 1
	pageSize := 20
	status := "all"

	if p := c.Query("page"); p != "" {
		if v := parseInt(p); v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v := parseInt(ps); v > 0 {
			pageSize = v
		}
	}
	if s := c.Query("status"); s != "" {
		status = s
	}

	resp, err := h.svc.GetPriceWatchList(c.Request.Context(), userID, page, pageSize, status)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "get price watch list failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

// UpdatePriceWatch updates watch settings
func (h *Handler) updatePriceWatch(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		code, body := httpx.Fail(errx.CodeUnauthorized, "not authenticated")
		c.JSON(code, body)
		return
	}

	skuID := c.Param("sku_id")
	if skuID == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "sku_id required")
		c.JSON(code, body)
		return
	}

	var req UpdatePriceWatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "invalid request")
		c.JSON(code, body)
		return
	}

	if err := h.svc.UpdatePriceWatch(c.Request.Context(), userID, skuID, &req); err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "update price watch failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(gin.H{"message": "更新成功"})
	c.JSON(code, body)
}

// GetPriceHistory returns price history for a product
func (h *Handler) getPriceHistory(c *gin.Context) {
	skuID := c.Param("sku_id")
	if skuID == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "sku_id required")
		c.JSON(code, body)
		return
	}

	period := c.DefaultQuery("period", "30d")

	resp, err := h.svc.GetPriceHistory(c.Request.Context(), skuID, period)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "get price history failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

// GetNotifications returns user's price watch notifications
func (h *Handler) getNotifications(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		code, body := httpx.Fail(errx.CodeUnauthorized, "not authenticated")
		c.JSON(code, body)
		return
	}

	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if v := parseInt(p); v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v := parseInt(ps); v > 0 {
			pageSize = v
		}
	}

	resp, err := h.svc.GetNotifications(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "get notifications failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

func getUserID(c *gin.Context) int64 {
	if uid, exists := c.Get(middleware.CtxUserID); exists {
		if id, ok := uid.(int64); ok {
			return id
		}
	}
	return 0
}

func parseInt(s string) int {
	v := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			v = v*10 + int(c-'0')
		}
	}
	return v
}
