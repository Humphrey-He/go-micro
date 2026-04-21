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

export type PaymentStatus = 'PENDING' | 'PAID' | 'FAILED' | 'REFUNDED'

export interface OrderItem {
  skuId: string
  quantity: number
  price: number
}

export interface OrderListItem {
  orderId: string
  bizNo: string
  userId: string
  status: OrderStatus
  totalAmount: number
  itemCount: number
  paymentStatus: PaymentStatus
  createdAt: number
  updatedAt: number
}

export interface Order {
  orderId: string
  bizNo: string
  userId: string
  status: OrderStatus
  totalAmount: number
  items: OrderItem[]
  viewStatus?: ViewStatus
  paymentStatus?: PaymentStatus
  createdAt?: number
  updatedAt?: number
}

export interface OrderListParams {
  page?: number
  pageSize?: number
  orderNo?: string
  userId?: string
  status?: OrderStatus
  startTime?: number
  endTime?: number
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}

export interface OrderListResponse {
  orders: OrderListItem[]
  total: number
  page: number
  pageSize: number
}

export interface OrderDetailResponse extends Order {
  paymentStatus: PaymentStatus
}
