import { useState, useCallback, useEffect } from 'react'
import { getUnreadCount } from '../notificationApi'

export const useUnreadCount = () => {
  const [count, setCount] = useState(0)
  const [loading, setLoading] = useState(false)

  const fetch = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getUnreadCount()
      setCount(res.count || 0)
    } catch {
      // 静默失败
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetch()
    // 轮询每30秒刷新一次
    const interval = setInterval(fetch, 30000)
    return () => clearInterval(interval)
  }, [fetch])

  return { count, loading, refresh: fetch }
}