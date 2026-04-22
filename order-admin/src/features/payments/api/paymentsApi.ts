import { get } from '@/api/request'
import type { Payment, PaymentListParams, PaymentListResponse } from '../types/payment'

const buildQueryString = (params: PaymentListParams): string => {
  const filtered: Record<string, string> = {}
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '' && value !== null) {
      filtered[key] = String(value)
    }
  }
  return new URLSearchParams(filtered).toString()
}

export const getPaymentList = (params: PaymentListParams) =>
  get<PaymentListResponse>(`/admin/payments?${buildQueryString(params)}`)

export const getPaymentDetail = (id: string) =>
  get<Payment>(`/payments/${id}`)
