import React, { useState, useCallback } from 'react'
import { Card, Form, Row, Col, Input, Select, Button, Statistic, Tag, Space, Table, Typography } from 'antd'
import {
  SearchOutlined,
  ReloadOutlined,
  RollbackOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons'
import { useRefundList } from '@/features/refund/hooks/useRefundList'
import { formatDateTimeStr } from '@/utils/format'
import type { RefundListParams, RefundStatus } from '@/features/refund/types/refund'

const { Text, Title } = Typography

const STATUS_MAP: Record<RefundStatus, { label: string; color: string; bg: string }> = {
  PENDING: { label: '处理中', color: '#d97706', bg: '#fffbeb' },
  SUCCESS: { label: '已退款', color: '#16a34a', bg: '#f0fdf4' },
  FAILED: { label: '退款失败', color: '#dc2626', bg: '#fef2f2' },
}

const STATUS_OPTIONS = [
  { label: '全部状态', value: '' },
  { label: '处理中', value: 'PENDING' },
  { label: '已退款', value: 'SUCCESS' },
  { label: '退款失败', value: 'FAILED' },
]

export const RefundListPage: React.FC = () => {
  const [form] = Form.useForm()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [refreshKey, setRefreshKey] = useState(0)

  const { loading, data, total, fetchData } = useRefundList()

  const loadData = useCallback(
    (values?: Record<string, unknown>) => {
      const vals = values || {}
      const { orderId, status } = vals as Record<string, unknown>
      const params: RefundListParams = {
        page,
        page_size: pageSize,
        order_id: orderId as string | undefined,
        status: status as RefundStatus | undefined,
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
      fetchData({ order_id: orderId, status: status as RefundStatus | undefined, page: newPage, page_size: newPageSize })
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
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: RefundStatus) => {
        const config = STATUS_MAP[status] || STATUS_MAP.PENDING
        return (
          <Tag style={{ color: config.color, background: config.bg, border: 'none', fontWeight: 500 }}>
            {config.label}
          </Tag>
        )
      },
    },
    {
      title: '原因',
      dataIndex: 'reason',
      key: 'reason',
      width: 120,
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
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 170,
      render: (text: string) => formatDateTimeStr(text),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 170,
      render: (text: string) => formatDateTimeStr(text),
    },
  ]

  return (
    <div style={{ padding: 24 }}>
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card size="small" hoverable>
            <Statistic
              title="处理中"
              value={data.filter((r) => r.status === 'PENDING').length}
              prefix={<RollbackOutlined style={{ color: '#d97706' }} />}
              valueStyle={{ color: '#d97706' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" hoverable>
            <Statistic
              title="已退款"
              value={data.filter((r) => r.status === 'SUCCESS').length}
              prefix={<CheckCircleOutlined style={{ color: '#16a34a' }} />}
              valueStyle={{ color: '#16a34a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" hoverable>
            <Statistic
              title="退款失败"
              value={data.filter((r) => r.status === 'FAILED').length}
              prefix={<CloseCircleOutlined style={{ color: '#dc2626' }} />}
              valueStyle={{ color: '#dc2626' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" hoverable>
            <Statistic
              title="总退款数"
              value={total}
              valueStyle={{ color: '#7c3aed' }}
            />
          </Card>
        </Col>
      </Row>

      <Card
        title={<Title level={5} style={{ margin: 0 }}>退款管理</Title>}
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
              <Form.Item name="status" label="退款状态">
                <Select options={STATUS_OPTIONS} placeholder="请选择" allowClear />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label=" " colon={false}>
                <Space>
                  <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>
                    查询
                  </Button>
                  <Button onClick={handleReset}>
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
          rowKey="refund_id"
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (t) => `共 ${t} 条`,
            onChange: handlePageChange,
          }}
          scroll={{ x: 1100 }}
          size="middle"
        />
      </Card>
    </div>
  )
}

export default RefundListPage
