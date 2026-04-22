import { get } from '@/api/request'
import type { DashboardStats } from './types/stats'

export interface DashboardStatsParams {
  start_time?: number
  end_time?: number
  period?: 'day' | 'week' | 'month'
}

const buildQueryString = (params: DashboardStatsParams): string => {
  const filtered: Record<string, string> = {}
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '' && value !== null) {
      filtered[key] = String(value)
    }
  }
  return new URLSearchParams(filtered).toString()
}

export const getDashboardStats = (params?: DashboardStatsParams) => {
  const queryString = params ? buildQueryString(params) : ''
  return get<DashboardStats>(`/admin/dashboard/stats${queryString ? `?${queryString}` : ''}`)
}
