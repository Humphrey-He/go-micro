import React, { useState, useCallback } from 'react'
import { Card, Form, Row, Col, Input, Select, Button, Statistic, Typography, Tag, Space, Table, Badge } from 'antd'
import {
  SearchOutlined,
  ReloadOutlined,
  CreditCardOutlined,
  DollarOutlined,
  FilterOutlined,
} from '@ant-design/icons'
import { usePaymentList } from '@/features/payments/hooks/usePaymentList'
import { formatDateTimeStr } from '@/utils/format'
import type { PaymentListParams, PaymentStatus } from '@/features/payments/types/payment'

const { Text, Title } = Typography

const STATUS_MAP: Record<PaymentStatus, { label: string; color: string; bg: string }> = {
  PENDING: { label: '待支付', color: '#d97706', bg: '#fffbeb' },
  SUCCESS: { label: '已支付', color: '#16a34a', bg: '#f0fdf4' },
  FAILED: { label: '支付失败', color: '#dc2626', bg: '#fef2f2' },
  REFUNDED: { label: '已退款', color: '#7c3aed', bg: '#f5f3ff' },
}

const STATUS_OPTIONS = [
  { label: '全部状态', value: '' },
  { label: '待支付', value: 'PENDING' },
  { label: '已支付', value: 'SUCCESS' },
  { label: '支付失败', value: 'FAILED' },
  { label: '已退款', value: 'REFUNDED' },
]

export const PaymentListPage: React.FC = () => {
  const [form] = Form.useForm()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [refreshKey, setRefreshKey] = useState(0)

  const { loading, data, total, fetchData } = usePaymentList()

  const loadData = useCallback(
    (values?: Record<string, unknown>) => {
      const { orderId, status, ...rest } = values as Record<string, unknown>
      const params: PaymentListParams = {
        page,
        page_size: pageSize,
        order_id: orderId as string | undefined,
        status: status as PaymentStatus | undefined,
        ...rest,
      }
      fetchData(params)
    },
    [page, pageSize, fetchData]
  )

  const handleSearch = useCallback(
    (values: Record<string, unknown>) => {
      setPage(1)
      loadData(values)
    },
    [loadData]
  )

  const handleReset = useCallback(() => {
    form.resetFields()
    setPage(1)
    loadData({})
  }, [form, loadData])

  const handlePageChange = useCallback(
    (newPage: number, newPageSize: number) => {
      setPage(newPage)
      setPageSize(newPageSize)
      const values = form.getFieldsValue()
      const { orderId, status } = values as Record<string, string | undefined>
      fetchData({ order_id: orderId, status: status as PaymentStatus | undefined, page: newPage, page_size: newPageSize })
    },
    [form, fetchData]
  )

  const handleReload = useCallback(() => {
    setRefreshKey((k) => k + 1)
    const values = form.getFieldsValue()
    loadData(values)
  }, [form, loadData])

  React.useEffect(() => {
    loadData()
  }, [refreshKey])

  const successPayments = data.filter((p) => p.status === 'SUCCESS')
  const successAmount = successPayments.reduce((sum, p) => sum + p.amount, 0)

  const columns = [
    {
      title: '支付单号',
      dataIndex: 'paymentId',
      key: 'paymentId',
      width: 200,
      ellipsis: true,
      render: (text: string) => <Text copyable={{ text }}>{text}</Text>,
    },
    {
      title: '订单号',
      dataIndex: 'order_id',
      key: 'order_id',
      width: 200,
      ellipsis: true,
    },
    {
      title: '金额(元)',
      dataIndex: 'amount',
      key: 'amount',
      width: 120,
      render: (amount: number) => (
        <Text strong style={{ color: '#16a34a' }}>
          ¥{(amount / 100).toFixed(2)}
        </Text>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: PaymentStatus) => {
        const config = STATUS_MAP[status] || STATUS_MAP.PENDING
        return (
          <Tag style={{ color: config.color, background: config.bg, border: 'none', fontWeight: 500 }}>
            {config.label}
          </Tag>
        )
      },
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 170,
      render: (text: string) => formatDateTimeStr(text),
    },
    {
      title: '支付时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 170,
      render: (text: string, record: { status: PaymentStatus }) =>
        text && record.status === 'SUCCESS' ? formatDateTimeStr(text) : '-',
    },
  ]

  return (
    <div style={{ padding: 24 }}>
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card size="small" hoverable>
            <Statistic
              title="待处理"
              value={data.filter((p) => p.status === 'PENDING').length}
              prefix={<Badge status="warning" />}
              valueStyle={{ color: '#d97706' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" hoverable>
            <Statistic
              title="已支付"
              value={successPayments.length}
              prefix={<DollarOutlined style={{ color: '#16a34a' }} />}
              valueStyle={{ color: '#16a34a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" hoverable>
            <Statistic
              title="实收金额(元)"
              value={(successAmount / 100).toFixed(2)}
              prefix={<CreditCardOutlined style={{ color: '#16a34a' }} />}
              valueStyle={{ color: '#16a34a', fontSize: 20 }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" hoverable>
            <Statistic
              title="已退款"
              value={data.filter((p) => p.status === 'REFUNDED').length}
              valueStyle={{ color: '#7c3aed' }}
            />
          </Card>
        </Col>
      </Row>

      <Card
        title={<Title level={5} style={{ margin: 0 }}>支付订单</Title>}
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={handleReload}>
              刷新
            </Button>
          </Space>
        }
      >
        <Form form={form} layout="vertical" onFinish={handleSearch}>
          <Row gutter={16}>
            <Col span={6}>
              <Form.Item name="orderId" label="订单号">
                <Input placeholder="请输入订单号" allowClear />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="status" label="支付状态">
                <Select options={STATUS_OPTIONS} placeholder="请选择" allowClear />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label=" " colon={false}>
                <Space>
                  <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>
                    查询
                  </Button>
                  <Button icon={<FilterOutlined />} onClick={handleReset}>
                    重置
                  </Button>
                </Space>
              </Form.Item>
            </Col>
          </Row>
        </Form>

        <div style={{ marginBottom: 16 }}>
          <Text type="secondary">共 {total} 条记录</Text>
        </div>

        <Table
          columns={columns}
          dataSource={data}
          loading={loading}
          rowKey="paymentId"
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (t) => `共 ${t} 条`,
            onChange: handlePageChange,
          }}
          scroll={{ x: 900 }}
          size="middle"
        />
      </Card>
    </div>
  )
}

export default PaymentListPage
