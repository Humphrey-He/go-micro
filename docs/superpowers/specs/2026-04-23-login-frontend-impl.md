# 登录页面前端实现文档

> 更新时间: 2026-04-23
>
> 基于: `docs/superpowers/specs/2026-04-23-login-theme-design.md`

---

## 一、组件架构

### 1.1 组件层级

```
LoginPage (主容器)
├── ThemeProvider (主题上下文)
├── ThemeSwitcher (主题切换)
├── LoginHeader (头部区域)
├── BrandSection (品牌展示区) [Vibrant/Luxury]
├── FunctionSection (功能区) [Stable]
├── LoginForm (登录表单)
│   ├── TabSwitcher (密码/验证码切换)
│   ├── AccountInput / PasswordInput
│   ├── SMSCodeInput (验证码输入)
│   ├── SubmitButton (提交按钮)
│   └── SocialLogin (社交登录)
├── ProductShowcase (商品展示区)
├── FooterLinks (底部链接)
└── SkipButton (跳过按钮)
```

### 1.2 主题配置

```typescript
// src/styles/login-themes/theme-config.ts

export type LoginTheme = 'vibrant' | 'stable' | 'luxury'

export interface ThemeConfig {
  id: LoginTheme
  name: string
  nameCn: string
  colors: {
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
  cardStyle: {
    borderRadius: number
    shadow: string
    hoverShadow: string
  }
  socialStyle: 'colorful' | 'subtle' | 'outline'
  animation: {
    mountDuration: number
    hoverLift: boolean
  }
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
    cardStyle: {
      borderRadius: 24,
      shadow: '0 20px 60px rgba(255, 107, 53, 0.3)',
      hoverShadow: '0 30px 80px rgba(255, 107, 53, 0.4)',
    },
    socialStyle: 'colorful',
    animation: {
      mountDuration: 600,
      hoverLift: true,
    },
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
    cardStyle: {
      borderRadius: 8,
      shadow: '0 2px 8px rgba(0, 0, 0, 0.08)',
      hoverShadow: '0 4px 16px rgba(0, 0, 0, 0.12)',
    },
    socialStyle: 'subtle',
    animation: {
      mountDuration: 400,
      hoverLift: false,
    },
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
    cardStyle: {
      borderRadius: 4,
      shadow: '0 1px 3px rgba(0, 0, 0, 0.05)',
      hoverShadow: '0 8px 30px rgba(0, 0, 0, 0.08)',
    },
    socialStyle: 'outline',
    animation: {
      mountDuration: 800,
      hoverLift: true,
    },
  },
}
```

---

## 二、文件清单

### 2.1 新增文件

| 文件路径 | 说明 |
|----------|------|
| `src/styles/login-themes/theme-config.ts` | 主题配置常量 |
| `src/stores/themeStore.ts` | 主题状态管理 (zustand) |
| `src/components/ThemeSwitcher/index.tsx` | 主题切换组件 |
| `src/components/ThemeSwitcher/ThemeDot.tsx` | 主题圆点组件 |
| `src/hooks/useLoginTheme.ts` | 主题相关 hooks |
| `src/api/recommendation.ts` (修改) | 添加 login-showcase 接口 |

### 2.2 修改文件

| 文件路径 | 修改内容 |
|----------|----------|
| `src/pages/Auth/Login.tsx` | 重构为三主题支持 |
| `src/components/SocialLoginButtons/index.tsx` | 添加 QQ 登录按钮 |
| `src/utils/socialLogin.ts` | 添加 qqLogin 函数 |
| `src/api/auth.ts` | 添加 QQ 回调类型 |
| `vite.config.ts` (如需要) | 环境变量配置 |

---

## 三、Login.tsx 重构

### 3.1 组件 Props

```typescript
interface LoginPageProps {
  // 从路由获取当前主题（可选）
  initialTheme?: LoginTheme
}
```

### 3.2 主题上下文

```tsx
// src/context/ThemeContext.tsx
import { createContext, useContext } from 'react'
import type { LoginTheme } from '@/styles/login-themes/theme-config'

interface ThemeContextType {
  theme: LoginTheme
  setTheme: (theme: LoginTheme) => void
  isDark: boolean
}

export const ThemeContext = createContext<ThemeContextType>({
  theme: 'vibrant',
  setTheme: () => {},
  isDark: false,
})

export const useTheme = () => useContext(ThemeContext)
```

### 3.3 三主题布局渲染逻辑

```tsx
// src/pages/Auth/Login.tsx

export default function Login() {
  const { theme } = useTheme()
  const isDesktop = window.innerWidth >= 768

  // 根据主题渲染不同布局
  const renderDesktopLayout = () => {
    switch (theme) {
      case 'vibrant':
        return <VibrantDesktopLayout {...props} />
      case 'stable':
        return <StableDesktopLayout {...props} />
      case 'luxury':
        return <LuxuryDesktopLayout {...props} />
    }
  }

  const renderMobileLayout = () => {
    switch (theme) {
      case 'vibrant':
        return <VibrantMobileLayout {...props} />
      case 'stable':
        return <StableMobileLayout {...props} />
      case 'luxury':
        return <LuxuryMobileLayout {...props} />
    }
  }

  return isDesktop ? renderDesktopLayout() : renderMobileLayout()
}
```

---

## 四、ThemeSwitcher 组件

### 4.1 组件代码

```tsx
// src/components/ThemeSwitcher/index.tsx
import { useTheme } from '@/context/ThemeContext'
import { themes } from '@/styles/login-themes/theme-config'

export const ThemeSwitcher: React.FC = () => {
  const { theme, setTheme } = useTheme()

  return (
    <div className="theme-switcher flex items-center gap-2">
      {Object.values(themes).map((t) => (
        <button
          key={t.id}
          onClick={() => setTheme(t.id)}
          className={`
            theme-dot w-8 h-8 rounded-full border-2 transition-all
            ${theme === t.id ? 'border-primary scale-110' : 'border-transparent opacity-60'}
          `}
          style={{ background: t.colors.primary }}
          title={t.nameCn}
        />
      ))}
    </div>
  )
}
```

### 4.2 样式

```css
/* Tailwind classes + inline styles */
.theme-switcher {
  position: fixed;
  top: 16px;
  right: 16px;
  z-index: 1000;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(8px);
  border-radius: 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.theme-dot:hover {
  opacity: 1 !important;
}
```

---

## 五、SocialLoginButtons 增强

### 5.1 添加 QQ 图标

```tsx
// src/components/SocialLoginButtons/index.tsx

const QQIcon = () => (
  <svg viewBox="0 0 24 24" className="w-6 h-6">
    <path fill="#12B7F5" d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z"/>
    <path fill="#12B7F5" d="M12 6c-3.31 0-6 2.69-6 6s2.69 6 6 6 6-2.69 6-6-2.69-6-6-6zm0 10c-2.21 0-4-1.79-4-4s1.79-4 4-4 4 1.79 4 4-1.79 4-4 4z"/>
    <circle fill="#12B7F5" cx="8" cy="11" r="1.5"/>
    <circle fill="#12B7F5" cx="16" cy="11" r="1.5"/>
    <path fill="#12B7F5" d="M12 17c2 0 3.5-1 3.5-2H8.5c0 1 1.5 2 3.5 2z"/>
  </svg>
)
```

### 5.2 QQ 登录处理

```tsx
// 新增 qqLogin 处理
const handleQQ = () => {
  try {
    qqLogin(`${window.location.origin}/auth/social/callback/qq`)
  } catch {
    Toast.show('QQ 登录暂不可用')
  }
}

// 主题化按钮样式
const getButtonClass = (theme: string) => {
  switch (theme) {
    case 'vibrant':
      return 'w-12 h-12 rounded-full bg-[#12B7F5] text-white shadow-md hover:shadow-lg hover:scale-105'
    case 'stable':
      return 'w-10 h-10 rounded bg-gray-100 text-gray-600 hover:bg-gray-200'
    case 'luxury':
      return 'w-10 h-10 rounded-full border-2 border-gray-300 text-gray-600 hover:border-[#C9A96E]'
    default:
      return ''
  }
}
```

---

## 六、社交登录工具函数

### 6.1 qqLogin 实现

```typescript
// src/utils/socialLogin.ts

interface QQOAuthConfig {
  appId: string
  redirectURI: string
  scope: string
  state: string
}

export const qqLogin = (callbackURL: string) => {
  const appId = import.meta.env.VITE_QQ_APP_ID
  const redirectURI = encodeURIComponent(callbackURL)
  const scope = 'get_user_info'
  const state = crypto.randomUUID()

  const authURL = `https://graph.qq.com/oauth2.0/authorize?response_type=code&client_id=${appId}&redirect_uri=${redirectURI}&scope=${scope}&state=${state}`

  window.location.href = authURL
}

export const wechatLogin = (callbackURL: string) => {
  // 现有实现保持不变
}

export const googleLogin = () => {
  // 现有实现保持不变
}

export const appleLogin = () => {
  // 现有实现保持不变
}
```

### 6.2 回调页面

```tsx
// src/pages/Auth/SocialCallback.tsx
// 添加 QQ 回调处理

const handleQQCallback = async (code: string) => {
  try {
    const res = await post<SocialLoginResponse>('/auth/social/callback/qq', { code })
    // 处理登录
  } catch (error) {
    Toast.show('QQ 登录失败')
  }
}
```

---

## 七、商品展示接口

### 7.1 新增 API

```typescript
// src/api/recommendation.ts

export interface ShowcaseProduct {
  sku_id: string
  title: string
  price: number
  original_price: number
  image: string
  discount?: number
}

// 获取登录页展示商品
export const getLoginShowcase = (theme: string, limit = 6) =>
  get<{ items: ShowcaseProduct[] }>('/rec/login-showcase', { theme, limit })
```

### 7.2 商品展示组件

```tsx
// src/components/ProductShowcase/index.tsx

interface ProductShowcaseProps {
  theme: LoginTheme
  limit?: number
}

export const ProductShowcase: React.FC<ProductShowcaseProps> = ({ theme, limit = 6 }) => {
  const { data } = useSWR(['/rec/login-showcase', theme, limit], () =>
    getLoginShowcase(theme, limit)
  )

  return (
    <div className="product-showcase">
      {data?.items.map((product) => (
        <ProductCard key={product.sku_id} product={product} theme={theme} />
      ))}
    </div>
  )
}
```

---

## 八、主题 Store

### 8.1 Zustand Store

```typescript
// src/stores/themeStore.ts
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
      theme: 'vibrant', // 默认主题

      setTheme: (theme) => set({ theme }),

      toggleTheme: () => {
        const themes: LoginTheme[] = ['vibrant', 'stable', 'luxury']
        const current = get().theme
        const index = themes.indexOf(current)
        const next = themes[(index + 1) % themes.length]
        set({ theme: next })
      },
    }),
    {
      name: 'login-theme-storage',
    }
  )
)
```

---

## 九、环境变量

### 9.1 前端 .env

```bash
# 社交登录配置
VITE_WECHAT_APP_ID=your_wechat_appid
VITE_GOOGLE_CLIENT_ID=your_google_clientid.apps.googleusercontent.com
VITE_APPLE_CLIENT_ID=com.yourcompany.yourapp
VITE_QQ_APP_ID=your_qq_appid
```

---

## 十、实现步骤

### Phase 1: 基础架构
1. 创建主题配置文件
2. 创建 themeStore
3. 创建 ThemeContext
4. 实现 ThemeSwitcher 组件

### Phase 2: 登录页面重构
1. 重构 Login.tsx 使用主题上下文
2. 实现 Vibrant 主题布局
3. 实现 Stable 主题布局
4. 实现 Luxury 主题布局

### Phase 3: 社交登录增强
1. 添加 QQ 图标和登录函数
2. 更新 SocialLoginButtons 支持主题化样式
3. 添加 QQ 回调处理

### Phase 4: 商品展示
1. 添加 login-showcase API
2. 实现 ProductShowcase 组件
3. 集成到各主题布局

---

## 十一、测试检查清单

- [ ] 主题切换正常工作
- [ ] 主题选择持久化到 localStorage
- [ ] 三套主题在桌面端渲染正确
- [ ] 三套主题在移动端渲染正确
- [ ] QQ 登录按钮显示且可点击
- [ ] 微信/Google/Apple 登录保持正常
- [ ] 商品展示区域正常显示
- [ ] 跳过登录按钮正常工作
