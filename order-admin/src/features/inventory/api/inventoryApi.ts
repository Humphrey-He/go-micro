export interface InventoryItem {
  sku_id: string
  available: number
  reserved: number
  updated_at?: string
}

export const getInventoryList = async (): Promise<InventoryItem[]> => {
  const res = await fetch('/api/v1/admin/inventory')
  const json = await res.json()
  if (json.code !== 0) {
    throw new Error(json.message || '获取库存列表失败')
  }
  return json.data || []
}
