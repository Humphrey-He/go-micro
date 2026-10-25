package distribution

import (
	"errors"
	"strconv"

	"ecomm-micro/pkg/errx"
	"ecomm-micro/pkg/httpx"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc   *Service
	sched *Scheduler
	recon *Reconciler
}

func NewHandler(svc *Service, sched *Scheduler, recon *Reconciler) *Handler {
	return &Handler{svc: svc, sched: sched, recon: recon}
}

func (h *Handler) Register(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		code, body := httpx.OK(gin.H{"status": "ok"})
		c.JSON(code, body)
	})
	api := r.Group("/distribution")
	api.POST("/dispatch", h.dispatch)
	api.GET("/tasks", h.listTasks)
	api.GET("/task-counts", h.taskCounts)
	api.POST("/scheduler/tick", h.schedulerTick)
	api.POST("/reconcile/tick", h.reconcileTick)
}

// @Summary 分发商品到所有已映射平台
// @Tags Distribution
// @Accept json
// @Produce json
// @Param body body object true "{sku_id}"
// @Success 200 {object} httpx.Response
// @Router /distribution/dispatch [post]
func (h *Handler) dispatch(c *gin.Context) {
	var req struct {
		SkuID string `json:"sku_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.SkuID == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	batch, err := h.svc.Dispatch(c.Request.Context(), req.SkuID)
	if err != nil {
		if errors.Is(err, ErrNoMapping) {
			code, body := httpx.Fail(errx.CodeNotFound, "no active mapping for sku")
			c.JSON(code, body)
			return
		}
		code, body := httpx.Fail(errx.CodeInternalError, "dispatch failed")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(batch)
	c.JSON(code, body)
}

func (h *Handler) listTasks(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	tasks, err := h.svc.ListTasks(c.Request.Context(), c.Query("status"), limit)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "list tasks failed")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"tasks": tasks})
	c.JSON(code, body)
}

func (h *Handler) taskCounts(c *gin.Context) {
	counts, err := h.svc.CountByStatus(c.Request.Context())
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "count failed")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(counts)
	c.JSON(code, body)
}

// 手动触发一轮投递 / 一轮回执比对（演示用）
func (h *Handler) schedulerTick(c *gin.Context) {
	sent, queued, rejected, err := h.sched.Tick(c.Request.Context())
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "scheduler tick failed")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"sent": sent, "queued": queued, "rejected_429": rejected})
	c.JSON(code, body)
}

func (h *Handler) reconcileTick(c *gin.Context) {
	checked, missing, confirmed, err := h.recon.Tick(c.Request.Context())
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "reconcile tick failed")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"checked": checked, "missing": missing, "confirmed": confirmed})
	c.JSON(code, body)
}
