import { get, post } from '@/api/request'
import type {
  OrderListParams,
  OrderListResponse,
  OrderDetailResponse,
} from '@/features/orders/types/order'

export const getOrderList = (params: OrderListParams) =>
  get<OrderListResponse>(`/admin/orders?${new URLSearchParams(params as Record<string, string>)}`)

export const getOrderDetail = (orderNo: string) =>
  get<OrderDetailResponse>(`/order-views/${orderNo}`)

export const cancelOrder = (orderId: string) =>
  post<{ code: number; message: string }>(`/orders/${orderId}/cancel`)
