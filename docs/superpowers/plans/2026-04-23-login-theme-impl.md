# 登录页面多主题实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为商城登录页面实现三套可切换的主题（兴趣电商/大厂稳重/轻奢简约），支持 QQ 登录和个性化商品展示

**Architecture:** 基于 Zustand 管理主题状态，React Context 提供主题上下文，组件根据当前主题渲染不同布局。主题配置集中管理，支持 localStorage 持久化。

**Tech Stack:** React 18, TypeScript, Zustand, Tailwind CSS, antd-mobile

---

## 文件结构

```
user-mall/src/
├── styles/login-themes/
│   └── theme-config.ts              # 主题配置常量
├── stores/
│   └── themeStore.ts                # 主题状态管理
├── context/
│   └── ThemeContext.tsx             # 主题上下文
├── components/ThemeSwitcher/
│   └── index.tsx                    # 主题切换组件
├── components/ProductShowcase/
│   └── index.tsx                    # 商品展示组件
├── utils/
│   └── socialLogin.ts               # 社交登录工具（修改）
├── api/
│   └── recommendation.ts             # 添加 login-showcase API
└── pages/Auth/
    ├── Login.tsx                    # 主登录页（重构）
    ├── LoginVibrant.tsx             # 兴趣电商主题
    ├── LoginStable.tsx              # 大厂稳重主题
    └── LoginLuxury.tsx              # 轻奢简约主题
```

---

## Phase 1: 基础架构

### Task 1: 创建主题配置文件

**Files:**
- Create: `user-mall/src/styles/login-themes/theme-config.ts`

- [ ] **Step 1: 创建主题配置目录和文件**

```typescript
// user-mall/src/styles/login-themes/theme-config.ts

export type LoginTheme = 'vibrant' | 'stable' | 'luxury'

export interface ThemeColors {
  primary: string
  secondary: string
  accent: string
  background: string
  backgroundGradient: string
  cardBg: string
  text: string
  textSecondary: string
  border: string
}

export interface ThemeConfig {
  id: LoginTheme
  name: string
  nameCn: string
  colors: ThemeColors
  cardRadius: number
  socialStyle: 'colorful' | 'subtle' | 'outline'
}

export const themes: Record<LoginTheme, ThemeConfig> = {
  vibrant: {
    id: 'vibrant',
    name: 'Youth/Vibrant',
    nameCn: '兴趣电商',
    colors: {
      primary: '#FF6B35',
      secondary: '#F7C948',
      accent: '#00D9C0',
      background: '#FFFFFF',
      backgroundGradient: 'linear-gradient(135deg, #FF6B35 0%, #FF4D94 100%)',
      cardBg: '#FFFFFF',
      text: '#1F2937',
      textSecondary: '#6B7280',
      border: '#E5E7EB',
    },
    cardRadius: 24,
    socialStyle: 'colorful',
  },
  stable: {
    id: 'stable',
    name: 'Amazon Stable',
    nameCn: '大厂稳重',
    colors: {
      primary: '#007185',
      secondary: '#FFA41C',
      accent: '#00A4A4',
      background: '#FFFFFF',
      backgroundGradient: 'linear-gradient(180deg, #F7F7F7 0%, #FFFFFF 100%)',
      cardBg: '#FFFFFF',
      text: '#0F1111',
      textSecondary: '#565959',
      border: '#E7E7E7',
    },
    cardRadius: 8,
    socialStyle: 'subtle',
  },
  luxury: {
    id: 'luxury',
    name: 'Luxury/Minimalist',
    nameCn: '轻奢简约',
    colors: {
      primary: '#1A1A1A',
      secondary: '#C9A96E',
      accent: '#E8E8E8',
      background: '#FAFAFA',
      backgroundGradient: 'linear-gradient(180deg, #FAFAFA 0%, #F5F5F5 100%)',
      cardBg: '#FFFFFF',
      text: '#1A1A1A',
      textSecondary: '#6B6B6B',
      border: '#E8E8E8',
    },
    cardRadius: 4,
    socialStyle: 'outline',
  },
}
```

- [ ] **Step 2: 提交**

```bash
git add user-mall/src/styles/login-themes/theme-config.ts
git commit -m "feat: add login theme configuration"
```

---

### Task 2: 创建主题状态管理 Store

**Files:**
- Create: `user-mall/src/stores/themeStore.ts`

- [ ] **Step 1: 创建 themeStore**

```typescript
// user-mall/src/stores/themeStore.ts
import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { LoginTheme } from '@/styles/login-themes/theme-config'

interface ThemeState {
  theme: LoginTheme
  setTheme: (theme: LoginTheme) => void
  toggleTheme: () => void
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set, get) => ({
      theme: 'vibrant',

      setTheme: (theme) => set({ theme }),

      toggleTheme: () => {
        const themeOrder: LoginTheme[] = ['vibrant', 'stable', 'luxury']
        const current = get().theme
        const index = themeOrder.indexOf(current)
        const next = themeOrder[(index + 1) % themeOrder.length]
        set({ theme: next })
      },
    }),
    {
      name: 'login-theme-storage',
    }
  )
)
```

- [ ] **Step 2: 提交**

```bash
git add user-mall/src/stores/themeStore.ts
git commit -m "feat: add login theme zustand store"
```

---

### Task 3: 创建 ThemeContext

**Files:**
- Create: `user-mall/src/context/ThemeContext.tsx`

- [ ] **Step 1: 创建 ThemeContext**

```tsx
// user-mall/src/context/ThemeContext.tsx
import { createContext, useContext } from 'react'
import { useThemeStore } from '@/stores/themeStore'
import { themes, type LoginTheme } from '@/styles/login-themes/theme-config'

interface ThemeContextType {
  theme: LoginTheme
  setTheme: (theme: LoginTheme) => void
  config: (typeof themes)['vibrant']
}

export const ThemeContext = createContext<ThemeContextType>({
  theme: 'vibrant',
  setTheme: () => {},
  config: themes.vibrant,
})

export const useLoginTheme = () => {
  const store = useThemeStore()
  return {
    theme: store.theme,
    setTheme: store.setTheme,
    config: themes[store.theme],
  }
}
```

- [ ] **Step 2: 创建 context 目录并提交**

```bash
mkdir -p user-mall/src/context
git add user-mall/src/context/ThemeContext.tsx
git commit -m "feat: add login theme React context"
```

---

### Task 4: 创建 ThemeSwitcher 组件

**Files:**
- Create: `user-mall/src/components/ThemeSwitcher/index.tsx`

- [ ] **Step 1: 创建 ThemeSwitcher 组件**

```tsx
// user-mall/src/components/ThemeSwitcher/index.tsx
import { useLoginTheme } from '@/context/ThemeContext'
import { themes } from '@/styles/login-themes/theme-config'

export const ThemeSwitcher: React.FC = () => {
  const { theme, setTheme } = useLoginTheme()

  return (
    <div
      className="fixed top-4 right-4 z-50 flex items-center gap-3 px-4 py-2 rounded-full bg-white/90 backdrop-blur-sm shadow-lg"
      style={{ boxShadow: '0 4px 20px rgba(0,0,0,0.1)' }}
    >
      {Object.values(themes).map((t) => (
        <button
          key={t.id}
          onClick={() => setTheme(t.id)}
          className={`relative w-8 h-8 rounded-full transition-all duration-300 ${
            theme === t.id ? 'scale-110 ring-2 ring-offset-2' : 'opacity-60 hover:opacity-100'
          }`}
          style={{
            background: `linear-gradient(135deg, ${t.colors.primary}, ${t.colors.secondary})`,
            ringColor: t.colors.primary,
          }}
          title={t.nameCn}
        >
          {theme === t.id && (
            <span className="absolute inset-0 flex items-center justify-center text-white text-xs font-bold">
              {t.nameCn.charAt(0)}
            </span>
          )}
        </button>
      ))}
    </div>
  )
}
```

- [ ] **Step 2: 提交**

```bash
git add user-mall/src/components/ThemeSwitcher/index.tsx
git commit -m "feat: add theme switcher component"
```

---

## Phase 2: 社交登录增强

### Task 5: 添加 QQ 登录到 SocialLoginButtons

**Files:**
- Modify: `user-mall/src/components/SocialLoginButtons/index.tsx`
- Modify: `user-mall/src/utils/socialLogin.ts`
- Modify: `user-mall/src/api/auth.ts`

- [ ] **Step 1: 更新 SocialLoginButtons 添加 QQ 图标和属性**

```tsx
// user-mall/src/components/SocialLoginButtons/index.tsx
import { Toast } from 'antd-mobile'
import { wechatLogin, googleLogin, appleLogin, qqLogin } from '@/utils/socialLogin'
import type { LoginTheme } from '@/styles/login-themes/theme-config'

interface SocialLoginButtonsProps {
  theme?: LoginTheme
  layout?: 'horizontal' | 'vertical'
  onWechatClick?: () => void
  onGoogleClick?: () => void
  onAppleClick?: () => void
  onQQClick?: () => void
}

const WechatIcon = () => (
  <svg viewBox="0 0 24 24" className="w-6 h-6">
    <path fill="#07C160" d="M8.5 11c-.83 0-1.5-.67-1.5-1.5S7.67 8 8.5 8s1.5.67 1.5 1.5-.67 1.5-1.5 1.5zm7 0c-.83 0-1.5-.67-1.5-1.5S14.67 8 15.5 8s1.5.67 1.5 1.5-.67 1.5-1.5 1.5zM12 4c-4.97 0-9 2.69-9 6 0 1.97.94 3.74 2.44 5.02L3.5 18.5l3.5-.5c1.5 1.26 3.28 2.04 5.19 2.23l-.19 1.77 1.98-.52c1.82.19 3.69.19 5.51 0 .91-.09 1.82-.22 2.72-.4l.52-1.98-1.47-.19c1.08-.69 2.01-1.56 2.75-2.58l1.98.5-.5-2.02c1.74-1.58 2.81-3.79 2.81-6.29 0-3.31-4.03-6-9-6z"/>
  </svg>
)

const GoogleIcon = () => (
  <svg viewBox="0 0 24 24" className="w-6 h-6">
    <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
    <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
    <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
    <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
  </svg>
)

const AppleIcon = () => (
  <svg viewBox="0 0 24 24" className="w-6 h-6">
    <path fill="#000" d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.81-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5.13 1.17-.34 2.35-1.04 3.19-.69.85-1.83 1.51-2.95 1.42-.15-1.15.41-2.35 1.05-3.11z"/>
  </svg>
)

const QQIcon = () => (
  <svg viewBox="0 0 24 24" className="w-6 h-6">
    <path fill="#12B7F5" d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z"/>
    <path fill="#12B7F5" d="M12 6c-3.31 0-6 2.69-6 6s2.69 6 6 6 6-2.69 6-6-2.69-6-6-6zm0 10c-2.21 0-4-1.79-4-4s1.79-4 4-4 4 1.79 4 4-1.79 4-4 4z"/>
    <circle fill="#12B7F5" cx="8" cy="11" r="1.5"/>
    <circle fill="#12B7F5" cx="16" cy="11" r="1.5"/>
  </svg>
)

const getSocialButtonClass = (socialStyle: string, provider: string) => {
  switch (socialStyle) {
    case 'colorful':
      return `w-12 h-12 rounded-full flex items-center justify-center text-white shadow-md hover:shadow-lg hover:scale-105 transition-all duration-200 active:scale-95`
    case 'subtle':
      return `w-10 h-10 rounded flex items-center justify-center bg-gray-100 text-gray-600 hover:bg-gray-200 transition-colors`
    case 'outline':
      return `w-10 h-10 rounded-full border-2 flex items-center justify-center text-gray-500 hover:border-[#C9A96E] hover:text-[#C9A96E] transition-all`
    default:
      return ''
  }
}

export default function SocialLoginButtons({
  theme = 'vibrant',
  layout = 'horizontal',
  onWechatClick,
  onGoogleClick,
  onAppleClick,
  onQQClick,
}: SocialLoginButtonsProps) {
  const socialConfig = themes[theme]

  const handleWechat = () => {
    if (onWechatClick) onWechatClick()
    else {
      try {
        wechatLogin(`${window.location.origin}/auth/social/callback/wechat`)
      } catch {
        Toast.show('微信登录暂不可用')
      }
    }
  }

  const handleGoogle = () => {
    if (onGoogleClick) onGoogleClick()
    else {
      try {
        googleLogin()
      } catch {
        Toast.show('Google 登录暂不可用')
      }
    }
  }

  const handleApple = () => {
    if (onAppleClick) onAppleClick()
    else {
      try {
        appleLogin()
      } catch {
        Toast.show('Apple 登录暂不可用')
      }
    }
  }

  const handleQQ = () => {
    if (onQQClick) onQQClick()
    else {
      try {
        qqLogin(`${window.location.origin}/auth/social/callback/qq`)
      } catch {
        Toast.show('QQ 登录暂不可用')
      }
    }
  }

  return (
    <div className={`flex ${layout === 'horizontal' ? 'justify-center gap-4' : 'flex-col gap-3'}`}>
      <button
        onClick={handleWechat}
        className={`${getSocialButtonClass(socialConfig.socialStyle, 'wechat')} ${socialConfig.socialStyle === 'colorful' ? 'bg-[#07C160]' : ''}`}
        title="微信"
      >
        <WechatIcon />
      </button>

      <button
        onClick={handleQQ}
        className={`${getSocialButtonClass(socialConfig.socialStyle, 'qq')} ${socialConfig.socialStyle === 'colorful' ? 'bg-[#12B7F5]' : ''}`}
        title="QQ"
      >
        <QQIcon />
      </button>

      <button
        onClick={handleGoogle}
        className={`${getSocialButtonClass(socialConfig.socialStyle, 'google')} ${socialConfig.socialStyle === 'colorful' ? 'bg-white border border-gray-200' : ''}`}
        title="Google"
      >
        <GoogleIcon />
      </button>

      <button
        onClick={handleApple}
        className={`${getSocialButtonClass(socialConfig.socialStyle, 'apple')} ${socialConfig.socialStyle === 'colorful' ? 'bg-black' : ''}`}
        title="Apple"
      >
        <AppleIcon />
      </button>
    </div>
  )
}
```

- [ ] **Step 2: 添加 qqLogin 到 socialLogin.ts**

```typescript
// user-mall/src/utils/socialLogin.ts
// 在现有文件末尾添加

export const qqLogin = (callbackURL: string) => {
  const appId = import.meta.env.VITE_QQ_APP_ID
  if (!appId) {
    throw new Error('QQ_APP_ID not configured')
  }
  const redirectURI = encodeURIComponent(callbackURL)
  const scope = 'get_user_info'
  const state = crypto.randomUUID()

  const authURL = `https://graph.qq.com/oauth2.0/authorize?response_type=code&client_id=${appId}&redirect_uri=${redirectURI}&scope=${scope}&state=${state}`

  window.location.href = authURL
}
```

- [ ] **Step 3: 更新 auth.ts 添加 QQ 类型**

```typescript
// user-mall/src/api/auth.ts
// 添加 QQ 平台到 SocialProvider 类型

export type SocialProvider = 'wechat' | 'google' | 'apple' | 'qq'
```

- [ ] **Step 4: 提交**

```bash
git add user-mall/src/components/SocialLoginButtons/index.tsx user-mall/src/utils/socialLogin.ts user-mall/src/api/auth.ts
git commit -m "feat: add QQ login support to social login buttons"
```

---

## Phase 3: 登录页面重构

### Task 6: 重构 Login.tsx 为主容器

**Files:**
- Modify: `user-mall/src/pages/Auth/Login.tsx`

- [ ] **Step 1: 重构 Login.tsx 作为主题路由分发器**

```tsx
// user-mall/src/pages/Auth/Login.tsx
import { useEffect, useState } from 'react'
import { useThemeStore } from '@/stores/themeStore'
import { themes } from '@/styles/login-themes/theme-config'
import { ThemeSwitcher } from '@/components/ThemeSwitcher'

const LoginVibrant = () => import('./LoginVibrant')
const LoginStable = () => import('./LoginStable')
const LoginLuxury = () => import('./LoginLuxury')

export default function Login() {
  const theme = useThemeStore((s) => s.theme)
  const [DynamicComponent, setDynamicComponent] = useState<React.LazyExoticComponent<() => JSX.Element> | null>(null)

  useEffect(() => {
    let loader: () => Promise<{ default: () => JSX.Element }>
    switch (theme) {
      case 'vibrant':
        loader = LoginVibrant
        break
      case 'stable':
        loader = LoginStable
        break
      case 'luxury':
        loader = LoginLuxury
        break
      default:
        loader = LoginVibrant
    }
    loader().then((module) => setDynamicComponent(() => module.default))
  }, [theme])

  if (!DynamicComponent) {
    return (
      <div className="min-h-screen flex items-center justify-center" style={{ background: themes[theme].colors.backgroundGradient }}>
        <div className="animate-pulse text-white">加载中...</div>
      </div>
    )
  }

  return (
    <>
      <ThemeSwitcher />
      <DynamicComponent />
    </>
  )
}
```

- [ ] **Step 2: 提交**

```bash
git add user-mall/src/pages/Auth/Login.tsx
git commit -m "refactor: split Login.tsx into theme-based components"
```

---

### Task 7: 实现兴趣电商主题 (LoginVibrant)

**Files:**
- Create: `user-mall/src/pages/Auth/LoginVibrant.tsx`

- [ ] **Step 1: 创建 LoginVibrant 组件**

```tsx
// user-mall/src/pages/Auth/LoginVibrant.tsx
import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Form, Input, Toast } from 'antd-mobile'
import { login, smsLogin, sendSms } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'
import SocialLoginButtons from '@/components/SocialLoginButtons'

const isDesktop = () => window.innerWidth >= 768

export default function LoginVibrant() {
  const navigate = useNavigate()
  const { login: setAuth } = useAuthStore()
  const [loading, setLoading] = useState(false)
  const [loginType, setLoginType] = useState<'password' | 'sms'>('password')
  const [smsCountdown, setSmsCountdown] = useState(0)
  const [smsLoading, setSmsLoading] = useState(false)
  const [mounted, setMounted] = useState(false)
  const [form] = Form.useForm()
  const [tabStyle, setTabStyle] = useState({ left: '0%', width: '50%' })
  const [isDesktopView, setIsDesktopView] = useState(false)

  useEffect(() => {
    setMounted(true)
    const check = () => setIsDesktopView(isDesktop())
    check()
    window.addEventListener('resize', check)
    return () => window.removeEventListener('resize', check)
  }, [])

  useEffect(() => {
    if (smsCountdown > 0) {
      const t = setTimeout(() => setSmsCountdown(smsCountdown - 1), 1000)
      return () => clearTimeout(t)
    }
  }, [smsCountdown])

  const handleTab = (type: 'password' | 'sms') => {
    setLoginType(type)
    setTabStyle({ left: type === 'password' ? '0%' : '50%', width: '50%' })
  }

  const handleSendSms = async () => {
    const phone = form.getFieldValue('phone')
    if (!phone || !/^1[3-9]\d{9}$/.test(phone)) {
      Toast.show('请输入正确的手机号')
      return
    }
    try {
      setSmsLoading(true)
      await sendSms({ phone, type: 'login' })
      Toast.show('验证码已发送')
      setSmsCountdown(60)
    } catch {
      Toast.show('发送失败')
    } finally {
      setSmsLoading(false)
    }
  }

  const handleSubmit = async (values: any) => {
    try {
      setLoading(true)
      if (loginType === 'password') {
        const res = await login({ account: values.account || '', password: values.password || '' })
        setAuth(res.token, res.user)
        Toast.show('登录成功')
        navigate('/')
      } else {
        const res = await smsLogin({ phone: values.phone || '', code: values.code || '' })
        setAuth(res.token, res.user)
        Toast.show('登录成功')
        navigate('/')
      }
    } catch {
      Toast.show('登录失败')
    } finally {
      setLoading(false)
    }
  }

  // Desktop Layout
  if (isDesktopView) {
    return (
      <div className="min-h-screen flex" style={{ background: 'linear-gradient(135deg, #FF6B35 0%, #FF4D94 100%)' }}>
        {/* Left side - Brand */}
        <div className="flex-1 p-12 flex flex-col justify-center items-center">
          <div className="text-center text-white">
            <h1 className="text-5xl font-bold mb-4">🎉 兴趣电商</h1>
            <p className="text-xl text-white/80 mb-8">发现你的热爱，好物即刻享</p>
            <div className="flex gap-8 justify-center text-white/90">
              <div><span className="text-3xl font-bold">1000+</span><br/><span className="text-sm">精选好物</span></div>
              <div><span className="text-3xl font-bold">500+</span><br/><span className="text-sm">品牌商家</span></div>
              <div><span className="text-3xl font-bold">10万+</span><br/><span className="text-sm">满意用户</span></div>
            </div>
          </div>
          {/* Floating elements */}
          <div className="absolute top-20 left-20 w-20 h-20 rounded-full bg-yellow-400/30 animate-bounce" />
          <div className="absolute bottom-40 right-32 w-16 h-16 rounded-full bg-cyan-400/30 animate-pulse" />
          <div className="absolute top-40 right-40 w-12 h-12 rounded-full bg-pink-400/30 animate-ping" />
        </div>

        {/* Right side - Form */}
        <div className="w-[500px] bg-white rounded-l-3xl shadow-2xl p-10 flex flex-col justify-center" style={{ boxShadow: '-20px 0 60px rgba(0,0,0,0.2)' }}>
          <h2 className="text-3xl font-bold text-gray-800 mb-2">欢迎回来</h2>
          <p className="text-gray-500 mb-6">登录即享更多优惠</p>

          {/* Tab */}
          <div className="relative flex mb-6 bg-gray-100 rounded-xl p-1">
            <div className="absolute top-0 bottom-0 rounded-xl bg-white shadow transition-all duration-300" style={{ left: tabStyle.left, width: tabStyle.width }} />
            <div className={`relative flex-1 py-2.5 text-center font-medium text-sm rounded-xl cursor-pointer z-10 ${loginType === 'password' ? 'text-orange-500' : 'text-gray-500'}`} onClick={() => handleTab('password')}>密码登录</div>
            <div className={`relative flex-1 py-2.5 text-center font-medium text-sm rounded-xl cursor-pointer z-10 ${loginType === 'sms' ? 'text-orange-500' : 'text-gray-500'}`} onClick={() => handleTab('sms')}>验证码登录</div>
          </div>

          {/* Form */}
          <Form layout="vertical" onFinish={handleSubmit}>
            {loginType === 'password' ? (
              <>
                <div className="mb-4">
                  <Input placeholder="手机号/用户名" className="h-12 rounded-xl bg-gray-50 text-base" />
                </div>
                <div className="mb-4">
                  <Input type="password" placeholder="请输入密码" className="h-12 rounded-xl bg-gray-50 text-base" />
                </div>
              </>
            ) : (
              <>
                <div className="mb-4">
                  <Input placeholder="请输入手机号" className="h-12 rounded-xl bg-gray-50 text-base" />
                </div>
                <div className="mb-4 flex gap-3">
                  <Input placeholder="验证码" className="h-12 rounded-xl bg-gray-50 text-base flex-1" />
                  <div className={`h-12 px-4 rounded-xl flex items-center text-sm font-medium cursor-pointer ${smsCountdown > 0 ? 'bg-gray-100 text-gray-400' : 'bg-orange-100 text-orange-500'}`} onClick={smsCountdown > 0 ? undefined : handleSendSms}>
                    {smsCountdown > 0 ? `${smsCountdown}s` : '获取验证码'}
                  </div>
                </div>
              </>
            )}
            <Button block color="primary" size="large" loading={loading} type="submit" className="h-12 rounded-xl text-base font-bold" style={{ background: 'linear-gradient(135deg, #FF6B35, #FF4D94)' }}>
              {loading ? '登录中...' : '登 录'}
            </Button>
          </Form>

          <div className="mt-6">
            <div className="relative"><div className="absolute inset-0 flex items-center"><div className="w-full border-t border-gray-200" /></div><div className="relative flex justify-center text-sm"><span className="px-4 bg-white text-gray-400">其他登录方式</span></div></div>
            <div className="mt-4 flex justify-center"><SocialLoginButtons theme="vibrant" layout="horizontal" /></div>
          </div>

          <div className="mt-6 text-center">
            <button onClick={() => navigate('/')} className="text-gray-400 text-sm hover:text-orange-500 transition-colors">跳过登录，先看看 →</button>
          </div>
        </div>
      </div>
    )
  }

  // Mobile Layout
  return (
    <div className="min-h-screen flex flex-col" style={{ background: 'linear-gradient(135deg, #FF6B35 0%, #FF4D94 100%)' }}>
      <div className="p-4 flex justify-end">
        <button onClick={() => navigate('/')} className="text-white/80 text-sm px-3 py-1.5 rounded-full bg-white/10 backdrop-blur-sm">跳过</button>
      </div>
      <div className="text-center text-white py-8">
        <h1 className="text-3xl font-bold mb-2">🎉 兴趣电商</h1>
        <p className="text-white/70">发现你的热爱</p>
      </div>
      <div className="flex-1 bg-white rounded-t-3xl p-6">
        <div className="relative flex mb-6 bg-gray-100 rounded-xl p-1">
          <div className="absolute top-0 bottom-0 rounded-xl bg-white shadow transition-all duration-300" style={{ left: tabStyle.left, width: tabStyle.width }} />
          <div className={`relative flex-1 py-2.5 text-center font-medium text-sm rounded-xl cursor-pointer z-10 ${loginType === 'password' ? 'text-orange-500' : 'text-gray-500'}`} onClick={() => handleTab('password')}>密码</div>
          <div className={`relative flex-1 py-2.5 text-center font-medium text-sm rounded-xl cursor-pointer z-10 ${loginType === 'sms' ? 'text-orange-500' : 'text-gray-500'}`} onClick={() => handleTab('sms')}>验证码</div>
        </div>
        <Form layout="vertical" onFinish={handleSubmit}>
          <div className="mb-4"><Input placeholder="手机号/用户名" className="h-12 rounded-xl bg-gray-50" /></div>
          {loginType === 'sms' && <div className="mb-4"><Input placeholder="验证码" className="h-12 rounded-xl bg-gray-50" /></div>}
          {loginType === 'password' && <div className="mb-4"><Input type="password" placeholder="密码" className="h-12 rounded-xl bg-gray-50" /></div>}
          <Button block color="primary" size="large" loading={loading} type="submit" className="h-12 rounded-xl font-bold" style={{ background: 'linear-gradient(135deg, #FF6B35, #FF4D94)' }}>
            {loading ? '登录中...' : '登 录'}
          </Button>
        </Form>
        <div className="mt-6"><div className="relative"><div className="absolute inset-0 flex items-center"><div className="w-full border-t border-gray-200" /></div><div className="relative flex justify-center text-sm"><span className="px-4 bg-white text-gray-400">其他方式</span></div></div></div>
        <div className="mt-4 flex justify-center"><SocialLoginButtons theme="vibrant" /></div>
        <div className="mt-6 text-center text-sm text-gray-500">还没有账号？<span className="text-orange-500 font-medium">立即注册</span></div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: 提交**

```bash
git add user-mall/src/pages/Auth/LoginVibrant.tsx
git commit -m "feat: implement LoginVibrant theme (兴趣电商)"
```

---

### Task 8: 实现大厂稳重主题 (LoginStable)

**Files:**
- Create: `user-mall/src/pages/Auth/LoginStable.tsx`

- [ ] **Step 1: 创建 LoginStable 组件**

```tsx
// user-mall/src/pages/Auth/LoginStable.tsx
import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Form, Input, Toast } from 'antd-mobile'
import { login, smsLogin, sendSms } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'
import SocialLoginButtons from '@/components/SocialLoginButtons'

const isDesktop = () => window.innerWidth >= 768

export default function LoginStable() {
  const navigate = useNavigate()
  const { login: setAuth } = useAuthStore()
  const [loading, setLoading] = useState(false)
  const [loginType, setLoginType] = useState<'password' | 'sms'>('password')
  const [smsCountdown, setSmsCountdown] = useState(0)
  const [smsLoading, setSmsLoading] = useState(false)
  const [form] = Form.useForm()
  const [isDesktopView, setIsDesktopView] = useState(false)

  useEffect(() => {
    const check = () => setIsDesktopView(isDesktop())
    check()
    window.addEventListener('resize', check)
    return () => window.removeEventListener('resize', check)
  }, [])

  useEffect(() => {
    if (smsCountdown > 0) {
      const t = setTimeout(() => setSmsCountdown(smsCountdown - 1), 1000)
      return () => clearTimeout(t)
    }
  }, [smsCountdown])

  const handleSendSms = async () => {
    const phone = form.getFieldValue('phone')
    if (!phone || !/^1[3-9]\d{9}$/.test(phone)) {
      Toast.show('请输入正确的手机号')
      return
    }
    try {
      setSmsLoading(true)
      await sendSms({ phone, type: 'login' })
      Toast.show('验证码已发送')
      setSmsCountdown(60)
    } catch {
      Toast.show('发送失败')
    } finally {
      setSmsLoading(false)
    }
  }

  const handleSubmit = async (values: any) => {
    try {
      setLoading(true)
      if (loginType === 'password') {
        const res = await login({ account: values.account || '', password: values.password || '' })
        setAuth(res.token, res.user)
        Toast.show('登录成功')
        navigate('/')
      } else {
        const res = await smsLogin({ phone: values.phone || '', code: values.code || '' })
        setAuth(res.token, res.user)
        Toast.show('登录成功')
        navigate('/')
      }
    } catch {
      Toast.show('登录失败，请检查账号信息')
    } finally {
      setLoading(false)
    }
  }

  const categories = ['数码家电', '服饰鞋包', '食品生鲜', '美妆护肤', '家居用品', '图书文具']
  const trustBadges = [
    { icon: '🔒', text: '安全支付' },
    { icon: '✓', text: '正品保障' },
    { icon: '🚚', text: '快速配送' },
  ]

  // Desktop Layout
  if (isDesktopView) {
    return (
      <div className="min-h-screen bg-gray-50">
        <div className="bg-white border-b border-gray-200 px-8 py-4 flex justify-between items-center">
          <div className="text-xl font-bold text-teal-700">🏪 商城</div>
          <button onClick={() => navigate('/')} className="text-gray-500 text-sm hover:text-teal-600">跳过登录 →</button>
        </div>

        <div className="flex max-w-6xl mx-auto mt-12 gap-8">
          {/* Left - Form */}
          <div className="flex-1 max-w-md">
            <div className="bg-white rounded-lg p-8 shadow-sm border border-gray-200">
              <h2 className="text-2xl font-bold text-gray-900 mb-1">登录</h2>
              <p className="text-gray-500 text-sm mb-6">欢迎回来</p>

              <div className="flex gap-4 mb-6">
                <button
                  className={`px-4 py-2 text-sm font-medium rounded ${loginType === 'password' ? 'text-teal-600 border-b-2 border-teal-600' : 'text-gray-500'}`}
                  onClick={() => setLoginType('password')}
                >密码登录</button>
                <button
                  className={`px-4 py-2 text-sm font-medium rounded ${loginType === 'sms' ? 'text-teal-600 border-b-2 border-teal-600' : 'text-gray-500'}`}
                  onClick={() => setLoginType('sms')}
                >验证码登录</button>
              </div>

              <Form layout="vertical" onFinish={handleSubmit}>
                <div className="mb-4">
                  <label className="text-sm text-gray-600">邮箱/手机</label>
                  <Input placeholder="请输入" className="mt-1 h-10 rounded" />
                </div>
                {loginType === 'password' && (
                  <div className="mb-4">
                    <label className="text-sm text-gray-600">密码</label>
                    <Input type="password" placeholder="请输入密码" className="mt-1 h-10 rounded" />
                  </div>
                )}
                {loginType === 'sms' && (
                  <div className="mb-4">
                    <label className="text-sm text-gray-600">验证码</label>
                    <div className="flex gap-2 mt-1">
                      <Input placeholder="请输入" className="h-10 rounded flex-1" />
                      <button
                        className={`px-3 h-10 rounded text-sm ${smsCountdown > 0 ? 'bg-gray-100 text-gray-400' : 'bg-teal-50 text-teal-600'}`}
                        onClick={smsCountdown > 0 ? undefined : handleSendSms}
                      >
                        {smsCountdown > 0 ? `${smsCountdown}s` : '获取验证码'}
                      </button>
                    </div>
                  </div>
                )}
                <Button block color="primary" size="large" loading={loading} type="submit" className="h-10 rounded text-base font-medium mt-4" style={{ background: '#007185' }}>
                  {loading ? '登录中...' : '登 录'}
                </Button>
              </Form>

              <div className="mt-6">
                <div className="flex items-center gap-2 mb-4">
                  <div className="flex-1 h-px bg-gray-200" />
                  <span className="text-xs text-gray-400">其他登录方式</span>
                  <div className="flex-1 h-px bg-gray-200" />
                </div>
                <SocialLoginButtons theme="stable" layout="horizontal" />
              </div>

              <div className="mt-6 flex justify-between text-sm">
                <button className="text-teal-600">忘记密码</button>
                <button onClick={() => navigate('/register')} className="text-teal-600">立即注册 →</button>
              </div>
            </div>
          </div>

          {/* Right - Features */}
          <div className="flex-1">
            <h3 className="text-lg font-bold text-gray-800 mb-4">快捷分类</h3>
            <div className="grid grid-cols-3 gap-3 mb-8">
              {categories.map((cat) => (
                <div key={cat} className="bg-white p-4 rounded-lg text-center text-sm text-gray-700 border border-gray-200 hover:border-teal-300 cursor-pointer transition-colors">
                  {cat}
                </div>
              ))}
            </div>

            <h3 className="text-lg font-bold text-gray-800 mb-4">信任保障</h3>
            <div className="flex gap-4">
              {trustBadges.map((badge) => (
                <div key={badge.text} className="flex items-center gap-2 text-sm text-gray-600">
                  <span>{badge.icon}</span>
                  <span>{badge.text}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="text-center text-xs text-gray-400 mt-12">© 2026 商城 · 品质保障</div>
      </div>
    )
  }

  // Mobile Layout
  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-white border-b border-gray-200 px-4 py-3 flex justify-between items-center">
        <div className="text-lg font-bold text-teal-700">🏪 商城</div>
        <button onClick={() => navigate('/')} className="text-gray-500 text-sm">跳过</button>
      </div>

      <div className="p-6">
        <div className="bg-white rounded-lg p-6 shadow-sm border border-gray-200">
          <h2 className="text-xl font-bold text-gray-900 mb-1">登录</h2>
          <p className="text-gray-500 text-sm mb-4">欢迎回来</p>

          <div className="flex gap-2 mb-4 text-sm">
            <button
              className={`px-3 py-1.5 rounded ${loginType === 'password' ? 'bg-teal-50 text-teal-700 font-medium' : 'text-gray-500'}`}
              onClick={() => setLoginType('password')}
            >密码</button>
            <button
              className={`px-3 py-1.5 rounded ${loginType === 'sms' ? 'bg-teal-50 text-teal-700 font-medium' : 'text-gray-500'}`}
              onClick={() => setLoginType('sms')}
            >验证码</button>
          </div>

          <Form layout="vertical" onFinish={handleSubmit}>
            <div className="mb-4"><Input placeholder="手机号/用户名" className="h-11 rounded" /></div>
            {loginType === 'password' && <div className="mb-4"><Input type="password" placeholder="密码" className="h-11 rounded" /></div>}
            {loginType === 'sms' && <div className="mb-4 flex gap-2"><Input placeholder="验证码" className="h-11 rounded flex-1" /><button className="px-3 h-11 bg-teal-50 text-teal-600 rounded text-sm">{smsCountdown > 0 ? `${smsCountdown}s` : '获取'}</button></div>}
            <Button block color="primary" size="large" loading={loading} type="submit" className="h-11 rounded" style={{ background: '#007185' }}>{loading ? '登录中...' : '登 录'}</Button>
          </Form>

          <div className="mt-4 flex justify-center"><SocialLoginButtons theme="stable" /></div>

          <div className="mt-4 flex justify-between text-sm">
            <button className="text-teal-600">忘记密码</button>
            <button onClick={() => navigate('/register')} className="text-teal-600">注册 →</button>
          </div>
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: 提交**

```bash
git add user-mall/src/pages/Auth/LoginStable.tsx
git commit -m "feat: implement LoginStable theme (大厂稳重)"
```

---

### Task 9: 实现轻奢简约主题 (LoginLuxury)

**Files:**
- Create: `user-mall/src/pages/Auth/LoginLuxury.tsx`

- [ ] **Step 1: 创建 LoginLuxury 组件**

```tsx
// user-mall/src/pages/Auth/LoginLuxury.tsx
import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Form, Input, Toast } from 'antd-mobile'
import { login, smsLogin, sendSms } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'
import SocialLoginButtons from '@/components/SocialLoginButtons'

const isDesktop = () => window.innerWidth >= 768

export default function LoginLuxury() {
  const navigate = useNavigate()
  const { login: setAuth } = useAuthStore()
  const [loading, setLoading] = useState(false)
  const [loginType, setLoginType] = useState<'password' | 'sms'>('password')
  const [smsCountdown, setSmsCountdown] = useState(0)
  const [smsLoading, setSmsLoading] = useState(false)
  const [mounted, setMounted] = useState(false)
  const [form] = Form.useForm()
  const [isDesktopView, setIsDesktopView] = useState(false)

  useEffect(() => {
    setMounted(true)
    const check = () => setIsDesktopView(isDesktop())
    check()
    window.addEventListener('resize', check)
    return () => window.removeEventListener('resize', check)
  }, [])

  useEffect(() => {
    if (smsCountdown > 0) {
      const t = setTimeout(() => setSmsCountdown(smsCountdown - 1), 1000)
      return () => clearTimeout(t)
    }
  }, [smsCountdown])

  const handleSendSms = async () => {
    const phone = form.getFieldValue('phone')
    if (!phone || !/^1[3-9]\d{9}$/.test(phone)) {
      Toast.show('请输入正确的手机号')
      return
    }
    try {
      setSmsLoading(true)
      await sendSms({ phone, type: 'login' })
      Toast.show('验证码已发送')
      setSmsCountdown(60)
    } catch {
      Toast.show('发送失败')
    } finally {
      setSmsLoading(false)
    }
  }

  const handleSubmit = async (values: any) => {
    try {
      setLoading(true)
      if (loginType === 'password') {
        const res = await login({ account: values.account || '', password: values.password || '' })
        setAuth(res.token, res.user)
        Toast.show('登录成功')
        navigate('/')
      } else {
        const res = await smsLogin({ phone: values.phone || '', code: values.code || '' })
        setAuth(res.token, res.user)
        Toast.show('登录成功')
        navigate('/')
      }
    } catch {
      Toast.show('登录失败，请检查账号信息')
    } finally {
      setLoading(false)
    }
  }

  // Desktop Layout
  if (isDesktopView) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-stone-100">
        <div
          className="w-full max-w-4xl flex shadow-2xl rounded-sm overflow-hidden"
          style={{
            opacity: mounted ? 1 : 0,
            transform: mounted ? 'translateY(0)' : 'translateY(20px)',
            transition: 'opacity 0.8s ease, transform 0.8s ease',
          }}
        >
          {/* Left - Brand */}
          <div className="flex-1 bg-stone-900 p-16 flex flex-col justify-center items-center text-white">
            <div className="w-16 h-px bg-stone-600 mb-8" />
            <h1 className="text-4xl font-light tracking-wider mb-4">LUXURY</h1>
            <p className="text-stone-400 tracking-widest text-sm mb-8">精 · 致 · 生 · 活</p>
            <div className="w-16 h-px bg-stone-600 mt-8" />
          </div>

          {/* Right - Form */}
          <div className="w-[400px] bg-white p-16 flex flex-col justify-center">
            <h2 className="text-2xl font-light text-stone-800 mb-2">欢迎回来</h2>
            <p className="text-stone-400 text-sm mb-8">探索专属精致体验</p>

            <div className="flex gap-6 mb-8">
              <button
                className={`text-sm pb-2 ${loginType === 'password' ? 'text-stone-800 border-b border-stone-800' : 'text-stone-400'}`}
                onClick={() => setLoginType('password')}
              >密码登录</button>
              <button
                className={`text-sm pb-2 ${loginType === 'sms' ? 'text-stone-800 border-b border-stone-800' : 'text-stone-400'}`}
                onClick={() => setLoginType('sms')}
              >验证码</button>
            </div>

            <Form layout="vertical" onFinish={handleSubmit}>
              <div className="space-y-4">
                <div className="border-b border-stone-200 pb-2">
                  <input
                    type="text"
                    placeholder="邮箱 / 手机号"
                    className="w-full text-stone-800 placeholder-stone-300 outline-none bg-transparent text-sm"
                  />
                </div>
                {loginType === 'password' && (
                  <div className="border-b border-stone-200 pb-2">
                    <input
                      type="password"
                      placeholder="密码"
                      className="w-full text-stone-800 placeholder-stone-300 outline-none bg-transparent text-sm"
                    />
                  </div>
                )}
                {loginType === 'sms' && (
                  <div className="flex border-b border-stone-200 pb-2">
                    <input
                      type="text"
                      placeholder="验证码"
                      className="flex-1 text-stone-800 placeholder-stone-300 outline-none bg-transparent text-sm"
                    />
                    <button
                      className={`text-xs ${smsCountdown > 0 ? 'text-stone-300' : 'text-stone-500'}`}
                      onClick={smsCountdown > 0 ? undefined : handleSendSms}
                    >
                      {smsCountdown > 0 ? `${smsCountdown}s` : '获取验证码'}
                    </button>
                  </div>
                )}
              </div>

              <Button
                block
                type="submit"
                loading={loading}
                className="mt-8 h-12 rounded-sm bg-stone-800 text-white text-sm tracking-widest"
                style={{ background: '#1A1A1A' }}
              >
                {loading ? '登录中...' : '登 录'}
              </Button>
            </Form>

            <div className="mt-8 flex justify-center">
              <SocialLoginButtons theme="luxury" layout="horizontal" />
            </div>

            <div className="mt-8 text-center">
              <button onClick={() => navigate('/register')} className="text-stone-400 text-sm hover:text-stone-600 transition-colors">
                还没有账号？<span className="border-b border-stone-300">立即注册</span>
              </button>
            </div>

            <div className="mt-8 text-center">
              <button onClick={() => navigate('/')} className="text-stone-300 text-xs hover:text-stone-500 transition-colors">
                跳过登录 →
              </button>
            </div>
          </div>
        </div>
      </div>
    )
  }

  // Mobile Layout
  return (
    <div className="min-h-screen bg-stone-100 flex flex-col items-center justify-center p-6">
      <div
        className="w-full max-w-sm"
        style={{
          opacity: mounted ? 1 : 0,
          transform: mounted ? 'translateY(0)' : 'translateY(20px)',
          transition: 'opacity 0.6s ease, transform 0.6s ease',
        }}
      >
        <div className="text-center mb-12">
          <div className="w-12 h-px bg-stone-300 mx-auto mb-4" />
          <h1 className="text-2xl font-light tracking-widest text-stone-800">LUXURY</h1>
          <p className="text-stone-400 text-xs tracking-widest mt-1">精 · 致 · 生 · 活</p>
        </div>

        <div className="bg-white rounded-sm shadow-lg p-8">
          <div className="flex gap-4 mb-6">
            <button
              className={`text-sm ${loginType === 'password' ? 'text-stone-800 border-b border-stone-800' : 'text-stone-400'}`}
              onClick={() => setLoginType('password')}
            >密码</button>
            <button
              className={`text-sm ${loginType === 'sms' ? 'text-stone-800 border-b border-stone-800' : 'text-stone-400'}`}
              onClick={() => setLoginType('sms')}
            >验证码</button>
          </div>

          <Form layout="vertical" onFinish={handleSubmit}>
            <div className="space-y-3 mb-4">
              <input
                type="text"
                placeholder="邮箱 / 手机号"
                className="w-full text-stone-800 placeholder-stone-300 outline-none border-b border-stone-200 pb-2 text-sm"
              />
              {loginType === 'password' && (
                <input
                  type="password"
                  placeholder="密码"
                  className="w-full text-stone-800 placeholder-stone-300 outline-none border-b border-stone-200 pb-2 text-sm"
                />
              )}
              {loginType === 'sms' && (
                <div className="flex items-center border-b border-stone-200 pb-2">
                  <input
                    type="text"
                    placeholder="验证码"
                    className="flex-1 text-stone-800 placeholder-stone-300 outline-none text-sm"
                  />
                  <button className="text-xs text-stone-400">{smsCountdown > 0 ? `${smsCountdown}s` : '获取'}</button>
                </div>
              )}
            </div>

            <Button
              block
              type="submit"
              loading={loading}
              className="h-11 rounded-sm bg-stone-800 text-white text-sm tracking-wider"
              style={{ background: '#1A1A1A' }}
            >
              {loading ? '登录中...' : '登 录'}
            </Button>
          </Form>

          <div className="mt-6 flex justify-center">
            <SocialLoginButtons theme="luxury" />
          </div>

          <div className="mt-6 text-center">
            <button onClick={() => navigate('/register')} className="text-stone-400 text-xs">
              还没有账号？<span className="border-b border-stone-300">立即注册</span>
            </button>
          </div>
        </div>

        <div className="mt-8 text-center">
          <button onClick={() => navigate('/')} className="text-stone-400 text-xs hover:text-stone-600">
            跳过登录 →
          </button>
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: 提交**

```bash
git add user-mall/src/pages/Auth/LoginLuxury.tsx
git commit -m "feat: implement LoginLuxury theme (轻奢简约)"
```

---

## Phase 4: 商品展示 (可选，后续实现)

### Task 10: 添加商品展示 API 和组件 (后续)

此任务可后续实现，暂不包含在当前计划中。

---

## 验证清单

- [ ] 主题切换正常工作
- [ ] 主题选择持久化到 localStorage
- [ ] Vibrant 主题在桌面端/移动端渲染正确
- [ ] Stable 主题在桌面端/移动端渲染正确
- [ ] Luxury 主题在桌面端/移动端渲染正确
- [ ] QQ 登录按钮显示且可点击
- [ ] 微信/Google/Apple 登录保持正常
- [ ] 跳过登录按钮正常工作
