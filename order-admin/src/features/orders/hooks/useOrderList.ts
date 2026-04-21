import { useState, useCallback } from 'react'
import { getOrderList, cancelOrder } from '../ordersApi'
import type { OrderListItem, OrderListParams } from '../types/order'

interface UseOrderListResult {
  loading: boolean
  data: OrderListItem[]
  total: number
  fetchData: (params: OrderListParams) => Promise<void>
  cancelOrder: (orderId: string) => Promise<void>
}

export const useOrderList = (): UseOrderListResult => {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<OrderListItem[]>([])
  const [total, setTotal] = useState(0)

  const fetchData = useCallback(async (params: OrderListParams) => {
    setLoading(true)
    try {
      const res = await getOrderList(params)
      setData(res.orders || [])
      setTotal(res.total || 0)
    } finally {
      setLoading(false)
    }
  }, [])

  const handleCancelOrder = useCallback(async (orderId: string) => {
    await cancelOrder(orderId)
  }, [])

  return {
    loading,
    data,
    total,
    fetchData,
    cancelOrder: handleCancelOrder,
  }
}
