import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface UserInfo {
  user_id: string
  phone: string
  nickname: string
  avatar: string
  level: number
  member_title: string
  points: number
}

interface AuthState {
  token: string | null
  userInfo: UserInfo | null
  isLoggedIn: boolean
  login: (token: string, userInfo: UserInfo) => void
  logout: () => void
  updateUserInfo: (info: Partial<UserInfo>) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      userInfo: null,
      isLoggedIn: false,

      login: (token, userInfo) => set({ token, userInfo, isLoggedIn: true }),

      logout: () => set({ token: null, userInfo: null, isLoggedIn: false }),

      updateUserInfo: (info) =>
        set((state) => ({
          userInfo: state.userInfo ? { ...state.userInfo, ...info } : null,
        })),
    }),
    { name: 'mall-auth' }
  )
)