// user-mall/src/pages/Auth/LoginVibrant.tsx
import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Form, Input, Toast } from 'antd-mobile'
import { login, smsLogin, sendSms } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'
import SocialLoginButtons from '@/components/SocialLoginButtons'

const isDesktop = () => window.innerWidth >= 768

export default function LoginVibrant() {
  const navigate = useNavigate()
  const { login: setAuth } = useAuthStore()
  const [loading, setLoading] = useState(false)
  const [loginType, setLoginType] = useState<'password' | 'sms'>('password')
  const [smsCountdown, setSmsCountdown] = useState(0)
  const [smsLoading, setSmsLoading] = useState(false)
  const [mounted, setMounted] = useState(false)
  const [form] = Form.useForm()
  const [tabStyle, setTabStyle] = useState({ left: '0%', width: '50%' })
  const [isDesktopView, setIsDesktopView] = useState(false)

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
      Toast.show('登录失败')
    } finally {
      setLoading(false)
    }
  }

  // Desktop Layout
  if (isDesktopView) {
    return (
      <div className="min-h-screen flex" style={{ background: 'linear-gradient(135deg, #FF6B35 0%, #FF4D94 100%)' }}>
        {/* Left side - Brand */}
        <div className="flex-1 p-12 flex flex-col justify-center items-center">
          <div className="text-center text-white">
            <h1 className="text-5xl font-bold mb-4">🎉 兴趣电商</h1>
            <p className="text-xl text-white/80 mb-8">发现你的热爱，好物即刻享</p>
            <div className="flex gap-8 justify-center text-white/90">
              <div><span className="text-3xl font-bold">1000+</span><br/><span className="text-sm">精选好物</span></div>
              <div><span className="text-3xl font-bold">500+</span><br/><span className="text-sm">品牌商家</span></div>
              <div><span className="text-3xl font-bold">10万+</span><br/><span className="text-sm">满意用户</span></div>
            </div>
          </div>
          {/* Floating elements */}
          <div className="absolute top-20 left-20 w-20 h-20 rounded-full bg-yellow-400/30 animate-bounce" />
          <div className="absolute bottom-40 right-32 w-16 h-16 rounded-full bg-cyan-400/30 animate-pulse" />
          <div className="absolute top-40 right-40 w-12 h-12 rounded-full bg-pink-400/30 animate-ping" />
        </div>

        {/* Right side - Form */}
        <div className="w-[500px] bg-white rounded-l-3xl shadow-2xl p-10 flex flex-col justify-center" style={{ boxShadow: '-20px 0 60px rgba(0,0,0,0.2)' }}>
          <h2 className="text-3xl font-bold text-gray-800 mb-2">欢迎回来</h2>
          <p className="text-gray-500 mb-6">登录即享更多优惠</p>

          {/* Tab */}
          <div className="relative flex mb-6 bg-gray-100 rounded-xl p-1">
            <div className="absolute top-0 bottom-0 rounded-xl bg-white shadow transition-all duration-300" style={{ left: tabStyle.left, width: tabStyle.width }} />
            <div className={`relative flex-1 py-2.5 text-center font-medium text-sm rounded-xl cursor-pointer z-10 ${loginType === 'password' ? 'text-orange-500' : 'text-gray-500'}`} onClick={() => handleTab('password')}>密码登录</div>
            <div className={`relative flex-1 py-2.5 text-center font-medium text-sm rounded-xl cursor-pointer z-10 ${loginType === 'sms' ? 'text-orange-500' : 'text-gray-500'}`} onClick={() => handleTab('sms')}>验证码登录</div>
          </div>

          {/* Form */}
          <Form layout="vertical" onFinish={handleSubmit}>
            {loginType === 'password' ? (
              <>
                <div className="mb-4">
                  <Input placeholder="手机号/用户名" className="h-12 rounded-xl bg-gray-50 text-base" />
                </div>
                <div className="mb-4">
                  <Input type="password" placeholder="请输入密码" className="h-12 rounded-xl bg-gray-50 text-base" />
                </div>
              </>
            ) : (
              <>
                <div className="mb-4">
                  <Input placeholder="请输入手机号" className="h-12 rounded-xl bg-gray-50 text-base" />
                </div>
                <div className="mb-4 flex gap-3">
                  <Input placeholder="验证码" className="h-12 rounded-xl bg-gray-50 text-base flex-1" />
                  <div className={`h-12 px-4 rounded-xl flex items-center text-sm font-medium cursor-pointer ${smsCountdown > 0 ? 'bg-gray-100 text-gray-400' : 'bg-orange-100 text-orange-500'}`} onClick={smsCountdown > 0 ? undefined : handleSendSms}>
                    {smsCountdown > 0 ? `${smsCountdown}s` : '获取验证码'}
                  </div>
                </div>
              </>
            )}
            <Button block color="primary" size="large" loading={loading} type="submit" className="h-12 rounded-xl text-base font-bold" style={{ background: 'linear-gradient(135deg, #FF6B35, #FF4D94)' }}>
              {loading ? '登录中...' : '登 录'}
            </Button>
          </Form>

          <div className="mt-6">
            <div className="relative"><div className="absolute inset-0 flex items-center"><div className="w-full border-t border-gray-200" /></div><div className="relative flex justify-center text-sm"><span className="px-4 bg-white text-gray-400">其他登录方式</span></div></div>
            <div className="mt-4 flex justify-center"><SocialLoginButtons theme="vibrant" layout="horizontal" /></div>
          </div>

          <div className="mt-6 text-center">
            <button onClick={() => navigate('/')} className="text-gray-400 text-sm hover:text-orange-500 transition-colors">跳过登录，先看看 →</button>
          </div>
        </div>
      </div>
    )
  }

  // Mobile Layout
  return (
    <div className="min-h-screen flex flex-col" style={{ background: 'linear-gradient(135deg, #FF6B35 0%, #FF4D94 100%)' }}>
      <div className="p-4 flex justify-end">
        <button onClick={() => navigate('/')} className="text-white/80 text-sm px-3 py-1.5 rounded-full bg-white/10 backdrop-blur-sm">跳过</button>
      </div>
      <div className="text-center text-white py-8">
        <h1 className="text-3xl font-bold mb-2">🎉 兴趣电商</h1>
        <p className="text-white/70">发现你的热爱</p>
      </div>
      <div className="flex-1 bg-white rounded-t-3xl p-6">
        <div className="relative flex mb-6 bg-gray-100 rounded-xl p-1">
          <div className="absolute top-0 bottom-0 rounded-xl bg-white shadow transition-all duration-300" style={{ left: tabStyle.left, width: tabStyle.width }} />
          <div className={`relative flex-1 py-2.5 text-center font-medium text-sm rounded-xl cursor-pointer z-10 ${loginType === 'password' ? 'text-orange-500' : 'text-gray-500'}`} onClick={() => handleTab('password')}>密码</div>
          <div className={`relative flex-1 py-2.5 text-center font-medium text-sm rounded-xl cursor-pointer z-10 ${loginType === 'sms' ? 'text-orange-500' : 'text-gray-500'}`} onClick={() => handleTab('sms')}>验证码</div>
        </div>
        <Form layout="vertical" onFinish={handleSubmit}>
          <div className="mb-4"><Input placeholder="手机号/用户名" className="h-12 rounded-xl bg-gray-50" /></div>
          {loginType === 'sms' && <div className="mb-4"><Input placeholder="验证码" className="h-12 rounded-xl bg-gray-50" /></div>}
          {loginType === 'password' && <div className="mb-4"><Input type="password" placeholder="密码" className="h-12 rounded-xl bg-gray-50" /></div>}
          <Button block color="primary" size="large" loading={loading} type="submit" className="h-12 rounded-xl font-bold" style={{ background: 'linear-gradient(135deg, #FF6B35, #FF4D94)' }}>
            {loading ? '登录中...' : '登 录'}
          </Button>
        </Form>
        <div className="mt-6"><div className="relative"><div className="absolute inset-0 flex items-center"><div className="w-full border-t border-gray-200" /></div><div className="relative flex justify-center text-sm"><span className="px-4 bg-white text-gray-400">其他方式</span></div></div></div>
        <div className="mt-4 flex justify-center"><SocialLoginButtons theme="vibrant" /></div>
        <div className="mt-6 text-center text-sm text-gray-500">还没有账号？<span className="text-orange-500 font-medium">立即注册</span></div>
      </div>
    </div>
  )
}
