package social

import "time"

type Provider string

const (
	ProviderWechat Provider = "wechat"
	ProviderGoogle Provider = "google"
	ProviderApple  Provider = "apple"
)

type SocialBinding struct {
	ID             int64     `db:"id" json:"id"`
	UserID         int64     `db:"user_id" json:"user_id"`
	Provider       Provider  `db:"provider" json:"provider"`
	ProviderUserID string    `db:"provider_user_id" json:"provider_user_id"`
	AccessToken    string    `db:"access_token" json:"-"`
	RefreshToken   string    `db:"refresh_token" json:"-"`
	TokenExpiresAt time.Time `db:"token_expires_at" json:"token_expires_at,omitempty"`
	UnionID        string    `db:"union_id" json:"union_id,omitempty"`
	Nickname       string    `db:"nickname" json:"nickname,omitempty"`
	Avatar         string    `db:"avatar" json:"avatar,omitempty"`
	BindStatus     int       `db:"bind_status" json:"bind_status"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

type BindingStatus struct {
	Provider Provider `json:"provider"`
	Bound    bool     `json:"bound"`
	Nickname string   `json:"nickname,omitempty"`
	Avatar   string   `json:"avatar,omitempty"`
}

type WechatCallbackRequest struct {
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
	State        string `json:"state"`
}

type GoogleCallbackRequest struct {
	Credential string `json:"credential"`
}

type AppleCallbackRequest struct {
	IdentityToken     string `json:"identity_token"`
	AuthorizationCode string `json:"authorization_code"`
	User              string `json:"user,omitempty"`
}

type AssociatePhoneRequest struct {
	Phone  string `json:"phone"`
	Code   string `json:"code"`
	Action string `json:"action"` // bind or merge
}

type SocialLoginResponse struct {
	Token     string       `json:"token"`
	ExpiresIn int          `json:"expires_in"`
	IsNewUser bool         `json:"is_new_user"`
	UserInfo  UserInfoResp `json:"user_info"`
}

type UserInfoResp struct {
	UserID     string `json:"user_id"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	PhoneBound bool   `json:"phone_bound"`
}
