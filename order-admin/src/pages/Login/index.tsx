import React, { useState } from 'react'
import { Form, Input, Button, Card, message } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { post } from '@/api/request'
import { useAuthStore } from '@/stores/authStore'
import { paths } from '@/routes/paths'

interface LoginResponse {
  code: number
  message: string
  data: {
    token: string
    user: {
      userId: string
      username: string
      role: string
    }
  }
}

interface LoginFormValues {
  username: string
  password: string
}

export const LoginPage: React.FC = () => {
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const { setAuth } = useAuthStore()

  const handleSubmit = async (values: LoginFormValues) => {
    setLoading(true)
    try {
      const res = await post<LoginResponse>('/auth/login', {
        username: values.username,
        password: values.password,
      })

      if (res.code === 0 && res.data) {
        setAuth(res.data.token, res.data.user)
        message.success('登录成功')
        navigate(paths.orders)
      } else {
        message.error(res.message || '登录失败')
      }
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
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        padding: '20px',
      }}
    >
      <div style={{ width: '100%', maxWidth: 420 }}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <div
            style={{
              width: 64,
              height: 64,
              borderRadius: 16,
              background: 'rgba(255,255,255,0.2)',
              backdropFilter: 'blur(10px)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              margin: '0 auto 16px',
              border: '1px solid rgba(255,255,255,0.3)',
            }}
          >
            <svg width="32" height="32" viewBox="0 0 24 24" fill="none">
              <path
                d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"
                stroke="white"
                strokeWidth="2"
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
              letterSpacing: 1,
            }}
          >
            订单管理系统
          </h1>
          <p style={{ color: 'rgba(255,255,255,0.75)', fontSize: 14 }}>
            高效管理，智能运营
          </p>
        </div>

        <Card
          style={{
            width: '100%',
            boxShadow: '0 20px 60px rgba(0,0,0,0.2)',
            borderRadius: 16,
            border: 'none',
            overflow: 'hidden',
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
                style={{ borderRadius: 10, height: 44 }}
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
                style={{ borderRadius: 10, height: 44 }}
              />
            </Form.Item>

            <Form.Item style={{ marginBottom: 0 }}>
              <Button
                type="primary"
                htmlType="submit"
                block
                loading={loading}
                style={{
                  height: 48,
                  borderRadius: 10,
                  fontSize: 16,
                  fontWeight: 600,
                  border: 'none',
                  background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                  boxShadow: '0 4px 14px rgba(102, 126, 234, 0.4)',
                }}
              >
                登 录
              </Button>
            </Form.Item>
          </Form>

          <div
            style={{
              marginTop: 24,
              padding: '16px',
              background: '#f9fafb',
              borderRadius: 10,
              textAlign: 'center',
            }}
          >
            <p style={{ fontSize: 12, color: '#9ca3af', marginBottom: 4 }}>测试账号</p>
            <p style={{ fontSize: 13, color: '#6b7280', margin: 0, fontFamily: 'monospace' }}>
              admin / admin123
            </p>
          </div>
        </Card>

        <p
          style={{
            textAlign: 'center',
            color: 'rgba(255,255,255,0.5)',
            fontSize: 12,
            marginTop: 24,
          }}
        >
          © 2026 订单管理系统 · All Rights Reserved
        </p>
      </div>
    </div>
  )
}

export default LoginPage
