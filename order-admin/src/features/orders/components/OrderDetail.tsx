import React from 'react'
import { Card, Descriptions, Table, Spin } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { StatusTag } from '@/components/StatusTag'
import { DetailDrawer } from '@/components/DetailDrawer'
import { formatAmount, formatDateTime } from '@/utils/format'
import {
  ORDER_STATUS_MAP,
  VIEW_STATUS_MAP,
  PAYMENT_STATUS_MAP,
} from '@/utils/constants'
import type { OrderDetailResponse } from '../types/order'

interface OrderDetailProps {
  visible: boolean
  loading: boolean
  data: OrderDetailResponse | null
  onClose: () => void
}

export const OrderDetail: React.FC<OrderDetailProps> = ({
  visible,
  loading,
  data,
  onClose,
}) => {
  if (!data) {
    return null
  }

  const itemColumns: ColumnsType<{ sku_id: string; quantity: number; price: number }> = [
    {
      title: 'SKU ID',
      dataIndex: 'sku_id',
      key: 'sku_id',
    },
    {
      title: '数量',
      dataIndex: 'quantity',
      key: 'quantity',
      align: 'right',
    },
    {
      title: '单价',
      dataIndex: 'price',
      key: 'price',
      align: 'right',
      render: (price) => `¥${formatAmount(price)}`,
    },
    {
      title: '小计',
      key: 'subtotal',
      align: 'right',
      render: (_, record) => `¥${formatAmount(record.price * record.quantity)}`,
    },
  ]

  return (
    <DetailDrawer title="订单详情" open={visible} onClose={onClose}>
      <Spin spinning={loading}>
        <Card title="基本信息" style={{ marginBottom: 16 }}>
          <Descriptions column={2} size="small">
            <Descriptions.Item label="订单号">{data.biz_no}</Descriptions.Item>
            <Descriptions.Item label="订单ID">{data.order_id}</Descriptions.Item>
            <Descriptions.Item label="用户ID">{data.user_id}</Descriptions.Item>
            <Descriptions.Item label="订单状态">
              <StatusTag status={data.status} statusMap={ORDER_STATUS_MAP} />
            </Descriptions.Item>
            {data.view_status && (
              <Descriptions.Item label="聚合状态">
                <StatusTag status={data.view_status} statusMap={VIEW_STATUS_MAP} />
              </Descriptions.Item>
            )}
            <Descriptions.Item label="支付状态">
              <StatusTag
                status={data.payment_status || ''}
                statusMap={PAYMENT_STATUS_MAP}
              />
            </Descriptions.Item>
            <Descriptions.Item label="订单金额">
              ¥{formatAmount(data.total_amount)}
            </Descriptions.Item>
            <Descriptions.Item label="创建时间">
              {data.created_at ? formatDateTime(data.created_at) : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="更新时间">
              {data.updated_at ? formatDateTime(data.updated_at) : '-'}
            </Descriptions.Item>
          </Descriptions>
        </Card>

        <Card title="商品信息">
          <Table
            columns={itemColumns}
            dataSource={data.items.map((item, index) => ({
              ...item,
              key: index,
            }))}
            pagination={false}
            size="small"
          />
        </Card>
      </Spin>
    </DetailDrawer>
  )
}
