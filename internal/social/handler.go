package social

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

func (h *Handler) Register(r *gin.RouterGroup) {
	// OAuth 回调：登录前调用，不加鉴权
	social := r.Group("/auth/social")
	social.POST("/callback/wechat", h.wechatCallback)
	social.POST("/callback/google", h.googleCallback)
	social.POST("/callback/apple", h.appleCallback)

	// 账号绑定管理：需要登录，无论调用方分组是否已挂 JWTAuth
	authed := r.Group("/auth/social", middleware.JWTAuth())
	authed.GET("/bindings", h.getBindings)
	authed.POST("/unbind", h.unbind)
	authed.POST("/associate", h.associatePhone)
}

func (h *Handler) wechatCallback(c *gin.Context) {
	var req WechatCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "invalid request")
		c.JSON(code, body)
		return
	}

	resp, err := h.svc.WechatCallback(c.Request.Context(), req.Code, req.CodeVerifier, req.State)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "wechat callback failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

func (h *Handler) googleCallback(c *gin.Context) {
	var req GoogleCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "invalid request")
		c.JSON(code, body)
		return
	}

	resp, err := h.svc.GoogleCallback(c.Request.Context(), req.Credential)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "google callback failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

func (h *Handler) appleCallback(c *gin.Context) {
	var req AppleCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "invalid request")
		c.JSON(code, body)
		return
	}

	resp, err := h.svc.AppleCallback(c.Request.Context(), req.IdentityToken, req.AuthorizationCode, req.User)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "apple callback failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

func (h *Handler) getBindings(c *gin.Context) {
	userID, exists := c.Get(middleware.CtxUserID)
	if !exists {
		code, body := httpx.Fail(errx.CodeUnauthorized, "not authenticated")
		c.JSON(code, body)
		return
	}

	uidStr, ok := userID.(string)
	if !ok {
		code, body := httpx.Fail(errx.CodeUnauthorized, "invalid user")
		c.JSON(code, body)
		return
	}
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		code, body := httpx.Fail(errx.CodeUnauthorized, "invalid user")
		c.JSON(code, body)
		return
	}

	bindings, err := h.svc.GetUserBindings(c.Request.Context(), uid)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "get bindings failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(gin.H{"bindings": bindings})
	c.JSON(code, body)
}

func (h *Handler) unbind(c *gin.Context) {
	userID, exists := c.Get(middleware.CtxUserID)
	if !exists {
		code, body := httpx.Fail(errx.CodeUnauthorized, "not authenticated")
		c.JSON(code, body)
		return
	}

	uidStr, ok := userID.(string)
	if !ok {
		code, body := httpx.Fail(errx.CodeUnauthorized, "invalid user")
		c.JSON(code, body)
		return
	}
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		code, body := httpx.Fail(errx.CodeUnauthorized, "invalid user")
		c.JSON(code, body)
		return
	}

	var req struct {
		Provider string `json:"provider" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "invalid request")
		c.JSON(code, body)
		return
	}

	if err := h.svc.Unbind(c.Request.Context(), uid, req.Provider); err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "unbind failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(gin.H{"message": "解绑成功"})
	c.JSON(code, body)
}

func (h *Handler) associatePhone(c *gin.Context) {
	userID, exists := c.Get(middleware.CtxUserID)
	if !exists {
		code, body := httpx.Fail(errx.CodeUnauthorized, "not authenticated")
		c.JSON(code, body)
		return
	}

	uidStr, ok := userID.(string)
	if !ok {
		code, body := httpx.Fail(errx.CodeUnauthorized, "invalid user")
		c.JSON(code, body)
		return
	}
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		code, body := httpx.Fail(errx.CodeUnauthorized, "invalid user")
		c.JSON(code, body)
		return
	}

	var req AssociatePhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "invalid request")
		c.JSON(code, body)
		return
	}

	resp, err := h.svc.AssociatePhone(c.Request.Context(), uid, req.Phone, req.Code, req.Action)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, "associate phone failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(resp)
	c.JSON(code, body)
}
