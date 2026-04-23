import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Form, Input, Toast } from 'antd-mobile'
import { login, smsLogin, sendSms } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'
import SocialLoginButtons from '@/components/SocialLoginButtons'

const isDesktop = () => {
  return window.innerWidth >= 768
}

export default function Login() {
  const navigate = useNavigate()
  const { login: setAuth } = useAuthStore()
  const [loading, setLoading] = useState(false)
  const [loginType, setLoginType] = useState<'password' | 'sms'>('password')
  const [smsCountdown, setSmsCountdown] = useState(0)
  const [smsLoading, setSmsLoading] = useState(false)
  const [mounted, setMounted] = useState(false)
  const [form] = Form.useForm()
  const [tabIndicatorStyle, setTabIndicatorStyle] = useState({ left: '0%', width: '50%' })
  const [isDesktopView, setIsDesktopView] = useState(false)

  useEffect(() => {
    setMounted(true)
    const checkDesktop = () => setIsDesktopView(isDesktop())
    checkDesktop()
    window.addEventListener('resize', checkDesktop)
    return () => window.removeEventListener('resize', checkDesktop)
  }, [])

  useEffect(() => {
    if (smsCountdown > 0) {
      const timer = setTimeout(() => setSmsCountdown(smsCountdown - 1), 1000)
      return () => clearTimeout(timer)
    }
  }, [smsCountdown])

  const handleTabClick = (type: 'password' | 'sms') => {
    setLoginType(type)
    setTabIndicatorStyle({
      left: type === 'password' ? '0%' : '50%',
      width: '50%',
    })
  }

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

  const handleSkip = () => {
    navigate('/')
  }

  // Desktop Layout
  if (isDesktopView) {
    return (
      <div
        className="min-h-screen flex items-center justify-center"
        style={{ background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' }}
      >
        <div
          className="bg-white rounded-2xl shadow-2xl flex overflow-hidden"
          style={{
            width: '900px',
            maxWidth: '95vw',
            opacity: mounted ? 1 : 0,
            transform: mounted ? 'translateY(0) scale(1)' : 'translateY(20px) scale(0.98)',
            transition: 'opacity 0.5s ease-out, transform 0.5s ease-out',
          }}
        >
          {/* Left side - branding */}
          <div
            className="flex-1 p-12 flex flex-col justify-center items-center"
            style={{ background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' }}
          >
            <div
              className="w-24 h-24 rounded-2xl flex items-center justify-center mb-6"
              style={{ background: 'rgba(255,255,255,0.2)', backdropFilter: 'blur(10px)' }}
            >
              <svg width="48" height="48" viewBox="0 0 24 24" fill="none">
                <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </div>
            <h1 className="text-3xl font-bold text-white mb-3">欢迎来到商城</h1>
            <p className="text-white/70 text-center mb-8">探索海量商品，享受购物乐趣</p>
            <div className="flex gap-6 text-white/80 text-sm">
              <div className="text-center">
                <div className="text-2xl font-bold text-white">1000+</div>
                <div>精选商品</div>
              </div>
              <div className="w-px bg-white/30" />
              <div className="text-center">
                <div className="text-2xl font-bold text-white">500+</div>
                <div>品牌商家</div>
              </div>
              <div className="w-px bg-white/30" />
              <div className="text-center">
                <div className="text-2xl font-bold text-white">10万+</div>
                <div>满意用户</div>
              </div>
            </div>
          </div>

          {/* Right side - login form */}
          <div className="flex-1 p-12 flex flex-col justify-center">
            <div className="mb-8">
              <h2 className="text-2xl font-bold text-gray-900 mb-2">欢迎回来</h2>
              <p className="text-gray-500">登录后享受更多权益</p>
            </div>

            {/* Tab Switcher */}
            <div className="relative flex mb-6 bg-gray-100 rounded-xl p-1">
              <div
                className="absolute top-0 bottom-0 rounded-xl bg-white shadow-sm transition-all duration-300 ease-out"
                style={{ left: tabIndicatorStyle.left, width: tabIndicatorStyle.width }}
              />
              <div
                className={`relative flex-1 py-2.5 text-center font-medium text-sm rounded-xl transition-colors duration-300 z-10 cursor-pointer ${
                  loginType === 'password' ? 'text-primary-500' : 'text-gray-500'
                }`}
                onClick={() => handleTabClick('password')}
              >
                密码登录
              </div>
              <div
                className={`relative flex-1 py-2.5 text-center font-medium text-sm rounded-xl transition-colors duration-300 z-10 cursor-pointer ${
                  loginType === 'sms' ? 'text-primary-500' : 'text-gray-500'
                }`}
                onClick={() => handleTabClick('sms')}
              >
                验证码登录
              </div>
            </div>

            {/* Form */}
            <div className="transition-all duration-300">
              {loginType === 'password' ? (
                <Form layout="vertical" onFinish={handleSubmit}>
                  <div className="mb-4">
                    <label className="text-sm text-gray-600 mb-1.5 block">账号</label>
                    <Input
                      placeholder="手机号/用户名"
                      className="h-12 rounded-xl border-gray-200 bg-gray-50 text-base"
                    />
                  </div>
                  <div className="mb-4">
                    <label className="text-sm text-gray-600 mb-1.5 block">密码</label>
                    <Input
                      type="password"
                      placeholder="请输入密码"
                      className="h-12 rounded-xl border-gray-200 bg-gray-50 text-base"
                    />
                  </div>
                  <div className="flex justify-between items-center mb-6">
                    <span className="text-sm text-primary-500 cursor-pointer" onClick={() => navigate('/register')}>
                      忘记密码？
                    </span>
                  </div>
                  <Button
                    block
                    color="primary"
                    size="large"
                    loading={loading}
                    type="submit"
                    className="h-12 rounded-xl text-base font-medium"
                    style={{ '--background-color': '#667eea' } as React.CSSProperties}
                  >
                    {loading ? '登录中...' : '登 录'}
                  </Button>
                </Form>
              ) : (
                <Form layout="vertical" onFinish={handleSubmit} form={form}>
                  <div className="mb-4">
                    <label className="text-sm text-gray-600 mb-1.5 block">手机号</label>
                    <Input
                      placeholder="请输入手机号"
                      className="h-12 rounded-xl border-gray-200 bg-gray-50 text-base"
                    />
                  </div>
                  <div className="mb-6">
                    <label className="text-sm text-gray-600 mb-1.5 block">验证码</label>
                    <div className="flex gap-3">
                      <Input
                        placeholder="请输入验证码"
                        className="h-12 rounded-xl border-gray-200 bg-gray-50 text-base flex-1"
                      />
                      <div
                        className={`h-12 px-4 rounded-xl flex items-center justify-center text-sm font-medium cursor-pointer transition-all ${
                          smsCountdown > 0
                            ? 'bg-gray-100 text-gray-400'
                            : 'bg-primary-100 text-primary-500 hover:bg-primary-200'
                        }`}
                        onClick={smsCountdown > 0 ? undefined : handleSendSms}
                      >
                        {smsCountdown > 0 ? `${smsCountdown}s` : smsLoading ? '发送中' : '获取验证码'}
                      </div>
                    </div>
                  </div>
                  <Button
                    block
                    color="primary"
                    size="large"
                    loading={loading}
                    type="submit"
                    className="h-12 rounded-xl text-base font-medium"
                    style={{ '--background-color': '#667eea' } as React.CSSProperties}
                  >
                    {loading ? '登录中...' : '登 录'}
                  </Button>
                </Form>
              )}
            </div>

            {/* Social Login */}
            <div className="mt-6">
              <div className="relative">
                <div className="absolute inset-0 flex items-center">
                  <div className="w-full border-t border-gray-200" />
                </div>
                <div className="relative flex justify-center text-sm">
                  <span className="px-4 bg-white text-gray-400">其他登录方式</span>
                </div>
              </div>
              <div className="mt-4 flex justify-center">
                <SocialLoginButtons layout="horizontal" />
              </div>
            </div>

            {/* Skip */}
            <div className="mt-6 text-center">
              <button
                onClick={handleSkip}
                className="text-gray-400 hover:text-gray-600 text-sm transition-colors"
              >
                跳过登录，先看看有什么
              </button>
            </div>

            {/* Register */}
            <div className="mt-4 text-center text-sm text-gray-500">
              还没有账号？
              <span className="text-primary-500 font-medium ml-1 cursor-pointer hover:underline" onClick={() => navigate('/register')}>
                立即注册
              </span>
            </div>
          </div>
        </div>
      </div>
    )
  }

  // Mobile Layout
  return (
    <div className="min-h-screen flex flex-col" style={{ background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' }}>
      {/* Skip button - top right */}
      <div className="absolute top-4 right-4 z-10">
        <button
          onClick={handleSkip}
          className="text-white/80 text-sm px-3 py-1.5 rounded-full bg-white/10 backdrop-blur-sm hover:bg-white/20 transition-colors"
        >
          跳过
        </button>
      </div>

      {/* Header */}
      <div
        className="pt-12 pb-6 px-6 text-center"
        style={{
          opacity: mounted ? 1 : 0,
          transform: mounted ? 'translateY(0)' : 'translateY(-20px)',
          transition: 'opacity 0.6s ease-out, transform 0.6s ease-out',
        }}
      >
        <div
          className="w-20 h-20 mx-auto mb-4 rounded-2xl flex items-center justify-center shadow-lg"
          style={{ background: 'rgba(255,255,255,0.2)', backdropFilter: 'blur(10px)' }}
        >
          <svg width="40" height="40" viewBox="0 0 24 24" fill="none">
            <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        </div>
        <h1 className="text-2xl font-bold text-white mb-2">欢迎回来</h1>
        <p className="text-white/70">登录后享受更多权益</p>
      </div>

      {/* Login Form Card */}
      <div
        className="flex-1 rounded-t-3xl bg-white px-6 py-8"
        style={{
          opacity: mounted ? 1 : 0,
          transform: mounted ? 'translateY(0)' : 'translateY(40px)',
          transition: 'opacity 0.6s ease-out 0.2s, transform 0.6s ease-out 0.2s',
        }}
      >
        {/* Tab Switcher */}
        <div className="relative flex mb-6 bg-gray-100 rounded-xl p-1">
          <div
            className="absolute top-0 bottom-0 rounded-xl bg-white shadow-sm transition-all duration-300 ease-out"
            style={{ left: tabIndicatorStyle.left, width: tabIndicatorStyle.width }}
          />
          <div
            className={`relative flex-1 py-2.5 text-center font-medium text-sm rounded-xl transition-colors duration-300 z-10 cursor-pointer ${
              loginType === 'password' ? 'text-primary-500' : 'text-gray-500'
            }`}
            onClick={() => handleTabClick('password')}
          >
            密码登录
          </div>
          <div
            className={`relative flex-1 py-2.5 text-center font-medium text-sm rounded-xl transition-colors duration-300 z-10 cursor-pointer ${
              loginType === 'sms' ? 'text-primary-500' : 'text-gray-500'
            }`}
            onClick={() => handleTabClick('sms')}
          >
            验证码登录
          </div>
        </div>

        {/* Form */}
        <div className="transition-all duration-300">
          {loginType === 'password' ? (
            <Form layout="vertical" onFinish={handleSubmit}>
              <div className="mb-4">
                <label className="text-sm text-gray-600 mb-1.5 block">账号</label>
                <div className="relative">
                  <Input
                    placeholder="手机号/用户名"
                    className="h-12 rounded-xl border-gray-200 bg-gray-50 pl-10 text-base"
                  />
                  <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
                      <circle cx="12" cy="7" r="4"/>
                    </svg>
                  </div>
                </div>
              </div>
              <div className="mb-4">
                <label className="text-sm text-gray-600 mb-1.5 block">密码</label>
                <div className="relative">
                  <Input
                    type="password"
                    placeholder="请输入密码"
                    className="h-12 rounded-xl border-gray-200 bg-gray-50 pl-10 text-base"
                  />
                  <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
                      <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
                    </svg>
                  </div>
                </div>
              </div>
              <div className="flex justify-end mb-6">
                <span className="text-sm text-primary-500 cursor-pointer" onClick={() => navigate('/register')}>
                  忘记密码？
                </span>
              </div>
              <Button
                block
                color="primary"
                size="large"
                loading={loading}
                type="submit"
                className="h-12 rounded-xl text-base font-medium shadow-lg shadow-primary-500/30 hover:shadow-xl hover:shadow-primary-500/40 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0"
                style={{ '--background-color': '#667eea' } as React.CSSProperties}
              >
                {loading ? '登录中...' : '登 录'}
              </Button>
            </Form>
          ) : (
            <Form layout="vertical" onFinish={handleSubmit} form={form}>
              <div className="mb-4">
                <label className="text-sm text-gray-600 mb-1.5 block">手机号</label>
                <div className="relative">
                  <Input
                    placeholder="请输入手机号"
                    className="h-12 rounded-xl border-gray-200 bg-gray-50 pl-10 text-base"
                  />
                  <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"/>
                    </svg>
                  </div>
                </div>
              </div>
              <div className="mb-6">
                <label className="text-sm text-gray-600 mb-1.5 block">验证码</label>
                <div className="relative">
                  <Input
                    placeholder="请输入验证码"
                    className="h-12 rounded-xl border-gray-200 bg-gray-50 pl-10 text-base"
                  />
                  <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
                    </svg>
                  </div>
                  <div
                    className={`absolute right-3 top-1/2 -translate-y-1/2 text-sm font-medium px-3 py-1 rounded-full transition-all cursor-pointer ${
                      smsCountdown > 0
                        ? 'bg-gray-100 text-gray-400'
                        : 'bg-primary-100 text-primary-500 hover:bg-primary-200'
                    }`}
                    onClick={smsCountdown > 0 ? undefined : handleSendSms}
                  >
                    {smsCountdown > 0 ? `${smsCountdown}s` : smsLoading ? '发送中' : '获取验证码'}
                  </div>
                </div>
              </div>
              <Button
                block
                color="primary"
                size="large"
                loading={loading}
                type="submit"
                className="h-12 rounded-xl text-base font-medium shadow-lg shadow-primary-500/30 hover:shadow-xl hover:shadow-primary-500/40 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0"
                style={{ '--background-color': '#667eea' } as React.CSSProperties}
              >
                {loading ? '登录中...' : '登 录'}
              </Button>
            </Form>
          )}
        </div>

        {/* Social Login */}
        <div className="mt-8">
          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-gray-200" />
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="px-4 bg-white text-gray-400">其他登录方式</span>
            </div>
          </div>
          <div className="mt-4">
            <SocialLoginButtons />
          </div>
        </div>

        {/* Agreement */}
        <div className="mt-6 text-center text-xs text-gray-400">
          登录即表示同意
          <span className="text-primary-500">《用户协议》</span>和
          <span className="text-primary-500">《隐私政策》</span>
        </div>

        {/* Register Link */}
        <div className="mt-4 text-center text-sm text-gray-500">
          还没有账号？
          <span className="text-primary-500 font-medium ml-1 cursor-pointer" onClick={() => navigate('/register')}>
            立即注册
          </span>
        </div>

        {/* Skip Link */}
        <div className="mt-4 text-center">
          <button
            onClick={handleSkip}
            className="text-gray-400 text-sm hover:text-gray-600 transition-colors"
          >
            跳过登录，先看看有什么
          </button>
        </div>
      </div>
    </div>
  )
}
