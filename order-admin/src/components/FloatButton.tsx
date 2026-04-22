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
        boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
        transition: 'left 200ms ease-in-out',
      }}
    />
  )
}
