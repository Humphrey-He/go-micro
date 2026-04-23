// user-mall/src/context/ThemeContext.tsx
import { createContext } from 'react'
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
