import { get, put } from '@/api/request'
import type { Notification, NotificationConfig } from './types/notification'

export interface NotificationListParams {
  page?: number
  page_size?: number
}

const buildQueryString = (params: NotificationListParams): string => {
  const filtered: Record<string, string> = {}
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '' && value !== null) {
      filtered[key] = String(value)
    }
  }
  return new URLSearchParams(filtered).toString()
}

export const getNotifications = (params?: NotificationListParams) =>
  get<{
    notifications: Notification[]
    unread_count: number
    page: number
    page_size: number
  }>(params ? `/notifications?${buildQueryString(params)}` : '/notifications')

export const getUnreadCount = () =>
  get<{ count: number }>('/notifications/unread-count')

export const markAsRead = (id: number) =>
  put<{ success: boolean }>(`/notifications/${id}/read`)

export const markAllAsRead = () =>
  put<{ success: boolean }>('/notifications/read-all')

export const getNotificationConfig = (type: string) =>
  get<NotificationConfig>(`/notification/configs?type=${encodeURIComponent(type)}`)

export const updateNotificationConfig = (config: NotificationConfig) =>
  put<{ success: boolean }>('/notification/configs', config)