// user-mall/src/pages/Auth/LoginStable.tsx
import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Form, Input, Toast } from 'antd-mobile'
import { login, smsLogin, sendSms } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'
import SocialLoginButtons from '@/components/SocialLoginButtons'

const isDesktop = () => window.innerWidth >= 768

// Background with animated gradient and floating shapes
const StableBackground = () => (
  <div className="fixed inset-0 -z-10 overflow-hidden">
    {/* Base gradient */}
    <div
      className="absolute inset-0"
      style={{
        background: 'linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 25%, #f0fdf4 50%, #ecfeff 75%, #f0f9ff 100%)',
        backgroundSize: '400% 400%',
        animation: 'gradient-shift 15s ease infinite',
      }}
    />

    {/* Floating geometric shapes */}
    <div className="absolute top-20 left-[10%] w-64 h-64 rounded-full bg-teal-200/30 blur-3xl animate-float-slow" />
    <div className="absolute top-[40%] right-[5%] w-96 h-96 rounded-full bg-cyan-200/20 blur-3xl animate-float-slower" />
    <div className="absolute bottom-[20%] left-[20%] w-48 h-48 rounded-full bg-emerald-200/25 blur-3xl animate-float-medium" />
    <div className="absolute top-[10%] right-[30%] w-32 h-32 rounded-full bg-sky-200/20 blur-2xl animate-float-fast" />

    {/* Grid pattern overlay */}
    <div
      className="absolute inset-0 opacity-[0.03]"
      style={{
        backgroundImage: `url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23007185' fill-opacity='1'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E")`,
      }}
    />

    {/* Decorative lines */}
    <svg className="absolute top-0 left-0 w-full h-full opacity-[0.02]" xmlns="http://www.w3.org/2000/svg">
      <line x1="0" y1="20%" x2="100%" y2="20%" stroke="#007185" strokeWidth="1" />
      <line x1="0" y1="40%" x2="100%" y2="40%" stroke="#007185" strokeWidth="1" />
      <line x1="0" y1="60%" x2="100%" y2="60%" stroke="#007185" strokeWidth="1" />
      <line x1="0" y1="80%" x2="100%" y2="80%" stroke="#007185" strokeWidth="1" />
    </svg>
  </div>
)

export default function LoginStable() {
  const navigate = useNavigate()
  const { login: setAuth } = useAuthStore()
  const [loading, setLoading] = useState(false)
  const [loginType, setLoginType] = useState<'password' | 'sms'>('password')
  const [smsCountdown, setSmsCountdown] = useState(0)
  const [form] = Form.useForm()
  const [isDesktopView, setIsDesktopView] = useState(false)
  const [agreed, setAgreed] = useState(false)

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
      await sendSms({ phone, type: 'login' })
      Toast.show('验证码已发送')
      setSmsCountdown(60)
    } catch {
      Toast.show('发送失败')
    }
  }

  const handleSubmit = async (values: any) => {
    if (!agreed) {
      Toast.show('请先阅读并同意用户协议')
      return
    }
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

  const handleCategoryClick = (category: string) => {
    Toast.show(`即将跳转到${category}分类`)
    setTimeout(() => navigate('/'), 500)
  }

  const categories = [
    { name: '数码家电', icon: '📱', color: 'blue' },
    { name: '服饰鞋包', icon: '👟', color: 'purple' },
    { name: '食品生鲜', icon: '🍎', color: 'green' },
    { name: '美妆护肤', icon: '💄', color: 'pink' },
    { name: '家居用品', icon: '🏠', color: 'orange' },
    { name: '图书文具', icon: '📚', color: 'indigo' },
  ]

  const trustBadges = [
    { icon: '🔒', text: '安全支付', desc: '多重加密保障' },
    { icon: '✓', text: '正品保障', desc: '100%正品承诺' },
    { icon: '🚚', text: '快速配送', desc: '准时送达服务' },
    { icon: '💯', text: '售后无忧', desc: '7天无理由退换' },
  ]

  const activities = [
    { tag: '限时特惠', title: '新人专享100元礼包', sub: '立即领取' },
    { tag: '热门', title: '今日秒杀 低至5折', sub: '查看详情' },
    { tag: '推荐', title: '为你精选的好物', sub: '查看全部' },
  ]

  // Desktop Layout
  if (isDesktopView) {
    return (
      <div className="min-h-screen relative">
        <StableBackground />
        {/* Header */}
        <div className="bg-white/80 backdrop-blur-md border-b border-gray-200 px-8 py-4 flex justify-between items-center shadow-sm">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-teal-500 to-teal-600 flex items-center justify-center text-white font-bold text-lg">
              🏪
            </div>
            <div>
              <div className="text-lg font-bold text-gray-800">品质商城</div>
              <div className="text-xs text-gray-500">品质保障 · 值得信赖</div>
            </div>
          </div>
          <div className="flex items-center gap-6">
            <button className="text-gray-500 hover:text-teal-600 text-sm transition-colors">帮助中心</button>
            <button className="text-gray-500 hover:text-teal-600 text-sm transition-colors">联系我们</button>
            <button onClick={() => navigate('/')} className="text-teal-600 text-sm font-medium hover:text-teal-700">
              跳过登录 →
            </button>
          </div>
        </div>

        {/* Main Content */}
        <div className="flex max-w-7xl mx-auto mt-8 gap-10 px-4">
          {/* Left - Form */}
          <div className="flex-1 max-w-lg">
            <div className="bg-white rounded-xl p-8 shadow-sm border border-gray-200">
              <div className="flex items-center gap-3 mb-6">
                <div className="flex-1">
                  <h2 className="text-2xl font-bold text-gray-900">登录</h2>
                  <p className="text-gray-500 text-sm mt-1">欢迎回到品质商城</p>
                </div>
                <div className="px-3 py-1 bg-teal-50 text-teal-600 text-xs font-medium rounded-full">
                  安全登录
                </div>
              </div>

              {/* Tab */}
              <div className="flex gap-6 mb-6 border-b border-gray-100 pb-3">
                <button
                  className={`pb-2 text-sm font-medium transition-colors ${
                    loginType === 'password'
                      ? 'text-teal-600 border-b-2 border-teal-600'
                      : 'text-gray-500 hover:text-gray-700'
                  }`}
                  onClick={() => setLoginType('password')}
                >
                  密码登录
                </button>
                <button
                  className={`pb-2 text-sm font-medium transition-colors ${
                    loginType === 'sms'
                      ? 'text-teal-600 border-b-2 border-teal-600'
                      : 'text-gray-500 hover:text-gray-700'
                  }`}
                  onClick={() => setLoginType('sms')}
                >
                  验证码登录
                </button>
              </div>

              <Form layout="vertical" onFinish={handleSubmit}>
                <div className="mb-4">
                  <label className="text-sm font-medium text-gray-700 mb-1.5 block">邮箱/手机</label>
                  <Input
                    placeholder="请输入手机号或邮箱"
                    className="h-11 rounded-lg bg-gray-50 border-gray-200 focus:border-teal-500"
                  />
                </div>

                {loginType === 'password' && (
                  <div className="mb-4">
                    <label className="text-sm font-medium text-gray-700 mb-1.5 block">密码</label>
                    <Input
                      type="password"
                      placeholder="请输入密码"
                      className="h-11 rounded-lg bg-gray-50 border-gray-200 focus:border-teal-500"
                    />
                  </div>
                )}

                {loginType === 'sms' && (
                  <div className="mb-4">
                    <label className="text-sm font-medium text-gray-700 mb-1.5 block">验证码</label>
                    <div className="flex gap-2">
                      <Input
                        placeholder="请输入验证码"
                        className="h-11 rounded-lg bg-gray-50 border-gray-200 focus:border-teal-500 flex-1"
                      />
                      <button
                        type="button"
                        className={`px-4 h-11 rounded-lg text-sm font-medium transition-colors ${
                          smsCountdown > 0
                            ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                            : 'bg-teal-50 text-teal-600 hover:bg-teal-100'
                        }`}
                        onClick={smsCountdown > 0 ? undefined : handleSendSms}
                        disabled={smsCountdown > 0}
                      >
                        {smsCountdown > 0 ? `${smsCountdown}s` : '获取验证码'}
                      </button>
                    </div>
                  </div>
                )}

                {/* Agreement checkbox */}
                <div className="mb-4 flex items-start gap-2">
                  <input
                    type="checkbox"
                    id="agreement"
                    checked={agreed}
                    onChange={(e) => setAgreed(e.target.checked)}
                    className="mt-1 w-4 h-4 text-teal-600 rounded border-gray-300 focus:ring-teal-500"
                  />
                  <label htmlFor="agreement" className="text-xs text-gray-500 leading-relaxed">
                    我已阅读并同意
                    <button type="button" className="text-teal-600 hover:underline mx-0.5">《用户服务协议》</button>
                    和
                    <button type="button" className="text-teal-600 hover:underline mx-0.5">《隐私政策》</button>
                  </label>
                </div>

                <Button
                  block
                  type="submit"
                  loading={loading}
                  disabled={!agreed}
                  className="h-11 rounded-lg text-base font-medium"
                  style={{ background: agreed ? '#007185' : '#9ca3af' }}
                >
                  {loading ? '登录中...' : '登 录'}
                </Button>
              </Form>

              {/* Social Login */}
              <div className="mt-6">
                <div className="flex items-center gap-3 mb-4">
                  <div className="flex-1 h-px bg-gray-200" />
                  <span className="text-xs text-gray-400">其他登录方式</span>
                  <div className="flex-1 h-px bg-gray-200" />
                </div>
                <SocialLoginButtons theme="stable" layout="horizontal" />
              </div>

              {/* Footer links */}
              <div className="mt-6 flex justify-between text-sm">
                <button type="button" className="text-teal-600 hover:underline">忘记密码？</button>
                <button type="button" onClick={() => navigate('/register')} className="text-teal-600 hover:underline">
                  立即注册 →
                </button>
              </div>
            </div>

            {/* Disclaimer */}
            <div className="mt-4 text-xs text-gray-400 text-center leading-relaxed">
              登录即表示您同意我们的
              <button type="button" className="text-teal-600 hover:underline">服务条款</button>和
              <button type="button" className="text-teal-600 hover:underline">隐私政策</button>。
              我们致力于保护您的个人信息安全。
            </div>
          </div>

          {/* Right - Features & Activities */}
          <div className="flex-1 max-w-md space-y-6">
            {/* Activities */}
            <div className="bg-gradient-to-br from-teal-500 to-cyan-500 rounded-xl p-5 text-white">
              <h3 className="font-bold text-lg mb-3">今日活动</h3>
              <div className="space-y-2">
                {activities.map((act) => (
                  <div key={act.title} className="flex items-center justify-between bg-white/10 rounded-lg px-4 py-3 cursor-pointer hover:bg-white/20 transition-colors">
                    <div>
                      <div className="text-xs opacity-80">{act.tag}</div>
                      <div className="font-medium text-sm">{act.title}</div>
                    </div>
                    <div className="text-xs opacity-80">{act.sub} →</div>
                  </div>
                ))}
              </div>
            </div>

            {/* Categories */}
            <div className="bg-white rounded-xl p-5 shadow-sm border border-gray-200">
              <h3 className="font-bold text-gray-800 mb-4 flex items-center gap-2">
                <span>📂</span> 快捷分类
              </h3>
              <div className="grid grid-cols-3 gap-3">
                {categories.map((cat) => (
                  <button
                    key={cat.name}
                    type="button"
                    onClick={() => handleCategoryClick(cat.name)}
                    className={`p-3 rounded-lg text-center border border-gray-200 hover:border-${cat.color}-300 hover:bg-${cat.color}-50 transition-all`}
                  >
                    <div className="text-2xl mb-1">{cat.icon}</div>
                    <div className="text-xs text-gray-600">{cat.name}</div>
                  </button>
                ))}
              </div>
            </div>

            {/* Trust Badges */}
            <div className="bg-white rounded-xl p-5 shadow-sm border border-gray-200">
              <h3 className="font-bold text-gray-800 mb-4 flex items-center gap-2">
                <span>🛡️</span> 信任保障
              </h3>
              <div className="grid grid-cols-2 gap-4">
                {trustBadges.map((badge) => (
                  <div key={badge.text} className="flex items-start gap-3">
                    <div className="text-xl">{badge.icon}</div>
                    <div>
                      <div className="text-sm font-medium text-gray-800">{badge.text}</div>
                      <div className="text-xs text-gray-500">{badge.desc}</div>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* App Download QR */}
            <div className="bg-white rounded-xl p-5 shadow-sm border border-gray-200">
              <h3 className="font-bold text-gray-800 mb-3 flex items-center gap-2">
                <span>📱</span> 下载APP
              </h3>
              <div className="flex items-center gap-4">
                <div className="w-20 h-20 bg-gray-200 rounded-lg flex items-center justify-center text-gray-400 text-xs">
                  QR码
                </div>
                <div className="text-sm text-gray-600">
                  <div className="font-medium">扫码下载APP</div>
                  <div className="text-xs text-gray-500 mt-1">享受更多优惠</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="text-center text-xs text-gray-400 mt-10 pb-6">
          <div className="flex justify-center gap-4 mb-2">
            <button className="hover:text-gray-600 transition-colors">关于我们</button>
            <span>|</span>
            <button className="hover:text-gray-600 transition-colors">商家入驻</button>
            <span>|</span>
            <button className="hover:text-gray-600 transition-colors">人才招聘</button>
            <span>|</span>
            <button className="hover:text-gray-600 transition-colors">联系我们</button>
          </div>
          <div>© 2026 品质商城 · 品质保障 · ICP备xxxxxxxx号</div>
        </div>
      </div>
    )
  }

  // Mobile Layout
  return (
    <div className="min-h-screen relative">
      <StableBackground />
      {/* Header */}
      <div className="bg-white/80 backdrop-blur-md border-b border-gray-200 px-4 py-3 flex justify-between items-center relative z-10">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-teal-500 to-teal-600 flex items-center justify-center text-white font-bold">
            🏪
          </div>
          <div className="text-base font-bold text-gray-800">品质商城</div>
        </div>
        <button onClick={() => navigate('/')} className="text-teal-600 text-sm font-medium">
          跳过
        </button>
      </div>

      {/* Activities Banner */}
      <div className="bg-gradient-to-r from-teal-500 to-cyan-500 px-4 py-3">
        <div className="flex items-center justify-between text-white text-sm">
          <div>
            <div className="text-xs opacity-80">新用户专享</div>
            <div className="font-bold">100元新人礼包</div>
          </div>
          <button className="bg-white/20 px-3 py-1 rounded-full text-xs font-medium">
            立即领取
          </button>
        </div>
      </div>

      {/* Login Form */}
      <div className="p-4">
        <div className="bg-white rounded-xl p-5 shadow-sm border border-gray-200">
          <h2 className="text-xl font-bold text-gray-900 mb-1">登录</h2>
          <p className="text-gray-500 text-sm mb-4">欢迎回到品质商城</p>

          {/* Tab */}
          <div className="flex gap-3 mb-4 text-sm">
            <button
              className={`px-4 py-2 rounded-lg font-medium ${
                loginType === 'password'
                  ? 'bg-teal-50 text-teal-700'
                  : 'text-gray-500'
              }`}
              onClick={() => setLoginType('password')}
            >
              密码登录
            </button>
            <button
              className={`px-4 py-2 rounded-lg font-medium ${
                loginType === 'sms'
                  ? 'bg-teal-50 text-teal-700'
                  : 'text-gray-500'
              }`}
              onClick={() => setLoginType('sms')}
            >
              验证码
            </button>
          </div>

          <Form layout="vertical" onFinish={handleSubmit}>
            <div className="mb-3">
              <Input
                placeholder="手机号/用户名"
                className="h-11 rounded-lg"
              />
            </div>
            {loginType === 'password' && (
              <div className="mb-3">
                <Input
                  type="password"
                  placeholder="密码"
                  className="h-11 rounded-lg"
                />
              </div>
            )}
            {loginType === 'sms' && (
              <div className="mb-3 flex gap-2">
                <Input
                  placeholder="验证码"
                  className="h-11 rounded-lg flex-1"
                />
                <button
                  type="button"
                  className={`px-4 h-11 rounded-lg text-sm font-medium ${
                    smsCountdown > 0
                      ? 'bg-gray-100 text-gray-400'
                      : 'bg-teal-50 text-teal-600'
                  }`}
                  onClick={smsCountdown > 0 ? undefined : handleSendSms}
                >
                  {smsCountdown > 0 ? `${smsCountdown}s` : '获取'}
                </button>
              </div>
            )}

            {/* Agreement checkbox */}
            <div className="mb-3 flex items-start gap-2">
              <input
                type="checkbox"
                id="agreement-mobile"
                checked={agreed}
                onChange={(e) => setAgreed(e.target.checked)}
                className="mt-1 w-4 h-4 text-teal-600 rounded"
              />
              <label htmlFor="agreement-mobile" className="text-xs text-gray-500 leading-relaxed">
                我已阅读并同意
                <button type="button" className="text-teal-600">《用户协议》</button>
                和
                <button type="button" className="text-teal-600">《隐私政策》</button>
              </label>
            </div>

            <Button
              block
              type="submit"
              loading={loading}
              disabled={!agreed}
              className="h-11 rounded-lg"
              style={{ background: agreed ? '#007185' : '#9ca3af' }}
            >
              {loading ? '登录中...' : '登 录'}
            </Button>
          </Form>

          {/* Social Login */}
          <div className="mt-4 flex justify-center">
            <SocialLoginButtons theme="stable" />
          </div>

          {/* Footer links */}
          <div className="mt-4 flex justify-between text-sm">
            <button type="button" className="text-teal-600">忘记密码</button>
            <button type="button" onClick={() => navigate('/register')} className="text-teal-600">
              注册 →
            </button>
          </div>
        </div>

        {/* Categories */}
        <div className="bg-white rounded-xl p-4 shadow-sm border border-gray-200 mt-4">
          <h3 className="font-bold text-gray-800 mb-3 flex items-center gap-2 text-sm">
            <span>📂</span> 快捷分类
          </h3>
          <div className="grid grid-cols-4 gap-2">
            {categories.slice(0, 4).map((cat) => (
              <button
                key={cat.name}
                type="button"
                onClick={() => handleCategoryClick(cat.name)}
                className="p-2 rounded-lg text-center border border-gray-200 hover:border-teal-300 transition-colors"
              >
                <div className="text-xl mb-0.5">{cat.icon}</div>
                <div className="text-xs text-gray-600">{cat.name}</div>
              </button>
            ))}
          </div>
        </div>

        {/* Trust */}
        <div className="flex justify-center gap-6 mt-4 py-3 text-xs text-gray-500">
          <span>🔒 安全支付</span>
          <span>✓ 正品保障</span>
          <span>🚚 快速配送</span>
        </div>

        {/* Disclaimer */}
        <div className="text-center text-xs text-gray-400 mt-4 pb-6">
          登录即表示您同意我们的服务条款和隐私政策
        </div>
      </div>
    </div>
  )
}
