import { useState } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { Button, Form, Input, Toast } from 'antd-mobile'
import { associatePhone, sendSms } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'

export default function SocialBind() {
  const navigate = useNavigate()
  const location = useLocation()
  const { login: setAuth } = useAuthStore()
  const [loading, setLoading] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const [isNewUser, setIsNewUser] = useState(
    (location.state as { isNewUser?: boolean })?.isNewUser ?? true
  )

  const handleSendCode = async (phone: string) => {
    if (!phone) {
      Toast.show('请输入手机号')
      return
    }
    try {
      await sendSms({ phone, type: 'bind' })
      Toast.show('验证码已发送')
      setCountdown(60)
      const timer = setInterval(() => {
        setCountdown((prev) => {
          if (prev <= 1) {
            clearInterval(timer)
            return 0
          }
          return prev - 1
        })
      }, 1000)
    } catch {
      Toast.show('发送失败，请稍后重试')
    }
  }

  const handleSubmit = async (values: { phone: string; code: string }) => {
    try {
      setLoading(true)
      const response = await associatePhone({
        phone: values.phone,
        code: values.code,
        action: 'bind',
      })
      // 更新 token（关联成功后用新的 token）
      const userInfo = useAuthStore.getState().userInfo
      if (userInfo) {
        setAuth(response.token, userInfo)
      }
      Toast.show('绑定成功')
      navigate('/')
    } catch {
      Toast.show('绑定失败，请检查验证码')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-white p-6">
      <div className="mt-8 mb-8">
        <h1 className="text-2xl font-bold text-gray-900">
          {isNewUser ? '完善手机号' : '绑定手机号'}
        </h1>
        <p className="text-gray-500 mt-2">
          {isNewUser
            ? '为提供更好的服务，请绑定您的手机号'
            : '绑定手机号后可使用手机号登录'}
        </p>
      </div>

      <Form layout="vertical" onFinish={handleSubmit}>
        <Form.Item
          name="phone"
          rules={[
            { required: true, message: '请输入手机号' },
            { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确' },
          ]}
        >
          <Input placeholder="请输入手机号" className="py-3" />
        </Form.Item>

        <Form.Item
          name="code"
          rules={[{ required: true, message: '请输入验证码' }]}
          extra={
            <span
              className={`text-sm ${countdown > 0 ? 'text-gray-400' : 'text-primary-500'}`}
              onClick={() => {
                if (countdown === 0) {
                  const phone = (document.querySelector('input[placeholder="请输入手机号"]') as HTMLInputElement)?.value
                  handleSendCode(phone || '')
                }
              }}
            >
              {countdown > 0 ? `${countdown}s后重发` : '获取验证码'}
            </span>
          }
        >
          <Input placeholder="请输入验证码" className="py-3" />
        </Form.Item>

        <Form.Item>
          <Button block color="primary" size="large" loading={loading} type="submit">
            绑定
          </Button>
        </Form.Item>
      </Form>

      <div className="mt-6 text-center text-sm text-gray-400">
        <span
          className="text-primary-500"
          onClick={() => navigate('/')}
        >
          跳过
        </span>
        {' '}，后续可在设置中绑定
      </div>
    </div>
  )
}
