import React from 'react'
import { Space, Button, Dropdown, MenuProps, Tooltip } from 'antd'
import { MoreOutlined, EyeOutlined, StopOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { BusinessTable } from '@/components/BusinessTable'
import { StatusTag } from '@/components/StatusTag'
import { formatAmount, formatDateTime } from '@/utils/format'
import { ORDER_STATUS_MAP, PAYMENT_STATUS_MAP } from '@/utils/constants'
import type { OrderListItem } from '../types/order'

interface OrderTableProps {
  loading: boolean
  data: OrderListItem[]
  total: number
  page: number
  pageSize: number
  onPageChange: (page: number, pageSize: number) => void
  onViewDetail: (order: OrderListItem) => void
  onCancel: (order: OrderListItem) => void
}

export const OrderTable: React.FC<OrderTableProps> = ({
  loading,
  data,
  total,
  page,
  pageSize,
  onPageChange,
  onViewDetail,
  onCancel,
}) => {
  const getActionItems = (record: OrderListItem): MenuProps['items'] => {
    const items: MenuProps['items'] = [
      {
        key: 'view',
        label: '查看详情',
        icon: <EyeOutlined />,
        onClick: () => onViewDetail(record),
      },
    ]

    if (record.status !== 'CANCELED' && record.status !== 'SUCCESS') {
      items.push({
        key: 'cancel',
        label: '取消订单',
        icon: <StopOutlined />,
        danger: true,
        onClick: () => onCancel(record),
      })
    }

    return items
  }

  const columns: ColumnsType<OrderListItem> = [
    {
      title: '订单号',
      dataIndex: 'biz_no',
      key: 'biz_no',
      width: 160,
      fixed: 'left',
      render: (bizNo: string) => (
        <span
          style={{
            fontFamily: 'monospace',
            fontSize: 13,
            fontWeight: 600,
            color: '#1677ff',
          }}
        >
          {bizNo}
        </span>
      ),
    },
    {
      title: '用户ID',
      dataIndex: 'user_id',
      key: 'user_id',
      width: 110,
      render: (userId: string) => (
        <span style={{ color: '#4b5563', fontSize: 13 }}>{userId}</span>
      ),
    },
    {
      title: '订单状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status) => <StatusTag status={status} statusMap={ORDER_STATUS_MAP} />,
    },
    {
      title: '支付状态',
      dataIndex: 'payment_status',
      key: 'payment_status',
      width: 100,
      render: (status) => <StatusTag status={status} statusMap={PAYMENT_STATUS_MAP} />,
    },
    {
      title: '商品数',
      dataIndex: 'item_count',
      key: 'item_count',
      width: 80,
      align: 'center',
      render: (count: number) => (
        <span
          style={{
            background: '#f3f4f6',
            padding: '2px 8px',
            borderRadius: 10,
            fontSize: 12,
            fontWeight: 600,
            color: '#374151',
          }}
        >
          {count}
        </span>
      ),
    },
    {
      title: '订单金额',
      dataIndex: 'total_amount',
      key: 'total_amount',
      width: 120,
      align: 'right',
      render: (amount: number) => (
        <span
          style={{
            fontWeight: 700,
            fontSize: 14,
            color: '#16a34a',
          }}
        >
          ¥{formatAmount(amount)}
        </span>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 170,
      render: (timestamp: number) => (
        <span style={{ color: '#6b7280', fontSize: 12 }}>{formatDateTime(timestamp)}</span>
      ),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 170,
      render: (timestamp: number) => (
        <span style={{ color: '#9ca3af', fontSize: 12 }}>{formatDateTime(timestamp)}</span>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      fixed: 'right',
      render: (_, record) => (
        <Space size={4}>
          <Tooltip title="查看详情">
            <Button
              type="text"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => onViewDetail(record)}
              style={{ color: '#1677ff' }}
            />
          </Tooltip>
          <Dropdown
            menu={{ items: getActionItems(record) }}
            trigger={['click']}
            placement="bottomRight"
          >
            <Button type="text" size="small" icon={<MoreOutlined />} />
          </Dropdown>
        </Space>
      ),
    },
  ]

  return (
    <BusinessTable<OrderListItem>
      loading={loading}
      columns={columns}
      dataSource={data}
      rowKey="order_id"
      scroll={{ x: 1200 }}
      pagination={{
        current: page,
        pageSize,
        total,
      }}
      onPageChange={onPageChange}
    />
  )
}
