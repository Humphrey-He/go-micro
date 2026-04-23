import React, { useState, useEffect } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Layout, Avatar, Dropdown, Space, Button, Tooltip } from 'antd'
import {
  UserOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'
import { NotificationBell } from '@/features/notification/components/NotificationBell'
import { SortableMenu } from '@/components/SortableMenu'
import { FloatButton } from '@/components/FloatButton'
import { useSidebarStore } from '@/stores/sidebarStore'

const { Header, Sider, Content } = Layout

export const BasicLayout: React.FC = () => {
  const navigate = useNavigate()
  const location = useLocation()
  const { userInfo, logout } = useAuthStore()
  const { collapsed, floatMode, toggleCollapse } = useSidebarStore()
  const [isRouteChanging, setIsRouteChanging] = useState(false)

  useEffect(() => {
    setIsRouteChanging(true)
    const timer = setTimeout(() => setIsRouteChanging(false), 150)
    return () => clearTimeout(timer)
  }, [location.pathname])

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
      {/* 悬浮模式下的侧边栏 */}
      <Sider
        trigger={null}
        width={220}
        style={{
          borderRight: '1px solid #e2e8f0',
          background: '#fff',
          position: 'fixed',
          left: floatMode ? 0 : -220,
          top: 0,
          bottom: 0,
          zIndex: 100,
          transition: 'left 250ms cubic-bezier(0.4, 0, 0.2, 1)',
          overflow: 'auto',
          boxShadow: floatMode ? '2px 0 12px rgba(0,0,0,0.08)' : 'none',
        }}
      >
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderBottom: '1px solid #e2e8f0',
            background: 'linear-gradient(135deg, #1e3a5f 0%, #0f172a 100%)',
            position: 'relative',
            overflow: 'hidden',
          }}
        >
          {/* Subtle accent line */}
          <div
            style={{
              position: 'absolute',
              bottom: 0,
              left: 0,
              right: 0,
              height: 3,
              background: 'linear-gradient(90deg, #3b82f6, #1d4ed8)',
            }}
          />
          <span style={{ fontSize: 16, fontWeight: 700, color: '#fff', letterSpacing: 1 }}>
            订单管理系统
          </span>
        </div>
        <SortableMenu collapsed={false} />
      </Sider>

      {/* 折叠模式的侧边栏 */}
      <Sider
        trigger={null}
        width={220}
        collapsedWidth={64}
        style={{
          borderRight: '1px solid #e2e8f0',
          background: '#fff',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
          zIndex: 100,
          transition: 'width 250ms cubic-bezier(0.4, 0, 0.2, 1)',
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderBottom: '1px solid #e2e8f0',
            background: 'linear-gradient(135deg, #1e3a5f 0%, #0f172a 100%)',
            position: 'relative',
            overflow: 'hidden',
            transition: 'all 250ms ease',
          }}
        >
          {/* Subtle accent line */}
          <div
            style={{
              position: 'absolute',
              bottom: 0,
              left: 0,
              right: 0,
              height: 3,
              background: 'linear-gradient(90deg, #3b82f6, #1d4ed8)',
              opacity: collapsed ? 0 : 1,
              transition: 'opacity 200ms ease',
            }}
          />
          {!collapsed && (
            <span style={{ fontSize: 16, fontWeight: 700, color: '#fff', letterSpacing: 1 }}>
              订单管理系统
            </span>
          )}
          {collapsed && (
            <span style={{ fontSize: 16, fontWeight: 700, color: '#fff' }}>订</span>
          )}
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', height: 'calc(100vh - 64px)' }}>
          <div style={{ flex: 1, overflow: 'auto' }}>
            <SortableMenu collapsed={collapsed} />
          </div>

          {/* 折叠/展开按钮 - 优化样式 */}
          <div
            style={{
              padding: '12px 8px',
              borderTop: '1px solid #e2e8f0',
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
            }}
          >
            <Tooltip title={collapsed ? '展开菜单' : '折叠菜单'} placement="right">
              <Button
                type="text"
                shape="circle"
                size="large"
                icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
                onClick={toggleCollapse}
                style={{
                  color: '#64748b',
                  background: 'transparent',
                  transition: 'all 200ms ease',
                }}
                className="admin-collapse-btn"
              />
            </Tooltip>
          </div>
        </div>
      </Sider>

      {/* 悬浮按钮 - 优化样式 */}
      <FloatButton />

      <Layout style={{
        marginLeft: collapsed ? 64 : 220,
        transition: 'margin-left 250ms cubic-bezier(0.4, 0, 0.2, 1)',
      }}>
        <Header
          style={{
            background: '#fff',
            padding: '0 24px',
            display: 'flex',
            justifyContent: 'flex-end',
            alignItems: 'center',
            borderBottom: '1px solid #e2e8f0',
            position: 'sticky',
            top: 0,
            zIndex: 99,
            boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
          }}
        >
          <Space size={16}>
            <NotificationBell />
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight" trigger={['click']}>
              <Space
                style={{
                  cursor: 'pointer',
                  padding: '6px 12px',
                  borderRadius: 8,
                  transition: 'all 200ms ease',
                  border: '1px solid transparent',
                }}
                className="admin-user-menu"
              >
                <Avatar size={32} style={{ background: 'linear-gradient(135deg, #1e3a5f 0%, #0f172a 100%)' }} icon={<UserOutlined />} />
                <span style={{ fontWeight: 500, fontSize: 14, color: '#374151' }}>{userInfo?.username || 'Admin'}</span>
              </Space>
            </Dropdown>
          </Space>
        </Header>
        <Content
          style={{
            background: '#f8fafc',
            minHeight: 'calc(100vh - 64px)',
            position: 'relative',
          }}
        >
          {isRouteChanging && (
            <div
              style={{
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
                height: 3,
                background: 'linear-gradient(90deg, #3b82f6, #1d4ed8, #3b82f6)',
                backgroundSize: '200% 100%',
                zIndex: 100,
                animation: 'routeLoading 1s ease-in-out infinite',
              }}
            />
          )}
          <div
            style={{
              opacity: isRouteChanging ? 0.85 : 1,
              transition: 'opacity 150ms ease-out',
            }}
          >
            <Outlet />
          </div>
        </Content>
      </Layout>

      {/* 添加全局样式 */}
      <style>{`
        .admin-collapse-btn:hover {
          background: #f1f5f9 !important;
          color: #1e3a5f !important;
        }
        .float-btn:hover {
          transform: scale(1.05);
          box-shadow: 0 6px 20px rgba(59, 130, 246, 0.35) !important;
        }
        .admin-user-menu:hover {
          background: #f8fafc !important;
          border-color: #e2e8f0 !important;
        }
        @keyframes routeLoading {
          0% { background-position: 100% 0; }
          50% { background-position: 0% 0; }
          100% { background-position: 100% 0; }
        }
      `}</style>
    </Layout>
  )
}
