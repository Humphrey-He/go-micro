import { useState, useCallback, useEffect } from 'react'
import { getDashboardStats } from '../dashboardApi'
import type { DashboardStats } from '../types/stats'
import type { DashboardStatsParams } from '../dashboardApi'

interface UseDashboardStatsResult {
  loading: boolean
  data: DashboardStats | null
  error: string | null
  refresh: () => Promise<void>
}

export const useDashboardStats = (params?: DashboardStatsParams): UseDashboardStatsResult => {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<DashboardStats | null>(null)
  const [error, setError] = useState<string | null>(null)

  const fetchStats = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await getDashboardStats(params)
      setData(res)
    } catch {
      setError('获取统计数据失败')
    } finally {
      setLoading(false)
    }
  }, [params])

  useEffect(() => {
    fetchStats()
  }, [fetchStats])

  return {
    loading,
    data,
    error,
    refresh: fetchStats,
  }
}
