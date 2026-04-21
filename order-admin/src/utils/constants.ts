export const ORDER_STATUS_MAP: Record<string, { label: string; color: string; bg: string }> = {
  CREATED: { label: '已创建', color: '#1677ff', bg: '#e6f4ff' },
  RESERVED: { label: '已预留', color: '#0891b2', bg: '#ecfeff' },
  PROCESSING: { label: '处理中', color: '#7c3aed', bg: '#f5f3ff' },
  SUCCESS: { label: '成功', color: '#16a34a', bg: '#f0fdf4' },
  FAILED: { label: '失败', color: '#dc2626', bg: '#fef2f2' },
  CANCELED: { label: '已取消', color: '#6b7280', bg: '#f9fafb' },
  TIMEOUT: { label: '超时', color: '#ea580c', bg: '#fff7ed' },
}

export const VIEW_STATUS_MAP: Record<string, { label: string; color: string; bg: string }> = {
  PENDING: { label: '待处理', color: '#6b7280', bg: '#f9fafb' },
  PROCESSING: { label: '处理中', color: '#7c3aed', bg: '#f5f3ff' },
  SUCCESS: { label: '成功', color: '#16a34a', bg: '#f0fdf4' },
  FAILED: { label: '失败', color: '#dc2626', bg: '#fef2f2' },
  DEAD: { label: '死单', color: '#dc2626', bg: '#fef2f2' },
  CANCELED: { label: '已取消', color: '#6b7280', bg: '#f9fafb' },
  TIMEOUT: { label: '超时', color: '#ea580c', bg: '#fff7ed' },
}

export const PAYMENT_STATUS_MAP: Record<string, { label: string; color: string; bg: string }> = {
  PENDING: { label: '待支付', color: '#d97706', bg: '#fffbeb' },
  PAID: { label: '已支付', color: '#16a34a', bg: '#f0fdf4' },
  SUCCESS: { label: '已支付', color: '#16a34a', bg: '#f0fdf4' },
  FAILED: { label: '支付失败', color: '#dc2626', bg: '#fef2f2' },
  REFUNDED: { label: '已退款', color: '#7c3aed', bg: '#f5f3ff' },
}
