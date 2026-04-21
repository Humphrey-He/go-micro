import React from 'react'
import { ConfigProvider, App as AntdApp } from 'antd'
import { AppRoutes } from '@/routes'

const App: React.FC = () => {
  return (
    <ConfigProvider
      theme={{
        token: {
          colorPrimary: '#1677ff',
          borderRadius: 8,
          colorBgContainer: '#ffffff',
          colorBgElevated: '#ffffff',
          colorBorder: '#e5e7eb',
          colorText: '#1f2937',
          colorTextSecondary: '#6b7280',
          fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif",
          boxShadow: '0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)',
          boxShadowSecondary: '0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)',
        },
        components: {
          Card: {
            paddingLG: 20,
          },
          Table: {
            headerBg: '#f9fafb',
            headerColor: '#374151',
            rowHoverBg: '#f9fafb',
          },
          Button: {
            primaryShadow: 'none',
          },
        },
      }}
    >
      <AntdApp>
        <AppRoutes />
      </AntdApp>
    </ConfigProvider>
  )
}

export default App
