import React from 'react'
import { Tag } from 'antd'

interface StatusTagProps {
  status: string
  statusMap: Record<string, { label: string; color: string; bg: string }>
}

export const StatusTag: React.FC<StatusTagProps> = ({ status, statusMap }) => {
  const config = statusMap[status] || { label: status, color: '#6b7280', bg: '#f9fafb' }
  return (
    <Tag
      style={{
        color: config.color,
        background: config.bg,
        borderColor: config.color,
        fontWeight: 500,
        borderRadius: 6,
      }}
    >
      {config.label}
    </Tag>
  )
}
