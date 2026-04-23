import React, { useState, useEffect, useRef } from 'react'
import { Form, Input, Button, Card, message } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { post } from '@/api/request'
import { useAuthStore } from '@/stores/authStore'
import { paths } from '@/routes/paths'

interface LoginData {
  token: string
  user: {
    user_id: string
    username: string
    role: string
    status: number
    created_at?: string
    updated_at?: string
  }
}

interface LoginFormValues {
  username: string
  password: string
}

export const LoginPage: React.FC = () => {
  const [loading, setLoading] = useState(false)
  const [btnState, setBtnState] = useState<'idle' | 'hover' | 'active'>('idle')
  const [mounted, setMounted] = useState(false)
  const [shimmerPos, setShimmerPos] = useState(-100)
  const navigate = useNavigate()
  const { setAuth } = useAuthStore()
  const shimmerRef = useRef<number | null>(null)

  useEffect(() => {
    setMounted(true)
    return () => {
      if (shimmerRef.current) cancelAnimationFrame(shimmerRef.current)
    }
  }, [])

  useEffect(() => {
    if (btnState === 'hover' && !loading) {
      const animate = () => {
        setShimmerPos((prev) => {
          const next = prev + 2
          if (next > 200) return -100
          return next
        })
        shimmerRef.current = requestAnimationFrame(animate)
      }
      shimmerRef.current = requestAnimationFrame(animate)
    } else {
      if (shimmerRef.current) cancelAnimationFrame(shimmerRef.current)
      setShimmerPos(-100)
    }
    return () => {
      if (shimmerRef.current) cancelAnimationFrame(shimmerRef.current)
    }
  }, [btnState, loading])

  const handleSubmit = async (values: LoginFormValues) => {
    setLoading(true)
    try {
      const data = await post<LoginData>('/auth/login', {
        username: values.username,
        password: values.password,
      })

      setAuth(data.token, {
        userId: data.user.user_id,
        username: data.user.username,
        role: data.user.role,
      })
      message.success('登录成功')
      navigate(paths.orders)
    } catch {
      message.error('登录失败，请检查用户名和密码')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #0f172a 0%, #1e3a5f 50%, #0f172a 100%)',
        backgroundSize: '400% 400%',
        animation: 'bgShift 20s ease infinite',
        padding: '20px',
        position: 'relative',
        overflow: 'hidden',
      }}
    >
      {/* Subtle grid pattern */}
      <div
        style={{
          position: 'absolute',
          inset: 0,
          backgroundImage: `
            linear-gradient(rgba(255,255,255,0.03) 1px, transparent 1px),
            linear-gradient(90deg, rgba(255,255,255,0.03) 1px, transparent 1px)
          `,
          backgroundSize: '60px 60px',
          pointerEvents: 'none',
        }}
      />

      {/* Floating decorative elements */}
      <div
        style={{
          position: 'absolute',
          top: '15%',
          left: '10%',
          width: 300,
          height: 300,
          background: 'radial-gradient(circle, rgba(59, 130, 246, 0.1) 0%, transparent 70%)',
          borderRadius: '50%',
          pointerEvents: 'none',
        }}
      />
      <div
        style={{
          position: 'absolute',
          bottom: '20%',
          right: '15%',
          width: 250,
          height: 250,
          background: 'radial-gradient(circle, rgba(99, 102, 241, 0.08) 0%, transparent 70%)',
          borderRadius: '50%',
          pointerEvents: 'none',
        }}
      />

      <div
        style={{
          width: '100%',
          maxWidth: 420,
          opacity: mounted ? 1 : 0,
          transform: mounted ? 'translateY(0)' : 'translateY(20px)',
          transition: 'opacity 0.6s ease-out, transform 0.6s ease-out',
          position: 'relative',
          zIndex: 1,
        }}
      >
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          {/* Logo */}
          <div
            style={{
              width: 64,
              height: 64,
              borderRadius: 16,
              background: 'linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              margin: '0 auto 20px',
              boxShadow: '0 8px 32px rgba(59, 130, 246, 0.3)',
              position: 'relative',
            }}
          >
            <svg width="32" height="32" viewBox="0 0 24 24" fill="none">
              <path
                d="M3 3h7v7H3zM14 3h7v7h-7zM14 14h7v7h-7zM3 14h7v7H3z"
                stroke="white"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                opacity="0.4"
              />
              <path
                d="M7 7h4v4H7zM14 7h4v4h-4zM14 14h4v4h-4zM7 14h4v4H7z"
                stroke="white"
                strokeWidth="1.5"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </svg>
          </div>
          <h1
            style={{
              fontSize: 26,
              fontWeight: 800,
              color: '#fff',
              marginBottom: 8,
              letterSpacing: 2,
            }}
          >
            订单管理系统
          </h1>
          <p style={{ color: 'rgba(255,255,255,0.6)', fontSize: 14 }}>
            高效管理，智能运营
          </p>
        </div>

        <Card
          style={{
            width: '100%',
            boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.4)',
            borderRadius: 16,
            border: '1px solid rgba(255,255,255,0.1)',
            overflow: 'hidden',
            background: 'rgba(255,255,255,0.95)',
            backdropFilter: 'blur(20px)',
          }}
          bodyStyle={{ padding: '32px 32px' }}
        >
          <div style={{ marginBottom: 24 }}>
            <h2 style={{ fontSize: 18, fontWeight: 700, color: '#1f2937', marginBottom: 4 }}>
              欢迎回来
            </h2>
            <p style={{ color: '#6b7280', fontSize: 13 }}>请登录您的管理员账号</p>
          </div>

          <Form name="login" onFinish={handleSubmit} autoComplete="off" size="large">
            <Form.Item
              name="username"
              rules={[{ required: true, message: '请输入用户名' }]}
              style={{ marginBottom: 20 }}
            >
              <Input
                prefix={<UserOutlined style={{ color: '#9ca3af' }} />}
                placeholder="用户名"
                style={{
                  borderRadius: 10,
                  height: 48,
                  border: '1px solid #e5e7eb',
                  transition: 'all 0.2s ease',
                }}
                className="admin-input"
              />
            </Form.Item>

            <Form.Item
              name="password"
              rules={[{ required: true, message: '请输入密码' }]}
              style={{ marginBottom: 16 }}
            >
              <Input.Password
                prefix={<LockOutlined style={{ color: '#9ca3af' }} />}
                placeholder="密码"
                style={{
                  borderRadius: 10,
                  height: 48,
                  border: '1px solid #e5e7eb',
                  transition: 'all 0.2s ease',
                }}
                className="admin-input"
              />
            </Form.Item>

            <Form.Item style={{ marginBottom: 0 }}>
              <div
                style={{ position: 'relative' }}
                onMouseEnter={() => setBtnState('hover')}
                onMouseLeave={() => setBtnState('idle')}
                onMouseDown={() => setBtnState('active')}
                onMouseUp={() => setBtnState('hover')}
              >
                <Button
                  type="primary"
                  htmlType="submit"
                  block
                  loading={loading}
                  style={{
                    height: 52,
                    borderRadius: 12,
                    fontSize: 16,
                    fontWeight: 700,
                    border: 'none',
                    background: 'linear-gradient(135deg, #1d4ed8 0%, #1e40af 100%)',
                    boxShadow: btnState === 'idle'
                      ? '0 4px 14px rgba(29, 78, 216, 0.4)'
                      : btnState === 'hover'
                        ? '0 8px 25px rgba(29, 78, 216, 0.5)'
                        : '0 2px 8px rgba(29, 78, 216, 0.3)',
                    transform: btnState === 'hover'
                      ? 'translateY(-2px)'
                      : btnState === 'active'
                        ? 'translateY(0)'
                        : 'translateY(0)',
                    transition: 'all 0.25s ease',
                    overflow: 'hidden',
                  }}
                >
                  <span style={{ position: 'relative', zIndex: 1 }}>
                    {loading ? '登录中...' : '登 录'}
                  </span>
                  {/* Shimmer effect */}
                  <span
                    style={{
                      position: 'absolute',
                      top: 0,
                      left: 0,
                      right: 0,
                      bottom: 0,
                      background: 'linear-gradient(90deg, transparent, rgba(255,255,255,0.2), transparent)',
                      transform: `translateX(${shimmerPos}%)`,
                      transition: 'transform 0.1s linear',
                      pointerEvents: 'none',
                    }}
                  />
                </Button>
              </div>
            </Form.Item>
          </Form>

          <div
            style={{
              marginTop: 24,
              padding: '16px',
              background: '#f8fafc',
              borderRadius: 10,
              textAlign: 'center',
              border: '1px solid #e2e8f0',
            }}
          >
            <p style={{ fontSize: 12, color: '#9ca3af', marginBottom: 4 }}>测试账号</p>
            <p style={{ fontSize: 13, color: '#64748b', margin: 0, fontFamily: 'monospace' }}>
              testuser / test123
            </p>
          </div>
        </Card>

        <p
          style={{
            textAlign: 'center',
            color: 'rgba(255,255,255,0.4)',
            fontSize: 12,
            marginTop: 24,
          }}
        >
          © 2026 订单管理系统 · All Rights Reserved
        </p>
      </div>

      <style>{`
        @keyframes bgShift {
          0%, 100% { backgroundPosition: '0% 50%'; }
          50% { backgroundPosition: '100% 50%'; }
        }
        .admin-input:hover {
          border-color: #3b82f6 !important;
        }
        .admin-input:focus {
          border-color: #3b82f6 !important;
          box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1) !important;
        }
      `}</style>
    </div>
  )
}

export default LoginPage
