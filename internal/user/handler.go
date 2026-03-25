package user

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
	r.GET("/healthz", func(c *gin.Context) {
		code, body := httpx.OK(gin.H{"status": "ok"})
		c.JSON(code, body)
	})
	r.GET("/readyz", func(c *gin.Context) {
		code, body := httpx.OK(gin.H{"status": "ready"})
		c.JSON(code, body)
	})

	r.POST("/users", h.create)
	r.GET("/users/:id", h.get)
}

// @Summary ????
// @Tags User
// @Accept json
// @Produce json
// @Param body body CreateUserRequest true "????"
// @Success 200 {object} httpx.Response
// @Router /users [post]
func (h *Handler) create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	user, err := h.svc.Create(req)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "create user failed")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(user)
	c.JSON(code, body)
}

// @Summary ????
// @Tags User
// @Produce json
// @Param id path string true "??ID"
// @Success 200 {object} httpx.Response
// @Router /users/{id} [get]
func (h *Handler) get(c *gin.Context) {
	id := c.Param("id")
	user, err := h.svc.Get(id)
	if err != nil {
		code, body := httpx.Fail(errx.CodeNotFound, "user not found")
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(user)
	c.JSON(code, body)
}
