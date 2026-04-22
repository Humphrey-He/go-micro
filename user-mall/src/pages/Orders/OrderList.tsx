import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Empty, Tabs } from 'antd-mobile'
import type { Order, OrderStatus } from '@/api/order'
import OrderCard from './OrderCard'

const statusTabs = [
  { key: 'ALL', label: '全部' },
  { key: 'PENDING_PAYMENT', label: '待付款' },
  { key: 'PAID', label: '待发货' },
  { key: 'SHIPPED', label: '待收货' },
  { key: 'CONFIRMED', label: '待评价' },
]

export default function OrderList() {
  const navigate = useNavigate()
  const [orders, setOrders] = useState<Order[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState<OrderStatus>('ALL')

  useEffect(() => {
    loadOrders()
  }, [activeTab])

  const loadOrders = async () => {
    try {
      setLoading(true)
      // 模拟数据
      setOrders([])
    } catch (error) {
      console.error(error)
    } finally {
      setLoading(false)
    }
  }

  const handleTabChange = (key: string) => {
    setActiveTab(key as OrderStatus)
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="sticky top-0 z-50 bg-white">
        <div className="flex items-center justify-between px-4 h-12">
          <span className="text-lg font-bold">我的订单</span>
        </div>
        <Tabs activeKey={activeTab} onChange={handleTabChange} className="bg-white">
          {statusTabs.map((tab) => (
            <Tabs.Tab title={tab.label} key={tab.key} />
          ))}
        </Tabs>
      </div>

      {/* 订单列表 */}
      {loading ? (
        <div className="p-4 text-center text-gray-500">加载中...</div>
      ) : orders.length === 0 ? (
        <Empty description="暂无订单" />
      ) : (
        <div className="p-2 space-y-3">
          {orders.map((order) => (
            <OrderCard
              key={order.order_no}
              order={order}
              onClick={() => navigate(`/order/${order.order_no}`)}
            />
          ))}
        </div>
      )}
    </div>
  )
}