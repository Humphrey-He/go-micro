import { get, post, del, put } from '@/api/request'

// ============ 价格监控 API ============

// 设置价格监控
export interface SetPriceWatchParams {
  sku_id: string
  target_price?: number  // 可选，NULL表示任意降幅提醒
}

export interface SetPriceWatchResponse {
  watch_id: number
  current_price: number
  target_price?: number
}

export const setPriceWatch = (params: SetPriceWatchParams) =>
  post<SetPriceWatchResponse>('/price-watch', params)

// 取消价格监控
export const cancelPriceWatch = (skuId: string) =>
  del<void>(`/price-watch/${skuId}`)

// 获取我的价格监控列表
export interface PriceWatchItem {
  watch_id: number
  sku_id: string
  product_name: string
  image: string
  current_price: number
  original_price: number
  target_price?: number
  notify_enabled: boolean
  lowest_price: number
  lowest_price_date: string
  created_at: string
  price_trend: 'down' | 'stable' | 'up'
}

export interface PriceWatchListResponse {
  items: PriceWatchItem[]
  page: number
  page_size: number
  total: number
}

export interface GetPriceWatchListParams {
  page?: number
  page_size?: number
  status?: 'active' | 'all'
}

export const getPriceWatchList = (params?: GetPriceWatchListParams) =>
  get<PriceWatchListResponse>('/price-watch/list', { params })

// 更新监控设置
export interface UpdatePriceWatchParams {
  target_price?: number
  notify_enabled?: boolean
}

export const updatePriceWatch = (skuId: string, params: UpdatePriceWatchParams) =>
  put<void>(`/price-watch/${skuId}`, params)

// 获取商品价格走势
export interface PricePoint {
  date: string
  price: number
}

export interface PriceHistoryResponse {
  sku_id: string
  current_price: number
  lowest_price: number
  lowest_date: string
  highest_price: number
  highest_date: string
  average_price: number
  price_points: PricePoint[]
}

export type PriceHistoryPeriod = '30d' | '90d' | '1y'

export const getPriceHistory = (skuId: string, period: PriceHistoryPeriod = '30d') =>
  get<PriceHistoryResponse>(`/products/${skuId}/price-history`, { params: { period } })

// 价格提醒通知列表
export interface PriceWatchNotification {
  notification_id: number
  sku_id: string
  product_name: string
  image: string
  old_price: number
  new_price: number
  discount_amount: number
  discount_rate: number
  is_read: boolean
  created_at: string
}

export interface PriceWatchNotificationResponse {
  items: PriceWatchNotification[]
  unread_count: number
  page: number
  page_size: number
  total: number
}

export interface GetPriceWatchNotificationsParams {
  page?: number
  page_size?: number
}

export const getPriceWatchNotifications = (params?: GetPriceWatchNotificationsParams) =>
  get<PriceWatchNotificationResponse>('/notifications/price-watch', { params })
