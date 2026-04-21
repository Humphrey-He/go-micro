import { useState, useCallback } from 'react'
import { getOrderDetail } from '../ordersApi'
import type { OrderDetailResponse } from '../types/order'

interface UseOrderDetailResult {
  loading: boolean
  data: OrderDetailResponse | null
  error: string | null
  fetchDetail: (orderNo: string) => Promise<void>
  reset: () => void
}

export const useOrderDetail = (): UseOrderDetailResult => {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<OrderDetailResponse | null>(null)
  const [error, setError] = useState<string | null>(null)

  const fetchDetail = useCallback(async (orderNo: string) => {
    setLoading(true)
    setError(null)
    try {
      const res = await getOrderDetail(orderNo)
      console.log('[useOrderDetail] raw response:', JSON.stringify(res))
      // API returns {code, message, data: {...}}, data is nested
      const orderData = (res as unknown as { data: OrderDetailResponse }).data
      console.log('[useOrderDetail] order data:', JSON.stringify(orderData))
      setData(orderData)
    } catch (err) {
      console.error('[useOrderDetail] error:', err)
      setError('获取订单详情失败')
    } finally {
      setLoading(false)
    }
  }, [])

  const reset = useCallback(() => {
    setData(null)
    setError(null)
  }, [])

  return {
    loading,
    data,
    error,
    fetchDetail,
    reset,
  }
}
