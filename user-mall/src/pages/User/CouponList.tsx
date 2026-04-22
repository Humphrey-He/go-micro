import { useState, useEffect } from 'react'
import { Empty, Tabs, Skeleton } from 'antd-mobile'
import type { Coupon } from '@/api/coupon'

const mockCoupons: Coupon[] = [
  {
    id: '1',
    name: '新人专享券',
    type: 'CASH',
    value: 20,
    min_amount: 100,
    status: 'AVAILABLE',
    valid_from: '2024-01-01',
    valid_until: '2024-12-31',
  },
  {
    id: '2',
    name: '满减券',
    type: 'DISCOUNT',
    value: 10,
    min_amount: 50,
    status: 'AVAILABLE',
    valid_from: '2024-01-01',
    valid_until: '2024-12-31',
  },
]

export default function CouponList() {
  const [coupons, setCoupons] = useState<Coupon[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState('AVAILABLE')

  useEffect(() => {
    loadCoupons()
  }, [])

  const loadCoupons = async () => {
    try {
      setLoading(true)
      setCoupons(mockCoupons)
    } finally {
      setLoading(false)
    }
  }

  const filteredCoupons = coupons.filter((c) => c.status === activeTab)

  const getCouponStyle = (coupon: Coupon) => {
    if (coupon.status === 'USED' || coupon.status === 'EXPIRED') {
      return 'opacity-50 bg-gray-100'
    }
    return 'bg-gradient-to-r from-primary-500 to-primary-400'
  }

  const formatValue = (coupon: Coupon) => {
    if (coupon.type === 'CASH') {
      return `¥${coupon.value}`
    }
    return `${coupon.value}折`
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <Tabs activeKey={activeTab} onChange={(key) => setActiveTab(key)}>
        <Tabs.Tab title="可用" key="AVAILABLE" />
        <Tabs.Tab title="已使用" key="USED" />
        <Tabs.Tab title="已过期" key="EXPIRED" />
      </Tabs>

      {loading ? (
        <div className="p-4 space-y-4">
          <Skeleton animated />
          <Skeleton animated />
        </div>
      ) : filteredCoupons.length === 0 ? (
        <Empty description="暂无优惠券" />
      ) : (
        <div className="p-2 space-y-3">
          {filteredCoupons.map((coupon) => (
            <div
              key={coupon.id}
              className={`rounded-lg p-4 flex ${getCouponStyle(coupon)}`}
            >
              <div className="text-white text-center w-24">
                <div className="text-2xl font-bold">{formatValue(coupon)}</div>
                <div className="text-xs mt-1 opacity-80">
                  {coupon.min_amount > 0 ? `满${coupon.min_amount}可用` : '无门槛'}
                </div>
              </div>
              <div className="flex-1 text-white pl-4 border-l border-white border-opacity-30">
                <div className="font-bold">{coupon.name}</div>
                <div className="text-xs mt-1 opacity-80">
                  有效期至 {coupon.valid_until}
                </div>
                {coupon.status === 'AVAILABLE' && (
                  <div className="text-xs mt-2 bg-white text-primary-500 inline-block px-2 py-0.5 rounded">
                    待使用
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}