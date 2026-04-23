import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { SpinLoading, Toast } from 'antd-mobile'
import { post } from '@/api/request'
import { useAuthStore } from '@/stores/authStore'
import { getStoredPKCE, clearPKCE } from '@/utils/socialLogin'
import type { SocialProvider } from '@/api/auth'

interface CallbackParams {
  code?: string
  state?: string
  error?: string
  error_description?: string
}

export default function SocialCallback() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { login: setAuth } = useAuthStore()
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading')
  const [errorMsg, setErrorMsg] = useState('')

  useEffect(() => {
    const handleCallback = async () => {
      // 获取 URL 参数
      const params: CallbackParams = {
        code: searchParams.get('code') || undefined,
        state: searchParams.get('state') || undefined,
        error: searchParams.get('error') || undefined,
        error_description: searchParams.get('error_description') || undefined,
      }

      // 错误处理
      if (params.error) {
        setStatus('error')
        setErrorMsg(params.error_description || '授权失败')
        Toast.show(params.error_description || '授权失败')
        setTimeout(() => navigate('/login'), 2000)
        return
      }

      // 判断是哪个平台
      const path = window.location.pathname
      let provider: SocialProvider
      let callbackParams: Record<string, unknown>

      if (path.includes('/wechat')) {
        provider = 'wechat'
        // 微信需要 PKCE
        const pkce = getStoredPKCE()
        if (!pkce || pkce.state !== params.state) {
          setStatus('error')
          setErrorMsg('State 验证失败，请重试')
          Toast.show('授权验证失败')
          clearPKCE()
          setTimeout(() => navigate('/login'), 2000)
          return
        }
        callbackParams = {
          code: params.code,
          code_verifier: pkce.verifier,
          state: params.state,
        }
        clearPKCE()
      } else if (path.includes('/google')) {
        provider = 'google'
        // Google 通过 URL hash 传递 token
        const hashParams = new URLSearchParams(window.location.hash.substring(1))
        const credential = hashParams.get('access_token')
        if (!credential) {
          setStatus('error')
          setErrorMsg('获取凭证失败')
          Toast.show('授权失败')
          setTimeout(() => navigate('/login'), 2000)
          return
        }
        callbackParams = { credential }
        // 清理 hash
        window.history.replaceState(null, '', window.location.pathname)
      } else if (path.includes('/apple')) {
        provider = 'apple'
        // Apple 通过 URL 参数传递
        const identityToken = searchParams.get('id_token') || ''
        const authorizationCode = searchParams.get('code') || ''
        const user = searchParams.get('user') || undefined
        callbackParams = { identity_token: identityToken, authorization_code: authorizationCode, user }
      } else {
        setStatus('error')
        setErrorMsg('未知的登录平台')
        Toast.show('未知错误')
        setTimeout(() => navigate('/login'), 2000)
        return
      }

      try {
        // 调用后端社交登录接口
        const response = await post<{
          token: string
          expires_in: number
          is_new_user: boolean
          user_info: {
            user_id: string
            nickname: string
            avatar: string
            phone_bound: boolean
          }
        }>(`/auth/social/callback/${provider}`, callbackParams)

        // 登录成功
        setAuth(response.token, response.user_info)
        setStatus('success')
        Toast.show('登录成功')

        // 跳转
        if (!response.user_info.phone_bound) {
          // 新用户或未绑定手机号，跳转到绑定手机号页面
          navigate('/auth/social/bind', { state: { isNewUser: response.is_new_user } })
        } else {
          navigate('/')
        }
      } catch (error) {
        setStatus('error')
        setErrorMsg('登录失败，请重试')
        Toast.show('登录失败，请重试')
        setTimeout(() => navigate('/login'), 2000)
      }
    }

    handleCallback()
  }, [searchParams, navigate, setAuth])

  return (
    <div className="min-h-screen bg-white flex flex-col items-center justify-center">
      {status === 'loading' && (
        <>
          <SpinLoading color="primary" />
          <p className="mt-4 text-gray-500">正在处理登录...</p>
        </>
      )}
      {status === 'error' && (
        <>
          <div className="text-6xl mb-4">😢</div>
          <p className="text-gray-700 font-medium">{errorMsg}</p>
          <p className="text-gray-400 text-sm mt-2">即将跳转到登录页...</p>
        </>
      )}
      {status === 'success' && (
        <>
          <div className="text-6xl mb-4">🎉</div>
          <p className="text-gray-700 font-medium">登录成功</p>
          <p className="text-gray-400 text-sm mt-2">正在跳转...</p>
        </>
      )}
    </div>
  )
}
