import { get, post } from '@/api/request'
import type { Refund, RefundListParams, RefundListResponse, InitiateRefundParams } from '../types/refund'

const buildQueryString = (params: RefundListParams): string => {
  const filtered: Record<string, string> = {}
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '' && value !== null) {
      filtered[key] = String(value)
    }
  }
  return new URLSearchParams(filtered).toString()
}

export const getRefundList = (params: RefundListParams) =>
  get<RefundListResponse>(`/admin/refunds?${buildQueryString(params)}`)

export const getRefundDetail = (id: string) =>
  get<Refund>(`/refunds/${id}`)

export const initiateRefund = (params: InitiateRefundParams) =>
  post<Refund>('/refunds', params)
