import React, { useState, useEffect, useCallback } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import {
  Card,
  Form,
  Row,
  Col,
  Input,
  Select,
  Button,
  Space,
  DatePicker,
  Typography,
  Badge,
  Tooltip,
  Tag,
} from 'antd'
import {
  SearchOutlined,
  ReloadOutlined,
  ShoppingOutlined,
  FilterOutlined,
  ClearOutlined,
} from '@ant-design/icons'
import { OrderTable } from '@/features/orders/components'
import { useOrderList } from '@/features/orders/hooks'
import { useConfirm } from '@/hooks'
import type { OrderListItem, OrderStatus } from '@/features/orders/types/order'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker
const { Text } = Typography

const STATUS_OPTIONS = [
  { label: '全部状态', value: '' },
  { label: '已创建', value: 'CREATED' },
  { label: '已预留', value: 'RESERVED' },
  { label: '处理中', value: 'PROCESSING' },
  { label: '成功', value: 'SUCCESS' },
  { label: '失败', value: 'FAILED' },
  { label: '已取消', value: 'CANCELED' },
  { label: '超时', value: 'TIMEOUT' },
]

export const OrderListPage: React.FC = () => {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const [form] = Form.useForm()

  const [refreshKey, setRefreshKey] = useState(0)
  const [searchCollapsed, setSearchCollapsed] = useState(false)

  const { confirm } = useConfirm()
  const { loading, data, total, fetchData, cancelOrder } = useOrderList()

  const page = parseInt(searchParams.get('page') || '1', 10)
  const pageSize = parseInt(searchParams.get('pageSize') || '20', 10)

  const activeFilterCount = useCallback(() => {
    let count = 0
    if (searchParams.get('orderNo')) count++
    if (searchParams.get('userId')) count++
    if (searchParams.get('status')) count++
    if (searchParams.get('startTime')) count++
    return count
  }, [searchParams])

  const buildQueryParams = useCallback(() => {
    const order_no = searchParams.get('orderNo') || undefined
    const user_id = searchParams.get('userId') || undefined
    const status = (searchParams.get('status') as OrderStatus) || undefined
    const startTimeStr = searchParams.get('startTime')
    const endTimeStr = searchParams.get('endTime')

    return {
      page,
      page_size: pageSize,
      order_no,
      user_id,
      status: status || undefined,
      start_time: startTimeStr ? parseInt(startTimeStr, 10) : undefined,
      end_time: endTimeStr ? parseInt(endTimeStr, 10) : undefined,
    }
  }, [searchParams, page, pageSize])

  useEffect(() => {
    fetchData(buildQueryParams())
  }, [fetchData, buildQueryParams, refreshKey])

  const handleSearch = useCallback(
    (values: {
      orderNo?: string
      userId?: string
      status?: string
      dateRange?: [dayjs.Dayjs, dayjs.Dayjs]
    }) => {
      const params: Record<string, string | number | undefined> = {}

      if (values.orderNo) params.orderNo = values.orderNo
      if (values.userId) params.userId = values.userId
      if (values.status) params.status = values.status
      if (values.dateRange) {
        params.startTime = values.dateRange[0].startOf('day').unix()
        params.endTime = values.dateRange[1].endOf('day').unix()
      }

      setSearchParams(new URLSearchParams(params as Record<string, string>))
    },
    [setSearchParams]
  )

  const handleReset = useCallback(() => {
    form.resetFields()
    setSearchParams(new URLSearchParams())
  }, [form, setSearchParams])

  const handlePageChange = useCallback(
    (newPage: number, newPageSize: number) => {
      const params: Record<string, string | number | undefined> = {}
      searchParams.forEach((value, key) => {
        params[key] = value
      })

      if (newPage === 1) {
        delete params.page
      } else {
        params.page = newPage
      }

      if (newPageSize === 20) {
        delete params.pageSize
      } else {
        params.pageSize = newPageSize
      }

      setSearchParams(new URLSearchParams(params as Record<string, string>))
    },
    [searchParams, setSearchParams]
  )

  const handleViewDetail = useCallback(
    (order: OrderListItem) => {
      navigate(`/orders/${order.biz_no}`)
    },
    [navigate]
  )

  const handleCancel = useCallback(
    (order: OrderListItem) => {
      confirm({
        title: '取消订单',
        content: `确定要取消订单 ${order.biz_no} 吗？`,
        onConfirm: async () => {
          await cancelOrder(order.order_id)
          setRefreshKey((k) => k + 1)
        },
      })
    },
    [confirm, cancelOrder]
  )

  const filterCount = activeFilterCount()

  return (
    <div style={{ padding: '24px 28px' }}>
      {/* Page Header */}
      <div
        style={{
          marginBottom: 20,
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <div>
          <h1
            style={{
              fontSize: 22,
              fontWeight: 700,
              color: '#1f2937',
              marginBottom: 4,
            }}
          >
            订单管理
          </h1>
          <Space size={16}>
            <Text type="secondary">
              共 <Text strong style={{ color: '#1f2937' }}>{total}</Text> 个订单
            </Text>
            {filterCount > 0 && (
              <Tag color="blue">
                {filterCount} 个筛选条件
              </Tag>
            )}
          </Space>
        </div>
        <Space>
          <Tooltip title="刷新数据">
            <Button
              icon={<ReloadOutlined spin={loading} />}
              onClick={() => setRefreshKey((k) => k + 1)}
            >
              刷新
            </Button>
          </Tooltip>
        </Space>
      </div>

      {/* Filter Card */}
      <Card
        style={{ marginBottom: 16, borderRadius: 12, border: '1px solid #e5e7eb' }}
        bodyStyle={{ padding: searchCollapsed ? '12px 20px' : '16px 20px' }}
      >
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            cursor: 'pointer',
            marginBottom: searchCollapsed ? 0 : 12,
          }}
          onClick={() => setSearchCollapsed(!searchCollapsed)}
        >
          <Space>
            <FilterOutlined style={{ color: '#1677ff' }} />
            <Text strong style={{ fontSize: 14 }}>筛选查询</Text>
            {filterCount > 0 && (
              <Badge count={filterCount} style={{ backgroundColor: '#1677ff' }} />
            )}
          </Space>
          <Button type="link" size="small" style={{ padding: '0 4px' }}>
            {searchCollapsed ? '展开' : '收起'}
          </Button>
        </div>

        {!searchCollapsed && (
          <Form form={form} onFinish={handleSearch} layout="vertical">
            <Row gutter={[16, 8]}>
              <Col xs={24} sm={12} lg={6}>
                <Form.Item name="orderNo" label="订单号" style={{ marginBottom: 8 }}>
                  <Input
                    placeholder="请输入订单号"
                    allowClear
                    prefix={<ShoppingOutlined style={{ color: '#9ca3af' }} />}
                    style={{ borderRadius: 8 }}
                  />
                </Form.Item>
              </Col>
              <Col xs={24} sm={12} lg={6}>
                <Form.Item name="userId" label="用户ID" style={{ marginBottom: 8 }}>
                  <Input
                    placeholder="请输入用户ID"
                    allowClear
                    style={{ borderRadius: 8 }}
                  />
                </Form.Item>
              </Col>
              <Col xs={24} sm={12} lg={6}>
                <Form.Item name="status" label="订单状态" style={{ marginBottom: 8 }}>
                  <Select
                    options={STATUS_OPTIONS}
                    placeholder="请选择状态"
                    allowClear
                    style={{ borderRadius: 8 }}
                  />
                </Form.Item>
              </Col>
              <Col xs={24} sm={12} lg={6}>
                <Form.Item name="dateRange" label="创建时间" style={{ marginBottom: 8 }}>
                  <RangePicker style={{ width: '100%', borderRadius: 8 }} />
                </Form.Item>
              </Col>
            </Row>
            <Row>
              <Col span={24} style={{ textAlign: 'right', marginTop: 8 }}>
                <Space>
                  {filterCount > 0 && (
                    <Button
                      icon={<ClearOutlined />}
                      onClick={handleReset}
                      style={{ borderRadius: 8 }}
                    >
                      清除筛选
                    </Button>
                  )}
                  <Button
                    onClick={handleReset}
                    style={{ borderRadius: 8 }}
                  >
                    重置
                  </Button>
                  <Button
                    type="primary"
                    htmlType="submit"
                    icon={<SearchOutlined />}
                    style={{ borderRadius: 8 }}
                  >
                    搜索
                  </Button>
                </Space>
              </Col>
            </Row>
          </Form>
        )}
      </Card>

      {/* Table Card */}
      <Card
        bodyStyle={{ padding: 0 }}
        style={{ borderRadius: 12, border: '1px solid #e5e7eb' }}
      >
        <div
          style={{
            padding: '14px 20px',
            borderBottom: '1px solid #f3f4f6',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
          }}
        >
          <Space>
            <Text strong style={{ fontSize: 14 }}>订单列表</Text>
          </Space>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {loading ? '加载中...' : `显示 ${data.length} 条，共 ${total} 条`}
          </Text>
        </div>
        <OrderTable
          loading={loading}
          data={data}
          total={total}
          page={page}
          pageSize={pageSize}
          onPageChange={handlePageChange}
          onViewDetail={handleViewDetail}
          onCancel={handleCancel}
        />
      </Card>
    </div>
  )
}

export default OrderListPage
