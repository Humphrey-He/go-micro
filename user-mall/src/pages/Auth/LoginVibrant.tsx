// user-mall/src/pages/Auth/LoginVibrant.tsx
import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Form, Input, Toast } from 'antd-mobile'
import { login, smsLogin, sendSms } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'
import SocialLoginButtons from '@/components/SocialLoginButtons'

const isDesktop = () => window.innerWidth >= 768

// Floating particle component
const FloatingParticle = ({ delay, size, x, y, color }: { delay: number; size: number; x: number; y: number; color: string }) => (
  <div
    className="absolute rounded-full opacity-60 animate-float"
    style={{
      width: size,
      height: size,
      left: `${x}%`,
      top: `${y}%`,
      background: color,
      animationDelay: `${delay}s`,
      animationDuration: `${3 + Math.random() * 2}s`,
    }}
  />
)

// Animated gradient background
const AnimatedGradientBg = () => {
  const colors = [
    'rgba(255, 107, 53, 0.4)',
    'rgba(247, 201, 72, 0.3)',
    'rgba(0, 217, 192, 0.3)',
    'rgba(255, 77, 148, 0.3)',
  ]
  return (
    <div className="absolute inset-0 overflow-hidden">
      {colors.map((color, i) => (
        <div
          key={i}
          className="absolute rounded-full blur-3xl animate-pulse-slow"
          style={{
            width: 300 + i * 100,
            height: 300 + i * 100,
            background: color,
            left: `${(i * 25) % 100}%`,
            top: `${(i * 30) % 100}%`,
            animationDelay: `${i * 0.5}s`,
            animationDuration: `${4 + i}s`,
          }}
        />
      ))}
    </div>
  )
}

export default function LoginVibrant() {
  const navigate = useNavigate()
  const { login: setAuth } = useAuthStore()
  const [loading, setLoading] = useState(false)
  const [loginType, setLoginType] = useState<'password' | 'sms'>('password')
  const [smsCountdown, setSmsCountdown] = useState(0)
  const [form] = Form.useForm()
  const [tabStyle, setTabStyle] = useState({ left: '0%', width: '50%' })
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

  const handleTab = (type: 'password' | 'sms') => {
    setLoginType(type)
    setTabStyle({ left: type === 'password' ? '0%' : '50%', width: '50%' })
  }

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
      Toast.show('登录失败')
    } finally {
      setLoading(false)
    }
  }

  // Desktop Layout
  if (isDesktopView) {
    return (
      <div className="min-h-screen flex relative overflow-hidden" style={{ background: 'linear-gradient(135deg, #FF6B35 0%, #FF4D94 50%, #FF6B35 100%)', backgroundSize: '200% 200%' }}>
        <AnimatedGradientBg />

        {/* Floating particles */}
        <FloatingParticle delay={0} size={20} x={10} y={20} color="rgba(255,255,255,0.3)" />
        <FloatingParticle delay={0.5} size={30} x={80} y={15} color="rgba(255,255,255,0.2)" />
        <FloatingParticle delay={1} size={15} x={60} y={70} color="rgba(255,255,255,0.25)" />
        <FloatingParticle delay={1.5} size={25} x={25} y={75} color="rgba(255,255,255,0.2)" />
        <FloatingParticle delay={2} size={18} x={90} y={50} color="rgba(255,255,255,0.15)" />

        {/* Left side - Brand */}
        <div
          className="flex-1 p-12 flex flex-col justify-center items-center relative z-10"
          style={{
            opacity: mounted ? 1 : 0,
            transform: mounted ? 'translateX(0)' : 'translateX(-30px)',
            transition: 'opacity 0.8s ease, transform 0.8s ease',
          }}
        >
          <div className="text-center text-white">
            <h1 className="text-6xl font-black mb-4 tracking-tight">
              <span className="animate-bounce inline-block" style={{ animationDuration: '2s' }}>🎉</span> 兴趣电商
            </h1>
            <p className="text-2xl text-white/90 mb-10 font-light">发现你的热爱，好物即刻享</p>
            <div className="flex gap-12 justify-center text-white/95">
              <div className="text-center">
                <div className="text-4xl font-black mb-1">1000+</div>
                <div className="text-sm opacity-80">精选好物</div>
              </div>
              <div className="text-center">
                <div className="text-4xl font-black mb-1">500+</div>
                <div className="text-sm opacity-80">品牌商家</div>
              </div>
              <div className="text-center">
                <div className="text-4xl font-black mb-1">10万+</div>
                <div className="text-sm opacity-80">满意用户</div>
              </div>
            </div>
          </div>

          {/* Promo badge */}
          <div className="absolute bottom-20 left-12 bg-white/20 backdrop-blur-md rounded-2xl px-6 py-3 text-white">
            <div className="text-sm font-medium">新用户专享</div>
            <div className="text-lg font-bold">首单立减 ¥10</div>
          </div>
        </div>

        {/* Right side - Form */}
        <div
          className="w-[520px] bg-white rounded-l-[40px] shadow-2xl p-12 flex flex-col justify-center relative z-10"
          style={{
            boxShadow: '-20px 0 80px rgba(255, 77, 148, 0.3)',
            opacity: mounted ? 1 : 0,
            transform: mounted ? 'translateX(0)' : 'translateX(30px)',
            transition: 'opacity 0.8s ease 0.2s, transform 0.8s ease 0.2s',
          }}
        >
          <div className="mb-2">
            <span className="inline-block px-3 py-1 bg-gradient-to-r from-orange-500 to-pink-500 text-white text-xs font-medium rounded-full mb-3">欢迎回来</span>
          </div>
          <h2 className="text-3xl font-bold text-gray-800 mb-1">登录即可领取</h2>
          <p className="text-gray-500 mb-8">专属优惠券等你拿</p>

          {/* Tab */}
          <div className="relative flex mb-8 bg-gray-100 rounded-2xl p-1.5">
            <div
              className="absolute top-1.5 bottom-1.5 rounded-xl bg-white shadow-md transition-all duration-300"
              style={{ left: tabStyle.left, width: tabStyle.width }}
            />
            <div
              className={`relative flex-1 py-3 text-center font-bold text-sm rounded-xl cursor-pointer z-10 transition-colors ${loginType === 'password' ? 'text-orange-500' : 'text-gray-500 hover:text-gray-600'}`}
              onClick={() => handleTab('password')}
            >
              密码登录
            </div>
            <div
              className={`relative flex-1 py-3 text-center font-bold text-sm rounded-xl cursor-pointer z-10 transition-colors ${loginType === 'sms' ? 'text-orange-500' : 'text-gray-500 hover:text-gray-600'}`}
              onClick={() => handleTab('sms')}
            >
              验证码登录
            </div>
          </div>

          {/* Form */}
          <Form layout="vertical" onFinish={handleSubmit}>
            {loginType === 'password' ? (
              <>
                <div className="mb-5">
                  <Input
                    placeholder="手机号/用户名"
                    className="h-14 rounded-2xl bg-gray-50 text-base border-0"
                  />
                </div>
                <div className="mb-6">
                  <Input
                    type="password"
                    placeholder="请输入密码"
                    className="h-14 rounded-2xl bg-gray-50 text-base border-0"
                  />
                </div>
              </>
            ) : (
              <>
                <div className="mb-5">
                  <Input
                    placeholder="请输入手机号"
                    className="h-14 rounded-2xl bg-gray-50 text-base border-0"
                  />
                </div>
                <div className="mb-6 flex gap-3">
                  <Input
                    placeholder="验证码"
                    className="h-14 rounded-2xl bg-gray-50 text-base flex-1 border-0"
                  />
                  <button
                    type="button"
                    className={`h-14 px-5 rounded-2xl text-sm font-bold transition-all duration-300 ${
                      smsCountdown > 0
                        ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                        : 'bg-gradient-to-r from-orange-500 to-pink-500 text-white hover:shadow-lg hover:scale-105 active:scale-95'
                    }`}
                    onClick={smsCountdown > 0 ? undefined : handleSendSms}
                    disabled={smsCountdown > 0}
                  >
                    {smsCountdown > 0 ? `${smsCountdown}s` : '获取验证码'}
                  </button>
                </div>
              </>
            )}

            <Button
              block
              type="submit"
              loading={loading}
              className="h-14 rounded-2xl text-base font-bold shadow-md hover:shadow-xl transition-shadow"
              style={{
                background: 'linear-gradient(135deg, #FF6B35 0%, #FF4D94 100%)',
                border: 'none',
              }}
            >
              {loading ? (
                <span className="text-white/90">登录中...</span>
              ) : (
                <span className="text-white font-bold">立即登录</span>
              )}
            </Button>
          </Form>

          <div className="mt-8">
            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-gray-200" />
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-4 bg-white text-gray-400">其他登录方式</span>
              </div>
            </div>
            <div className="mt-5 flex justify-center">
              <SocialLoginButtons theme="vibrant" layout="horizontal" />
            </div>
          </div>

          <div className="mt-8 text-center space-y-3">
            <button
              onClick={() => navigate('/')}
              className="block w-full text-gray-400 text-sm hover:text-orange-500 transition-colors py-2"
            >
              跳过登录，先看看 →
            </button>
            <div className="text-xs text-gray-400">
              还没有账号？<button onClick={() => navigate('/register')} className="text-orange-500 font-medium hover:underline">立即注册</button>
            </div>
          </div>
        </div>
      </div>
    )
  }

  // Mobile Layout
  return (
    <div className="min-h-screen flex flex-col relative overflow-hidden" style={{ background: 'linear-gradient(135deg, #FF6B35 0%, #FF4D94 100%)' }}>
      <AnimatedGradientBg />

      <div className="p-4 flex justify-between items-center relative z-10">
        <div className="text-white font-bold">🎉 兴趣电商</div>
        <button onClick={() => navigate('/')} className="text-white/80 text-sm px-4 py-2 rounded-full bg-white/20 backdrop-blur-sm">
          跳过
        </button>
      </div>

      <div
        className="flex-1 flex flex-col items-center justify-center px-6 relative z-10"
        style={{
          opacity: mounted ? 1 : 0,
          transform: mounted ? 'translateY(0)' : 'translateY(20px)',
          transition: 'opacity 0.6s ease, transform 0.6s ease',
        }}
      >
        <div className="text-center text-white mb-8">
          <h1 className="text-4xl font-black mb-2">🎉 兴趣电商</h1>
          <p className="text-white/80">发现你的热爱</p>
        </div>

        {/* Promo badge */}
        <div className="bg-white/20 backdrop-blur-md rounded-2xl px-6 py-3 text-white mb-6">
          <div className="text-sm font-medium">新用户专享</div>
          <div className="text-lg font-bold">首单立减 ¥10</div>
        </div>
      </div>

      <div
        className="bg-white rounded-t-[32px] p-6 relative z-10"
        style={{
          opacity: mounted ? 1 : 0,
          transform: mounted ? 'translateY(0)' : 'translateY(30px)',
          transition: 'opacity 0.6s ease 0.2s, transform 0.6s ease 0.2s',
        }}
      >
        <div className="relative flex mb-6 bg-gray-100 rounded-2xl p-1.5">
          <div
            className="absolute top-1.5 bottom-1.5 rounded-xl bg-white shadow transition-all duration-300"
            style={{ left: tabStyle.left, width: tabStyle.width }}
          />
          <div
            className={`relative flex-1 py-2.5 text-center font-bold text-sm rounded-xl cursor-pointer z-10 ${loginType === 'password' ? 'text-orange-500' : 'text-gray-500'}`}
            onClick={() => handleTab('password')}
          >
            密码登录
          </div>
          <div
            className={`relative flex-1 py-2.5 text-center font-bold text-sm rounded-xl cursor-pointer z-10 ${loginType === 'sms' ? 'text-orange-500' : 'text-gray-500'}`}
            onClick={() => handleTab('sms')}
          >
            验证码
          </div>
        </div>

        <Form layout="vertical" onFinish={handleSubmit}>
          <div className="mb-4">
            <Input
              placeholder="手机号/用户名"
              className="h-12 rounded-xl bg-gray-50 text-base"
            />
          </div>
          {loginType === 'password' ? (
            <div className="mb-4">
              <Input
                type="password"
                placeholder="密码"
                className="h-12 rounded-xl bg-gray-50 text-base"
              />
            </div>
          ) : (
            <div className="mb-4 flex gap-2">
              <Input
                placeholder="验证码"
                className="h-12 rounded-xl bg-gray-50 text-base flex-1"
              />
              <button
                type="button"
                className={`h-12 px-4 rounded-xl text-sm font-bold ${
                  smsCountdown > 0
                    ? 'bg-gray-100 text-gray-400'
                    : 'bg-gradient-to-r from-orange-500 to-pink-500 text-white'
                }`}
                onClick={smsCountdown > 0 ? undefined : handleSendSms}
              >
                {smsCountdown > 0 ? `${smsCountdown}s` : '获取'}
              </button>
            </div>
          )}

          <Button
            block
            type="submit"
            loading={loading}
            className="h-12 rounded-xl font-bold text-base"
            style={{
              background: 'linear-gradient(135deg, #FF6B35 0%, #FF4D94 100%)',
              border: 'none',
            }}
          >
            <span className="text-white">立即登录</span>
          </Button>
        </Form>

        <div className="mt-6">
          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-gray-200" />
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="px-4 bg-white text-gray-400">其他方式</span>
            </div>
          </div>
        </div>

        <div className="mt-4 flex justify-center">
          <SocialLoginButtons theme="vibrant" />
        </div>

        <div className="mt-6 text-center text-sm text-gray-500">
          还没有账号？<button onClick={() => navigate('/register')} className="text-orange-500 font-bold">立即注册</button>
        </div>
      </div>
    </div>
  )
}
