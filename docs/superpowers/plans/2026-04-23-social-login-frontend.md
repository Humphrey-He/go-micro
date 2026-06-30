# 社交登录前端实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现微信、Google、Apple 三大平台的社交登录前端功能，包括登录按钮、OAuth回调处理、账号绑定/解绑功能。

**Architecture:**
- 在现有 `Login.tsx` 页面增加社交登录入口（三个平台的登录按钮）
- 创建独立的 OAuth 回调处理页面 `SocialCallback.tsx`，处理各平台授权后的回调
- 社交登录 API 与现有 auth API 统一管理，扩展 `src/api/auth.ts`
- 使用 Ant Design Mobile 的 Button 组件，保持与现有 UI 一致

**Tech Stack:** React 18, TypeScript, Ant Design Mobile 5, Zustand, React Router 6

---

## File Structure

```
user-mall/src/
├── api/auth.ts                    # 扩展：增加社交登录API
├── stores/authStore.ts            # 扩展：增加社交绑定状态
├── pages/Auth/
│   ├── Login.tsx                  # 修改：增加社交登录按钮
│   ├── SocialCallback.tsx         # 创建：OAuth回调处理页
│   └── SocialBind.tsx             # 创建：账号关联页面
├── components/
│   └── SocialLoginButtons.tsx     # 创建：社交登录按钮组件
└── utils/
    └── socialLogin.ts             # 创建：社交登录工具函数
```

---

## Task 1: 创建社交登录 API 接口

**Files:**
- Modify: `user-mall/src/api/auth.ts`

- [ ] **Step 1: 添加社交登录类型定义和接口**

在 `user-mall/src/api/auth.ts` 末尾添加：

```typescript
// ============ 社交登录相关 ============

export type SocialProvider = 'wechat' | 'google' | 'apple'

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
```

- [ ] **Step 2: 添加微信 PKCE 工具函数引用**

在 `utils/socialLogin.ts` 中创建（见 Task 4）

---

## Task 2: 创建社交登录工具函数

**Files:**
- Create: `user-mall/src/utils/socialLogin.ts`

- [ ] **Step 1: 创建 PKCE 工具函数**

创建 `user-mall/src/utils/socialLogin.ts`：

```typescript
/**
 * 社交登录工具函数
 * 实现 OAuth2.0 PKCE 流程
 */

// 生成随机字符串
export const generateRandomString = (length: number = 64): string => {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~'
  const randomValues = crypto.getRandomValues(new Uint8Array(length))
  return Array.from(randomValues)
    .map(v => chars[v % chars.length])
    .join('')
}

// 计算 SHA256 hash，返回 Base64URL 编码
export const generateCodeChallenge = async (codeVerifier: string): Promise<string> => {
  const encoder = new TextEncoder()
  const data = encoder.encode(codeVerifier)
  const hash = await crypto.subtle.digest('SHA-256', data)
  return base64URLEncode(new Uint8Array(hash))
}

// Base64URL 编码
export const base64URLEncode = (bytes: Uint8Array): string => {
  let str = ''
  for (const byte of bytes) {
    str += String.fromCharCode(byte)
  }
  return btoa(str)
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '')
}

// 生成 PKCE 参数
export const generatePKCE = async () => {
  const codeVerifier = generateRandomString(64)
  const codeChallenge = await generateCodeChallenge(codeVerifier)
  return { codeVerifier, codeChallenge }
}

// 存储 PKCE verifier（临时存储，5分钟过期）
const PKCE_STORAGE_KEY = 'social_pkce_verifier'
const PKCE_STATE_KEY = 'social_pkce_state'

export const storePKCE = (verifier: string, state: string) => {
  const expiresAt = Date.now() + 5 * 60 * 1000 // 5分钟
  localStorage.setItem(PKCE_STORAGE_KEY, JSON.stringify({ verifier, expiresAt }))
  localStorage.setItem(PKCE_STATE_KEY, state)
}

export const getStoredPKCE = (): { verifier: string; state: string } | null => {
  try {
    const verifierData = localStorage.getItem(PKCE_STORAGE_KEY)
    const state = localStorage.getItem(PKCE_STATE_KEY)
    if (!verifierData || !state) return null

    const { verifier, expiresAt } = JSON.parse(verifierData)
    if (Date.now() > expiresAt) {
      clearPKCE()
      return null
    }
    return { verifier, state }
  } catch {
    return null
  }
}

export const clearPKCE = () => {
  localStorage.removeItem(PKCE_STORAGE_KEY)
  localStorage.removeItem(PKCE_STATE_KEY)
}

// 微信登录跳转
export const wechatLogin = async (redirectUri: string) => {
  const { codeVerifier, codeChallenge } = await generatePKCE()
  const state = generateRandomString(32)

  storePKCE(codeVerifier, state)

  const appId = import.meta.env.VITE_WECHAT_APP_ID || 'your_wechat_appid'
  const encodedRedirectUri = encodeURIComponent(redirectUri)

  const authUrl = `https://open.weixin.qq.com/connect/qrconnect?appid=${appId}&redirect_uri=${encodedRedirectUri}&response_type=code&scope=snsapi_login&state=${state}&code_challenge=${codeChallenge}&code_challenge_method=S256`

  window.location.href = authUrl
}

// Google 登录（使用 Google Identity Services）
export const googleLogin = () => {
  const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID || 'your_google_clientid'
  const redirectUri = `${window.location.origin}/auth/social/callback/google`

  const authUrl = `https://accounts.google.com/o/oauth2/v2/auth?client_id=${clientId}&redirect_uri=${encodeURIComponent(redirectUri)}&response_type=token&scope=openid%20email%20profile&include_granted_scopes=true&state=random_state_string`

  window.location.href = authUrl
}

// Apple 登录
export const appleLogin = () => {
  const clientId = import.meta.env.VITE_APPLE_CLIENT_ID || 'your_apple_clientid'
  const redirectUri = `${window.location.origin}/auth/social/callback/apple`

  const authUrl = `https://appleid.apple.com/auth/authorize?client_id=${clientId}&redirect_uri=${encodeURIComponent(redirectUri)}&response_type=code%20id_token&scope=name%20email&state=random_state_string`

  window.location.href = authUrl
}

// 获取环境变量类型定义
declare global {
  interface ImportMetaEnv {
    VITE_WECHAT_APP_ID?: string
    VITE_GOOGLE_CLIENT_ID?: string
    VITE_APPLE_CLIENT_ID?: string
  }
}
```

---

## Task 3: 创建社交登录按钮组件

**Files:**
- Create: `user-mall/src/components/SocialLoginButtons.tsx`

- [ ] **Step 1: 创建 SocialLoginButtons 组件**

创建 `user-mall/src/components/SocialLoginButtons/index.tsx`：

```tsx
import { Button, Toast } from 'antd-mobile'
import { wechatLogin, googleLogin, appleLogin } from '@/utils/socialLogin'

interface SocialLoginButtonsProps {
  onWechatClick?: () => void
  onGoogleClick?: () => void
  onAppleClick?: () => void
  layout?: 'horizontal' | 'vertical'
}

const WechatIcon = () => (
  <svg viewBox="0 0 24 24" className="w-5 h-5">
    <path fill="#07C160" d="M8.5 11c-.83 0-1.5-.67-1.5-1.5S7.67 8 8.5 8s1.5.67 1.5 1.5-.67 1.5-1.5 1.5zm7 0c-.83 0-1.5-.67-1.5-1.5S14.67 8 15.5 8s1.5.67 1.5 1.5-.67 1.5-1.5 1.5zM12 4c-4.97 0-9 2.69-9 6 0 1.97.94 3.74 2.44 5.02L3.5 18.5l3.5-.5c1.5 1.26 3.28 2.04 5.19 2.23l-.19 1.77 1.98-.52c1.82.19 3.69.19 5.51 0 .91-.09 1.82-.22 2.72-.4l.52-1.98-1.47-.19c1.08-.69 2.01-1.56 2.75-2.58l1.98.5-.5-2.02c1.74-1.58 2.81-3.79 2.81-6.29 0-3.31-4.03-6-9-6zm5.75 8.75c-.19.48-.44.92-.75 1.33l-.72-1.48.72.15zm-2.54 2.54c-.31.41-.66.79-1.05 1.12l-.91-1.47 1.47-.28.49 1.63zm-3.23-1.29l-1.47.91-.28-1.47 1.12-1.05 1.63.49v1.12zm-4.25 1.29c.19.48.44.92.75 1.33l-.72 1.48-.72-.15.69-2.66zm-1.63-3.25l1.47-.91.28 1.47-1.12 1.05-1.63-.49v-1.12zm2.54-2.54c.31-.41.66-.79 1.05-1.12l.91 1.47-1.47.28-.49-1.63zm5.75-1.29l-1.47-.91-.28 1.47 1.12 1.05 1.63-.49v1.12z"/>
  </svg>
)

const GoogleIcon = () => (
  <svg viewBox="0 0 24 24" className="w-5 h-5">
    <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
    <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
    <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
    <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
  </svg>
)

const AppleIcon = () => (
  <svg viewBox="0 0 24 24" className="w-5 h-5">
    <path fill="#000" d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.81-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5.13 1.17-.34 2.35-1.04 3.19-.69.85-1.83 1.51-2.95 1.42-.15-1.15.41-2.35 1.05-3.11z"/>
  </svg>
)

export default function SocialLoginButtons({
  onWechatClick,
  onGoogleClick,
  onAppleClick,
  layout = 'vertical'
}: SocialLoginButtonsProps) {
  const handleWechat = () => {
    if (onWechatClick) {
      onWechatClick()
    } else {
      try {
        wechatLogin(`${window.location.origin}/auth/social/callback/wechat`)
      } catch {
        Toast.show('微信登录暂不可用')
      }
    }
  }

  const handleGoogle = () => {
    if (onGoogleClick) {
      onGoogleClick()
    } else {
      try {
        googleLogin()
      } catch {
        Toast.show('Google 登录暂不可用')
      }
    }
  }

  const handleApple = () => {
    if (onAppleClick) {
      onAppleClick()
    } else {
      try {
        appleLogin()
      } catch {
        Toast.show('Apple 登录暂不可用')
      }
    }
  }

  const buttonStyle = layout === 'horizontal'
    ? 'flex-1'
    : 'w-full'

  return (
    <div className={`flex ${layout === 'horizontal' ? 'gap-3' : 'flex-col gap-3'}`}>
      {/* 微信登录 */}
      <Button
        className={`${buttonStyle} !bg-[#07C160] !border-[#07C160]`}
        color="primary"
        fill="outline"
        onClick={handleWechat}
      >
        <span className="flex items-center justify-center gap-2">
          <WechatIcon />
          <span className="text-[#07C160]">微信</span>
        </span>
      </Button>

      {/* Google 登录 */}
      <Button
        className={`${buttonStyle} !bg-white !border-gray-300`}
        color="primary"
        fill="outline"
        onClick={handleGoogle}
      >
        <span className="flex items-center justify-center gap-2">
          <GoogleIcon />
          <span className="text-gray-700">Google</span>
        </span>
      </Button>

      {/* Apple 登录 */}
      <Button
        className={`${buttonStyle} !bg-black !border-black`}
        color="primary"
        fill="outline"
        onClick={handleApple}
      >
        <span className="flex items-center justify-center gap-2">
          <AppleIcon />
          <span className="text-black">Apple</span>
        </span>
      </Button>
    </div>
  )
}
```

- [ ] **Step 2: 创建组件索引文件**

创建 `user-mall/src/components/SocialLoginButtons/index.tsx` 已在上面完成，无需额外索引文件

---

## Task 4: 更新登录页面（增加社交登录入口）

**Files:**
- Modify: `user-mall/src/pages/Auth/Login.tsx`

- [ ] **Step 1: 添加社交登录区域**

在 `user-mall/src/pages/Auth/Login.tsx` 中，在登录表单下方、用户协议上方添加社交登录区域：

找到登录表单结束位置 (`</Form>` 之后) 和用户协议 div 之间的位置，添加：

```tsx
{/* 社交登录 */}
<div className="mt-8">
  <div className="relative">
    <div className="absolute inset-0 flex items-center">
      <div className="w-full border-t border-gray-200" />
    </div>
    <div className="relative flex justify-center text-sm">
      <span className="px-2 bg-white text-gray-500">其他登录方式</span>
    </div>
  </div>
  <div className="mt-4">
    <SocialLoginButtons />
  </div>
</div>
```

并在文件顶部添加导入：

```tsx
import SocialLoginButtons from '@/components/SocialLoginButtons'
```

---

## Task 5: 创建 OAuth 回调处理页面

**Files:**
- Create: `user-mall/src/pages/Auth/SocialCallback.tsx`
- Modify: `user-mall/src/App.tsx` (添加路由)

- [ ] **Step 1: 创建 SocialCallback 页面**

创建 `user-mall/src/pages/Auth/SocialCallback.tsx`：

```tsx
import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { SpinLoading, Toast } from 'antd-mobile'
import { post } from '@/api/request'
import { useAuthStore } from '@/stores/authStore'
import { getStoredPKCE, clearPKCE } from '@/utils/socialLogin'
import type { SocialProvider } from '@/api/auth'

interface CallbackParams {
  code?: string
  state?: string
  error?: string
  error_description?: string
}

export default function SocialCallback() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { login: setAuth } = useAuthStore()
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading')
  const [errorMsg, setErrorMsg] = useState('')

  useEffect(() => {
    const handleCallback = async () => {
      // 获取 URL 参数
      const params: CallbackParams = {
        code: searchParams.get('code') || undefined,
        state: searchParams.get('state') || undefined,
        error: searchParams.get('error') || undefined,
        error_description: searchParams.get('error_description') || undefined,
      }

      // 错误处理
      if (params.error) {
        setStatus('error')
        setErrorMsg(params.error_description || '授权失败')
        Toast.show(params.error_description || '授权失败')
        setTimeout(() => navigate('/login'), 2000)
        return
      }

      // 判断是哪个平台
      const path = window.location.pathname
      let provider: SocialProvider
      let callbackParams: Record<string, unknown>

      if (path.includes('/wechat')) {
        provider = 'wechat'
        // 微信需要 PKCE
        const pkce = getStoredPKCE()
        if (!pkce || pkce.state !== params.state) {
          setStatus('error')
          setErrorMsg('State 验证失败，请重试')
          Toast.show('授权验证失败')
          clearPKCE()
          setTimeout(() => navigate('/login'), 2000)
          return
        }
        callbackParams = {
          code: params.code,
          code_verifier: pkce.verifier,
          state: params.state,
        }
        clearPKCE()
      } else if (path.includes('/google')) {
        provider = 'google'
        // Google 通过 URL hash 传递 token
        const hashParams = new URLSearchParams(window.location.hash.substring(1))
        const credential = hashParams.get('access_token')
        if (!credential) {
          setStatus('error')
          setErrorMsg('获取凭证失败')
          Toast.show('授权失败')
          setTimeout(() => navigate('/login'), 2000)
          return
        }
        callbackParams = { credential }
        // 清理 hash
        window.history.replaceState(null, '', window.location.pathname)
      } else if (path.includes('/apple')) {
        provider = 'apple'
        // Apple 通过 URL 参数传递
        const identityToken = searchParams.get('id_token') || ''
        const authorizationCode = searchParams.get('code') || ''
        const user = searchParams.get('user') || undefined
        callbackParams = { identity_token: identityToken, authorization_code: authorizationCode, user }
      } else {
        setStatus('error')
        setErrorMsg('未知的登录平台')
        Toast.show('未知错误')
        setTimeout(() => navigate('/login'), 2000)
        return
      }

      try {
        // 调用后端社交登录接口
        const response = await post<{
          token: string
          expires_in: number
          is_new_user: boolean
          user_info: {
            user_id: string
            nickname: string
            avatar: string
            phone_bound: boolean
          }
        }>(`/auth/social/callback/${provider}`, callbackParams)

        // 登录成功
        setAuth(response.token, response.user_info)
        setStatus('success')
        Toast.show('登录成功')

        // 跳转
        if (!response.user_info.phone_bound) {
          // 新用户或未绑定手机号，跳转到绑定手机号页面
          navigate('/auth/social/bind', { state: { isNewUser: response.is_new_user } })
        } else {
          navigate('/')
        }
      } catch (error) {
        setStatus('error')
        setErrorMsg('登录失败，请重试')
        Toast.show('登录失败，请重试')
        setTimeout(() => navigate('/login'), 2000)
      }
    }

    handleCallback()
  }, [searchParams, navigate, setAuth])

  return (
    <div className="min-h-screen bg-white flex flex-col items-center justify-center">
      {status === 'loading' && (
        <>
          <SpinLoading color="primary" />
          <p className="mt-4 text-gray-500">正在处理登录...</p>
        </>
      )}
      {status === 'error' && (
        <>
          <div className="text-6xl mb-4">😢</div>
          <p className="text-gray-700 font-medium">{errorMsg}</p>
          <p className="text-gray-400 text-sm mt-2">即将跳转到登录页...</p>
        </>
      )}
      {status === 'success' && (
        <>
          <div className="text-6xl mb-4">🎉</div>
          <p className="text-gray-700 font-medium">登录成功</p>
          <p className="text-gray-400 text-sm mt-2">正在跳转...</p>
        </>
      )}
    </div>
  )
}
```

- [ ] **Step 2: 添加路由**

在 `user-mall/src/App.tsx` 中：

添加导入：
```tsx
const SocialCallback = lazy(() => import('./pages/Auth/SocialCallback'))
const SocialBind = lazy(() => import('./pages/Auth/SocialBind'))
```

在 routes 数组中添加：
```tsx
// 社交登录回调
{ path: '/auth/social/callback/wechat', component: SocialCallback, layout: BlankLayout },
{ path: '/auth/social/callback/google', component: SocialCallback, layout: BlankLayout },
{ path: '/auth/social/callback/apple', component: SocialCallback, layout: BlankLayout },
{ path: '/auth/social/bind', component: SocialBind, layout: BlankLayout },
```

---

## Task 6: 创建账号关联页面

**Files:**
- Create: `user-mall/src/pages/Auth/SocialBind.tsx`

- [ ] **Step 1: 创建 SocialBind 页面**

创建 `user-mall/src/pages/Auth/SocialBind.tsx`：

```tsx
import { useState } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { Button, Form, Input, Toast } from 'antd-mobile'
import { associatePhone, sendSms } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'

export default function SocialBind() {
  const navigate = useNavigate()
  const location = useLocation()
  const { login: setAuth } = useAuthStore()
  const [loading, setLoading] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const [isNewUser, setIsNewUser] = useState(
    (location.state as { isNewUser?: boolean })?.isNewUser ?? true
  )

  const handleSendCode = async (phone: string) => {
    if (!phone) {
      Toast.show('请输入手机号')
      return
    }
    try {
      await sendSms({ phone, type: 'bind' })
      Toast.show('验证码已发送')
      setCountdown(60)
      const timer = setInterval(() => {
        setCountdown((prev) => {
          if (prev <= 1) {
            clearInterval(timer)
            return 0
          }
          return prev - 1
        })
      }, 1000)
    } catch {
      Toast.show('发送失败，请稍后重试')
    }
  }

  const handleSubmit = async (values: { phone: string; code: string }) => {
    try {
      setLoading(true)
      const response = await associatePhone({
        phone: values.phone,
        code: values.code,
        action: 'bind',
      })
      // 更新 token（关联成功后用新的 token）
      const userInfo = useAuthStore.getState().userInfo
      if (userInfo) {
        setAuth(response.token, userInfo)
      }
      Toast.show('绑定成功')
      navigate('/')
    } catch {
      Toast.show('绑定失败，请检查验证码')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-white p-6">
      <div className="mt-8 mb-8">
        <h1 className="text-2xl font-bold text-gray-900">
          {isNewUser ? '完善手机号' : '绑定手机号'}
        </h1>
        <p className="text-gray-500 mt-2">
          {isNewUser
            ? '为提供更好的服务，请绑定您的手机号'
            : '绑定手机号后可使用手机号登录'}
        </p>
      </div>

      <Form layout="vertical" onFinish={handleSubmit}>
        <Form.Item
          name="phone"
          rules={[
            { required: true, message: '请输入手机号' },
            { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确' },
          ]}
        >
          <Input placeholder="请输入手机号" className="py-3" />
        </Form.Item>

        <Form.Item
          name="code"
          rules={[{ required: true, message: '请输入验证码' }]}
          extra={
            <span
              className={`text-sm ${countdown > 0 ? 'text-gray-400' : 'text-primary-500'}`}
              onClick={() => {
                if (countdown === 0) {
                  const phone = (document.querySelector('input[placeholder="请输入手机号"]') as HTMLInputElement)?.value
                  handleSendCode(phone || '')
                }
              }}
            >
              {countdown > 0 ? `${countdown}s后重发` : '获取验证码'}
            </span>
          }
        >
          <Input placeholder="请输入验证码" className="py-3" />
        </Form.Item>

        <Form.Item>
          <Button block color="primary" size="large" loading={loading} type="submit">
            绑定
          </Button>
        </Form.Item>
      </Form>

      <div className="mt-6 text-center text-sm text-gray-400">
        <span
          className="text-primary-500"
          onClick={() => navigate('/')}
        >
          跳过
        </span>
        {' '}，后续可在设置中绑定
      </div>
    </div>
  )
}
```

---

## Task 7: 更新 authStore 支持社交登录

**Files:**
- Modify: `user-mall/src/stores/authStore.ts`

- [ ] **Step 1: 扩展 AuthState 类型**

在 `user-mall/src/stores/authStore.ts` 中，更新 `UserInfo` 接口和 `AuthState`：

```typescript
// 新增字段
interface UserInfo {
  user_id: string
  phone: string
  nickname: string
  avatar: string
  level: number
  member_title: string
  points: number
  phone_bound?: boolean  // 新增：是否已绑定手机号
  is_new_user?: boolean // 新增：是否新用户
}

// SocialBinding 类型
export interface SocialBinding {
  provider: 'wechat' | 'google' | 'apple'
  bound: boolean
  nickname?: string
  avatar?: string
}

// 扩展 AuthState
interface AuthState {
  // ... 现有字段
  socialBindings: SocialBinding[]
  setSocialBindings: (bindings: SocialBinding[]) => void
}
```

- [ ] **Step 2: 添加新方法**

在 `persist` 的 `set` 函数中添加：

```typescript
socialBindings: [],

setSocialBindings: (bindings) => set({ socialBindings: bindings }),
```

---

## Task 8: 环境变量配置

**Files:**
- Modify: `user-mall/src/env.d.ts` (或创建 `user-mall/src/vite-env.d.ts`)

- [ ] **Step 1: 添加环境变量类型定义**

创建或修改 `user-mall/src/vite-env.d.ts`：

```typescript
/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_WECHAT_APP_ID: string
  readonly VITE_GOOGLE_CLIENT_ID: string
  readonly VITE_APPLE_CLIENT_ID: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
```

创建或修改 `.env.example` 文件：

```bash
# 社交登录 App ID 配置
VITE_WECHAT_APP_ID=your_wechat_appid
VITE_GOOGLE_CLIENT_ID=your_google_clientid.apps.googleusercontent.com
VITE_APPLE_CLIENT_ID=com.yourcompany.yourapp
```

---

## Task 9: 验证与测试

**Files:**
- Modify: `user-mall/.env` (本地开发环境)

- [ ] **Step 1: 本地环境变量配置**

创建 `user-mall/.env` 文件（如果不存在），添加：

```bash
# 社交登录配置（开发环境使用测试 ID）
VITE_WECHAT_APP_ID=wx_test_appid
VITE_GOOGLE_CLIENT_ID=test_clientid.apps.googleusercontent.com
VITE_APPLE_CLIENT_ID=com.test.app
```

- [ ] **Step 2: 验证开发服务器启动**

```bash
cd user-mall && npm run dev
```

访问 `/login` 页面，确认：
1. 登录页显示三个社交登录按钮
2. 点击各按钮不报 JS 错误（即使后端未实现）
3. 页面布局正常

---

## Task 10: 提交代码

- [ ] **Step 1: 提交所有更改**

```bash
git add -A
git commit -m "$(cat <<'EOF'
feat(auth): add social login UI (WeChat, Google, Apple)

- Add SocialLoginButtons component with OAuth login buttons
- Add OAuth callback handler page (SocialCallback)
- Add phone binding page for new social users
- Extend auth API with social login endpoints
- Extend authStore with social binding state
- Add PKCE utilities for WeChat OAuth

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Spec Coverage Checklist

| Spec Section | Task | Status |
|-------------|------|--------|
| 微信登录 UI 按钮 | Task 3 | ✅ |
| Google 登录 UI 按钮 | Task 3 | ✅ |
| Apple 登录 UI 按钮 | Task 3 | ✅ |
| OAuth 回调处理 | Task 5 | ✅ |
| 账号关联/绑定手机号 | Task 6 | ✅ |
| 社交绑定状态管理 | Task 7 | ✅ |
| PKCE 流程工具 | Task 2 | ✅ |
| 环境变量配置 | Task 8 | ✅ |
| 前端路由配置 | Task 5 | ✅ |

---

## Execution Options

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?