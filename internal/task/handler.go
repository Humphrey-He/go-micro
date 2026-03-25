package task

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

	r.POST("/tasks", h.create)
	r.POST("/tasks/:id/retry", h.retry)
	r.GET("/tasks/:id", h.get)
}

// @Summary ????
// @Tags Task
// @Accept json
// @Produce json
// @Param body body CreateTaskRequest true "????"
// @Success 200 {object} httpx.Response
// @Router /tasks [post]
func (h *Handler) create(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.OrderID == "" || req.BizNo == "" || req.Type == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	t, err := h.svc.Create(req)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "create task failed")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(t)
	c.JSON(code, body)
}

// @Summary ????
// @Tags Task
// @Produce json
// @Param id path string true "??ID"
// @Success 200 {object} httpx.Response
// @Router /tasks/{id} [get]
func (h *Handler) get(c *gin.Context) {
	id := c.Param("id")
	t, err := h.svc.Get(id)
	if err != nil {
		code, body := httpx.Fail(errx.CodeNotFound, "task not found")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(t)
	c.JSON(code, body)
}

// @Summary ????
// @Tags Task
// @Produce json
// @Param id path string true "??ID"
// @Success 200 {object} httpx.Response
// @Router /tasks/{id}/retry [post]
func (h *Handler) retry(c *gin.Context) {
	id := c.Param("id")
	t, err := h.svc.Retry(id)
	if err != nil {
		code, body := httpx.Fail(errx.CodeNotFound, "task not found")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(t)
	c.JSON(code, body)
}
