import { useParams, useSearchParams, useNavigate } from 'react-router-dom'
import { Button, Card, Result } from 'antd-mobile'

export default function PaymentResult() {
  const { orderNo } = useParams<{ orderNo: string }>()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const status = searchParams.get('status')

  const isSuccess = status === 'success'

  return (
    <div className="min-h-screen bg-gray-50">
      <Card className="m-2">
        <Result
          status={isSuccess ? 'success' : 'error'}
          title={isSuccess ? '支付成功' : '支付失败'}
          description={isSuccess ? '感谢您的购买，订单已创建' : '请重新支付或联系客服'}
        />
      </Card>

      <Card className="m-2">
        <div className="flex justify-between py-2">
          <span className="text-gray-500">订单号</span>
          <span className="font-mono">{orderNo}</span>
        </div>
        {isSuccess && (
          <div className="flex justify-between py-2">
            <span className="text-gray-500">支付方式</span>
            <span>余额支付</span>
          </div>
        )}
      </Card>

      <div className="p-4 space-y-3">
        <Button
          block
          color="primary"
          size="large"
          onClick={() => navigate('/orders')}
        >
          查看订单
        </Button>
        <Button
          block
          size="large"
          onClick={() => navigate('/')}
        >
          返回首页
        </Button>
      </div>
    </div>
  )
}