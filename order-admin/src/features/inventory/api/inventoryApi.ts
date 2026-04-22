import { get } from '@/api/request'

export interface InventoryItem {
  sku_id: string
  available: number
  reserved: number
  updated_at?: string
}

export const getInventoryList = async (): Promise<InventoryItem[]> => {
  const data = await get<InventoryItem[]>('/admin/inventory')
  return data || []
}
