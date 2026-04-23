/**
 * 社交登录工具函数
 * 实现 OAuth2.0 PKCE 流程
 */

// 生成随机字符串
export const generateRandomString = (length: number = 64): string => {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~'
  const randomValues = crypto.getRandomValues(new Uint8Array(length))
  return Array.from(randomValues)
    .map(v => chars[v % chars.length])
    .join('')
}

// 计算 SHA256 hash，返回 Base64URL 编码
export const generateCodeChallenge = async (codeVerifier: string): Promise<string> => {
  const encoder = new TextEncoder()
  const data = encoder.encode(codeVerifier)
  const hash = await crypto.subtle.digest('SHA-256', data)
  return base64URLEncode(new Uint8Array(hash))
}

// Base64URL 编码
export const base64URLEncode = (bytes: Uint8Array): string => {
  let str = ''
  for (const byte of bytes) {
    str += String.fromCharCode(byte)
  }
  return btoa(str)
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '')
}

// 生成 PKCE 参数
export const generatePKCE = async () => {
  const codeVerifier = generateRandomString(64)
  const codeChallenge = await generateCodeChallenge(codeVerifier)
  return { codeVerifier, codeChallenge }
}

// 存储 PKCE verifier（临时存储，5分钟过期）
const PKCE_STORAGE_KEY = 'social_pkce_verifier'
const PKCE_STATE_KEY = 'social_pkce_state'

export const storePKCE = (verifier: string, state: string) => {
  const expiresAt = Date.now() + 5 * 60 * 1000 // 5分钟
  localStorage.setItem(PKCE_STORAGE_KEY, JSON.stringify({ verifier, expiresAt }))
  localStorage.setItem(PKCE_STATE_KEY, state)
}

export const getStoredPKCE = (): { verifier: string; state: string } | null => {
  try {
    const verifierData = localStorage.getItem(PKCE_STORAGE_KEY)
    const state = localStorage.getItem(PKCE_STATE_KEY)
    if (!verifierData || !state) return null

    const { verifier, expiresAt } = JSON.parse(verifierData)
    if (Date.now() > expiresAt) {
      clearPKCE()
      return null
    }
    return { verifier, state }
  } catch {
    return null
  }
}

export const clearPKCE = () => {
  localStorage.removeItem(PKCE_STORAGE_KEY)
  localStorage.removeItem(PKCE_STATE_KEY)
}

// 微信登录跳转
export const wechatLogin = async (redirectUri: string) => {
  const { codeVerifier, codeChallenge } = await generatePKCE()
  const state = generateRandomString(32)

  storePKCE(codeVerifier, state)

  const appId = import.meta.env.VITE_WECHAT_APP_ID || 'your_wechat_appid'
  const encodedRedirectUri = encodeURIComponent(redirectUri)

  const authUrl = `https://open.weixin.qq.com/connect/qrconnect?appid=${appId}&redirect_uri=${encodedRedirectUri}&response_type=code&scope=snsapi_login&state=${state}&code_challenge=${codeChallenge}&code_challenge_method=S256`

  window.location.href = authUrl
}

// Google 登录（使用 Google Identity Services）
export const googleLogin = () => {
  const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID || 'your_google_clientid'
  const redirectUri = `${window.location.origin}/auth/social/callback/google`

  const authUrl = `https://accounts.google.com/o/oauth2/v2/auth?client_id=${clientId}&redirect_uri=${encodeURIComponent(redirectUri)}&response_type=token&scope=openid%20email%20profile&include_granted_scopes=true&state=random_state_string`

  window.location.href = authUrl
}

// Apple 登录
export const appleLogin = () => {
  const clientId = import.meta.env.VITE_APPLE_CLIENT_ID || 'your_apple_clientid'
  const redirectUri = `${window.location.origin}/auth/social/callback/apple`

  const authUrl = `https://appleid.apple.com/auth/authorize?client_id=${clientId}&redirect_uri=${encodeURIComponent(redirectUri)}&response_type=code%20id_token&scope=name%20email&state=random_state_string`

  window.location.href = authUrl
}

// QQ 登录
export const qqLogin = (callbackURL: string) => {
  const appId = import.meta.env.VITE_QQ_APP_ID
  if (!appId) {
    throw new Error('QQ_APP_ID not configured')
  }
  const redirectURI = encodeURIComponent(callbackURL)
  const scope = 'get_user_info'
  const state = crypto.randomUUID()

  const authURL = `https://graph.qq.com/oauth2.0/authorize?response_type=code&client_id=${appId}&redirect_uri=${redirectURI}&scope=${scope}&state=${state}`

  window.location.href = authURL
}
