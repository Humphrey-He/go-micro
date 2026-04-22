import { get } from '@/api/request'
import type { DashboardStats } from './types/stats'

export const getDashboardStats = () =>
  get<DashboardStats>('/admin/dashboard/stats')
