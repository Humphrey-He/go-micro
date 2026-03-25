package gateway

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go-micro/internal/payment"
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
	r.GET("/healthz", func(c *gin.Context) {
		code, body := httpx.OK(gin.H{"status": "ok"})
		c.JSON(code, body)
	})
	r.GET("/readyz", func(c *gin.Context) {
		code, body := httpx.OK(gin.H{"status": "ready"})
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
	api.GET("/users/me", h.me)
}

// @Summary ??
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body LoginRequest true "??"
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
		code, body := httpx.Fail(errx.CodeUnauthorized, "invalid username or password")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

// @Summary ????
// @Tags Order
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body CreateOrderRequest true "????"
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

// @Summary ????
// @Tags Order
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "??ID"
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

// @Summary ??????
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
