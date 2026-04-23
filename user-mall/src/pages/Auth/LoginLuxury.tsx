// user-mall/src/pages/Auth/LoginLuxury.tsx
import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Form, Toast } from 'antd-mobile'
import { login, smsLogin, sendSms } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'
import SocialLoginButtons from '@/components/SocialLoginButtons'

const isDesktop = () => window.innerWidth >= 768

export default function LoginLuxury() {
  const navigate = useNavigate()
  const { login: setAuth } = useAuthStore()
  const [loading, setLoading] = useState(false)
  const [loginType, setLoginType] = useState<'password' | 'sms'>('password')
  const [smsCountdown, setSmsCountdown] = useState(0)
  const [form] = Form.useForm()
  const [isDesktopView, setIsDesktopView] = useState(false)
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
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
      await sendSms({ phone, type: 'login' })
      Toast.show('验证码已发送')
      setSmsCountdown(60)
    } catch {
      Toast.show('发送失败')
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

  // Desktop Layout
  if (isDesktopView) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-stone-100">
        <div
          className="w-full max-w-4xl flex shadow-2xl rounded-sm overflow-hidden"
          style={{
            opacity: mounted ? 1 : 0,
            transform: mounted ? 'translateY(0)' : 'translateY(20px)',
            transition: 'opacity 0.8s ease, transform 0.8s ease',
          }}
        >
          {/* Left - Brand */}
          <div className="flex-1 bg-stone-900 p-16 flex flex-col justify-center items-center text-white">
            <div className="w-16 h-px bg-stone-600 mb-8" />
            <h1 className="text-4xl font-light tracking-wider mb-4">LUXURY</h1>
            <p className="text-stone-400 tracking-widest text-sm mb-8">精 · 致 · 生 · 活</p>
            <div className="w-16 h-px bg-stone-600 mt-8" />
          </div>

          {/* Right - Form */}
          <div className="w-[400px] bg-white p-16 flex flex-col justify-center">
            <h2 className="text-2xl font-light text-stone-800 mb-2">欢迎回来</h2>
            <p className="text-stone-400 text-sm mb-8">探索专属精致体验</p>

            <div className="flex gap-6 mb-8">
              <button
                className={`text-sm pb-2 ${loginType === 'password' ? 'text-stone-800 border-b border-stone-800' : 'text-stone-400'}`}
                onClick={() => setLoginType('password')}
              >密码登录</button>
              <button
                className={`text-sm pb-2 ${loginType === 'sms' ? 'text-stone-800 border-b border-stone-800' : 'text-stone-400'}`}
                onClick={() => setLoginType('sms')}
              >验证码</button>
            </div>

            <Form layout="vertical" onFinish={handleSubmit}>
              <div className="space-y-4">
                <div className="border-b border-stone-200 pb-2">
                  <input
                    type="text"
                    placeholder="邮箱 / 手机号"
                    className="w-full text-stone-800 placeholder-stone-300 outline-none bg-transparent text-sm"
                  />
                </div>
                {loginType === 'password' && (
                  <div className="border-b border-stone-200 pb-2">
                    <input
                      type="password"
                      placeholder="密码"
                      className="w-full text-stone-800 placeholder-stone-300 outline-none bg-transparent text-sm"
                    />
                  </div>
                )}
                {loginType === 'sms' && (
                  <div className="flex border-b border-stone-200 pb-2">
                    <input
                      type="text"
                      placeholder="验证码"
                      className="flex-1 text-stone-800 placeholder-stone-300 outline-none bg-transparent text-sm"
                    />
                    <button
                      className={`text-xs ${smsCountdown > 0 ? 'text-stone-300' : 'text-stone-500'}`}
                      onClick={smsCountdown > 0 ? undefined : handleSendSms}
                    >
                      {smsCountdown > 0 ? `${smsCountdown}s` : '获取验证码'}
                    </button>
                  </div>
                )}
              </div>

              <Button
                block
                type="submit"
                loading={loading}
                className="mt-8 h-12 rounded-sm bg-stone-800 text-white text-sm tracking-widest"
                style={{ background: '#1A1A1A' }}
              >
                {loading ? '登录中...' : '登 录'}
              </Button>
            </Form>

            <div className="mt-8 flex justify-center">
              <SocialLoginButtons theme="luxury" layout="horizontal" />
            </div>

            <div className="mt-8 text-center">
              <button onClick={() => navigate('/register')} className="text-stone-400 text-sm hover:text-stone-600 transition-colors">
                还没有账号？<span className="border-b border-stone-300">立即注册</span>
              </button>
            </div>

            <div className="mt-8 text-center">
              <button onClick={() => navigate('/')} className="text-stone-300 text-xs hover:text-stone-500 transition-colors">
                跳过登录 →
              </button>
            </div>
          </div>
        </div>
      </div>
    )
  }

  // Mobile Layout
  return (
    <div className="min-h-screen bg-stone-100 flex flex-col items-center justify-center p-6">
      <div
        className="w-full max-w-sm"
        style={{
          opacity: mounted ? 1 : 0,
          transform: mounted ? 'translateY(0)' : 'translateY(20px)',
          transition: 'opacity 0.6s ease, transform 0.6s ease',
        }}
      >
        <div className="text-center mb-12">
          <div className="w-12 h-px bg-stone-300 mx-auto mb-4" />
          <h1 className="text-2xl font-light tracking-widest text-stone-800">LUXURY</h1>
          <p className="text-stone-400 text-xs tracking-widest mt-1">精 · 致 · 生 · 活</p>
        </div>

        <div className="bg-white rounded-sm shadow-lg p-8">
          <div className="flex gap-4 mb-6">
            <button
              className={`text-sm ${loginType === 'password' ? 'text-stone-800 border-b border-stone-800' : 'text-stone-400'}`}
              onClick={() => setLoginType('password')}
            >密码</button>
            <button
              className={`text-sm ${loginType === 'sms' ? 'text-stone-800 border-b border-stone-800' : 'text-stone-400'}`}
              onClick={() => setLoginType('sms')}
            >验证码</button>
          </div>

          <Form layout="vertical" onFinish={handleSubmit}>
            <div className="space-y-3 mb-4">
              <input
                type="text"
                placeholder="邮箱 / 手机号"
                className="w-full text-stone-800 placeholder-stone-300 outline-none border-b border-stone-200 pb-2 text-sm"
              />
              {loginType === 'password' && (
                <input
                  type="password"
                  placeholder="密码"
                  className="w-full text-stone-800 placeholder-stone-300 outline-none border-b border-stone-200 pb-2 text-sm"
                />
              )}
              {loginType === 'sms' && (
                <div className="flex items-center border-b border-stone-200 pb-2">
                  <input
                    type="text"
                    placeholder="验证码"
                    className="flex-1 text-stone-800 placeholder-stone-300 outline-none text-sm"
                  />
                  <button className="text-xs text-stone-400">{smsCountdown > 0 ? `${smsCountdown}s` : '获取'}</button>
                </div>
              )}
            </div>

            <Button
              block
              type="submit"
              loading={loading}
              className="h-11 rounded-sm bg-stone-800 text-white text-sm tracking-wider"
              style={{ background: '#1A1A1A' }}
            >
              {loading ? '登录中...' : '登 录'}
            </Button>
          </Form>

          <div className="mt-6 flex justify-center">
            <SocialLoginButtons theme="luxury" />
          </div>

          <div className="mt-6 text-center">
            <button onClick={() => navigate('/register')} className="text-stone-400 text-xs">
              还没有账号？<span className="border-b border-stone-300">立即注册</span>
            </button>
          </div>
        </div>

        <div className="mt-8 text-center">
          <button onClick={() => navigate('/')} className="text-stone-400 text-xs hover:text-stone-600">
            跳过登录 →
          </button>
        </div>
      </div>
    </div>
  )
}