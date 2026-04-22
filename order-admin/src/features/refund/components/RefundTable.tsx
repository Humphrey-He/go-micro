import React from 'react'
import { Table, Tag, Typography } from 'antd'
import type { Refund, RefundStatus } from '../types/refund'
import { formatDateTimeStr } from '@/utils/format'

const { Text } = Typography

const STATUS_MAP: Record<RefundStatus, { label: string; color: string; bg: string }> = {
  PENDING: { label: '处理中', color: '#d97706', bg: '#fffbeb' },
  SUCCESS: { label: '已退款', color: '#16a34a', bg: '#f0fdf4' },
  FAILED: { label: '退款失败', color: '#dc2626', bg: '#fef2f2' },
}

interface RefundTableProps {
  data: Refund[]
  loading: boolean
  total: number
  page: number
  pageSize: number
  onPageChange: (page: number, pageSize: number) => void
}

export const RefundTable: React.FC<RefundTableProps> = ({
  data,
  loading,
  total,
  page,
  pageSize,
  onPageChange,
}) => {
  const columns = [
    {
      title: '退款单号',
      dataIndex: 'refund_id',
      key: 'refund_id',
      width: 200,
      ellipsis: true,
      render: (text: string) => <Text copyable={{ text }}>{text}</Text>,
    },
    {
      title: '订单号',
      dataIndex: 'order_id',
      key: 'order_id',
      width: 180,
      ellipsis: true,
    },
    {
      title: '退款类型',
      dataIndex: 'refund_type',
      key: 'refund_type',
      width: 100,
      render: (type: string) => {
        const label = type === 'manual' ? '手动退款' : type === 'payment_failed' ? '支付失败' : type === 'order_cancel' ? '订单取消' : type
        return <Tag>{label}</Tag>
      },
    },
    {
      title: '退款状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: RefundStatus) => {
        const config = STATUS_MAP[status] || STATUS_MAP.PENDING
        return (
          <Tag style={{ color: config.color, background: config.bg, border: 'none' }}>
            {config.label}
          </Tag>
        )
      },
    },
    {
      title: '原因',
      dataIndex: 'reason',
      key: 'reason',
      width: 150,
      ellipsis: true,
    },
    {
      title: '重试次数',
      dataIndex: 'retry_count',
      key: 'retry_count',
      width: 80,
      render: (count: number) => (count > 0 ? <Text type="warning">{count}</Text> : count),
    },
    {
      title: '最后错误',
      dataIndex: 'last_error',
      key: 'last_error',
      width: 150,
      ellipsis: true,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (text: string) => formatDateTimeStr(text),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 160,
      render: (text: string) => formatDateTimeStr(text),
    },
  ]

  return (
    <Table
      columns={columns}
      dataSource={data}
      loading={loading}
      rowKey="refund_id"
      pagination={{
        current: page,
        pageSize,
        total,
        showSizeChanger: true,
        showQuickJumper: true,
        showTotal: (t) => `共 ${t} 条`,
        onChange: onPageChange,
      }}
      scroll={{ x: 1400 }}
    />
  )
}
