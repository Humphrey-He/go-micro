import { useState, useCallback, useEffect } from 'react'
import {
  getNotifications,
  markAsRead,
  markAllAsRead,
} from '../notificationApi'
import type { Notification } from '../types/notification'

interface UseNotificationsResult {
  notifications: Notification[]
  unreadCount: number
  loading: boolean
  error: string | null
  page: number
  pageSize: number
  refresh: () => Promise<void>
  loadMore: () => Promise<void>
  markRead: (id: number) => Promise<void>
  markAllRead: () => Promise<void>
}

export const useNotifications = (): UseNotificationsResult => {
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [unreadCount, setUnreadCount] = useState(0)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [pageSize] = useState(20)

  const fetchNotifications = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await getNotifications({ page: 1, page_size: pageSize })
      setNotifications(res.notifications || [])
      setUnreadCount(res.unread_count || 0)
      setPage(1)
    } catch {
      setError('获取通知失败')
    } finally {
      setLoading(false)
    }
  }, [pageSize])

  const loadMore = useCallback(async () => {
    if (loading) return
    setLoading(true)
    try {
      const nextPage = page + 1
      const res = await getNotifications({ page: nextPage, page_size: pageSize })
      setNotifications((prev) => [...prev, ...(res.notifications || [])])
      setPage(nextPage)
    } catch {
      setError('加载更多失败')
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, loading])

  const markRead = useCallback(async (id: number) => {
    try {
      await markAsRead(id)
      setNotifications((prev) =>
        prev.map((n) => (n.id === id ? { ...n, is_read: true } : n))
      )
      setUnreadCount((prev) => Math.max(0, prev - 1))
    } catch {
      setError('标记已读失败')
    }
  }, [])

  const markAllRead = useCallback(async () => {
    try {
      await markAllAsRead()
      setNotifications((prev) => prev.map((n) => ({ ...n, is_read: true })))
      setUnreadCount(0)
    } catch {
      setError('标记全部已读失败')
    }
  }, [])

  useEffect(() => {
    fetchNotifications()
  }, [fetchNotifications])

  return {
    notifications,
    unreadCount,
    loading,
    error,
    page,
    pageSize,
    refresh: fetchNotifications,
    loadMore,
    markRead,
    markAllRead,
  }
}