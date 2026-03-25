package gateway

import (
	"net/http"

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
	r.GET("/healthz", func(c *gin.Context) {
		code, body := httpx.OK(gin.H{"status": "ok"})
		c.JSON(code, body)
	})
	r.GET("/readyz", func(c *gin.Context) {
		code, body := httpx.OK(gin.H{"status": "ready"})
		c.JSON(code, body)
	})

	api := r.Group("/api/v1")
	api.Use(middleware.AuthRequired())
	api.POST("/orders", h.createOrder)
	api.GET("/orders/:id", h.getOrder)
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

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
