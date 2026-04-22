import { useState, useCallback } from 'react'
import { getRefundList } from '../api/refundApi'
import type { Refund, RefundListParams } from '../types/refund'

interface UseRefundListResult {
  loading: boolean
  data: Refund[]
  total: number
  fetchData: (params: RefundListParams) => Promise<void>
}

export const useRefundList = (): UseRefundListResult => {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<Refund[]>([])
  const [total, setTotal] = useState(0)

  const fetchData = useCallback(async (params: RefundListParams) => {
    setLoading(true)
    try {
      const res = await getRefundList(params)
      setData(res.refunds || [])
      setTotal(res.total || 0)
    } finally {
      setLoading(false)
    }
  }, [])

  return { loading, data, total, fetchData }
}
