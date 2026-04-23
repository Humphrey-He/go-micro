import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { logout as logoutApi } from '@/api/auth'

export interface UserInfo {
  user_id: string
  phone: string
  nickname: string
  avatar: string
  level: number
  member_title: string
  points: number
  phone_bound?: boolean
  is_new_user?: boolean
}

interface SocialBinding {
  platform: string
  openid: string
  bound_at: string
}

interface AuthState {
  token: string | null
  userInfo: UserInfo | null
  isLoggedIn: boolean
  socialBindings: SocialBinding[]
  login: (token: string, userInfo: UserInfo) => void
  logout: () => Promise<void>
  updateUserInfo: (info: Partial<UserInfo>) => void
  setSocialBindings: (bindings: SocialBinding[]) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      userInfo: null,
      isLoggedIn: false,
      socialBindings: [],

      login: (token, userInfo) => set({ token, userInfo, isLoggedIn: true }),

      logout: async () => {
        try {
          await logoutApi()
        } catch (error) {
          console.error('Logout API failed:', error)
        } finally {
          set({ token: null, userInfo: null, isLoggedIn: false })
        }
      },

      updateUserInfo: (info) =>
        set((state) => ({
          userInfo: state.userInfo ? { ...state.userInfo, ...info } : null,
        })),

      setSocialBindings: (bindings) => set({ socialBindings: bindings }),
    }),
    { name: 'mall-auth' }
  )
)