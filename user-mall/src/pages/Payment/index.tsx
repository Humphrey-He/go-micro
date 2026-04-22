import { useParams, useNavigate } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { Button, Card, Toast } from 'antd-mobile'

export default function Payment() {
  const { orderNo } = useParams<{ orderNo: string }>()
  const navigate = useNavigate()
  const [countdown, setCountdown] = useState(30 * 60) // 30分钟倒计时
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    const timer = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          clearInterval(timer)
          return 0
        }
        return prev - 1
      })
    }, 1000)
    return () => clearInterval(timer)
  }, [])

  const formatTime = (seconds: number) => {
    const m = Math.floor(seconds / 60)
    const s = seconds % 60
    return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
  }

  const handlePay = async () => {
    try {
      setLoading(true)
      // 模拟支付
      await new Promise((resolve) => setTimeout(resolve, 1500))
      Toast.show('支付成功')
      navigate(`/payment/result/${orderNo}?status=success`)
    } catch (error) {
      Toast.show('支付失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <Card className="m-2">
        <div className="text-center">
          <div className="text-4xl mb-4">💰</div>
          <div className="text-gray-500">订单金额</div>
          <div className="text-3xl font-bold text-primary-500 mt-2">
            ¥{((Math.random() * 1000 + 100) / 100).toFixed(2)}
          </div>
        </div>
      </Card>

      <Card className="m-2">
        <div className="text-sm text-gray-500 mb-3">支付方式</div>
        <div className="flex items-center justify-between py-2">
          <div className="flex items-center gap-2">
            <span className="text-xl">💼</span>
            <span className="font-medium">余额支付</span>
          </div>
          <div className="text-primary-500 text-sm">可用 ¥8888.00</div>
        </div>
      </Card>

      <Card className="m-2">
        <div className="flex items-center justify-between">
          <span className="text-gray-500">订单号</span>
          <span className="font-mono text-sm">{orderNo}</span>
        </div>
        <div className="flex items-center justify-between mt-2">
          <span className="text-gray-500">支付倒计时</span>
          <span className={countdown < 300 ? 'text-red-500' : 'text-gray-500'}>
            {countdown > 0 ? formatTime(countdown) : '已过期'}
          </span>
        </div>
      </Card>

      <div className="fixed bottom-4 left-4 right-4">
        <Button
          block
          color="primary"
          size="large"
          loading={loading}
          onClick={handlePay}
          disabled={countdown === 0}
        >
          确认支付
        </Button>
      </div>
    </div>
  )
}