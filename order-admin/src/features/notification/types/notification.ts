export interface Notification {
  id: number
  user_id: string
  type: 'refund_pending' | 'low_stock' | 'payment_failed' | 'daily_report' | 'weekly_report'
  title: string
  content: string
  is_read: boolean
  created_at: string
}

export interface NotificationConfig {
  id?: number
  user_id: string
  type: string
  email_enabled: boolean
  push_enabled: boolean
  threshold: number
}