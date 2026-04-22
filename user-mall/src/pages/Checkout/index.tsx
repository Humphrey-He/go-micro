import { useNavigate } from 'react-router-dom'
import { Button, Card, Checkbox, Input, Toast } from 'antd-mobile'
import { useState } from 'react'
import { useCartStore } from '@/stores/cartStore'

export default function Checkout() {
  const navigate = useNavigate()
  const { getSelectedItems, getTotalAmount, clearSelected } = useCartStore()
  const [loading] = useState(false)
  const [remark, setRemark] = useState('')

  const selectedItems = getSelectedItems()
  const totalAmount = getTotalAmount()

  const handleSubmit = () => {
    if (selectedItems.length === 0) {
      Toast.show('请选择商品')
      return
    }
    // 模拟创建订单
    const orderNo = 'ORD' + Date.now()
    clearSelected()
    navigate(`/payment/${orderNo}`)
  }

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      {/* 收货地址 */}
      <Card className="m-2">
        <div className="flex items-center justify-between">
          <div>
            <div className="font-bold">张三 138****8888</div>
            <div className="text-sm text-gray-500 mt-1">
              广东省深圳市南山区科技园南路88号
            </div>
          </div>
          <span className="text-gray-400">&gt;</span>
        </div>
      </Card>

      {/* 商品列表 */}
      <Card className="m-2">
        <div className="space-y-3">
          {selectedItems.map((item) => (
            <div key={item.id} className="flex items-center gap-3">
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

      {/* 备注 */}
      <Card className="m-2">
        <div className="text-sm">订单备注</div>
        <Input
          placeholder="选填，可备注特殊需求"
          value={remark}
          onChange={setRemark}
          className="mt-2"
        />
      </Card>

      {/* 支付方式 */}
      <Card className="m-2">
        <div className="font-bold mb-3">支付方式</div>
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span className="text-lg">💼</span>
              <span>余额支付</span>
            </div>
            <Checkbox checked />
          </div>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span className="text-lg">💳</span>
              <span>微信支付</span>
            </div>
            <Checkbox />
          </div>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span className="text-lg">💰</span>
              <span>支付宝</span>
            </div>
            <Checkbox />
          </div>
        </div>
      </Card>

      {/* 底部结算 */}
      <div className="fixed bottom-16 left-0 right-0 bg-white border-t p-3 flex items-center justify-between">
        <div className="text-right">
          <div className="text-sm text-gray-500">合计</div>
          <div className="text-primary-500 font-bold text-xl">
            ¥{(totalAmount / 100).toFixed(2)}
          </div>
        </div>
        <Button
          color="primary"
          size="large"
          loading={loading}
          onClick={handleSubmit}
          disabled={selectedItems.length === 0}
        >
          提交订单
        </Button>
      </div>
    </div>
  )
}