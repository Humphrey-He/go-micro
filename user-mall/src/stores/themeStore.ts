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
