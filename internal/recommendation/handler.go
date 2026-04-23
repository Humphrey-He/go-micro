package recommendation

import (
	"strconv"

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

func (h *Handler) Register(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		code, body := httpx.OK(gin.H{"status": "ok"})
		c.JSON(code, body)
	})

	api := r.Group("/api/v1/rec")
	api.POST("/report", h.reportBehavior)
	api.GET("/similar/:sku_id", h.getSimilarProducts)
	api.GET("/home", h.getHomeRecommendations)
	api.GET("/cold-start", h.getColdStart)
	api.POST("/preference", h.setPreference)
	api.POST("/cart-addon", h.getCartAddons)
	api.GET("/pay-complete", h.getPayCompleteRecommendations)
}

func (h *Handler) reportBehavior(c *gin.Context) {
	var req BehaviorReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "invalid request")
		c.JSON(code, body)
		return
	}

	userID := int64(0)
	if uid, exists := c.Get(middleware.CtxUserID); exists {
		if id, ok := uid.(int64); ok {
			userID = id
		}
	}

	if err := h.svc.ReportBehavior(c.Request.Context(), &req, userID); err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "report failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(nil)
	c.JSON(code, body)
}

func (h *Handler) getSimilarProducts(c *gin.Context) {
	skuID, err := strconv.ParseInt(c.Param("sku_id"), 10, 64)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "invalid sku_id")
		c.JSON(code, body)
		return
	}

	scene := c.DefaultQuery("scene", "purchase")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	resp, err := h.svc.GetSimilarProducts(c.Request.Context(), skuID, scene, limit)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "get similar products failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

func (h *Handler) getHomeRecommendations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	userID := int64(0)
	if uid, exists := c.Get(middleware.CtxUserID); exists {
		if id, ok := uid.(int64); ok {
			userID = id
		}
	}

	resp, err := h.svc.GetHomeRecommendations(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "get home recommendations failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

func (h *Handler) getColdStart(c *gin.Context) {
	resp, err := h.svc.GetColdStartData(c.Request.Context())
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "get cold start data failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

func (h *Handler) setPreference(c *gin.Context) {
	userID := int64(0)
	if uid, exists := c.Get(middleware.CtxUserID); exists {
		if id, ok := uid.(int64); ok {
			userID = id
		}
	}

	var req SetPreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "invalid request")
		c.JSON(code, body)
		return
	}

	if err := h.svc.SetUserPreference(c.Request.Context(), userID, req.CategoryIDs); err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "set preference failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(nil)
	c.JSON(code, body)
}

type SetPreferenceRequest struct {
	CategoryIDs []int64 `json:"category_ids" binding:"required,min=1"`
}

func (h *Handler) getCartAddons(c *gin.Context) {
	var req struct {
		CartSKUIDs []int64 `json:"cart_sku_ids" binding:"required"`
		Limit      int      `json:"limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "invalid request")
		c.JSON(code, body)
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	resp, err := h.svc.GetCartAddons(c.Request.Context(), req.CartSKUIDs, req.Limit)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "get cart addons failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

func (h *Handler) getPayCompleteRecommendations(c *gin.Context) {
	purchasedStr := c.Query("purchased_sku_ids")
	if purchasedStr == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "purchased_sku_ids required")
		c.JSON(code, body)
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse comma-separated SKU IDs
	var purchasedSKUIDs []int64
	for _, s := range splitString(purchasedStr) {
		if id, err := strconv.ParseInt(s, 10, 64); err == nil {
			purchasedSKUIDs = append(purchasedSKUIDs, id)
		}
	}

	if len(purchasedSKUIDs) == 0 {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "no valid sku_ids")
		c.JSON(code, body)
		return
	}

	resp, err := h.svc.GetPayCompleteRecommendations(c.Request.Context(), purchasedSKUIDs, limit)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "get pay complete recommendations failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

func splitString(s string) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == ',' {
			if current != "" {
				result = append(result, current)
			}
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}