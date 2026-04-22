import React, { useMemo, useState } from 'react'
import { Card, Row, Col, Spin, Progress, Table, Space, Typography, Segmented, Button } from 'antd'
import {
  ShoppingOutlined,
  DollarOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
  TeamOutlined,
  UnorderedListOutlined,
  CalendarOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import dayjs from 'dayjs'
import { useDashboardStats } from '@/features/dashboard/hooks'
import { formatAmount, formatDateTime } from '@/utils/format'
import { StatusTag } from '@/components/StatusTag'
import { ORDER_STATUS_MAP } from '@/utils/constants'
import type { OrderListItem } from '@/features/orders/types/order'

const { Text } = Typography

interface StatCardProps {
  title: string
  value: number | string
  suffix?: string
  prefix?: React.ReactNode
  bgColor: string
  trend?: number
  subValue?: string
}

const StatCard: React.FC<StatCardProps> = ({
  title,
  value,
  suffix,
  prefix,
  bgColor,
  trend,
  subValue,
}) => (
  <Card
    style={{
      borderRadius: 12,
      border: '1px solid #e5e7eb',
      overflow: 'hidden',
    }}
    bodyStyle={{ padding: '20px 24px' }}
  >
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
      <div>
        <div
          style={{
            fontSize: 13,
            color: '#6b7280',
            fontWeight: 500,
            marginBottom: 8,
          }}
        >
          {title}
        </div>
        <div style={{ display: 'flex', alignItems: 'baseline', gap: 4 }}>
          <span style={{ fontSize: 28, fontWeight: 800, color: '#1f2937', lineHeight: 1 }}>
            {value}
          </span>
          {suffix && (
            <span style={{ fontSize: 14, color: '#6b7280', fontWeight: 500 }}>{suffix}</span>
          )}
        </div>
        {subValue && (
          <div style={{ fontSize: 12, color: '#9ca3af', marginTop: 4 }}>{subValue}</div>
        )}
        {trend !== undefined && (
          <div style={{ marginTop: 6, display: 'flex', alignItems: 'center', gap: 4 }}>
            {trend >= 0 ? (
              <ArrowUpOutlined style={{ fontSize: 12, color: '#16a34a' }} />
            ) : (
              <ArrowDownOutlined style={{ fontSize: 12, color: '#dc2626' }} />
            )}
            <span
              style={{
                fontSize: 12,
                color: trend >= 0 ? '#16a34a' : '#dc2626',
                fontWeight: 600,
              }}
            >
              {Math.abs(trend)}%
            </span>
            <span style={{ fontSize: 12, color: '#9ca3af' }}>较昨日</span>
          </div>
        )}
      </div>
      <div
        style={{
          width: 48,
          height: 48,
          borderRadius: 12,
          background: bgColor,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontSize: 22,
        }}
      >
        {prefix}
      </div>
    </div>
  </Card>
)

interface MiniChartProps {
  data: { label: string; value: number; color: string }[]
  total: number
}

const MiniBarChart: React.FC<MiniChartProps> = ({ data, total }) => (
  <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
    {data.map((item) => (
      <div key={item.label}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            marginBottom: 4,
            fontSize: 13,
          }}
        >
          <span style={{ color: '#374151' }}>{item.label}</span>
          <span style={{ fontWeight: 600, color: '#1f2937' }}>
            {item.value}{' '}
            <span style={{ color: '#9ca3af', fontWeight: 400 }}>
              ({total > 0 ? ((item.value / total) * 100).toFixed(1) : 0}%)
            </span>
          </span>
        </div>
        <div
          style={{
            height: 6,
            background: '#f3f4f6',
            borderRadius: 3,
            overflow: 'hidden',
          }}
        >
          <div
            style={{
              height: '100%',
              width: `${total > 0 ? (item.value / total) * 100 : 0}%`,
              background: item.color,
              borderRadius: 3,
              transition: 'width 0.5s ease',
            }}
          />
        </div>
      </div>
    ))}
  </div>
)

export const DashboardPage: React.FC = () => {
  const navigate = useNavigate()

  const [selectedPeriod, setSelectedPeriod] = useState<'day' | 'week' | 'month'>('day')

  const timeRanges = [
    { label: '今日', value: 'today' },
    { label: '本周', value: 'week' },
    { label: '本月', value: 'month' },
    { label: '本年', value: 'year' },
  ]
  const [selectedRange, setSelectedRange] = useState('today')

  const getTimeRange = (range: string) => {
    const now = dayjs()
    switch (range) {
      case 'today':
        return {
          start_time: now.startOf('day').unix(),
          end_time: now.unix(),
        }
      case 'week':
        return {
          start_time: now.startOf('week').unix(),
          end_time: now.unix(),
        }
      case 'month':
        return {
          start_time: now.startOf('month').unix(),
          end_time: now.unix(),
        }
      case 'year':
        return {
          start_time: now.startOf('year').unix(),
          end_time: now.unix(),
        }
      default:
        return {
          start_time: now.startOf('day').unix(),
          end_time: now.unix(),
        }
    }
  }

  const timeRange = getTimeRange(selectedRange)
  const { loading, data, error } = useDashboardStats({
    ...timeRange,
    period: selectedPeriod,
  })

  const stats = useMemo(
    () => [
      {
        title: '今日订单数',
        value: data?.today_order_count ?? 0,
        prefix: <ShoppingOutlined style={{ color: '#1677ff' }} />,
        bgColor: '#e6f4ff',
        trend: 12,
      },
      {
        title: '今日成交额',
        value: formatAmount(data?.today_order_amount ?? 0),
        suffix: '元',
        prefix: <DollarOutlined style={{ color: '#16a34a' }} />,
        bgColor: '#f0fdf4',
        trend: 8,
      },
      {
        title: '待处理退款',
        value: data?.pending_refund_count ?? 0,
        prefix: <ClockCircleOutlined style={{ color: '#ea580c' }} />,
        bgColor: '#fff7ed',
        trend: -3,
      },
      {
        title: '支付成功率',
        value: ((data?.payment_success_rate ?? 0)).toFixed(1),
        suffix: '%',
        prefix: <CheckCircleOutlined style={{ color: '#7c3aed' }} />,
        bgColor: '#f5f3ff',
        subValue: `${((data?.payment_success_rate ?? 0)).toFixed(1)}% 成功支付`,
      },
      {
        title: '库存预警',
        value: data?.low_stock_sku_count ?? 0,
        prefix: <WarningOutlined style={{ color: '#dc2626' }} />,
        bgColor: '#fef2f2',
      },
      {
        title: '用户总数',
        value: '50',
        prefix: <TeamOutlined style={{ color: '#0891b2' }} />,
        bgColor: '#ecfeff',
        subValue: '活跃用户 48',
      },
    ],
    [data]
  )

  if (error) {
    return (
      <div
        style={{
          padding: 40,
          textAlign: 'center',
          color: '#dc2626',
          fontSize: 15,
        }}
      >
        {error}
      </div>
    )
  }

  const mockRecentOrders: OrderListItem[] = [
    {
      order_id: 'ord-1',
      biz_no: 'ORD20260421001',
      user_id: 'user-001',
      status: 'SUCCESS',
      total_amount: 29900,
      item_count: 2,
      payment_status: 'PAID',
      created_at: Math.floor(Date.now() / 1000) - 3600,
      updated_at: Math.floor(Date.now() / 1000) - 1800,
    },
    {
      order_id: 'ord-2',
      biz_no: 'ORD20260421002',
      user_id: 'user-005',
      status: 'PROCESSING',
      total_amount: 159900,
      item_count: 1,
      payment_status: 'PAID',
      created_at: Math.floor(Date.now() / 1000) - 7200,
      updated_at: Math.floor(Date.now() / 1000) - 3600,
    },
    {
      order_id: 'ord-3',
      biz_no: 'ORD20260421003',
      user_id: 'user-012',
      status: 'CREATED',
      total_amount: 8900,
      item_count: 3,
      payment_status: 'PENDING',
      created_at: Math.floor(Date.now() / 1000) - 10800,
      updated_at: Math.floor(Date.now() / 1000) - 10800,
    },
    {
      order_id: 'ord-4',
      biz_no: 'ORD20260421004',
      user_id: 'user-008',
      status: 'SUCCESS',
      total_amount: 45900,
      item_count: 1,
      payment_status: 'PAID',
      created_at: Math.floor(Date.now() / 1000) - 14400,
      updated_at: Math.floor(Date.now() / 1000) - 10800,
    },
    {
      order_id: 'ord-5',
      biz_no: 'ORD20260421005',
      user_id: 'user-023',
      status: 'CANCELED',
      total_amount: 12900,
      item_count: 2,
      payment_status: 'FAILED',
      created_at: Math.floor(Date.now() / 1000) - 18000,
      updated_at: Math.floor(Date.now() / 1000) - 14400,
    },
  ]

  const mockStatusDistribution = [
    { label: '成功', value: 210, color: '#16a34a' },
    { label: '处理中', value: 95, color: '#7c3aed' },
    { label: '已创建', value: 65, color: '#1677ff' },
    { label: '已预留', value: 42, color: '#0891b2' },
    { label: '超时', value: 23, color: '#ea580c' },
    { label: '已取消', value: 15, color: '#6b7280' },
    { label: '失败', value: 10, color: '#dc2626' },
  ]
  const totalOrders = mockStatusDistribution.reduce((sum, item) => sum + item.value, 0)

  const recentColumns = [
    {
      title: '订单号',
      dataIndex: 'biz_no',
      key: 'biz_no',
      width: 150,
      render: (biz_no: string) => (
        <Text strong style={{ fontFamily: 'monospace', fontSize: 13 }}>
          {biz_no}
        </Text>
      ),
    },
    {
      title: '用户',
      dataIndex: 'user_id',
      key: 'user_id',
      width: 100,
      render: (user_id: string) => <Text type="secondary">{user_id}</Text>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (status: string) => <StatusTag status={status} statusMap={ORDER_STATUS_MAP} />,
    },
    {
      title: '金额',
      dataIndex: 'total_amount',
      key: 'total_amount',
      width: 100,
      align: 'right' as const,
      render: (amount: number) => (
        <Text strong style={{ color: '#16a34a' }}>
          ¥{formatAmount(amount)}
        </Text>
      ),
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
      render: (ts: number) => (
        <Text type="secondary" style={{ fontSize: 12 }}>
          {formatDateTime(ts)}
        </Text>
      ),
    },
  ]

  return (
    <div style={{ padding: '24px 28px' }}>
      <Spin spinning={loading}>
        {/* Page Header */}
        <div
          style={{
            marginBottom: 24,
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
              运营看板
            </h1>
            <Text type="secondary">{dayjs().format('YYYY年MM月DD日 dddd')} · 数据实时更新</Text>
            <div style={{ marginTop: 12, display: 'flex', alignItems: 'center', gap: 16 }}>
  <Space size={4}>
    {timeRanges.map((range) => (
      <Button
        key={range.value}
        type={selectedRange === range.value ? 'primary' : 'default'}
        size="small"
        onClick={() => setSelectedRange(range.value)}
      >
        {range.label}
      </Button>
    ))}
  </Space>
  <Segmented
    value={selectedPeriod}
    onChange={(value) => setSelectedPeriod(value as 'day' | 'week' | 'month')}
    options={[
      { label: '按天', value: 'day' },
      { label: '按周', value: 'week' },
      { label: '按月', value: 'month' },
    ]}
  />
</div>
          </div>
          <div style={{
  fontSize: 13,
  color: '#6b7280',
  background: '#f3f4f6',
  padding: '6px 14px',
  borderRadius: 20,
}}>
  <Space size={8}>
    <CalendarOutlined />
    <span>
      数据统计周期：{dayjs.unix(timeRange.start_time).format('MM月DD日 HH:mm')} - {dayjs.unix(timeRange.end_time).format('MM月DD日 HH:mm')}
    </span>
  </Space>
</div>
        </div>

        {/* Stats Cards */}
        <Row gutter={[16, 16]} style={{ marginBottom: 20 }}>
          {stats.map((stat, idx) => (
            <Col key={idx} xs={24} sm={12} lg={8} xl={4}>
              <StatCard {...stat} />
            </Col>
          ))}
        </Row>

        {/* Charts Row */}
        <Row gutter={[16, 16]} style={{ marginBottom: 20 }}>
          <Col xs={24} lg={14}>
            <Card
              title={
                <Space>
                  <UnorderedListOutlined />
                  <span>订单状态分布</span>
                </Space>
              }
              style={{ borderRadius: 12, border: '1px solid #e5e7eb' }}
              bodyStyle={{ padding: '20px 24px' }}
            >
              <Row gutter={[24, 0]}>
                <Col span={14}>
                  <MiniBarChart data={mockStatusDistribution} total={totalOrders} />
                </Col>
                <Col span={10}>
                  <div
                    style={{
                      height: '100%',
                      display: 'flex',
                      flexDirection: 'column',
                      justifyContent: 'center',
                      alignItems: 'center',
                      background: 'linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%)',
                      borderRadius: 12,
                      padding: 20,
                    }}
                  >
                    <div style={{ fontSize: 42, fontWeight: 800, color: '#16a34a' }}>
                      {totalOrders}
                    </div>
                    <div style={{ fontSize: 14, color: '#6b7280', marginTop: 4 }}>总订单数</div>
                    <Progress
                      type="circle"
                      percent={((210 / totalOrders) * 100).toFixed(1) as unknown as number}
                      size={80}
                      strokeColor="#16a34a"
                      format={(p) => `${p}%`}
                      style={{ marginTop: 12 }}
                    />
                    <div style={{ fontSize: 12, color: '#6b7280', marginTop: 4 }}>成功率</div>
                  </div>
                </Col>
              </Row>
            </Card>
          </Col>

          <Col xs={24} lg={10}>
            <Card
              title={
                <Space>
                  <ClockCircleOutlined />
                  <span>关键指标</span>
                </Space>
              }
              style={{ borderRadius: 12, border: '1px solid #e5e7eb', height: '100%' }}
              bodyStyle={{ padding: '20px 24px' }}
            >
              <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    padding: '12px 16px',
                    background: '#f0fdf4',
                    borderRadius: 10,
                    border: '1px solid #bbf7d0',
                  }}
                >
                  <div>
                    <div style={{ fontSize: 12, color: '#6b7280' }}>今日成交额</div>
                    <div style={{ fontSize: 22, fontWeight: 800, color: '#16a34a' }}>
                      ¥{formatAmount(data?.today_order_amount ?? 0)}
                    </div>
                  </div>
                  <ArrowUpOutlined style={{ fontSize: 20, color: '#16a34a' }} />
                </div>
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    padding: '12px 16px',
                    background: '#e6f4ff',
                    borderRadius: 10,
                    border: '1px solid #bae0ff',
                  }}
                >
                  <div>
                    <div style={{ fontSize: 12, color: '#6b7280' }}>订单完成率</div>
                    <div style={{ fontSize: 22, fontWeight: 800, color: '#1677ff' }}>
                      {((data?.payment_success_rate ?? 0)).toFixed(1)}%
                    </div>
                  </div>
                  <CheckCircleOutlined style={{ fontSize: 20, color: '#1677ff' }} />
                </div>
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    padding: '12px 16px',
                    background: '#fff7ed',
                    borderRadius: 10,
                    border: '1px solid #fed7aa',
                  }}
                >
                  <div>
                    <div style={{ fontSize: 12, color: '#6b7280' }}>退款待处理</div>
                    <div style={{ fontSize: 22, fontWeight: 800, color: '#ea580c' }}>
                      {data?.pending_refund_count ?? 0} 件
                    </div>
                  </div>
                  <WarningOutlined style={{ fontSize: 20, color: '#ea580c' }} />
                </div>
              </div>
            </Card>
          </Col>
        </Row>

        {/* Recent Orders */}
        <Card
          title={
            <Space>
              <ShoppingOutlined />
              <span>最新订单</span>
            </Space>
          }
          extra={
            <Text
              style={{ cursor: 'pointer', color: '#1677ff', fontSize: 13 }}
              onClick={() => navigate('/orders')}
            >
              查看全部 →
            </Text>
          }
          style={{ borderRadius: 12, border: '1px solid #e5e7eb' }}
          bodyStyle={{ padding: 0 }}
        >
          <Table
            columns={recentColumns}
            dataSource={mockRecentOrders}
            rowKey="order_id"
            pagination={false}
            size="middle"
            onRow={(record) => ({
              style: { cursor: 'pointer' },
              onClick: () => navigate(`/orders/${record.biz_no}`),
            })}
          />
        </Card>
      </Spin>
    </div>
  )
}

export default DashboardPage
