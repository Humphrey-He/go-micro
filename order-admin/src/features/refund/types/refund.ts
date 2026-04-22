export type RefundStatus = 'PENDING' | 'SUCCESS' | 'FAILED'

export type RefundType = 'manual' | 'payment_failed' | 'order_cancel'

export interface Refund {
  refund_id: string
  order_id: string
  refund_type: RefundType
  status: RefundStatus
  reason?: string
  retry_count: number
  next_retry_time?: string
  last_error?: string
  created_at: string
  updated_at: string
}

export interface RefundListResponse {
  refunds: Refund[]
  total: number
  page: number
  page_size: number
}

export interface RefundListParams {
  page?: number
  page_size?: number
  order_id?: string
  status?: RefundStatus
}

export interface InitiateRefundParams {
  refund_id: string
  order_id: string
  refund_type?: RefundType
  reason?: string
}
