package price

import (
	"net/http"
	"strconv"

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
	api.POST("/price/calculate", h.calculate)
	api.GET("/price/history", h.history)
}

// @Summary Calculate price
// @Tags Price
// @Accept json
// @Produce json
// @Param body body CalculateRequest true "calculate price"
// @Success 200 {object} httpx.Response
// @Router /api/v1/price/calculate [post]
func (h *Handler) calculate(c *gin.Context) {
	var req CalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.SkuID == "" || req.BasePrice <= 0 {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	resp, err := h.svc.Calculate(req)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, err.Error())
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

// @Summary Price history
// @Tags Price
// @Produce json
// @Param sku_id query string true "sku id"
// @Param limit query int false "limit"
// @Success 200 {object} httpx.Response
// @Router /api/v1/price/history [get]
func (h *Handler) history(c *gin.Context) {
	skuID := c.Query("sku_id")
	limitStr := c.Query("limit")
	limit := 20
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil {
			limit = n
		}
	}
	rows, err := h.svc.History(skuID, limit)
	if err != nil {
		code, body := httpx.Fail(errx.CodeNotFound, errx.MsgNotFound)
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(rows)
	c.JSON(code, body)
}
