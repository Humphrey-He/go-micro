import { post } from '@/api/request'

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