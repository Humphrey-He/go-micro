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
