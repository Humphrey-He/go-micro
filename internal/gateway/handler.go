package gateway

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go-micro/internal/activity"
	"go-micro/internal/payment"
	"go-micro/internal/price"
	"go-micro/internal/refund"
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

	api := r.Group("/api/v1")
	api.POST("/auth/login", h.login)
	api.Use(middleware.JWTAuth())
	api.POST("/orders", h.createOrder)
	api.GET("/orders/:id", h.getOrder)
	api.GET("/order-views/:order_no", h.getOrderView)
	api.POST("/payments", h.createPayment)
	api.GET("/payments/:id", h.getPayment)
	api.POST("/payments/:id/success", h.markPaymentSuccess)
	api.POST("/payments/:id/failed", h.markPaymentFailed)
	api.POST("/payments/:id/timeout", h.markPaymentTimeout)
	api.POST("/refund/initiate", h.refundInitiate)
	api.POST("/refund/status", h.refundStatus)
	api.POST("/refund/rollback", h.refundRollback)
	api.POST("/activity/coupon", h.issueCoupon)
	api.POST("/activity/seckill", h.seckill)
	api.GET("/activity/status", h.activityStatus)
	api.POST("/price/calculate", h.calculatePrice)
	api.GET("/price/history", h.priceHistory)
	api.GET("/users/me", h.me)
}

// @Summary 用户登录
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body LoginRequest true "登录请求"
// @Success 200 {object} httpx.Response
// @Router /api/v1/auth/login [post]
func (h *Handler) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	resp, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		code, body := httpx.Fail(errx.CodeUnauthorized, errx.MsgInvalidCredentials)
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

// @Summary 创建订单
// @Tags Order
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body CreateOrderRequest true "订单创建请求"
// @Success 200 {object} httpx.Response
// @Router /api/v1/orders [post]
func (h *Handler) createOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.RequestID == "" || req.UserID == "" || len(req.Items) == 0 {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}

	rid, _ := c.Get(middleware.HeaderRequestID)
	resp, err := h.svc.CreateOrder(req, toString(rid))
	if err != nil {
		code, body := httpx.Fail(errx.CodeUpstreamUnavail, "order service unavailable")
		c.JSON(code, body)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary 查询订单
// @Tags Order
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "订单ID"
// @Success 200 {object} httpx.Response
// @Router /api/v1/orders/{id} [get]
func (h *Handler) getOrder(c *gin.Context) {
	id := c.Param("id")
	rid, _ := c.Get(middleware.HeaderRequestID)
	resp, err := h.svc.GetOrder(id, toString(rid))
	if err != nil {
		code, body := httpx.Fail(errx.CodeUpstreamUnavail, "order service unavailable")
		c.JSON(code, body)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary 订单聚合视图
// @Description view_status 为聚合层统一展示状态，优先级高于单一服务原始状态。
// @Description view_status 枚举：PENDING、PROCESSING、SUCCESS、FAILED、DEAD、CANCELED、TIMEOUT。
// @Tags Order
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param order_no path string true "业务订单号"
// @Success 200 {object} httpx.Response
// @Router /api/v1/order-views/{order_no} [get]
func (h *Handler) getOrderView(c *gin.Context) {
	orderNo := c.Param("order_no")
	resp, err := h.svc.GetOrderView(orderNo)
	if err != nil {
		code, body := httpx.Fail(errx.CodeUpstreamUnavail, "order view unavailable")
		c.JSON(code, body)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary 创建支付
// @Tags Payment
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body payment.CreatePaymentRequest true "创建支付请求"
// @Success 200 {object} httpx.Response
// @Router /api/v1/payments [post]
func (h *Handler) createPayment(c *gin.Context) {
	var req payment.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.OrderID == "" || req.RequestID == "" || req.Amount <= 0 {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	resp, err := h.svc.CreatePayment(req)
	if err != nil {
		code, body := httpx.Fail(errx.CodeUpstreamUnavail, "payment service unavailable")
		c.JSON(code, body)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary 查询支付
// @Tags Payment
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "支付ID"
// @Success 200 {object} httpx.Response
// @Router /api/v1/payments/{id} [get]
func (h *Handler) getPayment(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.svc.GetPayment(id)
	if err != nil {
		code, body := httpx.Fail(errx.CodeUpstreamUnavail, "payment service unavailable")
		c.JSON(code, body)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary 支付成功
// @Tags Payment
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "支付ID"
// @Success 200 {object} httpx.Response
// @Router /api/v1/payments/{id}/success [post]
func (h *Handler) markPaymentSuccess(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.svc.MarkPaymentSuccess(id)
	if err != nil {
		code, body := httpx.Fail(errx.CodeUpstreamUnavail, "payment service unavailable")
		c.JSON(code, body)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary 支付失败
// @Tags Payment
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "支付ID"
// @Success 200 {object} httpx.Response
// @Router /api/v1/payments/{id}/failed [post]
func (h *Handler) markPaymentFailed(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.svc.MarkPaymentFailed(id)
	if err != nil {
		code, body := httpx.Fail(errx.CodeUpstreamUnavail, "payment service unavailable")
		c.JSON(code, body)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary 支付超时
// @Tags Payment
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "支付ID"
// @Success 200 {object} httpx.Response
// @Router /api/v1/payments/{id}/timeout [post]
func (h *Handler) markPaymentTimeout(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.svc.MarkPaymentTimeout(id)
	if err != nil {
		code, body := httpx.Fail(errx.CodeUpstreamUnavail, "payment service unavailable")
		c.JSON(code, body)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary Initiate refund
// @Tags Refund
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body refund.InitiateRequest true "refund initiate"
// @Success 200 {object} httpx.Response
// @Router /api/v1/refund/initiate [post]
func (h *Handler) refundInitiate(c *gin.Context) {
	var req refund.InitiateRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.RefundID == "" || req.OrderID == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	resp, err := h.svc.RefundInitiate(req)
	if err != nil {
		code, body := httpx.Fail(errx.CodeUpstreamUnavail, "refund service unavailable")
		c.JSON(code, body)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary Refund status
// @Tags Refund
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body refund.StatusRequest true "refund status"
// @Success 200 {object} httpx.Response
// @Router /api/v1/refund/status [post]
func (h *Handler) refundStatus(c *gin.Context) {
	var req refund.StatusRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.RefundID == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	resp, err := h.svc.RefundStatus(req)
	if err != nil {
		code, body := httpx.Fail(errx.CodeUpstreamUnavail, "refund service unavailable")
		c.JSON(code, body)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary Refund rollback
// @Tags Refund
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body refund.RollbackRequest true "refund rollback"
// @Success 200 {object} httpx.Response
// @Router /api/v1/refund/rollback [post]
func (h *Handler) refundRollback(c *gin.Context) {
	var req refund.RollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.RefundID == "" || req.OrderID == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	resp, err := h.svc.RefundRollback(req)
	if err != nil {
		code, body := httpx.Fail(errx.CodeUpstreamUnavail, "refund service unavailable")
		c.JSON(code, body)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary Issue coupon
// @Tags Activity
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body activity.CouponRequest true "issue coupon"
// @Success 200 {object} httpx.Response
// @Router /api/v1/activity/coupon [post]
func (h *Handler) issueCoupon(c *gin.Context) {
	var req activity.CouponRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.CouponID == "" || req.UserID == "" || req.Amount <= 0 {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	resp, err := h.svc.IssueCoupon(req)
	if err != nil {
		code, body := httpx.Fail(errx.CodeUpstreamUnavail, "activity service unavailable")
		c.JSON(code, body)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary Seckill request
// @Tags Activity
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body activity.SeckillRequest true "seckill request"
// @Success 200 {object} httpx.Response
// @Router /api/v1/activity/seckill [post]
func (h *Handler) seckill(c *gin.Context) {
	var req activity.SeckillRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.SkuID == "" || req.UserID == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	resp, err := h.svc.Seckill(req)
	if err != nil {
		code, body := httpx.Fail(errx.CodeUpstreamUnavail, "activity service unavailable")
		c.JSON(code, body)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary Activity status
// @Tags Activity
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param coupon_id query string false "coupon id"
// @Param sku_id query string false "sku id"
// @Success 200 {object} httpx.Response
// @Router /api/v1/activity/status [get]
func (h *Handler) activityStatus(c *gin.Context) {
	couponID := c.Query("coupon_id")
	skuID := c.Query("sku_id")
	resp, err := h.svc.GetActivityStatus(couponID, skuID)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary Calculate price
// @Tags Price
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body price.CalculateRequest true "calculate price"
// @Success 200 {object} httpx.Response
// @Router /api/v1/price/calculate [post]
func (h *Handler) calculatePrice(c *gin.Context) {
	var req price.CalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.SkuID == "" || req.BasePrice <= 0 {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	resp, err := h.svc.CalculatePrice(req)
	if err != nil {
		code, body := httpx.Fail(errx.CodeUpstreamUnavail, "price service unavailable")
		c.JSON(code, body)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary Price history
// @Tags Price
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param sku_id query string true "sku id"
// @Param limit query int false "limit"
// @Success 200 {object} httpx.Response
// @Router /api/v1/price/history [get]
func (h *Handler) priceHistory(c *gin.Context) {
	skuID := c.Query("sku_id")
	limit := 20
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	resp, err := h.svc.GetPriceHistory(skuID, limit)
	if err != nil {
		code, body := httpx.Fail(errx.CodeUpstreamUnavail, "price service unavailable")
		c.JSON(code, body)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary 获取当前用户信息
// @Tags User
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} httpx.Response
// @Router /api/v1/users/me [get]
func (h *Handler) me(c *gin.Context) {
	userID, _ := c.Get(middleware.CtxUserID)
	username, _ := c.Get(middleware.CtxName)
	role, _ := c.Get(middleware.CtxRole)
	code, body := httpx.OK(gin.H{"user_id": userID, "username": username, "role": role})
	c.JSON(code, body)
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
