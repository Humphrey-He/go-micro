package reconcile

import (
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
	r.GET("/health", func(c *gin.Context) {
		code, body := httpx.OK(gin.H{"status": "ok"})
		c.JSON(code, body)
	})
	r.POST("/reconcile/run", h.run)
	r.GET("/reconcile/diffs", h.list)
	r.POST("/reconcile/diffs/:id/resolve", h.resolve)
	r.POST("/reconcile/diffs/:id/ignore", h.ignore)
}

// @Summary 手动触发一轮订单/库存对账
// @Tags Reconcile
// @Produce json
// @Success 200 {object} httpx.Response
// @Router /reconcile/run [post]
func (h *Handler) run(c *gin.Context) {
	report, err := h.svc.RunOnce(c.Request.Context())
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "reconcile run failed")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(report)
	c.JSON(code, body)
}

// @Summary 查询差异台账
// @Tags Reconcile
// @Produce json
// @Param status query string false "PENDING/RESOLVED/IGNORED"
// @Param limit query int false "default 50"
// @Success 200 {object} httpx.Response
// @Router /reconcile/diffs [get]
func (h *Handler) list(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	diffs, err := h.svc.ListDiffs(c.Request.Context(), c.Query("status"), limit)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "list diffs failed")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"diffs": diffs})
	c.JSON(code, body)
}

// @Summary 关闭差异（PENDING → RESOLVED）
// @Tags Reconcile
// @Produce json
// @Param id path int true "差异ID"
// @Success 200 {object} httpx.Response
// @Router /reconcile/diffs/{id}/resolve [post]
func (h *Handler) resolve(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "invalid id")
		c.JSON(code, body)
		return
	}
	if err := h.svc.Resolve(c.Request.Context(), id); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "diff not pending or not found")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"id": id, "status": DiffResolved})
	c.JSON(code, body)
}

// @Summary 忽略差异（PENDING → IGNORED）
// @Tags Reconcile
// @Produce json
// @Param id path int true "差异ID"
// @Success 200 {object} httpx.Response
// @Router /reconcile/diffs/{id}/ignore [post]
func (h *Handler) ignore(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "invalid id")
		c.JSON(code, body)
		return
	}
	if err := h.svc.Ignore(c.Request.Context(), id); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "diff not pending or not found")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"id": id, "status": DiffIgnored})
	c.JSON(code, body)
}
