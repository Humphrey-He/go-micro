import React from 'react'
import { Button } from 'antd'
import { LeftOutlined, MenuOutlined } from '@ant-design/icons'
import { useSidebarStore } from '@/stores/sidebarStore'

export const FloatButton: React.FC = () => {
  const { floatMode, toggleFloat, collapsed } = useSidebarStore()

  if (!floatMode && !collapsed) return null

  return (
    <Button
      type="primary"
      shape="circle"
      size="large"
      icon={collapsed ? <MenuOutlined /> : <LeftOutlined />}
      onClick={toggleFloat}
      style={{
        position: 'fixed',
        left: collapsed ? 16 : 224,
        top: 80,
        zIndex: 99,
        boxShadow: '0 4px 16px rgba(22, 119, 255, 0.3)',
        transition: 'left 250ms cubic-bezier(0.4, 0, 0.2, 1), box-shadow 150ms ease, transform 150ms ease',
        transform: 'scale(1)',
      }}
      className="float-btn"
    />
  )
}
