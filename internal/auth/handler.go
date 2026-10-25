package auth

import (
	"errors"
	"strings"

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

// RegisterPublic 注册无需登录的认证路由（验证码、短信登录、注册）。
// 必须挂在 JWTAuth 中间件之前/之外。
func (h *Handler) RegisterPublic(rg *gin.RouterGroup) {
	rg.POST("/auth/sms/send", h.sendSms)
	rg.POST("/auth/sms/login", h.smsLogin)
	rg.POST("/auth/register", h.register)
}

// RegisterAuthed 注册需要登录的认证路由（登出）。
// rg 所在分组必须已应用 JWTAuth 中间件。
func (h *Handler) RegisterAuthed(rg *gin.RouterGroup) {
	rg.POST("/auth/logout", h.logout)
}

// @Summary 发送短信验证码
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body SendSmsRequest true "发送验证码请求 {phone, type: login|register|reset_password}"
// @Success 200 {object} httpx.Response
// @Router /api/v1/auth/sms/send [post]
func (h *Handler) sendSms(c *gin.Context) {
	var req SendSmsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	resp, err := h.svc.SendCode(c.Request.Context(), req.Phone, req.Type)
	if err != nil {
		code, body := httpx.Fail(smsErrCode(err), smsErrMsg(err))
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

// @Summary 短信验证码登录
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body SmsLoginRequest true "短信登录请求 {phone, code}"
// @Success 200 {object} httpx.Response
// @Router /api/v1/auth/sms/login [post]
func (h *Handler) smsLogin(c *gin.Context) {
	var req SmsLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	resp, err := h.svc.SmsLogin(c.Request.Context(), req.Phone, req.Code)
	if err != nil {
		code, body := httpx.Fail(smsErrCode(err), smsErrMsg(err))
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

// @Summary 用户注册
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body RegisterRequest true "注册请求 {phone, code, password, invite_code?}"
// @Success 200 {object} httpx.Response
// @Router /api/v1/auth/register [post]
func (h *Handler) register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
		c.JSON(code, body)
		return
	}
	resp, err := h.svc.Register(c.Request.Context(), req.Phone, req.Code, req.Password, req.InviteCode)
	if err != nil {
		code, body := httpx.Fail(smsErrCode(err), smsErrMsg(err))
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

// @Summary 登出
// @Tags Auth
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} httpx.Response
// @Router /api/v1/auth/logout [post]
func (h *Handler) logout(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	tokenStr := strings.TrimPrefix(auth, "Bearer ")
	if tokenStr == "" {
		code, body := httpx.Fail(errx.CodeUnauthorized, errx.MsgUnauthorized)
		c.JSON(code, body)
		return
	}
	if err := h.svc.Logout(c.Request.Context(), tokenStr); err != nil {
		code, body := httpx.Fail(errx.CodeInternalError, errx.MsgInternalError)
		c.JSON(code, body)
		return
	}
	code, body := httpx.OK(gin.H{"message": "logged out"})
	c.JSON(code, body)
}

func smsErrCode(err error) int {
	switch {
	case errors.Is(err, ErrInvalidPhone),
		errors.Is(err, ErrInvalidCodeType),
		errors.Is(err, ErrInvalidPassword),
		errors.Is(err, ErrInvalidInviteCode):
		return errx.CodeInvalidRequest
	case errors.Is(err, ErrSendCooldown):
		return errx.CodeSmsSendTooFrequent
	case errors.Is(err, ErrDailyLimit):
		return errx.CodeSmsDailyLimit
	case errors.Is(err, ErrCodeInvalid):
		return errx.CodeSmsCodeInvalid
	case errors.Is(err, ErrCodeExpired):
		return errx.CodeSmsCodeExpired
	case errors.Is(err, ErrTooManyAttempts):
		return errx.CodeSmsTooManyAttempts
	case errors.Is(err, ErrUserNotFound):
		return errx.CodeUserNotFound
	case errors.Is(err, ErrUserAlreadyExists):
		return errx.CodeUserAlreadyExists
	case errors.Is(err, ErrUserDisabled):
		return errx.CodeUserDisabled
	default:
		return errx.CodeInternalError
	}
}

func smsErrMsg(err error) string {
	switch {
	case errors.Is(err, ErrInvalidPhone):
		return "invalid phone number"
	case errors.Is(err, ErrInvalidCodeType):
		return "invalid sms code type"
	case errors.Is(err, ErrInvalidPassword):
		return "password must be 6-64 characters"
	case errors.Is(err, ErrInvalidInviteCode):
		return "invalid invite code"
	case errors.Is(err, ErrSendCooldown):
		return errx.MsgSmsSendTooFrequent
	case errors.Is(err, ErrDailyLimit):
		return errx.MsgSmsDailyLimit
	case errors.Is(err, ErrCodeInvalid):
		return errx.MsgSmsCodeInvalid
	case errors.Is(err, ErrCodeExpired):
		return errx.MsgSmsCodeExpired
	case errors.Is(err, ErrTooManyAttempts):
		return errx.MsgSmsTooManyAttempts
	case errors.Is(err, ErrUserNotFound):
		return "phone not registered, please register first"
	case errors.Is(err, ErrUserAlreadyExists):
		return "phone already registered, please login"
	case errors.Is(err, ErrUserDisabled):
		return errx.MsgUserDisabled
	default:
		return errx.MsgInternalError
	}
}
