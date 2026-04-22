import { useParams, useNavigate } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { Card, Button, Steps } from 'antd-mobile'

const statusSteps = [
  { title: '下单', description: '' },
  { title: '支付', description: '' },
  { title: '发货', description: '' },
  { title: '收货', description: '' },
  { title: '完成', description: '' },
]

const statusIndex: Record<string, number> = {
  PENDING_PAYMENT: 0,
  PAID: 1,
  SHIPPED: 2,
  CONFIRMED: 3,
  COMPLETED: 4,
}

interface OrderItem {
  sku_id: string
  title: string
  price: number
  quantity: number
  image: string
  attributes: string[]
}

interface MockOrder {
  order_no: string
  status: string
  status_text: string
  total_amount: number
  pay_amount: number
  created_at: string
  items: OrderItem[]
  address: {
    name: string
    phone: string
    detail: string
  }
}

export default function OrderDetail() {
  const { orderNo } = useParams<{ orderNo: string }>()
  const navigate = useNavigate()
  const [order, setOrder] = useState<MockOrder | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadDetail()
  }, [orderNo])

  const loadDetail = async () => {
    try {
      setLoading(true)
      setOrder({
        order_no: orderNo || 'mock123',
        status: 'PAID',
        status_text: '待发货',
        total_amount: 29900,
        pay_amount: 29900,
        created_at: new Date().toISOString(),
        items: [
          {
            sku_id: '1',
            title: 'iPhone 15 Pro Max 256GB 深空黑',
            price: 899900,
            quantity: 1,
            image: 'https://picsum.photos/200',
            attributes: ['颜色: 深空黑', '内存: 256GB'],
          },
        ],
        address: {
          name: '张三',
          phone: '138****8888',
          detail: '广东省深圳市南山区科技园南路88号',
        },
      })
    } finally {
      setLoading(false)
    }
  }

  if (loading || !order) {
    return (
      <div className="p-4 text-center text-gray-500">
        加载中...
      </div>
    )
  }

  const currentStep = statusIndex[order.status] ?? 0

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      {/* 物流状态 */}
      <div className="bg-primary-500 text-white p-4">
        <div className="text-lg font-bold mb-2">{order.status_text}</div>
        <div className="text-sm opacity-80">
          {order.status === 'PENDING_PAYMENT' && '请尽快完成支付'}
          {order.status === 'PAID' && '商家正在准备商品'}
          {order.status === 'SHIPPED' && '商品已发货，请注意查收'}
          {order.status === 'CONFIRMED' && '确认收货后即可评价'}
          {order.status === 'COMPLETED' && '交易已完成，感谢购买'}
        </div>
      </div>

      {/* 物流进度 */}
      <Card className="m-2">
        <Steps current={currentStep}>
          {statusSteps.map((step, index) => (
            <Steps.Step key={index} title={step.title} description={step.description} />
          ))}
        </Steps>
      </Card>

      {/* 收货地址 */}
      {order.address && (
        <Card className="m-2">
          <div className="flex items-start gap-3">
            <div className="text-xl">📍</div>
            <div>
              <div className="font-bold">{order.address.name} {order.address.phone}</div>
              <div className="text-sm text-gray-500 mt-1">{order.address.detail}</div>
            </div>
          </div>
        </Card>
      )}

      {/* 商品列表 */}
      <Card className="m-2">
        <div className="space-y-3">
          {order.items.map((item) => (
            <div key={item.sku_id} className="flex items-center gap-3">
              <img
                src={item.image}
                alt=""
                className="w-16 h-16 rounded object-cover bg-gray-100"
              />
              <div className="flex-1 min-w-0">
                <div className="text-sm line-clamp-1">{item.title}</div>
                <div className="text-xs text-gray-400 mt-0.5">
                  {item.attributes.join(' / ')}
                </div>
              </div>
              <div className="text-right">
                <div className="text-sm">¥{(item.price / 100).toFixed(2)}</div>
                <div className="text-xs text-gray-400">x{item.quantity}</div>
              </div>
            </div>
          ))}
        </div>
      </Card>

      {/* 订单信息 */}
      <Card className="m-2">
        <div className="text-sm text-gray-500 space-y-2">
          <div className="flex justify-between">
            <span>订单编号</span>
            <span className="font-mono">{order.order_no}</span>
          </div>
          <div className="flex justify-between">
            <span>创建时间</span>
            <span>{new Date(order.created_at).toLocaleString()}</span>
          </div>
          <div className="flex justify-between">
            <span>订单金额</span>
            <span className="text-primary-500">¥{(order.total_amount / 100).toFixed(2)}</span>
          </div>
        </div>
      </Card>

      {/* 底部操作 */}
      <div className="fixed bottom-0 left-0 right-0 bg-white border-t p-3 flex justify-end gap-3">
        {order.status === 'PENDING_PAYMENT' && (
          <>
            <Button>取消订单</Button>
            <Button color="primary" onClick={() => navigate(`/payment/${order.order_no}`)}>
              立即支付
            </Button>
          </>
        )}
        {order.status === 'SHIPPED' && (
          <Button color="primary">确认收货</Button>
        )}
        {order.status === 'CONFIRMED' && (
          <Button color="primary">立即评价</Button>
        )}
      </div>
    </div>
  )
}