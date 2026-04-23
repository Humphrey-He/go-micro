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
	userID, _ := c.Get(middleware.CtxUserID)
	uid, _ := userID.(int64)

	var req SetPreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "invalid request")
		c.JSON(code, body)
		return
	}

	if err := h.svc.SetUserPreference(c.Request.Context(), uid, req.CategoryIDs); err != nil {
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