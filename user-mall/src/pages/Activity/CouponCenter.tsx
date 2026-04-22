import { useState, useEffect } from 'react'
import { Card, Button, Tabs, Empty } from 'antd-mobile'

interface Coupon {
  id: string
  name: string
  desc: string
  type: 'CASH' | 'DISCOUNT'
  value: number
  minAmount: number
  validFrom: string
  validUntil: string
  received?: boolean
}

const mockCoupons: Coupon[] = [
  {
    id: '1',
    name: '新人专享',
    desc: '全场通用',
    type: 'CASH',
    value: 20,
    minAmount: 100,
    validFrom: '2024-01-01',
    validUntil: '2024-12-31',
  },
  {
    id: '2',
    name: '满减券',
    desc: '满50减10',
    type: 'DISCOUNT',
    value: 10,
    minAmount: 50,
    validFrom: '2024-01-01',
    validUntil: '2024-12-31',
  },
  {
    id: '3',
    name: 'VIP专享',
    desc: '满200可用',
    type: 'CASH',
    value: 50,
    minAmount: 200,
    validFrom: '2024-01-01',
    validUntil: '2024-12-31',
  },
]

export default function CouponCenter() {
  const [coupons, setCoupons] = useState<Coupon[]>([])
  const [loading, setLoading] = useState(true)

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

  const handleReceive = (id: string) => {
    setCoupons((prev) =>
      prev.map((c) => (c.id === id ? { ...c, received: true } : c))
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <Tabs>
        <Tabs.Tab title="可领取" key="available">
          {loading ? (
            <div className="p-4 text-center text-gray-500">加载中...</div>
          ) : (
            <div className="p-2 space-y-3">
              {coupons.map((coupon) => (
                <Card key={coupon.id} className="bg-white">
                  <div className="flex">
                    <div className="w-28 bg-gradient-to-br from-primary-500 to-primary-400 text-white p-3 flex flex-col items-center justify-center">
                      <div className="text-2xl font-bold">
                        {coupon.type === 'CASH'
                          ? `¥${coupon.value}`
                          : `${coupon.value}折`}
                      </div>
                      <div className="text-xs mt-1 opacity-80">
                        {coupon.minAmount > 0
                          ? `满${coupon.minAmount}可用`
                          : '无门槛'}
                      </div>
                    </div>
                    <div className="flex-1 p-3">
                      <div className="font-bold">{coupon.name}</div>
                      <div className="text-sm text-gray-500 mt-1">
                        {coupon.desc}
                      </div>
                      <div className="text-xs text-gray-400 mt-1">
                        有效期至 {coupon.validUntil}
                      </div>
                      <Button
                        size="small"
                        color="primary"
                        className="mt-2"
                        disabled={coupon.received}
                        onClick={() => handleReceive(coupon.id)}
                      >
                        {coupon.received ? '已领取' : '立即领取'}
                      </Button>
                    </div>
                  </div>
                </Card>
              ))}
            </div>
          )}
        </Tabs.Tab>
        <Tabs.Tab title="已领取" key="received">
          <Empty
            description="暂无领取记录"
          />
        </Tabs.Tab>
      </Tabs>
    </div>
  )
}