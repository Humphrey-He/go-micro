import React from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Layout, Menu, Avatar, Dropdown, Space } from 'antd'
import {
  DashboardOutlined,
  ShoppingOutlined,
  CreditCardOutlined,
  RollbackOutlined,
  AppstoreOutlined,
  UserOutlined,
  LogoutOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'
import { NotificationBell } from '@/features/notification/components/NotificationBell'

const { Header, Sider, Content } = Layout

const menuItems = [
  {
    key: '/dashboard',
    icon: <DashboardOutlined />,
    label: '运营看板',
  },
  {
    key: '/orders',
    icon: <ShoppingOutlined />,
    label: '订单管理',
  },
  {
    key: '/payments',
    icon: <CreditCardOutlined />,
    label: '支付管理',
  },
  {
    key: '/refunds',
    icon: <RollbackOutlined />,
    label: '退款管理',
  },
  {
    key: '/inventory',
    icon: <AppstoreOutlined />,
    label: '库存管理',
  },
]

export const BasicLayout: React.FC = () => {
  const navigate = useNavigate()
  const location = useLocation()
  const { userInfo, logout } = useAuthStore()

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key)
  }

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: userInfo?.username || 'Admin',
      disabled: true,
    },
    { type: 'divider' as const },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
    },
  ]

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        theme="light"
        width={220}
        style={{
          borderRight: '1px solid #e5e7eb',
          background: '#fff',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
          zIndex: 100,
        }}
      >
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderBottom: '1px solid #e5e7eb',
            background: 'linear-gradient(135deg, #1677ff 0%, #0958d9 100%)',
          }}
        >
          <span style={{ fontSize: 17, fontWeight: 700, color: '#fff', letterSpacing: 0.5 }}>
            订单管理系统
          </span>
        </div>
        <Menu
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={handleMenuClick}
          style={{
            borderRight: 0,
            marginTop: 8,
            padding: '0 8px',
          }}
        />
      </Sider>
      <Layout style={{ marginLeft: 220 }}>
        <Header
          style={{
            background: '#fff',
            padding: '0 24px',
            display: 'flex',
            justifyContent: 'flex-end',
            alignItems: 'center',
            borderBottom: '1px solid #e5e7eb',
            position: 'sticky',
            top: 0,
            zIndex: 99,
          }}
        >
          <Space size={16}>
            <NotificationBell />
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight" trigger={['click']}>
              <Space style={{ cursor: 'pointer', padding: '4px 8px', borderRadius: 8, transition: 'background 0.2s' }}
                onMouseEnter={(e) => (e.currentTarget.style.background = '#f3f4f6')}
                onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
              >
                <Avatar size={32} style={{ background: '#1677ff' }} icon={<UserOutlined />} />
                <span style={{ fontWeight: 500, fontSize: 14 }}>{userInfo?.username || 'Admin'}</span>
              </Space>
            </Dropdown>
          </Space>
        </Header>
        <Content style={{ background: '#f5f5f5', minHeight: 'calc(100vh - 64px)' }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
