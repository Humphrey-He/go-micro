import React, { useState, useEffect, useCallback } from 'react'
import { Card, Row, Col, Table, Typography, Tag, Statistic, Button, Space, Progress, Tooltip } from 'antd'
import {
  ReloadOutlined,
  AppstoreOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons'
import { getInventoryList } from '@/features/inventory/api/inventoryApi'
import type { InventoryItem } from '@/features/inventory/api/inventoryApi'

const { Text, Title } = Typography

const LOW_STOCK_THRESHOLD = 100

interface InventoryRow extends InventoryItem {
  total: number
  usagePercent: number
  isLowStock: boolean
}

export const InventoryPage: React.FC = () => {
  const [data, setData] = useState<InventoryRow[]>([])
  const [loading, setLoading] = useState(false)

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const items = await getInventoryList()
      const rows: InventoryRow[] = items.map((item) => {
        const total = item.available + item.reserved
        return {
          ...item,
          total,
          usagePercent: total > 0 ? Math.round((item.reserved / total) * 100) : 0,
          isLowStock: item.available < LOW_STOCK_THRESHOLD,
        }
      })
      setData(rows)
    } catch (err) {
      console.error('Failed to fetch inventory:', err)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const totalStock = data.reduce((sum, item) => sum + item.available, 0)
  const totalReserved = data.reduce((sum, item) => sum + item.reserved, 0)
  const lowStockCount = data.filter((item) => item.isLowStock).length

  const columns = [
    {
      title: 'SKU ID',
      dataIndex: 'skuId',
      key: 'skuId',
      width: 120,
      render: (text: string) => (
        <Text strong copyable={{ text }}>
          {text}
        </Text>
      ),
    },
    {
      title: '可用库存',
      dataIndex: 'available',
      key: 'available',
      width: 140,
      render: (val: number, record: InventoryRow) => (
        <Space>
          <Text strong style={{ color: record.isLowStock ? '#dc2626' : '#16a34a', fontSize: 16 }}>
            {val}
          </Text>
          {record.isLowStock && (
            <Tooltip title={`库存低于阈值 ${LOW_STOCK_THRESHOLD}`}>
              <ExclamationCircleOutlined style={{ color: '#dc2626' }} />
            </Tooltip>
          )}
        </Space>
      ),
    },
    {
      title: '已预扣',
      dataIndex: 'reserved',
      key: 'reserved',
      width: 120,
      render: (val: number) => (
        <Text type="secondary">{val}</Text>
      ),
    },
    {
      title: '总库存',
      dataIndex: 'total',
      key: 'total',
      width: 100,
      render: (val: number) => <Text>{val}</Text>,
    },
    {
      title: '使用率',
      dataIndex: 'usagePercent',
      key: 'usagePercent',
      width: 200,
      render: (percent: number, _record: InventoryRow) => (
        <Progress
          percent={percent}
          size="small"
          strokeColor={percent > 80 ? '#dc2626' : percent > 60 ? '#d97706' : '#16a34a'}
          style={{ width: 150 }}
        />
      ),
    },
    {
      title: '状态',
      key: 'status',
      width: 100,
      render: (_: unknown, record: InventoryRow) =>
        record.isLowStock ? (
          <Tag color="error">库存预警</Tag>
        ) : (
          <Tag color="success">正常</Tag>
        ),
    },
  ]

  return (
    <div style={{ padding: 24 }}>
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card size="small" hoverable>
            <Statistic
              title="SKU 种类"
              value={data.length}
              prefix={<AppstoreOutlined />}
              valueStyle={{ color: '#1677ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" hoverable>
            <Statistic
              title="总可用库存"
              value={totalStock}
              valueStyle={{ color: '#16a34a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" hoverable>
            <Statistic
              title="总预扣库存"
              value={totalReserved}
              valueStyle={{ color: '#d97706' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" hoverable>
            <Statistic
              title="库存预警"
              value={lowStockCount}
              prefix={<ExclamationCircleOutlined />}
              valueStyle={{ color: lowStockCount > 0 ? '#dc2626' : '#16a34a' }}
            />
          </Card>
        </Col>
      </Row>

      <Card
        title={<Title level={5} style={{ margin: 0 }}>库存概览</Title>}
        extra={
          <Button icon={<ReloadOutlined />} onClick={fetchData}>
            刷新
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={data}
          loading={loading}
          rowKey="skuId"
          pagination={false}
          size="middle"
        />
      </Card>
    </div>
  )
}

export default InventoryPage
