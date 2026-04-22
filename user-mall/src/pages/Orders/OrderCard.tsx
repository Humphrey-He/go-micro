import { Card, Tag } from 'antd-mobile'
import type { Order } from '@/api/order'

interface Props {
  order: Order
  onClick: () => void
}

const statusColors: Record<string, string> = {
  PENDING_PAYMENT: 'warning',
  PAID: 'primary',
  SHIPPED: 'primary',
  CONFIRMED: 'success',
  COMPLETED: 'success',
  CANCELED: 'default',
}

export default function OrderCard({ order, onClick }: Props) {
  return (
    <Card className="cursor-pointer" onClick={onClick}>
      <div className="flex justify-between items-center mb-2">
        <span className="text-sm text-gray-500">订单号：{order.order_no}</span>
        <Tag color={statusColors[order.status] || 'default'}>{order.status_text}</Tag>
      </div>

      <div className="space-y-2">
        {order.items.slice(0, 3).map((item) => (
          <div key={item.sku_id} className="flex items-center gap-2">
            <img
              src={item.image}
              alt=""
              className="w-12 h-12 rounded object-cover bg-gray-100"
            />
            <div className="flex-1 min-w-0">
              <div className="text-sm line-clamp-1">{item.title}</div>
              <div className="text-xs text-gray-400">{item.attributes.join(' / ')}</div>
            </div>
            <div className="text-right">
              <div className="text-sm">¥{(item.price / 100).toFixed(2)}</div>
              <div className="text-xs text-gray-400">x{item.quantity}</div>
            </div>
          </div>
        ))}
        {order.items.length > 3 && (
          <div className="text-sm text-gray-400 text-center">
            还有{order.items.length - 3}件商品
          </div>
        )}
      </div>

      <div className="flex justify-between items-center mt-3 pt-3 border-t">
        <span className="text-sm text-gray-500">
          {new Date(order.created_at).toLocaleDateString()}
        </span>
        <div className="text-right">
          <span className="text-primary-500 font-bold">
            ¥{(order.pay_amount / 100).toFixed(2)}
          </span>
        </div>
      </div>
    </Card>
  )
}