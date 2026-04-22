export interface DashboardStats {
  today_order_count: number
  today_order_amount: number
  pending_refund_count: number
  payment_success_rate: number
  low_stock_sku_count: number
  period?: 'day' | 'week' | 'month'
  start_time?: number
  end_time?: number
}
