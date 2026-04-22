import { get, post } from '@/api/request'
import type {
  OrderListParams,
  OrderListResponse,
  OrderDetailResponse,
} from '@/features/orders/types/order'

const buildQueryString = (params: OrderListParams): string => {
  const filtered: Record<string, string> = {}
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '' && value !== null) {
      filtered[key] = String(value)
    }
  }
  return new URLSearchParams(filtered).toString()
}

export const getOrderList = (params: OrderListParams) =>
  get<OrderListResponse>(`/admin/orders?${buildQueryString(params)}`)

export const getOrderDetail = (orderNo: string) =>
  get<OrderDetailResponse>(`/order-views/${orderNo}`)

export const cancelOrder = (orderId: string) =>
  post<{ code: number; message: string }>(`/orders/${orderId}/cancel`)
