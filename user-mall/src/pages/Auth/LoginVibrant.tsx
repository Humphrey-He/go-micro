// user-mall/src/pages/Auth/LoginVibrant.tsx
import { useState, useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Form, Input, Toast } from 'antd-mobile'
import { login, smsLogin, sendSms } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'
import SocialLoginButtons from '@/components/SocialLoginButtons'

const isDesktop = () => window.innerWidth >= 768

// Mak's Store Logo Component
const MakLogo = ({ size = 'large' }: { size?: 'large' | 'small' }) => {
  const sizeClasses = size === 'large' ? 'w-20 h-20 text-4xl' : 'w-10 h-10 text-xl'
  return (
    <div className={`${sizeClasses} relative flex items-center justify-center`}>
      <svg viewBox="0 0 100 100" className="w-full h-full">
        {/* Outer ring with gradient */}
        <defs>
          <linearGradient id="logoGradient" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stopColor="#FF6B35" />
            <stop offset="50%" stopColor="#FF4D94" />
            <stop offset="100%" stopColor="#FF6B35" />
          </linearGradient>
          <filter id="glow">
            <feGaussianBlur stdDeviation="2" result="coloredBlur"/>
            <feMerge>
              <feMergeNode in="coloredBlur"/>
              <feMergeNode in="SourceGraphic"/>
            </feMerge>
          </filter>
        </defs>
        {/* Background circle */}
        <circle cx="50" cy="50" r="45" fill="url(#logoGradient)" filter="url(#glow)" />
        {/* M letter */}
        <text x="50" y="62" textAnchor="middle" fill="white" fontSize="40" fontWeight="bold" fontFamily="Arial, sans-serif">M</text>
        {/* Decorative dots */}
        <circle cx="20" cy="50" r="3" fill="#F7C948" />
        <circle cx="80" cy="50" r="3" fill="#00D9C0" />
      </svg>
    </div>
  )
}

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

// Click burst effect component
const ClickBurst = ({ x, y, active }: { x: number; y: number; active: boolean }) => {
  if (!active) return null

  const particles = Array.from({ length: 12 }, (_, i) => {
    const angle = (i / 12) * Math.PI * 2
    const distance = 60 + Math.random() * 40
    const endX = Math.cos(angle) * distance
    const endY = Math.sin(angle) * distance
    const color = ['#FF6B35', '#FF4D94', '#F7C948', '#00D9C0'][i % 4]
    const size = 6 + Math.random() * 6

    return (
      <div
        key={i}
        className="absolute rounded-full animate-burst-particle"
        style={{
          width: size,
          height: size,
          background: color,
          '--end-x': `${endX}px`,
          '--end-y': `${endY}px`,
        } as React.CSSProperties}
      />
    )
  })

  return (
    <div
      className="pointer-events-none fixed z-[9999]"
      style={{
        left: x,
        top: y,
        transform: 'translate(-50%, -50%)',
      }}
    >
      {particles}
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
  const [clickEffect, setClickEffect] = useState({ active: false, x: 0, y: 0 })
  const [isHovered, setIsHovered] = useState(false)
  const btnRef = useRef<HTMLDivElement>(null)

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

  const handleLoginClick = (e: React.MouseEvent) => {
    // Trigger burst effect at click position
    setClickEffect({ active: true, x: e.clientX, y: e.clientY })
    setTimeout(() => setClickEffect(prev => ({ ...prev, active: false })), 600)
  }

  const handleSubmit = async (values: any) => {
    try {
      setLoading(true)
      if (loginType === 'password') {
        const res = await login({ account: values.account || '', password: values.password || '' })
        setAuth(res.token, res.user)
        Toast.show('🎉 登录成功，欢迎回来！')
        navigate('/')
      } else {
        const res = await smsLogin({ phone: values.phone || '', code: values.code || '' })
        setAuth(res.token, res.user)
        Toast.show('🎉 登录成功，欢迎回来！')
        navigate('/')
      }
    } catch {
      Toast.show('登录失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  // Desktop Layout
  if (isDesktopView) {
    return (
      <div className="min-h-screen flex relative overflow-hidden" style={{ background: 'linear-gradient(135deg, #FF6B35 0%, #FF4D94 50%, #FF6B35 100%)', backgroundSize: '200% 200%' }}>
        <ClickBurst x={clickEffect.x} y={clickEffect.y} active={clickEffect.active} />
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
            {/* Logo and Brand Name */}
            <div className="flex items-center justify-center gap-4 mb-6">
              <MakLogo size="large" />
              <div className="text-left">
                <h1 className="text-5xl font-black tracking-tight">Mak's Store</h1>
                <p className="text-lg text-white/80 font-light">探索品质生活</p>
              </div>
            </div>

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

            {/* Breathing gradient login button */}
            <div
              ref={btnRef}
              className="relative"
              style={{ cursor: isHovered ? 'pointer' : 'default' }}
              onMouseEnter={() => setIsHovered(true)}
              onMouseLeave={() => setIsHovered(false)}
            >
              {/* Breathing glow effect */}
              <div
                className="absolute inset-0 rounded-2xl animate-breathing-glow"
                style={{
                  background: 'linear-gradient(135deg, #FF6B35 0%, #FF4D94 50%, #FF6B35 100%)',
                  filter: 'blur(15px)',
                  opacity: isHovered ? 0.6 : 0.3,
                  transform: 'scale(1.05)',
                }}
              />

              <Button
                block
                type="submit"
                loading={loading}
                onClick={handleLoginClick}
                className="relative h-14 rounded-2xl text-base font-bold shadow-md hover:shadow-xl transition-all duration-300 animate-gradient-shift"
                style={{
                  background: 'linear-gradient(135deg, #FF6B35 0%, #FF4D94 100%)',
                  border: 'none',
                  backgroundSize: '200% 200%',
                }}
              >
                {loading ? (
                  <span className="text-white/90">登录中...</span>
                ) : (
                  <span className="text-white font-bold relative z-10">✨ 立即登录 ✨</span>
                )}
              </Button>
            </div>
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
      <ClickBurst x={clickEffect.x} y={clickEffect.y} active={clickEffect.active} />
      <AnimatedGradientBg />

      <div className="p-4 flex justify-between items-center relative z-10">
        <div className="flex items-center gap-2 text-white">
          <MakLogo size="small" />
          <span className="font-bold text-lg">Mak's Store</span>
        </div>
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
          <div className="flex items-center justify-center gap-3 mb-4">
            <MakLogo size="large" />
          </div>
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

          {/* Breathing gradient login button for mobile */}
          <div className="relative">
            <div
              className="absolute inset-0 rounded-xl animate-breathing-glow"
              style={{
                background: 'linear-gradient(135deg, #FF6B35 0%, #FF4D94 100%)',
                filter: 'blur(10px)',
                opacity: 0.4,
                transform: 'scale(1.02)',
              }}
            />
            <Button
              block
              type="submit"
              loading={loading}
              onClick={handleLoginClick}
              className="relative h-12 rounded-xl font-bold text-base"
              style={{
                background: 'linear-gradient(135deg, #FF6B35 0%, #FF4D94 100%)',
                border: 'none',
              }}
            >
              <span className="text-white">✨ 立即登录 ✨</span>
            </Button>
          </div>
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
