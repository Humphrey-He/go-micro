import { get, post } from '@/api/request'
import type {
  OrderListParams,
  OrderListResponse,
  OrderItem,
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
  get<{ orderId: string; bizNo: string; userId: string; status: string; totalAmount: number; items: { skuId: string; quantity: number; price: number }[]; paymentStatus: string; viewStatus: string }>(`/order-views/${orderNo}`).then((res) => ({
    order_id: res.orderId,
    biz_no: res.bizNo,
    user_id: res.userId,
    status: res.status as OrderDetailResponse['status'],
    total_amount: res.totalAmount,
    items: res.items.map((item) => ({ sku_id: item.skuId, quantity: item.quantity, price: item.price } as OrderItem)),
    payment_status: res.paymentStatus as OrderDetailResponse['payment_status'],
    view_status: (res.viewStatus || undefined) as OrderDetailResponse['view_status'],
  }))

export const cancelOrder = (orderId: string) =>
  post<{ code: number; message: string }>(`/orders/${orderId}/cancel`)
