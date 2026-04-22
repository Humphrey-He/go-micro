import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Form, Input, Toast } from 'antd-mobile'
import { login } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'

export default function Login() {
  const navigate = useNavigate()
  const { login: setAuth } = useAuthStore()
  const [loading, setLoading] = useState(false)
  const [loginType, setLoginType] = useState<'password' | 'sms'>('password')

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
        Toast.show('请使用密码登录')
      }
    } catch (error) {
      Toast.show('登录失败，请检查账号密码')
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
        <Form layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="phone" rules={[{ required: true, message: '请输入手机号' }]}>
            <Input placeholder="手机号" className="py-3" />
          </Form.Item>
          <Form.Item
            name="code"
            rules={[{ required: true, message: '请输入验证码' }]}
            extra={
              <span className="text-primary-500 text-sm">
                获取验证码
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