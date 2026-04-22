package notification

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

func (h *Handler) Register(r *gin.Engine) {
    api := r.Group("/api/v1")
    api.Use(middleware.JWTAuth())
    api.GET("/notifications", h.listNotifications)
    api.GET("/notifications/unread-count", h.unreadCount)
    api.PUT("/notifications/:id/read", h.markRead)
    api.PUT("/notifications/read-all", h.markAllRead)
    api.GET("/notification/configs", h.getConfigs)
    api.PUT("/notification/configs", h.updateConfigs)
}

func (h *Handler) listNotifications(c *gin.Context) {
    userID, _ := c.Get(middleware.CtxUserID)
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

    notifications, unreadCount, err := h.svc.ListNotifications(c.Request.Context(), userID.(string), page, pageSize)
    if err != nil {
        code, body := httpx.Fail(errx.CodeInternalError, "failed to get notifications")
        c.JSON(code, body)
        return
    }

    code, body := httpx.OK(gin.H{
        "notifications": notifications,
        "unread_count":  unreadCount,
        "page":          page,
        "page_size":     pageSize,
    })
    c.JSON(code, body)
}

func (h *Handler) unreadCount(c *gin.Context) {
    userID, _ := c.Get(middleware.CtxUserID)
    count, err := h.svc.GetUnreadCount(c.Request.Context(), userID.(string))
    if err != nil {
        code, body := httpx.Fail(errx.CodeInternalError, "failed to get unread count")
        c.JSON(code, body)
        return
    }
    code, body := httpx.OK(gin.H{"count": count})
    c.JSON(code, body)
}

func (h *Handler) markRead(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
        c.JSON(code, body)
        return
    }
    if err := h.svc.MarkRead(c.Request.Context(), id); err != nil {
        code, body := httpx.Fail(errx.CodeInternalError, "failed to mark read")
        c.JSON(code, body)
        return
    }
    code, body := httpx.OK(gin.H{"success": true})
    c.JSON(code, body)
}

func (h *Handler) markAllRead(c *gin.Context) {
    userID, _ := c.Get(middleware.CtxUserID)
    if err := h.svc.MarkAllRead(c.Request.Context(), userID.(string)); err != nil {
        code, body := httpx.Fail(errx.CodeInternalError, "failed to mark all read")
        c.JSON(code, body)
        return
    }
    code, body := httpx.OK(gin.H{"success": true})
    c.JSON(code, body)
}

func (h *Handler) getConfigs(c *gin.Context) {
    userID, _ := c.Get(middleware.CtxUserID)
    notifType := c.Query("type")

    cfg, err := h.svc.GetConfig(c.Request.Context(), userID.(string), notifType)
    if err != nil {
        cfg = &NotificationConfig{
            UserID:       userID.(string),
            Type:         notifType,
            EmailEnabled: true,
            PushEnabled:  true,
            Threshold:    0,
        }
    }
    code, body := httpx.OK(cfg)
    c.JSON(code, body)
}

func (h *Handler) updateConfigs(c *gin.Context) {
    var req NotificationConfig
    if err := c.ShouldBindJSON(&req); err != nil {
        code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
        c.JSON(code, body)
        return
    }
    userID, _ := c.Get(middleware.CtxUserID)
    req.UserID = userID.(string)

    if err := h.svc.UpdateConfig(c.Request.Context(), &req); err != nil {
        code, body := httpx.Fail(errx.CodeInternalError, "failed to update config")
        c.JSON(code, body)
        return
    }
    code, body := httpx.OK(gin.H{"success": true})
    c.JSON(code, body)
}