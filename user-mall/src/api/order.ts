import { get, post } from '@/api/request'

// 确认订单（计算价格）
export interface CheckoutParams {
  cart_item_ids: string[]
  address_id: string
  coupon_id?: string
  use_points?: number
  remark?: string
}

export interface CheckoutResponse {
  items: Array<{
    sku_id: string
    title: string
    image: string
    attributes: string[]
    price: number
    quantity: number
    subtotal: number
  }>
  address: {
    id: string
    receiver: string
    phone: string
    detail: string
  }
  price_summary: {
    goods_amount: number
    freight_amount: number
    coupon_discount: number
    points_discount: number
    total_amount: number
  }
}

export const checkout = (params: CheckoutParams) =>
  post<CheckoutResponse>('/orders/checkout', params)

// 创建订单
export interface CreateOrderParams extends CheckoutParams {}

export interface CreateOrderResponse {
  order_no: string
  total_amount: number
  pay_amount: number
  pay_deadline: string
}

export const createOrder = (params: CreateOrderParams) =>
  post<CreateOrderResponse>('/orders', params)

// 订单列表
export type OrderStatus = 'ALL' | 'PENDING_PAYMENT' | 'PAID' | 'SHIPPED' | 'CONFIRMED' | 'COMPLETED' | 'CANCELED'

export interface Order {
  order_no: string
  status: OrderStatus
  status_text: string
  total_amount: number
  pay_amount: number
  created_at: string
  shop?: {
    shop_id: string
    name: string
  }
  items: Array<{
    sku_id: string
    title: string
    image: string
    attributes: string[]
    price: number
    quantity: number
  }>
  address?: {
    name: string
    phone: string
    detail: string
  }
  action_buttons?: string[]
}

export interface OrderListResponse {
  orders: Order[]
  pagination: {
    page: number
    page_size: number
    total: number
  }
  status_counts: Record<string, number>
}

export const getOrderList = (params?: { page?: number; status?: OrderStatus }) =>
  get<OrderListResponse>('/user/orders', { params })

// 订单详情
export interface OrderDetailResponse {
  order_no: string
  status: OrderStatus
  status_text: string
  total_amount: number
  pay_amount: number
  created_at: string
  pay_time?: string
  ship_time?: string
  receive_time?: string
  address: {
    receiver: string
    phone: string
    province: string
    city: string
    district: string
    detail: string
  }
  shop: {
    shop_id: string
    name: string
    phone: string
  }
  remark?: string
  logistics?: {
    company: string
    tracking_no: string
    status: string
    steps: Array<{
      time: string
      status: string
      description: string
    }>
  }
  action_logs: Array<{
    action: string
    time: string
  }>
}

export const getOrderDetail = (orderNo: string) =>
  get<OrderDetailResponse>(`/user/orders/${orderNo}`)

// 取消订单
export const cancelOrder = (orderNo: string, reason: string) =>
  post<void>(`/orders/${orderNo}/cancel`, { reason })

// 确认收货
export const confirmOrder = (orderNo: string) =>
  post<void>(`/orders/${orderNo}/confirm`)

// 支付
export interface PaymentInfo {
  order_no: string
  pay_amount: number
  pay_deadline: string
  payment_methods: Array<{
    id: string
    name: string
    icon: string
  }>
}

export const getPaymentInfo = (orderNo: string) =>
  get<PaymentInfo>(`/payments/${orderNo}`)

export interface CreatePaymentParams {
  order_no: string
  method: string
}

export interface CreatePaymentResponse {
  payment_id: string
  order_no: string
  amount: number
  status: string
  qr_code?: string
  h5_url?: string
}

export const createPayment = (params: CreatePaymentParams) =>
  post<CreatePaymentResponse>('/payments', params)