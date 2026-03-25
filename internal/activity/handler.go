package activity

import (
	"net/http"

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
	api.POST("/activity/coupon", h.issueCoupon)
	api.POST("/activity/seckill", h.seckill)
	api.GET("/activity/status", h.status)
}

// @Summary Issue coupon
// @Tags Activity
// @Accept json
// @Produce json
// @Param body body CouponRequest true "issue coupon"
// @Success 200 {object} httpx.Response
// @Router /api/v1/activity/coupon [post]
func (h *Handler) issueCoupon(c *gin.Context) {
	var req CouponRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.CouponID == "" || req.UserID == "" || req.Amount <= 0 {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	cp, err := h.svc.IssueCoupon(req)
	if err != nil && err != ErrIdempotentHit {
		code, body := httpx.Fail(errx.CodeInternalError, err.Error())
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(cp)
	c.JSON(code, body)
}

// @Summary Seckill request
// @Tags Activity
// @Accept json
// @Produce json
// @Param body body SeckillRequest true "seckill request"
// @Success 200 {object} httpx.Response
// @Router /api/v1/activity/seckill [post]
func (h *Handler) seckill(c *gin.Context) {
	var req SeckillRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.SkuID == "" || req.UserID == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	order, err := h.svc.Seckill(req)
	if err != nil && err != ErrIdempotentHit {
		code, body := httpx.Fail(errx.CodeInternalError, err.Error())
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(order)
	c.JSON(code, body)
}

// @Summary Activity status
// @Tags Activity
// @Produce json
// @Param coupon_id query string false "coupon id"
// @Param sku_id query string false "sku id"
// @Success 200 {object} httpx.Response
// @Router /api/v1/activity/status [get]
func (h *Handler) status(c *gin.Context) {
	couponID := c.Query("coupon_id")
	skuID := c.Query("sku_id")
	if couponID != "" {
		cp, err := h.svc.GetCoupon(couponID)
		if err != nil {
			code, body := httpx.Fail(errx.CodeNotFound, errx.MsgNotFound)
			c.JSON(code, body)
			return
		}
		code, body := httpx.OK(cp)
		c.JSON(code, body)
		return
	}
	if skuID != "" {
		sk, err := h.svc.GetSeckill(skuID)
		if err != nil {
			code, body := httpx.Fail(errx.CodeNotFound, errx.MsgNotFound)
			c.JSON(code, body)
			return
		}
		code, body := httpx.OK(sk)
		c.JSON(code, body)
		return
	}
	code, body := httpx.Fail(errx.CodeInvalidRequest, "coupon_id or sku_id required")
	c.JSON(code, body)
}
