import { get } from '@/api/request'
import type { Payment, PaymentListParams, PaymentListResponse } from '../types/payment'

export const getPaymentList = (params: PaymentListParams) =>
  get<PaymentListResponse>(`/admin/payments?${new URLSearchParams(params as Record<string, string>)}`)

export const getPaymentDetail = (id: string) =>
  get<Payment>(`/payments/${id}`)
