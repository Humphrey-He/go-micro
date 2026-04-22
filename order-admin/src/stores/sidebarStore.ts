import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export interface MenuItem {
  key: string
  icon: string
  label: string
}

export interface SidebarState {
  collapsed: boolean
  floatMode: boolean
  menuOrder: string[]
  menuItems: MenuItem[]
  toggleCollapse: () => void
  toggleFloat: () => void
  setMenuOrder: (order: string[]) => void
  setMenuItems: (items: MenuItem[]) => void
}

const defaultMenuItems: MenuItem[] = [
  { key: '/dashboard', label: '运营看板', icon: 'DashboardOutlined' },
  { key: '/orders', label: '订单管理', icon: 'ShoppingOutlined' },
  { key: '/payments', label: '支付管理', icon: 'CreditCardOutlined' },
  { key: '/refunds', label: '退款管理', icon: 'RollbackOutlined' },
  { key: '/inventory', label: '库存管理', icon: 'AppstoreOutlined' },
]

export const useSidebarStore = create<SidebarState>()(
  persist(
    (set) => ({
      collapsed: false,
      floatMode: false,
      menuOrder: defaultMenuItems.map((i) => i.key),
      menuItems: defaultMenuItems,
      toggleCollapse: () => set((s) => ({ collapsed: !s.collapsed, floatMode: false })),
      toggleFloat: () => set((s) => ({ floatMode: !s.floatMode, collapsed: false })),
      setMenuOrder: (order) => set({ menuOrder: order }),
      setMenuItems: (items) => set({ menuItems: items }),
    }),
    { name: 'sidebar-config' }
  )
)