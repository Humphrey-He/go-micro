import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export interface CartItem {
  id: string
  sku_id: string
  title: string
  image: string
  attributes: string[]
  price: number
  quantity: number
  stock: number
  is_selected: boolean
  is_valid: boolean
  invalid_reason?: string
  shop_id: string
  shop_name: string
}

interface CartState {
  items: CartItem[]
  addItem: (item: Omit<CartItem, 'id' | 'is_selected' | 'is_valid'>) => void
  removeItem: (id: string) => void
  updateQuantity: (id: string, quantity: number) => void
  toggleSelect: (id: string) => void
  selectAll: (selected: boolean) => void
  clearSelected: () => void
  getSelectedItems: () => CartItem[]
  getTotalAmount: () => number
}

export const useCartStore = create<CartState>()(
  persist(
    (set, get) => ({
      items: [],

      addItem: (item) => {
        const existing = get().items.find(
          (i) => i.sku_id === item.sku_id && i.attributes.join(',') === item.attributes.join(',')
        )
        if (existing) {
          set((state) => ({
            items: state.items.map((i) =>
              i.id === existing.id ? { ...i, quantity: i.quantity + item.quantity } : i
            ),
          }))
        } else {
          const newItem: CartItem = {
            ...item,
            id: `${Date.now()}-${Math.random()}`,
            is_selected: true,
            is_valid: true,
          }
          set((state) => ({ items: [...state.items, newItem] }))
        }
      },

      removeItem: (id) =>
        set((state) => ({
          items: state.items.filter((i) => i.id !== id),
        })),

      updateQuantity: (id, quantity) =>
        set((state) => ({
          items: state.items.map((i) =>
            i.id === id ? { ...i, quantity: Math.min(quantity, i.stock) } : i
          ),
        })),

      toggleSelect: (id) =>
        set((state) => ({
          items: state.items.map((i) =>
            i.id === id ? { ...i, is_selected: !i.is_selected } : i
          ),
        })),

      selectAll: (selected) =>
        set((state) => ({
          items: state.items.map((i) => ({ ...i, is_selected: selected && i.is_valid })),
        })),

      clearSelected: () =>
        set((state) => ({
          items: state.items.map((i) => ({ ...i, is_selected: false })),
        })),

      getSelectedItems: () => get().items.filter((i) => i.is_selected && i.is_valid),

      getTotalAmount: () =>
        get()
          .getSelectedItems()
          .reduce((sum, i) => sum + i.price * i.quantity, 0),
    }),
    { name: 'mall-cart' }
  )
)