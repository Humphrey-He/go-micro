export type PaymentStatus = 'PENDING' | 'SUCCESS' | 'FAILED' | 'REFUNDED'

export interface Payment {
  payment_id: string
  order_id: string
  user_id: string
  amount: number
  status: PaymentStatus
  payment_method?: string
  transaction_id?: string
  paid_at?: string
  created_at: string
  updated_at: string
}

export interface PaymentListResponse {
  payments: Payment[]
  total: number
  page: number
  page_size: number
}

export interface PaymentListParams {
  page?: number
  page_size?: number
  order_id?: string
  status?: PaymentStatus
  startDate?: string
  endDate?: string
}
