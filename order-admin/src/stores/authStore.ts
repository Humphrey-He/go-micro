import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface UserInfo {
  userId: string
  username: string
  role: string
}

interface AuthState {
  token: string | null
  userInfo: UserInfo | null
  setAuth: (token: string, userInfo: UserInfo) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      userInfo: null,
      setAuth: (token, userInfo) => set({ token, userInfo }),
      logout: () => set({ token: null, userInfo: null }),
    }),
    { name: 'admin-auth' }
  )
)
