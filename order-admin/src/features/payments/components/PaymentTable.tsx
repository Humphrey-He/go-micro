import React from 'react'
import { Table, Tag, Typography } from 'antd'
import type { Payment, PaymentStatus } from '../types/payment'
import { formatDateTimeStr } from '@/utils/format'

const { Text } = Typography

const STATUS_MAP: Record<PaymentStatus, { label: string; color: string; bg: string }> = {
  PENDING: { label: '待支付', color: '#d97706', bg: '#fffbeb' },
  SUCCESS: { label: '已支付', color: '#16a34a', bg: '#f0fdf4' },
  FAILED: { label: '支付失败', color: '#dc2626', bg: '#fef2f2' },
  REFUNDED: { label: '已退款', color: '#7c3aed', bg: '#f5f3ff' },
}

interface PaymentTableProps {
  data: Payment[]
  loading: boolean
  total: number
  page: number
  pageSize: number
  onPageChange: (page: number, pageSize: number) => void
}

export const PaymentTable: React.FC<PaymentTableProps> = ({
  data,
  loading,
  total,
  page,
  pageSize,
  onPageChange,
}) => {
  const columns = [
    {
      title: '支付单号',
      dataIndex: 'payment_id',
      key: 'payment_id',
      width: 180,
      ellipsis: true,
    },
    {
      title: '订单号',
      dataIndex: 'order_id',
      key: 'order_id',
      width: 180,
      ellipsis: true,
    },
    {
      title: '用户ID',
      dataIndex: 'user_id',
      key: 'user_id',
      width: 100,
      ellipsis: true,
    },
    {
      title: '金额(元)',
      dataIndex: 'amount',
      key: 'amount',
      width: 100,
      render: (amount: number) => (
        <Text strong>¥{(amount / 100).toFixed(2)}</Text>
      ),
    },
    {
      title: '支付状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: PaymentStatus) => {
        const config = STATUS_MAP[status] || STATUS_MAP.PENDING
        return (
          <Tag style={{ color: config.color, background: config.bg, border: 'none' }}>
            {config.label}
          </Tag>
        )
      },
    },
    {
      title: '支付方式',
      dataIndex: 'payment_method',
      key: 'payment_method',
      width: 100,
      ellipsis: true,
    },
    {
      title: '交易流水号',
      dataIndex: 'transaction_id',
      key: 'transaction_id',
      width: 160,
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
      title: '支付时间',
      dataIndex: 'paid_at',
      key: 'paid_at',
      width: 160,
      render: (text: string) => formatDateTimeStr(text),
    },
  ]

  return (
    <Table
      columns={columns}
      dataSource={data}
      loading={loading}
      rowKey="payment_id"
      pagination={{
        current: page,
        pageSize,
        total,
        showSizeChanger: true,
        showQuickJumper: true,
        showTotal: (t) => `共 ${t} 条`,
        onChange: onPageChange,
      }}
      scroll={{ x: 1200 }}
    />
  )
}
