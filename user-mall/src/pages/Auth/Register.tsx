import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Form, Input, Button, Toast } from 'antd-mobile'
import { register, sendSms } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'

export default function Register() {
  const navigate = useNavigate()
  const { login: setAuth } = useAuthStore()
  const [loading, setLoading] = useState(false)
  const [sendingCode, setSendingCode] = useState(false)
  const [countdown, setCountdown] = useState(0)

  const handleSendCode = async (phone: string) => {
    try {
      setSendingCode(true)
      await sendSms({ phone, type: 'register' })
      Toast.show('验证码已发送')
      setCountdown(60)
    } catch (error) {
      Toast.show('发送失败，请稍后重试')
    } finally {
      setSendingCode(false)
    }
  }

  const handleSubmit = async (values: { phone: string; code: string; password: string; confirmPassword: string }) => {
    if (values.password !== values.confirmPassword) {
      Toast.show('两次密码输入不一致')
      return
    }
    try {
      setLoading(true)
      const res = await register({
        phone: values.phone,
        code: values.code,
        password: values.password,
      })
      setAuth(res.token, res.user)
      Toast.show('注册成功')
      navigate('/')
    } catch (error) {
      Toast.show('注册失败，请检查验证码')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-white p-6">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">注册账号</h1>
        <p className="text-gray-500 mt-2">创建新账号，享受购物乐趣</p>
      </div>

      <Form layout="vertical" onFinish={handleSubmit}>
        <Form.Item
          name="phone"
          rules={[
            { required: true, message: '请输入手机号' },
            { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确' },
          ]}
        >
          <Input placeholder="手机号" className="py-3" />
        </Form.Item>

        <Form.Item
          name="code"
          rules={[{ required: true, message: '请输入验证码' }]}
          extra={
            <button
              className="text-primary-500 text-sm"
              disabled={countdown > 0 || sendingCode}
              onClick={() => {
                const phone = (document.querySelector('[name=phone]') as HTMLInputElement)?.value
                if (phone) handleSendCode(phone)
              }}
              type="button"
            >
              {countdown > 0 ? `${countdown}s` : '获取验证码'}
            </button>
          }
        >
          <Input placeholder="验证码" className="py-3" />
        </Form.Item>

        <Form.Item
          name="password"
          rules={[
            { required: true, message: '请输入密码' },
            { min: 8, message: '密码至少8位' },
          ]}
        >
          <Input type="password" placeholder="设置密码（8-20位）" className="py-3" />
        </Form.Item>

        <Form.Item
          name="confirmPassword"
          rules={[
            { required: true, message: '请确认密码' },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value || getFieldValue('password') === value) {
                  return Promise.resolve()
                }
                return Promise.reject(new Error('两次密码输入不一致'))
              },
            }),
          ]}
        >
          <Input type="password" placeholder="确认密码" className="py-3" />
        </Form.Item>

        <div className="my-4 text-sm text-gray-500">
          <label className="flex items-start gap-2">
            <input type="checkbox" className="mt-1" required />
            <span>
              我已阅读并同意
              <span className="text-primary-500">《用户协议》</span>和
              <span className="text-primary-500">《隐私政策》</span>
            </span>
          </label>
        </div>

        <Button block color="primary" size="large" loading={loading} type="submit">
          注册
        </Button>
      </Form>

      <div className="fixed bottom-8 left-0 right-0 text-center text-sm text-gray-400">
        已有账号？
        <span className="text-primary-500" onClick={() => navigate('/login')}>
          立即登录
        </span>
      </div>
    </div>
  )
}