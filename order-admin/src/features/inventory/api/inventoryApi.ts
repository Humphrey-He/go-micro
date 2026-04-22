export interface InventoryItem {
  skuId: string
  available: number
  reserved: number
  updatedAt?: string
}

export const getInventoryList = async (): Promise<InventoryItem[]> => {
  const res = await fetch('http://localhost:8082/inventory')
  const json = await res.json()
  if (json.code !== 0) {
    throw new Error(json.message || '获取库存列表失败')
  }
  return json.data || []
}
