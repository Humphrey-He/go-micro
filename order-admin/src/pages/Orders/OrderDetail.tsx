import React, { useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  Card,
  Descriptions,
  Table,
  Button,
  Space,
  Result,
  Spin,
  Row,
  Col,
  Steps,
  Typography,
  Divider,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import {
  ArrowLeftOutlined,
  ShopOutlined,
  CreditCardOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons'
import { StatusTag } from '@/components/StatusTag'
import { formatAmount, formatDateTime } from '@/utils/format'
import {
  ORDER_STATUS_MAP,
  VIEW_STATUS_MAP,
  PAYMENT_STATUS_MAP,
} from '@/utils/constants'
import { useOrderDetail } from '@/features/orders/hooks'
import type { OrderDetailResponse } from '@/features/orders/types/order'

const { Text } = Typography

const getStepStatus = (
  status: string
): 'wait' | 'process' | 'finish' | 'error' => {
  switch (status) {
    case 'SUCCESS':
      return 'finish'
    case 'PROCESSING':
      return 'process'
    case 'FAILED':
    case 'CANCELED':
    case 'TIMEOUT':
      return 'error'
    default:
      return 'wait'
  }
}

const getStepIcon = (status: string, index: number) => {
  switch (status) {
    case 'SUCCESS':
      return <CheckCircleOutlined style={{ color: '#16a34a' }} />
    case 'PROCESSING':
      return <ClockCircleOutlined style={{ color: '#7c3aed' }} />
    case 'FAILED':
    case 'CANCELED':
    case 'TIMEOUT':
      return <CloseCircleOutlined style={{ color: '#dc2626' }} />
    default:
      return index + 1
  }
}

const OrderItemsTable: React.FC<{ items: OrderDetailResponse['items'] }> = ({ items }) => {
  const columns: ColumnsType<{ sku_id: string; quantity: number; price: number }> = [
    {
      title: '商品SKU',
      dataIndex: 'sku_id',
      key: 'sku_id',
      render: (skuId: string) => (
        <Space>
          <div
            style={{
              width: 36,
              height: 36,
              borderRadius: 8,
              background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: '#fff',
              fontWeight: 700,
              fontSize: 12,
            }}
          >
            {skuId.slice(-4)}
          </div>
          <Text strong>{skuId}</Text>
        </Space>
      ),
    },
    {
      title: '单价',
      dataIndex: 'price',
      key: 'price',
      align: 'right',
      render: (price: number) => (
        <Text style={{ fontFamily: 'monospace' }}>¥{formatAmount(price)}</Text>
      ),
    },
    {
      title: '数量',
      dataIndex: 'quantity',
      key: 'quantity',
      align: 'center',
      width: 80,
      render: (qty: number) => (
        <span
          style={{
            background: '#f3f4f6',
            padding: '2px 10px',
            borderRadius: 10,
            fontSize: 13,
            fontWeight: 600,
          }}
        >
          {qty}
        </span>
      ),
    },
    {
      title: '小计',
      key: 'subtotal',
      align: 'right',
      render: (_, record) => (
        <Text
          strong
          style={{ fontSize: 14, color: '#16a34a', fontFamily: 'monospace' }}
        >
          ¥{formatAmount(record.price * record.quantity)}
        </Text>
      ),
    },
  ]

  const totalAmount = items.reduce((sum, item) => sum + item.price * item.quantity, 0)

  return (
    <Table
      columns={columns}
      dataSource={items.map((item, index) => ({ ...item, key: index }))}
      pagination={false}
      size="middle"
      summary={() => (
        <Table.Summary fixed>
          <Table.Summary.Row style={{ background: '#f9fafb' }}>
            <Table.Summary.Cell index={0}>
              <Text type="secondary">合计</Text>
            </Table.Summary.Cell>
            <Table.Summary.Cell index={1} align="right">
              <Text type="secondary">{items.length} 件商品</Text>
            </Table.Summary.Cell>
            <Table.Summary.Cell index={2} align="right">
              <Text strong style={{ fontSize: 16, color: '#16a34a' }}>
                ¥{formatAmount(totalAmount)}
              </Text>
            </Table.Summary.Cell>
          </Table.Summary.Row>
        </Table.Summary>
      )}
    />
  )
}

export const OrderDetailPage: React.FC = () => {
  const { orderNo } = useParams<{ orderNo: string }>()
  const navigate = useNavigate()
  const { loading, data, error, fetchDetail } = useOrderDetail()

  useEffect(() => {
    if (orderNo) {
      fetchDetail(orderNo)
    }
  }, [orderNo, fetchDetail])

  if (error) {
    return (
      <div style={{ padding: 40 }}>
        <Result
          status="error"
          title="获取订单详情失败"
          subTitle={error}
          extra={
            <Button type="primary" onClick={() => navigate('/orders')}>
              返回列表
            </Button>
          }
        />
      </div>
    )
  }

  const currentStep =
    data?.status === 'SUCCESS'
      ? 3
      : data?.status === 'PROCESSING' || data?.status === 'RESERVED'
        ? 2
        : data?.status === 'CREATED'
          ? 1
          : 0

  return (
    <div style={{ padding: '24px 28px' }}>
      {/* Header */}
      <div
        style={{
          marginBottom: 20,
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/orders')}
            style={{ borderRadius: 8 }}
          >
            返回
          </Button>
          {data && (
            <Text strong style={{ fontSize: 16 }}>
              订单详情
            </Text>
          )}
        </Space>
        {data && (
          <Space>
            <StatusTag status={data.status} statusMap={ORDER_STATUS_MAP} />
            {data.payment_status && (
              <StatusTag status={data.payment_status} statusMap={PAYMENT_STATUS_MAP} />
            )}
          </Space>
        )}
      </div>

      <Spin spinning={loading}>
        {data ? (
          <Row gutter={[16, 16]}>
            {/* Order Progress Stepper */}
            <Col span={24}>
              <Card
                style={{ borderRadius: 12, border: '1px solid #e5e7eb', marginBottom: 16 }}
                bodyStyle={{ padding: '24px 40px' }}
              >
                <Steps
                  current={currentStep}
                  size="small"
                  items={[
                    {
                      title: '订单创建',
                      icon: getStepIcon(data.status, 0),
                      status: getStepStatus(data.status),
                    },
                    {
                      title: '支付处理',
                      icon: getStepIcon(data.status, 1),
                      status:
                        data.status === 'CREATED'
                          ? 'wait'
                          : getStepStatus(data.status),
                    },
                    {
                      title: '库存预留',
                      icon: getStepIcon(data.status, 2),
                      status:
                        ['RESERVED', 'SUCCESS'].includes(data.status)
                          ? 'finish'
                          : data.status === 'PROCESSING'
                            ? 'process'
                            : 'wait',
                    },
                    {
                      title: '订单完成',
                      icon:
                        data.status === 'SUCCESS' ? (
                          <CheckCircleOutlined style={{ color: '#16a34a' }} />
                        ) : (
                          4
                        ),
                      status: getStepStatus(data.status),
                    },
                  ]}
                />
              </Card>
            </Col>

            {/* Basic Info */}
            <Col xs={24} lg={12}>
              <Card
                title={
                  <Space>
                    <ShopOutlined style={{ color: '#1677ff' }} />
                    <span>基本信息</span>
                  </Space>
                }
                style={{ borderRadius: 12, border: '1px solid #e5e7eb', height: '100%' }}
                bodyStyle={{ padding: '12px 24px' }}
              >
                <Descriptions column={1} size="small" labelStyle={{ width: 90, color: '#6b7280' }}>
                  <Descriptions.Item label="订单号">
                    <Text strong style={{ fontFamily: 'monospace' }}>{data.biz_no}</Text>
                  </Descriptions.Item>
                  <Descriptions.Item label="订单ID">
                    <Text type="secondary" style={{ fontSize: 12 }}>{data.order_id}</Text>
                  </Descriptions.Item>
                  <Descriptions.Item label="用户ID">
                    <Text>{data.user_id}</Text>
                  </Descriptions.Item>
                  {data.view_status && (
                    <Descriptions.Item label="聚合状态">
                      <StatusTag status={data.view_status} statusMap={VIEW_STATUS_MAP} />
                    </Descriptions.Item>
                  )}
                  <Descriptions.Item label="订单金额">
                    <Text strong style={{ color: '#16a34a', fontSize: 16 }}>
                      ¥{formatAmount(data.total_amount)}
                    </Text>
                  </Descriptions.Item>
                  <Descriptions.Item label="创建时间">
                    {data.created_at ? formatDateTime(data.created_at) : '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label="更新时间">
                    {data.updated_at ? formatDateTime(data.updated_at) : '-'}
                  </Descriptions.Item>
                </Descriptions>
              </Card>
            </Col>

            {/* Payment Info */}
            <Col xs={24} lg={12}>
              <Card
                title={
                  <Space>
                    <CreditCardOutlined style={{ color: '#7c3aed' }} />
                    <span>支付信息</span>
                  </Space>
                }
                style={{ borderRadius: 12, border: '1px solid #e5e7eb', height: '100%' }}
                bodyStyle={{ padding: '12px 24px' }}
              >
                <Descriptions column={1} size="small" labelStyle={{ width: 90, color: '#6b7280' }}>
                  <Descriptions.Item label="支付状态">
                    {data.payment_status ? (
                      <StatusTag status={data.payment_status} statusMap={PAYMENT_STATUS_MAP} />
                    ) : (
                      <Text type="secondary">-</Text>
                    )}
                  </Descriptions.Item>
                  <Descriptions.Item label="订单状态">
                    <StatusTag status={data.status} statusMap={ORDER_STATUS_MAP} />
                  </Descriptions.Item>
                </Descriptions>

                <Divider style={{ margin: '12px 0' }} />

                {data.items && data.items.length > 0 && (
                  <div
                    style={{
                      background: '#f9fafb',
                      borderRadius: 8,
                      padding: '12px 16px',
                    }}
                  >
                    <Row gutter={16}>
                      <Col span={12}>
                        <div style={{ fontSize: 12, color: '#6b7280' }}>商品种类</div>
                        <div style={{ fontSize: 18, fontWeight: 700, color: '#1f2937' }}>
                          {data.items.length}
                        </div>
                      </Col>
                      <Col span={12}>
                        <div style={{ fontSize: 12, color: '#6b7280' }}>商品总数</div>
                        <div style={{ fontSize: 18, fontWeight: 700, color: '#1f2937' }}>
                          {data.items.reduce((sum, i) => sum + i.quantity, 0)}
                        </div>
                      </Col>
                    </Row>
                  </div>
                )}
              </Card>
            </Col>

            {/* Items */}
            <Col span={24}>
              <Card
                title={
                  <Space>
                    <ExclamationCircleOutlined style={{ color: '#0891b2' }} />
                    <span>商品明细</span>
                  </Space>
                }
                style={{ borderRadius: 12, border: '1px solid #e5e7eb' }}
                bodyStyle={{ padding: 0 }}
              >
                <div style={{ padding: '0 0 0 0' }}>
                  <OrderItemsTable items={data.items} />
                </div>
              </Card>
            </Col>
          </Row>
        ) : (
          <Card style={{ textAlign: 'center', padding: 60, borderRadius: 12 }}>
            <Text type="secondary">加载中...</Text>
          </Card>
        )}
      </Spin>
    </div>
  )
}

export default OrderDetailPage
