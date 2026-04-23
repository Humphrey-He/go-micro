import { get, post } from '@/api/request'

// 登录
export interface LoginParams {
  account: string
  password: string
  remember?: boolean
}

export interface LoginResponse {
  token: string
  expires_in: number
  user: {
    user_id: string
    phone: string
    nickname: string
    avatar: string
    level: number
    member_title: string
    points: number
  }
}

export const login = (params: LoginParams) =>
  post<LoginResponse>('/auth/login', params)

// 发送验证码
export interface SendSmsParams {
  phone: string
  type: 'login' | 'register' | 'reset_password'
}

export const sendSms = (params: SendSmsParams) =>
  post<{ expires_in: number }>('/auth/sms/send', params)

// 验证码登录
export interface SmsLoginParams {
  phone: string
  code: string
}

export const smsLogin = (params: SmsLoginParams) =>
  post<LoginResponse>('/auth/sms/login', params)

// 注册
export interface RegisterParams {
  phone: string
  code: string
  password: string
  invite_code?: string
}

export const register = (params: RegisterParams) =>
  post<LoginResponse>('/auth/register', params)

// 退出
export const logout = () => post<void>('/auth/logout')

// ============ 社交登录相关 ============

export type SocialProvider = 'wechat' | 'google' | 'apple' | 'qq'

// 社交登录响应
export interface SocialLoginResponse {
  token: string
  expires_in: number
  is_new_user: boolean
  user_info: {
    user_id: string
    nickname: string
    avatar: string
    phone_bound: boolean
  }
}

// 社交绑定状态
export interface SocialBinding {
  provider: SocialProvider
  bound: boolean
  nickname?: string
  avatar?: string
}

// 微信登录参数 (PKCE)
export interface WechatCallbackParams {
  code: string
  code_verifier: string
  state: string
}

// Google 登录参数
export interface GoogleCallbackParams {
  credential: string
}

// Apple 登录参数
export interface AppleCallbackParams {
  identity_token: string
  authorization_code: string
  user?: string
}

// 社交账号绑定查询
export const getSocialBindings = () =>
  get<{ bindings: SocialBinding[] }>('/auth/social/bindings')

// 解绑社交账号
export const unbindSocial = (provider: SocialProvider) =>
  post<void>('/auth/social/unbind', { provider })

// 关联手机号
export interface AssociatePhoneParams {
  phone: string
  code: string
  action: 'bind' | 'merge'
}

export const associatePhone = (params: AssociatePhoneParams) =>
  post<{ token: string; user_id: string }>('/auth/social/associate', params)