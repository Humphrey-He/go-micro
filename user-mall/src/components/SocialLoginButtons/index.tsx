import { Toast } from 'antd-mobile'
import { wechatLogin, googleLogin, appleLogin } from '@/utils/socialLogin'

interface SocialLoginButtonsProps {
  onWechatClick?: () => void
  onGoogleClick?: () => void
  onAppleClick?: () => void
  layout?: 'horizontal' | 'vertical'
}

const WechatIcon = () => (
  <svg viewBox="0 0 24 24" className="w-6 h-6">
    <path fill="#07C160" d="M8.5 11c-.83 0-1.5-.67-1.5-1.5S7.67 8 8.5 8s1.5.67 1.5 1.5-.67 1.5-1.5 1.5zm7 0c-.83 0-1.5-.67-1.5-1.5S14.67 8 15.5 8s1.5.67 1.5 1.5-.67 1.5-1.5 1.5zM12 4c-4.97 0-9 2.69-9 6 0 1.97.94 3.74 2.44 5.02L3.5 18.5l3.5-.5c1.5 1.26 3.28 2.04 5.19 2.23l-.19 1.77 1.98-.52c1.82.19 3.69.19 5.51 0 .91-.09 1.82-.22 2.72-.4l.52-1.98-1.47-.19c1.08-.69 2.01-1.56 2.75-2.58l1.98.5-.5-2.02c1.74-1.58 2.81-3.79 2.81-6.29 0-3.31-4.03-6-9-6zm5.75 8.75c-.19.48-.44.92-.75 1.33l-.72-1.48.72.15zm-2.54 2.54c-.31.41-.66.79-1.05 1.12l-.91-1.47 1.47-.28.49 1.63zm-3.23-1.29l-1.47.91-.28-1.47 1.12-1.05 1.63.49v1.12zm-4.25 1.29c.19.48.44.92.75 1.33l-.72 1.48-.72-.15.69-2.66zm-1.63-3.25l1.47-.91.28 1.47-1.12 1.05-1.63-.49v-1.12zm2.54-2.54c.31-.41.66-.79 1.05-1.12l.91 1.47-1.47.28-.49-1.63zm5.75-1.29l-1.47-.91-.28 1.47 1.12 1.05 1.63-.49v1.12z"/>
  </svg>
)

const GoogleIcon = () => (
  <svg viewBox="0 0 24 24" className="w-6 h-6">
    <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
    <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
    <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
    <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
  </svg>
)

const AppleIcon = () => (
  <svg viewBox="0 0 24 24" className="w-6 h-6">
    <path fill="#000" d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.81-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5.13 1.17-.34 2.35-1.04 3.19-.69.85-1.83 1.51-2.95 1.42-.15-1.15.41-2.35 1.05-3.11z"/>
  </svg>
)

export default function SocialLoginButtons({
  onWechatClick,
  onGoogleClick,
  onAppleClick,
  layout = 'horizontal'
}: SocialLoginButtonsProps) {
  const handleWechat = () => {
    if (onWechatClick) {
      onWechatClick()
    } else {
      try {
        wechatLogin(`${window.location.origin}/auth/social/callback/wechat`)
      } catch {
        Toast.show('微信登录暂不可用')
      }
    }
  }

  const handleGoogle = () => {
    if (onGoogleClick) {
      onGoogleClick()
    } else {
      try {
        googleLogin()
      } catch {
        Toast.show('Google 登录暂不可用')
      }
    }
  }

  const handleApple = () => {
    if (onAppleClick) {
      onAppleClick()
    } else {
      try {
        appleLogin()
      } catch {
        Toast.show('Apple 登录暂不可用')
      }
    }
  }

  return (
    <div className={`flex ${layout === 'horizontal' ? 'justify-center gap-4' : 'flex-col gap-3'}`}>
      <button
        onClick={handleWechat}
        className="w-12 h-12 rounded-full bg-[#07C160] flex items-center justify-center text-white shadow-md hover:shadow-lg hover:scale-105 transition-all duration-200 active:scale-95"
        title="微信登录"
      >
        <WechatIcon />
      </button>

      <button
        onClick={handleGoogle}
        className="w-12 h-12 rounded-full bg-white border border-gray-200 flex items-center justify-center shadow-md hover:shadow-lg hover:scale-105 transition-all duration-200 active:scale-95"
        title="Google 登录"
      >
        <GoogleIcon />
      </button>

      <button
        onClick={handleApple}
        className="w-12 h-12 rounded-full bg-black flex items-center justify-center text-white shadow-md hover:shadow-lg hover:scale-105 transition-all duration-200 active:scale-95"
        title="Apple 登录"
      >
        <AppleIcon />
      </button>
    </div>
  )
}
