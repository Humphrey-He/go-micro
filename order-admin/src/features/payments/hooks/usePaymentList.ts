import { useState, useCallback } from 'react'
import { getPaymentList } from '../api/paymentsApi'
import type { Payment, PaymentListParams } from '../types/payment'

interface UsePaymentListResult {
  loading: boolean
  data: Payment[]
  total: number
  fetchData: (params: PaymentListParams) => Promise<void>
}

export const usePaymentList = (): UsePaymentListResult => {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<Payment[]>([])
  const [total, setTotal] = useState(0)

  const fetchData = useCallback(async (params: PaymentListParams) => {
    setLoading(true)
    try {
      const res = await getPaymentList(params)
      setData(res.payments || [])
      setTotal(res.total || 0)
    } finally {
      setLoading(false)
    }
  }, [])

  return { loading, data, total, fetchData }
}
