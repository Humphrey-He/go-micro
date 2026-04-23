import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Form, Input, Toast } from 'antd-mobile'
import { login, smsLogin, sendSms } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'
import SocialLoginButtons from '@/components/SocialLoginButtons'

export default function Login() {
  const navigate = useNavigate()
  const { login: setAuth } = useAuthStore()
  const [loading, setLoading] = useState(false)
  const [loginType, setLoginType] = useState<'password' | 'sms'>('password')
  const [smsCountdown, setSmsCountdown] = useState(0)
  const [smsLoading, setSmsLoading] = useState(false)
  const [form] = Form.useForm()

  // 验证码倒计时
  useEffect(() => {
    if (smsCountdown > 0) {
      const timer = setTimeout(() => setSmsCountdown(smsCountdown - 1), 1000)
      return () => clearTimeout(timer)
    }
  }, [smsCountdown])

  const handleSendSms = async () => {
    const phone = form.getFieldValue('phone')
    if (!phone) {
      Toast.show('请输入手机号')
      return
    }
    if (!/^1[3-9]\d{9}$/.test(phone)) {
      Toast.show('请输入正确的手机号')
      return
    }
    try {
      setSmsLoading(true)
      await sendSms({ phone, type: 'login' })
      Toast.show('验证码已发送')
      setSmsCountdown(60)
    } catch (error) {
      Toast.show('发送失败，请稍后重试')
    } finally {
      setSmsLoading(false)
    }
  }

  const handleSubmit = async (values: { account?: string; password?: string; phone?: string; code?: string }) => {
    try {
      setLoading(true)
      if (loginType === 'password') {
        const res = await login({
          account: values.account || '',
          password: values.password || '',
        })
        setAuth(res.token, res.user)
        Toast.show('登录成功')
        navigate('/')
      } else {
        const res = await smsLogin({
          phone: values.phone || '',
          code: values.code || '',
        })
        setAuth(res.token, res.user)
        Toast.show('登录成功')
        navigate('/')
      }
    } catch (error) {
      Toast.show('登录失败，请检查账号信息')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-white p-6">
      <div className="mt-12 mb-8">
        <h1 className="text-2xl font-bold text-gray-900">欢迎回来</h1>
        <p className="text-gray-500 mt-2">登录后享受更多权益</p>
      </div>

      {/* 登录方式切换 */}
      <div className="flex mb-6">
        <div
          className={`flex-1 py-2 text-center border-b-2 ${
            loginType === 'password' ? 'border-primary-500 text-primary-500' : 'border-gray-200 text-gray-500'
          }`}
          onClick={() => setLoginType('password')}
        >
          密码登录
        </div>
        <div
          className={`flex-1 py-2 text-center border-b-2 ${
            loginType === 'sms' ? 'border-primary-500 text-primary-500' : 'border-gray-200 text-gray-500'
          }`}
          onClick={() => setLoginType('sms')}
        >
          验证码登录
        </div>
      </div>

      {loginType === 'password' ? (
        <Form layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="account" rules={[{ required: true, message: '请输入手机号或用户名' }]}>
            <Input placeholder="手机号/用户名" className="py-3" />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input type="password" placeholder="密码" className="py-3" />
          </Form.Item>
          <div className="flex justify-between text-sm mb-4">
            <span className="text-primary-500" onClick={() => navigate('/register')}>
              忘记密码？
            </span>
          </div>
          <Button block color="primary" size="large" loading={loading} type="submit">
            登录
          </Button>
        </Form>
      ) : (
        <Form layout="vertical" onFinish={handleSubmit} form={form}>
          <Form.Item name="phone" rules={[{ required: true, message: '请输入手机号' }]}>
            <Input placeholder="手机号" className="py-3" />
          </Form.Item>
          <Form.Item
            name="code"
            rules={[{ required: true, message: '请输入验证码' }]}
            extra={
              <span
                className={`text-sm ${smsCountdown > 0 ? 'text-gray-400' : 'text-primary-500'}`}
                onClick={smsCountdown > 0 ? undefined : handleSendSms}
              >
                {smsCountdown > 0 ? `${smsCountdown}s后重发` : smsLoading ? '发送中...' : '获取验证码'}
              </span>
            }
          >
            <Input placeholder="验证码" className="py-3" />
          </Form.Item>
          <Button block color="primary" size="large" loading={loading} type="submit">
            登录
          </Button>
        </Form>
      )}

      {/* 社交登录 */}
      <div className="mt-8">
        <div className="relative">
          <div className="absolute inset-0 flex items-center">
            <div className="w-full border-t border-gray-200" />
          </div>
          <div className="relative flex justify-center text-sm">
            <span className="px-2 bg-white text-gray-500">其他登录方式</span>
          </div>
        </div>
        <div className="mt-4">
          <SocialLoginButtons />
        </div>
      </div>

      <div className="mt-6 text-center text-sm text-gray-500">
        登录即表示同意
        <span className="text-primary-500">《用户协议》</span>和
        <span className="text-primary-500">《隐私政策》</span>
      </div>

      <div className="fixed bottom-8 left-0 right-0 text-center text-sm text-gray-400">
        还没有账号？
        <span className="text-primary-500" onClick={() => navigate('/register')}>
          立即注册
        </span>
      </div>
    </div>
  )
}