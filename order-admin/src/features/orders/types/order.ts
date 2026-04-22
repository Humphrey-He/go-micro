export type OrderStatus =
  | 'CREATED'
  | 'RESERVED'
  | 'PROCESSING'
  | 'SUCCESS'
  | 'FAILED'
  | 'CANCELED'
  | 'TIMEOUT'

export type ViewStatus =
  | 'PENDING'
  | 'PROCESSING'
  | 'SUCCESS'
  | 'FAILED'
  | 'DEAD'
  | 'CANCELED'
  | 'TIMEOUT'

export type PaymentStatus = 'PENDING' | 'PAID' | 'SUCCESS' | 'FAILED' | 'REFUNDED'

export interface OrderItem {
  sku_id: string
  quantity: number
  price: number
}

export interface OrderListItem {
  order_id: string
  biz_no: string
  user_id: string
  status: OrderStatus
  total_amount: number
  item_count: number
  payment_status: PaymentStatus
  created_at: number
  updated_at: number
}

export interface Order {
  order_id: string
  biz_no: string
  user_id: string
  status: OrderStatus
  total_amount: number
  items: OrderItem[]
  view_status?: ViewStatus
  payment_status?: PaymentStatus
  created_at?: number
  updated_at?: number
}

export interface OrderListParams {
  page?: number
  page_size?: number
  order_no?: string
  user_id?: string
  status?: OrderStatus
  start_time?: number
  end_time?: number
  sort_by?: string
  sort_order?: 'asc' | 'desc'
}

export interface OrderListResponse {
  orders: OrderListItem[]
  total: number
  page: number
  page_size: number
}

export interface OrderDetailResponse extends Order {
  payment_status: PaymentStatus
}
