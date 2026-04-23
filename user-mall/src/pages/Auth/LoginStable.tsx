// user-mall/src/pages/Auth/LoginStable.tsx
import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Form, Input, Toast } from 'antd-mobile'
import { login, smsLogin, sendSms } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'
import SocialLoginButtons from '@/components/SocialLoginButtons'

const isDesktop = () => window.innerWidth >= 768

export default function LoginStable() {
  const navigate = useNavigate()
  const { login: setAuth } = useAuthStore()
  const [loading, setLoading] = useState(false)
  const [loginType, setLoginType] = useState<'password' | 'sms'>('password')
  const [smsCountdown, setSmsCountdown] = useState(0)
  const [smsLoading, setSmsLoading] = useState(false)
  const [form] = Form.useForm()
  const [isDesktopView, setIsDesktopView] = useState(false)

  useEffect(() => {
    const check = () => setIsDesktopView(isDesktop())
    check()
    window.addEventListener('resize', check)
    return () => window.removeEventListener('resize', check)
  }, [])

  useEffect(() => {
    if (smsCountdown > 0) {
      const t = setTimeout(() => setSmsCountdown(smsCountdown - 1), 1000)
      return () => clearTimeout(t)
    }
  }, [smsCountdown])

  const handleSendSms = async () => {
    const phone = form.getFieldValue('phone')
    if (!phone || !/^1[3-9]\d{9}$/.test(phone)) {
      Toast.show('请输入正确的手机号')
      return
    }
    try {
      setSmsLoading(true)
      await sendSms({ phone, type: 'login' })
      Toast.show('验证码已发送')
      setSmsCountdown(60)
    } catch {
      Toast.show('发送失败')
    } finally {
      setSmsLoading(false)
    }
  }

  const handleSubmit = async (values: any) => {
    try {
      setLoading(true)
      if (loginType === 'password') {
        const res = await login({ account: values.account || '', password: values.password || '' })
        setAuth(res.token, res.user)
        Toast.show('登录成功')
        navigate('/')
      } else {
        const res = await smsLogin({ phone: values.phone || '', code: values.code || '' })
        setAuth(res.token, res.user)
        Toast.show('登录成功')
        navigate('/')
      }
    } catch {
      Toast.show('登录失败，请检查账号信息')
    } finally {
      setLoading(false)
    }
  }

  const categories = ['数码家电', '服饰鞋包', '食品生鲜', '美妆护肤', '家居用品', '图书文具']
  const trustBadges = [
    { icon: '🔒', text: '安全支付' },
    { icon: '✓', text: '正品保障' },
    { icon: '🚚', text: '快速配送' },
  ]

  // Desktop Layout
  if (isDesktopView) {
    return (
      <div className="min-h-screen bg-gray-50">
        <div className="bg-white border-b border-gray-200 px-8 py-4 flex justify-between items-center">
          <div className="text-xl font-bold text-teal-700">🏪 商城</div>
          <button onClick={() => navigate('/')} className="text-gray-500 text-sm hover:text-teal-600">跳过登录 →</button>
        </div>

        <div className="flex max-w-6xl mx-auto mt-12 gap-8">
          {/* Left - Form */}
          <div className="flex-1 max-w-md">
            <div className="bg-white rounded-lg p-8 shadow-sm border border-gray-200">
              <h2 className="text-2xl font-bold text-gray-900 mb-1">登录</h2>
              <p className="text-gray-500 text-sm mb-6">欢迎回来</p>

              <div className="flex gap-4 mb-6">
                <button
                  className={`px-4 py-2 text-sm font-medium rounded ${loginType === 'password' ? 'text-teal-600 border-b-2 border-teal-600' : 'text-gray-500'}`}
                  onClick={() => setLoginType('password')}
                >密码登录</button>
                <button
                  className={`px-4 py-2 text-sm font-medium rounded ${loginType === 'sms' ? 'text-teal-600 border-b-2 border-teal-600' : 'text-gray-500'}`}
                  onClick={() => setLoginType('sms')}
                >验证码登录</button>
              </div>

              <Form layout="vertical" onFinish={handleSubmit}>
                <div className="mb-4">
                  <label className="text-sm text-gray-600">邮箱/手机</label>
                  <Input placeholder="请输入" className="mt-1 h-10 rounded" />
                </div>
                {loginType === 'password' && (
                  <div className="mb-4">
                    <label className="text-sm text-gray-600">密码</label>
                    <Input type="password" placeholder="请输入密码" className="mt-1 h-10 rounded" />
                  </div>
                )}
                {loginType === 'sms' && (
                  <div className="mb-4">
                    <label className="text-sm text-gray-600">验证码</label>
                    <div className="flex gap-2 mt-1">
                      <Input placeholder="请输入" className="h-10 rounded flex-1" />
                      <button
                        className={`px-3 h-10 rounded text-sm ${smsCountdown > 0 ? 'bg-gray-100 text-gray-400' : 'bg-teal-50 text-teal-600'}`}
                        onClick={smsCountdown > 0 ? undefined : handleSendSms}
                      >
                        {smsCountdown > 0 ? `${smsCountdown}s` : '获取验证码'}
                      </button>
                    </div>
                  </div>
                )}
                <Button block color="primary" size="large" loading={loading} type="submit" className="h-10 rounded text-base font-medium mt-4" style={{ background: '#007185' }}>
                  {loading ? '登录中...' : '登 录'}
                </Button>
              </Form>

              <div className="mt-6">
                <div className="flex items-center gap-2 mb-4">
                  <div className="flex-1 h-px bg-gray-200" />
                  <span className="text-xs text-gray-400">其他登录方式</span>
                  <div className="flex-1 h-px bg-gray-200" />
                </div>
                <SocialLoginButtons theme="stable" layout="horizontal" />
              </div>

              <div className="mt-6 flex justify-between text-sm">
                <button className="text-teal-600">忘记密码</button>
                <button onClick={() => navigate('/register')} className="text-teal-600">立即注册 →</button>
              </div>
            </div>
          </div>

          {/* Right - Features */}
          <div className="flex-1">
            <h3 className="text-lg font-bold text-gray-800 mb-4">快捷分类</h3>
            <div className="grid grid-cols-3 gap-3 mb-8">
              {categories.map((cat) => (
                <div key={cat} className="bg-white p-4 rounded-lg text-center text-sm text-gray-700 border border-gray-200 hover:border-teal-300 cursor-pointer transition-colors">
                  {cat}
                </div>
              ))}
            </div>

            <h3 className="text-lg font-bold text-gray-800 mb-4">信任保障</h3>
            <div className="flex gap-4">
              {trustBadges.map((badge) => (
                <div key={badge.text} className="flex items-center gap-2 text-sm text-gray-600">
                  <span>{badge.icon}</span>
                  <span>{badge.text}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="text-center text-xs text-gray-400 mt-12">© 2026 商城 · 品质保障</div>
      </div>
    )
  }

  // Mobile Layout
  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-white border-b border-gray-200 px-4 py-3 flex justify-between items-center">
        <div className="text-lg font-bold text-teal-700">🏪 商城</div>
        <button onClick={() => navigate('/')} className="text-gray-500 text-sm">跳过</button>
      </div>

      <div className="p-6">
        <div className="bg-white rounded-lg p-6 shadow-sm border border-gray-200">
          <h2 className="text-xl font-bold text-gray-900 mb-1">登录</h2>
          <p className="text-gray-500 text-sm mb-4">欢迎回来</p>

          <div className="flex gap-2 mb-4 text-sm">
            <button
              className={`px-3 py-1.5 rounded ${loginType === 'password' ? 'bg-teal-50 text-teal-700 font-medium' : 'text-gray-500'}`}
              onClick={() => setLoginType('password')}
            >密码</button>
            <button
              className={`px-3 py-1.5 rounded ${loginType === 'sms' ? 'bg-teal-50 text-teal-700 font-medium' : 'text-gray-500'}`}
              onClick={() => setLoginType('sms')}
            >验证码</button>
          </div>

          <Form layout="vertical" onFinish={handleSubmit}>
            <div className="mb-4"><Input placeholder="手机号/用户名" className="h-11 rounded" /></div>
            {loginType === 'password' && <div className="mb-4"><Input type="password" placeholder="密码" className="h-11 rounded" /></div>}
            {loginType === 'sms' && <div className="mb-4 flex gap-2"><Input placeholder="验证码" className="h-11 rounded flex-1" /><button className="px-3 h-11 bg-teal-50 text-teal-600 rounded text-sm">{smsCountdown > 0 ? `${smsCountdown}s` : '获取'}</button></div>}
            <Button block color="primary" size="large" loading={loading} type="submit" className="h-11 rounded" style={{ background: '#007185' }}>{loading ? '登录中...' : '登 录'}</Button>
          </Form>

          <div className="mt-4 flex justify-center"><SocialLoginButtons theme="stable" /></div>

          <div className="mt-4 flex justify-between text-sm">
            <button className="text-teal-600">忘记密码</button>
            <button onClick={() => navigate('/register')} className="text-teal-600">注册 →</button>
          </div>
        </div>
      </div>
    </div>
  )
}
