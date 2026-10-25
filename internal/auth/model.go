package auth

import "go-micro/internal/user"

// SendSmsRequest 发送短信验证码请求
type SendSmsRequest struct {
	Phone string `json:"phone" binding:"required"`
	Type  string `json:"type" binding:"required"` // login | register | reset_password
}

// SmsLoginRequest 短信验证码登录请求
type SmsLoginRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Phone      string `json:"phone" binding:"required"`
	Code       string `json:"code" binding:"required"`
	Password   string `json:"password" binding:"required"`
	InviteCode string `json:"invite_code"`
}

// SendSmsResponse 发送验证码响应
type SendSmsResponse struct {
	ExpiresIn int `json:"expires_in"` // 验证码有效期（秒）
}

// LoginResponse 登录/注册成功响应，与 /auth/login 的返回结构保持一致
type LoginResponse struct {
	Token     string     `json:"token"`
	ExpiresIn int        `json:"expires_in"`
	User      *user.User `json:"user"`
}
